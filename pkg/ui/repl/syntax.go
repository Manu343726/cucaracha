package repl

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/ui/debugger"
	"github.com/Manu343726/cucaracha/pkg/utils"
)

// =============================================================================
// REPLSyntaxDescriptor - Interactive REPL syntax
// =============================================================================

// REPLSyntaxDescriptor formats commands for interactive REPL
// Example: step count:5
type REPLSyntax struct{}

// Ensure REPLSyntaxDescriptor implements CommandSyntaxDescriptor
var _ debugger.CommandsSyntax = (*REPLSyntax)(nil)

func (r REPLSyntax) Name() string {
	return "repl"
}

func (r REPLSyntax) FormatCommandName(command debugger.DebuggerCommandId) string {
	return utils.KebabCase(strings.TrimPrefix(command.String(), "DebuggerCommand"))
}

func (r REPLSyntax) FormatArgumentName(name string, isRequired bool) string {
	// REPL uses simple name syntax
	return utils.KebabCase(name)
}

func (r REPLSyntax) FormatArgumentValue(value interface{}) string {
	return fmt.Sprintf("%v", value)
}

func (r REPLSyntax) ParseOptionalArgumentName(arg any) (string, any, error) {
	strArg, ok := arg.(string)
	if !ok {
		return "", nil, fmt.Errorf("optional arguments must be strings in the format 'name=value' or 'name' for boolean flags")
	}

	parts := strings.SplitN(strArg, "=", 2)
	name := parts[0]

	if len(parts) == 2 {
		return name, parts[1], nil
	} else {
		return name, true, nil // Boolean flag, set to true
	}
}

func (r REPLSyntax) ParseOptionalArguments(args any, expectedArgs map[string]debugger.SyntaxOptionValueParser) (map[string]interface{}, error) {
	strArgs, ok := args.([]string)
	if !ok {
		return nil, fmt.Errorf("optional arguments must be a slice of strings in the format 'name:value'")
	}

	result := make(map[string]interface{})

	for _, arg := range strArgs {
		name, value, err := debugger.ParseSingleOptionalArgument(r, arg, expectedArgs)
		if err != nil {
			return nil, err
		}
		result[name] = value
	}

	return result, nil
}

func (r REPLSyntax) ParseSourceLocation(loc interface{}) (*debugger.SourceLocation, error) {
	locStr, ok := loc.(string)
	if !ok {
		return nil, fmt.Errorf("source location must be a string in the format 'file:line'")
	}

	// The syntax for source locations in the REPL is "file:line"
	var file string
	var line int
	n, err := fmt.Sscanf(locStr, "%[^:]:%d", &file, &line)
	if err != nil {
		return nil, fmt.Errorf("invalid source location format: %w", err)
	}
	if n != 2 {
		return nil, fmt.Errorf("invalid source location format: expected 'file:line'")
	}
	return &debugger.SourceLocation{
		File: file,
		Line: line,
	}, nil
}

func (r REPLSyntax) ParseCommand(input interface{}) (*debugger.DebuggerCommand, error) {
	inputArgs, ok := input.([]string)
	if !ok {
		return nil, fmt.Errorf("command input must be a slice of strings")
	}

	if len(inputArgs) == 0 {
		return nil, fmt.Errorf("command input cannot be empty")
	}

	// Find the command by matching against formatted command names using the enum values map
	// This is case-insensitive and dynamic based on all available command enum values
	inputCmdLower := strings.ToLower(inputArgs[0])
	var command debugger.DebuggerCommandId
	var found bool

	// Try to find a matching command by checking all valid command IDs from the map
	for cmdId := range debugger.DebuggerCommandIdValues {
		cmdStr := cmdId.String()

		// Format the command name using REPL syntax (removes prefix, kebab-cases)
		formattedName := r.FormatCommandName(cmdId)

		// Try both formatted name and raw lowercased command name for maximum compatibility
		if strings.EqualFold(formattedName, inputArgs[0]) ||
			strings.EqualFold(strings.ToLower(strings.TrimPrefix(cmdStr, "DebuggerCommand")), inputCmdLower) {
			command = cmdId
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("unknown command: %s", inputArgs[0])
	}

	return debugger.ParseCommandArguments(r, command, inputArgs[1:])
}

// =============================================================================
// CommandsArgsParser methods - ordered according to the interface
// =============================================================================

func (r REPLSyntax) ParseStepArguments(args any) (*debugger.StepArgs, error) {
	strArgs, ok := args.([]string)
	if !ok {
		return nil, fmt.Errorf("step command arguments must be a slice of strings")
	}

	optionals, err := r.ParseOptionalArguments(strArgs, map[string]debugger.SyntaxOptionValueParser{
		"stepMode":  utils.UntypeFunction(debugger.StepModeFromString),
		"countMode": utils.UntypeFunction(debugger.StepCountModeFromString),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse step command optional arguments: %w", err)
	}

	stepArgs := &debugger.StepArgs{}

	if stepMode, ok := optionals["stepMode"]; ok {
		stepArgs.StepMode = stepMode.(debugger.StepMode)
	}
	if countMode, ok := optionals["countMode"]; ok {
		stepArgs.CountMode = countMode.(debugger.StepCountMode)
	}

	return stepArgs, nil
}

func (r REPLSyntax) ParseBreakArguments(args any) (*debugger.BreakArgs, error) {
	strArgs, ok := args.([]string)
	if !ok {
		return nil, fmt.Errorf("break command arguments must be a slice of strings")
	}

	if len(strArgs) == 0 {
		return nil, fmt.Errorf("break command requires at least a source location argument")
	}

	if srcLoc, err := r.ParseSourceLocation(strArgs[0]); err == nil {
		if len(strArgs) > 1 {
			return nil, fmt.Errorf("unexpected extra arguments for break command: %v", strArgs[1:])
		}

		return &debugger.BreakArgs{
			SourceLocation: srcLoc,
		}, nil
	}

	// If the first argument is not a valid source location, treat the entire argument list as an address expression
	return &debugger.BreakArgs{
		Address: utils.Ptr(strings.Join(strArgs, " ")),
	}, nil
}

func (r REPLSyntax) ParseWatchArguments(args any) (*debugger.WatchArgs, error) {
	strArgs, ok := args.([]string)
	if !ok {
		return nil, fmt.Errorf("watch command arguments must be a slice of strings")
	}

	if len(strArgs) == 0 {
		return nil, fmt.Errorf("watch command requires at least a start address argument")
	}

	startAddress := strArgs[0]
	var endAddress *string
	var size *string
	var watchType *debugger.WatchpointType

	optionals, err := r.ParseOptionalArguments(strArgs[1:], map[string]debugger.SyntaxOptionValueParser{
		"end":  utils.UntypeFunction(func(s string) (*string, error) { return utils.Ptr(s), nil }),
		"size": utils.UntypeFunction(func(s string) (*string, error) { return utils.Ptr(s), nil }),
		"type": utils.UntypeFunction(func(s string) (*debugger.WatchpointType, error) {
			if wt, err := debugger.WatchpointTypeFromString(s); err != nil {
				return nil, err
			} else {
				return utils.Ptr(wt), nil
			}
		}),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse watch command optional arguments: %w", err)
	}

	if val, ok := optionals["end"]; ok {
		endAddress = val.(*string)
	}
	if val, ok := optionals["size"]; ok {
		size = val.(*string)
	}
	if val, ok := optionals["type"]; ok {
		watchType = val.(*debugger.WatchpointType)
	}

	return &debugger.WatchArgs{
		StartAddress: startAddress,
		EndAddress:   endAddress,
		Size:         size,
		Type:         watchType,
	}, nil
}

func (r REPLSyntax) ParseRemoveBreakpointArguments(args any) (*debugger.RemoveBreakpointArgs, error) {
	strArgs, ok := args.([]string)
	if !ok {
		return nil, fmt.Errorf("remove breakpoint command arguments must be a slice of strings")
	}

	if len(strArgs) == 0 {
		return nil, fmt.Errorf("remove breakpoint command requires a breakpoint ID argument")
	}

	if len(strArgs) > 1 {
		return nil, fmt.Errorf("unexpected extra arguments for remove breakpoint command: %v", strArgs[1:])
	}

	breakpointID, err := strconv.Atoi(strArgs[0])
	if err != nil {
		return nil, fmt.Errorf("invalid breakpoint ID: %w", err)
	}

	return &debugger.RemoveBreakpointArgs{
		ID: breakpointID,
	}, nil
}

func (r REPLSyntax) ParseRemoveWatchpointArguments(args any) (*debugger.RemoveWatchpointArgs, error) {
	strArgs, ok := args.([]string)
	if !ok {
		return nil, fmt.Errorf("remove watchpoint command arguments must be a slice of strings")
	}

	if len(strArgs) == 0 {
		return nil, fmt.Errorf("remove watchpoint command requires a watchpoint ID argument")
	}

	if len(strArgs) > 1 {
		return nil, fmt.Errorf("unexpected extra arguments for remove watchpoint command: %v", strArgs[1:])
	}

	watchpointID, err := strconv.Atoi(strArgs[0])
	if err != nil {
		return nil, fmt.Errorf("invalid watchpoint ID: %w", err)
	}

	return &debugger.RemoveWatchpointArgs{
		ID: watchpointID,
	}, nil
}

func (r REPLSyntax) ParseDisasmArguments(args any) (*debugger.DisasmArgs, error) {
	strArgs, ok := args.([]string)
	if !ok {
		return nil, fmt.Errorf("disasm command arguments must be a slice of strings")
	}

	if len(strArgs) == 0 {
		return nil, fmt.Errorf("disasm command requires an address expression argument")
	}

	// Treat the first argument as an address expression, the second (if there is) as a count expression
	addrExpression := strArgs[0]
	var countExpression *string

	if len(strArgs) > 1 {
		if len(strArgs) > 2 {
			return nil, fmt.Errorf("unexpected extra arguments for disasm command: %v", strArgs[2:])
		}

		countExpression = utils.Ptr(strArgs[1])
	}

	return &debugger.DisasmArgs{
		Address:   addrExpression,
		CountExpr: countExpression,
	}, nil
}

func (r REPLSyntax) ParseMemoryArguments(args any) (*debugger.MemoryArgs, error) {
	strArgs, ok := args.([]string)
	if !ok {
		return nil, fmt.Errorf("memory command arguments must be a slice of strings")
	}

	if len(strArgs) == 0 {
		return nil, fmt.Errorf("memory command requires an address expression argument")
	}

	// Treat the first argument as an address expression, the second (if there is) as a count expression
	addrExpression := strArgs[0]
	var countExpression *string

	if len(strArgs) > 1 {
		if len(strArgs) > 2 {
			return nil, fmt.Errorf("unexpected extra arguments for memory command: %v", strArgs[2:])
		}

		countExpression = utils.Ptr(strArgs[1])
	}

	return &debugger.MemoryArgs{
		AddressExpr: addrExpression,
		CountExpr:   countExpression,
	}, nil
}

func (r REPLSyntax) ParseSourceArguments(args any) (*debugger.SourceArgs, error) {
	strArgs, ok := args.([]string)
	if !ok {
		return nil, fmt.Errorf("source command arguments must be a slice of strings")
	}

	if len(strArgs) == 0 {
		return nil, fmt.Errorf("source command requires a file path argument")
	}

	location, err := r.ParseSourceLocation(strArgs[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse source location: %w", err)
	}

	optionals, err := r.ParseOptionalArguments(strArgs[1:], map[string]debugger.SyntaxOptionValueParser{
		"contextMode": utils.UntypeFunction(debugger.SourceContextModeFromString),
		"lines": utils.UntypeFunction(func(s string) (int, error) {
			if n, err := strconv.Atoi(s); err != nil {
				return 0, err
			} else {
				return n, nil
			}
		}),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse source command optional arguments: %w", err)
	}

	lines := 10
	if val, ok := optionals["lines"]; ok {
		lines = val.(int)
	}

	contextMode := debugger.SourceContextTop
	if val, ok := optionals["contextMode"]; ok {
		contextMode = val.(debugger.SourceContextMode)
	}

	return &debugger.SourceArgs{
		Location:     location,
		ContextLines: lines,
		ContextMode:  contextMode,
	}, nil

}

func (r REPLSyntax) ParseCurrentSourceArguments(args any) (*debugger.CurrentSourceArgs, error) {
	strArgs, ok := args.([]string)
	if !ok {
		return nil, fmt.Errorf("current-source command arguments must be a slice of strings")
	}

	optionals, err := r.ParseOptionalArguments(strArgs, map[string]debugger.SyntaxOptionValueParser{
		"contextMode": utils.UntypeFunction(debugger.SourceContextModeFromString),
		"lines": utils.UntypeFunction(func(s string) (int, error) {
			if n, err := strconv.Atoi(s); err != nil {
				return 0, err
			} else {
				return n, nil
			}
		}),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse current-source command optional arguments: %w", err)
	}

	lines := 10
	if val, ok := optionals["lines"]; ok {
		lines = val.(int)
	}

	contextMode := debugger.SourceContextTop
	if val, ok := optionals["contextMode"]; ok {
		contextMode = val.(debugger.SourceContextMode)
	}

	return &debugger.CurrentSourceArgs{
		ContextLines: lines,
		ContextMode:  contextMode,
	}, nil
}

func (r REPLSyntax) ParseEvalArguments(args any) (*debugger.EvalArgs, error) {
	strArgs, ok := args.([]string)
	if !ok {
		return nil, fmt.Errorf("eval command arguments must be a slice of strings")
	}

	if len(strArgs) == 0 {
		return nil, fmt.Errorf("eval command requires an expression argument")
	}

	// Treat the entire argument list as an expression
	return &debugger.EvalArgs{
		Expression: strings.Join(strArgs, " "),
	}, nil
}

func (r REPLSyntax) ParseInfoArguments(args any) (*debugger.InfoArgs, error) {
	strArgs, ok := args.([]string)
	if !ok {
		return nil, fmt.Errorf("info command arguments must be a slice of strings")
	}

	var infoType debugger.InfoType
	if len(strArgs) == 0 {
		infoType = debugger.InfoTypeGeneral
	} else {
		var err error
		infoType, err = debugger.InfoTypeFromString(strArgs[0])
		if err != nil {
			return nil, fmt.Errorf("invalid info type: %w", err)
		}
	}

	return &debugger.InfoArgs{
		Type: infoType,
	}, nil
}

func (r REPLSyntax) ParseSymbolsArguments(args any) (*debugger.SymbolsArgs, error) {
	strArgs, ok := args.([]string)
	if !ok {
		return nil, fmt.Errorf("symbols command arguments must be a slice of strings")
	}

	var symbolName *string

	if len(strArgs) > 0 {
		symbolName = utils.Ptr(strings.Join(strArgs, " "))
	}

	return &debugger.SymbolsArgs{
		SymbolName: symbolName,
	}, nil
}

func (r REPLSyntax) ParseLoadSystemArguments(args any) (*debugger.LoadSystemArgs, error) {
	// LoadSystem takes no arguments (uses embedded default)
	return &debugger.LoadSystemArgs{}, nil
}

func (r REPLSyntax) ParseLoadProgramArguments(args any) (*debugger.LoadProgramArgs, error) {
	strArgs, ok := args.([]string)
	if !ok {
		return nil, fmt.Errorf("load program command arguments must be a slice of strings")
	}

	if len(strArgs) == 0 {
		return nil, fmt.Errorf("load program requires a file path argument")
	}

	return &debugger.LoadProgramArgs{
		FilePath: strArgs[0],
		// AutoBuildClang and ForceRebuildClang default to nil (will use defaults in implementation)
	}, nil
}

func (r REPLSyntax) ParseLoadSystemFromFileArguments(args any) (*debugger.LoadSystemFromFileArgs, error) {
	strArgs, ok := args.([]string)
	if !ok {
		return nil, fmt.Errorf("load-system-from-file command arguments must be a slice of strings")
	}

	if len(strArgs) == 0 {
		return nil, fmt.Errorf("load-system-from-file command requires a file path argument")
	}

	filePath := strings.Join(strArgs, " ")

	return &debugger.LoadSystemFromFileArgs{
		FilePath: filePath,
	}, nil
}

func (r REPLSyntax) ParseLoadProgramFromFileArguments(args any) (*debugger.LoadProgramFromFileArgs, error) {
	strArgs, ok := args.([]string)
	if !ok {
		return nil, fmt.Errorf("load-program-from-file command arguments must be a slice of strings")
	}

	if len(strArgs) == 0 {
		return nil, fmt.Errorf("load-program-from-file command requires a file path argument")
	}

	filePath := strings.Join(strArgs, " ")

	return &debugger.LoadProgramFromFileArgs{
		FilePath: filePath,
	}, nil
}

func (r REPLSyntax) ParseLoadRuntimeArguments(args any) (*debugger.LoadRuntimeArgs, error) {
	strArgs, ok := args.([]string)
	if !ok {
		return nil, fmt.Errorf("load-runtime command arguments must be a slice of strings")
	}

	if len(strArgs) == 0 {
		return nil, fmt.Errorf("load-runtime command requires a file path argument")
	}

	if len(strArgs) > 1 {
		return nil, fmt.Errorf("unexpected extra arguments for load-runtime command: %v", strArgs[1:])
	}

	runtime, err := debugger.RuntimeTypeFromString(strArgs[0])
	if err != nil {
		return nil, fmt.Errorf("invalid runtime type: %w", err)
	}

	return &debugger.LoadRuntimeArgs{
		Runtime: runtime,
	}, nil
}

func (r REPLSyntax) ParseLoadArguments(args any) (*debugger.LoadArgs, error) {
	strArgs, ok := args.([]string)
	if !ok {
		return nil, fmt.Errorf("load command arguments must be a slice of strings")
	}

	if len(strArgs) == 0 {
		return nil, fmt.Errorf("load command requires a file path argument")
	}

	var fullDescriptorPath *string
	var systemConfigPath *string
	var programPath *string
	var runtime *debugger.RuntimeType

	if len(strArgs) == 1 {
		fullDescriptorPath = utils.Ptr(strArgs[0])
	}

	if len(strArgs) >= 2 {
		systemConfigPath = utils.Ptr(strArgs[0])
		programPath = utils.Ptr(strArgs[1])
	}

	if len(strArgs) == 3 {
		runtimeVal, err := debugger.RuntimeTypeFromString(strArgs[2])
		if err != nil {
			return nil, fmt.Errorf("invalid runtime type: %w", err)
		}
		runtime = utils.Ptr(runtimeVal)
	}

	return &debugger.LoadArgs{
		FullDescriptorPath: fullDescriptorPath,
		SystemConfigPath:   systemConfigPath,
		ProgramPath:        programPath,
		Runtime:            runtime,
	}, nil
}

// GetREPLSyntax returns the REPL syntax descriptor instance
func GetREPLSyntax() debugger.CommandsSyntax {
	return REPLSyntax{}
}
