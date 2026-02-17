package contract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEnsureThat_WithAnonymousStruct tests EnsureThat with anonymous struct
func TestEnsureThat_WithAnonymousStruct(t *testing.T) {
	logger, _ := createTestLogger()
	condition := true

	// Should not panic with struct variables
	EnsureThat(logger, condition, struct {
		Age  int
		Name string
	}{
		Age:  25,
		Name: "Alice",
	}, "user validation")
}

// TestEnsureThat_WithAnonymousStruct_Fails tests EnsureThat failure with struct
func TestEnsureThat_WithAnonymousStruct_Fails(t *testing.T) {
	logger, _ := createTestLogger()
	condition := false

	assert.Panics(t, func() {
		EnsureThat(logger, condition, struct {
			Age  int
			Name string
		}{
			Age:  16,
			Name: "Bob",
		}, "age must be >= 18")
	}, "EnsureThat should panic when condition is false")
}

// TestRequireThat_WithAnonymousStruct tests RequireThat with anonymous struct
func TestRequireThat_WithAnonymousStruct(t *testing.T) {
	logger, _ := createTestLogger()
	condition := true

	// Should not panic with struct variables
	RequireThat(logger, condition, struct {
		Email string
		Age   int
	}{
		Email: "user@example.com",
		Age:   25,
	}, "")
}

// TestRequireThat_WithAnonymousStruct_Fails tests RequireThat failure with struct
func TestRequireThat_WithAnonymousStruct_Fails(t *testing.T) {
	logger, _ := createTestLogger()
	condition := false

	assert.Panics(t, func() {
		RequireThat(logger, condition, struct {
			Email string
		}{
			Email: "",
		}, "")
	}, "RequireThat should panic when condition is false")
}

// TestInvariantThat_WithAnonymousStruct tests InvariantThat with anonymous struct
func TestInvariantThat_WithAnonymousStruct(t *testing.T) {
	logger, _ := createTestLogger()
	condition := true

	// Should not panic with struct variables
	InvariantThat(logger, condition, struct {
		Balance int
		Status  string
	}{
		Balance: 1000,
		Status:  "active",
	}, "")
}

// TestInvariantThat_WithAnonymousStruct_Fails tests InvariantThat failure with struct
func TestInvariantThat_WithAnonymousStruct_Fails(t *testing.T) {
	logger, _ := createTestLogger()
	condition := false

	assert.Panics(t, func() {
		InvariantThat(logger, condition, struct {
			Balance int
		}{
			Balance: -100,
		}, "")
	}, "InvariantThat should panic when condition is false")
}

// TestAutoCapture_MixedMapAndStruct tests that both map and struct work together
func TestAutoCapture_MixedMapAndStruct(t *testing.T) {
	logger, _ := createTestLogger()

	// Map version
	EnsureThat(logger, true, map[string]interface{}{
		"x": 10,
	}, "")

	// Struct version
	EnsureThat(logger, true, struct {
		X int
	}{
		X: 10,
	}, "")
	// Both should work without panic
}

// TestAutoCapture_VariableProvider_Integration tests VariableProvider integration
func TestAutoCapture_VariableProvider_Integration(t *testing.T) {
	logger, _ := createTestLogger()

	provider := &testVariableProvider{
		data: map[string]interface{}{
			"status": "active",
			"count":  5,
		},
	}

	// Should work with VariableProvider
	EnsureThat(logger, true, provider, "")
}

type testVariableProvider struct {
	data map[string]interface{}
}

func (tvp *testVariableProvider) GetVariables() map[string]interface{} {
	return tvp.data
}

// TestAutoCapture_ComplexStruct tests struct with complex field types
func TestAutoCapture_ComplexStruct(t *testing.T) {
	logger, _ := createTestLogger()

	assert.NotPanics(t, func() {
		EnsureThat(logger, true, struct {
			Tags    []string
			Config  map[string]interface{}
			Enabled bool
		}{
			Tags:    []string{"tag1", "tag2"},
			Config:  map[string]interface{}{"version": "1.0"},
			Enabled: true,
		}, "")
	})
}

// TestAutoCapture_StructPointer tests that struct pointers work
func TestAutoCapture_StructPointer(t *testing.T) {
	logger, _ := createTestLogger()

	type Config struct {
		Host string
		Port int
	}

	cfg := &Config{
		Host: "localhost",
		Port: 8080,
	}

	assert.NotPanics(t, func() {
		EnsureThat(logger, cfg.Port > 0, cfg, "")
	})
}

// TestThatCapture_UnexportedFields tests that unexported fields are included
func TestThatCapture_UnexportedFields(t *testing.T) {
	s := struct {
		Public  int
		private int // Unexported - should still be included
	}{
		Public:  100,
		private: 999,
	}

	// The conversion should include private fields
	result := ToVariableMap(s)
	assert.Equal(t, 100, result["Public"])
	assert.Equal(t, 999, result["private"], "unexported fields should be included")
	assert.Equal(t, 2, len(result))
}

// BenchmarkToVariableMap_Map benchmarks map conversion
func BenchmarkToVariableMap_Map(b *testing.B) {
	input := map[string]interface{}{"x": 10, "y": 20}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ToVariableMap(input)
	}
}

// BenchmarkToVariableMap_Struct benchmarks struct conversion
func BenchmarkToVariableMap_Struct(b *testing.B) {
	input := struct {
		X int
		Y int
	}{X: 10, Y: 20}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ToVariableMap(input)
	}
}

// BenchmarkToVariableMap_StructPointer benchmarks struct pointer conversion
func BenchmarkToVariableMap_StructPointer(b *testing.B) {
	input := &struct {
		X int
		Y int
	}{X: 10, Y: 20}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ToVariableMap(input)
	}
}
