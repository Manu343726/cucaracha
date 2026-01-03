// Package interpreter provides an automatic interpreter for Cucaracha machine code.
//
// # Formatting - Output Formatting Utilities
//
// This file provides utilities for formatting execution output, including
// instruction colorization and trace formatting. These utilities are used
// by both the CLI and can be used by other tools that want consistent output.
package interpreter

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
)

// FormatStyle controls the output style for formatting functions
type FormatStyle int

const (
	// StylePlain produces plain text output without colors
	StylePlain FormatStyle = iota
	// StyleColored produces colorized output using ANSI escape codes
	StyleColored
)

// OutputConfig configures output formatting
type OutputConfig struct {
	// Style controls whether output is colorized
	Style FormatStyle
	// Writer is where output is written (default: os.Stderr)
	Writer io.Writer
}

// InstructionFormatter formats instructions for display
type InstructionFormatter struct {
	config OutputConfig
}

// NewInstructionFormatter creates a new instruction formatter
func NewInstructionFormatter(config OutputConfig) *InstructionFormatter {
	return &InstructionFormatter{config: config}
}

// Regular expressions for parsing instruction parts
var (
	regPattern    = regexp.MustCompile(`\b(r[0-9]{1,2}|sp|lr|pc|cpsr)\b`)
	immPattern    = regexp.MustCompile(`#-?[0-9]+|#-?0x[0-9a-fA-F]+|\b-?[0-9]+\b`)
	opcodePattern = regexp.MustCompile(`^[A-Z][A-Z0-9]+`)
)

// FormatInstruction formats an instruction string
func (f *InstructionFormatter) FormatInstruction(instr string) string {
	if f.config.Style == StylePlain {
		return instr
	}
	return colorizeInstruction(instr)
}

// colorizeInstruction applies ANSI colors to instruction parts
func colorizeInstruction(instr string) string {
	instr = strings.TrimSpace(instr)
	if instr == "" {
		return instr
	}

	// Find and extract the opcode first
	opcodeLoc := opcodePattern.FindStringIndex(instr)
	if opcodeLoc == nil {
		return instr
	}

	opcode := instr[opcodeLoc[0]:opcodeLoc[1]]
	rest := instr[opcodeLoc[1]:]

	// ANSI color codes
	const (
		reset  = "\033[0m"
		yellow = "\033[1;33m" // Bold yellow for opcode
		green  = "\033[32m"   // Green for registers
		cyan   = "\033[36m"   // Cyan for immediates
		white  = "\033[37m"   // White for punctuation
	)

	// Build the result with colored opcode
	result := yellow + opcode + reset

	// Find all matches
	regMatches := regPattern.FindAllStringIndex(rest, -1)
	immMatches := immPattern.FindAllStringIndex(rest, -1)

	// Create spans for coloring
	type colorSpan struct {
		start int
		end   int
		color string
		text  string
	}
	var spans []colorSpan

	for _, m := range regMatches {
		spans = append(spans, colorSpan{m[0], m[1], green, rest[m[0]:m[1]]})
	}
	for _, m := range immMatches {
		// Check if this immediate overlaps with a register
		overlaps := false
		for _, rm := range regMatches {
			if m[0] < rm[1] && m[1] > rm[0] {
				overlaps = true
				break
			}
		}
		if !overlaps {
			spans = append(spans, colorSpan{m[0], m[1], cyan, rest[m[0]:m[1]]})
		}
	}

	// Sort spans by start position
	for i := 0; i < len(spans); i++ {
		for j := i + 1; j < len(spans); j++ {
			if spans[j].start < spans[i].start {
				spans[i], spans[j] = spans[j], spans[i]
			}
		}
	}

	// Build the rest of the string with colors
	pos := 0
	for _, span := range spans {
		if span.start > pos {
			result += white + rest[pos:span.start] + reset
		}
		result += span.color + span.text + reset
		pos = span.end
	}
	if pos < len(rest) {
		result += white + rest[pos:] + reset
	}

	return result
}

// TraceFormatter formats execution trace output
type TraceFormatter struct {
	config    OutputConfig
	formatter *InstructionFormatter
}

// NewTraceFormatter creates a new trace formatter
func NewTraceFormatter(config OutputConfig) *TraceFormatter {
	return &TraceFormatter{
		config:    config,
		formatter: NewInstructionFormatter(config),
	}
}

// FormatStep formats a single execution step for trace output
func (t *TraceFormatter) FormatStep(step int, pc uint32, instrText string, state *CPUState) string {
	if t.config.Style == StylePlain {
		return fmt.Sprintf("[%4d] PC=0x%04X sp=%6d lr=%6d r0=%6d | %s",
			step, pc, *state.SP, *state.LR, state.Registers[16], instrText)
	}

	// Colored output
	const (
		reset   = "\033[0m"
		gray    = "\033[90m"   // Bright black for step
		cyan    = "\033[36m"   // Cyan for PC
		magenta = "\033[35m"   // Magenta for word
		green   = "\033[32m"   // Green for register names
		white   = "\033[1;37m" // Bold white for values
	)

	return fmt.Sprintf("%s[%s%4d%s] %sPC%s=%s0x%04X%s %ssp%s=%s%6d%s %slr%s=%s%6d%s %sr0%s=%s%6d%s | %s",
		reset,
		gray, step, reset,
		green, reset, cyan, pc, reset,
		green, reset, white, *state.SP, reset,
		green, reset, white, *state.LR, reset,
		green, reset, white, state.Registers[16], reset,
		t.formatter.FormatInstruction(instrText))
}

// FormatSourceLocation formats a source location for display
func (t *TraceFormatter) FormatSourceLocation(loc *mc.SourceLocation, srcLine string) string {
	if loc == nil || !loc.IsValid() {
		return ""
	}

	srcStr := ""
	if srcLine != "" {
		srcStr = strings.TrimSpace(srcLine)
	}

	if t.config.Style == StylePlain {
		return fmt.Sprintf("  %s:%d  %s", loc.File, loc.Line, srcStr)
	}

	// Colored output
	const (
		reset     = "\033[0m"
		blue      = "\033[94m"   // Bright blue for file
		cyan      = "\033[96m"   // Bright cyan for line number
		boldWhite = "\033[1;37m" // Bold white for source
	)

	return fmt.Sprintf("  %s%s%s:%s%d%s  %s%s%s",
		blue, loc.File, reset,
		cyan, loc.Line, reset,
		boldWhite, srcStr, reset)
}

// ExecutionSummary contains summary information about an execution
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
func (t *TraceFormatter) FormatSummary(summary *ExecutionSummary, verbose bool) string {
	var sb strings.Builder

	if verbose {
		if summary.ExitedNormally {
			sb.WriteString("\n=== Execution completed (returned from main) ===\n")
		} else {
			sb.WriteString(fmt.Sprintf("\n=== Execution %s ===\n", summary.StopReason.String()))
		}
		sb.WriteString(fmt.Sprintf("Steps executed: %d\n", summary.StepsExecuted))
		sb.WriteString(fmt.Sprintf("Final PC: 0x%08X\n", summary.FinalPC))
	}

	if verbose {
		sb.WriteString(fmt.Sprintf("\nReturn value (r0): %d\n", summary.ReturnValue))
	} else {
		sb.WriteString(fmt.Sprintf("%d\n", summary.ReturnValue))
	}

	return sb.String()
}
