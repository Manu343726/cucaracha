package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// MUX2 Tests
// =============================================================================

func TestMux2(t *testing.T) {
	t.Run("NewMux2 creates mux with correct ports", func(t *testing.T) {
		mux := NewMux2("Mux", 32)

		assert.Equal(t, "Mux", mux.Name())
		assert.Equal(t, "MUX2", mux.Type())
		assert.Equal(t, 32, mux.Width())
		assert.NotNil(t, mux.InputA())
		assert.NotNil(t, mux.InputB())
		assert.NotNil(t, mux.Select())
		assert.NotNil(t, mux.Output())
	})

	t.Run("Select=0 outputs A", func(t *testing.T) {
		mux := NewMux2("Mux", 32)
		mux.InputA().SetValue(0xAAAA)
		mux.InputB().SetValue(0xBBBB)
		mux.Select().Set(Low)

		mux.Compute()
		assert.Equal(t, uint64(0xAAAA), mux.Output().GetValue())
	})

	t.Run("Select=1 outputs B", func(t *testing.T) {
		mux := NewMux2("Mux", 32)
		mux.InputA().SetValue(0xAAAA)
		mux.InputB().SetValue(0xBBBB)
		mux.Select().Set(High)

		mux.Compute()
		assert.Equal(t, uint64(0xBBBB), mux.Output().GetValue())
	})

	t.Run("Registry", func(t *testing.T) {
		desc, err := Registry.Get("MUX2")
		require.NoError(t, err)
		assert.Equal(t, "MUX2", desc.Name)
	})
}

// =============================================================================
// MUX4 Tests
// =============================================================================

func TestMux4(t *testing.T) {
	t.Run("NewMux4 creates mux with correct ports", func(t *testing.T) {
		mux := NewMux4("Mux", 32)

		assert.Equal(t, "Mux", mux.Name())
		assert.Equal(t, "MUX4", mux.Type())
		assert.Equal(t, 32, mux.Width())
	})

	t.Run("Select=0 outputs A", func(t *testing.T) {
		mux := NewMux4("Mux", 32)
		mux.InputA().SetValue(0xAAAA)
		mux.InputB().SetValue(0xBBBB)
		mux.InputC().SetValue(0xCCCC)
		mux.InputD().SetValue(0xDDDD)
		mux.Select().SetValue(0)

		mux.Compute()
		assert.Equal(t, uint64(0xAAAA), mux.Output().GetValue())
	})

	t.Run("Select=1 outputs B", func(t *testing.T) {
		mux := NewMux4("Mux", 32)
		mux.InputA().SetValue(0xAAAA)
		mux.InputB().SetValue(0xBBBB)
		mux.InputC().SetValue(0xCCCC)
		mux.InputD().SetValue(0xDDDD)
		mux.Select().SetValue(1)

		mux.Compute()
		assert.Equal(t, uint64(0xBBBB), mux.Output().GetValue())
	})

	t.Run("Select=2 outputs C", func(t *testing.T) {
		mux := NewMux4("Mux", 32)
		mux.InputA().SetValue(0xAAAA)
		mux.InputB().SetValue(0xBBBB)
		mux.InputC().SetValue(0xCCCC)
		mux.InputD().SetValue(0xDDDD)
		mux.Select().SetValue(2)

		mux.Compute()
		assert.Equal(t, uint64(0xCCCC), mux.Output().GetValue())
	})

	t.Run("Select=3 outputs D", func(t *testing.T) {
		mux := NewMux4("Mux", 32)
		mux.InputA().SetValue(0xAAAA)
		mux.InputB().SetValue(0xBBBB)
		mux.InputC().SetValue(0xCCCC)
		mux.InputD().SetValue(0xDDDD)
		mux.Select().SetValue(3)

		mux.Compute()
		assert.Equal(t, uint64(0xDDDD), mux.Output().GetValue())
	})
}

// =============================================================================
// Program Counter Tests
// =============================================================================

func TestProgramCounter(t *testing.T) {
	t.Run("NewProgramCounter creates PC with correct ports", func(t *testing.T) {
		pc := NewProgramCounter("PC", 4, 0)

		assert.Equal(t, "PC", pc.Name())
		assert.Equal(t, "PC", pc.Type())
		assert.Equal(t, uint32(4), pc.Step())
		assert.Equal(t, uint32(0), pc.Value())
	})

	t.Run("Initial value", func(t *testing.T) {
		pc := NewProgramCounter("PC", 4, 0x1000)
		assert.Equal(t, uint32(0x1000), pc.Value())
	})

	t.Run("Increment", func(t *testing.T) {
		pc := NewProgramCounter("PC", 4, 0)
		pc.Increment()
		assert.Equal(t, uint32(4), pc.Value())
		pc.Increment()
		assert.Equal(t, uint32(8), pc.Value())
	})

	t.Run("Load", func(t *testing.T) {
		pc := NewProgramCounter("PC", 4, 0)
		pc.Load(0x1000)
		assert.Equal(t, uint32(0x1000), pc.Value())
	})

	t.Run("Clock with increment enable", func(t *testing.T) {
		pc := NewProgramCounter("PC", 4, 0)
		pc.IncrementEnable().Set(High)
		pc.Clock()
		assert.Equal(t, uint32(4), pc.Value())
	})

	t.Run("Clock with load enable", func(t *testing.T) {
		pc := NewProgramCounter("PC", 4, 0)
		pc.LoadValue().SetValue(0x2000)
		pc.LoadEnable().Set(High)
		pc.Clock()
		assert.Equal(t, uint32(0x2000), pc.Value())
	})

	t.Run("Load has priority over increment", func(t *testing.T) {
		pc := NewProgramCounter("PC", 4, 0)
		pc.LoadValue().SetValue(0x2000)
		pc.LoadEnable().Set(High)
		pc.IncrementEnable().Set(High)
		pc.Clock()
		assert.Equal(t, uint32(0x2000), pc.Value())
	})

	t.Run("Reset", func(t *testing.T) {
		pc := NewProgramCounter("PC", 4, 0x100)
		pc.Increment()
		pc.Increment()
		assert.Equal(t, uint32(0x108), pc.Value())
		pc.Reset()
		assert.Equal(t, uint32(0x100), pc.Value())
	})

	t.Run("Reset pin", func(t *testing.T) {
		pc := NewProgramCounter("PC", 4, 0x100)
		pc.Increment()
		pc.ResetPin().Set(High)
		pc.Clock()
		assert.Equal(t, uint32(0x100), pc.Value())
	})

	t.Run("Registry", func(t *testing.T) {
		desc, err := Registry.Get("PC")
		require.NoError(t, err)
		assert.Equal(t, "PC", desc.Name)
	})
}

// =============================================================================
// Instruction Decoder Tests
// =============================================================================

func TestInstructionDecoder(t *testing.T) {
	t.Run("NewInstructionDecoder creates decoder with correct ports", func(t *testing.T) {
		dec := NewInstructionDecoder("Decoder")

		assert.Equal(t, "Decoder", dec.Name())
		assert.Equal(t, "DECODER", dec.Type())
		assert.NotNil(t, dec.Instruction())
		assert.NotNil(t, dec.Opcode())
		assert.NotNil(t, dec.Op1())
		assert.NotNil(t, dec.Op2())
		assert.NotNil(t, dec.Op3())
		assert.NotNil(t, dec.Imm16())
	})

	t.Run("Decode opcode", func(t *testing.T) {
		dec := NewInstructionDecoder("Decoder")
		// Instruction with opcode 6 (ADD)
		instr := EncodeInstruction(6, 0, 0, 0)
		dec.Decode(instr)
		assert.Equal(t, uint8(6), dec.GetOpcode())
	})

	t.Run("Decode operands", func(t *testing.T) {
		dec := NewInstructionDecoder("Decoder")
		// ADD r3, r1, r2 (opcode=6, op1=1, op2=2, op3=3)
		instr := EncodeInstruction(6, 1, 2, 3)
		dec.Decode(instr)
		assert.Equal(t, uint8(6), dec.GetOpcode())
		assert.Equal(t, uint8(1), dec.GetOp1())
		assert.Equal(t, uint8(2), dec.GetOp2())
		assert.Equal(t, uint8(3), dec.GetOp3())
	})

	t.Run("Decode immediate", func(t *testing.T) {
		dec := NewInstructionDecoder("Decoder")
		// MOVIMM16L with immediate 0x1234
		instr := EncodeImmInstruction(2, 0x1234, 5)
		dec.Decode(instr)
		assert.Equal(t, uint8(2), dec.GetOpcode())
		assert.Equal(t, uint16(0x1234), dec.GetImm16())
	})

	t.Run("Registry", func(t *testing.T) {
		desc, err := Registry.Get("DECODER")
		require.NoError(t, err)
		assert.Equal(t, "DECODER", desc.Name)
	})
}

// =============================================================================
// Control Unit Tests
// =============================================================================

func TestControlUnit(t *testing.T) {
	t.Run("NewControlUnit creates control unit", func(t *testing.T) {
		cu := NewControlUnit("CU")

		assert.Equal(t, "CU", cu.Name())
		assert.Equal(t, "CONTROL_UNIT", cu.Type())
		assert.Equal(t, State_Fetch, cu.CurrentState())
		assert.False(t, cu.IsHalted())
	})

	t.Run("FSM transitions from Fetch to Decode", func(t *testing.T) {
		cu := NewControlUnit("CU")
		assert.Equal(t, State_Fetch, cu.CurrentState())

		cu.Clock()
		assert.Equal(t, State_Decode, cu.CurrentState())
	})

	t.Run("Halt", func(t *testing.T) {
		cu := NewControlUnit("CU")
		cu.Halt()
		assert.True(t, cu.IsHalted())
		assert.Equal(t, State_Halt, cu.CurrentState())
	})

	t.Run("Reset", func(t *testing.T) {
		cu := NewControlUnit("CU")
		cu.Clock() // Move to Decode
		cu.Halt()
		cu.Reset()
		assert.False(t, cu.IsHalted())
		assert.Equal(t, State_Fetch, cu.CurrentState())
	})

	t.Run("Registry", func(t *testing.T) {
		desc, err := Registry.Get("CONTROL_UNIT")
		require.NoError(t, err)
		assert.Equal(t, "CONTROL_UNIT", desc.Name)
	})
}
