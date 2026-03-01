package debugger

import (
	"encoding/json"
	"fmt"
)

// BreakArgs specifies where to set a breakpoint, either by source location or instruction address.
type BreakArgs struct {
	// Location in source code where to set the breakpoint. If specified, the debugger will resolve this to an instruction address.
	SourceLocation *SourceLocation `json:"sourceLocation"`
	// Instruction address as an eval expression where to set the breakpoint (e.g., "0x1000", "_start", "sp+8").
	// Used when source location is not available.
	Address *string `json:"address"`
}

// WatchArgs specifies a memory range to monitor for reads, writes, or both.
type WatchArgs struct {
	// Starting address of the memory range to watch. Can be an eval expression (e.g., "0x2000", "sp", "_data").
	StartAddress string `json:"startAddress"`
	// Ending address of the memory range to watch (optional). If omitted, Size must be specified.
	// Can be an eval expression (e.g., "0x2100", "sp+256").
	EndAddress *string `json:"endAddress"`
	// Size of the memory range to watch (optional). Used if EndAddress is not provided.
	// Can be an eval expression (e.g., "256", "4").
	Size *string `json:"size"`
	// Type of access to watch (read, write, or both). Optional; defaults to [WatchpointTypeReadWrite].
	Type *WatchpointType `json:"type"`
}

// RemoveBreakpointArgs specifies which [Breakpoint] to delete by ID.
type RemoveBreakpointArgs struct {
	// ID of the breakpoint to remove. Obtained from [Breakpoint] or [ListResult].
	ID int `json:"id"`
}

// RemoveWatchpointArgs specifies which [Watchpoint] to delete by ID.
type RemoveWatchpointArgs struct {
	// ID of the watchpoint to remove. Obtained from [Watchpoint] or [ListResult].
	ID int `json:"id"`
}

// DisasmArgs specifies parameters for disassembling instructions.
type DisasmArgs struct {
	// Starting address to disassemble from. Can be an eval expression (e.g., "0x1000", "r0", "sp+8", "_start").
	Address string `json:"addressExpr"`
	// Number of instructions to disassemble (optional). Can be an eval expression (e.g., "10", "100").
	// If omitted, a default count will be used.
	CountExpr *string `json:"count"`
	// Whether to include source code lines in the disassembly output.
	ShowSource bool `json:"showSource"`
	// Whether to include control flow graph information in the disassembly output.
	ShowCFG bool `json:"showCFG"`
}

// StepArgs specifies how to execute a step operation.
type StepArgs struct {
	// Determines the stepping behavior (into, over, out, or return). See [StepMode] for details.
	StepMode StepMode `json:"stepMode"`
	// Determines what counts as a single step: instructions or source lines. See [StepCountMode] for details.
	CountMode StepCountMode `json:"countMode"`
}

// PrintArgs specifies what value or expression to print.
type PrintArgs struct {
	// What to print: register name (e.g., "r0", "sp"), memory address (e.g., "0x2000"),
	// or complex eval expression (e.g., "r0 + sp", "[0x2000]").
	Expression string `json:"expression"`
}

// SetArgs specifies a register and value to write to it.
type SetArgs struct {
	// Name of the register to modify (e.g., "r0", "sp", "pc").
	Target string `json:"target"`
	// Value to write to the register.
	Value uint32 `json:"value"`
}

// SourceContextMode controls how lines around a source location are displayed.
type SourceContextMode int

const (
	// SourceContextTop shows N lines starting from the target line.
	SourceContextTop SourceContextMode = iota
	// SourceContextCentered shows N lines centered around the target line.
	SourceContextCentered
	// SourceContextBottom shows N lines ending at the target line.
	SourceContextBottom
)

func (s SourceContextMode) String() string {
	switch s {
	case SourceContextTop:
		return "top"
	case SourceContextCentered:
		return "centered"
	case SourceContextBottom:
		return "bottom"
	default:
		return "unknown"
	}
}

func SourceContextModeFromString(s string) (SourceContextMode, error) {
	switch s {
	case "top":
		return SourceContextTop, nil
	case "centered":
		return SourceContextCentered, nil
	case "bottom":
		return SourceContextBottom, nil
	default:
		return 0, fmt.Errorf("unknown SourceContextMode: \"%s\"", s)
	}
}

func (s SourceContextMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *SourceContextMode) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	val, err := SourceContextModeFromString(str)
	if err != nil {
		return err
	}
	*s = val
	return nil
}

// SourceArgs specifies parameters for displaying source code.
type SourceArgs struct {
	// Location in source code to display. If provided, the debugger will jump to this location.
	Location *SourceLocation `json:"location"`
	// Number of lines to display around the target location.
	ContextLines int `json:"contextLines"`
	// How to display the context lines relative to the target. See [SourceContextMode] for options.
	ContextMode SourceContextMode `json:"contextMode"`
}

// EvalArgs specifies an expression to evaluate.
type EvalArgs struct {
	// Expression to evaluate. Can reference registers (e.g., "r0", "sp"),
	// memory (e.g., "[0x2000]"), or be a mathematical expression (e.g., "sp + 8").
	Expression string `json:"expression"`
}

// InfoType specifies what category of debugger information to display.
type InfoType int

const (
	// InfoTypeGeneral shows general debugger info (status, registers, flags).
	InfoTypeGeneral InfoType = iota
	// InfoTypeRuntime shows runtime info (system config, memory layout).
	InfoTypeRuntime
	// InfoTypeProgram shows program info (source file, entry point, code layout).
	InfoTypeProgram
)

func (i InfoType) String() string {
	switch i {
	case InfoTypeGeneral:
		return "general"
	case InfoTypeRuntime:
		return "runtime"
	case InfoTypeProgram:
		return "program"
	default:
		return "unknown"
	}
}

func InfoTypeFromString(s string) (InfoType, error) {
	switch s {
	case "general", "":
		return InfoTypeGeneral, nil
	case "runtime", "system":
		return InfoTypeRuntime, nil
	case "program":
		return InfoTypeProgram, nil
	default:
		return 0, fmt.Errorf("unknown InfoType: \"%s\"", s)
	}
}

func (i InfoType) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

func (i *InfoType) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	val, err := InfoTypeFromString(str)
	if err != nil {
		return err
	}
	*i = val
	return nil
}

// InfoArgs specifies what category of information to retrieve.
type InfoArgs struct {
	// Category of information to display: general (status, registers, flags),
	// runtime (system configuration, memory layout), or program (entry point, symbols).
	// See [InfoType] for details.
	Type InfoType `json:"type"`
}

// CurrentSourceArgs specifies parameters for displaying source code at the current location.
type CurrentSourceArgs struct {
	// Number of lines to display around the current location.
	ContextLines int `json:"contextLines"`
	// How to display the context lines relative to the current location.
	// See [SourceContextMode] for options.
	ContextMode SourceContextMode `json:"contextMode"`
}

// LoadSystemFromFileArgs specifies the path to a system configuration file.
type LoadSystemFromFileArgs struct {
	// Path to a YAML file containing system configuration.
	FilePath string `json:"filePath"`
}

// LoadSystemArgs specifies arguments for loading the embedded default system configuration.
// This type has no fields; the embedded system is always used.
type LoadSystemArgs struct {
	// No arguments - uses the embedded default system configuration
}

// LoadProgramFromFileArgs specifies how to load a program from a file.
type LoadProgramFromFileArgs struct {
	// Path to the program file (object file or C source file).
	FilePath string `json:"filePath"`
	// Whether to automatically build the clang compiler if needed (defaults to true).
	AutoBuildClang *bool `json:"autoBuildClang,omitempty"`
	// Force a rebuild of the clang compiler even if it already exists (defaults to false).
	ForceRebuildClang *bool `json:"forceRebuildClang,omitempty"`
}

// LoadProgramArgs specifies how to load a program. This is equivalent to [LoadProgramFromFileArgs].
// Used by the shorthand LoadProgram command.
type LoadProgramArgs struct {
	// Path to the program file (object file or C source file).
	FilePath string `json:"filePath"`
	// Whether to automatically build clang if needed (defaults to true).
	AutoBuildClang *bool `json:"autoBuildClang,omitempty"`
	// Force rebuild of clang (defaults to false).
	ForceRebuildClang *bool `json:"forceRebuildClang,omitempty"`
}

// LoadArgs specifies parameters for loading program, system, and runtime in a single operation.
type LoadArgs struct {
	// Path to a YAML file containing program, system, and runtime configuration (optional).
	// If specified, other fields are ignored.
	FullDescriptorPath *string `json:"fullDescriptorPath,omitempty"`
	// Path to YAML file containing system configuration (optional).
	// Used if FullDescriptorPath is not provided.
	SystemConfigPath *string `json:"systemConfigPath,omitempty"`
	// Path to program file (optional). Used if FullDescriptorPath is not provided.
	ProgramPath *string `json:"programPath,omitempty"`
	// [RuntimeType] to use (optional). Used if FullDescriptorPath is not provided.
	Runtime *RuntimeType `json:"runtime,omitempty"`
}

// RuntimeType specifies the execution engine type.
type RuntimeType uint

const (
	// RuntimeTypeInterpreter uses the software interpreter runtime for program execution.
	RuntimeTypeInterpreter RuntimeType = iota
)

func (r RuntimeType) String() string {
	switch r {
	case RuntimeTypeInterpreter:
		return "interpreter"
	default:
		return "unknown"
	}
}

func RuntimeTypeFromString(s string) (RuntimeType, error) {
	switch s {
	case "interpreter":
		return RuntimeTypeInterpreter, nil
	default:
		return 0, fmt.Errorf("unknown RuntimeType: \"%s\"", s)
	}
}

func (r RuntimeType) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

func (r *RuntimeType) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	val, err := RuntimeTypeFromString(str)
	if err != nil {
		return err
	}
	*r = val
	return nil
}

// LoadRuntimeArgs specifies which runtime to load and use for execution.
type LoadRuntimeArgs struct {
	// The type of runtime to load and use. See [RuntimeType] for available options.
	Runtime RuntimeType `json:"runtimeType"`
}

// SymbolsArgs specifies optional filtering for the Symbols command.
type SymbolsArgs struct {
	// Optional symbol name pattern to filter which symbols to display.
	// If nil, all known symbols are returned. Supports wildcard patterns (e.g., "_*", "*_init").
	SymbolName *string `json:"symbolName"`
}
