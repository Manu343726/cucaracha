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
	data, err := mem.Read(0, len(initialData))
	assert.NoError(t, err)
	assert.Equal(t, initialData, data)

	// Test write
	err = mem.Write(2, []byte{0xFF})
	assert.NoError(t, err)
	value, err := mem.Read(2, 1)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0xFF}, value)

	// Test out-of-bounds read
	value, err = mem.Read(10, 1)
	assert.Error(t, err)

	// Test out-of-bounds write (should not panic)
	err = mem.Write(10, []byte{0xAA})
	assert.Error(t, err)
	// Test size
	assert.Equal(t, len(initialData), mem.Size())

	// Test reset
	mem.Reset()
	data, err = mem.Read(0, len(initialData))
	assert.NoError(t, err)
	for _, b := range data {
		assert.Equal(t, byte(0x00), b)
	}

	// Test write of out-of-bounds address
	err = mem.Write(uint32(len(initialData)+5), []byte{0x55})
	assert.Error(t, err)

	// Test read of out-of-bounds address
	_, err = mem.Read(uint32(len(initialData)+5), 1)
	assert.Error(t, err)

	// Test Ranges
	ranges := mem.Ranges()
	assert.Len(t, ranges, 1)
	assert.Equal(t, uint32(0), ranges[0].Start)
	assert.Equal(t, uint32(len(initialData)), ranges[0].Size)
	assert.Equal(t, FlagReadable|FlagWritable|FlagExecutable, ranges[0].Flags)
}
