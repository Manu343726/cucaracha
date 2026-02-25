package components

import (
	"fmt"
	"testing"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/registers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test Helpers - Instruction Encoding
// =============================================================================

// GPR encodes a general purpose register index (r0-r9) as a proper operand value
// that includes the register class encoding
func GPR(index uint8) uint32 {
	reg, err := registers.RegisterClasses.RegisterByName(fmt.Sprintf("r%d", index))
	if err != nil {
		panic(err)
	}
	return uint32(reg.Encode())
}

// EncodeInstruction creates an encoded instruction using the proper ISA encoding.
// This uses the mc/instructions package to ensure correct encoding.
func EncodeInstruction(opcode instructions.OpCode, operands ...uint32) uint32 {
	desc, err := instructions.Instructions.Instruction(opcode)
	if err != nil {
		panic(err)
	}

	operandValues := make([]uint64, len(operands))
	for i, v := range operands {
		operandValues[i] = uint64(v)
	}

	raw := &instructions.RawInstruction{
		Descriptor:    desc,
		OperandValues: operandValues,
	}
	return raw.Encode()
}

// Opcode constants for testing (using the actual instructions.OpCode values)
const (
	OP_NOP        = instructions.OpCode_NOP
	OP_MOV        = instructions.OpCode_MOV
	OP_MOV_IMM16L = instructions.OpCode_MOV_IMM16L
	OP_MOV_IMM16H = instructions.OpCode_MOV_IMM16H
	OP_ADD        = instructions.OpCode_ADD
	OP_SUB        = instructions.OpCode_SUB
	OP_MUL        = instructions.OpCode_MUL
	OP_DIV        = instructions.OpCode_DIV
	OP_MOD        = instructions.OpCode_MOD
	OP_CMP        = instructions.OpCode_CMP
	OP_JMP        = instructions.OpCode_JMP
	OP_CJMP       = instructions.OpCode_CJMP
	OP_LD         = instructions.OpCode_LD
	OP_ST         = instructions.OpCode_ST
	OP_LSL        = instructions.OpCode_LSL
	OP_LSR        = instructions.OpCode_LSR
	OP_ASL        = instructions.OpCode_ASL
	OP_ASR        = instructions.OpCode_ASR
)

func newCPUWithRAM(memorySize int) *CPU {
	cpu := NewCPU("CPU", memorySize, 256)
	cpu.AttachMemory(NewRAM("RAM", memorySize))
	return cpu
}

// EncodeImmInstruction is a helper for immediate instructions (MOVIMM16L/H)
func EncodeImmInstruction(opcode instructions.OpCode, imm16 uint16, dst uint8) uint32 {
	return EncodeInstruction(opcode, uint32(imm16), GPR(dst))
}

// =============================================================================
// CPU Creation Tests
// =============================================================================

func TestCPU(t *testing.T) {
	t.Run("NewCPU creates CPU with internal components", func(t *testing.T) {
		cpu := newCPUWithRAM(1024)

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
		cpu := newCPUWithRAM(1024)
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
		cpu := newCPUWithRAM(1024)

		cpu.WriteMemory(0, 0x12345678)
		assert.Equal(t, uint32(0x12345678), cpu.ReadMemory(0))

		cpu.WriteMemory(100, 0xDEADBEEF)
		assert.Equal(t, uint32(0xDEADBEEF), cpu.ReadMemory(100))
	})

	t.Run("LoadBinary", func(t *testing.T) {
		cpu := newCPUWithRAM(1024)
		data := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

		err := cpu.LoadBinary(data, 0x100)
		require.NoError(t, err)

		assert.Equal(t, uint32(0x04030201), cpu.ReadMemory(0x100))
		assert.Equal(t, uint32(0x08070605), cpu.ReadMemory(0x104))
		assert.Equal(t, uint32(0x100), cpu.GetPC())
	})

	t.Run("LoadProgram", func(t *testing.T) {
		cpu := newCPUWithRAM(1024)
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
	cpu := newCPUWithRAM(1024)

	// NOP instruction (no operands)
	program := []uint32{EncodeInstruction(OP_NOP)}
	cpu.LoadProgram(program, 0)

	err := cpu.StepInstruction()
	require.NoError(t, err)

	// PC should advance by 4
	assert.Equal(t, uint32(4), cpu.GetPC())
}

func TestCPUMovImm16L(t *testing.T) {
	cpu := newCPUWithRAM(1024)

	// MOVIMM16L r5, #0x1234
	program := []uint32{EncodeImmInstruction(OP_MOV_IMM16L, 0x1234, 5)}
	cpu.LoadProgram(program, 0)

	err := cpu.StepInstruction()
	require.NoError(t, err)

	assert.Equal(t, uint32(0x1234), cpu.GetRegister(5))
}

func TestCPUMovImm16H(t *testing.T) {
	cpu := newCPUWithRAM(1024)

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
	cpu := newCPUWithRAM(1024)

	cpu.SetRegister(1, 0x42)

	// MOV dst, src -> operands are (src, dst)
	program := []uint32{EncodeInstruction(OP_MOV, GPR(1), GPR(2))} // src=r1, dst=r2
	cpu.LoadProgram(program, 0)

	err := cpu.StepInstruction()
	require.NoError(t, err)

	assert.Equal(t, uint32(0x42), cpu.GetRegister(2))
}

func TestCPUAdd(t *testing.T) {
	cpu := newCPUWithRAM(1024)

	cpu.SetRegister(1, 10)
	cpu.SetRegister(2, 20)

	// ADD dst, src1, src2 -> operands are (src1, src2, dst)
	program := []uint32{EncodeInstruction(OP_ADD, GPR(1), GPR(2), GPR(3))}
	cpu.LoadProgram(program, 0)

	err := cpu.StepInstruction()
	require.NoError(t, err)

	assert.Equal(t, uint32(30), cpu.GetRegister(3))
}

func TestCPUSub(t *testing.T) {
	cpu := newCPUWithRAM(1024)

	cpu.SetRegister(1, 50)
	cpu.SetRegister(2, 20)

	// SUB dst, src1, src2 -> operands are (src1, src2, dst)
	program := []uint32{EncodeInstruction(OP_SUB, GPR(1), GPR(2), GPR(3))}
	cpu.LoadProgram(program, 0)

	err := cpu.StepInstruction()
	require.NoError(t, err)

	assert.Equal(t, uint32(30), cpu.GetRegister(3))
}

func TestCPUMul(t *testing.T) {
	cpu := newCPUWithRAM(1024)

	cpu.SetRegister(1, 6)
	cpu.SetRegister(2, 7)

	// MUL dst, src1, src2 -> operands are (src1, src2, dst)
	program := []uint32{EncodeInstruction(OP_MUL, GPR(1), GPR(2), GPR(3))}
	cpu.LoadProgram(program, 0)

	err := cpu.StepInstruction()
	require.NoError(t, err)

	assert.Equal(t, uint32(42), cpu.GetRegister(3))
}

func TestCPUDiv(t *testing.T) {
	cpu := newCPUWithRAM(1024)

	cpu.SetRegister(1, 100)
	cpu.SetRegister(2, 10)

	// DIV dst, src1, src2 -> operands are (src1, src2, dst)
	program := []uint32{EncodeInstruction(OP_DIV, GPR(1), GPR(2), GPR(3))}
	cpu.LoadProgram(program, 0)

	err := cpu.StepInstruction()
	require.NoError(t, err)

	assert.Equal(t, uint32(10), cpu.GetRegister(3))
}

func TestCPULoadStore(t *testing.T) {
	cpu := newCPUWithRAM(1024)

	// Write a value to memory
	cpu.WriteMemory(0x100, 0xDEADBEEF)

	// Set r1 = address
	cpu.SetRegister(1, 0x100)

	// LD dst, [addr] -> operands are (addr, dst)
	program := []uint32{EncodeInstruction(OP_LD, GPR(1), GPR(2))} // addr=r1, dst=r2
	cpu.LoadProgram(program, 0)

	err := cpu.StepInstruction()
	require.NoError(t, err)

	assert.Equal(t, uint32(0xDEADBEEF), cpu.GetRegister(2))
}

func TestCPUStore(t *testing.T) {
	cpu := newCPUWithRAM(1024)

	// Set value and address
	cpu.SetRegister(1, 0x12345678) // Value
	cpu.SetRegister(2, 0x200)      // Address

	// ST src, [addr] -> operands are (src, addr)
	program := []uint32{EncodeInstruction(OP_ST, GPR(1), GPR(2))} // src=r1, addr=r2
	cpu.LoadProgram(program, 0)

	err := cpu.StepInstruction()
	require.NoError(t, err)

	assert.Equal(t, uint32(0x12345678), cpu.ReadMemory(0x200))
}

func TestCPUJump(t *testing.T) {
	cpu := newCPUWithRAM(1024)

	// Set target address
	cpu.SetRegister(1, 0x100)

	// JMP target, link -> operands are (target, link)
	program := []uint32{EncodeInstruction(OP_JMP, GPR(1), GPR(3))} // target=r1, link=r3
	cpu.LoadProgram(program, 0)

	err := cpu.StepInstruction()
	require.NoError(t, err)

	// PC should be at target
	assert.Equal(t, uint32(0x100), cpu.GetPC())
	// Link register should have return address
	assert.Equal(t, uint32(4), cpu.GetRegister(3))
}

func TestCPUConditionalJump(t *testing.T) {
	// CJMP op1=condition, op2=target, op3=link
	// Condition codes: 0=EQ(Z=1), 1=NE(Z=0), 2=CS(C=1), 3=CC(C=0), etc.

	t.Run("CJMP takes branch when condition satisfied (EQ with Z=1)", func(t *testing.T) {
		cpu := newCPUWithRAM(1024)

		// Use r4 for condition, r5 for target, r6 for link (avoid cpsr at r2)
		cpu.SetRegister(4, 0)     // condition code EQ (test Z flag)
		cpu.SetRegister(5, 0x100) // target address
		cpu.SetRegister(6, 0)     // link register

		// Set CPSR (register 2) with Zero flag set
		cpu.registers.Set(2, uint64(FlagZero))

		// CJMP r4, r5, r6
		program := []uint32{EncodeInstruction(OP_CJMP, GPR(4), GPR(5), GPR(6))}
		cpu.LoadProgram(program, 0)

		err := cpu.StepInstruction()
		require.NoError(t, err)

		// Branch should be taken
		assert.Equal(t, uint32(0x100), cpu.GetPC(), "PC should jump to target")
		assert.Equal(t, uint32(4), cpu.GetRegister(6), "Link register should have return address")
	})

	t.Run("CJMP does not branch when condition not satisfied (EQ with Z=0)", func(t *testing.T) {
		cpu := newCPUWithRAM(1024)

		// Use r4 for condition, r5 for target, r6 for link
		cpu.SetRegister(4, 0)     // condition code EQ (test Z flag)
		cpu.SetRegister(5, 0x100) // target address
		cpu.SetRegister(6, 0)     // link register

		// Set CPSR (register 2) with Zero flag NOT set
		cpu.registers.Set(2, 0)

		// CJMP r4, r5, r6
		program := []uint32{EncodeInstruction(OP_CJMP, GPR(4), GPR(5), GPR(6))}
		cpu.LoadProgram(program, 0)

		err := cpu.StepInstruction()
		require.NoError(t, err)

		// Branch should NOT be taken, PC should just advance
		assert.Equal(t, uint32(4), cpu.GetPC(), "PC should advance to next instruction")
		assert.Equal(t, uint32(0), cpu.GetRegister(6), "Link register should be unchanged")
	})

	t.Run("CJMP NE condition (Z=0 means branch)", func(t *testing.T) {
		cpu := newCPUWithRAM(1024)

		cpu.SetRegister(4, 1)     // condition code NE (branch if Z=0)
		cpu.SetRegister(5, 0x200) // target address
		cpu.SetRegister(6, 0)     // link register

		// CPSR with Z=0 (no flags set)
		cpu.registers.Set(2, 0)

		program := []uint32{EncodeInstruction(OP_CJMP, GPR(4), GPR(5), GPR(6))}
		cpu.LoadProgram(program, 0)

		err := cpu.StepInstruction()
		require.NoError(t, err)

		// NE with Z=0 should branch
		assert.Equal(t, uint32(0x200), cpu.GetPC(), "PC should jump to target")
		assert.Equal(t, uint32(4), cpu.GetRegister(6), "Link register should have return address")
	})

	t.Run("CJMP GT condition (Z=0 and N=V)", func(t *testing.T) {
		cpu := newCPUWithRAM(1024)

		// GT: Z=0 AND N=V (condition code 12)
		cpu.SetRegister(4, 12)    // condition code GT
		cpu.SetRegister(5, 0x300) // target address
		cpu.SetRegister(6, 0)     // link register

		// CPSR with Z=0, N=0, V=0 (N=V satisfied, Z=0 satisfied)
		cpu.registers.Set(2, 0)

		program := []uint32{EncodeInstruction(OP_CJMP, GPR(4), GPR(5), GPR(6))}
		cpu.LoadProgram(program, 0)

		err := cpu.StepInstruction()
		require.NoError(t, err)

		// GT with Z=0 and N=V should branch
		assert.Equal(t, uint32(0x300), cpu.GetPC(), "PC should jump to target")
	})

	t.Run("CJMP loop pattern", func(t *testing.T) {
		cpu := newCPUWithRAM(1024)

		// Simulate a simple countdown loop using CMP to set flags:
		// NOTE: Register 2 is the implicit CPSR register, so avoid using it for data!
		// r3 = counter (starts at 3)
		// r4 = 1 (decrement value)
		// r5 = 0 (zero for comparison)
		// r6 = loop target address (0 = start of program)
		// r7 = condition code register (NE = 1)
		// r8 = link register
		// r9 = CPSR destination for CMP (though also writes to r2)
		// Loop: SUB r3, r4, r3; CMP r3, r5, r9; CJMP r7, r6, r8

		cpu.SetRegister(3, 3) // counter
		cpu.SetRegister(4, 1) // decrement
		cpu.SetRegister(5, 0) // zero for comparison
		cpu.SetRegister(6, 0) // loop target (address 0)
		cpu.SetRegister(7, 1) // condition NE (continue while Z=0)
		cpu.SetRegister(8, 0) // link (unused)
		cpu.SetRegister(9, 0) // CPSR destination

		program := []uint32{
			EncodeInstruction(OP_SUB, GPR(3), GPR(4), GPR(3)),  // r3 = r3 - r4
			EncodeInstruction(OP_CMP, GPR(3), GPR(5), GPR(9)),  // CMP r3, r5 -> sets flags in register 2 (CPSR)
			EncodeInstruction(OP_CJMP, GPR(7), GPR(6), GPR(8)), // if NE, jump to r6
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
		assert.Equal(t, uint32(0), cpu.GetRegister(3), "Counter should be 0")
	})
}

func TestCPUShifts(t *testing.T) {
	cpu := newCPUWithRAM(1024)

	t.Run("LSL", func(t *testing.T) {
		cpu.Reset()
		cpu.SetRegister(1, 1)
		cpu.SetRegister(2, 4)
		program := []uint32{EncodeInstruction(OP_LSL, GPR(1), GPR(2), GPR(3))}
		cpu.LoadProgram(program, 0)
		cpu.StepInstruction()
		assert.Equal(t, uint32(16), cpu.GetRegister(3))
	})

	t.Run("LSR", func(t *testing.T) {
		cpu.Reset()
		cpu.SetRegister(1, 16)
		cpu.SetRegister(2, 2)
		program := []uint32{EncodeInstruction(OP_LSR, GPR(1), GPR(2), GPR(3))}
		cpu.LoadProgram(program, 0)
		cpu.StepInstruction()
		assert.Equal(t, uint32(4), cpu.GetRegister(3))
	})
}

// =============================================================================
// CPU Program Execution Tests
// =============================================================================

func TestCPUMultipleInstructions(t *testing.T) {
	cpu := newCPUWithRAM(1024)

	// Simple program: r1 = 10, r2 = 20, r3 = r1 + r2
	program := []uint32{
		EncodeImmInstruction(OP_MOV_IMM16L, 10, 1),        // r1 = 10
		EncodeImmInstruction(OP_MOV_IMM16L, 20, 2),        // r2 = 20
		EncodeInstruction(OP_ADD, GPR(1), GPR(2), GPR(3)), // r3 = r1 + r2
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
	cpu := newCPUWithRAM(1024)

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
	cpu := newCPUWithRAM(1024)

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
	t.Run("NOP encoding", func(t *testing.T) {
		instr := EncodeInstruction(OP_NOP)
		// NOP is opcode 0
		assert.Equal(t, uint8(0), uint8(instr&0x1F))
	})

	t.Run("ADD encoding", func(t *testing.T) {
		// ADD dst, src1, src2 -> operands are (src1, src2, dst)
		instr := EncodeInstruction(OP_ADD, GPR(1), GPR(2), GPR(3))
		// Just verify it decodes without error
		assert.NotEqual(t, uint32(0), instr)
	})
}
