// Package system provides YAML-based system configuration for the Cucaracha CPU.
//
// The system configuration describes the hardware setup including:
//   - Memory layout (code, data, stack, heap sizes)
//   - Peripheral devices (UART, GPIO, timers, etc.)
//   - Interrupt vector configuration
//
// Example YAML configuration:
//
//	version: 1
//	name: "Basic System"
//
//	memory:
//	  total: 65536        # 64KB total memory
//	  code_size: 16384    # 16KB for code
//	  data_size: 4096     # 4KB for data
//	  stack_size: 8192    # 8KB stack
//	  heap_size: 0        # Use remaining space
//
//	interrupts:
//	  vectors: 32         # Number of interrupt vectors
//
//	peripherals:
//	  - name: uart0
//	    type: uart
//	    size: 256
//	    interrupt: 16
//
//	  - name: gpio
//	    type: gpio
//	    size: 128
//	    interrupt: 17
package system

import (
	"fmt"
	"os"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/hw/peripheral"
	"github.com/Manu343726/cucaracha/pkg/hw/peripherals"
	"gopkg.in/yaml.v3"
)

// The current configuration format version.
const ConfigVersion = 1

// Represents a complete system configuration.
type SystemConfig struct {
	// Version of the configuration format.
	Version int `yaml:"version"`

	// An optional human-readable name for this configuration.
	Name string `yaml:"name,omitempty"`

	// Provides additional details about the configuration.
	Description string `yaml:"description,omitempty"`

	// Configures the memory requirements.
	Memory MemoryConfig `yaml:"memory"`

	// Configures the interrupt system.
	Interrupts InterruptConfig `yaml:"interrupts,omitempty"`

	// List of peripheral devices available in the system.
	Peripherals []PeripheralConfig `yaml:"peripherals,omitempty"`
}

// Describes the memory requirements of the system.
type MemoryConfig struct {
	// Total memory size in bytes.
	Total uint32 `yaml:"total,omitempty"`

	// Space for program code in bytes.
	//
	// If not set, defaults to Instructions * 4 bytes.
	CodeSize uint32 `yaml:"code_size,omitempty"`

	// Number of instructions that fit in the code region.
	//
	// If not set, defaults to CodeSize / 4.
	Instructions uint32 `yaml:"instructions,omitempty"`

	// Vector table size in bytes.
	//
	// If not set, defaults to NumInterruptVectors * VectorSize (See InterruptConfig).
	VectorTableSize uint32 `yaml:"vector_table_size,omitempty"`

	// Space for static/global data in bytes.
	DataSize uint32 `yaml:"data_size,omitempty"`

	// Space for dynamic memory allocations in bytes.
	HeapSize uint32 `yaml:"heap_size,omitempty"`

	// Space for the call stack in bytes.
	StackSize uint32 `yaml:"stack_size,omitempty"`

	// Space for peripheral devices MMIO in bytes.
	PeripheralSize uint32 `yaml:"peripheral_size,omitempty"`
}

// Describes the interrupt system.
type InterruptConfig struct {
	// Number of interrupt vectors.
	Vectors uint32 `yaml:"vectors,omitempty"`
	// Size of each interrupt vector in bytes.
	VectorSize uint32 `yaml:"vector_size,omitempty"`
}

// Describes a single peripheral device.
type PeripheralConfig struct {
	// The peripheral type (uart, gpio, timer, etc.).
	Type string `yaml:"type"`

	// Parameters for creating the peripheral.
	Params peripheral.PeripheralParams `yaml:"params"`
}

// Loads a system configuration from a YAML file.
func LoadFile(path string) (*SystemConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	return Parse(data)
}

// Parse parses a system configuration from YAML data.
func Parse(data []byte) (*SystemConfig, error) {
	var config SystemConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if config.Version == 0 {
		config.Version = ConfigVersion
	}
	if config.Version > ConfigVersion {
		return nil, fmt.Errorf("unsupported config version %d (max supported: %d)", config.Version, ConfigVersion)
	}

	return &config, nil
}

// Finds peripheral descriptors and builds metadata for each configured peripheral.
func (c *SystemConfig) ResolvePeripherals() ([]peripheral.Metadata, error) {
	metadata := make([]peripheral.Metadata, 0, len(c.Peripherals))

	for _, p := range c.Peripherals {
		descriptor, err := peripherals.Descriptor.GetByType(peripheral.Type(p.Type))
		if err != nil {
			return nil, fmt.Errorf("unknown peripheral type %s: %w", p.Type, err)
		}

		description := descriptor.Description
		if p.Params.Description != "" {
			description = p.Params.Description
		}

		size := p.Params.Size
		if size == 0 {
			size = descriptor.DefaultSize
		}

		interruptVector := p.Params.InterruptVector
		if interruptVector == 0 && descriptor.HasInterrupt {
			interruptVector = descriptor.DefaultInterruptVector
		}

		metadata = append(metadata, peripheral.Metadata{
			Name:            p.Params.Name,
			Description:     description,
			BaseAddress:     p.Params.BaseAddress,
			Size:            size,
			InterruptVector: interruptVector,
			Descriptor:      descriptor,
		})
	}

	return metadata, nil
}

// Returns the memory allocation requirements for this configuration.
func (c *SystemConfig) memoryRequirements() (*MemoryRequirements, error) {
	peripherals, err := c.ResolvePeripherals()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peripherals: %w", err)
	}

	instructions := c.Memory.Instructions

	if instructions == 0 {
		if c.Memory.CodeSize == 0 {
			return nil, fmt.Errorf("either memory.instructions or memory.code_size must be set")
		}
		instructions = c.Memory.CodeSize / 4
	}

	vectorSize := c.Interrupts.VectorSize
	if vectorSize == 0 {
		vectorSize = 4 // Default vector size
	}

	numVectors := c.Interrupts.Vectors
	if numVectors == 0 {
		if c.Memory.VectorTableSize == 0 {
			return nil, fmt.Errorf("either interrupts.vectors or memory.vector_table_size must be set")
		}
		numVectors = c.Memory.VectorTableSize / vectorSize
	}

	req := MemoryRequirements{
		TotalSize:           c.Memory.Total,
		CodeInstructions:    instructions,
		DataSize:            c.Memory.DataSize,
		HeapSize:            c.Memory.HeapSize,
		StackSize:           c.Memory.StackSize,
		NumInterruptVectors: c.Interrupts.Vectors,
		VectorSize:          c.Interrupts.VectorSize,
		MinPeripheralRegion: c.Memory.PeripheralSize,
		Peripherals:         peripherals,
	}

	return &req, nil
}

// Computes memory layout and peripheral allocation.
func (c *SystemConfig) Setup() (*SystemDescriptor, error) {
	memoryRequirements, err := c.memoryRequirements()
	if err != nil {
		return nil, err
	}

	allocation, err := Allocate(*memoryRequirements)

	if err != nil {
		return nil, fmt.Errorf("memory allocation failed: %w", err)
	}

	peripherals := make([]peripheral.Peripheral, 0, len(c.Peripherals))

	for i, metadata := range memoryRequirements.Peripherals {
		params := c.Peripherals[i].Params
		params.BaseAddress = allocation.PeripheralAddresses[i]
		peripheral, err := metadata.Descriptor.Factory(params)
		if err != nil {
			return nil, fmt.Errorf("failed to create peripheral %s: %w", metadata.Name, err)
		}
		peripherals = append(peripherals, peripheral)
	}

	return &SystemDescriptor{
		MemoryLayout: allocation.Layout,
		Peripherals:  peripherals,
	}, nil
}

// Returns a human-readable summary of the configuration.
func (c *SystemConfig) String() string {
	// Dump as YAML
	data, _ := yaml.Marshal(c)
	return strings.TrimSpace(string(data))
}
