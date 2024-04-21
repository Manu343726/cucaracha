package instructions

import "github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"

func Add(src1 uint64, src2 uint64, dest uint64) *mc.Instruction {
	descriptor, err := mc.Descriptor_Instructions.Instruction(mc.OpCode_ADD)

	if err != nil {
		panic(err)
	}

	return &mc.Instruction{
		Descriptor:    descriptor,
		OperandValues: []uint64{src1, src2, dest},
	}
}
