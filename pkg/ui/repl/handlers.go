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

func (r *REPL) handleCommand(args []string, parse func([]string) (*debuggerUI.DebuggerCommand, error), resultError func(*debuggerUI.DebuggerCommandResult) error) error {
	cmd, err := parse(args)
	if err != nil {
		return err
	}

	result, err := r.debugger.Execute(cmd)
	if err != nil {
		return err
	}
	if err := resultError(result); err != nil {
		return err
	}
	r.printCommandResult(result)
	return nil
}

func (r *REPL) handleStep(args []string) error {
	return r.handleCommand(args, func(args []string) (*debuggerUI.DebuggerCommand, error) {
		stepArgs := &debuggerUI.StepArgs{
			StepMode:  debuggerUI.StepModeInto,
			CountMode: debuggerUI.StepCountSourceLines,
		}

		return &debuggerUI.DebuggerCommand{
			Command:  debuggerUI.DebuggerCommandStep,
			StepArgs: stepArgs,
		}, nil
	}, func(result *debuggerUI.DebuggerCommandResult) error {
		return result.StepResult.Error
	})
}

func (r *REPL) handleContinue(args []string) error {
	return r.handleCommand(args, func(args []string) (*debuggerUI.DebuggerCommand, error) {
		return &debuggerUI.DebuggerCommand{
			Command: debuggerUI.DebuggerCommandContinue,
		}, nil
	}, func(result *debuggerUI.DebuggerCommandResult) error {
		return result.ContinueResult.Error
	})
}

func (r *REPL) handleInterrupt(args []string) error {
	return r.handleCommand(args, func(args []string) (*debuggerUI.DebuggerCommand, error) {
		return &debuggerUI.DebuggerCommand{
			Command: debuggerUI.DebuggerCommandInterrupt,
		}, nil
	}, func(result *debuggerUI.DebuggerCommandResult) error {
		return result.InterruptResult.Error
	})
}

func (r *REPL) handleRun(args []string) error {
	return r.handleCommand(args, func(args []string) (*debuggerUI.DebuggerCommand, error) {
		return &debuggerUI.DebuggerCommand{
			Command: debuggerUI.DebuggerCommandContinue,
		}, nil
	}, func(result *debuggerUI.DebuggerCommandResult) error {
		return result.ContinueResult.Error
	})
}

func (r *REPL) handleReset(args []string) error {
	return r.handleCommand(args, func(args []string) (*debuggerUI.DebuggerCommand, error) {
		return &debuggerUI.DebuggerCommand{
			Command: debuggerUI.DebuggerCommandReset,
		}, nil
	}, func(result *debuggerUI.DebuggerCommandResult) error {
		return result.ResetResult.Error
	})
}

func (r *REPL) handleRestart(args []string) error {
	return r.handleCommand(args, func(args []string) (*debuggerUI.DebuggerCommand, error) {
		return &debuggerUI.DebuggerCommand{
			Command: debuggerUI.DebuggerCommandRestart,
		}, nil
	}, func(result *debuggerUI.DebuggerCommandResult) error {
		return result.RestartResult.Error
	})
}

// ============================================================================
// Breakpoint Commands
// ============================================================================

func (r *REPL) handleBreak(args []string) error {
	return r.handleCommand(args, func(args []string) (*debuggerUI.DebuggerCommand, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("break requires an address")
		}

		addr, err := parseAddress(args[0])
		if err != nil {
			return nil, err
		}

		return &debuggerUI.DebuggerCommand{
			Command: debuggerUI.DebuggerCommandBreak,
			BreakArgs: &debuggerUI.BreakArgs{
				Address: &addr,
			},
		}, nil
	}, func(result *debuggerUI.DebuggerCommandResult) error {
		return result.BreakResult.Error
	})
}

func (r *REPL) handleRemoveBreakpoint(args []string) error {
	return r.handleCommand(args, func(args []string) (*debuggerUI.DebuggerCommand, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("removebreakpoint requires a breakpoint ID")
		}

		id, err := strconv.Atoi(args[0])
		if err != nil {
			return nil, fmt.Errorf("invalid breakpoint ID: %v", err)
		}

		return &debuggerUI.DebuggerCommand{
			Command: debuggerUI.DebuggerCommandRemoveBreakpoint,
			RemoveBreakpointArgs: &debuggerUI.RemoveBreakpointArgs{
				ID: id,
			},
		}, nil
	}, func(result *debuggerUI.DebuggerCommandResult) error {
		return result.RemoveBreakpointResult.Error
	})
}

func (r *REPL) handleWatch(args []string) error {
	return r.handleCommand(args, func(args []string) (*debuggerUI.DebuggerCommand, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("watch requires an address")
		}

		addr, err := parseAddress(args[0])
		if err != nil {
			return nil, err
		}

		// Default to 4-byte watch
		memRegion := &debuggerUI.MemoryRegion{
			Start: addr,
			Size:  4,
		}

		return &debuggerUI.DebuggerCommand{
			Command: debuggerUI.DebuggerCommandWatch,
			WatchArgs: &debuggerUI.WatchArgs{
				Range: memRegion,
			},
		}, nil
	}, func(result *debuggerUI.DebuggerCommandResult) error {
		return result.WatchResult.Error
	})
}

func (r *REPL) handleRemoveWatchpoint(args []string) error {
	return r.handleCommand(args, func(args []string) (*debuggerUI.DebuggerCommand, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("removewatchpoint requires a watchpoint ID")
		}

		id, err := strconv.Atoi(args[0])
		if err != nil {
			return nil, fmt.Errorf("invalid watchpoint ID: %v", err)
		}

		return &debuggerUI.DebuggerCommand{
			Command: debuggerUI.DebuggerCommandRemoveWatchpoint,
			RemoveWatchpointArgs: &debuggerUI.RemoveWatchpointArgs{
				ID: id,
			},
		}, nil
	}, func(result *debuggerUI.DebuggerCommandResult) error {
		return result.RemoveWatchpointResult.Error
	})
}

func (r *REPL) handleList(args []string) error {
	return r.handleCommand(args, func(args []string) (*debuggerUI.DebuggerCommand, error) {
		return &debuggerUI.DebuggerCommand{
			Command: debuggerUI.DebuggerCommandList,
		}, nil
	}, func(result *debuggerUI.DebuggerCommandResult) error {
		return result.ListResult.Error
	})
}

// ============================================================================
// Inspection Commands
// ============================================================================

func (r *REPL) handleDisasm(args []string) error {
	return r.handleCommand(args, func(args []string) (*debuggerUI.DebuggerCommand, error) {
		var disasmArgs debuggerUI.DisasmArgs
		// Set defaults to show everything
		disasmArgs.ShowSource = true
		disasmArgs.ShowCFG = true

		// Parse positional and flag arguments
		var positionalCount int
		for _, arg := range args {
			if arg == "-source" || arg == "-s" {
				disasmArgs.ShowSource = true
			} else if arg == "-no-source" || arg == "-ns" {
				disasmArgs.ShowSource = false
			} else if arg == "-cfg" || arg == "-g" {
				disasmArgs.ShowCFG = true
			} else if arg == "-no-cfg" || arg == "-ng" {
				disasmArgs.ShowCFG = false
			} else if positionalCount == 0 {
				// First positional arg: address
				disasmArgs.AddressExpr = arg
				positionalCount++
			} else if positionalCount == 1 {
				// Second positional arg: count
				count, err := strconv.Atoi(arg)
				if err != nil {
					return nil, fmt.Errorf("invalid count: %v", err)
				}
				disasmArgs.Count = count
				positionalCount++
			}
		}

		// Store the args for use in output formatting
		r.lastDisasmArgs = &disasmArgs

		return &debuggerUI.DebuggerCommand{
			Command:    debuggerUI.DebuggerCommandDisassemble,
			DisasmArgs: &disasmArgs,
		}, nil
	}, func(result *debuggerUI.DebuggerCommandResult) error {
		return result.DisassemblyResult.Error
	})
}

func (r *REPL) handleCurrent(args []string) error {
	return r.handleCommand(args, func(args []string) (*debuggerUI.DebuggerCommand, error) {
		return &debuggerUI.DebuggerCommand{
			Command: debuggerUI.DebuggerCommandCurrentInstruction,
		}, nil
	}, func(result *debuggerUI.DebuggerCommandResult) error {
		return result.CurrentInstructionResult.Error
	})
}

func (r *REPL) handleMemory(args []string) error {
	return r.handleCommand(args, func(args []string) (*debuggerUI.DebuggerCommand, error) {
		memArgs := &debuggerUI.MemoryArgs{}

		if len(args) > 0 {
			memArgs.AddressExpr = args[0]
		}

		if len(args) > 1 {
			count, err := strconv.Atoi(args[1])
			if err != nil {
				return nil, fmt.Errorf("invalid count: %v", err)
			}
			memArgs.Count = count
		}

		return &debuggerUI.DebuggerCommand{
			Command:    debuggerUI.DebuggerCommandMemory,
			MemoryArgs: memArgs,
		}, nil
	}, func(result *debuggerUI.DebuggerCommandResult) error {
		return result.MemoryResult.Error
	})
}

func (r *REPL) handleSource(args []string) error {
	return r.handleCommand(args, func(args []string) (*debuggerUI.DebuggerCommand, error) {
		sourceArgs := &debuggerUI.SourceArgs{}

		if len(args) > 0 {
			sourceArgs.File = args[0]
		}

		return &debuggerUI.DebuggerCommand{
			Command:    debuggerUI.DebuggerCommandSource,
			SourceArgs: sourceArgs,
		}, nil
	}, func(result *debuggerUI.DebuggerCommandResult) error {
		return result.SourceResult.Error
	})
}

func (r *REPL) handleInfo(args []string) error {
	infoType := debuggerUI.InfoTypeGeneral
	if len(args) > 0 {
		var err error
		infoType, err = debuggerUI.InfoTypeFromString(args[0])
		if err != nil {
			return err
		}
	}

	return r.handleCommand(args, func(args []string) (*debuggerUI.DebuggerCommand, error) {
		return &debuggerUI.DebuggerCommand{
			Command: debuggerUI.DebuggerCommandInfo,
			InfoArgs: &debuggerUI.InfoArgs{
				Type: infoType,
			},
		}, nil
	}, func(result *debuggerUI.DebuggerCommandResult) error {
		return result.InfoResult.Error
	})
}

func (r *REPL) handleSymbols(args []string) error {
	return r.handleCommand(args, func(args []string) (*debuggerUI.DebuggerCommand, error) {
		symbolsArgs := &debuggerUI.SymbolsArgs{}

		if len(args) > 0 {
			symbolsArgs.SymbolName = &args[0]
		}

		return &debuggerUI.DebuggerCommand{
			Command:     debuggerUI.DebuggerCommandSymbols,
			SymbolsArgs: symbolsArgs,
		}, nil
	}, func(result *debuggerUI.DebuggerCommandResult) error {
		return result.SymbolsResult.Error
	})
}

func (r *REPL) handleRegisters(args []string) error {
	return r.handleCommand(args, func(args []string) (*debuggerUI.DebuggerCommand, error) {
		return &debuggerUI.DebuggerCommand{
			Command: debuggerUI.DebuggerCommandRegisters,
		}, nil
	}, func(result *debuggerUI.DebuggerCommandResult) error {
		return result.RegistersResult.Error
	})
}

func (r *REPL) handleStack(args []string) error {
	return r.handleCommand(args, func(args []string) (*debuggerUI.DebuggerCommand, error) {
		return &debuggerUI.DebuggerCommand{
			Command: debuggerUI.DebuggerCommandStack,
		}, nil
	}, func(result *debuggerUI.DebuggerCommandResult) error {
		return result.StackResult.Error
	})
}

func (r *REPL) handleVars(args []string) error {
	return r.handleCommand(args, func(args []string) (*debuggerUI.DebuggerCommand, error) {
		return &debuggerUI.DebuggerCommand{
			Command: debuggerUI.DebuggerCommandVariables,
		}, nil
	}, func(result *debuggerUI.DebuggerCommandResult) error {
		return result.VariablesResult.Error
	})
}

func (r *REPL) handleEval(args []string) error {
	return r.handleCommand(args, func(args []string) (*debuggerUI.DebuggerCommand, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("eval requires an expression")
		}

		expression := strings.Join(args, " ")
		return &debuggerUI.DebuggerCommand{
			Command: debuggerUI.DebuggerCommandEvaluateExpression,
			EvalArgs: &debuggerUI.EvalArgs{
				Expression: expression,
			},
		}, nil
	}, func(result *debuggerUI.DebuggerCommandResult) error {
		return result.EvalResult.Error
	})
}

// ============================================================================
// Program Loading Commands
// ============================================================================

func (r *REPL) handleLoad(args []string) error {
	return r.handleCommand(args, func(args []string) (*debuggerUI.DebuggerCommand, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("load requires a file path")
		}

		// Read the auto_build_clang setting
		autoBuildClang, err := r.settings.GetBool(SettingKeyBuildAutoClang)
		if err != nil {
			autoBuildClang = true // Default to true if setting not found
		}

		// Read the force_rebuild_clang setting
		forceRebuildClang, err := r.settings.GetBool(SettingKeyBuildForceClang)
		if err != nil {
			forceRebuildClang = false // Default to false if setting not found
		}

		return &debuggerUI.DebuggerCommand{
			Command: debuggerUI.DebuggerCommandLoadProgramFromFile,
			LoadProgramArgs: &debuggerUI.LoadProgramArgs{
				FilePath:          args[0],
				AutoBuildClang:    &autoBuildClang,
				ForceRebuildClang: &forceRebuildClang,
			},
		}, nil
	}, func(result *debuggerUI.DebuggerCommandResult) error {
		return result.LoadProgramResult.Error
	})
}

func (r *REPL) handleLoadProgram(args []string) error {
	return r.handleCommand(args, func(args []string) (*debuggerUI.DebuggerCommand, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("loadprogram requires a file path")
		}

		// Read the auto_build_clang setting
		autoBuildClang, err := r.settings.GetBool(SettingKeyBuildAutoClang)
		if err != nil {
			autoBuildClang = true // Default to true if setting not found
		}

		// Read the force_rebuild_clang setting
		forceRebuildClang, err := r.settings.GetBool(SettingKeyBuildForceClang)
		if err != nil {
			forceRebuildClang = false // Default to false if setting not found
		}

		return &debuggerUI.DebuggerCommand{
			Command: debuggerUI.DebuggerCommandLoadProgramFromFile,
			LoadProgramArgs: &debuggerUI.LoadProgramArgs{
				FilePath:          args[0],
				AutoBuildClang:    &autoBuildClang,
				ForceRebuildClang: &forceRebuildClang,
			},
		}, nil
	}, func(result *debuggerUI.DebuggerCommandResult) error {
		return result.LoadProgramResult.Error
	})
}

func (r *REPL) handleLoadSystem(args []string) error {
	return r.handleCommand(args, func(args []string) (*debuggerUI.DebuggerCommand, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("loadsystem requires a file path")
		}

		return &debuggerUI.DebuggerCommand{
			Command: debuggerUI.DebuggerCommandLoadSystemFromFile,
			LoadSystemArgs: &debuggerUI.LoadSystemArgs{
				FilePath: args[0],
			},
		}, nil
	}, func(result *debuggerUI.DebuggerCommandResult) error {
		return result.LoadSystemResult.Error
	})
}

func (r *REPL) handleLoadRuntime(args []string) error {
	return r.handleCommand(args, func(args []string) (*debuggerUI.DebuggerCommand, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("loadruntime requires a runtime name")
		}

		runtimeType, err := debuggerUI.RuntimeTypeFromString(args[0])
		if err != nil {
			return nil, err
		}

		return &debuggerUI.DebuggerCommand{
			Command: debuggerUI.DebuggerCommandLoadRuntime,
			LoadRuntimeArgs: &debuggerUI.LoadRuntimeArgs{
				Runtime: runtimeType,
			},
		}, nil
	}, func(result *debuggerUI.DebuggerCommandResult) error {
		return result.LoadRuntimeResult.Error
	})
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
