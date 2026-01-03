package mc

import (
	"errors"
	"fmt"
)

// MemoryResolverConfig contains configuration for memory resolution
type MemoryResolverConfig struct {
	// BaseAddress is the starting address for the program
	BaseAddress uint32
	// MaxSize is the maximum allowed size for the program (0 = unlimited)
	MaxSize uint32
	// DataAlignment is the alignment for the data section (default: 4)
	DataAlignment uint32
	// InstructionSize is the size of each instruction in bytes (default: 4)
	InstructionSize uint32
}

// DefaultMemoryResolverConfig returns a config with sensible defaults
func DefaultMemoryResolverConfig() MemoryResolverConfig {
	return MemoryResolverConfig{
		BaseAddress:     0,
		MaxSize:         0, // unlimited
		DataAlignment:   4,
		InstructionSize: 4,
	}
}

var (
	ErrProgramTooLarge        = errors.New("program exceeds maximum size")
	ErrUnresolvedSymbol       = errors.New("unresolved symbol reference")
	ErrUnresolvedMemory       = errors.New("unresolved memory address")
	ErrInvalidInstructionSize = errors.New("instruction size must be greater than 0")
)

// ResolveMemory assigns memory addresses to all instructions and globals in a ProgramFile.
// It returns a new ProgramFile with all addresses resolved and a MemoryLayout.
//
// The memory layout is:
//   - Code section starts at BaseAddress
//   - Data section follows the code section (aligned to DataAlignment)
//   - Symbol references are updated with the resolved addresses
func ResolveMemory(pf ProgramFile, config MemoryResolverConfig) (ProgramFile, error) {
	if config.InstructionSize == 0 {
		config.InstructionSize = 4
	}
	if config.DataAlignment == 0 {
		config.DataAlignment = 4
	}

	srcInstructions := pf.Instructions()
	srcGlobals := pf.Globals()
	srcFunctions := pf.Functions()
	srcLabels := pf.Labels()

	// Calculate code size
	codeSize := uint32(len(srcInstructions)) * config.InstructionSize
	codeStart := config.BaseAddress

	// Calculate data section start (aligned)
	dataStart := alignAddress(codeStart+codeSize, config.DataAlignment)

	// Calculate data size
	var dataSize uint32 = 0
	for _, g := range srcGlobals {
		dataSize += uint32(g.Size)
	}

	totalSize := (dataStart - config.BaseAddress) + dataSize

	// Check size limit
	if config.MaxSize > 0 && totalSize > config.MaxSize {
		return nil, fmt.Errorf("%w: program size %d exceeds max %d", ErrProgramTooLarge, totalSize, config.MaxSize)
	}

	// Create resolved instructions with addresses
	resolvedInstructions := make([]Instruction, len(srcInstructions))
	for i, inst := range srcInstructions {
		addr := codeStart + uint32(i)*config.InstructionSize
		resolvedInstructions[i] = Instruction{
			LineNumber:  inst.LineNumber,
			Address:     &addr,
			Text:        inst.Text,
			Raw:         inst.Raw,
			Instruction: inst.Instruction,
			Symbols:     make([]SymbolReference, len(inst.Symbols)),
		}
		// Copy symbols initially (will update references later)
		copy(resolvedInstructions[i].Symbols, inst.Symbols)
	}

	// Create resolved globals with addresses
	resolvedGlobals := make([]Global, len(srcGlobals))
	globalAddresses := make(map[string]uint32)
	currentDataAddr := dataStart
	for i, g := range srcGlobals {
		addr := currentDataAddr
		resolvedGlobals[i] = Global{
			Name:        g.Name,
			Address:     &addr,
			Size:        g.Size,
			InitialData: g.InitialData,
			Type:        g.Type,
		}
		globalAddresses[g.Name] = addr
		currentDataAddr += uint32(g.Size)
	}

	// Create resolved functions (copy, addresses derived from instruction addresses)
	resolvedFunctions := make(map[string]Function, len(srcFunctions))
	functionAddresses := make(map[string]uint32)
	for name, fn := range srcFunctions {
		resolvedFunctions[name] = fn
		// Function address is the address of its first instruction
		if len(fn.InstructionRanges) > 0 && fn.InstructionRanges[0].Count > 0 {
			startIdx := fn.InstructionRanges[0].Start
			if startIdx >= 0 && startIdx < len(resolvedInstructions) {
				functionAddresses[name] = *resolvedInstructions[startIdx].Address
			}
		}
	}

	// Create resolved labels with addresses (derived from instruction addresses)
	resolvedLabels := make([]Label, len(srcLabels))
	labelAddresses := make(map[string]uint32)
	for i, lbl := range srcLabels {
		resolvedLabels[i] = Label{
			Name:             lbl.Name,
			InstructionIndex: lbl.InstructionIndex,
		}
		if lbl.InstructionIndex >= 0 && lbl.InstructionIndex < len(resolvedInstructions) {
			labelAddresses[lbl.Name] = *resolvedInstructions[lbl.InstructionIndex].Address
		}
	}

	// Create pointer maps for symbol references
	globalPtrs := make(map[string]*Global)
	for i := range resolvedGlobals {
		globalPtrs[resolvedGlobals[i].Name] = &resolvedGlobals[i]
	}

	functionPtrMap := make(map[string]*Function)
	for name := range resolvedFunctions {
		fn := resolvedFunctions[name]
		functionPtrMap[name] = &fn
	}

	labelPtrs := make(map[string]*Label)
	for i := range resolvedLabels {
		labelPtrs[resolvedLabels[i].Name] = &resolvedLabels[i]
	}

	// Update symbol references in instructions
	for i := range resolvedInstructions {
		for j, sym := range resolvedInstructions[i].Symbols {
			resolved := SymbolReference{
				Name:  sym.Name,
				Usage: sym.Usage,
			}

			lookupName := sym.BaseName()

			// Look up symbol and set pointer
			if fn, ok := functionPtrMap[lookupName]; ok {
				resolved.Function = fn
			} else if g, ok := globalPtrs[lookupName]; ok {
				resolved.Global = g
			} else if lbl, ok := labelPtrs[lookupName]; ok {
				resolved.Label = lbl
			} else if sym.Function != nil || sym.Global != nil || sym.Label != nil {
				// Symbol was already resolved, keep the reference type but update pointers
				if sym.Function != nil {
					if fn, ok := functionPtrMap[lookupName]; ok {
						resolved.Function = fn
					}
				}
				if sym.Global != nil {
					if g, ok := globalPtrs[lookupName]; ok {
						resolved.Global = g
					}
				}
				if sym.Label != nil {
					if lbl, ok := labelPtrs[lookupName]; ok {
						resolved.Label = lbl
					}
				}
			}
			// Note: If symbol is still unresolved, we leave it unresolved
			// The caller can check for unresolved symbols if needed

			resolvedInstructions[i].Symbols[j] = resolved
		}
	}

	layout := &MemoryLayout{
		BaseAddress: config.BaseAddress,
		TotalSize:   totalSize,
		CodeSize:    codeSize,
		DataSize:    dataSize,
		CodeStart:   codeStart,
		DataStart:   dataStart,
	}

	// Remap debug info addresses to match the relocated code
	debugInfo := pf.DebugInfo()
	if debugInfo != nil {
		debugInfo = remapDebugInfoAddresses(debugInfo, codeStart)
	}

	return &ProgramFileContents{
		FileNameValue:     pf.FileName(),
		SourceFileValue:   pf.SourceFile(),
		FunctionsValue:    resolvedFunctions,
		InstructionsValue: resolvedInstructions,
		GlobalsValue:      resolvedGlobals,
		LabelsValue:       resolvedLabels,
		MemoryLayoutValue: layout,
		DebugInfoValue:    debugInfo,
	}, nil
}

// remapDebugInfoAddresses adjusts all addresses in debug info to account for code relocation.
//
// When DWARF debug information is parsed from an ELF object file, the addresses are
// relative to the ELF file's virtual address layout (typically starting at 0x0 for
// relocatable objects). When the code is loaded into memory at a different base address
// (e.g., 0x10000), all debug info addresses must be adjusted by adding the code start address.
//
// This function creates a new DebugInfo with all address-containing fields remapped:
//   - InstructionLocations: Maps instruction addresses to source locations
//   - InstructionVariables: Maps instruction addresses to accessible variables
//   - Functions: Function start/end addresses and scope addresses
//
// Example: If DWARF says function main() starts at 0x100 and code is loaded at 0x10000,
// the remapped function will have StartAddress = 0x10100.
//
// Parameters:
//   - original: The debug info with addresses relative to the ELF file
//   - codeStart: The actual memory address where code is loaded (e.g., 0x10000)
//
// Returns a new DebugInfo with all addresses adjusted, or nil if original is nil.
func remapDebugInfoAddresses(original *DebugInfo, codeStart uint32) *DebugInfo {
	if original == nil {
		return nil
	}

	remapped := NewDebugInfo()
	remapped.CompilationUnit = original.CompilationUnit
	remapped.Producer = original.Producer
	remapped.SourceFiles = original.SourceFiles

	// Remap instruction locations
	for addr, loc := range original.InstructionLocations {
		remapped.InstructionLocations[codeStart+addr] = loc
	}

	// Remap instruction variables
	for addr, vars := range original.InstructionVariables {
		remapped.InstructionVariables[codeStart+addr] = vars
	}

	// Remap functions
	for name, fn := range original.Functions {
		remappedFn := &FunctionDebugInfo{
			Name:           fn.Name,
			StartAddress:   codeStart + fn.StartAddress,
			EndAddress:     codeStart + fn.EndAddress,
			SourceFile:     fn.SourceFile,
			StartLine:      fn.StartLine,
			EndLine:        fn.EndLine,
			Parameters:     fn.Parameters,
			LocalVariables: fn.LocalVariables,
		}
		// Remap scopes
		for _, scope := range fn.Scopes {
			remappedFn.Scopes = append(remappedFn.Scopes, ScopeInfo{
				StartAddress: codeStart + scope.StartAddress,
				EndAddress:   codeStart + scope.EndAddress,
				Variables:    scope.Variables,
			})
		}
		remapped.Functions[name] = remappedFn
	}

	return remapped
}

// alignAddress aligns an address to the given alignment
func alignAddress(addr, alignment uint32) uint32 {
	if alignment == 0 {
		return addr
	}
	remainder := addr % alignment
	if remainder == 0 {
		return addr
	}
	return addr + (alignment - remainder)
}

// GetSymbolAddress returns the resolved address for a symbol reference.
// Returns 0 and false if the symbol is not resolved or has no address.
func GetSymbolAddress(sym *SymbolReference) (uint32, bool) {
	if sym == nil {
		return 0, false
	}

	if sym.Function != nil {
		// Function address is derived from its first instruction
		// For now, we need to look it up from the program
		// This is a limitation - functions don't store their own address
		return 0, false
	}

	if sym.Global != nil && sym.Global.Address != nil {
		return *sym.Global.Address, true
	}

	if sym.Label != nil {
		// Label address needs to be looked up from the instruction
		// This is a limitation similar to functions
		return 0, false
	}

	return 0, false
}

// GetSymbolAddressFromProgram returns the resolved address for a symbol reference
// using the program's resolved data.
func GetSymbolAddressFromProgram(sym *SymbolReference, pf ProgramFile) (uint32, bool) {
	if sym == nil || pf == nil {
		return 0, false
	}

	lookupName := sym.BaseName()

	// Check functions
	if sym.Function != nil {
		functions := pf.Functions()
		if fn, ok := functions[lookupName]; ok {
			if len(fn.InstructionRanges) > 0 && fn.InstructionRanges[0].Count > 0 {
				instructions := pf.Instructions()
				startIdx := fn.InstructionRanges[0].Start
				if startIdx >= 0 && startIdx < len(instructions) && instructions[startIdx].Address != nil {
					return *instructions[startIdx].Address, true
				}
			}
		}
		return 0, false
	}

	// Check globals
	if sym.Global != nil && sym.Global.Address != nil {
		return *sym.Global.Address, true
	}

	// Check labels
	if sym.Label != nil {
		instructions := pf.Instructions()
		if sym.Label.InstructionIndex >= 0 && sym.Label.InstructionIndex < len(instructions) {
			inst := instructions[sym.Label.InstructionIndex]
			if inst.Address != nil {
				return *inst.Address, true
			}
		}
		return 0, false
	}

	return 0, false
}
