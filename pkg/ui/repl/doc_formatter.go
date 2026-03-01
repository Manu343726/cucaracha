package repl

import (
	"fmt"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/docs"
	"github.com/Manu343726/cucaracha/pkg/reflect"
)

// REPLLinkFormatter implements docs.LinkFormatter for REPL help output.
// It formats links using REPL syntax (command names in backticks).
type REPLLinkFormatter struct {
	syntaxFormatter REPLSyntax
}

// NewREPLLinkFormatter creates a new REPL link formatter
func NewREPLLinkFormatter() *REPLLinkFormatter {
	return &REPLLinkFormatter{
		syntaxFormatter: REPLSyntax{},
	}
}

// FormatPackage formats a reference to a package
func (f *REPLLinkFormatter) FormatPackage(pkg *reflect.Package) string {
	return fmt.Sprintf("`%s`", pkg.Name)
}

// FormatType formats a reference to a type
func (f *REPLLinkFormatter) FormatType(typ *reflect.Type, pkg *reflect.Package) string {
	return fmt.Sprintf("`%s`", typ.Name)
}

// FormatFunction formats a reference to a function
func (f *REPLLinkFormatter) FormatFunction(fn *reflect.Function, pkg *reflect.Package) string {
	return fmt.Sprintf("`%s()`", fn.Name)
}

// FormatMethod formats a reference to a method on a type
func (f *REPLLinkFormatter) FormatMethod(method *reflect.Method, typ *reflect.Type, pkg *reflect.Package) string {
	return fmt.Sprintf("`%s.%s()`", typ.Name, method.Name)
}

// FormatConstant formats a reference to a constant
func (f *REPLLinkFormatter) FormatConstant(const_ *reflect.Constant, pkg *reflect.Package) string {
	return fmt.Sprintf("`%s`", const_.Name)
}

// FormatField formats a reference to a struct field
func (f *REPLLinkFormatter) FormatField(field *reflect.Field, typ *reflect.Type, pkg *reflect.Package) string {
	return fmt.Sprintf("`%s.%s`", typ.Name, field.Name)
}

// FormatEnumValue formats a reference to an enum value (constant in an enum)
func (f *REPLLinkFormatter) FormatEnumValue(const_ *reflect.Constant, enum *reflect.Enum, pkg *reflect.Package) string {
	if enum.Type != nil && enum.Type.Type != nil {
		return fmt.Sprintf("`%s.%s`", enum.Type.Type.Name, const_.Name)
	}
	return fmt.Sprintf("`%s`", const_.Name)
}

// FormatLink formats a generic link when the source entity is unknown
func (f *REPLLinkFormatter) FormatLink(target string, sourceEntity any) string {
	return fmt.Sprintf("`%s`", target)
}

// REPLDocumentationRenderer renders documentation for REPL help output.
// It wraps text to respect the REPL layout constraints (indentation, line width).
type REPLDocumentationRenderer struct {
	lineWidth       int // Maximum line width (0 = no limit)
	indentation     string
	linkFormatter   docs.LinkFormatter
	docIndex        *docs.DocumentationIndex // Reference to the documentation index for field lookup
	packagePath     string                   // Package path for qualified name lookups
	syntaxFormatter REPLSyntax               // For converting field names to REPL syntax
}

// NewREPLDocumentationRenderer creates a new REPL documentation renderer
func NewREPLDocumentationRenderer(lineWidth int, indentation string) *REPLDocumentationRenderer {
	return &REPLDocumentationRenderer{
		lineWidth:     lineWidth,
		indentation:   indentation,
		linkFormatter: NewREPLLinkFormatter(),
	}
}

// NewREPLDocumentationRendererWithIndex creates a renderer with access to the documentation index
func NewREPLDocumentationRendererWithIndex(lineWidth int, indentation string, docIndex *docs.DocumentationIndex, packagePath string) *REPLDocumentationRenderer {
	return &REPLDocumentationRenderer{
		lineWidth:       lineWidth,
		indentation:     indentation,
		linkFormatter:   NewREPLLinkFormatter(),
		docIndex:        docIndex,
		packagePath:     packagePath,
		syntaxFormatter: REPLSyntax{},
	}
}

// RenderDocumentation renders documentation text with proper wrapping and link formatting.
// It takes the raw documentation text and returns formatted text suitable for REPL output.
// The text respects the specified indentation and line width constraints.
func (r *REPLDocumentationRenderer) RenderDocumentation(entry *docs.DocumentationEntry, indent string) string {
	if entry == nil {
		return ""
	}

	var output strings.Builder

	// Render summary if available
	if entry.Summary != "" {
		output.WriteString(r.wrapAndIndent(entry.Summary, indent))
	}

	// Render details if available
	if entry.Details != "" {
		if output.Len() > 0 {
			output.WriteString("\n")
		}
		output.WriteString(r.wrapAndIndent(entry.Details, indent))
	}

	return output.String()
}

// wrapAndIndent wraps text to the specified line width while preserving indentation.
// It handles paragraph breaks (double newlines) and respects the indentation prefix.
func (r *REPLDocumentationRenderer) wrapAndIndent(text, indent string) string {
	// Split by double newlines to preserve paragraph breaks
	paragraphs := strings.Split(text, "\n\n")

	var result strings.Builder
	for i, paragraph := range paragraphs {
		if i > 0 {
			result.WriteString("\n")
		}

		// Clean up and wrap each paragraph
		wrapped := r.wrapParagraph(strings.TrimSpace(paragraph), indent)
		result.WriteString(wrapped)
	}

	return result.String()
}

// wrapParagraph wraps a single paragraph to the specified line width,
// respecting the indentation prefix.
func (r *REPLDocumentationRenderer) wrapParagraph(paragraph, indent string) string {
	if r.lineWidth == 0 {
		// No line wrapping
		lines := strings.Split(paragraph, "\n")
		var result strings.Builder
		for i, line := range lines {
			if i > 0 {
				result.WriteString("\n")
			}
			result.WriteString(indent)
			result.WriteString(strings.TrimSpace(line))
		}
		return result.String()
	}

	// Wrap text to respect line width
	availableWidth := r.lineWidth - len(indent)
	if availableWidth < 20 {
		availableWidth = 80 // Fallback if indent is too long
	}

	// Split paragraph into lines and join them for wrapping
	lines := strings.Split(paragraph, "\n")
	cleanText := strings.Join(lines, " ")

	var result strings.Builder
	words := strings.Fields(cleanText)

	currentLine := ""
	for _, word := range words {
		// Check if adding this word would exceed line width
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) <= availableWidth || currentLine == "" {
			currentLine = testLine
		} else {
			// Write current line and start a new one
			if result.Len() > 0 {
				result.WriteString("\n")
			}
			result.WriteString(indent)
			result.WriteString(currentLine)
			currentLine = word
		}
	}

	// Write the last line
	if currentLine != "" {
		if result.Len() > 0 {
			result.WriteString("\n")
		}
		result.WriteString(indent)
		result.WriteString(currentLine)
	}

	return result.String()
}

// wrapFieldText wraps text for a field item where the first line doesn't get indented
// (it comes right after the field name and colon), but continuation lines do.
func (r *REPLDocumentationRenderer) wrapFieldText(text, continuationIndent string) string {
	if r.lineWidth == 0 {
		// No line wrapping
		return text
	}

	// Calculate available width: line width minus continuation indent length
	availableWidth := r.lineWidth - len(continuationIndent)
	if availableWidth < 20 {
		availableWidth = 80 // Fallback if indent is too long
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var result strings.Builder
	currentLine := ""

	for i, word := range words {
		if i == 0 {
			// First word goes on the first line (no indent needed)
			currentLine = word
			continue
		}

		// For subsequent words, check if they fit on current line
		testLine := currentLine + " " + word

		// Calculate length without the continuation indent (since it's not on the first line)
		if len(testLine) <= availableWidth || currentLine == "" {
			currentLine = testLine
		} else {
			// Current line is full, write it and start a new one
			result.WriteString(currentLine)
			result.WriteString("\n")
			result.WriteString(continuationIndent)
			currentLine = word
		}
	}

	// Write the last line
	if currentLine != "" {
		result.WriteString(currentLine)
	}

	return result.String()
}

// FormatCommandLinks replaces {{Links[N]}} placeholders with formatted links
// This is used to render documentation entries with resolved links
func (r *REPLDocumentationRenderer) FormatCommandLinks(text string, entry *docs.DocumentationEntry) string {
	if entry == nil || len(entry.Links) == 0 {
		return text
	}

	result := text
	for i := range entry.Links {
		placeholder := fmt.Sprintf("{{Links[%d]}}", i)
		link := &entry.Links[i]

		formattedLink := r.formatLink(link)
		result = strings.ReplaceAll(result, placeholder, formattedLink)
	}

	return result
}

// formatLink uses the LinkFormatter to format a link
func (r *REPLDocumentationRenderer) formatLink(link *docs.Link) string {
	if link == nil {
		return ""
	}
	if r.linkFormatter == nil || link.Source == nil {
		return link.Target
	}

	source := link.Source
	pkg := link.SourcePackage

	// Switch on the non-nil field in the DocumentationSource
	switch {
	case source.Package != nil:
		return r.linkFormatter.FormatPackage(source.Package)

	case source.Type != nil:
		return r.linkFormatter.FormatType(source.Type, pkg)

	case source.Function != nil:
		return r.linkFormatter.FormatFunction(source.Function, pkg)

	case source.Method != nil:
		return r.linkFormatter.FormatMethod(source.Method, source.Type, pkg)

	case source.Constant != nil:
		return r.linkFormatter.FormatConstant(source.Constant, pkg)

	case source.Field != nil:
		return r.linkFormatter.FormatField(source.Field, source.Type, pkg)

	case source.Enum != nil:
		if source.Enum.Type != nil && source.Enum.Type.Type != nil {
			return r.linkFormatter.FormatType(source.Enum.Type.Type, pkg)
		}
		return r.linkFormatter.FormatLink(link.Target, source)

	default:
		return r.linkFormatter.FormatLink(link.Target, source)
	}
}

// RenderCommandHelp renders a complete command help with structured sections.
// Format:
//
//	command name: summary
//	                            next line of summary
//
//	                            details
//
//	    args:
//	        - argName: arg summary
//	                                       arg details
//
//	    results:
//	        - resultName: result summary
//	                                       result details
func (r *REPLDocumentationRenderer) RenderCommandHelp(commandName string, methodDoc, argsDoc, resultDoc *docs.DocumentationEntry) string {
	var output strings.Builder
	baseIndent := "    "

	// Command name: summary
	output.WriteString(fmt.Sprintf("%s: ", commandName))
	if methodDoc != nil && methodDoc.Summary != "" {
		// Resolve links in summary
		summary := methodDoc.Summary
		if len(methodDoc.Links) > 0 {
			summary = r.FormatCommandLinks(summary, methodDoc)
		}
		summaryIndent := strings.Repeat(" ", len(commandName)+2)
		output.WriteString(r.wrapFieldText(summary, summaryIndent))
	}

	// Details section
	if methodDoc != nil && methodDoc.Details != "" {
		// Resolve links in details
		details := methodDoc.Details
		if len(methodDoc.Links) > 0 {
			details = r.FormatCommandLinks(details, methodDoc)
		}
		output.WriteString("\n\n")
		output.WriteString(r.wrapAndIndent(details, ""))
	}

	// Args section
	if argsDoc != nil {
		output.WriteString(r.renderFieldsSection("args", argsDoc, baseIndent))
	}

	// Results section
	if resultDoc != nil {
		output.WriteString(r.renderFieldsSection("results", resultDoc, baseIndent))
	}

	return output.String()
}

// renderFieldsSection renders a section (args or results) with field names and documentation
func (r *REPLDocumentationRenderer) renderFieldsSection(sectionName string, doc *docs.DocumentationEntry, baseIndent string) string {
	if doc == nil {
		return ""
	}

	var output strings.Builder
	output.WriteString("\n\n")
	output.WriteString(baseIndent)
	output.WriteString(sectionName)
	output.WriteString(":\n")

	// Try to get struct fields from the documentation index
	fields := r.getStructFields(doc.QualifiedName)
	if len(fields) > 0 {
		r.renderStructFields(fields, &output, baseIndent)
		return output.String()
	}

	// Fallback: just show the struct documentation itself
	fieldIndent := baseIndent + "    "

	if doc.Summary != "" {
		summary := doc.Summary
		if len(doc.Links) > 0 {
			summary = r.FormatCommandLinks(summary, doc)
		}

		output.WriteString(fieldIndent)
		output.WriteString("- ")
		
		// Calculate continuation indent for this single item
		continuationIndent := strings.Repeat(" ", len(fieldIndent)+2)
		output.WriteString(r.wrapFieldText(summary, continuationIndent))

		if doc.Details != "" {
			details := doc.Details
			if len(doc.Links) > 0 {
				details = r.FormatCommandLinks(details, doc)
			}

			output.WriteString("\n")
			output.WriteString(continuationIndent)
			output.WriteString(r.wrapAndIndent(details, continuationIndent))
		}
	}

	return output.String()
}

// getStructFields retrieves all field entries for a given struct type
func (r *REPLDocumentationRenderer) getStructFields(structQualName string) []*docs.DocumentationEntry {
	var fields []*docs.DocumentationEntry

	// Look for all entries that are fields of this struct
	// Field qualified names follow the pattern: "packagePath.StructName.FieldName"
	for _, entry := range r.docIndex.Entries {
		if entry.Kind == docs.KindField && strings.HasPrefix(entry.QualifiedName, structQualName+".") {
			fields = append(fields, entry)
		}
	}

	return fields
}

// renderStructFields renders the fields of a struct type
func (r *REPLDocumentationRenderer) renderStructFields(fields []*docs.DocumentationEntry, output *strings.Builder, baseIndent string) {
	if len(fields) == 0 {
		return
	}

	fieldIndent := baseIndent + "    "

	for i, field := range fields {
		if i > 0 {
			output.WriteString("\n\n")
		}

		// Extract field name from qualified name (last part after the last dot)
		parts := strings.Split(field.QualifiedName, ".")
		rawFieldName := parts[len(parts)-1]
		// Format field name using REPL syntax convention
		fieldName := r.syntaxFormatter.FormatArgumentName(rawFieldName, true)

		output.WriteString(fieldIndent)
		output.WriteString("- ")
		output.WriteString(fieldName)
		output.WriteString(": ")

		// Calculate continuation indent: field indent (4 + 2) + field name + ": "
		// Total indent = len(baseIndent) + len("    " + "- ") + len(fieldName) + len(": ")
		continuationIndent := strings.Repeat(" ", len(fieldIndent)+2+len(fieldName)+2)

		// Render field documentation - first line doesn't get extra indent
		if field.Summary != "" {
			summary := field.Summary
			if len(field.Links) > 0 {
				summary = r.FormatCommandLinks(summary, field)
			}
			output.WriteString(r.wrapFieldText(summary, continuationIndent))
		}

		// Add details if available
		if field.Details != "" {
			output.WriteString("\n")
			output.WriteString(continuationIndent)
			details := field.Details
			if len(field.Links) > 0 {
				details = r.FormatCommandLinks(details, field)
			}
			output.WriteString(r.wrapAndIndent(details, continuationIndent))
		}
	}
}
