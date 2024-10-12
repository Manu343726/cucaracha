package registers

type RegisterClass uint

const (
	// State registers
	RegisterClass_StateRegisters RegisterClass = iota

	// General purpose integer integer registers
	RegisterClass_GeneralPurposeInteger

	// Number of register classes
	TOTAL_REGISTER_CLASSES
)

func (rc RegisterClass) String() string {
	switch rc {
	case RegisterClass_StateRegisters:
		return "state registers"
	case RegisterClass_GeneralPurposeInteger:
		return "general purpose integer registers"
	}

	panic("unreachable")
}
