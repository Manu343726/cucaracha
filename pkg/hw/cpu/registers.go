package cpu

import (
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/registers"
)

// Defines an interface for manipulating CPU registers
type Registers interface {
	// Get retrieves the value of the register at the given index
	Read(idx uint32) (uint32, error)
	// Set sets the value of the register at the given index
	Write(idx uint32, value uint32) error

	// Retrieves the value of the register described by regDesc
	ReadByDescriptor(regDesc *registers.RegisterDescriptor) (uint32, error)
	// Sets the value of the register described by regDesc
	WriteByDescriptor(regDesc *registers.RegisterDescriptor, value uint32) error

	// Resets all registers to their default state
	Reset() error
}

// Returns the value of SP register
func ReadSP(regs Registers) (uint32, error) {
	return regs.ReadByDescriptor(mc.Descriptor.Registers.SP)
}

// Sets the value of SP register
func WriteSP(regs Registers, value uint32) error {
	return regs.WriteByDescriptor(mc.Descriptor.Registers.SP, value)
}

// Returns the value of LR register
func ReadLR(regs Registers) (uint32, error) {
	return regs.ReadByDescriptor(mc.Descriptor.Registers.LR)
}

// Sets the value of LR register
func WriteLR(regs Registers, value uint32) error {
	return regs.WriteByDescriptor(mc.Descriptor.Registers.LR, value)
}

// Returns the value of PC register
func ReadPC(regs Registers) (uint32, error) {
	return regs.ReadByDescriptor(mc.Descriptor.Registers.PC)
}

// Advances the PC register by N instructions
func AdvancePC(regs Registers, n uint32) error {
	pc, err := ReadPC(regs)
	if err != nil {
		return err
	}
	return WritePC(regs, pc+n*4) // Instructions are 4 bytes
}

// Advances the PC register N instructions if current PC is equal to expectedPC
func AdvancePCIfEqual(regs Registers, expectedPC uint32, n uint32) error {
	pc, err := ReadPC(regs)
	if err != nil {
		return err
	}
	if pc == expectedPC {
		return WritePC(regs, pc+n*4) // Instructions are 4 bytes
	}
	return nil
}

// Sets the value of PC register
func WritePC(regs Registers, value uint32) error {
	return regs.WriteByDescriptor(mc.Descriptor.Registers.PC, value)
}

// Returns the value of CPSR register
func ReadCPSR(regs Registers) (uint32, error) {
	return regs.ReadByDescriptor(mc.Descriptor.Registers.CPSR)
}

// Sets the value of CPSR register
func WriteCPSR(regs Registers, value uint32) error {
	return regs.WriteByDescriptor(mc.Descriptor.Registers.CPSR, value)
}
