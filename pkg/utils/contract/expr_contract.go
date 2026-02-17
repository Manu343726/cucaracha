package contract

import (
	"fmt"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/utils/logging"
)

// formatASTFailure creates a detailed failure message from an AST evaluation result.
func formatASTFailure(contractType string, result EvalResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s: %s\n", contractType, result.Expression))

	// Write variable values in a consistent order
	if len(result.Variables) > 0 {
		for name, value := range result.Variables {
			sb.WriteString(fmt.Sprintf("  %s = %v\n", name, value))
		}
	}

	// Show analysis of compound expressions
	if len(result.Parts) > 0 {
		sb.WriteString("\n  Analysis:\n")
		for i, part := range result.Parts {
			symbol := "✓"
			if !part.Value {
				symbol = "✗"
			}

			var line string
			if i == 0 {
				line = fmt.Sprintf("    %s %s", symbol, part.Expression)
			} else {
				line = fmt.Sprintf("    %s %s %s", part.Operator, symbol, part.Expression)
			}

			if !part.Value && (part.Operator == "&&" || part.Operator == "") {
				line += " ← FAILED"
			}

			sb.WriteString(line + "\n")
		}
	}

	return sb.String()
}

// checkExprWithType evaluates an Expr and panics with a specific contract type message.
// contractType is used in the detailed output: "Ensure", "Expect", "Invariant"
// failureMsg is the panic message prefix: "postcondition failed", "precondition failed", "invariant failed"
// userMessage is optional additional context to append after the analysis
func checkExprWithType(logger *logging.Logger, expr *Expr, contractType, failureMsg string, userMessage ...any) {
	result, err := expr.ast.Evaluate(expr.variables)
	if err != nil {
		logger.Panic(fmt.Sprintf("%s: failed to evaluate expression: %v", contractType, err))
	}

	if !result.Value {
		details := formatASTFailure(contractType, result)
		message := fmt.Sprintf("%s:\n%s", failureMsg, details)
		if len(userMessage) > 0 {
			message += "\n\n" + FormatErrorMessage("", userMessage...)
		}
		logger.Panic(message)
	}
}

// RequireExpr checks a precondition expressed as a parsed expression.
func RequireExpr(logger *logging.Logger, expr *Expr, fmtAndArgs ...any) {
	checkExprWithType(logger, expr, "Require", "precondition failed", fmtAndArgs...)
}

// EnsureExpr checks a postcondition expressed as a parsed expression.
func EnsureExpr(logger *logging.Logger, expr *Expr, fmtAndArgs ...any) {
	checkExprWithType(logger, expr, "Ensure", "postcondition failed", fmtAndArgs...)
}

// InvariantExpr checks an invariant expressed as a parsed expression.
func InvariantExpr(logger *logging.Logger, expr *Expr, fmtAndArgs ...any) {
	checkExprWithType(logger, expr, "Invariant", "invariant failed", fmtAndArgs...)
}
