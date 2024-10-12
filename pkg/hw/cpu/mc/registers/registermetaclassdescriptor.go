package registers

import (
	"fmt"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/types"
	"github.com/Manu343726/cucaracha/pkg/utils"
)

// Represents a set of register classes that share the same value type, thus making them compatible instruction operands
type RegisterMetaClass struct {
	classes   map[RegisterClass]*RegisterClassDescriptor
	valueType types.ValueType
	Name      string
}

// Returns the value type of all registers in the metaclass
func (mc *RegisterMetaClass) ValueType() types.ValueType {
	return mc.valueType
}

func (mc *RegisterMetaClass) String() string {
	return fmt.Sprintf("<%v:%v>", mc.valueType, utils.FormatSlice(utils.Keys(mc.classes), ","))
}

// Returns the descriptor for the given register class in the metaclass
func (mc *RegisterMetaClass) Class(class RegisterClass) (*RegisterClassDescriptor, error) {
	if descriptor, hasClass := mc.classes[class]; hasClass {
		return descriptor, nil
	} else {
		return nil, utils.MakeError(ErrWrongRegisterClass, "'%v' is not part of this %v register metaclass", class, mc)
	}
}

// Checks if a given register belong to any of the register classes of this metaclass. If not returns a detailed error
func (mc *RegisterMetaClass) RegisterBelongsToClass(register *RegisterDescriptor) error {
	if _, err := mc.Class(register.Class.Class); err == nil {
		return nil
	} else {
		return utils.MakeError(ErrWrongRegisterClass, "expected a %v register, '%v' is %v", mc, register, register.Class)
	}
}

// Returns all registers classes in the metaclass
func (mc *RegisterMetaClass) AllClasses() []*RegisterClassDescriptor {
	return utils.Values(mc.classes)
}

// Returns all registers in the metaclass
func (mc *RegisterMetaClass) AllRegisters() []*RegisterDescriptor {
	return utils.ConcatMap(mc.AllClasses(), (*RegisterClassDescriptor).AllRegisters)
}

// Returns a metaclass of all the given register classes
func MakeRegisterMetaClass(name string, classes []RegisterClass) *RegisterMetaClass {
	if len(classes) <= 0 {
		panic("register metaclass cannot be empty")
	}

	classesMap := utils.GenMapFromKeys(classes, RegisterClasses.Class)

	valueType := classesMap[classes[0]].ValueType

	for _, class := range classesMap {
		if class.ValueType != valueType {
			panic(fmt.Errorf("all classes of this register metaclass must have value type %v. Class %v has value type %v", valueType, class.Class, class.ValueType))
		}
	}

	return &RegisterMetaClass{
		classes:   classesMap,
		valueType: valueType,
		Name:      name,
	}
}
