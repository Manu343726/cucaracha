package cpu

import (
	"golang.org/x/exp/constraints"
)

type MicroCpu[Register RegisterName, Word constraints.Integer, Float constraints.Float] interface {
	StateRegisters() RegisterBank[Register, Word]
	InternalWordRegisters() RegisterBank[Register, Word]
	PublicWordRegisters() RegisterBank[Register, Word]
	AllWordRegisters() RegisterBank[Register, Word]
	FloatRegisters() RegisterBank[Register, Float]

	MovWord() RegisterInterchange[Register]
	MovFloat() RegisterInterchange[Register]
	WordToFloat() RegisterConversion[Register]
	FloatToWord() RegisterConversion[Register]

	WordAlu() IntegerAlu[Register, Word]
	FloatAlu() FloatAlu[Register, Float]
	Memory() MemoryAccess[Register]
}

type MicroCpuFactory[Register RegisterName, Word constraints.Integer, Float constraints.Float] func(MicroCpuFactories[Register, Word, Float]) MicroCpu[Register, Word, Float]

type microCpu[Register RegisterName, Word constraints.Integer, Float constraints.Float] struct {
	stateRegisters        RegisterBank[Register, Word]
	internalWordRegisters RegisterBank[Register, Word]
	publicWordRegisters   RegisterBank[Register, Word]
	allWordRegisters      RegisterBank[Register, Word]
	floatRegisters        RegisterBank[Register, Float]

	wordMov     RegisterInterchange[Register]
	floatMov    RegisterInterchange[Register]
	wordToFloat RegisterConversion[Register]
	floatToWord RegisterConversion[Register]

	wordAlu  IntegerAlu[Register, Word]
	floatAlu FloatAlu[Register, Float]

	memoryBus    MemoryBus[Word]
	memoryAccess MemoryAccess[Register]
}

type RegisterFactory[Register RegisterName] func(index int) Register
type RegisterParser[Register RegisterName] func(name string) (Register, error)

type StateRegisters[Register RegisterName] struct {
	ProgramCounter Register
	StackPointer   Register
	FramePointer   Register
	LinkRegister   Register
}

type MicroCpuFactories[Register RegisterName, Word constraints.Integer, Float constraints.Float] struct {
	StateRegisterNames       StateRegisters[Register]
	InternalWordRegisterName RegisterFactory[Register]
	PublicWordRegisterName   RegisterFactory[Register]
	FloatRegisterName        RegisterFactory[Register]
	StateRegisters           RegisterBankFactory[Register, Word]
	InternalWordRegisters    RegisterBankFactory[Register, Word]
	PublicWordRegisters      RegisterBankFactory[Register, Word]
	FloatRegisters           RegisterBankFactory[Register, Float]
	WordMov                  RegisterInterchangeFactory[Register, Word]
	FloatMov                 RegisterInterchangeFactory[Register, Float]
	WordToFloat              RegisterConversionFactory[Register, Word, Float]
	FloatToWord              RegisterConversionFactory[Register, Float, Word]
	WordAlu                  IntergerAluFactory[Register, Word]
	FloatAlu                 FloatAluFactory[Register, Float]
	MemoryBus                MemoryBusFactory[Word]
	MemoryAccess             MemoryAccessFactory[Register, Word]
}

func registerNames[Register RegisterName](factory RegisterFactory[Register], count int) []Register {
	rs := make([]Register, count)

	for i := range rs {
		rs[i] = factory(i)
	}

	return rs
}

func (f *MicroCpuFactories[Register, Word, Float]) CreateStateRegisters() RegisterBank[Register, Word] {
	return f.StateRegisters(
		f.StateRegisterNames.FramePointer,
		f.StateRegisterNames.LinkRegister,
		f.StateRegisterNames.ProgramCounter,
		f.StateRegisterNames.StackPointer)
}

func (f *MicroCpuFactories[Register, Word, Float]) CreateInternalWordRegisters(count int) RegisterBank[Register, Word] {
	return f.InternalWordRegisters(registerNames[Register](f.InternalWordRegisterName, count)...)
}

func (f *MicroCpuFactories[Register, Word, Float]) CreatePublicWordRegisters(count int) RegisterBank[Register, Word] {
	return f.PublicWordRegisters(registerNames[Register](f.PublicWordRegisterName, count)...)
}

func (f *MicroCpuFactories[Register, Word, Float]) CreateFloatRegisters(count int) RegisterBank[Register, Float] {
	return f.FloatRegisters(registerNames[Register](f.FloatRegisterName, count)...)
}

type MicroCpuSettings struct {
	TotalInternalWordRegisters int
	TotalPublicWordRegisters   int
	TotalFloatRegisters        int
	TotalMemory                int
}

func MakeMicroCpu[Register RegisterName, Word constraints.Integer, Float constraints.Float](settings MicroCpuSettings, factories MicroCpuFactories[Register, Word, Float]) MicroCpu[Register, Word, Float] {
	stateRegisters := factories.StateRegisters()
	internalWordRegisters := factories.CreateInternalWordRegisters(settings.TotalInternalWordRegisters)
	publicWordRegisters := factories.CreatePublicWordRegisters(settings.TotalPublicWordRegisters)
	allWordRegisters := JoinRegisterBanks(stateRegisters, internalWordRegisters, publicWordRegisters)
	floatRegisters := factories.CreateFloatRegisters(settings.TotalFloatRegisters)
	memoryBus := factories.MemoryBus()

	microCpu := &microCpu[Register, Word, Float]{
		stateRegisters:        stateRegisters,
		internalWordRegisters: internalWordRegisters,
		publicWordRegisters:   publicWordRegisters,
		allWordRegisters:      allWordRegisters,
		floatRegisters:        floatRegisters,
		wordMov:               factories.WordMov(allWordRegisters),
		floatMov:              factories.FloatMov(floatRegisters),
		wordToFloat:           factories.WordToFloat(allWordRegisters, floatRegisters),
		floatToWord:           factories.FloatToWord(floatRegisters, allWordRegisters),
		wordAlu:               factories.WordAlu(allWordRegisters),
		floatAlu:              factories.FloatAlu(floatRegisters),
		memoryBus:             memoryBus,
		memoryAccess:          factories.MemoryAccess(allWordRegisters, memoryBus),
	}

	return microCpu
}

func (m *microCpu[Register, Word, Float]) StateRegisters() RegisterBank[Register, Word] {
	return m.stateRegisters
}

func (m *microCpu[Register, Word, Float]) InternalWordRegisters() RegisterBank[Register, Word] {
	return m.internalWordRegisters
}

func (m *microCpu[Register, Word, Float]) PublicWordRegisters() RegisterBank[Register, Word] {
	return m.publicWordRegisters
}

func (m *microCpu[Register, Word, Float]) AllWordRegisters() RegisterBank[Register, Word] {
	return m.allWordRegisters
}

func (m *microCpu[Register, Word, Float]) FloatRegisters() RegisterBank[Register, Float] {
	return m.floatRegisters
}

func (m *microCpu[Register, Word, Float]) MovWord() RegisterInterchange[Register] {
	return m.wordMov
}

func (m *microCpu[Register, Word, Float]) MovFloat() RegisterInterchange[Register] {
	return m.floatMov
}

func (m *microCpu[Register, Word, Float]) WordToFloat() RegisterConversion[Register] {
	return m.wordToFloat
}

func (m *microCpu[Register, Word, Float]) FloatToWord() RegisterConversion[Register] {
	return m.floatToWord
}

func (m *microCpu[Register, Word, Float]) WordAlu() IntegerAlu[Register, Word] {
	return m.wordAlu
}

func (m *microCpu[Register, Word, Float]) FloatAlu() FloatAlu[Register, Float] {
	return m.floatAlu
}

func (m *microCpu[Register, Word, Float]) Memory() MemoryAccess[Register] {
	return m.memoryAccess
}
