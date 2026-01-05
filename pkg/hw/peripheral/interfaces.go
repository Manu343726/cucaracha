// Package peripheral provides interfaces and utilities for memory-mapped peripherals
// that bridge the CPU simulation and user interface.
package peripheral

import "github.com/Manu343726/cucaracha/pkg/hw/memory"

type Metadata struct {
	// Name of the peripheral instance
	Name string
	// Purpose and usafe of the peripheral
	Description string
	// Memory-mapped I/O base address
	BaseAddress uint32
	// Memory-mapped I/O region size
	Size uint32
	// Interrupt vector number (if applicable)
	InterruptVector uint8
	// Peripheral type descriptor
	Descriptor *Descriptor
}

// Provides access to the environment needed to execute peripheral actions.
type Environment struct {
	MemoryLayout memory.MemoryLayout
	RAM          memory.Memory
}

// Peripheral represents a device that communicates with the CPU through
// memory-mapped I/O and exposes functionality to the simulator UI.
type Peripheral interface {
	// Information about the peripheral.
	Metadata() Metadata

	// Lifecycle
	Reset()

	// Clock is called each CPU cycle, allowing peripherals to update state
	Clock(env Environment) error
}

// UIPeripheral extends Peripheral with UI-facing capabilities.
// Peripherals that want to expose functionality to the simulator UI
// should implement this interface.
type UIPeripheral interface {
	Peripheral

	// UIState returns a map of named values that the UI can display.
	// Keys are field names, values can be any displayable type.
	UIState() map[string]interface{}

	// UIActions returns available actions the user can trigger.
	// Returns a map of action name -> action description.
	UIActions() map[string]string

	// UITrigger executes a named action with optional parameters.
	// Returns an error if the action is unknown or fails.
	UITrigger(action string, params map[string]interface{}) error
}

// InterruptSource is implemented by peripherals that can raise interrupts.
type InterruptSource interface {
	Peripheral

	// InterruptPending returns true if the peripheral has a pending interrupt.
	InterruptPending() bool

	// InterruptVector returns the interrupt vector number.
	InterruptVector() uint8

	// AcknowledgeInterrupt clears the pending interrupt.
	AcknowledgeInterrupt()
}
