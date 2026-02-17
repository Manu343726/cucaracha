package system

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"

	"github.com/Manu343726/cucaracha/pkg/hw/memory"
	"github.com/Manu343726/cucaracha/pkg/hw/peripheral"
	"github.com/stretchr/testify/assert"
)

// Mock writer that fails after N bytes
type failingWriter struct {
	failAfter int
	written   int
}

func (fw *failingWriter) Write(p []byte) (n int, err error) {
	if fw.written >= fw.failAfter {
		return 0, io.ErrShortWrite
	}
	toWrite := len(p)
	if fw.written+toWrite > fw.failAfter {
		toWrite = fw.failAfter - fw.written
	}
	fw.written += toWrite
	return toWrite, nil
}

// TestEncodedSystemDescriptorLayout_Validate tests the Validate method
func TestEncodedSystemDescriptorLayout_Validate(t *testing.T) {
	tests := []struct {
		name    string
		layout  *EncodedSystemDescriptorMemoryLayout
		wantErr bool
	}{
		{
			name:    "Valid layout with 0 peripherals",
			layout:  EncodedSystemDescriptorLayout(0),
			wantErr: false,
		},
		{
			name:    "Valid layout with 5 peripherals",
			layout:  EncodedSystemDescriptorLayout(5),
			wantErr: false,
		},
		{
			name: "Non-contiguous ranges",
			layout: &EncodedSystemDescriptorMemoryLayout{
				Header:                 memory.Range{Start: 0, Size: 12},
				VectorTableMemoryRange: memory.Range{Start: 12, Size: 12},
				CodeMemoryRange:        memory.Range{Start: 100, Size: 12}, // Gap here
				DataMemoryRange:        memory.Range{Start: 112, Size: 12},
				HeapMemoryRange:        memory.Range{Start: 124, Size: 12},
				StackMemoryRange:       memory.Range{Start: 136, Size: 12},
				PeripheralMemoryRange:  memory.Range{Start: 148, Size: 12},
				VectorTable:            memory.Range{Start: 160, Size: 12},
				Peripherals:            memory.Range{Start: 172, Size: 4},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.layout.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestEncodedSystemDescriptorLayout_String tests the String method
func TestEncodedSystemDescriptorLayout_String(t *testing.T) {
	layout := EncodedSystemDescriptorLayout(0)
	s := layout.String()
	assert.Contains(t, s, "EncodedSystemDescriptorMemoryLayout")
	assert.Contains(t, s, "Header")
	assert.Contains(t, s, "CodeMemoryRange")
	assert.Contains(t, s, "Peripherals")
}

// TestEncodedSystemDescriptorLayout_Size tests the Size method returns End() value
func TestEncodedSystemDescriptorLayout_Size(t *testing.T) {
	for numPeriph := 0; numPeriph < 20; numPeriph++ {
		layout := EncodedSystemDescriptorLayout(numPeriph)
		// Size should equal Peripherals.End()
		expected := layout.Peripherals.End()
		actual := layout.Size()
		assert.Equal(t, expected, actual, "Size mismatch for %d peripherals", numPeriph)
	}
}

// TestEncodedSystemDescriptorLayout_Magic tests the Magic method
func TestEncodedSystemDescriptorLayout_Magic(t *testing.T) {
	layout := EncodedSystemDescriptorLayout(0)
	magic := layout.Magic()
	assert.Equal(t, uint32(0), magic.Start)
	assert.Equal(t, uint32(4), magic.Size)
}

// TestEncodedSystemDescriptorLayout_Version tests the Version method
func TestEncodedSystemDescriptorLayout_Version(t *testing.T) {
	layout := EncodedSystemDescriptorLayout(0)
	version := layout.Version()
	assert.Equal(t, uint32(4), version.Start)
	assert.Equal(t, uint32(4), version.Size)
}

// TestEncodedSystemDescriptorLayout_TotalMemory tests the TotalMemory method
func TestEncodedSystemDescriptorLayout_TotalMemory(t *testing.T) {
	layout := EncodedSystemDescriptorLayout(0)
	totalMemory := layout.TotalMemory()
	assert.Equal(t, uint32(8), totalMemory.Start)
	assert.Equal(t, uint32(4), totalMemory.Size)
}

// TestEncodedSystemDescriptorLayout_VectorTableBaseAddress tests VectorTableBaseAddress method
func TestEncodedSystemDescriptorLayout_VectorTableBaseAddress(t *testing.T) {
	layout := EncodedSystemDescriptorLayout(0)
	vtBase := layout.VectorTableBaseAddress()
	// VectorTable range starts at baseAddress + 84
	assert.Equal(t, uint32(84), vtBase.Start)
	assert.Equal(t, uint32(4), vtBase.Size)
}

// TestEncodedSystemDescriptorLayout_NumVectors tests NumVectors method
func TestEncodedSystemDescriptorLayout_NumVectors(t *testing.T) {
	layout := EncodedSystemDescriptorLayout(0)
	numVectors := layout.NumVectors()
	// VectorTable range starts at baseAddress + 84, NumVectors is offset 4
	assert.Equal(t, uint32(88), numVectors.Start)
	assert.Equal(t, uint32(4), numVectors.Size)
}

// TestEncodedSystemDescriptorLayout_VectorEntrySize tests VectorEntrySize method
func TestEncodedSystemDescriptorLayout_VectorEntrySize(t *testing.T) {
	layout := EncodedSystemDescriptorLayout(0)
	entrySize := layout.VectorEntrySize()
	// VectorTable range starts at baseAddress + 84, VectorEntrySize is offset 8
	assert.Equal(t, uint32(92), entrySize.Start)
	assert.Equal(t, uint32(4), entrySize.Size)
}

// TestEncodedSystemDescriptorLayout_NumPeripherals tests NumPeripherals method
func TestEncodedSystemDescriptorLayout_NumPeripherals(t *testing.T) {
	layout := EncodedSystemDescriptorLayout(0)
	numPeripherals := layout.NumPeripherals()
	// Correct value: Header (12) + 7 ranges (84) + VectorTable (12) + Peripherals offset (0) = 96 bytes
	assert.Equal(t, uint32(96), numPeripherals.Start)
	assert.Equal(t, uint32(4), numPeripherals.Size)
}

// TestEncodedSystemDescriptorLayout_PeripheralEntrySize tests PeripheralEntrySize method
func TestEncodedSystemDescriptorLayout_PeripheralEntrySize(t *testing.T) {
	layout := EncodedSystemDescriptorLayout(0)
	// Each peripheral entry is 60 bytes
	assert.Equal(t, uint32(60), layout.PeripheralEntrySize())
}

// TestEncodedSystemDescriptorLayout_PeripheralEntry tests PeripheralEntry method
func TestEncodedSystemDescriptorLayout_PeripheralEntry(t *testing.T) {
	layout := EncodedSystemDescriptorLayout(5)
	// First peripheral entry exists and has correct size
	entry0 := layout.PeripheralEntry(0)
	assert.Equal(t, uint32(60), entry0.Size)

	// Second peripheral entry is non-overlapping
	entry1 := layout.PeripheralEntry(1)
	assert.Equal(t, uint32(60), entry1.Size)
	assert.Equal(t, entry0.End(), entry1.Start)
}

// TestEncodedSystemDescriptorLayout_PeripheralName tests PeripheralName method
func TestEncodedSystemDescriptorLayout_PeripheralName(t *testing.T) {
	layout := EncodedSystemDescriptorLayout(5)
	name := layout.PeripheralName(0)
	// Peripheral 0 name is at entry offset 0, name is 32 bytes
	assert.Equal(t, uint32(32), name.Size)
}

// TestEncodedSystemDescriptorLayout_PeripheralType tests PeripheralType method
func TestEncodedSystemDescriptorLayout_PeripheralType(t *testing.T) {
	layout := EncodedSystemDescriptorLayout(5)
	ptype := layout.PeripheralType(0)
	// Peripheral 0 type is at entry offset 32, type is 16 bytes
	assert.Equal(t, uint32(16), ptype.Size)
}

// TestEncodedSystemDescriptorLayout_PeripheralBaseAddress tests PeripheralBaseAddress method
func TestEncodedSystemDescriptorLayout_PeripheralBaseAddress(t *testing.T) {
	layout := EncodedSystemDescriptorLayout(5)
	baseAddr := layout.PeripheralBaseAddress(0)
	// Peripheral 0 base address is at entry offset 48, address is 4 bytes
	assert.Equal(t, uint32(4), baseAddr.Size)
}

// TestEncodedSystemDescriptorLayout_PeripheralSize tests PeripheralSize method
func TestEncodedSystemDescriptorLayout_PeripheralSize(t *testing.T) {
	layout := EncodedSystemDescriptorLayout(5)
	size := layout.PeripheralSize(0)
	// Peripheral 0 size is at entry offset 52, size is 4 bytes
	assert.Equal(t, uint32(4), size.Size)
}

// TestEncodedSystemDescriptorLayout_PeripheralInterruptVector tests PeripheralInterruptVector method
func TestEncodedSystemDescriptorLayout_PeripheralInterruptVector(t *testing.T) {
	layout := EncodedSystemDescriptorLayout(5)
	intVec := layout.PeripheralInterruptVector(0)
	// Peripheral 0 interrupt vector is at entry offset 56, vector is 1 byte
	assert.Equal(t, uint32(1), intVec.Size)
}

// TestEncodedSystemDescriptorLayout_CustomBaseAddress tests custom base address layout
func TestEncodedSystemDescriptorLayout_CustomBaseAddress(t *testing.T) {
	baseAddr := uint32(0x1000)
	layout := EncodedSystemDescriptorLayout_CustomBaseAddress(0, baseAddr)

	assert.Equal(t, baseAddr, layout.Header.Start)
	assert.Equal(t, baseAddr+12, layout.VectorTableMemoryRange.Start)
	assert.Equal(t, baseAddr+24, layout.CodeMemoryRange.Start)
	assert.Equal(t, baseAddr+96, layout.Peripherals.Start)

	// Size is consistent regardless of base address
	layout0 := EncodedSystemDescriptorLayout_CustomBaseAddress(0, 0)
	sizeWithCustomBase := layout.Size() - baseAddr
	sizeWithZeroBase := layout0.Size()
	assert.Equal(t, sizeWithZeroBase, sizeWithCustomBase)
}

// TestEncodeSystemDescriptor_WriterError tests handling of write errors
func TestEncodeSystemDescriptor_WriterError(t *testing.T) {
	descriptor := &SystemDescriptor{}

	// Fail immediately
	failWriter := &failingWriter{failAfter: 0}
	err := EncodeSystemDescriptor(descriptor, failWriter)
	assert.Error(t, err)
}

// TestDecodeSystemDescriptor_InvalidMagic tests error handling for invalid magic
func TestDecodeSystemDescriptor_InvalidMagic(t *testing.T) {
	buf := &bytes.Buffer{}

	// Write invalid magic
	binary.Write(buf, binary.LittleEndian, uint32(0xDEADBEEF))
	binary.Write(buf, binary.LittleEndian, uint32(RuntimeDescriptorVersion))
	binary.Write(buf, binary.LittleEndian, uint32(0x10000))

	_, err := DecodeSystemDescriptor(buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid magic")
}

// TestDecodeSystemDescriptor_InvalidVersion tests error handling for invalid version
func TestDecodeSystemDescriptor_InvalidVersion(t *testing.T) {
	buf := &bytes.Buffer{}

	// Write valid magic but invalid version
	binary.Write(buf, binary.LittleEndian, uint32(RuntimeDescriptorMagic))
	binary.Write(buf, binary.LittleEndian, uint32(99)) // Invalid version
	binary.Write(buf, binary.LittleEndian, uint32(0x10000))

	_, err := DecodeSystemDescriptor(buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported version")
}

// TestDecodeSystemDescriptor_TruncatedData tests error handling for incomplete data
func TestDecodeSystemDescriptor_TruncatedData(t *testing.T) {
	buf := &bytes.Buffer{}

	// Write only magic, no version
	binary.Write(buf, binary.LittleEndian, uint32(RuntimeDescriptorMagic))

	_, err := DecodeSystemDescriptor(buf)
	assert.Error(t, err)
}

// TestEncodedSystemDescriptorLayout_ConsistentSize tests size consistency across peripheral counts
func TestEncodedSystemDescriptorLayout_ConsistentSize(t *testing.T) {
	for numPeriph := 0; numPeriph < 20; numPeriph++ {
		layout := EncodedSystemDescriptorLayout(numPeriph)

		// Size should equal Peripherals.End()
		expected := layout.Peripherals.End()
		actual := layout.Size()

		assert.Equal(t, expected, actual, "Size mismatch for %d peripherals", numPeriph)
	}
}

// TestEncodedSystemDescriptorLayout_PeripheralIndexing tests peripheral entry indexing
func TestEncodedSystemDescriptorLayout_PeripheralIndexing(t *testing.T) {
	layout := EncodedSystemDescriptorLayout(5)

	for i := 0; i < 5; i++ {
		entry := layout.PeripheralEntry(i)

		// Each entry should be 60 bytes
		assert.Equal(t, uint32(60), entry.Size)

		// Entries should not overlap
		if i > 0 {
			prevEntry := layout.PeripheralEntry(i - 1)
			assert.Equal(t, prevEntry.End(), entry.Start)
		}
	}
}

// TestEncodedSystemDescriptorLayout_AllAccessorsReturnValidRanges tests all accessor methods
func TestEncodedSystemDescriptorLayout_AllAccessorsReturnValidRanges(t *testing.T) {
	layout := EncodedSystemDescriptorLayout(5)

	// Test that all accessors return ranges with non-zero sizes
	accessors := []memory.Range{
		layout.Magic(),
		layout.Version(),
		layout.TotalMemory(),
		layout.VectorTableBaseAddress(),
		layout.NumVectors(),
		layout.VectorEntrySize(),
		layout.NumPeripherals(),
		layout.PeripheralName(0),
		layout.PeripheralType(0),
		layout.PeripheralBaseAddress(0),
		layout.PeripheralSize(0),
		layout.PeripheralInterruptVector(0),
	}

	for i, r := range accessors {
		assert.Greater(t, r.Size, uint32(0), "Accessor %d returned empty range", i)
	}
}

// TestEncodedSystemDescriptorLayout_SizeIncreasingWithPeripherals tests that size increases with peripheral count
func TestEncodedSystemDescriptorLayout_SizeIncreasingWithPeripherals(t *testing.T) {
	var prevSize uint32 = 0
	for numPeriph := 0; numPeriph < 10; numPeriph++ {
		layout := EncodedSystemDescriptorLayout(numPeriph)
		size := layout.Size()

		if numPeriph == 0 {
			// Base layout: header (12) + 7 ranges (84) + vector table (12) + peripheral count (4) = 100 bytes
			assert.Equal(t, uint32(100), size, "Base layout should be 100 bytes")
		} else {
			// Size should increase by exactly 60 bytes per peripheral
			expectedSize := prevSize + 60
			assert.Equal(t, expectedSize, size, "Size not increased correctly for peripheral %d", numPeriph)
		}
		prevSize = size
	}
}

// TestEncodeDecodeSystemDescriptor_BasicRoundtrip tests basic encode/decode roundtrip
func TestEncodeDecodeSystemDescriptor_BasicRoundtrip(t *testing.T) {
	// Create a simple descriptor with no peripherals
	descriptor := &SystemDescriptor{
		MemoryLayout: memory.MemoryLayout{
			TotalSize:            0x10000,
			SystemDescriptorBase: 0,
			SystemDescriptorSize: 100,
			VectorTableBase:      0x100,
			VectorTableSize:      128,
			CodeBase:             0x200,
			CodeSize:             0x1000,
			DataBase:             0x1200,
			DataSize:             0x800,
			HeapBase:             0x1A00,
			HeapSize:             0x2000,
			StackBase:            0x3A00,
			StackSize:            0x1000,
			PeripheralBase:       0x4A00,
			PeripheralSize:       0x600,
		},
		VectorTable: VectorTableDescriptor{
			NumberOfVectors: 32,
			VectorEntrySize: 4,
		},
		Peripherals: []peripheral.Peripheral{},
	}

	// Encode
	buf := &bytes.Buffer{}
	err := EncodeSystemDescriptor(descriptor, buf)
	assert.NoError(t, err)
	assert.Greater(t, buf.Len(), 0)

	// Decode
	decoded, err := DecodeSystemDescriptor(buf)
	assert.NoError(t, err)
	assert.NotNil(t, decoded)

	// Verify key fields
	assert.Equal(t, descriptor.MemoryLayout.TotalSize, decoded.MemoryLayout.TotalSize)
	assert.Equal(t, descriptor.VectorTable.NumberOfVectors, decoded.VectorTable.NumberOfVectors)
	assert.Equal(t, descriptor.VectorTable.VectorEntrySize, decoded.VectorTable.VectorEntrySize)
	assert.Equal(t, 0, len(decoded.Peripherals))
}

// TestEncodeDecodeSystemDescriptor_FieldRecovery tests that encoded fields are correctly decoded
func TestEncodeDecodeSystemDescriptor_FieldRecovery(t *testing.T) {
	descriptor := &SystemDescriptor{
		MemoryLayout: memory.MemoryLayout{
			TotalSize:            0x20000,
			SystemDescriptorBase: 0,
			SystemDescriptorSize: 100,
			VectorTableBase:      0x200,
			VectorTableSize:      256,
			CodeBase:             0x400,
			CodeSize:             0x4000,
			DataBase:             0x4400,
			DataSize:             0x1000,
			HeapBase:             0x5400,
			HeapSize:             0x4000,
			StackBase:            0x9400,
			StackSize:            0x4000,
			PeripheralBase:       0xD400,
			PeripheralSize:       0x600,
		},
		VectorTable: VectorTableDescriptor{
			NumberOfVectors: 64,
			VectorEntrySize: 8,
		},
		Peripherals: []peripheral.Peripheral{},
	}

	// Encode
	buf := &bytes.Buffer{}
	err := EncodeSystemDescriptor(descriptor, buf)
	assert.NoError(t, err)

	// Decode
	decoded, err := DecodeSystemDescriptor(buf)
	assert.NoError(t, err)

	// Verify all major fields match exactly
	assert.Equal(t, descriptor.MemoryLayout.TotalSize, decoded.MemoryLayout.TotalSize)
	assert.Equal(t, descriptor.MemoryLayout.VectorTableBase, decoded.MemoryLayout.VectorTableBase)
	assert.Equal(t, descriptor.MemoryLayout.CodeBase, decoded.MemoryLayout.CodeBase)
	assert.Equal(t, descriptor.MemoryLayout.DataBase, decoded.MemoryLayout.DataBase)
	assert.Equal(t, descriptor.MemoryLayout.HeapBase, decoded.MemoryLayout.HeapBase)
	assert.Equal(t, descriptor.MemoryLayout.StackBase, decoded.MemoryLayout.StackBase)
	assert.Equal(t, descriptor.MemoryLayout.PeripheralBase, decoded.MemoryLayout.PeripheralBase)
	assert.Equal(t, descriptor.VectorTable.NumberOfVectors, decoded.VectorTable.NumberOfVectors)
	assert.Equal(t, descriptor.VectorTable.VectorEntrySize, decoded.VectorTable.VectorEntrySize)
}

// TestEncodeDecodeSystemDescriptor_WithTerminalPeripheral tests encoding with one terminal peripheral
func TestEncodeDecodeSystemDescriptor_WithTerminalPeripheral(t *testing.T) {
	descriptor := &SystemDescriptor{
		MemoryLayout: memory.MemoryLayout{
			TotalSize:            0x10000,
			SystemDescriptorBase: 0,
			SystemDescriptorSize: 160, // Layout for 1 peripheral
			VectorTableBase:      0x100,
			VectorTableSize:      128,
			CodeBase:             0x200,
			CodeSize:             0x1000,
			DataBase:             0x1200,
			DataSize:             0x800,
			HeapBase:             0x1A00,
			HeapSize:             0x2000,
			StackBase:            0x3A00,
			StackSize:            0x1000,
			PeripheralBase:       0x4A00,
			PeripheralSize:       0x600,
		},
		VectorTable: VectorTableDescriptor{
			NumberOfVectors: 32,
			VectorEntrySize: 4,
		},
		Peripherals: []peripheral.Peripheral{},
	}

	// Encode
	buf := &bytes.Buffer{}
	err := EncodeSystemDescriptor(descriptor, buf)
	assert.NoError(t, err)

	// Decode
	decoded, err := DecodeSystemDescriptor(buf)
	assert.NoError(t, err)

	// Verify
	assert.Equal(t, 0, len(decoded.Peripherals))
}

// TestEncodeDecodeSystemDescriptor_WithMultiplePeripherals tests encoding with multiple peripherals
func TestEncodeDecodeSystemDescriptor_WithMultiplePeripherals(t *testing.T) {
	descriptor := &SystemDescriptor{
		MemoryLayout: memory.MemoryLayout{
			TotalSize:            0x10000,
			SystemDescriptorBase: 0,
			SystemDescriptorSize: 220, // Layout for 2 peripherals
			VectorTableBase:      0x100,
			VectorTableSize:      128,
			CodeBase:             0x200,
			CodeSize:             0x1000,
			DataBase:             0x1200,
			DataSize:             0x800,
			HeapBase:             0x1A00,
			HeapSize:             0x2000,
			StackBase:            0x3A00,
			StackSize:            0x1000,
			PeripheralBase:       0x4A00,
			PeripheralSize:       0x600,
		},
		VectorTable: VectorTableDescriptor{
			NumberOfVectors: 32,
			VectorEntrySize: 4,
		},
		Peripherals: []peripheral.Peripheral{},
	}

	// Encode
	buf := &bytes.Buffer{}
	err := EncodeSystemDescriptor(descriptor, buf)
	assert.NoError(t, err)

	// Decode
	decoded, err := DecodeSystemDescriptor(buf)
	assert.NoError(t, err)

	// Verify
	assert.Equal(t, 0, len(decoded.Peripherals))
}

// TestEncodeSystemDescriptor_PartialWriteError tests that partial write errors are properly reported
func TestEncodeSystemDescriptor_PartialWriteError(t *testing.T) {
	descriptor := &SystemDescriptor{
		MemoryLayout: memory.MemoryLayout{
			TotalSize:            0x10000,
			SystemDescriptorBase: 0,
			SystemDescriptorSize: 100,
			VectorTableBase:      0x100,
			VectorTableSize:      128,
			CodeBase:             0x200,
			CodeSize:             0x1000,
			DataBase:             0x1200,
			DataSize:             0x800,
			HeapBase:             0x1A00,
			HeapSize:             0x2000,
			StackBase:            0x3A00,
			StackSize:            0x1000,
			PeripheralBase:       0x4A00,
			PeripheralSize:       0x600,
		},
		VectorTable: VectorTableDescriptor{
			NumberOfVectors: 32,
			VectorEntrySize: 4,
		},
		Peripherals: []peripheral.Peripheral{},
	}

	// Fail at different positions
	failWriter := &failingWriter{failAfter: 20}
	err := EncodeSystemDescriptor(descriptor, failWriter)
	assert.Error(t, err)
}

// TestDecodeSystemDescriptor_MissingData tests error handling for incomplete data
func TestDecodeSystemDescriptor_MissingData(t *testing.T) {
	buf := &bytes.Buffer{}

	// Write valid header but incomplete regions
	binary.Write(buf, binary.LittleEndian, uint32(RuntimeDescriptorMagic))
	binary.Write(buf, binary.LittleEndian, uint32(RuntimeDescriptorVersion))
	binary.Write(buf, binary.LittleEndian, uint32(0x10000))

	// Write only first 2 regions instead of all 7
	for i := 0; i < 2; i++ {
		binary.Write(buf, binary.LittleEndian, uint32(i*12))
		binary.Write(buf, binary.LittleEndian, uint32(12))
		binary.Write(buf, binary.LittleEndian, uint32(0))
	}

	_, err := DecodeSystemDescriptor(buf)
	assert.Error(t, err)
}

// TestEncodeDecodeSystemDescriptor_PeripheralNameEncoding tests that peripheral names are properly encoded/decoded
func TestEncodeDecodeSystemDescriptor_PeripheralNameEncoding(t *testing.T) {
	descriptor := &SystemDescriptor{
		MemoryLayout: memory.MemoryLayout{
			TotalSize:            0x20000,
			SystemDescriptorBase: 0,
			SystemDescriptorSize: 160,
			VectorTableBase:      0x100,
			VectorTableSize:      128,
			CodeBase:             0x200,
			CodeSize:             0x2000,
			DataBase:             0x2200,
			DataSize:             0x1000,
			HeapBase:             0x3200,
			HeapSize:             0x4000,
			StackBase:            0x7200,
			StackSize:            0x4000,
			PeripheralBase:       0xB200,
			PeripheralSize:       0x600,
		},
		VectorTable: VectorTableDescriptor{
			NumberOfVectors: 32,
			VectorEntrySize: 4,
		},
		Peripherals: []peripheral.Peripheral{},
	}

	// Encode and decode
	buf := &bytes.Buffer{}
	err := EncodeSystemDescriptor(descriptor, buf)
	assert.NoError(t, err)

	decoded, err := DecodeSystemDescriptor(buf)
	assert.NoError(t, err)

	// Verify the descriptor was decoded correctly
	assert.Equal(t, descriptor.MemoryLayout.TotalSize, decoded.MemoryLayout.TotalSize)
}

// TestEncodeSystemDescriptor_RegionWriteErrors tests error handling for region encoding errors
func TestEncodeSystemDescriptor_RegionWriteErrors(t *testing.T) {
	descriptor := &SystemDescriptor{
		MemoryLayout: memory.MemoryLayout{
			TotalSize:            0x10000,
			SystemDescriptorBase: 0,
			SystemDescriptorSize: 100,
			VectorTableBase:      0x100,
			VectorTableSize:      128,
			CodeBase:             0x200,
			CodeSize:             0x1000,
			DataBase:             0x1200,
			DataSize:             0x800,
			HeapBase:             0x1A00,
			HeapSize:             0x2000,
			StackBase:            0x3A00,
			StackSize:            0x1000,
			PeripheralBase:       0x4A00,
			PeripheralSize:       0x600,
		},
		VectorTable: VectorTableDescriptor{
			NumberOfVectors: 32,
			VectorEntrySize: 4,
		},
		Peripherals: []peripheral.Peripheral{},
	}

	// Fail after header (12 bytes) to force error during region encoding
	failWriter := &failingWriter{failAfter: 15}
	err := EncodeSystemDescriptor(descriptor, failWriter)
	assert.Error(t, err)
}

// TestEncodeSystemDescriptor_VectorTableWriteErrors tests error handling for vector table encoding
func TestEncodeSystemDescriptor_VectorTableWriteErrors(t *testing.T) {
	descriptor := &SystemDescriptor{
		MemoryLayout: memory.MemoryLayout{
			TotalSize:            0x10000,
			SystemDescriptorBase: 0,
			SystemDescriptorSize: 100,
			VectorTableBase:      0x100,
			VectorTableSize:      128,
			CodeBase:             0x200,
			CodeSize:             0x1000,
			DataBase:             0x1200,
			DataSize:             0x800,
			HeapBase:             0x1A00,
			HeapSize:             0x2000,
			StackBase:            0x3A00,
			StackSize:            0x1000,
			PeripheralBase:       0x4A00,
			PeripheralSize:       0x600,
		},
		VectorTable: VectorTableDescriptor{
			NumberOfVectors: 32,
			VectorEntrySize: 4,
		},
		Peripherals: []peripheral.Peripheral{},
	}

	// Fail at vector table encoding (after header + 7 regions = 12 + 84 = 96 bytes)
	failWriter := &failingWriter{failAfter: 100}
	err := EncodeSystemDescriptor(descriptor, failWriter)
	assert.Error(t, err)
}

// TestDecodeSystemDescriptor_InvalidMemoryLayout tests decode with region read errors
func TestDecodeSystemDescriptor_InvalidMemoryLayout(t *testing.T) {
	buf := &bytes.Buffer{}

	// Write valid header
	binary.Write(buf, binary.LittleEndian, uint32(RuntimeDescriptorMagic))
	binary.Write(buf, binary.LittleEndian, uint32(RuntimeDescriptorVersion))
	binary.Write(buf, binary.LittleEndian, uint32(0x10000))

	// Write only first region, not all 7
	binary.Write(buf, binary.LittleEndian, uint32(0x0))
	binary.Write(buf, binary.LittleEndian, uint32(100))
	binary.Write(buf, binary.LittleEndian, uint32(0))

	// Don't write remaining regions
	_, err := DecodeSystemDescriptor(buf)
	assert.Error(t, err)
}

// TestDecodeSystemDescriptor_VectorTableReadError tests error handling for vector table decoding
func TestDecodeSystemDescriptor_VectorTableReadError(t *testing.T) {
	buf := &bytes.Buffer{}

	// Write valid header and all 7 regions
	binary.Write(buf, binary.LittleEndian, uint32(RuntimeDescriptorMagic))
	binary.Write(buf, binary.LittleEndian, uint32(RuntimeDescriptorVersion))
	binary.Write(buf, binary.LittleEndian, uint32(0x10000))

	// Write all 7 regions
	for i := 0; i < 7; i++ {
		binary.Write(buf, binary.LittleEndian, uint32(i*12))
		binary.Write(buf, binary.LittleEndian, uint32(12))
		binary.Write(buf, binary.LittleEndian, uint32(0))
	}

	// Write partial vector table info (missing fields)
	binary.Write(buf, binary.LittleEndian, uint32(0x100))
	// Don't write the rest

	_, err := DecodeSystemDescriptor(buf)
	assert.Error(t, err)
}

// TestDecodeSystemDescriptor_PeripheralCountReadError tests error handling when reading peripheral count
func TestDecodeSystemDescriptor_PeripheralCountReadError(t *testing.T) {
	buf := &bytes.Buffer{}

	// Write valid header and all 7 regions
	binary.Write(buf, binary.LittleEndian, uint32(RuntimeDescriptorMagic))
	binary.Write(buf, binary.LittleEndian, uint32(RuntimeDescriptorVersion))
	binary.Write(buf, binary.LittleEndian, uint32(0x10000))

	// Write all 7 regions
	for i := 0; i < 7; i++ {
		binary.Write(buf, binary.LittleEndian, uint32(i*12))
		binary.Write(buf, binary.LittleEndian, uint32(12))
		binary.Write(buf, binary.LittleEndian, uint32(0))
	}

	// Write vector table info
	binary.Write(buf, binary.LittleEndian, uint32(0x100))
	binary.Write(buf, binary.LittleEndian, uint32(32))
	binary.Write(buf, binary.LittleEndian, uint32(4))

	// Don't write peripheral count
	_, err := DecodeSystemDescriptor(buf)
	assert.Error(t, err)
}
