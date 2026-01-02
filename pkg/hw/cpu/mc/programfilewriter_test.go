package mc

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteProgramFile_EmptyProgram(t *testing.T) {
	pf := &ProgramFileContents{
		FileNameValue:     "empty.cucaracha",
		SourceFileValue:   "empty.c",
		FunctionsValue:    map[string]Function{},
		InstructionsValue: []Instruction{},
		GlobalsValue:      []Global{},
		LabelsValue:       []Label{},
	}

	var buf bytes.Buffer
	err := WriteProgramFile(&buf, pf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, ".text")
	assert.Contains(t, output, `.file	"empty.c"`)
}

func TestWriteProgramFile_SingleFunction(t *testing.T) {
	pf := &ProgramFileContents{
		FileNameValue:   "main.cucaracha",
		SourceFileValue: "main.c",
		FunctionsValue: map[string]Function{
			"main": {
				Name:       "main",
				SourceFile: "main.c",
				StartLine:  1,
				EndLine:    5,
				InstructionRanges: []InstructionRange{
					{Start: 0, Count: 3},
				},
			},
		},
		InstructionsValue: []Instruction{
			{LineNumber: 1, Text: "MOVIMM16L #10, r0"},
			{LineNumber: 2, Text: "MOVIMM16L #20, r1"},
			{LineNumber: 3, Text: "ADD r0, r1, r2"},
		},
		GlobalsValue: []Global{},
		LabelsValue:  []Label{},
	}

	var buf bytes.Buffer
	err := WriteProgramFile(&buf, pf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, ".text")
	assert.Contains(t, output, `.file	"main.c"`)
	assert.Contains(t, output, ".globl\tmain")
	assert.Contains(t, output, ".type\tmain,@function")
	assert.Contains(t, output, "main:")
	assert.Contains(t, output, "\tMOVIMM16L #10, r0")
	assert.Contains(t, output, "\tMOVIMM16L #20, r1")
	assert.Contains(t, output, "\tADD r0, r1, r2")
	assert.Contains(t, output, ".Lfunc_endmain:")
	assert.Contains(t, output, ".size\tmain, .Lfunc_endmain-main")
}

func TestWriteProgramFile_WithLabels(t *testing.T) {
	pf := &ProgramFileContents{
		FileNameValue:   "loop.cucaracha",
		SourceFileValue: "loop.c",
		FunctionsValue: map[string]Function{
			"main": {
				Name:       "main",
				SourceFile: "loop.c",
				StartLine:  1,
				EndLine:    10,
				InstructionRanges: []InstructionRange{
					{Start: 0, Count: 5},
				},
			},
		},
		InstructionsValue: []Instruction{
			{LineNumber: 1, Text: "MOVIMM16L #0, r0"},
			{LineNumber: 2, Text: "MOVIMM16L #10, r1"},
			{LineNumber: 3, Text: "CMP r0, r1, r2"},
			{LineNumber: 4, Text: "MOVIMM16L #1, r3"},
			{LineNumber: 5, Text: "ADD r0, r3, r0"},
		},
		GlobalsValue: []Global{},
		LabelsValue: []Label{
			{Name: ".LBB0_1", InstructionIndex: 2},
		},
	}

	var buf bytes.Buffer
	err := WriteProgramFile(&buf, pf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, ".LBB0_1:")
	// Verify label appears before instruction at index 2
	labelPos := strings.Index(output, ".LBB0_1:")
	instrPos := strings.Index(output, "\tCMP r0, r1, r2")
	assert.True(t, labelPos < instrPos, "Label should appear before its instruction")
}

func TestWriteProgramFile_WithGlobals(t *testing.T) {
	pf := &ProgramFileContents{
		FileNameValue:     "data.cucaracha",
		SourceFileValue:   "data.c",
		FunctionsValue:    map[string]Function{},
		InstructionsValue: []Instruction{},
		GlobalsValue: []Global{
			{
				Name:        ".L__const.main.arr",
				Size:        20,
				InitialData: []byte{1, 0, 0, 0, 2, 0, 0, 0, 3, 0, 0, 0, 4, 0, 0, 0, 5, 0, 0, 0},
				Type:        GlobalObject,
			},
		},
		LabelsValue: []Label{},
	}

	var buf bytes.Buffer
	err := WriteProgramFile(&buf, pf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, ".type\t.L__const.main.arr,@object")
	assert.Contains(t, output, ".section\t.rodata,\"a\",@progbits")
	assert.Contains(t, output, ".p2align\t2, 0x0")
	assert.Contains(t, output, ".L__const.main.arr:")
	assert.Contains(t, output, ".long\t1")
	assert.Contains(t, output, ".long\t2")
	assert.Contains(t, output, ".long\t3")
	assert.Contains(t, output, ".long\t4")
	assert.Contains(t, output, ".long\t5")
	assert.Contains(t, output, ".size\t.L__const.main.arr, 20")
}

func TestWriteProgramFile_GlobalWithBytes(t *testing.T) {
	pf := &ProgramFileContents{
		FileNameValue:     "bytes.cucaracha",
		SourceFileValue:   "bytes.c",
		FunctionsValue:    map[string]Function{},
		InstructionsValue: []Instruction{},
		GlobalsValue: []Global{
			{
				Name:        ".L__const.main.str",
				Size:        5,
				InitialData: []byte{72, 101, 108, 108, 111}, // "Hello"
				Type:        GlobalObject,
			},
		},
		LabelsValue: []Label{},
	}

	var buf bytes.Buffer
	err := WriteProgramFile(&buf, pf)
	require.NoError(t, err)

	output := buf.String()
	// 4 bytes as .long, 1 remaining byte
	assert.Contains(t, output, ".long\t")
	assert.Contains(t, output, ".byte\t111") // 'o' = 111
}

func TestWriteProgramFile_MultipleFunctions(t *testing.T) {
	pf := &ProgramFileContents{
		FileNameValue:   "multi.cucaracha",
		SourceFileValue: "multi.c",
		FunctionsValue: map[string]Function{
			"main": {
				Name:       "main",
				SourceFile: "multi.c",
				StartLine:  1,
				EndLine:    3,
				InstructionRanges: []InstructionRange{
					{Start: 0, Count: 2},
				},
			},
			"helper": {
				Name:       "helper",
				SourceFile: "multi.c",
				StartLine:  5,
				EndLine:    7,
				InstructionRanges: []InstructionRange{
					{Start: 2, Count: 2},
				},
			},
		},
		InstructionsValue: []Instruction{
			{LineNumber: 1, Text: "MOVIMM16L #1, r0"},
			{LineNumber: 2, Text: "bx lr"},
			{LineNumber: 5, Text: "MOVIMM16L #2, r0"},
			{LineNumber: 6, Text: "bx lr"},
		},
		GlobalsValue: []Global{},
		LabelsValue:  []Label{},
	}

	var buf bytes.Buffer
	err := WriteProgramFile(&buf, pf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, ".globl\tmain")
	assert.Contains(t, output, ".globl\thelper")
	assert.Contains(t, output, "main:")
	assert.Contains(t, output, "helper:")

	// Functions should be in order by instruction index
	mainPos := strings.Index(output, "main:")
	helperPos := strings.Index(output, "helper:")
	assert.True(t, mainPos < helperPos, "main should appear before helper")
}

func TestWriteProgramFile_NoSourceFile(t *testing.T) {
	pf := &ProgramFileContents{
		FileNameValue:     "nosrc.cucaracha",
		SourceFileValue:   "",
		FunctionsValue:    map[string]Function{},
		InstructionsValue: []Instruction{},
		GlobalsValue:      []Global{},
		LabelsValue:       []Label{},
	}

	var buf bytes.Buffer
	err := WriteProgramFile(&buf, pf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, ".text")
	assert.NotContains(t, output, ".file")
}

func TestWriteProgramFile_GlobalWithZero(t *testing.T) {
	pf := &ProgramFileContents{
		FileNameValue:     "zero.cucaracha",
		SourceFileValue:   "zero.c",
		FunctionsValue:    map[string]Function{},
		InstructionsValue: []Instruction{},
		GlobalsValue: []Global{
			{
				Name:        ".L__bss.data",
				Size:        16,
				InitialData: []byte{}, // No initial data
				Type:        GlobalObject,
			},
		},
		LabelsValue: []Label{},
	}

	var buf bytes.Buffer
	err := WriteProgramFile(&buf, pf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, ".zero\t16")
}

func TestWriteProgramFile_FunctionLabelNotDuplicated(t *testing.T) {
	// When a label has the same name as a function, it shouldn't be written twice
	pf := &ProgramFileContents{
		FileNameValue:   "nodup.cucaracha",
		SourceFileValue: "nodup.c",
		FunctionsValue: map[string]Function{
			"main": {
				Name:       "main",
				SourceFile: "nodup.c",
				StartLine:  1,
				EndLine:    3,
				InstructionRanges: []InstructionRange{
					{Start: 0, Count: 2},
				},
			},
		},
		InstructionsValue: []Instruction{
			{LineNumber: 1, Text: "MOVIMM16L #1, r0"},
			{LineNumber: 2, Text: "bx lr"},
		},
		GlobalsValue: []Global{},
		LabelsValue: []Label{
			{Name: "main", InstructionIndex: 0}, // Label with same name as function
		},
	}

	var buf bytes.Buffer
	err := WriteProgramFile(&buf, pf)
	require.NoError(t, err)

	output := buf.String()
	// Count occurrences of "main:" as a line (should be exactly 1)
	// We need to count lines that are exactly "main:" (possibly with leading whitespace)
	count := 0
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "main:" {
			count++
		}
	}
	assert.Equal(t, 1, count, "Function label should appear exactly once")
}
