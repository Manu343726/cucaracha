package ui

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSourceContextModeString(t *testing.T) {
	tests := []struct {
		mode     SourceContextMode
		expected string
	}{
		{SourceContextTop, "top"},
		{SourceContextCentered, "centered"},
		{SourceContextBottom, "bottom"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.mode.String()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestSourceContextModeFromString(t *testing.T) {
	tests := []struct {
		str      string
		expected SourceContextMode
		wantErr  bool
	}{
		{"top", SourceContextTop, false},
		{"centered", SourceContextCentered, false},
		{"bottom", SourceContextBottom, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			got, err := SourceContextModeFromString(tt.str)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestSourceContextModeJSON(t *testing.T) {
	tests := []SourceContextMode{
		SourceContextTop,
		SourceContextCentered,
		SourceContextBottom,
	}

	for _, mode := range tests {
		t.Run(mode.String(), func(t *testing.T) {
			data, err := json.Marshal(mode)
			assert.NoError(t, err)

			var unmarshaled SourceContextMode
			err = json.Unmarshal(data, &unmarshaled)
			assert.NoError(t, err)

			assert.Equal(t, mode, unmarshaled)
		})
	}
}

func TestBreakArgsJSON(t *testing.T) {
	args := &BreakArgs{
		SourceLocation: &SourceLocation{
			File: "main.go",
			Line: 42,
		},
	}

	data, err := json.Marshal(args)
	assert.NoError(t, err)

	var unmarshaled BreakArgs
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, args.SourceLocation.File, unmarshaled.SourceLocation.File)
	assert.Equal(t, args.SourceLocation.Line, unmarshaled.SourceLocation.Line)
}

func TestWatchArgsJSON(t *testing.T) {
	args := &WatchArgs{
		Range: &MemoryRegion{
			Name:       "stack",
			Start:      0x2000,
			Size:       0x100,
			RegionType: RegionStack,
		},
		Type: WatchpointTypeWrite,
	}

	data, err := json.Marshal(args)
	assert.NoError(t, err)

	var unmarshaled WatchArgs
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, args.Range.Name, unmarshaled.Range.Name)
	assert.Equal(t, args.Type, unmarshaled.Type)
}

func TestRemoveBreakpointArgsJSON(t *testing.T) {
	args := &RemoveBreakpointArgs{
		ID: 5,
	}

	data, err := json.Marshal(args)
	assert.NoError(t, err)

	var unmarshaled RemoveBreakpointArgs
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, args.ID, unmarshaled.ID)
}

func TestRemoveWatchpointArgsJSON(t *testing.T) {
	args := &RemoveWatchpointArgs{
		ID: 3,
	}

	data, err := json.Marshal(args)
	assert.NoError(t, err)

	var unmarshaled RemoveWatchpointArgs
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, args.ID, unmarshaled.ID)
}

func TestDisasmArgsJSON(t *testing.T) {
	args := &DisasmArgs{
		Address: 0x1000,
		Count:   20,
	}

	data, err := json.Marshal(args)
	assert.NoError(t, err)

	var unmarshaled DisasmArgs
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, args.Address, unmarshaled.Address)
	assert.Equal(t, args.Count, unmarshaled.Count)
}

func TestStepArgsJSON(t *testing.T) {
	args := &StepArgs{
		StepMode:  StepModeInto,
		CountMode: StepCountInstructions,
	}

	data, err := json.Marshal(args)
	assert.NoError(t, err)

	var unmarshaled StepArgs
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, args.StepMode, unmarshaled.StepMode)
	assert.Equal(t, args.CountMode, unmarshaled.CountMode)
}

func TestPrintArgsJSON(t *testing.T) {
	args := &PrintArgs{
		Expression: "x + y",
	}

	data, err := json.Marshal(args)
	assert.NoError(t, err)

	var unmarshaled PrintArgs
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, args.Expression, unmarshaled.Expression)
}

func TestSetArgsJSON(t *testing.T) {
	args := &SetArgs{
		Target: "x",
		Value:  10,
	}

	data, err := json.Marshal(args)
	assert.NoError(t, err)

	var unmarshaled SetArgs
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, args.Target, unmarshaled.Target)
	assert.Equal(t, args.Value, unmarshaled.Value)
}

func TestSourceArgsJSON(t *testing.T) {
	args := &SourceArgs{
		File:         "main.go",
		Line:         100,
		ContextLines: 10,
		ContextMode:  SourceContextCentered,
	}

	data, err := json.Marshal(args)
	assert.NoError(t, err)

	var unmarshaled SourceArgs
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, args.File, unmarshaled.File)
	assert.Equal(t, args.Line, unmarshaled.Line)
	assert.Equal(t, args.ContextMode, unmarshaled.ContextMode)
}

func TestEvalArgsJSON(t *testing.T) {
	args := &EvalArgs{
		Expression: "ptr->field",
	}

	data, err := json.Marshal(args)
	assert.NoError(t, err)

	var unmarshaled EvalArgs
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, args.Expression, unmarshaled.Expression)
}

func TestCurrentSourceArgsJSON(t *testing.T) {
	args := &CurrentSourceArgs{
		ContextMode: SourceContextTop,
	}

	data, err := json.Marshal(args)
	assert.NoError(t, err)

	var unmarshaled CurrentSourceArgs
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, args.ContextMode, unmarshaled.ContextMode)
}
