package components

import (
	"fmt"

	"github.com/Manu343726/cucaracha/pkg/hw/component"
)

func init() {
	registerCPUComponents()
}

const CategoryCPU = "cpu"

func registerCPUComponents() {
	Registry.Register(component.NewDescriptor("CPU").
		DisplayName("Cucaracha CPU").
		Description("Complete Cucaracha CPU with all internal components").
		Category(CategoryCPU).
		Version("1.0.0").
		Param("memorySize", "int", 65536, "Memory size in bytes").
		Param("numRegisters", "int", 256, "Number of general purpose registers").
		Factory(func(name string, params map[string]interface{}) (component.Component, error) {
			memSize := getIntParam(params, "memorySize", 65536)
			numRegs := getIntParam(params, "numRegisters", 256)
			return NewCPU(name, memSize, numRegs), nil
		}).
		Build())
}

// =============================================================================
// CPU Component
// =============================================================================

// CPU is a complete Cucaracha processor
type CPU struct {
	*component.BaseComponent

	// Internal components
	pc          *ProgramCounter
	ir          *Reg32 // Instruction Register
	decoder     *InstructionDecoder
	controlUnit *ControlUnit
	alu         *ALU
	registers   *RegisterBank
	memory      *RAM

	// Internal buses/connections
	aluResult uint32
	memData   uint32

	// Configuration
	memorySize   int
	numRegisters int

	// Cycle counter
	cycles uint64
}

// NewCPU creates a new CPU with the specified memory size and register count
func NewCPU(name string, memorySize, numRegisters int) *CPU {
	cpu := &CPU{
		BaseComponent: component.NewBaseComponent(name, "CPU"),
		memorySize:    memorySize,
		numRegisters:  numRegisters,
	}

	// Create internal components
	cpu.pc = NewProgramCounter("PC", 4, 0)
	cpu.ir = NewRegister("IR", 32)
	cpu.decoder = NewInstructionDecoder("Decoder")
	cpu.controlUnit = NewControlUnit("ControlUnit")
	cpu.alu = NewALU("ALU")
	cpu.registers = NewRegisterBank("Registers", numRegisters, 32)
	cpu.memory = NewRAM("Memory", memorySize)

	return cpu
}

// Component accessors
func (cpu *CPU) PC() *ProgramCounter          { return cpu.pc }
func (cpu *CPU) IR() *Reg32                   { return cpu.ir }
func (cpu *CPU) Decoder() *InstructionDecoder { return cpu.decoder }
func (cpu *CPU) ControlUnit() *ControlUnit    { return cpu.controlUnit }
func (cpu *CPU) ALU() *ALU                    { return cpu.alu }
func (cpu *CPU) Registers() *RegisterBank     { return cpu.registers }
func (cpu *CPU) Memory() *RAM                 { return cpu.memory }

// Cycles returns the number of clock cycles executed
func (cpu *CPU) Cycles() uint64 { return cpu.cycles }

// IsHalted returns true if the CPU is halted
func (cpu *CPU) IsHalted() bool { return cpu.controlUnit.IsHalted() }

// Halt halts the CPU
func (cpu *CPU) Halt() { cpu.controlUnit.Halt() }

// GetRegister returns the value of a register by index
func (cpu *CPU) GetRegister(idx int) uint32 {
	return uint32(cpu.registers.Get(idx))
}

// SetRegister sets the value of a register by index
func (cpu *CPU) SetRegister(idx int, value uint32) {
	cpu.registers.Set(idx, uint64(value))
}

// GetPC returns the current program counter
func (cpu *CPU) GetPC() uint32 {
	return cpu.pc.Value()
}

// SetPC sets the program counter
func (cpu *CPU) SetPC(value uint32) {
	cpu.pc.SetValue(value)
}

// ReadMemory reads a 32-bit word from memory
func (cpu *CPU) ReadMemory(addr uint32) uint32 {
	if int(addr+3) >= cpu.memorySize {
		return 0
	}
	// Read 4 bytes little-endian
	return uint32(cpu.memory.ReadByte(int(addr))) |
		uint32(cpu.memory.ReadByte(int(addr+1)))<<8 |
		uint32(cpu.memory.ReadByte(int(addr+2)))<<16 |
		uint32(cpu.memory.ReadByte(int(addr+3)))<<24
}

// WriteMemory writes a 32-bit word to memory
func (cpu *CPU) WriteMemory(addr uint32, value uint32) {
	if int(addr+3) >= cpu.memorySize {
		return
	}
	// Write 4 bytes little-endian
	cpu.memory.WriteByte(int(addr), byte(value))
	cpu.memory.WriteByte(int(addr+1), byte(value>>8))
	cpu.memory.WriteByte(int(addr+2), byte(value>>16))
	cpu.memory.WriteByte(int(addr+3), byte(value>>24))
}

// LoadProgram loads a program (slice of 32-bit instructions) into memory
func (cpu *CPU) LoadProgram(program []uint32, startAddr uint32) error {
	for i, instr := range program {
		addr := startAddr + uint32(i*4)
		if int(addr+3) >= cpu.memorySize {
			return fmt.Errorf("program too large for memory")
		}
		cpu.WriteMemory(addr, instr)
	}
	cpu.SetPC(startAddr)
	return nil
}

// LoadBinary loads raw binary data into memory
func (cpu *CPU) LoadBinary(data []byte, startAddr uint32) error {
	if int(startAddr)+len(data) > cpu.memorySize {
		return fmt.Errorf("data too large for memory")
	}
	for i, b := range data {
		cpu.memory.WriteByte(int(startAddr)+i, b)
	}
	cpu.SetPC(startAddr)
	return nil
}

// Step executes one CPU cycle (one FSM state transition)
func (cpu *CPU) Step() error {
	if !cpu.IsEnabled() || cpu.IsHalted() {
		return nil
	}

	state := cpu.controlUnit.CurrentState()

	switch state {
	case State_Fetch:
		cpu.doFetch()

	case State_Decode:
		cpu.doDecode()

	case State_Execute:
		cpu.doExecute()

	case State_MemRead:
		cpu.doMemRead()

	case State_MemWrite:
		cpu.doMemWrite()

	case State_WriteBack:
		cpu.doWriteBack()
	}

	// Advance control unit
	cpu.controlUnit.Clock()
	cpu.cycles++

	return nil
}

// StepInstruction executes one complete instruction (multiple cycles)
func (cpu *CPU) StepInstruction() error {
	if cpu.IsHalted() {
		return fmt.Errorf("CPU is halted")
	}

	// Execute until we return to Fetch state
	startState := cpu.controlUnit.CurrentState()
	for {
		if err := cpu.Step(); err != nil {
			return err
		}
		// If we've completed a cycle back to Fetch (or halted)
		if cpu.controlUnit.CurrentState() == State_Fetch && startState != State_Fetch {
			break
		}
		if cpu.controlUnit.CurrentState() == State_Fetch && startState == State_Fetch {
			// Single-cycle instruction (like NOP)
			if err := cpu.Step(); err != nil {
				return err
			}
			break
		}
		if cpu.IsHalted() {
			break
		}
	}
	return nil
}

// Run executes instructions until halted
func (cpu *CPU) Run() error {
	for !cpu.IsHalted() {
		if err := cpu.StepInstruction(); err != nil {
			return err
		}
	}
	return nil
}

// RunN executes at most n instructions
func (cpu *CPU) RunN(n int) error {
	for i := 0; i < n && !cpu.IsHalted(); i++ {
		if err := cpu.StepInstruction(); err != nil {
			return err
		}
	}
	return nil
}

// doFetch fetches the instruction at PC
func (cpu *CPU) doFetch() {
	pc := cpu.pc.Value()
	instruction := cpu.ReadMemory(pc)
	cpu.ir.SetValue(uint64(instruction))
}

// doDecode decodes the instruction in IR
func (cpu *CPU) doDecode() {
	instruction := cpu.ir.Value()
	cpu.decoder.Decode(uint32(instruction))

	// Set up control unit with opcode
	cpu.controlUnit.Opcode().SetValue(uint64(cpu.decoder.GetOpcode()))
}

// doExecute executes the instruction
func (cpu *CPU) doExecute() {
	opcode := cpu.decoder.GetOpcode()
	op1 := cpu.decoder.GetOp1()
	op2 := cpu.decoder.GetOp2()
	op3 := cpu.decoder.GetOp3()
	imm16 := cpu.decoder.GetImm16()

	switch opcode {
	case OP_NOP:
		cpu.pc.Increment()

	case OP_MOV_IMM16L:
		// Load lower 16 bits of immediate into register
		// Format: imm16 (bits 5-20), dst (bits 21-28) => dst is op3
		dst := op3
		cpu.registers.Set(int(dst), uint64(imm16))

	case OP_MOV_IMM16H:
		// Load upper 16 bits of immediate into register (preserve low 16)
		// Format: imm16 (bits 5-20), dst (bits 21-28) => dst is op3
		dst := op3
		current := cpu.registers.Get(int(dst))
		cpu.registers.Set(int(dst), (current&0xFFFF)|(uint64(imm16)<<16))

	case OP_MOV:
		// Copy register to register
		src := op1
		dst := op2
		cpu.registers.Set(int(dst), cpu.registers.Get(int(src)))

	case OP_ADD, OP_SUB, OP_MUL, OP_DIV, OP_MOD, OP_LSL, OP_LSR, OP_ASL, OP_ASR:
		// Binary ALU operations
		src1 := uint32(cpu.registers.Get(int(op1)))
		src2 := uint32(cpu.registers.Get(int(op2)))
		cpu.alu.SetOperands(src1, src2)
		cpu.alu.SetOperation(opcodeToALUOp(opcode))
		cpu.alu.Compute()
		cpu.aluResult = cpu.alu.Result()

	case OP_CMP:
		// Compare (set flags only)
		src1 := uint32(cpu.registers.Get(int(op1)))
		src2 := uint32(cpu.registers.Get(int(op2)))
		cpu.alu.SetOperands(src1, src2)
		cpu.alu.SetOperation(ALUOp_CMP)
		cpu.alu.Compute()
		// Store flags to destination register
		dst := op3
		flags := cpu.alu.GetFlags()
		cpu.registers.Set(int(dst), uint64(flags))
		// Also write to implicit CPSR register (index 2) for CJMP to read
		cpu.registers.Set(2, uint64(flags))

	case OP_LD:
		// Load: address is in src register
		addr := uint32(cpu.registers.Get(int(op1)))
		cpu.aluResult = addr // Store address for mem read stage

	case OP_ST:
		// Store: value in src, address in addr register
		cpu.aluResult = uint32(cpu.registers.Get(int(op2))) // Address

	case OP_JMP:
		// Unconditional jump
		target := uint32(cpu.registers.Get(int(op1)))
		link := op2
		cpu.registers.Set(int(link), uint64(cpu.pc.Value()+4))
		cpu.pc.Load(target)
		return // Don't increment PC

	case OP_CJMP:
		// Conditional jump
		condReg := uint32(cpu.registers.Get(int(op1)))
		target := uint32(cpu.registers.Get(int(op2)))
		link := op3
		// Use ALU flags stored in cpsr register (index 2)
		flags := uint32(cpu.registers.Get(2))
		cpu.controlUnit.Flags().SetValue(uint64(flags))
		cpu.controlUnit.Cond().SetValue(uint64(condReg))

		if cpu.controlUnit.testCondition(flags, uint8(condReg)) {
			cpu.registers.Set(int(link), uint64(cpu.pc.Value()+4))
			cpu.pc.Load(target)
		} else {
			// Condition not met - increment PC to next instruction
			cpu.pc.Increment()
		}
		return // Don't fall through to default PC increment
	}
}

// doMemRead reads data from memory
func (cpu *CPU) doMemRead() {
	addr := cpu.aluResult
	cpu.memData = cpu.ReadMemory(addr)
}

// doMemWrite writes data to memory
func (cpu *CPU) doMemWrite() {
	opcode := cpu.decoder.GetOpcode()
	if opcode == OP_ST {
		op1 := cpu.decoder.GetOp1()
		value := uint32(cpu.registers.Get(int(op1)))
		addr := cpu.aluResult
		cpu.WriteMemory(addr, value)
	}
	cpu.pc.Increment()
}

// doWriteBack writes result to destination register
func (cpu *CPU) doWriteBack() {
	opcode := cpu.decoder.GetOpcode()
	op3 := cpu.decoder.GetOp3()
	op2 := cpu.decoder.GetOp2()

	switch opcode {
	case OP_LD:
		// Write memory data to destination register
		cpu.registers.Set(int(op2), uint64(cpu.memData))

	case OP_ADD, OP_SUB, OP_MUL, OP_DIV, OP_MOD, OP_LSL, OP_LSR, OP_ASL, OP_ASR:
		// Write ALU result to destination register
		cpu.registers.Set(int(op3), uint64(cpu.aluResult))
	}

	cpu.pc.Increment()
}

// opcodeToALUOp converts instruction opcode to ALU operation
func opcodeToALUOp(opcode uint8) ALUOp {
	switch opcode {
	case OP_ADD:
		return ALUOp_ADD
	case OP_SUB:
		return ALUOp_SUB
	case OP_MUL:
		return ALUOp_MUL
	case OP_DIV:
		return ALUOp_DIV
	case OP_MOD:
		return ALUOp_MOD
	case OP_LSL:
		return ALUOp_LSL
	case OP_LSR:
		return ALUOp_LSR
	case OP_ASL:
		return ALUOp_ASL
	case OP_ASR:
		return ALUOp_ASR
	case OP_CMP:
		return ALUOp_CMP
	default:
		return ALUOp_NOP
	}
}

// Reset resets the CPU to initial state
func (cpu *CPU) Reset() {
	cpu.pc.Reset()
	cpu.ir.Reset()
	cpu.decoder.Reset()
	cpu.controlUnit.Reset()
	cpu.alu.Reset()
	cpu.registers.Reset()
	cpu.memory.Reset()
	cpu.cycles = 0
	cpu.aluResult = 0
	cpu.memData = 0
}

// EncodeInstruction encodes an instruction from opcode and operands
// This is a helper for testing
func EncodeInstruction(opcode uint8, op1, op2, op3 uint8) uint32 {
	return uint32(opcode) |
		uint32(op1)<<5 |
		uint32(op2)<<13 |
		uint32(op3)<<21
}

// EncodeImmInstruction encodes an instruction with 16-bit immediate
func EncodeImmInstruction(opcode uint8, imm16 uint16, dst uint8) uint32 {
	return uint32(opcode) |
		uint32(imm16)<<5 |
		uint32(dst)<<21
}
