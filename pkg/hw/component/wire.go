// Package component provides hardware abstraction interfaces for cucaracha components.
package component

import (
	"fmt"
	"sync"
)

// Represents a connection between ports that transfers data
type Connection interface {
	// Source returns the source port
	Source() Port

	// Destination returns the destination port
	Destination() Port

	// Transfer copies data from source to destination
	Transfer() error

	// IsEnabled returns whether the connection is active
	IsEnabled() bool

	// Enable/Disable the connection
	Enable()
	Disable()
}

// =============================================================================
// Bus Implementation
// =============================================================================

// A shared communication pathway connecting two ports
type Bus struct {
	name        string
	source      Port
	destination Port
	enabled     bool
	mu          sync.RWMutex

	// Optional bit mapping (nil means 1:1 mapping)
	// Maps source bit position to destination bit position
	bitMap map[int]int

	// Optional transform function
	transform func(value uint64) uint64
}

// A functional option for configuring a Bus
type BusOption func(*Bus)

// Sets a custom bit mapping for the bus
func WithBitMapping(mapping map[int]int) BusOption {
	return func(b *Bus) {
		b.bitMap = mapping
	}
}

// Sets a transform function for the bus
func WithTransform(fn func(value uint64) uint64) BusOption {
	return func(b *Bus) {
		b.transform = fn
	}
}

// Sets the initial enabled state of the bus
func WithBusEnabled(enabled bool) BusOption {
	return func(b *Bus) {
		b.enabled = enabled
	}
}

// Creates a new bus between two ports
func NewBus(name string, source, destination Port, opts ...BusOption) *Bus {
	b := &Bus{
		name:        name,
		source:      source,
		destination: destination,
		enabled:     true,
	}

	for _, opt := range opts {
		opt(b)
	}

	return b
}

// Returns the bus name
func (b *Bus) Name() string {
	return b.name
}

// Returns the source port
func (b *Bus) Source() Port {
	return b.source
}

// Returns the destination port
func (b *Bus) Destination() Port {
	return b.destination
}

// Transfer copies data from source to destination
func (b *Bus) Transfer() error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.enabled {
		return nil
	}

	if b.source.IsTristate() {
		// Don't transfer from tristate source
		return nil
	}

	// If there's a bit mapping, use it
	if b.bitMap != nil {
		for srcBit, dstBit := range b.bitMap {
			val, err := b.source.GetBit(srcBit)
			if err != nil {
				return fmt.Errorf("failed to get source bit %d: %w", srcBit, err)
			}
			if err := b.destination.SetBit(dstBit, val); err != nil {
				return fmt.Errorf("failed to set destination bit %d: %w", dstBit, err)
			}
		}
		return nil
	}

	// Default: transfer entire value
	value := b.source.GetValue()

	if b.transform != nil {
		value = b.transform(value)
	}

	return b.destination.SetValue(value)
}

// Returns whether the bus is enabled
func (b *Bus) IsEnabled() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.enabled
}

// Enable enables the bus
func (b *Bus) Enable() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.enabled = true
}

// Disable disables the bus
func (b *Bus) Disable() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.enabled = false
}

// Returns a string representation of the bus
func (b *Bus) String() string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	status := ""
	if !b.enabled {
		status = " [disabled]"
	}

	return fmt.Sprintf("%s: %s -> %s%s", b.name, b.source.Name(), b.destination.Name(), status)
}

// =============================================================================
// Interconnect (Bus Switch)
// =============================================================================

// Interconnect allows selecting between multiple source ports to drive a destination
type Interconnect struct {
	name        string
	sources     []Port
	destination Port
	selectedIdx int
	enabled     bool
	mu          sync.RWMutex
}

// Creates a new interconnect
func NewInterconnect(name string, destination Port) *Interconnect {
	return &Interconnect{
		name:        name,
		sources:     make([]Port, 0),
		destination: destination,
		selectedIdx: -1,
		enabled:     true,
	}
}

// Returns the interconnect name
func (ic *Interconnect) Name() string {
	return ic.name
}

// AddSource adds a source port to the interconnect
func (ic *Interconnect) AddSource(port Port) int {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	idx := len(ic.sources)
	ic.sources = append(ic.sources, port)
	return idx
}

// Select chooses which source bus to use (-1 for none)
func (ic *Interconnect) Select(index int) error {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	if index < -1 || index >= len(ic.sources) {
		return fmt.Errorf("source index %d out of range [-1, %d)", index, len(ic.sources))
	}

	ic.selectedIdx = index
	return nil
}

// Returns the currently selected source index (-1 if none)
func (ic *Interconnect) Selected() int {
	ic.mu.RLock()
	defer ic.mu.RUnlock()
	return ic.selectedIdx
}

// Returns the currently selected source port (nil if none)
func (ic *Interconnect) SelectedSource() Port {
	ic.mu.RLock()
	defer ic.mu.RUnlock()

	if ic.selectedIdx < 0 || ic.selectedIdx >= len(ic.sources) {
		return nil
	}
	return ic.sources[ic.selectedIdx]
}

// Returns the currently selected source (implements Connection)
func (ic *Interconnect) Source() Port {
	return ic.SelectedSource()
}

// Returns the destination port (implements Connection)
func (ic *Interconnect) Destination() Port {
	return ic.destination
}

// Transfer copies from selected source to destination (implements Connection)
func (ic *Interconnect) Transfer() error {
	ic.mu.RLock()
	defer ic.mu.RUnlock()

	if !ic.enabled {
		return nil
	}

	if ic.selectedIdx < 0 || ic.selectedIdx >= len(ic.sources) {
		// No source selected, put destination in tristate
		ic.destination.SetTristate(true)
		return nil
	}

	source := ic.sources[ic.selectedIdx]
	if source.IsTristate() {
		ic.destination.SetTristate(true)
		return nil
	}

	ic.destination.SetTristate(false)
	return ic.destination.SetValue(source.GetValue())
}

// Returns whether the interconnect is enabled
func (ic *Interconnect) IsEnabled() bool {
	ic.mu.RLock()
	defer ic.mu.RUnlock()
	return ic.enabled
}

// Enable enables the interconnect
func (ic *Interconnect) Enable() {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	ic.enabled = true
}

// Disable disables the interconnect
func (ic *Interconnect) Disable() {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	ic.enabled = false
}

// Returns a string representation of the interconnect
func (ic *Interconnect) String() string {
	ic.mu.RLock()
	defer ic.mu.RUnlock()

	selected := "none"
	if ic.selectedIdx >= 0 && ic.selectedIdx < len(ic.sources) {
		selected = ic.sources[ic.selectedIdx].Name()
	}

	return fmt.Sprintf("%s: [%d sources] -> %s (selected: %s)",
		ic.name, len(ic.sources), ic.destination.Name(), selected)
}

// =============================================================================
// Circuit (Component Collection)
// =============================================================================

// Represents a collection of components and their interconnections
type Circuit struct {
	name        string
	components  map[string]Component
	connections []Connection
	mu          sync.RWMutex
}

// Creates a new circuit
func NewCircuit(name string) *Circuit {
	return &Circuit{
		name:        name,
		components:  make(map[string]Component),
		connections: make([]Connection, 0),
	}
}

// Returns the circuit name
func (c *Circuit) Name() string {
	return c.name
}

// AddComponent adds a component to the circuit
func (c *Circuit) AddComponent(comp Component) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.components[comp.Name()]; exists {
		return fmt.Errorf("component %q already exists in circuit", comp.Name())
	}

	c.components[comp.Name()] = comp
	return nil
}

// Returns a component by name
func (c *Circuit) GetComponent(name string) (Component, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	comp, exists := c.components[name]
	if !exists {
		return nil, fmt.Errorf("component %q not found in circuit", name)
	}
	return comp, nil
}

// Returns all components
func (c *Circuit) Components() []Component {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]Component, 0, len(c.components))
	for _, comp := range c.components {
		result = append(result, comp)
	}
	return result
}

// Creates a bus between two ports and adds it to the circuit
func (c *Circuit) Connect(name string, source, destination Port, opts ...BusOption) *Bus {
	c.mu.Lock()
	defer c.mu.Unlock()

	bus := NewBus(name, source, destination, opts...)
	c.connections = append(c.connections, bus)
	return bus
}

// AddConnection adds an existing connection to the circuit
func (c *Circuit) AddConnection(conn Connection) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connections = append(c.connections, conn)
}

// Returns all connections
func (c *Circuit) Connections() []Connection {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]Connection, len(c.connections))
	copy(result, c.connections)
	return result
}

// Propagate transfers all connection values
func (c *Circuit) Propagate() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, conn := range c.connections {
		if err := conn.Transfer(); err != nil {
			return fmt.Errorf("failed to transfer connection: %w", err)
		}
	}
	return nil
}

// Clock advances all components by one cycle
func (c *Circuit) Clock() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, comp := range c.components {
		if err := comp.Clock(); err != nil {
			return fmt.Errorf("failed to clock component %q: %w", comp.Name(), err)
		}
	}
	return nil
}

// Reset resets all components
func (c *Circuit) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, comp := range c.components {
		comp.Reset()
	}
}

// Returns a string representation of the circuit
func (c *Circuit) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return fmt.Sprintf("Circuit %s: %d components, %d connections",
		c.name, len(c.components), len(c.connections))
}
