package cpu

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/interpreter"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/llvm"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/registers"
	"github.com/spf13/cobra"
)

var (
	debugMemorySize uint32
	debugVerbose    bool
)

var debugCmd = &cobra.Command{
	Use:   "debug <file>",
	Short: "Interactive debugger for cucaracha programs",
	Long: `Starts an interactive debugging session for a cucaracha program.

The command accepts either:
  - Assembly files (.cucaracha) - parsed by the LLVM assembly parser
  - Binary/object files (.o) - parsed by the ELF binary parser

Available debugger commands:
  step, s [n]        - Step n instructions (default: 1)
  continue, c        - Continue execution until breakpoint
  run, r             - Run until termination or breakpoint
  break, b <addr>    - Set breakpoint at address (hex)
  watch, w <addr>    - Set watchpoint on memory address
  delete, d <id>     - Delete breakpoint/watchpoint by ID
  list, l            - List all breakpoints and watchpoints
  print, p <what>    - Print register (r0-r9, sp, lr, pc) or memory (@addr)
  set <reg> <value>  - Set register value
  disasm, x [addr] [n] - Disassemble n instructions at addr (default: PC, 10)
  info, i            - Show CPU state (registers, flags)
  stack              - Show stack contents
  memory, m <addr> [n] - Show n bytes of memory at addr
  help, h            - Show this help
  quit, q            - Exit debugger

Example:
  cucaracha cpu debug program.cucaracha`,
	Args: cobra.ExactArgs(1),
	Run:  runDebug,
}

func init() {
	CpuCmd.AddCommand(debugCmd)
	debugCmd.Flags().Uint32VarP(&debugMemorySize, "memory", "m", 0x20000, "Memory size in bytes (default: 128KB)")
	debugCmd.Flags().BoolVarP(&debugVerbose, "verbose", "v", false, "Print verbose output")
}

// debugSession holds the state of an interactive debugging session
type debugSession struct {
	dbg             *interpreter.Debugger
	pf              mc.ProgramFile
	addrToIdx       map[uint32]int
	idxToAddr       map[int]uint32
	running         bool
	lastCmd         string
	terminationAddr uint32
}

func runDebug(cmd *cobra.Command, args []string) {
	inputPath := args[0]
	ext := strings.ToLower(filepath.Ext(inputPath))

	var pf mc.ProgramFile
	var err error

	switch ext {
	case ".cucaracha", ".s":
		fmt.Printf("Loading assembly file: %s\n", inputPath)
		pf, err = llvm.ParseAssemblyFile(inputPath)
	case ".o":
		fmt.Printf("Loading binary file: %s\n", inputPath)
		pf, err = llvm.ParseBinaryFile(inputPath)
	default:
		fmt.Fprintf(os.Stderr, "Error: unsupported file extension '%s'\n", ext)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading program: %v\n", err)
		os.Exit(2)
	}

	// Resolve the program
	memConfig := mc.MemoryResolverConfig{
		BaseAddress:     0x10000,
		MaxSize:         0,
		DataAlignment:   4,
		InstructionSize: 4,
	}
	resolved, err := mc.Resolve(pf, memConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving program: %v\n", err)
		os.Exit(3)
	}

	// Create interpreter and debugger
	interp := interpreter.NewInterpreter(debugMemorySize)
	dbg := interpreter.NewDebugger(interp)

	if err := loadProgramFile(interp, resolved); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading program: %v\n", err)
		os.Exit(4)
	}

	// Set up termination address
	const TerminationAddress uint32 = 0x0000FFFC
	dbg.AddTerminationAddress(TerminationAddress)

	// Find main and set entry point
	mainFunc, hasMain := resolved.Functions()["main"]
	if hasMain && len(mainFunc.InstructionRanges) > 0 {
		startIdx := mainFunc.InstructionRanges[0].Start
		instrs := resolved.Instructions()
		if startIdx < len(instrs) && instrs[startIdx].Address != nil {
			interp.State().PC = *instrs[startIdx].Address
			*interp.State().LR = TerminationAddress
		}
	}

	// Build address maps
	addrToIdx := make(map[uint32]int)
	idxToAddr := make(map[int]uint32)
	for i, instr := range resolved.Instructions() {
		if instr.Address != nil {
			addrToIdx[*instr.Address] = i
			idxToAddr[i] = *instr.Address
		}
	}

	session := &debugSession{
		dbg:             dbg,
		pf:              resolved,
		addrToIdx:       addrToIdx,
		idxToAddr:       idxToAddr,
		running:         true,
		terminationAddr: TerminationAddress,
	}

	fmt.Printf("Loaded %d instructions\n", len(resolved.Instructions()))
	fmt.Printf("Entry point: 0x%08X\n", interp.State().PC)
	fmt.Printf("Type 'help' for available commands.\n\n")

	// Show initial state
	session.showCurrentInstruction()

	// Start interactive loop
	reader := bufio.NewReader(os.Stdin)
	for session.running {
		fmt.Print("(cucaracha) ")
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			// Repeat last command
			line = session.lastCmd
		}
		if line != "" {
			session.lastCmd = line
			session.executeCommand(line)
		}
	}
}

func (s *debugSession) executeCommand(line string) {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return
	}

	cmd := strings.ToLower(parts[0])
	args := parts[1:]

	switch cmd {
	case "step", "s":
		s.cmdStep(args)
	case "continue", "c":
		s.cmdContinue()
	case "run", "r":
		s.cmdRun()
	case "break", "b":
		s.cmdBreak(args)
	case "watch", "w":
		s.cmdWatch(args)
	case "delete", "d":
		s.cmdDelete(args)
	case "list", "l":
		s.cmdList()
	case "print", "p":
		s.cmdPrint(args)
	case "set":
		s.cmdSet(args)
	case "disasm", "x":
		s.cmdDisasm(args)
	case "info", "i":
		s.cmdInfo()
	case "stack":
		s.cmdStack()
	case "memory", "m":
		s.cmdMemory(args)
	case "help", "h", "?":
		s.cmdHelp()
	case "quit", "q", "exit":
		s.running = false
		fmt.Println("Exiting debugger.")
	default:
		fmt.Printf("Unknown command: %s. Type 'help' for available commands.\n", cmd)
	}
}

func (s *debugSession) cmdStep(args []string) {
	n := 1
	if len(args) > 0 {
		if val, err := strconv.Atoi(args[0]); err == nil && val > 0 {
			n = val
		}
	}

	for i := 0; i < n; i++ {
		result := s.dbg.Step()
		if result.Error != nil {
			fmt.Printf("Error: %v\n", result.Error)
			return
		}
		if result.StopReason == interpreter.StopTermination {
			fmt.Println("Program terminated.")
			s.showReturnValue()
			return
		}
		if result.StopReason == interpreter.StopHalt {
			fmt.Println("CPU halted.")
			return
		}
		if result.StopReason == interpreter.StopBreakpoint {
			fmt.Printf("Breakpoint hit at 0x%08X\n", s.dbg.State().PC)
			break
		}
		if result.StopReason == interpreter.StopWatchpoint {
			fmt.Printf("Watchpoint triggered at 0x%08X\n", s.dbg.State().PC)
			break
		}
	}
	s.showCurrentInstruction()
}

func (s *debugSession) cmdContinue() {
	result := s.dbg.Continue()
	s.handleExecutionResult(result)
}

func (s *debugSession) cmdRun() {
	result := s.dbg.Run(0)
	s.handleExecutionResult(result)
}

func (s *debugSession) handleExecutionResult(result *interpreter.ExecutionResult) {
	switch result.StopReason {
	case interpreter.StopBreakpoint:
		fmt.Printf("Breakpoint hit at 0x%08X\n", s.dbg.State().PC)
		s.showCurrentInstruction()
	case interpreter.StopWatchpoint:
		fmt.Printf("Watchpoint triggered at 0x%08X\n", s.dbg.State().PC)
		s.showCurrentInstruction()
	case interpreter.StopTermination:
		fmt.Printf("Program terminated after %d steps.\n", result.StepsExecuted)
		s.showReturnValue()
	case interpreter.StopHalt:
		fmt.Println("CPU halted.")
	case interpreter.StopError:
		fmt.Printf("Error: %v\n", result.Error)
	default:
		fmt.Printf("Stopped: %s after %d steps\n", result.StopReason, result.StepsExecuted)
		s.showCurrentInstruction()
	}
}

func (s *debugSession) cmdBreak(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: break <address>")
		fmt.Println("  Address can be hex (0x...) or decimal")
		return
	}

	addr, err := parseAddress(args[0])
	if err != nil {
		fmt.Printf("Invalid address: %s\n", args[0])
		return
	}

	bp := s.dbg.AddBreakpoint(addr)
	fmt.Printf("Breakpoint %d set at 0x%08X\n", bp.ID, addr)
}

func (s *debugSession) cmdWatch(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: watch <address>")
		return
	}

	addr, err := parseAddress(args[0])
	if err != nil {
		fmt.Printf("Invalid address: %s\n", args[0])
		return
	}

	wp := s.dbg.AddWatchpoint(addr, 4, interpreter.WatchWrite)
	fmt.Printf("Watchpoint %d set at 0x%08X (4 bytes, write)\n", wp.ID, addr)
}

func (s *debugSession) cmdDelete(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: delete <breakpoint-id>")
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("Invalid ID: %s\n", args[0])
		return
	}

	if s.dbg.RemoveBreakpoint(id) {
		fmt.Printf("Breakpoint %d deleted.\n", id)
	} else if s.dbg.RemoveWatchpoint(id) {
		fmt.Printf("Watchpoint %d deleted.\n", id)
	} else {
		fmt.Printf("No breakpoint or watchpoint with ID %d.\n", id)
	}
}

func (s *debugSession) cmdList() {
	bps := s.dbg.ListBreakpoints()
	wps := s.dbg.ListWatchpoints()

	if len(bps) == 0 && len(wps) == 0 {
		fmt.Println("No breakpoints or watchpoints set.")
		return
	}

	if len(bps) > 0 {
		fmt.Println("Breakpoints:")
		for _, bp := range bps {
			status := "enabled"
			if !bp.Enabled {
				status = "disabled"
			}
			instrText := s.getInstructionText(bp.Address)
			fmt.Printf("  %d: 0x%08X (%s) %s\n", bp.ID, bp.Address, status, instrText)
		}
	}

	if len(wps) > 0 {
		fmt.Println("Watchpoints:")
		for _, wp := range wps {
			status := "enabled"
			if !wp.Enabled {
				status = "disabled"
			}
			typeStr := "read/write"
			switch wp.Type {
			case interpreter.WatchRead:
				typeStr = "read"
			case interpreter.WatchWrite:
				typeStr = "write"
			}
			fmt.Printf("  %d: 0x%08X (%d bytes, %s, %s)\n", wp.ID, wp.Address, wp.Size, typeStr, status)
		}
	}
}

func (s *debugSession) cmdPrint(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: print <register|@address>")
		fmt.Println("  Registers: r0-r9, sp, lr, pc, cpsr")
		fmt.Println("  Memory: @0x1234 or @1234")
		return
	}

	what := strings.ToLower(args[0])

	// Memory access
	if strings.HasPrefix(what, "@") {
		addr, err := parseAddress(what[1:])
		if err != nil {
			fmt.Printf("Invalid address: %s\n", what[1:])
			return
		}
		data, err := s.dbg.ReadMemory(addr, 4)
		if err != nil || len(data) != 4 {
			fmt.Printf("Could not read memory at 0x%08X\n", addr)
			return
		}
		val := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
		fmt.Printf("[@0x%08X] = %d (0x%08X)\n", addr, int32(val), val)
		return
	}

	// Special registers
	switch what {
	case "pc":
		val := s.dbg.GetPC()
		fmt.Printf("pc = %d (0x%08X)\n", val, val)
		return
	case "sp":
		val := s.dbg.GetSP()
		fmt.Printf("sp = %d (0x%08X)\n", val, val)
		return
	case "lr":
		val := s.dbg.GetLR()
		fmt.Printf("lr = %d (0x%08X)\n", val, val)
		return
	}

	// General register access by name
	regIdx, ok := parseRegisterName(what)
	if !ok {
		fmt.Printf("Unknown register: %s\n", what)
		return
	}
	val := s.dbg.GetRegister(regIdx)
	fmt.Printf("%s = %d (0x%08X)\n", what, int32(val), val)
}

func (s *debugSession) cmdSet(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: set <register> <value>")
		return
	}

	reg := strings.ToLower(args[0])
	val, err := parseValue(args[1])
	if err != nil {
		fmt.Printf("Invalid value: %s\n", args[1])
		return
	}

	// Special registers
	switch reg {
	case "pc":
		s.dbg.SetPC(val)
		fmt.Printf("pc = %d (0x%08X)\n", val, val)
		return
	case "sp":
		s.dbg.SetSP(val)
		fmt.Printf("sp = %d (0x%08X)\n", val, val)
		return
	case "lr":
		s.dbg.SetLR(val)
		fmt.Printf("lr = %d (0x%08X)\n", val, val)
		return
	}

	// General register by name
	regIdx, ok := parseRegisterName(reg)
	if !ok {
		fmt.Printf("Unknown register: %s\n", reg)
		return
	}
	s.dbg.SetRegister(regIdx, val)
	fmt.Printf("%s = %d (0x%08X)\n", reg, int32(val), val)
}

func (s *debugSession) cmdDisasm(args []string) {
	addr := s.dbg.State().PC
	count := 10

	if len(args) > 0 {
		if a, err := parseAddress(args[0]); err == nil {
			addr = a
		}
	}
	if len(args) > 1 {
		if n, err := strconv.Atoi(args[1]); err == nil && n > 0 {
			count = n
		}
	}

	instrs := s.pf.Instructions()
	idx, ok := s.addrToIdx[addr]
	if !ok {
		// Try to find closest instruction
		fmt.Printf("No instruction at 0x%08X\n", addr)
		return
	}

	fmt.Printf("Disassembly at 0x%08X:\n", addr)
	for i := 0; i < count && idx+i < len(instrs); i++ {
		instr := instrs[idx+i]
		if instr.Address == nil {
			continue
		}
		marker := "  "
		if *instr.Address == s.dbg.State().PC {
			marker = "=>"
		}
		// Check for breakpoint
		for _, bp := range s.dbg.ListBreakpoints() {
			if bp.Address == *instr.Address && bp.Enabled {
				marker = "* "
				if *instr.Address == s.dbg.State().PC {
					marker = "*>"
				}
				break
			}
		}
		fmt.Printf("%s 0x%08X: %s\n", marker, *instr.Address, instr.Text)
	}
}

func (s *debugSession) cmdInfo() {
	state := s.dbg.State()

	fmt.Println("=== CPU State ===")
	fmt.Printf("PC:   0x%08X\n", state.PC)
	fmt.Printf("SP:   %d (0x%08X)\n", *state.SP, *state.SP)
	fmt.Printf("LR:   0x%08X\n", *state.LR)

	// Get CPSR from the cpsr register
	cpsrIdx := uint32(registers.Register("cpsr").Encode())
	cpsr := state.Registers[cpsrIdx]
	fmt.Printf("CPSR: 0x%08X (N=%d Z=%d C=%d V=%d)\n",
		cpsr,
		(cpsr>>3)&1, // FLAG_N is bit 3
		(cpsr>>0)&1, // FLAG_Z is bit 0
		(cpsr>>1)&1, // FLAG_C is bit 1
		(cpsr>>2)&1) // FLAG_V is bit 2
	fmt.Println()

	fmt.Println("General Purpose Registers:")
	for i := 0; i < 10; i++ {
		regIdx := 16 + i // r0 starts at index 16
		val := state.Registers[regIdx]
		fmt.Printf("  r%d = %10d (0x%08X)\n", i, int32(val), val)
	}
}

func (s *debugSession) cmdStack() {
	state := s.dbg.State()
	sp := *state.SP

	fmt.Printf("Stack (SP = 0x%08X):\n", sp)
	// Show 10 stack entries
	for i := 0; i < 10; i++ {
		addr := sp + uint32(i*4)
		if int(addr)+4 > len(state.Memory) {
			break
		}
		data, err := s.dbg.ReadMemory(addr, 4)
		if err != nil || len(data) != 4 {
			break
		}
		val := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
		marker := ""
		if i == 0 {
			marker = " <- SP"
		}
		fmt.Printf("  0x%08X: %10d (0x%08X)%s\n", addr, int32(val), val, marker)
	}
}

func (s *debugSession) cmdMemory(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: memory <address> [count]")
		return
	}

	addr, err := parseAddress(args[0])
	if err != nil {
		fmt.Printf("Invalid address: %s\n", args[0])
		return
	}

	count := 64
	if len(args) > 1 {
		if n, err := strconv.Atoi(args[1]); err == nil && n > 0 {
			count = n
		}
	}

	data, err := s.dbg.ReadMemory(addr, count)
	if err != nil || len(data) == 0 {
		fmt.Printf("Could not read memory at 0x%08X\n", addr)
		return
	}

	fmt.Printf("Memory at 0x%08X:\n", addr)
	for i := 0; i < len(data); i += 16 {
		fmt.Printf("  0x%08X: ", addr+uint32(i))
		// Hex bytes
		for j := 0; j < 16 && i+j < len(data); j++ {
			fmt.Printf("%02X ", data[i+j])
		}
		// Padding if needed
		for j := len(data) - i; j < 16; j++ {
			fmt.Print("   ")
		}
		// ASCII
		fmt.Print(" |")
		for j := 0; j < 16 && i+j < len(data); j++ {
			b := data[i+j]
			if b >= 32 && b < 127 {
				fmt.Printf("%c", b)
			} else {
				fmt.Print(".")
			}
		}
		fmt.Println("|")
	}
}

func (s *debugSession) cmdHelp() {
	fmt.Println(`Cucaracha Debugger Commands:

Execution:
  step, s [n]        - Step n instructions (default: 1)
  continue, c        - Continue execution until breakpoint
  run, r             - Run until termination or breakpoint

Breakpoints:
  break, b <addr>    - Set breakpoint at address (hex: 0x... or decimal)
  watch, w <addr>    - Set watchpoint on memory address
  delete, d <id>     - Delete breakpoint/watchpoint by ID
  list, l            - List all breakpoints and watchpoints

Inspection:
  print, p <what>    - Print register (r0-r9, sp, lr, pc) or memory (@addr)
  set <reg> <value>  - Set register value
  disasm, x [addr] [n] - Disassemble n instructions at addr
  info, i            - Show CPU state (registers, flags)
  stack              - Show stack contents
  memory, m <addr> [n] - Show n bytes of memory at addr

Other:
  help, h            - Show this help
  quit, q            - Exit debugger

Press Enter to repeat the last command.`)
}

func (s *debugSession) showCurrentInstruction() {
	state := s.dbg.State()
	pc := state.PC

	instrText := s.getInstructionText(pc)
	word, _ := state.ReadMemory32(pc)

	fmt.Printf("=> 0x%08X [%08X]: %s\n", pc, word, instrText)
}

func (s *debugSession) showReturnValue() {
	state := s.dbg.State()
	r0 := state.Registers[16]
	fmt.Printf("Return value (r0): %d (0x%08X)\n", int32(r0), r0)
}

func (s *debugSession) getInstructionText(addr uint32) string {
	idx, ok := s.addrToIdx[addr]
	if !ok {
		return "???"
	}
	instrs := s.pf.Instructions()
	if idx >= len(instrs) {
		return "???"
	}
	return instrs[idx].Text
}

func parseAddress(s string) (uint32, error) {
	s = strings.TrimPrefix(strings.ToLower(s), "0x")
	if strings.HasPrefix(s, "0x") {
		s = s[2:]
	}
	// Try hex first if it looks like hex
	if val, err := strconv.ParseUint(s, 16, 32); err == nil {
		return uint32(val), nil
	}
	// Try decimal
	if val, err := strconv.ParseUint(s, 10, 32); err == nil {
		return uint32(val), nil
	}
	return 0, fmt.Errorf("invalid address: %s", s)
}

func parseValue(s string) (uint32, error) {
	s = strings.ToLower(s)
	negative := false
	if strings.HasPrefix(s, "-") {
		negative = true
		s = s[1:]
	}

	var val uint64
	var err error

	if strings.HasPrefix(s, "0x") {
		val, err = strconv.ParseUint(s[2:], 16, 32)
	} else {
		val, err = strconv.ParseUint(s, 10, 32)
	}

	if err != nil {
		return 0, err
	}

	if negative {
		return uint32(-int32(val)), nil
	}
	return uint32(val), nil
}

// parseRegisterName converts a register name to its encoded index
func parseRegisterName(name string) (uint32, bool) {
	name = strings.ToLower(name)

	// Handle numbered registers r0-r9
	if strings.HasPrefix(name, "r") && len(name) >= 2 {
		numStr := name[1:]
		if num, err := strconv.Atoi(numStr); err == nil && num >= 0 && num <= 9 {
			// r0-r9 map to indices 16-25
			return uint32(16 + num), true
		}
	}

	// Use the registers package for named registers
	reg := registers.Register(name)
	if reg.Name() != "" {
		return uint32(reg.Encode()), true
	}

	return 0, false
}
