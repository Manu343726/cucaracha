package debugger

import (
	"encoding/json"
	"fmt"
)

// WatchpointType specifies what kind of memory access triggers a watchpoint.
type WatchpointType int

const (
	// WatchpointTypeRead triggers when memory at the watched address is read.
	WatchpointTypeRead WatchpointType = iota + 1
	// WatchpointTypeWrite triggers when memory at the watched address is written.
	WatchpointTypeWrite
	// WatchpointTypeReadWrite triggers on both reads and writes to the watched address.
	WatchpointTypeReadWrite
)

func (t WatchpointType) String() string {
	switch t {
	case WatchpointTypeRead:
		return "read"
	case WatchpointTypeWrite:
		return "write"
	case WatchpointTypeReadWrite:
		return "readWrite"
	default:
		return "unknown"
	}
}

func WatchpointTypeFromString(s string) (WatchpointType, error) {
	switch s {
	case "read":
		return WatchpointTypeRead, nil
	case "write":
		return WatchpointTypeWrite, nil
	case "readWrite":
		return WatchpointTypeReadWrite, nil
	default:
		return 0, fmt.Errorf("unknown WatchpointType: \"%s\"", s)
	}
}

func (t WatchpointType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

func (t *WatchpointType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	val, err := WatchpointTypeFromString(s)
	if err != nil {
		return err
	}
	*t = val
	return nil
}

// Watchpoint represents a memory breakpoint that triggers on read/write access to a memory range.
type Watchpoint struct {
	// Unique identifier for this watchpoint.
	ID int `json:"id"`
	// Memory range being watched. See [MemoryRegion] for details.
	Range *MemoryRegion `json:"range"`
	// Type of memory access that triggers this watchpoint. See [WatchpointType] for options.
	Type WatchpointType `json:"type"`
	// Whether this watchpoint is currently active and monitoring memory access.
	Enabled bool `json:"enabled"`
}

// MemoryRegion represents a contiguous block of memory with a specific purpose.
type MemoryRegion struct {
	// Human-readable name for this region (e.g., "stack", "heap", ".text").
	Name string `json:"name"`
	// Starting address of this memory region.
	Start uint32 `json:"start"`
	// Size in bytes of this memory region.
	Size uint32 `json:"size"`
	// Classification of this region's purpose. See [MemoryRegionType] for options.
	RegionType MemoryRegionType `json:"regionType"`
}

func (r *MemoryRegion) End() uint32 {
	return r.Start + r.Size
}

// MemoryRegionType classifies a memory region according to its function and purpose.
type MemoryRegionType int

const (
	// RegionUnknown indicates the purpose of this region is not known.
	RegionUnknown MemoryRegionType = iota
	// RegionCode contains executable program instructions.
	RegionCode
	// RegionData contains initialized program data.
	RegionData
	// RegionStack contains the automatic storage / call stack.
	RegionStack
	// RegionHeap contains dynamically allocated memory.
	RegionHeap
	// RegionIO contains memory-mapped I/O registers and peripheral memory.
	RegionIO
)

func (m MemoryRegionType) String() string {
	switch m {
	case RegionUnknown:
		return "unknown"
	case RegionCode:
		return "code"
	case RegionData:
		return "data"
	case RegionStack:
		return "stack"
	case RegionHeap:
		return "heap"
	case RegionIO:
		return "io"
	default:
		return "unknown"
	}
}

func MemoryRegionTypeFromString(s string) (MemoryRegionType, error) {
	switch s {
	case "unknown":
		return RegionUnknown, nil
	case "code":
		return RegionCode, nil
	case "data":
		return RegionData, nil
	case "stack":
		return RegionStack, nil
	case "heap":
		return RegionHeap, nil
	case "io":
		return RegionIO, nil
	default:
		return 0, fmt.Errorf("unknown MemoryRegionType: \"%s\"", s)
	}
}

func (m MemoryRegionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

func (m *MemoryRegionType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	val, err := MemoryRegionTypeFromString(s)
	if err != nil {
		return err
	}
	*m = val
	return nil
}

// MemoryArgs specifies parameters for the Memory command to display memory contents.
type MemoryArgs struct {
	// Starting address as an eval expression (e.g., "0x1000", "sp", "heap_base", "sp-16").
	AddressExpr string `json:"addressExpr"`
	// Number of bytes to display as an eval expression (optional).
	// If nil, displays a default number of bytes (typically 16 or 32).
	CountExpr *string `json:"countExpr"`
}

// MemoryResult contains the result of a Memory command displaying memory contents.
type MemoryResult struct {
	// Error message, if the memory read failed (e.g., address out of bounds).
	Error error `json:"error"`
	// Starting address where memory data begins.
	Address uint32 `json:"address"`
	// The byte contents of memory starting at Address.
	Data []byte `json:"data"`
	// Active memory regions in the system for context. See [MemoryRegion] for structure.
	Regions []*MemoryRegion `json:"regions"`
}
