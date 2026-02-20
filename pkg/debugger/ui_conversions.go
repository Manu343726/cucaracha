package debugger

import (
	"github.com/Manu343726/cucaracha/pkg/debugger/core"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/registers"
	"github.com/Manu343726/cucaracha/pkg/hw/memory"
	"github.com/Manu343726/cucaracha/pkg/runtime"
	"github.com/Manu343726/cucaracha/pkg/runtime/program"
	"github.com/Manu343726/cucaracha/pkg/runtime/program/sourcecode"
	uiDebugger "github.com/Manu343726/cucaracha/pkg/ui/debugger"
	"github.com/Manu343726/cucaracha/pkg/utils"
)

func SourceLocationToUI(loc *sourcecode.Location) *uiDebugger.SourceLocation {
	if loc == nil {
		return nil
	}

	return &uiDebugger.SourceLocation{
		File: loc.File.Path(),
		Line: loc.Line,
	}
}

func SourceLocationFromUI(loc *uiDebugger.SourceLocation) *sourcecode.Location {
	if loc == nil {
		return nil
	}

	return &sourcecode.Location{
		File: sourcecode.FileNamed(loc.File),
		Line: loc.Line,
	}
}

func SourceLineToUI(line *sourcecode.Line) *uiDebugger.SourceLine {
	if line == nil {
		return nil
	}

	return &uiDebugger.SourceLine{
		Location: SourceLocationToUI(line.Location),
		Text:     line.Text,
	}
}

func SourceRangeFromUI(rng *uiDebugger.SourceRange) *sourcecode.Range {
	if rng == nil {
		return nil
	}

	return &sourcecode.Range{
		File:      sourcecode.FileNamed(rng.Start.File),
		StartLine: rng.Start.Line,
		LineCount: rng.Lines,
	}
}

func SourceRangeToUI(rng *sourcecode.Range) *uiDebugger.SourceRange {
	if rng == nil {
		return nil
	}

	return &uiDebugger.SourceRange{
		Start: &uiDebugger.SourceLocation{
			File: rng.File.Path(),
			Line: rng.StartLine,
		},
		Lines: rng.LineCount,
	}
}

func SourceCodeSnippetToUI(snippet *sourcecode.Snippet) *uiDebugger.SourceCodeSnippet {
	if snippet == nil {
		return nil
	}

	return &uiDebugger.SourceCodeSnippet{
		SourceRange: SourceRangeToUI(snippet.Range),
		Lines:       utils.Map(snippet.Lines, SourceLineToUI),
	}
}

func MemoryRangeToUI(region *memory.Range) *uiDebugger.MemoryRegion {
	if region == nil {
		return nil
	}

	return &uiDebugger.MemoryRegion{
		Start: region.Start,
		Size:  region.Size,
	}
}

func MemoryRangeFromUI(region *uiDebugger.MemoryRegion) *memory.Range {
	if region == nil {
		return nil
	}

	return &memory.Range{
		Start: region.Start,
		Size:  region.Size,
	}
}

func WatchpointTypeFromUI(t uiDebugger.WatchpointType) core.WatchpointType {
	switch t {
	case uiDebugger.WatchpointTypeRead:
		return core.WatchRead
	case uiDebugger.WatchpointTypeWrite:
		return core.WatchWrite
	case uiDebugger.WatchpointTypeReadWrite:
		return core.WatchReadWrite
	}

	panic("unknown UI watchpoint type")
}

func WatchpointTypeToUI(t core.WatchpointType) uiDebugger.WatchpointType {
	switch t {
	case core.WatchRead:
		return uiDebugger.WatchpointTypeRead
	case core.WatchWrite:
		return uiDebugger.WatchpointTypeWrite
	case core.WatchReadWrite:
		return uiDebugger.WatchpointTypeReadWrite
	}

	panic("unknown core watchpoint type")
}

func WatchpointToUI(wp *core.Watchpoint) *uiDebugger.Watchpoint {
	if wp == nil {
		return nil
	}

	return &uiDebugger.Watchpoint{
		ID:      wp.ID,
		Range:   MemoryRangeToUI(wp.Memory),
		Type:    WatchpointTypeToUI(wp.Type),
		Enabled: wp.Enabled,
	}
}

func WatchpointFromUI(wp *uiDebugger.Watchpoint) *core.Watchpoint {
	if wp == nil {
		return nil
	}

	return &core.Watchpoint{
		ID:      wp.ID,
		Memory:  MemoryRangeFromUI(wp.Range),
		Type:    WatchpointTypeFromUI(wp.Type),
		Enabled: wp.Enabled,
	}
}

func BreakpointToUI(bp *core.Breakpoint) *uiDebugger.Breakpoint {
	if bp == nil {
		return nil
	}

	return &uiDebugger.Breakpoint{
		ID:      bp.ID,
		Address: bp.Address,
		Enabled: bp.Enabled,
	}
}

func BreakpointFromUI(bp *uiDebugger.Breakpoint) *core.Breakpoint {
	if bp == nil {
		return nil
	}

	return &core.Breakpoint{
		ID:      bp.ID,
		Address: bp.Address,
		Enabled: bp.Enabled,
	}
}

func InstructionOperandKindToUI(kind instructions.OperandKind) uiDebugger.InstructionOperandKind {
	switch kind {
	case instructions.OperandKind_Register:
		return uiDebugger.OperandKindRegister
	case instructions.OperandKind_Immediate:
		return uiDebugger.OperandKindImmediate
	}

	panic("unknown instruction operand kind")
}

func RegisterToUI(reg *registers.RegisterDescriptor) *uiDebugger.Register {
	if reg == nil {
		return nil
	}

	return &uiDebugger.Register{
		Name:     reg.Name(),
		Encoding: uint32(reg.Encode()),
	}
}

func RegisterValueToUI(reg *registers.RegisterDescriptor, value uint32) *uiDebugger.Register {
	result := RegisterToUI(reg)
	if result != nil {
		result.Value = value
	}

	return result
}

func RegistersValuesToUI(regs map[string]uint32) map[string]*uiDebugger.Register {
	result := make(map[string]*uiDebugger.Register, len(regs))

	for name, value := range regs {
		result[name] = &uiDebugger.Register{
			Name:  name,
			Value: value,
		}
	}

	return result
}

func InstructionOperandToUI(op *instructions.OperandValue) *uiDebugger.InstructionOperand {
	if op == nil {
		return nil
	}

	result := &uiDebugger.InstructionOperand{
		Kind: InstructionOperandKindToUI(op.Kind()),
	}

	switch op.Kind() {
	case instructions.OperandKind_Register:
		result.Register = RegisterToUI(op.Register())
	case instructions.OperandKind_Immediate:
		result.Immediate = utils.Ptr(uint32(op.Immediate().Encode()))
	default:
		panic("unsupported operand kind")
	}

	return result
}

func InstructionToUI(instr *program.Instruction) *uiDebugger.Instruction {
	if instr == nil {
		return nil
	}

	uiInstr := &uiDebugger.Instruction{
		Address:  utils.DerefOr(instr.Address, 0),
		Encoding: instr.Raw.Encode(),
		Mnemonic: instr.Instruction.Descriptor.OpCode.Mnemonic,
		Text:     instr.Text,
		Operands: utils.Map(instr.Instruction.OperandValues, utils.PtrFunc(InstructionOperandToUI)),
	}

	return uiInstr
}

func StopReasonToUI(r core.StopReason) uiDebugger.StopReason {
	switch r {
	case core.StopNone:
		return uiDebugger.StopReasonNone
	case core.StopStep:
		return uiDebugger.StopReasonStep
	case core.StopBreakpoint:
		return uiDebugger.StopReasonBreakpoint
	case core.StopWatchpoint:
		return uiDebugger.StopReasonWatchpoint
	case core.StopHalt:
		return uiDebugger.StopReasonHalt
	case core.StopError:
		return uiDebugger.StopReasonError
	case core.StopTermination:
		return uiDebugger.StopReasonTermination
	case core.StopMaxSteps:
		return uiDebugger.StopReasonMaxSteps
	case core.StopInterrupt:
		return uiDebugger.StopReasonInterrupt
	default:
		return uiDebugger.StopReasonNone
	}
}

func ExecutionResultToUI(result *core.ExecutionResult) *uiDebugger.ExecutionResult {
	if result == nil {
		return nil
	}

	return &uiDebugger.ExecutionResult{
		Error:           result.Error,
		StopReason:      StopReasonToUI(result.StopReason),
		Steps:           uint64(result.StepsExecuted),
		Cycles:          uint64(result.CyclesExecuted),
		Breakpoint:      BreakpointToUI(result.Breakpoint),
		Watchpoint:      WatchpointToUI(result.Watchpoint),
		LastInstruction: result.LastPC,
		LaggingCycles:   uint32(result.LagCycles),
	}
}

func EventTypeToUI(eventType core.DebugEvent) uiDebugger.DebuggerEventType {
	switch eventType {
	case core.EventProgramLoaded:
		return uiDebugger.DebuggerEventProgramLoaded
	case core.EventStepped:
		return uiDebugger.DebuggerEventStepped
	case core.EventBreakpointHit:
		return uiDebugger.DebuggerEventBreakpointHit
	case core.EventWatchpointHit:
		return uiDebugger.DebuggerEventWatchpointHit
	case core.EventProgramTerminated:
		return uiDebugger.DebuggerEventProgramTerminated
	case core.EventProgramHalted:
		return uiDebugger.DebuggerEventProgramHalted
	case core.EventError:
		return uiDebugger.DebuggerEventError
	case core.EventSourceLocationChanged:
		return uiDebugger.DebuggerEventSourceLocationChanged
	case core.EventInterrupted:
		return uiDebugger.DebuggerEventInterrupted
	case core.EventLagging:
		return uiDebugger.DebuggerEventLagging
	default:
		return uiDebugger.DebuggerEventType(0) // Unknown/undefined event type
	}
}

func EventToUI(event *core.Event) *uiDebugger.DebuggerEvent {
	if event == nil {
		return nil
	}

	return &uiDebugger.DebuggerEvent{
		Type:   EventTypeToUI(event.Event),
		Result: ExecutionResultToUI(event.Result),
	}
}

// DetermineDetailedDebuggerStatus returns a more detailed debugger status by examining
// the runner state and the last execution result's stop reason.
// This provides better insight into why the debugger is in its current state.
func DetermineDetailedDebuggerStatus(debugger core.Debugger) uiDebugger.DebuggerStatus {
	if debugger == nil {
		return uiDebugger.DebuggerStatusNotReady_MissingProgram
	}

	runnerState := debugger.GetRunnerState()
	lastResult := debugger.LastResult()

	switch runnerState {
	case runtime.RunnerStateIdle:
		return uiDebugger.DebuggerStatusIdle

	case runtime.RunnerStateRunning:
		return uiDebugger.DebuggerStatusRunning

	case runtime.RunnerStateInterrupted:
		// When interrupted, check the stop reason for more details
		if lastResult != nil {
			switch lastResult.StopReason {
			case core.StopTermination:
				return uiDebugger.DebuggerStatusTerminated
			case core.StopError:
				return uiDebugger.DebuggerStatusPaused
			case core.StopHalt:
				return uiDebugger.DebuggerStatusPaused
			// All other stop reasons indicate paused state (breakpoint, watchpoint, step, etc.)
			default:
				return uiDebugger.DebuggerStatusPaused
			}
		}
		return uiDebugger.DebuggerStatusPaused

	case runtime.RunnerStateStopped:
		// Runner stopped, check if it's terminated or halted
		if lastResult != nil {
			switch lastResult.StopReason {
			case core.StopTermination:
				return uiDebugger.DebuggerStatusTerminated
			case core.StopHalt:
				return uiDebugger.DebuggerStatusPaused
			case core.StopError:
				return uiDebugger.DebuggerStatusPaused
			}
		}
		return uiDebugger.DebuggerStatusTerminated

	default:
		return uiDebugger.DebuggerStatusRunning
	}
}
