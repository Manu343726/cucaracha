package codegen

import (
	"bytes"
	"testing"

	"github.com/Manu343726/cucaracha/pkg/reflect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for GenerateTypeCode for various types

func TestGenerateTypeCodeStruct(t *testing.T) {
	t.Run("Generate code for struct type", func(t *testing.T) {
		fields := []*reflect.Field{
			{Name: "ID", Type: reflect.MakeTypeReference(reflect.TypeInt)},
			{Name: "Name", Type: reflect.MakeTypeReference(reflect.TypeString)},
		}
		userType := reflect.Struct("User", fields)

		var buf bytes.Buffer
		gen := NewGenerator(&reflect.Package{}, &buf)
		err := gen.GenerateStructCode(userType)
		require.NoError(t, err)

		code := buf.String()
		assert.Contains(t, code, "type User struct")
		assert.Contains(t, code, "ID")
		assert.Contains(t, code, "Name")
	})
}

func TestGenerateTypeCodeAlias(t *testing.T) {
	t.Run("Generate code for type alias", func(t *testing.T) {
		myIntAlias := reflect.Alias("MyInt", reflect.TypeInt)

		var buf bytes.Buffer
		gen := NewGenerator(&reflect.Package{}, &buf)
		err := gen.GenerateAliasCode(myIntAlias)
		require.NoError(t, err)

		code := buf.String()
		assert.Contains(t, code, "type MyInt")
		assert.Contains(t, code, "int")
	})
}

func TestGenerateTypeCodeTypedef(t *testing.T) {
	t.Run("Generate code for typedef", func(t *testing.T) {
		statusTypedef := reflect.Typedef("Status", reflect.TypeInt)

		var buf bytes.Buffer
		gen := NewGenerator(&reflect.Package{}, &buf)
		err := gen.GenerateTypedefCode(statusTypedef)
		require.NoError(t, err)

		code := buf.String()
		assert.Contains(t, code, "type Status")
		assert.Contains(t, code, "int")
	})
}

// Tests for GenerateValue with different basic types
func TestGenerateValueBasicTypes(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		typ      *reflect.Type
		contains []string
	}{
		{
			name:     "Boolean true",
			value:    true,
			typ:      reflect.TypeBool,
			contains: []string{"true"},
		},
		{
			name:     "Boolean false",
			value:    false,
			typ:      reflect.TypeBool,
			contains: []string{"false"},
		},
		{
			name:     "Integer",
			value:    42,
			typ:      reflect.TypeInt,
			contains: []string{"42"},
		},
		{
			name:     "String",
			value:    "hello",
			typ:      reflect.TypeString,
			contains: []string{"hello"},
		},
		{
			name:     "Float",
			value:    3.14,
			typ:      reflect.TypeFloat64,
			contains: []string{"3.14"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := &reflect.Value{
				Value: tt.value,
				Type:  reflect.MakeTypeReference(tt.typ),
			}

			var buf bytes.Buffer
			gen := NewGenerator(&reflect.Package{}, &buf)
			err := gen.GenerateValue(value)
			require.NoError(t, err)

			code := buf.String()
			for _, expected := range tt.contains {
				assert.Contains(t, code, expected)
			}
		})
	}
}

// Tests for GenerateMethodSignature with various parameter counts
func TestGenerateMethodSignatureComplex(t *testing.T) {
	t.Run("Method with multiple parameters and returns", func(t *testing.T) {
		method := &reflect.Method{
			Name: "Process",
			Receiver: &reflect.Parameter{
				Name: "r",
				Type: reflect.MakeTypeReference(reflect.Struct("Reader", nil)),
			},
			Args: []*reflect.Parameter{
				{Name: "ctx", Type: reflect.MakeTypeReference(reflect.TypeString)},
				{Name: "data", Type: reflect.MakeTypeReference(reflect.TypeInt)},
			},
			Results: []*reflect.Parameter{
				{Type: reflect.MakeTypeReference(reflect.TypeString)},
				{Type: reflect.MakeTypeReference(reflect.TypeBool)},
			},
		}

		var buf bytes.Buffer
		gen := NewGenerator(&reflect.Package{}, &buf)
		gen.GenerateMethodSignature(method)

		code := buf.String()
		assert.Contains(t, code, "Process")
	})
}

// Tests for GenerateFunctionSignature with various signatures
func TestGenerateFunctionSignatureComplex(t *testing.T) {
	t.Run("Function with variadic parameters", func(t *testing.T) {
		fn := &reflect.Function{
			Name: "Format",
			Args: []*reflect.Parameter{
				{Name: "format", Type: reflect.MakeTypeReference(reflect.TypeString)},
				{Name: "args", Type: reflect.MakeTypeReference(reflect.Slice(reflect.TypeString))},
			},
			Results: []*reflect.Parameter{
				{Type: reflect.MakeTypeReference(reflect.TypeString)},
			},
		}

		var buf bytes.Buffer
		gen := NewGenerator(&reflect.Package{}, &buf)
		gen.GenerateFunctionSignature(fn)

		code := buf.String()
		assert.Contains(t, code, "Format")
	})
}

// Tests for complex type generation scenarios
func TestGenerateComplexNestedStructure(t *testing.T) {
	t.Run("Generate nested struct", func(t *testing.T) {
		addressFields := []*reflect.Field{
			{Name: "Street", Type: reflect.MakeTypeReference(reflect.TypeString)},
			{Name: "City", Type: reflect.MakeTypeReference(reflect.TypeString)},
		}
		addressType := reflect.Struct("Address", addressFields)

		userFields := []*reflect.Field{
			{Name: "Name", Type: reflect.MakeTypeReference(reflect.TypeString)},
			{Name: "Age", Type: reflect.MakeTypeReference(reflect.TypeInt)},
			{Name: "Address", Type: reflect.MakeTypeReference(addressType)},
		}
		userType := reflect.Struct("User", userFields)

		var buf bytes.Buffer
		gen := NewGenerator(&reflect.Package{}, &buf)
		err := gen.GenerateStructCode(userType)
		require.NoError(t, err)

		code := buf.String()
		assert.Contains(t, code, "User")
		assert.Contains(t, code, "Address")
	})
}

// Tests for parameter formatting
func TestFormatParametersFunc(t *testing.T) {
	t.Run("Format basic parameters indirectly", func(t *testing.T) {
		fn := &reflect.Function{
			Name: "Add",
			Args: []*reflect.Parameter{
				{Name: "x", Type: reflect.MakeTypeReference(reflect.TypeInt)},
				{Name: "y", Type: reflect.MakeTypeReference(reflect.TypeInt)},
			},
			Results: []*reflect.Parameter{
				{Type: reflect.MakeTypeReference(reflect.TypeInt)},
			},
		}

		var buf bytes.Buffer
		gen := NewGenerator(&reflect.Package{}, &buf)
		gen.GenerateFunctionSignature(fn)

		code := buf.String()
		assert.Contains(t, code, "Add")
	})
}

// Tests for GenerateMethodSignature with receiver
func TestGenerateMethodSignatureWithReceiverFunc(t *testing.T) {
	t.Run("Method signature with receiver", func(t *testing.T) {
		method := &reflect.Method{
			Name: "String",
			Receiver: &reflect.Parameter{
				Name: "e",
				Type: reflect.MakeTypeReference(reflect.Typedef("Status", reflect.TypeInt)),
			},
			Results: []*reflect.Parameter{
				{Type: reflect.MakeTypeReference(reflect.TypeString)},
			},
		}

		var buf bytes.Buffer
		gen := NewGenerator(&reflect.Package{}, &buf)
		gen.GenerateMethodSignature(method)

		code := buf.String()
		assert.NotEmpty(t, code)
	})
}

// Tests for GenerateInterfaceCode
func TestGenerateInterfaceCodeComplex(t *testing.T) {
	t.Run("Generate interface with multiple methods", func(t *testing.T) {
		methods := []*reflect.Method{
			{
				Name: "Read",
				Args: []*reflect.Parameter{
					{Name: "p", Type: reflect.MakeTypeReference(reflect.Slice(reflect.TypeInt8))},
				},
				Results: []*reflect.Parameter{
					{Type: reflect.MakeTypeReference(reflect.TypeInt)},
					{Type: reflect.MakeTypeReference(reflect.TypeString)},
				},
			},
			{
				Name: "Close",
				Results: []*reflect.Parameter{
					{Type: reflect.MakeTypeReference(reflect.TypeString)},
				},
			},
		}
		readerInterface := reflect.Interface("Reader", methods)

		var buf bytes.Buffer
		gen := NewGenerator(&reflect.Package{}, &buf)
		err := gen.GenerateInterfaceCode(readerInterface)
		require.NoError(t, err)

		code := buf.String()
		assert.Contains(t, code, "type Reader interface")
		assert.Contains(t, code, "Read")
		assert.Contains(t, code, "Close")
	})
}

// Tests for GenerateTypeCode with variable assignments
func TestGenerateTypeCodeWithVariables(t *testing.T) {
	t.Run("Generate code for variable assignment", func(t *testing.T) {
		variable := reflect.NewVariable("result", 42)

		assert.NotNil(t, variable)
		assert.Equal(t, "result", variable.Name)
		assert.NotNil(t, variable.Value)
	})
}

// Tests for GenerateTypeCode with package context
func TestGenerateTypeCodeWithPackageContext(t *testing.T) {
	t.Run("Generate code within package context", func(t *testing.T) {
		pkg := &reflect.Package{
			Name: "mypackage",
		}

		userType := reflect.Struct("User", nil)
		var buf bytes.Buffer
		gen := NewGenerator(pkg, &buf)
		err := gen.GenerateStructCode(userType)
		require.NoError(t, err)

		code := buf.String()
		assert.NotEmpty(t, code)
		assert.Contains(t, code, "User")
	})
}

// Tests for GenerateTypeCode with multiple fields
func TestGenerateTypeCodeMultipleFields(t *testing.T) {
	t.Run("Generate struct with many fields", func(t *testing.T) {
		fields := []*reflect.Field{
			{Name: "ID", Type: reflect.MakeTypeReference(reflect.TypeInt)},
			{Name: "Name", Type: reflect.MakeTypeReference(reflect.TypeString)},
			{Name: "Email", Type: reflect.MakeTypeReference(reflect.TypeString)},
			{Name: "Active", Type: reflect.MakeTypeReference(reflect.TypeBool)},
		}
		personType := reflect.Struct("Person", fields)

		var buf bytes.Buffer
		gen := NewGenerator(&reflect.Package{}, &buf)
		err := gen.GenerateStructCode(personType)
		require.NoError(t, err)

		code := buf.String()
		assert.Contains(t, code, "Person")
		assert.Contains(t, code, "ID")
		assert.Contains(t, code, "Name")
		assert.Contains(t, code, "Email")
		assert.Contains(t, code, "Active")
	})
}
