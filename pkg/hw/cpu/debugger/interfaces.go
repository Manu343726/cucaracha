// Package debugger provides an abstraction layer for debugging Cucaracha programs.
// It separates the debugger logic from the presentation layer, allowing different
// frontends (CLI, GUI, TUI, REST API, etc.) to reuse the same debugger core.
package debugger

import (
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/interpreter"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
)

// DebugEvent represents events that can be sent to the UI
type DebugEvent int

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
	default:
		return "unknown"
	}
}

// EventData contains data associated with a debug event
type EventData struct {
	// Event type
	Event DebugEvent
	// Address where the event occurred (e.g., breakpoint address, PC)
	Address uint32
	// Message associated with the event (e.g., error message)
	Message string
	// Error if any
	Error error
	// ReturnValue for program termination
	ReturnValue uint32
	// BreakpointID for breakpoint events
	BreakpointID int
	// WatchpointID for watchpoint events
	WatchpointID int
	// SourceLocation for source-related events
	SourceLocation *mc.SourceLocation
	// SourceText contains the source code text at the current location
	SourceText string
	// StepsExecuted for step events
	StepsExecuted int
}

// RegisterInfo contains information about a register
type RegisterInfo struct {
	Name  string
	Index uint32
	Value uint32
}

// InstructionInfo contains information about an instruction
type InstructionInfo struct {
	Address       uint32
	Encoding      uint32
	Mnemonic      string
	Operands      string
	RawBytes      []byte
	HasBreakpoint bool
	IsCurrentPC   bool
	// Branch target information (for CFG visualization)
	BranchTarget    uint32 // Resolved branch target address (0 if not a branch or unknown)
	BranchTargetSym string // Symbol name of branch target (empty if none)
}

// MemoryRegion represents a memory region for display
type MemoryRegion struct {
	Name       string
	StartAddr  uint32
	EndAddr    uint32
	RegionType MemoryRegionType
}

// MemoryRegionType classifies memory regions
type MemoryRegionType int

const (
	RegionUnknown MemoryRegionType = iota
	RegionCode
	RegionData
	RegionStack
	RegionHeap
	RegionIO
)

// VariableValue contains a variable's current value
type VariableValue struct {
	Name        string
	TypeName    string
	Value       interface{} // Can be uint32, int32, string, etc.
	ValueString string      // Formatted value string for display (includes "<optimized out>" etc.)
	Location    string      // Human-readable location (e.g., "[sp+16]", "r0", "<optimized out>")
	Size        int
}

// StackFrame represents a stack frame
type StackFrame struct {
	Address  uint32
	Function string
	File     string
	Line     int
}

// DebuggerState contains a snapshot of the debugger state
type DebuggerState struct {
	PC           uint32
	SP           uint32
	LR           uint32
	CPSR         uint32
	Registers    []RegisterInfo
	Flags        FlagState
	IsRunning    bool
	IsTerminated bool
}

// FlagState contains the CPU flags
type FlagState struct {
	N bool // Negative
	Z bool // Zero
	C bool // Carry
	V bool // Overflow
}

// TerminalSize represents the dimensions of a terminal
type TerminalSize struct {
	Width  int
	Height int
}

// ResizeHandler is a callback function called when terminal size changes
type ResizeHandler func(size TerminalSize)

// DebuggerUI is the interface that presentation layers must implement.
// This allows different frontends (CLI, GUI, TUI, REST API) to present
// debugger information in their own way.
type DebuggerUI interface {
	// OnEvent is called when a debug event occurs
	OnEvent(event EventData)

	// GetTerminalSize returns the current terminal dimensions
	// For non-terminal UIs, this may return a default or configured size
	GetTerminalSize() TerminalSize

	// OnResize registers a callback to be called when the terminal is resized
	// The callback will be called with the new dimensions
	// Returns a function to unregister the callback
	OnResize(handler ResizeHandler) (unregister func())

	// ShowMessage displays a message to the user
	ShowMessage(level MessageLevel, format string, args ...interface{})

	// ShowInstruction displays the current instruction
	ShowInstruction(info InstructionInfo)

	// ShowRegisters displays register values
	ShowRegisters(regs []RegisterInfo, flags FlagState)

	// ShowMemory displays memory contents
	ShowMemory(addr uint32, data []byte, regions []MemoryRegion)

	// ShowDisassembly displays disassembled instructions
	ShowDisassembly(instructions []InstructionInfo, currentPC uint32)

	// ShowBreakpoints displays the list of breakpoints
	ShowBreakpoints(breakpoints []BreakpointInfo)

	// ShowWatchpoints displays the list of watchpoints
	ShowWatchpoints(watchpoints []WatchpointInfo)

	// ShowStack displays the stack contents
	ShowStack(sp uint32, data []byte, frames []StackFrame)

	// ShowBacktrace displays the call stack (function frames)
	ShowBacktrace(frames []StackFrame)

	// ShowSource displays source code
	ShowSource(location *mc.SourceLocation, lines []SourceLine, currentLine int)

	// ShowVariables displays accessible variables
	ShowVariables(variables []VariableValue)

	// ShowEvalResult displays the result of an expression evaluation
	ShowEvalResult(expr string, value uint32, err error)

	// ShowHelp displays help information
	ShowHelp(commands []CommandHelp)

	// Prompt requests input from the user (for interactive UIs)
	// Returns the input string and any error
	Prompt(prompt string) (string, error)

	// PromptConfirm requests a yes/no confirmation
	PromptConfirm(message string) bool
}

// MessageLevel indicates the severity of a message
type MessageLevel int

const (
	LevelInfo MessageLevel = iota
	LevelSuccess
	LevelWarning
	LevelError
	LevelDebug
)

// SourceLine represents a line of source code
type SourceLine struct {
	LineNumber    int
	Text          string
	IsCurrent     bool
	HasBreakpoint bool
}

// CommandHelp contains help information for a command
type CommandHelp struct {
	Name        string
	Aliases     []string
	Description string
	Usage       string
	Examples    []string
}

// BreakpointInfo contains information about a breakpoint for display
type BreakpointInfo struct {
	ID              int
	Address         uint32
	Enabled         bool
	HitCount        int
	InstructionText string // Pre-formatted instruction text
	SourceFile      string // Source file if available
	SourceLine      int    // Source line number if available
	SourceText      string // Source code text if available
}

// WatchpointInfo contains information about a watchpoint for display
type WatchpointInfo struct {
	ID       int
	Address  uint32
	Size     int
	Type     string // "read", "write", or "read/write"
	Enabled  bool
	HitCount int
}

// DebuggerBackend is the interface for the debugger operations.
// This wraps the interpreter.Runner and provides high-level debugging operations.
type DebuggerBackend interface {
	// Program operations
	LoadProgram(program mc.ProgramFile) error
	Program() mc.ProgramFile
	DebugInfo() *mc.DebugInfo

	// Execution control
	Step(count int) ExecutionResult
	Continue() ExecutionResult
	Run() ExecutionResult
	Reset()

	// State inspection
	GetState() DebuggerState
	ReadRegister(name string) (uint32, error)
	WriteRegister(name string, value uint32) error
	ReadMemory(addr uint32, size int) ([]byte, error)
	WriteMemory(addr uint32, data []byte) error

	// Breakpoint management
	AddBreakpoint(addr uint32) (*interpreter.Breakpoint, error)
	RemoveBreakpoint(id int) error
	ListBreakpoints() []*interpreter.Breakpoint
	EnableBreakpoint(id int, enabled bool) error
	GetBreakpointInfos() []BreakpointInfo

	// Watchpoint management
	AddWatchpoint(addr uint32) (*interpreter.Watchpoint, error)
	RemoveWatchpoint(id int) error
	ListWatchpoints() []*interpreter.Watchpoint
	GetWatchpointInfos() []WatchpointInfo

	// Disassembly
	Disassemble(addr uint32, count int) ([]InstructionInfo, error)
	GetInstructionText(addr uint32) string

	// Expression evaluation
	EvalExpression(expr string) (uint32, error)

	// Symbol resolution
	ResolveSymbol(name string) (uint32, error)
	GetSymbolAt(addr uint32) (string, bool)

	// Source-level debugging
	GetSourceLocation(pc uint32) *mc.SourceLocation
	GetVariables(pc uint32) []VariableValue
	GetStackFrames() []StackFrame
	GetSourceLines(file string, startLine, endLine int) []SourceLine

	// Memory region information
	GetMemoryRegions() []MemoryRegion
	ClassifyAddress(addr uint32) (MemoryRegion, bool)
}

// ExecutionResult contains the result of an execution operation
type ExecutionResult struct {
	StopReason    interpreter.StopReason
	StepsExecuted int
	Error         error
	BreakpointID  int
	WatchpointID  int
	LastPC        uint32
	ReturnValue   uint32
}

// =============================================================================
// Command Result Types - All commands return structured results for display
// =============================================================================

// PrintResult contains the result of a print command
type PrintResult struct {
	Success bool
	Error   string
	// What was printed (register name, memory address, etc.)
	Target string
	// The value read
	Value       uint32
	ValueSigned int32
	// For memory reads
	IsMemory bool
	Address  uint32
}

// SetResult contains the result of a set command
type SetResult struct {
	Success bool
	Error   string
	// Register that was set
	Register string
	// New value
	Value       uint32
	ValueSigned int32
}

// BreakpointResult contains the result of a breakpoint operation
type BreakpointResult struct {
	Success bool
	Error   string
	// Operation performed
	Operation string // "add", "remove", "enable", "disable"
	// Breakpoint info
	ID      int
	Address uint32
}

// WatchpointResult contains the result of a watchpoint operation
type WatchpointResult struct {
	Success bool
	Error   string
	// Operation performed
	Operation string // "add", "remove"
	// Watchpoint info
	ID      int
	Address uint32
	Size    int
}

// DeleteResult contains the result of a delete command
type DeleteResult struct {
	Success bool
	Error   string
	// What was deleted
	WasBreakpoint bool
	WasWatchpoint bool
	ID            int
}

// DisassemblyResult contains disassembled instructions
type DisassemblyResult struct {
	Success      bool
	Error        string
	Address      uint32
	Instructions []InstructionInfo
}

// MemoryResult contains memory dump results
type MemoryResult struct {
	Success bool
	Error   string
	Address uint32
	Data    []byte
	Regions []MemoryRegion
}

// SourceResult contains source code display results
type SourceResult struct {
	Success bool
	Error   string
	File    string
	Lines   []SourceLine
}

// EvalResult contains expression evaluation results
type EvalResult struct {
	Success     bool
	Error       string
	Expression  string
	Value       uint32
	ValueSigned int32
	ValueBinary string
}

// CurrentInstructionResult contains the current instruction for display
type CurrentInstructionResult struct {
	PC              uint32
	InstructionWord uint32
	InstructionText string
	// Source location if available
	HasSource  bool
	SourceFile string
	SourceLine int
	SourceText string
}

// =============================================================================
// DebuggerCommands interface - High-level command operations
// =============================================================================

// DebuggerCommands provides high-level command operations.
// These methods take string arguments (as entered by user) and return
// structured results for display. All parsing and validation is done here.
type DebuggerCommands interface {
	// CmdStep executes step command. Args: optional count
	CmdStep(args []string) ExecutionResult

	// CmdContinue executes continue command
	CmdContinue() ExecutionResult

	// CmdRun executes run command
	CmdRun() ExecutionResult

	// CmdPrint prints a register or memory value. Args: what to print
	CmdPrint(args []string) PrintResult

	// CmdSet sets a register value. Args: register, value
	CmdSet(args []string) SetResult

	// CmdBreak adds a breakpoint. Args: address or symbol
	CmdBreak(args []string) BreakpointResult

	// CmdWatch adds a watchpoint. Args: address or symbol
	CmdWatch(args []string) WatchpointResult

	// CmdDelete deletes a breakpoint or watchpoint. Args: id
	CmdDelete(args []string) DeleteResult

	// CmdDisasm disassembles instructions. Args: optional address, optional count
	CmdDisasm(args []string) DisassemblyResult

	// CmdMemory displays memory. Args: address expression, optional count
	CmdMemory(args []string) MemoryResult

	// CmdSource displays source code. Args: optional context lines
	CmdSource(args []string) SourceResult

	// CmdEval evaluates an expression. Args: expression
	CmdEval(args []string) EvalResult

	// CmdInfo returns current CPU state
	CmdInfo() DebuggerState

	// CmdRegisters returns all register values
	CmdRegisters() []RegisterInfo

	// CmdStack returns stack information
	CmdStack() (uint32, []byte, []StackFrame)

	// CmdVars returns accessible variables
	CmdVars() []VariableValue

	// CmdList returns breakpoints and watchpoints
	CmdList() ([]BreakpointInfo, []WatchpointInfo)

	// GetCurrentInstruction returns the current instruction for display
	GetCurrentInstruction() CurrentInstructionResult
}
