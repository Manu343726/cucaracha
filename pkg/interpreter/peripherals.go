package interpreter

import (
	"fmt"

	"github.com/Manu343726/cucaracha/pkg/hw/peripheral"
)

// Represents the collection of peripherals used by the interpreter CPU
type Peripherals struct {
	byName            map[string]peripheral.Peripheral
	byInterruptVector map[uint8]peripheral.InterruptSource
}

// Creates a new Peripherals collection
func NewPeripherals(peripherals []peripheral.Peripheral) *Peripherals {
	byName := make(map[string]peripheral.Peripheral)
	for _, p := range peripherals {
		byName[p.Metadata().Name] = p
	}

	byInterruptVector := make(map[uint8]peripheral.InterruptSource)
	for _, p := range peripherals {
		if is, ok := p.(peripheral.InterruptSource); ok {
			byInterruptVector[is.InterruptVector()] = is
		}
	}

	return &Peripherals{
		byName:            byName,
		byInterruptVector: byInterruptVector,
	}
}

// Returns all peripherals in the collection
func (p *Peripherals) GetAll() []peripheral.Peripheral {
	result := make([]peripheral.Peripheral, 0, len(p.byName))
	for _, periph := range p.byName {
		result = append(result, periph)
	}
	return result
}

// Returns a peripheral by its name
func (p *Peripherals) GetByName(name string) (peripheral.Peripheral, error) {
	peripheral, exists := p.byName[name]
	if !exists {
		return nil, fmt.Errorf("peripheral %q not found", name)
	}
	return peripheral, nil
}

// Returns all registered interrupt sources
func (p *Peripherals) GetInterruptSources() []peripheral.InterruptSource {
	result := make([]peripheral.InterruptSource, 0, len(p.byInterruptVector))
	for _, source := range p.byInterruptVector {
		result = append(result, source)
	}
	return result
}

// Returns an interrupt source by its interrupt vector
func (p *Peripherals) GetInterruptSourceByVector(vector uint8) (peripheral.InterruptSource, error) {
	source, exists := p.byInterruptVector[vector]
	if !exists {
		return nil, fmt.Errorf("interrupt source with vector %d not found", vector)
	}
	return source, nil
}

// Clock advances the state of all peripherals by one clock cycle
func (p *Peripherals) Clock(env peripheral.Environment) error {
	for _, periph := range p.byName {
		if err := periph.Clock(env); err != nil {
			return err
		}
	}

	return nil
}

func (p *Peripherals) Reset() {
	for _, periph := range p.byName {
		periph.Reset()
	}
}
