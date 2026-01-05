package memory

// Returns the next address aligned to the given alignment.
func NextAlignedAddress(addr uint32, alignment uint32) uint32 {
	if alignment == 0 {
		return addr
	}
	remainder := addr % alignment
	if remainder == 0 {
		return addr
	}
	return addr + (alignment - remainder)
}

// Checks if the given address is aligned to the given boundary.
func IsAligned(addr uint32, alignment uint32) bool {
	if alignment == 0 {
		return true
	}
	return addr%alignment == 0
}

// Returns the size aligned up to the given boundary.
func AlignSize(size uint32, alignment uint32) uint32 {
	if alignment == 0 {
		return size
	}
	remainder := size % alignment
	if remainder == 0 {
		return size
	}
	return size + (alignment - remainder)
}
