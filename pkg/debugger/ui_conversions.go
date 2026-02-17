package debugger

import (
	"github.com/Manu343726/cucaracha/pkg/debugger/core"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/registers"
	"github.com/Manu343726/cucaracha/pkg/hw/memory"
	"github.com/Manu343726/cucaracha/pkg/runtime/program"
	"github.com/Manu343726/cucaracha/pkg/runtime/program/sourcecode"
	"github.com/Manu343726/cucaracha/pkg/ui"
	"github.com/Manu343726/cucaracha/pkg/utils"
)

func SourceLocationToUI(loc *sourcecode.Location) *ui.SourceLocation {
	if loc == nil {
		return nil
	}

	return &ui.SourceLocation{
		File: loc.File.Path(),
		Line: loc.Line,
	}
}

func SourceLocationFromUI(loc *ui.SourceLocation) *sourcecode.Location {
	if loc == nil {
		return nil
	}

	return &sourcecode.Location{
		File: sourcecode.FileNamed(loc.File),
		Line: loc.Line,
	}
}

func SourceLineToUI(line *sourcecode.Line) *ui.SourceLine {
	if line == nil {
		return nil
	}

	return &ui.SourceLine{
		Location: SourceLocationToUI(line.Location),
		Text:     line.Text,
	}
}

func SourceRangeFromUI(rng *ui.SourceRange) *sourcecode.Range {
	if rng == nil {
		return nil
	}

	return &sourcecode.Range{
		File:      sourcecode.FileNamed(rng.Start.File),
		StartLine: rng.Start.Line,
		LineCount: rng.Lines,
	}
}

func SourceRangeToUI(rng *sourcecode.Range) *ui.SourceRange {
	if rng == nil {
		return nil
	}

	return &ui.SourceRange{
		Start: &ui.SourceLocation{
			File: rng.File.Path(),
			Line: rng.StartLine,
		},
		Lines: rng.LineCount,
	}
}

func SourceCodeSnippetToUI(snippet *sourcecode.Snippet) *ui.SourceCodeSnippet {
	if snippet == nil {
		return nil
	}

	return &ui.SourceCodeSnippet{
		SourceRange: SourceRangeToUI(snippet.Range),
		Lines:       utils.Map(snippet.Lines, SourceLineToUI),
	}
}

func MemoryRangeToUI(region *memory.Range) *ui.MemoryRegion {
	if region == nil {
		return nil
	}

	return &ui.MemoryRegion{
		Start: region.Start,
		Size:  region.Size,
	}
}

func MemoryRangeFromUI(region *ui.MemoryRegion) *memory.Range {
	if region == nil {
		return nil
	}

	return &memory.Range{
		Start: region.Start,
		Size:  region.Size,
	}
}

func WatchpointTypeFromUI(t ui.WatchpointType) core.WatchpointType {
	switch t {
	case ui.WatchpointTypeRead:
		return core.WatchRead
	case ui.WatchpointTypeWrite:
		return core.WatchWrite
	case ui.WatchpointTypeReadWrite:
		return core.WatchReadWrite
	}

	panic("unknown UI watchpoint type")
}

func WatchpointTypeToUI(t core.WatchpointType) ui.WatchpointType {
	switch t {
	case core.WatchRead:
		return ui.WatchpointTypeRead
	case core.WatchWrite:
		return ui.WatchpointTypeWrite
	case core.WatchReadWrite:
		return ui.WatchpointTypeReadWrite
	}

	panic("unknown core watchpoint type")
}

func WatchpointToUI(wp *core.Watchpoint) *ui.Watchpoint {
	if wp == nil {
		return nil
	}

	return &ui.Watchpoint{
		ID:      wp.ID,
		Range:   MemoryRangeToUI(wp.Memory),
		Type:    WatchpointTypeToUI(wp.Type),
		Enabled: wp.Enabled,
	}
}

func WatchpointFromUI(wp *ui.Watchpoint) *core.Watchpoint {
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

func BreakpointToUI(bp *core.Breakpoint) *ui.Breakpoint {
	if bp == nil {
		return nil
	}

	return &ui.Breakpoint{
		ID:      bp.ID,
		Address: bp.Address,
		Enabled: bp.Enabled,
	}
}

func BreakpointFromUI(bp *ui.Breakpoint) *core.Breakpoint {
	if bp == nil {
		return nil
	}

	return &core.Breakpoint{
		ID:      bp.ID,
		Address: bp.Address,
		Enabled: bp.Enabled,
	}
}

func InstructionOperandKindToUI(kind instructions.OperandKind) ui.InstructionOperandKind {
	switch kind {
	case instructions.OperandKind_Register:
		return ui.OperandKindRegister
	case instructions.OperandKind_Immediate:
		return ui.OperandKindImmediate
	}

	panic("unknown instruction operand kind")
}

func RegisterToUI(reg *registers.RegisterDescriptor) *ui.Register {
	if reg == nil {
		return nil
	}

	return &ui.Register{
		Name:     reg.Name(),
		Encoding: uint32(reg.Encode()),
	}
}

func RegisterValueToUI(reg *registers.RegisterDescriptor, value uint32) *ui.Register {
	result := RegisterToUI(reg)
	if result != nil {
		result.Value = value
	}

	return result
}

func RegistersValuesToUI(regs map[string]uint32) map[string]*ui.Register {
	result := make(map[string]*ui.Register, len(regs))

	for name, value := range regs {
		result[name] = &ui.Register{
			Name:  name,
			Value: value,
		}
	}

	return result
}

func InstructionOperandToUI(op *instructions.OperandValue) *ui.InstructionOperand {
	if op == nil {
		return nil
	}

	result := &ui.InstructionOperand{
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

func InstructionToUI(instr *program.Instruction) *ui.Instruction {
	if instr == nil {
		return nil
	}

	uiInstr := &ui.Instruction{
		Address:  utils.DerefOr(instr.Address, 0),
		Encoding: instr.Raw.Encode(),
		Mnemonic: instr.Instruction.Descriptor.OpCode.Mnemonic,
		Text:     instr.Text,
		Operands: utils.Map(instr.Instruction.OperandValues, utils.PtrFunc(InstructionOperandToUI)),
	}

	return uiInstr
}

func StopReasonToUI(r core.StopReason) ui.StopReason {
	switch r {
	case core.StopNone:
		return ui.StopReasonNone
	case core.StopStep:
		return ui.StopReasonStep
	case core.StopBreakpoint:
		return ui.StopReasonBreakpoint
	case core.StopWatchpoint:
		return ui.StopReasonWatchpoint
	case core.StopHalt:
		return ui.StopReasonHalt
	case core.StopError:
		return ui.StopReasonError
	case core.StopTermination:
		return ui.StopReasonTermination
	case core.StopMaxSteps:
		return ui.StopReasonMaxSteps
	case core.StopInterrupt:
		return ui.StopReasonInterrupt
	default:
		return ui.StopReasonNone
	}
}

func ExecutionResultToUI(result *core.ExecutionResult) *ui.ExecutionResult {
	if result == nil {
		return nil
	}

	return &ui.ExecutionResult{
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

func EventTypeToUI(eventType core.DebugEvent) ui.DebuggerEventType {
	switch eventType {
	case core.EventProgramLoaded:
		return ui.DebuggerEventProgramLoaded
	case core.EventStepped:
		return ui.DebuggerEventStepped
	case core.EventBreakpointHit:
		return ui.DebuggerEventBreakpointHit
	case core.EventWatchpointHit:
		return ui.DebuggerEventWatchpointHit
	case core.EventProgramTerminated:
		return ui.DebuggerEventProgramTerminated
	case core.EventProgramHalted:
		return ui.DebuggerEventProgramHalted
	case core.EventError:
		return ui.DebuggerEventError
	case core.EventSourceLocationChanged:
		return ui.DebuggerEventSourceLocationChanged
	case core.EventInterrupted:
		return ui.DebuggerEventInterrupted
	case core.EventLagging:
		return ui.DebuggerEventLagging
	default:
		return ui.DebuggerEventType(0) // Unknown/undefined event type
	}
}

func EventToUI(event *core.Event) *ui.DebuggerEvent {
	if event == nil {
		return nil
	}

	return &ui.DebuggerEvent{
		Type:   EventTypeToUI(event.Event),
		Result: ExecutionResultToUI(event.Result),
	}
}
