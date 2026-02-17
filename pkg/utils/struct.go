package utils

import (
	"fmt"
	"reflect"
	"unsafe"
)

// StructToMap converts various input types to a map[string]interface{}.
// It supports:
// 1. map[string]interface{} - returned as-is
// 2. Struct (by pointer or value) - reflects on fields to extract names and values
// 3. nil - returns empty map
//
// For structs, ALL fields (both exported and unexported) are included.
// Field names from the struct become the keys in the returned map.
// Nested structs are kept as-is, not flattened.
//
// This function uses reflection and unsafe operations to access unexported fields.
// It's particularly useful for contracts, testing, and debugging where you need
// to capture all struct fields as a map regardless of their export status.
func StructToMap(v interface{}) map[string]interface{} {
	if v == nil {
		return make(map[string]interface{})
	}

	// If it's already a map, return it
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}

	// Handle struct types using reflection
	result := make(map[string]interface{})
	val := reflect.ValueOf(v)

	// Dereference pointer if needed
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Only process struct types
	if val.Kind() != reflect.Struct {
		// For non-struct, non-map types, try to convert to string representation
		return map[string]interface{}{"value": fmt.Sprintf("%v", v)}
	}

	// For accessing unexported fields, we need an addressable value
	// If the value is not addressable, convert it through pointer
	if !val.CanAddr() && val.Kind() == reflect.Struct {
		// Create a pointer to the struct value
		ptr := reflect.New(val.Type())
		ptr.Elem().Set(val)
		val = ptr.Elem()
	}

	// Extract all fields from the struct (both exported and unexported)
	structType := val.Type()
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldVal := val.Field(i)

		// For unexported fields, we need to use unsafe to get the value
		// since reflect.Value.Interface() panics on unexported fields
		var fieldValue interface{}
		if field.IsExported() {
			fieldValue = fieldVal.Interface()
		} else {
			// Use unsafe to read unexported field value
			// This is safe because we're just reading the value, not modifying it
			if fieldVal.CanAddr() {
				fieldValue = reflect.NewAt(fieldVal.Type(), unsafe.Pointer(fieldVal.UnsafeAddr())).Elem().Interface()
			} else {
				// Fallback: convert to string representation
				fieldValue = fmt.Sprintf("%v", fieldVal)
			}
		}
		result[field.Name] = fieldValue
	}

	return result
}
