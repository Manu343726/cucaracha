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
