package debugger

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/interpreter"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/registers"
)

// Register index helpers
// r0-r9 map to indices 16-25 (general purpose registers start at offset 16)
const (
	regR0Idx   = 16
	regR9Idx   = 25
	regCountGP = 10 // General purpose register count (r0-r9)
)

// getRegisterIndex returns the register index for a named register
func getRegisterIndex(name string) uint32 {
	name = strings.ToLower(name)

	// Handle numbered registers r0-r9
	if strings.HasPrefix(name, "r") && len(name) >= 2 {
		if num, err := strconv.Atoi(name[1:]); err == nil && num >= 0 && num <= 9 {
			return uint32(regR0Idx + num)
		}
	}

	// Use the registers package for named registers
	return uint32(registers.Register(name).Encode())
}

// Backend implements DebuggerBackend using an interpreter.Runner
type Backend struct {
	runner     *interpreter.Runner
	addrToIdx  map[uint32]int
	idxToAddr  map[int]uint32
	lastResult *ExecutionResult // Last execution result for termination check
}

// NewBackend creates a new debugger backend with the given memory size
func NewBackend(memorySize uint32) *Backend {
	return &Backend{
		runner:    interpreter.NewRunner(memorySize),
		addrToIdx: make(map[uint32]int),
		idxToAddr: make(map[int]uint32),
	}
}

// NewBackendWithRunner creates a new debugger backend using an existing runner
func NewBackendWithRunner(runner *interpreter.Runner) *Backend {
	return &Backend{
		runner:    runner,
		addrToIdx: make(map[uint32]int),
		idxToAddr: make(map[int]uint32),
	}
}

// Runner returns the underlying runner (for advanced use cases)
func (b *Backend) Runner() *interpreter.Runner {
	return b.runner
}

// Interrupt signals the debugger to stop execution.
// This is safe to call from signal handlers or other goroutines.
func (b *Backend) Interrupt() {
	b.runner.Debugger().Interrupt()
}

// SetExecutionDelay sets the delay between instruction executions in milliseconds.
// Use 0 for full speed, higher values for slower execution (slow-motion mode).
// Predefined speeds: 0=instant, 50=fast, 100=normal, 250=slow, 500=very slow
// Deprecated: Use SetTargetSpeed for more accurate timing control.
func (b *Backend) SetExecutionDelay(delayMs int) {
	b.runner.Debugger().SetExecutionDelay(delayMs)
}

// GetExecutionDelay returns the current execution delay in milliseconds.
// Deprecated: Use GetTargetSpeed instead.
func (b *Backend) GetExecutionDelay() int {
	return b.runner.Debugger().GetExecutionDelay()
}

// SetTargetSpeed sets the target execution speed in Hz (cycles per second).
// Use 0 for unlimited speed (no timing simulation).
// For example, 1000 Hz means 1000 cycles per second.
func (b *Backend) SetTargetSpeed(hz float64) {
	b.runner.SetTargetSpeed(hz)
}

// GetTargetSpeed returns the current target execution speed in Hz.
func (b *Backend) GetTargetSpeed() float64 {
	return b.runner.GetTargetSpeed()
}

// ExecutionCallback is called during execution on each step.
// Return true to continue execution, false to stop.
type ExecutionCallback func(event interpreter.ExecutionEvent, stepsExecuted int, pc uint32) bool

// FullExecutionCallback is called during execution with full result information.
// Includes lagging info and other timing details.
// Return true to continue execution, false to stop.
type FullExecutionCallback func(event interpreter.ExecutionEvent, result *interpreter.ExecutionResult) bool

// SetExecutionCallback sets a callback to be invoked during execution.
// The callback receives the event type, steps executed so far, and current PC.
// It should return true to continue execution, false to stop.
// Pass nil to clear the callback.
func (b *Backend) SetExecutionCallback(callback ExecutionCallback) {
	if callback == nil {
		b.runner.Debugger().SetEventCallback(nil)
		return
	}

	b.runner.Debugger().SetEventCallback(func(event interpreter.ExecutionEvent, result *interpreter.ExecutionResult) bool {
		return callback(event, result.StepsExecuted, result.LastPC)
	})
}

// SetFullExecutionCallback sets a callback with full result information.
// Use this to receive lagging events and detailed timing info.
// Pass nil to clear the callback.
func (b *Backend) SetFullExecutionCallback(callback FullExecutionCallback) {
	if callback == nil {
		b.runner.Debugger().SetEventCallback(nil)
		return
	}

	b.runner.Debugger().SetEventCallback(func(event interpreter.ExecutionEvent, result *interpreter.ExecutionResult) bool {
		return callback(event, result)
	})
}

// HasExecutionCallback returns true if an execution callback is set (for debugging)
func (b *Backend) HasExecutionCallback() bool {
	return b.runner.Debugger().HasEventCallback()
}

// IsTerminated returns true if the program has terminated.
func (b *Backend) IsTerminated() bool {
	if b.lastResult == nil {
		return false
	}
	return b.lastResult.StopReason == interpreter.StopTermination
}

// LoadProgram loads a program into the backend
func (b *Backend) LoadProgram(program mc.ProgramFile) error {
	if err := b.runner.LoadProgram(program); err != nil {
		return err
	}

	// Build address maps
	b.addrToIdx = make(map[uint32]int)
	b.idxToAddr = make(map[int]uint32)
	for i, instr := range program.Instructions() {
		if instr.Address != nil {
			b.addrToIdx[*instr.Address] = i
			b.idxToAddr[i] = *instr.Address
		}
	}

	return nil
}

// Program returns the loaded program
func (b *Backend) Program() mc.ProgramFile {
	return b.runner.Program()
}

// DebugInfo returns the debug information
func (b *Backend) DebugInfo() *mc.DebugInfo {
	return b.runner.DebugInfo()
}

// Step executes count instructions
func (b *Backend) Step(count int) ExecutionResult {
	var lastResult ExecutionResult

	for i := 0; i < count; i++ {
		result := b.runner.Debugger().Step()
		lastResult = ExecutionResult{
			StopReason:    result.StopReason,
			StepsExecuted: result.StepsExecuted,
			Error:         result.Error,
			BreakpointID:  result.BreakpointID,
			WatchpointID:  result.WatchpointID,
			LastPC:        result.LastPC,
		}

		if result.Error != nil {
			break
		}
		if result.StopReason != interpreter.StopStep && result.StopReason != interpreter.StopNone {
			break
		}
	}

	// Get return value if terminated
	if lastResult.StopReason == interpreter.StopTermination {
		lastResult.ReturnValue = b.runner.Debugger().State().Registers[regR0Idx]
	}

	return lastResult
}

// Continue runs until a breakpoint or termination
func (b *Backend) Continue() ExecutionResult {
	result := b.runner.Debugger().Continue() // Continue takes no arguments
	execResult := ExecutionResult{
		StopReason:     result.StopReason,
		StepsExecuted:  result.StepsExecuted,
		CyclesExecuted: result.CyclesExecuted,
		Error:          result.Error,
		BreakpointID:   result.BreakpointID,
		WatchpointID:   result.WatchpointID,
		LastPC:         result.LastPC,
		Lagging:        result.Lagging,
		LagCycles:      result.LagCycles,
	}

	if execResult.StopReason == interpreter.StopTermination {
		execResult.ReturnValue = b.runner.Debugger().State().Registers[regR0Idx]
	}

	return execResult
}

// Run runs until termination
func (b *Backend) Run() ExecutionResult {
	result := b.runner.Debugger().Run(0) // 0 = no limit
	execResult := ExecutionResult{
		StopReason:     result.StopReason,
		StepsExecuted:  result.StepsExecuted,
		CyclesExecuted: result.CyclesExecuted,
		Error:          result.Error,
		BreakpointID:   result.BreakpointID,
		WatchpointID:   result.WatchpointID,
		LastPC:         result.LastPC,
		Lagging:        result.Lagging,
		LagCycles:      result.LagCycles,
	}

	if execResult.StopReason == interpreter.StopTermination {
		execResult.ReturnValue = b.runner.Debugger().State().Registers[regR0Idx]
	}

	// Store result for termination check
	b.lastResult = &execResult

	return execResult
}

// Next executes until the next source line, stepping over function calls
// If the current instruction is a call, it sets a temporary breakpoint at the return address
// and continues until that breakpoint is hit
func (b *Backend) Next(count int) ExecutionResult {
	var lastResult ExecutionResult

	for i := 0; i < count; i++ {
		result := b.nextOne()
		lastResult = result

		if result.Error != nil {
			break
		}
		if result.StopReason != interpreter.StopStep && result.StopReason != interpreter.StopNone {
			break
		}
	}

	return lastResult
}

// nextOne executes one "next" step (stepping over calls)
func (b *Backend) nextOne() ExecutionResult {
	pc := b.runner.State().PC

	// Check if the current instruction is a call
	if b.isCallInstruction(pc) {
		// Set temporary breakpoint at return address (PC + 4)
		returnAddr := pc + 4

		// Add temporary breakpoint
		bp, err := b.AddBreakpoint(returnAddr)
		if err != nil {
			// If we can't set breakpoint, just do a regular step
			return b.Step(1)
		}

		// Continue execution
		result := b.Continue()

		// Remove temporary breakpoint
		b.RemoveBreakpoint(bp.ID)

		// If we stopped at our temporary breakpoint, report as step
		if result.StopReason == interpreter.StopBreakpoint && b.runner.State().PC == returnAddr {
			result.StopReason = interpreter.StopStep
		}

		return result
	}

	// Not a call, just do a regular step
	return b.Step(1)
}

// isCallInstruction checks if the instruction at addr is a function call
// A branch is a call if the branch target is a function symbol
func (b *Backend) isCallInstruction(addr uint32) bool {
	program := b.runner.Program()
	if program == nil {
		return false
	}

	idx, found := b.addrToIdx[addr]
	if !found {
		return false
	}

	instrs := program.Instructions()
	if idx >= len(instrs) {
		return false
	}

	instr := instrs[idx]
	if instr.Instruction == nil || instr.Instruction.Descriptor == nil {
		return false
	}

	// Check if it's a branch instruction (JMP or CJMP)
	mnemonic := strings.ToUpper(instr.Instruction.Descriptor.OpCode.Mnemonic)
	if mnemonic != "JMP" && mnemonic != "CJMP" {
		return false
	}

	// Check if the branch target is a function
	// Use the same logic as getBranchTarget - backtrack to find the MOVIMM16L/H
	// that loads the target register, and check if the symbol is a function
	return b.isBranchTargetFunction(idx, instrs)
}

// isBranchTargetFunction checks if the branch target at the given instruction index is a function
func (b *Backend) isBranchTargetFunction(instrIdx int, instrs []mc.Instruction) bool {
	instr := instrs[instrIdx]

	// Get mnemonic
	if instr.Instruction == nil || instr.Instruction.Descriptor == nil {
		return false
	}
	mnemonic := strings.ToUpper(instr.Instruction.Descriptor.OpCode.Mnemonic)

	// Determine which operand is the target register
	targetRegIdx := 0
	if mnemonic == "CJMP" && len(instr.Instruction.OperandValues) >= 2 {
		targetRegIdx = 1 // CJMP: condcode, target, link
	}

	// Get the target register
	if targetRegIdx >= len(instr.Instruction.OperandValues) {
		return false
	}
	targetOp := instr.Instruction.OperandValues[targetRegIdx]
	if targetOp.Kind() != instructions.OperandKind_Register {
		return false
	}
	targetReg := targetOp.Register()
	if targetReg == nil {
		return false
	}
	targetRegName := targetReg.Name()

	// Backtrack through previous instructions looking for MOVIMM16L/MOVIMM16H
	// that write to this register and have a function symbol
	for i := instrIdx - 1; i >= 0 && i >= instrIdx-20; i-- {
		prevInstr := instrs[i]
		if prevInstr.Instruction == nil || prevInstr.Instruction.Descriptor == nil {
			continue
		}

		prevMnemonic := strings.ToUpper(prevInstr.Instruction.Descriptor.OpCode.Mnemonic)

		// Check if this instruction writes to our target register with an immediate
		if (prevMnemonic == "MOVIMM16L" || prevMnemonic == "MOVIMM16H") &&
			len(prevInstr.Instruction.OperandValues) >= 2 {

			// MOVIMM16L/H format: imm, dest_reg
			destOp := prevInstr.Instruction.OperandValues[1]
			if destOp.Kind() == instructions.OperandKind_Register {
				destReg := destOp.Register()
				if destReg != nil && destReg.Name() == targetRegName {
					// Found an immediate load to our target register
					// Check for associated function symbol
					for _, sym := range prevInstr.Symbols {
						if sym.Function != nil {
							return true // Branch target is a function - this is a call
						}
					}
				}
			}
		}

		// If we found a different instruction that writes to our register, stop
		if prevMnemonic != "MOVIMM16L" && prevMnemonic != "MOVIMM16H" {
			if prevInstr.Instruction != nil && prevInstr.Instruction.Descriptor != nil {
				for opIdx, opDesc := range prevInstr.Instruction.Descriptor.Operands {
					if opIdx < len(prevInstr.Instruction.OperandValues) &&
						opDesc.Role == instructions.OperandRole_Destination {
						destOp := prevInstr.Instruction.OperandValues[opIdx]
						if destOp.Kind() == instructions.OperandKind_Register {
							destReg := destOp.Register()
							if destReg != nil && destReg.Name() == targetRegName {
								// Different instruction writes to target reg, stop backtracking
								return false
							}
						}
					}
				}
			}
		}
	}

	return false
}

// Reset resets the program state
func (b *Backend) Reset() error {
	return b.runner.Reset()
}

// GetState returns the current debugger state
func (b *Backend) GetState() DebuggerState {
	state := b.runner.Debugger().State()

	regs := make([]RegisterInfo, 0)
	regNames := []string{"r0", "r1", "r2", "r3", "r4", "r5", "r6", "r7", "r8", "r9", "sp", "lr", "pc", "cpsr"}
	regIndices := make([]uint32, len(regNames))
	for i, name := range regNames {
		regIndices[i] = getRegisterIndex(name)
	}
	cpsrIdx := getRegisterIndex("cpsr")

	for i, name := range regNames {
		idx := regIndices[i]
		var value uint32
		if name == "pc" {
			value = state.PC
		} else if name == "sp" {
			value = *state.SP
		} else if name == "lr" {
			value = *state.LR
		} else if name == "cpsr" {
			value = state.Registers[cpsrIdx]
		} else {
			value = state.Registers[idx]
		}
		regs = append(regs, RegisterInfo{
			Name:  name,
			Index: idx,
			Value: value,
		})
	}

	cpsr := state.Registers[cpsrIdx]
	flags := FlagState{
		N: cpsr&(1<<31) != 0,
		Z: cpsr&(1<<30) != 0,
		C: cpsr&(1<<29) != 0,
		V: cpsr&(1<<28) != 0,
	}

	return DebuggerState{
		PC:        state.PC,
		SP:        *state.SP,
		LR:        *state.LR,
		CPSR:      cpsr,
		Registers: regs,
		Flags:     flags,
		IsRunning: false, // TODO: track running state
	}
}

// ReadRegister reads a register by name
func (b *Backend) ReadRegister(name string) (uint32, error) {
	name = strings.ToLower(name)
	state := b.runner.Debugger().State()

	switch name {
	case "sp":
		return *state.SP, nil
	case "lr":
		return *state.LR, nil
	case "pc":
		return state.PC, nil
	}

	// Handle numbered registers r0-r9 and cpsr
	if strings.HasPrefix(name, "r") && len(name) >= 2 {
		if n, err := strconv.Atoi(name[1:]); err == nil && n >= 0 && n <= 9 {
			return state.Registers[regR0Idx+n], nil
		}
	}

	if name == "cpsr" {
		return state.Registers[getRegisterIndex("cpsr")], nil
	}

	return 0, fmt.Errorf("unknown register: %s", name)
}

// WriteRegister writes a value to a register
func (b *Backend) WriteRegister(name string, value uint32) error {
	name = strings.ToLower(name)
	state := b.runner.Debugger().State()

	switch name {
	case "sp":
		*state.SP = value
		return nil
	case "lr":
		*state.LR = value
		return nil
	case "pc":
		state.PC = value
		return nil
	}

	// Handle numbered registers r0-r9
	if strings.HasPrefix(name, "r") && len(name) >= 2 {
		if n, err := strconv.Atoi(name[1:]); err == nil && n >= 0 && n <= 9 {
			state.Registers[regR0Idx+n] = value
			return nil
		}
	}

	if name == "cpsr" {
		state.Registers[getRegisterIndex("cpsr")] = value
		return nil
	}

	return fmt.Errorf("unknown register: %s", name)
}

// ReadMemory reads memory from the given address
func (b *Backend) ReadMemory(addr uint32, size int) ([]byte, error) {
	return b.runner.Debugger().ReadMemory(addr, size)
}

// WriteMemory writes memory at the given address
func (b *Backend) WriteMemory(addr uint32, data []byte) error {
	return b.runner.Debugger().WriteMemory(addr, data)
}

// AddBreakpoint adds a breakpoint at the given address
func (b *Backend) AddBreakpoint(addr uint32) (*interpreter.Breakpoint, error) {
	bp := b.runner.Debugger().AddBreakpoint(addr)
	return bp, nil
}

// RemoveBreakpoint removes a breakpoint by ID
func (b *Backend) RemoveBreakpoint(id int) error {
	if !b.runner.Debugger().RemoveBreakpoint(id) {
		return fmt.Errorf("breakpoint %d not found", id)
	}
	return nil
}

// ListBreakpoints returns all breakpoints
func (b *Backend) ListBreakpoints() []*interpreter.Breakpoint {
	return b.runner.Debugger().ListBreakpoints()
}

// EnableBreakpoint enables or disables a breakpoint
func (b *Backend) EnableBreakpoint(id int, enabled bool) error {
	if !b.runner.Debugger().EnableBreakpoint(id, enabled) {
		return fmt.Errorf("breakpoint %d not found", id)
	}
	return nil
}

// AddWatchpoint adds a watchpoint at the given address
func (b *Backend) AddWatchpoint(addr uint32) (*interpreter.Watchpoint, error) {
	wp := b.runner.Debugger().AddWatchpoint(addr, 4, interpreter.WatchWrite)
	return wp, nil
}

// RemoveWatchpoint removes a watchpoint by ID
func (b *Backend) RemoveWatchpoint(id int) error {
	if !b.runner.Debugger().RemoveWatchpoint(id) {
		return fmt.Errorf("watchpoint %d not found", id)
	}
	return nil
}

// ListWatchpoints returns all watchpoints
func (b *Backend) ListWatchpoints() []*interpreter.Watchpoint {
	return b.runner.Debugger().ListWatchpoints()
}

// Disassemble disassembles instructions at the given address
func (b *Backend) Disassemble(addr uint32, count int) ([]InstructionInfo, error) {
	program := b.runner.Program()
	if program == nil {
		return nil, fmt.Errorf("no program loaded")
	}

	instrs := program.Instructions()
	result := make([]InstructionInfo, 0, count)

	// Find starting index
	startIdx, found := b.addrToIdx[addr]
	if !found {
		// Try to find closest instruction
		for i, instr := range instrs {
			if instr.Address != nil && *instr.Address >= addr {
				startIdx = i
				found = true
				break
			}
		}
	}

	if !found {
		return nil, fmt.Errorf("address 0x%08X not found in program", addr)
	}

	pc := b.runner.Debugger().State().PC

	for i := startIdx; i < len(instrs) && len(result) < count; i++ {
		instr := instrs[i]
		if instr.Address == nil {
			continue
		}

		info := InstructionInfo{
			Address:       *instr.Address,
			IsCurrentPC:   *instr.Address == pc,
			HasBreakpoint: b.runner.Debugger().GetBreakpointAt(*instr.Address) != nil,
		}

		// Get encoding from Raw instruction
		if instr.Raw != nil {
			info.Encoding = instr.Raw.Encode()
		}

		// Get mnemonic and operands from the fully decoded instruction
		if instr.Instruction != nil {
			if instr.Instruction.Descriptor != nil {
				info.Mnemonic = instr.Instruction.Descriptor.OpCode.Mnemonic
			}
			info.Operands = formatOperands(instr.Instruction)
		} else {
			// Fall back to the text representation
			info.Mnemonic = instr.Text
		}

		// Extract branch target from symbols (for CFG visualization)
		info.BranchTarget, info.BranchTargetSym = b.getBranchTarget(i, instrs)

		result = append(result, info)
	}

	return result, nil
}

// getBranchTarget extracts branch target address and symbol from an instruction.
// Cucaracha branch instructions (JMP, CJMP) use register-indirect addressing,
// so we backtrack to find MOVIMM16L/MOVIMM16H instructions that load the target register.
func (b *Backend) getBranchTarget(instrIdx int, instrs []mc.Instruction) (uint32, string) {
	instr := instrs[instrIdx]

	// Check if this is a branch instruction
	mnemonic := ""
	if instr.Instruction != nil && instr.Instruction.Descriptor != nil {
		mnemonic = strings.ToUpper(instr.Instruction.Descriptor.OpCode.Mnemonic)
	}

	if mnemonic != "JMP" && mnemonic != "CJMP" && !strings.HasPrefix(mnemonic, "B") {
		return 0, ""
	}

	// Backtrack to find MOVIMM16L/MOVIMM16H that loads the target register
	// For JMP: first operand is target register
	// For CJMP: second operand is target register
	if instr.Instruction == nil || len(instr.Instruction.OperandValues) == 0 {
		return 0, ""
	}

	// Determine which operand is the target register
	targetRegIdx := 0
	if mnemonic == "CJMP" && len(instr.Instruction.OperandValues) >= 2 {
		targetRegIdx = 1 // CJMP: condcode, target, link
	}

	// Get the target register
	if targetRegIdx >= len(instr.Instruction.OperandValues) {
		return 0, ""
	}
	targetOp := instr.Instruction.OperandValues[targetRegIdx]
	if targetOp.Kind() != instructions.OperandKind_Register {
		return 0, ""
	}
	targetReg := targetOp.Register()
	if targetReg == nil {
		return 0, ""
	}
	targetRegName := targetReg.Name()

	// Backtrack through previous instructions looking for MOVIMM16L/MOVIMM16H
	// that write to this register
	var loValue, hiValue uint32
	foundLo, foundHi := false, false

	for i := instrIdx - 1; i >= 0 && i >= instrIdx-20; i-- {
		prevInstr := instrs[i]
		if prevInstr.Instruction == nil || prevInstr.Instruction.Descriptor == nil {
			continue
		}

		prevMnemonic := strings.ToUpper(prevInstr.Instruction.Descriptor.OpCode.Mnemonic)

		// Check if this instruction writes to our target register
		if (prevMnemonic == "MOVIMM16L" || prevMnemonic == "MOVIMM16H") &&
			len(prevInstr.Instruction.OperandValues) >= 2 {

			// MOVIMM16L/H format: imm, dest_reg
			destOp := prevInstr.Instruction.OperandValues[1]
			if destOp.Kind() == instructions.OperandKind_Register {
				destReg := destOp.Register()
				if destReg != nil && destReg.Name() == targetRegName {
					// Found an immediate load to our target register
					immOp := prevInstr.Instruction.OperandValues[0]
					if immOp.Kind() == instructions.OperandKind_Immediate {
						imm := immOp.Immediate()
						immVal := uint32(imm.Encode())
						if prevMnemonic == "MOVIMM16L" {
							loValue = immVal & 0xFFFF
							foundLo = true
						} else {
							hiValue = (immVal & 0xFFFF) << 16
							foundHi = true
						}

						// Check for associated symbol
						for _, sym := range prevInstr.Symbols {
							if sym.Label != nil || sym.Function != nil {
								symAddr, ok := mc.GetSymbolAddressFromProgram(&sym, b.runner.Program())
								if ok {
									return symAddr, sym.Name
								}
							}
						}
					}
				}
			}
		}

		// If we found a different instruction that writes to our register, stop
		// (the value might have been overwritten)
		if prevMnemonic != "MOVIMM16L" && prevMnemonic != "MOVIMM16H" {
			// Check if this instruction writes to targetReg
			if prevInstr.Instruction != nil && prevInstr.Instruction.Descriptor != nil {
				for opIdx, opDesc := range prevInstr.Instruction.Descriptor.Operands {
					if opIdx < len(prevInstr.Instruction.OperandValues) &&
						opDesc.Role == instructions.OperandRole_Destination {
						destOp := prevInstr.Instruction.OperandValues[opIdx]
						if destOp.Kind() == instructions.OperandKind_Register {
							destReg := destOp.Register()
							if destReg != nil && destReg.Name() == targetRegName {
								// Different instruction writes to target reg, stop backtracking
								break
							}
						}
					}
				}
			}
		}

		// If we found both lo and hi, we can compute the address
		if foundLo && foundHi {
			return loValue | hiValue, ""
		}
	}

	// If we only found lo (16-bit address), return it
	if foundLo {
		return loValue, ""
	}

	return 0, ""
}

// formatOperands formats instruction operands for display
func formatOperands(instr *instructions.Instruction) string {
	if instr == nil || len(instr.OperandValues) == 0 {
		return ""
	}

	parts := make([]string, 0, len(instr.OperandValues))
	for _, op := range instr.OperandValues {
		parts = append(parts, op.String())
	}
	return strings.Join(parts, ", ")
}

// EvalExpression evaluates an expression and returns the result
func (b *Backend) EvalExpression(expr string) (uint32, error) {
	eval := NewExpressionEvaluator(b)
	return eval.Eval(expr)
}

// ResolveSymbol resolves a symbol name to an address
func (b *Backend) ResolveSymbol(name string) (uint32, error) {
	program := b.runner.Program()
	if program == nil {
		return 0, fmt.Errorf("no program loaded")
	}

	// Check functions
	if funcs := program.Functions(); funcs != nil {
		if fn, ok := funcs[name]; ok {
			if len(fn.InstructionRanges) > 0 {
				idx := fn.InstructionRanges[0].Start
				if addr, ok := b.idxToAddr[idx]; ok {
					return addr, nil
				}
			}
		}
	}

	// Check globals
	for _, g := range program.Globals() {
		if g.Name == name && g.Address != nil {
			return *g.Address, nil
		}
	}

	// Check labels
	for _, l := range program.Labels() {
		if l.Name == name {
			if addr, ok := b.idxToAddr[l.InstructionIndex]; ok {
				return addr, nil
			}
		}
	}

	return 0, fmt.Errorf("unknown symbol: %s", name)
}

// GetSymbolAt returns the symbol name at the given address
func (b *Backend) GetSymbolAt(addr uint32) (string, bool) {
	program := b.runner.Program()
	if program == nil {
		return "", false
	}

	// Check functions
	if funcs := program.Functions(); funcs != nil {
		for name, fn := range funcs {
			if len(fn.InstructionRanges) > 0 {
				idx := fn.InstructionRanges[0].Start
				if fnAddr, ok := b.idxToAddr[idx]; ok && fnAddr == addr {
					return name, true
				}
			}
		}
	}

	// Check labels
	for _, l := range program.Labels() {
		if labelAddr, ok := b.idxToAddr[l.InstructionIndex]; ok && labelAddr == addr {
			return l.Name, true
		}
	}

	return "", false
}

// GetSourceLocation returns the source location for the given PC
func (b *Backend) GetSourceLocation(pc uint32) *mc.SourceLocation {
	debugInfo := b.runner.DebugInfo()
	if debugInfo == nil {
		return nil
	}
	return debugInfo.GetSourceLocation(pc)
}

// ResolveSourceLocation finds the first instruction address for a given source file and line
// If file is empty, uses the current source file (based on PC)
func (b *Backend) ResolveSourceLocation(file string, line int) (uint32, error) {
	debugInfo := b.runner.DebugInfo()
	if debugInfo == nil {
		return 0, fmt.Errorf("no debug info available")
	}

	// If no file specified, use current file
	if file == "" {
		currentLoc := b.GetSourceLocation(b.runner.State().PC)
		if currentLoc == nil || currentLoc.File == "" {
			return 0, fmt.Errorf("no current source file")
		}
		file = currentLoc.File
	}

	// Find the first instruction that matches the file and line
	var bestAddr uint32
	var found bool

	for addr, loc := range debugInfo.InstructionLocations {
		if loc.File == file && loc.Line == line {
			if !found || addr < bestAddr {
				bestAddr = addr
				found = true
			}
		}
	}

	// Also try matching just the basename for convenience
	if !found {
		fileBase := filepath.Base(file)
		for addr, loc := range debugInfo.InstructionLocations {
			locBase := filepath.Base(loc.File)
			if locBase == fileBase && loc.Line == line {
				if !found || addr < bestAddr {
					bestAddr = addr
					found = true
				}
			}
		}
	}

	if !found {
		if file != "" {
			return 0, fmt.Errorf("no instruction found for %s:%d", file, line)
		}
		return 0, fmt.Errorf("no instruction found for line %d", line)
	}

	return bestAddr, nil
}

// GetVariables returns the variables accessible at the given PC
func (b *Backend) GetVariables(pc uint32) []VariableValue {
	debugInfo := b.runner.DebugInfo()
	if debugInfo == nil {
		return nil
	}

	vars := debugInfo.GetVariables(pc)
	result := make([]VariableValue, 0, len(vars))

	for _, v := range vars {
		vv := VariableValue{
			Name:     v.Name,
			TypeName: v.TypeName,
			Size:     v.Size,
			Location: formatVariableLocation(v.Location),
		}

		// Try to read the value and format it
		if v.Location == nil {
			vv.ValueString = "<optimized out>"
		} else {
			vv.Value = b.readVariableValue(v)
			if vv.Value != nil {
				vv.ValueString = formatVariableValue(vv.Value)
			} else {
				vv.ValueString = "<unavailable>"
			}
		}

		result = append(result, vv)
	}

	return result
}

// formatVariableValue formats a variable value for display
func formatVariableValue(val interface{}) string {
	switch v := val.(type) {
	case uint32:
		return fmt.Sprintf("%d (0x%08X)", int32(v), v)
	case int32:
		return fmt.Sprintf("%d (0x%08X)", v, uint32(v))
	case int:
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// formatVariableLocation formats a variable location for display
func formatVariableLocation(loc mc.VariableLocation) string {
	if loc == nil {
		return "<optimized out>"
	}

	switch l := loc.(type) {
	case mc.RegisterLocation:
		return formatRegisterName(l.Register)
	case mc.MemoryLocation:
		baseName := formatRegisterName(l.BaseRegister)
		if l.Offset >= 0 {
			return fmt.Sprintf("[%s+%d]", baseName, l.Offset)
		}
		return fmt.Sprintf("[%s%d]", baseName, l.Offset)
	case mc.ConstantLocation:
		return fmt.Sprintf("const(%d)", l.Value)
	default:
		return "<complex expr>"
	}
}

// Cached register indices for special registers
var (
	regSPIdx   = getRegisterIndex("sp")
	regLRIdx   = getRegisterIndex("lr")
	regPCIdx   = getRegisterIndex("pc")
	regCPSRIdx = getRegisterIndex("cpsr")
)

// formatRegisterName returns the human-readable name for a register index
func formatRegisterName(regIdx uint32) string {
	switch regIdx {
	case regSPIdx:
		return "sp"
	case regLRIdx:
		return "lr"
	case regPCIdx:
		return "pc"
	case regCPSRIdx:
		return "cpsr"
	default:
		if regIdx >= regR0Idx && regIdx <= regR9Idx {
			return fmt.Sprintf("r%d", regIdx-regR0Idx)
		}
		return fmt.Sprintf("reg%d", regIdx)
	}
}

// readVariableValue reads the value of a variable
// Returns nil if the variable location cannot be resolved (optimized out, complex expression, etc.)
func (b *Backend) readVariableValue(v mc.VariableInfo) interface{} {
	// Handle nil location - variable was optimized out or has no location info
	if v.Location == nil {
		return nil // Will display as "<optimized out>" in UI
	}

	state := b.runner.Debugger().State()

	switch loc := v.Location.(type) {
	case mc.RegisterLocation:
		return state.Registers[loc.Register]

	case mc.MemoryLocation:
		var baseValue uint32
		switch loc.BaseRegister {
		case regSPIdx:
			baseValue = *state.SP
		case regLRIdx:
			baseValue = *state.LR
		case regPCIdx:
			baseValue = state.PC
		default:
			if loc.BaseRegister >= regR0Idx && loc.BaseRegister <= regR9Idx {
				baseValue = state.Registers[loc.BaseRegister]
			} else {
				return nil
			}
		}

		addr := uint32(int32(baseValue) + loc.Offset)
		data, err := b.ReadMemory(addr, 4)
		if err != nil {
			return nil
		}
		// Little-endian
		return uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24

	case mc.ConstantLocation:
		return uint32(loc.Value)

	default:
		return nil
	}
}

// GetStackFrames returns the current stack frames by unwinding the call stack.
// The stack is returned with the current frame first (index 0).
func (b *Backend) GetStackFrames() []StackFrame {
	return b.GetCallStack()
}

// GetSourceLines returns source lines for display
func (b *Backend) GetSourceLines(file string, startLine, endLine int) []SourceLine {
	debugInfo := b.runner.DebugInfo()
	if debugInfo == nil {
		return nil
	}

	pc := b.runner.Debugger().State().PC
	currentLoc := debugInfo.GetSourceLocation(pc)
	currentLine := 0
	if currentLoc != nil && currentLoc.File == file {
		currentLine = currentLoc.Line
	}

	result := make([]SourceLine, 0, endLine-startLine+1)
	for line := startLine; line <= endLine; line++ {
		srcText := debugInfo.GetSourceLine(file, line)
		// Stop at end of file (empty lines after current line)
		if srcText == "" && line > currentLine {
			break
		}
		result = append(result, SourceLine{
			LineNumber:    line,
			Text:          srcText,
			IsCurrent:     line == currentLine,
			HasBreakpoint: false, // TODO: implement source-level breakpoint tracking
		})
	}

	return result
}

// GetInstructionText returns the text representation of an instruction at the given address
func (b *Backend) GetInstructionText(addr uint32) string {
	idx, ok := b.addrToIdx[addr]
	if !ok {
		return "???"
	}
	instrs := b.runner.Program().Instructions()
	if idx >= len(instrs) {
		return "???"
	}
	return instrs[idx].Text
}

// GetBreakpointInfos returns breakpoints with display information
func (b *Backend) GetBreakpointInfos() []BreakpointInfo {
	bps := b.runner.Debugger().ListBreakpoints()
	result := make([]BreakpointInfo, len(bps))

	for i, bp := range bps {
		info := BreakpointInfo{
			ID:              bp.ID,
			Address:         bp.Address,
			Enabled:         bp.Enabled,
			HitCount:        bp.HitCount,
			InstructionText: b.GetInstructionText(bp.Address),
		}

		// Get source location if available
		if srcLoc := b.GetSourceLocation(bp.Address); srcLoc != nil {
			info.SourceFile = srcLoc.File
			info.SourceLine = srcLoc.Line
			// Get the source code text
			if debugInfo := b.runner.DebugInfo(); debugInfo != nil {
				info.SourceText = debugInfo.GetSourceLine(srcLoc.File, srcLoc.Line)
			}
		}

		result[i] = info
	}

	return result
}

// GetWatchpointInfos returns watchpoints with display information
func (b *Backend) GetWatchpointInfos() []WatchpointInfo {
	wps := b.runner.Debugger().ListWatchpoints()
	result := make([]WatchpointInfo, len(wps))

	for i, wp := range wps {
		typeStr := "read/write"
		switch wp.Type {
		case interpreter.WatchRead:
			typeStr = "read"
		case interpreter.WatchWrite:
			typeStr = "write"
		}

		result[i] = WatchpointInfo{
			ID:       wp.ID,
			Address:  wp.Address,
			Size:     wp.Size,
			Type:     typeStr,
			Enabled:  wp.Enabled,
			HitCount: wp.HitCount,
		}
	}

	return result
}

// GetMemoryRegions returns memory regions for display and classification
func (b *Backend) GetMemoryRegions() []MemoryRegion {
	var regions []MemoryRegion
	state := b.runner.Debugger().State()
	program := b.runner.Program()

	// Add code region
	if layout := program.MemoryLayout(); layout != nil {
		regions = append(regions, MemoryRegion{
			Name:       "code section",
			StartAddr:  layout.CodeStart,
			EndAddr:    layout.CodeStart + layout.CodeSize,
			RegionType: RegionCode,
		})

		// Add data region if present
		if layout.DataSize > 0 {
			regions = append(regions, MemoryRegion{
				Name:       "data section",
				StartAddr:  layout.DataStart,
				EndAddr:    layout.DataStart + layout.DataSize,
				RegionType: RegionData,
			})
		}

		// Add stack region (from SP to estimated stack top)
		sp := *state.SP
		stackTop := layout.BaseAddress + layout.TotalSize + 0x10000 // 64KB stack space
		if sp > 0 && sp < stackTop {
			regions = append(regions, MemoryRegion{
				Name:       "stack",
				StartAddr:  sp,
				EndAddr:    stackTop,
				RegionType: RegionStack,
			})
		}
	}

	// Add global variables
	for _, g := range program.Globals() {
		if g.Address != nil && g.Size > 0 {
			regions = append(regions, MemoryRegion{
				Name:       "global: " + g.Name,
				StartAddr:  *g.Address,
				EndAddr:    *g.Address + uint32(g.Size),
				RegionType: RegionData, // Globals are in data
			})
		}
	}

	// Add known variables from debug info
	if debugInfo := b.runner.DebugInfo(); debugInfo != nil {
		pc := state.PC
		vars := debugInfo.GetVariables(pc)
		for _, v := range vars {
			if memLoc, ok := v.Location.(mc.MemoryLocation); ok {
				// Calculate actual address
				var baseAddr uint32
				switch {
				case memLoc.BaseRegister == 13:
					baseAddr = *state.SP
				case memLoc.BaseRegister == 14:
					baseAddr = *state.LR
				default:
					if memLoc.BaseRegister >= regR0Idx && memLoc.BaseRegister <= regR9Idx {
						baseAddr = state.Registers[memLoc.BaseRegister]
					}
				}
				varAddr := uint32(int32(baseAddr) + memLoc.Offset)
				if v.Size > 0 {
					regions = append(regions, MemoryRegion{
						Name:       v.Name + " (" + v.TypeName + ")",
						StartAddr:  varAddr,
						EndAddr:    varAddr + uint32(v.Size),
						RegionType: RegionStack, // Stack variables
					})
				}
			}
		}
	}

	return regions
}

// ClassifyAddress returns the memory region containing the given address
func (b *Backend) ClassifyAddress(addr uint32) (MemoryRegion, bool) {
	regions := b.GetMemoryRegions()

	// Priority: stack variables > stack > data > code
	for _, r := range regions {
		if r.RegionType == RegionStack && addr >= r.StartAddr && addr < r.EndAddr {
			return r, true
		}
	}
	for _, r := range regions {
		if r.RegionType == RegionData && addr >= r.StartAddr && addr < r.EndAddr {
			return r, true
		}
	}
	for _, r := range regions {
		if r.RegionType == RegionCode && addr >= r.StartAddr && addr < r.EndAddr {
			return r, true
		}
	}

	return MemoryRegion{RegionType: RegionUnknown}, false
}

// =============================================================================
// Command implementations (DebuggerCommands interface)
// =============================================================================

// CmdStep executes the step command
func (b *Backend) CmdStep(args []string) ExecutionResult {
	count := 1
	if len(args) > 0 {
		if n, err := strconv.Atoi(args[0]); err == nil && n > 0 {
			count = n
		}
	}
	return b.Step(count)
}

// CmdContinue executes the continue command
func (b *Backend) CmdContinue() ExecutionResult {
	return b.Continue()
}

// CmdRun executes the run command
func (b *Backend) CmdRun() ExecutionResult {
	return b.Run()
}

// CmdPrint prints a register or memory value
func (b *Backend) CmdPrint(args []string) PrintResult {
	if len(args) == 0 {
		return PrintResult{Success: false, Error: "Usage: print <register|@address>"}
	}

	what := strings.ToLower(args[0])

	// Memory access with @ prefix
	if strings.HasPrefix(what, "@") {
		addrStr := what[1:]
		addr, err := b.parseAddress(addrStr)
		if err != nil {
			return PrintResult{Success: false, Error: fmt.Sprintf("Invalid address: %s", addrStr)}
		}

		data, err := b.ReadMemory(addr, 4)
		if err != nil || len(data) != 4 {
			return PrintResult{Success: false, Error: fmt.Sprintf("Could not read memory at 0x%08X", addr)}
		}

		val := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
		return PrintResult{
			Success:     true,
			Target:      fmt.Sprintf("@0x%08X", addr),
			Value:       val,
			ValueSigned: int32(val),
			IsMemory:    true,
			Address:     addr,
		}
	}

	// Register access
	val, err := b.ReadRegister(what)
	if err != nil {
		return PrintResult{Success: false, Error: fmt.Sprintf("Unknown register: %s", what)}
	}

	return PrintResult{
		Success:     true,
		Target:      what,
		Value:       val,
		ValueSigned: int32(val),
		IsMemory:    false,
	}
}

// CmdSet sets a register value
func (b *Backend) CmdSet(args []string) SetResult {
	if len(args) < 2 {
		return SetResult{Success: false, Error: "Usage: set <register> <value>"}
	}

	reg := strings.ToLower(args[0])
	val, err := b.parseValue(args[1])
	if err != nil {
		return SetResult{Success: false, Error: fmt.Sprintf("Invalid value: %s", args[1])}
	}

	if err := b.WriteRegister(reg, val); err != nil {
		return SetResult{Success: false, Error: fmt.Sprintf("Unknown register: %s", reg)}
	}

	return SetResult{
		Success:     true,
		Register:    reg,
		Value:       val,
		ValueSigned: int32(val),
	}
}

// CmdBreak adds a breakpoint
func (b *Backend) CmdBreak(args []string) BreakpointResult {
	if len(args) == 0 {
		return BreakpointResult{Success: false, Error: "Usage: break <address|symbol>"}
	}

	// Try to resolve as symbol first, then as address
	addr, err := b.resolveAddressOrSymbol(args[0])
	if err != nil {
		return BreakpointResult{Success: false, Error: err.Error()}
	}

	bp, err := b.AddBreakpoint(addr)
	if err != nil {
		return BreakpointResult{Success: false, Error: err.Error()}
	}

	return BreakpointResult{
		Success:   true,
		Operation: "add",
		ID:        bp.ID,
		Address:   addr,
	}
}

// CmdWatch adds a watchpoint
func (b *Backend) CmdWatch(args []string) WatchpointResult {
	if len(args) == 0 {
		return WatchpointResult{Success: false, Error: "Usage: watch <address|symbol>"}
	}

	addr, err := b.resolveAddressOrSymbol(args[0])
	if err != nil {
		return WatchpointResult{Success: false, Error: err.Error()}
	}

	wp, err := b.AddWatchpoint(addr)
	if err != nil {
		return WatchpointResult{Success: false, Error: err.Error()}
	}

	return WatchpointResult{
		Success:   true,
		Operation: "add",
		ID:        wp.ID,
		Address:   addr,
		Size:      wp.Size,
	}
}

// CmdDelete deletes a breakpoint or watchpoint
func (b *Backend) CmdDelete(args []string) DeleteResult {
	if len(args) == 0 {
		return DeleteResult{Success: false, Error: "Usage: delete <id>"}
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return DeleteResult{Success: false, Error: fmt.Sprintf("Invalid ID: %s", args[0])}
	}

	// Try breakpoint first, then watchpoint
	if b.RemoveBreakpoint(id) == nil {
		return DeleteResult{Success: true, WasBreakpoint: true, ID: id}
	}
	if b.RemoveWatchpoint(id) == nil {
		return DeleteResult{Success: true, WasWatchpoint: true, ID: id}
	}

	return DeleteResult{Success: false, Error: fmt.Sprintf("No breakpoint or watchpoint with ID %d", id)}
}

// CmdDisasm disassembles instructions
func (b *Backend) CmdDisasm(args []string) DisassemblyResult {
	addr := b.GetState().PC
	count := 10

	if len(args) > 0 {
		if a, err := b.parseAddress(args[0]); err == nil {
			addr = a
		}
	}
	if len(args) > 1 {
		if n, err := strconv.Atoi(args[1]); err == nil && n > 0 {
			count = n
		}
	}

	instrs, err := b.Disassemble(addr, count)
	if err != nil {
		return DisassemblyResult{Success: false, Error: fmt.Sprintf("No instruction at 0x%08X", addr)}
	}

	return DisassemblyResult{
		Success:      true,
		Address:      addr,
		Instructions: instrs,
	}
}

// CmdMemory displays memory
func (b *Backend) CmdMemory(args []string) MemoryResult {
	if len(args) == 0 {
		return MemoryResult{Success: false, Error: "Usage: memory <address> [count]"}
	}

	// Parse address expression and optional count
	fullArg := strings.Join(args, " ")
	expr := fullArg
	count := uint32(64)

	// Check for comma separator for count
	if commaIdx := strings.LastIndex(fullArg, ","); commaIdx != -1 {
		countExpr := strings.TrimSpace(fullArg[commaIdx+1:])
		if countExpr != "" {
			eval := NewExpressionEvaluator(b)
			if n, err := eval.Eval(countExpr); err == nil && n > 0 {
				count = n
				expr = strings.TrimSpace(fullArg[:commaIdx])
			} else if err != nil {
				return MemoryResult{Success: false, Error: fmt.Sprintf("Invalid count expression: %v", err)}
			}
		}
	}

	// Evaluate address expression
	eval := NewExpressionEvaluator(b)
	addr, err := eval.Eval(expr)
	if err != nil {
		return MemoryResult{Success: false, Error: fmt.Sprintf("Invalid address expression: %v", err)}
	}

	data, err := b.ReadMemory(addr, int(count))
	if err != nil || len(data) == 0 {
		return MemoryResult{Success: false, Error: fmt.Sprintf("Could not read memory at 0x%08X", addr)}
	}

	return MemoryResult{
		Success: true,
		Address: addr,
		Data:    data,
		Regions: b.GetMemoryRegions(),
	}
}

// CmdSource displays source code
func (b *Backend) CmdSource(args []string) SourceResult {
	pc := b.GetState().PC
	loc := b.GetSourceLocation(pc)
	if loc == nil || !loc.IsValid() {
		return SourceResult{Success: false, Error: "No source location for current address"}
	}

	contextLines := 5
	if len(args) > 0 {
		if n, err := strconv.Atoi(args[0]); err == nil && n > 0 {
			contextLines = n / 2
		}
	}

	startLine := loc.Line - contextLines
	if startLine < 1 {
		startLine = 1
	}
	endLine := loc.Line + contextLines

	lines := b.GetSourceLines(loc.File, startLine, endLine)
	if len(lines) == 0 {
		return SourceResult{Success: false, Error: "No source available"}
	}

	return SourceResult{
		Success: true,
		File:    loc.File,
		Lines:   lines,
	}
}

// CmdEval evaluates an expression
func (b *Backend) CmdEval(args []string) EvalResult {
	if len(args) == 0 {
		return EvalResult{Success: false, Error: "Usage: eval <expression>"}
	}

	expr := strings.Join(args, " ")
	eval := NewExpressionEvaluator(b)
	result, err := eval.Eval(expr)
	if err != nil {
		return EvalResult{Success: false, Error: err.Error(), Expression: expr}
	}

	return EvalResult{
		Success:     true,
		Expression:  expr,
		Value:       result,
		ValueSigned: int32(result),
		ValueBinary: FormatBinary(result),
	}
}

// CmdInfo returns the current CPU state
func (b *Backend) CmdInfo() DebuggerState {
	return b.GetState()
}

// CmdRegisters returns all register values
func (b *Backend) CmdRegisters() []RegisterInfo {
	return b.GetState().Registers
}

// CmdStack returns stack information
func (b *Backend) CmdStack() (uint32, []byte, []StackFrame) {
	state := b.GetState()
	sp := state.SP

	// Read 10 stack entries (40 bytes)
	data, _ := b.ReadMemory(sp, 40)
	frames := b.GetStackFrames()

	return sp, data, frames
}

// CmdVars returns accessible variables
func (b *Backend) CmdVars() []VariableValue {
	return b.GetVariables(b.GetState().PC)
}

// CmdList returns breakpoints and watchpoints
func (b *Backend) CmdList() ([]BreakpointInfo, []WatchpointInfo) {
	return b.GetBreakpointInfos(), b.GetWatchpointInfos()
}

// GetCurrentInstruction returns the current instruction for display
func (b *Backend) GetCurrentInstruction() CurrentInstructionResult {
	state := b.GetState()
	pc := state.PC

	// Read instruction word
	data, _ := b.ReadMemory(pc, 4)
	word := uint32(0)
	if len(data) == 4 {
		word = uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
	}

	result := CurrentInstructionResult{
		PC:              pc,
		InstructionWord: word,
		InstructionText: b.GetInstructionText(pc),
	}

	// Get source location if available
	if loc := b.GetSourceLocation(pc); loc != nil && loc.IsValid() {
		result.HasSource = true
		result.SourceFile = loc.File
		result.SourceLine = loc.Line
		if debugInfo := b.DebugInfo(); debugInfo != nil {
			result.SourceText = debugInfo.GetSourceLine(loc.File, loc.Line)
		}
	}

	return result
}

// Helper: parse an address string (hex or decimal)
func (b *Backend) parseAddress(s string) (uint32, error) {
	s = strings.TrimPrefix(strings.ToLower(s), "0x")
	// Try hex
	if val, err := strconv.ParseUint(s, 16, 32); err == nil {
		return uint32(val), nil
	}
	// Try decimal
	if val, err := strconv.ParseUint(s, 10, 32); err == nil {
		return uint32(val), nil
	}
	return 0, fmt.Errorf("invalid address: %s", s)
}

// Helper: parse a value string (hex, decimal, or negative)
func (b *Backend) parseValue(s string) (uint32, error) {
	s = strings.TrimSpace(s)

	// Handle negative numbers
	if strings.HasPrefix(s, "-") {
		if val, err := strconv.ParseInt(s, 10, 32); err == nil {
			return uint32(val), nil
		}
	}

	// Handle hex
	if strings.HasPrefix(strings.ToLower(s), "0x") {
		if val, err := strconv.ParseUint(s[2:], 16, 32); err == nil {
			return uint32(val), nil
		}
	}

	// Handle decimal
	if val, err := strconv.ParseUint(s, 10, 32); err == nil {
		return uint32(val), nil
	}

	return 0, fmt.Errorf("invalid value: %s", s)
}

// Helper: resolve an address or symbol name
func (b *Backend) resolveAddressOrSymbol(s string) (uint32, error) {
	// Try as address first
	if addr, err := b.parseAddress(s); err == nil {
		return addr, nil
	}

	// Try as symbol
	if addr, err := b.ResolveSymbol(s); err == nil {
		return addr, nil
	}

	return 0, fmt.Errorf("invalid address or unknown symbol: %s", s)
}

// GetFunctionAtAddress returns the function containing the given address
func (b *Backend) GetFunctionAtAddress(addr uint32) (*mc.FunctionDebugInfo, bool) {
	debugInfo := b.runner.DebugInfo()
	if debugInfo == nil {
		return nil, false
	}

	for _, fn := range debugInfo.Functions {
		if addr >= fn.StartAddress && addr < fn.EndAddress {
			return fn, true
		}
	}

	// Fallback: check program functions
	program := b.runner.Program()
	if program == nil {
		return nil, false
	}

	for name, fn := range program.Functions() {
		if len(fn.InstructionRanges) > 0 {
			startIdx := fn.InstructionRanges[0].Start
			// Calculate end index from last range: Start + Count - 1 + 1 = Start + Count
			lastRange := fn.InstructionRanges[len(fn.InstructionRanges)-1]
			endIdx := lastRange.Start + lastRange.Count
			if startAddr, ok := b.idxToAddr[startIdx]; ok {
				endAddr := startAddr
				if ea, ok := b.idxToAddr[endIdx]; ok {
					endAddr = ea
				} else if endIdx > 0 {
					// Try the last instruction index
					if ea, ok := b.idxToAddr[endIdx-1]; ok {
						endAddr = ea + 4 // Include the last instruction
					}
				}
				if addr >= startAddr && addr < endAddr {
					// Create a minimal FunctionDebugInfo
					return &mc.FunctionDebugInfo{
						Name:         name,
						StartAddress: startAddr,
						EndAddress:   endAddr,
						SourceFile:   fn.SourceFile,
						StartLine:    fn.StartLine,
						EndLine:      fn.EndLine,
					}, true
				}
			}
		}
	}

	return nil, false
}

// GetCallStack returns the current call stack by unwinding frames.
// The stack is returned with the current frame first (index 0).
func (b *Backend) GetCallStack() []StackFrame {
	state := b.runner.Debugger().State()
	frames := make([]StackFrame, 0, 16)

	// Frame 0: Current PC
	pc := state.PC
	lr := *state.LR
	sp := *state.SP

	// Add current frame
	frame := b.buildStackFrame(pc)
	frames = append(frames, frame)

	// Memory size for bounds checking
	memSize := uint32(len(state.Memory))

	// Termination address (return from main)
	terminationAddr := interpreter.TerminationAddress

	// Unwind stack by following return addresses
	// We look for the LR value which points to the return address
	// In a typical calling convention:
	// - LR contains the return address for the current function
	// - When a function calls another, it saves LR on the stack

	// First, check if LR points to a valid code address (not terminated yet)
	if lr != terminationAddr && lr != 0 && lr >= 0x10000 && lr < memSize {
		frame := b.buildStackFrame(lr)
		if frame.Function != "" || frame.File != "" {
			frames = append(frames, frame)
		}
	}

	// Try to unwind further by scanning the stack for return addresses
	// This is a heuristic approach since we don't have frame pointers
	currentSP := sp
	seenAddrs := make(map[uint32]bool)
	seenAddrs[pc] = true
	if lr != 0 {
		seenAddrs[lr] = true
	}

	// Scan stack entries looking for potential return addresses
	maxFrames := 20
	maxStackScan := uint32(256) // Scan up to 256 bytes of stack

	for i := uint32(0); i < maxStackScan && len(frames) < maxFrames; i += 4 {
		addr := currentSP + i
		if addr+4 > memSize {
			break
		}

		// Read potential return address from stack
		potentialRA := uint32(state.Memory[addr]) |
			uint32(state.Memory[addr+1])<<8 |
			uint32(state.Memory[addr+2])<<16 |
			uint32(state.Memory[addr+3])<<24

		// Check if this looks like a valid return address
		if potentialRA == 0 || potentialRA == terminationAddr {
			continue
		}
		if potentialRA < 0x10000 || potentialRA >= memSize {
			continue
		}
		if seenAddrs[potentialRA] {
			continue
		}

		// Check if this address is within a known function
		if fn, ok := b.GetFunctionAtAddress(potentialRA); ok {
			seenAddrs[potentialRA] = true
			frame := StackFrame{
				Address:  potentialRA,
				Function: fn.Name,
				File:     fn.SourceFile,
				Line:     fn.StartLine,
			}
			// Get more precise line info if available
			if srcLoc := b.GetSourceLocation(potentialRA); srcLoc != nil {
				frame.File = srcLoc.File
				frame.Line = srcLoc.Line
			}
			frames = append(frames, frame)
		}
	}

	return frames
}

// buildStackFrame creates a StackFrame for the given address
func (b *Backend) buildStackFrame(addr uint32) StackFrame {
	frame := StackFrame{
		Address: addr,
	}

	// Try to get function name
	if fn, ok := b.GetFunctionAtAddress(addr); ok {
		frame.Function = fn.Name
		frame.File = fn.SourceFile
		frame.Line = fn.StartLine
	} else if sym, ok := b.GetSymbolAt(addr); ok {
		frame.Function = sym
	}

	// Try to get source location
	if srcLoc := b.GetSourceLocation(addr); srcLoc != nil {
		frame.File = srcLoc.File
		frame.Line = srcLoc.Line
	}

	return frame
}
