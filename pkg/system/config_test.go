package system

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_Basic(t *testing.T) {
	yaml := `
version: 1
name: "Test System"
description: "A test system configuration"

memory:
  total: 65536
  code_size: 16384
  data_size: 4096
  stack_size: 8192
  heap_size: 2048

interrupts:
  vectors: 32
`

	config, err := Parse([]byte(yaml))
	require.NoError(t, err)

	assert.Equal(t, 1, config.Version)
	assert.Equal(t, "Test System", config.Name)
	assert.Equal(t, "A test system configuration", config.Description)

	assert.Equal(t, uint32(65536), config.Memory.Total)
	assert.Equal(t, uint32(16384), config.Memory.CodeSize)
	assert.Equal(t, uint32(4096), config.Memory.DataSize)
	assert.Equal(t, uint32(8192), config.Memory.StackSize)
	assert.Equal(t, uint32(2048), config.Memory.HeapSize)

	assert.Equal(t, uint32(32), config.Interrupts.Vectors)
}

func TestParse_WithPeripherals(t *testing.T) {
	yaml := `
version: 1
name: "System with Peripherals"

memory:
  total: 65536

peripherals:
  - type: terminal
    params:
      name: uart0
      size: 256
      interrupt_vector: 16

  - type: terminal
    params:
      name: uart1
      size: 256
      interrupt_vector: 17
`

	config, err := Parse([]byte(yaml))
	require.NoError(t, err)

	assert.Len(t, config.Peripherals, 2)

	assert.Equal(t, "terminal", config.Peripherals[0].Type)
	assert.Equal(t, "uart0", config.Peripherals[0].Params.Name)
	assert.Equal(t, uint32(256), config.Peripherals[0].Params.Size)
	assert.Equal(t, uint8(16), config.Peripherals[0].Params.InterruptVector)

	assert.Equal(t, "terminal", config.Peripherals[1].Type)
	assert.Equal(t, "uart1", config.Peripherals[1].Params.Name)
	assert.Equal(t, uint8(17), config.Peripherals[1].Params.InterruptVector)
}

func TestParse_DefaultVersion(t *testing.T) {
	yaml := `
name: "No Version"
memory:
  total: 32768
`

	config, err := Parse([]byte(yaml))
	require.NoError(t, err)

	// Should default to ConfigVersion
	assert.Equal(t, ConfigVersion, config.Version)
}

func TestParse_UnsupportedVersion(t *testing.T) {
	yaml := `
version: 999
name: "Future Version"
memory:
  total: 32768
`

	_, err := Parse([]byte(yaml))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported config version")
}

func TestParse_InvalidYAML(t *testing.T) {
	yaml := `
version: 1
name: [invalid yaml structure
memory: not a map
`

	_, err := Parse([]byte(yaml))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config")
}

func TestParse_MinimalConfig(t *testing.T) {
	yaml := `
memory:
  total: 4096
`

	config, err := Parse([]byte(yaml))
	require.NoError(t, err)

	assert.Equal(t, ConfigVersion, config.Version)
	assert.Equal(t, "", config.Name)
	assert.Equal(t, uint32(4096), config.Memory.Total)
	assert.Len(t, config.Peripherals, 0)
}

func TestLoadFile_Success(t *testing.T) {
	yaml := `
version: 1
name: "File Test"
memory:
  total: 65536
  code_size: 8192
`

	// Create a temporary file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.yaml")
	err := os.WriteFile(configPath, []byte(yaml), 0644)
	require.NoError(t, err)

	config, err := LoadFile(configPath)
	require.NoError(t, err)

	assert.Equal(t, "File Test", config.Name)
	assert.Equal(t, uint32(65536), config.Memory.Total)
	assert.Equal(t, uint32(8192), config.Memory.CodeSize)
}

func TestLoadFile_NotFound(t *testing.T) {
	_, err := LoadFile("/nonexistent/path/config.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestResolvePeripherals_Success(t *testing.T) {
	yaml := `
version: 1
memory:
  total: 65536

peripherals:
  - type: terminal
    params:
      name: uart0
      size: 256
      interrupt_vector: 16
`

	config, err := Parse([]byte(yaml))
	require.NoError(t, err)

	metadata, err := config.ResolvePeripherals()
	require.NoError(t, err)

	require.Len(t, metadata, 1)
	assert.Equal(t, "uart0", metadata[0].Name)
	assert.Equal(t, uint32(256), metadata[0].Size)
	assert.Equal(t, uint8(16), metadata[0].InterruptVector)
	assert.NotNil(t, metadata[0].Descriptor)
}

func TestResolvePeripherals_UnknownType(t *testing.T) {
	yaml := `
version: 1
memory:
  total: 65536

peripherals:
  - type: unknown_peripheral_type
    params:
      name: unknown
`

	config, err := Parse([]byte(yaml))
	require.NoError(t, err)

	_, err = config.ResolvePeripherals()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown peripheral type")
}

func TestResolvePeripherals_DefaultValues(t *testing.T) {
	yaml := `
version: 1
memory:
  total: 65536

peripherals:
  - type: terminal
    params:
      name: uart0
`

	config, err := Parse([]byte(yaml))
	require.NoError(t, err)

	metadata, err := config.ResolvePeripherals()
	require.NoError(t, err)

	require.Len(t, metadata, 1)
	// Should use default size from descriptor
	assert.Greater(t, metadata[0].Size, uint32(0))
}

func TestSetup_Basic(t *testing.T) {
	yaml := `
version: 1
name: "Setup Test"

memory:
  total: 65536
  code_size: 16384
  data_size: 4096
  stack_size: 8192

interrupts:
  vectors: 32
`

	config, err := Parse([]byte(yaml))
	require.NoError(t, err)

	descriptor, err := config.Setup()
	require.NoError(t, err)

	assert.NotNil(t, descriptor)
	assert.NotNil(t, descriptor.MemoryLayout)
	assert.Len(t, descriptor.Peripherals, 0)
}

func TestSetup_WithPeripherals(t *testing.T) {
	yaml := `
version: 1
name: "Setup with Peripherals"

memory:
  total: 65536
  code_size: 8192
  data_size: 2048
  stack_size: 4096

interrupts:
  vectors: 32

peripherals:
  - type: terminal
    params:
      name: uart0
      size: 256
      interrupt_vector: 16
`

	config, err := Parse([]byte(yaml))
	require.NoError(t, err)

	descriptor, err := config.Setup()
	require.NoError(t, err)

	assert.NotNil(t, descriptor)
	require.Len(t, descriptor.Peripherals, 1)
	metadata := descriptor.Peripherals[0].Metadata()
	assert.Equal(t, "uart0", metadata.Name)
	assert.Greater(t, metadata.BaseAddress, uint32(0))
}

func TestSetup_InsufficientMemory(t *testing.T) {
	yaml := `
version: 1
memory:
  total: 256
  code_size: 16384
  data_size: 4096
  stack_size: 8192

interrupts:
  vectors: 32
`

	config, err := Parse([]byte(yaml))
	require.NoError(t, err)

	_, err = config.Setup()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "memory allocation failed")
}

func TestString_ValidConfig(t *testing.T) {
	yaml := `
version: 1
name: "String Test"

memory:
  total: 65536
  code_size: 8192
  data_size: 2048
  stack_size: 4096

interrupts:
  vectors: 32

peripherals:
  - type: terminal
    params:
      name: uart0
      size: 256
`

	config, err := Parse([]byte(yaml))
	require.NoError(t, err)

	str := config.String()
	assert.Contains(t, str, "String Test")
	assert.Contains(t, str, "Memory Layout")
	assert.Contains(t, str, "Peripherals")
	assert.Contains(t, str, "uart0")
}

func TestString_InvalidConfig(t *testing.T) {
	yaml := `
version: 1
memory:
  total: 100
  code_size: 99999
`

	config, err := Parse([]byte(yaml))
	require.NoError(t, err)

	str := config.String()
	assert.Contains(t, str, "invalid system")
}

func TestMemoryConfig_AllFields(t *testing.T) {
	yaml := `
version: 1
memory:
  total: 131072
  code_size: 32768
  data_size: 8192
  stack_size: 16384
  heap_size: 8192
  peripheral_size: 4096
`

	config, err := Parse([]byte(yaml))
	require.NoError(t, err)

	assert.Equal(t, uint32(131072), config.Memory.Total)
	assert.Equal(t, uint32(32768), config.Memory.CodeSize)
	assert.Equal(t, uint32(8192), config.Memory.DataSize)
	assert.Equal(t, uint32(16384), config.Memory.StackSize)
	assert.Equal(t, uint32(8192), config.Memory.HeapSize)
	assert.Equal(t, uint32(4096), config.Memory.PeripheralSize)
}

func TestPeripheralConfig_UnknownPeripheralType(t *testing.T) {
	yaml := `
version: 1
memory:
  total: 65536

peripherals:
  - type: unknown_type
	params:
	  name: test_peripheral
`
	config, err := Parse([]byte(yaml))
	require.NoError(t, err)

	_, err = config.ResolvePeripherals()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown peripheral type")
}

func TestPeripheralConfig_CustomDescription(t *testing.T) {
	yaml := `
version: 1
memory:
  total: 65536

peripherals:
  - type: terminal
    params:
      name: debug_uart
      description: "Debug UART for console output"
      size: 512
`

	config, err := Parse([]byte(yaml))
	require.NoError(t, err)

	metadata, err := config.ResolvePeripherals()
	require.NoError(t, err)

	require.Len(t, metadata, 1)
	assert.Equal(t, "Debug UART for console output", metadata[0].Description)
}

func TestMultiplePeripherals_DifferentTypes(t *testing.T) {
	yaml := `
version: 1
memory:
  total: 131072

interrupts:
  vectors: 64

peripherals:
  - type: terminal
    params:
      name: uart0
      size: 256
      interrupt_vector: 16

  - type: terminal
    params:
      name: uart1
      size: 256
      interrupt_vector: 17

  - type: terminal
    params:
      name: debug
      size: 128
      interrupt_vector: 18
`

	config, err := Parse([]byte(yaml))
	require.NoError(t, err)

	descriptor, err := config.Setup()
	require.NoError(t, err)

	assert.NotNil(t, descriptor)
	require.Len(t, descriptor.Peripherals, 3)
	// Each peripheral should have a unique base address
	addresses := make(map[uint32]bool)
	for _, p := range descriptor.Peripherals {
		addr := p.Metadata().BaseAddress
		assert.False(t, addresses[addr], "duplicate address: 0x%X", addr)
		addresses[addr] = true
	}
}
