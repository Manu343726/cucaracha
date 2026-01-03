// Package utils provides utility functions for the cucaracha project.
package utils

import (
	"regexp"
	"strings"

	"github.com/fatih/color"
)

// C syntax highlighting colors
var (
	// Keywords
	cKeywordColor = color.New(color.FgMagenta, color.Bold)
	// Types
	cTypeColor = color.New(color.FgCyan)
	// Strings and characters
	cStringColor = color.New(color.FgGreen)
	// Numbers
	cNumberColor = color.New(color.FgYellow)
	// Comments
	cCommentColor = color.New(color.FgHiBlack)
	// Preprocessor directives
	cPreprocessorColor = color.New(color.FgBlue)
	// Operators
	cOperatorColor = color.New(color.FgRed)
	// Function calls
	cFunctionColor = color.New(color.FgHiYellow)
	// Default text
	cDefaultColor = color.New(color.FgWhite)
)

// C language keywords
var cKeywords = map[string]bool{
	"auto": true, "break": true, "case": true, "const": true,
	"continue": true, "default": true, "do": true, "else": true,
	"enum": true, "extern": true, "for": true, "goto": true,
	"if": true, "inline": true, "register": true, "restrict": true,
	"return": true, "sizeof": true, "static": true, "struct": true,
	"switch": true, "typedef": true, "union": true, "volatile": true,
	"while": true, "_Alignas": true, "_Alignof": true, "_Atomic": true,
	"_Bool": true, "_Complex": true, "_Generic": true, "_Imaginary": true,
	"_Noreturn": true, "_Static_assert": true, "_Thread_local": true,
}

// C type keywords
var cTypes = map[string]bool{
	"void": true, "char": true, "short": true, "int": true,
	"long": true, "float": true, "double": true, "signed": true,
	"unsigned": true, "bool": true, "size_t": true, "ssize_t": true,
	"int8_t": true, "int16_t": true, "int32_t": true, "int64_t": true,
	"uint8_t": true, "uint16_t": true, "uint32_t": true, "uint64_t": true,
	"intptr_t": true, "uintptr_t": true, "ptrdiff_t": true,
	"NULL": true, "true": true, "false": true,
}

// Patterns for syntax elements
var (
	// Matches C-style strings (handles escaped quotes)
	cStringPattern = regexp.MustCompile(`"(?:[^"\\]|\\.)*"`)
	// Matches C-style characters
	cCharPattern = regexp.MustCompile(`'(?:[^'\\]|\\.)*'`)
	// Matches single-line comments
	cLineCommentPattern = regexp.MustCompile(`//.*$`)
	// Matches numbers (hex, octal, binary, decimal, float)
	cNumberPattern = regexp.MustCompile(`\b(?:0[xX][0-9a-fA-F]+|0[bB][01]+|0[0-7]+|[0-9]+(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?)[uUlLfF]*\b`)
	// Matches preprocessor directives
	cPreprocessorPattern = regexp.MustCompile(`^\s*#\s*\w+`)
	// Matches identifiers (for keyword/type matching)
	cIdentifierPattern = regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`)
	// Matches function calls (identifier followed by open paren)
	cFunctionCallPattern = regexp.MustCompile(`\b([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`)
	// Matches operators
	cOperatorPattern = regexp.MustCompile(`[+\-*/%&|^!~<>=?:]+|&&|\|\||<<|>>|->|\.`)
)

// token represents a syntax-highlighted token
type token struct {
	text  string
	color *color.Color
	start int
	end   int
}

// HighlightCCode applies syntax highlighting to C source code and returns the colored string
func HighlightCCode(code string) string {
	if code == "" {
		return ""
	}

	// Build a list of tokens with their positions
	var tokens []token

	// First pass: find strings (highest priority - nothing inside strings should be highlighted)
	stringMatches := cStringPattern.FindAllStringIndex(code, -1)
	for _, match := range stringMatches {
		tokens = append(tokens, token{
			text:  code[match[0]:match[1]],
			color: cStringColor,
			start: match[0],
			end:   match[1],
		})
	}

	// Find character literals
	charMatches := cCharPattern.FindAllStringIndex(code, -1)
	for _, match := range charMatches {
		if !overlapsAny(match[0], match[1], tokens) {
			tokens = append(tokens, token{
				text:  code[match[0]:match[1]],
				color: cStringColor,
				start: match[0],
				end:   match[1],
			})
		}
	}

	// Find comments
	commentMatches := cLineCommentPattern.FindAllStringIndex(code, -1)
	for _, match := range commentMatches {
		if !overlapsAny(match[0], match[1], tokens) {
			tokens = append(tokens, token{
				text:  code[match[0]:match[1]],
				color: cCommentColor,
				start: match[0],
				end:   match[1],
			})
		}
	}

	// Find preprocessor directives
	if strings.HasPrefix(strings.TrimSpace(code), "#") {
		preprocMatches := cPreprocessorPattern.FindAllStringIndex(code, -1)
		for _, match := range preprocMatches {
			if !overlapsAny(match[0], match[1], tokens) {
				tokens = append(tokens, token{
					text:  code[match[0]:match[1]],
					color: cPreprocessorColor,
					start: match[0],
					end:   match[1],
				})
			}
		}
	}

	// Find numbers
	numberMatches := cNumberPattern.FindAllStringIndex(code, -1)
	for _, match := range numberMatches {
		if !overlapsAny(match[0], match[1], tokens) {
			tokens = append(tokens, token{
				text:  code[match[0]:match[1]],
				color: cNumberColor,
				start: match[0],
				end:   match[1],
			})
		}
	}

	// Find function calls (before identifiers to prioritize function highlighting)
	funcMatches := cFunctionCallPattern.FindAllStringSubmatchIndex(code, -1)
	for _, match := range funcMatches {
		// match[2]:match[3] is the capture group (function name)
		if len(match) >= 4 && match[2] >= 0 && match[3] >= 0 {
			funcName := code[match[2]:match[3]]
			// Don't highlight keywords/types as functions
			if !cKeywords[funcName] && !cTypes[funcName] {
				if !overlapsAny(match[2], match[3], tokens) {
					tokens = append(tokens, token{
						text:  funcName,
						color: cFunctionColor,
						start: match[2],
						end:   match[3],
					})
				}
			}
		}
	}

	// Find identifiers (keywords and types)
	identMatches := cIdentifierPattern.FindAllStringIndex(code, -1)
	for _, match := range identMatches {
		if !overlapsAny(match[0], match[1], tokens) {
			word := code[match[0]:match[1]]
			var c *color.Color
			if cKeywords[word] {
				c = cKeywordColor
			} else if cTypes[word] {
				c = cTypeColor
			}
			if c != nil {
				tokens = append(tokens, token{
					text:  word,
					color: c,
					start: match[0],
					end:   match[1],
				})
			}
		}
	}

	// Find operators
	opMatches := cOperatorPattern.FindAllStringIndex(code, -1)
	for _, match := range opMatches {
		if !overlapsAny(match[0], match[1], tokens) {
			tokens = append(tokens, token{
				text:  code[match[0]:match[1]],
				color: cOperatorColor,
				start: match[0],
				end:   match[1],
			})
		}
	}

	// Build the final highlighted string
	return buildHighlightedString(code, tokens)
}

// overlapsAny checks if a range overlaps with any existing token
func overlapsAny(start, end int, tokens []token) bool {
	for _, t := range tokens {
		if start < t.end && end > t.start {
			return true
		}
	}
	return false
}

// buildHighlightedString constructs the final string with color codes
func buildHighlightedString(code string, tokens []token) string {
	if len(tokens) == 0 {
		return code
	}

	// Sort tokens by start position
	sortTokens(tokens)

	var result strings.Builder
	pos := 0

	for _, t := range tokens {
		// Add unhighlighted text before this token
		if t.start > pos {
			result.WriteString(code[pos:t.start])
		}
		// Add highlighted token
		result.WriteString(t.color.Sprint(t.text))
		pos = t.end
	}

	// Add remaining unhighlighted text
	if pos < len(code) {
		result.WriteString(code[pos:])
	}

	return result.String()
}

// sortTokens sorts tokens by start position (simple insertion sort for small arrays)
func sortTokens(tokens []token) {
	for i := 1; i < len(tokens); i++ {
		key := tokens[i]
		j := i - 1
		for j >= 0 && tokens[j].start > key.start {
			tokens[j+1] = tokens[j]
			j--
		}
		tokens[j+1] = key
	}
}

// PrintHighlightedCCode prints C code with syntax highlighting to stdout
func PrintHighlightedCCode(code string) {
	print(HighlightCCode(code))
}
