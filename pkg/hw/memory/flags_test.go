package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlags_Has(t *testing.T) {
	tests := []struct {
		flags    Flags
		check    Flags
		expected bool
	}{
		{FlagReadable, FlagReadable, true},
		{FlagWritable, FlagWritable, true},
		{FlagExecutable, FlagExecutable, true},
		{FlagReadable | FlagWritable, FlagReadable, true},
		{FlagReadable | FlagWritable, FlagWritable, true},
		{FlagReadable | FlagExecutable, FlagExecutable, true},
		{FlagWritable | FlagExecutable, FlagWritable, true},
		{FlagReadable | FlagWritable | FlagExecutable, FlagReadable, true},
		{FlagReadable | FlagWritable | FlagExecutable, FlagWritable, true},
		{FlagReadable | FlagWritable | FlagExecutable, FlagExecutable, true},
		{FlagReadable, FlagWritable, false},
		{FlagWritable, FlagExecutable, false},
		{FlagExecutable, FlagReadable, false},
	}

	for _, test := range tests {
		t.Run(test.flags.String()+"_has_"+test.check.String(), func(t *testing.T) {
			result := test.flags&test.check != 0
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestFlags_IsReadable(t *testing.T) {
	tests := []struct {
		flags    Flags
		expected bool
	}{
		{FlagReadable, true},
		{FlagWritable, false},
		{FlagExecutable, false},
		{FlagReadable | FlagWritable, true},
		{FlagReadable | FlagExecutable, true},
		{FlagWritable | FlagExecutable, false},
		{FlagReadable | FlagWritable | FlagExecutable, true},
		{0, false},
	}

	for _, test := range tests {
		t.Run(test.flags.String()+"_is_readable", func(t *testing.T) {
			result := test.flags.IsReadable()
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestFlags_IsWritable(t *testing.T) {
	tests := []struct {
		flags    Flags
		expected bool
	}{
		{FlagReadable, false},
		{FlagWritable, true},
		{FlagExecutable, false},
		{FlagReadable | FlagWritable, true},
		{FlagReadable | FlagExecutable, false},
		{FlagWritable | FlagExecutable, true},
		{FlagReadable | FlagWritable | FlagExecutable, true},
		{0, false},
	}

	for _, test := range tests {
		t.Run(test.flags.String()+"_is_writable", func(t *testing.T) {
			result := test.flags.IsWritable()
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestFlags_IsExecutable(t *testing.T) {
	tests := []struct {
		flags    Flags
		expected bool
	}{
		{FlagReadable, false},
		{FlagWritable, false},
		{FlagExecutable, true},
		{FlagReadable | FlagWritable, false},
		{FlagReadable | FlagExecutable, true},
		{FlagWritable | FlagExecutable, true},
		{FlagReadable | FlagWritable | FlagExecutable, true},
		{0, false},
	}

	for _, test := range tests {
		t.Run(test.flags.String()+"_is_executable", func(t *testing.T) {
			result := test.flags.IsExecutable()
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestFlags_String(t *testing.T) {
	tests := []struct {
		flags    Flags
		expected string
	}{
		{FlagReadable, "[R]"},
		{FlagWritable, "[W]"},
		{FlagExecutable, "[X]"},
		{FlagReadable | FlagWritable, "[R,W]"},
		{FlagReadable | FlagExecutable, "[R,X]"},
		{FlagWritable | FlagExecutable, "[W,X]"},
		{FlagReadable | FlagWritable | FlagExecutable, "[R,W,X]"},
		{0, "None"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			result := test.flags.String()
			assert.Equal(t, test.expected, result)
		})
	}
}
