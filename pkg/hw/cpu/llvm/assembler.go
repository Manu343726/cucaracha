package llvm

import (
	"strings"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
)

// GlobalType is an enum for global symbol types
// (If you want string conversion, run go generate)
//
//go:generate stringer -type=GlobalType
type GlobalType int

const (
	GlobalUnknown GlobalType = iota
	GlobalFunction
	GlobalObject
)

// GlobalSymbol holds metadata for a global variable or constant
type GlobalSymbol struct {
	Name        string
	Size        int
	InitialData []byte
	Type        GlobalType // enum instead of string
}

// InstructionRange represents a contiguous range of instructions in the global array
type InstructionRange struct {
	Start int // index of first instruction in AssemblyFile.Instructions
	Count int // number of instructions in the range
}

// FunctionBody holds metadata and instruction ranges for a function
type FunctionBody struct {
	Name              string
	SourceFile        string
	StartLine         int
	EndLine           int
	InstructionRanges []InstructionRange
}

// Instruction holds a single instruction and its line number/location
type Instruction struct {
	LineNumber int
	Text       string
	Symbols    []string // referenced symbol names
}

// LabelSymbol represents a label and the instruction it points to
type LabelSymbol struct {
	Name        string
	Instruction *Instruction // nil if not pointing to an instruction
}

// AssemblyFile represents the parsed contents of a .cucaracha assembly file
type AssemblyFile struct {
	FileNameValue   string
	SourceFileValue string
	GlobalsValue    []GlobalSymbol
	FunctionsMap    map[string]*FunctionBody
	LabelsValue     []LabelSymbol // all labels found in code
	InstructionsAll []Instruction // all instructions in file, in order
}

// ParseAssemblyFile parses a .cucaracha assembly file and returns an in-memory representation
func ParseAssemblyFile(path string) (mc.ProgramFile, error) {
	parser := NewAssemblyFileParser(path)
	return parser.Parse()
}

// FileName returns the path to the assembly file
func (f *AssemblyFile) FileName() string {
	return f.FileNameValue
}

// SourceFile returns the original source file name
func (f *AssemblyFile) SourceFile() string {
	return f.SourceFileValue
}

// Functions returns all functions in the program as mc.Function values
func (f *AssemblyFile) Functions() map[string]mc.Function {
	result := make(map[string]mc.Function, len(f.FunctionsMap))
	for name, fn := range f.FunctionsMap {
		ranges := make([]mc.InstructionRange, len(fn.InstructionRanges))
		for i, r := range fn.InstructionRanges {
			ranges[i] = mc.InstructionRange{Start: r.Start, Count: r.Count}
		}
		result[name] = mc.Function{
			Name:              fn.Name,
			SourceFile:        fn.SourceFile,
			StartLine:         fn.StartLine,
			EndLine:           fn.EndLine,
			InstructionRanges: ranges,
		}
	}
	return result
}

// Instructions returns all instructions in the program
func (f *AssemblyFile) Instructions() []mc.Instruction {
	result := make([]mc.Instruction, len(f.InstructionsAll))
	for i, inst := range f.InstructionsAll {
		symbols := make([]mc.SymbolReference, len(inst.Symbols))
		for j, sym := range inst.Symbols {
			name := sym
			usage := mc.SymbolUsageFull
			if strings.HasSuffix(sym, "@lo") {
				name = strings.TrimSuffix(sym, "@lo")
				usage = mc.SymbolUsageLo
			} else if strings.HasSuffix(sym, "@hi") {
				name = strings.TrimSuffix(sym, "@hi")
				usage = mc.SymbolUsageHi
			}
			symbols[j] = mc.SymbolReference{Name: name, Usage: usage}
		}
		result[i] = mc.Instruction{
			LineNumber: inst.LineNumber,
			Text:       inst.Text,
			Symbols:    symbols,
		}
	}
	return result
}

// Globals returns all global symbols in the program
func (f *AssemblyFile) Globals() []mc.Global {
	result := make([]mc.Global, len(f.GlobalsValue))
	for i, g := range f.GlobalsValue {
		data := make([]byte, len(g.InitialData))
		copy(data, g.InitialData)
		result[i] = mc.Global{
			Name:        g.Name,
			Size:        g.Size,
			InitialData: data,
			Type:        mc.GlobalType(g.Type),
		}
	}
	return result
}

// Labels returns all labels in the program
func (f *AssemblyFile) Labels() []mc.Label {
	result := make([]mc.Label, len(f.LabelsValue))
	for i, lbl := range f.LabelsValue {
		instrIdx := -1
		if lbl.Instruction != nil {
			// Find the index of the instruction in the global array
			for j := range f.InstructionsAll {
				if &f.InstructionsAll[j] == lbl.Instruction {
					instrIdx = j
					break
				}
			}
		}
		result[i] = mc.Label{
			Name:             lbl.Name,
			InstructionIndex: instrIdx,
		}
	}
	return result
}

// MemoryLayout returns nil since parsed files don't have memory layout yet
func (f *AssemblyFile) MemoryLayout() *mc.MemoryLayout {
	return nil
}
