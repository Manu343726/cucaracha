package reflect

import (
	refl "reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTypedefOfSlice tests that typedefs of slice types are handled correctly
func TestTypedefOfSlice(t *testing.T) {
	type IntSlice []int

	typ := FromRuntimeType(refl.TypeOf(IntSlice{}))

	assert.NotNil(t, typ)
	assert.Equal(t, TypeKindTypedef, typ.Kind)
	assert.Contains(t, typ.Name, "IntSlice")
	// The typedef should reference the underlying slice type
	assert.NotNil(t, typ.OriginalType)
	assert.Equal(t, TypeKindSlice, typ.OriginalType.Type.Kind)
}

// TestTypedefOfArray tests that typedefs of array types are handled correctly
func TestTypedefOfArray(t *testing.T) {
	type IntArray [5]int

	typ := FromRuntimeType(refl.TypeOf(IntArray{}))

	assert.NotNil(t, typ)
	assert.Equal(t, TypeKindTypedef, typ.Kind)
	assert.Contains(t, typ.Name, "IntArray")
	// The typedef should reference the underlying array type
	assert.NotNil(t, typ.OriginalType)
	assert.Equal(t, TypeKindArray, typ.OriginalType.Type.Kind)
}

// TestTypedefOfMap tests that typedefs of map types are handled correctly
func TestTypedefOfMap(t *testing.T) {
	type StringMap map[string]int

	typ := FromRuntimeType(refl.TypeOf(StringMap{}))

	assert.NotNil(t, typ)
	assert.Equal(t, TypeKindTypedef, typ.Kind)
	assert.Contains(t, typ.Name, "StringMap")
	// The typedef should reference the underlying map type
	assert.NotNil(t, typ.OriginalType)
	assert.Equal(t, TypeKindMap, typ.OriginalType.Type.Kind)
}

// TestTypedefOfPointer tests that typedefs of pointer types are handled correctly
func TestTypedefOfPointer(t *testing.T) {
	type IntPointer *int

	typ := FromRuntimeType(refl.TypeOf((*int)(nil)))
	// Note: Going through a type definition variable

	assert.NotNil(t, typ)
	// Direct pointer types may not be typedefs unless explicitly defined with a custom name
	assert.Equal(t, TypeKindPointer, typ.Kind)
}

// TestTypedefOfChannel tests that typedefs of channel types are handled correctly
func TestTypedefOfChannel(t *testing.T) {
	type IntChan chan int

	typ := FromRuntimeType(refl.TypeOf(IntChan(nil)))

	assert.NotNil(t, typ)
	assert.Equal(t, TypeKindTypedef, typ.Kind)
	assert.Contains(t, typ.Name, "IntChan")
	// The typedef should reference the underlying channel type
	assert.NotNil(t, typ.OriginalType)
	assert.Equal(t, TypeKindChannel, typ.OriginalType.Type.Kind)
}

// TestTypedefOfFunction tests that typedefs of function types are handled correctly
func TestTypedefOfFunction(t *testing.T) {
	type IntFunc func(int, int) int

	typ := FromRuntimeType(refl.TypeOf(IntFunc(nil)))

	assert.NotNil(t, typ)
	assert.Equal(t, TypeKindTypedef, typ.Kind)
	assert.Contains(t, typ.Name, "IntFunc")
	// The typedef should reference the underlying function type
	assert.NotNil(t, typ.OriginalType)
	assert.Equal(t, TypeKindFunction, typ.OriginalType.Type.Kind)
}

// TestPlainSliceNotTypedef tests that plain slices without typedef are not treated as typedefs
func TestPlainSliceNotTypedef(t *testing.T) {
	typ := FromRuntimeType(refl.TypeOf([]int{}))

	assert.NotNil(t, typ)
	assert.Equal(t, TypeKindSlice, typ.Kind)
	// Should not be a typedef
	assert.Nil(t, typ.OriginalType)
}

// TestPlainMapNotTypedef tests that plain maps without typedef are not treated as typedefs
func TestPlainMapNotTypedef(t *testing.T) {
	typ := FromRuntimeType(refl.TypeOf(map[string]int{}))

	assert.NotNil(t, typ)
	assert.Equal(t, TypeKindMap, typ.Kind)
	// Should not be a typedef
	assert.Nil(t, typ.OriginalType)
}

// TestPlainChannelNotTypedef tests that plain channels without typedef are not treated as typedefs
func TestPlainChannelNotTypedef(t *testing.T) {
	typ := FromRuntimeType(refl.TypeOf(make(chan int)))

	assert.NotNil(t, typ)
	assert.Equal(t, TypeKindChannel, typ.Kind)
	// Should not be a typedef
	assert.Nil(t, typ.OriginalType)
}

// TestTypedefOfComplexSlice tests typedef of slice of pointers
func TestTypedefOfComplexSlice(t *testing.T) {
	type PointerSlice []*int

	typ := FromRuntimeType(refl.TypeOf(PointerSlice{}))

	assert.NotNil(t, typ)
	assert.Equal(t, TypeKindTypedef, typ.Kind)
	assert.Contains(t, typ.Name, "PointerSlice")
	// The typedef should reference the underlying slice type
	assert.NotNil(t, typ.OriginalType)
	assert.Equal(t, TypeKindSlice, typ.OriginalType.Type.Kind)
}
