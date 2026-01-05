package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_End(t *testing.T) {
	tests := []struct {
		r        Range
		expected uint32
	}{
		{Range{Start: 0x1000, Size: 0x2000}, 0x3000},
		{Range{Start: 0x0, Size: 0x1000}, 0x1000},
	}

	for _, test := range tests {
		t.Run(test.r.String(), func(t *testing.T) {
			result := test.r.End()
			assert.Equal(t, test.expected, result)
		})
	}
}

func Test_Range_String(t *testing.T) {
	tests := []struct {
		r        Range
		expected string
	}{
		{
			Range{Start: 0x1000, Size: 0x2000, Flags: FlagReadable | FlagWritable},
			"[0x00001000 - 0x00003000, Size: 8192b, Flags: [R,W]]",
		},
		{
			Range{Start: 0x0, Size: 0x1000, Flags: 0},
			"[0x00000000 - 0x00001000, Size: 4096b, Flags: None]",
		},
	}

	for _, test := range tests {
		t.Run(test.r.String(), func(t *testing.T) {
			result := test.r.String()
			assert.Equal(t, test.expected, result)
		})
	}
}

func Test_Overlaps(t *testing.T) {
	tests := []struct {
		r1, r2   Range
		expected bool
	}{
		{
			Range{Start: 0x1000, Size: 0x2000},
			Range{Start: 0x1800, Size: 0x1000},
			true,
		},
		{
			Range{Start: 0x1000, Size: 0x1000},
			Range{Start: 0x2000, Size: 0x1000},
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.r1.String()+" & "+test.r2.String(), func(t *testing.T) {
			result := test.r1.Overlaps(test.r2)
			assert.Equal(t, test.expected, result)
		})
	}
}

func Test_RangesOverlap(t *testing.T) {
	tests := []struct {
		ranges   []Range
		expected bool
	}{
		{
			[]Range{
				{Start: 0x1000, Size: 0x1000},
				{Start: 0x1800, Size: 0x1000},
			},
			true,
		},
		{
			[]Range{
				{Start: 0x1000, Size: 0x1000},
				{Start: 0x2000, Size: 0x1000},
			},
			false,
		},
	}

	for _, test := range tests {
		t.Run("RangesOverlap", func(t *testing.T) {
			result := RangesOverlap(test.ranges)
			assert.Equal(t, test.expected, result)
		})
	}
}

func Test_ContainsAddress(t *testing.T) {
	tests := []struct {
		r        Range
		addr     uint32
		expected bool
	}{
		{
			Range{Start: 0x1000, Size: 0x1000},
			0x1800,
			true,
		},
		{
			Range{Start: 0x1000, Size: 0x1000},
			0x2000,
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.r.String(), func(t *testing.T) {
			result := test.r.ContainsAddress(test.addr)
			assert.Equal(t, test.expected, result)
		})
	}
}

func Test_ContainsRange(t *testing.T) {
	tests := []struct {
		r        Range
		other    Range
		expected bool
	}{
		{
			Range{Start: 0x1000, Size: 0x2000},
			Range{Start: 0x1800, Size: 0x500},
			true,
		},
		{
			Range{Start: 0x1000, Size: 0x1000},
			Range{Start: 0x1800, Size: 0x1000},
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.r.String()+" contains "+test.other.String(), func(t *testing.T) {
			result := test.r.ContainsRange(test.other)
			assert.Equal(t, test.expected, result)
		})
	}
}

func Test_SubRange(t *testing.T) {
	tests := []struct {
		r        Range
		offset   uint32
		size     uint32
		expected Range
	}{
		{
			Range{Start: 0x1000, Size: 0x2000, Flags: FlagReadable},
			0x500,
			0x1000,
			Range{Start: 0x1500, Size: 0x1000, Flags: FlagReadable},
		},
		{
			Range{Start: 0x0, Size: 0x1000, Flags: FlagWritable},
			0x200,
			0x400,
			Range{Start: 0x200, Size: 0x400, Flags: FlagWritable},
		},
	}

	for _, test := range tests {
		t.Run(test.r.String()+" SubRange", func(t *testing.T) {
			result := test.r.SubRange(test.offset, test.size)
			assert.Equal(t, test.expected, result)
		})
	}
}

func Test_IsAdjacentTo(t *testing.T) {
	tests := []struct {
		r1, r2   Range
		expected bool
	}{
		{
			Range{Start: 0x1000, Size: 0x1000},
			Range{Start: 0x2000, Size: 0x1000},
			true,
		},
		{
			Range{Start: 0x1000, Size: 0x1000},
			Range{Start: 0x2500, Size: 0x1000},
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.r1.String()+" & "+test.r2.String(), func(t *testing.T) {
			result := test.r1.IsAdjacentTo(test.r2)
			assert.Equal(t, test.expected, result)
		})
	}
}

func Test_ContiguousRanges(t *testing.T) {
	tests := []struct {
		ranges   []Range
		expected bool
	}{
		{
			[]Range{
				{Start: 0x1000, Size: 0x1000},
				{Start: 0x2000, Size: 0x1000},
				{Start: 0x3000, Size: 0x1000},
			},
			true,
		},
		{
			[]Range{
				{Start: 0x1000, Size: 0x1000},
				{Start: 0x2500, Size: 0x1000},
			},
			false,
		},
	}

	for _, test := range tests {
		t.Run("ContiguousRanges", func(t *testing.T) {
			result := ContiguousRanges(test.ranges)
			assert.Equal(t, test.expected, result)
		})
	}
}
