package cpu_test

import (
	"testing"

	"github.com/Manu343726/cucaracha/pkg/hw/components"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/interpreter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCPURunnerWithEmulator tests CPURunner with the software emulator
func TestCPURunnerWithEmulator(t *testing.T) {
	createRunner := func() *cpu.CPURunner {
		emulator := interpreter.NewEmulator(4096)
		return cpu.NewCPURunner(emulator)
	}

	testCPURunner(t, "Emulator", createRunner)
}

// TestCPURunnerWithHardwareCPU tests CPURunner with the hardware CPU
func TestCPURunnerWithHardwareCPU(t *testing.T) {
	createRunner := func() *cpu.CPURunner {
		hwCPU := components.NewHardwareCPU(4096)
		return cpu.NewCPURunner(hwCPU)
	}

	testCPURunner(t, "HardwareCPU", createRunner)
}

func testCPURunner(t *testing.T, name string, createRunner func() *cpu.CPURunner) {
	t.Run(name+"/Step", func(t *testing.T) {
		runner := createRunner()

		// Load NOP instruction
		err := runner.LoadProgram([]uint32{0x00}, 0x100)
		require.NoError(t, err)

		result := runner.Step()
		require.NotNil(t, result)
		assert.Equal(t, cpu.StopStep, result.StopReason)
		assert.Equal(t, 1, result.StepsExecuted)
	})

	t.Run(name+"/RunN", func(t *testing.T) {
		runner := createRunner()

		// Load 5 NOP instructions
		program := []uint32{0x00, 0x00, 0x00, 0x00, 0x00}
		err := runner.LoadProgram(program, 0x100)
		require.NoError(t, err)

		result := runner.RunN(3)
		require.NotNil(t, result)
		assert.Equal(t, 3, result.StepsExecuted)
	})

	t.Run(name+"/Breakpoint", func(t *testing.T) {
		runner := createRunner()

		// Load 5 NOP instructions
		program := []uint32{0x00, 0x00, 0x00, 0x00, 0x00}
		err := runner.LoadProgram(program, 0x100)
		require.NoError(t, err)

		// Add breakpoint at third instruction
		bpID := runner.AddBreakpoint(0x108)

		result := runner.Run()
		require.NotNil(t, result)
		assert.Equal(t, cpu.StopBreakpoint, result.StopReason)
		assert.Equal(t, bpID, result.BreakpointID)
		assert.Equal(t, uint32(0x108), result.LastPC)
	})

	t.Run(name+"/TerminationAddress", func(t *testing.T) {
		runner := createRunner()

		// Load NOP instructions
		program := []uint32{0x00, 0x00, 0x00}
		err := runner.LoadProgram(program, 0x100)
		require.NoError(t, err)

		// Add termination at address after program
		runner.AddTerminationAddress(0x10C)

		result := runner.Run()
		require.NotNil(t, result)
		assert.Equal(t, cpu.StopTermination, result.StopReason)
		// 3 NOP instructions executed before reaching termination address
		assert.Equal(t, 3, result.StepsExecuted)
	})

	t.Run(name+"/EventCallback", func(t *testing.T) {
		runner := createRunner()

		// Load NOP instructions
		program := []uint32{0x00, 0x00, 0x00, 0x00, 0x00}
		err := runner.LoadProgram(program, 0x100)
		require.NoError(t, err)

		stepCount := 0
		runner.SetEventCallback(func(event cpu.ExecutionEvent, result *cpu.ExecutionResult) bool {
			if event == cpu.EventStep {
				stepCount++
				// Stop after 3 steps
				return stepCount < 3
			}
			return true
		})

		runner.Run()
		// Callback stops after 3 steps
		assert.Equal(t, 3, stepCount)
	})

	t.Run(name+"/Continue", func(t *testing.T) {
		runner := createRunner()

		// Load 5 NOP instructions
		program := []uint32{0x00, 0x00, 0x00, 0x00, 0x00}
		err := runner.LoadProgram(program, 0x100)
		require.NoError(t, err)

		// Add breakpoint at 2nd instruction (0x104)
		bp1ID := runner.AddBreakpoint(0x104)

		// Add termination after program (0x114)
		runner.AddTerminationAddress(0x114)

		// Run to first breakpoint
		result := runner.Run()
		assert.Equal(t, cpu.StopBreakpoint, result.StopReason)
		assert.Equal(t, bp1ID, result.BreakpointID)
		assert.Equal(t, uint32(0x104), runner.GetPC())

		// Continue to termination (should skip past breakpoint and run to end)
		result = runner.Continue()
		assert.Equal(t, cpu.StopTermination, result.StopReason)
	})
}
