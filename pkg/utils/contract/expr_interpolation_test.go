package contract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ========== Brace Interpolation Syntax Tests ==========

func TestParseExpr_BraceSyntax_Simple(t *testing.T) {
	expr, err := ParseExpr("{x} > 5")
	assert.NoError(t, err, "ParseExpr with brace syntax failed")
	assert.NotNil(t, expr, "ParseExpr returned nil")
}

func TestExpr_BraceSyntax_Evaluation(t *testing.T) {
	expr, _ := ParseExpr("{x} > 5")
	expr.SetVar("x", 10)

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.True(t, result.Value, "expected true, got false")
}

func TestExpr_BraceSyntax_MultipleVars(t *testing.T) {
	expr, _ := ParseExpr("{x} > 5 && {y} < 10")
	expr.SetVar("x", 7).SetVar("y", 8)

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.True(t, result.Value, "expected true, got false")
}

func TestExpr_BraceSyntax_WithStrings(t *testing.T) {
	expr, _ := ParseExpr("{name} == \"Alice\"")
	expr.SetVar("name", "Alice")

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.True(t, result.Value, "expected true, got false")
}

func TestExpr_BraceSyntax_WithArithmetic(t *testing.T) {
	expr, _ := ParseExpr("{x} + 5 > 10")
	expr.SetVar("x", 10)

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.True(t, result.Value, "expected true, got false")
}

func TestExpr_MixedSyntax_DollarAndBrace(t *testing.T) {
	expr, _ := ParseExpr("$x > 5 && {y} < 10")
	expr.SetVar("x", 7).SetVar("y", 8)

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.True(t, result.Value, "expected true, got false")
}

func TestExpr_BraceSyntax_StringInterpolation(t *testing.T) {
	expr, _ := ParseExpr("{age} > 18")
	expr.SetVar("age", 21)

	strResult := expr.String()
	assert.Equal(t, "21 > 18", strResult, "string interpolation mismatch")
}

func TestExpr_BraceSyntax_Chaining(t *testing.T) {
	expr, _ := ParseExpr("{x} > 5 && {y} < 10 && {z} == 0")

	// Test method chaining
	expr.SetVar("x", 7).SetVar("y", 8).SetVar("z", 0)

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.True(t, result.Value, "expected true, got false")
}

func TestExpr_BraceSyntax_ComplexExpression(t *testing.T) {
	expr, _ := ParseExpr("{balance} >= 100 || {status} == \"active\"")
	expr.SetVar("balance", 50).
		SetVar("status", "active")

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.True(t, result.Value, "expected true, got false")
}

func TestExpr_BraceSyntax_Error_MissingVar(t *testing.T) {
	expr, _ := ParseExpr("{missing} > 5")
	expr.SetVar("other", 10)

	_, err := expr.Eval()
	assert.Error(t, err, "expected error for missing variable")
}

func TestExpr_BraceSyntax_Negative(t *testing.T) {
	expr, _ := ParseExpr("{x} > -5")
	expr.SetVar("x", -3)

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.True(t, result.Value, "expected true, got false")
}

func TestExpr_BraceSyntax_FloatingPoint(t *testing.T) {
	expr, _ := ParseExpr("{x} > 5.5")
	expr.SetVar("x", 6.0)

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.True(t, result.Value, "expected true, got false")
}

func TestExpr_BraceSyntax_OrOperator(t *testing.T) {
	expr, _ := ParseExpr("{x} > 5 || {y} < 10")
	expr.SetVar("x", 2).SetVar("y", 3)

	result, err := expr.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.True(t, result.Value, "expected true, got false")
}

func TestExpr_SetVars_ReturnsExpr(t *testing.T) {
	expr, _ := ParseExpr("{x} > 5 && {y} < 10")

	// Test that SetVars returns the Expr for chaining
	result := expr.SetVars(map[string]interface{}{
		"x": 7,
		"y": 8,
	})

	assert.Equal(t, expr, result, "SetVars should return the same Expr instance")

	evalResult, err := result.Eval()
	assert.NoError(t, err, "Eval failed")
	assert.True(t, evalResult.Value, "expected true, got false")
}
