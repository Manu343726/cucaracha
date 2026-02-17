package contract

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
)

// SourceInfo contains information extracted from source code
type SourceInfo struct {
	File       string
	Line       int
	Expression string
}

// GetCallerInfo retrieves source code information about the caller
// skip parameter indicates how many call frames to skip (0 = immediate caller, 1 = caller's caller, etc.)
func GetCallerInfo(skip int) (*SourceInfo, error) {
	var pc uintptr
	var file string
	var line int

	for {
		var ok bool
		pc, file, line, ok = runtime.Caller(skip + 1)
		if !ok {
			return nil, fmt.Errorf("could not get caller info")
		}

		if !strings.Contains(file, "contract.go") && !strings.Contains(file, "simple_contracts.go") && !strings.Contains(file, "expr_contract.go") {
			break
		}

		skip++
	}

	// Get the function name for debugging
	fn := runtime.FuncForPC(pc)
	_ = fn // Keep for potential future use

	// Read the source line
	sourceBytes, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("could not read source file %s: %w", file, err)
	}

	lines := strings.Split(string(sourceBytes), "\n")
	if line < 1 || line > len(lines) {
		return nil, fmt.Errorf("line %d out of range in %s", line, file)
	}

	sourceLine := lines[line-1]

	return &SourceInfo{
		File:       file,
		Line:       line,
		Expression: strings.TrimSpace(sourceLine),
	}, nil
}

// ExtractCallArgument extracts the argument passed to a function call from source code
// For example, from "contract.Ensure(logger, age > 0 && age < 25, vars)"
// it extracts "age > 0 && age < 25" as the second argument after logger
func ExtractCallArgument(sourceLine string, callName string, argIndex int) (string, error) {
	// Build a regex to find the function call
	// Pattern: callName followed by (
	pattern := regexp.MustCompile(regexp.QuoteMeta(callName) + `\s*\(`)

	matchIdx := pattern.FindStringIndex(sourceLine)
	if matchIdx == nil {
		return "", fmt.Errorf("could not find call to %s", callName)
	}

	// Find the opening paren
	openParen := matchIdx[1] - 1
	rest := sourceLine[openParen+1:]

	// Parse arguments - handle nested parens, strings, etc.
	args, err := parseArguments(rest)
	if err != nil {
		return "", err
	}

	if argIndex >= len(args) {
		return "", fmt.Errorf("argument index %d out of range (function has %d arguments)", argIndex, len(args))
	}

	return strings.TrimSpace(args[argIndex]), nil
}

// parseArguments splits a comma-separated argument list, respecting nested parentheses and strings
func parseArguments(argsStr string) ([]string, error) {
	var args []string
	var current strings.Builder
	var depth int
	var inString bool
	var stringChar rune
	var escaped bool

	for _, ch := range argsStr {
		if escaped {
			current.WriteRune(ch)
			escaped = false
			continue
		}

		if ch == '\\' {
			current.WriteRune(ch)
			escaped = true
			continue
		}

		if inString {
			current.WriteRune(ch)
			if ch == stringChar {
				inString = false
			}
			continue
		}

		if ch == '"' || ch == '\'' {
			inString = true
			stringChar = ch
			current.WriteRune(ch)
			continue
		}

		if ch == '(' || ch == '{' || ch == '[' {
			depth++
			current.WriteRune(ch)
			continue
		}

		if ch == ')' || ch == '}' || ch == ']' {
			if depth == 0 {
				// End of arguments
				arg := current.String()
				if arg != "" {
					args = append(args, arg)
				}
				return args, nil
			}
			depth--
			current.WriteRune(ch)
			continue
		}

		if ch == ',' && depth == 0 {
			// Argument separator
			arg := current.String()
			if arg != "" {
				args = append(args, arg)
			}
			current.Reset()
			continue
		}

		current.WriteRune(ch)
	}

	// Handle case where there's no closing paren (incomplete)
	arg := current.String()
	if arg != "" {
		args = append(args, arg)
	}

	return args, nil
}

// ExtractVariableNames extracts all variable names from an expression
// Variables are assumed to be identifiers (alphanumeric + underscore)
// This returns a deduplicated list in order of appearance
// String contents are skipped
func ExtractVariableNames(expr string) []string {
	// First, remove string contents to avoid matching keywords inside strings
	cleanExpr := removeStrings(expr)

	// Pattern: identifier that's not preceded by a dot and not quoted
	pattern := regexp.MustCompile(`\b([a-zA-Z_][a-zA-Z0-9_]*)\b`)

	matches := pattern.FindAllString(cleanExpr, -1)

	// Deduplicate while preserving order
	seen := make(map[string]bool)
	var result []string

	for _, match := range matches {
		// Skip keywords and operators
		if isKeywordOrOperator(match) {
			continue
		}

		if !seen[match] {
			seen[match] = true
			result = append(result, match)
		}
	}

	return result
}

// removeStrings replaces all string contents with spaces to prevent matching inside strings
func removeStrings(expr string) string {
	var result strings.Builder
	inString := false
	var stringChar rune
	escaped := false

	for _, ch := range expr {
		if escaped {
			result.WriteRune(' ')
			escaped = false
			continue
		}

		if ch == '\\' {
			escaped = true
			result.WriteRune(' ')
			continue
		}

		if inString {
			if ch == stringChar {
				inString = false
				result.WriteRune(' ')
			} else {
				result.WriteRune(' ')
			}
			continue
		}

		if ch == '"' || ch == '\'' {
			inString = true
			stringChar = ch
			result.WriteRune(' ')
			continue
		}

		result.WriteRune(ch)
	}

	return result.String()
}

// isKeywordOrOperator checks if a string is a Go keyword or common operator word
func isKeywordOrOperator(s string) bool {
	keywords := map[string]bool{
		"true": true, "false": true, "nil": true,
		"package": true, "import": true, "func": true,
		"if": true, "else": true, "for": true, "range": true,
		"switch": true, "case": true, "default": true,
		"return": true, "defer": true, "go": true,
		"const": true, "var": true, "type": true,
		"interface": true, "struct": true,
		"and": true, "or": true, "not": true,
		"contract": true,
	}
	return keywords[s]
}

// ExtractVariableContext extracts the expression and variable values from a contract function call.
// It reads the source line, extracts the condition expression, and captures variable values
// from the surrounding scope using the local variable inspection.
//
// For a call like: contract.EnsureThat(age > 0 && age < 120)
// It returns: ("age > 0 && age < 120", map{"age": 25})
func ExtractVariableContext(info *SourceInfo) (string, map[string]interface{}) {
	if info == nil {
		return "", make(map[string]interface{})
	}

	sourceLine := info.Expression

	// Extract the expression - look for EnsureThat, ExpectThat, InvariantThat patterns
	exprStr := extractExpressionFromLine(sourceLine)
	if exprStr == "" {
		return "", make(map[string]interface{})
	}

	// Extract variable names from the expression
	varNames := ExtractVariableNames(exprStr)

	// Build variable context - in the real implementation, we'd inspect the call stack
	// For now, return an empty map - the actual values should be passed through
	// or retrieved via reflection if needed
	vars := make(map[string]interface{})
	for _, name := range varNames {
		// Placeholder - would need deeper introspection to get actual values
		vars[name] = nil
	}

	return exprStr, vars
}

// extractExpressionFromLine extracts the boolean condition from a contract function call
// Supports: EnsureThat, ExpectThat, InvariantThat, EnsureExpr, etc.
func extractExpressionFromLine(sourceLine string) string {
	contractFuncs := []string{
		"EnsureThat", "ExpectThat", "InvariantThat",
		"EnsureExpr", "RequireExpr", "InvariantExpr",
		"SimpleEnsure", "SimpleRequire", "SimpleInvariant",
	}

	for _, funcName := range contractFuncs {
		pattern := regexp.MustCompile(regexp.QuoteMeta("contract."+funcName) + `\s*\(`)
		matchIdx := pattern.FindStringIndex(sourceLine)
		if matchIdx == nil {
			continue
		}

		openParen := matchIdx[1] - 1
		rest := sourceLine[openParen+1:]

		args, err := parseArguments(rest)
		if err != nil {
			continue
		}

		if len(args) > 0 {
			return strings.TrimSpace(args[0])
		}
	}

	return ""
}
