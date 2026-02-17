package ui

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstructionOperandKindString(t *testing.T) {
	tests := []struct {
		kind     InstructionOperandKind
		expected string
	}{
		{OperandKindRegister, "register"},
		{OperandKindImmediate, "immediate"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.kind.String()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestInstructionOperandKindFromString(t *testing.T) {
	tests := []struct {
		str      string
		expected InstructionOperandKind
		wantErr  bool
	}{
		{"register", OperandKindRegister, false},
		{"immediate", OperandKindImmediate, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			got, err := InstructionOperandKindFromString(tt.str)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestInstructionOperandKindJSON(t *testing.T) {
	tests := []InstructionOperandKind{
		OperandKindRegister,
		OperandKindImmediate,
	}

	for _, kind := range tests {
		t.Run(kind.String(), func(t *testing.T) {
			data, err := json.Marshal(kind)
			assert.NoError(t, err)

			var unmarshaled InstructionOperandKind
			err = json.Unmarshal(data, &unmarshaled)
			assert.NoError(t, err)

			assert.Equal(t, kind, unmarshaled)
		})
	}
}

func TestBreakpointJSON(t *testing.T) {
	bp := &Breakpoint{
		ID:       1,
		Address:  0x1000,
		Enabled:  true,
		Location: &SourceLocation{File: "test.c", Line: 10},
	}

	data, err := json.Marshal(bp)
	assert.NoError(t, err)

	var unmarshaled Breakpoint
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, bp.ID, unmarshaled.ID)
	assert.Equal(t, bp.Address, unmarshaled.Address)
	assert.Equal(t, bp.Enabled, unmarshaled.Enabled)
}

func TestInstructionOperandJSON(t *testing.T) {
	operand := &InstructionOperand{
		Kind:     OperandKindRegister,
		Register: &Register{Name: "r0", Value: 0x100},
	}

	data, err := json.Marshal(operand)
	assert.NoError(t, err)

	var unmarshaled InstructionOperand
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, operand.Kind, unmarshaled.Kind)
}

func TestInstructionJSON(t *testing.T) {
	instr := &Instruction{
		Address:     0x1000,
		Encoding:    0x12345678,
		Mnemonic:    "ADD",
		Text:        "ADD r0, r1",
		IsCurrentPC: true,
		Operands: []*InstructionOperand{
			{Kind: OperandKindRegister, Register: &Register{Name: "r0"}},
		},
		Breakpoints: []*Breakpoint{},
		Watchpoints: []*Watchpoint{},
	}

	data, err := json.Marshal(instr)
	assert.NoError(t, err)

	var unmarshaled Instruction
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, instr.Address, unmarshaled.Address)
	assert.Equal(t, instr.Mnemonic, unmarshaled.Mnemonic)
}

func TestDisassemblyResultJSON(t *testing.T) {
	result := &DisassemblyResult{
		Instructions: []*Instruction{
			{
				Address:  0x1000,
				Encoding: 0x12345678,
				Mnemonic: "MOV",
				Text:     "MOV r0, #1",
			},
		},
	}

	data, err := json.Marshal(result)
	assert.NoError(t, err)

	var unmarshaled DisassemblyResult
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Len(t, unmarshaled.Instructions, len(result.Instructions))
}

func TestCurrentInstructionResultJSON(t *testing.T) {
	result := &CurrentInstructionResult{
		Instruction: &Instruction{
			Address:  0x2000,
			Mnemonic: "JMP",
			Text:     "JMP #0x3000",
		},
	}

	data, err := json.Marshal(result)
	assert.NoError(t, err)

	var unmarshaled CurrentInstructionResult
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.NotNil(t, unmarshaled.Instruction)
	assert.Equal(t, result.Instruction.Mnemonic, unmarshaled.Instruction.Mnemonic)
}
