package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu"
	"github.com/Manu343726/cucaracha/pkg/hw/memory"
	"github.com/Manu343726/cucaracha/pkg/hw/peripheral"
	"github.com/Manu343726/cucaracha/pkg/runtime/program"
)

// Mock implementation of program.ProgramFile
type MockProgramFile struct {
	mock.Mock
}

func (m *MockProgramFile) FileName() string {
	return m.Called().String(0)
}

func (m *MockProgramFile) SourceFile() string {
	return m.Called().String(0)
}

func (m *MockProgramFile) Functions() map[string]program.Function {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(map[string]program.Function)
}

func (m *MockProgramFile) Instructions() []program.Instruction {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]program.Instruction)
}

func (m *MockProgramFile) Globals() []program.Global {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]program.Global)
}

func (m *MockProgramFile) Labels() []program.Label {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]program.Label)
}

func (m *MockProgramFile) MemoryLayout() *memory.MemoryLayout {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*memory.MemoryLayout)
}

func (m *MockProgramFile) DebugInfo() *program.DebugInfo {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*program.DebugInfo)
}

// Mock implementation of Runtime
type MockRuntime struct {
	mock.Mock
}

func (m *MockRuntime) SetBreakpoint(addr uint32) error {
	return m.Called(addr).Error(0)
}

func (m *MockRuntime) ClearBreakpoint(addr uint32) error {
	return m.Called(addr).Error(0)
}

func (m *MockRuntime) SetWatchpoint(r memory.Range) error {
	return m.Called(r).Error(0)
}

func (m *MockRuntime) ClearWatchpoint(r memory.Range) error {
	return m.Called(r).Error(0)
}

func (m *MockRuntime) CPU() cpu.CPU {
	return m.Called().Get(0).(cpu.CPU)
}

func (m *MockRuntime) Memory() memory.Memory {
	return m.Called().Get(0).(memory.Memory)
}

func (m *MockRuntime) MemoryLayout() memory.MemoryLayout {
	return m.Called().Get(0).(memory.MemoryLayout)
}

func (m *MockRuntime) Peripherals() map[string]peripheral.Peripheral {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(map[string]peripheral.Peripheral)
}

func (m *MockRuntime) Reset() error {
	return m.Called().Error(0)
}

func (m *MockRuntime) Step() (*cpu.StepInfo, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cpu.StepInfo), args.Error(1)
}

// Test helper to create a memory layout with specific code region
func createMemoryLayout(codeBase, codeSize uint32) *memory.MemoryLayout {
	return &memory.MemoryLayout{
		TotalSize:               0x10000,
		SystemDescriptorBase:    0x0,
		SystemDescriptorSize:    0x200,
		VectorTableBase:         0x200,
		VectorTableSize:         0x100,
		DataBase:                0x300,
		DataSize:                0x100,
		CodeBase:                codeBase,
		CodeSize:                codeSize,
		HeapBase:                0x500,
		HeapSize:                0x2000,
		StackBase:               0xFFF0,
		StackSize:               0x1000,
		PeripheralBase:          0xF000,
		PeripheralSize:          0x1000,
		PeripheralBaseAddresses: []uint32{},
	}
}

// Test that LoadCode succeeds when program and runtime have matching code regions
func TestLoadCode_MatchingCodeRegion(t *testing.T) {
	mockProgram := new(MockProgramFile)
	mockRuntime := new(MockRuntime)

	// Setup program with code region 0x400, size 0x400
	progLayout := createMemoryLayout(0x400, 0x400)
	mockProgram.On("MemoryLayout").Return(progLayout)
	mockProgram.On("Instructions").Return([]program.Instruction{})

	// Setup runtime with same code region
	runtimeLayout := createMemoryLayout(0x400, 0x400)
	mockRuntime.On("MemoryLayout").Return(*runtimeLayout)

	loader := NewProgramLoader(mockProgram, mockRuntime)
	err := loader.LoadCode()

	assert.NoError(t, err)
	mockProgram.AssertExpectations(t)
	mockRuntime.AssertExpectations(t)
}

// Test that LoadCode fails when program code base doesn't match runtime
func TestLoadCode_CodeBaseMismatch(t *testing.T) {
	mockProgram := new(MockProgramFile)
	mockRuntime := new(MockRuntime)

	// Setup program with code base 0x400
	progLayout := createMemoryLayout(0x400, 0x400)
	mockProgram.On("MemoryLayout").Return(progLayout)

	// Setup runtime with different code base 0x500
	runtimeLayout := createMemoryLayout(0x500, 0x400)
	mockRuntime.On("MemoryLayout").Return(*runtimeLayout)

	loader := NewProgramLoader(mockProgram, mockRuntime)
	err := loader.LoadCode()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "program code base does not match runtime")
	assert.Contains(t, err.Error(), "0x400")
	assert.Contains(t, err.Error(), "0x500")
}

// Test that LoadCode succeeds when runtime code size is larger than program
func TestLoadCode_RuntimeCodeSizeLarger(t *testing.T) {
	mockProgram := new(MockProgramFile)
	mockRuntime := new(MockRuntime)

	// Setup program with code size 0x400
	progLayout := createMemoryLayout(0x400, 0x400)
	mockProgram.On("MemoryLayout").Return(progLayout)
	mockProgram.On("Instructions").Return([]program.Instruction{})

	// Setup runtime with larger code size 0x800
	runtimeLayout := createMemoryLayout(0x400, 0x800)
	mockRuntime.On("MemoryLayout").Return(*runtimeLayout)

	loader := NewProgramLoader(mockProgram, mockRuntime)
	err := loader.LoadCode()

	// Should succeed - program fits in runtime
	assert.NoError(t, err)
	mockProgram.AssertExpectations(t)
	mockRuntime.AssertExpectations(t)
}

// Test that LoadCode fails when program code size is larger than runtime
func TestLoadCode_ProgramCodeSizeTooLarge(t *testing.T) {
	mockProgram := new(MockProgramFile)
	mockRuntime := new(MockRuntime)

	// Setup program with code size 0x800
	progLayout := createMemoryLayout(0x400, 0x800)
	mockProgram.On("MemoryLayout").Return(progLayout)

	// Setup runtime with smaller code size 0x400
	runtimeLayout := createMemoryLayout(0x400, 0x400)
	mockRuntime.On("MemoryLayout").Return(*runtimeLayout)

	loader := NewProgramLoader(mockProgram, mockRuntime)
	err := loader.LoadCode()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "program code does not fit in runtime")
	assert.Contains(t, err.Error(), "0x800")
	assert.Contains(t, err.Error(), "0x400")
}

// Test that LoadCode fails when program code base mismatches or doesn't fit
func TestLoadCode_CodeBaseAndSizeMismatch(t *testing.T) {
	mockProgram := new(MockProgramFile)
	mockRuntime := new(MockRuntime)

	// Setup program with code region 0x400, size 0x800
	progLayout := createMemoryLayout(0x400, 0x800)
	mockProgram.On("MemoryLayout").Return(progLayout)

	// Setup runtime with different code base and smaller size
	runtimeLayout := createMemoryLayout(0x600, 0x400)
	mockRuntime.On("MemoryLayout").Return(*runtimeLayout)

	loader := NewProgramLoader(mockProgram, mockRuntime)
	err := loader.LoadCode()

	assert.Error(t, err)
	// Should fail on base mismatch first
	assert.Contains(t, err.Error(), "program code base does not match runtime")
	assert.Contains(t, err.Error(), "0x400") // program base
	assert.Contains(t, err.Error(), "0x600") // runtime base
}

// Test that LoadCode fails when program has no memory layout
func TestLoadCode_NoMemoryLayout(t *testing.T) {
	mockProgram := new(MockProgramFile)
	mockRuntime := new(MockRuntime)

	// Setup program with nil memory layout
	mockProgram.On("MemoryLayout").Return(nil)

	loader := NewProgramLoader(mockProgram, mockRuntime)
	err := loader.LoadCode()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "program file has no resolved memory addresses")
}

// Test that other regions don't affect the code region check
func TestLoadCode_OtherRegionsMismatch(t *testing.T) {
	mockProgram := new(MockProgramFile)
	mockRuntime := new(MockRuntime)

	// Setup program with code region matching
	progLayout := createMemoryLayout(0x400, 0x400)
	// But different data region
	progLayout.DataBase = 0x800
	mockProgram.On("MemoryLayout").Return(progLayout)
	mockProgram.On("Instructions").Return([]program.Instruction{})

	// Setup runtime with same code region but different data region
	runtimeLayout := createMemoryLayout(0x400, 0x400)
	// Different data region
	runtimeLayout.DataBase = 0x900
	mockRuntime.On("MemoryLayout").Return(*runtimeLayout)

	loader := NewProgramLoader(mockProgram, mockRuntime)
	err := loader.LoadCode()

	// Should succeed because code regions match
	assert.NoError(t, err)
	mockProgram.AssertExpectations(t)
	mockRuntime.AssertExpectations(t)
}

// Test that different heap/stack regions don't affect the check
func TestLoadCode_DifferentHeapStackRegions(t *testing.T) {
	mockProgram := new(MockProgramFile)
	mockRuntime := new(MockRuntime)

	// Setup program
	progLayout := createMemoryLayout(0x400, 0x400)
	progLayout.HeapBase = 0x1000
	progLayout.StackBase = 0x8000
	mockProgram.On("MemoryLayout").Return(progLayout)
	mockProgram.On("Instructions").Return([]program.Instruction{})

	// Setup runtime with different heap/stack but same code region
	runtimeLayout := createMemoryLayout(0x400, 0x400)
	runtimeLayout.HeapBase = 0x2000
	runtimeLayout.StackBase = 0x7000
	mockRuntime.On("MemoryLayout").Return(*runtimeLayout)

	loader := NewProgramLoader(mockProgram, mockRuntime)
	err := loader.LoadCode()

	// Should succeed because code regions match
	assert.NoError(t, err)
	mockProgram.AssertExpectations(t)
	mockRuntime.AssertExpectations(t)
}

// Mock for memory.Memory interface
type MockMemory struct {
	mock.Mock
}

func (m *MockMemory) WriteUint32(address uint32, value uint32) error {
	return m.Called(address, value).Error(0)
}

func (m *MockMemory) ReadUint32(address uint32) (uint32, error) {
	args := m.Called(address)
	return args.Get(0).(uint32), args.Error(1)
}

func (m *MockMemory) WriteUint8(address uint32, value uint8) error {
	return m.Called(address, value).Error(0)
}

func (m *MockMemory) ReadUint8(address uint32) (uint8, error) {
	args := m.Called(address)
	return args.Get(0).(uint8), args.Error(1)
}
