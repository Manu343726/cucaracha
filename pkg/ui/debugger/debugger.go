package debugger

import (
	"encoding/json"
	"fmt"
)

// StopReason indicates why execution of the debugged program stopped or paused.
type StopReason int

const (
	// StopReasonNone indicates execution is still running.
	StopReasonNone StopReason = iota
	// StopReasonStep indicates execution stopped after completing a step instruction.
	StopReasonStep
	// StopReasonBreakpoint indicates a code breakpoint was hit.
	StopReasonBreakpoint
	// StopReasonWatchpoint indicates a data watchpoint was hit.
	StopReasonWatchpoint
	// StopReasonHalt indicates the program executed a halt instruction.
	StopReasonHalt
	// StopReasonError indicates execution stopped due to a runtime error.
	StopReasonError
	// StopReasonTermination indicates the program has terminated normally.
	StopReasonTermination
	// StopReasonMaxSteps indicates execution stopped after reaching max step limit.
	StopReasonMaxSteps
	// StopReasonInterrupt indicates execution was interrupted by the user.
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

// ExecutionResult contains a complete snapshot of execution state after a step/continue operation.
type ExecutionResult struct {
	// The reason why execution stopped, or [StopReasonNone] if still running.
	StopReason StopReason `json:"stopReason"`
	// Error message if an error occurred during execution (nil if successful).
	Error error `json:"error"`
	// Number of instructions executed in this operation.
	Steps uint64 `json:"steps"`
	// Number of CPU cycles executed in this operation.
	Cycles uint64 `json:"cycles"`
	// Hit breakpoint, if any (nil if execution did not hit a breakpoint).
	Breakpoint *Breakpoint `json:"breakpoint"`
	// Hit watchpoint, if any (nil if execution did not hit a watchpoint).
	Watchpoint *Watchpoint `json:"watchpoint"`
	// Memory address of the last instruction executed.
	LastInstruction uint32 `json:"lastInstruction"`
	// Source code location of the last executed instruction (nil if unknown).
	LastLocation *SourceLocation `json:"lastLocation"`
	// Cycles the CPU is lagging behind target timing (for real-time systems).
	LaggingCycles uint32 `json:"laggingCycles"`
}

// DebuggerStatus describes the current execution state of the debugged program.
type DebuggerStatus int

const (
	// DebuggerStatusNotReady_MissingProgram: Debugger cannot run, no program has been loaded.
	DebuggerStatusNotReady_MissingProgram DebuggerStatus = iota
	// DebuggerStatusNotReady_MissingRuntime: Debugger cannot run, no runtime has been configured.
	DebuggerStatusNotReady_MissingRuntime
	// DebuggerStatusNotReady_MissingSystemConfig: Debugger cannot run, no system configuration has been loaded.
	DebuggerStatusNotReady_MissingSystemConfig
	// DebuggerStatusIdle: Debugger is ready but the debugged program has not yet started.
	DebuggerStatusIdle
	// DebuggerStatusRunning: The debugged program is currently executing.
	DebuggerStatusRunning
	// DebuggerStatusPaused: Execution has paused at a breakpoint/watchpoint and is waiting to continue.
	DebuggerStatusPaused
	// DebuggerStatusTerminated: The debugged program has finished and cannot be executed further.
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

// DebuggerState is a snapshot of the debugger state including CPU state and execution status.
type DebuggerState struct {
	// Current execution status of the debugged program. See [DebuggerStatus] for details.
	Status DebuggerStatus `json:"status"`
	// Current CPU register values keyed by register name.
	Registers map[string]*Register `json:"registers"`
	// Current CPU status flags. See [FlagState] for details.
	Flags *FlagState `json:"flags"`
}

// DebuggerEventType describes the kind of event reported by the debugger.
type DebuggerEventType int

const (
	// DebuggerEventProgramLoaded fires when a program is loaded and ready to execute.
	DebuggerEventProgramLoaded DebuggerEventType = iota
	// DebuggerEventStepped fires when a step operation completes.
	DebuggerEventStepped
	// DebuggerEventBreakpointHit fires when a code breakpoint is hit.
	DebuggerEventBreakpointHit
	// DebuggerEventWatchpointHit fires when a data watchpoint is hit.
	DebuggerEventWatchpointHit
	// DebuggerEventProgramTerminated fires when the program terminates normally.
	DebuggerEventProgramTerminated
	// DebuggerEventProgramHalted fires when the program executes a halt instruction.
	DebuggerEventProgramHalted
	// DebuggerEventError fires when a runtime error occurs.
	DebuggerEventError
	// DebuggerEventSourceLocationChanged fires when the current source location changes.
	DebuggerEventSourceLocationChanged
	// DebuggerEventInterrupted fires when execution is interrupted by the user.
	DebuggerEventInterrupted
	// DebuggerEventLagging fires when the emulator cannot keep up with target timing.
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

// DebuggerEvent represents a notification from the debugger to the UI layer.
type DebuggerEvent struct {
	// Type of event that occurred. See [DebuggerEventType] for details.
	Type DebuggerEventType
	// Execution result associated with this event (if applicable).
	Result *ExecutionResult
}

// DebuggerEventCallback is a function type for handling debugger events.
// The UI layer registers a callback to receive notifications of significant debugger events.
// See [DebuggerEventType] and [DebuggerEvent] for details.
type DebuggerEventCallback func(event *DebuggerEvent)

// VariableValue describes a program variable's current value and location.
type VariableValue struct {
	// Name of the variable.
	Name string `json:"name"`
	// Type name of the variable (e.g., "int", "unsigned int").
	TypeName string `json:"typeName"`
	// Formatted value string for display, may include special values like "<optimized out>".
	ValueString string `json:"valueString"`
	// Human-readable storage location description (e.g., "[sp+16]", "r0", "<optimized out>").
	Location string `json:"location"`
	// Size of the variable in bytes.
	Size int `json:"size"`
	// Memory region where the variable is stored (nil if optimized out or in register).
	MemoryLocation *MemoryRegion `json:"memoryLocation"`
}

// StackFrame represents one frame in the call stack.
type StackFrame struct {
	// Source code location of the function call (nil if unknown).
	SourceLocation *SourceLocation `json:"sourceLocation"`
	// Name of the function in this frame (nil if unknown).
	Function *string `json:"function"`
	// Memory region occupied by this stack frame.
	Memory *MemoryRegion `json:"memory"`
}

// CommandHelp provides usage documentation for a debugger command.
type CommandHelp struct {
	// Command name.
	Name string `json:"name"`
	// Alternative command names that invoke the same command.
	Aliases []string `json:"aliases"`
	// Short description of what the command does.
	Description string `json:"description"`
	// Usage syntax for the command.
	Usage string `json:"usage"`
	// Example usages of the command.
	Examples []string `json:"examples"`
}

// InfoResult contains the debugger state, system configuration, program information, or runtime info.
type InfoResult struct {
	// Error message if the info command failed (nil if successful).
	Error error `json:"error"`
	// Debugger state and CPU registers (present when requesting general info). See [DebuggerState].
	DebuggerState *DebuggerState `json:"debuggerState"`
	// System and peripheral configuration (present when requesting system info). See [SystemInfo].
	SystemInfo *SystemInfo `json:"systemInfo"`
	// Loaded program info (present when requesting program info). See [ProgramInfo].
	ProgramInfo *ProgramInfo `json:"programInfo"`
	// Configured runtime info (present when requesting runtime info). See [RuntimeInfo].
	RuntimeInfo *RuntimeInfo `json:"runtimeInfo"`
}

// StackResult contains the call stack and stack memory contents.
type StackResult struct {
	// Error message if stack inspection failed (nil if successful).
	Error error `json:"error"`
	// Current stack pointer (top of stack) value.
	SP uint32 `json:"sp"`
	// Complete stack memory contents for inspection.
	StackData []byte `json:"stackData"`
	// Call stack frames with source locations and function names.
	StackFrames []*StackFrame `json:"stackFrames"`
}

// VarsResult contains the values of accessible variables at the current location.
type VarsResult struct {
	// Error message if variable inspection failed (nil if successful).
	Error error `json:"error"`
	// Current local and parameter variable values. See [VariableValue] for structure.
	Variables []*VariableValue `json:"variables"`
}
