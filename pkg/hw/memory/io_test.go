package memory

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Reader(t *testing.T) {
	data := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}
	mem := NewSimulatedMemory(data)

	reader := NewMemoryReader(MemorySlice(mem))

	buf := make([]byte, 4)
	n, err := reader.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, []byte{0x00, 0x01, 0x02, 0x03}, buf)

	n, err = reader.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, []byte{0x04, 0x05, 0x06, 0x07}, buf)

	n, err = reader.Read(buf)
	assert.Equal(t, 2, n)
	assert.Equal(t, io.EOF, err)
}

func Test_Writer(t *testing.T) {
	mem := NewSimulatedMemory(make([]byte, 10))
	writer := NewMemoryWriter(MemorySlice(mem))

	data := []byte{0xAA, 0xBB, 0xCC, 0xDD}
	n, err := writer.Write(data)
	assert.NoError(t, err)
	assert.Equal(t, 4, n)

	expected := []byte{0xAA, 0xBB, 0xCC, 0xDD, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	actual, _ := MemorySlice(mem).ReadAll()
	assert.Equal(t, expected, actual)

	n, err = writer.Write(data)
	assert.NoError(t, err)
	assert.Equal(t, 4, n)

	expected = []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xAA, 0xBB, 0xCC, 0xDD, 0x00, 0x00}
	actual, _ = MemorySlice(mem).ReadAll()
	assert.Equal(t, expected, actual)

	n, err = writer.Write(data)
	assert.Equal(t, 2, n)
	assert.Equal(t, io.EOF, err)

	expected = []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xAA, 0xBB, 0xCC, 0xDD, 0xAA, 0xBB}
	actual, _ = MemorySlice(mem).ReadAll()
	assert.Equal(t, expected, actual)
}
