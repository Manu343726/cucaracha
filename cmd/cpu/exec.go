package cpu

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/interpreter"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/llvm"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
	"github.com/spf13/cobra"
)

var (
	execMemorySize uint32
	execVerbose    bool
	execMaxSteps   int
	execTrace      bool
)

var execCmd = &cobra.Command{
	Use:   "exec <file>",
	Short: "Execute a cucaracha program",
	Long: `Loads and executes a cucaracha program file.

The command accepts either:
  - Assembly files (.cucaracha) - parsed by the LLVM assembly parser
  - Binary/object files (.o) - parsed by the ELF binary parser

The program is loaded, resolved (symbols, memory layout, instructions),
and then executed through the cucaracha interpreter.

Example:
  cucaracha cpu exec program.cucaracha
  cucaracha cpu exec program.o`,
	Args: cobra.ExactArgs(1),
	Run:  runExec,
}

func init() {
	CpuCmd.AddCommand(execCmd)
	execCmd.Flags().Uint32VarP(&execMemorySize, "memory", "m", 0x20000, "Memory size in bytes (default: 128KB to accommodate code at 0x10000)")
	execCmd.Flags().BoolVarP(&execVerbose, "verbose", "v", false, "Print execution details")
	execCmd.Flags().IntVarP(&execMaxSteps, "max-steps", "n", 0, "Maximum number of steps to execute (0 = unlimited)")
	execCmd.Flags().BoolVarP(&execTrace, "trace", "t", false, "Trace each instruction execution")
}

func runExec(cmd *cobra.Command, args []string) {
	inputPath := args[0]

	// Determine file type by extension
	ext := strings.ToLower(filepath.Ext(inputPath))

	var pf mc.ProgramFile
	var err error

	switch ext {
	case ".cucaracha", ".s":
		if execVerbose {
			fmt.Fprintf(os.Stderr, "Loading assembly file: %s\n", inputPath)
		}
		pf, err = llvm.ParseAssemblyFile(inputPath)
	case ".o":
		if execVerbose {
			fmt.Fprintf(os.Stderr, "Loading binary file: %s\n", inputPath)
		}
		pf, err = llvm.ParseBinaryFile(inputPath)
	default:
		fmt.Fprintf(os.Stderr, "Error: unsupported file extension '%s'\n", ext)
		fmt.Fprintln(os.Stderr, "Supported extensions: .cucaracha, .s (assembly), .o (binary)")
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading program: %v\n", err)
		os.Exit(2)
	}

	if execVerbose {
		fmt.Fprintf(os.Stderr, "Loaded %d instructions\n", len(pf.Instructions()))
	}

	// Resolve the program
	// Use a higher base address to avoid conflict with LLVM's absolute local variable addresses
	memConfig := mc.MemoryResolverConfig{
		BaseAddress:     0x10000, // 64KB offset to avoid overlap with low absolute addresses
		MaxSize:         0,
		DataAlignment:   4,
		InstructionSize: 4,
	}
	resolved, err := mc.Resolve(pf, memConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving program: %v\n", err)
		os.Exit(3)
	}

	if execVerbose {
		layout := resolved.MemoryLayout()
		if layout != nil {
			fmt.Fprintf(os.Stderr, "Memory layout: base=0x%08X, code=%d bytes, data=%d bytes\n",
				layout.BaseAddress, layout.CodeSize, layout.DataSize)
		}
	}

	// Create interpreter and debugger
	interp := interpreter.NewInterpreter(execMemorySize)
	dbg := interpreter.NewDebugger(interp)

	if err := loadProgramFile(interp, resolved); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading program into interpreter: %v\n", err)
		os.Exit(4)
	}

	// Define a magic termination address (just before the base address)
	// When main returns (JMP lr, lr), PC will jump here and we'll detect program end
	const TerminationAddress uint32 = 0x0000FFFC
	dbg.AddTerminationAddress(TerminationAddress)

	// Find the main function entry point
	mainFunc, hasMain := resolved.Functions()["main"]
	if hasMain && len(mainFunc.InstructionRanges) > 0 {
		startIdx := mainFunc.InstructionRanges[0].Start
		instrs := resolved.Instructions()
		if startIdx < len(instrs) && instrs[startIdx].Address != nil {
			interp.State().PC = *instrs[startIdx].Address
			// Set LR to termination address so "JMP lr, lr" at end of main will terminate
			*interp.State().LR = TerminationAddress
			if execVerbose {
				fmt.Fprintf(os.Stderr, "Entry point: main at 0x%08X\n", interp.State().PC)
				fmt.Fprintf(os.Stderr, "Termination address (LR): 0x%08X\n", TerminationAddress)
			}
		}
	}

	if execVerbose {
		fmt.Fprintf(os.Stderr, "Starting execution at PC=0x%08X\n", interp.State().PC)
	}

	// Execute the program using the debugger API
	var result *interpreter.ExecutionResult
	if execTrace {
		result = executeWithTraceDebugger(dbg, execMaxSteps, resolved)
	} else {
		result = dbg.Run(execMaxSteps)
	}

	// Check if execution stopped due to jumping past end of code (normal return from main)
	state := interp.State()
	layout := resolved.MemoryLayout()
	endOfCode := layout.CodeStart + layout.CodeSize
	normalExit := result.StopReason == interpreter.StopTermination ||
		(result.Error != nil && state.PC >= endOfCode)

	// Print final state
	if execVerbose {
		if normalExit {
			fmt.Fprintf(os.Stderr, "\n=== Execution completed (returned from main) ===\n")
		} else {
			fmt.Fprintf(os.Stderr, "\n=== Execution %s ===\n", result.StopReason.String())
		}
		fmt.Fprintf(os.Stderr, "Steps executed: %d\n", result.StepsExecuted)
		fmt.Fprintf(os.Stderr, "Final PC: 0x%08X\n", state.PC)
		fmt.Fprintf(os.Stderr, "Registers:\n")
		for i := 0; i < 10; i++ {
			fmt.Fprintf(os.Stderr, "  r%d = %d (0x%08X)\n", i, state.Registers[16+i], state.Registers[16+i])
		}
		fmt.Fprintf(os.Stderr, "  sp = %d (0x%08X)\n", *state.SP, *state.SP)
		fmt.Fprintf(os.Stderr, "  lr = %d (0x%08X)\n", *state.LR, *state.LR)
	}

	// Return value is in r0 (register index 16)
	r0 := state.Registers[16]
	if execVerbose {
		fmt.Fprintf(os.Stderr, "\nReturn value (r0): %d\n", r0)
	} else {
		fmt.Printf("%d\n", r0)
	}

	// If execution stopped due to returning past end of code, that's normal
	if normalExit {
		os.Exit(0)
	}

	if result.Error != nil && !state.Halted {
		fmt.Fprintf(os.Stderr, "Execution error: %v\n", result.Error)
		os.Exit(5)
	}
}

// loadProgramFile loads a resolved ProgramFile into the interpreter
func loadProgramFile(interp *interpreter.Interpreter, pf mc.ProgramFile) error {
	layout := pf.MemoryLayout()
	if layout == nil {
		return fmt.Errorf("program has no memory layout (not resolved)")
	}

	instrList := pf.Instructions()
	if len(instrList) == 0 {
		return fmt.Errorf("program has no instructions")
	}

	// Load each instruction's binary encoding into memory
	for i, instr := range instrList {
		if instr.Address == nil {
			return fmt.Errorf("instruction %d has no address (not resolved)", i)
		}
		if instr.Instruction == nil {
			return fmt.Errorf("instruction %d has no decoded instruction", i)
		}

		// Encode the instruction to binary via RawInstruction
		rawInstr := instr.Instruction.Raw()
		encoded := rawInstr.Encode()

		// Debug: show first 15 instructions being loaded
		if execTrace && i < 15 {
			fmt.Fprintf(os.Stderr, "Loading [%3d] addr=0x%08X enc=0x%08X %s\n", i, *instr.Address, encoded, instr.Text)
		}

		// Write to memory
		if err := interp.State().WriteMemory32(*instr.Address, encoded); err != nil {
			return fmt.Errorf("failed to write instruction %d at 0x%08X: %w", i, *instr.Address, err)
		}
	}

	// Debug: verify instruction at 0x0004 after loading
	if execTrace {
		word, _ := interp.State().ReadMemory32(0x0004)
		fmt.Fprintf(os.Stderr, "After loading code: memory[0x0004] = 0x%08X\n", word)
	}

	// Load global data
	for _, global := range pf.Globals() {
		if global.Address == nil {
			continue // Skip unresolved globals
		}
		if execTrace {
			fmt.Fprintf(os.Stderr, "Loading global %q at 0x%08X, %d bytes\n", global.Name, *global.Address, len(global.InitialData))
		}
		if len(global.InitialData) > 0 {
			addr := *global.Address
			for j, b := range global.InitialData {
				if int(addr)+j >= len(interp.State().Memory) {
					return fmt.Errorf("global '%s' data exceeds memory bounds", global.Name)
				}
				interp.State().Memory[addr+uint32(j)] = b
			}
		}
	}

	// Debug: verify instruction at 0x0004 after globals
	if execTrace {
		word, _ := interp.State().ReadMemory32(0x0004)
		fmt.Fprintf(os.Stderr, "After loading globals: memory[0x0004] = 0x%08X\n", word)
	}

	// Set initial PC to the start of code
	interp.State().PC = layout.CodeStart

	return nil
}

// executeWithTraceDebugger runs the debugger with instruction tracing via event callback
func executeWithTraceDebugger(dbg *interpreter.Debugger, maxSteps int, pf mc.ProgramFile) *interpreter.ExecutionResult {
	state := dbg.State()
	instrs := pf.Instructions()

	// Build address to instruction index map
	addrToIdx := make(map[uint32]int)
	for i, instr := range instrs {
		if instr.Address != nil {
			addrToIdx[*instr.Address] = i
		}
	}

	stepCount := 0

	// Set up tracing callback
	dbg.SetEventCallback(func(event interpreter.ExecutionEvent, result *interpreter.ExecutionResult) bool {
		if event == interpreter.EventStep {
			// Print current state before execution
			pc := result.LastPC
			idx, ok := addrToIdx[pc]
			instrText := "???"
			if ok && idx < len(instrs) {
				instrText = instrs[idx].Text
			}

			// Read the instruction word and show it
			word, _ := state.ReadMemory32(pc)

			fmt.Fprintf(os.Stderr, "[%4d] PC=0x%04X word=0x%08X sp=%6d (R[1]=%d) lr=%6d r0=%6d r4=%6d r5=%6d | %s\n",
				stepCount, pc, word, *state.SP, state.Registers[1], *state.LR,
				state.Registers[16], state.Registers[20], state.Registers[21],
				instrText)
			stepCount++
		}
		return true // Continue execution
	})

	return dbg.Run(maxSteps)
}
