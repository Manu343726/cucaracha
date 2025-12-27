package instructions

import (
	"testing"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/registers"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/types"
	"github.com/stretchr/testify/assert"
)

func TestLLVMInstructionOperandDescriptor_Bits(t *testing.T) {
	operand := &OperandDescriptor{EncodingBits: 8}
	desc := &LLVMInstructionOperandDescriptor{Operand: operand}
	assert.Equal(t, 8, desc.Bits())
}

func TestLLVMInstructionOperandDescriptor_MostSignificantBit(t *testing.T) {
	operand := &OperandDescriptor{EncodingBits: 8, EncodingPosition: 4}
	desc := &LLVMInstructionOperandDescriptor{Operand: operand}
	assert.Equal(t, 11, desc.MostSignificantBit())
}

func TestLLVMInstructionOperandDescriptor_LeastSignificantBit(t *testing.T) {
	operand := &OperandDescriptor{EncodingBits: 8, EncodingPosition: 4}
	desc := &LLVMInstructionOperandDescriptor{Operand: operand}
	assert.Equal(t, 4, desc.LeastSignificantBit())
}

func TestLLVMInstructionOperandDescriptor_Param(t *testing.T) {
	desc := &LLVMInstructionOperandDescriptor{Name: "op"}
	assert.Equal(t, "$op", desc.Param())
}

func TestLLVMInstructionOperandDescriptor_ParamDeclaration(t *testing.T) {
	desc := &LLVMInstructionOperandDescriptor{Type: "i32", Name: "op"}
	assert.Equal(t, "i32:$op", desc.ParamDeclaration())
}

func TestLLVMInstructionOperandDescriptor_ParamPattern(t *testing.T) {
	desc := &LLVMInstructionOperandDescriptor{Pattern: "pat", Name: "op"}
	assert.Equal(t, "pat:$op", desc.ParamPattern())
}

func TestLLVMInstructionFlags_FlagsSet(t *testing.T) {
	flags := LLVMInstructionFlags_IsReturn | LLVMInstructionFlags_IsBranch
	expected := []LLVMInstructionFlags{LLVMInstructionFlags_IsReturn, LLVMInstructionFlags_IsBranch}
	assert.ElementsMatch(t, expected, flags.FlagsSet())
}

func TestLLVMInstructionFlags_Flags(t *testing.T) {
	flags := LLVMInstructionFlags_IsReturn | LLVMInstructionFlags_IsBranch
	expected := map[LLVMInstructionFlags]bool{
		LLVMInstructionFlags_IsReturn:  true,
		LLVMInstructionFlags_IsBranch:  true,
		LLVMInstructionFlags_IsCompare: false,
		LLVMInstructionFlags_IsMoveImm: false,
		LLVMInstructionFlags_IsMoveReg: false,
		LLVMInstructionFlags_IsBitcast: false,
		LLVMInstructionFlags_IsCall:    false,
		LLVMInstructionFlags_MayLoad:   false,
		LLVMInstructionFlags_MayStore:  false,
		LLVMInstructionFlags_IsPseudo:  false,
	}
	assert.Equal(t, expected, flags.Flags())
}

func TestLLVMInstructionFlags_String(t *testing.T) {
	flags := LLVMInstructionFlags_IsReturn | LLVMInstructionFlags_IsBranch
	assert.Equal(t, "isReturn|isBranch", flags.String())
}

func TestLLVMType(t *testing.T) {
	assert.Equal(t, "i32", LLVMType(types.ValueType_Int32))
}

func TestLLVMImmediateType(t *testing.T) {
	assert.Equal(t, "i32imm", LLVMImmediateType(types.ValueType_Int32))
}

func TestLLVMOperandType(t *testing.T) {
	operand := &OperandDescriptor{Kind: OperandKind_Immediate, ValueType: types.ValueType_Int32}
	assert.Equal(t, "i32imm", LLVMOperandType(operand))
}

func TestLLVMOperandPattern(t *testing.T) {
	operand := &OperandDescriptor{ValueType: types.ValueType_Int32}
	assert.Equal(t, "i32", LLVMOperandPattern(operand))
}

func TestNewLLVMInstructionDescriptor(t *testing.T) {
	instr := &InstructionDescriptor{
		Operands: []*OperandDescriptor{
			{Role: OperandRole_Source, ValueType: types.ValueType_Int32},
			{Role: OperandRole_Destination, ValueType: types.ValueType_Int32},
		},
		LLVM_PatternTemplate: "template",
	}
	desc := NewLLVMInstructionDescriptor(instr)
	assert.Equal(t, "template", desc.Pattern)
	assert.Len(t, desc.Operands, 2)
	assert.Equal(t, "src", desc.Operands[0].Name)
	assert.Equal(t, "dst", desc.Operands[1].Name)
}

func TestLLVMInstructionDescriptor_Ins(t *testing.T) {
	instr := &InstructionDescriptor{
		Operands: []*OperandDescriptor{
			{Role: OperandRole_Source, ValueType: types.ValueType_Int32, Kind: OperandKind_Register, RegisterMetaClass: registers.RegisterMetaClasses[0]},
			{Role: OperandRole_Destination, ValueType: types.ValueType_Int32, Kind: OperandKind_Register, RegisterMetaClass: registers.RegisterMetaClasses[0]},
		},
	}
	desc := NewLLVMInstructionDescriptor(instr)
	assert.Equal(t, []string{"IntegerRegisters:$src"}, desc.Ins())
}

func TestLLVMInstructionDescriptor_Outs(t *testing.T) {
	instr := &InstructionDescriptor{
		Operands: []*OperandDescriptor{
			{Role: OperandRole_Source, ValueType: types.ValueType_Int32, Kind: OperandKind_Register, RegisterMetaClass: registers.RegisterMetaClasses[0]},
			{Role: OperandRole_Destination, ValueType: types.ValueType_Int32, Kind: OperandKind_Register, RegisterMetaClass: registers.RegisterMetaClasses[0]},
		},
	}
	desc := NewLLVMInstructionDescriptor(instr)
	assert.Equal(t, []string{"IntegerRegisters:$dst"}, desc.Outs())
}

func TestLLVMInstructionDescriptor_Ins_Immediate(t *testing.T) {
	instr := &InstructionDescriptor{
		Operands: []*OperandDescriptor{
			{Role: OperandRole_Source, ValueType: types.ValueType_Int32, Kind: OperandKind_Immediate},
			{Role: OperandRole_Destination, ValueType: types.ValueType_Int32, Kind: OperandKind_Register, RegisterMetaClass: registers.RegisterMetaClasses[0]},
		},
	}
	desc := NewLLVMInstructionDescriptor(instr)
	assert.Equal(t, []string{"i32imm:$src"}, desc.Ins())
}

func TestLLVMInstructionDescriptor_Outs_Immediate(t *testing.T) {
	instr := &InstructionDescriptor{
		Operands: []*OperandDescriptor{
			{Role: OperandRole_Source, ValueType: types.ValueType_Int32, Kind: OperandKind_Immediate},
			{Role: OperandRole_Destination, ValueType: types.ValueType_Int32, Kind: OperandKind_Register, RegisterMetaClass: registers.RegisterMetaClasses[0]},
		},
	}
	desc := NewLLVMInstructionDescriptor(instr)
	assert.Equal(t, []string{"IntegerRegisters:$dst"}, desc.Outs())
}

func TestLLVMInstructionDescriptor_Params(t *testing.T) {
	instr := &InstructionDescriptor{
		Operands: []*OperandDescriptor{
			{Role: OperandRole_Source, ValueType: types.ValueType_Int32, Kind: OperandKind_Register, RegisterMetaClass: registers.RegisterMetaClasses[0]},
			{Role: OperandRole_Destination, ValueType: types.ValueType_Int32, Kind: OperandKind_Register, RegisterMetaClass: registers.RegisterMetaClasses[0]},
		},
	}
	desc := NewLLVMInstructionDescriptor(instr)
	assert.Equal(t, []string{"$src", "$dst"}, desc.Params())
}

func TestLLVMInstructionDescriptor_Params_MultipleOperands(t *testing.T) {
	instr := &InstructionDescriptor{
		Operands: []*OperandDescriptor{
			{Role: OperandRole_Source, ValueType: types.ValueType_Int32},
			{Role: OperandRole_Source, ValueType: types.ValueType_Int32},
			{Role: OperandRole_Destination, ValueType: types.ValueType_Int32},
		},
	}
	desc := NewLLVMInstructionDescriptor(instr)
	assert.Equal(t, []string{"$src1", "$src2", "$dst"}, desc.Params())
}

func TestLLVMInstructionDescriptor_Ins_MultipleOperands(t *testing.T) {
	instr := &InstructionDescriptor{
		Operands: []*OperandDescriptor{
			{Role: OperandRole_Source, ValueType: types.ValueType_Int32},
			{Role: OperandRole_Source, ValueType: types.ValueType_Int32},
			{Role: OperandRole_Destination, ValueType: types.ValueType_Int32},
		},
	}
	desc := NewLLVMInstructionDescriptor(instr)
	assert.Equal(t, []string{"i32imm:$src1", "i32imm:$src2"}, desc.Ins())
}

func TestLLVMInstructionDescriptor_Outs_MultipleOperands(t *testing.T) {
	instr := &InstructionDescriptor{
		Operands: []*OperandDescriptor{
			{Role: OperandRole_Source, ValueType: types.ValueType_Int32},
			{Role: OperandRole_Source, ValueType: types.ValueType_Int32},
			{Role: OperandRole_Destination, ValueType: types.ValueType_Int32},
		},
	}
	desc := NewLLVMInstructionDescriptor(instr)
	assert.Equal(t, []string{"i32imm:$dst"}, desc.Outs())
}
