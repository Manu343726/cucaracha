package mc

import (
	"fmt"
	"strings"
)

// Contains implementation information about the machine code
type MachineCodeDescriptor struct {
	// Information about instruction opcodes
	OpCodes *OpCodesDescriptor
	// Information about machine instructions
	Instructions *InstructionsDescriptor
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

func makeMachineFunctionDescriptor() MachineCodeDescriptor {
	return MachineCodeDescriptor{
		OpCodes:      &Descriptor_Opcodes,
		Instructions: &Descriptor_Instructions,
	}
}

// Contains implementation information about the machine code
var Descriptor MachineCodeDescriptor = makeMachineFunctionDescriptor()
