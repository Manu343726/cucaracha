package debugger

import (
	"encoding/json"
	"fmt"
)

// InstructionOperandKind specifies the type of an instruction operand.
type InstructionOperandKind int

const (
	// OperandKindRegister indicates this operand is a CPU register.
	OperandKindRegister InstructionOperandKind = iota
	// OperandKindImmediate indicates this operand is an immediate/literal value.
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

// Breakpoint represents a code breakpoint at a specific instruction address.
type Breakpoint struct {
	// Unique identifier for this breakpoint.
	ID int `json:"id"`
	// Instruction memory address where this breakpoint is set.
	Address uint32 `json:"address"`
	// Whether this breakpoint is currently active and will trigger execution stops.
	Enabled bool `json:"enabled"`
	// Source code location of the breakpoint (nil if the source location is unknown).
	Location *SourceLocation `json:"location"`
}

// InstructionOperand represents a single operand in an instruction.
type InstructionOperand struct {
	// Type of this operand. See [InstructionOperandKind] for options.
	Kind InstructionOperandKind `json:"kind"`
	// Register operand data (present when Kind == OperandKindRegister). See [Register].
	Register *Register `json:"register"`
	// Immediate value data (present when Kind == OperandKindImmediate).
	Immediate *uint32 `json:"immediate"`
}

// Instruction represents a disassembled CPU instruction with its operands, metadata, and associated breakpoints/watchpoints.
type Instruction struct {
	// Memory address of this instruction.
	Address uint32 `json:"address"`
	// Binary encoding of the instruction.
	Encoding uint32 `json:"encoding"`
	// Assembly mnemonic (e.g., "add", "beq", "sw").
	Mnemonic string `json:"mnemonic"`
	// Complete assembly language representation of the instruction.
	Text string `json:"text"`
	// Instruction operands. See [InstructionOperand] for structure.
	Operands []*InstructionOperand `json:"operands"`
	// Code breakpoints set on this instruction.
	Breakpoints []*Breakpoint `json:"breakpoints"`
	// Memory watchpoints triggered by this instruction.
	Watchpoints []*Watchpoint `json:"watchpoints"`
	// Whether the program counter is currently at this instruction.
	IsCurrentPC bool `json:"isCurrentPc"`
	// Resolved branch target address if this is a branch instruction (nil if not a branch or target unknown).
	BranchTarget *uint32 `json:"branchTarget"`
	// Symbol name at the branch target if known (nil if unknown or not a branch).
	BranchTargetSym *string `json:"branchTargetSym"`
	// Source code line this instruction came from (nil if source location unknown). See [SourceLine].
	SourceLine *SourceLine `json:"sourceLine"`
}

// DisasmResult contains disassembled instructions and optionally their control flow graph.
type DisasmResult struct {
	// Error message if disassembly failed (nil if successful).
	Error error `json:"error"`
	// Disassembled instructions in memory order. See [Instruction] for structure.
	Instructions []*Instruction `json:"instructions"`
	// Control flow graph showing branch relationships between instructions (nil if not generated). See [ControlFlowGraph].
	ControlFlowGraph *ControlFlowGraph `json:"controlFlowGraph"`
}

// ControlFlowGraph represents the control flow of disassembled instructions showing branch relationships.
type ControlFlowGraph struct {
	// Map from instruction address to target address for branch instructions.
	Edges map[uint32]uint32 `json:"edges"`
}

// CurrentInstructionResult contains the instruction at the current program counter.
type CurrentInstructionResult struct {
	// Error message if retrieval failed (nil if successful).
	Error error `json:"error"`
	// Current instruction at the program counter. See [Instruction] for structure.
	Instruction *Instruction `json:"instruction"`
}

// FunctionSymbol describes a function defined in the loaded program.
type FunctionSymbol struct {
	// Function name.
	Name string `json:"name"`
	// Memory address where this function starts (nil if not resolved).
	Address *uint32 `json:"address"`
	// Size in bytes of this function's code (nil if unknown).
	Size *uint32 `json:"size"`
	// Path to the source file containing this function (if available).
	SourceFile string `json:"sourceFile"`
	// Starting line number in source file where this function is defined.
	StartLine int `json:"startLine"`
	// Ending line number in source file where this function definition ends.
	EndLine int `json:"endLine"`
	// Ranges of instruction addresses for this function (e.g., ["0x1000-0x1020", "0x2000-0x2010"] for non-contiguous code).
	InstructionRanges []string `json:"instructionRanges"`
}

// GlobalSymbol describes a global variable or object defined in the loaded program.
type GlobalSymbol struct {
	// Symbol name (as declared in source or debug info).
	Name string `json:"name"`
	// Memory address of this global (nil if not resolved/located).
	Address *uint32 `json:"address"`
	// Size in bytes of this global object.
	Size int `json:"size"`
	// Type/category of this symbol (e.g., "function", "object", "variable").
	SymbolType string `json:"symbolType"`
	// Whether this global has initial data (initialized with non-zero values).
	HasInitData bool `json:"hasInitData"`
	// Length in bytes of initial data if HasInitData is true.
	InitDataLen int `json:"initDataLen"`
}

// LabelSymbol describes a symbolic label in the program (typically branch targets).
type LabelSymbol struct {
	// Label name (symbol identifier).
	Name string `json:"name"`
	// Index into the instructions array if this label points to an instruction (-1 if not pointing to an instruction).
	InstructionIndex int `json:"instructionIndex"`
	// Resolved memory address of this label (nil if not resolved/located).
	Address *uint32 `json:"address"`
}

// SymbolsResult contains function symbols, global symbols, and label symbols matching the request.
type SymbolsResult struct {
	// Error message if symbol lookup failed (nil if successful).
	Error error `json:"error"`
	// Total number of symbols matching the search criteria.
	TotalCount int `json:"totalCount"`
	// Matching function symbols. See [FunctionSymbol] for structure.
	Functions []*FunctionSymbol `json:"functions"`
	// Matching global variable/object symbols. See [GlobalSymbol] for structure.
	Globals []*GlobalSymbol `json:"globals"`
	// Matching label symbols. See [LabelSymbol] for structure.
	Labels []*LabelSymbol `json:"labels"`
}
