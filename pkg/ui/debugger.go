package ui

import (
	"encoding/json"
	"fmt"
)

// Indicates the reason why the execution stopped
type StopReason int

const (
	StopReasonNone StopReason = iota
	StopReasonStep
	StopReasonBreakpoint
	StopReasonWatchpoint
	StopReasonHalt
	StopReasonError
	StopReasonTermination
	StopReasonMaxSteps
	StopReasonInterrupt
)

func (r StopReason) String() string {
	switch r {
	case StopReasonNone:
		return "none"
	case StopReasonStep:
		return "step"
	case StopReasonBreakpoint:
		return "breakpoint"
	case StopReasonWatchpoint:
		return "watchpoint"
	case StopReasonHalt:
		return "halt"
	case StopReasonError:
		return "error"
	case StopReasonTermination:
		return "termination"
	case StopReasonMaxSteps:
		return "maxSteps"
	case StopReasonInterrupt:
		return "interrupt"
	default:
		return "unknown"
	}
}

func StopReasonFromString(s string) (StopReason, error) {
	switch s {
	case "none":
		return StopReasonNone, nil
	case "step":
		return StopReasonStep, nil
	case "breakpoint":
		return StopReasonBreakpoint, nil
	case "watchpoint":
		return StopReasonWatchpoint, nil
	case "halt":
		return StopReasonHalt, nil
	case "error":
		return StopReasonError, nil
	case "termination":
		return StopReasonTermination, nil
	case "maxSteps":
		return StopReasonMaxSteps, nil
	case "interrupt":
		return StopReasonInterrupt, nil
	default:
		return 0, fmt.Errorf("unknown StopReason: \"%s\"", s)
	}
}

func (r StopReason) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

func (r *StopReason) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	val, err := StopReasonFromString(s)
	if err != nil {
		return err
	}
	*r = val
	return nil
}

// Contains general information about an execution step in the debugger
type ExecutionResult struct {
	StopReason      StopReason      `json:"stopReason"`      // Reason why execution stopped, StopReasonNone if still running
	Error           error           `json:"error"`           // Error that occurred during execution, if any
	Steps           uint64          `json:"steps"`           // Number of steps executed
	Cycles          uint64          `json:"cycles"`          // Number of cycles executed
	Breakpoint      *Breakpoint     `json:"breakpoint"`      // Hit breakpoint, if any
	Watchpoint      *Watchpoint     `json:"watchpoint"`      // Hit watchpoint, if any
	LastInstruction uint32          `json:"lastInstruction"` // Address of the last executed instruction
	LastLocation    *SourceLocation `json:"lastLocation"`    // Source location of the last executed instruction
	LaggingCycles   uint32          `json:"laggingCycles"`   // Number of cycles the CPU is lagging behind expected timing
}

// Represents the status of the debugger
type DebuggerStatus int

const (
	// The debugger is not ready, no program is loaded
	DebuggerStatusNotReady_MissingProgram DebuggerStatus = iota
	// The debugger is not ready, no runtime is loaded
	DebuggerStatusNotReady_MissingRuntime
	// The debugger is not ready, no system config is loaded
	DebuggerStatusNotReady_MissingSystemConfig
	// The debugged program has not been started yet
	DebuggerStatusIdle
	// The debugged program is running
	DebuggerStatusRunning
	// Debugger is waiting to continue program execution
	DebuggerStatusPaused
	// The debugged program has finished execution
	DebuggerStatusTerminated
)

func (d DebuggerStatus) String() string {
	switch d {
	case DebuggerStatusNotReady_MissingProgram:
		return "notReadyMissingProgram"
	case DebuggerStatusNotReady_MissingRuntime:
		return "notReadyMissingRuntime"
	case DebuggerStatusNotReady_MissingSystemConfig:
		return "notReadyMissingSystemConfig"
	case DebuggerStatusIdle:
		return "idle"
	case DebuggerStatusRunning:
		return "running"
	case DebuggerStatusPaused:
		return "paused"
	case DebuggerStatusTerminated:
		return "terminated"
	default:
		return "unknown"
	}
}

func DebuggerStatusFromString(s string) (DebuggerStatus, error) {
	switch s {
	case "notReadyMissingProgram":
		return DebuggerStatusNotReady_MissingProgram, nil
	case "notReadyMissingRuntime":
		return DebuggerStatusNotReady_MissingRuntime, nil
	case "notReadyMissingSystemConfig":
		return DebuggerStatusNotReady_MissingSystemConfig, nil
	case "idle":
		return DebuggerStatusIdle, nil
	case "running":
		return DebuggerStatusRunning, nil
	case "paused":
		return DebuggerStatusPaused, nil
	case "terminated":
		return DebuggerStatusTerminated, nil
	default:
		return 0, fmt.Errorf("unknown DebuggerStatus: \"%s\"", s)
	}
}

func (d DebuggerStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *DebuggerStatus) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	val, err := DebuggerStatusFromString(s)
	if err != nil {
		return err
	}
	*d = val
	return nil
}

// DebuggerState contains a snapshot of the debugger state
type DebuggerState struct {
	Status    DebuggerStatus       `json:"status"`    // Current execution status of the debugged program
	Registers map[string]*Register `json:"registers"` // Current CPU registers
	Flags     *FlagState           `json:"flags"`     // Current CPU flags
}

// Specified the different types of events a debugger can report
type DebuggerEventType int

const (
	// DebuggerEventProgramLoaded is fired when a program is loaded
	DebuggerEventProgramLoaded DebuggerEventType = iota
	// DebuggerEventStepped is fired when a step operation completes
	DebuggerEventStepped
	// DebuggerEventBreakpointHit is fired when a breakpoint is hit
	DebuggerEventBreakpointHit
	// DebuggerEventWatchpointHit is fired when a watchpoint is hit
	DebuggerEventWatchpointHit
	// DebuggerEventProgramTerminated is fired when the program terminates
	DebuggerEventProgramTerminated
	// DebuggerEventProgramHalted is fired when the program is halted
	DebuggerEventProgramHalted
	// DebuggerEventError is fired when an error occurs
	DebuggerEventError
	// DebuggerEventSourceLocationChanged is fired when the source location changes
	DebuggerEventSourceLocationChanged
	// DebuggerEventInterrupted is fired when execution is interrupted by user (Ctrl+C)
	DebuggerEventInterrupted
	// DebuggerEventLagging is fired when the emulator can't keep up with target execution speed
	DebuggerEventLagging
)

func (e DebuggerEventType) String() string {
	switch e {
	case DebuggerEventProgramLoaded:
		return "programLoaded"
	case DebuggerEventStepped:
		return "stepped"
	case DebuggerEventBreakpointHit:
		return "breakpointHit"
	case DebuggerEventWatchpointHit:
		return "watchpointHit"
	case DebuggerEventProgramTerminated:
		return "programTerminated"
	case DebuggerEventProgramHalted:
		return "programHalted"
	case DebuggerEventError:
		return "error"
	case DebuggerEventSourceLocationChanged:
		return "sourceLocationChanged"
	case DebuggerEventInterrupted:
		return "interrupted"
	case DebuggerEventLagging:
		return "lagging"
	default:
		return "unknown"
	}
}

func DebuggerEventTypeFromString(s string) (DebuggerEventType, error) {
	switch s {
	case "programLoaded":
		return DebuggerEventProgramLoaded, nil
	case "stepped":
		return DebuggerEventStepped, nil
	case "breakpointHit":
		return DebuggerEventBreakpointHit, nil
	case "watchpointHit":
		return DebuggerEventWatchpointHit, nil
	case "programTerminated":
		return DebuggerEventProgramTerminated, nil
	case "programHalted":
		return DebuggerEventProgramHalted, nil
	case "error":
		return DebuggerEventError, nil
	case "sourceLocationChanged":
		return DebuggerEventSourceLocationChanged, nil
	case "interrupted":
		return DebuggerEventInterrupted, nil
	case "lagging":
		return DebuggerEventLagging, nil
	default:
		return 0, fmt.Errorf("unknown DebuggerEventType: \"%s\"", s)
	}
}

func (e DebuggerEventType) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

func (e *DebuggerEventType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	val, err := DebuggerEventTypeFromString(s)
	if err != nil {
		return err
	}
	*e = val
	return nil
}

// Represents an event sent by the debugger to the UI
type DebuggerEvent struct {
	// Type of event
	Type DebuggerEventType
	// Associated execution result, if applicable
	Result *ExecutionResult
}

// Callback function type for debugger events
type DebuggerEventCallback func(event *DebuggerEvent)

// Contains the current value of a program variable
type VariableValue struct {
	Name           string        `json:"name"`           // Variable name
	TypeName       string        `json:"typeName"`       // Variable type name
	ValueString    string        `json:"valueString"`    // Formatted value string for display (includes "<optimized out>" etc.)
	Location       string        `json:"location"`       // Human-readable location (e.g., "[sp+16]", "r0", "<optimized out>")
	Size           int           `json:"size"`           // Size in bytes
	MemoryLocation *MemoryRegion `json:"memoryLocation"` // Memory region where the variable is located (if applicable)
}

// Represents a stack frame
type StackFrame struct {
	SourceLocation *SourceLocation `json:"sourceLocation"` // Source code location (nil if unknown)
	Function       *string         `json:"function"`       // Name of the function (nil if unknown)
	Memory         *MemoryRegion   `json:"memory"`         // Memory region of the stack frame
}

// CommandHelp contains help information for a command
type CommandHelp struct {
	Name        string   `json:"name"`        // Command name
	Aliases     []string `json:"aliases"`     // Command aliases
	Description string   `json:"description"` // Command description
	Usage       string   `json:"usage"`       // Command usage
	Examples    []string `json:"examples"`    // Usage examples
}

// Result of Info command
type InfoResult struct {
	Error         error          `json:"error"`         // Error, if any
	DebuggerState *DebuggerState `json:"debuggerState"` // Debugger state
}

// Result of Stack command
type StackResult struct {
	Error       error         `json:"error"`       // Error, if any
	SP          uint32        `json:"sp"`          // Current stack pointer (Current top of stack)
	StackData   []byte        `json:"stackData"`   // Full stack data
	StackFrames []*StackFrame `json:"stackFrames"` // Call stack frames
}

// Result of Vars command
type VarsResult struct {
	Error     error            `json:"error"`     // Error, if any
	Variables []*VariableValue `json:"variables"` // Variable values
}
