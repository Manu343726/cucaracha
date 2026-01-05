// Package component provides hardware abstraction interfaces for cucaracha components.
package component

import (
	"fmt"
	"sync"
)

// Represents a hardware component with inputs, outputs, and internal state.
type Component interface {
	// Identity
	Name() string
	Type() string

	// Port access
	Inputs() []Port
	Outputs() []Port
	GetInput(name string) (Port, error)
	GetOutput(name string) (Port, error)
	GetPort(name string) (Port, error)

	// Lifecycle
	Reset()
	Clock() error // Advance one clock cycle (for synchronous components)

	// State
	IsEnabled() bool
	Enable()
	Disable()
}

// =============================================================================
// Base Component Implementation
// =============================================================================

// Provides a standard implementation of the Component interface
// that can be embedded in concrete component types.
type BaseComponent struct {
	name     string
	compType string
	enabled  bool
	mu       sync.RWMutex

	inputs  map[string]Port
	outputs map[string]Port

	// Clock callback (optional)
	onClock func() error

	// Reset callback (optional)
	onReset func()
}

// A functional option for configuring a BaseComponent
type ComponentOption func(*BaseComponent)

// Sets a callback for clock events
func WithClock(callback func() error) ComponentOption {
	return func(c *BaseComponent) {
		c.onClock = callback
	}
}

// Sets a callback for reset events
func WithReset(callback func()) ComponentOption {
	return func(c *BaseComponent) {
		c.onReset = callback
	}
}

// Sets the initial enabled state
func WithEnabled(enabled bool) ComponentOption {
	return func(c *BaseComponent) {
		c.enabled = enabled
	}
}

// Creates a new base component
func NewBaseComponent(name, compType string, opts ...ComponentOption) *BaseComponent {
	c := &BaseComponent{
		name:     name,
		compType: compType,
		enabled:  true,
		inputs:   make(map[string]Port),
		outputs:  make(map[string]Port),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Returns the component name
func (c *BaseComponent) Name() string {
	return c.name
}

// Returns the component type
func (c *BaseComponent) Type() string {
	return c.compType
}

// AddInput adds an input port to the component
func (c *BaseComponent) AddInput(port Port) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.inputs[port.Name()]; exists {
		return fmt.Errorf("input port %q already exists", port.Name())
	}
	c.inputs[port.Name()] = port
	return nil
}

// AddOutput adds an output port to the component
func (c *BaseComponent) AddOutput(port Port) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.outputs[port.Name()]; exists {
		return fmt.Errorf("output port %q already exists", port.Name())
	}
	c.outputs[port.Name()] = port
	return nil
}

// Returns all input ports
func (c *BaseComponent) Inputs() []Port {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]Port, 0, len(c.inputs))
	for _, port := range c.inputs {
		result = append(result, port)
	}
	return result
}

// Returns all output ports
func (c *BaseComponent) Outputs() []Port {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]Port, 0, len(c.outputs))
	for _, port := range c.outputs {
		result = append(result, port)
	}
	return result
}

// Returns an input port by name
func (c *BaseComponent) GetInput(name string) (Port, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	port, exists := c.inputs[name]
	if !exists {
		return nil, fmt.Errorf("input port %q not found", name)
	}
	return port, nil
}

// GetOutput returns an output port by name
func (c *BaseComponent) GetOutput(name string) (Port, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	port, exists := c.outputs[name]
	if !exists {
		return nil, fmt.Errorf("output port %q not found", name)
	}
	return port, nil
}

// GetPort returns any port by name (checks both inputs and outputs)
func (c *BaseComponent) GetPort(name string) (Port, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if port, exists := c.inputs[name]; exists {
		return port, nil
	}
	if port, exists := c.outputs[name]; exists {
		return port, nil
	}
	return nil, fmt.Errorf("port %q not found", name)
}

// Reset resets the component and all its ports
func (c *BaseComponent) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, port := range c.inputs {
		port.Reset()
	}
	for _, port := range c.outputs {
		port.Reset()
	}

	if c.onReset != nil {
		c.onReset()
	}
}

// Clock advances the component by one clock cycle
func (c *BaseComponent) Clock() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.enabled {
		return nil
	}

	if c.onClock != nil {
		return c.onClock()
	}
	return nil
}

// IsEnabled returns whether the component is enabled
func (c *BaseComponent) IsEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.enabled
}

// Enable enables the component
func (c *BaseComponent) Enable() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enabled = true
}

// Disable disables the component
func (c *BaseComponent) Disable() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enabled = false
}

// String returns a string representation of the component
func (c *BaseComponent) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	status := "enabled"
	if !c.enabled {
		status = "disabled"
	}

	return fmt.Sprintf("%s[%s] (%d inputs, %d outputs) [%s]",
		c.name, c.compType, len(c.inputs), len(c.outputs), status)
}
