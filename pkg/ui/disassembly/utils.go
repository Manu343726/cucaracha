package disassembly

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseAddress parses an address string (hex or decimal)
func ParseAddress(addrStr string) (uint32, error) {
	addrStr = strings.TrimSpace(addrStr)

	// Try hex format
	if strings.HasPrefix(addrStr, "0x") || strings.HasPrefix(addrStr, "0X") {
		val, err := strconv.ParseUint(addrStr[2:], 16, 32)
		return uint32(val), err
	}

	// Try decimal format
	val, err := strconv.ParseUint(addrStr, 10, 32)
	return uint32(val), err
}

// FormatAddress formats an address for display
func FormatAddress(addr uint32) string {
	return fmt.Sprintf("0x%08x", addr)
}

// FormatRange formats an address range for display
func FormatRange(start, end uint32) string {
	return fmt.Sprintf("0x%x-0x%x", start, end)
}

// TruncateString truncates a string to a maximum length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// PadRight pads a string to the right with spaces
func PadRight(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(" ", length-len(s))
}

// PadLeft pads a string to the left with spaces
func PadLeft(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return strings.Repeat(" ", length-len(s)) + s
}
