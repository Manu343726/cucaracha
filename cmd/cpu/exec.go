package cpu

import (
	"fmt"
	"os"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/interpreter"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/loader"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Color definitions for exec output
var (
	execColorHeader  = color.New(color.FgWhite, color.Bold)
	execColorSuccess = color.New(color.FgGreen, color.Bold)
	execColorWarning = color.New(color.FgYellow)
	execColorError   = color.New(color.FgRed, color.Bold)
	execColorAddr    = color.New(color.FgCyan)
	execColorValue   = color.New(color.FgWhite, color.Bold)
	execColorReg     = color.New(color.FgGreen)
	execColorLabel   = color.New(color.FgHiBlack)
	execColorFile    = color.New(color.FgHiBlue)
)

var (
	execMemorySize    uint32
	execVerbose       bool
	execMaxSteps      int
	execTrace         bool
	execCompileFormat string
	execTargetSpeed   float64
	execBuildClang    bool
)

var execCmd = &cobra.Command{
	Use:   "exec <file>",
	Short: "Execute a cucaracha program",
	Long: `Loads and executes a cucaracha program file.

The command accepts:
  - Assembly files (.cucaracha, .s) - parsed by the LLVM assembly parser
  - Binary/object files (.o) - parsed by the ELF binary parser
  - C/C++ source files (.c, .cpp, etc.) - compiled first, then executed

When a source file is provided, it is automatically compiled using clang
with the Cucaracha target before execution.

If --build-clang is specified and llvm-project is found but clang hasn't
been built yet, it will automatically build clang from source.

Example:
  cucaracha cpu exec program.cucaracha
  cucaracha cpu exec program.o
  cucaracha cpu exec program.c
  cucaracha cpu exec --build-clang program.c`,
	Args: cobra.ExactArgs(1),
	Run:  runExec,
}

func init() {
	CpuCmd.AddCommand(execCmd)
	execCmd.Flags().Uint32VarP(&execMemorySize, "memory", "m", 0x20000, "Memory size in bytes (default: 128KB to accommodate code at 0x10000)")
	execCmd.Flags().BoolVarP(&execVerbose, "verbose", "v", false, "Print execution details")
	execCmd.Flags().IntVarP(&execMaxSteps, "max-steps", "n", 0, "Maximum number of steps to execute (0 = unlimited)")
	execCmd.Flags().BoolVarP(&execTrace, "trace", "t", false, "Trace each instruction execution")
	execCmd.Flags().StringVar(&execCompileFormat, "compile-to", "object", "Compilation output format for source files: assembly, object")
	execCmd.Flags().Float64VarP(&execTargetSpeed, "speed", "s", 0, "Target execution speed in Hz (cycles per second). 0 = unlimited")
	execCmd.Flags().BoolVar(&execBuildClang, "build-clang", false, "Build clang from llvm-project if not found")
}

func runExec(cmd *cobra.Command, args []string) {
	inputPath := args[0]

	// Load the program using the loader package
	loadOpts := &loader.Options{
		Verbose:        execVerbose,
		OutputFormat:   execCompileFormat,
		AutoBuildClang: execBuildClang,
	}
	loadResult, err := loader.LoadFile(inputPath, loadOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading program: %v\n", err)
		os.Exit(1)
	}
	defer loadResult.Cleanup()

	// Display any warnings from loading
	for _, warning := range loadResult.Warnings {
		execColorWarning.Fprintf(os.Stderr, "WARNING: %s\n", warning)
	}

	resolved := loadResult.Program

	if execVerbose {
		fmt.Fprintf(os.Stderr, "Loaded %s instructions from %s\n",
			execColorValue.Sprintf("%d", len(resolved.Instructions())),
			execColorFile.Sprint(loadResult.CompiledPath))
		layout := resolved.MemoryLayout()
		if layout != nil {
			fmt.Fprintf(os.Stderr, "Memory layout: base=%s, code=%s bytes, data=%s bytes\n",
				execColorAddr.Sprintf("0x%08X", layout.BaseAddress),
				execColorValue.Sprintf("%d", layout.CodeSize),
				execColorValue.Sprintf("%d", layout.DataSize))
		}
	}

	// Create a runner and load the program
	runner := interpreter.NewRunner(execMemorySize)
	if err := runner.LoadProgram(resolved); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading program into interpreter: %v\n", err)
		os.Exit(2)
	}

	if execVerbose {
		fmt.Fprintf(os.Stderr, "Entry point: %s at %s\n",
			execColorReg.Sprint("main"),
			execColorAddr.Sprintf("0x%08X", runner.PC()))
		fmt.Fprintf(os.Stderr, "Termination address (LR): %s\n",
			execColorAddr.Sprintf("0x%08X", interpreter.TerminationAddress))
		fmt.Fprintf(os.Stderr, "Starting execution at PC=%s\n",
			execColorAddr.Sprintf("0x%08X", runner.PC()))
	}

	// Set target execution speed if specified
	if execTargetSpeed > 0 {
		runner.SetTargetSpeed(execTargetSpeed)
		if execVerbose {
			fmt.Fprintf(os.Stderr, "Target execution speed: %s Hz\n",
				execColorValue.Sprintf("%.2f", execTargetSpeed))
		}
	}

	// Set up lagging warning handler
	laggingWarningShown := false
	runner.Debugger().SetEventCallback(func(event interpreter.ExecutionEvent, result *interpreter.ExecutionResult) bool {
		if event == interpreter.EventLagging && !laggingWarningShown {
			execColorWarning.Fprintf(os.Stderr, "WARNING: Emulator running slower than target speed (%.2f Hz). Lagging by %d cycles.\n",
				execTargetSpeed, result.LagCycles)
			laggingWarningShown = true
		}
		return true // Continue execution
	})

	// Execute the program
	var result *interpreter.ExecutionResult
	if execTrace {
		result = executeWithTrace(runner, execMaxSteps)
	} else {
		result = runner.Run(execMaxSteps)
	}

	// Check for normal exit
	normalExit := runner.IsNormalExit()

	// Print final state
	state := runner.State()
	if execVerbose {
		if normalExit {
			execColorSuccess.Fprintf(os.Stderr, "\n=== Execution completed (returned from main) ===\n")
		} else {
			execColorHeader.Fprintf(os.Stderr, "\n=== Execution %s ===\n", result.StopReason.String())
		}
		fmt.Fprintf(os.Stderr, "%s %s\n",
			execColorLabel.Sprint("Steps executed:"),
			execColorValue.Sprintf("%d", result.StepsExecuted))
		fmt.Fprintf(os.Stderr, "%s %s\n",
			execColorLabel.Sprint("Cycles executed:"),
			execColorValue.Sprintf("%d", result.CyclesExecuted))
		if result.Lagging {
			execColorWarning.Fprintf(os.Stderr, "Lagging: yes (%d cycles behind)\n", result.LagCycles)
		}
		fmt.Fprintf(os.Stderr, "%s %s\n",
			execColorLabel.Sprint("Final PC:"),
			execColorAddr.Sprintf("0x%08X", state.PC))
		execColorHeader.Fprintln(os.Stderr, "Registers:")
		for i := 0; i < 10; i++ {
			fmt.Fprintf(os.Stderr, "  %s = %s (%s)\n",
				execColorReg.Sprintf("r%d", i),
				execColorValue.Sprintf("%d", state.Registers[16+i]),
				execColorAddr.Sprintf("0x%08X", state.Registers[16+i]))
		}
		fmt.Fprintf(os.Stderr, "  %s = %s (%s)\n",
			execColorReg.Sprint("sp"),
			execColorValue.Sprintf("%d", *state.SP),
			execColorAddr.Sprintf("0x%08X", *state.SP))
		fmt.Fprintf(os.Stderr, "  %s = %s (%s)\n",
			execColorReg.Sprint("lr"),
			execColorValue.Sprintf("%d", *state.LR),
			execColorAddr.Sprintf("0x%08X", *state.LR))
	}

	// Return value is in r0 (register index 16)
	returnValue := runner.ReturnValue()
	if execVerbose {
		fmt.Fprintf(os.Stderr, "\nReturn value (%s): %s\n",
			execColorReg.Sprint("r0"),
			execColorSuccess.Sprintf("%d", returnValue))
	} else {
		fmt.Printf("%d\n", returnValue)
	}

	// Exit with appropriate code
	if normalExit {
		os.Exit(0)
	}

	if result.Error != nil && !state.Halted {
		execColorError.Fprintf(os.Stderr, "Execution error: %v\n", result.Error)
		os.Exit(5)
	}
}

// executeWithTrace runs the program with instruction tracing
func executeWithTrace(runner *interpreter.Runner, maxSteps int) *interpreter.ExecutionResult {
	debugInfo := runner.DebugInfo()
	state := runner.State()

	// Set up trace formatting
	traceFormatter := interpreter.NewTraceFormatter(interpreter.OutputConfig{
		Style:  interpreter.StyleColored,
		Writer: os.Stderr,
	})

	var lastSourceLoc *mc.SourceLocation
	laggingWarningShown := false

	return runner.RunWithTraceAndEvents(maxSteps,
		func(step int, pc uint32, instrText string, srcLoc *mc.SourceLocation) bool {
			// Show source location if changed
			if srcLoc != nil && srcLoc.IsValid() {
				showSource := lastSourceLoc == nil ||
					lastSourceLoc.File != srcLoc.File ||
					lastSourceLoc.Line != srcLoc.Line
				if showSource {
					lastSourceLoc = srcLoc
					srcLine := ""
					if debugInfo != nil {
						srcLine = debugInfo.GetSourceLine(srcLoc.File, srcLoc.Line)
					}
					line := traceFormatter.FormatSourceLocation(srcLoc, srcLine)
					if line != "" {
						fmt.Fprintln(os.Stderr, line)
					}
				}
			}

			// Print trace line
			fmt.Fprintln(os.Stderr, traceFormatter.FormatStep(step, pc, instrText, state))
			return true // Continue execution
		},
		func(event interpreter.ExecutionEvent, result *interpreter.ExecutionResult) bool {
			if event == interpreter.EventLagging && !laggingWarningShown {
				execColorWarning.Fprintf(os.Stderr, "WARNING: Emulator running slower than target speed (%.2f Hz). Lagging by %d cycles.\n",
					execTargetSpeed, result.LagCycles)
				laggingWarningShown = true
			}
			return true
		})
}
