package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// CPU Creation Tests
// =============================================================================

func TestCPU(t *testing.T) {
	t.Run("NewCPU creates CPU with internal components", func(t *testing.T) {
		cpu := NewCPU("CPU", 1024, 256)

		assert.Equal(t, "CPU", cpu.Name())
		assert.Equal(t, "CPU", cpu.Type())
		assert.NotNil(t, cpu.PC())
		assert.NotNil(t, cpu.IR())
		assert.NotNil(t, cpu.Decoder())
		assert.NotNil(t, cpu.ControlUnit())
		assert.NotNil(t, cpu.ALU())
		assert.NotNil(t, cpu.Registers())
		assert.NotNil(t, cpu.Memory())
	})

	t.Run("Initial state", func(t *testing.T) {
		cpu := NewCPU("CPU", 1024, 256)
		assert.Equal(t, uint32(0), cpu.GetPC())
		assert.Equal(t, uint64(0), cpu.Cycles())
		assert.False(t, cpu.IsHalted())
	})

	t.Run("Registry", func(t *testing.T) {
		desc, err := Registry.Get("CPU")
		require.NoError(t, err)
		assert.Equal(t, "CPU", desc.Name)
		assert.Equal(t, CategoryCPU, desc.Category)
	})
}

// =============================================================================
// CPU Memory Tests
// =============================================================================

func TestCPUMemory(t *testing.T) {
	t.Run("ReadMemory and WriteMemory", func(t *testing.T) {
		cpu := NewCPU("CPU", 1024, 256)

		cpu.WriteMemory(0, 0x12345678)
		assert.Equal(t, uint32(0x12345678), cpu.ReadMemory(0))

		cpu.WriteMemory(100, 0xDEADBEEF)
		assert.Equal(t, uint32(0xDEADBEEF), cpu.ReadMemory(100))
	})

	t.Run("LoadBinary", func(t *testing.T) {
		cpu := NewCPU("CPU", 1024, 256)
		data := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

		err := cpu.LoadBinary(data, 0x100)
		require.NoError(t, err)

		assert.Equal(t, uint32(0x04030201), cpu.ReadMemory(0x100))
		assert.Equal(t, uint32(0x08070605), cpu.ReadMemory(0x104))
		assert.Equal(t, uint32(0x100), cpu.GetPC())
	})

	t.Run("LoadProgram", func(t *testing.T) {
		cpu := NewCPU("CPU", 1024, 256)
		program := []uint32{0x12345678, 0xDEADBEEF}

		err := cpu.LoadProgram(program, 0)
		require.NoError(t, err)

		assert.Equal(t, uint32(0x12345678), cpu.ReadMemory(0))
		assert.Equal(t, uint32(0xDEADBEEF), cpu.ReadMemory(4))
	})
}

// =============================================================================
// CPU Instruction Execution Tests
// =============================================================================

func TestCPUNop(t *testing.T) {
	cpu := NewCPU("CPU", 1024, 256)

	// NOP instruction
	program := []uint32{EncodeInstruction(OP_NOP, 0, 0, 0)}
	cpu.LoadProgram(program, 0)

	err := cpu.StepInstruction()
	require.NoError(t, err)

	// PC should advance by 4
	assert.Equal(t, uint32(4), cpu.GetPC())
}

func TestCPUMovImm16L(t *testing.T) {
	cpu := NewCPU("CPU", 1024, 256)

	// MOVIMM16L r5, #0x1234
	program := []uint32{EncodeImmInstruction(OP_MOV_IMM16L, 0x1234, 5)}
	cpu.LoadProgram(program, 0)

	err := cpu.StepInstruction()
	require.NoError(t, err)

	assert.Equal(t, uint32(0x1234), cpu.GetRegister(5))
}

func TestCPUMovImm16H(t *testing.T) {
	cpu := NewCPU("CPU", 1024, 256)

	// First set low 16 bits
	cpu.SetRegister(5, 0x5678)

	// MOVIMM16H r5, #0x1234 - should set high 16 bits, preserve low
	program := []uint32{EncodeImmInstruction(OP_MOV_IMM16H, 0x1234, 5)}
	cpu.LoadProgram(program, 0)

	err := cpu.StepInstruction()
	require.NoError(t, err)

	assert.Equal(t, uint32(0x12345678), cpu.GetRegister(5))
}

func TestCPUMov(t *testing.T) {
	cpu := NewCPU("CPU", 1024, 256)

	cpu.SetRegister(1, 0x42)

	// MOV r2, r1
	program := []uint32{EncodeInstruction(OP_MOV, 1, 2, 0)}
	cpu.LoadProgram(program, 0)

	err := cpu.StepInstruction()
	require.NoError(t, err)

	assert.Equal(t, uint32(0x42), cpu.GetRegister(2))
}

func TestCPUAdd(t *testing.T) {
	cpu := NewCPU("CPU", 1024, 256)

	cpu.SetRegister(1, 10)
	cpu.SetRegister(2, 20)

	// ADD r3, r1, r2
	program := []uint32{EncodeInstruction(OP_ADD, 1, 2, 3)}
	cpu.LoadProgram(program, 0)

	err := cpu.StepInstruction()
	require.NoError(t, err)

	assert.Equal(t, uint32(30), cpu.GetRegister(3))
}

func TestCPUSub(t *testing.T) {
	cpu := NewCPU("CPU", 1024, 256)

	cpu.SetRegister(1, 50)
	cpu.SetRegister(2, 20)

	// SUB r3, r1, r2
	program := []uint32{EncodeInstruction(OP_SUB, 1, 2, 3)}
	cpu.LoadProgram(program, 0)

	err := cpu.StepInstruction()
	require.NoError(t, err)

	assert.Equal(t, uint32(30), cpu.GetRegister(3))
}

func TestCPUMul(t *testing.T) {
	cpu := NewCPU("CPU", 1024, 256)

	cpu.SetRegister(1, 6)
	cpu.SetRegister(2, 7)

	// MUL r3, r1, r2
	program := []uint32{EncodeInstruction(OP_MUL, 1, 2, 3)}
	cpu.LoadProgram(program, 0)

	err := cpu.StepInstruction()
	require.NoError(t, err)

	assert.Equal(t, uint32(42), cpu.GetRegister(3))
}

func TestCPUDiv(t *testing.T) {
	cpu := NewCPU("CPU", 1024, 256)

	cpu.SetRegister(1, 100)
	cpu.SetRegister(2, 10)

	// DIV r3, r1, r2
	program := []uint32{EncodeInstruction(OP_DIV, 1, 2, 3)}
	cpu.LoadProgram(program, 0)

	err := cpu.StepInstruction()
	require.NoError(t, err)

	assert.Equal(t, uint32(10), cpu.GetRegister(3))
}

func TestCPULoadStore(t *testing.T) {
	cpu := NewCPU("CPU", 1024, 256)

	// Write a value to memory
	cpu.WriteMemory(0x100, 0xDEADBEEF)

	// Set r1 = address
	cpu.SetRegister(1, 0x100)

	// LD r2, [r1]
	program := []uint32{EncodeInstruction(OP_LD, 1, 2, 0)}
	cpu.LoadProgram(program, 0)

	err := cpu.StepInstruction()
	require.NoError(t, err)

	assert.Equal(t, uint32(0xDEADBEEF), cpu.GetRegister(2))
}

func TestCPUStore(t *testing.T) {
	cpu := NewCPU("CPU", 1024, 256)

	// Set value and address
	cpu.SetRegister(1, 0x12345678) // Value
	cpu.SetRegister(2, 0x200)      // Address

	// ST r1, [r2]
	program := []uint32{EncodeInstruction(OP_ST, 1, 2, 0)}
	cpu.LoadProgram(program, 0)

	err := cpu.StepInstruction()
	require.NoError(t, err)

	assert.Equal(t, uint32(0x12345678), cpu.ReadMemory(0x200))
}

func TestCPUJump(t *testing.T) {
	cpu := NewCPU("CPU", 1024, 256)

	// Set target address
	cpu.SetRegister(1, 0x100)

	// JMP r1, r2 (link register)
	program := []uint32{EncodeInstruction(OP_JMP, 1, 2, 0)}
	cpu.LoadProgram(program, 0)

	err := cpu.StepInstruction()
	require.NoError(t, err)

	// PC should be at target
	assert.Equal(t, uint32(0x100), cpu.GetPC())
	// Link register should have return address
	assert.Equal(t, uint32(4), cpu.GetRegister(2))
}

func TestCPUConditionalJump(t *testing.T) {
	// CJMP op1=condition, op2=target, op3=link
	// Condition codes: 0=EQ(Z=1), 1=NE(Z=0), 2=CS(C=1), 3=CC(C=0), etc.

	t.Run("CJMP takes branch when condition satisfied (EQ with Z=1)", func(t *testing.T) {
		cpu := NewCPU("CPU", 1024, 256)

		// Setup:
		// r1 = condition code (0 = EQ, which tests Z flag)
		// r2 = target address
		// r3 = link register
		// CPSR (r2 in register bank) has Z flag set
		cpu.SetRegister(1, 0)     // condition code EQ
		cpu.SetRegister(2, 0x100) // target address
		cpu.SetRegister(3, 0)     // link will be written here

		// Set CPSR flags in register 2 (but we use a different register for flags)
		// Actually, looking at CPU code: flags come from register index 2
		// We need to set up properly - let me use different registers

		// Use r10 for condition, r11 for target, r12 for link
		cpu.SetRegister(10, 0)     // condition code EQ (test Z flag)
		cpu.SetRegister(11, 0x100) // target address
		cpu.SetRegister(12, 0)     // link register

		// Set CPSR (register 2) with Zero flag set
		cpu.registers.Set(2, uint64(FlagZero))

		// CJMP r10, r11, r12
		program := []uint32{EncodeInstruction(OP_CJMP, 10, 11, 12)}
		cpu.LoadProgram(program, 0)

		err := cpu.StepInstruction()
		require.NoError(t, err)

		// Branch should be taken
		assert.Equal(t, uint32(0x100), cpu.GetPC(), "PC should jump to target")
		assert.Equal(t, uint32(4), cpu.GetRegister(12), "Link register should have return address")
	})

	t.Run("CJMP does not branch when condition not satisfied (EQ with Z=0)", func(t *testing.T) {
		cpu := NewCPU("CPU", 1024, 256)

		// Use r10 for condition, r11 for target, r12 for link
		cpu.SetRegister(10, 0)     // condition code EQ (test Z flag)
		cpu.SetRegister(11, 0x100) // target address
		cpu.SetRegister(12, 0)     // link register

		// Set CPSR (register 2) with Zero flag NOT set
		cpu.registers.Set(2, 0)

		// CJMP r10, r11, r12
		program := []uint32{EncodeInstruction(OP_CJMP, 10, 11, 12)}
		cpu.LoadProgram(program, 0)

		err := cpu.StepInstruction()
		require.NoError(t, err)

		// Branch should NOT be taken, PC should just advance
		assert.Equal(t, uint32(4), cpu.GetPC(), "PC should advance to next instruction")
		assert.Equal(t, uint32(0), cpu.GetRegister(12), "Link register should be unchanged")
	})

	t.Run("CJMP NE condition (Z=0 means branch)", func(t *testing.T) {
		cpu := NewCPU("CPU", 1024, 256)

		cpu.SetRegister(10, 1)     // condition code NE (branch if Z=0)
		cpu.SetRegister(11, 0x200) // target address
		cpu.SetRegister(12, 0)     // link register

		// CPSR with Z=0 (no flags set)
		cpu.registers.Set(2, 0)

		program := []uint32{EncodeInstruction(OP_CJMP, 10, 11, 12)}
		cpu.LoadProgram(program, 0)

		err := cpu.StepInstruction()
		require.NoError(t, err)

		// NE with Z=0 should branch
		assert.Equal(t, uint32(0x200), cpu.GetPC(), "PC should jump to target")
		assert.Equal(t, uint32(4), cpu.GetRegister(12), "Link register should have return address")
	})

	t.Run("CJMP GT condition (Z=0 and N=V)", func(t *testing.T) {
		cpu := NewCPU("CPU", 1024, 256)

		// GT: Z=0 AND N=V (condition code 12)
		cpu.SetRegister(10, 12)    // condition code GT
		cpu.SetRegister(11, 0x300) // target address
		cpu.SetRegister(12, 0)     // link register

		// CPSR with Z=0, N=0, V=0 (N=V satisfied, Z=0 satisfied)
		cpu.registers.Set(2, 0)

		program := []uint32{EncodeInstruction(OP_CJMP, 10, 11, 12)}
		cpu.LoadProgram(program, 0)

		err := cpu.StepInstruction()
		require.NoError(t, err)

		// GT with Z=0 and N=V should branch
		assert.Equal(t, uint32(0x300), cpu.GetPC(), "PC should jump to target")
	})

	t.Run("CJMP loop pattern", func(t *testing.T) {
		cpu := NewCPU("CPU", 1024, 256)

		// Simulate a simple countdown loop using CMP to set flags:
		// NOTE: Register 2 is the implicit CPSR register, so avoid using it for data!
		// r10 = counter (starts at 3)
		// r11 = 1 (decrement value)
		// r12 = 0 (zero for comparison)
		// r13 = loop target address (0 = start of program)
		// r14 = condition code register (NE = 1)
		// r15 = link register
		// r16 = CPSR destination for CMP (though also writes to r2)
		// Loop: SUB r10, r11, r10; CMP r10, r12, r16; CJMP r14, r13, r15

		cpu.SetRegister(10, 3) // counter
		cpu.SetRegister(11, 1) // decrement
		cpu.SetRegister(12, 0) // zero for comparison
		cpu.SetRegister(13, 0) // loop target (address 0)
		cpu.SetRegister(14, 1) // condition NE (continue while Z=0)
		cpu.SetRegister(15, 0) // link (unused)
		cpu.SetRegister(16, 0) // CPSR destination

		program := []uint32{
			EncodeInstruction(OP_SUB, 10, 11, 10),  // r10 = r10 - r11
			EncodeInstruction(OP_CMP, 10, 12, 16),  // CMP r10, r12 -> sets flags in register 2 (CPSR)
			EncodeInstruction(OP_CJMP, 14, 13, 15), // if NE, jump to r13
		}
		cpu.LoadProgram(program, 0)

		// Execute the loop - should run 3 iterations (when counter reaches 0, Z=1, NE fails)
		maxSteps := 15 // safety limit (3 iterations * 3 instructions + buffer)
		for step := 0; step < maxSteps; step++ {
			pc := cpu.GetPC()
			err := cpu.StepInstruction()
			require.NoError(t, err, "Step %d at PC %d", step, pc)

			// If we fell through the CJMP, we're done
			if cpu.GetPC() == 12 {
				break
			}
		}

		// Counter should be 0 after 3 full iterations
		assert.Equal(t, uint32(0), cpu.GetRegister(10), "Counter should be 0")
	})
}

func TestCPUShifts(t *testing.T) {
	cpu := NewCPU("CPU", 1024, 256)

	t.Run("LSL", func(t *testing.T) {
		cpu.Reset()
		cpu.SetRegister(1, 1)
		cpu.SetRegister(2, 4)
		program := []uint32{EncodeInstruction(OP_LSL, 1, 2, 3)}
		cpu.LoadProgram(program, 0)
		cpu.StepInstruction()
		assert.Equal(t, uint32(16), cpu.GetRegister(3))
	})

	t.Run("LSR", func(t *testing.T) {
		cpu.Reset()
		cpu.SetRegister(1, 16)
		cpu.SetRegister(2, 2)
		program := []uint32{EncodeInstruction(OP_LSR, 1, 2, 3)}
		cpu.LoadProgram(program, 0)
		cpu.StepInstruction()
		assert.Equal(t, uint32(4), cpu.GetRegister(3))
	})
}

// =============================================================================
// CPU Program Execution Tests
// =============================================================================

func TestCPUMultipleInstructions(t *testing.T) {
	cpu := NewCPU("CPU", 1024, 256)

	// Simple program: r1 = 10, r2 = 20, r3 = r1 + r2
	program := []uint32{
		EncodeImmInstruction(OP_MOV_IMM16L, 10, 1), // r1 = 10
		EncodeImmInstruction(OP_MOV_IMM16L, 20, 2), // r2 = 20
		EncodeInstruction(OP_ADD, 1, 2, 3),         // r3 = r1 + r2
	}
	cpu.LoadProgram(program, 0)

	// Execute all instructions
	for i := 0; i < 3; i++ {
		err := cpu.StepInstruction()
		require.NoError(t, err)
	}

	assert.Equal(t, uint32(10), cpu.GetRegister(1))
	assert.Equal(t, uint32(20), cpu.GetRegister(2))
	assert.Equal(t, uint32(30), cpu.GetRegister(3))
}

func TestCPUReset(t *testing.T) {
	cpu := NewCPU("CPU", 1024, 256)

	// Execute some instructions
	cpu.SetRegister(1, 100)
	cpu.SetPC(0x100)

	cpu.Reset()

	assert.Equal(t, uint32(0), cpu.GetPC())
	assert.Equal(t, uint32(0), cpu.GetRegister(1))
	assert.Equal(t, uint64(0), cpu.Cycles())
	assert.False(t, cpu.IsHalted())
}

func TestCPUHalt(t *testing.T) {
	cpu := NewCPU("CPU", 1024, 256)

	cpu.Halt()
	assert.True(t, cpu.IsHalted())

	// Step should not execute when halted
	oldPC := cpu.GetPC()
	cpu.Step()
	assert.Equal(t, oldPC, cpu.GetPC())
}

// =============================================================================
// Encode Instruction Helper Tests
// =============================================================================

func TestEncodeInstruction(t *testing.T) {
	t.Run("Opcode encoding", func(t *testing.T) {
		instr := EncodeInstruction(6, 0, 0, 0)
		assert.Equal(t, uint8(6), uint8(instr&0x1F))
	})

	t.Run("Full encoding", func(t *testing.T) {
		instr := EncodeInstruction(6, 1, 2, 3)
		assert.Equal(t, uint8(6), uint8(instr&0x1F))
		assert.Equal(t, uint8(1), uint8((instr>>5)&0xFF))
		assert.Equal(t, uint8(2), uint8((instr>>13)&0xFF))
		assert.Equal(t, uint8(3), uint8((instr>>21)&0xFF))
	})

	t.Run("Immediate encoding", func(t *testing.T) {
		instr := EncodeImmInstruction(2, 0x1234, 5)
		assert.Equal(t, uint8(2), uint8(instr&0x1F))
		assert.Equal(t, uint16(0x1234), uint16((instr>>5)&0xFFFF))
		assert.Equal(t, uint8(5), uint8((instr>>21)&0xFF))
	})
}
