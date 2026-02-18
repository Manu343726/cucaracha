// Package debugger provides an abstraction layer for debugging Cucaracha programs.
// It separates the debugger logic from the presentation layer, allowing different
// frontends (CLI, GUI, TUI, REST API, etc.) to reuse the same debugger core.
package debugger

import (
	"github.com/Manu343726/cucaracha/pkg/ui"
)

// =============================================================================
// DebuggerCommands interface - High-level command operations
// =============================================================================

// DebuggerCommands provides high-level command operations.
// These methods take string arguments (as entered by user) and return
// structured results for display. All parsing and validation is done here.
type DebuggerCommands interface {
	// Executes a single execution step
	Step(args *ui.StepArgs) *ui.ExecutionResult

	// Continues execution until the next breakpoint/watchpoint/termination
	Continue() *ui.ExecutionResult

	// CmdRun executes run command
	Run() *ui.ExecutionResult

	// Interrupts execution
	Interrupt() *ui.ExecutionResult

	// Adds a breakpoint. Args: address or symbol
	Break(args *ui.BreakArgs) *ui.BreakResult

	// Adds a watchpoint. Args: address or symbol
	Watch(args *ui.WatchArgs) *ui.WatchResult

	// Deletes a breakpoint or watchpoint. Args: id
	RemoveBreakpoint(args *ui.RemoveBreakpointArgs) *ui.RemoveBreakpointResult

	// Deletes a watchpoint. Args: id
	RemoveWatchpoint(args *ui.RemoveWatchpointArgs) *ui.RemoveWatchpointResult

	// Returns breakpoints and watchpoints
	List() *ui.ListResult
	// Disassembles instructions. Args: optional address, optional count
	Disasm(args *ui.DisasmArgs) *ui.DisassemblyResult

	// Returns the current instruction for display
	CurrentInstruction() *ui.CurrentInstructionResult

	// Displays memory. Args: address expression, optional count
	Memory(args *ui.MemoryArgs) *ui.MemoryResult

	// Displays source code
	Source(args *ui.SourceArgs) *ui.SourceResult

	// Displays the current source
	CurrentSource(args *ui.CurrentSourceArgs) *ui.SourceResult

	// Evaluates an expression
	Eval(args *ui.EvalArgs) *ui.EvalResult
	// Returns current debugger state, system info, or program info
	Info(args *ui.InfoArgs) *ui.InfoResult

	// Returns all register values
	Registers() *ui.RegistersResult

	// Returns stack information
	Stack() *ui.StackResult

	// Returns accessible variables
	Vars() *ui.VarsResult

	// Returns symbols from the loaded program
	Symbols(args *ui.SymbolsArgs) *ui.SymbolsResult

	// Resets the program to its initial state
	Reset() *ui.ExecutionResult

	// Restarts the program (reset + continue)
	Restart() *ui.ExecutionResult
}

type Runtime uint

const (
	// Software interpreter runtime
	RuntimeInterpreter Runtime = iota
)

type Debugger interface {
	DebuggerCommands

	// Set callback used to notify UI of debug events
	SetEventCallback(callback ui.DebuggerEventCallback)

	// Load system configuration from a YAML file
	LoadSystemFromFile(args *ui.LoadSystemArgs) *ui.LoadSystemResult

	// Load the embedded default system configuration
	LoadSystemFromEmbedded() *ui.LoadSystemResult

	// Load a program into the debugged runtime from a file.
	//
	// Note that this may involve compiling the input file. Any error or warning during
	// compilation or loading will be returned.
	LoadProgramFromFile(args *ui.LoadProgramArgs) *ui.LoadProgramResult

	// Configure the execution runtime used to run the debugged program
	LoadRuntime(args *ui.LoadRuntimeArgs) *ui.LoadRuntimeResult

	// Loads program, system, and runtime from a single YAML file
	Load(args *ui.LoadArgs) *ui.LoadResult
}
