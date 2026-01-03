package mc

// Source File Loading
//
// This file provides utilities for loading original source files into the debug info.
// When debugging, it's useful to display the actual source code lines alongside
// the machine instructions. This requires reading the source files referenced
// in the DWARF debug information.
//
// The loader handles cases where:
//   - Source files may be at absolute paths from the compilation machine
//   - Source files may need to be found relative to the current directory
//   - Source files may not be available (silently ignored)

import (
	"bufio"
	"os"
	"path/filepath"
)

// loadSourceFileImpl loads a source file's contents into the debug info.
// It reads all lines from the file and stores them in DebugInfo.SourceFiles.
// If the file cannot be opened at its original path, it tries the basename only.
func loadSourceFileImpl(d *DebugInfo, file string) error {
	// Try to open the file
	f, err := os.Open(file)
	if err != nil {
		// Try relative path variations
		baseName := filepath.Base(file)
		f, err = os.Open(baseName)
		if err != nil {
			return err
		}
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	d.SourceFiles[file] = lines
	return nil
}

// TryLoadSourceFiles attempts to load all referenced source files
// Errors are silently ignored since source files may not be available
func (d *DebugInfo) TryLoadSourceFiles() {
	if d == nil {
		return
	}

	// Collect unique file paths from instruction locations
	files := make(map[string]bool)
	for _, loc := range d.InstructionLocations {
		if loc != nil && loc.File != "" {
			files[loc.File] = true
		}
	}

	// Try to load each file
	for file := range files {
		_ = d.LoadSourceFile(file)
	}
}
