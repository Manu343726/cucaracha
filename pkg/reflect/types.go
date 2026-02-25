package reflect

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
)

// ParsingOptions controls the behavior of parsing operations
type ParsingOptions struct {
	// ResolveExternalTypes enables resolution of types from other packages.
	// When enabled, the parser will recursively parse and resolve types from imported packages.
	ResolveExternalTypes bool `json:"resolveExternalTypes"`

	// MaxResolutionDepth limits the recursion depth for external type resolution.
	// A value of 0 means no external resolution. A negative value means unlimited depth.
	// Default is -1 (unlimited).
	MaxResolutionDepth int `json:"maxResolutionDepth"`
}

// DefaultParsingOptions returns a ParsingOptions with sensible defaults
func DefaultParsingOptions() ParsingOptions {
	return ParsingOptions{
		ResolveExternalTypes: false,
		MaxResolutionDepth:   -1,
	}
}

// Package represents a parsed Go package with all its types, functions, and constants
type Package struct {
	Name      string           `json:"name"`      // Package name (e.g., "reflect")
	Path      string           `json:"path"`      // Full package import path (e.g., "github.com/Manu343726/cucaracha/pkg/reflect")
	Files     []*File          `json:"files"`     // List of files in the package
	Types     map[string]*Type `json:"types"`     // All types in the package, keyed by type name
	Functions []*Function      `json:"functions"` // All functions in the package
	Constants []*Constant      `json:"constants"` // All constants in the package
	Enums     map[string]*Enum `json:"enums"`     // All enums in the package, keyed by enum name
	FileSet   *token.FileSet   `json:"-"`         // FileSet for position information (not serialized)
}

// File represents a parsed Go source file.
// It contains all parsed definitions from a single Go source file including imports, type definitions,
// functions, and constants.
type File struct {
	// Name is the base filename (e.g., "types.go")
	Name string `json:"name"`

	// Path is the absolute file path (e.g., "/workspace/pkg/reflect/types.go")
	Path string `json:"path"`

	// Package is the package name this file belongs to (e.g., "reflect")
	Package string `json:"package"`

	// Doc is the file-level documentation comment if present
	Doc string `json:"doc,omitempty"`

	// Imports is a list of all import statements in this file
	Imports []Import `json:"imports"`

	// Types maps type names to their Type definitions declared in this file
	Types map[string]*Type `json:"types"`

	// Functions is a list of all top-level functions declared in this file
	Functions []*Function `json:"functions"`

	// Constants is a list of all top-level constants declared in this file
	Constants []*Constant `json:"constants"`

	// Comments is a concatenation of all file comments for reference
	Comments string `json:"comments,omitempty"`
}

// Import represents an import statement in a Go source file.
// It captures the imported package path and any alias used to reference the import.
type Import struct {
	// Name is the local name used to reference the imported package.
	// This is the package name for standard imports, or an alias if one was specified.
	// Examples: "os", "fmt", "sqlc" (for "database/sql" with alias)
	Name string `json:"name"`

	// Path is the full import path of the package.
	// Examples: "os", "fmt", "database/sql", "github.com/user/repo/pkg"
	Path string `json:"path"`

	// Alias is the import alias if one was explicitly specified.
	// Empty string means no alias (use package name), "." for dot import, "_" for blank import
	Alias string `json:"alias"`
}

// Type represents a parsed Go type definition
type Type struct {
	Name       string    `json:"name"`                 // Name of the type (e.g., "User", "DebuggerCommandId")
	Kind       TypeKind  `json:"kind"`                 // struct, interface, alias, slice, map, pointer, etc.
	Doc        string    `json:"doc,omitempty"`        // Documentation comment
	Comments   []string  `json:"comments,omitempty"`   // Inline comments
	Underlying ast.Expr  `json:"-"`                    // The underlying AST expression
	Fields     []*Field  `json:"fields,omitempty"`     // For struct types
	Methods    []*Method `json:"methods,omitempty"`    // For named types with methods
	Interfaces []string  `json:"interfaces,omitempty"` // For types implementing interfaces
	SourcePos  token.Pos `json:"sourcePos"`            // Position in source file

	OriginalType *TypeReference `json:"originalType,omitempty"` // For aliases: reference to the original type

	// Fields for composite types

	Elem    *TypeReference `json:"elem,omitempty"`    // For slices, arrays, pointers: the element/item type
	Key     *TypeReference `json:"key,omitempty"`     // For maps: the key type
	Value   *TypeReference `json:"value,omitempty"`   // For maps: the value type
	Size    int            `json:"size,omitempty"`    // For arrays: the size (0 means unbounded/slice)
	ChanDir ChanDirection  `json:"chanDir,omitempty"` // For channels: the direction (send, recv, or both)

	// Fields for function/method types

	Args     []*TypeReference `json:"args,omitempty"`     // For function/method types: the argument types
	Results  []*TypeReference `json:"results,omitempty"`  // For function/method types: the result types
	Receiver *TypeReference   `json:"receiver,omitempty"` // For method types: the receiver type
}

func (t *Type) WithDoc(doc string) *Type {
	t.Doc = doc
	return t
}

func (t *Type) WithComments(comments []string) *Type {
	t.Comments = comments
	return t
}

func (t *Type) IsComposite() bool {
	return t.Kind == TypeKindSlice || t.Kind == TypeKindArray || t.Kind == TypeKindMap || t.Kind == TypeKindPointer || t.Kind == TypeKindChannel
}

func (t *Type) IsFunction() bool {
	return t.Kind == TypeKindFunction
}

func (t *Type) IsStruct() bool {
	return t.Kind == TypeKindStruct
}

func (t *Type) IsInterface() bool {
	return t.Kind == TypeKindInterface
}

func (t *Type) IsAlias() bool {
	return t.Kind == TypeKindAlias
}

func (t *Type) IsTypedef() bool {
	return t.Kind == TypeKindTypedef
}

func (t *Type) IsBasic() bool {
	return t.Kind == TypeKindBasic
}

func (t *Type) IsMethod() bool {
	return t.Kind == TypeKindMethod
}

// TypeKind indicates the kind of type
type TypeKind string

const (
	TypeKindStruct    TypeKind = "struct"
	TypeKindInterface TypeKind = "interface"
	TypeKindAlias     TypeKind = "alias"
	TypeKindTypedef   TypeKind = "typedef"
	TypeKindBasic     TypeKind = "basic"
	TypeKindSlice     TypeKind = "slice"
	TypeKindArray     TypeKind = "array"
	TypeKindMap       TypeKind = "map"
	TypeKindPointer   TypeKind = "pointer"
	TypeKindChannel   TypeKind = "channel"
	TypeKindFunction  TypeKind = "function"
	TypeKindMethod    TypeKind = "method"
)

func (k TypeKind) String() string {
	return string(k)
}

func TypeKindFromString(s string) (TypeKind, error) {
	switch s {
	case "struct":
		return TypeKindStruct, nil
	case "interface":
		return TypeKindInterface, nil
	case "alias":
		return TypeKindAlias, nil
	case "typedef":
		return TypeKindTypedef, nil
	case "basic":
		return TypeKindBasic, nil
	case "slice":
		return TypeKindSlice, nil
	case "array":
		return TypeKindArray, nil
	case "map":
		return TypeKindMap, nil
	case "pointer":
		return TypeKindPointer, nil
	case "channel":
		return TypeKindChannel, nil
	case "function":
		return TypeKindFunction, nil
	case "method":
		return TypeKindMethod, nil
	default:
		return "", fmt.Errorf("unknown TypeKind: %s", s)
	}
}

func (k TypeKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.String())
}

func (k *TypeKind) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	val, err := TypeKindFromString(s)
	if err != nil {
		return err
	}
	*k = val
	return nil
}

// ChanDirection indicates the direction of a channel
type ChanDirection int

const (
	ChanBidirectional ChanDirection = iota
	ChanSend
	ChanRecv
)

func (d ChanDirection) String() string {
	switch d {
	case ChanBidirectional:
		return "bidirectional"
	case ChanSend:
		return "send"
	case ChanRecv:
		return "recv"
	default:
		return "unknown"
	}
}

func ChanDirectionFromString(s string) (ChanDirection, error) {
	switch s {
	case "bidirectional":
		return ChanBidirectional, nil
	case "send":
		return ChanSend, nil
	case "recv":
		return ChanRecv, nil
	default:
		return 0, fmt.Errorf("unknown ChanDirection: %s", s)
	}
}

func (d ChanDirection) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *ChanDirection) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	val, err := ChanDirectionFromString(s)
	if err != nil {
		return err
	}
	*d = val
	return nil
}

// TypeReference represents a reference to a type by name and optionally a pointer to the resolved Type.
// This allows for lazy resolution of types and handling of forward references and unresolved types.
type TypeReference struct {
	// Name is the string representation of the type name.
	// Examples: "string", "*int", "[]byte", "map[string]interface{}", "error"
	Name string `json:"name"`

	// Type points to the resolved Type struct, if the type was successfully resolved.
	// May be nil if the type could not be resolved or resolution was not attempted.
	Type *Type `json:"type"`
}

// Returns a reference to the given Type
func MakeTypeReference(t *Type) *TypeReference {
	if t == nil {
		return nil
	}
	return &TypeReference{
		Name: t.Name,
		Type: t,
	}
}

// Field represents a struct field definition.
// It captures the field name, type, struct tags, and documentation.
type Field struct {
	// Name is the field name as declared in the struct.
	// Examples: "Id", "Name", "Version"
	Name string `json:"name"`

	// Type is a reference to the field's type.
	// Can reference basic types, user-defined types, or composite types.
	Type *TypeReference `json:"type"`

	// Tag contains the struct tag string if present (e.g., `json:"name" xml:"name"`)
	Tag string `json:"tag"`

	// Doc is the documentation comment for the field if present.
	Doc string `json:"doc"`

	// IsEmbedded indicates if this is an embedded/anonymous field.
	IsEmbedded bool `json:"isEmbedded"`
}

// Method represents a method definition on a type.
// It captures the method signature, receiver, parameters, return types, and documentation.
type Method struct {
	// Name is the method name.
	// Examples: "String", "ServeHTTP", "Read"
	Name string `json:"name"`

	// Signature is the method's full signature string (including receiver and return types).
	// Example: "(ctx context.Context, id string) (string, error)"
	Signature string `json:"signature"`

	// Doc is the documentation comment for the method if present.
	Doc string `json:"doc"`

	// Receiver is the method receiver (the type the method is defined on).
	Receiver *Parameter `json:"receiver"`

	// Args is a list of the method's parameters.
	Args []*Parameter `json:"args"`

	// Results is a list of the method's return types.
	Results []*Parameter `json:"results"`

	// SourcePos is the position in the source file where this method is defined.
	SourcePos token.Pos `json:"sourcePos"`
}

// Function represents a package-level function or method
type Function struct {
	Name      string       `json:"name"`      // Function name (e.g., "Parse", "NewUser")
	Package   string       `json:"package"`   // Package it belongs to
	Doc       string       `json:"doc"`       // Documentation comment
	Signature string       `json:"signature"` // Full signature
	Args      []*Parameter `json:"args"`      // Function parameters
	Results   []*Parameter `json:"results"`   // Function return types
	SourcePos token.Pos    `json:"sourcePos"` // Position in source file
}

// Parameter represents a function parameter
type Parameter struct {
	Name string         `json:"name"` // Parameter name (e.g., "ctx", "id")
	Type *TypeReference `json:"type"` // Reference to the parameter type
}

// Constant represents a package-level constant or enum value
type Constant struct {
	Name  string `json:"name"`  // Constant name (e.g., "MaxSize", "DebuggerCommandStep")
	Doc   string `json:"doc"`   // Documentation comment
	Value *Value `json:"value"` // Value of the constant
}

// Enum represents a group of constants forming a logical enum
type Enum struct {
	Type         *TypeReference `json:"type"`         // Enum type
	Values       []*Constant    `json:"values"`       // The constant values
	StringMethod bool           `json:"stringMethod"` // Whether a String() method was found
	SourcePos    token.Pos      `json:"sourcePos"`    // Position in source file
}

// ============================================================================
// Global Basic Type Variables
// ============================================================================

// TypeBool is the basic bool type
var TypeBool = &Type{Name: "bool", Kind: TypeKindBasic}

// TypeByte is the basic byte type (alias for uint8)
var TypeByte = &Type{Name: "byte", Kind: TypeKindBasic}

// TypeRune is the basic rune type (alias for int32)
var TypeRune = &Type{Name: "rune", Kind: TypeKindBasic}

// TypeInt is the basic int type
var TypeInt = &Type{Name: "int", Kind: TypeKindBasic}

// TypeInt8 is the basic int8 type
var TypeInt8 = &Type{Name: "int8", Kind: TypeKindBasic}

// TypeInt16 is the basic int16 type
var TypeInt16 = &Type{Name: "int16", Kind: TypeKindBasic}

// TypeInt32 is the basic int32 type
var TypeInt32 = &Type{Name: "int32", Kind: TypeKindBasic}

// TypeInt64 is the basic int64 type
var TypeInt64 = &Type{Name: "int64", Kind: TypeKindBasic}

// TypeUint is the basic uint type
var TypeUint = &Type{Name: "uint", Kind: TypeKindBasic}

// TypeUint8 is the basic uint8 type
var TypeUint8 = &Type{Name: "uint8", Kind: TypeKindBasic}

// TypeUint16 is the basic uint16 type
var TypeUint16 = &Type{Name: "uint16", Kind: TypeKindBasic}

// TypeUint32 is the basic uint32 type
var TypeUint32 = &Type{Name: "uint32", Kind: TypeKindBasic}

// TypeUint64 is the basic uint64 type
var TypeUint64 = &Type{Name: "uint64", Kind: TypeKindBasic}

// TypeUintptr is the basic uintptr type
var TypeUintptr = &Type{Name: "uintptr", Kind: TypeKindBasic}

// TypeFloat32 is the basic float32 type
var TypeFloat32 = &Type{Name: "float32", Kind: TypeKindBasic}

// TypeFloat64 is the basic float64 type
var TypeFloat64 = &Type{Name: "float64", Kind: TypeKindBasic}

// TypeComplex64 is the basic complex64 type
var TypeComplex64 = &Type{Name: "complex64", Kind: TypeKindBasic}

// TypeComplex128 is the basic complex128 type
var TypeComplex128 = &Type{Name: "complex128", Kind: TypeKindBasic}

// TypeString is the basic string type
var TypeString = &Type{Name: "string", Kind: TypeKindBasic}

// TypeError is the error interface type
var TypeError = &Type{
	Name:       "error",
	Kind:       TypeKindInterface,
	Methods:    []*Method{},
	Fields:     []*Field{},
	Interfaces: []string{},
}

// basicTypes maps from type name strings to their global Type variables
var basicTypes = map[string]*Type{
	"bool":       TypeBool,
	"byte":       TypeByte,
	"rune":       TypeRune,
	"int":        TypeInt,
	"int8":       TypeInt8,
	"int16":      TypeInt16,
	"int32":      TypeInt32,
	"int64":      TypeInt64,
	"uint":       TypeUint,
	"uint8":      TypeUint8,
	"uint16":     TypeUint16,
	"uint32":     TypeUint32,
	"uint64":     TypeUint64,
	"uintptr":    TypeUintptr,
	"float32":    TypeFloat32,
	"float64":    TypeFloat64,
	"complex64":  TypeComplex64,
	"complex128": TypeComplex128,
	"string":     TypeString,
	"error":      TypeError,
}

// ============================================================================
// Type Construction Functions for Composite Types
// ============================================================================

// Pointer creates a pointer type to the given base type
func Pointer(baseType *Type) *Type {
	if baseType == nil {
		return nil
	}
	return &Type{
		Name: "*" + baseType.Name,
		Kind: TypeKindPointer,
		Elem: &TypeReference{
			Name: baseType.Name,
			Type: baseType,
		},
	}
}

// Slice creates a slice type of the given element type
func Slice(elemType *Type) *Type {
	if elemType == nil {
		return nil
	}
	return &Type{
		Name: "[]" + elemType.Name,
		Kind: TypeKindSlice,
		Elem: &TypeReference{
			Name: elemType.Name,
			Type: elemType,
		},
	}
}

// Array creates a fixed-size array type of the given element type and size
func Array(elemType *Type, size int) *Type {
	if elemType == nil {
		return nil
	}
	arraySize := ""
	if size > 0 {
		arraySize = strconv.Itoa(size)
	}
	return &Type{
		Name: "[" + arraySize + "]" + elemType.Name,
		Kind: TypeKindArray,
		Size: size,
		Elem: &TypeReference{
			Name: elemType.Name,
			Type: elemType,
		},
	}
}

// Map creates a map type with the given key and value types
func Map(keyType, valueType *Type) *Type {
	if keyType == nil || valueType == nil {
		return nil
	}
	return &Type{
		Name: "map[" + keyType.Name + "]" + valueType.Name,
		Kind: TypeKindMap,
		Key: &TypeReference{
			Name: keyType.Name,
			Type: keyType,
		},
		Value: &TypeReference{
			Name: valueType.Name,
			Type: valueType,
		},
	}
}

// Chan creates a channel type of the given element type with the specified direction
func Chan(elemType *Type, direction ChanDirection) *Type {
	if elemType == nil {
		return nil
	}

	name := "chan " + elemType.Name
	if direction == ChanSend {
		name = "chan<- " + elemType.Name
	} else if direction == ChanRecv {
		name = "<-chan " + elemType.Name
	}

	return &Type{
		Name:    name,
		Kind:    TypeKindChannel,
		ChanDir: direction,
		Elem: &TypeReference{
			Name: elemType.Name,
			Type: elemType,
		},
	}
}

// Struct creates a struct type with the given name and fields
func Struct(name string, fields []*Field) *Type {
	return &Type{
		Name:   name,
		Kind:   TypeKindStruct,
		Fields: fields,
	}
}

// Interface creates an interface type with the given name and methods
func Interface(name string, methods []*Method) *Type {
	return &Type{
		Name:    name,
		Kind:    TypeKindInterface,
		Methods: methods,
	}
}

// Alias creates a new type that is an alias for the given type
func Alias(name string, t *Type) *Type {
	return &Type{
		Name:         name,
		Kind:         TypeKindAlias,
		Underlying:   t.Underlying, // Preserve the original underlying type for aliases
		OriginalType: MakeTypeReference(t),
	}
}

// Typedef creates a new type that is a typedef of the given type
// Unlike Alias, Typedef creates a distinct type identity
func Typedef(name string, t *Type) *Type {
	return &Type{
		Name:         name,
		Kind:         TypeKindTypedef,
		Underlying:   t.Underlying, // Preserve the original underlying type for typedefs
		OriginalType: MakeTypeReference(t),
	}
}

// GetBasicType returns the global Type instance for a basic type name
// Returns nil if the name is not a basic type
func GetBasicType(name string) *Type {
	return basicTypes[name]
}

// IsBasicType returns true if the given name is a basic Go type
func IsBasicType(name string) bool {
	_, exists := basicTypes[name]
	return exists
}

// Value represents a value with its associated type information.
type Value struct {
	Type  *TypeReference `json:"type"`  // Type information of the value
	Value interface{}    `json:"value"` // The actual runtime value
}

// NewValue creates a new Value instance from a runtime value, extracting its type information.
func NewValue(v interface{}) *Value {
	return &Value{
		Type:  MakeTypeReference(FromRuntimeValue(v)),
		Value: v,
	}
}

// Variable represents a variable declaration with its initialization value
type Variable struct {
	Name  string `json:"name"`  // Variable name (e.g., "config", "userData")
	Doc   string `json:"doc"`   // Documentation comment for the variable
	Value *Value `json:"value"` // Variable initialization value
}

// NewVariable creates a new Variable instance with the given name and initial value from a runtime value
func NewVariable(name string, value interface{}) *Variable {
	return &Variable{
		Name:  name,
		Value: NewValue(value),
	}
}
