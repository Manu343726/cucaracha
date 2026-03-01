//go:generate go run generator.go -out commands.go -api api.go -docs-out docs.gob

package debugger

import (
	"bytes"
	_ "embed"
	"encoding/gob"
	"encoding/json"
	"fmt"

	"github.com/Manu343726/cucaracha/pkg/docs"
)

//go:embed docs.gob
var documentationData []byte

// CommandDocumentation holds the structured documentation for all debugger commands, generated from the source code
var Documentation *docs.DocumentationIndex = func() *docs.DocumentationIndex {
	var doc docs.DocumentationIndex
	if err := gob.NewDecoder(bytes.NewReader(documentationData)).Decode(&doc); err != nil {
		panic(fmt.Sprintf("failed to decode documentation: %v", err))
	}
	return &doc
}()

// Controls the behavior of the Step command
type StepMode int

const (
	// Step one source line (steps into function calls)
	StepModeInto StepMode = iota
	// Step one source line, stepping over function calls
	StepModeOver
	// Step out of the current function
	StepModeOut
)

func (s StepMode) String() string {
	switch s {
	case StepModeInto:
		return "into"
	case StepModeOver:
		return "over"
	case StepModeOut:
		return "out"
	default:
		return "unknown"
	}
}

func StepModeFromString(s string) (StepMode, error) {
	switch s {
	case "into":
		return StepModeInto, nil
	case "over":
		return StepModeOver, nil
	case "out":
		return StepModeOut, nil
	default:
		return 0, fmt.Errorf("unknown StepMode: \"%s\"", s)
	}
}

func (s StepMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *StepMode) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	val, err := StepModeFromString(str)
	if err != nil {
		return err
	}
	*s = val
	return nil
}

type StepCountMode int

const (
	// Count by instructions
	StepCountInstructions StepCountMode = iota
	// Count by source lines
	StepCountSourceLines
)

func (s StepCountMode) String() string {
	switch s {
	case StepCountInstructions:
		return "instructions"
	case StepCountSourceLines:
		return "sourceLines"
	default:
		return "unknown"
	}
}

func StepCountModeFromString(s string) (StepCountMode, error) {
	switch s {
	case "instructions":
		return StepCountInstructions, nil
	case "sourceLines":
		return StepCountSourceLines, nil
	default:
		return 0, fmt.Errorf("unknown StepCountMode: \"%s\"", s)
	}
}

func (s StepCountMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *StepCountMode) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	val, err := StepCountModeFromString(str)
	if err != nil {
		return err
	}
	*s = val
	return nil
}

// CommandBasedDebugger is an interface to interact with the debugger in the UI using command structures.
// It differs from the high-level Debugger interface by processing commands through a command/result pattern
// instead of direct method calls.
type CommandBasedDebugger interface {
	// Sends a command to the debugger and returns the result
	Execute(cmd *DebuggerCommand) (*DebuggerCommandResult, error)

	// Sets a callback to receive debugger events
	SetEventCallback(callback DebuggerEventCallback)
}

type commandBasedDebuggerAdapter struct {
	debugger Debugger
}

// Returns a CommandBasedDebugger that wraps a regular Debugger implementation
func MakeCommandBased(debugger Debugger) CommandBasedDebugger {
	return &commandBasedDebuggerAdapter{debugger: debugger}
}

func (d *commandBasedDebuggerAdapter) SetEventCallback(callback DebuggerEventCallback) {
	d.debugger.SetEventCallback(callback)
}
