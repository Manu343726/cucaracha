package mc

import (
	"strings"
	"testing"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/registers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstructionResolver_FailsWithoutMemoryLayout(t *testing.T) {
	resolver := NewInstructionResolver()

	pf := &ProgramFileContents{
		FileNameValue:     "test.cucaracha",
		InstructionsValue: []Instruction{},
	}

	_, err := resolver.ResolveInstructions(pf)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnresolvedMemory)
}

func TestInstructionResolver_FailsWithoutInstructionAddresses(t *testing.T) {
	resolver := NewInstructionResolver()

	pf := &ProgramFileContents{
		FileNameValue: "test.cucaracha",
		InstructionsValue: []Instruction{
			{Text: "NOP"},
		},
		MemoryLayoutValue: &MemoryLayout{
			BaseAddress: 0x1000,
			CodeStart:   0x1000,
			CodeSize:    4,
		},
	}

	_, err := resolver.ResolveInstructions(pf)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnresolvedMemory)
	assert.Contains(t, err.Error(), "instruction 0")
}

func TestInstructionResolver_FailsWithoutGlobalAddresses(t *testing.T) {
	resolver := NewInstructionResolver()
	addr := uint32(0x1000)

	pf := &ProgramFileContents{
		FileNameValue: "test.cucaracha",
		InstructionsValue: []Instruction{
			{Text: "NOP", Address: &addr},
		},
		GlobalsValue: []Global{
			{Name: ".data", Size: 4}, // No address
		},
		MemoryLayoutValue: &MemoryLayout{
			BaseAddress: 0x1000,
			CodeStart:   0x1000,
			CodeSize:    4,
		},
	}

	_, err := resolver.ResolveInstructions(pf)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnresolvedMemory)
	assert.Contains(t, err.Error(), ".data")
}

func TestInstructionResolver_FailsWithUnresolvedSymbol(t *testing.T) {
	resolver := NewInstructionResolver()
	addr := uint32(0x1000)

	pf := &ProgramFileContents{
		FileNameValue: "test.cucaracha",
		InstructionsValue: []Instruction{
			{
				Text:    "MOVIMM16L .unknown@lo, r0",
				Address: &addr,
				Symbols: []SymbolReference{
					{Name: ".unknown", Usage: SymbolUsageLo}, // Not resolved
				},
			},
		},
		MemoryLayoutValue: &MemoryLayout{
			BaseAddress: 0x1000,
			CodeStart:   0x1000,
			CodeSize:    4,
		},
	}

	_, err := resolver.ResolveInstructions(pf)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnresolvedSymbol)
	assert.Contains(t, err.Error(), ".unknown")
}

func TestInstructionResolver_ResolvesFromText_NOP(t *testing.T) {
	resolver := NewInstructionResolver()
	addr := uint32(0x1000)

	pf := &ProgramFileContents{
		FileNameValue: "test.cucaracha",
		InstructionsValue: []Instruction{
			{Text: "NOP", Address: &addr},
		},
		MemoryLayoutValue: &MemoryLayout{
			BaseAddress: 0x1000,
			CodeStart:   0x1000,
			CodeSize:    4,
		},
	}

	result, err := resolver.ResolveInstructions(pf)
	require.NoError(t, err)

	instr := result.Instructions()[0]
	require.NotNil(t, instr.Raw)
	require.NotNil(t, instr.Instruction)

	// Verify Text
	assert.Equal(t, "NOP", instr.Text)

	// Verify Instruction descriptor
	assert.Equal(t, instructions.OpCode_NOP, instr.Instruction.Descriptor.OpCode.OpCode)
	assert.Equal(t, "NOP", instr.Instruction.Descriptor.OpCode.Mnemonic)
	assert.Len(t, instr.Instruction.OperandValues, 0, "NOP should have no operands")

	// Verify Raw instruction
	assert.Equal(t, instructions.OpCode_NOP, instr.Raw.Descriptor.OpCode.OpCode)
	assert.Len(t, instr.Raw.OperandValues, 0, "NOP raw should have no operand values")
}

func TestInstructionResolver_ResolvesFromText_ADD(t *testing.T) {
	resolver := NewInstructionResolver()
	addr := uint32(0x1000)

	pf := &ProgramFileContents{
		FileNameValue: "test.cucaracha",
		InstructionsValue: []Instruction{
			{Text: "ADD r0, r1, r2", Address: &addr},
		},
		MemoryLayoutValue: &MemoryLayout{
			BaseAddress: 0x1000,
			CodeStart:   0x1000,
			CodeSize:    4,
		},
	}

	result, err := resolver.ResolveInstructions(pf)
	require.NoError(t, err)

	instr := result.Instructions()[0]
	require.NotNil(t, instr.Raw)
	require.NotNil(t, instr.Instruction)

	// Verify Text
	assert.Equal(t, "ADD r0, r1, r2", instr.Text)

	// Verify Instruction descriptor
	assert.Equal(t, instructions.OpCode_ADD, instr.Instruction.Descriptor.OpCode.OpCode)
	assert.Equal(t, "ADD", instr.Instruction.Descriptor.OpCode.Mnemonic)
	require.Len(t, instr.Instruction.OperandValues, 3, "ADD should have 3 operands")

	// Verify operand values (r0, r1, r2)
	assert.Equal(t, instructions.OperandKind_Register, instr.Instruction.OperandValues[0].Kind())
	assert.Equal(t, instructions.OperandKind_Register, instr.Instruction.OperandValues[1].Kind())
	assert.Equal(t, instructions.OperandKind_Register, instr.Instruction.OperandValues[2].Kind())
	assert.Equal(t, "r0", instr.Instruction.OperandValues[0].Register().Name())
	assert.Equal(t, "r1", instr.Instruction.OperandValues[1].Register().Name())
	assert.Equal(t, "r2", instr.Instruction.OperandValues[2].Register().Name())

	// Verify Raw instruction operand values (encoded register values)
	assert.Equal(t, instructions.OpCode_ADD, instr.Raw.Descriptor.OpCode.OpCode)
	require.Len(t, instr.Raw.OperandValues, 3)
	assert.Equal(t, registers.Register("r0").Encode(), instr.Raw.OperandValues[0], "raw operand 0 should match r0 encoding")
	assert.Equal(t, registers.Register("r1").Encode(), instr.Raw.OperandValues[1], "raw operand 1 should match r1 encoding")
	assert.Equal(t, registers.Register("r2").Encode(), instr.Raw.OperandValues[2], "raw operand 2 should match r2 encoding")
}

func TestInstructionResolver_ResolvesFromText_MOV(t *testing.T) {
	resolver := NewInstructionResolver()
	addr := uint32(0x1000)

	pf := &ProgramFileContents{
		FileNameValue: "test.cucaracha",
		InstructionsValue: []Instruction{
			{Text: "MOV r0, r1", Address: &addr},
		},
		MemoryLayoutValue: &MemoryLayout{
			BaseAddress: 0x1000,
			CodeStart:   0x1000,
			CodeSize:    4,
		},
	}

	result, err := resolver.ResolveInstructions(pf)
	require.NoError(t, err)

	instr := result.Instructions()[0]
	require.NotNil(t, instr.Raw)
	require.NotNil(t, instr.Instruction)

	// Verify Text
	assert.Equal(t, "MOV r0, r1", instr.Text)

	// Verify Instruction descriptor
	assert.Equal(t, instructions.OpCode_MOV, instr.Instruction.Descriptor.OpCode.OpCode)
	assert.Equal(t, "MOV", instr.Instruction.Descriptor.OpCode.Mnemonic)
	require.Len(t, instr.Instruction.OperandValues, 2, "MOV should have 2 operands")

	// Verify operand values (r0, r1)
	assert.Equal(t, instructions.OperandKind_Register, instr.Instruction.OperandValues[0].Kind())
	assert.Equal(t, instructions.OperandKind_Register, instr.Instruction.OperandValues[1].Kind())
	assert.Equal(t, "r0", instr.Instruction.OperandValues[0].Register().Name())
	assert.Equal(t, "r1", instr.Instruction.OperandValues[1].Register().Name())

	// Verify Raw instruction
	assert.Equal(t, instructions.OpCode_MOV, instr.Raw.Descriptor.OpCode.OpCode)
	require.Len(t, instr.Raw.OperandValues, 2)
	assert.Equal(t, registers.Register("r0").Encode(), instr.Raw.OperandValues[0], "raw operand 0 should match r0 encoding")
	assert.Equal(t, registers.Register("r1").Encode(), instr.Raw.OperandValues[1], "raw operand 1 should match r1 encoding")
}

func TestInstructionResolver_ResolvesFromText_MOVIMM16L(t *testing.T) {
	resolver := NewInstructionResolver()
	addr := uint32(0x1000)

	pf := &ProgramFileContents{
		FileNameValue: "test.cucaracha",
		InstructionsValue: []Instruction{
			{Text: "MOVIMM16L #42, r0", Address: &addr},
		},
		MemoryLayoutValue: &MemoryLayout{
			BaseAddress: 0x1000,
			CodeStart:   0x1000,
			CodeSize:    4,
		},
	}

	result, err := resolver.ResolveInstructions(pf)
	require.NoError(t, err)

	instr := result.Instructions()[0]
	require.NotNil(t, instr.Raw)
	require.NotNil(t, instr.Instruction)

	// Verify Text
	assert.Equal(t, "MOVIMM16L #42, r0", instr.Text)

	// Verify Instruction descriptor
	assert.Equal(t, instructions.OpCode_MOV_IMM16L, instr.Instruction.Descriptor.OpCode.OpCode)
	assert.Equal(t, "MOVIMM16L", instr.Instruction.Descriptor.OpCode.Mnemonic)
	require.Len(t, instr.Instruction.OperandValues, 2, "MOVIMM16L should have 2 operands")

	// Verify operand values (#42, r0)
	assert.Equal(t, instructions.OperandKind_Immediate, instr.Instruction.OperandValues[0].Kind())
	assert.Equal(t, instructions.OperandKind_Register, instr.Instruction.OperandValues[1].Kind())
	imm := instr.Instruction.OperandValues[0].Immediate()
	assert.Equal(t, int32(42), imm.Int32())
	assert.Equal(t, "r0", instr.Instruction.OperandValues[1].Register().Name())

	// Verify Raw instruction
	assert.Equal(t, instructions.OpCode_MOV_IMM16L, instr.Raw.Descriptor.OpCode.OpCode)
	require.Len(t, instr.Raw.OperandValues, 2)
	assert.Equal(t, uint64(42), instr.Raw.OperandValues[0], "immediate should encode as 42")
	assert.Equal(t, registers.Register("r0").Encode(), instr.Raw.OperandValues[1], "raw operand 1 should match r0 encoding")
}

func TestInstructionResolver_ResolvesFromRaw(t *testing.T) {
	resolver := NewInstructionResolver()
	addr := uint32(0x1000)

	// Create a raw NOP instruction
	nopDesc, err := instructions.Instructions.Instruction(instructions.OpCode_NOP)
	require.NoError(t, err)

	rawNop := &instructions.RawInstruction{
		Descriptor:    nopDesc,
		OperandValues: []uint64{},
	}

	pf := &ProgramFileContents{
		FileNameValue: "test.cucaracha",
		InstructionsValue: []Instruction{
			{Raw: rawNop, Address: &addr},
		},
		MemoryLayoutValue: &MemoryLayout{
			BaseAddress: 0x1000,
			CodeStart:   0x1000,
			CodeSize:    4,
		},
	}

	result, err := resolver.ResolveInstructions(pf)
	require.NoError(t, err)

	instr := result.Instructions()[0]
	require.NotNil(t, instr.Instruction)

	// Verify Text was generated
	assert.NotEmpty(t, instr.Text)
	assert.Contains(t, instr.Text, "NOP")

	// Verify Instruction was decoded from Raw
	assert.Equal(t, instructions.OpCode_NOP, instr.Instruction.Descriptor.OpCode.OpCode)
	assert.Equal(t, "NOP", instr.Instruction.Descriptor.OpCode.Mnemonic)
	assert.Len(t, instr.Instruction.OperandValues, 0, "NOP should have no operands")

	// Verify Raw is preserved
	assert.Equal(t, rawNop, instr.Raw)
}

func TestInstructionResolver_ResolvesFromInstruction(t *testing.T) {
	resolver := NewInstructionResolver()
	addr := uint32(0x1000)

	// Create a decoded NOP instruction
	nopDesc, err := instructions.Instructions.Instruction(instructions.OpCode_NOP)
	require.NoError(t, err)

	decoded, err := instructions.NewInstruction(nopDesc, []instructions.OperandValue{})
	require.NoError(t, err)

	pf := &ProgramFileContents{
		FileNameValue: "test.cucaracha",
		InstructionsValue: []Instruction{
			{Instruction: decoded, Address: &addr},
		},
		MemoryLayoutValue: &MemoryLayout{
			BaseAddress: 0x1000,
			CodeStart:   0x1000,
			CodeSize:    4,
		},
	}

	result, err := resolver.ResolveInstructions(pf)
	require.NoError(t, err)

	instr := result.Instructions()[0]
	require.NotNil(t, instr.Raw)

	// Verify Text was generated
	assert.NotEmpty(t, instr.Text)
	assert.Contains(t, instr.Text, "NOP")

	// Verify Instruction is preserved
	assert.Equal(t, decoded, instr.Instruction)
	assert.Equal(t, instructions.OpCode_NOP, instr.Instruction.Descriptor.OpCode.OpCode)

	// Verify Raw was generated from Instruction
	assert.Equal(t, instructions.OpCode_NOP, instr.Raw.Descriptor.OpCode.OpCode)
	assert.Len(t, instr.Raw.OperandValues, 0, "NOP raw should have no operand values")
}

func TestInstructionResolver_ResolvesMultipleInstructions(t *testing.T) {
	resolver := NewInstructionResolver()
	addr0 := uint32(0x1000)
	addr1 := uint32(0x1004)
	addr2 := uint32(0x1008)

	pf := &ProgramFileContents{
		FileNameValue: "test.cucaracha",
		InstructionsValue: []Instruction{
			{Text: "NOP", Address: &addr0},
			{Text: "ADD r0, r1, r2", Address: &addr1},
			{Text: "MOV r2, r3", Address: &addr2},
		},
		MemoryLayoutValue: &MemoryLayout{
			BaseAddress: 0x1000,
			CodeStart:   0x1000,
			CodeSize:    12,
		},
	}

	result, err := resolver.ResolveInstructions(pf)
	require.NoError(t, err)

	instrs := result.Instructions()
	require.Len(t, instrs, 3)

	// Verify instruction 0: NOP
	require.NotNil(t, instrs[0].Raw)
	require.NotNil(t, instrs[0].Instruction)
	assert.Equal(t, "NOP", instrs[0].Text)
	assert.Equal(t, instructions.OpCode_NOP, instrs[0].Instruction.Descriptor.OpCode.OpCode)
	assert.Len(t, instrs[0].Instruction.OperandValues, 0)

	// Verify instruction 1: ADD r0, r1, r2
	require.NotNil(t, instrs[1].Raw)
	require.NotNil(t, instrs[1].Instruction)
	assert.Equal(t, "ADD r0, r1, r2", instrs[1].Text)
	assert.Equal(t, instructions.OpCode_ADD, instrs[1].Instruction.Descriptor.OpCode.OpCode)
	require.Len(t, instrs[1].Instruction.OperandValues, 3)
	assert.Equal(t, "r0", instrs[1].Instruction.OperandValues[0].Register().Name())
	assert.Equal(t, "r1", instrs[1].Instruction.OperandValues[1].Register().Name())
	assert.Equal(t, "r2", instrs[1].Instruction.OperandValues[2].Register().Name())

	// Verify instruction 2: MOV r2, r3
	require.NotNil(t, instrs[2].Raw)
	require.NotNil(t, instrs[2].Instruction)
	assert.Equal(t, "MOV r2, r3", instrs[2].Text)
	assert.Equal(t, instructions.OpCode_MOV, instrs[2].Instruction.Descriptor.OpCode.OpCode)
	require.Len(t, instrs[2].Instruction.OperandValues, 2)
	assert.Equal(t, "r2", instrs[2].Instruction.OperandValues[0].Register().Name())
	assert.Equal(t, "r3", instrs[2].Instruction.OperandValues[1].Register().Name())
}

func TestInstructionResolver_PreservesExistingFields(t *testing.T) {
	resolver := NewInstructionResolver()
	addr := uint32(0x1000)

	// Create instruction with both text and raw using actual register encodings
	addDesc, err := instructions.Instructions.Instruction(instructions.OpCode_ADD)
	require.NoError(t, err)

	// Get actual register descriptors to compute correct encodings
	r0, err := addDesc.Operands[0].ParseValue("r0")
	require.NoError(t, err)
	r1, err := addDesc.Operands[1].ParseValue("r1")
	require.NoError(t, err)
	r2, err := addDesc.Operands[2].ParseValue("r2")
	require.NoError(t, err)

	rawAdd := &instructions.RawInstruction{
		Descriptor:    addDesc,
		OperandValues: []uint64{r0.Encode(), r1.Encode(), r2.Encode()},
	}

	pf := &ProgramFileContents{
		FileNameValue: "test.cucaracha",
		InstructionsValue: []Instruction{
			{Text: "ADD r0, r1, r2", Raw: rawAdd, Address: &addr},
		},
		MemoryLayoutValue: &MemoryLayout{
			BaseAddress: 0x1000,
			CodeStart:   0x1000,
			CodeSize:    4,
		},
	}

	result, err := resolver.ResolveInstructions(pf)
	require.NoError(t, err)

	instr := result.Instructions()[0]
	require.NotNil(t, instr.Instruction)

	// Verify Text is preserved
	assert.Equal(t, "ADD r0, r1, r2", instr.Text)

	// Verify Raw is preserved
	assert.Equal(t, rawAdd, instr.Raw)
	assert.Equal(t, r0.Encode(), instr.Raw.OperandValues[0])
	assert.Equal(t, r1.Encode(), instr.Raw.OperandValues[1])
	assert.Equal(t, r2.Encode(), instr.Raw.OperandValues[2])

	// Verify Instruction was decoded correctly
	assert.Equal(t, instructions.OpCode_ADD, instr.Instruction.Descriptor.OpCode.OpCode)
	require.Len(t, instr.Instruction.OperandValues, 3)
	assert.Equal(t, "r0", instr.Instruction.OperandValues[0].Register().Name())
	assert.Equal(t, "r1", instr.Instruction.OperandValues[1].Register().Name())
	assert.Equal(t, "r2", instr.Instruction.OperandValues[2].Register().Name())
}

func TestInstructionResolver_WithContext_ResolvesLabelSymbol(t *testing.T) {
	resolver := NewInstructionResolver()
	addr0 := uint32(0x1000)
	addr1 := uint32(0x1004)
	globalAddr := uint32(0x2000)

	label := Label{Name: ".LBB0_1", InstructionIndex: 1}
	global := Global{Name: ".data", Size: 4, Address: &globalAddr}

	pf := &ProgramFileContents{
		FileNameValue: "test.cucaracha",
		InstructionsValue: []Instruction{
			{
				Text:    "MOVIMM16L .LBB0_1@lo, r0",
				Address: &addr0,
				Symbols: []SymbolReference{
					{Name: ".LBB0_1", Usage: SymbolUsageLo, Label: &label},
				},
			},
			{Text: "NOP", Address: &addr1},
		},
		GlobalsValue: []Global{global},
		LabelsValue:  []Label{label},
		MemoryLayoutValue: &MemoryLayout{
			BaseAddress: 0x1000,
			CodeStart:   0x1000,
			CodeSize:    8,
			DataStart:   0x2000,
			DataSize:    4,
		},
	}

	result, err := resolver.ResolveWithContext(pf)
	require.NoError(t, err)

	instr := result.Instructions()[0]
	require.NotNil(t, instr.Raw)
	require.NotNil(t, instr.Instruction)

	// Verify Instruction descriptor
	assert.Equal(t, instructions.OpCode_MOV_IMM16L, instr.Instruction.Descriptor.OpCode.OpCode)
	assert.Equal(t, "MOVIMM16L", instr.Instruction.Descriptor.OpCode.Mnemonic)
	require.Len(t, instr.Instruction.OperandValues, 2)

	// Verify operand values
	// The immediate should be the low 16 bits of the label's target address (0x1004)
	// 0x1004 & 0xFFFF = 0x1004 = 4100
	assert.Equal(t, instructions.OperandKind_Immediate, instr.Instruction.OperandValues[0].Kind())
	labelImm := instr.Instruction.OperandValues[0].Immediate()
	assert.Equal(t, int32(0x1004), labelImm.Int32())
	assert.Equal(t, instructions.OperandKind_Register, instr.Instruction.OperandValues[1].Kind())
	assert.Equal(t, "r0", instr.Instruction.OperandValues[1].Register().Name())

	// Verify Raw instruction
	assert.Equal(t, instructions.OpCode_MOV_IMM16L, instr.Raw.Descriptor.OpCode.OpCode)
	require.Len(t, instr.Raw.OperandValues, 2)
	assert.Equal(t, uint64(0x1004), instr.Raw.OperandValues[0], "immediate should be label address 0x1004")
	assert.Equal(t, registers.Register("r0").Encode(), instr.Raw.OperandValues[1], "raw operand 1 should match r0 encoding")
}

func TestInstructionResolver_WithContext_ResolvesGlobalSymbol(t *testing.T) {
	resolver := NewInstructionResolver()
	addr0 := uint32(0x1000)
	globalAddr := uint32(0x2000)

	global := Global{Name: ".data", Size: 4, Address: &globalAddr, Type: GlobalObject}

	pf := &ProgramFileContents{
		FileNameValue: "test.cucaracha",
		InstructionsValue: []Instruction{
			{
				Text:    "MOVIMM16L .data@lo, r0",
				Address: &addr0,
				Symbols: []SymbolReference{
					{Name: ".data", Usage: SymbolUsageLo, Global: &global},
				},
			},
		},
		GlobalsValue: []Global{global},
		MemoryLayoutValue: &MemoryLayout{
			BaseAddress: 0x1000,
			CodeStart:   0x1000,
			CodeSize:    4,
			DataStart:   0x2000,
			DataSize:    4,
		},
	}

	result, err := resolver.ResolveWithContext(pf)
	require.NoError(t, err)

	instr := result.Instructions()[0]
	require.NotNil(t, instr.Raw)
	require.NotNil(t, instr.Instruction)

	// Verify Instruction descriptor
	assert.Equal(t, instructions.OpCode_MOV_IMM16L, instr.Instruction.Descriptor.OpCode.OpCode)
	assert.Equal(t, "MOVIMM16L", instr.Instruction.Descriptor.OpCode.Mnemonic)
	require.Len(t, instr.Instruction.OperandValues, 2)

	// Verify operand values
	// The immediate should be the low 16 bits of the global's address (0x2000)
	// 0x2000 & 0xFFFF = 0x2000 = 8192
	assert.Equal(t, instructions.OperandKind_Immediate, instr.Instruction.OperandValues[0].Kind())
	globalImm := instr.Instruction.OperandValues[0].Immediate()
	assert.Equal(t, int32(0x2000), globalImm.Int32())
	assert.Equal(t, instructions.OperandKind_Register, instr.Instruction.OperandValues[1].Kind())
	assert.Equal(t, "r0", instr.Instruction.OperandValues[1].Register().Name())

	// Verify Raw instruction
	assert.Equal(t, instructions.OpCode_MOV_IMM16L, instr.Raw.Descriptor.OpCode.OpCode)
	require.Len(t, instr.Raw.OperandValues, 2)
	assert.Equal(t, uint64(0x2000), instr.Raw.OperandValues[0], "immediate should be global address 0x2000")
	assert.Equal(t, registers.Register("r0").Encode(), instr.Raw.OperandValues[1], "raw operand 1 should match r0 encoding")
}

func TestInstructionResolver_FailsOnUnknownMnemonic(t *testing.T) {
	resolver := NewInstructionResolver()
	addr := uint32(0x1000)

	pf := &ProgramFileContents{
		FileNameValue: "test.cucaracha",
		InstructionsValue: []Instruction{
			{Text: "UNKNOWN r0, r1", Address: &addr},
		},
		MemoryLayoutValue: &MemoryLayout{
			BaseAddress: 0x1000,
			CodeStart:   0x1000,
			CodeSize:    4,
		},
	}

	_, err := resolver.ResolveInstructions(pf)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInstructionParsing)
	assert.Contains(t, err.Error(), "UNKNOWN")
}

func TestInstructionResolver_FailsOnWrongOperandCount(t *testing.T) {
	resolver := NewInstructionResolver()
	addr := uint32(0x1000)

	pf := &ProgramFileContents{
		FileNameValue: "test.cucaracha",
		InstructionsValue: []Instruction{
			{Text: "ADD r0, r1", Address: &addr}, // Missing third operand
		},
		MemoryLayoutValue: &MemoryLayout{
			BaseAddress: 0x1000,
			CodeStart:   0x1000,
			CodeSize:    4,
		},
	}

	_, err := resolver.ResolveInstructions(pf)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInstructionParsing)
}

func TestInstructionResolver_GenerateTextFromInstruction_SymbolOnlyForImmediate(t *testing.T) {
	// This test ensures that symbol references are only applied to immediate operands,
	// not to register operands. This was a bug where MOVIMM16H label@hi, r0 would
	// be generated as MOVIMM16H label@hi label@hi (symbol applied to both operands).

	resolver := NewInstructionResolver()

	// Create a MOVIMM16H instruction with immediate value and register
	desc, err := instructions.Instructions.Instruction(instructions.OpCode_MOV_IMM16H)
	require.NoError(t, err)

	// Create operand values: immediate (0x1234), dst register (r0), and src register (r0, tied to dst)
	// MOVIMM16H has 3 operands in LLVM: imm, dst, src (where src is tied to dst and hidden from asm)
	r0 := registers.Register("r0")

	instr := &instructions.Instruction{
		Descriptor: desc,
		OperandValues: []instructions.OperandValue{
			Imm16(0x1234),
			instructions.RegisterOperandValue(r0), // dst
			instructions.RegisterOperandValue(r0), // src (tied to dst, LLVM_HideFromAsm=true)
		},
	}

	// Symbol reference that should only be applied to the immediate operand
	symbols := []SymbolReference{
		{Name: "myLabel", Usage: SymbolUsageHi},
	}

	// Generate text
	text := resolver.generateTextFromInstruction(instr, symbols)

	// Should be "MOVIMM16H myLabel@hi, r0" - the third operand (src) should be hidden
	assert.Equal(t, "MOVIMM16H myLabel@hi, r0", text)
	assert.NotContains(t, text, "myLabel@hi, myLabel@hi", "symbol should not be applied to register operand")
	assert.NotContains(t, text, "r0, r0", "hidden src operand should not appear in text")
}

func TestInstructionResolver_GenerateTextFromInstruction_HidesLLVMInternalOperands(t *testing.T) {
	// This test specifically verifies that operands marked with LLVM_HideFromAsm=true
	// are not included in the generated text. This is important for MOVIMM16H which
	// has a third operand (src tied to dst) that is only used by LLVM for register allocation.

	resolver := NewInstructionResolver()

	desc, err := instructions.Instructions.Instruction(instructions.OpCode_MOV_IMM16H)
	require.NoError(t, err)

	r5 := registers.Register("r5")

	// All 3 operands including the hidden src
	instr := &instructions.Instruction{
		Descriptor: desc,
		OperandValues: []instructions.OperandValue{
			Imm16(0xABCD),
			instructions.RegisterOperandValue(r5), // dst
			instructions.RegisterOperandValue(r5), // src (hidden)
		},
	}

	text := resolver.generateTextFromInstruction(instr, nil)

	// Should only show 2 operands: immediate and dst register
	assert.Equal(t, "MOVIMM16H #43981, r5", text) // 0xABCD = 43981
	// Verify only one r5 appears (not r5, r5)
	assert.Equal(t, 1, strings.Count(text, "r5"), "should only have one r5 in output, hidden src operand should not appear")
}

func TestInstructionResolver_GenerateTextFromInstruction_MultipleSymbols(t *testing.T) {
	// Test that multiple symbols are applied to different immediate operands correctly
	resolver := NewInstructionResolver()

	// Create a MOVIMM16L instruction
	desc, err := instructions.Instructions.Instruction(instructions.OpCode_MOV_IMM16L)
	require.NoError(t, err)

	r1 := registers.Register("r1")

	instr := &instructions.Instruction{
		Descriptor: desc,
		OperandValues: []instructions.OperandValue{
			Imm16(0x5678),
			instructions.RegisterOperandValue(r1),
		},
	}

	symbols := []SymbolReference{
		{Name: "funcAddr", Usage: SymbolUsageLo},
	}

	text := resolver.generateTextFromInstruction(instr, symbols)

	assert.Equal(t, "MOVIMM16L funcAddr@lo, r1", text)
}

func TestInstructionResolver_GenerateTextFromInstruction_NoSymbols(t *testing.T) {
	// Test that instructions without symbols generate correct text
	resolver := NewInstructionResolver()

	desc, err := instructions.Instructions.Instruction(instructions.OpCode_ADD)
	require.NoError(t, err)

	r0 := registers.Register("r0")
	r1 := registers.Register("r1")
	r2 := registers.Register("r2")

	// ADD r0, r1, r2
	instr := &instructions.Instruction{
		Descriptor: desc,
		OperandValues: []instructions.OperandValue{
			instructions.RegisterOperandValue(r0),
			instructions.RegisterOperandValue(r1),
			instructions.RegisterOperandValue(r2),
		},
	}

	text := resolver.generateTextFromInstruction(instr, nil)

	assert.Equal(t, "ADD r0, r1, r2", text)
}
