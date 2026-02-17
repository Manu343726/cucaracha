package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu"
	"github.com/Manu343726/cucaracha/pkg/hw/memory"
	"github.com/Manu343726/cucaracha/pkg/runtime"
	"github.com/Manu343726/cucaracha/pkg/runtime/program"
)

// Token types for expression parsing
type TokenType int

const (
	TokenNumber TokenType = iota
	TokenRegister
	TokenSymbol
	TokenPlus
	TokenMinus
	TokenMul
	TokenDiv
	TokenMod
	TokenAnd
	TokenOr
	TokenXor
	TokenShiftLeft
	TokenShiftRight
	TokenLBracket
	TokenRBracket
	TokenLParen
	TokenRParen
)

// Token represents a lexical token in an expression
type Token struct {
	Type  TokenType
	Value string
	Num   uint32 // For number tokens
}

// ExpressionEvaluator evaluates expressions in the context of a debugger backend
type ExpressionEvaluator struct {
	runtime     runtime.Runtime
	programFile program.ProgramFile
}

// NewExpressionEvaluator creates a new expression evaluator
func NewExpressionEvaluator(runtime runtime.Runtime, programFile program.ProgramFile) *ExpressionEvaluator {
	return &ExpressionEvaluator{runtime: runtime, programFile: programFile}
}

// Eval evaluates an expression string and returns the result
func (e *ExpressionEvaluator) Eval(expr string) (uint32, error) {
	tokens, err := e.Tokenize(expr)
	if err != nil {
		return 0, err
	}

	if len(tokens) == 0 {
		return 0, fmt.Errorf("empty expression")
	}

	result, remaining, err := e.parseAddSub(tokens)
	if err != nil {
		return 0, err
	}

	if len(remaining) > 0 {
		return 0, fmt.Errorf("unexpected token: %s", remaining[0].Value)
	}

	return result, nil
}

// Tokenize breaks an expression into tokens
func (e *ExpressionEvaluator) Tokenize(expr string) ([]Token, error) {
	var tokens []Token
	expr = strings.TrimSpace(expr)

	for len(expr) > 0 {
		expr = strings.TrimSpace(expr)
		if len(expr) == 0 {
			break
		}

		// Check single-character operators first
		switch expr[0] {
		case '+':
			tokens = append(tokens, Token{Type: TokenPlus, Value: "+"})
			expr = expr[1:]
			continue
		case '-':
			tokens = append(tokens, Token{Type: TokenMinus, Value: "-"})
			expr = expr[1:]
			continue
		case '*':
			tokens = append(tokens, Token{Type: TokenMul, Value: "*"})
			expr = expr[1:]
			continue
		case '/':
			tokens = append(tokens, Token{Type: TokenDiv, Value: "/"})
			expr = expr[1:]
			continue
		case '%':
			tokens = append(tokens, Token{Type: TokenMod, Value: "%"})
			expr = expr[1:]
			continue
		case '&':
			tokens = append(tokens, Token{Type: TokenAnd, Value: "&"})
			expr = expr[1:]
			continue
		case '|':
			tokens = append(tokens, Token{Type: TokenOr, Value: "|"})
			expr = expr[1:]
			continue
		case '^':
			tokens = append(tokens, Token{Type: TokenXor, Value: "^"})
			expr = expr[1:]
			continue
		case '[':
			tokens = append(tokens, Token{Type: TokenLBracket, Value: "["})
			expr = expr[1:]
			continue
		case ']':
			tokens = append(tokens, Token{Type: TokenRBracket, Value: "]"})
			expr = expr[1:]
			continue
		case '(':
			tokens = append(tokens, Token{Type: TokenLParen, Value: "("})
			expr = expr[1:]
			continue
		case ')':
			tokens = append(tokens, Token{Type: TokenRParen, Value: ")"})
			expr = expr[1:]
			continue
		case '<':
			if len(expr) >= 2 && expr[1] == '<' {
				tokens = append(tokens, Token{Type: TokenShiftLeft, Value: "<<"})
				expr = expr[2:]
				continue
			}
		case '>':
			if len(expr) >= 2 && expr[1] == '>' {
				tokens = append(tokens, Token{Type: TokenShiftRight, Value: ">>"})
				expr = expr[2:]
				continue
			}
		}

		// Check for hex number (0x...)
		if len(expr) >= 2 && expr[0] == '0' && (expr[1] == 'x' || expr[1] == 'X') {
			end := 2
			for end < len(expr) && IsHexDigit(expr[end]) {
				end++
			}
			numStr := expr[2:end]
			num, err := strconv.ParseUint(numStr, 16, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid hex number: 0x%s", numStr)
			}
			tokens = append(tokens, Token{Type: TokenNumber, Value: expr[:end], Num: uint32(num)})
			expr = expr[end:]
			continue
		}

		// Check for binary number (0b...)
		if len(expr) >= 2 && expr[0] == '0' && (expr[1] == 'b' || expr[1] == 'B') {
			end := 2
			for end < len(expr) && (expr[end] == '0' || expr[end] == '1' || expr[end] == '_') {
				end++
			}
			numStr := strings.ReplaceAll(expr[2:end], "_", "")
			num, err := strconv.ParseUint(numStr, 2, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid binary number: 0b%s", numStr)
			}
			tokens = append(tokens, Token{Type: TokenNumber, Value: expr[:end], Num: uint32(num)})
			expr = expr[end:]
			continue
		}

		// Check for decimal number
		if IsDigit(expr[0]) {
			end := 0
			for end < len(expr) && IsDigit(expr[end]) {
				end++
			}
			numStr := expr[:end]
			num, err := strconv.ParseUint(numStr, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid number: %s", numStr)
			}
			tokens = append(tokens, Token{Type: TokenNumber, Value: numStr, Num: uint32(num)})
			expr = expr[end:]
			continue
		}

		// Check for register or symbol (identifier)
		if IsAlpha(expr[0]) || expr[0] == '_' {
			end := 0
			for end < len(expr) && (IsAlphaNum(expr[end]) || expr[end] == '_') {
				end++
			}
			name := expr[:end]
			nameLower := strings.ToLower(name)

			// Check if it's a register
			if IsRegisterName(nameLower) {
				tokens = append(tokens, Token{Type: TokenRegister, Value: nameLower})
			} else {
				tokens = append(tokens, Token{Type: TokenSymbol, Value: name})
			}
			expr = expr[end:]
			continue
		}

		return nil, fmt.Errorf("unexpected character: %c", expr[0])
	}

	return tokens, nil
}

// Character classification helpers
func IsDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func IsHexDigit(c byte) bool {
	return IsDigit(c) || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

func IsAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func IsAlphaNum(c byte) bool {
	return IsAlpha(c) || IsDigit(c)
}

func IsRegisterName(name string) bool {
	switch name {
	case "r0", "r1", "r2", "r3", "r4", "r5", "r6", "r7", "r8", "r9":
		return true
	case "sp", "lr", "pc", "cpsr":
		return true
	}
	return false
}

// Recursive descent parser with operator precedence
// Precedence (lowest to highest):
// 1. | (OR)
// 2. ^ (XOR)
// 3. & (AND)
// 4. << >> (shifts)
// 5. + - (add/sub)
// 6. * / % (mul/div/mod)
// 7. unary -, []

func (e *ExpressionEvaluator) parseAddSub(tokens []Token) (uint32, []Token, error) {
	left, tokens, err := e.parseMulDiv(tokens)
	if err != nil {
		return 0, nil, err
	}

	for len(tokens) > 0 {
		if tokens[0].Type == TokenPlus {
			tokens = tokens[1:]
			right, remaining, err := e.parseMulDiv(tokens)
			if err != nil {
				return 0, nil, err
			}
			left = left + right
			tokens = remaining
		} else if tokens[0].Type == TokenMinus {
			tokens = tokens[1:]
			right, remaining, err := e.parseMulDiv(tokens)
			if err != nil {
				return 0, nil, err
			}
			left = left - right
			tokens = remaining
		} else {
			break
		}
	}

	return left, tokens, nil
}

func (e *ExpressionEvaluator) parseMulDiv(tokens []Token) (uint32, []Token, error) {
	left, tokens, err := e.parseBitwise(tokens)
	if err != nil {
		return 0, nil, err
	}

	for len(tokens) > 0 {
		switch tokens[0].Type {
		case TokenMul:
			tokens = tokens[1:]
			right, remaining, err := e.parseBitwise(tokens)
			if err != nil {
				return 0, nil, err
			}
			left = left * right
			tokens = remaining
		case TokenDiv:
			tokens = tokens[1:]
			right, remaining, err := e.parseBitwise(tokens)
			if err != nil {
				return 0, nil, err
			}
			if right == 0 {
				return 0, nil, fmt.Errorf("division by zero")
			}
			left = left / right
			tokens = remaining
		case TokenMod:
			tokens = tokens[1:]
			right, remaining, err := e.parseBitwise(tokens)
			if err != nil {
				return 0, nil, err
			}
			if right == 0 {
				return 0, nil, fmt.Errorf("modulo by zero")
			}
			left = left % right
			tokens = remaining
		default:
			return left, tokens, nil
		}
	}

	return left, tokens, nil
}

func (e *ExpressionEvaluator) parseBitwise(tokens []Token) (uint32, []Token, error) {
	left, tokens, err := e.parseShift(tokens)
	if err != nil {
		return 0, nil, err
	}

	for len(tokens) > 0 {
		switch tokens[0].Type {
		case TokenAnd:
			tokens = tokens[1:]
			right, remaining, err := e.parseShift(tokens)
			if err != nil {
				return 0, nil, err
			}
			left = left & right
			tokens = remaining
		case TokenOr:
			tokens = tokens[1:]
			right, remaining, err := e.parseShift(tokens)
			if err != nil {
				return 0, nil, err
			}
			left = left | right
			tokens = remaining
		case TokenXor:
			tokens = tokens[1:]
			right, remaining, err := e.parseShift(tokens)
			if err != nil {
				return 0, nil, err
			}
			left = left ^ right
			tokens = remaining
		default:
			return left, tokens, nil
		}
	}

	return left, tokens, nil
}

func (e *ExpressionEvaluator) parseShift(tokens []Token) (uint32, []Token, error) {
	left, tokens, err := e.parseUnary(tokens)
	if err != nil {
		return 0, nil, err
	}

	for len(tokens) > 0 {
		switch tokens[0].Type {
		case TokenShiftLeft:
			tokens = tokens[1:]
			right, remaining, err := e.parseUnary(tokens)
			if err != nil {
				return 0, nil, err
			}
			left = left << right
			tokens = remaining
		case TokenShiftRight:
			tokens = tokens[1:]
			right, remaining, err := e.parseUnary(tokens)
			if err != nil {
				return 0, nil, err
			}
			left = left >> right
			tokens = remaining
		default:
			return left, tokens, nil
		}
	}

	return left, tokens, nil
}

func (e *ExpressionEvaluator) parseUnary(tokens []Token) (uint32, []Token, error) {
	if len(tokens) == 0 {
		return 0, nil, fmt.Errorf("unexpected end of expression")
	}

	// Unary minus
	if tokens[0].Type == TokenMinus {
		tokens = tokens[1:]
		val, remaining, err := e.parseUnary(tokens)
		if err != nil {
			return 0, nil, err
		}
		return uint32(-int32(val)), remaining, nil
	}

	// Memory dereference [expr]
	if tokens[0].Type == TokenLBracket {
		tokens = tokens[1:]
		addr, remaining, err := e.parseAddSub(tokens)
		if err != nil {
			return 0, nil, err
		}
		if len(remaining) == 0 || remaining[0].Type != TokenRBracket {
			return 0, nil, fmt.Errorf("expected ']' after memory address")
		}
		remaining = remaining[1:]

		// Read 4 bytes from memory
		val, err := memory.ReadUint32(e.runtime.Memory(), addr)
		if err != nil {
			return 0, nil, err
		}
		return val, remaining, nil
	}

	return e.parsePrimary(tokens)
}

func (e *ExpressionEvaluator) parsePrimary(tokens []Token) (uint32, []Token, error) {
	if len(tokens) == 0 {
		return 0, nil, fmt.Errorf("unexpected end of expression")
	}

	tok := tokens[0]
	tokens = tokens[1:]

	switch tok.Type {
	case TokenNumber:
		return tok.Num, tokens, nil

	case TokenRegister:
		val, err := cpu.ReadRegisterByName(e.runtime.CPU().Registers(), tok.Value)
		if err != nil {
			return 0, nil, err
		}
		return val, tokens, nil

	case TokenSymbol:
		val, err := e.resolveSymbol(tok.Value)
		if err != nil {
			return 0, nil, err
		}
		return val, tokens, nil

	case TokenLParen:
		val, remaining, err := e.parseAddSub(tokens)
		if err != nil {
			return 0, nil, err
		}
		if len(remaining) == 0 || remaining[0].Type != TokenRParen {
			return 0, nil, fmt.Errorf("expected ')' after expression")
		}
		return val, remaining[1:], nil

	default:
		return 0, nil, fmt.Errorf("unexpected token: %s", tok.Value)
	}
}

// resolveSymbol resolves a symbol name to its address/value
func (e *ExpressionEvaluator) resolveSymbol(name string) (uint32, error) {
	return runtime.ResolveSymbol(e.runtime, e.programFile, name)
}

// readVariableValue reads the current value of a source-level variable
func (e *ExpressionEvaluator) readVariableValue(v *program.VariableInfo) (uint32, error) {
	return runtime.ReadVariable(e.runtime, v)
}

// FormatBinary formats a 32-bit value as a binary string with underscore separators
func FormatBinary(val uint32) string {
	s := fmt.Sprintf("%032b", val)
	return s[0:8] + "_" + s[8:16] + "_" + s[16:24] + "_" + s[24:32]
}

// Evaluates an expression and returns its value
func Eval(r runtime.Runtime, pf program.ProgramFile, expr string) (uint32, error) {
	evaluator := NewExpressionEvaluator(r, pf)
	return evaluator.Eval(expr)
}
