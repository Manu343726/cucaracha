package instructions

import (
	"fmt"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/registers"
	"github.com/Manu343726/cucaracha/pkg/utils"
)

// ExecuteContext provides the CPU state needed for instruction execution
type ExecuteContext interface {
	// GetRegister returns the value of a register by index
	GetRegister(idx uint32) uint32
	// SetRegister sets the value of a register by index
	SetRegister(idx uint32, value uint32)
	// GetPC returns the current program counter
	GetPC() uint32
	// SetPC sets the program counter
	SetPC(pc uint32)
	// ReadMemory32 reads a 32-bit word from memory
	ReadMemory32(addr uint32) (uint32, error)
	// WriteMemory32 writes a 32-bit word to memory
	WriteMemory32(addr uint32, value uint32) error
}

// ExecuteFunc is the signature for instruction execution functions
// operands contains the decoded operand values (register indices or immediate values)
type ExecuteFunc func(ctx ExecuteContext, operands []uint32) error

// Contains information describing an instruction
type InstructionDescriptor struct {
	// Instruction opcode
	OpCode *OpCodeDescriptor
	// Instruction operands
	Operands []*OperandDescriptor
	// Instruction description (for documentation and debugging)
	Description string

	// Execute is the function that implements the instruction behavior
	Execute ExecuteFunc

	// LLVM instruction selection pattern template
	LLVM_PatternTemplate string
	// Flags controlling high level semantics of the instruction in LLVM instruction definition. See
	// class Instruction definition bit flags in LLVM's source llvm/include/Target/Target.td
	LLVM_InstructionFlags LLVMInstructionFlags
	// Set of non operand registers that are implicitly modified by the instruction
	LLVM_Defs []*registers.RegisterDescriptor
	// Set of non operand registers that are implicitly read by the instruction
	LLVM_Uses []*registers.RegisterDescriptor
	// LLVM operand constraints (e.g. "$dst = $src" for tied operands)
	LLVM_Constraints string
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

	asciiFrame, err := utils.AsciiFrame(fields, d.InstructionBits(), "bits", utils.AsciiFrameUnitLayout_RightToLeft, leftpad+2)
	if err != nil {
		panic(fmt.Errorf("error generating documentation for instruction %s: %w", d.OpCode.String(), err))
	}

	builder.WriteString(asciiFrame)
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
