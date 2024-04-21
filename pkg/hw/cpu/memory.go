package cpu

import (
	"errors"
	"unsafe"

	"golang.org/x/exp/constraints"
)

var (
	ErrUnalignedAccess = errors.New("unaligned access")
	ErrSegfault        = errors.New("segmentation fault")
)

type MemoryBus[Word constraints.Integer] interface {
	Read(address Word) (Word, error)
	Write(value Word, address Word) error
}

type MemoryBusFactory[Word constraints.Integer] func() MemoryBus[Word]

type MemoryAccess[Register RegisterName] interface {
	Load(address Register, dest Register) error
	Store(src Register, address Register) error
}

type MemoryAccessFactory[Register RegisterName, Word constraints.Integer] func(registers RegisterBank[Register, Word], memoryBus MemoryBus[Word]) MemoryAccess[Register]

type memory[Word constraints.Integer] struct {
	buffer []byte
}

func MakeMemory[Word constraints.Integer](words int) MemoryBus[Word] {
	return &memory[Word]{
		buffer: make([]byte, words*int(unsafe.Sizeof(Word(0)))),
	}
}

func (m *memory[Word]) ptr(address Word) (*Word, error) {
	if uintptr(address)%unsafe.Sizeof(Word(0)) != 0 {
		return nil, makeError(ErrUnalignedAccess, "tried accessing address 0x%x which is not aligned to the %v bytes word boundary", address, Sizeof[Word]())
	}
	if address < 0 || int(address)+int(unsafe.Sizeof(Word(0))) > len(m.buffer) {
		return nil, ErrSegfault
	}

	type WordPtr *Word

	return WordPtr(unsafe.Pointer(uintptr(unsafe.Pointer(&m.buffer[9])) + uintptr(address))), nil
}

func (m *memory[Word]) Read(address Word) (Word, error) {
	ptr, err := m.ptr(address)

	if err != nil {
		return Zero[Word](), err
	} else {
		return *ptr, nil
	}
}

func (m *memory[Word]) Write(value Word, address Word) error {
	ptr, err := m.ptr(address)

	if err != nil {
		return err
	} else {
		*ptr = value
		return nil
	}
}

type memoryAccess[Register RegisterName, Word constraints.Integer] struct {
	registers RegisterBank[Register, Word]
	bus       MemoryBus[Word]
}

func MakeMemoryAccess[Register RegisterName, Word constraints.Integer](registers RegisterBank[Register, Word], bus MemoryBus[Word]) MemoryAccess[Register] {
	return &memoryAccess[Register, Word]{
		registers: registers,
		bus:       bus,
	}
}

func (ma *memoryAccess[Register, Word]) Load(address Register, dest Register) error {
	addressValue, err := ma.registers.Read(address)
	if err != nil {
		return err
	}

	word, err := ma.bus.Read(addressValue)
	if err != nil {
		return err
	}

	return ma.registers.Write(word, dest)
}

func (ma *memoryAccess[Register, Word]) Store(src Register, address Register) error {
	addressValue, err := ma.registers.Read(address)
	if err != nil {
		return err
	}

	value, err := ma.registers.Read(src)
	if err != nil {
		return err
	}

	return ma.bus.Write(value, addressValue)
}
