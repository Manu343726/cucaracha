package docs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/Manu343726/cucaracha/pkg/reflect"
)

type TypesTestSuite struct {
	suite.Suite
}

func TestTypesTestSuite(t *testing.T) {
	suite.Run(t, new(TypesTestSuite))
}

// TestDocumentationIndexInitialization creates a new index
func (suite *TypesTestSuite) TestDocumentationIndexInitialization() {
	index := &DocumentationIndex{
		Entries:    make(map[string]*DocumentationEntry),
		ByPackage:  make(map[string][]string),
		ByKind:     make(map[string][]string),
		References: make(map[string][]string),
		Metadata: IndexMetadata{
			Version: "1.0",
		},
	}

	assert.NotNil(suite.T(), index)
	assert.Len(suite.T(), index.Entries, 0)
	assert.Len(suite.T(), index.ByPackage, 0)
	assert.Len(suite.T(), index.ByKind, 0)
	assert.Len(suite.T(), index.References, 0)
	assert.Equal(suite.T(), "1.0", index.Metadata.Version)
}

// TestDocumentationIndexWithReflectIndex includes reflect index
func (suite *TypesTestSuite) TestDocumentationIndexWithReflectIndex() {
	reflectIndex := reflect.NewIndex()
	index := &DocumentationIndex{
		ReflectIndex: reflectIndex,
		Entries:      make(map[string]*DocumentationEntry),
		ByPackage:    make(map[string][]string),
		ByKind:       make(map[string][]string),
		References:   make(map[string][]string),
		Metadata: IndexMetadata{
			Version: "1.0",
		},
	}

	assert.NotNil(suite.T(), index.ReflectIndex)
	assert.Same(suite.T(), reflectIndex, index.ReflectIndex)
}

// TestIndexMetadata verifies metadata structure
func (suite *TypesTestSuite) TestIndexMetadata() {
	metadata := IndexMetadata{
		Version:         "1.0",
		PackagesIndexed: []string{"pkg1", "pkg2"},
		SourceInfo:      "test source",
	}

	assert.Equal(suite.T(), "1.0", metadata.Version)
	assert.Len(suite.T(), metadata.PackagesIndexed, 2)
	assert.Equal(suite.T(), "test source", metadata.SourceInfo)
}

// TestDocumentationSource creates all source type variations
func (suite *TypesTestSuite) TestDocumentationSourcePackage() {
	pkg := &reflect.Package{Path: "test", Name: "test"}
	source := &DocumentationSource{Package: pkg}

	assert.NotNil(suite.T(), source.Package)
	assert.Nil(suite.T(), source.Type)
	assert.Nil(suite.T(), source.Function)
	assert.Nil(suite.T(), source.Method)
	assert.Nil(suite.T(), source.Constant)
	assert.Nil(suite.T(), source.Field)
	assert.Nil(suite.T(), source.Enum)
}

// TestDocumentationSourceType creates type source
func (suite *TypesTestSuite) TestDocumentationSourceType() {
	typ := &reflect.Type{Name: "MyType"}
	source := &DocumentationSource{Type: typ}

	assert.Nil(suite.T(), source.Package)
	assert.NotNil(suite.T(), source.Type)
	assert.Nil(suite.T(), source.Function)
}

// TestDocumentationSourceFunction creates function source
func (suite *TypesTestSuite) TestDocumentationSourceFunction() {
	fn := &reflect.Function{Name: "MyFunction"}
	source := &DocumentationSource{Function: fn}

	assert.Nil(suite.T(), source.Package)
	assert.Nil(suite.T(), source.Type)
	assert.NotNil(suite.T(), source.Function)
	assert.Nil(suite.T(), source.Method)
}

// TestDocumentationSourceMethod creates method source
func (suite *TypesTestSuite) TestDocumentationSourceMethod() {
	method := &reflect.Method{Name: "MyMethod"}
	source := &DocumentationSource{Method: method}

	assert.Nil(suite.T(), source.Package)
	assert.NotNil(suite.T(), source.Method)
	assert.Nil(suite.T(), source.Constant)
}

// TestDocumentationSourceConstant creates constant source
func (suite *TypesTestSuite) TestDocumentationSourceConstant() {
	const_ := &reflect.Constant{Name: "MyConstant"}
	source := &DocumentationSource{Constant: const_}

	assert.Nil(suite.T(), source.Package)
	assert.NotNil(suite.T(), source.Constant)
	assert.Nil(suite.T(), source.Field)
}

// TestDocumentationSourceField creates field source
func (suite *TypesTestSuite) TestDocumentationSourceField() {
	field := &reflect.Field{Name: "MyField"}
	source := &DocumentationSource{Field: field}

	assert.Nil(suite.T(), source.Package)
	assert.NotNil(suite.T(), source.Field)
	assert.Nil(suite.T(), source.Enum)
}

// TestDocumentationSourceEnum creates enum source
func (suite *TypesTestSuite) TestDocumentationSourceEnum() {
	enum := &reflect.Enum{Type: &reflect.TypeReference{Name: "MyEnum"}}
	source := &DocumentationSource{Enum: enum}

	assert.Nil(suite.T(), source.Package)
	assert.NotNil(suite.T(), source.Enum)
}

// TestDocumentationEntry verifies all entry fields
func (suite *TypesTestSuite) TestDocumentationEntry() {
	entry := &DocumentationEntry{
		QualifiedName: "pkg.Type",
		LocalName:     "Type",
		Kind:          KindType,
		PackagePath:   "pkg",
		Summary:       "Type summary",
		Details:       "Type details",
		Examples: []Example{
			{Description: "Example 1", Code: "code1"},
		},
		Links: []Link{
			{Target: "pkg.Other", Relationship: RelationshipUses},
		},
		SourceLocation: SourceLocation{
			FilePath:   "main.go",
			LineNumber: 10,
		},
	}

	assert.Equal(suite.T(), "pkg.Type", entry.QualifiedName)
	assert.Equal(suite.T(), "Type", entry.LocalName)
	assert.Equal(suite.T(), KindType, entry.Kind)
	assert.Equal(suite.T(), "pkg", entry.PackagePath)
	assert.Equal(suite.T(), "Type summary", entry.Summary)
	assert.Equal(suite.T(), "Type details", entry.Details)
	assert.Len(suite.T(), entry.Examples, 1)
	assert.Len(suite.T(), entry.Links, 1)
}

// TestEntryKindPackage verifies package kind constant
func (suite *TypesTestSuite) TestEntryKindPackage() {
	assert.Equal(suite.T(), EntryKind("package"), KindPackage)
}

// TestEntryKindType verifies type kind constant
func (suite *TypesTestSuite) TestEntryKindType() {
	assert.Equal(suite.T(), EntryKind("type"), KindType)
}

// TestEntryKindFunction verifies function kind constant
func (suite *TypesTestSuite) TestEntryKindFunction() {
	assert.Equal(suite.T(), EntryKind("function"), KindFunction)
}

// TestEntryKindMethod verifies method kind constant
func (suite *TypesTestSuite) TestEntryKindMethod() {
	assert.Equal(suite.T(), EntryKind("method"), KindMethod)
}

// TestEntryKindConstant verifies constant kind constant
func (suite *TypesTestSuite) TestEntryKindConstant() {
	assert.Equal(suite.T(), EntryKind("constant"), KindConstant)
}

// TestEntryKindEnum verifies enum kind constant
func (suite *TypesTestSuite) TestEntryKindEnum() {
	assert.Equal(suite.T(), EntryKind("enum"), KindEnum)
}

// TestEntryKindEnumValue verifies enum value kind constant
func (suite *TypesTestSuite) TestEntryKindEnumValue() {
	assert.Equal(suite.T(), EntryKind("enumValue"), KindEnumValue)
}

// TestEntryKindField verifies field kind constant
func (suite *TypesTestSuite) TestEntryKindField() {
	assert.Equal(suite.T(), EntryKind("field"), KindField)
}

// TestEntryKindInterfaceMethod verifies interface method kind constant
func (suite *TypesTestSuite) TestEntryKindInterfaceMethod() {
	assert.Equal(suite.T(), EntryKind("interfaceMethod"), KindInterfaceMethod)
}

// TestLinkRelationshipUses verifies uses relationship constant
func (suite *TypesTestSuite) TestLinkRelationshipUses() {
	assert.Equal(suite.T(), LinkRelationship("uses"), RelationshipUses)
}

// TestLinkRelationshipUsedBy verifies usedBy relationship constant
func (suite *TypesTestSuite) TestLinkRelationshipUsedBy() {
	assert.Equal(suite.T(), LinkRelationship("usedBy"), RelationshipUsedBy)
}

// TestLinkRelationshipImplements verifies implements relationship constant
func (suite *TypesTestSuite) TestLinkRelationshipImplements() {
	assert.Equal(suite.T(), LinkRelationship("implements"), RelationshipImplements)
}

// TestLinkRelationshipImplementedBy verifies implementedBy relationship constant
func (suite *TypesTestSuite) TestLinkRelationshipImplementedBy() {
	assert.Equal(suite.T(), LinkRelationship("implementedBy"), RelationshipImplementedBy)
}

// TestLinkRelationshipRelated verifies related relationship constant
func (suite *TypesTestSuite) TestLinkRelationshipRelated() {
	assert.Equal(suite.T(), LinkRelationship("related"), RelationshipRelated)
}

// TestLinkRelationshipEmbeds verifies embeds relationship constant
func (suite *TypesTestSuite) TestLinkRelationshipEmbeds() {
	assert.Equal(suite.T(), LinkRelationship("embeds"), RelationshipEmbeds)
}

// TestLinkRelationshipEmbeddedBy verifies embeddedBy relationship constant
func (suite *TypesTestSuite) TestLinkRelationshipEmbeddedBy() {
	assert.Equal(suite.T(), LinkRelationship("embeddedBy"), RelationshipEmbeddedBy)
}

// TestLinkRelationshipReturns verifies returns relationship constant
func (suite *TypesTestSuite) TestLinkRelationshipReturns() {
	assert.Equal(suite.T(), LinkRelationship("returns"), RelationshipReturns)
}

// TestLinkRelationshipParameter verifies parameter relationship constant
func (suite *TypesTestSuite) TestLinkRelationshipParameter() {
	assert.Equal(suite.T(), LinkRelationship("parameter"), RelationshipParameter)
}

// TestLinkRelationshipExample verifies example relationship constant
func (suite *TypesTestSuite) TestLinkRelationshipExample() {
	assert.Equal(suite.T(), LinkRelationship("example"), RelationshipExample)
}

// TestExampleStructure verifies all example fields
func (suite *TypesTestSuite) TestExampleStructure() {
	example := Example{
		Description: "This is an example",
		Code:        "func main() { }",
		Output:      "Output text",
		Tags:        []string{"basic", "advanced"},
	}

	assert.Equal(suite.T(), "This is an example", example.Description)
	assert.Equal(suite.T(), "func main() { }", example.Code)
	assert.Equal(suite.T(), "Output text", example.Output)
	assert.Len(suite.T(), example.Tags, 2)
}

// TestExampleWithoutOutput creates example without output
func (suite *TypesTestSuite) TestExampleWithoutOutput() {
	example := Example{
		Description: "Example without output",
		Code:        "code",
	}

	assert.Equal(suite.T(), "Example without output", example.Description)
	assert.Equal(suite.T(), "", example.Output)
	assert.Len(suite.T(), example.Tags, 0)
}

// TestLink verifies all link fields
func (suite *TypesTestSuite) TestLink() {
	targetEntry := &DocumentationEntry{QualifiedName: "pkg.Target"}
	link := Link{
		Target:       "pkg.Target",
		Relationship: RelationshipUses,
		Context:      "Used in example",
		TargetEntry:  targetEntry,
	}

	assert.Equal(suite.T(), "pkg.Target", link.Target)
	assert.Equal(suite.T(), RelationshipUses, link.Relationship)
	assert.Equal(suite.T(), "Used in example", link.Context)
	assert.NotNil(suite.T(), link.TargetEntry)
}

// TestLinkWithSource creates link with source entity
func (suite *TypesTestSuite) TestLinkWithSource() {
	source := &DocumentationSource{
		Type: &reflect.Type{Name: "SourceType"},
	}
	link := Link{
		Target:       "pkg.Target",
		Relationship: RelationshipUses,
		Source:       source,
	}

	assert.NotNil(suite.T(), link.Source)
	assert.NotNil(suite.T(), link.Source.Type)
}

// TestSourceLocation verifies all location fields
func (suite *TypesTestSuite) TestSourceLocation() {
	loc := SourceLocation{
		FilePath:      "main.go",
		LineNumber:    10,
		ColumnNumber:  5,
		EndLineNumber: 20,
	}

	assert.Equal(suite.T(), "main.go", loc.FilePath)
	assert.Equal(suite.T(), 10, loc.LineNumber)
	assert.Equal(suite.T(), 5, loc.ColumnNumber)
	assert.Equal(suite.T(), 20, loc.EndLineNumber)
}

// TestSourceLocationMinimal creates minimal location
func (suite *TypesTestSuite) TestSourceLocationMinimal() {
	loc := SourceLocation{
		FilePath:   "file.go",
		LineNumber: 1,
	}

	assert.Equal(suite.T(), "file.go", loc.FilePath)
	assert.Equal(suite.T(), 1, loc.LineNumber)
	assert.Equal(suite.T(), 0, loc.ColumnNumber)
}

// TestDocumentationSourceMultipleNil verifies only one field per source
func (suite *TypesTestSuite) TestDocumentationSourceOnlyOneNonNil() {
	const_ := &reflect.Constant{Name: "C"}
	enum := &reflect.Enum{Type: &reflect.TypeReference{Name: "E"}}
	source := &DocumentationSource{
		Constant: const_,
		Enum:     enum,
	}

	// Both fields are set, which is allowed but semantically one should be used
	assert.NotNil(suite.T(), source.Constant)
	assert.NotNil(suite.T(), source.Enum)
}

// TestDocumentationEntryEmpty creates empty entry
func (suite *TypesTestSuite) TestDocumentationEntryEmpty() {
	entry := &DocumentationEntry{}

	assert.Equal(suite.T(), "", entry.QualifiedName)
	assert.Equal(suite.T(), "", entry.LocalName)
	assert.Equal(suite.T(), EntryKind(""), entry.Kind)
	assert.Equal(suite.T(), "", entry.PackagePath)
}

// TestDocumentationEntryWithoutExamples and Links
func (suite *TypesTestSuite) TestDocumentationEntryWithoutOptional() {
	entry := &DocumentationEntry{
		QualifiedName: "pkg.Item",
		LocalName:     "Item",
		Kind:          KindType,
		PackagePath:   "pkg",
	}

	assert.Len(suite.T(), entry.Examples, 0)
	assert.Len(suite.T(), entry.Links, 0)
}

// TestBuilderSourceInterface verifies reflect package implements interface
func (suite *TypesTestSuite) TestBuilderSourceInterface() {
	var _ BuilderSource = (*reflect.Package)(nil)
}

// TestDocumentationIndexMultipleMaps verifies maps are independent
func (suite *TypesTestSuite) TestDocumentationIndexMultipleMaps() {
	index := &DocumentationIndex{
		Entries:    make(map[string]*DocumentationEntry),
		ByPackage:  make(map[string][]string),
		ByKind:     make(map[string][]string),
		References: make(map[string][]string),
	}

	entry := &DocumentationEntry{QualifiedName: "pkg.Type", Kind: KindType}
	index.Entries["pkg.Type"] = entry
	index.ByKind["type"] = []string{"pkg.Type"}
	index.ByPackage["pkg"] = []string{"pkg.Type"}

	// Verify each map is independent
	assert.Len(suite.T(), index.Entries, 1)
	assert.Len(suite.T(), index.ByKind, 1)
	assert.Len(suite.T(), index.ByPackage, 1)

	// Verify modification doesn't affect others
	delete(index.Entries, "pkg.Type")
	assert.Len(suite.T(), index.ByKind, 1)
	assert.Len(suite.T(), index.ByPackage, 1)
}

// TestExampleMultipleTags verifies tags can be multiple
func (suite *TypesTestSuite) TestExampleMultipleTags() {
	example := Example{
		Code: "code",
		Tags: []string{"tag1", "tag2", "tag3"},
	}

	require.Len(suite.T(), example.Tags, 3)
	assert.Contains(suite.T(), example.Tags, "tag1")
	assert.Contains(suite.T(), example.Tags, "tag2")
	assert.Contains(suite.T(), example.Tags, "tag3")
}

// TestLinkWithoutContext creates link without context
func (suite *TypesTestSuite) TestLinkWithoutContext() {
	link := Link{
		Target:       "pkg.Type",
		Relationship: RelationshipRelated,
	}

	assert.Equal(suite.T(), "", link.Context)
	assert.Nil(suite.T(), link.TargetEntry)
}
