package instructions

import (
	"fmt"
	"strings"
)

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

	first := true
	for j, operand := range i.OperandValues {
		// Skip operands that are hidden from assembly (e.g., tied operands)
		if i.Descriptor.Operands[j].LLVM_HideFromAsm {
			continue
		}

		if first {
			builder.WriteString(" ")
			first = false
		} else {
			builder.WriteString(", ")
		}
		builder.WriteString(operand.String())
	}

	return builder.String()
}

func NewInstruction(descriptor *InstructionDescriptor, operands []OperandValue) (*Instruction, error) {
	// Validate operand count
	expectedOperands := len(descriptor.Operands)
	if len(operands) != expectedOperands {
		return nil, fmt.Errorf("instruction %s expects %d operands, got %d",
			descriptor.OpCode.Mnemonic, expectedOperands, len(operands))
	}

	// Validate operand types
	for i, op := range operands {
		desc := descriptor.Operands[i]
		if op.Kind() != desc.Kind {
			return nil, fmt.Errorf("operand %d of %s: expected %s, got %s",
				i, descriptor.OpCode.Mnemonic, desc.Kind, op.Kind())
		}

		if op.ValueType() != desc.ValueType {
			return nil, fmt.Errorf("operand %d of %s: expected operand of type %s, got %s",
				i, descriptor.OpCode.Mnemonic, desc.ValueType, op.ValueType())
		}
	}

	return &Instruction{
		Descriptor:    descriptor,
		OperandValues: operands,
	}, nil
}
