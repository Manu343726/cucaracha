package ui

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterJSON(t *testing.T) {
	reg := &Register{
		Name:     "r0",
		Encoding: 0,
		Value:    0x1234,
	}

	data, err := json.Marshal(reg)
	assert.NoError(t, err)

	var unmarshaled Register
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, reg.Name, unmarshaled.Name)
	assert.Equal(t, reg.Value, unmarshaled.Value)
}

func TestFlagStateJSON(t *testing.T) {
	flags := &FlagState{
		N: true,
		Z: false,
		C: true,
		V: false,
	}

	data, err := json.Marshal(flags)
	assert.NoError(t, err)

	var unmarshaled FlagState
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, flags.N, unmarshaled.N)
	assert.Equal(t, flags.Z, unmarshaled.Z)
	assert.Equal(t, flags.C, unmarshaled.C)
	assert.Equal(t, flags.V, unmarshaled.V)
}

func TestRegistersResultJSON(t *testing.T) {
	result := &RegistersResult{
		Registers: map[string]*Register{
			"r0": {Name: "r0", Value: 0x1000},
			"r1": {Name: "r1", Value: 0x2000},
		},
		Flags: &FlagState{N: true, Z: false, C: false, V: true},
	}

	data, err := json.Marshal(result)
	assert.NoError(t, err)

	var unmarshaled RegistersResult
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Len(t, unmarshaled.Registers, len(result.Registers))
}
