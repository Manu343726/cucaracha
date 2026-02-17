package program

import (
	"fmt"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/Manu343726/cucaracha/pkg/hw/memory"
	"github.com/Manu343726/cucaracha/pkg/runtime/program/sourcecode"
)

// Returns the function starting at the given memory address
//
// Note that this function requires the ProgramFile to have resolved memory addresses.
// Also the result only will be correct if the program file was resolved with the same
// memory layout as the runtime where the address comes from.
func FunctionAtAddress(pf ProgramFile, addr uint32) (*Function, error) {
	instruction, err := InstructionAtAddress(pf, addr)
	if err != nil {
		return nil, err
	}

	instructionIndex := (*instruction.Address - pf.MemoryLayout().CodeBase) / 4

	functions := pf.Functions()

	for _, fn := range functions {
		if fn.FirstInstructionIndex() == int(instructionIndex) {
			return &fn, nil
		}
	}

	return nil, nil
}

// Returns the global variable at the given memory address
//
// Note that this function requires the ProgramFile to have resolved memory addresses.
// Also the result only will be correct if the program file was resolved with the same
// memory layout as the runtime where the address comes from.
func GlobalAtAddress(pf ProgramFile, addr uint32) (*Global, error) {
	if pf.MemoryLayout() == nil {
		return nil, fmt.Errorf("program file memory addresses are not resolved")
	}

	if !pf.MemoryLayout().Data().ContainsAddress(addr) {
		return nil, fmt.Errorf("address 0x%X is outside of data segment", addr)
	}

	globals := pf.Globals()

	for i := range globals {
		g := &globals[i]
		if g.Address != nil && *g.Address == addr {
			return g, nil
		}
	}

	return nil, fmt.Errorf("no global found at address 0x%X", addr)
}

// Returns the program instruction at the given memory address
//
// Note that this function requires the ProgramFile to have resolved memory addresses.
// Also the result only will be correct if the program file was resolved with the same
// memory layout as the runtime where the address comes from.
func InstructionAtAddress(pf ProgramFile, addr uint32) (*Instruction, error) {
	if pf.MemoryLayout() == nil {
		return nil, fmt.Errorf("program file memory addresses are not resolved")
	}

	if !pf.MemoryLayout().Code().ContainsAddress(addr) {
		return nil, fmt.Errorf("address 0x%X is outside of code segment", addr)
	}

	relativeAddr := addr - pf.MemoryLayout().CodeBase
	if relativeAddr%4 != 0 {
		return nil, fmt.Errorf("address 0x%X is not aligned to instruction size", addr)
	}

	index := relativeAddr / 4
	instructions := pf.Instructions()
	if index >= uint32(len(instructions)) {
		panic("calculated instruction index out of bounds")
	}

	if instructions[index].Address == nil {
		panic("instruction at calculated index has no resolved address")
	}

	if *instructions[index].Address != addr {
		panic("instruction address mismatch at calculated index")
	}

	return &instructions[index], nil
}

// Returns the source location for a given instruction address
func SourceLocationAtInstructionAddress(pf ProgramFile, addr uint32) (*sourcecode.Location, error) {
	if pf.DebugInfo() == nil {
		return nil, fmt.Errorf("program has no debug information")
	}

	if !pf.MemoryLayout().Code().ContainsAddress(addr) {
		return nil, fmt.Errorf("address 0x%X is outside of code segment", addr)
	}

	srcLoc, exists := pf.DebugInfo().InstructionLocations[addr]
	if !exists {
		return nil, fmt.Errorf("no source location found for instruction address 0x%X", addr)
	}

	return srcLoc, nil
}

// Returns the first instruction address for a given source location
func InstructionAddressAtSourceLocation(pf ProgramFile, loc *sourcecode.Location) (uint32, error) {
	if pf.DebugInfo() == nil {
		return 0, fmt.Errorf("program has no debug information")
	}

	for _, entry := range pf.DebugInfo().SortedSourceLocations() {
		addr := entry.Address
		srcLoc := entry.Location

		if srcLoc.File == loc.File && srcLoc.Line == loc.Line {
			return addr, nil
		}
	}

	return 0, fmt.Errorf("no instruction address found for source location %s:%d", loc.File.Path(), loc.Line)
}

// Returns all instruction addresses for a given source location
//
// The function returns the addresses as an slice of contiguous instruction memory addresses as memory ranges.
func InstructionAddressesAtSourceLocation(pf ProgramFile, loc *sourcecode.Location) ([]memory.Range, error) {
	if pf.DebugInfo() == nil {
		return nil, fmt.Errorf("program has no debug information")
	}

	ranges := make([]memory.Range, 0)
	locationMatches := false

	for _, entry := range pf.DebugInfo().SortedSourceLocations() {
		addr := entry.Address
		srcLoc := entry.Location
		locationMatched := locationMatches

		if srcLoc.File == loc.File && srcLoc.Line == loc.Line {
			locationMatches = true

			if !locationMatched {
				// Start of a new range
				ranges = append(ranges, memory.Range{
					Start: addr,
					Size:  4,
				})
			} else {
				// Extend the last range
				lastRange := &ranges[len(ranges)-1]
				lastRange.Size += 4
			}
		}
	}

	return ranges, nil
}

// Returns the source line for a given instruction address
func SourceLineAtInstructionAddress(pf ProgramFile, addr uint32) (*sourcecode.Line, error) {
	srcLoc, err := SourceLocationAtInstructionAddress(pf, addr)
	if err != nil {
		return nil, err
	}

	snippet, err := sourcecode.ReadSnippet(pf.DebugInfo().SourceLibrary, sourcecode.SingleLineRange(srcLoc))
	if err != nil {
		return nil, err
	}

	if len(snippet.Lines) != 1 {
		panic("expected single line snippet")
	}

	return snippet.Lines[0], nil
}

// Given an instruction address, returns the branch target address and if possible the symbol that target address refers to.
func BranchTargetAtInstruction(pf ProgramFile, addr uint32) (*uint32, *SymbolReference, error) {
	instr, err := InstructionAtAddress(pf, addr)
	if err != nil {
		return nil, nil, err
	}

	if instr.Instruction == nil {
		return nil, nil, fmt.Errorf("instruction at address 0x%X is not fully resolved", addr)
	}

	branchTargetReg, err := instructions.BranchTargetRegister(instr.Instruction)
	if err != nil {
		return nil, nil, err
	}

	// Since cucaracha ISA is KISS as fuck, when the compiler emits a jump instruction (conditional or unconditional)
	// it always follows the same pattern:
	//
	//    MOVIMM16H r, TARGET_ADDRESS_16BIT_HIGH  // Load higher 16 bits of the target address into register r
	// 	  MOVIMM16L r, TARGET_ADDRESS_16BIT_LOW   // Load lower 16 bits of the target address into register r
	//    ...
	//    ...                                     // If you look at how we implemented this in our LLVM backend, we lowered jumps
	//    ...                                     // using this exact pattern, so we don't really expect anything else between
	//    ...                                     // the mov immediate instructions and the jump instruction.
	//    ...                                     // But who knows, LLVM works in mysterious ways...
	//    ...
	//    (C)JMP ... r ...                        // Jump to the address contained in register r
	//
	// So what we do to find the target address is to backtrack along the program to find the latest mov immediate pair of
	// instructions involving our target register before the jump instruction, read the immediates, and combine them to get
	// the full 32 bit target address.

	// We cap backtracking to up to 20 instructions before the branch instruction so we don't end up scanning the whole program...
	const maxBacktrackInstructions = 20
	tries := 0
	var targetAddress *uint32
	foundLowerBits := false
	foundHigherBits := false
	var symbolRef *SymbolReference

	for (addr >= pf.MemoryLayout().CodeBase+4 && tries < maxBacktrackInstructions) && !(foundLowerBits && foundHigherBits) {
		addr -= 4
		tries++

		prevInstr, err := InstructionAtAddress(pf, addr)
		if err != nil {
			return nil, nil, fmt.Errorf("error backtracking to find branch target: %w", err)
		}

		if prevInstr.Instruction == nil {
			return nil, nil, fmt.Errorf("error backtracking to find branch target: instruction at address 0x%X is not fully resolved", addr)
		}

		switch prevInstr.Instruction.Descriptor.OpCode.OpCode {
		case instructions.OpCode_MOV_IMM16H:
			if len(prevInstr.Instruction.OperandValues) != 2 {
				panic("expected 2 operands for MOVIMM16H instruction")
			}

			if prevInstr.Instruction.OperandValues[0].Kind() != instructions.OperandKind_Register {
				panic("expected first operand of MOVIMM16H to be a register")
			}

			if prevInstr.Instruction.OperandValues[0].Register() == branchTargetReg {
				immValue := uint32(prevInstr.Instruction.OperandValues[1].Immediate().Encode())
				if targetAddress == nil {
					targetAddress = new(uint32)
				}
				*targetAddress = (*targetAddress & 0x0000FFFF) | (immValue << 16)
				foundHigherBits = true

				// Try to find symbol reference for this instruction
				for _, symRef := range prevInstr.Symbols {
					if symRef.Usage == SymbolUsageHi && symRef.Name != "" {
						if symbolRef != nil && symbolRef.Name != symRef.Name {
							// Conflicting symbol references for the same target register
							return nil, nil, fmt.Errorf("conflicting symbol references for branch target register %s: '%s' and '%s'", branchTargetReg.Name(), symbolRef.Name, symRef.Name)
						}
						symbolRef = &symRef
						break
					}
				}
			}

		case instructions.OpCode_MOV_IMM16L:
			if len(prevInstr.Instruction.OperandValues) != 2 {
				panic("expected 2 operands for MOVIMM16L instruction")
			}

			if prevInstr.Instruction.OperandValues[0].Kind() != instructions.OperandKind_Register {
				panic("expected first operand of MOVIMM16L to be a register")
			}

			if prevInstr.Instruction.OperandValues[0].Register() == branchTargetReg {
				immValue := uint32(prevInstr.Instruction.OperandValues[1].Immediate().Encode())
				if targetAddress == nil {
					targetAddress = new(uint32)
				}
				*targetAddress = (*targetAddress & 0xFFFF0000) | (immValue & 0x0000FFFF)
				foundLowerBits = true

				// Try to find symbol reference for this instruction
				for _, symRef := range prevInstr.Symbols {
					if symRef.Usage == SymbolUsageLo && symRef.Name != "" {
						if symbolRef != nil && symbolRef.Name != symRef.Name {
							// Conflicting symbol references for the same target register
							return nil, nil, fmt.Errorf("conflicting symbol references for branch target register %s: '%s' and '%s'", branchTargetReg.Name(), symbolRef.Name, symRef.Name)
						}
						symbolRef = &symRef
						break
					}
				}
			}
		}
	}

	if (targetAddress == nil) || !(foundLowerBits && foundHigherBits) {
		return nil, nil, fmt.Errorf("could not determine branch target address for instruction at 0x%X", addr)
	}

	return targetAddress, symbolRef, nil
}

// Returns N instructions starting at the given memory address
func InstructionsAtAddress(program ProgramFile, addr uint32, count int) ([]*Instruction, error) {
	instructions := make([]*Instruction, 0, count)

	for i := 0; i < count; i++ {
		instr, err := InstructionAtAddress(program, addr)
		if err != nil {
			return nil, err
		}

		instructions = append(instructions, instr)
		addr += 4
	}

	return instructions, nil
}
