package interpreter

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/registers"
	"github.com/Manu343726/cucaracha/pkg/hw/memory"
)

type executeFunc func(*Interpreter, []*uint32) error

func generateExecutorMap() map[instructions.OpCode]executeFunc {
	executorMap := make(map[instructions.OpCode]executeFunc, len(mc.Descriptor.OpCodes.AllOpCodes()))
	instrImpl := &cpuInstructionsImplementation{}
	instrImplType := reflect.TypeOf(instrImpl)

	for _, opCode := range mc.Descriptor.OpCodes.AllOpCodes() {
		methodName := strings.Title(strings.ToLower(opCode.Mnemonic))
		method, exists := instrImplType.MethodByName(methodName)

		if !exists {
			panic(fmt.Sprintf("no execution function defined for instruction opcode %s", opCode))
		}

		executorMap[opCode.OpCode] = method.Func.Interface().(executeFunc)
	}

	return executorMap
}

// Map of instruction opcodes to their execution functions
var Instruction_Executors map[instructions.OpCode]executeFunc = generateExecutorMap()

type cpuInstructionsImplementation struct{}

func ExecuteInstruction(interpreter *Interpreter, instr *instructions.Instruction) error {
	execFunc, exists := Instruction_Executors[instr.Descriptor.OpCode.OpCode]
	if !exists {
		return fmt.Errorf("no execution function defined for instruction %s", instr)
	}

	operandValues := make([]uint32, len(instr.Descriptor.Operands))
	operandReferences := make([]*uint32, len(instr.Descriptor.Operands))
	destinationOperandIndices := make(map[*registers.RegisterDescriptor]int, len(instr.Descriptor.Operands))

	for i, operand := range instr.OperandValues {
		switch operand.Kind() {
		case instructions.OperandKind_Register:
			var err error
			operandValues[i], err = interpreter.Registers().ReadByDescriptor(operand.Register())
			if err != nil {
				return fmt.Errorf("could not execute instruction %s: failed to read register operand [%d] %s: %w", instr, i, operand.Register().Name(), err)
			}

			if instr.Descriptor.Operands[i].Role == instructions.OperandRole_Destination {
				destinationOperandIndices[operand.Register()] = i
			}
		case instructions.OperandKind_Immediate:
			operandValues[i] = uint32(operand.Immediate().Int32())
		default:
			return fmt.Errorf("could not execute instruction %s: unsupported operand [%d] type", instr, i)
		}

		operandReferences[i] = &operandValues[i]
	}

	err := execFunc(interpreter, operandReferences)

	if err != nil {
		return fmt.Errorf("error executing instruction %s: %w", instr, err)
	}

	// Write back destination operands
	for regDesc, operandIdx := range destinationOperandIndices {
		if err := interpreter.Registers().WriteByDescriptor(regDesc, operandValues[operandIdx]); err != nil {
			return fmt.Errorf("could not execute instruction %s: failed to write back destination register %s: %w", instr, regDesc.Name(), err)
		}
	}

	return nil
}

func NOP(i *Interpreter, operands []uint32) error {
	// No operation
	return nil
}

func MOVIMM16H(i *Interpreter, operands []*uint32) error {
	immediateHighBits := *operands[0]
	current := *operands[1]

	value := (immediateHighBits << 16) | (current & 0x0000FFFF)

	*operands[1] = value
	return nil
}

func MOVIMM16L(i *Interpreter, operands []*uint32) error {
	immediateLowBits := *operands[0]
	current := *operands[1]

	value := (current & 0xFFFF0000) | (immediateLowBits & 0x0000FFFF)

	*operands[1] = value
	return nil
}

func MOV(i *Interpreter, operands []*uint32) error {
	*operands[1] = *operands[0]
	return nil
}

func LD(i *Interpreter, operands []*uint32) error {
	address := *operands[0]
	value, err := memory.ReadUint32(i.Ram(), address)
	if err != nil {
		return fmt.Errorf("error executing LD: failed to read memory at address 0x%X: %w", address, err)
	}

	*operands[1] = value
	return nil
}

func ST(i *Interpreter, operands []*uint32) error {
	address := *operands[0]
	value := *operands[1]

	if err := memory.WriteUint32(i.Ram(), address, value); err != nil {
		return fmt.Errorf("error executing ST: failed to write memory at address 0x%X: %w", address, err)
	}

	return nil
}

func CMP(i *Interpreter, operands []*uint32) error {
	val1 := *operands[0]
	val2 := *operands[1]

	if err := cpu.WriteCPSR(i.Registers(), instructions.ComputeCPSR(val1, val2)); err != nil {
		return fmt.Errorf("error executing CMP: failed to write CPSR register: %w", err)
	}

	return nil
}

func JMP(i *Interpreter, operands []*uint32) error {
	target := *operands[0]
	link := operands[1]
	// Write link register first before changing PC
	if pc, err := cpu.ReadPC(i.Registers()); err != nil {
		return fmt.Errorf("error executing JMP: failed to read PC register: %w", err)
	} else {
		*link = pc
	}

	if err := cpu.WritePC(i.Registers(), target); err != nil {
		return fmt.Errorf("error executing JMP: failed to write PC register: %w", err)
	}

	return nil
}

func CJMP(i *Interpreter, operands []*uint32) error {
	condCode := instructions.ConditionCode(*operands[0])
	target := *operands[1]
	link := operands[2]

	cpsr, err := cpu.ReadCPSR(i.Registers())
	if err != nil {
		return fmt.Errorf("error executing CJMP: failed to read CPSR register: %w", err)
	}

	if instructions.TestCondition(cpsr, condCode) {
		// Write link register first before changing PC
		if pc, err := cpu.ReadPC(i.Registers()); err != nil {
			return fmt.Errorf("error executing CJMP: failed to read PC register: %w", err)
		} else {
			*link = pc
		}

		if err = cpu.WritePC(i.Registers(), target); err != nil {
			return fmt.Errorf("error executing CJMP: failed to write PC register: %w", err)
		}
	}

	return nil
}

func ADD(i *Interpreter, operands []*uint32) error {
	*operands[2] = *operands[0] + *operands[1]
	return nil
}

func SUB(i *Interpreter, operands []*uint32) error {
	*operands[2] = *operands[0] - *operands[1]
	return nil
}

func MUL(i *Interpreter, operands []*uint32) error {
	*operands[2] = *operands[0] * *operands[1]
	return nil
}

func DIV(i *Interpreter, operands []*uint32) error {
	if *operands[1] == 0 {
		// Handle division by zero as needed; here we just return zero
		*operands[2] = 0
	} else {
		*operands[2] = *operands[0] / *operands[1]
	}
	return nil
}

func MOD(i *Interpreter, operands []*uint32) error {
	if *operands[1] == 0 {
		// Handle modulo by zero as needed; here we just return zero
		*operands[2] = 0
	} else {
		*operands[2] = *operands[0] % *operands[1]
	}
	return nil
}

func AND(i *Interpreter, operands []*uint32) error {
	*operands[2] = *operands[0] & *operands[1]
	return nil
}

func OR(i *Interpreter, operands []*uint32) error {
	*operands[2] = *operands[0] | *operands[1]
	return nil
}

func XOR(i *Interpreter, operands []*uint32) error {
	*operands[2] = *operands[0] ^ *operands[1]
	return nil
}

func NOT(i *Interpreter, operands []*uint32) error {
	*operands[1] = ^(*operands[0])
	return nil
}

func LSL(i *Interpreter, operands []*uint32) error {
	*operands[2] = *operands[0] << (*operands[1] & 0x1F)
	return nil
}

func LSR(i *Interpreter, operands []*uint32) error {
	*operands[2] = *operands[0] >> (*operands[1] & 0x1F)
	return nil
}

func ASL(i *Interpreter, operands []*uint32) error {
	*operands[2] = uint32(int32(*operands[0]) << (*operands[1] & 0x1F))
	return nil
}

func ASR(i *Interpreter, operands []*uint32) error {
	*operands[2] = uint32(int32(*operands[0]) >> (*operands[1] & 0x1F))
	return nil
}

func HLT(i *Interpreter, operands []*uint32) error {
	return i.Halt()
}

func EI(i *Interpreter, operands []*uint32) error {
	return i.Interrupts().Enable()
}

func DI(i *Interpreter, operands []*uint32) error {
	return i.Interrupts().Disable()
}

func SWI(i *Interpreter, operands []*uint32) error {
	vector := uint8(*operands[0] & 0xFF)
	return i.Interrupts().Interrupt(vector)
}

func RETI(i *Interpreter, operands []*uint32) error {
	return i.ReturnFromInterrupt()
}
