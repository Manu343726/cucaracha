package cpu

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/exp/constraints"
)

type Interpreter interface {
	Run(instruction string, args ...string) (*string, error)
}

type CommandInterpreter interface {
	Run(command string) (*string, error)
}

type ProgramInterpreter interface {
	Run(commands []string) (*string, error)
}

type commandInterpreter struct {
	impl Interpreter
}

func MakeCommandInterpreter(i Interpreter) CommandInterpreter {
	return &commandInterpreter{
		impl: i,
	}
}

func (i *commandInterpreter) Run(command string) (*string, error) {
	// ignore comments:
	if strings.HasPrefix(command, "//") {
		return nil, nil
	} else if strings.Contains(command, "//") {
		command = strings.Split(command, "//")[0]
	}

	args := strings.Fields(command)

	if len(args) <= 0 {
		return nil, MakeInterpreterError(ErrBadParameters, "invalid command, cannot be empty")
	} else {
		return i.impl.Run(args[0], args[1:]...)
	}
}

type sanitizedCommandInterpreter struct {
	CommandInterpreter
}

func (i *sanitizedCommandInterpreter) Run(command string) (*string, error) {
	command = strings.TrimSpace(command)

	// ignore empty lines:
	if len(command) <= 0 {
		return nil, nil
	}

	return i.CommandInterpreter.Run(command)
}

func MakeSanitizedCommandInterpreter(i CommandInterpreter) CommandInterpreter {
	return &sanitizedCommandInterpreter{
		CommandInterpreter: i,
	}
}

type programInterpreter struct {
	impl CommandInterpreter
}

func MakeProgramInterpreter(i CommandInterpreter) ProgramInterpreter {
	return &programInterpreter{
		impl: i,
	}
}

func MakeProgramError(line int, command string, err error) error {
	return fmt.Errorf("error at line %v (%v): %w", line, command, err)
}

func (i *programInterpreter) Run(commands []string) (*string, error) {
	var lastResult *string

	for line, command := range commands {
		if result, err := i.impl.Run(command); err != nil {
			return nil, MakeProgramError(line, command, err)
		} else if result != nil {
			lastResult = result
		}
	}

	return lastResult, nil
}

type microCpuInterpreter[Register RegisterName, Word constraints.Integer, Float constraints.Float] struct {
	MicroCpu[Register, Word, Float]
	registerParser RegisterParser[Register]
}

func MakeMicroCpuInterpreter[Register RegisterName, Word constraints.Integer, Float constraints.Float](impl MicroCpu[Register, Word, Float], registerParser RegisterParser[Register]) Interpreter {
	return &microCpuInterpreter[Register, Word, Float]{
		MicroCpu:       impl,
		registerParser: registerParser,
	}
}

func MakeInterpreterError(err error, args ...any) error {
	if len(args) <= 0 {
		return fmt.Errorf("%w: %w", ErrInterpreter, err)
	} else {
		switch message := args[0].(type) {
		case string:
			return fmt.Errorf("%w: %w: "+message, append([]any{ErrInterpreter, err}, args[1:]...)...)
		default:
			return fmt.Errorf("%w: %w: "+fmt.Sprint(message), append([]any{ErrInterpreter, err}, args[1:]...)...)
		}
	}
}

var (
	ErrInterpreter    = errors.New("interpreter error")
	ErrBadParameters  = errors.New("bad paramters")
	ErrBadInstruction = errors.New("bad instruction")
)

const (
	MicroCpuInstruction_ReadWord  string = "RW"
	MicroCpuInstruction_WriteWord string = "WW"
)

func (i *microCpuInterpreter[Register, Word, Float]) readWord(args ...string) (*string, error) {
	if len(args) > 1 {
		return nil, MakeInterpreterError(ErrBadParameters, "expected one register argument, got %v arguments", len(args))
	}

	if r, err := i.registerParser(args[0]); err != nil {
		return nil, MakeInterpreterError(ErrBadParameters, "could not parse register argument '%v': %w", args[0], err)
	} else {
		value, err := i.AllWordRegisters().Read(r)
		strValue := fmt.Sprint(value)
		return &strValue, err
	}
}

func parseWord[Word constraints.Integer](str string) (Word, error) {
	if value, err := strconv.ParseInt(str, 10, Sizeof[Word]()*8); err != nil {
		return 0, MakeInterpreterError(ErrBadParameters, "%w", err)
	} else {
		return Word(value), nil
	}
}

func (i *microCpuInterpreter[Register, Word, Float]) writeWord(args ...string) error {
	if len(args) > 2 {
		return MakeInterpreterError(ErrBadParameters, "expected one word argument and one register argument, got %v arguments", len(args))
	}

	value, err := parseWord[Word](args[0])
	if err != nil {
		return err
	}

	if r, err := i.registerParser(args[1]); err != nil {
		return MakeInterpreterError(ErrBadParameters, "could not parse register argument '%v': %w", args[1], err)
	} else {
		return i.AllWordRegisters().Write(value, r)
	}
}

func (i *microCpuInterpreter[Register, Word, Float]) Run(instruction string, args ...string) (*string, error) {
	switch instruction {
	case MicroCpuInstruction_ReadWord:
		return i.readWord(args...)
	case MicroCpuInstruction_WriteWord:
		return nil, i.writeWord(args...)
	}

	return nil, MakeInterpreterError(ErrBadInstruction, "unsupported instruction '%v'", instruction)
}
