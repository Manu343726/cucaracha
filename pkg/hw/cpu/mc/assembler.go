package mc

import (
	"os"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/Manu343726/cucaracha/pkg/utils"
)

// Comment string for assembly files
const Comment string = "//"

// Parses an instruction of the form 'mnemonic [operands, ...]'
//
// The parser is as stupid as it gets, the commas separating operands are
// optional, operands are really splitted through the whitespaces between them
// (So multi-arg operands are not supported), and anything preceded by a // is
// considered a comment and ignored.
func ParseInstruction(instrString string) (*instructions.Instruction, error) {
	sanitizedStr := strings.TrimSpace(instrString)
	splitted := strings.Split(sanitizedStr, " ")

	if len(splitted) <= 1 {
		return nil, utils.MakeError(instructions.ErrInvalidInstruction, "instruction '%s' invalid. Must have at least an opcode", sanitizedStr)
	}

	if strings.HasPrefix(splitted[0], Comment) {
		return nil, nil
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

// Represents a parsed assembly instruction
type AssemblyInstruction struct {
	Instruction *instructions.Instruction
	Line        int
	File        *string
}

func ParseAssembly(assembly string) ([]AssemblyInstruction, error) {
	lines := strings.Split(assembly, "\n")
	result := make([]AssemblyInstruction, 0, len(lines))

	for i, line := range lines {
		instruction, err := ParseInstruction(line)

		if err != nil {
			return nil, utils.MakeError(err, "error parsing instruction at line %v: %v", i, line)
		}

		if instruction != nil {
			result = append(result, AssemblyInstruction{
				Instruction: instruction,
				Line:        i,
				File:        nil,
			})
		}
	}

	return result, nil
}

func ParseAssemblyFile(path *string) ([]AssemblyInstruction, error) {
	assembly, err := os.ReadFile(*path)
	if err != nil {
		return nil, err
	}

	instructions, err := ParseAssembly(string(assembly))
	if err != nil {
		return nil, utils.MakeError(err, "error parsing assembly file '%v'", *path)
	}

	for i := range instructions {
		instructions[i].File = path
	}

	return instructions, nil
}
