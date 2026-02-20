// Package interpreter provides debugging and step-by-step execution support
// for the Cucaracha CPU interpreter.
package core

import (
	"fmt"
	"iter"
	"log/slog"
	"maps"
	"slices"
	"sort"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/Manu343726/cucaracha/pkg/hw/memory"
	"github.com/Manu343726/cucaracha/pkg/runtime"
	"github.com/Manu343726/cucaracha/pkg/runtime/program"
	"github.com/Manu343726/cucaracha/pkg/runtime/program/sourcecode"
	"github.com/Manu343726/cucaracha/pkg/utils"
	"github.com/Manu343726/cucaracha/pkg/utils/contract"
	"github.com/Manu343726/cucaracha/pkg/utils/logging"
)

// debugger provides debugging capabilities for the interpreter
type debugger struct {
	contract.Base

	// execution runtime that is being debugged
	runtime runtime.Runtime
	runner  *runtime.Runner

	// Program being debugged
	programFile program.ProgramFile

	// Breakpoints indexed by ID
	breakpoints map[int]*Breakpoint
	// Breakpoint addresses for fast lookup
	breakpointAddrs map[uint32]*Breakpoint
	// Next breakpoint ID
	nextBreakpointID int

	// Watchpoints indexed by ID
	watchpoints map[int]*Watchpoint
	// Next watchpoint ID
	nextWatchpointID int

	// Termination addresses (e.g., return address from main)
	terminationAddrs map[uint32]bool

	// Event callback
	eventCallback EventCallback

	// Internal events queue
	eventsQueue *InternalEventsQueue

	// Execution state
	lastResult         *ExecutionResult
	lastSourceLocation *sourcecode.Location
}

// NewDebugger creates a new debugger for the given runtime
func NewDebugger(r runtime.Runtime, program program.ProgramFile) Debugger {
	runner := runtime.NewRunner(r)

	// We use the runtime adapter from the runner instead
	// of the original runtime to ensure that all interactions with the debugged
	// runtime go through the runner and are properly synchronized.
	runtime := runner.Runtime()

	d := &debugger{
		Base:             contract.NewBase(log().Child("debugger")),
		runtime:          runtime,
		runner:           runner,
		programFile:      program,
		breakpoints:      make(map[int]*Breakpoint),
		breakpointAddrs:  make(map[uint32]*Breakpoint),
		watchpoints:      make(map[int]*Watchpoint),
		terminationAddrs: make(map[uint32]bool),
		eventsQueue:      NewInternalEventsQueue(),
	}

	go d.monitorRunnerEvents()

	return d
}

func (d *debugger) monitorRunnerEvents() {
	for event := range d.runner.Events() {
		d.handleRunnerEvent(event)
	}
}

func (d *debugger) handleRunnerEvent(event runtime.RunnerEvent) {
	// Convert runner event to debugger event and trigger callback
	debugEvent := &Event{
		Result: &ExecutionResult{},
	}

	switch event.Type {
	case runtime.RunnerEventBreakpointHit:
		if addr, ok := event.Data.(uint32); ok {
			debugEvent.Event = EventBreakpointHit
			debugEvent.Result.StopReason = StopBreakpoint

			// Find the corresponding breakpoint in the debugger's breakpoint list
			if bp := d.GetBreakpointAt(addr); bp != nil {
				bp.HitCount++
				debugEvent.Result.Breakpoint = bp
			}
		}

	case runtime.RunnerEventWatchpointHit:
		if watchRange, ok := event.Data.(memory.Range); ok {
			debugEvent.Event = EventWatchpointHit
			debugEvent.Result.StopReason = StopWatchpoint

			// Find the corresponding watchpoint in the debugger's watchpoint list
			for _, wp := range d.ListWatchpoints() {
				if wp.Memory != nil && wp.Memory.Start == watchRange.Start && wp.Memory.End() == watchRange.End() {
					wp.HitCount++
					debugEvent.Result.Watchpoint = wp
					break
				}
			}
		}

	case runtime.RunnerEventStepCompleted:
		if stepInfo, ok := event.Data.(*cpu.StepInfo); ok {
			debugEvent.Event = EventStepped
			debugEvent.Result.StopReason = StopStep
			debugEvent.Result.CyclesExecuted = int64(stepInfo.CyclesUsed)
			debugEvent.Result.LastPC = stepInfo.InstructionAddress

			if stepInfo.InstructionAddress != stepInfo.NextInstructionAddress {
				newLocation, err := program.SourceLocationAtInstructionAddress(d.programFile, stepInfo.InstructionAddress)
				if err != nil {
					log().Warn("failed to get source location for last executed instruction", logging.Address("address", stepInfo.InstructionAddress), slog.Any("error", err))
					break
				}

				lastSourceLocation := d.lastSourceLocation
				d.lastSourceLocation = newLocation
				debugEvent.Result.SourceLocation = newLocation

				if lastSourceLocation == nil || *lastSourceLocation != *newLocation {
					d.handleSourceLocationChange(debugEvent.Result)
				}
			}
		}

	case runtime.RunnerEventRuntimeError:
		if err, ok := event.Data.(error); ok {
			debugEvent.Event = EventError
			debugEvent.Result.StopReason = StopError
			debugEvent.Result.Error = err
		}

	case runtime.RunnerEventInterrupted:
		debugEvent.Event = EventInterrupted
		debugEvent.Result.StopReason = StopInterrupt

	case runtime.RunnerEventStopped:
		debugEvent.Event = EventProgramHalted
		debugEvent.Result.StopReason = StopHalt

	default:
		d.Log().Warn("unknown runner event type", slog.String("type", event.Type.String()))
		return
	}

	//log().Debug("runner event", slog.String("runner_event", spew.Sdump(event)), slog.String("debug_event", spew.Sdump(debugEvent)))

	d.lastResult = debugEvent.Result

	d.fireEvent(debugEvent)
}

func (d *debugger) Runtime() runtime.Runtime {
	return d.runtime
}

func (d *debugger) IsRunning() bool {
	return d.runner.State() == runtime.RunnerStateRunning
}

func (d *debugger) Interrupt() *ExecutionResult {
	result := <-d.runner.SendCommand(runtime.RunnerCommand{Type: runtime.RunnerCommandInterrupt})
	if result.Error != nil {
		return &ExecutionResult{
			StopReason: StopError,
			Error:      result.Error,
		}
	}

	return &ExecutionResult{
		StopReason: StopInterrupt,
	}
}

// Returns if the user has requested the debugger to interrupt execution.
func (d *debugger) IsInterrupted() bool {
	return d.runner.State() == runtime.RunnerStateInterrupted
}

// SetEventCallback sets the callback for execution events
func (d *debugger) SetEventCallback(callback EventCallback) {
	d.eventCallback = callback
}

// HasEventCallback returns true if an event callback is set (for debugging)
func (d *debugger) HasEventCallback() bool {
	return d.eventCallback != nil
}

// LastResult returns the result of the last execution operation
func (d *debugger) LastResult() *ExecutionResult {
	return d.lastResult
}

// GetRunnerState returns the current state of the underlying runner
func (d *debugger) GetRunnerState() runtime.RunnerState {
	return d.runner.State()
}

// --- Breakpoint Management ---

// AddBreakpoint adds a breakpoint at the given address
func (d *debugger) AddBreakpoint(addr uint32) (*Breakpoint, error) {
	if err := d.runtime.SetBreakpoint(addr); err != nil {
		return nil, fmt.Errorf("failed to add breakpoint at address 0x%X: %w", addr, err)
	}

	bp := &Breakpoint{
		ID:      d.nextBreakpointID,
		Address: addr,
		Enabled: true,
	}
	d.nextBreakpointID++
	d.breakpoints[bp.ID] = bp
	d.breakpointAddrs[addr] = bp

	if sourceLocation, err := program.SourceLocationAtInstructionAddress(d.programFile, addr); err == nil {
		d.Log().Debug("breakpoint added", logging.Address("address", addr), sourceLocation.LoggingAttribute("source_location"))
	}

	return bp, nil
}

// RemoveBreakpoint removes a breakpoint by ID
func (d *debugger) RemoveBreakpoint(id int) (*Breakpoint, error) {
	bp, exists := d.breakpoints[id]
	if !exists {
		return nil, fmt.Errorf("breakpoint with ID %d does not exist", id)
	}

	delete(d.breakpointAddrs, bp.Address)
	delete(d.breakpoints, id)

	if err := d.runtime.ClearBreakpoint(bp.Address); err != nil {
		return nil, fmt.Errorf("failed to remove breakpoint at address 0x%X: %w", bp.Address, err)
	}

	if sourceLocation, err := program.SourceLocationAtInstructionAddress(d.programFile, bp.Address); err == nil {
		d.Log().Debug("breakpoint removed", logging.Address("address", bp.Address), sourceLocation.LoggingAttribute("source_location"))
	}

	return bp, nil
}

// GetBreakpoint returns a breakpoint by ID
func (d *debugger) GetBreakpoint(id int) *Breakpoint {
	return d.breakpoints[id]
}

// GetBreakpointAt returns a breakpoint at the given address
func (d *debugger) GetBreakpointAt(addr uint32) *Breakpoint {
	return d.breakpointAddrs[addr]
}

// ListBreakpoints returns all breakpoints sorted by address
func (d *debugger) ListBreakpoints() []*Breakpoint {
	bps := make([]*Breakpoint, 0, len(d.breakpoints))
	for _, bp := range d.breakpoints {
		bps = append(bps, bp)
	}
	sort.Slice(bps, func(i, j int) bool {
		return bps[i].Address < bps[j].Address
	})
	return bps
}

// EnableBreakpoint enables or disables a breakpoint
func (d *debugger) EnableBreakpoint(id int, enabled bool) bool {
	bp, exists := d.breakpoints[id]
	if !exists {
		return false
	}
	bp.Enabled = enabled
	return true
}

// ClearBreakpoints removes all breakpoints
func (d *debugger) ClearBreakpoints() {
	d.breakpoints = make(map[int]*Breakpoint)
	d.breakpointAddrs = make(map[uint32]*Breakpoint)
}

// --- Watchpoint Management ---

// AddWatchpoint adds a memory watchpoint
func (d *debugger) AddWatchpoint(r *memory.Range, wpType WatchpointType) (*Watchpoint, error) {
	if r.Size > 4 {
		return nil, fmt.Errorf("watchpoint size too large: %d bytes (max 4)", r.Size)
	}

	var lastValue []byte
	view := memory.NewSlice(d.runtime.Memory(), r)
	if err := view.ReadInto(lastValue); err != nil {
		return nil, fmt.Errorf("failed to read initial value for watchpoint at address range 0x%X-0x%X: %w", r.Start, r.End(), err)
	}

	wp := &Watchpoint{
		ID:        d.nextWatchpointID,
		Memory:    r,
		Type:      wpType,
		Enabled:   true,
		LastValue: lastValue,
	}
	d.nextWatchpointID++
	d.watchpoints[wp.ID] = wp
	return wp, nil
}

// RemoveWatchpoint removes a watchpoint by ID
func (d *debugger) RemoveWatchpoint(id int) (*Watchpoint, error) {
	wp, exists := d.watchpoints[id]
	if !exists {
		return nil, fmt.Errorf("watchpoint with ID %d does not exist", id)
	}
	delete(d.watchpoints, id)

	if err := d.runtime.ClearWatchpoint(*wp.Memory); err != nil {
		return nil, fmt.Errorf("failed to clear watchpoint at address range 0x%X-0x%X: %w", wp.Memory.Start, wp.Memory.End(), err)
	}

	return wp, nil
}

// GetWatchpoint returns a watchpoint by ID
func (d *debugger) GetWatchpoint(id int) *Watchpoint {
	return d.watchpoints[id]
}

// ListWatchpoints returns all watchpoints sorted by address
func (d *debugger) ListWatchpoints() []*Watchpoint {
	wps := make([]*Watchpoint, 0, len(d.watchpoints))
	for _, wp := range d.watchpoints {
		wps = append(wps, wp)
	}
	sort.Slice(wps, func(i, j int) bool {
		return wps[i].Memory.Start < wps[j].Memory.Start
	})
	return wps
}

// ClearWatchpoints removes all watchpoints
func (d *debugger) ClearWatchpoints() {
	d.watchpoints = make(map[int]*Watchpoint)
}

// --- Termination Addresses ---

// AddTerminationAddress adds an address that will cause execution to stop
func (d *debugger) AddTerminationAddress(addr uint32) {
	d.terminationAddrs[addr] = true
}

// RemoveTerminationAddress removes a termination address
func (d *debugger) RemoveTerminationAddress(addr uint32) {
	delete(d.terminationAddrs, addr)
}

// IsTerminationAddress checks if an address is a termination address
func (d *debugger) IsTerminationAddress(addr uint32) bool {
	return d.terminationAddrs[addr]
}

// ClearTerminationAddresses removes all termination addresses
func (d *debugger) ClearTerminationAddresses() {
	d.terminationAddrs = make(map[uint32]bool)
}

// --- Execution Control ---

func (d *debugger) executionError(err error) *ExecutionResult {
	result := &ExecutionResult{
		StopReason: StopError,
		Error:      err,
	}
	d.lastResult = result
	return result
}

func (d *debugger) reportExecutionError(err error) *ExecutionResult {
	result := d.executionError(err)
	d.fireEvent(&Event{Event: EventError, Result: result})
	return result
}

func (d *debugger) readPC() (uint32, error) {
	return cpu.ReadPC(d.runtime.CPU().Registers())
}

// Step executes a single instruction
func (d *debugger) Step() *ExecutionResult {
	pc, err := d.readPC()
	if err != nil {
		return d.reportExecutionError(fmt.Errorf("failed to read PC register before Step: %w", err))
	}

	bp, err := d.AddBreakpoint(pc + 4)
	if err != nil {
		return d.reportExecutionError(fmt.Errorf("failed to add temporary breakpoint at 0x%X: %w", pc+4, err))
	}

	// Subscribe to stepping event to remove temporary breakpoint after stepping
	d.eventsQueue.SubscribeOnce(EventStepped, NewEventHandler(func(event *Event) bool {
		// On step event, remove temporary breakpoint and stop execution
		d.RemoveBreakpoint(bp.ID)
		return false
	}))

	return d.Continue()
}

// Continue executes until a stop condition is met
func (d *debugger) Continue() *ExecutionResult {
	result := <-d.runner.SendCommand(runtime.RunnerCommand{
		Type: runtime.RunnerCommandContinue,
	})

	if result.Error != nil {
		return d.reportExecutionError(fmt.Errorf("failed to continue execution: %w", result.Error))
	}

	return &ExecutionResult{
		StopReason: StopNone,
	}
}

// RunUntil executes until the PC reaches the target address
func (d *debugger) RunUntil(targetAddr uint32) *ExecutionResult {
	// Add temporary breakpoint
	bp, err := d.AddBreakpoint(targetAddr)
	if err != nil {
		return d.reportExecutionError(fmt.Errorf("failed to add temporary breakpoint at 0x%X: %w", targetAddr, err))
	}

	d.eventsQueue.Subscribe(EventBreakpointHit, NewEventHandler(func(event *Event) bool {
		if event.Result.Breakpoint.ID == bp.ID {
			// On hitting our temporary breakpoint, remove it and stop execution
			d.RemoveBreakpoint(bp.ID)
			return false
		}
		return true
	}))

	return d.Continue()
}

// StepOver executes one instruction, stepping over function calls
func (d *debugger) StepOver() *ExecutionResult {
	pc, err := cpu.ReadPC(d.runtime.CPU().Registers())
	if err != nil {
		return d.reportExecutionError(fmt.Errorf("failed to read PC register before StepOver: %w", err))
	}

	isCall, err := d.isCallInstruction(pc)
	if err != nil {
		return d.reportExecutionError(fmt.Errorf("failed to determine if instruction at 0x%X is a call: %w", pc, err))
	}

	// Check if the current instruction is a call
	if isCall {
		// Set temporary breakpoint at return address (PC + 4)
		returnAddr := pc + 4

		// Add temporary breakpoint
		bp, err := d.AddBreakpoint(returnAddr)
		if err != nil {
			return d.reportExecutionError(fmt.Errorf("failed to add temporary breakpoint at return address 0x%X: %w", returnAddr, err))
		}

		// Continue execution
		result := d.Continue()

		if result.StopReason == StopNone {
			panic("Continue returned with no stop reason???")
		}

		// Remove temporary breakpoint
		d.RemoveBreakpoint(bp.ID)

		// Abort if continue stopped due to error
		if result.StopReason == StopError {
			return result
		}

		pc, err = cpu.ReadPC(d.runtime.CPU().Registers())
		if err != nil {
			return d.reportExecutionError(fmt.Errorf("failed to read PC register after StepOver: %w", err))
		}

		// If we stopped at our temporary breakpoint, report as step
		if result.StopReason == StopBreakpoint && pc == returnAddr {
			result.StopReason = StopStep
		}

		return result
	}

	// Not a call, just do a regular step
	return d.Step()
}

// For a given instruction address, returns all possible next instruction addresses (e.g., branches or fallthrough)
func (d *debugger) nextInstructionAddresses(addr uint32) ([]uint32, error) {
	log := d.Log().Child("nextInstructionAddresses").WithAttrs(logging.Address("address", addr))

	instr, err := cpu.DecodeInstruction(d.runtime.Memory(), addr)
	if err != nil {
		return nil, log.Errorf("failed to decode instruction at 0x%X: %w", addr, err)
	}

	log = log.WithAttrs(instr.LoggingAttribute("instruction"))

	var nextAddrs []uint32
	switch instr.Descriptor.OpCode.OpCode {
	case instructions.OpCode_JMP:
		if len(instr.OperandValues) < 1 {
			log.Panic("JMP instruction at 0x%X missing operand???", addr)
		}

		operandValue := instr.OperandValues[0]
		if operandValue.Kind() != instructions.OperandKind_Register {
			log.Panic("JMP instruction at 0x%X has non-register operand???", addr)
		}

		targetReg := operandValue.Register()
		if targetReg == nil {
			log.Panic("JMP instruction at 0x%X has nil register operand???", addr)
		}

		targetAddr, err := d.runtime.CPU().Registers().ReadByDescriptor(targetReg)
		if err != nil {
			return nil, log.Errorf("failed to read jump target register %s for JMP instruction at 0x%X: %w", targetReg.Name(), addr, err)
		}

		log.Debug("unconditional jump target decoded", logging.Address("target_address", targetAddr))

		nextAddrs = append(nextAddrs, targetAddr)

	case instructions.OpCode_CJMP:
		if len(instr.OperandValues) < 2 {
			log.Panic("CJMP instruction at 0x%X missing operands???", addr)
		}

		operandValue := instr.OperandValues[1]
		if operandValue.Kind() != instructions.OperandKind_Register {
			log.Panic("CJMP instruction at 0x%X has non-register jump target operand???", addr)
		}

		targetReg := operandValue.Register()
		if targetReg == nil {
			log.Panic("CJMP instruction at 0x%X has nil register jump target operand???", addr)
		}

		targetAddr, err := d.runtime.CPU().Registers().ReadByDescriptor(targetReg)
		if err != nil {
			return nil, log.Errorf("failed to read jump target register %s for CJMP instruction at 0x%X: %w", targetReg.Name(), addr, err)
		}

		nextAddrs = append(nextAddrs, targetAddr)

		log.Debug("conditional jump target decoded", logging.Address("target_address", targetAddr))

		// CJMP can also fall through to the next instruction
		fallthrough
	default:
		// Add fallthrough address for non-jump instructions and CJMP (which can also fall through)
		nextAddrs = append(nextAddrs, addr+4)
		log.Debug("fallthrough to next instruction", logging.Address("next_address", addr+4))
	}

	return nextAddrs, nil
}

// Stores a graph of source instructions addresses -> jump target instruction addresses for all instructions within a code region.
//
// Sources are only tracked within the code region covered by the graph, but targets can be outside the region.
type ControlFlowGraph struct {
	sourceToTarget  map[uint32]uint32
	targetToSources map[uint32][]uint32
}

// Returns the branch target address for a given source instruction address, or nil if the instruction at that address is not a branch or is outside the graph's code region.
func (c *ControlFlowGraph) Target(addr uint32) *uint32 {
	if target, exists := c.sourceToTarget[addr]; exists {
		return &target
	}
	return nil
}

// Returns the source instruction addresses that can branch to the given target address, or nil if there are no such instructions or the target is outside the graph's code region.
func (c *ControlFlowGraph) Sources(addr uint32) []uint32 {
	if sources, exists := c.targetToSources[addr]; exists {
		return sources
	}
	return nil
}

// Returns all source instruction addresses in the graph's code region that have a branch target within the graph's code region.
func (c *ControlFlowGraph) AllSources() iter.Seq[uint32] {
	return maps.Keys(c.sourceToTarget)
}

// Returns all target instruction addresses in the graph that are branched to from a source instruction within the graph's code region.
func (c *ControlFlowGraph) AllTargets() iter.Seq[uint32] {
	return maps.Keys(c.targetToSources)
}

// Returns a control flow graph of the instructions within a given code region
func (d *debugger) BuildControlFlowGraph(region *memory.Range) (*ControlFlowGraph, error) {
	log := d.Log().Child("BuildControlFlowGraph").WithAttrs(region.AddressRangeLoggingAttribute("region"))

	sourceToTarget := make(map[uint32]uint32)
	targetToSources := make(map[uint32][]uint32)

	for addr := range region.Addresses(4) {
		targets, err := d.nextInstructionAddresses(addr)
		if err != nil {
			return nil, log.Errorf("failed to get next instruction addresses for instruction at 0x%X: %v", addr, err)
		}

		for _, target := range targets {
			sourceToTarget[addr] = target
			targetToSources[target] = append(targetToSources[target], addr)

			log.Debug("edge", logging.Address("source", addr), logging.Address("target", target))
		}
	}

	return &ControlFlowGraph{
		sourceToTarget:  sourceToTarget,
		targetToSources: targetToSources,
	}, nil
}

// For a given instruction address, returns the source code locations of all possible next instructions (e.g., branches or fallthrough)
func (d *debugger) nextSourceLocations(addr uint32) ([]*sourcecode.Location, error) {
	log := d.Log().Child("nextSourceLocations").WithAttrs(logging.Address("address", addr))

	originalSourceLocation, err := program.SourceLocationAtInstructionAddress(d.programFile, addr)
	if err != nil {
		return nil, log.Errorf("failed to get source location for instruction at 0x%X: %w", addr, err)
	}

	log = log.WithAttrs(originalSourceLocation.LoggingAttribute("original_source_location"))

	addresses, err := d.nextInstructionAddresses(addr)
	if err != nil {
		return nil, log.Errorf("failed to get next instruction addresses for instruction at 0x%X: %w", addr, err)
	}

	locations, err := utils.MapMayFail(addresses, func(nextAddr uint32) (*sourcecode.Location, error) {
		loc, err := program.SourceLocationAtInstructionAddress(d.programFile, nextAddr)
		if err != nil {
			return nil, log.Errorf("failed to get source location for next instruction address 0x%X: %w", nextAddr, err)
		}

		log.Debug("possible target source location", logging.Address("next_instruction_address", nextAddr), loc.LoggingAttribute("target_source_location"))
		return loc, nil
	})

	if err != nil {
		return nil, err
	}

	return utils.Filter(utils.Set(locations), func(loc *sourcecode.Location) bool {
		// Filter out locations that are the same as the original instruction's source location, since we only want to stop at a new source location.
		return *loc != *originalSourceLocation
	}), nil
}

// For a given instruction address, returns all the possible next source code locations (e.g., branches or fallthrough) first instruction addresses.
// This is used for determining where to stop when stepping through source code lines.
func (d *debugger) nextSourceLocationInstructionAddresses(addr uint32) ([]uint32, error) {
	log := d.Log().Child("nextSourceLocationInstructionAddresses").WithAttrs(logging.Address("address", addr))

	locations, err := d.nextSourceLocations(addr)
	if err != nil {
		return nil, log.Errorf("failed to get next source locations for instruction at 0x%X: %w", addr, err)
	}

	return utils.MapMayFail(locations, func(loc *sourcecode.Location) (uint32, error) {
		if addr, err := program.InstructionAddressAtSourceLocation(d.programFile, loc); err != nil {
			return 0, log.Errorf("failed to get instruction address for source location %s:%d:%d: %w", loc.File.Name(), loc.Line, loc.Column, err)
		} else {
			return addr, nil
		}
	})
}

// isCallInstruction checks if the instruction at addr is a function call
// A branch is a call if the branch target is a function symbol
func (d *debugger) isCallInstruction(addr uint32) (bool, error) {
	log := d.Log().Child("isCallInstruction").WithAttrs(logging.Address("address", addr))

	nextAddrs, err := d.nextInstructionAddresses(addr)
	if err != nil {
		return false, log.Errorf("failed to get next instruction addresses for instruction at 0x%X: %w", addr, err)
	}

	if len(nextAddrs) <= 0 {
		return false, log.Errorf("no next instruction addresses found for instruction at 0x%X", addr)
	}

	// If you look at nextInstructionAddresses(), the target address of a jump instruction (either JMP or CJMP) will always be the first address in the returned list. So we can just check if that address is a function symbol to determine if this is a call instruction.
	targetAddr := nextAddrs[0]

	function, err := program.FunctionAtAddress(d.programFile, targetAddr)
	if err != nil {
		log.Debug("no function symbol found at jump target address, not a call instruction", logging.Address("target_address", targetAddr))
		return false, nil
	}

	log.Debug("function symbol found at jump target address, this is a call instruction", logging.Address("target_address", targetAddr), slog.String("function_name", function.Name))
	return true, nil
}

// StepOut executes until returning from the current function
// For now, this runs until LR is reached (simplified implementation)
func (d *debugger) StepOut() *ExecutionResult {
	lr, err := cpu.ReadLR(d.runtime.CPU().Registers())
	if err != nil {
		return d.reportExecutionError(fmt.Errorf("failed to read LR register for StepOut: %w", err))
	}

	return d.RunUntil(lr)
}

func (d *debugger) fireEvent(event *Event) bool {
	doContinue := d.eventsQueue.EventFired(event)

	if d.eventCallback != nil {
		doContinue = doContinue && d.eventCallback(event)
	}

	return doContinue
}

func (d *debugger) checkWatchpoints() (*Watchpoint, error) {
	for _, wp := range d.watchpoints {
		if !wp.Enabled {
			continue
		}

		if wp.Memory.Size > 4 {
			return nil, fmt.Errorf("watchpoint size too large: %d bytes (max 4)", wp.Memory.Size)
		}

		currentValue, err := memory.NewSlice(d.runtime.Memory(), wp.Memory).ReadAll()
		if err != nil {
			return nil, fmt.Errorf("failed to read watchpoint memory: %w", err)
		}

		if !slices.Equal(currentValue, wp.LastValue) && (wp.Type&WatchRead) != 0 {
			wp.HitCount++
			wp.LastValue = currentValue
			return wp, nil
		}
	}

	return nil, nil
}

func (d *debugger) CurrentSourceLocation() (*sourcecode.Location, error) {
	pc, err := cpu.ReadPC(d.runtime.CPU().Registers())
	if err != nil {
		return nil, fmt.Errorf("failed to read PC register: %w", err)
	}

	return program.SourceLocationAtInstructionAddress(d.programFile, pc)
}

func (d *debugger) handleSourceLocationChange(result *ExecutionResult) {
	d.fireEvent(&Event{Event: EventSourceLocationChanged, Result: result})

	d.Log().Debug("source location changed", result.SourceLocation.LoggingAttribute("new_source_location"), logging.Address("pc", result.LastPC))
}

func (d *debugger) StepIntoSource() *ExecutionResult {
	log := d.Log().Child("StepIntoSource")

	pc, err := d.readPC()
	if err != nil {
		return d.reportExecutionError(log.Errorf("failed to read PC register before StepIntoSource: %w", err))
	}

	log = log.WithAttrs(logging.Address("starting_instruction_address", pc))

	var targetAddrs []uint32

	for {
		var err error
		targetAddrs, err = d.nextSourceLocationInstructionAddresses(pc)
		if err != nil {
			return d.reportExecutionError(log.Errorf("failed to get next source location instruction addresses before StepIntoSource: %w", err))
		}

		if len(targetAddrs) == 0 {
			log.Warn("no next source location instruction addresses found after instruction, trying next instruction", logging.Address("instruction_address", pc), logging.Address("next_instruction_address", pc+4))
			pc += 4
			continue
		} else {
			break
		}
	}

	log.Debug("next source location instruction addresses", logging.Addresses("addresses", targetAddrs))

	// Add temporary breakpoints at all possible next source location instruction addresses
	var breakpoints map[int]*Breakpoint = make(map[int]*Breakpoint)
	for _, targetAddr := range targetAddrs {
		bp, err := d.AddBreakpoint(targetAddr)
		if err != nil {
			// Clean up any breakpoints we added before returning error
			for id := range breakpoints {
				d.RemoveBreakpoint(id)
			}
			return d.reportExecutionError(log.Errorf("failed to add temporary breakpoint at 0x%X: %w", targetAddr, err))
		}
		breakpoints[bp.ID] = bp

		d.eventsQueue.SubscribeOnce(EventBreakpointHit, NewEventHandler(func(event *Event) bool {
			if event.Result.Breakpoint.ID == bp.ID {
				// On hitting one of our temporary breakpoints, remove all of them and stop execution
				// If go closures capture by value, then we have a nice bug here
				for id := range breakpoints {
					d.RemoveBreakpoint(id)
				}
				return false
			}
			return true
		}))
	}

	// Continue execution

	return d.Continue()
}

func (d *debugger) StepOverSource() *ExecutionResult {
	log := d.Log().Child("StepOverSource")

	currentSourceLocation, err := d.CurrentSourceLocation()
	if err != nil {
		return d.reportExecutionError(log.Errorf("failed to get current source location before StepOverSource: %w", err))
	}

	instructionRanges, err := program.InstructionAddressesAtSourceLocation(d.programFile, currentSourceLocation)
	if err != nil {
		return d.reportExecutionError(log.Errorf("failed to get instruction addresses at source location before StepOverSource: %w", err))
	}

	if len(instructionRanges) == 0 {
		return d.reportExecutionError(log.Errorf("no instruction addresses found at current source location before StepOverSource"))
	}

	for _, instrRange := range instructionRanges {
		for addr := instrRange.Start; addr < instrRange.End(); addr += 4 {
			if isCall, err := d.isCallInstruction(addr); err != nil {
				return d.reportExecutionError(log.Errorf("failed to determine if instruction at 0x%X is a call: %w", addr, err))
			} else if isCall {
				// Set temporary breakpoint at return address (addr + 4)
				returnAddr := addr + 4

				// Add temporary breakpoint
				bp, err := d.AddBreakpoint(returnAddr)
				if err != nil {
					return d.reportExecutionError(log.Errorf("failed to add temporary breakpoint at return address 0x%X: %w", returnAddr, err))
				}

				// Continue execution
				result := d.Continue()

				if result.StopReason == StopNone {
					log.Panic("Continue returned with no stop reason???")
				}

				// Remove temporary breakpoint
				d.RemoveBreakpoint(bp.ID)

				// Abort if continue stopped due to error
				if result.StopReason == StopError {
					return result
				}

				pc, err := cpu.ReadPC(d.runtime.CPU().Registers())
				if err != nil {
					return d.reportExecutionError(log.Errorf("failed to read PC register after StepOverSource: %w", err))
				}

				// If we stopped at our temporary breakpoint, report as step
				if result.StopReason == StopBreakpoint && pc == returnAddr {
					result.StopReason = StopStep
				}

				return result
			}
		}
	}

	// No calls found in current source line, just do a regular step
	return d.StepIntoSource()
}

func (d *debugger) Program() program.ProgramFile {
	return d.programFile
}

func (d *debugger) Reset() *ExecutionResult {
	result := <-d.runner.SendCommand(runtime.RunnerCommand{Type: runtime.RunnerCommandReset})
	if result.Error != nil {
		return &ExecutionResult{
			StopReason: StopError,
			Error:      result.Error,
		}
	}

	// Clear breakpoints and watchpoints after reset
	return &ExecutionResult{
		StopReason: StopHalt,
	}
}
