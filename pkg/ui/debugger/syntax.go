package debugger

import (
	"fmt"
	"reflect"
	"sort"
	"time"
)

// Interface for parsing commands according to a specific syntax
type CommandsParser interface {
	// Parses a debugger command according to the syntax
	ParseCommand(input interface{}) (*DebuggerCommand, error)
}

// Interface for parsing optional arguments of commands according to a specific syntax
type CommandsOptionalArgumentsParser interface {
	// Parses command optional arguments according to the syntax, returning a map of argument names to values
	//
	// An implementation may choose to implement full parsing from scratch in this method, or it can choose to
	// implement parsing of individual arguments in ParseOptionalArgumentName and use the helper function
	// ParseSingleOptionalArgument to parse each argument and validate them against the expected arguments
	ParseOptionalArguments(args any, expectedArgs map[string]SyntaxOptionValueParser) (map[string]interface{}, error)
	// Parses a single optional argument and returns the argument name and raw value ready for parsing
	//
	// An implementation may return a boolean as raw value for boolean flags that don't have an explicit value (e.g., "verbose" flag can be passed as just "verbose" without "=true")
	ParseOptionalArgumentName(arg any) (string, any, error)
}

// Interface for parsing command arguments according to a specific syntax
type CommandsArgsParser interface {
	// Parses Step command arguments according to the syntax
	ParseStepArguments(args any) (*StepArgs, error)
	// Parses Break command arguments according to the syntax
	ParseBreakArguments(args any) (*BreakArgs, error)
	// Parses Watch command arguments according to the syntax
	ParseWatchArguments(args any) (*WatchArgs, error)
	// Parses RemoveBreakpoint command arguments according to the syntax
	ParseRemoveBreakpointArguments(args any) (*RemoveBreakpointArgs, error)
	// Parses RemoveWatchpoint command arguments according to the syntax
	ParseRemoveWatchpointArguments(args any) (*RemoveWatchpointArgs, error)
	// Parses Disasm command arguments according to the syntax
	ParseDisasmArguments(args any) (*DisasmArgs, error)
	// Parses Memory command arguments according to the syntax
	ParseMemoryArguments(args any) (*MemoryArgs, error)
	// Parses Source command arguments according to the syntax
	ParseSourceArguments(args any) (*SourceArgs, error)
	// Parses CurrentSource command arguments according to the syntax
	ParseCurrentSourceArguments(args any) (*CurrentSourceArgs, error)
	// Parses Eval command arguments according to the syntax
	ParseEvalArguments(args any) (*EvalArgs, error)
	// Parses Info command arguments according to the syntax
	ParseInfoArguments(args any) (*InfoArgs, error)
	// Parses Symbols command arguments according to the syntax
	ParseSymbolsArguments(args any) (*SymbolsArgs, error)
	// Parses LoadSystem command arguments according to the syntax
	ParseLoadSystemArguments(args any) (*LoadSystemArgs, error)
	// Parses LoadSystemFromFile command arguments according to the syntax
	ParseLoadSystemFromFileArguments(args any) (*LoadSystemFromFileArgs, error)
	// Parses LoadProgram command arguments according to the syntax
	ParseLoadProgramArguments(args any) (*LoadProgramArgs, error)
	// Parses LoadProgramFromFile command arguments according to the syntax
	ParseLoadProgramFromFileArguments(args any) (*LoadProgramFromFileArgs, error)
	// Parses LoadRuntime command arguments according to the syntax
	ParseLoadRuntimeArguments(args any) (*LoadRuntimeArgs, error)
	// Parses Load command arguments according to the syntax
	ParseLoadArguments(args any) (*LoadArgs, error)
}

// Interface for formatting commands and their arguments according to a specific syntax, used for help text generation and error messages
type CommandsFormatter interface {
	// Formats a command name according to the syntax
	FormatCommandName(command DebuggerCommandId) string
	// Formats an argument name according to the syntax, isRequired indicates if the argument is required or optional
	FormatArgumentName(name string, isRequired bool) string
	// Formats an argument value according to the syntax
	FormatArgumentValue(value interface{}) string
}

// Combines command parsing and formatting in a single interface
type CommandsSyntax interface {
	CommandsFormatter
	CommandsParser
}

// Signature of functions that parse option values for a specific syntax
type SyntaxOptionValueParser = func(arg any) (any, error)

// Helper function to parse command arguments using the provided syntax and expected arguments with their types
func ParseSingleOptionalArgument(syntax CommandsOptionalArgumentsParser, arg any, supportedArgs map[string]SyntaxOptionValueParser) (string, any, error) {
	name, rawValue, err := syntax.ParseOptionalArgumentName(arg)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse optional argument: %w", err)
	}
	if parser, ok := supportedArgs[name]; !ok {
		// Build list of supported arguments
		var supportedArgsList []string
		for argName := range supportedArgs {
			supportedArgsList = append(supportedArgsList, argName)
		}
		sort.Strings(supportedArgsList)
		return "", nil, fmt.Errorf("unexpected argument: \"%s\" (expected: %v)", name, supportedArgsList)
	} else {
		// If it's not a boolean flag, parse the value
		if reflect.TypeOf(rawValue) != reflect.TypeOf(true) {
			if parsedValue, err := parser(rawValue); err != nil {
				return "", nil, fmt.Errorf("invalid value for argument \"%s\": %w", name, err)
			} else {
				return name, parsedValue, nil
			}
		} else {
			// Return the boolean returned by the syntax directly (i.e. the argument is a flag)
			return name, rawValue, nil
		}
	}
}

// Helper function to parse command arguments using the provided syntax
func ParseCommandArguments(syntax CommandsArgsParser, command DebuggerCommandId, args any) (*DebuggerCommand, error) {
	result := &DebuggerCommand{
		Id:      uint64(time.Now().UnixNano()),
		Command: command,
	}

	switch command {
	case DebuggerCommandStep:
		stepArgs, err := syntax.ParseStepArguments(args)
		if err != nil {
			return nil, fmt.Errorf("failed to parse step command arguments: %w", err)
		}
		result.StepArgs = stepArgs
	case DebuggerCommandBreak:
		breakArgs, err := syntax.ParseBreakArguments(args)
		if err != nil {
			return nil, fmt.Errorf("failed to parse break command arguments: %w", err)
		}
		result.BreakArgs = breakArgs
	case DebuggerCommandWatch:
		watchArgs, err := syntax.ParseWatchArguments(args)
		if err != nil {
			return nil, fmt.Errorf("failed to parse watch command arguments: %w", err)
		}
		result.WatchArgs = watchArgs
	case DebuggerCommandRemoveBreakpoint:
		removeBreakpointArgs, err := syntax.ParseRemoveBreakpointArguments(args)
		if err != nil {
			return nil, fmt.Errorf("failed to parse remove breakpoint command arguments: %w", err)
		}
		result.RemoveBreakpointArgs = removeBreakpointArgs
	case DebuggerCommandRemoveWatchpoint:
		removeWatchpointArgs, err := syntax.ParseRemoveWatchpointArguments(args)
		if err != nil {
			return nil, fmt.Errorf("failed to parse remove watchpoint command arguments: %w", err)
		}
		result.RemoveWatchpointArgs = removeWatchpointArgs
	case DebuggerCommandDisasm:
		disasmArgs, err := syntax.ParseDisasmArguments(args)
		if err != nil {
			return nil, fmt.Errorf("failed to parse disasm command arguments: %w", err)
		}
		result.DisasmArgs = disasmArgs
	case DebuggerCommandMemory:
		memoryArgs, err := syntax.ParseMemoryArguments(args)
		if err != nil {
			return nil, fmt.Errorf("failed to parse memory command arguments: %w", err)
		}
		result.MemoryArgs = memoryArgs
	case DebuggerCommandSource:
		sourceArgs, err := syntax.ParseSourceArguments(args)
		if err != nil {
			return nil, fmt.Errorf("failed to parse source command arguments: %w", err)
		}
		result.SourceArgs = sourceArgs
	case DebuggerCommandCurrentSource:
		currentSourceArgs, err := syntax.ParseCurrentSourceArguments(args)
		if err != nil {
			return nil, fmt.Errorf("failed to parse current source command arguments: %w", err)
		}
		result.CurrentSourceArgs = currentSourceArgs
	case DebuggerCommandEval:
		evalArgs, err := syntax.ParseEvalArguments(args)
		if err != nil {
			return nil, fmt.Errorf("failed to parse eval command arguments: %w", err)
		}
		result.EvalArgs = evalArgs
	case DebuggerCommandInfo:
		infoArgs, err := syntax.ParseInfoArguments(args)
		if err != nil {
			return nil, fmt.Errorf("failed to parse info command arguments: %w", err)
		}
		result.InfoArgs = infoArgs
	case DebuggerCommandSymbols:
		symbolsArgs, err := syntax.ParseSymbolsArguments(args)
		if err != nil {
			return nil, fmt.Errorf("failed to parse symbols command arguments: %w", err)
		}
		result.SymbolsArgs = symbolsArgs
	case DebuggerCommandLoadSystem:
		loadSystemArgs, err := syntax.ParseLoadSystemArguments(args)
		if err != nil {
			return nil, fmt.Errorf("failed to parse load system command arguments: %w", err)
		}
		result.LoadSystemArgs = loadSystemArgs
	case DebuggerCommandLoadProgram:
		loadProgramArgs, err := syntax.ParseLoadProgramArguments(args)
		if err != nil {
			return nil, fmt.Errorf("failed to parse load program command arguments: %w", err)
		}
		result.LoadProgramArgs = loadProgramArgs
	case DebuggerCommandLoadSystemFromFile:
		loadSystemFromFileArgs, err := syntax.ParseLoadSystemFromFileArguments(args)
		if err != nil {
			return nil, fmt.Errorf("failed to parse load system from file command arguments: %w", err)
		}
		result.LoadSystemFromFileArgs = loadSystemFromFileArgs
	case DebuggerCommandLoadProgramFromFile:
		loadProgramFromFileArgs, err := syntax.ParseLoadProgramFromFileArguments(args)
		if err != nil {
			return nil, fmt.Errorf("failed to parse load program from file command arguments: %w", err)
		}
		result.LoadProgramFromFileArgs = loadProgramFromFileArgs
	case DebuggerCommandLoadRuntime:
		loadRuntimeArgs, err := syntax.ParseLoadRuntimeArguments(args)
		if err != nil {
			return nil, fmt.Errorf("failed to parse load runtime command arguments: %w", err)
		}
		result.LoadRuntimeArgs = loadRuntimeArgs
	case DebuggerCommandLoad:
		loadArgs, err := syntax.ParseLoadArguments(args)
		if err != nil {
			return nil, fmt.Errorf("failed to parse load command arguments: %w", err)
		}
		result.LoadArgs = loadArgs
	}

	return result, nil
}
