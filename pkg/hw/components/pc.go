package components

import (
	"github.com/Manu343726/cucaracha/pkg/hw/component"
)

func init() {
	registerPCComponents()
}

const CategoryControl = "control"

func registerPCComponents() {
	// Program Counter
	Registry.Register(component.NewDescriptor("PC").
		DisplayName("Program Counter").
		Description("Program Counter - tracks instruction address with increment and load").
		Category(CategoryControl).
		Version("1.0.0").
		Input("LOAD", 1, "Load enable: High=load new value").
		Input("IN", 32, "Value to load when LOAD is high").
		Input("INC", 1, "Increment enable: High=increment by step").
		Input("RESET", 1, "Reset to initial value").
		Output("OUT", 32, "Current PC value").
		Param("step", "int", 4, "Increment step size (default 4 for 32-bit instructions)").
		Param("initial", "int", 0, "Initial/reset value").
		Factory(func(name string, params map[string]interface{}) (component.Component, error) {
			step := getIntParam(params, "step", 4)
			initial := getIntParam(params, "initial", 0)
			return NewProgramCounter(name, uint32(step), uint32(initial)), nil
		}).
		Build())
}

// =============================================================================
// Program Counter
// =============================================================================

// ProgramCounter is a register that tracks the current instruction address
type ProgramCounter struct {
	*component.BaseComponent

	// Control inputs
	loadEnable *component.Pin          // Load new value
	loadValue  *component.StandardPort // Value to load
	incEnable  *component.Pin          // Increment enable
	resetPin   *component.Pin          // Reset to initial

	// Output
	output *component.StandardPort

	// Internal state
	value   uint32
	step    uint32 // Increment amount (typically 4)
	initial uint32 // Reset value
}

// NewProgramCounter creates a new program counter
func NewProgramCounter(name string, step, initial uint32) *ProgramCounter {
	pc := &ProgramCounter{
		BaseComponent: component.NewBaseComponent(name, "PC"),
		value:         initial,
		step:          step,
		initial:       initial,
	}

	pc.loadEnable = component.NewInputPin("LOAD")
	pc.AddInput(pc.loadEnable)

	pc.loadValue = component.NewInputPort("IN", 32)
	pc.AddInput(pc.loadValue)

	pc.incEnable = component.NewInputPin("INC")
	pc.AddInput(pc.incEnable)

	pc.resetPin = component.NewInputPin("RESET")
	pc.AddInput(pc.resetPin)

	pc.output = component.NewOutputPort("OUT", 32)
	pc.output.SetValue(uint64(initial))
	pc.AddOutput(pc.output)

	return pc
}

// LoadEnable returns the load enable pin
func (pc *ProgramCounter) LoadEnable() *component.Pin { return pc.loadEnable }

// LoadValue returns the load value input port
func (pc *ProgramCounter) LoadValue() *component.StandardPort { return pc.loadValue }

// IncrementEnable returns the increment enable pin
func (pc *ProgramCounter) IncrementEnable() *component.Pin { return pc.incEnable }

// ResetPin returns the reset pin
func (pc *ProgramCounter) ResetPin() *component.Pin { return pc.resetPin }

// Output returns the output port
func (pc *ProgramCounter) Output() *component.StandardPort { return pc.output }

// Value returns the current PC value
func (pc *ProgramCounter) Value() uint32 { return pc.value }

// Step returns the increment step size
func (pc *ProgramCounter) Step() uint32 { return pc.step }

// SetValue directly sets the PC value (for initialization)
func (pc *ProgramCounter) SetValue(v uint32) {
	pc.value = v
	pc.output.SetValue(uint64(v))
}

// Clock updates the PC on clock edge
func (pc *ProgramCounter) Clock() error {
	if !pc.IsEnabled() {
		return nil
	}

	// Reset has highest priority
	if pc.resetPin.IsHigh() {
		pc.value = pc.initial
	} else if pc.loadEnable.IsHigh() {
		// Load has priority over increment
		pc.value = uint32(pc.loadValue.GetValue())
	} else if pc.incEnable.IsHigh() {
		// Increment
		pc.value += pc.step
	}

	pc.output.SetValue(uint64(pc.value))
	return nil
}

// Reset resets the PC to initial value
func (pc *ProgramCounter) Reset() {
	pc.value = pc.initial
	pc.output.SetValue(uint64(pc.initial))
}

// Increment increments the PC by step (convenience method)
func (pc *ProgramCounter) Increment() {
	pc.value += pc.step
	pc.output.SetValue(uint64(pc.value))
}

// Load loads a new value into the PC (convenience method)
func (pc *ProgramCounter) Load(v uint32) {
	pc.value = v
	pc.output.SetValue(uint64(v))
}
