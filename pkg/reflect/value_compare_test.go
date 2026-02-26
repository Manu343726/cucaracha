package reflect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValueCompareNilHandling(t *testing.T) {
	tests := []struct {
		name     string
		a        *Value
		b        *Value
		expected int
	}{
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: 0,
		},
		{
			name:     "a nil, b non-nil",
			a:        nil,
			b:        NewValue(42),
			expected: -1,
		},
		{
			name:     "a non-nil, b nil",
			a:        NewValue(42),
			b:        nil,
			expected: 1,
		},
		{
			name:     "both have nil Value",
			a:        &Value{Value: nil},
			b:        &Value{Value: nil},
			expected: 0,
		},
		{
			name:     "a has nil Value, b non-nil",
			a:        &Value{Value: nil},
			b:        NewValue(42),
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.a.Compare(tt.b)
			assert.Equal(t, tt.expected, result, "Compare(%v, %v) should equal %d", tt.a, tt.b, tt.expected)
		})
	}
}

func TestValueCompareIntegers(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected int
	}{
		{"int: 10 < 20", int(10), int(20), -1},
		{"int: 20 > 10", int(20), int(10), 1},
		{"int: 10 == 10", int(10), int(10), 0},
		{"int: -5 < 5", int(-5), int(5), -1},
		{"int8: 5 < 10", int8(5), int8(10), -1},
		{"int16: 100 > 50", int16(100), int16(50), 1},
		{"int32: 0 == 0", int32(0), int32(0), 0},
		{"int64: -100 < 0", int64(-100), int64(0), -1},
		{"uint: 5 < 10", uint(5), uint(10), -1},
		{"uint8: 255 > 0", uint8(255), uint8(0), 1},
		{"uint32: 100 == 100", uint32(100), uint32(100), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			va := NewValue(tt.a)
			vb := NewValue(tt.b)
			result := va.Compare(vb)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValueCompareFloats(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected int
	}{
		{"float32: 1.5 < 2.5", float32(1.5), float32(2.5), -1},
		{"float32: 3.14 > 2.71", float32(3.14), float32(2.71), 1},
		{"float64: 0.0 == 0.0", float64(0.0), float64(0.0), 0},
		{"float64: -1.5 < 1.5", float64(-1.5), float64(1.5), -1},
		{"float64: 99.99 > 99.98", float64(99.99), float64(99.98), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			va := NewValue(tt.a)
			vb := NewValue(tt.b)
			result := va.Compare(vb)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValueCompareStrings(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected int
	}{
		{"apple < banana", "apple", "banana", -1},
		{"zebra > apple", "zebra", "apple", 1},
		{"hello == hello", "hello", "hello", 0},
		{"a < b", "a", "b", -1},
		{"empty == empty", "", "", 0},
		{"empty < zero", "", "0", -1},
		{"123 < 456", "123", "456", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			va := NewValue(tt.a)
			vb := NewValue(tt.b)
			result := va.Compare(vb)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValueCompareBooleans(t *testing.T) {
	tests := []struct {
		name     string
		a        bool
		b        bool
		expected int
	}{
		{"false < true", false, true, -1},
		{"true > false", true, false, 1},
		{"false == false", false, false, 0},
		{"true == true", true, true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			va := NewValue(tt.a)
			vb := NewValue(tt.b)
			result := va.Compare(vb)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValueCompareSlices(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected int
	}{
		{
			name:     "[]int with equal elements",
			a:        []int{1, 2, 3},
			b:        []int{1, 2, 3},
			expected: 0,
		},
		{
			name:     "[]int first element smaller",
			a:        []int{1, 5, 9},
			b:        []int{2, 5, 9},
			expected: -1,
		},
		{
			name:     "[]int first element larger",
			a:        []int{3, 2, 1},
			b:        []int{2, 2, 1},
			expected: 1,
		},
		{
			name:     "[]int shorter slice",
			a:        []int{1, 2},
			b:        []int{1, 2, 3},
			expected: -1,
		},
		{
			name:     "[]int longer slice",
			a:        []int{1, 2, 3, 4},
			b:        []int{1, 2, 3},
			expected: 1,
		},
		{
			name:     "[]string comparison",
			a:        []string{"a", "b"},
			b:        []string{"a", "c"},
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			va := NewValue(tt.a)
			vb := NewValue(tt.b)
			result := va.Compare(vb)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValueCompareArrays(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected int
	}{
		{
			name:     "[3]int with equal elements",
			a:        [3]int{1, 2, 3},
			b:        [3]int{1, 2, 3},
			expected: 0,
		},
		{
			name:     "[3]int first element smaller",
			a:        [3]int{1, 5, 9},
			b:        [3]int{2, 5, 9},
			expected: -1,
		},
		{
			name:     "[3]int first element larger",
			a:        [3]int{3, 2, 1},
			b:        [3]int{2, 2, 1},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			va := NewValue(tt.a)
			vb := NewValue(tt.b)
			result := va.Compare(vb)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValueCompareDifferentKinds(t *testing.T) {
	tests := []struct {
		name string
		a    interface{}
		b    interface{}
	}{
		{"int vs string", 42, "hello"},
		{"bool vs int", true, 1},
		{"float vs string", 3.14, "3.14"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			va := NewValue(tt.a)
			vb := NewValue(tt.b)
			// Should return a comparison value (not panic)
			result := va.Compare(vb)
			// Just verify it doesn't panic and returns an int
			_ = result
		})
	}
}

func TestValueCompareSorting(t *testing.T) {
	t.Run("sorting integers", func(t *testing.T) {
		values := []*Value{
			NewValue(30),
			NewValue(10),
			NewValue(20),
			NewValue(5),
		}

		// Bubble sort using Compare
		for i := 0; i < len(values); i++ {
			for j := i + 1; j < len(values); j++ {
				if values[i].Compare(values[j]) > 0 {
					values[i], values[j] = values[j], values[i]
				}
			}
		}

		// Verify sorted order
		expected := []interface{}{5, 10, 20, 30}
		for i, v := range values {
			assert.Equal(t, expected[i], v.Value)
		}
	})

	t.Run("sorting strings", func(t *testing.T) {
		values := []*Value{
			NewValue("charlie"),
			NewValue("alpha"),
			NewValue("bravo"),
		}

		// Bubble sort using Compare
		for i := 0; i < len(values); i++ {
			for j := i + 1; j < len(values); j++ {
				if values[i].Compare(values[j]) > 0 {
					values[i], values[j] = values[j], values[i]
				}
			}
		}

		// Verify sorted order
		expected := []string{"alpha", "bravo", "charlie"}
		for i, v := range values {
			assert.Equal(t, expected[i], v.Value)
		}
	})
}

func TestValueCompareDeterminism(t *testing.T) {
	t.Run("consistent results", func(t *testing.T) {
		va := NewValue(42)
		vb := NewValue(100)

		// Compare multiple times - should get same result
		for i := 0; i < 10; i++ {
			result1 := va.Compare(vb)
			result2 := va.Compare(vb)
			assert.Equal(t, result1, result2, "iteration %d: Compare should return consistent results", i)
		}
	})

	t.Run("reflexive property: a.Compare(a) == 0", func(t *testing.T) {
		testValues := []interface{}{
			int(42),
			float64(3.14),
			"hello",
			true,
			[]int{1, 2, 3},
		}

		for _, v := range testValues {
			val := NewValue(v)
			result := val.Compare(val)
			assert.Equal(t, 0, result, "val.Compare(val) should be 0 for %v", v)
		}
	})

	t.Run("antisymmetric property: a.Compare(b) == -b.Compare(a)", func(t *testing.T) {
		pairs := []struct {
			a interface{}
			b interface{}
		}{
			{int(10), int(20)},
			{"apple", "banana"},
			{float64(1.5), float64(2.5)},
			{[]int{1, 2}, []int{1, 3}},
		}

		for _, pair := range pairs {
			va := NewValue(pair.a)
			vb := NewValue(pair.b)

			cmpAB := va.Compare(vb)
			cmpBA := vb.Compare(va)

			assert.Equal(t, -cmpBA, cmpAB, "Compare(%v, %v) and Compare(%v, %v) should be antisymmetric",
				pair.a, pair.b, pair.b, pair.a)
		}
	})

	t.Run("transitive property: if a < b and b < c then a < c", func(t *testing.T) {
		values := []interface{}{int(10), int(20), int(30)}
		va := NewValue(values[0])
		vb := NewValue(values[1])
		vc := NewValue(values[2])

		if va.Compare(vb) >= 0 {
			t.Fatalf("va.Compare(vb) should be < 0")
		}
		if vb.Compare(vc) >= 0 {
			t.Fatalf("vb.Compare(vc) should be < 0")
		}

		assert.Less(t, va.Compare(vc), 0, "va.Compare(vc) should be < 0 (transitivity)")
	})
}
