// Package cpu provides the Cucaracha CPU abstraction and implementations.
package cpu

import (
	"fmt"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/Manu343726/cucaracha/pkg/hw/memory"
)

// Contains statistics about interruptions during a CPU step
type InterruptionStatistics struct {
	GeneratedInterrupts []uint32
	HandledInterrupts   []uint32
	FinishedInterrupts  []uint32
}

// Returns information about the execution of a CPU step
type StepInfo struct {
	// Number of  cycles used during the last step
	CyclesUsed int
	// Details about interruptions during the last step
	InterruptionDetails InterruptionStatistics
	// Whether the CPU is halted after the last step
	Halted bool
}

// CPU defines the minimal contract for a Cucaracha CPU implementation.
type CPU interface {
	// Provides access to CPU registers and their manipulation
	Registers() Registers
	// Provides access to interruption handling features of the CPU
	Interrupts() Interrupts

	// Executes a single CPU step, returning information about the execution
	Step() (*StepInfo, error)

	// Checks whether the CPU is currently halted
	IsHalted() bool

	// Halts the CPU execution
	Halt() error

	// Resets the CPU to its initial state
	Reset() error
}

// Decodes the instruction at the given address in memory
func DecodeInstruction(ram memory.Memory, addr uint32) (*instructions.Instruction, error) {
	instructionData, err := memory.ReadUint32(ram, addr)
	if err != nil {
		return nil, fmt.Errorf("error decoding instruction at address 0x%X: error reading memory: %w", addr, err)
	}

	instr, err := mc.Descriptor.Instructions.Decode(instructionData)
	if err != nil {
		return nil, fmt.Errorf("error decoding instruction at address 0x%X: %w", addr, err)
	}

	return instr, nil
}

// Decodes the instruction currently pointed to by the PC register
func DecodeCurrentInstruction(cpu CPU, ram memory.Memory) (*instructions.Instruction, error) {
	pc, err := ReadPC(cpu.Registers())
	if err != nil {
		return nil, fmt.Errorf("error decoding current CPU instruction: error reading PC register: %w", err)
	}

	return DecodeInstruction(ram, pc)
}

func LoadProgram(cpu CPU, ram memory.Memory, layout memory.MemoryLayout, program *mc.Program) error {
	if err := cpu.Reset(); err != nil {
		return fmt.Errorf("failed to reset CPU before loading program: %w", err)
	}

	binary, err := program.Encode()
	if err != nil {
		return fmt.Errorf("failed to encode program: %w", err)
	}

	if uint32(len(binary))+layout.CodeBase > uint32(ram.Size()) {
		return fmt.Errorf("program too large to fit in memory at address 0x%X", layout.CodeBase)
	}

	for offset, b := range binary {
		err := ram.WriteByte(layout.CodeBase+uint32(offset), b)
		if err != nil {
			return fmt.Errorf("failed to write program to memory at address 0x%X: %w", layout.CodeBase+uint32(offset), err)
		}
	}

	err = WritePC(cpu.Registers(), layout.CodeBase)
	if err != nil {
		return fmt.Errorf("failed to set PC to program start address 0x%X: %w", layout.CodeBase, err)
	}

	return nil
}
