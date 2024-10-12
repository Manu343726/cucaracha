package instructions

import (
	"fmt"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/registers"
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

	// LLVM instruction selection pattern template
	LLVM_PatternTemplate string
	// Flags controlling high level semantics of the instruction in LLVM instruction definition. See
	// class Instruction definition bit flags in LLVM's source llvm/include/Target/Target.td
	LLVM_InstructionFlags LLVMInstructionFlags
	// Set of non operand registers that are implicitly modified by the instruction
	LLVM_Defs []*registers.RegisterDescriptor
	// Set of non operand registers that are implicitly read by the instruction
	LLVM_Uses []*registers.RegisterDescriptor
	// LLVM instruction definition metadata
	LLVM *LLVMInstructionDescriptor
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
			Name:  utils.FormatUintBinary(d.OpCode.BinaryRepresentation, Opcodes.OpCodeBits()),
			Begin: 0,
			Width: Opcodes.OpCodeBits(),
		},
	}
	fields = append(fields, utils.Map(d.Operands, func(op *OperandDescriptor) utils.AsciiFrameField {
		return utils.AsciiFrameField{
			Name:  op.String(),
			Begin: op.EncodingPosition,
			Width: op.EncodingBits,
		}
	})...)
	builder.WriteString(utils.AsciiFrame(fields, Instructions.InstructionBits(), "bits", utils.AsciiFrameUnitLayout_RightToLeft, leftpad+2))
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
