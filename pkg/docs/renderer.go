package docs

import (
	"fmt"
	"io"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/reflect"
)

// LinkFormatter is an interface for formatting links in documentation.
// Different implementations can format links differently based on the use case
// (e.g., REPL documentation vs HTML documentation).
// LinkFormatter receives pointers to the actual reflect entities, enabling
// rich formatting based on the complete type information.
type LinkFormatter interface {
	// FormatPackage formats a reference to a package
	FormatPackage(pkg *reflect.Package) string

	// FormatType formats a reference to a type
	FormatType(typ *reflect.Type, pkg *reflect.Package) string

	// FormatFunction formats a reference to a function
	FormatFunction(fn *reflect.Function, pkg *reflect.Package) string

	// FormatMethod formats a reference to a method on a type
	FormatMethod(method *reflect.Method, typ *reflect.Type, pkg *reflect.Package) string

	// FormatConstant formats a reference to a constant
	FormatConstant(const_ *reflect.Constant, pkg *reflect.Package) string

	// FormatField formats a reference to a struct field
	FormatField(field *reflect.Field, typ *reflect.Type, pkg *reflect.Package) string

	// FormatEnumValue formats a reference to an enum value (constant in an enum)
	FormatEnumValue(const_ *reflect.Constant, enum *reflect.Enum, pkg *reflect.Package) string

	// FormatLink formats a generic link when the source entity is unknown or stored as any
	// This is the fallback for links that don't have specific type information
	FormatLink(target string, sourceEntity any) string
}

// RenderOptions contains formatting options for rendering documentation
type RenderOptions struct {
	// Indentation is the string used for indenting content
	Indentation string

	// Prefix is a string prefixed to every line of output
	Prefix string

	// LinkFormatter is used to format links in documentation text
	LinkFormatter LinkFormatter

	// ReflectIndex provides efficient typed access to reflect entities for link formatting
	// If provided, enables O(1) lookups for parent types and packages.
	// If nil, the renderer falls back to linear search (less efficient).
	ReflectIndex *reflect.Index
}

// DefaultRenderOptions returns sensible default rendering options
func DefaultRenderOptions(formatter LinkFormatter) *RenderOptions {
	return &RenderOptions{
		Indentation:   "  ",
		Prefix:        "",
		LinkFormatter: formatter,
	}
}

// EntryRenderer renders documentation entries to a writer
type EntryRenderer struct {
	writer  io.Writer
	options *RenderOptions

	// The entry being rendered
	entry *DocumentationEntry

	// Current indentation level
	indentLevel int
}

// NewEntryRenderer creates a new EntryRenderer
func NewEntryRenderer(writer io.Writer, options *RenderOptions) *EntryRenderer {
	if options == nil {
		options = &RenderOptions{
			Indentation:   "  ",
			Prefix:        "",
			LinkFormatter: &DefaultLinkFormatter{},
		}
	}
	return &EntryRenderer{
		writer:      writer,
		options:     options,
		indentLevel: 0,
	}
}

// SetEntry sets the documentation entry to render
func (r *EntryRenderer) SetEntry(entry *DocumentationEntry) {
	r.entry = entry
}

// RenderSummary renders the summary section of the documentation entry
func (r *EntryRenderer) RenderSummary() error {
	if r.entry == nil || r.entry.Summary == "" {
		return nil
	}

	if err := r.writeLine(""); err != nil {
		return err
	}

	return r.writeResolvedWrapped(r.entry.Summary)
}

// RenderDetails renders the detailed documentation section
func (r *EntryRenderer) RenderDetails() error {
	if r.entry == nil || r.entry.Details == "" {
		return nil
	}

	if err := r.writeLine(""); err != nil {
		return err
	}

	return r.writeResolvedWrapped(r.entry.Details)
}

// RenderExamples renders the examples section
func (r *EntryRenderer) RenderExamples() error {
	if r.entry == nil || len(r.entry.Examples) == 0 {
		return nil
	}

	if err := r.writeLine(""); err != nil {
		return err
	}

	if err := r.writeLine("Examples:"); err != nil {
		return err
	}

	r.indentLevel++
	defer func() { r.indentLevel-- }()

	for i, example := range r.entry.Examples {
		if i > 0 {
			if err := r.writeLine(""); err != nil {
				return err
			}
		}

		if example.Description != "" {
			if err := r.writeLine(example.Description); err != nil {
				return err
			}
		}

		// Write code block with indentation
		if err := r.writeLine(""); err != nil {
			return err
		}

		r.indentLevel++
		for _, codeLine := range strings.Split(example.Code, "\n") {
			if err := r.writeLine(codeLine); err != nil {
				return err
			}
		}
		r.indentLevel--

		if example.Output != "" {
			if err := r.writeLine(""); err != nil {
				return err
			}
			if err := r.writeLine("Output:"); err != nil {
				return err
			}
			r.indentLevel++
			for _, outLine := range strings.Split(example.Output, "\n") {
				if err := r.writeLine(outLine); err != nil {
					return err
				}
			}
			r.indentLevel--
		}
	}

	return nil
}

// RenderLinks renders cross-references to other documentation entries
func (r *EntryRenderer) RenderLinks() error {
	if r.entry == nil || len(r.entry.Links) == 0 {
		return nil
	}

	if err := r.writeLine(""); err != nil {
		return err
	}

	if err := r.writeLine("Related:"); err != nil {
		return err
	}

	r.indentLevel++
	defer func() { r.indentLevel-- }()

	for idx := range r.entry.Links {
		linkText := r.formatLink(&r.entry.Links[idx])
		relStr := ""
		if r.entry.Links[idx].Relationship != "" {
			relStr = fmt.Sprintf(" (%s)", r.entry.Links[idx].Relationship)
		}

		if err := r.writeLine(fmt.Sprintf("%s%s", linkText, relStr)); err != nil {
			return err
		}
	}

	return nil
}

// RenderFull renders the complete documentation entry
func (r *EntryRenderer) RenderFull() error {
	if r.entry == nil {
		return fmt.Errorf("no entry to render")
	}

	if err := r.RenderSummary(); err != nil {
		return err
	}
	if err := r.RenderDetails(); err != nil {
		return err
	}
	if err := r.RenderExamples(); err != nil {
		return err
	}
	if err := r.RenderLinks(); err != nil {
		return err
	}

	return nil
}

// writeResolvedText writes text with {{Links[N]}} placeholders resolved directly to writer
func (r *EntryRenderer) writeResolvedText(text string) error {
	if r.entry == nil || len(r.entry.Links) == 0 {
		// No links, just write the text directly
		_, err := io.WriteString(r.writer, text)
		return err
	}

	// Scan through text and write segments with resolved links
	remaining := text
	for i := range r.entry.Links {
		placeholder := fmt.Sprintf("{{Links[%d]}}", i)
		idx := strings.Index(remaining, placeholder)
		if idx == -1 {
			continue
		}

		// Write text before placeholder
		if _, err := io.WriteString(r.writer, remaining[:idx]); err != nil {
			return err
		}

		// Write formatted link
		formatted := r.formatLink(&r.entry.Links[i])
		if _, err := io.WriteString(r.writer, formatted); err != nil {
			return err
		}

		// Move past the placeholder
		remaining = remaining[idx+len(placeholder):]
	}

	// Write any remaining text
	_, err := io.WriteString(r.writer, remaining)
	return err
}

// formatLink uses the LinkFormatter to format a single link based on its source entity type
func (r *EntryRenderer) formatLink(link *Link) string {
	if link == nil {
		return ""
	}
	if r.options.LinkFormatter == nil || link.Source == nil {
		return link.Target
	}

	source := link.Source
	pkg := link.SourcePackage

	// Switch on the non-nil field in the DocumentationSource
	switch {
	case source.Package != nil:
		return r.options.LinkFormatter.FormatPackage(source.Package)

	case source.Type != nil:
		return r.options.LinkFormatter.FormatType(source.Type, pkg)

	case source.Function != nil:
		return r.options.LinkFormatter.FormatFunction(source.Function, pkg)

	case source.Method != nil:
		parentType := r.options.ReflectIndex.MethodParent(source.Method)
		if parentType != nil {
			return r.options.LinkFormatter.FormatMethod(source.Method, parentType, pkg)
		}
		return r.options.LinkFormatter.FormatLink(link.Target, source)

	case source.Constant != nil:
		enum := r.options.ReflectIndex.ConstantEnum(source.Constant)
		if enum != nil {
			return r.options.LinkFormatter.FormatEnumValue(source.Constant, enum, pkg)
		}
		return r.options.LinkFormatter.FormatConstant(source.Constant, pkg)

	case source.Field != nil:
		parentType := r.options.ReflectIndex.FieldParent(source.Field)
		if parentType != nil {
			return r.options.LinkFormatter.FormatField(source.Field, parentType, pkg)
		}
		return r.options.LinkFormatter.FormatLink(link.Target, source)

	case source.Enum != nil:
		// For enum types, format as a type
		if source.Enum.Type != nil && source.Enum.Type.Type != nil {
			return r.options.LinkFormatter.FormatType(source.Enum.Type.Type, pkg)
		}
		return r.options.LinkFormatter.FormatLink(link.Target, source)

	default:
		return r.options.LinkFormatter.FormatLink(link.Target, source)
	}
}

// writeResolvedWrapped writes wrapped text with link resolution directly to writer
func (r *EntryRenderer) writeResolvedWrapped(text string) error {
	// Split by double newlines to preserve paragraph breaks
	paragraphs := strings.Split(text, "\n\n")

	for i, para := range paragraphs {
		if i > 0 {
			if err := r.writeLine(""); err != nil {
				return err
			}
		}

		// Split by single newlines within paragraph and write each line
		lines := strings.Split(para, "\n")
		for _, line := range lines {
			indent := r.options.Prefix + strings.Repeat(r.options.Indentation, r.indentLevel)
			if _, err := fmt.Fprintf(r.writer, "%s", indent); err != nil {
				return err
			}

			if err := r.writeResolvedText(strings.TrimSpace(line)); err != nil {
				return err
			}

			if _, err := io.WriteString(r.writer, "\n"); err != nil {
				return err
			}
		}
	}

	return nil
}

// writeLine writes a line of text with proper indentation and prefix
func (r *EntryRenderer) writeLine(text string) error {
	indent := r.options.Prefix + strings.Repeat(r.options.Indentation, r.indentLevel)
	_, err := fmt.Fprintf(r.writer, "%s%s\n", indent, text)
	return err
}

// writeWrapped writes wrapped text, preserving paragraph breaks
func (r *EntryRenderer) writeWrapped(text string) error {
	// Split by double newlines to preserve paragraph breaks
	paragraphs := strings.Split(text, "\n\n")

	for i, para := range paragraphs {
		if i > 0 {
			if err := r.writeLine(""); err != nil {
				return err
			}
		}

		// Split by single newlines within paragraph
		lines := strings.Split(para, "\n")
		for _, line := range lines {
			if err := r.writeLine(strings.TrimSpace(line)); err != nil {
				return err
			}
		}
	}

	return nil
}

// DefaultLinkFormatter is a basic implementation of LinkFormatter
type DefaultLinkFormatter struct{}

// FormatLink formats a generic link target
func (f *DefaultLinkFormatter) FormatLink(target string, sourceEntity any) string {
	return fmt.Sprintf("[%s]", target)
}

// FormatPackage formats a reference to a package
func (f *DefaultLinkFormatter) FormatPackage(pkg *reflect.Package) string {
	return fmt.Sprintf("[%s]", pkg.Name)
}

// FormatType formats a reference to a type
func (f *DefaultLinkFormatter) FormatType(typ *reflect.Type, pkg *reflect.Package) string {
	return fmt.Sprintf("[%s]", typ.Name)
}

// FormatFunction formats a reference to a function
func (f *DefaultLinkFormatter) FormatFunction(fn *reflect.Function, pkg *reflect.Package) string {
	return fmt.Sprintf("[%s()]", fn.Name)
}

// FormatMethod formats a reference to a method on a type
func (f *DefaultLinkFormatter) FormatMethod(method *reflect.Method, typ *reflect.Type, pkg *reflect.Package) string {
	return fmt.Sprintf("[%s.%s()]", typ.Name, method.Name)
}

// FormatConstant formats a reference to a constant
func (f *DefaultLinkFormatter) FormatConstant(const_ *reflect.Constant, pkg *reflect.Package) string {
	return fmt.Sprintf("[%s]", const_.Name)
}

// FormatField formats a reference to a struct field
func (f *DefaultLinkFormatter) FormatField(field *reflect.Field, typ *reflect.Type, pkg *reflect.Package) string {
	return fmt.Sprintf("[%s.%s]", typ.Name, field.Name)
}

// FormatEnumValue formats a reference to an enum value (constant in an enum)
func (f *DefaultLinkFormatter) FormatEnumValue(const_ *reflect.Constant, enum *reflect.Enum, pkg *reflect.Package) string {
	if enum.Type != nil && enum.Type.Type != nil {
		return fmt.Sprintf("[%s.%s]", enum.Type.Type.Name, const_.Name)
	}
	return fmt.Sprintf("[%s]", const_.Name)
}
