// Package cpu provides CPU execution runners that work with any cpu.CPU implementation.
package cpu

import (
	"fmt"
	"sync/atomic"
)

// ExecutionEvent represents events that can occur during execution
type ExecutionEvent int

const (
	// EventStep is fired after each instruction execution
	EventStep ExecutionEvent = iota
	// EventBreakpoint is fired when a breakpoint is hit
	EventBreakpoint
	// EventHalt is fired when the CPU halts
	EventHalt
	// EventError is fired when an execution error occurs
	EventError
	// EventTermination is fired when execution reaches a termination address
	EventTermination
)

// StopReason indicates why execution stopped
type StopReason int

const (
	// StopNone indicates execution has not stopped
	StopNone StopReason = iota
	// StopStep indicates execution stopped after a single step
	StopStep
	// StopBreakpoint indicates execution stopped at a breakpoint
	StopBreakpoint
	// StopHalt indicates the CPU halted
	StopHalt
	// StopError indicates an execution error occurred
	StopError
	// StopTermination indicates normal program termination
	StopTermination
	// StopMaxSteps indicates max steps limit was reached
	StopMaxSteps
	// StopInterrupt indicates execution was interrupted by user
	StopInterrupt
)

// Breakpoint represents a code breakpoint
type Breakpoint struct {
	ID      int
	Address uint32
	Enabled bool
}

// ExecutionResult contains the result of an execution operation
type ExecutionResult struct {
	StopReason    StopReason
	StepsExecuted int
	Error         error
	BreakpointID  int
	LastPC        uint32
}

// EventCallback is called when an execution event occurs
// Return true to continue execution, false to stop
type EventCallback func(event ExecutionEvent, result *ExecutionResult) bool

// CPURunner provides debugging and execution control for any cpu.CPU implementation.
// It allows breakpoints, termination addresses, and step-by-step execution.
type CPURunner struct {
	cpu CPU

	// Breakpoints indexed by ID
	breakpoints map[int]*Breakpoint
	// Breakpoint addresses for fast lookup
	breakpointAddrs map[uint32]*Breakpoint
	// Next breakpoint ID
	nextBreakpointID int

	// Termination addresses
	terminationAddrs map[uint32]bool

	// Event callback
	eventCallback EventCallback

	// Interrupt flag (atomic for thread-safety)
	interrupted int32

	// Execution state
	lastResult *ExecutionResult
}

// NewCPURunner creates a new runner for the given CPU
func NewCPURunner(cpu CPU) *CPURunner {
	return &CPURunner{
		cpu:              cpu,
		breakpoints:      make(map[int]*Breakpoint),
		breakpointAddrs:  make(map[uint32]*Breakpoint),
		terminationAddrs: make(map[uint32]bool),
	}
}

// CPU returns the underlying CPU
func (r *CPURunner) CPU() CPU {
	return r.cpu
}

// AddBreakpoint adds a breakpoint at the specified address
func (r *CPURunner) AddBreakpoint(addr uint32) int {
	id := r.nextBreakpointID
	r.nextBreakpointID++

	bp := &Breakpoint{
		ID:      id,
		Address: addr,
		Enabled: true,
	}

	r.breakpoints[id] = bp
	r.breakpointAddrs[addr] = bp

	return id
}

// RemoveBreakpoint removes a breakpoint by ID
func (r *CPURunner) RemoveBreakpoint(id int) error {
	bp, ok := r.breakpoints[id]
	if !ok {
		return fmt.Errorf("breakpoint %d not found", id)
	}

	delete(r.breakpointAddrs, bp.Address)
	delete(r.breakpoints, id)

	return nil
}

// GetBreakpointAt returns the breakpoint at the given address, or nil
func (r *CPURunner) GetBreakpointAt(addr uint32) *Breakpoint {
	return r.breakpointAddrs[addr]
}

// AddTerminationAddress adds an address that signals program termination
func (r *CPURunner) AddTerminationAddress(addr uint32) {
	r.terminationAddrs[addr] = true
}

// RemoveTerminationAddress removes a termination address
func (r *CPURunner) RemoveTerminationAddress(addr uint32) {
	delete(r.terminationAddrs, addr)
}

// SetEventCallback sets the callback for execution events
func (r *CPURunner) SetEventCallback(callback EventCallback) {
	r.eventCallback = callback
}

// Interrupt signals the runner to stop execution
func (r *CPURunner) Interrupt() {
	atomic.StoreInt32(&r.interrupted, 1)
}

// ClearInterrupt clears the interrupt flag
func (r *CPURunner) ClearInterrupt() {
	atomic.StoreInt32(&r.interrupted, 0)
}

// IsInterrupted returns true if the interrupt flag is set
func (r *CPURunner) IsInterrupted() bool {
	return atomic.LoadInt32(&r.interrupted) != 0
}

// Step executes a single instruction
func (r *CPURunner) Step() *ExecutionResult {
	r.ClearInterrupt()

	result := &ExecutionResult{
		StopReason: StopStep,
		LastPC:     r.cpu.GetPC(),
	}

	// Check for termination address before executing
	if r.terminationAddrs[r.cpu.GetPC()] {
		result.StopReason = StopTermination
		r.lastResult = result
		r.fireEvent(EventTermination, result)
		return result
	}

	// Check for breakpoint (skip if we just set it)
	if bp := r.GetBreakpointAt(r.cpu.GetPC()); bp != nil && bp.Enabled {
		result.StopReason = StopBreakpoint
		result.BreakpointID = bp.ID
		r.lastResult = result
		r.fireEvent(EventBreakpoint, result)
		return result
	}

	// Execute one instruction
	if err := r.cpu.Step(); err != nil {
		result.StopReason = StopError
		result.Error = err
		r.lastResult = result
		r.fireEvent(EventError, result)
		return result
	}

	result.StepsExecuted = 1
	result.LastPC = r.cpu.GetPC()

	// Check for halt
	if r.cpu.IsHalted() {
		result.StopReason = StopHalt
		r.lastResult = result
		r.fireEvent(EventHalt, result)
		return result
	}

	r.lastResult = result
	r.fireEvent(EventStep, result)
	return result
}

// Run executes until a stop condition is reached
func (r *CPURunner) Run() *ExecutionResult {
	return r.RunN(0)
}

// RunN executes at most n instructions (0 = unlimited)
func (r *CPURunner) RunN(maxSteps int) *ExecutionResult {
	r.ClearInterrupt()

	result := &ExecutionResult{
		StopReason: StopNone,
		LastPC:     r.cpu.GetPC(),
	}

	unlimited := maxSteps == 0

	for unlimited || result.StepsExecuted < maxSteps {
		// Check interrupt
		if r.IsInterrupted() {
			result.StopReason = StopInterrupt
			break
		}

		// Check for termination address
		if r.terminationAddrs[r.cpu.GetPC()] {
			result.StopReason = StopTermination
			r.fireEvent(EventTermination, result)
			break
		}

		// Check for breakpoint
		if bp := r.GetBreakpointAt(r.cpu.GetPC()); bp != nil && bp.Enabled {
			result.StopReason = StopBreakpoint
			result.BreakpointID = bp.ID
			r.fireEvent(EventBreakpoint, result)
			break
		}

		// Execute one instruction
		if err := r.cpu.Step(); err != nil {
			result.StopReason = StopError
			result.Error = err
			r.fireEvent(EventError, result)
			break
		}

		result.StepsExecuted++
		result.LastPC = r.cpu.GetPC()

		// Fire step event and check if we should continue
		if !r.fireEvent(EventStep, result) {
			result.StopReason = StopInterrupt
			break
		}

		// Check for halt
		if r.cpu.IsHalted() {
			result.StopReason = StopHalt
			r.fireEvent(EventHalt, result)
			break
		}
	}

	if result.StopReason == StopNone && !unlimited && result.StepsExecuted >= maxSteps {
		result.StopReason = StopMaxSteps
	}

	r.lastResult = result
	return result
}

// Continue continues execution after a breakpoint
func (r *CPURunner) Continue() *ExecutionResult {
	// Skip past current breakpoint by executing one instruction
	if bp := r.GetBreakpointAt(r.cpu.GetPC()); bp != nil && bp.Enabled {
		if err := r.cpu.Step(); err != nil {
			return &ExecutionResult{
				StopReason: StopError,
				Error:      err,
				LastPC:     r.cpu.GetPC(),
			}
		}
	}

	return r.Run()
}

// Reset resets the CPU
func (r *CPURunner) Reset() {
	r.cpu.Reset()
	r.lastResult = nil
}

// LastResult returns the result of the last execution
func (r *CPURunner) LastResult() *ExecutionResult {
	return r.lastResult
}

// fireEvent fires an event to the callback
// Returns true if execution should continue, false to stop
func (r *CPURunner) fireEvent(event ExecutionEvent, result *ExecutionResult) bool {
	if r.eventCallback != nil {
		return r.eventCallback(event, result)
	}
	return true
}

// GetPC returns the current program counter
func (r *CPURunner) GetPC() uint32 {
	return r.cpu.GetPC()
}

// GetRegister returns a register value
func (r *CPURunner) GetRegister(idx int) uint32 {
	return r.cpu.GetRegister(idx)
}

// SetRegister sets a register value
func (r *CPURunner) SetRegister(idx int, value uint32) {
	r.cpu.SetRegister(idx, value)
}

// ReadMemory reads a 32-bit word from memory
func (r *CPURunner) ReadMemory(addr uint32) uint32 {
	return r.cpu.ReadMemory(addr)
}

// WriteMemory writes a 32-bit word to memory
func (r *CPURunner) WriteMemory(addr uint32, value uint32) {
	r.cpu.WriteMemory(addr, value)
}

// LoadBinary loads binary data into memory
func (r *CPURunner) LoadBinary(data []byte, addr uint32) error {
	return r.cpu.LoadBinary(data, addr)
}

// LoadProgram loads a program (slice of instructions) into memory
func (r *CPURunner) LoadProgram(program []uint32, addr uint32) error {
	return r.cpu.LoadProgram(program, addr)
}

// IsHalted returns true if the CPU is halted
func (r *CPURunner) IsHalted() bool {
	return r.cpu.IsHalted()
}
