package mc

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
)

var (
	// ErrInstructionResolution indicates an error during instruction resolution
	ErrInstructionResolution = errors.New("instruction resolution error")
	// ErrInstructionParsing indicates an error parsing instruction text
	ErrInstructionParsing = errors.New("instruction parsing error")
)

// InstructionResolver resolves instructions in a ProgramFile, filling in
// Raw and Instruction fields from Text, or vice versa.
type InstructionResolver struct{}

// NewInstructionResolver creates a new instruction resolver
func NewInstructionResolver() *InstructionResolver {
	return &InstructionResolver{}
}

// ResolveInstructions processes all instructions in the program file and ensures each has:
// - Raw: the partially decoded instruction
// - Instruction: the fully decoded instruction ready for the interpreter
// - Text: the assembly text representation
//
// Returns a new ProgramFile with resolved instructions.
//
// The function will fail if:
// - Any symbol reference is not resolved (Function, Global, or Label pointer is nil)
// - Memory layout is not resolved (addresses are not assigned)
func (r *InstructionResolver) ResolveInstructions(pf ProgramFile) (ProgramFile, error) {
	// Check that memory layout is resolved
	if pf.MemoryLayout() == nil {
		return nil, fmt.Errorf("%w: memory layout must be resolved before instruction resolution", ErrUnresolvedMemory)
	}

	// Copy instructions for modification
	instructions := make([]Instruction, len(pf.Instructions()))
	copy(instructions, pf.Instructions())

	// Check that all instructions have addresses
	for i, instr := range instructions {
		if instr.Address == nil {
			return nil, fmt.Errorf("%w: instruction %d has no address assigned", ErrUnresolvedMemory, i)
		}
	}

	// Check that all globals have addresses
	for i, g := range pf.Globals() {
		if g.Address == nil {
			return nil, fmt.Errorf("%w: global %q (index %d) has no address assigned", ErrUnresolvedMemory, g.Name, i)
		}
	}

	// Resolve each instruction
	for i := range instructions {
		if err := r.resolveInstruction(&instructions[i], i); err != nil {
			return nil, err
		}
	}

	// Create a new ProgramFileContents with the resolved instructions
	return &ProgramFileContents{
		FileNameValue:     pf.FileName(),
		InstructionsValue: instructions,
		FunctionsValue:    pf.Functions(),
		GlobalsValue:      pf.Globals(),
		LabelsValue:       pf.Labels(),
		SourceFileValue:   pf.SourceFile(),
		MemoryLayoutValue: pf.MemoryLayout(),
		DebugInfoValue:    pf.DebugInfo(),
	}, nil
}

// resolveInstruction resolves a single instruction
func (r *InstructionResolver) resolveInstruction(instr *Instruction, index int) error {
	// First, check that all symbol references are resolved
	for _, sym := range instr.Symbols {
		if sym.Unresolved() {
			return fmt.Errorf("%w: instruction %d references unresolved symbol %q", ErrUnresolvedSymbol, index, sym.Name)
		}
	}

	hasText := instr.Text != ""
	hasRaw := instr.Raw != nil
	hasDecoded := instr.Instruction != nil

	// Case 1: Has text but no Raw or Instruction - parse from text
	if hasText && !hasRaw && !hasDecoded {
		raw, decoded, err := r.parseInstructionText(instr.Text, instr.Symbols)
		if err != nil {
			return fmt.Errorf("%w at instruction %d: %v", ErrInstructionParsing, index, err)
		}
		instr.Raw = raw
		instr.Instruction = decoded
		return nil
	}

	// Case 2: Has Raw but no text - generate text from Raw
	if hasRaw && !hasText {
		instr.Text = r.generateTextFromRaw(instr.Raw, instr.Symbols)
	}

	// Case 3: Has Raw but no Instruction - decode from Raw
	if hasRaw && !hasDecoded {
		decoded, err := instr.Raw.Decode()
		if err != nil {
			return fmt.Errorf("%w at instruction %d: failed to decode raw instruction: %v", ErrInstructionResolution, index, err)
		}
		instr.Instruction = decoded
	}

	// Case 4: Has Instruction but no Raw - generate Raw from Instruction
	if hasDecoded && !hasRaw {
		raw := instr.Instruction.Raw()
		instr.Raw = &raw
	}

	// Case 5: Has Instruction but no text - generate text from Instruction
	if hasDecoded && !hasText {
		instr.Text = r.generateTextFromInstruction(instr.Instruction, instr.Symbols)
	}

	// Verify we now have all three
	if instr.Text == "" || instr.Raw == nil || instr.Instruction == nil {
		return fmt.Errorf("%w at instruction %d: could not fully resolve instruction (text=%v, raw=%v, decoded=%v)",
			ErrInstructionResolution, index, instr.Text != "", instr.Raw != nil, instr.Instruction != nil)
	}

	return nil
}

// parseInstructionText parses assembly text into Raw and decoded Instruction
func (r *InstructionResolver) parseInstructionText(text string, symbols []SymbolReference) (*instructions.RawInstruction, *instructions.Instruction, error) {
	// Parse the instruction text
	// Format: MNEMONIC operand1, operand2, ...
	// Operands can be: registers (r0-r15, sp, lr), immediates (#value), or symbols (.name@lo/@hi)

	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil, fmt.Errorf("empty instruction text")
	}

	// Split into mnemonic and operands
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return nil, nil, fmt.Errorf("empty instruction text")
	}

	mnemonic := parts[0]

	// Look up the opcode
	opcode, err := instructions.Opcodes.ParseOpCode(mnemonic)
	if err != nil {
		return nil, nil, fmt.Errorf("unknown mnemonic %q: %w", mnemonic, err)
	}

	// Get the instruction descriptor
	descriptor, err := instructions.Instructions.Instruction(opcode)
	if err != nil {
		return nil, nil, fmt.Errorf("no descriptor for opcode %q: %w", mnemonic, err)
	}

	// Parse operands
	operandStr := ""
	if len(parts) > 1 {
		operandStr = strings.Join(parts[1:], " ")
	}

	operandValues, err := r.parseOperands(operandStr, descriptor, symbols)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse operands for %q: %w", mnemonic, err)
	}

	// Create the instruction
	decoded, err := instructions.NewInstruction(descriptor, operandValues)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create instruction %q: %w", mnemonic, err)
	}

	// Generate Raw from decoded
	raw := decoded.Raw()
	return &raw, decoded, nil
}

// parseOperands parses the operand string into operand values
func (r *InstructionResolver) parseOperands(operandStr string, descriptor *instructions.InstructionDescriptor, symbols []SymbolReference) ([]instructions.OperandValue, error) {
	if len(descriptor.Operands) == 0 {
		return []instructions.OperandValue{}, nil
	}

	// Count visible operands (those without LLVM_HideFromAsm)
	visibleCount := 0
	for _, op := range descriptor.Operands {
		if !op.LLVM_HideFromAsm {
			visibleCount++
		}
	}

	// Split operands by comma
	operandStrs := strings.Split(operandStr, ",")
	if len(operandStrs) != visibleCount {
		return nil, fmt.Errorf("expected %d operands, got %d", visibleCount, len(operandStrs))
	}

	operandValues := make([]instructions.OperandValue, len(descriptor.Operands))
	visibleIdx := 0
	for i, opDesc := range descriptor.Operands {
		if opDesc.LLVM_HideFromAsm {
			// Hidden operand - copy from the tied operand (previous operand with same type)
			// For MOVIMM16H, src is tied to dst, so copy dst value
			if i > 0 {
				operandValues[i] = operandValues[i-1]
			}
			continue
		}

		opStr := strings.TrimSpace(operandStrs[visibleIdx])
		value, err := r.parseOperandValue(opStr, opDesc, symbols)
		if err != nil {
			return nil, fmt.Errorf("operand %d: %w", i, err)
		}
		operandValues[i] = value
		visibleIdx++
	}

	return operandValues, nil
}

// parseOperandValue parses a single operand value
func (r *InstructionResolver) parseOperandValue(opStr string, opDesc *instructions.OperandDescriptor, symbols []SymbolReference) (instructions.OperandValue, error) {
	opStr = strings.TrimSpace(opStr)

	// Check if it's a symbol reference
	if r.isSymbolReference(opStr) {
		return r.resolveSymbolOperand(opStr, opDesc, symbols)
	}

	// Check if it's an immediate
	if strings.HasPrefix(opStr, "#") {
		immStr := strings.TrimPrefix(opStr, "#")
		return opDesc.ParseValue(immStr)
	}

	// Otherwise it's a register or direct value
	return opDesc.ParseValue(opStr)
}

// isSymbolReference checks if the operand string is a symbol reference
func (r *InstructionResolver) isSymbolReference(opStr string) bool {
	// Symbol references start with . or are followed by @lo/@hi
	return strings.HasPrefix(opStr, ".") || strings.Contains(opStr, "@")
}

// resolveSymbolOperand resolves a symbol reference to an immediate value
func (r *InstructionResolver) resolveSymbolOperand(opStr string, opDesc *instructions.OperandDescriptor, symbols []SymbolReference) (instructions.OperandValue, error) {
	// Extract base name and usage
	baseName := opStr
	var usage SymbolReferenceUsage = SymbolUsageFull

	if strings.HasSuffix(opStr, "@lo") {
		baseName = strings.TrimSuffix(opStr, "@lo")
		usage = SymbolUsageLo
	} else if strings.HasSuffix(opStr, "@hi") {
		baseName = strings.TrimSuffix(opStr, "@hi")
		usage = SymbolUsageHi
	}

	// Find the symbol in the references
	var addr uint32
	found := false
	for _, sym := range symbols {
		if sym.Name == baseName && sym.Usage == usage {
			// Get the address from the resolved symbol
			addr, found = r.getSymbolAddress(&sym)
			break
		}
	}

	if !found {
		return instructions.OperandValue{}, fmt.Errorf("symbol %q not found or not resolved", opStr)
	}

	// Apply @lo/@hi masking
	var value int32
	switch usage {
	case SymbolUsageLo:
		value = int32(addr & 0xFFFF)
	case SymbolUsageHi:
		value = int32((addr >> 16) & 0xFFFF)
	default:
		value = int32(addr)
	}

	return opDesc.ParseValue(fmt.Sprintf("%d", value))
}

// getSymbolAddress returns the address of a resolved symbol
func (r *InstructionResolver) getSymbolAddress(sym *SymbolReference) (uint32, bool) {
	if sym.Function != nil {
		// For functions, we'd need the address of the first instruction
		// This requires looking up the function's instruction range
		return 0, false // Not implemented yet - would need ProgramFile context
	}

	if sym.Global != nil && sym.Global.Address != nil {
		return *sym.Global.Address, true
	}

	if sym.Label != nil {
		// For labels, we'd need to look up the instruction address
		// This requires access to the instructions array
		return 0, false // Not implemented yet - would need ProgramFile context
	}

	return 0, false
}

// generateTextFromRaw generates assembly text from a RawInstruction
func (r *InstructionResolver) generateTextFromRaw(raw *instructions.RawInstruction, symbols []SymbolReference) string {
	// Decode to get the full instruction, then generate text
	decoded, err := raw.Decode()
	if err != nil {
		return raw.String()
	}
	return r.generateTextFromInstruction(decoded, symbols)
}

// generateTextFromInstruction generates assembly text from a decoded Instruction
func (r *InstructionResolver) generateTextFromInstruction(instr *instructions.Instruction, symbols []SymbolReference) string {
	var sb strings.Builder

	sb.WriteString(instr.Descriptor.OpCode.Mnemonic)

	// Track which symbols have been used
	usedSymbols := make(map[int]bool)

	// Track if we've written any operands (for comma placement)
	firstOperand := true

	for i, operand := range instr.OperandValues {
		// Skip operands marked as hidden from assembly (e.g., MOVIMM16H's tied src register)
		if i < len(instr.Descriptor.Operands) && instr.Descriptor.Operands[i].LLVM_HideFromAsm {
			continue
		}

		if !firstOperand {
			sb.WriteString(",")
		}
		sb.WriteString(" ")
		firstOperand = false

		// Only look for symbol references for immediate operands
		symbolFound := false
		if operand.Kind() == instructions.OperandKind_Immediate {
			// Check if this operand corresponds to a symbol reference
			for symIdx, sym := range symbols {
				if usedSymbols[symIdx] {
					continue // Skip already used symbols
				}
				if sym.Usage == SymbolUsageLo || sym.Usage == SymbolUsageHi {
					suffix := ""
					if sym.Usage == SymbolUsageLo {
						suffix = "@lo"
					} else if sym.Usage == SymbolUsageHi {
						suffix = "@hi"
					}
					sb.WriteString(sym.Name)
					sb.WriteString(suffix)
					symbolFound = true
					usedSymbols[symIdx] = true
					break
				}
			}
		}

		if !symbolFound {
			// Use the operand's string representation
			if operand.Kind() == instructions.OperandKind_Immediate {
				sb.WriteString("#")
			}
			sb.WriteString(operand.String())
		}
	}

	return sb.String()
}

// ResolveWithContext resolves instructions with full program context
// This version can resolve function and label addresses.
// Returns a new ProgramFile with resolved instructions.
func (r *InstructionResolver) ResolveWithContext(pf ProgramFile) (ProgramFile, error) {
	// Check that memory layout is resolved
	if pf.MemoryLayout() == nil {
		return nil, fmt.Errorf("%w: memory layout must be resolved before instruction resolution", ErrUnresolvedMemory)
	}

	// Copy instructions for modification
	instructionsCopy := make([]Instruction, len(pf.Instructions()))
	copy(instructionsCopy, pf.Instructions())

	// Check that all instructions have addresses
	for i, instr := range instructionsCopy {
		if instr.Address == nil {
			return nil, fmt.Errorf("%w: instruction %d has no address assigned", ErrUnresolvedMemory, i)
		}
	}

	// Check that all globals have addresses
	for i, g := range pf.Globals() {
		if g.Address == nil {
			return nil, fmt.Errorf("%w: global %q (index %d) has no address assigned", ErrUnresolvedMemory, g.Name, i)
		}
	}

	// Create intermediate result for context-based resolution
	result := &ProgramFileContents{
		FileNameValue:     pf.FileName(),
		InstructionsValue: instructionsCopy,
		FunctionsValue:    pf.Functions(),
		GlobalsValue:      pf.Globals(),
		LabelsValue:       pf.Labels(),
		SourceFileValue:   pf.SourceFile(),
		MemoryLayoutValue: pf.MemoryLayout(),
		DebugInfoValue:    pf.DebugInfo(),
	}

	// Resolve each instruction with context
	for i := range result.InstructionsValue {
		if err := r.resolveInstructionWithContext(&result.InstructionsValue[i], i, result); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// resolveInstructionWithContext resolves a single instruction with full program context
func (r *InstructionResolver) resolveInstructionWithContext(instr *Instruction, index int, pf *ProgramFileContents) error {
	// First, check that all symbol references are resolved
	for _, sym := range instr.Symbols {
		if sym.Unresolved() {
			return fmt.Errorf("%w: instruction %d references unresolved symbol %q", ErrUnresolvedSymbol, index, sym.Name)
		}
	}

	hasText := instr.Text != ""
	hasRaw := instr.Raw != nil && instr.Raw.Descriptor != nil // Only consider Raw valid if it has a descriptor
	hasDecoded := instr.Instruction != nil

	// If we have a Raw with nil Descriptor, this is an unknown instruction - skip resolution
	if instr.Raw != nil && instr.Raw.Descriptor == nil {
		// Unknown instruction - nothing to resolve, leave as-is
		return nil
	}

	// Case 1: Has text but no Raw or Instruction - parse from text with context
	if hasText && !hasRaw && !hasDecoded {
		raw, decoded, err := r.parseInstructionTextWithContext(instr.Text, instr.Symbols, pf)
		if err != nil {
			return fmt.Errorf("%w at instruction %d: %v", ErrInstructionParsing, index, err)
		}
		instr.Raw = raw
		instr.Instruction = decoded
		return nil
	}

	// Case 2: Has Raw but no text - generate text from Raw
	if hasRaw && !hasText {
		instr.Text = r.generateTextFromRaw(instr.Raw, instr.Symbols)
	}

	// Case 3: Has Raw but no Instruction - decode from Raw
	if hasRaw && !hasDecoded {
		decoded, err := instr.Raw.Decode()
		if err != nil {
			return fmt.Errorf("%w at instruction %d: failed to decode raw instruction: %v", ErrInstructionResolution, index, err)
		}
		instr.Instruction = decoded
	}

	// Case 4: Has Instruction but no Raw - generate Raw from Instruction
	if hasDecoded && !hasRaw {
		raw := instr.Instruction.Raw()
		instr.Raw = &raw
	}

	// Case 5: Has Instruction but no text - generate text from Instruction
	if hasDecoded && !hasText {
		instr.Text = r.generateTextFromInstruction(instr.Instruction, instr.Symbols)
	}

	// Case 6: If we have symbols that need to be patched into the Raw instruction
	// This is needed when loading from binary files where the immediate values
	// are unresolved but we have symbol references attached
	if len(instr.Symbols) > 0 && instr.Raw != nil && instr.Raw.Descriptor != nil {
		if err := r.patchSymbolsIntoInstruction(instr, pf); err != nil {
			return fmt.Errorf("%w at instruction %d: %v", ErrInstructionResolution, index, err)
		}
	}

	// Verify we now have all three
	if instr.Text == "" || instr.Raw == nil || instr.Instruction == nil {
		return fmt.Errorf("%w at instruction %d: could not fully resolve instruction (text=%v, raw=%v, decoded=%v)",
			ErrInstructionResolution, index, instr.Text != "", instr.Raw != nil, instr.Instruction != nil)
	}

	return nil
}

// parseInstructionTextWithContext parses assembly text with full program context
func (r *InstructionResolver) parseInstructionTextWithContext(text string, symbols []SymbolReference, pf *ProgramFileContents) (*instructions.RawInstruction, *instructions.Instruction, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil, fmt.Errorf("empty instruction text")
	}

	// Split into mnemonic and operands
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return nil, nil, fmt.Errorf("empty instruction text")
	}

	mnemonic := parts[0]

	// Look up the opcode
	opcode, err := instructions.Opcodes.ParseOpCode(mnemonic)
	if err != nil {
		return nil, nil, fmt.Errorf("unknown mnemonic %q: %w", mnemonic, err)
	}

	// Get the instruction descriptor
	descriptor, err := instructions.Instructions.Instruction(opcode)
	if err != nil {
		return nil, nil, fmt.Errorf("no descriptor for opcode %q: %w", mnemonic, err)
	}

	// Parse operands with context
	operandStr := ""
	if len(parts) > 1 {
		operandStr = strings.Join(parts[1:], " ")
	}

	operandValues, err := r.parseOperandsWithContext(operandStr, descriptor, symbols, pf)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse operands for %q: %w", mnemonic, err)
	}

	// Create the instruction
	decoded, err := instructions.NewInstruction(descriptor, operandValues)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create instruction %q: %w", mnemonic, err)
	}

	// Generate Raw from decoded
	raw := decoded.Raw()
	return &raw, decoded, nil
}

// parseOperandsWithContext parses operands with full program context
func (r *InstructionResolver) parseOperandsWithContext(operandStr string, descriptor *instructions.InstructionDescriptor, symbols []SymbolReference, pf *ProgramFileContents) ([]instructions.OperandValue, error) {
	if len(descriptor.Operands) == 0 {
		return []instructions.OperandValue{}, nil
	}

	// Count visible operands (those without LLVM_HideFromAsm)
	visibleCount := 0
	for _, op := range descriptor.Operands {
		if !op.LLVM_HideFromAsm {
			visibleCount++
		}
	}

	// Split operands by comma
	operandStrs := strings.Split(operandStr, ",")
	if len(operandStrs) != visibleCount {
		return nil, fmt.Errorf("expected %d operands, got %d", visibleCount, len(operandStrs))
	}

	operandValues := make([]instructions.OperandValue, len(descriptor.Operands))
	visibleIdx := 0
	for i, opDesc := range descriptor.Operands {
		if opDesc.LLVM_HideFromAsm {
			// Hidden operand - copy from the tied operand (previous operand with same type)
			// For MOVIMM16H, src is tied to dst, so copy dst value
			if i > 0 {
				operandValues[i] = operandValues[i-1]
			}
			continue
		}

		opStr := strings.TrimSpace(operandStrs[visibleIdx])
		value, err := r.parseOperandValueWithContext(opStr, opDesc, symbols, pf)
		if err != nil {
			return nil, fmt.Errorf("operand %d: %w", i, err)
		}
		operandValues[i] = value
		visibleIdx++
	}

	return operandValues, nil
}

// parseOperandValueWithContext parses a single operand value with program context
func (r *InstructionResolver) parseOperandValueWithContext(opStr string, opDesc *instructions.OperandDescriptor, symbols []SymbolReference, pf *ProgramFileContents) (instructions.OperandValue, error) {
	opStr = strings.TrimSpace(opStr)

	// Check if it's a symbol reference
	if r.isSymbolReference(opStr) {
		return r.resolveSymbolOperandWithContext(opStr, opDesc, symbols, pf)
	}

	// Check if it's an immediate
	if strings.HasPrefix(opStr, "#") {
		immStr := strings.TrimPrefix(opStr, "#")
		return opDesc.ParseValue(immStr)
	}

	// Otherwise it's a register or direct value
	return opDesc.ParseValue(opStr)
}

// resolveSymbolOperandWithContext resolves a symbol reference with full program context
func (r *InstructionResolver) resolveSymbolOperandWithContext(opStr string, opDesc *instructions.OperandDescriptor, symbols []SymbolReference, pf *ProgramFileContents) (instructions.OperandValue, error) {
	// Extract base name and usage
	baseName := opStr
	var usage SymbolReferenceUsage = SymbolUsageFull

	if strings.HasSuffix(opStr, "@lo") {
		baseName = strings.TrimSuffix(opStr, "@lo")
		usage = SymbolUsageLo
	} else if strings.HasSuffix(opStr, "@hi") {
		baseName = strings.TrimSuffix(opStr, "@hi")
		usage = SymbolUsageHi
	}

	// Find the symbol in the references
	var addr uint32
	found := false
	for _, sym := range symbols {
		if sym.Name == baseName && sym.Usage == usage {
			addr, found = r.getSymbolAddressWithContext(&sym, pf)
			break
		}
	}

	if !found {
		return instructions.OperandValue{}, fmt.Errorf("symbol %q not found or not resolved", opStr)
	}

	// Apply @lo/@hi masking
	var value int32
	switch usage {
	case SymbolUsageLo:
		value = int32(addr & 0xFFFF)
	case SymbolUsageHi:
		value = int32((addr >> 16) & 0xFFFF)
	default:
		value = int32(addr)
	}

	return opDesc.ParseValue(fmt.Sprintf("%d", value))
}

// getSymbolAddressWithContext returns the address of a resolved symbol using program context
func (r *InstructionResolver) getSymbolAddressWithContext(sym *SymbolReference, pf *ProgramFileContents) (uint32, bool) {
	if sym.Function != nil {
		// Get the address of the first instruction in the function
		if len(sym.Function.InstructionRanges) > 0 {
			firstInstrIdx := sym.Function.InstructionRanges[0].Start
			if firstInstrIdx < len(pf.InstructionsValue) {
				instr := pf.InstructionsValue[firstInstrIdx]
				if instr.Address != nil {
					return *instr.Address, true
				}
			}
		}
		return 0, false
	}

	if sym.Global != nil && sym.Global.Address != nil {
		return *sym.Global.Address, true
	}

	if sym.Label != nil {
		// Get the address of the instruction at the label's index
		if sym.Label.InstructionIndex >= 0 && sym.Label.InstructionIndex < len(pf.InstructionsValue) {
			instr := pf.InstructionsValue[sym.Label.InstructionIndex]
			if instr.Address != nil {
				return *instr.Address, true
			}
		}
		return 0, false
	}

	return 0, false
}

// patchSymbolsIntoInstruction patches symbol addresses into the instruction's immediate operands
// This is needed for binary files where instructions have unresolved immediates
func (r *InstructionResolver) patchSymbolsIntoInstruction(instr *Instruction, pf *ProgramFileContents) error {
	if instr.Raw == nil || instr.Raw.Descriptor == nil {
		return nil
	}

	// For MOVIMM16L/MOVIMM16H instructions, the first operand is the immediate
	// We need to patch it with the symbol's address (lo or hi part)
	for _, sym := range instr.Symbols {
		addr, ok := r.getSymbolAddressWithContext(&sym, pf)
		if !ok {
			return fmt.Errorf("failed to get address for symbol %q", sym.Name)
		}

		// Calculate the value based on usage
		var value uint32
		switch sym.Usage {
		case SymbolUsageLo:
			value = addr & 0xFFFF
		case SymbolUsageHi:
			value = (addr >> 16) & 0xFFFF
		default:
			value = addr
		}

		// Find the immediate operand and patch it
		// For MOVIMM16L/MOVIMM16H, the first operand is the 16-bit immediate
		for i, opDesc := range instr.Raw.Descriptor.Operands {
			if opDesc.Kind == instructions.OperandKind_Immediate {
				if i < len(instr.Raw.OperandValues) {
					instr.Raw.OperandValues[i] = uint64(value)
				}
				break
			}
		}

		// Re-decode the instruction from the patched Raw
		decoded, err := instr.Raw.Decode()
		if err != nil {
			return fmt.Errorf("failed to re-decode instruction after patching: %w", err)
		}
		instr.Instruction = decoded

		// Update the text representation
		instr.Text = r.generateTextFromInstruction(decoded, instr.Symbols)
	}

	return nil
}
