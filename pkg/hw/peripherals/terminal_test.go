package peripherals

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTerminal_ReadWrite(t *testing.T) {
	term := NewTerminal("test", 0x1000)

	// Write some data
	term.WriteByte(TerminalTXData, 'H')
	term.WriteByte(TerminalTXData, 'i')

	// Read output
	output := term.ReadOutput()
	assert.Equal(t, "Hi", output)

	// Check status
	status := term.ReadByte(TerminalStatus)
	assert.Equal(t, byte(TerminalTXReady), status&TerminalTXReady)
}

func TestTerminal_SendInput(t *testing.T) {
	term := NewTerminal("test", 0x1000)

	// Send input
	term.SendInput([]byte("Hello"))

	// Check status shows RX ready
	status := term.ReadByte(TerminalStatus)
	assert.Equal(t, byte(TerminalRXReady), status&TerminalRXReady)

	// Read input
	var received []byte
	for i := 0; i < 5; i++ {
		received = append(received, term.ReadByte(TerminalRXData))
	}
	assert.Equal(t, "Hello", string(received))
}
