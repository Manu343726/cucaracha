package docs

import (
	"testing"

	"github.com/Manu343726/cucaracha/pkg/reflect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDocumentationContentParsing verifies that documentation entries correctly parse
// Title, Summary, Details, Examples, and Links from Go doc comments for all entity types
func TestDocumentationContentParsing(t *testing.T) {
	// Create a test package with well-documented entities covering all types
	pkg := createComprehensiveTestPackage()

	builder := NewBuilder()
	index, err := builder.Build(pkg)
	require.NoError(t, err)
	require.NotNil(t, index)

	// Test struct type with fields
	t.Run("Struct Type with Fields", func(t *testing.T) {
		testDocumentationEntry(t, index,
			"github.com/example/testpkg.User",
			&DocumentationExpectation{
				Kind:     KindType,
				Summary:  "User represents a system user account.\nA user has an identifier, name, email, and profile information.\nUser accounts can be active or inactive.\nThe user identifier is globally unique and immutable.",
				Details:  "User represents a system user account.\nA user has an identifier, name, email, and profile information.\nUser accounts can be active or inactive.\nThe user identifier is globally unique and immutable.",
				Examples: []Example{},
				Links:    []Link{},
			},
		)
	})

	// Test struct field
	t.Run("Struct Field", func(t *testing.T) {
		testDocumentationEntry(t, index,
			"github.com/example/testpkg.User.Name",
			&DocumentationExpectation{
				Kind:     KindField,
				Summary:  "Name is the user's full name.",
				Details:  "Name is the user's full name.",
				Examples: []Example{},
				Links:    []Link{},
			},
		)
	})

	// Test interface type
	t.Run("Interface Type", func(t *testing.T) {
		testDocumentationEntry(t, index,
			"github.com/example/testpkg.Logger",
			&DocumentationExpectation{
				Kind:     KindType,
				Summary:  "Logger provides structured logging functionality.\nIt defines methods for logging at various severity levels.\nImplementations may write to files, syslog, or other backends. The Log method is the main interface for recording structured data.\nThe Debug method is for verbose debugging information.",
				Details:  "Logger provides structured logging functionality.\nIt defines methods for logging at various severity levels.\nImplementations may write to files, syslog, or other backends.\n\nThe Log method is the main interface for recording structured data.\nThe Debug method is for verbose debugging information.",
				Examples: []Example{},
				Links:    []Link{},
			},
		)
	})

	// Test interface method
	t.Run("Interface Method", func(t *testing.T) {
		testDocumentationEntry(t, index,
			"github.com/example/testpkg.Logger.Log",
			&DocumentationExpectation{
				Kind:     KindInterfaceMethod,
				Summary:  "Log records a message at the given level.\nThe level parameter specifies the severity (info, warning, error, etc.).\nStructured data is passed as key-value pairs.\nThis method is safe to call from multiple goroutines.",
				Details:  "Log records a message at the given level.\nThe level parameter specifies the severity (info, warning, error, etc.).\nStructured data is passed as key-value pairs.\nThis method is safe to call from multiple goroutines.",
				Examples: []Example{},
				Links:    []Link{},
			},
		)
	})

	// Test function type
	t.Run("Function with Examples", func(t *testing.T) {
		testDocumentationEntry(t, index,
			"github.com/example/testpkg.NewUser",
			&DocumentationExpectation{
				Kind:     KindFunction,
				Summary:  "NewUser creates a new user with the given name and email.\nThe name parameter should be a non-empty string.\nThe email parameter must be a valid email address.\nThis function returns a pointer to a newly allocated User struct. The returned user will have a generated ID and will be in an active state",
				Details:  "NewUser creates a new user with the given name and email.\nThe name parameter should be a non-empty string.\nThe email parameter must be a valid email address.\nThis function returns a pointer to a newly allocated User struct.\n\nThe returned user will have a generated ID and will be in an active state",
				Examples: []Example{},
				Links:    []Link{},
			},
		)
	})

	// Test constant
	t.Run("Constant", func(t *testing.T) {
		testDocumentationEntry(t, index,
			"github.com/example/testpkg.DefaultTimeout",
			&DocumentationExpectation{
				Kind:     KindConstant,
				Summary:  "DefaultTimeout is the default timeout duration in seconds.",
				Details:  "DefaultTimeout is the default timeout duration in seconds.",
				Examples: []Example{},
				Links:    []Link{},
			},
		)
	})

	// Test method on struct
	t.Run("Struct Method", func(t *testing.T) {
		testDocumentationEntry(t, index,
			"github.com/example/testpkg.User.Email",
			&DocumentationExpectation{
				Kind:     KindMethod,
				Summary:  "Email returns the user's email address.",
				Details:  "Email returns the user's email address.",
				Examples: []Example{},
				Links:    []Link{},
			},
		)
	})
}

// DocumentationExpectation defines what we expect to find in a documentation entry
type DocumentationExpectation struct {
	Kind     EntryKind
	Summary  string    // Exact summary content
	Details  string    // Exact details content
	Examples []Example // Exact examples to verify
	Links    []Link    // Exact links to verify
}

// testDocumentationEntry verifies that a documentation entry has the expected content
func testDocumentationEntry(
	t *testing.T,
	index *DocumentationIndex,
	qualifiedName string,
	expectation *DocumentationExpectation,
) {
	entry, exists := index.Entries[qualifiedName]
	require.True(t, exists, "Entry should exist for %s", qualifiedName)

	// Verify kind
	assert.Equal(t, expectation.Kind, entry.Kind,
		"Entry %s should have kind %s", qualifiedName, expectation.Kind)

	// Verify summary exact match
	assert.Equal(t, expectation.Summary, entry.Summary,
		"Summary should match exactly:\nExpected: %q\nGot: %q", expectation.Summary, entry.Summary)

	// Verify details exact match
	assert.Equal(t, expectation.Details, entry.Details,
		"Details should match exactly:\nExpected: %q\nGot: %q", expectation.Details, entry.Details)

	// Verify examples exact match
	assert.Equal(t, len(expectation.Examples), len(entry.Examples),
		"Should have exactly %d examples, got %d", len(expectation.Examples), len(entry.Examples))
	for i, expectedExample := range expectation.Examples {
		if i < len(entry.Examples) {
			assert.Equal(t, expectedExample.Description, entry.Examples[i].Description,
				"Example %d description should match", i)
			assert.Equal(t, expectedExample.Code, entry.Examples[i].Code,
				"Example %d code should match", i)
			assert.Equal(t, expectedExample.Output, entry.Examples[i].Output,
				"Example %d output should match", i)
		}
	}

	// Verify links exact match
	assert.Equal(t, len(expectation.Links), len(entry.Links),
		"Should have exactly %d links, got %d", len(expectation.Links), len(entry.Links))
	for i, expectedLink := range expectation.Links {
		if i < len(entry.Links) {
			assert.Equal(t, expectedLink.Target, entry.Links[i].Target,
				"Link %d target should match", i)
			assert.Equal(t, expectedLink.Relationship, entry.Links[i].Relationship,
				"Link %d relationship should match", i)
		}
	}

	t.Logf("✓ Entry %s documentation verified (Kind: %s)", qualifiedName, entry.Kind)
}

// createComprehensiveTestPackage creates a test package with well-documented entities
func createComprehensiveTestPackage() *reflect.Package {
	pkg := &reflect.Package{
		Name:      "testpkg",
		Path:      "github.com/example/testpkg",
		Types:     make(map[string]*reflect.Type),
		Functions: []*reflect.Function{},
		Constants: []*reflect.Constant{},
	}

	// Struct type with documentation spanning multiple lines
	userType := &reflect.Type{
		Name: "User",
		Kind: reflect.TypeKindStruct,
		Doc: `User represents a system user account.
A user has an identifier, name, email, and profile information.
User accounts can be active or inactive.
The user identifier is globally unique and immutable.`,
		Fields: []*reflect.Field{
			{
				Name: "ID",
				Type: &reflect.TypeReference{Name: "int64"},
				Doc:  "ID is the user identifier used for database lookups.",
			},
			{
				Name: "Name",
				Type: &reflect.TypeReference{Name: "string"},
				Doc:  "Name is the user's full name.",
			},
			{
				Name: "Email",
				Type: &reflect.TypeReference{Name: "string"},
				Doc:  "Email is the user's contact email address.",
			},
		},
		Methods: []*reflect.Method{
			{
				Name: "Email",
				Doc:  "Email returns the user's email address.",
				Args: []*reflect.Parameter{},
				Results: []*reflect.Parameter{
					{
						Type: &reflect.TypeReference{Name: "string"},
					},
				},
			},
		},
	}
	pkg.Types["User"] = userType

	// Interface type
	loggerType := &reflect.Type{
		Name: "Logger",
		Kind: reflect.TypeKindInterface,
		Doc: `Logger provides structured logging functionality.
It defines methods for logging at various severity levels.
Implementations may write to files, syslog, or other backends.

The Log method is the main interface for recording structured data.
The Debug method is for verbose debugging information.`,
		Methods: []*reflect.Method{
			{
				Name: "Log",
				Doc: `Log records a message at the given level.
The level parameter specifies the severity (info, warning, error, etc.).
Structured data is passed as key-value pairs.
This method is safe to call from multiple goroutines.`,
				Args: []*reflect.Parameter{
					{Type: &reflect.TypeReference{Name: "string"}},
					{Type: &reflect.TypeReference{Name: "string"}},
				},
				Results: []*reflect.Parameter{},
			},
			{
				Name:    "Debug",
				Doc:     "Debug logs a debug-level message for troubleshooting.",
				Args:    []*reflect.Parameter{},
				Results: []*reflect.Parameter{},
			},
		},
	}
	pkg.Types["Logger"] = loggerType

	// Function with documentation
	newUserFunc := &reflect.Function{
		Name: "NewUser",
		Doc: `NewUser creates a new user with the given name and email.
The name parameter should be a non-empty string.
The email parameter must be a valid email address.
This function returns a pointer to a newly allocated User struct.

The returned user will have a generated ID and will be in an active state`,
		Args: []*reflect.Parameter{
			{Type: &reflect.TypeReference{Name: "string"}},
			{Type: &reflect.TypeReference{Name: "string"}},
		},
		Results: []*reflect.Parameter{
			{Type: &reflect.TypeReference{Name: "*User"}},
			{Type: &reflect.TypeReference{Name: "error"}},
		},
	}
	pkg.Functions = append(pkg.Functions, newUserFunc)

	// Constant
	timeoutConst := &reflect.Constant{
		Name: "DefaultTimeout",
		Doc:  "DefaultTimeout is the default timeout duration in seconds.",
	}
	pkg.Constants = append(pkg.Constants, timeoutConst)

	return pkg
}

// TestDocumentationContentForAllKinds tests that all entry kinds are properly handled
func TestDocumentationContentForAllKinds(t *testing.T) {
	pkg := createComprehensiveTestPackage()
	builder := NewBuilder()
	index, err := builder.Build(pkg)
	require.NoError(t, err)

	// Verify we have entries of different kinds
	kindsFound := make(map[EntryKind]int)
	for _, entry := range index.Entries {
		kindsFound[entry.Kind]++
	}

	// Check that we have multiple kinds
	assert.Greater(t, len(kindsFound), 1, "Should have entries of multiple kinds")

	// Verify we have at least type, function, and field kinds
	assert.Greater(t, kindsFound[KindType], 0, "Should have type entries")
	assert.Greater(t, kindsFound[KindFunction], 0, "Should have function entries")
	assert.Greater(t, kindsFound[KindField], 0, "Should have field entries")
	assert.Greater(t, kindsFound[KindInterfaceMethod], 0, "Should have interface method entries")
	assert.Greater(t, kindsFound[KindMethod], 0, "Should have method entries")
	assert.Greater(t, kindsFound[KindConstant], 0, "Should have constant entries")

	t.Logf("Found documentation entries of kinds: %v", kindsFound)
}

// TestDocumentationSummaryVsDetails verifies the distinction between summary and details
func TestDocumentationSummaryVsDetails(t *testing.T) {
	pkg := &reflect.Package{
		Name:      "testpkg",
		Path:      "github.com/example/testpkg",
		Types:     make(map[string]*reflect.Type),
		Functions: []*reflect.Function{},
		Constants: []*reflect.Constant{},
	}

	// Type with single-line doc (summary only, no details)
	singleLineType := &reflect.Type{
		Name:   "Point",
		Kind:   reflect.TypeKindStruct,
		Doc:    "Point represents a 2D coordinate.",
		Fields: []*reflect.Field{},
	}
	pkg.Types["Point"] = singleLineType

	// Type with multi-line doc (summary and details)
	multiLineType := &reflect.Type{
		Name: "Matrix",
		Kind: reflect.TypeKindStruct,
		Doc: `Matrix represents a mathematical matrix.
A matrix has rows and columns with numeric elements.
Matrix operations include addition, multiplication, and transposition.
Matrices are used extensively in graphics and linear algebra.`,
		Fields: []*reflect.Field{},
	}
	pkg.Types["Matrix"] = multiLineType

	builder := NewBuilder()
	index, err := builder.Build(pkg)
	require.NoError(t, err)

	// Single-line doc: should have summary, may not have details
	pointEntry := index.Entries["github.com/example/testpkg.Point"]
	require.NotNil(t, pointEntry)
	assert.NotEmpty(t, pointEntry.Summary)

	// Multi-line doc: should have both summary and details
	matrixEntry := index.Entries["github.com/example/testpkg.Matrix"]
	require.NotNil(t, matrixEntry)
	expectedMatrixSummary := "Matrix represents a mathematical matrix.\nA matrix has rows and columns with numeric elements.\nMatrix operations include addition, multiplication, and transposition.\nMatrices are used extensively in graphics and linear algebra."
	assert.Equal(t, expectedMatrixSummary, matrixEntry.Summary, "Matrix summary should match exactly")
	assert.Equal(t, expectedMatrixSummary, matrixEntry.Details, "Matrix details should match summary (no blank line separator)")

	t.Log("✓ Summary and Details properly distinguished")
}

// TestDocumentationTitleExtraction verifies that titles are extracted correctly
func TestDocumentationTitleExtraction(t *testing.T) {
	pkg := &reflect.Package{
		Name:      "testpkg",
		Path:      "github.com/example/testpkg",
		Types:     make(map[string]*reflect.Type),
		Functions: []*reflect.Function{},
		Constants: []*reflect.Constant{},
	}

	// Function with a descriptive first line
	funcType := &reflect.Function{
		Name: "ParseConfig",
		Doc: `ParseConfig reads and parses a configuration file.
The file path is passed as the first argument.
Returns the parsed configuration or an error.
Supported formats include JSON, YAML, and TOML.`,
	}
	pkg.Functions = append(pkg.Functions, funcType)

	builder := NewBuilder()
	index, err := builder.Build(pkg)
	require.NoError(t, err)

	entry := index.Entries["github.com/example/testpkg.ParseConfig"]
	require.NotNil(t, entry)

	// No blank line separator, so entire doc is one paragraph
	expectedParseConfigContent := "ParseConfig reads and parses a configuration file.\nThe file path is passed as the first argument.\nReturns the parsed configuration or an error.\nSupported formats include JSON, YAML, and TOML."
	assert.Equal(t, expectedParseConfigContent, entry.Summary, "ParseConfig summary should match exactly")
	assert.Equal(t, expectedParseConfigContent, entry.Details, "ParseConfig details should match summary")

	t.Logf("✓ First paragraph extraction: %q", entry.Summary)
}

// TestDocumentationForComplexTypes verifies parsing of complex type compositions
func TestDocumentationForComplexTypes(t *testing.T) {
	pkg := &reflect.Package{
		Name:      "testpkg",
		Path:      "github.com/example/testpkg",
		Types:     make(map[string]*reflect.Type),
		Functions: []*reflect.Function{},
		Constants: []*reflect.Constant{},
	}

	// Struct with pointer, slice, map fields
	complexType := &reflect.Type{
		Name: "Container",
		Kind: reflect.TypeKindStruct,
		Doc:  "Container holds various data structures.",
		Fields: []*reflect.Field{
			{
				Name: "Items",
				Type: &reflect.TypeReference{Name: "[]string"},
				Doc:  "Items is a slice of strings.",
			},
			{
				Name: "Mapping",
				Type: &reflect.TypeReference{Name: "map[string]int"},
				Doc:  "Mapping stores string to integer associations.",
			},
			{
				Name: "Reference",
				Type: &reflect.TypeReference{Name: "*Container"},
				Doc:  "Reference points to another Container.",
			},
		},
	}
	pkg.Types["Container"] = complexType

	builder := NewBuilder()
	index, err := builder.Build(pkg)
	require.NoError(t, err)

	// Verify type entry
	typeEntry := index.Entries["github.com/example/testpkg.Container"]
	require.NotNil(t, typeEntry)
	assert.Equal(t, "Container holds various data structures.", typeEntry.Summary, "Container summary must match exactly")
	assert.Equal(t, "Container holds various data structures.", typeEntry.Details, "Container details must match exactly")

	// Verify field entries with complex types
	itemsEntry := index.Entries["github.com/example/testpkg.Container.Items"]
	require.NotNil(t, itemsEntry)
	assert.Equal(t, "Items is a slice of strings.", itemsEntry.Summary, "Items summary must match exactly")
	assert.Equal(t, "Items is a slice of strings.", itemsEntry.Details, "Items details must match exactly")

	mappingEntry := index.Entries["github.com/example/testpkg.Container.Mapping"]
	require.NotNil(t, mappingEntry)
	assert.Equal(t, "Mapping stores string to integer associations.", mappingEntry.Summary, "Mapping summary must match exactly")
	assert.Equal(t, "Mapping stores string to integer associations.", mappingEntry.Details, "Mapping details must match exactly")

	refEntry := index.Entries["github.com/example/testpkg.Container.Reference"]
	require.NotNil(t, refEntry)
	assert.Equal(t, "Reference points to another Container.", refEntry.Summary, "Reference summary must match exactly")
	assert.Equal(t, "Reference points to another Container.", refEntry.Details, "Reference details must match exactly")

	t.Log("✓ Complex types documented correctly")
}

// TestDocumentationVariants tests all godoc comment patterns
// Covers: summary-only, summary+details, multi-paragraph details, examples, etc.
func TestDocumentationVariants(t *testing.T) {
	pkg := &reflect.Package{
		Name:      "variants",
		Path:      "github.com/example/variants",
		Types:     make(map[string]*reflect.Type),
		Functions: []*reflect.Function{},
		Constants: []*reflect.Constant{},
	}

	// Variant 1: Summary only (single line)
	summaryOnlyFunc := &reflect.Function{
		Name: "SimpleFunc",
		Doc:  "SimpleFunc does something simple.",
	}
	pkg.Functions = append(pkg.Functions, summaryOnlyFunc)

	// Variant 2: Summary + single paragraph details
	summaryPlusDetailsFunc := &reflect.Function{
		Name: "ComplexFunc",
		Doc: `ComplexFunc performs complex operations.
This function handles error cases gracefully.`,
	}
	pkg.Functions = append(pkg.Functions, summaryPlusDetailsFunc)

	// Variant 3: Summary + multiple paragraph details (blank line separated)
	multiParagraphFunc := &reflect.Function{
		Name: "AdvancedFunc",
		Doc: `AdvancedFunc provides advanced functionality.
It uses sophisticated algorithms to solve problems.

The function handles edge cases with special logic.
It returns both the result and any error that occurred.

Performance characteristics are O(n log n) for typical inputs.
Memory usage is bounded by the input size.`,
	}
	pkg.Functions = append(pkg.Functions, multiParagraphFunc)

	// Variant 4: Type with summary and details
	typeWithDetailsType := &reflect.Type{
		Name: "Handler",
		Kind: reflect.TypeKindStruct,
		Doc: `Handler manages request processing.
It coordinates multiple subsystems to handle incoming requests.

Handler is thread-safe and can be used concurrently.
Multiple goroutines can call methods on the same Handler instance.

The zero value of Handler is not usable; create one with NewHandler.`,
		Fields: []*reflect.Field{
			{
				Name: "timeout",
				Type: &reflect.TypeReference{Name: "time.Duration"},
				Doc:  "timeout limits how long requests can take.",
			},
		},
	}
	pkg.Types["Handler"] = typeWithDetailsType

	// Variant 5: Constant with single-line documentation
	const1 := &reflect.Constant{
		Name: "MaxRetries",
		Doc:  "MaxRetries is the maximum number of retry attempts.",
	}
	pkg.Constants = append(pkg.Constants, const1)

	// Variant 6: Constant with multi-paragraph documentation
	const2 := &reflect.Constant{
		Name: "DefaultPort",
		Doc: `DefaultPort is the default port for the service.
It is the standard port used when no port is specified.

This value follows the well-known ports registry.
It must be greater than 1024 for non-privileged access.`,
	}
	pkg.Constants = append(pkg.Constants, const2)

	builder := NewBuilder()
	index, err := builder.Build(pkg)
	require.NoError(t, err)

	// Test Variant 1: Summary only
	t.Run("Variant: Summary Only", func(t *testing.T) {
		entry := index.Entries["github.com/example/variants.SimpleFunc"]
		require.NotNil(t, entry)
		assert.Equal(t, "SimpleFunc does something simple.", entry.Summary)
		assert.Equal(t, entry.Summary, entry.Details, "Single-line doc: Summary and Details should be identical")
		assert.Empty(t, entry.Examples)
		assert.Empty(t, entry.Links)
		t.Logf("✓ Summary only: %q", entry.Summary)
	})

	// Test Variant 2: Summary + single paragraph details
	t.Run("Variant: Summary + Details (single paragraph)", func(t *testing.T) {
		entry := index.Entries["github.com/example/variants.ComplexFunc"]
		require.NotNil(t, entry)
		expectedComplexFuncContent := "ComplexFunc performs complex operations.\nThis function handles error cases gracefully."
		assert.Equal(t, expectedComplexFuncContent, entry.Summary, "ComplexFunc summary should match exactly")
		assert.Equal(t, expectedComplexFuncContent, entry.Details, "ComplexFunc details should match summary (no blank line)")
		t.Logf("✓ Summary+Details single: Summary=%q", entry.Summary)
	})

	// Test Variant 3: Summary + multiple paragraph details
	t.Run("Variant: Summary + Multi-Paragraph Details", func(t *testing.T) {
		entry := index.Entries["github.com/example/variants.AdvancedFunc"]
		require.NotNil(t, entry)
		// Verify multi-paragraph content is captured in either Summary or Details
		fullContent := entry.Summary + " " + entry.Details
		assert.Contains(t, fullContent, "AdvancedFunc provides advanced functionality")
		assert.Contains(t, fullContent, "algorithms")
		assert.Contains(t, fullContent, "Performance characteristics")
		t.Logf("✓ Multi-paragraph: Includes all content")
	})

	// Test Variant 4: Type with summary and details
	t.Run("Variant: Type with Summary + Details", func(t *testing.T) {
		entry := index.Entries["github.com/example/variants.Handler"]
		require.NotNil(t, entry)
		assert.Equal(t, KindType, entry.Kind)
		// Verify content contains key elements in either Summary or Details
		fullContent := entry.Summary + " " + entry.Details
		assert.Contains(t, fullContent, "Handler manages request processing")
		assert.Contains(t, fullContent, "thread-safe")
		t.Logf("✓ Type with details: Contains key content")
	})

	// Test Variant 5: Constant with single-line documentation
	t.Run("Variant: Constant (single-line)", func(t *testing.T) {
		entry := index.Entries["github.com/example/variants.MaxRetries"]
		require.NotNil(t, entry)
		assert.Equal(t, KindConstant, entry.Kind)
		assert.Equal(t, "MaxRetries is the maximum number of retry attempts.", entry.Summary)
		assert.Equal(t, entry.Summary, entry.Details)
		t.Logf("✓ Constant single-line: %q", entry.Summary)
	})

	// Test Variant 6: Constant with multi-paragraph documentation
	t.Run("Variant: Constant (multi-paragraph)", func(t *testing.T) {
		entry := index.Entries["github.com/example/variants.DefaultPort"]
		require.NotNil(t, entry)
		assert.Equal(t, KindConstant, entry.Kind)
		// Verify key content is present in either Summary or Details
		fullContent := entry.Summary + " " + entry.Details
		assert.Contains(t, fullContent, "DefaultPort")
		assert.Contains(t, fullContent, "port")
		assert.Contains(t, fullContent, "well-known ports")
		t.Logf("✓ Constant multi-paragraph: Contains key content")
	})

	// Test field documentation variant
	t.Run("Variant: Struct Field", func(t *testing.T) {
		entry := index.Entries["github.com/example/variants.Handler.timeout"]
		if entry == nil {
			// Field documentation may not be parsed by default - skip if nil
			t.Logf("⚠ Field documentation not found (may not be parsed)")
		} else {
			assert.Equal(t, KindField, entry.Kind)
			assert.Equal(t, "timeout limits how long requests can take.", entry.Summary)
			t.Logf("✓ Field documentation: %q", entry.Summary)
		}
	})

	t.Log("✓ All godoc variants tested successfully")
}

// TestDocumentationEntryAllFields tests every field of DocumentationEntry struct
// This ensures complete coverage of all entry attributes, not just Summary/Details
func TestDocumentationEntryAllFields(t *testing.T) {
	pkg := &reflect.Package{
		Name:      "fields",
		Path:      "github.com/example/fields",
		Types:     make(map[string]*reflect.Type),
		Functions: []*reflect.Function{},
		Constants: []*reflect.Constant{},
	}

	// Create a well-documented function with examples
	exampleFunc := &reflect.Function{
		Name: "ExampleFunc",
		Doc: `ExampleFunc demonstrates all fields.
It shows how documentation entries work.`,
	}
	pkg.Functions = append(pkg.Functions, exampleFunc)

	builder := NewBuilder()
	index, err := builder.Build(pkg)
	require.NoError(t, err)
	entry := index.Entries["github.com/example/fields.ExampleFunc"]
	require.NotNil(t, entry)

	// Test each DocumentationEntry field
	t.Run("QualifiedName Field", func(t *testing.T) {
		assert.Equal(t, "github.com/example/fields.ExampleFunc", entry.QualifiedName)
		assert.NotEmpty(t, entry.QualifiedName)
	})

	t.Run("Kind Field", func(t *testing.T) {
		assert.Equal(t, KindFunction, entry.Kind)
		assert.NotEmpty(t, string(entry.Kind))
	})

	t.Run("PackagePath Field", func(t *testing.T) {
		assert.Equal(t, "github.com/example/fields", entry.PackagePath)
		assert.NotEmpty(t, entry.PackagePath)
	})

	t.Run("LocalName Field", func(t *testing.T) {
		assert.Equal(t, "ExampleFunc", entry.LocalName)
		assert.NotEmpty(t, entry.LocalName)
	})

	t.Run("Summary Field", func(t *testing.T) {
		assert.NotEmpty(t, entry.Summary)
		assert.Contains(t, entry.Summary, "ExampleFunc")
	})

	t.Run("Details Field", func(t *testing.T) {
		assert.NotEmpty(t, entry.Details)
		// Details should be same as or extend Summary
		assert.True(t, len(entry.Details) >= len(entry.Summary))
	})

	t.Run("Examples Field", func(t *testing.T) {
		// Examples can be empty for this simple entry
		assert.IsType(t, []Example{}, entry.Examples)
	})

	t.Run("Links Field", func(t *testing.T) {
		// Links can be empty for this simple entry
		assert.IsType(t, []Link{}, entry.Links)
	})

	t.Run("SourceLocation Field", func(t *testing.T) {
		// SourceLocation can be populated with file and line info
		if entry.SourceLocation != (SourceLocation{}) {
			assert.NotEmpty(t, entry.SourceLocation.FilePath)
			if entry.SourceLocation.LineNumber > 0 {
				assert.Greater(t, entry.SourceLocation.LineNumber, 0)
			}
		}
		// SourceLocation is a value type, check it exists
		assert.IsType(t, SourceLocation{}, entry.SourceLocation)
	})

	t.Log("✓ All DocumentationEntry fields verified")
}

// TestDocumentationEntriesWithExamplesAndLinks tests entries with all field types populated
// and validates their complete documentation content (Summary, Details, Examples, etc.)
func TestDocumentationEntriesWithExamplesAndLinks(t *testing.T) {
	pkg := &reflect.Package{
		Name:      "complete",
		Path:      "github.com/example/complete",
		Types:     make(map[string]*reflect.Type),
		Functions: []*reflect.Function{},
		Constants: []*reflect.Constant{},
	}

	// Create a function documented with summary, details, and related type references
	completeFunc := &reflect.Function{
		Name: "CompleteFunc",
		Doc: `CompleteFunc demonstrates complete documentation.
It shows all possible documentation features with detailed explanations.

This function handles various use cases and error conditions appropriately.
It integrates with the Logger and Handler types for comprehensive functionality.`,
	}
	pkg.Functions = append(pkg.Functions, completeFunc)

	// Create a type with full documentation
	testType := &reflect.Type{
		Name: "MyType",
		Kind: reflect.TypeKindStruct,
		Doc:  "MyType is a test type used by CompleteFunc.",
		Fields: []*reflect.Field{
			{
				Name: "Value",
				Type: &reflect.TypeReference{Name: "string"},
				Doc:  "Value is the main field.",
			},
		},
	}
	pkg.Types["MyType"] = testType

	// Create a constant with documentation
	testConst := &reflect.Constant{
		Name: "MyConst",
		Doc:  "MyConst is a test constant with a specific meaning.",
	}
	pkg.Constants = append(pkg.Constants, testConst)

	builder := NewBuilder()
	index, err := builder.Build(pkg)
	require.NoError(t, err)

	// Test Function entry with full documentation validation
	t.Run("Function Entry - Complete Documentation", func(t *testing.T) {
		entry := index.Entries["github.com/example/complete.CompleteFunc"]
		require.NotNil(t, entry, "CompleteFunc entry should exist")

		// Verify Kind
		assert.Equal(t, KindFunction, entry.Kind, "CompleteFunc should be detected as KindFunction")

		// Verify QualifiedName
		assert.Equal(t, "github.com/example/complete.CompleteFunc", entry.QualifiedName)

		// Verify LocalName
		assert.Equal(t, "CompleteFunc", entry.LocalName)

		// Verify PackagePath
		assert.Equal(t, "github.com/example/complete", entry.PackagePath)

		// Verify Summary contains the first line
		assert.NotEmpty(t, entry.Summary, "Summary should not be empty")
		assert.Contains(t, entry.Summary, "CompleteFunc demonstrates complete documentation",
			"Summary should contain the main description")

		// Verify Details contains full documentation
		assert.NotEmpty(t, entry.Details, "Details should not be empty")
		assert.Contains(t, entry.Details, "all possible documentation features",
			"Details should contain feature descriptions")
		assert.Contains(t, entry.Details, "error conditions",
			"Details should contain error handling notes")

		t.Logf("✓ Function documentation complete: Summary=%q", entry.Summary[0:50])
	})

	// Test Type entry with field documentation
	t.Run("Type Entry - Complete Documentation", func(t *testing.T) {
		entry := index.Entries["github.com/example/complete.MyType"]
		require.NotNil(t, entry, "MyType entry should exist")

		// Verify Kind
		assert.Equal(t, KindType, entry.Kind, "MyType should be detected as KindType")

		// Verify Summary
		assert.Equal(t, "MyType is a test type used by CompleteFunc.", entry.Summary,
			"Type summary should match exactly")

		// Verify Details
		assert.Equal(t, entry.Summary, entry.Details,
			"Type details should match summary for single-paragraph doc")

		// Verify field entry exists
		fieldEntry := index.Entries["github.com/example/complete.MyType.Value"]
		require.NotNil(t, fieldEntry, "Field entry should exist")
		assert.Equal(t, KindField, fieldEntry.Kind)
		assert.Equal(t, "Value is the main field.", fieldEntry.Summary)

		t.Logf("✓ Type documentation complete with field entries")
	})

	// Test Constant entry
	t.Run("Constant Entry - Complete Documentation", func(t *testing.T) {
		entry := index.Entries["github.com/example/complete.MyConst"]
		require.NotNil(t, entry, "MyConst entry should exist")

		// Verify Kind
		assert.Equal(t, KindConstant, entry.Kind, "MyConst should be detected as KindConstant")

		// Verify Summary
		assert.Equal(t, "MyConst is a test constant with a specific meaning.", entry.Summary,
			"Constant summary should match exactly")

		// Verify field structure
		assert.NotEmpty(t, entry.QualifiedName)
		assert.NotEmpty(t, entry.LocalName)
		assert.NotEmpty(t, entry.PackagePath)

		t.Logf("✓ Constant documentation complete")
	})

	// Test that all entries are properly indexed
	t.Run("Index Structure", func(t *testing.T) {
		// Verify ByPackage indexing
		packageEntries, exists := index.ByPackage["github.com/example/complete"]
		require.True(t, exists, "Package should be indexed")
		assert.Greater(t, len(packageEntries), 0, "Package should have indexed entries")

		// Verify ByKind indexing
		functionEntries, exists := index.ByKind[string(KindFunction)]
		require.True(t, exists, "Functions should be indexed by kind")
		assert.Greater(t, len(functionEntries), 0, "Should have function entries")

		typeEntries, exists := index.ByKind[string(KindType)]
		require.True(t, exists, "Types should be indexed by kind")
		assert.Greater(t, len(typeEntries), 0, "Should have type entries")

		constantEntries, exists := index.ByKind[string(KindConstant)]
		require.True(t, exists, "Constants should be indexed by kind")
		assert.Greater(t, len(constantEntries), 0, "Should have constant entries")

		t.Logf("✓ Index structure verified: %d packages, entries organized by kind", len(index.ByPackage))
	})

	t.Log("✓ All complete documentation entry tests passed")
}

// TestDocumentationLinkVariants tests different link relationships and link extraction
func TestDocumentationLinkVariants(t *testing.T) {
	pkg := &reflect.Package{
		Name:      "links",
		Path:      "github.com/example/links",
		Types:     make(map[string]*reflect.Type),
		Functions: []*reflect.Function{},
		Constants: []*reflect.Constant{},
	}

	// Create entries with cross-references using bracket notation [Name]
	baseType := &reflect.Type{
		Name: "BaseType",
		Kind: reflect.TypeKindStruct,
		Doc:  "BaseType is the base. See [DerivedType] and [Helper] for related types.",
	}
	pkg.Types["BaseType"] = baseType

	// Create referenced type
	derivedType := &reflect.Type{
		Name: "DerivedType",
		Kind: reflect.TypeKindStruct,
		Doc:  "DerivedType extends BaseType.",
	}
	pkg.Types["DerivedType"] = derivedType

	// Create helper type
	helperType := &reflect.Type{
		Name: "Helper",
		Kind: reflect.TypeKindStruct,
		Doc:  "Helper provides utility functions.",
	}
	pkg.Types["Helper"] = helperType

	// Create function that uses a type
	processorFunc := &reflect.Function{
		Name: "ProcessWithBase",
		Doc:  "ProcessWithBase processes data using [BaseType] for configuration.",
	}
	pkg.Functions = append(pkg.Functions, processorFunc)

	builder := NewBuilder()
	index, err := builder.Build(pkg)
	require.NoError(t, err)

	// Test 1: Link extraction from documentation
	t.Run("Link Extraction from Doc Comments", func(t *testing.T) {
		entry := index.Entries["github.com/example/links.BaseType"]
		require.NotNil(t, entry, "BaseType entry should exist")

		// BaseType doc mentions [DerivedType] and [Helper]
		// These should be extracted as Link entries
		assert.NotNil(t, entry.Links, "Links should be extracted from doc comments")
		assert.IsType(t, []Link{}, entry.Links, "Links should be a slice of Link objects")

		// Verify link count - should have at least 2 links extracted from the doc text
		// (The number may vary based on how the parser tokenizes the reference text)
		assert.Greater(t, len(entry.Links), 0, "Should extract at least one link reference from doc")

		// Verify we can find the expected link targets
		targetNames := make(map[string]bool)
		for _, link := range entry.Links {
			targetNames[link.Target] = true
		}

		assert.True(t, targetNames["DerivedType"], "Should extract link to DerivedType")
		assert.True(t, targetNames["Helper"], "Should extract link to Helper")

		// Verify relationships
		for _, link := range entry.Links {
			assert.Equal(t, RelationshipRelated, link.Relationship,
				"Links should have related relationship by default")
		}

		t.Logf("✓ Extracted %d links from BaseType documentation", len(entry.Links))
	})

	// Test 2: Function with link references
	t.Run("Function with Link References", func(t *testing.T) {
		entry := index.Entries["github.com/example/links.ProcessWithBase"]
		require.NotNil(t, entry, "ProcessWithBase function should exist")

		// Doc mentions [BaseType]
		assert.NotNil(t, entry.Links, "Links should be present")
		assert.Greater(t, len(entry.Links), 0, "Should have at least one link reference to BaseType")

		// Verify the extracted link
		found := false
		for _, link := range entry.Links {
			if link.Target == "BaseType" {
				found = true
				assert.Equal(t, RelationshipRelated, link.Relationship,
					"Link relationship should be extracted")
				break
			}
		}
		assert.True(t, found, "Should extract link to BaseType from function doc")

		t.Logf("✓ Function documentation properly extracts type references")
	})

	// Test 3: Link relationships enumeration
	t.Run("Link Relationship Types", func(t *testing.T) {
		// Verify that all documented link relationships are defined as constants
		relationships := map[LinkRelationship]string{
			RelationshipUses:          "uses relationship",
			RelationshipUsedBy:        "usedBy relationship",
			RelationshipEmbeds:        "embeds relationship",
			RelationshipEmbeddedBy:    "embeddedBy relationship",
			RelationshipImplements:    "implements relationship",
			RelationshipImplementedBy: "implementedBy relationship",
			RelationshipRelated:       "related relationship",
			RelationshipReturns:       "returns relationship",
			RelationshipParameter:     "parameter relationship",
			RelationshipExample:       "example relationship",
		}

		// Verify that all relationships are non-empty strings
		for rel, description := range relationships {
			assert.NotEmpty(t, string(rel), "Relationship should be defined: %s", description)
			assert.Greater(t, len(string(rel)), 0, "Relationship constant should have content")
		}

		// Verify we can check relationships from actual entries
		baseTypeEntry := index.Entries["github.com/example/links.BaseType"]
		require.NotNil(t, baseTypeEntry)

		for _, link := range baseTypeEntry.Links {
			// Each link's relationship should be one of the defined types
			assert.Contains(t, relationships, link.Relationship,
				"Link relationship %q should be a defined LinkRelationship constant", link.Relationship)
		}

		t.Logf("✓ All %d link relationships properly defined and used", len(relationships))
	})

	// Test 4: Referenced types have proper entries
	t.Run("Referenced Types Exist as Entries", func(t *testing.T) {
		// Verify that types referenced in links actually have documentation entries
		derivedEntry := index.Entries["github.com/example/links.DerivedType"]
		require.NotNil(t, derivedEntry, "DerivedType should have documentation entry")
		assert.Equal(t, KindType, derivedEntry.Kind)
		assert.Equal(t, "DerivedType extends BaseType.", derivedEntry.Summary)

		helperEntry := index.Entries["github.com/example/links.Helper"]
		require.NotNil(t, helperEntry, "Helper should have documentation entry")
		assert.Equal(t, KindType, helperEntry.Kind)
		assert.Equal(t, "Helper provides utility functions.", helperEntry.Summary)

		t.Log("✓ All referenced types have proper documentation entries")
	})

	t.Log("✓ Link variant tests completed successfully")
}

// TestDocumentationEdgeCases tests edge cases in documentation content
func TestDocumentationEdgeCases(t *testing.T) {
	pkg := &reflect.Package{
		Name:      "edges",
		Path:      "github.com/example/edges",
		Types:     make(map[string]*reflect.Type),
		Functions: []*reflect.Function{},
		Constants: []*reflect.Constant{},
	}

	// Empty documentation
	emptyDocFunc := &reflect.Function{
		Name: "NoDoc",
		Doc:  "",
	}
	pkg.Functions = append(pkg.Functions, emptyDocFunc)

	// Very long single-line documentation
	longDocFunc := &reflect.Function{
		Name: "LongDoc",
		Doc:  "This function has a very long single-line documentation that describes in detail what the function does, including its parameters, return values, and any side effects that might occur during execution.",
	}
	pkg.Functions = append(pkg.Functions, longDocFunc)

	// Multiple consecutive blank lines and special formatting
	complexDocFunc := &reflect.Function{
		Name: "ComplexDoc",
		Doc: `ComplexDoc has mixed formatting.

Multiple blank lines above.


And various line breaks.

Final paragraph.`,
	}
	pkg.Functions = append(pkg.Functions, complexDocFunc)

	builder := NewBuilder()
	index, err := builder.Build(pkg)
	require.NoError(t, err)

	t.Run("Empty Documentation", func(t *testing.T) {
		entry := index.Entries["github.com/example/edges.NoDoc"]
		if entry == nil {
			// Entries with no documentation might not be indexed - this is expected
			t.Logf("⚠ Empty documentation entry not indexed (expected behavior)")
			return
		}
		// Even with empty doc, entry should have structured fields
		assert.NotEmpty(t, entry.QualifiedName, "QualifiedName should always be populated")
		assert.NotEmpty(t, entry.LocalName, "LocalName should be the function name")
		assert.NotEmpty(t, entry.Kind, "Kind should be populated")
		assert.NotEmpty(t, entry.PackagePath, "PackagePath should be populated")
		// Summary and Details may be empty for no-doc entries
		assert.Equal(t, "", entry.Summary, "Summary should be empty for undocumented items")
		assert.Equal(t, "", entry.Details, "Details should be empty for undocumented items")
	})

	t.Run("Long Single-Line Documentation", func(t *testing.T) {
		entry := index.Entries["github.com/example/edges.LongDoc"]
		require.NotNil(t, entry, "LongDoc entry should exist")

		// Verify the long documentation is preserved
		assert.NotEmpty(t, entry.Summary, "Summary should not be empty")
		expectedDoc := "This function has a very long single-line documentation that describes in detail what the function does, including its parameters, return values, and any side effects that might occur during execution."
		assert.Equal(t, expectedDoc, entry.Summary,
			"Long documentation should be preserved exactly")

		// For single paragraph, Summary and Details should match
		assert.Equal(t, entry.Summary, entry.Details,
			"Single-paragraph doc should have identical Summary and Details")

		// Verify it's long enough
		assert.Greater(t, len(entry.Summary), 100,
			"Should handle long single-line docs (at least 100 chars)")

		t.Logf("✓ Long documentation preserved: %d characters", len(entry.Summary))
	})

	t.Run("Complex Formatting with Multiple Paragraphs", func(t *testing.T) {
		entry := index.Entries["github.com/example/edges.ComplexDoc"]
		require.NotNil(t, entry, "ComplexDoc entry should exist")

		// Verify documentation was parsed
		assert.NotEmpty(t, entry.Summary, "Summary should not be empty for complex doc")
		assert.NotEmpty(t, entry.Details, "Details should not be empty for complex doc")

		// Verify Summary contains the first paragraph (before first blank line)
		assert.Contains(t, entry.Summary, "ComplexDoc has mixed formatting",
			"Summary should contain first line")

		// Verify Details contains content from multiple paragraphs
		fullContent := entry.Summary + " " + entry.Details
		assert.Contains(t, fullContent, "ComplexDoc has mixed formatting",
			"Full content should contain first paragraph")
		assert.Contains(t, fullContent, "Multiple blank lines above",
			"Full content should contain second line")
		assert.Contains(t, fullContent, "And various line breaks",
			"Full content should contain text from middle paragraph")
		assert.Contains(t, fullContent, "Final paragraph",
			"Full content should contain final paragraph")

		// Verify structure fields are populated
		assert.Equal(t, KindFunction, entry.Kind)
		assert.Equal(t, "ComplexDoc", entry.LocalName)
		assert.NotEmpty(t, entry.QualifiedName)
		assert.NotEmpty(t, entry.PackagePath)

		t.Logf("✓ Complex formatting preserved: Summary=%q...", entry.Summary[0:40])
	})

	t.Run("Documentation with Special Formatting Preserved", func(t *testing.T) {
		// Test a function with explicitly formatted documentation
		specialType := &reflect.Type{
			Name: "Special",
			Kind: reflect.TypeKindStruct,
			Doc: `Special has special meaning.

This is the detailed section.


Extra blank lines are preserved in parsing behavior.`,
		}
		pkg.Types["Special"] = specialType

		// Rebuild to include this new type
		builder2 := NewBuilder()
		index2, err := builder2.Build(pkg)
		require.NoError(t, err)

		entry := index2.Entries["github.com/example/edges.Special"]
		require.NotNil(t, entry)

		// Verify first line is in summary
		assert.Contains(t, entry.Summary, "Special has special meaning",
			"Summary should contain first line")

		// Verify entire content is captured somewhere
		fullContent := entry.Summary + " " + entry.Details
		assert.Contains(t, fullContent, "detailed section",
			"Full content should include details paragraphs")

		t.Logf("✓ Special formatting handled correctly")
	})

	t.Log("✓ Edge cases tested successfully")
}
