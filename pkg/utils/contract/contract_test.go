package contract

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Manu343726/cucaracha/pkg/utils/logging"
)

// Helper to create a test logger
func createTestLogger() (*logging.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	sink := logging.NewFileSink("test", buf, slog.LevelDebug)
	registry := logging.NewRegistry()
	regLogger := logging.NewRegisteredLogger("test", sink)
	registry.RegisterLogger(regLogger)
	return registry.Get("test"), buf
}

// TestRequire_Success tests that Require doesn't panic on true condition
func TestRequire_Success(t *testing.T) {
	logger, _ := createTestLogger()
	// Should not panic
	Require(logger, true, "should not panic")
}

// TestRequire_Failure tests that Require panics on false condition
func TestRequire_Failure(t *testing.T) {
	logger, _ := createTestLogger()
	assert.Panics(t, func() {
		Require(logger, false, "test precondition failed")
	}, "expected panic on false condition")
}

// TestRequireValue_WithValidator_Success tests RequireValue with validator that passes
func TestRequireValue_WithValidator_Success(t *testing.T) {
	logger, _ := createTestLogger()
	// Should not panic
	RequireValue(logger, 10, func(v int) bool { return v > 0 }, "value must be positive")
}

// TestRequireValue_WithValidator_Failure tests RequireValue with validator that fails
func TestRequireValue_WithValidator_Failure(t *testing.T) {
	logger, _ := createTestLogger()
	assert.Panics(t, func() {
		RequireValue(logger, -5, func(v int) bool { return v > 0 }, "value must be positive")
	}, "expected panic on validation failure")
}

// TestRequireEach_AllPass tests RequireEach when all elements pass
func TestRequireEach_AllPass(t *testing.T) {
	logger, _ := createTestLogger()
	values := []int{1, 2, 3, 4, 5}
	// Should not panic
	RequireEach(logger, values, func(v int) bool { return v > 0 }, "all values must be positive")
}

// TestRequireEach_OneFails tests RequireEach panics on first failure
func TestRequireEach_OneFails(t *testing.T) {
	logger, _ := createTestLogger()
	values := []int{1, 2, -3, 4, 5}
	assert.Panics(t, func() {
		RequireEach(logger, values, func(v int) bool { return v > 0 }, "all values must be positive")
	}, "expected panic when element fails validation")
}

// TestEnsure_Success tests that Ensure doesn't panic on true condition
func TestEnsure_Success(t *testing.T) {
	logger, _ := createTestLogger()
	// Should not panic
	Ensure(logger, true, "postcondition should hold")
}

// TestEnsure_Failure tests that Ensure panics on false condition
func TestEnsure_Failure(t *testing.T) {
	logger, _ := createTestLogger()
	assert.Panics(t, func() {
		Ensure(logger, false, "postcondition violated")
	}, "expected panic on failed postcondition")
}

// TestEnsureValue_WithValidator tests EnsureValue with validator
func TestEnsureValue_WithValidator(t *testing.T) {
	logger, _ := createTestLogger()
	result := 42
	// Should not panic
	EnsureValue(logger, result, func(v int) bool { return v > 0 }, "result must be positive")
}

// TestEnsureValue_WithValidator_Failure tests EnsureValue panics on failed validation
func TestEnsureValue_WithValidator_Failure(t *testing.T) {
	logger, _ := createTestLogger()
	result := -1
	assert.Panics(t, func() {
		EnsureValue(logger, result, func(v int) bool { return v > 0 }, "result must be positive")
	}, "expected panic on postcondition failure")
}

// TestInvariant_Success tests that Invariant doesn't panic on true condition
func TestInvariant_Success(t *testing.T) {
	logger, _ := createTestLogger()
	balance := 100
	// Should not panic
	Invariant(logger, balance >= 0, "balance must be non-negative")
}

// TestInvariant_Failure tests that Invariant panics on false condition
func TestInvariant_Failure(t *testing.T) {
	logger, _ := createTestLogger()
	balance := -50
	assert.Panics(t, func() {
		Invariant(logger, balance >= 0, "balance must be non-negative")
	}, "expected panic on invariant violation")
}

// TestInvariantEach_AllPass tests InvariantEach when all elements pass
func TestInvariantEach_AllPass(t *testing.T) {
	logger, _ := createTestLogger()
	items := []int{1, 2, 3, 4, 5}
	// Should not panic
	InvariantEach(logger, items, func(v int) bool { return v > 0 }, "all items must be positive")
}

// TestInvariantEach_OneFails tests InvariantEach panics on first failure
func TestInvariantEach_OneFails(t *testing.T) {
	logger, _ := createTestLogger()
	items := []int{1, 2, 0, 4, 5}
	assert.Panics(t, func() {
		InvariantEach(logger, items, func(v int) bool { return v > 0 }, "all items must be positive")
	}, "expected panic on invariant violation")
}

// TestRangeInt_InBounds tests RangeInt with value in bounds
func TestRangeInt_InBounds(t *testing.T) {
	logger, _ := createTestLogger()
	// Should not panic
	RangeInt(logger, 50, 0, 100, "value must be 0-100")
}

// TestRangeInt_BelowMin tests RangeInt with value below minimum
func TestRangeInt_BelowMin(t *testing.T) {
	logger, _ := createTestLogger()
	assert.Panics(t, func() {
		RangeInt(logger, -1, 0, 100, "value must be 0-100")
	}, "expected panic when value below minimum")
}

// TestRangeInt_AboveMax tests RangeInt with value above maximum
func TestRangeInt_AboveMax(t *testing.T) {
	logger, _ := createTestLogger()
	assert.Panics(t, func() {
		RangeInt(logger, 101, 0, 100, "value must be 0-100")
	}, "expected panic when value above maximum")
}

// TestNotNil_Valid tests NotNil with non-nil pointer
func TestNotNil_Valid(t *testing.T) {
	logger, _ := createTestLogger()
	value := 42
	// Should not panic
	NotNil(logger, &value, "value must not be nil")
}

// TestNotNil_Nil tests NotNil with nil pointer
func TestNotNil_Nil(t *testing.T) {
	logger, _ := createTestLogger()
	var value *int = nil
	assert.Panics(t, func() {
		NotNil(logger, value, "value must not be nil")
	}, "expected panic when pointer is nil")
}

// TestNotEmpty_Valid tests NotEmpty with non-empty string
func TestNotEmpty_Valid(t *testing.T) {
	logger, _ := createTestLogger()
	// Should not panic
	NotEmpty(logger, "hello", "string must not be empty")
}

// TestNotEmpty_Empty tests NotEmpty with empty string
func TestNotEmpty_Empty(t *testing.T) {
	logger, _ := createTestLogger()
	assert.Panics(t, func() {
		NotEmpty(logger, "", "string must not be empty")
	}, "expected panic when string is empty")
}

// TestNotEmptySlice_Valid tests NotEmptySlice with non-empty slice
func TestNotEmptySlice_Valid(t *testing.T) {
	logger, _ := createTestLogger()
	values := []int{1, 2, 3}
	// Should not panic
	NotEmptySlice(logger, values, "slice must not be empty")
}

// TestNotEmptySlice_Empty tests NotEmptySlice with empty slice
func TestNotEmptySlice_Empty(t *testing.T) {
	logger, _ := createTestLogger()
	var values []int
	assert.Panics(t, func() {
		NotEmptySlice(logger, values, "slice must not be empty")
	}, "expected panic when slice is empty")
}

// TestRequireValue_WithMultipleTypes tests generic RequireValue works with different types
func TestRequireValue_WithMultipleTypes(t *testing.T) {
	logger, _ := createTestLogger()

	// Test with string
	RequireValue(logger, "hello", func(s string) bool {
		return len(s) > 0
	}, "string must not be empty")

	// Test with struct
	type Point struct {
		X, Y int
	}
	RequireValue(logger, Point{1, 2}, func(p Point) bool {
		return p.X != p.Y
	}, "coordinates must be different")
}

// TestContractWithAttributes tests contract checks include additional attributes
func TestContractWithAttributes(t *testing.T) {
	logger, buf := createTestLogger()
	assert.Panics(t, func() {
		Require(logger, false, "test failed", slog.String("code", "ERR_001"))
	}, "expected panic with attributes")
	// Check that attributes were logged
	output := buf.String()
	assert.NotEmpty(t, output, "expected log output with attributes")
}
