package program

// Debug Information Support
//
// This file defines types for source-level debugging support in the Cucaracha CPU.
// Debug information allows the debugger and execution tracer to:
//
//   - Map machine instruction addresses to original source code locations (file, line, column)
//   - Track variable names, types, and storage locations at each instruction
//   - Display original source code during debugging sessions
//   - Provide source-level stepping and breakpoints
//
// Debug information is typically extracted from DWARF debug sections in ELF object files.
// When a C/C++ program is compiled with -g flag, the compiler embeds DWARF data that
// describes the mapping between machine code and source code.
//
// The debug information flow is:
//
//   1. C source compiled with clang -g produces ELF with DWARF sections
//   2. DWARFParser (in llvm package) extracts debug info from the ELF file
//   3. MemoryResolver remaps addresses when code is relocated to its runtime address
//   4. Debugger and exec tracer use the debug info to display source context
//
// Example usage in the debugger:
//
//   debugInfo := pf.DebugInfo()
//   if loc := debugInfo.GetSourceLocation(pc); loc != nil {
//       fmt.Printf("%s:%d %s\n", loc.File, loc.Line, debugInfo.GetSourceLine(loc.File, loc.Line))
//   }

import (
	"sort"

	"github.com/Manu343726/cucaracha/pkg/runtime/program/sourcecode"
)

// Defines the type of location where a variable's value is stored
type VariableLocationType int

const (
	VariableLocationRegister VariableLocationType = iota
	VariableLocationMemory
	VariableLocationConstant
)

func (vlt VariableLocationType) String() string {
	switch vlt {
	case VariableLocationRegister:
		return "Register"
	case VariableLocationMemory:
		return "Memory"
	case VariableLocationConstant:
		return "Constant"
	default:
		return "Unknown"
	}
}

// VariableLocation describes where a variable's value can be found at runtime
type VariableLocation interface {
	Type() VariableLocationType
}

// RegisterLocation indicates a variable is stored in a CPU register
type RegisterLocation struct {
	// Register is the register number (0-15 for r0-r15, or special register encoding)
	Register uint32
}

func (RegisterLocation) Type() VariableLocationType {
	return VariableLocationRegister
}

// MemoryLocation indicates a variable is stored in memory
type MemoryLocation struct {
	// BaseRegister is the base register for the address calculation (e.g., SP, FP)
	BaseRegister uint32
	// Offset is the offset from the base register
	Offset int32
}

func (MemoryLocation) Type() VariableLocationType {
	return VariableLocationMemory
}

// ConstantLocation indicates a variable has a constant value
type ConstantLocation struct {
	// Value is the constant value
	Value int64
}

func (ConstantLocation) Type() VariableLocationType {
	return VariableLocationConstant
}

// VariableInfo describes a source-level variable accessible at a given point
type VariableInfo struct {
	// Name is the variable name as it appears in the source code
	Name string
	// TypeName is the type of the variable (e.g., "int", "char*")
	TypeName string
	// Size is the size of the variable in bytes
	Size int
	// Location describes where the variable's value can be found
	Location VariableLocation
	// IsParameter indicates if this is a function parameter
	IsParameter bool
}

// DebugInfo contains all debug information for a program
type DebugInfo struct {
	// Source code library
	SourceLibrary sourcecode.Library

	// InstructionLocations maps instruction addresses to their source locations
	InstructionLocations map[uint32]*sourcecode.Location

	// InstructionVariables maps instruction addresses to variables accessible at that point
	// The slice is ordered with innermost scope variables first
	InstructionVariables map[uint32][]VariableInfo

	// Functions contains debug info for each function
	Functions map[string]*FunctionDebugInfo

	// CompilationUnit contains info about the compilation unit (source file)
	CompilationUnit string

	// Producer is the compiler/tool that produced the debug info
	Producer string
}

// FunctionDebugInfo contains debug information for a single function
type FunctionDebugInfo struct {
	// Name is the function name
	Name string
	// StartAddress is the address of the first instruction
	StartAddress uint32
	// EndAddress is the address after the last instruction
	EndAddress uint32
	// SourceFile is the file where the function is defined
	SourceFile string
	// StartLine is the line number where the function starts
	StartLine int
	// EndLine is the line number where the function ends
	EndLine int
	// Parameters are the function parameters
	Parameters []VariableInfo
	// LocalVariables are the local variables
	LocalVariables []VariableInfo
	// Scopes contains nested scopes within the function
	Scopes []ScopeInfo
}

// ScopeInfo describes a lexical scope (e.g., a block or loop body)
type ScopeInfo struct {
	// StartAddress is the address where this scope begins
	StartAddress uint32
	// EndAddress is the address where this scope ends
	EndAddress uint32
	// Variables are the variables declared in this scope
	Variables []VariableInfo
}

// NewDebugInfo creates an empty DebugInfo structure
func NewDebugInfo() *DebugInfo {
	return &DebugInfo{
		SourceLibrary:        sourcecode.NewSourceLibraryOnDisk(),
		InstructionLocations: make(map[uint32]*sourcecode.Location),
		InstructionVariables: make(map[uint32][]VariableInfo),
		Functions:            make(map[string]*FunctionDebugInfo),
	}
}

// GetSourceLocation returns the source location for an instruction address
func (d *DebugInfo) GetSourceLocation(addr uint32) *sourcecode.Location {
	if d == nil {
		return nil
	}
	return d.InstructionLocations[addr]
}

// GetVariables returns the variables accessible at an instruction address
func (d *DebugInfo) GetVariables(addr uint32) []VariableInfo {
	if d == nil {
		return nil
	}
	return d.InstructionVariables[addr]
}

// AddressLocation pairs an address with its source location for sorting
type AddressLocation struct {
	Address  uint32
	Location *sourcecode.Location
}

// SortedSourceLocations returns source locations sorted by address
func (d *DebugInfo) SortedSourceLocations() []AddressLocation {
	if d == nil {
		return nil
	}

	result := make([]AddressLocation, 0, len(d.InstructionLocations))
	for addr, loc := range d.InstructionLocations {
		result = append(result, AddressLocation{addr, loc})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Address < result[j].Address
	})

	return result
}
