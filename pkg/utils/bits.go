package utils

import (
	"unsafe"

	"golang.org/x/exp/constraints"
)

const BitsPerByte = 8

// Returns the size in bits of n bytes
func Bits(bytes int) int {
	return bytes * 8
}

// Returns the size in bytes of values of a type
func Sizeof[T any]() int {
	var val T
	return int(unsafe.Sizeof(val))
}

// Returns the size in bits of values of a type
func SizeofBits[T any]() int {
	return Bits(Sizeof[T]())
}

// Returns an all ones bitmask of n bits of the given unsigned integer type
func AllOnes[T constraints.Unsigned](bits int) T {
	return (T(1) << bits) - T(1)
}

// Implements a read/write view over an unsigned interger, allowing manipullating individual bits easily
type BitView[T constraints.Unsigned] struct {
	Bits *T
}

// Returns the viewed unsigned int value
func (v BitView[T]) Value() T {
	return *v.Bits
}

// Returns the size in bits of the viewed value
func (v BitView[T]) SizeofBits() int {
	return SizeofBits[T]()
}

// Extracts a range of bits given a first bit and a width
func (v BitView[T]) Read(bit int, width int) T {
	mask := AllOnes[T](width)
	return (v.Value() >> bit) & mask
}

// Copies a value into a range of bits, given the start and width of the range.
// All most significant bits of the value not fitting into the destination range are ignored.
func (v BitView[T]) Write(value T, bit int, width int) {
	clearedValue := value & AllOnes[T](width)
	*v.Bits = (*v.Bits) | (clearedValue << bit)
}

// Sets all bits in a range to 1
func (v BitView[T]) SetBits(bit int, width int) {
	v.Write(AllOnes[T](width), bit, width)
}

// Sets all bits in a range to 0
func (v BitView[T]) ClearBits(bit int, width int) {
	v.Write(T(0), bit, width)
}

// Sets bit to 1
func (v BitView[T]) SetBit(bit int) {
	v.SetBits(bit, 1)
}

// Sets bit to 0
func (v BitView[T]) ClearBit(bit int) {
	v.ClearBits(bit, 1)
}

// Creates a bit view out of an unsigned int
func CreateBitView[T constraints.Unsigned](value *T) BitView[T] {
	return BitView[T]{
		Bits: value,
	}
}
