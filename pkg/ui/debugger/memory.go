package debugger

import (
	"encoding/json"
	"fmt"
)

// Represents the action that triggers a watchpoint
type WatchpointType int

const (
	WatchpointTypeRead WatchpointType = iota + 1
	WatchpointTypeWrite
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

// Represents a memory watchpoint
type Watchpoint struct {
	ID      int            `json:"id"`      // Watchpoint ID
	Range   *MemoryRegion  `json:"range"`   // Memory range
	Type    WatchpointType `json:"type"`    // Watchpoint type
	Enabled bool           `json:"enabled"` // Whether the watchpoint is enabled
}

// Represents a memory region for display
type MemoryRegion struct {
	Name       string           `json:"name"`       // Region name
	Start      uint32           `json:"start"`      // Start address
	Size       uint32           `json:"size"`       // Region size
	RegionType MemoryRegionType `json:"regionType"` // Region type
}

func (r *MemoryRegion) End() uint32 {
	return r.Start + r.Size
}

// Classifies memory regions depending on their purpose
type MemoryRegionType int

const (
	RegionUnknown MemoryRegionType = iota
	RegionCode
	RegionData
	RegionStack
	RegionHeap
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

// Memory command arguments
type MemoryArgs struct {
	AddressExpr string `json:"addressExpr"` // Address expression
	Count       int    `json:"count"`       // Number of bytes to display (optional)
}

// Result of Memory command
type MemoryResult struct {
	Error   error           `json:"error"`   // Error, if any
	Address uint32          `json:"address"` // Start address
	Data    []byte          `json:"data"`    // Memory data
	Regions []*MemoryRegion `json:"regions"` // Memory regions
}
