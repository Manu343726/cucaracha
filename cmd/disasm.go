package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Manu343726/cucaracha/pkg/debugger"
	debuggerUI "github.com/Manu343726/cucaracha/pkg/ui/debugger"
	"github.com/Manu343726/cucaracha/pkg/ui/disassembly"
	"github.com/Manu343726/cucaracha/pkg/utils/logging"
	"github.com/spf13/cobra"
)

var (
	disasmSystem     string
	disasmRuntime    string
	disasmShowSource bool
	disasmShowCFG    bool
	disasmShowDeps   bool
)

// DisasmCmd represents the disasm command
var DisasmCmd = &cobra.Command{
	Use:   "disasm [source]",
	Short: "Interactive disassembler for Cucaracha programs",
	Long: `Launch the interactive disassembly viewer for Cucaracha programs.

This command loads a C/C++ source file, compiles it to Cucaracha machine code,
and provides an interactive interface for exploring the compiled code with various
analysis features including control flow graphs, dependency analysis, and multi-criteria search.

Features:
  - Load and compile C/C++ programs
  - Interactive scrollable disassembly view
  - Instruction dependency graphs
  - Jump/branch graphs  
  - Search by mnemonic, operand, address, symbol, or source line
  - Vim-like command prompt for advanced navigation
  - Source code correlation

Usage:
  cucaracha disasm hello.c           # Compile and disassemble hello.c
  cucaracha disasm program.yaml      # Load program from descriptor

Examples:
  cucaracha disasm hello.c
  cucaracha disasm --show-source hello.c
  cucaracha disasm --show-cfg hello.c
  cucaracha disasm --show-deps hello.c`,
	Args: cobra.MaximumNArgs(1),
	Run:  runDisasm,
}

func init() {
	DisasmCmd.Flags().StringVarP(&disasmSystem, "system", "s", "", "System configuration YAML file")
	DisasmCmd.Flags().StringVarP(&disasmRuntime, "runtime", "r", "interpreter", "Runtime type: interpreter or llvm")
	DisasmCmd.Flags().BoolVar(&disasmShowSource, "show-source", false, "Show source code correlation")
	DisasmCmd.Flags().BoolVar(&disasmShowCFG, "show-cfg", true, "Show control flow graph information")
	DisasmCmd.Flags().BoolVar(&disasmShowDeps, "show-deps", true, "Show instruction dependency graphs")
}

func runDisasm(cmd *cobra.Command, args []string) {
	// Parse command line arguments
	loadArgs, err := parseDisasmArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Configure logging: redirect all components to stderr for visibility during setup
	reg := logging.DefaultRegistry()
	stderrSink := logging.NewTextSink("stderr", os.Stderr, slog.LevelDebug)
	if err := reg.RegisterSink(stderrSink); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to register stderr sink: %v\n", err)
	}

	// Replace all registered loggers to use stderr during setup
	rootLogger := logging.NewRegisteredLogger("cucaracha", stderrSink)
	reg.ReplaceLogger(rootLogger)

	// Create the underlying debugger
	dbg := debugger.NewDebugger()

	// Create the interactive disassembly session
	session := disassembly.NewSession(dbg, loadArgs)

	// Configure display options
	session.SetShowSource(disasmShowSource)
	session.SetShowCFG(disasmShowCFG)
	session.SetShowDeps(disasmShowDeps)

	// Run the interactive disassembly interface
	if err := session.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// parseDisasmArgs parses command line arguments into LoadArgs
func parseDisasmArgs(args []string) (*debuggerUI.LoadArgs, error) {
	loadArgs := &debuggerUI.LoadArgs{}

	// Extract program path if provided
	if len(args) > 0 {
		ext := filepath.Ext(args[0])
		if ext == ".c" || ext == ".cpp" || ext == ".cc" || ext == ".cxx" {
			loadArgs.ProgramPath = &args[0]
		} else if ext == ".yaml" || ext == ".yml" {
			loadArgs.FullDescriptorPath = &args[0]
		} else {
			return nil, fmt.Errorf("unrecognized file extension '%s' for argument '%s'\nSupported extensions: .c, .cpp, .cc, .cxx, .yaml, .yml", ext, args[0])
		}
	}

	// Set system config path if provided
	if disasmSystem != "" {
		loadArgs.SystemConfigPath = &disasmSystem
	}

	// Parse runtime type
	runtimeType, err := debuggerUI.RuntimeTypeFromString(disasmRuntime)
	if err != nil {
		return nil, fmt.Errorf("invalid runtime type: %v", err)
	}
	loadArgs.Runtime = &runtimeType

	return loadArgs, nil
}
