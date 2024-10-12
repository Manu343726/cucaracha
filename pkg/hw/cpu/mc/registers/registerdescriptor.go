package registers

import (
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/types"
	"github.com/Manu343726/cucaracha/pkg/utils"
)

type RegisterDescriptor struct {
	// Register class
	Class *RegisterClassDescriptor

	// Index within the register class
	Index int

	// Custom name for the register instead of the default RegisterNamePrefix + Index name
	CustomName string

	// Register description (for documentation/debugging)
	Description string

	// Description details (for documentation/debugging)
	Details string
}

// Returns the register name
func (d *RegisterDescriptor) Name() string {
	if len(d.CustomName) > 0 {
		return d.CustomName
	} else {
		return d.Class.DefaultRegisterName(d.Index)
	}
}

func (d *RegisterDescriptor) String() string {
	return d.Name()
}

// Returns the binary representation of the register
func (d *RegisterDescriptor) Encode() uint64 {
	var result uint64 = 0
	view := utils.CreateBitView(&result)
	classBits := RegisterClasses.RegisterClassBits()
	totalBits := RegisterClasses.RegisterBits()
	registerIndexBits := totalBits - classBits

	view.Write(uint64(d.Index), 0, registerIndexBits)
	view.Write(d.Class.Encode(), registerIndexBits, classBits)

	return result
}

// Returns the value type of the register
func (d *RegisterDescriptor) ValueType() types.ValueType {
	return d.Class.ValueType
}

// Creates multiple consecutive indexed registers
func MakeRegisters(count int) []*RegisterDescriptor {
	return utils.Iota(count, func(i int) *RegisterDescriptor {
		return &RegisterDescriptor{
			Index: i,
		}
	})
}
