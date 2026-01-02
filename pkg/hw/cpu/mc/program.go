package mc

import (
	"fmt"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/registers"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/types"
)

// Reg creates a register operand from a register descriptor
func Reg(name string) (instructions.OperandValue, error) {
	reg, err := Descriptor.RegisterClasses.RegisterByName(name)
	if err != nil {
		return instructions.OperandValue{}, fmt.Errorf("failed to create register operand: %w", err)
	}

	return instructions.RegisterOperandValue(reg), nil
}

// Imm creates an immediate operand with a 32-bit signed integer value
func Imm(value int32) instructions.OperandValue {
	return instructions.ImmediateValue(types.Int32(value))
}

// UImm creates an immediate operand with a 32-bit unsigned integer value
func UImm(value uint32) instructions.OperandValue {
	return instructions.ImmediateValue(types.Int32(int32(value)))
}

// Imm16 creates a 16-bit immediate operand
func Imm16(value uint16) instructions.OperandValue {
	return instructions.ImmediateValue(types.Int32(int32(value)))
}

// Program represents a sequence of instructions
type Program struct {
	Instructions []*instructions.Instruction
}

// NewProgram creates a new empty program
func NewProgram() *Program {
	return &Program{
		Instructions: make([]*instructions.Instruction, 0),
	}
}

// Add appends an instruction to the program
func (p *Program) Add(instr *instructions.Instruction) *Program {
	p.Instructions = append(p.Instructions, instr)
	return p
}

// Len returns the number of instructions in the program
func (p *Program) Len() int {
	return len(p.Instructions)
}

// At returns the instruction at the given index
func (p *Program) At(index int) *instructions.Instruction {
	return p.Instructions[index]
}

// String returns a human-readable representation of the program
func (p *Program) String() string {
	var builder strings.Builder
	for i, instr := range p.Instructions {
		if i > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(fmt.Sprintf("%4d: %s", i, instr.String()))
	}
	return builder.String()
}

// Encode encodes the program to binary format
func (p *Program) Encode() ([]byte, error) {
	result := make([]byte, 0, len(p.Instructions)*4) // 4 bytes per instruction
	for i, instr := range p.Instructions {
		raw := instr.Raw()
		encoded := raw.Encode()
		// Write as little-endian 32-bit
		result = append(result,
			byte(encoded),
			byte(encoded>>8),
			byte(encoded>>16),
			byte(encoded>>24),
		)
		_ = i // avoid unused variable warning
	}
	return result, nil
}

// InstructionBuilder provides a fluent interface for building instructions
type InstructionBuilder struct {
	descriptor *instructions.InstructionDescriptor
	operands   []instructions.OperandValue
	err        error
}

// Instr creates an instruction builder for the given opcode
func Instr(opcode instructions.OpCode) *InstructionBuilder {
	desc, err := instructions.Instructions.Instruction(opcode)
	return &InstructionBuilder{
		descriptor: desc,
		operands:   make([]instructions.OperandValue, 0),
		err:        err,
	}
}

// InstrByName creates an instruction builder for the given mnemonic
func InstrByName(mnemonic string) *InstructionBuilder {
	// Search through all instructions for the matching mnemonic
	for _, desc := range instructions.Instructions.AllInstructions() {
		if desc.OpCode.Mnemonic == mnemonic {
			return &InstructionBuilder{
				descriptor: desc,
				operands:   make([]instructions.OperandValue, 0),
				err:        nil,
			}
		}
	}
	return &InstructionBuilder{
		descriptor: nil,
		operands:   make([]instructions.OperandValue, 0),
		err:        fmt.Errorf("unknown instruction mnemonic: %s", mnemonic),
	}
}

// Op adds an operand to the instruction
func (b *InstructionBuilder) Op(op instructions.OperandValue) *InstructionBuilder {
	if b.err != nil {
		return b
	}
	b.operands = append(b.operands, op)
	return b
}

// Reg adds a register operand by register descriptor
func (b *InstructionBuilder) NamedR(name string) *InstructionBuilder {
	reg, err := Descriptor.RegisterClasses.RegisterByName(name)
	if err != nil {
		b.err = fmt.Errorf("failed to add register operand: %w", err)
		return b
	}

	return b.Op(instructions.RegisterOperandValue(reg))
}

// R adds a general purpose register operand
func (b *InstructionBuilder) R(index int) *InstructionBuilder {
	reg, err := Descriptor.RegisterClasses.Class(registers.RegisterClass_GeneralPurposeInteger).Register(index)

	if err != nil {
		b.err = fmt.Errorf("failed to add register operand: %w", err)
		return b
	}

	return b.Op(instructions.RegisterOperandValue(reg))
}

// I adds a signed immediate operand
func (b *InstructionBuilder) I(value int32) *InstructionBuilder {
	return b.Op(Imm(value))
}

// U adds an unsigned immediate operand
func (b *InstructionBuilder) U(value uint32) *InstructionBuilder {
	return b.Op(UImm(value))
}

// Build constructs the ProgramInstruction, validating operand count and types
func (b *InstructionBuilder) Build() (*instructions.Instruction, error) {
	if b.err != nil {
		return nil, b.err
	}

	return instructions.NewInstruction(b.descriptor, b.operands)
}

// MustBuild constructs the ProgramInstruction, panicking on error
func (b *InstructionBuilder) MustBuild() *instructions.Instruction {
	instr, err := b.Build()
	if err != nil {
		panic(err)
	}
	return instr
}

// Convenience functions for common instructions

// Nop creates a NOP instruction
func Nop() *instructions.Instruction {
	return Instr(instructions.OpCode_NOP).MustBuild()
}

// Mov creates a MOV instruction (dst = src)
func Mov(src, dst string) *instructions.Instruction {
	return Instr(instructions.OpCode_MOV).NamedR(src).NamedR(dst).MustBuild()
}

// MovImm16L creates a MOV_IMM16L instruction (dst = imm16, zeroing upper bits)
func MovImm16L(imm uint16, dst string) *instructions.Instruction {
	return Instr(instructions.OpCode_MOV_IMM16L).I(int32(imm)).NamedR(dst).MustBuild()
}

// MovImm16H creates a MOV_IMM16H instruction (dst = (dst & 0xFFFF) | (imm16 << 16))
func MovImm16H(imm uint16, dst string) *instructions.Instruction {
	return Instr(instructions.OpCode_MOV_IMM16H).I(int32(imm)).NamedR(dst).NamedR(dst).MustBuild()
}

// Ld creates a LD instruction (dst = memory[addr])
func Ld(addr, dst string) *instructions.Instruction {
	return Instr(instructions.OpCode_LD).NamedR(addr).NamedR(dst).MustBuild()
}

// St creates a ST instruction (memory[addr] = src)
func St(src, addr string) *instructions.Instruction {
	return Instr(instructions.OpCode_ST).NamedR(src).NamedR(addr).MustBuild()
}

// Add creates an ADD instruction (dst = src1 + src2)
func Add(src1, src2, dst string) *instructions.Instruction {
	return Instr(instructions.OpCode_ADD).NamedR(src1).NamedR(src2).NamedR(dst).MustBuild()
}

// Sub creates a SUB instruction (dst = src1 - src2)
func Sub(src1, src2, dst string) *instructions.Instruction {
	return Instr(instructions.OpCode_SUB).NamedR(src1).NamedR(src2).NamedR(dst).MustBuild()
}

// Mul creates a MUL instruction (dst = src1 * src2)
func Mul(src1, src2, dst string) *instructions.Instruction {
	return Instr(instructions.OpCode_MUL).NamedR(src1).NamedR(src2).NamedR(dst).MustBuild()
}

// Div creates a DIV instruction (dst = src1 / src2)
func Div(src1, src2, dst string) *instructions.Instruction {
	return Instr(instructions.OpCode_DIV).NamedR(src1).NamedR(src2).NamedR(dst).MustBuild()
}

// Mod creates a MOD instruction (dst = src1 % src2)
func Mod(src1, src2, dst string) *instructions.Instruction {
	return Instr(instructions.OpCode_MOD).NamedR(src1).NamedR(src2).NamedR(dst).MustBuild()
}

// Cmp creates a CMP instruction (dst = compare(src1, src2))
func Cmp(src1, src2, dst string) *instructions.Instruction {
	return Instr(instructions.OpCode_CMP).NamedR(src1).NamedR(src2).NamedR(dst).MustBuild()
}

// Jmp creates a JMP instruction (link = PC+4; PC = target)
func Jmp(target, link string) *instructions.Instruction {
	return Instr(instructions.OpCode_JMP).NamedR(target).NamedR(link).MustBuild()
}

// CJmp creates a CJMP instruction (if mask != 0: link = PC+4; PC = target)
func CJmp(mask, target, link string) *instructions.Instruction {
	return Instr(instructions.OpCode_CJMP).NamedR(mask).NamedR(target).NamedR(link).MustBuild()
}

// Lsl creates a LSL instruction (dst = src1 << src2)
func Lsl(src1, src2, dst string) *instructions.Instruction {
	return Instr(instructions.OpCode_LSL).NamedR(src1).NamedR(src2).NamedR(dst).MustBuild()
}

// Lsr creates a LSR instruction (dst = src1 >> src2, logical)
func Lsr(src1, src2, dst string) *instructions.Instruction {
	return Instr(instructions.OpCode_LSR).NamedR(src1).NamedR(src2).NamedR(dst).MustBuild()
}

// Asl creates an ASL instruction (dst = src1 << src2, arithmetic)
func Asl(src1, src2, dst string) *instructions.Instruction {
	return Instr(instructions.OpCode_ASL).NamedR(src1).NamedR(src2).NamedR(dst).MustBuild()
}

// Asr creates an ASR instruction (dst = src1 >> src2, arithmetic)
func Asr(src1, src2, dst string) *instructions.Instruction {
	return Instr(instructions.OpCode_ASR).NamedR(src1).NamedR(src2).NamedR(dst).MustBuild()
}
