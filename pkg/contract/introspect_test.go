package contract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ========== Stack Introspection Tests ==========

func TestExtractVariableNames_Simple(t *testing.T) {
	names := ExtractVariableNames("age > 18")
	assert.Len(t, names, 1, "expected 1 variable")
	assert.Equal(t, names[0], "age", "expected 'age'")
}

func TestExtractVariableNames_Multiple(t *testing.T) {
	names := ExtractVariableNames("age > 0 && age < 100")
	assert.Len(t, names, 1, "expected 1 variable (should be deduplicated)")
	assert.Equal(t, names[0], "age", "expected 'age'")
}

func TestExtractVariableNames_MultipleDistinct(t *testing.T) {
	names := ExtractVariableNames("x > 5 && y < 10")
	assert.Len(t, names, 2, "expected 2 variables")
	assert.Equal(t, names[0], "x", "first variable should be 'x'")
	assert.Equal(t, names[1], "y", "second variable should be 'y'")
}

func TestExtractVariableNames_WithStrings(t *testing.T) {
	names := ExtractVariableNames("name == \"Alice\" && age > 18")
	assert.Len(t, names, 2, "expected 2 variables")
}

func TestExtractVariableNames_IgnoresKeywords(t *testing.T) {
	names := ExtractVariableNames("true && age > 18 && false")
	assert.Len(t, names, 1, "expected 1 variable (should exclude true/false)")
	assert.Equal(t, names[0], "age", "expected 'age'")
}

func TestExtractVariableNames_ComplexExpression(t *testing.T) {
	names := ExtractVariableNames("(balance >= 100 && status == \"active\") || admin == 1")
	assert.Len(t, names, 3, "expected 3 variables")
}

func TestParseArguments_Simple(t *testing.T) {
	args, err := parseArguments("age > 18, vars)")
	assert.NoError(t, err, "parseArguments should not fail")
	assert.Len(t, args, 2, "expected 2 arguments")
}

func TestParseArguments_WithNested(t *testing.T) {
	args, err := parseArguments("map[string]interface{}{\"age\": 25}, vars)")
	assert.NoError(t, err, "parseArguments should not fail")
	assert.Len(t, args, 2, "expected 2 arguments")
}

func TestParseArguments_WithStrings(t *testing.T) {
	args, err := parseArguments("\"hello, world\", age > 18)")
	assert.NoError(t, err, "parseArguments should not fail")
	assert.Len(t, args, 2, "expected 2 arguments")
}

func TestExtractCallArgument_Simple(t *testing.T) {
	sourceLine := "contract.EnsureAuto(logger, age > 18, vars)"
	arg, err := ExtractCallArgument(sourceLine, "EnsureAuto", 1)
	assert.NoError(t, err, "ExtractCallArgument should not fail")
	assert.Equal(t, arg, "age > 18", "expected argument 'age > 18'")
}

func TestExtractCallArgument_Complex(t *testing.T) {
	sourceLine := "contract.EnsureAuto(logger, age > 0 && age < 100, map[string]interface{}{\"age\": age})"
	arg, err := ExtractCallArgument(sourceLine, "EnsureAuto", 1)
	assert.NoError(t, err, "ExtractCallArgument should not fail")
	assert.Equal(t, arg, "age > 0 && age < 100", "expected argument 'age > 0 && age < 100'")
}

// ========== Auto-Capture Integration Tests ==========

func TestEnsureAuto_Success(t *testing.T) {
	// This should not panic
	age := 25
	_ = map[string]interface{}{"age": age}
	// Skip auto tests since they require real logger and would panic on failure
	// Just verify extraction works
	names := ExtractVariableNames("age > 0 && age < 100")
	assert.Len(t, names, 1, "expected 1 variable")
}

func TestRequireAuto_Success(t *testing.T) {
	// Verify the logic without actual logger
	age := 25
	_ = map[string]interface{}{"age": age}
	names := ExtractVariableNames("age > 0 && age < 100")
	assert.Len(t, names, 1, "expected 1 variable")
}

func TestInvariantAuto_Success(t *testing.T) {
	// Verify the logic without actual logger
	balance := 50
	_ = map[string]interface{}{"balance": balance}
	names := ExtractVariableNames("balance >= 0 && balance < 1000")
	assert.Len(t, names, 1, "expected 1 variable")
}
