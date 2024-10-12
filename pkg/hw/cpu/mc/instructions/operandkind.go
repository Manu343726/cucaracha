package instructions

// Represents the kind of operand (Register, immediate, etc)
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
