// Package memmap provides memory-mapped address space management for the CPU.
//
// The MemoryMap type manages the CPU's address space, routing memory accesses
// to RAM or memory-mapped peripherals based on address ranges.
package memory

import "fmt"

// Memory defines the minimal contract the CPU or a peripheral needs to talk to an external
// memory device.
type Memory interface {
	// Returns the address ranges where operations are valid
	Ranges() []Range
	// Reads a byte from memory at the given address.
	ReadByte(addr uint32) (byte, error)
	// Writes a byte to memory at the given address.
	WriteByte(addr uint32, value byte) error
	// Total memory size in bytes
	Size() int
	// Clears all memory contents to zero.
	Reset() error
}

// Reads data from memory starting at the given address and saves it into the provided buffer.
func ReadTo(mem Memory, addr uint32, buffer []byte) error {
	for i := 0; i < len(buffer); i++ {
		b, err := mem.ReadByte(addr + uint32(i))
		if err != nil {
			return fmt.Errorf("memory read failed at address 0x%08X: %w", addr+uint32(i), err)
		}
		buffer[i] = b
	}
	return nil
}

// Reads data from memory starting at the given address.
func Read(mem Memory, addr uint32, size int) ([]byte, error) {
	data := make([]byte, size)
	if err := ReadTo(mem, addr, data); err != nil {
		return nil, err
	}
	return data, nil
}

// Writes data to memory starting at the given address.
func Write(mem Memory, addr uint32, data []byte) error {
	for i, b := range data {
		if err := mem.WriteByte(addr+uint32(i), b); err != nil {
			return fmt.Errorf("memory write failed at address 0x%08X: %w", addr+uint32(i), err)
		}
	}
	return nil
}

// Writes an unsigned 32 bit integer to memory in little-endian format.
func WriteUint32(mem Memory, addr uint32, value uint32) error {
	if err := mem.WriteByte(addr, byte(value&0xFF)); err != nil {
		return fmt.Errorf("uint32 write failed: byte 0 write failed: %w", err)
	}
	if err := mem.WriteByte(addr+1, byte((value>>8)&0xFF)); err != nil {
		return fmt.Errorf("uint32 write failed: byte 1 write failed: %w", err)
	}
	if err := mem.WriteByte(addr+2, byte((value>>16)&0xFF)); err != nil {
		return fmt.Errorf("uint32 write failed: byte 2 write failed: %w", err)
	}
	if err := mem.WriteByte(addr+3, byte((value>>24)&0xFF)); err != nil {
		return fmt.Errorf("uint32 write failed: byte 3 write failed: %w", err)
	}
	return nil
}

// Reads an unsigned 32 bit integer from memory in little-endian format.
func ReadUint32(mem Memory, addr uint32) (uint32, error) {
	var value uint32

	if b, err := mem.ReadByte(addr); err != nil {
		return 0, fmt.Errorf("uint32 read failed: byte 0 read failed: %w", err)
	} else {
		value = value | uint32(b)
	}

	if b, err := mem.ReadByte(addr + 1); err != nil {
		return 0, fmt.Errorf("uint32 read failed: byte 1 read failed: %w", err)
	} else {
		value = value | uint32(b)<<8
	}

	if b, err := mem.ReadByte(addr + 2); err != nil {
		return 0, fmt.Errorf("uint32 read failed: byte 2 read failed: %w", err)
	} else {
		value = value | uint32(b)<<16
	}

	if b, err := mem.ReadByte(addr + 3); err != nil {
		return 0, fmt.Errorf("uint32 read failed: byte 3 read failed: %w", err)
	} else {
		value = value | uint32(b)<<24
	}

	return value, nil
}
