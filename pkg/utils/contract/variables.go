package contract

import (
	"github.com/Manu343726/cucaracha/pkg/utils"
)

// VariableProvider is an interface that allows flexible passing of variables.
// It can be implemented by custom types that want to provide variables in a special way.
type VariableProvider interface {
	// GetVariables returns the variables as a map[string]interface{}
	GetVariables() map[string]interface{}
}

// ToVariableMap converts various input types to a map[string]interface{}.
// It supports:
// 1. map[string]interface{} - returned as-is
// 2. VariableProvider interface - calls GetVariables()
// 3. Struct (by pointer or value) - reflects on fields to extract names and values
// 4. nil - returns empty map
//
// For structs, ALL fields (both exported and unexported) are included.
// Field names from the struct become the keys in the returned map.
// Nested structs are kept as-is, not flattened.
//
// This is a wrapper around utils.StructToMap that adds VariableProvider support.
func ToVariableMap(v interface{}) map[string]interface{} {
	// If it implements VariableProvider interface, use that
	if vp, ok := v.(VariableProvider); ok {
		return vp.GetVariables()
	}

	// Otherwise, delegate to utils.StructToMap
	return utils.StructToMap(v)
}
