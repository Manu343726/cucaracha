package components

import (
	"github.com/Manu343726/cucaracha/pkg/hw/component"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/registers"
)

// =============================================================================
// Register Bank Component
// =============================================================================

// RegisterBank is a collection of registers with address-based selection
type RegisterBank struct {
	*component.BaseComponent

	// Ports
	data      *component.StandardPort // Bidirectional data port
	address   *component.StandardPort // Address input port
	readWrite *component.Pin          // Read/Write control (input)
	ready     *component.Pin          // Ready signal (output)

	// Internal state
	registers []uint64 // Register values
	count     int      // Number of registers
	width     int      // Bit width per register
}

// NewRegisterBank creates a new register bank with the specified count and width
func NewRegisterBank(name string, count, width int) *RegisterBank {
	rb := &RegisterBank{
		BaseComponent: component.NewBaseComponent(name, "REGISTER_BANK"),
		registers:     make([]uint64, count),
		count:         count,
		width:         width,
	}

	// Create bidirectional data port
	rb.data = component.NewPort("DATA", width)
	rb.AddInput(rb.data)
	rb.AddOutput(rb.data)

	// Create address input port (32-bit)
	rb.address = component.NewInputPort("ADDR", 32)
	rb.AddInput(rb.address)

	// Create control pins
	rb.readWrite = component.NewInputPin("RW")
	rb.AddInput(rb.readWrite)

	rb.ready = component.NewOutputPin("READY")
	rb.ready.Set(High) // Initially ready
	rb.AddOutput(rb.ready)

	return rb
}

// Data returns the bidirectional data port
func (rb *RegisterBank) Data() *component.StandardPort {
	return rb.data
}

// Address returns the address input port
func (rb *RegisterBank) Address() *component.StandardPort {
	return rb.address
}

// ReadWrite returns the read/write control pin
func (rb *RegisterBank) ReadWrite() *component.Pin {
	return rb.readWrite
}

// Ready returns the ready signal pin
func (rb *RegisterBank) Ready() *component.Pin {
	return rb.ready
}

// Count returns the number of registers in the bank
func (rb *RegisterBank) Count() int {
	return rb.count
}

// Width returns the bit width of each register
func (rb *RegisterBank) Width() int {
	return rb.width
}

// Get returns the value at the specified register index
func (rb *RegisterBank) Get(index uint32) uint64 {
	if index >= uint32(rb.count) {
		return 0
	}
	return rb.registers[index]
}

// Set directly sets the value at the specified register index
func (rb *RegisterBank) Set(index uint32, value uint64) {
	if index >= uint32(rb.count) {
		return
	}
	mask := uint64((1 << rb.width) - 1)
	rb.registers[index] = value & mask
}

// Clock performs the register bank operation on clock edge
func (rb *RegisterBank) Clock() error {
	if !rb.IsEnabled() {
		return nil
	}

	rb.ready.Set(Low) // Busy during operation

	addr := int(rb.address.GetValue())

	// Check address bounds
	if addr >= 0 && addr < rb.count {
		if rb.readWrite.IsHigh() {
			// Write operation: capture data from port into selected register
			mask := uint64((1 << rb.width) - 1)
			rb.registers[addr] = rb.data.GetValue() & mask
		} else {
			// Read operation: put selected register value on data port
			rb.data.SetValue(rb.registers[addr])
		}
	}

	rb.ready.Set(High) // Ready - operation complete
	return nil
}

// Reset clears all registers
func (rb *RegisterBank) Reset() {
	for i := range rb.registers {
		rb.registers[i] = 0
	}
	rb.data.Reset()
	rb.ready.Set(High)
}

// Implements registers.Registers interface Read method
func (rb *RegisterBank) Read(idx uint32) (uint32, error) {
	return uint32(rb.Get(idx)), nil
}

// Implements registers.Registers interface ReadByDescriptor method
func (rb *RegisterBank) ReadByDescriptor(regDesc *registers.RegisterDescriptor) (uint32, error) {
	return uint32(rb.Get(uint32(regDesc.Index))), nil
}

// Implements registers.Registers interface WriteByDescriptor method
func (rb *RegisterBank) WriteByDescriptor(regDesc *registers.RegisterDescriptor, value uint32) error {
	rb.Set(uint32(regDesc.Index), uint64(value))
	return nil
}

// Implements registers.Registers interface Write method
func (rb *RegisterBank) Write(idx uint32, value uint32) error {
	rb.Set(idx, uint64(value))
	return nil
}
