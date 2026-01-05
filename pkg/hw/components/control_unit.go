package components

import (
	"github.com/Manu343726/cucaracha/pkg/hw/component"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
)

func init() {
	registerControlUnitComponents()
}

// CPU States for the control unit FSM
type CPUState uint8

const (
	State_Fetch CPUState = iota
	State_Decode
	State_Execute
	State_MemRead
	State_MemWrite
	State_WriteBack
	State_Halt
)

func (s CPUState) String() string {
	switch s {
	case State_Fetch:
		return "FETCH"
	case State_Decode:
		return "DECODE"
	case State_Execute:
		return "EXECUTE"
	case State_MemRead:
		return "MEM_READ"
	case State_MemWrite:
		return "MEM_WRITE"
	case State_WriteBack:
		return "WRITEBACK"
	case State_Halt:
		return "HALT"
	default:
		return "UNKNOWN"
	}
}

// Control signals for CPU components
type ControlSignals struct {
	// PC control
	PCIncrement bool // Increment PC
	PCLoad      bool // Load PC from ALU/register

	// Memory control
	MemRead  bool // Read from memory
	MemWrite bool // Write to memory

	// Register control
	RegWrite bool // Write to register bank

	// ALU control
	ALUOp   ALUOp // ALU operation
	ALUSrcB uint8 // ALU B source: 0=reg, 1=imm

	// Mux selects
	RegDstSel uint8 // Register destination select
	MemToReg  bool  // Memory to register (vs ALU result)
	BranchSel bool  // Branch select
}

func registerControlUnitComponents() {
	Registry.Register(component.NewDescriptor("CONTROL_UNIT").
		DisplayName("Control Unit").
		Description("Control Unit - FSM that generates control signals based on opcode").
		Category(CategoryControl).
		Version("1.0.0").
		Input("OPCODE", 5, "Instruction opcode").
		Input("FLAGS", 32, "ALU flags for conditional branches").
		Input("COND", 8, "Condition code for CJMP").
		Output("STATE", 3, "Current FSM state").
		Output("PC_INC", 1, "PC increment enable").
		Output("PC_LOAD", 1, "PC load enable").
		Output("MEM_READ", 1, "Memory read enable").
		Output("MEM_WRITE", 1, "Memory write enable").
		Output("REG_WRITE", 1, "Register write enable").
		Output("ALU_OP", 4, "ALU operation code").
		Output("ALU_SRC_B", 1, "ALU source B select").
		Output("MEM_TO_REG", 1, "Memory to register select").
		Output("BRANCH", 1, "Branch taken").
		Factory(func(name string, params map[string]interface{}) (component.Component, error) {
			return NewControlUnit(name), nil
		}).
		Build())
}

// =============================================================================
// Control Unit
// =============================================================================

// ControlUnit generates control signals based on the current state and opcode
type ControlUnit struct {
	*component.BaseComponent

	// Inputs
	opcode *component.StandardPort
	flags  *component.StandardPort
	cond   *component.StandardPort

	// Outputs
	state    *component.StandardPort
	pcInc    *component.Pin
	pcLoad   *component.Pin
	memRead  *component.Pin
	memWrite *component.Pin
	regWrite *component.Pin
	aluOp    *component.StandardPort
	aluSrcB  *component.Pin
	memToReg *component.Pin
	branch   *component.Pin

	// Internal state
	currentState CPUState
	halted       bool
}

// NewControlUnit creates a new control unit
func NewControlUnit(name string) *ControlUnit {
	cu := &ControlUnit{
		BaseComponent: component.NewBaseComponent(name, "CONTROL_UNIT"),
		currentState:  State_Fetch,
	}

	// Inputs
	cu.opcode = component.NewInputPort("OPCODE", 5)
	cu.AddInput(cu.opcode)

	cu.flags = component.NewInputPort("FLAGS", 32)
	cu.AddInput(cu.flags)

	cu.cond = component.NewInputPort("COND", 8)
	cu.AddInput(cu.cond)

	// Outputs
	cu.state = component.NewOutputPort("STATE", 3)
	cu.AddOutput(cu.state)

	cu.pcInc = component.NewOutputPin("PC_INC")
	cu.AddOutput(cu.pcInc)

	cu.pcLoad = component.NewOutputPin("PC_LOAD")
	cu.AddOutput(cu.pcLoad)

	cu.memRead = component.NewOutputPin("MEM_READ")
	cu.AddOutput(cu.memRead)

	cu.memWrite = component.NewOutputPin("MEM_WRITE")
	cu.AddOutput(cu.memWrite)

	cu.regWrite = component.NewOutputPin("REG_WRITE")
	cu.AddOutput(cu.regWrite)

	cu.aluOp = component.NewOutputPort("ALU_OP", 4)
	cu.AddOutput(cu.aluOp)

	cu.aluSrcB = component.NewOutputPin("ALU_SRC_B")
	cu.AddOutput(cu.aluSrcB)

	cu.memToReg = component.NewOutputPin("MEM_TO_REG")
	cu.AddOutput(cu.memToReg)

	cu.branch = component.NewOutputPin("BRANCH")
	cu.AddOutput(cu.branch)

	return cu
}

// Accessor methods
func (cu *ControlUnit) Opcode() *component.StandardPort { return cu.opcode }
func (cu *ControlUnit) Flags() *component.StandardPort  { return cu.flags }
func (cu *ControlUnit) Cond() *component.StandardPort   { return cu.cond }
func (cu *ControlUnit) State() *component.StandardPort  { return cu.state }
func (cu *ControlUnit) PCInc() *component.Pin           { return cu.pcInc }
func (cu *ControlUnit) PCLoad() *component.Pin          { return cu.pcLoad }
func (cu *ControlUnit) MemRead() *component.Pin         { return cu.memRead }
func (cu *ControlUnit) MemWrite() *component.Pin        { return cu.memWrite }
func (cu *ControlUnit) RegWrite() *component.Pin        { return cu.regWrite }
func (cu *ControlUnit) ALUOp() *component.StandardPort  { return cu.aluOp }
func (cu *ControlUnit) ALUSrcB() *component.Pin         { return cu.aluSrcB }
func (cu *ControlUnit) MemToReg() *component.Pin        { return cu.memToReg }
func (cu *ControlUnit) Branch() *component.Pin          { return cu.branch }

// CurrentState returns the current FSM state
func (cu *ControlUnit) CurrentState() CPUState { return cu.currentState }

// IsHalted returns true if the CPU is halted
func (cu *ControlUnit) IsHalted() bool { return cu.halted }

// Halt halts the control unit
func (cu *ControlUnit) Halt() {
	cu.halted = true
	cu.currentState = State_Halt
}

// clearSignals resets all control signals to default
func (cu *ControlUnit) clearSignals() {
	cu.pcInc.Set(Low)
	cu.pcLoad.Set(Low)
	cu.memRead.Set(Low)
	cu.memWrite.Set(Low)
	cu.regWrite.Set(Low)
	cu.aluOp.SetValue(uint64(ALUOp_NOP))
	cu.aluSrcB.Set(Low)
	cu.memToReg.Set(Low)
	cu.branch.Set(Low)
}

// Clock advances the control unit FSM
func (cu *ControlUnit) Clock() error {
	if !cu.IsEnabled() || cu.halted {
		return nil
	}

	cu.clearSignals()

	opcode := instructions.OpCode(cu.opcode.GetValue())

	switch cu.currentState {
	case State_Fetch:
		// Fetch: Read instruction from memory at PC
		cu.memRead.Set(High)
		cu.currentState = State_Decode

	case State_Decode:
		// Decode: Instruction is decoded combinationally
		// Determine next state based on opcode
		cu.currentState = cu.decodeNextState(opcode)

	case State_Execute:
		// Execute: Perform ALU operation or branch
		cu.generateExecuteSignals(opcode)
		cu.currentState = cu.executeNextState(opcode)

	case State_MemRead:
		// Memory read for LD instruction
		cu.memRead.Set(High)
		cu.currentState = State_WriteBack

	case State_MemWrite:
		// Memory write for ST instruction
		cu.memWrite.Set(High)
		cu.pcInc.Set(High)
		cu.currentState = State_Fetch

	case State_WriteBack:
		// Write result to register
		cu.regWrite.Set(High)
		if opcode == instructions.OpCode_LD {
			cu.memToReg.Set(High)
		}
		cu.pcInc.Set(High)
		cu.currentState = State_Fetch

	case State_Halt:
		// Do nothing
	}

	cu.state.SetValue(uint64(cu.currentState))
	return nil
}

// decodeNextState determines the next state after decode
func (cu *ControlUnit) decodeNextState(opcode instructions.OpCode) CPUState {
	switch opcode {
	case instructions.OpCode_LD:
		return State_Execute // Calculate address first
	case instructions.OpCode_ST:
		return State_Execute // Calculate address first
	default:
		return State_Execute // All instructions go through execute
	}
}

// executeNextState determines the next state after execute
func (cu *ControlUnit) executeNextState(opcode instructions.OpCode) CPUState {
	switch opcode {
	case instructions.OpCode_NOP:
		return State_Fetch // NOP completes after execute
	case instructions.OpCode_LD:
		return State_MemRead
	case instructions.OpCode_ST:
		return State_MemWrite
	case instructions.OpCode_JMP, instructions.OpCode_CJMP:
		return State_Fetch // Branch completes in execute
	default:
		return State_WriteBack
	}
}

// generateExecuteSignals generates control signals for the execute state
func (cu *ControlUnit) generateExecuteSignals(opcode instructions.OpCode) {
	switch opcode {
	case instructions.OpCode_NOP:
		cu.pcInc.Set(High)

	case instructions.OpCode_MOV_IMM16H, instructions.OpCode_MOV_IMM16L:
		cu.aluSrcB.Set(High)                 // Use immediate
		cu.aluOp.SetValue(uint64(ALUOp_NOP)) // Pass through

	case instructions.OpCode_MOV:
		cu.aluOp.SetValue(uint64(ALUOp_NOP))

	case instructions.OpCode_ADD:
		cu.aluOp.SetValue(uint64(ALUOp_ADD))

	case instructions.OpCode_SUB:
		cu.aluOp.SetValue(uint64(ALUOp_SUB))

	case instructions.OpCode_MUL:
		cu.aluOp.SetValue(uint64(ALUOp_MUL))

	case instructions.OpCode_DIV:
		cu.aluOp.SetValue(uint64(ALUOp_DIV))

	case instructions.OpCode_MOD:
		cu.aluOp.SetValue(uint64(ALUOp_MOD))

	case instructions.OpCode_CMP:
		cu.aluOp.SetValue(uint64(ALUOp_CMP))

	case instructions.OpCode_LSL:
		cu.aluOp.SetValue(uint64(ALUOp_LSL))

	case instructions.OpCode_LSR:
		cu.aluOp.SetValue(uint64(ALUOp_LSR))

	case instructions.OpCode_ASL:
		cu.aluOp.SetValue(uint64(ALUOp_ASL))

	case instructions.OpCode_ASR:
		cu.aluOp.SetValue(uint64(ALUOp_ASR))

	case instructions.OpCode_JMP:
		cu.pcLoad.Set(High)
		cu.branch.Set(High)

	case instructions.OpCode_CJMP:
		// Check condition
		flags := uint32(cu.flags.GetValue())
		cond := uint8(cu.cond.GetValue())
		if cu.testCondition(flags, cond) {
			cu.pcLoad.Set(High)
			cu.branch.Set(High)
		} else {
			cu.pcInc.Set(High)
		}

	case instructions.OpCode_LD:
		// Address calculation (pass through register value)
		cu.aluOp.SetValue(uint64(ALUOp_NOP))

	case instructions.OpCode_ST:
		// Address calculation
		cu.aluOp.SetValue(uint64(ALUOp_NOP))
	}
}

// testCondition tests a condition code against flags (simplified)
func (cu *ControlUnit) testCondition(flags uint32, cond uint8) bool {
	z := (flags & FlagZero) != 0
	n := (flags & FlagNegative) != 0
	c := (flags & FlagCarry) != 0
	v := (flags & FlagOverflow) != 0

	switch cond {
	case 0: // EQ - Equal (Z=1)
		return z
	case 1: // NE - Not Equal (Z=0)
		return !z
	case 2: // CS/HS - Carry Set / Unsigned Higher or Same (C=1)
		return c
	case 3: // CC/LO - Carry Clear / Unsigned Lower (C=0)
		return !c
	case 4: // MI - Minus / Negative (N=1)
		return n
	case 5: // PL - Plus / Positive or Zero (N=0)
		return !n
	case 6: // VS - Overflow Set (V=1)
		return v
	case 7: // VC - Overflow Clear (V=0)
		return !v
	case 8: // HI - Unsigned Higher (C=1 and Z=0)
		return c && !z
	case 9: // LS - Unsigned Lower or Same (C=0 or Z=1)
		return !c || z
	case 10: // GE - Signed Greater or Equal (N=V)
		return n == v
	case 11: // LT - Signed Less Than (N!=V)
		return n != v
	case 12: // GT - Signed Greater Than (Z=0 and N=V)
		return !z && (n == v)
	case 13: // LE - Signed Less or Equal (Z=1 or N!=V)
		return z || (n != v)
	case 14: // AL - Always
		return true
	default:
		return false
	}
}

// Reset resets the control unit
func (cu *ControlUnit) Reset() {
	cu.currentState = State_Fetch
	cu.halted = false
	cu.clearSignals()
	cu.state.SetValue(uint64(State_Fetch))
}

// SetState sets the FSM state directly (for testing)
func (cu *ControlUnit) SetState(s CPUState) {
	cu.currentState = s
	cu.state.SetValue(uint64(s))
}
