package sourcecode

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type sourceFileOnDisk struct {
	path string
	name string
}

// Returns a SourceFile backed on disk. All operations will read from the file system.
func NewSourceFileOnDisk(filePath string) (File, error) {
	if fileInfo, err := os.Stat(filePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("source file '%s' does not exist", filePath)
		}

		return nil, fmt.Errorf("failed to stat source file '%s': %w", filePath, err)
	} else if !fileInfo.IsDir() {
		return nil, fmt.Errorf("source file '%s' is a directory", filePath)
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for file '%s': %w", filePath, err)
	}

	return &sourceFileOnDisk{
		path: absPath,
		name: path.Base(absPath),
	}, nil
}

func (s *sourceFileOnDisk) Name() string {
	return s.name
}

func (s *sourceFileOnDisk) Path() string {
	return s.path
}

func (s *sourceFileOnDisk) Snippet(r *Range) (*Snippet, error) {
	reader, err := os.Open(s.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open source file '%s': %w", s.path, err)
	}

	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	var currentLine int = 1

	result := &Snippet{
		Range:    r,
		Lines:    make([]*Line, 0, r.LineCount),
		FullText: "",
	}

	for scanner.Scan() {
		if r.ContainsLine(currentLine) {
			lineText := scanner.Text()
			result.FullText += lineText + "\n"
		}

		currentLine++
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading source file '%s': %w", s.path, err)
	}

	// We compose the lines from the full text to reduce memory usage (we don't store each line
	// separately as a new string but as slices of the full text)
	for i, line := range strings.Split(result.FullText, "\n") {
		result.Lines = append(result.Lines, &Line{
			Location: &Location{
				File:      s,
				Line:      r.StartLine + i,
				Column:    0,
				EndColumn: len(line),
			},
			Text: line,
		})
	}

	return result, nil
}

type sourceLibraryOnDisk struct {
	sourceFiles map[string]File
}

// Creates a new source code library backed on disk.
// The library will load source files on demand from the file system.
func NewSourceLibraryOnDisk() Library {
	return &sourceLibraryOnDisk{
		sourceFiles: make(map[string]File),
	}
}

func (s *sourceLibraryOnDisk) File(filePath string) (File, error) {
	if sourceFile, exists := s.sourceFiles[filePath]; exists {
		return sourceFile, nil
	}

	sourceFile, err := NewSourceFileOnDisk(filePath)
	if err != nil {
		return nil, err
	}

	s.sourceFiles[filePath] = sourceFile
	return sourceFile, nil
}

func (s *sourceLibraryOnDisk) KnownFiles() []File {
	files := make([]File, 0, len(s.sourceFiles))
	for _, file := range s.sourceFiles {
		files = append(files, file)
	}
	return files
}
