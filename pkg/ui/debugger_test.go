package ui

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStopReasonString(t *testing.T) {
	tests := []struct {
		sr       StopReason
		expected string
	}{
		{StopReasonNone, "none"},
		{StopReasonStep, "step"},
		{StopReasonBreakpoint, "breakpoint"},
		{StopReasonWatchpoint, "watchpoint"},
		{StopReasonHalt, "halt"},
		{StopReasonError, "error"},
		{StopReasonTermination, "termination"},
		{StopReasonMaxSteps, "maxSteps"},
		{StopReasonInterrupt, "interrupt"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.sr.String()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestStopReasonFromString(t *testing.T) {
	tests := []struct {
		str      string
		expected StopReason
		wantErr  bool
	}{
		{"none", StopReasonNone, false},
		{"step", StopReasonStep, false},
		{"breakpoint", StopReasonBreakpoint, false},
		{"watchpoint", StopReasonWatchpoint, false},
		{"halt", StopReasonHalt, false},
		{"error", StopReasonError, false},
		{"termination", StopReasonTermination, false},
		{"maxSteps", StopReasonMaxSteps, false},
		{"interrupt", StopReasonInterrupt, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			got, err := StopReasonFromString(tt.str)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestStopReasonJSON(t *testing.T) {
	tests := []StopReason{
		StopReasonNone,
		StopReasonStep,
		StopReasonBreakpoint,
		StopReasonWatchpoint,
		StopReasonHalt,
		StopReasonError,
		StopReasonTermination,
		StopReasonMaxSteps,
		StopReasonInterrupt,
	}

	for _, sr := range tests {
		t.Run(sr.String(), func(t *testing.T) {
			data, err := json.Marshal(sr)
			assert.NoError(t, err)

			var unmarshaled StopReason
			err = json.Unmarshal(data, &unmarshaled)
			assert.NoError(t, err)

			assert.Equal(t, sr, unmarshaled)
		})
	}
}

func TestDebuggerStatusString(t *testing.T) {
	tests := []struct {
		ds       DebuggerStatus
		expected string
	}{
		{DebuggerStatusNotReady_MissingProgram, "notReadyMissingProgram"},
		{DebuggerStatusNotReady_MissingRuntime, "notReadyMissingRuntime"},
		{DebuggerStatusNotReady_MissingSystemConfig, "notReadyMissingSystemConfig"},
		{DebuggerStatusIdle, "idle"},
		{DebuggerStatusRunning, "running"},
		{DebuggerStatusPaused, "paused"},
		{DebuggerStatusTerminated, "terminated"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.ds.String()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestDebuggerStatusFromString(t *testing.T) {
	tests := []struct {
		str      string
		expected DebuggerStatus
		wantErr  bool
	}{
		{"notReadyMissingProgram", DebuggerStatusNotReady_MissingProgram, false},
		{"notReadyMissingRuntime", DebuggerStatusNotReady_MissingRuntime, false},
		{"notReadyMissingSystemConfig", DebuggerStatusNotReady_MissingSystemConfig, false},
		{"idle", DebuggerStatusIdle, false},
		{"running", DebuggerStatusRunning, false},
		{"paused", DebuggerStatusPaused, false},
		{"terminated", DebuggerStatusTerminated, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			got, err := DebuggerStatusFromString(tt.str)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestDebuggerStatusJSON(t *testing.T) {
	tests := []DebuggerStatus{
		DebuggerStatusNotReady_MissingProgram,
		DebuggerStatusNotReady_MissingRuntime,
		DebuggerStatusNotReady_MissingSystemConfig,
		DebuggerStatusIdle,
		DebuggerStatusRunning,
		DebuggerStatusPaused,
		DebuggerStatusTerminated,
	}

	for _, ds := range tests {
		t.Run(ds.String(), func(t *testing.T) {
			data, err := json.Marshal(ds)
			assert.NoError(t, err)

			var unmarshaled DebuggerStatus
			err = json.Unmarshal(data, &unmarshaled)
			assert.NoError(t, err)

			assert.Equal(t, ds, unmarshaled)
		})
	}
}

func TestDebuggerEventTypeString(t *testing.T) {
	tests := []struct {
		et       DebuggerEventType
		expected string
	}{
		{DebuggerEventProgramLoaded, "programLoaded"},
		{DebuggerEventStepped, "stepped"},
		{DebuggerEventBreakpointHit, "breakpointHit"},
		{DebuggerEventWatchpointHit, "watchpointHit"},
		{DebuggerEventProgramTerminated, "programTerminated"},
		{DebuggerEventProgramHalted, "programHalted"},
		{DebuggerEventError, "error"},
		{DebuggerEventSourceLocationChanged, "sourceLocationChanged"},
		{DebuggerEventInterrupted, "interrupted"},
		{DebuggerEventLagging, "lagging"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.et.String()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestDebuggerEventTypeFromString(t *testing.T) {
	tests := []struct {
		str      string
		expected DebuggerEventType
		wantErr  bool
	}{
		{"programLoaded", DebuggerEventProgramLoaded, false},
		{"stepped", DebuggerEventStepped, false},
		{"breakpointHit", DebuggerEventBreakpointHit, false},
		{"watchpointHit", DebuggerEventWatchpointHit, false},
		{"programTerminated", DebuggerEventProgramTerminated, false},
		{"programHalted", DebuggerEventProgramHalted, false},
		{"error", DebuggerEventError, false},
		{"sourceLocationChanged", DebuggerEventSourceLocationChanged, false},
		{"interrupted", DebuggerEventInterrupted, false},
		{"lagging", DebuggerEventLagging, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			got, err := DebuggerEventTypeFromString(tt.str)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestDebuggerEventTypeJSON(t *testing.T) {
	tests := []DebuggerEventType{
		DebuggerEventProgramLoaded,
		DebuggerEventStepped,
		DebuggerEventBreakpointHit,
		DebuggerEventWatchpointHit,
		DebuggerEventProgramTerminated,
		DebuggerEventProgramHalted,
		DebuggerEventError,
		DebuggerEventSourceLocationChanged,
		DebuggerEventInterrupted,
		DebuggerEventLagging,
	}

	for _, et := range tests {
		t.Run(et.String(), func(t *testing.T) {
			data, err := json.Marshal(et)
			assert.NoError(t, err)

			var unmarshaled DebuggerEventType
			err = json.Unmarshal(data, &unmarshaled)
			assert.NoError(t, err)

			assert.Equal(t, et, unmarshaled)
		})
	}
}

func TestExecutionResultJSON(t *testing.T) {
	result := &ExecutionResult{
		StopReason: StopReasonBreakpoint,
		Error:      nil,
		Steps:      10,
	}

	data, err := json.Marshal(result)
	assert.NoError(t, err)

	var unmarshaled ExecutionResult
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, result.StopReason, unmarshaled.StopReason)
}

func TestDebuggerStateJSON(t *testing.T) {
	state := &DebuggerState{
		Status: DebuggerStatusRunning,
	}

	data, err := json.Marshal(state)
	assert.NoError(t, err)

	var unmarshaled DebuggerState
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, state.Status, unmarshaled.Status)
}

func TestDebuggerEventJSON(t *testing.T) {
	event := &DebuggerEvent{
		Type: DebuggerEventStepped,
		Result: &ExecutionResult{
			StopReason: StopReasonStep,
			Steps:      1,
		},
	}

	data, err := json.Marshal(event)
	assert.NoError(t, err)

	var unmarshaled DebuggerEvent
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, event.Type, unmarshaled.Type)
}

func TestVariableValueJSON(t *testing.T) {
	varValue := &VariableValue{
		Name:        "x",
		ValueString: "42",
		TypeName:    "int",
		Location:    "r0",
	}

	data, err := json.Marshal(varValue)
	assert.NoError(t, err)

	var unmarshaled VariableValue
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, varValue.Name, unmarshaled.Name)
	assert.Equal(t, varValue.ValueString, unmarshaled.ValueString)
}

func TestStackFrameJSON(t *testing.T) {
	funcName := "main"
	frame := &StackFrame{
		Function:       &funcName,
		SourceLocation: &SourceLocation{File: "main.go", Line: 10},
	}

	data, err := json.Marshal(frame)
	assert.NoError(t, err)

	var unmarshaled StackFrame
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	if frame.Function != nil {
		assert.NotNil(t, unmarshaled.Function)
		assert.Equal(t, *frame.Function, *unmarshaled.Function)
	} else {
		assert.Nil(t, unmarshaled.Function)
	}
	assert.Equal(t, frame.SourceLocation.File, unmarshaled.SourceLocation.File)
}

func TestInfoResultJSON(t *testing.T) {
	result := &InfoResult{
		DebuggerState: &DebuggerState{
			Status: DebuggerStatusIdle,
		},
	}

	data, err := json.Marshal(result)
	assert.NoError(t, err)

	var unmarshaled InfoResult
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, result.DebuggerState.Status, unmarshaled.DebuggerState.Status)
}

func TestStackResultJSON(t *testing.T) {
	result := &StackResult{
		StackFrames: []*StackFrame{
			{Function: func() *string { s := "main"; return &s }(), SourceLocation: &SourceLocation{File: "main.go", Line: 10}},
		},
	}

	data, err := json.Marshal(result)
	assert.NoError(t, err)

	var unmarshaled StackResult
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Len(t, unmarshaled.StackFrames, len(result.StackFrames))
}

func TestVarsResultJSON(t *testing.T) {
	result := &VarsResult{
		Variables: []*VariableValue{
			{Name: "x", ValueString: "42", TypeName: "int", Location: "r0"},
		},
	}

	data, err := json.Marshal(result)
	assert.NoError(t, err)

	var unmarshaled VarsResult
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Len(t, unmarshaled.Variables, len(result.Variables))
}
