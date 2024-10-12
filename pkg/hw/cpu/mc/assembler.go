package mc

import (
	"strings"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/Manu343726/cucaracha/pkg/utils"
)

// Parses an instruction of the form 'mnemonic [operands, ...]'
//
// The parser is as stupid as it gets, the commas separating operands are
// optional, operands are really splitted through the whitespaces between them
// (So multi-arg operands are not supported)
func ParseInstruction(instrString string) (*instructions.Instruction, error) {
	sanitizedStr := strings.TrimSpace(instrString)
	splitted := strings.Split(sanitizedStr, " ")

	if len(splitted) <= 1 {
		return nil, utils.MakeError(instructions.ErrInvalidInstruction, "instruction '%s' invalid. Must have at least an opcode", sanitizedStr)
	}

	opCodeMnemonic := &splitted[0]

	opCode, err := instructions.Opcodes.ParseOpCode(*opCodeMnemonic)

	if err != nil {
		return nil, err
	}

	var result instructions.Instruction
	result.Descriptor, err = instructions.Instructions.Instruction(opCode)

	if err != nil {
		return nil, err
	}

	operands := splitted[1:]

	// Remove the commas between operands
	for i := range operands {
		if withoutComma, hasComma := strings.CutSuffix(operands[i], ","); hasComma {
			operands[i] = withoutComma
		}
	}

	if len(operands) != len(result.Descriptor.Operands) {
		return nil, utils.MakeError(instructions.ErrInvalidInstruction, "'%v': expected %v operands for %v instruction, got %v", sanitizedStr, len(result.Descriptor.Operands), opCode, len(operands))
	}

	result.OperandValues = make([]instructions.OperandValue, len(result.Descriptor.Operands))

	for i, operandDescriptor := range result.Descriptor.Operands {
		value, err := operandDescriptor.ParseValue(operands[i])

		if err != nil {
			return nil, utils.MakeError(instructions.ErrInvalidInstruction, "error parsing operand [%v] '%v': %w", i, operands[i], err)
		}

		result.OperandValues[i] = value
	}

	return &result, nil
}
