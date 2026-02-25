package debugger

import (
	"encoding/json"
	"fmt"
)

// Break command arguments
type BreakArgs struct {
	SourceLocation *SourceLocation `json:"sourceLocation"` // Location in source code where to set the breakpoint
	Address        *string         `json:"address"`        // Instruction address as an eval expression where to set the breakpoint
}

// Watch command arguments
type WatchArgs struct {
	StartAddress string          `json:"startAddress"` // Starting address of the memory range to watch
	EndAddress   *string         `json:"endAddress"`   // Ending address of the memory range to watch (optional)
	Size         *string         `json:"size"`         // Size of the memory range to watch as an eval expression (optional, used if endAddress is not provided)
	Type         *WatchpointType `json:"type"`         // Type of access to watch (optional, defaults to WatchpointTypeReadWrite)
}

// RemoveBreakpoint command arguments
type RemoveBreakpointArgs struct {
	ID int `json:"id"` // Breakpoint ID
}

// RemoveWatchpoint command arguments
type RemoveWatchpointArgs struct {
	ID int `json:"id"` // Watchpoint ID
}

// Disasm command arguments
type DisasmArgs struct {
	Address    string  `json:"addressExpr"` // Address eval expression (e.g., "0x1000", "r0", "sp+8")
	CountExpr  *string `json:"count"`       // Number of instructions to disassemble (optional) as an eval expression (e.g. "sp+8", "10")
	ShowSource bool    `json:"showSource"`  // Whether to include source code in disassembly output
	ShowCFG    bool    `json:"showCFG"`     // Whether to include control flow graph in disassembly output
}

// Step command arguments
type StepArgs struct {
	StepMode  StepMode      `json:"stepMode"`  // Determines the stepping behavior
	CountMode StepCountMode `json:"countMode"` // Determines how the debugger interprets what is a single step, e.g., by instructions or source lines
}

// Print command arguments
type PrintArgs struct {
	Expression string `json:"expression"` // What to print (register name, memory address, complex expression)
}

// Set command arguments
type SetArgs struct {
	Target string `json:"target"` // Register name
	Value  uint32 `json:"value"`  // Value to set
}

// Controls what lines around the target line of source commands are shown
type SourceContextMode int

const (
	// Show N lines starting from the target line
	SourceContextTop SourceContextMode = iota
	// Show N lines centered around the target line
	SourceContextCentered
	// Show N lines ending at the target line
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

// Source command arguments
type SourceArgs struct {
	Location     *SourceLocation   `json:"location"`     // Location in source code to display
	ContextLines int               `json:"contextLines"` // Number of lines to show
	ContextMode  SourceContextMode `json:"contextMode"`  // How to display context lines
}

// Eval command arguments
type EvalArgs struct {
	Expression string `json:"expression"` // Expression to evaluate
}

// Info command type
type InfoType int

const (
	// Show general debugger info (status, registers, flags)
	InfoTypeGeneral InfoType = iota
	// Show runtime info (system config, memory layout)
	InfoTypeRuntime
	// Show program info (source file, entry point, code layout)
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

// Info command arguments
type InfoArgs struct {
	Type InfoType `json:"type"` // Type of info to display (general, runtime, or program)
}

// CurrentSource command arguments
type CurrentSourceArgs struct {
	ContextLines int               `json:"contextLines"` // Number of lines to show
	ContextMode  SourceContextMode `json:"contextMode"`  // How to display context lines
}

// LoadSystemFromFile command arguments
type LoadSystemFromFileArgs struct {
	FilePath string `json:"filePath"` // Path to system configuration file
}

// LoadProgramFromFile command arguments
type LoadProgramFromFileArgs struct {
	FilePath          string `json:"filePath"`                    // Path to program file
	AutoBuildClang    *bool  `json:"autoBuildClang,omitempty"`    // Enable automatic building of clang (defaults to true if not specified)
	ForceRebuildClang *bool  `json:"forceRebuildClang,omitempty"` // Force rebuild of clang  (defaults to false if not specified)
}

// Load command arguments
type LoadArgs struct {
	FullDescriptorPath *string      `json:"fullDescriptorPath,omitempty"` // Path to YAML file containing program, system, and runtime configuration
	SystemConfigPath   *string      `json:"systemConfigPath,omitempty"`   // Path to YAML file containing system configuration (used if fullDescriptorPath is not provided)
	ProgramPath        *string      `json:"programPath,omitempty"`        // Path to program file (used if fullDescriptorPath is not provided)
	Runtime            *RuntimeType `json:"runtime,omitempty"`            // Runtime type (used if fullDescriptorPath is not provided)
}

// Runtime types
type RuntimeType uint

const (
	// Software interpreter runtime
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

// LoadRuntime command arguments
type LoadRuntimeArgs struct {
	Runtime RuntimeType `json:"runtimeType"` // Type of runtime to load (e.g., "interpreter")
}

// Symbols command arguments
type SymbolsArgs struct {
	SymbolName *string `json:"symbolName"` // Optional symbol name pattern to filter symbols (nil shows all)
}
