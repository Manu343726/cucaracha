package llvm

import (
	"fmt"
	"testing"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeFixupValue(t *testing.T) {
	tests := []struct {
		name      string
		immediate uint32
		expected  uint32
	}{
		{
			name:      "zero immediate",
			immediate: 0x0000,
			expected:  0x00000000,
		},
		{
			name:      "small immediate",
			immediate: 0x001C,     // 28 decimal
			expected:  0x00000380, // 28 << 5 = 896
		},
		{
			name:      "medium immediate",
			immediate: 0x1234,
			expected:  0x00024680, // 0x1234 << 5
		},
		{
			name:      "max 16-bit immediate",
			immediate: 0xFFFF,
			expected:  0x001FFFE0, // 0xFFFF << 5
		},
		{
			name:      "truncates to 16 bits",
			immediate: 0x1FFFF,    // 17 bits
			expected:  0x001FFFE0, // only lower 16 bits used
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeFixupValue(tt.immediate)
			assert.Equal(t, tt.expected, result, "EncodeFixupValue(%#x)", tt.immediate)
		})
	}
}

func TestDecodeFixupValue(t *testing.T) {
	tests := []struct {
		name       string
		fixupValue uint32
		expected   uint32
	}{
		{
			name:       "zero fixup",
			fixupValue: 0x00000000,
			expected:   0x0000,
		},
		{
			name:       "small fixup",
			fixupValue: 0x00000380, // 28 << 5
			expected:   0x001C,
		},
		{
			name:       "medium fixup",
			fixupValue: 0x00024680, // 0x1234 << 5
			expected:   0x1234,
		},
		{
			name:       "max 16-bit fixup",
			fixupValue: 0x001FFFE0, // 0xFFFF << 5
			expected:   0xFFFF,
		},
		{
			name:       "ignores bits outside immediate field",
			fixupValue: 0xFF24068F, // has opcode bits and register bits set
			expected:   0x2034,     // (0xFF24068F >> 5) & 0xFFFF = 0x2034
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DecodeFixupValue(tt.fixupValue)
			assert.Equal(t, tt.expected, result, "DecodeFixupValue(%#x)", tt.fixupValue)
		})
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	// Test that encoding then decoding returns the original value
	values := []uint32{0, 1, 28, 100, 0x1234, 0x8000, 0xFFFF}

	for _, original := range values {
		encoded := EncodeFixupValue(original)
		decoded := DecodeFixupValue(encoded)
		assert.Equal(t, original, decoded, "roundtrip for %#x", original)
	}
}

func TestDecodeFixupFromInstruction(t *testing.T) {
	// Test decoding from a full instruction word
	// Instruction: MOVIMM16L #0x50, r4
	// opcode = 2 (bits 0-4)
	// immediate = 0x50 (bits 5-20)
	// register = 20 (bits 21-28)

	// Build the instruction: (reg << 21) | (imm << 5) | opcode
	instruction := uint32((20 << 21) | (0x50 << 5) | 2)

	immediate := DecodeFixupFromInstruction(instruction)
	assert.Equal(t, uint32(0x50), immediate)
}

func TestExtractOpcode(t *testing.T) {
	tests := []struct {
		name        string
		instruction uint32
		expected    uint32
	}{
		{
			name:        "MOVIMM16L opcode",
			instruction: 0x02800A02, // some instruction with opcode 2
			expected:    2,
		},
		{
			name:        "MOVIMM16H opcode",
			instruction: 0x02800A01, // some instruction with opcode 1
			expected:    1,
		},
		{
			name:        "NOP opcode",
			instruction: 0x00000000,
			expected:    0,
		},
		{
			name:        "max opcode",
			instruction: 0x0000001F,
			expected:    31,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractOpcode(tt.instruction)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractRegister(t *testing.T) {
	tests := []struct {
		name        string
		instruction uint32
		expected    uint32
	}{
		{
			name:        "r0",
			instruction: uint32(0 << 21),
			expected:    0,
		},
		{
			name:        "r4 (index 20 in Cucaracha)",
			instruction: uint32(20 << 21),
			expected:    20,
		},
		{
			name:        "max register",
			instruction: uint32(255 << 21),
			expected:    255,
		},
		{
			name:        "with other bits set",
			instruction: uint32((20 << 21) | (0x1234 << 5) | 2),
			expected:    20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractRegister(tt.instruction)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCombineLoHiImmediate(t *testing.T) {
	tests := []struct {
		name     string
		loImm    uint32
		hiImm    uint32
		expected uint32
	}{
		{
			name:     "zero address",
			loImm:    0,
			hiImm:    0,
			expected: 0,
		},
		{
			name:     "lo only",
			loImm:    0x1234,
			hiImm:    0,
			expected: 0x00001234,
		},
		{
			name:     "hi only",
			loImm:    0,
			hiImm:    0x5678,
			expected: 0x56780000,
		},
		{
			name:     "both lo and hi",
			loImm:    0x1234,
			hiImm:    0x5678,
			expected: 0x56781234,
		},
		{
			name:     "max address",
			loImm:    0xFFFF,
			hiImm:    0xFFFF,
			expected: 0xFFFFFFFF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CombineLoHiImmediate(tt.loImm, tt.hiImm)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSplitToLoHiImmediate(t *testing.T) {
	tests := []struct {
		name       string
		address    uint32
		expectedLo uint32
		expectedHi uint32
	}{
		{
			name:       "zero address",
			address:    0,
			expectedLo: 0,
			expectedHi: 0,
		},
		{
			name:       "small address",
			address:    0x00001234,
			expectedLo: 0x1234,
			expectedHi: 0,
		},
		{
			name:       "large address",
			address:    0x56780000,
			expectedLo: 0,
			expectedHi: 0x5678,
		},
		{
			name:       "full address",
			address:    0x56781234,
			expectedLo: 0x1234,
			expectedHi: 0x5678,
		},
		{
			name:       "max address",
			address:    0xFFFFFFFF,
			expectedLo: 0xFFFF,
			expectedHi: 0xFFFF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lo, hi := SplitToLoHiImmediate(tt.address)
			assert.Equal(t, tt.expectedLo, lo, "lo for %#x", tt.address)
			assert.Equal(t, tt.expectedHi, hi, "hi for %#x", tt.address)
		})
	}
}

func TestSplitCombineRoundTrip(t *testing.T) {
	// Test that splitting then combining returns the original address
	addresses := []uint32{0, 0x1234, 0x56780000, 0x56781234, 0xFFFFFFFF}

	for _, original := range addresses {
		lo, hi := SplitToLoHiImmediate(original)
		combined := CombineLoHiImmediate(lo, hi)
		assert.Equal(t, original, combined, "roundtrip for %#x", original)
	}
}

// Tests for legacy ARM encoding (for compatibility with old .o files)

func TestDecodeARMFixupLo12Hi4(t *testing.T) {
	tests := []struct {
		name        string
		instruction uint32
		expected    uint32
	}{
		{
			name:        "zero",
			instruction: 0x00000000,
			expected:    0x0000,
		},
		{
			name:        "small value fits in lo12",
			instruction: 0x0000001C, // 28 in lo12
			expected:    0x001C,
		},
		{
			name:        "value needs hi4",
			instruction: 0x00050234, // hi4=5, lo12=0x234 -> 0x5234
			expected:    0x5234,
		},
		{
			name:        "max value",
			instruction: 0x000F0FFF, // hi4=F, lo12=FFF -> 0xFFFF
			expected:    0xFFFF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DecodeARMFixupLo12Hi4(tt.instruction)
			assert.Equal(t, tt.expected, result, "DecodeARMFixupLo12Hi4(%#x)", tt.instruction)
		})
	}
}

func TestEncodeARMFixupLo12Hi4(t *testing.T) {
	tests := []struct {
		name      string
		immediate uint32
		expected  uint32
	}{
		{
			name:      "zero",
			immediate: 0x0000,
			expected:  0x00000000,
		},
		{
			name:      "small value fits in lo12",
			immediate: 0x001C,
			expected:  0x0000001C,
		},
		{
			name:      "value needs hi4",
			immediate: 0x5234,
			expected:  0x00050234, // hi4=5, lo12=0x234
		},
		{
			name:      "max value",
			immediate: 0xFFFF,
			expected:  0x000F0FFF, // hi4=F, lo12=FFF
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeARMFixupLo12Hi4(tt.immediate)
			assert.Equal(t, tt.expected, result, "EncodeARMFixupLo12Hi4(%#x)", tt.immediate)
		})
	}
}

func TestARMEncodingRoundTrip(t *testing.T) {
	// Test that ARM encoding then decoding returns the original value
	values := []uint32{0, 1, 28, 100, 0x1234, 0x8000, 0xFFFF}

	for _, original := range values {
		encoded := EncodeARMFixupLo12Hi4(original)
		decoded := DecodeARMFixupLo12Hi4(encoded)
		assert.Equal(t, original, decoded, "ARM roundtrip for %#x", original)
	}
}

// Test showing the problem with ARM encoding in Cucaracha instructions

func TestARMEncodingCorruptsOpcode(t *testing.T) {
	// Demonstrate why ARM encoding doesn't work for Cucaracha:
	// ARM encoding puts Lo12 in bits 0-11, which overlaps with opcode (bits 0-4)

	// Start with a clean MOVIMM16L instruction (opcode=2, reg=r4/20, imm=0)
	baseInstruction := uint32((20 << 21) | (0 << 5) | 2)
	originalOpcode := ExtractOpcode(baseInstruction)
	assert.Equal(t, uint32(2), originalOpcode, "base opcode should be 2")

	// Try to apply ARM-encoded fixup for immediate 0x1C (28)
	armFixup := EncodeARMFixupLo12Hi4(0x1C) // = 0x1C in bits 0-11
	corruptedInstruction := baseInstruction | armFixup
	corruptedOpcode := ExtractOpcode(corruptedInstruction)

	// The opcode is now corrupted! 2 | 0x1C = 0x1E = 30
	assert.NotEqual(t, uint32(2), corruptedOpcode, "ARM fixup should corrupt opcode")
	assert.Equal(t, uint32(0x1E), corruptedOpcode, "corrupted opcode = 2 | 0x1C")
}

func TestCucarachaEncodingPreservesOpcode(t *testing.T) {
	// Demonstrate that Cucaracha encoding preserves the opcode:
	// Cucaracha encoding puts the immediate in bits 5-20, not touching opcode

	// Start with a clean MOVIMM16L instruction (opcode=2, reg=r4/20, imm=0)
	baseInstruction := uint32((20 << 21) | (0 << 5) | 2)
	originalOpcode := ExtractOpcode(baseInstruction)
	assert.Equal(t, uint32(2), originalOpcode, "base opcode should be 2")

	// Apply Cucaracha-encoded fixup for immediate 0x1C (28)
	cucarachaFixup := EncodeFixupValue(0x1C) // = 0x1C << 5 = 0x380
	fixedInstruction := baseInstruction | cucarachaFixup
	fixedOpcode := ExtractOpcode(fixedInstruction)

	// The opcode is preserved!
	assert.Equal(t, uint32(2), fixedOpcode, "Cucaracha fixup should preserve opcode")

	// And we can extract the correct immediate
	extractedImmediate := DecodeFixupValue(fixedInstruction)
	assert.Equal(t, uint32(0x1C), extractedImmediate, "immediate should be extractable")
}

// Tests ensuring fixup encoding is compatible with MOV immediate instructions
// These tests use mc.Instruction and instruction descriptors to verify
// that the fixup encoding is compatible with the actual instruction encoding.

// encodeInstruction is a helper to encode an instruction to uint32
func encodeInstruction(instr *instructions.Instruction) uint32 {
	raw := instr.Raw()
	return raw.Encode()
}

func TestFixupEncodingCompatibleWithMOVIMM16L(t *testing.T) {
	// Test that fixup encoding is compatible with properly encoded MOVIMM16L instructions

	tests := []struct {
		name      string
		immediate uint16
		register  string
	}{
		{"zero immediate, r0", 0x0000, "r0"},
		{"small immediate, r1", 0x001C, "r1"},
		{"medium immediate, r4", 0x1234, "r4"},
		{"max immediate, r7", 0xFFFF, "r7"},
		{"address low bits, r2", 0x5678, "r2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build instruction using mc.Instr with actual instruction descriptors
			instr := mc.Instr(instructions.OpCode_MOV_IMM16L).
				U(uint32(tt.immediate)).
				NamedR(tt.register).
				MustBuild()

			// Encode the instruction using the real encoder
			encoded := encodeInstruction(instr)

			// Verify opcode is correct
			assert.Equal(t, uint32(instructions.OpCode_MOV_IMM16L), ExtractOpcode(encoded),
				"opcode should be MOVIMM16L")

			// Verify immediate can be decoded using our fixup decoder
			decodedImm := DecodeFixupFromInstruction(encoded)
			assert.Equal(t, uint32(tt.immediate), decodedImm,
				"immediate should be decodable from properly encoded instruction")

			// Verify that applying a fixup to a zero-immediate instruction produces same result
			zeroImmInstr := mc.Instr(instructions.OpCode_MOV_IMM16L).
				U(0).
				NamedR(tt.register).
				MustBuild()
			zeroEncoded := encodeInstruction(zeroImmInstr)

			// Apply fixup to zero-immediate instruction
			fixup := EncodeFixupValue(uint32(tt.immediate))
			withFixup := zeroEncoded | fixup

			// The result should match the properly encoded instruction
			assert.Equal(t, encoded, withFixup,
				"fixup-applied instruction should match properly encoded instruction")
		})
	}
}

func TestFixupEncodingCompatibleWithMOVIMM16H(t *testing.T) {
	// Test that fixup encoding is compatible with properly encoded MOVIMM16H instructions
	// Now that MOVIMM16H uses the same operand layout as MOVIMM16L:
	// - opcode bits 0-4
	// - immediate bits 5-20
	// - register bits 21-28
	//
	// The fixup encoding works for both instructions!

	tests := []struct {
		name      string
		immediate uint16
		register  string
	}{
		{"zero immediate, r0", 0x0000, "r0"},
		{"small immediate, r1", 0x001C, "r1"},
		{"medium immediate, r4", 0x1234, "r4"},
		{"max immediate, r7", 0xFFFF, "r7"},
		{"address high bits, r2", 0x5678, "r2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build instruction using mc.Instr with actual instruction descriptors
			// MOVIMM16H now has format: imm, dst, src (same encoding order as MOVIMM16L)
			instr := mc.Instr(instructions.OpCode_MOV_IMM16H).
				U(uint32(tt.immediate)).
				NamedR(tt.register). // dst
				NamedR(tt.register). // src (tied to dst)
				MustBuild()

			// Encode the instruction using the real encoder
			encoded := encodeInstruction(instr)

			// Verify opcode is correct
			assert.Equal(t, uint32(instructions.OpCode_MOV_IMM16H), ExtractOpcode(encoded),
				"opcode should be MOVIMM16H")

			// Verify immediate can be decoded using our fixup decoder
			// (now works because MOVIMM16H uses same immediate position as MOVIMM16L)
			decodedImm := DecodeFixupFromInstruction(encoded)
			assert.Equal(t, uint32(tt.immediate), decodedImm,
				"immediate should be decodable from properly encoded instruction")

			// Verify that applying a fixup to a zero-immediate instruction produces same result
			zeroImmInstr := mc.Instr(instructions.OpCode_MOV_IMM16H).
				U(0).
				NamedR(tt.register).
				NamedR(tt.register).
				MustBuild()
			zeroEncoded := encodeInstruction(zeroImmInstr)

			// Apply fixup to zero-immediate instruction
			fixup := EncodeFixupValue(uint32(tt.immediate))
			withFixup := zeroEncoded | fixup

			// The result should match the properly encoded instruction
			assert.Equal(t, encoded, withFixup,
				"fixup-applied instruction should match properly encoded instruction")
		})
	}
}

func TestFixupEncoding32BitAddressWithMOVInstructions(t *testing.T) {
	// Test loading a full 32-bit address using MOVIMM16L + MOVIMM16H pair
	// This is the pattern used by LLVM for loading branch targets
	//
	// Both MOVIMM16L and MOVIMM16H now use the same immediate position:
	// - opcode bits 0-4
	// - immediate bits 5-20
	// - register bits 21-28
	//
	// The fixup encoding works for both instructions!

	testAddresses := []uint32{
		0x00000000, // zero
		0x00001234, // lo only
		0x56780000, // hi only
		0x56781234, // both lo and hi
		0xDEADBEEF, // typical debug address
		0xFFFFFFFF, // max address
	}

	for _, targetAddr := range testAddresses {
		t.Run(fmt.Sprintf("addr_%08X", targetAddr), func(t *testing.T) {
			register := "r4"

			// Split address into lo and hi parts
			loImm, hiImm := SplitToLoHiImmediate(targetAddr)

			// Build MOVIMM16L instruction using mc.Instr
			loInstr := mc.Instr(instructions.OpCode_MOV_IMM16L).
				U(loImm).
				NamedR(register).
				MustBuild()
			loEncoded := encodeInstruction(loInstr)

			// Verify MOVIMM16L has correct opcode
			assert.Equal(t, uint32(instructions.OpCode_MOV_IMM16L), ExtractOpcode(loEncoded),
				"lo instruction should have MOVIMM16L opcode")

			// Decode immediate from MOVIMM16L
			decodedLo := DecodeFixupFromInstruction(loEncoded)
			lowWord := uint16(targetAddr & 0xFFFF)
			assert.Equal(t, uint32(lowWord), decodedLo,
				"MOVIMM16L immediate should decode correctly")

			// Build MOVIMM16H instruction for the high bits (now uses same operand order)
			hiInstr := mc.Instr(instructions.OpCode_MOV_IMM16H).
				U(hiImm).
				NamedR(register).
				NamedR(register).
				MustBuild()
			hiEncoded := encodeInstruction(hiInstr)

			// Verify MOVIMM16H has correct opcode
			assert.Equal(t, uint32(instructions.OpCode_MOV_IMM16H), ExtractOpcode(hiEncoded),
				"hi instruction should have MOVIMM16H opcode")

			// MOVIMM16H now uses same immediate position as MOVIMM16L (bits 5-20)
			decodedHi := DecodeFixupFromInstruction(hiEncoded)
			highWord := uint16(targetAddr >> 16)
			assert.Equal(t, uint32(highWord), decodedHi,
				"MOVIMM16H immediate should decode correctly")

			// Reconstruct the address using CombineLoHiImmediate
			reconstructedAddr := CombineLoHiImmediate(decodedLo, decodedHi)
			assert.Equal(t, targetAddr, reconstructedAddr,
				"reconstructed address should match original")
		})
	}
}

func TestFixupEncodingBitFieldsDoNotOverlap(t *testing.T) {
	// Verify that fixup encoding doesn't interfere with opcode or register fields
	// using actual instruction encoding

	// Use max immediate with MOVIMM16L to r7
	instr := mc.Instr(instructions.OpCode_MOV_IMM16L).
		U(0xFFFF). // max immediate
		NamedR("r7").
		MustBuild()

	encoded := encodeInstruction(instr)

	// Verify opcode is preserved
	assert.Equal(t, uint32(instructions.OpCode_MOV_IMM16L), ExtractOpcode(encoded),
		"opcode should be MOVIMM16L")

	// Verify immediate can be decoded
	decodedImm := DecodeFixupFromInstruction(encoded)
	assert.Equal(t, uint32(0xFFFF), decodedImm,
		"max immediate should be decodable")

	// Verify applying fixup to zero-imm produces same result
	zeroInstr := mc.Instr(instructions.OpCode_MOV_IMM16L).
		U(0).
		NamedR("r7").
		MustBuild()
	zeroEncoded := encodeInstruction(zeroInstr)

	fixup := EncodeFixupValue(0xFFFF)
	withFixup := zeroEncoded | fixup

	assert.Equal(t, encoded, withFixup,
		"fixup-applied instruction should match properly encoded instruction")
}

func TestFixupEncodingWithRealCucarachaOpcodes(t *testing.T) {
	// Test with actual Cucaracha instructions built using mc.Instr

	immediates := []uint16{0, 0x1C, 0x1234, 0x8000, 0xFFFF}
	registers := []string{"r0", "r1", "r2", "r3", "r4", "r5", "r6", "r7"}

	// Test MOVIMM16L
	for _, imm := range immediates {
		for _, reg := range registers {
			name := fmt.Sprintf("MOVIMM16L_imm%04X_%s", imm, reg)
			t.Run(name, func(t *testing.T) {
				instr := mc.Instr(instructions.OpCode_MOV_IMM16L).
					U(uint32(imm)).
					NamedR(reg).
					MustBuild()

				encoded := encodeInstruction(instr)

				// Verify opcode
				assert.Equal(t, uint32(instructions.OpCode_MOV_IMM16L), ExtractOpcode(encoded))

				// Verify immediate can be decoded with the fixup decoder
				assert.Equal(t, uint32(imm), DecodeFixupFromInstruction(encoded))
			})
		}
	}

	// Test MOVIMM16H - now uses same operand layout as MOVIMM16L
	for _, imm := range immediates {
		for _, reg := range registers {
			name := fmt.Sprintf("MOVIMM16H_imm%04X_%s", imm, reg)
			t.Run(name, func(t *testing.T) {
				instr := mc.Instr(instructions.OpCode_MOV_IMM16H).
					U(uint32(imm)).
					NamedR(reg).
					NamedR(reg).
					MustBuild()

				encoded := encodeInstruction(instr)

				// Verify opcode
				assert.Equal(t, uint32(instructions.OpCode_MOV_IMM16H), ExtractOpcode(encoded))

				// MOVIMM16H now uses same immediate position as MOVIMM16L (bits 5-20)
				assert.Equal(t, uint32(imm), DecodeFixupFromInstruction(encoded))
			})
		}
	}
}

func TestFixupEncodingMatchesInstructionDescriptor(t *testing.T) {
	// Verify that the fixup encoding matches the instruction descriptor's operand layout

	// Get MOVIMM16L descriptor
	movimm16lDesc, err := instructions.Instructions.Instruction(instructions.OpCode_MOV_IMM16L)
	require.NoError(t, err)

	// Find the immediate operand descriptor
	var immOperand *instructions.OperandDescriptor
	for _, op := range movimm16lDesc.Operands {
		if op.Kind == instructions.OperandKind_Immediate {
			immOperand = op
			break
		}
	}
	require.NotNil(t, immOperand, "MOVIMM16L should have an immediate operand")

	// Verify our encoding constants match the descriptor
	assert.Equal(t, 5, immOperand.EncodingPosition,
		"immediate should start at bit 5")
	assert.Equal(t, 16, immOperand.EncodingBits,
		"immediate should be 16 bits")

	// Test with a specific value
	testImm := uint16(0xABCD)
	instr := mc.Instr(instructions.OpCode_MOV_IMM16L).
		U(uint32(testImm)).
		NamedR("r0").
		MustBuild()

	encoded := encodeInstruction(instr)
	decoded := DecodeFixupFromInstruction(encoded)

	assert.Equal(t, uint32(testImm), decoded,
		"fixup decoding should match instruction encoding")
}

func TestFixupEncodingWithInstructionDecoder(t *testing.T) {
	// Test that instructions created by applying fixups can be properly decoded

	testCases := []struct {
		immediate uint16
		register  string
	}{
		{0x0000, "r0"},
		{0x1234, "r4"},
		{0xFFFF, "r7"},
	}

	for _, tc := range testCases {
		name := fmt.Sprintf("imm_%04X_%s", tc.immediate, tc.register)
		t.Run(name, func(t *testing.T) {
			// Create zero-immediate instruction
			zeroInstr := mc.Instr(instructions.OpCode_MOV_IMM16L).
				U(0).
				NamedR(tc.register).
				MustBuild()
			zeroEncoded := encodeInstruction(zeroInstr)

			// Apply fixup
			fixup := EncodeFixupValue(uint32(tc.immediate))
			withFixup := zeroEncoded | fixup

			// Decode the instruction
			decodedInstr, err := instructions.Instructions.Decode(withFixup)
			require.NoError(t, err)

			// Verify the decoded instruction
			assert.Equal(t, instructions.OpCode_MOV_IMM16L, decodedInstr.Descriptor.OpCode.OpCode)

			// Check the immediate operand
			for i, op := range decodedInstr.Descriptor.Operands {
				if op.Kind == instructions.OperandKind_Immediate {
					immValue := decodedInstr.OperandValues[i]
					assert.Equal(t, uint32(tc.immediate), uint32(immValue.Encode()),
						"decoded immediate should match")
				}
			}
		})
	}
}
