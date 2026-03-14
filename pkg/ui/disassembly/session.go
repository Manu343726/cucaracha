// Package disassembly provides an interactive disassembly viewer for Cucaracha programs.
// It offers features like instruction dependency graphs, jump graphs, and advanced search capabilities.
package disassembly

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/Manu343726/cucaracha/pkg/ui/debugger"
	"github.com/Manu343726/cucaracha/pkg/utils"
	"github.com/Manu343726/cucaracha/pkg/utils/logging"
	"golang.org/x/term"
)

// Session manages an interactive disassembly session
type Session struct {
	debugger        debugger.Debugger
	loadArgs        *debugger.LoadArgs
	disasmResult    *debugger.DisasmResult           // Currently displayed disassembly
	instructions    map[uint32]*debugger.Instruction // Quick lookup by address
	showSource      bool
	showCFG         bool
	showDeps        bool
	viewport        *Viewport
	searchEngine    *SearchEngine
	dependencyGraph *InstructionDependencyGraph
	jumpGraph       *JumpGraph
	currentAddress  uint32
	currentIndex    int // Index in the instructions array
	exit            bool

	logger *logging.Logger
}

// Viewport manages the scrollable view of disassembly
type Viewport struct {
	StartIndex int // Index of the first visible instruction
	Height     int // Number of lines visible
	Width      int // Console width
}

// NewSession creates a new interactive disassembly session
func NewSession(dbg debugger.Debugger, loadArgs *debugger.LoadArgs) *Session {
	// Detect terminal size
	width, height := getTerminalSize()

	return &Session{
		debugger:       dbg,
		loadArgs:       loadArgs,
		instructions:   make(map[uint32]*debugger.Instruction),
		showSource:     false,
		showCFG:        true,
		showDeps:       true,
		searchEngine:   NewSearchEngine(),
		viewport:       &Viewport{StartIndex: 0, Height: height - 5, Width: width - 2},
		currentAddress: 0,
		currentIndex:   0,
		exit:           false,
		logger:         log(),
	}
}

// SetShowSource enables/disables source code display
func (s *Session) SetShowSource(show bool) {
	s.showSource = show
}

// SetShowCFG enables/disables control flow graph display
func (s *Session) SetShowCFG(show bool) {
	s.showCFG = show
}

// SetShowDeps enables/disables instruction dependency display
func (s *Session) SetShowDeps(show bool) {
	s.showDeps = show
}

// Run starts the interactive disassembly session
func (s *Session) Run() error {
	// Load the program (logs go to stderr)
	if err := s.loadProgram(); err != nil {
		return fmt.Errorf("failed to load program: %w", err)
	}

	// Silence all logging to prevent TUI interference
	// Re-register root logger with a silent (discard) sink
	reg := logging.DefaultRegistry()
	discardSink := logging.NewTextSink("discard", io.Discard, slog.LevelDebug)
	if err := reg.RegisterSink(discardSink); err == nil {
		reg.RegisterLogger(logging.NewRegisteredLogger("cucaracha", discardSink))
	}

	// Use tview-based UI
	app, err := NewDisasmApp(s)
	if err != nil {
		return fmt.Errorf("failed to create UI: %w", err)
	}

	return app.Run()
}

// loadProgram loads and disassembles the program
func (s *Session) loadProgram() error {
	// Ensure loadArgs are properly set
	if s.loadArgs == nil {
		s.loadArgs = &debugger.LoadArgs{}
	}

	// Set defaults
	if s.loadArgs.SystemConfigPath == nil {
		s.loadArgs.SystemConfigPath = utils.Ptr("default")
	}

	if s.loadArgs.Runtime == nil {
		s.loadArgs.Runtime = utils.Ptr(debugger.RuntimeTypeInterpreter)
	}

	// Log loading progress
	s.logger.Info("Loading program", "runtime", (*s.loadArgs.Runtime).String())

	// Load system, runtime, and program using the unified Load command
	loadResult := s.debugger.Load(s.loadArgs)
	if loadResult.Error != nil {
		s.logger.Error("Failed to load program", "error", fmt.Sprintf("%v", loadResult.Error))
		return fmt.Errorf("failed to load: %v", loadResult.Error)
	}

	s.logger.Info("Program loaded successfully")

	// Initial disassembly from the beginning
	return s.disassembleFromStart()
}

// disassembleFromStart performs initial disassembly
func (s *Session) disassembleFromStart() error {
	s.logger.Debug("Finding entry point for disassembly")

	// Get program info to find entry point
	progResult := s.debugger.Info(&debugger.InfoArgs{Type: debugger.InfoTypeProgram})

	if progResult != nil && progResult.ProgramInfo != nil {
		entryPoint := progResult.ProgramInfo.EntryPoint
		s.logger.Debug("Found program entry point", "address", fmt.Sprintf("0x%x", entryPoint))
		// Try disassembling from the entry point first
		if err := s.tryDisassembleFromAddress(entryPoint); err == nil {
			s.logger.Info("Disassembly started from entry point")
			return nil
		}
		s.logger.Debug("Entry point disassembly failed, trying alternatives")
	}

	// Try to get all symbols to find alternative starting location
	s.logger.Debug("Attempting to load program symbols")
	symbolsResult := s.debugger.Symbols(&debugger.SymbolsArgs{SymbolName: nil})
	if symbolsResult.Error == nil && len(symbolsResult.Functions) > 0 {
		s.logger.Debug("Symbol table loaded", "function_count", len(symbolsResult.Functions))
		// Try to find main
		for _, fn := range symbolsResult.Functions {
			if fn.Name == "main" && fn.Address != nil {
				s.logger.Debug("Found main function", "address", fmt.Sprintf("0x%x", *fn.Address))
				if err := s.tryDisassembleFromAddress(*fn.Address); err == nil {
					s.logger.Info("Disassembly started from main")
					return nil
				}
			}
		}

		// Use the first function if main not found
		if symbolsResult.Functions[0].Address != nil {
			fnName := symbolsResult.Functions[0].Name
			fnAddr := *symbolsResult.Functions[0].Address
			s.logger.Debug("Using first function as entry point", "name", fnName, "address", fmt.Sprintf("0x%x", fnAddr))
			if err := s.tryDisassembleFromAddress(fnAddr); err == nil {
				s.logger.Info("Disassembly started from first function")
				return nil
			}
		}
	}

	// Default fallback: try from address 0
	s.logger.Debug("Using fallback address 0 for disassembly")
	return s.tryDisassembleFromAddress(0)
}

// tryDisassembleFromAddress tries to disassemble from a specific address
func (s *Session) tryDisassembleFromAddress(addr uint32) error {
	count := 100
	countStr := fmt.Sprintf("%d", count)
	result := s.debugger.Disasm(&debugger.DisasmArgs{
		Address:    fmt.Sprintf("0x%x", addr),
		CountExpr:  &countStr,
		ShowSource: s.showSource,
		ShowCFG:    s.showCFG,
	})

	if result.Error != nil {
		return fmt.Errorf("failed to disassemble: %v", result.Error)
	}

	s.disasmResult = result
	s.rebuildIndexes()

	// Build analysis graphs
	if s.showCFG {
		s.jumpGraph = NewJumpGraph(result)
	}
	if s.showDeps {
		s.dependencyGraph = NewInstructionDependencyGraph(result)
	}

	return nil
}

// rebuildIndexes rebuilds the instruction map and searches
func (s *Session) rebuildIndexes() {
	s.instructions = make(map[uint32]*debugger.Instruction)
	for _, instr := range s.disasmResult.Instructions {
		s.instructions[instr.Address] = instr
	}
	s.searchEngine.Index(s.disasmResult.Instructions)
}

// ExecuteCommand executes a single user command
func (s *Session) ExecuteCommand(cmd string) error {
	executor := NewCommandExecutor(s)
	return executor.Execute(cmd)
}

// GetInstructions returns the current disassembled instructions
func (s *Session) GetInstructions() []*debugger.Instruction {
	if s.disasmResult == nil {
		return nil
	}
	return s.disasmResult.Instructions
}

// GetInstruction returns a specific instruction by address
func (s *Session) GetInstruction(addr uint32) *debugger.Instruction {
	return s.instructions[addr]
}

// GetDependencies returns dependencies for an instruction
func (s *Session) GetDependencies(addr uint32) []uint32 {
	if s.dependencyGraph == nil {
		return nil
	}
	return s.dependencyGraph.GetDependencies(addr)
}

// GetDependents returns dependents for an instruction
func (s *Session) GetDependents(addr uint32) []uint32 {
	if s.dependencyGraph == nil {
		return nil
	}
	return s.dependencyGraph.GetDependents(addr)
}

// GetJumpGraph returns the jump graph
func (s *Session) GetJumpGraph() *JumpGraph {
	return s.jumpGraph
}

// GetDependencyGraph returns the instruction dependency graph
func (s *Session) GetDependencyGraph() *InstructionDependencyGraph {
	return s.dependencyGraph
}

// GetControlFlowGraph returns the control flow graph
func (s *Session) GetControlFlowGraph() *debugger.ControlFlowGraph {
	if s.disasmResult == nil {
		return nil
	}
	return s.disasmResult.ControlFlowGraph
}

// GetViewport returns the current viewport
func (s *Session) GetViewport() *Viewport {
	return s.viewport
}

// GetSearchEngine returns the search engine
func (s *Session) GetSearchEngine() *SearchEngine {
	return s.searchEngine
}

// GetDebugger returns the underlying debugger
func (s *Session) GetDebugger() debugger.Debugger {
	return s.debugger
}

// GetLoadArgs returns the load arguments
func (s *Session) GetLoadArgs() *debugger.LoadArgs {
	return s.loadArgs
}

// SetExit marks the session for exit
func (s *Session) SetExit(exit bool) {
	s.exit = exit
}

// JumpToAddress navigates to a specific instruction address
func (s *Session) JumpToAddress(addr uint32) error {
	// Find the instruction index
	for i, instr := range s.disasmResult.Instructions {
		if instr.Address == addr {
			s.viewport.StartIndex = i
			s.currentAddress = addr
			s.currentIndex = i
			return nil
		}
	}

	return fmt.Errorf("address 0x%x not found in disassembly", addr)
}

// IsShowSource returns if source display is enabled
func (s *Session) IsShowSource() bool {
	return s.showSource
}

// ScrollUp scrolls up in the viewport
func (s *Session) ScrollUp(lines int) {
	s.viewport.StartIndex -= lines
	if s.viewport.StartIndex < 0 {
		s.viewport.StartIndex = 0
	}
}

// ScrollDown scrolls down in the viewport
func (s *Session) ScrollDown(lines int) {
	maxStart := len(s.disasmResult.Instructions) - s.viewport.Height
	s.viewport.StartIndex += lines
	if s.viewport.StartIndex > maxStart {
		s.viewport.StartIndex = maxStart
	}
	if s.viewport.StartIndex < 0 {
		s.viewport.StartIndex = 0
	}
}

// getTerminalSize returns the current terminal width and height
func getTerminalSize() (width, height int) {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w == 0 || h == 0 {
		// Fallback to reasonable defaults
		return 100, 30
	}
	return w, h
}
