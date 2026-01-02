package mc

import (
	"bytes"
	"testing"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDumpProgramFile_EmptyProgram(t *testing.T) {
	pf := &ProgramFileContents{
		FileNameValue:     "empty.cucaracha",
		SourceFileValue:   "empty.c",
		FunctionsValue:    map[string]Function{},
		InstructionsValue: []Instruction{},
		GlobalsValue:      []Global{},
		LabelsValue:       []Label{},
	}

	var buf bytes.Buffer
	err := DumpProgramFile(&buf, pf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "=== Program File ===")
	assert.Contains(t, output, "File: empty.cucaracha")
	assert.Contains(t, output, "Source: empty.c")
	assert.Contains(t, output, "=== Memory Layout ===")
	assert.Contains(t, output, "(not resolved)")
	assert.Contains(t, output, "=== Functions (0) ===")
	assert.Contains(t, output, "=== Labels (0) ===")
	assert.Contains(t, output, "=== Globals (0) ===")
	assert.Contains(t, output, "=== Instructions (0) ===")
}

func TestDumpProgramFile_TextOnlyProgram(t *testing.T) {
	// A program with only text (assembly) representation, nothing resolved
	pf := &ProgramFileContents{
		FileNameValue:   "text_only.cucaracha",
		SourceFileValue: "main.c",
		FunctionsValue: map[string]Function{
			"main": {
				Name:       "main",
				SourceFile: "main.c",
				StartLine:  5,
				EndLine:    15,
				InstructionRanges: []InstructionRange{
					{Start: 0, Count: 4},
				},
			},
		},
		InstructionsValue: []Instruction{
			{LineNumber: 6, Text: "MOVIMM16L #10, r0"},
			{LineNumber: 7, Text: "MOVIMM16L #20, r1"},
			{LineNumber: 8, Text: "ADD r0, r1, r2"},
			{LineNumber: 9, Text: "JMP r4, r5"},
		},
		GlobalsValue: []Global{},
		LabelsValue: []Label{
			{Name: ".LBB0_1", InstructionIndex: 2},
		},
	}

	var buf bytes.Buffer
	err := DumpProgramFile(&buf, pf)
	require.NoError(t, err)

	output := buf.String()

	// Header
	assert.Contains(t, output, "File: text_only.cucaracha")
	assert.Contains(t, output, "Source: main.c")

	// Memory layout not resolved
	assert.Contains(t, output, "=== Memory Layout ===")
	assert.Contains(t, output, "(not resolved)")

	// Function info
	assert.Contains(t, output, "=== Functions (1) ===")
	assert.Contains(t, output, "main:")
	assert.Contains(t, output, "Source: main.c:5-15")
	assert.Contains(t, output, "Instruction Ranges: [0..3]")

	// Labels
	assert.Contains(t, output, "=== Labels (1) ===")
	assert.Contains(t, output, ".LBB0_1 -> instruction #2")

	// Instructions with text only (no raw, no address, not decoded)
	assert.Contains(t, output, "=== Instructions (4) ===")
	assert.Contains(t, output, "; function: main")
	assert.Contains(t, output, "[   0] 0x--------  (no raw)  (not decoded)  MOVIMM16L #10, r0")
	assert.Contains(t, output, "; line 6")
	assert.Contains(t, output, ".LBB0_1:")
	assert.Contains(t, output, "[   2] 0x--------  (no raw)  (not decoded)  ADD r0, r1, r2")
}

func TestDumpProgramFile_WithResolvedSymbols(t *testing.T) {
	// Create function, global, and label that will be referenced
	mainFunc := Function{
		Name:       "main",
		SourceFile: "symbols.c",
		StartLine:  1,
		EndLine:    10,
		InstructionRanges: []InstructionRange{
			{Start: 0, Count: 3},
		},
	}

	globalVar := Global{
		Name:        ".L__const.data",
		Size:        8,
		InitialData: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		Type:        GlobalObject,
	}

	label := Label{
		Name:             ".LBB0_1",
		InstructionIndex: 1,
	}

	pf := &ProgramFileContents{
		FileNameValue:   "symbols.cucaracha",
		SourceFileValue: "symbols.c",
		FunctionsValue: map[string]Function{
			"main": mainFunc,
		},
		InstructionsValue: []Instruction{
			{
				LineNumber: 2,
				Text:       "MOVIMM16L .L__const.data@lo, r0",
				Symbols: []SymbolReference{
					{Name: ".L__const.data", Usage: SymbolUsageLo, Global: &globalVar},
				},
			},
			{
				LineNumber: 3,
				Text:       "MOVIMM16H .L__const.data@hi, r0",
				Symbols: []SymbolReference{
					{Name: ".L__const.data", Usage: SymbolUsageHi, Global: &globalVar},
				},
			},
			{
				LineNumber: 4,
				Text:       "JMP .LBB0_1, r5",
				Symbols: []SymbolReference{
					{Name: ".LBB0_1", Usage: SymbolUsageFull, Label: &label},
				},
			},
		},
		GlobalsValue: []Global{globalVar},
		LabelsValue:  []Label{label},
	}

	var buf bytes.Buffer
	err := DumpProgramFile(&buf, pf)
	require.NoError(t, err)

	output := buf.String()

	// Symbol references should show as resolved
	assert.Contains(t, output, ".L__const.data@lo (global)")
	assert.Contains(t, output, ".L__const.data@hi (global)")
	assert.Contains(t, output, ".LBB0_1 (label)")

	// Global should show data
	assert.Contains(t, output, ".L__const.data:")
	assert.Contains(t, output, "Type: object")
	assert.Contains(t, output, "Size: 8 bytes")
	assert.Contains(t, output, "Data: 01 02 03 04 05 06 07 08")
}

func TestDumpProgramFile_WithUnresolvedSymbols(t *testing.T) {
	pf := &ProgramFileContents{
		FileNameValue:   "unresolved.cucaracha",
		SourceFileValue: "unresolved.c",
		FunctionsValue: map[string]Function{
			"main": {
				Name:              "main",
				InstructionRanges: []InstructionRange{{Start: 0, Count: 2}},
			},
		},
		InstructionsValue: []Instruction{
			{
				LineNumber: 1,
				Text:       "MOVIMM16L .unknown@lo, r0",
				Symbols: []SymbolReference{
					{Name: ".unknown", Usage: SymbolUsageLo}, // No Function, Global, or Label pointer
				},
			},
			{
				LineNumber: 2,
				Text:       "JMP .missing, r5",
				Symbols: []SymbolReference{
					{Name: ".missing", Usage: SymbolUsageFull},
				},
			},
		},
		GlobalsValue: []Global{},
		LabelsValue:  []Label{},
	}

	var buf bytes.Buffer
	err := DumpProgramFile(&buf, pf)
	require.NoError(t, err)

	output := buf.String()

	// Unresolved symbols should be marked as such
	assert.Contains(t, output, ".unknown@lo (unresolved)")
	assert.Contains(t, output, ".missing (unresolved)")
}

func TestDumpProgramFile_WithResolvedMemory(t *testing.T) {
	addr0 := uint32(0x1000)
	addr1 := uint32(0x1004)
	addr2 := uint32(0x1008)
	globalAddr := uint32(0x2000)

	pf := &ProgramFileContents{
		FileNameValue:   "memory.cucaracha",
		SourceFileValue: "memory.c",
		FunctionsValue: map[string]Function{
			"main": {
				Name:              "main",
				InstructionRanges: []InstructionRange{{Start: 0, Count: 3}},
			},
		},
		InstructionsValue: []Instruction{
			{LineNumber: 1, Text: "MOVIMM16L #10, r0", Address: &addr0},
			{LineNumber: 2, Text: "MOVIMM16L #20, r1", Address: &addr1},
			{LineNumber: 3, Text: "ADD r0, r1, r2", Address: &addr2},
		},
		GlobalsValue: []Global{
			{
				Name:        ".data",
				Size:        4,
				InitialData: []byte{0xDE, 0xAD, 0xBE, 0xEF},
				Type:        GlobalObject,
				Address:     &globalAddr,
			},
		},
		LabelsValue: []Label{},
		MemoryLayoutValue: &MemoryLayout{
			BaseAddress: 0x1000,
			TotalSize:   0x1010,
			CodeSize:    12,
			DataSize:    4,
			CodeStart:   0x1000,
			DataStart:   0x2000,
		},
	}

	var buf bytes.Buffer
	err := DumpProgramFile(&buf, pf)
	require.NoError(t, err)

	output := buf.String()

	// Memory layout should be resolved
	assert.Contains(t, output, "=== Memory Layout ===")
	assert.Contains(t, output, "Base Address: 0x00001000")
	assert.Contains(t, output, "Total Size:   4112 bytes")
	assert.Contains(t, output, "Code Section: 0x00001000 - 0x0000100C (12 bytes)")
	assert.Contains(t, output, "Data Section: 0x00002000 - 0x00002004 (4 bytes)")

	// Instructions should show addresses
	assert.Contains(t, output, "[   0] 0x00001000")
	assert.Contains(t, output, "[   1] 0x00001004")
	assert.Contains(t, output, "[   2] 0x00001008")

	// Global should show address
	assert.Contains(t, output, "Address: 0x00002000")
}

func TestDumpProgramFile_WithDecodedInstructions(t *testing.T) {
	// Create raw instructions with descriptor
	nopDescriptor, err := instructions.Instructions.Instruction(instructions.OpCode_NOP)
	require.NoError(t, err)
	addDescriptor, err := instructions.Instructions.Instruction(instructions.OpCode_ADD)
	require.NoError(t, err)

	rawNop := &instructions.RawInstruction{
		Descriptor:    nopDescriptor,
		OperandValues: []uint64{},
	}

	rawAdd := &instructions.RawInstruction{
		Descriptor:    addDescriptor,
		OperandValues: []uint64{0, 1, 2}, // r0, r1, r2
	}

	pf := &ProgramFileContents{
		FileNameValue:   "decoded.cucaracha",
		SourceFileValue: "decoded.c",
		FunctionsValue: map[string]Function{
			"main": {
				Name:              "main",
				InstructionRanges: []InstructionRange{{Start: 0, Count: 2}},
			},
		},
		InstructionsValue: []Instruction{
			{LineNumber: 1, Text: "NOP", Raw: rawNop},
			{LineNumber: 2, Text: "ADD r0, r1, r2", Raw: rawAdd},
		},
		GlobalsValue: []Global{},
		LabelsValue:  []Label{},
	}

	var buf bytes.Buffer
	err = DumpProgramFile(&buf, pf)
	require.NoError(t, err)

	output := buf.String()

	// Instructions should show raw decoded info
	assert.Contains(t, output, "NOP")
	assert.Contains(t, output, "ADD 0x00, 0x01, 0x02") // Raw instruction string format
}

func TestDumpProgramFile_WithFullyDecodedInstructions(t *testing.T) {
	// Create fully decoded instruction
	addDescriptor, err := instructions.Instructions.Instruction(instructions.OpCode_ADD)
	require.NoError(t, err)

	rawAdd := &instructions.RawInstruction{
		Descriptor:    addDescriptor,
		OperandValues: []uint64{0, 1, 2},
	}
	decodedAdd, err := rawAdd.Decode()
	require.NoError(t, err)

	addr := uint32(0x1000)

	pf := &ProgramFileContents{
		FileNameValue:   "full.cucaracha",
		SourceFileValue: "full.c",
		FunctionsValue: map[string]Function{
			"main": {
				Name:              "main",
				InstructionRanges: []InstructionRange{{Start: 0, Count: 1}},
			},
		},
		InstructionsValue: []Instruction{
			{
				LineNumber:  1,
				Text:        "ADD r0, r1, r2",
				Address:     &addr,
				Raw:         rawAdd,
				Instruction: decodedAdd,
			},
		},
		GlobalsValue: []Global{},
		LabelsValue:  []Label{},
	}

	var buf bytes.Buffer
	err = DumpProgramFile(&buf, pf)
	require.NoError(t, err)

	output := buf.String()

	// Should show address, raw, and text
	assert.Contains(t, output, "[   0] 0x00001000")
	assert.Contains(t, output, "ADD")
}

func TestDumpProgramFile_MultipleFunctions(t *testing.T) {
	pf := &ProgramFileContents{
		FileNameValue:   "multi.cucaracha",
		SourceFileValue: "multi.c",
		FunctionsValue: map[string]Function{
			"main": {
				Name:              "main",
				SourceFile:        "multi.c",
				StartLine:         1,
				EndLine:           5,
				InstructionRanges: []InstructionRange{{Start: 0, Count: 2}},
			},
			"helper": {
				Name:              "helper",
				SourceFile:        "multi.c",
				StartLine:         10,
				EndLine:           15,
				InstructionRanges: []InstructionRange{{Start: 2, Count: 2}},
			},
			"another": {
				Name:              "another",
				SourceFile:        "multi.c",
				StartLine:         20,
				InstructionRanges: []InstructionRange{{Start: 4, Count: 1}},
			},
		},
		InstructionsValue: []Instruction{
			{LineNumber: 2, Text: "MOVIMM16L #1, r0"},
			{LineNumber: 3, Text: "JMP r4, r5"},
			{LineNumber: 11, Text: "MOVIMM16L #2, r1"},
			{LineNumber: 12, Text: "JMP r4, r5"},
			{LineNumber: 21, Text: "NOP"},
		},
		GlobalsValue: []Global{},
		LabelsValue:  []Label{},
	}

	var buf bytes.Buffer
	err := DumpProgramFile(&buf, pf)
	require.NoError(t, err)

	output := buf.String()

	// Functions should be sorted alphabetically
	assert.Contains(t, output, "=== Functions (3) ===")
	assert.Contains(t, output, "another:")
	assert.Contains(t, output, "helper:")
	assert.Contains(t, output, "main:")

	// Check source line formatting
	assert.Contains(t, output, "Source: multi.c:1-5")
	assert.Contains(t, output, "Source: multi.c:10-15")
	assert.Contains(t, output, "Source: multi.c:20") // No end line

	// Function markers in instruction dump
	assert.Contains(t, output, "; function: main")
	assert.Contains(t, output, "; function: helper")
	assert.Contains(t, output, "; function: another")
}

func TestDumpProgramFile_GlobalTypes(t *testing.T) {
	funcAddr := uint32(0x1000)
	objAddr := uint32(0x2000)

	pf := &ProgramFileContents{
		FileNameValue:     "globals.cucaracha",
		SourceFileValue:   "globals.c",
		FunctionsValue:    map[string]Function{},
		InstructionsValue: []Instruction{},
		GlobalsValue: []Global{
			{
				Name:    "my_function",
				Type:    GlobalFunction,
				Size:    0,
				Address: &funcAddr,
			},
			{
				Name:        "my_object",
				Type:        GlobalObject,
				Size:        16,
				InitialData: []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF},
				Address:     &objAddr,
			},
			{
				Name: "unknown_type",
				Type: GlobalUnknown,
				Size: 4,
			},
		},
		LabelsValue: []Label{},
	}

	var buf bytes.Buffer
	err := DumpProgramFile(&buf, pf)
	require.NoError(t, err)

	output := buf.String()

	assert.Contains(t, output, "=== Globals (3) ===")

	// Function global
	assert.Contains(t, output, "my_function:")
	assert.Contains(t, output, "Type: function")
	assert.Contains(t, output, "Address: 0x00001000")

	// Object global with data
	assert.Contains(t, output, "my_object:")
	assert.Contains(t, output, "Type: object")
	assert.Contains(t, output, "Address: 0x00002000")
	assert.Contains(t, output, "Size: 16 bytes")
	assert.Contains(t, output, "Data: 00 11 22 33 44 55 66 77 88 99 AA BB CC DD EE FF")

	// Unknown type
	assert.Contains(t, output, "unknown_type:")
	assert.Contains(t, output, "Type: unknown")
	assert.Contains(t, output, "Address: (unresolved)")
}

func TestDumpProgramFile_LargeGlobalData(t *testing.T) {
	// Create data larger than 32 bytes to test truncation
	largeData := make([]byte, 50)
	for i := range largeData {
		largeData[i] = byte(i)
	}

	pf := &ProgramFileContents{
		FileNameValue:     "large.cucaracha",
		SourceFileValue:   "large.c",
		FunctionsValue:    map[string]Function{},
		InstructionsValue: []Instruction{},
		GlobalsValue: []Global{
			{
				Name:        ".large_data",
				Type:        GlobalObject,
				Size:        50,
				InitialData: largeData,
			},
		},
		LabelsValue: []Label{},
	}

	var buf bytes.Buffer
	err := DumpProgramFile(&buf, pf)
	require.NoError(t, err)

	output := buf.String()

	// Should show truncation message
	assert.Contains(t, output, "... (18 more bytes)")
}

func TestDumpProgramFile_MultipleLabelsAtSameInstruction(t *testing.T) {
	pf := &ProgramFileContents{
		FileNameValue:   "labels.cucaracha",
		SourceFileValue: "labels.c",
		FunctionsValue: map[string]Function{
			"main": {
				Name:              "main",
				InstructionRanges: []InstructionRange{{Start: 0, Count: 2}},
			},
		},
		InstructionsValue: []Instruction{
			{LineNumber: 1, Text: "MOVIMM16L #1, r0"},
			{LineNumber: 2, Text: "JMP r4, r5"},
		},
		GlobalsValue: []Global{},
		LabelsValue: []Label{
			{Name: ".LBB0_start", InstructionIndex: 0},
			{Name: ".LBB0_entry", InstructionIndex: 0},
			{Name: ".Lfunc_end", InstructionIndex: -1}, // Unresolved
		},
	}

	var buf bytes.Buffer
	err := DumpProgramFile(&buf, pf)
	require.NoError(t, err)

	output := buf.String()

	// Multiple labels at same instruction
	assert.Contains(t, output, ".LBB0_start -> instruction #0")
	assert.Contains(t, output, ".LBB0_entry -> instruction #0")
	assert.Contains(t, output, ".Lfunc_end -> (unresolved)")

	// Both labels should appear before instruction 0 in dump
	assert.Contains(t, output, ".LBB0_start:")
	assert.Contains(t, output, ".LBB0_entry:")
}

func TestDumpProgramFile_FunctionReferencedInSymbol(t *testing.T) {
	mainFunc := Function{
		Name:              "main",
		InstructionRanges: []InstructionRange{{Start: 0, Count: 2}},
	}
	helperFunc := Function{
		Name:              "helper",
		InstructionRanges: []InstructionRange{{Start: 2, Count: 1}},
	}

	pf := &ProgramFileContents{
		FileNameValue:   "funcref.cucaracha",
		SourceFileValue: "funcref.c",
		FunctionsValue: map[string]Function{
			"main":   mainFunc,
			"helper": helperFunc,
		},
		InstructionsValue: []Instruction{
			{
				LineNumber: 1,
				Text:       "MOVIMM16L helper@lo, r0",
				Symbols: []SymbolReference{
					{Name: "helper", Usage: SymbolUsageLo, Function: &helperFunc},
				},
			},
			{
				LineNumber: 2,
				Text:       "MOVIMM16H helper@hi, r0",
				Symbols: []SymbolReference{
					{Name: "helper", Usage: SymbolUsageHi, Function: &helperFunc},
				},
			},
			{
				LineNumber: 10,
				Text:       "JMP r4, r5",
			},
		},
		GlobalsValue: []Global{},
		LabelsValue:  []Label{},
	}

	var buf bytes.Buffer
	err := DumpProgramFile(&buf, pf)
	require.NoError(t, err)

	output := buf.String()

	// Function references should be shown correctly
	assert.Contains(t, output, "helper@lo (func)")
	assert.Contains(t, output, "helper@hi (func)")
}

func TestDumpProgramFile_NoSourceFile(t *testing.T) {
	pf := &ProgramFileContents{
		FileNameValue:   "nosource.cucaracha",
		SourceFileValue: "", // Empty source file
		FunctionsValue: map[string]Function{
			"main": {
				Name:              "main",
				InstructionRanges: []InstructionRange{{Start: 0, Count: 1}},
			},
		},
		InstructionsValue: []Instruction{
			{Text: "NOP"}, // No line number
		},
		GlobalsValue: []Global{},
		LabelsValue:  []Label{},
	}

	var buf bytes.Buffer
	err := DumpProgramFile(&buf, pf)
	require.NoError(t, err)

	output := buf.String()

	assert.Contains(t, output, "Source: ")
	// Should not have "; line" comment for instruction without line number
	assert.NotContains(t, output, "; line 0")
}

func TestDumpProgramFile_FunctionWithMultipleRanges(t *testing.T) {
	pf := &ProgramFileContents{
		FileNameValue:   "multirange.cucaracha",
		SourceFileValue: "multirange.c",
		FunctionsValue: map[string]Function{
			"fragmented": {
				Name:       "fragmented",
				SourceFile: "multirange.c",
				InstructionRanges: []InstructionRange{
					{Start: 0, Count: 2},
					{Start: 5, Count: 3},
				},
			},
		},
		InstructionsValue: []Instruction{
			{Text: "MOVIMM16L #1, r0"},
			{Text: "JMP r4, r5"},
			{Text: "NOP"},
			{Text: "NOP"},
			{Text: "NOP"},
			{Text: "MOVIMM16L #2, r1"},
			{Text: "ADD r0, r1, r2"},
			{Text: "JMP r4, r5"},
		},
		GlobalsValue: []Global{},
		LabelsValue:  []Label{},
	}

	var buf bytes.Buffer
	err := DumpProgramFile(&buf, pf)
	require.NoError(t, err)

	output := buf.String()

	// Should show multiple ranges
	assert.Contains(t, output, "Instruction Ranges: [0..1], [5..7]")
}

func TestDumpProgramFile_CompleteProgram(t *testing.T) {
	// A fully resolved program with all features
	addr0 := uint32(0x1000)
	addr1 := uint32(0x1004)
	addr2 := uint32(0x1008)
	addr3 := uint32(0x100C)
	globalAddr := uint32(0x2000)

	addDescriptor, err := instructions.Instructions.Instruction(instructions.OpCode_ADD)
	require.NoError(t, err)
	rawAdd := &instructions.RawInstruction{
		Descriptor:    addDescriptor,
		OperandValues: []uint64{0, 1, 2},
	}
	decodedAdd, _ := rawAdd.Decode()

	globalData := Global{
		Name:        ".L__const",
		Type:        GlobalObject,
		Size:        4,
		InitialData: []byte{0x01, 0x02, 0x03, 0x04},
		Address:     &globalAddr,
	}

	label := Label{
		Name:             ".LBB0_1",
		InstructionIndex: 2,
	}

	pf := &ProgramFileContents{
		FileNameValue:   "complete.cucaracha",
		SourceFileValue: "complete.c",
		FunctionsValue: map[string]Function{
			"main": {
				Name:              "main",
				SourceFile:        "complete.c",
				StartLine:         1,
				EndLine:           20,
				InstructionRanges: []InstructionRange{{Start: 0, Count: 4}},
			},
		},
		InstructionsValue: []Instruction{
			{
				LineNumber: 2,
				Text:       "MOVIMM16L .L__const@lo, r0",
				Address:    &addr0,
				Symbols: []SymbolReference{
					{Name: ".L__const", Usage: SymbolUsageLo, Global: &globalData},
				},
			},
			{
				LineNumber: 3,
				Text:       "MOVIMM16H .L__const@hi, r0",
				Address:    &addr1,
				Symbols: []SymbolReference{
					{Name: ".L__const", Usage: SymbolUsageHi, Global: &globalData},
				},
			},
			{
				LineNumber:  4,
				Text:        "ADD r0, r1, r2",
				Address:     &addr2,
				Raw:         rawAdd,
				Instruction: decodedAdd,
			},
			{
				LineNumber: 5,
				Text:       "JMP .LBB0_1, r5",
				Address:    &addr3,
				Symbols: []SymbolReference{
					{Name: ".LBB0_1", Usage: SymbolUsageFull, Label: &label},
				},
			},
		},
		GlobalsValue: []Global{globalData},
		LabelsValue:  []Label{label},
		MemoryLayoutValue: &MemoryLayout{
			BaseAddress: 0x1000,
			TotalSize:   0x1004,
			CodeSize:    16,
			DataSize:    4,
			CodeStart:   0x1000,
			DataStart:   0x2000,
		},
	}

	var buf bytes.Buffer
	err = DumpProgramFile(&buf, pf)
	require.NoError(t, err)

	output := buf.String()

	// Verify all sections are present and complete
	assert.Contains(t, output, "=== Program File ===")
	assert.Contains(t, output, "File: complete.cucaracha")
	assert.Contains(t, output, "Source: complete.c")

	assert.Contains(t, output, "=== Memory Layout ===")
	assert.Contains(t, output, "Base Address: 0x00001000")
	assert.Contains(t, output, "Code Section: 0x00001000")
	assert.Contains(t, output, "Data Section: 0x00002000")

	assert.Contains(t, output, "=== Functions (1) ===")
	assert.Contains(t, output, "main:")
	assert.Contains(t, output, "Source: complete.c:1-20")
	assert.Contains(t, output, "Instruction Ranges: [0..3]")

	assert.Contains(t, output, "=== Labels (1) ===")
	assert.Contains(t, output, ".LBB0_1 -> instruction #2")

	assert.Contains(t, output, "=== Globals (1) ===")
	assert.Contains(t, output, ".L__const:")
	assert.Contains(t, output, "Type: object")
	assert.Contains(t, output, "Address: 0x00002000")
	assert.Contains(t, output, "Data: 01 02 03 04")

	assert.Contains(t, output, "=== Instructions (4) ===")
	assert.Contains(t, output, "; function: main")
	assert.Contains(t, output, "[   0] 0x00001000")
	assert.Contains(t, output, ".L__const@lo (global)")
	assert.Contains(t, output, ".LBB0_1:")
	assert.Contains(t, output, "ADD 0x00, 0x01, 0x02")
	assert.Contains(t, output, ".LBB0_1 (label)")
}
