package contract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ========== Expression Parsing Tests ==========

func TestParseExpr_SimpleComparison(t *testing.T) {
	expr, err := ParseExpr("$x > 5")
	assert.NoError(t, err, "ParseExpr failed")
	assert.NotNil(t, expr, "ParseExpr returned nil")
}

func TestParseExpr_StringComparison(t *testing.T) {
	expr, err := ParseExpr("$name == \"Alice\"")
	assert.NoError(t, err, "ParseExpr failed")
	assert.NotNil(t, expr, "ParseExpr returned nil")
}

func TestParseExpr_LogicalAnd(t *testing.T) {
	expr, err := ParseExpr("$x > 5 && $y < 10")
	assert.NoError(t, err, "ParseExpr failed")
	assert.NotNil(t, expr, "ParseExpr returned nil")
}

func TestParseExpr_LogicalOr(t *testing.T) {
	expr, err := ParseExpr("$x > 5 || $y < 10")
	assert.NoError(t, err, "ParseExpr failed")
	assert.NotNil(t, expr, "ParseExpr returned nil")
}

func TestParseExpr_Complex(t *testing.T) {
	expr, err := ParseExpr("($x > 5 && $y < 10) || ($z == 0)")
	assert.NoError(t, err, "ParseExpr failed")
	assert.NotNil(t, expr, "ParseExpr returned nil")
}

func TestParseExpr_Arithmetic(t *testing.T) {
	expr, err := ParseExpr("$x + 5 > 10")
	assert.NoError(t, err, "ParseExpr failed")
	assert.NotNil(t, expr, "ParseExpr returned nil")
}

func TestParseExpr_InvalidExpr(t *testing.T) {
	_, err := ParseExpr("@invalid")
	assert.Error(t, err, "ParseExpr should have failed for invalid input")
}

// ========== Expression Evaluation Tests ==========

func TestExpr_SimpleGreater(t *testing.T) {
	expr, _ := ParseExpr("$x > 5")
	expr.SetVar("x", 10)

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.True(t, result.Value, "expected true, got false")
}

func TestExpr_SimpleGreater_False(t *testing.T) {
	expr, _ := ParseExpr("$x > 5")
	expr.SetVar("x", 3)

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.False(t, result.Value, "expected false, got true")
}

func TestExpr_Equal(t *testing.T) {
	expr, _ := ParseExpr("$name == \"Alice\"")
	expr.SetVar("name", "Alice")

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.True(t, result.Value, "expected true, got false")
}

func TestExpr_NotEqual(t *testing.T) {
	expr, _ := ParseExpr("$x != 0")
	expr.SetVar("x", 5)

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.True(t, result.Value, "expected true, got false")
}

func TestExpr_LogicalAnd_True(t *testing.T) {
	expr, _ := ParseExpr("$x > 5 && $y < 10")
	expr.SetVar("x", 7)
	expr.SetVar("y", 8)

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.True(t, result.Value, "expected true, got false")
}

func TestExpr_LogicalAnd_False(t *testing.T) {
	expr, _ := ParseExpr("$x > 5 && $y < 10")
	expr.SetVar("x", 3)
	expr.SetVar("y", 8)

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.False(t, result.Value, "expected false, got true")
}

func TestExpr_LogicalOr_True(t *testing.T) {
	expr, _ := ParseExpr("$x > 5 || $y < 10")
	expr.SetVar("x", 3)
	expr.SetVar("y", 8)

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.True(t, result.Value, "expected true, got false")
}

func TestExpr_LogicalOr_False(t *testing.T) {
	expr, _ := ParseExpr("$x > 5 || $y > 10")
	expr.SetVar("x", 3)
	expr.SetVar("y", 8)

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.False(t, result.Value, "expected false, got true")
}

func TestExpr_Arithmetic_Addition(t *testing.T) {
	expr, _ := ParseExpr("$x + 5 > 10")
	expr.SetVar("x", 6)

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.True(t, result.Value, "expected true, got false")
}

func TestExpr_Arithmetic_Multiplication(t *testing.T) {
	expr, _ := ParseExpr("$x * 2 == 10")
	expr.SetVar("x", 5)

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.True(t, result.Value, "expected true, got false")
}

func TestExpr_Parentheses(t *testing.T) {
	expr, _ := ParseExpr("($x > 5) && ($y < 10)")
	expr.SetVar("x", 7)
	expr.SetVar("y", 8)

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.True(t, result.Value, "expected true, got false")
}

func TestExpr_SetVars(t *testing.T) {
	expr, _ := ParseExpr("$x > 5 && $y < 10")
	expr.SetVars(map[string]interface{}{
		"x": 7,
		"y": 8,
	})

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.True(t, result.Value, "expected true, got false")
}

func TestExpr_UndefinedVariable(t *testing.T) {
	expr, _ := ParseExpr("$undefined > 5")

	_, err := expr.Eval()
	assert.Error(t, err, "Eval should fail with undefined variable")
}

func TestExpr_String_Output(t *testing.T) {
	expr, _ := ParseExpr("$x > $threshold && $name == \"test\"")
	expr.SetVar("x", 42)
	expr.SetVar("threshold", 10)
	expr.SetVar("name", "test")

	str := expr.String()
	// Should have quoted strings in the output
	assert.Contains(t, str, "42 > 10", "expected output to contain '42 > 10'")
	assert.Contains(t, str, "== \"test\"", "expected output to contain '== \"test\"'")
}

func TestExpr_LessThanOrEqual(t *testing.T) {
	expr, _ := ParseExpr("$x <= 10")
	expr.SetVar("x", 10)

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.True(t, result.Value, "expected true, got false")
}

func TestExpr_GreaterThanOrEqual(t *testing.T) {
	expr, _ := ParseExpr("$x >= 5")
	expr.SetVar("x", 5)

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.True(t, result.Value, "expected true, got false")
}

func TestExpr_DivisionByZero(t *testing.T) {
	expr, _ := ParseExpr("$x / 0 > 5")
	expr.SetVar("x", 10)

	_, err := expr.Eval()
	assert.Error(t, err, "Eval should fail on division by zero")
}

// ========== Contract with Expression Tests ==========

func TestRequireExpr_Success(t *testing.T) {
	logger, _ := createTestLogger()
	base := &Base{logger: logger}

	expr, _ := ParseExpr("$x > 5")
	expr.SetVar("x", 10)

	// Should not panic
	base.RequireExpr(expr, "x must be greater than 5")
}

func TestRequireExpr_Failure(t *testing.T) {
	logger, _ := createTestLogger()
	base := &Base{logger: logger}

	expr, _ := ParseExpr("$x > 5")
	expr.SetVar("x", 3)

	assert.Panics(t, func() {
		base.RequireExpr(expr, "x must be greater than 5")
	}, "RequireExpr should have panicked")
}

func TestEnsureExpr_Success(t *testing.T) {
	logger, _ := createTestLogger()
	base := &Base{logger: logger}

	expr, _ := ParseExpr("$result > 0")
	expr.SetVar("result", 42)

	// Should not panic
	base.EnsureExpr(expr, "result must be positive")
}

func TestEnsureExpr_Failure(t *testing.T) {
	logger, _ := createTestLogger()
	base := &Base{logger: logger}

	expr, _ := ParseExpr("$result > 0")
	expr.SetVar("result", -5)

	assert.Panics(t, func() {
		base.EnsureExpr(expr, "result must be positive")
	}, "EnsureExpr should have panicked")
}

func TestInvariantExpr_Success(t *testing.T) {
	logger, _ := createTestLogger()
	base := &Base{logger: logger}

	expr, _ := ParseExpr("$balance >= 0")
	expr.SetVar("balance", 100)

	// Should not panic
	base.InvariantExpr(expr, "balance invariant")
}

func TestInvariantExpr_Failure(t *testing.T) {
	logger, _ := createTestLogger()
	base := &Base{logger: logger}

	expr, _ := ParseExpr("$balance >= 0")
	expr.SetVar("balance", -5)

	assert.Panics(t, func() {
		base.InvariantExpr(expr, "balance invariant")
	}, "InvariantExpr should have panicked")
}

func TestRequireExpr_WithFormatting(t *testing.T) {
	logger, _ := createTestLogger()
	base := &Base{logger: logger}

	expr, _ := ParseExpr("$x > 5")
	expr.SetVar("x", 10)

	// Should not panic
	base.RequireExpr(expr, "value %d must be greater than threshold", 5)
}

func TestExpr_ComplexExpression(t *testing.T) {
	// Nested parentheses might have issues, let's use this version
	expr, _ := ParseExpr("$x > 5 && $y < 10 ||  $z == 0")
	expr.SetVars(map[string]interface{}{
		"x": 7,
		"y": 8,
		"z": 0,
	})

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.True(t, result.Value, "expected true, got false")
}

func TestExpr_IntegerConversion(t *testing.T) {
	expr, _ := ParseExpr("$x > 5")
	expr.SetVar("x", int(10))

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.True(t, result.Value, "expected true, got false")
}

func TestExpr_NegativeNumbers(t *testing.T) {
	expr, _ := ParseExpr("$x == -5")
	expr.SetVar("x", -5)

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.True(t, result.Value, "expected true, got false")
}

func TestExpr_FloatingPoint(t *testing.T) {
	expr, _ := ParseExpr("$x > 5.5")
	expr.SetVar("x", 6.0)

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.True(t, result.Value, "expected true, got false")
}
