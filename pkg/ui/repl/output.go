package repl

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	debuggerUI "github.com/Manu343726/cucaracha/pkg/ui/debugger"
	"github.com/Manu343726/cucaracha/pkg/utils/logging"
)

func (r *REPL) printCommandResult(result *debuggerUI.DebuggerCommandResult) {
	if result == nil {
		r.printError("No result returned")
		return
	}

	// Print the appropriate result based on command type
	if result.StepResult != nil {
		r.printExecutionResult(result.StepResult)
	}
	if result.ContinueResult != nil {
		r.printExecutionResult(result.ContinueResult)
	}
	if result.InterruptResult != nil {
		r.printExecutionResult(result.InterruptResult)
	}
	if result.ResetResult != nil {
		r.printExecutionResult(result.ResetResult)
	}
	if result.RestartResult != nil {
		r.printExecutionResult(result.RestartResult)
	}
	if result.BreakResult != nil {
		r.printBreakResult(result.BreakResult)
	}
	if result.WatchResult != nil {
		r.printWatchResult(result.WatchResult)
	}
	if result.RemoveBreakpointResult != nil {
		r.printRemoveBreakpointResult(result.RemoveBreakpointResult)
	}
	if result.RemoveWatchpointResult != nil {
		r.printRemoveWatchpointResult(result.RemoveWatchpointResult)
	}
	if result.ListResult != nil {
		r.printListResult(result.ListResult)
	}
	if result.DisassemblyResult != nil {
		r.printDisassemblyResult(result.DisassemblyResult)
	}
	if result.CurrentInstructionResult != nil {
		r.printCurrentInstructionResult(result.CurrentInstructionResult)
	}
	if result.MemoryResult != nil {
		r.printMemoryResult(result.MemoryResult)
	}
	if result.SourceResult != nil {
		r.printSourceResult(result.SourceResult)
	}
	if result.CurrentSourceResult != nil {
		r.printSourceResult(result.CurrentSourceResult)
	}
	if result.InfoResult != nil {
		r.printInfoResult(result.InfoResult)
	}
	if result.RegistersResult != nil {
		r.printRegistersResult(result.RegistersResult)
	}
	if result.StackResult != nil {
		r.printStackResult(result.StackResult)
	}
	if result.VariablesResult != nil {
		r.printVarsResult(result.VariablesResult)
	}
	if result.SymbolsResult != nil {
		r.printSymbolsResult(result.SymbolsResult)
	}
	if result.EvalResult != nil {
		r.printEvalResult(result.EvalResult)
	}
	if result.LoadProgramResult != nil {
		r.printLoadProgramResult(result.LoadProgramResult)
	}
	if result.LoadSystemResult != nil {
		r.printLoadSystemResult(result.LoadSystemResult)
	}
	if result.LoadRuntimeResult != nil {
		r.printLoadRuntimeResult(result.LoadRuntimeResult)
	}
	if result.LoadResult != nil {
		r.printLoadResult(result.LoadResult)
	}
}

func (r *REPL) printWelcome() {
	r.write("\n")
	r.write("Welcome to Cucaracha Debugger REPL\n")
	r.write("Type 'help' for available commands\n")
	r.write("\n")
}

func (r *REPL) printGoodbye() {
	r.write("\nGoodbye!\n")
}

func (r *REPL) printHelp() {
	help := `
Available Commands:

Execution:
  step, s                - Step through one instruction
  continue, c            - Continue execution until breakpoint
  run, r                 - Run the program
  interrupt              - Interrupt execution
  reset                  - Reset program to initial state
  restart                - Reset and continue execution

Breakpoints:
  break <addr>, b <addr> - Set breakpoint at address
  rbp <id>               - Remove breakpoint by ID
  watch <addr>, w <addr> - Set watchpoint at address
  rw <id>                - Remove watchpoint by ID
  list, l                - List all breakpoints and watchpoints

Inspection:
  disasm [addr] [cnt]    - Disassemble instructions
  current                - Show current instruction
  memory <addr> [cnt]    - Display memory
  source [path]          - Show source code
  info [general|runtime| - Show debugger/system/program info
    program], i
  registers, reg         - Show CPU registers
  stack, st              - Show stack trace
  vars, v                - Show variables
  eval <expr>, e <expr>  - Evaluate expression
  symbols [name], sym    - Show loaded symbols

Program Loading:
  load <file>            - Load program from file
  loadprogram <file>     - Load program from file
  loadsystem <file>      - Load system configuration
  loadruntime <name>     - Load runtime (interpreter)

Settings:
  set [name] [value]     - Set a REPL setting (or show all with descriptions)
  get [name]             - Get a setting value (or show all current values)

Utility:
  loggers                - List all registered loggers and their sinks
  help, h                - Show this help message
  exit, quit, q          - Exit the debugger
`
	r.write("%s", help)
}

func (r *REPL) printError(msg string) {
	r.write("Error: %s\n", msg)
}

func (r *REPL) printExecutionResult(result *debuggerUI.ExecutionResult) {
	if result == nil {
		r.printError("No execution result")
		return
	}

	if result.Error != nil {
		r.printError(result.Error.Error())
		return
	}

	r.write("Stopped: %s\n", result.StopReason)
	r.write("Steps: %d, Cycles: %d\n", result.Steps, result.Cycles)

	if result.LastInstruction > 0 {
		r.write("Last Instruction: 0x%x\n", result.LastInstruction)
	}

	if result.LastLocation != nil {
		r.write("Location: %s:%d\n",
			result.LastLocation.File,
			result.LastLocation.Line)
	}

	if result.Breakpoint != nil {
		r.write("Hit Breakpoint: %d at 0x%x\n",
			result.Breakpoint.ID,
			result.Breakpoint.Address)
	}

	if result.Watchpoint != nil {
		r.write("Hit Watchpoint: %d at 0x%x\n",
			result.Watchpoint.ID,
			result.Watchpoint.Range.Start)
	}
}

func (r *REPL) printBreakResult(result *debuggerUI.BreakResult) {
	if result == nil {
		r.printError("Failed to set breakpoint")
		return
	}

	if result.Error != nil {
		r.printError(result.Error.Error())
		return
	}

	if result.Breakpoint != nil {
		r.write("Breakpoint %d set at 0x%x\n",
			result.Breakpoint.ID,
			result.Breakpoint.Address)
	}
}

func (r *REPL) printWatchResult(result *debuggerUI.WatchResult) {
	if result == nil {
		r.printError("Failed to set watchpoint")
		return
	}

	if result.Error != nil {
		r.printError(result.Error.Error())
		return
	}

	if result.Watchpoint != nil {
		r.write("Watchpoint %d set at 0x%x\n",
			result.Watchpoint.ID,
			result.Watchpoint.Range.Start)
	}
}

func (r *REPL) printRemoveBreakpointResult(result *debuggerUI.RemoveBreakpointResult) {
	if result == nil {
		r.printError("Failed to remove breakpoint")
		return
	}

	if result.Error != nil {
		r.printError(result.Error.Error())
		return
	}

	r.write("Breakpoint removed\n")
}

func (r *REPL) printRemoveWatchpointResult(result *debuggerUI.RemoveWatchpointResult) {
	if result == nil {
		r.printError("Failed to remove watchpoint")
		return
	}

	if result.Error != nil {
		r.printError(result.Error.Error())
		return
	}

	r.write("Watchpoint removed\n")
}

func (r *REPL) printListResult(result *debuggerUI.ListResult) {
	if result == nil {
		r.printError("Failed to list breakpoints")
		return
	}

	if result.Error != nil {
		r.printError(result.Error.Error())
		return
	}

	if len(result.Breakpoints) == 0 && len(result.Watchpoints) == 0 {
		r.write("No breakpoints or watchpoints\n")
		return
	}

	if len(result.Breakpoints) > 0 {
		r.write("\nBreakpoints:\n")
		for _, bp := range result.Breakpoints {
			r.write("  %d at 0x%x\n", bp.ID, bp.Address)
		}
	}

	if len(result.Watchpoints) > 0 {
		r.write("\nWatchpoints:\n")
		for _, wp := range result.Watchpoints {
			r.write("  %d at 0x%x (size: %d)\n", wp.ID, wp.Range.Start, wp.Range.Size)
		}
	}
}

func (r *REPL) printDisassemblyResult(result *debuggerUI.DisassemblyResult) {
	if result == nil {
		r.printError("Failed to disassemble")
		return
	}

	if result.Error != nil {
		r.printError(result.Error.Error())
		return
	}

	if len(result.Instructions) == 0 {
		r.write("No instructions\n")
		return
	}

	r.write("\nInstructions:\n")

	// Get display flags from last disasm args, default to showing everything
	showSource := true
	showCFG := true
	if r.lastDisasmArgs != nil {
		showSource = r.lastDisasmArgs.ShowSource
		showCFG = r.lastDisasmArgs.ShowCFG
	}

	// Build a set of branch targets from the CFG
	branchTargets := make(map[uint32]bool)
	if result.CFG != nil {
		for target := range result.CFG.Edges {
			branchTargets[target] = true
		}
	}

	// First pass: calculate column widths
	maxLocationWidth := 0
	maxSourceWidth := 0
	maxCFGWidth := 0

	if showSource {
		for _, inst := range result.Instructions {
			if inst.SourceLine != nil && inst.SourceLine.Location != nil {
				locationStr := fmt.Sprintf("%s:%d", inst.SourceLine.Location.File, inst.SourceLine.Location.Line)
				if len(locationStr) > maxLocationWidth {
					maxLocationWidth = len(locationStr)
				}
				if len(inst.SourceLine.Text) > maxSourceWidth {
					maxSourceWidth = len(inst.SourceLine.Text)
				}
			}
		}
	}

	if showCFG {
		for _, inst := range result.Instructions {
			// Calculate CFG column width
			cfgStr := r.getCFGSymbol(inst, result.CFG)
			if len(cfgStr) > maxCFGWidth {
				maxCFGWidth = len(cfgStr)
			}
		}
		// Ensure minimum width for CFG column
		if maxCFGWidth < 3 {
			maxCFGWidth = 3
		}
	}

	// Second pass: print with aligned columns
	var prevLocationStr string
	for _, inst := range result.Instructions {
		marker := " "
		if inst.IsCurrentPC {
			marker = ">"
		}

		// Get location and source information
		var locationStr, sourceStr string
		if inst.SourceLine != nil && inst.SourceLine.Location != nil {
			locationStr = fmt.Sprintf("%s:%d", inst.SourceLine.Location.File, inst.SourceLine.Location.Line)
			sourceStr = inst.SourceLine.Text
		}

		// Build the line with conditionally included columns
		var line string

		// Add location and code columns if enabled
		if showSource {
			// Add location column (only if different from previous)
			showLocation := locationStr != prevLocationStr && locationStr != ""
			if showLocation {
				line += locationStr
				prevLocationStr = locationStr
			}
			// Pad location column
			for len(line) < maxLocationWidth {
				line += " "
			}

			// Add spacing between columns
			line += "    "

			// Add source code column (only if we showed location)
			if showLocation {
				line += sourceStr
			}
			// Pad source column
			for len(line) < maxLocationWidth+4+maxSourceWidth {
				line += " "
			}

			// Add spacing between columns
			line += "    "
		}

		// Add CFG column if enabled
		if showCFG {
			cfgStr := r.getCFGSymbol(inst, result.CFG)
			line += cfgStr
			// Pad CFG column
			for len(line) < maxLocationWidth+4+maxSourceWidth+4+maxCFGWidth {
				line += " "
			}

			// Add spacing between columns
			line += "    "
		}

		// Add instruction
		instrPart := fmt.Sprintf("%s 0x%08x: %s %s", marker, inst.Address, inst.Mnemonic, inst.Text)
		line += instrPart

		fmt.Fprintf(r.writer, "%s\n", line)
	}
}

// getCFGSymbol returns a string representing the control flow graph information for an instruction
func (r *REPL) getCFGSymbol(inst *debuggerUI.Instruction, cfg *debuggerUI.ControlFlowGraph) string {
	if cfg == nil {
		return "○"
	}

	// Check if this instruction is a branch source (has an edge in the CFG)
	_, isBranchSource := cfg.Edges[inst.Address]

	// Check if this instruction is a branch target
	isBranchTarget := false
	for _, target := range cfg.Edges {
		if target == inst.Address {
			isBranchTarget = true
			break
		}
	}

	switch {
	case isBranchSource && isBranchTarget:
		// Both source and target of branches
		return "⟲↔"
	case isBranchSource:
		// Branch source - shows arrow pointing to target
		target := cfg.Edges[inst.Address]
		return fmt.Sprintf("→%X", target&0xFFFF) // Show last 4 hex digits
	case isBranchTarget:
		// Branch target - shows incoming arrow
		return "←"
	default:
		// Regular fall-through instruction
		return "○"
	}
}

func (r *REPL) printCurrentInstructionResult(result *debuggerUI.CurrentInstructionResult) {
	if result == nil {
		r.printError("Failed to get current instruction")
		return
	}

	if result.Error != nil {
		r.printError(result.Error.Error())
		return
	}

	if result.Instruction != nil {
		output := fmt.Sprintf("Current: 0x%08x: %s %s",
			result.Instruction.Address,
			result.Instruction.Mnemonic,
			result.Instruction.Text)

		// Append source location and source line information if available
		if result.Instruction.SourceLine != nil {
			if result.Instruction.SourceLine.Location != nil {
				output += fmt.Sprintf(" (%s:%d", result.Instruction.SourceLine.Location.File, result.Instruction.SourceLine.Location.Line)
				if result.Instruction.SourceLine.Text != "" {
					output += fmt.Sprintf(" - %s", result.Instruction.SourceLine.Text)
				}
				output += ")"
			}
		}

		// Use standard output pattern to avoid vet warnings
		fmt.Fprintf(r.writer, "%s\n", output)
	}
}

func (r *REPL) printMemoryResult(result *debuggerUI.MemoryResult) {
	if result == nil {
		r.printError("Failed to read memory")
		return
	}

	if result.Error != nil {
		r.printError(result.Error.Error())
		return
	}

	if len(result.Data) == 0 {
		r.write("No data\n")
		return
	}

	r.write("\nMemory at 0x%x:\n", result.Address)
	for i := 0; i < len(result.Data); i += 16 {
		end := i + 16
		if end > len(result.Data) {
			end = len(result.Data)
		}

		r.write("0x%08x: ", uint32(result.Address)+uint32(i))
		for j := i; j < end; j++ {
			r.write("%02x ", result.Data[j])
		}
		r.write("\n")
	}
}

func (r *REPL) printSourceResult(result *debuggerUI.SourceResult) {
	if result == nil {
		r.printError("Failed to read source")
		return
	}

	if result.Error != nil {
		r.printError(result.Error.Error())
		return
	}

	if result.Snippet == nil || len(result.Snippet.Lines) == 0 {
		r.write("No source\n")
		return
	}

	r.write("\nSource:\n")
	for _, line := range result.Snippet.Lines {
		marker := " "
		if line.IsCurrent {
			marker = ">"
		}
		if line.Location != nil {
			r.write("%s %4d: %s\n", marker, line.Location.Line, line.Text)
		} else {
			r.write("%s %s\n", marker, line.Text)
		}
	}
}

func (r *REPL) printInfoResult(result *debuggerUI.InfoResult) {
	if result == nil {
		r.printError("Failed to get info")
		return
	}

	if result.Error != nil {
		r.printError(result.Error.Error())
		return
	}

	// Print debugger state if available
	if result.DebuggerState != nil {
		r.write("\nDebugger State:\n")
		r.write("  Status: %v\n", result.DebuggerState.Status)
		if result.DebuggerState.Registers != nil && len(result.DebuggerState.Registers) > 0 {
			r.write("  Registers:\n")
			for name, reg := range result.DebuggerState.Registers {
				r.write("    %s = 0x%x\n", name, reg.Value)
			}
		}
		if result.DebuggerState.Flags != nil {
			r.write("  Flags: N=%v Z=%v C=%v V=%v\n",
				result.DebuggerState.Flags.N,
				result.DebuggerState.Flags.Z,
				result.DebuggerState.Flags.C,
				result.DebuggerState.Flags.V)
		}
	}

	// Print system info if available
	if result.SystemInfo != nil {
		r.write("\nSystem Configuration:\n")
		r.write("  Total Memory: %d bytes (0x%x)\n", result.SystemInfo.TotalMemory, result.SystemInfo.TotalMemory)
		r.write("  Code Region: %d bytes (0x%x)\n", result.SystemInfo.CodeSize, result.SystemInfo.CodeSize)
		r.write("  Data Region: %d bytes (0x%x)\n", result.SystemInfo.DataSize, result.SystemInfo.DataSize)
		r.write("  Stack Region: %d bytes (0x%x)\n", result.SystemInfo.StackSize, result.SystemInfo.StackSize)
		r.write("  Heap Region: %d bytes (0x%x)\n", result.SystemInfo.HeapSize, result.SystemInfo.HeapSize)
		r.write("  Peripheral Region: %d bytes (0x%x)\n", result.SystemInfo.PeripheralSize, result.SystemInfo.PeripheralSize)
		r.write("  Interrupt Vectors: %d (entry size: %d bytes)\n", result.SystemInfo.NumberOfVectors, result.SystemInfo.VectorEntrySize)

		if result.SystemInfo.NumPeripherals > 0 {
			r.write("  Peripherals (%d):\n", result.SystemInfo.NumPeripherals)
			for _, p := range result.SystemInfo.Peripherals {
				r.write("    - %s (%s): %s\n", p.Name, p.Type, p.DisplayName)
				r.write("      Base Address: 0x%x, Size: %d bytes, IRQ: %d\n",
					p.BaseAddress, p.Size, p.InterruptVector)
			}
		}
	}

	// Print program info if available
	if result.ProgramInfo != nil {
		r.write("\nProgram Information:\n")
		if result.ProgramInfo.SourceFile != nil {
			r.write("  Source File: %s\n", *result.ProgramInfo.SourceFile)
		}
		if result.ProgramInfo.ObjectFile != nil {
			r.write("  Object File: %s\n", *result.ProgramInfo.ObjectFile)
		}
		r.write("  Entry Point: 0x%x\n", result.ProgramInfo.EntryPoint)
		r.write("  Debug Symbols: %v\n", result.ProgramInfo.HasDebugInfo)
		if len(result.ProgramInfo.Warnings) > 0 {
			r.write("  Warnings:\n")
			for _, w := range result.ProgramInfo.Warnings {
				r.write("    - %s\n", w)
			}
		}
	}

	// Print runtime info if available
	if result.RuntimeInfo != nil {
		r.write("\nRuntime Information:\n")
		r.write("  Type: %s\n", result.RuntimeInfo.Runtime)
	}
}

func (r *REPL) printRegistersResult(result *debuggerUI.RegistersResult) {
	if result == nil {
		r.printError("Failed to get registers")
		return
	}

	if result.Error != nil {
		r.printError(result.Error.Error())
		return
	}

	if len(result.Registers) == 0 {
		r.write("No registers\n")
		return
	}

	r.write("\nRegisters:\n")
	for name, reg := range result.Registers {
		if reg != nil {
			r.write("  %s = 0x%x\n", name, reg.Value)
		}
	}
}

func (r *REPL) printStackResult(result *debuggerUI.StackResult) {
	if result == nil {
		r.printError("Failed to get stack")
		return
	}

	if result.Error != nil {
		r.printError(result.Error.Error())
		return
	}

	if len(result.StackFrames) == 0 {
		r.write("Empty stack\n")
		return
	}

	r.write("\nStack Frames:\n")
	for i, frame := range result.StackFrames {
		functionName := "unknown"
		if frame.Function != nil {
			functionName = *frame.Function
		}
		address := uint32(0)
		if frame.Memory != nil {
			address = frame.Memory.Start
		}
		r.write("  #%d %s at 0x%x\n", i, functionName, address)
	}
}

func (r *REPL) printVarsResult(result *debuggerUI.VarsResult) {
	if result == nil {
		r.printError("Failed to get variables")
		return
	}

	if result.Error != nil {
		r.printError(result.Error.Error())
		return
	}

	if len(result.Variables) == 0 {
		r.write("No variables\n")
		return
	}

	r.write("\nVariables:\n")
	for _, variable := range result.Variables {
		if variable != nil {
			r.write("  %s: %s = %s\n", variable.Name, variable.TypeName, variable.ValueString)
		}
	}
}

func (r *REPL) printSymbolsResult(result *debuggerUI.SymbolsResult) {
	if result == nil {
		r.printError("Failed to list symbols")
		return
	}

	if result.Error != nil {
		r.printError(result.Error.Error())
		return
	}

	if result.TotalCount == 0 {
		r.write("No symbols found\n")
		return
	}

	r.write("\nSymbols (%d total):\n", result.TotalCount)

	// Print functions
	if len(result.Functions) > 0 {
		r.write("\nFunctions:\n")
		for _, fn := range result.Functions {
			if fn != nil {
				addrStr := "???"
				if fn.Address != nil {
					addrStr = fmt.Sprintf("0x%x", *fn.Address)
				}
				sizeStr := "???"
				if fn.Size != nil {
					sizeStr = fmt.Sprintf("%d", *fn.Size)
				}
				rangeStr := ""
				if len(fn.InstructionRanges) > 0 {
					rangeStr = fmt.Sprintf(" [%s]", strings.Join(fn.InstructionRanges, ", "))
				}
				sourceStr := ""
				if fn.SourceFile != "" {
					sourceStr = fmt.Sprintf(" (%s:%d-%d)", fn.SourceFile, fn.StartLine, fn.EndLine)
				}
				r.write("  %s @ %s size=%s%s%s\n", fn.Name, addrStr, sizeStr, rangeStr, sourceStr)
			}
		}
	}

	// Print globals
	if len(result.Globals) > 0 {
		r.write("\nGlobals:\n")
		for _, global := range result.Globals {
			if global != nil {
				addrStr := "???"
				if global.Address != nil {
					addrStr = fmt.Sprintf("0x%x", *global.Address)
				}
				typeStr := global.SymbolType
				initStr := ""
				if global.HasInitData {
					initStr = fmt.Sprintf(" (init data: %d bytes)", global.InitDataLen)
				}
				r.write("  %s @ %s type=%s size=%d%s\n", global.Name, addrStr, typeStr, global.Size, initStr)
			}
		}
	}

	// Print labels
	if len(result.Labels) > 0 {
		r.write("\nLabels:\n")
		for _, label := range result.Labels {
			if label != nil {
				addrStr := "???"
				if label.Address != nil {
					addrStr = fmt.Sprintf("0x%x", *label.Address)
				}
				idxStr := ""
				if label.InstructionIndex >= 0 {
					idxStr = fmt.Sprintf(" [instr %d]", label.InstructionIndex)
				}
				r.write("  %s @ %s%s\n", label.Name, addrStr, idxStr)
			}
		}
	}
}

func (r *REPL) printEvalResult(result *debuggerUI.EvalResult) {
	if result == nil {
		r.printError("Failed to evaluate expression")
		return
	}

	if result.Error != nil {
		r.printError(result.Error.Error())
		return
	}

	r.write("Result: 0x%x (%s)\n", result.Value, result.ValueString)
}

func (r *REPL) printLoadProgramResult(result *debuggerUI.LoadProgramResult) {
	if result == nil {
		r.printError("Failed to load program")
		return
	}

	if result.Error != nil {
		r.printError(result.Error.Error())
		return
	}

	r.write("Program loaded successfully\n")
}

func (r *REPL) printLoadSystemResult(result *debuggerUI.LoadSystemResult) {
	if result == nil {
		r.printError("Failed to load system")
		return
	}

	if result.Error != nil {
		r.printError(result.Error.Error())
		return
	}

	r.write("System loaded successfully\n")
}

func (r *REPL) printLoadRuntimeResult(result *debuggerUI.LoadRuntimeResult) {
	if result == nil {
		r.printError("Failed to load runtime")
		return
	}

	if result.Error != nil {
		r.printError(result.Error.Error())
		return
	}

	r.write("Runtime loaded successfully\n")
}

func (r *REPL) printLoadResult(result *debuggerUI.LoadResult) {
	if result == nil {
		r.printError("Failed to load")
		return
	}

	if result.Error != nil {
		r.printError(result.Error.Error())
		return
	}

	r.write("Load completed successfully\n")
}

// ============================================================================
// Machine-Readable Output Support
// ============================================================================

// startCommandOutput starts buffering output for a command (machine readable mode)
func (r *REPL) startCommandOutput() {
	if r.outputFormat == MachineReadable {
		r.outputBuffer.Reset()
		r.commandStarted = true
	}
}

// finishCommandOutput emits the buffered output as JSONL (machine readable mode)
func (r *REPL) finishCommandOutput(success bool, err error) {
	if r.outputFormat != MachineReadable || !r.commandStarted {
		return
	}

	output := CommandOutput{
		Command: r.lastInput,
		Output:  strings.TrimSpace(r.outputBuffer.String()),
		Success: success,
		Index:   r.commandIndex,
	}

	if err != nil {
		output.Error = err.Error()
	}

	// Add location information if running from a script
	if r.scriptFile != "" {
		output.File = r.scriptFile
		output.Line = r.scriptLine
	}

	jsonLine, jsonErr := json.Marshal(output)
	if jsonErr != nil {
		// Fallback: output error as plain text
		fmt.Fprintf(r.writer, "Error: failed to marshal output to JSON: %v\n", jsonErr)
		r.commandStarted = false
		r.outputBuffer.Reset()
		return
	}

	// Write JSONL directly to writer (not buffered)
	fmt.Fprintf(r.writer, "%s\n", jsonLine)
	r.commandStarted = false
	r.outputBuffer.Reset()
}

func (r *REPL) printAllSettings() {
	r.write("\nAvailable Settings:\n\n")

	categoryTree := r.settings.ListByCategory()

	// Print root-level settings first if any
	if len(categoryTree.Settings) > 0 {
		for _, setting := range categoryTree.Settings {
			r.write("  %s\n", setting.Name)
			r.write("    %s\n", setting.Description)
			r.write("    Default: %v\n\n", setting.DefaultValue)
		}
	}

	// Print categories in hierarchical order
	r.settings.IterateCategories(func(path []string, category *SettingCategory, indent string) {
		if len(category.Settings) > 0 || category.Description != "" {
			// Print category header with description if available
			categoryPath := strings.Join(path, ".")
			r.write("%s[%s]\n", indent, categoryPath)
			if category.Description != "" {
				r.write("%s  %s\n", indent, category.Description)
			}

			// Print settings in this category
			settingNames := make([]string, 0, len(category.Settings))
			for name := range category.Settings {
				settingNames = append(settingNames, name)
			}
			sort.Strings(settingNames)

			for _, settingName := range settingNames {
				setting := category.Settings[settingName]
				r.write("%s  %s\n", indent, setting.Name)
				r.write("%s    %s\n", indent, setting.Description)
				r.write("%s    Default: %v\n\n", indent, setting.DefaultValue)
			}
		}
	})
}

func (r *REPL) printCurrentSettings() {
	r.write("\nCurrent Settings:\n\n")

	categoryTree := r.settings.ListByCategory()

	// Print root-level settings first if any
	if len(categoryTree.Settings) > 0 {
		for _, setting := range categoryTree.Settings {
			r.write("  %s = %v\n", setting.Name, setting.Value)
		}
		r.write("\n")
	}

	// Print categories in hierarchical order
	r.settings.IterateCategories(func(path []string, category *SettingCategory, indent string) {
		if len(category.Settings) > 0 {
			// Print category header
			categoryPath := strings.Join(path, ".")
			r.write("%s[%s]\n", indent, categoryPath)

			// Print settings in this category
			settingNames := make([]string, 0, len(category.Settings))
			for name := range category.Settings {
				settingNames = append(settingNames, name)
			}
			sort.Strings(settingNames)

			for _, settingName := range settingNames {
				setting := category.Settings[settingName]
				r.write("%s  %s = %v\n", indent, setting.Name, setting.Value)
			}
			r.write("\n")
		}
	})
}

func (r *REPL) printEvent(eventType string, details map[string]interface{}) {
	r.write("[%s event]\n", eventType)
	for key, value := range details {
		r.write("  %s: %v\n", key, value)
	}
}

// write outputs text, buffering in machine-readable mode, direct output in human-readable mode
func (r *REPL) write(format string, args ...interface{}) {
	if r.outputFormat == MachineReadable && r.commandStarted {
		fmt.Fprintf(&r.outputBuffer, format, args...)
	} else {
		// If waiting for input, move output above the prompt line
		if r.waitingForInput {
			// Go to start of line, clear the line (clearing the prompt), then move to new line
			fmt.Fprintf(r.writer, "\r\033[K\n")
		}
		fmt.Fprintf(r.writer, format, args...)
	}
}

func (r *REPL) printLogEntry(entry logging.UILogEntry) {
	fmt.Fprintf(r.writer, "%s\n", entry.String())
}
