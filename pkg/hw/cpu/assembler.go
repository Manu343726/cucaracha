package cpu

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// AssemblyFile represents the parsed contents of a .cucaracha assembly file
// It contains sections, global symbols, and a list of instructions per function
// For now, we focus on the .text section and function bodies

type AssemblyFile struct {
	FileName  string
	Globals   []string
	Functions map[string]*FunctionBody
}

type FunctionBody struct {
	Name         string
	Instructions []string // Raw instruction lines
}

// ParseAssemblyFile parses a .cucaracha assembly file and returns an in-memory representation
func ParseAssemblyFile(path string) (*AssemblyFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	file := &AssemblyFile{
		FileName:  path,
		Globals:   []string{},
		Functions: map[string]*FunctionBody{},
	}

	var currentFunc *FunctionBody
	funcHeader := regexp.MustCompile(`^([A-Za-z0-9_\.]+):`) // e.g. main:
	globalHeader := regexp.MustCompile(`^\.globl\s+(\S+)`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // skip empty and comment lines
		}

		if funcHeader.MatchString(line) {
			name := funcHeader.FindStringSubmatch(line)[1]
			currentFunc = &FunctionBody{Name: name, Instructions: []string{}}
			file.Functions[name] = currentFunc
			continue
		}

		if globalHeader.MatchString(line) {
			name := globalHeader.FindStringSubmatch(line)[1]
			file.Globals = append(file.Globals, name)
			continue
		}

		if currentFunc != nil {
			currentFunc.Instructions = append(currentFunc.Instructions, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return file, nil
}
