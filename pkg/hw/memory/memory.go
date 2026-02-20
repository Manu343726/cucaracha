// Package memmap provides memory-mapped address space management for the CPU.
//
// The MemoryMap type manages the CPU's address space, routing memory accesses
// to RAM or memory-mapped peripherals based on address ranges.
package memory

// Memory defines the minimal contract the CPU or a peripheral needs to talk to an external
// memory device.
type Memory interface {
	// Returns the address ranges where operations are valid
	Ranges() []Range
	// Reads data from memory starting at the given address.
	Read(addr uint32, size int) ([]byte, error)
	// Writes data to memory starting at the given address.
	Write(addr uint32, data []byte) error
	// Total memory size in bytes
	Size() int
	// Clears all memory contents to zero.
	Reset() error
}

// Reads data from memory starting at the given address and saves it into the provided buffer.
func ReadTo(mem Memory, addr uint32, buffer []byte) error {
	data, err := mem.Read(addr, len(buffer))
	if err != nil {
		return err
	}
	copy(buffer, data)
	return nil
}

// Reads data from memory starting at the given address.
func Read(mem Memory, addr uint32, size int) ([]byte, error) {
	return mem.Read(addr, size)
}

// Writes data to memory starting at the given address.
func Write(mem Memory, addr uint32, data []byte) error {
	return mem.Write(addr, data)
}

// Writes an unsigned 32 bit integer to memory in little-endian format.
func WriteUint32(mem Memory, addr uint32, value uint32) error {
	data := []byte{
		byte(value & 0xFF),
		byte((value >> 8) & 0xFF),
		byte((value >> 16) & 0xFF),
		byte((value >> 24) & 0xFF),
	}
	if err := mem.Write(addr, data); err != nil {
		return err
	}

	return nil
}

// Reads an unsigned 32 bit integer from memory in little-endian format.
func ReadUint32(mem Memory, addr uint32) (uint32, error) {
	data, err := mem.Read(addr, 4)
	if err != nil {
		return 0, err
	}
	if len(data) != 4 {
		return 0, err
	}
	value := uint32(data[0]) | (uint32(data[1]) << 8) | (uint32(data[2]) << 16) | (uint32(data[3]) << 24)

	return value, nil
}
