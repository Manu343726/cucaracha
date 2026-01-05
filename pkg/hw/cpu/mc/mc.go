package mc

import (
	"fmt"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/registers"
)

// Contains implementation information about the machine code
type MachineCodeDescriptor struct {
	// Information about instruction opcodes
	OpCodes *instructions.OpCodesDescriptor
	// Information about machine instructions
	Instructions *instructions.InstructionsDescriptor
	// Information about machine registers classes
	RegisterClasses *registers.RegisterClassesDescriptor
	// Information of machine register meta classes
	RegisterMetaClasses []*registers.RegisterMetaClass
	// Information about condition codes
	ConditionCodes *instructions.ConditionCodesTemplateData
	// Fast lookup of ISA registers (filled automatically from descriptors at initialization)
	Registers FastRegisterLookup

	Name string
}

// Dumps all the MC description as one big multiline string
func (d *MachineCodeDescriptor) Documentation(leftpad int) string {
	leftpad_str := strings.Repeat(" ", leftpad)

	var builder strings.Builder

	builder.WriteString(leftpad_str)
	builder.WriteString(fmt.Sprintf("total supported opcodes: %v\n", d.OpCodes.TotalOpCodes()))
	builder.WriteString(leftpad_str)
	builder.WriteString(fmt.Sprintf("total implemented instructions: %v\n", d.OpCodes.TotalOpCodes()))
	builder.WriteString(leftpad_str)
	builder.WriteString(fmt.Sprintf("instruction encoding lengh (bits): %v\n", d.Instructions.InstructionBits()))
	builder.WriteString(leftpad_str)
	builder.WriteString(fmt.Sprintf("opcode encoding lengh (bits): %v\n\n", d.OpCodes.OpCodeBits()))

	builder.WriteString(leftpad_str)
	builder.WriteString("Opcodes:\n\n")

	for _, opCode := range d.OpCodes.AllOpCodes() {
		builder.WriteString(fmt.Sprintf(" - %v%v\n", leftpad_str, opCode))
	}

	builder.WriteString("\n")
	builder.WriteString(leftpad_str)
	builder.WriteString("Instructions:\n\n")

	for _, instruction := range d.Instructions.AllInstructions() {
		builder.WriteString(instruction.Documentation(leftpad + 2))
		builder.WriteString("\n\n")
	}

	return builder.String()
}

// Like Documentation(), but with zero leftpad
func (d *MachineCodeDescriptor) DocString() string {
	return d.Documentation(0)
}

// Stores references to the standard registers for fast lookup
type FastRegisterLookup struct {
	LR   *registers.RegisterDescriptor
	SP   *registers.RegisterDescriptor
	PC   *registers.RegisterDescriptor
	CPSR *registers.RegisterDescriptor
	R    []*registers.RegisterDescriptor
}

func NewFastRegisterLookup() FastRegisterLookup {
	SP, err := registers.RegisterClasses.RegisterByName("sp")
	if err != nil {
		panic("failed to initialize fast register lookup: " + err.Error())
	}
	LR, err := registers.RegisterClasses.RegisterByName("lr")
	if err != nil {
		panic("failed to initialize fast register lookup: " + err.Error())
	}
	PC, err := registers.RegisterClasses.RegisterByName("pc")
	if err != nil {
		panic("failed to initialize fast register lookup: " + err.Error())
	}
	CPSR, err := registers.RegisterClasses.RegisterByName("cpsr")
	if err != nil {
		panic("failed to initialize fast register lookup: " + err.Error())
	}

	R := make([]*registers.RegisterDescriptor, registers.RegisterClasses.Class(registers.RegisterClass_GeneralPurposeInteger).TotalRegisters())
	for i := 0; i < registers.RegisterClasses.Class(registers.RegisterClass_GeneralPurposeInteger).TotalRegisters(); i++ {
		reg, err := registers.RegisterClasses.Register(registers.RegisterClass_GeneralPurposeInteger, i)
		if err != nil {
			panic("failed to initialize fast register lookup: " + err.Error())
		}
		R[i] = reg
	}

	return FastRegisterLookup{
		SP:   SP,
		LR:   LR,
		PC:   PC,
		CPSR: CPSR,
		R:    R,
	}
}

func makeMachineFunctionDescriptor() MachineCodeDescriptor {
	conditionCodes := instructions.GetConditionCodesTemplateData()
	return MachineCodeDescriptor{
		OpCodes:             &instructions.Opcodes,
		Instructions:        &instructions.Instructions,
		RegisterClasses:     &registers.RegisterClasses,
		RegisterMetaClasses: registers.RegisterMetaClasses,
		ConditionCodes:      &conditionCodes,
		Registers:           NewFastRegisterLookup(),
		Name:                "Cucaracha MC",
	}
}

// Contains implementation information about the machine code
var Descriptor MachineCodeDescriptor = makeMachineFunctionDescriptor()
