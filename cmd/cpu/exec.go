package cpu

import (
	"fmt"
	"os"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/interpreter"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/loader"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
	"github.com/spf13/cobra"
)

var (
	execMemorySize    uint32
	execVerbose       bool
	execMaxSteps      int
	execTrace         bool
	execCompileFormat string
	execTargetSpeed   float64
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

Example:
  cucaracha cpu exec program.cucaracha
  cucaracha cpu exec program.o
  cucaracha cpu exec program.c
  cucaracha cpu exec --compile-to assembly program.c`,
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
}

func runExec(cmd *cobra.Command, args []string) {
	inputPath := args[0]

	// Load the program using the loader package
	loadOpts := &loader.Options{
		Verbose:      execVerbose,
		OutputFormat: execCompileFormat,
	}
	loadResult, err := loader.LoadFile(inputPath, loadOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading program: %v\n", err)
		os.Exit(1)
	}
	defer loadResult.Cleanup()

	resolved := loadResult.Program

	if execVerbose {
		fmt.Fprintf(os.Stderr, "Loaded %d instructions from %s\n", len(resolved.Instructions()), loadResult.CompiledPath)
		layout := resolved.MemoryLayout()
		if layout != nil {
			fmt.Fprintf(os.Stderr, "Memory layout: base=0x%08X, code=%d bytes, data=%d bytes\n",
				layout.BaseAddress, layout.CodeSize, layout.DataSize)
		}
	}

	// Create a runner and load the program
	runner := interpreter.NewRunner(execMemorySize)
	if err := runner.LoadProgram(resolved); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading program into interpreter: %v\n", err)
		os.Exit(2)
	}

	if execVerbose {
		fmt.Fprintf(os.Stderr, "Entry point: main at 0x%08X\n", runner.PC())
		fmt.Fprintf(os.Stderr, "Termination address (LR): 0x%08X\n", interpreter.TerminationAddress)
		fmt.Fprintf(os.Stderr, "Starting execution at PC=0x%08X\n", runner.PC())
	}

	// Set target execution speed if specified
	if execTargetSpeed > 0 {
		runner.SetTargetSpeed(execTargetSpeed)
		if execVerbose {
			fmt.Fprintf(os.Stderr, "Target execution speed: %.2f Hz\n", execTargetSpeed)
		}
	}

	// Set up lagging warning handler
	laggingWarningShown := false
	runner.Debugger().SetEventCallback(func(event interpreter.ExecutionEvent, result *interpreter.ExecutionResult) bool {
		if event == interpreter.EventLagging && !laggingWarningShown {
			fmt.Fprintf(os.Stderr, "WARNING: Emulator running slower than target speed (%.2f Hz). Lagging by %d cycles.\n",
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
			fmt.Fprintf(os.Stderr, "\n=== Execution completed (returned from main) ===\n")
		} else {
			fmt.Fprintf(os.Stderr, "\n=== Execution %s ===\n", result.StopReason.String())
		}
		fmt.Fprintf(os.Stderr, "Steps executed: %d\n", result.StepsExecuted)
		fmt.Fprintf(os.Stderr, "Cycles executed: %d\n", result.CyclesExecuted)
		if result.Lagging {
			fmt.Fprintf(os.Stderr, "Lagging: yes (%d cycles behind)\n", result.LagCycles)
		}
		fmt.Fprintf(os.Stderr, "Final PC: 0x%08X\n", state.PC)
		fmt.Fprintf(os.Stderr, "Registers:\n")
		for i := 0; i < 10; i++ {
			fmt.Fprintf(os.Stderr, "  r%d = %d (0x%08X)\n", i, state.Registers[16+i], state.Registers[16+i])
		}
		fmt.Fprintf(os.Stderr, "  sp = %d (0x%08X)\n", *state.SP, *state.SP)
		fmt.Fprintf(os.Stderr, "  lr = %d (0x%08X)\n", *state.LR, *state.LR)
	}

	// Return value is in r0 (register index 16)
	returnValue := runner.ReturnValue()
	if execVerbose {
		fmt.Fprintf(os.Stderr, "\nReturn value (r0): %d\n", returnValue)
	} else {
		fmt.Printf("%d\n", returnValue)
	}

	// Exit with appropriate code
	if normalExit {
		os.Exit(0)
	}

	if result.Error != nil && !state.Halted {
		fmt.Fprintf(os.Stderr, "Execution error: %v\n", result.Error)
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
				fmt.Fprintf(os.Stderr, "WARNING: Emulator running slower than target speed (%.2f Hz). Lagging by %d cycles.\n",
					execTargetSpeed, result.LagCycles)
				laggingWarningShown = true
			}
			return true
		})
}
