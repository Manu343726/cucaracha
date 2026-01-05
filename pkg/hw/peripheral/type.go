package peripheral

// Identified the type of a peripheral
type Type string

func (t Type) String() string {
	return string(t)
}

// Returns the byte-encoded representation of the peripheral type.
func (t Type) Encode() []byte {
	b := make([]byte, 16)
	copy(b, t)
	return b
}
