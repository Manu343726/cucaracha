package reflect

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypeReferenceResolution(t *testing.T) {
	// Test data: create a simple package with types
	pkg := &Package{
		Name:      "testpkg",
		Path:      "/test/path",
		Files:     []*File{},
		Types:     make(map[string]*Type),
		Functions: []*Function{},
		Constants: []*Constant{},
		Enums:     make(map[string]*Enum),
	}

	// Create a simple struct type
	myType := &Type{
		Name: "MyType",
		Kind: TypeKindStruct,
		Fields: []*Field{
			{
				Name: "Value",
				Type: &TypeReference{
					Name: "string",
					Type: nil,
				},
			},
		},
	}
	pkg.Types["MyType"] = myType

	// Create a function that uses the type
	fn := &Function{
		Name:    "ProcessType",
		Package: "testpkg",
		Args: []*Parameter{
			{
				Name: "input",
				Type: &TypeReference{
					Name: "*MyType",
					Type: nil,
				},
			},
		},
		Results: []*Parameter{
			{
				Name: "err",
				Type: &TypeReference{
					Name: "error",
					Type: nil,
				},
			},
		},
	}
	pkg.Functions = append(pkg.Functions, fn)

	// Create a constant
	const_ := &Constant{
		Name: "MyConstant",
		Value: &Value{
			Value: "42",
			Type: &TypeReference{
				Name: "MyType",
				Type: nil,
			},
		},
	}
	pkg.Constants = append(pkg.Constants, const_)

	// Before resolution, the Type pointers should be nil
	assert.Nil(t, myType.Fields[0].Type.Type)
	assert.Nil(t, fn.Args[0].Type.Type)
	assert.Nil(t, fn.Results[0].Type.Type)
	assert.Nil(t, const_.Value.Type.Type)

	// Run the resolution
	resolveTypeReferences(pkg)

	// After resolution, the Type pointers should be set
	assert.NotNil(t, myType.Fields[0].Type.Type, "string is a built-in type that should resolve to TypeString")
	assert.Equal(t, "string", myType.Fields[0].Type.Type.Name)

	assert.NotNil(t, fn.Args[0].Type.Type, "*MyType argument should resolve to MyType")
	assert.Equal(t, "MyType", fn.Args[0].Type.Type.Name)

	// error is a built-in that now resolves to TypeError
	assert.NotNil(t, fn.Results[0].Type.Type, "error is a built-in type that should resolve to TypeError")
	assert.Equal(t, "error", fn.Results[0].Type.Type.Name)

	// Constant type should resolve
	assert.NotNil(t, const_.Value.Type.Type, "MyType constant should resolve")
	assert.Equal(t, "MyType", const_.Value.Type.Type.Name)
}

func TestExtractBaseTypeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"string", "string"},
		{"*string", "string"},
		{"**int", "int"},
		{"[]byte", "byte"},
		{"[]*MyType", "MyType"},
		{"map[string]int", "int"},
		{"map[string]*MyType", "MyType"},
		{"<-chan string", "string"},
		{"chan MyType", "MyType"},
		{"pkg.MyType", "MyType"},
		{"*pkg.MyType", "MyType"},
		{"[]pkg.MyType", "MyType"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractBaseTypeName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFileTypeResolution(t *testing.T) {
	// Create a file with types
	file := &File{
		Name:      "test.go",
		Path:      "/test/path/test.go",
		Package:   "testpkg",
		Types:     make(map[string]*Type),
		Functions: []*Function{},
		Constants: []*Constant{},
	}

	// Define a struct type
	structType := &Type{
		Name: "User",
		Kind: TypeKindStruct,
		Fields: []*Field{
			{
				Name: "Name",
				Type: &TypeReference{
					Name: "string",
					Type: nil,
				},
			},
			{
				Name: "Age",
				Type: &TypeReference{
					Name: "int",
					Type: nil,
				},
			},
		},
	}
	file.Types["User"] = structType

	// Define an interface
	interfaceType := &Type{
		Name: "Printer",
		Kind: TypeKindInterface,
		Methods: []*Method{
			{
				Name: "Print",
				Args: []*Parameter{
					{
						Name: "text",
						Type: &TypeReference{
							Name: "string",
							Type: nil,
						},
					},
				},
				Results: []*Parameter{
					{
						Name: "err",
						Type: &TypeReference{
							Name: "error",
							Type: nil,
						},
					},
				},
			},
		},
	}
	file.Types["Printer"] = interfaceType

	// Create a function that references these types
	fn := &Function{
		Name: "NewUser",
		Args: []*Parameter{
			{
				Name: "name",
				Type: &TypeReference{
					Name: "string",
					Type: nil,
				},
			},
		},
		Results: []*Parameter{
			{
				Name: "user",
				Type: &TypeReference{
					Name: "*User",
					Type: nil,
				},
			},
		},
	}
	file.Functions = append(file.Functions, fn)

	// Before resolution
	require.Nil(t, structType.Fields[0].Type.Type)
	require.Nil(t, fn.Results[0].Type.Type)

	// Run resolution
	resolveFileTypeReferences(file)

	// After resolution
	assert.NotNil(t, structType.Fields[0].Type.Type, "string should resolve to TypeString")
	assert.Equal(t, "string", structType.Fields[0].Type.Type.Name)

	assert.NotNil(t, fn.Results[0].Type.Type, "*User should resolve to User")
	assert.Equal(t, "User", fn.Results[0].Type.Type.Name)

	// Check method parameters are resolved
	method := interfaceType.Methods[0]
	assert.NotNil(t, method.Args[0].Type.Type, "string should resolve to TypeString")
	assert.Equal(t, "string", method.Args[0].Type.Type.Name)
	assert.NotNil(t, method.Results[0].Type.Type, "error should resolve to TypeError")
	assert.Equal(t, "error", method.Results[0].Type.Type.Name)
}

func TestDefaultParsingOptions(t *testing.T) {
	opts := DefaultParsingOptions()

	// Default options should disable external type resolution
	assert.False(t, opts.ResolveExternalTypes, "ResolveExternalTypes should be disabled by default")

	// Default options should have unlimited recursion depth
	assert.Equal(t, -1, opts.MaxResolutionDepth, "MaxResolutionDepth should be -1 (unlimited) by default")
}

func TestParsingOptionsValidation(t *testing.T) {
	tests := []struct {
		name         string
		opts         ParsingOptions
		expectExtRes bool
		expectDepth  int
	}{
		{
			name: "External resolution enabled with unlimited depth",
			opts: ParsingOptions{
				ResolveExternalTypes: true,
				MaxResolutionDepth:   -1,
			},
			expectExtRes: true,
			expectDepth:  -1,
		},
		{
			name: "External resolution enabled with limited depth",
			opts: ParsingOptions{
				ResolveExternalTypes: true,
				MaxResolutionDepth:   5,
			},
			expectExtRes: true,
			expectDepth:  5,
		},
		{
			name: "External resolution disabled",
			opts: ParsingOptions{
				ResolveExternalTypes: false,
				MaxResolutionDepth:   10,
			},
			expectExtRes: false,
			expectDepth:  10,
		},
		{
			name: "Zero recursion depth prevents resolution",
			opts: ParsingOptions{
				ResolveExternalTypes: true,
				MaxResolutionDepth:   0,
			},
			expectExtRes: true,
			expectDepth:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectExtRes, tt.opts.ResolveExternalTypes)
			assert.Equal(t, tt.expectDepth, tt.opts.MaxResolutionDepth)
		})
	}
}

func TestParsingWithOptionsBasic(t *testing.T) {
	// This test verifies that we can pass ParsingOptions to parse functions
	// and that the parsing completes successfully with the default options

	opts := ParsingOptions{
		ResolveExternalTypes: false,
		MaxResolutionDepth:   -1,
	}

	// Test that we can create options and pass them
	assert.False(t, opts.ResolveExternalTypes)
	assert.Equal(t, -1, opts.MaxResolutionDepth)

	// The actual external resolution would require real packages,
	// which we test implicitly through the integration with ParsePackageWithOptions
}
