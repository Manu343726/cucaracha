package runtime

import (
	"github.com/Manu343726/cucaracha/pkg/hw/cpu"
	"github.com/Manu343726/cucaracha/pkg/hw/memory"
	"github.com/Manu343726/cucaracha/pkg/hw/peripheral"
)

type Runtime interface {
	// Provides access to the CPU
	CPU() cpu.CPU

	// Provides access to the RAM memory
	Memory() memory.Memory

	// Returns the memory layout
	MemoryLayout() memory.MemoryLayout

	// Provides access to the peripherals
	Peripherals() map[string]peripheral.Peripheral

	// Resets the runtime to its initial state
	Reset() error

	// Executes a single step of the runtime simulation
	Step() (*cpu.StepInfo, error)

	SetBreakpoint(addr uint32) error
	ClearBreakpoint(addr uint32) error

	SetWatchpoint(r memory.Range) error
	ClearWatchpoint(r memory.Range) error
}
