package program

import (
	"testing"

	"github.com/Manu343726/cucaracha/pkg/hw/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlignAddress(t *testing.T) {
	tests := []struct {
		name      string
		addr      uint32
		alignment uint32
		expected  uint32
	}{
		{"already aligned", 0x100, 4, 0x100},
		{"needs alignment by 1", 0x101, 4, 0x104},
		{"needs alignment by 2", 0x102, 4, 0x104},
		{"needs alignment by 3", 0x103, 4, 0x104},
		{"zero address", 0, 4, 0},
		{"alignment 8", 0x105, 8, 0x108},
		{"alignment 16", 0x110, 16, 0x110},
		{"alignment 16 needs align", 0x111, 16, 0x120},
		{"zero alignment", 0x105, 0, 0x105},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := alignAddress(tt.addr, tt.alignment)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResolveMemory_EmptyProgram(t *testing.T) {
	pf := &ProgramFileContents{
		FileNameValue:     "empty.cucaracha",
		SourceFileValue:   "empty.c",
		FunctionsValue:    map[string]Function{},
		InstructionsValue: []Instruction{},
		GlobalsValue:      []Global{},
		LabelsValue:       []Label{},
	}

	layout := &memory.MemoryLayout{
		CodeBase: 0x1000,
		CodeSize: 512,
		DataBase: 0x1200,
		DataSize: 256,
	}

	resolved, err := ResolveMemory(pf, layout)
	require.NoError(t, err)
	require.NotNil(t, resolved)

	resultLayout := resolved.MemoryLayout()
	require.NotNil(t, resultLayout)
	assert.Equal(t, uint32(0x1000), resultLayout.CodeBase)
	assert.Equal(t, uint32(0), resultLayout.CodeSize)
}

func TestResolveMemory_InstructionsOnly(t *testing.T) {
	pf := &ProgramFileContents{
		FileNameValue:   "code.cucaracha",
		SourceFileValue: "code.c",
		FunctionsValue:  map[string]Function{},
		InstructionsValue: []Instruction{
			{LineNumber: 1, Text: "nop"},
			{LineNumber: 2, Text: "nop"},
			{LineNumber: 3, Text: "nop"},
		},
		GlobalsValue: []Global{},
		LabelsValue:  []Label{},
	}

	layout := &memory.MemoryLayout{
		CodeBase: 0x2000,
		CodeSize: 1024,
		DataBase: 0x2400,
		DataSize: 256,
	}

	resolved, err := ResolveMemory(pf, layout)
	require.NoError(t, err)
	require.NotNil(t, resolved)

	resultLayout := resolved.MemoryLayout()
	require.NotNil(t, resultLayout)
	assert.Equal(t, uint32(0x2000), resultLayout.CodeBase)
	assert.Equal(t, uint32(12), resultLayout.CodeSize) // 3 instructions * 4 bytes

	// Check instruction addresses
	instructions := resolved.Instructions()
	require.Len(t, instructions, 3)
	require.NotNil(t, instructions[0].Address)
	require.NotNil(t, instructions[1].Address)
	require.NotNil(t, instructions[2].Address)
	assert.Equal(t, uint32(0x2000), *instructions[0].Address)
	assert.Equal(t, uint32(0x2004), *instructions[1].Address)
	assert.Equal(t, uint32(0x2008), *instructions[2].Address)
}

func TestResolveMemory_GlobalsOnly(t *testing.T) {
	pf := &ProgramFileContents{
		FileNameValue:     "data.cucaracha",
		SourceFileValue:   "data.c",
		FunctionsValue:    map[string]Function{},
		InstructionsValue: []Instruction{},
		GlobalsValue: []Global{
			{Name: "var1", Size: 4, Type: GlobalObject},
			{Name: "var2", Size: 8, Type: GlobalObject},
		},
		LabelsValue: []Label{},
	}

	layout := &memory.MemoryLayout{
		CodeBase: 0x1000,
		CodeSize: 512,
		DataBase: 0x1200,
		DataSize: 256,
	}

	resolved, err := ResolveMemory(pf, layout)
	require.NoError(t, err)
	require.NotNil(t, resolved)

	resultLayout := resolved.MemoryLayout()
	require.NotNil(t, resultLayout)
	assert.Equal(t, uint32(0x1200), resultLayout.DataBase)
	assert.Equal(t, uint32(12), resultLayout.DataSize) // 4 + 8 bytes

	// Check global addresses
	globals := resolved.Globals()
	require.Len(t, globals, 2)
	require.NotNil(t, globals[0].Address)
	require.NotNil(t, globals[1].Address)
	assert.Equal(t, uint32(0x1200), *globals[0].Address)
	assert.Equal(t, uint32(0x1204), *globals[1].Address)
}

func TestResolveMemory_CodeAndData(t *testing.T) {
	pf := &ProgramFileContents{
		FileNameValue:   "mixed.cucaracha",
		SourceFileValue: "mixed.c",
		FunctionsValue:  map[string]Function{},
		InstructionsValue: []Instruction{
			{LineNumber: 1, Text: "movimm16l 0, r0"},
			{LineNumber: 2, Text: "ld r0, r1"},
		},
		GlobalsValue: []Global{
			{Name: "data", Size: 16, Type: GlobalObject},
		},
		LabelsValue: []Label{},
	}

	layout := &memory.MemoryLayout{
		CodeBase: 0x0,
		CodeSize: 256,
		DataBase: 0x100,
		DataSize: 256,
	}

	resolved, err := ResolveMemory(pf, layout)
	require.NoError(t, err)
	require.NotNil(t, resolved)

	resultLayout := resolved.MemoryLayout()
	require.NotNil(t, resultLayout)
	assert.Equal(t, uint32(0), resultLayout.CodeBase)
	assert.Equal(t, uint32(8), resultLayout.CodeSize)  // 2 instructions * 4 bytes
	assert.Equal(t, uint32(16), resultLayout.DataSize) // 16 bytes
	assert.Equal(t, uint32(0x100), resultLayout.DataBase)

	// Check addresses
	instructions := resolved.Instructions()
	assert.Equal(t, uint32(0x0), *instructions[0].Address)
	assert.Equal(t, uint32(0x4), *instructions[1].Address)

	globals := resolved.Globals()
	assert.Equal(t, uint32(0x100), *globals[0].Address)
}

func TestResolveMemory_DataAlignment(t *testing.T) {
	pf := &ProgramFileContents{
		FileNameValue:   "align.cucaracha",
		SourceFileValue: "align.c",
		FunctionsValue:  map[string]Function{},
		InstructionsValue: []Instruction{
			{LineNumber: 1, Text: "nop"},
			{LineNumber: 2, Text: "nop"},
			{LineNumber: 3, Text: "nop"}, // 12 bytes of code
		},
		GlobalsValue: []Global{
			{Name: "data", Size: 4, Type: GlobalObject},
		},
		LabelsValue: []Label{},
	}

	layout := &memory.MemoryLayout{
		CodeBase: 0x0,
		CodeSize: 512,
		DataBase: 0x200,
		DataSize: 256,
	}

	resolved, err := ResolveMemory(pf, layout)
	require.NoError(t, err)
	require.NotNil(t, resolved)

	resultLayout := resolved.MemoryLayout()
	require.NotNil(t, resultLayout)
	assert.Equal(t, uint32(12), resultLayout.CodeSize)
	assert.Equal(t, uint32(4), resultLayout.DataSize)
	assert.Equal(t, uint32(0x200), resultLayout.DataBase)
}

func TestResolveMemory_MaxSizeExceeded(t *testing.T) {
	pf := &ProgramFileContents{
		FileNameValue:   "large.cucaracha",
		SourceFileValue: "large.c",
		FunctionsValue:  map[string]Function{},
		InstructionsValue: []Instruction{
			{LineNumber: 1, Text: "nop"},
			{LineNumber: 2, Text: "nop"},
			{LineNumber: 3, Text: "nop"},
		},
		GlobalsValue: []Global{
			{Name: "data", Size: 100, Type: GlobalObject},
		},
		LabelsValue: []Label{},
	}

	// Program needs 12 bytes code + 100 bytes data = 112 bytes total
	// But layout only provides 11 bytes in code section to trigger overflow
	layout := &memory.MemoryLayout{
		CodeBase: 0x0,
		CodeSize: 11,
		DataBase: 0x100,
		DataSize: 256,
	}

	resolved, err := ResolveMemory(pf, layout)
	assert.Error(t, err)
	assert.Nil(t, resolved)
	assert.ErrorIs(t, err, ErrProgramTooLarge)
}

func TestResolveMemory_MaxSizeJustRight(t *testing.T) {
	pf := &ProgramFileContents{
		FileNameValue:   "exact.cucaracha",
		SourceFileValue: "exact.c",
		FunctionsValue:  map[string]Function{},
		InstructionsValue: []Instruction{
			{LineNumber: 1, Text: "nop"},
		},
		GlobalsValue: []Global{
			{Name: "data", Size: 4, Type: GlobalObject},
		},
		LabelsValue: []Label{},
	}

	// Program needs 4 bytes code + 4 bytes data = 8 bytes total
	layout := &memory.MemoryLayout{
		CodeBase: 0x0,
		CodeSize: 256,
		DataBase: 0x100,
		DataSize: 256,
	}

	resolved, err := ResolveMemory(pf, layout)
	require.NoError(t, err)
	require.NotNil(t, resolved)
	resultLayout := resolved.MemoryLayout()
	assert.Equal(t, uint32(4), resultLayout.CodeSize)
	assert.Equal(t, uint32(4), resultLayout.DataSize)
}

func TestResolveMemory_WithLabels(t *testing.T) {
	pf := &ProgramFileContents{
		FileNameValue:   "labels.cucaracha",
		SourceFileValue: "labels.c",
		FunctionsValue:  map[string]Function{},
		InstructionsValue: []Instruction{
			{LineNumber: 1, Text: "nop"},        // 0x1000
			{LineNumber: 2, Text: "nop"},        // 0x1004 - loop_start
			{LineNumber: 3, Text: "jmp r0, r1"}, // 0x1008
		},
		GlobalsValue: []Global{},
		LabelsValue: []Label{
			{Name: "loop_start", InstructionIndex: 1},
		},
	}

	layout := &memory.MemoryLayout{
		CodeBase: 0x1000,
		CodeSize: 512,
		DataBase: 0x1200,
		DataSize: 256,
	}

	resolved, err := ResolveMemory(pf, layout)
	require.NoError(t, err)
	require.NotNil(t, resolved)

	// Check that label address can be looked up
	labels := resolved.Labels()
	require.Len(t, labels, 1)
	assert.Equal(t, "loop_start", labels[0].Name)
	assert.Equal(t, 1, labels[0].InstructionIndex)

	// The instruction at index 1 should have address 0x1004
	instructions := resolved.Instructions()
	assert.Equal(t, uint32(0x1004), *instructions[1].Address)
}

func TestResolveMemory_WithFunctions(t *testing.T) {
	pf := &ProgramFileContents{
		FileNameValue:   "funcs.cucaracha",
		SourceFileValue: "funcs.c",
		FunctionsValue: map[string]Function{
			"main": {
				Name:              "main",
				InstructionRanges: []InstructionRange{{Start: 0, Count: 2}},
			},
			"helper": {
				Name:              "helper",
				InstructionRanges: []InstructionRange{{Start: 2, Count: 1}},
			},
		},
		InstructionsValue: []Instruction{
			{LineNumber: 1, Text: "movimm16l 0, r0"}, // main:     0x0
			{LineNumber: 2, Text: "jmp r0, r1"},      //           0x4
			{LineNumber: 3, Text: "nop"},             // helper:   0x8
		},
		GlobalsValue: []Global{},
		LabelsValue:  []Label{},
	}

	layout := &memory.MemoryLayout{
		CodeBase: 0x0,
		CodeSize: 512,
		DataBase: 0x200,
		DataSize: 256,
	}

	resolved, err := ResolveMemory(pf, layout)
	require.NoError(t, err)
	require.NotNil(t, resolved)

	functions := resolved.Functions()
	require.Contains(t, functions, "main")
	require.Contains(t, functions, "helper")

	// Function addresses are derived from their first instruction
	instructions := resolved.Instructions()
	assert.Equal(t, uint32(0x0), *instructions[0].Address) // main starts here
	assert.Equal(t, uint32(0x8), *instructions[2].Address) // helper starts here
}

func TestResolveMemory_WithSymbolReferences(t *testing.T) {
	global := Global{Name: "counter", Size: 4, Type: GlobalObject}

	pf := &ProgramFileContents{
		FileNameValue:   "symbols.cucaracha",
		SourceFileValue: "symbols.c",
		FunctionsValue:  map[string]Function{},
		InstructionsValue: []Instruction{
			{
				LineNumber: 1,
				Text:       "movimm16l counter@lo, r0",
				Symbols: []SymbolReference{
					{Name: "counter", Usage: SymbolUsageLo},
				},
			},
			{
				LineNumber: 2,
				Text:       "movimm16h counter@hi, r0",
				Symbols: []SymbolReference{
					{Name: "counter", Usage: SymbolUsageHi},
				},
			},
		},
		GlobalsValue: []Global{global},
		LabelsValue:  []Label{},
	}

	layout := &memory.MemoryLayout{
		CodeBase: 0x0,
		CodeSize: 512,
		DataBase: 0x200,
		DataSize: 256,
	}

	resolved, err := ResolveMemory(pf, layout)
	require.NoError(t, err)
	require.NotNil(t, resolved)

	// Check that symbol references point to the resolved global
	instructions := resolved.Instructions()
	require.Len(t, instructions[0].Symbols, 1)
	require.NotNil(t, instructions[0].Symbols[0].Global)
	require.NotNil(t, instructions[0].Symbols[0].Global.Address)

	// Global should be at address 0x200 (data section)
	assert.Equal(t, uint32(0x200), *instructions[0].Symbols[0].Global.Address)
}

func TestGetSymbolAddressFromProgram(t *testing.T) {
	pf := &ProgramFileContents{
		FileNameValue:   "addrlookup.cucaracha",
		SourceFileValue: "addrlookup.c",
		FunctionsValue: map[string]Function{
			"func1": {
				Name:              "func1",
				InstructionRanges: []InstructionRange{{Start: 0, Count: 1}},
			},
		},
		InstructionsValue: []Instruction{
			{LineNumber: 1, Text: "nop"},
			{LineNumber: 2, Text: "nop"},
		},
		GlobalsValue: []Global{
			{Name: "myvar", Size: 4, Type: GlobalObject},
		},
		LabelsValue: []Label{
			{Name: "mylabel", InstructionIndex: 1},
		},
	}

	layout := &memory.MemoryLayout{
		CodeBase: 0x1000,
		CodeSize: 512,
		DataBase: 0x1200,
		DataSize: 256,
	}

	resolved, err := ResolveMemory(pf, layout)
	require.NoError(t, err)

	t.Run("global address", func(t *testing.T) {
		globals := resolved.Globals()
		sym := &SymbolReference{Name: "myvar", Global: &globals[0]}
		addr, ok := GetSymbolAddressFromProgram(sym, resolved)
		assert.True(t, ok)
		assert.Equal(t, uint32(0x1200), addr) // Data section base
	})

	t.Run("label address", func(t *testing.T) {
		labels := resolved.Labels()
		sym := &SymbolReference{Name: "mylabel", Label: &labels[0]}
		addr, ok := GetSymbolAddressFromProgram(sym, resolved)
		assert.True(t, ok)
		assert.Equal(t, uint32(0x1004), addr) // Second instruction
	})

	t.Run("function address", func(t *testing.T) {
		functions := resolved.Functions()
		fn := functions["func1"]
		sym := &SymbolReference{Name: "func1", Function: &fn}
		addr, ok := GetSymbolAddressFromProgram(sym, resolved)
		assert.True(t, ok)
		assert.Equal(t, uint32(0x1000), addr) // First instruction
	})

	t.Run("nil symbol", func(t *testing.T) {
		addr, ok := GetSymbolAddressFromProgram(nil, resolved)
		assert.False(t, ok)
		assert.Equal(t, uint32(0), addr)
	})
}

func TestResolveMemory_PreservesFileInfo(t *testing.T) {
	pf := &ProgramFileContents{
		FileNameValue:     "test.cucaracha",
		SourceFileValue:   "test.c",
		FunctionsValue:    map[string]Function{},
		InstructionsValue: []Instruction{{LineNumber: 1, Text: "nop"}},
		GlobalsValue:      []Global{},
		LabelsValue:       []Label{},
	}

	layout := &memory.MemoryLayout{
		CodeBase: 0x0,
		CodeSize: 512,
		DataBase: 0x200,
		DataSize: 256,
	}

	resolved, err := ResolveMemory(pf, layout)
	require.NoError(t, err)

	assert.Equal(t, "test.cucaracha", resolved.FileName())
	assert.Equal(t, "test.c", resolved.SourceFile())
}
