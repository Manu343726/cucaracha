package memory

import (
	"sort"

	"github.com/Manu343726/cucaracha/pkg/utils"
)

// MemoryLayout describes how the CPU's address space is organized.
// It allows configuring where code, stack, heap, and peripherals reside.
//
// Memory layout (from low to high addresses):
//   - System Descriptor: Hardware configuration info at address 0
//   - Vector Table: Interrupt handler addresses
//   - Code: Program instructions
//   - Data: Static/global variables
//   - Heap: Dynamic allocations (grows up)
//   - (free space)
//   - Stack: Call stack (grows down)
//   - Peripherals: Memory-mapped I/O
type MemoryLayout struct {
	// TotalSize is the total addressable memory in bytes.
	TotalSize uint32

	// SystemDescriptorBase is where the system descriptor starts (typically 0).
	SystemDescriptorBase uint32

	// SystemDescriptorSize is the size of the system descriptor region.
	SystemDescriptorSize uint32

	// VectorTableBase is where the interrupt vector table starts.
	VectorTableBase uint32

	// VectorTableSize is the size of the vector table region.
	VectorTableSize uint32

	// DataBase is where static/global data starts.
	DataBase uint32

	// DataSize is the size of the data region.
	DataSize uint32

	// CodeBase is where program code is loaded.
	CodeBase uint32

	// CodeSize is the size of the code region.
	CodeSize uint32

	// HeapBase is where dynamic allocations start.
	HeapBase uint32

	// HeapSize is the maximum heap size in bytes.
	HeapSize uint32

	// StackBase is the initial stack pointer value.
	// Stack grows downward, so this should point to the top of the stack region.
	StackBase uint32

	// StackSize is the maximum stack size in bytes.
	StackSize uint32

	// PeripheralBase is where memory-mapped peripherals begin.
	// Peripherals are typically placed at the end of the address space.
	PeripheralBase uint32

	// PeripheralSize is the total size reserved for peripherals.
	PeripheralSize uint32

	// Base address of each peripheral in order
	PeripheralBaseAddresses []uint32
}

const (
	// MinSystemDescriptorSize is the minimum size of the system descriptor region.
	MinSystemDescriptorSize uint32 = 512
)

// Returns a default memory layout for the given total size.
// Layout (for 64KB example):
//
//	0x00000000 - 0x000000BF: System Descriptor (192 bytes)
//	0x000000C0 - 0x0000013F: Vector Table (128 bytes, 32 vectors)
//	0x00000140 - 0x0000BFFF: Code + Data (48KB - overhead)
//	0x0000C000 - 0x0000EFFF: Stack (12KB, grows down from 0x0000EFFF)
//	0x0000F000 - 0x0000FFFF: Peripherals (4KB)
func DefaultLayout(totalSize uint32) MemoryLayout {
	return CustomLayout(LayoutOptions{
		TotalSize:            totalSize,
		SystemDescriptorSize: MinSystemDescriptorSize,
		VectorTableSize:      32 * 4,
		CodeSize:             totalSize / 2,
		DataSize:             totalSize / 8,
		HeapSize:             totalSize / 8,
		StackSize:            totalSize / 8,
		PeripheralSize:       totalSize / 16,
	})
}

type LayoutOptions struct {
	TotalSize            uint32
	SystemDescriptorSize uint32
	VectorTableSize      uint32
	CodeSize             uint32
	DataSize             uint32
	HeapSize             uint32
	StackSize            uint32
	PeripheralSize       uint32
}

// Creates a memory layout with explicit region sizes.
func CustomLayout(options LayoutOptions) MemoryLayout {
	if options.SystemDescriptorSize < MinSystemDescriptorSize {
		options.SystemDescriptorSize = MinSystemDescriptorSize
	}

	// Vector table after system descriptor
	vectorTableBase := options.SystemDescriptorSize
	vectorTableSz := options.VectorTableSize

	// Code starts after vector table
	codeBase := vectorTableBase + vectorTableSz
	dataBase := codeBase + options.CodeSize
	heapBase := dataBase + options.DataSize
	stackBase := utils.NextAligned(heapBase+options.HeapSize, 4)
	peripheralsBase := stackBase + options.StackSize

	return MemoryLayout{
		TotalSize:            options.TotalSize,
		SystemDescriptorBase: 0,
		SystemDescriptorSize: options.SystemDescriptorSize,
		VectorTableBase:      vectorTableBase,
		VectorTableSize:      vectorTableSz,
		CodeBase:             codeBase,
		CodeSize:             options.CodeSize,
		DataBase:             dataBase,
		DataSize:             options.DataSize,
		StackBase:            stackBase,
		StackSize:            options.StackSize,
		HeapBase:             heapBase,
		HeapSize:             options.HeapSize,
		PeripheralBase:       peripheralsBase,
		PeripheralSize:       options.PeripheralSize,
	}
}

// Returns the RAM memory range used for the system descriptor.
func (l MemoryLayout) SystemDescriptor() Range {
	return Range{Start: l.SystemDescriptorBase, Size: l.SystemDescriptorSize, Flags: FlagReadable}
}

// Returns the RAM memory range used for the interrupt vector table.
func (l MemoryLayout) VectorTable() Range {
	return Range{Start: l.VectorTableBase, Size: l.VectorTableSize, Flags: FlagReadable | FlagWritable}
}

// Returns the RAM memory range used for code
func (l MemoryLayout) Code() Range {
	return Range{Start: l.CodeBase, Size: l.CodeSize, Flags: FlagExecutable | FlagReadable}
}

// Returns the RAM memory range user for global data
func (l MemoryLayout) Data() Range {
	return Range{Start: l.DataBase, Size: l.DataSize, Flags: FlagReadable | FlagWritable}
}

// Returns the RAM memory range used for the heap.
func (l MemoryLayout) Heap() Range {
	return Range{Start: l.HeapBase, Size: l.HeapSize, Flags: FlagReadable | FlagWritable}
}

// Returns the RAM memory range used for the stack.
func (l MemoryLayout) Stack() Range {
	return Range{Start: l.StackBase, Size: l.StackSize, Flags: FlagReadable | FlagWritable}
}

// Returns the first address used for the stack (bottom of the stack).
// Remarks: Stack grows down, so this is the highest address of the stack region.
func (l MemoryLayout) StackBottom() uint32 {
	return l.StackBase + l.StackSize - 4
}

// Returns the last address used for the stack (top of the stack).
// Remarks: Stack grows down, so this is the lowest address of the stack region.
func (l MemoryLayout) StackTop() uint32 {
	return l.StackBase
}

// Returns the RAM memory range used for peripherals.
func (l MemoryLayout) Peripherals() Range {
	return Range{Start: l.PeripheralBase, Size: l.PeripheralSize, Flags: FlagReadable | FlagWritable}
}

// Validate checks if the layout is valid.
func (l MemoryLayout) Validate() error {
	// Check system descriptor
	if l.SystemDescriptorBase > l.VectorTableBase {
		return &LayoutError{"system descriptor is after vector table"}
	}
	if l.SystemDescriptorBase+l.SystemDescriptorSize > l.VectorTableBase {
		return &LayoutError{"system descriptor overlaps with vector table"}
	}
	if l.SystemDescriptorBase+l.SystemDescriptorSize > l.TotalSize {
		return &LayoutError{"system descriptor exceeds total size"}
	}

	// Check vector table
	if l.VectorTableBase > l.CodeBase {
		return &LayoutError{"vector table is after code"}
	}
	if l.VectorTableBase+l.VectorTableSize > l.CodeBase {
		return &LayoutError{"vector table overlaps with code"}
	}
	if l.VectorTableBase+l.VectorTableSize > l.TotalSize {
		return &LayoutError{"vector table exceeds total size"}
	}

	// Check code
	if l.CodeBase > l.DataBase {
		return &LayoutError{"code region is after data region"}
	}
	if l.CodeBase+l.CodeSize > l.DataBase {
		return &LayoutError{"code region overlaps with data region"}
	}
	if l.CodeBase+l.CodeSize > l.TotalSize {
		return &LayoutError{"code region exceeds total size"}
	}

	// Check data
	if l.DataBase > l.HeapBase {
		return &LayoutError{"data region is after heap"}
	}
	if l.DataBase+l.DataSize > l.HeapBase {
		return &LayoutError{"data region overlaps with heap"}
	}
	if l.DataBase+l.DataSize > l.TotalSize {
		return &LayoutError{"data region exceeds total size"}
	}

	// Check heap
	if l.HeapBase > l.StackBase {
		return &LayoutError{"heap is after stack"}
	}
	if l.HeapBase+l.HeapSize > l.StackBase {
		return &LayoutError{"heap overlaps with stack"}
	}
	if l.HeapBase+l.HeapSize > l.TotalSize {
		return &LayoutError{"heap exceeds total size"}
	}

	// Check stack
	if l.StackBase > l.PeripheralBase {
		return &LayoutError{"stack is after peripherals"}
	}
	if l.StackBase+l.StackSize > l.PeripheralBase {
		return &LayoutError{"stack overlaps with peripherals"}
	}
	if l.StackBase+l.StackSize > l.TotalSize {
		return &LayoutError{"stack exceeds total size"}
	}

	// Check peripherals
	if l.PeripheralBase+l.PeripheralSize > l.TotalSize {
		return &LayoutError{"peripheral region exceeds total size"}
	}

	return nil
}

// Returns all defined memory ranges in the layout.
func (l MemoryLayout) Ranges() []Range {
	return []Range{
		l.SystemDescriptor(),
		l.VectorTable(),
		l.Code(),
		l.Data(),
		l.Heap(),
		l.Stack(),
		l.Peripherals(),
	}
}

func (l MemoryLayout) HasOverlappingRanges() bool {
	return RangesOverlap(l.Ranges())
}

func (l MemoryLayout) UnusedRanges() []Range {
	var unused []Range

	ranges := l.Ranges()
	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].Start < ranges[j].Start
	})

	// Check gaps between regions
	prevEnd := uint32(0)
	for _, region := range ranges {
		if region.Start > prevEnd {
			unused = append(unused, Range{
				Start: prevEnd,
				Size:  region.Start - prevEnd,
			})
		}
		prevEnd = region.End()
	}

	// Check for unused space at the end
	if prevEnd < l.TotalSize {
		unused = append(unused, Range{
			Start: prevEnd,
			Size:  l.TotalSize - prevEnd,
		})
	}

	return unused
}

// Represents a memory layout configuration error.
type LayoutError struct {
	Message string
}

func (e *LayoutError) Error() string {
	return "invalid memory layout: " + e.Message
}
