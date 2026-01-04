package components

import (
	"github.com/Manu343726/cucaracha/pkg/hw/component"
)

func init() {
	registerMuxComponents()
}

const CategoryMux = "mux"

func registerMuxComponents() {
	// MUX2 - 2:1 Multiplexer
	Registry.Register(component.NewDescriptor("MUX2").
		DisplayName("2:1 MUX").
		Description("2-to-1 Multiplexer - selects between two inputs").
		Category(CategoryMux).
		Version("1.0.0").
		Input("A", 32, "First input").
		Input("B", 32, "Second input").
		Input("SEL", 1, "Select signal: 0=A, 1=B").
		Output("OUT", 32, "Selected output").
		Param("width", "int", 32, "Data width in bits").
		Factory(func(name string, params map[string]interface{}) (component.Component, error) {
			width := getIntParam(params, "width", 32)
			return NewMux2(name, width), nil
		}).
		Build())

	// MUX4 - 4:1 Multiplexer
	Registry.Register(component.NewDescriptor("MUX4").
		DisplayName("4:1 MUX").
		Description("4-to-1 Multiplexer - selects between four inputs").
		Category(CategoryMux).
		Version("1.0.0").
		Input("A", 32, "Input 0").
		Input("B", 32, "Input 1").
		Input("C", 32, "Input 2").
		Input("D", 32, "Input 3").
		Input("SEL", 2, "Select signal: 0-3").
		Output("OUT", 32, "Selected output").
		Param("width", "int", 32, "Data width in bits").
		Factory(func(name string, params map[string]interface{}) (component.Component, error) {
			width := getIntParam(params, "width", 32)
			return NewMux4(name, width), nil
		}).
		Build())
}

// =============================================================================
// 2:1 Multiplexer
// =============================================================================

// Mux2 is a 2-to-1 multiplexer
type Mux2 struct {
	*component.BaseComponent

	inputA *component.StandardPort
	inputB *component.StandardPort
	sel    *component.Pin
	output *component.StandardPort

	width int
}

// NewMux2 creates a new 2:1 multiplexer with the specified data width
func NewMux2(name string, width int) *Mux2 {
	mux := &Mux2{
		BaseComponent: component.NewBaseComponent(name, "MUX2"),
		width:         width,
	}

	mux.inputA = component.NewInputPort("A", width)
	mux.AddInput(mux.inputA)

	mux.inputB = component.NewInputPort("B", width)
	mux.AddInput(mux.inputB)

	mux.sel = component.NewInputPin("SEL")
	mux.AddInput(mux.sel)

	mux.output = component.NewOutputPort("OUT", width)
	mux.AddOutput(mux.output)

	return mux
}

// InputA returns the first input port
func (m *Mux2) InputA() *component.StandardPort { return m.inputA }

// InputB returns the second input port
func (m *Mux2) InputB() *component.StandardPort { return m.inputB }

// Select returns the select pin
func (m *Mux2) Select() *component.Pin { return m.sel }

// Output returns the output port
func (m *Mux2) Output() *component.StandardPort { return m.output }

// Width returns the data width
func (m *Mux2) Width() int { return m.width }

// Compute performs the multiplexer selection (combinational logic)
func (m *Mux2) Compute() error {
	if !m.IsEnabled() {
		return nil
	}

	if m.sel.IsHigh() {
		m.output.SetValue(m.inputB.GetValue())
	} else {
		m.output.SetValue(m.inputA.GetValue())
	}

	return nil
}

// Reset resets the multiplexer output
func (m *Mux2) Reset() {
	m.output.Reset()
}

// =============================================================================
// 4:1 Multiplexer
// =============================================================================

// Mux4 is a 4-to-1 multiplexer
type Mux4 struct {
	*component.BaseComponent

	inputs [4]*component.StandardPort
	sel    *component.StandardPort // 2-bit select
	output *component.StandardPort

	width int
}

// NewMux4 creates a new 4:1 multiplexer with the specified data width
func NewMux4(name string, width int) *Mux4 {
	mux := &Mux4{
		BaseComponent: component.NewBaseComponent(name, "MUX4"),
		width:         width,
	}

	names := []string{"A", "B", "C", "D"}
	for i := 0; i < 4; i++ {
		mux.inputs[i] = component.NewInputPort(names[i], width)
		mux.AddInput(mux.inputs[i])
	}

	mux.sel = component.NewInputPort("SEL", 2)
	mux.AddInput(mux.sel)

	mux.output = component.NewOutputPort("OUT", width)
	mux.AddOutput(mux.output)

	return mux
}

// Input returns the input port at the specified index (0-3)
func (m *Mux4) Input(idx int) *component.StandardPort {
	if idx < 0 || idx >= 4 {
		return nil
	}
	return m.inputs[idx]
}

// InputA returns input 0
func (m *Mux4) InputA() *component.StandardPort { return m.inputs[0] }

// InputB returns input 1
func (m *Mux4) InputB() *component.StandardPort { return m.inputs[1] }

// InputC returns input 2
func (m *Mux4) InputC() *component.StandardPort { return m.inputs[2] }

// InputD returns input 3
func (m *Mux4) InputD() *component.StandardPort { return m.inputs[3] }

// Select returns the 2-bit select port
func (m *Mux4) Select() *component.StandardPort { return m.sel }

// Output returns the output port
func (m *Mux4) Output() *component.StandardPort { return m.output }

// Width returns the data width
func (m *Mux4) Width() int { return m.width }

// Compute performs the multiplexer selection (combinational logic)
func (m *Mux4) Compute() error {
	if !m.IsEnabled() {
		return nil
	}

	idx := int(m.sel.GetValue() & 0x3)
	m.output.SetValue(m.inputs[idx].GetValue())

	return nil
}

// Reset resets the multiplexer output
func (m *Mux4) Reset() {
	m.output.Reset()
}
