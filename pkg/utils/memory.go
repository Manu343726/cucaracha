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

// Returns the value pointed to by ptr, or a default value if ptr is nil
func DerefOr[T any](ptr *T, defaultValue T) T {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

// Maps a pointer value using the given function, returning a pointer to the result, or nil if the input pointer is nil
func MapPtr[T any, U any](ptr *T, f func(T) U) *U {
	if ptr == nil {
		return nil
	}
	v := f(*ptr)
	return &v
}

// Converts a function taking a pointer to T and returning U into a function taking a value of type T and returning U
func PtrFunc[T any, U any](f func(*T) U) func(T) U {
	return func(v T) U {
		return f(&v)
	}
}
