package instructions

import (
	"fmt"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/utils"
)

// Stores a partially decoded instruction
//
// Raw instructions are generated as a middle step in the instruction decoding process, when
// the instruction opcode has been decoded and identified (So we already have access to the instruction descriptor)
// but the instruction operands values have not been decoded yet
type RawInstruction struct {
	Descriptor    *InstructionDescriptor
	OperandValues []uint64
}

// Generates am ASCII frame representation of the instruction, showing all opcode and operand bits
func (instr RawInstruction) PrettyPrint(leftpad int) string {
	fields := []utils.AsciiFrameField{
		{
			Name:  utils.FormatUintBinary(instr.Descriptor.OpCode.BinaryRepresentation, Opcodes.OpCodeBits()),
			Begin: 0,
			Width: Opcodes.OpCodeBits(),
		},
	}
	fields = append(fields, utils.Map(instr.Descriptor.Operands, func(op *OperandDescriptor) utils.AsciiFrameField {
		return utils.AsciiFrameField{
			Name:  fmt.Sprintf("[%v] %v (%v)", op, utils.FormatUintBinary(instr.OperandValues[op.Index], op.EncodingBits), utils.FormatUintHex(instr.OperandValues[op.Index], op.EncodingBits/4)),
			Begin: op.EncodingPosition,
			Width: op.EncodingBits,
		}
	})...)

	return utils.AsciiFrame(fields, Instructions.InstructionBits(), "bits", utils.AsciiFrameUnitLayout_RightToLeft, leftpad)
}

func (instr *RawInstruction) String() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("%v ", instr.Descriptor.OpCode.Mnemonic))

	for i, operand := range instr.Descriptor.Operands {
		builder.WriteString(utils.FormatUintHex(instr.OperandValues[i], operand.EncodingBits/4))

		if i < len(instr.OperandValues)-1 {
			builder.WriteString(", ")
		}
	}

	return builder.String()
}

// Returns the instruction with all its operands decoded
func (instr *RawInstruction) Decode() (*Instruction, error) {
	i := Instruction{
		Descriptor:    instr.Descriptor,
		OperandValues: make([]OperandValue, len(instr.OperandValues)),
	}

	for j, operandDescriptor := range instr.Descriptor.Operands {
		value, err := operandDescriptor.DecodeValue(instr.OperandValues[j])

		if err != nil {
			return nil, utils.MakeError(ErrInvalidInstruction, "error decoding operand [%v]: %w", j, err)
		}

		i.OperandValues[j] = value
	}

	return &i, nil
}

// Returns the binary representation of the instruction, with the opcode and all operands encoded
func (instr *RawInstruction) Encode() uint32 {
	if len(instr.OperandValues) != len(instr.Descriptor.Operands) {
		panic(fmt.Errorf("mistmatched operand values, the instruction must have %v operands, we have %v values", len(instr.Descriptor.Operands), len(instr.OperandValues)))
	}

	var binaryRepresentation uint32 = 0
	view := utils.CreateBitView(&binaryRepresentation)

	view.Write(uint32(instr.Descriptor.OpCode.BinaryRepresentation), 0, Opcodes.OpCodeBits())

	for i, operand := range instr.Descriptor.Operands {
		view.Write(uint32(instr.OperandValues[i]), operand.EncodingPosition, operand.EncodingBits)
	}

	return binaryRepresentation
}
