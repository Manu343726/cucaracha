package instructions

import (
	"fmt"

	"github.com/Manu343726/cucaracha/pkg/utils"
)

// Contains implementation information of an instruction opcode
type OpCodeDescriptor struct {
	OpCode               OpCode
	BinaryRepresentation uint64
	Mnemonic             string
}

func (d *OpCodeDescriptor) String() string {
	return fmt.Sprintf("%v (code: %v, binary: %v, hex: %v)", d.Mnemonic, d.BinaryRepresentation, utils.FormatUintBinary(d.BinaryRepresentation, Opcodes.OpCodeBits()), utils.FormatUintHex(d.BinaryRepresentation, Opcodes.OpCodeBits()/4))
}

// Returns the number of bits used to encode an instruction opcode
func (d *OpCodeDescriptor) EncodingBits() int {
	return Opcodes.OpCodeBits()
}

// Returns the first bit within an instruction used to encode the opcode
func (d *OpCodeDescriptor) EncodingPosition() int {
	return 0
}

// Same as EncodingPosition()
func (d *OpCodeDescriptor) LeastSignificantBit() int {
	return 0
}

// Returns the last bit within an instruction used to encode the opcode
func (d *OpCodeDescriptor) MostSignificantBit() int {
	return d.LeastSignificantBit() + d.EncodingBits() - 1
}
