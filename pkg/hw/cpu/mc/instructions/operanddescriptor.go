package instructions

import (
	"fmt"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/registers"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/types"
)

// Contains information about an instruction operand
type OperandDescriptor struct {
	// Type of operand
	Kind OperandKind
	// Role the operand takes in the instruction
	Role OperandRole
	// Register classes compatible with the operand in case the operand is a register, nil otherwise
	RegisterMetaClass *registers.RegisterMetaClass
	// Type of the value of the operand
	ValueType types.ValueType
	// First bit within the instruction used to encode this operand
	EncodingPosition int
	// Total bits used to encode the value into the instruction. The remaining significant bits left are truncated (ignored) from the value during encoding
	EncodingBits int
	// Operand description (for documentation and debugging)
	Description string
	// Position within the set of operands of the instruction, indexed from 0 to total operands - 1
	Index int

	// Custom operand name for LLVM instruction generation
	LLVM_CustomName string
	// Custom operand type for LLVM instruction generation
	LLVM_CustomType string
	// Custom operand pattern for LLVM instruction generation
	LLVM_CustomPattern string
	// If true, this operand is hidden from the assembly string (useful for tied operands)
	LLVM_HideFromAsm bool
}

// Returns true if the operand is a register operand
func (u *OperandDescriptor) IsRegister() bool {
	return u.Kind == OperandKind_Register
}

// Returns true if the operand is an immediate operand
func (u *OperandDescriptor) IsImmediate() bool {
	return u.Kind == OperandKind_Immediate
}

// Returns an human readable string describing the operand (See [InstructionDescriptor.PrettyPrint])
func (o *OperandDescriptor) String() string {
	if o.IsRegister() {
		return fmt.Sprintf("%v", o.RegisterMetaClass)
	} else {
		return fmt.Sprintf("<%v:%v:%v>", o.Role, o.Kind, o.ValueType)
	}
}

// Parses an operand value
func (o *OperandDescriptor) ParseValue(value string) (OperandValue, error) {
	switch o.Kind {
	case OperandKind_Immediate:
		return ParseImmediate(value, o.ValueType)
	case OperandKind_Register:
		register, err := registers.RegisterClasses.RegisterByName(value)

		if err != nil {
			return OperandValue{}, err
		}

		if err := o.RegisterMetaClass.RegisterBelongsToClass(register); err != nil {
			return OperandValue{}, err
		}

		return RegisterOperandValue(register), nil
	}

	panic("unreachable")
}

// Decodes an operand value
func (o *OperandDescriptor) DecodeValue(value uint64) (OperandValue, error) {
	switch o.Kind {
	case OperandKind_Immediate:
		return DecodeImmediate(value, o.ValueType)
	case OperandKind_Register:
		register, err := registers.RegisterClasses.DecodeRegister(value)

		if err != nil {
			return OperandValue{}, err
		}

		if err := o.RegisterMetaClass.RegisterBelongsToClass(register); err != nil {
			return OperandValue{}, err
		}

		return RegisterOperandValue(register), nil
	}

	panic("unreachable")
}

// Initializes a register operand descriptor
func RegisterOperandDescriptor(rmc *registers.RegisterMetaClass, opd OperandDescriptor) *OperandDescriptor {
	opd.RegisterMetaClass = rmc
	opd.ValueType = rmc.ValueType()
	opd.Kind = OperandKind_Register

	return &opd
}
