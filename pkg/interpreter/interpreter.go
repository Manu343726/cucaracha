// Package interpreter provides an automatic interpreter for Cucaracha machine code
// based on the instruction descriptors.
package interpreter

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu"
	"github.com/Manu343726/cucaracha/pkg/hw/memory"
	"github.com/Manu343726/cucaracha/pkg/hw/peripheral"
	"github.com/Manu343726/cucaracha/pkg/system"
	"github.com/Manu343726/cucaracha/pkg/utils/contract"
	"github.com/Manu343726/cucaracha/pkg/utils/logging"
)

// --- Interpreter ---

// Executes Cucaracha machine code using the instruction descriptors.
// It supports peripherals and interrupts.
type Interpreter struct {
	contract.Base

	state *CPUState

	// Target execution speed in Hz (cycles per second)
	// 0 means unlimited (full speed, no timing simulation)
	targetSpeedHz float64
	// Track timing for speed control
	cycleAccumulator int64     // Accumulated cycles since last timing reset
	timingStartTime  time.Time // When timing measurement started

	breakpoints       map[uint32]struct{} // Set of breakpoint addresses
	lastBreakpointHit *uint32
}

// Creates a new interpreter with the given system configuration.
func NewInterpreter(system *system.SystemDescriptor) (*Interpreter, error) {
	state, err := NewCPUState(system)
	if err != nil {
		return nil, fmt.Errorf("failed to setup CPU state: %w", err)
	}
	interp := &Interpreter{
		Base:        contract.NewBase(log().Child("interpreter")),
		state:       state,
		breakpoints: make(map[uint32]struct{}),
	}
	return interp, nil
}

func (i *Interpreter) SetBreakpoint(addr uint32) error {
	i.Log().Debug("SetBreakpoint", slog.Uint64("addr", uint64(addr)))

	if _, exists := i.breakpoints[addr]; exists {
		return i.Log().Errorf("breakpoint already exists at address 0x%X", addr)
	}

	i.breakpoints[addr] = struct{}{}

	i.Log().Info("breakpoint set", logging.Address("address", addr))
	return nil
}

func (i *Interpreter) ClearBreakpoint(addr uint32) error {
	i.Log().Debug("ClearBreakpoint", slog.Uint64("addr", uint64(addr)))

	if _, exists := i.breakpoints[addr]; !exists {
		return i.Log().Errorf("no breakpoint found at address 0x%X", addr)
	}

	delete(i.breakpoints, addr)

	i.Log().Info("breakpoint cleared", logging.Address("address", addr))
	return nil
}

func (i *Interpreter) SetWatchpoint(r memory.Range) error {
	i.Log().Debug("SetWatchpoint", slog.Uint64("start", uint64(r.Start)), slog.Uint64("Size", uint64(r.Size)))
	// TODO: Implement watchpoint logic
	return nil
}

func (i *Interpreter) ClearWatchpoint(r memory.Range) error {
	i.Log().Debug("ClearWatchpoint", slog.Uint64("start", uint64(r.Start)), slog.Uint64("Size", uint64(r.Size)))

	// TODO: Implement watchpoint logic
	return nil
}

// Sets the target execution speed in Hz (cycles per second).
// Use 0 for unlimited speed (no timing simulation).
func (i *Interpreter) SetTargetSpeed(hz float64) {
	if hz < 0 {
		hz = 0
	}
	i.targetSpeedHz = hz
	i.ResetTiming()
}

// Returns the current target execution speed in Hz.
func (i *Interpreter) GetTargetSpeed() float64 {
	return i.targetSpeedHz
}

// ResetTiming resets the timing accumulator for speed control.
func (i *Interpreter) ResetTiming() {
	i.cycleAccumulator = 0
	i.timingStartTime = time.Now()
}

// Deprecated. Use SetTargetSpeed instead.
func (i *Interpreter) SetExecutionDelay(delayMs int) {
	if delayMs <= 0 {
		i.SetTargetSpeed(0)
	} else {
		hz := 1000.0 / float64(delayMs)
		i.SetTargetSpeed(hz)
	}
}

// Deprecated. Use GetTargetSpeed instead.
func (i *Interpreter) GetExecutionDelay() int {
	if i.targetSpeedHz <= 0 {
		return 0
	}
	return int(1000.0 / i.targetSpeedHz)
}

func (i *Interpreter) Registers() cpu.Registers {
	return i.state.Registers
}

func (i *Interpreter) Interrupts() cpu.Interrupts {
	return i.state.IntController
}

func (i *Interpreter) Memory() memory.Memory {
	return i.state.Ram
}

func (i *Interpreter) MemoryLayout() memory.MemoryLayout {
	return i.state.MemoryLayout
}

func (i *Interpreter) CPU() cpu.CPU {
	return i
}

func (i *Interpreter) Peripherals() map[string]peripheral.Peripheral {
	return i.state.Peripherals.byName
}

// Executes a single instruction, handling interrupts and peripherals.
func (i *Interpreter) Step() (*cpu.StepInfo, error) {
	log := i.Log().Child("Step")

	if i.state.Halted {
		log.Debug("CPU is halted")
		return nil, log.Errorf("CPU is halted")
	}

	// Save current PC for detecting branches
	oldPC, err := cpu.ReadPC(i.Registers())
	if err != nil {
		return nil, log.Errorf("failed to read PC: %v", err)
	}

	if !i.MemoryLayout().Code().ContainsAddress(oldPC) {
		return nil, log.Errorf("PC outside of code memory range: 0x%X (code range: 0x%X - 0x%X)", oldPC, i.MemoryLayout().Code().Start, i.MemoryLayout().Code().End)
	}

	log = log.WithAttrs(logging.Address("pc", oldPC))

	if i.lastBreakpointHit != nil && *i.lastBreakpointHit == oldPC {
		log.Info("resuming from breakpoint")
		i.lastBreakpointHit = nil
	} else if _, isBreakpoint := i.breakpoints[oldPC]; isBreakpoint {
		i.lastBreakpointHit = &oldPC

		info := &cpu.StepInfo{
			CyclesUsed:             0,
			Halted:                 i.IsHalted(),
			InstructionAddress:     oldPC,
			NextInstructionAddress: oldPC,
			BreakpointHit:          &oldPC,
		}

		log.Debug("finished (breakpoint hit)", slog.Uint64("cycles", uint64(info.CyclesUsed)), slog.Bool("halted", info.Halted))
		return info, nil
	}

	instruction, err := cpu.DecodeCurrentInstruction(i.Registers(), i.Memory())
	if err != nil {
		return nil, log.Errorf("failed to decode instruction: %v", err)
	}

	log = log.WithAttrs(instruction.LoggingAttribute("instruction"))
	log.Debug("executing")

	// Check for pending interrupts before executing
	if err := i.checkInterrupts(); err != nil {
		return nil, log.Errorf("Interrupt handling failed: %v", err)
	}

	// Execute the instruction
	if err := ExecuteInstruction(i, instruction); err != nil {
		return nil, err
	}

	// Advance PC one instruction if it wasn't changed by the instruction itself (e.g., JMP)
	if err := cpu.AdvancePCIfEqual(i.Registers(), oldPC, 1); err != nil {
		return nil, log.Errorf("failed to advance PC: %w", err)
	}

	newPC, err := cpu.ReadPC(i.Registers())
	if err != nil {
		return nil, log.Errorf("failed to read new PC: %v", err)
	}

	if !i.MemoryLayout().Code().ContainsAddress(newPC) {
		i.Log().Warn("PC is outside of code memory range after instruction execution", instruction.LoggingAttribute("instruction"), logging.Address("pc", newPC), logging.Address("code_start", i.MemoryLayout().Code().Start), logging.Address("code_end", i.MemoryLayout().Code().End()))
	}

	// Clock peripherals
	env := peripheral.Environment{
		MemoryLayout: i.state.MemoryLayout,
		RAM:          i.state.Ram,
	}

	if err := i.state.Peripherals.Clock(env); err != nil {
		return nil, log.Errorf("peripheral clock error: %v", err)
	}

	// Poll interrupt sources after each instruction
	i.state.IntController.Poll()

	info := &cpu.StepInfo{
		CyclesUsed:             instruction.Descriptor.Cycles,
		Halted:                 i.IsHalted(),
		InstructionAddress:     oldPC,
		NextInstructionAddress: newPC,
	}

	log.Debug("finished", slog.Uint64("cycles", uint64(info.CyclesUsed)), slog.Bool("halted", info.Halted), logging.Address("next_pc", newPC))
	return info, nil
}

// checkInterrupts checks for and handles any pending interrupts.
func (i *Interpreter) checkInterrupts() error {
	if !i.state.IntController.HasPendingInterrupt() {
		return nil
	}

	vector := i.state.IntController.GetNextInterrupt()
	if vector < 0 {
		return nil
	}

	pc, err := cpu.ReadPC(i.state.Registers)
	if err != nil {
		return fmt.Errorf("failed to read PC during interrupt: %w", err)
	}

	// Save current state
	if err := i.state.SaveState(pc); err != nil {
		return fmt.Errorf("failed to save interrupt state: %w", err)
	}

	// Disable interrupts while servicing
	i.state.DisableInterrupts()

	// Get handler address from vector table
	handlerAddrLoc := i.state.IntController.BeginService(vector)
	handlerAddr, err := memory.ReadUint32(i.state.Ram, handlerAddrLoc)
	if err != nil {
		return fmt.Errorf("failed to read interrupt handler address: %w", err)
	}

	// Jump to interrupt handler
	if err := cpu.WritePC(i.state.Registers, handlerAddr); err != nil {
		return fmt.Errorf("failed to write PC during interrupt: %w", err)
	}

	return nil
}

// ReturnFromInterrupt restores state and returns from an interrupt handler.
func (i *Interpreter) ReturnFromInterrupt() error {
	// Mark interrupt as complete
	i.state.IntController.EndService()

	// Restore saved state
	pc, err := i.state.RestoreState()
	if err != nil {
		return err
	}

	if err := cpu.WritePC(i.state.Registers, pc); err != nil {
		return fmt.Errorf("failed to write PC during interrupt return: %w", err)
	}

	return nil
}

// SoftwareInterrupt triggers a software interrupt with the given vector.
func (i *Interpreter) SoftwareInterrupt(vector uint8) error {
	i.state.IntController.SetPending(vector)
	return nil
}

// Reset resets the CPU state including memory, registers, and interrupt state.
func (i *Interpreter) Reset() error {
	i.state.Reset()
	i.ResetTiming()
	return nil
}

// Sets the CPU to a halted state.
func (i *Interpreter) Halt() error {
	i.state.Halted = true
	return nil
}

// Returns true if the CPU is currently halted.
func (i *Interpreter) IsHalted() bool {
	return i.state.Halted
}
