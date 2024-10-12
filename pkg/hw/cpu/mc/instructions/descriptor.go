package instructions

import (
	"fmt"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/registers"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/types"
)

var Opcodes OpCodesDescriptor = NewOpCodesDescriptor(
	map[OpCode]string{
		OpCode_NOP:        "NOP",
		OpCode_MOV_IMM16H: "MOVIMM16H",
		OpCode_MOV_IMM16L: "MOVIMM16L",
		OpCode_MOV:        "MOV",
		OpCode_LD:         "LD",
		OpCode_ST:         "ST",
		OpCode_ADD:        "ADD",
		OpCode_SUB:        "SUB",
		OpCode_MUL:        "MUL",
		OpCode_DIV:        "DIV",
		OpCode_MOD:        "MOD",
	},
)

var Instructions InstructionsDescriptor = NewInstructionsDescriptor([]*InstructionDescriptor{
	Nop(),
	MovImm16H(),
	MovImm16L(),
	Mov(),
	Add(),
	Sub(),
	Mul(),
	Div(),
	Mod(),
})

func Nop() *InstructionDescriptor {
	return &InstructionDescriptor{
		OpCode:      Opcodes.Descriptor(OpCode_NOP),
		Description: "No operation. All internal state except from the program counter stays the same afer the execution of the instruction. Takes only 1 CPU cicle",
		Operands:    nil,
	}
}

func Mov() *InstructionDescriptor {
	return &InstructionDescriptor{
		OpCode:      Opcodes.Descriptor(OpCode_MOV),
		Description: "Copies the value of a 32 bit integer register into another 32 bit integer register",
		Operands: []*OperandDescriptor{
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingPosition: 4,
					EncodingBits:     8,
					Role:             OperandRole_Source,
					Description:      "source register",
				}),
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingPosition: 12,
					EncodingBits:     8,
					Role:             OperandRole_Destination,
					Description:      "destination register",
				}),
		},
		LLVM_PatternTemplate:  "mov $dst, $src",
		LLVM_InstructionFlags: LLVMInstructionFlags_IsMoveReg,
	}
}

func MovImm16H() *InstructionDescriptor {
	return &InstructionDescriptor{
		OpCode:      Opcodes.Descriptor(OpCode_MOV_IMM16H),
		Description: "Copies the 16 most significant bits of a 32 bit immediate into a 32 bit integer register",
		Operands: []*OperandDescriptor{
			{
				EncodingPosition: 4,
				EncodingBits:     16,
				Role:             OperandRole_Source,
				Description:      "source immediate. The lower 16 bits are ignored",
				LLVM_CustomName:  "imm",
			},
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingPosition: 20,
					EncodingBits:     8,
					Role:             OperandRole_Destination,
					Description:      "destination register",
					LLVM_CustomName:  "dst",
				}),
		},
		LLVM_InstructionFlags: LLVMInstructionFlags_IsMoveReg,
	}
}

func MovImm16L() *InstructionDescriptor {
	return &InstructionDescriptor{
		OpCode:      Opcodes.Descriptor(OpCode_MOV_IMM16L),
		Description: "Copies the 16 least significant bits of a 32 bit immediate into a 32 bit integer register",
		Operands: []*OperandDescriptor{
			{
				EncodingPosition:   4,
				EncodingBits:       16,
				Role:               OperandRole_Source,
				Description:        "source immediate. The higher 16 bits are ignored",
				LLVM_CustomName:    "imm",
				LLVM_CustomPattern: "i32imm_lo",
			},
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingPosition: 20,
					EncodingBits:     8,
					Role:             OperandRole_Destination,
					Description:      "destination register",
					LLVM_CustomName:  "dst",
				}),
		},
		LLVM_PatternTemplate:  "set $dst, $imm",
		LLVM_InstructionFlags: LLVMInstructionFlags_IsMoveReg,
	}
}

func Ld() *InstructionDescriptor {
	return &InstructionDescriptor{
		OpCode:      Opcodes.Descriptor(OpCode_LD),
		Description: "Copies a word (32 bit) from a memory location and stores it in a register",
		Operands: []*OperandDescriptor{
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingPosition:   4,
					EncodingBits:       8,
					Role:               OperandRole_Source,
					Description:        "Register containing the memory address to copy from",
					LLVM_CustomName:    "addr",
					LLVM_CustomType:    "memsrc",
					LLVM_CustomPattern: "addr",
				}),
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingPosition: 12,
					EncodingBits:     8,
					Role:             OperandRole_Destination,
					Description:      "destination register",
					LLVM_CustomName:  "dst",
				}),
		},
		LLVM_PatternTemplate:  "(set $dst, (load $addr))",
		LLVM_InstructionFlags: LLVMInstructionFlags_MayLoad,
	}
}

func St() *InstructionDescriptor {
	return &InstructionDescriptor{
		OpCode:      Opcodes.Descriptor(OpCode_ST),
		Description: "Copies a word (32 bit) from a 32 bit integer register and writes it to a location in memory",
		Operands: []*OperandDescriptor{
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingPosition:   4,
					EncodingBits:       8,
					Role:               OperandRole_Source,
					Description:        "Register containing the memory address to copy from",
					LLVM_CustomName:    "addr",
					LLVM_CustomType:    "memsrc",
					LLVM_CustomPattern: "addr",
				}),
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingPosition:   12,
					EncodingBits:       8,
					Role:               OperandRole_Destination,
					Description:        "Register containing the memory address to write to",
					LLVM_CustomName:    "addr",
					LLVM_CustomType:    "memsrc",
					LLVM_CustomPattern: "addr",
				}),
		},
		LLVM_PatternTemplate:  "(str $dst, $addr)",
		LLVM_InstructionFlags: LLVMInstructionFlags_MayStore,
	}
}

func Cmp() *InstructionDescriptor {
	return &InstructionDescriptor{
		OpCode:      Opcodes.Descriptor(OpCode_ST),
		Description: "Compares the values of two 32 bit integer registers and stores the comparison result into CPSR",
		Operands: []*OperandDescriptor{
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingPosition: 4,
					EncodingBits:     8,
					Role:             OperandRole_Source,
					Description:      "First operand",
					LLVM_CustomName:  "lhs",
				}),
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingPosition: 12,
					EncodingBits:     8,
					Role:             OperandRole_Source,
					Description:      "Second operand",
					LLVM_CustomName:  "rhs",
				}),
		},
		LLVM_Defs: []*registers.RegisterDescriptor{registers.Register("cpsr")},
	}
}

func binaryInstruction(opcode OpCode, description string, LLVM_DagNode string) *InstructionDescriptor {
	return &InstructionDescriptor{
		OpCode:      Opcodes.Descriptor(opcode),
		Description: fmt.Sprintf("%v the values of two integer registers and save the result into an integer register", description),
		Operands: []*OperandDescriptor{
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingPosition: 4,
					EncodingBits:     8,
					Role:             OperandRole_Source,
					Description:      "first source register",
				}),
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingPosition: 12,
					EncodingBits:     8,
					Role:             OperandRole_Source,
					Description:      "second source register",
				}),
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					Kind:             OperandKind_Register,
					ValueType:        types.ValueType_Int32,
					EncodingPosition: 20,
					EncodingBits:     8,
					Role:             OperandRole_Destination,
					Description:      "destination register",
				}),
		},
		LLVM_PatternTemplate: fmt.Sprintf("(set $dst, (%v $src1, $src2))", LLVM_DagNode),
	}
}

func Add() *InstructionDescriptor {
	return binaryInstruction(OpCode_ADD, "Adds", "add")
}

func Sub() *InstructionDescriptor {
	return binaryInstruction(OpCode_SUB, "Subsctracts", "sub")
}

func Mul() *InstructionDescriptor {
	return binaryInstruction(OpCode_MUL, "Multiplies", "mul")
}

func Div() *InstructionDescriptor {
	return binaryInstruction(OpCode_DIV, "Divides", "div")
}

func Mod() *InstructionDescriptor {
	return binaryInstruction(OpCode_MOD, "Applies integer modulo between", "mod")
}
