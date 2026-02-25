package reflect

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompositeTypeStructures(t *testing.T) {
	tests := []struct {
		name      string
		typeStr   string
		checkType func(t *testing.T, typ *Type)
	}{
		{
			name:    "Slice type",
			typeStr: "[]Foo",
			checkType: func(t *testing.T, typ *Type) {
				assert.Equal(t, TypeKindSlice, typ.Kind)
				assert.Equal(t, "[]Foo", typ.Name)
				assert.NotNil(t, typ.Elem)
				assert.Equal(t, "Foo", typ.Elem.Name)
			},
		},
		{
			name:    "Pointer type",
			typeStr: "*Foo",
			checkType: func(t *testing.T, typ *Type) {
				assert.Equal(t, TypeKindPointer, typ.Kind)
				assert.Equal(t, "*Foo", typ.Name)
				assert.NotNil(t, typ.Elem)
				assert.Equal(t, "Foo", typ.Elem.Name)
			},
		},
		{
			name:    "Pointer to slice",
			typeStr: "*[]Foo",
			checkType: func(t *testing.T, typ *Type) {
				assert.Equal(t, TypeKindPointer, typ.Kind)
				assert.Equal(t, "*[]Foo", typ.Name)
				assert.NotNil(t, typ.Elem)
				assert.Equal(t, TypeKindSlice, typ.Elem.Type.Kind)
				assert.Equal(t, "Foo", typ.Elem.Type.Elem.Name)
			},
		},
		{
			name:    "Array type",
			typeStr: "[5]Foo",
			checkType: func(t *testing.T, typ *Type) {
				assert.Equal(t, TypeKindArray, typ.Kind)
				assert.Equal(t, "Foo", typ.Elem.Name)
				assert.Equal(t, 5, typ.Size)
			},
		},
		{
			name:    "Map type with string key and int value",
			typeStr: "map[string]int",
			checkType: func(t *testing.T, typ *Type) {
				assert.Equal(t, TypeKindMap, typ.Kind)
				assert.Equal(t, "map[string]int", typ.Name)
				assert.NotNil(t, typ.Key)
				assert.Equal(t, "string", typ.Key.Name)
				assert.NotNil(t, typ.Value)
				assert.Equal(t, "int", typ.Value.Name)
			},
		},
		{
			name:    "Map with complex value",
			typeStr: "map[string]*Foo",
			checkType: func(t *testing.T, typ *Type) {
				assert.Equal(t, TypeKindMap, typ.Kind)
				assert.Equal(t, "string", typ.Key.Name)
				assert.NotNil(t, typ.Value)
				assert.Equal(t, TypeKindPointer, typ.Value.Type.Kind)
				assert.Equal(t, "Foo", typ.Value.Type.Elem.Name)
			},
		},
		{
			name:    "Channel type (bidirectional)",
			typeStr: "chan Foo",
			checkType: func(t *testing.T, typ *Type) {
				assert.Equal(t, TypeKindChannel, typ.Kind)
				assert.Equal(t, ChanBidirectional, typ.ChanDir)
				assert.NotNil(t, typ.Elem)
				assert.Equal(t, "Foo", typ.Elem.Name)
			},
		},
		{
			name:    "Receive-only channel",
			typeStr: "<-chan Foo",
			checkType: func(t *testing.T, typ *Type) {
				assert.Equal(t, TypeKindChannel, typ.Kind)
				assert.Equal(t, ChanRecv, typ.ChanDir)
				assert.NotNil(t, typ.Elem)
				assert.Equal(t, "Foo", typ.Elem.Name)
			},
		},
		{
			name:    "Basic type",
			typeStr: "string",
			checkType: func(t *testing.T, typ *Type) {
				assert.Equal(t, TypeKindBasic, typ.Kind)
				assert.Equal(t, "string", typ.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the Go code to get an AST
			code := "package test\ntype T " + tt.typeStr
			fset := token.NewFileSet()
			astFile, err := parser.ParseFile(fset, "test.go", code, parser.AllErrors)
			require.NoError(t, err)

			// Extract the type from the parsed code
			require.NotEmpty(t, astFile.Decls)
			genDecl, ok := astFile.Decls[0].(*ast.GenDecl)
			require.True(t, ok)
			require.NotEmpty(t, genDecl.Specs)

			typeSpec := genDecl.Specs[0].(*ast.TypeSpec)
			typ := astExprToType(typeSpec.Type)
			require.NotNil(t, typ)

			tt.checkType(t, typ)
		})
	}
}

func TestSliceTypeResolution(t *testing.T) {
	// Test that parseParameters correctly creates Type structures for composite types
	code := "func GetUsers() []User { return nil }"

	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, "test.go", "package test\n"+code, parser.AllErrors)
	require.NoError(t, err)

	// Extract the function declaration
	funcDecl := astFile.Decls[0].(*ast.FuncDecl)
	require.NotNil(t, funcDecl.Type.Results)

	// Parse the results
	results := parseParameters(funcDecl.Type.Results)
	require.Len(t, results, 1)

	result := results[0]
	assert.Equal(t, "[]User", result.Type.Name, "Result type name should preserve composite syntax")

	// The Type field should contain a proper Type structure for the slice
	require.NotNil(t, result.Type.Type, "Type field should contain the composite type")
	resolvedType := result.Type.Type
	assert.Equal(t, TypeKindSlice, resolvedType.Kind, "Should be a slice type")
	assert.NotNil(t, resolvedType.Elem, "Slice should have an item type")
	assert.Equal(t, "User", resolvedType.Elem.Name, "Item should be User")
}

func TestMapTypeResolution(t *testing.T) {
	// Test that parseParameters correctly creates Type structures for map types
	code := "func Process(data map[string]int) {}"

	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, "test.go", "package test\n"+code, parser.AllErrors)
	require.NoError(t, err)

	// Extract the function declaration
	funcDecl := astFile.Decls[0].(*ast.FuncDecl)
	require.NotNil(t, funcDecl.Type.Params)

	// Parse the parameters
	params := parseParameters(funcDecl.Type.Params)
	require.Len(t, params, 1)

	param := params[0]
	assert.Equal(t, "data", param.Name)
	assert.Equal(t, "map[string]int", param.Type.Name)

	// The Type field should contain a proper Type structure for the map
	require.NotNil(t, param.Type.Type, "Type field should contain the composite type")
	mapType := param.Type.Type
	assert.Equal(t, TypeKindMap, mapType.Kind, "Should be a map type")
	assert.NotNil(t, mapType.Key, "Map should have a key type")
	assert.Equal(t, "string", mapType.Key.Name)
	assert.NotNil(t, mapType.Value, "Map should have a value type")
	assert.Equal(t, "int", mapType.Value.Name)
}

func TestPointerSliceResolution(t *testing.T) {
	// Test that parseParameters correctly creates Type structures for []*User
	code := "func ProcessUsers(users []*User) {}"

	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, "test.go", "package test\n"+code, parser.AllErrors)
	require.NoError(t, err)

	// Extract the function declaration
	funcDecl := astFile.Decls[0].(*ast.FuncDecl)
	require.NotNil(t, funcDecl.Type.Params)

	// Parse the parameters
	params := parseParameters(funcDecl.Type.Params)
	require.Len(t, params, 1)

	param := params[0]
	assert.Equal(t, "users", param.Name)
	assert.Equal(t, "[]*User", param.Type.Name)

	// The Type field should contain a proper Type structure
	require.NotNil(t, param.Type.Type, "Type field should contain the composite type")
	sliceType := param.Type.Type
	assert.Equal(t, TypeKindSlice, sliceType.Kind, "Should be a slice type")
	assert.NotNil(t, sliceType.Elem, "Slice should have an item type")
	assert.Equal(t, TypeKindPointer, sliceType.Elem.Type.Kind, "Item should be a pointer type")
	assert.Equal(t, "User", sliceType.Elem.Type.Elem.Name, "Pointer should point to User")
}

func TestFunctionTypeStructures(t *testing.T) {
	tests := []struct {
		name      string
		typeStr   string
		checkType func(t *testing.T, typ *Type)
	}{
		{
			name:    "Simple function type with args and results",
			typeStr: "func(int, string) (bool, error)",
			checkType: func(t *testing.T, typ *Type) {
				assert.Equal(t, TypeKindFunction, typ.Kind)
				require.NotNil(t, typ.Args)
				require.NotNil(t, typ.Results)
				require.Equal(t, 2, len(typ.Args), "Function should have 2 arguments")
				require.Equal(t, 2, len(typ.Results), "Function should have 2 results")

				// Check argument types
				assert.Equal(t, "int", typ.Args[0].Name)
				assert.Equal(t, "string", typ.Args[1].Name)

				// Check result types
				assert.Equal(t, "bool", typ.Results[0].Name)
				assert.Equal(t, "error", typ.Results[1].Name)
			},
		},
		{
			name:    "Function with no arguments",
			typeStr: "func() string",
			checkType: func(t *testing.T, typ *Type) {
				assert.Equal(t, TypeKindFunction, typ.Kind)
				require.NotNil(t, typ.Args)
				require.NotNil(t, typ.Results)
				require.Equal(t, 0, len(typ.Args), "Function should have 0 arguments")
				require.Equal(t, 1, len(typ.Results), "Function should have 1 result")
				assert.Equal(t, "string", typ.Results[0].Name)
			},
		},
		{
			name:    "Function with no results",
			typeStr: "func(string)",
			checkType: func(t *testing.T, typ *Type) {
				assert.Equal(t, TypeKindFunction, typ.Kind)
				require.NotNil(t, typ.Args)
				require.NotNil(t, typ.Results)
				require.Equal(t, 1, len(typ.Args), "Function should have 1 argument")
				require.Equal(t, 0, len(typ.Results), "Function should have 0 results")
				assert.Equal(t, "string", typ.Args[0].Name)
			},
		},
		{
			name:    "Function with composite types",
			typeStr: "func([]*string, map[string]int) (*[]error, chan bool)",
			checkType: func(t *testing.T, typ *Type) {
				assert.Equal(t, TypeKindFunction, typ.Kind)
				require.NotNil(t, typ.Args)
				require.NotNil(t, typ.Results)
				require.Equal(t, 2, len(typ.Args), "Function should have 2 arguments")
				require.Equal(t, 2, len(typ.Results), "Function should have 2 results")

				// Check argument types are properly parsed as composite types
				require.NotNil(t, typ.Args[0].Type)
				assert.Equal(t, TypeKindSlice, typ.Args[0].Type.Kind)
				require.NotNil(t, typ.Args[1].Type)
				assert.Equal(t, TypeKindMap, typ.Args[1].Type.Kind)

				// Check result types are properly parsed as composite types
				require.NotNil(t, typ.Results[0].Type)
				assert.Equal(t, TypeKindPointer, typ.Results[0].Type.Kind)
				require.NotNil(t, typ.Results[1].Type)
				assert.Equal(t, TypeKindChannel, typ.Results[1].Type.Kind)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.ParseExpr(tt.typeStr)
			require.NoError(t, err)

			typ := astExprToType(expr)
			require.NotNil(t, typ)
			tt.checkType(t, typ)
		})
	}
}
