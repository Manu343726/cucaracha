package components

import (
	"github.com/Manu343726/cucaracha/pkg/hw/component"
)

// CategoryMemory is the category name for memory components
const CategoryMemory = "memory"

func init() {
	registerMemoryComponents()
}

func registerMemoryComponents() {
	// Register
	Registry.Register(component.NewDescriptor("REGISTER").
		DisplayName("32-bit Register").
		Description("A 32-bit register with read/write control and ready signaling").
		Category(CategoryMemory).
		Version("1.0.0").
		Input("RW", 1, "Read/Write control: High=Write, Low=Read").
		Output("READY", 1, "Ready signal: High=ready, Low=busy").
		Param("width", "int", 32, "Register width in bits").
		Factory(func(name string, params map[string]interface{}) (component.Component, error) {
			width := getIntParam(params, "width", 32)
			return NewRegister(name, width), nil
		}).
		Build())

	// Register Bank
	Registry.Register(component.NewDescriptor("REGISTER_BANK").
		DisplayName("Register Bank").
		Description("A bank of N registers with address-based selection").
		Category(CategoryMemory).
		Version("1.0.0").
		Input("ADDR", 32, "Address to select register").
		Input("RW", 1, "Read/Write control: High=Write, Low=Read").
		Output("READY", 1, "Ready signal: High=ready, Low=busy").
		Param("width", "int", 32, "Register width in bits").
		Param("count", "int", 16, "Number of registers").
		Factory(func(name string, params map[string]interface{}) (component.Component, error) {
			width := getIntParam(params, "width", 32)
			count := getIntParam(params, "count", 16)
			return NewRegisterBank(name, count, width), nil
		}).
		Build())

	// RAM
	Registry.Register(component.NewDescriptor("RAM").
		DisplayName("RAM").
		Description("Random Access Memory - N bytes of storage").
		Category(CategoryMemory).
		Version("1.0.0").
		Input("ADDR", 32, "Byte address").
		Input("RW", 1, "Read/Write control: High=Write, Low=Read").
		Output("READY", 1, "Ready signal: High=ready, Low=busy").
		Param("size", "int", 1024, "Size in bytes").
		Factory(func(name string, params map[string]interface{}) (component.Component, error) {
			size := getIntParam(params, "size", 1024)
			return NewRAM(name, size), nil
		}).
		Build())
}

// =============================================================================
// Register Component
// =============================================================================

// Reg32 is a memory component that stores a value
type Reg32 struct {
	*component.BaseComponent

	// Ports
	data      *component.StandardPort // Bidirectional data port
	readWrite *component.Pin          // Read/Write control (input)
	ready     *component.Pin          // Ready signal (output)

	// Internal state
	value uint64 // Stored value
	width int    // Bit width
}

// NewRegister creates a new register with the specified bit width
func NewRegister(name string, width int) *Reg32 {
	r := &Reg32{
		BaseComponent: component.NewBaseComponent(name, "REGISTER"),
		width:         width,
	}

	// Create bidirectional data port
	r.data = component.NewPort("DATA", width)
	r.AddInput(r.data)
	r.AddOutput(r.data)

	// Create control pins
	r.readWrite = component.NewInputPin("RW")
	r.AddInput(r.readWrite)

	r.ready = component.NewOutputPin("READY")
	r.ready.Set(High) // Initially ready
	r.AddOutput(r.ready)

	return r
}

// Data returns the bidirectional data port
func (r *Reg32) Data() *component.StandardPort {
	return r.data
}

// ReadWrite returns the read/write control pin
func (r *Reg32) ReadWrite() *component.Pin {
	return r.readWrite
}

// Ready returns the ready signal pin
func (r *Reg32) Ready() *component.Pin {
	return r.ready
}

// Value returns the currently stored value
func (r *Reg32) Value() uint64 {
	return r.value
}

// SetValue directly sets the stored value (for initialization)
func (r *Reg32) SetValue(v uint64) {
	mask := uint64((1 << r.width) - 1)
	r.value = v & mask
}

// Clock performs the register operation on clock edge
func (r *Reg32) Clock() error {
	if !r.IsEnabled() {
		return nil
	}

	r.ready.Set(Low) // Busy during operation

	if r.readWrite.IsHigh() {
		// Write operation: capture data from port
		r.value = r.data.GetValue()
	} else {
		// Read operation: put value on data port
		r.data.SetValue(r.value)
	}

	r.ready.Set(High) // Ready - operation complete
	return nil
}

// Reset clears the register
func (r *Reg32) Reset() {
	r.value = 0
	r.data.Reset()
	r.ready.Set(High)
}

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
func (rb *RegisterBank) Get(index int) uint64 {
	if index < 0 || index >= rb.count {
		return 0
	}
	return rb.registers[index]
}

// Set directly sets the value at the specified register index
func (rb *RegisterBank) Set(index int, value uint64) {
	if index < 0 || index >= rb.count {
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
func (ram *RAM) ReadByte(addr int) byte {
	if addr < 0 || addr >= ram.size {
		return 0
	}
	return ram.memory[addr]
}

// WriteByte writes a byte at the specified address
func (ram *RAM) WriteByte(addr int, value byte) {
	if addr < 0 || addr >= ram.size {
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
