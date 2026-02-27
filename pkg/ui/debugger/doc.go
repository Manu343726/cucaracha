package debugger

import (
	"fmt"
	"io"

	"github.com/Manu343726/cucaracha/pkg/utils/contract"
)

// ============================================================================
// Documentation System for Debugger Commands
// ============================================================================
//
// This documentation system provides a template-based approach to command
// documentation that is resolved at runtime based on the active CommandsSyntax.
//
// Documentation is parsed from Go doc comments of:
// - DebuggerCommands interface methods (command docs)
// - Command argument types and their fields (argument docs)
// - Command result types and their fields (result docs)
//
// The documentation can reference command names, argument names, and values
// using placeholder syntax. At runtime, these placeholders are resolved to
// the appropriate syntax-specific format (REPL, JSON-RPC, etc.).
//
// Example:
//   Go doc comment: "Executes step command with {StepMode} argument"
//   Resolved in REPL: "Executes step command with step-mode argument"
//   Resolved in JSON-RPC: "Executes step command with 'stepMode' argument"

// DocumentationPlaceholder represents a placeholder in documentation text
// that should be replaced with syntax-specific formatting.
type DocumentationPlaceholder struct {
	// The raw placeholder text, e.g. "{StepMode}" or "{StepArgs.CountMode}"
	RawText string
	// Type of placeholder: "command", "argument", "field"
	Type string
	// The referenced entity name, e.g. "StepMode" or "CountMode"
	EntityName string
	// Optional parent for nested references, e.g. "StepArgs" for "StepArgs.CountMode"
	Parent string
	// Start position in the original text
	StartPos int
	// End position in the original text
	EndPos int
}

// CommandDocumentation contains documentation for a single debugger command
type CommandDocumentation struct {
	// Command identifier and name
	CommandID   DebuggerCommandId
	CommandName string // e.g., "step", "break", etc.
	MethodName  string // e.g., "Step", "Break", etc.

	// Main command documentation (from interface method doc comment)
	Summary string

	// Detailed description (optional, from interface method doc comment)
	Description string

	// Placeholders found in documentation
	Placeholders []*DocumentationPlaceholder

	// Argument documentation, keyed by argument struct field name
	// e.g., "StepArgs.StepMode" -> ArgumentDocumentation
	Arguments map[string]*ArgumentDocumentation

	// Result documentation, keyed by result struct field name
	// e.g., "ExecutionResult.State" -> ArgumentDocumentation
	Results map[string]*ArgumentDocumentation

	// Examples of how to use the command in different scenarios
	Examples []CommandExample
}

// ArgumentDocumentation contains documentation for a command argument or result field
type ArgumentDocumentation struct {
	// Field name in the argument/result struct
	FieldName string

	// Type name as it appears in code
	TypeName string

	// Whether this argument is required (for arguments, not results)
	IsRequired bool

	// Documentation text from Go doc comment
	Summary string

	// Detailed description (optional)
	Description string

	// Placeholders found in documentation
	Placeholders []*DocumentationPlaceholder

	// For enum types, the valid values with their documentation
	EnumValues map[string]string
}

// CommandExample demonstrates how to use a command
type CommandExample struct {
	// Brief description of what the example does
	Description string

	// The command invocation (will be formatted using the syntax)
	Command string

	// Expected output or behavior description
	Output string
}

// Documentation is the compiled documentation for all debugger commands.
// It is generated from Go doc comments of the DebuggerCommands interface
// and related types. It can be rendered using any CommandsSyntax.
type Documentation map[DebuggerCommandId]*CommandDocumentation

// DocsRenderer renders documentation using a specific syntax.
// It takes an output writer and a formatter (syntax) as constructor arguments.
type DocsRenderer struct {
	contract.Base
	writer    io.Writer
	formatter CommandsFormatter
}

// NewDocsRenderer creates a new documentation renderer.
func NewDocsRenderer(writer io.Writer, formatter CommandsFormatter) *DocsRenderer {
	return &DocsRenderer{
		Base:      contract.NewBase(log().Child("DocsRenderer")),
		writer:    writer,
		formatter: formatter,
	}
}

// RenderName renders the command name to the output writer.
func (r *DocsRenderer) RenderName(cmd DebuggerCommandId, docs Documentation) error {
	if docs == nil {
		return fmt.Errorf("documentation is nil")
	}

	cmdDoc, ok := docs[cmd]
	if !ok {
		return fmt.Errorf("no documentation found for command %v", cmd)
	}

	r.Log().Debug("rendering command name", "command", cmd, "name", cmdDoc.CommandName)

	if _, err := fmt.Fprintf(r.writer, "%s", r.formatter.FormatCommandName(cmd)); err != nil {
		return fmt.Errorf("failed to write command name: %w", err)
	}

	return nil
}

// RenderSummary renders the command summary to the output writer.
func (r *DocsRenderer) RenderSummary(cmd DebuggerCommandId, docs Documentation) error {
	if docs == nil {
		return fmt.Errorf("documentation is nil")
	}

	cmdDoc, ok := docs[cmd]
	if !ok {
		return fmt.Errorf("no documentation found for command %v", cmd)
	}

	r.Log().Debug("rendering command summary", "command", cmd)

	summary := cmdDoc.Summary
	if summary == "" {
		summary = fmt.Sprintf("%s command", cmdDoc.CommandName)
	}

	// Resolve placeholders in summary
	resolved, err := r.resolvePlaceholders(summary, cmdDoc.Placeholders)
	if err != nil {
		return fmt.Errorf("failed to resolve placeholders in summary: %w", err)
	}

	if _, err := fmt.Fprint(r.writer, resolved); err != nil {
		return fmt.Errorf("failed to write summary: %w", err)
	}

	return nil
}

// RenderDescription renders the command description to the output writer.
func (r *DocsRenderer) RenderDescription(cmd DebuggerCommandId, docs Documentation) error {
	if docs == nil {
		return fmt.Errorf("documentation is nil")
	}

	cmdDoc, ok := docs[cmd]
	if !ok {
		return fmt.Errorf("no documentation found for command %v", cmd)
	}

	r.Log().Debug("rendering command description", "command", cmd)

	if cmdDoc.Description == "" {
		return nil
	}

	// Resolve placeholders in description
	resolved, err := r.resolvePlaceholders(cmdDoc.Description, cmdDoc.Placeholders)
	if err != nil {
		return fmt.Errorf("failed to resolve placeholders in description: %w", err)
	}

	if _, err := fmt.Fprint(r.writer, resolved); err != nil {
		return fmt.Errorf("failed to write description: %w", err)
	}

	return nil
}

// RenderArguments renders the command arguments documentation to the output writer.
func (r *DocsRenderer) RenderArguments(cmd DebuggerCommandId, docs Documentation) error {
	if docs == nil {
		return fmt.Errorf("documentation is nil")
	}

	cmdDoc, ok := docs[cmd]
	if !ok {
		return fmt.Errorf("no documentation found for command %v", cmd)
	}

	if len(cmdDoc.Arguments) == 0 {
		r.Log().Debug("no arguments to render", "command", cmd)
		return nil
	}

	r.Log().Debug("rendering command arguments", "command", cmd, "count", len(cmdDoc.Arguments))

	for name, argDoc := range cmdDoc.Arguments {
		r.Log().Debug("rendering argument", "argument", name, "required", argDoc.IsRequired)

		formatted := r.formatter.FormatArgumentName(name, argDoc.IsRequired)
		if _, err := fmt.Fprintf(r.writer, "  %s: %s\n", formatted, argDoc.Summary); err != nil {
			return fmt.Errorf("failed to write argument documentation: %w", err)
		}
	}

	return nil
}

// RenderResults renders the command results documentation to the output writer.
func (r *DocsRenderer) RenderResults(cmd DebuggerCommandId, docs Documentation) error {
	if docs == nil {
		return fmt.Errorf("documentation is nil")
	}

	cmdDoc, ok := docs[cmd]
	if !ok {
		return fmt.Errorf("no documentation found for command %v", cmd)
	}

	if len(cmdDoc.Results) == 0 {
		r.Log().Debug("no results to render", "command", cmd)
		return nil
	}

	r.Log().Debug("rendering command results", "command", cmd, "count", len(cmdDoc.Results))

	for name, resultDoc := range cmdDoc.Results {
		r.Log().Debug("rendering result", "result", name)

		if _, err := fmt.Fprintf(r.writer, "  %s: %s\n", name, resultDoc.Summary); err != nil {
			return fmt.Errorf("failed to write result documentation: %w", err)
		}
	}

	return nil
}

// RenderExamples renders the command examples to the output writer.
func (r *DocsRenderer) RenderExamples(cmd DebuggerCommandId, docs Documentation) error {
	if docs == nil {
		return fmt.Errorf("documentation is nil")
	}

	cmdDoc, ok := docs[cmd]
	if !ok {
		return fmt.Errorf("no documentation found for command %v", cmd)
	}

	if len(cmdDoc.Examples) == 0 {
		r.Log().Debug("no examples to render", "command", cmd)
		return nil
	}

	r.Log().Debug("rendering command examples", "command", cmd, "count", len(cmdDoc.Examples))

	for i, example := range cmdDoc.Examples {
		r.Log().Debug("rendering example", "index", i)

		if _, err := fmt.Fprintf(r.writer, "Example: %s\n", example.Description); err != nil {
			return fmt.Errorf("failed to write example description: %w", err)
		}

		if _, err := fmt.Fprintf(r.writer, "  %s\n", example.Command); err != nil {
			return fmt.Errorf("failed to write example command: %w", err)
		}

		if example.Output != "" {
			if _, err := fmt.Fprintf(r.writer, "  Output: %s\n", example.Output); err != nil {
				return fmt.Errorf("failed to write example output: %w", err)
			}
		}
	}

	return nil
}

// ============================================================================
// Private Helper Methods
// ============================================================================

// resolvePlaceholders replaces all documentation placeholders with
// syntax-specific formatting using the configured formatter.
func (r *DocsRenderer) resolvePlaceholders(text string, placeholders []*DocumentationPlaceholder) (string, error) {
	if len(placeholders) == 0 {
		return text, nil
	}

	result := text
	for _, ph := range placeholders {
		r.Log().Debug("resolving placeholder", "placeholder", ph.RawText, "type", ph.Type, "entity", ph.EntityName)

		replacement := r.formatPlaceholder(ph)
		result = r.replacePlaceholder(result, ph.RawText, replacement)
	}

	return result, nil
}

// formatPlaceholder converts a single placeholder to its formatted representation
func (r *DocsRenderer) formatPlaceholder(ph *DocumentationPlaceholder) string {
	switch ph.Type {
	case "argument":
		return r.formatter.FormatArgumentName(ph.EntityName, true)
	case "field":
		return r.formatter.FormatArgumentName(ph.EntityName, false)
	default:
		r.Log().Debug("unknown placeholder type", "type", ph.Type, "text", ph.RawText)
		return ph.RawText
	}
}

// replacePlaceholder replaces a single placeholder in text
func (r *DocsRenderer) replacePlaceholder(text, oldStr, newStr string) string {
	// Simple string replacement - can be optimized if needed
	for {
		idx := findSubstring(text, oldStr)
		if idx == -1 {
			break
		}
		text = text[:idx] + newStr + text[idx+len(oldStr):]
	}
	return text
}

// findSubstring finds the first occurrence of substr in s
func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
