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

	// Total opcodes implemented
	TOTAL_OPCODES
)

// Returns the mnemonic of the instruction opcode
func (op OpCode) String() string {
	return Opcodes.Mnemonic(op)
}
