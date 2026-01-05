package memory

import "strings"

// Flags represent memory region properties.
type Flags uint32

const (
	FlagReadable   Flags = 1 << 0
	FlagWritable   Flags = 1 << 1
	FlagExecutable Flags = 1 << 2
)

func (f Flags) IsReadable() bool {
	return f&FlagReadable != 0
}

func (f Flags) IsWritable() bool {
	return f&FlagWritable != 0
}

func (f Flags) IsExecutable() bool {
	return f&FlagExecutable != 0
}

func (f Flags) String() string {
	var flags []string
	if f.IsReadable() {
		flags = append(flags, "R")
	}
	if f.IsWritable() {
		flags = append(flags, "W")
	}
	if f.IsExecutable() {
		flags = append(flags, "X")
	}
	if len(flags) == 0 {
		return "None"
	}

	return "[" + strings.Join(flags, ",") + "]"
}
