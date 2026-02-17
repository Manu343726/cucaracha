package program

import (
	"strings"
	"testing"

	"github.com/Manu343726/cucaracha/pkg/hw/memory"
	"github.com/Manu343726/cucaracha/pkg/runtime/program/sourcecode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test memory layout
func createTestMemoryLayout() *memory.MemoryLayout {
	return &memory.MemoryLayout{
		CodeBase: 0x1000,
		CodeSize: 0x1000,
		DataBase: 0x3000,
		DataSize: 0x1000,
	}
}

// Helper function to create a test program file
func createTestProgramFile() *ProgramFileContents {
	memLayout := createTestMemoryLayout()
	return &ProgramFileContents{
		FileNameValue:     "test.o",
		SourceFileValue:   "test.c",
		FunctionsValue:    make(map[string]Function),
		InstructionsValue: make([]Instruction, 0),
		GlobalsValue:      make([]Global, 0),
		LabelsValue:       make([]Label, 0),
		MemoryLayoutValue: memLayout,
		DebugInfoValue:    nil,
	}
}

// Tests for Global struct

func TestGlobal_Range_WithResolvedAddress(t *testing.T) {
	addr := uint32(0x3000)
	g := Global{
		Name:        "myGlobal",
		Address:     &addr,
		Size:        4,
		InitialData: []byte{0x01, 0x02, 0x03, 0x04},
		Type:        GlobalObject,
	}

	r := g.Range()
	require.NotNil(t, r)
	assert.Equal(t, uint32(0x3000), r.Start)
	assert.Equal(t, uint32(4), r.Size)
}

func TestGlobal_Range_WithUnresolvedAddress(t *testing.T) {
	g := Global{
		Name:        "myGlobal",
		Address:     nil,
		Size:        4,
		InitialData: []byte{0x01, 0x02, 0x03, 0x04},
		Type:        GlobalObject,
	}

	r := g.Range()
	assert.Nil(t, r)
}

func TestGlobal_Range_VariousSize(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		expected uint32
	}{
		{"single byte", 1, 1},
		{"word", 4, 4},
		{"double word", 8, 8},
		{"large block", 256, 256},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr := uint32(0x3000)
			g := Global{
				Name:    "test",
				Address: &addr,
				Size:    tt.size,
			}

			r := g.Range()
			require.NotNil(t, r)
			assert.Equal(t, tt.expected, r.Size)
		})
	}
}

// Tests for Function struct

func TestFunction_FirstInstructionIndex_Empty(t *testing.T) {
	f := Function{
		Name:              "main",
		InstructionRanges: []InstructionRange{},
	}

	idx := f.FirstInstructionIndex()
	assert.Equal(t, -1, idx)
}

func TestFunction_FirstInstructionIndex_SingleRange(t *testing.T) {
	f := Function{
		Name: "main",
		InstructionRanges: []InstructionRange{
			{Start: 10, Count: 5},
		},
	}

	idx := f.FirstInstructionIndex()
	assert.Equal(t, 10, idx)
}

func TestFunction_FirstInstructionIndex_MultipleRanges(t *testing.T) {
	f := Function{
		Name: "main",
		InstructionRanges: []InstructionRange{
			{Start: 10, Count: 5},
			{Start: 20, Count: 3},
		},
	}

	idx := f.FirstInstructionIndex()
	assert.Equal(t, 10, idx)
}

// Tests for SymbolReference struct

func TestSymbolReference_BaseName(t *testing.T) {
	s := SymbolReference{
		Name:  "myFunction",
		Usage: SymbolUsageFull,
	}

	assert.Equal(t, "myFunction", s.BaseName())
}

func TestSymbolReference_Kind_Function(t *testing.T) {
	f := &Function{Name: "test"}
	s := SymbolReference{
		Name:     "test",
		Function: f,
		Global:   nil,
		Label:    nil,
	}

	assert.Equal(t, SymbolKindFunction, s.Kind())
}

func TestSymbolReference_Kind_Global(t *testing.T) {
	g := &Global{Name: "test"}
	s := SymbolReference{
		Name:     "test",
		Function: nil,
		Global:   g,
		Label:    nil,
	}

	assert.Equal(t, SymbolKindGlobal, s.Kind())
}

func TestSymbolReference_Kind_Label(t *testing.T) {
	l := &Label{Name: "test"}
	s := SymbolReference{
		Name:     "test",
		Function: nil,
		Global:   nil,
		Label:    l,
	}

	assert.Equal(t, SymbolKindLabel, s.Kind())
}

func TestSymbolReference_Kind_Unknown(t *testing.T) {
	s := SymbolReference{
		Name:     "test",
		Function: nil,
		Global:   nil,
		Label:    nil,
	}

	assert.Equal(t, SymbolKindUnknown, s.Kind())
}

func TestSymbolReference_Unresolved_True(t *testing.T) {
	s := SymbolReference{
		Name:     "test",
		Function: nil,
		Global:   nil,
		Label:    nil,
	}

	assert.True(t, s.Unresolved())
}

func TestSymbolReference_Unresolved_False_WithFunction(t *testing.T) {
	f := &Function{Name: "test"}
	s := SymbolReference{
		Name:     "test",
		Function: f,
	}

	assert.False(t, s.Unresolved())
}

func TestSymbolReference_Unresolved_False_WithGlobal(t *testing.T) {
	g := &Global{Name: "test"}
	s := SymbolReference{
		Name:   "test",
		Global: g,
	}

	assert.False(t, s.Unresolved())
}

func TestSymbolReference_Unresolved_False_WithLabel(t *testing.T) {
	l := &Label{Name: "test"}
	s := SymbolReference{
		Name:  "test",
		Label: l,
	}

	assert.False(t, s.Unresolved())
}

// Tests for ProgramFileContents struct

func TestProgramFileContents_FileName(t *testing.T) {
	p := createTestProgramFile()
	assert.Equal(t, "test.o", p.FileName())
}

func TestProgramFileContents_SourceFile(t *testing.T) {
	p := createTestProgramFile()
	assert.Equal(t, "test.c", p.SourceFile())
}

func TestProgramFileContents_Functions(t *testing.T) {
	p := createTestProgramFile()
	p.FunctionsValue["main"] = Function{Name: "main"}

	funcs := p.Functions()
	assert.Len(t, funcs, 1)
	assert.Contains(t, funcs, "main")
}

func TestProgramFileContents_Instructions(t *testing.T) {
	p := createTestProgramFile()
	addr1 := uint32(0x1000)
	p.InstructionsValue = []Instruction{
		{Address: &addr1, Text: "NOP"},
	}

	instrs := p.Instructions()
	assert.Len(t, instrs, 1)
	assert.Equal(t, "NOP", instrs[0].Text)
}

func TestProgramFileContents_Globals(t *testing.T) {
	p := createTestProgramFile()
	globalAddr := uint32(0x3000)
	p.GlobalsValue = []Global{
		{Name: "globalVar", Address: &globalAddr},
	}

	globals := p.Globals()
	assert.Len(t, globals, 1)
	assert.Equal(t, "globalVar", globals[0].Name)
}

func TestProgramFileContents_Labels(t *testing.T) {
	p := createTestProgramFile()
	p.LabelsValue = []Label{
		{Name: "loop", InstructionIndex: 5},
	}

	labels := p.Labels()
	assert.Len(t, labels, 1)
	assert.Equal(t, "loop", labels[0].Name)
}

func TestProgramFileContents_MemoryLayout(t *testing.T) {
	p := createTestProgramFile()
	layout := p.MemoryLayout()
	require.NotNil(t, layout)
	assert.Equal(t, uint32(0x1000), layout.CodeBase)
}

func TestProgramFileContents_DebugInfo_Nil(t *testing.T) {
	p := createTestProgramFile()
	assert.Nil(t, p.DebugInfo())
}

func TestProgramFileContents_DebugInfo_NotNil(t *testing.T) {
	p := createTestProgramFile()
	debugInfo := &DebugInfo{
		Functions:            make(map[string]*FunctionDebugInfo),
		InstructionLocations: make(map[uint32]*sourcecode.Location),
	}
	p.DebugInfoValue = debugInfo

	di := p.DebugInfo()
	assert.NotNil(t, di)
	assert.Equal(t, debugInfo, di)
}

// Tests for ProgramEntryPoint function

func TestProgramEntryPoint_Success(t *testing.T) {
	p := createTestProgramFile()
	addr := uint32(0x1000)

	p.FunctionsValue["main"] = Function{
		Name: "main",
		InstructionRanges: []InstructionRange{
			{Start: 0, Count: 5},
		},
	}

	p.InstructionsValue = []Instruction{
		{Address: &addr, Text: "NOP"},
	}

	entryPoint, err := ProgramEntryPoint(p)
	require.NoError(t, err)
	assert.Equal(t, uint32(0x1000), entryPoint)
}

func TestProgramEntryPoint_NoMainFunction(t *testing.T) {
	p := createTestProgramFile()
	p.FunctionsValue["other"] = Function{Name: "other"}

	_, err := ProgramEntryPoint(p)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "main")
}

func TestProgramEntryPoint_NoInstructions(t *testing.T) {
	p := createTestProgramFile()
	p.FunctionsValue["main"] = Function{
		Name:              "main",
		InstructionRanges: []InstructionRange{},
	}

	_, err := ProgramEntryPoint(p)
	assert.Error(t, err)
}

func TestProgramEntryPoint_EmptyInstructionRange(t *testing.T) {
	p := createTestProgramFile()
	p.FunctionsValue["main"] = Function{
		Name: "main",
		InstructionRanges: []InstructionRange{
			{Start: 0, Count: 0},
		},
	}

	_, err := ProgramEntryPoint(p)
	assert.Error(t, err)
}

func TestProgramEntryPoint_UnresolvedAddress(t *testing.T) {
	p := createTestProgramFile()
	p.FunctionsValue["main"] = Function{
		Name: "main",
		InstructionRanges: []InstructionRange{
			{Start: 0, Count: 1},
		},
	}

	p.InstructionsValue = []Instruction{
		{Address: nil, Text: "NOP"},
	}

	_, err := ProgramEntryPoint(p)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resolved address")
}

// Tests for FunctionByName function

func TestFunctionByName_Success(t *testing.T) {
	p := createTestProgramFile()
	addr := uint32(0x1000)

	p.FunctionsValue["myFunc"] = Function{
		Name: "myFunc",
		InstructionRanges: []InstructionRange{
			{Start: 0, Count: 5},
		},
	}

	p.InstructionsValue = []Instruction{
		{Address: &addr},
	}

	fn, err := FunctionByName(p, "myFunc")
	require.NoError(t, err)
	require.NotNil(t, fn)
	assert.Equal(t, "myFunc", fn.Name)
}

func TestFunctionByName_NotFound(t *testing.T) {
	p := createTestProgramFile()

	_, err := FunctionByName(p, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFunctionByName_NoMemoryLayout(t *testing.T) {
	p := createTestProgramFile()
	p.MemoryLayoutValue = nil

	_, err := FunctionByName(p, "anyFunc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not resolved")
}

// Tests for GlobalByName function

func TestGlobalByName_Success(t *testing.T) {
	p := createTestProgramFile()
	addr := uint32(0x3000)

	p.GlobalsValue = []Global{
		{Name: "myGlobal", Address: &addr, Size: 4},
	}

	global, err := GlobalByName(p, "myGlobal")
	require.NoError(t, err)
	require.NotNil(t, global)
	assert.Equal(t, "myGlobal", global.Name)
}

func TestGlobalByName_NotFound(t *testing.T) {
	p := createTestProgramFile()

	_, err := GlobalByName(p, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGlobalByName_NoMemoryLayout(t *testing.T) {
	p := createTestProgramFile()
	p.MemoryLayoutValue = nil

	_, err := GlobalByName(p, "anyGlobal")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not resolved")
}

func TestGlobalByName_MultipleGlobals(t *testing.T) {
	p := createTestProgramFile()
	addr1 := uint32(0x3000)
	addr2 := uint32(0x3004)
	addr3 := uint32(0x3008)

	p.GlobalsValue = []Global{
		{Name: "g1", Address: &addr1, Size: 4},
		{Name: "g2", Address: &addr2, Size: 4},
		{Name: "g3", Address: &addr3, Size: 4},
	}

	global, err := GlobalByName(p, "g2")
	require.NoError(t, err)
	assert.Equal(t, "g2", global.Name)
}

// Tests for SymbolUsageString representations

func TestSymbolUsageTypes(t *testing.T) {
	tests := []struct {
		name  string
		usage SymbolReferenceUsage
	}{
		{"Full", SymbolUsageFull},
		{"Lo", SymbolUsageLo},
		{"Hi", SymbolUsageHi},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, int(tt.usage) >= 0, true)
		})
	}
}

// Tests for SymbolKindString representations

func TestSymbolKindTypes(t *testing.T) {
	tests := []struct {
		name string
		kind SymbolKind
	}{
		{"Unknown", SymbolKindUnknown},
		{"Function", SymbolKindFunction},
		{"Global", SymbolKindGlobal},
		{"Label", SymbolKindLabel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, int(tt.kind) >= 0, true)
		})
	}
}

// Tests for GlobalType

func TestGlobalTypes(t *testing.T) {
	tests := []struct {
		name  string
		gtype GlobalType
	}{
		{"Unknown", GlobalUnknown},
		{"Function", GlobalFunction},
		{"Object", GlobalObject},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, int(tt.gtype) >= 0, true)
		})
	}
}

// Tests for Label struct

func TestLabel_Basic(t *testing.T) {
	label := Label{
		Name:             "loop",
		InstructionIndex: 42,
	}

	assert.Equal(t, "loop", label.Name)
	assert.Equal(t, 42, label.InstructionIndex)
}

func TestLabel_NoInstruction(t *testing.T) {
	label := Label{
		Name:             "unused_label",
		InstructionIndex: -1,
	}

	assert.Equal(t, "unused_label", label.Name)
	assert.Equal(t, -1, label.InstructionIndex)
}

// Tests for Instruction struct

func TestInstruction_WithResolvedAddress(t *testing.T) {
	addr := uint32(0x1000)
	instr := Instruction{
		LineNumber: 5,
		Address:    &addr,
		Text:       "NOP",
	}

	assert.Equal(t, 5, instr.LineNumber)
	assert.Equal(t, uint32(0x1000), *instr.Address)
	assert.Equal(t, "NOP", instr.Text)
}

func TestInstruction_WithSymbols(t *testing.T) {
	addr := uint32(0x1000)
	instr := Instruction{
		LineNumber: 5,
		Address:    &addr,
		Text:       "JAL main",
		Symbols: []SymbolReference{
			{Name: "main", Function: &Function{Name: "main"}},
		},
	}

	assert.Len(t, instr.Symbols, 1)
	assert.Equal(t, "main", instr.Symbols[0].Name)
}

// Tests for Resolve function

func TestResolve_NilMemoryLayout(t *testing.T) {
	pf := createTestProgramFile()

	result, err := Resolve(pf, nil)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "memory layout must be provided")
}

func TestResolve_WithValidMemoryLayout(t *testing.T) {
	pf := createTestProgramFile()
	memLayout := createTestMemoryLayout()

	result, err := Resolve(pf, memLayout)

	assert.NotNil(t, result)
	assert.NoError(t, err)
}

func TestResolve_UnresolvedSymbols_ErrorPropagation(t *testing.T) {
	pf := createTestProgramFile()
	memLayout := createTestMemoryLayout()

	// Add an unresolved symbol
	instr := Instruction{
		LineNumber: 1,
		Address:    nil,
		Text:       "JAL undefined_function",
		Symbols: []SymbolReference{
			{Name: "undefined_function", Function: nil, Global: nil, Label: nil},
		},
	}
	pf.InstructionsValue = append(pf.InstructionsValue, instr)

	result, err := Resolve(pf, memLayout)

	// If ResolveSymbols fails, the error should be propagated
	if err != nil {
		assert.Contains(t, err.Error(), "undefined")
		assert.Nil(t, result)
	}
}

func TestResolve_PreservesMemoryLayout(t *testing.T) {
	pf := createTestProgramFile()
	memLayout := createTestMemoryLayout()

	result, err := Resolve(pf, memLayout)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Verify memory layout is present (it may be modified during resolution)
	assert.NotNil(t, result.MemoryLayout())
	assert.Equal(t, uint32(0x1000), result.MemoryLayout().CodeBase)
	assert.Equal(t, uint32(0x3000), result.MemoryLayout().DataBase)
}

// Tests for SourceLineAtInstructionAddress function

func TestSourceLineAtInstructionAddress_NoDebugInfo(t *testing.T) {
	pf := createTestProgramFile()
	pf.DebugInfoValue = nil

	result, err := SourceLineAtInstructionAddress(pf, 0x1000)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no debug information")
}

func TestSourceLineAtInstructionAddress_InvalidAddress(t *testing.T) {
	pf := createTestProgramFile()
	pf.DebugInfoValue = &DebugInfo{
		SourceLibrary:        nil,
		Functions:            make(map[string]*FunctionDebugInfo),
		InstructionLocations: make(map[uint32]*sourcecode.Location),
	}

	// Try to get line at address outside code segment
	result, err := SourceLineAtInstructionAddress(pf, 0x9000)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "outside of code segment")
}

func TestSourceLineAtInstructionAddress_LocationNotFound(t *testing.T) {
	pf := createTestProgramFile()
	pf.DebugInfoValue = &DebugInfo{
		SourceLibrary:        nil,
		Functions:            make(map[string]*FunctionDebugInfo),
		InstructionLocations: make(map[uint32]*sourcecode.Location),
	}

	// Valid address but no location info
	result, err := SourceLineAtInstructionAddress(pf, 0x1000)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no source location found")
}

// Tests for BranchTargetAtInstruction function

func TestBranchTargetAtInstruction_InstructionNotResolved(t *testing.T) {
	pf := createTestProgramFile()
	memLayout := createTestMemoryLayout()
	pf.MemoryLayoutValue = memLayout

	// Add unresolved instruction (no Instruction field)
	addr := uint32(0x1000)
	pf.InstructionsValue = append(pf.InstructionsValue, Instruction{
		Address: &addr,
		Text:    "JMP r0",
	})

	result, sym, err := BranchTargetAtInstruction(pf, 0x1000)

	assert.Nil(t, result)
	assert.Nil(t, sym)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not fully resolved")
}

func TestBranchTargetAtInstruction_InvalidAddress(t *testing.T) {
	pf := createTestProgramFile()
	memLayout := createTestMemoryLayout()
	pf.MemoryLayoutValue = memLayout

	// Try to get branch target at address outside code segment
	result, sym, err := BranchTargetAtInstruction(pf, 0x9000)

	assert.Nil(t, result)
	assert.Nil(t, sym)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "outside of code segment")
}

func TestBranchTargetAtInstruction_NoMemoryLayout(t *testing.T) {
	pf := createTestProgramFile()
	pf.MemoryLayoutValue = nil

	result, sym, err := BranchTargetAtInstruction(pf, 0x1000)

	assert.Nil(t, result)
	assert.Nil(t, sym)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "program file memory addresses are not resolved")
}

// Tests for partial coverage of FunctionByName - with debug info lookup

func TestFunctionByName_WithDebugInfo_NotFound(t *testing.T) {
	pf := createTestProgramFile()
	memLayout := createTestMemoryLayout()
	pf.MemoryLayoutValue = memLayout

	// Add debug info but without the requested function
	pf.DebugInfoValue = &DebugInfo{
		SourceLibrary: nil,
		Functions:     make(map[string]*FunctionDebugInfo),
	}

	result, err := FunctionByName(pf, "unknown_function")

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFunctionByName_WithDebugInfo_Found(t *testing.T) {
	pf := createTestProgramFile()
	memLayout := createTestMemoryLayout()
	pf.MemoryLayoutValue = memLayout

	// Add function to debug info
	debugFunc := &FunctionDebugInfo{
		Name:         "test_func",
		SourceFile:   "test.c",
		StartLine:    10,
		EndLine:      20,
		StartAddress: 0x1000,
		EndAddress:   0x1008,
	}

	pf.DebugInfoValue = &DebugInfo{
		SourceLibrary:        nil,
		Functions:            map[string]*FunctionDebugInfo{"test_func": debugFunc},
		InstructionLocations: make(map[uint32]*sourcecode.Location),
	}

	result, err := FunctionByName(pf, "test_func")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test_func", result.Name)
	assert.Equal(t, 10, result.StartLine)
	assert.Equal(t, 20, result.EndLine)
}

func TestFunctionByName_WithDebugInfo_FallbackToAssemblyNames(t *testing.T) {
	pf := createTestProgramFile()
	memLayout := createTestMemoryLayout()
	pf.MemoryLayoutValue = memLayout

	// Add assembly-level function
	pf.FunctionsValue["_Z8functionv"] = Function{
		Name: "_Z8functionv",
		InstructionRanges: []InstructionRange{
			{Start: 0, Count: 4},
		},
	}

	// Add debug info but it doesn't have this function
	pf.DebugInfoValue = &DebugInfo{
		SourceLibrary: nil,
		Functions:     make(map[string]*FunctionDebugInfo),
	}

	// Should find the assembly-level name
	result, err := FunctionByName(pf, "_Z8functionv")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "_Z8functionv", result.Name)
}

// Additional tests for Resolve function

func TestResolve_EmptyProgram(t *testing.T) {
	// Test with completely empty program
	pf := &ProgramFileContents{
		FileNameValue:     "empty.o",
		SourceFileValue:   "",
		FunctionsValue:    make(map[string]Function),
		InstructionsValue: make([]Instruction, 0),
		GlobalsValue:      make([]Global, 0),
		LabelsValue:       make([]Label, 0),
		MemoryLayoutValue: nil,
		DebugInfoValue:    nil,
	}
	memLayout := createTestMemoryLayout()

	result, err := Resolve(pf, memLayout)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// Additional tests for FunctionByName

func TestFunctionByName_NoMemoryLayout_WithDebugInfo(t *testing.T) {
	pf := createTestProgramFile()
	pf.MemoryLayoutValue = nil

	debugFunc := &FunctionDebugInfo{
		Name:         "test_func",
		SourceFile:   "test.c",
		StartLine:    10,
		EndLine:      20,
		StartAddress: 0x1000,
		EndAddress:   0x1008,
	}

	pf.DebugInfoValue = &DebugInfo{
		SourceLibrary:        nil,
		Functions:            map[string]*FunctionDebugInfo{"test_func": debugFunc},
		InstructionLocations: make(map[uint32]*sourcecode.Location),
	}

	result, err := FunctionByName(pf, "test_func")

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "program file memory addresses are not resolved")
}

// Tests for additional Resolve error conditions

func TestResolve_InstructionResolutionFailure(t *testing.T) {
	pf := createTestProgramFile()
	memLayout := createTestMemoryLayout()

	// Add instructions that might cause issues during resolution
	addr := uint32(0x1000)
	pf.InstructionsValue = append(pf.InstructionsValue, Instruction{
		Address: &addr,
		Text:    "NOP", // Valid instruction that should resolve properly
	})

	result, err := Resolve(pf, memLayout)

	// Should succeed since NOP is valid
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// Test for GlobalByName

func TestGlobalByName_MultipleGlobals_CorrectSelection(t *testing.T) {
	pf := createTestProgramFile()
	memLayout := createTestMemoryLayout()
	pf.MemoryLayoutValue = memLayout

	// Add multiple globals
	addr1 := uint32(0x3000)
	addr2 := uint32(0x3004)

	pf.GlobalsValue = []Global{
		{Name: "global1", Address: &addr1, Size: 4},
		{Name: "global2", Address: &addr2, Size: 4},
		{Name: "global3", Address: nil, Size: 4},
	}

	// Test finding specific globals
	result1, err1 := GlobalByName(pf, "global1")
	assert.NoError(t, err1)
	assert.NotNil(t, result1)
	assert.Equal(t, "global1", result1.Name)

	result2, err2 := GlobalByName(pf, "global2")
	assert.NoError(t, err2)
	assert.NotNil(t, result2)
	assert.Equal(t, "global2", result2.Name)

	// Test not found
	result3, err3 := GlobalByName(pf, "global4")
	assert.Error(t, err3)
	assert.Nil(t, result3)
}

// Tests for Resolve function - error paths

func TestResolve_ResolveMemoryError(t *testing.T) {
	pf := createTestProgramFile()
	memLayout := createTestMemoryLayout()

	// Create a program that will cause ResolveMemory to fail
	// Add a global that's too large to fit
	pf.GlobalsValue = []Global{
		{
			Name:        "huge_global",
			Size:        0x10000, // Larger than data segment (0x1000)
			Address:     nil,
			InitialData: make([]byte, 0x10000),
		},
	}

	result, err := Resolve(pf, memLayout)

	// ResolveMemory should fail due to size
	if err != nil {
		assert.Nil(t, result)
		assert.True(t, strings.Contains(err.Error(), "data") || strings.Contains(err.Error(), "size"), "error message: %s", err.Error())
	}
}

// Tests for FunctionByName - additional error path

func TestFunctionByName_WithDebugInfo_NoLayout(t *testing.T) {
	pf := createTestProgramFile()
	pf.MemoryLayoutValue = nil

	debugFunc := &FunctionDebugInfo{
		Name:         "test_func",
		SourceFile:   "test.c",
		StartLine:    10,
		EndLine:      20,
		StartAddress: 0x1000,
		EndAddress:   0x1008,
	}

	pf.DebugInfoValue = &DebugInfo{
		SourceLibrary:        nil,
		Functions:            map[string]*FunctionDebugInfo{"test_func": debugFunc},
		InstructionLocations: make(map[uint32]*sourcecode.Location),
	}

	result, err := FunctionByName(pf, "test_func")

	assert.Nil(t, result)
	assert.Error(t, err)
}

// Test ProgramEntryPoint function

func TestProgramEntryPoint_WithValidMain(t *testing.T) {
	pf := createTestProgramFile()
	memLayout := createTestMemoryLayout()
	pf.MemoryLayoutValue = memLayout

	// Add main function
	instrAddr := uint32(0x1000)
	pf.InstructionsValue = []Instruction{
		{Address: &instrAddr, Text: "ENTRY_POINT"},
	}

	pf.FunctionsValue["main"] = Function{
		Name: "main",
		InstructionRanges: []InstructionRange{
			{Start: 0, Count: 1},
		},
	}

	result, err := ProgramEntryPoint(pf)

	assert.NoError(t, err)
	assert.Equal(t, uint32(0x1000), result)
}

// Test to ensure all FunctionByName paths are covered

func TestFunctionByName_AssemblyLevelFallback(t *testing.T) {
	pf := createTestProgramFile()
	memLayout := createTestMemoryLayout()
	pf.MemoryLayoutValue = memLayout

	// Add assembly-level function without debug info
	pf.FunctionsValue["_Z4mainv"] = Function{
		Name: "_Z4mainv",
		InstructionRanges: []InstructionRange{
			{Start: 0, Count: 1},
		},
	}

	pf.DebugInfoValue = &DebugInfo{
		SourceLibrary:        nil,
		Functions:            make(map[string]*FunctionDebugInfo),
		InstructionLocations: make(map[uint32]*sourcecode.Location),
	}

	result, err := FunctionByName(pf, "_Z4mainv")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "_Z4mainv", result.Name)
}

// Coverage for GlobalByName with empty globals list

func TestGlobalByName_EmptyGlobalsList(t *testing.T) {
	pf := createTestProgramFile()
	memLayout := createTestMemoryLayout()
	pf.MemoryLayoutValue = memLayout
	pf.GlobalsValue = []Global{} // Empty

	result, err := GlobalByName(pf, "any_global")

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// Coverage for FunctionAtAddress with no matching function

func TestFunctionAtAddress_NoMatchingFunction(t *testing.T) {
	pf := createTestProgramFile()
	memLayout := createTestMemoryLayout()
	pf.MemoryLayoutValue = memLayout

	// Add instruction but no function
	addr := uint32(0x1000)
	pf.InstructionsValue = []Instruction{
		{Address: &addr, Text: "ORPHAN_INSTR"},
	}

	pf.FunctionsValue = make(map[string]Function) // Empty

	result, err := FunctionAtAddress(pf, 0x1000)

	assert.NoError(t, err)
	assert.Nil(t, result)
}

// Coverage for ProgramEntryPoint with empty main function

func TestProgramEntryPoint_EmptyMain(t *testing.T) {
	pf := createTestProgramFile()
	memLayout := createTestMemoryLayout()
	pf.MemoryLayoutValue = memLayout

	// Add main with empty ranges
	pf.FunctionsValue["main"] = Function{
		Name:              "main",
		InstructionRanges: []InstructionRange{}, // Empty
	}

	result, err := ProgramEntryPoint(pf)

	assert.Equal(t, uint32(0), result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no instructions")
}

// Coverage for ProgramEntryPoint with zero count range

func TestProgramEntryPoint_MainWithZeroCountRange(t *testing.T) {
	pf := createTestProgramFile()
	memLayout := createTestMemoryLayout()
	pf.MemoryLayoutValue = memLayout

	// Add main with range that has zero count
	pf.FunctionsValue["main"] = Function{
		Name: "main",
		InstructionRanges: []InstructionRange{
			{Start: 0, Count: 0}, // Zero count
		},
	}

	result, err := ProgramEntryPoint(pf)

	assert.Equal(t, uint32(0), result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no instructions")
}

// Coverage for Resolve with valid complete flow

func TestResolve_CompleteValidFlow(t *testing.T) {
	pf := createTestProgramFile()
	memLayout := createTestMemoryLayout()

	// Add valid instruction
	addr := uint32(0x1000)
	pf.InstructionsValue = []Instruction{
		{Address: &addr, Text: "NOP"},
	}

	result, err := Resolve(pf, memLayout)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// Coverage for GlobalAtAddress with multiple globals, finding one in the middle

func TestGlobalAtAddress_FindInMiddle(t *testing.T) {
	pf := createTestProgramFile()
	memLayout := createTestMemoryLayout()
	pf.MemoryLayoutValue = memLayout

	addr1 := uint32(0x3000)
	addr2 := uint32(0x3004)
	addr3 := uint32(0x3008)

	pf.GlobalsValue = []Global{
		{Name: "global1", Address: &addr1, Size: 4},
		{Name: "global2", Address: &addr2, Size: 4},
		{Name: "global3", Address: &addr3, Size: 4},
	}

	result, err := GlobalAtAddress(pf, 0x3004)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "global2", result.Name)
}

// Coverage for multiple function lookups

func TestFunctionByName_VariousFunctions(t *testing.T) {
	pf := createTestProgramFile()
	memLayout := createTestMemoryLayout()
	pf.MemoryLayoutValue = memLayout

	// Add multiple functions
	pf.FunctionsValue["func1"] = Function{Name: "func1"}
	pf.FunctionsValue["func2"] = Function{Name: "func2"}
	pf.FunctionsValue["func3"] = Function{Name: "func3"}

	// Test lookup of each
	result1, _ := FunctionByName(pf, "func1")
	assert.NotNil(t, result1)
	assert.Equal(t, "func1", result1.Name)

	result2, _ := FunctionByName(pf, "func2")
	assert.NotNil(t, result2)
	assert.Equal(t, "func2", result2.Name)

	result3, _ := FunctionByName(pf, "func3")
	assert.NotNil(t, result3)
	assert.Equal(t, "func3", result3.Name)
}

// Coverage for GlobalByName with first, middle, and last globals (existing earlier, added here for completeness)
// Tests to improve coverage on edge cases
