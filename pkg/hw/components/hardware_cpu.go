package components

import (
	"fmt"

	hwcpu "github.com/Manu343726/cucaracha/pkg/hw/cpu"
)

// HardwareCPU wraps the hardware-based CPU component to implement
// the cpu.CPU and cpu.DebuggableCPU interfaces.
// This allows the component-based CPU to be used interchangeably
// with the software emulator.
type HardwareCPU struct {
	cpu *CPU
}

// NewHardwareCPU creates a new hardware CPU adapter with the given memory size
func NewHardwareCPU(memorySize int) *HardwareCPU {
	return &HardwareCPU{
		cpu: NewCPU("CPU", memorySize, 256),
	}
}

// NewHardwareCPUWithComponent wraps an existing CPU component
func NewHardwareCPUWithComponent(cpu *CPU) *HardwareCPU {
	return &HardwareCPU{cpu: cpu}
}

// CPU returns the underlying CPU component for advanced use cases
func (h *HardwareCPU) Component() *CPU {
	return h.cpu
}

// --- hwcpu.CPU interface implementation ---

// GetRegister returns the value of a register by index
func (h *HardwareCPU) GetRegister(idx int) uint32 {
	return h.cpu.GetRegister(idx)
}

// SetRegister sets the value of a register by index
func (h *HardwareCPU) SetRegister(idx int, value uint32) {
	h.cpu.SetRegister(idx, value)
}

// GetPC returns the current program counter
func (h *HardwareCPU) GetPC() uint32 {
	return h.cpu.GetPC()
}

// SetPC sets the program counter
func (h *HardwareCPU) SetPC(value uint32) {
	h.cpu.SetPC(value)
}

// GetSP returns the stack pointer
func (h *HardwareCPU) GetSP() uint32 {
	return h.GetRegister(hwcpu.RegSP)
}

// SetSP sets the stack pointer
func (h *HardwareCPU) SetSP(value uint32) {
	h.SetRegister(hwcpu.RegSP, value)
}

// GetLR returns the link register
func (h *HardwareCPU) GetLR() uint32 {
	return h.GetRegister(hwcpu.RegLR)
}

// SetLR sets the link register
func (h *HardwareCPU) SetLR(value uint32) {
	h.SetRegister(hwcpu.RegLR, value)
}

// ReadMemory reads a 32-bit word from memory
func (h *HardwareCPU) ReadMemory(addr uint32) uint32 {
	return h.cpu.ReadMemory(addr)
}

// WriteMemory writes a 32-bit word to memory
func (h *HardwareCPU) WriteMemory(addr uint32, value uint32) {
	h.cpu.WriteMemory(addr, value)
}

// ReadByte reads a single byte from memory
func (h *HardwareCPU) ReadByte(addr uint32) byte {
	return h.cpu.Memory().ReadByte(int(addr))
}

// WriteByte writes a single byte to memory
func (h *HardwareCPU) WriteByte(addr uint32, value byte) {
	h.cpu.Memory().WriteByte(int(addr), value)
}

// MemorySize returns the size of memory in bytes
func (h *HardwareCPU) MemorySize() int {
	return h.cpu.memorySize
}

// LoadBinary loads binary machine code into memory at the given address
func (h *HardwareCPU) LoadBinary(data []byte, startAddr uint32) error {
	return h.cpu.LoadBinary(data, startAddr)
}

// LoadProgram loads a program (slice of 32-bit instructions) into memory
func (h *HardwareCPU) LoadProgram(program []uint32, startAddr uint32) error {
	return h.cpu.LoadProgram(program, startAddr)
}

// Step executes one instruction
func (h *HardwareCPU) Step() error {
	return h.cpu.StepInstruction()
}

// Run executes instructions until halted
func (h *HardwareCPU) Run() error {
	return h.cpu.Run()
}

// RunN executes at most n instructions
func (h *HardwareCPU) RunN(n int) error {
	return h.cpu.RunN(n)
}

// IsHalted returns true if the CPU is halted
func (h *HardwareCPU) IsHalted() bool {
	return h.cpu.IsHalted()
}

// Halt halts the CPU
func (h *HardwareCPU) Halt() {
	h.cpu.Halt()
}

// Reset resets the CPU state
func (h *HardwareCPU) Reset() {
	h.cpu.Reset()
}

// Cycles returns the number of cycles executed
func (h *HardwareCPU) Cycles() uint64 {
	return h.cpu.Cycles()
}

// --- hwcpu.DebuggableCPU interface implementation ---

// DecodeInstruction decodes the instruction at the given address
func (h *HardwareCPU) DecodeInstruction(addr uint32) (mnemonic string, operands string, err error) {
	// Read instruction from memory
	instr := h.cpu.ReadMemory(addr)

	// Extract opcode (bits 0-4)
	opcode := uint8(instr & 0x1F)

	// Map opcode to mnemonic
	mnemonic = opcodeToMnemonic(opcode)

	// Extract operands based on instruction format
	op1 := (instr >> 5) & 0xFF
	op2 := (instr >> 13) & 0xFF
	op3 := (instr >> 21) & 0xFF
	imm16 := uint16((instr >> 5) & 0xFFFF)

	// Format operands based on instruction type
	operands = formatOperands(opcode, op1, op2, op3, imm16)

	return mnemonic, operands, nil
}

// GetFlags returns the CPU flags register (CPSR)
func (h *HardwareCPU) GetFlags() uint32 {
	return h.GetRegister(hwcpu.RegCPSR)
}

// SetFlags sets the CPU flags register (CPSR)
func (h *HardwareCPU) SetFlags(flags uint32) {
	h.SetRegister(hwcpu.RegCPSR, flags)
}

// opcodeToMnemonic maps opcodes to mnemonics
func opcodeToMnemonic(opcode uint8) string {
	mnemonics := map[uint8]string{
		OP_NOP:        "nop",
		OP_MOV_IMM16L: "mov_imm16l",
		OP_MOV_IMM16H: "mov_imm16h",
		OP_MOV:        "mov",
		OP_ADD:        "add",
		OP_SUB:        "sub",
		OP_MUL:        "mul",
		OP_DIV:        "div",
		OP_MOD:        "mod",
		OP_LSL:        "lsl",
		OP_LSR:        "lsr",
		OP_ASL:        "asl",
		OP_ASR:        "asr",
		OP_CMP:        "cmp",
		OP_LD:         "ld",
		OP_ST:         "st",
		OP_JMP:        "jmp",
		OP_CJMP:       "cjmp",
	}
	if m, ok := mnemonics[opcode]; ok {
		return m
	}
	return fmt.Sprintf("unknown(%d)", opcode)
}

// formatOperands formats instruction operands based on opcode
func formatOperands(opcode uint8, op1, op2, op3 uint32, imm16 uint16) string {
	switch opcode {
	case OP_NOP:
		return ""
	case OP_MOV_IMM16L, OP_MOV_IMM16H:
		return fmt.Sprintf("r%d, #%d", op3, imm16)
	case OP_MOV:
		return fmt.Sprintf("r%d, r%d", op2, op1)
	case OP_ADD, OP_SUB, OP_MUL, OP_DIV, OP_MOD, OP_LSL, OP_LSR, OP_ASL, OP_ASR:
		return fmt.Sprintf("r%d, r%d, r%d", op3, op1, op2)
	case OP_CMP:
		return fmt.Sprintf("r%d, r%d, r%d", op3, op1, op2)
	case OP_LD:
		return fmt.Sprintf("r%d, [r%d]", op2, op1)
	case OP_ST:
		return fmt.Sprintf("[r%d], r%d", op2, op1)
	case OP_JMP:
		return fmt.Sprintf("r%d, r%d", op1, op2)
	case OP_CJMP:
		return fmt.Sprintf("r%d, r%d, r%d", op1, op2, op3)
	default:
		return fmt.Sprintf("r%d, r%d, r%d", op1, op2, op3)
	}
}

// Ensure HardwareCPU implements the interfaces
var _ hwcpu.CPU = (*HardwareCPU)(nil)
var _ hwcpu.DebuggableCPU = (*HardwareCPU)(nil)

// NewCPUAdapter is a factory function that creates a HardwareCPU implementing hwcpu.CPU
func NewCPUAdapter(memorySize int) hwcpu.CPU {
	return NewHardwareCPU(memorySize)
}
