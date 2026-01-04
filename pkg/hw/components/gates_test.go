package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// NOT Gate Tests
// =============================================================================

func TestNOTGate(t *testing.T) {
	tests := []struct {
		input    BitValue
		expected BitValue
	}{
		{Low, High},
		{High, Low},
	}

	for _, tt := range tests {
		gate := NewNotGate("test_not")

		gate.Input(0).Set(tt.input)
		gate.Compute()

		assert.Equal(t, tt.expected, gate.Output().Get(), "NOT(%v) should be %v", tt.input, tt.expected)
	}
}

// =============================================================================
// AND Gate Tests
// =============================================================================

func TestANDGate(t *testing.T) {
	tests := []struct {
		a, b     BitValue
		expected BitValue
	}{
		{Low, Low, Low},
		{Low, High, Low},
		{High, Low, Low},
		{High, High, High},
	}

	for _, tt := range tests {
		gate := NewAndGate("test_and", 2)

		gate.Input(0).Set(tt.a)
		gate.Input(1).Set(tt.b)
		gate.Compute()

		assert.Equal(t, tt.expected, gate.Output().Get(), "%v AND %v should be %v", tt.a, tt.b, tt.expected)
	}
}

// =============================================================================
// OR Gate Tests
// =============================================================================

func TestORGate(t *testing.T) {
	tests := []struct {
		a, b     BitValue
		expected BitValue
	}{
		{Low, Low, Low},
		{Low, High, High},
		{High, Low, High},
		{High, High, High},
	}

	for _, tt := range tests {
		gate := NewOrGate("test_or", 2)

		gate.Input(0).Set(tt.a)
		gate.Input(1).Set(tt.b)
		gate.Compute()

		assert.Equal(t, tt.expected, gate.Output().Get(), "%v OR %v should be %v", tt.a, tt.b, tt.expected)
	}
}

// =============================================================================
// XOR Gate Tests
// =============================================================================

func TestXORGate(t *testing.T) {
	tests := []struct {
		a, b     BitValue
		expected BitValue
	}{
		{Low, Low, Low},
		{Low, High, High},
		{High, Low, High},
		{High, High, Low},
	}

	for _, tt := range tests {
		gate := NewXorGate("test_xor", 2)

		gate.Input(0).Set(tt.a)
		gate.Input(1).Set(tt.b)
		gate.Compute()

		assert.Equal(t, tt.expected, gate.Output().Get(), "%v XOR %v should be %v", tt.a, tt.b, tt.expected)
	}
}

// =============================================================================
// NAND Gate Tests
// =============================================================================

func TestNANDGate(t *testing.T) {
	tests := []struct {
		a, b     BitValue
		expected BitValue
	}{
		{Low, Low, High},
		{Low, High, High},
		{High, Low, High},
		{High, High, Low},
	}

	for _, tt := range tests {
		gate := NewNandGate("test_nand", 2)

		gate.Input(0).Set(tt.a)
		gate.Input(1).Set(tt.b)
		gate.Compute()

		assert.Equal(t, tt.expected, gate.Output().Get(), "%v NAND %v should be %v", tt.a, tt.b, tt.expected)
	}
}

// =============================================================================
// NOR Gate Tests
// =============================================================================

func TestNORGate(t *testing.T) {
	tests := []struct {
		a, b     BitValue
		expected BitValue
	}{
		{Low, Low, High},
		{Low, High, Low},
		{High, Low, Low},
		{High, High, Low},
	}

	for _, tt := range tests {
		gate := NewNorGate("test_nor", 2)

		gate.Input(0).Set(tt.a)
		gate.Input(1).Set(tt.b)
		gate.Compute()

		assert.Equal(t, tt.expected, gate.Output().Get(), "%v NOR %v should be %v", tt.a, tt.b, tt.expected)
	}
}

// =============================================================================
// XNOR Gate Tests
// =============================================================================

func TestXNORGate(t *testing.T) {
	tests := []struct {
		a, b     BitValue
		expected BitValue
	}{
		{Low, Low, High},
		{Low, High, Low},
		{High, Low, Low},
		{High, High, High},
	}

	for _, tt := range tests {
		gate := NewXnorGate("test_xnor", 2)

		gate.Input(0).Set(tt.a)
		gate.Input(1).Set(tt.b)
		gate.Compute()

		assert.Equal(t, tt.expected, gate.Output().Get(), "%v XNOR %v should be %v", tt.a, tt.b, tt.expected)
	}
}

// =============================================================================
// Buffer Tests
// =============================================================================

func TestBuffer(t *testing.T) {
	tests := []struct {
		input    BitValue
		expected BitValue
	}{
		{Low, Low},
		{High, High},
	}

	for _, tt := range tests {
		gate := NewBuffer("test_buffer")

		gate.Input(0).Set(tt.input)
		gate.Compute()

		assert.Equal(t, tt.expected, gate.Output().Get(), "Buffer(%v) should be %v", tt.input, tt.expected)
	}
}

// =============================================================================
// Multi-Input Gate Tests
// =============================================================================

func TestMultiInputANDGate(t *testing.T) {
	gate := NewAndGate("test_and3", 3)

	// All ones
	gate.Input(0).Set(High)
	gate.Input(1).Set(High)
	gate.Input(2).Set(High)
	gate.Compute()
	assert.Equal(t, High, gate.Output().Get(), "1 AND 1 AND 1 should be 1")

	// One zero
	gate.Input(1).Set(Low)
	gate.Compute()
	assert.Equal(t, Low, gate.Output().Get(), "1 AND 0 AND 1 should be 0")
}

func TestMultiInputORGate(t *testing.T) {
	gate := NewOrGate("test_or4", 4)

	// All zeros
	for i := 0; i < 4; i++ {
		gate.Input(i).Set(Low)
	}
	gate.Compute()
	assert.Equal(t, Low, gate.Output().Get(), "0 OR 0 OR 0 OR 0 should be 0")

	// One high
	gate.Input(2).Set(High)
	gate.Compute()
	assert.Equal(t, High, gate.Output().Get(), "should be 1 with one input high")
}

// =============================================================================
// Disabled Gate Tests
// =============================================================================

func TestDisabledGate(t *testing.T) {
	gate := NewAndGate("test_and", 2)

	// Set output to high first
	gate.Output().Set(High)

	// Set inputs to low (AND of 0,0 = 0)
	gate.Input(0).Set(Low)
	gate.Input(1).Set(Low)

	// Disable the gate - Compute should not run
	gate.Disable()
	gate.Compute()

	// Output should remain high since gate was disabled
	assert.Equal(t, High, gate.Output().Get(), "disabled gate should not update output")
}
