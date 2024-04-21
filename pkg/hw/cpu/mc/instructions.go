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
	// fill operand indices
	for _, instr := range instructions {
		for i := range instr.Operands {
			instr.Operands[i].Index = i
		}
	}

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

type Instruction struct {
	Descriptor    *InstructionDescriptor
	OperandValues []uint64
}

// Generates am ASCII frame representation of the instruction, showing all opcode and operand bits
func (instr *Instruction) PrettyPrint(leftpad int) string {
	fields := []utils.AsciiFrameField{
		{
			Name:  utils.FormatUintBinary(instr.Descriptor.OpCode.BinaryRepresentation, Descriptor_Opcodes.OpCodeBits()),
			Begin: 0,
			Width: Descriptor_Opcodes.OpCodeBits(),
		},
	}
	fields = append(fields, utils.Map(instr.Descriptor.Operands, func(op *OperandDescriptor) utils.AsciiFrameField {
		return utils.AsciiFrameField{
			Name:  fmt.Sprintf("[%v] %v (%v)", op, utils.FormatUintBinary(instr.OperandValues[op.Index], op.EncodingBits), utils.FormatUintHex(instr.OperandValues[op.Index], op.EncodingBits/4)),
			Begin: op.EncodingPosition,
			Width: op.EncodingBits,
		}
	})...)

	return utils.AsciiFrame(fields, Descriptor_Instructions.InstructionBits(), "bits", utils.AsciiFrameUnitLayout_RightToLeft, leftpad)
}

func (instr *Instruction) String() string {
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

// Returns the binary representation of the instruction, with the opcode and all operands encoded
func (instr *Instruction) Encode() uint32 {
	if len(instr.OperandValues) != len(instr.Descriptor.Operands) {
		panic(fmt.Errorf("mistmatched operand values, the instruction must have %v operands, we have %v values", len(instr.Descriptor.Operands), len(instr.OperandValues)))
	}

	var binaryRepresentation uint32 = 0
	view := utils.CreateBitView(&binaryRepresentation)

	view.Write(uint32(instr.Descriptor.OpCode.BinaryRepresentation), 0, Descriptor_Opcodes.OpCodeBits())

	for i, operand := range instr.Descriptor.Operands {
		view.Write(uint32(instr.OperandValues[i]), operand.EncodingPosition, operand.EncodingBits)
	}

	return binaryRepresentation
}

// Decode an instruction
func (d *InstructionsDescriptor) Decode(binaryRepresentation uint32) (*Instruction, error) {
	view := utils.CreateBitView(&binaryRepresentation)
	opCode, err := Descriptor_Opcodes.DecodeOpCode(uint64(view.Read(0, Descriptor_Opcodes.OpCodeBits())))

	if err != nil {
		return nil, err
	}

	descriptor, err := d.Instruction(opCode)

	if err != nil {
		return nil, err
	}

	operandValues := make([]uint64, len(descriptor.Operands))

	for i, operand := range descriptor.Operands {
		operandValues[i] = uint64(view.Read(operand.EncodingPosition, operand.EncodingBits))
	}

	return &Instruction{
		Descriptor:    descriptor,
		OperandValues: operandValues,
	}, nil
}

// Decode an instruction
func DecodeInstruction(binaryRepresentation uint32) (*Instruction, error) {
	return Descriptor_Instructions.Decode(binaryRepresentation)
}
