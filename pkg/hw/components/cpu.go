package components

import (
	"fmt"

	"github.com/Manu343726/cucaracha/pkg/hw/component"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
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
			cpu := NewCPU(name, memSize, numRegs)
			ram := NewRAM("Memory", memSize)
			cpu.AttachMemory(ram)
			return cpu, nil
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
	memory      Memory

	// External memory interface (bus-like signals)
	memDataPort    *component.StandardPort
	memAddressPort *component.StandardPort
	memReadWrite   *component.Pin
	memReady       *component.Pin

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

	// Memory interface ports (bidirectional data + address/control pins)
	cpu.memDataPort = component.NewPort("MEM_DATA", 32)
	_ = cpu.AddInput(cpu.memDataPort)
	_ = cpu.AddOutput(cpu.memDataPort)

	cpu.memAddressPort = component.NewOutputPort("MEM_ADDR", 32)
	_ = cpu.AddOutput(cpu.memAddressPort)

	cpu.memReadWrite = component.NewOutputPin("MEM_RW")
	_ = cpu.AddOutput(cpu.memReadWrite)

	cpu.memReady = component.NewInputPin("MEM_READY")
	_ = cpu.AddInput(cpu.memReady)

	return cpu
}

// Component accessors
func (cpu *CPU) PC() *ProgramCounter          { return cpu.pc }
func (cpu *CPU) IR() *Reg32                   { return cpu.ir }
func (cpu *CPU) Decoder() *InstructionDecoder { return cpu.decoder }
func (cpu *CPU) ControlUnit() *ControlUnit    { return cpu.controlUnit }
func (cpu *CPU) ALU() *ALU                    { return cpu.alu }
func (cpu *CPU) Registers() *RegisterBank     { return cpu.registers }
func (cpu *CPU) Memory() Memory               { return cpu.memory }

// External memory bus ports (for wiring to an external memory component)
func (cpu *CPU) MemDataPort() *component.StandardPort    { return cpu.memDataPort }
func (cpu *CPU) MemAddressPort() *component.StandardPort { return cpu.memAddressPort }
func (cpu *CPU) MemReadWritePin() *component.Pin         { return cpu.memReadWrite }
func (cpu *CPU) MemReadyPin() *component.Pin             { return cpu.memReady }

// AttachMemory connects an external memory device to the CPU.
func (cpu *CPU) AttachMemory(memory Memory) {
	cpu.memory = memory
	if memory != nil {
		cpu.memorySize = memory.Size()
	}
}

// Cycles returns the number of clock cycles executed
func (cpu *CPU) Cycles() uint64 { return cpu.cycles }

// IsHalted returns true if the CPU is halted
func (cpu *CPU) IsHalted() bool { return cpu.controlUnit.IsHalted() }

// Halt halts the CPU
func (cpu *CPU) Halt() { cpu.controlUnit.Halt() }

// GetRegister returns the value of a register by index
func (cpu *CPU) GetRegister(idx uint32) uint32 {
	return uint32(cpu.registers.Get(idx))
}

// SetRegister sets the value of a register by index
func (cpu *CPU) SetRegister(idx uint32, value uint32) {
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
	if cpu.memory == nil {
		return 0
	}
	_ = cpu.memAddressPort.SetValue(uint64(addr))
	_ = cpu.memReadWrite.Set(component.Low)
	if int(addr+3) >= cpu.memorySize {
		return 0
	}
	// Read 4 bytes little-endian
	value := uint32(cpu.memory.ReadByte(addr)) |
		uint32(cpu.memory.ReadByte(addr+1))<<8 |
		uint32(cpu.memory.ReadByte(addr+2))<<16 |
		uint32(cpu.memory.ReadByte(addr+3))<<24
	_ = cpu.memDataPort.SetValue(uint64(value))
	return value
}

// WriteMemory writes a 32-bit word to memory
func (cpu *CPU) WriteMemory(addr uint32, value uint32) {
	if cpu.memory == nil {
		return
	}
	_ = cpu.memAddressPort.SetValue(uint64(addr))
	_ = cpu.memDataPort.SetValue(uint64(value))
	_ = cpu.memReadWrite.Set(component.High)
	if int(addr+3) >= cpu.memorySize {
		return
	}
	// Write 4 bytes little-endian
	cpu.memory.WriteByte(addr, byte(value))
	cpu.memory.WriteByte(addr+1, byte(value>>8))
	cpu.memory.WriteByte(addr+2, byte(value>>16))
	cpu.memory.WriteByte(addr+3, byte(value>>24))
}

// LoadProgram loads a program (slice of 32-bit instructions) into memory
func (cpu *CPU) LoadProgram(program []uint32, startAddr uint32) error {
	if cpu.memory == nil {
		return fmt.Errorf("no memory attached to CPU")
	}
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
	if cpu.memory == nil {
		return fmt.Errorf("no memory attached to CPU")
	}
	if int(startAddr)+len(data) > cpu.memorySize {
		return fmt.Errorf("data too large for memory")
	}
	for i, b := range data {
		cpu.memory.WriteByte(startAddr+uint32(i), b)
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

func (cpu *CPU) executeNOP(opcode instructions.OpCode, op1, op2, op3 uint64, imm16 uint16) {
	// NOP - just increment PC (NOP goes directly to Fetch, skipping WriteBack)
	cpu.pc.Increment()
}

func (cpu *CPU) executeMOVImm16L(opcode instructions.OpCode, op1, op2, op3 uint64, imm16 uint16) {
	// MOVIMM16L: operands are (imm16, dst)
	// imm16 is passed separately, dst is in op1 (first register operand)
	dst := uint32(op1)
	cpu.registers.Set(dst, uint64(imm16))
}

func (cpu *CPU) executeMOVImm16H(opcode instructions.OpCode, op1, op2, op3 uint64, imm16 uint16) {
	// MOVIMM16H: operands are (imm16, dst, src) where src is tied to dst
	// imm16 is passed separately, dst is in op1, src is in op2
	dst := uint32(op1)
	current := cpu.registers.Get(dst)
	cpu.registers.Set(dst, (current&0xFFFF)|(uint64(imm16)<<16))
}

func (cpu *CPU) executeMOV(opcode instructions.OpCode, op1, op2, op3 uint64, imm16 uint16) {
	src := uint32(op1)
	dst := uint32(op2)
	cpu.registers.Set(dst, cpu.registers.Get(src))
}

func (cpu *CPU) executeALUOp(opcode instructions.OpCode, op1, op2, op3 uint64, imm16 uint16) {
	// ALU ops: operands are (src1, src2, dst)
	// op1 = src1, op2 = src2, op3 = dst
	src1 := uint32(op1)
	src2 := uint32(op2)
	srcVal1 := uint32(cpu.registers.Get(src1))
	srcVal2 := uint32(cpu.registers.Get(src2))
	cpu.alu.SetOperands(srcVal1, srcVal2)
	cpu.alu.SetOperation(opcodeToALUOp(opcode))
	cpu.alu.Compute()
	cpu.aluResult = cpu.alu.Result()
}

func (cpu *CPU) executeCMP(opcode instructions.OpCode, op1, op2, op3 uint64, imm16 uint16) {
	// CMP: operands are (lhs, rhs, dst)
	// op1 = lhs, op2 = rhs, op3 = dst
	src1 := uint32(op1)
	src2 := uint32(op2)
	dst := uint32(op3)
	srcVal1 := uint32(cpu.registers.Get(src1))
	srcVal2 := uint32(cpu.registers.Get(src2))
	cpu.alu.SetOperands(srcVal1, srcVal2)
	cpu.alu.SetOperation(ALUOp_CMP)
	cpu.alu.Compute()
	// Store flags to destination register
	flags := cpu.alu.GetFlags()
	cpu.registers.Set(dst, uint64(flags))
	// Also write to implicit CPSR register (index 2) for CJMP to read
	cpu.registers.Set(2, uint64(flags))
}

func (cpu *CPU) executeLD(opcode instructions.OpCode, op1, op2, op3 uint64, imm16 uint16) {
	addrReg := uint32(op1)
	addr := uint32(cpu.registers.Get(addrReg))
	cpu.aluResult = addr // Store address for mem read stage
}

func (cpu *CPU) executeST(opcode instructions.OpCode, op1, op2, op3 uint64, imm16 uint16) {
	// op1 = src register (value), op2 = addr register
	addrReg := uint32(op2)
	addr := uint32(cpu.registers.Get(addrReg))
	cpu.aluResult = addr // Store address for mem write stage
}

func (cpu *CPU) executeJMP(opcode instructions.OpCode, op1, op2, op3 uint64, imm16 uint16) {
	targetReg := uint32(op1)
	linkReg := uint32(op2)
	target := uint32(cpu.registers.Get(targetReg))
	link := linkReg
	cpu.registers.Set(link, uint64(cpu.pc.Value()+4))
	cpu.pc.Load(target)
}

func (cpu *CPU) executeCJMP(opcode instructions.OpCode, op1, op2, op3 uint64, imm16 uint16) {
	condReg := uint32(op1)
	targetReg := uint32(op2)
	linkReg := uint32(op3)
	cond := uint8(cpu.registers.Get(condReg))
	target := uint32(cpu.registers.Get(targetReg))
	link := linkReg
	// Use ALU flags stored in cpsr register (index 2)
	flags := uint32(cpu.registers.Get(2))
	cpu.controlUnit.Flags().SetValue(uint64(flags))
	cpu.controlUnit.Cond().SetValue(uint64(cond))

	if cpu.controlUnit.testCondition(flags, cond) {
		cpu.registers.Set(link, uint64(cpu.pc.Value()+4))
		cpu.pc.Load(target)
	} else {
		// Condition not met - increment PC to next instruction
		cpu.pc.Increment()
	}
}

// doExecute executes the instruction
func (cpu *CPU) doExecute() {
	opcode := cpu.decoder.GetOpcode()
	op1 := cpu.decoder.GetOp1()
	op2 := cpu.decoder.GetOp2()
	op3 := cpu.decoder.GetOp3()
	imm16 := cpu.decoder.GetImm16()

	switch opcode {
	case instructions.OpCode_NOP:
		cpu.executeNOP(opcode, op1, op2, op3, imm16)
	case instructions.OpCode_MOV_IMM16L:
		cpu.executeMOVImm16L(opcode, op1, op2, op3, imm16)
	case instructions.OpCode_MOV_IMM16H:
		cpu.executeMOVImm16H(opcode, op1, op2, op3, imm16)
	case instructions.OpCode_MOV:
		cpu.executeMOV(opcode, op1, op2, op3, imm16)
	case instructions.OpCode_ADD, instructions.OpCode_SUB, instructions.OpCode_MUL, instructions.OpCode_DIV, instructions.OpCode_MOD, instructions.OpCode_LSL, instructions.OpCode_LSR, instructions.OpCode_ASL, instructions.OpCode_ASR:
		cpu.executeALUOp(opcode, op1, op2, op3, imm16)
	case instructions.OpCode_CMP:
		cpu.executeCMP(opcode, op1, op2, op3, imm16)
	case instructions.OpCode_LD:
		cpu.executeLD(opcode, op1, op2, op3, imm16)
	case instructions.OpCode_ST:
		cpu.executeST(opcode, op1, op2, op3, imm16)
	case instructions.OpCode_JMP:
		cpu.executeJMP(opcode, op1, op2, op3, imm16)
	case instructions.OpCode_CJMP:
		cpu.executeCJMP(opcode, op1, op2, op3, imm16)
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
	if opcode == instructions.OpCode_ST {
		op1 := cpu.decoder.GetOp1()
		value := uint32(cpu.registers.Get(uint32(op1)))
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
	case instructions.OpCode_LD:
		// Write memory data to destination register
		cpu.registers.Set(uint32(op2), uint64(cpu.memData))

	case instructions.OpCode_ADD, instructions.OpCode_SUB, instructions.OpCode_MUL, instructions.OpCode_DIV, instructions.OpCode_MOD, instructions.OpCode_LSL, instructions.OpCode_LSR, instructions.OpCode_ASL, instructions.OpCode_ASR:
		// Write ALU result to destination register
		cpu.registers.Set(uint32(op3), uint64(cpu.aluResult))
	}

	cpu.pc.Increment()
}

// opcodeToALUOp converts instruction opcode to ALU operation
func opcodeToALUOp(opcode instructions.OpCode) ALUOp {
	switch opcode {
	case instructions.OpCode_ADD:
		return ALUOp_ADD
	case instructions.OpCode_SUB:
		return ALUOp_SUB
	case instructions.OpCode_MUL:
		return ALUOp_MUL
	case instructions.OpCode_DIV:
		return ALUOp_DIV
	case instructions.OpCode_MOD:
		return ALUOp_MOD
	case instructions.OpCode_LSL:
		return ALUOp_LSL
	case instructions.OpCode_LSR:
		return ALUOp_LSR
	case instructions.OpCode_ASL:
		return ALUOp_ASL
	case instructions.OpCode_ASR:
		return ALUOp_ASR
	case instructions.OpCode_CMP:
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
	if cpu.memory != nil {
		cpu.memory.Reset()
	}
	cpu.memDataPort.Reset()
	cpu.memAddressPort.Reset()
	cpu.memReadWrite.Reset()
	cpu.memReady.Reset()
	cpu.cycles = 0
	cpu.aluResult = 0
	cpu.memData = 0
}
