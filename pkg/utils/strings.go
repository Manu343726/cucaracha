package utils

import (
	"fmt"
	"strconv"
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
