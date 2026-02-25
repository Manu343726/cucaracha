package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// Formats an uint value into a fixed width binary string of n bits
func FormatUintBinary(value uint64, bits int) string {
	leadingZerosFormat := "%0" + fmt.Sprint(bits) + "s"
	return fmt.Sprintf(leadingZerosFormat, strconv.FormatUint(value, 2))
}

// Formats an uint value into an fixed width hex string of n characters
func FormatUintHex(value uint64, bits int) string {
	leadingZerosFormat := "0x%0" + fmt.Sprint(bits) + "s"
	return fmt.Sprintf(leadingZerosFormat, strconv.FormatUint(value, 16))
}

// Returns an string containing all formatted sequence items separated by a given separator
func FormatSlice[T any](input []T, separator string) string {
	var builder strings.Builder

	for i, value := range input {
		builder.WriteString(fmt.Sprint(value))

		if i < len(input)-1 {
			builder.WriteString(separator)
		}
	}

	return builder.String()
}

// Converts a string to snake_case
func SnakeCase(input string) string {
	var builder strings.Builder

	for i, r := range input {
		if i > 0 && r >= 'A' && r <= 'Z' {
			builder.WriteRune('_')
		}
		builder.WriteRune(r)
	}

	return strings.ToLower(builder.String())
}

// Converts a string to camelCase
func CamelCase(input string) string {
	var builder strings.Builder
	upperNext := false

	for i, r := range input {
		if r == '_' {
			upperNext = true
			continue
		}
		if i == 0 {
			builder.WriteRune(r)
		} else if upperNext {
			builder.WriteRune(r - ('a' - 'A')) // Convert to uppercase
			upperNext = false
		} else {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

// Converts a string to PascalCase
// Handles underscores and hyphens as delimiters
func PascalCase(input string) string {
	var builder strings.Builder
	upperNext := true

	for _, r := range input {
		if r == '_' || r == '-' {
			upperNext = true
			continue
		}
		if upperNext {
			builder.WriteRune(r - ('a' - 'A')) // Convert to uppercase
			upperNext = false
		} else {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

// Converts a string to kebab-case
func KebabCase(input string) string {
	var builder strings.Builder

	for i, r := range input {
		if i > 0 && r >= 'A' && r <= 'Z' {
			builder.WriteRune('-')
		}
		builder.WriteRune(r)
	}

	return strings.ToLower(builder.String())
}
