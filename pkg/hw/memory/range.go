package memory

import (
	"fmt"
	"iter"
	"log/slog"

	"github.com/Manu343726/cucaracha/pkg/utils/logging"
)

// Contains addressing information of a continuous memory segment
type Range struct {
	// Start address of the memory region
	Start uint32
	// Size of the memory region in bytes
	Size uint32
	// Access flags for the memory region
	Flags Flags
}

// Returns the end address of the memory region.
func (r Range) End() uint32 {
	return r.Start + r.Size
}

// Returns a sequence of addresses within the memory range, stepping by the given amount.
func (r Range) Addresses(step uint32) iter.Seq[uint32] {
	if step == 0 {
		step = 1
	}

	return func(yield func(uint32) bool) {
		for addr := r.Start; addr < r.End(); addr += step {
			if !yield(addr) {
				return
			}
		}
	}
}

func (r Range) String() string {
	return fmt.Sprintf("[0x%08X - 0x%08X, Size: %db, Flags: %s]", r.Start, r.End(), r.Size, r.Flags.String())
}

// Returns a slog attribute representing the memory range in a human-readable format.
func (r Range) AddressRangeLoggingAttribute(name string) slog.Attr {
	return logging.Stringf(name, "0x%08X-0x%08X", r.Start, r.End())
}

// Checks if this memory range overlaps with another.
func (r Range) Overlaps(other Range) bool {
	return r.Start < other.End() && other.Start < r.End()
}

// Checks if any of the provided memory ranges overlap.
func RangesOverlap(ranges []Range) bool {
	for i := 0; i < len(ranges); i++ {
		for j := i + 1; j < len(ranges); j++ {
			if ranges[i].Overlaps(ranges[j]) {
				return true
			}
		}
	}
	return false
}

// Checks if the given address is within the memory range.
func (r Range) ContainsAddress(addr uint32) bool {
	return addr >= r.Start && addr < r.End()
}

// Checks if the given memory range is fully contained within this memory range.
func (r Range) ContainsRange(other Range) bool {
	return other.Start >= r.Start && other.End() <= r.End()
}

// Returns a range within the current range, given an offset and size.
func (r Range) SubRange(offset uint32, size uint32) Range {
	return Range{
		Start: r.Start + offset,
		Size:  size,
		Flags: r.Flags,
	}
}

// Checks whether a second range is right at the end of this range.
func (r Range) IsAdjacentTo(other Range) bool {
	return r.End() == other.Start || other.End() == r.Start
}

// Checks if the provided list of ranges are all contiguous (i.e., each range is adjacent to the next).
func ContiguousRanges(ranges []Range) bool {
	if len(ranges) == 0 {
		return true
	}

	for i := 1; i < len(ranges); i++ {
		if !ranges[i-1].IsAdjacentTo(ranges[i]) {
			return false
		}
	}

	return true
}
