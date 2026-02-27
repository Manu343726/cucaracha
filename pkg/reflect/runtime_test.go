package reflect

import (
	refl "reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestExtractInterfaceMethodsNoParams tests that methods with no parameters are extracted correctly
func TestExtractInterfaceMethodsNoParams(t *testing.T) {
	// Create an interface with a method that has no parameters
	type NoParamInterface interface {
		NoParams()
	}

	rawType := refl.TypeOf((*NoParamInterface)(nil)).Elem()
	methods := extractInterfaceMethods(rawType, make(map[string]*Type))

	assert.Len(t, methods, 1)
	method := methods[0]
	assert.Equal(t, "NoParams", method.Name)
	assert.Len(t, method.Args, 0)
	assert.Len(t, method.Results, 0)
}

// TestExtractInterfaceMethodsNoResults tests that methods with no results are extracted correctly
func TestExtractInterfaceMethodsNoResults(t *testing.T) {
	// Create an interface with a method that has parameters but no results
	type NoResultInterface interface {
		NoResult(x int, y string)
	}

	rawType := refl.TypeOf((*NoResultInterface)(nil)).Elem()
	methods := extractInterfaceMethods(rawType, make(map[string]*Type))

	assert.Len(t, methods, 1)
	method := methods[0]
	assert.Equal(t, "NoResult", method.Name)
	assert.Len(t, method.Args, 2)
	assert.Len(t, method.Results, 0)
}

// TestExtractInterfaceMethodsWithResults tests that methods with both params and results are extracted correctly
func TestExtractInterfaceMethodsWithResults(t *testing.T) {
	// Create an interface with a method that has parameters and results
	type WithResultInterface interface {
		WithResult(x int) (string, error)
	}

	rawType := refl.TypeOf((*WithResultInterface)(nil)).Elem()
	methods := extractInterfaceMethods(rawType, make(map[string]*Type))

	assert.Len(t, methods, 1)
	method := methods[0]
	assert.Equal(t, "WithResult", method.Name)
	assert.Len(t, method.Args, 1)
	assert.Len(t, method.Results, 2)
}

// TestExtractInterfaceMultipleMethods tests interface with multiple methods
func TestExtractInterfaceMultipleMethods(t *testing.T) {
	type ComplexInterface interface {
		NoParams()
		WithParams(int, string)
		WithResults() error
		Full(int) (string, error)
	}

	rawType := refl.TypeOf((*ComplexInterface)(nil)).Elem()
	methods := extractInterfaceMethods(rawType, make(map[string]*Type))

	assert.Len(t, methods, 4)

	// Create a map to check methods by name (they may be in any order)
	methodMap := make(map[string]*Method)
	for _, m := range methods {
		methodMap[m.Name] = m
	}

	testCases := map[string]struct {
		expectedArgs    int
		expectedResults int
	}{
		"NoParams":    {0, 0},
		"WithParams":  {2, 0},
		"WithResults": {0, 1},
		"Full":        {1, 2},
	}

	for name, tc := range testCases {
		method, ok := methodMap[name]
		assert.True(t, ok, "expected method '%s' not found", name)
		if ok {
			assert.Len(t, method.Args, tc.expectedArgs, "method '%s' args count mismatch", name)
			assert.Len(t, method.Results, tc.expectedResults, "method '%s' results count mismatch", name)
		}
	}
}

// TestExtractFunctionParamsWithCount tests the extractFunctionParams function with explicit count
func TestExtractFunctionParamsWithCount(t *testing.T) {
	cache := make(map[string]*Type)

	// Test function type with 0 params and 0 results
	funcNoParamsNoResults := refl.TypeOf(func() {})
	params := extractFunctionParams(funcNoParamsNoResults.In, funcNoParamsNoResults.NumIn(), cache)
	results := extractFunctionParams(funcNoParamsNoResults.Out, funcNoParamsNoResults.NumOut(), cache)

	assert.Len(t, params, 0)
	assert.Len(t, results, 0)

	// Test function type with params but no results
	funcWithParamsNoResults := refl.TypeOf(func(int, string) {})
	params = extractFunctionParams(funcWithParamsNoResults.In, funcWithParamsNoResults.NumIn(), cache)
	results = extractFunctionParams(funcWithParamsNoResults.Out, funcWithParamsNoResults.NumOut(), cache)

	assert.Len(t, params, 2)
	assert.Len(t, results, 0)

	// Test function type with no params but with results
	funcNoParamsWithResults := refl.TypeOf(func() (string, error) { return "", nil })
	params = extractFunctionParams(funcNoParamsWithResults.In, funcNoParamsWithResults.NumIn(), cache)
	results = extractFunctionParams(funcNoParamsWithResults.Out, funcNoParamsWithResults.NumOut(), cache)

	assert.Len(t, params, 0)
	assert.Len(t, results, 2)
}

// TestFromRuntimeTypeFunction tests that function types are extracted correctly
func TestFromRuntimeTypeFunction(t *testing.T) {
	testCases := []struct {
		name            string
		fn              interface{}
		expectedArgs    int
		expectedResults int
	}{
		{"no params no results", func() {}, 0, 0},
		{"one param no result", func(int) {}, 1, 0},
		{"two params no result", func(int, string) {}, 2, 0},
		{"no param one result", func() error { return nil }, 0, 1},
		{"no param two results", func() (string, error) { return "", nil }, 0, 2},
		{"params and results", func(int, string) (bool, error) { return false, nil }, 2, 2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			funcType := refl.TypeOf(tc.fn)
			extracted := FromRuntimeType(funcType)

			assert.Equal(t, TypeKindFunction, extracted.Kind)
			assert.Len(t, extracted.Args, tc.expectedArgs)
			assert.Len(t, extracted.Results, tc.expectedResults)
		})
	}
}

// BenchmarkExtractInterfaceMethods benchmarks the interface method extraction
func BenchmarkExtractInterfaceMethods(b *testing.B) {
	type BenchInterface interface {
		Method1(int, string) (bool, error)
		Method2() string
		Method3(float64)
		Method4() (int, int, int)
	}

	rawType := refl.TypeOf((*BenchInterface)(nil)).Elem()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractInterfaceMethods(rawType, make(map[string]*Type))
	}
}

// TestEdgeCaseEmptyInterface tests an empty interface by checking we can handle interface{} types
func TestEdgeCaseEmptyInterface(t *testing.T) {
	// Test with a concrete interface type, not an empty interface variable
	var i interface{}
	rt := refl.TypeOf(i)
	// For interface{}, TypeOf returns nil, and that's expected behavior
	if rt != nil {
		// If we do have a type, verify we handle it gracefully
		extracted := FromRuntimeType(rt)
		// Should not panic
		assert.NotNil(t, extracted)
	}
}

// TestEdgeCaseNilType tests that nil types are handled gracefully
func TestEdgeCaseNilType(t *testing.T) {
	extracted := FromRuntimeType(nil)
	assert.Nil(t, extracted)
}

// TestComplexInterfaceSignature tests a realistic complex interface
func TestComplexInterfaceSignature(t *testing.T) {
	type Reader interface {
		Read(p []byte) (n int, err error)
	}

	rawType := refl.TypeOf((*Reader)(nil)).Elem()
	methods := extractInterfaceMethods(rawType, make(map[string]*Type))

	assert.Len(t, methods, 1)
	method := methods[0]
	assert.Equal(t, "Read", method.Name)
	assert.Len(t, method.Args, 1)
	assert.Len(t, method.Results, 2)
	assert.NotEmpty(t, method.Signature, "expected non-empty signature string")
}

// TestExtractFunctionParamsNilFunc tests that nil function accessors are handled
func TestExtractFunctionParamsNilFunc(t *testing.T) {
	cache := make(map[string]*Type)
	params := extractFunctionParams(nil, 0, cache)

	assert.Len(t, params, 0)
}

// TestTypeReferencesToParameters tests conversion of TypeReferences to Parameters
func TestTypeReferencesToParameters(t *testing.T) {
	refs := []*TypeReference{
		{Name: "int", Type: &Type{Name: "int", Kind: TypeKindBasic}},
		{Name: "string", Type: &Type{Name: "string", Kind: TypeKindBasic}},
	}

	params := typeReferencesToParameters(refs)

	assert.Len(t, params, 2)
	for i, param := range params {
		assert.Equal(t, refs[i], param.Type)
	}
}

// TestEmptyTypeReferencesToParameters tests conversion with empty slice
func TestEmptyTypeReferencesToParameters(t *testing.T) {
	refs := []*TypeReference{}
	params := typeReferencesToParameters(refs)

	assert.Len(t, params, 0)
}

// TestInterfaceWithManyMethods tests an interface with many methods to ensure all are extracted
func TestInterfaceWithManyMethods(t *testing.T) {
	type ManyMethodsInterface interface {
		Method1()
		Method2(int) error
		Method3(string, bool) (int, error)
		Method4() (string, string, string)
		Method5(int, int, int, int, int) (bool, bool, bool, bool, bool)
	}

	rawType := refl.TypeOf((*ManyMethodsInterface)(nil)).Elem()
	methods := extractInterfaceMethods(rawType, make(map[string]*Type))

	assert.Len(t, methods, 5)

	// Create a map to check methods by name
	methodMap := make(map[string]*Method)
	for _, m := range methods {
		methodMap[m.Name] = m
	}

	expectedCounts := map[string]struct {
		args    int
		results int
	}{
		"Method1": {0, 0},
		"Method2": {1, 1},
		"Method3": {2, 2},
		"Method4": {0, 3},
		"Method5": {5, 5},
	}

	for name, expected := range expectedCounts {
		method, ok := methodMap[name]
		assert.True(t, ok, "expected method '%s' not found", name)
		if ok {
			assert.Len(t, method.Args, expected.args, "method '%s' args count mismatch", name)
			assert.Len(t, method.Results, expected.results, "method '%s' results count mismatch", name)
		}
	}
}
