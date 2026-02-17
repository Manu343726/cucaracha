package ui

// PeripheralInfo contains information about a peripheral
type PeripheralInfo struct {
	Name            string `json:"name"`            // Instance name of the peripheral
	Type            string `json:"type"`            // Type of the peripheral (e.g., "terminal", "timer")
	DisplayName     string `json:"displayName"`     // Human-readable display name (e.g., "UART Serial Port")
	Description     string `json:"description"`     // Description of the peripheral
	BaseAddress     uint32 `json:"baseAddress"`     // Memory-mapped I/O base address
	Size            uint32 `json:"size"`            // Memory-mapped I/O region size
	InterruptVector uint8  `json:"interruptVector"` // Interrupt vector number
}

// Result of debugger Break command
type BreakResult struct {
	Error      error       `json:"error"`      // Error, if any
	Breakpoint *Breakpoint `json:"breakpoint"` // Created breakpoint
}

// Result of debugger Watch command
type WatchResult struct {
	Error      error       `json:"error"`      // Error, if any
	Watchpoint *Watchpoint `json:"watchpoint"` // Created watchpoint
}

// Result of RemoveBreakpoint command
type RemoveBreakpointResult struct {
	Error      error       `json:"error"`      // Error, if any
	Breakpoint *Breakpoint `json:"breakpoint"` // Breakpoint that was removed
}

// Result of RemoveWatchpoint command
type RemoveWatchpointResult struct {
	Error      error       `json:"error"`      // Error, if any
	Watchpoint *Watchpoint `json:"watchpoint"` // Watchpoint that was removed
}

// Result of Source command
type SourceResult struct {
	Error   error              `json:"error"`   // Error, if any
	Snippet *SourceCodeSnippet `json:"snippet"` // Source code snippet
}

// Result of Eval command
type EvalResult struct {
	Error       error  `json:"error"`       // Error, if any
	Value       uint32 `json:"value"`       // Evaluated value
	ValueString string `json:"valueString"` // String representation
	ValueBytes  []byte `json:"valueBytes"`  // Byte representation
}

// Result of List command
type ListResult struct {
	Error       error         `json:"error"`       // Error, if any
	Breakpoints []*Breakpoint `json:"breakpoints"` // Active breakpoints
	Watchpoints []*Watchpoint `json:"watchpoints"` // Active watchpoints
}

// Result of LoadSystemFromFile command
type LoadSystemResult struct {
	Error           error            `json:"error"`           // Error, if any
	TotalMemory     uint32           `json:"totalMemory"`     // Total memory size in bytes
	CodeSize        uint32           `json:"codeSize"`        // Code region size in bytes
	DataSize        uint32           `json:"dataSize"`        // Data region size in bytes
	StackSize       uint32           `json:"stackSize"`       // Stack region size in bytes
	HeapSize        uint32           `json:"heapSize"`        // Heap region size in bytes
	PeripheralSize  uint32           `json:"peripheralSize"`  // Peripheral MMIO region size in bytes
	NumberOfVectors uint32           `json:"numberOfVectors"` // Number of interrupt vectors
	VectorEntrySize uint32           `json:"vectorEntrySize"` // Size of each interrupt vector entry in bytes
	NumPeripherals  int              `json:"numPeripherals"`  // Number of available peripherals
	Peripherals     []PeripheralInfo `json:"peripherals"`     // Detailed peripheral information
}

// Result of LoadProgramFromFile command
type LoadProgramResult struct {
	Error      error    `json:"error"`      // Error, if any
	Warnings   []string `json:"warnings"`   // Compilation/loading warnings, if any
	SourceFile *string  `json:"sourceFile"` // Path to the source file that was loaded
	ObjectFile *string  `json:"objectFile"` // Path to the generated object file, if the input program was compiled
	EntryPoint uint32   `json:"entryPoint"` // Program entry point address
}

// Result of LoadRuntime command
type LoadRuntimeResult struct {
	Error   error       `json:"error"`   // Error, if any
	Runtime RuntimeType `json:"runtime"` // Loaded runtime type
}

// Result of Load command
type LoadResult struct {
	Error   error              `json:"error"`             // Error, if any
	System  *LoadSystemResult  `json:"system,omitempty"`  // Result of loading system configuration (present if system was loaded)
	Program *LoadProgramResult `json:"program,omitempty"` // Result of loading program (present if program was loaded)
	Runtime *LoadRuntimeResult `json:"runtime,omitempty"` // Result of loading runtime (present if runtime was loaded)
}
