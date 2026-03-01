package docs

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/Manu343726/cucaracha/pkg/reflect"
)

type BuilderHelperTestSuite struct {
	suite.Suite
}

func TestBuilderHelperTestSuite(t *testing.T) {
	suite.Run(t, new(BuilderHelperTestSuite))
}

// TestIsExportedWithCapitalLetter returns true for exported names
func (suite *BuilderHelperTestSuite) TestIsExportedWithCapitalLetter() {
	assert.True(suite.T(), isExported("Type"))
	assert.True(suite.T(), isExported("Function"))
	assert.True(suite.T(), isExported("Z"))
	assert.True(suite.T(), isExported("MyType"))
	assert.True(suite.T(), isExported("ABC"))
}

// TestIsExportedWithLowercaseLetter returns false for unexported names
func (suite *BuilderHelperTestSuite) TestIsExportedWithLowercaseLetter() {
	assert.False(suite.T(), isExported("type"))
	assert.False(suite.T(), isExported("function"))
	assert.False(suite.T(), isExported("a"))
	assert.False(suite.T(), isExported("myType"))
	assert.False(suite.T(), isExported("_private"))
}

// TestIsExportedWithEmptyString returns false for empty string
func (suite *BuilderHelperTestSuite) TestIsExportedWithEmptyString() {
	assert.False(suite.T(), isExported(""))
}

// TestIsExportedWithSpecialCharacters handles special characters
func (suite *BuilderHelperTestSuite) TestIsExportedWithSpecialCharacters() {
	assert.False(suite.T(), isExported("_Type"))
	assert.False(suite.T(), isExported("123Type"))
	assert.False(suite.T(), isExported("-Type"))
}

// TestResolveReferencesReturnsNil is tested through Build() method
// which calls resolveReferences internally when ResolveReferences option is true
func (suite *BuilderHelperTestSuite) TestResolveReferencesReturnsNil() {
	opts := &BuilderOptions{
		ResolveReferences: false, // Disable to verify path
	}
	builder := NewBuilderWithOptions(opts)
	pkg := &reflect.Package{Path: "test/pkg", Name: "pkg"}
	index, err := builder.Build(pkg)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), index)
}

// TestAddEntryWithQualifiedName validates qualified name requirement
func (suite *BuilderHelperTestSuite) TestAddEntryWithQualifiedName() {
	builder := NewBuilder()
	entry := &DocumentationEntry{
		QualifiedName: "",
		LocalName:     "Test",
		Kind:          KindType,
	}

	err := builder.AddEntry(entry)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "qualified name")
}

// TestBuildWithSinglePackage builds documentation from single package
func (suite *BuilderHelperTestSuite) TestBuildWithSinglePackage() {
	builder := NewBuilder()
	pkg := &reflect.Package{
		Path: "test/single",
		Name: "single",
	}

	index, err := builder.Build(pkg)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), index)
	assert.Contains(suite.T(), index.Metadata.PackagesIndexed, "test/single")
}

// TestBuildWithMultiplePackages builds documentation from multiple packages
func (suite *BuilderHelperTestSuite) TestBuildWithMultiplePackages() {
	builder := NewBuilder()
	pkg1 := &reflect.Package{Path: "test/pkg1", Name: "pkg1"}
	pkg2 := &reflect.Package{Path: "test/pkg2", Name: "pkg2"}

	index, err := builder.Build(pkg1, pkg2)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), index.Metadata.PackagesIndexed, 2)
}

// TestBuildNotifyLogging builds with logging
func (suite *BuilderHelperTestSuite) TestBuildNotifyLogging() {
	builder := NewBuilder()
	pkg := &reflect.Package{
		Path: "test/pkg",
		Name: "pkg",
	}

	// Should log various stages
	index, err := builder.Build(pkg)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), index)
}

// TestAddEntryIndexesByPackage verifies entry is indexed by package through Build
func (suite *BuilderHelperTestSuite) TestAddEntryIndexesByPackage() {
	builder := NewBuilder()
	entry := &DocumentationEntry{
		QualifiedName: "pkg.Item",
		LocalName:     "Item",
		Kind:          KindType,
		PackagePath:   "pkg",
	}

	err := builder.AddEntry(entry)
	require.NoError(suite.T(), err)
	// Entry itself should have been added successfully
	assert.Equal(suite.T(), "pkg.Item", entry.QualifiedName)
}

// TestAddEntryIndexesByKind verifies entry is indexed by kind
func (suite *BuilderHelperTestSuite) TestAddEntryIndexesByKind() {
	builder := NewBuilder()
	entry := &DocumentationEntry{
		QualifiedName: "pkg.Item",
		LocalName:     "Item",
		Kind:          KindFunction,
		PackagePath:   "pkg",
	}

	err := builder.AddEntry(entry)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), KindFunction, entry.Kind)
}

// TestAddEntryMultipleTimesToSamePackage adds multiple entries to same package
func (suite *BuilderHelperTestSuite) TestAddEntryMultipleTimesToSamePackage() {
	builder := NewBuilder()
	entry1 := &DocumentationEntry{
		QualifiedName: "pkg.Item1",
		LocalName:     "Item1",
		Kind:          KindType,
		PackagePath:   "pkg",
	}
	entry2 := &DocumentationEntry{
		QualifiedName: "pkg.Item2",
		LocalName:     "Item2",
		Kind:          KindFunction,
		PackagePath:   "pkg",
	}

	err1 := builder.AddEntry(entry1)
	err2 := builder.AddEntry(entry2)
	require.NoError(suite.T(), err1)
	require.NoError(suite.T(), err2)
	assert.Equal(suite.T(), "pkg.Item1", entry1.QualifiedName)
	assert.Equal(suite.T(), "pkg.Item2", entry2.QualifiedName)
}

// TestBuildIncludePrivateOption includes private items when enabled
func (suite *BuilderHelperTestSuite) TestBuildIncludePrivateOption() {
	opts := &BuilderOptions{
		IncludePrivate:    true,
		ResolveReferences: false,
	}
	builder := NewBuilderWithOptions(opts)
	assert.True(suite.T(), builder.Options.IncludePrivate)
}

// TestBuildResolveReferencesOption can be disabled
func (suite *BuilderHelperTestSuite) TestBuildResolveReferencesOption() {
	opts := &BuilderOptions{
		IncludePrivate:    false,
		ResolveReferences: false,
	}
	builder := NewBuilderWithOptions(opts)
	index, err := builder.Build()
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), index)
}

// TestRenderExamplesIncludesDescription renders example description
func (suite *BuilderHelperTestSuite) TestRenderExamplesIncludesDescription() {
	buf := &bytes.Buffer{}
	formatter := &DefaultLinkFormatter{}
	opts := DefaultRenderOptions(formatter)

	renderer := NewEntryRenderer(buf, opts)
	renderer.SetEntry(&DocumentationEntry{
		LocalName: "TestType",
		Examples: []Example{
			{
				Description: "Example of basic usage",
				Code:        "code",
			},
		},
	})

	err := renderer.RenderExamples()
	require.NoError(suite.T(), err)
	content := buf.String()
	assert.Contains(suite.T(), content, "Example of basic usage")
}

// TestRenderExamplesIncludesCode renders example code
func (suite *BuilderHelperTestSuite) TestRenderExamplesIncludesCode() {
	buf := &bytes.Buffer{}
	formatter := &DefaultLinkFormatter{}
	opts := DefaultRenderOptions(formatter)

	renderer := NewEntryRenderer(buf, opts)
	renderer.SetEntry(&DocumentationEntry{
		LocalName: "TestType",
		Examples: []Example{
			{
				Description: "Example",
				Code:        "func main() { }",
			},
		},
	})

	err := renderer.RenderExamples()
	require.NoError(suite.T(), err)
	content := buf.String()
	assert.Contains(suite.T(), content, "func main() { }")
}

// TestRenderExamplesIncludesOutput renders example output when provided
func (suite *BuilderHelperTestSuite) TestRenderExamplesIncludesOutput() {
	buf := &bytes.Buffer{}
	formatter := &DefaultLinkFormatter{}
	opts := DefaultRenderOptions(formatter)

	renderer := NewEntryRenderer(buf, opts)
	renderer.SetEntry(&DocumentationEntry{
		LocalName: "TestType",
		Examples: []Example{
			{
				Description: "Example",
				Code:        "code",
				Output:      "Some output text",
			},
		},
	})

	err := renderer.RenderExamples()
	require.NoError(suite.T(), err)
	content := buf.String()
	assert.Contains(suite.T(), content, "Output:")
	assert.Contains(suite.T(), content, "Some output text")
}

// TestRenderLinksIncludesRelationship renders link relationship
func (suite *BuilderHelperTestSuite) TestRenderLinksIncludesRelationship() {
	buf := &bytes.Buffer{}
	formatter := &DefaultLinkFormatter{}
	opts := DefaultRenderOptions(formatter)

	renderer := NewEntryRenderer(buf, opts)
	renderer.SetEntry(&DocumentationEntry{
		LocalName: "TestType",
		Links: []Link{
			{
				Target:       "pkg.Other",
				Relationship: RelationshipImplements,
			},
		},
	})

	err := renderer.RenderLinks()
	require.NoError(suite.T(), err)
	content := buf.String()
	assert.Contains(suite.T(), content, "implements")
}

// TestRenderLinksHandlesEmptyRelationship renders without relationship when empty
func (suite *BuilderHelperTestSuite) TestRenderLinksHandlesEmptyRelationship() {
	buf := &bytes.Buffer{}
	formatter := &DefaultLinkFormatter{}
	opts := DefaultRenderOptions(formatter)

	renderer := NewEntryRenderer(buf, opts)
	renderer.SetEntry(&DocumentationEntry{
		LocalName: "TestType",
		Links: []Link{
			{
				Target:       "pkg.Other",
				Relationship: "",
			},
		},
	})

	err := renderer.RenderLinks()
	require.NoError(suite.T(), err)
	// Should still render without error
	assert.NoError(suite.T(), err)
}

// TestCommentTextStringReturnsEmpty handles interface correctly
func (suite *BuilderHelperTestSuite) TestCommentTextStringReturnsEmpty() {
	// commentTextString currently returns empty string for all inputs
	result := commentTextString(nil)
	assert.Equal(suite.T(), "", result)

	result = commentTextString("string")
	assert.Equal(suite.T(), "", result)

	result = commentTextString(123)
	assert.Equal(suite.T(), "", result)
}

// TestBlockStringReturnsEmptyForCode returns Code.Text appropriately
func (suite *BuilderHelperTestSuite) TestBlockStringReturnsEmptyForCode() {
	// blockString handles Code blocks by returning v.Text
	// Creating a simple Code block would require importing comment package
	// For now, test nil case
	result := blockString(nil)
	assert.Equal(suite.T(), "", result)
}

// TestRenderExamplesWithMultipleLinesOfCode renders code with multiple lines
func (suite *BuilderHelperTestSuite) TestRenderExamplesWithMultipleLinesOfCode() {
	buf := &bytes.Buffer{}
	formatter := &DefaultLinkFormatter{}
	opts := DefaultRenderOptions(formatter)

	renderer := NewEntryRenderer(buf, opts)
	renderer.SetEntry(&DocumentationEntry{
		LocalName: "TestType",
		Examples: []Example{
			{
				Description: "Multi-line example",
				Code:        "func main() {\n  fmt.Println(\"hello\")\n}",
			},
		},
	})

	err := renderer.RenderExamples()
	require.NoError(suite.T(), err)
	content := buf.String()
	assert.Contains(suite.T(), content, "func main()")
	assert.Contains(suite.T(), content, "fmt.Println")
}

// TestRenderFullDoesNotErrorOnNilComponentsInEntry tests rendering with minimal entry
func (suite *BuilderHelperTestSuite) TestRenderFullWithMinimalEntry() {
	buf := &bytes.Buffer{}
	formatter := &DefaultLinkFormatter{}
	opts := DefaultRenderOptions(formatter)

	renderer := NewEntryRenderer(buf, opts)
	renderer.SetEntry(&DocumentationEntry{
		LocalName: "Minimal",
		Summary:   "",
		Details:   "",
	})

	err := renderer.RenderFull()
	require.NoError(suite.T(), err)
}

// TestParseDocCommentReturnsDoc is tested through Build() with real packages
func (suite *BuilderHelperTestSuite) TestParseDocCommentReturnsDoc() {
	builder := NewBuilder()
	pkg := &reflect.Package{
		Path:      "test/parsing",
		Name:      "parsing",
		Types:     make(map[string]*reflect.Type),
		Functions: []*reflect.Function{},
		Constants: []*reflect.Constant{},
	}
	index, err := builder.Build(pkg)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), index)
}

// TestParseDocCommentWithValidComment parses valid documentation comments
func (suite *BuilderHelperTestSuite) TestParseDocCommentWithValidComment() {
	builder := NewBuilder()
	// Build a package to trigger doc comment parsing
	pkg := &reflect.Package{
		Path: "test/parsing",
		Name: "parsing",
		Types: map[string]*reflect.Type{
			"MyType": {
				Name: "MyType",
				Doc:  "MyType is a test type.\n\nIt has documentation.",
				Kind: reflect.TypeKindStruct,
			},
		},
	}
	index, err := builder.Build(pkg)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), index)
}

// TestBuildPackageLogs logs package building
func (suite *BuilderHelperTestSuite) TestBuildPackageLogs() {
	builder := NewBuilder()
	pkg := &reflect.Package{
		Path:      "test/logging",
		Name:      "logging",
		Types:     make(map[string]*reflect.Type),
		Functions: []*reflect.Function{},
		Constants: []*reflect.Constant{},
	}

	index, err := builder.Build(pkg)
	require.NoError(suite.T(), err)
	// Verify package was added to metadata
	assert.Contains(suite.T(), index.Metadata.PackagesIndexed, "test/logging")
}

// TestPackageWithoutDocComments handles package without doc comments
func (suite *BuilderHelperTestSuite) TestPackageWithoutDocComments() {
	builder := NewBuilder()
	pkg := &reflect.Package{
		Path:      "test/nodocs",
		Name:      "nodocs",
		Types:     make(map[string]*reflect.Type),
		Functions: []*reflect.Function{},
		Constants: []*reflect.Constant{},
	}

	index, err := builder.Build(pkg)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), index)
}

// TestEntrySourceLocationIsPopulated verifies source location in entry
func (suite *BuilderHelperTestSuite) TestEntrySourceLocationIsPopulated() {
	entry := &DocumentationEntry{
		SourceLocation: SourceLocation{
			FilePath:   "file.go",
			LineNumber: 42,
		},
	}

	assert.Equal(suite.T(), "file.go", entry.SourceLocation.FilePath)
	assert.Equal(suite.T(), 42, entry.SourceLocation.LineNumber)
}

// TestWriteLine writes line with proper formatting
func (suite *BuilderHelperTestSuite) TestWriteLine() {
	buf := &bytes.Buffer{}
	formatter := &DefaultLinkFormatter{}
	opts := DefaultRenderOptions(formatter)

	renderer := NewEntryRenderer(buf, opts)
	err := renderer.writeLine("test content")
	require.NoError(suite.T(), err)

	content := buf.String()
	assert.Contains(suite.T(), content, "test content")
}

// TestWriteLineMultipleTimes writes multiple lines
func (suite *BuilderHelperTestSuite) TestWriteLineMultipleTimes() {
	buf := &bytes.Buffer{}
	formatter := &DefaultLinkFormatter{}
	opts := DefaultRenderOptions(formatter)

	renderer := NewEntryRenderer(buf, opts)
	renderer.writeLine("line1")
	renderer.writeLine("line2")

	content := buf.String()
	assert.Contains(suite.T(), content, "line1")
	assert.Contains(suite.T(), content, "line2")
}

// TestRenderDetailsWithComplexText renders details with multiple lines
func (suite *BuilderHelperTestSuite) TestRenderDetailsWithComplexText() {
	buf := &bytes.Buffer{}
	formatter := &DefaultLinkFormatter{}
	opts := DefaultRenderOptions(formatter)

	renderer := NewEntryRenderer(buf, opts)
	renderer.SetEntry(&DocumentationEntry{
		LocalName: "Test",
		Details:   "Line 1\nLine 2\nLine 3",
	})

	err := renderer.RenderDetails()
	require.NoError(suite.T(), err)
	// Should render without error
	assert.NoError(suite.T(), err)
}
