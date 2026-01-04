// Package interpreter provides an automatic interpreter for Cucaracha machine code
// based on the instruction descriptors.
package interpreter

import (
	"fmt"
	"time"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/registers"
)

// CPUState represents the complete state of the CPU
type CPUState struct {
	// General purpose registers (r0-r255)
	Registers [256]uint32
	// Program counter (byte address)
	PC uint32
	// Stack pointer (alias for r7)
	SP *uint32
	// Link register (alias for r5)
	LR *uint32
	// Memory (simplified as a byte slice)
	Memory []byte
	// Halted flag
	Halted bool
}

// NewCPUState creates a new CPU state with the given memory size
func NewCPUState(memorySize uint32) *CPUState {
	state := &CPUState{
		Memory: make([]byte, memorySize),
	}
	// Set up register aliases using encoded register indices
	spIdx := registers.Register("sp").Encode()
	lrIdx := registers.Register("lr").Encode()
	state.SP = &state.Registers[spIdx]
	state.LR = &state.Registers[lrIdx]
	// Initialize stack pointer to last valid word-aligned address
	// Stack grows downward, so SP points to the next address to write
	*state.SP = memorySize - 4
	return state
}

// ReadMemory32 reads a 32-bit word from memory (little-endian)
func (s *CPUState) ReadMemory32(addr uint32) (uint32, error) {
	if addr+4 > uint32(len(s.Memory)) {
		return 0, fmt.Errorf("memory access out of bounds: 0x%08X", addr)
	}
	return uint32(s.Memory[addr]) |
		uint32(s.Memory[addr+1])<<8 |
		uint32(s.Memory[addr+2])<<16 |
		uint32(s.Memory[addr+3])<<24, nil
}

// WriteMemory32 writes a 32-bit word to memory (little-endian)
func (s *CPUState) WriteMemory32(addr uint32, value uint32) error {
	if addr+4 > uint32(len(s.Memory)) {
		return fmt.Errorf("memory access out of bounds: 0x%08X", addr)
	}
	s.Memory[addr] = byte(value)
	s.Memory[addr+1] = byte(value >> 8)
	s.Memory[addr+2] = byte(value >> 16)
	s.Memory[addr+3] = byte(value >> 24)
	return nil
}

// GetRegister returns the value of a register by index (implements ExecuteContext)
func (s *CPUState) GetRegister(idx uint32) uint32 {
	return s.Registers[idx]
}

// SetRegister sets the value of a register by index (implements ExecuteContext)
func (s *CPUState) SetRegister(idx uint32, value uint32) {
	s.Registers[idx] = value
}

// GetPC returns the current program counter (implements ExecuteContext)
func (s *CPUState) GetPC() uint32 {
	return s.PC
}

// SetPC sets the program counter (implements ExecuteContext)
func (s *CPUState) SetPC(pc uint32) {
	s.PC = pc
}

// Interpreter executes Cucaracha machine code using the instruction descriptors
type Interpreter struct {
	state *CPUState
	// Target execution speed in Hz (cycles per second)
	// 0 means unlimited (full speed, no timing simulation)
	targetSpeedHz float64
	// Track timing for speed control
	cycleAccumulator int64     // Accumulated cycles since last timing reset
	timingStartTime  time.Time // When timing measurement started
}

// NewInterpreter creates a new interpreter with the given memory size
func NewInterpreter(memorySize uint32) *Interpreter {
	return &Interpreter{
		state: NewCPUState(memorySize),
	}
}

// SetTargetSpeed sets the target execution speed in Hz (cycles per second).
// Use 0 for unlimited speed (no timing simulation).
// For example, 1000 Hz means 1000 cycles per second.
func (i *Interpreter) SetTargetSpeed(hz float64) {
	if hz < 0 {
		hz = 0
	}
	i.targetSpeedHz = hz
	i.ResetTiming()
}

// GetTargetSpeed returns the current target execution speed in Hz.
func (i *Interpreter) GetTargetSpeed() float64 {
	return i.targetSpeedHz
}

// ResetTiming resets the timing accumulator for speed control.
// Should be called when starting a new execution run.
func (i *Interpreter) ResetTiming() {
	i.cycleAccumulator = 0
	i.timingStartTime = time.Now()
}

// SetExecutionDelay is deprecated. Use SetTargetSpeed instead.
// This method converts the delay to an approximate Hz value for backward compatibility.
// Use 0 for full speed.
func (i *Interpreter) SetExecutionDelay(delayMs int) {
	if delayMs <= 0 {
		i.SetTargetSpeed(0) // Unlimited speed
	} else {
		// Convert ms delay to Hz: if delay is N ms per instruction,
		// that's 1000/N instructions per second
		hz := 1000.0 / float64(delayMs)
		i.SetTargetSpeed(hz)
	}
}

// GetExecutionDelay is deprecated. Use GetTargetSpeed instead.
// Returns an approximate delay in ms for backward compatibility.
func (i *Interpreter) GetExecutionDelay() int {
	if i.targetSpeedHz <= 0 {
		return 0 // Unlimited speed
	}
	// Convert Hz to ms: if target is N Hz, delay is 1000/N ms
	return int(1000.0 / i.targetSpeedHz)
}

// State returns the current CPU state
func (i *Interpreter) State() *CPUState {
	return i.state
}

// LoadBinary loads binary machine code into memory at the given address
func (i *Interpreter) LoadBinary(binary []byte, addr uint32) error {
	if addr+uint32(len(binary)) > uint32(len(i.state.Memory)) {
		return fmt.Errorf("program too large for memory")
	}
	copy(i.state.Memory[addr:], binary)
	i.state.PC = addr
	i.state.Halted = false
	return nil
}

// LoadProgram encodes a program to binary and loads it into memory at the given address
func (i *Interpreter) LoadProgram(program *mc.Program, addr uint32) error {
	binary, err := program.Encode()
	if err != nil {
		return fmt.Errorf("failed to encode program: %w", err)
	}
	return i.LoadBinary(binary, addr)
}

// DecodeInstruction decodes an instruction from memory at the current PC
func (i *Interpreter) DecodeInstruction() (*instructions.InstructionDescriptor, []uint32, error) {
	// Read the 32-bit instruction word
	word, err := i.state.ReadMemory32(i.state.PC)
	if err != nil {
		return nil, nil, err
	}

	// Extract opcode (bits 0-4)
	opcode := instructions.OpCode(word & 0x1F)

	// Find the instruction descriptor
	desc, err := instructions.Instructions.Instruction(opcode)
	if err != nil {
		return nil, nil, fmt.Errorf("unknown opcode: %d - %w", opcode, err)
	}

	// Decode operands based on descriptor
	operands := make([]uint32, len(desc.Operands))
	for idx, op := range desc.Operands {
		if op.EncodingBits == 0 {
			// Tied operand - skip encoding, will be same as tied register
			continue
		}
		// Extract bits from instruction word
		mask := uint32((1 << op.EncodingBits) - 1)
		operands[idx] = (word >> op.EncodingPosition) & mask
	}

	return desc, operands, nil
}

// StepResult contains the result of executing a single instruction
type StepResult struct {
	// Cycles is the number of CPU cycles consumed by the instruction
	Cycles int
	// Instruction is the executed instruction descriptor
	Instruction *instructions.InstructionDescriptor
	// Operands are the decoded operand values
	Operands []uint32
}

// Step executes a single instruction and returns the cycle count
func (i *Interpreter) Step() (*StepResult, error) {
	if i.state.Halted {
		return nil, fmt.Errorf("CPU is halted")
	}

	desc, operands, err := i.DecodeInstruction()
	if err != nil {
		return nil, err
	}

	// Save current PC for detecting branches
	oldPC := i.state.PC

	// Execute the instruction
	if err := i.executeInstruction(desc, operands); err != nil {
		return nil, fmt.Errorf("error executing %s at 0x%08X: %w", desc.OpCode.Mnemonic, oldPC, err)
	}

	// Advance PC if not modified by a branch
	if i.state.PC == oldPC {
		i.state.PC += 4 // Instructions are 32 bits
	}

	return &StepResult{
		Cycles:      desc.GetCycles(),
		Instruction: desc,
		Operands:    operands,
	}, nil
}

// executeInstruction executes an instruction using its descriptor's Execute function
func (i *Interpreter) executeInstruction(desc *instructions.InstructionDescriptor, operands []uint32) error {
	if desc.Execute == nil {
		return fmt.Errorf("unimplemented opcode: %s (no Execute function)", desc.OpCode.Mnemonic)
	}
	return desc.Execute(i.state, operands)
}

// Run executes instructions until halted or an error occurs
func (i *Interpreter) Run() error {
	for !i.state.Halted {
		if _, err := i.Step(); err != nil {
			return err
		}
	}
	return nil
}

// RunN executes at most n instructions
func (i *Interpreter) RunN(n int) error {
	for count := 0; count < n && !i.state.Halted; count++ {
		if _, err := i.Step(); err != nil {
			return err
		}
	}
	return nil
}

// Reset resets the CPU state
func (i *Interpreter) Reset() {
	memSize := uint32(len(i.state.Memory))
	i.state = NewCPUState(memSize)
}

// Utility function to get register name
func RegisterName(idx uint32) string {
	for _, reg := range registers.IntegerRegisters.AllRegisters() {
		if uint32(reg.Encode()) == idx {
			return reg.Name()
		}
	}
	return fmt.Sprintf("r%d", idx)
}
