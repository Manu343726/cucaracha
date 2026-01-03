package cpu

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/interpreter"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/llvm"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/registers"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Color definitions for debugger output
var (
	// Address colors
	colorAddr = color.New(color.FgCyan)
	// Instruction/opcode colors
	colorInstr = color.New(color.FgYellow)
	// Register name colors
	colorReg = color.New(color.FgGreen)
	// Value colors (numeric values)
	colorValue = color.New(color.FgWhite, color.Bold)
	// Hex value colors
	colorHex = color.New(color.FgMagenta)
	// Prompt/marker colors
	colorPrompt = color.New(color.FgBlue, color.Bold)
	// Error colors
	colorError = color.New(color.FgRed, color.Bold)
	// Success/info colors
	colorSuccess = color.New(color.FgGreen)
	// Warning colors
	colorWarning = color.New(color.FgYellow)
	// Header/title colors
	colorHeader = color.New(color.FgWhite, color.Bold, color.Underline)
	// Breakpoint marker
	colorBreakpoint = color.New(color.FgRed, color.Bold)
	// Current PC marker
	colorPC = color.New(color.FgGreen, color.Bold)
	// Flag colors
	colorFlagSet   = color.New(color.FgGreen, color.Bold)
	colorFlagClear = color.New(color.FgHiBlack)
	// Source code colors
	colorSource     = color.New(color.FgHiWhite)
	colorSourceFile = color.New(color.FgHiBlue)
	colorSourceLine = color.New(color.FgHiCyan)
	colorVarName    = color.New(color.FgHiGreen)
	colorVarType    = color.New(color.FgHiYellow)
)

// Instruction part colors
var (
	instrOpcode = color.New(color.FgYellow, color.Bold)
	instrReg    = color.New(color.FgGreen)
	instrImm    = color.New(color.FgCyan)
	instrPunct  = color.New(color.FgWhite)
)

// Regex patterns for parsing instruction parts
var (
	debugRegPattern    = regexp.MustCompile(`\b(r[0-9]{1,2}|sp|lr|pc|cpsr)\b`)
	debugImmPattern    = regexp.MustCompile(`#-?[0-9]+|#-?0x[0-9a-fA-F]+|\b-?[0-9]+\b`)
	debugOpcodePattern = regexp.MustCompile(`^[A-Z][A-Z0-9]+`)
)

// colorizeInstruction applies syntax highlighting to an instruction string
func colorizeInstructionDebug(instr string) string {
	instr = strings.TrimSpace(instr)
	if instr == "" {
		return instr
	}

	// Find and extract the opcode first
	opcodeLoc := debugOpcodePattern.FindStringIndex(instr)
	if opcodeLoc == nil {
		return instr // No opcode found, return as-is
	}

	opcode := instr[opcodeLoc[0]:opcodeLoc[1]]
	rest := instr[opcodeLoc[1]:]

	// Build the result with colored opcode
	result := instrOpcode.Sprint(opcode)

	// Process the rest character by character, applying colors
	// First, find all register and immediate matches
	regMatches := debugRegPattern.FindAllStringIndex(rest, -1)
	immMatches := debugImmPattern.FindAllStringIndex(rest, -1)

	// Create a map of positions to colors
	type colorSpan struct {
		start int
		end   int
		color *color.Color
		text  string
	}
	var spans []colorSpan

	for _, m := range regMatches {
		spans = append(spans, colorSpan{m[0], m[1], instrReg, rest[m[0]:m[1]]})
	}
	for _, m := range immMatches {
		// Check if this immediate overlaps with a register (registers take priority)
		overlaps := false
		for _, rm := range regMatches {
			if m[0] < rm[1] && m[1] > rm[0] {
				overlaps = true
				break
			}
		}
		if !overlaps {
			spans = append(spans, colorSpan{m[0], m[1], instrImm, rest[m[0]:m[1]]})
		}
	}

	// Sort spans by start position
	for i := 0; i < len(spans); i++ {
		for j := i + 1; j < len(spans); j++ {
			if spans[j].start < spans[i].start {
				spans[i], spans[j] = spans[j], spans[i]
			}
		}
	}

	// Build the rest of the string with colors
	pos := 0
	for _, span := range spans {
		if span.start > pos {
			// Add uncolored text between spans
			result += instrPunct.Sprint(rest[pos:span.start])
		}
		result += span.color.Sprint(span.text)
		pos = span.end
	}
	// Add any remaining text
	if pos < len(rest) {
		result += instrPunct.Sprint(rest[pos:])
	}

	return result
}

var (
	debugMemorySize    uint32
	debugVerbose       bool
	debugCompileFormat string
)

var debugCmd = &cobra.Command{
	Use:   "debug <file>",
	Short: "Interactive debugger for cucaracha programs",
	Long: `Starts an interactive debugging session for a cucaracha program.

The command accepts:
  - Assembly files (.cucaracha, .s) - parsed by the LLVM assembly parser
  - Binary/object files (.o) - parsed by the ELF binary parser
  - C/C++ source files (.c, .cpp, etc.) - compiled first, then debugged

When a source file is provided, it is automatically compiled using clang
with the Cucaracha target before debugging.

For binary/object files compiled with debug info (-g), source-level debugging
is available:
  - The debugger will show the corresponding source code line when stepping
  - Use 'source' to view source code around the current location
  - Use 'vars' to see variables accessible at the current location

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
  source, src [n]    - Show n lines of source code around current location
  vars, v            - Show variables accessible at current location
  help, h            - Show this help
  quit, q            - Exit debugger

Example:
  cucaracha cpu debug program.cucaracha
  cucaracha cpu debug program.c`,
	Args: cobra.ExactArgs(1),
	Run:  runDebug,
}

func init() {
	CpuCmd.AddCommand(debugCmd)
	debugCmd.Flags().Uint32VarP(&debugMemorySize, "memory", "m", 0x20000, "Memory size in bytes (default: 128KB)")
	debugCmd.Flags().BoolVarP(&debugVerbose, "verbose", "v", false, "Print verbose output")
	debugCmd.Flags().StringVar(&debugCompileFormat, "compile-to", "object", "Compilation output format for source files: assembly, object (default: object for DWARF debug info)")
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
	// Debug information for source-level debugging
	debugInfo *mc.DebugInfo
	// Last shown source location (to avoid repetitive display)
	lastSourceLoc *mc.SourceLocation
}

func runDebug(cmd *cobra.Command, args []string) {
	inputPath := args[0]

	// Check if it's a source file that needs compilation
	var cleanup func()
	if llvm.IsSourceFile(inputPath) {
		var outputFormat llvm.OutputFormat
		switch strings.ToLower(debugCompileFormat) {
		case "assembly", "asm":
			outputFormat = llvm.OutputAssembly
		case "object", "obj", "o":
			outputFormat = llvm.OutputObject
		default:
			outputFormat = llvm.OutputObject // Default to object for DWARF debug info
		}

		compiledPath, cleanupFn, err := CompileSourceFile(inputPath, outputFormat, debugVerbose)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error compiling source file: %v\n", err)
			os.Exit(1)
		}
		cleanup = cleanupFn
		inputPath = compiledPath
		fmt.Printf("Compiled %s -> %s\n", args[0], compiledPath)
	}
	if cleanup != nil {
		defer cleanup()
	}

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

	// Get debug info from resolved program
	debugInfo := resolved.DebugInfo()
	if debugInfo != nil {
		// Try to load source files if debug info available
		debugInfo.TryLoadSourceFiles()
	}

	session := &debugSession{
		dbg:             dbg,
		pf:              resolved,
		addrToIdx:       addrToIdx,
		idxToAddr:       idxToAddr,
		running:         true,
		terminationAddr: TerminationAddress,
		debugInfo:       debugInfo,
	}

	fmt.Printf("Loaded %d instructions\n", len(resolved.Instructions()))
	if debugInfo != nil && len(debugInfo.InstructionLocations) > 0 {
		colorSuccess.Printf("Debug info: %d source locations\n", len(debugInfo.InstructionLocations))
	}
	fmt.Printf("Entry point: %s\n", colorAddr.Sprintf("0x%08X", interp.State().PC))
	colorSuccess.Println("Type 'help' for available commands.")
	fmt.Println()

	// Show initial state
	session.showCurrentInstruction()

	// Start interactive loop
	reader := bufio.NewReader(os.Stdin)
	for session.running {
		colorPrompt.Print("(cucaracha) ")
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
	case "source", "src":
		s.cmdSource(args)
	case "vars", "v":
		s.cmdVars()
	case "help", "h", "?":
		s.cmdHelp()
	case "quit", "q", "exit":
		s.running = false
		colorSuccess.Println("Exiting debugger.")
	default:
		colorError.Printf("Unknown command: %s. ", cmd)
		fmt.Println("Type 'help' for available commands.")
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
			colorError.Printf("Error: %v\n", result.Error)
			return
		}
		if result.StopReason == interpreter.StopTermination {
			colorSuccess.Println("Program terminated.")
			s.showReturnValue()
			return
		}
		if result.StopReason == interpreter.StopHalt {
			colorWarning.Println("CPU halted.")
			return
		}
		if result.StopReason == interpreter.StopBreakpoint {
			colorBreakpoint.Printf("Breakpoint hit at %s\n", colorAddr.Sprintf("0x%08X", s.dbg.State().PC))
			break
		}
		if result.StopReason == interpreter.StopWatchpoint {
			colorWarning.Printf("Watchpoint triggered at %s\n", colorAddr.Sprintf("0x%08X", s.dbg.State().PC))
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
		colorBreakpoint.Printf("Breakpoint hit at %s\n", colorAddr.Sprintf("0x%08X", s.dbg.State().PC))
		s.showCurrentInstruction()
	case interpreter.StopWatchpoint:
		colorWarning.Printf("Watchpoint triggered at %s\n", colorAddr.Sprintf("0x%08X", s.dbg.State().PC))
		s.showCurrentInstruction()
	case interpreter.StopTermination:
		colorSuccess.Printf("Program terminated after %s steps.\n", colorValue.Sprintf("%d", result.StepsExecuted))
		s.showReturnValue()
	case interpreter.StopHalt:
		colorWarning.Println("CPU halted.")
	case interpreter.StopError:
		colorError.Printf("Error: %v\n", result.Error)
	default:
		fmt.Printf("Stopped: %s after %s steps\n", result.StopReason, colorValue.Sprintf("%d", result.StepsExecuted))
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
		colorError.Printf("Invalid address: %s\n", args[0])
		return
	}

	bp := s.dbg.AddBreakpoint(addr)
	colorSuccess.Printf("Breakpoint %s set at %s\n", colorValue.Sprintf("%d", bp.ID), colorAddr.Sprintf("0x%08X", addr))
}

func (s *debugSession) cmdWatch(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: watch <address>")
		return
	}

	addr, err := parseAddress(args[0])
	if err != nil {
		colorError.Printf("Invalid address: %s\n", args[0])
		return
	}

	wp := s.dbg.AddWatchpoint(addr, 4, interpreter.WatchWrite)
	colorSuccess.Printf("Watchpoint %s set at %s (4 bytes, write)\n", colorValue.Sprintf("%d", wp.ID), colorAddr.Sprintf("0x%08X", addr))
}

func (s *debugSession) cmdDelete(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: delete <breakpoint-id>")
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		colorError.Printf("Invalid ID: %s\n", args[0])
		return
	}

	if s.dbg.RemoveBreakpoint(id) {
		colorSuccess.Printf("Breakpoint %d deleted.\n", id)
	} else if s.dbg.RemoveWatchpoint(id) {
		colorSuccess.Printf("Watchpoint %d deleted.\n", id)
	} else {
		colorWarning.Printf("No breakpoint or watchpoint with ID %d.\n", id)
	}
}

func (s *debugSession) cmdList() {
	bps := s.dbg.ListBreakpoints()
	wps := s.dbg.ListWatchpoints()

	if len(bps) == 0 && len(wps) == 0 {
		colorWarning.Println("No breakpoints or watchpoints set.")
		return
	}

	if len(bps) > 0 {
		colorHeader.Println("Breakpoints:")
		for _, bp := range bps {
			var status string
			if bp.Enabled {
				status = colorSuccess.Sprint("enabled")
			} else {
				status = colorFlagClear.Sprint("disabled")
			}
			instrText := s.getInstructionText(bp.Address)
			fmt.Printf("  %s: %s (%s) %s\n",
				colorValue.Sprintf("%d", bp.ID),
				colorAddr.Sprintf("0x%08X", bp.Address),
				status,
				colorizeInstructionDebug(instrText))
		}
	}

	if len(wps) > 0 {
		colorHeader.Println("Watchpoints:")
		for _, wp := range wps {
			var status string
			if wp.Enabled {
				status = colorSuccess.Sprint("enabled")
			} else {
				status = colorFlagClear.Sprint("disabled")
			}
			typeStr := "read/write"
			switch wp.Type {
			case interpreter.WatchRead:
				typeStr = "read"
			case interpreter.WatchWrite:
				typeStr = "write"
			}
			fmt.Printf("  %s: %s (%d bytes, %s, %s)\n",
				colorValue.Sprintf("%d", wp.ID),
				colorAddr.Sprintf("0x%08X", wp.Address),
				wp.Size,
				typeStr,
				status)
		}
	}
}

func (s *debugSession) cmdPrint(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: print <register|@address>")
		fmt.Printf("  Registers: %s\n", colorReg.Sprint("r0-r9, sp, lr, pc, cpsr"))
		fmt.Printf("  Memory: %s or %s\n", colorAddr.Sprint("@0x1234"), colorAddr.Sprint("@1234"))
		return
	}

	what := strings.ToLower(args[0])

	// Memory access
	if strings.HasPrefix(what, "@") {
		addr, err := parseAddress(what[1:])
		if err != nil {
			colorError.Printf("Invalid address: %s\n", what[1:])
			return
		}
		data, err := s.dbg.ReadMemory(addr, 4)
		if err != nil || len(data) != 4 {
			colorError.Printf("Could not read memory at %s\n", colorAddr.Sprintf("0x%08X", addr))
			return
		}
		val := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
		fmt.Printf("[%s] = %s (%s)\n",
			colorAddr.Sprintf("@0x%08X", addr),
			colorValue.Sprintf("%d", int32(val)),
			colorHex.Sprintf("0x%08X", val))
		return
	}

	// Special registers
	switch what {
	case "pc":
		val := s.dbg.GetPC()
		fmt.Printf("%s = %s (%s)\n", colorReg.Sprint("pc"), colorValue.Sprintf("%d", val), colorHex.Sprintf("0x%08X", val))
		return
	case "sp":
		val := s.dbg.GetSP()
		fmt.Printf("%s = %s (%s)\n", colorReg.Sprint("sp"), colorValue.Sprintf("%d", val), colorHex.Sprintf("0x%08X", val))
		return
	case "lr":
		val := s.dbg.GetLR()
		fmt.Printf("%s = %s (%s)\n", colorReg.Sprint("lr"), colorValue.Sprintf("%d", val), colorHex.Sprintf("0x%08X", val))
		return
	}

	// General register access by name
	regIdx, ok := parseRegisterName(what)
	if !ok {
		colorError.Printf("Unknown register: %s\n", what)
		return
	}
	val := s.dbg.GetRegister(regIdx)
	fmt.Printf("%s = %s (%s)\n", colorReg.Sprint(what), colorValue.Sprintf("%d", int32(val)), colorHex.Sprintf("0x%08X", val))
}

func (s *debugSession) cmdSet(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: set <register> <value>")
		return
	}

	reg := strings.ToLower(args[0])
	val, err := parseValue(args[1])
	if err != nil {
		colorError.Printf("Invalid value: %s\n", args[1])
		return
	}

	// Special registers
	switch reg {
	case "pc":
		s.dbg.SetPC(val)
		fmt.Printf("%s = %s (%s)\n", colorReg.Sprint("pc"), colorValue.Sprintf("%d", val), colorHex.Sprintf("0x%08X", val))
		return
	case "sp":
		s.dbg.SetSP(val)
		fmt.Printf("%s = %s (%s)\n", colorReg.Sprint("sp"), colorValue.Sprintf("%d", val), colorHex.Sprintf("0x%08X", val))
		return
	case "lr":
		s.dbg.SetLR(val)
		fmt.Printf("%s = %s (%s)\n", colorReg.Sprint("lr"), colorValue.Sprintf("%d", val), colorHex.Sprintf("0x%08X", val))
		return
	}

	// General register by name
	regIdx, ok := parseRegisterName(reg)
	if !ok {
		colorError.Printf("Unknown register: %s\n", reg)
		return
	}
	s.dbg.SetRegister(regIdx, val)
	fmt.Printf("%s = %s (%s)\n", colorReg.Sprint(reg), colorValue.Sprintf("%d", int32(val)), colorHex.Sprintf("0x%08X", val))
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
		colorError.Printf("No instruction at %s\n", colorAddr.Sprintf("0x%08X", addr))
		return
	}

	fmt.Printf("Disassembly at %s:\n", colorAddr.Sprintf("0x%08X", addr))
	for i := 0; i < count && idx+i < len(instrs); i++ {
		instr := instrs[idx+i]
		if instr.Address == nil {
			continue
		}
		marker := "  "
		markerColor := color.New()
		isPC := *instr.Address == s.dbg.State().PC
		isBP := false

		// Check for breakpoint
		for _, bp := range s.dbg.ListBreakpoints() {
			if bp.Address == *instr.Address && bp.Enabled {
				isBP = true
				break
			}
		}

		if isBP && isPC {
			marker = "*>"
			markerColor = colorBreakpoint
		} else if isBP {
			marker = "* "
			markerColor = colorBreakpoint
		} else if isPC {
			marker = "=>"
			markerColor = colorPC
		}

		fmt.Printf("%s %s: %s\n",
			markerColor.Sprint(marker),
			colorAddr.Sprintf("0x%08X", *instr.Address),
			colorizeInstructionDebug(instr.Text))
	}
}

func (s *debugSession) cmdInfo() {
	state := s.dbg.State()

	colorHeader.Println("=== CPU State ===")
	fmt.Printf("%s:   %s\n", colorReg.Sprint("PC"), colorAddr.Sprintf("0x%08X", state.PC))
	fmt.Printf("%s:   %s (%s)\n", colorReg.Sprint("SP"), colorValue.Sprintf("%d", *state.SP), colorHex.Sprintf("0x%08X", *state.SP))
	fmt.Printf("%s:   %s\n", colorReg.Sprint("LR"), colorAddr.Sprintf("0x%08X", *state.LR))

	// Get CPSR from the cpsr register
	cpsrIdx := uint32(registers.Register("cpsr").Encode())
	cpsr := state.Registers[cpsrIdx]

	// Format CPSR flags with colors
	flagN := (cpsr >> 3) & 1
	flagZ := (cpsr >> 0) & 1
	flagC := (cpsr >> 1) & 1
	flagV := (cpsr >> 2) & 1

	formatFlag := func(name string, val uint32) string {
		if val == 1 {
			return colorFlagSet.Sprintf("%s=%d", name, val)
		}
		return colorFlagClear.Sprintf("%s=%d", name, val)
	}

	fmt.Printf("%s: %s (%s %s %s %s)\n",
		colorReg.Sprint("CPSR"),
		colorHex.Sprintf("0x%08X", cpsr),
		formatFlag("N", flagN),
		formatFlag("Z", flagZ),
		formatFlag("C", flagC),
		formatFlag("V", flagV))
	fmt.Println()

	colorHeader.Println("General Purpose Registers:")
	for i := 0; i < 10; i++ {
		regIdx := 16 + i // r0 starts at index 16
		val := state.Registers[regIdx]
		fmt.Printf("  %s = %s (%s)\n",
			colorReg.Sprintf("r%d", i),
			colorValue.Sprintf("%10d", int32(val)),
			colorHex.Sprintf("0x%08X", val))
	}
}

func (s *debugSession) cmdStack() {
	state := s.dbg.State()
	sp := *state.SP

	fmt.Printf("Stack (%s = %s):\n", colorReg.Sprint("SP"), colorAddr.Sprintf("0x%08X", sp))
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
			marker = colorPC.Sprint(" <- SP")
		}
		fmt.Printf("  %s: %s (%s)%s\n",
			colorAddr.Sprintf("0x%08X", addr),
			colorValue.Sprintf("%10d", int32(val)),
			colorHex.Sprintf("0x%08X", val),
			marker)
	}
}

func (s *debugSession) cmdMemory(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: memory <address> [count]")
		return
	}

	addr, err := parseAddress(args[0])
	if err != nil {
		colorError.Printf("Invalid address: %s\n", args[0])
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
		colorError.Printf("Could not read memory at %s\n", colorAddr.Sprintf("0x%08X", addr))
		return
	}

	fmt.Printf("Memory at %s:\n", colorAddr.Sprintf("0x%08X", addr))
	for i := 0; i < len(data); i += 16 {
		fmt.Printf("  %s: ", colorAddr.Sprintf("0x%08X", addr+uint32(i)))
		// Hex bytes
		for j := 0; j < 16 && i+j < len(data); j++ {
			colorHex.Printf("%02X ", data[i+j])
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

func (s *debugSession) cmdSource(args []string) {
	if s.debugInfo == nil {
		colorWarning.Println("No debug information available.")
		return
	}

	pc := s.dbg.State().PC

	// Get current source location
	loc := s.debugInfo.GetSourceLocation(pc)
	if loc == nil || !loc.IsValid() {
		colorWarning.Println("No source location for current address.")
		return
	}

	// Determine how many lines to show (default: 10 centered on current line)
	contextLines := 5
	if len(args) > 0 {
		if n, err := strconv.Atoi(args[0]); err == nil && n > 0 {
			contextLines = n / 2
		}
	}

	fmt.Printf("%s:\n", colorSourceFile.Sprint(loc.File))

	// Show source lines
	startLine := loc.Line - contextLines
	if startLine < 1 {
		startLine = 1
	}
	endLine := loc.Line + contextLines

	for line := startLine; line <= endLine; line++ {
		srcLine := s.debugInfo.GetSourceLine(loc.File, line)
		if srcLine == "" && line > loc.Line {
			// End of file
			break
		}
		marker := "   "
		lineColor := colorSource
		if line == loc.Line {
			marker = colorPC.Sprint("=>")
			lineColor = color.New(color.FgWhite, color.Bold)
		}
		fmt.Printf("%s %s %s\n",
			marker,
			colorSourceLine.Sprintf("%4d", line),
			lineColor.Sprint(strings.TrimRight(srcLine, "\r\n")))
	}
}

func (s *debugSession) cmdVars() {
	if s.debugInfo == nil {
		colorWarning.Println("No debug information available.")
		return
	}

	pc := s.dbg.State().PC
	vars := s.debugInfo.GetVariables(pc)

	if len(vars) == 0 {
		colorWarning.Println("No variables accessible at current location.")
		return
	}

	colorHeader.Println("Accessible Variables:")
	for _, v := range vars {
		// Format variable location
		var locStr string
		switch loc := v.Location.(type) {
		case mc.RegisterLocation:
			regName := fmt.Sprintf("r%d", loc.Register)
			if loc.Register == 13 {
				regName = "sp"
			} else if loc.Register == 14 {
				regName = "lr"
			}
			locStr = colorReg.Sprintf("@%s", regName)
		case mc.MemoryLocation:
			baseReg := fmt.Sprintf("r%d", loc.BaseRegister)
			if loc.BaseRegister == 13 {
				baseReg = "sp"
			} else if loc.BaseRegister == 14 {
				baseReg = "lr"
			}
			if loc.Offset >= 0 {
				locStr = colorAddr.Sprintf("[%s+%d]", baseReg, loc.Offset)
			} else {
				locStr = colorAddr.Sprintf("[%s%d]", baseReg, loc.Offset)
			}
		case mc.ConstantLocation:
			locStr = colorValue.Sprintf("=%d", loc.Value)
		default:
			locStr = "?"
		}

		// Try to read the value if in register or memory
		var valueStr string
		switch loc := v.Location.(type) {
		case mc.RegisterLocation:
			regIdx := 16 + loc.Register // r0 starts at index 16
			if loc.Register < 10 {
				val := s.dbg.State().Registers[regIdx]
				valueStr = colorValue.Sprintf(" = %d", int32(val))
			}
		case mc.MemoryLocation:
			// Compute memory address
			var baseVal uint32
			if loc.BaseRegister == 13 {
				baseVal = s.dbg.GetSP()
			} else if loc.BaseRegister == 14 {
				baseVal = s.dbg.GetLR()
			} else if loc.BaseRegister < 10 {
				baseVal = s.dbg.State().Registers[16+loc.BaseRegister]
			}
			addr := uint32(int32(baseVal) + loc.Offset)
			if data, err := s.dbg.ReadMemory(addr, 4); err == nil && len(data) == 4 {
				val := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
				valueStr = colorValue.Sprintf(" = %d", int32(val))
			}
		}

		paramStr := ""
		if v.IsParameter {
			paramStr = colorVarType.Sprint(" (param)")
		}

		fmt.Printf("  %s %s: %s%s%s\n",
			colorVarName.Sprint(v.Name),
			colorVarType.Sprint(v.TypeName),
			locStr,
			valueStr,
			paramStr)
	}
}

func (s *debugSession) cmdHelp() {
	colorHeader.Println("Cucaracha Debugger Commands:")
	fmt.Println()

	colorHeader.Println("Execution:")
	fmt.Printf("  %s, %s [n]        - Step n instructions (default: 1)\n", colorInstr.Sprint("step"), colorInstr.Sprint("s"))
	fmt.Printf("  %s, %s        - Continue execution until breakpoint\n", colorInstr.Sprint("continue"), colorInstr.Sprint("c"))
	fmt.Printf("  %s, %s             - Run until termination or breakpoint\n", colorInstr.Sprint("run"), colorInstr.Sprint("r"))
	fmt.Println()

	colorHeader.Println("Breakpoints:")
	fmt.Printf("  %s, %s <addr>    - Set breakpoint at address (hex: 0x... or decimal)\n", colorInstr.Sprint("break"), colorInstr.Sprint("b"))
	fmt.Printf("  %s, %s <addr>    - Set watchpoint on memory address\n", colorInstr.Sprint("watch"), colorInstr.Sprint("w"))
	fmt.Printf("  %s, %s <id>     - Delete breakpoint/watchpoint by ID\n", colorInstr.Sprint("delete"), colorInstr.Sprint("d"))
	fmt.Printf("  %s, %s            - List all breakpoints and watchpoints\n", colorInstr.Sprint("list"), colorInstr.Sprint("l"))
	fmt.Println()

	colorHeader.Println("Inspection:")
	fmt.Printf("  %s, %s <what>    - Print register (%s) or memory (%s)\n",
		colorInstr.Sprint("print"), colorInstr.Sprint("p"),
		colorReg.Sprint("r0-r9, sp, lr, pc"),
		colorAddr.Sprint("@addr"))
	fmt.Printf("  %s <reg> <value>  - Set register value\n", colorInstr.Sprint("set"))
	fmt.Printf("  %s, %s [addr] [n] - Disassemble n instructions at addr\n", colorInstr.Sprint("disasm"), colorInstr.Sprint("x"))
	fmt.Printf("  %s, %s            - Show CPU state (registers, flags)\n", colorInstr.Sprint("info"), colorInstr.Sprint("i"))
	fmt.Printf("  %s              - Show stack contents\n", colorInstr.Sprint("stack"))
	fmt.Printf("  %s, %s <addr> [n] - Show n bytes of memory at addr\n", colorInstr.Sprint("memory"), colorInstr.Sprint("m"))
	fmt.Println()

	colorHeader.Println("Source-Level Debugging:")
	fmt.Printf("  %s, %s [n]   - Show source code around current line\n", colorInstr.Sprint("source"), colorInstr.Sprint("src"))
	fmt.Printf("  %s, %s            - Show accessible variables at current location\n", colorInstr.Sprint("vars"), colorInstr.Sprint("v"))
	fmt.Println()

	colorHeader.Println("Other:")
	fmt.Printf("  %s, %s            - Show this help\n", colorInstr.Sprint("help"), colorInstr.Sprint("h"))
	fmt.Printf("  %s, %s            - Exit debugger\n", colorInstr.Sprint("quit"), colorInstr.Sprint("q"))
	fmt.Println()
	fmt.Println("Press Enter to repeat the last command.")
}

func (s *debugSession) showCurrentInstruction() {
	state := s.dbg.State()
	pc := state.PC

	// Show source location if available and changed
	if s.debugInfo != nil {
		if loc := s.debugInfo.GetSourceLocation(pc); loc != nil && loc.IsValid() {
			// Check if source location changed (different file or line)
			showSource := s.lastSourceLoc == nil ||
				s.lastSourceLoc.File != loc.File ||
				s.lastSourceLoc.Line != loc.Line
			if showSource {
				s.lastSourceLoc = loc
				// Show source file and line
				fmt.Printf("%s %s\n",
					colorSourceFile.Sprint(loc.File+":"),
					colorSourceLine.Sprintf("%d", loc.Line))
				// Show actual source line if available
				if srcLine := s.debugInfo.GetSourceLine(loc.File, loc.Line); srcLine != "" {
					fmt.Printf("   %s\n", colorSource.Sprint(strings.TrimRight(srcLine, "\r\n")))
				}
			}
		}
	}

	instrText := s.getInstructionText(pc)
	word, _ := state.ReadMemory32(pc)

	fmt.Printf("%s %s [%s]: %s\n",
		colorPC.Sprint("=>"),
		colorAddr.Sprintf("0x%08X", pc),
		colorHex.Sprintf("%08X", word),
		colorizeInstructionDebug(instrText))
}

func (s *debugSession) showReturnValue() {
	state := s.dbg.State()
	r0 := state.Registers[16]
	fmt.Printf("Return value (%s): %s (%s)\n",
		colorReg.Sprint("r0"),
		colorValue.Sprintf("%d", int32(r0)),
		colorHex.Sprintf("0x%08X", r0))
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
