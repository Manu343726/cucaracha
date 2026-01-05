package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Register Bank Tests
// =============================================================================

func TestRegisterBank(t *testing.T) {
	t.Run("NewRegisterBank creates bank with correct ports", func(t *testing.T) {
		rb := NewRegisterBank("RB0", 16, 32)

		assert.Equal(t, "RB0", rb.Name())
		assert.Equal(t, "REGISTER_BANK", rb.Type())

		// Check data port
		assert.NotNil(t, rb.Data())
		assert.Equal(t, 32, rb.Data().Width())

		// Check address port
		assert.NotNil(t, rb.Address())
		assert.Equal(t, 32, rb.Address().Width())

		// Check control pins
		assert.NotNil(t, rb.ReadWrite())
		assert.NotNil(t, rb.Ready())

		// Check count and width
		assert.Equal(t, 16, rb.Count())
		assert.Equal(t, 32, rb.Width())

		// Initially ready
		assert.True(t, rb.Ready().IsHigh())
	})

	t.Run("Initial values are zero", func(t *testing.T) {
		rb := NewRegisterBank("RB0", 8, 32)
		for i := 0; i < 8; i++ {
			assert.Equal(t, uint64(0), rb.Get(uint32(i)))
		}
	})

	t.Run("Set and Get directly access registers", func(t *testing.T) {
		rb := NewRegisterBank("RB0", 8, 32)
		rb.Set(3, 0x12345678)
		assert.Equal(t, uint64(0x12345678), rb.Get(3))
	})

	t.Run("Set masks to register width", func(t *testing.T) {
		rb := NewRegisterBank("RB0", 4, 8)
		rb.Set(0, 0x1234)
		assert.Equal(t, uint64(0x34), rb.Get(0), "should be masked to 8 bits")
	})

	t.Run("Get with invalid index returns zero", func(t *testing.T) {
		rb := NewRegisterBank("RB0", 4, 32)
		assert.Equal(t, uint64(0), rb.Get(100))
	})
}

func TestRegisterBankWrite(t *testing.T) {
	t.Run("Write to addressed register", func(t *testing.T) {
		rb := NewRegisterBank("RB0", 8, 32)

		rb.ReadWrite().Set(High) // Write mode
		rb.Address().SetValue(3)
		rb.Data().SetValue(0xDEADBEEF)

		rb.Clock()

		assert.True(t, rb.Ready().IsHigh())
		assert.Equal(t, uint64(0xDEADBEEF), rb.Get(3))
		// Other registers should be untouched
		assert.Equal(t, uint64(0), rb.Get(0))
		assert.Equal(t, uint64(0), rb.Get(2))
		assert.Equal(t, uint64(0), rb.Get(4))
	})

	t.Run("Write to multiple registers", func(t *testing.T) {
		rb := NewRegisterBank("RB0", 8, 32)
		rb.ReadWrite().Set(High)

		// Write to register 0
		rb.Address().SetValue(0)
		rb.Data().SetValue(0x11111111)
		rb.Clock()

		// Write to register 5
		rb.Address().SetValue(5)
		rb.Data().SetValue(0x55555555)
		rb.Clock()

		assert.Equal(t, uint64(0x11111111), rb.Get(0))
		assert.Equal(t, uint64(0x55555555), rb.Get(5))
	})
}

func TestRegisterBankRead(t *testing.T) {
	t.Run("Read from addressed register", func(t *testing.T) {
		rb := NewRegisterBank("RB0", 8, 32)
		rb.Set(4, 0xCAFEBABE)

		rb.ReadWrite().Set(Low) // Read mode
		rb.Address().SetValue(4)

		rb.Clock()

		assert.True(t, rb.Ready().IsHigh())
		assert.Equal(t, uint64(0xCAFEBABE), rb.Data().GetValue())
	})

	t.Run("Read does not modify registers", func(t *testing.T) {
		rb := NewRegisterBank("RB0", 8, 32)
		rb.Set(2, 0x12345678)

		rb.ReadWrite().Set(Low)
		rb.Address().SetValue(2)
		rb.Clock()

		assert.Equal(t, uint64(0x12345678), rb.Get(2))
	})

	t.Run("Read from different registers", func(t *testing.T) {
		rb := NewRegisterBank("RB0", 4, 32)
		rb.Set(0, 0xAAAAAAAA)
		rb.Set(1, 0xBBBBBBBB)
		rb.Set(2, 0xCCCCCCCC)
		rb.Set(3, 0xDDDDDDDD)

		rb.ReadWrite().Set(Low)

		for i, expected := range []uint64{0xAAAAAAAA, 0xBBBBBBBB, 0xCCCCCCCC, 0xDDDDDDDD} {
			rb.Address().SetValue(uint64(i))
			rb.Clock()
			assert.Equal(t, expected, rb.Data().GetValue(), "register %d", i)
		}
	})
}

func TestRegisterBankReset(t *testing.T) {
	t.Run("Reset clears all registers", func(t *testing.T) {
		rb := NewRegisterBank("RB0", 4, 32)
		rb.Set(0, 0x11111111)
		rb.Set(1, 0x22222222)
		rb.Set(2, 0x33333333)
		rb.Set(3, 0x44444444)

		rb.Reset()

		for i := 0; i < 4; i++ {
			assert.Equal(t, uint64(0), rb.Get(uint32(i)), "register %d should be zero", i)
		}
		assert.True(t, rb.Ready().IsHigh())
	})
}

func TestRegisterBankDisabled(t *testing.T) {
	t.Run("Disabled bank does not perform operations", func(t *testing.T) {
		rb := NewRegisterBank("RB0", 4, 32)
		rb.Set(0, 0x11111111)

		rb.Disable()

		rb.ReadWrite().Set(High)
		rb.Address().SetValue(0)
		rb.Data().SetValue(0x22222222)
		rb.Clock()

		assert.Equal(t, uint64(0x11111111), rb.Get(0))
	})
}

func TestRegisterBankOutOfBounds(t *testing.T) {
	t.Run("Write to out of bounds address is ignored", func(t *testing.T) {
		rb := NewRegisterBank("RB0", 4, 32)

		rb.ReadWrite().Set(High)
		rb.Address().SetValue(100) // Out of bounds
		rb.Data().SetValue(0xFFFFFFFF)
		rb.Clock()

		// All registers should still be zero
		for i := 0; i < 4; i++ {
			assert.Equal(t, uint64(0), rb.Get(uint32(i)))
		}
	})
}

func TestRegisterBankRegistry(t *testing.T) {
	t.Run("REGISTER_BANK is registered in Registry", func(t *testing.T) {
		desc, err := Registry.Get("REGISTER_BANK")
		require.NoError(t, err)

		assert.Equal(t, "REGISTER_BANK", desc.Name)
		assert.Equal(t, CategoryMemory, desc.Category)
		assert.NotEmpty(t, desc.Description)
	})

	t.Run("Create register bank from registry", func(t *testing.T) {
		comp, err := Registry.Create("REGISTER_BANK", "my_rb", nil)
		require.NoError(t, err)

		assert.Equal(t, "my_rb", comp.Name())
		assert.Equal(t, "REGISTER_BANK", comp.Type())

		rb, ok := comp.(*RegisterBank)
		require.True(t, ok)
		assert.Equal(t, 16, rb.Count()) // Default count
		assert.Equal(t, 32, rb.Width()) // Default width
	})

	t.Run("Create register bank with custom params", func(t *testing.T) {
		comp, err := Registry.Create("REGISTER_BANK", "my_rb", map[string]interface{}{
			"count": 8,
			"width": 16,
		})
		require.NoError(t, err)

		rb, ok := comp.(*RegisterBank)
		require.True(t, ok)
		assert.Equal(t, 8, rb.Count())
		assert.Equal(t, 16, rb.Width())
	})
}
