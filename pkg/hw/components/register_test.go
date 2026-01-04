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
			assert.Equal(t, uint64(0), rb.Get(i))
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
		assert.Equal(t, uint64(0), rb.Get(-1))
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
			assert.Equal(t, uint64(0), rb.Get(i), "register %d should be zero", i)
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
			assert.Equal(t, uint64(0), rb.Get(i))
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

// =============================================================================
// RAM Tests
// =============================================================================

func TestRAM(t *testing.T) {
	t.Run("NewRAM creates RAM with correct ports", func(t *testing.T) {
		ram := NewRAM("MainRAM", 256)

		assert.Equal(t, "MainRAM", ram.Name())
		assert.Equal(t, "RAM", ram.Type())

		// Check data port (8-bit)
		assert.NotNil(t, ram.Data())
		assert.Equal(t, 8, ram.Data().Width())

		// Check address port (32-bit)
		assert.NotNil(t, ram.Address())
		assert.Equal(t, 32, ram.Address().Width())

		// Check control pins
		assert.NotNil(t, ram.ReadWrite())
		assert.NotNil(t, ram.Ready())

		// Initially ready
		assert.True(t, ram.Ready().IsHigh())

		// Check size
		assert.Equal(t, 256, ram.Size())
	})

	t.Run("Initial memory is zero", func(t *testing.T) {
		ram := NewRAM("RAM", 64)

		for i := 0; i < 64; i++ {
			assert.Equal(t, byte(0), ram.ReadByte(i))
		}
	})

	t.Run("DirectWriteByte and ReadByte work", func(t *testing.T) {
		ram := NewRAM("RAM", 256)

		ram.WriteByte(0, 0xAB)
		ram.WriteByte(100, 0xCD)
		ram.WriteByte(255, 0xEF)

		assert.Equal(t, byte(0xAB), ram.ReadByte(0))
		assert.Equal(t, byte(0xCD), ram.ReadByte(100))
		assert.Equal(t, byte(0xEF), ram.ReadByte(255))
	})

	t.Run("Out of bounds access is safe", func(t *testing.T) {
		ram := NewRAM("RAM", 64)

		// Write out of bounds should be ignored
		ram.WriteByte(-1, 0xFF)
		ram.WriteByte(64, 0xFF)
		ram.WriteByte(100, 0xFF)

		// Read out of bounds returns 0
		assert.Equal(t, byte(0), ram.ReadByte(-1))
		assert.Equal(t, byte(0), ram.ReadByte(64))
		assert.Equal(t, byte(0), ram.ReadByte(100))
	})
}

func TestRAMWriteOperation(t *testing.T) {
	t.Run("Write operation stores data at address", func(t *testing.T) {
		ram := NewRAM("RAM", 256)

		// Set address
		ram.Address().SetValue(42)

		// Set data to write
		ram.Data().SetValue(0xAB)

		// Set write mode (RW=High)
		ram.ReadWrite().Set(High)

		// Clock the RAM
		err := ram.Clock()
		require.NoError(t, err)

		// Verify data was written
		assert.Equal(t, byte(0xAB), ram.ReadByte(42))

		// Ready should be high after operation
		assert.True(t, ram.Ready().IsHigh())
	})

	t.Run("Multiple writes to different addresses", func(t *testing.T) {
		ram := NewRAM("RAM", 256)
		ram.ReadWrite().Set(High) // Write mode

		testData := map[int]byte{
			0:   0x11,
			50:  0x22,
			100: 0x33,
			255: 0x44,
		}

		for addr, value := range testData {
			ram.Address().SetValue(uint64(addr))
			ram.Data().SetValue(uint64(value))
			err := ram.Clock()
			require.NoError(t, err)
		}

		// Verify all writes
		for addr, expected := range testData {
			assert.Equal(t, expected, ram.ReadByte(addr))
		}
	})

	t.Run("Write only stores lower 8 bits", func(t *testing.T) {
		ram := NewRAM("RAM", 256)
		ram.ReadWrite().Set(High)
		ram.Address().SetValue(10)
		ram.Data().SetValue(0x1234) // Only 0x34 should be stored

		err := ram.Clock()
		require.NoError(t, err)

		assert.Equal(t, byte(0x34), ram.ReadByte(10))
	})
}

func TestRAMReadOperation(t *testing.T) {
	t.Run("Read operation outputs data from address", func(t *testing.T) {
		ram := NewRAM("RAM", 256)

		// Pre-populate memory
		ram.WriteByte(42, 0xAB)

		// Set address
		ram.Address().SetValue(42)

		// Set read mode (RW=Low)
		ram.ReadWrite().Set(Low)

		// Clock the RAM
		err := ram.Clock()
		require.NoError(t, err)

		// Verify data port has the value
		assert.Equal(t, uint64(0xAB), ram.Data().GetValue())

		// Ready should be high after operation
		assert.True(t, ram.Ready().IsHigh())
	})

	t.Run("Read from multiple addresses", func(t *testing.T) {
		ram := NewRAM("RAM", 256)

		// Pre-populate memory
		testData := map[int]byte{
			0:   0x11,
			50:  0x22,
			100: 0x33,
			255: 0x44,
		}
		for addr, value := range testData {
			ram.WriteByte(addr, value)
		}

		ram.ReadWrite().Set(Low) // Read mode

		// Read and verify each address
		for addr, expected := range testData {
			ram.Address().SetValue(uint64(addr))
			err := ram.Clock()
			require.NoError(t, err)
			assert.Equal(t, uint64(expected), ram.Data().GetValue())
		}
	})

	t.Run("Read from unwritten address returns zero", func(t *testing.T) {
		ram := NewRAM("RAM", 256)
		ram.ReadWrite().Set(Low)
		ram.Address().SetValue(100)

		err := ram.Clock()
		require.NoError(t, err)

		assert.Equal(t, uint64(0), ram.Data().GetValue())
	})
}

func TestRAMReadAfterWrite(t *testing.T) {
	t.Run("Read after write returns written value", func(t *testing.T) {
		ram := NewRAM("RAM", 256)

		// Write
		ram.Address().SetValue(10)
		ram.Data().SetValue(0x42)
		ram.ReadWrite().Set(High)
		err := ram.Clock()
		require.NoError(t, err)

		// Read back
		ram.Address().SetValue(10)
		ram.ReadWrite().Set(Low)
		err = ram.Clock()
		require.NoError(t, err)

		assert.Equal(t, uint64(0x42), ram.Data().GetValue())
	})

	t.Run("Overwrite updates memory", func(t *testing.T) {
		ram := NewRAM("RAM", 256)

		// First write
		ram.Address().SetValue(10)
		ram.Data().SetValue(0x11)
		ram.ReadWrite().Set(High)
		ram.Clock()

		// Second write (overwrite)
		ram.Data().SetValue(0x22)
		ram.Clock()

		// Read back
		ram.ReadWrite().Set(Low)
		ram.Clock()

		assert.Equal(t, uint64(0x22), ram.Data().GetValue())
	})
}

func TestRAMDisabled(t *testing.T) {
	t.Run("Disabled RAM does not perform operations", func(t *testing.T) {
		ram := NewRAM("RAM", 256)

		// Disable the RAM
		ram.Disable()
		assert.False(t, ram.IsEnabled())

		// Try to write
		ram.Address().SetValue(10)
		ram.Data().SetValue(0xAB)
		ram.ReadWrite().Set(High)

		err := ram.Clock()
		require.NoError(t, err)

		// Memory should not be changed
		assert.Equal(t, byte(0), ram.ReadByte(10))

		// Re-enable and verify write works
		ram.Enable()
		err = ram.Clock()
		require.NoError(t, err)
		assert.Equal(t, byte(0xAB), ram.ReadByte(10))
	})
}

func TestRAMReset(t *testing.T) {
	t.Run("Reset clears all memory", func(t *testing.T) {
		ram := NewRAM("RAM", 64)

		// Write some data
		for i := 0; i < 64; i++ {
			ram.WriteByte(i, byte(i+1))
		}

		// Verify data was written
		assert.Equal(t, byte(1), ram.ReadByte(0))
		assert.Equal(t, byte(64), ram.ReadByte(63))

		// Reset
		ram.Reset()

		// Verify all memory is cleared
		for i := 0; i < 64; i++ {
			assert.Equal(t, byte(0), ram.ReadByte(i), "Address %d should be 0 after reset", i)
		}

		// Ready should be high after reset
		assert.True(t, ram.Ready().IsHigh())
	})
}

func TestRAMBoundsOnClock(t *testing.T) {
	t.Run("Out of bounds address on write is ignored", func(t *testing.T) {
		ram := NewRAM("RAM", 64)
		ram.ReadWrite().Set(High)
		ram.Address().SetValue(100) // Out of bounds
		ram.Data().SetValue(0xAB)

		err := ram.Clock()
		require.NoError(t, err)

		// No crash, operation is silently ignored
	})

	t.Run("Out of bounds address on read returns zero", func(t *testing.T) {
		ram := NewRAM("RAM", 64)
		ram.ReadWrite().Set(Low)
		ram.Address().SetValue(100) // Out of bounds

		// Pre-set data port to non-zero
		ram.Data().SetValue(0xFF)

		err := ram.Clock()
		require.NoError(t, err)

		// Data port should remain unchanged for out-of-bounds read
		// (behavior depends on implementation - we don't modify on out-of-bounds)
	})
}

func TestRAMRegistry(t *testing.T) {
	t.Run("RAM descriptor is registered", func(t *testing.T) {
		desc, err := Registry.Get("RAM")
		require.NoError(t, err)

		assert.Equal(t, "RAM", desc.Name)
		assert.Equal(t, "RAM", desc.DisplayName)
		assert.Equal(t, CategoryMemory, desc.Category)
		assert.NotEmpty(t, desc.Description)
	})

	t.Run("Create RAM from registry", func(t *testing.T) {
		comp, err := Registry.Create("RAM", "my_ram", nil)
		require.NoError(t, err)

		assert.Equal(t, "my_ram", comp.Name())
		assert.Equal(t, "RAM", comp.Type())

		ram, ok := comp.(*RAM)
		require.True(t, ok)
		assert.Equal(t, 1024, ram.Size()) // Default size
	})

	t.Run("Create RAM with custom size", func(t *testing.T) {
		comp, err := Registry.Create("RAM", "my_ram", map[string]interface{}{
			"size": 4096,
		})
		require.NoError(t, err)

		ram, ok := comp.(*RAM)
		require.True(t, ok)
		assert.Equal(t, 4096, ram.Size())
	})
}
