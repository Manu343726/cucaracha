package llvm

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NOTE: Always write tests with testify (assert/require packages)

func TestParseAssemblyFile_Comprehensive(t *testing.T) {
	// Find the test data file using a path relative to this test source file
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller failed")
	path := filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "..", "llvm-project", "cucaracha-tests", "test_output", "arrays.cucaracha")
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
	// Total instructions should be > 100 (the file has ~140 instructions)
	assert.Greater(t, len(instructions), 100, "total instructions should be > 100")

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

		// Size is correct
		assert.Equal(t, 20, arrGlobal.Size, "global size should be 20 bytes")

		// Initialization value is correct (array of 5 int32: 1, 2, 3, 4, 5)
		expectedData := []byte{
			1, 0, 0, 0, // 1
			2, 0, 0, 0, // 2
			3, 0, 0, 0, // 3
			4, 0, 0, 0, // 4
			5, 0, 0, 0, // 5
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
		// First instructions of main function (before first label)
		// NOTE: LR is now callee-saved, so it's saved to the stack at function entry
		expectedInstructions := []struct {
			Text       string
			LineNumber int
			Symbols    []string
		}{
			{"MOVIMM16L #44, r4", 7, []string{}},                  // Stack frame size (includes LR)
			{"SUB sp, r4, sp", 8, []string{}},                     // Allocate stack frame
			{"MOVIMM16L #40, r4", 9, []string{}},                  // Offset for LR save
			{"ADD sp, r4, r5", 10, []string{}},                    // Compute LR save address
			{"ST lr, r5", 11, []string{}},                         // Save LR to stack
			{"MOVIMM16L #36, r0", 12, []string{}},                 // Offset for return value
			{"MOVIMM16H #0, r0", 13, []string{}},                  // High bits of offset
			{"ADD sp, r0, r1", 14, []string{}},                    // Compute return value address
			{"MOVIMM16L #0, r0", 15, []string{}},                  // Initialize return value to 0
			{"ST r0, r1", 16, []string{}},                         // Store return value
			{"MOVIMM16L #16, r1", 17, []string{}},                 // Array offset
			{"MOVIMM16H #0, r1", 18, []string{}},                  // High bits
			{"ADD sp, r1, r2", 19, []string{}},                    // Array base address
			{"MOVIMM16L #5, r1", 20, []string{}},                  // arr[4] = 5 (inlined constant)
			{"MOVIMM16L #16, r4", 21, []string{}},                 // Offset for arr[4]
			{"ADD r2, r4, r5", 22, []string{}},                    // Address of arr[4]
			{"ST r1, r5", 23, []string{}},                         // Store arr[4]
			{"MOVIMM16L #4, r1", 24, []string{}},                  // arr[3] = 4 (inlined constant)
			{"MOVIMM16L #12, r4", 25, []string{}},                 // Offset for arr[3]
			{"ADD r2, r4, r5", 26, []string{}},                    // Address of arr[3]
			{"ST r1, r5", 27, []string{}},                         // Store arr[3]
			{"MOVIMM16L #3, r1", 28, []string{}},                  // arr[2] = 3 (inlined constant)
			{"MOVIMM16L #8, r4", 29, []string{}},                  // Offset for arr[2]
			{"ADD r2, r4, r5", 30, []string{}},                    // Address of arr[2]
			{"ST r1, r5", 31, []string{}},                         // Store arr[2]
			{"MOVIMM16L #2, r1", 32, []string{}},                  // arr[1] = 2 (inlined constant)
			{"MOVIMM16L #4, r4", 33, []string{}},                  // Offset for arr[1]
			{"ADD r2, r4, r5", 34, []string{}},                    // Address of arr[1]
			{"ST r1, r5", 35, []string{}},                         // Store arr[1]
			{"MOVIMM16L #1, r1", 36, []string{}},                  // arr[0] = 1 (inlined constant)
			{"ST r1, r2", 37, []string{}},                         // Store arr[0]
			{"MOVIMM16L #12, r1", 38, []string{}},                 // Loop counter offset
			{"MOVIMM16H #0, r1", 39, []string{}},                  // High bits
			{"ADD sp, r1, r1", 40, []string{}},                    // Loop counter address
			{"ST r0, r1", 41, []string{}},                         // Initialize loop counter to 0
			{"MOVIMM16L #8, r1", 42, []string{}},                  // Sum offset
			{"MOVIMM16H #0, r1", 43, []string{}},                  // High bits
			{"ADD sp, r1, r1", 44, []string{}},                    // Sum address
			{"ST r0, r1", 45, []string{}},                         // Initialize sum to 0
			{"MOVIMM16L .LBB0_1@lo, r4", 46, []string{".LBB0_1"}}, // Loop start label low
			{"MOVIMM16H .LBB0_1@hi, r4", 47, []string{".LBB0_1"}}, // Loop start label high
			{"JMP r4, r5", 48, []string{}},                        // Jump to loop
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
		// Main function should cover most of the instructions (excluding directives)
		assert.Greater(t, totalCovered, 100, "main function should cover > 100 instructions")
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
