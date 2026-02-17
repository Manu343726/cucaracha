package runtime

import (
	"fmt"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
	"github.com/Manu343726/cucaracha/pkg/hw/memory"
	"github.com/Manu343726/cucaracha/pkg/runtime/program"
)

// Implements program loading into the runtime
type ProgramLoader struct {
	programFile program.ProgramFile
	runtime     Runtime
}

// Returns a program loader for the given program file and runtime
func NewProgramLoader(programFile program.ProgramFile, runtime Runtime) *ProgramLoader {
	return &ProgramLoader{
		programFile: programFile,
		runtime:     runtime,
	}
}

// Loads the program code into the runtime memory
func (pl *ProgramLoader) LoadCode() error {
	if pl.programFile.MemoryLayout() == nil {
		return fmt.Errorf("program file has no resolved memory addresses")
	}

	progLayout := pl.programFile.MemoryLayout()
	runtimeLayout := pl.runtime.MemoryLayout()

	// Program code base must match runtime (no relocation support)
	if progLayout.CodeBase != runtimeLayout.CodeBase {
		return fmt.Errorf("program code base does not match runtime: program 0x%X, runtime 0x%X",
			progLayout.CodeBase, runtimeLayout.CodeBase)
	}

	// Program code must fit within runtime code region
	if progLayout.CodeSize > runtimeLayout.CodeSize {
		return fmt.Errorf("program code does not fit in runtime: program size 0x%X, runtime size 0x%X",
			progLayout.CodeSize, runtimeLayout.CodeSize)
	}

	for i, instr := range pl.programFile.Instructions() {
		expectedAddr := pl.programFile.MemoryLayout().CodeBase + uint32(i*4)
		if instr.Address == nil {
			return fmt.Errorf("instruction %d has no resolved address", i)
		} else if *instr.Address != expectedAddr {
			return fmt.Errorf("instruction %d has unexpected address 0x%X (expected 0x%X)", i, *instr.Address, expectedAddr)
		}

		encoded, err := mc.Descriptor.Instructions.Encode(instr.Instruction)
		if err != nil {
			return fmt.Errorf("failed to encode instruction %d: %w", i, err)
		}

		err = memory.WriteUint32(pl.runtime.Memory(), *instr.Address, encoded)
		if err != nil {
			return fmt.Errorf("failed to write instruction %d to memory at address 0x%X: %w", i, *instr.Address, err)
		}
	}

	return nil
}

// Loads the program data into the runtime memory
func (pl *ProgramLoader) LoadData() error {
	for _, global := range pl.programFile.Globals() {
		if global.Range() == nil {
			return fmt.Errorf("global '%s' has no resolved memory location", global.Name)
		}

		if !pl.runtime.MemoryLayout().Data().ContainsRange(*global.Range()) {
			return fmt.Errorf("global '%s' range %s is outside of the data segment", global.Name, global.Range().String())
		}

		if err := memory.NewSlice(pl.runtime.Memory(), global.Range()).Write(global.InitialData); err != nil {
			return fmt.Errorf("failed to write global '%s' data to memory at address 0x%X: %w", global.Name, *global.Address, err)
		}
	}

	return nil
}

// Configures the CPU registers according to the program file
func (pl *ProgramLoader) SetupCPU() error {
	entrypoint, err := program.ProgramEntryPoint(pl.programFile)
	if err != nil {
		return fmt.Errorf("failed to get program entry point: %w", err)
	}

	// Adjust entrypoint to be absolute in memory
	entrypoint = entrypoint + pl.runtime.MemoryLayout().CodeBase

	// Set the LR to the end of program mark address
	if err := cpu.WriteLR(pl.runtime.CPU().Registers(), 0xffffffff); err != nil {
		return fmt.Errorf("failed to set LR to end of program mark address: %w", err)
	}

	// Set the PC to the program entry point
	if err := cpu.WritePC(pl.runtime.CPU().Registers(), entrypoint); err != nil {
		return fmt.Errorf("failed to set PC to program entry point 0x%X: %w", entrypoint, err)
	}

	// Set the SP to the stack bottom
	if err := cpu.WriteSP(pl.runtime.CPU().Registers(), pl.runtime.MemoryLayout().StackBottom()); err != nil {
		return fmt.Errorf("failed to set SP to stack bottom 0x%X: %w", pl.runtime.MemoryLayout().StackBottom(), err)
	}

	return nil
}

// Loads the entire program into the runtime
func (pl *ProgramLoader) Load() error {
	if err := pl.LoadCode(); err != nil {
		return fmt.Errorf("failed to load program code: %w", err)
	}

	if err := pl.LoadData(); err != nil {
		return fmt.Errorf("failed to load program data: %w", err)
	}

	if err := pl.SetupCPU(); err != nil {
		return fmt.Errorf("failed to setup CPU for program execution: %w", err)
	}

	return nil
}

// Loads a program into the given runtime
func LoadProgram(program program.ProgramFile, runtime Runtime) error {
	loader := NewProgramLoader(program, runtime)
	return loader.Load()
}
