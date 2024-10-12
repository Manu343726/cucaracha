package registers

import (
	"errors"
	"fmt"
	"math/bits"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/types"
	"github.com/Manu343726/cucaracha/pkg/utils"
)

type RegisterClassDescriptor struct {
	Class              RegisterClass
	Description        string
	ValueType          types.ValueType
	RegisterNamePrefix string

	registers []*RegisterDescriptor
}

// Returns the number of registers in the class
func (d *RegisterClassDescriptor) TotalRegisters() int {
	return len(d.registers)
}

// Returns the set of all registers in the class
func (d *RegisterClassDescriptor) AllRegisters() []*RegisterDescriptor {
	return d.registers
}

var ErrUnknownRegister = errors.New("unknown register")
var ErrWrongRegisterClass = errors.New("wrong register class")

// Returns a register of the class given its index
func (d *RegisterClassDescriptor) Register(index int) (*RegisterDescriptor, error) {
	if index < len(d.registers) {
		return d.registers[index], nil
	} else {
		return nil, utils.MakeError(ErrUnknownRegister, "register with index '%v' not found in register class, '%v' class has only %v registers", index, d.Class, d.TotalRegisters())
	}
}

// Returns the number of bits required to binary encode the index of a register of the class
func (d *RegisterClassDescriptor) RegisterBits() int {
	return bits.Len(uint(d.TotalRegisters()))
}

// Returns the name used to refer to a register of the class in case the register didn't specify a custom one
func (d *RegisterClassDescriptor) DefaultRegisterName(index int) string {
	return d.RegisterNamePrefix + fmt.Sprint(index)
}

// Returns the binary representation of the register class
func (d *RegisterClassDescriptor) Encode() uint64 {
	return uint64(d.Class)
}

// Initializes a register class descriptor with the given registers
func NewRegisterClassDescriptor(descriptor *RegisterClassDescriptor, registers []*RegisterDescriptor) *RegisterClassDescriptor {
	descriptor.registers = registers
	return descriptor
}
