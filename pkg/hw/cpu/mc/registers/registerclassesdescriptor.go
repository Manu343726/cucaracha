package registers

import (
	"errors"
	"fmt"
	"math/bits"
	"strconv"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/utils"
)

type RegisterClassesDescriptor struct {
	classes map[RegisterClass]*RegisterClassDescriptor
}

// Returns the descriptor of a register class
func (d *RegisterClassesDescriptor) Class(rc RegisterClass) *RegisterClassDescriptor {
	return d.classes[rc]
}

// Returns all the register classes
func (d *RegisterClassesDescriptor) AllClasses() []*RegisterClassDescriptor {
	return utils.Values(d.classes)
}

// Returns the minimal number of bits required to encode all register classes
func (d *RegisterClassesDescriptor) RegisterClassBits() int {
	return bits.Len(uint(len(d.classes)))
}

// Returns the minimal number of bits required to univocally encode a register
func (d *RegisterClassesDescriptor) RegisterBits() int {
	return d.RegisterClassBits() + utils.Max(utils.Map(d.AllClasses(), func(class *RegisterClassDescriptor) int { return class.RegisterBits() }))
}

var ErrInvalidRegisterClass = errors.New("invalid register class")

// Returns a register class given its binary representation
func (d *RegisterClassesDescriptor) Decode(classBinaryRepresentation uint64) (*RegisterClassDescriptor, error) {
	class := RegisterClass(classBinaryRepresentation)

	if class < TOTAL_REGISTER_CLASSES {
		return d.Class(class), nil
	} else {
		return nil, utils.MakeError(ErrInvalidRegisterClass, "%v (binary: %v, hex: %v) is not a valid register class",
			classBinaryRepresentation,
			utils.FormatUintBinary(classBinaryRepresentation, bits.Len64(classBinaryRepresentation)),
			utils.FormatUintHex(classBinaryRepresentation, bits.Len64((classBinaryRepresentation)/4)))
	}
}

// Returns a register given its class and index. Equivalent to Class(class).Register(index)
func (d *RegisterClassesDescriptor) Register(class RegisterClass, index int) (*RegisterDescriptor, error) {
	return d.Class(class).Register(index)
}

// Returns a register given its name
func (d *RegisterClassesDescriptor) RegisterByName(name string) (*RegisterDescriptor, error) {
	for _, class := range d.AllClasses() {
		if registerIndexStr, hasPrefix := strings.CutPrefix(name, class.RegisterNamePrefix); hasPrefix {
			registerIndex, err := strconv.Atoi(registerIndexStr)

			if err == nil {
				return class.Register(registerIndex)
			}
		}

		for _, register := range class.AllRegisters() {
			if register.Name() == name {
				return register, nil
			}
		}
	}

	return nil, utils.MakeError(ErrUnknownRegister, "'%v'", name)
}

// Returns a register given its binary representation
func (d *RegisterClassesDescriptor) DecodeRegister(binaryRepresentation uint64) (*RegisterDescriptor, error) {
	view := utils.CreateBitView(&binaryRepresentation)
	classBits := d.RegisterClassBits()
	totalBits := d.RegisterBits()
	registerIndexBits := totalBits - classBits

	class, err := d.Decode(view.Read(registerIndexBits, classBits))

	if err != nil {
		return nil, err
	}

	index := int(view.Read(0, registerIndexBits))

	return class.Register(index)
}

// Initializes a register classes descriptor with all the given register class descriptors
func NewRegisterClassesDescriptor(classes []*RegisterClassDescriptor) RegisterClassesDescriptor {
	classMap := utils.GenMap(classes, func(class *RegisterClassDescriptor) RegisterClass {
		return class.Class
	})

	for _, class := range utils.Iota(int(TOTAL_REGISTER_CLASSES), func(i int) RegisterClass { return RegisterClass(i) }) {
		if descriptor, hasClass := classMap[class]; !hasClass {
			panic(fmt.Sprintf("missing entry for register class '%v' in registers classes descriptor. Make sure you've added an entry for all register classes in the NewRegisterClassesDescriptor() call", class))
		} else {
			// Make sure all registers in the class have the right class
			for _, register := range descriptor.registers {
				register.Class = descriptor
			}
		}
	}

	d := RegisterClassesDescriptor{
		classes: classMap,
	}

	return d
}
