package tview

import (
	"fmt"
	"time"

	"github.com/Manu343726/cucaracha/pkg/ui"
	"github.com/Manu343726/cucaracha/pkg/ui/tui/tview/themes"
	"github.com/Manu343726/cucaracha/pkg/ui/tui/tview/widgets"
	"github.com/gdamore/tcell/v2"
	tvlib "github.com/rivo/tview"
)

// DebuggerTUI is the main tview implementation of the Cucaracha debugger TUI
type DebuggerTUI struct {
	debugger *ui.AsyncDebuggerUI

	// Main application
	app *tvlib.Application

	// Configuration for auto-loading
	loadArgs *ui.LoadArgs

	// Active widgets
	commandsTerminal *widgets.CommandsTerminal
	sourceFile       *widgets.SourceFile
	sourceSnippet    *widgets.SourceSnippet
	stack            *widgets.Stack
	registers        *widgets.Registers
	memory           *widgets.Memory
	events           *widgets.Events
	filePicker       *widgets.FilePicker
	themeSelector    *widgets.ThemeSelector

	// Layout components
	mainFlex  *tvlib.Flex
	rightFlex *tvlib.Flex
	pages     *tvlib.Pages

	// Theme management
	themeManager *themes.Manager

	// Current mode
	mode     UIMode
	prevMode UIMode

	// For file picker when loading
	filePickerAction string
}

// UIMode represents the current UI mode
type UIMode int

const (
	ModeNormal UIMode = iota
	ModeFilePicker
)

// NewDebuggerTUI creates a new debugger TUI using tview
func NewDebuggerTUI(debugger ui.Debugger, opts ...NewDebuggerTUIOpt) *tvlib.Application {
	// Create theme manager
	themeManager := themes.NewManager()

	app := tvlib.NewApplication()
	asyncDebugger := ui.NewAsyncDebuggerUI(debugger)

	tui := &DebuggerTUI{
		debugger:         asyncDebugger,
		app:              app,
		commandsTerminal: widgets.NewCommandsTerminal(asyncDebugger, app),
		sourceFile:       widgets.NewSourceFile(nil),
		sourceSnippet:    widgets.NewSourceSnippet(nil, 15),
		stack:            widgets.NewStack(nil),
		registers:        widgets.NewRegisters(nil),
		memory:           widgets.NewMemory(nil),
		events:           widgets.NewEvents(),
		filePicker:       widgets.NewFilePicker("."),
		themeManager:     themeManager,
		mode:             ModeNormal,
	}

	// Apply custom options
	config := &debuggerTUIConfig{
		catchPanics: true,
	}
	for _, opt := range opts {
		opt(config)
	}

	// Set configuration from options
	tui.loadArgs = config.loadArgs

	// Create theme selector
	tui.themeSelector = widgets.NewThemeSelector(themeManager)

	// Register callbacks for widgets to receive command results
	tui.registerCallbacks()

	// Build the UI
	tui.buildUI()

	// Apply current theme
	tui.applyCurrentTheme()

	// Set up input capture
	tui.setupInputHandling()

	// Add initial output
	tui.commandsTerminal.AddOutput("Welcome to Cucaracha Debugger TUI")
	tui.commandsTerminal.AddOutput("Type 'help' for available commands or '?' to toggle help")
	tui.commandsTerminal.AddOutput("Type 'theme' to change the color theme")

	// Auto-load system and program if provided
	tui.initializeFromOptions()

	return tui.app
}

// debuggerTUIConfig holds configuration options for the TUI
type debuggerTUIConfig struct {
	catchPanics bool
	loadArgs    *ui.LoadArgs
}

// NewDebuggerTUIOpt is a functional option for configuring the debugger TUI
type NewDebuggerTUIOpt func(*debuggerTUIConfig)

// WithoutCatchPanics returns an option that disables panic catching
func WithoutCatchPanics() NewDebuggerTUIOpt {
	return func(c *debuggerTUIConfig) {
		c.catchPanics = false
	}
}

func WithLoadArgs(args *ui.LoadArgs) NewDebuggerTUIOpt {
	return func(c *debuggerTUIConfig) {
		c.loadArgs = args
	}
}

// registerCallbacks registers callbacks for command results
func (m *DebuggerTUI) registerCallbacks() {
	m.commandsTerminal.RegisterCallback(ui.DebuggerCommandSource, func(result *ui.DebuggerCommandResult) {
		if result.SourceResult != nil && result.SourceResult.Error == nil {
			m.sourceFile.SetResult(result.SourceResult)
		}
	})

	m.commandsTerminal.RegisterCallback(ui.DebuggerCommandCurrentSource, func(result *ui.DebuggerCommandResult) {
		if result.CurrentSourceResult != nil && result.CurrentSourceResult.Error == nil {
			m.sourceSnippet.SetResult(result.CurrentSourceResult)
		}
	})

	m.commandsTerminal.RegisterCallback(ui.DebuggerCommandStack, func(result *ui.DebuggerCommandResult) {
		if result.StackResult != nil && result.StackResult.Error == nil {
			m.stack.SetResult(result.StackResult)
		}
	})

	m.commandsTerminal.RegisterCallback(ui.DebuggerCommandRegisters, func(result *ui.DebuggerCommandResult) {
		if result.RegistersResult != nil && result.RegistersResult.Error == nil {
			m.registers.SetResult(result.RegistersResult)
		}
	})

	m.commandsTerminal.RegisterCallback(ui.DebuggerCommandMemory, func(result *ui.DebuggerCommandResult) {
		if result.MemoryResult != nil && result.MemoryResult.Error == nil {
			m.memory.SetResult(result.MemoryResult)
		}
	})

	m.commandsTerminal.SetFilePickerCallback(func(action string) {
		m.filePickerAction = action
		m.mode = ModeFilePicker
		m.prevMode = ModeNormal
		m.pages.SwitchToPage("filePicker")
	})

	m.commandsTerminal.SetThemeCallback(func(action string) {
		m.mode = ModeFilePicker // Reuse file picker mode for theme selector
		m.prevMode = ModeNormal
		m.pages.SwitchToPage("themeSelector")
	})

	lastEvent := time.Now()
	eventsBuffer := make([]*ui.DebuggerEvent, 0)

	// Set debugger event callback
	m.debugger.SetEventCallback(func(event *ui.DebuggerEvent) {
		// Throttle event updates to avoid overwhelming the UI
		if time.Since(lastEvent) < 100*time.Millisecond {
			eventsBuffer = append(eventsBuffer, event)
			return
		}

		lastEvent = time.Now()
		m.app.QueueUpdateDraw(func() {
			for _, event := range eventsBuffer {
				m.events.AddEvent(event)
			}
			eventsBuffer = make([]*ui.DebuggerEvent, 0)
		})
	})
}

// colorToHex converts a tcell.Color to a hex color string
func colorToHex(c tcell.Color) string {
	r, g, b := c.RGB()
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

// formatError formats an error message using the theme's error color
func (m *DebuggerTUI) formatError(msg string) string {
	color := colorToHex(m.themeManager.GetTheme().UserIO.Error)
	return fmt.Sprintf("[%s]%s[:-:-]", color, msg)
}

// formatSuccess formats a success message using the theme's success color
func (m *DebuggerTUI) formatSuccess(msg string) string {
	color := colorToHex(m.themeManager.GetTheme().UserIO.Success)
	return fmt.Sprintf("[%s]%s[:-:-]", color, msg)
}

// formatWarning formats a warning message using the theme's warning color
func (m *DebuggerTUI) formatWarning(msg string) string {
	color := colorToHex(m.themeManager.GetTheme().UserIO.Warning)
	return fmt.Sprintf("[%s]%s[:-:-]", color, msg)
}

// formatInfo formats an info message using the theme's info color
func (m *DebuggerTUI) formatInfo(msg string) string {
	color := colorToHex(m.themeManager.GetTheme().UserIO.Info)
	return fmt.Sprintf("[%s]%s[:-:-]", color, msg)
}

// buildUI builds the main UI layout
func (m *DebuggerTUI) buildUI() {
	m.rightFlex = tvlib.NewFlex().
		SetDirection(tvlib.FlexRow).
		AddItem(m.sourceSnippet, 0, 1, false).
		AddItem(m.registers, 0, 1, false).
		AddItem(m.stack, 0, 1, false).
		AddItem(m.events, 0, 1, false)

	m.mainFlex = tvlib.NewFlex().
		SetDirection(tvlib.FlexColumn).
		AddItem(m.commandsTerminal, 0, 1, true).
		AddItem(m.rightFlex, 0, 1, false)

	m.pages = tvlib.NewPages().
		AddPage("normal", m.mainFlex, true, true).
		AddPage("filePicker", m.createFilePickerPage(), true, false).
		AddPage("themeSelector", m.createThemeSelectorPage(), true, false)

	m.app.SetRoot(m.pages, true)
}

// createFilePickerPage creates the file picker page
func (m *DebuggerTUI) createFilePickerPage() tvlib.Primitive {
	flex := tvlib.NewFlex().
		SetDirection(tvlib.FlexRow)

	title := tvlib.NewTextView().
		SetDynamicColors(true).
		SetText(fmt.Sprintf("[blue::b]Select file to load (%s)[::-]", m.filePickerAction))
	flex.AddItem(title, 1, 0, false)

	flex.AddItem(m.filePicker, 0, 1, true)

	footer := tvlib.NewTextView().
		SetDynamicColors(true).
		SetText("[gray]Press ESC to cancel, Enter to select[::-]")
	flex.AddItem(footer, 1, 0, false)

	return flex
}

// createThemeSelectorPage creates the theme selector page
func (m *DebuggerTUI) createThemeSelectorPage() tvlib.Primitive {
	flex := tvlib.NewFlex().
		SetDirection(tvlib.FlexRow)

	title := tvlib.NewTextView().
		SetDynamicColors(true).
		SetText("[#AE81FF::b]Select a Color Theme[::-]")
	flex.AddItem(title, 1, 0, false)

	m.themeSelector.SetOnThemeSelected(func(theme *themes.Theme) {
		m.switchTheme(theme)
		// After theme is applied, return to normal mode
		m.mode = ModeNormal
		m.pages.SwitchToPage("normal")
	})

	m.themeSelector.SetOnThemePreview(func(theme *themes.Theme) {
		// Apply theme preview without fully switching
		m.updateWidgetsTheme(theme)
	})

	// Add input capture to handle ESC
	themeList := m.themeSelector.GetList()
	themeList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			m.mode = ModeNormal
			m.pages.SwitchToPage("normal")
			return nil
		}
		// Let the theme selector handle other events (including Enter)
		return event
	})

	flex.AddItem(m.themeSelector, 0, 1, true)

	footer := tvlib.NewTextView().
		SetDynamicColors(true).
		SetText("[gray]Arrow Keys to Browse | Enter to Select | ESC to Close[::-]")
	flex.AddItem(footer, 1, 0, false)

	return flex
}

// applyMonokaiTheme applies the dark monokai theme to the application
func (m *DebuggerTUI) applyMonokaiTheme() {
	// Monokai colors are applied through text markup and terminal defaults
	// No additional theme setup needed - tcell handles terminal colors automatically
}

// applyCurrentTheme applies the current theme from the theme manager to all widgets
func (m *DebuggerTUI) applyCurrentTheme() {
	theme := m.themeManager.GetTheme()
	if theme == nil {
		return
	}
	m.updateWidgetsTheme(theme)
}

// switchTheme switches to a new theme and updates all widgets
func (m *DebuggerTUI) switchTheme(theme *themes.Theme) {
	err := m.themeManager.SetTheme(theme.Name)
	if err == nil {
		m.applyCurrentTheme()
		m.commandsTerminal.AddOutput(fmt.Sprintf("[#A6E22E]Theme changed to: %s[-]", theme.Name))
	}
}

// updateWidgetsTheme updates all widgets with the given theme
func (m *DebuggerTUI) updateWidgetsTheme(theme *themes.Theme) {
	if theme == nil {
		return
	}

	// Apply theme to all widgets recursively
	m.commandsTerminal.SetTheme(theme)
	m.sourceFile.SetTheme(theme)
	m.sourceSnippet.SetTheme(theme)
	m.stack.SetTheme(theme)
	m.registers.SetTheme(theme)
	m.memory.SetTheme(theme)
	m.events.SetTheme(theme)
	m.filePicker.SetTheme(theme)
	m.themeSelector.SetTheme(theme)
}

// setupInputHandling sets up global input handling for the TUI
func (m *DebuggerTUI) setupInputHandling() {
	m.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Let the application handle standard events
		// Input is primarily handled by the InputField in CommandsTerminal

		// Handle global shortcuts
		switch event.Key() {
		case tcell.KeyEscape:
			// Return to normal mode if in file picker
			if m.mode == ModeFilePicker {
				m.mode = ModeNormal
				m.pages.SwitchToPage("normal")
				return nil
			}
		case tcell.KeyCtrlC:
			// Send interrupt command to the debugger
			interruptCmd := &ui.DebuggerCommand{
				Command: ui.DebuggerCommandInterrupt,
			}
			m.debugger.Execute(interruptCmd, func(result *ui.DebuggerCommandResult, err error) {
				m.app.QueueUpdateDraw(func() {
					if err != nil {
						m.commandsTerminal.AddOutput(m.formatError(fmt.Sprintf("Interrupt error: %v", err)))
					} else if result.InterruptResult != nil && result.InterruptResult.Error != nil {
						m.commandsTerminal.AddOutput(m.formatError(fmt.Sprintf("Interrupt error: %v", result.InterruptResult.Error)))
					} else {
						m.commandsTerminal.AddOutput(m.formatSuccess("Interrupt signal sent to the program"))
					}
				})
			})
			return nil
		}
		return event
	})
}

// handleNormalModeInput handles input in normal mode
// This is handled by the InputField in CommandsTerminal
func (m *DebuggerTUI) handleNormalModeInput() {
	// Input is handled by the InputField widget
}

// handleFilePickerModeInput handles input in file picker mode
// This is handled by the tview file picker widget
func (m *DebuggerTUI) handleFilePickerModeInput() {
	// Input is handled by the List widget in FilePicker
}

// AddOutput is a helper method to add output to the command terminal
func (m *DebuggerTUI) AddOutput(msg string) {
	m.commandsTerminal.AddOutput(msg)
}

// GetCommandTerminal returns the CommandsTerminal widget
func (m *DebuggerTUI) GetCommandTerminal() *widgets.CommandsTerminal {
	return m.commandsTerminal
}

// GetStackWidget returns the Stack widget
func (m *DebuggerTUI) GetStackWidget() *widgets.Stack {
	return m.stack
}

// GetRegistersWidget returns the Registers widget
func (m *DebuggerTUI) GetRegistersWidget() *widgets.Registers {
	return m.registers
}

// GetMemoryWidget returns the Memory widget
func (m *DebuggerTUI) GetMemoryWidget() *widgets.Memory {
	return m.memory
}

// GetSourceFileWidget returns the SourceFile widget
func (m *DebuggerTUI) GetSourceFileWidget() *widgets.SourceFile {
	return m.sourceFile
}

func (m *DebuggerTUI) handleLoadResult(result *ui.DebuggerCommandResult, err error) {
	if err != nil {
		m.commandsTerminal.AddOutput(m.formatError(fmt.Sprintf("unexpected error running load command: %v", err)))
		return
	}

	if result.LoadResult.Error != nil {
		m.commandsTerminal.AddOutput(m.formatError(fmt.Sprintf("Error loading system/program/runtime: %v", result.LoadResult.Error)))
		return
	}

	// Display system information
	sysResult := result.LoadResult.System
	if sysResult == nil || sysResult.System == nil {
		m.commandsTerminal.AddOutput(m.formatError("System information not available"))
		return
	}

	sysInfo := sysResult.System
	m.commandsTerminal.AddOutput(m.formatInfo(fmt.Sprintf("Memory: %dKB (Code: %dB, Data: %dB, Stack: %dB, Heap: %dB)",
		sysInfo.TotalMemory/1024, sysInfo.CodeSize, sysInfo.DataSize, sysInfo.StackSize, sysInfo.HeapSize)))
	m.commandsTerminal.AddOutput(m.formatInfo(fmt.Sprintf("Interrupts: %d vectors, Entry size: %d bytes",
		sysInfo.NumberOfVectors, sysInfo.VectorEntrySize)))
	if sysInfo.NumPeripherals > 0 {
		m.commandsTerminal.AddOutput(m.formatInfo(fmt.Sprintf("Peripherals (%d):", sysInfo.NumPeripherals)))
		for _, p := range sysInfo.Peripherals {
			m.commandsTerminal.AddOutput(m.formatInfo(fmt.Sprintf("  - %s", p.Name)))
			if p.Type != "" {
				m.commandsTerminal.AddOutput(m.formatInfo(fmt.Sprintf("    Type: %s", p.Type)))
			}
			if p.DisplayName != "" {
				m.commandsTerminal.AddOutput(m.formatInfo(fmt.Sprintf("    Display Name: %s", p.DisplayName)))
			}
			if p.Description != "" {
				m.commandsTerminal.AddOutput(m.formatInfo(fmt.Sprintf("    Description: %s", p.Description)))
			}
			m.commandsTerminal.AddOutput(m.formatInfo(fmt.Sprintf("    Base Address: 0x%08X, Size: 0x%X (%d bytes)", p.BaseAddress, p.Size, p.Size)))
			m.commandsTerminal.AddOutput(m.formatInfo(fmt.Sprintf("    Interrupt Vector: %d", p.InterruptVector)))
		}
	}

	// Display program loading result
	progResult := result.LoadResult.Program
	if progResult != nil && progResult.Program != nil {
		m.commandsTerminal.AddOutput(m.formatInfo(fmt.Sprintf("Program entry point: 0x%08X", progResult.Program.EntryPoint)))
		if progResult.Program.SourceFile != nil {
			m.commandsTerminal.AddOutput(m.formatInfo(fmt.Sprintf("Loaded source file: %s", *progResult.Program.SourceFile)))
		}
		if progResult.Program.ObjectFile != nil {
			m.commandsTerminal.AddOutput(m.formatInfo(fmt.Sprintf("Generated object file: %s", *progResult.Program.ObjectFile)))
		}
		for _, warning := range progResult.Program.Warnings {
			m.commandsTerminal.AddOutput(m.formatWarning(fmt.Sprintf("compile warning: %v", warning)))
		}
	}

	// Display runtime loading result
	runtimeResult := result.LoadResult.Runtime
	if runtimeResult != nil && runtimeResult.Runtime != nil {
		m.commandsTerminal.AddOutput(m.formatInfo(fmt.Sprintf("Using runtime: %s", runtimeResult.Runtime.Runtime)))
	}
}

// initializeFromOptions loads system and program if configured via options
func (m *DebuggerTUI) initializeFromOptions() {
	if m.loadArgs == nil {
		return
	}

	m.debugger.Execute(&ui.DebuggerCommand{
		Command:  ui.DebuggerCommandLoad,
		LoadArgs: m.loadArgs,
	}, func(result *ui.DebuggerCommandResult, err error) {
		m.app.QueueUpdateDraw(func() {
			m.handleLoadResult(result, err)
		})
	})
}
