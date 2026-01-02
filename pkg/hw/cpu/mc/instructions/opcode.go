package instructions

// Represents an instruction opcode
type OpCode uint

const (
	// No-Operation
	OpCode_NOP OpCode = iota
	// Copy 16 most significant bits of immediate value into register
	OpCode_MOV_IMM16H
	// Copy 16 least significant bits of immediate value into register
	OpCode_MOV_IMM16L
	// Copy value of one register into another
	OpCode_MOV
	// Load value from memory into register
	OpCode_LD
	// Save value of register into memory
	OpCode_ST
	// Add values of two registers, save result into third
	OpCode_ADD
	// Substract values of two registers, save result into third
	OpCode_SUB
	// Multiply values of two registers, save result into third
	OpCode_MUL
	// Divide values of two registers, save result into third
	OpCode_DIV
	// Compute register value modulo other register value, save result into third
	OpCode_MOD
	// Compare values of two registers, set flags in CPSR
	OpCode_CMP
	// Uncondiional jump to address in register
	OpCode_JMP
	// Conditional jump to address in register if the result of masking CPSR with a given register is not zero
	OpCode_CJMP
	// Conditional jump to address in register if zero flag is not set
	OpCode_LSL
	// Logical shift right: shift first register right by second register amount, save to third
	OpCode_LSR
	// Arithmetic shift left: shift first register left by second register amount (sign-extend), save to third
	OpCode_ASL
	// Arithmetic shift right: shift first register right by second register amount (sign-extend), save to third
	OpCode_ASR

	// Total opcodes implemented
	TOTAL_OPCODES
)

// Returns the mnemonic of the instruction opcode
func (op OpCode) String() string {
	return Opcodes.Mnemonic(op)
}
