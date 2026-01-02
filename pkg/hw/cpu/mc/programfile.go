package mc

import (
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
)

// MemoryLayout contains the resolved memory layout information for a program
type MemoryLayout struct {
	// BaseAddress is the starting address of the program in memory
	BaseAddress uint32
	// TotalSize is the total size of the program in bytes (code + data)
	TotalSize uint32
	// CodeSize is the size of the code section in bytes
	CodeSize uint32
	// DataSize is the size of the data section in bytes
	DataSize uint32
	// CodeStart is the starting address of the code section
	CodeStart uint32
	// DataStart is the starting address of the data section
	DataStart uint32
}

// GlobalType is an enum for global symbol types
type GlobalType int

const (
	GlobalUnknown GlobalType = iota
	GlobalFunction
	GlobalObject
)

// Global represents a global variable or constant in a program
type Global struct {
	Name        string
	Address     *uint32 // Resolved memory address (nil if not resolved)
	Size        int
	InitialData []byte
	Type        GlobalType
}

// InstructionRange represents a contiguous range of instructions
type InstructionRange struct {
	Start int
	Count int
}

// Function represents a function in a program
type Function struct {
	Name              string
	SourceFile        string
	StartLine         int
	EndLine           int
	InstructionRanges []InstructionRange
}

// SymbolKind indicates the kind of symbol being referenced
type SymbolKind int

const (
	SymbolKindUnknown SymbolKind = iota
	SymbolKindFunction
	SymbolKindGlobal
	SymbolKindLabel
)

// SymbolReferenceUsage indicates how a symbol reference accesses the symbol's address
type SymbolReferenceUsage int

const (
	SymbolUsageFull SymbolReferenceUsage = iota // Accesses the full symbol address
	SymbolUsageLo                               // Accesses the lower bits of the symbol address (@lo)
	SymbolUsageHi                               // Accesses the higher bits of the symbol address (@hi)
)

// SymbolReference represents a reference to a symbol in an instruction
type SymbolReference struct {
	Name     string               // The name of the referenced symbol (without @lo/@hi suffix)
	Usage    SymbolReferenceUsage // How the reference uses the symbol address
	Function *Function            // Pointer to referenced function (nil if not a function)
	Global   *Global              // Pointer to referenced global (nil if not a global)
	Label    *Label               // Pointer to referenced label (nil if not a label)
}

// BaseName returns the symbol name (same as Name since @lo/@hi suffixes are stored in Usage)
func (s *SymbolReference) BaseName() string {
	return s.Name
}

// Kind returns the kind of symbol being referenced
func (s *SymbolReference) Kind() SymbolKind {
	if s.Function != nil {
		return SymbolKindFunction
	}
	if s.Global != nil {
		return SymbolKindGlobal
	}
	if s.Label != nil {
		return SymbolKindLabel
	}
	return SymbolKindUnknown
}

// Unresolved returns true if the symbol reference has not been resolved
func (s *SymbolReference) Unresolved() bool {
	return s.Function == nil && s.Global == nil && s.Label == nil
}

// Instruction represents a single instruction in a program
type Instruction struct {
	LineNumber  int
	Address     *uint32                      // Resolved memory address (nil if not resolved)
	Text        string                       // Assembly text representation
	Raw         *instructions.RawInstruction // Raw binary instruction (partially decoded)
	Instruction *instructions.Instruction    // Fully decoded instruction ready for the interpreter
	Symbols     []SymbolReference
}

// Label represents a label and its associated instruction
type Label struct {
	Name             string
	InstructionIndex int // index into Instructions array, -1 if not pointing to an instruction
}

// ProgramFile is an interface for loading cucaracha programs from a file.
// It provides access to the program's metadata, functions, instructions, globals, and labels.
type ProgramFile interface {
	// FileName returns the path to the program file
	FileName() string

	// SourceFile returns the original source file name (e.g., "main.c")
	SourceFile() string

	// Functions returns all functions in the program
	Functions() map[string]Function

	// Instructions returns all instructions in the program, in order
	Instructions() []Instruction

	// Globals returns all global symbols in the program
	Globals() []Global

	// Labels returns all labels in the program
	Labels() []Label

	// MemoryLayout returns the memory layout information, or nil if not resolved
	MemoryLayout() *MemoryLayout
}

// ProgramFileContents is a struct that stores all the contents of a program file.
// It implements the ProgramFile interface and can be embedded in other types
// to easily provide ProgramFile functionality.
type ProgramFileContents struct {
	FileNameValue     string
	SourceFileValue   string
	FunctionsValue    map[string]Function
	InstructionsValue []Instruction
	GlobalsValue      []Global
	LabelsValue       []Label
	MemoryLayoutValue *MemoryLayout
}

// FileName returns the path to the program file
func (p *ProgramFileContents) FileName() string {
	return p.FileNameValue
}

// SourceFile returns the original source file name
func (p *ProgramFileContents) SourceFile() string {
	return p.SourceFileValue
}

// Functions returns all functions in the program
func (p *ProgramFileContents) Functions() map[string]Function {
	return p.FunctionsValue
}

// Instructions returns all instructions in the program
func (p *ProgramFileContents) Instructions() []Instruction {
	return p.InstructionsValue
}

// Globals returns all global symbols in the program
func (p *ProgramFileContents) Globals() []Global {
	return p.GlobalsValue
}

// Labels returns all labels in the program
func (p *ProgramFileContents) Labels() []Label {
	return p.LabelsValue
}

// MemoryLayout returns the memory layout information, or nil if not resolved
func (p *ProgramFileContents) MemoryLayout() *MemoryLayout {
	return p.MemoryLayoutValue
}

// Resolve applies all resolvers to a ProgramFile in the correct order:
// 1. Symbol resolution - resolves all symbol references (functions, globals, labels)
// 2. Memory resolution - assigns memory addresses to instructions and globals
// 3. Instruction resolution - decodes/encodes instructions (Text <-> Raw <-> Instruction)
//
// Returns a new fully resolved ProgramFile, or an error if any resolution step fails.
func Resolve(pf ProgramFile, memoryConfig MemoryResolverConfig) (ProgramFile, error) {
	// Step 1: Resolve symbols
	symbolResolved, err := ResolveSymbols(pf)
	if err != nil {
		return nil, err
	}

	// Step 2: Resolve memory addresses
	memoryResolved, err := ResolveMemory(symbolResolved, memoryConfig)
	if err != nil {
		return nil, err
	}

	// Step 3: Resolve instructions
	resolver := NewInstructionResolver()
	instructionResolved, err := resolver.ResolveWithContext(memoryResolved)
	if err != nil {
		return nil, err
	}

	return instructionResolved, nil
}
