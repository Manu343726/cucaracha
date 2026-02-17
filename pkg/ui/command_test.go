package ui

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDebuggerCommandIdString(t *testing.T) {
	tests := []struct {
		cmdID    DebuggerCommandId
		expected string
	}{
		{DebuggerCommandLoadProgramFromFile, "loadProgramFromFile"},
		{DebuggerCommandLoadSystemFromFile, "loadSystemFromFile"},
		{DebuggerCommandLoadRuntime, "loadRuntime"},
		{DebuggerCommandStep, "step"},
		{DebuggerCommandContinue, "continue"},
		{DebuggerCommandInterrupt, "interrupt"},
		{DebuggerCommandBreak, "setBreakpoint"},
		{DebuggerCommandRemoveBreakpoint, "removeBreakpoint"},
		{DebuggerCommandWatch, "setWatchpoint"},
		{DebuggerCommandRemoveWatchpoint, "removeWatchpoint"},
		{DebuggerCommandList, "list"},
		{DebuggerCommandDisassemble, "disassemble"},
		{DebuggerCommandCurrentInstruction, "currentInstruction"},
		{DebuggerCommandMemory, "memory"},
		{DebuggerCommandSource, "source"},
		{DebuggerCommandCurrentSource, "currentSource"},
		{DebuggerCommandEvaluateExpression, "evaluateExpression"},
		{DebuggerCommandInfo, "info"},
		{DebuggerCommandRegisters, "registers"},
		{DebuggerCommandStack, "stack"},
		{DebuggerCommandVariables, "variables"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.cmdID.String()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestDebuggerCommandIdFromString(t *testing.T) {
	tests := []struct {
		str      string
		expected DebuggerCommandId
		wantErr  bool
	}{
		{"loadProgramFromFile", DebuggerCommandLoadProgramFromFile, false},
		{"loadSystemFromFile", DebuggerCommandLoadSystemFromFile, false},
		{"loadRuntime", DebuggerCommandLoadRuntime, false},
		{"step", DebuggerCommandStep, false},
		{"continue", DebuggerCommandContinue, false},
		{"interrupt", DebuggerCommandInterrupt, false},
		{"setBreakpoint", DebuggerCommandBreak, false},
		{"removeBreakpoint", DebuggerCommandRemoveBreakpoint, false},
		{"setWatchpoint", DebuggerCommandWatch, false},
		{"removeWatchpoint", DebuggerCommandRemoveWatchpoint, false},
		{"list", DebuggerCommandList, false},
		{"disassemble", DebuggerCommandDisassemble, false},
		{"currentInstruction", DebuggerCommandCurrentInstruction, false},
		{"memory", DebuggerCommandMemory, false},
		{"source", DebuggerCommandSource, false},
		{"currentSource", DebuggerCommandCurrentSource, false},
		{"evaluateExpression", DebuggerCommandEvaluateExpression, false},
		{"info", DebuggerCommandInfo, false},
		{"registers", DebuggerCommandRegisters, false},
		{"stack", DebuggerCommandStack, false},
		{"variables", DebuggerCommandVariables, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			got, err := DebuggerCommandIdFromString(tt.str)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestDebuggerCommandIdJSON(t *testing.T) {
	tests := []DebuggerCommandId{
		DebuggerCommandLoadProgramFromFile,
		DebuggerCommandLoadSystemFromFile,
		DebuggerCommandLoadRuntime,
		DebuggerCommandStep,
		DebuggerCommandContinue,
		DebuggerCommandInterrupt,
		DebuggerCommandBreak,
		DebuggerCommandRemoveBreakpoint,
		DebuggerCommandWatch,
		DebuggerCommandRemoveWatchpoint,
		DebuggerCommandList,
		DebuggerCommandDisassemble,
		DebuggerCommandCurrentInstruction,
		DebuggerCommandMemory,
		DebuggerCommandSource,
		DebuggerCommandCurrentSource,
		DebuggerCommandEvaluateExpression,
		DebuggerCommandInfo,
		DebuggerCommandRegisters,
		DebuggerCommandStack,
		DebuggerCommandVariables,
	}

	for _, cmdID := range tests {
		t.Run(cmdID.String(), func(t *testing.T) {
			data, err := json.Marshal(cmdID)
			assert.NoError(t, err)

			var unmarshaled DebuggerCommandId
			err = json.Unmarshal(data, &unmarshaled)
			assert.NoError(t, err)

			assert.Equal(t, cmdID, unmarshaled)
		})
	}
}

func TestStepModeString(t *testing.T) {
	tests := []struct {
		mode     StepMode
		expected string
	}{
		{StepModeInto, "into"},
		{StepModeOver, "over"},
		{StepModeOut, "out"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.mode.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStepModeFromString(t *testing.T) {
	tests := []struct {
		str      string
		expected StepMode
		wantErr  bool
	}{
		{"into", StepModeInto, false},
		{"over", StepModeOver, false},
		{"out", StepModeOut, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			got, err := StepModeFromString(tt.str)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestStepModeJSON(t *testing.T) {
	tests := []StepMode{
		StepModeInto,
		StepModeOver,
		StepModeOut,
	}

	for _, mode := range tests {
		t.Run(mode.String(), func(t *testing.T) {
			data, err := json.Marshal(mode)
			assert.NoError(t, err)

			var unmarshaled StepMode
			err = json.Unmarshal(data, &unmarshaled)
			assert.NoError(t, err)

			assert.Equal(t, mode, unmarshaled)
		})
	}
}

func TestStepCountModeString(t *testing.T) {
	tests := []struct {
		mode     StepCountMode
		expected string
	}{
		{StepCountInstructions, "instructions"},
		{StepCountSourceLines, "sourceLines"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.mode.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStepCountModeFromString(t *testing.T) {
	tests := []struct {
		str      string
		expected StepCountMode
		wantErr  bool
	}{
		{"instructions", StepCountInstructions, false},
		{"sourceLines", StepCountSourceLines, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			got, err := StepCountModeFromString(tt.str)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestStepCountModeJSON(t *testing.T) {
	tests := []StepCountMode{
		StepCountInstructions,
		StepCountSourceLines,
	}

	for _, mode := range tests {
		t.Run(mode.String(), func(t *testing.T) {
			data, err := json.Marshal(mode)
			assert.NoError(t, err)

			var unmarshaled StepCountMode
			err = json.Unmarshal(data, &unmarshaled)
			assert.NoError(t, err)

			assert.Equal(t, mode, unmarshaled)
		})
	}
}

func TestDebuggerCommandJSON(t *testing.T) {
	cmd := &DebuggerCommand{
		Command: DebuggerCommandStep,
	}

	data, err := json.Marshal(cmd)
	assert.NoError(t, err)

	var unmarshaled DebuggerCommand
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, cmd.Command, unmarshaled.Command)
}
