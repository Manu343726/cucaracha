package contract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEnsureThat_Success demonstrates that EnsureThat doesn't panic when condition is true
func TestEnsureThat_Success(t *testing.T) {
	logger, _ := createTestLogger()
	age := 25
	// Should not panic
	EnsureThat(logger, age > 0 && age < 100)
}

// TestEnsureThat_Failure demonstrates that EnsureThat panics when condition is false
func TestEnsureThat_Failure(t *testing.T) {
	logger, _ := createTestLogger()
	age := 42
	// Should panic with detailed message
	assert.Panics(t, func() {
		EnsureThat(logger, age > 0 && age < 25)
	}, "EnsureThat should panic when condition is false")
}

// TestExpectThat_Success demonstrates that ExpectThat doesn't panic when condition is true
func TestExpectThat_Success(t *testing.T) {
	logger, _ := createTestLogger()
	age := 25
	// Should not panic
	ExpectThat(logger, age >= 18)
}

// TestExpectThat_Failure demonstrates failure with precondition
func TestExpectThat_Failure(t *testing.T) {
	logger, _ := createTestLogger()
	age := 16
	// Should panic
	assert.Panics(t, func() {
		ExpectThat(logger, age >= 18)
	}, "ExpectThat should panic when condition is false")
}

// TestInvariantThat_Success demonstrates that InvariantThat doesn't panic when condition is true
func TestInvariantThat_Success(t *testing.T) {
	logger, _ := createTestLogger()
	balance := 100
	// Should not panic
	InvariantThat(logger, balance >= 0)
}

// TestInvariantThat_Failure demonstrates invariant violation
func TestInvariantThat_Failure(t *testing.T) {
	logger, _ := createTestLogger()
	balance := -50
	// Should panic
	assert.Panics(t, func() {
		InvariantThat(logger, balance >= 0)
	}, "InvariantThat should panic when condition is false")
}

// TestCompileExpression_SimpleComparison tests basic expression compilation to AST
func TestCompileExpression_SimpleComparison(t *testing.T) {
	ast, err := CompileExpression("x > 5")
	assert.NoError(t, err, "should compile simple comparison")
	assert.NotNil(t, ast, "AST should not be nil")
}

// TestCompileExpression_ComplexExpression tests compound expression compilation
func TestCompileExpression_ComplexExpression(t *testing.T) {
	ast, err := CompileExpression("x > 0 && x < 10")
	assert.NoError(t, err, "should compile complex expression")
	assert.NotNil(t, ast, "AST should not be nil")
}

// TestEvaluateAST_SimpleComparison tests AST evaluation
func TestEvaluateAST_SimpleComparison(t *testing.T) {
	ast, _ := CompileExpression("x > 5")
	vars := map[string]interface{}{"x": 10}
	result, err := ast.Evaluate(vars)

	assert.NoError(t, err, "evaluation should succeed")
	assert.True(t, result.Value, "10 > 5 should be true")
}

// TestEvaluateAST_FailingCondition tests AST evaluation with false result
func TestEvaluateAST_FailingCondition(t *testing.T) {
	ast, _ := CompileExpression("x > 5")
	vars := map[string]interface{}{"x": 3}
	result, err := ast.Evaluate(vars)

	assert.NoError(t, err, "evaluation should succeed")
	assert.False(t, result.Value, "3 > 5 should be false")
}

// TestEvaluateAST_CompoundExpression tests complex expression evaluation
func TestEvaluateAST_CompoundExpression(t *testing.T) {
	ast, _ := CompileExpression("x > 0 && x < 10")
	vars := map[string]interface{}{"x": 5}
	result, err := ast.Evaluate(vars)

	assert.NoError(t, err, "evaluation should succeed")
	assert.True(t, result.Value, "5 > 0 && 5 < 10 should be true")
	assert.NotEmpty(t, result.Parts, "compound expression should have parts")
}

// TestFormatASTFailure tests failure message formatting
func TestFormatASTFailure_SimpleCondition(t *testing.T) {
	ast, _ := CompileExpression("x > 5")
	vars := map[string]interface{}{"x": 3}
	result, _ := ast.Evaluate(vars)

	msg := formatASTFailure("Ensure", result)
	assert.NotEmpty(t, msg, "failure message should not be empty")
	assert.Contains(t, msg, "Ensure", "message should contain contract type")
	assert.Contains(t, msg, "x", "message should contain variable name")
}

// TestAST_ExtractVariableNames tests variable extraction from expressions
func TestAST_ExtractVariableNames_Simple(t *testing.T) {
	names := ExtractVariableNames("x > 5")
	assert.Contains(t, names, "x", "should extract variable x")
}

// TestAST_ExtractVariableNames_Multiple tests extracting multiple variables
func TestAST_ExtractVariableNames_Multiple(t *testing.T) {
	names := ExtractVariableNames("age > 0 && age < 100 && status != \"banned\"")
	assert.Contains(t, names, "age", "should extract variable age")
	assert.Contains(t, names, "status", "should extract variable status")
}
