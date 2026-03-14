package disassembly

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/ui/debugger"
	"github.com/chzyer/readline"
)

// CommandPrompt handles vim-like command input
type CommandPrompt struct {
	session *Session
	rl      *readline.Instance
}

// NewCommandPrompt creates a new command prompt
func NewCommandPrompt(session *Session) *CommandPrompt {
	historyFile, _ := getHistoryFilePath()
	rl, err := readline.NewEx(&readline.Config{
		Prompt:            "",
		HistoryFile:       historyFile,
		HistorySearchFold: true,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
	})
	if err != nil {
		// Fall back to simple input if readline fails
		return &CommandPrompt{session: session, rl: nil}
	}

	return &CommandPrompt{
		session: session,
		rl:      rl,
	}
}

// GetCommand reads a command from the user
func (cp *CommandPrompt) GetCommand() (string, error) {
	if cp.rl != nil {
		line, err := cp.rl.Readline()
		if err != nil {
			return "", nil
		}
		return strings.TrimSpace(line), nil
	}

	// Fallback to simple input
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", nil
	}

	return strings.TrimSpace(line), nil
}

// Close closes the readline instance
func (cp *CommandPrompt) Close() error {
	if cp.rl != nil {
		return cp.rl.Close()
	}
	return nil
}

// getHistoryFilePath returns the path to the disassembler history file
func getHistoryFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	cucarachaDir := filepath.Join(homeDir, ".cucaracha")
	historyFile := filepath.Join(cucarachaDir, "disasm_history")

	// Create the .cucaracha directory if it doesn't exist
	if err := os.MkdirAll(cucarachaDir, 0755); err != nil {
		return "", err
	}

	return historyFile, nil
}

// CommandExecutor handles command execution
type CommandExecutor struct {
	session *Session
}

// NewCommandExecutor creates a new command executor
func NewCommandExecutor(session *Session) *CommandExecutor {
	return &CommandExecutor{session: session}
}

// Execute executes a command
func (ce *CommandExecutor) Execute(cmd string) error {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return nil
	}

	// Handle vim-like single char commands
	if len(cmd) == 1 {
		return ce.executeSingleCharCommand(cmd)
	}

	// Handle prefixed commands
	if len(cmd) > 1 {
		switch cmd[0] {
		case '/':
			// Search command
			pattern := cmd[1:]
			return ce.search(pattern)
		case '@':
			// Jump to address
			addrStr := strings.TrimPrefix(cmd, "@")
			return ce.jumpToAddress(addrStr)
		case '#':
			// Jump to symbol
			symbol := strings.TrimPrefix(cmd, "#")
			return ce.jumpToSymbol(symbol)
		case ':':
			// Colon commands
			return ce.executeColonCommand(strings.TrimPrefix(cmd, ":"))
		}
	}

	return fmt.Errorf("unknown command: %s", cmd)
}

// executeSingleCharCommand executes vim-like single character commands
func (ce *CommandExecutor) executeSingleCharCommand(cmd string) error {
	switch cmd {
	case "q":
		ce.session.SetExit(true)
		return nil
	case "?":
		// Help - handled by tview UI
		return nil
	case "j":
		ce.session.ScrollDown(3)
		return nil
	case "k":
		ce.session.ScrollUp(3)
		return nil
	case "g":
		ce.session.GetViewport().StartIndex = 0
		return nil
	case "G":
		instructions := ce.session.GetInstructions()
		viewport := ce.session.GetViewport()
		viewport.StartIndex = len(instructions) - viewport.Height
		if viewport.StartIndex < 0 {
			viewport.StartIndex = 0
		}
		return nil
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

// executeColonCommand executes colon commands like :deps, :jumps, :info
func (ce *CommandExecutor) executeColonCommand(cmd string) error {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}

	switch parts[0] {
	case "deps":
		// Show dependencies for current instruction or specified address
		if len(parts) > 1 {
			return ce.showDependencies(parts[1])
		}
		return ce.showDependencies(fmt.Sprintf("0x%x", ce.session.currentAddress))

	case "jumps":
		// Show jump graph - handled by tview UI
		return nil

	case "info":
		// Show program information - handled by tview UI
		return nil

	case "help":
		// Help - handled by tview UI
		return nil

	default:
		return fmt.Errorf("unknown colon command: %s", parts[0])
	}
}

// search performs a search and updates the viewport
func (ce *CommandExecutor) search(pattern string) error {
	searchEngine := ce.session.GetSearchEngine()
	results := searchEngine.Search(pattern, SearchByMnemonic)

	if len(results) == 0 {
		// No results found - tview UI will handle messaging
		return nil
	}

	// Jump to first result
	if len(results) > 0 {
		return ce.jumpToAddress(fmt.Sprintf("0x%x", results[0]))
	}

	return nil
}

// jumpToAddress navigates to a specific instruction address
func (ce *CommandExecutor) jumpToAddress(addrStr string) error {
	addr, err := ParseAddress(addrStr)
	if err != nil {
		return fmt.Errorf("invalid address: %s", addrStr)
	}

	return ce.session.JumpToAddress(addr)
}

// jumpToSymbol navigates to a symbol
func (ce *CommandExecutor) jumpToSymbol(symbol string) error {
	dbg := ce.session.GetDebugger()
	symResult := dbg.Symbols(&debugger.SymbolsArgs{SymbolName: &symbol})

	if symResult.Error != nil {
		return fmt.Errorf("failed to look up symbol: %v", symResult.Error)
	}

	if len(symResult.Functions) > 0 && symResult.Functions[0].Address != nil {
		return ce.jumpToAddress(fmt.Sprintf("0x%x", *symResult.Functions[0].Address))
	}

	if len(symResult.Globals) > 0 && symResult.Globals[0].Address != nil {
		return ce.jumpToAddress(fmt.Sprintf("0x%x", *symResult.Globals[0].Address))
	}

	return fmt.Errorf("symbol '%s' not found", symbol)
}

// showDependencies displays instruction dependencies
func (ce *CommandExecutor) showDependencies(addrStr string) error {
	addr, err := ParseAddress(addrStr)
	if err != nil {
		return fmt.Errorf("invalid address: %s", addrStr)
	}

	// Display dependencies - tview UI will handle this
	_ = addr
	return nil
}
