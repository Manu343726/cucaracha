package core

import (
	"fmt"

	"github.com/Manu343726/cucaracha/pkg/hw/memory"
	"github.com/Manu343726/cucaracha/pkg/runtime"
	"github.com/Manu343726/cucaracha/pkg/runtime/program"
	"github.com/Manu343726/cucaracha/pkg/runtime/program/sourcecode"
)

// StopReason indicates why execution stopped
type StopReason int

const (
	// StopNone indicates execution has not stopped
	StopNone StopReason = iota
	// StopStep indicates execution stopped after a single step
	StopStep
	// StopBreakpoint indicates execution stopped at a breakpoint
	StopBreakpoint
	// StopWatchpoint indicates execution stopped due to a watchpoint
	StopWatchpoint
	// StopHalt indicates the CPU halted
	StopHalt
	// StopError indicates an execution error occurred
	StopError
	// StopTermination indicates normal program termination
	StopTermination
	// StopMaxSteps indicates max steps limit was reached
	StopMaxSteps
	// StopInterrupt indicates execution was interrupted by user
	StopInterrupt
)

// String returns the string representation of a StopReason
func (r StopReason) String() string {
	switch r {
	case StopNone:
		return "none"
	case StopStep:
		return "step"
	case StopBreakpoint:
		return "breakpoint"
	case StopWatchpoint:
		return "watchpoint"
	case StopHalt:
		return "halt"
	case StopError:
		return "error"
	case StopTermination:
		return "termination"
	case StopMaxSteps:
		return "max_steps"
	case StopInterrupt:
		return "interrupt"
	default:
		return fmt.Sprintf("unknown(%d)", r)
	}
}

// Breakpoint represents a code breakpoint
type Breakpoint struct {
	// ID is the unique breakpoint identifier
	ID int
	// Address is the memory address of the breakpoint
	Address uint32
	// Enabled indicates if the breakpoint is active
	Enabled bool
	// HitCount tracks how many times this breakpoint has been hit
	HitCount int
	// Condition is an optional condition expression (for future use)
	Condition string
}

// Watchpoint represents a memory watchpoint
type Watchpoint struct {
	// ID is the unique watchpoint identifier
	ID int
	// Memory to watch
	Memory *memory.Range
	// Type indicates read, write, or read/write
	Type WatchpointType
	// Enabled indicates if the watchpoint is active
	Enabled bool
	// HitCount tracks how many times this watchpoint has been triggered
	HitCount int
	// LastValue stores the last known value at this address
	LastValue []byte
}

// WatchpointType indicates what access triggers the watchpoint
type WatchpointType int

const (
	// WatchWrite triggers on write access
	WatchWrite WatchpointType = 1 << iota
	// WatchRead triggers on read access
	WatchRead
	// WatchReadWrite triggers on any access
	WatchReadWrite = WatchWrite | WatchRead
)

// DebugEvent represents events that can be sent to the UI
type DebugEvent int

// EventCallback is called when an execution event occurs
// Return true to continue execution, false to stop
type EventCallback func(event *Event) bool

const (
	// EventProgramLoaded is fired when a program is loaded
	EventProgramLoaded DebugEvent = iota
	// EventStepped is fired after stepping one or more instructions
	EventStepped
	// EventBreakpointHit is fired when a breakpoint is hit
	EventBreakpointHit
	// EventWatchpointHit is fired when a watchpoint triggers
	EventWatchpointHit
	// EventProgramTerminated is fired when the program exits normally
	EventProgramTerminated
	// EventProgramHalted is fired when the CPU halts
	EventProgramHalted
	// EventError is fired when an error occurs
	EventError
	// EventSourceLocationChanged is fired when the source location changes
	EventSourceLocationChanged
	// EventInterrupted is fired when execution is interrupted by user (Ctrl+C)
	EventInterrupted
	// EventLagging is fired when the emulator can't keep up with target execution speed
	EventLagging
)

// String returns the string representation of a DebugEvent
func (e DebugEvent) String() string {
	switch e {
	case EventProgramLoaded:
		return "program_loaded"
	case EventStepped:
		return "stepped"
	case EventBreakpointHit:
		return "breakpoint_hit"
	case EventWatchpointHit:
		return "watchpoint_hit"
	case EventProgramTerminated:
		return "program_terminated"
	case EventProgramHalted:
		return "program_halted"
	case EventError:
		return "error"
	case EventSourceLocationChanged:
		return "source_location_changed"
	case EventInterrupted:
		return "interrupted"
	case EventLagging:
		return "lagging"
	default:
		return "unknown"
	}
}

// Contains data associated with a debug event
type Event struct {
	// Event type
	Event DebugEvent
	// Execution result of the current instruction at the time of the event
	Result *ExecutionResult
}

// ExecutionResult contains the result of an execution operation
type ExecutionResult struct {
	StopReason     StopReason
	StepsExecuted  int
	CyclesExecuted int64
	Error          error
	Breakpoint     *Breakpoint
	Watchpoint     *Watchpoint
	LastPC         uint32
	ReturnValue    uint32
	Lagging        bool
	LagCycles      int64
	SourceLocation *sourcecode.Location
}

// Debugger provides low level debugging capabilities
type Debugger interface {
	// Runtime returns the underlying runtime
	Runtime() runtime.Runtime

	// Event callback management
	SetEventCallback(callback EventCallback)
	HasEventCallback() bool

	// Execution control
	Step() *ExecutionResult
	Continue() *ExecutionResult
	RunUntil(targetAddr uint32) *ExecutionResult
	StepOver() *ExecutionResult
	StepOut() *ExecutionResult
	StepIntoSource() *ExecutionResult
	StepOverSource() *ExecutionResult
	Interrupt() *ExecutionResult
	IsInterrupted() bool

	// Execution state
	LastResult() *ExecutionResult

	// Breakpoint management
	AddBreakpoint(addr uint32) (*Breakpoint, error)
	RemoveBreakpoint(id int) (*Breakpoint, error)
	GetBreakpoint(id int) *Breakpoint
	GetBreakpointAt(addr uint32) *Breakpoint
	ListBreakpoints() []*Breakpoint
	EnableBreakpoint(id int, enabled bool) bool
	ClearBreakpoints()

	// Watchpoint management
	AddWatchpoint(r *memory.Range, wpType WatchpointType) (*Watchpoint, error)
	RemoveWatchpoint(id int) (*Watchpoint, error)
	GetWatchpoint(id int) *Watchpoint
	ListWatchpoints() []*Watchpoint
	ClearWatchpoints()

	// Termination addresses
	AddTerminationAddress(addr uint32)
	RemoveTerminationAddress(addr uint32)
	IsTerminationAddress(addr uint32) bool
	ClearTerminationAddresses()

	// Program information
	Program() program.ProgramFile
}
