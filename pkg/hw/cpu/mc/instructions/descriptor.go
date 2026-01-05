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
		OpCode_CMP:        "CMP",
		OpCode_JMP:        "JMP",
		OpCode_CJMP:       "CJMP",
		OpCode_LSL:        "LSL",
		OpCode_LSR:        "LSR",
		OpCode_ASL:        "ASL",
		OpCode_ASR:        "ASR",
		OpCode_HLT:        "HLT",
		OpCode_EI:         "EI",
		OpCode_DI:         "DI",
		OpCode_INT:        "INT",
		OpCode_RETI:       "RETI",
		OpCode_SIG:        "SIG",
	},
)

var Instructions InstructionsDescriptor = NewInstructionsDescriptor([]*InstructionDescriptor{
	Nop(),
	MovImm16H(),
	MovImm16L(),
	Mov(),
	Cmp(),
	Ld(),
	St(),
	Jmp(),
	CJmp(),
	Add(),
	Sub(),
	Mul(),
	Div(),
	Mod(),
	Cmp(),
	Lsl(),
	Lsr(),
	Asr(),
	Asl(),
	Hlt(),
	Ei(),
	Di(),
	Int(),
	Reti(),
	Sig(),
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
					EncodingBits: 8,
					Role:         OperandRole_Source,
					Description:  "source register",
				}),
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingBits: 8,
					Role:         OperandRole_Destination,
					Description:  "destination register",
				}),
		},
		LLVM_PatternTemplate: "",
	}
}

func MovImm16H() *InstructionDescriptor {
	return &InstructionDescriptor{
		OpCode:      Opcodes.Descriptor(OpCode_MOV_IMM16H),
		Description: "Copies the 16 most significant bits of a 32 bit immediate into a 32 bit integer register, preserving the low 16 bits",
		Operands: []*OperandDescriptor{
			// Immediate value for high 16 bits (encoded first, like MOVIMM16L)
			{
				EncodingBits:    16,
				Role:            OperandRole_Source,
				Description:     "source immediate. The lower 16 bits are ignored",
				LLVM_CustomName: "imm",
			},
			// Destination register (output)
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingBits:    8,
					Role:            OperandRole_Destination,
					Description:     "destination register",
					LLVM_CustomName: "dst",
				}),
			// Source register (input) - tied to dst, not shown in assembly
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingBits:     0, // Not encoded separately - same as dst
					Role:             OperandRole_Source,
					Description:      "source register (tied to dst, implicit)",
					LLVM_CustomName:  "src",
					LLVM_HideFromAsm: true, // Don't show in assembly, it's always the same as dst
				}),
		},
		LLVM_Constraints: "$src = $dst",
	}
}

func MovImm16L() *InstructionDescriptor {
	return &InstructionDescriptor{
		OpCode:      Opcodes.Descriptor(OpCode_MOV_IMM16L),
		Description: "Copies the 16 least significant bits of a 32 bit immediate into a 32 bit integer register",
		Operands: []*OperandDescriptor{
			{
				EncodingBits:       16,
				Role:               OperandRole_Source,
				Description:        "source immediate. The higher 16 bits are ignored",
				LLVM_CustomName:    "imm",
				LLVM_CustomPattern: "i32imm_lo",
			},
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingBits:    8,
					Role:            OperandRole_Destination,
					Description:     "destination register",
					LLVM_CustomName: "dst",
				}),
		},
		LLVM_PatternTemplate: "(set $dst, $imm)",
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
					EncodingBits:    8,
					Role:            OperandRole_Source,
					Description:     "Register containing the memory address to copy from",
					LLVM_CustomName: "addr",
				}),
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingBits:    8,
					Role:            OperandRole_Destination,
					Description:     "destination register",
					LLVM_CustomName: "dst",
				}),
		},
		LLVM_PatternTemplate:  "", // No pattern - PseudoLD handles load patterns and expands to this
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
					EncodingBits:    8,
					Role:            OperandRole_Source,
					Description:     "Register containing the value to store",
					LLVM_CustomName: "src",
				}),
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingBits:    8,
					Role:            OperandRole_Source,
					Description:     "Register containing the memory address to write to",
					LLVM_CustomName: "addr",
				}),
		},
		LLVM_PatternTemplate:  "", // No pattern - PseudoST handles store patterns and expands to this
		LLVM_InstructionFlags: LLVMInstructionFlags_MayStore,
	}
}

func Cmp() *InstructionDescriptor {
	return &InstructionDescriptor{
		OpCode:      Opcodes.Descriptor(OpCode_CMP),
		Description: "Compares the values of two 32 bit integer registers and stores the comparison result into CPSR",
		Operands: []*OperandDescriptor{
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingBits:    8,
					Role:            OperandRole_Source,
					Description:     "First operand",
					LLVM_CustomName: "lhs",
				}),
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingBits:    8,
					Role:            OperandRole_Source,
					Description:     "Second operand",
					LLVM_CustomName: "rhs",
				}),
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingBits:    8,
					Role:            OperandRole_Destination,
					Description:     "Destination register (CPSR)",
					LLVM_CustomName: "dst",
				}),
		},
		LLVM_PatternTemplate: "(set $dst, (cucaracha_cmp $lhs, $rhs))",
	}
}

func Jmp() *InstructionDescriptor {
	return &InstructionDescriptor{
		OpCode:      Opcodes.Descriptor(OpCode_JMP),
		Description: "Unconditional jump to the address contained in a 32 bit integer register",
		Operands: []*OperandDescriptor{
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingBits:    8,
					Role:            OperandRole_Source,
					Description:     "Register containing the target address",
					LLVM_CustomName: "target",
				}),
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingBits:    8,
					Role:            OperandRole_Destination,
					Description:     "Register where the current instruction address will be saved before the jump",
					LLVM_CustomName: "link",
				},
			),
		},
		LLVM_PatternTemplate:  "", // No pattern - needs custom lowering due to link register output
		LLVM_InstructionFlags: LLVMInstructionFlags_IsBranch | LLVMInstructionFlags_IsBarrier | LLVMInstructionFlags_IsTerminator | LLVMInstructionFlags_IsIndirectBranch,
	}
}

func CJmp() *InstructionDescriptor {
	return &InstructionDescriptor{
		OpCode:      Opcodes.Descriptor(OpCode_CJMP),
		Description: "Conditional jump to the address contained in a 32 bit integer register if the condition code (stored in a register) is satisfied by the current CPSR flags",
		Operands: []*OperandDescriptor{
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingBits:    8,
					Role:            OperandRole_Source,
					Description:     "Register containing the condition code (0-14: EQ, NE, CS, CC, MI, PL, VS, VC, HI, LS, GE, LT, GT, LE, AL)",
					LLVM_CustomName: "condcode",
				}),
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingBits:    8,
					Role:            OperandRole_Source,
					Description:     "Register containing the target address",
					LLVM_CustomName: "target",
				}),
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingBits:    8,
					Role:            OperandRole_Destination,
					Description:     "Link register - stores current PC before jump",
					LLVM_CustomName: "link",
				}),
		},
		LLVM_PatternTemplate:  "", // No pattern - PseudoBRCOND matches cucaracha_brcond and expands to this
		LLVM_InstructionFlags: LLVMInstructionFlags_IsBranch | LLVMInstructionFlags_IsTerminator,
	}
}

func binaryInstruction(opcode OpCode, description string, LLVM_DagNode string, op func(a, b uint32) uint32) *InstructionDescriptor {
	return &InstructionDescriptor{
		OpCode:      Opcodes.Descriptor(opcode),
		Description: fmt.Sprintf("%v the values of two integer registers and save the result into an integer register", description),
		Operands: []*OperandDescriptor{
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingBits: 8,
					Role:         OperandRole_Source,
					Description:  "first source register",
				}),
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					EncodingBits: 8,
					Role:         OperandRole_Source,
					Description:  "second source register",
				}),
			RegisterOperandDescriptor(
				registers.IntegerRegisters,
				OperandDescriptor{
					Kind:         OperandKind_Register,
					ValueType:    types.ValueType_Int32,
					EncodingBits: 8,
					Role:         OperandRole_Destination,
					Description:  "destination register",
				}),
		},
		LLVM_PatternTemplate: fmt.Sprintf("(set $dst, (%v $src1, $src2))", LLVM_DagNode),
	}
}

func Add() *InstructionDescriptor {
	return binaryInstruction(OpCode_ADD, "Adds", "add", func(a, b uint32) uint32 { return a + b })
}

func Sub() *InstructionDescriptor {
	return binaryInstruction(OpCode_SUB, "Subsctracts", "sub", func(a, b uint32) uint32 { return a - b })
}

func Mul() *InstructionDescriptor {
	return binaryInstruction(OpCode_MUL, "Multiplies", "mul", func(a, b uint32) uint32 { return a * b })
}

func Div() *InstructionDescriptor {
	return binaryInstruction(OpCode_DIV, "Divides", "sdiv", func(a, b uint32) uint32 {
		if b == 0 {
			return 0 // Division by zero returns 0 (could also panic)
		}
		return uint32(int32(a) / int32(b))
	})
}

func Mod() *InstructionDescriptor {
	return binaryInstruction(OpCode_MOD, "Applies integer modulo between", "srem", func(a, b uint32) uint32 {
		if b == 0 {
			return 0 // Division by zero returns 0 (could also panic)
		}
		return uint32(int32(a) % int32(b))
	})
}

func Lsl() *InstructionDescriptor {
	return binaryInstruction(OpCode_LSL, "Logical shift left", "shl", func(a, b uint32) uint32 {
		return a << (b & 0x1F)
	})
}

func Lsr() *InstructionDescriptor {
	return binaryInstruction(OpCode_LSR, "Logical shift right", "srl", func(a, b uint32) uint32 {
		return a >> (b & 0x1F)
	})
}

func Asr() *InstructionDescriptor {
	return binaryInstruction(OpCode_ASR, "Arithmetic shift right", "sra", func(a, b uint32) uint32 {
		return uint32(int32(a) >> (b & 0x1F))
	})
}

func Asl() *InstructionDescriptor {
	return binaryInstruction(OpCode_ASL, "Arithmetic shift left", "shl", func(a, b uint32) uint32 {
		return a << (b & 0x1F)
	})
}

// Hlt halts CPU execution
func Hlt() *InstructionDescriptor {
	return &InstructionDescriptor{
		OpCode:      Opcodes.Descriptor(OpCode_HLT),
		Description: "Halts CPU execution. The CPU will stop executing instructions until reset.",
		Operands:    nil,
	}
}

// Ei enables interrupts
func Ei() *InstructionDescriptor {
	return &InstructionDescriptor{
		OpCode:      Opcodes.Descriptor(OpCode_EI),
		Description: "Enables interrupts. After this instruction, the CPU will respond to interrupt requests.",
		Operands:    nil,
	}
}

// Di disables interrupts
func Di() *InstructionDescriptor {
	return &InstructionDescriptor{
		OpCode:      Opcodes.Descriptor(OpCode_DI),
		Description: "Disables interrupts. After this instruction, the CPU will ignore interrupt requests.",
		Operands:    nil,
	}
}

// Triggers an interrupt
func Int() *InstructionDescriptor {
	return &InstructionDescriptor{
		OpCode:      Opcodes.Descriptor(OpCode_INT),
		Description: "Triggers an interrupt with the specified vector number.",
		Operands: []*OperandDescriptor{
			{
				EncodingBits: 8,
				Role:         OperandRole_Source,
				Description:  "interrupt vector number (0-255)",
			},
		},
	}
}

// Reti returns from an interrupt handler
func Reti() *InstructionDescriptor {
	return &InstructionDescriptor{
		OpCode:      Opcodes.Descriptor(OpCode_RETI),
		Description: "Returns from an interrupt handler, restoring saved state and re-enabling interrupts.",
		Operands:    nil,
	}
}

// Sends a signal from the CPU to the outside world
func Sig() *InstructionDescriptor {
	return &InstructionDescriptor{
		OpCode:      Opcodes.Descriptor(OpCode_SIG),
		Description: "Sends a signal to an external peripheral or system component",
		Operands: []*OperandDescriptor{
			{
				EncodingBits: 8,
				Role:         OperandRole_Source,
				Description:  "Signal number",
			},
			{
				EncodingBits: 8,
				Role:         OperandRole_Source,
				Description:  "Target peripheral or component ID",
			},
		},
	}
}
