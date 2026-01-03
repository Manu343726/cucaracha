package llvm

import (
	"testing"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
	"github.com/stretchr/testify/assert"
)

// TestDecodeSLEB128 tests signed LEB128 decoding
func TestDecodeSLEB128(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected int32
	}{
		{
			name:     "zero",
			input:    []byte{0x00},
			expected: 0,
		},
		{
			name:     "positive single byte",
			input:    []byte{0x08},
			expected: 8,
		},
		{
			name:     "positive max single byte",
			input:    []byte{0x3F},
			expected: 63,
		},
		{
			name:     "negative single byte (-1)",
			input:    []byte{0x7F},
			expected: -1,
		},
		{
			name:     "negative single byte (-64)",
			input:    []byte{0x40},
			expected: -64,
		},
		{
			name:     "positive two bytes (128)",
			input:    []byte{0x80, 0x01},
			expected: 128,
		},
		{
			name:     "positive two bytes (624)",
			input:    []byte{0xF0, 0x04},
			expected: 624,
		},
		{
			name:     "negative two bytes (-128)",
			input:    []byte{0x80, 0x7F},
			expected: -128,
		},
		{
			name:     "large positive value",
			input:    []byte{0xE5, 0x8E, 0x26},
			expected: 624485,
		},
		{
			name:     "large negative value",
			input:    []byte{0x9B, 0xF1, 0x59},
			expected: -624485,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decodeSLEB128(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDecodeULEB128 tests unsigned LEB128 decoding
func TestDecodeULEB128(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected uint32
	}{
		{
			name:     "zero",
			input:    []byte{0x00},
			expected: 0,
		},
		{
			name:     "single byte (1)",
			input:    []byte{0x01},
			expected: 1,
		},
		{
			name:     "single byte max (127)",
			input:    []byte{0x7F},
			expected: 127,
		},
		{
			name:     "two bytes (128)",
			input:    []byte{0x80, 0x01},
			expected: 128,
		},
		{
			name:     "two bytes (255)",
			input:    []byte{0xFF, 0x01},
			expected: 255,
		},
		{
			name:     "three bytes (16383)",
			input:    []byte{0xFF, 0x7F},
			expected: 16383,
		},
		{
			name:     "three bytes (16384)",
			input:    []byte{0x80, 0x80, 0x01},
			expected: 16384,
		},
		{
			name:     "value 624485",
			input:    []byte{0xE5, 0x8E, 0x26},
			expected: 624485,
		},
		{
			name:     "offset 28 (0x1c)",
			input:    []byte{0x1C},
			expected: 28,
		},
		{
			name:     "offset 24 (0x18)",
			input:    []byte{0x18},
			expected: 24,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decodeULEB128(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDecodeLocationExpr tests DWARF location expression decoding
func TestDecodeLocationExpr(t *testing.T) {
	// Create a minimal parser just for location decoding tests
	parser := &DWARFParser{}

	tests := []struct {
		name        string
		expr        []byte
		expectedLoc mc.VariableLocation
		expectedNil bool
	}{
		{
			name:        "empty expression",
			expr:        []byte{},
			expectedNil: true,
		},
		// DW_OP_reg tests (0x50-0x6f)
		{
			name:        "DW_OP_reg0 (r0)",
			expr:        []byte{0x50},
			expectedLoc: mc.RegisterLocation{Register: 16}, // r0 maps to 16
		},
		{
			name:        "DW_OP_reg1 (r1)",
			expr:        []byte{0x51},
			expectedLoc: mc.RegisterLocation{Register: 17}, // r1 maps to 17
		},
		{
			name:        "DW_OP_reg9 (r9)",
			expr:        []byte{0x59},
			expectedLoc: mc.RegisterLocation{Register: 25}, // r9 maps to 25
		},
		{
			name:        "DW_OP_reg13 (sp)",
			expr:        []byte{0x5D},
			expectedLoc: mc.RegisterLocation{Register: 13}, // sp stays 13
		},
		{
			name:        "DW_OP_reg14 (lr)",
			expr:        []byte{0x5E},
			expectedLoc: mc.RegisterLocation{Register: 14}, // lr stays 14
		},
		{
			name:        "DW_OP_reg15 (pc)",
			expr:        []byte{0x5F},
			expectedLoc: mc.RegisterLocation{Register: 15}, // pc stays 15
		},
		// DW_OP_breg tests (0x70-0x8f) - base register + offset
		{
			name: "DW_OP_breg0 with positive offset",
			expr: []byte{0x70, 0x08}, // [r0 + 8]
			expectedLoc: mc.MemoryLocation{
				BaseRegister: 16, // r0 maps to 16
				Offset:       8,
			},
		},
		{
			name: "DW_OP_breg13 (sp) with positive offset",
			expr: []byte{0x7D, 0x10}, // [sp + 16]
			expectedLoc: mc.MemoryLocation{
				BaseRegister: 13, // sp stays 13
				Offset:       16,
			},
		},
		{
			name: "DW_OP_breg13 (sp) with negative offset",
			expr: []byte{0x7D, 0x7C}, // [sp - 4]
			expectedLoc: mc.MemoryLocation{
				BaseRegister: 13, // sp stays 13
				Offset:       -4,
			},
		},
		{
			name: "DW_OP_breg13 with zero offset",
			expr: []byte{0x7D, 0x00}, // [sp + 0]
			expectedLoc: mc.MemoryLocation{
				BaseRegister: 13,
				Offset:       0,
			},
		},
		{
			name: "DW_OP_breg with large positive offset",
			expr: []byte{0x7D, 0x80, 0x01}, // [sp + 128]
			expectedLoc: mc.MemoryLocation{
				BaseRegister: 13,
				Offset:       128,
			},
		},
		// DW_OP_fbreg tests (0x91) - frame base relative
		{
			name: "DW_OP_fbreg with positive offset",
			expr: []byte{0x91, 0x10}, // [fp + 16]
			expectedLoc: mc.MemoryLocation{
				BaseRegister: 13, // Uses SP as frame base
				Offset:       16,
			},
		},
		{
			name: "DW_OP_fbreg with negative offset",
			expr: []byte{0x91, 0x7C}, // [fp - 4]
			expectedLoc: mc.MemoryLocation{
				BaseRegister: 13,
				Offset:       -4,
			},
		},
		{
			name: "DW_OP_fbreg with zero offset",
			expr: []byte{0x91, 0x00}, // [fp + 0]
			expectedLoc: mc.MemoryLocation{
				BaseRegister: 13,
				Offset:       0,
			},
		},
		// DW_OP_plus_uconst tests (0x23) - add unsigned constant
		{
			name: "DW_OP_plus_uconst offset 28",
			expr: []byte{0x23, 0x1C}, // [sp + 28]
			expectedLoc: mc.MemoryLocation{
				BaseRegister: 13, // SP
				Offset:       28,
			},
		},
		{
			name: "DW_OP_plus_uconst offset 24",
			expr: []byte{0x23, 0x18}, // [sp + 24]
			expectedLoc: mc.MemoryLocation{
				BaseRegister: 13,
				Offset:       24,
			},
		},
		{
			name: "DW_OP_plus_uconst offset 0",
			expr: []byte{0x23, 0x00}, // [sp + 0]
			expectedLoc: mc.MemoryLocation{
				BaseRegister: 13,
				Offset:       0,
			},
		},
		{
			name: "DW_OP_plus_uconst large offset",
			expr: []byte{0x23, 0x80, 0x02}, // [sp + 256]
			expectedLoc: mc.MemoryLocation{
				BaseRegister: 13,
				Offset:       256,
			},
		},
		// Unknown/unsupported opcodes
		{
			name:        "unsupported opcode DW_OP_addr",
			expr:        []byte{0x03, 0x00, 0x00, 0x00, 0x00}, // DW_OP_addr
			expectedNil: true,
		},
		{
			name:        "unsupported opcode DW_OP_stack_val",
			expr:        []byte{0x9F},
			expectedNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.decodeLocationExpr(tt.expr)

			if tt.expectedNil {
				assert.Nil(t, result, "expected nil location")
			} else {
				assert.NotNil(t, result, "expected non-nil location")
				assert.Equal(t, tt.expectedLoc, result)
			}
		})
	}
}

// TestMapDWARFRegister tests the DWARF to Cucaracha register mapping
func TestMapDWARFRegister(t *testing.T) {
	parser := &DWARFParser{}

	tests := []struct {
		name           string
		dwarfReg       uint32
		expectedCucReg uint32
	}{
		// General purpose registers r0-r9 map to internal indices 16-25
		{"r0", 0, 16},
		{"r1", 1, 17},
		{"r2", 2, 18},
		{"r3", 3, 19},
		{"r4", 4, 20},
		{"r5", 5, 21},
		{"r6", 6, 22},
		{"r7", 7, 23},
		{"r8", 8, 24},
		{"r9", 9, 25},
		// Special registers stay as-is
		{"sp (r13)", 13, 13},
		{"lr (r14)", 14, 14},
		{"pc (r15)", 15, 15},
		// Registers outside r0-r9 and special regs pass through
		{"r10", 10, 10},
		{"r11", 11, 11},
		{"r12", 12, 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.mapDWARFRegister(tt.dwarfReg)
			assert.Equal(t, tt.expectedCucReg, result)
		})
	}
}

// TestLocationExprWithRegisterMapping ensures register mapping is applied
// correctly when decoding location expressions
func TestLocationExprWithRegisterMapping(t *testing.T) {
	parser := &DWARFParser{}

	tests := []struct {
		name    string
		expr    []byte
		checkFn func(t *testing.T, loc mc.VariableLocation)
	}{
		{
			name: "DW_OP_reg0 should map to internal register 16",
			expr: []byte{0x50},
			checkFn: func(t *testing.T, loc mc.VariableLocation) {
				regLoc, ok := loc.(mc.RegisterLocation)
				assert.True(t, ok, "expected RegisterLocation")
				assert.Equal(t, uint32(16), regLoc.Register, "r0 should map to 16")
			},
		},
		{
			name: "DW_OP_breg0 should map base register to 16",
			expr: []byte{0x70, 0x04},
			checkFn: func(t *testing.T, loc mc.VariableLocation) {
				memLoc, ok := loc.(mc.MemoryLocation)
				assert.True(t, ok, "expected MemoryLocation")
				assert.Equal(t, uint32(16), memLoc.BaseRegister, "r0 should map to 16")
				assert.Equal(t, int32(4), memLoc.Offset)
			},
		},
		{
			name: "DW_OP_breg13 should keep SP as 13",
			expr: []byte{0x7D, 0x10},
			checkFn: func(t *testing.T, loc mc.VariableLocation) {
				memLoc, ok := loc.(mc.MemoryLocation)
				assert.True(t, ok, "expected MemoryLocation")
				assert.Equal(t, uint32(13), memLoc.BaseRegister, "sp should stay 13")
				assert.Equal(t, int32(16), memLoc.Offset)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := parser.decodeLocationExpr(tt.expr)
			assert.NotNil(t, loc)
			tt.checkFn(t, loc)
		})
	}
}

// TestNewDebugInfo tests that NewDebugInfo creates properly initialized maps
func TestNewDebugInfo(t *testing.T) {
	info := mc.NewDebugInfo()

	assert.NotNil(t, info)
	assert.NotNil(t, info.InstructionLocations, "InstructionLocations should be initialized")
	assert.NotNil(t, info.Functions, "Functions should be initialized")
	assert.NotNil(t, info.InstructionVariables, "InstructionVariables should be initialized")

	// Should be empty
	assert.Empty(t, info.InstructionLocations)
	assert.Empty(t, info.Functions)
	assert.Empty(t, info.InstructionVariables)
}

// TestVariableLocationTypes tests that all VariableLocation types implement the interface
func TestVariableLocationTypes(t *testing.T) {
	// This test ensures the interface is implemented correctly
	var loc mc.VariableLocation

	// RegisterLocation
	loc = mc.RegisterLocation{Register: 16}
	assert.NotNil(t, loc)
	regLoc, ok := loc.(mc.RegisterLocation)
	assert.True(t, ok)
	assert.Equal(t, uint32(16), regLoc.Register)

	// MemoryLocation
	loc = mc.MemoryLocation{BaseRegister: 13, Offset: 24}
	assert.NotNil(t, loc)
	memLoc, ok := loc.(mc.MemoryLocation)
	assert.True(t, ok)
	assert.Equal(t, uint32(13), memLoc.BaseRegister)
	assert.Equal(t, int32(24), memLoc.Offset)

	// ConstantLocation
	loc = mc.ConstantLocation{Value: 42}
	assert.NotNil(t, loc)
	constLoc, ok := loc.(mc.ConstantLocation)
	assert.True(t, ok)
	assert.Equal(t, int64(42), constLoc.Value)
}

// TestLEB128EdgeCases tests edge cases in LEB128 encoding/decoding
func TestLEB128EdgeCases(t *testing.T) {
	t.Run("SLEB128 with trailing bytes", func(t *testing.T) {
		// Should only read until continuation bit is clear
		data := []byte{0x08, 0xFF, 0xFF} // 8 followed by garbage
		result := decodeSLEB128(data)
		assert.Equal(t, int32(8), result)
	})

	t.Run("ULEB128 with trailing bytes", func(t *testing.T) {
		// Should only read until continuation bit is clear
		data := []byte{0x7F, 0xFF, 0xFF} // 127 followed by garbage
		result := decodeULEB128(data)
		assert.Equal(t, uint32(127), result)
	})

	t.Run("SLEB128 sign extension boundary", func(t *testing.T) {
		// -65 in signed LEB128 (0x3F with sign bit)
		data := []byte{0xBF, 0x7F}
		result := decodeSLEB128(data)
		assert.Equal(t, int32(-65), result)
	})
}

// TestDecodeLocationExprEdgeCases tests edge cases in location expression decoding
func TestDecodeLocationExprEdgeCases(t *testing.T) {
	parser := &DWARFParser{}

	t.Run("DW_OP_breg without offset byte", func(t *testing.T) {
		// Only the opcode, no offset - should use 0
		expr := []byte{0x7D} // DW_OP_breg13 (sp)
		loc := parser.decodeLocationExpr(expr)
		assert.NotNil(t, loc)
		memLoc, ok := loc.(mc.MemoryLocation)
		assert.True(t, ok)
		assert.Equal(t, int32(0), memLoc.Offset)
	})

	t.Run("DW_OP_fbreg without offset byte returns nil", func(t *testing.T) {
		// DW_OP_fbreg requires at least one offset byte
		expr := []byte{0x91} // DW_OP_fbreg alone
		loc := parser.decodeLocationExpr(expr)
		assert.Nil(t, loc)
	})

	t.Run("DW_OP_plus_uconst without offset byte returns nil", func(t *testing.T) {
		// DW_OP_plus_uconst requires at least one offset byte
		expr := []byte{0x23} // DW_OP_plus_uconst alone
		loc := parser.decodeLocationExpr(expr)
		assert.Nil(t, loc)
	})
}

// TestSourceLocationString tests the String() method of SourceLocation
func TestSourceLocationString(t *testing.T) {
	tests := []struct {
		name     string
		loc      mc.SourceLocation
		expected string
	}{
		{
			name: "with column",
			loc: mc.SourceLocation{
				File:   "test.c",
				Line:   42,
				Column: 10,
			},
			expected: "test.c:42:10",
		},
		{
			name: "without column",
			loc: mc.SourceLocation{
				File:   "main.c",
				Line:   100,
				Column: 0,
			},
			expected: "main.c:100",
		},
		{
			name: "line 1 column 1",
			loc: mc.SourceLocation{
				File:   "source.cpp",
				Line:   1,
				Column: 1,
			},
			expected: "source.cpp:1:1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.loc.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestVariableInfoFields tests VariableInfo structure
func TestVariableInfoFields(t *testing.T) {
	v := mc.VariableInfo{
		Name:        "myVar",
		TypeName:    "int",
		Size:        4,
		Location:    mc.MemoryLocation{BaseRegister: 13, Offset: 24},
		IsParameter: false,
	}

	assert.Equal(t, "myVar", v.Name)
	assert.Equal(t, "int", v.TypeName)
	assert.Equal(t, 4, v.Size)
	assert.False(t, v.IsParameter)

	memLoc, ok := v.Location.(mc.MemoryLocation)
	assert.True(t, ok)
	assert.Equal(t, uint32(13), memLoc.BaseRegister)
	assert.Equal(t, int32(24), memLoc.Offset)
}

// TestFunctionDebugInfoFields tests FunctionDebugInfo structure
func TestFunctionDebugInfoFields(t *testing.T) {
	f := mc.FunctionDebugInfo{
		Name:         "main",
		StartAddress: 0x10000,
		EndAddress:   0x10100,
		SourceFile:   "main.c",
		StartLine:    5,
		Parameters: []mc.VariableInfo{
			{Name: "argc", TypeName: "int", Size: 4},
			{Name: "argv", TypeName: "char**", Size: 4},
		},
		LocalVariables: []mc.VariableInfo{
			{Name: "x", TypeName: "int", Size: 4},
		},
	}

	assert.Equal(t, "main", f.Name)
	assert.Equal(t, uint32(0x10000), f.StartAddress)
	assert.Equal(t, uint32(0x10100), f.EndAddress)
	assert.Len(t, f.Parameters, 2)
	assert.Len(t, f.LocalVariables, 1)
}

// TestSourceLocationPropagation tests that DWARF line info is propagated to all
// instruction addresses between entries. DWARF only records line info at statement
// boundaries, so the parser must fill in intermediate instruction addresses.
func TestSourceLocationPropagation(t *testing.T) {
	// Simulate the result of parsing DWARF line entries at specific addresses
	// and verify that all instruction addresses (every 4 bytes) are filled in

	debugInfo := mc.NewDebugInfo()

	// Simulate what parseLineInfo does: given line entries at certain addresses,
	// it should fill in all 4-byte aligned addresses between them

	// Entry 1: Address 0x100, file "test.c", line 10
	// Entry 2: Address 0x110, file "test.c", line 15
	// Entry 3: Address 0x120, file "test.c", line 20

	// After propagation, addresses 0x100, 0x104, 0x108, 0x10C should all have line 10
	// Addresses 0x110, 0x114, 0x118, 0x11C should all have line 15
	// Address 0x120 should have line 20

	type lineEntryData struct {
		addr   uint32
		file   string
		line   int
		column int
	}
	entries := []lineEntryData{
		{addr: 0x100, file: "test.c", line: 10, column: 1},
		{addr: 0x110, file: "test.c", line: 15, column: 5},
		{addr: 0x120, file: "test.c", line: 20, column: 1},
	}

	// Simulate the propagation logic from parseLineInfo
	const instrSize = 4
	for i, entry := range entries {
		loc := &mc.SourceLocation{
			File:   entry.file,
			Line:   entry.line,
			Column: entry.column,
		}

		var endAddr uint32
		if i+1 < len(entries) {
			endAddr = entries[i+1].addr
		} else {
			endAddr = entry.addr + instrSize
		}

		for addr := entry.addr; addr < endAddr; addr += instrSize {
			debugInfo.InstructionLocations[addr] = loc
		}
	}

	// Verify propagation for first statement (line 10)
	// Should cover addresses 0x100, 0x104, 0x108, 0x10C
	for addr := uint32(0x100); addr < 0x110; addr += 4 {
		loc := debugInfo.GetSourceLocation(addr)
		assert.NotNil(t, loc, "Expected source location at address 0x%X", addr)
		assert.Equal(t, "test.c", loc.File)
		assert.Equal(t, 10, loc.Line, "Address 0x%X should have line 10", addr)
	}

	// Verify propagation for second statement (line 15)
	// Should cover addresses 0x110, 0x114, 0x118, 0x11C
	for addr := uint32(0x110); addr < 0x120; addr += 4 {
		loc := debugInfo.GetSourceLocation(addr)
		assert.NotNil(t, loc, "Expected source location at address 0x%X", addr)
		assert.Equal(t, "test.c", loc.File)
		assert.Equal(t, 15, loc.Line, "Address 0x%X should have line 15", addr)
	}

	// Verify last entry (line 20)
	loc := debugInfo.GetSourceLocation(0x120)
	assert.NotNil(t, loc)
	assert.Equal(t, 20, loc.Line)

	// Verify address before first entry has no location
	assert.Nil(t, debugInfo.GetSourceLocation(0x0FC))

	// Verify address after last entry has no location (unless it's within instrSize)
	assert.Nil(t, debugInfo.GetSourceLocation(0x124))
}

// TestSourceLocationPropagationMultipleFiles tests propagation across different source files
func TestSourceLocationPropagationMultipleFiles(t *testing.T) {
	debugInfo := mc.NewDebugInfo()

	type lineEntryData struct {
		addr   uint32
		file   string
		line   int
		column int
	}
	entries := []lineEntryData{
		{addr: 0x100, file: "main.c", line: 5, column: 1},
		{addr: 0x108, file: "utils.c", line: 10, column: 1}, // Different file
		{addr: 0x114, file: "main.c", line: 6, column: 1},   // Back to main.c
	}

	// Simulate propagation
	const instrSize = 4
	for i, entry := range entries {
		loc := &mc.SourceLocation{
			File:   entry.file,
			Line:   entry.line,
			Column: entry.column,
		}

		var endAddr uint32
		if i+1 < len(entries) {
			endAddr = entries[i+1].addr
		} else {
			endAddr = entry.addr + instrSize
		}

		for addr := entry.addr; addr < endAddr; addr += instrSize {
			debugInfo.InstructionLocations[addr] = loc
		}
	}

	// Verify first file (main.c, line 5) at 0x100 and 0x104
	loc1 := debugInfo.GetSourceLocation(0x100)
	assert.NotNil(t, loc1)
	assert.Equal(t, "main.c", loc1.File)
	assert.Equal(t, 5, loc1.Line)

	loc2 := debugInfo.GetSourceLocation(0x104)
	assert.NotNil(t, loc2)
	assert.Equal(t, "main.c", loc2.File)
	assert.Equal(t, 5, loc2.Line)

	// Verify second file (utils.c, line 10) at 0x108, 0x10C, 0x110
	loc3 := debugInfo.GetSourceLocation(0x108)
	assert.NotNil(t, loc3)
	assert.Equal(t, "utils.c", loc3.File)
	assert.Equal(t, 10, loc3.Line)

	loc4 := debugInfo.GetSourceLocation(0x110)
	assert.NotNil(t, loc4)
	assert.Equal(t, "utils.c", loc4.File)
	assert.Equal(t, 10, loc4.Line)

	// Verify back to main.c at 0x114
	loc5 := debugInfo.GetSourceLocation(0x114)
	assert.NotNil(t, loc5)
	assert.Equal(t, "main.c", loc5.File)
	assert.Equal(t, 6, loc5.Line)
}

// TestSourceLocationPropagationSingleInstruction tests a statement with only one instruction
func TestSourceLocationPropagationSingleInstruction(t *testing.T) {
	debugInfo := mc.NewDebugInfo()

	// Two consecutive DWARF entries 4 bytes apart (single instruction per statement)
	type lineEntryData struct {
		addr   uint32
		file   string
		line   int
		column int
	}
	entries := []lineEntryData{
		{addr: 0x200, file: "test.c", line: 1, column: 1},
		{addr: 0x204, file: "test.c", line: 2, column: 1},
		{addr: 0x208, file: "test.c", line: 3, column: 1},
	}

	// Simulate propagation
	const instrSize = 4
	for i, entry := range entries {
		loc := &mc.SourceLocation{
			File:   entry.file,
			Line:   entry.line,
			Column: entry.column,
		}

		var endAddr uint32
		if i+1 < len(entries) {
			endAddr = entries[i+1].addr
		} else {
			endAddr = entry.addr + instrSize
		}

		for addr := entry.addr; addr < endAddr; addr += instrSize {
			debugInfo.InstructionLocations[addr] = loc
		}
	}

	// Each address should have exactly its own line
	loc1 := debugInfo.GetSourceLocation(0x200)
	assert.NotNil(t, loc1)
	assert.Equal(t, 1, loc1.Line)

	loc2 := debugInfo.GetSourceLocation(0x204)
	assert.NotNil(t, loc2)
	assert.Equal(t, 2, loc2.Line)

	loc3 := debugInfo.GetSourceLocation(0x208)
	assert.NotNil(t, loc3)
	assert.Equal(t, 3, loc3.Line)
}

// TestSourceLocationPropagationLargeGap tests propagation across a large address gap
func TestSourceLocationPropagationLargeGap(t *testing.T) {
	debugInfo := mc.NewDebugInfo()

	// A statement that spans many instructions (e.g., complex expression)
	type lineEntryData struct {
		addr   uint32
		file   string
		line   int
		column int
	}
	entries := []lineEntryData{
		{addr: 0x300, file: "test.c", line: 50, column: 1},
		{addr: 0x340, file: "test.c", line: 51, column: 1}, // 16 instructions (64 bytes) for line 50
	}

	// Simulate propagation
	const instrSize = 4
	for i, entry := range entries {
		loc := &mc.SourceLocation{
			File:   entry.file,
			Line:   entry.line,
			Column: entry.column,
		}

		var endAddr uint32
		if i+1 < len(entries) {
			endAddr = entries[i+1].addr
		} else {
			endAddr = entry.addr + instrSize
		}

		for addr := entry.addr; addr < endAddr; addr += instrSize {
			debugInfo.InstructionLocations[addr] = loc
		}
	}

	// All 16 instructions from 0x300 to 0x33C should have line 50
	for addr := uint32(0x300); addr < 0x340; addr += 4 {
		loc := debugInfo.GetSourceLocation(addr)
		assert.NotNil(t, loc, "Expected source location at address 0x%X", addr)
		assert.Equal(t, 50, loc.Line, "Address 0x%X should have line 50", addr)
	}

	// Address 0x340 should have line 51
	loc := debugInfo.GetSourceLocation(0x340)
	assert.NotNil(t, loc)
	assert.Equal(t, 51, loc.Line)

	// Count total entries for line 50 (should be 16)
	count := 0
	for addr := uint32(0x300); addr < 0x340; addr += 4 {
		if debugInfo.GetSourceLocation(addr) != nil {
			count++
		}
	}
	assert.Equal(t, 16, count, "Expected 16 instruction addresses for line 50")
}
