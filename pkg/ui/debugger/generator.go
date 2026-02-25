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
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/Manu343726/cucaracha/pkg/codegen"
	"github.com/Manu343726/cucaracha/pkg/reflect"
	"github.com/Manu343726/cucaracha/pkg/utils"
)

func main() {
	outputFile := flag.String("out", "", "output file path for Execute method")
	apiPath := flag.String("api", "", "path to api.go file (DebuggerCommands interface)")
	flag.Parse()

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
	commands, err := extractCommands(debuggerCommandsIface)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error extracting commands: %v\n", err)
		os.Exit(1)
	}

	// Generate all code in a single file
	if err := generate(*outputFile, commands); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating code: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Generated code in %s\n", *outputFile)
}

// CommandMetadata represents extracted command information from interface methods
type CommandMetadata struct {
	MethodName  string                 // e.g., "Step"
	CommandID   string                 // e.g., "DebuggerCommandStep"
	CommandName string                 // e.g., "step" (camelCase for JSON)
	Doc         string                 // Documentation comment
	ArgsType    *reflect.TypeReference // e.g., "*StepArgs"
	ResultType  *reflect.TypeReference // e.g., "*ExecutionResult"
	Index       int                    // Command ordered index (0, 1, 2, ...)
}

// extractCommands extracts command metadata from interface methods using reflect API
func extractCommands(iface *reflect.Type) ([]CommandMetadata, error) {
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
		if len(method.Args) > 0 {
			argsType = method.Args[0].Type
			expectedArgsType := fmt.Sprintf("*%sArgs", method.Name)
			if argsType.Name != expectedArgsType {
				return nil, fmt.Errorf("method %s: parameter type %q does not match convention, expected %q",
					method.Name, argsType.Name, expectedArgsType)
			}
		}

		resultType := method.Results[0].Type
		if !strings.HasSuffix(resultType.Name, "Result") {
			return nil, fmt.Errorf("method %s: return type %q does not follow Result naming convention",
				method.Name, resultType.Name)
		}

		cmd := CommandMetadata{
			MethodName:  method.Name,
			CommandID:   methodNameToCommandIdEnumConstant(method.Name),
			CommandName: methodNameToCommandName(method.Name),
			Doc:         method.Doc,
			ArgsType:    argsType,
			ResultType:  resultType,
			Index:       len(commands),
		}

		commands = append(commands, cmd)
	}

	return commands, nil
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

// Generates everything command related in one go (Execute method, command structs, CommandId enum, and documentation schema)
func generate(outputPath string, commands []CommandMetadata) error {
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
		}).AddMethod(&codegen.MethodImplementation{
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
