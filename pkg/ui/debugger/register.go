package debugger

// Contains information about a register
type Register struct {
	Name     string `json:"name"`     // Register name
	Encoding uint32 `json:"encoding"` // Binary representation of the register
	Value    uint32 `json:"value"`    // Current value of the register
}

// FlagState contains the CPU flags
type FlagState struct {
	N bool `json:"n"` // Negative
	Z bool `json:"z"` // Zero
	C bool `json:"c"` // Carry
	V bool `json:"v"` // Overflow
}

// Result of Registers command
type RegistersResult struct {
	Error     error                `json:"error"`     // Error, if any
	Registers map[string]*Register `json:"registers"` // Register values keyed by name
	Flags     *FlagState           `json:"flags"`     // CPU flags
}
