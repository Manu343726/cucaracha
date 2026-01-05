package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Register Tests
// =============================================================================

func TestRegister(t *testing.T) {
	t.Run("NewRegister creates register with correct ports", func(t *testing.T) {
		reg := NewRegister("R0", 32)

		assert.Equal(t, "R0", reg.Name())
		assert.Equal(t, "REGISTER", reg.Type())

		// Check data port
		assert.NotNil(t, reg.Data())
		assert.Equal(t, 32, reg.Data().Width())

		// Check control pins
		assert.NotNil(t, reg.ReadWrite())
		assert.NotNil(t, reg.Ready())

		// Initially ready
		assert.True(t, reg.Ready().IsHigh())
	})

	t.Run("Initial value is zero", func(t *testing.T) {
		reg := NewRegister("R0", 32)
		assert.Equal(t, uint64(0), reg.Value())
	})

	t.Run("SetValue directly sets value", func(t *testing.T) {
		reg := NewRegister("R0", 32)
		reg.SetValue(0x12345678)
		assert.Equal(t, uint64(0x12345678), reg.Value())
	})

	t.Run("SetValue masks to register width", func(t *testing.T) {
		reg := NewRegister("R0", 8)
		reg.SetValue(0x1234)
		assert.Equal(t, uint64(0x34), reg.Value(), "should be masked to 8 bits")
	})
}

func TestRegisterWrite(t *testing.T) {
	t.Run("Write operation stores value", func(t *testing.T) {
		reg := NewRegister("R0", 32)

		// Set up write operation
		reg.ReadWrite().Set(High) // Write mode
		reg.Data().SetValue(0xDEADBEEF)

		// Execute on clock
		reg.Clock()
		assert.True(t, reg.Ready().IsHigh(), "should be ready after clock")
		assert.Equal(t, uint64(0xDEADBEEF), reg.Value(), "value should be stored")
	})

	t.Run("Multiple writes", func(t *testing.T) {
		reg := NewRegister("R0", 32)
		reg.ReadWrite().Set(High) // Write mode

		// First write
		reg.Data().SetValue(0x11111111)
		reg.Clock()
		assert.Equal(t, uint64(0x11111111), reg.Value())

		// Second write
		reg.Data().SetValue(0x22222222)
		reg.Clock()
		assert.Equal(t, uint64(0x22222222), reg.Value())
	})
}

func TestRegisterRead(t *testing.T) {
	t.Run("Read operation outputs stored value", func(t *testing.T) {
		reg := NewRegister("R0", 32)
		reg.SetValue(0xCAFEBABE)

		// Set up read operation
		reg.ReadWrite().Set(Low) // Read mode

		// Execute on clock
		reg.Clock()
		assert.True(t, reg.Ready().IsHigh(), "should be ready after clock")
		assert.Equal(t, uint64(0xCAFEBABE), reg.Data().GetValue(), "data port should have stored value")
	})

	t.Run("Read does not modify stored value", func(t *testing.T) {
		reg := NewRegister("R0", 32)
		reg.SetValue(0x12345678)

		reg.ReadWrite().Set(Low) // Read mode
		reg.Clock()

		assert.Equal(t, uint64(0x12345678), reg.Value(), "stored value should not change")
	})
}

func TestRegisterReadWriteSequence(t *testing.T) {
	t.Run("Write then read sequence", func(t *testing.T) {
		reg := NewRegister("R0", 32)

		// Write phase
		reg.ReadWrite().Set(High)
		reg.Data().SetValue(0xABCD1234)
		reg.Clock()

		assert.Equal(t, uint64(0xABCD1234), reg.Value())

		// Read phase
		reg.ReadWrite().Set(Low)
		reg.Data().SetValue(0) // Clear data port
		reg.Clock()

		assert.Equal(t, uint64(0xABCD1234), reg.Data().GetValue())
	})
}

func TestRegisterReset(t *testing.T) {
	t.Run("Reset clears value and sets ready", func(t *testing.T) {
		reg := NewRegister("R0", 32)

		// Store some value
		reg.SetValue(0xFFFFFFFF)

		// Reset
		reg.Reset()

		assert.Equal(t, uint64(0), reg.Value())
		assert.True(t, reg.Ready().IsHigh())
	})
}

func TestRegisterDisabled(t *testing.T) {
	t.Run("Disabled register does not perform operations", func(t *testing.T) {
		reg := NewRegister("R0", 32)
		reg.SetValue(0x11111111)

		reg.Disable()

		// Try to write
		reg.ReadWrite().Set(High)
		reg.Data().SetValue(0x22222222)
		reg.Clock()

		// Value should not change
		assert.Equal(t, uint64(0x11111111), reg.Value())
	})
}

func TestRegisterRegistry(t *testing.T) {
	t.Run("REGISTER is registered in Registry", func(t *testing.T) {
		desc, err := Registry.Get("REGISTER")
		require.NoError(t, err)

		assert.Equal(t, "REGISTER", desc.Name)
		assert.Equal(t, CategoryMemory, desc.Category)
		assert.NotEmpty(t, desc.Description)
	})

	t.Run("Create register from registry", func(t *testing.T) {
		comp, err := Registry.Create("REGISTER", "my_reg", nil)
		require.NoError(t, err)

		assert.Equal(t, "my_reg", comp.Name())
		assert.Equal(t, "REGISTER", comp.Type())
	})

	t.Run("Create register with custom width", func(t *testing.T) {
		comp, err := Registry.Create("REGISTER", "my_reg", map[string]interface{}{
			"width": 16,
		})
		require.NoError(t, err)

		reg, ok := comp.(*Reg32)
		require.True(t, ok)
		assert.Equal(t, 16, reg.Data().Width())
	})
}
