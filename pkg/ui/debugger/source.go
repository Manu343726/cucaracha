package debugger

// SourceLocation describes a location in source code by file and line number.
type SourceLocation struct {
	// File path or name of the source file.
	File string `json:"file"`
	// Line number in the source file (1-indexed).
	Line int `json:"line"`
}

// SourceRange describes a contiguous range of source code lines.
type SourceRange struct {
	// Starting location of the range. See [SourceLocation] for structure.
	Start *SourceLocation `json:"start"`
	// Number of lines in this range starting from Start.
	Lines int `json:"lines"`
}

// SourceCodeSnippet represents a contiguous snippet of source code with detailed line information.
type SourceCodeSnippet struct {
	// Range of source lines contained in this snippet. See [SourceRange] for structure.
	SourceRange *SourceRange `json:"sourceRange"`
	// Individual source lines with metadata. See [SourceLine] for structure.
	Lines []*SourceLine `json:"lines"`
}

// SourceLine represents an individual line of source code with associated breakpoints and instruction mapping.
type SourceLine struct {
	// Source code location of this line. See [SourceLocation] for structure.
	Location *SourceLocation `json:"location"`
	// Text content of this line of source code.
	Text string `json:"text"`
	// Whether this line is the current execution location (PC).
	IsCurrent bool `json:"isCurrent"`
	// Code breakpoints set on this line.
	Breakpoints []*Breakpoint `json:"breakpoints"`
	// Data watchpoints that may trigger on this line.
	Watchpoints []*Watchpoint `json:"watchpoints"`
	// Memory address of the first instruction generated from this line (if available).
	Address uint32 `json:"address"`
}
