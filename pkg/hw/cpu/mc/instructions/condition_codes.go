package instructions

// CPSRFlag represents individual flags in the Current Program Status Register.
// The CPSR contains condition flags that are set by comparison and arithmetic instructions.
type CPSRFlag uint32

const (
	// FLAG_Z is the Zero flag (bit 0) - set when result is zero
	FLAG_Z CPSRFlag = 1 << iota // 0x1
	// FLAG_N is the Negative flag (bit 1) - set when result is negative
	FLAG_N // 0x2
	// FLAG_C is the Carry flag (bit 2) - set on unsigned overflow/borrow
	FLAG_C // 0x4
	// FLAG_V is the Overflow flag (bit 3) - set on signed overflow
	FLAG_V // 0x8
)

// ConditionCode represents condition codes used by CJMP instruction.
// These are numeric codes (0-14) that are evaluated using proper condition
// semantics (e.g., GT = Z=0 AND N=V, not just a simple AND mask).
type ConditionCode uint32

const (
	CC_EQ ConditionCode = iota // 0 - Equal (Z=1)
	CC_NE                      // 1 - Not Equal (Z=0)
	CC_CS                      // 2 - Carry Set (C=1, unsigned >=)
	CC_CC                      // 3 - Carry Clear (C=0, unsigned <)
	CC_MI                      // 4 - Minus (N=1)
	CC_PL                      // 5 - Plus (N=0)
	CC_VS                      // 6 - Overflow Set (V=1)
	CC_VC                      // 7 - Overflow Clear (V=0)
	CC_HI                      // 8 - Unsigned Higher (C=1 AND Z=0)
	CC_LS                      // 9 - Unsigned Lower or Same (C=0 OR Z=1)
	CC_GE                      // 10 - Signed Greater or Equal (N=V)
	CC_LT                      // 11 - Signed Less Than (N!=V)
	CC_GT                      // 12 - Signed Greater (Z=0 AND N=V)
	CC_LE                      // 13 - Signed Less or Equal (Z=1 OR N!=V)
	CC_AL                      // 14 - Always
	CC_INVALID
)

// String returns the condition code name
func (cc ConditionCode) String() string {
	names := []string{
		"EQ", "NE", "CS", "CC", "MI", "PL", "VS", "VC",
		"HI", "LS", "GE", "LT", "GT", "LE", "AL", "INVALID",
	}
	if int(cc) < len(names) {
		return names[cc]
	}
	return "UNKNOWN"
}

// Opposite returns the opposite condition code
func (cc ConditionCode) Opposite() ConditionCode {
	opposites := []ConditionCode{
		CC_NE, CC_EQ, CC_CC, CC_CS, CC_PL, CC_MI, CC_VC, CC_VS,
		CC_LS, CC_HI, CC_LT, CC_GE, CC_LE, CC_GT, CC_AL, CC_INVALID,
	}
	if int(cc) < len(opposites) {
		return opposites[cc]
	}
	return CC_INVALID
}

// ComputeCPSR computes the CPSR flags for a comparison between two values
func ComputeCPSR(lhs, rhs uint32) uint32 {
	var flags uint32 = 0

	// Zero flag - set if lhs == rhs
	if lhs == rhs {
		flags |= uint32(FLAG_Z)
	}

	// For subtraction lhs - rhs:
	diff := lhs - rhs

	// Negative flag - set if result is negative (bit 31 set)
	if diff&0x80000000 != 0 {
		flags |= uint32(FLAG_N)
	}

	// Carry flag - set if there was NO borrow (unsigned: lhs >= rhs)
	if lhs >= rhs {
		flags |= uint32(FLAG_C)
	}

	// Overflow flag - set if signed overflow occurred
	// Overflow happens when:
	// - positive - negative = negative (lhs pos, rhs neg, diff neg)
	// - negative - positive = positive (lhs neg, rhs pos, diff pos)
	lhsSign := lhs & 0x80000000
	rhsSign := rhs & 0x80000000
	diffSign := diff & 0x80000000
	if (lhsSign != rhsSign) && (diffSign != lhsSign) {
		flags |= uint32(FLAG_V)
	}

	return flags
}

// TestCondition tests if a condition code is satisfied given CPSR flags
func TestCondition(cpsr uint32, cc ConditionCode) bool {
	z := (cpsr & uint32(FLAG_Z)) != 0
	n := (cpsr & uint32(FLAG_N)) != 0
	c := (cpsr & uint32(FLAG_C)) != 0
	v := (cpsr & uint32(FLAG_V)) != 0

	switch cc {
	case CC_EQ:
		return z // Equal: Z=1
	case CC_NE:
		return !z // Not Equal: Z=0
	case CC_CS:
		return c // Carry Set: C=1 (unsigned >=)
	case CC_CC:
		return !c // Carry Clear: C=0 (unsigned <)
	case CC_MI:
		return n // Minus: N=1
	case CC_PL:
		return !n // Plus: N=0
	case CC_VS:
		return v // Overflow Set: V=1
	case CC_VC:
		return !v // Overflow Clear: V=0
	case CC_HI:
		return c && !z // Unsigned Higher: C=1 and Z=0
	case CC_LS:
		return !c || z // Unsigned Lower or Same: C=0 or Z=1
	case CC_GE:
		return n == v // Signed Greater or Equal: N=V
	case CC_LT:
		return n != v // Signed Less Than: N!=V
	case CC_GT:
		return !z && (n == v) // Signed Greater: Z=0 and N=V
	case CC_LE:
		return z || (n != v) // Signed Less or Equal: Z=1 or N!=V
	case CC_AL:
		return true // Always
	default:
		return false
	}
}
