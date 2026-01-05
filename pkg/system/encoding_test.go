package system

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodedSystemDescriptorLayout(t *testing.T) {
	// Test multiple instances of EncodedSystemDescriptorMemoryLayout with different peripheral counts
	for _, peripheralCount := range []int{0, 1, 5, 10} {
		t.Run(fmt.Sprintf("PeripheralCount_%d", peripheralCount), func(t *testing.T) {
			layout := EncodedSystemDescriptorLayout(peripheralCount)

			// Validate layout
			require.NoError(t, layout.Validate())

			// Check that the size is correct
			expectedSize := 4 + peripheralCount*24 // 4 bytes for count + 24 bytes per peripheral
			assert.Equal(t, expectedSize, layout.Size)
		})
	}
}
