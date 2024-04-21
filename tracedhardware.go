package main

import (
	"fmt"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu"
	"golang.org/x/exp/constraints"
)

type Trace struct {
	Operation    string
	ContextStack []string
	Operands     map[string]string
	Result       string
	Error        error
}

func (t *Trace) Depth() int {
	return len(t.ContextStack)
}

func (t *Trace) Context() string {
	if len(t.ContextStack) > 0 {
		return t.ContextStack[len(t.ContextStack)-1]
	} else {
		return ""
	}
}

func (t *Trace) resultString() string {
	if t.Error != nil {
		return fmt.Sprintf("error: %v", t.Error.Error())
	} else if len(t.Result) > 0 {
		return fmt.Sprintf("result: %v", t.Result)
	} else {
		return ""
	}
}

func (t *Trace) joinOperands() string {
	fields := make([]string, 0, len(t.Operands))

	for name, value := range t.Operands {
		fields = append(fields, fmt.Sprintf("%v: %v", name, value))
	}

	return strings.Join(fields, ", ")
}

func (t *Trace) String() string {
	return fmt.Sprintf("%v %v %s", t.Operation, t.joinOperands(), t.resultString())
}

func (t *Trace) backtrace(buffer *strings.Builder, prefix ...string) {
	for i := range t.ContextStack {
		frame := t.ContextStack[i]

		for _, p := range prefix {
			buffer.WriteString(p)
		}
		buffer.WriteString(fmt.Sprintf("[%v] %v", len(t.ContextStack)-i-1, frame))
		buffer.WriteString("\n")
	}
}

func (t *Trace) Backtrace() string {
	buffer := strings.Builder{}
	t.backtrace(&buffer)
	return buffer.String()
}

func (t *Trace) VerboseString() string {
	buffer := strings.Builder{}
	buffer.WriteString("Trace:\n")
	t.backtrace(&buffer, " ... ")
	buffer.WriteString(t.String())

	return buffer.String()
}

type Tracer interface {
	SaveTrace(t *Trace)
}

type TracerWithContextStack interface {
	Tracer
	CurrentContext() string
	PushContext(body string, args ...any)
	PopContext()
}

type tracerWithContextStack struct {
	Tracer
	ContextStack
}

func MakeTracerWithContextStack(tracer Tracer) TracerWithContextStack {
	return &tracerWithContextStack{
		Tracer:       tracer,
		ContextStack: MakeContextStack(),
	}
}

func (t *tracerWithContextStack) SaveTrace(trace *Trace) {
	trace.ContextStack = append([]string{}, t.stack...)
	t.Tracer.SaveTrace(trace)
}

func (t *tracerWithContextStack) PushContext(body string, args ...any) {
	context := fmt.Sprintf(body, args...)

	t.SaveTrace(&Trace{
		Operation: "BeginContext",
		Operands: map[string]string{
			"context": context,
		},
	})

	t.ContextStack.PushContext(context)
}

func (t *tracerWithContextStack) PopContext() {
	context := t.ContextStack.CurrentContext()
	t.ContextStack.PopContext()

	t.SaveTrace(&Trace{
		Operation: "EndContext",
		Operands: map[string]string{
			"context": context,
		},
	})
}

type ContextStack struct {
	stack []string
}

func MakeContextStack() ContextStack {
	stack := ContextStack{
		stack: make([]string, 0, 17),
	}

	stack.PushContext("root")

	return stack
}

func (s *ContextStack) PushContext(body string, args ...any) {
	s.stack = append(s.stack, fmt.Sprintf(body, args...))
}

func (s *ContextStack) PopContext() {
	if len(s.stack) <= 1 {
		return
	}

	s.stack = s.stack[:len(s.stack)-1]
}

func (s *ContextStack) CurrentContext() string {
	return s.stack[len(s.stack)-1]
}

type tracedModule struct {
	TracerWithContextStack
	name string
}

func makeTracedModule(name string, tracer TracerWithContextStack) tracedModule {
	return tracedModule{
		TracerWithContextStack: tracer,
		name:                   name,
	}
}

func (t *tracedModule) PushContext(body string, args ...any) {
	t.TracerWithContextStack.PushContext(t.name+" "+body, args...)
}

type tracedRegisterBank[Register cpu.RegisterName, Type cpu.Number] struct {
	tracedModule
	cpu.RegisterBank[Register, Type]
}

func MakeTracedRegisterBank[Register cpu.RegisterName, Type cpu.Number](impl cpu.RegisterBank[Register, Type], name string, tracer TracerWithContextStack) cpu.RegisterBank[Register, Type] {
	return &tracedRegisterBank[Register, Type]{
		tracedModule: makeTracedModule(name, tracer),
		RegisterBank: impl,
	}
}

func TracedRegisterBankFactory[Register cpu.RegisterName, Type cpu.Number](factory cpu.RegisterBankFactory[Register, Type], name string, tracer TracerWithContextStack) cpu.RegisterBankFactory[Register, Type] {
	return func(registers ...Register) cpu.RegisterBank[Register, Type] {
		return MakeTracedRegisterBank[Register, Type](factory(registers...), name, tracer)
	}
}

func (t *tracedRegisterBank[Register, Type]) Read(r Register) (Type, error) {
	t.PushContext("Read(r: %v)", r)

	result, err := t.RegisterBank.Read(r)

	t.SaveTrace(&Trace{
		Operation: "Read",
		Operands: map[string]string{
			"r": fmt.Sprint(r),
		},
		Result: fmt.Sprint(result),
		Error:  err,
	})

	t.PopContext()

	return result, err
}

func (t *tracedRegisterBank[Register, Type]) Write(value Type, dest Register) error {
	t.PushContext("Write(value: %v, dest: %v)", value, dest)

	err := t.RegisterBank.Write(value, dest)

	t.SaveTrace(&Trace{
		Operation: "Write",
		Operands: map[string]string{
			"value": fmt.Sprint(value),
			"dest":  fmt.Sprint(dest),
		},
		Error: err,
	})

	t.PopContext()

	return err
}

type tracedMemoryBus[Word constraints.Integer] struct {
	tracedModule
	cpu.MemoryBus[Word]
}

func (t *tracedMemoryBus[Word]) Read(address Word) (Word, error) {
	t.PushContext("Read(address: 0x%x)", address)

	word, err := t.MemoryBus.Read(address)

	t.SaveTrace(&Trace{
		Operation: "Read",
		Operands: map[string]string{
			"address": fmt.Sprintf("0x%x", address),
		},
		Result: fmt.Sprint(word),
		Error:  err,
	})

	t.PopContext()

	return word, err
}

func (t *tracedMemoryBus[Word]) Write(word Word, address Word) error {
	t.PushContext("Write(word: %v, address: 0x%x)", word, address)

	err := t.MemoryBus.Write(word, address)

	t.SaveTrace(&Trace{
		Operation: "Write",
		Operands: map[string]string{
			"word":    fmt.Sprint(word),
			"address": fmt.Sprintf("0x%x", address),
		},
		Error: err,
	})

	t.PopContext()

	return err
}

func MakeTracedMemoryBus[Word constraints.Integer](impl cpu.MemoryBus[Word], name string, tracer TracerWithContextStack) cpu.MemoryBus[Word] {
	return &tracedMemoryBus[Word]{
		tracedModule: makeTracedModule(name, tracer),
		MemoryBus:    impl,
	}
}

func TracedMemoryBusFactory[Word constraints.Integer](factory cpu.MemoryBusFactory[Word], name string, tracer TracerWithContextStack) cpu.MemoryBusFactory[Word] {
	return func() cpu.MemoryBus[Word] {
		return MakeTracedMemoryBus[Word](factory(), name, tracer)
	}
}

type tracedMemoryAccess[Register cpu.RegisterName, Word constraints.Integer] struct {
	tracedModule
	cpu.MemoryAccess[Register]
}

func (t *tracedMemoryAccess[Register, Word]) Load(address Register, dest Register) error {
	t.PushContext("Load(address: %v, dest: %v)", address, dest)

	err := t.MemoryAccess.Load(address, dest)

	t.SaveTrace(&Trace{
		Operation: "Load",
		Operands: map[string]string{
			"address": fmt.Sprint(address),
			"dest":    fmt.Sprint(dest),
		},
		Error: err,
	})

	t.PopContext()

	return err
}

func (t *tracedMemoryAccess[Register, Word]) Store(src Register, address Register) error {
	t.PushContext("Store(src: %v, address: %v)", src, address)

	err := t.MemoryAccess.Store(src, address)

	t.SaveTrace(&Trace{
		Operation: "Store",
		Operands: map[string]string{
			"src":     fmt.Sprint(src),
			"address": fmt.Sprint(address),
		},
		Error: err,
	})

	t.PopContext()

	return err
}

func MakeTracedMemoryAccess[Register cpu.RegisterName, Word constraints.Integer](impl cpu.MemoryAccess[Register], name string, tracer TracerWithContextStack) cpu.MemoryAccess[Register] {
	return &tracedMemoryAccess[Register, Word]{
		tracedModule: makeTracedModule(name, tracer),
		MemoryAccess: impl,
	}
}

func TracedMemoryAccessFactory[Register cpu.RegisterName, Word constraints.Integer](factory cpu.MemoryAccessFactory[Register, Word], name string, tracer TracerWithContextStack) cpu.MemoryAccessFactory[Register, Word] {
	return func(registers cpu.RegisterBank[Register, Word], memoryBus cpu.MemoryBus[Word]) cpu.MemoryAccess[Register] {
		return MakeTracedMemoryAccess[Register, Word](factory(registers, memoryBus), name, tracer)
	}
}

type tracedRegisterInterchange[Register cpu.RegisterName] struct {
	tracedModule
	cpu.RegisterInterchange[Register]
}

func (t *tracedRegisterInterchange[Register]) Move(src Register, dest Register) error {
	t.PushContext("Move(src: %v, dest: %v)", src, dest)

	err := t.RegisterInterchange.Move(src, dest)

	t.SaveTrace(&Trace{
		Operation: "Move",
		Operands: map[string]string{
			"src":  fmt.Sprint(src),
			"dest": fmt.Sprint(dest),
		},
		Error: err,
	})

	t.PopContext()

	return err
}

func MakeTracedRegisterInterchange[Register cpu.RegisterName](impl cpu.RegisterInterchange[Register], name string, tracer TracerWithContextStack) cpu.RegisterInterchange[Register] {
	return &tracedRegisterInterchange[Register]{
		tracedModule:        makeTracedModule(name, tracer),
		RegisterInterchange: impl,
	}
}

func TracedRegisterInterchangeFactory[Register cpu.RegisterName, Type cpu.Number](factory cpu.RegisterInterchangeFactory[Register, Type], name string, tracer TracerWithContextStack) cpu.RegisterInterchangeFactory[Register, Type] {
	return func(registers cpu.RegisterBank[Register, Type]) cpu.RegisterInterchange[Register] {
		return MakeTracedRegisterInterchange[Register](factory(registers), name, tracer)
	}
}

type tracedRegisterConversion[Register cpu.RegisterName, Word constraints.Integer] struct {
	tracedModule
	cpu.RegisterConversion[Register]
}

func (t *tracedRegisterConversion[Register, Word]) Convert(src Register, dest Register) error {
	t.PushContext("Convert(src: %v, dest: %v)", src, dest)

	err := t.RegisterConversion.Convert(src, dest)

	t.SaveTrace(&Trace{
		Operation: "Convert",
		Operands: map[string]string{
			"src":  fmt.Sprint(src),
			"dest": fmt.Sprint(dest),
		},
		Error: err,
	})

	t.PopContext()

	return err
}

func MakeTracedRegisterConversion[Register cpu.RegisterName, Word constraints.Integer](impl cpu.RegisterConversion[Register], name string, tracer TracerWithContextStack) cpu.RegisterConversion[Register] {
	return &tracedRegisterConversion[Register, Word]{
		tracedModule:       makeTracedModule(name, tracer),
		RegisterConversion: impl,
	}
}

func TracedRegisterConversionFactory[Register cpu.RegisterName, Src cpu.Number, Dest cpu.Number](factory cpu.RegisterConversionFactory[Register, Src, Dest], name string, tracer TracerWithContextStack) cpu.RegisterConversionFactory[Register, Src, Dest] {
	return func(src cpu.RegisterBank[Register, Src], dest cpu.RegisterBank[Register, Dest]) cpu.RegisterConversion[Register] {
		return MakeTracedRegisterConversion[Register, Word](factory(src, dest), name, tracer)
	}
}

type tracedMicroCpu[Register cpu.RegisterName, Word constraints.Integer, Float constraints.Float] struct {
	tracedModule
	cpu.MicroCpu[Register, Word, Float]
}

func MakeTracedMicroCpu[Register cpu.RegisterName, Word constraints.Integer, Float constraints.Float](impl cpu.MicroCpu[Register, Word, Float], name string, tracer TracerWithContextStack) cpu.MicroCpu[Register, Word, Float] {
	return &tracedMicroCpu[Register, Word, Float]{
		tracedModule: makeTracedModule(name, tracer),
		MicroCpu:     impl,
	}
}

func TracedMicroCpuFactory[Register cpu.RegisterName, Word constraints.Integer, Float constraints.Float](factory cpu.MicroCpuFactory[Register, Word, Float], name string, tracer TracerWithContextStack) cpu.MicroCpuFactory[Register, Word, Float] {
	return func(factories cpu.MicroCpuFactories[Register, Word, Float]) cpu.MicroCpu[Register, Word, Float] {
		return MakeTracedMicroCpu[Register, Word, Float](factory(factories), name, tracer)
	}
}
