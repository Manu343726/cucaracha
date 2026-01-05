package interpreter

import (
	"fmt"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu"
	"github.com/Manu343726/cucaracha/pkg/hw/memory"
	"github.com/Manu343726/cucaracha/pkg/system"
)

// Flag bit positions
const (
	FlagZ uint32 = 1 << 0 // Zero flag
	FlagN uint32 = 1 << 1 // Negative flag
	FlagC uint32 = 1 << 2 // Carry flag
	FlagV uint32 = 1 << 3 // Overflow flag
	FlagI uint32 = 1 << 4 // Interrupt enable flag
)

// InterruptState holds the CPU state saved during interrupt handling.
type InterruptState struct {
	// Saved program counter (return address)
	PC uint32
	// Saved flags register value
	Flags uint32
	// Whether interrupts were enabled
	InterruptsEnabled bool
}

// Complete state needed to simulate CPU execution.
type CPUState struct {
	// CPU Registers
	Registers cpu.Registers
	// Halted flag
	Halted bool

	// Full system RAM
	Ram memory.Memory
	// Memory layout configuration
	MemoryLayout memory.MemoryLayout

	// Peripheral collection
	Peripherals *Peripherals
	// Interrupt controller
	IntController *InterruptController
	// Saved interrupt states (for nested interrupts, if supported)
	SavedStates []InterruptState
	// Maximum nesting level (0 = no nesting allowed)
	MaxNesting int
	// Flags register (bits: Z=0, N=1, C=2, V=3, I=4 for interrupt enable)
	Flags uint32
}

func NewCPUState(system *system.SystemDescriptor) (*CPUState, error) {
	if err := system.MemoryLayout.Validate(); err != nil {
		return nil, fmt.Errorf("invalid memory layout: %w", err)
	}

	peripheralsCollection := NewPeripherals(system.Peripherals)
	controller := NewInterruptController(system.MemoryLayout.VectorTableBase, system.VectorTable.VectorEntrySize, int(system.VectorTable.NumberOfVectors), peripheralsCollection)

	state := &CPUState{
		Halted:        false,
		Ram:           memory.NewSimulatedMemory(make([]byte, system.MemoryLayout.TotalSize)),
		MemoryLayout:  system.MemoryLayout,
		Peripherals:   peripheralsCollection,
		IntController: controller,
		SavedStates:   make([]InterruptState, 0, 1),
		MaxNesting:    0, // No nesting by default
		Flags:         0,
	}

	state.Reset()

	return state, nil
}

func (s *CPUState) Reset() {
	s.Registers.Reset()
	s.Halted = false
	s.Ram.Reset()
	s.Peripherals.Reset()
	s.IntController.Reset()
	s.SavedStates = s.SavedStates[:0]
	s.Flags = 0

	cpu.WritePC(s.Registers, s.MemoryLayout.CodeBase)
	cpu.WriteSP(s.Registers, s.MemoryLayout.StackBottom())
}

// Sets the interrupt enable flag.
func (s *CPUState) EnableInterrupts() {
	s.Flags |= FlagI
	if s.IntController != nil {
		s.IntController.Enable()
	}
}

// DisableInterrupts clears the interrupt enable flag.
func (s *CPUState) DisableInterrupts() {
	s.Flags &^= FlagI
	if s.IntController != nil {
		s.IntController.Disable()
	}
}

// Returns true if interrupts are enabled.
func (s *CPUState) InterruptsEnabled() bool {
	return s.Flags&FlagI != 0
}

// Sets the flags register value.
func (s *CPUState) SetFlags(flags uint32) {
	s.Flags = flags
	// Sync interrupt enable with controller
	if s.IntController != nil {
		if flags&FlagI != 0 {
			s.IntController.Enable()
		} else {
			s.IntController.Disable()
		}
	}
}

// Sets or clears the zero flag based on a value.
func (s *CPUState) SetZeroFlag(value uint32) {
	if value == 0 {
		s.Flags |= FlagZ
	} else {
		s.Flags &^= FlagZ
	}
}

// Sets or clears the negative flag based on a value.
func (s *CPUState) SetNegativeFlag(value uint32) {
	if int32(value) < 0 {
		s.Flags |= FlagN
	} else {
		s.Flags &^= FlagN
	}
}

// Returns the current interrupt nesting level.
func (s *CPUState) NestingLevel() int {
	return len(s.SavedStates)
}

// Returns true if another interrupt can be nested.
func (s *CPUState) CanNest() bool {
	return len(s.SavedStates) < s.MaxNesting
}

// SaveState saves the current state for interrupt handling.
func (s *CPUState) SaveState(pc uint32) error {
	if !s.CanNest() && len(s.SavedStates) > 0 {
		return fmt.Errorf("interrupt nesting limit reached")
	}

	s.SavedStates = append(s.SavedStates, InterruptState{
		PC:                pc,
		Flags:             s.Flags,
		InterruptsEnabled: s.InterruptsEnabled(),
	})

	return nil
}

// RestoreState restores the most recently saved state.
// Returns the saved PC value to return to.
func (s *CPUState) RestoreState() (uint32, error) {
	if len(s.SavedStates) == 0 {
		return 0, fmt.Errorf("no saved interrupt state to restore")
	}

	// Pop the last saved state
	idx := len(s.SavedStates) - 1
	saved := s.SavedStates[idx]
	s.SavedStates = s.SavedStates[:idx]

	// Restore flags
	s.SetFlags(saved.Flags)

	return saved.PC, nil
}

// ResetInterrupts clears all interrupt state.
func (s *CPUState) ResetInterrupts() {
	s.SavedStates = s.SavedStates[:0]
	s.Flags = 0
	if s.IntController != nil {
		s.IntController.Reset()
	}
}
