package ui

import (
	"encoding/json"
	"fmt"
)

// Specifies the exact command sent to the debugger
type DebuggerCommandId int

const (
	// Load a program from a file
	DebuggerCommandLoadProgramFromFile DebuggerCommandId = iota
	// Load system configuration from a file
	DebuggerCommandLoadSystemFromFile
	// Load execution runtime
	DebuggerCommandLoadRuntime
	// Loads program, system, and runtime from a single YAML file
	DebuggerCommandLoad

	// Executes a single execution step
	DebuggerCommandStep
	// Continues execution until the next breakpoint/watchpoint/termination
	DebuggerCommandContinue
	// Interrupts execution
	DebuggerCommandInterrupt
	// Sets a breakpoint at the specified address
	DebuggerCommandBreak
	// Removes a breakpoint by ID
	DebuggerCommandRemoveBreakpoint
	// Sets a watchpoint on the specified memory range
	DebuggerCommandWatch
	// Removes a watchpoint by ID
	DebuggerCommandRemoveWatchpoint
	// Returns breakpoints and watchpoints
	DebuggerCommandList
	// Disassembles instructions.
	DebuggerCommandDisassemble
	// Returns the current instruction for display
	DebuggerCommandCurrentInstruction
	// Displays memory
	DebuggerCommandMemory
	// Displays source code
	DebuggerCommandSource
	// Displays the current source
	DebuggerCommandCurrentSource
	// Evaluates an expression
	DebuggerCommandEvaluateExpression
	// Returns current debugger state
	DebuggerCommandInfo
	// Returns all register values
	DebuggerCommandRegisters
	// Returns stack information
	DebuggerCommandStack
	// Returns accessible variables
	DebuggerCommandVariables
	// Returns symbols from the loaded program
	DebuggerCommandSymbols
	// Resets the program to its initial state
	DebuggerCommandReset
	// Resets the program and continues execution
	DebuggerCommandRestart
)

func (c DebuggerCommandId) String() string {
	switch c {
	case DebuggerCommandLoadProgramFromFile:
		return "loadProgramFromFile"
	case DebuggerCommandLoadSystemFromFile:
		return "loadSystemFromFile"
	case DebuggerCommandLoadRuntime:
		return "loadRuntime"
	case DebuggerCommandLoad:
		return "load"
	case DebuggerCommandStep:
		return "step"
	case DebuggerCommandContinue:
		return "continue"
	case DebuggerCommandInterrupt:
		return "interrupt"
	case DebuggerCommandBreak:
		return "setBreakpoint"
	case DebuggerCommandRemoveBreakpoint:
		return "removeBreakpoint"
	case DebuggerCommandWatch:
		return "setWatchpoint"
	case DebuggerCommandRemoveWatchpoint:
		return "removeWatchpoint"
	case DebuggerCommandList:
		return "list"
	case DebuggerCommandDisassemble:
		return "disassemble"
	case DebuggerCommandCurrentInstruction:
		return "currentInstruction"
	case DebuggerCommandMemory:
		return "memory"
	case DebuggerCommandSource:
		return "source"
	case DebuggerCommandCurrentSource:
		return "currentSource"
	case DebuggerCommandEvaluateExpression:
		return "evaluateExpression"
	case DebuggerCommandInfo:
		return "info"
	case DebuggerCommandRegisters:
		return "registers"
	case DebuggerCommandStack:
		return "stack"
	case DebuggerCommandVariables:
		return "variables"
	case DebuggerCommandSymbols:
		return "symbols"
	case DebuggerCommandReset:
		return "reset"
	case DebuggerCommandRestart:
		return "restart"
	default:
		return "unknown"
	}
}

func DebuggerCommandIdFromString(s string) (DebuggerCommandId, error) {
	switch s {
	case "loadProgramFromFile":
		return DebuggerCommandLoadProgramFromFile, nil
	case "loadSystemFromFile":
		return DebuggerCommandLoadSystemFromFile, nil
	case "loadRuntime":
		return DebuggerCommandLoadRuntime, nil
	case "load":
		return DebuggerCommandLoad, nil
	case "step":
		return DebuggerCommandStep, nil
	case "continue":
		return DebuggerCommandContinue, nil
	case "interrupt":
		return DebuggerCommandInterrupt, nil
	case "setBreakpoint":
		return DebuggerCommandBreak, nil
	case "removeBreakpoint":
		return DebuggerCommandRemoveBreakpoint, nil
	case "setWatchpoint":
		return DebuggerCommandWatch, nil
	case "removeWatchpoint":
		return DebuggerCommandRemoveWatchpoint, nil
	case "list":
		return DebuggerCommandList, nil
	case "disassemble":
		return DebuggerCommandDisassemble, nil
	case "currentInstruction":
		return DebuggerCommandCurrentInstruction, nil
	case "memory":
		return DebuggerCommandMemory, nil
	case "source":
		return DebuggerCommandSource, nil
	case "currentSource":
		return DebuggerCommandCurrentSource, nil
	case "evaluateExpression":
		return DebuggerCommandEvaluateExpression, nil
	case "info":
		return DebuggerCommandInfo, nil
	case "registers":
		return DebuggerCommandRegisters, nil
	case "stack":
		return DebuggerCommandStack, nil
	case "variables":
		return DebuggerCommandVariables, nil
	case "symbols":
		return DebuggerCommandSymbols, nil
	case "reset":
		return DebuggerCommandReset, nil
	case "restart":
		return DebuggerCommandRestart, nil
	default:
		return 0, fmt.Errorf("unknown DebuggerCommandId: \"%s\"", s)
	}
}

func (c DebuggerCommandId) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c *DebuggerCommandId) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	val, err := DebuggerCommandIdFromString(s)
	if err != nil {
		return err
	}
	*c = val
	return nil
}

// Controls the behavior of the Step command
type StepMode int

const (
	// Step one source line (steps into function calls)
	StepModeInto StepMode = iota
	// Step one source line, stepping over function calls
	StepModeOver
	// Step out of the current function
	StepModeOut
)

func (s StepMode) String() string {
	switch s {
	case StepModeInto:
		return "into"
	case StepModeOver:
		return "over"
	case StepModeOut:
		return "out"
	default:
		return "unknown"
	}
}

func StepModeFromString(s string) (StepMode, error) {
	switch s {
	case "into":
		return StepModeInto, nil
	case "over":
		return StepModeOver, nil
	case "out":
		return StepModeOut, nil
	default:
		return 0, fmt.Errorf("unknown StepMode: \"%s\"", s)
	}
}

func (s StepMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *StepMode) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	val, err := StepModeFromString(str)
	if err != nil {
		return err
	}
	*s = val
	return nil
}

type StepCountMode int

const (
	// Count by instructions
	StepCountInstructions StepCountMode = iota
	// Count by source lines
	StepCountSourceLines
)

func (s StepCountMode) String() string {
	switch s {
	case StepCountInstructions:
		return "instructions"
	case StepCountSourceLines:
		return "sourceLines"
	default:
		return "unknown"
	}
}

func StepCountModeFromString(s string) (StepCountMode, error) {
	switch s {
	case "instructions":
		return StepCountInstructions, nil
	case "sourceLines":
		return StepCountSourceLines, nil
	default:
		return 0, fmt.Errorf("unknown StepCountMode: \"%s\"", s)
	}
}

func (s StepCountMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *StepCountMode) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	val, err := StepCountModeFromString(str)
	if err != nil {
		return err
	}
	*s = val
	return nil
}

// Represents a debugger command
type DebuggerCommand struct {
	Id                   uint64                `json:"id"`                   // Unique ID for this command instance
	Command              DebuggerCommandId     `json:"command"`              // Type of command
	LoadSystemArgs       *LoadSystemArgs       `json:"loadSystemArgs"`       // Command arguments for LoadSystem command
	LoadProgramArgs      *LoadProgramArgs      `json:"loadProgramArgs"`      // Command arguments for LoadProgram command
	LoadRuntimeArgs      *LoadRuntimeArgs      `json:"loadRuntimeArgs"`      // Command arguments for LoadRuntime command
	LoadArgs             *LoadArgs             `json:"loadArgs"`             // Command arguments for Load command
	StepArgs             *StepArgs             `json:"stepArgs"`             // Command arguments for Step command
	BreakArgs            *BreakArgs            `json:"breakArgs"`            // Command arguments for Break command
	WatchArgs            *WatchArgs            `json:"watchArgs"`            // Command arguments for Watch command
	RemoveBreakpointArgs *RemoveBreakpointArgs `json:"removeBreakpointArgs"` // Command arguments for RemoveBreakpoint command
	RemoveWatchpointArgs *RemoveWatchpointArgs `json:"removeWatchpointArgs"` // Command arguments for RemoveWatchpoint command
	DisasmArgs           *DisasmArgs           `json:"disasmArgs"`           // Command arguments for Disasm command
	MemoryArgs           *MemoryArgs           `json:"memoryArgs"`           // Command arguments for Memory command
	SourceArgs           *SourceArgs           `json:"sourceArgs"`           // Command arguments for Source command
	CurrentSourceArgs    *CurrentSourceArgs    `json:"currentSourceArgs"`    // Command arguments for CurrentSource command
	EvalArgs             *EvalArgs             `json:"evalArgs"`             // Command arguments for Eval command
	InfoArgs             *InfoArgs             `json:"infoArgs"`             // Command arguments for Info command
	SymbolsArgs          *SymbolsArgs          `json:"symbolsArgs"`          // Command arguments for Symbols command
}

// Represents a debugger command result
type DebuggerCommandResult struct {
	Id                       uint64                    `json:"id"`                       // Unique ID for this command instance
	Command                  DebuggerCommandId         `json:"command"`                  // Command identifier
	LoadSystemResult         *LoadSystemResult         `json:"loadSystemResult"`         // Result of LoadSystem command
	LoadProgramResult        *LoadProgramResult        `json:"loadProgramResult"`        // Result of LoadProgram command
	LoadRuntimeResult        *LoadRuntimeResult        `json:"loadRuntimeResult"`        // Result of LoadRuntime command
	LoadResult               *LoadResult               `json:"loadResult"`               // Result of Load command
	StepResult               *ExecutionResult          `json:"stepResult"`               // Result of Step command
	ContinueResult           *ExecutionResult          `json:"continueResult"`           // Result of Continue command
	InterruptResult          *ExecutionResult          `json:"interruptResult"`          // Result of Interrupt command
	BreakResult              *BreakResult              `json:"breakResult"`              // Result of Break command
	WatchResult              *WatchResult              `json:"watchResult"`              // Result of Watch command
	RemoveBreakpointResult   *RemoveBreakpointResult   `json:"removeBreakpointResult"`   // Result of RemoveBreakpoint command
	RemoveWatchpointResult   *RemoveWatchpointResult   `json:"removeWatchpointResult"`   // Result of RemoveWatchpoint command
	ListResult               *ListResult               `json:"listResult"`               // Result of List command
	DisassemblyResult        *DisassemblyResult        `json:"disassemblyResult"`        // Result of Disasm command
	CurrentInstructionResult *CurrentInstructionResult `json:"currentInstructionResult"` // Result of CurrentInstruction command
	MemoryResult             *MemoryResult             `json:"memoryResult"`             // Result of Memory command
	SourceResult             *SourceResult             `json:"sourceResult"`             // Result of Source command
	CurrentSourceResult      *SourceResult             `json:"currentSourceResult"`      // Result of CurrentSource command
	EvalResult               *EvalResult               `json:"evalResult"`
	InfoResult               *InfoResult               `json:"infoResult"`      // Result of Info command
	RegistersResult          *RegistersResult          `json:"registersResult"` // Result of Registers command
	StackResult              *StackResult              `json:"stackResult"`     // Result of Stack command
	VariablesResult          *VarsResult               `json:"variablesResult"` // Result of Variables command
	SymbolsResult            *SymbolsResult            `json:"symbolsResult"`   // Result of Symbols command
	ResetResult              *ExecutionResult          `json:"resetResult"`     // Result of Reset command
	RestartResult            *ExecutionResult          `json:"restartResult"`   // Result of Restart command
}

// An interface to interact with the debugger in the UI
type Debugger interface {
	// Sends a command to the debugger and returns the result
	Execute(cmd *DebuggerCommand) (*DebuggerCommandResult, error)

	// Sets a callback to receive debugger events
	SetEventCallback(callback DebuggerEventCallback)

	// Resets the debugger to its initial state
	Reset() *ExecutionResult

	// Restarts the debugger (reset + continue)
	Restart() *ExecutionResult
}
