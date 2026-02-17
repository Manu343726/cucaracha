package llvm

// Cucaracha Fixup Encoding/Decoding
//
// This file implements the encoding and decoding of 16-bit immediate values
// for PC-relative relocations in Cucaracha MOVIMM16L/MOVIMM16H instructions.
//
// Cucaracha Instruction Format (32-bit):
//   ┌─────────────────────────────────────────────────────────────┐
//   │ 31-29 │   28-21    │       20-5        │       4-0         │
//   │unused │  register  │  16-bit immediate │   5-bit opcode    │
//   │ (3)   │    (8)     │       (16)        │       (5)         │
//   └─────────────────────────────────────────────────────────────┘
//
// The fixup value encodes a 16-bit immediate that gets OR'd into the instruction.
// For Cucaracha, this means placing the immediate in bits 5-20.
//
// Encoding: fixupValue = (immediate16 & 0xFFFF) << 5
// Decoding: immediate16 = (fixupValue >> 5) & 0xFFFF
//
// This is different from ARM MOVW encoding which places:
//   - Lo12 (bits 0-11 of immediate) in instruction bits 0-11
//   - Hi4 (bits 12-15 of immediate) in instruction bits 16-19
//
// The ARM encoding conflicts with Cucaracha's opcode bits (0-4), causing corruption.

const (
	// CucarachaImmediateBitOffset is the bit position where the 16-bit immediate starts
	CucarachaImmediateBitOffset = 5

	// CucarachaImmediateMask is the mask for extracting the 16-bit immediate after shifting
	CucarachaImmediateMask = 0xFFFF

	// CucarachaOpcodeMask is the mask for the 5-bit opcode (bits 0-4)
	CucarachaOpcodeMask = 0x1F

	// CucarachaRegisterBitOffset is the bit position where the register starts
	CucarachaRegisterBitOffset = 21

	// CucarachaRegisterMask is the mask for the 8-bit register after shifting
	CucarachaRegisterMask = 0xFF
)

// EncodeFixupValue encodes a 16-bit immediate value into the Cucaracha fixup format.
// The result can be OR'd into an instruction to set the immediate field.
//
// Parameters:
//   - immediate16: The 16-bit value to encode (only lower 16 bits are used)
//
// Returns:
//   - The encoded fixup value with the immediate in bits 5-20
func EncodeFixupValue(immediate16 uint32) uint32 {
	return (immediate16 & CucarachaImmediateMask) << CucarachaImmediateBitOffset
}

// DecodeFixupValue decodes a fixup value from the Cucaracha format.
// This extracts the 16-bit immediate from bits 5-20.
//
// Parameters:
//   - fixupValue: The encoded fixup value (typically OR'd into an instruction)
//
// Returns:
//   - The 16-bit immediate value
func DecodeFixupValue(fixupValue uint32) uint32 {
	return (fixupValue >> CucarachaImmediateBitOffset) & CucarachaImmediateMask
}

// DecodeFixupFromInstruction extracts the 16-bit immediate from a full instruction.
// This is equivalent to DecodeFixupValue but clarifies intent when working with
// complete instruction words.
//
// Parameters:
//   - instruction: The full 32-bit instruction word
//
// Returns:
//   - The 16-bit immediate value from bits 5-20
func DecodeFixupFromInstruction(instruction uint32) uint32 {
	return DecodeFixupValue(instruction)
}

// ExtractOpcode extracts the 5-bit opcode from an instruction.
//
// Parameters:
//   - instruction: The full 32-bit instruction word
//
// Returns:
//   - The 5-bit opcode value
func ExtractOpcode(instruction uint32) uint32 {
	return instruction & CucarachaOpcodeMask
}

// ExtractRegister extracts the 8-bit register index from an instruction.
//
// Parameters:
//   - instruction: The full 32-bit instruction word
//
// Returns:
//   - The 8-bit register index
func ExtractRegister(instruction uint32) uint32 {
	return (instruction >> CucarachaRegisterBitOffset) & CucarachaRegisterMask
}

// CombineLoHiImmediate combines Lo and Hi 16-bit values into a 32-bit address.
// The Lo value comes from a MOVIMM16L instruction (lower 16 bits).
// The Hi value comes from a MOVIMM16H instruction (upper 16 bits).
//
// Parameters:
//   - loImm: The lower 16 bits of the address
//   - hiImm: The upper 16 bits of the address
//
// Returns:
//   - The combined 32-bit address
func CombineLoHiImmediate(loImm, hiImm uint32) uint32 {
	return (hiImm << 16) | (loImm & CucarachaImmediateMask)
}

// SplitToLoHiImmediate splits a 32-bit address into Lo and Hi 16-bit components.
//
// Parameters:
//   - address: The 32-bit address to split
//
// Returns:
//   - loImm: The lower 16 bits for MOVIMM16L
//   - hiImm: The upper 16 bits for MOVIMM16H
func SplitToLoHiImmediate(address uint32) (loImm, hiImm uint32) {
	loImm = address & CucarachaImmediateMask
	hiImm = (address >> 16) & CucarachaImmediateMask
	return
}

// Legacy ARM MOVW Encoding Functions
// These are provided for compatibility with existing .o files that were generated
// with the old ARM-style encoding. Once the LLVM backend is updated, these can be removed.

// DecodeARMFixupLo12Hi4 decodes a 16-bit value from the legacy ARM MOVW format.
// ARM MOVW stores the immediate as:
//   - Bits 0-11: Lo12 (lower 12 bits of the 16-bit value)
//   - Bits 16-19: Hi4 (upper 4 bits of the 16-bit value)
//
// Parameters:
//   - instruction: The instruction word with ARM-style fixup applied
//
// Returns:
//   - The decoded 16-bit immediate value
func DecodeARMFixupLo12Hi4(instruction uint32) uint32 {
	lo12 := instruction & 0xFFF      // bits 0-11
	hi4 := (instruction >> 16) & 0xF // bits 16-19
	return (hi4 << 12) | lo12
}

// EncodeARMFixupLo12Hi4 encodes a 16-bit value into the legacy ARM MOVW format.
// This is the inverse of DecodeARMFixupLo12Hi4.
//
// Parameters:
//   - immediate16: The 16-bit value to encode
//
// Returns:
//   - The ARM-format fixup value
func EncodeARMFixupLo12Hi4(immediate16 uint32) uint32 {
	lo12 := immediate16 & 0xFFF
	hi4 := (immediate16 >> 12) & 0xF
	return (hi4 << 16) | lo12
}
