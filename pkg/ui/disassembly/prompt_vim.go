package disassembly

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"golang.org/x/term"
)

// VimKeyHandler provides vim-style key handling
type VimKeyHandler struct {
	session      *Session
	rl           *readline.Instance
	reader       *bufio.Reader
	oldTermState *term.State
}

// NewVimKeyHandler creates a new vim-style key handler
func NewVimKeyHandler(session *Session) *VimKeyHandler {
	historyFile, _ := getHistoryFilePath()
	rl, err := readline.NewEx(&readline.Config{
		Prompt:            "",
		HistoryFile:       historyFile,
		HistorySearchFold: true,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
	})
	if err != nil {
		rl = nil
	}

	handler := &VimKeyHandler{
		session: session,
		rl:      rl,
		reader:  bufio.NewReader(os.Stdin),
	}

	// Set terminal to raw mode for single-character input
	if state, err := term.MakeRaw(int(os.Stdin.Fd())); err == nil {
		handler.oldTermState = state
	}

	return handler
}

// GetNextAction reads the next key/input and returns the command to execute
// Returns: (command, mode) where mode is "normal" or "error"
// In normal mode, single keys like 'j', 'k', 'q' are returned as-is
// When ':' or '/' is pressed, readline takes over for full input
func (vkh *VimKeyHandler) GetNextAction() (string, error) {
	// Read a single byte directly in raw mode
	buf := make([]byte, 1)
	n, err := os.Stdin.Read(buf)
	if err != nil || n < 1 {
		return "", err
	}

	ch := rune(buf[0])

	// Handle direct navigation commands
	switch ch {
	case 'j', 'k', 'g', 'G', 'q', '?':
		// Single character commands - return immediately
		return string(ch), nil
	case 'J', 'K': // Also handle capital versions
		return string(ch), nil
	case ':', '/':
		// Enter command/search mode - use readline for full input
		return vkh.readCommandMode(ch)
	case '@':
		// Address jump - read address
		return vkh.readUntilSpace("@"), nil
	case '#':
		// Symbol jump - read symbol
		return vkh.readUntilSpace("#"), nil
	case 3: // Ctrl+C
		return "q", nil
	default:
		// Unknown command or arrow key prefix
		// For arrow keys, we need to handle escape sequences
		if ch == 27 { // ESC character
			return vkh.handleEscapeSequence()
		}
		return "", nil // Ignore unknown keys
	}
}

// readCommandMode reads a full command line when : or / is pressed
func (vkh *VimKeyHandler) readCommandMode(prefix rune) (string, error) {
	if vkh.rl != nil {
		// Temporarily restore cooked mode for readline
		if vkh.oldTermState != nil {
			term.Restore(int(os.Stdin.Fd()), vkh.oldTermState)
			defer func() {
				// Re-enter raw mode after readline
				if state, err := term.MakeRaw(int(os.Stdin.Fd())); err == nil {
					vkh.oldTermState = state
				}
			}()
		}

		// Set prompt for readline
		if prefix == ':' {
			vkh.rl.SetPrompt(":")
		} else {
			vkh.rl.SetPrompt("/")
		}
		line, err := vkh.rl.Readline()
		if err != nil {
			vkh.rl.SetPrompt("")
			return "", nil
		}
		vkh.rl.SetPrompt("")
		return string(prefix) + strings.TrimSpace(line), nil
	}

	// Fallback: read rest of line manually in raw mode
	// Print the prefix and read until newline
	fmt.Fprint(os.Stdout, string(prefix))
	result := string(prefix)
	buf := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n < 1 || buf[0] == '\n' || buf[0] == '\r' {
			break
		}
		result += string(buf[0])
		fmt.Fprint(os.Stdout, string(buf[0]))
	}
	fmt.Fprintln(os.Stdout) // Newline after command
	return result, nil
}

// readUntilSpace reads characters until space or newline
func (vkh *VimKeyHandler) readUntilSpace(prefix string) string {
	result := prefix
	buf := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n < 1 {
			break
		}
		if buf[0] == '\n' || buf[0] == ' ' || buf[0] == '\r' {
			break
		}
		result += string(buf[0])
	}
	return result
}

// handleEscapeSequence handles ANSI escape sequences (arrow keys, etc.)
func (vkh *VimKeyHandler) handleEscapeSequence() (string, error) {
	// Read [ for CSI sequence
	buf := make([]byte, 1)
	n, err := os.Stdin.Read(buf)
	if err != nil || n < 1 {
		return "", nil
	}

	if buf[0] != '[' {
		return "", nil
	}

	// Read the final character
	n, err = os.Stdin.Read(buf)
	if err != nil || n < 1 {
		return "", nil
	}

	// Map arrow keys to vim equivalents
	switch buf[0] {
	case 'A': // Up arrow
		return "k", nil
	case 'B': // Down arrow
		return "j", nil
	case 'C': // Right arrow
		return "l", nil
	case 'D': // Left arrow
		return "h", nil
	default:
		return "", nil
	}
}

// Close closes the readline instance and restores terminal state
func (vkh *VimKeyHandler) Close() error {
	// Restore terminal state
	if vkh.oldTermState != nil {
		term.Restore(int(os.Stdin.Fd()), vkh.oldTermState)
	}

	if vkh.rl != nil {
		return vkh.rl.Close()
	}
	return nil
}
