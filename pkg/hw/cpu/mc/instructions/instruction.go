package instructions

import "strings"

// Stores a fully decoded instruction
type Instruction struct {
	Descriptor    *InstructionDescriptor
	OperandValues []OperandValue
}

func (i *Instruction) Raw() RawInstruction {
	raw := RawInstruction{
		Descriptor:    i.Descriptor,
		OperandValues: make([]uint64, len(i.OperandValues)),
	}

	for i, operandValue := range i.OperandValues {
		raw.OperandValues[i] = operandValue.Encode()
	}

	return raw
}

func (i *Instruction) String() string {
	var builder strings.Builder

	builder.WriteString(i.Descriptor.OpCode.Mnemonic)

	for j, operand := range i.OperandValues {
		builder.WriteString(operand.String())

		if j < len(i.OperandValues)-1 {
			builder.WriteString(" ")
		}
	}

	return builder.String()
}
