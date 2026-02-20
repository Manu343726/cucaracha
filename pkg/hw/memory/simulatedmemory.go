package memory

import (
	"log/slog"

	"github.com/Manu343726/cucaracha/pkg/utils/contract"
	"github.com/Manu343726/cucaracha/pkg/utils/logging"
)

// Implements Memory interface with a simple byte slice.
// Suitable for tests and software simulators such as the CPU interpreter.
type SimulatedMemory struct {
	contract.Base

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
		Base: contract.NewBase(log().Child("SimulatedMemory")),
		data: make([]byte, len(initialData)),
	}
	copy(mem.data, initialData)
	return mem
}

// Max number of bytes to log for read/write operations to avoid overwhelming the logs with large memory dumps.
const MaxLoggedDataSize = 32

func (m *SimulatedMemory) Read(addr uint32, size int) ([]byte, error) {
	if int(addr)+size > len(m.data) {
		return nil, m.Log().Errorf("memory read out of bounds: addr=0x%X, size=%d, valid range: %v", addr, size, m.Ranges()[0])
	}

	result := make([]byte, size)
	copy(result, m.data[addr:addr+uint32(size)])
	m.Log().Debug("read", logging.Address("address", addr), slog.Uint64("size", uint64(size)), logging.HexBytes("data", result[:min(size, MaxLoggedDataSize)]))

	return result, nil
}

func (m *SimulatedMemory) Write(addr uint32, data []byte) error {
	if int(addr)+len(data) > len(m.data) {
		return m.Log().Errorf("memory write out of bounds: addr=0x%X, size=%d, valid range: %v", addr, len(data), m.Ranges()[0])
	}

	m.Log().Debug("write", logging.Address("address", addr), slog.Uint64("size", uint64(len(data))), logging.HexBytes("data", data[:min(len(data), MaxLoggedDataSize)]))

	copy(m.data[addr:], data)
	return nil
}

func (m *SimulatedMemory) Size() int {
	return len(m.data)
}

func (m *SimulatedMemory) Reset() error {
	for i := range m.data {
		m.data[i] = 0
	}

	m.Log().Debug("reset")

	return nil
}
