package cpu_test

import (
	"strings"
	"testing"

	"github.com/Manu343726/cucaracha/pkg/hw/components"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/interpreter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCPUInterface tests that both CPU implementations satisfy the interface
func TestCPUInterface(t *testing.T) {
	t.Run("Emulator implements CPU", func(t *testing.T) {
		var c cpu.CPU = interpreter.NewEmulator(1024)
		assert.NotNil(t, c)
	})

	t.Run("HardwareCPU implements CPU", func(t *testing.T) {
		var c cpu.CPU = components.NewHardwareCPU(1024)
		assert.NotNil(t, c)
	})
}

// TestCPUBasicOperations tests basic CPU operations on both implementations
func TestCPUBasicOperations(t *testing.T) {
	cpus := map[string]cpu.CPU{
		"Emulator":    interpreter.NewEmulator(1024),
		"HardwareCPU": components.NewHardwareCPU(1024),
	}

	for name, c := range cpus {
		t.Run(name+"/RegisterOperations", func(t *testing.T) {
			// Test register read/write
			c.SetRegister(cpu.RegR0, 42)
			assert.Equal(t, uint32(42), c.GetRegister(cpu.RegR0))

			c.SetRegister(cpu.RegR1, 100)
			assert.Equal(t, uint32(100), c.GetRegister(cpu.RegR1))
		})

		t.Run(name+"/PCOperations", func(t *testing.T) {
			// Reset for clean state
			c.Reset()

			c.SetPC(0x1000)
			assert.Equal(t, uint32(0x1000), c.GetPC())
		})

		t.Run(name+"/MemoryOperations", func(t *testing.T) {
			c.Reset()

			// Test word write/read
			c.WriteMemory(0x100, 0xDEADBEEF)
			assert.Equal(t, uint32(0xDEADBEEF), c.ReadMemory(0x100))

			// Test byte write/read
			c.WriteByte(0x200, 0x42)
			assert.Equal(t, byte(0x42), c.ReadByte(0x200))
		})

		t.Run(name+"/HaltOperation", func(t *testing.T) {
			c.Reset()
			assert.False(t, c.IsHalted())

			c.Halt()
			assert.True(t, c.IsHalted())

			c.Reset()
			assert.False(t, c.IsHalted())
		})

		t.Run(name+"/LoadProgram", func(t *testing.T) {
			c.Reset()

			// Create a simple program: NOP, NOP, HALT (we'll simulate halt by checking IsHalted after limited steps)
			program := []uint32{
				0x00, // NOP (opcode 0)
				0x00, // NOP
			}

			err := c.LoadProgram(program, 0x100)
			require.NoError(t, err)

			assert.Equal(t, uint32(0x100), c.GetPC())
			assert.Equal(t, uint32(0x00), c.ReadMemory(0x100))
		})
	}
}

// TestCPUExecution tests instruction execution on both implementations
func TestCPUExecution(t *testing.T) {
	cpus := map[string]cpu.CPU{
		"Emulator":    interpreter.NewEmulator(1024),
		"HardwareCPU": components.NewHardwareCPU(1024),
	}

	for name, c := range cpus {
		t.Run(name+"/ExecuteNOP", func(t *testing.T) {
			c.Reset()

			// Load NOP instruction
			program := []uint32{0x00} // NOP
			err := c.LoadProgram(program, 0x100)
			require.NoError(t, err)

			startPC := c.GetPC()
			err = c.Step()
			require.NoError(t, err)

			// PC should have advanced by 4
			assert.Equal(t, startPC+4, c.GetPC())
		})
	}
}

// TestDebuggableCPU tests the DebuggableCPU interface
func TestDebuggableCPU(t *testing.T) {
	cpus := map[string]cpu.DebuggableCPU{
		"Emulator":    interpreter.NewEmulator(1024),
		"HardwareCPU": components.NewHardwareCPU(1024),
	}

	for name, c := range cpus {
		t.Run(name+"/DecodeInstruction", func(t *testing.T) {
			c.Reset()

			// Load a NOP instruction
			c.WriteMemory(0x100, 0x00)

			mnemonic, _, err := c.DecodeInstruction(0x100)
			require.NoError(t, err)
			assert.Equal(t, "nop", strings.ToLower(mnemonic))
		})

		t.Run(name+"/FlagsOperations", func(t *testing.T) {
			c.Reset()

			c.SetFlags(0x0F) // Set all flags
			assert.Equal(t, uint32(0x0F), c.GetFlags())
		})
	}
}
