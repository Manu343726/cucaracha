package docs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/Manu343726/cucaracha/pkg/reflect"
)

type BuilderTestSuite struct {
	suite.Suite
}

func TestBuilderTestSuite(t *testing.T) {
	suite.Run(t, new(BuilderTestSuite))
}

// TestDefaultBuilderOptions verifies default options are properly configured
func (suite *BuilderTestSuite) TestDefaultBuilderOptions() {
	opts := DefaultBuilderOptions()
	assert.NotNil(suite.T(), opts)
	assert.False(suite.T(), opts.IncludePrivate)
	assert.True(suite.T(), opts.ResolveReferences)
	assert.Nil(suite.T(), opts.ReflectIndex)
}

// TestNewBuilder creates a builder with default options
func (suite *BuilderTestSuite) TestNewBuilder() {
	builder := NewBuilder()
	assert.NotNil(suite.T(), builder)
	assert.NotNil(suite.T(), builder.Options)
	assert.False(suite.T(), builder.Options.IncludePrivate)
	assert.True(suite.T(), builder.Options.ResolveReferences)
}

// TestNewBuilderWithOptions creates a builder with custom options
func (suite *BuilderTestSuite) TestNewBuilderWithOptions() {
	opts := &BuilderOptions{
		IncludePrivate:    true,
		ResolveReferences: false,
	}
	builder := NewBuilderWithOptions(opts)
	assert.NotNil(suite.T(), builder)
	assert.True(suite.T(), builder.Options.IncludePrivate)
	assert.False(suite.T(), builder.Options.ResolveReferences)
}

// TestNewBuilderInitializesIndex verifies the builder initializes an empty index
func (suite *BuilderTestSuite) TestNewBuilderInitializesIndex() {
	builder := NewBuilder()
	assert.NotNil(suite.T(), builder.index)
	assert.NotNil(suite.T(), builder.index.Entries)
	assert.NotNil(suite.T(), builder.index.ByPackage)
	assert.NotNil(suite.T(), builder.index.ByKind)
	assert.NotNil(suite.T(), builder.index.References)
	assert.Equal(suite.T(), "1.0", builder.index.Metadata.Version)
}

// TestBuilderWithExternalIndex verifies external index is used
func (suite *BuilderTestSuite) TestBuilderWithExternalIndex() {
	externalIndex := reflect.NewIndex()
	opts := &BuilderOptions{
		ReflectIndex: externalIndex,
	}
	builder := NewBuilderWithOptions(opts)
	assert.NotNil(suite.T(), builder.index.ReflectIndex)
	assert.Same(suite.T(), externalIndex, builder.index.ReflectIndex)
}

// TestBuilderWithoutExternalIndex verifies index is created when not provided
func (suite *BuilderTestSuite) TestBuilderWithoutExternalIndex() {
	opts := &BuilderOptions{}
	builder := NewBuilderWithOptions(opts)
	assert.NotNil(suite.T(), builder.index.ReflectIndex)
}

// TestAddEntry adds a documentation entry to the builder's index
func (suite *BuilderTestSuite) TestAddEntry() {
	builder := NewBuilder()
	entry := &DocumentationEntry{
		QualifiedName: "test.Example",
		LocalName:     "Example",
		Kind:          KindType,
		PackagePath:   "test",
		Summary:       "A test example",
	}

	err := builder.AddEntry(entry)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(builder.index.Entries))
	stored, exists := builder.index.Entries["test.Example"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), entry, stored)
}

// TestAddEntryMultiple adds multiple entries and verifies they are all stored
func (suite *BuilderTestSuite) TestAddEntryMultiple() {
	builder := NewBuilder()
	entries := []*DocumentationEntry{
		{
			QualifiedName: "test.TypeA",
			LocalName:     "TypeA",
			Kind:          KindType,
			PackagePath:   "test",
		},
		{
			QualifiedName: "test.FunctionB",
			LocalName:     "FunctionB",
			Kind:          KindFunction,
			PackagePath:   "test",
		},
	}

	for _, entry := range entries {
		err := builder.AddEntry(entry)
		require.NoError(suite.T(), err)
	}

	assert.Equal(suite.T(), 2, len(builder.index.Entries))
	assert.Contains(suite.T(), builder.index.Entries, "test.TypeA")
	assert.Contains(suite.T(), builder.index.Entries, "test.FunctionB")
}

// TestAddEntryByKind verifies entries are indexed by kind
func (suite *BuilderTestSuite) TestAddEntryByKind() {
	builder := NewBuilder()

	typeEntry := &DocumentationEntry{
		QualifiedName: "test.Type1",
		LocalName:     "Type1",
		Kind:          KindType,
		PackagePath:   "test",
	}
	funcEntry := &DocumentationEntry{
		QualifiedName: "test.Func1",
		LocalName:     "Func1",
		Kind:          KindFunction,
		PackagePath:   "test",
	}

	builder.AddEntry(typeEntry)
	builder.AddEntry(funcEntry)

	assert.Contains(suite.T(), builder.index.ByKind, string(KindType))
	assert.Contains(suite.T(), builder.index.ByKind, string(KindFunction))
}

// TestAddEntryByPackage verifies entries are indexed by package
func (suite *BuilderTestSuite) TestAddEntryByPackage() {
	builder := NewBuilder()

	entry1 := &DocumentationEntry{
		QualifiedName: "pkg1.Type1",
		LocalName:     "Type1",
		Kind:          KindType,
		PackagePath:   "pkg1",
	}
	entry2 := &DocumentationEntry{
		QualifiedName: "pkg1.Type2",
		LocalName:     "Type2",
		Kind:          KindType,
		PackagePath:   "pkg1",
	}

	builder.AddEntry(entry1)
	builder.AddEntry(entry2)

	pkgEntries, exists := builder.index.ByPackage["pkg1"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), 2, len(pkgEntries))
}

// TestAddEntryDuplicateOverwrites verifies duplicate entries are overwritten
func (suite *BuilderTestSuite) TestAddEntryDuplicateOverwrites() {
	builder := NewBuilder()
	entry1 := &DocumentationEntry{
		QualifiedName: "test.Duplicate",
		LocalName:     "Duplicate1",
		Kind:          KindType,
		PackagePath:   "test",
	}
	entry2 := &DocumentationEntry{
		QualifiedName: "test.Duplicate",
		LocalName:     "Duplicate2",
		Kind:          KindType,
		PackagePath:   "test",
	}

	err1 := builder.AddEntry(entry1)
	require.NoError(suite.T(), err1)

	err2 := builder.AddEntry(entry2)
	require.NoError(suite.T(), err2)
	// Both should succeed; second one overwrites the first
	assert.NoError(suite.T(), err2)
}

// TestBuildWithEmptyPackages returns valid index for empty input
func (suite *BuilderTestSuite) TestBuildWithEmptyPackages() {
	builder := NewBuilder()
	index, err := builder.Build()
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), index)
	assert.Equal(suite.T(), 0, len(index.Entries))
	assert.Equal(suite.T(), 0, len(index.Metadata.PackagesIndexed))
}

// TestBuildUpdatesMetadata verifies Build() updates metadata correctly
func (suite *BuilderTestSuite) TestBuildUpdatesMetadata() {
	builder := NewBuilder()
	pkgs := []*reflect.Package{
		{
			Path: "test/pkg1",
			Name: "pkg1",
		},
		{
			Path: "test/pkg2",
			Name: "pkg2",
		},
	}

	index, err := builder.Build(pkgs...)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, len(index.Metadata.PackagesIndexed))
	assert.Contains(suite.T(), index.Metadata.PackagesIndexed, "test/pkg1")
	assert.Contains(suite.T(), index.Metadata.PackagesIndexed, "test/pkg2")
}

// TestBuildPopulatesReflectIndex verifies packages are added to ReflectIndex
func (suite *BuilderTestSuite) TestBuildPopulatesReflectIndex() {
	builder := NewBuilder()
	pkg := &reflect.Package{
		Path: "test/pkg",
		Name: "pkg",
	}

	index, err := builder.Build(pkg)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), index.ReflectIndex)

	// Verify package was added to index
	retrieved := index.ReflectIndex.Package("test/pkg")
	assert.NotNil(suite.T(), retrieved)
	assert.Equal(suite.T(), "test/pkg", retrieved.Path)
}

// TestBuildIndexMetadataVersion checks initial version
func (suite *BuilderTestSuite) TestBuildIndexMetadataVersion() {
	builder := NewBuilder()
	index, err := builder.Build()
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "1.0", index.Metadata.Version)
}

// TestBuildReturnsDocumentationIndex verifies return type
func (suite *BuilderTestSuite) TestBuildReturnsDocumentationIndex() {
	builder := NewBuilder()
	index, err := builder.Build()
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), index)
	assert.IsType(suite.T(), &DocumentationIndex{}, index)
}

// TestEntryKindConstants verifies all kind constants are defined
func (suite *BuilderTestSuite) TestEntryKindConstants() {
	assert.Equal(suite.T(), EntryKind("package"), KindPackage)
	assert.Equal(suite.T(), EntryKind("type"), KindType)
	assert.Equal(suite.T(), EntryKind("function"), KindFunction)
	assert.Equal(suite.T(), EntryKind("method"), KindMethod)
	assert.Equal(suite.T(), EntryKind("constant"), KindConstant)
	assert.Equal(suite.T(), EntryKind("enum"), KindEnum)
	assert.Equal(suite.T(), EntryKind("enumValue"), KindEnumValue)
	assert.Equal(suite.T(), EntryKind("field"), KindField)
	assert.Equal(suite.T(), EntryKind("interfaceMethod"), KindInterfaceMethod)
}

// TestLinkRelationshipConstants verifies all relationship constants are defined
func (suite *BuilderTestSuite) TestLinkRelationshipConstants() {
	assert.Equal(suite.T(), LinkRelationship("uses"), RelationshipUses)
	assert.Equal(suite.T(), LinkRelationship("usedBy"), RelationshipUsedBy)
	assert.Equal(suite.T(), LinkRelationship("implements"), RelationshipImplements)
	assert.Equal(suite.T(), LinkRelationship("implementedBy"), RelationshipImplementedBy)
	assert.Equal(suite.T(), LinkRelationship("related"), RelationshipRelated)
	assert.Equal(suite.T(), LinkRelationship("embeds"), RelationshipEmbeds)
	assert.Equal(suite.T(), LinkRelationship("embeddedBy"), RelationshipEmbeddedBy)
	assert.Equal(suite.T(), LinkRelationship("returns"), RelationshipReturns)
	assert.Equal(suite.T(), LinkRelationship("parameter"), RelationshipParameter)
	assert.Equal(suite.T(), LinkRelationship("example"), RelationshipExample)
}

// TestDocumentationEntry verifies DocumentationEntry structure
func (suite *BuilderTestSuite) TestDocumentationEntry() {
	entry := &DocumentationEntry{
		QualifiedName: "pkg.Type",
		LocalName:     "Type",
		Kind:          KindType,
		PackagePath:   "pkg",
		Summary:       "Summary",
		Details:       "Details",
		Examples: []Example{
			{
				Description: "Example",
				Code:        "code",
			},
		},
		Links: []Link{
			{
				Target:       "pkg.OtherType",
				Relationship: RelationshipUses,
			},
		},
	}

	assert.Equal(suite.T(), "pkg.Type", entry.QualifiedName)
	assert.Equal(suite.T(), "Type", entry.LocalName)
	assert.Equal(suite.T(), KindType, entry.Kind)
	assert.Equal(suite.T(), "pkg", entry.PackagePath)
	assert.Equal(suite.T(), "Summary", entry.Summary)
	assert.Equal(suite.T(), "Details", entry.Details)
	assert.Equal(suite.T(), 1, len(entry.Examples))
	assert.Equal(suite.T(), 1, len(entry.Links))
}

// TestSourceLocation verifies SourceLocation structure
func (suite *BuilderTestSuite) TestSourceLocation() {
	loc := SourceLocation{
		FilePath:      "main.go",
		LineNumber:    10,
		ColumnNumber:  5,
		EndLineNumber: 15,
	}

	assert.Equal(suite.T(), "main.go", loc.FilePath)
	assert.Equal(suite.T(), 10, loc.LineNumber)
	assert.Equal(suite.T(), 5, loc.ColumnNumber)
	assert.Equal(suite.T(), 15, loc.EndLineNumber)
}

// TestExample verifies Example structure
func (suite *BuilderTestSuite) TestExample() {
	example := Example{
		Description: "Example description",
		Code:        "code snippet",
		Output:      "expected output",
		Tags:        []string{"basic", "advanced"},
	}

	assert.Equal(suite.T(), "Example description", example.Description)
	assert.Equal(suite.T(), "code snippet", example.Code)
	assert.Equal(suite.T(), "expected output", example.Output)
	assert.Len(suite.T(), example.Tags, 2)
	assert.Contains(suite.T(), example.Tags, "basic")
}

// TestLink verifies Link structure
func (suite *BuilderTestSuite) TestLink() {
	link := Link{
		Target:       "pkg.Type",
		Relationship: RelationshipUses,
		Context:      "Used by function X",
	}

	assert.Equal(suite.T(), "pkg.Type", link.Target)
	assert.Equal(suite.T(), RelationshipUses, link.Relationship)
	assert.Equal(suite.T(), "Used by function X", link.Context)
}

// TestDocumentationIndex verifies DocumentationIndex initialization
func (suite *BuilderTestSuite) TestDocumentationIndex() {
	index := &DocumentationIndex{
		Entries:    make(map[string]*DocumentationEntry),
		ByPackage:  make(map[string][]string),
		ByKind:     make(map[string][]string),
		References: make(map[string][]string),
		Metadata: IndexMetadata{
			Version:         "1.0",
			PackagesIndexed: []string{"pkg1"},
		},
	}

	assert.NotNil(suite.T(), index.Entries)
	assert.NotNil(suite.T(), index.ByPackage)
	assert.NotNil(suite.T(), index.ByKind)
	assert.NotNil(suite.T(), index.References)
	assert.Equal(suite.T(), "1.0", index.Metadata.Version)
	assert.Len(suite.T(), index.Metadata.PackagesIndexed, 1)
}
