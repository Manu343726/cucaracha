// Package cpu provides the Cucaracha CPU abstraction and implementations.
package cpu

// CPU is the core interface that abstracts a Cucaracha processor.
// It can be implemented by both software emulators and hardware-simulated CPUs.
type CPU interface {
	// Register operations
	GetRegister(idx int) uint32
	SetRegister(idx int, value uint32)

	// Program counter operations
	GetPC() uint32
	SetPC(value uint32)

	// Stack pointer operations (convenience methods for SP register)
	GetSP() uint32
	SetSP(value uint32)

	// Link register operations (convenience methods for LR register)
	GetLR() uint32
	SetLR(value uint32)

	// Memory operations
	ReadMemory(addr uint32) uint32
	WriteMemory(addr uint32, value uint32)
	ReadByte(addr uint32) byte
	WriteByte(addr uint32, value byte)

	// Memory size
	MemorySize() int

	// Program loading
	LoadBinary(data []byte, startAddr uint32) error
	LoadProgram(program []uint32, startAddr uint32) error

	// Execution
	Step() error  // Execute one instruction
	Run() error   // Run until halted
	RunN(n int) error // Run at most n instructions

	// State
	IsHalted() bool
	Halt()
	Reset()

	// Cycles executed
	Cycles() uint64
}

// DebuggableCPU extends CPU with debugging capabilities.
// This interface is used by the debugger to implement breakpoints,
// watchpoints, and execution control.
type DebuggableCPU interface {
	CPU

	// DecodeInstruction decodes the instruction at the given address
	DecodeInstruction(addr uint32) (mnemonic string, operands string, err error)

	// GetFlags returns the CPU flags register
	GetFlags() uint32
	SetFlags(flags uint32)
}

// CPUFactory creates CPU instances
type CPUFactory func(memorySize int) CPU

// Standard register indices used by Cucaracha ISA
const (
	// Control registers (0-15)
	RegPC   = 0  // Program Counter
	RegSP   = 1  // Stack Pointer
	RegCPSR = 2  // Current Program Status Register (flags)
	RegLR   = 3  // Link Register

	// General purpose registers start at index 16
	RegR0 = 16
	RegR1 = 17
	RegR2 = 18
	RegR3 = 19
	RegR4 = 20
	RegR5 = 21
	RegR6 = 22
	RegR7 = 23
	RegR8 = 24
	RegR9 = 25
)

// Flag bits in CPSR
const (
	FlagN = 1 << 3 // Negative
	FlagZ = 1 << 2 // Zero
	FlagC = 1 << 1 // Carry
	FlagV = 1 << 0 // Overflow
)
