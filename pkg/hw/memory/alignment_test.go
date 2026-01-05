package memory

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_AlignSize(t *testing.T) {
	tests := []struct {
		size      uint32
		alignment uint32
		expected  uint32
	}{
		{size: 0, alignment: 4, expected: 0},
		{size: 1, alignment: 4, expected: 4},
		{size: 2, alignment: 4, expected: 4},
		{size: 3, alignment: 4, expected: 4},
		{size: 4, alignment: 4, expected: 4},
		{size: 5, alignment: 4, expected: 8},
		{size: 15, alignment: 8, expected: 16},
		{size: 16, alignment: 8, expected: 16},
		{size: 17, alignment: 8, expected: 24},
		{size: 20, alignment: 0, expected: 20},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("size=%d_alignment=%d", tt.size, tt.alignment), func(t *testing.T) {
			result := AlignSize(tt.size, tt.alignment)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_IsAligned(t *testing.T) {
	tests := []struct {
		addr      uint32
		alignment uint32
		expected  bool
	}{
		{addr: 0, alignment: 4, expected: true},
		{addr: 1, alignment: 4, expected: false},
		{addr: 4, alignment: 4, expected: true},
		{addr: 5, alignment: 4, expected: false},
		{addr: 8, alignment: 8, expected: true},
		{addr: 12, alignment: 8, expected: false},
		{addr: 20, alignment: 0, expected: true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("addr=%d_alignment=%d", tt.addr, tt.alignment), func(t *testing.T) {
			result := IsAligned(tt.addr, tt.alignment)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_NextAlignedAddress(t *testing.T) {
	tests := []struct {
		addr      uint32
		alignment uint32
		expected  uint32
	}{
		{addr: 0, alignment: 4, expected: 0},
		{addr: 1, alignment: 4, expected: 4},
		{addr: 4, alignment: 4, expected: 4},
		{addr: 5, alignment: 4, expected: 8},
		{addr: 8, alignment: 8, expected: 8},
		{addr: 12, alignment: 8, expected: 16},
		{addr: 20, alignment: 0, expected: 20},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("addr=%d_alignment=%d", tt.addr, tt.alignment), func(t *testing.T) {
			result := NextAlignedAddress(tt.addr, tt.alignment)
			assert.Equal(t, tt.expected, result)
		})
	}
}
