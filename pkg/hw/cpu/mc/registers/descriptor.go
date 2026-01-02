package registers

import "github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/types"

// Contains all the metadata describing the registers and register classes supported by the Cucaracha architecture
var RegisterClasses RegisterClassesDescriptor = NewRegisterClassesDescriptor([]*RegisterClassDescriptor{
	StateRegisters(),
	GeneralPurposeInt32(10),
})

// Contains all the different meta-classes of registers defined by the Cucaracha architecture, which are used to specify the kind of registers accepted by instruction operands
var RegisterMetaClasses []*RegisterMetaClass = []*RegisterMetaClass{
	IntegerRegisters,
}

// CPU state registers descriptor
func StateRegisters() *RegisterClassDescriptor {
	return NewRegisterClassDescriptor(
		&RegisterClassDescriptor{
			Class:              RegisterClass_StateRegisters,
			Description:        "CPU state registers",
			ValueType:          types.ValueType_Int32,
			RegisterNamePrefix: "st",
		},
		[]*RegisterDescriptor{
			Pc(),
			Sp(),
			Cpsr(),
			Lr(),
		},
	)
}

// General purpuse 32 bit integer registers descriptor
func GeneralPurposeInt32(count int) *RegisterClassDescriptor {
	return NewRegisterClassDescriptor(&RegisterClassDescriptor{
		Class:              RegisterClass_GeneralPurposeInteger,
		Description:        "General purpose word-sized 32 bit integer registers",
		ValueType:          types.ValueType_Int32,
		RegisterNamePrefix: "r",
	}, MakeRegisters(count))
}

// Program counter descriptor
func Pc() *RegisterDescriptor {
	return &RegisterDescriptor{
		CustomName:  "pc",
		Description: "Program Counter",
	}
}

// Stack pointer descriptor
func Sp() *RegisterDescriptor {
	return &RegisterDescriptor{
		CustomName:  "sp",
		Description: "Stack Pointer. Points to the current position of the top of the stack",
		Details: `Cucaracha stack grows downwards (stack starts at address 0xffffffff).
				sp points to the lowest address currently used as stack memory. This is also the top of the current stack frame.`,
	}
}

// CPU state register descriptor
func Cpsr() *RegisterDescriptor {
	return &RegisterDescriptor{
		CustomName:  "cpsr",
		Description: "CPU state register. Stores CPU state such as comparison flag, carry flag, etc",
		Details:     "TODO",
	}
}

// Link register descriptor
func Lr() *RegisterDescriptor {
	return &RegisterDescriptor{
		CustomName:  "lr",
		Description: "Link register. Stores the return adress of a call",
		Details:     "TODO",
	}
}

// Returns a register descriptor by name, panics if no such register exists
func Register(name string) *RegisterDescriptor {
	reg, err := RegisterClasses.RegisterByName(name)

	if err != nil {
		panic(err)
	}

	return reg
}
