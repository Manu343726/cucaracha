package reflect

import (
	"fmt"
	refl "reflect"
	"strings"
)

// FromRuntimeType converts a golang reflect.Type to a cucaracha reflect.Type.
// It attempts to fill as much information as possible from the runtime reflect package,
// with graceful degradation when type information is not available.
func FromRuntimeType(t refl.Type) *Type {
	if t == nil {
		return nil
	}

	return fromRuntimeTypeInternal(t, make(map[string]*Type))
}

// fromRuntimeTypeInternal is the internal implementation with a cache to avoid infinite recursion
func fromRuntimeTypeInternal(t refl.Type, cache map[string]*Type) *Type {
	if t == nil {
		return nil
	}

	// Check cache first
	cacheKey := t.String()
	if cached, exists := cache[cacheKey]; exists {
		return cached
	}

	// Handle basic types first
	if basicType := GetBasicType(t.Name()); basicType != nil {
		cache[cacheKey] = basicType
		return basicType
	}

	// Create the result type early and add to cache to break cycles
	result := &Type{
		Name: t.String(),
	}
	cache[cacheKey] = result

	// Process based on kind
	kind := t.Kind()

	switch kind {
	case refl.Ptr:
		elem := fromRuntimeTypeInternal(t.Elem(), cache)
		return Pointer(elem)

	case refl.Slice:
		elem := fromRuntimeTypeInternal(t.Elem(), cache)
		return Slice(elem)

	case refl.Array:
		elem := fromRuntimeTypeInternal(t.Elem(), cache)
		return Array(elem, t.Len())

	case refl.Map:
		keyType := fromRuntimeTypeInternal(t.Key(), cache)
		valType := fromRuntimeTypeInternal(t.Elem(), cache)
		return Map(keyType, valType)

	case refl.Chan:
		elem := fromRuntimeTypeInternal(t.Elem(), cache)
		// Determine channel direction
		var chanDir ChanDirection
		if t.ChanDir() == refl.SendDir {
			chanDir = ChanSend
		} else if t.ChanDir() == refl.RecvDir {
			chanDir = ChanRecv
		} else {
			chanDir = ChanBidirectional
		}
		return Chan(elem, chanDir)

	case refl.Struct:
		result.Kind = TypeKindStruct
		result.Name = t.Name()
		result.Fields = extractStructFields(t, cache)

	case refl.Interface:
		result.Kind = TypeKindInterface
		result.Name = t.Name()
		result.Methods = extractInterfaceMethods(t, cache)

	case refl.Func:
		result.Kind = TypeKindFunction
		result.Args = extractFunctionParams(t.In, cache)
		result.Results = extractFunctionParams(t.Out, cache)
		result.Name = buildFunctionSignature(t, cache)

	default:
		// For unhandled kinds, assume it's a basic or aliased type
		result.Kind = TypeKindBasic
	}

	return result
}

// extractStructFields extracts fields from a struct type
func extractStructFields(t refl.Type, cache map[string]*Type) []*Field {
	var fields []*Field

	numFields := t.NumField()
	for i := 0; i < numFields; i++ {
		sf := t.Field(i)

		fieldType := fromRuntimeTypeInternal(sf.Type, cache)
		field := &Field{
			Name:       sf.Name,
			Type:       &TypeReference{Name: fieldType.Name, Type: fieldType},
			Tag:        string(sf.Tag),
			IsEmbedded: sf.Anonymous,
		}
		fields = append(fields, field)
	}

	return fields
}

// extractInterfaceMethods extracts methods from an interface type
func extractInterfaceMethods(t refl.Type, cache map[string]*Type) []*Method {
	var methods []*Method

	numMethods := t.NumMethod()
	for i := 0; i < numMethods; i++ {
		m := t.Method(i)

		method := &Method{
			Name:      m.Name,
			Signature: buildMethodSignature(m.Type, cache),
			Args:      typeReferencesToParameters(extractFunctionParams(m.Type.In, cache)),
			Results:   typeReferencesToParameters(extractFunctionParams(m.Type.Out, cache)),
		}
		methods = append(methods, method)
	}

	return methods
}

// extractFunctionParams extracts parameters from a variadic function input/output function
func extractFunctionParams(typeFunc func(int) refl.Type, cache map[string]*Type) []*TypeReference {
	var params []*TypeReference

	if typeFunc == nil {
		return params
	}

	// Count the number of parameters
	count := 0
	for {
		t := typeFunc(count)
		if t == nil {
			break
		}
		count++
	}

	for i := 0; i < count; i++ {
		paramType := fromRuntimeTypeInternal(typeFunc(i), cache)
		params = append(params, &TypeReference{
			Name: paramType.Name,
			Type: paramType,
		})
	}

	return params
}

// typeReferencesToParameters converts TypeReference slice to Parameter slice
func typeReferencesToParameters(refs []*TypeReference) []*Parameter {
	var params []*Parameter
	for _, ref := range refs {
		params = append(params, &Parameter{
			Type: ref,
		})
	}
	return params
}

// buildFunctionSignature builds a function signature string from a function type
func buildFunctionSignature(t refl.Type, cache map[string]*Type) string {
	var params []string
	for i := 0; i < t.NumIn(); i++ {
		inType := fromRuntimeTypeInternal(t.In(i), cache)
		params = append(params, inType.Name)
	}

	var results []string
	for i := 0; i < t.NumOut(); i++ {
		outType := fromRuntimeTypeInternal(t.Out(i), cache)
		results = append(results, outType.Name)
	}

	paramStr := "(" + strings.Join(params, ", ") + ")"
	resultStr := "(" + strings.Join(results, ", ") + ")"
	if len(results) == 1 && !strings.Contains(results[0], ",") {
		resultStr = results[0]
	}

	return fmt.Sprintf("func%s %s", paramStr, resultStr)
}

// buildMethodSignature builds a method signature string from a function type
func buildMethodSignature(t refl.Type, cache map[string]*Type) string {
	var params []string
	for i := 0; i < t.NumIn(); i++ {
		inType := fromRuntimeTypeInternal(t.In(i), cache)
		params = append(params, inType.Name)
	}

	var results []string
	for i := 0; i < t.NumOut(); i++ {
		outType := fromRuntimeTypeInternal(t.Out(i), cache)
		results = append(results, outType.Name)
	}

	paramStr := "(" + strings.Join(params, ", ") + ")"
	resultStr := "(" + strings.Join(results, ", ") + ")"
	if len(results) == 1 && !strings.Contains(results[0], ",") {
		resultStr = results[0]
	}

	return fmt.Sprintf("%s %s", paramStr, resultStr)
}

// FromRuntimeValue extracts the type information from a runtime value's type.
// This is a convenience function that calls FromRuntimeType on the value's reflect.Type.
func FromRuntimeValue(v interface{}) *Type {
	return FromRuntimeType(refl.TypeOf(v))
}

// mergeParameterNames fills in parameter names from parsed parameters into runtime parameters.
// Runtime reflection doesn't provide parameter names, only types, so this function copies
// the names from the parsed type information when available.
func mergeParameterNames(runtimeParams, parsedParams []*Parameter) {
	// Match parameters by position and copy names from parsed params to runtime params
	for i := 0; i < len(runtimeParams) && i < len(parsedParams); i++ {
		if runtimeParams[i].Name == "" && parsedParams[i].Name != "" {
			runtimeParams[i].Name = parsedParams[i].Name
		}
	}
}

// MergeRuntimeAndParsedType combines information from a runtime reflect.Type with
// information parsed from source code. The runtime type provides complete structural information,
// while the parsed type may provide documentation and source location information.
// The runtime type takes precedence for structural data, while parsed documentation is preserved.
func MergeRuntimeAndParsedType(runtimeType *Type, parsedType *Type) *Type {
	if runtimeType == nil {
		return parsedType
	}
	if parsedType == nil {
		return runtimeType
	}

	// Use runtime type as the base
	merged := *runtimeType

	// Preserve documentation from parsed type if runtime type lacks it
	if merged.Doc == "" && parsedType.Doc != "" {
		merged.Doc = parsedType.Doc
	}

	// Merge method documentation for interfaces
	if merged.Kind == TypeKindInterface && parsedType.Kind == TypeKindInterface {
		// Create a map of parsed methods for easy lookup
		parsedMethodMap := make(map[string]*Method)
		for _, m := range parsedType.Methods {
			parsedMethodMap[m.Name] = m
		}

		// Enhance runtime methods with parsed documentation and parameter names
		for _, runtimeMethod := range merged.Methods {
			if parsedMethod, exists := parsedMethodMap[runtimeMethod.Name]; exists {
				if runtimeMethod.Doc == "" {
					runtimeMethod.Doc = parsedMethod.Doc
				}
				// Merge parameter names from parsed method
				mergeParameterNames(runtimeMethod.Args, parsedMethod.Args)
				mergeParameterNames(runtimeMethod.Results, parsedMethod.Results)
			}
		}
	}

	// Merge field documentation for structs
	if merged.Kind == TypeKindStruct && parsedType.Kind == TypeKindStruct {
		// Create a map of parsed fields for easy lookup
		parsedFieldMap := make(map[string]*Field)
		for _, f := range parsedType.Fields {
			parsedFieldMap[f.Name] = f
		}

		// Enhance runtime fields with parsed documentation
		for _, runtimeField := range merged.Fields {
			if parsedField, exists := parsedFieldMap[runtimeField.Name]; exists {
				if runtimeField.Doc == "" {
					runtimeField.Doc = parsedField.Doc
				}
			}
		}
	}

	return &merged
}
