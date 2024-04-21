package cpu

import (
	"errors"
)

var (
	ErrUnknownRegister     = errors.New("unknown register")
	ErrInvalidRegisterName = errors.New("invalid register type")
)

type RegisterName interface {
	comparable
}

type RegisterIndex[Register RegisterName, Type Number] interface {
	Get(r Register) (*Type, error)
}

type RegisterBank[Register RegisterName, Type Number] interface {
	Read(r Register) (Type, error)
	Write(value Type, r Register) error
}

type RegisterBankFactory[Register RegisterName, Type Number] func(registers ...Register) RegisterBank[Register, Type]

type registerBankFromIndex[Register RegisterName, Type Number] struct {
	index RegisterIndex[Register, Type]
}

func MakeRegisterBank[Register RegisterName, Type Number](index RegisterIndex[Register, Type]) RegisterBank[Register, Type] {
	return &registerBankFromIndex[Register, Type]{
		index: index,
	}
}

func (b *registerBankFromIndex[Register, Type]) Read(r Register) (Type, error) {
	reg, err := b.index.Get(r)

	if err != nil {
		return 0, err
	}

	return *reg, nil
}

func (b *registerBankFromIndex[Register, Type]) Write(value Type, r Register) error {
	reg, err := b.index.Get(r)

	if err != nil {
		return err
	}

	*reg = value

	return nil
}

type registers[Register RegisterName, Type Number] struct {
	rs map[Register]*Type
}

func MakeRegisters[Register RegisterName, Type Number](usedRegisters ...Register) RegisterBank[Register, Type] {
	rs := make(map[Register]*Type, len(usedRegisters))

	for _, r := range usedRegisters {
		rs[r] = new(Type)
	}

	return MakeRegisterBank[Register, Type](&registers[Register, Type]{
		rs: rs,
	})
}

func (rs *registers[Register, Type]) Get(r Register) (*Type, error) {
	if ptr, contains := rs.rs[r]; contains {
		return ptr, nil
	} else {
		return nil, makeError(ErrUnknownRegister, "'%v'", r)
	}
}

type joinedRegisterBanks[Register RegisterName, Type Number] struct {
	banks []RegisterBank[Register, Type]
}

func (rs *joinedRegisterBanks[Register, Type]) Read(r Register) (Type, error) {
	for _, bank := range rs.banks {
		if value, err := bank.Read(r); err != nil {
			if errors.Is(err, ErrUnknownRegister) {
				continue
			} else {
				return Zero[Type](), err
			}
		} else {
			return value, nil
		}
	}

	return Zero[Type](), makeError(ErrUnknownRegister, "'%v'", r)
}

func (rs *joinedRegisterBanks[Register, Type]) Write(value Type, r Register) error {
	for _, bank := range rs.banks {
		if err := bank.Write(value, r); err != nil {
			if errors.Is(err, ErrUnknownRegister) {
				continue
			} else {
				return err
			}
		} else {
			return nil
		}
	}

	return makeError(ErrUnknownRegister, "'%v'", r)
}

func JoinRegisterBanks[Register RegisterName, Type Number](banks ...RegisterBank[Register, Type]) RegisterBank[Register, Type] {
	return &joinedRegisterBanks[Register, Type]{
		banks: banks,
	}
}

type RegisterInterchange[Register RegisterName] interface {
	Move(src Register, dst Register) error
}

type RegisterInterchangeFactory[Register RegisterName, Type Number] func(rs RegisterBank[Register, Type]) RegisterInterchange[Register]

type registerInterchange[Register RegisterName, Type Number] struct {
	rs RegisterBank[Register, Type]
}

func MakeRegisterInterchange[Register RegisterName, Type Number](rs RegisterBank[Register, Type]) RegisterInterchange[Register] {
	return &registerInterchange[Register, Type]{
		rs: rs,
	}
}

func (ri *registerInterchange[Register, Type]) Move(src Register, dst Register) error {
	srcValue, err := ri.rs.Read(src)

	if err != nil {
		return err
	}

	return ri.rs.Write(srcValue, dst)
}

type RegisterConversion[Register RegisterName] interface {
	Convert(src Register, dst Register) error
}

type RegisterConversionFactory[Register RegisterName, SrcType Number, DstType Number] func(src RegisterBank[Register, SrcType], dst RegisterBank[Register, DstType]) RegisterConversion[Register]

type registerConversion[Register RegisterName, SrcType Number, DstType Number] struct {
	src RegisterBank[Register, SrcType]
	dst RegisterBank[Register, DstType]
}

func MakeRegisterConversion[Register RegisterName, SrcType Number, DstType Number](src RegisterBank[Register, SrcType], dst RegisterBank[Register, DstType]) RegisterConversion[Register] {
	return &registerConversion[Register, SrcType, DstType]{
		src: src,
		dst: dst,
	}
}

func (rc *registerConversion[Register, SrcType, DstType]) Convert(src Register, dst Register) error {
	srcValue, err := rc.src.Read(src)

	if err != nil {
		return err
	}

	return rc.dst.Write(DstType(srcValue), dst)
}
