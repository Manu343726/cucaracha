package cpu

// Minimal interface for CPU interrupt handling
type Interrupts interface {
	// Enable interruptions
	Enable() error
	// Disable interruptions
	Disable() error
	// Check if interruptions are enabled
	Enabled() bool
	// Trigger a software interrupt with the given vector number
	Interrupt(vectorNumber uint8) error
	// Checks whether the CPU is currently servicing an interrupt
	Servicing() bool
	// Returns the vector number of the interrupt being serviced, or -1 if none
	CurrentInterrupt() int
	// Returns interrupts registered but now yet handled
	PendingInterrupts() []uint8
}
