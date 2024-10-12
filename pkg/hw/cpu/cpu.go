package cpu

import (
	"golang.org/x/exp/constraints"
)

type MicroCpu[Register RegisterName, Integer constraints.Integer, Float constraints.Float] interface {
	IntegerRegisters() RegisterBank[Register, Integer]
	FloatRegisters() RegisterBank[Register, Float]

	MovInteger() RegisterInterchange[Register]
	MovFloat() RegisterInterchange[Register]
	IntegerToFloat() RegisterConversion[Register]
	FloatToInteger() RegisterConversion[Register]

	IntegerAlu() IntegerAlu[Register, Integer]
	FloatAlu() FloatAlu[Register, Float]
	Memory() MemoryAccess[Register]
}

type MicroCpuFactory[Register RegisterName, Integer constraints.Integer, Float constraints.Float] func(MicroCpuFactories[Register, Integer, Float]) MicroCpu[Register, Integer, Float]

type microCpu[Register RegisterName, Integer constraints.Integer, Float constraints.Float] struct {
	integerRegisters RegisterBank[Register, Integer]
	floatRegisters   RegisterBank[Register, Float]

	integerMov     RegisterInterchange[Register]
	floatMov       RegisterInterchange[Register]
	integerToFloat RegisterConversion[Register]
	floatToInteger RegisterConversion[Register]

	integerAlu IntegerAlu[Register, Integer]
	floatAlu   FloatAlu[Register, Float]

	memoryBus    MemoryBus[Integer]
	memoryAccess MemoryAccess[Register]
}

type MicroCpuRegisters[Register RegisterName] struct {
	IntegerRegisters []Register
	FloatRegisters   []Register
}

type MicroCpuFactories[Register RegisterName, Integer constraints.Integer, Float constraints.Float] struct {
	IntegerRegisters RegisterBankFactory[Register, Integer]
	FloatRegisters   RegisterBankFactory[Register, Float]
	IntegerMov       RegisterInterchangeFactory[Register, Integer]
	FloatMov         RegisterInterchangeFactory[Register, Float]
	IntegerToFloat   RegisterConversionFactory[Register, Integer, Float]
	FloatToInteger   RegisterConversionFactory[Register, Float, Integer]
	IntegerAlu       IntergerAluFactory[Register, Integer]
	FloatAlu         FloatAluFactory[Register, Float]
	MemoryBus        MemoryBusFactory[Integer]
	MemoryAccess     MemoryAccessFactory[Register, Integer]
}

func MakeMicroCpu[Register RegisterName, Integer constraints.Integer, Float constraints.Float](factories MicroCpuFactories[Register, Integer, Float], registers MicroCpuRegisters[Register]) MicroCpu[Register, Integer, Float] {
	integerRegisters := factories.IntegerRegisters(registers.IntegerRegisters...)
	floatRegisters := factories.FloatRegisters(registers.FloatRegisters...)
	memoryBus := factories.MemoryBus()

	microCpu := &microCpu[Register, Integer, Float]{
		integerRegisters: integerRegisters,
		floatRegisters:   floatRegisters,
		integerMov:       factories.IntegerMov(integerRegisters),
		floatMov:         factories.FloatMov(floatRegisters),
		integerToFloat:   factories.IntegerToFloat(integerRegisters, floatRegisters),
		floatToInteger:   factories.FloatToInteger(floatRegisters, integerRegisters),
		integerAlu:       factories.IntegerAlu(integerRegisters),
		floatAlu:         factories.FloatAlu(floatRegisters),
		memoryBus:        memoryBus,
		memoryAccess:     factories.MemoryAccess(integerRegisters, memoryBus),
	}

	return microCpu
}

func (m *microCpu[Register, Integer, Float]) IntegerRegisters() RegisterBank[Register, Integer] {
	return m.integerRegisters
}

func (m *microCpu[Register, Integer, Float]) FloatRegisters() RegisterBank[Register, Float] {
	return m.floatRegisters
}

func (m *microCpu[Register, Integer, Float]) MovInteger() RegisterInterchange[Register] {
	return m.integerMov
}

func (m *microCpu[Register, Integer, Float]) MovFloat() RegisterInterchange[Register] {
	return m.floatMov
}

func (m *microCpu[Register, Integer, Float]) IntegerToFloat() RegisterConversion[Register] {
	return m.integerToFloat
}

func (m *microCpu[Register, Integer, Float]) FloatToInteger() RegisterConversion[Register] {
	return m.floatToInteger
}

func (m *microCpu[Register, Integer, Float]) IntegerAlu() IntegerAlu[Register, Integer] {
	return m.integerAlu
}

func (m *microCpu[Register, Integer, Float]) FloatAlu() FloatAlu[Register, Float] {
	return m.floatAlu
}

func (m *microCpu[Register, Integer, Float]) Memory() MemoryAccess[Register] {
	return m.memoryAccess
}
