package system

import (
	"fmt"
	"sort"

	"github.com/Manu343726/cucaracha/pkg/hw/memory"
	"github.com/Manu343726/cucaracha/pkg/hw/peripheral"
)

// MemoryRequirements describes the memory needs for different CPU components.
// The allocator uses these requirements to calculate a valid memory layout.
type MemoryRequirements struct {
	// TotalSize is the total memory size in bytes.
	// If 0, the allocator calculates a suitable size.
	TotalSize uint32

	// Number of instructions in the code region. If not set, defaults to MinCodeSize / 4.
	CodeInstructions uint32

	// DataSize is the space needed for static/global data.
	// This is placed after code.
	DataSize uint32

	// HeapSize is the space for dynamic allocations.
	// If 0, remaining space after other allocations is used.
	HeapSize uint32

	// StackSize is the space for the call stack.
	StackSize uint32

	// NumInterruptVectors is the number of interrupt vectors.
	NumInterruptVectors uint32

	// Size of each interrupt vector in bytes.
	// If 0, defaults to DefaultVectorSize (4 bytes).
	VectorSize uint32

	// Peripherals describes the memory-mapped peripherals to allocate.
	// Each peripheral has a size requirement; the allocator assigns addresses.
	Peripherals []peripheral.Metadata

	// MinPeripheralRegion is the minimum size for the peripheral region.
	// Even with no peripherals, some space may be reserved.
	// Default is 0 (no minimum).
	MinPeripheralRegion uint32
}

// AllocationResult contains the calculated memory layout and peripheral addresses.
type AllocationResult struct {
	// Layout is the computed memory layout.
	Layout memory.MemoryLayout

	// Resulting base addresses for each peripheral
	PeripheralAddresses []uint32

	// Summary provides a human-readable summary of the allocation.
	Summary string
}

// Computes the memory layout and peripheral MMIO addresses based on the system requirements.
func Allocate(req MemoryRequirements) (*AllocationResult, error) {
	// System descriptor size depends on number of peripherals
	sysDescSize := EncodedSystemDescriptorLayout(len(req.Peripherals)).Size()

	// Vector table size
	vectorTableSz := req.NumInterruptVectors * req.VectorSize

	// Ensure sizes meet minimums
	codeSize := req.CodeInstructions * 4

	// Calculate peripheral space needed
	peripheralSize := req.MinPeripheralRegion
	for _, p := range req.Peripherals {
		peripheralSize += memory.AlignSize(p.Size, peripheralAlignment(p))
	}

	// Validate we have enough space
	minRequired := sysDescSize + vectorTableSz + codeSize + req.DataSize + req.HeapSize + req.StackSize + peripheralSize

	if req.TotalSize == 0 {
		req.TotalSize = minRequired
	}

	layout := memory.CustomLayout(memory.LayoutOptions{
		TotalSize:            req.TotalSize,
		SystemDescriptorSize: sysDescSize,
		VectorTableSize:      vectorTableSz,
		CodeSize:             codeSize,
		DataSize:             req.DataSize,
		HeapSize:             req.HeapSize,
		StackSize:            req.StackSize,
		PeripheralSize:       peripheralSize,
	})

	// Validate the layout
	if err := layout.Validate(); err != nil {
		return nil, fmt.Errorf("memory layout validation failed: %w", err)
	}

	// Allocate peripheral addresses and build peripheral info
	peripheralAddrs := make([]uint32, 0, len(req.Peripherals))
	currentAddr := layout.PeripheralBase

	// Sort peripherals: preferred addresses first, then by size (largest first)
	sortedPeripherals := make([]peripheral.Metadata, len(req.Peripherals))
	copy(sortedPeripherals, req.Peripherals)
	sort.Slice(sortedPeripherals, func(i, j int) bool {
		// Then by size (largest first for better packing)
		return sortedPeripherals[i].Size > sortedPeripherals[j].Size
	})

	// Assign addresses and build peripheral info
	for _, p := range sortedPeripherals {
		alignment := peripheralAlignment(p)
		alignedAddr := memory.NextAlignedAddress(currentAddr, alignment)

		// Check if preferred address can be used
		if p.BaseAddress != 0 {
			preferredAligned := memory.NextAlignedAddress(p.BaseAddress, alignment)
			if preferredAligned >= layout.PeripheralBase && preferredAligned+p.Size <= req.TotalSize {
				// Check for conflicts with already allocated peripherals
				conflict := false
				for _, allocated := range peripheralAddrs {
					// Simple overlap check
					if preferredAligned < allocated+p.Size && preferredAligned+p.Size > allocated {
						conflict = true
						break
					}
				}
				if !conflict {
					alignedAddr = preferredAligned
				}
			}
		}

		// Ensure we don't exceed peripheral region
		if alignedAddr+p.Size > req.TotalSize {
			return nil, fmt.Errorf("peripheral %q (size %d) doesn't fit in remaining peripheral space", p.Name, p.Size)
		}

		peripheralAddrs = append(peripheralAddrs, alignedAddr)
		currentAddr = alignedAddr + p.Size
		p.BaseAddress = alignedAddr
	}

	// Build summary
	summary := fmt.Sprintf(`Memory Layout (Total: %d bytes / 0x%X):
  SysDescriptor: 0x%08X - 0x%08X (%d bytes)
  VectorTable:   0x%08X - 0x%08X (%d bytes, %d vectors)
  Code:          0x%08X - 0x%08X (%d bytes)
  Data:          0x%08X - 0x%08X (%d bytes)
  Heap:          0x%08X - 0x%08X (%d bytes)
  Stack:         0x%08X - 0x%08X (%d bytes, grows down from 0x%08X)
  Peripherals:   0x%08X - 0x%08X (%d bytes)
  Total Size:   %d bytes`,
		req.TotalSize, req.TotalSize,
		layout.SystemDescriptor().Start, layout.SystemDescriptor().End(), layout.SystemDescriptorSize,
		layout.VectorTable().Start, layout.VectorTable().End(), layout.VectorTableSize, req.NumInterruptVectors,
		layout.Code().Start, layout.Code().End(), layout.CodeSize,
		layout.Data().Start, layout.Data().End(), layout.DataSize,
		layout.Heap().Start, layout.Heap().End(), layout.HeapSize,
		layout.StackBottom(), layout.StackBase, layout.StackSize, layout.StackTop(),
		layout.Peripherals().Start, layout.Peripherals().End(), layout.PeripheralSize,
		layout.TotalSize)

	if len(peripheralAddrs) > 0 {
		summary += "\n  Peripheral Addresses:"
		for i, addr := range peripheralAddrs {
			summary += fmt.Sprintf("\n    %s: 0x%08X", sortedPeripherals[i].Name, addr)
		}
	}

	return &AllocationResult{
		Layout:              layout,
		PeripheralAddresses: peripheralAddrs,
		Summary:             summary,
	}, nil
}

func peripheralAlignment(p peripheral.Metadata) uint32 {
	if p.Descriptor != nil {
		return p.Descriptor.DefaultAlignment
	}

	// TODO: peripheral parameters could specify alignment
	return 0
}
