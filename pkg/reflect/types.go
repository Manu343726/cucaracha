package reflect

import (
	"go/ast"
	"go/token"
)

// Package represents a parsed Go package with all its types, functions, and constants
type Package struct {
	Name      string
	Path      string
	Files     []*File
	Types     map[string]*Type
	Functions []*Function
	Constants []*Constant
	Enums     map[string]*Enum
	FileSet   *token.FileSet
}

// File represents a parsed Go source file
type File struct {
	Name      string
	Path      string
	Package   string
	Imports   []Import
	Types     map[string]*Type
	Functions []*Function
	Constants []*Constant
	Comments  string
}

// Import represents an import statement
type Import struct {
	Name  string // e.g., "os", "fmt", or alias like "sqlc" for "database/sql"
	Path  string // e.g., "os", "fmt", "database/sql"
	Alias string // e.g., "" (no alias), "." (dot import), "_" (blank import)
}

// Type represents a parsed Go type definition
type Type struct {
	Name       string
	Kind       TypeKind  // struct, interface, alias, etc.
	Doc        string    // Documentation comment
	Comments   []string  // Inline comments
	Underlying ast.Expr  // The underlying AST expression
	Fields     []*Field  // For struct types
	Methods    []*Method // For named types with methods
	Interfaces []string  // For types implementing interfaces
	SourcePos  token.Pos // Position in source file
}

// TypeKind indicates the kind of type
type TypeKind string

const (
	TypeKindStruct    TypeKind = "struct"
	TypeKindInterface TypeKind = "interface"
	TypeKindAlias     TypeKind = "alias"
	TypeKindBasic     TypeKind = "basic"
)

// Field represents a struct field
type Field struct {
	Name       string
	Type       string // String representation of the type (e.g., "string", "*int", "[]byte")
	Tag        string // Struct tag (e.g., "json:\"name\"")
	Doc        string // Documentation comment
	IsEmbedded bool   // Whether this is an embedded field
}

// Method represents a method on a type
type Method struct {
	Name      string
	Signature string // e.g., "(ctx context.Context, id string) (string, error)"
	Doc       string // Documentation comment
	Receiver  string // The receiver type
	Args      []*Parameter
	Results   []*Parameter
	SourcePos token.Pos
}

// Function represents a package-level function or method
type Function struct {
	Name      string
	Package   string // Package it belongs to
	Doc       string // Documentation comment
	Signature string // Full signature
	Args      []*Parameter
	Results   []*Parameter
	SourcePos token.Pos
	IsMethod  bool   // Whether this is a method (has receiver)
	Receiver  string // Receiver type if method
}

// Parameter represents a function parameter
type Parameter struct {
	Name string
	Type string // String representation (e.g., "string", "*int", "[]interface{}")
}

// Constant represents a package-level constant or enum value
type Constant struct {
	Name  string
	Type  string // Type of the constant (e.g., "string", "int")
	Doc   string // Documentation comment
	Value string // String representation of the value
}

// Enum represents a group of constants forming a logical enum
type Enum struct {
	Name         string      // Name of the enum type (e.g., "DebuggerCommandId")
	Type         string      // Underlying type (e.g., "int", "string")
	Doc          string      // Documentation comment
	Values       []*Constant // The constant values
	StringMethod bool        // Whether a String() method was found
	SourcePos    token.Pos
}

// Interface represents an interface type with its methods
type Interface struct {
	Name    string
	Methods []*Method
	Doc     string
}
