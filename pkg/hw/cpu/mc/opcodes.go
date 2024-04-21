package mc

import (
	"errors"
	"fmt"
	"math/bits"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/utils"
)

// Represents an instruction opcode
type OpCode uint

const (
	// No-Operation
	OpCode_NOP OpCode = iota
	// Write immediate value into register
	OpCode_IMM
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

// Contains implementation information of an opcode
type OpCodeDescriptor struct {
	OpCode               OpCode
	BinaryRepresentation uint64
	Mnemonic             string
}

func (d *OpCodeDescriptor) String() string {
	return fmt.Sprintf("%v (code: %v, binary: %v, hex: %v)", d.Mnemonic, d.BinaryRepresentation, utils.FormatUintBinary(d.BinaryRepresentation, Descriptor_Opcodes.OpCodeBits()), utils.FormatUintHex(d.BinaryRepresentation, Descriptor_Opcodes.OpCodeBits()/4))
}

// Returns information about the implemented opcodes
type OpCodesDescriptor struct {
	mnemonics         map[OpCode]string
	mnemonicsToOpCode map[string]OpCode
}

func (d *OpCodesDescriptor) Descriptor(op OpCode) *OpCodeDescriptor {
	return &OpCodeDescriptor{
		OpCode:               op,
		BinaryRepresentation: d.EncodeOpCode(op),
		Mnemonic:             d.Mnemonic(op),
	}
}

// Returns the descriptors of all implemented opcodes
func (d *OpCodesDescriptor) AllOpCodes() []*OpCodeDescriptor {
	return utils.Map(utils.Keys(d.mnemonics), d.Descriptor)
}

// Number of opcodes implemented
func (d *OpCodesDescriptor) TotalOpCodes() int {
	return len(d.mnemonics)
}

// Miminum number of bits required to binary encode an opcode
func (d *OpCodesDescriptor) OpCodeBits() int {
	return bits.Len(uint(d.TotalOpCodes() - 1))
}

// Decodes an opcode from its binary representation
func (d *OpCodesDescriptor) DecodeOpCode(binaryRepresentation uint64) (OpCode, error) {
	opCode := OpCode(binaryRepresentation)

	if opCode >= TOTAL_OPCODES {
		return 0, utils.MakeError(ErrInvalidOpCode, "%v (hex: %v, bin: %v)", utils.FormatUintBinary(binaryRepresentation, 64), utils.FormatUintHex(binaryRepresentation, 16))
	}

	return opCode, nil
}

// Encodes an opcode into its binary representation
func (d *OpCodesDescriptor) EncodeOpCode(op OpCode) uint64 {
	return uint64(op)
}

// Returns the mnemonic string representation of the opcode
func (d *OpCodesDescriptor) Mnemonic(op OpCode) string {
	return d.mnemonics[op]
}

var ErrInvalidOpCode error = errors.New("invalid instruction opcode")

// Returns the opcode corresponding to the given mnemonic
func (d *OpCodesDescriptor) ParseOpCode(mnemonic string) (OpCode, error) {
	if opcode, hasOpCode := d.mnemonicsToOpCode[strings.ToUpper(mnemonic)]; hasOpCode {
		return opcode, nil
	} else {
		return 0, utils.MakeError(ErrInvalidOpCode, "'%v'", mnemonic)
	}
}

func makeOpCodesDescriptor(mnemonics map[OpCode]string) OpCodesDescriptor {
	for i, opCode := range utils.Iota(int(TOTAL_OPCODES), func(i int) OpCode { return OpCode(i) }) {
		if _, hasOpCode := mnemonics[opCode]; !hasOpCode {
			panic(fmt.Sprintf("missing entry for opcode %v in mnemonics table. Make sure you've added all OpCode -> Mnemonic entries in the makeOpCodesDescriptor() call", i))
		}
	}

	d := OpCodesDescriptor{
		mnemonics:         mnemonics,
		mnemonicsToOpCode: utils.InvertedMap(mnemonics),
	}

	if d.TotalOpCodes() != int(TOTAL_OPCODES) {
		panic("missing entry in opcode mnemonics table??? Make sure you've added all OpCode -> Mnemonic entries in the makeOpCodesDescriptor() call")
	}
	return d
}

var Descriptor_Opcodes OpCodesDescriptor = makeOpCodesDescriptor(
	map[OpCode]string{
		OpCode_NOP: "NOP",
		OpCode_IMM: "IMM",
		OpCode_MOV: "MOV",
		OpCode_LD:  "LD",
		OpCode_ST:  "ST",
		OpCode_ADD: "ADD",
		OpCode_SUB: "SUB",
		OpCode_MUL: "MUL",
		OpCode_DIV: "DIV",
		OpCode_MOD: "MOD",
	},
)

// Returns the mnemonic of the instruction opcode
func (op OpCode) String() string {
	return Descriptor_Opcodes.Mnemonic(op)
}
