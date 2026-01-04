package interpreter

import (
	"testing"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/registers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// regIdx returns the encoded register index for a named register.
// This is the index to use when accessing CPUState.Registers[].
func regIdx(name string) uint32 {
	return uint32(registers.Register(name).Encode())
}

func TestNewCPUState(t *testing.T) {
	state := NewCPUState(1024)

	assert.NotNil(t, state)
	assert.Len(t, state.Memory, 1024)
	assert.Equal(t, uint32(0), state.PC)
	assert.False(t, state.Halted)

	// Stack pointer should be initialized to last valid word-aligned address
	assert.Equal(t, uint32(1020), *state.SP)

	// SP should alias the sp register
	assert.Equal(t, &state.Registers[regIdx("sp")], state.SP)

	// LR should alias the lr register
	assert.Equal(t, &state.Registers[regIdx("lr")], state.LR)
}

func TestCPUState_ReadWriteMemory32(t *testing.T) {
	state := NewCPUState(1024)

	t.Run("write and read", func(t *testing.T) {
		err := state.WriteMemory32(0x100, 0xDEADBEEF)
		require.NoError(t, err)

		value, err := state.ReadMemory32(0x100)
		require.NoError(t, err)
		assert.Equal(t, uint32(0xDEADBEEF), value)
	})

	t.Run("little endian", func(t *testing.T) {
		err := state.WriteMemory32(0x200, 0x04030201)
		require.NoError(t, err)

		// Verify little-endian byte order
		assert.Equal(t, byte(0x01), state.Memory[0x200])
		assert.Equal(t, byte(0x02), state.Memory[0x201])
		assert.Equal(t, byte(0x03), state.Memory[0x202])
		assert.Equal(t, byte(0x04), state.Memory[0x203])
	})

	t.Run("out of bounds read", func(t *testing.T) {
		_, err := state.ReadMemory32(1024) // At the boundary
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "out of bounds")
	})

	t.Run("out of bounds write", func(t *testing.T) {
		err := state.WriteMemory32(1024, 0x12345678)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "out of bounds")
	})
}

func TestCPUState_RegisterAccess(t *testing.T) {
	state := NewCPUState(1024)

	t.Run("get and set register", func(t *testing.T) {
		state.SetRegister(10, 0x12345678)
		assert.Equal(t, uint32(0x12345678), state.GetRegister(10))
	})

	t.Run("PC access", func(t *testing.T) {
		state.SetPC(0x1000)
		assert.Equal(t, uint32(0x1000), state.GetPC())
	})
}

func TestNewInterpreter(t *testing.T) {
	interp := NewInterpreter(4096)

	assert.NotNil(t, interp)
	assert.NotNil(t, interp.State())
	assert.Len(t, interp.State().Memory, 4096)
}

func TestInterpreter_LoadBinary(t *testing.T) {
	interp := NewInterpreter(1024)

	t.Run("load valid binary", func(t *testing.T) {
		binary := []byte{0x01, 0x02, 0x03, 0x04}
		err := interp.LoadBinary(binary, 0x100)
		require.NoError(t, err)

		// Check binary was loaded
		assert.Equal(t, byte(0x01), interp.State().Memory[0x100])
		assert.Equal(t, byte(0x02), interp.State().Memory[0x101])
		assert.Equal(t, byte(0x03), interp.State().Memory[0x102])
		assert.Equal(t, byte(0x04), interp.State().Memory[0x103])

		// PC should be set to load address
		assert.Equal(t, uint32(0x100), interp.State().PC)
		assert.False(t, interp.State().Halted)
	})

	t.Run("binary too large", func(t *testing.T) {
		binary := make([]byte, 2000)
		err := interp.LoadBinary(binary, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "too large")
	})
}

func TestInterpreter_LoadProgram(t *testing.T) {
	interp := NewInterpreter(4096)

	program := mc.NewProgram().Add(mc.Nop())

	err := interp.LoadProgram(program, 0)
	require.NoError(t, err)

	assert.Equal(t, uint32(0), interp.State().PC)
}

func TestInterpreter_Reset(t *testing.T) {
	interp := NewInterpreter(1024)

	// Modify state
	interp.State().Registers[regIdx("r0")] = 0x12345678
	interp.State().PC = 0x100
	interp.State().Halted = true

	interp.Reset()

	// State should be reset
	assert.Equal(t, uint32(0), interp.State().Registers[regIdx("r0")])
	assert.Equal(t, uint32(0), interp.State().PC)
	assert.False(t, interp.State().Halted)
	assert.Equal(t, uint32(1020), *interp.State().SP)
}

func TestInterpreter_Step_Halted(t *testing.T) {
	interp := NewInterpreter(1024)
	interp.State().Halted = true

	_, err := interp.Step()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "halted")
}

func TestInterpreter_Nop(t *testing.T) {
	interp := NewInterpreter(4096)

	program := mc.NewProgram().Add(mc.Nop())

	err := interp.LoadProgram(program, 0)
	require.NoError(t, err)

	_, err = interp.Step()
	require.NoError(t, err)

	// PC should advance by 4
	assert.Equal(t, uint32(4), interp.State().PC)
}

func TestInterpreter_MovImm16L(t *testing.T) {
	interp := NewInterpreter(4096)

	program := mc.NewProgram().Add(mc.MovImm16L(0x1234, "r0"))

	err := interp.LoadProgram(program, 0)
	require.NoError(t, err)

	_, err = interp.Step()
	require.NoError(t, err)

	assert.Equal(t, uint32(0x1234), interp.State().Registers[regIdx("r0")])
}

func TestInterpreter_MovImm16H(t *testing.T) {
	interp := NewInterpreter(4096)

	program := mc.NewProgram().
		Add(mc.MovImm16L(0x5678, "r0")). // r0 = 0x00005678
		Add(mc.MovImm16H(0x1234, "r0"))  // r0 = 0x12345678

	err := interp.LoadProgram(program, 0)
	require.NoError(t, err)

	err = interp.RunN(2)
	require.NoError(t, err)

	assert.Equal(t, uint32(0x12345678), interp.State().Registers[regIdx("r0")])
}

func TestInterpreter_Mov(t *testing.T) {
	interp := NewInterpreter(4096)

	program := mc.NewProgram().
		Add(mc.MovImm16L(0x42, "r0")).
		Add(mc.Mov("r0", "r1"))

	err := interp.LoadProgram(program, 0)
	require.NoError(t, err)

	err = interp.RunN(2)
	require.NoError(t, err)

	assert.Equal(t, uint32(0x42), interp.State().Registers[regIdx("r0")])
	assert.Equal(t, uint32(0x42), interp.State().Registers[regIdx("r1")])
}

func TestInterpreter_Add(t *testing.T) {
	interp := NewInterpreter(4096)

	program := mc.NewProgram().
		Add(mc.MovImm16L(10, "r0")).
		Add(mc.MovImm16L(20, "r1")).
		Add(mc.Add("r0", "r1", "r2"))

	err := interp.LoadProgram(program, 0)
	require.NoError(t, err)

	err = interp.RunN(3)
	require.NoError(t, err)

	assert.Equal(t, uint32(30), interp.State().Registers[regIdx("r2")])
}

func TestInterpreter_Sub(t *testing.T) {
	interp := NewInterpreter(4096)

	program := mc.NewProgram().
		Add(mc.MovImm16L(50, "r0")).
		Add(mc.MovImm16L(20, "r1")).
		Add(mc.Sub("r0", "r1", "r2"))

	err := interp.LoadProgram(program, 0)
	require.NoError(t, err)

	err = interp.RunN(3)
	require.NoError(t, err)

	assert.Equal(t, uint32(30), interp.State().Registers[regIdx("r2")])
}

func TestInterpreter_Mul(t *testing.T) {
	interp := NewInterpreter(4096)

	program := mc.NewProgram().
		Add(mc.MovImm16L(6, "r0")).
		Add(mc.MovImm16L(7, "r1")).
		Add(mc.Mul("r0", "r1", "r2"))

	err := interp.LoadProgram(program, 0)
	require.NoError(t, err)

	err = interp.RunN(3)
	require.NoError(t, err)

	assert.Equal(t, uint32(42), interp.State().Registers[regIdx("r2")])
}

func TestInterpreter_Div(t *testing.T) {
	interp := NewInterpreter(4096)

	program := mc.NewProgram().
		Add(mc.MovImm16L(100, "r0")).
		Add(mc.MovImm16L(10, "r1")).
		Add(mc.Div("r0", "r1", "r2"))

	err := interp.LoadProgram(program, 0)
	require.NoError(t, err)

	err = interp.RunN(3)
	require.NoError(t, err)

	assert.Equal(t, uint32(10), interp.State().Registers[regIdx("r2")])
}

func TestInterpreter_Mod(t *testing.T) {
	interp := NewInterpreter(4096)

	program := mc.NewProgram().
		Add(mc.MovImm16L(17, "r0")).
		Add(mc.MovImm16L(5, "r1")).
		Add(mc.Mod("r0", "r1", "r2"))

	err := interp.LoadProgram(program, 0)
	require.NoError(t, err)

	err = interp.RunN(3)
	require.NoError(t, err)

	assert.Equal(t, uint32(2), interp.State().Registers[regIdx("r2")])
}

func TestInterpreter_LoadStore(t *testing.T) {
	interp := NewInterpreter(8192) // Need more memory to access 0x1000

	program := mc.NewProgram().
		Add(mc.MovImm16L(0xBEEF, "r0")).
		Add(mc.MovImm16H(0xDEAD, "r0")).
		Add(mc.MovImm16L(0x1000, "r1")).
		Add(mc.MovImm16H(0x0000, "r1")).
		Add(mc.St("r0", "r1")).
		Add(mc.Ld("r1", "r2"))

	err := interp.LoadProgram(program, 0)
	require.NoError(t, err)

	err = interp.RunN(6)
	require.NoError(t, err)

	assert.Equal(t, uint32(0xDEADBEEF), interp.State().Registers[regIdx("r0")])
	assert.Equal(t, uint32(0x1000), interp.State().Registers[regIdx("r1")])
	assert.Equal(t, uint32(0xDEADBEEF), interp.State().Registers[regIdx("r2")])

	// Verify memory was written
	value, err := interp.State().ReadMemory32(0x1000)
	require.NoError(t, err)
	assert.Equal(t, uint32(0xDEADBEEF), value)
}

func TestInterpreter_Cmp(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		interp := NewInterpreter(4096)

		program := mc.NewProgram().
			Add(mc.MovImm16L(42, "r0")).
			Add(mc.MovImm16L(42, "r1")).
			Add(mc.Cmp("r0", "r1", "r2"))

		err := interp.LoadProgram(program, 0)
		require.NoError(t, err)

		err = interp.RunN(3)
		require.NoError(t, err)

		// Equal: Z=1 (zero flag) and C=1 (no borrow since lhs >= rhs)
		// flags = FLAG_Z | FLAG_C = 0x1 | 0x4 = 0x5
		cpsr := interp.State().Registers[regIdx("r2")]
		assert.True(t, cpsr&uint32(instructions.FLAG_Z) != 0, "Z flag should be set for equal values")
		assert.True(t, cpsr&uint32(instructions.FLAG_C) != 0, "C flag should be set (no borrow)")
	})

	t.Run("greater than", func(t *testing.T) {
		interp := NewInterpreter(4096)

		program := mc.NewProgram().
			Add(mc.MovImm16L(100, "r0")).
			Add(mc.MovImm16L(50, "r1")).
			Add(mc.Cmp("r0", "r1", "r2"))

		err := interp.LoadProgram(program, 0)
		require.NoError(t, err)

		err = interp.RunN(3)
		require.NoError(t, err)

		// Greater (100 > 50): Z=0 (not zero), C=1 (no borrow since lhs > rhs), N=0 (positive result)
		// flags = FLAG_C = 0x4
		cpsr := interp.State().Registers[regIdx("r2")]
		assert.True(t, cpsr&uint32(instructions.FLAG_Z) == 0, "Z flag should not be set")
		assert.True(t, cpsr&uint32(instructions.FLAG_C) != 0, "C flag should be set (no borrow)")
		assert.True(t, cpsr&uint32(instructions.FLAG_N) == 0, "N flag should not be set (positive result)")
	})

	t.Run("less than", func(t *testing.T) {
		interp := NewInterpreter(4096)

		program := mc.NewProgram().
			Add(mc.MovImm16L(50, "r0")).
			Add(mc.MovImm16L(100, "r1")).
			Add(mc.Cmp("r0", "r1", "r2"))

		err := interp.LoadProgram(program, 0)
		require.NoError(t, err)

		err = interp.RunN(3)
		require.NoError(t, err)

		// Less (50 < 100): Z=0, C=0 (borrow since lhs < rhs), N=1 (negative result for 50-100)
		// flags = FLAG_N = 0x2
		cpsr := interp.State().Registers[regIdx("r2")]
		assert.True(t, cpsr&uint32(instructions.FLAG_Z) == 0, "Z flag should not be set")
		assert.True(t, cpsr&uint32(instructions.FLAG_C) == 0, "C flag should not be set (borrow occurred)")
		assert.True(t, cpsr&uint32(instructions.FLAG_N) != 0, "N flag should be set (negative result)")
	})
}

func TestInterpreter_Jmp(t *testing.T) {
	interp := NewInterpreter(4096)

	program := mc.NewProgram().
		Add(mc.MovImm16L(0x100, "r0")).
		Add(mc.MovImm16H(0x0000, "r0")).
		Add(mc.Jmp("r0", "r5"))

	err := interp.LoadProgram(program, 0)
	require.NoError(t, err)

	err = interp.RunN(3)
	require.NoError(t, err)

	// PC should be at jump target
	assert.Equal(t, uint32(0x100), interp.State().PC)

	// Link register should have return address (instruction after jump)
	assert.Equal(t, uint32(12), interp.State().Registers[regIdx("r5")]) // 3 instructions * 4 bytes
}

func TestInterpreter_CJmp(t *testing.T) {
	t.Run("condition true - EQ", func(t *testing.T) {
		interp := NewInterpreter(4096)

		// CJMP branches if condition code is satisfied by CPSR flags
		// First we need to set CPSR flags via a CMP instruction
		// CMP 5, 5 sets Z flag (equal)
		// Then CJMP with condition code 0 (CC_EQ) should branch
		program := mc.NewProgram().
			Add(mc.MovImm16L(5, "r0")).                          // r0 = 5
			Add(mc.MovImm16L(5, "r2")).                          // r2 = 5
			Add(mc.Cmp("r0", "r2", "r3")).                       // CMP r0, r2 -> sets Z flag (equal)
			Add(mc.MovImm16L(uint16(instructions.CC_EQ), "r0")). // r0 = CC_EQ condition code
			Add(mc.MovImm16L(0x100, "r1")).                      // r1 = 0x100 (target address)
			Add(mc.MovImm16H(0x0000, "r1")).                     // r1 high bits
			Add(mc.CJmp("r0", "r1", "r5"))                       // if CC_EQ satisfied, jump to r1

		err := interp.LoadProgram(program, 0)
		require.NoError(t, err)

		err = interp.RunN(7)
		require.NoError(t, err)

		// Should have jumped because Z flag is set and condition CC_EQ is satisfied
		assert.Equal(t, uint32(0x100), interp.State().PC)
	})

	t.Run("condition false - EQ when not equal", func(t *testing.T) {
		interp := NewInterpreter(4096)

		// CJMP should NOT branch if condition code is not satisfied
		// CMP 5, 3 clears Z flag (not equal)
		// CJMP with condition code 0 (CC_EQ) should NOT branch
		program := mc.NewProgram().
			Add(mc.MovImm16L(5, "r0")).                          // r0 = 5
			Add(mc.MovImm16L(3, "r2")).                          // r2 = 3
			Add(mc.Cmp("r0", "r2", "r3")).                       // CMP r0, r2 -> clears Z flag (not equal)
			Add(mc.MovImm16L(uint16(instructions.CC_EQ), "r0")). // r0 = CC_EQ condition code
			Add(mc.MovImm16L(0x100, "r1")).                      // r1 = 0x100 (target address)
			Add(mc.MovImm16H(0x0000, "r1")).                     // r1 high bits
			Add(mc.CJmp("r0", "r1", "r5"))                       // if CC_EQ satisfied, jump to r1

		err := interp.LoadProgram(program, 0)
		require.NoError(t, err)

		err = interp.RunN(7)
		require.NoError(t, err)

		// Should NOT have jumped because Z flag is clear (CC_EQ not satisfied)
		assert.Equal(t, uint32(28), interp.State().PC) // 7 instructions * 4 bytes
	})

	t.Run("condition true - NE", func(t *testing.T) {
		interp := NewInterpreter(4096)

		// CMP 5, 3 clears Z flag (not equal)
		// CJMP with condition code 1 (CC_NE) should branch
		program := mc.NewProgram().
			Add(mc.MovImm16L(5, "r0")).                          // r0 = 5
			Add(mc.MovImm16L(3, "r2")).                          // r2 = 3
			Add(mc.Cmp("r0", "r2", "r3")).                       // CMP r0, r2 -> clears Z flag (not equal)
			Add(mc.MovImm16L(uint16(instructions.CC_NE), "r0")). // r0 = CC_NE condition code
			Add(mc.MovImm16L(0x100, "r1")).                      // r1 = 0x100 (target address)
			Add(mc.MovImm16H(0x0000, "r1")).                     // r1 high bits
			Add(mc.CJmp("r0", "r1", "r5"))                       // if CC_NE satisfied, jump to r1

		err := interp.LoadProgram(program, 0)
		require.NoError(t, err)

		err = interp.RunN(7)
		require.NoError(t, err)

		// Should have jumped because Z flag is clear (CC_NE satisfied)
		assert.Equal(t, uint32(0x100), interp.State().PC)
	})

	t.Run("condition true - GT", func(t *testing.T) {
		interp := NewInterpreter(4096)

		// CMP 10, 5 should set: Z=0, N=0, C=1, V=0
		// GT requires Z=0 AND N=V, which is satisfied (both N and V are 0)
		program := mc.NewProgram().
			Add(mc.MovImm16L(10, "r0")).                         // r0 = 10
			Add(mc.MovImm16L(5, "r2")).                          // r2 = 5
			Add(mc.Cmp("r0", "r2", "r3")).                       // CMP 10, 5 -> Z=0, N=0, C=1, V=0
			Add(mc.MovImm16L(uint16(instructions.CC_GT), "r0")). // r0 = CC_GT condition code
			Add(mc.MovImm16L(0x100, "r1")).                      // r1 = 0x100 (target address)
			Add(mc.MovImm16H(0x0000, "r1")).                     // r1 high bits
			Add(mc.CJmp("r0", "r1", "r5"))                       // if CC_GT satisfied, jump to r1

		err := interp.LoadProgram(program, 0)
		require.NoError(t, err)

		err = interp.RunN(7)
		require.NoError(t, err)

		// Should have jumped because GT condition is satisfied (Z=0, N=V)
		assert.Equal(t, uint32(0x100), interp.State().PC)
	})
}

func TestInterpreter_ShiftOperations(t *testing.T) {
	t.Run("LSL", func(t *testing.T) {
		interp := NewInterpreter(4096)

		program := mc.NewProgram().
			Add(mc.MovImm16L(1, "r0")).
			Add(mc.MovImm16L(4, "r1")).
			Add(mc.Lsl("r0", "r1", "r2"))

		err := interp.LoadProgram(program, 0)
		require.NoError(t, err)

		err = interp.RunN(3)
		require.NoError(t, err)

		assert.Equal(t, uint32(16), interp.State().Registers[regIdx("r2")]) // 1 << 4 = 16
	})

	t.Run("LSR", func(t *testing.T) {
		interp := NewInterpreter(4096)

		program := mc.NewProgram().
			Add(mc.MovImm16L(64, "r0")).
			Add(mc.MovImm16L(2, "r1")).
			Add(mc.Lsr("r0", "r1", "r2"))

		err := interp.LoadProgram(program, 0)
		require.NoError(t, err)

		err = interp.RunN(3)
		require.NoError(t, err)

		assert.Equal(t, uint32(16), interp.State().Registers[regIdx("r2")]) // 64 >> 2 = 16
	})

	t.Run("ASR negative", func(t *testing.T) {
		interp := NewInterpreter(4096)

		program := mc.NewProgram().
			Add(mc.MovImm16L(0xFFF0, "r0")).
			Add(mc.MovImm16H(0xFFFF, "r0")).
			Add(mc.MovImm16L(2, "r1")).
			Add(mc.Asr("r0", "r1", "r2"))

		err := interp.LoadProgram(program, 0)
		require.NoError(t, err)

		err = interp.RunN(4)
		require.NoError(t, err)

		// -16 >> 2 = -4 (arithmetic shift preserves sign)
		assert.Equal(t, uint32(0xFFFFFFFC), interp.State().Registers[regIdx("r2")])
	})
}

func TestInterpreter_SimpleLoop(t *testing.T) {
	// Test a simple loop that sums numbers 1 to 5
	// Loop continues while counter >= 1
	// After CMP counter, 1:
	//   - When counter >= 1: C=1 (carry set for unsigned >=)
	//   - When counter < 1 (counter = 0): C=0 (borrow/no carry)
	// We use condition code CC_CS (2) which tests C=1 (carry set)
	// So we branch on CS condition while counter >= 1
	interp := NewInterpreter(4096)

	program := mc.NewProgram().
		Add(mc.MovImm16L(5, "r0")).                          // 0: r0 = counter = 5
		Add(mc.MovImm16L(0, "r1")).                          // 4: r1 = sum = 0
		Add(mc.MovImm16L(1, "r2")).                          // 8: r2 = 1 (decrement value and comparison target)
		Add(mc.Add("r1", "r0", "r1")).                       // 12: loop start - sum += counter
		Add(mc.Sub("r0", "r2", "r0")).                       // 16: counter -= 1
		Add(mc.Cmp("r0", "r2", "r4")).                       // 20: compare counter with 1 (sets C=1 if counter >= 1)
		Add(mc.MovImm16L(12, "r3")).                         // 24: r3 = 12 (loop start address)
		Add(mc.MovImm16L(uint16(instructions.CC_CS), "r5")). // 28: r5 = CC_CS condition code (carry set)
		Add(mc.CJmp("r5", "r3", "r6"))                       // 32: if CS condition satisfied (counter >= 1), continue loop

	err := interp.LoadProgram(program, 0)
	require.NoError(t, err)

	// Run enough instructions to complete the loop (5 iterations + setup)
	// Setup: 3 instructions (0,4,8)
	// Each loop iteration: 6 instructions (12,16,20,24,28,32)
	// Total: 3 + 5*6 = 33 instructions, but last iteration doesn't loop back
	// Run 100 to be safe
	err = interp.RunN(100)
	require.NoError(t, err)

	// Sum should be 5+4+3+2+1 = 15
	assert.Equal(t, uint32(15), interp.State().Registers[regIdx("r1")])
	// Counter should be 0 (decremented from 1 to 0 which failed the >= 1 test)
	assert.Equal(t, uint32(0), interp.State().Registers[regIdx("r0")])
}

func TestRegisterName(t *testing.T) {
	// Test that RegisterName correctly resolves register names
	assert.Equal(t, "r0", RegisterName(regIdx("r0")))
	assert.Equal(t, "r1", RegisterName(regIdx("r1")))
	assert.Equal(t, "r2", RegisterName(regIdx("r2")))
	assert.Equal(t, "r9", RegisterName(regIdx("r9")))
	assert.Equal(t, "pc", RegisterName(regIdx("pc")))
	assert.Equal(t, "sp", RegisterName(regIdx("sp")))
	assert.Equal(t, "cpsr", RegisterName(regIdx("cpsr")))
	assert.Equal(t, "lr", RegisterName(regIdx("lr")))
}

func TestInterpreter_TargetSpeed(t *testing.T) {
	interp := NewInterpreter(4096)

	// Default should be unlimited (0)
	assert.Equal(t, float64(0), interp.GetTargetSpeed())

	// Set target speed
	interp.SetTargetSpeed(1000)
	assert.Equal(t, float64(1000), interp.GetTargetSpeed())

	// Reset to unlimited
	interp.SetTargetSpeed(0)
	assert.Equal(t, float64(0), interp.GetTargetSpeed())

	// Negative values should be treated as 0
	interp.SetTargetSpeed(-100)
	assert.Equal(t, float64(0), interp.GetTargetSpeed())
}

func TestInterpreter_ExecutionDelayBackwardCompat(t *testing.T) {
	interp := NewInterpreter(4096)

	// Test backward compatibility with SetExecutionDelay
	interp.SetExecutionDelay(100) // 100ms delay = 10 Hz
	// The conversion is approximate
	assert.InDelta(t, float64(10), interp.GetTargetSpeed(), 0.1)

	// GetExecutionDelay should return approximate ms
	delay := interp.GetExecutionDelay()
	assert.InDelta(t, 100, delay, 1)

	// 0 delay means unlimited
	interp.SetExecutionDelay(0)
	assert.Equal(t, float64(0), interp.GetTargetSpeed())
	assert.Equal(t, 0, interp.GetExecutionDelay())
}

func TestInterpreter_StepReturnsCycles(t *testing.T) {
	interp := NewInterpreter(4096)

	program := mc.NewProgram().
		Add(mc.Nop()).
		Add(mc.MovImm16L(0x1234, "r0")).
		Add(mc.Add("r0", "r1", "r2"))

	err := interp.LoadProgram(program, 0)
	require.NoError(t, err)

	// Execute NOP - should return 1 cycle (default)
	result, err := interp.Step()
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.Cycles)
	assert.NotNil(t, result.Instruction)
	assert.Equal(t, "NOP", result.Instruction.OpCode.Mnemonic)

	// Execute MOVIMM16L - should return 1 cycle (default)
	result, err = interp.Step()
	require.NoError(t, err)
	assert.Equal(t, 1, result.Cycles)

	// Execute ADD - should return 1 cycle (default)
	result, err = interp.Step()
	require.NoError(t, err)
	assert.Equal(t, 1, result.Cycles)
}

func TestDebugger_TargetSpeed(t *testing.T) {
	interp := NewInterpreter(4096)
	dbg := NewDebugger(interp)

	// Default should be unlimited (0)
	assert.Equal(t, float64(0), dbg.GetTargetSpeed())

	// Set target speed
	dbg.SetTargetSpeed(1000)
	assert.Equal(t, float64(1000), dbg.GetTargetSpeed())

	// Should also be reflected in interpreter
	assert.Equal(t, float64(1000), interp.GetTargetSpeed())
}

func TestDebugger_ExecutionResultCycles(t *testing.T) {
	interp := NewInterpreter(4096)
	dbg := NewDebugger(interp)

	program := mc.NewProgram().
		Add(mc.Nop()).
		Add(mc.Nop()).
		Add(mc.Nop())

	err := interp.LoadProgram(program, 0)
	require.NoError(t, err)

	// Run 3 NOPs
	result := dbg.Run(3)

	assert.Equal(t, 3, result.StepsExecuted)
	assert.Equal(t, int64(3), result.CyclesExecuted) // 3 NOPs * 1 cycle each
}

func TestDebugger_LaggingEventType(t *testing.T) {
	// Verify the EventLagging constant exists and has expected string
	assert.Equal(t, "lagging", EventLagging.String())
}
