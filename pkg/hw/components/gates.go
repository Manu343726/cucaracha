package components

import (
	"fmt"

	"github.com/Manu343726/cucaracha/pkg/hw/component"
)

// CategoryLogic is the category name for logic gates
const CategoryLogic = "logic"

// init registers all logic gate components
func init() {
	registerLogicGates()
}

func registerLogicGates() {
	// NOT Gate
	Registry.Register(component.NewDescriptor("NOT").
		DisplayName("NOT Gate").
		Description("Inverts the input signal").
		Category(CategoryLogic).
		Version("1.0.0").
		Input("A", 1, "Input signal").
		Output("Y", 1, "Inverted output").
		Factory(func(name string, params map[string]interface{}) (component.Component, error) {
			return NewNotGate(name), nil
		}).
		Build())

	// Buffer
	Registry.Register(component.NewDescriptor("BUFFER").
		DisplayName("Buffer").
		Description("Non-inverting buffer for signal propagation").
		Category(CategoryLogic).
		Version("1.0.0").
		Input("A", 1, "Input signal").
		Output("Y", 1, "Buffered output").
		Factory(func(name string, params map[string]interface{}) (component.Component, error) {
			return NewBuffer(name), nil
		}).
		Build())

	// AND Gate
	Registry.Register(component.NewDescriptor("AND").
		DisplayName("AND Gate").
		Description("Multi-input AND gate, outputs high only when all inputs are high").
		Category(CategoryLogic).
		Version("1.0.0").
		Param("numInputs", "int", 2, "Number of inputs (2-8)").
		Output("Y", 1, "AND result").
		Factory(func(name string, params map[string]interface{}) (component.Component, error) {
			numInputs := getIntParam(params, "numInputs", 2)
			return NewAndGate(name, numInputs), nil
		}).
		Build())

	// OR Gate
	Registry.Register(component.NewDescriptor("OR").
		DisplayName("OR Gate").
		Description("Multi-input OR gate, outputs high when any input is high").
		Category(CategoryLogic).
		Version("1.0.0").
		Param("numInputs", "int", 2, "Number of inputs (2-8)").
		Output("Y", 1, "OR result").
		Factory(func(name string, params map[string]interface{}) (component.Component, error) {
			numInputs := getIntParam(params, "numInputs", 2)
			return NewOrGate(name, numInputs), nil
		}).
		Build())

	// XOR Gate
	Registry.Register(component.NewDescriptor("XOR").
		DisplayName("XOR Gate").
		Description("Multi-input XOR gate, outputs high when odd number of inputs are high").
		Category(CategoryLogic).
		Version("1.0.0").
		Param("numInputs", "int", 2, "Number of inputs (2-8)").
		Output("Y", 1, "XOR result").
		Factory(func(name string, params map[string]interface{}) (component.Component, error) {
			numInputs := getIntParam(params, "numInputs", 2)
			return NewXorGate(name, numInputs), nil
		}).
		Build())

	// NAND Gate
	Registry.Register(component.NewDescriptor("NAND").
		DisplayName("NAND Gate").
		Description("Multi-input NAND gate (inverted AND)").
		Category(CategoryLogic).
		Version("1.0.0").
		Param("numInputs", "int", 2, "Number of inputs (2-8)").
		Output("Y", 1, "NAND result").
		Factory(func(name string, params map[string]interface{}) (component.Component, error) {
			numInputs := getIntParam(params, "numInputs", 2)
			return NewNandGate(name, numInputs), nil
		}).
		Build())

	// NOR Gate
	Registry.Register(component.NewDescriptor("NOR").
		DisplayName("NOR Gate").
		Description("Multi-input NOR gate (inverted OR)").
		Category(CategoryLogic).
		Version("1.0.0").
		Param("numInputs", "int", 2, "Number of inputs (2-8)").
		Output("Y", 1, "NOR result").
		Factory(func(name string, params map[string]interface{}) (component.Component, error) {
			numInputs := getIntParam(params, "numInputs", 2)
			return NewNorGate(name, numInputs), nil
		}).
		Build())

	// XNOR Gate
	Registry.Register(component.NewDescriptor("XNOR").
		DisplayName("XNOR Gate").
		Description("Multi-input XNOR gate (inverted XOR)").
		Category(CategoryLogic).
		Version("1.0.0").
		Param("numInputs", "int", 2, "Number of inputs (2-8)").
		Output("Y", 1, "XNOR result").
		Factory(func(name string, params map[string]interface{}) (component.Component, error) {
			numInputs := getIntParam(params, "numInputs", 2)
			return NewXnorGate(name, numInputs), nil
		}).
		Build())
}

// getIntParam safely extracts an int parameter with a default value
func getIntParam(params map[string]interface{}, name string, defaultVal int) int {
	if v, ok := params[name]; ok {
		switch val := v.(type) {
		case int:
			return val
		case float64:
			return int(val)
		case int64:
			return int(val)
		}
	}
	return defaultVal
}

// createGateInputDescriptors creates input descriptors for a gate
func createGateInputDescriptors(numInputs int) []component.PortDescriptor {
	inputs := make([]component.PortDescriptor, numInputs)
	for i := 0; i < numInputs; i++ {
		inputs[i] = component.PortDescriptor{
			Name:        string(rune('A' + i)),
			Width:       1,
			Direction:   component.Input,
			Description: fmt.Sprintf("Input %c", rune('A'+i)),
		}
	}
	return inputs
}

// =============================================================================
// Basic Logic Gates
// =============================================================================

// Gate is a basic logic gate component
type Gate struct {
	*component.BaseComponent
	inputs  []*component.Pin
	output  *component.Pin
	compute func(inputs []BitValue) BitValue
}

// NewGate creates a new logic gate with the specified number of inputs
func NewGate(name, gateType string, numInputs int, compute func([]BitValue) BitValue) *Gate {
	g := &Gate{
		BaseComponent: component.NewBaseComponent(name, gateType),
		inputs:        make([]*component.Pin, numInputs),
		compute:       compute,
	}

	// Create input pins
	for i := 0; i < numInputs; i++ {
		pin := component.NewInputPin(string(rune('A' + i)))
		g.inputs[i] = pin
		g.AddInput(pin)
	}

	// Create output pin
	g.output = component.NewOutputPin("Y")
	g.AddOutput(g.output)

	return g
}

// Compute evaluates the gate and updates the output
func (g *Gate) Compute() error {
	if !g.IsEnabled() {
		return nil
	}

	inputs := make([]BitValue, len(g.inputs))
	for i, pin := range g.inputs {
		inputs[i] = pin.Get()
	}

	result := g.compute(inputs)
	return g.output.Set(result)
}

// Output returns the gate's output pin
func (g *Gate) Output() *component.Pin {
	return g.output
}

// Input returns the gate's input pin at the given index
func (g *Gate) Input(idx int) *component.Pin {
	if idx < 0 || idx >= len(g.inputs) {
		return nil
	}
	return g.inputs[idx]
}

// =============================================================================
// NOT Gate (Inverter)
// =============================================================================

// NotGate is a single-input inverter
type NotGate struct {
	*Gate
}

// NewNotGate creates a NOT gate
func NewNotGate(name string) *NotGate {
	return &NotGate{
		Gate: NewGate(name, "NOT", 1, func(inputs []BitValue) BitValue {
			if inputs[0] == Low {
				return High
			}
			return Low
		}),
	}
}

// =============================================================================
// AND Gate
// =============================================================================

// AndGate is a multi-input AND gate
type AndGate struct {
	*Gate
}

// NewAndGate creates an AND gate with the specified number of inputs
func NewAndGate(name string, numInputs int) *AndGate {
	return &AndGate{
		Gate: NewGate(name, "AND", numInputs, func(inputs []BitValue) BitValue {
			for _, v := range inputs {
				if v == Low {
					return Low
				}
			}
			return High
		}),
	}
}

// =============================================================================
// OR Gate
// =============================================================================

// OrGate is a multi-input OR gate
type OrGate struct {
	*Gate
}

// NewOrGate creates an OR gate with the specified number of inputs
func NewOrGate(name string, numInputs int) *OrGate {
	return &OrGate{
		Gate: NewGate(name, "OR", numInputs, func(inputs []BitValue) BitValue {
			for _, v := range inputs {
				if v == High {
					return High
				}
			}
			return Low
		}),
	}
}

// =============================================================================
// XOR Gate
// =============================================================================

// XorGate is a multi-input XOR gate
type XorGate struct {
	*Gate
}

// NewXorGate creates an XOR gate with the specified number of inputs
func NewXorGate(name string, numInputs int) *XorGate {
	return &XorGate{
		Gate: NewGate(name, "XOR", numInputs, func(inputs []BitValue) BitValue {
			count := 0
			for _, v := range inputs {
				if v == High {
					count++
				}
			}
			if count%2 == 1 {
				return High
			}
			return Low
		}),
	}
}

// =============================================================================
// NAND Gate
// =============================================================================

// NandGate is a multi-input NAND gate
type NandGate struct {
	*Gate
}

// NewNandGate creates a NAND gate with the specified number of inputs
func NewNandGate(name string, numInputs int) *NandGate {
	return &NandGate{
		Gate: NewGate(name, "NAND", numInputs, func(inputs []BitValue) BitValue {
			for _, v := range inputs {
				if v == Low {
					return High
				}
			}
			return Low
		}),
	}
}

// =============================================================================
// NOR Gate
// =============================================================================

// NorGate is a multi-input NOR gate
type NorGate struct {
	*Gate
}

// NewNorGate creates a NOR gate with the specified number of inputs
func NewNorGate(name string, numInputs int) *NorGate {
	return &NorGate{
		Gate: NewGate(name, "NOR", numInputs, func(inputs []BitValue) BitValue {
			for _, v := range inputs {
				if v == High {
					return Low
				}
			}
			return High
		}),
	}
}

// =============================================================================
// XNOR Gate
// =============================================================================

// XnorGate is a multi-input XNOR gate
type XnorGate struct {
	*Gate
}

// NewXnorGate creates an XNOR gate with the specified number of inputs
func NewXnorGate(name string, numInputs int) *XnorGate {
	return &XnorGate{
		Gate: NewGate(name, "XNOR", numInputs, func(inputs []BitValue) BitValue {
			count := 0
			for _, v := range inputs {
				if v == High {
					count++
				}
			}
			if count%2 == 0 {
				return High
			}
			return Low
		}),
	}
}

// =============================================================================
// Buffer
// =============================================================================

// Buffer is a non-inverting buffer (useful for signal propagation)
type Buffer struct {
	*Gate
}

// NewBuffer creates a buffer gate
func NewBuffer(name string) *Buffer {
	return &Buffer{
		Gate: NewGate(name, "BUFFER", 1, func(inputs []BitValue) BitValue {
			return inputs[0]
		}),
	}
}
