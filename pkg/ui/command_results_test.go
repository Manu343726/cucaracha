package ui

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBreakResultJSON(t *testing.T) {
	result := &BreakResult{
		Breakpoint: &Breakpoint{
			ID: 1,
			Location: &SourceLocation{
				File: "main.go",
				Line: 42,
			},
			Enabled: true,
		},
	}

	data, err := json.Marshal(result)
	assert.NoError(t, err)

	var unmarshaled BreakResult
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, result.Breakpoint.ID, unmarshaled.Breakpoint.ID)
}

func TestWatchResultJSON(t *testing.T) {
	result := &WatchResult{
		Watchpoint: &Watchpoint{
			ID: 2,
			Range: &MemoryRegion{
				Name:       "stack",
				Start:      0x2000,
				Size:       0x100,
				RegionType: RegionStack,
			},
			Type:    WatchpointTypeWrite,
			Enabled: true,
		},
	}

	data, err := json.Marshal(result)
	assert.NoError(t, err)

	var unmarshaled WatchResult
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, result.Watchpoint.ID, unmarshaled.Watchpoint.ID)
}

func TestRemoveBreakpointResultJSON(t *testing.T) {
	result := &RemoveBreakpointResult{
		Breakpoint: &Breakpoint{
			ID: 3,
			Location: &SourceLocation{
				File: "test.go",
				Line: 50,
			},
			Enabled: false,
		},
	}

	data, err := json.Marshal(result)
	assert.NoError(t, err)

	var unmarshaled RemoveBreakpointResult
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, result.Breakpoint.ID, unmarshaled.Breakpoint.ID)
}

func TestRemoveWatchpointResultJSON(t *testing.T) {
	result := &RemoveWatchpointResult{
		Watchpoint: &Watchpoint{
			ID: 4,
			Range: &MemoryRegion{
				Name:       "heap",
				Start:      0x3000,
				Size:       0x200,
				RegionType: RegionHeap,
			},
			Type:    WatchpointTypeRead,
			Enabled: false,
		},
	}

	data, err := json.Marshal(result)
	assert.NoError(t, err)

	var unmarshaled RemoveWatchpointResult
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, result.Watchpoint.ID, unmarshaled.Watchpoint.ID)
}

func TestSourceResultJSON(t *testing.T) {
	result := &SourceResult{
		Snippet: &SourceCodeSnippet{
			SourceRange: &SourceRange{
				Start: &SourceLocation{File: "main.go", Line: 1},
				Lines: 2,
			},
			Lines: []*SourceLine{
				{Location: &SourceLocation{File: "main.go", Line: 1}, Text: "package main", IsCurrent: false},
				{Location: &SourceLocation{File: "main.go", Line: 2}, Text: "func main() {", IsCurrent: true},
			},
		},
	}

	data, err := json.Marshal(result)
	assert.NoError(t, err)

	var unmarshaled SourceResult
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, result.Snippet.SourceRange.Start.File, unmarshaled.Snippet.SourceRange.Start.File)
	assert.Len(t, unmarshaled.Snippet.Lines, len(result.Snippet.Lines))
}

func TestEvalResultJSON(t *testing.T) {
	result := &EvalResult{
		Value:       123,
		ValueString: "123",
		ValueBytes:  []byte{0x7b},
	}

	data, err := json.Marshal(result)
	assert.NoError(t, err)

	var unmarshaled EvalResult
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, result.Value, unmarshaled.Value)
	assert.Equal(t, result.ValueString, unmarshaled.ValueString)
}

func TestListResultJSON(t *testing.T) {
	result := &ListResult{
		Breakpoints: []*Breakpoint{
			{ID: 1, Location: &SourceLocation{File: "main.go", Line: 42}, Enabled: true},
		},
		Watchpoints: []*Watchpoint{
			{ID: 1, Range: &MemoryRegion{Name: "stack", Start: 0x2000, Size: 0x100, RegionType: RegionStack}, Type: WatchpointTypeWrite, Enabled: true},
		},
	}

	data, err := json.Marshal(result)
	assert.NoError(t, err)

	var unmarshaled ListResult
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Len(t, unmarshaled.Breakpoints, len(result.Breakpoints))
	assert.Len(t, unmarshaled.Watchpoints, len(result.Watchpoints))
}
