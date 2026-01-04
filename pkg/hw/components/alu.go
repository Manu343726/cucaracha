package components

import (
	"github.com/Manu343726/cucaracha/pkg/hw/component"
)

func init() {
	registerALUComponents()
}

// ALU operation codes matching cucaracha CPU operations
type ALUOp uint8

const (
	ALUOp_NOP ALUOp = iota // No operation - output = A
	ALUOp_ADD              // A + B
	ALUOp_SUB              // A - B
	ALUOp_MUL              // A * B
	ALUOp_DIV              // A / B (signed)
	ALUOp_MOD              // A % B (signed)
	ALUOp_AND              // A & B
	ALUOp_OR               // A | B
	ALUOp_XOR              // A ^ B
	ALUOp_NOT              // ~A
	ALUOp_LSL              // A << B (logical shift left)
	ALUOp_LSR              // A >> B (logical shift right)
	ALUOp_ASL              // A << B (arithmetic shift left, same as LSL)
	ALUOp_ASR              // A >> B (arithmetic shift right, sign-extend)
	ALUOp_CMP              // Compare A and B, output CPSR flags
)

const CategoryALU = "alu"

func registerALUComponents() {
	// ALU
	Registry.Register(component.NewDescriptor("ALU").
		DisplayName("ALU").
		Description("32-bit Arithmetic Logic Unit - performs arithmetic and logic operations").
		Category(CategoryALU).
		Version("1.0.0").
		Input("A", 32, "First operand (32-bit)").
		Input("B", 32, "Second operand (32-bit)").
		Input("OP", 4, "Operation code (0-14)").
		Output("OUT", 32, "Result (32-bit)").
		Output("FLAGS", 32, "CPSR flags (Zero, Negative, Carry, Overflow)").
		Factory(func(name string, params map[string]interface{}) (component.Component, error) {
			return NewALU(name), nil
		}).
		Build())
}

// =============================================================================
// ALU Component
// =============================================================================

// CPSR flag bit positions
const (
	FlagZero     uint32 = 1 << 0 // Z: Result is zero
	FlagNegative uint32 = 1 << 1 // N: Result is negative
	FlagCarry    uint32 = 1 << 2 // C: Carry out
	FlagOverflow uint32 = 1 << 3 // V: Signed overflow
)

// ALU is a 32-bit Arithmetic Logic Unit
type ALU struct {
	*component.BaseComponent

	// Input ports
	inputA *component.StandardPort // 32-bit first operand
	inputB *component.StandardPort // 32-bit second operand
	opCode *component.StandardPort // 4-bit operation code

	// Output ports
	output *component.StandardPort // 32-bit result
	flags  *component.StandardPort // 32-bit flags (CPSR)
}

// NewALU creates a new 32-bit ALU
func NewALU(name string) *ALU {
	alu := &ALU{
		BaseComponent: component.NewBaseComponent(name, "ALU"),
	}

	// Create input ports
	alu.inputA = component.NewInputPort("A", 32)
	alu.AddInput(alu.inputA)

	alu.inputB = component.NewInputPort("B", 32)
	alu.AddInput(alu.inputB)

	alu.opCode = component.NewInputPort("OP", 4)
	alu.AddInput(alu.opCode)

	// Create output ports
	alu.output = component.NewOutputPort("OUT", 32)
	alu.AddOutput(alu.output)

	alu.flags = component.NewOutputPort("FLAGS", 32)
	alu.AddOutput(alu.flags)

	return alu
}

// InputA returns the first operand input port
func (alu *ALU) InputA() *component.StandardPort {
	return alu.inputA
}

// InputB returns the second operand input port
func (alu *ALU) InputB() *component.StandardPort {
	return alu.inputB
}

// OpCode returns the operation code input port
func (alu *ALU) OpCode() *component.StandardPort {
	return alu.opCode
}

// Output returns the result output port
func (alu *ALU) Output() *component.StandardPort {
	return alu.output
}

// Flags returns the flags output port
func (alu *ALU) Flags() *component.StandardPort {
	return alu.flags
}

// SetOperands is a convenience method to set both operands
func (alu *ALU) SetOperands(a, b uint32) {
	alu.inputA.SetValue(uint64(a))
	alu.inputB.SetValue(uint64(b))
}

// SetOperation sets the ALU operation code
func (alu *ALU) SetOperation(op ALUOp) {
	alu.opCode.SetValue(uint64(op))
}

// Result returns the current output value
func (alu *ALU) Result() uint32 {
	return uint32(alu.output.GetValue())
}

// GetFlags returns the current flags value
func (alu *ALU) GetFlags() uint32 {
	return uint32(alu.flags.GetValue())
}

// IsZero returns true if the zero flag is set
func (alu *ALU) IsZero() bool {
	return (alu.GetFlags() & FlagZero) != 0
}

// IsNegative returns true if the negative flag is set
func (alu *ALU) IsNegative() bool {
	return (alu.GetFlags() & FlagNegative) != 0
}

// HasCarry returns true if the carry flag is set
func (alu *ALU) HasCarry() bool {
	return (alu.GetFlags() & FlagCarry) != 0
}

// HasOverflow returns true if the overflow flag is set
func (alu *ALU) HasOverflow() bool {
	return (alu.GetFlags() & FlagOverflow) != 0
}

// Compute performs the ALU operation (combinational logic)
func (alu *ALU) Compute() error {
	if !alu.IsEnabled() {
		return nil
	}

	a := uint32(alu.inputA.GetValue())
	b := uint32(alu.inputB.GetValue())
	op := ALUOp(alu.opCode.GetValue() & 0xF)

	var result uint32
	var flags uint32

	switch op {
	case ALUOp_NOP:
		result = a

	case ALUOp_ADD:
		result = a + b
		// Check for carry (unsigned overflow)
		if result < a {
			flags |= FlagCarry
		}
		// Check for signed overflow: overflow occurs when signs of operands are same
		// but sign of result is different
		if ((a ^ result) & (b ^ result) & 0x80000000) != 0 {
			flags |= FlagOverflow
		}

	case ALUOp_SUB:
		result = a - b
		// Carry flag is set if there's NO borrow (a >= b for unsigned)
		if a >= b {
			flags |= FlagCarry
		}
		// Check for signed overflow
		if ((a ^ b) & (a ^ result) & 0x80000000) != 0 {
			flags |= FlagOverflow
		}

	case ALUOp_MUL:
		result = a * b
		// For multiplication, overflow if result doesn't fit in 32 bits
		// We can detect this by checking if result/b != a (when b != 0)
		if b != 0 && result/b != a {
			flags |= FlagOverflow
		}

	case ALUOp_DIV:
		if b == 0 {
			result = 0 // Division by zero returns 0
		} else {
			result = uint32(int32(a) / int32(b))
		}

	case ALUOp_MOD:
		if b == 0 {
			result = 0 // Modulo by zero returns 0
		} else {
			result = uint32(int32(a) % int32(b))
		}

	case ALUOp_AND:
		result = a & b

	case ALUOp_OR:
		result = a | b

	case ALUOp_XOR:
		result = a ^ b

	case ALUOp_NOT:
		result = ^a

	case ALUOp_LSL:
		shift := b & 0x1F // Mask to 5 bits (0-31)
		result = a << shift
		// Carry is the last bit shifted out
		if shift > 0 {
			if (a & (1 << (32 - shift))) != 0 {
				flags |= FlagCarry
			}
		}

	case ALUOp_LSR:
		shift := b & 0x1F
		result = a >> shift
		// Carry is the last bit shifted out
		if shift > 0 {
			if (a & (1 << (shift - 1))) != 0 {
				flags |= FlagCarry
			}
		}

	case ALUOp_ASL:
		shift := b & 0x1F
		result = a << shift
		if shift > 0 {
			if (a & (1 << (32 - shift))) != 0 {
				flags |= FlagCarry
			}
		}

	case ALUOp_ASR:
		shift := b & 0x1F
		result = uint32(int32(a) >> shift) // Sign-extend
		if shift > 0 {
			if (a & (1 << (shift - 1))) != 0 {
				flags |= FlagCarry
			}
		}

	case ALUOp_CMP:
		// Compare: compute A - B and set flags, but don't store result
		result = a - b
		if a >= b {
			flags |= FlagCarry
		}
		if ((a ^ b) & (a ^ result) & 0x80000000) != 0 {
			flags |= FlagOverflow
		}
	}

	// Set Zero flag
	if result == 0 {
		flags |= FlagZero
	}

	// Set Negative flag (sign bit)
	if (result & 0x80000000) != 0 {
		flags |= FlagNegative
	}

	alu.output.SetValue(uint64(result))
	alu.flags.SetValue(uint64(flags))

	return nil
}

// Reset resets the ALU outputs
func (alu *ALU) Reset() {
	alu.output.Reset()
	alu.flags.Reset()
}

// Execute performs an operation directly and returns the result
// This is a convenience method for testing
func (alu *ALU) Execute(op ALUOp, a, b uint32) uint32 {
	alu.SetOperation(op)
	alu.SetOperands(a, b)
	alu.Compute()
	return alu.Result()
}
