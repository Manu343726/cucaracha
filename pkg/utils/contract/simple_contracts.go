package contract

import (
	"fmt"

	"github.com/Manu343726/cucaracha/pkg/utils/logging"
)

// parseVarArgs parses variadic arguments for *That functions.
// Returns (variables map, user message, or empty vars and message if no variables provided)
// Detects if first arg is a variables map/struct or a message.
func parseVarArgs(args ...any) (map[string]interface{}, []any) {
	if len(args) == 0 {
		return nil, nil
	}

	// Check if first arg is a map or struct (variables), not a string (message)
	var variables map[string]interface{}
	var messageArgs []any

	// Try to interpret first arg as variables
	if varMap, ok := args[0].(map[string]interface{}); ok {
		variables = varMap
		messageArgs = args[1:]
	} else if _, ok := args[0].(string); !ok {
		// Not a string, might be a struct with variables
		// Use reflection through ToVariableMap
		if ptrVal, isPtrToStruct := args[0].(*struct{}); isPtrToStruct {
			variables = ToVariableMap(ptrVal)
			messageArgs = args[1:]
		} else if structVal, isStruct := args[0].(struct{}); isStruct {
			variables = ToVariableMap(structVal)
			messageArgs = args[1:]
		} else {
			// Assume it's just message args
			messageArgs = args
		}
	} else {
		// First arg is a string or message-like
		messageArgs = args
	}

	return variables, messageArgs
}

// checkSimpleWithType extracts source, compiles to AST, and evaluates with contract type.
// If variables are provided, they override source extraction.
func checkSimpleWithTypeWithVars(logger *logging.Logger, condition bool, contractType, failureMsg string, variables map[string]interface{}, userMessage ...any) {
	if !condition {
		var exprStr string
		var vars map[string]interface{}

		// If variables provided, use them; otherwise extract from source
		if variables != nil {
			vars = variables
			// Still need to extract expression from source
			info, _ := GetCallerInfo(1)
			exprStr, _ = ExtractVariableContext(info)
		} else {
			info, _ := GetCallerInfo(1)
			exprStr, vars = ExtractVariableContext(info)
		}

		if exprStr == "" {
			msg := failureMsg + ": (unable to extract source)"
			if len(userMessage) > 0 {
				msg += "\n\n" + FormatErrorMessage("", userMessage...)
			}
			logger.Panic(msg)
			return
		}

		// Compile expression to AST
		ast, err := CompileExpression(exprStr)
		if err != nil {
			msg := fmt.Sprintf("%s: %s (parse error: %v)", failureMsg, exprStr, err)
			if len(userMessage) > 0 {
				msg += "\n\n" + FormatErrorMessage("", userMessage...)
			}
			logger.Panic(msg)
			return
		}

		// Create Expr and evaluate
		expr := &Expr{
			raw:       exprStr,
			ast:       ast,
			variables: vars,
		}
		checkExprWithType(logger, expr, contractType, failureMsg, userMessage)
	}
}

// EnsureThat checks a postcondition with AST-based evaluation.
// Can optionally pass variables map/struct and/or custom error message:
//
//	EnsureThat(logger, condition)
//	EnsureThat(logger, condition, map[string]interface{}{"x": 5})
//	EnsureThat(logger, condition, "custom error message")
//	EnsureThat(logger, condition, map[string]interface{}{"x": 5}, "custom message")
func EnsureThat(logger *logging.Logger, condition bool, args ...any) {
	vars, msg := parseVarArgs(args...)
	checkSimpleWithTypeWithVars(logger, condition, "Ensure", "postcondition failed", vars, msg)
}

// ExpectThat checks a precondition with AST-based evaluation.
// Can optionally pass variables map/struct and/or custom error message:
//
//	ExpectThat(logger, condition)
//	ExpectThat(logger, condition, map[string]interface{}{"x": 5})
//	ExpectThat(logger, condition, "custom error message")
//	ExpectThat(logger, condition, map[string]interface{}{"x": 5}, "custom message")
func ExpectThat(logger *logging.Logger, condition bool, args ...any) {
	vars, msg := parseVarArgs(args...)
	checkSimpleWithTypeWithVars(logger, condition, "Expect", "precondition failed", vars, msg)
}

// InvariantThat checks an invariant with AST-based evaluation.
// Can optionally pass variables map/struct and/or custom error message:
//
//	InvariantThat(logger, condition)
//	InvariantThat(logger, condition, map[string]interface{}{"x": 5})
//	InvariantThat(logger, condition, "custom error message")
//	InvariantThat(logger, condition, map[string]interface{}{"x": 5}, "custom message")
func InvariantThat(logger *logging.Logger, condition bool, args ...any) {
	vars, msg := parseVarArgs(args...)
	checkSimpleWithTypeWithVars(logger, condition, "Invariant", "invariant failed", vars, msg)
}

// RequireThat checks a precondition with AST-based evaluation and automatic variable capture.
// This is a simplified version that only captures variables if the first argument is a map or struct.
// For more complex cases, users can use RequireExpr with manual variable passing.
//
//	RequireThat(logger, condition)
//	RequireThat(logger, condition, map[string]interface{}{"x": 5})
//	RequireThat(logger, condition, "custom error message")
//	RequireThat(logger, condition, map[string]interface{}{"x": 5}, "custom message")
func RequireThat(logger *logging.Logger, condition bool, args ...any) {
	vars, msg := parseVarArgs(args...)
	checkSimpleWithTypeWithVars(logger, condition, "Require", "precondition failed", vars, msg)
}
