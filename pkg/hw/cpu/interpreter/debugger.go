// Package interpreter provides debugging and step-by-step execution support
// for the Cucaracha CPU interpreter.
package interpreter

import (
	"fmt"
	"sort"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
)

// ExecutionEvent represents events that can occur during execution
type ExecutionEvent int

const (
	// EventStep is fired after each instruction execution
	EventStep ExecutionEvent = iota
	// EventBreakpoint is fired when a breakpoint is hit
	EventBreakpoint
	// EventWatchpoint is fired when a watched memory location is modified
	EventWatchpoint
	// EventHalt is fired when the CPU halts
	EventHalt
	// EventError is fired when an execution error occurs
	EventError
	// EventTermination is fired when execution reaches a termination address
	EventTermination
)

// String returns the string representation of an ExecutionEvent
func (e ExecutionEvent) String() string {
	switch e {
	case EventStep:
		return "step"
	case EventBreakpoint:
		return "breakpoint"
	case EventWatchpoint:
		return "watchpoint"
	case EventHalt:
		return "halt"
	case EventError:
		return "error"
	case EventTermination:
		return "termination"
	default:
		return fmt.Sprintf("unknown(%d)", e)
	}
}

// StopReason indicates why execution stopped
type StopReason int

const (
	// StopNone indicates execution has not stopped
	StopNone StopReason = iota
	// StopStep indicates execution stopped after a single step
	StopStep
	// StopBreakpoint indicates execution stopped at a breakpoint
	StopBreakpoint
	// StopWatchpoint indicates execution stopped due to a watchpoint
	StopWatchpoint
	// StopHalt indicates the CPU halted
	StopHalt
	// StopError indicates an execution error occurred
	StopError
	// StopTermination indicates normal program termination
	StopTermination
	// StopMaxSteps indicates max steps limit was reached
	StopMaxSteps
)

// String returns the string representation of a StopReason
func (r StopReason) String() string {
	switch r {
	case StopNone:
		return "none"
	case StopStep:
		return "step"
	case StopBreakpoint:
		return "breakpoint"
	case StopWatchpoint:
		return "watchpoint"
	case StopHalt:
		return "halt"
	case StopError:
		return "error"
	case StopTermination:
		return "termination"
	case StopMaxSteps:
		return "max_steps"
	default:
		return fmt.Sprintf("unknown(%d)", r)
	}
}

// Breakpoint represents a code breakpoint
type Breakpoint struct {
	// ID is the unique breakpoint identifier
	ID int
	// Address is the memory address of the breakpoint
	Address uint32
	// Enabled indicates if the breakpoint is active
	Enabled bool
	// HitCount tracks how many times this breakpoint has been hit
	HitCount int
	// Condition is an optional condition expression (for future use)
	Condition string
}

// Watchpoint represents a memory watchpoint
type Watchpoint struct {
	// ID is the unique watchpoint identifier
	ID int
	// Address is the memory address to watch
	Address uint32
	// Size is the number of bytes to watch (1, 2, or 4)
	Size int
	// Type indicates read, write, or read/write
	Type WatchpointType
	// Enabled indicates if the watchpoint is active
	Enabled bool
	// HitCount tracks how many times this watchpoint has been triggered
	HitCount int
	// LastValue stores the last known value at this address
	LastValue uint32
}

// WatchpointType indicates what access triggers the watchpoint
type WatchpointType int

const (
	// WatchWrite triggers on write access
	WatchWrite WatchpointType = 1 << iota
	// WatchRead triggers on read access
	WatchRead
	// WatchReadWrite triggers on any access
	WatchReadWrite = WatchWrite | WatchRead
)

// ExecutionResult contains the result of an execution operation
type ExecutionResult struct {
	// StopReason indicates why execution stopped
	StopReason StopReason
	// StepsExecuted is the number of instructions executed
	StepsExecuted int
	// Error contains any error that occurred (nil if none)
	Error error
	// BreakpointID is set if stopped at a breakpoint
	BreakpointID int
	// WatchpointID is set if stopped at a watchpoint
	WatchpointID int
	// LastPC is the PC value when execution stopped
	LastPC uint32
	// LastInstruction is the last decoded instruction (if available)
	LastInstruction *instructions.InstructionDescriptor
	// LastOperands are the operands of the last instruction
	LastOperands []uint32
}

// EventCallback is called when an execution event occurs
// Return true to continue execution, false to stop
type EventCallback func(event ExecutionEvent, result *ExecutionResult) bool

// Debugger provides debugging capabilities for the interpreter
type Debugger struct {
	interp *Interpreter

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
}

// NewDebugger creates a new debugger for the given interpreter
func NewDebugger(interp *Interpreter) *Debugger {
	return &Debugger{
		interp:           interp,
		breakpoints:      make(map[int]*Breakpoint),
		breakpointAddrs:  make(map[uint32]*Breakpoint),
		watchpoints:      make(map[int]*Watchpoint),
		terminationAddrs: make(map[uint32]bool),
	}
}

// Interpreter returns the underlying interpreter
func (d *Debugger) Interpreter() *Interpreter {
	return d.interp
}

// State returns the current CPU state
func (d *Debugger) State() *CPUState {
	return d.interp.State()
}

// SetEventCallback sets the callback for execution events
func (d *Debugger) SetEventCallback(callback EventCallback) {
	d.eventCallback = callback
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
func (d *Debugger) AddWatchpoint(addr uint32, size int, wpType WatchpointType) *Watchpoint {
	// Read initial value
	var lastValue uint32
	switch size {
	case 1:
		if int(addr) < len(d.interp.state.Memory) {
			lastValue = uint32(d.interp.state.Memory[addr])
		}
	case 2:
		if int(addr)+1 < len(d.interp.state.Memory) {
			lastValue = uint32(d.interp.state.Memory[addr]) | uint32(d.interp.state.Memory[addr+1])<<8
		}
	case 4:
		lastValue, _ = d.interp.state.ReadMemory32(addr)
	}

	wp := &Watchpoint{
		ID:        d.nextWatchpointID,
		Address:   addr,
		Size:      size,
		Type:      wpType,
		Enabled:   true,
		LastValue: lastValue,
	}
	d.nextWatchpointID++
	d.watchpoints[wp.ID] = wp
	return wp
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
		return wps[i].Address < wps[j].Address
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
func (d *Debugger) Step() *ExecutionResult {
	result := &ExecutionResult{
		LastPC: d.interp.state.PC,
	}

	// Check for termination address before executing
	if d.terminationAddrs[d.interp.state.PC] {
		result.StopReason = StopTermination
		d.lastResult = result
		d.fireEvent(EventTermination, result)
		return result
	}

	// Check for breakpoint (after first step, stepping over breakpoint)
	if bp := d.breakpointAddrs[d.interp.state.PC]; bp != nil && bp.Enabled {
		bp.HitCount++
		result.StopReason = StopBreakpoint
		result.BreakpointID = bp.ID
		d.lastResult = result
		d.fireEvent(EventBreakpoint, result)
		return result
	}

	// Check if halted
	if d.interp.state.Halted {
		result.StopReason = StopHalt
		d.lastResult = result
		d.fireEvent(EventHalt, result)
		return result
	}

	// Decode instruction for introspection
	desc, operands, decodeErr := d.interp.DecodeInstruction()
	if decodeErr != nil {
		result.StopReason = StopError
		result.Error = decodeErr
		d.lastResult = result
		d.fireEvent(EventError, result)
		return result
	}
	result.LastInstruction = desc
	result.LastOperands = operands

	// Execute the instruction
	err := d.interp.Step()
	result.StepsExecuted = 1

	if err != nil {
		result.StopReason = StopError
		result.Error = err
		d.lastResult = result
		d.fireEvent(EventError, result)
		return result
	}

	// Check watchpoints after execution
	if wp := d.checkWatchpoints(); wp != nil {
		result.StopReason = StopWatchpoint
		result.WatchpointID = wp.ID
		d.lastResult = result
		d.fireEvent(EventWatchpoint, result)
		return result
	}

	result.StopReason = StopStep
	d.lastResult = result
	d.fireEvent(EventStep, result)
	return result
}

// Continue executes until a stop condition is met
func (d *Debugger) Continue() *ExecutionResult {
	return d.Run(0)
}

// Run executes up to maxSteps instructions (0 = unlimited)
func (d *Debugger) Run(maxSteps int) *ExecutionResult {
	result := &ExecutionResult{
		LastPC: d.interp.state.PC,
	}

	for {
		// Check step limit
		if maxSteps > 0 && result.StepsExecuted >= maxSteps {
			result.StopReason = StopMaxSteps
			break
		}

		// Check for termination address
		if d.terminationAddrs[d.interp.state.PC] {
			result.StopReason = StopTermination
			d.fireEvent(EventTermination, result)
			break
		}

		// Check for breakpoint (not on first step if we're already there)
		if result.StepsExecuted > 0 {
			if bp := d.breakpointAddrs[d.interp.state.PC]; bp != nil && bp.Enabled {
				bp.HitCount++
				result.StopReason = StopBreakpoint
				result.BreakpointID = bp.ID
				d.fireEvent(EventBreakpoint, result)
				break
			}
		}

		// Check if halted
		if d.interp.state.Halted {
			result.StopReason = StopHalt
			d.fireEvent(EventHalt, result)
			break
		}

		// Decode for introspection
		desc, operands, decodeErr := d.interp.DecodeInstruction()
		if decodeErr != nil {
			result.StopReason = StopError
			result.Error = decodeErr
			d.fireEvent(EventError, result)
			break
		}
		result.LastInstruction = desc
		result.LastOperands = operands
		result.LastPC = d.interp.state.PC

		// Execute
		if err := d.interp.Step(); err != nil {
			result.StopReason = StopError
			result.Error = err
			d.fireEvent(EventError, result)
			break
		}
		result.StepsExecuted++

		// Fire step event and check if we should continue
		if !d.fireEvent(EventStep, result) {
			result.StopReason = StopStep
			break
		}

		// Check watchpoints
		if wp := d.checkWatchpoints(); wp != nil {
			result.StopReason = StopWatchpoint
			result.WatchpointID = wp.ID
			d.fireEvent(EventWatchpoint, result)
			break
		}
	}

	d.lastResult = result
	return result
}

// RunUntil executes until the PC reaches the target address
func (d *Debugger) RunUntil(targetAddr uint32) *ExecutionResult {
	// Add temporary breakpoint
	bp := d.AddBreakpoint(targetAddr)
	defer d.RemoveBreakpoint(bp.ID)

	return d.Continue()
}

// StepOver executes one instruction, stepping over function calls
// For now, this is the same as Step (full implementation would track call depth)
func (d *Debugger) StepOver() *ExecutionResult {
	// TODO: Implement proper step-over by detecting call instructions
	// and setting a temporary breakpoint at the return address
	return d.Step()
}

// StepOut executes until returning from the current function
// For now, this runs until LR is reached (simplified implementation)
func (d *Debugger) StepOut() *ExecutionResult {
	lr := *d.interp.state.LR
	return d.RunUntil(lr)
}

// --- Introspection ---

// CurrentInstruction decodes and returns the current instruction
func (d *Debugger) CurrentInstruction() (*instructions.InstructionDescriptor, []uint32, error) {
	return d.interp.DecodeInstruction()
}

// DisassembleAt disassembles the instruction at the given address
func (d *Debugger) DisassembleAt(addr uint32) (string, error) {
	// Save and restore PC
	savedPC := d.interp.state.PC
	d.interp.state.PC = addr
	defer func() { d.interp.state.PC = savedPC }()

	desc, operands, err := d.interp.DecodeInstruction()
	if err != nil {
		return "", err
	}

	return formatInstruction(desc, operands), nil
}

// DisassembleRange disassembles a range of addresses
func (d *Debugger) DisassembleRange(startAddr, endAddr uint32) ([]string, error) {
	var result []string
	for addr := startAddr; addr < endAddr; addr += 4 {
		line, err := d.DisassembleAt(addr)
		if err != nil {
			result = append(result, fmt.Sprintf("0x%08X: <error: %v>", addr, err))
		} else {
			result = append(result, fmt.Sprintf("0x%08X: %s", addr, line))
		}
	}
	return result, nil
}

// ReadMemory reads memory at the given address
func (d *Debugger) ReadMemory(addr uint32, size int) ([]byte, error) {
	if int(addr)+size > len(d.interp.state.Memory) {
		return nil, fmt.Errorf("memory access out of bounds: 0x%08X + %d", addr, size)
	}
	result := make([]byte, size)
	copy(result, d.interp.state.Memory[addr:addr+uint32(size)])
	return result, nil
}

// ReadMemory32 reads a 32-bit word from memory
func (d *Debugger) ReadMemory32(addr uint32) (uint32, error) {
	return d.interp.state.ReadMemory32(addr)
}

// WriteMemory writes data to memory
func (d *Debugger) WriteMemory(addr uint32, data []byte) error {
	if int(addr)+len(data) > len(d.interp.state.Memory) {
		return fmt.Errorf("memory access out of bounds: 0x%08X + %d", addr, len(data))
	}
	copy(d.interp.state.Memory[addr:], data)
	return nil
}

// WriteMemory32 writes a 32-bit word to memory
func (d *Debugger) WriteMemory32(addr uint32, value uint32) error {
	return d.interp.state.WriteMemory32(addr, value)
}

// GetRegister returns the value of a register by index
func (d *Debugger) GetRegister(idx uint32) uint32 {
	return d.interp.state.GetRegister(idx)
}

// SetRegister sets the value of a register by index
func (d *Debugger) SetRegister(idx uint32, value uint32) {
	d.interp.state.SetRegister(idx, value)
}

// GetPC returns the current program counter
func (d *Debugger) GetPC() uint32 {
	return d.interp.state.PC
}

// SetPC sets the program counter
func (d *Debugger) SetPC(pc uint32) {
	d.interp.state.PC = pc
}

// GetSP returns the stack pointer
func (d *Debugger) GetSP() uint32 {
	return *d.interp.state.SP
}

// SetSP sets the stack pointer
func (d *Debugger) SetSP(sp uint32) {
	*d.interp.state.SP = sp
}

// GetLR returns the link register
func (d *Debugger) GetLR() uint32 {
	return *d.interp.state.LR
}

// SetLR sets the link register
func (d *Debugger) SetLR(lr uint32) {
	*d.interp.state.LR = lr
}

// IsHalted returns whether the CPU is halted
func (d *Debugger) IsHalted() bool {
	return d.interp.state.Halted
}

// --- Helper functions ---

func (d *Debugger) fireEvent(event ExecutionEvent, result *ExecutionResult) bool {
	if d.eventCallback != nil {
		return d.eventCallback(event, result)
	}
	return true // Continue by default
}

func (d *Debugger) checkWatchpoints() *Watchpoint {
	for _, wp := range d.watchpoints {
		if !wp.Enabled {
			continue
		}

		var currentValue uint32
		switch wp.Size {
		case 1:
			if int(wp.Address) < len(d.interp.state.Memory) {
				currentValue = uint32(d.interp.state.Memory[wp.Address])
			}
		case 2:
			if int(wp.Address)+1 < len(d.interp.state.Memory) {
				currentValue = uint32(d.interp.state.Memory[wp.Address]) |
					uint32(d.interp.state.Memory[wp.Address+1])<<8
			}
		case 4:
			currentValue, _ = d.interp.state.ReadMemory32(wp.Address)
		}

		if currentValue != wp.LastValue && (wp.Type&WatchWrite) != 0 {
			wp.HitCount++
			wp.LastValue = currentValue
			return wp
		}
	}
	return nil
}

func formatInstruction(desc *instructions.InstructionDescriptor, operands []uint32) string {
	result := desc.OpCode.Mnemonic
	for i, op := range desc.Operands {
		if i < len(operands) {
			if op.Kind == instructions.OperandKind_Immediate {
				result += fmt.Sprintf(" #%d", operands[i])
			} else {
				result += fmt.Sprintf(" %s", RegisterName(operands[i]))
			}
		}
	}
	return result
}
