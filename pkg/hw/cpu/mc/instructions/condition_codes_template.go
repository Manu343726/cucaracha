package instructions

// ConditionCodesTemplateData provides data for generating CucarachaConditionCodes.h
type ConditionCodesTemplateData struct {
	Flags     []FlagTemplateData
	CondCodes []CondCodeTemplateData
}

// FlagTemplateData represents a single CPSR flag
type FlagTemplateData struct {
	Name    string
	Value   uint32
	Comment string
}

// CondCodeTemplateData represents a named condition code
type CondCodeTemplateData struct {
	Name         string
	Code         uint32 // Numeric code value (0-14)
	Comment      string
	OppositeName string
	TestExpr     string // C++ expression to test the condition (uses z, n, c, v booleans)
}

// GetConditionCodesTemplateData returns the template data for generating the C++ header
func GetConditionCodesTemplateData() ConditionCodesTemplateData {
	return ConditionCodesTemplateData{
		Flags: []FlagTemplateData{
			{Name: "FLAG_Z", Value: uint32(FLAG_Z), Comment: "Zero flag - set when result is zero"},
			{Name: "FLAG_N", Value: uint32(FLAG_N), Comment: "Negative flag - set when result is negative"},
			{Name: "FLAG_C", Value: uint32(FLAG_C), Comment: "Carry flag - set on unsigned overflow/no borrow"},
			{Name: "FLAG_V", Value: uint32(FLAG_V), Comment: "Overflow flag - set on signed overflow"},
		},
		CondCodes: []CondCodeTemplateData{
			{Name: "EQ", Code: uint32(CC_EQ), Comment: "Equal", OppositeName: "NE", TestExpr: "z"},
			{Name: "NE", Code: uint32(CC_NE), Comment: "Not Equal", OppositeName: "EQ", TestExpr: "!z"},
			{Name: "CS", Code: uint32(CC_CS), Comment: "Carry Set (unsigned >=)", OppositeName: "CC", TestExpr: "c"},
			{Name: "CC", Code: uint32(CC_CC), Comment: "Carry Clear (unsigned <)", OppositeName: "CS", TestExpr: "!c"},
			{Name: "MI", Code: uint32(CC_MI), Comment: "Minus (negative)", OppositeName: "PL", TestExpr: "n"},
			{Name: "PL", Code: uint32(CC_PL), Comment: "Plus (positive or zero)", OppositeName: "MI", TestExpr: "!n"},
			{Name: "VS", Code: uint32(CC_VS), Comment: "Overflow Set", OppositeName: "VC", TestExpr: "v"},
			{Name: "VC", Code: uint32(CC_VC), Comment: "Overflow Clear", OppositeName: "VS", TestExpr: "!v"},
			{Name: "HI", Code: uint32(CC_HI), Comment: "Unsigned Higher", OppositeName: "LS", TestExpr: "c && !z"},
			{Name: "LS", Code: uint32(CC_LS), Comment: "Unsigned Lower or Same", OppositeName: "HI", TestExpr: "!c || z"},
			{Name: "GE", Code: uint32(CC_GE), Comment: "Signed Greater or Equal", OppositeName: "LT", TestExpr: "n == v"},
			{Name: "LT", Code: uint32(CC_LT), Comment: "Signed Less Than", OppositeName: "GE", TestExpr: "n != v"},
			{Name: "GT", Code: uint32(CC_GT), Comment: "Signed Greater", OppositeName: "LE", TestExpr: "!z && (n == v)"},
			{Name: "LE", Code: uint32(CC_LE), Comment: "Signed Less or Equal", OppositeName: "GT", TestExpr: "z || (n != v)"},
			{Name: "AL", Code: uint32(CC_AL), Comment: "Always", OppositeName: "AL", TestExpr: "true"},
		},
	}
}
