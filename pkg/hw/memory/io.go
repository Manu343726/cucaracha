package memory

import "io"

// Implements an io.Reader that reads from a memory slice
type memoryReader struct {
	slice Slice
	pos   uint32
}

// Returns a new io.Reader that reads from the given memory slice
func NewMemoryReader(slice Slice) io.Reader {
	return &memoryReader{
		slice: slice,
		pos:   0,
	}
}

// Reads data from the memory slice into the provided byte slice
func (mr *memoryReader) Read(p []byte) (n int, err error) {
	if mr.pos >= mr.slice.r.Size {
		return 0, io.EOF
	}

	remaining := mr.slice.r.Size - mr.pos
	toRead := uint32(len(p))
	if toRead > remaining {
		toRead = remaining
	}

	subSlice := mr.slice.SubSlice(mr.pos, toRead)
	if err := subSlice.ReadInto(p[:toRead]); err != nil {
		return 0, err
	}

	if toRead < uint32(len(p)) {
		err = io.EOF
	}

	mr.pos += toRead
	return int(toRead), err
}

// Implements an io.Writer that writes to a memory slice
type memoryWriter struct {
	slice Slice
	pos   uint32
}

// Returns a new io.Writer that writes to the given memory slice
func NewMemoryWriter(slice Slice) io.Writer {
	return &memoryWriter{
		slice: slice,
		pos:   0,
	}
}

// Writes data from the provided byte slice into the memory slice
func (mw *memoryWriter) Write(p []byte) (n int, err error) {
	if mw.pos >= mw.slice.r.Size {
		return 0, io.EOF
	}

	remaining := mw.slice.r.Size - mw.pos
	toWrite := uint32(len(p))
	if toWrite > remaining {
		toWrite = remaining
	}

	subSlice := mw.slice.SubSlice(mw.pos, toWrite)
	if err := subSlice.Write(p[:toWrite]); err != nil {
		return 0, err
	}

	if toWrite < uint32(len(p)) {
		err = io.EOF
	}

	mw.pos += toWrite
	return int(toWrite), err
}
