// Package repl provides a simple Read-Eval-Print Loop CLI interface for the debugger.
package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/ui"
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
}

// CommandHandler is a function that handles a debugger command
type CommandHandler func(args []string) error

// NewREPL creates a new REPL instance
func NewREPL(debugger ui.Debugger) *REPL {
	rl, err := readline.New("(cucaracha) ")
	if err != nil {
		panic(err)
	}

	repl := &REPL{
		debugger: debugger,
		readline: rl,
		writer:   os.Stdout,
		commands: make(map[string]CommandHandler),
	}

	repl.registerCommands()
	return repl
}

// NewREPLWithLoadArgs creates a new REPL instance with load arguments
func NewREPLWithLoadArgs(debugger ui.Debugger, loadArgs *ui.LoadArgs) *REPL {
	rl, err := readline.New("(cucaracha) ")
	if err != nil {
		panic(err)
	}

	repl := &REPL{
		debugger: debugger,
		readline: rl,
		writer:   os.Stdout,
		commands: make(map[string]CommandHandler),
		loadArgs: loadArgs,
	}

	repl.registerCommands()
	return repl
}

// NewREPLWithIO creates a new REPL instance with custom input/output
func NewREPLWithIO(debugger ui.Debugger, reader io.Reader, writer io.Writer) *REPL {
	// Wrap reader to implement io.ReadCloser if needed
	var readCloser io.ReadCloser
	if rc, ok := reader.(io.ReadCloser); ok {
		readCloser = rc
	} else {
		readCloser = io.NopCloser(reader)
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:            "(cucaracha) ",
		Stdin:             readCloser,
		Stdout:            writer,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	})
	if err != nil {
		panic(err)
	}

	repl := &REPL{
		debugger: debugger,
		readline: rl,
		writer:   writer,
		commands: make(map[string]CommandHandler),
	}

	repl.registerCommands()
	return repl
}

// NewREPLWithIOQuiet creates a new REPL instance with custom input/output in quiet mode (no welcome/goodbye messages)
func NewREPLWithIOQuiet(debugger ui.Debugger, reader io.Reader, writer io.Writer) *REPL {
	// Wrap reader to implement io.ReadCloser if needed
	var readCloser io.ReadCloser
	if rc, ok := reader.(io.ReadCloser); ok {
		readCloser = rc
	} else {
		readCloser = io.NopCloser(reader)
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:            "(cucaracha) ",
		Stdin:             readCloser,
		Stdout:            writer,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	})
	if err != nil {
		panic(err)
	}

	repl := &REPL{
		debugger:     debugger,
		readline:     rl,
		writer:       writer,
		commands:     make(map[string]CommandHandler),
		quiet:        true,
		outputFormat: HumanReadable,
	}

	repl.registerCommands()
	return repl
}

// NewREPLWithOutputFormat creates a new REPL instance with a specific output format
func NewREPLWithOutputFormat(debugger ui.Debugger, reader io.Reader, writer io.Writer, format OutputFormat) *REPL {
	// Wrap reader to implement io.ReadCloser if needed
	var readCloser io.ReadCloser
	if rc, ok := reader.(io.ReadCloser); ok {
		readCloser = rc
	} else {
		readCloser = io.NopCloser(reader)
	}

	// In machine-readable mode, suppress readline echo by using a discard writer
	readlineStdout := writer
	if format == MachineReadable {
		readlineStdout = io.Discard
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:            "(cucaracha) ",
		Stdin:             readCloser,
		Stdout:            readlineStdout,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	})
	if err != nil {
		panic(err)
	}

	repl := &REPL{
		debugger:     debugger,
		readline:     rl,
		writer:       writer,
		commands:     make(map[string]CommandHandler),
		quiet:        format == MachineReadable, // Suppress welcome/goodbye in machine readable mode
		outputFormat: format,
	}

	repl.registerCommands()
	return repl
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

	// Program loading commands
	r.commands["load"] = r.handleLoad
	r.commands["loadprogram"] = r.handleLoadProgram
	r.commands["loadsystem"] = r.handleLoadSystem
	r.commands["loadruntime"] = r.handleLoadRuntime
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

	// Check if there's anything to load
	if r.loadArgs.FullDescriptorPath == nil &&
		r.loadArgs.ProgramPath == nil &&
		r.loadArgs.SystemConfigPath == nil &&
		r.loadArgs.Runtime == nil {
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
