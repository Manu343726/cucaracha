package reflect

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewIndex tests the creation of a new index
func TestNewIndex(t *testing.T) {
	pkg := &Package{
		Name:      "testpkg",
		Path:      "github.com/test/testpkg",
		Types:     make(map[string]*Type),
		Functions: []*Function{},
		Constants: []*Constant{},
		Enums:     make(map[string]*Enum),
	}

	idx := NewIndexFromPackage(pkg)

	assert.NotNil(t, idx)
	assert.NotNil(t, idx.FieldToType)
	assert.NotNil(t, idx.MethodToType)
	assert.NotNil(t, idx.TypeToPackage)
	assert.NotNil(t, idx.FunctionToPackage)
	assert.NotNil(t, idx.ConstantToPackage)
	assert.NotNil(t, idx.EnumToPackage)
	assert.NotNil(t, idx.ConstantToEnum)
	assert.NotNil(t, idx.TypeToEnum)
}

// TestFieldParent tests finding the parent type of a field
func TestFieldParent(t *testing.T) {
	// Create a struct type with a field
	field := &Field{Name: "Id", Doc: "unique identifier"}
	typ := &Type{
		Name:   "User",
		Kind:   TypeKindStruct,
		Fields: []*Field{field},
	}

	pkg := &Package{
		Name:      "testpkg",
		Path:      "github.com/test/testpkg",
		Types:     map[string]*Type{"User": typ},
		Functions: []*Function{},
		Constants: []*Constant{},
		Enums:     make(map[string]*Enum),
	}

	idx := NewIndexFromPackage(pkg)

	// Test finding parent type
	parent := idx.FieldParent(field)
	assert.Equal(t, typ, parent)
	assert.Equal(t, "User", parent.Name)

	// Test nil field panics
	assert.Panics(t, func() { idx.FieldParent(nil) })

	// Test non-indexed field returns nil
	otherField := &Field{Name: "Name"}
	assert.Nil(t, idx.FieldParent(otherField))
}

// TestMethodParent tests finding the parent type of a method
func TestMethodParent(t *testing.T) {
	// Create a struct type with a method
	method := &Method{
		Name:      "String",
		Signature: "() string",
	}
	typ := &Type{
		Name:    "User",
		Kind:    TypeKindStruct,
		Methods: []*Method{method},
	}

	pkg := &Package{
		Name:      "testpkg",
		Path:      "github.com/test/testpkg",
		Types:     map[string]*Type{"User": typ},
		Functions: []*Function{},
		Constants: []*Constant{},
		Enums:     make(map[string]*Enum),
	}

	idx := NewIndexFromPackage(pkg)

	// Test finding parent type
	parent := idx.MethodParent(method)
	assert.Equal(t, typ, parent)
	assert.Equal(t, "User", parent.Name)

	// Test nil method panics
	assert.Panics(t, func() { idx.MethodParent(nil) })

	// Test non-indexed method returns nil
	otherMethod := &Method{Name: "Name"}
	assert.Nil(t, idx.MethodParent(otherMethod))
}

// TestTypePackage tests finding the package of a type
func TestTypePackage(t *testing.T) {
	typ := &Type{Name: "User", Kind: TypeKindStruct}
	pkg := &Package{
		Name:      "testpkg",
		Path:      "github.com/test/testpkg",
		Types:     map[string]*Type{"User": typ},
		Functions: []*Function{},
		Constants: []*Constant{},
		Enums:     make(map[string]*Enum),
	}

	idx := NewIndexFromPackage(pkg)

	// Test finding package
	result := idx.TypePackage(typ)
	assert.Equal(t, pkg, result)
	assert.Equal(t, "testpkg", result.Name)

	// Test nil type panics
	assert.Panics(t, func() { idx.TypePackage(nil) })

	// Test non-indexed type returns nil
	otherType := &Type{Name: "Other"}
	assert.Nil(t, idx.TypePackage(otherType))
}

// TestFunctionPackage tests finding the package of a function
func TestFunctionPackage(t *testing.T) {
	fn := &Function{Name: "NewUser", Package: "testpkg"}
	pkg := &Package{
		Name:      "testpkg",
		Path:      "github.com/test/testpkg",
		Types:     make(map[string]*Type),
		Functions: []*Function{fn},
		Constants: []*Constant{},
		Enums:     make(map[string]*Enum),
	}

	idx := NewIndexFromPackage(pkg)

	// Test finding package
	result := idx.FunctionPackage(fn)
	assert.Equal(t, pkg, result)
	assert.Equal(t, "testpkg", result.Name)

	// Test nil function panics
	assert.Panics(t, func() { idx.FunctionPackage(nil) })

	// Test non-indexed function returns nil
	otherFn := &Function{Name: "Other"}
	assert.Nil(t, idx.FunctionPackage(otherFn))
}

// TestConstantPackage tests finding the package of a constant
func TestConstantPackage(t *testing.T) {
	const_ := &Constant{Name: "MaxSize", Doc: "maximum size"}
	pkg := &Package{
		Name:      "testpkg",
		Path:      "github.com/test/testpkg",
		Types:     make(map[string]*Type),
		Functions: []*Function{},
		Constants: []*Constant{const_},
		Enums:     make(map[string]*Enum),
	}

	idx := NewIndexFromPackage(pkg)

	// Test finding package
	result := idx.ConstantPackage(const_)
	assert.Equal(t, pkg, result)
	assert.Equal(t, "testpkg", result.Name)

	// Test nil constant panics
	assert.Panics(t, func() { idx.ConstantPackage(nil) })

	// Test non-indexed constant returns nil
	otherConst := &Constant{Name: "Other"}
	assert.Nil(t, idx.ConstantPackage(otherConst))
}

// TestEnumPackage tests finding the package of an enum
func TestEnumPackage(t *testing.T) {
	enumType := &Type{Name: "Color", Kind: TypeKindTypedef}
	enum := &Enum{
		Type: &TypeReference{Name: "Color", Type: enumType},
	}

	pkg := &Package{
		Name:      "testpkg",
		Path:      "github.com/test/testpkg",
		Types:     map[string]*Type{"Color": enumType},
		Functions: []*Function{},
		Constants: []*Constant{},
		Enums:     map[string]*Enum{"Color": enum},
	}

	idx := NewIndexFromPackage(pkg)

	// Test finding package
	result := idx.EnumPackage(enum)
	assert.Equal(t, pkg, result)
	assert.Equal(t, "testpkg", result.Name)

	// Test nil enum panics
	assert.Panics(t, func() { idx.EnumPackage(nil) })

	// Test non-indexed enum returns nil
	otherEnum := &Enum{}
	assert.Nil(t, idx.EnumPackage(otherEnum))
}

// TestConstantEnum tests finding the enum of an enum value
func TestConstantEnum(t *testing.T) {
	enumType := &Type{Name: "Status", Kind: TypeKindTypedef}
	const1 := &Constant{Name: "StatusActive"}
	const2 := &Constant{Name: "StatusInactive"}

	enum := &Enum{
		Type:   &TypeReference{Name: "Status", Type: enumType},
		Values: []*Constant{const1, const2},
	}

	pkg := &Package{
		Name:      "testpkg",
		Path:      "github.com/test/testpkg",
		Types:     map[string]*Type{"Status": enumType},
		Functions: []*Function{},
		Constants: []*Constant{const1, const2},
		Enums:     map[string]*Enum{"Status": enum},
	}

	idx := NewIndexFromPackage(pkg)

	// Test finding enum for value 1
	result1 := idx.ConstantEnum(const1)
	assert.Equal(t, enum, result1)

	// Test finding enum for value 2
	result2 := idx.ConstantEnum(const2)
	assert.Equal(t, enum, result2)

	// Test nil constant returns nil
	assert.Nil(t, idx.ConstantEnum(nil))

	// Test non-enum value returns nil
	otherConst := &Constant{Name: "MaxSize"}
	assert.Nil(t, idx.ConstantEnum(otherConst))
}

// TestTypeEnum tests finding the enum for an enum type
func TestTypeEnum(t *testing.T) {
	enumType := &Type{Name: "Priority", Kind: TypeKindTypedef}
	enum := &Enum{
		Type: &TypeReference{Name: "Priority", Type: enumType},
	}

	pkg := &Package{
		Name:      "testpkg",
		Path:      "github.com/test/testpkg",
		Types:     map[string]*Type{"Priority": enumType},
		Functions: []*Function{},
		Constants: []*Constant{},
		Enums:     map[string]*Enum{"Priority": enum},
	}

	idx := NewIndexFromPackage(pkg)

	// Test finding enum for type
	result := idx.TypeEnum(enumType)
	assert.Equal(t, enum, result)

	// Test nil type returns nil
	assert.Nil(t, idx.TypeEnum(nil))

	// Test non-enum type returns nil
	otherType := &Type{Name: "User", Kind: TypeKindStruct}
	assert.Nil(t, idx.TypeEnum(otherType))
}

// TestGetFieldsInType tests retrieving all fields of a type
func TestGetFieldsInType(t *testing.T) {
	field1 := &Field{Name: "Id"}
	field2 := &Field{Name: "Name"}
	typ := &Type{
		Name:   "User",
		Kind:   TypeKindStruct,
		Fields: []*Field{field1, field2},
	}

	pkg := &Package{
		Name:      "testpkg",
		Path:      "github.com/test/testpkg",
		Types:     map[string]*Type{"User": typ},
		Functions: []*Function{},
		Constants: []*Constant{},
		Enums:     make(map[string]*Enum),
	}

	idx := NewIndexFromPackage(pkg)

	fields := idx.GetFieldsInType(typ)
	require.NotNil(t, fields)
	assert.Len(t, fields, 2)
	assert.Equal(t, field1, fields[0])
	assert.Equal(t, field2, fields[1])

	// Test nil type panics
	assert.Panics(t, func() { idx.GetFieldsInType(nil) })
}

// TestGetMethodsInType tests retrieving all methods of a type
func TestGetMethodsInType(t *testing.T) {
	method1 := &Method{Name: "String"}
	method2 := &Method{Name: "MarshalJSON"}
	typ := &Type{
		Name:    "User",
		Kind:    TypeKindStruct,
		Methods: []*Method{method1, method2},
	}

	pkg := &Package{
		Name:      "testpkg",
		Path:      "github.com/test/testpkg",
		Types:     map[string]*Type{"User": typ},
		Functions: []*Function{},
		Constants: []*Constant{},
		Enums:     make(map[string]*Enum),
	}

	idx := NewIndexFromPackage(pkg)

	methods := idx.GetMethodsInType(typ)
	require.NotNil(t, methods)
	assert.Len(t, methods, 2)
	assert.Equal(t, method1, methods[0])
	assert.Equal(t, method2, methods[1])

	// Test nil type panics
	assert.Panics(t, func() { idx.GetMethodsInType(nil) })
}

// TestGetEnumValues tests retrieving all values of an enum
func TestGetEnumValues(t *testing.T) {
	enumType := &Type{Name: "Status", Kind: TypeKindTypedef}
	const1 := &Constant{Name: "StatusActive"}
	const2 := &Constant{Name: "StatusInactive"}

	enum := &Enum{
		Type:   &TypeReference{Name: "Status", Type: enumType},
		Values: []*Constant{const1, const2},
	}

	pkg := &Package{
		Name:      "testpkg",
		Path:      "github.com/test/testpkg",
		Types:     map[string]*Type{"Status": enumType},
		Functions: []*Function{},
		Constants: []*Constant{const1, const2},
		Enums:     map[string]*Enum{"Status": enum},
	}

	idx := NewIndexFromPackage(pkg)

	values := idx.GetEnumValues(enum)
	require.NotNil(t, values)
	assert.Len(t, values, 2)
	assert.Equal(t, const1, values[0])
	assert.Equal(t, const2, values[1])

	// Test nil enum panics
	assert.Panics(t, func() { idx.GetEnumValues(nil) })
}

// TestComplexPackageIndexing tests indexing a complex package with multiple entities
func TestComplexPackageIndexing(t *testing.T) {
	// Create multiple types with fields and methods
	field1 := &Field{Name: "Id"}
	field2 := &Field{Name: "Name"}
	method1 := &Method{Name: "String"}

	userType := &Type{
		Name:    "User",
		Kind:    TypeKindStruct,
		Fields:  []*Field{field1, field2},
		Methods: []*Method{method1},
	}

	// Create functions
	fn := &Function{Name: "NewUser", Package: "testpkg"}

	// Create constants
	const1 := &Constant{Name: "MaxSize"}
	const2 := &Constant{Name: "StatusActive"}

	// Create enum
	statusType := &Type{Name: "Status", Kind: TypeKindTypedef}
	const3 := &Constant{Name: "StatusInactive"}
	enum := &Enum{
		Type:   &TypeReference{Name: "Status", Type: statusType},
		Values: []*Constant{const2, const3},
	}

	pkg := &Package{
		Name:      "testpkg",
		Path:      "github.com/test/testpkg",
		Types:     map[string]*Type{"User": userType, "Status": statusType},
		Functions: []*Function{fn},
		Constants: []*Constant{const1, const2, const3},
		Enums:     map[string]*Enum{"Status": enum},
	}

	idx := NewIndexFromPackage(pkg)

	// Verify all relationships
	assert.Equal(t, userType, idx.FieldParent(field1))
	assert.Equal(t, userType, idx.FieldParent(field2))
	assert.Equal(t, userType, idx.MethodParent(method1))
	assert.Equal(t, pkg, idx.TypePackage(userType))
	assert.Equal(t, pkg, idx.TypePackage(statusType))
	assert.Equal(t, pkg, idx.FunctionPackage(fn))
	assert.Equal(t, pkg, idx.ConstantPackage(const1))
	assert.Equal(t, pkg, idx.ConstantPackage(const2))
	assert.Equal(t, pkg, idx.ConstantPackage(const3))
	assert.Equal(t, enum, idx.ConstantEnum(const2))
	assert.Equal(t, enum, idx.ConstantEnum(const3))
	assert.Nil(t, idx.ConstantEnum(const1)) // Not an enum value
	assert.Equal(t, enum, idx.TypeEnum(statusType))
	assert.Nil(t, idx.TypeEnum(userType)) // Not an enum type
}

// TestEmptyPackageIndexing tests indexing an empty package
func TestEmptyPackageIndexing(t *testing.T) {
	pkg := &Package{
		Name:      "empty",
		Path:      "github.com/test/empty",
		Types:     make(map[string]*Type),
		Functions: []*Function{},
		Constants: []*Constant{},
		Enums:     make(map[string]*Enum),
	}

	idx := NewIndexFromPackage(pkg)
	assert.NotNil(t, idx)

	// All queries should return nil for non-existent entities
	assert.Nil(t, idx.FieldParent(&Field{}))
	assert.Nil(t, idx.MethodParent(&Method{}))
	assert.Nil(t, idx.TypePackage(&Type{}))
	assert.Nil(t, idx.FunctionPackage(&Function{}))
	assert.Nil(t, idx.ConstantPackage(&Constant{}))
	assert.Nil(t, idx.EnumPackage(&Enum{}))
}

// TestSearchMethod tests the Method search by qualified name
func TestSearchMethod(t *testing.T) {
	// Create a type with a method
	method := &Method{
		Name:      "Debug",
		Signature: "(string) error",
	}
	typ := &Type{
		Name:    "Logger",
		Kind:    TypeKindStruct,
		Methods: []*Method{method},
	}

	pkg := &Package{
		Name:      "logging",
		Path:      "github.com/Manu343726/cucaracha/pkg/utils/logging",
		Types:     map[string]*Type{"Logger": typ},
		Functions: []*Function{},
		Constants: []*Constant{},
		Enums:     make(map[string]*Enum),
	}

	idx := NewIndexFromPackage(pkg)

	// Test finding method by qualified name
	found := idx.Method("github.com/Manu343726/cucaracha/pkg/utils/logging/Logger/Debug")
	require.NotNil(t, found)
	assert.Equal(t, "Debug", found.Name)
	assert.Equal(t, "(string) error", found.Signature)

	// Test non-existent method
	assert.Nil(t, idx.Method("github.com/Manu343726/cucaracha/pkg/utils/logging/Logger/NonExistent"))

	// Test invalid format
	assert.Nil(t, idx.Method("invalid"))
}

// TestSearchField tests the Field search by qualified name
func TestSearchField(t *testing.T) {
	// Create a type with a field
	field := &Field{
		Name: "writer",
		Doc:  "output writer",
	}
	typ := &Type{
		Name:   "Logger",
		Kind:   TypeKindStruct,
		Fields: []*Field{field},
	}

	pkg := &Package{
		Name:      "logging",
		Path:      "github.com/Manu343726/cucaracha/pkg/utils/logging",
		Types:     map[string]*Type{"Logger": typ},
		Functions: []*Function{},
		Constants: []*Constant{},
		Enums:     make(map[string]*Enum),
	}

	idx := NewIndexFromPackage(pkg)

	// Test finding field by qualified name
	found := idx.Field("github.com/Manu343726/cucaracha/pkg/utils/logging/Logger/writer")
	require.NotNil(t, found)
	assert.Equal(t, "writer", found.Name)

	// Test non-existent field
	assert.Nil(t, idx.Field("github.com/Manu343726/cucaracha/pkg/utils/logging/Logger/NonExistent"))

	// Test invalid format
	assert.Nil(t, idx.Field("invalid"))
}

// TestSearchFunction tests the Function search by qualified name
func TestSearchFunction(t *testing.T) {
	// Create a function
	fn := &Function{
		Name:      "NewLogger",
		Signature: "(io.Writer) *Logger",
		Doc:       "creates a new logger",
	}

	pkg := &Package{
		Name:      "logging",
		Path:      "github.com/Manu343726/cucaracha/pkg/utils/logging",
		Types:     make(map[string]*Type),
		Functions: []*Function{fn},
		Constants: []*Constant{},
		Enums:     make(map[string]*Enum),
	}

	idx := NewIndexFromPackage(pkg)

	// Test finding function by qualified name
	found := idx.Function("github.com/Manu343726/cucaracha/pkg/utils/logging/NewLogger")
	require.NotNil(t, found)
	assert.Equal(t, "NewLogger", found.Name)

	// Test non-existent function
	assert.Nil(t, idx.Function("github.com/Manu343726/cucaracha/pkg/utils/logging/NonExistent"))

	// Test invalid format
	assert.Nil(t, idx.Function("invalid"))
}

// TestSearchType tests the Type search by qualified name
func TestSearchType(t *testing.T) {
	// Create a type
	typ := &Type{
		Name: "Logger",
		Kind: TypeKindStruct,
	}

	pkg := &Package{
		Name:      "logging",
		Path:      "github.com/Manu343726/cucaracha/pkg/utils/logging",
		Types:     map[string]*Type{"Logger": typ},
		Functions: []*Function{},
		Constants: []*Constant{},
		Enums:     make(map[string]*Enum),
	}

	idx := NewIndexFromPackage(pkg)

	// Test finding type by qualified name
	found := idx.Type("github.com/Manu343726/cucaracha/pkg/utils/logging/Logger")
	require.NotNil(t, found)
	assert.Equal(t, "Logger", found.Name)

	// Test non-existent type
	assert.Nil(t, idx.Type("github.com/Manu343726/cucaracha/pkg/utils/logging/NonExistent"))

	// Test invalid format
	assert.Nil(t, idx.Type("invalid"))
}

// TestSearchConstant tests the Constant search by qualified name
func TestSearchConstant(t *testing.T) {
	// Create a constant
	const_ := &Constant{
		Name:  "DebugLevel",
		Value: &Value{Value: "0"},
	}

	pkg := &Package{
		Name:      "logging",
		Path:      "github.com/Manu343726/cucaracha/pkg/utils/logging",
		Types:     make(map[string]*Type),
		Functions: []*Function{},
		Constants: []*Constant{const_},
		Enums:     make(map[string]*Enum),
	}

	idx := NewIndexFromPackage(pkg)

	// Test finding constant by qualified name (2-part format)
	found := idx.Constant("github.com/Manu343726/cucaracha/pkg/utils/logging/DebugLevel")
	require.NotNil(t, found)
	assert.Equal(t, "DebugLevel", found.Name)

	// Test non-existent constant
	assert.Nil(t, idx.Constant("github.com/Manu343726/cucaracha/pkg/utils/logging/NonExistent"))

	// Test invalid format
	assert.Nil(t, idx.Constant("invalid"))
}

// TestSearchEnumValue tests searching for enum values by qualified name
func TestSearchEnumValue(t *testing.T) {
	// Create an enum with values
	val1 := &Constant{Name: "Active", Value: &Value{Value: "0"}}
	val2 := &Constant{Name: "Inactive", Value: &Value{Value: "1"}}

	enumType := &Type{Name: "Status", Kind: TypeKindTypedef}
	enum := &Enum{
		Type:   &TypeReference{Name: "Status", Type: enumType},
		Values: []*Constant{val1, val2},
	}

	pkg := &Package{
		Name:      "utils",
		Path:      "github.com/Manu343726/cucaracha/pkg/utils",
		Types:     map[string]*Type{"Status": enumType},
		Functions: []*Function{},
		Constants: []*Constant{},
		Enums:     map[string]*Enum{"Status": enum},
	}

	idx := NewIndexFromPackage(pkg)

	// Test finding enum value with EnumValue method (3-part format)
	found := idx.EnumValue("github.com/Manu343726/cucaracha/pkg/utils/Status/Active")
	require.NotNil(t, found)
	assert.Equal(t, "Active", found.Name)

	found = idx.EnumValue("github.com/Manu343726/cucaracha/pkg/utils/Status/Inactive")
	require.NotNil(t, found)
	assert.Equal(t, "Inactive", found.Name)

	// Test non-existent enum value
	assert.Nil(t, idx.EnumValue("github.com/Manu343726/cucaracha/pkg/utils/Status/NonExistent"))

	// Test non-existent enum
	assert.Nil(t, idx.EnumValue("github.com/Manu343726/cucaracha/pkg/utils/NonExistent/Value"))

	// Test invalid format (too few parts)
	assert.Nil(t, idx.EnumValue("invalid"))

	// Also test that Constant method can find enum values with 3-part format
	found = idx.Constant("github.com/Manu343726/cucaracha/pkg/utils/Status/Active")
	require.NotNil(t, found)
	assert.Equal(t, "Active", found.Name)
}

// TestSearchPackage tests the Package search by path
func TestSearchPackage(t *testing.T) {
	pkg := &Package{
		Name:      "logging",
		Path:      "github.com/Manu343726/cucaracha/pkg/utils/logging",
		Types:     make(map[string]*Type),
		Functions: []*Function{},
		Constants: []*Constant{},
		Enums:     make(map[string]*Enum),
	}

	idx := NewIndexFromPackage(pkg)

	// Test finding package by path
	found := idx.Package("github.com/Manu343726/cucaracha/pkg/utils/logging")
	require.NotNil(t, found)
	assert.Equal(t, "logging", found.Name)
	assert.Equal(t, "github.com/Manu343726/cucaracha/pkg/utils/logging", found.Path)

	// Test non-existent package
	assert.Nil(t, idx.Package("non.existent/package"))
}
