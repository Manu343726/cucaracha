package ui

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWatchpointTypeString(t *testing.T) {
	tests := []struct {
		wt       WatchpointType
		expected string
	}{
		{WatchpointTypeRead, "read"},
		{WatchpointTypeWrite, "write"},
		{WatchpointTypeReadWrite, "readWrite"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.wt.String()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestWatchpointTypeFromString(t *testing.T) {
	tests := []struct {
		str      string
		expected WatchpointType
		wantErr  bool
	}{
		{"read", WatchpointTypeRead, false},
		{"write", WatchpointTypeWrite, false},
		{"readWrite", WatchpointTypeReadWrite, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			got, err := WatchpointTypeFromString(tt.str)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestWatchpointTypeJSON(t *testing.T) {
	tests := []WatchpointType{
		WatchpointTypeRead,
		WatchpointTypeWrite,
		WatchpointTypeReadWrite,
	}

	for _, wt := range tests {
		t.Run(wt.String(), func(t *testing.T) {
			data, err := json.Marshal(wt)
			assert.NoError(t, err)

			var unmarshaled WatchpointType
			err = json.Unmarshal(data, &unmarshaled)
			assert.NoError(t, err)

			assert.Equal(t, wt, unmarshaled)
		})
	}
}

func TestWatchpointJSON(t *testing.T) {
	wp := &Watchpoint{
		ID: 1,
		Range: &MemoryRegion{
			Name:       "stack",
			Start:      0x2000,
			Size:       0x1000,
			RegionType: RegionStack,
		},
		Type:    WatchpointTypeWrite,
		Enabled: true,
	}

	data, err := json.Marshal(wp)
	assert.NoError(t, err)

	var unmarshaled Watchpoint
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, wp.ID, unmarshaled.ID)
	assert.Equal(t, wp.Type, unmarshaled.Type)
}

func TestMemoryRegionTypeString(t *testing.T) {
	tests := []struct {
		rt       MemoryRegionType
		expected string
	}{
		{RegionUnknown, "unknown"},
		{RegionCode, "code"},
		{RegionData, "data"},
		{RegionStack, "stack"},
		{RegionHeap, "heap"},
		{RegionIO, "io"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.rt.String()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestMemoryRegionTypeFromString(t *testing.T) {
	tests := []struct {
		str      string
		expected MemoryRegionType
		wantErr  bool
	}{
		{"unknown", RegionUnknown, false},
		{"code", RegionCode, false},
		{"data", RegionData, false},
		{"stack", RegionStack, false},
		{"heap", RegionHeap, false},
		{"io", RegionIO, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			got, err := MemoryRegionTypeFromString(tt.str)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestMemoryRegionTypeJSON(t *testing.T) {
	tests := []MemoryRegionType{
		RegionUnknown,
		RegionCode,
		RegionData,
		RegionStack,
		RegionHeap,
		RegionIO,
	}

	for _, rt := range tests {
		t.Run(rt.String(), func(t *testing.T) {
			data, err := json.Marshal(rt)
			assert.NoError(t, err)

			var unmarshaled MemoryRegionType
			err = json.Unmarshal(data, &unmarshaled)
			assert.NoError(t, err)

			assert.Equal(t, rt, unmarshaled)
		})
	}
}

func TestMemoryRegionJSON(t *testing.T) {
	region := &MemoryRegion{
		Name:       "code",
		Start:      0x0000,
		Size:       0x1000,
		RegionType: RegionCode,
	}

	data, err := json.Marshal(region)
	assert.NoError(t, err)

	var unmarshaled MemoryRegion
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, region.Name, unmarshaled.Name)
	assert.Equal(t, region.RegionType, unmarshaled.RegionType)
}

func TestMemoryRegionEnd(t *testing.T) {
	region := &MemoryRegion{
		Start: 0x1000,
		Size:  0x1000,
	}

	expected := uint32(0x2000)
	got := region.End()
	assert.Equal(t, expected, got)
}

func TestMemoryArgsJSON(t *testing.T) {
	args := &MemoryArgs{
		AddressExpr: "0x1000",
		Count:       16,
	}

	data, err := json.Marshal(args)
	assert.NoError(t, err)

	var unmarshaled MemoryArgs
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, args.AddressExpr, unmarshaled.AddressExpr)
	assert.Equal(t, args.Count, unmarshaled.Count)
}

func TestMemoryResultJSON(t *testing.T) {
	result := &MemoryResult{
		Address: 0x1000,
		Data:    []byte{0x12, 0x34, 0x56, 0x78},
		Regions: []*MemoryRegion{
			{Name: "code", Start: 0x0000, Size: 0x1000, RegionType: RegionCode},
		},
	}

	data, err := json.Marshal(result)
	assert.NoError(t, err)

	var unmarshaled MemoryResult
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, result.Address, unmarshaled.Address)
	assert.Len(t, unmarshaled.Data, len(result.Data))
}
