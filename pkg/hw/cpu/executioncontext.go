package cpu

import "github.com/Manu343726/cucaracha/pkg/hw/memory"

// ExecuteContext provides the minimal functionality required to execute instructions
type ExecuteContext interface {
	// Provides access to CPU registers
	Registers() Registers
	// Provides access to the RAM memory attached to the CPU
	Ram() memory.Memory
	// Halt stops CPU execution
	Halt()
	// EnableInterrupts enables interrupt handling
	EnableInterrupts()
	// DisableInterrupts disables interrupt handling
	DisableInterrupts()
	// SoftwareInterrupt triggers a software interrupt with the given vector
	SoftwareInterrupt(vector uint8) error
	// ReturnFromInterrupt returns from an interrupt handler
	ReturnFromInterrupt() error
}

// ExecuteFunc is the signature for instruction execution functions
// operands contains the decoded operand values (register indices or immediate values)
type ExecuteFunc func(ctx ExecuteContext, operands []uint32) error
