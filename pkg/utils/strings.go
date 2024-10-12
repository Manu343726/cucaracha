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
