package reflect

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for Type helper methods
func TestTypeIsComposite(t *testing.T) {
	tests := []struct {
		name     string
		typ      *Type
		expected bool
	}{
		{
			name:     "Struct is not composite (based on kinds)",
			typ:      Struct("User", nil),
			expected: false,
		},
		{
			name:     "Interface is not composite",
			typ:      Interface("Reader", nil),
			expected: false,
		},
		{
			name:     "Slice is composite",
			typ:      Slice(TypeInt),
			expected: true,
		},
		{
			name:     "Array is composite",
			typ:      Array(TypeInt, 5),
			expected: true,
		},
		{
			name:     "Map is composite",
			typ:      Map(TypeString, TypeInt),
			expected: true,
		},
		{
			name:     "Pointer is composite",
			typ:      Pointer(TypeInt),
			expected: true,
		},
		{
			name:     "Channel is composite",
			typ:      Chan(TypeInt, ChanBidirectional),
			expected: true,
		},
		{
			name:     "Bool basic type is not composite",
			typ:      TypeBool,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.typ.IsComposite()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTypeIsFunction(t *testing.T) {
	tests := []struct {
		name     string
		typ      *Type
		expected bool
	}{
		{
			name:     "Function type is function",
			typ:      &Type{Kind: TypeKindFunction},
			expected: true,
		},
		{
			name:     "Method type is not function (separate kind)",
			typ:      &Type{Kind: TypeKindMethod},
			expected: false,
		},
		{
			name:     "Struct is not function",
			typ:      Struct("User", nil),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.typ.IsFunction()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTypeIsStruct(t *testing.T) {
	tests := []struct {
		name     string
		typ      *Type
		expected bool
	}{
		{
			name:     "Struct type is struct",
			typ:      Struct("User", nil),
			expected: true,
		},
		{
			name:     "Interface is not struct",
			typ:      Interface("Reader", nil),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.typ.IsStruct()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTypeIsInterface(t *testing.T) {
	tests := []struct {
		name     string
		typ      *Type
		expected bool
	}{
		{
			name:     "Interface type is interface",
			typ:      Interface("Reader", nil),
			expected: true,
		},
		{
			name:     "Struct is not interface",
			typ:      Struct("User", nil),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.typ.IsInterface()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTypeIsAlias(t *testing.T) {
	tests := []struct {
		name     string
		typ      *Type
		expected bool
	}{
		{
			name:     "Alias type is alias",
			typ:      Alias("MyInt", TypeInt),
			expected: true,
		},
		{
			name:     "Typedef is not alias",
			typ:      Typedef("MyInt", TypeInt),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.typ.IsAlias()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTypeIsTypedef(t *testing.T) {
	tests := []struct {
		name     string
		typ      *Type
		expected bool
	}{
		{
			name:     "Typedef type is typedef",
			typ:      Typedef("MyInt", TypeInt),
			expected: true,
		},
		{
			name:     "Alias is not typedef",
			typ:      Alias("MyInt", TypeInt),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.typ.IsTypedef()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTypeIsBasic(t *testing.T) {
	tests := []struct {
		name     string
		typ      *Type
		expected bool
	}{
		{
			name:     "Int type is basic",
			typ:      TypeInt,
			expected: true,
		},
		{
			name:     "String type is basic",
			typ:      TypeString,
			expected: true,
		},
		{
			name:     "Struct is not basic",
			typ:      Struct("User", nil),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.typ.IsBasic()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTypeIsMethod(t *testing.T) {
	tests := []struct {
		name     string
		typ      *Type
		expected bool
	}{
		{
			name:     "Method type is method",
			typ:      &Type{Kind: TypeKindMethod},
			expected: true,
		},
		{
			name:     "Function is not method",
			typ:      &Type{Kind: TypeKindFunction},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.typ.IsMethod()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTypeWithDoc(t *testing.T) {
	typ := TypeInt
	result := typ.WithDoc("This is an integer")
	assert.Equal(t, "This is an integer", result.Doc)
	assert.Equal(t, typ, result) // Should return the same pointer
}

func TestTypeWithComments(t *testing.T) {
	typ := TypeInt
	comments := []string{"Line 1", "Line 2"}
	result := typ.WithComments(comments)
	assert.Equal(t, comments, result.Comments)
	assert.Equal(t, typ, result) // Should return the same pointer
}

func TestTypeName(t *testing.T) {
	tests := []struct {
		name     string
		typ      *Type
		contains string
	}{
		{
			name:     "Struct name representation",
			typ:      Struct("User", nil),
			contains: "User",
		},
		{
			name:     "Typedef name representation",
			typ:      Typedef("MyInt", TypeInt),
			contains: "MyInt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, tt.typ.Name, tt.contains)
		})
	}
}

// Tests for TypeKind helper methods
func TestTypeKindString(t *testing.T) {
	tests := []struct {
		kind     TypeKind
		expected string
	}{
		{TypeKindBasic, "basic"},
		{TypeKindStruct, "struct"},
		{TypeKindInterface, "interface"},
		{TypeKindPointer, "pointer"},
		{TypeKindSlice, "slice"},
		{TypeKindArray, "array"},
		{TypeKindMap, "map"},
		{TypeKindFunction, "function"},
		{TypeKindMethod, "method"},
		{TypeKindChannel, "channel"},
		{TypeKindTypedef, "typedef"},
		{TypeKindAlias, "alias"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.kind.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTypeKindFromString(t *testing.T) {
	tests := []struct {
		name     string
		expected TypeKind
	}{
		{"basic", TypeKindBasic},
		{"struct", TypeKindStruct},
		{"interface", TypeKindInterface},
		{"pointer", TypeKindPointer},
		{"slice", TypeKindSlice},
		{"array", TypeKindArray},
		{"map", TypeKindMap},
		{"function", TypeKindFunction},
		{"method", TypeKindMethod},
		{"channel", TypeKindChannel},
		{"typedef", TypeKindTypedef},
		{"alias", TypeKindAlias},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TypeKindFromString(tt.name)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}

	t.Run("unknown returns error", func(t *testing.T) {
		_, err := TypeKindFromString("unknown")
		assert.Error(t, err)
	})
}

func TestTypeKindMarshalUnmarshalJSON(t *testing.T) {
	tests := []TypeKind{
		TypeKindBasic,
		TypeKindStruct,
		TypeKindInterface,
		TypeKindPointer,
		TypeKindSlice,
		TypeKindArray,
		TypeKindMap,
		TypeKindFunction,
		TypeKindMethod,
		TypeKindChannel,
		TypeKindTypedef,
		TypeKindAlias,
	}

	for _, kind := range tests {
		t.Run(kind.String(), func(t *testing.T) {
			marshaled, err := kind.MarshalJSON()
			require.NoError(t, err)

			var unmarshaled TypeKind
			err = unmarshaled.UnmarshalJSON(marshaled)
			require.NoError(t, err)

			assert.Equal(t, kind, unmarshaled)
		})
	}
}

// Tests for ChanDirection helper methods
func TestChanDirectionString(t *testing.T) {
	tests := []struct {
		direction ChanDirection
		expected  string
	}{
		{ChanBidirectional, "bidirectional"},
		{ChanRecv, "recv"},
		{ChanSend, "send"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.direction.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestChanDirectionFromString(t *testing.T) {
	tests := []struct {
		name     string
		expected ChanDirection
	}{
		{"bidirectional", ChanBidirectional},
		{"recv", ChanRecv},
		{"send", ChanSend},
		{"unknown", ChanBidirectional}, // Default to bidirectional
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := ChanDirectionFromString(tt.name)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestChanDirectionMarshalUnmarshalJSON(t *testing.T) {
	tests := []ChanDirection{
		ChanBidirectional,
		ChanRecv,
		ChanSend,
	}

	for _, direction := range tests {
		t.Run(direction.String(), func(t *testing.T) {
			marshaled, err := direction.MarshalJSON()
			require.NoError(t, err)

			var unmarshaled ChanDirection
			err = unmarshaled.UnmarshalJSON(marshaled)
			require.NoError(t, err)

			assert.Equal(t, direction, unmarshaled)
		})
	}
}

// Tests for factory functions with nil handling
func TestTypeFactoriesWithNil(t *testing.T) {
	t.Run("MakeTypeReference with nil", func(t *testing.T) {
		result := MakeTypeReference(nil)
		assert.Nil(t, result)
	})

	t.Run("Pointer with nil", func(t *testing.T) {
		result := Pointer(nil)
		assert.Nil(t, result)
	})

	t.Run("Slice with nil", func(t *testing.T) {
		result := Slice(nil)
		assert.Nil(t, result)
	})

	t.Run("Array with nil", func(t *testing.T) {
		result := Array(nil, 5)
		assert.Nil(t, result)
	})

	t.Run("Map with nil key", func(t *testing.T) {
		result := Map(nil, TypeInt)
		assert.Nil(t, result)
	})

	t.Run("Map with nil value", func(t *testing.T) {
		result := Map(TypeString, nil)
		assert.Nil(t, result)
	})

	t.Run("Chan with nil", func(t *testing.T) {
		result := Chan(nil, ChanBidirectional)
		assert.Nil(t, result)
	})
}

// Tests for Variable and Value factory functions
func TestNewVariable(t *testing.T) {
	result := NewVariable("myVar", 42)
	assert.NotNil(t, result)
	assert.Equal(t, "myVar", result.Name)
	assert.NotNil(t, result.Value)
}

func TestNewVariableWithType(t *testing.T) {
	result := NewVariableWithType("myVar", 42, TypeInt)
	assert.NotNil(t, result)
	assert.Equal(t, "myVar", result.Name)
	assert.NotNil(t, result.Value)
}

func TestTransformToValue(t *testing.T) {
	transformed := TransformToValue(42)
	assert.NotNil(t, transformed)
	assert.Equal(t, 42, transformed.Value)
	assert.Equal(t, "int", transformed.Type.Name)
}

func TestTransformToValueWithType(t *testing.T) {
	transformed := TransformToValueWithType(42, TypeInt)
	assert.NotNil(t, transformed)
	assert.Equal(t, 42, transformed.Value)
	assert.Equal(t, "int", transformed.Type.Name)
}

func TestVariableWithDoc(t *testing.T) {
	v := &Variable{Name: "test"}
	result := v.WithDoc("This is a test variable")

	assert.Equal(t, "This is a test variable", result.Doc)
	assert.Equal(t, v, result) // Should return the same pointer
}

// Tests for factory functions that construct complex types
func TestStructFactory(t *testing.T) {
	fields := []*Field{
		{Name: "ID", Type: MakeTypeReference(TypeInt)},
		{Name: "Name", Type: MakeTypeReference(TypeString)},
	}

	typ := Struct("User", fields)

	assert.Equal(t, "User", typ.Name)
	assert.Equal(t, TypeKindStruct, typ.Kind)
	assert.Equal(t, 2, len(typ.Fields))
	assert.Equal(t, "ID", typ.Fields[0].Name)
}

func TestInterfaceFactory(t *testing.T) {
	typ := Interface("Reader", nil)

	assert.Equal(t, "Reader", typ.Name)
	assert.Equal(t, TypeKindInterface, typ.Kind)
}

func TestAliasFactory(t *testing.T) {
	alias := Alias("MyInt", TypeInt)

	assert.Equal(t, "MyInt", alias.Name)
	assert.Equal(t, TypeKindAlias, alias.Kind)
	assert.NotNil(t, alias.OriginalType)
	assert.Equal(t, "int", alias.OriginalType.Name)
	assert.Equal(t, TypeInt, alias.OriginalType.Type)
}

func TestTypedefFactory(t *testing.T) {
	typedef := Typedef("Status", TypeInt)

	assert.Equal(t, "Status", typedef.Name)
	assert.Equal(t, TypeKindTypedef, typedef.Kind)
	assert.NotNil(t, typedef.OriginalType)
	assert.Equal(t, "int", typedef.OriginalType.Name)
	assert.Equal(t, TypeInt, typedef.OriginalType.Type)
}

func TestGetBasicType(t *testing.T) {
	tests := []struct {
		name     string
		typename string
		expected *Type
	}{
		{"int", "int", TypeInt},
		{"string", "string", TypeString},
		{"bool", "bool", TypeBool},
		{"float64", "float64", TypeFloat64},
		{"unknown", "unknown_type", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetBasicType(tt.typename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsBasicTypeFunc(t *testing.T) {
	tests := []struct {
		name     string
		typename string
		expected bool
	}{
		{"int type", "int", true},
		{"string type", "string", true},
		{"user struct", "User", false},
		{"unknown type", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBasicType(tt.typename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Tests for complex composed types
func TestPointerToSliceOfMaps(t *testing.T) {
	// Create: *[]map[string]bool
	mapType := Map(TypeString, TypeBool)
	require.NotNil(t, mapType)

	sliceType := Slice(mapType)
	require.NotNil(t, sliceType)

	pointerType := Pointer(sliceType)
	require.NotNil(t, pointerType)

	assert.Equal(t, TypeKindPointer, pointerType.Kind)
	assert.NotNil(t, pointerType.Elem)
	if pointerType.Elem.Type != nil {
		assert.Equal(t, TypeKindSlice, pointerType.Elem.Type.Kind)
	}
}

func TestArrayOfPointers(t *testing.T) {
	// Create: [10]*User
	userStruct := Struct("User", []*Field{
		{Name: "ID", Type: MakeTypeReference(TypeInt)},
	})

	ptrType := Pointer(userStruct)
	require.NotNil(t, ptrType)

	arrayType := Array(ptrType, 10)
	require.NotNil(t, arrayType)

	assert.Equal(t, TypeKindArray, arrayType.Kind)
	if arrayType.Elem != nil && arrayType.Elem.Type != nil {
		assert.Equal(t, TypeKindPointer, arrayType.Elem.Type.Kind)
	}
}

func TestMapValueMarshalJSON(t *testing.T) {
	v := &Value{
		Value: map[string]interface{}{"key": "value"},
		Type:  MakeTypeReference(Map(TypeString, TypeString)),
	}

	data, err := json.Marshal(v)
	require.NoError(t, err)
	assert.NotNil(t, data)
}
