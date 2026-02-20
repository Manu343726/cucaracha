package sourcecode

import (
	"fmt"
	"log/slog"
)

// Represents a location within a source file
type Location struct {
	File      File
	Line      int
	Column    int
	EndColumn int
}

func (s *Location) IsValid() bool {
	return s.File != nil && s.Line > 0
}

func (s *Location) String() string {
	if !s.IsValid() {
		return "<unknown>"
	}

	if s.Column > 0 {
		return fmt.Sprintf("%s:%d:%d", s.File.Path(), s.Line, s.Column)
	}

	return fmt.Sprintf("%s:%d", s.File.Path(), s.Line)
}

// Returns a slog attribute representing the source code location in a human-readable format.
func (s *Location) LoggingAttribute(name string) slog.Attr {
	return slog.String(name, s.String())
}

// Represents a range of lines within a source file
type Range struct {
	File      File
	StartLine int
	LineCount int
}

// Returns a Range that covers a single line at the given location
func SingleLineRange(location *Location) *Range {
	return &Range{
		File:      location.File,
		StartLine: location.Line,
		LineCount: 1,
	}
}

func (r *Range) IsValid() bool {
	return r.File != nil && r.StartLine > 0 && r.LineCount > 0
}

func (r *Range) EndLine() int {
	return r.StartLine + r.LineCount - 1
}

func (r *Range) String() string {
	if !r.IsValid() {
		return "<unknown>"
	}

	return fmt.Sprintf("%s:%d-%d", r.File.Path(), r.StartLine, r.EndLine())
}

func (r *Range) StartLocation() *Location {
	return &Location{
		File: r.File,
		Line: r.StartLine,
	}
}

func (r *Range) EndLocation() *Location {
	return &Location{
		File: r.File,
		Line: r.EndLine(),
	}
}

func (r *Range) ContainsLine(line int) bool {
	return line >= r.StartLine && line <= r.EndLine()
}

// Represents a snippet of source code within a file
type Snippet struct {
	Range    *Range
	Lines    []*Line
	FullText string
}

func (s *Snippet) IsValid() bool {
	return s.Range != nil && s.Range.IsValid()
}

func (s *Snippet) String() string {
	return s.Range.String()
}

// Represents a source code line, with text and its location within a file
type Line struct {
	Location *Location
	Text     string
}

// Represents a source code file
type File interface {
	Name() string
	Path() string
	Snippet(r *Range) (*Snippet, error)
}

// Provides access to a program source code
//
// A source code library allows retrieving source files, their contents, and searching
// for symbols and locations within the source code.
type Library interface {
	// Retrieves a source file by its path. If the file is not found, returns an error.
	File(filepath string) (File, error)
	// Lists all known source files in the library. A library implementation may choose to
	// only return files that have been loaded on demand when doing File(path) calls.
	KnownFiles() []File
}

type fileNamed string

func (f fileNamed) Name() string {
	return string(f)
}

func (f fileNamed) Path() string {
	return string(f)
}

func (f fileNamed) Snippet(r *Range) (*Snippet, error) {
	return nil, fmt.Errorf("cannot return snippet: this is not a real sourcecode.File but a placeholder for name '%s'", string(f))
}

// Returns a File instance that represents a source file with the given name.
//
// This is intended for use in situations where only the name of the file is needed,
// such as in source code locations, but the actual file contents are not required.
func FileNamed(name string) File {
	return fileNamed(name)
}

// Returns a code snippet from the given library
func ReadSnippet(lib Library, r *Range) (*Snippet, error) {
	sourceFile, err := lib.File(r.File.Path())
	if err != nil {
		return nil, fmt.Errorf("failed to get source file '%s': %w", r.File.Path(), err)
	}

	return sourceFile.Snippet(r)
}
