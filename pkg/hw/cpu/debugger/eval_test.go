package debugger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTokenize tests the expression tokenizer
func TestTokenize(t *testing.T) {
	// Create an evaluator without backend for tokenization tests
	eval := &ExpressionEvaluator{}

	tests := []struct {
		name     string
		expr     string
		expected []Token
		wantErr  bool
	}{
		// Numbers
		{
			name: "decimal number",
			expr: "123",
			expected: []Token{
				{Type: TokenNumber, Value: "123", Num: 123},
			},
		},
		{
			name: "hex number lowercase",
			expr: "0x1a2b",
			expected: []Token{
				{Type: TokenNumber, Value: "0x1a2b", Num: 0x1a2b},
			},
		},
		{
			name: "hex number uppercase",
			expr: "0X1A2B",
			expected: []Token{
				{Type: TokenNumber, Value: "0X1A2B", Num: 0x1a2b},
			},
		},
		{
			name: "binary number",
			expr: "0b1010",
			expected: []Token{
				{Type: TokenNumber, Value: "0b1010", Num: 10},
			},
		},
		{
			name: "binary number with separators",
			expr: "0b1111_0000",
			expected: []Token{
				{Type: TokenNumber, Value: "0b1111_0000", Num: 0xF0},
			},
		},
		{
			name: "zero",
			expr: "0",
			expected: []Token{
				{Type: TokenNumber, Value: "0", Num: 0},
			},
		},

		// Registers
		{
			name: "register r0",
			expr: "r0",
			expected: []Token{
				{Type: TokenRegister, Value: "r0"},
			},
		},
		{
			name: "register r9",
			expr: "r9",
			expected: []Token{
				{Type: TokenRegister, Value: "r9"},
			},
		},
		{
			name: "register sp",
			expr: "sp",
			expected: []Token{
				{Type: TokenRegister, Value: "sp"},
			},
		},
		{
			name: "register lr",
			expr: "lr",
			expected: []Token{
				{Type: TokenRegister, Value: "lr"},
			},
		},
		{
			name: "register pc",
			expr: "pc",
			expected: []Token{
				{Type: TokenRegister, Value: "pc"},
			},
		},
		{
			name: "register cpsr",
			expr: "cpsr",
			expected: []Token{
				{Type: TokenRegister, Value: "cpsr"},
			},
		},
		{
			name: "register uppercase converted to lowercase",
			expr: "R0",
			expected: []Token{
				{Type: TokenRegister, Value: "r0"},
			},
		},
		{
			name: "register SP uppercase",
			expr: "SP",
			expected: []Token{
				{Type: TokenRegister, Value: "sp"},
			},
		},

		// Symbols
		{
			name: "symbol main",
			expr: "main",
			expected: []Token{
				{Type: TokenSymbol, Value: "main"},
			},
		},
		{
			name: "symbol with underscore",
			expr: "_start",
			expected: []Token{
				{Type: TokenSymbol, Value: "_start"},
			},
		},
		{
			name: "symbol with numbers",
			expr: "func123",
			expected: []Token{
				{Type: TokenSymbol, Value: "func123"},
			},
		},

		// Operators
		{
			name: "plus operator",
			expr: "+",
			expected: []Token{
				{Type: TokenPlus, Value: "+"},
			},
		},
		{
			name: "minus operator",
			expr: "-",
			expected: []Token{
				{Type: TokenMinus, Value: "-"},
			},
		},
		{
			name: "multiply operator",
			expr: "*",
			expected: []Token{
				{Type: TokenMul, Value: "*"},
			},
		},
		{
			name: "divide operator",
			expr: "/",
			expected: []Token{
				{Type: TokenDiv, Value: "/"},
			},
		},
		{
			name: "modulo operator",
			expr: "%",
			expected: []Token{
				{Type: TokenMod, Value: "%"},
			},
		},
		{
			name: "and operator",
			expr: "&",
			expected: []Token{
				{Type: TokenAnd, Value: "&"},
			},
		},
		{
			name: "or operator",
			expr: "|",
			expected: []Token{
				{Type: TokenOr, Value: "|"},
			},
		},
		{
			name: "xor operator",
			expr: "^",
			expected: []Token{
				{Type: TokenXor, Value: "^"},
			},
		},
		{
			name: "shift left operator",
			expr: "<<",
			expected: []Token{
				{Type: TokenShiftLeft, Value: "<<"},
			},
		},
		{
			name: "shift right operator",
			expr: ">>",
			expected: []Token{
				{Type: TokenShiftRight, Value: ">>"},
			},
		},

		// Brackets and parentheses
		{
			name: "brackets",
			expr: "[]",
			expected: []Token{
				{Type: TokenLBracket, Value: "["},
				{Type: TokenRBracket, Value: "]"},
			},
		},
		{
			name: "parentheses",
			expr: "()",
			expected: []Token{
				{Type: TokenLParen, Value: "("},
				{Type: TokenRParen, Value: ")"},
			},
		},

		// Complex expressions
		{
			name: "simple addition",
			expr: "r0 + 1",
			expected: []Token{
				{Type: TokenRegister, Value: "r0"},
				{Type: TokenPlus, Value: "+"},
				{Type: TokenNumber, Value: "1", Num: 1},
			},
		},
		{
			name: "arithmetic expression",
			expr: "r0 + r1 * 2",
			expected: []Token{
				{Type: TokenRegister, Value: "r0"},
				{Type: TokenPlus, Value: "+"},
				{Type: TokenRegister, Value: "r1"},
				{Type: TokenMul, Value: "*"},
				{Type: TokenNumber, Value: "2", Num: 2},
			},
		},
		{
			name: "memory access",
			expr: "[sp]",
			expected: []Token{
				{Type: TokenLBracket, Value: "["},
				{Type: TokenRegister, Value: "sp"},
				{Type: TokenRBracket, Value: "]"},
			},
		},
		{
			name: "memory access with offset",
			expr: "[sp + 4]",
			expected: []Token{
				{Type: TokenLBracket, Value: "["},
				{Type: TokenRegister, Value: "sp"},
				{Type: TokenPlus, Value: "+"},
				{Type: TokenNumber, Value: "4", Num: 4},
				{Type: TokenRBracket, Value: "]"},
			},
		},
		{
			name: "parenthesized expression",
			expr: "(r0 + 1) * 2",
			expected: []Token{
				{Type: TokenLParen, Value: "("},
				{Type: TokenRegister, Value: "r0"},
				{Type: TokenPlus, Value: "+"},
				{Type: TokenNumber, Value: "1", Num: 1},
				{Type: TokenRParen, Value: ")"},
				{Type: TokenMul, Value: "*"},
				{Type: TokenNumber, Value: "2", Num: 2},
			},
		},
		{
			name: "hex in expression",
			expr: "r0 | 0xFF",
			expected: []Token{
				{Type: TokenRegister, Value: "r0"},
				{Type: TokenOr, Value: "|"},
				{Type: TokenNumber, Value: "0xFF", Num: 0xFF},
			},
		},
		{
			name: "shift expression",
			expr: "1 << 4",
			expected: []Token{
				{Type: TokenNumber, Value: "1", Num: 1},
				{Type: TokenShiftLeft, Value: "<<"},
				{Type: TokenNumber, Value: "4", Num: 4},
			},
		},
		{
			name: "complex expression with spaces",
			expr: "  r0  +  r1  ",
			expected: []Token{
				{Type: TokenRegister, Value: "r0"},
				{Type: TokenPlus, Value: "+"},
				{Type: TokenRegister, Value: "r1"},
			},
		},
		{
			name: "unary minus",
			expr: "-1",
			expected: []Token{
				{Type: TokenMinus, Value: "-"},
				{Type: TokenNumber, Value: "1", Num: 1},
			},
		},
		{
			name:     "empty expression",
			expr:     "",
			expected: []Token{},
		},
		{
			name:     "whitespace only",
			expr:     "   ",
			expected: []Token{},
		},

		// Error cases
		{
			name:    "invalid character",
			expr:    "r0 @ 1",
			wantErr: true,
		},
		{
			name:    "incomplete shift left",
			expr:    "1 < 2",
			wantErr: true,
		},
		{
			name:    "incomplete shift right",
			expr:    "1 > 2",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := eval.Tokenize(tt.expr)

			if tt.wantErr {
				assert.Error(t, err, "expected error")
				return
			}

			require.NoError(t, err)
			require.Len(t, tokens, len(tt.expected), "token count mismatch")

			for i, tok := range tokens {
				exp := tt.expected[i]
				assert.Equal(t, exp.Type, tok.Type, "token[%d].Type mismatch", i)
				assert.Equal(t, exp.Value, tok.Value, "token[%d].Value mismatch", i)
				if tok.Type == TokenNumber {
					assert.Equal(t, exp.Num, tok.Num, "token[%d].Num mismatch", i)
				}
			}
		})
	}
}

// TestIsRegisterName tests the register name detection
func TestIsRegisterName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid registers
		{"r0", "r0", true},
		{"r1", "r1", true},
		{"r9", "r9", true},
		{"sp", "sp", true},
		{"lr", "lr", true},
		{"pc", "pc", true},
		{"cpsr", "cpsr", true},

		// Invalid registers
		{"r10", "r10", false},
		{"r-1", "r-1", false},
		{"rr0", "rr0", false},
		{"main", "main", false},
		{"R0 uppercase", "R0", false}, // IsRegisterName expects lowercase
		{"empty", "", false},
		{"space", "sp ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRegisterName(tt.input)
			assert.Equal(t, tt.expected, result, "IsRegisterName(%q)", tt.input)
		})
	}
}

// TestHelperFunctions tests the character classification helpers
func TestHelperFunctions(t *testing.T) {
	t.Run("IsDigit", func(t *testing.T) {
		for c := byte('0'); c <= '9'; c++ {
			assert.True(t, IsDigit(c), "IsDigit(%c) should be true", c)
		}
		assert.False(t, IsDigit('a'), "IsDigit('a') should be false")
		assert.False(t, IsDigit('A'), "IsDigit('A') should be false")
	})

	t.Run("IsHexDigit", func(t *testing.T) {
		validHex := "0123456789abcdefABCDEF"
		for _, c := range validHex {
			assert.True(t, IsHexDigit(byte(c)), "IsHexDigit(%c) should be true", c)
		}
		assert.False(t, IsHexDigit('g'), "IsHexDigit('g') should be false")
		assert.False(t, IsHexDigit('G'), "IsHexDigit('G') should be false")
	})

	t.Run("IsAlpha", func(t *testing.T) {
		for c := byte('a'); c <= 'z'; c++ {
			assert.True(t, IsAlpha(c), "IsAlpha(%c) should be true", c)
		}
		for c := byte('A'); c <= 'Z'; c++ {
			assert.True(t, IsAlpha(c), "IsAlpha(%c) should be true", c)
		}
		assert.False(t, IsAlpha('0'), "IsAlpha('0') should be false")
		assert.False(t, IsAlpha('_'), "IsAlpha('_') should be false")
	})

	t.Run("IsAlphaNum", func(t *testing.T) {
		assert.True(t, IsAlphaNum('a'), "IsAlphaNum('a') should be true")
		assert.True(t, IsAlphaNum('Z'), "IsAlphaNum('Z') should be true")
		assert.True(t, IsAlphaNum('5'), "IsAlphaNum('5') should be true")
		assert.False(t, IsAlphaNum('_'), "IsAlphaNum('_') should be false")
	})
}

// TestFormatBinary tests the binary formatting function
func TestFormatBinary(t *testing.T) {
	tests := []struct {
		input    uint32
		expected string
	}{
		{0, "00000000_00000000_00000000_00000000"},
		{1, "00000000_00000000_00000000_00000001"},
		{0xFF, "00000000_00000000_00000000_11111111"},
		{0xFFFF, "00000000_00000000_11111111_11111111"},
		{0xFFFFFFFF, "11111111_11111111_11111111_11111111"},
		{0x12345678, "00010010_00110100_01010110_01111000"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatBinary(tt.input)
			assert.Equal(t, tt.expected, result, "FormatBinary(0x%X)", tt.input)
		})
	}
}

// TestEvalArithmetic tests arithmetic expression evaluation
func TestEvalArithmetic(t *testing.T) {
	backend := NewBackend(1024 * 1024)
	eval := NewExpressionEvaluator(backend)

	tests := []struct {
		name     string
		expr     string
		expected uint32
		wantErr  bool
	}{
		// Basic numbers
		{"decimal", "42", 42, false},
		{"hex", "0x10", 16, false},
		{"binary", "0b1010", 10, false},
		{"zero", "0", 0, false},

		// Addition
		{"add two numbers", "1 + 2", 3, false},
		{"add three numbers", "1 + 2 + 3", 6, false},
		{"add with hex", "0x10 + 16", 32, false},

		// Subtraction
		{"subtract", "10 - 3", 7, false},
		{"subtract chain", "10 - 3 - 2", 5, false},

		// Multiplication
		{"multiply", "3 * 4", 12, false},
		{"multiply chain", "2 * 3 * 4", 24, false},

		// Division
		{"divide", "12 / 3", 4, false},
		{"divide integer", "10 / 3", 3, false},
		{"divide by zero", "10 / 0", 0, true},

		// Modulo
		{"modulo", "10 % 3", 1, false},
		{"modulo by zero", "10 % 0", 0, true},

		// Precedence
		{"mul before add", "1 + 2 * 3", 7, false},
		{"add then mul", "2 * 3 + 1", 7, false},
		{"parentheses", "(1 + 2) * 3", 9, false},

		// Bitwise
		{"and", "0xFF & 0x0F", 0x0F, false},
		{"or", "0xF0 | 0x0F", 0xFF, false},
		{"xor", "0xFF ^ 0x0F", 0xF0, false},

		// Shifts
		{"shift left", "1 << 4", 16, false},
		{"shift right", "16 >> 2", 4, false},

		// Unary minus
		{"unary minus", "-1", 0xFFFFFFFF, false},
		{"unary minus with add", "10 + -3", 7, false},

		// Complex
		{"complex", "((1 + 2) * 3 + 4) / 2", 6, false},
		{"complex with bits", "(0xFF & 0x0F) << 4", 0xF0, false},

		// Errors
		{"empty", "", 0, true},
		{"unbalanced paren", "(1 + 2", 0, true},
		{"extra paren", "1 + 2)", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := eval.Eval(tt.expr)

			if tt.wantErr {
				assert.Error(t, err, "expected error for %q", tt.expr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result, "Eval(%q)", tt.expr)
		})
	}
}

// TestEvalRegisters tests register evaluation with a backend
func TestEvalRegisters(t *testing.T) {
	backend := NewBackend(1024 * 1024)
	eval := NewExpressionEvaluator(backend)

	// Set up some register values using Backend methods
	require.NoError(t, backend.WriteRegister("r0", 100))
	require.NoError(t, backend.WriteRegister("r1", 200))
	require.NoError(t, backend.WriteRegister("sp", 0x1000))
	require.NoError(t, backend.WriteRegister("lr", 0x2000))
	// PC is set through state directly since WriteRegister doesn't handle it
	backend.Runner().Debugger().State().PC = 0x3000

	tests := []struct {
		name     string
		expr     string
		expected uint32
	}{
		{"r0", "r0", 100},
		{"r1", "r1", 200},
		{"sp", "sp", 0x1000},
		{"lr", "lr", 0x2000},
		{"pc", "pc", 0x3000},
		{"r0 + r1", "r0 + r1", 300},
		{"sp + 16", "sp + 16", 0x1010},
		{"r0 * 2", "r0 * 2", 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := eval.Eval(tt.expr)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result, "Eval(%q)", tt.expr)
		})
	}
}

// TestEvalMemory tests memory dereference evaluation
func TestEvalMemory(t *testing.T) {
	backend := NewBackend(1024 * 1024)
	eval := NewExpressionEvaluator(backend)

	// Set up memory
	require.NoError(t, backend.WriteRegister("sp", 0x1000))

	// Write test values to memory (little-endian)
	mem := backend.Runner().Debugger().State().Memory
	mem[0x100] = 0x78
	mem[0x101] = 0x56
	mem[0x102] = 0x34
	mem[0x103] = 0x12 // Value at 0x100 = 0x12345678

	mem[0x1000] = 0xEF
	mem[0x1001] = 0xBE
	mem[0x1002] = 0xAD
	mem[0x1003] = 0xDE // Value at sp = 0xDEADBEEF

	tests := []struct {
		name     string
		expr     string
		expected uint32
	}{
		{"memory at address", "[0x100]", 0x12345678},
		{"memory at sp", "[sp]", 0xDEADBEEF},
		{"memory at expr", "[0x100 + 0]", 0x12345678},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := eval.Eval(tt.expr)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result, "Eval(%q)", tt.expr)
		})
	}
}
