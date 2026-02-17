package runtime

import (
	"fmt"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu"
	"github.com/Manu343726/cucaracha/pkg/hw/memory"
	"github.com/Manu343726/cucaracha/pkg/runtime/program"
)

func ReadVariable(r Runtime, variable *program.VariableInfo) (uint32, error) {
	if variable.Location == nil {
		return 0, fmt.Errorf("variable '%s' was optimized out (no location info)", variable.Name)
	}

	switch variable.Location.Type() {
	case program.VariableLocationRegister:
		regLoc := variable.Location.(program.RegisterLocation)
		return cpu.ReadRegisterByEncodedId(r.CPU().Registers(), regLoc.Register)
	case program.VariableLocationMemory:
		memLoc := variable.Location.(program.MemoryLocation)
		baseAddr, err := cpu.ReadRegisterByEncodedId(r.CPU().Registers(), memLoc.BaseRegister)
		if err != nil {
			return 0, fmt.Errorf("failed to read base register for variable '%s': %w", variable.Name, err)
		}

		address := baseAddr + uint32(memLoc.Offset)
		return memory.ReadUint32(r.Memory(), address)
	case program.VariableLocationConstant:
		constLoc := variable.Location.(program.ConstantLocation)
		return uint32(constLoc.Value), nil
	default:
		return 0, fmt.Errorf("unknown location type for variable '%s'", variable.Name)
	}
}

// Returns the value of a symbol by its name
//
// The symbol may be a global variable, a function parameter, a function local variable, or a function.
// If the symbol is a variable, its current value is returned.
// If the symbol is a function, its starting address is returned.
func ResolveSymbol(r Runtime, p program.ProgramFile, name string) (uint32, error) {
	// First use debug info if available to search using original source-level names
	if p.DebugInfo() != nil {
		pc, err := cpu.ReadPC(r.CPU().Registers())
		if err != nil {
			return 0, fmt.Errorf("failed to read PC register: %w", err)
		}

		// Try to find a variable in the current function scope
		for _, variable := range p.DebugInfo().GetVariables(pc) {
			if variable.Name == name {
				return ReadVariable(r, &variable)
			}
		}
	}

	// Try to find a function by name
	if function, err := program.FunctionByName(p, name); err == nil {
		firstInstruction := p.Instructions()[function.FirstInstructionIndex()]
		if firstInstruction.Address == nil {
			return 0, fmt.Errorf("function '%s' has no resolved address", name)
		}

		return *firstInstruction.Address, nil
	}

	// Try to find a global variable by name
	if global, err := program.GlobalByName(p, name); err == nil {
		if global.Address == nil {
			return 0, fmt.Errorf("global variable '%s' has no resolved address", name)
		}

		return memory.ReadUint32(r.Memory(), *global.Address)
	}

	return 0, fmt.Errorf("symbol '%s' not found", name)
}
