package program

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/utils/logging"
)

// ResolveSymbols resolves all symbol references in the given ProgramFile.
// It returns a new ProgramFile with all symbols resolved, or an error if any
// symbol could not be resolved.
func ResolveSymbols(pf ProgramFile) (ProgramFile, error) {
	log := log().Child(pf.FileName()).Child("ResolveSymbols")

	// Build lookup maps for symbols
	functions := pf.Functions()
	globals := pf.Globals()
	labels := pf.Labels()

	// Create maps for quick lookup
	globalMap := make(map[string]*Global)
	for i := range globals {
		globalMap[globals[i].Name] = &globals[i]
	}

	labelMap := make(map[string]*Label)
	for i := range labels {
		labelMap[labels[i].Name] = &labels[i]
	}

	// Copy functions (they need to be addressable for references)
	resolvedFunctions := make(map[string]Function, len(functions))
	functionPtrs := make(map[string]*Function)
	for name, fn := range functions {
		resolvedFunctions[name] = fn
		fnCopy := resolvedFunctions[name]
		functionPtrs[name] = &fnCopy
	}

	// Copy globals
	resolvedGlobals := make([]Global, len(globals))
	copy(resolvedGlobals, globals)
	resolvedGlobalPtrs := make(map[string]*Global)
	for i := range resolvedGlobals {
		resolvedGlobalPtrs[resolvedGlobals[i].Name] = &resolvedGlobals[i]
	}

	// Copy labels
	resolvedLabels := make([]Label, len(labels))
	copy(resolvedLabels, labels)
	resolvedLabelPtrs := make(map[string]*Label)
	for i := range resolvedLabels {
		resolvedLabelPtrs[resolvedLabels[i].Name] = &resolvedLabels[i]
	}

	// Resolve symbols in instructions
	srcInstructions := pf.Instructions()
	resolvedInstructions := make([]Instruction, len(srcInstructions))

	var unresolvedSymbols []string

	for i, inst := range srcInstructions {
		resolvedInstructions[i] = Instruction{
			LineNumber:  inst.LineNumber,
			Address:     inst.Address,
			Text:        inst.Text,
			Raw:         inst.Raw,
			Instruction: inst.Instruction,
			Symbols:     make([]SymbolReference, len(inst.Symbols)),
		}

		for j, sym := range inst.Symbols {
			resolved := SymbolReference{
				Name:  sym.Name,
				Usage: sym.Usage,
			}

			// Use BaseName() for lookup (strips @lo/@hi suffixes)
			lookupName := sym.BaseName()

			// Try function first
			if fn, ok := functionPtrs[lookupName]; ok {
				resolved.Function = fn
				log.Debug("resolved symbol as function", slog.String("symbol", sym.Name), slog.String("function", fn.Name), logging.Address("instruction_address", *inst.Address), slog.String("instruction", fmt.Sprintf("{%s}", inst.Text)))
			} else if g, ok := resolvedGlobalPtrs[lookupName]; ok {
				resolved.Global = g
				log.Debug("resolved symbol as global", slog.String("symbol", sym.Name), slog.String("global", g.Name), logging.Address("instruction_address", *inst.Address), slog.String("instruction", fmt.Sprintf("{%s}", inst.Text)))
			} else if lbl, ok := resolvedLabelPtrs[lookupName]; ok {
				resolved.Label = lbl
				log.Debug("resolved symbol as label", slog.String("symbol", sym.Name), slog.String("label", lbl.Name), logging.Address("instruction_address", *inst.Address), slog.String("instruction", fmt.Sprintf("{%s}", inst.Text)))
			} else {
				log.Debug("failed to resolve symbol", slog.String("symbol", sym.Name), logging.Address("instruction_address", *inst.Address), slog.String("instruction", fmt.Sprintf("{%s}", inst.Text)))
				unresolvedSymbols = append(unresolvedSymbols, fmt.Sprintf("%s (instruction %d, line %d)", sym.Name, i, inst.LineNumber))
			}

			resolvedInstructions[i].Symbols[j] = resolved
		}
	}

	if len(unresolvedSymbols) > 0 {
		return nil, log.Errorf("unresolved symbols: %s", strings.Join(unresolvedSymbols, ", "))
	}

	return &ProgramFileContents{
		FileNameValue:     pf.FileName(),
		SourceFileValue:   pf.SourceFile(),
		FunctionsValue:    resolvedFunctions,
		InstructionsValue: resolvedInstructions,
		GlobalsValue:      resolvedGlobals,
		LabelsValue:       resolvedLabels,
		DebugInfoValue:    pf.DebugInfo(),
	}, nil
}
