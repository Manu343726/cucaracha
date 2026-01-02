package instructions

import (
	"strconv"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/registers"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/types"
	"github.com/Manu343726/cucaracha/pkg/utils"
)

// Stores the value of an instruction operand
type OperandValue struct {
	register  *registers.RegisterDescriptor
	immediate *types.Value
}

// Returns the kind of operand this value refers to
func (v *OperandValue) Kind() OperandKind {
	if v.register != nil {
		return OperandKind_Register
	} else if v.immediate != nil {
		return OperandKind_Immediate
	}

	panic("unreachable")
}

// Returns the value type of the operand value
func (v *OperandValue) ValueType() types.ValueType {
	if v.register != nil {
		return v.register.ValueType()
	} else if v.immediate != nil {
		return v.immediate.Type()
	}

	panic("unreachable")
}

// Returns the string representation of the operand value
func (v *OperandValue) String() string {
	if v.register != nil {
		return v.register.Name()
	} else if v.immediate != nil {
		return v.immediate.String()
	}

	panic("unreachable")
}

func (v *OperandValue) Immediate() types.Value {
	if v.immediate != nil {
		return *v.immediate
	}

	panic("operand value is not an immediate")
}

func (v *OperandValue) Register() *registers.RegisterDescriptor {
	if v.register != nil {
		return v.register
	}

	panic("operand value is not a register")
}

// Returns the binary representation of the operand value
func (v OperandValue) Encode() uint64 {
	if v.register != nil {
		return v.register.Encode()
	} else if v.immediate != nil {
		return v.immediate.Encode()
	}

	panic("unreachable")
}

// Returns a register operand value
func RegisterOperandValue(register *registers.RegisterDescriptor) OperandValue {
	return OperandValue{
		register:  register,
		immediate: nil,
	}
}

// Returns an immediate operand value
func ImmediateValue(value types.Value) OperandValue {
	return OperandValue{
		register:  nil,
		immediate: &value,
	}
}

// Parses an string as a 32 bit integer immediate operand value
func ParseInt32Immediate(value string) (OperandValue, error) {
	result, err := strconv.ParseInt(value, 0, 32)
	return ImmediateValue(types.Int32(int32(result))), err
}

// Parses an string as an immediate operand value
func ParseImmediate(value string, valueType types.ValueType) (OperandValue, error) {
	switch valueType {
	case types.ValueType_Int32:
		return ParseInt32Immediate(value)
	}
	return OperandValue{}, utils.MakeError(ErrInvalidInstruction, "unsupported operand value type %v for operand '%v'", valueType, value)
}

// Decodes a 32 bit integer immediate operand value
func DecodeInt32Immediate(binaryRepresentation uint64) OperandValue {
	return ImmediateValue(types.Int32(utils.BitCast[int32](binaryRepresentation)))
}

// Decodes an immediate operand value
func DecodeImmediate(binaryRepresentation uint64, valueType types.ValueType) (OperandValue, error) {
	switch valueType {
	case types.ValueType_Int32:
		return DecodeInt32Immediate(binaryRepresentation), nil
	}
	return OperandValue{}, utils.MakeError(ErrInvalidInstruction, "unsupported operand value type %v for immediate operand '%v'", valueType, utils.FormatUintHex(binaryRepresentation, 16))
}
