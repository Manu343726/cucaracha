package program

import (
	"testing"

	"github.com/Manu343726/cucaracha/pkg/runtime/program/sourcecode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for VariableLocationType.String()

func TestVariableLocationType_String_Register(t *testing.T) {
	vlt := VariableLocationRegister
	assert.Equal(t, "Register", vlt.String())
}

func TestVariableLocationType_String_Memory(t *testing.T) {
	vlt := VariableLocationMemory
	assert.Equal(t, "Memory", vlt.String())
}

func TestVariableLocationType_String_Constant(t *testing.T) {
	vlt := VariableLocationConstant
	assert.Equal(t, "Constant", vlt.String())
}

func TestVariableLocationType_String_Unknown(t *testing.T) {
	vlt := VariableLocationType(999)
	assert.Equal(t, "Unknown", vlt.String())
}

// Tests for RegisterLocation.Type()

func TestRegisterLocation_Type(t *testing.T) {
	loc := RegisterLocation{Register: 5}
	assert.Equal(t, VariableLocationRegister, loc.Type())
}

func TestRegisterLocation_Type_RegisterZero(t *testing.T) {
	loc := RegisterLocation{Register: 0}
	assert.Equal(t, VariableLocationRegister, loc.Type())
}

// Tests for MemoryLocation.Type()

func TestMemoryLocation_Type(t *testing.T) {
	loc := MemoryLocation{BaseRegister: 14, Offset: -8}
	assert.Equal(t, VariableLocationMemory, loc.Type())
}

func TestMemoryLocation_Type_PositiveOffset(t *testing.T) {
	loc := MemoryLocation{BaseRegister: 15, Offset: 256}
	assert.Equal(t, VariableLocationMemory, loc.Type())
}

// Tests for ConstantLocation.Type()

func TestConstantLocation_Type(t *testing.T) {
	loc := ConstantLocation{Value: 42}
	assert.Equal(t, VariableLocationConstant, loc.Type())
}

func TestConstantLocation_Type_NegativeValue(t *testing.T) {
	loc := ConstantLocation{Value: -100}
	assert.Equal(t, VariableLocationConstant, loc.Type())
}

// Tests for NewDebugInfo

func TestNewDebugInfo_CreatesEmptyStructure(t *testing.T) {
	debugInfo := NewDebugInfo()

	require.NotNil(t, debugInfo)
	assert.NotNil(t, debugInfo.SourceLibrary)
	assert.NotNil(t, debugInfo.InstructionLocations)
	assert.NotNil(t, debugInfo.InstructionVariables)
	assert.NotNil(t, debugInfo.Functions)
	assert.Equal(t, 0, len(debugInfo.InstructionLocations))
	assert.Equal(t, 0, len(debugInfo.InstructionVariables))
	assert.Equal(t, 0, len(debugInfo.Functions))
}

func TestNewDebugInfo_IndependentInstances(t *testing.T) {
	debugInfo1 := NewDebugInfo()
	debugInfo2 := NewDebugInfo()

	assert.NotSame(t, debugInfo1, debugInfo2)

	// Add data to one and ensure it doesn't affect the other
	file := sourcecode.FileNamed("test.c")
	loc := &sourcecode.Location{File: file, Line: 10}
	debugInfo1.InstructionLocations[0x1000] = loc

	// Verify the other is still empty
	assert.Nil(t, debugInfo2.GetSourceLocation(0x1000))
}

// Tests for GetSourceLocation

func TestGetSourceLocation_NilDebugInfo(t *testing.T) {
	var debugInfo *DebugInfo
	loc := debugInfo.GetSourceLocation(0x1000)
	assert.Nil(t, loc)
}

func TestGetSourceLocation_AddressNotFound(t *testing.T) {
	debugInfo := NewDebugInfo()
	loc := debugInfo.GetSourceLocation(0x1000)
	assert.Nil(t, loc)
}

func TestGetSourceLocation_AddressFound(t *testing.T) {
	debugInfo := NewDebugInfo()
	file := sourcecode.FileNamed("test.c")
	expectedLoc := &sourcecode.Location{
		File:   file,
		Line:   10,
		Column: 5,
	}
	debugInfo.InstructionLocations[0x1000] = expectedLoc

	loc := debugInfo.GetSourceLocation(0x1000)
	require.NotNil(t, loc)
	assert.Equal(t, expectedLoc, loc)
}

func TestGetSourceLocation_MultipleAddresses(t *testing.T) {
	debugInfo := NewDebugInfo()
	file1 := sourcecode.FileNamed("test.c")
	file2 := sourcecode.FileNamed("other.c")
	loc1 := &sourcecode.Location{File: file1, Line: 10, Column: 5}
	loc2 := &sourcecode.Location{File: file1, Line: 20, Column: 8}
	loc3 := &sourcecode.Location{File: file2, Line: 5, Column: 0}

	debugInfo.InstructionLocations[0x1000] = loc1
	debugInfo.InstructionLocations[0x1004] = loc2
	debugInfo.InstructionLocations[0x2000] = loc3

	assert.Equal(t, loc1, debugInfo.GetSourceLocation(0x1000))
	assert.Equal(t, loc2, debugInfo.GetSourceLocation(0x1004))
	assert.Equal(t, loc3, debugInfo.GetSourceLocation(0x2000))
	assert.Nil(t, debugInfo.GetSourceLocation(0x3000))
}

// Tests for GetVariables

func TestGetVariables_NilDebugInfo(t *testing.T) {
	var debugInfo *DebugInfo
	vars := debugInfo.GetVariables(0x1000)
	assert.Nil(t, vars)
}

func TestGetVariables_AddressNotFound(t *testing.T) {
	debugInfo := NewDebugInfo()
	vars := debugInfo.GetVariables(0x1000)
	assert.Nil(t, vars)
}

func TestGetVariables_NoVariablesAtAddress(t *testing.T) {
	debugInfo := NewDebugInfo()
	debugInfo.InstructionVariables[0x1000] = make([]VariableInfo, 0)

	vars := debugInfo.GetVariables(0x1000)
	require.NotNil(t, vars)
	assert.Equal(t, 0, len(vars))
}

func TestGetVariables_SingleVariable(t *testing.T) {
	debugInfo := NewDebugInfo()
	varInfo := VariableInfo{
		Name:        "counter",
		TypeName:    "int",
		Size:        4,
		Location:    RegisterLocation{Register: 1},
		IsParameter: false,
	}
	debugInfo.InstructionVariables[0x1000] = []VariableInfo{varInfo}

	vars := debugInfo.GetVariables(0x1000)
	require.NotNil(t, vars)
	require.Equal(t, 1, len(vars))
	assert.Equal(t, varInfo, vars[0])
}

func TestGetVariables_MultipleVariables(t *testing.T) {
	debugInfo := NewDebugInfo()
	vars := []VariableInfo{
		{Name: "x", TypeName: "int", Size: 4, Location: RegisterLocation{Register: 1}},
		{Name: "y", TypeName: "int", Size: 4, Location: RegisterLocation{Register: 2}},
		{Name: "ptr", TypeName: "int*", Size: 4, Location: MemoryLocation{BaseRegister: 14, Offset: -8}},
	}
	debugInfo.InstructionVariables[0x1000] = vars

	retrievedVars := debugInfo.GetVariables(0x1000)
	require.NotNil(t, retrievedVars)
	require.Equal(t, len(vars), len(retrievedVars))
	for i, v := range retrievedVars {
		assert.Equal(t, vars[i], v)
	}
}

// Tests for SortedSourceLocations

func TestSortedSourceLocations_NilDebugInfo(t *testing.T) {
	var debugInfo *DebugInfo
	result := debugInfo.SortedSourceLocations()
	assert.Nil(t, result)
}

func TestSortedSourceLocations_EmptyDebugInfo(t *testing.T) {
	debugInfo := NewDebugInfo()
	result := debugInfo.SortedSourceLocations()
	require.NotNil(t, result)
	assert.Equal(t, 0, len(result))
}

func TestSortedSourceLocations_SingleLocation(t *testing.T) {
	debugInfo := NewDebugInfo()
	file := sourcecode.FileNamed("test.c")
	loc := &sourcecode.Location{File: file, Line: 10, Column: 5}
	debugInfo.InstructionLocations[0x1000] = loc

	result := debugInfo.SortedSourceLocations()
	require.NotNil(t, result)
	require.Equal(t, 1, len(result))
	assert.Equal(t, uint32(0x1000), result[0].Address)
	assert.Equal(t, loc, result[0].Location)
}

func TestSortedSourceLocations_MultipleLocations_IsSorted(t *testing.T) {
	debugInfo := NewDebugInfo()
	file := sourcecode.FileNamed("test.c")
	loc1 := &sourcecode.Location{File: file, Line: 10, Column: 5}
	loc2 := &sourcecode.Location{File: file, Line: 20, Column: 8}
	loc3 := &sourcecode.Location{File: file, Line: 30, Column: 0}

	// Add in reverse order to verify sorting
	debugInfo.InstructionLocations[0x2000] = loc3
	debugInfo.InstructionLocations[0x1000] = loc1
	debugInfo.InstructionLocations[0x1004] = loc2

	result := debugInfo.SortedSourceLocations()
	require.NotNil(t, result)
	require.Equal(t, 3, len(result))

	// Verify sorted order
	assert.Equal(t, uint32(0x1000), result[0].Address)
	assert.Equal(t, uint32(0x1004), result[1].Address)
	assert.Equal(t, uint32(0x2000), result[2].Address)

	// Verify locations match
	assert.Equal(t, loc1, result[0].Location)
	assert.Equal(t, loc2, result[1].Location)
	assert.Equal(t, loc3, result[2].Location)
}

func TestSortedSourceLocations_RandomOrder(t *testing.T) {
	debugInfo := NewDebugInfo()
	addresses := []uint32{0x5000, 0x1000, 0x3000, 0x2000, 0x4000}
	locations := make(map[uint32]*sourcecode.Location)
	file := sourcecode.FileNamed("test.c")

	for i, addr := range addresses {
		loc := &sourcecode.Location{
			File:   file,
			Line:   10 + i,
			Column: 0,
		}
		debugInfo.InstructionLocations[addr] = loc
		locations[addr] = loc
	}

	result := debugInfo.SortedSourceLocations()
	require.NotNil(t, result)
	require.Equal(t, len(addresses), len(result))

	// Verify strictly increasing addresses
	for i := 0; i < len(result)-1; i++ {
		assert.Less(t, result[i].Address, result[i+1].Address)
	}
}

func TestSortedSourceLocations_LargeAddressGaps(t *testing.T) {
	debugInfo := NewDebugInfo()
	fileA := sourcecode.FileNamed("a.c")
	fileB := sourcecode.FileNamed("b.c")
	fileC := sourcecode.FileNamed("c.c")
	debugInfo.InstructionLocations[0x00000001] = &sourcecode.Location{File: fileA, Line: 1}
	debugInfo.InstructionLocations[0x80000000] = &sourcecode.Location{File: fileB, Line: 2}
	debugInfo.InstructionLocations[0x00001000] = &sourcecode.Location{File: fileC, Line: 3}

	result := debugInfo.SortedSourceLocations()
	require.NotNil(t, result)
	require.Equal(t, 3, len(result))

	assert.Equal(t, uint32(0x00000001), result[0].Address)
	assert.Equal(t, uint32(0x00001000), result[1].Address)
	assert.Equal(t, uint32(0x80000000), result[2].Address)
}

// Tests for VariableInfo integration

func TestVariableInfo_RegisterLocation(t *testing.T) {
	info := VariableInfo{
		Name:        "x",
		TypeName:    "int",
		Size:        4,
		Location:    RegisterLocation{Register: 1},
		IsParameter: true,
	}

	assert.Equal(t, "x", info.Name)
	assert.Equal(t, "int", info.TypeName)
	assert.Equal(t, 4, info.Size)
	assert.Equal(t, VariableLocationRegister, info.Location.Type())
	assert.True(t, info.IsParameter)
}

func TestVariableInfo_MemoryLocation(t *testing.T) {
	info := VariableInfo{
		Name:        "buffer",
		TypeName:    "char[256]",
		Size:        256,
		Location:    MemoryLocation{BaseRegister: 15, Offset: -256},
		IsParameter: false,
	}

	assert.Equal(t, "buffer", info.Name)
	assert.Equal(t, "char[256]", info.TypeName)
	assert.Equal(t, 256, info.Size)
	assert.Equal(t, VariableLocationMemory, info.Location.Type())
	assert.False(t, info.IsParameter)
}

func TestVariableInfo_ConstantLocation(t *testing.T) {
	info := VariableInfo{
		Name:        "MAX_SIZE",
		TypeName:    "const int",
		Size:        4,
		Location:    ConstantLocation{Value: 1024},
		IsParameter: false,
	}

	assert.Equal(t, "MAX_SIZE", info.Name)
	assert.Equal(t, "const int", info.TypeName)
	assert.Equal(t, VariableLocationConstant, info.Location.Type())
}

// Tests for FunctionDebugInfo

func TestFunctionDebugInfo_Creation(t *testing.T) {
	funcInfo := FunctionDebugInfo{
		Name:           "main",
		StartAddress:   0x1000,
		EndAddress:     0x1100,
		SourceFile:     "main.c",
		StartLine:      1,
		EndLine:        50,
		Parameters:     make([]VariableInfo, 0),
		LocalVariables: make([]VariableInfo, 0),
		Scopes:         make([]ScopeInfo, 0),
	}

	assert.Equal(t, "main", funcInfo.Name)
	assert.Equal(t, uint32(0x1000), funcInfo.StartAddress)
	assert.Equal(t, uint32(0x1100), funcInfo.EndAddress)
	assert.Equal(t, "main.c", funcInfo.SourceFile)
	assert.Equal(t, 1, funcInfo.StartLine)
	assert.Equal(t, 50, funcInfo.EndLine)
}

// Tests for ScopeInfo

func TestScopeInfo_Creation(t *testing.T) {
	scope := ScopeInfo{
		StartAddress: 0x1020,
		EndAddress:   0x1050,
		Variables: []VariableInfo{
			{Name: "local_var", TypeName: "int", Size: 4, Location: MemoryLocation{BaseRegister: 14, Offset: -4}},
		},
	}

	assert.Equal(t, uint32(0x1020), scope.StartAddress)
	assert.Equal(t, uint32(0x1050), scope.EndAddress)
	require.Equal(t, 1, len(scope.Variables))
	assert.Equal(t, "local_var", scope.Variables[0].Name)
}

// Integration tests

func TestDebugInfo_CompleteFlow(t *testing.T) {
	debugInfo := NewDebugInfo()
	file := sourcecode.FileNamed("test.c")

	// Add locations
	loc1 := &sourcecode.Location{File: file, Line: 10, Column: 5}
	loc2 := &sourcecode.Location{File: file, Line: 15, Column: 0}
	debugInfo.InstructionLocations[0x1000] = loc1
	debugInfo.InstructionLocations[0x1004] = loc2

	// Add variables
	debugInfo.InstructionVariables[0x1000] = []VariableInfo{
		{Name: "x", TypeName: "int", Size: 4, Location: RegisterLocation{Register: 1}},
	}

	// Verify retrieval
	assert.Equal(t, loc1, debugInfo.GetSourceLocation(0x1000))
	assert.Equal(t, loc2, debugInfo.GetSourceLocation(0x1004))

	vars := debugInfo.GetVariables(0x1000)
	require.NotNil(t, vars)
	require.Equal(t, 1, len(vars))
	assert.Equal(t, "x", vars[0].Name)

	// Verify sorted locations
	sorted := debugInfo.SortedSourceLocations()
	require.Equal(t, 2, len(sorted))
	assert.Equal(t, uint32(0x1000), sorted[0].Address)
	assert.Equal(t, uint32(0x1004), sorted[1].Address)
}

func TestDebugInfo_ManyVariablesPerAddress(t *testing.T) {
	debugInfo := NewDebugInfo()

	vars := []VariableInfo{
		{Name: "param1", TypeName: "int", Size: 4, Location: RegisterLocation{Register: 0}, IsParameter: true},
		{Name: "param2", TypeName: "int*", Size: 4, Location: RegisterLocation{Register: 1}, IsParameter: true},
		{Name: "local1", TypeName: "char", Size: 1, Location: MemoryLocation{BaseRegister: 14, Offset: -1}},
		{Name: "local2", TypeName: "int[10]", Size: 40, Location: MemoryLocation{BaseRegister: 14, Offset: -44}},
	}

	debugInfo.InstructionVariables[0x1000] = vars

	retrievedVars := debugInfo.GetVariables(0x1000)
	require.NotNil(t, retrievedVars)
	require.Equal(t, len(vars), len(retrievedVars))

	for i, expected := range vars {
		assert.Equal(t, expected, retrievedVars[i])
	}
}
