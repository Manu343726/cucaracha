package repl

import (
	"encoding/json"
	"fmt"
	"strings"
)

// OutputFormat specifies how the REPL outputs results
type OutputFormat int

const (
	// HumanReadable outputs human-friendly text
	HumanReadable OutputFormat = iota
	// MachineReadable outputs JSONL format (one entry per command)
	MachineReadable
)

// String returns the string representation of the output format
func (of OutputFormat) String() string {
	switch of {
	case HumanReadable:
		return "human_readable"
	case MachineReadable:
		return "machine_readable"
	default:
		return "unknown"
	}
}

// OutputFormatFromString parses an output format from a string
func OutputFormatFromString(s string) (OutputFormat, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "human_readable":
		return HumanReadable, nil
	case "machine_readable":
		return MachineReadable, nil
	default:
		return HumanReadable, fmt.Errorf("unknown output format: %s", s)
	}
}

// CommandOutput represents a single command execution in machine-readable format
type CommandOutput struct {
	Command string  `json:"command"`
	Output  string  `json:"output"`
	Success bool    `json:"success"`
	Error   string  `json:"error,omitempty"`
	File    string  `json:"file,omitempty"`
	Line    int     `json:"line,omitempty"`
	Index   int     `json:"index"`
}

// ToJSON converts the command output to JSON
func (co *CommandOutput) ToJSON() (string, error) {
	data, err := json.Marshal(co)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
