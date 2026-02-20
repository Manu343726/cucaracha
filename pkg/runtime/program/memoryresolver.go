package program

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/Manu343726/cucaracha/pkg/hw/memory"
	"github.com/Manu343726/cucaracha/pkg/utils/logging"
)

var (
	ErrProgramTooLarge        = errors.New("program exceeds maximum size")
	ErrUnresolvedSymbol       = errors.New("unresolved symbol reference")
	ErrUnresolvedMemory       = errors.New("unresolved memory address")
	ErrInvalidInstructionSize = errors.New("instruction size must be greater than 0")
)

// ResolveMemory assigns memory addresses to all instructions and globals in a ProgramFile.
// It returns a new ProgramFile with all addresses resolved
func ResolveMemory(pf ProgramFile, memoryLayout *memory.MemoryLayout) (ProgramFile, error) {
	log := log().Child(pf.FileName()).Child("ResolveMemory")

	srcInstructions := pf.Instructions()
	srcGlobals := pf.Globals()
	srcFunctions := pf.Functions()
	srcLabels := pf.Labels()

	// Calculate code size
	codeSize := uint32(len(srcInstructions)) * 4
	codeStart := memoryLayout.CodeBase

	if codeSize >= memoryLayout.CodeSize {
		return nil, log.Errorf("%w: code size %d exceeds allocated code section size %d", ErrProgramTooLarge, codeSize, memoryLayout.CodeSize)
	}

	// Calculate data size
	var dataSize uint32 = 0
	for _, g := range srcGlobals {
		dataSize += uint32(g.Size)
	}

	if memoryLayout.DataBase+dataSize > memoryLayout.Data().End() {
		return nil, log.Errorf("%w: data size %d exceeds allocated data section size %d", ErrProgramTooLarge, dataSize, memoryLayout.Data().Size)
	}

	log.Debug("relocating instructions...", logging.Address("code_start", codeStart), slog.Uint64("code_size", uint64(codeSize)))

	// Create resolved instructions with addresses
	resolvedInstructions := make([]Instruction, len(srcInstructions))
	for i, inst := range srcInstructions {
		addr := codeStart + uint32(i)*4
		resolvedInstructions[i] = Instruction{
			LineNumber:  inst.LineNumber,
			Address:     &addr,
			Text:        inst.Text,
			Raw:         inst.Raw,
			Instruction: inst.Instruction,
			Symbols:     make([]SymbolReference, len(inst.Symbols)),
		}

		log.Debug("instruction relocated", slog.Int("index", i), logging.Address("address", addr), slog.String("instruction", fmt.Sprintf("{%s}", inst.Text)))

		// Copy symbols initially (will update references later)
		copy(resolvedInstructions[i].Symbols, inst.Symbols)
	}

	log.Debug("relocating globals...", logging.Address("data_start", memoryLayout.DataBase), slog.Uint64("data_size", uint64(dataSize)))

	// Create resolved globals with addresses
	resolvedGlobals := make([]Global, len(srcGlobals))
	currentDataAddr := memoryLayout.DataBase
	for i, g := range srcGlobals {
		addr := currentDataAddr
		resolvedGlobals[i] = Global{
			Name:        g.Name,
			Address:     &addr,
			Size:        g.Size,
			InitialData: g.InitialData,
			Type:        g.Type,
		}

		log.Debug("global relocated", slog.String("global", g.Name), logging.Address("address", addr), slog.Uint64("size", uint64(g.Size)))

		currentDataAddr += uint32(g.Size)
	}

	// Create resolved functions (copy, addresses derived from instruction addresses)
	resolvedFunctions := make(map[string]Function, len(srcFunctions))
	for name, fn := range srcFunctions {
		resolvedFunctions[name] = fn
	}

	// Create resolved labels with addresses (derived from instruction addresses)
	resolvedLabels := make([]Label, len(srcLabels))
	for i, lbl := range srcLabels {
		resolvedLabels[i] = Label{
			Name:             lbl.Name,
			InstructionIndex: lbl.InstructionIndex,
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

			if sym.Unresolved() {
				return nil, log.Errorf("%w: symbol '%q' in instruction {%s} at address 0x%X is not resolved", ErrUnresolvedSymbol, sym.Name, resolvedInstructions[i].Instruction.String(), *resolvedInstructions[i].Address)
			}

			if sym.Function != nil {
				if fn, ok := functionPtrMap[lookupName]; ok {
					// Functions don't have their own address, so we derive it from the first instruction in the function's instruction range
					// So technically we resolved the function symbol address before when relocating instructions.
					// Here we just set the resolved reference to point to the resolved function (a copy from the source function) so that we don't have references to the original source function objects.
					resolved.Function = fn
					log.Debug("function reference updated", slog.String("function", fn.Name), logging.Address("instruction_address", *resolvedInstructions[i].Address), slog.String("instruction", fmt.Sprintf("{%s}", resolvedInstructions[i].Instruction.String())))
				} else {
					return nil, log.Errorf("%w: function symbol '%q' in instruction {%s} at address 0x%X could not be resolved", ErrUnresolvedSymbol, sym.Name, resolvedInstructions[i].Instruction.String(), *resolvedInstructions[i].Address)
				}
			} else if sym.Global != nil {
				if g, ok := globalPtrs[lookupName]; ok {
					resolved.Global = g
					log.Debug("global reference updated", slog.String("global", g.Name), logging.Address("global_address", *g.Address), logging.Address("instruction_address", *resolvedInstructions[i].Address), slog.String("instruction", fmt.Sprintf("{%s}", resolvedInstructions[i].Instruction.String())))
				} else {
					return nil, log.Errorf("%w: global symbol '%q' in instruction {%s} at address 0x%X could not be resolved", ErrUnresolvedSymbol, sym.Name, resolvedInstructions[i].Instruction.String(), *resolvedInstructions[i].Address)
				}
			} else if sym.Label != nil {
				if lbl, ok := labelPtrs[lookupName]; ok {
					// Labels also don't have their own address, so we derive it from the instruction they reference (the instruction index is stored in the label)
					// Similar to functions, we just update the reference to point to the resolved label object to avoid references to the original source label objects.
					resolved.Label = lbl
					log.Debug("label reference updated", slog.String("label", lbl.Name), logging.Address("instruction_address", *resolvedInstructions[i].Address), slog.String("instruction", fmt.Sprintf("{%s}", resolvedInstructions[i].Instruction.String())))
				} else {
					return nil, log.Errorf("%w: label symbol '%q' in instruction {%s} at address 0x%X could not be resolved", ErrUnresolvedSymbol, sym.Name, resolvedInstructions[i].Instruction.String(), *resolvedInstructions[i].Address)
				}
			} else {
				// This should not happen - we should have caught unresolved symbols with the previous check
				panic(fmt.Errorf("symbol '%q' in instruction {%s} at address 0x%X is not resolved", sym.Name, resolvedInstructions[i].Instruction.String(), *resolvedInstructions[i].Address))
			}

			resolvedInstructions[i].Symbols[j] = resolved
		}
	}

	// Remap debug info addresses to match the relocated code
	debugInfo := pf.DebugInfo()
	if debugInfo != nil {
		debugInfo = relocateDebugInfoAddresses(debugInfo, codeStart)
	}

	// Create a new MemoryLayout with the actual used sizes
	resolvedLayout := &memory.MemoryLayout{
		TotalSize:               memoryLayout.TotalSize,
		SystemDescriptorBase:    memoryLayout.SystemDescriptorBase,
		SystemDescriptorSize:    memoryLayout.SystemDescriptorSize,
		VectorTableBase:         memoryLayout.VectorTableBase,
		VectorTableSize:         memoryLayout.VectorTableSize,
		DataBase:                memoryLayout.DataBase,
		DataSize:                dataSize,
		CodeBase:                memoryLayout.CodeBase,
		CodeSize:                codeSize,
		HeapBase:                memoryLayout.HeapBase,
		HeapSize:                memoryLayout.HeapSize,
		StackBase:               memoryLayout.StackBase,
		StackSize:               memoryLayout.StackSize,
		PeripheralBase:          memoryLayout.PeripheralBase,
		PeripheralSize:          memoryLayout.PeripheralSize,
		PeripheralBaseAddresses: memoryLayout.PeripheralBaseAddresses,
	}

	log.Debug("final program memory layout", slog.Uint64("TotalSize", uint64(resolvedLayout.TotalSize)))
	log.Debug("final program memory layout", logging.Address("SystemDescriptorBase", resolvedLayout.SystemDescriptorBase))
	log.Debug("final program memory layout", slog.Uint64("SystemDescriptorSize", uint64(resolvedLayout.SystemDescriptorSize)))
	log.Debug("final program memory layout", logging.Address("VectorTableBase", resolvedLayout.VectorTableBase))
	log.Debug("final program memory layout", slog.Uint64("VectorTableSize", uint64(resolvedLayout.VectorTableSize)))
	log.Debug("final program memory layout", logging.Address("DataBase", resolvedLayout.DataBase))
	log.Debug("final program memory layout", slog.Uint64("DataSize", uint64(resolvedLayout.DataSize)))
	log.Debug("final program memory layout", logging.Address("CodeBase", resolvedLayout.CodeBase))
	log.Debug("final program memory layout", slog.Uint64("CodeSize", uint64(resolvedLayout.CodeSize)))
	log.Debug("final program memory layout", logging.Address("HeapBase", resolvedLayout.HeapBase))
	log.Debug("final program memory layout", slog.Uint64("HeapSize", uint64(resolvedLayout.HeapSize)))
	log.Debug("final program memory layout", logging.Address("StackBase", resolvedLayout.StackBase))
	log.Debug("final program memory layout", slog.Uint64("StackSize", uint64(resolvedLayout.StackSize)))
	log.Debug("final program memory layout", logging.Address("PeripheralBase", resolvedLayout.PeripheralBase))
	log.Debug("final program memory layout", slog.Uint64("PeripheralSize", uint64(resolvedLayout.PeripheralSize)))
	for i, addr := range resolvedLayout.PeripheralBaseAddresses {
		log.Debug("final program memory layout", logging.Address("base_address", addr), slog.Int("peripheral_index", i))
	}

	return &ProgramFileContents{
		FileNameValue:     pf.FileName(),
		SourceFileValue:   pf.SourceFile(),
		FunctionsValue:    resolvedFunctions,
		InstructionsValue: resolvedInstructions,
		GlobalsValue:      resolvedGlobals,
		LabelsValue:       resolvedLabels,
		MemoryLayoutValue: resolvedLayout,
		DebugInfoValue:    debugInfo,
	}, nil
}

// relocateDebugInfoAddresses adjusts all addresses in debug info to account for code relocation.
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
func relocateDebugInfoAddresses(original *DebugInfo, codeStart uint32) *DebugInfo {
	if original == nil {
		return nil
	}

	log := log().Child("relocateDebugInfoAddresses")

	relocated := NewDebugInfo()
	relocated.CompilationUnit = original.CompilationUnit
	relocated.Producer = original.Producer
	relocated.SourceLibrary = original.SourceLibrary

	// Relocate instruction locations
	for addr, loc := range original.InstructionLocations {
		relocated.InstructionLocations[codeStart+addr] = loc
		log.Debug("instruction location relocated", logging.Address("original_address", addr), logging.Address("relocated_address", codeStart+addr), slog.String("source_location", loc.String()))
	}

	// Relocate instruction variables
	for addr, vars := range original.InstructionVariables {
		relocated.InstructionVariables[codeStart+addr] = vars

		for _, v := range vars {
			log.Debug("instruction variable relocated", logging.Address("original_address", addr), logging.Address("relocated_address", codeStart+addr), slog.String("variable", v.Name), slog.String("type", v.TypeName), slog.Bool("is_parameter", v.IsParameter), slog.Int("size", v.Size))
		}
	}

	// Relocate functions
	for name, fn := range original.Functions {
		relocatedFn := &FunctionDebugInfo{
			Name:           fn.Name,
			StartAddress:   codeStart + fn.StartAddress,
			EndAddress:     codeStart + fn.EndAddress,
			SourceFile:     fn.SourceFile,
			StartLine:      fn.StartLine,
			EndLine:        fn.EndLine,
			Parameters:     fn.Parameters,
			LocalVariables: fn.LocalVariables,
		}

		log.Debug("function debug info relocated", slog.String("function", fn.Name), logging.Address("original_start_address", fn.StartAddress), logging.Address("relocated_start_address", relocatedFn.StartAddress), logging.Address("original_end_address", fn.EndAddress), logging.Address("relocated_end_address", relocatedFn.EndAddress), slog.String("source_file", fn.SourceFile), slog.Int("start_line", fn.StartLine), slog.Int("end_line", fn.EndLine))

		// Relocate scopes
		for _, scope := range fn.Scopes {
			relocatedFn.Scopes = append(relocatedFn.Scopes, ScopeInfo{
				StartAddress: codeStart + scope.StartAddress,
				EndAddress:   codeStart + scope.EndAddress,
				Variables:    scope.Variables,
			})

			log.Debug("function scope relocated", slog.String("function", fn.Name), logging.Address("original_scope_start_address", scope.StartAddress), logging.Address("relocated_scope_start_address", codeStart+scope.StartAddress), logging.Address("original_scope_end_address", scope.EndAddress), logging.Address("relocated_scope_end_address", codeStart+scope.EndAddress), slog.Int("variable_count", len(scope.Variables)))
		}
		relocated.Functions[name] = relocatedFn
	}

	return relocated
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
