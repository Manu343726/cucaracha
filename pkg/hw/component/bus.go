// Package component provides hardware abstraction interfaces for cucaracha components.
package component

import (
	"fmt"
	"sync"
)

// BitValue represents a single bit state
type BitValue bool

const (
	Low  BitValue = false
	High BitValue = true
)

// String returns a string representation of the bit value
func (b BitValue) String() string {
	if b {
		return "1"
	}
	return "0"
}

// Direction indicates whether a bus is input or output
type Direction int

const (
	Input Direction = iota
	Output
	Bidirectional
)

// String returns a string representation of the direction
func (d Direction) String() string {
	switch d {
	case Input:
		return "IN"
	case Output:
		return "OUT"
	case Bidirectional:
		return "INOUT"
	default:
		return "?"
	}
}

// Port represents an N-bit connection point that can be read and written
// either bit-by-bit or as a whole value.
type Port interface {
	// Name returns the bus name
	Name() string

	// Width returns the number of bits in the bus
	Width() int

	// Direction returns whether this is an input, output, or bidirectional bus
	Direction() Direction

	// Bit-level access

	// GetBit returns the value of bit at position (0 = LSB)
	GetBit(position int) (BitValue, error)

	// SetBit sets the value of bit at position (0 = LSB)
	SetBit(position int, value BitValue) error

	// GetBits returns a slice of bit values from start to end (inclusive)
	GetBits(start, end int) ([]BitValue, error)

	// SetBits sets multiple consecutive bits starting at position
	SetBits(start int, values []BitValue) error

	// Word-level access (for buses up to 64 bits)

	// GetValue returns the entire bus value as uint64
	// For buses wider than 64 bits, only the lower 64 bits are returned
	GetValue() uint64

	// SetValue sets the entire bus value from uint64
	// For buses wider than 64 bits, only the lower 64 bits are set
	SetValue(value uint64) error

	// Byte-level access (for arbitrary width buses)

	// GetBytes returns the bus value as a byte slice (little-endian)
	GetBytes() []byte

	// SetBytes sets the bus value from a byte slice (little-endian)
	SetBytes(data []byte) error

	// State

	// IsTristate returns true if the bus is in high-impedance state
	IsTristate() bool

	// SetTristate puts the bus in high-impedance state
	SetTristate(enabled bool)

	// Reset sets all bits to low
	Reset()
}

// =============================================================================
// Standard Port Implementation
// =============================================================================

// StandardPort is a basic implementation of the Port interface
type StandardPort struct {
	name      string
	width     int
	direction Direction
	bits      []BitValue
	tristate  bool
	mu        sync.RWMutex

	// Optional callback when value changes
	onChange func(port *StandardPort, oldValue, newValue uint64)
}

// PortOption is a functional option for configuring a StandardPort
type PortOption func(*StandardPort)

// WithDirection sets the port direction
func WithDirection(dir Direction) PortOption {
	return func(p *StandardPort) {
		p.direction = dir
	}
}

// WithOnChange sets a callback for value changes
func WithOnChange(callback func(port *StandardPort, oldValue, newValue uint64)) PortOption {
	return func(p *StandardPort) {
		p.onChange = callback
	}
}

// WithInitialValue sets the initial value of the port
func WithInitialValue(value uint64) PortOption {
	return func(p *StandardPort) {
		p.setValueInternal(value)
	}
}

// NewPort creates a new port with the specified name and width
func NewPort(name string, width int, opts ...PortOption) *StandardPort {
	if width < 1 {
		width = 1
	}
	if width > 64 {
		// For wider ports, we still support them but word access is limited
	}

	p := &StandardPort{
		name:      name,
		width:     width,
		direction: Bidirectional,
		bits:      make([]BitValue, width),
		tristate:  false,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// NewInputPort creates a new input port
func NewInputPort(name string, width int, opts ...PortOption) *StandardPort {
	opts = append([]PortOption{WithDirection(Input)}, opts...)
	return NewPort(name, width, opts...)
}

// NewOutputPort creates a new output port
func NewOutputPort(name string, width int, opts ...PortOption) *StandardPort {
	opts = append([]PortOption{WithDirection(Output)}, opts...)
	return NewPort(name, width, opts...)
}

// Name returns the port name
func (p *StandardPort) Name() string {
	return p.name
}

// Width returns the number of bits in the port
func (p *StandardPort) Width() int {
	return p.width
}

// Direction returns the port direction
func (p *StandardPort) Direction() Direction {
	return p.direction
}

// GetBit returns the value of bit at position (0 = LSB)
func (p *StandardPort) GetBit(position int) (BitValue, error) {
	if position < 0 || position >= p.width {
		return Low, fmt.Errorf("bit position %d out of range [0, %d)", position, p.width)
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.tristate {
		return Low, nil // Could also return an error or special value
	}

	return p.bits[position], nil
}

// SetBit sets the value of bit at position (0 = LSB)
func (p *StandardPort) SetBit(position int, value BitValue) error {
	if position < 0 || position >= p.width {
		return fmt.Errorf("bit position %d out of range [0, %d)", position, p.width)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.tristate {
		return fmt.Errorf("cannot set bit on tristate port")
	}

	oldValue := p.getValueInternal()
	p.bits[position] = value
	newValue := p.getValueInternal()

	if p.onChange != nil && oldValue != newValue {
		p.onChange(p, oldValue, newValue)
	}

	return nil
}

// GetBits returns a slice of bit values from start to end (inclusive)
func (p *StandardPort) GetBits(start, end int) ([]BitValue, error) {
	if start < 0 || start >= p.width {
		return nil, fmt.Errorf("start position %d out of range [0, %d)", start, p.width)
	}
	if end < start || end >= p.width {
		return nil, fmt.Errorf("end position %d invalid (start=%d, width=%d)", end, start, p.width)
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]BitValue, end-start+1)
	copy(result, p.bits[start:end+1])
	return result, nil
}

// SetBits sets multiple consecutive bits starting at position
func (p *StandardPort) SetBits(start int, values []BitValue) error {
	if start < 0 || start >= p.width {
		return fmt.Errorf("start position %d out of range [0, %d)", start, p.width)
	}
	if start+len(values) > p.width {
		return fmt.Errorf("values exceed port width (start=%d, len=%d, width=%d)", start, len(values), p.width)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.tristate {
		return fmt.Errorf("cannot set bits on tristate port")
	}

	oldValue := p.getValueInternal()
	copy(p.bits[start:], values)
	newValue := p.getValueInternal()

	if p.onChange != nil && oldValue != newValue {
		p.onChange(p, oldValue, newValue)
	}

	return nil
}

// getValueInternal returns the value without locking (caller must hold lock)
func (p *StandardPort) getValueInternal() uint64 {
	var value uint64
	maxBits := p.width
	if maxBits > 64 {
		maxBits = 64
	}

	for i := 0; i < maxBits; i++ {
		if p.bits[i] {
			value |= 1 << i
		}
	}
	return value
}

// setValueInternal sets the value without locking (caller must hold lock)
func (p *StandardPort) setValueInternal(value uint64) {
	maxBits := p.width
	if maxBits > 64 {
		maxBits = 64
	}

	for i := 0; i < maxBits; i++ {
		p.bits[i] = BitValue((value>>i)&1 == 1)
	}
}

// GetValue returns the entire port value as uint64
func (p *StandardPort) GetValue() uint64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.tristate {
		return 0
	}

	return p.getValueInternal()
}

// SetValue sets the entire port value from uint64
func (p *StandardPort) SetValue(value uint64) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.tristate {
		return fmt.Errorf("cannot set value on tristate port")
	}

	oldValue := p.getValueInternal()

	// Mask value to port width
	if p.width < 64 {
		mask := uint64((1 << p.width) - 1)
		value &= mask
	}

	p.setValueInternal(value)

	if p.onChange != nil && oldValue != value {
		p.onChange(p, oldValue, value)
	}

	return nil
}

// GetBytes returns the port value as a byte slice (little-endian)
func (p *StandardPort) GetBytes() []byte {
	p.mu.RLock()
	defer p.mu.RUnlock()

	numBytes := (p.width + 7) / 8
	result := make([]byte, numBytes)

	for i := 0; i < p.width; i++ {
		if p.bits[i] {
			byteIdx := i / 8
			bitIdx := i % 8
			result[byteIdx] |= 1 << bitIdx
		}
	}

	return result
}

// SetBytes sets the port value from a byte slice (little-endian)
func (p *StandardPort) SetBytes(data []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.tristate {
		return fmt.Errorf("cannot set bytes on tristate port")
	}

	oldValue := p.getValueInternal()

	// Clear all bits first
	for i := range p.bits {
		p.bits[i] = Low
	}

	// Set bits from data
	for byteIdx, byteVal := range data {
		for bitIdx := 0; bitIdx < 8; bitIdx++ {
			bitPos := byteIdx*8 + bitIdx
			if bitPos >= p.width {
				break
			}
			p.bits[bitPos] = BitValue((byteVal>>bitIdx)&1 == 1)
		}
	}

	newValue := p.getValueInternal()
	if p.onChange != nil && oldValue != newValue {
		p.onChange(p, oldValue, newValue)
	}

	return nil
}

// IsTristate returns true if the port is in high-impedance state
func (p *StandardPort) IsTristate() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.tristate
}

// SetTristate puts the port in high-impedance state
func (p *StandardPort) SetTristate(enabled bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.tristate = enabled
}

// Reset sets all bits to low
func (p *StandardPort) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()

	oldValue := p.getValueInternal()

	for i := range p.bits {
		p.bits[i] = Low
	}
	p.tristate = false

	if p.onChange != nil && oldValue != 0 {
		p.onChange(p, oldValue, 0)
	}
}

// String returns a string representation of the port
func (p *StandardPort) String() string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.tristate {
		return fmt.Sprintf("%s[%d]<%s>=Z", p.name, p.width, p.direction)
	}

	// Build binary string (MSB first for readability)
	bits := make([]byte, p.width)
	for i := 0; i < p.width; i++ {
		if p.bits[p.width-1-i] {
			bits[i] = '1'
		} else {
			bits[i] = '0'
		}
	}

	return fmt.Sprintf("%s[%d]<%s>=0b%s (0x%X)", p.name, p.width, p.direction, string(bits), p.getValueInternal())
}

// =============================================================================
// Pin Implementation (1-bit Port)
// =============================================================================

// Pin is a 1-bit port with convenient single-bit access methods.
// It embeds StandardPort and adds pin-specific helper methods.
type Pin struct {
	*StandardPort
}

// NewPin creates a new pin (1-bit port) with the specified name
func NewPin(name string, opts ...PortOption) *Pin {
	return &Pin{
		StandardPort: NewPort(name, 1, opts...),
	}
}

// NewInputPin creates a new input pin
func NewInputPin(name string, opts ...PortOption) *Pin {
	opts = append([]PortOption{WithDirection(Input)}, opts...)
	return NewPin(name, opts...)
}

// NewOutputPin creates a new output pin
func NewOutputPin(name string, opts ...PortOption) *Pin {
	opts = append([]PortOption{WithDirection(Output)}, opts...)
	return NewPin(name, opts...)
}

// Get returns the current pin value
func (p *Pin) Get() BitValue {
	val, _ := p.GetBit(0)
	return val
}

// Set sets the pin value
func (p *Pin) Set(value BitValue) error {
	return p.SetBit(0, value)
}

// IsHigh returns true if the pin is high
func (p *Pin) IsHigh() bool {
	return p.Get() == High
}

// IsLow returns true if the pin is low
func (p *Pin) IsLow() bool {
	return p.Get() == Low
}

// Toggle inverts the pin value
func (p *Pin) Toggle() error {
	current := p.Get()
	return p.Set(!current)
}

// String returns a string representation of the pin
func (p *Pin) String() string {
	if p.IsTristate() {
		return fmt.Sprintf("%s<%s>=Z", p.Name(), p.Direction())
	}
	return fmt.Sprintf("%s<%s>=%s", p.Name(), p.Direction(), p.Get())
}
