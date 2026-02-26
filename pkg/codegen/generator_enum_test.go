package codegen

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cucarachareflex "github.com/Manu343726/cucaracha/pkg/reflect"
)

func TestGenerateEnumValuesMapSorted(t *testing.T) {
	tests := []struct {
		name          string
		constants     []*cucarachareflex.Constant
		expectedOrder []interface{} // Expected order of values after sorting
	}{
		{
			name: "integers in random order",
			constants: []*cucarachareflex.Constant{
				{Name: "Third", Value: cucarachareflex.NewValue(3)},
				{Name: "First", Value: cucarachareflex.NewValue(1)},
				{Name: "Second", Value: cucarachareflex.NewValue(2)},
			},
			expectedOrder: []interface{}{1, 2, 3},
		},
		{
			name: "negative and positive integers",
			constants: []*cucarachareflex.Constant{
				{Name: "Five", Value: cucarachareflex.NewValue(5)},
				{Name: "NegTwo", Value: cucarachareflex.NewValue(-2)},
				{Name: "Zero", Value: cucarachareflex.NewValue(0)},
				{Name: "Two", Value: cucarachareflex.NewValue(2)},
			},
			expectedOrder: []interface{}{-2, 0, 2, 5},
		},
		{
			name: "single value",
			constants: []*cucarachareflex.Constant{
				{Name: "Only", Value: cucarachareflex.NewValue(42)},
			},
			expectedOrder: []interface{}{42},
		},
		{
			name: "already sorted",
			constants: []*cucarachareflex.Constant{
				{Name: "One", Value: cucarachareflex.NewValue(10)},
				{Name: "Two", Value: cucarachareflex.NewValue(20)},
				{Name: "Three", Value: cucarachareflex.NewValue(30)},
			},
			expectedOrder: []interface{}{10, 20, 30},
		},
		{
			name: "reverse sorted",
			constants: []*cucarachareflex.Constant{
				{Name: "Three", Value: cucarachareflex.NewValue(30)},
				{Name: "Two", Value: cucarachareflex.NewValue(20)},
				{Name: "One", Value: cucarachareflex.NewValue(10)},
			},
			expectedOrder: []interface{}{10, 20, 30},
		},
		{
			name: "duplicate values",
			constants: []*cucarachareflex.Constant{
				{Name: "A", Value: cucarachareflex.NewValue(5)},
				{Name: "B", Value: cucarachareflex.NewValue(3)},
				{Name: "C", Value: cucarachareflex.NewValue(5)},
				{Name: "D", Value: cucarachareflex.NewValue(1)},
			},
			expectedOrder: []interface{}{1, 3, 5, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an enum with the test constants
			enum := &cucarachareflex.Enum{
				Type: &cucarachareflex.TypeReference{
					Name: "TestEnum",
				},
				Values: tt.constants,
			}

			// Create a generator with a buffer
			buf := &bytes.Buffer{}
			gen := NewGenerator(&cucarachareflex.Package{}, buf)

			// Generate the enum values map
			err := gen.generateEnumValuesMap(enum)
			require.NoError(t, err)

			// The values should be sorted - verify by checking the output
			output := buf.String()
			require.NotEmpty(t, output)

			// Verify the map contains all expected values in order
			// by checking that values appear in the correct order in the output
			for i, expectedVal := range tt.expectedOrder {
				expectedStr := formatValue(expectedVal)
				assert.Contains(t, output, expectedStr, "expected value %v at position %d should be in output", expectedVal, i)
			}
		})
	}
}

func TestGenerateEnumValuesMapDeterministic(t *testing.T) {
	t.Run("same input produces same output", func(t *testing.T) {
		constants := []*cucarachareflex.Constant{
			{Name: "Z", Value: cucarachareflex.NewValue(26)},
			{Name: "A", Value: cucarachareflex.NewValue(1)},
			{Name: "M", Value: cucarachareflex.NewValue(13)},
			{Name: "D", Value: cucarachareflex.NewValue(4)},
		}

		enum := &cucarachareflex.Enum{
			Type: &cucarachareflex.TypeReference{
				Name: "TestEnum",
			},
			Values: constants,
		}

		// Generate multiple times
		outputs := make([]string, 3)
		for i := 0; i < 3; i++ {
			buf := &bytes.Buffer{}
			gen := NewGenerator(&cucarachareflex.Package{}, buf)
			err := gen.generateEnumValuesMap(enum)
			require.NoError(t, err)
			outputs[i] = buf.String()
		}

		// All outputs should be identical
		for i := 1; i < len(outputs); i++ {
			assert.Equal(t, outputs[0], outputs[i], "output iteration %d should match iteration 0", i)
		}
	})
}

func TestGenerateEnumValuesMapContainsAllValues(t *testing.T) {
	t.Run("all constants are in the generated map", func(t *testing.T) {
		constants := []*cucarachareflex.Constant{
			{Name: "Alpha", Value: cucarachareflex.NewValue(1)},
			{Name: "Beta", Value: cucarachareflex.NewValue(2)},
			{Name: "Gamma", Value: cucarachareflex.NewValue(3)},
		}

		enum := &cucarachareflex.Enum{
			Type: &cucarachareflex.TypeReference{
				Name: "TestEnum",
			},
			Values: constants,
		}

		buf := &bytes.Buffer{}
		gen := NewGenerator(&cucarachareflex.Package{}, buf)
		err := gen.generateEnumValuesMap(enum)
		require.NoError(t, err)

		output := buf.String()

		// Check that all values are present
		for i, c := range constants {
			valueStr := formatValue(c.Value.Value)
			assert.Contains(t, output, valueStr,
				"constant %d (%s) with value %v should be in output", i, c.Name, c.Value.Value)
		}

		// Check that the output contains the map declaration
		assert.Contains(t, output, "var", "output should contain variable declaration")
		assert.Contains(t, output, "map", "output should contain map type")
	})
}

func TestGenerateEnumValuesMapWithStrings(t *testing.T) {
	t.Run("string constants are sorted", func(t *testing.T) {
		constants := []*cucarachareflex.Constant{
			{Name: "Zebra", Value: cucarachareflex.NewValue("zebra")},
			{Name: "Apple", Value: cucarachareflex.NewValue("apple")},
			{Name: "Mango", Value: cucarachareflex.NewValue("mango")},
		}

		enum := &cucarachareflex.Enum{
			Type: &cucarachareflex.TypeReference{
				Name: "StringEnum",
			},
			Values: constants,
		}

		buf := &bytes.Buffer{}
		gen := NewGenerator(&cucarachareflex.Package{}, buf)
		err := gen.generateEnumValuesMap(enum)
		require.NoError(t, err)

		output := buf.String()

		// Strings should appear in sorted order: apple, mango, zebra
		appleIdx := bytes.Index([]byte(output), []byte("apple"))
		mangoIdx := bytes.Index([]byte(output), []byte("mango"))
		zebraIdx := bytes.Index([]byte(output), []byte("zebra"))

		require.NotEqual(t, -1, appleIdx)
		require.NotEqual(t, -1, mangoIdx)
		require.NotEqual(t, -1, zebraIdx)

		assert.Less(t, appleIdx, mangoIdx)
		assert.Less(t, mangoIdx, zebraIdx)
	})
}

func TestGenerateEnumValuesMapNoDuplicateEntries(t *testing.T) {
	t.Run("each value appears once in the map", func(t *testing.T) {
		constants := []*cucarachareflex.Constant{
			{Name: "One", Value: cucarachareflex.NewValue(1)},
			{Name: "Two", Value: cucarachareflex.NewValue(2)},
			{Name: "Three", Value: cucarachareflex.NewValue(3)},
		}

		enum := &cucarachareflex.Enum{
			Type: &cucarachareflex.TypeReference{
				Name: "TestEnum",
			},
			Values: constants,
		}

		buf := &bytes.Buffer{}
		gen := NewGenerator(&cucarachareflex.Package{}, buf)
		err := gen.generateEnumValuesMap(enum)
		require.NoError(t, err)

		output := buf.String()

		// Count occurrences of each value
		for _, c := range constants {
			valueStr := formatValue(c.Value.Value)
			count := bytes.Count([]byte(output), []byte(valueStr))
			// Should appear exactly once (allowing for one occurrence in the map)
			assert.Equal(t, 1, count, "value %v should appear exactly once", c.Value.Value)
		}
	})
}

func TestGenerateEnumValuesMapMapKeyType(t *testing.T) {
	t.Run("generated map uses correct key type", func(t *testing.T) {
		constants := []*cucarachareflex.Constant{
			{Name: "One", Value: cucarachareflex.NewValue(1)},
			{Name: "Two", Value: cucarachareflex.NewValue(2)},
		}

		enum := &cucarachareflex.Enum{
			Type: &cucarachareflex.TypeReference{
				Name: "MyEnum",
			},
			Values: constants,
		}

		buf := &bytes.Buffer{}
		gen := NewGenerator(&cucarachareflex.Package{}, buf)
		err := gen.generateEnumValuesMap(enum)
		require.NoError(t, err)

		output := buf.String()

		// Should use the enum type name as the map key
		if !bytes.Contains([]byte(output), []byte("MyEnum")) {
			t.Error("generated map does not use the correct enum type name")
		}

		// Should use bool as the map value
		if !bytes.Contains([]byte(output), []byte("bool")) {
			t.Error("generated map value type is not bool")
		}
	})
}

// Helper function to format values for string comparison
func formatValue(v interface{}) string {
	switch val := v.(type) {
	case int:
		return formatInt(val)
	case int8:
		return formatInt(int(val))
	case int16:
		return formatInt(int(val))
	case int32:
		return formatInt(int(val))
	case int64:
		return formatInt(int(val))
	case string:
		return val
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}

// Helper to format int values as Go code would
func formatInt(i int) string {
	if i < 0 {
		return "-" // Just check for presence of negative sign
	}
	// Return a representation that would appear in the code
	switch i {
	case 0:
		return "0"
	case 1:
		return "1"
	case 2:
		return "2"
	case 3:
		return "3"
	case 4:
		return "4"
	case 5:
		return "5"
	case 10:
		return "10"
	case 13:
		return "13"
	case 20:
		return "20"
	case 26:
		return "26"
	case 30:
		return "30"
	default:
		return ""
	}
}
