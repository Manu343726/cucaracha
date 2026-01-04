package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// ALU Creation Tests
// =============================================================================

func TestALU(t *testing.T) {
	t.Run("NewALU creates ALU with correct ports", func(t *testing.T) {
		alu := NewALU("MainALU")

		assert.Equal(t, "MainALU", alu.Name())
		assert.Equal(t, "ALU", alu.Type())

		// Check input ports
		assert.NotNil(t, alu.InputA())
		assert.Equal(t, 32, alu.InputA().Width())

		assert.NotNil(t, alu.InputB())
		assert.Equal(t, 32, alu.InputB().Width())

		assert.NotNil(t, alu.OpCode())
		assert.Equal(t, 4, alu.OpCode().Width())

		// Check output ports
		assert.NotNil(t, alu.Output())
		assert.Equal(t, 32, alu.Output().Width())

		assert.NotNil(t, alu.Flags())
		assert.Equal(t, 32, alu.Flags().Width())
	})

	t.Run("Initial output is zero", func(t *testing.T) {
		alu := NewALU("ALU")
		assert.Equal(t, uint32(0), alu.Result())
		assert.Equal(t, uint32(0), alu.GetFlags())
	})
}

// =============================================================================
// ALU Arithmetic Operations Tests
// =============================================================================

func TestALUAddition(t *testing.T) {
	t.Run("ADD basic", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_ADD, 5, 3)
		assert.Equal(t, uint32(8), result)
	})

	t.Run("ADD with zero", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_ADD, 42, 0)
		assert.Equal(t, uint32(42), result)
	})

	t.Run("ADD sets zero flag", func(t *testing.T) {
		alu := NewALU("ALU")
		alu.Execute(ALUOp_ADD, 0, 0)
		assert.True(t, alu.IsZero())
		assert.False(t, alu.IsNegative())
	})

	t.Run("ADD sets negative flag", func(t *testing.T) {
		alu := NewALU("ALU")
		alu.Execute(ALUOp_ADD, 0xFFFFFFFF, 0) // -1 in two's complement
		assert.True(t, alu.IsNegative())
	})

	t.Run("ADD sets carry on unsigned overflow", func(t *testing.T) {
		alu := NewALU("ALU")
		alu.Execute(ALUOp_ADD, 0xFFFFFFFF, 1) // Wraps around
		assert.True(t, alu.HasCarry())
		assert.Equal(t, uint32(0), alu.Result())
	})

	t.Run("ADD sets overflow on signed overflow", func(t *testing.T) {
		alu := NewALU("ALU")
		// 0x7FFFFFFF + 1 = overflow (positive + positive = negative)
		alu.Execute(ALUOp_ADD, 0x7FFFFFFF, 1)
		assert.True(t, alu.HasOverflow())
		assert.True(t, alu.IsNegative())
	})
}

func TestALUSubtraction(t *testing.T) {
	t.Run("SUB basic", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_SUB, 10, 3)
		assert.Equal(t, uint32(7), result)
	})

	t.Run("SUB same values", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_SUB, 42, 42)
		assert.Equal(t, uint32(0), result)
		assert.True(t, alu.IsZero())
	})

	t.Run("SUB sets carry when no borrow", func(t *testing.T) {
		alu := NewALU("ALU")
		alu.Execute(ALUOp_SUB, 10, 3)
		assert.True(t, alu.HasCarry()) // Carry = no borrow
	})

	t.Run("SUB clears carry on borrow", func(t *testing.T) {
		alu := NewALU("ALU")
		alu.Execute(ALUOp_SUB, 3, 10)
		assert.False(t, alu.HasCarry())
		assert.True(t, alu.IsNegative())
	})

	t.Run("SUB sets overflow on signed underflow", func(t *testing.T) {
		alu := NewALU("ALU")
		// 0x80000000 - 1 = overflow (negative - positive = positive)
		alu.Execute(ALUOp_SUB, 0x80000000, 1)
		assert.True(t, alu.HasOverflow())
	})
}

func TestALUMultiplication(t *testing.T) {
	t.Run("MUL basic", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_MUL, 6, 7)
		assert.Equal(t, uint32(42), result)
	})

	t.Run("MUL by zero", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_MUL, 12345, 0)
		assert.Equal(t, uint32(0), result)
		assert.True(t, alu.IsZero())
	})

	t.Run("MUL by one", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_MUL, 12345, 1)
		assert.Equal(t, uint32(12345), result)
	})

	t.Run("MUL large numbers", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_MUL, 1000, 1000)
		assert.Equal(t, uint32(1000000), result)
	})
}

func TestALUDivision(t *testing.T) {
	t.Run("DIV basic", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_DIV, 42, 7)
		assert.Equal(t, uint32(6), result)
	})

	t.Run("DIV with remainder", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_DIV, 43, 7)
		assert.Equal(t, uint32(6), result) // Integer division truncates
	})

	t.Run("DIV by zero returns zero", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_DIV, 42, 0)
		assert.Equal(t, uint32(0), result)
	})

	t.Run("DIV signed negative", func(t *testing.T) {
		alu := NewALU("ALU")
		// -10 / 3 = -3 (signed division)
		neg10 := uint32(0xFFFFFFF6) // -10 in two's complement
		result := alu.Execute(ALUOp_DIV, neg10, 3)
		assert.Equal(t, int32(-3), int32(result))
	})
}

func TestALUModulo(t *testing.T) {
	t.Run("MOD basic", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_MOD, 43, 7)
		assert.Equal(t, uint32(1), result)
	})

	t.Run("MOD exact division", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_MOD, 42, 7)
		assert.Equal(t, uint32(0), result)
		assert.True(t, alu.IsZero())
	})

	t.Run("MOD by zero returns zero", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_MOD, 42, 0)
		assert.Equal(t, uint32(0), result)
	})

	t.Run("MOD signed negative", func(t *testing.T) {
		alu := NewALU("ALU")
		// -10 % 3 = -1 (signed modulo)
		neg10 := uint32(0xFFFFFFF6) // -10 in two's complement
		result := alu.Execute(ALUOp_MOD, neg10, 3)
		assert.Equal(t, int32(-1), int32(result))
	})
}

// =============================================================================
// ALU Bitwise Operations Tests
// =============================================================================

func TestALUBitwiseOperations(t *testing.T) {
	t.Run("AND", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_AND, 0xFF00FF00, 0x0F0F0F0F)
		assert.Equal(t, uint32(0x0F000F00), result)
	})

	t.Run("OR", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_OR, 0xFF00FF00, 0x00FF00FF)
		assert.Equal(t, uint32(0xFFFFFFFF), result)
	})

	t.Run("XOR", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_XOR, 0xAAAAAAAA, 0x55555555)
		assert.Equal(t, uint32(0xFFFFFFFF), result)
	})

	t.Run("XOR same value is zero", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_XOR, 0x12345678, 0x12345678)
		assert.Equal(t, uint32(0), result)
		assert.True(t, alu.IsZero())
	})

	t.Run("NOT", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_NOT, 0xFF00FF00, 0) // B is ignored
		assert.Equal(t, uint32(0x00FF00FF), result)
	})

	t.Run("NOT zero", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_NOT, 0, 0)
		assert.Equal(t, uint32(0xFFFFFFFF), result)
	})
}

// =============================================================================
// ALU Shift Operations Tests
// =============================================================================

func TestALUShiftOperations(t *testing.T) {
	t.Run("LSL basic", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_LSL, 1, 4)
		assert.Equal(t, uint32(16), result)
	})

	t.Run("LSL by zero", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_LSL, 42, 0)
		assert.Equal(t, uint32(42), result)
	})

	t.Run("LSL overflow", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_LSL, 0x80000000, 1)
		assert.Equal(t, uint32(0), result)
		assert.True(t, alu.HasCarry()) // Last bit shifted out was 1
	})

	t.Run("LSR basic", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_LSR, 16, 4)
		assert.Equal(t, uint32(1), result)
	})

	t.Run("LSR fills with zeros", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_LSR, 0x80000000, 1)
		assert.Equal(t, uint32(0x40000000), result) // Logical shift fills with 0
		assert.False(t, alu.IsNegative())
	})

	t.Run("ASR sign extends", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_ASR, 0x80000000, 1)
		assert.Equal(t, uint32(0xC0000000), result) // Arithmetic shift preserves sign
		assert.True(t, alu.IsNegative())
	})

	t.Run("ASR positive number", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_ASR, 0x40000000, 1)
		assert.Equal(t, uint32(0x20000000), result)
		assert.False(t, alu.IsNegative())
	})

	t.Run("ASL is same as LSL", func(t *testing.T) {
		alu := NewALU("ALU")
		lslResult := alu.Execute(ALUOp_LSL, 0x12345678, 4)
		aslResult := alu.Execute(ALUOp_ASL, 0x12345678, 4)
		assert.Equal(t, lslResult, aslResult)
	})

	t.Run("Shift amount is masked to 5 bits", func(t *testing.T) {
		alu := NewALU("ALU")
		// Shift by 33 should be same as shift by 1 (33 & 0x1F = 1)
		result := alu.Execute(ALUOp_LSL, 1, 33)
		assert.Equal(t, uint32(2), result)
	})
}

// =============================================================================
// ALU Compare Operation Tests
// =============================================================================

func TestALUCompare(t *testing.T) {
	t.Run("CMP equal values sets zero flag", func(t *testing.T) {
		alu := NewALU("ALU")
		alu.Execute(ALUOp_CMP, 42, 42)
		assert.True(t, alu.IsZero())
		assert.True(t, alu.HasCarry()) // No borrow
	})

	t.Run("CMP A > B (unsigned)", func(t *testing.T) {
		alu := NewALU("ALU")
		alu.Execute(ALUOp_CMP, 10, 5)
		assert.False(t, alu.IsZero())
		assert.True(t, alu.HasCarry())
		assert.False(t, alu.IsNegative())
	})

	t.Run("CMP A < B (unsigned)", func(t *testing.T) {
		alu := NewALU("ALU")
		alu.Execute(ALUOp_CMP, 5, 10)
		assert.False(t, alu.IsZero())
		assert.False(t, alu.HasCarry())
		assert.True(t, alu.IsNegative())
	})

	t.Run("CMP signed comparison", func(t *testing.T) {
		alu := NewALU("ALU")
		// -1 vs 1 (signed: -1 < 1)
		alu.Execute(ALUOp_CMP, 0xFFFFFFFF, 1)
		// Result is 0xFFFFFFFE, which is negative
		assert.True(t, alu.IsNegative())
	})
}

// =============================================================================
// ALU NOP Tests
// =============================================================================

func TestALUNop(t *testing.T) {
	t.Run("NOP passes through A", func(t *testing.T) {
		alu := NewALU("ALU")
		result := alu.Execute(ALUOp_NOP, 42, 100)
		assert.Equal(t, uint32(42), result)
	})
}

// =============================================================================
// ALU Disabled Tests
// =============================================================================

func TestALUDisabled(t *testing.T) {
	t.Run("Disabled ALU does not compute", func(t *testing.T) {
		alu := NewALU("ALU")

		// Set up an operation
		alu.SetOperands(5, 3)
		alu.SetOperation(ALUOp_ADD)

		// Disable and compute
		alu.Disable()
		err := alu.Compute()
		require.NoError(t, err)

		// Output should remain zero (initial value)
		assert.Equal(t, uint32(0), alu.Result())

		// Re-enable and compute
		alu.Enable()
		err = alu.Compute()
		require.NoError(t, err)
		assert.Equal(t, uint32(8), alu.Result())
	})
}

// =============================================================================
// ALU Reset Tests
// =============================================================================

func TestALUReset(t *testing.T) {
	t.Run("Reset clears outputs", func(t *testing.T) {
		alu := NewALU("ALU")

		// Perform an operation
		alu.Execute(ALUOp_ADD, 100, 200)
		assert.Equal(t, uint32(300), alu.Result())

		// Reset
		alu.Reset()
		assert.Equal(t, uint32(0), alu.Result())
		assert.Equal(t, uint32(0), alu.GetFlags())
	})
}

// =============================================================================
// ALU Registry Tests
// =============================================================================

func TestALURegistry(t *testing.T) {
	t.Run("ALU descriptor is registered", func(t *testing.T) {
		desc, err := Registry.Get("ALU")
		require.NoError(t, err)

		assert.Equal(t, "ALU", desc.Name)
		assert.Equal(t, "ALU", desc.DisplayName)
		assert.Equal(t, CategoryALU, desc.Category)
		assert.NotEmpty(t, desc.Description)
	})

	t.Run("Create ALU from registry", func(t *testing.T) {
		comp, err := Registry.Create("ALU", "my_alu", nil)
		require.NoError(t, err)

		assert.Equal(t, "my_alu", comp.Name())
		assert.Equal(t, "ALU", comp.Type())

		alu, ok := comp.(*ALU)
		require.True(t, ok)

		// Verify it works
		result := alu.Execute(ALUOp_ADD, 10, 20)
		assert.Equal(t, uint32(30), result)
	})
}

// =============================================================================
// ALU Edge Cases
// =============================================================================

func TestALUEdgeCases(t *testing.T) {
	t.Run("Max uint32 values", func(t *testing.T) {
		alu := NewALU("ALU")
		alu.Execute(ALUOp_AND, 0xFFFFFFFF, 0xFFFFFFFF)
		assert.Equal(t, uint32(0xFFFFFFFF), alu.Result())
	})

	t.Run("Chained operations", func(t *testing.T) {
		alu := NewALU("ALU")

		// (5 + 3) then use result
		r1 := alu.Execute(ALUOp_ADD, 5, 3)
		r2 := alu.Execute(ALUOp_MUL, r1, 2)
		assert.Equal(t, uint32(16), r2)
	})

	t.Run("All flags can be set independently", func(t *testing.T) {
		alu := NewALU("ALU")

		// Zero flag only
		alu.Execute(ALUOp_SUB, 5, 5)
		assert.True(t, alu.IsZero())
		assert.False(t, alu.IsNegative())

		// Negative flag only
		alu.Execute(ALUOp_SUB, 0, 1)
		assert.False(t, alu.IsZero())
		assert.True(t, alu.IsNegative())
	})
}
