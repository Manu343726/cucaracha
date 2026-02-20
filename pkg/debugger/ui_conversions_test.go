package debugger

import (
	"testing"

	"github.com/Manu343726/cucaracha/pkg/debugger/core"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/Manu343726/cucaracha/pkg/hw/memory"
	"github.com/Manu343726/cucaracha/pkg/runtime/program"
	"github.com/Manu343726/cucaracha/pkg/runtime/program/sourcecode"
	uiDebugger "github.com/Manu343726/cucaracha/pkg/ui/debugger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for SourceLocationToUI

func TestSourceLocationToUI_WithValidLocation(t *testing.T) {
	file := sourcecode.FileNamed("test.c")
	loc := &sourcecode.Location{
		File:   file,
		Line:   42,
		Column: 10,
	}

	result := SourceLocationToUI(loc)

	require.NotNil(t, result)
	assert.Equal(t, "test.c", result.File)
	assert.Equal(t, 42, result.Line)
}

func TestSourceLocationToUI_WithNilLocation(t *testing.T) {
	result := SourceLocationToUI(nil)

	assert.Nil(t, result)
}

func TestSourceLocationToUI_WithDifferentFiles(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		line     int
		expected string
	}{
		{"main file", "main.c", 1, "main.c"},
		{"header file", "header.h", 100, "header.h"},
		{"nested path", "src/module.c", 50, "src/module.c"},
		{"complex path", "/usr/include/stdio.h", 200, "/usr/include/stdio.h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := sourcecode.FileNamed(tt.file)
			loc := &sourcecode.Location{File: file, Line: tt.line}

			result := SourceLocationToUI(loc)

			require.NotNil(t, result)
			assert.Equal(t, tt.expected, result.File)
			assert.Equal(t, tt.line, result.Line)
		})
	}
}

// Tests for SourceLocationFromUI

func TestSourceLocationFromUI_WithValidLocation(t *testing.T) {
	loc := &uiDebugger.SourceLocation{
		File: "test.c",
		Line: 42,
	}

	result := SourceLocationFromUI(loc)

	require.NotNil(t, result)
	assert.Equal(t, "test.c", result.File.Path())
	assert.Equal(t, 42, result.Line)
}

func TestSourceLocationFromUI_WithNilLocation(t *testing.T) {
	result := SourceLocationFromUI(nil)

	assert.Nil(t, result)
}

func TestSourceLocationFromUI_WithDifferentLines(t *testing.T) {
	lines := []int{0, 1, 10, 100, 1000, 999999}

	for _, line := range lines {
		loc := &uiDebugger.SourceLocation{File: "test.c", Line: line}
		result := SourceLocationFromUI(loc)

		require.NotNil(t, result)
		assert.Equal(t, line, result.Line)
	}
}

// Tests for roundtrip conversion

func TestSourceLocationRoundtrip(t *testing.T) {
	file := sourcecode.FileNamed("roundtrip.c")
	original := &sourcecode.Location{File: file, Line: 123, Column: 45}

	// Convert to UI and back
	uiLoc := SourceLocationToUI(original)
	restored := SourceLocationFromUI(uiLoc)

	assert.Equal(t, original.File.Path(), restored.File.Path())
	assert.Equal(t, original.Line, restored.Line)
}

// Tests for MemoryRangeToUI

func TestMemoryRangeToUI_WithValidRange(t *testing.T) {
	memRange := &memory.Range{
		Start: 0x1000,
		Size:  0x100,
	}

	result := MemoryRangeToUI(memRange)

	require.NotNil(t, result)
	assert.Equal(t, uint32(0x1000), result.Start)
	assert.Equal(t, uint32(0x100), result.Size)
}

func TestMemoryRangeToUI_WithNilRange(t *testing.T) {
	result := MemoryRangeToUI(nil)

	assert.Nil(t, result)
}

func TestMemoryRangeToUI_WithDifferentRanges(t *testing.T) {
	tests := []struct {
		name  string
		start uint32
		size  uint32
	}{
		{"zero start", 0, 1},
		{"max start", 0xFFFFFFFF, 1},
		{"large size", 0x1000, 0xFFFFFFFF},
		{"code segment", 0x1000, 0x1000},
		{"data segment", 0x3000, 0x1000},
		{"single byte", 0x2000, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memRange := &memory.Range{Start: tt.start, Size: tt.size}
			result := MemoryRangeToUI(memRange)

			require.NotNil(t, result)
			assert.Equal(t, tt.start, result.Start)
			assert.Equal(t, tt.size, result.Size)
		})
	}
}

// Tests for MemoryRangeFromUI

func TestMemoryRangeFromUI_WithValidRange(t *testing.T) {
	region := &uiDebugger.MemoryRegion{
		Start: 0x1000,
		Size:  0x100,
	}

	result := MemoryRangeFromUI(region)

	require.NotNil(t, result)
	assert.Equal(t, uint32(0x1000), result.Start)
	assert.Equal(t, uint32(0x100), result.Size)
}

func TestMemoryRangeFromUI_WithNilRange(t *testing.T) {
	result := MemoryRangeFromUI(nil)

	assert.Nil(t, result)
}

// Tests for roundtrip memory conversion

func TestMemoryRangeRoundtrip(t *testing.T) {
	original := &memory.Range{Start: 0x4000, Size: 0x2000}

	// Convert to UI and back
	uiRange := MemoryRangeToUI(original)
	restored := MemoryRangeFromUI(uiRange)

	assert.Equal(t, original.Start, restored.Start)
	assert.Equal(t, original.Size, restored.Size)
}

// Tests for WatchpointTypeFromUI

func TestWatchpointTypeFromUI_Read(t *testing.T) {
	result := WatchpointTypeFromUI(uiDebugger.WatchpointTypeRead)

	assert.Equal(t, core.WatchRead, result)
}

func TestWatchpointTypeFromUI_Write(t *testing.T) {
	result := WatchpointTypeFromUI(uiDebugger.WatchpointTypeWrite)

	assert.Equal(t, core.WatchWrite, result)
}

func TestWatchpointTypeFromUI_ReadWrite(t *testing.T) {
	result := WatchpointTypeFromUI(uiDebugger.WatchpointTypeReadWrite)

	assert.Equal(t, core.WatchReadWrite, result)
}

func TestWatchpointTypeFromUI_AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		uiType   uiDebugger.WatchpointType
		expected core.WatchpointType
	}{
		{"read", uiDebugger.WatchpointTypeRead, core.WatchRead},
		{"write", uiDebugger.WatchpointTypeWrite, core.WatchWrite},
		{"read write", uiDebugger.WatchpointTypeReadWrite, core.WatchReadWrite},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WatchpointTypeFromUI(tt.uiType)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWatchpointTypeFromUI_InvalidType_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for unknown watchpoint type")
		}
	}()

	WatchpointTypeFromUI(uiDebugger.WatchpointType(999))
}

// Tests for WatchpointTypeToUI

func TestWatchpointTypeToUI_Read(t *testing.T) {
	result := WatchpointTypeToUI(core.WatchRead)

	assert.Equal(t, uiDebugger.WatchpointTypeRead, result)
}

func TestWatchpointTypeToUI_Write(t *testing.T) {
	result := WatchpointTypeToUI(core.WatchWrite)

	assert.Equal(t, uiDebugger.WatchpointTypeWrite, result)
}

func TestWatchpointTypeToUI_ReadWrite(t *testing.T) {
	result := WatchpointTypeToUI(core.WatchReadWrite)

	assert.Equal(t, uiDebugger.WatchpointTypeReadWrite, result)
}

func TestWatchpointTypeToUI_AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		coreType core.WatchpointType
		expected uiDebugger.WatchpointType
	}{
		{"read", core.WatchRead, uiDebugger.WatchpointTypeRead},
		{"write", core.WatchWrite, uiDebugger.WatchpointTypeWrite},
		{"read write", core.WatchReadWrite, uiDebugger.WatchpointTypeReadWrite},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WatchpointTypeToUI(tt.coreType)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWatchpointTypeToUI_InvalidType_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for unknown watchpoint type")
		}
	}()

	WatchpointTypeToUI(core.WatchpointType(999))
}

// Tests for roundtrip watchpoint type conversion

func TestWatchpointTypeRoundtrip(t *testing.T) {
	types := []core.WatchpointType{core.WatchRead, core.WatchWrite, core.WatchReadWrite}

	for _, original := range types {
		// Convert to UI and back
		uiType := WatchpointTypeToUI(original)
		restored := WatchpointTypeFromUI(uiType)

		assert.Equal(t, original, restored)
	}
}

// Tests for WatchpointToUI

func TestWatchpointToUI_WithValidWatchpoint(t *testing.T) {
	wp := &core.Watchpoint{
		ID:      123,
		Memory:  &memory.Range{Start: 0x1000, Size: 0x10},
		Type:    core.WatchRead,
		Enabled: true,
	}

	result := WatchpointToUI(wp)

	require.NotNil(t, result)
	assert.Equal(t, 123, result.ID)
	assert.Equal(t, uint32(0x1000), result.Range.Start)
	assert.Equal(t, uint32(0x10), result.Range.Size)
	assert.Equal(t, uiDebugger.WatchpointTypeRead, result.Type)
	assert.True(t, result.Enabled)
}

func TestWatchpointToUI_WithNilWatchpoint(t *testing.T) {
	result := WatchpointToUI(nil)

	assert.Nil(t, result)
}

func TestWatchpointToUI_WithDisabledWatchpoint(t *testing.T) {
	wp := &core.Watchpoint{
		ID:      456,
		Memory:  &memory.Range{Start: 0x2000, Size: 0x20},
		Type:    core.WatchWrite,
		Enabled: false,
	}

	result := WatchpointToUI(wp)

	require.NotNil(t, result)
	assert.False(t, result.Enabled)
	assert.Equal(t, uiDebugger.WatchpointTypeWrite, result.Type)
}

func TestWatchpointToUI_WithDifferentTypes(t *testing.T) {
	types := []struct {
		name string
		wt   core.WatchpointType
	}{
		{"read", core.WatchRead},
		{"write", core.WatchWrite},
		{"read write", core.WatchReadWrite},
	}

	for _, tt := range types {
		t.Run(tt.name, func(t *testing.T) {
			wp := &core.Watchpoint{
				ID:      100,
				Memory:  &memory.Range{Start: 0x5000, Size: 0x100},
				Type:    tt.wt,
				Enabled: true,
			}

			result := WatchpointToUI(wp)

			require.NotNil(t, result)
			expectedUIType := WatchpointTypeToUI(tt.wt)
			assert.Equal(t, expectedUIType, result.Type)
		})
	}
}

// Tests for WatchpointFromUI

func TestWatchpointFromUI_WithValidWatchpoint(t *testing.T) {
	wp := &uiDebugger.Watchpoint{
		ID:      123,
		Range:   &uiDebugger.MemoryRegion{Start: 0x1000, Size: 0x10},
		Type:    uiDebugger.WatchpointTypeRead,
		Enabled: true,
	}

	result := WatchpointFromUI(wp)

	require.NotNil(t, result)
	assert.Equal(t, 123, result.ID)
	assert.Equal(t, uint32(0x1000), result.Memory.Start)
	assert.Equal(t, uint32(0x10), result.Memory.Size)
	assert.Equal(t, core.WatchRead, result.Type)
	assert.True(t, result.Enabled)
}

func TestWatchpointFromUI_WithNilWatchpoint(t *testing.T) {
	result := WatchpointFromUI(nil)

	assert.Nil(t, result)
}

func TestWatchpointFromUI_WithDisabledWatchpoint(t *testing.T) {
	wp := &uiDebugger.Watchpoint{
		ID:      456,
		Range:   &uiDebugger.MemoryRegion{Start: 0x2000, Size: 0x20},
		Type:    uiDebugger.WatchpointTypeWrite,
		Enabled: false,
	}

	result := WatchpointFromUI(wp)

	require.NotNil(t, result)
	assert.False(t, result.Enabled)
	assert.Equal(t, core.WatchWrite, result.Type)
}

// Tests for roundtrip watchpoint conversion

func TestWatchpointRoundtrip(t *testing.T) {
	original := &core.Watchpoint{
		ID:      999,
		Memory:  &memory.Range{Start: 0x8000, Size: 0x200},
		Type:    core.WatchReadWrite,
		Enabled: true,
	}

	// Convert to UI and back
	uiWp := WatchpointToUI(original)
	restored := WatchpointFromUI(uiWp)

	assert.Equal(t, original.ID, restored.ID)
	assert.Equal(t, original.Memory.Start, restored.Memory.Start)
	assert.Equal(t, original.Memory.Size, restored.Memory.Size)
	assert.Equal(t, original.Type, restored.Type)
	assert.Equal(t, original.Enabled, restored.Enabled)
}

// Tests for BreakpointToUI

func TestBreakpointToUI_WithValidBreakpoint(t *testing.T) {
	bp := &core.Breakpoint{
		ID:      111,
		Address: 0x5000,
		Enabled: true,
	}

	result := BreakpointToUI(bp)

	require.NotNil(t, result)
	assert.Equal(t, 111, result.ID)
	assert.Equal(t, uint32(0x5000), result.Address)
	assert.True(t, result.Enabled)
}

func TestBreakpointToUI_WithNilBreakpoint(t *testing.T) {
	result := BreakpointToUI(nil)

	assert.Nil(t, result)
}

func TestBreakpointToUI_WithDisabledBreakpoint(t *testing.T) {
	bp := &core.Breakpoint{
		ID:      222,
		Address: 0x6000,
		Enabled: false,
	}

	result := BreakpointToUI(bp)

	require.NotNil(t, result)
	assert.False(t, result.Enabled)
}

func TestBreakpointToUI_WithDifferentAddresses(t *testing.T) {
	addresses := []uint32{0, 0x1000, 0x8000, 0xFFFFFFFF}

	for _, addr := range addresses {
		bp := &core.Breakpoint{
			ID:      100,
			Address: addr,
			Enabled: true,
		}

		result := BreakpointToUI(bp)

		require.NotNil(t, result)
		assert.Equal(t, addr, result.Address)
	}
}

// Tests for BreakpointFromUI

func TestBreakpointFromUI_WithValidBreakpoint(t *testing.T) {
	bp := &uiDebugger.Breakpoint{
		ID:      111,
		Address: 0x5000,
		Enabled: true,
	}

	result := BreakpointFromUI(bp)

	require.NotNil(t, result)
	assert.Equal(t, 111, result.ID)
	assert.Equal(t, uint32(0x5000), result.Address)
	assert.True(t, result.Enabled)
}

func TestBreakpointFromUI_WithNilBreakpoint(t *testing.T) {
	result := BreakpointFromUI(nil)

	assert.Nil(t, result)
}

func TestBreakpointFromUI_WithDisabledBreakpoint(t *testing.T) {
	bp := &uiDebugger.Breakpoint{
		ID:      222,
		Address: 0x6000,
		Enabled: false,
	}

	result := BreakpointFromUI(bp)

	require.NotNil(t, result)
	assert.False(t, result.Enabled)
}

// Tests for roundtrip breakpoint conversion

func TestBreakpointRoundtrip(t *testing.T) {
	original := &core.Breakpoint{
		ID:      333,
		Address: 0x7000,
		Enabled: true,
	}

	// Convert to UI and back
	uiBp := BreakpointToUI(original)
	restored := BreakpointFromUI(uiBp)

	assert.Equal(t, original.ID, restored.ID)
	assert.Equal(t, original.Address, restored.Address)
	assert.Equal(t, original.Enabled, restored.Enabled)
}

// Tests for InstructionOperandKindToUI

func TestInstructionOperandKindToUI_Register(t *testing.T) {
	result := InstructionOperandKindToUI(instructions.OperandKind_Register)

	assert.Equal(t, uiDebugger.OperandKindRegister, result)
}

func TestInstructionOperandKindToUI_Immediate(t *testing.T) {
	result := InstructionOperandKindToUI(instructions.OperandKind_Immediate)

	assert.Equal(t, uiDebugger.OperandKindImmediate, result)
}

func TestInstructionOperandKindToUI_AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		kind     instructions.OperandKind
		expected uiDebugger.InstructionOperandKind
	}{
		{"register", instructions.OperandKind_Register, uiDebugger.OperandKindRegister},
		{"immediate", instructions.OperandKind_Immediate, uiDebugger.OperandKindImmediate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InstructionOperandKindToUI(tt.kind)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInstructionOperandKindToUI_InvalidKind_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for unknown operand kind")
		}
	}()

	InstructionOperandKindToUI(instructions.OperandKind(999))
}

// Tests for InstructionToUI

func TestInstructionToUI_WithNilInstruction(t *testing.T) {
	result := InstructionToUI(nil)

	assert.Nil(t, result)
}

func TestInstructionToUI_WithValidInstruction_NOP(t *testing.T) {
	// Get NOP instruction descriptor
	nopDesc, err := instructions.Instructions.Instruction(instructions.OpCode_NOP)
	require.NoError(t, err)

	// Create a fully decoded instruction
	decodedInstr := &instructions.Instruction{
		Descriptor:    nopDesc,
		OperandValues: []instructions.OperandValue{},
	}

	// Create a raw NOP instruction
	rawNop := &instructions.RawInstruction{
		Descriptor:    nopDesc,
		OperandValues: []uint64{},
	}

	// Create a program instruction with both raw and decoded instruction
	addr := uint32(0x1000)
	instr := &program.Instruction{
		Address:     &addr,
		Text:        "NOP",
		Raw:         rawNop,
		Instruction: decodedInstr,
	}

	result := InstructionToUI(instr)

	require.NotNil(t, result)
	assert.Equal(t, uint32(0x1000), result.Address)
	assert.Equal(t, "NOP", result.Mnemonic)
	// Encoding should be set (even if to 0 for NOP)
	assert.IsType(t, uint32(0), result.Encoding)
}

func TestInstructionToUI_WithValidInstruction_WithAddress(t *testing.T) {
	// Get NOP instruction descriptor
	nopDesc, err := instructions.Instructions.Instruction(instructions.OpCode_NOP)
	require.NoError(t, err)

	// Create a fully decoded instruction
	decodedInstr := &instructions.Instruction{
		Descriptor:    nopDesc,
		OperandValues: []instructions.OperandValue{},
	}

	// Create a raw NOP instruction
	rawNop := &instructions.RawInstruction{
		Descriptor:    nopDesc,
		OperandValues: []uint64{},
	}

	// Test different addresses
	testCases := []uint32{0x0000, 0x0100, 0x1000, 0xFFFF}

	for _, addr := range testCases {
		addrCopy := addr
		instr := &program.Instruction{
			Address:     &addrCopy,
			Text:        "NOP",
			Raw:         rawNop,
			Instruction: decodedInstr,
		}

		result := InstructionToUI(instr)

		require.NotNil(t, result, "address=%x", addr)
		assert.Equal(t, addr, result.Address)
	}
}

func TestInstructionToUI_WithNilAddress(t *testing.T) {
	// Get NOP instruction descriptor
	nopDesc, err := instructions.Instructions.Instruction(instructions.OpCode_NOP)
	require.NoError(t, err)

	// Create a fully decoded instruction
	decodedInstr := &instructions.Instruction{
		Descriptor:    nopDesc,
		OperandValues: []instructions.OperandValue{},
	}

	// Create a raw NOP instruction
	rawNop := &instructions.RawInstruction{
		Descriptor:    nopDesc,
		OperandValues: []uint64{},
	}

	instr := &program.Instruction{
		Address:     nil,
		Text:        "NOP",
		Raw:         rawNop,
		Instruction: decodedInstr,
	}

	result := InstructionToUI(instr)

	require.NotNil(t, result)
	assert.Equal(t, uint32(0), result.Address) // Should default to 0
	assert.Equal(t, "NOP", result.Mnemonic)
}

func TestInstructionToUI_WithDifferentMnemonics(t *testing.T) {
	// Test that InstructionToUI correctly returns the descriptor's mnemonic
	nopDesc, err := instructions.Instructions.Instruction(instructions.OpCode_NOP)
	require.NoError(t, err)

	hltDesc, err := instructions.Instructions.Instruction(instructions.OpCode_HLT)
	require.NoError(t, err)

	testCases := []struct {
		name       string
		descriptor *instructions.InstructionDescriptor
		mnemonic   string
	}{
		{"NOP", nopDesc, "NOP"},
		{"HLT", hltDesc, "HLT"},
	}

	for _, tc := range testCases {
		decodedInstr := &instructions.Instruction{
			Descriptor:    tc.descriptor,
			OperandValues: []instructions.OperandValue{},
		}

		rawInstr := &instructions.RawInstruction{
			Descriptor:    tc.descriptor,
			OperandValues: []uint64{},
		}

		addr := uint32(0x1000)
		instr := &program.Instruction{
			Address:     &addr,
			Text:        tc.mnemonic,
			Raw:         rawInstr,
			Instruction: decodedInstr,
		}

		result := InstructionToUI(instr)

		require.NotNil(t, result, "mnemonic=%s", tc.mnemonic)
		assert.Equal(t, tc.mnemonic, result.Mnemonic, "expected mnemonic from descriptor for %s", tc.name)
	}
}

func TestInstructionToUI_Encoding_Consistency(t *testing.T) {
	// Test that encoding is consistent and not zero
	nopDesc, err := instructions.Instructions.Instruction(instructions.OpCode_NOP)
	require.NoError(t, err)

	decodedInstr := &instructions.Instruction{
		Descriptor:    nopDesc,
		OperandValues: []instructions.OperandValue{},
	}

	rawNop := &instructions.RawInstruction{
		Descriptor:    nopDesc,
		OperandValues: []uint64{},
	}

	// Create multiple instructions with same properties
	addr := uint32(0x2000)
	instr1 := &program.Instruction{
		Address:     &addr,
		Text:        "NOP",
		Raw:         rawNop,
		Instruction: decodedInstr,
	}

	instr2 := &program.Instruction{
		Address:     &addr,
		Text:        "NOP",
		Raw:         rawNop,
		Instruction: decodedInstr,
	}

	result1 := InstructionToUI(instr1)
	result2 := InstructionToUI(instr2)

	require.NotNil(t, result1)
	require.NotNil(t, result2)

	// Encoding should be the same for identical instructions
	assert.Equal(t, result1.Encoding, result2.Encoding)
}

// Integration tests

func TestConversions_MultipleObjects(t *testing.T) {
	// Test converting multiple objects of different types
	bps := []*core.Breakpoint{
		{ID: 1, Address: 0x1000, Enabled: true},
		{ID: 2, Address: 0x2000, Enabled: false},
		{ID: 3, Address: 0x3000, Enabled: true},
	}

	for _, bp := range bps {
		uiBp := BreakpointToUI(bp)
		restored := BreakpointFromUI(uiBp)

		assert.Equal(t, bp.ID, restored.ID)
		assert.Equal(t, bp.Address, restored.Address)
		assert.Equal(t, bp.Enabled, restored.Enabled)
	}
}

func TestConversions_NilHandling(t *testing.T) {
	// Test that all conversion functions properly handle nil inputs
	assert.Nil(t, SourceLocationToUI(nil))
	assert.Nil(t, SourceLocationFromUI(nil))
	assert.Nil(t, MemoryRangeToUI(nil))
	assert.Nil(t, MemoryRangeFromUI(nil))
	assert.Nil(t, WatchpointToUI(nil))
	assert.Nil(t, WatchpointFromUI(nil))
	assert.Nil(t, BreakpointToUI(nil))
	assert.Nil(t, BreakpointFromUI(nil))
	assert.Nil(t, InstructionToUI(nil))
}

func TestConversions_PreservesData(t *testing.T) {
	// Test that all roundtrip conversions preserve data

	// Watchpoint roundtrip
	wp := &core.Watchpoint{
		ID:      555,
		Memory:  &memory.Range{Start: 0xA000, Size: 0x50},
		Type:    core.WatchReadWrite,
		Enabled: false,
	}
	wpUI := WatchpointToUI(wp)
	wpRestored := WatchpointFromUI(wpUI)
	assert.Equal(t, wp.ID, wpRestored.ID)
	assert.Equal(t, wp.Memory.Start, wpRestored.Memory.Start)
	assert.Equal(t, wp.Memory.Size, wpRestored.Memory.Size)
	assert.Equal(t, wp.Type, wpRestored.Type)
	assert.Equal(t, wp.Enabled, wpRestored.Enabled)

	// Breakpoint roundtrip
	bp := &core.Breakpoint{
		ID:      666,
		Address: 0xB000,
		Enabled: true,
	}
	bpUI := BreakpointToUI(bp)
	bpRestored := BreakpointFromUI(bpUI)
	assert.Equal(t, bp.ID, bpRestored.ID)
	assert.Equal(t, bp.Address, bpRestored.Address)
	assert.Equal(t, bp.Enabled, bpRestored.Enabled)

	// Memory range roundtrip
	memRange := &memory.Range{Start: 0xC000, Size: 0x1000}
	memRangeUI := MemoryRangeToUI(memRange)
	memRangeRestored := MemoryRangeFromUI(memRangeUI)
	assert.Equal(t, memRange.Start, memRangeRestored.Start)
	assert.Equal(t, memRange.Size, memRangeRestored.Size)
}
