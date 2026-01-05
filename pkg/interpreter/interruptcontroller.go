package interpreter

import (
	"sync"
)

// InterruptController manages interrupt sources and prioritizes them.
// It provides a centralized way to handle hardware interrupts from peripherals.
type InterruptController struct {
	mu sync.Mutex

	// System peripherals (Peripherals not implementing InterruptSource are ignored)
	peripherals *Peripherals

	// Global interrupt enable flag
	enabled bool

	// Interrupt mask - bit N controls whether vector N is enabled
	mask uint32

	// Pending interrupts bitmap
	pending uint32

	// Currently servicing interrupt (-1 if none)
	servicing int

	// Interrupt vector table base address
	vectorTableBase uint32

	// Size of each vector entry in bytes
	vectorEntrySize uint32

	// Number of supported interrupt vectors
	numVectors int
}

// Creates a new interrupt controller.
// vectorTableBase is the memory address where interrupt vectors are stored.
// numVectors is the number of interrupt vectors supported (typically 32 or 256).
func NewInterruptController(vectorTableBase uint32, vectorEntrySize uint32, numVectors int, peripherals *Peripherals) *InterruptController {
	return &InterruptController{
		peripherals:     peripherals,
		enabled:         false,
		mask:            0xFFFFFFFF, // All interrupts enabled by default
		pending:         0,
		servicing:       -1,
		vectorTableBase: vectorTableBase,
		vectorEntrySize: vectorEntrySize,
		numVectors:      numVectors,
	}
}

// Enable globally enables interrupt handling.
func (ic *InterruptController) Enable() error {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	ic.enabled = true
	return nil
}

// Disable globally disables interrupt handling.
func (ic *InterruptController) Disable() error {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	ic.enabled = false
	return nil
}

// Returns true if interrupts are globally enabled.
func (ic *InterruptController) Enabled() bool {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	return ic.enabled
}

// Sets the interrupt mask. Bit N controls vector N.
// A bit value of 1 means the interrupt is enabled.
func (ic *InterruptController) SetMask(mask uint32) {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	ic.mask = mask
}

// Returns the current interrupt mask.
func (ic *InterruptController) GetMask() uint32 {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	return ic.mask
}

// EnableVector enables a specific interrupt vector.
func (ic *InterruptController) EnableVector(vector uint8) {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	if int(vector) < ic.numVectors {
		ic.mask |= 1 << vector
	}
}

// DisableVector disables a specific interrupt vector.
func (ic *InterruptController) DisableVector(vector uint8) {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	if int(vector) < ic.numVectors {
		ic.mask &^= 1 << vector
	}
}

// Returns true if a specific vector is enabled.
func (ic *InterruptController) IsVectorEnabled(vector uint8) bool {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	if int(vector) >= ic.numVectors {
		return false
	}
	return ic.mask&(1<<vector) != 0
}

// Returns the pending interrupt bitmap.
func (ic *InterruptController) GetPending() uint32 {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	return ic.pending
}

// SetPending manually sets a pending interrupt (for software interrupts).
func (ic *InterruptController) SetPending(vector uint8) {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	if int(vector) < ic.numVectors {
		ic.pending |= 1 << vector
	}
}

// Equivalent to SetPending(), for interface compliance.
func (ic *InterruptController) Interrupt(vectorNumber uint8) error {
	ic.SetPending(vectorNumber)
	return nil
}

// ClearPending clears a pending interrupt.
func (ic *InterruptController) ClearPending(vector uint8) {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	if int(vector) < ic.numVectors {
		ic.pending &^= 1 << vector
	}
}

// Poll checks all interrupt sources for pending interrupts.
// This should be called each CPU cycle.
func (ic *InterruptController) Poll() {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	for _, source := range ic.peripherals.GetInterruptSources() {
		if source.InterruptPending() {
			vector := source.InterruptVector()
			if int(vector) < ic.numVectors {
				ic.pending |= 1 << vector
			}
		}
	}
}

// Returns true if there's an interrupt that should be serviced.
// Takes into account global enable, mask, and current servicing state.
func (ic *InterruptController) HasPendingInterrupt() bool {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	if !ic.enabled {
		return false
	}

	// Don't interrupt if already servicing (simple non-nested model)
	if ic.servicing >= 0 {
		return false
	}

	// Check if any enabled interrupt is pending
	effective := ic.pending & ic.mask
	return effective != 0
}

// Returns the highest priority pending interrupt vector.
// Returns -1 if no interrupt is pending.
// Lower vector numbers have higher priority.
func (ic *InterruptController) GetNextInterrupt() int {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	if !ic.enabled || ic.servicing >= 0 {
		return -1
	}

	effective := ic.pending & ic.mask
	if effective == 0 {
		return -1
	}

	// Find lowest set bit (highest priority)
	for i := 0; i < ic.numVectors; i++ {
		if effective&(1<<i) != 0 {
			return i
		}
	}

	return -1
}

// BeginService marks an interrupt as being serviced and clears its pending bit.
// Returns the vector table entry address for the interrupt handler.
func (ic *InterruptController) BeginService(vector int) uint32 {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	if vector < 0 || vector >= ic.numVectors {
		return 0
	}

	ic.servicing = vector
	ic.pending &^= 1 << vector

	// Acknowledge the interrupt on all sources with this vector
	for _, source := range ic.peripherals.GetInterruptSources() {
		if int(source.InterruptVector()) == vector && source.InterruptPending() {
			source.AcknowledgeInterrupt()
		}
	}

	// Return address of vector table entry (each entry is 4 bytes)
	return ic.vectorTableBase + uint32(vector)*4
}

// EndService marks the current interrupt as complete.
// Call this when returning from an interrupt handler.
func (ic *InterruptController) EndService() {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	ic.servicing = -1
}

// Returns the currently servicing interrupt vector, or -1 if none.
func (ic *InterruptController) CurrentInterrupt() int {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	return ic.servicing
}

// Returns true if the CPU is currently servicing an interrupt.
func (ic *InterruptController) Servicing() bool {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	return ic.servicing >= 0
}

// Returns the base address of the interrupt vector table.
func (ic *InterruptController) VectorTableBase() uint32 {
	return ic.vectorTableBase
}

// Returns the number of supported interrupt vectors.
func (ic *InterruptController) NumVectors() int {
	return ic.numVectors
}

// Reset resets the interrupt controller to its initial state.
func (ic *InterruptController) Reset() {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	ic.enabled = false
	ic.mask = 0xFFFFFFFF
	ic.pending = 0
	ic.servicing = -1
}

// Returns interrupts queued but not yet handled.
func (ic *InterruptController) PendingInterrupts() []uint8 {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	var pending []uint8
	for i := 0; i < ic.numVectors; i++ {
		if ic.pending&(1<<i) != 0 {
			pending = append(pending, uint8(i))
		}
	}
	return pending
}
