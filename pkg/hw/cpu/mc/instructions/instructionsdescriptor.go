package instructions

import (
	"errors"
	"fmt"

	"github.com/Manu343726/cucaracha/pkg/utils"
)

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

// Returns the number of bytes required to encode a machine instruction
func (d *InstructionsDescriptor) InstructionBytes() int {
	return d.InstructionBits() / 8
}

func fixInstructionOperands(instr *InstructionDescriptor) {
	currentEncodingPosition := Opcodes.OpCodeBits()

	for i := range instr.Operands {
		instr.Operands[i].Index = i
		if instr.Operands[i].EncodingPosition == 0 {
			instr.Operands[i].EncodingPosition = currentEncodingPosition
		} else if instr.Operands[i].EncodingPosition < currentEncodingPosition {
			panic(fmt.Errorf("operand %v of instruction %s has invalid encoding position %v (overlaps with previous operand or the instruction opcode)", instr.Operands[i], instr.OpCode.String(), instr.Operands[i].EncodingPosition))
		} else if instr.Operands[i].EncodingBits <= 0 {
			panic(fmt.Errorf("operand %v of instruction %s has invalid encoding bits %v", instr.Operands[i], instr.OpCode.String(), instr.Operands[i].EncodingBits))
		}

		currentEncodingPosition += instr.Operands[i].EncodingBits
	}
}

// Initializes an instructions descriptor with all the given instructions
func NewInstructionsDescriptor(instructions []*InstructionDescriptor) InstructionsDescriptor {
	// fill operand indices and LLVM metadata
	for _, instr := range instructions {
		fixInstructionOperands(instr)
		instr.LLVM = NewLLVMInstructionDescriptor(instr)
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

// Decode an instruction
func (d *InstructionsDescriptor) Decode(binaryRepresentation uint32) (*Instruction, error) {
	view := utils.CreateBitView(&binaryRepresentation)
	opCode, err := Opcodes.DecodeOpCode(uint64(view.Read(0, Opcodes.OpCodeBits())))

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

	raw := RawInstruction{
		Descriptor:    descriptor,
		OperandValues: operandValues,
	}

	return raw.Decode()
}
