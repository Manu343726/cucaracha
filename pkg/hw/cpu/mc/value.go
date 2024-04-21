package mc

import "reflect"

// Represents the type of a machine instruction operand value
type ValueType uint

const (
	ValueType_Int32 ValueType = iota
)

func (vt ValueType) String() string {
	switch vt {
	case ValueType_Int32:
		return "Int32"
	}

	panic("unreachable")
}

// Returns the golang equivalent of the value type
func (vt ValueType) GoType() reflect.Type {
	switch vt {
	case ValueType_Int32:
		return reflect.TypeFor[int32]()
	}

	panic("unreachable")
}
