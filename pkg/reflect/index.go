package reflect

import (
	"strings"
)

// Index maintains parent/child relationships between reflect entities across one or more packages.
// It allows efficient typed queries to find the parent or package of any entity
// without adding circular references to the entity types themselves.
// When used for multi-package indexing, call AddPackage to add packages on demand.
type Index struct {
	// Packages stores all packages indexed by this index, keyed by package path
	Packages map[string]*Package

	// Maps fields to their parent types
	FieldToType map[*Field]*Type

	// Maps methods to their parent types
	MethodToType map[*Method]*Type

	// Maps types to their parent packages
	TypeToPackage map[*Type]*Package

	// Maps functions to their parent packages
	FunctionToPackage map[*Function]*Package

	// Maps constants to their parent packages
	ConstantToPackage map[*Constant]*Package

	// Maps enums to their parent packages
	EnumToPackage map[*Enum]*Package

	// Maps enum values (constants) to their enums
	ConstantToEnum map[*Constant]*Enum

	// Maps enum types to their enums
	TypeToEnum map[*Type]*Enum
}

// NewIndex creates a new empty index.
// Use AddPackage to add packages to the index.
func NewIndex() *Index {
	return &Index{
		Packages:          make(map[string]*Package),
		FieldToType:       make(map[*Field]*Type),
		MethodToType:      make(map[*Method]*Type),
		TypeToPackage:     make(map[*Type]*Package),
		FunctionToPackage: make(map[*Function]*Package),
		ConstantToPackage: make(map[*Constant]*Package),
		EnumToPackage:     make(map[*Enum]*Package),
		ConstantToEnum:    make(map[*Constant]*Enum),
		TypeToEnum:        make(map[*Type]*Enum),
	}
}

// AddPackage adds a package and all its entities to the index.
// This allows building the index incrementally as packages are parsed.
// If the package is already in the index (by path), it will be replaced.
func (idx *Index) AddPackage(pkg *Package) {
	if pkg == nil {
		return
	}

	idx.Packages[pkg.Path] = pkg

	// Index all types and their fields/methods
	for _, typ := range pkg.Types {
		idx.TypeToPackage[typ] = pkg

		// Index all fields in this type
		for _, field := range typ.Fields {
			idx.FieldToType[field] = typ
		}

		// Index all methods on this type
		for _, method := range typ.Methods {
			idx.MethodToType[method] = typ
		}
	}

	// Index all functions
	for _, fn := range pkg.Functions {
		idx.FunctionToPackage[fn] = pkg
	}

	// Index all constants
	for _, const_ := range pkg.Constants {
		idx.ConstantToPackage[const_] = pkg
	}

	// Index all enums
	for _, enum := range pkg.Enums {
		idx.EnumToPackage[enum] = pkg

		// Index the enum type
		if enum.Type != nil && enum.Type.Type != nil {
			idx.TypeToEnum[enum.Type.Type] = enum
		}

		// Index the enum values
		for _, val := range enum.Values {
			idx.ConstantToEnum[val] = enum
			idx.ConstantToPackage[val] = pkg
		}
	}
}

// NewIndexFromPackage creates a new index initialized with a single package.
// This is equivalent to NewIndex followed by AddPackage.
func NewIndexFromPackage(pkg *Package) *Index {
	idx := NewIndex()
	idx.AddPackage(pkg)
	return idx
}

// FieldParent returns the Type (struct or interface) that contains the given field.
// Returns nil if the field is not in the index.
func (idx *Index) FieldParent(f *Field) *Type {
	if f == nil {
		panic("FieldParent: field cannot be nil")
	}
	return idx.FieldToType[f]
}

// MethodParent returns the Type (struct or interface) that contains the given method.
// Returns nil if the method is not in the index.
func (idx *Index) MethodParent(m *Method) *Type {
	if m == nil {
		panic("MethodParent: method cannot be nil")
	}
	return idx.MethodToType[m]
}

// TypePackage returns the Package that contains the given type.
// Returns nil if the type is not in the index.
func (idx *Index) TypePackage(t *Type) *Package {
	if t == nil {
		panic("TypePackage: type cannot be nil")
	}
	return idx.TypeToPackage[t]
}

// FunctionPackage returns the Package that contains the given function.
// Returns nil if the function is not in the index.
func (idx *Index) FunctionPackage(f *Function) *Package {
	if f == nil {
		panic("FunctionPackage: function cannot be nil")
	}
	return idx.FunctionToPackage[f]
}

// ConstantPackage returns the Package that contains the given constant.
// Returns nil if the constant is not in the index.
func (idx *Index) ConstantPackage(c *Constant) *Package {
	if c == nil {
		panic("ConstantPackage: constant cannot be nil")
	}
	return idx.ConstantToPackage[c]
}

// EnumPackage returns the Package that contains the given enum.
// Returns nil if the enum is not in the index.
func (idx *Index) EnumPackage(e *Enum) *Package {
	if e == nil {
		panic("EnumPackage: enum cannot be nil")
	}
	return idx.EnumToPackage[e]
}

// ConstantEnum returns the Enum that contains the given constant (if it's an enum value).
// Returns nil if the constant is not an enum value or is not in the index.
func (idx *Index) ConstantEnum(c *Constant) *Enum {
	return idx.ConstantToEnum[c]
}

// TypeEnum returns the Enum associated with the given type (if the type is an enum).
// Returns nil if the type is not an enum or is not in the index.
func (idx *Index) TypeEnum(t *Type) *Enum {
	return idx.TypeToEnum[t]
}

// GetFieldsInType returns all fields of the given type.
// This is a convenience method that avoids copying the fields slice directly.
func (idx *Index) GetFieldsInType(t *Type) []*Field {
	if t == nil {
		panic("GetFieldsInType: type cannot be nil")
	}
	return t.Fields
}

// GetMethodsInType returns all methods of the given type.
// This is a convenience method that avoids copying the methods slice directly.
func (idx *Index) GetMethodsInType(t *Type) []*Method {
	if t == nil {
		panic("GetMethodsInType: type cannot be nil")
	}
	return t.Methods
}

// GetEnumValues returns all constant values in the given enum.
// This is a convenience method that avoids copying the enum values slice directly.
func (idx *Index) GetEnumValues(e *Enum) []*Constant {
	if e == nil {
		panic("GetEnumValues: enum cannot be nil")
	}
	return e.Values
}

// Method returns the method with the given fully qualified name.
// The name format is: PackagePath/TypeName/MethodName
// Example: "github.com/Manu343726/cucaracha/pkg/utils/logging/Logger/Debug"
// Returns nil if the method is not found.
func (idx *Index) Method(qualifiedName string) *Method {
	parts := strings.Split(qualifiedName, "/")
	if len(parts) < 3 {
		return nil
	}

	// Last part is method name
	methodName := parts[len(parts)-1]
	// Second-to-last part is type name
	typeName := parts[len(parts)-2]
	// Everything else is the package path
	packagePath := strings.Join(parts[:len(parts)-2], "/")

	pkg, exists := idx.Packages[packagePath]
	if !exists {
		return nil
	}

	typ, exists := pkg.Types[typeName]
	if !exists {
		return nil
	}

	// Search for the method by name
	for _, method := range typ.Methods {
		if method.Name == methodName {
			return method
		}
	}

	return nil
}

// Field returns the field with the given fully qualified name.
// The name format is: PackagePath/TypeName/FieldName
// Example: "github.com/Manu343726/cucaracha/pkg/utils/logging/Logger/writer"
// Returns nil if the field is not found.
func (idx *Index) Field(qualifiedName string) *Field {
	parts := strings.Split(qualifiedName, "/")
	if len(parts) < 3 {
		return nil
	}

	// Last part is field name
	fieldName := parts[len(parts)-1]
	// Second-to-last part is type name
	typeName := parts[len(parts)-2]
	// Everything else is the package path
	packagePath := strings.Join(parts[:len(parts)-2], "/")

	pkg, exists := idx.Packages[packagePath]
	if !exists {
		return nil
	}

	typ, exists := pkg.Types[typeName]
	if !exists {
		return nil
	}

	// Search for the field by name
	for _, field := range typ.Fields {
		if field.Name == fieldName {
			return field
		}
	}

	return nil
}

// Function returns the function with the given fully qualified name.
// The name format is: PackagePath/FunctionName
// Example: "github.com/Manu343726/cucaracha/pkg/utils/logging/NewLogger"
// Returns nil if the function is not found.
func (idx *Index) Function(qualifiedName string) *Function {
	parts := strings.Split(qualifiedName, "/")
	if len(parts) < 2 {
		return nil
	}

	// Last part is function name
	functionName := parts[len(parts)-1]
	// Everything else is the package path
	packagePath := strings.Join(parts[:len(parts)-1], "/")

	pkg, exists := idx.Packages[packagePath]
	if !exists {
		return nil
	}

	// Search for the function by name
	for _, fn := range pkg.Functions {
		if fn.Name == functionName {
			return fn
		}
	}

	return nil
}

// Type returns the type with the given fully qualified name.
// The name format is: PackagePath/TypeName
// Example: "github.com/Manu343726/cucaracha/pkg/utils/logging/Logger"
// Returns nil if the type is not found.
func (idx *Index) Type(qualifiedName string) *Type {
	parts := strings.Split(qualifiedName, "/")
	if len(parts) < 2 {
		return nil
	}

	// Last part is type name
	typeName := parts[len(parts)-1]
	// Everything else is the package path
	packagePath := strings.Join(parts[:len(parts)-1], "/")

	pkg, exists := idx.Packages[packagePath]
	if !exists {
		return nil
	}

	typ, exists := pkg.Types[typeName]
	if exists {
		return typ
	}

	return nil
}

// Constant returns the constant with the given fully qualified name.
// The name format can be:
//   - PackagePath/ConstantName for package-level constants
//     Example: "github.com/Manu343726/cucaracha/pkg/utils/logging/DebugLevel"
//   - PackagePath/EnumName/ConstantName for enum values
//     Example: "github.com/Manu343726/cucaracha/pkg/utils/logging/Status/Active"
//
// Returns nil if the constant is not found.
func (idx *Index) Constant(qualifiedName string) *Constant {
	parts := strings.Split(qualifiedName, "/")
	if len(parts) < 2 {
		return nil
	}

	pkg, exists := idx.Packages[strings.Join(parts[:len(parts)-1], "/")]
	if !exists && len(parts) >= 3 {
		// Try the 3-part format (package/enum/value)
		packagePath := strings.Join(parts[:len(parts)-2], "/")
		pkg, exists = idx.Packages[packagePath]
		if !exists {
			return nil
		}

		enumName := parts[len(parts)-2]
		constantName := parts[len(parts)-1]

		enum, exists := pkg.Enums[enumName]
		if !exists {
			return nil
		}

		// Search for the constant in the enum values
		for _, const_ := range enum.Values {
			if const_.Name == constantName {
				return const_
			}
		}

		return nil
	}

	if !exists {
		return nil
	}

	// 2-part format (package/constant)
	constantName := parts[len(parts)-1]

	// Search for the constant by name
	for _, const_ := range pkg.Constants {
		if const_.Name == constantName {
			return const_
		}
	}

	return nil
}

// EnumValue returns an enum constant (value) with the given fully qualified name.
// The name format is: PackagePath/EnumName/ConstantName
// Example: "github.com/Manu343726/cucaracha/pkg/utils/logging/Status/Active"
// This is equivalent to Constant with 3-part format but more explicit.
// Returns nil if the enum value is not found.
func (idx *Index) EnumValue(qualifiedName string) *Constant {
	parts := strings.Split(qualifiedName, "/")
	if len(parts) < 3 {
		return nil
	}

	packagePath := strings.Join(parts[:len(parts)-2], "/")
	enumName := parts[len(parts)-2]
	constantName := parts[len(parts)-1]

	pkg, exists := idx.Packages[packagePath]
	if !exists {
		return nil
	}

	enum, exists := pkg.Enums[enumName]
	if !exists {
		return nil
	}

	// Search for the constant in the enum values
	for _, const_ := range enum.Values {
		if const_.Name == constantName {
			return const_
		}
	}

	return nil
}

// Package returns the package with the given path.
// Returns nil if the package is not in the index.
func (idx *Index) Package(packagePath string) *Package {
	return idx.Packages[packagePath]
}
