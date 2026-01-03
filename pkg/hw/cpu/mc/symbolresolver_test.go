package mc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProgramFile is a test implementation of ProgramFile
type mockProgramFile struct {
	fileName     string
	sourceFile   string
	functions    map[string]Function
	instructions []Instruction
	globals      []Global
	labels       []Label
}

func (m *mockProgramFile) FileName() string               { return m.fileName }
func (m *mockProgramFile) SourceFile() string             { return m.sourceFile }
func (m *mockProgramFile) Functions() map[string]Function { return m.functions }
func (m *mockProgramFile) Instructions() []Instruction    { return m.instructions }
func (m *mockProgramFile) Globals() []Global              { return m.globals }
func (m *mockProgramFile) Labels() []Label                { return m.labels }
func (m *mockProgramFile) MemoryLayout() *MemoryLayout    { return nil }
func (m *mockProgramFile) DebugInfo() *DebugInfo          { return nil }

func TestResolveSymbols_ResolvesAllSymbolTypes(t *testing.T) {
	// Create a program with functions, globals, and labels
	pf := &mockProgramFile{
		fileName:   "test.cucaracha",
		sourceFile: "test.c",
		functions: map[string]Function{
			"main": {
				Name:       "main",
				SourceFile: "test.c",
				StartLine:  1,
				EndLine:    10,
				InstructionRanges: []InstructionRange{
					{Start: 0, Count: 3},
				},
			},
			"helper": {
				Name:       "helper",
				SourceFile: "test.c",
				StartLine:  12,
				EndLine:    20,
				InstructionRanges: []InstructionRange{
					{Start: 3, Count: 2},
				},
			},
		},
		globals: []Global{
			{Name: ".L__const.data", Size: 16, Type: GlobalObject},
		},
		labels: []Label{
			{Name: ".LBB0_1", InstructionIndex: 1},
			{Name: ".LBB0_2", InstructionIndex: 2},
		},
		instructions: []Instruction{
			{LineNumber: 1, Text: "MOVIMM16L helper@lo, r0", Symbols: []SymbolReference{{Name: "helper", Usage: SymbolUsageLo}}},
			{LineNumber: 2, Text: "MOVIMM16L .L__const.data@hi, r1", Symbols: []SymbolReference{{Name: ".L__const.data", Usage: SymbolUsageHi}}},
			{LineNumber: 3, Text: "JMP .LBB0_1", Symbols: []SymbolReference{{Name: ".LBB0_1"}}},
			{LineNumber: 4, Text: "ADD r0, r1, r2", Symbols: []SymbolReference{}},
			{LineNumber: 5, Text: "RET", Symbols: []SymbolReference{}},
		},
	}

	resolved, err := ResolveSymbols(pf)
	require.NoError(t, err)
	require.NotNil(t, resolved)

	// Verify metadata is preserved
	assert.Equal(t, "test.cucaracha", resolved.FileName())
	assert.Equal(t, "test.c", resolved.SourceFile())
	assert.Len(t, resolved.Functions(), 2)
	assert.Len(t, resolved.Globals(), 1)
	assert.Len(t, resolved.Labels(), 2)

	instructions := resolved.Instructions()
	require.Len(t, instructions, 5)

	// Instruction 0: references function "helper"
	require.Len(t, instructions[0].Symbols, 1)
	sym0 := instructions[0].Symbols[0]
	assert.Equal(t, "helper", sym0.Name)
	assert.Equal(t, SymbolUsageLo, sym0.Usage)
	assert.False(t, sym0.Unresolved(), "symbol should be resolved")
	assert.Equal(t, SymbolKindFunction, sym0.Kind())
	assert.NotNil(t, sym0.Function)
	assert.Equal(t, "helper", sym0.Function.Name)

	// Instruction 1: references global ".L__const.data"
	require.Len(t, instructions[1].Symbols, 1)
	sym1 := instructions[1].Symbols[0]
	assert.Equal(t, ".L__const.data", sym1.Name)
	assert.Equal(t, SymbolUsageHi, sym1.Usage)
	assert.False(t, sym1.Unresolved(), "symbol should be resolved")
	assert.Equal(t, SymbolKindGlobal, sym1.Kind())
	assert.NotNil(t, sym1.Global)
	assert.Equal(t, ".L__const.data", sym1.Global.Name)

	// Instruction 2: references label ".LBB0_1"
	require.Len(t, instructions[2].Symbols, 1)
	sym2 := instructions[2].Symbols[0]
	assert.Equal(t, ".LBB0_1", sym2.Name)
	assert.False(t, sym2.Unresolved(), "symbol should be resolved")
	assert.Equal(t, SymbolKindLabel, sym2.Kind())
	assert.NotNil(t, sym2.Label)
	assert.Equal(t, ".LBB0_1", sym2.Label.Name)

	// Instructions 3 and 4: no symbols
	assert.Len(t, instructions[3].Symbols, 0)
	assert.Len(t, instructions[4].Symbols, 0)
}

func TestResolveSymbols_ErrorOnUnresolvedSymbol(t *testing.T) {
	pf := &mockProgramFile{
		fileName:   "test.cucaracha",
		sourceFile: "test.c",
		functions:  map[string]Function{},
		globals:    []Global{},
		labels:     []Label{},
		instructions: []Instruction{
			{LineNumber: 1, Text: "MOVIMM16L unknown_symbol, r0", Symbols: []SymbolReference{{Name: "unknown_symbol"}}},
		},
	}

	resolved, err := ResolveSymbols(pf)
	assert.Nil(t, resolved)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unresolved symbols")
	assert.Contains(t, err.Error(), "unknown_symbol")
}

func TestResolveSymbols_MultipleUnresolvedSymbols(t *testing.T) {
	pf := &mockProgramFile{
		fileName:   "test.cucaracha",
		sourceFile: "test.c",
		functions:  map[string]Function{},
		globals:    []Global{},
		labels:     []Label{},
		instructions: []Instruction{
			{LineNumber: 1, Text: "MOVIMM16L sym1, r0", Symbols: []SymbolReference{{Name: "sym1"}}},
			{LineNumber: 2, Text: "MOVIMM16L sym2, r1", Symbols: []SymbolReference{{Name: "sym2"}}},
		},
	}

	resolved, err := ResolveSymbols(pf)
	assert.Nil(t, resolved)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sym1")
	assert.Contains(t, err.Error(), "sym2")
}

func TestResolveSymbols_NoSymbols(t *testing.T) {
	pf := &mockProgramFile{
		fileName:   "test.cucaracha",
		sourceFile: "test.c",
		functions: map[string]Function{
			"main": {Name: "main"},
		},
		globals: []Global{},
		labels:  []Label{},
		instructions: []Instruction{
			{LineNumber: 1, Text: "ADD r0, r1, r2", Symbols: []SymbolReference{}},
			{LineNumber: 2, Text: "RET", Symbols: []SymbolReference{}},
		},
	}

	resolved, err := ResolveSymbols(pf)
	require.NoError(t, err)
	require.NotNil(t, resolved)

	instructions := resolved.Instructions()
	assert.Len(t, instructions, 2)
	assert.Len(t, instructions[0].Symbols, 0)
	assert.Len(t, instructions[1].Symbols, 0)
}

func TestResolveSymbols_PreservesInstructionData(t *testing.T) {
	addr := uint32(0x1000)
	pf := &mockProgramFile{
		fileName:   "test.cucaracha",
		sourceFile: "test.c",
		functions:  map[string]Function{},
		globals:    []Global{},
		labels:     []Label{},
		instructions: []Instruction{
			{
				LineNumber: 42,
				Address:    &addr,
				Text:       "NOP",
				Symbols:    []SymbolReference{},
			},
		},
	}

	resolved, err := ResolveSymbols(pf)
	require.NoError(t, err)

	instructions := resolved.Instructions()
	require.Len(t, instructions, 1)
	assert.Equal(t, 42, instructions[0].LineNumber)
	assert.Equal(t, "NOP", instructions[0].Text)
	require.NotNil(t, instructions[0].Address)
	assert.Equal(t, uint32(0x1000), *instructions[0].Address)
}

func TestResolveSymbols_UsageField(t *testing.T) {
	// Test that Usage field is properly preserved through resolution
	pf := &mockProgramFile{
		fileName:   "test.cucaracha",
		sourceFile: "test.c",
		functions: map[string]Function{
			"myFunc": {Name: "myFunc"},
		},
		globals: []Global{
			{Name: "myGlobal", Type: GlobalObject},
		},
		labels: []Label{
			{Name: ".LBB0_1", InstructionIndex: 0},
		},
		instructions: []Instruction{
			{LineNumber: 1, Text: "MOVIMM16L myFunc@lo, r0", Symbols: []SymbolReference{{Name: "myFunc", Usage: SymbolUsageLo}}},
			{LineNumber: 2, Text: "MOVIMM16H myFunc@hi, r0", Symbols: []SymbolReference{{Name: "myFunc", Usage: SymbolUsageHi}}},
			{LineNumber: 3, Text: "MOVIMM16L myGlobal@lo, r1", Symbols: []SymbolReference{{Name: "myGlobal", Usage: SymbolUsageLo}}},
			{LineNumber: 4, Text: "MOVIMM16L .LBB0_1@lo, r2", Symbols: []SymbolReference{{Name: ".LBB0_1", Usage: SymbolUsageLo}}},
			{LineNumber: 5, Text: "JMP r0, lr", Symbols: []SymbolReference{}}, // JMP uses registers only, no symbols
		},
	}

	resolved, err := ResolveSymbols(pf)
	require.NoError(t, err)

	instructions := resolved.Instructions()

	// All symbols should be resolved and Usage preserved
	assert.Equal(t, SymbolKindFunction, instructions[0].Symbols[0].Kind())
	assert.Equal(t, "myFunc", instructions[0].Symbols[0].Name)
	assert.Equal(t, "myFunc", instructions[0].Symbols[0].Function.Name)
	assert.Equal(t, SymbolUsageLo, instructions[0].Symbols[0].Usage)

	assert.Equal(t, SymbolKindFunction, instructions[1].Symbols[0].Kind())
	assert.Equal(t, "myFunc", instructions[1].Symbols[0].Name)
	assert.Equal(t, SymbolUsageHi, instructions[1].Symbols[0].Usage)

	assert.Equal(t, SymbolKindGlobal, instructions[2].Symbols[0].Kind())
	assert.Equal(t, "myGlobal", instructions[2].Symbols[0].Name)
	assert.Equal(t, "myGlobal", instructions[2].Symbols[0].Global.Name)
	assert.Equal(t, SymbolUsageLo, instructions[2].Symbols[0].Usage)

	assert.Equal(t, SymbolKindLabel, instructions[3].Symbols[0].Kind())
	assert.Equal(t, ".LBB0_1", instructions[3].Symbols[0].Name)
	assert.Equal(t, ".LBB0_1", instructions[3].Symbols[0].Label.Name)
	assert.Equal(t, SymbolUsageLo, instructions[3].Symbols[0].Usage)

	// JMP instruction has no symbols (register-indirect only)
	assert.Equal(t, 0, len(instructions[4].Symbols))
}

func TestProgramFileContents_ImplementsInterface(t *testing.T) {
	// Compile-time check that ProgramFileContents implements ProgramFile
	var _ ProgramFile = (*ProgramFileContents)(nil)
}
