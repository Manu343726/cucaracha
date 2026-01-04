package interpreter

import (
	"fmt"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/registers"
)

// Emulator wraps an Interpreter to implement the cpu.CPU interface.
// This allows the interpreter-based CPU to be used interchangeably with
// hardware-based CPU implementations.
type Emulator struct {
	interpreter *Interpreter
	cycles      uint64
}

// NewEmulator creates a new emulator with the given memory size
func NewEmulator(memorySize int) *Emulator {
	return &Emulator{
		interpreter: NewInterpreter(uint32(memorySize)),
	}
}

// NewEmulatorWithInterpreter wraps an existing interpreter
func NewEmulatorWithInterpreter(interp *Interpreter) *Emulator {
	return &Emulator{
		interpreter: interp,
	}
}

// Interpreter returns the underlying interpreter for advanced use cases
func (e *Emulator) Interpreter() *Interpreter {
	return e.interpreter
}

// --- cpu.CPU interface implementation ---

// GetRegister returns the value of a register by index
func (e *Emulator) GetRegister(idx int) uint32 {
	return e.interpreter.state.GetRegister(uint32(idx))
}

// SetRegister sets the value of a register by index
func (e *Emulator) SetRegister(idx int, value uint32) {
	e.interpreter.state.SetRegister(uint32(idx), value)
}

// GetPC returns the current program counter
func (e *Emulator) GetPC() uint32 {
	return e.interpreter.state.GetPC()
}

// SetPC sets the program counter
func (e *Emulator) SetPC(value uint32) {
	e.interpreter.state.SetPC(value)
}

// GetSP returns the stack pointer
func (e *Emulator) GetSP() uint32 {
	return *e.interpreter.state.SP
}

// SetSP sets the stack pointer
func (e *Emulator) SetSP(value uint32) {
	*e.interpreter.state.SP = value
}

// GetLR returns the link register
func (e *Emulator) GetLR() uint32 {
	return *e.interpreter.state.LR
}

// SetLR sets the link register
func (e *Emulator) SetLR(value uint32) {
	*e.interpreter.state.LR = value
}

// ReadMemory reads a 32-bit word from memory
func (e *Emulator) ReadMemory(addr uint32) uint32 {
	val, _ := e.interpreter.state.ReadMemory32(addr)
	return val
}

// WriteMemory writes a 32-bit word to memory
func (e *Emulator) WriteMemory(addr uint32, value uint32) {
	_ = e.interpreter.state.WriteMemory32(addr, value)
}

// ReadByte reads a single byte from memory
func (e *Emulator) ReadByte(addr uint32) byte {
	if int(addr) >= len(e.interpreter.state.Memory) {
		return 0
	}
	return e.interpreter.state.Memory[addr]
}

// WriteByte writes a single byte to memory
func (e *Emulator) WriteByte(addr uint32, value byte) {
	if int(addr) >= len(e.interpreter.state.Memory) {
		return
	}
	e.interpreter.state.Memory[addr] = value
}

// MemorySize returns the size of memory in bytes
func (e *Emulator) MemorySize() int {
	return len(e.interpreter.state.Memory)
}

// LoadBinary loads binary machine code into memory at the given address
func (e *Emulator) LoadBinary(data []byte, startAddr uint32) error {
	return e.interpreter.LoadBinary(data, startAddr)
}

// LoadProgram loads a program (slice of 32-bit instructions) into memory
func (e *Emulator) LoadProgram(program []uint32, startAddr uint32) error {
	if int(startAddr)+len(program)*4 > len(e.interpreter.state.Memory) {
		return fmt.Errorf("program too large for memory")
	}
	for i, instr := range program {
		addr := startAddr + uint32(i*4)
		if err := e.interpreter.state.WriteMemory32(addr, instr); err != nil {
			return err
		}
	}
	e.interpreter.state.PC = startAddr
	e.interpreter.state.Halted = false
	return nil
}

// Step executes one instruction
func (e *Emulator) Step() error {
	result, err := e.interpreter.Step()
	if err != nil {
		return err
	}
	e.cycles += uint64(result.Cycles)
	return nil
}

// Run executes instructions until halted
func (e *Emulator) Run() error {
	for !e.IsHalted() {
		if err := e.Step(); err != nil {
			return err
		}
	}
	return nil
}

// RunN executes at most n instructions
func (e *Emulator) RunN(n int) error {
	for i := 0; i < n && !e.IsHalted(); i++ {
		if err := e.Step(); err != nil {
			return err
		}
	}
	return nil
}

// IsHalted returns true if the CPU is halted
func (e *Emulator) IsHalted() bool {
	return e.interpreter.state.Halted
}

// Halt halts the CPU
func (e *Emulator) Halt() {
	e.interpreter.state.Halted = true
}

// Reset resets the CPU state
func (e *Emulator) Reset() {
	e.interpreter.Reset()
	e.cycles = 0
}

// Cycles returns the number of cycles executed
func (e *Emulator) Cycles() uint64 {
	return e.cycles
}

// --- cpu.DebuggableCPU interface implementation ---

// DecodeInstruction decodes the instruction at the given address
func (e *Emulator) DecodeInstruction(addr uint32) (mnemonic string, operands string, err error) {
	// Save current PC
	savedPC := e.interpreter.state.PC
	defer func() { e.interpreter.state.PC = savedPC }()

	// Temporarily set PC to the address to decode
	e.interpreter.state.PC = addr

	desc, ops, err := e.interpreter.DecodeInstruction()
	if err != nil {
		return "", "", err
	}

	// Format operands
	operandStrs := make([]string, 0, len(ops))
	for i, op := range ops {
		if i < len(desc.Operands) {
			operandStrs = append(operandStrs, formatOperand(desc.Operands[i], op))
		}
	}

	return desc.OpCode.Mnemonic, joinOperands(operandStrs), nil
}

// GetFlags returns the CPU flags register (CPSR)
func (e *Emulator) GetFlags() uint32 {
	return e.GetRegister(cpu.RegCPSR)
}

// SetFlags sets the CPU flags register (CPSR)
func (e *Emulator) SetFlags(flags uint32) {
	e.SetRegister(cpu.RegCPSR, flags)
}

// formatOperand formats a single operand for display
func formatOperand(opDesc *instructions.OperandDescriptor, value uint32) string {
	switch opDesc.Kind {
	case instructions.OperandKind_Register:
		return RegisterName(value)
	case instructions.OperandKind_Immediate:
		return fmt.Sprintf("#%d", value)
	default:
		return fmt.Sprintf("%d", value)
	}
}

// joinOperands joins operand strings with commas
func joinOperands(operands []string) string {
	result := ""
	for i, op := range operands {
		if i > 0 {
			result += ", "
		}
		result += op
	}
	return result
}

// --- Timing methods (not part of interface but useful) ---

// SetTargetSpeed sets the target execution speed in Hz
func (e *Emulator) SetTargetSpeed(hz float64) {
	e.interpreter.SetTargetSpeed(hz)
}

// GetTargetSpeed returns the current target execution speed in Hz
func (e *Emulator) GetTargetSpeed() float64 {
	return e.interpreter.GetTargetSpeed()
}

// Ensure Emulator implements the interfaces
var _ cpu.CPU = (*Emulator)(nil)
var _ cpu.DebuggableCPU = (*Emulator)(nil)

// NewCPU is a factory function that creates an Emulator implementing cpu.CPU
func NewCPU(memorySize int) cpu.CPU {
	return NewEmulator(memorySize)
}

// Helper to get register name from registers package
func getRegName(idx uint32) string {
	for _, reg := range registers.IntegerRegisters.AllRegisters() {
		if uint32(reg.Encode()) == idx {
			return reg.Name()
		}
	}
	return fmt.Sprintf("r%d", idx)
}
