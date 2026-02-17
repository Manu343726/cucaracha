package ui

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSourceLocationJSON(t *testing.T) {
	loc := &SourceLocation{
		File: "main.c",
		Line: 42,
	}

	data, err := json.Marshal(loc)
	assert.NoError(t, err)

	var unmarshaled SourceLocation
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, loc.File, unmarshaled.File)
	assert.Equal(t, loc.Line, unmarshaled.Line)
}

func TestSourceRangeJSON(t *testing.T) {
	sr := &SourceRange{
		Start: &SourceLocation{File: "test.c", Line: 10},
		Lines: 5,
	}

	data, err := json.Marshal(sr)
	assert.NoError(t, err)

	var unmarshaled SourceRange
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, sr.Lines, unmarshaled.Lines)
}

func TestSourceCodeSnippetJSON(t *testing.T) {
	snippet := &SourceCodeSnippet{
		SourceRange: &SourceRange{
			Start: &SourceLocation{File: "app.c", Line: 1},
			Lines: 3,
		},
		Lines: []*SourceLine{
			{
				Location:  &SourceLocation{File: "app.c", Line: 1},
				Text:      "int main() {",
				IsCurrent: true,
				Address:   0x1000,
			},
		},
	}

	data, err := json.Marshal(snippet)
	assert.NoError(t, err)

	var unmarshaled SourceCodeSnippet
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, snippet.SourceRange.Lines, unmarshaled.SourceRange.Lines)
}

func TestSourceLineJSON(t *testing.T) {
	sl := &SourceLine{
		Location:    &SourceLocation{File: "test.c", Line: 10},
		Text:        "return 0;",
		IsCurrent:   true,
		Address:     0x2000,
		Breakpoints: []*Breakpoint{},
		Watchpoints: []*Watchpoint{},
	}

	data, err := json.Marshal(sl)
	assert.NoError(t, err)

	var unmarshaled SourceLine
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, sl.Text, unmarshaled.Text)
	assert.Equal(t, sl.IsCurrent, unmarshaled.IsCurrent)
}
