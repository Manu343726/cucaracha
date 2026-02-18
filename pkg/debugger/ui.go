package debugger

import (
	"fmt"

	"github.com/Manu343726/cucaracha/pkg/ui"
)

// Implements the command execution ui.Debugger interface
type DebuggerUI struct {
	commands Debugger
}

func NewDebuggerForUI(commands Debugger) ui.Debugger {
	return &DebuggerUI{
		commands: commands,
	}
}

func (d *DebuggerUI) SetEventCallback(callback ui.DebuggerEventCallback) {
	d.commands.SetEventCallback(callback)
}

func (d *DebuggerUI) Reset() *ui.ExecutionResult {
	return d.commands.Reset()
}

func (d *DebuggerUI) Restart() *ui.ExecutionResult {
	return d.commands.Restart()
}

func (d *DebuggerUI) Execute(args *ui.DebuggerCommand) (*ui.DebuggerCommandResult, error) {
	result := &ui.DebuggerCommandResult{
		Id:      args.Id,
		Command: args.Command,
	}

	switch args.Command {
	case ui.DebuggerCommandStep:
		result.StepResult = d.commands.Step(args.StepArgs)
	case ui.DebuggerCommandContinue:
		result.ContinueResult = d.commands.Continue()
	case ui.DebuggerCommandInterrupt:
		result.InterruptResult = d.commands.Interrupt()
	case ui.DebuggerCommandBreak:
		result.BreakResult = d.commands.Break(args.BreakArgs)
	case ui.DebuggerCommandWatch:
		result.WatchResult = d.commands.Watch(args.WatchArgs)
	case ui.DebuggerCommandRemoveBreakpoint:
		result.RemoveBreakpointResult = d.commands.RemoveBreakpoint(args.RemoveBreakpointArgs)
	case ui.DebuggerCommandRemoveWatchpoint:
		result.RemoveWatchpointResult = d.commands.RemoveWatchpoint(args.RemoveWatchpointArgs)
	case ui.DebuggerCommandList:
		result.ListResult = d.commands.List()
	case ui.DebuggerCommandDisassemble:
		result.DisassemblyResult = d.commands.Disasm(args.DisasmArgs)
	case ui.DebuggerCommandCurrentInstruction:
		result.CurrentInstructionResult = d.commands.CurrentInstruction()
	case ui.DebuggerCommandMemory:
		result.MemoryResult = d.commands.Memory(args.MemoryArgs)
	case ui.DebuggerCommandSource:
		result.SourceResult = d.commands.Source(args.SourceArgs)
	case ui.DebuggerCommandCurrentSource:
		result.CurrentSourceResult = d.commands.CurrentSource(args.CurrentSourceArgs)
	case ui.DebuggerCommandEvaluateExpression:
		result.EvalResult = d.commands.Eval(args.EvalArgs)
	case ui.DebuggerCommandInfo:
		infoArgs := args.InfoArgs
		if infoArgs == nil {
			infoArgs = &ui.InfoArgs{Type: ui.InfoTypeGeneral}
		}
		result.InfoResult = d.commands.Info(infoArgs)
	case ui.DebuggerCommandRegisters:
		result.RegistersResult = d.commands.Registers()
	case ui.DebuggerCommandStack:
		result.StackResult = d.commands.Stack()
	case ui.DebuggerCommandVariables:
		result.VariablesResult = d.commands.Vars()
	case ui.DebuggerCommandSymbols:
		result.SymbolsResult = d.commands.Symbols(args.SymbolsArgs)
	case ui.DebuggerCommandReset:
		result.ResetResult = d.commands.Reset()
	case ui.DebuggerCommandRestart:
		result.RestartResult = d.commands.Restart()
	case ui.DebuggerCommandLoadProgramFromFile:
		result.LoadProgramResult = d.commands.LoadProgramFromFile(args.LoadProgramArgs)
	case ui.DebuggerCommandLoadSystemFromFile:
		result.LoadSystemResult = d.commands.LoadSystemFromFile(args.LoadSystemArgs)
	case ui.DebuggerCommandLoadRuntime:
		result.LoadRuntimeResult = d.commands.LoadRuntime(args.LoadRuntimeArgs)
	case ui.DebuggerCommandLoad:
		result.LoadResult = d.commands.Load(args.LoadArgs)
	default:
		return nil, fmt.Errorf("unknown debugger command: '%s'", args.Command)
	}

	return result, nil
}
