package instructions

// Represents the role an operand has within an instruction
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
