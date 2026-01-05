// Package interpreter provides debugging and step-by-step execution support
// for the Cucaracha CPU interpreter.
package debugger

import (
	"fmt"
	goruntime "runtime"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/interpreter"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/Manu343726/cucaracha/pkg/hw/memory"
	"github.com/Manu343726/cucaracha/pkg/runtime"
	"github.com/Manu343726/cucaracha/pkg/utils"
)

// Debugger provides debugging capabilities for the interpreter
type Debugger struct {
	// execution runtime that is being debugged
	runtime runtime.Runtime

	// Program being debugged
	programFile mc.ProgramFile

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

	// Execution state
	lastResult *ExecutionResult

	// Interrupt flag (atomic for thread-safety)
	interrupted int32

	// Target execution speed
	targetSpeedHz float64
}

// NewDebugger creates a new debugger for the given interpreter
func NewDebugger(runtime runtime.Runtime, program mc.ProgramFile) *Debugger {
	return &Debugger{
		runtime:          runtime,
		programFile:      program,
		breakpoints:      make(map[int]*Breakpoint),
		breakpointAddrs:  make(map[uint32]*Breakpoint),
		watchpoints:      make(map[int]*Watchpoint),
		terminationAddrs: make(map[uint32]bool),
	}
}

func (d *Debugger) ProgramFile() mc.ProgramFile {
	return d.programFile
}

// LoadProgram loads a program into the debugger's runtime
func (d *Debugger) LoadProgram(prog *mc.Program) error {
	if err := d.runtime.LoadProgram(prog); err != nil {
		return err
	}

	d.program = prog
	return nil
}

// Interrupt signals the debugger to stop execution.
// This is safe to call from signal handlers or other goroutines.
func (d *Debugger) Interrupt() {
	atomic.StoreInt32(&d.interrupted, 1)
}

// ClearInterrupt clears the interrupt flag.
func (d *Debugger) ClearInterrupt() {
	atomic.StoreInt32(&d.interrupted, 0)
}

// IsInterrupted returns true if the interrupt flag is set.
func (d *Debugger) IsInterrupted() bool {
	return atomic.LoadInt32(&d.interrupted) != 0
}

// SetTargetSpeed sets the target execution speed in Hz (cycles per second).
// Use 0 for unlimited speed (no timing simulation).
// This delegates to the underlying Interpreter.
func (d *Debugger) SetTargetSpeed(hz float64) {
	d.targetSpeedHz = hz
}

// GetTargetSpeed returns the current target execution speed in Hz.
func (d *Debugger) GetTargetSpeed() float64 {
	return d.targetSpeedHz
}

// Runtime returns the underlying runtime
func (d *Debugger) Runtime() runtime.Runtime {
	return d.runtime
}

// SetEventCallback sets the callback for execution events
func (d *Debugger) SetEventCallback(callback EventCallback) {
	d.eventCallback = callback
}

// HasEventCallback returns true if an event callback is set (for debugging)
func (d *Debugger) HasEventCallback() bool {
	return d.eventCallback != nil
}

// LastResult returns the result of the last execution operation
func (d *Debugger) LastResult() *ExecutionResult {
	return d.lastResult
}

// --- Breakpoint Management ---

// AddBreakpoint adds a breakpoint at the given address
func (d *Debugger) AddBreakpoint(addr uint32) *Breakpoint {
	bp := &Breakpoint{
		ID:      d.nextBreakpointID,
		Address: addr,
		Enabled: true,
	}
	d.nextBreakpointID++
	d.breakpoints[bp.ID] = bp
	d.breakpointAddrs[addr] = bp
	return bp
}

// RemoveBreakpoint removes a breakpoint by ID
func (d *Debugger) RemoveBreakpoint(id int) bool {
	bp, exists := d.breakpoints[id]
	if !exists {
		return false
	}
	delete(d.breakpointAddrs, bp.Address)
	delete(d.breakpoints, id)
	return true
}

// GetBreakpoint returns a breakpoint by ID
func (d *Debugger) GetBreakpoint(id int) *Breakpoint {
	return d.breakpoints[id]
}

// GetBreakpointAt returns a breakpoint at the given address
func (d *Debugger) GetBreakpointAt(addr uint32) *Breakpoint {
	return d.breakpointAddrs[addr]
}

// ListBreakpoints returns all breakpoints sorted by address
func (d *Debugger) ListBreakpoints() []*Breakpoint {
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
func (d *Debugger) EnableBreakpoint(id int, enabled bool) bool {
	bp, exists := d.breakpoints[id]
	if !exists {
		return false
	}
	bp.Enabled = enabled
	return true
}

// ClearBreakpoints removes all breakpoints
func (d *Debugger) ClearBreakpoints() {
	d.breakpoints = make(map[int]*Breakpoint)
	d.breakpointAddrs = make(map[uint32]*Breakpoint)
}

// --- Watchpoint Management ---

// AddWatchpoint adds a memory watchpoint
func (d *Debugger) AddWatchpoint(r memory.Range, wpType WatchpointType) (*Watchpoint, error) {
	if r.Size > 4 {
		return nil, fmt.Errorf("watchpoint size too large: %d bytes (max 4)", r.Size)
	}

	// Read initial value
	lastValue, err := memory.NewSlice(d.runtime.Memory(), r).ReadAsUint32()
	if err != nil {
		return nil, fmt.Errorf("failed to read initial watchpoint value: %w", err)
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
func (d *Debugger) RemoveWatchpoint(id int) bool {
	_, exists := d.watchpoints[id]
	if !exists {
		return false
	}
	delete(d.watchpoints, id)
	return true
}

// GetWatchpoint returns a watchpoint by ID
func (d *Debugger) GetWatchpoint(id int) *Watchpoint {
	return d.watchpoints[id]
}

// ListWatchpoints returns all watchpoints sorted by address
func (d *Debugger) ListWatchpoints() []*Watchpoint {
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
func (d *Debugger) ClearWatchpoints() {
	d.watchpoints = make(map[int]*Watchpoint)
}

// --- Termination Addresses ---

// AddTerminationAddress adds an address that will cause execution to stop
func (d *Debugger) AddTerminationAddress(addr uint32) {
	d.terminationAddrs[addr] = true
}

// RemoveTerminationAddress removes a termination address
func (d *Debugger) RemoveTerminationAddress(addr uint32) {
	delete(d.terminationAddrs, addr)
}

// IsTerminationAddress checks if an address is a termination address
func (d *Debugger) IsTerminationAddress(addr uint32) bool {
	return d.terminationAddrs[addr]
}

// ClearTerminationAddresses removes all termination addresses
func (d *Debugger) ClearTerminationAddresses() {
	d.terminationAddrs = make(map[uint32]bool)
}

// --- Execution Control ---

// Step executes a single instruction
func (d *Debugger) Step() (*ExecutionResult, error) {
	pc, err := d.runtime.CPU().Registers().ReadByDescriptor(mc.Descriptor.Registers.PC)
	if err != nil {
		return nil, fmt.Errorf("failed to read PC register before debugger step: %w", err)
	}

	result := &ExecutionResult{
		LastPC: pc,
	}

	// Check for termination address before executing
	if d.terminationAddrs[pc] {
		result.StopReason = utils.Ptr(StopTermination)
		d.lastResult = result
		d.fireEvent(EventProgramTerminated, result)
		return result, nil
	}

	// Check for breakpoint (after first step, stepping over breakpoint)
	if bp := d.breakpointAddrs[pc]; bp != nil && bp.Enabled {
		bp.HitCount++
		result.StopReason = utils.Ptr(StopBreakpoint)
		result.BreakpointID = bp.ID
		d.lastResult = result
		d.fireEvent(EventBreakpointHit, result)
		return result, nil
	}

	// Check if halted
	if d.runtime.CPU().IsHalted() {
		result.StopReason = utils.Ptr(StopHalt)
		d.lastResult = result
		d.fireEvent(EventProgramHalted, result)
		return result, nil
	}

	// Execute simulation step
	stepResult, err := d.runtime.Step()

	if err != nil {
		result.StopReason = utils.Ptr(StopError)
		result.Error = err
		d.lastResult = result
		d.fireEvent(EventError, result)
		return result, nil
	}

	// Track cycles
	if stepResult != nil {
		result.CyclesExecuted = int64(stepResult.CyclesUsed)
	}

	// Check watchpoints after execution
	if wp, err := d.checkWatchpoints(); err != nil {
		result.StopReason = utils.Ptr(StopError)
		result.Error = err
		d.lastResult = result
		d.fireEvent(EventError, result)
		return result, nil
	} else if wp != nil {
		result.StopReason = utils.Ptr(StopWatchpoint)
		result.WatchpointID = wp.ID
		d.lastResult = result
		d.fireEvent(EventWatchpointHit, result)
		return result, nil
	}

	result.StopReason = nil
	d.lastResult = result
	d.fireEvent(EventStepped, result)
	return result, nil
}

// Continue executes until a stop condition is met
func (d *Debugger) Continue() (*ExecutionResult, error) {
	return d.Run(0)
}

// Run executes up to maxSteps instructions (0 = unlimited)
func (d *Debugger) Run(maxSteps int) (*ExecutionResult, error) {
	startTime := time.Now()
	var totalCycles int64
	totalSteps := 0

	for {
		result, err := d.Step()
		if err != nil {
			return result, err
		}

		totalSteps++
		totalCycles += result.CyclesExecuted

		// Yield to allow signal handlers and other goroutines to run
		// This is important for Ctrl+C handling on Windows
		if totalSteps%100 == 0 {
			goruntime.Gosched()
		}

		if result.StopReason != nil {
			result.StepsExecuted = totalSteps
			result.CyclesExecuted = totalCycles
			d.lastResult = result
			return result, nil
		}

		if totalSteps == maxSteps {
			result.StepsExecuted = totalSteps
			result.CyclesExecuted = totalCycles
			result.StopReason = utils.Ptr(StopMaxSteps)
			d.lastResult = result
			return result, nil
		}

		// Apply speed control: compute dynamic delay to match target Hz
		if d.targetSpeedHz > 0 {
			elapsed := time.Since(startTime)
			// Calculate expected time based on cycles executed and target speed
			// expectedTime = totalCycles / targetHz (in seconds)
			expectedTime := time.Duration(float64(totalCycles) / d.targetSpeedHz * float64(time.Second))

			if elapsed < expectedTime {
				// We're running ahead of schedule, sleep to catch up
				sleepDuration := expectedTime - elapsed
				time.Sleep(sleepDuration)
				result.Lagging = false
				result.LagCycles = 0
			} else {
				// We're running behind schedule
				lagTime := elapsed - expectedTime
				// Convert lag time back to cycles: lagCycles = lagTime * targetHz
				result.LagCycles = int64(lagTime.Seconds() * d.targetSpeedHz)
				result.Lagging = result.LagCycles > 0

				// Fire lagging event if we're behind by more than 10% of target speed
				// (i.e., if we've fallen behind by more than 0.1 seconds worth of cycles)
				lagThresholdCycles := int64(d.targetSpeedHz * 0.1) // 10% of one second's worth
				if result.LagCycles > lagThresholdCycles {
					d.fireEvent(EventLagging, result)
				}
			}
		}

		d.lastResult = result
	}
}

// RunUntil executes until the PC reaches the target address
func (d *Debugger) RunUntil(targetAddr uint32) (*ExecutionResult, error) {
	// Add temporary breakpoint
	bp := d.AddBreakpoint(targetAddr)
	defer d.RemoveBreakpoint(bp.ID)

	return d.Continue()
}

// StepOver executes one instruction, stepping over function calls
// For now, this is the same as Step (full implementation would track call depth)
func (d *Debugger) StepOver() (*ExecutionResult, error) {
	pc, err := cpu.ReadPC(d.runtime.CPU().Registers())
	if err != nil {
		return nil, fmt.Errorf("failed to read PC register before StepOver: %w", err)
	}

	// Check if the current instruction is a call
	if d.isCallInstruction(pc) {
		// Set temporary breakpoint at return address (PC + 4)
		returnAddr := pc + 4

		// Add temporary breakpoint
		bp, err := b.AddBreakpoint(returnAddr)
		if err != nil {
			// If we can't set breakpoint, just do a regular step
			return b.Step(1)
		}

		// Continue execution
		result := b.Continue()

		// Remove temporary breakpoint
		b.RemoveBreakpoint(bp.ID)

		// If we stopped at our temporary breakpoint, report as step
		if result.StopReason == interpreter.StopBreakpoint && b.runner.State().PC == returnAddr {
			result.StopReason = interpreter.StopStep
		}

		return result
	}

	// Not a call, just do a regular step
	return b.Step(1)
}

// isCallInstruction checks if the instruction at addr is a function call
// A branch is a call if the branch target is a function symbol
func (d *Debugger) isCallInstruction(addr uint32) (bool, error) {
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

	instr, err = d.program.InstructionAtAddress(targetAddress)
	if err != nil {
		return false, fmt.Errorf("failed to get instruction at branch target address 0x%X: %w", targetAddress, err)
	}

	// Check if the branch target is a function
	// Use the same logic as getBranchTarget - backtrack to find the MOVIMM16L/H
	// that loads the target register, and check if the symbol is a function
	return b.isBranchTargetFunction(idx, instrs)

}

// isBranchTargetFunction checks if the branch target at the given instruction index is a function
func (b *Backend) isBranchTargetFunction(instrIdx int, instrs []mc.Instruction) bool {
	instr := instrs[instrIdx]

	// Get mnemonic
	if instr.Instruction == nil || instr.Instruction.Descriptor == nil {
		return false
	}
	mnemonic := strings.ToUpper(instr.Instruction.Descriptor.OpCode.Mnemonic)

	// Determine which operand is the target register
	targetRegIdx := 0
	if mnemonic == "CJMP" && len(instr.Instruction.OperandValues) >= 2 {
		targetRegIdx = 1 // CJMP: condcode, target, link
	}

	// Get the target register
	if targetRegIdx >= len(instr.Instruction.OperandValues) {
		return false
	}
	targetOp := instr.Instruction.OperandValues[targetRegIdx]
	if targetOp.Kind() != instructions.OperandKind_Register {
		return false
	}
	targetReg := targetOp.Register()
	if targetReg == nil {
		return false
	}
	targetRegName := targetReg.Name()

	// Backtrack through previous instructions looking for MOVIMM16L/MOVIMM16H
	// that write to this register and have a function symbol
	for i := instrIdx - 1; i >= 0 && i >= instrIdx-20; i-- {
		prevInstr := instrs[i]
		if prevInstr.Instruction == nil || prevInstr.Instruction.Descriptor == nil {
			continue
		}

		prevMnemonic := strings.ToUpper(prevInstr.Instruction.Descriptor.OpCode.Mnemonic)

		// Check if this instruction writes to our target register with an immediate
		if (prevMnemonic == "MOVIMM16L" || prevMnemonic == "MOVIMM16H") &&
			len(prevInstr.Instruction.OperandValues) >= 2 {

			// MOVIMM16L/H format: imm, dest_reg
			destOp := prevInstr.Instruction.OperandValues[1]
			if destOp.Kind() == instructions.OperandKind_Register {
				destReg := destOp.Register()
				if destReg != nil && destReg.Name() == targetRegName {
					// Found an immediate load to our target register
					// Check for associated function symbol
					for _, sym := range prevInstr.Symbols {
						if sym.Function != nil {
							return true // Branch target is a function - this is a call
						}
					}
				}
			}
		}

		// If we found a different instruction that writes to our register, stop
		if prevMnemonic != "MOVIMM16L" && prevMnemonic != "MOVIMM16H" {
			if prevInstr.Instruction != nil && prevInstr.Instruction.Descriptor != nil {
				for opIdx, opDesc := range prevInstr.Instruction.Descriptor.Operands {
					if opIdx < len(prevInstr.Instruction.OperandValues) &&
						opDesc.Role == instructions.OperandRole_Destination {
						destOp := prevInstr.Instruction.OperandValues[opIdx]
						if destOp.Kind() == instructions.OperandKind_Register {
							destReg := destOp.Register()
							if destReg != nil && destReg.Name() == targetRegName {
								// Different instruction writes to target reg, stop backtracking
								return false
							}
						}
					}
				}
			}
		}
	}

	return false
}

// StepOut executes until returning from the current function
// For now, this runs until LR is reached (simplified implementation)
func (d *Debugger) StepOut() (*ExecutionResult, error) {
	lr, err := cpu.ReadLR(d.runtime.CPU().Registers())
	if err != nil {
		return nil, fmt.Errorf("failed to read LR register for StepOut: %w", err)
	}

	return d.RunUntil(lr)
}

// --- Helper functions ---

func (d *Debugger) fireEvent(event DebugEvent, result *ExecutionResult) bool {
	if d.eventCallback != nil {
		return d.eventCallback(event, result)
	}
	return true // Continue by default
}

func (d *Debugger) checkWatchpoints() (*Watchpoint, error) {
	for _, wp := range d.watchpoints {
		if !wp.Enabled {
			continue
		}

		if wp.Memory.Size > 4 {
			return nil, fmt.Errorf("watchpoint size too large: %d bytes (max 4)", wp.Memory.Size)
		}

		var currentValue uint32

		currentValue, err := memory.NewSlice(d.runtime.Memory(), wp.Memory).ReadAsUint32()
		if err != nil {
			return nil, fmt.Errorf("failed to read watchpoint memory: %w", err)
		}

		if currentValue != wp.LastValue && (wp.Type&WatchRead) != 0 {
			wp.HitCount++
			wp.LastValue = currentValue
			return wp, nil
		}
	}

	return nil, nil
}
