package widgets

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/utils/logging"
	"github.com/Manu343726/cucaracha/pkg/ui"
	"github.com/Manu343726/cucaracha/pkg/ui/tui/tview/themes"
	"github.com/gdamore/tcell/v2"
	tvlib "github.com/rivo/tview"
)

// parseAddress parses a hex or decimal address string
func parseAddress(s string) uint32 {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		val, _ := strconv.ParseUint(s[2:], 16, 32)
		return uint32(val)
	}
	val, _ := strconv.ParseUint(s, 10, 32)
	return uint32(val)
}

// parseID parses an integer ID
func parseID(s string) int {
	val, _ := strconv.Atoi(s)
	return val
}

// CommandCallback is called when a command is executed
type CommandCallback func(result *ui.DebuggerCommandResult)

// CommandExecutor executes commands in the debugger
type CommandExecutor interface {
	Execute(cmd *ui.DebuggerCommand, callback ui.AsyncDebuggerCommandResultCallback)
}

// CommandInfo describes a debugger command
type CommandInfo struct {
	Name        string
	Description string
	Syntax      string
	Example     string
}

// CommandsTerminal is a widget for executing debugger commands
type CommandsTerminal struct {
	*tvlib.Flex
	app                 *tvlib.Application
	outputView          *tvlib.TextView
	inputView           *tvlib.InputField
	input               string
	history             []string
	historyIdx          int
	executor            CommandExecutor
	commandCallbacks    map[ui.DebuggerCommandId][]CommandCallback
	outputLines         []string
	commandDescriptions map[string]CommandInfo
	pendingCommand      *ui.DebuggerCommand
	filePickerCallback  func(action string)
	themeCallback       func(string)
	focusable           bool
	theme               *themes.Theme
}

// NewCommandsTerminal creates a new CommandsTerminal widget
func NewCommandsTerminal(executor CommandExecutor, app *tvlib.Application) *CommandsTerminal {
	flex := tvlib.NewFlex().SetDirection(tvlib.FlexRow)

	// Output view
	outputView := tvlib.NewTextView()
	outputView.SetBorder(true)
	outputView.SetTitle("Commands")
	outputView.SetDynamicColors(true)
	outputView.SetScrollable(true)

	// Input field
	inputView := tvlib.NewInputField()
	inputView.SetLabel("[#AE81FF]> [-]")

	ct := &CommandsTerminal{
		Flex:                flex,
		app:                 app,
		outputView:          outputView,
		inputView:           inputView,
		input:               "",
		history:             []string{},
		historyIdx:          -1,
		executor:            executor,
		commandCallbacks:    make(map[ui.DebuggerCommandId][]CommandCallback),
		outputLines:         []string{},
		commandDescriptions: initCommandDescriptions(),
		pendingCommand:      nil,
		filePickerCallback:  nil,
		themeCallback:       nil,
		focusable:           true,
		theme:               nil,
	}

	// Set up input field with command execution
	inputView.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			text := inputView.GetText()
			if text != "" {
				ct.HandleInput(text)
				inputView.SetText("")
			}
		}
	})

	// Set up input field autocomplete/history
	inputView.SetAutocompleteFunc(func(currentText string) (entries []string) {
		// Return matching history
		for _, h := range ct.history {
			if strings.HasPrefix(h, currentText) {
				entries = append(entries, h)
			}
		}
		return
	})

	// Layout: output on top, input on bottom
	flex.AddItem(outputView, 0, 1, false)
	flex.AddItem(inputView, 1, 0, true)

	return ct
}

func (ct *CommandsTerminal) log() *logging.Logger {
	return log().Child("commandsterminal")
}

// initCommandDescriptions initializes the command descriptions
func initCommandDescriptions() map[string]CommandInfo {
	return map[string]CommandInfo{
		"step":                {Name: "step", Description: "Step through execution", Syntax: "step [mode] [count]", Example: "step into 1"},
		"stepi":               {Name: "stepi", Description: "Step one instruction", Syntax: "stepi", Example: "stepi"},
		"continue":            {Name: "continue", Description: "Continue execution until next breakpoint", Syntax: "continue", Example: "continue"},
		"interrupt":           {Name: "interrupt", Description: "Interrupt execution", Syntax: "interrupt", Example: "interrupt"},
		"break":               {Name: "break", Description: "Set a breakpoint", Syntax: "break <address>", Example: "break 0x1000"},
		"removeBreakpoint":    {Name: "removeBreakpoint", Description: "Remove a breakpoint", Syntax: "removeBreakpoint <id>", Example: "removeBreakpoint 1"},
		"watch":               {Name: "watch", Description: "Set a watchpoint", Syntax: "watch <address> <size>", Example: "watch 0x2000 4"},
		"removeWatchpoint":    {Name: "removeWatchpoint", Description: "Remove a watchpoint", Syntax: "removeWatchpoint <id>", Example: "removeWatchpoint 1"},
		"list":                {Name: "list", Description: "List breakpoints and watchpoints", Syntax: "list", Example: "list"},
		"disassemble":         {Name: "disassemble", Description: "Disassemble instructions", Syntax: "disassemble [address] [count]", Example: "disassemble 0x1000 10"},
		"currentInstruction":  {Name: "currentInstruction", Description: "Show current instruction", Syntax: "currentInstruction", Example: "currentInstruction"},
		"memory":              {Name: "memory", Description: "Display memory contents", Syntax: "memory <address> [count]", Example: "memory 0x2000 256"},
		"source":              {Name: "source", Description: "Display source code", Syntax: "source <file> [line]", Example: "source main.c 10"},
		"currentSource":       {Name: "currentSource", Description: "Display current source code", Syntax: "currentSource", Example: "currentSource"},
		"evaluateExpression":  {Name: "evaluateExpression", Description: "Evaluate an expression", Syntax: "evaluateExpression <expr>", Example: "evaluateExpression x + 5"},
		"info":                {Name: "info", Description: "Show debugger info", Syntax: "info", Example: "info"},
		"registers":           {Name: "registers", Description: "Show CPU registers", Syntax: "registers", Example: "registers"},
		"stack":               {Name: "stack", Description: "Show call stack", Syntax: "stack", Example: "stack"},
		"variables":           {Name: "variables", Description: "Show variables", Syntax: "variables", Example: "variables"},
		"loadProgramFromFile": {Name: "loadProgramFromFile", Description: "Load program from file", Syntax: "loadProgramFromFile <file>", Example: "loadProgramFromFile program.bin"},
		"load":                {Name: "load", Description: "Load program, system, and runtime from YAML file", Syntax: "load <file>", Example: "load config.yaml"},
		"help":                {Name: "help", Description: "Show help", Syntax: "help [command]", Example: "help step"},
	}
}

// RegisterCommandCallback registers a callback for a command ID
func (ct *CommandsTerminal) RegisterCommandCallback(cmdID ui.DebuggerCommandId, callback CommandCallback) {
	ct.commandCallbacks[cmdID] = append(ct.commandCallbacks[cmdID], callback)
}

// RegisterCallback is an alias for RegisterCommandCallback for compatibility
func (ct *CommandsTerminal) RegisterCallback(cmdID ui.DebuggerCommandId, callback CommandCallback) {
	ct.RegisterCommandCallback(cmdID, callback)
}

// displayResult displays a command result
func (ct *CommandsTerminal) displayResult(result *ui.DebuggerCommandResult) {
	if ct.theme != nil {
		ct.displayOutput(ct.theme.FormatSuccess(result.Command.String()))
	} else {
		ct.displayOutput(result.Command.String())
	}
}

// executeCommand executes a parsed command
func (ct *CommandsTerminal) executeCommand(cmd *ui.DebuggerCommand) {
	if ct.executor == nil {
		if ct.theme != nil {
			ct.displayOutput(ct.theme.FormatError("No executor available"))
		} else {
			ct.displayOutput("No executor available")
		}
		return
	}

	log().Debug("executing command", "command", cmd.Command.String(), "args", cmd)

	ct.executor.Execute(cmd, func(result *ui.DebuggerCommandResult, err error) {
		ct.app.QueueUpdateDraw(func() {
			log().Debug("command executed", "command", cmd.Command.String(), "result", result, "error", err)

			if err != nil {
				if ct.theme != nil {
					ct.displayOutput(ct.theme.FormatError(fmt.Sprintf("Execution error: %v", err)))
				} else {
					ct.displayOutput(fmt.Sprintf("Execution error: %v", err))
				}
				return
			}

			ct.displayResult(result)

			// Call registered callbacks
			if callbacks, exists := ct.commandCallbacks[cmd.Command]; exists {
				for _, callback := range callbacks {
					if callback != nil {
						callback(result)
					}
				}
			}
		})
	})
}

// parseCommand parses a command string into a DebuggerCommand
func (ct *CommandsTerminal) parseCommand(input string) (*ui.DebuggerCommand, error) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil, fmt.Errorf("Invalid command '%s'", input)
	}

	cmdName := parts[0]
	cmdID, err := ui.DebuggerCommandIdFromString(cmdName)
	if err != nil {
		return nil, err
	}

	return ct.parseCommandArgs(cmdID, parts[1:])
}

func (ct *CommandsTerminal) parseStepCommandArgs(args []string) (*ui.DebuggerCommand, error) {
	return &ui.DebuggerCommand{
		Command: ui.DebuggerCommandStep,
		StepArgs: &ui.StepArgs{
			StepMode:  ui.StepModeInto,
			CountMode: ui.StepCountSourceLines,
		},
	}, nil
}

func (ct *CommandsTerminal) parseSourceCommandArgs(args []string) (*ui.DebuggerCommand, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("source command requires a file:line argument")
	}

	fileLine := strings.Split(args[0], ":")
	if len(fileLine) != 2 {
		return nil, fmt.Errorf("invalid source argument format, expected file:line")
	}

	lineNum, err := strconv.Atoi(fileLine[1])
	if err != nil {
		return nil, fmt.Errorf("invalid line number: %v", err)
	}

	sourceArgs := &ui.SourceArgs{
		File: fileLine[0],
		Line: lineNum,
	}

	if len(args) >= 2 {
		contextLines, err := strconv.Atoi(args[1])
		if err != nil {
			return nil, fmt.Errorf("invalid source context lines: %v", err)
		}
		sourceArgs.ContextLines = contextLines
	} else {
		sourceArgs.ContextLines = 20
	}

	if len(args) >= 3 {
		mode, err := ui.SourceContextModeFromString(args[2])
		if err != nil {
			return nil, fmt.Errorf("invalid source context mode: %v", err)
		}
		sourceArgs.ContextMode = mode
	} else {
		sourceArgs.ContextMode = ui.SourceContextTop
	}

	return &ui.DebuggerCommand{
		Command:    ui.DebuggerCommandSource,
		SourceArgs: sourceArgs,
	}, nil
}

func (ct *CommandsTerminal) parseCommandArgs(cmdID ui.DebuggerCommandId, args []string) (*ui.DebuggerCommand, error) {
	switch cmdID {
	case ui.DebuggerCommandStep:
		return ct.parseStepCommandArgs(args)
	case ui.DebuggerCommandSource:
		return ct.parseSourceCommandArgs(args)
	case ui.DebuggerCommandEvaluateExpression:
		return &ui.DebuggerCommand{
			Command: ui.DebuggerCommandEvaluateExpression,
			EvalArgs: &ui.EvalArgs{
				Expression: strings.Join(args, " "),
			},
		}, nil
	case ui.DebuggerCommandMemory:
		return &ui.DebuggerCommand{
			Command: ui.DebuggerCommandMemory,
			MemoryArgs: &ui.MemoryArgs{
				AddressExpr: strings.Join(args, " "),
			},
		}, nil
	default:
		return &ui.DebuggerCommand{
			Command: cmdID,
		}, nil
	}
}

// HandleInput processes input from user - can be overridden by TUI
func (ct *CommandsTerminal) HandleInput(input string) {
	ct.input = input
	if input == "" {
		return
	}

	// Add to history
	ct.history = append(ct.history, input)
	ct.historyIdx = -1

	// Display the command
	if ct.theme != nil {
		ct.displayOutput(ct.theme.FormatPrompt("> " + input))
	} else {
		ct.displayOutput("> " + input)
	}

	// Handle special commands
	if input == "theme" || input == "themes" {
		if ct.themeCallback != nil {
			ct.themeCallback("theme")
		}
		ct.input = ""
		return
	}

	if input == "help" || input == "?" {
		ct.displayHelp()
		ct.input = ""
		return
	}

	// Parse and execute regular debugger commands
	cmd, err := ct.parseCommand(input)
	if err != nil {
		ct.displayOutput(ct.theme.FormatError(err.Error()))
		return
	}

	ct.executeCommand(cmd)
	ct.input = ""
}

// refreshDisplay updates the output view content
func (ct *CommandsTerminal) refreshDisplay() {
	var text string
	for _, line := range ct.outputLines {
		text += line + "\n"
	}
	ct.outputView.SetText(text)
	// Auto-scroll to bottom
	ct.outputView.ScrollToEnd()
}

// displayOutput adds output lines and refreshes the display
func (ct *CommandsTerminal) displayOutput(text string) {
	ct.outputLines = append(ct.outputLines, text)
	// Limit output to last 100 lines
	if len(ct.outputLines) > 100 {
		ct.outputLines = ct.outputLines[len(ct.outputLines)-100:]
	}
	ct.refreshDisplay()
}

// AddOutput is an alias for displayOutput for compatibility
func (ct *CommandsTerminal) AddOutput(text string) {
	ct.displayOutput(text)
}

// displayHelp displays available commands
func (ct *CommandsTerminal) displayHelp() {
	ct.displayOutput("Available commands:")
	ct.displayOutput("")
	ct.displayOutput("Loading:")
	ct.displayOutput("  load <file>                 - Load program, system, and runtime from YAML")
	ct.displayOutput("  loadProgramFromFile <file>  - Load program from executable")
	ct.displayOutput("  loadSystemFromFile <file>   - Load system configuration")
	ct.displayOutput("  loadRuntime <file>          - Load runtime configuration")
	ct.displayOutput("")
	ct.displayOutput("Execution:")
	ct.displayOutput("  step                        - Execute one instruction")
	ct.displayOutput("  continue                    - Resume execution until breakpoint")
	ct.displayOutput("  interrupt                   - Stop execution")
	ct.displayOutput("")
	ct.displayOutput("Debugging:")
	ct.displayOutput("  setBreakpoint <addr>        - Set breakpoint at address")
	ct.displayOutput("  removeBreakpoint <id>       - Remove breakpoint by ID")
	ct.displayOutput("  setWatchpoint <addr> <len>  - Watch memory region")
	ct.displayOutput("  removeWatchpoint <id>       - Remove watchpoint by ID")
	ct.displayOutput("  list                        - Show breakpoints and watchpoints")
	ct.displayOutput("")
	ct.displayOutput("Inspection:")
	ct.displayOutput("  info                        - Show current debugger state")
	ct.displayOutput("  registers                   - Show all register values")
	ct.displayOutput("  stack                       - Show stack information")
	ct.displayOutput("  variables                   - Show accessible variables")
	ct.displayOutput("  memory <addr> [len]         - Display memory")
	ct.displayOutput("  disassemble [addr] [len]    - Show assembly instructions")
	ct.displayOutput("  source <file>               - Display source code")
	ct.displayOutput("  currentInstruction          - Show current instruction")
	ct.displayOutput("  currentSource               - Show source at current PC")
	ct.displayOutput("  evaluateExpression <expr>   - Evaluate expression")
	ct.displayOutput("")
	ct.displayOutput("UI:")
	ct.displayOutput("  theme / themes              - Show theme selector")
	ct.displayOutput("  help / ?                    - Show this help")
}

// SetFilePickerCallback sets the file picker callback
func (ct *CommandsTerminal) SetFilePickerCallback(callback func(action string)) {
	ct.filePickerCallback = callback
}

// SetThemeCallback sets the theme selector callback
func (ct *CommandsTerminal) SetThemeCallback(callback func(action string)) {
	ct.themeCallback = callback
}

// SetTheme applies the theme to the CommandsTerminal and its children recursively
func (ct *CommandsTerminal) SetTheme(theme *themes.Theme) *CommandsTerminal {
	ct.Flex.SetBackgroundColor(theme.PrimitiveBackgroundColor)
	ct.outputView.SetBackgroundColor(theme.PrimitiveBackgroundColor)
	ct.outputView.SetTextColor(theme.PrimaryTextColor)
	ct.outputView.SetBorderColor(theme.BorderColor)
	ct.inputView.SetFieldBackgroundColor(theme.PrimitiveBackgroundColor)
	ct.inputView.SetFieldTextColor(theme.PrimaryTextColor)
	ct.inputView.SetBorderColor(theme.BorderColor)
	ct.theme = theme
	return ct
}
