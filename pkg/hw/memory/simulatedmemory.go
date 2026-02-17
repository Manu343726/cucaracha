package memory

import "fmt"

// Implements Memory interface with a simple byte slice.
// Suitable for tests and software simulators such as the CPU interpreter.
type SimulatedMemory struct {
	data []byte
}

func (s *SimulatedMemory) Ranges() []Range {
	return []Range{
		{
			Start: 0,
			Size:  uint32(len(s.data)),
			Flags: FlagReadable | FlagWritable | FlagExecutable,
		},
	}
}

// Creates a new SimulatedMemory with the given initial data.
func NewSimulatedMemory(initialData []byte) Memory {
	mem := &SimulatedMemory{
		data: make([]byte, len(initialData)),
	}
	copy(mem.data, initialData)
	return mem
}

func (m *SimulatedMemory) ReadByte(addr uint32) (byte, error) {
	if int(addr) >= len(m.data) {
		return 0, fmt.Errorf("memory read out of bounds: addr=0x%X, valid range: %v", addr, m.Ranges()[0])
	}

	return m.data[addr], nil
}

func (m *SimulatedMemory) WriteByte(addr uint32, value byte) error {
	if int(addr) >= len(m.data) {
		return fmt.Errorf("memory write out of bounds: addr=0x%X, valid range: %v", addr, m.Ranges()[0])
	}

	m.data[addr] = value
	return nil
}

func (m *SimulatedMemory) Size() int {
	return len(m.data)
}

func (m *SimulatedMemory) Reset() error {
	for i := range m.data {
		m.data[i] = 0
	}
	return nil
}
