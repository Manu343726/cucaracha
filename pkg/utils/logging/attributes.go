package logging

import (
	"fmt"
	"log/slog"
	"strings"

	"golang.org/x/exp/constraints"
)

// Formats an unsigned integer as a hexadecimal string for logging.
func Hex[T constraints.Unsigned](name string, value T) slog.Attr {
	return slog.String(name, fmt.Sprintf("0x%X", value))
}

// Formats an unsigned integer as a binary string for logging.
func Binary[T constraints.Unsigned](name string, value T) slog.Attr {
	return slog.String(name, fmt.Sprintf("0b%b", value))
}

// Formats a byte array as a hexadecimal string for logging.
func HexBytes(name string, value []byte) slog.Attr {
	return slog.String(name, fmt.Sprintf("0x%X", value))
}

// Formats a cucaracha memory address for logging.
func Address(name string, value uint32) slog.Attr {
	return Hex(name, value)
}

// Formats multiple cucaracha memory addresses as a comma-separated list for logging.
func Addresses(name string, addresses []uint32) slog.Attr {
	var builder strings.Builder
	builder.WriteString("[")

	first := true
	for _, addr := range addresses {
		if first {
			first = false
		} else {
			builder.WriteString(", ")
		}
		builder.WriteString(fmt.Sprintf("0x%X", addr))
	}

	builder.WriteString("]")
	return slog.String(name, builder.String())
}

// Formats a binary encoded instruction for logging.
func EncodedInstruction(name string, value uint32) slog.Attr {
	return Hex(name, value)
}

// Formats a string logging attribute using printf-style formatting.
func Stringf(name string, format string, args ...any) slog.Attr {
	return slog.String(name, fmt.Sprintf(format, args...))
}

// Formats a cucaracha instruction for logging.
func Instruction(name string, assembly string) slog.Attr {
	return Stringf(name, "{%v}", assembly)
}
