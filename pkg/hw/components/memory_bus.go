package components

import (
	"github.com/Manu343726/cucaracha/pkg/hw/component"
	"github.com/Manu343726/cucaracha/pkg/hw/memory"
)

// Memory defines the minimal contract the CPU needs to talk to an external
// memory device. Implementations can be simple byte-addressable memories or
// richer peripherals that expose the same API.
type Memory interface {
	ReadByte(addr uint32) byte
	WriteByte(addr uint32, value byte)
	Size() int
	Reset()
}

// MemoryPort augments Memory with the set of ports/pins typically used to wire
// the CPU to a memory device. Implementations may return nil for any of these
// if they are not port-based, but hardware-style memories like RAM expose
// concrete ports that can be connected.
type MemoryPort interface {
	Memory
	Data() *component.StandardPort
	Address() *component.StandardPort
	ReadWrite() *component.Pin
	Ready() *component.Pin
}

// Adapts a Memory implementation to the memory.Memory interface
type memoryAdapter struct {
	memory Memory
}

// NewMemoryAdapter creates a new memory.Memory adapter for the given Memory
// implementation
func NewMemoryAdapter(mem Memory) memory.Memory {
	return &memoryAdapter{memory: mem}
}

// ReadByte reads a byte from the memory at the given address
func (m *memoryAdapter) ReadByte(addr uint32) (byte, error) {
	return m.memory.ReadByte(addr), nil
}

// WriteByte writes a byte to the memory at the given address
func (m *memoryAdapter) WriteByte(addr uint32, value byte) error {
	m.memory.WriteByte(addr, value)
	return nil
}

// Size returns the total size of the memory in bytes
func (m *memoryAdapter) Size() int {
	return m.memory.Size()
}

// Reset resets the memory to its initial state
func (m *memoryAdapter) Reset() error {
	m.memory.Reset()
	return nil
}

// Ranges returns the address ranges where operations are valid
func (m *memoryAdapter) Ranges() []memory.Range {
	return []memory.Range{
		{Start: 0, Size: uint32(m.memory.Size())},
	}
}
