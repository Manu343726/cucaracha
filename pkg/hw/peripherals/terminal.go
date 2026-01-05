package peripherals

import (
	"sync"

	"github.com/Manu343726/cucaracha/pkg/hw/peripheral"
)

// A simple text I/O peripheral.
// It provides a character-based interface for programs to output text
// and read user input through memory-mapped registers.
//
// Memory Map (8 bytes):
//   - Offset 0: TX_DATA (W) - Write a byte to output buffer
//   - Offset 1: RX_DATA (R) - Read a byte from input buffer (consumes it)
//   - Offset 2: STATUS  (R) - Status register
//   - Bit 0: TX_READY - 1 if output buffer can accept data
//   - Bit 1: RX_READY - 1 if input buffer has data
//   - Offset 3: CONTROL (W) - Control register
//   - Bit 0: TX_INT_EN - Enable interrupt on TX ready
//   - Bit 1: RX_INT_EN - Enable interrupt on RX ready
//   - Offset 4-7: Reserved
type Terminal struct {
	name        string
	baseAddress uint32
	mu          sync.Mutex

	// Output buffer - bytes written by the CPU
	txBuffer []byte
	// Input buffer - bytes to be read by the CPU
	rxBuffer []byte

	// Control register
	txIntEnable bool
	rxIntEnable bool

	// Interrupt state
	interruptPending bool
	interruptVector  uint8

	// Callbacks for UI integration
	onOutput func(b byte)   // Called when CPU writes a byte
	onFlush  func(s string) // Called when output should be displayed
}

// Register offsets
const (
	TerminalTXData  = 0
	TerminalRXData  = 1
	TerminalStatus  = 2
	TerminalControl = 3
)

// Status bits
const (
	TerminalTXReady = 1 << 0
	TerminalRXReady = 1 << 1
)

// Control bits
const (
	TerminalTXIntEn = 1 << 0
	TerminalRXIntEn = 1 << 1
)

// Creates a new terminal peripheral at the given base address.
func NewTerminal(name string, baseAddress uint32) *Terminal {
	return &Terminal{
		name:            name,
		baseAddress:     baseAddress,
		txBuffer:        make([]byte, 0, 256),
		rxBuffer:        make([]byte, 0, 256),
		interruptVector: 1, // Default vector
	}
}

func (t *Terminal) Metadata() peripheral.Metadata {
	return peripheral.Metadata{
		Name:        t.name,
		Description: "Character-based terminal I/O device",
		BaseAddress: t.baseAddress,
		Size:        8,
		Descriptor:  Peripheral(PeripheralType_Terminal),
	}
}

// ReadByte reads from a memory-mapped register.
func (t *Terminal) ReadByte(offset uint32) byte {
	t.mu.Lock()
	defer t.mu.Unlock()

	switch offset {
	case TerminalTXData:
		// TX is write-only, reading returns 0
		return 0

	case TerminalRXData:
		// Read and consume one byte from input buffer
		if len(t.rxBuffer) > 0 {
			b := t.rxBuffer[0]
			t.rxBuffer = t.rxBuffer[1:]
			return b
		}
		return 0

	case TerminalStatus:
		var status byte
		status |= TerminalTXReady // Always ready to accept output
		if len(t.rxBuffer) > 0 {
			status |= TerminalRXReady
		}
		return status

	case TerminalControl:
		var ctrl byte
		if t.txIntEnable {
			ctrl |= TerminalTXIntEn
		}
		if t.rxIntEnable {
			ctrl |= TerminalRXIntEn
		}
		return ctrl

	default:
		return 0
	}
}

// WriteByte writes to a memory-mapped register.
func (t *Terminal) WriteByte(offset uint32, value byte) {
	t.mu.Lock()
	defer t.mu.Unlock()

	switch offset {
	case TerminalTXData:
		// Write byte to output buffer
		t.txBuffer = append(t.txBuffer, value)
		if t.onOutput != nil {
			t.onOutput(value)
		}
		// Flush on newline
		if value == '\n' && t.onFlush != nil {
			t.onFlush(string(t.txBuffer))
			t.txBuffer = t.txBuffer[:0]
		}

	case TerminalControl:
		t.txIntEnable = value&TerminalTXIntEn != 0
		t.rxIntEnable = value&TerminalRXIntEn != 0
	}
}

// Reset clears all buffers and state.
func (t *Terminal) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.txBuffer = t.txBuffer[:0]
	t.rxBuffer = t.rxBuffer[:0]
	t.txIntEnable = false
	t.rxIntEnable = false
	t.interruptPending = false
}

// Called each CPU cycle.
func (t *Terminal) Clock(env peripheral.Environment) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Write registers and buffers to RAM
	// TODO

	// Check for interrupt conditions
	if t.rxIntEnable && len(t.rxBuffer) > 0 {
		t.interruptPending = true
	}

	return nil
}

// --- UIPeripheral interface ---

// Returns the current state for UI display.
func (t *Terminal) UIState() map[string]interface{} {
	t.mu.Lock()
	defer t.mu.Unlock()

	return map[string]interface{}{
		"output":       string(t.txBuffer),
		"inputPending": len(t.rxBuffer),
		"txIntEnable":  t.txIntEnable,
		"rxIntEnable":  t.rxIntEnable,
	}
}

// Returns available UI actions.
func (t *Terminal) UIActions() map[string]string {
	return map[string]string{
		"sendChar":   "Send a character to the terminal input",
		"sendString": "Send a string to the terminal input",
		"clear":      "Clear the output buffer",
	}
}

// Executes a UI action.
func (t *Terminal) UITrigger(action string, params map[string]interface{}) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	switch action {
	case "sendChar":
		if ch, ok := params["char"].(byte); ok {
			t.rxBuffer = append(t.rxBuffer, ch)
		} else if s, ok := params["char"].(string); ok && len(s) > 0 {
			t.rxBuffer = append(t.rxBuffer, s[0])
		}

	case "sendString":
		if s, ok := params["string"].(string); ok {
			t.rxBuffer = append(t.rxBuffer, []byte(s)...)
		}

	case "clear":
		t.txBuffer = t.txBuffer[:0]
	}

	return nil
}

// --- InterruptSource interface ---

// Returns true if an interrupt is pending.
func (t *Terminal) InterruptPending() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.interruptPending
}

// Returns the interrupt vector number.
func (t *Terminal) InterruptVector() uint8 {
	return t.interruptVector
}

// AcknowledgeInterrupt clears the pending interrupt.
func (t *Terminal) AcknowledgeInterrupt() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.interruptPending = false
}

// --- Additional Terminal-specific methods ---

// Sets the callback for output bytes.
func (t *Terminal) SetOutputCallback(cb func(byte)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onOutput = cb
}

// Sets the callback for line output.
func (t *Terminal) SetFlushCallback(cb func(string)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onFlush = cb
}

// SendInput queues bytes for the CPU to read.
func (t *Terminal) SendInput(data []byte) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.rxBuffer = append(t.rxBuffer, data...)
}

// Returns and clears the output buffer.
func (t *Terminal) ReadOutput() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	s := string(t.txBuffer)
	t.txBuffer = t.txBuffer[:0]
	return s
}

// Returns true if there's pending input data.
func (t *Terminal) HasInput() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.rxBuffer) > 0
}

// Sets the interrupt vector number.
func (t *Terminal) SetInterruptVector(vector uint8) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.interruptVector = vector
}

const PeripheralType_Terminal peripheral.Type = "terminal"

func TerminalDescriptor() *peripheral.Descriptor {
	return &peripheral.Descriptor{
		Type:        PeripheralType_Terminal,
		Description: "Character-based terminal I/O device",
		Factory: func(params peripheral.PeripheralParams) (peripheral.Peripheral, error) {
			return NewTerminal(params.Name, params.BaseAddress), nil
		},
		DefaultSize:  8,
		HasInterrupt: true,
	}
}
