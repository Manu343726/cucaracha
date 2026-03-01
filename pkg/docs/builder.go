package docs

import (
	"fmt"
	"go/doc/comment"
	"log/slog"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/reflect"
	"github.com/Manu343726/cucaracha/pkg/utils/contract"
)

// Builder orchestrates the creation of a documentation index from reflect packages
// It parses Doc comments from reflect entries and builds a structured documentation index
type Builder struct {
	contract.Base

	// Options for building documentation
	Options *BuilderOptions

	// index being built
	index *DocumentationIndex
}

// BuilderOptions controls behavior of the documentation builder
type BuilderOptions struct {
	// IncludePrivate includes unexported (private) items in the documentation
	IncludePrivate bool

	// ResolveReferences enables linking between entries based on type usage
	ResolveReferences bool

	// ReflectIndex is an optional pre-built reflect.Index for typed entity searching
	// If provided, the builder will use this index instead of creating its own.
	// If nil, the builder will create a new index from the packages being documented.
	ReflectIndex *reflect.Index
}

// DefaultBuilderOptions returns builder options with sensible defaults
func DefaultBuilderOptions() *BuilderOptions {
	return &BuilderOptions{
		IncludePrivate:    false,
		ResolveReferences: true,
	}
}

// NewBuilder creates a new documentation builder with default options
func NewBuilder() *Builder {
	return NewBuilderWithOptions(DefaultBuilderOptions())
}

// NewBuilderWithOptions creates a new documentation builder with custom options
func NewBuilderWithOptions(opts *BuilderOptions) *Builder {
	index := &DocumentationIndex{
		Entries:    make(map[string]*DocumentationEntry),
		ByPackage:  make(map[string][]string),
		ByKind:     make(map[string][]string),
		References: make(map[string][]string),
		Metadata:   IndexMetadata{Version: "1.0"},
	}

	// Initialize the reflect.Index if provided in options, otherwise create empty
	if opts.ReflectIndex != nil {
		index.ReflectIndex = opts.ReflectIndex
	} else {
		index.ReflectIndex = reflect.NewIndex()
	}

	return &Builder{
		Base:    contract.NewBase(log().Child("Builder")),
		Options: opts,
		index:   index,
	}
}

// Build creates a documentation index from one or more packages
// It extracts and parses Doc comments from reflect package entries and builds a structured index
// If no reflect.Index was provided in options, it builds one from the packages during processing.
func (b *Builder) Build(packages ...*reflect.Package) (*DocumentationIndex, error) {
	b.Log().Info("Starting documentation build", slog.Int("packageCount", len(packages)))

	for _, pkg := range packages {
		// Add package to reflect.Index if it's not already there
		// (works for both self-created and externally provided indices)
		if b.index.ReflectIndex.Package(pkg.Path) == nil {
			b.index.ReflectIndex.AddPackage(pkg)
		}

		if err := b.buildPackage(pkg); err != nil {
			return nil, fmt.Errorf("failed to build documentation for package %s: %w", pkg.Name, err)
		}
	}

	if b.Options.ResolveReferences {
		if err := b.resolveReferences(); err != nil {
			return nil, fmt.Errorf("failed to resolve references: %w", err)
		}
	}

	b.Log().Info("Documentation build complete",
		slog.Int("entriesCount", len(b.index.Entries)),
		slog.Int("packagesIndexed", len(b.index.Metadata.PackagesIndexed)))

	return b.index, nil
}

// buildPackage processes a single reflect.Package to extract documentation
func (b *Builder) buildPackage(pkg *reflect.Package) error {
	b.Log().Debug("Building package documentation", slog.String("package", pkg.Path))

	// Track this package in metadata
	b.index.Metadata.PackagesIndexed = append(b.index.Metadata.PackagesIndexed, pkg.Path)
	b.index.ByPackage[pkg.Path] = []string{}

	// Extract type documentation
	for _, typ := range pkg.Types {
		if !b.Options.IncludePrivate && !isExported(typ.Name) {
			continue
		}
		if err := b.buildTypeEntry(pkg, typ); err != nil {
			b.Log().Warn("Failed to build type documentation",
				slog.String("type", typ.Name),
				slog.String("error", err.Error()))
		}

		// Extract interface method documentation
		if typ.IsInterface() {
			for _, method := range typ.Methods {
				if !b.Options.IncludePrivate && !isExported(method.Name) {
					continue
				}
				if err := b.buildMethodEntry(pkg, typ, method); err != nil {
					b.Log().Warn("Failed to build method documentation",
						slog.String("type", typ.Name),
						slog.String("method", method.Name),
						slog.String("error", err.Error()))
				}
			}
		}
	}

	// Extract function documentation
	for _, fn := range pkg.Functions {
		if !b.Options.IncludePrivate && !isExported(fn.Name) {
			continue
		}
		if err := b.buildFunctionEntry(pkg, fn); err != nil {
			b.Log().Warn("Failed to build function documentation",
				slog.String("function", fn.Name),
				slog.String("error", err.Error()))
		}
	}

	// Extract constant documentation
	for _, const_ := range pkg.Constants {
		if !b.Options.IncludePrivate && !isExported(const_.Name) {
			continue
		}
		if err := b.buildConstantEntry(pkg, const_); err != nil {
			b.Log().Warn("Failed to build constant documentation",
				slog.String("constant", const_.Name),
				slog.String("error", err.Error()))
		}
	}

	return nil
}

// buildTypeEntry parses a reflect.Type's Doc comment and creates a documentation entry
func (b *Builder) buildTypeEntry(pkg *reflect.Package, typ *reflect.Type) error {
	if typ.Doc == "" {
		return nil // No documentation for this type
	}

	parsedDoc, err := b.parseDocComment(typ.Doc)
	if err != nil {
		return err
	}

	qualName := pkg.Path + "." + typ.Name
	entry := b.buildEntryFromParsedDoc(
		qualName,
		typ.Name,
		KindType,
		pkg.Path,
		parsedDoc,
	)

	if err := b.AddEntry(entry); err != nil {
		return err
	}

	// Extract field documentation for any type with fields
	if len(typ.Fields) > 0 {
		for _, field := range typ.Fields {
			if !b.Options.IncludePrivate && !isExported(field.Name) {
				continue
			}
			if err := b.buildStructFieldEntry(pkg, typ, field); err != nil {
				b.Log().Warn("Failed to build field documentation",
					slog.String("type", typ.Name),
					slog.String("field", field.Name),
					slog.String("error", err.Error()))
			}
		}
	}

	// Extract method documentation for struct/concrete types only
	// Interface methods are handled separately in buildPackage as KindInterfaceMethod
	if !typ.IsInterface() && len(typ.Methods) > 0 {
		for _, method := range typ.Methods {
			if !b.Options.IncludePrivate && !isExported(method.Name) {
				continue
			}
			if err := b.buildStructMethodEntry(pkg, typ, method); err != nil {
				b.Log().Warn("Failed to build method documentation",
					slog.String("type", typ.Name),
					slog.String("method", method.Name),
					slog.String("error", err.Error()))
			}
		}
	}

	return nil
}

// buildStructFieldEntry parses a struct field's Doc comment and creates a documentation entry
func (b *Builder) buildStructFieldEntry(pkg *reflect.Package, typ *reflect.Type, field *reflect.Field) error {
	if field.Doc == "" {
		return nil // No documentation for this field
	}

	parsedDoc, err := b.parseDocComment(field.Doc)
	if err != nil {
		return err
	}

	// Qualified name format: pkg.path.TypeName.FieldName
	qualName := pkg.Path + "." + typ.Name + "." + field.Name
	entry := b.buildEntryFromParsedDoc(
		qualName,
		field.Name,
		KindField,
		pkg.Path,
		parsedDoc,
	)

	return b.AddEntry(entry)
}

// buildStructMethodEntry parses a struct method's Doc comment and creates a documentation entry
func (b *Builder) buildStructMethodEntry(pkg *reflect.Package, typ *reflect.Type, method *reflect.Method) error {
	if method.Doc == "" {
		return nil // No documentation for this method
	}

	parsedDoc, err := b.parseDocComment(method.Doc)
	if err != nil {
		return err
	}

	// Qualified name format: pkg.path.TypeName.MethodName
	qualName := pkg.Path + "." + typ.Name + "." + method.Name
	entry := b.buildEntryFromParsedDoc(
		qualName,
		method.Name,
		KindMethod,
		pkg.Path,
		parsedDoc,
	)

	return b.AddEntry(entry)
}

// buildMethodEntry parses a reflect.Method's Doc comment and creates a documentation entry
func (b *Builder) buildMethodEntry(pkg *reflect.Package, typ *reflect.Type, method *reflect.Method) error {
	if method.Doc == "" {
		return nil // No documentation for this method
	}

	parsedDoc, err := b.parseDocComment(method.Doc)
	if err != nil {
		return err
	}

	// Qualified name format: pkg.path.TypeName.MethodName
	qualName := pkg.Path + "." + typ.Name + "." + method.Name
	entry := b.buildEntryFromParsedDoc(
		qualName,
		method.Name,
		KindInterfaceMethod,
		pkg.Path,
		parsedDoc,
	)

	return b.AddEntry(entry)
}

// buildFunctionEntry parses a reflect.Function's Doc comment and creates a documentation entry
func (b *Builder) buildFunctionEntry(pkg *reflect.Package, fn *reflect.Function) error {
	if fn.Doc == "" {
		return nil // No documentation for this function
	}

	parsedDoc, err := b.parseDocComment(fn.Doc)
	if err != nil {
		return err
	}

	qualName := pkg.Path + "." + fn.Name
	entry := b.buildEntryFromParsedDoc(
		qualName,
		fn.Name,
		KindFunction,
		pkg.Path,
		parsedDoc,
	)

	return b.AddEntry(entry)
}

// buildConstantEntry parses a reflect.Constant's Doc comment and creates a documentation entry
func (b *Builder) buildConstantEntry(pkg *reflect.Package, const_ *reflect.Constant) error {
	if const_.Doc == "" {
		return nil // No documentation for this constant
	}

	parsedDoc, err := b.parseDocComment(const_.Doc)
	if err != nil {
		return err
	}

	qualName := pkg.Path + "." + const_.Name
	entry := b.buildEntryFromParsedDoc(
		qualName,
		const_.Name,
		KindConstant,
		pkg.Path,
		parsedDoc,
	)

	return b.AddEntry(entry)
}

// parseDocComment parses a Go doc comment string into a structured comment.Doc
func (b *Builder) parseDocComment(rawDoc string) (*comment.Doc, error) {
	parser := &comment.Parser{}
	doc := parser.Parse(rawDoc)
	if doc == nil {
		return nil, fmt.Errorf("failed to parse doc comment")
	}
	return doc, nil
}

// buildEntryFromParsedDoc converts a parsed comment.Doc into a DocumentationEntry
// It also extracts links and replaces them in the text with {{Links[N]}} placeholders
func (b *Builder) buildEntryFromParsedDoc(
	qualName, localName string,
	kind EntryKind,
	pkgPath string,
	doc *comment.Doc,
) *DocumentationEntry {
	// Extract links and get processed text with link placeholders
	var allLinks []Link
	linker := &linkReplacer{links: &allLinks}

	entry := &DocumentationEntry{
		QualifiedName: qualName,
		LocalName:     localName,
		Kind:          kind,
		PackagePath:   pkgPath,
		Summary:       linker.processText(extractSummaryBlocks(doc)),
		Details:       linker.processText(extractDetailsBlocks(doc)),
		Examples:      extractExamples(doc),
		Links:         allLinks,
	}
	return entry
}

// linkReplacer helps replace doc links with indexed placeholders
type linkReplacer struct {
	links *[]Link
	index int
}

// processText walks through text elements, extracts links, and returns text with {{Links[N]}} placeholders
func (lr *linkReplacer) processText(text string) string {
	if text == "" {
		return ""
	}

	// Scan for [name] or [pkg.name] patterns and replace with {{Links[N]}}
	// This is a simple regex-based replacement
	var result strings.Builder
	i := 0
	for i < len(text) {
		if text[i] == '[' {
			// Find the closing bracket
			close := strings.IndexByte(text[i+1:], ']')
			if close != -1 {
				linkText := text[i+1 : i+1+close]
				// Create a link entry
				link := Link{
					Target:       linkText,
					Relationship: RelationshipRelated, // default relationship
				}
				*lr.links = append(*lr.links, link)
				// Replace [linkText] with {{Links[N]}}
				fmt.Fprintf(&result, "{{Links[%d]}}", lr.index)
				lr.index++
				i = i + 1 + close + 1
				continue
			}
		}
		result.WriteByte(text[i])
		i++
	}

	return result.String()
}

// extractSummaryBlocks extracts text from first 1-2 paragraphs for summary
func extractSummaryBlocks(doc *comment.Doc) string {
	var summary strings.Builder
	parCount := 0
	for _, block := range doc.Content {
		if parCount >= 2 {
			break
		}
		if para, ok := block.(*comment.Paragraph); ok {
			if summary.Len() > 0 {
				summary.WriteString(" ")
			}
			summary.WriteString(blockString(para))
			parCount++
		}
	}
	return strings.TrimSpace(summary.String())
}

// extractDetailsBlocks extracts all content as detailed documentation
func extractDetailsBlocks(doc *comment.Doc) string {
	var details strings.Builder
	for i, block := range doc.Content {
		if i > 0 {
			details.WriteString("\n\n")
		}
		details.WriteString(blockString(block))
	}
	return strings.TrimSpace(details.String())
}

// commentTextString converts a comment item to string
func commentTextString(t interface{}) string {
	// comment.Text is an interface with unexported methods
	// For now, return empty string to avoid type assertion issues
	return ""
}

// blockString converts a comment block to string, preserving link syntax for processing
func blockString(b comment.Block) string {
	switch v := b.(type) {
	case *comment.Paragraph:
		// Parse paragraph text by extracting content from Text elements
		return textSliceToString(v.Text)
	case *comment.Code:
		return v.Text
	case *comment.Heading:
		// Parse heading text
		return textSliceToString(v.Text)
	case *comment.List:
		var text strings.Builder
		for i, item := range v.Items {
			if i > 0 {
				text.WriteString("\n")
			}
			for _, b := range item.Content {
				text.WriteString(blockString(b))
			}
		}
		return text.String()
	}
	return ""
}

// textSliceToString converts a slice of comment.Text elements to a string
func textSliceToString(texts []comment.Text) string {
	var buf strings.Builder
	for _, t := range texts {
		switch v := t.(type) {
		case comment.Plain:
			buf.WriteString(string(v))
		case comment.Italic:
			buf.WriteString(string(v))
		case *comment.Link:
			// For links, include the text and target notation: text[target]
			if v != nil {
				buf.WriteString("[")
				buf.WriteString(textSliceToString(v.Text))
				buf.WriteString("]")
				if v.URL != "" {
					buf.WriteString("(" + v.URL + ")")
				}
			}
		case *comment.DocLink:
			// For doc links, include the reference text
			if v != nil {
				buf.WriteString("[")
				buf.WriteString(textSliceToString(v.Text))
				buf.WriteString("]")
			}
		}
	}
	return buf.String()
}

// extractExamples extracts code examples from the parsed doc
func extractExamples(doc *comment.Doc) []Example {
	var examples []Example
	for _, block := range doc.Content {
		if code, ok := block.(*comment.Code); ok {
			examples = append(examples, Example{
				Code: code.Text,
			})
		}
	}
	return examples
}

// resolveReferences creates links between entries based on doc link references
func (b *Builder) resolveReferences() error {
	// TODO: Implement reference resolution by scanning for doc links
	// This will create Links between entries based on [Name] references in doc comments
	return nil
}

// AddEntry adds a documentation entry to the index
func (b *Builder) AddEntry(entry *DocumentationEntry) error {
	if entry.QualifiedName == "" {
		return fmt.Errorf("entry must have a qualified name")
	}

	b.index.Entries[entry.QualifiedName] = entry

	// Index by package
	if _, exists := b.index.ByPackage[entry.PackagePath]; !exists {
		b.index.ByPackage[entry.PackagePath] = []string{}
	}
	b.index.ByPackage[entry.PackagePath] = append(
		b.index.ByPackage[entry.PackagePath],
		entry.QualifiedName,
	)

	// Index by kind
	kindStr := string(entry.Kind)
	if _, exists := b.index.ByKind[kindStr]; !exists {
		b.index.ByKind[kindStr] = []string{}
	}
	b.index.ByKind[kindStr] = append(b.index.ByKind[kindStr], entry.QualifiedName)

	b.Log().Debug("Added documentation entry",
		slog.String("name", entry.QualifiedName),
		slog.String("kind", kindStr))

	return nil
}

// isExported reports whether a name is exported
func isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}
