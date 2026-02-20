// Package debugger provides an abstraction layer for debugging Cucaracha programs.
// It separates the debugger logic from the presentation layer, allowing different
// frontends (CLI, GUI, TUI, REST API, etc.) to reuse the same debugger core.
package debugger

// =============================================================================
// DebuggerCommands interface - High-level command operations
// =============================================================================

// DebuggerCommands provides high-level command operations.
// These methods take string arguments (as entered by user) and return
// structured results for display. All parsing and validation is done here.
type DebuggerCommands interface {
	// Executes a single execution step
	Step(args *StepArgs) *ExecutionResult

	// Continues execution until the next breakpoint/watchpoint/termination
	Continue() *ExecutionResult

	// CmdRun executes run command
	Run() *ExecutionResult

	// Interrupts execution
	Interrupt() *ExecutionResult

	// Adds a breakpoint. Args: address or symbol
	Break(args *BreakArgs) *BreakResult

	// Adds a watchpoint. Args: address or symbol
	Watch(args *WatchArgs) *WatchResult

	// Deletes a breakpoint or watchpoint. Args: id
	RemoveBreakpoint(args *RemoveBreakpointArgs) *RemoveBreakpointResult

	// Deletes a watchpoint. Args: id
	RemoveWatchpoint(args *RemoveWatchpointArgs) *RemoveWatchpointResult

	// Returns breakpoints and watchpoints
	List() *ListResult
	// Disassembles instructions. Args: optional address, optional count
	Disasm(args *DisasmArgs) *DisasmResult

	// Returns the current instruction for display
	CurrentInstruction() *CurrentInstructionResult

	// Displays memory. Args: address expression, optional count
	Memory(args *MemoryArgs) *MemoryResult

	// Displays source code
	Source(args *SourceArgs) *SourceResult

	// Displays the current source
	CurrentSource(args *CurrentSourceArgs) *SourceResult

	// Evaluates an expression
	Eval(args *EvalArgs) *EvalResult
	// Returns current debugger state, system info, or program info
	Info(args *InfoArgs) *InfoResult

	// Returns all register values
	Registers() *RegistersResult

	// Returns stack information
	Stack() *StackResult

	// Returns accessible variables
	Vars() *VarsResult

	// Returns symbols from the loaded program
	Symbols(args *SymbolsArgs) *SymbolsResult

	// Resets the program to its initial state
	Reset() *ExecutionResult

	// Restarts the program (reset + continue)
	Restart() *ExecutionResult

	// Load system configuration from a YAML file
	LoadSystemFromFile(args *LoadSystemFromFileArgs) *LoadSystemFromFileResult

	// Load the embedded default system configuration
	LoadSystemFromEmbedded() *LoadSystemFromEmbeddedResult

	// Load a program into the debugged runtime from a file.
	//
	// Note that this may involve compiling the input file. Any error or warning during
	// compilation or loading will be returned.
	LoadProgramFromFile(args *LoadProgramFromFileArgs) *LoadProgramFromFileResult

	// Configure the execution runtime used to run the debugged program
	LoadRuntime(args *LoadRuntimeArgs) *LoadRuntimeResult

	// Loads program, system, and runtime from a single YAML file
	Load(args *LoadArgs) *LoadResult
}

type Runtime uint

const (
	// Software interpreter runtime
	RuntimeInterpreter Runtime = iota
)

type Debugger interface {
	DebuggerCommands

	// Set callback used to notify UI of debug events
	SetEventCallback(callback DebuggerEventCallback)
}
