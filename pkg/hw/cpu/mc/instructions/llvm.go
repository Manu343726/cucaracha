package instructions

import (
	"fmt"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/types"
	"github.com/Manu343726/cucaracha/pkg/utils"
)

// Contains metadata to generate correct operand descriptions for LLVM instruction definition
type LLVMInstructionOperandDescriptor struct {
	// Operand descriptor
	Operand *OperandDescriptor
	// Operand type
	Type string
	// Operand name
	Name string
	// Operand pattern in the DAG selection pattern of the instruction
	Pattern string
}

// Returns the number of bits used to encode the operand
func (d *LLVMInstructionOperandDescriptor) Bits() int {
	return d.Operand.EncodingBits
}

// Returns the most significant bit within the instruction used to encode the operand
func (d *LLVMInstructionOperandDescriptor) MostSignificantBit() int {
	return d.Operand.EncodingPosition + d.Operand.EncodingBits - 1
}

// Returns the least significant bit within the instruction used to encode the operand
func (d *LLVMInstructionOperandDescriptor) LeastSignificantBit() int {
	return d.Operand.EncodingPosition
}

// Returns the operand paramter string, that is, the operand name prefixed with a $
func (d *LLVMInstructionOperandDescriptor) Param() string {
	return "$" + d.Name
}

// Returns the full operand paramter declaration, that is, the operand type followed by the operand name prefixed with a $
func (d *LLVMInstructionOperandDescriptor) ParamDeclaration() string {
	return fmt.Sprintf("%v:$%v", d.Type, d.Name)
}

// Returns the full operand pattern declaration, that is, the operand pattern followed by the operand name prefixed with a $
func (d *LLVMInstructionOperandDescriptor) ParamPattern() string {
	return fmt.Sprintf("%v:$%v", d.Pattern, d.Name)
}

// Represents flags in the LLVM tablegen Instruction class that control the high level semantics of an instruction
// See class Instruction in LLVM's source code llvm/include/llvm/Target/Target.td for details
type LLVMInstructionFlags uint

const (
	LLVMInstructionFlags_IsReturn LLVMInstructionFlags = 1 << iota
	LLVMInstructionFlags_IsBranch
	LLVMInstructionFlags_IsCompare
	LLVMInstructionFlags_IsMoveImm
	LLVMInstructionFlags_IsMoveReg
	LLVMInstructionFlags_IsBitcast
	LLVMInstructionFlags_IsCall
	LLVMInstructionFlags_MayLoad
	LLVMInstructionFlags_MayStore
	LLVMInstructionFlags_IsPseudo
	LLVMInstructionFlags_IsBarrier
	LLVMInstructionFlags_IsTerminator
	LLVMInstructionFlags_UsesCustomInserter
	LLVMInstructionFlags_IsIndirectBranch
)

var llvmInstructionFlagsToString map[LLVMInstructionFlags]string = map[LLVMInstructionFlags]string{
	LLVMInstructionFlags_IsReturn:           "isReturn",
	LLVMInstructionFlags_IsBranch:           "isBranch",
	LLVMInstructionFlags_IsCompare:          "isCompare",
	LLVMInstructionFlags_IsMoveImm:          "isMovImm",
	LLVMInstructionFlags_IsMoveReg:          "isMovReg",
	LLVMInstructionFlags_IsBitcast:          "isBitcast",
	LLVMInstructionFlags_IsCall:             "isCall",
	LLVMInstructionFlags_MayLoad:            "mayLoad",
	LLVMInstructionFlags_MayStore:           "mayStore",
	LLVMInstructionFlags_IsPseudo:           "isPseudo",
	LLVMInstructionFlags_IsBarrier:          "isBarrier",
	LLVMInstructionFlags_IsTerminator:       "isTerminator",
	LLVMInstructionFlags_UsesCustomInserter: "usesCustomInserter",
	LLVMInstructionFlags_IsIndirectBranch:   "isIndirectBranch",
}

// Returns all the flags enabled
func (f LLVMInstructionFlags) FlagsSet() []LLVMInstructionFlags {
	result := make([]LLVMInstructionFlags, 0, len(llvmInstructionFlagsToString))

	for i := 0; i < len(llvmInstructionFlagsToString); i++ {
		flag := LLVMInstructionFlags(1 << i)

		if (f & flag) != 0 {
			result = append(result, flag)
		}
	}

	return result
}

// Returns all flags with a corresponding boolean stating if the flag is set or not
func (f LLVMInstructionFlags) Flags() map[LLVMInstructionFlags]bool {
	result := make(map[LLVMInstructionFlags]bool, len(llvmInstructionFlagsToString))

	for i := 0; i < len(llvmInstructionFlagsToString); i++ {
		flag := LLVMInstructionFlags(1 << i)

		result[flag] = (f & flag) != 0
	}

	return result
}

func (f LLVMInstructionFlags) String() string {
	if str, hasFlag := llvmInstructionFlagsToString[f]; hasFlag {
		return str
	} else {
		return strings.Join(utils.Map(f.FlagsSet(), LLVMInstructionFlags.String), "|")
	}
}

// Contains metadata to generate LLVM instruction definitions
type LLVMInstructionDescriptor struct {
	Instruction *InstructionDescriptor
	Operands    []*LLVMInstructionOperandDescriptor
	Pattern     string
	Flags       map[LLVMInstructionFlags]bool
	// LLVM operand constraints (e.g. "$dst = $src" for tied operands)
	Constraints string
}

// Returns the set of input operands declarations
func (d *LLVMInstructionDescriptor) Ins() []string {
	return utils.Map(
		utils.Filter(d.Operands,
			func(op *LLVMInstructionOperandDescriptor) bool { return op.Operand.Role == OperandRole_Source }),
		(*LLVMInstructionOperandDescriptor).ParamDeclaration)
}

// Returns the set of output operands declarations
func (d *LLVMInstructionDescriptor) Outs() []string {
	return utils.Map(
		utils.Filter(d.Operands,
			func(op *LLVMInstructionOperandDescriptor) bool { return op.Operand.Role == OperandRole_Destination }),
		(*LLVMInstructionOperandDescriptor).ParamDeclaration)
}

// Returns all instruction operands in order in parameter form, that is, a $ followed by the operand name
// Operands marked with LLVM_HideFromAsm are excluded from the assembly string
func (d *LLVMInstructionDescriptor) Params() []string {
	return utils.Map(
		utils.Filter(d.Operands,
			func(op *LLVMInstructionOperandDescriptor) bool { return !op.Operand.LLVM_HideFromAsm }),
		(*LLVMInstructionOperandDescriptor).Param)
}

// Returns the LLVM type name of a given value type
func LLVMType(valueType types.ValueType) string {
	switch valueType {
	case types.ValueType_Int32:
		return "i32"
	}

	panic("unsupported value type")
}

// Returns the LLVM type name of a given immediate value type
func LLVMImmediateType(valueType types.ValueType) string {
	return LLVMType(valueType) + "imm"
}

// Returns the operand type for a LLVM instruction definition
func LLVMOperandType(op *OperandDescriptor) string {
	if len(op.LLVM_CustomType) > 0 {
		return op.LLVM_CustomType
	}

	switch op.Kind {
	case OperandKind_Immediate:
		return LLVMImmediateType(op.ValueType)
	case OperandKind_Register:
		return op.RegisterMetaClass.Name
	}

	panic("unreachable")
}

// Returns the operand pattern for a LLVM instruction definition
func LLVMOperandPattern(op *OperandDescriptor) string {
	if len(op.LLVM_CustomPattern) > 0 {
		return op.LLVM_CustomPattern
	}

	return LLVMType(op.ValueType)
}

// Generates a descriptor for LLVM instruction generation from a cucaracha instruction descriptor
func NewLLVMInstructionDescriptor(i *InstructionDescriptor) *LLVMInstructionDescriptor {
	d := &LLVMInstructionDescriptor{
		Instruction: i,
		Operands:    make([]*LLVMInstructionOperandDescriptor, len(i.Operands)),
		Flags:       i.LLVM_InstructionFlags.Flags(),
		Constraints: i.LLVM_Constraints,
	}

	totalSources := 0
	totalDestinations := 0
	totalOperandsWithCustomName := 0

	for i, descriptor := range i.Operands {
		op := &LLVMInstructionOperandDescriptor{
			Operand: descriptor,
			Type:    LLVMOperandType(descriptor),
			Pattern: LLVMOperandPattern(descriptor),
		}

		if len(descriptor.LLVM_CustomName) > 0 {
			totalOperandsWithCustomName++
			op.Name = descriptor.LLVM_CustomName
		} else {
			switch descriptor.Role {
			case OperandRole_Source:
				totalSources++
			case OperandRole_Destination:
				totalDestinations++
			default:
				panic("operand role not implemented by LLVM target descriptor generator")
			}
		}

		d.Operands[i] = op
	}

	if totalOperandsWithCustomName > 0 {
		if totalOperandsWithCustomName != len(i.Operands) {
			panic("you can either use a custom name for all operands or let the system automatically name them for you, but not both")
		}
	} else {
		if (totalSources + totalDestinations) != len(i.Operands) {
			panic("unexpected number of operands")
		}

		sourceIndex := 1
		destinationIndex := 1

		for _, op := range d.Operands {
			switch op.Operand.Role {
			case OperandRole_Source:
				if totalSources > 1 {
					op.Name = fmt.Sprintf("src%v", sourceIndex)
				} else {
					op.Name = "src"
				}

				sourceIndex++
			case OperandRole_Destination:
				if totalDestinations > 1 {
					op.Name = fmt.Sprintf("dst%v", destinationIndex)
				} else {
					op.Name = "dst"
				}

				destinationIndex++
			}
		}
	}

	replaces := utils.ConcatMap(d.Operands, func(op *LLVMInstructionOperandDescriptor) []string {
		return []string{
			op.Param(),
			op.ParamPattern(),
		}
	})

	d.Pattern = strings.NewReplacer(replaces...).Replace(i.LLVM_PatternTemplate)

	return d
}
