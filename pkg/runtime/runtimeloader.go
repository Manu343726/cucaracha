package runtime

import (
	"fmt"

	"github.com/Manu343726/cucaracha/pkg/hw/memory"
	"github.com/Manu343726/cucaracha/pkg/runtime/program"
	"github.com/Manu343726/cucaracha/pkg/system"
)

// Saves the system descriptor in runtime memory
func LoadSystemDescriptor(r Runtime, s *system.SystemDescriptor) error {
	log().Debug("loading system descriptor")

	if err := system.EncodeSystemDescriptor(s, memory.NewMemoryWriter(memory.MemorySlice(r.Memory()))); err != nil {
		return fmt.Errorf("failed to save system descriptor to memory: %w", err)
	}
	return nil
}

// Loads a program into the given runtime environment
func Load(r Runtime, s *system.SystemDescriptor, programFile program.ProgramFile) error {
	if err := LoadSystemDescriptor(r, s); err != nil {
		return fmt.Errorf("failed to load system descriptor: %w", err)
	}

	if err := LoadProgram(programFile, r); err != nil {
		return fmt.Errorf("failed to load program: %w", err)
	}

	return nil
}
