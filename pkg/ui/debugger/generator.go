//go:build ignore
// +build ignore

// Generator for debugger command infrastructure.
// This is a code generation tool invoked via go:generate.
// Run with: go run generator.go -out <output> -api <api.go> ...
//
// This file is compiled separately and excluded from normal package builds
// due to the //go:build ignore directive above.
//
// It imports and reuses the implementation functions from generator_impl.go
// which is part of the debugger package.

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Manu343726/cucaracha/pkg/ui/debugger"
)

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
	commands, err := debugger.ParseDebuggerInterface(*apiPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing API file: %v\n", err)
		os.Exit(1)
	}

	// Generate code for the Execute() method
	output, err := debugger.GenerateExecuteMethod(commands)
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
		enumCode := debugger.GenerateEnumConstants(commands)
		if err := os.WriteFile(*enumOutput, []byte(enumCode), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing enum output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated enum constants in %s\n", *enumOutput)
	}

	// Generate structs if requested
	if *structsOutput != "" {
		structsCode := debugger.GenerateCommandStructs(commands)
		if err := os.WriteFile(*structsOutput, []byte(structsCode), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing structs output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated command structs in %s\n", *structsOutput)
	}

	// Generate documentation schema if requested
	if *docsOutput != "" {
		docsCode := debugger.GenerateDocsSchema(commands)
		if err := os.WriteFile(*docsOutput, []byte(docsCode), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing docs output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated documentation schema in %s\n", *docsOutput)
	}
}
