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

func makeMachineFunctionDescriptor() MachineCodeDescriptor {
	return MachineCodeDescriptor{
		OpCodes:             &instructions.Opcodes,
		Instructions:        &instructions.Instructions,
		RegisterClasses:     &registers.RegisterClasses,
		RegisterMetaClasses: registers.RegisterMetaClasses,
	}
}

// Contains implementation information about the machine code
var Descriptor MachineCodeDescriptor = makeMachineFunctionDescriptor()
