package types

import (
	"fmt"
	"reflect"

	"github.com/Manu343726/cucaracha/pkg/utils"
)

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

func (vt ValueType) IsInteger() bool {
	switch vt {
	case ValueType_Int32:
		return true
	}

	return false
}

func (vt ValueType) IsSigned() bool {
	switch vt {
	case ValueType_Int32:
		return true
	}

	return false
}

func (vt ValueType) Bits() int {
	switch vt {
	case ValueType_Int32:
		return 32
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

type Value struct {
	value     interface{}
	valueType ValueType
}

func (v *Value) Type() ValueType {
	return v.valueType
}

func (v *Value) Int32() int32 {
	return v.value.(int32)
}

func (v *Value) String() string {
	return fmt.Sprint(v.value)
}

func (v *Value) Encode() uint64 {
	switch v.valueType {
	case ValueType_Int32:
		return utils.BitCast[uint64](v.Int32())
	}

	panic("unreachable")
}

// Stores a 32 bits signed integer value
func Int32(value int32) Value {
	return Value{
		value:     value,
		valueType: ValueType_Int32,
	}
}
