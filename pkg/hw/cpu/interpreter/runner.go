// Package interpreter provides an automatic interpreter for Cucaracha machine code.
//
// # Runner - High-Level Program Execution
//
// This file provides high-level APIs for loading and executing Cucaracha programs.
// It abstracts the details of program loading, memory layout, and execution setup
// so that CLI tools can focus on user interaction rather than implementation details.
//
// The typical execution flow is:
//
//  1. Create a Runner with NewRunner(memorySize)
//  2. Load a program with runner.LoadProgram(pf)
//  3. Execute with runner.Run() or runner.RunWithTrace(callback)
//  4. Get results with runner.Result() and runner.ReturnValue()
//
// For simpler usage, use RunFile() which handles the entire flow:
//
//	result, err := interpreter.RunFile("program.c", nil)
//	fmt.Println("Return value:", result.ReturnValue)
package interpreter

import (
	"fmt"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
)

// DefaultMemorySize is the default memory size for the interpreter (128KB)
const DefaultMemorySize uint32 = 0x20000

// DefaultBaseAddress is the default base address for loading code
const DefaultBaseAddress uint32 = 0x10000

// TerminationAddress is the magic address that signals program termination
// When execution reaches this address, the program has returned from main
const TerminationAddress uint32 = 0x0000FFFC

// Runner provides high-level program execution functionality.
// It wraps the Interpreter and Debugger to provide a simpler API for
// loading and running programs.
type Runner struct {
	interp    *Interpreter
	dbg       *Debugger
	program   mc.ProgramFile
	addrToIdx map[uint32]int
	result    *ExecutionResult
}

// NewRunner creates a new Runner with the specified memory size.
// Use DefaultMemorySize if unsure.
func NewRunner(memorySize uint32) *Runner {
	interp := NewInterpreter(memorySize)
	dbg := NewDebugger(interp)
	dbg.AddTerminationAddress(TerminationAddress)

	return &Runner{
		interp:    interp,
		dbg:       dbg,
		addrToIdx: make(map[uint32]int),
	}
}

// LoadProgram loads a resolved ProgramFile into memory and sets up execution.
// The program must already be resolved (have addresses assigned).
// Returns an error if loading fails.
func (r *Runner) LoadProgram(pf mc.ProgramFile) error {
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

		// Build address map
		r.addrToIdx[*instr.Address] = i

		// Encode the instruction to binary
		rawInstr := instr.Instruction.Raw()
		encoded := rawInstr.Encode()

		// Write to memory
		if err := r.interp.state.WriteMemory32(*instr.Address, encoded); err != nil {
			return fmt.Errorf("failed to write instruction %d at 0x%08X: %w", i, *instr.Address, err)
		}
	}

	// Load global data
	for _, global := range pf.Globals() {
		if global.Address == nil {
			continue // Skip unresolved globals
		}
		if len(global.InitialData) > 0 {
			addr := *global.Address
			for j, b := range global.InitialData {
				if int(addr)+j >= len(r.interp.state.Memory) {
					return fmt.Errorf("global '%s' data exceeds memory bounds", global.Name)
				}
				r.interp.state.Memory[addr+uint32(j)] = b
			}
		}
	}

	// Set initial PC to the start of code
	r.interp.state.PC = layout.CodeStart

	// Find main function and set entry point
	if mainFunc, hasMain := pf.Functions()["main"]; hasMain && len(mainFunc.InstructionRanges) > 0 {
		startIdx := mainFunc.InstructionRanges[0].Start
		if startIdx < len(instrList) && instrList[startIdx].Address != nil {
			r.interp.state.PC = *instrList[startIdx].Address
			*r.interp.state.LR = TerminationAddress
		}
	}

	r.program = pf
	return nil
}

// State returns the current CPU state.
func (r *Runner) State() *CPUState {
	return r.interp.State()
}

// Debugger returns the underlying debugger for advanced usage.
func (r *Runner) Debugger() *Debugger {
	return r.dbg
}

// Program returns the loaded program.
func (r *Runner) Program() mc.ProgramFile {
	return r.program
}

// DebugInfo returns the debug information from the loaded program, or nil if none.
func (r *Runner) DebugInfo() *mc.DebugInfo {
	if r.program == nil {
		return nil
	}
	return r.program.DebugInfo()
}

// Run executes the program until termination or error.
// maxSteps limits the number of instructions (0 = unlimited).
// Returns the execution result.
func (r *Runner) Run(maxSteps int) *ExecutionResult {
	r.result = r.dbg.Run(maxSteps)
	return r.result
}

// Result returns the result of the last execution, or nil if not run yet.
func (r *Runner) Result() *ExecutionResult {
	return r.result
}

// ReturnValue returns the program's return value (r0 register).
func (r *Runner) ReturnValue() int32 {
	return int32(r.interp.state.Registers[16])
}

// PC returns the current program counter.
func (r *Runner) PC() uint32 {
	return r.interp.state.PC
}

// IsNormalExit returns true if the program exited normally (returned from main).
func (r *Runner) IsNormalExit() bool {
	if r.result == nil {
		return false
	}
	if r.result.StopReason == StopTermination {
		return true
	}
	// Also check if we jumped past end of code
	if r.program != nil {
		layout := r.program.MemoryLayout()
		if layout != nil {
			endOfCode := layout.CodeStart + layout.CodeSize
			return r.result.Error != nil && r.interp.state.PC >= endOfCode
		}
	}
	return false
}

// GetInstructionAt returns the instruction at the given address, or nil if not found.
func (r *Runner) GetInstructionAt(addr uint32) *mc.Instruction {
	if r.program == nil {
		return nil
	}
	idx, ok := r.addrToIdx[addr]
	if !ok {
		return nil
	}
	instrs := r.program.Instructions()
	if idx >= len(instrs) {
		return nil
	}
	return &instrs[idx]
}

// TraceCallback is called for each instruction during traced execution.
// It receives the step number, program counter, and instruction text.
// Return true to continue execution, false to stop.
type TraceCallback func(step int, pc uint32, instrText string, srcLoc *mc.SourceLocation) bool

// RunWithTrace executes with per-instruction callbacks for tracing.
// The callback is called before each instruction executes.
// maxSteps limits the number of instructions (0 = unlimited).
func (r *Runner) RunWithTrace(maxSteps int, callback TraceCallback) *ExecutionResult {
	debugInfo := r.DebugInfo()
	if debugInfo != nil {
		debugInfo.TryLoadSourceFiles()
	}

	stepCount := 0
	var lastSourceLoc *mc.SourceLocation

	r.dbg.SetEventCallback(func(event ExecutionEvent, result *ExecutionResult) bool {
		if event == EventStep {
			pc := result.LastPC
			instrText := "???"
			if instr := r.GetInstructionAt(pc); instr != nil {
				instrText = instr.Text
			}

			// Get source location
			var srcLoc *mc.SourceLocation
			if debugInfo != nil {
				if loc := debugInfo.GetSourceLocation(pc); loc != nil && loc.IsValid() {
					// Only report if changed
					if lastSourceLoc == nil || lastSourceLoc.File != loc.File || lastSourceLoc.Line != loc.Line {
						lastSourceLoc = loc
						srcLoc = loc
					}
				}
			}

			cont := callback(stepCount, pc, instrText, srcLoc)
			stepCount++
			return cont
		}
		return true
	})

	r.result = r.dbg.Run(maxSteps)
	r.dbg.SetEventCallback(nil) // Clear callback
	return r.result
}

// Step executes a single instruction.
func (r *Runner) Step() *ExecutionResult {
	r.result = r.dbg.Step()
	return r.result
}

// Continue executes until a breakpoint or termination.
func (r *Runner) Continue() *ExecutionResult {
	r.result = r.dbg.Continue()
	return r.result
}

// AddBreakpoint adds a breakpoint at the given address.
func (r *Runner) AddBreakpoint(addr uint32) *Breakpoint {
	return r.dbg.AddBreakpoint(addr)
}

// RemoveBreakpoint removes a breakpoint by ID.
func (r *Runner) RemoveBreakpoint(id int) bool {
	return r.dbg.RemoveBreakpoint(id)
}

// ListBreakpoints returns all breakpoints.
func (r *Runner) ListBreakpoints() []*Breakpoint {
	return r.dbg.ListBreakpoints()
}
