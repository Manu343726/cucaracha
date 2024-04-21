package cpu

import (
	"unsafe"

	"golang.org/x/exp/constraints"
)

type Number interface {
	constraints.Integer | constraints.Float
}

func Sizeof[Type Number]() int {
	return int(unsafe.Sizeof(Type(0)))
}
