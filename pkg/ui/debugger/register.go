package debugger

// Register represents a CPU register including its current value and binary encoding.
type Register struct {
	// Name of the register (e.g., "r0", "sp", "ra").
	Name string `json:"name"`
	// Binary representation of the register's encoding/index in the CPU.
	Encoding uint32 `json:"encoding"`
	// Current numerical value stored in this register.
	Value uint32 `json:"value"`
}

// FlagState describes the CPU status flags that indicate the result of the last operation.
type FlagState struct {
	// Negative flag: set if the result of the last operation was negative.
	N bool `json:"n"`
	// Zero flag: set if the result of the last operation was zero.
	Z bool `json:"z"`
	// Carry flag: set if the last operation produced a carry/borrow.
	C bool `json:"c"`
	// Overflow flag: set if the last operation overflowed/underflowed.
	V bool `json:"v"`
}

// RegistersResult contains all CPU register values and status flags.
type RegistersResult struct {
	// Error message if register retrieval failed (nil if successful).
	Error error `json:"error"`
	// All CPU registers keyed by register name. See [Register] for structure.
	Registers map[string]*Register `json:"registers"`
	// CPU status flags. See [FlagState] for structure.
	Flags *FlagState `json:"flags"`
}
