// Package interpreter provides an automatic interpreter for Cucaracha machine code.
//
// # Formatting - Output Formatting Utilities
//
// This file provides utilities for formatting execution output, including
// instruction colorization and trace formatting. These utilities are used
// by both the CLI and can be used by other tools that want consistent output.
package core

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/Manu343726/cucaracha/pkg/runtime"
	"github.com/Manu343726/cucaracha/pkg/runtime/program/sourcecode"
	"github.com/Manu343726/cucaracha/pkg/utils"
)

// Defines ANSI colors to be used when formatting
type FormatPalette struct {
	Step        *color.Color
	Opcode      *color.Color
	Register    *color.Color
	Immediate   *color.Color
	Punctuation *color.Color
	Registers   []*color.Color
	SourceCode  *utils.CSyntaxHighlightPalette
}

// Returns the default instruction format palette
func DefaultFormatPalette() *FormatPalette {
	return &FormatPalette{
		Step:        color.New(color.FgHiBlack),
		Opcode:      color.New(color.FgYellow),
		Register:    color.New(color.FgGreen),
		Immediate:   color.New(color.FgCyan),
		Punctuation: color.New(color.FgWhite),
		Registers: []*color.Color{
			color.New(color.FgMagenta),
			color.New(color.FgCyan),
			color.New(color.FgGreen),
			color.New(color.FgWhite),
		},
		SourceCode: utils.DefaultCSyntaxHighlightPalette(),
	}
}

// OutputConfig configures output formatting
type OutputConfig struct {
	// String used to separate instruction operands
	OperandSeparator string
	// Palette defines the colors used for formatting different tokens. If nil, no colors are used.
	Palette *FormatPalette
}

// Returns the default output configuration
func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		OperandSeparator: ", ",
		Palette:          DefaultFormatPalette(),
	}
}

// InstructionFormatter formats instructions for display
type InstructionFormatter struct {
	config OutputConfig
	output io.Writer
}

// NewInstructionFormatter creates a new instruction formatter
func NewInstructionFormatter(config OutputConfig, output io.Writer) *InstructionFormatter {
	return &InstructionFormatter{config: config, output: output}
}

// Executes formatting of the given instruction
func (f *InstructionFormatter) Run(instr *instructions.Instruction) {
	if f.config.Palette == nil {
		f.plainFormat(instr)
	} else {
		f.colorizedFormat(instr)
	}
}

func (f *InstructionFormatter) plainFormat(instr *instructions.Instruction) {
	f.output.Write([]byte(instr.Descriptor.OpCode.Mnemonic))

	if len(instr.Descriptor.Operands) > 0 {
		f.output.Write([]byte(" "))

		for i, op := range instr.OperandValues {
			f.plainFormatInstructionOperand(op)

			if i < len(instr.OperandValues)-1 {
				f.output.Write([]byte(f.config.OperandSeparator))
			}
		}
	}
}

func (f *InstructionFormatter) plainFormatInstructionOperand(op instructions.OperandValue) {
	switch op.Kind() {
	case instructions.OperandKind_Register:
		f.output.Write([]byte(op.Register().Name()))
	case instructions.OperandKind_Immediate:
		f.output.Write([]byte(op.Immediate().String()))
	default:
		panic("unsupported operand kind")
	}
}

func (f *InstructionFormatter) colorizedFormat(instr *instructions.Instruction) {
	f.config.Palette.Opcode.Fprint(f.output, instr.Descriptor.OpCode.Mnemonic)

	if len(instr.Descriptor.Operands) > 0 {
		f.config.Palette.Punctuation.Fprint(f.output, f.config.Palette.Punctuation, "")
	}

	for i, op := range instr.OperandValues {
		f.colorizedFormatInstructionOperand(op)

		if i < len(instr.OperandValues)-1 {
			f.config.Palette.Punctuation.Fprint(f.output, f.config.Palette.Punctuation, f.config.OperandSeparator)
		}
	}
}

func (f *InstructionFormatter) colorizedFormatInstructionOperand(op instructions.OperandValue) {
	switch op.Kind() {
	case instructions.OperandKind_Register:
		f.config.Palette.Register.Fprint(f.output, op.Register().Name())
	case instructions.OperandKind_Immediate:
		f.config.Palette.Immediate.Fprint(f.output, op.Immediate().String())
	default:
		panic("unsupported operand kind")
	}
}

// TraceFormatter formats execution trace output
type TraceFormatter struct {
	config OutputConfig
	output io.Writer
}

// NewTraceFormatter creates a new trace formatter
func NewTraceFormatter(config OutputConfig, output io.Writer) *TraceFormatter {
	return &TraceFormatter{
		config: config,
		output: output,
	}
}

// FormatStep formats a single execution step for trace output
func (t *TraceFormatter) FormatStep(step int, instr *instructions.Instruction, runtime runtime.Runtime) {
	if t.config.Palette == nil {
		t.formatStepPlain(step, instr, runtime)
	} else {
		t.formatStepColored(step, instr, runtime)
	}
}

func (t *TraceFormatter) formatStepPlain(step int, instr *instructions.Instruction, runtime runtime.Runtime) {
	registers := NewRegisters(runtime)
	stateRegisters := registers.ReadStateRegisters()

	fmt.Fprintf(t.output, "[%4d] ", step)
	for name, value := range stateRegisters {
		fmt.Fprintf(t.output, "%s="+RecommendedRegisterStringFormat()[name]+" ", name, value)
	}

	NewInstructionFormatter(t.config, t.output).Run(instr)
}

func (t *TraceFormatter) formatStepColored(step int, instr *instructions.Instruction, runtime runtime.Runtime) {
	registers := NewRegisters(runtime)
	stateRegisters := registers.ReadStateRegisters()

	t.config.Palette.Step.Fprintf(t.output, "[%4d] ", step)

	for name, value := range stateRegisters {
		format := RecommendedRegisterStringFormat()[name]
		color := RecommendedRegisterColor(t.config.Palette.Registers)[name]
		color.Fprintf(t.output, "%s=", name)
		t.config.Palette.Immediate.Fprintf(t.output, format+" ", value)
	}

	NewInstructionFormatter(t.config, t.output).Run(instr)
}

// FormatSourceLocation formats a source location for display
func (t *TraceFormatter) FormatSourceLocation(loc *sourcecode.Location, srcLine string) {
	if loc == nil || !loc.IsValid() {
		return
	}

	srcLine = strings.TrimSpace(srcLine)

	if t.config.Palette == nil || t.config.Palette.SourceCode == nil {
		t.formatSourceLocationPlain(loc, srcLine)
	} else {
		t.formatSourceLocationColored(loc, srcLine)
	}
}

func (t *TraceFormatter) formatSourceLocationPlain(loc *sourcecode.Location, srcLine string) {
	if loc == nil || !loc.IsValid() {
		return
	}

	srcLine = strings.TrimSpace(srcLine)

	fmt.Fprintf(t.output, "  %s:%d  %s\n", loc.File, loc.Line, srcLine)
}

func (t *TraceFormatter) formatSourceLocationColored(loc *sourcecode.Location, srcLine string) {
	if loc == nil || !loc.IsValid() {
		return
	}

	srcLine = strings.TrimSpace(srcLine)

	fmt.Fprintf(t.output, "  %s:%d  ", loc.File, loc.Line)

	utils.HighlightCCode(t.output, srcLine, t.config.Palette.SourceCode)

	fmt.Fprintln(t.output)
}

type ExecutionSummary struct {
	// StepsExecuted is the total number of instructions executed
	StepsExecuted int
	// FinalPC is the program counter at termination
	FinalPC uint32
	// ReturnValue is the return value (r0)
	ReturnValue int32
	// ExitedNormally indicates if the program returned from main
	ExitedNormally bool
	// StopReason is the reason execution stopped
	StopReason StopReason
	// Error is any error that occurred (may be nil)
	Error error
}

// FormatSummary formats an execution summary for display
func (t *TraceFormatter) FormatSummary(summary *ExecutionSummary, verbose bool) {
	if verbose {
		if summary.ExitedNormally {
			fmt.Fprint(t.output, "\n=== Execution completed (returned from main) ===\n")
		} else {
			fmt.Fprintf(t.output, "\n=== Execution %s ===\n", summary.StopReason.String())
		}
		fmt.Fprintf(t.output, "Steps executed: %d\n", summary.StepsExecuted)
		fmt.Fprintf(t.output, "Final PC: 0x%08X\n", summary.FinalPC)
	}

	if verbose {
		fmt.Fprintf(t.output, "\nReturn value (r0): %d\n", summary.ReturnValue)
	} else {
		fmt.Fprintf(t.output, "%d\n", summary.ReturnValue)
	}
}
