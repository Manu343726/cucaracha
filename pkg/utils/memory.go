package utils

// Returns the next address greater than or equal to addr that is aligned to align bytes.
func NextAligned(addr, align uint32) uint32 {
	if align == 0 {
		return addr
	}
	if addr%align == 0 {
		return addr
	}
	return ((addr / align) + 1) * align
}

// Returns a pointer to the given value
func Ptr[T any](v T) *T {
	return &v
}
