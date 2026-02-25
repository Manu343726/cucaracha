package repl

import (
	"fmt"
	"strconv"
	"strings"

	debuggerUI "github.com/Manu343726/cucaracha/pkg/ui/debugger"
	"github.com/Manu343726/cucaracha/pkg/utils/logging"
)

// ============================================================================
// Execution Commands
// ============================================================================

func (r *REPL) handleDebuggerCommand(args []string) error {
	var syntax REPLSyntax

	command, err := syntax.ParseCommand(args)
	if err != nil {
		return err
	}

	result, err := r.debugger.Execute(command)
	if err != nil {
		return err
	}

	return r.printCommandResult(result)
}

// ============================================================================
// Info Commands
// ============================================================================

func (r *REPL) handleHelp(args []string) error {
	r.printHelp()
	return nil
}

func (r *REPL) handleExit(args []string) error {
	r.exit = true
	return nil
}

// ============================================================================
// Settings Commands
// ============================================================================

func (r *REPL) handleSet(args []string) error {
	if len(args) == 0 {
		// Display all available settings with their descriptions
		r.printAllSettings()
		return nil
	}

	if len(args) < 1 {
		return fmt.Errorf("set requires a setting name")
	}

	settingName := args[0]

	// For all settings, require exactly one value (or multiple for logging.show which is a slice)
	if len(args) < 2 {
		return fmt.Errorf("set requires a setting name and value")
	}

	// For display.logs, collect all remaining arguments as logger names
	if settingName == SettingKeyDisplayLogs {
		// Pass remaining args as a list for the logging setting
		loggerNames := args[1:]
		if err := r.settings.Set(settingName, loggerNames); err != nil {
			return err
		}
		value, _ := r.settings.Get(settingName)
		r.write("Set %s = %v\n", settingName, value)
		return nil
	}

	// For other settings, only use the first value argument
	settingValue := args[1]
	if err := r.settings.Set(settingName, settingValue); err != nil {
		return err
	}

	// Print confirmation
	value, _ := r.settings.Get(settingName)
	r.write("Set %s = %v\n", settingName, value)
	return nil
}

func (r *REPL) handleGet(args []string) error {
	if len(args) == 0 {
		// Display all settings with their current values
		r.printCurrentSettings()
		return nil
	}

	settingName := args[0]
	value, err := r.settings.Get(settingName)
	if err != nil {
		return err
	}

	r.write("%s = %v\n", settingName, value)
	return nil
}

// ============================================================================
// Alias Commands
// ============================================================================

func (r *REPL) handleDefine(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("define requires an alias name (usage: define <name>\nenter commands, one per line\nend to finish)")
	}

	aliasName := strings.ToLower(args[0])

	// Check if trying to override a built-in command
	if _, isBuiltin := r.commands[aliasName]; isBuiltin {
		return fmt.Errorf("cannot define alias '%s': conflicts with built-in command", aliasName)
	}

	// Check if trying to override an existing alias
	if _, exists := r.aliases[aliasName]; exists {
		return fmt.Errorf("alias '%s' already defined; use 'undefine %s' first", aliasName, aliasName)
	}

	// Check for optional documentation in args (format: define name "documentation")
	var doc string
	if len(args) > 1 && strings.HasPrefix(args[1], "\"") && strings.HasSuffix(args[1], "\"") {
		doc = strings.TrimPrefix(strings.TrimSuffix(args[1], "\""), "\"")
	}

	// Enter define mode
	r.definingAlias = true
	r.defineAliasName = aliasName
	r.defineCommands = make([][]string, 0)
	r.defineDoc = doc

	if doc != "" {
		r.write("Defining alias '%s' (%s)\nEnter commands, one per line. Type 'end' to finish.\n", aliasName, doc)
	} else {
		r.write("Defining alias '%s'\nEnter commands, one per line. Type 'end' to finish.\n", aliasName)
	}

	return nil
}

func (r *REPL) handleUndefine(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("undefine requires an alias name (usage: undefine <name>)")
	}

	aliasName := strings.ToLower(args[0])

	if _, exists := r.aliases[aliasName]; !exists {
		return fmt.Errorf("alias '%s' not defined", aliasName)
	}

	delete(r.aliases, aliasName)
	r.write("Removed alias '%s'\n", aliasName)

	return nil
}

func (r *REPL) handleSaveAliases(args []string) error {
	var filePath string

	// If a file path is provided as argument, use it
	if len(args) > 0 {
		filePath = args[0]
	} else if r.settingsFilePath != "" {
		// Use the loaded settings file path
		filePath = r.settingsFilePath
	} else {
		return fmt.Errorf("no settings file loaded; specify a file path (usage: save-aliases [file])")
	}

	if err := r.saveAliasesToSettingsFile(filePath); err != nil {
		return fmt.Errorf("failed to save aliases: %w", err)
	}

	r.write("Saved %d alias(es) to %s\n", len(r.aliases), filePath)
	return nil
}

func (r *REPL) handleSaveSettings(args []string) error {
	var filePath string

	// If a file path is provided as argument, use it
	if len(args) > 0 {
		filePath = args[0]
	} else if r.settingsFilePath != "" {
		// Use the loaded settings file path
		filePath = r.settingsFilePath
	} else {
		return fmt.Errorf("no settings file loaded; specify a file path (usage: save-settings [file])")
	}

	if err := r.saveSettingsToFile(filePath); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	r.write("Saved settings to %s\n", filePath)
	return nil
}

// ============================================================================
// Event Handler
// ============================================================================

func (r *REPL) handleDebuggerEvent(event *debuggerUI.DebuggerEvent) {
	if event == nil {
		return
	}

	displayEvents, _ := r.settings.GetBool(SettingKeyDisplayEvents)
	if !displayEvents {
		return
	}

	// Convert debugger event type to string
	eventTypeStr := event.Type.String()

	// Print event header
	r.write("\n>>> [%s]\n", eventTypeStr)

	// Print event details based on type
	switch event.Type {
	case debuggerUI.DebuggerEventStepped:
		r.write("  Instruction stepped\n")
		r.printEventExecutionDetails(event.Result)

	case debuggerUI.DebuggerEventBreakpointHit:
		r.write("  Breakpoint hit\n")
		if event.Result != nil && event.Result.Breakpoint != nil {
			r.write("    Breakpoint ID: %d\n", event.Result.Breakpoint.ID)
			r.write("    Address: 0x%x\n", event.Result.Breakpoint.Address)
		}
		r.printEventExecutionDetails(event.Result)

	case debuggerUI.DebuggerEventWatchpointHit:
		r.write("  Watchpoint hit\n")
		if event.Result != nil && event.Result.Watchpoint != nil {
			r.write("    Watchpoint ID: %d\n", event.Result.Watchpoint.ID)
			r.write("    Address Range: 0x%x - 0x%x (size: %d)\n",
				event.Result.Watchpoint.Range.Start,
				event.Result.Watchpoint.Range.Start+event.Result.Watchpoint.Range.Size,
				event.Result.Watchpoint.Range.Size)
		}
		r.printEventExecutionDetails(event.Result)

	case debuggerUI.DebuggerEventProgramTerminated:
		r.write("  Program terminated\n")
		r.printEventExecutionDetails(event.Result)

	case debuggerUI.DebuggerEventProgramHalted:
		r.write("  Program halted\n")
		r.printEventExecutionDetails(event.Result)

	case debuggerUI.DebuggerEventError:
		r.write("  Error occurred\n")
		if event.Result != nil && event.Result.Error != nil {
			r.write("    Error: %v\n", event.Result.Error)
		}
		r.printEventExecutionDetails(event.Result)

	case debuggerUI.DebuggerEventSourceLocationChanged:
		r.write("  Source location changed\n")
		if event.Result != nil && event.Result.LastLocation != nil {
			r.write("    Location: %s:%d\n",
				event.Result.LastLocation.File,
				event.Result.LastLocation.Line)
		}
		r.printEventExecutionDetails(event.Result)

	case debuggerUI.DebuggerEventInterrupted:
		r.write("  Execution interrupted\n")
		r.printEventExecutionDetails(event.Result)

	case debuggerUI.DebuggerEventProgramLoaded:
		r.write("  Program loaded\n")

	case debuggerUI.DebuggerEventLagging:
		r.write("  Emulator lagging\n")
		if event.Result != nil && event.Result.LaggingCycles > 0 {
			r.write("    Lagging by: %d cycles\n", event.Result.LaggingCycles)
		}
		r.printEventExecutionDetails(event.Result)
	}
}

// printEventExecutionDetails prints common execution result details
func (r *REPL) printEventExecutionDetails(result *debuggerUI.ExecutionResult) {
	if result == nil {
		return
	}

	// Print execution statistics
	if result.Steps > 0 {
		r.write("    Steps: %d\n", result.Steps)
	}
	if result.Cycles > 0 {
		r.write("    Cycles: %d\n", result.Cycles)
	}

	// Print stop reason
	if result.StopReason != debuggerUI.StopReasonNone {
		r.write("    Stop Reason: %s\n", result.StopReason.String())
	}

	// Print last instruction
	if result.LastInstruction > 0 {
		r.write("    Last Instruction: 0x%x\n", result.LastInstruction)
	}

	// Print source location
	if result.LastLocation != nil {
		r.write("    Source Location: %s:%d\n",
			result.LastLocation.File,
			result.LastLocation.Line)
	}

	// Print lagging cycles if present
	if result.LaggingCycles > 0 {
		r.write("    Lagging Cycles: %d\n", result.LaggingCycles)
	}
}

// ============================================================================
// Utility Commands
// ============================================================================

func (r *REPL) handleLoggers(args []string) error {
	registry := logging.DefaultRegistry()
	loggerNames := registry.ListLoggers()

	if len(loggerNames) == 0 {
		r.write("No loggers registered\n")
		return nil
	}

	r.write("Registered loggers:\n")
	for _, name := range loggerNames {
		// Get the logger to check what sinks it has
		logger, err := registry.GetRegisteredLogger(name)
		if err == nil {
			sinks := logger.GetSinks()
			sinkNames := make([]string, len(sinks))
			for i, sink := range sinks {
				sinkNames[i] = sink.Name()
			}
			r.write("  %s [sinks: %s]\n", name, strings.Join(sinkNames, ", "))
		} else {
			r.write("  %s\n", name)
		}
	}

	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

func parseAddress(addrStr string) (uint32, error) {
	// Handle hex format (0x...)
	if strings.HasPrefix(addrStr, "0x") || strings.HasPrefix(addrStr, "0X") {
		var addr uint32
		_, err := fmt.Sscanf(addrStr, "0x%x", &addr)
		return addr, err
	}

	// Try as decimal
	addr, err := strconv.ParseUint(addrStr, 10, 32)
	return uint32(addr), err
}
