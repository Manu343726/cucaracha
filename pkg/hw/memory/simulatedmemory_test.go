package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Memory_Interface(t *testing.T) {
	var _ Memory = &SimulatedMemory{}
}

func Test_SimulatedMemory_ReadWrite(t *testing.T) {
	initialData := []byte{0x00, 0x01, 0x02, 0x03, 0x04}
	mem := NewSimulatedMemory(initialData)

	// Test initial data
	for addr, expected := range initialData {
		value, err := mem.ReadByte(uint32(addr))
		assert.NoError(t, err)
		assert.Equal(t, expected, value)
	}

	// Test write
	mem.WriteByte(2, 0xFF)
	value, err := mem.ReadByte(2)
	assert.NoError(t, err)
	assert.Equal(t, byte(0xFF), value)

	// Test out-of-bounds read
	value, err = mem.ReadByte(10)
	assert.Error(t, err)
	assert.Equal(t, byte(0x00), value)

	// Test out-of-bounds write (should not panic)
	err = mem.WriteByte(10, 0xAA)
	assert.Error(t, err)
	// Test size
	assert.Equal(t, len(initialData), mem.Size())

	// Test reset
	mem.Reset()
	for addr := range initialData {
		value, err := mem.ReadByte(uint32(addr))
		assert.NoError(t, err)
		assert.Equal(t, byte(0x00), value)
	}

	// Test write of out-of-bounds address
	err = mem.WriteByte(uint32(len(initialData)+5), 0x55)
	assert.Error(t, err)

	// Test read of out-of-bounds address
	_, err = mem.ReadByte(uint32(len(initialData) + 5))
	assert.Error(t, err)

	// Test Ranges
	ranges := mem.Ranges()
	assert.Len(t, ranges, 1)
	assert.Equal(t, uint32(0), ranges[0].Start)
	assert.Equal(t, uint32(len(initialData)), ranges[0].Size)
	assert.Equal(t, FlagReadable|FlagWritable|FlagExecutable, ranges[0].Flags)
}
