package contract

import (
	"fmt"
	"strconv"
	"strings"
)

// ExprAST represents an Abstract Syntax Tree for contract expressions.
// All expressions can be compiled down to this AST format for consistent evaluation.
type ExprAST interface {
	Evaluate(vars map[string]interface{}) (EvalResult, error)
	String() string
}

// EvalResult contains the result of evaluating an expression along with detailed information.
type EvalResult struct {
	Value      bool                   // The final boolean result
	Parts      []PartResult           // Results of each part in compound expressions
	Expression string                 // Original expression string
	Variables  map[string]interface{} // Variable values used in evaluation
	Error      error                  // Any evaluation error
}

// PartResult describes the evaluation of a single part of a compound expression.
type PartResult struct {
	Expression string                 // The sub-expression
	Value      bool                   // Result of this part
	Variables  map[string]interface{} // Variables used in this part
	Operator   string                 // Operator before this part (&& or ||)
}

// BinaryOpAST represents a binary operation (comparison or logical).
type BinaryOpAST struct {
	Left     ExprAST
	Right    ExprAST
	Operator string // "==", "!=", "<", ">", "<=", ">=", "&&", "||", "+", "-", "*", "/"
	LeftStr  string // String representation of left side
	RightStr string // String representation of right side
}

func (b *BinaryOpAST) Evaluate(vars map[string]interface{}) (EvalResult, error) {
	result := EvalResult{
		Expression: fmt.Sprintf("%s %s %s", b.LeftStr, b.Operator, b.RightStr),
		Variables:  vars,
	}

	// For logical operators, we might short-circuit
	if b.Operator == "&&" || b.Operator == "||" {
		leftResult, err := b.Left.Evaluate(vars)
		if err != nil {
			result.Error = err
			return result, err
		}

		rightResult, err := b.Right.Evaluate(vars)
		if err != nil {
			result.Error = err
			return result, err
		}

		if b.Operator == "&&" {
			result.Value = leftResult.Value && rightResult.Value
		} else { // "||"
			result.Value = leftResult.Value || rightResult.Value
		}

		result.Parts = []PartResult{
			{Expression: leftResult.Expression, Value: leftResult.Value, Operator: ""},
			{Expression: rightResult.Expression, Value: rightResult.Value, Operator: b.Operator},
		}

		return result, nil
	}

	// For comparison/arithmetic operators
	leftVal, err := evaluateValue(b.Left, vars)
	if err != nil {
		result.Error = err
		return result, err
	}

	rightVal, err := evaluateValue(b.Right, vars)
	if err != nil {
		result.Error = err
		return result, err
	}

	switch b.Operator {
	case ">":
		result.Value = compareGreater(leftVal, rightVal)
	case "<":
		result.Value = compareLess(leftVal, rightVal)
	case ">=":
		result.Value = compareGreaterEqual(leftVal, rightVal)
	case "<=":
		result.Value = compareLessEqual(leftVal, rightVal)
	case "==":
		result.Value = compareEqual(leftVal, rightVal)
	case "!=":
		result.Value = compareNotEqual(leftVal, rightVal)
	case "+":
		leftVal, rightVal = arithmetic(leftVal, rightVal, "+")
	case "-":
		leftVal, rightVal = arithmetic(leftVal, rightVal, "-")
	case "*":
		leftVal, rightVal = arithmetic(leftVal, rightVal, "*")
	case "/":
		if toFloat(rightVal) == 0 {
			result.Error = fmt.Errorf("division by zero")
			return result, result.Error
		}
		leftVal, rightVal = arithmetic(leftVal, rightVal, "/")
	}

	return result, nil
}

func (b *BinaryOpAST) String() string {
	return fmt.Sprintf("(%s %s %s)", b.LeftStr, b.Operator, b.RightStr)
}

// VariableAST represents a variable reference.
type VariableAST struct {
	Name string
}

func (v *VariableAST) Evaluate(vars map[string]interface{}) (EvalResult, error) {
	result := EvalResult{
		Expression: v.Name,
		Variables:  vars,
	}

	if val, ok := vars[v.Name]; ok {
		// Try to convert to bool
		switch v := val.(type) {
		case bool:
			result.Value = v
		case int:
			result.Value = v != 0
		case float64:
			result.Value = v != 0
		case string:
			result.Value = v != ""
		default:
			result.Value = val != nil
		}
	} else {
		result.Error = fmt.Errorf("undefined variable: %s", v.Name)
		return result, result.Error
	}

	return result, nil
}

func (v *VariableAST) String() string {
	return v.Name
}

// LiteralAST represents a literal value (number or string).
type LiteralAST struct {
	Value interface{}
	Text  string // Original text representation
}

func (l *LiteralAST) Evaluate(vars map[string]interface{}) (EvalResult, error) {
	result := EvalResult{
		Expression: l.Text,
		Variables:  vars,
	}

	switch v := l.Value.(type) {
	case bool:
		result.Value = v
	case float64:
		result.Value = v != 0
	case string:
		result.Value = v != ""
	}

	return result, nil
}

func (l *LiteralAST) String() string {
	return l.Text
}

// CompileExpression compiles a string expression into an AST.
func CompileExpression(exprStr string) (ExprAST, error) {
	tokens := tokenizeExpr(exprStr)
	parser := &exprParser{tokens: tokens, pos: 0}
	ast, err := parser.parseExpression()
	if err != nil {
		return nil, err
	}

	// Verify we consumed all tokens (except EOF)
	if parser.pos < len(tokens)-1 {
		return nil, fmt.Errorf("unexpected token after expression: %s", tokens[parser.pos].text)
	}

	return ast, nil
}

// exprParser is a simple recursive descent parser.
type exprParser struct {
	tokens []exprToken
	pos    int
}

type exprToken struct {
	typ  string // "ident", "number", "string", "op", "comp", "logical", "(", ")", "EOF"
	text string
	val  interface{}
}

func tokenizeExpr(expr string) []exprToken {
	var tokens []exprToken
	expr = strings.TrimSpace(expr)

	i := 0
	for i < len(expr) {
		switch {
		case expr[i] == '(':
			tokens = append(tokens, exprToken{typ: "(", text: "("})
			i++
		case expr[i] == ')':
			tokens = append(tokens, exprToken{typ: ")", text: ")"})
			i++
		case expr[i] == '{':
			// Variable in braces
			j := strings.Index(expr[i:], "}")
			if j == -1 {
				i++
				continue
			}
			varName := expr[i+1 : i+j]
			tokens = append(tokens, exprToken{typ: "ident", text: varName})
			i += j + 1
		case expr[i] == '$':
			// Variable with $
			i++
			j := i
			for j < len(expr) && (isAlphaNum(expr[j]) || expr[j] == '_') {
				j++
			}
			tokens = append(tokens, exprToken{typ: "ident", text: expr[i:j]})
			i = j
		case expr[i] == '"':
			// String literal
			j := i + 1
			for j < len(expr) && expr[j] != '"' {
				if expr[j] == '\\' {
					j++
				}
				j++
			}
			if j < len(expr) {
				j++
			}
			tokens = append(tokens, exprToken{typ: "string", text: expr[i:j], val: expr[i+1 : j-1]})
			i = j
		case isDigit(expr[i]):
			// Number
			j := i
			for j < len(expr) && (isDigit(expr[j]) || expr[j] == '.') {
				j++
			}
			numStr := expr[i:j]
			if num, err := strconv.ParseFloat(numStr, 64); err == nil {
				tokens = append(tokens, exprToken{typ: "number", text: numStr, val: num})
			}
			i = j
		case i+1 < len(expr) && expr[i:i+2] == "==":
			tokens = append(tokens, exprToken{typ: "comp", text: "=="})
			i += 2
		case i+1 < len(expr) && expr[i:i+2] == "!=":
			tokens = append(tokens, exprToken{typ: "comp", text: "!="})
			i += 2
		case i+1 < len(expr) && expr[i:i+2] == "<=":
			tokens = append(tokens, exprToken{typ: "comp", text: "<="})
			i += 2
		case i+1 < len(expr) && expr[i:i+2] == ">=":
			tokens = append(tokens, exprToken{typ: "comp", text: ">="})
			i += 2
		case i+1 < len(expr) && expr[i:i+2] == "&&":
			tokens = append(tokens, exprToken{typ: "logical", text: "&&"})
			i += 2
		case i+1 < len(expr) && expr[i:i+2] == "||":
			tokens = append(tokens, exprToken{typ: "logical", text: "||"})
			i += 2
		case expr[i] == '<':
			tokens = append(tokens, exprToken{typ: "comp", text: "<"})
			i++
		case expr[i] == '>':
			tokens = append(tokens, exprToken{typ: "comp", text: ">"})
			i++
		case expr[i] == '+' || expr[i] == '-' || expr[i] == '*' || expr[i] == '/':
			tokens = append(tokens, exprToken{typ: "op", text: string(expr[i])})
			i++
		case isAlpha(expr[i]):
			// Identifier or keyword
			j := i
			for j < len(expr) && isAlphaNum(expr[j]) {
				j++
			}
			word := expr[i:j]
			if word == "true" {
				tokens = append(tokens, exprToken{typ: "bool", text: "true", val: true})
			} else if word == "false" {
				tokens = append(tokens, exprToken{typ: "bool", text: "false", val: false})
			} else {
				tokens = append(tokens, exprToken{typ: "ident", text: word})
			}
			i = j
		case expr[i] == ' ' || expr[i] == '\t' || expr[i] == '\n':
			i++
		default:
			// Unknown character - mark as invalid
			tokens = append(tokens, exprToken{typ: "invalid", text: string(expr[i])})
			i++
		}
	}

	tokens = append(tokens, exprToken{typ: "EOF"})
	return tokens
}

func (p *exprParser) parseExpression() (ExprAST, error) {
	return p.parseLogicalOr()
}

func (p *exprParser) parseLogicalOr() (ExprAST, error) {
	left, err := p.parseLogicalAnd()
	if err != nil {
		return nil, err
	}

	for p.current().typ == "logical" && p.current().text == "||" {
		p.pos++
		right, err := p.parseLogicalAnd()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpAST{
			Left:     left,
			Right:    right,
			Operator: "||",
			LeftStr:  left.String(),
			RightStr: right.String(),
		}
	}

	return left, nil
}

func (p *exprParser) parseLogicalAnd() (ExprAST, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	for p.current().typ == "logical" && p.current().text == "&&" {
		p.pos++
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpAST{
			Left:     left,
			Right:    right,
			Operator: "&&",
			LeftStr:  left.String(),
			RightStr: right.String(),
		}
	}

	return left, nil
}

func (p *exprParser) parseComparison() (ExprAST, error) {
	left, err := p.parseAdditive()
	if err != nil {
		return nil, err
	}

	if p.current().typ == "comp" {
		op := p.current().text
		p.pos++
		right, err := p.parseAdditive()
		if err != nil {
			return nil, err
		}
		return &BinaryOpAST{
			Left:     left,
			Right:    right,
			Operator: op,
			LeftStr:  left.String(),
			RightStr: right.String(),
		}, nil
	}

	return left, nil
}

func (p *exprParser) parseAdditive() (ExprAST, error) {
	left, err := p.parseMultiplicative()
	if err != nil {
		return nil, err
	}

	for p.current().typ == "op" && (p.current().text == "+" || p.current().text == "-") {
		op := p.current().text
		p.pos++
		right, err := p.parseMultiplicative()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpAST{
			Left:     left,
			Right:    right,
			Operator: op,
			LeftStr:  left.String(),
			RightStr: right.String(),
		}
	}

	return left, nil
}

func (p *exprParser) parseMultiplicative() (ExprAST, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for p.current().typ == "op" && (p.current().text == "*" || p.current().text == "/") {
		op := p.current().text
		p.pos++
		right, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpAST{
			Left:     left,
			Right:    right,
			Operator: op,
			LeftStr:  left.String(),
			RightStr: right.String(),
		}
	}

	return left, nil
}

func (p *exprParser) parsePrimary() (ExprAST, error) {
	tok := p.current()

	// Check for invalid tokens first
	if tok.typ == "invalid" {
		return nil, fmt.Errorf("invalid character in expression: %s", tok.text)
	}

	// Check for EOF immediately
	if tok.typ == "EOF" {
		return nil, fmt.Errorf("unexpected end of expression")
	}

	// Handle unary minus
	if tok.typ == "op" && tok.text == "-" {
		p.pos++
		operand, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		// Create a unary negation as (0 - operand)
		return &BinaryOpAST{
			Left:     &LiteralAST{Value: float64(0), Text: "0"},
			Right:    operand,
			Operator: "-",
			LeftStr:  "0",
			RightStr: operand.String(),
		}, nil
	}

	if tok.typ == "(" {
		p.pos++
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if p.current().typ != ")" {
			return nil, fmt.Errorf("expected )")
		}
		p.pos++
		return expr, nil
	}

	if tok.typ == "ident" {
		p.pos++
		return &VariableAST{Name: tok.text}, nil
	}

	if tok.typ == "number" {
		p.pos++
		return &LiteralAST{Value: tok.val, Text: tok.text}, nil
	}

	if tok.typ == "string" {
		p.pos++
		return &LiteralAST{Value: tok.val, Text: tok.text}, nil
	}

	if tok.typ == "bool" {
		p.pos++
		return &LiteralAST{Value: tok.val, Text: tok.text}, nil
	}

	return nil, fmt.Errorf("unexpected token: %v", tok)
}

func (p *exprParser) current() exprToken {
	if p.pos >= len(p.tokens) {
		return exprToken{typ: "EOF"}
	}
	return p.tokens[p.pos]
}

// Helper functions

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlphaNum(c byte) bool {
	return isAlpha(c) || isDigit(c)
}

func evaluateValue(ast ExprAST, vars map[string]interface{}) (interface{}, error) {
	switch a := ast.(type) {
	case *LiteralAST:
		return a.Value, nil
	case *VariableAST:
		if v, ok := vars[a.Name]; ok {
			return v, nil
		}
		return nil, fmt.Errorf("undefined variable: %s", a.Name)
	case *BinaryOpAST:
		if a.Operator == "+" || a.Operator == "-" || a.Operator == "*" || a.Operator == "/" {
			left, _ := evaluateValue(a.Left, vars)
			right, _ := evaluateValue(a.Right, vars)
			leftF := toFloat(left)
			rightF := toFloat(right)

			switch a.Operator {
			case "+":
				return leftF + rightF, nil
			case "-":
				return leftF - rightF, nil
			case "*":
				return leftF * rightF, nil
			case "/":
				if rightF == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				return leftF / rightF, nil
			}
		}
	}
	return nil, fmt.Errorf("cannot evaluate value")
}

func toFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case float32:
		return float64(val)
	}
	return 0
}

func compareGreater(left, right interface{}) bool {
	return toFloat(left) > toFloat(right)
}

func compareLess(left, right interface{}) bool {
	return toFloat(left) < toFloat(right)
}

func compareGreaterEqual(left, right interface{}) bool {
	return toFloat(left) >= toFloat(right)
}

func compareLessEqual(left, right interface{}) bool {
	return toFloat(left) <= toFloat(right)
}

func compareEqual(left, right interface{}) bool {
	return fmt.Sprintf("%v", left) == fmt.Sprintf("%v", right)
}

func compareNotEqual(left, right interface{}) bool {
	return fmt.Sprintf("%v", left) != fmt.Sprintf("%v", right)
}

func arithmetic(left, right interface{}, op string) (interface{}, interface{}) {
	return left, right
}
