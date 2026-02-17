package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestStructToMap_WithMap tests conversion of a map
func TestStructToMap_WithMap(t *testing.T) {
	input := map[string]interface{}{"x": 10, "y": 20}
	result := StructToMap(input)

	assert.Equal(t, input, result)
	assert.Equal(t, 10, result["x"])
	assert.Equal(t, 20, result["y"])
}

// TestStructToMap_WithNil tests conversion of nil
func TestStructToMap_WithNil(t *testing.T) {
	result := StructToMap(nil)

	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result))
}

// TestStructToMap_WithAnonymousStruct tests conversion of anonymous struct
func TestStructToMap_WithAnonymousStruct(t *testing.T) {
	input := struct {
		Age   int
		Name  string
		Score float64
	}{
		Age:   25,
		Name:  "Alice",
		Score: 95.5,
	}

	result := StructToMap(input)

	assert.Equal(t, 25, result["Age"])
	assert.Equal(t, "Alice", result["Name"])
	assert.Equal(t, 95.5, result["Score"])
	assert.Equal(t, 3, len(result))
}

// TestStructToMap_WithStructPointer tests conversion of pointer to struct
func TestStructToMap_WithStructPointer(t *testing.T) {
	input := &struct {
		Age  int
		Name string
	}{
		Age:  30,
		Name: "Bob",
	}

	result := StructToMap(input)

	assert.Equal(t, 30, result["Age"])
	assert.Equal(t, "Bob", result["Name"])
}

// TestStructToMap_WithMixedFields tests struct with exported and unexported fields
func TestStructToMap_WithMixedFields(t *testing.T) {
	input := struct {
		Public   int
		private  int // Unexported, should still be included
		Exported string
	}{
		Public:   100,
		private:  999, // Should appear in result
		Exported: "visible",
	}

	result := StructToMap(input)

	assert.Equal(t, 100, result["Public"])
	assert.Equal(t, "visible", result["Exported"])
	assert.Equal(t, 999, result["private"], "unexported fields should be included")
	assert.Equal(t, 3, len(result))
}

// TestStructToMap_WithComplexTypes tests struct with complex field types
func TestStructToMap_WithComplexTypes(t *testing.T) {
	input := struct {
		Numbers []int
		Mapping map[string]int
		Active  bool
	}{
		Numbers: []int{1, 2, 3},
		Mapping: map[string]int{"a": 1, "b": 2},
		Active:  true,
	}

	result := StructToMap(input)

	assert.Equal(t, []int{1, 2, 3}, result["Numbers"])
	assert.Equal(t, map[string]int{"a": 1, "b": 2}, result["Mapping"])
	assert.Equal(t, true, result["Active"])
}

// TestStructToMap_WithNestedStruct tests struct with nested structs
func TestStructToMap_WithNestedStruct(t *testing.T) {
	nested := struct {
		X int
		Y int
	}{X: 5, Y: 10}

	input := struct {
		Nested struct {
			X int
			Y int
		}
		Label string
	}{
		Nested: nested,
		Label:  "test",
	}

	result := StructToMap(input)

	assert.Equal(t, nested, result["Nested"])
	assert.Equal(t, "test", result["Label"])
}

// TestStructToMap_WithPrimitiveType tests non-struct, non-map type
func TestStructToMap_WithPrimitiveType(t *testing.T) {
	result := StructToMap(42)

	assert.NotNil(t, result)
	assert.Equal(t, "42", result["value"])
}

// TestStructToMap_WithString tests string type
func TestStructToMap_WithString(t *testing.T) {
	result := StructToMap("hello")

	assert.NotNil(t, result)
	assert.Equal(t, "hello", result["value"])
}

// TestStructToMap_WithEmptyStruct tests empty struct
func TestStructToMap_WithEmptyStruct(t *testing.T) {
	result := StructToMap(struct{}{})

	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result))
}

// TestStructToMap_RealWorldExample tests a realistic scenario
func TestStructToMap_RealWorldExample(t *testing.T) {
	// User validation scenario
	user := struct {
		ID       int
		Email    string
		Age      int
		IsActive bool
	}{
		ID:       123,
		Email:    "alice@example.com",
		Age:      25,
		IsActive: true,
	}

	result := StructToMap(user)

	assert.Equal(t, 123, result["ID"])
	assert.Equal(t, "alice@example.com", result["Email"])
	assert.Equal(t, 25, result["Age"])
	assert.Equal(t, true, result["IsActive"])
}

// TestStructToMap_LowercaseFieldNames tests struct with lowercase (unexported) field names
func TestStructToMap_LowercaseFieldNames(t *testing.T) {
	input := struct {
		x   int
		y   int
		sum int
	}{
		x:   10,
		y:   20,
		sum: 30,
	}

	result := StructToMap(input)

	// All lowercase field names should be preserved and accessible
	assert.Equal(t, 10, result["x"])
	assert.Equal(t, 20, result["y"])
	assert.Equal(t, 30, result["sum"])
	assert.Equal(t, 3, len(result))
}

// TestStructToMap_MixedCaseFieldNames tests struct with mixed case field names
func TestStructToMap_MixedCaseFieldNames(t *testing.T) {
	input := struct {
		age     int
		Name    string
		balance float64
		Status  string
	}{
		age:     25,
		Name:    "Alice",
		balance: 1000.0,
		Status:  "active",
	}

	result := StructToMap(input)

	assert.Equal(t, 25, result["age"])
	assert.Equal(t, "Alice", result["Name"])
	assert.Equal(t, 1000.0, result["balance"])
	assert.Equal(t, "active", result["Status"])
	assert.Equal(t, 4, len(result))
}

// BenchmarkStructToMap_Map benchmarks map conversion
func BenchmarkStructToMap_Map(b *testing.B) {
	input := map[string]interface{}{"x": 10, "y": 20}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		StructToMap(input)
	}
}

// BenchmarkStructToMap_Struct benchmarks struct conversion
func BenchmarkStructToMap_Struct(b *testing.B) {
	input := struct {
		X int
		Y int
	}{X: 10, Y: 20}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		StructToMap(input)
	}
}

// BenchmarkStructToMap_StructPointer benchmarks struct pointer conversion
func BenchmarkStructToMap_StructPointer(b *testing.B) {
	input := &struct {
		X int
		Y int
	}{X: 10, Y: 20}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		StructToMap(input)
	}
}
