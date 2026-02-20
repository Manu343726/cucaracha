package memory

import "fmt"

type Slice struct {
	m Memory
	r Range
}

// Returns an slice of the specified range of memory
// Remarks: Note that slices do not perform bounds checking until
// an actual read or write is attempted.
func NewSlice(m Memory, r *Range) Slice {
	return Slice{m: m, r: *r}
}

// Returns a slice covering the entire memory
func MemorySlice(m Memory) Slice {
	return Slice{
		m: m,
		r: Range{Start: 0, Size: uint32(m.Size())},
	}
}

// Returns the range of memory covered by the slice
func (s Slice) Range() Range {
	return s.r
}

// Returns a slice from the given offset and size within the memory slice
// Remarks: Note that slices do not perform bounds checking until
// an actual read or write is attempted.
func (s Slice) SubSlice(offset uint32, size uint32) Slice {
	return Slice{
		m: s.m,
		r: Range{
			Start: s.r.Start + offset,
			Size:  size,
		},
	}
}

// Returns a slice from the given start to end addresses within the memory slice
// Remarks: Note that slices do not perform bounds checking until
// an actual read or write is attempted.
func (s Slice) FromTo(start uint32, end uint32) Slice {
	size := end - start
	return s.SubSlice(start, size)
}

// Reads all bytes from the memory slice
func (s Slice) ReadAll() ([]byte, error) {
	data := make([]byte, s.r.Size)
	if err := s.ReadInto(data); err != nil {
		return nil, err
	}

	return data, nil
}

// Reads all bytes from the memory slice into the provided buffer
func (s Slice) ReadInto(buffer []byte) error {
	if s.r.End() > uint32(s.m.Size()) {
		return fmt.Errorf("slice read out of bounds: slice end 0x%X exceeds memory size 0x%X", s.r.End(), s.m.Size())
	}

	if uint32(len(buffer)) < s.r.Size {
		return fmt.Errorf("buffer too small for slice read: buffer size %d bytes, slice size %d bytes", len(buffer), s.r.Size)
	}

	data, err := s.m.Read(s.r.Start, int(s.r.Size))
	if err != nil {
		return fmt.Errorf("slice read failed: %w", err)
	}
	copy(buffer, data)
	return nil
}

// Writes all bytes from the provided buffer into the memory slice
func (s Slice) Write(buffer []byte) error {
	if s.r.End() > uint32(s.m.Size()) {
		return fmt.Errorf("slice write out of bounds: slice end 0x%X exceeds memory size 0x%X", s.r.End(), s.m.Size())
	}

	if uint32(len(buffer)) < s.r.Size {
		return fmt.Errorf("buffer too small for slice write: buffer size %d bytes, slice size %d bytes", len(buffer), s.r.Size)
	}

	if err := s.m.Write(s.r.Start, buffer[:s.r.Size]); err != nil {
		return fmt.Errorf("slice write failed: %w", err)
	}

	return nil
}

// Writes an unsigned 32 bit integer to the start of the memory slice in little-endian format.
func (s Slice) WriteUint32(value uint32) error {
	if s.r.Size < 4 {
		return fmt.Errorf("slice too small for uint32 write: slice size %d bytes", s.r.Size)
	}

	return WriteUint32(s.m, s.r.Start, value)
}

// Reads an unsigned 32 bit integer from the start of the memory slice in little-endian format.
func (s Slice) ReadUint32() (uint32, error) {
	if s.r.Size < 4 {
		return 0, fmt.Errorf("slice too small for uint32 read: slice size %d bytes", s.r.Size)
	}

	return ReadUint32(s.m, s.r.Start)
}

// Reads the memory slice and returns its value as an unsigned 32 bit integer in little-endian format.
// In contrast to ReadUint32, this function can read slices smaller than 4 bytes, returning the value
// with zero-extension.
func (s Slice) ReadAsUint32() (uint32, error) {
	if s.r.Size > 4 {
		return 0, fmt.Errorf("slice too large for uint32 read: slice size %d bytes", s.r.Size)
	}

	data, err := s.m.Read(s.r.Start, int(s.r.Size))
	if err != nil {
		return 0, fmt.Errorf("slice uint32 read failed: %w", err)
	}

	var value uint32
	for i, b := range data {
		value |= uint32(b) << (8 * i)
	}

	return value, nil
}
