// Package peripheral provides interfaces and utilities for memory-mapped peripherals.
package peripheral

// Descriptor describes a type of peripheral device.
// This provides metadata and a factory for creating peripheral instances.
type Descriptor struct {
	// Peripheral type
	Type Type

	// DisplayName is a human-readable name (e.g., "UART Serial Port").
	DisplayName string

	// Description provides documentation about the peripheral.
	Description string

	// DefaultSize is the default memory-mapped region size in bytes.
	DefaultSize uint32

	// DefaultAlignment is the default address alignment requirement.
	DefaultAlignment uint32

	// HasInterrupt indicates if this peripheral typically uses interrupts.
	HasInterrupt bool

	// DefaultInterruptVector is the default interrupt vector if HasInterrupt is true.
	DefaultInterruptVector uint8

	// Factory creates an instance of this peripheral type.
	// Standard parameters:
	//   - name: unique identifier for this instance
	//   - baseAddress: memory-mapped base address
	//   - interruptVector: interrupt vector number (0xFF = no interrupt)
	// Extra parameters are peripheral-specific (e.g., buffer sizes, frequencies).
	Factory func(params PeripheralParams) (Peripheral, error)
}

// PeripheralParams contains parameters for creating a peripheral instance.
type PeripheralParams struct {
	// Name is the unique identifier for this instance.
	Name string

	// BaseAddress is the memory-mapped base address.
	BaseAddress uint32

	// Size is the memory-mapped region size (0 = use default).
	Size uint32

	// InterruptVector is the interrupt vector number (0xFF = no interrupt).
	InterruptVector uint8

	// Instance is the instance number for multiple peripherals of the same type.
	Instance uint8

	// Description of the specific purpose and usage of this peripheral instance.
	Description string

	// Extra contains peripheral-specific parameters.
	// Keys and expected types depend on the peripheral type.
	Extra map[string]interface{}
}

// GetExtra retrieves an extra parameter with type assertion.
func (p PeripheralParams) GetExtra(key string, defaultValue interface{}) interface{} {
	if p.Extra == nil {
		return defaultValue
	}
	if v, ok := p.Extra[key]; ok {
		return v
	}
	return defaultValue
}

// GetExtraInt retrieves an integer extra parameter.
func (p PeripheralParams) GetExtraInt(key string, defaultValue int) int {
	v := p.GetExtra(key, defaultValue)
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case uint32:
		return int(val)
	case float64:
		return int(val)
	default:
		return defaultValue
	}
}

// GetExtraBool retrieves a boolean extra parameter.
func (p PeripheralParams) GetExtraBool(key string, defaultValue bool) bool {
	v := p.GetExtra(key, defaultValue)
	if b, ok := v.(bool); ok {
		return b
	}
	return defaultValue
}

// GetExtraString retrieves a string extra parameter.
func (p PeripheralParams) GetExtraString(key, defaultValue string) string {
	v := p.GetExtra(key, defaultValue)
	if s, ok := v.(string); ok {
		return s
	}
	return defaultValue
}
