package peripherals

import (
	"fmt"

	"github.com/Manu343726/cucaracha/pkg/hw/peripheral"
)

// Holds descriptors for all available peripherals.
type PeripheralsDescriptor struct {
	byType     map[peripheral.Type]*peripheral.Descriptor
	byTypeName map[string]*peripheral.Descriptor
}

// Returns the descriptor for the given peripheral type.
func (p *PeripheralsDescriptor) GetByType(t peripheral.Type) (*peripheral.Descriptor, error) {
	desc, ok := p.byType[t]
	if !ok {
		return nil, fmt.Errorf("peripheral type %q not found", t)
	}
	return desc, nil
}

// Returns the descriptor for the given peripheral type by name
func (p *PeripheralsDescriptor) GetByTypeName(typeName string) (*peripheral.Descriptor, error) {
	desc, ok := p.byTypeName[typeName]
	if !ok {
		return nil, fmt.Errorf("peripheral type name %q not found", typeName)
	}
	return desc, nil
}

// Registers known peripherals
func NewPeripheralsDescriptor(peripherals []*peripheral.Descriptor) *PeripheralsDescriptor {
	byType := make(map[peripheral.Type]*peripheral.Descriptor)
	for i := range peripherals {
		p := peripherals[i]
		byType[p.Type] = p
	}

	return &PeripheralsDescriptor{
		byType: byType,
	}
}

// Global descriptor for all available peripherals
var Descriptor = NewPeripheralsDescriptor([]*peripheral.Descriptor{
	TerminalDescriptor(),
})

// Convenience function to get peripheral descriptor by type.
//
// Used by peripheral implementations to return their own descriptor
func Peripheral(typeName peripheral.Type) *peripheral.Descriptor {
	desc, err := Descriptor.GetByType(typeName)
	if err != nil {
		panic(err)
	}

	return desc
}
