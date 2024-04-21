package main

import (
	"fmt"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu"
)

type Word = int32
type Float float32
type Register = string
type MicroCpu = cpu.MicroCpu[Register, Word, Float]
type MicroCpuFactories = cpu.MicroCpuFactories[Register, Word, Float]

type StdoutTracer struct{}

func (t *StdoutTracer) SaveTrace(trace *Trace) {
	buffer := strings.Builder{}

	for i := 0; i < trace.Depth(); i++ {
		buffer.WriteByte(' ')
	}

	buffer.WriteString(trace.String())
	fmt.Println(buffer.String())
}

func makeCpu() MicroCpu {
	tracer := MakeTracerWithContextStack(&StdoutTracer{})

	settings := cpu.MicroCpuSettings{
		TotalInternalWordRegisters: 4,
		TotalPublicWordRegisters:   8,
		TotalFloatRegisters:        8,
		TotalMemory:                1024,
	}

	cpuFactories := MicroCpuFactories{
		StateRegisterNames: cpu.StateRegisters[Register]{
			ProgramCounter: "pc",
			StackPointer:   "sp",
			FramePointer:   "fp",
			LinkRegister:   "lr",
		},
		InternalWordRegisterName: func(index int) Register {
			return fmt.Sprintf("_w%v", index)
		},
		PublicWordRegisterName: func(index int) Register {
			return fmt.Sprintf("w%v", index)
		},
		FloatRegisterName: func(index int) Register {
			return fmt.Sprintf("f%v", index)
		},
		StateRegisters:        TracedRegisterBankFactory[Register, Word](cpu.MakeRegisters[Register, Word], "state registers", tracer),
		InternalWordRegisters: TracedRegisterBankFactory[Register, Word](cpu.MakeRegisters[Register, Word], "internal word registers", tracer),
		PublicWordRegisters:   TracedRegisterBankFactory[Register, Word](cpu.MakeRegisters[Register, Word], "public word registers", tracer),
		FloatRegisters:        TracedRegisterBankFactory[Register, Float](cpu.MakeRegisters[Register, Float], "float registers", tracer),
		WordMov:               TracedRegisterInterchangeFactory[Register, Word](cpu.MakeRegisterInterchange[Register, Word], "word mov", tracer),
		FloatMov:              TracedRegisterInterchangeFactory[Register, Float](cpu.MakeRegisterInterchange[Register, Float], "float mov", tracer),
		WordToFloat:           TracedRegisterConversionFactory[Register, Word, Float](cpu.MakeRegisterConversion[Register, Word, Float], "word to float conversion", tracer),
		FloatToWord:           TracedRegisterConversionFactory[Register, Float, Word](cpu.MakeRegisterConversion[Register, Float, Word], "float to word conversion", tracer),
		WordAlu:               cpu.MakeIntegerAlu[Register, Word],
		FloatAlu:              cpu.MakeFloatAlu[Register, Float],
		MemoryBus: TracedMemoryBusFactory[Word](func() cpu.MemoryBus[Word] {
			return cpu.MakeMemory[Word](settings.TotalMemory)
		}, "main memory bus", tracer),
		MemoryAccess: TracedMemoryAccessFactory[Register, Word](cpu.MakeMemoryAccess[Register, Word], "main memory access", tracer),
	}

	return MakeTracedMicroCpu[Register, Word, Float](cpu.MakeMicroCpu[Register, Word, Float](settings, cpuFactories), "main cpu", tracer)
}

func main() {
	myCpu := makeCpu()

	registerParser := func(name string) (Register, error) {
		return name, nil
	}

	interpreter := cpu.MakeProgramInterpreter(cpu.MakeSanitizedCommandInterpreter(cpu.MakeCommandInterpreter(cpu.MakeMicroCpuInterpreter(myCpu, registerParser))))

	result, err := interpreter.Run(strings.Split(`
		// hello world I guess...

		WW 1 w0 // write 1 into general purpose word register 0
		RW w0   // read general purpose word register 0
	`, "\n"))

	if err != nil {
		fmt.Printf("program stopped with errors: %v\n", err)
	} else if result != nil {
		fmt.Printf("program finished. result: %v\n", *result)
	} else {
		fmt.Println("program finished")
	}
}
