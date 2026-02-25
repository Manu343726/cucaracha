package instructions

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/registers"
	"github.com/Manu343726/cucaracha/pkg/utils/logging"
)

// Stores a fully decoded instruction
type Instruction struct {
	Descriptor    *InstructionDescriptor
	OperandValues []OperandValue
}

func (i *Instruction) Raw() *RawInstruction {
	raw := &RawInstruction{
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

// Returns a logging attribute for the instruction, which includes its assembly representation.
func (i *Instruction) LoggingAttribute(name string) slog.Attr {
	return logging.Instruction(name, i.String())
}

func NewInstruction(descriptor *InstructionDescriptor, operands []OperandValue) (*Instruction, error) {
	// Validate operand count
	realOperands := descriptor.RealOperands()
	expectedOperands := len(realOperands)
	if len(operands) != expectedOperands {
		return nil, fmt.Errorf("instruction %s expects %d operands, got %d",
			descriptor.OpCode.Mnemonic, expectedOperands, len(operands))
	}

	// Validate operand types
	for i, op := range operands {
		desc := realOperands[i]
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

// Returns the register containing the branch target for the given branch instruction
func BranchTargetRegister(instr *Instruction) (*registers.RegisterDescriptor, error) {
	switch instr.Descriptor.OpCode.OpCode {
	case OpCode_CJMP:
		return cjmpBranchTargetRegister(instr)
	case OpCode_JMP:
		return jmpBranchTargetRegister(instr)
	default:
		return nil, fmt.Errorf("instruction is not a branch")
	}
}

func cjmpBranchTargetRegister(instr *Instruction) (*registers.RegisterDescriptor, error) {
	if instr.Descriptor.OpCode.OpCode != OpCode_CJMP {
		panic("expected CJMP instruction for cjmpBranchTarget()")
	}

	// CJMP has three operands: condition register, jump target address register, and link register
	if len(instr.OperandValues) != 3 || len(instr.Descriptor.Operands) != 3 {
		panic("expected 3 operands for CJMP instruction")
	}

	if instr.Descriptor.Operands[1].Role != OperandRole_Source {
		panic("expected jump target operand to have source role")
	}

	if instr.OperandValues[1].Kind() != OperandKind_Register {
		panic("expected jump target operand to be a register")
	}

	return instr.OperandValues[1].Register(), nil
}

func jmpBranchTargetRegister(instr *Instruction) (*registers.RegisterDescriptor, error) {
	if instr.Descriptor.OpCode.OpCode != OpCode_JMP {
		panic("expected JMP instruction for jmpBranchTarget()")
	}

	// JMP has two operands: jump target address register and link register
	if len(instr.OperandValues) != 2 || len(instr.Descriptor.Operands) != 2 {
		panic("expected 2 operands for JMP instruction")
	}

	if instr.Descriptor.Operands[0].Role != OperandRole_Source {
		panic("expected jump target operand to have source role")
	}

	if instr.OperandValues[0].Kind() != OperandKind_Register {
		panic("expected jump target operand to be a register")
	}

	return instr.OperandValues[0].Register(), nil
}
