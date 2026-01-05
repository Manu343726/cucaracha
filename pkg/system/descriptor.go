package system

import (
	"github.com/Manu343726/cucaracha/pkg/hw/memory"
	"github.com/Manu343726/cucaracha/pkg/hw/peripheral"
)

// Describes the interrupt vector table characteristics
type VectorTableDescriptor struct {
	// Total number of interrupts supported
	NumberOfVectors uint32

	// Size of each interrupt vector entry in bytes
	// (e.g., 4 bytes for a 32-bit handler address, 8 bytes 32-bit handler address + 32-bit handler parameter data address, etc.)
	VectorEntrySize uint32
}

// Describes a cucaracha system and its hardware characteristics
type SystemDescriptor struct {
	// Describes the memory layout of the system
	MemoryLayout memory.MemoryLayout

	// Describes the interrupt vector table characteristics
	VectorTable VectorTableDescriptor

	// Describes the available peripherals in the system
	Peripherals []peripheral.Peripheral
}
