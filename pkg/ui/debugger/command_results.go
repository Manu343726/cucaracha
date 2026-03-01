package debugger

// PeripheralInfo describes a hardware peripheral and its memory-mapped I/O configuration.
type PeripheralInfo struct {
	// Instance name of the peripheral (e.g., "uart0", "timer1").
	Name string `json:"name"`
	// Type classification of the peripheral (e.g., "terminal", "timer", "rtc").
	Type string `json:"type"`
	// Human-readable display name for the peripheral (e.g., "UART Serial Port", "Programmable Interval Timer").
	DisplayName string `json:"displayName"`
	// Detailed description of the peripheral's functionality.
	Description string `json:"description"`
	// Memory-mapped I/O base address for this peripheral.
	BaseAddress uint32 `json:"baseAddress"`
	// Size in bytes of the memory-mapped I/O region for this peripheral.
	Size uint32 `json:"size"`
	// Interrupt vector number this peripheral uses for interrupts.
	InterruptVector uint8 `json:"interruptVector"`
}

// SystemInfo describes the loaded hardware configuration including memory layout and peripherals.
type SystemInfo struct {
	// Total addressable memory size in bytes.
	TotalMemory uint32 `json:"totalMemory"`
	// Size in bytes of the code/instruction region.
	CodeSize uint32 `json:"codeSize"`
	// Size in bytes of the initialized data region.
	DataSize uint32 `json:"dataSize"`
	// Size in bytes of the automatic storage (stack) region.
	StackSize uint32 `json:"stackSize"`
	// Size in bytes of the heap (dynamic allocation) region.
	HeapSize uint32 `json:"heapSize"`
	// Size in bytes of the memory-mapped I/O peripheral region.
	PeripheralSize uint32 `json:"peripheralSize"`
	// Total number of interrupt vectors supported by this system.
	NumberOfVectors uint32 `json:"numberOfVectors"`
	// Size in bytes of each interrupt vector table entry.
	VectorEntrySize uint32 `json:"vectorEntrySize"`
	// Number of available hardware peripherals in this system.
	NumPeripherals int `json:"numPeripherals"`
	// Detailed information about each peripheral. See [PeripheralInfo] for structure.
	Peripherals []PeripheralInfo `json:"peripherals"`
}

// ProgramInfo describes the loaded program including its compilation status and entry point.
type ProgramInfo struct {
	// Warnings generated during program loading or compilation (empty if none).
	Warnings []string `json:"warnings"`
	// Path to the source file that was loaded (nil if a pre-compiled object file was loaded).
	SourceFile *string `json:"sourceFile"`
	// Path to the compiled object file (nil if the input was a pre-compiled object file).
	ObjectFile *string `json:"objectFile"`
	// Entry point address where program execution begins.
	EntryPoint uint32 `json:"entryPoint"`
	// Whether the loaded program contains debug information (DWARF) for source-level debugging.
	HasDebugInfo bool `json:"hasDebugInfo"`
}

// RuntimeInfo describes the currently loaded execution runtime.
type RuntimeInfo struct {
	// Type of execution runtime. See [RuntimeType] for available options.
	Runtime RuntimeType `json:"runtime"`
}

// BreakResult contains the newly created breakpoint after a Break command.
type BreakResult struct {
	// Error message if breakpoint creation failed (nil if successful).
	Error error `json:"error"`
	// The created breakpoint. See [Breakpoint] for structure.
	Breakpoint *Breakpoint `json:"breakpoint"`
}

// WatchResult contains the newly created watchpoint after a Watch command.
type WatchResult struct {
	// Error message if watchpoint creation failed (nil if successful).
	Error error `json:"error"`
	// The created watchpoint. See [Watchpoint] for structure.
	Watchpoint *Watchpoint `json:"watchpoint"`
}

// RemoveBreakpointResult contains the deleted breakpoint after a RemoveBreakpoint command.
type RemoveBreakpointResult struct {
	// Error message if removal failed (nil if successful).
	Error error `json:"error"`
	// The breakpoint that was removed. See [Breakpoint] for structure.
	Breakpoint *Breakpoint `json:"breakpoint"`
}

// RemoveWatchpointResult contains the deleted watchpoint after a RemoveWatchpoint command.
type RemoveWatchpointResult struct {
	// Error message if removal failed (nil if successful).
	Error error `json:"error"`
	// The watchpoint that was removed. See [Watchpoint] for structure.
	Watchpoint *Watchpoint `json:"watchpoint"`
}

// SourceResult contains a source code snippet around a specified location.
type SourceResult struct {
	// Error message if source retrieval failed (nil if successful).
	Error error `json:"error"`
	// Source code snippet. See [SourceCodeSnippet] for structure.
	Snippet *SourceCodeSnippet `json:"snippet"`
}

// EvalResult contains the computed value of an evaluated expression.
type EvalResult struct {
	// Error message if evaluation failed (nil if successful).
	Error error `json:"error"`
	// The numerical result of evaluating the expression.
	Value uint32 `json:"value"`
	// String representation of the result (for display).
	ValueString string `json:"valueString"`
	// Byte representation of the result.
	ValueBytes []byte `json:"valueBytes"`
}

// ListResult contains all active breakpoints and watchpoints.
type ListResult struct {
	// Error message if retrieval failed (nil if successful).
	Error error `json:"error"`
	// Currently active code breakpoints. See [Breakpoint] for structure.
	Breakpoints []*Breakpoint `json:"breakpoints"`
	// Currently active memory watchpoints. See [Watchpoint] for structure.
	Watchpoints []*Watchpoint `json:"watchpoints"`
}

// LoadSystemFromEmbeddedResult contains the loaded system information from the embedded default configuration.
type LoadSystemFromEmbeddedResult struct {
	// Error message if loading failed (nil if successful).
	Error error `json:"error"`
	// Loaded system configuration (nil if loading failed). See [SystemInfo] for structure.
	System *SystemInfo `json:"system"`
}

// LoadSystemFromFileResult contains the loaded system information from a YAML configuration file.
type LoadSystemFromFileResult struct {
	// Error message if loading or parsing failed (nil if successful).
	Error error `json:"error"`
	// Loaded system configuration (nil if loading failed). See [SystemInfo] for structure.
	System *SystemInfo `json:"system"`
}

// LoadProgramFromFileResult contains the loaded program information after compilation/loading from a file.
type LoadProgramFromFileResult struct {
	// Error message if loading or compilation failed (nil if successful).
	Error error `json:"error"`
	// Loaded program information (nil if loading failed). See [ProgramInfo] for structure.
	Program *ProgramInfo `json:"program"`
}

// LoadRuntimeResult contains the loaded runtime configuration.
type LoadRuntimeResult struct {
	// Error message if loading failed (nil if successful).
	Error error `json:"error"`
	// Loaded runtime information (nil if loading failed). See [RuntimeInfo] for structure.
	Runtime *RuntimeInfo `json:"runtime"`
}

// LoadResult contains the results of loading program, system, and runtime from a combined configuration.
type LoadResult struct {
	// Error message if loading failed (nil if successful).
	Error error `json:"error"`
	// Result of loading system configuration (present if system was loaded). See [LoadSystemFromFileResult].
	System *LoadSystemFromFileResult `json:"system,omitempty"`
	// Result of loading program (present if program was loaded). See [LoadProgramFromFileResult].
	Program *LoadProgramFromFileResult `json:"program,omitempty"`
	// Result of loading runtime (present if runtime was loaded). See [LoadRuntimeResult].
	Runtime *LoadRuntimeResult `json:"runtime,omitempty"`
}
