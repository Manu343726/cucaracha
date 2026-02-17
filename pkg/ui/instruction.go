package ui

import (
	"encoding/json"
	"fmt"
)

// Kind of an instruction operand
type InstructionOperandKind int

const (
	// The operand is a register
	OperandKindRegister InstructionOperandKind = iota
	// The operand is an immediate value
	OperandKindImmediate
)

func (k InstructionOperandKind) String() string {
	switch k {
	case OperandKindRegister:
		return "register"
	case OperandKindImmediate:
		return "immediate"
	default:
		return "unknown"
	}
}

func InstructionOperandKindFromString(s string) (InstructionOperandKind, error) {
	switch s {
	case "register":
		return OperandKindRegister, nil
	case "immediate":
		return OperandKindImmediate, nil
	default:
		return 0, fmt.Errorf("unknown InstructionOperandKind: \"%s\"", s)
	}
}

func (k InstructionOperandKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.String())
}

func (k *InstructionOperandKind) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	val, err := InstructionOperandKindFromString(s)
	if err != nil {
		return err
	}
	*k = val
	return nil
}

// Represents a code breakpoint
type Breakpoint struct {
	ID       int             `json:"id"`       // Breakpoint ID
	Address  uint32          `json:"address"`  // Instruction address of the breakpoint
	Enabled  bool            `json:"enabled"`  // Whether the breakpoint is enabled
	Location *SourceLocation `json:"location"` // Source location of the breakpoint (nil if unknown)
}

// Represents an operand of an instruction
type InstructionOperand struct {
	Kind      InstructionOperandKind `json:"kind"`      // Operand kind
	Register  *Register              `json:"register"`  // Present if Kind == OperandKindRegister
	Immediate *uint32                `json:"immediate"` // Present if Kind == OperandKindImmediate
}

// Contains information about an instruction
type Instruction struct {
	Address         uint32                `json:"address"`         // Instruction address
	Encoding        uint32                `json:"encoding"`        // Binary encoding of the instruction
	Mnemonic        string                `json:"mnemonic"`        // Assembly mnemonic
	Text            string                `json:"text"`            // Full assembly text
	Operands        []*InstructionOperand `json:"operands"`        // Instruction operands
	Breakpoints     []*Breakpoint         `json:"breakpoints"`     // Associated breakpoints
	Watchpoints     []*Watchpoint         `json:"watchpoints"`     // Associated watchpoints
	IsCurrentPC     bool                  `json:"isCurrentPc"`     // Whether this is the current PC
	BranchTarget    *uint32               `json:"branchTarget"`    // Resolved branch target address (nil if not a branch or unknown)
	BranchTargetSym *string               `json:"branchTargetSym"` // Symbol name of branch target (nil if none)
	SourceLocation  *SourceLocation       `json:"sourceLocation"`  // Source location of the instruction (nil if unknown)
}

// Result of Disasm command
type DisassemblyResult struct {
	Error        error          `json:"error"`        // Error, if any
	Instructions []*Instruction `json:"instructions"` // Disassembled instructions
}

// Result of CurrentInstruction command
type CurrentInstructionResult struct {
	Error       error        `json:"error"`       // Error, if any
	Instruction *Instruction `json:"instruction"` // Current instruction
}
