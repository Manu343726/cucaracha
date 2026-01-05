package debugger

import (
	"github.com/Manu343726/cucaracha/pkg/hw/cpu"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/registers"
	"github.com/Manu343726/cucaracha/pkg/runtime"
)

type Registers struct {
	r cpu.Registers
}

func NewRegisters(runtime runtime.Runtime) Registers {
	return Registers{
		r: runtime.CPU().Registers(),
	}
}

func (r Registers) ReadRegisters(class registers.RegisterClass) map[string]uint32 {
	result := make(map[string]uint32)

	classDesc := mc.Descriptor.RegisterClasses.Class(class)

	for _, descriptor := range classDesc.AllRegisters() {
		value, err := r.r.ReadByDescriptor(descriptor)
		if err != nil {
			panic("failed to read register value: " + err.Error())
		}
		result[descriptor.Name()] = value
	}

	return result
}

var recommendedStringFormat map[string]string = func() map[string]string {
	result := make(map[string]string)

	for _, classDesc := range mc.Descriptor.RegisterClasses.AllClasses() {
		for _, descriptor := range classDesc.AllRegisters() {
			result[descriptor.Name()] = descriptor.RecommendedStringFormat
		}
	}

	return result
}()

func RecommendedRegisterStringFormat() map[string]string {
	return recommendedStringFormat
}

func (r Registers) ReadStateRegisters() map[string]uint32 {
	return r.ReadRegisters(registers.RegisterClass_StateRegisters)
}

func (r Registers) ReadGeneralPurposeRegisters() map[string]uint32 {
	return r.ReadRegisters(registers.RegisterClass_GeneralPurposeInteger)
}

func RecommendedRegisterColor[Color any](palette []Color) map[string]Color {
	result := make(map[string]Color)
	registerClasses := mc.Descriptor.RegisterClasses.AllClasses()

	for i, classDesc := range registerClasses {
		classColor := palette[i%len(registerClasses)]

		for _, descriptor := range classDesc.AllRegisters() {
			result[descriptor.Name()] = classColor
		}
	}

	return result
}
