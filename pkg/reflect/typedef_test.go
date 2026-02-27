package reflect

import (
	refl "reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTypedefOfBasicInt tests that typedefs of basic int type are handled correctly
func TestTypedefOfBasicInt(t *testing.T) {
	type MyInt int

	typ := FromRuntimeType(refl.TypeOf(MyInt(0)))

	assert.NotNil(t, typ)
	assert.Equal(t, TypeKindTypedef, typ.Kind)
	// The typedef should preserve its custom name
	assert.Contains(t, typ.Name, "MyInt")
	// The typedef should have the underlying type reference
	assert.NotNil(t, typ.OriginalType)
	assert.Equal(t, "int", typ.OriginalType.Name)
}

// TestTypedefOfBasicString tests that typedefs of basic string type are handled correctly
func TestTypedefOfBasicString(t *testing.T) {
	type MyString string

	typ := FromRuntimeType(refl.TypeOf(MyString("")))

	assert.NotNil(t, typ)
	assert.Equal(t, TypeKindTypedef, typ.Kind)
	// The typedef should preserve its custom name
	assert.Contains(t, typ.Name, "MyString")
	// The typedef should have the underlying type reference
	assert.NotNil(t, typ.OriginalType)
	assert.Equal(t, "string", typ.OriginalType.Name)
}

// TestTypedefOfBasicFloat tests that typedefs of basic float type are handled correctly
func TestTypedefOfBasicFloat(t *testing.T) {
	type MyFloat float64

	typ := FromRuntimeType(refl.TypeOf(MyFloat(0)))

	assert.NotNil(t, typ)
	assert.Equal(t, TypeKindTypedef, typ.Kind)
	// The typedef should preserve its custom name
	assert.Contains(t, typ.Name, "MyFloat")
	// The typedef should have the underlying type reference
	assert.NotNil(t, typ.OriginalType)
	assert.Equal(t, "float64", typ.OriginalType.Name)
}

// TestTypedefOfBasicBool tests that typedefs of basic bool type are handled correctly
func TestTypedefOfBasicBool(t *testing.T) {
	type MyBool bool

	typ := FromRuntimeType(refl.TypeOf(MyBool(false)))

	assert.NotNil(t, typ)
	assert.Equal(t, TypeKindTypedef, typ.Kind)
	// The typedef should preserve its custom name
	assert.Contains(t, typ.Name, "MyBool")
	// The typedef should have the underlying type reference
	assert.NotNil(t, typ.OriginalType)
	assert.Equal(t, "bool", typ.OriginalType.Name)
}

// TestTypedefOfInt64 tests that typedefs of int64 are handled correctly
func TestTypedefOfInt64(t *testing.T) {
	type ID int64

	typ := FromRuntimeType(refl.TypeOf(ID(0)))

	assert.NotNil(t, typ)
	assert.Equal(t, TypeKindTypedef, typ.Kind)
	// The typedef should preserve its custom name
	assert.Contains(t, typ.Name, "ID")
	// The typedef should have the underlying type reference
	assert.NotNil(t, typ.OriginalType)
	assert.Equal(t, "int64", typ.OriginalType.Name)
}

// TestTypedefUintptr tests that typedefs of uintptr are handled correctly
func TestTypedefUintptr(t *testing.T) {
	type Pointer uintptr

	typ := FromRuntimeType(refl.TypeOf(Pointer(0)))

	assert.NotNil(t, typ)
	assert.Equal(t, TypeKindTypedef, typ.Kind)
	// The typedef should preserve its custom name
	assert.Contains(t, typ.Name, "Pointer")
	// The typedef should have the underlying type reference
	assert.NotNil(t, typ.OriginalType)
	assert.Equal(t, "uintptr", typ.OriginalType.Name)
}

// TestMultipleTypedefsOfSameBase tests that multiple typedefs of the same base type work correctly
func TestMultipleTypedefsOfSameBase(t *testing.T) {
	type TypeA int
	type TypeB int

	typA := FromRuntimeType(refl.TypeOf(TypeA(0)))
	typB := FromRuntimeType(refl.TypeOf(TypeB(0)))

	assert.NotNil(t, typA)
	assert.NotNil(t, typB)

	// Both should be typedef types
	assert.Equal(t, TypeKindTypedef, typA.Kind)
	assert.Equal(t, TypeKindTypedef, typB.Kind)

	// Both should have original type pointing to int
	assert.NotNil(t, typA.OriginalType)
	assert.NotNil(t, typB.OriginalType)
	assert.Equal(t, "int", typA.OriginalType.Name)
	assert.Equal(t, "int", typB.OriginalType.Name)
}

// TestBuiltInBasicTypeNotModified tests that built-in basic types are still handled correctly
func TestBuiltInBasicTypeNotModified(t *testing.T) {
	var x int
	typ := FromRuntimeType(refl.TypeOf(x))

	assert.NotNil(t, typ)
	assert.Equal(t, TypeKindBasic, typ.Kind)
	assert.Equal(t, "int", typ.Name)
	// Built-in types should not have elem
	assert.Nil(t, typ.Elem)
}

// TestTypedefOfComplex tests that typedefs of complex types are handled correctly
func TestTypedefOfComplex(t *testing.T) {
	type MyComplex complex128

	typ := FromRuntimeType(refl.TypeOf(MyComplex(0)))

	assert.NotNil(t, typ)
	assert.Equal(t, TypeKindTypedef, typ.Kind)
	// The typedef should preserve its custom name
	assert.Contains(t, typ.Name, "MyComplex")
	// The typedef should have the underlying type reference
	assert.NotNil(t, typ.OriginalType)
	assert.Equal(t, "complex128", typ.OriginalType.Name)
}
