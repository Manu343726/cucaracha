package interpreter

import (
	"github.com/Manu343726/cucaracha/pkg/hw/cpu"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/registers"
)

// Holds CPU register values
type Registers struct {
	registers map[registers.RegisterClass][]uint32
}

func NewRegisters() cpu.Registers {
	registers := make(map[registers.RegisterClass][]uint32)
	for _, classDesc := range mc.Descriptor.RegisterClasses.AllClasses() {
		registers[classDesc.Class] = make([]uint32, classDesc.TotalRegisters())
	}

	return &Registers{
		registers: registers,
	}
}

func (r *Registers) Reset() error {
	for class, regs := range r.registers {
		for i := range regs {
			regs[i] = 0
		}
		r.registers[class] = regs
	}

	return nil
}

func (r *Registers) lookup(idx uint32) (*registers.RegisterDescriptor, *uint32, error) {
	regDesc, err := mc.Descriptor.RegisterClasses.DecodeRegister(uint64(idx))
	if err != nil {
		return nil, nil, err
	}

	regValue := r.lookupByDescriptor(regDesc)
	return regDesc, regValue, nil
}

func (r *Registers) lookupByDescriptor(regDesc *registers.RegisterDescriptor) *uint32 {
	classRegs, ok := r.registers[regDesc.Class.Class]
	if !ok {
		panic("register class not found in registers map")
	}

	if int(regDesc.Index) >= len(classRegs) {
		panic("register index out of bounds")
	}

	return &classRegs[regDesc.Index]
}

func (r *Registers) ReadByDescriptor(regDesc *registers.RegisterDescriptor) (uint32, error) {
	return *r.lookupByDescriptor(regDesc), nil
}

func (r *Registers) WriteByDescriptor(regDesc *registers.RegisterDescriptor, value uint32) error {
	*r.lookupByDescriptor(regDesc) = value
	return nil
}

func (r *Registers) Read(idx uint32) (uint32, error) {
	_, value, err := r.lookup(idx)
	if err != nil {
		return 0, err
	}

	return *value, nil
}

func (r *Registers) Write(idx uint32, value uint32) error {
	_, regValue, err := r.lookup(idx)
	if err != nil {
		return err
	}

	*regValue = value
	return nil
}
