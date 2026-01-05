package llvm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NOTE: Always write tests with testify (assert/require packages)

func TestParseAssemblyFile_Comprehensive(t *testing.T) {
	// Create a comprehensive test assembly file with multiple features
	content := `	.text
	.file	"arrays.c"
	.globl	main
	.type	main,@function
main:
	.loc	1 5 0
	MOVIMM16L #0, r0
	MOVIMM16L #4, r1
	MOVIMM16L .LBB0_1@lo, r4
	MOVIMM16H .LBB0_1@hi, r4
	JMP r4, r5
.LBB0_1:
	.loc	1 10 0
	MOVIMM16L #8, r0
	ADD r0, r1, r2
	MOVIMM16L .LBB0_2@lo, r4
	MOVIMM16H .LBB0_2@hi, r4
	JMP r4, r5
.LBB0_2:
	.loc	1 15 0
	MOVIMM16L #12, r0
	SUB r2, r1, r3
	MOVIMM16L .LBB0_3@lo, r4
	MOVIMM16H .LBB0_3@hi, r4
	JMP r4, r5
.LBB0_3:
	.loc	1 20 0
	MOVIMM16L #16, r0
	MUL r0, r1, r2
	MOVIMM16L .LBB0_4@lo, r4
	MOVIMM16H .LBB0_4@hi, r4
	JMP r4, r5
.LBB0_4:
	.loc	1 25 0
	MOVIMM16L #20, r0
	DIV r2, r1, r3
	MOVIMM16L .LBB0_5@lo, r4
	MOVIMM16H .LBB0_5@hi, r4
	JMP r4, r5
.LBB0_5:
	.loc	1 30 0
	MOVIMM16L #24, r0
	CMP r0, r1, r2
	MOVIMM16L .LBB0_6@lo, r4
	MOVIMM16H .LBB0_6@hi, r4
	CJMP r2, r4, r5
.LBB0_6:
	.loc	1 35 0
	MOVIMM16L #28, r0
	LD r0, r1
	MOVIMM16L .LBB0_7@lo, r4
	MOVIMM16H .LBB0_7@hi, r4
	JMP r4, r5
.LBB0_7:
	.loc	1 40 0
	MOVIMM16L #32, r0
	ST r1, r0
	MOVIMM16L .LBB0_8@lo, r4
	MOVIMM16H .LBB0_8@hi, r4
	JMP r4, r5
.LBB0_8:
	.loc	1 45 0
	MOVIMM16L #36, r0
	LSL r0, r1, r2
	MOVIMM16L .LBB0_9@lo, r4
	MOVIMM16H .LBB0_9@hi, r4
	JMP r4, r5
.LBB0_9:
	.loc	1 50 0
	MOVIMM16L #40, r0
	LSR r2, r1, r3
	MOVIMM16L .LBB0_10@lo, r4
	MOVIMM16H .LBB0_10@hi, r4
	JMP r4, r5
.LBB0_10:
	.loc	1 55 0
	MOVIMM16L #44, r0
	MOV r0, r1
	MOVIMM16L .LBB0_11@lo, r4
	MOVIMM16H .LBB0_11@hi, r4
	JMP r4, r5
.LBB0_11:
	.loc	1 60 0
	MOVIMM16L #48, r0
	NOP
	MOVIMM16L .LBB0_12@lo, r4
	MOVIMM16H .LBB0_12@hi, r4
	JMP r4, r5
.LBB0_12:
	.loc	1 65 0
	MOVIMM16L #36, r0
	bx lr
.Lfunc_end0:
	.size	main, .Lfunc_end0-main

	.type	.L__const.main.arr,@object
	.section	.rodata
.L__const.main.arr:
	.long	1
	.long	2
	.long	3
	.long	4
	.size	.L__const.main.arr, 16
`

	// Write to a temp file
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "arrays.cucaracha")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err, "failed to write temp file")

	f, err := ParseAssemblyFile(path)
	require.NoError(t, err, "ParseAssemblyFile failed")

	// ==========================================
	// 1. Assembly file name is correct
	// ==========================================
	assert.Equal(t, path, f.FileName(), "assembly file name")

	// ==========================================
	// 2. Source file name is correct
	// ==========================================
	assert.Equal(t, "arrays.c", f.SourceFile(), "source file name")

	// ==========================================
	// 3. All functions are parsed
	// ==========================================
	functions := f.Functions()
	expectedFunctions := []string{"main"}
	assert.Len(t, functions, len(expectedFunctions), "number of functions")
	for _, name := range expectedFunctions {
		assert.Contains(t, functions, name, "function %s should be present", name)
	}

	// ==========================================
	// 4. All labels are parsed
	// ==========================================
	labels := f.Labels()
	expectedLabels := []string{
		".LBB0_1", ".LBB0_2", ".LBB0_3", ".LBB0_4",
		".LBB0_5", ".LBB0_6", ".LBB0_7", ".LBB0_8",
		".LBB0_9", ".LBB0_10", ".LBB0_11", ".LBB0_12",
		".Lfunc_end0",
	}
	assert.Len(t, labels, len(expectedLabels), "number of labels")
	labelNames := make(map[string]bool)
	for _, lbl := range labels {
		labelNames[lbl.Name] = true
	}
	for _, name := range expectedLabels {
		assert.True(t, labelNames[name], "label %s should be present", name)
	}
	// No unexpected labels
	for name := range labelNames {
		found := false
		for _, expected := range expectedLabels {
			if name == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "unexpected label %s", name)
	}

	// ==========================================
	// 4b. Function names are NOT in labels
	// ==========================================
	for funcName := range functions {
		assert.False(t, labelNames[funcName], "function '%s' should not be in labels", funcName)
	}

	// ==========================================
	// 5. All globals are parsed
	// ==========================================
	globals := f.Globals()
	expectedGlobals := []string{".L__const.main.arr"}
	assert.Len(t, globals, len(expectedGlobals), "number of globals")
	for _, name := range expectedGlobals {
		found := false
		for _, g := range globals {
			if g.Name == name {
				found = true
				break
			}
		}
		assert.True(t, found, "global %s should be present", name)
	}

	// ==========================================
	// 6. All instructions are parsed
	// ==========================================
	instructions := f.Instructions()
	// Our test file has 58 instructions, so verify we have a reasonable number
	assert.Greater(t, len(instructions), 50, "total instructions should be > 50")

	// ==========================================
	// Per-function validation
	// ==========================================
	t.Run("Functions", func(t *testing.T) {
		functions := f.Functions()
		mainFunc, ok := functions["main"]
		require.True(t, ok, "main function should exist")

		// Function name is correct
		assert.Equal(t, "main", mainFunc.Name, "function name")

		// Function source file
		assert.Equal(t, "arrays.c", mainFunc.SourceFile, "main.SourceFile")

		// Function line numbers
		assert.NotZero(t, mainFunc.StartLine, "main StartLine should not be zero")
		assert.NotZero(t, mainFunc.EndLine, "main EndLine should not be zero")
		assert.Less(t, mainFunc.StartLine, mainFunc.EndLine, "main StartLine < EndLine")

		// Function instruction ranges are correct
		assert.NotEmpty(t, mainFunc.InstructionRanges, "main should have instruction ranges")

		// Helper to flatten all instructions for a function
		getFuncInstructions := func(allInstructions []mc.Instruction, fn mc.Function) []mc.Instruction {
			var out []mc.Instruction
			for _, r := range fn.InstructionRanges {
				out = append(out, allInstructions[r.Start:r.Start+r.Count]...)
			}
			return out
		}

		instructions := f.Instructions()
		mainInstructions := getFuncInstructions(instructions, mainFunc)
		assert.NotEmpty(t, mainInstructions, "main function should have instructions")

		// Validate instruction ranges point to valid indices
		for i, r := range mainFunc.InstructionRanges {
			assert.GreaterOrEqual(t, r.Start, 0, "instruction range %d start >= 0", i)
			assert.Greater(t, r.Count, 0, "instruction range %d count > 0", i)
			assert.LessOrEqual(t, r.Start+r.Count, len(instructions),
				"instruction range %d end <= total instructions", i)
		}
	})

	// ==========================================
	// Per-label validation
	// ==========================================
	t.Run("Labels", func(t *testing.T) {
		labels := f.Labels()
		instructions := f.Instructions()

		for _, lbl := range labels {
			// Label name is correct (not empty)
			assert.NotEmpty(t, lbl.Name, "label name should not be empty")

			// InstructionIndex should be valid if >= 0
			if lbl.InstructionIndex >= 0 {
				assert.Less(t, lbl.InstructionIndex, len(instructions),
					"label '%s' instruction index should be valid", lbl.Name)
			}
		}

		// Check specific labels point to correct instructions
		labelMap := make(map[string]mc.Label)
		for _, lbl := range labels {
			labelMap[lbl.Name] = lbl
		}

		// .LBB0_1 should point to the first instruction after the JMP (line 40)
		if lbl, ok := labelMap[".LBB0_1"]; ok && lbl.InstructionIndex >= 0 {
			assert.Equal(t, "MOVIMM16L #8, r0", instructions[lbl.InstructionIndex].Text,
				".LBB0_1 should point to first instruction of for.cond block")
		}

		// .LBB0_12 should point to the return block
		if lbl, ok := labelMap[".LBB0_12"]; ok && lbl.InstructionIndex >= 0 {
			assert.Equal(t, "MOVIMM16L #36, r0", instructions[lbl.InstructionIndex].Text,
				".LBB0_12 should point to first instruction of return block")
		}
	})

	// ==========================================
	// Per-global validation
	// ==========================================
	t.Run("Globals", func(t *testing.T) {
		globals := f.Globals()

		var arrGlobal *mc.Global
		for i := range globals {
			if globals[i].Name == ".L__const.main.arr" {
				arrGlobal = &globals[i]
				break
			}
		}
		require.NotNil(t, arrGlobal, "global .L__const.main.arr should exist")

		// Global name is correct
		assert.Equal(t, ".L__const.main.arr", arrGlobal.Name, "global name")

		// Type is correct
		assert.Equal(t, mc.GlobalObject, arrGlobal.Type, "global type should be GlobalObject")

		// Size is correct (4 int32s = 16 bytes)
		assert.Equal(t, 16, arrGlobal.Size, "global size should be 16 bytes")

		// Initialization value is correct (array of 4 int32: 1, 2, 3, 4)
		expectedData := []byte{
			1, 0, 0, 0, // 1
			2, 0, 0, 0, // 2
			3, 0, 0, 0, // 3
			4, 0, 0, 0, // 4
		}
		assert.Equal(t, expectedData, arrGlobal.InitialData, "global initial data")

		// LLVM assembly parser does not resolve memory addresses
		for i, g := range globals {
			assert.Nil(t, g.Address, "global %d Address should be nil (text-only parser)", i)
		}
	})

	// ==========================================
	// Per-instruction validation
	// ==========================================
	t.Run("Instructions", func(t *testing.T) {
		instructions := f.Instructions()

		// Define expected instructions with line numbers and symbol references
		// Line numbers are assembly file line numbers, not source .loc line numbers
		expectedInstructions := []struct {
			Text       string
			LineNumber int
			Symbols    []string
		}{
			{"MOVIMM16L #0, r0", 7, []string{}},
			{"MOVIMM16L #4, r1", 8, []string{}},
			{"MOVIMM16L .LBB0_1@lo, r4", 9, []string{".LBB0_1"}},
			{"MOVIMM16H .LBB0_1@hi, r4", 10, []string{".LBB0_1"}},
			{"JMP r4, r5", 11, []string{}},
			// .LBB0_1: (line 12)
			{"MOVIMM16L #8, r0", 14, []string{}},
			{"ADD r0, r1, r2", 15, []string{}},
			{"MOVIMM16L .LBB0_2@lo, r4", 16, []string{".LBB0_2"}},
			{"MOVIMM16H .LBB0_2@hi, r4", 17, []string{".LBB0_2"}},
			{"JMP r4, r5", 18, []string{}},
		}

		require.GreaterOrEqual(t, len(instructions), len(expectedInstructions),
			"should have at least %d instructions", len(expectedInstructions))

		for i, expected := range expectedInstructions {
			inst := instructions[i]

			// Instruction location is correct
			assert.Equal(t, expected.LineNumber, inst.LineNumber,
				"instruction %d line number", i)

			// Instruction text is correct
			assert.Equal(t, expected.Text, inst.Text,
				"instruction %d text", i)

			// References to other symbols are correct
			require.Len(t, inst.Symbols, len(expected.Symbols),
				"instruction %d symbol count", i)
			for j, expectedSym := range expected.Symbols {
				assert.Equal(t, expectedSym, inst.Symbols[j].Name,
					"instruction %d symbol %d name", i, j)
				// LLVM parser doesn't resolve symbol references
				assert.Equal(t, mc.SymbolKindUnknown, inst.Symbols[j].Kind(),
					"instruction %d symbol %d kind should be unknown (unresolved)", i, j)
			}
		}

		// Validate all instructions have valid data
		for i, inst := range instructions {
			assert.NotZero(t, inst.LineNumber, "instruction %d LineNumber should not be zero", i)
			assert.NotEmpty(t, inst.Text, "instruction %d Text should not be empty", i)
			// LLVM assembly parser only provides text representation
			assert.Nil(t, inst.Raw, "instruction %d Raw should be nil (text-only parser)", i)
			assert.Nil(t, inst.Instruction, "instruction %d Instruction should be nil (text-only parser)", i)
			assert.Nil(t, inst.Address, "instruction %d Address should be nil (text-only parser)", i)
		}

		// Check some instructions with symbol references
		symbolRefCount := 0
		for _, inst := range instructions {
			if len(inst.Symbols) > 0 {
				symbolRefCount++
				for _, sym := range inst.Symbols {
					assert.NotEmpty(t, sym.Name, "symbol reference name should not be empty")
					// LLVM parser doesn't resolve symbol references
					assert.True(t, sym.Unresolved(), "symbol reference '%s' should be unresolved (text-only parser)", sym.Name)
				}
			}
		}
		assert.Greater(t, symbolRefCount, 10, "should have multiple instructions with symbol references")
	})

	// ==========================================
	// Cross-validation: instruction ranges cover all function instructions
	// ==========================================
	t.Run("InstructionRangesCoverage", func(t *testing.T) {
		functions := f.Functions()
		mainFunc, ok := functions["main"]
		require.True(t, ok, "main function should exist")

		totalCovered := 0
		for _, r := range mainFunc.InstructionRanges {
			totalCovered += r.Count
		}
		// Main function should cover most of the instructions (we have 62)
		assert.Greater(t, totalCovered, 50, "main function should cover > 50 instructions")
	})
}

func TestParseAssemblyFile_FunctionsNotInLabels(t *testing.T) {
	// Create a test file with multiple functions
	content := `	.text
	.file	"multi.c"
	.globl	main
	.type	main,@function
main:
	MOVIMM16L #1, r0
	MOVIMM16L .LBB0_1@lo, r4
	MOVIMM16H .LBB0_1@hi, r4
	JMP r4, r5
.LBB0_1:
	MOVIMM16L #2, r0
	bx lr
.Lfunc_end0:
	.size	main, .Lfunc_end0-main

	.globl	helper
	.type	helper,@function
helper:
	MOVIMM16L #10, r0
	bx lr
.Lfunc_end1:
	.size	helper, .Lfunc_end1-helper
`

	// Write to a temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "multi.cucaracha")
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(t, err, "failed to write temp file")

	// Parse the file
	f, err := ParseAssemblyFile(tmpFile)
	require.NoError(t, err, "ParseAssemblyFile failed")

	// Get functions and labels
	functions := f.Functions()
	labels := f.Labels()

	// Verify functions are parsed
	require.Contains(t, functions, "main", "main function should be parsed")
	require.Contains(t, functions, "helper", "helper function should be parsed")

	// Build a set of label names
	labelNames := make(map[string]bool)
	for _, lbl := range labels {
		labelNames[lbl.Name] = true
	}

	// Verify function names are NOT in labels
	assert.False(t, labelNames["main"], "function 'main' should NOT be in labels")
	assert.False(t, labelNames["helper"], "function 'helper' should NOT be in labels")

	// Verify actual labels ARE in labels
	assert.True(t, labelNames[".LBB0_1"], "label '.LBB0_1' should be in labels")
	assert.True(t, labelNames[".Lfunc_end0"], "label '.Lfunc_end0' should be in labels")
	assert.True(t, labelNames[".Lfunc_end1"], "label '.Lfunc_end1' should be in labels")

	// Verify label count is exactly what we expect
	assert.Len(t, labels, 3, "should have exactly 3 labels (not including function names)")
}
