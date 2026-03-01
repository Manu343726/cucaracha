// Package debugger provides an abstraction layer for debugging Cucaracha programs.
// It separates the debugger logic from the presentation layer, allowing different
// frontends (CLI, GUI, TUI, REST API, etc.) to reuse the same debugger core.
package debugger

// =============================================================================
// DebuggerCommands interface - High-level command operations
// =============================================================================

// DebuggerCommands provides high-level command operations.
// These methods take structured arguments and return structured results for display.
// All parsing, validation, and error handling is performed by the implementation.
type DebuggerCommands interface {
	// Step executes a single instruction step. Returns execution state after the step completes.
	//
	// Args specify step mode and optional step count. See [StepArgs] and [ExecutionResult].
	Step(args *StepArgs) *ExecutionResult

	// Continue resumes execution until the next breakpoint, watchpoint, or program termination.
	//
	// Returns execution state when execution stops. See [ExecutionResult].
	Continue() *ExecutionResult

	// Run starts program execution from the beginning. Equivalent to Reset followed by Continue.
	//
	// Returns execution state when execution stops. See [ExecutionResult].
	Run() *ExecutionResult

	// Interrupt stops the currently running program and returns control to the debugger.
	// Returns execution state after interruption. See [ExecutionResult].
	Interrupt() *ExecutionResult

	// Break adds a code breakpoint. Args specify the breakpoint location by source or address.
	//
	// See [BreakArgs] and [BreakResult] for details.
	Break(args *BreakArgs) *BreakResult

	// Watch adds a memory/data watchpoint. Args specify the memory location and access type.
	//
	// See [WatchArgs] and [WatchResult] for details.
	Watch(args *WatchArgs) *WatchResult

	// RemoveBreakpoint removes a code breakpoint by ID. See [RemoveBreakpointArgs] and [RemoveBreakpointResult].
	RemoveBreakpoint(args *RemoveBreakpointArgs) *RemoveBreakpointResult

	// RemoveWatchpoint removes a memory watchpoint by ID. See [RemoveWatchpointArgs] and [RemoveWatchpointResult].
	RemoveWatchpoint(args *RemoveWatchpointArgs) *RemoveWatchpointResult

	// List returns all active code breakpoints and memory watchpoints. See [ListResult].
	List() *ListResult

	// Disasm disassembles instructions from a specified address. Args specify the address and optional count.
	// See [DisasmArgs] and [DisasmResult] for details.
	Disasm(args *DisasmArgs) *DisasmResult

	// CurrentInstruction returns the instruction at the current program counter. See [CurrentInstructionResult].
	CurrentInstruction() *CurrentInstructionResult

	// Memory displays memory contents at a specified address. Args specify the address and optional byte count.
	//
	// See [MemoryArgs] and [MemoryResult] for details.
	Memory(args *MemoryArgs) *MemoryResult

	// Source displays source code around a specified location. Args specify the location and context display mode.
	//
	// See [SourceArgs] and [SourceResult] for details.
	Source(args *SourceArgs) *SourceResult

	// CurrentSource displays source code around the current execution location. Args specify context display options.
	//
	// See [CurrentSourceArgs] and [SourceResult] for details.
	CurrentSource(args *CurrentSourceArgs) *SourceResult

	// Eval evaluates an expression in the current CPU context. Args specify the expression to evaluate.
	//
	// See [EvalArgs] and [EvalResult] for details.
	Eval(args *EvalArgs) *EvalResult

	// Info returns debugger state, system configuration, program info, or runtime info based on args.
	//
	// See [InfoArgs] and [InfoResult] for details.
	Info(args *InfoArgs) *InfoResult

	// Registers returns all CPU register values and status flags. See [RegistersResult].
	Registers() *RegistersResult

	// Stack returns the call stack and stack memory contents. See [StackResult].
	Stack() *StackResult

	// Vars returns the values of accessible local variables and parameters. See [VarsResult].
	Vars() *VarsResult

	// Symbols returns function, global, and label symbols matching optional filter criteria.
	//
	// See [SymbolsArgs] and [SymbolsResult] for details.
	Symbols(args *SymbolsArgs) *SymbolsResult

	// Reset resets the debugged program to its initial state. Returns execution state. See [ExecutionResult].
	Reset() *ExecutionResult

	// Restart resets and continues program execution (equivalent to Reset followed by Continue). See [ExecutionResult].
	Restart() *ExecutionResult

	// LoadSystem loads the embedded default system configuration. See [LoadSystemArgs] and [LoadSystemFromEmbeddedResult].
	LoadSystem(args *LoadSystemArgs) *LoadSystemFromEmbeddedResult

	// LoadProgram loads a program from a file, optionally compiling it if needed.
	//
	// See [LoadProgramArgs] and [LoadProgramFromFileResult] for details.
	LoadProgram(args *LoadProgramArgs) *LoadProgramFromFileResult

	// LoadSystemFromFile loads system configuration from a YAML file.
	//
	// See [LoadSystemFromFileArgs] and [LoadSystemFromFileResult] for details.
	LoadSystemFromFile(args *LoadSystemFromFileArgs) *LoadSystemFromFileResult

	// LoadSystemFromEmbedded loads the embedded default system configuration.
	//
	// Returns the loaded system information. See [LoadSystemFromEmbeddedResult].
	LoadSystemFromEmbedded() *LoadSystemFromEmbeddedResult

	// LoadProgramFromFile loads a program from a file, optionally compiling it if needed.
	//
	// This is the full-named version of LoadProgram().
	//
	// See [LoadProgramFromFileArgs] and [LoadProgramFromFileResult] for details.
	LoadProgramFromFile(args *LoadProgramFromFileArgs) *LoadProgramFromFileResult

	// LoadRuntime configures the execution runtime (interpreter engine) used to run the debugged program.
	//
	// See [LoadRuntimeArgs] and [LoadRuntimeResult] for details.
	LoadRuntime(args *LoadRuntimeArgs) *LoadRuntimeResult

	// Load loads program, system, and runtime configuration from a single YAML descriptor file or individual components.
	//
	// See [LoadArgs] and [LoadResult] for details.
	Load(args *LoadArgs) *LoadResult
}

// Runtime specifies the type of execution engine used to run the debugged program.
type Runtime uint

const (
	// RuntimeInterpreter uses the software interpreter to execute the debugged program.
	RuntimeInterpreter Runtime = iota
)

// Debugger combines command operations with event notification and represents the complete debugger interface.
type Debugger interface {
	// High-level command operations. See [DebuggerCommands].
	DebuggerCommands

	// SetEventCallback registers a callback function to receive notifications of debugger events.
	// Events include execution stops, program state changes, and errors.
	// See [DebuggerEventCallback] and [DebuggerEvent] for details.
	SetEventCallback(callback DebuggerEventCallback)
}
