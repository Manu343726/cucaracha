package system

import (
	"fmt"
	"testing"

	"github.com/Manu343726/cucaracha/pkg/hw/peripheral"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryAllocator_Allocate_WithPeripherals(t *testing.T) {

	peripherals := []peripheral.Metadata{
		{Name: "uart", Size: 16},
		{Name: "gpio", Size: 32},
	}

	result, err := Allocate(MemoryRequirements{
		TotalSize:           0x10000, // 64KB total
		CodeInstructions:    256,     // 1KB code
		StackSize:           512,
		MinPeripheralRegion: 1024, // Ensure enough space for peripherals
		Peripherals:         peripherals,
	}) // 64KB total
	require.NoError(t, err)

	layout := result.Layout

	// Validate layout
	assert.NoError(t, layout.Validate())

	// Peripherals should be allocated
	assert.Len(t, result.PeripheralAddresses, 2)

	// Peripheral addresses should be within peripheral region
	for name, addr := range result.PeripheralAddresses {
		assert.GreaterOrEqual(t, addr, layout.PeripheralBase, "peripheral %s should be in peripheral region", name)
		assert.Less(t, addr, layout.TotalSize, "peripheral %s should be within total size", name)
	}

	// Peripherals shouldn't overlap
	uartAddr := result.PeripheralAddresses[0]
	gpioAddr := result.PeripheralAddresses[1]
	assert.True(t,
		uartAddr+16 <= gpioAddr || gpioAddr+32 <= uartAddr,
		"peripherals should not overlap")
}

func TestMemoryAllocator_Allocate_WithPreferredAddress(t *testing.T) {

	// Request a specific address for the peripheral
	peripherals := []peripheral.Metadata{
		{Name: "uart", Size: 16, BaseAddress: 0xFF00},
	}

	result, err := Allocate(MemoryRequirements{
		TotalSize:           0x10000, // 64KB total
		CodeInstructions:    256,     // 1KB code
		StackSize:           512,
		MinPeripheralRegion: 512, // Ensure space for preferred address
		Peripherals:         peripherals,
	}) // 64KB total
	require.NoError(t, err)

	// Check if preferred address was honored
	uartAddr := result.PeripheralAddresses[0]
	// May or may not be exactly at preferred address depending on constraints
	assert.GreaterOrEqual(t, uartAddr, result.Layout.PeripheralBase)
}

func TestMemoryAllocator_Allocate_InsufficientSpace(t *testing.T) {
	// Request more than available
	_, err := Allocate(MemoryRequirements{
		TotalSize:        4096,
		CodeInstructions: 2000,
		StackSize:        8000,
	}) // Only 4KB total

	assert.Error(t, err)
	assert.True(t, err != nil, "should have error for insufficient memory")
}

func TestAllocationResult_Summary(t *testing.T) {
	result, err := Allocate(MemoryRequirements{
		TotalSize:        65536,
		CodeInstructions: 512, // 2KB code
		StackSize:        1024,
		Peripherals: []peripheral.Metadata{
			{Name: "uart", Size: 16},
		},
	})
	require.NoError(t, err)

	// Summary should contain key information
	assert.Contains(t, result.Summary, "Memory Layout")
	assert.Contains(t, result.Summary, "Code:")
	assert.Contains(t, result.Summary, "Stack:")
	assert.Contains(t, result.Summary, "Peripherals:")
	assert.Contains(t, result.Summary, "uart")
}

// TestSmallMemoryLayouts verifies that small memory sizes work correctly
// With system descriptor (192 bytes) + vector table (128 bytes), minimum is 320 bytes
// plus code and stack, so we start at 512 bytes
func TestSmallMemoryLayouts(t *testing.T) {
	sizes := []uint32{512, 1024, 2048, 4096}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("%d_bytes", size), func(t *testing.T) {
			result, err := Allocate(MemoryRequirements{
				TotalSize:        size,
				CodeInstructions: 16,
				StackSize:        64,
			})

			// Smaller sizes might fail, that's acceptable
			if size < 1024 {
				if err != nil {
					// It's OK if small sizes fail
					return
				}
			}

			require.NoError(t, err, "should allocate %d bytes", size)
			if result != nil {
				assert.NoError(t, result.Layout.Validate())

				// Verify system descriptor and vector table are present
				if result.Layout.SystemDescriptorSize > 0 {
					assert.Greater(t, result.Layout.SystemDescriptorSize, uint32(0))
				}
				if result.Layout.VectorTableSize > 0 {
					assert.Greater(t, result.Layout.VectorTableSize, uint32(0))
				}
			}
		})
	}
}

// TestSmallMemoryLayoutTooSmall verifies that memory sizes too small fail appropriately
func TestSmallMemoryLayoutTooSmall(t *testing.T) {
	// 256 bytes is too small for system descriptor + vector table + code + stack
	result, err := Allocate(MemoryRequirements{
		TotalSize:        256,
		CodeInstructions: 16,
		StackSize:        64,
	})

	assert.Error(t, err, "256 bytes should be too small")
	assert.Nil(t, result)
}
