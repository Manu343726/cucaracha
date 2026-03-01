package docs

import (
	"github.com/Manu343726/cucaracha/pkg/reflect"
)

// DocumentationIndex is the main structured index of documentation entries.
// It provides both indexed access and organized retrieval of documentation.
type DocumentationIndex struct {
	// ReflectIndex is an internal index for efficient typed searching of reflect entities
	// It allows looking up packages, types, functions, methods, fields, and constants
	// by their fully qualified names. Built by the builder during documentation extraction.
	// Not serialized to JSON as it contains code references.
	ReflectIndex *reflect.Index `json:"-"`

	// Entries maps fully qualified names to documentation entries
	// Examples: "github.com/user/repo/pkg.TypeName", "github.com/user/repo/pkg.FunctionName"
	Entries map[string]*DocumentationEntry `json:"entries"`

	// ByPackage groups documentation entries by package path
	// Maps package import path to a slice of entry keys
	ByPackage map[string][]string `json:"byPackage"`

	// ByKind groups documentation entries by their kind (Type, Function, Constant, etc.)
	// Maps kind string to a slice of entry keys
	ByKind map[string][]string `json:"byKind"`

	// References maps entry keys to all other entries they reference
	// Used for building dependency graphs and discovering related documentation
	References map[string][]string `json:"references"`

	// Metadata about the index (creation time, processing info, etc.)
	Metadata IndexMetadata `json:"metadata"`
}

// IndexMetadata contains metadata about the documentation index
type IndexMetadata struct {
	// Version of the documentation format
	Version string `json:"version"`

	// PackagesIndexed is a list of all package paths that were indexed
	PackagesIndexed []string `json:"packagesIndexed"`

	// SourceInfo describes where the documentation was extracted from
	SourceInfo string `json:"sourceInfo,omitempty"`
}

// DocumentationSource acts as a sum type for the various reflect entities that can be documented.
// At most one of these fields will be non-nil for any given DocumentationSource.
// This provides type-safe access to the original reflect entities during rendering and analysis.
type DocumentationSource struct {
	// Package is the reflect.Package if this source documents a package
	Package *reflect.Package `json:"-"`

	// Type is the reflect.Type if this source documents a type (struct, interface, alias, etc.)
	Type *reflect.Type `json:"-"`

	// Function is the reflect.Function if this source documents a function
	Function *reflect.Function `json:"-"`

	// Method is the reflect.Method if this source documents a method
	Method *reflect.Method `json:"-"`

	// Constant is the reflect.Constant if this source documents a constant (or enum value)
	Constant *reflect.Constant `json:"-"`

	// Field is the reflect.Field if this source documents a struct field
	Field *reflect.Field `json:"-"`

	// Enum is the reflect.Enum if this source documents an enum type
	Enum *reflect.Enum `json:"-"`
}

// DocumentationEntry represents a single documented item (type, function, package, etc.)
// Each entry is structured with multiple sections for comprehensive documentation
type DocumentationEntry struct {
	// QualifiedName is the fully qualified name of the documented item
	// Format: "packagePath.ItemName" or "packagePath.TypeName.MethodName"
	QualifiedName string `json:"qualifiedName"`

	// Kind indicates what this entry documents (Type, Function, Constant, Package, Method, etc.)
	Kind EntryKind `json:"kind"`

	// PackagePath is the import path of the package containing this item
	PackagePath string `json:"packagePath"`

	// LocalName is the unqualified name of the item (e.g., "User", "ParsePackage")
	LocalName string `json:"localName"`

	// Summary is a brief synopsis (typically the first paragraph)
	Summary string `json:"summary"`

	// Details contains comprehensive documentation with full descriptions,
	// parameter details, return values, and behavioral notes
	Details string `json:"details"`

	// Examples contains code examples demonstrating usage of the documented item
	Examples []Example `json:"examples,omitempty"`

	// Links are references to related documentation entries
	Links []Link `json:"links,omitempty"`

	// SourceLocation indicates where in the source this was documented from
	SourceLocation SourceLocation `json:"sourceLocation,omitempty"`

	// Source is a pointer to the original reflect entities that this entry documents
	// Contains exactly one non-nil field indicating what type of entity this is
	// Not serialized to JSON as it contains code references
	Source *DocumentationSource `json:"-"`
}

// EntryKind represents the kind of documentation entry
type EntryKind string

const (
	// Package documentation
	KindPackage EntryKind = "package"

	// Type documentation (struct, interface, etc.)
	KindType EntryKind = "type"

	// Function or method documentation
	KindFunction EntryKind = "function"

	// Method on a type
	KindMethod EntryKind = "method"

	// Constant documentation
	KindConstant EntryKind = "constant"

	// Enum type documentation
	KindEnum EntryKind = "enum"

	// Enum value documentation
	KindEnumValue EntryKind = "enumValue"

	// Field of a struct
	KindField EntryKind = "field"

	// Interface method signature
	KindInterfaceMethod EntryKind = "interfaceMethod"
)

// Example represents a code example in documentation
type Example struct {
	// Description of what this example demonstrates
	Description string `json:"description"`

	// Code is the example code snippet
	Code string `json:"code"`

	// Expected output of running the example (if applicable)
	Output string `json:"output,omitempty"`

	// Tags for categorizing examples (e.g., "basic", "advanced", "error-handling")
	Tags []string `json:"tags,omitempty"`
}

// Link represents a reference to another documentation entry
type Link struct {
	// QualifiedName of the referenced entry
	Target string `json:"target"`

	// Relationship describes the type of link (uses, implements, inherits, related, etc.)
	Relationship LinkRelationship `json:"relationship"`

	// Context describes why this link is relevant in this entry's documentation
	Context string `json:"context,omitempty"`

	// TargetEntry is a pointer to the documentation entry being referenced
	// Populated after the index is built and references are resolved
	// Not serialized to JSON as it creates circular references
	TargetEntry *DocumentationEntry `json:"-"`

	// Source is a pointer to the original reflect entity being referenced
	// Contains exactly one non-nil field indicating what type of entity this is
	// Not serialized to JSON as it contains code references
	Source *DocumentationSource `json:"-"`

	// SourcePackage is a pointer to the package containing the source entity
	// Used by renderers to access full type information for formatting
	// Not serialized to JSON as it contains code references
	SourcePackage *reflect.Package `json:"-"`
}

// LinkRelationship describes the relationship between two documentation entries
type LinkRelationship string

const (
	// RelationshipUses: This entry uses/references the target
	RelationshipUses LinkRelationship = "uses"

	// RelationshipUsedBy: The target uses/references this entry
	RelationshipUsedBy LinkRelationship = "usedBy"

	// RelationshipImplements: This entry implements the target interface
	RelationshipImplements LinkRelationship = "implements"

	// RelationshipImplementedBy: The target implements this interface
	RelationshipImplementedBy LinkRelationship = "implementedBy"

	// RelationshipRelated: The target is related to this entry
	RelationshipRelated LinkRelationship = "related"

	// RelationshipEmbeds: This type embeds the target type
	RelationshipEmbeds LinkRelationship = "embeds"

	// RelationshipEmbeddedBy: The target type embeds this type
	RelationshipEmbeddedBy LinkRelationship = "embeddedBy"

	// RelationshipReturns: The target is returned by this function
	RelationshipReturns LinkRelationship = "returns"

	// RelationshipParameter: The target is a parameter of this function
	RelationshipParameter LinkRelationship = "parameter"

	// RelationshipExample: The target is used in examples
	RelationshipExample LinkRelationship = "example"
)

// SourceLocation indicates where in the source code a documentation entry came from
type SourceLocation struct {
	// FilePath is the path to the source file
	FilePath string `json:"filePath"`

	// LineNumber is the line where the documented item is defined
	LineNumber int `json:"lineNumber"`

	// ColumnNumber is the column where the documented item is defined
	ColumnNumber int `json:"columnNumber"`

	// EndLineNumber indicates where the definition ends (for multi-line items)
	EndLineNumber int `json:"endLineNumber,omitempty"`
}

// BuilderSource represents the source data used to build documentation
// Can be a reflect.Package or other structure
type BuilderSource interface {
	// Source types that can be indexed
}

// ensure our source types satisfy the interface
var (
	_ BuilderSource = (*reflect.Package)(nil)
)
