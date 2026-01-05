package component

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// DescriptorBuilder Tests
// =============================================================================

func TestDescriptorBuilder(t *testing.T) {
	t.Run("builds complete descriptor", func(t *testing.T) {
		desc := NewDescriptor("AND").
			DisplayName("AND Gate").
			Description("2-input AND gate").
			Category("gates").
			Version("1.0.0").
			Input("A", 1, "First input").
			Input("B", 1, "Second input").
			Output("Y", 1, "Output = A AND B").
			Param("width", "int", 1, "Bit width").
			Factory(func(name string, params map[string]interface{}) (Component, error) {
				return NewBaseComponent(name, "logic"), nil
			}).
			Build()

		assert.Equal(t, "AND", desc.Name)
		assert.Equal(t, "AND Gate", desc.DisplayName)
		assert.Equal(t, "2-input AND gate", desc.Description)
		assert.Equal(t, "gates", desc.Category)
		assert.Equal(t, "1.0.0", desc.Version)
		assert.Len(t, desc.Inputs, 2)
		assert.Len(t, desc.Outputs, 1)
		assert.Len(t, desc.Parameters, 1)
		assert.NotNil(t, desc.Factory)
	})
}

// =============================================================================
// Registry Tests
// =============================================================================

func TestRegistry(t *testing.T) {
	t.Run("Register and Get", func(t *testing.T) {
		reg := NewRegistry()
		desc := NewDescriptor("TEST").Build()

		require.NoError(t, reg.Register(desc))

		got, err := reg.Get("TEST")
		require.NoError(t, err)
		assert.Equal(t, "TEST", got.Name)
	})

	t.Run("Get returns error for unknown", func(t *testing.T) {
		reg := NewRegistry()
		_, err := reg.Get("UNKNOWN")
		assert.Error(t, err)
	})

	t.Run("List returns all names", func(t *testing.T) {
		reg := NewRegistry()
		reg.Register(NewDescriptor("A").Build())
		reg.Register(NewDescriptor("B").Build())
		reg.Register(NewDescriptor("C").Build())

		list := reg.List()
		assert.Len(t, list, 3)
		assert.Contains(t, list, "A")
		assert.Contains(t, list, "B")
		assert.Contains(t, list, "C")
	})

	t.Run("ListByCategory filters by category", func(t *testing.T) {
		reg := NewRegistry()
		reg.Register(NewDescriptor("AND").Category("logic").Build())
		reg.Register(NewDescriptor("REG").Category("memory").Build())
		reg.Register(NewDescriptor("OR").Category("logic").Build())

		logicList := reg.ListByCategory("logic")
		assert.Len(t, logicList, 2)
		assert.Contains(t, logicList, "AND")
		assert.Contains(t, logicList, "OR")
	})

	t.Run("Categories returns all categories", func(t *testing.T) {
		reg := NewRegistry()
		reg.Register(NewDescriptor("AND").Category("logic").Build())
		reg.Register(NewDescriptor("REG").Category("memory").Build())

		cats := reg.Categories()
		assert.Len(t, cats, 2)
		assert.Contains(t, cats, "logic")
		assert.Contains(t, cats, "memory")
	})

	t.Run("Unregister removes descriptor", func(t *testing.T) {
		reg := NewRegistry()
		reg.Register(NewDescriptor("TEST").Build())

		require.NoError(t, reg.Unregister("TEST"))
		_, err := reg.Get("TEST")
		assert.Error(t, err)

		assert.Error(t, reg.Unregister("UNKNOWN"))
	})
}

// =============================================================================
// Registry Create Tests
// =============================================================================

func TestRegistryCreate(t *testing.T) {
	t.Run("Create instantiates component", func(t *testing.T) {
		reg := NewRegistry()
		reg.Register(NewDescriptor("TEST").
			Factory(func(name string, params map[string]interface{}) (Component, error) {
				comp := NewBaseComponent(name, "test")
				return comp, nil
			}).
			Build())

		comp, err := reg.Create("TEST", "my_test", nil)
		require.NoError(t, err)
		assert.Equal(t, "my_test", comp.Name())
	})

	t.Run("Create returns error for unknown", func(t *testing.T) {
		reg := NewRegistry()
		_, err := reg.Create("UNKNOWN", "inst", nil)
		assert.Error(t, err)
	})

	t.Run("Create returns error when no factory", func(t *testing.T) {
		reg := NewRegistry()
		reg.Register(NewDescriptor("TEST").Build())

		_, err := reg.Create("TEST", "inst", nil)
		assert.Error(t, err)
	})
}

// =============================================================================
// Registry Search Tests
// =============================================================================

func TestRegistrySearch(t *testing.T) {
	setupRegistry := func() *Registry {
		reg := NewRegistry()
		reg.Register(NewDescriptor("AND2").
			Description("2-input AND gate").
			Category("gates").
			Build())
		reg.Register(NewDescriptor("AND4").
			Description("4-input AND gate").
			Category("gates").
			Build())
		reg.Register(NewDescriptor("DFLIPFLOP").
			Description("D flip-flop with async reset").
			Category("flip-flops").
			Build())
		return reg
	}

	t.Run("Search by name substring", func(t *testing.T) {
		reg := setupRegistry()
		results := reg.Search(SearchQuery{NameContains: "AND"})
		assert.Len(t, results, 2)
	})

	t.Run("Search by description", func(t *testing.T) {
		reg := setupRegistry()
		results := reg.Search(SearchQuery{NameContains: "flip-flop"})
		assert.Len(t, results, 1)
		assert.Equal(t, "DFLIPFLOP", results[0].Name)
	})

	t.Run("Search case insensitive", func(t *testing.T) {
		reg := setupRegistry()
		results := reg.Search(SearchQuery{NameContains: "and"})
		assert.Len(t, results, 2)
	})

	t.Run("Search by category", func(t *testing.T) {
		reg := setupRegistry()
		results := reg.Search(SearchQuery{Category: "gates"})
		assert.Len(t, results, 2)

		for _, r := range results {
			assert.Equal(t, "gates", r.Category)
		}
	})

	t.Run("Search empty returns empty", func(t *testing.T) {
		reg := setupRegistry()
		results := reg.Search(SearchQuery{NameContains: "xyz"})
		assert.Empty(t, results)
	})
}

// =============================================================================
// Required Parameters Tests
// =============================================================================

func TestRequiredParams(t *testing.T) {
	t.Run("RequiredParam sets required flag", func(t *testing.T) {
		desc := NewDescriptor("TEST").
			RequiredParam("size", "int", "Required size").
			Param("name", "string", "default", "Optional name").
			Build()

		var sizeParam *ParameterDescriptor
		for i := range desc.Parameters {
			if desc.Parameters[i].Name == "size" {
				sizeParam = &desc.Parameters[i]
				break
			}
		}

		require.NotNil(t, sizeParam)
		assert.True(t, sizeParam.Required)
	})

	t.Run("Create fails without required param", func(t *testing.T) {
		reg := NewRegistry()
		reg.Register(NewDescriptor("TEST").
			RequiredParam("size", "int", "Required size").
			Factory(func(name string, params map[string]interface{}) (Component, error) {
				return NewBaseComponent(name, "test"), nil
			}).
			Build())

		_, err := reg.Create("TEST", "inst", nil)
		assert.Error(t, err, "should fail without required param")
	})
}
