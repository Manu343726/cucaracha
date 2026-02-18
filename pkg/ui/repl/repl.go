// Package repl provides a simple Read-Eval-Print Loop CLI interface for the debugger.
package repl

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/ui"
	"github.com/Manu343726/cucaracha/pkg/utils"
	"github.com/Manu343726/cucaracha/pkg/utils/logging"
	"github.com/chzyer/readline"
)

// REPL represents a debugger Read-Eval-Print Loop interface
type REPL struct {
	debugger       ui.Debugger
	readline       *readline.Instance
	writer         io.Writer
	exit           bool
	commands       map[string]CommandHandler
	lastInput      string
	loadArgs       *ui.LoadArgs
	quiet          bool            // When true, no welcome/goodbye messages
	outputFormat   OutputFormat    // Human readable or machine readable
	outputBuffer   strings.Builder // Buffer for collecting output when in machine readable mode
	commandStarted bool            // Track if we're in the middle of processing a command
	scriptFile     string          // Current script file path (for location tracking)
	scriptLine     int             // Current line number in script (for location tracking)
	commandIndex   int             // Current command index (0-based)
	settings       *Settings       // REPL settings (display.events, etc)
	uiSink         *logging.Sink   // UI sink for capturing log entries
}

// CommandHandler is a function that handles a debugger command
type CommandHandler func(args []string) error

// readlineOptions holds configuration for readline creation
type readlineOptions struct {
	reader          io.Reader
	writer          io.Writer
	useCustomConfig bool
}

// createReadline creates a readline instance with the given options
func createReadline(opts readlineOptions) *readline.Instance {
	if !opts.useCustomConfig {
		// Simple readline with default prompt
		rl, err := readline.New("(cucaracha) ")
		if err != nil {
			panic(err)
		}
		return rl
	}

	// Wrap reader to implement io.ReadCloser if needed
	var readCloser io.ReadCloser
	if rc, ok := opts.reader.(io.ReadCloser); ok {
		readCloser = rc
	} else {
		readCloser = io.NopCloser(opts.reader)
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:            "(cucaracha) ",
		Stdin:             readCloser,
		Stdout:            opts.writer,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	})
	if err != nil {
		panic(err)
	}
	return rl
}

// newREPLInternal creates a new REPL instance with the provided configuration
func newREPLInternal(debugger ui.Debugger, readline *readline.Instance, writer io.Writer,
	loadArgs *ui.LoadArgs, quiet bool, outputFormat OutputFormat) *REPL {
	repl := &REPL{
		debugger:     debugger,
		readline:     readline,
		writer:       writer,
		commands:     make(map[string]CommandHandler),
		loadArgs:     loadArgs,
		quiet:        quiet,
		outputFormat: outputFormat,
		settings:     NewSettings(),
	}

	repl.registerCommands()
	return repl
}

// NewREPL creates a new REPL instance
func NewREPL(debugger ui.Debugger) *REPL {
	rl := createReadline(readlineOptions{useCustomConfig: false})
	return newREPLInternal(debugger, rl, os.Stdout, nil, false, HumanReadable)
}

// NewREPLWithLoadArgs creates a new REPL instance with load arguments
func NewREPLWithLoadArgs(debugger ui.Debugger, loadArgs *ui.LoadArgs) *REPL {
	rl := createReadline(readlineOptions{useCustomConfig: false})
	return newREPLInternal(debugger, rl, os.Stdout, loadArgs, false, HumanReadable)
}

// NewREPLWithIO creates a new REPL instance with custom input/output
func NewREPLWithIO(debugger ui.Debugger, reader io.Reader, writer io.Writer) *REPL {
	rl := createReadline(readlineOptions{
		reader:          reader,
		writer:          writer,
		useCustomConfig: true,
	})
	return newREPLInternal(debugger, rl, writer, nil, false, HumanReadable)
}

// NewREPLWithIOQuiet creates a new REPL instance with custom input/output in quiet mode (no welcome/goodbye messages)
func NewREPLWithIOQuiet(debugger ui.Debugger, reader io.Reader, writer io.Writer) *REPL {
	rl := createReadline(readlineOptions{
		reader:          reader,
		writer:          writer,
		useCustomConfig: true,
	})
	return newREPLInternal(debugger, rl, writer, nil, true, HumanReadable)
}

// NewREPLWithOutputFormat creates a new REPL instance with a specific output format
func NewREPLWithOutputFormat(debugger ui.Debugger, reader io.Reader, writer io.Writer, format OutputFormat) *REPL {
	// In machine-readable mode, suppress readline echo by using a discard writer
	readlineStdout := writer
	if format == MachineReadable {
		readlineStdout = io.Discard
	}

	rl := createReadline(readlineOptions{
		reader:          reader,
		writer:          readlineStdout,
		useCustomConfig: true,
	})
	return newREPLInternal(debugger, rl, writer, nil, format == MachineReadable, format)
}

// ApplySettingsFromFile loads and applies settings from a YAML file
func (r *REPL) ApplySettingsFromFile(filePath string) error {
	return r.settings.LoadFromFile(filePath)
}

// ApplySettingsKeyValue applies a single key=value setting string
func (r *REPL) ApplySettingsKeyValue(kvStr string) error {
	return r.settings.ApplyKeyValue(kvStr)
}

// registerCommands registers all available REPL commands
func (r *REPL) registerCommands() {
	r.commands["help"] = r.handleHelp
	r.commands["h"] = r.handleHelp
	r.commands["exit"] = r.handleExit
	r.commands["quit"] = r.handleExit
	r.commands["q"] = r.handleExit

	// Execution commands
	r.commands["step"] = r.handleStep
	r.commands["s"] = r.handleStep
	r.commands["continue"] = r.handleContinue
	r.commands["c"] = r.handleContinue
	r.commands["interrupt"] = r.handleInterrupt
	r.commands["run"] = r.handleRun
	r.commands["r"] = r.handleRun
	r.commands["reset"] = r.handleReset
	r.commands["restart"] = r.handleRestart

	// Breakpoint commands
	r.commands["break"] = r.handleBreak
	r.commands["b"] = r.handleBreak
	r.commands["removebreakpoint"] = r.handleRemoveBreakpoint
	r.commands["rbp"] = r.handleRemoveBreakpoint
	r.commands["watch"] = r.handleWatch
	r.commands["w"] = r.handleWatch
	r.commands["removewatchpoint"] = r.handleRemoveWatchpoint
	r.commands["rw"] = r.handleRemoveWatchpoint
	r.commands["list"] = r.handleList
	r.commands["l"] = r.handleList

	// Inspection commands
	r.commands["disasm"] = r.handleDisasm
	r.commands["d"] = r.handleDisasm
	r.commands["current"] = r.handleCurrent
	r.commands["memory"] = r.handleMemory
	r.commands["m"] = r.handleMemory
	r.commands["source"] = r.handleSource
	r.commands["info"] = r.handleInfo
	r.commands["i"] = r.handleInfo
	r.commands["registers"] = r.handleRegisters
	r.commands["reg"] = r.handleRegisters
	r.commands["stack"] = r.handleStack
	r.commands["st"] = r.handleStack
	r.commands["vars"] = r.handleVars
	r.commands["v"] = r.handleVars
	r.commands["eval"] = r.handleEval
	r.commands["e"] = r.handleEval
	r.commands["symbols"] = r.handleSymbols
	r.commands["sym"] = r.handleSymbols

	// Program loading commands
	r.commands["load"] = r.handleLoad
	r.commands["loadprogram"] = r.handleLoadProgram
	r.commands["loadsystem"] = r.handleLoadSystem
	r.commands["loadruntime"] = r.handleLoadRuntime

	// Settings commands
	r.commands["set"] = r.handleSet
	r.commands["get"] = r.handleGet

	// Utility commands
	r.commands["loggers"] = r.handleLoggers
}

// Run starts the REPL interactive loop
func (r *REPL) Run() error {
	defer r.readline.Close()

	// Initialize debugger with load arguments if provided
	if r.loadArgs != nil {
		if err := r.initializeDebugger(); err != nil {
			r.printError(fmt.Sprintf("Failed to initialize debugger: %v", err))
		}
	}

	// Create UI sink (but don't register it yet - will be applied to specific loggers)
	r.uiSink = logging.NewUISink("ui-repl", slog.LevelDebug, 1000, r.printLogEntry)

	// Set up event callback if display.events is enabled
	displayEvents, _ := r.settings.GetBool(SettingKeyDisplayEvents)
	if displayEvents {
		r.debugger.SetEventCallback(func(event *ui.DebuggerEvent) {
			r.handleDebuggerEvent(event)
		})
	}

	if !r.quiet {
		r.printWelcome()
	}

	for !r.exit {
		line, err := r.readline.Readline()
		if err != nil {
			if err == readline.ErrInterrupt || err.Error() == "Interrupt" {
				continue
			}
			if err.Error() == "EOF" {
				break
			}
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		r.lastInput = line
		r.processCommand(line)
	}

	if !r.quiet {
		r.printGoodbye()
	}
	return nil
}

// initializeDebugger sets up the debugger with the provided load arguments
func (r *REPL) initializeDebugger() error {
	// Only attempt to initialize if we have meaningful load args
	if r.loadArgs == nil {
		return nil
	}

	if r.loadArgs.SystemConfigPath == nil {
		r.loadArgs.SystemConfigPath = utils.Ptr("default")
	}

	if r.loadArgs.Runtime == nil {
		r.loadArgs.Runtime = utils.Ptr(ui.RuntimeTypeInterpreter)
	}

	// Check if there's anything to load
	if r.loadArgs.FullDescriptorPath == nil &&
		r.loadArgs.ProgramPath == nil {
		return nil
	}

	cmd := &ui.DebuggerCommand{
		Command:  ui.DebuggerCommandLoad,
		LoadArgs: r.loadArgs,
	}

	result, err := r.debugger.Execute(cmd)
	if err != nil {
		return err
	}

	// Print the result if there's something to display
	if result != nil {
		r.printCommandResult(result)
	}

	return nil
}

// processCommand parses and executes a command
func (r *REPL) processCommand(input string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}

	cmd := strings.ToLower(parts[0])
	args := parts[1:]

	handler, exists := r.commands[cmd]
	if !exists {
		r.startCommandOutput()
		r.printError(fmt.Sprintf("Unknown command: %s", cmd))
		r.finishCommandOutput(false, fmt.Errorf("unknown command: %s", cmd))
		return
	}

	r.startCommandOutput()
	err := handler(args)
	if err != nil {
		r.printError(err.Error())
		r.finishCommandOutput(false, err)
	} else {
		r.finishCommandOutput(true, nil)
	}
}

// RunScript executes a script file containing debugger commands
func (r *REPL) RunScript(scriptPath string) error {
	defer r.readline.Close()

	// Open the script file
	file, err := os.Open(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to open script file: %w", err)
	}
	defer file.Close()

	// Set script file for location tracking
	r.scriptFile = scriptPath
	r.commandIndex = 0

	// Create UI sink (but don't register it yet - will be applied to specific loggers)
	r.uiSink = logging.NewUISink("ui-repl", slog.LevelDebug, 1000, r.printLogEntry)

	// Only initialize debugger if we have a program to debug
	// Script mode typically just provides commands, not program setup
	if r.loadArgs != nil && r.loadArgs.ProgramPath != nil {
		if err := r.initializeDebugger(); err != nil {
			r.printError(fmt.Sprintf("Failed to initialize debugger: %v", err))
		}
	}

	if !r.quiet {
		r.printWelcome()
		fmt.Fprintf(r.writer, "Running script: %s\n\n", scriptPath)
	}

	// Read and execute commands from the script
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() && !r.exit {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Set current line for location tracking
		r.scriptLine = lineNum

		// Print the command being executed (only in human-readable mode)
		if r.outputFormat == HumanReadable {
			fmt.Fprintf(r.writer, "> %s\n", line)
		}

		// Execute the command
		r.lastInput = line
		r.processCommand(line)

		// Increment command index after execution
		r.commandIndex++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading script file: %w", err)
	}

	if !r.quiet {
		r.printGoodbye()
	}
	return nil
}

// applyUILoggers applies the UI sink to the specified logger names.
// This removes the UI sink from any previously configured loggers and applies it to the new list.
func (r *REPL) applyUILoggers(loggerNames []string) error {
	if r.uiSink == nil {
		return fmt.Errorf("UI sink not initialized")
	}

	registry := logging.DefaultRegistry()

	registry.RemoveSinkFromLoggers(r.uiSink)

	if len(loggerNames) == 0 {
		// If no loggers specified, we just remove the UI sink from all loggers and return
		return nil
	}

	if err := registry.AddSinkToLoggers(r.uiSink, loggerNames); err != nil {
		return fmt.Errorf("failed to apply UI sink to loggers: %w", err)
	}

	// Store the logger names in the setting
	return r.settings.Set(SettingKeyShowLogs, loggerNames)
}
