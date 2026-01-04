package components

import (
	"github.com/Manu343726/cucaracha/pkg/hw/component"
)

func init() {
	registerDecoderComponents()
}

func registerDecoderComponents() {
	// Instruction Decoder
	Registry.Register(component.NewDescriptor("DECODER").
		DisplayName("Instruction Decoder").
		Description("Decodes 32-bit instruction into opcode and operands").
		Category(CategoryControl).
		Version("1.0.0").
		Input("INSTR", 32, "32-bit instruction word").
		Output("OPCODE", 5, "5-bit opcode").
		Output("OP1", 8, "First operand (8-bit register index or immediate low)").
		Output("OP2", 8, "Second operand (8-bit register index)").
		Output("OP3", 8, "Third operand (8-bit register index)").
		Output("IMM16", 16, "16-bit immediate value").
		Factory(func(name string, params map[string]interface{}) (component.Component, error) {
			return NewInstructionDecoder(name), nil
		}).
		Build())
}

// =============================================================================
// Instruction Decoder
// =============================================================================

// Cucaracha instruction format (32 bits):
// Bits 0-4:   Opcode (5 bits)
// Bits 5-12:  Operand 1 (8 bits) - register or immediate low
// Bits 13-20: Operand 2 (8 bits) - register
// Bits 21-28: Operand 3 (8 bits) - register
// For MOVIMM16L/H: bits 5-20 are 16-bit immediate

const (
	OpcodeMask = 0x1F        // 5 bits
	Op1Mask    = 0xFF << 5   // 8 bits at position 5
	Op2Mask    = 0xFF << 13  // 8 bits at position 13
	Op3Mask    = 0xFF << 21  // 8 bits at position 21
	Imm16Mask  = 0xFFFF << 5 // 16 bits at position 5
)

// InstructionDecoder decodes a 32-bit instruction word
type InstructionDecoder struct {
	*component.BaseComponent

	// Input
	instruction *component.StandardPort

	// Outputs
	opcode *component.StandardPort // 5-bit opcode
	op1    *component.StandardPort // 8-bit operand 1
	op2    *component.StandardPort // 8-bit operand 2
	op3    *component.StandardPort // 8-bit operand 3
	imm16  *component.StandardPort // 16-bit immediate
}

// NewInstructionDecoder creates a new instruction decoder
func NewInstructionDecoder(name string) *InstructionDecoder {
	dec := &InstructionDecoder{
		BaseComponent: component.NewBaseComponent(name, "DECODER"),
	}

	dec.instruction = component.NewInputPort("INSTR", 32)
	dec.AddInput(dec.instruction)

	dec.opcode = component.NewOutputPort("OPCODE", 5)
	dec.AddOutput(dec.opcode)

	dec.op1 = component.NewOutputPort("OP1", 8)
	dec.AddOutput(dec.op1)

	dec.op2 = component.NewOutputPort("OP2", 8)
	dec.AddOutput(dec.op2)

	dec.op3 = component.NewOutputPort("OP3", 8)
	dec.AddOutput(dec.op3)

	dec.imm16 = component.NewOutputPort("IMM16", 16)
	dec.AddOutput(dec.imm16)

	return dec
}

// Instruction returns the instruction input port
func (d *InstructionDecoder) Instruction() *component.StandardPort { return d.instruction }

// Opcode returns the opcode output port
func (d *InstructionDecoder) Opcode() *component.StandardPort { return d.opcode }

// Op1 returns the first operand output port
func (d *InstructionDecoder) Op1() *component.StandardPort { return d.op1 }

// Op2 returns the second operand output port
func (d *InstructionDecoder) Op2() *component.StandardPort { return d.op2 }

// Op3 returns the third operand output port
func (d *InstructionDecoder) Op3() *component.StandardPort { return d.op3 }

// Imm16 returns the 16-bit immediate output port
func (d *InstructionDecoder) Imm16() *component.StandardPort { return d.imm16 }

// GetOpcode returns the decoded opcode value
func (d *InstructionDecoder) GetOpcode() uint8 {
	return uint8(d.opcode.GetValue())
}

// GetOp1 returns the first operand value
func (d *InstructionDecoder) GetOp1() uint8 {
	return uint8(d.op1.GetValue())
}

// GetOp2 returns the second operand value
func (d *InstructionDecoder) GetOp2() uint8 {
	return uint8(d.op2.GetValue())
}

// GetOp3 returns the third operand value
func (d *InstructionDecoder) GetOp3() uint8 {
	return uint8(d.op3.GetValue())
}

// GetImm16 returns the 16-bit immediate value
func (d *InstructionDecoder) GetImm16() uint16 {
	return uint16(d.imm16.GetValue())
}

// Compute decodes the instruction (combinational logic)
func (d *InstructionDecoder) Compute() error {
	if !d.IsEnabled() {
		return nil
	}

	instr := uint32(d.instruction.GetValue())

	// Extract fields
	opcode := instr & OpcodeMask
	op1 := (instr >> 5) & 0xFF
	op2 := (instr >> 13) & 0xFF
	op3 := (instr >> 21) & 0xFF
	imm16 := (instr >> 5) & 0xFFFF

	d.opcode.SetValue(uint64(opcode))
	d.op1.SetValue(uint64(op1))
	d.op2.SetValue(uint64(op2))
	d.op3.SetValue(uint64(op3))
	d.imm16.SetValue(uint64(imm16))

	return nil
}

// Decode is a convenience method to decode an instruction directly
func (d *InstructionDecoder) Decode(instr uint32) {
	d.instruction.SetValue(uint64(instr))
	d.Compute()
}

// Reset resets all outputs
func (d *InstructionDecoder) Reset() {
	d.opcode.Reset()
	d.op1.Reset()
	d.op2.Reset()
	d.op3.Reset()
	d.imm16.Reset()
}
