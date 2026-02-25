// Package repl provides a simple Read-Eval-Print Loop CLI interface for the debugger.
package repl

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	debuggerUI "github.com/Manu343726/cucaracha/pkg/ui/debugger"
	"github.com/Manu343726/cucaracha/pkg/utils"
	"github.com/Manu343726/cucaracha/pkg/utils/logging"
	"github.com/chzyer/readline"
)

// REPL represents a debugger Read-Eval-Print Loop interface
type REPL struct {
	debugger         debuggerUI.CommandBasedDebugger
	readline         *readline.Instance
	writer           io.Writer
	exit             bool
	commands         map[string]CommandHandler
	aliases          map[string]*Alias // Map of alias name to Alias
	lastInput        string
	loadArgs         *debuggerUI.LoadArgs
	lastDisasmArgs   *debuggerUI.DisasmArgs // Store last disasm args for output formatting
	quiet            bool                   // When true, no welcome/goodbye messages
	outputFormat     OutputFormat           // Human readable or machine readable
	outputBuffer     strings.Builder        // Buffer for collecting output when in machine readable mode
	commandStarted   bool                   // Track if we're in the middle of processing a command
	scriptFile       string                 // Current script file path (for location tracking)
	scriptLine       int                    // Current line number in script (for location tracking)
	commandIndex     int                    // Current command index (0-based)
	settings         *Settings              // REPL settings (display.events, etc)
	uiSink           *logging.Sink          // UI sink for capturing log entries
	waitingForInput  bool                   // Track if REPL is waiting for user input
	definingAlias    bool                   // Are we currently defining a multi-line alias?
	defineAliasName  string                 // Name of the alias being defined
	defineCommands   [][]string             // Commands being collected for the alias
	defineDoc        string                 // Documentation for the alias being defined
	settingsFilePath string                 // Path to the loaded settings file
}

// CommandHandler is a function that handles a debugger command
type CommandHandler func(args []string) error

// Alias represents a command alias with optional documentation
type Alias struct {
	Commands [][]string // Each element is a command with its arguments (supports multi-command aliases)
	Doc      string     // Optional documentation
}

// readlineOptions holds configuration for readline creation
type readlineOptions struct {
	reader          io.Reader
	writer          io.Writer
	useCustomConfig bool
}

// getHistoryFilePath returns the path to the REPL history file
func getHistoryFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	cucarachaDir := filepath.Join(homeDir, ".cucaracha")
	historyFile := filepath.Join(cucarachaDir, "history")

	// Create the .cucaracha directory if it doesn't exist
	if err := os.MkdirAll(cucarachaDir, 0755); err != nil {
		return "", err
	}

	return historyFile, nil
}

// createReadline creates a readline instance with the given options
func createReadline(opts readlineOptions) *readline.Instance {
	if !opts.useCustomConfig {
		// Create readline with persistent history
		historyFile, _ := getHistoryFilePath()
		rl, err := readline.NewEx(&readline.Config{
			Prompt:            "(cucaracha) ",
			HistoryFile:       historyFile,
			HistorySearchFold: true,
			InterruptPrompt:   "^C",
			EOFPrompt:         "exit",
		})
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

	// Get history file path for custom config as well
	historyFile, _ := getHistoryFilePath()

	rl, err := readline.NewEx(&readline.Config{
		Prompt:            "(cucaracha) ",
		Stdin:             readCloser,
		Stdout:            opts.writer,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
		HistoryFile:       historyFile,
	})
	if err != nil {
		panic(err)
	}
	return rl
}

// getAliasesConfigPath returns the path to the aliases configuration file
func getAliasesConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	cucarachaDir := filepath.Join(homeDir, ".cucaracha")
	aliasesFile := filepath.Join(cucarachaDir, "aliases.yaml")

	// Create the .cucaracha directory if it doesn't exist
	if err := os.MkdirAll(cucarachaDir, 0755); err != nil {
		return "", err
	}

	return aliasesFile, nil
}

// newREPLInternal creates a new REPL instance with the provided configuration
func newREPLInternal(debugger debuggerUI.Debugger, readline *readline.Instance, writer io.Writer,
	loadArgs *debuggerUI.LoadArgs, quiet bool, outputFormat OutputFormat) *REPL {
	repl := &REPL{
		debugger:     debuggerUI.MakeCommandBased(debugger),
		readline:     readline,
		writer:       writer,
		commands:     make(map[string]CommandHandler),
		aliases:      make(map[string]*Alias),
		loadArgs:     loadArgs,
		quiet:        quiet,
		outputFormat: outputFormat,
		settings:     NewSettings(),
	}

	repl.registerCommands()

	// Create UI sink with printLogEntry callback
	// This is created early so that setting callbacks can use it before Run() is called
	repl.uiSink = logging.NewUISink("ui-repl", slog.LevelDebug, 1000, repl.printLogEntry)

	// Register setting change callbacks now that uiSink is initialized
	// This must happen before settings are applied (which happens before Run())
	if err := repl.registerSettingCallbacks(); err != nil {
		// Log error but don't prevent REPL creation
		slog.Error("failed to register setting callbacks", "error", err)
	}

	return repl
}

// NewREPL creates a new REPL instance
func NewREPL(debugger debuggerUI.Debugger) *REPL {
	rl := createReadline(readlineOptions{useCustomConfig: false})
	return newREPLInternal(debugger, rl, os.Stdout, nil, false, HumanReadable)
}

// NewREPLWithLoadArgs creates a new REPL instance with load arguments
func NewREPLWithLoadArgs(debugger debuggerUI.Debugger, loadArgs *debuggerUI.LoadArgs) *REPL {
	rl := createReadline(readlineOptions{useCustomConfig: false})
	return newREPLInternal(debugger, rl, os.Stdout, loadArgs, false, HumanReadable)
}

// NewREPLWithIO creates a new REPL instance with custom input/output
func NewREPLWithIO(debugger debuggerUI.Debugger, reader io.Reader, writer io.Writer) *REPL {
	rl := createReadline(readlineOptions{
		reader:          reader,
		writer:          writer,
		useCustomConfig: true,
	})
	return newREPLInternal(debugger, rl, writer, nil, false, HumanReadable)
}

// NewREPLWithIOQuiet creates a new REPL instance with custom input/output in quiet mode (no welcome/goodbye messages)
func NewREPLWithIOQuiet(debugger debuggerUI.Debugger, reader io.Reader, writer io.Writer) *REPL {
	rl := createReadline(readlineOptions{
		reader:          reader,
		writer:          writer,
		useCustomConfig: true,
	})
	return newREPLInternal(debugger, rl, writer, nil, true, HumanReadable)
}

// NewREPLWithOutputFormat creates a new REPL instance with a specific output format
func NewREPLWithOutputFormat(debugger debuggerUI.Debugger, reader io.Reader, writer io.Writer, format OutputFormat) *REPL {
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
	if err := r.settings.LoadFromFile(filePath); err != nil {
		return err
	}

	// Store the settings file path so we can save aliases back to it
	r.settingsFilePath = filePath

	// Also load aliases from the same settings file
	return r.loadAliasesFromSettingsFile(filePath)
}

// ApplySettingsKeyValue applies a single key=value setting string
func (r *REPL) ApplySettingsKeyValue(kvStr string) error {
	return r.settings.ApplyKeyValue(kvStr)
}

// loadAliasesFromSettingsFile loads aliases from the YAML settings file
func (r *REPL) loadAliasesFromSettingsFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		// If file doesn't exist, that's okay - just return nil
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read settings file: %w", err)
	}

	// Parse YAML with aliases key
	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		// If parsing fails, just return nil - settings were loaded, ignoring aliases
		return nil
	}

	aliasesData, ok := config["aliases"]
	if !ok {
		// No aliases section in file
		return nil
	}

	// Convert to map of alias definitions
	aliasesMap, ok := aliasesData.(map[string]interface{})
	if !ok {
		slog.Warn("aliases in settings file must be a map")
		return nil
	}

	// Load each alias
	for name, aliasData := range aliasesMap {
		alias, err := parseAliasFromYAML(name, aliasData)
		if err != nil {
			slog.Warn("failed to parse alias from settings", "name", name, "error", err)
			continue
		}
		r.aliases[strings.ToLower(name)] = alias
	}

	return nil
}

// saveAliasesToSettingsFile saves all current aliases back to the settings file
// maintaining the structured format with 'settings' and 'aliases' sections
func (r *REPL) saveAliasesToSettingsFile(settingsFilePath string) error {
	// Read existing config to preserve settings
	config := make(map[string]interface{})
	existingSettings := make(map[string]interface{})

	if _, err := os.Stat(settingsFilePath); err == nil {
		data, err := os.ReadFile(settingsFilePath)
		if err == nil {
			_ = yaml.Unmarshal(data, &config)

			// Extract existing settings section if it exists
			if settingsData, ok := config["settings"]; ok {
				if settingsMap, ok := settingsData.(map[string]interface{}); ok {
					existingSettings = settingsMap
				}
			} else {
				// Old flat format - everything is settings
				for key, value := range config {
					if key != "aliases" {
						existingSettings[key] = value
					}
				}
			}
		}
	}

	// Build new config with structured sections
	newConfig := make(map[string]interface{})

	// Add settings section
	if len(existingSettings) > 0 {
		newConfig["settings"] = existingSettings
	}

	// Build aliases map
	aliasesMap := make(map[string]interface{})
	for name, alias := range r.aliases {
		aliasEntry := make(map[string]interface{})

		// Convert commands to strings
		var cmdStrs []string
		for _, cmdParts := range alias.Commands {
			cmdStrs = append(cmdStrs, strings.Join(cmdParts, " "))
		}
		aliasEntry["commands"] = cmdStrs

		if alias.Doc != "" {
			aliasEntry["doc"] = alias.Doc
		}

		aliasesMap[name] = aliasEntry
	}

	// Add aliases section if there are any aliases
	if len(aliasesMap) > 0 {
		newConfig["aliases"] = aliasesMap
	}

	// Marshal to YAML
	data, err := yaml.Marshal(newConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal to YAML: %w", err)
	}

	// Write to file
	if err := os.WriteFile(settingsFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

// saveSettingsToFile saves all current settings back to the settings file
// maintaining the structured format with 'settings' and 'aliases' sections
func (r *REPL) saveSettingsToFile(settingsFilePath string) error {
	// Read existing config to preserve aliases
	config := make(map[string]interface{})
	existingAliases := make(map[string]interface{})

	if _, err := os.Stat(settingsFilePath); err == nil {
		data, err := os.ReadFile(settingsFilePath)
		if err == nil {
			_ = yaml.Unmarshal(data, &config)

			// Extract existing aliases section if it exists
			if aliasesData, ok := config["aliases"]; ok {
				if aliasesMap, ok := aliasesData.(map[string]interface{}); ok {
					existingAliases = aliasesMap
				}
			}
		}
	}

	// Build new config with structured sections
	newConfig := make(map[string]interface{})

	// Add settings section - export current settings
	settingsMap := r.settings.ExportSettings()
	if len(settingsMap) > 0 {
		newConfig["settings"] = settingsMap
	}

	// Add aliases section if there are any
	if len(existingAliases) > 0 {
		newConfig["aliases"] = existingAliases
	}

	// Marshal to YAML
	data, err := yaml.Marshal(newConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal to YAML: %w", err)
	}

	// Write to file
	if err := os.WriteFile(settingsFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

// parseAliasFromYAML parses an alias from YAML data
func parseAliasFromYAML(name string, aliasData interface{}) (*Alias, error) {
	alertMap, ok := aliasData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("alias data must be a map")
	}

	// Get commands
	commandsData, ok := alertMap["commands"]
	if !ok {
		return nil, fmt.Errorf("missing 'commands' field")
	}

	var commands [][]string

	// Handle commands as array or string
	switch v := commandsData.(type) {
	case []interface{}:
		for i, cmd := range v {
			cmdStr, ok := cmd.(string)
			if !ok {
				return nil, fmt.Errorf("command at index %d is not a string", i)
			}
			commands = append(commands, strings.Fields(cmdStr))
		}
	case string:
		commands = append(commands, strings.Fields(v))
	default:
		return nil, fmt.Errorf("commands must be a string or array of strings")
	}

	alias := &Alias{
		Commands: commands,
	}

	// Get optional doc
	if docData, ok := alertMap["doc"]; ok {
		if docStr, ok := docData.(string); ok {
			alias.Doc = docStr
		}
	}

	return alias, nil
}

// registerCommands registers all available REPL commands
func (r *REPL) registerCommands() {
	r.commands["help"] = r.handleHelp
	r.commands["h"] = r.handleHelp
	r.commands["exit"] = r.handleExit
	r.commands["quit"] = r.handleExit
	r.commands["q"] = r.handleExit

	// Settings commands
	r.commands["set"] = r.handleSet
	r.commands["get"] = r.handleGet
	r.commands["save-settings"] = r.handleSaveSettings

	// Alias commands
	r.commands["define"] = r.handleDefine
	r.commands["undefine"] = r.handleUndefine
	r.commands["unalias"] = r.handleUndefine
	r.commands["save-aliases"] = r.handleSaveAliases

	// Utility commands
	r.commands["loggers"] = r.handleLoggers
}

// Run starts the REPL interactive loop
func (r *REPL) Run() error {
	defer r.readline.Close()

	// Initialize debugger with load arguments if provided
	if r.loadArgs != nil {
		if err := r.initializeDebugger(); err != nil {
			return r.printError(fmt.Errorf("Failed to initialize debugger: %v", err))
		}
	}

	r.debugger.SetEventCallback(func(event *debuggerUI.DebuggerEvent) {
		r.handleDebuggerEvent(event)
	})

	if !r.quiet {
		r.printWelcome()
	}

	for !r.exit {
		r.waitingForInput = true

		// Use different prompt for define mode
		if r.definingAlias {
			r.readline.SetPrompt(fmt.Sprintf("define %s> ", r.defineAliasName))
		}

		line, err := r.readline.Readline()
		r.waitingForInput = false

		if err != nil {
			if err == readline.ErrInterrupt || err.Error() == "Interrupt" {
				// Cancel define mode on interrupt
				if r.definingAlias {
					r.write("Define mode cancelled\n")
					r.definingAlias = false
					r.defineAliasName = ""
					r.defineCommands = nil
					r.defineDoc = ""
					r.readline.SetPrompt("(cucaracha) ")
				}
				continue
			}
			if err.Error() == "EOF" {
				break
			}
			return err
		}

		line = strings.TrimSpace(line)

		// Handle define mode
		if r.definingAlias {
			// Check for end markers
			if line == "end" || line == "end-alias" {
				// Create the alias
				r.aliases[r.defineAliasName] = &Alias{
					Commands: r.defineCommands,
					Doc:      r.defineDoc,
				}

				outputMsg := fmt.Sprintf("Defined alias '%s' with %d command(s)", r.defineAliasName, len(r.defineCommands))
				if r.defineDoc != "" {
					outputMsg += fmt.Sprintf("\n  Documentation: %s", r.defineDoc)
				}
				r.write("%s\n", outputMsg)

				// Exit define mode and reset prompt
				r.definingAlias = false
				r.defineAliasName = ""
				r.defineCommands = nil
				r.defineDoc = ""
				r.readline.SetPrompt("(cucaracha) ")
			} else if line != "" {
				// Validate and resolve the command
				parts := strings.Fields(line)
				_, resolution, err := r.validateAndResolveCommand(parts)
				if err != nil {
					r.write("  Warning: %v\n", err)
					continue
				}

				// Show the command being added with resolution info
				r.write("  ✓ %s\n", strings.Join(parts, " "))
				r.write("    %s\n", resolution)

				// Add command to the sequence
				r.defineCommands = append(r.defineCommands, parts)
			}
			continue
		}

		if line == "" {
			// If empty line and we have a previous command, execute the last command
			if r.lastInput != "" {
				line = r.lastInput
			} else {
				continue
			}
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
		r.loadArgs.Runtime = utils.Ptr(debuggerUI.RuntimeTypeInterpreter)
	}

	// Check if there's anything to load
	if r.loadArgs.FullDescriptorPath == nil &&
		r.loadArgs.ProgramPath == nil {
		return nil
	}

	cmd := &debuggerUI.DebuggerCommand{
		Command:  debuggerUI.DebuggerCommandLoad,
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

// expandAliasRecursively expands an alias recursively, detecting cycles
// For multi-command aliases, returns the first command if it's part of an alias sequence
func (r *REPL) expandAliasRecursively(cmd string, visited map[string]bool, depth int) ([]string, error) {
	const maxDepth = 10 // Prevent deep recursion

	if depth > maxDepth {
		return nil, fmt.Errorf("alias expansion depth exceeded (circular alias detected)")
	}

	// Check for cycles
	if visited[cmd] {
		return nil, fmt.Errorf("circular alias detected: %s", cmd)
	}

	// If not an alias, return just the command
	alias, isAlias := r.aliases[cmd]
	if !isAlias {
		return []string{cmd}, nil
	}

	// Mark as visited
	visited[cmd] = true

	// For multi-command aliases, expand the first command if it's also an alias
	if len(alias.Commands) == 0 {
		return nil, fmt.Errorf("alias %s has no commands", cmd)
	}

	firstCmdParts := alias.Commands[0]
	if len(firstCmdParts) == 0 {
		return nil, fmt.Errorf("alias %s has empty first command", cmd)
	}

	// Try to expand the first part of the first command
	firstCmd := firstCmdParts[0]
	expandedFirst, err := r.expandAliasRecursively(firstCmd, visited, depth+1)
	if err != nil {
		return nil, err
	}

	// Combine expanded first part with remaining parts of first command
	result := append(expandedFirst, firstCmdParts[1:]...)
	return result, nil
}

// validateAndResolveCommand validates a command using REPL syntax and resolves any alias references
// Returns the resolved command parts and a description of what it resolves to
func (r *REPL) validateAndResolveCommand(cmdParts []string) ([]string, string, error) {
	if len(cmdParts) == 0 {
		return nil, "", fmt.Errorf("empty command")
	}

	firstCmd := cmdParts[0]

	// Check if it's a built-in command
	if _, isBuiltin := r.commands[firstCmd]; isBuiltin {
		// Built-in command, just return it as-is
		return cmdParts, fmt.Sprintf("built-in command: %s", firstCmd), nil
	}

	// Check if it's an alias
	if alias, isAlias := r.aliases[firstCmd]; isAlias {
		// Resolve the alias
		if len(alias.Commands) == 0 {
			return nil, "", fmt.Errorf("alias '%s' has no commands", firstCmd)
		}

		// Try to expand the alias recursively
		expandedParts, err := r.expandAliasRecursively(firstCmd, make(map[string]bool), 0)
		if err != nil {
			return nil, "", err
		}

		// Combine expanded parts with any user-provided arguments
		resolvedCmd := append(expandedParts, cmdParts[1:]...)

		// Show the alias expansion
		aliasDescription := fmt.Sprintf("alias '%s' → resolved to: %s", firstCmd, strings.Join(expandedParts, " "))
		if alias.Doc != "" {
			aliasDescription = fmt.Sprintf("alias '%s' (%s) → resolved to: %s", firstCmd, alias.Doc, strings.Join(expandedParts, " "))
		}

		return resolvedCmd, aliasDescription, nil
	}

	// Check if it's a debugger command by trying to parse it
	var syntax REPLSyntax
	_, err := syntax.ParseCommand(cmdParts)
	if err == nil {
		// Valid debugger command
		return cmdParts, fmt.Sprintf("debugger command: %s", firstCmd), nil
	}

	// Not a built-in command, alias, or debugger command
	return nil, "", fmt.Errorf("unknown command: '%s' (not a built-in command, defined alias, or debugger command)", firstCmd)
}

// processCommand parses and executes a command
func (r *REPL) processCommand(input string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}

	cmd := parts[0]
	args := parts[1:]

	// Check if this is an alias
	if alias, isAlias := r.aliases[cmd]; isAlias {
		// Execute multi-command alias sequence
		r.startCommandOutput()
		for _, cmdParts := range alias.Commands {
			if len(cmdParts) == 0 {
				continue
			}

			// Expand each command part in case it's also an alias
			expandedParts, err := r.expandAliasRecursively(cmdParts[0], make(map[string]bool), 0)
			if err != nil {
				r.printError(err)
				r.finishCommandOutput(false, err)
				return
			}

			// Combine expanded parts with rest of command
			fullParts := append(expandedParts, cmdParts[1:]...)
			// Append original user args to each command
			fullParts = append(fullParts, args...)

			if len(fullParts) == 0 {
				continue
			}

			// Execute the command
			finalCmd := fullParts[0]
			finalArgs := fullParts[1:]

			handler, exists := r.commands[finalCmd]
			if !exists {
				handler = r.handleDebuggerCommand
				finalArgs = fullParts
			}

			err = handler(finalArgs)
			if err != nil {
				r.printError(err)
				r.finishCommandOutput(false, err)
				return
			}
		}
		r.finishCommandOutput(true, nil)
		return
	}

	// Not an alias - execute as single command
	// Recursively expand in case it's a command that aliases to other aliases
	expandedParts, err := r.expandAliasRecursively(cmd, make(map[string]bool), 0)
	if err != nil {
		r.startCommandOutput()
		r.printError(err)
		r.finishCommandOutput(false, err)
		return
	}

	// expandedParts now contains the full expanded command
	// Append original args to the fully expanded command
	parts = append(expandedParts, args...)
	cmd = strings.ToLower(parts[0])
	args = parts[1:]

	handler, exists := r.commands[cmd]
	if !exists {
		handler = r.handleDebuggerCommand
		args = parts // Pass the entire input as args for the debugger command parser
	}

	r.startCommandOutput()
	err = handler(args)
	if err != nil {
		r.printError(err)
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
			return fmt.Errorf("Failed to initialize debugger: %v", err)
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

// applyUILoggersImpl applies the UI sink to specified logger names without updating settings.
// This is used by both the callback and the applyUILoggers method.
func (r *REPL) applyUILoggersImpl(loggerNames []string) error {
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

	return nil
}

// registerSettingCallbacks registers callbacks for setting changes.
// These callbacks are triggered whenever a setting value changes, either from
// file loading or manual user input through the REPL.
func (r *REPL) registerSettingCallbacks() error {
	// Register callback for display.logs setting
	if err := r.settings.SetChangeCallback(SettingKeyDisplayLogs, func(name string, newValue interface{}) error {
		loggerNames, ok := newValue.([]string)
		if !ok {
			return fmt.Errorf("invalid value type for %s: expected []string", name)
		}
		return r.applyUILoggersImpl(loggerNames)
	}); err != nil {
		return err
	}

	return nil
}
