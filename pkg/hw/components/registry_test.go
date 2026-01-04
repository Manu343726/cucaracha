package components

import (
	"testing"

	"github.com/Manu343726/cucaracha/pkg/hw/component"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Registered Gates Tests
// =============================================================================

func TestRegisteredGates(t *testing.T) {
	gates := []string{"NOT", "AND", "OR", "XOR", "NAND", "NOR", "XNOR", "BUFFER"}

	for _, name := range gates {
		t.Run(name+" is registered", func(t *testing.T) {
			desc, err := Registry.Get(name)
			require.NoError(t, err, "%s should be registered", name)
			assert.Equal(t, name, desc.Name)
			assert.Equal(t, CategoryLogic, desc.Category)
		})
	}
}

// =============================================================================
// Create Gate From Registry Tests
// =============================================================================

func TestCreateGateFromRegistry(t *testing.T) {
	t.Run("Create AND gate from registry", func(t *testing.T) {
		comp, err := Registry.Create("AND", "my_and", nil)
		require.NoError(t, err)

		assert.Equal(t, "my_and", comp.Name())
		assert.Equal(t, "AND", comp.Type())
	})

	t.Run("Create OR gate with numInputs param", func(t *testing.T) {
		comp, err := Registry.Create("OR", "my_or", map[string]interface{}{
			"numInputs": 4,
		})
		require.NoError(t, err)

		assert.Len(t, comp.Inputs(), 4)
	})

	t.Run("Create NOT gate", func(t *testing.T) {
		comp, err := Registry.Create("NOT", "my_not", nil)
		require.NoError(t, err)

		assert.Len(t, comp.Inputs(), 1)
		assert.Len(t, comp.Outputs(), 1)
	})
}

// =============================================================================
// List Logic Gates Tests
// =============================================================================

func TestListLogicGates(t *testing.T) {
	logicGates := Registry.ListByCategory(CategoryLogic)
	assert.GreaterOrEqual(t, len(logicGates), 8, "should have at least 8 logic gates")
}

// =============================================================================
// Search Logic Gates Tests
// =============================================================================

func TestSearchLogicGates(t *testing.T) {
	t.Run("Search by name", func(t *testing.T) {
		results := Registry.Search(component.SearchQuery{NameContains: "AND"})
		assert.NotEmpty(t, results)

		found := false
		for _, r := range results {
			if r.Name == "AND" {
				found = true
				break
			}
		}
		assert.True(t, found, "should find AND gate")
	})

	t.Run("Search by category", func(t *testing.T) {
		results := Registry.Search(component.SearchQuery{Category: CategoryLogic})
		assert.GreaterOrEqual(t, len(results), 8, "all gates should be in logic category")
	})
}

// =============================================================================
// Gate Descriptor Details Tests
// =============================================================================

func TestGateDescriptorDetails(t *testing.T) {
	t.Run("AND gate descriptor", func(t *testing.T) {
		desc, err := Registry.Get("AND")
		require.NoError(t, err)

		assert.Equal(t, "AND", desc.Name)
		assert.Equal(t, CategoryLogic, desc.Category)
		assert.NotEmpty(t, desc.Description)
		assert.Len(t, desc.Outputs, 1)
	})

	t.Run("Gate has numInputs parameter", func(t *testing.T) {
		desc, err := Registry.Get("OR")
		require.NoError(t, err)

		var numInputsParam *component.ParameterDescriptor
		for i := range desc.Parameters {
			if desc.Parameters[i].Name == "numInputs" {
				numInputsParam = &desc.Parameters[i]
				break
			}
		}

		require.NotNil(t, numInputsParam, "should have numInputs parameter")
		assert.Equal(t, "int", numInputsParam.Type)
		assert.Equal(t, 2, numInputsParam.Default)
		assert.False(t, numInputsParam.Required)
	})
}
