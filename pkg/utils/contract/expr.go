package contract

import (
	"fmt"
	"strings"
)

// Expr represents a compiled expression using AST-based evaluation.
// It supports comparison operations, logical operators, and variable substitution.
type Expr struct {
	raw       string                 // Original expression string
	ast       ExprAST                // Compiled Abstract Syntax Tree
	variables map[string]interface{} // Variable values
}

// ParseExpr parses an expression string and compiles it to an AST.
// Variables can be referenced with either syntax:
//   - Dollar syntax: $varname
//   - Brace syntax: {varname} (automatically interpolated)
//
// Supported operators: ==, !=, <, >, <=, >=, &&, ||, +, -, *, /
//
// Example expressions:
//
//	"age > 18"              // implicit variable interpolation
//	"{balance} >= 0"         // explicit brace syntax
//	"$legacy > 5"            // dollar syntax
//	"name == \"Alice\""      // string literals
//	"{x} > 5 && {y} < 10"    // multiple variables
func ParseExpr(expr string) (*Expr, error) {
	// Compile the expression to AST
	ast, err := CompileExpression(expr)
	if err != nil {
		return nil, fmt.Errorf("failed to compile expression: %w", err)
	}

	e := &Expr{
		raw:       expr,
		ast:       ast,
		variables: make(map[string]interface{}),
	}

	return e, nil
}

// SetVar sets a variable value for expression evaluation and returns the Expr for chaining.
func (e *Expr) SetVar(name string, value interface{}) *Expr {
	e.variables[name] = value
	return e
}

// SetVars sets multiple variables at once and returns the Expr for chaining.
func (e *Expr) SetVars(vars map[string]interface{}) *Expr {
	for k, v := range vars {
		e.variables[k] = v
	}
	return e
}

// Eval evaluates the expression with the set variables and returns the full evaluation result.
func (e *Expr) Eval() (EvalResult, error) {
	return e.ast.Evaluate(e.variables)
}

// String returns the expression with variable values substituted.
func (e *Expr) String() string {
	result := e.raw
	for name, value := range e.variables {
		valueStr := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, "$"+name, valueStr)
		result = strings.ReplaceAll(result, "{"+name+"}", valueStr)
	}
	return result
}

// getAST returns the compiled AST for this expression.
// This is used internally by contract functions.
func (e *Expr) getAST() ExprAST {
	return e.ast
}

// getVariables returns the variables map for this expression.
// This is used internally by contract functions.
func (e *Expr) getVariables() map[string]interface{} {
	return e.variables
}
