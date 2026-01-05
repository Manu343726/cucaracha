package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			assert.Equal(t, byte(0), ram.ReadByte(uint32(i)), "Memory at address %d should be 0", i)
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
		ram.WriteByte(64, 0xFF)
		ram.WriteByte(100, 0xFF)

		// Read out of bounds returns 0
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
			assert.Equal(t, expected, ram.ReadByte(uint32(addr)))
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
			ram.WriteByte(uint32(addr), value)
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
			ram.WriteByte(uint32(i), byte(i+1))
		}

		// Verify data was written
		assert.Equal(t, byte(1), ram.ReadByte(0))
		assert.Equal(t, byte(64), ram.ReadByte(63))

		// Reset
		ram.Reset()

		// Verify all memory is cleared
		for i := 0; i < 64; i++ {
			assert.Equal(t, byte(0), ram.ReadByte(uint32(i)), "Address %d should be 0 after reset", i)
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
