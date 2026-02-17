// Package interpreter provides debugging and step-by-step execution support
// for the Cucaracha CPU interpreter.
package core

import (
	"fmt"
	"slices"
	"sort"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/Manu343726/cucaracha/pkg/hw/memory"
	"github.com/Manu343726/cucaracha/pkg/runtime"
	"github.com/Manu343726/cucaracha/pkg/runtime/program"
	"github.com/Manu343726/cucaracha/pkg/runtime/program/sourcecode"
)

// debugger provides debugging capabilities for the interpreter
type debugger struct {
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
	lastResult *ExecutionResult
}

// NewDebugger creates a new debugger for the given runtime
func NewDebugger(r runtime.Runtime, program program.ProgramFile) Debugger {
	runner := runtime.NewRunner(r)

	// We use the runtime adapter from the runner instead
	// of the original runtime to ensure that all interactions with the debugged
	// runtime go through the runner and are properly synchronized.
	runtime := runner.Runtime()

	return &debugger{
		runtime:          runtime,
		runner:           runner,
		programFile:      program,
		breakpoints:      make(map[int]*Breakpoint),
		breakpointAddrs:  make(map[uint32]*Breakpoint),
		watchpoints:      make(map[int]*Watchpoint),
		terminationAddrs: make(map[uint32]bool),
		eventsQueue:      NewInternalEventsQueue(),
	}
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

// isCallInstruction checks if the instruction at addr is a function call
// A branch is a call if the branch target is a function symbol
func (d *debugger) isCallInstruction(addr uint32) (bool, error) {
	instr, err := cpu.DecodeInstruction(d.runtime.Memory(), addr)
	if err != nil {
		return false, fmt.Errorf("failed to decode instruction at 0x%X: %w", addr, err)
	}

	callOpcodes := map[instructions.OpCode]struct{}{
		instructions.OpCode_JMP:  {},
		instructions.OpCode_CJMP: {},
	}

	if _, isCall := callOpcodes[instr.Descriptor.OpCode.OpCode]; !isCall {
		return false, nil
	}

	var operandValue instructions.OperandValue

	switch instr.Descriptor.OpCode.OpCode {
	case instructions.OpCode_JMP:
		if len(instr.OperandValues) < 1 {
			panic("JMP instruction missing operand???")
		}

		operandValue = instr.OperandValues[0]
	case instructions.OpCode_CJMP:
		if len(instr.OperandValues) < 2 {
			panic("CJMP instruction missing operands???")
		}

		operandValue = instr.OperandValues[1]
	}

	if operandValue.Kind() != instructions.OperandKind_Register {
		panic("branch target operand is not a register???")
	}

	targetReg := operandValue.Register()
	if targetReg == nil {
		panic("branch target operand register is nil???")
	}

	targetAddress, err := d.runtime.CPU().Registers().ReadByDescriptor(targetReg)
	if err != nil {
		return false, fmt.Errorf("failed to read branch target register %s: %w", targetReg.Name(), err)
	}

	function, err := program.FunctionAtAddress(d.programFile, targetAddress)
	if err != nil {
		return false, fmt.Errorf("failed to get instruction at branch target address 0x%X: %w", targetAddress, err)
	}

	return function != nil, nil
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
}

func (d *debugger) StepIntoSource() *ExecutionResult {
	sourceCodeChangeHandler := NewEventHandler(func(event *Event) bool {
		// On source line change, stop execution and report as step
		event.Result.StopReason = StopStep
		return false
	})

	d.eventsQueue.Subscribe(EventSourceLocationChanged, sourceCodeChangeHandler)
	defer d.eventsQueue.Unsubscribe(sourceCodeChangeHandler)

	return d.Continue()
}

func (d *debugger) StepOverSource() *ExecutionResult {
	currentSourceLocation, err := d.CurrentSourceLocation()
	if err != nil {
		return d.reportExecutionError(fmt.Errorf("failed to get current source location before StepOverSource: %w", err))
	}

	instructionRanges, err := program.InstructionAddressesAtSourceLocation(d.programFile, currentSourceLocation)
	if err != nil {
		return d.reportExecutionError(fmt.Errorf("failed to get instruction addresses at source location before StepOverSource: %w", err))
	}

	if len(instructionRanges) == 0 {
		return d.reportExecutionError(fmt.Errorf("no instruction addresses found at current source location before StepOverSource"))
	}

	for _, instrRange := range instructionRanges {
		for addr := instrRange.Start; addr < instrRange.End(); addr += 4 {
			if isCall, err := d.isCallInstruction(addr); err != nil {
				return d.reportExecutionError(fmt.Errorf("failed to determine if instruction at 0x%X is a call: %w", addr, err))
			} else if isCall {
				// Set temporary breakpoint at return address (addr + 4)
				returnAddr := addr + 4

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

				pc, err := cpu.ReadPC(d.runtime.CPU().Registers())
				if err != nil {
					return d.reportExecutionError(fmt.Errorf("failed to read PC register after StepOverSource: %w", err))
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
