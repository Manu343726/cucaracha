package system

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/hw/memory"
	"github.com/Manu343726/cucaracha/pkg/hw/peripheral"
	"github.com/Manu343726/cucaracha/pkg/hw/peripherals"
)

// Shows the memory layout of an encoded system descriptor
type EncodedSystemDescriptorMemoryLayout struct {
	Header                memory.Range
	CodeMemoryRange       memory.Range
	DataMemoryRange       memory.Range
	HeapMemoryRange       memory.Range
	StackMemoryRange      memory.Range
	PeripheralMemoryRange memory.Range
	VectorTable           memory.Range
	Peripherals           memory.Range
}

// Validates that memory layout of the encoded system descriptor is well-formed
func (l *EncodedSystemDescriptorMemoryLayout) Validate() error {
	if !memory.ContiguousRanges([]memory.Range{
		l.Header,
		l.CodeMemoryRange,
		l.DataMemoryRange,
		l.HeapMemoryRange,
		l.StackMemoryRange,
		l.PeripheralMemoryRange,
		l.VectorTable,
		l.Peripherals,
	}) {
		return fmt.Errorf("encoded system descriptor memory layout ranges are not contiguous: %v", l)
	}

	return nil
}

func (l *EncodedSystemDescriptorMemoryLayout) String() string {
	return fmt.Sprintf("EncodedSystemDescriptorMemoryLayout{\n"+
		"  Header: %s\n"+
		"  CodeMemoryRange: %s\n"+
		"  DataMemoryRange: %s\n"+
		"  HeapMemoryRange: %s\n"+
		"  StackMemoryRange: %s\n"+
		"  PeripheralMemoryRange: %s\n"+
		"  VectorTable: %s\n"+
		"  Peripherals: %s\n"+
		"}",
		l.Header.String(),
		l.CodeMemoryRange.String(),
		l.DataMemoryRange.String(),
		l.HeapMemoryRange.String(),
		l.StackMemoryRange.String(),
		l.PeripheralMemoryRange.String(),
		l.VectorTable.String(),
		l.Peripherals.String(),
	)
}

// Returns the total size of the encoded system descriptor in bytes
func (l *EncodedSystemDescriptorMemoryLayout) Size() uint32 {
	return l.Peripherals.End()
}

// Returns the range within the encoded descriptor where the magic number is stored
func (l *EncodedSystemDescriptorMemoryLayout) Magic() memory.Range {
	return l.Header.SubRange(0, 4)
}

// Returns the range within the encoded descriptor where the version is stored
func (l *EncodedSystemDescriptorMemoryLayout) Version() memory.Range {
	return l.Header.SubRange(4, 4)
}

// Returns the range within the encoded descriptor where the total memory size is stored
func (l *EncodedSystemDescriptorMemoryLayout) TotalMemory() memory.Range {
	return l.Header.SubRange(8, 4)
}

// Returns the range within the encoded descriptor where the vector table base address is stored
func (l *EncodedSystemDescriptorMemoryLayout) VectorTableBaseAddress() memory.Range {
	return l.VectorTable.SubRange(0, 4)
}

// Returns the range within the encoded descriptor where the number of interrupt vectors is stored
func (l *EncodedSystemDescriptorMemoryLayout) NumVectors() memory.Range {
	return l.VectorTable.SubRange(4, 4)
}

// Returns the range within the encoded descriptor where the vector entry size is stored
func (l *EncodedSystemDescriptorMemoryLayout) VectorEntrySize() memory.Range {
	return l.VectorTable.SubRange(8, 4)
}

// Returns the range within the encoded descriptor where the number of peripherals is stored
func (l *EncodedSystemDescriptorMemoryLayout) NumPeripherals() memory.Range {
	return l.Peripherals.SubRange(0, 4)
}

// Returns the number of bytes used by each peripheral entry in the encoded descriptor
func (l *EncodedSystemDescriptorMemoryLayout) PeripheralEntrySize() uint32 {
	// Name(32) + Type(16) + BaseAddress(4) + Size(4) + InterruptVector(1) + Padding(3) = 60 bytes
	return 60
}

// Returns the range within the encoded descriptor where a peripheral entry is stored given its index
func (l *EncodedSystemDescriptorMemoryLayout) PeripheralEntry(index int) memory.Range {
	entrySize := l.PeripheralEntrySize()
	start := l.Peripherals.Start + 4 + uint32(index)*entrySize
	return memory.Range{Start: start, Size: entrySize}
}

// Returns the range within the encoded descriptor where the peripheral name is stored
func (l *EncodedSystemDescriptorMemoryLayout) PeripheralName(index int) memory.Range {
	return l.PeripheralEntry(index).SubRange(0, 32)
}

// Returns the range within the encoded descriptor where the peripheral type is stored
func (l *EncodedSystemDescriptorMemoryLayout) PeripheralType(index int) memory.Range {
	return l.PeripheralEntry(index).SubRange(32, 16)
}

// Returns the range within the encoded descriptor where the peripheral base address is stored
func (l *EncodedSystemDescriptorMemoryLayout) PeripheralBaseAddress(index int) memory.Range {
	return l.PeripheralEntry(index).SubRange(48, 4)
}

// Returns the range within the encoded descriptor where the peripheral size is stored
func (l *EncodedSystemDescriptorMemoryLayout) PeripheralSize(index int) memory.Range {
	return l.PeripheralEntry(index).SubRange(52, 4)
}

// Returns the range within the encoded descriptor where the peripheral interrupt vector is stored
func (l *EncodedSystemDescriptorMemoryLayout) PeripheralInterruptVector(index int) memory.Range {
	return l.PeripheralEntry(index).SubRange(56, 1)
}

// Returns the layout of an encoded system descriptor given the number of peripherals and a custom base address
func EncodedSystemDescriptorLayout_CustomBaseAddress(numPeripherals int, baseAddress uint32) *EncodedSystemDescriptorMemoryLayout {
	return &EncodedSystemDescriptorMemoryLayout{
		Header:                memory.Range{Start: baseAddress + 0, Size: 12},
		VectorTable:           memory.Range{Start: baseAddress + 12, Size: 16},
		CodeMemoryRange:       memory.Range{Start: baseAddress + 28, Size: 12},
		DataMemoryRange:       memory.Range{Start: baseAddress + 40, Size: 12},
		HeapMemoryRange:       memory.Range{Start: baseAddress + 52, Size: 12},
		StackMemoryRange:      memory.Range{Start: baseAddress + 64, Size: 12},
		PeripheralMemoryRange: memory.Range{Start: baseAddress + 76, Size: 12},
		Peripherals:           memory.Range{Start: baseAddress + 88, Size: uint32(4 + numPeripherals*60)},
	}
}

// Returns the layout of an encoded system descriptor given the number of peripherals
func EncodedSystemDescriptorLayout(numPeripherals int) *EncodedSystemDescriptorMemoryLayout {
	return EncodedSystemDescriptorLayout_CustomBaseAddress(numPeripherals, 0)
}

// The magic number that identifies a valid system descriptor.
const RuntimeDescriptorMagic = 0xCCCACACA

// The current version of the descriptor format.
const RuntimeDescriptorVersion = 1

// Encodes a SystemDescriptor into its binary format
//
// The binary format is as follows:
// - Header:
//   - Magic (4 bytes, uint32)
//   - Version (4 bytes, uint32)
//   - TotalMemory (4 bytes, uint32)
//
// - Memory Regions (7 regions, each with BaseAddress (4 bytes, uint32), Size (4 bytes, uint32), Flags (4 bytes, uint32)):
//   - SystemDescriptor
//   - VectorTable
//   - Code
//   - Data
//   - Heap
//   - Stack
//   - Peripherals
//
// - VectorTable:
//   - BaseAddress (4 bytes, uint32)
//   - NumVectors (4 bytes, uint32)
//   - EntrySize (4 bytes, uint32)
//
// - Number of Peripherals (4 bytes, uint32)
// - Peripherals (for each peripheral):
//   - Name (32 bytes, null-padded string)
//   - Type (16 bytes, null-padded string)
//   - BaseAddress (4 bytes, uint32)
//   - Size (4 bytes, uint32)
//   - InterruptVector (1 byte, uint8) + 3 bytes padding
//
// This binary reprsentation is used by the cucaracha emulator to inject system
// information into the runtime by writing it into memory at a predefined location.
// This allows programs running inside the emulator to query system details at runtime.
// It is also used by debuggers and other tools to inspect the system configuration.
func EncodeSystemDescriptor(descriptor *SystemDescriptor, writer io.Writer) error {
	// Header
	if err := binary.Write(writer, binary.LittleEndian, RuntimeDescriptorMagic); err != nil {
		return fmt.Errorf("failed to write magic: %w", err)
	}
	if err := binary.Write(writer, binary.LittleEndian, RuntimeDescriptorVersion); err != nil {
		return fmt.Errorf("failed to write version: %w", err)
	}
	if err := binary.Write(writer, binary.LittleEndian, descriptor.MemoryLayout.TotalSize); err != nil {
		return fmt.Errorf("failed to write total size: %w", err)
	}

	// Regions (each region is BaseAddress + Size + Flags)
	encodeRegion := func(r memory.Range) error {
		if err := binary.Write(writer, binary.LittleEndian, r.Start); err != nil {
			return fmt.Errorf("failed to write region start: %w", err)
		}
		if err := binary.Write(writer, binary.LittleEndian, r.Size); err != nil {
			return fmt.Errorf("failed to write region size: %w", err)
		}
		if err := binary.Write(writer, binary.LittleEndian, uint32(r.Flags)); err != nil {
			return fmt.Errorf("failed to write region flags: %w", err)
		}
		return nil
	}

	if err := encodeRegion(descriptor.MemoryLayout.SystemDescriptor()); err != nil {
		return fmt.Errorf("failed to encode system description memory region: %w", err)
	}
	if err := encodeRegion(descriptor.MemoryLayout.VectorTable()); err != nil {
		return fmt.Errorf("failed to encode vector table memory region: %w", err)
	}
	if err := encodeRegion(descriptor.MemoryLayout.Code()); err != nil {
		return fmt.Errorf("failed to encode code memory region: %w", err)
	}
	if err := encodeRegion(descriptor.MemoryLayout.Data()); err != nil {
		return fmt.Errorf("failed to encode data memory region: %w", err)
	}
	if err := encodeRegion(descriptor.MemoryLayout.Heap()); err != nil {
		return fmt.Errorf("failed to encode heap memory region: %w", err)
	}
	if err := encodeRegion(descriptor.MemoryLayout.Stack()); err != nil {
		return fmt.Errorf("failed to encode stack memory region: %w", err)
	}
	if err := encodeRegion(descriptor.MemoryLayout.Peripherals()); err != nil {
		return fmt.Errorf("failed to encode peripherals memory region: %w", err)
	}

	// VectorTable
	if err := binary.Write(writer, binary.LittleEndian, descriptor.MemoryLayout.VectorTableBase); err != nil {
		return fmt.Errorf("failed to write vector table base address: %w", err)
	}
	if err := binary.Write(writer, binary.LittleEndian, descriptor.VectorTable.NumberOfVectors); err != nil {
		return fmt.Errorf("failed to write number of vectors: %w", err)
	}
	if err := binary.Write(writer, binary.LittleEndian, descriptor.VectorTable.VectorEntrySize); err != nil {
		return fmt.Errorf("failed to write vector entry size: %w", err)
	}

	// Number of peripherals
	if err := binary.Write(writer, binary.LittleEndian, uint32(len(descriptor.Peripherals))); err != nil {
		return fmt.Errorf("failed to write number of peripherals: %w", err)
	}

	// Peripherals
	for _, p := range descriptor.Peripherals {
		// Name: 32 bytes (null-padded)
		nameBytes := make([]byte, 32)
		copy(nameBytes, []byte(p.Metadata().Name))
		if _, err := writer.Write(nameBytes); err != nil {
			return fmt.Errorf("failed to write peripheral name: %w", err)
		}

		// Type: 16 bytes (null-padded)
		typeBytes := p.Metadata().Descriptor.Type.Encode()
		if _, err := writer.Write(typeBytes); err != nil {
			return fmt.Errorf("failed to write peripheral type: %w", err)
		}

		// BaseAddress: 4 bytes
		if err := binary.Write(writer, binary.LittleEndian, p.Metadata().BaseAddress); err != nil {
			return fmt.Errorf("failed to write peripheral base address: %w", err)
		}

		// Size: 4 bytes
		if err := binary.Write(writer, binary.LittleEndian, p.Metadata().Size); err != nil {
			return fmt.Errorf("failed to write peripheral size: %w", err)
		}

		// InterruptVector: 1 byte + 3 bytes padding
		if err := binary.Write(writer, binary.LittleEndian, p.Metadata().InterruptVector); err != nil {
			return fmt.Errorf("failed to write peripheral interrupt vector: %w", err)
		}
		var padding [3]byte
		if err := binary.Write(writer, binary.LittleEndian, padding); err != nil {
			return fmt.Errorf("failed to write peripheral interrupt vector padding: %w", err)
		}
	}

	return nil
}

// Decodes a SystemDescriptor dumped in its binary format
//
// For more details on the binary format, see EncodeSystemDescriptor
func DecodeSystemDescriptor(reader io.Reader) (*SystemDescriptor, error) {
	descriptor := &SystemDescriptor{}

	// Header
	var magic uint32
	if err := binary.Read(reader, binary.LittleEndian, &magic); err != nil {
		return nil, fmt.Errorf("failed to read magic: %w", err)
	}
	if magic != RuntimeDescriptorMagic {
		return nil, fmt.Errorf("invalid magic: expected 0x%08X, got 0x%08X", RuntimeDescriptorMagic, magic)
	}
	var version uint32
	if err := binary.Read(reader, binary.LittleEndian, &version); err != nil {
		return nil, fmt.Errorf("failed to read version: %w", err)
	}
	if version != RuntimeDescriptorVersion {
		return nil, fmt.Errorf("unsupported version: expected %d, got %d", RuntimeDescriptorVersion, version)
	}
	if err := binary.Read(reader, binary.LittleEndian, &descriptor.MemoryLayout.TotalSize); err != nil {
		return nil, fmt.Errorf("failed to read total size: %w", err)
	}

	// Regions
	decodeRegion := func() (memory.Range, error) {
		var start uint32
		if err := binary.Read(reader, binary.LittleEndian, &start); err != nil {
			return memory.Range{}, fmt.Errorf("failed to read region start: %w", err)
		}
		var size uint32
		if err := binary.Read(reader, binary.LittleEndian, &size); err != nil {
			return memory.Range{}, fmt.Errorf("failed to read region size: %w", err)
		}
		var flags memory.Flags
		if err := binary.Read(reader, binary.LittleEndian, &flags); err != nil {
			return memory.Range{}, fmt.Errorf("failed to read region flags: %w", err)
		}

		return memory.Range{
			Start: start,
			Size:  size,
			Flags: flags,
		}, nil
	}

	systemDescriptionMemoryRange, err := decodeRegion()
	if err != nil {
		return nil, fmt.Errorf("failed to decode system description memory region: %w", err)
	}
	vectorTableMemoryRange, err := decodeRegion()
	if err != nil {
		return nil, fmt.Errorf("failed to decode vector table memory region: %w", err)
	}
	codeMemoryRange, err := decodeRegion()
	if err != nil {
		return nil, fmt.Errorf("failed to decode code memory region: %w", err)
	}
	dataMemoryRange, err := decodeRegion()
	if err != nil {
		return nil, fmt.Errorf("failed to decode data memory region: %w", err)
	}
	heapMemoryRange, err := decodeRegion()
	if err != nil {
		return nil, fmt.Errorf("failed to decode heap memory region: %w", err)
	}
	stackMemoryRange, err := decodeRegion()
	if err != nil {
		return nil, fmt.Errorf("failed to decode stack memory region: %w", err)
	}
	peripheralsMemoryRange, err := decodeRegion()
	if err != nil {
		return nil, fmt.Errorf("failed to decode peripherals memory region: %w", err)
	}

	descriptor.MemoryLayout.SystemDescriptorBase = systemDescriptionMemoryRange.Start
	descriptor.MemoryLayout.SystemDescriptorSize = systemDescriptionMemoryRange.Size

	descriptor.MemoryLayout.VectorTableBase = vectorTableMemoryRange.Start
	descriptor.MemoryLayout.VectorTableSize = vectorTableMemoryRange.Size

	descriptor.MemoryLayout.CodeBase = codeMemoryRange.Start
	descriptor.MemoryLayout.CodeSize = codeMemoryRange.Size

	descriptor.MemoryLayout.DataBase = dataMemoryRange.Start
	descriptor.MemoryLayout.DataSize = dataMemoryRange.Size

	descriptor.MemoryLayout.HeapBase = heapMemoryRange.Start
	descriptor.MemoryLayout.HeapSize = heapMemoryRange.Size

	descriptor.MemoryLayout.StackBase = stackMemoryRange.Start
	descriptor.MemoryLayout.StackSize = stackMemoryRange.Size

	descriptor.MemoryLayout.PeripheralBase = peripheralsMemoryRange.Start
	descriptor.MemoryLayout.PeripheralSize = peripheralsMemoryRange.Size

	// VectorTable
	if err := binary.Read(reader, binary.LittleEndian, &descriptor.MemoryLayout.VectorTableBase); err != nil {
		return nil, fmt.Errorf("failed to read vector table base address: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &descriptor.VectorTable.NumberOfVectors); err != nil {
		return nil, fmt.Errorf("failed to read number of vectors: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &descriptor.VectorTable.VectorEntrySize); err != nil {
		return nil, fmt.Errorf("failed to read vector entry size: %w", err)
	}

	if err := descriptor.MemoryLayout.Validate(); err != nil {
		return nil, fmt.Errorf("invalid memory layout: %w", err)
	}

	// Number of peripherals
	var numPeripherals uint32
	if err := binary.Read(reader, binary.LittleEndian, &numPeripherals); err != nil {
		return nil, fmt.Errorf("failed to read number of peripherals: %w", err)
	}

	// Peripherals
	descriptor.Peripherals = make([]peripheral.Peripheral, numPeripherals)
	for i := uint32(0); i < numPeripherals; i++ {
		// Name: 32 bytes (null-terminated)
		var nameBytes [32]byte
		if err := binary.Read(reader, binary.LittleEndian, &nameBytes); err != nil {
			return nil, fmt.Errorf("failed to read peripheral name: %w", err)
		}
		name := strings.TrimRight(string(nameBytes[:]), "\x00")

		// Type: 16 bytes (null-terminated)
		var typeBytes [16]byte
		if err := binary.Read(reader, binary.LittleEndian, &typeBytes); err != nil {
			return nil, fmt.Errorf("failed to read peripheral type: %w", err)
		}
		typeName := strings.TrimRight(string(typeBytes[:]), "\x00")

		// BaseAddress: 4 bytes
		var baseAddress uint32
		if err := binary.Read(reader, binary.LittleEndian, &baseAddress); err != nil {
			return nil, fmt.Errorf("failed to read peripheral base address: %w", err)
		}

		// Size: 4 bytes
		var size uint32
		if err := binary.Read(reader, binary.LittleEndian, &size); err != nil {
			return nil, fmt.Errorf("failed to read peripheral size: %w", err)
		}

		// InterruptVector: 1 byte + 3 bytes padding
		var interruptVector uint8
		if err := binary.Read(reader, binary.LittleEndian, &interruptVector); err != nil {
			return nil, fmt.Errorf("failed to read peripheral interrupt vector: %w", err)
		}
		// Skip 3 bytes padding
		var padding [3]byte
		if err := binary.Read(reader, binary.LittleEndian, &padding); err != nil {
			return nil, fmt.Errorf("failed to read peripheral interrupt vector padding: %w", err)
		}

		peripheralDescriptor, err := peripherals.Descriptor.GetByTypeName(typeName)
		if err != nil {
			return nil, fmt.Errorf("failed to get peripheral descriptor for type '%q': %w", typeName, err)
		}

		peripheral, err := peripheralDescriptor.Factory(peripheral.PeripheralParams{
			Name:            name,
			BaseAddress:     baseAddress,
			Size:            size,
			InterruptVector: interruptVector,
			Instance:        0,   // TODO: Support multiple instances
			Description:     "",  // TODO: Support description
			Extra:           nil, // TODO: Support extra params
		})

		if err != nil {
			return nil, fmt.Errorf("failed to create %s peripheral '%s': %w", typeName, name, err)
		}

		descriptor.Peripherals[i] = peripheral
	}

	return descriptor, nil
}
