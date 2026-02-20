package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"unicode"
)

type CommandInfo struct {
	MethodName      string // e.g., "Disasm"
	CommandID       string // e.g., "DebuggerCommandDisasm"
	CommandName     string // e.g., "disasm" (JSON command name)
	Comments        string // Documentation comment
	ArgsType        string // e.g., "*DisasmArgs" (or "" if no args)
	ResultType      string // e.g., "*DisasmResult"
	ResultFieldName string // e.g., "DisasmResult"
	ArgsFieldName   string // e.g., "cmd.DisasmArgs" (or "" if no args)
}

func main() {
	outputFile := flag.String("out", "", "output file path for Execute method")
	apiPath := flag.String("api", "", "path to api.go file (DebuggerCommands interface)")
	docsOutput := flag.String("docs-out", "", "output file path for documentation schema")
	enumOutput := flag.String("enum-out", "", "output file path for DebuggerCommandId enum")
	structsOutput := flag.String("structs-out", "", "output file path for DebuggerCommand and DebuggerCommandResult structs")
	flag.Parse()

	if *outputFile == "" || *apiPath == "" {
		fmt.Fprintf(os.Stderr, "Usage: generator -out <execute_output> -api <api.go path> [-docs-out <docs_output>] [-enum-out <enum_output>] [-structs-out <structs_output>]\n")
		os.Exit(1)
	}

	// Parse the API file to extract DebuggerCommands interface methods
	commands, err := parseDebuggerInterface(*apiPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing API file: %v\n", err)
		os.Exit(1)
	}

	// Generate code for the Execute() method
	output, err := generateExecuteMethod(commands)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating Execute method: %v\n", err)
		os.Exit(1)
	}

	// Write Execute method output
	if err := os.WriteFile(*outputFile, output, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing execute output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated Execute method in %s\n", *outputFile)

	// Generate enum if requested
	if *enumOutput != "" {
		enumCode := generateEnumConstants(commands)
		if err := os.WriteFile(*enumOutput, []byte(enumCode), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing enum output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated enum constants in %s\n", *enumOutput)
	}

	// Generate structs if requested
	if *structsOutput != "" {
		structsCode := generateCommandStructs(commands)
		if err := os.WriteFile(*structsOutput, []byte(structsCode), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing structs output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated command structs in %s\n", *structsOutput)
	}

	// Generate documentation schema if requested
	if *docsOutput != "" {
		docsCode := generateDocsSchema(commands)
		if err := os.WriteFile(*docsOutput, []byte(docsCode), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing docs output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated documentation schema in %s\n", *docsOutput)
	}
}

// parseDebuggerInterface extracts all methods from the DebuggerCommands interface
// and generates command information based on naming conventions
func parseDebuggerInterface(apiPath string) ([]CommandInfo, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, apiPath, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var commands []CommandInfo

	// Find the DebuggerCommands interface
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != "DebuggerCommands" {
				continue
			}

			interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok {
				continue
			}

			// Extract methods from the DebuggerCommands interface
			for _, method := range interfaceType.Methods.List {
				// Skip embedded interfaces
				if len(method.Names) == 0 {
					continue
				}

				methodName := method.Names[0].Name

				// Skip excluded methods
				if !shouldIncludeInExecute(methodName) {
					continue
				}

				// Extract method signature
				funcType, ok := method.Type.(*ast.FuncType)
				if !ok {
					return nil, fmt.Errorf("failed to extract function type for method %s", methodName)
				}

				// Extract argument type if present
				var argsType string
				var argsFieldName string
				if funcType.Params != nil && len(funcType.Params.List) > 0 {
					param := funcType.Params.List[0]
					argsType = typeToString(param.Type)
					// Expected format: "*<MethodName>Args"
					expectedArgsType := fmt.Sprintf("*%sArgs", methodName)
					if argsType != expectedArgsType {
						return nil, fmt.Errorf("method %s: parameter type %q does not match naming convention, expected %q", methodName, argsType, expectedArgsType)
					}
					argsFieldName = fmt.Sprintf("cmd.%s", baseTypeName(argsType))
				}

				// Extract return type
				var resultType string
				var resultFieldName string
				if funcType.Results != nil && len(funcType.Results.List) > 0 {
					result := funcType.Results.List[0]
					resultType = typeToString(result.Type)

					// Check if it matches the convention
					if !strings.HasSuffix(resultType, "Result") {
						return nil, fmt.Errorf("method %s: return type %q does not follow Result naming convention", methodName, resultType)
					}

					// Field name should be <MethodName>Result for consistency with DebuggerCommandResult struct
					resultFieldName = methodName + "Result"
				}

				// Generate command information from method name
				cmd := CommandInfo{
					MethodName:      methodName,
					CommandID:       "DebuggerCommand" + methodName,
					CommandName:     methodNameToCommandName(methodName),
					ResultFieldName: resultFieldName,
					ArgsFieldName:   argsFieldName,
					ArgsType:        argsType,
					ResultType:      resultType,
				}

				// Get documentation if available
				if method.Doc != nil {
					cmd.Comments = method.Doc.Text()
				}

				commands = append(commands, cmd)
			}
		}
	}

	return commands, nil
}

// typeToString converts an AST type to its string representation
func typeToString(expr ast.Expr) string {
	switch v := expr.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.StarExpr:
		return "*" + typeToString(v.X)
	case *ast.SelectorExpr:
		return typeToString(v.X) + "." + v.Sel.Name
	default:
		return ""
	}
}

// baseTypeName extracts the base type name from a pointer type string
// e.g., "*StepResult" -> "StepResult"
func baseTypeName(typeName string) string {
	return strings.TrimPrefix(typeName, "*")
}

func shouldIncludeInExecute(methodName string) bool {
	// Exclude methods that are callbacks or not command-based
	excluded := map[string]bool{
		"SetEventCallback": true,
	}
	return !excluded[methodName]
}

// methodNameToCommandName converts a method name to a command name (for JSON)
// e.g., "Disasm" -> "disasm", "RemoveBreakpoint" -> "removeBreakpoint"
func methodNameToCommandName(methodName string) string {
	if len(methodName) == 0 {
		return methodName
	}
	// Convert first letter to lowercase
	runes := []rune(methodName)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// generateExecuteMethod generates the Execute() method implementation
func generateExecuteMethod(commands []CommandInfo) ([]byte, error) {
	var buf bytes.Buffer

	// Write header
	buf.WriteString(`// Code generated by debuggen; DO NOT EDIT.

package debugger

import (
	"fmt"
)

// Execute processes a DebuggerCommand and returns the result
// This method is generated from the Debugger interface
func (d *commandBasedDebuggerAdapter) Execute(cmd *DebuggerCommand) (*DebuggerCommandResult, error) {
	switch cmd.Command {
`)

	// Write each case
	for _, c := range commands {
		buf.WriteString(fmt.Sprintf(`	case %s:
		return &DebuggerCommandResult{
			Id:      cmd.Id,
			Command: cmd.Command,
			%s: d.debugger.%s(%s),
		}, nil
`,
			c.CommandID,
			c.ResultFieldName,
			c.MethodName,
			c.ArgsFieldName,
		))
	}

	// Write default case
	buf.WriteString(`	default:
		return nil, fmt.Errorf("unsupported command: %s", cmd.Command)
	}
}
`)

	return buf.Bytes(), nil
}

// generateEnumConstants generates the DebuggerCommandId enum and String() method
func generateEnumConstants(commands []CommandInfo) string {
	var buf bytes.Buffer

	buf.WriteString(`// Code generated by debuggen; DO NOT EDIT.

package debugger

import (
	"fmt"
)

// Specifies the exact command sent to the debugger
type DebuggerCommandId int

const (
`)

	// Write each constant with iota
	for i, c := range commands {
		comments := strings.TrimSpace(c.Comments)
		if comments != "" {
			// Add comment lines
			for _, line := range strings.Split(comments, "\n") {
				buf.WriteString(fmt.Sprintf("\t// %s\n", strings.TrimPrefix(strings.TrimSpace(line), "// ")))
			}
		}

		if i == 0 {
			buf.WriteString(fmt.Sprintf("\t%s DebuggerCommandId = iota\n", c.CommandID))
		} else {
			buf.WriteString(fmt.Sprintf("\t%s\n", c.CommandID))
		}
	}

	buf.WriteString(`
)

func (c DebuggerCommandId) String() string {
	switch c {
`)

	// Write each case in String() method
	for _, cmd := range commands {
		buf.WriteString(fmt.Sprintf("\tcase %s:\n", cmd.CommandID))
		buf.WriteString(fmt.Sprintf("\t\treturn %q\n", cmd.CommandName))
	}

	buf.WriteString(`	default:
		return "unknown"
	}
}

func DebuggerCommandIdFromString(s string) (DebuggerCommandId, error) {
	switch s {
`)

	// Write each case in FromString() method
	for _, cmd := range commands {
		buf.WriteString(fmt.Sprintf("\tcase %q:\n", cmd.CommandName))
		buf.WriteString(fmt.Sprintf("\t\treturn %s, nil\n", cmd.CommandID))
	}

	buf.WriteString(`	default:
		return 0, fmt.Errorf("unknown DebuggerCommandId: %q", s)
	}
}
`)

	return buf.String()
}

// generateCommandStructs generates the DebuggerCommand and DebuggerCommandResult structs
func generateCommandStructs(commands []CommandInfo) string {
	var buf bytes.Buffer

	buf.WriteString(`// Code generated by debuggen; DO NOT EDIT.

package debugger

// DebuggerCommand represents a debugger command with its arguments
type DebuggerCommand struct {
	Id      uint64            ` + "`" + `json:"id"` + "`" + `      // Unique ID for this command instance
	Command DebuggerCommandId ` + "`" + `json:"command"` + "`" + ` // Type of command
`)

	// Add argument fields
	for _, cmd := range commands {
		if cmd.ArgsFieldName != "" {
			fieldName := cmd.MethodName + "Args"
			jsonName := strings.ToLower(fieldName)
			buf.WriteString("\t" + fieldName + " " + cmd.ArgsType + " `json:\"" + jsonName + "\"` // Command arguments for " + cmd.MethodName + " command\n")
		}
	}

	buf.WriteString(`}

// DebuggerCommandResult represents the result of a debugger command
type DebuggerCommandResult struct {
	Id      uint64            ` + "`" + `json:"id"` + "`" + `      // Unique ID for this command instance
	Command DebuggerCommandId ` + "`" + `json:"command"` + "`" + ` // Command identifier
`)

	// Add result fields
	for _, cmd := range commands {
		if cmd.ResultType != "" {
			fieldName := cmd.MethodName + "Result"
			jsonName := strings.ToLower(fieldName)
			buf.WriteString("\t" + fieldName + " " + cmd.ResultType + " `json:\"" + jsonName + "\"` // Result of " + cmd.MethodName + " command\n")
		}
	}

	buf.WriteString(`}
`)

	return buf.String()
}

// generateDocsSchema generates the documentation schema Go code
func generateDocsSchema(commands []CommandInfo) string {
	var buf bytes.Buffer

	buf.WriteString(`// Code generated by docs.go; DO NOT EDIT.

package debugger

// CommandsDocsSchema is the embedded documentation schema for debugger commands
var CommandsDocsSchema = &CommandsDocumentationSchema{
	Version: "1.0",
	Commands: map[DebuggerCommandId]CommandDocumentation{
`)

	for i, cmd := range commands {
		buf.WriteString(fmt.Sprintf("\t\t%d: {\n", i))
		buf.WriteString(fmt.Sprintf("\t\t\tID: %q,\n", fmt.Sprintf("%d", i)))
		buf.WriteString(fmt.Sprintf("\t\t\tName: %q,\n", cmd.CommandName))
		buf.WriteString(fmt.Sprintf("\t\t\tDescription: %q,\n", strings.TrimSpace(cmd.Comments)))
		buf.WriteString("\t\t\tArguments: []CommandArgumentInfo{},\n")
		buf.WriteString("\t\t\tResult: \"\",\n")
		buf.WriteString("\t\t\tResultFields: []CommandResultField{},\n")
		buf.WriteString("\t\t},\n")
	}

	buf.WriteString(`	},
	Enums: make(map[string]EnumDocumentation),
}
`)

	return buf.String()
}
