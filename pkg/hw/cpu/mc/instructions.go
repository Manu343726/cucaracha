package mc

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/utils"
)

// Contains information describing an instruction
type InstructionDescriptor struct {
	// Instruction opcode
	OpCode *OpCodeDescriptor
	// Instruction operands
	Operands []*OperandDescriptor
	// Instruction description (for documentation and debugging)
	Description string
}

// Returns a human readable string representation of the instruction
func (d *InstructionDescriptor) String() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("%v ", d.OpCode))

	for i := range d.Operands {
		operand := d.Operands[i]

		builder.WriteString(operand.String())

		if i < len(d.Operands)-1 {
			builder.WriteString(" ")
		}
	}

	return builder.String()
}

// Returns full documentation for the instruction
func (d *InstructionDescriptor) Documentation(leftpad int) string {
	var builder strings.Builder
	leftpad_str := strings.Repeat(" ", leftpad)

	builder.WriteString(leftpad_str)
	builder.WriteString(fmt.Sprintf("%v\n\n", d))

	leftpad_str += "  "
	leftpad += 2

	builder.WriteString(leftpad_str)
	builder.WriteString("Description:\n\n  ")
	builder.WriteString(leftpad_str)
	builder.WriteString(d.Description)
	builder.WriteString("\n\n")
	builder.WriteString(leftpad_str)
	builder.WriteString("Memory layout:\n\n")
	fields := []utils.AsciiFrameField{
		{
			Name:  utils.FormatUintBinary(d.OpCode.BinaryRepresentation, Descriptor_Opcodes.OpCodeBits()),
			Begin: 0,
			Width: Descriptor_Opcodes.OpCodeBits(),
		},
	}
	fields = append(fields, utils.Map(d.Operands, func(op *OperandDescriptor) utils.AsciiFrameField {
		return utils.AsciiFrameField{
			Name:  op.String(),
			Begin: op.EncodingPosition,
			Width: op.EncodingBits,
		}
	})...)
	builder.WriteString(utils.AsciiFrame(fields, Descriptor_Instructions.InstructionBits(), "bits", utils.AsciiFrameUnitLayout_RightToLeft, leftpad+2))
	builder.WriteString("\n")
	builder.WriteString(leftpad_str)
	builder.WriteString("Operands:\n\n")

	if len(d.Operands) > 0 {
		for i, operand := range d.Operands {
			builder.WriteString(leftpad_str)
			builder.WriteString(fmt.Sprintf(" [%v] %v: %v\n", i, operand, operand.Description))
		}
	} else {
		builder.WriteString(leftpad_str)
		builder.WriteString("  (none)\n")
	}

	return builder.String()
}

// Returns the minimum bits required to encode the instruction
func (d *InstructionDescriptor) InstructionBits() int {
	return utils.Reduce(d.Operands, func(op *OperandDescriptor, totalBits int) int {
		return op.EncodingBits + totalBits
	})
}

// Constains information about all implemented instructions
type InstructionsDescriptor struct {
	instructions map[OpCode]*InstructionDescriptor
}

// Returns all implemented instructions
func (d *InstructionsDescriptor) AllInstructions() []*InstructionDescriptor {
	return utils.Values(d.instructions)
}

var ErrInstructionNotImplemented = errors.New("instruction not implemented")

// Returns the instruction corresponding to the given opcode
func (d *InstructionsDescriptor) Instruction(op OpCode) (*InstructionDescriptor, error) {
	if instruction, hasInstruction := d.instructions[op]; hasInstruction {
		return instruction, nil
	} else {
		return nil, utils.MakeError(ErrInstructionNotImplemented, "no instruction implemented for opcode '%v'", op)
	}
}

// Returns the number of bits required to encode a machine instruction
func (d *InstructionsDescriptor) InstructionBits() int {
	return 32
}

func makeInstructionsDescriptor(instructions []*InstructionDescriptor) InstructionsDescriptor {
	d := InstructionsDescriptor{
		instructions: utils.GenMap(instructions, func(i *InstructionDescriptor) OpCode { return i.OpCode.OpCode }),
	}

	for _, instruction := range d.AllInstructions() {
		if instruction.InstructionBits() > d.InstructionBits() {
			panic(fmt.Errorf("instruction '%v' requires %v bits for encoding, instructions should fit in %v bits", instruction.String(), instruction.InstructionBits(), d.InstructionBits()))
		}
	}

	return d
}

var Descriptor_Instructions InstructionsDescriptor = makeInstructionsDescriptor([]*InstructionDescriptor{
	{
		OpCode:      Descriptor_Opcodes.Descriptor(OpCode_NOP),
		Description: "No operation. All internal state except from the program counter stays the same afer the execution of the instruction. Takes only 1 CPU cicle",
		Operands:    nil,
	},
	{
		OpCode:      Descriptor_Opcodes.Descriptor(OpCode_ADD),
		Description: "Add the values of two word registers and save the result into a word register",
		Operands: []*OperandDescriptor{
			{
				Kind:             OperandKind_Register,
				ValueType:        ValueType_Int32,
				EncodingPosition: 4,
				EncodingBits:     8,
				Role:             OperandRole_Source,
				Description:      "first source register",
			},
			{
				Kind:             OperandKind_Register,
				ValueType:        ValueType_Int32,
				EncodingPosition: 12,
				EncodingBits:     8,
				Role:             OperandRole_Source,
				Description:      "second source register",
			},
			{
				Kind:             OperandKind_Register,
				ValueType:        ValueType_Int32,
				EncodingPosition: 20,
				EncodingBits:     8,
				Role:             OperandRole_Destination,
				Description:      "destination register",
			},
		},
	},
})
