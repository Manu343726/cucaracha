package cpu

import (
	"errors"
	"math"

	"golang.org/x/exp/constraints"
)

func Zero[Type Number]() Type {
	return 0
}

func One[Type Number]() Type {
	return 1
}

func True[Type Number]() Type {
	return One[Type]()
}

func False[Type Number]() Type {
	return Zero[Type]()
}

type RingArithmeticUnit[Register RegisterName, Type Number] interface {
	Add(lhs Register, rhs Register, dest Register) error
	Sub(lhs Register, rhs Register, dest Register) error
	Mul(lhs Register, rhs Register, dest Register) error
	Div(lhs Register, rhs Register, dest Register) error
}

type BitOpsUnit[Register RegisterName, Type constraints.Integer] interface {
	LeftShift(lhs Register, rhs Register, dest Register) error
	RightShift(lhs Register, rhs Register, dest Register) error
	And(lhs Register, rhs Register, dest Register) error
	Or(lhs Register, rhs Register, dest Register) error
	Xor(lhs Register, rhs Register, dest Register) error
	Not(src Register, dest Register) error
}

type ComparisonUnit[Register RegisterName, Type Number] interface {
	Equal(lhs Register, rhs Register, dest Register) error
	LessThan(lhs Register, rhs Register, dest Register) error
}

type IntegerAlu[Register RegisterName, Type constraints.Integer] interface {
	RingArithmeticUnit[Register, Type]
	BitOpsUnit[Register, Type]
	ComparisonUnit[Register, Type]
}

type IntergerAluFactory[Register RegisterName, Type constraints.Integer] func(rs RegisterBank[Register, Type]) IntegerAlu[Register, Type]

type FloatAlu[Register RegisterName, Type constraints.Float] interface {
	RingArithmeticUnit[Register, Type]
	ComparisonUnit[Register, Type]

	Sin(src Register, dest Register) error
	Cos(src Register, dest Register) error
	Tan(src Register, dest Register) error
	Asin(src Register, dest Register) error
	Acos(src Register, dest Register) error
	Atan(src Register, dest Register) error
}

type FloatAluFactory[Register RegisterName, Type constraints.Float] func(rs RegisterBank[Register, Type]) FloatAlu[Register, Type]

type arithmeticUnit[Register RegisterName, Type Number] struct {
	rs RegisterBank[Register, Type]
}

func (u *arithmeticUnit[Register, Type]) UnaryOp(src Register, dest Register, opBody func(Type) Type) error {
	srcValue, err := u.rs.Read(src)

	if err != nil {
		return err
	}

	return u.rs.Write(opBody(srcValue), dest)
}

func (u *arithmeticUnit[Register, Type]) BinaryOp(lhs Register, rhs Register, dest Register, opBody func(Type, Type) Type) error {
	lhsValue, lhsErr := u.rs.Read(lhs)
	rhsValue, rhsErr := u.rs.Read(rhs)

	if lhsErr != nil || rhsErr != nil {
		return errors.Join(lhsErr, rhsErr)
	}

	return u.rs.Write(opBody(lhsValue, rhsValue), dest)
}

type ringArithmeticUnit[Register RegisterName, Type Number] struct {
	au arithmeticUnit[Register, Type]
}

func MakeRingArithmeticUnit[Register RegisterName, Type Number](rs RegisterBank[Register, Type]) RingArithmeticUnit[Register, Type] {
	return &ringArithmeticUnit[Register, Type]{au: arithmeticUnit[Register, Type]{rs: rs}}
}

func (u *ringArithmeticUnit[Register, Type]) Add(lhs Register, rhs Register, dest Register) error {
	return u.au.BinaryOp(lhs, rhs, dest, func(lhs Type, rhs Type) Type {
		return lhs + rhs
	})
}

func (u *ringArithmeticUnit[Register, Type]) Sub(lhs Register, rhs Register, dest Register) error {
	return u.au.BinaryOp(lhs, rhs, dest, func(lhs Type, rhs Type) Type {
		return lhs - rhs
	})
}

func (u *ringArithmeticUnit[Register, Type]) Mul(lhs Register, rhs Register, dest Register) error {
	return u.au.BinaryOp(lhs, rhs, dest, func(lhs Type, rhs Type) Type {
		return lhs * rhs
	})
}

func (u *ringArithmeticUnit[Register, Type]) Div(lhs Register, rhs Register, dest Register) error {
	return u.au.BinaryOp(lhs, rhs, dest, func(lhs Type, rhs Type) Type {
		return lhs / rhs
	})
}

type bitOpsUnit[Register RegisterName, Type constraints.Integer] struct {
	au arithmeticUnit[Register, Type]
}

func MakeBitOptsUnit[Register RegisterName, Type constraints.Integer](rs RegisterBank[Register, Type]) BitOpsUnit[Register, Type] {
	return &bitOpsUnit[Register, Type]{au: arithmeticUnit[Register, Type]{rs: rs}}
}

func (u *bitOpsUnit[Register, Type]) LeftShift(lhs Register, rhs Register, dest Register) error {
	return u.au.BinaryOp(lhs, rhs, dest, func(lhs Type, rhs Type) Type {
		return lhs << rhs
	})
}

func (u *bitOpsUnit[Register, Type]) RightShift(lhs Register, rhs Register, dest Register) error {
	return u.au.BinaryOp(lhs, rhs, dest, func(lhs Type, rhs Type) Type {
		return lhs >> rhs
	})
}

func (u *bitOpsUnit[Register, Type]) And(lhs Register, rhs Register, dest Register) error {
	return u.au.BinaryOp(lhs, rhs, dest, func(lhs Type, rhs Type) Type {
		return lhs & rhs
	})
}

func (u *bitOpsUnit[Register, Type]) Or(lhs Register, rhs Register, dest Register) error {
	return u.au.BinaryOp(lhs, rhs, dest, func(lhs Type, rhs Type) Type {
		return lhs | rhs
	})
}

func (u *bitOpsUnit[Register, Type]) Xor(lhs Register, rhs Register, dest Register) error {
	return u.au.BinaryOp(lhs, rhs, dest, func(lhs Type, rhs Type) Type {
		return lhs ^ rhs
	})
}

func (u *bitOpsUnit[Register, Type]) Not(src Register, dest Register) error {
	return u.au.UnaryOp(src, dest, func(src Type) Type {
		return ^src
	})
}

type comparisonUnit[Register RegisterName, Type Number] struct {
	au arithmeticUnit[Register, Type]
}

func MakeComparisonUnit[Register RegisterName, Type Number](rs RegisterBank[Register, Type]) ComparisonUnit[Register, Type] {
	return &comparisonUnit[Register, Type]{au: arithmeticUnit[Register, Type]{rs: rs}}
}

func (u *comparisonUnit[Register, Type]) Equal(lhs Register, rhs Register, dest Register) error {
	return u.au.BinaryOp(lhs, rhs, dest, func(lhs Type, rhs Type) Type {
		if lhs == rhs {
			return True[Type]()
		} else {
			return False[Type]()
		}
	})
}

func (u *comparisonUnit[Register, Type]) LessThan(lhs Register, rhs Register, dest Register) error {
	return u.au.BinaryOp(lhs, rhs, dest, func(lhs Type, rhs Type) Type {
		if lhs < rhs {
			return True[Type]()
		} else {
			return False[Type]()
		}
	})
}

type integerAlu[Register RegisterName, Type constraints.Integer] struct {
	RingArithmeticUnit[Register, Type]
	BitOpsUnit[Register, Type]
	ComparisonUnit[Register, Type]
}

func MakeIntegerAlu[Register RegisterName, Type constraints.Integer](rs RegisterBank[Register, Type]) IntegerAlu[Register, Type] {
	return &integerAlu[Register, Type]{
		RingArithmeticUnit: MakeRingArithmeticUnit[Register, Type](rs),
		BitOpsUnit:         MakeBitOptsUnit[Register, Type](rs),
		ComparisonUnit:     MakeComparisonUnit[Register, Type](rs),
	}
}

type floatAlu[Register RegisterName, Type constraints.Float] struct {
	RingArithmeticUnit[Register, Type]
	ComparisonUnit[Register, Type]
	au arithmeticUnit[Register, Type]
}

func MakeFloatAlu[Register RegisterName, Type constraints.Float](rs RegisterBank[Register, Type]) FloatAlu[Register, Type] {
	return &floatAlu[Register, Type]{
		RingArithmeticUnit: MakeRingArithmeticUnit[Register, Type](rs),
		ComparisonUnit:     MakeComparisonUnit[Register, Type](rs),
		au:                 arithmeticUnit[Register, Type]{rs: rs},
	}
}

func (u *floatAlu[Register, Type]) Sin(src Register, dest Register) error {
	return u.au.UnaryOp(src, dest, func(src Type) Type {
		return Type(math.Sin(float64(src)))
	})
}

func (u *floatAlu[Register, Type]) Cos(src Register, dest Register) error {
	return u.au.UnaryOp(src, dest, func(src Type) Type {
		return Type(math.Cos(float64(src)))
	})
}

func (u *floatAlu[Register, Type]) Tan(src Register, dest Register) error {
	return u.au.UnaryOp(src, dest, func(src Type) Type {
		return Type(math.Tan(float64(src)))
	})
}

func (u *floatAlu[Register, Type]) Asin(src Register, dest Register) error {
	return u.au.UnaryOp(src, dest, func(src Type) Type {
		return Type(math.Asin(float64(src)))
	})
}

func (u *floatAlu[Register, Type]) Acos(src Register, dest Register) error {
	return u.au.UnaryOp(src, dest, func(src Type) Type {
		return Type(math.Acos(float64(src)))
	})
}

func (u *floatAlu[Register, Type]) Atan(src Register, dest Register) error {
	return u.au.UnaryOp(src, dest, func(src Type) Type {
		return Type(math.Atan(float64(src)))
	})
}
