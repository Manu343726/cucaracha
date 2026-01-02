package mc

import (
	"testing"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReg(t *testing.T) {
	t.Run("valid register", func(t *testing.T) {
		op, err := Reg("r0")
		require.NoError(t, err)
		assert.Equal(t, instructions.OperandKind_Register, op.Kind())
		assert.Equal(t, "r0", op.String())
	})

	t.Run("invalid register", func(t *testing.T) {
		_, err := Reg("invalid")
		assert.Error(t, err)
	})
}

func TestImm(t *testing.T) {
	t.Run("positive value", func(t *testing.T) {
		op := Imm(42)
		assert.Equal(t, instructions.OperandKind_Immediate, op.Kind())
	})

	t.Run("negative value", func(t *testing.T) {
		op := Imm(-42)
		assert.Equal(t, instructions.OperandKind_Immediate, op.Kind())
	})

	t.Run("zero", func(t *testing.T) {
		op := Imm(0)
		assert.Equal(t, instructions.OperandKind_Immediate, op.Kind())
	})
}

func TestUImm(t *testing.T) {
	op := UImm(0xFFFF)
	assert.Equal(t, instructions.OperandKind_Immediate, op.Kind())
}

func TestImm16(t *testing.T) {
	op := Imm16(0x1234)
	assert.Equal(t, instructions.OperandKind_Immediate, op.Kind())
}

func TestNewProgram(t *testing.T) {
	p := NewProgram()
	assert.NotNil(t, p)
	assert.Equal(t, 0, p.Len())
}

func TestProgram_Add(t *testing.T) {
	p := NewProgram()
	instr := Nop()

	result := p.Add(instr)

	assert.Same(t, p, result) // Fluent interface returns same pointer
	assert.Equal(t, 1, p.Len())
	assert.Same(t, instr, p.At(0))
}

func TestProgram_Len(t *testing.T) {
	p := NewProgram()
	assert.Equal(t, 0, p.Len())

	p.Add(Nop())
	assert.Equal(t, 1, p.Len())

	p.Add(Nop())
	assert.Equal(t, 2, p.Len())
}

func TestProgram_At(t *testing.T) {
	p := NewProgram()
	instr1 := Nop()
	instr2 := Add("r0", "r1", "r2")

	p.Add(instr1).Add(instr2)

	assert.Same(t, instr1, p.At(0))
	assert.Same(t, instr2, p.At(1))
}

func TestProgram_String(t *testing.T) {
	p := NewProgram()
	p.Add(Nop())

	str := p.String()

	assert.Contains(t, str, "NOP") // Mnemonics are uppercase
}

func TestProgram_Encode(t *testing.T) {
	p := NewProgram()
	p.Add(Nop())

	encoded, err := p.Encode()

	require.NoError(t, err)
	assert.Equal(t, 4, len(encoded)) // 4 bytes per instruction
}

func TestProgram_Encode_MultipleInstructions(t *testing.T) {
	p := NewProgram()
	p.Add(Nop())
	p.Add(Nop())

	encoded, err := p.Encode()

	require.NoError(t, err)
	assert.Equal(t, 8, len(encoded)) // 4 bytes per instruction * 2
}

func TestInstr(t *testing.T) {
	t.Run("valid opcode", func(t *testing.T) {
		builder := Instr(instructions.OpCode_NOP)
		assert.NotNil(t, builder)

		instr, err := builder.Build()
		require.NoError(t, err)
		assert.NotNil(t, instr)
	})

	t.Run("build with operands", func(t *testing.T) {
		builder := Instr(instructions.OpCode_ADD).R(0).R(1).R(2)
		instr, err := builder.Build()
		require.NoError(t, err)
		assert.NotNil(t, instr)
	})
}

func TestInstrByName(t *testing.T) {
	t.Run("valid mnemonic", func(t *testing.T) {
		builder := InstrByName("NOP") // Mnemonics are uppercase
		instr, err := builder.Build()
		require.NoError(t, err)
		assert.NotNil(t, instr)
	})

	t.Run("invalid mnemonic", func(t *testing.T) {
		builder := InstrByName("invalid_instruction")
		_, err := builder.Build()
		assert.Error(t, err)
	})
}

func TestInstructionBuilder_Op(t *testing.T) {
	op := Imm(42)
	builder := Instr(instructions.OpCode_MOV_IMM16L).Op(op).NamedR("r0")

	instr, err := builder.Build()

	require.NoError(t, err)
	assert.NotNil(t, instr)
}

func TestInstructionBuilder_NamedR(t *testing.T) {
	t.Run("valid register", func(t *testing.T) {
		builder := Instr(instructions.OpCode_MOV).NamedR("r0").NamedR("r1")
		instr, err := builder.Build()
		require.NoError(t, err)
		assert.NotNil(t, instr)
	})

	t.Run("invalid register", func(t *testing.T) {
		builder := Instr(instructions.OpCode_MOV).NamedR("invalid").NamedR("r1")
		_, err := builder.Build()
		assert.Error(t, err)
	})
}

func TestInstructionBuilder_R(t *testing.T) {
	t.Run("valid register index", func(t *testing.T) {
		builder := Instr(instructions.OpCode_MOV).R(0).R(1)
		instr, err := builder.Build()
		require.NoError(t, err)
		assert.NotNil(t, instr)
	})

	t.Run("invalid register index", func(t *testing.T) {
		builder := Instr(instructions.OpCode_MOV).R(999).R(1)
		_, err := builder.Build()
		assert.Error(t, err)
	})
}

func TestInstructionBuilder_I(t *testing.T) {
	builder := Instr(instructions.OpCode_MOV_IMM16L).I(42).NamedR("r0")
	instr, err := builder.Build()
	require.NoError(t, err)
	assert.NotNil(t, instr)
}

func TestInstructionBuilder_U(t *testing.T) {
	builder := Instr(instructions.OpCode_MOV_IMM16L).U(42).NamedR("r0")
	instr, err := builder.Build()
	require.NoError(t, err)
	assert.NotNil(t, instr)
}

func TestInstructionBuilder_MustBuild(t *testing.T) {
	t.Run("valid instruction", func(t *testing.T) {
		instr := Instr(instructions.OpCode_NOP).MustBuild()
		assert.NotNil(t, instr)
	})

	t.Run("invalid instruction panics", func(t *testing.T) {
		assert.Panics(t, func() {
			InstrByName("invalid").MustBuild()
		})
	})
}

// Test convenience functions

func TestNop(t *testing.T) {
	instr := Nop()
	assert.NotNil(t, instr)
	assert.Equal(t, instructions.OpCode_NOP, instr.Descriptor.OpCode.OpCode)
}

func TestMov(t *testing.T) {
	instr := Mov("r0", "r1")
	assert.NotNil(t, instr)
	assert.Equal(t, instructions.OpCode_MOV, instr.Descriptor.OpCode.OpCode)
}

func TestMovImm16L(t *testing.T) {
	instr := MovImm16L(0x1234, "r0")
	assert.NotNil(t, instr)
	assert.Equal(t, instructions.OpCode_MOV_IMM16L, instr.Descriptor.OpCode.OpCode)
}

func TestMovImm16H(t *testing.T) {
	instr := MovImm16H(0x5678, "r0")
	assert.NotNil(t, instr)
	assert.Equal(t, instructions.OpCode_MOV_IMM16H, instr.Descriptor.OpCode.OpCode)
}

func TestLd(t *testing.T) {
	instr := Ld("r0", "r1")
	assert.NotNil(t, instr)
	assert.Equal(t, instructions.OpCode_LD, instr.Descriptor.OpCode.OpCode)
}

func TestSt(t *testing.T) {
	instr := St("r0", "r1")
	assert.NotNil(t, instr)
	assert.Equal(t, instructions.OpCode_ST, instr.Descriptor.OpCode.OpCode)
}

func TestAdd(t *testing.T) {
	instr := Add("r0", "r1", "r2")
	assert.NotNil(t, instr)
	assert.Equal(t, instructions.OpCode_ADD, instr.Descriptor.OpCode.OpCode)
}

func TestSub(t *testing.T) {
	instr := Sub("r0", "r1", "r2")
	assert.NotNil(t, instr)
	assert.Equal(t, instructions.OpCode_SUB, instr.Descriptor.OpCode.OpCode)
}

func TestMul(t *testing.T) {
	instr := Mul("r0", "r1", "r2")
	assert.NotNil(t, instr)
	assert.Equal(t, instructions.OpCode_MUL, instr.Descriptor.OpCode.OpCode)
}

func TestDiv(t *testing.T) {
	instr := Div("r0", "r1", "r2")
	assert.NotNil(t, instr)
	assert.Equal(t, instructions.OpCode_DIV, instr.Descriptor.OpCode.OpCode)
}

func TestMod(t *testing.T) {
	instr := Mod("r0", "r1", "r2")
	assert.NotNil(t, instr)
	assert.Equal(t, instructions.OpCode_MOD, instr.Descriptor.OpCode.OpCode)
}

func TestCmp(t *testing.T) {
	instr := Cmp("r0", "r1", "r2")
	assert.NotNil(t, instr)
	assert.Equal(t, instructions.OpCode_CMP, instr.Descriptor.OpCode.OpCode)
}

func TestJmp(t *testing.T) {
	instr := Jmp("r0", "r5")
	assert.NotNil(t, instr)
	assert.Equal(t, instructions.OpCode_JMP, instr.Descriptor.OpCode.OpCode)
}

func TestCJmp(t *testing.T) {
	instr := CJmp("r0", "r1", "r5")
	assert.NotNil(t, instr)
	assert.Equal(t, instructions.OpCode_CJMP, instr.Descriptor.OpCode.OpCode)
}

func TestLsl(t *testing.T) {
	instr := Lsl("r0", "r1", "r2")
	assert.NotNil(t, instr)
	assert.Equal(t, instructions.OpCode_LSL, instr.Descriptor.OpCode.OpCode)
}

func TestLsr(t *testing.T) {
	instr := Lsr("r0", "r1", "r2")
	assert.NotNil(t, instr)
	assert.Equal(t, instructions.OpCode_LSR, instr.Descriptor.OpCode.OpCode)
}

func TestAsl(t *testing.T) {
	instr := Asl("r0", "r1", "r2")
	assert.NotNil(t, instr)
	assert.Equal(t, instructions.OpCode_ASL, instr.Descriptor.OpCode.OpCode)
}

func TestAsr(t *testing.T) {
	instr := Asr("r0", "r1", "r2")
	assert.NotNil(t, instr)
	assert.Equal(t, instructions.OpCode_ASR, instr.Descriptor.OpCode.OpCode)
}

// Integration tests

func TestProgram_FullExample(t *testing.T) {
	// Create a simple program that adds two numbers
	p := NewProgram()
	p.Add(MovImm16L(10, "r0")). // r0 = 10
					Add(MovImm16L(20, "r1")).   // r1 = 20
					Add(Add("r0", "r1", "r2")). // r2 = r0 + r1
					Add(Nop())

	assert.Equal(t, 4, p.Len())

	// Verify encoding works
	encoded, err := p.Encode()
	require.NoError(t, err)
	assert.Equal(t, 16, len(encoded)) // 4 instructions * 4 bytes
}

func TestProgram_FluentInterface(t *testing.T) {
	// Test that the fluent interface works correctly
	p := NewProgram().
		Add(Nop()).
		Add(Nop()).
		Add(Nop())

	assert.Equal(t, 3, p.Len())
}

func TestInstructionBuilder_ErrorPropagation(t *testing.T) {
	// Test that errors propagate through the builder chain
	builder := Instr(instructions.OpCode_MOV).
		NamedR("invalid_register"). // This should cause an error
		NamedR("r1")                // This should not clear the error

	_, err := builder.Build()
	assert.Error(t, err)
}
