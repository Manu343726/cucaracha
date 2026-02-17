package contract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestToVariableMap_WithMap tests conversion of a map
func TestToVariableMap_WithMap(t *testing.T) {
	input := map[string]interface{}{"x": 10, "y": 20}
	result := ToVariableMap(input)

	assert.Equal(t, input, result)
	assert.Equal(t, 10, result["x"])
	assert.Equal(t, 20, result["y"])
}

// TestToVariableMap_WithNil tests conversion of nil
func TestToVariableMap_WithNil(t *testing.T) {
	result := ToVariableMap(nil)

	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result))
}

// TestToVariableMap_WithAnonymousStruct tests conversion of anonymous struct
func TestToVariableMap_WithAnonymousStruct(t *testing.T) {
	input := struct {
		Age   int
		Name  string
		Score float64
	}{
		Age:   25,
		Name:  "Alice",
		Score: 95.5,
	}

	result := ToVariableMap(input)

	assert.Equal(t, 25, result["Age"])
	assert.Equal(t, "Alice", result["Name"])
	assert.Equal(t, 95.5, result["Score"])
	assert.Equal(t, 3, len(result))
}

// TestToVariableMap_WithStructPointer tests conversion of pointer to struct
func TestToVariableMap_WithStructPointer(t *testing.T) {
	input := &struct {
		Age  int
		Name string
	}{
		Age:  30,
		Name: "Bob",
	}

	result := ToVariableMap(input)

	assert.Equal(t, 30, result["Age"])
	assert.Equal(t, "Bob", result["Name"])
}

// TestToVariableMap_WithMixedFields tests struct with exported and unexported fields
func TestToVariableMap_WithMixedFields(t *testing.T) {
	input := struct {
		Public   int
		private  int // Unexported, should still be included
		Exported string
	}{
		Public:   100,
		private:  999, // Should appear in result
		Exported: "visible",
	}

	result := ToVariableMap(input)

	assert.Equal(t, 100, result["Public"])
	assert.Equal(t, "visible", result["Exported"])
	assert.Equal(t, 999, result["private"], "unexported fields should be included")
	assert.Equal(t, 3, len(result))
}

// TestToVariableMap_WithComplexTypes tests struct with complex field types
func TestToVariableMap_WithComplexTypes(t *testing.T) {
	input := struct {
		Numbers []int
		Mapping map[string]int
		Active  bool
	}{
		Numbers: []int{1, 2, 3},
		Mapping: map[string]int{"a": 1, "b": 2},
		Active:  true,
	}

	result := ToVariableMap(input)

	assert.Equal(t, []int{1, 2, 3}, result["Numbers"])
	assert.Equal(t, map[string]int{"a": 1, "b": 2}, result["Mapping"])
	assert.Equal(t, true, result["Active"])
}

// TestToVariableMap_WithNestedStruct tests struct with nested structs
func TestToVariableMap_WithNestedStruct(t *testing.T) {
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

	result := ToVariableMap(input)

	assert.Equal(t, nested, result["Nested"])
	assert.Equal(t, "test", result["Label"])
}

// TestToVariableMap_WithVariableProvider tests custom VariableProvider implementation
func TestToVariableMap_WithVariableProvider(t *testing.T) {
	provider := &customProvider{
		data: map[string]interface{}{"custom": "value"},
	}

	result := ToVariableMap(provider)

	assert.Equal(t, "value", result["custom"])
}

// customProvider is a test implementation of VariableProvider
type customProvider struct {
	data map[string]interface{}
}

func (cp *customProvider) GetVariables() map[string]interface{} {
	return cp.data
}

// TestToVariableMap_WithPrimitiveType tests non-struct, non-map type
func TestToVariableMap_WithPrimitiveType(t *testing.T) {
	result := ToVariableMap(42)

	assert.NotNil(t, result)
	assert.Equal(t, "42", result["value"])
}

// TestToVariableMap_WithString tests string type
func TestToVariableMap_WithString(t *testing.T) {
	result := ToVariableMap("hello")

	assert.NotNil(t, result)
	assert.Equal(t, "hello", result["value"])
}

// TestToVariableMap_WithEmptyStruct tests empty struct
func TestToVariableMap_WithEmptyStruct(t *testing.T) {
	result := ToVariableMap(struct{}{})

	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result))
}

// TestToVariableMap_RealWorldExample tests a realistic scenario
func TestToVariableMap_RealWorldExample(t *testing.T) {
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

	result := ToVariableMap(user)

	assert.Equal(t, 123, result["ID"])
	assert.Equal(t, "alice@example.com", result["Email"])
	assert.Equal(t, 25, result["Age"])
	assert.Equal(t, true, result["IsActive"])
}

// TestToVariableMap_LowercaseFieldNames tests struct with lowercase (unexported) field names
// This is a key feature: variables in expressions often use lowercase names
func TestToVariableMap_LowercaseFieldNames(t *testing.T) {
	input := struct {
		x   int
		y   int
		sum int
	}{
		x:   10,
		y:   20,
		sum: 30,
	}

	result := ToVariableMap(input)

	// All lowercase field names should be preserved and accessible
	assert.Equal(t, 10, result["x"])
	assert.Equal(t, 20, result["y"])
	assert.Equal(t, 30, result["sum"])
	assert.Equal(t, 3, len(result))
}

// TestToVariableMap_MixedCaseFieldNames tests struct with mixed case field names
func TestToVariableMap_MixedCaseFieldNames(t *testing.T) {
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

	result := ToVariableMap(input)

	assert.Equal(t, 25, result["age"])
	assert.Equal(t, "Alice", result["Name"])
	assert.Equal(t, 1000.0, result["balance"])
	assert.Equal(t, "active", result["Status"])
	assert.Equal(t, 4, len(result))
}
