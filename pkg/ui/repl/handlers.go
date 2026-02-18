package repl

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/ui"
	"github.com/Manu343726/cucaracha/pkg/utils/logging"
)

// ============================================================================
// Execution Commands
// ============================================================================

func (r *REPL) handleCommand(args []string, parse func([]string) (*ui.DebuggerCommand, error), resultError func(*ui.DebuggerCommandResult) error) error {
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
	return r.handleCommand(args, func(args []string) (*ui.DebuggerCommand, error) {
		stepArgs := &ui.StepArgs{
			StepMode:  ui.StepModeInto,
			CountMode: ui.StepCountSourceLines,
		}

		return &ui.DebuggerCommand{
			Command:  ui.DebuggerCommandStep,
			StepArgs: stepArgs,
		}, nil
	}, func(result *ui.DebuggerCommandResult) error {
		return result.StepResult.Error
	})
}

func (r *REPL) handleContinue(args []string) error {
	return r.handleCommand(args, func(args []string) (*ui.DebuggerCommand, error) {
		return &ui.DebuggerCommand{
			Command: ui.DebuggerCommandContinue,
		}, nil
	}, func(result *ui.DebuggerCommandResult) error {
		return result.ContinueResult.Error
	})
}

func (r *REPL) handleInterrupt(args []string) error {
	return r.handleCommand(args, func(args []string) (*ui.DebuggerCommand, error) {
		return &ui.DebuggerCommand{
			Command: ui.DebuggerCommandInterrupt,
		}, nil
	}, func(result *ui.DebuggerCommandResult) error {
		return result.InterruptResult.Error
	})
}

func (r *REPL) handleRun(args []string) error {
	return r.handleCommand(args, func(args []string) (*ui.DebuggerCommand, error) {
		return &ui.DebuggerCommand{
			Command: ui.DebuggerCommandContinue,
		}, nil
	}, func(result *ui.DebuggerCommandResult) error {
		return result.ContinueResult.Error
	})
}

func (r *REPL) handleReset(args []string) error {
	return r.handleCommand(args, func(args []string) (*ui.DebuggerCommand, error) {
		return &ui.DebuggerCommand{
			Command: ui.DebuggerCommandReset,
		}, nil
	}, func(result *ui.DebuggerCommandResult) error {
		return result.ResetResult.Error
	})
}

func (r *REPL) handleRestart(args []string) error {
	return r.handleCommand(args, func(args []string) (*ui.DebuggerCommand, error) {
		return &ui.DebuggerCommand{
			Command: ui.DebuggerCommandRestart,
		}, nil
	}, func(result *ui.DebuggerCommandResult) error {
		return result.RestartResult.Error
	})
}

// ============================================================================
// Breakpoint Commands
// ============================================================================

func (r *REPL) handleBreak(args []string) error {
	return r.handleCommand(args, func(args []string) (*ui.DebuggerCommand, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("break requires an address")
		}

		addr, err := parseAddress(args[0])
		if err != nil {
			return nil, err
		}

		return &ui.DebuggerCommand{
			Command: ui.DebuggerCommandBreak,
			BreakArgs: &ui.BreakArgs{
				Address: &addr,
			},
		}, nil
	}, func(result *ui.DebuggerCommandResult) error {
		return result.BreakResult.Error
	})
}

func (r *REPL) handleRemoveBreakpoint(args []string) error {
	return r.handleCommand(args, func(args []string) (*ui.DebuggerCommand, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("removebreakpoint requires a breakpoint ID")
		}

		id, err := strconv.Atoi(args[0])
		if err != nil {
			return nil, fmt.Errorf("invalid breakpoint ID: %v", err)
		}

		return &ui.DebuggerCommand{
			Command: ui.DebuggerCommandRemoveBreakpoint,
			RemoveBreakpointArgs: &ui.RemoveBreakpointArgs{
				ID: id,
			},
		}, nil
	}, func(result *ui.DebuggerCommandResult) error {
		return result.RemoveBreakpointResult.Error
	})
}

func (r *REPL) handleWatch(args []string) error {
	return r.handleCommand(args, func(args []string) (*ui.DebuggerCommand, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("watch requires an address")
		}

		addr, err := parseAddress(args[0])
		if err != nil {
			return nil, err
		}

		// Default to 4-byte watch
		memRegion := &ui.MemoryRegion{
			Start: addr,
			Size:  4,
		}

		return &ui.DebuggerCommand{
			Command: ui.DebuggerCommandWatch,
			WatchArgs: &ui.WatchArgs{
				Range: memRegion,
			},
		}, nil
	}, func(result *ui.DebuggerCommandResult) error {
		return result.WatchResult.Error
	})
}

func (r *REPL) handleRemoveWatchpoint(args []string) error {
	return r.handleCommand(args, func(args []string) (*ui.DebuggerCommand, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("removewatchpoint requires a watchpoint ID")
		}

		id, err := strconv.Atoi(args[0])
		if err != nil {
			return nil, fmt.Errorf("invalid watchpoint ID: %v", err)
		}

		return &ui.DebuggerCommand{
			Command: ui.DebuggerCommandRemoveWatchpoint,
			RemoveWatchpointArgs: &ui.RemoveWatchpointArgs{
				ID: id,
			},
		}, nil
	}, func(result *ui.DebuggerCommandResult) error {
		return result.RemoveWatchpointResult.Error
	})
}

func (r *REPL) handleList(args []string) error {
	return r.handleCommand(args, func(args []string) (*ui.DebuggerCommand, error) {
		return &ui.DebuggerCommand{
			Command: ui.DebuggerCommandList,
		}, nil
	}, func(result *ui.DebuggerCommandResult) error {
		return result.ListResult.Error
	})
}

// ============================================================================
// Inspection Commands
// ============================================================================

func (r *REPL) handleDisasm(args []string) error {
	return r.handleCommand(args, func(args []string) (*ui.DebuggerCommand, error) {
		var disasmArgs ui.DisasmArgs
		if len(args) > 0 {
			disasmArgs.AddressExpr = args[0]
		}

		if len(args) > 1 {
			count, err := strconv.Atoi(args[1])
			if err != nil {
				return nil, fmt.Errorf("invalid count: %v", err)
			}
			disasmArgs.Count = count
		}

		return &ui.DebuggerCommand{
			Command:    ui.DebuggerCommandDisassemble,
			DisasmArgs: &disasmArgs,
		}, nil
	}, func(result *ui.DebuggerCommandResult) error {
		return result.DisassemblyResult.Error
	})
}

func (r *REPL) handleCurrent(args []string) error {
	return r.handleCommand(args, func(args []string) (*ui.DebuggerCommand, error) {
		return &ui.DebuggerCommand{
			Command: ui.DebuggerCommandCurrentInstruction,
		}, nil
	}, func(result *ui.DebuggerCommandResult) error {
		return result.CurrentInstructionResult.Error
	})
}

func (r *REPL) handleMemory(args []string) error {
	return r.handleCommand(args, func(args []string) (*ui.DebuggerCommand, error) {
		memArgs := &ui.MemoryArgs{}

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

		return &ui.DebuggerCommand{
			Command:    ui.DebuggerCommandMemory,
			MemoryArgs: memArgs,
		}, nil
	}, func(result *ui.DebuggerCommandResult) error {
		return result.MemoryResult.Error
	})
}

func (r *REPL) handleSource(args []string) error {
	return r.handleCommand(args, func(args []string) (*ui.DebuggerCommand, error) {
		sourceArgs := &ui.SourceArgs{}

		if len(args) > 0 {
			sourceArgs.File = args[0]
		}

		return &ui.DebuggerCommand{
			Command:    ui.DebuggerCommandSource,
			SourceArgs: sourceArgs,
		}, nil
	}, func(result *ui.DebuggerCommandResult) error {
		return result.SourceResult.Error
	})
}

func (r *REPL) handleInfo(args []string) error {
	infoType := ui.InfoTypeGeneral
	if len(args) > 0 {
		var err error
		infoType, err = ui.InfoTypeFromString(args[0])
		if err != nil {
			return err
		}
	}

	return r.handleCommand(args, func(args []string) (*ui.DebuggerCommand, error) {
		return &ui.DebuggerCommand{
			Command: ui.DebuggerCommandInfo,
			InfoArgs: &ui.InfoArgs{
				Type: infoType,
			},
		}, nil
	}, func(result *ui.DebuggerCommandResult) error {
		return result.InfoResult.Error
	})
}

func (r *REPL) handleSymbols(args []string) error {
	return r.handleCommand(args, func(args []string) (*ui.DebuggerCommand, error) {
		symbolsArgs := &ui.SymbolsArgs{}

		if len(args) > 0 {
			symbolsArgs.SymbolName = &args[0]
		}

		return &ui.DebuggerCommand{
			Command:     ui.DebuggerCommandSymbols,
			SymbolsArgs: symbolsArgs,
		}, nil
	}, func(result *ui.DebuggerCommandResult) error {
		return result.SymbolsResult.Error
	})
}

func (r *REPL) handleRegisters(args []string) error {
	return r.handleCommand(args, func(args []string) (*ui.DebuggerCommand, error) {
		return &ui.DebuggerCommand{
			Command: ui.DebuggerCommandRegisters,
		}, nil
	}, func(result *ui.DebuggerCommandResult) error {
		return result.RegistersResult.Error
	})
}

func (r *REPL) handleStack(args []string) error {
	return r.handleCommand(args, func(args []string) (*ui.DebuggerCommand, error) {
		return &ui.DebuggerCommand{
			Command: ui.DebuggerCommandStack,
		}, nil
	}, func(result *ui.DebuggerCommandResult) error {
		return result.StackResult.Error
	})
}

func (r *REPL) handleVars(args []string) error {
	return r.handleCommand(args, func(args []string) (*ui.DebuggerCommand, error) {
		return &ui.DebuggerCommand{
			Command: ui.DebuggerCommandVariables,
		}, nil
	}, func(result *ui.DebuggerCommandResult) error {
		return result.VariablesResult.Error
	})
}

func (r *REPL) handleEval(args []string) error {
	return r.handleCommand(args, func(args []string) (*ui.DebuggerCommand, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("eval requires an expression")
		}

		expression := strings.Join(args, " ")
		return &ui.DebuggerCommand{
			Command: ui.DebuggerCommandEvaluateExpression,
			EvalArgs: &ui.EvalArgs{
				Expression: expression,
			},
		}, nil
	}, func(result *ui.DebuggerCommandResult) error {
		return result.EvalResult.Error
	})
}

// ============================================================================
// Program Loading Commands
// ============================================================================

func (r *REPL) handleLoad(args []string) error {
	return r.handleCommand(args, func(args []string) (*ui.DebuggerCommand, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("load requires a file path")
		}

		return &ui.DebuggerCommand{
			Command: ui.DebuggerCommandLoadProgramFromFile,
			LoadProgramArgs: &ui.LoadProgramArgs{
				FilePath: args[0],
			},
		}, nil
	}, func(result *ui.DebuggerCommandResult) error {
		return result.LoadProgramResult.Error
	})
}

func (r *REPL) handleLoadProgram(args []string) error {
	return r.handleCommand(args, func(args []string) (*ui.DebuggerCommand, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("loadprogram requires a file path")
		}

		return &ui.DebuggerCommand{
			Command: ui.DebuggerCommandLoadProgramFromFile,
			LoadProgramArgs: &ui.LoadProgramArgs{
				FilePath: args[0],
			},
		}, nil
	}, func(result *ui.DebuggerCommandResult) error {
		return result.LoadProgramResult.Error
	})
}

func (r *REPL) handleLoadSystem(args []string) error {
	return r.handleCommand(args, func(args []string) (*ui.DebuggerCommand, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("loadsystem requires a file path")
		}

		return &ui.DebuggerCommand{
			Command: ui.DebuggerCommandLoadSystemFromFile,
			LoadSystemArgs: &ui.LoadSystemArgs{
				FilePath: args[0],
			},
		}, nil
	}, func(result *ui.DebuggerCommandResult) error {
		return result.LoadSystemResult.Error
	})
}

func (r *REPL) handleLoadRuntime(args []string) error {
	return r.handleCommand(args, func(args []string) (*ui.DebuggerCommand, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("loadruntime requires a runtime name")
		}

		runtimeType, err := ui.RuntimeTypeFromString(args[0])
		if err != nil {
			return nil, err
		}

		return &ui.DebuggerCommand{
			Command: ui.DebuggerCommandLoadRuntime,
			LoadRuntimeArgs: &ui.LoadRuntimeArgs{
				Runtime: runtimeType,
			},
		}, nil
	}, func(result *ui.DebuggerCommandResult) error {
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

	// Special handling for show.logs - it takes multiple logger names
	if settingName == SettingKeyShowLogs {
		loggerNames := args[1:] // Get all remaining arguments as logger names
		if err := r.applyUILoggers(loggerNames); err != nil {
			return err
		}
		r.write("UI logging enabled for loggers: %v\n", loggerNames)
		return nil
	}

	// For other settings, require exactly one value
	if len(args) < 2 {
		return fmt.Errorf("set requires a setting name and value")
	}

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

func (r *REPL) handleDebuggerEvent(event *ui.DebuggerEvent) {
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
	case ui.DebuggerEventStepped:
		r.write("  Instruction stepped\n")
		r.printEventExecutionDetails(event.Result)

	case ui.DebuggerEventBreakpointHit:
		r.write("  Breakpoint hit\n")
		if event.Result != nil && event.Result.Breakpoint != nil {
			r.write("    Breakpoint ID: %d\n", event.Result.Breakpoint.ID)
			r.write("    Address: 0x%x\n", event.Result.Breakpoint.Address)
		}
		r.printEventExecutionDetails(event.Result)

	case ui.DebuggerEventWatchpointHit:
		r.write("  Watchpoint hit\n")
		if event.Result != nil && event.Result.Watchpoint != nil {
			r.write("    Watchpoint ID: %d\n", event.Result.Watchpoint.ID)
			r.write("    Address Range: 0x%x - 0x%x (size: %d)\n",
				event.Result.Watchpoint.Range.Start,
				event.Result.Watchpoint.Range.Start+event.Result.Watchpoint.Range.Size,
				event.Result.Watchpoint.Range.Size)
		}
		r.printEventExecutionDetails(event.Result)

	case ui.DebuggerEventProgramTerminated:
		r.write("  Program terminated\n")
		r.printEventExecutionDetails(event.Result)

	case ui.DebuggerEventProgramHalted:
		r.write("  Program halted\n")
		r.printEventExecutionDetails(event.Result)

	case ui.DebuggerEventError:
		r.write("  Error occurred\n")
		if event.Result != nil && event.Result.Error != nil {
			r.write("    Error: %v\n", event.Result.Error)
		}
		r.printEventExecutionDetails(event.Result)

	case ui.DebuggerEventSourceLocationChanged:
		r.write("  Source location changed\n")
		if event.Result != nil && event.Result.LastLocation != nil {
			r.write("    Location: %s:%d\n",
				event.Result.LastLocation.File,
				event.Result.LastLocation.Line)
		}
		r.printEventExecutionDetails(event.Result)

	case ui.DebuggerEventInterrupted:
		r.write("  Execution interrupted\n")
		r.printEventExecutionDetails(event.Result)

	case ui.DebuggerEventProgramLoaded:
		r.write("  Program loaded\n")

	case ui.DebuggerEventLagging:
		r.write("  Emulator lagging\n")
		if event.Result != nil && event.Result.LaggingCycles > 0 {
			r.write("    Lagging by: %d cycles\n", event.Result.LaggingCycles)
		}
		r.printEventExecutionDetails(event.Result)
	}
}

// printEventExecutionDetails prints common execution result details
func (r *REPL) printEventExecutionDetails(result *ui.ExecutionResult) {
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
	if result.StopReason != ui.StopReasonNone {
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
