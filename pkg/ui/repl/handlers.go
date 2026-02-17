package repl

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/ui"
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
			addr, err := parseAddress(args[0])
			if err != nil {
				return nil, err
			}
			disasmArgs.Address = addr
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
	return r.handleCommand(args, func(args []string) (*ui.DebuggerCommand, error) {
		return &ui.DebuggerCommand{
			Command: ui.DebuggerCommandInfo,
		}, nil
	}, func(result *ui.DebuggerCommandResult) error {
		return result.InfoResult.Error
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
