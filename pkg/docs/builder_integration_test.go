package docs

import (
	"testing"

	"github.com/Manu343726/cucaracha/pkg/reflect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuilderPackagePathCorrect verifies that builder uses correct package import paths, not directory paths
func TestBuilderPackagePathCorrect(t *testing.T) {
	// Create a test package with a specific import path
	pkg := &reflect.Package{
		Name:      "testpkg",
		Path:      "github.com/example/testpkg", // Full import path, not a directory
		Types:     make(map[string]*reflect.Type),
		Functions: []*reflect.Function{},
		Constants: []*reflect.Constant{},
	}

	// Add a test type with documentation
	pkg.Types["TestType"] = &reflect.Type{
		Name: "TestType",
		Kind: reflect.TypeKindStruct,
		Doc:  "TestType is a test type.",
	}

	builder := NewBuilder()
	index, err := builder.Build(pkg)

	require.NoError(t, err, "Builder should successfully build documentation")
	require.NotNil(t, index)

	// Verify the package path is stored correctly, not as "."
	assert.Equal(t, "github.com/example/testpkg", pkg.Path, "Package path should be full import path")
	assert.NotEqual(t, ".", pkg.Path, "ISSUE: Package path should NOT be relative path")

	// Verify entry uses correct package path
	expectedKey := "github.com/example/testpkg.TestType"
	entry, exists := index.Entries[expectedKey]
	require.True(t, exists, "Entry should exist with full package path key")

	assert.Equal(t, "github.com/example/testpkg", entry.PackagePath, "Entry PackagePath should match package import path")
	assert.NotEqual(t, ".", entry.PackagePath, "ISSUE: Entry PackagePath should NOT be '.'")

	// Verify ByPackage uses correct key
	pkgEntries, exists := index.ByPackage["github.com/example/testpkg"]
	require.True(t, exists, "ByPackage should use full package path as key, not '.'")
	assert.NotEqual(t, ".", "", "ISSUE: ByPackage keyed by '.' instead of full path")
	assert.Contains(t, pkgEntries, expectedKey)
}

// TestBuilderInterfaceMethodsDocumented verifies that builder documents interface methods
func TestBuilderInterfaceMethodsDocumented(t *testing.T) {
	pkg := &reflect.Package{
		Name:      "testpkg",
		Path:      "github.com/example/testpkg",
		Types:     make(map[string]*reflect.Type),
		Functions: []*reflect.Function{},
		Constants: []*reflect.Constant{},
	}

	// Create an interface type with methods
	pkg.Types["ExampleInterface"] = &reflect.Type{
		Name: "ExampleInterface",
		Kind: reflect.TypeKindInterface,
		Doc:  "ExampleInterface is an example interface.",
		Methods: []*reflect.Method{
			{
				Name: "Method1",
				Doc:  "Method1 does something.",
				Args: []*reflect.Parameter{
					{Name: "arg1", Type: &reflect.TypeReference{Name: "string"}},
				},
				Results: []*reflect.Parameter{
					{Name: "result", Type: &reflect.TypeReference{Name: "error"}},
				},
			},
			{
				Name: "Method2",
				Doc:  "Method2 does something else.",
				Args: []*reflect.Parameter{},
				Results: []*reflect.Parameter{
					{Name: "result", Type: &reflect.TypeReference{Name: "int"}},
				},
			},
		},
	}

	builder := NewBuilder()
	index, err := builder.Build(pkg)

	require.NoError(t, err, "Builder should successfully build documentation")
	require.NotNil(t, index)

	// ISSUE: Builder should document interface methods but currently doesn't
	_, exists := index.Entries["github.com/example/testpkg.ExampleInterface"]
	require.True(t, exists, "Interface type itself should be documented")

	// Check if methods are documented (they currently are not)
	method1Key := "github.com/example/testpkg.ExampleInterface.Method1"
	method1Entry, method1Exists := index.Entries[method1Key]

	require.True(t, method1Exists, "ISSUE: Interface methods are not documented. Expected entry for %q. The builder only documents types, functions, and constants, but NOT interface methods", method1Key)
	require.Equal(t, KindInterfaceMethod, method1Entry.Kind, "Method should be documented as interfaceMethod kind")

	method2Key := "github.com/example/testpkg.ExampleInterface.Method2"
	method2Entry, method2Exists := index.Entries[method2Key]
	require.True(t, method2Exists, "ISSUE: Second interface method not documented. Expected entry for %q", method2Key)
	require.Equal(t, KindInterfaceMethod, method2Entry.Kind, "Method2 should be documented as interfaceMethod kind")
}

// TestBuilderAllTypesDocumented verifies that all exported types are documented
func TestBuilderAllTypesDocumented(t *testing.T) {
	pkg := &reflect.Package{
		Name:      "testpkg",
		Path:      "github.com/example/testpkg",
		Types:     make(map[string]*reflect.Type),
		Functions: []*reflect.Function{},
		Constants: []*reflect.Constant{},
	}

	// Add multiple types
	typeNames := []string{"Type1", "Type2", "Type3"}
	for _, name := range typeNames {
		pkg.Types[name] = &reflect.Type{
			Name: name,
			Kind: reflect.TypeKindStruct,
			Doc:  name + " is a test type.",
		}
	}

	builder := NewBuilder()
	index, err := builder.Build(pkg)

	require.NoError(t, err)
	require.NotNil(t, index)

	// Verify all types are documented
	for _, name := range typeNames {
		key := "github.com/example/testpkg." + name
		entry, exists := index.Entries[key]
		require.True(t, exists, "All exported types should be documented, missing: %s", name)
		assert.Equal(t, KindType, entry.Kind)
	}
}

// TestBuilderByPackageIndexConsistency verifies ByPackage index stays consistent with Entries
func TestBuilderByPackageIndexConsistency(t *testing.T) {
	pkg := &reflect.Package{
		Name:      "testpkg",
		Path:      "github.com/example/testpkg",
		Types:     make(map[string]*reflect.Type),
		Functions: []*reflect.Function{},
		Constants: []*reflect.Constant{},
	}

	// Add various entries
	pkg.Types["TypeA"] = &reflect.Type{
		Name: "TypeA",
		Kind: reflect.TypeKindStruct,
		Doc:  "TypeA is a test type.",
	}

	pkg.Functions = append(pkg.Functions, &reflect.Function{
		Name: "FuncA",
		Doc:  "FuncA does something.",
	})

	pkg.Constants = append(pkg.Constants, &reflect.Constant{
		Name: "ConstA",
		Doc:  "ConstA is a constant.",
	})

	builder := NewBuilder()
	index, err := builder.Build(pkg)

	require.NoError(t, err)
	require.NotNil(t, index)

	// Check ByPackage consistency
	pkgPath := "github.com/example/testpkg"
	require.NotEqual(t, ".", pkgPath, "Package path should be full import path, not '.'")

	byPkgEntries, exists := index.ByPackage[pkgPath]
	require.True(t, exists, "ByPackage should have entry for package path %q", pkgPath)

	// Verify all entries in ByPackage exist in Entries
	for _, entryKey := range byPkgEntries {
		entry, exists := index.Entries[entryKey]
		require.True(t, exists, "Entry %q listed in ByPackage should exist in Entries", entryKey)
		assert.Equal(t, pkgPath, entry.PackagePath, "Entry should have matching PackagePath")
	}

	// Verify count is correct
	expectedCount := 3 // 1 type + 1 function + 1 constant
	assert.Equal(t, expectedCount, len(byPkgEntries), "ByPackage should have correct number of entries")
}

// TestBuilderByKindIndexConsistency verifies ByKind index stays consistent with Entries
func TestBuilderByKindIndexConsistency(t *testing.T) {
	pkg := &reflect.Package{
		Name:      "testpkg",
		Path:      "github.com/example/testpkg",
		Types:     make(map[string]*reflect.Type),
		Functions: []*reflect.Function{},
		Constants: []*reflect.Constant{},
	}

	pkg.Types["TestType"] = &reflect.Type{
		Name: "TestType",
		Kind: reflect.TypeKindStruct,
		Doc:  "TestType is a test type.",
	}

	pkg.Functions = append(pkg.Functions, &reflect.Function{
		Name: "TestFunc",
		Doc:  "TestFunc is a test function.",
	})

	builder := NewBuilder()
	index, err := builder.Build(pkg)

	require.NoError(t, err)
	require.NotNil(t, index)

	// Check ByKind entries
	typeEntries, typeExists := index.ByKind[string(KindType)]
	funcEntries, funcExists := index.ByKind[string(KindFunction)]

	require.True(t, typeExists, "ByKind should have KindType entries")
	require.True(t, funcExists, "ByKind should have KindFunction entries")

	// Verify entries exist and have matching kinds
	for _, entryKey := range typeEntries {
		entry, exists := index.Entries[entryKey]
		require.True(t, exists, "Entry %q listed in ByKind[type] should exist", entryKey)
		assert.Equal(t, KindType, entry.Kind, "Entry in ByKind[type] should have KindType")
	}

	for _, entryKey := range funcEntries {
		entry, exists := index.Entries[entryKey]
		require.True(t, exists, "Entry %q listed in ByKind[function] should exist", entryKey)
		assert.Equal(t, KindFunction, entry.Kind, "Entry in ByKind[function] should have KindFunction")
	}
}

// TestBuilderMetadataPackagePathTracking verifies metadata tracks correct package paths
func TestBuilderMetadataPackagePathTracking(t *testing.T) {
	pkg1 := &reflect.Package{
		Name:      "pkg1",
		Path:      "github.com/example/pkg1",
		Types:     make(map[string]*reflect.Type),
		Functions: []*reflect.Function{},
		Constants: []*reflect.Constant{},
	}

	pkg2 := &reflect.Package{
		Name:      "pkg2",
		Path:      "github.com/example/pkg2",
		Types:     make(map[string]*reflect.Type),
		Functions: []*reflect.Function{},
		Constants: []*reflect.Constant{},
	}

	pkg1.Types["Type1"] = &reflect.Type{
		Name: "Type1",
		Kind: reflect.TypeKindStruct,
		Doc:  "Type1 from pkg1.",
	}

	pkg2.Types["Type2"] = &reflect.Type{
		Name: "Type2",
		Kind: reflect.TypeKindStruct,
		Doc:  "Type2 from pkg2.",
	}

	builder := NewBuilder()
	index, err := builder.Build(pkg1, pkg2)

	require.NoError(t, err)
	require.NotNil(t, index)

	// Check metadata tracks correct package paths
	require.Contains(t, index.Metadata.PackagesIndexed, "github.com/example/pkg1", "Metadata should track pkg1 import path")
	require.Contains(t, index.Metadata.PackagesIndexed, "github.com/example/pkg2", "Metadata should track pkg2 import path")

	// ISSUE: Should NOT contain "." as a package path
	for _, pkgPath := range index.Metadata.PackagesIndexed {
		require.NotEqual(t, ".", pkgPath, "ISSUE: Metadata should NOT track '.' as package path, found: %q", pkgPath)
	}
}

// TestBuilderQualifiedNameFormat verifies qualified names use correct format
func TestBuilderQualifiedNameFormat(t *testing.T) {
	pkg := &reflect.Package{
		Name:      "testpkg",
		Path:      "github.com/example/testpkg",
		Types:     make(map[string]*reflect.Type),
		Functions: []*reflect.Function{},
		Constants: []*reflect.Constant{},
	}

	pkg.Types["MyType"] = &reflect.Type{
		Name: "MyType",
		Kind: reflect.TypeKindStruct,
		Doc:  "MyType is a test type.",
	}

	pkg.Functions = append(pkg.Functions, &reflect.Function{
		Name: "MyFunc",
		Doc:  "MyFunc is a test function.",
	})

	pkg.Constants = append(pkg.Constants, &reflect.Constant{
		Name: "MyConst",
		Doc:  "MyConst is a constant.",
	})

	builder := NewBuilder()
	index, err := builder.Build(pkg)

	require.NoError(t, err)
	require.NotNil(t, index)

	// Verify qualified name format is correct
	expectedTypeKey := "github.com/example/testpkg.MyType"
	expectedFuncKey := "github.com/example/testpkg.MyFunc"
	expectedConstKey := "github.com/example/testpkg.MyConst"

	typeEntry, typeExists := index.Entries[expectedTypeKey]
	funcEntry, funcExists := index.Entries[expectedFuncKey]
	constEntry, constExists := index.Entries[expectedConstKey]

	require.True(t, typeExists, "Type entry should exist with key %q", expectedTypeKey)
	require.True(t, funcExists, "Function entry should exist with key %q", expectedFuncKey)
	require.True(t, constExists, "Constant entry should exist with key %q", expectedConstKey)

	// Verify QualifiedName field matches key
	assert.Equal(t, expectedTypeKey, typeEntry.QualifiedName)
	assert.Equal(t, expectedFuncKey, funcEntry.QualifiedName)
	assert.Equal(t, expectedConstKey, constEntry.QualifiedName)

	// Verify format is NOT using "." prefix
	assert.False(t, len(typeEntry.QualifiedName) > 0 && typeEntry.QualifiedName[0] == '.',
		"ISSUE: QualifiedName should NOT start with '.'")
}

// TestBuilderPrivateItemsExcluded verifies private (unexported) items are excluded by default
func TestBuilderPrivateItemsExcluded(t *testing.T) {
	pkg := &reflect.Package{
		Name:      "testpkg",
		Path:      "github.com/example/testpkg",
		Types:     make(map[string]*reflect.Type),
		Functions: []*reflect.Function{},
		Constants: []*reflect.Constant{},
	}

	// Add both exported and unexported items
	pkg.Types["PublicType"] = &reflect.Type{
		Name: "PublicType",
		Kind: reflect.TypeKindStruct,
		Doc:  "PublicType is exported.",
	}

	pkg.Types["privateType"] = &reflect.Type{
		Name: "privateType",
		Kind: reflect.TypeKindStruct,
		Doc:  "privateType is not exported.",
	}

	builder := NewBuilder()
	index, err := builder.Build(pkg)

	require.NoError(t, err)
	require.NotNil(t, index)

	// Public type should be documented
	publicEntry, publicExists := index.Entries["github.com/example/testpkg.PublicType"]
	require.True(t, publicExists, "Exported type should be documented")
	assert.Equal(t, KindType, publicEntry.Kind)

	// Private type should NOT be documented by default
	privateEntry, privateExists := index.Entries["github.com/example/testpkg.privateType"]
	assert.False(t, privateExists, "Unexported types should NOT be documented by default")
	assert.Nil(t, privateEntry)
}

// TestBuilderPrivateItemsIncluded verifies private items can be included with option
func TestBuilderPrivateItemsIncluded(t *testing.T) {
	pkg := &reflect.Package{
		Name:      "testpkg",
		Path:      "github.com/example/testpkg",
		Types:     make(map[string]*reflect.Type),
		Functions: []*reflect.Function{},
		Constants: []*reflect.Constant{},
	}

	pkg.Types["privateType"] = &reflect.Type{
		Name: "privateType",
		Kind: reflect.TypeKindStruct,
		Doc:  "privateType is not exported.",
	}

	opts := &BuilderOptions{
		IncludePrivate: true,
	}

	builder := NewBuilderWithOptions(opts)
	index, err := builder.Build(pkg)

	require.NoError(t, err)
	require.NotNil(t, index)

	// Private type should be documented when IncludePrivate is true
	privateEntry, privateExists := index.Entries["github.com/example/testpkg.privateType"]
	require.True(t, privateExists, "Unexported types should be documented when IncludePrivate=true")
	assert.Equal(t, KindType, privateEntry.Kind)
}

// TestBuilderDocumentationContentExtraction verifies documentation content is properly extracted
func TestBuilderDocumentationContentExtraction(t *testing.T) {
	pkg := &reflect.Package{
		Name:      "testpkg",
		Path:      "github.com/example/testpkg",
		Types:     make(map[string]*reflect.Type),
		Functions: []*reflect.Function{},
		Constants: []*reflect.Constant{},
	}

	// Doc comments need to match Go doc format
	docComment := "TestType is a complex type."

	pkg.Types["TestType"] = &reflect.Type{
		Name: "TestType",
		Kind: reflect.TypeKindStruct,
		Doc:  docComment,
	}

	builder := NewBuilder()
	index, err := builder.Build(pkg)

	require.NoError(t, err)
	require.NotNil(t, index)

	entry, exists := index.Entries["github.com/example/testpkg.TestType"]
	require.True(t, exists)

	// Verify documentation is extracted
	// When a type has documentation, Summary or Details should be populated
	hasDocumentation := entry.Summary != "" || entry.Details != ""
	require.True(t, hasDocumentation, "ISSUE: Documentation extraction may not be working correctly - both Summary and Details are empty. Original doc: %q", docComment)
}

// TestBuilderStructFieldsDocumented verifies that struct fields are documented as separate entries
func TestBuilderStructFieldsDocumented(t *testing.T) {
	// Create a test package with a struct type containing documented fields
	pkg := &reflect.Package{
		Name:      "testpkg",
		Path:      "github.com/example/testpkg",
		Types:     make(map[string]*reflect.Type),
		Functions: []*reflect.Function{},
		Constants: []*reflect.Constant{},
	}

	// Create a struct type with documented fields
	testStruct := &reflect.Type{
		Name: "TestStruct",
		Kind: reflect.TypeKindStruct,
		Doc:  "TestStruct is a test structure.",
		Fields: []*reflect.Field{
			{
				Name: "FieldOne",
				Type: &reflect.TypeReference{Name: "string"},
				Doc:  "FieldOne is the first field.",
			},
			{
				Name: "FieldTwo",
				Type: &reflect.TypeReference{Name: "int"},
				Doc:  "FieldTwo is the second field.",
			},
			{
				Name: "privateField",
				Type: &reflect.TypeReference{Name: "bool"},
				Doc:  "privateField is a private field.",
			},
		},
	}
	pkg.Types["TestStruct"] = testStruct

	builder := NewBuilder()
	index, err := builder.Build(pkg)

	require.NoError(t, err, "Builder should successfully build documentation")
	require.NotNil(t, index)

	// Verify the struct type entry exists
	typeEntry, exists := index.Entries["github.com/example/testpkg.TestStruct"]
	require.True(t, exists, "Struct type entry should exist")
	assert.Equal(t, KindType, typeEntry.Kind, "Type entry should have KindType")

	// Verify field entries are created
	fieldOneEntry, exists := index.Entries["github.com/example/testpkg.TestStruct.FieldOne"]
	require.True(t, exists, "Field entry should exist for FieldOne")
	assert.Equal(t, KindField, fieldOneEntry.Kind, "Field entry should have KindField")
	assert.Equal(t, "FieldOne", fieldOneEntry.LocalName)
	assert.NotEmpty(t, fieldOneEntry.Summary, "Field entry should have documentation summary")

	fieldTwoEntry, exists := index.Entries["github.com/example/testpkg.TestStruct.FieldTwo"]
	require.True(t, exists, "Field entry should exist for FieldTwo")
	assert.Equal(t, KindField, fieldTwoEntry.Kind, "Field entry should have KindField")
	assert.Equal(t, "FieldTwo", fieldTwoEntry.LocalName)
	assert.NotEmpty(t, fieldTwoEntry.Summary, "Field entry should have documentation summary")

	// Verify private field is excluded (default behavior)
	_, exists = index.Entries["github.com/example/testpkg.TestStruct.privateField"]
	assert.False(t, exists, "Private field should not be documented by default")

	// Verify ByKind is organized correctly
	fieldEntries, exists := index.ByKind[string(KindField)]
	require.True(t, exists, "ByKind should have entries for KindField")
	assert.Greater(t, len(fieldEntries), 0, "Should have at least one field entry in ByKind")

	// Verify ByPackage includes field entries
	pkgEntries, exists := index.ByPackage["github.com/example/testpkg"]
	require.True(t, exists, "ByPackage should have entries for the package")
	assert.GreaterOrEqual(t, len(pkgEntries), 3, "ByPackage should include struct type + field entries")

	t.Logf("✓ Struct fields are properly documented: 2 field entries created")
}

// TestBuilderStructFieldsWithIncludePrivate verifies private fields are documented when IncludePrivate option is enabled
func TestBuilderStructFieldsWithIncludePrivate(t *testing.T) {
	pkg := &reflect.Package{
		Name:      "testpkg",
		Path:      "github.com/example/testpkg",
		Types:     make(map[string]*reflect.Type),
		Functions: []*reflect.Function{},
		Constants: []*reflect.Constant{},
	}

	testStruct := &reflect.Type{
		Name: "TestStruct",
		Kind: reflect.TypeKindStruct,
		Doc:  "TestStruct is a test structure.",
		Fields: []*reflect.Field{
			{
				Name: "PublicField",
				Type: &reflect.TypeReference{Name: "string"},
				Doc:  "PublicField is public.",
			},
			{
				Name: "privateField",
				Type: &reflect.TypeReference{Name: "int"},
				Doc:  "privateField is private.",
			},
		},
	}
	pkg.Types["TestStruct"] = testStruct

	// Build with IncludePrivate option enabled
	opts := DefaultBuilderOptions()
	opts.IncludePrivate = true
	builder := NewBuilderWithOptions(opts)
	index, err := builder.Build(pkg)

	require.NoError(t, err)
	require.NotNil(t, index)

	// Verify public field exists
	_, exists := index.Entries["github.com/example/testpkg.TestStruct.PublicField"]
	require.True(t, exists, "Public field should be documented")

	// Verify private field exists with IncludePrivate enabled
	privateEntry, exists := index.Entries["github.com/example/testpkg.TestStruct.privateField"]
	require.True(t, exists, "Private field should be documented when IncludePrivate=true")
	assert.Equal(t, KindField, privateEntry.Kind)

	t.Log("✓ Private fields are documented when IncludePrivate option is enabled")
}
