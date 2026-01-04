package debugger

import (
	"strconv"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/interpreter"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
)

// Controller coordinates between the debugger backend and UI.
// It implements the command processing logic while delegating
// presentation to the UI interface.
type Controller struct {
	backend       *Backend
	ui            DebuggerUI
	running       bool
	lastSourceLoc *mc.SourceLocation
	lastCommand   string
	selectedFrame int // Currently selected stack frame (0 = current/innermost)
}

// NewController creates a new debugger controller
func NewController(backend *Backend, ui DebuggerUI) *Controller {
	return &Controller{
		backend: backend,
		ui:      ui,
		running: true,
	}
}

// Backend returns the underlying backend
func (c *Controller) Backend() *Backend {
	return c.backend
}

// UI returns the UI interface
func (c *Controller) UI() DebuggerUI {
	return c.ui
}

// IsRunning returns true if the debugger session is active
func (c *Controller) IsRunning() bool {
	return c.running
}

// Stop stops the debugger session
func (c *Controller) Stop() {
	c.running = false
}

// SetLastCommand sets the last command (for command repetition)
func (c *Controller) SetLastCommand(cmd string) {
	c.lastCommand = cmd
}

// LastCommand returns the last command
func (c *Controller) LastCommand() string {
	return c.lastCommand
}

// SelectedFrame returns the currently selected stack frame index (0 = innermost/current)
func (c *Controller) SelectedFrame() int {
	return c.selectedFrame
}

// resetSelectedFrame resets the selected frame to the current (innermost) frame.
// Called after any execution (step, continue, run, etc.)
func (c *Controller) resetSelectedFrame() {
	c.selectedFrame = 0
}

// --- Command Implementations ---

// CmdStep steps through source lines when debug info is available.
// If no debug info, steps by single instruction.
// When stepping through source lines, shows instruction traces for each instruction.
func (c *Controller) CmdStep(count int) {
	c.resetSelectedFrame() // Reset frame selection on execution
	if count <= 0 {
		count = 1
	}

	// Check if we have debug info for source-level stepping
	debugInfo := c.backend.DebugInfo()
	hasDebugInfo := debugInfo != nil && len(debugInfo.InstructionLocations) > 0

	if !hasDebugInfo {
		// No debug info: fall back to instruction stepping
		c.CmdInstructionStep(count)
		return
	}

	// Source-level stepping
	for i := 0; i < count; i++ {
		if !c.stepSourceLine() {
			return
		}
	}
}

// CmdNext steps over function calls at source level (like step but doesn't enter functions)
func (c *Controller) CmdNext(count int) {
	c.resetSelectedFrame() // Reset frame selection on execution
	if count <= 0 {
		count = 1
	}

	// Check if we have debug info for source-level stepping
	debugInfo := c.backend.DebugInfo()
	hasDebugInfo := debugInfo != nil && len(debugInfo.InstructionLocations) > 0

	if !hasDebugInfo {
		// No debug info: fall back to instruction-level next
		c.CmdInstructionNext(count)
		return
	}

	// Source-level next (step over)
	for i := 0; i < count; i++ {
		if !c.nextSourceLine() {
			return
		}
	}
}

// nextSourceLine executes instructions until the source line changes, stepping over calls.
// Returns false if execution should stop (termination, breakpoint, etc.)
func (c *Controller) nextSourceLine() bool {
	// Get current source location
	state := c.backend.GetState()
	startLoc := c.backend.GetSourceLocation(state.PC)

	var startFile string
	var startLine int

	if startLoc != nil && startLoc.IsValid() {
		startFile = startLoc.File
		startLine = startLoc.Line
	}

	// Step over until source line changes
	const maxInstructions = 10000 // Safety limit
	for step := 0; step < maxInstructions; step++ {
		// Execute one "next" (step over calls)
		if !c.nextOneInstruction(true) {
			return false
		}

		// Check if source line changed
		state = c.backend.GetState()
		currentLoc := c.backend.GetSourceLocation(state.PC)

		if startLoc == nil || !startLoc.IsValid() {
			// We started without source info - stop when we find source info
			if currentLoc != nil && currentLoc.IsValid() {
				c.checkSourceLocationChange()
				return true
			}
			continue
		}

		// We started with source info - stop when it changes
		if currentLoc == nil || !currentLoc.IsValid() {
			continue
		}

		if currentLoc.File != startFile || currentLoc.Line != startLine {
			c.checkSourceLocationChange()
			return true
		}
	}

	c.ui.ShowMessage(LevelWarning, "Stepped %d instructions without source line change", maxInstructions)
	return true
}

// nextOneInstruction executes a single "next" (step over calls).
// If showTrace is true, shows the instruction after stepping.
// Returns false if execution should stop.
func (c *Controller) nextOneInstruction(showTrace bool) bool {
	result := c.backend.Next(1)

	if result.Error != nil {
		c.ui.OnEvent(EventData{
			Event:   EventError,
			Error:   result.Error,
			Message: result.Error.Error(),
		})
		return false
	}

	switch result.StopReason {
	case interpreter.StopTermination:
		c.ui.OnEvent(EventData{
			Event:       EventProgramTerminated,
			ReturnValue: result.ReturnValue,
		})
		return false

	case interpreter.StopHalt:
		c.ui.OnEvent(EventData{
			Event: EventProgramHalted,
		})
		return false

	case interpreter.StopBreakpoint:
		c.ui.OnEvent(EventData{
			Event:        EventBreakpointHit,
			Address:      result.LastPC,
			BreakpointID: result.BreakpointID,
		})
		c.showCurrentInstruction()
		return false

	case interpreter.StopWatchpoint:
		c.ui.OnEvent(EventData{
			Event:        EventWatchpointHit,
			Address:      result.LastPC,
			WatchpointID: result.WatchpointID,
		})
		c.showCurrentInstruction()
		return false
	}

	if showTrace {
		c.showCurrentInstructionOnly()
	}

	return true
}

// CmdInstructionNext steps over function calls at instruction level
func (c *Controller) CmdInstructionNext(count int) {
	c.resetSelectedFrame() // Reset frame selection on execution
	if count <= 0 {
		count = 1
	}

	for i := 0; i < count; i++ {
		result := c.backend.Next(1)

		if result.Error != nil {
			c.ui.ShowMessage(LevelError, "Error: %v", result.Error)
			return
		}

		switch result.StopReason {
		case interpreter.StopTermination:
			c.ui.OnEvent(EventData{
				Event:         EventProgramTerminated,
				ReturnValue:   result.ReturnValue,
				StepsExecuted: result.StepsExecuted,
			})
			return

		case interpreter.StopBreakpoint:
			c.ui.OnEvent(EventData{
				Event:         EventBreakpointHit,
				Address:       result.LastPC,
				BreakpointID:  result.BreakpointID,
				StepsExecuted: result.StepsExecuted,
			})
			c.showCurrentInstruction()
			return

		case interpreter.StopWatchpoint:
			c.ui.OnEvent(EventData{
				Event:         EventWatchpointHit,
				Address:       result.LastPC,
				WatchpointID:  result.WatchpointID,
				StepsExecuted: result.StepsExecuted,
			})
			c.showCurrentInstruction()
			return
		}

		// Show source location change during multi-step
		if count > 1 {
			c.checkSourceLocationChange()
		}
	}

	c.showCurrentInstruction()
}

// stepSourceLine executes instructions until the source line changes.
// Returns false if execution should stop (termination, breakpoint, etc.)
func (c *Controller) stepSourceLine() bool {
	// Get current source location
	state := c.backend.GetState()
	startLoc := c.backend.GetSourceLocation(state.PC)

	var startFile string
	var startLine int

	if startLoc != nil && startLoc.IsValid() {
		startFile = startLoc.File
		startLine = startLoc.Line
	}

	// Step until source line changes (or we find source info if we don't have any)
	const maxInstructions = 10000 // Safety limit
	for step := 0; step < maxInstructions; step++ {
		// Execute one instruction and show trace
		if !c.stepOneInstruction(true) {
			return false
		}

		// Check if source line changed
		state = c.backend.GetState()
		currentLoc := c.backend.GetSourceLocation(state.PC)

		if startLoc == nil || !startLoc.IsValid() {
			// We started without source info - stop when we find source info
			if currentLoc != nil && currentLoc.IsValid() {
				c.checkSourceLocationChange()
				return true
			}
			// Keep stepping until we find source info
			continue
		}

		// We started with source info - stop when it changes
		if currentLoc == nil || !currentLoc.IsValid() {
			// Lost source info - keep stepping to find it again
			continue
		}

		if currentLoc.File != startFile || currentLoc.Line != startLine {
			// Source line changed - show source location (instruction already shown)
			c.checkSourceLocationChange()
			return true
		}
	}

	c.ui.ShowMessage(LevelWarning, "Stepped %d instructions without source line change", maxInstructions)
	return true
}

// stepOneInstruction executes a single instruction.
// If showTrace is true, shows the instruction after stepping.
// Returns false if execution should stop.
func (c *Controller) stepOneInstruction(showTrace bool) bool {
	result := c.backend.Step(1)

	if result.Error != nil {
		c.ui.OnEvent(EventData{
			Event:   EventError,
			Error:   result.Error,
			Message: result.Error.Error(),
		})
		return false
	}

	switch result.StopReason {
	case interpreter.StopTermination:
		c.ui.OnEvent(EventData{
			Event:       EventProgramTerminated,
			ReturnValue: result.ReturnValue,
		})
		return false

	case interpreter.StopHalt:
		c.ui.OnEvent(EventData{
			Event: EventProgramHalted,
		})
		return false

	case interpreter.StopBreakpoint:
		c.ui.OnEvent(EventData{
			Event:        EventBreakpointHit,
			Address:      result.LastPC,
			BreakpointID: result.BreakpointID,
		})
		c.showCurrentInstruction()
		return false

	case interpreter.StopWatchpoint:
		c.ui.OnEvent(EventData{
			Event:        EventWatchpointHit,
			Address:      result.LastPC,
			WatchpointID: result.WatchpointID,
		})
		c.showCurrentInstruction()
		return false
	}

	if showTrace {
		c.showCurrentInstructionOnly()
	}

	return true
}

// CmdInstructionStep executes exactly n instructions (ignores source-level stepping)
func (c *Controller) CmdInstructionStep(count int) {
	c.resetSelectedFrame() // Reset frame selection on execution
	if count <= 0 {
		count = 1
	}

	for i := 0; i < count; i++ {
		if !c.stepOneInstruction(false) {
			return
		}

		// Show source location change during multi-step
		if count > 1 {
			c.checkSourceLocationChange()
		}
	}

	c.showCurrentInstruction()
}

// CmdContinue continues execution until breakpoint or termination
func (c *Controller) CmdContinue() {
	c.resetSelectedFrame() // Reset frame selection on execution

	// Set up callback to handle lagging events during execution
	laggingWarningShown := false
	c.backend.SetFullExecutionCallback(func(event interpreter.ExecutionEvent, result *interpreter.ExecutionResult) bool {
		if event == interpreter.EventLagging && !laggingWarningShown {
			c.ui.OnEvent(EventData{
				Event:         EventLagging,
				LagCycles:     result.LagCycles,
				TargetSpeedHz: c.backend.GetTargetSpeed(),
			})
			laggingWarningShown = true
		}
		return true // Continue execution
	})

	result := c.backend.Continue()

	// Clear callback
	c.backend.SetFullExecutionCallback(nil)

	// Populate lagging info in final event
	cyclesExecuted := result.CyclesExecuted
	lagging := result.Lagging
	lagCycles := result.LagCycles

	switch result.StopReason {
	case interpreter.StopTermination:
		c.ui.OnEvent(EventData{
			Event:          EventProgramTerminated,
			ReturnValue:    result.ReturnValue,
			StepsExecuted:  result.StepsExecuted,
			CyclesExecuted: cyclesExecuted,
		})

	case interpreter.StopBreakpoint:
		c.ui.OnEvent(EventData{
			Event:          EventBreakpointHit,
			Address:        result.LastPC,
			BreakpointID:   result.BreakpointID,
			StepsExecuted:  result.StepsExecuted,
			CyclesExecuted: cyclesExecuted,
		})
		c.showCurrentInstruction()

	case interpreter.StopWatchpoint:
		c.ui.OnEvent(EventData{
			Event:          EventWatchpointHit,
			Address:        result.LastPC,
			WatchpointID:   result.WatchpointID,
			StepsExecuted:  result.StepsExecuted,
			CyclesExecuted: cyclesExecuted,
		})
		c.showCurrentInstruction()

	case interpreter.StopHalt:
		c.ui.OnEvent(EventData{
			Event: EventProgramHalted,
		})

	case interpreter.StopInterrupt:
		c.ui.OnEvent(EventData{
			Event:          EventInterrupted,
			Address:        c.backend.GetState().PC,
			StepsExecuted:  result.StepsExecuted,
			CyclesExecuted: cyclesExecuted,
		})
		c.showCurrentInstruction()

	case interpreter.StopError:
		c.ui.OnEvent(EventData{
			Event:   EventError,
			Error:   result.Error,
			Message: result.Error.Error(),
		})
	}

	// Report lagging at the end if still lagging
	if lagging && !laggingWarningShown {
		c.ui.OnEvent(EventData{
			Event:         EventLagging,
			LagCycles:     lagCycles,
			TargetSpeedHz: c.backend.GetTargetSpeed(),
		})
	}
}

// CmdRun runs until termination. If program already terminated, asks to restart.
func (c *Controller) CmdRun() {
	c.resetSelectedFrame() // Reset frame selection on execution
	// Check if program already terminated - if so, ask to restart
	if c.backend.IsTerminated() {
		if !c.ui.PromptConfirm("Program already terminated. Restart?") {
			return
		}
		if err := c.backend.Reset(); err != nil {
			c.ui.ShowMessage(LevelError, "Failed to reset: %v", err)
			return
		}
	}

	// Set up callback to handle lagging events during execution
	laggingWarningShown := false
	c.backend.SetFullExecutionCallback(func(event interpreter.ExecutionEvent, result *interpreter.ExecutionResult) bool {
		if event == interpreter.EventLagging && !laggingWarningShown {
			c.ui.OnEvent(EventData{
				Event:         EventLagging,
				LagCycles:     result.LagCycles,
				TargetSpeedHz: c.backend.GetTargetSpeed(),
			})
			laggingWarningShown = true
		}
		return true // Continue execution
	})

	result := c.backend.Run()

	// Clear callback
	c.backend.SetFullExecutionCallback(nil)

	// Populate lagging info
	cyclesExecuted := result.CyclesExecuted
	lagging := result.Lagging
	lagCycles := result.LagCycles

	switch result.StopReason {
	case interpreter.StopTermination:
		c.ui.OnEvent(EventData{
			Event:          EventProgramTerminated,
			ReturnValue:    result.ReturnValue,
			StepsExecuted:  result.StepsExecuted,
			CyclesExecuted: cyclesExecuted,
		})

	case interpreter.StopBreakpoint:
		c.ui.OnEvent(EventData{
			Event:          EventBreakpointHit,
			Address:        result.LastPC,
			BreakpointID:   result.BreakpointID,
			StepsExecuted:  result.StepsExecuted,
			CyclesExecuted: cyclesExecuted,
		})
		c.showCurrentInstruction()

	case interpreter.StopWatchpoint:
		c.ui.OnEvent(EventData{
			Event:          EventWatchpointHit,
			Address:        result.LastPC,
			WatchpointID:   result.WatchpointID,
			StepsExecuted:  result.StepsExecuted,
			CyclesExecuted: cyclesExecuted,
		})
		c.showCurrentInstruction()

	case interpreter.StopHalt:
		c.ui.OnEvent(EventData{
			Event: EventProgramHalted,
		})

	case interpreter.StopInterrupt:
		c.ui.OnEvent(EventData{
			Event:          EventInterrupted,
			Address:        c.backend.GetState().PC,
			StepsExecuted:  result.StepsExecuted,
			CyclesExecuted: cyclesExecuted,
		})
		c.showCurrentInstruction()

	case interpreter.StopError:
		c.ui.OnEvent(EventData{
			Event:   EventError,
			Error:   result.Error,
			Message: result.Error.Error(),
		})
	}

	// Report lagging at the end if still lagging
	if lagging && !laggingWarningShown {
		c.ui.OnEvent(EventData{
			Event:         EventLagging,
			LagCycles:     lagCycles,
			TargetSpeedHz: c.backend.GetTargetSpeed(),
		})
	}
}

// CmdBreak adds a breakpoint at the given address
func (c *Controller) CmdBreak(addr uint32) {
	bp, err := c.backend.AddBreakpoint(addr)
	if err != nil {
		c.ui.ShowMessage(LevelError, "Failed to add breakpoint: %v", err)
		return
	}
	c.ui.ShowMessage(LevelSuccess, "Breakpoint %d set at 0x%08X", bp.ID, addr)
}

// CmdWatch adds a watchpoint at the given address
func (c *Controller) CmdWatch(addr uint32) {
	wp, err := c.backend.AddWatchpoint(addr)
	if err != nil {
		c.ui.ShowMessage(LevelError, "Failed to add watchpoint: %v", err)
		return
	}
	c.ui.ShowMessage(LevelSuccess, "Watchpoint %d set at 0x%08X", wp.ID, addr)
}

// CmdDelete removes a breakpoint or watchpoint by ID
func (c *Controller) CmdDelete(id int) {
	// Try breakpoint first
	if err := c.backend.RemoveBreakpoint(id); err == nil {
		c.ui.ShowMessage(LevelSuccess, "Breakpoint %d deleted", id)
		return
	}

	// Try watchpoint
	if err := c.backend.RemoveWatchpoint(id); err == nil {
		c.ui.ShowMessage(LevelSuccess, "Watchpoint %d deleted", id)
		return
	}

	c.ui.ShowMessage(LevelError, "No breakpoint or watchpoint with ID %d", id)
}

// CmdList lists all breakpoints and watchpoints
func (c *Controller) CmdList() {
	bps := c.backend.GetBreakpointInfos()
	wps := c.backend.GetWatchpointInfos()

	if len(bps) == 0 && len(wps) == 0 {
		c.ui.ShowMessage(LevelInfo, "No breakpoints or watchpoints set.")
		return
	}

	c.ui.ShowBreakpoints(bps)
	c.ui.ShowWatchpoints(wps)
}

// CmdPrint prints a register or memory value
func (c *Controller) CmdPrint(what string) {
	// Try as register
	if val, err := c.backend.ReadRegister(what); err == nil {
		c.ui.ShowEvalResult(what, val, nil)
		return
	}

	// Try evaluating as expression
	val, err := c.backend.EvalExpression(what)
	c.ui.ShowEvalResult(what, val, err)
}

// CmdSet sets a register value
func (c *Controller) CmdSet(regName string, value uint32) {
	if err := c.backend.WriteRegister(regName, value); err != nil {
		c.ui.ShowMessage(LevelError, "Failed to set %s: %v", regName, err)
		return
	}
	c.ui.ShowMessage(LevelSuccess, "%s = 0x%08X", regName, value)
}

// CmdDisasm disassembles instructions
func (c *Controller) CmdDisasm(addr uint32, count int) {
	if count <= 0 {
		count = 10
	}

	instructions, err := c.backend.Disassemble(addr, count)
	if err != nil {
		c.ui.ShowMessage(LevelError, "Failed to disassemble: %v", err)
		return
	}

	state := c.backend.GetState()
	c.ui.ShowDisassembly(instructions, state.PC)
}

// CmdInfo shows CPU state
func (c *Controller) CmdInfo() {
	state := c.backend.GetState()
	c.ui.ShowRegisters(state.Registers, state.Flags)
}

// CmdStack shows stack contents and call frames
func (c *Controller) CmdStack() {
	state := c.backend.GetState()

	// Get call stack frames (this doesn't read raw memory, just uses debug info)
	frames := c.backend.GetStackFrames()

	// Try to read stack memory, handling boundary conditions
	// Stack grows downward, so SP may be near top of memory
	stackSize := 64
	memSize := len(c.backend.Runner().State().Memory)

	// Calculate how much we can safely read
	if int(state.SP)+stackSize > memSize {
		stackSize = memSize - int(state.SP)
	}
	if stackSize < 0 {
		stackSize = 0
	}

	var data []byte
	var err error
	if stackSize > 0 {
		data, err = c.backend.ReadMemory(state.SP, stackSize)
		if err != nil {
			// Just show frames without raw stack data
			data = nil
		}
	}

	c.ui.ShowStack(state.SP, data, frames)
}

// CmdBacktrace shows the call stack (function frames)
func (c *Controller) CmdBacktrace() {
	frames := c.backend.GetCallStack()
	c.ui.ShowBacktrace(frames, c.selectedFrame)
}

// CmdMemory shows memory contents
func (c *Controller) CmdMemory(addr uint32, count int) {
	if count <= 0 {
		count = 64
	}

	data, err := c.backend.ReadMemory(addr, count)
	if err != nil {
		c.ui.ShowMessage(LevelError, "Failed to read memory at 0x%08X: %v", addr, err)
		return
	}

	regions := c.getMemoryRegions()
	c.ui.ShowMemory(addr, data, regions)
}

// CmdSource shows source code around current location or selected frame
func (c *Controller) CmdSource(lines int) {
	if lines <= 0 {
		lines = 10
	}

	// Get location for the selected frame
	pc := c.getSelectedFramePC()
	loc := c.backend.GetSourceLocation(pc)
	if loc == nil {
		if c.selectedFrame > 0 {
			c.ui.ShowMessage(LevelWarning, "No source information available at frame #%d", c.selectedFrame)
		} else {
			c.ui.ShowMessage(LevelWarning, "No source information available at current location")
		}
		return
	}

	debugInfo := c.backend.DebugInfo()
	if debugInfo == nil {
		c.ui.ShowMessage(LevelWarning, "No debug information available")
		return
	}

	// Get source lines
	sourceLines := make([]SourceLine, 0)
	startLine := loc.Line - lines/2
	if startLine < 1 {
		startLine = 1
	}

	for lineNum := startLine; lineNum < startLine+lines; lineNum++ {
		text := debugInfo.GetSourceLine(loc.File, lineNum)
		if text == "" && lineNum > loc.Line+lines/2 {
			break
		}

		sourceLines = append(sourceLines, SourceLine{
			LineNumber: lineNum,
			Text:       text,
			IsCurrent:  lineNum == loc.Line,
		})
	}

	c.ui.ShowSource(loc, sourceLines, loc.Line)
}

// CmdVars shows accessible variables
// Uses the currently selected frame's PC for variable lookup
func (c *Controller) CmdVars() {
	// Get the PC for the selected frame
	pc := c.getSelectedFramePC()

	vars := c.backend.GetVariables(pc)

	if len(vars) == 0 {
		if c.selectedFrame > 0 {
			c.ui.ShowMessage(LevelInfo, "No variables accessible at frame #%d.", c.selectedFrame)
		} else {
			c.ui.ShowMessage(LevelInfo, "No variables accessible at current location.")
		}
		return
	}

	c.ui.ShowVariables(vars)
}

// getSelectedFramePC returns the PC for the currently selected frame
func (c *Controller) getSelectedFramePC() uint32 {
	if c.selectedFrame == 0 {
		return c.backend.GetState().PC
	}

	frames := c.backend.GetCallStack()
	if c.selectedFrame < len(frames) {
		return frames[c.selectedFrame].Address
	}

	// Fall back to current PC if frame index is out of range
	return c.backend.GetState().PC
}

// CmdUp moves up the call stack (to older/caller frames)
func (c *Controller) CmdUp(count int) {
	if count <= 0 {
		count = 1
	}

	frames := c.backend.GetCallStack()
	if len(frames) == 0 {
		c.ui.ShowMessage(LevelWarning, "No call stack available")
		return
	}

	newFrame := c.selectedFrame + count
	if newFrame >= len(frames) {
		newFrame = len(frames) - 1
		if c.selectedFrame == newFrame {
			c.ui.ShowMessage(LevelWarning, "Already at outermost frame")
			return
		}
	}

	c.selectedFrame = newFrame
	c.showSelectedFrame(frames)
}

// CmdDown moves down the call stack (to newer/callee frames)
func (c *Controller) CmdDown(count int) {
	if count <= 0 {
		count = 1
	}

	if c.selectedFrame == 0 {
		c.ui.ShowMessage(LevelWarning, "Already at innermost frame")
		return
	}

	newFrame := c.selectedFrame - count
	if newFrame < 0 {
		newFrame = 0
	}

	c.selectedFrame = newFrame
	frames := c.backend.GetCallStack()
	c.showSelectedFrame(frames)
}

// CmdFrame selects a specific frame by number
func (c *Controller) CmdFrame(frameNum int) {
	frames := c.backend.GetCallStack()
	if len(frames) == 0 {
		c.ui.ShowMessage(LevelWarning, "No call stack available")
		return
	}

	if frameNum < 0 || frameNum >= len(frames) {
		c.ui.ShowMessage(LevelError, "Frame %d out of range (0-%d)", frameNum, len(frames)-1)
		return
	}

	c.selectedFrame = frameNum
	c.showSelectedFrame(frames)
}

// showSelectedFrame displays information about the currently selected frame
func (c *Controller) showSelectedFrame(frames []StackFrame) {
	if c.selectedFrame >= len(frames) {
		return
	}

	frame := frames[c.selectedFrame]

	// Get source line if available
	var srcText string
	if frame.File != "" && frame.Line > 0 {
		if debugInfo := c.backend.DebugInfo(); debugInfo != nil {
			srcText = debugInfo.GetSourceLine(frame.File, frame.Line)
		}
	}

	c.ui.ShowFrameInfo(frame, c.selectedFrame, srcText)
}

// CmdEval evaluates an expression
func (c *Controller) CmdEval(expr string) {
	val, err := c.backend.EvalExpression(expr)
	c.ui.ShowEvalResult(expr, val, err)
}

// CmdHelp shows help information
func (c *Controller) CmdHelp() {
	commands := []CommandHelp{
		{Name: "step", Aliases: []string{"s"}, Description: "Step to next source line", Usage: "step [n]"},
		{Name: "stepi", Aliases: []string{"si"}, Description: "Step n instructions", Usage: "stepi [n]"},
		{Name: "continue", Aliases: []string{"c"}, Description: "Continue until breakpoint", Usage: "continue"},
		{Name: "run", Aliases: []string{"r"}, Description: "Run until termination", Usage: "run"},
		{Name: "break", Aliases: []string{"b"}, Description: "Set breakpoint", Usage: "break <addr>"},
		{Name: "watch", Aliases: []string{"w"}, Description: "Set watchpoint", Usage: "watch <addr>"},
		{Name: "delete", Aliases: []string{"d"}, Description: "Delete breakpoint/watchpoint", Usage: "delete <id>"},
		{Name: "list", Aliases: []string{"l"}, Description: "List breakpoints/watchpoints", Usage: "list"},
		{Name: "print", Aliases: []string{"p"}, Description: "Print register or memory", Usage: "print <what>"},
		{Name: "set", Aliases: nil, Description: "Set register value", Usage: "set <reg> <value>"},
		{Name: "disasm", Aliases: []string{"x"}, Description: "Disassemble", Usage: "disasm [addr] [n]"},
		{Name: "info", Aliases: []string{"i"}, Description: "Show CPU state", Usage: "info"},
		{Name: "stack", Aliases: nil, Description: "Show stack", Usage: "stack"},
		{Name: "memory", Aliases: []string{"m"}, Description: "Show memory", Usage: "memory <addr> [, count]"},
		{Name: "source", Aliases: []string{"src"}, Description: "Show source code", Usage: "source [lines]"},
		{Name: "vars", Aliases: []string{"v"}, Description: "Show variables", Usage: "vars"},
		{Name: "eval", Aliases: []string{"e"}, Description: "Evaluate expression", Usage: "eval <expr>"},
		{Name: "help", Aliases: []string{"h", "?"}, Description: "Show help", Usage: "help"},
		{Name: "quit", Aliases: []string{"q", "exit"}, Description: "Exit debugger", Usage: "quit"},
	}
	c.ui.ShowHelp(commands)
}

// CmdQuit exits the debugger
func (c *Controller) CmdQuit() {
	c.running = false
	c.ui.ShowMessage(LevelSuccess, "Exiting debugger.")
}

// --- Helper Methods ---

// showCurrentInstruction shows the current instruction with source location check
func (c *Controller) showCurrentInstruction() {
	// Check for source location change first
	c.checkSourceLocationChange()
	// Then show the instruction
	c.showCurrentInstructionOnly()
}

// showCurrentInstructionOnly shows just the current instruction without source location check
func (c *Controller) showCurrentInstructionOnly() {
	state := c.backend.GetState()

	// Get instruction info
	instructions, err := c.backend.Disassemble(state.PC, 1)
	if err != nil || len(instructions) == 0 {
		c.ui.ShowMessage(LevelWarning, "Cannot disassemble at 0x%08X", state.PC)
		return
	}

	c.ui.ShowInstruction(instructions[0])
}

// checkSourceLocationChange checks if source location changed and shows it
func (c *Controller) checkSourceLocationChange() {
	state := c.backend.GetState()
	loc := c.backend.GetSourceLocation(state.PC)

	if loc == nil || !loc.IsValid() {
		return
	}

	// Check if location changed
	if c.lastSourceLoc != nil &&
		c.lastSourceLoc.File == loc.File &&
		c.lastSourceLoc.Line == loc.Line {
		return
	}

	c.lastSourceLoc = loc

	// Get the source text for this line
	sourceText := ""
	debugInfo := c.backend.DebugInfo()
	if debugInfo != nil {
		sourceText = debugInfo.GetSourceLine(loc.File, loc.Line)
	}

	c.ui.OnEvent(EventData{
		Event:          EventSourceLocationChanged,
		Address:        state.PC,
		SourceLocation: loc,
		SourceText:     sourceText,
	})
}

// getMemoryRegions returns memory region information
func (c *Controller) getMemoryRegions() []MemoryRegion {
	program := c.backend.Program()
	if program == nil {
		return nil
	}

	layout := program.MemoryLayout()
	state := c.backend.GetState()

	regions := []MemoryRegion{
		{
			Name:       "Code",
			StartAddr:  layout.CodeStart,
			EndAddr:    layout.CodeStart + layout.CodeSize,
			RegionType: RegionCode,
		},
	}

	if layout.DataSize > 0 {
		regions = append(regions, MemoryRegion{
			Name:       "Data",
			StartAddr:  layout.DataStart,
			EndAddr:    layout.DataStart + layout.DataSize,
			RegionType: RegionData,
		})
	}

	// Add stack region (grows down from SP)
	regions = append(regions, MemoryRegion{
		Name:       "Stack",
		StartAddr:  state.SP,
		EndAddr:    state.SP + 1024, // Assume 1KB visible
		RegionType: RegionStack,
	})

	return regions
}

// ResolveAddress resolves an address expression (for command parsing)
// Supports:
//   - Hex addresses: 0x10000
//   - Decimal: 65536
//   - Symbols: main, loop
//   - Registers: pc, sp, $r0
//   - Source locations: file.c:10, :10 (current file)
func (c *Controller) ResolveAddress(expr string) (uint32, error) {
	expr = strings.TrimSpace(expr)

	// Check for source location syntax (file:line or :line)
	if strings.Contains(expr, ":") {
		parts := strings.SplitN(expr, ":", 2)
		if len(parts) == 2 {
			file := strings.TrimSpace(parts[0])
			lineStr := strings.TrimSpace(parts[1])

			// Try to parse the line number
			line, err := strconv.Atoi(lineStr)
			if err == nil && line > 0 {
				// This looks like a source location
				return c.backend.ResolveSourceLocation(file, line)
			}
		}
	}

	// Fall back to expression evaluation
	return c.backend.EvalExpression(expr)
}
