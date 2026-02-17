package contract

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createTestBase(t *testing.T) Base {
	t.Helper()
	logger, _ := createTestLogger()
	return NewBase(logger)
}

func TestBase_Log(t *testing.T) {
	base := createTestBase(t)

	result := base.Log()
	assert.Equal(t, base.logger, result, "log() returned wrong logger")
}

func TestBase_Require_Success(t *testing.T) {
	base := createTestBase(t)

	// Should not panic
	base.Require(true, "condition is true")
}

func TestBase_Require_SuccessWithFormatted(t *testing.T) {
	base := createTestBase(t)

	x := 10
	// Should not panic
	base.Require(x > 0, "x=%d must be positive", x)
}

func TestBase_Require_Failure(t *testing.T) {
	base := createTestBase(t)

	assert.Panics(t, func() {
		base.Require(false, "condition is false")
	}, "Require should have panicked")
}

func TestBase_Require_FailureWithFormatted(t *testing.T) {
	base := createTestBase(t)

	x := -5
	assert.Panics(t, func() {
		base.Require(x > 0, "x=%d but expected x > 0", x)
	}, "Require should have panicked")
}

func TestBase_RequireDefaultMessage(t *testing.T) {
	base := createTestBase(t)

	// Should use default message
	base.Require(true)
}

func TestBase_Ensure_Success(t *testing.T) {
	base := createTestBase(t)

	base.Ensure(true, "postcondition met")
}

func TestBase_Ensure_SuccessWithFormatted(t *testing.T) {
	base := createTestBase(t)

	result := 42
	base.Ensure(result > 0, "result=%d must be positive", result)
}

func TestBase_Ensure_Failure(t *testing.T) {
	base := createTestBase(t)

	assert.Panics(t, func() {
		base.Ensure(false, "postcondition failed")
	}, "Ensure should have panicked")
}

func TestBase_Ensure_FailureWithFormatted(t *testing.T) {
	base := createTestBase(t)

	result := -1
	assert.Panics(t, func() {
		base.Ensure(result > 0, "result=%d but expected positive", result)
	}, "Ensure should have panicked")
}

func TestBase_Invariant_Success(t *testing.T) {
	base := createTestBase(t)

	base.Invariant(true, "invariant holds")
}

func TestBase_Invariant_SuccessWithFormatted(t *testing.T) {
	base := createTestBase(t)

	state := "active"
	base.Invariant(state != "", "state=%s must not be empty", state)
}

func TestBase_Invariant_Failure(t *testing.T) {
	base := createTestBase(t)

	assert.Panics(t, func() {
		base.Invariant(false, "invariant violated")
	}, "Invariant should have panicked")
}

func TestBase_Invariant_FailureWithFormatted(t *testing.T) {
	base := createTestBase(t)

	state := ""
	assert.Panics(t, func() {
		base.Invariant(state != "", "state=%s must not be empty", state)
	}, "Invariant should have panicked")
}

func TestBase_NotNil_Success(t *testing.T) {
	base := createTestBase(t)

	value := 42
	base.NotNil(&value)
}

func TestBase_NotNil_SuccessWithMessage(t *testing.T) {
	base := createTestBase(t)

	value := 42
	base.NotNil(&value, "value must not be nil")
}

func TestBase_NotNil_SuccessWithFormatted(t *testing.T) {
	base := createTestBase(t)

	value := 42
	key := "mykey"
	base.NotNil(&value, "pointer %s must not be nil", key)
}

func TestBase_NotNil_Failure(t *testing.T) {
	base := createTestBase(t)

	var value *int
	assert.Panics(t, func() {
		base.NotNil(value)
	}, "NotNil should have panicked")
}

func TestBase_NotNil_FailureWithMessage(t *testing.T) {
	base := createTestBase(t)

	var value *int
	assert.Panics(t, func() {
		base.NotNil(value, "pointer must not be nil")
	}, "NotNil should have panicked")
}

func TestBase_NotNil_FailureWithFormatted(t *testing.T) {
	base := createTestBase(t)

	var value *int
	key := "badkey"
	assert.Panics(t, func() {
		base.NotNil(value, "pointer %s must not be nil", key)
	}, "NotNil should have panicked")
}

func TestBase_NotEmpty_Success(t *testing.T) {
	base := createTestBase(t)

	base.NotEmpty("hello")
}

func TestBase_NotEmpty_SuccessWithMessage(t *testing.T) {
	base := createTestBase(t)

	base.NotEmpty("hello", "string not empty")
}

func TestBase_NotEmpty_SuccessWithFormatted(t *testing.T) {
	base := createTestBase(t)

	field := "username"
	base.NotEmpty("alice", "field %s must not be empty", field)
}

func TestBase_NotEmpty_Failure(t *testing.T) {
	base := createTestBase(t)

	assert.Panics(t, func() {
		base.NotEmpty("")
	}, "NotEmpty should have panicked")
}

func TestBase_NotEmptySlice_Success(t *testing.T) {
	base := createTestBase(t)

	base.NotEmptySlice([]int{1, 2, 3})
}

func TestBase_NotEmptySlice_SuccessWithMessage(t *testing.T) {
	base := createTestBase(t)

	base.NotEmptySlice([]int{1, 2, 3}, "slice not empty")
}

func TestBase_NotEmptySlice_SuccessWithFormatted(t *testing.T) {
	base := createTestBase(t)

	count := 3
	base.NotEmptySlice([]int{1, 2, 3}, "need at least %d items", count)
}

func TestBase_NotEmptySlice_Failure(t *testing.T) {
	base := createTestBase(t)

	assert.Panics(t, func() {
		base.NotEmptySlice([]int{})
	}, "NotEmptySlice should have panicked")
}

func TestBase_NotEmptySlice_FailureWithMessage(t *testing.T) {
	base := createTestBase(t)

	assert.Panics(t, func() {
		base.NotEmptySlice([]string{}, "data required")
	}, "NotEmptySlice should have panicked")
}

func TestBase_NotEmptySlice_FailureWithFormatted(t *testing.T) {
	base := createTestBase(t)

	expected := 5
	assert.Panics(t, func() {
		base.NotEmptySlice([]int{}, "expected at least %d items", expected)
	}, "NotEmptySlice should have panicked")
}

func TestBase_IntegrationExample(t *testing.T) {
	base := createTestBase(t)

	// Simulate a method that uses Base for contracts and logging
	processValue := func(b *Base, data []int) int {
		b.NotEmptySlice(data)
		b.Log().Debug("processing data", slog.Int("count", len(data)))

		sum := 0
		for _, v := range data {
			sum += v
		}

		b.Ensure(sum >= 0)
		return sum
	}

	result := processValue(&base, []int{1, 2, 3})

	assert.Equal(t, 6, result, "expected sum to be 6")
}

func TestBase_EmbeddedInCustomType(t *testing.T) {
	type Calculator struct {
		Base
	}

	logger, _ := createTestLogger()
	calc := NewBase(logger)

	// Test that embedded methods work
	calc.Require(true)
	calc.Log().Debug("test message")

	// Verify logger is set
	assert.Equal(t, logger, calc.logger, "embedded Base logger not set correctly")
}

func TestBase_ChainedCalls(t *testing.T) {
	base := createTestBase(t)

	// Should successfully chain multiple contract calls
	base.Require(true)
	base.NotNil(base)
	base.NotEmpty("test")
	base.NotEmptySlice([]int{1})
	base.Ensure(true)
	base.Invariant(true)
}
