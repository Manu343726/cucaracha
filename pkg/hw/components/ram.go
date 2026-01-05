package components

import "github.com/Manu343726/cucaracha/pkg/hw/component"

// =============================================================================
// RAM Component
// =============================================================================

// RAM is a random access memory component (N bytes)
type RAM struct {
	*component.BaseComponent

	// Ports
	data      *component.StandardPort // 8-bit bidirectional data port
	address   *component.StandardPort // 32-bit address input port
	readWrite *component.Pin          // Read/Write control (input)
	ready     *component.Pin          // Ready signal (output)

	// Internal state - 8-bit register bank
	memory []byte // Byte storage
	size   int    // Size in bytes
}

// NewRAM creates a new RAM with the specified size in bytes
func NewRAM(name string, size int) *RAM {
	ram := &RAM{
		BaseComponent: component.NewBaseComponent(name, "RAM"),
		memory:        make([]byte, size),
		size:          size,
	}

	// Create 8-bit bidirectional data port
	ram.data = component.NewPort("DATA", 8)
	ram.AddInput(ram.data)
	ram.AddOutput(ram.data)

	// Create 32-bit address input port
	ram.address = component.NewInputPort("ADDR", 32)
	ram.AddInput(ram.address)

	// Create control pins
	ram.readWrite = component.NewInputPin("RW")
	ram.AddInput(ram.readWrite)

	ram.ready = component.NewOutputPin("READY")
	ram.ready.Set(High) // Initially ready
	ram.AddOutput(ram.ready)

	return ram
}

// Data returns the 8-bit bidirectional data port
func (ram *RAM) Data() *component.StandardPort {
	return ram.data
}

// Address returns the 32-bit address input port
func (ram *RAM) Address() *component.StandardPort {
	return ram.address
}

// ReadWrite returns the read/write control pin
func (ram *RAM) ReadWrite() *component.Pin {
	return ram.readWrite
}

// Ready returns the ready signal pin
func (ram *RAM) Ready() *component.Pin {
	return ram.ready
}

// Size returns the RAM size in bytes
func (ram *RAM) Size() int {
	return ram.size
}

// ReadByte returns the byte at the specified address
func (ram *RAM) ReadByte(addr uint32) byte {
	if addr >= uint32(ram.size) {
		return 0
	}
	return ram.memory[addr]
}

// WriteByte writes a byte at the specified address
func (ram *RAM) WriteByte(addr uint32, value byte) {
	if addr >= uint32(ram.size) {
		return
	}
	ram.memory[addr] = value
}

// Clock performs the RAM operation on clock edge
func (ram *RAM) Clock() error {
	if !ram.IsEnabled() {
		return nil
	}

	ram.ready.Set(Low) // Busy during operation

	addr := int(ram.address.GetValue())

	// Check address bounds
	if addr >= 0 && addr < ram.size {
		if ram.readWrite.IsHigh() {
			// Write operation: capture data from port into memory
			ram.memory[addr] = byte(ram.data.GetValue() & 0xFF)
		} else {
			// Read operation: put memory value on data port
			ram.data.SetValue(uint64(ram.memory[addr]))
		}
	}

	ram.ready.Set(High) // Ready - operation complete
	return nil
}

// Reset clears all memory
func (ram *RAM) Reset() {
	for i := range ram.memory {
		ram.memory[i] = 0
	}
	ram.data.Reset()
	ram.ready.Set(High)
}
