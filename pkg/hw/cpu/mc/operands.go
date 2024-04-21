package mc

import "fmt"

type OperandKind uint

const (
	OperandKind_Immediate OperandKind = iota
	OperandKind_Register
)

func (o OperandKind) String() string {
	switch o {
	case OperandKind_Immediate:
		return "Immediate"
	case OperandKind_Register:
		return "Register"
	}

	panic("unreachable")
}

type OperandRole uint

const (
	OperandRole_Source OperandRole = iota
	OperandRole_Destination
)

func (o OperandRole) String() string {
	switch o {
	case OperandRole_Source:
		return "Source"
	case OperandRole_Destination:
		return "Destination"
	}

	panic("unreachable")
}

type OperandDescriptor struct {
	// Type of operand
	Kind OperandKind
	// Role the operand takes in the instruction
	Role OperandRole
	// Type of the value of the operand
	ValueType ValueType
	// First bit within the instruction used to encode this operand
	EncodingPosition int
	// Total bits used to encode the value into the instruction. The remaining significant bits left are truncated (ignored) from the value during encoding
	EncodingBits int
	// Operand description (for documentation and debugging)
	Description string
}

// Returns an human readable string describing the operand (See [InstructionDescriptor.PrettyPrint])
func (o *OperandDescriptor) String() string {
	return fmt.Sprintf("%v <%v:%v>", o.Role, o.Kind, o.ValueType)
}

// Returns the operand value ready to be encoded into the instruction with all unused most significant bits cleared.
// We prefer doing this and getting a weird operand value in case of encoding overflow than having the instruction completely
// trashed because we touched the wrong bits while encoding the operand
func (o *OperandDescriptor) EncodeValue(value uint64) uint64 {
	if o.EncodingBits > 64 {
		panic("cannot encode operand values that require more than 64 bits")
	}

	return value & ((uint64(o.EncodingBits) & 63) - 1)
}
