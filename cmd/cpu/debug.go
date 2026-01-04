package cpu

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/debugger"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/interpreter"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/loader"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/Manu343726/cucaracha/pkg/utils"
	"github.com/fatih/color"
	"github.com/peterh/liner"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// =============================================================================
// Color definitions for CLI output
// =============================================================================

var (
	colorAddr       = color.New(color.FgCyan)
	colorInstr      = color.New(color.FgYellow)
	colorReg        = color.New(color.FgGreen)
	colorValue      = color.New(color.FgWhite, color.Bold)
	colorHex        = color.New(color.FgMagenta)
	colorPrompt     = color.New(color.FgBlue, color.Bold)
	colorError      = color.New(color.FgRed, color.Bold)
	colorSuccess    = color.New(color.FgGreen)
	colorWarning    = color.New(color.FgYellow)
	colorHeader     = color.New(color.FgWhite, color.Bold, color.Underline)
	colorBreakpoint = color.New(color.FgRed, color.Bold)
	colorPC         = color.New(color.FgGreen, color.Bold)
	colorFlagSet    = color.New(color.FgGreen, color.Bold)
	colorFlagClear  = color.New(color.FgHiBlack)
	colorSource     = color.New(color.FgHiWhite)
	colorSourceFile = color.New(color.FgHiBlue)
	colorSourceLine = color.New(color.FgHiCyan)
	colorVarName    = color.New(color.FgHiGreen)
	colorVarType    = color.New(color.FgHiYellow)
	colorHiBlack    = color.New(color.FgHiBlack)
	colorFunc       = color.New(color.FgHiMagenta, color.Bold)
)

// Instruction part colors
var (
	instrOpcode = color.New(color.FgYellow, color.Bold)
	instrReg    = color.New(color.FgGreen)
	instrImm    = color.New(color.FgCyan)
	instrPunct  = color.New(color.FgWhite)
)

// Memory region colors
var (
	colorMemCode    = color.New(color.FgYellow)
	colorMemData    = color.New(color.FgGreen)
	colorMemStack   = color.New(color.FgCyan)
	colorMemUnknown = color.New(color.FgHiBlack)
)

// Variable color palette - distinct colors for different variables
var variableColorPalette = []*color.Color{
	color.New(color.FgHiGreen),
	color.New(color.FgHiYellow),
	color.New(color.FgHiCyan),
	color.New(color.FgHiMagenta),
	color.New(color.FgHiBlue),
	color.New(color.FgHiRed),
	color.New(color.FgGreen),
	color.New(color.FgYellow),
	color.New(color.FgCyan),
	color.New(color.FgMagenta),
}

// Regex patterns for instruction highlighting
var (
	debugRegPattern    = regexp.MustCompile(`\b(r[0-9]{1,2}|sp|lr|pc|cpsr)\b`)
	debugImmPattern    = regexp.MustCompile(`#-?[0-9]+|#-?0x[0-9a-fA-F]+|\b-?[0-9]+\b`)
	debugOpcodePattern = regexp.MustCompile(`^[A-Z][A-Z0-9]+`)
)

// =============================================================================
// CLI UI Implementation - Implements debugger.DebuggerUI
// =============================================================================

// cliUI implements the debugger.DebuggerUI interface for terminal output
type cliUI struct {
	liner          *liner.State
	resizeHandlers []debugger.ResizeHandler
	resizeStop     chan struct{}
	resizeMu       sync.Mutex
}

// Ensure cliUI implements DebuggerUI
var _ debugger.DebuggerUI = (*cliUI)(nil)

// OnEvent handles debugger events
func (ui *cliUI) OnEvent(event debugger.EventData) {
	switch event.Event {
	case debugger.EventBreakpointHit:
		colorBreakpoint.Printf("Breakpoint %d hit at %s\n",
			event.BreakpointID,
			colorAddr.Sprintf("0x%08X", event.Address))

	case debugger.EventWatchpointHit:
		colorWarning.Printf("Watchpoint %d triggered at %s\n",
			event.WatchpointID,
			colorAddr.Sprintf("0x%08X", event.Address))

	case debugger.EventProgramTerminated:
		colorSuccess.Printf("Program terminated after %d steps.\n", event.StepsExecuted)
		fmt.Printf("Return value (%s): %s (%s)\n",
			colorReg.Sprint("r0"),
			colorValue.Sprintf("%d", int32(event.ReturnValue)),
			colorHex.Sprintf("0x%08X", event.ReturnValue))

	case debugger.EventProgramHalted:
		colorWarning.Println("CPU halted.")

	case debugger.EventInterrupted:
		colorWarning.Printf("\nInterrupted after %d steps at %s\n",
			event.StepsExecuted,
			colorAddr.Sprintf("0x%08X", event.Address))

	case debugger.EventError:
		colorError.Printf("Error: %v\n", event.Error)

	case debugger.EventSourceLocationChanged:
		if event.SourceLocation != nil && event.SourceLocation.IsValid() {
			fmt.Printf("%s %s\n",
				colorSourceFile.Sprint(event.SourceLocation.File+":"),
				colorSourceLine.Sprintf("%d", event.SourceLocation.Line))
			if event.SourceText != "" {
				fmt.Printf("   %s\n", colorSource.Sprint(strings.TrimRight(event.SourceText, "\r\n")))
			}
		}
	}
}

// ShowMessage displays a message with appropriate color based on level
func (ui *cliUI) ShowMessage(level debugger.MessageLevel, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	switch level {
	case debugger.LevelError:
		colorError.Println(message)
	case debugger.LevelWarning:
		colorWarning.Println(message)
	case debugger.LevelSuccess:
		colorSuccess.Println(message)
	case debugger.LevelDebug:
		colorHiBlack.Println(message)
	default:
		fmt.Println(message)
	}
}

// ShowInstruction displays the current instruction
func (ui *cliUI) ShowInstruction(info debugger.InstructionInfo) {
	instrText := info.Mnemonic
	if info.Operands != "" {
		instrText += " " + info.Operands
	}

	fmt.Printf("%s %s [%s]: %s\n",
		colorPC.Sprint("=>"),
		colorAddr.Sprintf("0x%08X", info.Address),
		colorHex.Sprintf("%08X", info.Encoding),
		colorizeInstruction(instrText))
}

// ShowRegisters displays register values
func (ui *cliUI) ShowRegisters(regs []debugger.RegisterInfo, flags debugger.FlagState) {
	colorHeader.Println("=== CPU State ===")

	// Find special registers
	var pc, sp, lr, cpsr uint32
	for _, reg := range regs {
		switch reg.Name {
		case "pc":
			pc = reg.Value
		case "sp":
			sp = reg.Value
		case "lr":
			lr = reg.Value
		case "cpsr":
			cpsr = reg.Value
		}
	}

	fmt.Printf("%s:   %s\n", colorReg.Sprint("PC"), colorAddr.Sprintf("0x%08X", pc))
	fmt.Printf("%s:   %s (%s)\n", colorReg.Sprint("SP"), colorValue.Sprintf("%d", sp), colorHex.Sprintf("0x%08X", sp))
	fmt.Printf("%s:   %s\n", colorReg.Sprint("LR"), colorAddr.Sprintf("0x%08X", lr))

	// Format flags
	formatFlag := func(name string, set bool) string {
		if set {
			return colorFlagSet.Sprint(name)
		}
		return colorFlagClear.Sprint(name)
	}
	fmt.Printf("%s: %s %s %s %s (%s)\n",
		colorReg.Sprint("CPSR"),
		formatFlag("N", flags.N),
		formatFlag("Z", flags.Z),
		formatFlag("C", flags.C),
		formatFlag("V", flags.V),
		colorHex.Sprintf("0x%08X", cpsr))
	fmt.Println()

	colorHeader.Println("Registers:")
	for _, reg := range regs {
		if strings.HasPrefix(reg.Name, "r") {
			fmt.Printf("  %s = %s (%s)\n",
				colorReg.Sprint(reg.Name),
				colorValue.Sprintf("%10d", int32(reg.Value)),
				colorHex.Sprintf("0x%08X", reg.Value))
		}
	}
}

// ShowMemory displays memory contents
func (ui *cliUI) ShowMemory(addr uint32, data []byte, regions []debugger.MemoryRegion) {
	fmt.Printf("Memory at %s:\n", colorAddr.Sprintf("0x%08X", addr))
	ui.printMemoryLegend()

	for i := 0; i < len(data); i += 16 {
		lineAddr := addr + uint32(i)
		regionType, regionLabel := ui.getRegionForAddress(lineAddr, regions)

		// Print region marker
		marker := ui.getRegionMarker(regionType)
		fmt.Printf("%s ", marker)
		fmt.Printf("%s: ", colorAddr.Sprintf("0x%08X", lineAddr))

		// Hex bytes with color based on region
		for j := 0; j < 16 && i+j < len(data); j++ {
			byteAddr := lineAddr + uint32(j)
			byteRegion, _ := ui.getRegionForAddress(byteAddr, regions)
			ui.printColoredByte(data[i+j], byteRegion)
		}
		// Padding
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
		fmt.Print("|")
		if regionLabel != "" {
			fmt.Printf(" %s", colorHiBlack.Sprint(regionLabel))
		}
		fmt.Println()
	}
}

// ShowDisassembly displays disassembled instructions
func (ui *cliUI) ShowDisassembly(instructions []debugger.InstructionInfo, currentPC uint32) {
	if len(instructions) == 0 {
		return
	}
	fmt.Printf("Disassembly at %s:\n", colorAddr.Sprintf("0x%08X", instructions[0].Address))
	for _, instr := range instructions {
		marker := "  "
		markerColor := color.New()

		if instr.HasBreakpoint && instr.Address == currentPC {
			marker = "*>"
			markerColor = colorBreakpoint
		} else if instr.HasBreakpoint {
			marker = "* "
			markerColor = colorBreakpoint
		} else if instr.Address == currentPC {
			marker = "=>"
			markerColor = colorPC
		}

		instrText := instr.Mnemonic
		if instr.Operands != "" {
			instrText += " " + instr.Operands
		}

		fmt.Printf("%s %s: %s\n",
			markerColor.Sprint(marker),
			colorAddr.Sprintf("0x%08X", instr.Address),
			colorizeInstruction(instrText))
	}
}

// ShowBreakpoints displays the list of breakpoints
func (ui *cliUI) ShowBreakpoints(breakpoints []debugger.BreakpointInfo) {
	if len(breakpoints) == 0 {
		return
	}
	colorHeader.Println("Breakpoints:")
	for _, bp := range breakpoints {
		status := colorFlagClear.Sprint("disabled")
		if bp.Enabled {
			status = colorSuccess.Sprint("enabled")
		}

		// Build the main line: ID, address, status
		fmt.Printf("  %s: %s (%s)",
			colorValue.Sprintf("%d", bp.ID),
			colorAddr.Sprintf("0x%08X", bp.Address),
			status)

		// Add hit count if > 0
		if bp.HitCount > 0 {
			colorHiBlack.Printf(" [%d hits]", bp.HitCount)
		}
		fmt.Println()

		// Show instruction text if available
		if bp.InstructionText != "" {
			fmt.Printf("       ")
			printHighlightedInstruction(bp.InstructionText)
			fmt.Println()
		}

		// Show source location and code if available
		if bp.SourceFile != "" {
			fmt.Printf("       %s:%s",
				colorSourceFile.Sprint(filepath.Base(bp.SourceFile)),
				colorSourceLine.Sprintf("%d", bp.SourceLine))
			if bp.SourceText != "" {
				colorHiBlack.Printf("  %s", strings.TrimSpace(bp.SourceText))
			}
			fmt.Println()
		}
	}
}

// ShowWatchpoints displays the list of watchpoints
func (ui *cliUI) ShowWatchpoints(watchpoints []debugger.WatchpointInfo) {
	if len(watchpoints) == 0 {
		return
	}
	colorHeader.Println("Watchpoints:")
	for _, wp := range watchpoints {
		status := colorFlagClear.Sprint("disabled")
		if wp.Enabled {
			status = colorSuccess.Sprint("enabled")
		}
		fmt.Printf("  %s: %s (%d bytes, %s, %s)",
			colorValue.Sprintf("%d", wp.ID),
			colorAddr.Sprintf("0x%08X", wp.Address),
			wp.Size,
			wp.Type,
			status)

		// Add hit count if > 0
		if wp.HitCount > 0 {
			colorHiBlack.Printf(" [%d hits]", wp.HitCount)
		}
		fmt.Println()
	}
}

// ShowStack displays the stack contents
func (ui *cliUI) ShowStack(sp uint32, data []byte, frames []debugger.StackFrame) {
	fmt.Printf("Stack (%s = %s):\n", colorReg.Sprint("SP"), colorAddr.Sprintf("0x%08X", sp))

	// Show stack entries (4 bytes each)
	for i := 0; i+4 <= len(data) && i < 40; i += 4 {
		addr := sp + uint32(i)
		val := uint32(data[i]) | uint32(data[i+1])<<8 | uint32(data[i+2])<<16 | uint32(data[i+3])<<24
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

// ShowBacktrace displays the call stack (function frames)
func (ui *cliUI) ShowBacktrace(frames []debugger.StackFrame, selectedFrame int) {
	if len(frames) == 0 {
		fmt.Println("No call stack information available")
		return
	}

	colorHeader.Println("Backtrace:")
	for i, frame := range frames {
		funcName := frame.Function
		if funcName == "" {
			funcName = "??"
		}

		location := ""
		if frame.File != "" {
			if frame.Line > 0 {
				location = fmt.Sprintf(" at %s:%d", frame.File, frame.Line)
			} else {
				location = fmt.Sprintf(" at %s", frame.File)
			}
		}

		// Show marker for selected frame
		marker := "  "
		if i == selectedFrame {
			marker = colorPC.Sprint("=>")
		}

		fmt.Printf("%s#%d  %s %s%s\n",
			marker,
			i,
			colorFunc.Sprint(funcName),
			colorAddr.Sprintf("[0x%08X]", frame.Address),
			colorSourceFile.Sprint(location))
	}
}

// ShowFrameInfo displays information about a single frame (for up/down commands)
func (ui *cliUI) ShowFrameInfo(frame debugger.StackFrame, frameNum int, sourceLine string) {
	funcName := frame.Function
	if funcName == "" {
		funcName = "??"
	}

	// Build location string
	location := ""
	if frame.File != "" {
		if frame.Line > 0 {
			location = fmt.Sprintf(" at %s:%d", frame.File, frame.Line)
		} else {
			location = fmt.Sprintf(" at %s", frame.File)
		}
	}

	// Print colored frame info
	fmt.Printf("#%d  %s %s%s\n",
		frameNum,
		colorFunc.Sprint(funcName),
		colorAddr.Sprintf("[0x%08X]", frame.Address),
		colorSourceFile.Sprint(location))

	// Print source line if available with syntax highlighting
	if sourceLine != "" && frame.Line > 0 {
		fmt.Printf("%s %s\n",
			colorSourceLine.Sprintf("%4d", frame.Line),
			utils.HighlightCCode(strings.TrimRight(sourceLine, "\r\n")))
	}
}

// ShowSource displays source code
func (ui *cliUI) ShowSource(location *mc.SourceLocation, lines []debugger.SourceLine, currentLine int) {
	if location == nil {
		return
	}
	fmt.Printf("%s:\n", colorSourceFile.Sprint(location.File))

	for _, line := range lines {
		marker := "   "
		lineColor := colorSource
		if line.IsCurrent {
			marker = colorPC.Sprint("=>")
			lineColor = color.New(color.FgWhite, color.Bold)
		}
		fmt.Printf("%s %s %s\n",
			marker,
			colorSourceLine.Sprintf("%4d", line.LineNumber),
			lineColor.Sprint(strings.TrimRight(line.Text, "\r\n")))
	}
}

// ShowVariables displays accessible variables
func (ui *cliUI) ShowVariables(variables []debugger.VariableValue) {
	colorHeader.Println("Accessible Variables:")
	for _, v := range variables {
		fmt.Printf("  %s %s: %s = %s\n",
			colorVarName.Sprint(v.Name),
			colorVarType.Sprint(v.TypeName),
			colorReg.Sprint(v.Location),
			colorValue.Sprint(v.ValueString))
	}
}

// ShowEvalResult displays the result of an expression evaluation
func (ui *cliUI) ShowEvalResult(expr string, value uint32, err error) {
	if err != nil {
		colorError.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("%s = %s (%s) [%s]\n",
		colorValue.Sprint(expr),
		colorValue.Sprintf("%d", int32(value)),
		colorHex.Sprintf("0x%08X", value),
		debugger.FormatBinary(value))
}

// ShowHelp displays help information
func (ui *cliUI) ShowHelp(commands []debugger.CommandHelp) {
	colorHeader.Println("Cucaracha Debugger Commands:")
	fmt.Println()

	// Group commands by category
	execution := []debugger.CommandHelp{}
	breakpoints := []debugger.CommandHelp{}
	inspection := []debugger.CommandHelp{}
	source := []debugger.CommandHelp{}
	other := []debugger.CommandHelp{}

	for _, cmd := range commands {
		switch cmd.Name {
		case "step", "stepi", "continue", "run":
			execution = append(execution, cmd)
		case "break", "watch", "delete", "list":
			breakpoints = append(breakpoints, cmd)
		case "print", "set", "disasm", "info", "stack", "memory", "eval":
			inspection = append(inspection, cmd)
		case "source", "vars":
			source = append(source, cmd)
		default:
			other = append(other, cmd)
		}
	}

	printGroup := func(name string, cmds []debugger.CommandHelp) {
		if len(cmds) == 0 {
			return
		}
		colorHeader.Println(name + ":")
		for _, cmd := range cmds {
			aliases := ""
			if len(cmd.Aliases) > 0 {
				aliases = ", " + strings.Join(cmd.Aliases, ", ")
			}
			fmt.Printf("  %s%s - %s\n",
				colorInstr.Sprint(cmd.Name),
				colorInstr.Sprint(aliases),
				cmd.Description)
		}
		fmt.Println()
	}

	printGroup("Execution", execution)
	printGroup("Breakpoints", breakpoints)
	printGroup("Inspection", inspection)
	printGroup("Source-Level Debugging", source)
	printGroup("Other", other)

	fmt.Println("Press Enter to repeat the last command.")
}

// Prompt requests input from the user
func (ui *cliUI) Prompt(prompt string) (string, error) {
	return ui.liner.Prompt(prompt)
}

// PromptConfirm requests a yes/no confirmation
func (ui *cliUI) PromptConfirm(message string) bool {
	input, err := ui.liner.Prompt(message + " (y/n) ")
	if err != nil {
		return false
	}
	return strings.ToLower(strings.TrimSpace(input)) == "y"
}

// GetTerminalSize returns the current terminal dimensions
func (ui *cliUI) GetTerminalSize() debugger.TerminalSize {
	width, height := getTerminalSize()
	return debugger.TerminalSize{Width: width, Height: height}
}

// OnResize registers a callback to be called when the terminal is resized
func (ui *cliUI) OnResize(handler debugger.ResizeHandler) (unregister func()) {
	ui.resizeMu.Lock()
	defer ui.resizeMu.Unlock()

	// Start the resize monitor if not already running
	if ui.resizeStop == nil {
		ui.resizeStop = make(chan struct{})
		go ui.monitorResize()
	}

	// Add the handler
	ui.resizeHandlers = append(ui.resizeHandlers, handler)
	handlerIndex := len(ui.resizeHandlers) - 1

	// Return unregister function
	return func() {
		ui.resizeMu.Lock()
		defer ui.resizeMu.Unlock()
		if handlerIndex < len(ui.resizeHandlers) {
			// Mark as nil instead of removing to preserve indices
			ui.resizeHandlers[handlerIndex] = nil
		}
	}
}

// notifyResizeHandlers calls all registered resize handlers
func (ui *cliUI) notifyResizeHandlers(size debugger.TerminalSize) {
	ui.resizeMu.Lock()
	handlers := make([]debugger.ResizeHandler, len(ui.resizeHandlers))
	copy(handlers, ui.resizeHandlers)
	ui.resizeMu.Unlock()

	for _, handler := range handlers {
		if handler != nil {
			handler(size)
		}
	}
}

// StopResizeMonitor stops the resize monitoring goroutine
func (ui *cliUI) StopResizeMonitor() {
	ui.resizeMu.Lock()
	defer ui.resizeMu.Unlock()
	if ui.resizeStop != nil {
		close(ui.resizeStop)
		ui.resizeStop = nil
	}
}

// --- Helper methods for cliUI ---

func (ui *cliUI) printMemoryLegend() {
	fmt.Print("  Legend: ")
	colorMemCode.Print("■")
	fmt.Print("=code ")
	colorMemData.Print("■")
	fmt.Print("=data ")
	colorMemStack.Print("■")
	fmt.Print("=stack ")
	colorMemUnknown.Print("■")
	fmt.Println("=unknown")
}

func (ui *cliUI) getRegionForAddress(addr uint32, regions []debugger.MemoryRegion) (debugger.MemoryRegionType, string) {
	for _, region := range regions {
		if addr >= region.StartAddr && addr < region.EndAddr {
			return region.RegionType, region.Name
		}
	}
	return debugger.RegionUnknown, ""
}

func (ui *cliUI) getRegionMarker(typ debugger.MemoryRegionType) string {
	switch typ {
	case debugger.RegionCode:
		return colorMemCode.Sprint("C")
	case debugger.RegionData:
		return colorMemData.Sprint("D")
	case debugger.RegionStack:
		return colorMemStack.Sprint("S")
	default:
		return colorMemUnknown.Sprint(".")
	}
}

func (ui *cliUI) printColoredByte(b byte, typ debugger.MemoryRegionType) {
	switch typ {
	case debugger.RegionCode:
		colorMemCode.Printf("%02X ", b)
	case debugger.RegionData:
		colorMemData.Printf("%02X ", b)
	case debugger.RegionStack:
		colorMemStack.Printf("%02X ", b)
	default:
		colorMemUnknown.Printf("%02X ", b)
	}
}

// colorizeInstruction applies syntax highlighting to an instruction string
func colorizeInstruction(instr string) string {
	instr = strings.TrimSpace(instr)
	if instr == "" {
		return instr
	}

	opcodeLoc := debugOpcodePattern.FindStringIndex(instr)
	if opcodeLoc == nil {
		return instr
	}

	opcode := instr[opcodeLoc[0]:opcodeLoc[1]]
	rest := instr[opcodeLoc[1]:]

	// Colorize registers
	rest = debugRegPattern.ReplaceAllStringFunc(rest, func(s string) string {
		return instrReg.Sprint(s)
	})

	// Colorize immediates
	rest = debugImmPattern.ReplaceAllStringFunc(rest, func(s string) string {
		return instrImm.Sprint(s)
	})

	return instrOpcode.Sprint(opcode) + rest
}

// =============================================================================
// Command definitions and flags
// =============================================================================

var (
	debugMemorySize    uint32
	debugVerbose       bool
	debugCompileFormat string
)

var debugCmd = &cobra.Command{
	Use:   "debug <program>",
	Short: "Run the Cucaracha debugger",
	Long: `Interactive debugger for Cucaracha CPU programs.

Commands:
  step, s [n]        - Execute n source lines (default: 1)
  next, n [n]        - Step over function calls (default: 1)
  stepi, si [n]      - Execute n instructions (default: 1)
  nexti, ni [n]      - Step over calls at instruction level (default: 1)
  continue, c        - Continue execution until breakpoint
  run, r             - Run until termination or breakpoint
  break, b [addr|file:line] - Set breakpoint (default: PC)
  watch, w <addr>    - Set watchpoint on memory address
  delete, d <id>     - Delete breakpoint/watchpoint by ID
  list, l            - List all breakpoints and watchpoints
  print, p <what>    - Print register or expression
  set <reg> <value>  - Set register value
  disasm, x [addr] [n] - Disassemble n instructions
  info, i            - Show CPU state
  stack              - Interactive stack memory view
  bt                 - Show backtrace (call stack)
  up [n]             - Move up n frames (to caller)
  down [n]           - Move down n frames (to callee)
  frame, f [n]       - Select frame n (or show current)
  memory, m <addr> [n] - Show memory contents
  source, src [n]    - Show source code
  vars, v            - Show variables (in selected frame)
  eval, e <expr>     - Evaluate expression
  exec, ex           - Interactive execution view (combined panels)
  help, h            - Show help
  quit, q            - Exit debugger`,
	Args: cobra.ExactArgs(1),
	Run:  runDebug,
}

func init() {
	CpuCmd.AddCommand(debugCmd)
	debugCmd.Flags().Uint32VarP(&debugMemorySize, "memory", "m", 0x20000, "Memory size in bytes (default: 128KB)")
	debugCmd.Flags().BoolVarP(&debugVerbose, "verbose", "v", false, "Print verbose output")
	debugCmd.Flags().StringVar(&debugCompileFormat, "compile-to", "object", "Compilation output format: assembly, object")
}

// =============================================================================
// Main debug entry point
// =============================================================================

func runDebug(cmd *cobra.Command, args []string) {
	inputPath := args[0]

	// Load the program
	loadOpts := &loader.Options{
		Verbose:      debugVerbose,
		OutputFormat: debugCompileFormat,
	}
	loadResult, err := loader.LoadFile(inputPath, loadOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading program: %v\n", err)
		os.Exit(1)
	}
	defer loadResult.Cleanup()

	if loadResult.WasCompiled {
		fmt.Printf("Compiled %s -> %s\n", inputPath, loadResult.CompiledPath)
	}

	resolved := loadResult.Program
	fmt.Printf("Loaded %d instructions from %s\n", len(resolved.Instructions()), loadResult.CompiledPath)

	// Create backend
	backend := debugger.NewBackend(debugMemorySize)
	if err := backend.LoadProgram(resolved); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading program: %v\n", err)
		os.Exit(2)
	}

	// Set up signal handler for Ctrl+C interrupt during execution
	// On Windows/MSYS2, we need to be careful about signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	defer signal.Stop(sigChan)

	// Goroutine to handle interrupt signals
	go func() {
		for range sigChan {
			// Signal the debugger to interrupt execution
			backend.Interrupt()
		}
	}()

	// Get debug info
	debugInfo := backend.DebugInfo()
	if debugInfo != nil {
		debugInfo.TryLoadSourceFiles()
		colorSuccess.Printf("Debug info: %d source locations\n", len(debugInfo.InstructionLocations))
		if debugVerbose {
			for path, lines := range debugInfo.SourceFiles {
				fmt.Printf("  Loaded source: %s (%d lines)\n", path, len(lines))
			}
		}
	}

	// Set up liner for readline support
	line := liner.NewLiner()
	defer line.Close()

	// Disable liner's Ctrl+C handling - we handle it ourselves
	// Note: On MSYS2/Git Bash, signal handling may not work as expected
	line.SetCtrlCAborts(false)
	line.SetMultiLineMode(false)

	// Set up tab completion
	line.SetCompleter(func(input string) []string {
		commands := []string{
			"step", "s", "next", "n", "stepi", "si", "nexti", "ni", "continue", "c", "run", "r",
			"break", "b", "watch", "w", "delete", "d", "list", "l",
			"print", "p", "set", "disasm", "x",
			"info", "i", "stack", "memory", "m", "bt", "backtrace",
			"up", "down", "frame", "f",
			"source", "src", "vars", "v", "eval", "e",
			"exec", "ex",
			"help", "h", "quit", "q", "exit",
		}
		var completions []string
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, strings.ToLower(input)) {
				completions = append(completions, cmd)
			}
		}
		return completions
	})

	// Load history
	historyFile := getHistoryFilePath()
	if f, err := os.Open(historyFile); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	// Create UI and controller
	ui := &cliUI{liner: line}
	controller := debugger.NewController(backend, ui)

	// Show initial state
	fmt.Printf("Entry point: %s\n", colorAddr.Sprintf("0x%08X", backend.GetState().PC))
	colorSuccess.Println("Type 'help' for available commands.")
	fmt.Println()

	// Show initial instruction
	showInitialInstruction(backend, ui)

	// Main loop
	for controller.IsRunning() {
		input, err := line.Prompt("(cucaracha) ")
		if err != nil {
			if err == io.EOF {
				colorSuccess.Println("\nExiting debugger.")
				break
			}
			// Ctrl+C at prompt - tell user to use quit command
			if err == liner.ErrPromptAborted {
				fmt.Println()
				colorWarning.Println("Use 'quit' or 'exit' to leave the debugger.")
				continue
			}
			colorError.Printf("Error reading input: %v\n", err)
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			input = controller.LastCommand()
		}
		if input != "" {
			if input != controller.LastCommand() {
				line.AppendHistory(input)
			}
			controller.SetLastCommand(input)
			// Take over signal handling during command execution
			signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
			executeCommand(controller, input)
		}
	}

	// Save history
	if f, err := os.Create(historyFile); err == nil {
		line.WriteHistory(f)
		f.Close()
	}
}

// showInitialInstruction displays the initial instruction without stepping
func showInitialInstruction(backend *debugger.Backend, ui *cliUI) {
	instructions, err := backend.Disassemble(backend.GetState().PC, 1)
	if err == nil && len(instructions) > 0 {
		ui.ShowInstruction(instructions[0])
	}
}

// getHistoryFilePath returns the path to the debugger history file
func getHistoryFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".cucaracha_history"
	}
	return filepath.Join(homeDir, ".cucaracha_history")
}

// =============================================================================
// Command parsing and dispatching
// =============================================================================

func executeCommand(c *debugger.Controller, line string) {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return
	}

	cmd := strings.ToLower(parts[0])
	args := parts[1:]

	switch cmd {
	case "step", "s":
		count := 1
		if len(args) > 0 {
			if n, err := strconv.Atoi(args[0]); err == nil && n > 0 {
				count = n
			}
		}
		c.CmdStep(count)

	case "next", "n":
		count := 1
		if len(args) > 0 {
			if n, err := strconv.Atoi(args[0]); err == nil && n > 0 {
				count = n
			}
		}
		c.CmdNext(count)

	case "nexti", "ni":
		count := 1
		if len(args) > 0 {
			if n, err := strconv.Atoi(args[0]); err == nil && n > 0 {
				count = n
			}
		}
		c.CmdInstructionNext(count)

	case "stepi", "si":
		count := 1
		if len(args) > 0 {
			if n, err := strconv.Atoi(args[0]); err == nil && n > 0 {
				count = n
			}
		}
		c.CmdInstructionStep(count)

	case "continue", "c":
		c.CmdContinue()

	case "run", "r":
		c.CmdRun()

	case "break", "b":
		var addr uint32
		if len(args) == 0 {
			// No argument: use current PC
			addr = c.Backend().GetState().PC
		} else {
			var err error
			addr, err = c.ResolveAddress(args[0])
			if err != nil {
				c.UI().ShowMessage(debugger.LevelError, "Invalid address: %s", args[0])
				return
			}
		}
		c.CmdBreak(addr)

	case "watch", "w":
		if len(args) == 0 {
			c.UI().ShowMessage(debugger.LevelError, "Usage: watch <address>")
			return
		}
		addr, err := c.ResolveAddress(args[0])
		if err != nil {
			c.UI().ShowMessage(debugger.LevelError, "Invalid address: %s", args[0])
			return
		}
		c.CmdWatch(addr)

	case "delete", "d":
		if len(args) == 0 {
			c.UI().ShowMessage(debugger.LevelError, "Usage: delete <id>")
			return
		}
		id, err := strconv.Atoi(args[0])
		if err != nil {
			c.UI().ShowMessage(debugger.LevelError, "Invalid ID: %s", args[0])
			return
		}
		c.CmdDelete(id)

	case "list", "l":
		c.CmdList()

	case "print", "p":
		if len(args) == 0 {
			c.UI().ShowMessage(debugger.LevelError, "Usage: print <register|expression>")
			return
		}
		c.CmdPrint(strings.Join(args, " "))

	case "set":
		if len(args) < 2 {
			c.UI().ShowMessage(debugger.LevelError, "Usage: set <register> <value>")
			return
		}
		value, err := parseValue(args[1])
		if err != nil {
			c.UI().ShowMessage(debugger.LevelError, "Invalid value: %s", args[1])
			return
		}
		c.CmdSet(args[0], value)

	case "disasm", "x":
		addr := c.Backend().GetState().PC
		count := -1 // -1 means interactive mode
		if len(args) > 0 {
			if a, err := c.ResolveAddress(args[0]); err == nil {
				addr = a
			}
		}
		if len(args) > 1 {
			if n, err := strconv.Atoi(args[1]); err == nil && n > 0 {
				count = n
			}
		}
		if count > 0 {
			c.CmdDisasm(addr, count)
		} else {
			interactiveDisassemblyView(c, addr)
		}

	case "info", "i":
		c.CmdInfo()

	case "stack":
		// Interactive stack view - restricted to stack region (SP to end of memory)
		state := c.Backend().GetState()
		memSize := uint32(len(c.Backend().Runner().State().Memory))
		bounds := &MemoryViewBounds{
			MinAddr: state.SP,
			MaxAddr: memSize - 4,
			Title:   "Stack View",
		}
		interactiveMemoryViewWithBounds(c, state.SP, bounds)

	case "bt", "backtrace":
		c.CmdBacktrace()

	case "up":
		count := 1
		if len(args) > 0 {
			if n, err := strconv.Atoi(args[0]); err == nil && n > 0 {
				count = n
			}
		}
		c.CmdUp(count)

	case "down":
		count := 1
		if len(args) > 0 {
			if n, err := strconv.Atoi(args[0]); err == nil && n > 0 {
				count = n
			}
		}
		c.CmdDown(count)

	case "frame", "f":
		if len(args) == 0 {
			// Show current frame
			frames := c.Backend().GetCallStack()
			if len(frames) > 0 {
				frame := frames[c.SelectedFrame()]
				funcName := frame.Function
				if funcName == "" {
					funcName = "??"
				}
				c.UI().ShowMessage(debugger.LevelInfo, "#%d  %s [0x%08X]", c.SelectedFrame(), funcName, frame.Address)
			}
			return
		}
		frameNum, err := strconv.Atoi(args[0])
		if err != nil {
			c.UI().ShowMessage(debugger.LevelError, "Invalid frame number: %s", args[0])
			return
		}
		c.CmdFrame(frameNum)

	case "memory", "m":
		if len(args) == 0 {
			// Interactive mode - start at code section
			program := c.Backend().Program()
			layout := program.MemoryLayout()
			interactiveMemoryView(c, layout.CodeStart)
			return
		}
		// Parse: memory <expr>[, count]
		fullArg := strings.Join(args, " ")
		expr := fullArg
		count := -1 // -1 means interactive mode

		if commaIdx := strings.LastIndex(fullArg, ","); commaIdx != -1 {
			countStr := strings.TrimSpace(fullArg[commaIdx+1:])
			if n, err := strconv.Atoi(countStr); err == nil && n > 0 {
				count = n
				expr = strings.TrimSpace(fullArg[:commaIdx])
			}
		}

		addr, err := c.ResolveAddress(expr)
		if err != nil {
			c.UI().ShowMessage(debugger.LevelError, "Invalid address expression: %v", err)
			return
		}

		if count > 0 {
			c.CmdMemory(addr, count)
		} else {
			interactiveMemoryView(c, addr)
		}

	case "source", "src":
		lines := 10
		if len(args) > 0 {
			if n, err := strconv.Atoi(args[0]); err == nil && n > 0 {
				lines = n
			}
		}
		c.CmdSource(lines)

	case "vars", "v":
		c.CmdVars()

	case "eval", "e":
		if len(args) == 0 {
			c.UI().ShowMessage(debugger.LevelError, "Usage: eval <expression>")
			return
		}
		c.CmdEval(strings.Join(args, " "))

	case "exec", "ex":
		// Interactive execution display
		interactiveExecutionView(c)

	case "help", "h", "?":
		c.CmdHelp()

	case "quit", "q", "exit":
		c.CmdQuit()

	default:
		c.UI().ShowMessage(debugger.LevelError, "Unknown command: %s. Type 'help' for available commands.", cmd)
	}
}

func showMemoryUsage(c *debugger.Controller) {
	state := c.Backend().GetState()
	program := c.Backend().Program()
	layout := program.MemoryLayout()

	fmt.Println("Usage: memory <expr> [, count]")
	fmt.Println("  <expr> can be: address (0x...), register (sp, pc), symbol, or expression")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Printf("  %s       - Memory at hex address\n", colorAddr.Sprint("memory 0x10000"))
	fmt.Printf("  %s            - Memory at stack pointer\n", colorReg.Sprint("memory sp"))
	fmt.Printf("  %s  - 128 bytes at sp\n", colorReg.Sprint("memory sp, 128"))
	fmt.Println()
	fmt.Println("Useful addresses:")
	fmt.Printf("  Code section:   %s - %s\n",
		colorAddr.Sprintf("0x%08X", layout.CodeStart),
		colorAddr.Sprintf("0x%08X", layout.CodeStart+layout.CodeSize))
	if layout.DataSize > 0 {
		fmt.Printf("  Data section:   %s - %s\n",
			colorAddr.Sprintf("0x%08X", layout.DataStart),
			colorAddr.Sprintf("0x%08X", layout.DataStart+layout.DataSize))
	}
	fmt.Printf("  Stack (SP):     %s\n", colorAddr.Sprintf("0x%08X", state.SP))
	fmt.Printf("  Current PC:     %s\n", colorAddr.Sprintf("0x%08X", state.PC))
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

// =============================================================================
// Interactive Memory View
// =============================================================================

const (
	memViewBytesPerLine    = 16
	memViewHeaderLines     = 3 // Header + help + blank line
	memViewFooterLines     = 3 // Blank + SP/PC line + buffer
	memViewMinContentLines = 4 // Minimum content lines to show
)

// getTerminalSize returns terminal width and height, with fallback defaults
func getTerminalSize() (width, height int) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 || height <= 0 {
		// Fallback to reasonable defaults
		return 80, 24
	}
	return width, height
}

// calculateMemViewLines returns how many content lines fit in the terminal
func calculateMemViewLines() int {
	_, height := getTerminalSize()
	contentLines := height - memViewHeaderLines - memViewFooterLines
	if contentLines < memViewMinContentLines {
		contentLines = memViewMinContentLines
	}
	return contentLines
}

// MemoryViewBounds defines the address range for an interactive memory view
type MemoryViewBounds struct {
	MinAddr uint32 // Minimum viewable address (inclusive)
	MaxAddr uint32 // Maximum viewable address (inclusive, will be aligned)
	Title   string // Optional title for the view (e.g., "Stack View")
}

// interactiveMemoryView displays memory with keyboard navigation
func interactiveMemoryView(c *debugger.Controller, startAddr uint32) {
	interactiveMemoryViewWithBounds(c, startAddr, nil)
}

// interactiveMemoryViewWithBounds displays memory with keyboard navigation, optionally restricted to bounds
func interactiveMemoryViewWithBounds(c *debugger.Controller, startAddr uint32, bounds *MemoryViewBounds) {
	// Get memory size for bounds checking
	memSize := uint32(len(c.Backend().Runner().State().Memory))
	if memSize == 0 {
		c.UI().ShowMessage(debugger.LevelError, "No memory available")
		return
	}

	// Determine effective bounds
	minAddr := uint32(0)
	maxAddr := memSize - 4
	viewTitle := "Memory View"
	if bounds != nil {
		minAddr = bounds.MinAddr &^ 0x3 // Align to word boundary
		if bounds.MaxAddr < memSize {
			maxAddr = bounds.MaxAddr &^ 0x3
		}
		if bounds.Title != "" {
			viewTitle = bounds.Title
		}
	}

	// Align to word boundary and clamp to valid range
	addr := startAddr &^ 0x3
	if addr < minAddr {
		addr = minAddr
	}
	if addr > maxAddr {
		addr = maxAddr
	}

	// Save terminal state and enable raw mode
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		c.UI().ShowMessage(debugger.LevelError, "Cannot enable raw terminal mode: %v", err)
		return
	}
	defer term.Restore(fd, oldState)

	// Set up resize detection
	resizeChan := make(chan debugger.TerminalSize, 1)
	unregisterResize := c.UI().OnResize(func(size debugger.TerminalSize) {
		select {
		case resizeChan <- size:
		default:
			// Channel full, skip this resize event
		}
	})
	defer unregisterResize()

	// Clear screen and show initial view
	showInteractiveMemoryWithBounds(c, addr, memSize, minAddr, maxAddr, viewTitle)

	// Read keyboard input with resize detection
	buf := make([]byte, 8)
	inputChan := make(chan []byte, 1)
	done := make(chan struct{})
	go func() {
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				close(inputChan)
				return
			}
			key := make([]byte, n)
			copy(key, buf[:n])
			select {
			case inputChan <- key:
			case <-done:
				return
			}
		}
	}()

	cleanup := func() {
		close(done)
		// Drain any pending input to prevent the zombie goroutine from
		// blocking when it tries to send to inputChan
		select {
		case <-inputChan:
		default:
		}
	}

	for {
		var key []byte
		select {
		case <-resizeChan:
			// Terminal resized, redraw
			showInteractiveMemoryWithBounds(c, addr, memSize, minAddr, maxAddr, viewTitle)
			continue
		case k, ok := <-inputChan:
			if !ok {
				cleanup()
				return
			}
			key = k
		}

		newAddr := addr

		// Parse key sequences
		switch {
		case len(key) == 1 && (key[0] == 'q' || key[0] == 'Q' || key[0] == 27): // q, Q, or ESC
			// Exit
			fmt.Print("\r\n")
			cleanup()
			return

		case len(key) == 3 && key[0] == 27 && key[1] == '[':
			// Arrow keys, Home, End
			switch key[2] {
			case 'A': // Up arrow - move up 1 word (4 bytes)
				if addr >= minAddr+4 {
					newAddr = addr - 4
				}
			case 'B': // Down arrow - move down 1 word (4 bytes)
				if addr+4 <= maxAddr {
					newAddr = addr + 4
				}
			case 'H': // Home key - go to start
				newAddr = minAddr
			case 'F': // End key - go to end
				newAddr = maxAddr &^ 0x3
			case '5': // Page Up - need to read one more byte
				// This is actually incomplete - PgUp sends ESC [ 5 ~
			case '6': // Page Down - need to read one more byte
				// This is actually incomplete - PgDn sends ESC [ 6 ~
			}

		case len(key) == 4 && key[0] == 27 && key[1] == '[':
			// Page Up/Down, Home/End (ESC [ N ~)
			if key[3] == '~' {
				switch key[2] {
				case '1': // Home key (alternate) - go to start
					newAddr = minAddr
				case '4': // End key (alternate) - go to end
					newAddr = maxAddr &^ 0x3
				case '5': // Page Up - move up 100 words (400 bytes)
					if addr >= minAddr+400 {
						newAddr = addr - 400
					} else {
						newAddr = minAddr
					}
				case '6': // Page Down - move down 100 words (400 bytes)
					if addr+400 <= maxAddr {
						newAddr = addr + 400
					} else {
						newAddr = maxAddr &^ 0x3
					}
				}
			}

		case len(key) == 6 && key[0] == 27 && key[1] == '[' && key[2] == '1' && key[3] == ';' && key[4] == '2':
			// Shift+Arrow keys (ESC [ 1 ; 2 A/B)
			switch key[5] {
			case 'A': // Shift+Up - move up 10 words (40 bytes)
				if addr >= minAddr+40 {
					newAddr = addr - 40
				} else {
					newAddr = minAddr
				}
			case 'B': // Shift+Down - move down 10 words (40 bytes)
				if addr+40 <= maxAddr {
					newAddr = addr + 40
				} else {
					newAddr = maxAddr &^ 0x3
				}
			}

		case len(key) == 1 && key[0] == 'G': // G - go to stack pointer
			sp := c.Backend().GetState().SP &^ 0x3
			if sp >= minAddr && sp <= maxAddr {
				newAddr = sp
			}

		case len(key) == 1 && key[0] == '?': // ? - show help
			showMemoryViewHelp()
			// Wait for any key
			os.Stdin.Read(buf)
			showInteractiveMemoryWithBounds(c, addr, memSize, minAddr, maxAddr, viewTitle)
			continue
		}

		if newAddr != addr {
			addr = newAddr
			showInteractiveMemoryWithBounds(c, addr, memSize, minAddr, maxAddr, viewTitle)
		}
	}
}

// showInteractiveMemory displays the memory view at the given address (full memory range)
func showInteractiveMemory(c *debugger.Controller, addr uint32, memSize uint32) {
	showInteractiveMemoryWithBounds(c, addr, memSize, 0, memSize-4, "Memory View")
}

// showInteractiveMemoryWithBounds displays the memory view at the given address with restricted bounds
func showInteractiveMemoryWithBounds(c *debugger.Controller, addr uint32, memSize uint32, minAddr, maxAddr uint32, title string) {
	// Move cursor to top-left and clear entire screen
	fmt.Print("\033[H\033[2J")

	// Get terminal size and calculate how many lines we can show
	termWidth, _ := getTerminalSize()
	contentLines := calculateMemViewLines()

	// Header (use \r\n for raw terminal mode)
	colorHeader.Printf("%s - %s", title, colorAddr.Sprintf("0x%08X", addr))
	colorHiBlack.Printf(" (range: 0x%X - 0x%X)\r\n", minAddr, maxAddr)
	colorHiBlack.Print("↑/↓: ±1 word | Shift+↑/↓: ±10 words | PgUp/PgDn: ±100 words | Home/End | G: SP | q/ESC: exit\r\n")
	fmt.Print("\r\n")

	// Get regions and memory layout info first to determine view mode
	regions := c.Backend().GetMemoryRegions()
	state := c.Backend().GetState()
	program := c.Backend().Program()
	layout := program.MemoryLayout()

	// Determine the dominant region type for this view
	mainRegion, _ := getRegionForAddressStatic(addr, regions)

	// Calculate read size based on view mode and terminal size
	var readSize uint32
	if mainRegion == debugger.RegionCode && layout != nil {
		// Code view: 4 bytes per line (1 instruction)
		readSize = uint32(contentLines * 4)
	} else {
		// Data view: 16 bytes per line
		readSize = uint32(contentLines * memViewBytesPerLine)
	}

	if addr+readSize > memSize {
		readSize = memSize - addr
	}
	if readSize == 0 {
		colorWarning.Printf("Address 0x%08X is at end of memory\r\n", addr)
		return
	}

	// Read memory
	data, err := c.Backend().ReadMemory(addr, int(readSize))
	if err != nil {
		colorError.Printf("Cannot read memory at 0x%08X: %v\r\n", addr, err)
		return
	}

	// Build a map of addresses to variable/symbol info for quick lookup
	addrAnnotations := buildAddressAnnotations(c, regions, state)

	// Display memory based on region type
	if mainRegion == debugger.RegionCode && layout != nil {
		// Code section: show instruction-aligned view (4 bytes per line = 1 instruction)
		showCodeMemoryView(c, addr, data, regions, state, layout, addrAnnotations, contentLines, termWidth)
	} else {
		// Data/Stack/Unknown: show standard hex dump with annotations
		showDataMemoryView(c, addr, data, regions, state, addrAnnotations, contentLines, termWidth)
	}

	// Footer with context (no trailing \r\n on last line to prevent scroll)
	fmt.Print("\r\n")
	fmt.Printf("SP: %s  PC: %s",
		colorAddr.Sprintf("0x%08X", state.SP),
		colorAddr.Sprintf("0x%08X", state.PC))
}

// addressAnnotation contains contextual info for a memory address
type addressAnnotation struct {
	label      string // e.g., "counter (int)", "loop:", "main"
	regionType debugger.MemoryRegionType
	isPC       bool   // Current program counter
	isBP       bool   // Has breakpoint
	colorIdx   int    // Color index for variable coloring (-1 = no color)
	size       uint32 // Size of the variable in bytes
}

// variableColorMap tracks assigned colors for variables
type variableColorMap struct {
	nameToColor map[string]int
	nextColor   int
}

func newVariableColorMap() *variableColorMap {
	return &variableColorMap{
		nameToColor: make(map[string]int),
		nextColor:   0,
	}
}

func (m *variableColorMap) getColor(name string) int {
	if idx, ok := m.nameToColor[name]; ok {
		return idx
	}
	idx := m.nextColor % len(variableColorPalette)
	m.nameToColor[name] = idx
	m.nextColor++
	return idx
}

func getVariableColor(colorIdx int) *color.Color {
	if colorIdx < 0 || colorIdx >= len(variableColorPalette) {
		return colorHiBlack
	}
	return variableColorPalette[colorIdx]
}

// buildAddressAnnotations creates a map of address annotations
func buildAddressAnnotations(c *debugger.Controller, regions []debugger.MemoryRegion, state debugger.DebuggerState) map[uint32]addressAnnotation {
	annotations := make(map[uint32]addressAnnotation)
	colorMap := newVariableColorMap()

	// Add variable annotations from regions - annotate ALL bytes of each variable
	for _, r := range regions {
		// Only annotate specific variable/symbol regions
		if strings.HasPrefix(r.Name, "global:") || strings.Contains(r.Name, "(") {
			varSize := r.EndAddr - r.StartAddr
			colorIdx := colorMap.getColor(r.Name)

			// Mark all bytes belonging to this variable with its color
			for offset := uint32(0); offset < varSize; offset++ {
				byteAddr := r.StartAddr + offset
				ann := addressAnnotation{
					regionType: r.RegionType,
					colorIdx:   colorIdx,
					size:       varSize,
				}
				// Only set label on the first byte
				if offset == 0 {
					ann.label = r.Name
				}
				annotations[byteAddr] = ann
			}
		}
	}

	// Add function/symbol annotations
	program := c.Backend().Program()
	if funcs := program.Functions(); funcs != nil {
		for name, fn := range funcs {
			if len(fn.InstructionRanges) > 0 {
				for _, rng := range fn.InstructionRanges {
					instrs := program.Instructions()
					if rng.Start < len(instrs) && instrs[rng.Start].Address != nil {
						fnAddr := *instrs[rng.Start].Address
						annotations[fnAddr] = addressAnnotation{
							label:      name + ":",
							regionType: debugger.RegionCode,
						}
					}
				}
			}
		}
	}

	// Mark current PC and breakpoints
	pc := state.PC
	if ann, ok := annotations[pc]; ok {
		ann.isPC = true
		annotations[pc] = ann
	} else {
		annotations[pc] = addressAnnotation{isPC: true}
	}

	for _, bp := range c.Backend().ListBreakpoints() {
		if ann, ok := annotations[bp.Address]; ok {
			ann.isBP = true
			annotations[bp.Address] = ann
		} else {
			annotations[bp.Address] = addressAnnotation{isBP: true}
		}
	}

	return annotations
}

// showCodeMemoryView displays memory as decoded instructions
func showCodeMemoryView(c *debugger.Controller, addr uint32, data []byte, regions []debugger.MemoryRegion, state debugger.DebuggerState, layout *mc.MemoryLayout, annotations map[uint32]addressAnnotation, maxLines int, termWidth int) {
	// Code view: 4 bytes per line (1 instruction)
	// Multi-line format with aligned sub-lines for source info
	const indentWidth = 14 // "  " + "0x00000000: " alignment

	codeEnd := layout.CodeStart + layout.CodeSize
	linesShown := 0

	for i := 0; i < len(data) && linesShown < maxLines; i += 4 {
		lineAddr := addr + uint32(i)

		// Check if we're still in code section
		inCode := lineAddr >= layout.CodeStart && lineAddr < codeEnd

		// Get annotation for this address
		ann := annotations[lineAddr]

		// Print marker and address
		var marker string
		if ann.isPC {
			marker = colorPC.Sprint("►")
		} else if ann.isBP {
			marker = colorBreakpoint.Sprint("●")
		} else if inCode {
			marker = colorMemCode.Sprint("C")
		} else {
			regionType, _ := getRegionForAddressStatic(lineAddr, regions)
			marker = getRegionMarkerStatic(regionType)
		}
		fmt.Printf("%s ", marker)
		fmt.Printf("%s: ", colorAddr.Sprintf("0x%08X", lineAddr))

		// Print raw bytes (4 bytes for instruction)
		if i+4 <= len(data) {
			for j := 0; j < 4; j++ {
				if inCode {
					colorMemCode.Printf("%02X ", data[i+j])
				} else {
					regionType, _ := getRegionForAddressStatic(lineAddr+uint32(j), regions)
					printColoredByteStatic(data[i+j], regionType)
				}
			}
		} else {
			// Partial read at end
			for j := 0; i+j < len(data); j++ {
				fmt.Printf("%02X ", data[i+j])
			}
			for j := len(data) - i; j < 4; j++ {
				fmt.Print("   ")
			}
		}

		// Decode and show instruction if in code section
		if inCode && i+4 <= len(data) {
			word := binary.LittleEndian.Uint32(data[i : i+4])
			instrText := decodeInstructionWord(word)
			fmt.Printf(" │ ")
			printHighlightedInstruction(instrText)
		}

		// Show annotation if present (function labels, etc.) on same line
		if ann.label != "" {
			colorHiBlack.Printf("  ; %s", ann.label)
		}

		fmt.Print("\r\n")
		linesShown++

		// Show source info on additional lines if available
		if inCode && i+4 <= len(data) {
			if srcLoc := c.Backend().GetSourceLocation(lineAddr); srcLoc != nil {
				srcFileName := filepath.Base(srcLoc.File)

				// Get source text
				var srcText string
				if debugInfo := c.Backend().DebugInfo(); debugInfo != nil {
					srcText = strings.TrimSpace(debugInfo.GetSourceLine(srcLoc.File, srcLoc.Line))
				}

				// Print source location line
				if linesShown < maxLines {
					fmt.Printf("%*s", indentWidth, "")
					if srcText != "" {
						colorHiBlack.Print("├─ ")
					} else {
						colorHiBlack.Print("└─ ")
					}
					colorSourceFile.Printf("%s", srcFileName)
					colorHiBlack.Print(":")
					colorSourceLine.Printf("%d", srcLoc.Line)
					fmt.Print("\r\n")
					linesShown++
				}

				// Print source text line
				if srcText != "" && linesShown < maxLines {
					fmt.Printf("%*s", indentWidth, "")
					colorHiBlack.Print("└─ ")
					// Truncate if too long
					maxSrcWidth := termWidth - indentWidth - 3
					if len(srcText) > maxSrcWidth && maxSrcWidth > 3 {
						srcText = srcText[:maxSrcWidth-3] + "..."
					}
					fmt.Print(utils.HighlightCCode(srcText) + "\r\n")
					linesShown++
				}
			}
		}
	}
}

// Width constants for data memory view line components
const (
	dataMemViewIndent = 14 // Alignment indent for sub-lines (matches "  0x00000000: ")
)

// showDataMemoryView displays memory as hex dump with variable annotations
func showDataMemoryView(c *debugger.Controller, addr uint32, data []byte, regions []debugger.MemoryRegion, state debugger.DebuggerState, annotations map[uint32]addressAnnotation, maxLines int, termWidth int) {
	// Standard 16-byte per line hex dump with multi-line format for additional info
	linesShown := 0
	for i := 0; i < len(data) && linesShown < maxLines; i += memViewBytesPerLine {
		lineAddr := addr + uint32(i)
		regionType, regionName := getRegionForAddressStatic(lineAddr, regions)

		// Collect all variable annotations on this line with their offsets
		type varAnnotation struct {
			offset   int
			label    string
			colorIdx int
		}
		var lineVars []varAnnotation
		for j := 0; j < memViewBytesPerLine && i+j < len(data); j++ {
			byteAddr := lineAddr + uint32(j)
			if ann, ok := annotations[byteAddr]; ok && ann.label != "" {
				lineVars = append(lineVars, varAnnotation{
					offset:   j,
					label:    ann.label,
					colorIdx: ann.colorIdx,
				})
			}
		}

		// === Line 1: Main hex dump line ===
		marker := getRegionMarkerStatic(regionType)
		fmt.Printf("%s ", marker)
		fmt.Printf("%s: ", colorAddr.Sprintf("0x%08X", lineAddr))

		// Hex bytes - use variable colors when applicable
		for j := 0; j < memViewBytesPerLine && i+j < len(data); j++ {
			byteAddr := lineAddr + uint32(j)
			if ann, ok := annotations[byteAddr]; ok && ann.colorIdx >= 0 {
				varColor := getVariableColor(ann.colorIdx)
				varColor.Printf("%02X ", data[i+j])
			} else {
				byteRegion, _ := getRegionForAddressStatic(byteAddr, regions)
				printColoredByteStatic(data[i+j], byteRegion)
			}
		}
		// Padding for incomplete lines
		for j := len(data) - i; j < memViewBytesPerLine; j++ {
			fmt.Print("   ")
		}
		// ASCII column - also colored by variable
		fmt.Print(" │")
		for j := 0; j < memViewBytesPerLine && i+j < len(data); j++ {
			b := data[i+j]
			byteAddr := lineAddr + uint32(j)
			var ch string
			if b >= 32 && b < 127 {
				ch = string(b)
			} else {
				ch = "."
			}
			if ann, ok := annotations[byteAddr]; ok && ann.colorIdx >= 0 {
				varColor := getVariableColor(ann.colorIdx)
				varColor.Print(ch)
			} else {
				fmt.Print(ch)
			}
		}
		fmt.Print("│\r\n")
		linesShown++

		// Determine what additional lines we need
		hasWordValues := (regionType == debugger.RegionStack || regionType == debugger.RegionData) && i+4 <= len(data)
		hasVarAnnotations := len(lineVars) > 0
		hasRegionName := regionName != "" && !strings.HasPrefix(regionName, "code") && !strings.HasPrefix(regionName, "data") && !strings.HasPrefix(regionName, "stack")

		// === Line 2: Word values (for data/stack regions) ===
		if hasWordValues && linesShown < maxLines {
			fmt.Printf("%*s", dataMemViewIndent, "")
			if hasVarAnnotations || hasRegionName {
				colorHiBlack.Print("├─ W: ")
			} else {
				colorHiBlack.Print("└─ W: ")
			}
			for j := 0; j < memViewBytesPerLine && i+j+4 <= len(data); j += 4 {
				word := binary.LittleEndian.Uint32(data[i+j : i+j+4])
				// Color word value if it belongs to a variable
				byteAddr := lineAddr + uint32(j)
				if ann, ok := annotations[byteAddr]; ok && ann.colorIdx >= 0 {
					varColor := getVariableColor(ann.colorIdx)
					varColor.Printf("0x%08X ", word)
				} else {
					colorHiBlack.Printf("0x%08X ", word)
				}
			}
			fmt.Print("\r\n")
			linesShown++
		}

		// === Line 3: Variable annotations ===
		if hasVarAnnotations && linesShown < maxLines {
			fmt.Printf("%*s", dataMemViewIndent, "")
			if hasRegionName {
				colorHiBlack.Print("├─ V: ")
			} else {
				colorHiBlack.Print("└─ V: ")
			}
			maxVarWidth := termWidth - dataMemViewIndent - 6 // 6 for "└─ V: "
			usedWidth := 0
			for idx, v := range lineVars {
				varColor := getVariableColor(v.colorIdx)
				varText := fmt.Sprintf("@%X:%s", v.offset, v.label)
				separator := ""
				if idx > 0 {
					separator = ", "
				}
				totalLen := len(separator) + len(varText)
				if usedWidth+totalLen <= maxVarWidth {
					if idx > 0 {
						colorHiBlack.Print(", ")
					}
					varColor.Printf("@%X:%s", v.offset, v.label)
					usedWidth += totalLen
				} else if usedWidth+4 <= maxVarWidth {
					colorHiBlack.Print("...")
					break
				}
			}
			fmt.Print("\r\n")
			linesShown++
		}

		// === Line 4: Region name (if special) ===
		if hasRegionName && linesShown < maxLines {
			fmt.Printf("%*s", dataMemViewIndent, "")
			colorHiBlack.Print("└─ R: ")
			maxNameWidth := termWidth - dataMemViewIndent - 6
			if len(regionName) > maxNameWidth && maxNameWidth > 3 {
				colorHiBlack.Printf("%s...", regionName[:maxNameWidth-3])
			} else {
				colorHiBlack.Print(regionName)
			}
			fmt.Print("\r\n")
			linesShown++
		}
	}
}

// decodeInstructionWord decodes a 32-bit word as an instruction
func decodeInstructionWord(word uint32) string {
	decoded, err := instructions.Instructions.Decode(word)
	if err != nil {
		return fmt.Sprintf("??? (0x%08X)", word)
	}

	mnemonic := decoded.Descriptor.OpCode.Mnemonic
	if len(decoded.OperandValues) == 0 {
		return mnemonic
	}

	parts := make([]string, 0, len(decoded.OperandValues))
	for _, op := range decoded.OperandValues {
		parts = append(parts, op.String())
	}
	return mnemonic + " " + strings.Join(parts, ", ")
}

// printHighlightedInstruction prints an instruction with syntax highlighting
func printHighlightedInstruction(text string) {
	// Apply highlighting patterns
	result := text

	// First highlight the opcode (at the start)
	if idx := strings.Index(result, " "); idx > 0 {
		opcode := result[:idx]
		rest := result[idx:]
		instrOpcode.Print(opcode)
		// Highlight registers and immediates in the rest
		rest = debugRegPattern.ReplaceAllStringFunc(rest, func(s string) string {
			return instrReg.Sprint(s)
		})
		rest = debugImmPattern.ReplaceAllStringFunc(rest, func(s string) string {
			return instrImm.Sprint(s)
		})
		fmt.Print(rest)
	} else {
		// Just opcode, no operands
		instrOpcode.Print(result)
	}
}

// printHighlightedInstructionLen prints a highlighted instruction and returns its display length
func printHighlightedInstructionLen(text string) int {
	printHighlightedInstruction(text)
	return len(text)
}

// =============================================================================
// Interactive Disassembly View
// =============================================================================

// cfgInstrData holds instruction data for CFG analysis
type cfgInstrData struct {
	addr          uint32
	mnemonic      string
	operands      string
	isPC          bool
	isBP          bool
	srcFile       string
	srcLine       int
	srcText       string
	isBranch      bool   // True if this is a branch instruction
	branchTarget  uint32 // Target address of branch (0 if unknown)
	branchTargetS string // Symbol name of branch target
	label         string // Label/symbol at this address (if any)
}

// cfgEdge represents a control flow edge in the CFG visualization
type cfgEdge struct {
	startIdx   int          // Index of branch instruction
	endIdx     int          // Index of target instruction (-1 if outside view)
	column     int          // Column for drawing the line
	isBackward bool         // Loop (backward branch)
	destAddr   uint32       // Target address
	color      *color.Color // Color for this edge
}

// cfgEdgeColors is a palette of colors for different branch edges
var cfgEdgeColors = []*color.Color{
	color.New(color.FgHiCyan),
	color.New(color.FgHiMagenta),
	color.New(color.FgHiYellow),
	color.New(color.FgHiGreen),
	color.New(color.FgHiBlue),
	color.New(color.FgHiRed),
	color.New(color.FgCyan),
	color.New(color.FgMagenta),
	color.New(color.FgYellow),
	color.New(color.FgGreen),
	color.New(color.FgBlue),
	color.New(color.FgRed),
}

// interactiveDisassemblyView displays disassembly with keyboard navigation
func interactiveDisassemblyView(c *debugger.Controller, startAddr uint32) {
	// Get code bounds
	program := c.Backend().Program()
	layout := program.MemoryLayout()
	if layout == nil {
		c.UI().ShowMessage(debugger.LevelError, "No program loaded")
		return
	}

	codeStart := layout.CodeStart
	codeEnd := layout.CodeStart + layout.CodeSize

	// Align to instruction boundary and clamp to code range
	addr := startAddr &^ 0x3
	if addr < codeStart {
		addr = codeStart
	}
	if addr >= codeEnd {
		addr = (codeEnd - 4) &^ 0x3
	}

	// Save terminal state and enable raw mode
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		c.UI().ShowMessage(debugger.LevelError, "Cannot enable raw terminal mode: %v", err)
		return
	}
	defer term.Restore(fd, oldState)

	// Set up resize detection
	resizeChan := make(chan debugger.TerminalSize, 1)
	unregisterResize := c.UI().OnResize(func(size debugger.TerminalSize) {
		select {
		case resizeChan <- size:
		default:
			// Channel full, skip this resize event
		}
	})
	defer unregisterResize()

	// Clear screen and show initial view
	showInteractiveDisassembly(c, addr, codeStart, codeEnd)

	// Read keyboard input with resize detection
	buf := make([]byte, 8)
	inputChan := make(chan []byte, 1)
	done := make(chan struct{})
	go func() {
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				close(inputChan)
				return
			}
			key := make([]byte, n)
			copy(key, buf[:n])
			select {
			case inputChan <- key:
			case <-done:
				return
			}
		}
	}()

	cleanup := func() {
		close(done)
		// Drain any pending input to prevent the zombie goroutine from
		// blocking when it tries to send to inputChan
		select {
		case <-inputChan:
		default:
		}
	}

	for {
		var key []byte
		select {
		case <-resizeChan:
			// Terminal resized, redraw
			showInteractiveDisassembly(c, addr, codeStart, codeEnd)
			continue
		case k, ok := <-inputChan:
			if !ok {
				cleanup()
				return
			}
			key = k
		}

		newAddr := addr

		// Calculate max address
		maxAddr := codeEnd - 4
		if maxAddr < codeStart {
			maxAddr = codeStart
		}

		// Parse key sequences
		switch {
		case len(key) == 1 && (key[0] == 'q' || key[0] == 'Q' || key[0] == 27): // q, Q, or ESC
			fmt.Print("\r\n")
			cleanup()
			return

		case len(key) == 3 && key[0] == 27 && key[1] == '[':
			switch key[2] {
			case 'A': // Up arrow - move up 1 instruction
				if addr >= codeStart+4 {
					newAddr = addr - 4
				}
			case 'B': // Down arrow - move down 1 instruction
				if addr+4 <= maxAddr {
					newAddr = addr + 4
				}
			case 'H': // Home key - go to start of code
				newAddr = codeStart
			case 'F': // End key - go to end of code
				newAddr = maxAddr &^ 0x3
			}

		case len(key) == 4 && key[0] == 27 && key[1] == '[':
			if key[3] == '~' {
				switch key[2] {
				case '1': // Home key (alternate)
					newAddr = codeStart
				case '4': // End key (alternate)
					newAddr = maxAddr &^ 0x3
				case '5': // Page Up - move up by half screen
					_, height := getTerminalSize()
					instrs := height - 6
					if instrs < 5 {
						instrs = 5
					}
					offset := uint32(instrs * 4)
					if addr >= codeStart+offset {
						newAddr = addr - offset
					} else {
						newAddr = codeStart
					}
				case '6': // Page Down - move down by half screen
					_, height := getTerminalSize()
					instrs := height - 6
					if instrs < 5 {
						instrs = 5
					}
					offset := uint32(instrs * 4)
					if addr+offset <= maxAddr {
						newAddr = addr + offset
					} else {
						newAddr = maxAddr &^ 0x3
					}
				}
			}

		case len(key) == 6 && key[0] == 27 && key[1] == '[' && key[2] == '1' && key[3] == ';' && key[4] == '2':
			// Shift+Arrow keys
			switch key[5] {
			case 'A': // Shift+Up - move up 10 instructions
				if addr >= codeStart+40 {
					newAddr = addr - 40
				} else {
					newAddr = codeStart
				}
			case 'B': // Shift+Down - move down 10 instructions
				if addr+40 <= maxAddr {
					newAddr = addr + 40
				} else {
					newAddr = maxAddr &^ 0x3
				}
			}

		case len(key) == 1 && key[0] == 'P': // P - go to PC
			pc := c.Backend().GetState().PC &^ 0x3
			if pc >= codeStart && pc <= maxAddr {
				newAddr = pc
			}

		case len(key) == 1 && key[0] == '?': // ? - show help
			showDisassemblyViewHelp()
			os.Stdin.Read(buf)
			showInteractiveDisassembly(c, addr, codeStart, codeEnd)
			continue
		}

		if newAddr != addr {
			addr = newAddr
			showInteractiveDisassembly(c, addr, codeStart, codeEnd)
		}
	}
}

// showInteractiveDisassembly displays the disassembly view at the given address
func showInteractiveDisassembly(c *debugger.Controller, addr uint32, codeStart, codeEnd uint32) {
	// Clear screen: move to home position, then clear entire screen
	fmt.Print("\033[H\033[2J")

	termWidth, termHeight := getTerminalSize()

	// Calculate how many instructions we can show
	// With the new single-line layout, each instruction takes exactly 1 line
	headerLines := 3
	footerLines := 2
	availableLines := termHeight - headerLines - footerLines
	if availableLines < 4 {
		availableLines = 4
	}

	// Header (use \r\n for raw terminal mode)
	colorHeader.Printf("Disassembly View - %s", colorAddr.Sprintf("0x%08X", addr))
	colorHiBlack.Printf(" (code: 0x%X - 0x%X)\r\n", codeStart, codeEnd)
	colorHiBlack.Print("↑/↓: ±1 instr | Shift+↑/↓: ±10 instrs | PgUp/PgDn | Home/End | P: PC | q/ESC: exit\r\n")
	fmt.Print("\r\n")

	// Get state for PC and breakpoints
	state := c.Backend().GetState()
	breakpoints := c.Backend().ListBreakpoints()
	bpSet := make(map[uint32]bool)
	for _, bp := range breakpoints {
		bpSet[bp.Address] = true
	}

	// Collect all instructions with their metadata
	var instructions []cfgInstrData
	currentAddr := addr

	// Collect instructions for the visible range (1 instruction per line)
	for currentAddr < codeEnd && len(instructions) < availableLines {
		instrs, err := c.Backend().Disassemble(currentAddr, 1)
		if err != nil || len(instrs) == 0 {
			currentAddr += 4
			continue
		}
		instr := instrs[0]

		data := cfgInstrData{
			addr:          currentAddr,
			mnemonic:      instr.Mnemonic,
			operands:      instr.Operands,
			isPC:          currentAddr == state.PC,
			isBP:          bpSet[currentAddr],
			branchTarget:  instr.BranchTarget,
			branchTargetS: instr.BranchTargetSym,
		}

		// Check if this is a branch instruction
		data.isBranch = isBranchMnemonic(instr.Mnemonic)

		// Check if there's a label/symbol at this address
		if label, ok := c.Backend().GetSymbolAt(currentAddr); ok {
			data.label = label
		}

		// Get source info
		if srcLoc := c.Backend().GetSourceLocation(currentAddr); srcLoc != nil {
			data.srcFile = srcLoc.File
			data.srcLine = srcLoc.Line
			if debugInfo := c.Backend().DebugInfo(); debugInfo != nil {
				data.srcText = strings.TrimSpace(debugInfo.GetSourceLine(srcLoc.File, srcLoc.Line))
			}
		}

		instructions = append(instructions, data)
		currentAddr += 4
	}

	// Build CFG edges for visible range
	var edges []cfgEdge
	addrToIdx := make(map[uint32]int)
	for i, instr := range instructions {
		addrToIdx[instr.addr] = i
	}

	// Calculate visible address range
	visibleStartAddr := addr
	visibleEndAddr := addr
	if len(instructions) > 0 {
		visibleEndAddr = instructions[len(instructions)-1].addr
	}

	// Scan a wide range around the visible area to find ALL branches whose edge
	// passes through the visible range (even if both endpoints are outside)
	scanRange := uint32(200 * 4) // Scan 200 instructions in each direction
	scanStart := codeStart
	if addr > scanRange && addr-scanRange > codeStart {
		scanStart = addr - scanRange
	}
	scanEnd := codeEnd
	if currentAddr+scanRange < codeEnd {
		scanEnd = currentAddr + scanRange
	}

	// Find all branches in the scan range
	maxColumn := 0
	seenEdges := make(map[uint64]bool) // Track edges we've already added (source<<32 | target)

	for scanAddr := scanStart; scanAddr < scanEnd; scanAddr += 4 {
		var branchAddr uint32
		var branchTarget uint32
		var isBranch bool

		// Check if this address is in our visible instructions (already collected)
		if idx, ok := addrToIdx[scanAddr]; ok {
			instr := instructions[idx]
			if instr.isBranch && instr.branchTarget != 0 {
				branchAddr = instr.addr
				branchTarget = instr.branchTarget
				isBranch = true
			}
		} else {
			// Need to disassemble
			instrs, err := c.Backend().Disassemble(scanAddr, 1)
			if err != nil || len(instrs) == 0 {
				continue
			}
			instr := instrs[0]
			if isBranchMnemonic(instr.Mnemonic) && instr.BranchTarget != 0 {
				branchAddr = scanAddr
				branchTarget = instr.BranchTarget
				isBranch = true
			}
		}

		if !isBranch {
			continue
		}

		// Check if this edge overlaps with the visible range
		// An edge from branchAddr to branchTarget overlaps if:
		// min(branchAddr, branchTarget) <= visibleEndAddr AND max(branchAddr, branchTarget) >= visibleStartAddr
		edgeMin := branchAddr
		edgeMax := branchTarget
		if branchTarget < branchAddr {
			edgeMin, edgeMax = branchTarget, branchAddr
		}

		if edgeMin <= visibleEndAddr && edgeMax >= visibleStartAddr {
			// Edge overlaps with visible range - include it
			edgeKey := uint64(branchAddr)<<32 | uint64(branchTarget)
			if seenEdges[edgeKey] {
				continue
			}
			seenEdges[edgeKey] = true

			// Determine start and end indices
			startIdx := -1 // Default: source before visible range
			if idx, ok := addrToIdx[branchAddr]; ok {
				startIdx = idx
			} else if branchAddr > visibleEndAddr {
				startIdx = len(instructions) // Source after visible range
			}

			endIdx := -1 // Default: target before visible range
			if idx, ok := addrToIdx[branchTarget]; ok {
				endIdx = idx
			} else if branchTarget > visibleEndAddr {
				endIdx = len(instructions) // Target after visible range
			}

			isBackward := branchTarget < branchAddr
			edge := cfgEdge{
				startIdx:   startIdx,
				endIdx:     endIdx,
				isBackward: isBackward,
				destAddr:   branchTarget,
				color:      cfgEdgeColors[len(edges)%len(cfgEdgeColors)],
			}

			edge.column = assignEdgeColumn(edges, edge, len(instructions))
			if edge.column > maxColumn {
				maxColumn = edge.column
			}

			edges = append(edges, edge)
		}
	}

	// Calculate CFG column width
	cfgWidth := 0
	if len(edges) > 0 {
		cfgWidth = (maxColumn + 1) * 2 // 2 chars per column level
	}

	// Find the maximum label width for alignment
	maxLabelWidth := 0
	for _, instr := range instructions {
		if len(instr.label) > maxLabelWidth {
			maxLabelWidth = len(instr.label)
		}
	}
	// Cap label width to avoid eating up too much space
	if maxLabelWidth > 16 {
		maxLabelWidth = 16
	}

	// Calculate max instruction width for alignment
	maxInstrWidth := 0
	for _, instr := range instructions {
		instrLen := len(instr.mnemonic)
		if instr.operands != "" {
			instrLen += 1 + len(instr.operands)
		}
		if instrLen > maxInstrWidth {
			maxInstrWidth = instrLen
		}
	}
	if maxInstrWidth > 24 {
		maxInstrWidth = 24
	}
	if maxInstrWidth < 12 {
		maxInstrWidth = 12
	}

	// Color for labels
	colorLabel := color.New(color.FgMagenta, color.Bold)

	// Pre-compute source line groups: find ranges of consecutive instructions with same source location
	// Each instruction gets: isFirstOfGroup, isLastOfGroup, isMiddleOfGroup
	type srcGroupInfo struct {
		isFirst  bool // First instruction of a source line group
		isLast   bool // Last instruction of a source line group
		isSingle bool // Only instruction in the group (first and last)
	}
	srcGroups := make([]srcGroupInfo, len(instructions))

	for i := range instructions {
		instr := &instructions[i]
		// Check if this is first of a group (different from previous, or first instruction)
		isFirst := i == 0 || instr.srcFile != instructions[i-1].srcFile || instr.srcLine != instructions[i-1].srcLine
		// Check if this is last of a group (different from next, or last instruction)
		isLast := i == len(instructions)-1 || instr.srcFile != instructions[i+1].srcFile || instr.srcLine != instructions[i+1].srcLine
		srcGroups[i] = srcGroupInfo{
			isFirst:  isFirst,
			isLast:   isLast,
			isSingle: isFirst && isLast,
		}
	}

	// Render instructions with CFG - single line per instruction
	linesShown := 0
	for i, instr := range instructions {
		if linesShown >= availableLines {
			break
		}

		// Determine marker (column 0)
		var marker string
		if instr.isPC && instr.isBP {
			marker = colorBreakpoint.Sprint("*") + colorPC.Sprint(">")
		} else if instr.isPC {
			marker = colorPC.Sprint("=>")
		} else if instr.isBP {
			marker = colorBreakpoint.Sprint("* ")
		} else {
			marker = "  "
		}
		fmt.Printf("%s ", marker)

		// Column 1: CFG
		cfgStr := renderCFGLine(edges, i, false, cfgWidth, len(instructions))
		fmt.Print(cfgStr)

		// Column 2: Label/Function name (fixed width, right-aligned)
		labelText := ""
		if instr.label != "" {
			labelText = instr.label
			if len(labelText) > maxLabelWidth {
				labelText = labelText[:maxLabelWidth-1] + "…"
			}
		}
		if labelText != "" {
			colorLabel.Printf("%*s ", maxLabelWidth, labelText)
		} else {
			fmt.Printf("%*s ", maxLabelWidth, "")
		}

		// Column 3: Memory address
		fmt.Printf("%s ", colorAddr.Sprintf("0x%08X", instr.addr))

		// Column 4: Disassembled instruction (fixed width for alignment)
		instrText := instr.mnemonic
		if instr.operands != "" {
			instrText += " " + instr.operands
		}
		// Print instruction with highlighting, then pad to fixed width
		instrLen := printHighlightedInstructionLen(instrText)
		if instrLen < maxInstrWidth {
			fmt.Printf("%*s", maxInstrWidth-instrLen, "")
		}

		// Column 5: Source location with grouping indicator
		// Show source text only on first instruction of each source line group
		// Use visual brackets to show which instructions belong to the same source line
		srcGroup := srcGroups[i]
		if instr.srcFile != "" {
			// Draw grouping bracket on the left of source info
			if srcGroup.isSingle {
				colorHiBlack.Print(" ─ ")
			} else if srcGroup.isFirst {
				colorHiBlack.Print(" ╭ ")
			} else if srcGroup.isLast {
				colorHiBlack.Print(" ╰ ")
			} else {
				colorHiBlack.Print(" │ ")
			}

			// Show file:line on first instruction of group, or for single-instruction groups
			if srcGroup.isFirst {
				colorSourceFile.Printf("%s", filepath.Base(instr.srcFile))
				colorHiBlack.Print(":")
				colorSourceLine.Printf("%-4d", instr.srcLine)

				if instr.srcText != "" {
					colorHiBlack.Print(" ")
					// Calculate remaining width for source text
					// marker(3) + cfg + label + addr(12) + instr + bracket(3) + file:line(~20) + space
					usedWidth := 3 + cfgWidth + maxLabelWidth + 1 + 12 + maxInstrWidth + 3 + 20 + 1
					remainingWidth := termWidth - usedWidth
					if remainingWidth < 10 {
						remainingWidth = 10
					}
					srcText := instr.srcText
					if len(srcText) > remainingWidth {
						srcText = srcText[:remainingWidth-3] + "..."
					}
					fmt.Print(utils.HighlightCCode(srcText))
				}
			}
			// For continuation lines (not first), we just show the bracket - no repeated source info
		}

		// Show branch target hint at the end (after source info to not break alignment)
		if instr.isBranch && instr.branchTarget != 0 {
			if instr.branchTarget < instr.addr {
				colorHiBlack.Printf("  ; ↰ %s", instr.branchTargetS)
			} else {
				colorHiBlack.Printf("  ; ↴ %s", instr.branchTargetS)
			}
		} else if instr.isBranch {
			colorHiBlack.Print("  ; ↪ ?")
		}

		fmt.Print("\r\n")
		linesShown++
	}

	// Footer (no trailing \r\n on last line to prevent scroll)
	fmt.Print("\r\n")
	fmt.Printf("PC: %s  SP: %s",
		colorAddr.Sprintf("0x%08X", state.PC),
		colorAddr.Sprintf("0x%08X", state.SP))
}

// isBranchMnemonic checks if the mnemonic is a branch instruction
func isBranchMnemonic(mnemonic string) bool {
	m := strings.ToUpper(mnemonic)
	return m == "JMP" || m == "CJMP" || strings.HasPrefix(m, "B")
}

// assignEdgeColumn assigns a column to an edge avoiding overlaps
func assignEdgeColumn(edges []cfgEdge, newEdge cfgEdge, totalLines int) int {
	// Find minimum and maximum line covered by this edge
	minLine := newEdge.startIdx
	maxLine := newEdge.endIdx

	// Handle edges from outside the visible range
	if newEdge.startIdx == -1 {
		// Source is before visible range
		minLine = 0
	} else if newEdge.startIdx >= totalLines {
		// Source is after visible range
		minLine = totalLines - 1
	}

	if newEdge.endIdx == -1 {
		// Target is before visible range
		maxLine = 0
	} else if newEdge.endIdx >= totalLines {
		// Target is after visible range
		maxLine = totalLines - 1
	}

	if minLine > maxLine {
		minLine, maxLine = maxLine, minLine
	}

	// Find first column not used by overlapping edges
	column := 0
	for {
		conflict := false
		for _, e := range edges {
			eMin := e.startIdx
			eMax := e.endIdx

			// Handle edges from outside the visible range
			if e.startIdx == -1 {
				eMin = 0
			} else if e.startIdx >= totalLines {
				eMin = totalLines - 1
			}

			if e.endIdx == -1 {
				eMax = 0
			} else if e.endIdx >= totalLines {
				eMax = totalLines - 1
			}

			if eMin > eMax {
				eMin, eMax = eMax, eMin
			}

			// Check overlap
			if e.column == column && !(maxLine < eMin || minLine > eMax) {
				conflict = true
				break
			}
		}
		if !conflict {
			break
		}
		column++
	}

	return column
}

// renderCFGLine renders the CFG visualization for a given instruction line
func renderCFGLine(edges []cfgEdge, lineIdx int, isSubLine bool, totalWidth int, totalLines int) string {
	if totalWidth == 0 {
		return ""
	}

	// Build character buffer for this line
	buf := make([]rune, totalWidth)
	colors := make([]*color.Color, totalWidth)
	for i := range buf {
		buf[i] = ' '
	}

	for _, edge := range edges {
		col := edge.column * 2
		if col >= totalWidth {
			continue
		}

		// Determine effective min/max lines for this edge
		minLine := edge.startIdx
		maxLine := edge.endIdx

		// Handle edges from outside the visible range
		sourceOutsideTop := edge.startIdx == -1
		sourceOutsideBottom := edge.startIdx >= totalLines
		targetOutsideTop := edge.endIdx == -1
		targetOutsideBottom := edge.endIdx >= totalLines

		if sourceOutsideTop {
			minLine = 0
		} else if sourceOutsideBottom {
			minLine = totalLines - 1
		}

		if targetOutsideTop {
			maxLine = 0
		} else if targetOutsideBottom {
			maxLine = totalLines - 1
		}

		if minLine > maxLine {
			minLine, maxLine = maxLine, minLine
		}

		edgeColor := edge.color
		if edgeColor == nil {
			edgeColor = colorHiBlack
		}

		// Check if this line is part of this edge
		if lineIdx >= minLine && lineIdx <= maxLine {
			isActualStart := lineIdx == edge.startIdx && !isSubLine && !sourceOutsideTop && !sourceOutsideBottom
			isActualEnd := lineIdx == edge.endIdx && !isSubLine && !targetOutsideTop && !targetOutsideBottom

			if isActualStart {
				// Source of branch
				if edge.isBackward {
					buf[col] = '╰'
					if col+1 < totalWidth {
						buf[col+1] = '─'
					}
				} else {
					buf[col] = '╭'
					if col+1 < totalWidth {
						buf[col+1] = '─'
					}
				}
				colors[col] = edgeColor
				if col+1 < totalWidth {
					colors[col+1] = edgeColor
				}
			} else if isActualEnd {
				// Target of branch
				if edge.isBackward || sourceOutsideBottom {
					buf[col] = '╭'
					if col+1 < totalWidth {
						buf[col+1] = '>'
					}
				} else {
					buf[col] = '╰'
					if col+1 < totalWidth {
						buf[col+1] = '>'
					}
				}
				colors[col] = edgeColor
				if col+1 < totalWidth {
					colors[col+1] = edgeColor
				}
			} else {
				// In between (or sub-line), or edge from outside
				buf[col] = '│'
				colors[col] = edgeColor
			}
		}
	}

	// Render with colors
	var result strings.Builder
	for i, r := range buf {
		if colors[i] != nil {
			result.WriteString(colors[i].Sprint(string(r)))
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// showDisassemblyViewHelp displays help for the disassembly view
func showDisassemblyViewHelp() {
	fmt.Print("\033[H\033[J")
	colorHeader.Println("Disassembly View Help")
	fmt.Println()
	fmt.Println("Navigation:")
	fmt.Println("  ↑/↓         Move up/down by 1 instruction")
	fmt.Println("  Shift+↑/↓   Move up/down by 10 instructions")
	fmt.Println("  PgUp/PgDn   Move up/down by screen")
	fmt.Println("  Home        Go to start of code section")
	fmt.Println("  End         Go to end of code section")
	fmt.Println("  P           Go to program counter (PC)")
	fmt.Println()
	fmt.Println("Markers:")
	fmt.Printf("  %s  Current program counter (PC)\n", colorPC.Sprint("=>"))
	fmt.Printf("  %s   Breakpoint\n", colorBreakpoint.Sprint("*"))
	fmt.Printf("  %s  Breakpoint at PC\n", colorBreakpoint.Sprint("*")+colorPC.Sprint(">"))
	fmt.Println()
	fmt.Println("Control Flow Graph:")
	loopColor := color.New(color.FgHiMagenta)
	jumpColor := color.New(color.FgHiCyan)
	fmt.Printf("  %s  Backward branch (loop)\n", loopColor.Sprint("╭─>"))
	fmt.Printf("  %s  Forward branch (jump)\n", jumpColor.Sprint("╰─>"))
	fmt.Printf("  %s  Vertical connector\n", colorHiBlack.Sprint("│"))
	fmt.Println()
	fmt.Println("Display:")
	labelColor := color.New(color.FgMagenta, color.Bold)
	fmt.Printf("  %s  Labels/function names (branch targets)\n", labelColor.Sprint("label:"))
	fmt.Printf("  %s  Branch to target (loop back)\n", colorHiBlack.Sprint("; ↰ target"))
	fmt.Printf("  %s  Branch to target (jump forward)\n", colorHiBlack.Sprint("; ↴ target"))
	fmt.Println()
	fmt.Println("Other:")
	fmt.Println("  q/ESC       Exit disassembly view")
	fmt.Println("  ?           Show this help")
	fmt.Println()
	colorHiBlack.Println("Press any key to continue...")
}

// showMemoryViewHelp displays help for the memory view
func showMemoryViewHelp() {
	fmt.Print("\033[H\033[J")
	colorHeader.Println("Memory View Help")
	fmt.Println()
	fmt.Println("Navigation:")
	fmt.Println("  ↑/↓         Move up/down by 1 word (4 bytes)")
	fmt.Println("  Shift+↑/↓   Move up/down by 10 words (40 bytes)")
	fmt.Println("  PgUp/PgDn   Move up/down by 100 words (400 bytes)")
	fmt.Println("  Home        Go to start of memory (0x00000000)")
	fmt.Println("  End         Go to end of memory")
	fmt.Println("  G           Go to stack pointer (SP)")
	fmt.Println()
	fmt.Println("Region Markers:")
	fmt.Printf("  %s  Code section (shows decoded instructions)\n", colorMemCode.Sprint("C"))
	fmt.Printf("  %s  Data section (global variables)\n", colorMemData.Sprint("D"))
	fmt.Printf("  %s  Stack region (local variables)\n", colorMemStack.Sprint("S"))
	fmt.Printf("  %s  Unknown/unmapped region\n", colorMemUnknown.Sprint("·"))
	fmt.Printf("  %s  Current program counter (PC)\n", colorPC.Sprint("►"))
	fmt.Printf("  %s  Breakpoint\n", colorBreakpoint.Sprint("●"))
	fmt.Println()
	fmt.Println("Other:")
	fmt.Println("  q/ESC       Exit memory view")
	fmt.Println("  ?           Show this help")
	fmt.Println()
	colorHiBlack.Println("Press any key to continue...")
}

// =============================================================================
// Interactive Execution Display
// =============================================================================

// execViewLayout defines the layout dimensions for the execution view
type execViewLayout struct {
	termWidth, termHeight int
	// Panel dimensions
	disasmWidth, disasmHeight int
	regsWidth, regsHeight     int
	varsWidth, varsHeight     int
	stackWidth, stackHeight   int
	statusHeight              int
}

// calculateExecViewLayout computes panel dimensions based on terminal size
func calculateExecViewLayout() execViewLayout {
	w, h := getTerminalSize()

	// Minimum terminal size requirements
	if w < 80 {
		w = 80
	}
	if h < 20 {
		h = 20
	}

	layout := execViewLayout{
		termWidth:    w,
		termHeight:   h,
		statusHeight: 3, // Status bar at bottom
	}

	// Available height for panels (minus header and status)
	availableHeight := h - 4 - layout.statusHeight
	if availableHeight < 10 {
		availableHeight = 10
	}

	// Left side: Disassembly (takes ~60% width, minimum 50 chars)
	layout.disasmWidth = w * 60 / 100
	if layout.disasmWidth < 50 {
		layout.disasmWidth = 50
	}
	// Cap disassembly width to leave room for right panel
	maxDisasmWidth := w - 35 // Leave at least 32 chars + 3 separator for right panel
	if layout.disasmWidth > maxDisasmWidth {
		layout.disasmWidth = maxDisasmWidth
	}
	layout.disasmHeight = availableHeight

	// Right side width (minimum 30 chars for readable register display)
	rightWidth := w - layout.disasmWidth - 3 // 3 for separator " │ "
	if rightWidth < 30 {
		rightWidth = 30
	}

	// Right panels: Registers (top), Variables (middle), Stack (bottom)
	layout.regsWidth = rightWidth

	// Calculate register panel height based on content:
	// - 1 line header
	// - 2 lines for PC/SP/LR/CPSR
	// - 1 line for flags
	// - 7 lines for r0-r12 (pairs)
	// Total: ~11 lines ideal, but cap at 1/3 of available
	layout.regsHeight = 11
	maxRegsHeight := availableHeight / 3
	if layout.regsHeight > maxRegsHeight {
		layout.regsHeight = maxRegsHeight
	}
	if layout.regsHeight < 5 {
		layout.regsHeight = 5 // Minimum to show essential registers
	}

	// Split remaining height between variables and stack
	remainingHeight := availableHeight - layout.regsHeight
	if remainingHeight < 6 {
		remainingHeight = 6
	}

	layout.varsWidth = rightWidth
	layout.varsHeight = remainingHeight / 2
	if layout.varsHeight < 3 {
		layout.varsHeight = 3
	}

	layout.stackWidth = rightWidth
	layout.stackHeight = remainingHeight - layout.varsHeight
	if layout.stackHeight < 3 {
		layout.stackHeight = 3
	}

	return layout
}

// interactiveExecutionView displays a combined execution view with all panels
func interactiveExecutionView(c *debugger.Controller) {
	// Get code bounds
	program := c.Backend().Program()
	layout := program.MemoryLayout()
	if layout == nil {
		c.UI().ShowMessage(debugger.LevelError, "No program loaded")
		return
	}

	// Save terminal state and enable raw mode
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		c.UI().ShowMessage(debugger.LevelError, "Cannot enable raw terminal mode: %v", err)
		return
	}
	defer term.Restore(fd, oldState)

	// Set up resize detection
	resizeChan := make(chan debugger.TerminalSize, 1)
	unregisterResize := c.UI().OnResize(func(size debugger.TerminalSize) {
		select {
		case resizeChan <- size:
		default:
		}
	})
	defer unregisterResize()

	// Track execution state
	execState := &execViewState{
		followPC:    true,
		disasmAddr:  c.Backend().GetState().PC,
		codeStart:   layout.CodeStart,
		codeEnd:     layout.CodeStart + layout.CodeSize,
		memSize:     uint32(len(c.Backend().Runner().State().Memory)),
		statusMsg:   "Ready. Press ? for help.",
		inputMode:   false,
		inputBuffer: "",
		inputPrompt: "",
	}

	// Initial render
	renderExecutionView(c, execState)

	// Input handling goroutine
	buf := make([]byte, 16)
	inputChan := make(chan []byte, 1)
	done := make(chan struct{})
	go func() {
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				close(inputChan)
				return
			}
			key := make([]byte, n)
			copy(key, buf[:n])
			select {
			case inputChan <- key:
			case <-done:
				return
			}
		}
	}()

	cleanup := func() {
		close(done)
		select {
		case <-inputChan:
		default:
		}
	}

	for {
		var key []byte
		select {
		case <-resizeChan:
			renderExecutionView(c, execState)
			continue
		case k, ok := <-inputChan:
			if !ok {
				cleanup()
				return
			}
			key = k
		}

		// Handle input mode (for breakpoint addresses, etc.)
		if execState.inputMode {
			if handleExecViewInput(c, execState, key) {
				renderExecutionView(c, execState)
			}
			continue
		}

		// Handle normal key commands
		action := handleExecViewKey(c, execState, key)
		switch action {
		case execActionQuit:
			fmt.Print("\033[H\033[2J") // Clear screen before exit
			cleanup()
			return
		case execActionRedraw:
			renderExecutionView(c, execState)
		case execActionHelp:
			showExecViewHelp()
			os.Stdin.Read(buf) // Wait for key
			renderExecutionView(c, execState)
		}
	}
}

// execViewState holds the state for the execution view
type execViewState struct {
	followPC    bool   // Auto-follow PC in disassembly
	disasmAddr  uint32 // Current top address in disassembly view
	codeStart   uint32
	codeEnd     uint32
	memSize     uint32
	statusMsg   string // Status message to display
	inputMode   bool   // Whether we're in input mode
	inputBuffer string // Current input buffer
	inputPrompt string // Input prompt text
	inputAction string // What to do with the input ("break", "delete", etc.)
}

// execViewAction represents the action to take after key handling
type execViewAction int

const (
	execActionNone execViewAction = iota
	execActionRedraw
	execActionQuit
	execActionHelp
)

// handleExecViewKey processes a key press in the execution view
func handleExecViewKey(c *debugger.Controller, state *execViewState, key []byte) execViewAction {
	// Get current CPU state
	cpuState := c.Backend().GetState()

	switch {
	// Quit commands
	case len(key) == 1 && (key[0] == 'q' || key[0] == 'Q' || key[0] == 27): // q, Q, ESC
		return execActionQuit

	// Help
	case len(key) == 1 && key[0] == '?':
		return execActionHelp

	// Step commands
	case len(key) == 1 && (key[0] == 's' || key[0] == 'S'): // Step (source-level)
		execStep(c, state, 1, true)
		return execActionRedraw

	case len(key) == 1 && (key[0] == 'i' || key[0] == 'I'): // Step instruction
		execStep(c, state, 1, false)
		return execActionRedraw

	case len(key) == 1 && (key[0] == 'c' || key[0] == 'C'): // Continue
		execContinue(c, state)
		return execActionRedraw

	case len(key) == 1 && (key[0] == 'r' || key[0] == 'R'): // Run
		execRun(c, state)
		return execActionRedraw

	// Breakpoint commands
	case len(key) == 1 && key[0] == 'b': // Add breakpoint at PC
		pc := c.Backend().GetState().PC
		bp, err := c.Backend().AddBreakpoint(pc)
		if err != nil {
			state.statusMsg = fmt.Sprintf("Error: %v", err)
		} else {
			state.statusMsg = fmt.Sprintf("Breakpoint %d set at 0x%08X", bp.ID, pc)
		}
		return execActionRedraw

	case len(key) == 1 && key[0] == 'B': // Add breakpoint at address (prompt)
		state.inputMode = true
		state.inputPrompt = "Break at address: "
		state.inputBuffer = ""
		state.inputAction = "break"
		return execActionRedraw

	case len(key) == 1 && (key[0] == 'd' || key[0] == 'D'): // Delete breakpoint
		state.inputMode = true
		state.inputPrompt = "Delete breakpoint ID: "
		state.inputBuffer = ""
		state.inputAction = "delete"
		return execActionRedraw

	// Toggle follow PC
	case len(key) == 1 && (key[0] == 'f' || key[0] == 'F'):
		state.followPC = !state.followPC
		if state.followPC {
			state.disasmAddr = cpuState.PC
			state.statusMsg = "Follow PC: ON"
		} else {
			state.statusMsg = "Follow PC: OFF"
		}
		return execActionRedraw

	// Navigation (arrow keys)
	case len(key) == 3 && key[0] == 27 && key[1] == '[':
		switch key[2] {
		case 'A': // Up arrow
			if state.disasmAddr >= state.codeStart+4 {
				state.disasmAddr -= 4
				state.followPC = false
			}
			return execActionRedraw
		case 'B': // Down arrow
			if state.disasmAddr+4 < state.codeEnd {
				state.disasmAddr += 4
				state.followPC = false
			}
			return execActionRedraw
		}

	// Page Up/Down
	case len(key) == 4 && key[0] == 27 && key[1] == '[' && key[3] == '~':
		viewLayout := calculateExecViewLayout()
		pageSize := uint32((viewLayout.disasmHeight / 2) * 4) // Half page
		if pageSize < 20 {
			pageSize = 20
		}
		switch key[2] {
		case '5': // Page Up
			if state.disasmAddr >= state.codeStart+pageSize {
				state.disasmAddr -= pageSize
			} else {
				state.disasmAddr = state.codeStart
			}
			state.followPC = false
			return execActionRedraw
		case '6': // Page Down
			if state.disasmAddr+pageSize < state.codeEnd {
				state.disasmAddr += pageSize
			} else {
				state.disasmAddr = state.codeEnd - 4
			}
			state.followPC = false
			return execActionRedraw
		}

	// Go to PC
	case len(key) == 1 && (key[0] == 'p' || key[0] == 'P'):
		state.disasmAddr = cpuState.PC
		state.followPC = true
		state.statusMsg = "Jumped to PC"
		return execActionRedraw

	// Reset
	case len(key) == 1 && key[0] == '0':
		if err := c.Backend().Reset(); err != nil {
			state.statusMsg = fmt.Sprintf("Reset failed: %v", err)
		} else {
			state.disasmAddr = c.Backend().GetState().PC
			state.followPC = true
			state.statusMsg = "Program reset"
		}
		return execActionRedraw

	// Speed control: decrease delay (faster)
	case len(key) == 1 && (key[0] == '+' || key[0] == '='):
		currentDelay := c.Backend().GetExecutionDelay()
		newDelay := decreaseDelay(currentDelay)
		c.Backend().SetExecutionDelay(newDelay)
		state.statusMsg = fmt.Sprintf("Speed: %s", getSpeedName(newDelay))
		return execActionRedraw

	// Speed control: increase delay (slower)
	case len(key) == 1 && key[0] == '-':
		currentDelay := c.Backend().GetExecutionDelay()
		newDelay := increaseDelay(currentDelay)
		c.Backend().SetExecutionDelay(newDelay)
		state.statusMsg = fmt.Sprintf("Speed: %s", getSpeedName(newDelay))
		return execActionRedraw
	}

	return execActionNone
}

// Speed presets (in milliseconds)
var speedPresets = []int{0, 25, 50, 100, 200, 500, 1000}
var speedNames = []string{"instant", "very fast", "fast", "normal", "slow", "very slow", "ultra slow"}

// getSpeedName returns the name for a delay value
func getSpeedName(delayMs int) string {
	for i, preset := range speedPresets {
		if delayMs <= preset {
			return speedNames[i]
		}
	}
	return fmt.Sprintf("%dms", delayMs)
}

// decreaseDelay returns the next faster speed
func decreaseDelay(currentDelay int) int {
	for i := len(speedPresets) - 1; i >= 0; i-- {
		if speedPresets[i] < currentDelay {
			return speedPresets[i]
		}
	}
	return 0
}

// increaseDelay returns the next slower speed
func increaseDelay(currentDelay int) int {
	for _, preset := range speedPresets {
		if preset > currentDelay {
			return preset
		}
	}
	return speedPresets[len(speedPresets)-1]
}

// handleExecViewInput handles input mode (text entry)
func handleExecViewInput(c *debugger.Controller, state *execViewState, key []byte) bool {
	if len(key) == 0 {
		return false
	}

	switch {
	case key[0] == 27: // ESC - cancel input
		state.inputMode = false
		state.inputBuffer = ""
		state.statusMsg = "Cancelled"
		return true

	case key[0] == 13 || key[0] == 10: // Enter - submit
		state.inputMode = false
		processExecViewInput(c, state)
		return true

	case key[0] == 127 || key[0] == 8: // Backspace
		if len(state.inputBuffer) > 0 {
			state.inputBuffer = state.inputBuffer[:len(state.inputBuffer)-1]
		}
		return true

	case key[0] >= 32 && key[0] < 127: // Printable char
		state.inputBuffer += string(key[0])
		return true
	}

	return false
}

// processExecViewInput processes the completed input
func processExecViewInput(c *debugger.Controller, state *execViewState) {
	input := strings.TrimSpace(state.inputBuffer)
	if input == "" {
		state.statusMsg = "No input"
		return
	}

	switch state.inputAction {
	case "break":
		addr, err := c.ResolveAddress(input)
		if err != nil {
			state.statusMsg = fmt.Sprintf("Invalid address: %s", input)
			return
		}
		bp, err := c.Backend().AddBreakpoint(addr)
		if err != nil {
			state.statusMsg = fmt.Sprintf("Error: %v", err)
			return
		}
		state.statusMsg = fmt.Sprintf("Breakpoint %d set at 0x%08X", bp.ID, addr)

	case "delete":
		id, err := strconv.Atoi(input)
		if err != nil {
			state.statusMsg = fmt.Sprintf("Invalid ID: %s", input)
			return
		}
		err = c.Backend().RemoveBreakpoint(id)
		if err != nil {
			state.statusMsg = fmt.Sprintf("Error: %v", err)
			return
		}
		state.statusMsg = fmt.Sprintf("Breakpoint %d deleted", id)
	}
}

// execStep executes step(s) and updates state
func execStep(c *debugger.Controller, state *execViewState, count int, sourceLevel bool) {
	// Execute step without UI output (we'll render ourselves)
	var result debugger.ExecutionResult
	if sourceLevel {
		// Source-level step
		result = stepSourceLineQuiet(c, count)
	} else {
		// Instruction step
		result = c.Backend().Step(count)
	}

	// Update state based on result
	updateStateAfterExec(c, state, result, "step")
}

// stepSourceLineQuiet does source-level stepping without UI output
func stepSourceLineQuiet(c *debugger.Controller, count int) debugger.ExecutionResult {
	var lastResult debugger.ExecutionResult

	debugInfo := c.Backend().DebugInfo()
	hasDebugInfo := debugInfo != nil && len(debugInfo.InstructionLocations) > 0

	if !hasDebugInfo {
		// No debug info: fall back to instruction stepping
		return c.Backend().Step(count)
	}

	for i := 0; i < count; i++ {
		cpuState := c.Backend().GetState()
		startLoc := c.Backend().GetSourceLocation(cpuState.PC)

		var startFile string
		var startLine int
		if startLoc != nil && startLoc.IsValid() {
			startFile = startLoc.File
			startLine = startLoc.Line
		}

		// Step until source line changes
		const maxInstructions = 10000
		for step := 0; step < maxInstructions; step++ {
			lastResult = c.Backend().Step(1)
			if lastResult.Error != nil || lastResult.StopReason != 0 {
				return lastResult
			}

			cpuState = c.Backend().GetState()
			currentLoc := c.Backend().GetSourceLocation(cpuState.PC)

			if startLoc == nil || !startLoc.IsValid() {
				if currentLoc != nil && currentLoc.IsValid() {
					break
				}
				continue
			}

			if currentLoc == nil || !currentLoc.IsValid() {
				continue
			}

			if currentLoc.File != startFile || currentLoc.Line != startLine {
				break
			}
		}

		if lastResult.Error != nil || lastResult.StopReason != 0 {
			return lastResult
		}
	}

	return lastResult
}

// execContinue runs until breakpoint/termination
func execContinue(c *debugger.Controller, state *execViewState) {
	state.statusMsg = "Running..."
	result := c.Backend().Continue()
	updateStateAfterExec(c, state, result, "continue")
}

// execRun runs until termination
func execRun(c *debugger.Controller, state *execViewState) {
	state.statusMsg = "Running..."
	result := c.Backend().Run()
	updateStateAfterExec(c, state, result, "run")
}

// updateStateAfterExec updates the view state after execution
func updateStateAfterExec(c *debugger.Controller, state *execViewState, result debugger.ExecutionResult, cmd string) {
	cpuState := c.Backend().GetState()

	// Update disassembly address if following PC
	if state.followPC {
		state.disasmAddr = cpuState.PC
	}

	// Update status based on result
	switch result.StopReason {
	case interpreter.StopNone:
		state.statusMsg = "Ready"

	case interpreter.StopStep:
		state.statusMsg = fmt.Sprintf("Stepped %d instruction(s)", result.StepsExecuted)

	case interpreter.StopBreakpoint:
		state.statusMsg = fmt.Sprintf("Breakpoint %d hit at 0x%08X",
			result.BreakpointID, result.LastPC)

	case interpreter.StopWatchpoint:
		state.statusMsg = fmt.Sprintf("Watchpoint %d triggered at 0x%08X",
			result.WatchpointID, result.LastPC)

	case interpreter.StopHalt:
		state.statusMsg = "CPU halted"

	case interpreter.StopError:
		state.statusMsg = fmt.Sprintf("Error at 0x%08X", result.LastPC)

	case interpreter.StopTermination:
		state.statusMsg = fmt.Sprintf("Program terminated. Return value: %d (0x%08X)",
			int32(result.ReturnValue), result.ReturnValue)

	case interpreter.StopMaxSteps:
		state.statusMsg = fmt.Sprintf("Max steps reached (%d)", result.StepsExecuted)

	case interpreter.StopInterrupt:
		state.statusMsg = fmt.Sprintf("Interrupted after %d steps", result.StepsExecuted)

	default:
		state.statusMsg = fmt.Sprintf("Unknown stop reason: %d", result.StopReason)
	}

	if result.Error != nil {
		state.statusMsg = fmt.Sprintf("Error: %v", result.Error)
	}
}

// renderExecutionView renders the entire execution view
func renderExecutionView(c *debugger.Controller, state *execViewState) {
	// Clear screen
	fmt.Print("\033[H\033[2J")

	layout := calculateExecViewLayout()
	cpuState := c.Backend().GetState()

	// Header
	colorHeader.Printf("═══ Cucaracha Execution View ")
	fmt.Print(strings.Repeat("═", layout.termWidth-30))
	fmt.Print("\r\n")

	// Calculate available height for main content
	mainHeight := layout.termHeight - 4 - layout.statusHeight

	// Collect disassembly panel content
	disasmLines := renderDisasmPanel(c, state, layout.disasmWidth, mainHeight)

	// Collect right panel content (registers, variables, stack)
	rightLines := renderRightPanels(c, cpuState, layout, mainHeight)

	// Render both panels side by side
	for i := 0; i < mainHeight; i++ {
		// Disassembly line
		if i < len(disasmLines) {
			fmt.Print(disasmLines[i])
		} else {
			fmt.Print(strings.Repeat(" ", layout.disasmWidth))
		}

		// Separator
		colorHiBlack.Print(" │ ")

		// Right panel line
		if i < len(rightLines) {
			fmt.Print(rightLines[i])
		}

		fmt.Print("\r\n")
	}

	// Separator
	fmt.Print(strings.Repeat("─", layout.termWidth))
	fmt.Print("\r\n")

	// Status bar
	renderStatusBar(c, state, layout)
}

// renderDisasmPanel renders the disassembly panel with source code grouping
func renderDisasmPanel(c *debugger.Controller, state *execViewState, width, height int) []string {
	lines := make([]string, 0, height)

	// Panel header - calculate visible length properly
	headerPrefix := "┌─ Disassembly @ "
	headerAddr := fmt.Sprintf("0x%08X ", state.disasmAddr)
	headerSuffix := ""
	if state.followPC {
		headerSuffix = "[follow PC]"
	}
	// Calculate remaining dashes needed (prefix + addr + suffix + dashes = width)
	visibleLen := len(headerPrefix) + len(headerAddr) + len(headerSuffix)
	dashCount := width - visibleLen
	if dashCount < 0 {
		dashCount = 0
	}
	header := headerPrefix + colorAddr.Sprint(headerAddr)
	if state.followPC {
		header += colorSuccess.Sprint(headerSuffix)
	}
	header += strings.Repeat("─", dashCount)
	lines = append(lines, truncateOrPadString(header, width))

	// Get breakpoints
	bpSet := make(map[uint32]bool)
	for _, bp := range c.Backend().ListBreakpoints() {
		bpSet[bp.Address] = true
	}

	cpuState := c.Backend().GetState()

	// First pass: collect all instructions and their source info
	type instrInfo struct {
		addr     uint32
		mnemonic string
		operands string
		srcFile  string
		srcLine  int
		srcText  string
		isPC     bool
		isBP     bool
	}
	instructions := make([]instrInfo, 0, height-1)
	addr := state.disasmAddr

	for len(instructions) < height-1 && addr < state.codeEnd {
		instrs, err := c.Backend().Disassemble(addr, 1)
		if err != nil || len(instrs) == 0 {
			instructions = append(instructions, instrInfo{
				addr:     addr,
				mnemonic: "???",
				isPC:     addr == cpuState.PC,
				isBP:     bpSet[addr],
			})
			addr += 4
			continue
		}

		instr := instrs[0]
		info := instrInfo{
			addr:     addr,
			mnemonic: instr.Mnemonic,
			operands: instr.Operands,
			isPC:     addr == cpuState.PC,
			isBP:     bpSet[addr],
		}

		// Get source location
		if srcLoc := c.Backend().GetSourceLocation(addr); srcLoc != nil {
			info.srcFile = srcLoc.File
			info.srcLine = srcLoc.Line
			// Get source text if debug info is available
			if debugInfo := c.Backend().DebugInfo(); debugInfo != nil {
				info.srcText = strings.TrimSpace(debugInfo.GetSourceLine(srcLoc.File, srcLoc.Line))
			}
		}

		instructions = append(instructions, info)
		addr += 4
	}

	// Second pass: compute source line groups
	type srcGroupInfo struct {
		isFirst  bool
		isLast   bool
		isSingle bool
	}
	srcGroups := make([]srcGroupInfo, len(instructions))

	for i := range instructions {
		instr := &instructions[i]
		// Check if this is first of a group (different from previous, or first instruction)
		isFirst := i == 0 || instr.srcFile != instructions[i-1].srcFile || instr.srcLine != instructions[i-1].srcLine
		// Check if this is last of a group (different from next, or last instruction)
		isLast := i == len(instructions)-1 || instr.srcFile != instructions[i+1].srcFile || instr.srcLine != instructions[i+1].srcLine
		srcGroups[i] = srcGroupInfo{
			isFirst:  isFirst,
			isLast:   isLast,
			isSingle: isFirst && isLast,
		}
	}

	// Colors for source display
	colorSourceFile := color.New(color.FgCyan)
	colorSourceLine := color.New(color.FgYellow)
	colorSourceText := color.New(color.FgWhite)

	// Third pass: render instructions with source grouping
	for i, instr := range instructions {
		line := "│ "

		// Marker
		if instr.isPC && instr.isBP {
			line += colorBreakpoint.Sprint("*") + colorPC.Sprint(">")
		} else if instr.isPC {
			line += colorPC.Sprint("=>")
		} else if instr.isBP {
			line += colorBreakpoint.Sprint("* ")
		} else {
			line += "  "
		}

		// Address
		line += colorAddr.Sprintf(" 0x%08X", instr.addr) + " "

		// Instruction (fixed width for alignment)
		instrText := instr.mnemonic
		if instr.operands != "" {
			instrText += " " + instr.operands
		}
		instrColored := colorizeInstruction(instrText)
		// Pad instruction to 18 chars for alignment
		instrPadding := 18 - len(instrText)
		if instrPadding < 0 {
			instrPadding = 0
		}
		line += instrColored + strings.Repeat(" ", instrPadding)

		// Source grouping indicator and source info
		srcGroup := srcGroups[i]
		if instr.srcFile != "" {
			// Draw grouping bracket
			if srcGroup.isSingle {
				line += colorHiBlack.Sprint(" ─ ")
			} else if srcGroup.isFirst {
				line += colorHiBlack.Sprint(" ╭ ")
			} else if srcGroup.isLast {
				line += colorHiBlack.Sprint(" ╰ ")
			} else {
				line += colorHiBlack.Sprint(" │ ")
			}

			// Show file:line on first instruction of group
			if srcGroup.isFirst {
				fileBase := filepath.Base(instr.srcFile)
				line += colorSourceFile.Sprint(fileBase)
				line += colorHiBlack.Sprint(":")
				line += colorSourceLine.Sprintf("%-4d", instr.srcLine)

				// Show source text if there's room
				if instr.srcText != "" {
					srcTextTrimmed := strings.TrimSpace(instr.srcText)
					// Calculate remaining width (rough estimate: border + marker + addr + instr + bracket + file:line)
					usedLen := 2 + 3 + 12 + 18 + 3 + len(fileBase) + 1 + 4 + 1
					remainingWidth := width - usedLen
					if remainingWidth > 5 {
						if len(srcTextTrimmed) > remainingWidth {
							srcTextTrimmed = srcTextTrimmed[:remainingWidth-1] + "…"
						}
						line += " " + colorSourceText.Sprint(srcTextTrimmed)
					}
				}
			}
		}

		lines = append(lines, truncateOrPadString(line, width))
	}

	// Pad remaining lines
	for len(lines) < height {
		lines = append(lines, "│"+strings.Repeat(" ", width-1))
	}

	return lines
}

// renderRightPanels renders the right side panels (registers, variables, stack)
func renderRightPanels(c *debugger.Controller, cpuState debugger.DebuggerState, layout execViewLayout, totalHeight int) []string {
	lines := make([]string, 0, totalHeight)

	// Calculate panel heights
	regsHeight := layout.regsHeight
	varsHeight := layout.varsHeight
	stackHeight := layout.stackHeight
	rightWidth := layout.regsWidth

	// Adjust heights to fit
	usedHeight := regsHeight + varsHeight + stackHeight
	if usedHeight > totalHeight {
		// Reduce proportionally
		ratio := float64(totalHeight) / float64(usedHeight)
		regsHeight = int(float64(regsHeight) * ratio)
		varsHeight = int(float64(varsHeight) * ratio)
		stackHeight = totalHeight - regsHeight - varsHeight
	}

	// Registers panel
	lines = append(lines, renderRegistersPanel(cpuState, rightWidth, regsHeight)...)

	// Variables panel
	vars := c.Backend().GetVariables(cpuState.PC)
	lines = append(lines, renderVariablesPanel(vars, rightWidth, varsHeight)...)

	// Call Stack panel - show unwound call stack frames
	callStack := c.Backend().GetCallStack()
	lines = append(lines, renderCallStackPanel(callStack, rightWidth, stackHeight)...)

	// Pad if needed
	for len(lines) < totalHeight {
		lines = append(lines, strings.Repeat(" ", rightWidth))
	}

	return lines
}

// renderRegistersPanel renders the registers panel
func renderRegistersPanel(state debugger.DebuggerState, width, height int) []string {
	lines := make([]string, 0, height)

	// Header
	headerText := "┌─ Registers "
	header := headerText + strings.Repeat("─", width-len(headerText))
	lines = append(lines, truncateOrPadString(header, width))

	if height <= 1 {
		return lines
	}

	// PC, SP, LR, CPSR
	lines = append(lines, truncateOrPadString(fmt.Sprintf("│ %s: %s  %s: %s",
		colorReg.Sprint("PC"), colorAddr.Sprintf("0x%08X", state.PC),
		colorReg.Sprint("SP"), colorAddr.Sprintf("0x%08X", state.SP)), width))

	if len(lines) >= height {
		return lines
	}

	lines = append(lines, truncateOrPadString(fmt.Sprintf("│ %s: %s  %s: %s",
		colorReg.Sprint("LR"), colorAddr.Sprintf("0x%08X", state.LR),
		colorReg.Sprint("CPSR"), colorHex.Sprintf("0x%08X", state.CPSR)), width))

	if len(lines) >= height {
		return lines
	}

	// Flags
	formatFlag := func(name string, set bool) string {
		if set {
			return colorFlagSet.Sprint(name)
		}
		return colorFlagClear.Sprint(name)
	}
	flagsLine := fmt.Sprintf("│ Flags: %s %s %s %s",
		formatFlag("N", state.Flags.N),
		formatFlag("Z", state.Flags.Z),
		formatFlag("C", state.Flags.C),
		formatFlag("V", state.Flags.V))
	lines = append(lines, truncateOrPadString(flagsLine, width))

	if len(lines) >= height {
		return lines
	}

	// General purpose registers (r0-r12) in pairs
	for i := 0; i < 13 && len(lines) < height; i += 2 {
		var r0, r1 uint32
		for _, reg := range state.Registers {
			if reg.Name == fmt.Sprintf("r%d", i) {
				r0 = reg.Value
			}
			if reg.Name == fmt.Sprintf("r%d", i+1) {
				r1 = reg.Value
			}
		}

		if i+1 < 13 {
			line := fmt.Sprintf("│ %s: %s  %s: %s",
				colorReg.Sprintf("r%-2d", i), colorValue.Sprintf("%10d", int32(r0)),
				colorReg.Sprintf("r%-2d", i+1), colorValue.Sprintf("%10d", int32(r1)))
			lines = append(lines, truncateOrPadString(line, width))
		} else {
			line := fmt.Sprintf("│ %s: %s",
				colorReg.Sprintf("r%-2d", i), colorValue.Sprintf("%10d", int32(r0)))
			lines = append(lines, truncateOrPadString(line, width))
		}
	}

	// Pad
	for len(lines) < height {
		lines = append(lines, "│"+strings.Repeat(" ", width-1))
	}

	return lines
}

// renderVariablesPanel renders the variables panel
func renderVariablesPanel(vars []debugger.VariableValue, width, height int) []string {
	lines := make([]string, 0, height)

	// Header
	headerText := "├─ Variables "
	header := headerText + strings.Repeat("─", width-len(headerText))
	lines = append(lines, truncateOrPadString(header, width))

	if height <= 1 {
		return lines
	}

	if len(vars) == 0 {
		lines = append(lines, truncateOrPadString("│ (no variables in scope)", width))
	} else {
		for i, v := range vars {
			if len(lines) >= height {
				break
			}
			line := fmt.Sprintf("│ %s %s = %s",
				colorVarName.Sprint(v.Name),
				colorVarType.Sprint(v.TypeName),
				colorValue.Sprint(v.ValueString))
			if v.Location != "" && v.Location != "<optimized out>" {
				line += colorHiBlack.Sprintf(" @ %s", v.Location)
			}
			lines = append(lines, truncateOrPadString(line, width))
			_ = i
		}
	}

	// Pad
	for len(lines) < height {
		lines = append(lines, "│"+strings.Repeat(" ", width-1))
	}

	return lines
}

// renderCallStackPanel renders the call stack panel showing unwound stack frames
func renderCallStackPanel(frames []debugger.StackFrame, width, height int) []string {
	lines := make([]string, 0, height)

	// Header
	headerText := fmt.Sprintf("├─ Call Stack (%d frames) ", len(frames))
	header := headerText + strings.Repeat("─", width-len(headerText))
	lines = append(lines, truncateOrPadString(header, width))

	if height <= 1 {
		return lines
	}

	if len(frames) == 0 {
		line := "│ " + colorHiBlack.Sprint("(no call stack)")
		lines = append(lines, truncateOrPadString(line, width))
	} else {
		// Show frames with most recent (current) first
		for i, frame := range frames {
			if len(lines) >= height {
				break
			}

			// Frame number and marker
			marker := "  "
			if i == 0 {
				marker = colorPC.Sprint("→ ")
			}

			// Build frame line
			funcName := frame.Function
			if funcName == "" {
				funcName = "??"
			}

			// Format: #N funcName at file:line [addr]
			var line string
			if frame.File != "" && frame.Line > 0 {
				fileName := filepath.Base(frame.File)
				line = fmt.Sprintf("│ %s#%d %s at %s:%d",
					marker, i,
					colorFunc.Sprint(funcName),
					colorSourceFile.Sprint(fileName),
					frame.Line)
			} else if frame.File != "" {
				fileName := filepath.Base(frame.File)
				line = fmt.Sprintf("│ %s#%d %s at %s",
					marker, i,
					colorFunc.Sprint(funcName),
					colorSourceFile.Sprint(fileName))
			} else {
				line = fmt.Sprintf("│ %s#%d %s",
					marker, i,
					colorFunc.Sprint(funcName))
			}

			// Add address on same line if room
			addrStr := colorAddr.Sprintf(" [0x%08X]", frame.Address)
			if len(line)+14 <= width { // 14 = len(" [0x00000000]")
				line += addrStr
			}

			lines = append(lines, truncateOrPadString(line, width))
		}
	}

	// Pad
	for len(lines) < height {
		lines = append(lines, "│"+strings.Repeat(" ", width-1))
	}

	return lines
}

// renderStatusBar renders the status bar at the bottom
func renderStatusBar(c *debugger.Controller, state *execViewState, layout execViewLayout) {
	// Input mode line
	if state.inputMode {
		colorPrompt.Printf("%s", state.inputPrompt)
		fmt.Print(state.inputBuffer)
		colorHiBlack.Print("_")
		fmt.Print(strings.Repeat(" ", layout.termWidth-len(state.inputPrompt)-len(state.inputBuffer)-2))
		fmt.Print("\r\n")
	} else {
		// Status message
		fmt.Print(state.statusMsg)
		fmt.Print(strings.Repeat(" ", layout.termWidth-len(state.statusMsg)))
		fmt.Print("\r\n")
	}

	// Breakpoints line
	bps := c.Backend().ListBreakpoints()
	bpInfo := fmt.Sprintf("Breakpoints: %d", len(bps))
	if len(bps) > 0 && len(bps) <= 5 {
		addrs := make([]string, 0, len(bps))
		for _, bp := range bps {
			addrs = append(addrs, fmt.Sprintf("%d@0x%X", bp.ID, bp.Address))
		}
		bpInfo += " [" + strings.Join(addrs, ", ") + "]"
	}

	// Add speed indicator
	speedDelay := c.Backend().GetExecutionDelay()
	speedInfo := fmt.Sprintf(" | Speed: %s", getSpeedName(speedDelay))
	bpInfo += speedInfo

	colorHiBlack.Print(bpInfo)
	fmt.Print(strings.Repeat(" ", layout.termWidth-len(bpInfo)))
	fmt.Print("\r\n")

	// Help hint (no trailing newline)
	helpHint := "s:step i:stepi c:run r:run b:bp d:del f:follow p:PC 0:rst +/-:speed ?:help q:quit"
	colorHiBlack.Print(helpHint)
}

// truncateOrPadString truncates or pads a string to exactly the specified width, accounting for ANSI codes
func truncateOrPadString(s string, width int) string {
	// Count visible length and track positions
	visibleLen := 0
	inEscape := false
	result := strings.Builder{}

	for _, r := range s {
		if r == '\033' {
			inEscape = true
			result.WriteRune(r)
			continue
		}
		if inEscape {
			result.WriteRune(r)
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		if visibleLen >= width {
			break
		}
		result.WriteRune(r)
		visibleLen++
	}

	// Pad with spaces if needed
	for visibleLen < width {
		result.WriteRune(' ')
		visibleLen++
	}

	return result.String()
}

// truncateString truncates a string to fit within width, accounting for ANSI codes
func truncateString(s string, width int) string {
	// Simple truncation - doesn't account for ANSI codes perfectly but good enough
	visibleLen := 0
	inEscape := false
	result := strings.Builder{}

	for _, r := range s {
		if r == '\033' {
			inEscape = true
			result.WriteRune(r)
			continue
		}
		if inEscape {
			result.WriteRune(r)
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		if visibleLen >= width {
			break
		}
		result.WriteRune(r)
		visibleLen++
	}

	return result.String()
}

// showExecViewHelp displays help for the execution view
func showExecViewHelp() {
	fmt.Print("\033[H\033[2J")
	colorHeader.Println("Execution View Help")
	fmt.Println()
	colorHeader.Println("Execution Commands:")
	fmt.Println("  s           Step (source-level, steps to next source line)")
	fmt.Println("  i           Step instruction (single machine instruction)")
	fmt.Println("  c           Continue until breakpoint or termination")
	fmt.Println("  r           Run until termination")
	fmt.Println("  0           Reset program to initial state")
	fmt.Println()
	colorHeader.Println("Speed Control:")
	fmt.Println("  + or =      Increase execution speed (shorter delay)")
	fmt.Println("  -           Decrease execution speed (longer delay)")
	fmt.Println("              Presets: instant, very fast, fast, normal, slow, very slow, ultra slow")
	fmt.Println()
	colorHeader.Println("Breakpoints:")
	fmt.Println("  b           Add breakpoint (prompts for address)")
	fmt.Println("  d           Delete breakpoint (prompts for ID)")
	fmt.Println()
	colorHeader.Println("Navigation:")
	fmt.Println("  ↑/↓         Scroll disassembly up/down")
	fmt.Println("  PgUp/PgDn   Scroll disassembly by half page")
	fmt.Println("  p/P         Jump to current PC")
	fmt.Println("  f/F         Toggle follow PC mode")
	fmt.Println()
	colorHeader.Println("Display:")
	fmt.Printf("  %s  Current program counter (PC)\n", colorPC.Sprint("=>"))
	fmt.Printf("  %s   Breakpoint set\n", colorBreakpoint.Sprint("*"))
	fmt.Printf("  %s  Breakpoint at PC\n", colorBreakpoint.Sprint("*")+colorPC.Sprint(">"))
	fmt.Println()
	colorHeader.Println("Other:")
	fmt.Println("  ?           Show this help")
	fmt.Println("  q/Q/ESC     Exit execution view")
	fmt.Println()
	colorHiBlack.Println("Press any key to continue...")
}

// Static helper functions for interactive mode (don't need cliUI)
func getRegionForAddressStatic(addr uint32, regions []debugger.MemoryRegion) (debugger.MemoryRegionType, string) {
	for _, r := range regions {
		if addr >= r.StartAddr && addr < r.EndAddr {
			return r.RegionType, r.Name
		}
	}
	return debugger.RegionUnknown, ""
}

func getRegionMarkerStatic(regionType debugger.MemoryRegionType) string {
	switch regionType {
	case debugger.RegionCode:
		return colorMemCode.Sprint("C")
	case debugger.RegionData:
		return colorMemData.Sprint("D")
	case debugger.RegionStack:
		return colorMemStack.Sprint("S")
	default:
		return colorMemUnknown.Sprint("·")
	}
}

func printColoredByteStatic(b byte, regionType debugger.MemoryRegionType) {
	var c *color.Color
	switch regionType {
	case debugger.RegionCode:
		c = colorMemCode
	case debugger.RegionData:
		c = colorMemData
	case debugger.RegionStack:
		c = colorMemStack
	default:
		c = colorMemUnknown
	}
	c.Printf("%02X ", b)
}
