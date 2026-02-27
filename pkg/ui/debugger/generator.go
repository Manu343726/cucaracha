//go:build ignore
// +build ignore

// Generator for debugger command infrastructure.
// This is a code generation tool invoked via go:generate.
// Run with: go run generator.go -out <output> -api <api.go> ...
//
// This file is compiled separately and excluded from normal package builds
// due to the //go:build ignore directive above.
//
// Uses the reflect and codegen APIs to generate code from the DebuggerCommands interface.

package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/Manu343726/cucaracha/pkg/codegen"
	"github.com/Manu343726/cucaracha/pkg/reflect"
	"github.com/Manu343726/cucaracha/pkg/ui/debugger"
	"github.com/Manu343726/cucaracha/pkg/utils"
	"github.com/Manu343726/cucaracha/pkg/utils/logging"
)

func main() {
	outputFile := flag.String("out", "", "output file path for Execute method")
	apiPath := flag.String("api", "", "path to api.go file (DebuggerCommands interface)")
	flag.Parse()

	// Set up logging to stdout with DEBUG level
	stdoutSink := logging.NewTextSink("generator", os.Stdout, slog.LevelDebug)
	logging.DefaultRegistry().RegisterSink(stdoutSink)
	rootLogger := logging.NewRegisteredLogger("cucaracha", stdoutSink)
	logging.DefaultRegistry().RegisterLogger(rootLogger)

	if *outputFile == "" || *apiPath == "" {
		fmt.Fprintf(os.Stderr, "Usage: generator -out <execute_output> -api <api.go path> [-docs-out <docs_output>] [-enum-out <enum_output>] [-structs-out <structs_output>]\n")
		os.Exit(1)
	}

	// Parse the API file's package directory via reflect API
	apiDir := filepath.Dir(*apiPath)
	pkg, err := reflect.ParsePackage(apiDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing package: %v\n", err)
		os.Exit(1)
	}

	// Extract DebuggerCommands interface
	debuggerCommandsIface, ok := pkg.Types["DebuggerCommands"]
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: DebuggerCommands interface not found in package\n")
		os.Exit(1)
	}

	if debuggerCommandsIface.Kind != reflect.TypeKindInterface {
		fmt.Fprintf(os.Stderr, "Error: DebuggerCommands is not an interface\n")
		os.Exit(1)
	}

	// Extract command information from interface methods
	commands, err := extractCommands(debuggerCommandsIface, pkg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error extracting commands: %v\n", err)
		os.Exit(1)
	}

	// Generate all code in a single file
	if err := generate(*outputFile, commands, pkg); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating code: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Generated code in %s\n", *outputFile)
}

// CommandMetadata represents extracted command information from interface methods
type CommandMetadata struct {
	MethodName  string                            // e.g., "Step"
	CommandID   string                            // e.g., "DebuggerCommandStep"
	CommandName string                            // e.g., "step" (camelCase for JSON)
	Doc         string                            // Documentation comment
	ArgsType    *reflect.TypeReference            // e.g., "*StepArgs"
	ResultType  *reflect.TypeReference            // e.g., "*ExecutionResult"
	Arguments   []*debugger.ArgumentDocumentation // Documentation for arguments
	Results     []*debugger.ArgumentDocumentation // Documentation for results
	Index       int                               // Command ordered index (0, 1, 2, ...)
}

// extractCommands extracts command metadata from interface methods using reflect API
func extractCommands(iface *reflect.Type, pkg *reflect.Package) ([]CommandMetadata, error) {
	var commands []CommandMetadata

	for _, method := range iface.Methods {
		// Validate method signature
		if len(method.Results) != 1 {
			return nil, fmt.Errorf("method %s must have one return type", method.Name)
		}
		if len(method.Args) > 1 {
			return nil, fmt.Errorf("method %s must have at most one parameter", method.Name)
		}

		// Extract argument type
		var argsType *reflect.TypeReference
		var arguments []*debugger.ArgumentDocumentation
		if len(method.Args) > 0 {
			argsType = method.Args[0].Type
			expectedArgsType := fmt.Sprintf("*%sArgs", method.Name)
			if argsType.Name != expectedArgsType {
				return nil, fmt.Errorf("method %s: parameter type %q does not match convention, expected %q",
					method.Name, argsType.Name, expectedArgsType)
			}

			// Extract argument documentation from the ArgsType
			var err error
			arguments, err = extractArgumentsDocs(pkg, argsType.Name)
			if err != nil {
				return nil, fmt.Errorf("method %s: failed to extract argument docs: %w", method.Name, err)
			}
		}

		resultType := method.Results[0].Type
		if !strings.HasSuffix(resultType.Name, "Result") {
			return nil, fmt.Errorf("method %s: return type %q does not follow Result naming convention",
				method.Name, resultType.Name)
		}

		// Extract result documentation
		results, err := extractResultsDocs(pkg, resultType.Name)
		if err != nil {
			return nil, fmt.Errorf("method %s: failed to extract result docs: %w", method.Name, err)
		}

		cmd := CommandMetadata{
			MethodName:  method.Name,
			CommandID:   methodNameToCommandIdEnumConstant(method.Name),
			CommandName: methodNameToCommandName(method.Name),
			Doc:         method.Doc,
			ArgsType:    argsType,
			ResultType:  resultType,
			Arguments:   arguments,
			Results:     results,
			Index:       len(commands),
		}

		commands = append(commands, cmd)
	}

	return commands, nil
}

// extractArgumentsDocs extracts documentation for command arguments from the ArgsType
func extractArgumentsDocs(pkg *reflect.Package, argsTypeName string) ([]*debugger.ArgumentDocumentation, error) {
	// Remove pointer prefix if present
	typeName := strings.TrimPrefix(argsTypeName, "*")

	typ, ok := pkg.Types[typeName]
	if !ok {
		// If the type doesn't exist, it might be an external type - return empty docs
		return []*debugger.ArgumentDocumentation{}, nil
	}

	if typ.Kind != reflect.TypeKindStruct {
		return nil, fmt.Errorf("argument type %s is not a struct", typeName)
	}

	var args []*debugger.ArgumentDocumentation
	for _, field := range typ.Fields {
		args = append(args, &debugger.ArgumentDocumentation{
			FieldName:  field.Name,
			TypeName:   field.Type.Name,
			IsRequired: true, // All arguments are considered required unless explicitly marked
			Summary:    field.Doc,
		})
	}

	return args, nil
}

// extractResultsDocs extracts documentation for command results from the ResultType
func extractResultsDocs(pkg *reflect.Package, resultTypeName string) ([]*debugger.ArgumentDocumentation, error) {
	// Remove pointer prefix if present
	typeName := strings.TrimPrefix(resultTypeName, "*")

	typ, ok := pkg.Types[typeName]
	if !ok {
		// If the type doesn't exist, it might be an external type - return empty docs
		return []*debugger.ArgumentDocumentation{}, nil
	}

	if typ.Kind != reflect.TypeKindStruct {
		return nil, fmt.Errorf("result type %s is not a struct", typeName)
	}

	var results []*debugger.ArgumentDocumentation
	for _, field := range typ.Fields {
		results = append(results, &debugger.ArgumentDocumentation{
			FieldName: field.Name,
			TypeName:  field.Type.Name,
			Summary:   field.Doc,
		})
	}

	return results, nil
}

// For a method name like "Step", this generates "DebuggerCommandStep"
func methodNameToCommandIdEnumConstant(methodName string) string {
	return "DebuggerCommand" + methodName
}

// methodNameToCommandName converts a method name to a command name (for JSON)
// e.g., "Step" -> "step", "RemoveBreakpoint" -> "removeBreakpoint"
func methodNameToCommandName(methodName string) string {
	if len(methodName) == 0 {
		return methodName
	}
	runes := []rune(methodName)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// Generates everything command related in one go (Execute method, command structs, CommandId enum, and documentation)
func generate(outputPath string, commands []CommandMetadata, pkg *reflect.Package) error {
	// Build Documentation object from commands
	docs := make(debugger.Documentation)
	for _, cmd := range commands {
		arguments := make(map[string]*debugger.ArgumentDocumentation)
		for _, arg := range cmd.Arguments {
			arguments[arg.FieldName] = arg
		}

		results := make(map[string]*debugger.ArgumentDocumentation)
		for _, res := range cmd.Results {
			results[res.FieldName] = res
		}

		docs[debugger.DebuggerCommandId(cmd.Index)] = &debugger.CommandDocumentation{
			CommandID:   debugger.DebuggerCommandId(cmd.Index),
			CommandName: cmd.CommandName,
			MethodName:  cmd.MethodName,
			Summary:     cmd.Doc,
			Arguments:   arguments,
			Results:     results,
		}
	}

	file := codegen.NewFile("debugger").AddType(
		reflect.Struct(
			"DebuggerCommand",
			append([]*reflect.Field{
				{Name: "Id", Type: reflect.MakeTypeReference(reflect.TypeUint64), Doc: "Unique ID for this command instance", Tag: `json:"id"`},
				{Name: "Command", Type: &reflect.TypeReference{Name: "DebuggerCommandId"}, Doc: "Type of command", Tag: `json:"command"`},
			}, utils.MapIf(commands, func(cmd CommandMetadata) (*reflect.Field, bool) {
				if cmd.ArgsType == nil {
					return nil, false
				}

				fieldName := cmd.MethodName + "Args"
				jsonName := strings.ToLower(fieldName)
				return &reflect.Field{
					Name: fieldName,
					Type: cmd.ArgsType,
					Doc:  fmt.Sprintf("Command arguments for %s command", cmd.MethodName),
					Tag:  fmt.Sprintf(`json:"%s"`, jsonName),
				}, true
			})...),
		).WithDoc("DebuggerCommand represents a debugger command with its arguments")).AddType(
		reflect.Struct(
			"DebuggerCommandResult",
			append([]*reflect.Field{
				{Name: "Id", Type: reflect.MakeTypeReference(reflect.TypeUint64), Doc: "Unique ID for this command instance", Tag: `json:"id"`},
				{Name: "Command", Type: &reflect.TypeReference{Name: "DebuggerCommandId"}, Doc: "Command identifier", Tag: `json:"command"`},
			}, utils.Map(commands, func(cmd CommandMetadata) *reflect.Field {
				fieldName := cmd.MethodName + "Result"
				jsonName := strings.ToLower(fieldName)
				return &reflect.Field{
					Name: fieldName,
					Type: cmd.ResultType,
					Doc:  fmt.Sprintf("Result of %s command", cmd.MethodName),
					Tag:  fmt.Sprintf(`json:"%s"`, jsonName),
				}
			})...),
		).WithDoc("DebuggerCommandResult represents the result of a debugger command")).AddEnum(
		&reflect.Enum{
			Type: reflect.MakeTypeReference(reflect.Typedef("DebuggerCommandId", reflect.TypeInt).WithDoc("DebuggerCommandId is an enum of all supported debugger commands")),
			Values: utils.Map(commands, func(cmd CommandMetadata) *reflect.Constant {
				return &reflect.Constant{
					Doc:  cmd.Doc,
					Name: cmd.CommandID,
					Value: &reflect.Value{
						Value: reflect.NewValue(cmd.Index),
					},
				}
			}),
		}).AddVariable(
		func() *reflect.Variable {
			// Create the proper type structure for Documentation typedef
			// Documentation = map[DebuggerCommandId]*CommandDocumentation

			// First create the underlying map type
			// Note: DebuggerCommandId is a typedef for int (the enum type)
			keyType := &reflect.Type{
				Name: "DebuggerCommandId",
				Kind: reflect.TypeKindTypedef,
				OriginalType: &reflect.TypeReference{
					Name: "int",
					Type: reflect.GetBasicType("int"),
				},
			}
			valueType := &reflect.Type{
				Name: "*CommandDocumentation",
				Kind: reflect.TypeKindPointer,
				Elem: &reflect.TypeReference{
					Name: "CommandDocumentation",
					Type: &reflect.Type{
						Name: "CommandDocumentation",
						Kind: reflect.TypeKindStruct,
					},
				},
			}
			mapType := &reflect.Type{
				Name: "map[DebuggerCommandId]*CommandDocumentation",
				Kind: reflect.TypeKindMap,
				Key: &reflect.TypeReference{
					Name: "DebuggerCommandId",
					Type: keyType,
				},
				Value: &reflect.TypeReference{
					Name: "*CommandDocumentation",
					Type: valueType,
				},
			}

			// Now create the typedef wrapper
			docType := &reflect.Type{
				Name: "Documentation",
				Kind: reflect.TypeKindTypedef,
				OriginalType: &reflect.TypeReference{
					Name: "map[DebuggerCommandId]*CommandDocumentation",
					Type: mapType,
				},
			}
			return reflect.NewVariableWithType("CommandsDocumentation", docs, docType).WithDoc("CommandsDocumentation contains the complete documentation for all debugger commands")
		}(),
	).AddMethod(&codegen.MethodImplementation{
		Method: &reflect.Method{
			Name: "Execute",
			Doc:  "Execute executes a debugger command and returns the result",
			Receiver: &reflect.Parameter{
				Name: "c",
				Type: &reflect.TypeReference{Name: "*commandBasedDebuggerAdapter"},
			},
			Args: []*reflect.Parameter{
				{Name: "cmd", Type: &reflect.TypeReference{Name: "*DebuggerCommand"}},
			},
			Results: []*reflect.Parameter{
				{Name: "result", Type: &reflect.TypeReference{Name: "*DebuggerCommandResult"}},
				{Name: "err", Type: &reflect.TypeReference{Name: "error"}},
			},
		},
		BodyTemplate: `
switch(cmd.Command) {
{{- range .commands }}
case {{.CommandID}}:
{{- if .ArgsType}}
	callResult := c.debugger.{{.MethodName}}(cmd.{{.MethodName}}Args)
{{- else}}
	callResult := c.debugger.{{.MethodName}}()
{{- end}}
	result = &DebuggerCommandResult{
		Id: cmd.Id,
		Command: cmd.Command,
		{{.MethodName}}Result: callResult,
	}
{{- end }}
default:
	err = fmt.Errorf("unknown command ID: %d", cmd.Command)
}

return
`,
		TemplateContext: map[string]interface{}{
			"commands": commands,
		}})

	return codegen.Generate(file, outputPath)
}
