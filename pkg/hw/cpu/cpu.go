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

	// If the step triggered a breakpoint, the address of the instruction that caused the breakpoint to be hit
	BreakpointHit *uint32

	// If the step triggered a watchpoint, the range of memory that caused the watchpoint to be hit
	WatchpointHit *memory.Range
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
func DecodeCurrentInstruction(registers Registers, ram memory.Memory) (*instructions.Instruction, error) {
	pc, err := ReadPC(registers)
	if err != nil {
		return nil, fmt.Errorf("error decoding current CPU instruction: error reading PC register: %w", err)
	}

	return DecodeInstruction(ram, pc)
}

// Returns the target address of a branch instruction in memory
//
// The target address is obtained by reading the register that contains the branch target
// address as specified by the instruction operands. If this function is called when the given
// instruction is not the current instruction being executed by the CPU, the returned
// address may not correspond to the actual branch target address used during execution.
//
// To get a more accurate branch target address for common cases, such as compile-time resolved
// branches where the target address is hardcoded, use program.BranchTargetAddress() instead, which
// uses debug information to read the target address of hardcoded branches.
func BranchTargetAddress(instr *instructions.Instruction, registers Registers) (uint32, error) {
	targetReg, err := instructions.BranchTargetRegister(instr)
	if err != nil {
		return 0, fmt.Errorf("error getting branch target address register: %w", err)
	}

	return registers.ReadByDescriptor(targetReg)
}
