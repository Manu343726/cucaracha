// Package components provides a library of pre-built hardware components
// for constructing digital circuits.
//
// This package builds on the primitives from the component package to provide
// ready-to-use implementations of common hardware building blocks:
//
//   - Logic gates (AND, OR, NOT, XOR, NAND, NOR, XNOR)
//   - Multiplexers and demultiplexers
//   - Registers and latches
//   - Arithmetic units (adders, ALUs)
//   - Memory components
//   - Counters and timers
//
// Each component follows the Component interface from the component package
// and can be composed into larger circuits.
package components

import (
	"github.com/Manu343726/cucaracha/pkg/hw/component"
)

// Re-export commonly used types for convenience
type (
	// Port is a connection point on a component
	Port = component.Port

	// Component is a hardware component interface
	Component = component.Component

	// BitValue represents a single bit (High or Low)
	BitValue = component.BitValue

	// Direction indicates whether a port is input, output, or bidirectional
	Direction = component.Direction
)

// Re-export constants
const (
	Low           = component.Low
	High          = component.High
	Input         = component.Input
	Output        = component.Output
	Bidirectional = component.Bidirectional
)

// Re-export constructors for convenience
var (
	NewPort          = component.NewPort
	NewInputPort     = component.NewInputPort
	NewOutputPort    = component.NewOutputPort
	NewPin           = component.NewPin
	NewInputPin      = component.NewInputPin
	NewOutputPin     = component.NewOutputPin
	NewBaseComponent = component.NewBaseComponent
	NewBus           = component.NewBus
	NewCircuit       = component.NewCircuit
	NewDescriptor    = component.NewDescriptor
)

// Registry is the component library registry
var Registry = component.NewRegistry()

// Register adds a component to the library registry
func Register(desc *component.ComponentDescriptor) error {
	return Registry.Register(desc)
}

// Get retrieves a component descriptor by name
func Get(name string) (*component.ComponentDescriptor, error) {
	return Registry.Get(name)
}

// Create instantiates a component by name from the library
func Create(componentName, instanceName string, params map[string]interface{}) (Component, error) {
	return Registry.Create(componentName, instanceName, params)
}

// List returns all registered component names
func List() []string {
	return Registry.List()
}

// ListByCategory returns component names in a category
func ListByCategory(category string) []string {
	return Registry.ListByCategory(category)
}

// Categories returns all registered categories
func Categories() []string {
	return Registry.Categories()
}

// Search finds components matching a query
func Search(query component.SearchQuery) []*component.ComponentDescriptor {
	return Registry.Search(query)
}
