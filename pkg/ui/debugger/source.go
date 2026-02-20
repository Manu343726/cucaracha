package debugger

// Represents a source code location
type SourceLocation struct {
	File string `json:"file"` // File path
	Line int    `json:"line"` // Line number
}

// Represents a range of source code
type SourceRange struct {
	Start *SourceLocation `json:"start"` // Start location
	Lines int             `json:"lines"` // Number of lines
}

// Represents a snippet of source code
type SourceCodeSnippet struct {
	SourceRange *SourceRange  `json:"sourceRange"` // Source range
	Lines       []*SourceLine `json:"lines"`       // Source lines
}

// SourceLine represents a line of source code
type SourceLine struct {
	Location    *SourceLocation `json:"location"`    // Source location
	Text        string          `json:"text"`        // Text of the line
	IsCurrent   bool            `json:"isCurrent"`   // Whether this is the current line
	Breakpoints []*Breakpoint   `json:"breakpoints"` // Breakpoints on this line
	Watchpoints []*Watchpoint   `json:"watchpoints"` // Watchpoints on this line
	Address     uint32          `json:"address"`     // Address of the first instruction on this line
}
