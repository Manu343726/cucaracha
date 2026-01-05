package component

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// BitValue Tests
// =============================================================================

func TestBitValue(t *testing.T) {
	t.Run("Low is false", func(t *testing.T) {
		assert.False(t, bool(Low), "Low should be false")
	})

	t.Run("High is true", func(t *testing.T) {
		assert.True(t, bool(High), "High should be true")
	})

	t.Run("String representation", func(t *testing.T) {
		assert.Equal(t, "0", Low.String())
		assert.Equal(t, "1", High.String())
	})
}

// =============================================================================
// Direction Tests
// =============================================================================

func TestDirection(t *testing.T) {
	t.Run("String representation", func(t *testing.T) {
		tests := []struct {
			dir  Direction
			want string
		}{
			{Input, "IN"},
			{Output, "OUT"},
			{Bidirectional, "INOUT"},
			{Direction(99), "?"},
		}

		for _, tt := range tests {
			assert.Equal(t, tt.want, tt.dir.String())
		}
	})
}

// =============================================================================
// StandardPort Tests
// =============================================================================

func TestStandardPort(t *testing.T) {
	t.Run("NewPort creates port with correct properties", func(t *testing.T) {
		port := NewPort("test", 8)

		assert.Equal(t, "test", port.Name())
		assert.Equal(t, 8, port.Width())
		assert.Equal(t, Bidirectional, port.Direction())
	})

	t.Run("NewInputPort creates input port", func(t *testing.T) {
		port := NewInputPort("input", 4)
		assert.Equal(t, Input, port.Direction())
	})

	t.Run("NewOutputPort creates output port", func(t *testing.T) {
		port := NewOutputPort("output", 4)
		assert.Equal(t, Output, port.Direction())
	})

	t.Run("SetBit and GetBit", func(t *testing.T) {
		port := NewPort("test", 8)

		require.NoError(t, port.SetBit(3, High))

		val, err := port.GetBit(3)
		require.NoError(t, err)
		assert.Equal(t, High, val)

		// Check that other bits are still low
		for i := 0; i < 8; i++ {
			if i == 3 {
				continue
			}
			val, _ := port.GetBit(i)
			assert.Equal(t, Low, val, "bit %d should be Low", i)
		}
	})

	t.Run("SetBit out of range returns error", func(t *testing.T) {
		port := NewPort("test", 8)

		assert.Error(t, port.SetBit(-1, High))
		assert.Error(t, port.SetBit(8, High))
	})

	t.Run("SetValue and GetValue", func(t *testing.T) {
		port := NewPort("test", 8)

		require.NoError(t, port.SetValue(0xA5))
		assert.Equal(t, uint64(0xA5), port.GetValue())
	})

	t.Run("SetValue masks to port width", func(t *testing.T) {
		port := NewPort("test", 4)

		require.NoError(t, port.SetValue(0xFF))
		assert.Equal(t, uint64(0x0F), port.GetValue(), "should be masked to 4 bits")
	})

	t.Run("GetBits returns slice", func(t *testing.T) {
		port := NewPort("test", 8)
		port.SetValue(0b11010110)

		bits, err := port.GetBits(1, 4)
		require.NoError(t, err)

		expected := []BitValue{High, High, Low, High}
		assert.Equal(t, expected, bits)
	})

	t.Run("SetBits sets multiple bits", func(t *testing.T) {
		port := NewPort("test", 8)

		values := []BitValue{High, Low, High, High}
		require.NoError(t, port.SetBits(2, values))

		for i, expected := range values {
			val, _ := port.GetBit(2 + i)
			assert.Equal(t, expected, val, "bit %d", 2+i)
		}
	})

	t.Run("GetBytes and SetBytes", func(t *testing.T) {
		port := NewPort("test", 16)
		port.SetValue(0x1234)

		bytes := port.GetBytes()
		require.Len(t, bytes, 2)
		assert.Equal(t, byte(0x34), bytes[0], "LSB first (little-endian)")
		assert.Equal(t, byte(0x12), bytes[1])

		port.SetBytes([]byte{0xAB, 0xCD})
		assert.Equal(t, uint64(0xCDAB), port.GetValue())
	})

	t.Run("Tristate behavior", func(t *testing.T) {
		port := NewPort("test", 8)
		port.SetValue(0xFF)

		port.SetTristate(true)
		assert.True(t, port.IsTristate())
		assert.Equal(t, uint64(0), port.GetValue(), "tristate should return 0")
		assert.Error(t, port.SetValue(0x55), "SetValue should fail when tristate")

		port.SetTristate(false)
		assert.Equal(t, uint64(0xFF), port.GetValue(), "original value should be preserved")
	})

	t.Run("Reset clears all bits", func(t *testing.T) {
		port := NewPort("test", 8)
		port.SetValue(0xFF)
		port.SetTristate(true)

		port.Reset()

		assert.Equal(t, uint64(0), port.GetValue())
		assert.False(t, port.IsTristate())
	})

	t.Run("WithInitialValue option", func(t *testing.T) {
		port := NewPort("test", 8, WithInitialValue(0x42))
		assert.Equal(t, uint64(0x42), port.GetValue())
	})

	t.Run("WithOnChange callback", func(t *testing.T) {
		var callCount int
		var lastOld, lastNew uint64

		port := NewPort("test", 8, WithOnChange(func(c *StandardPort, old, new uint64) {
			callCount++
			lastOld = old
			lastNew = new
		}))

		port.SetValue(0x10)

		assert.Equal(t, 1, callCount)
		assert.Equal(t, uint64(0), lastOld)
		assert.Equal(t, uint64(0x10), lastNew)

		// Setting same value shouldn't trigger callback
		port.SetValue(0x10)
		assert.Equal(t, 1, callCount, "callback should not be called for same value")
	})
}

// =============================================================================
// Pin Tests
// =============================================================================

func TestPin(t *testing.T) {
	t.Run("NewPin creates 1-bit port", func(t *testing.T) {
		pin := NewPin("CLK")

		assert.Equal(t, "CLK", pin.Name())
		assert.Equal(t, 1, pin.Width())
	})

	t.Run("NewInputPin creates input pin", func(t *testing.T) {
		pin := NewInputPin("IN")
		assert.Equal(t, Input, pin.Direction())
	})

	t.Run("NewOutputPin creates output pin", func(t *testing.T) {
		pin := NewOutputPin("OUT")
		assert.Equal(t, Output, pin.Direction())
	})

	t.Run("Get and Set", func(t *testing.T) {
		pin := NewPin("test")

		assert.Equal(t, Low, pin.Get())

		require.NoError(t, pin.Set(High))
		assert.Equal(t, High, pin.Get())
	})

	t.Run("IsHigh and IsLow", func(t *testing.T) {
		pin := NewPin("test")

		assert.True(t, pin.IsLow())
		assert.False(t, pin.IsHigh())

		pin.Set(High)

		assert.False(t, pin.IsLow())
		assert.True(t, pin.IsHigh())
	})

	t.Run("Toggle", func(t *testing.T) {
		pin := NewPin("test")

		pin.Toggle()
		assert.True(t, pin.IsHigh(), "Toggle from Low should result in High")

		pin.Toggle()
		assert.True(t, pin.IsLow(), "Toggle from High should result in Low")
	})

	t.Run("Pin can be used as Port", func(t *testing.T) {
		pin := NewPin("test")

		var port Port = pin

		require.NoError(t, port.SetValue(1))
		assert.Equal(t, uint64(1), port.GetValue())
		assert.True(t, pin.IsHigh())
	})
}

// =============================================================================
// Component Tests
// =============================================================================

func TestBaseComponent(t *testing.T) {
	t.Run("NewBaseComponent creates component", func(t *testing.T) {
		comp := NewBaseComponent("ALU", "arithmetic")

		assert.Equal(t, "ALU", comp.Name())
		assert.Equal(t, "arithmetic", comp.Type())
		assert.True(t, comp.IsEnabled())
	})

	t.Run("AddInput and GetInput", func(t *testing.T) {
		comp := NewBaseComponent("test", "test")
		port := NewInputPort("A", 8)

		require.NoError(t, comp.AddInput(port))

		got, err := comp.GetInput("A")
		require.NoError(t, err)
		assert.Same(t, port, got)
	})

	t.Run("AddInput duplicate returns error", func(t *testing.T) {
		comp := NewBaseComponent("test", "test")
		port := NewInputPort("A", 8)

		comp.AddInput(port)
		assert.Error(t, comp.AddInput(port))
	})

	t.Run("AddOutput and GetOutput", func(t *testing.T) {
		comp := NewBaseComponent("test", "test")
		port := NewOutputPort("Y", 8)

		require.NoError(t, comp.AddOutput(port))

		got, err := comp.GetOutput("Y")
		require.NoError(t, err)
		assert.Same(t, port, got)
	})

	t.Run("GetPort finds inputs and outputs", func(t *testing.T) {
		comp := NewBaseComponent("test", "test")
		input := NewInputPort("IN", 8)
		output := NewOutputPort("OUT", 8)

		comp.AddInput(input)
		comp.AddOutput(output)

		gotIn, err := comp.GetPort("IN")
		require.NoError(t, err)
		assert.Same(t, input, gotIn)

		gotOut, err := comp.GetPort("OUT")
		require.NoError(t, err)
		assert.Same(t, output, gotOut)

		_, err = comp.GetPort("NOTFOUND")
		assert.Error(t, err)
	})

	t.Run("Inputs and Outputs return slices", func(t *testing.T) {
		comp := NewBaseComponent("test", "test")
		comp.AddInput(NewInputPort("A", 8))
		comp.AddInput(NewInputPort("B", 8))
		comp.AddOutput(NewOutputPort("Y", 8))

		assert.Len(t, comp.Inputs(), 2)
		assert.Len(t, comp.Outputs(), 1)
	})

	t.Run("Enable and Disable", func(t *testing.T) {
		comp := NewBaseComponent("test", "test")

		assert.True(t, comp.IsEnabled())

		comp.Disable()
		assert.False(t, comp.IsEnabled())

		comp.Enable()
		assert.True(t, comp.IsEnabled())
	})

	t.Run("Reset resets all ports", func(t *testing.T) {
		comp := NewBaseComponent("test", "test")
		input := NewInputPort("IN", 8)
		output := NewOutputPort("OUT", 8)

		input.SetValue(0xFF)
		output.SetValue(0xFF)

		comp.AddInput(input)
		comp.AddOutput(output)

		comp.Reset()

		assert.Equal(t, uint64(0), input.GetValue())
		assert.Equal(t, uint64(0), output.GetValue())
	})

	t.Run("Clock calls onClock callback", func(t *testing.T) {
		var called bool
		comp := NewBaseComponent("test", "test", WithClock(func() error {
			called = true
			return nil
		}))

		comp.Clock()

		assert.True(t, called)
	})

	t.Run("Clock does nothing when disabled", func(t *testing.T) {
		var called bool
		comp := NewBaseComponent("test", "test", WithClock(func() error {
			called = true
			return nil
		}))

		comp.Disable()
		comp.Clock()

		assert.False(t, called)
	})

	t.Run("Reset calls onReset callback", func(t *testing.T) {
		var called bool
		comp := NewBaseComponent("test", "test", WithReset(func() {
			called = true
		}))

		comp.Reset()

		assert.True(t, called)
	})
}

// =============================================================================
// Bus Tests
// =============================================================================

func TestBus(t *testing.T) {
	t.Run("NewBus creates bus", func(t *testing.T) {
		src := NewOutputPort("SRC", 8)
		dst := NewInputPort("DST", 8)
		bus := NewBus("B1", src, dst)

		assert.Equal(t, "B1", bus.Name())
		assert.Same(t, src, bus.Source())
		assert.Same(t, dst, bus.Destination())
		assert.True(t, bus.IsEnabled())
	})

	t.Run("Transfer copies value", func(t *testing.T) {
		src := NewOutputPort("SRC", 8)
		dst := NewInputPort("DST", 8)
		bus := NewBus("B1", src, dst)

		src.SetValue(0x42)
		bus.Transfer()

		assert.Equal(t, uint64(0x42), dst.GetValue())
	})

	t.Run("Transfer does nothing when disabled", func(t *testing.T) {
		src := NewOutputPort("SRC", 8)
		dst := NewInputPort("DST", 8)
		bus := NewBus("B1", src, dst)

		src.SetValue(0x42)
		bus.Disable()
		bus.Transfer()

		assert.Equal(t, uint64(0), dst.GetValue())
	})

	t.Run("Transfer does nothing when source is tristate", func(t *testing.T) {
		src := NewOutputPort("SRC", 8)
		dst := NewInputPort("DST", 8)
		bus := NewBus("B1", src, dst)

		dst.SetValue(0xFF)
		src.SetTristate(true)
		bus.Transfer()

		assert.Equal(t, uint64(0xFF), dst.GetValue())
	})

	t.Run("WithTransform applies transformation", func(t *testing.T) {
		src := NewOutputPort("SRC", 8)
		dst := NewInputPort("DST", 8)
		bus := NewBus("B1", src, dst, WithTransform(func(v uint64) uint64 {
			return v ^ 0xFF // Invert
		}))

		src.SetValue(0x00)
		bus.Transfer()

		assert.Equal(t, uint64(0xFF), dst.GetValue())
	})

	t.Run("WithBitMapping maps bits", func(t *testing.T) {
		src := NewOutputPort("SRC", 8)
		dst := NewInputPort("DST", 8)
		bus := NewBus("B1", src, dst, WithBitMapping(map[int]int{
			0: 7,
			1: 6,
		}))

		src.SetValue(0b00000011) // bits 0 and 1 are high
		bus.Transfer()

		assert.Equal(t, uint64(0b11000000), dst.GetValue(), "bits 6 and 7 should be high")
	})
}

// =============================================================================
// Interconnect Tests
// =============================================================================

func TestInterconnect(t *testing.T) {
	t.Run("NewInterconnect creates interconnect", func(t *testing.T) {
		dst := NewPort("DST", 8)
		ic := NewInterconnect("MUX", dst)

		assert.Equal(t, "MUX", ic.Name())
		assert.Equal(t, -1, ic.Selected())
	})

	t.Run("AddSource and Select", func(t *testing.T) {
		dst := NewPort("DST", 8)
		ic := NewInterconnect("MUX", dst)

		src1 := NewPort("SRC1", 8)
		src2 := NewPort("SRC2", 8)

		idx1 := ic.AddSource(src1)
		idx2 := ic.AddSource(src2)

		assert.Equal(t, 0, idx1)
		assert.Equal(t, 1, idx2)

		ic.Select(1)
		assert.Equal(t, 1, ic.Selected())
		assert.Same(t, src2, ic.SelectedSource())
	})

	t.Run("Transfer copies from selected source", func(t *testing.T) {
		dst := NewPort("DST", 8)
		ic := NewInterconnect("MUX", dst)

		src1 := NewPort("SRC1", 8)
		src2 := NewPort("SRC2", 8)

		src1.SetValue(0x11)
		src2.SetValue(0x22)

		ic.AddSource(src1)
		ic.AddSource(src2)

		ic.Select(0)
		ic.Transfer()
		assert.Equal(t, uint64(0x11), dst.GetValue())

		ic.Select(1)
		ic.Transfer()
		assert.Equal(t, uint64(0x22), dst.GetValue())
	})

	t.Run("Transfer with no selection puts dst in tristate", func(t *testing.T) {
		dst := NewPort("DST", 8)
		ic := NewInterconnect("MUX", dst)

		ic.Transfer()

		assert.True(t, dst.IsTristate())
	})

	t.Run("Select out of range returns error", func(t *testing.T) {
		dst := NewPort("DST", 8)
		ic := NewInterconnect("MUX", dst)

		assert.Error(t, ic.Select(5))
	})
}

// =============================================================================
// Circuit Tests
// =============================================================================

func TestCircuit(t *testing.T) {
	t.Run("NewCircuit creates circuit", func(t *testing.T) {
		circuit := NewCircuit("CPU")
		assert.Equal(t, "CPU", circuit.Name())
	})

	t.Run("AddComponent and GetComponent", func(t *testing.T) {
		circuit := NewCircuit("test")
		comp := NewBaseComponent("ALU", "arithmetic")

		require.NoError(t, circuit.AddComponent(comp))

		got, err := circuit.GetComponent("ALU")
		require.NoError(t, err)
		assert.Same(t, comp, got)
	})

	t.Run("AddComponent duplicate returns error", func(t *testing.T) {
		circuit := NewCircuit("test")
		comp := NewBaseComponent("ALU", "arithmetic")

		circuit.AddComponent(comp)
		assert.Error(t, circuit.AddComponent(comp))
	})

	t.Run("Connect creates bus", func(t *testing.T) {
		circuit := NewCircuit("test")
		src := NewOutputPort("SRC", 8)
		dst := NewInputPort("DST", 8)

		bus := circuit.Connect("B1", src, dst)

		require.NotNil(t, bus)
		assert.Len(t, circuit.Connections(), 1)
	})

	t.Run("Propagate transfers all connections", func(t *testing.T) {
		circuit := NewCircuit("test")

		src1 := NewOutputPort("SRC1", 8)
		dst1 := NewInputPort("DST1", 8)
		src2 := NewOutputPort("SRC2", 8)
		dst2 := NewInputPort("DST2", 8)

		src1.SetValue(0x11)
		src2.SetValue(0x22)

		circuit.Connect("B1", src1, dst1)
		circuit.Connect("B2", src2, dst2)

		circuit.Propagate()

		assert.Equal(t, uint64(0x11), dst1.GetValue())
		assert.Equal(t, uint64(0x22), dst2.GetValue())
	})

	t.Run("Clock advances all components", func(t *testing.T) {
		circuit := NewCircuit("test")

		var count int
		comp1 := NewBaseComponent("C1", "test", WithClock(func() error {
			count++
			return nil
		}))
		comp2 := NewBaseComponent("C2", "test", WithClock(func() error {
			count++
			return nil
		}))

		circuit.AddComponent(comp1)
		circuit.AddComponent(comp2)

		circuit.Clock()

		assert.Equal(t, 2, count)
	})

	t.Run("Reset resets all components", func(t *testing.T) {
		circuit := NewCircuit("test")

		comp := NewBaseComponent("C1", "test")
		port := NewOutputPort("OUT", 8)
		port.SetValue(0xFF)
		comp.AddOutput(port)

		circuit.AddComponent(comp)
		circuit.Reset()

		assert.Equal(t, uint64(0), port.GetValue())
	})
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestIntegration_SimpleALU(t *testing.T) {
	circuit := NewCircuit("SimpleALU")

	alu := NewBaseComponent("ALU", "arithmetic")

	inputA := NewInputPort("A", 8)
	inputB := NewInputPort("B", 8)
	output := NewOutputPort("Y", 8)

	alu.onClock = func() error {
		a := inputA.GetValue()
		b := inputB.GetValue()
		return output.SetValue((a + b) & 0xFF)
	}

	alu.AddInput(inputA)
	alu.AddInput(inputB)
	alu.AddOutput(output)

	circuit.AddComponent(alu)

	inputA.SetValue(10)
	inputB.SetValue(20)

	circuit.Clock()

	assert.Equal(t, uint64(30), output.GetValue())
}

func TestIntegration_BusInterconnect(t *testing.T) {
	circuit := NewCircuit("MuxTest")

	regA := NewPort("RegA", 8)
	regB := NewPort("RegB", 8)
	regC := NewPort("RegC", 8)

	aluInput := NewPort("ALU_IN", 8)

	mux := NewInterconnect("InputMux", aluInput)
	mux.AddSource(regA)
	mux.AddSource(regB)
	mux.AddSource(regC)

	circuit.AddConnection(mux)

	regA.SetValue(0x11)
	regB.SetValue(0x22)
	regC.SetValue(0x33)

	testCases := []struct {
		sel  int
		want uint64
	}{
		{0, 0x11},
		{1, 0x22},
		{2, 0x33},
	}

	for _, tc := range testCases {
		mux.Select(tc.sel)
		circuit.Propagate()
		assert.Equal(t, tc.want, aluInput.GetValue(), "sel=%d", tc.sel)
	}
}
