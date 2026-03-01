package debugger

import (
	"strings"
	"testing"

	"github.com/Manu343726/cucaracha/pkg/docs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDocumentationIndexLoads verifies that the embedded documentation index loads correctly
func TestDocumentationIndexLoads(t *testing.T) {
	require.NotNil(t, Documentation, "Documentation index is nil - failed to load embedded documentation")
	require.NotNil(t, Documentation.Entries, "Documentation entries map is nil")
	require.NotEmpty(t, Documentation.Entries, "Documentation entries map is empty - no documentation was generated")

	t.Logf("Documentation index loaded with %d entries", len(Documentation.Entries))
}

// TestDocumentationIndexStructure verifies the overall structure of the documentation index
func TestDocumentationIndexStructure(t *testing.T) {
	assert.NotNil(t, Documentation.ByPackage, "Documentation.ByPackage is nil")
	assert.NotNil(t, Documentation.ByKind, "Documentation.ByKind is nil")
	assert.NotNil(t, Documentation.References, "Documentation.References is nil")
	assert.NotEmpty(t, Documentation.Metadata.Version, "Documentation metadata version is empty")
	assert.NotEmpty(t, Documentation.Metadata.PackagesIndexed, "No packages were indexed in documentation metadata")

	t.Logf("Indexed packages: %v", Documentation.Metadata.PackagesIndexed)
}

// TestDebuggerCommandMethodsDocumented verifies all DebuggerCommands interface methods have documentation
func TestDebuggerCommandMethodsDocumented(t *testing.T) {
	// List of all methods in the DebuggerCommands interface
	requiredMethods := []string{
		"Step",
		"Continue",
		"Run",
		"Interrupt",
		"Break",
		"Watch",
		"RemoveBreakpoint",
		"RemoveWatchpoint",
		"List",
		"Disasm",
		"CurrentInstruction",
		"Memory",
		"Source",
		"CurrentSource",
		"Eval",
		"Info",
		"Registers",
		"Stack",
		"Vars",
		"Symbols",
		"Reset",
		"Restart",
		"LoadSystem",
		"LoadProgram",
		"LoadSystemFromFile",
		"LoadSystemFromEmbedded",
		"LoadProgramFromFile",
		"LoadRuntime",
		"Load",
	}

	packagePath := "github.com/Manu343726/cucaracha/pkg/ui/debugger"
	interfaceName := "DebuggerCommands"

	missingMethods := []string{}
	for _, methodName := range requiredMethods {
		// Methods are now documented as part of the interface: pkg.Interface.Method
		qualName := packagePath + "." + interfaceName + "." + methodName
		entry, exists := Documentation.Entries[qualName]

		if !exists {
			missingMethods = append(missingMethods, methodName)
			continue
		}

		assert.NotNil(t, entry, "Documentation entry for %s is nil", methodName)
		t.Logf("✓ %s documented (kind: %s)", methodName, entry.Kind)
	}

	if len(missingMethods) > 0 {
		t.Logf("\n⚠️  ISSUE: %d interface methods are not documented (docs package does not process interface methods):", len(missingMethods))
		for _, m := range missingMethods {
			t.Logf("  - %s", m)
		}
	}
	assert.Empty(t, missingMethods, "Interface methods should be documented")
}

// TestCommandArgumentTypesDocumented verifies that command argument types are documented
func TestCommandArgumentTypesDocumented(t *testing.T) {
	// These types have corresponding methods with arguments
	requiredArgTypes := []string{
		"StepArgs",
		"BreakArgs",
		"WatchArgs",
		"RemoveBreakpointArgs",
		"RemoveWatchpointArgs",
		"DisasmArgs",
		"MemoryArgs",
		"SourceArgs",
		"CurrentSourceArgs",
		"EvalArgs",
		"InfoArgs",
		"SymbolsArgs",
		"LoadSystemArgs",
		"LoadProgramArgs",
		"LoadSystemFromFileArgs",
		"LoadProgramFromFileArgs",
		"LoadRuntimeArgs",
		"LoadArgs",
	}

	packagePath := "github.com/Manu343726/cucaracha/pkg/ui/debugger"

	for _, typeName := range requiredArgTypes {
		qualName := packagePath + "." + typeName
		entry, exists := Documentation.Entries[qualName]

		require.True(t, exists, "Documentation entry for argument type %s should exist", typeName)
		require.NotNil(t, entry, "Documentation entry for %s should not be nil", typeName)
		assert.Equal(t, docs.KindType, entry.Kind, "Expected %s to be documented as type, got %s", typeName, entry.Kind)

		t.Logf("✓ %s documented", typeName)
	}
}

// TestCommandResultTypesDocumented verifies that command result types are documented
func TestCommandResultTypesDocumented(t *testing.T) {
	// These types are return values from command methods
	requiredResultTypes := []string{
		"ExecutionResult",
		"BreakResult",
		"WatchResult",
		"RemoveBreakpointResult",
		"RemoveWatchpointResult",
		"ListResult",
		"DisasmResult",
		"CurrentInstructionResult",
		"MemoryResult",
		"SourceResult",
		"EvalResult",
		"InfoResult",
		"RegistersResult",
		"StackResult",
		"VarsResult",
		"SymbolsResult",
		"LoadSystemFromEmbeddedResult",
		"LoadSystemFromFileResult",
		"LoadProgramFromFileResult",
		"LoadRuntimeResult",
		"LoadResult",
	}

	packagePath := "github.com/Manu343726/cucaracha/pkg/ui/debugger"

	for _, typeName := range requiredResultTypes {
		qualName := packagePath + "." + typeName
		entry, exists := Documentation.Entries[qualName]

		require.True(t, exists, "Documentation entry for result type %s should exist", typeName)
		require.NotNil(t, entry, "Documentation entry for %s should not be nil", typeName)
		assert.Equal(t, docs.KindType, entry.Kind, "Expected %s to be documented as type, got %s", typeName, entry.Kind)

		t.Logf("✓ %s documented", typeName)
	}
}

// TestDocumentationEntryFields verifies that documentation entries have required fields populated
func TestDocumentationEntryFields(t *testing.T) {
	packagePath := "github.com/Manu343726/cucaracha/pkg/ui/debugger"

	// Check a few representative entries that should exist
	testCases := []string{
		"ExecutionResult",
		"StepArgs",
		"BreakArgs",
	}

	for _, itemName := range testCases {
		qualName := packagePath + "." + itemName
		entry, exists := Documentation.Entries[qualName]

		if !exists {
			t.Logf("Skipping %s (not found in documentation)", itemName)
			continue
		}

		t.Run(itemName, func(t *testing.T) {
			// Check required fields
			assert.NotEmpty(t, entry.QualifiedName, "QualifiedName should not be empty")
			assert.Equal(t, qualName, entry.QualifiedName, "QualifiedName mismatch")

			assert.NotEmpty(t, entry.LocalName, "LocalName should not be empty")
			assert.Equal(t, itemName, entry.LocalName, "LocalName mismatch")

			assert.NotEmpty(t, entry.Kind, "Kind should not be empty")

			assert.NotEmpty(t, entry.PackagePath, "PackagePath should not be empty")
			assert.Equal(t, packagePath, entry.PackagePath, "PackagePath mismatch")

			// Check that at least one of summary or details is populated
			hasContent := entry.Summary != "" || entry.Details != ""
			if !hasContent {
				t.Logf("WARNING: No documentation content found for %s (summary, details all empty)", itemName)
			}

			t.Logf("Fields OK - Summary: %q, Details: %.50s...",
				entry.Summary, entry.Details)
		})
	}
}

// TestDocumentationByPackageOrganization verifies entries are correctly organized by package
func TestDocumentationByPackageOrganization(t *testing.T) {
	packagePath := "github.com/Manu343726/cucaracha/pkg/ui/debugger"

	entries, exists := Documentation.ByPackage[packagePath]

	require.True(t, exists, "Package %s should be found in ByPackage map (ISSUE: docs package should use full qualified package path, not %q)", packagePath, ".")
	require.NotEmpty(t, entries, "Package should have entries in ByPackage map")

	t.Logf("Package has %d entries", len(entries))

	// Verify all entries in ByPackage actually exist
	for i, entryKey := range entries {
		if i >= 5 { // Just check first 5 to keep output manageable
			t.Logf("... and %d more entries", len(entries)-5)
			break
		}

		entry, exists := Documentation.Entries[entryKey]
		require.True(t, exists, "Entry key %q listed in ByPackage should exist in Entries map", entryKey)

		assert.Equal(t, packagePath, entry.PackagePath, "Entry %s PackagePath should match", entryKey)

		t.Logf("✓ Entry: %s (kind: %s)", entry.LocalName, entry.Kind)
	}
}

// TestDocumentationByKindOrganization verifies entries are correctly organized by kind
func TestDocumentationByKindOrganization(t *testing.T) {
	if len(Documentation.ByKind) == 0 {
		t.Errorf("ByKind map is empty (ISSUE: docs package should organize entries by kind)")
		return
	}

	t.Logf("Documentation organized by %d kinds", len(Documentation.ByKind))

	for kind, entries := range Documentation.ByKind {
		t.Logf("Kind %q: %d entries", kind, len(entries))

		if len(entries) == 0 {
			t.Logf("WARNING: Kind %q has no entries", kind)
			continue
		}

		// Verify entries match their declared kind
		for i, entryKey := range entries {
			if i >= 3 { // Just check first 3
				break
			}

			entry, exists := Documentation.Entries[entryKey]
			if !exists {
				t.Errorf("Entry key %q in ByKind[%s] but not found in Entries (ISSUE: docs package index inconsistency)",
					entryKey, kind)
				continue
			}

			if entry.Kind != docs.EntryKind(kind) {
				t.Errorf("Entry %s has kind %s but listed under ByKind[%s] (ISSUE: docs package kind mismatch)",
					entryKey, entry.Kind, kind)
			}
		}
	}
}

// TestDocumentationReferences verifies the references structure is populated correctly
func TestDocumentationReferences(t *testing.T) {
	if len(Documentation.References) == 0 {
		t.Logf("No references in documentation (may be intended if ResolveReferences was disabled)")
		return
	}

	t.Logf("Documentation has %d entries with references", len(Documentation.References))

	// Sample a few references
	sampleCount := 0
	for sourceKey, targetKeys := range Documentation.References {
		if sampleCount >= 3 {
			break
		}

		if len(targetKeys) == 0 {
			continue
		}

		sourceEntry, exists := Documentation.Entries[sourceKey]
		if !exists {
			t.Errorf("Reference source %q not found in Entries (ISSUE: docs package reference tracking)", sourceKey)
			continue
		}

		for _, targetKey := range targetKeys {
			targetEntry, exists := Documentation.Entries[targetKey]
			if !exists {
				t.Logf("Reference target %q not found in Entries for %s (may be external reference)",
					targetKey, sourceKey)
				continue
			}

			t.Logf("✓ %s references %s", sourceEntry.LocalName, targetEntry.LocalName)
		}

		sampleCount++
	}
}

// TestDocumentationForBasicCommand verifies a complete example of Step command documentation
func TestDocumentationForBasicCommand(t *testing.T) {
	packagePath := "github.com/Manu343726/cucaracha/pkg/ui/debugger"
	interfaceName := "DebuggerCommands"
	qualName := packagePath + "." + interfaceName + ".Step"

	entry, exists := Documentation.Entries[qualName]
	require.True(t, exists, "Step method should be documented as interface method with qualified name: %q", qualName)

	assert.Equal(t, "Step", entry.LocalName)
	assert.Equal(t, packagePath, entry.PackagePath)
	assert.Equal(t, docs.KindInterfaceMethod, entry.Kind, "Step should be documented as interfaceMethod")

	t.Logf("✓ Step documented - Summary: %q, Kind: %s", entry.Summary, entry.Kind)
}

// TestDocumentationForArgumentType verifies documentation for an argument type
func TestDocumentationForArgumentType(t *testing.T) {
	packagePath := "github.com/Manu343726/cucaracha/pkg/ui/debugger"
	qualName := packagePath + ".StepArgs"

	entry, exists := Documentation.Entries[qualName]
	require.True(t, exists, "StepArgs documentation should exist")

	assert.Equal(t, docs.KindType, entry.Kind)
	assert.Equal(t, "StepArgs", entry.LocalName)

	t.Logf("✓ StepArgs documented - Kind: %s, QualifiedName: %s", entry.Kind, entry.QualifiedName)
}

// TestDocumentationEntriesConsistency checks for internal consistency of the documentation index
func TestDocumentationEntriesConsistency(t *testing.T) {
	packagePath := "github.com/Manu343726/cucaracha/pkg/ui/debugger"
	inconsistencies := 0

	// Check for the package path issue
	if len(Documentation.ByPackage["."]) > 0 && len(Documentation.ByPackage[packagePath]) == 0 {
		t.Logf("\n⚠️  CRITICAL ISSUE: Docs builder uses \".\" as package path instead of full path %q", packagePath)
		inconsistencies++
	}

	// Check all entries in Entries map
	for qualName, entry := range Documentation.Entries {
		require.NotNil(t, entry, "Entry %s should not be nil", qualName)

		// Check QualifiedName matches key
		assert.Equal(t, qualName, entry.QualifiedName, "Entry key should match QualifiedName")

		// Check that entry is in ByPackage
		if entry.PackagePath != "" {
			found := false
			pkgEntries, pkgExists := Documentation.ByPackage[entry.PackagePath]
			if pkgExists {
				for _, key := range pkgEntries {
					if key == qualName {
						found = true
						break
					}
				}
			}

			if !found {
				t.Logf("ISSUE: Entry %s not found in ByPackage[%s]", qualName, entry.PackagePath)
				inconsistencies++
			}
		}

		// Check that entry is in ByKind
		if entry.Kind != "" {
			found := false
			kindEntries, kindExists := Documentation.ByKind[string(entry.Kind)]
			if kindExists {
				for _, key := range kindEntries {
					if key == qualName {
						found = true
						break
					}
				}
			}

			if !found {
				t.Logf("ISSUE: Entry %s not found in ByKind[%s]", qualName, entry.Kind)
				inconsistencies++
			}
		}
	}

	if inconsistencies > 0 {
		t.Logf("\nFound %d inconsistencies in documentation index", inconsistencies)
	} else {
		t.Logf("✓ All entries are consistent in the documentation index")
	}
}

// TestDocumentationVersionMetadata verifies metadata is properly set
func TestDocumentationVersionMetadata(t *testing.T) {
	assert.NotEmpty(t, Documentation.Metadata.Version, "Metadata.Version should not be empty")
	t.Logf("Documentation version: %s", Documentation.Metadata.Version)

	assert.NotEmpty(t, Documentation.Metadata.PackagesIndexed, "Metadata.PackagesIndexed should not be empty")
	t.Logf("Indexed packages: %d", len(Documentation.Metadata.PackagesIndexed))
	for i, pkg := range Documentation.Metadata.PackagesIndexed {
		if i >= 5 {
			t.Logf("... and %d more packages", len(Documentation.Metadata.PackagesIndexed)-5)
			break
		}
		t.Logf("  - %s", pkg)
	}
}

// TestDebugDocumentationEntries is a debug test that prints all documentation entries
// This helps identify what the docs package is actually capturing
func TestDebugDocumentationEntries(t *testing.T) {
	t.Logf("=== All Documentation Entries ===")
	t.Logf("Total entries: %d\n", len(Documentation.Entries))

	for qualName, entry := range Documentation.Entries {
		t.Logf("  %s", qualName)
		t.Logf("    Kind: %s", entry.Kind)
		t.Logf("    LocalName: %s", entry.LocalName)
		t.Logf("    PackagePath: %s", entry.PackagePath)
	}

	t.Logf("\n=== By Package ===")
	for pkgPath, entries := range Documentation.ByPackage {
		t.Logf("Package: %q (%d entries)", pkgPath, len(entries))
	}

	t.Logf("\n=== By Kind ===")
	for kind, entries := range Documentation.ByKind {
		t.Logf("Kind %q: %d entries", kind, len(entries))
	}
}

// TestDocumentationCompleteness provides a summary of documentation coverage
func TestDocumentationCompleteness(t *testing.T) {
	packagePath := "github.com/Manu343726/cucaracha/pkg/ui/debugger"

	var typeCount, functionCount, methodCount, constantCount int
	var entriesWithSummary, entriesWithDetails int
	var entriesWithExamples int

	for qualName, entry := range Documentation.Entries {
		if !strings.HasPrefix(qualName, packagePath) {
			continue
		}

		switch entry.Kind {
		case docs.KindType:
			typeCount++
		case docs.KindFunction:
			functionCount++
		case docs.KindMethod:
			methodCount++
		case docs.KindConstant:
			constantCount++
		}

		if entry.Summary != "" {
			entriesWithSummary++
		}
		if entry.Details != "" {
			entriesWithDetails++
		}
		if len(entry.Examples) > 0 {
			entriesWithExamples++
		}
	}

	totalEntries := typeCount + functionCount + methodCount + constantCount
	t.Logf("Documentation Coverage Summary:")
	t.Logf("  Total entries: %d", totalEntries)
	t.Logf("    - Types: %d", typeCount)
	t.Logf("    - Functions: %d", functionCount)
	t.Logf("    - Methods: %d", methodCount)
	t.Logf("    - Constants: %d", constantCount)
	t.Logf("")
	t.Logf("Content Coverage:")
	if totalEntries > 0 {
		t.Logf("  - With Summary: %d (%.1f%%)", entriesWithSummary, float64(entriesWithSummary)/float64(totalEntries)*100)
		t.Logf("  - With Details: %d (%.1f%%)", entriesWithDetails, float64(entriesWithDetails)/float64(totalEntries)*100)
		t.Logf("  - With Examples: %d (%.1f%%)", entriesWithExamples, float64(entriesWithExamples)/float64(totalEntries)*100)
	}
}

// Helper to extract method names from API
func getAllMethodNames() []string {
	return []string{
		"Step", "Continue", "Run", "Interrupt", "Break", "Watch",
		"RemoveBreakpoint", "RemoveWatchpoint", "List", "Disasm",
		"CurrentInstruction", "Memory", "Source", "CurrentSource",
		"Eval", "Info", "Registers", "Stack", "Vars", "Symbols",
		"Reset", "Restart", "LoadSystem", "LoadProgram",
		"LoadSystemFromFile", "LoadSystemFromEmbedded", "LoadProgramFromFile",
		"LoadRuntime", "Load",
	}
}

// TestMethodDocumentationContent verifies that extracted method documentation content is accurate
func TestMethodDocumentationContent(t *testing.T) {
	packagePath := "github.com/Manu343726/cucaracha/pkg/ui/debugger"
	interfaceName := "DebuggerCommands"

	// Test methods with specific expected documentation from godoc comments
	methodTests := []struct {
		methodName          string
		expectedKeywords    []string // Keywords that must appear in summary or details
		minSummaryLength    int
		shouldMentionArgs   bool      // Methods that take args should mention the arg type
		shouldMentionResult bool      // Methods should mention their return type
		resultTypeName      string    // Expected result type name to reference
	}{
		{
			methodName:          "Step",
			expectedKeywords:    []string{"step", "instruction", "execution"},
			minSummaryLength:    20,
			shouldMentionArgs:   true,
			shouldMentionResult: true,
			resultTypeName:      "ExecutionResult",
		},
		{
			methodName:          "Continue",
			expectedKeywords:    []string{"resumes", "execution"},
			minSummaryLength:    20,
			shouldMentionArgs:   false,
			shouldMentionResult: true,
			resultTypeName:      "ExecutionResult",
		},
		{
			methodName:          "Break",
			expectedKeywords:    []string{"breakpoint", "code"},
			minSummaryLength:    20,
			shouldMentionArgs:   true,
			shouldMentionResult: true,
			resultTypeName:      "BreakResult",
		},
		{
			methodName:          "List",
			expectedKeywords:    []string{"active", "breakpoints"},
			minSummaryLength:    15,
			shouldMentionArgs:   false,
			shouldMentionResult: true,
			resultTypeName:      "ListResult",
		},
		{
			methodName:          "Registers",
			expectedKeywords:    []string{"cpu", "register"},
			minSummaryLength:    15,
			shouldMentionArgs:   false,
			shouldMentionResult: true,
			resultTypeName:      "RegistersResult",
		},
	}

	for _, tc := range methodTests {
		t.Run(tc.methodName, func(t *testing.T) {
			qualName := packagePath + "." + interfaceName + "." + tc.methodName
			entry, exists := Documentation.Entries[qualName]
			require.True(t, exists, "Method %s should be documented", tc.methodName)
			require.NotNil(t, entry, "Entry for %s should not be nil", tc.methodName)

			// 1. Verify summary is not empty
			assert.NotEmpty(t, entry.Summary, "Method %s must have a summary", tc.methodName)

			// 2. Verify summary meets minimum length
			assert.GreaterOrEqual(t, len(entry.Summary), tc.minSummaryLength,
				"Method %s summary is too short (%d chars), expected at least %d chars: %q",
				tc.methodName, len(entry.Summary), tc.minSummaryLength, entry.Summary)

			// 3. Verify key concepts are mentioned
			documentationText := strings.ToLower(entry.Summary + " " + entry.Details)
			foundKeywords := 0
			for _, keyword := range tc.expectedKeywords {
				if strings.Contains(documentationText, strings.ToLower(keyword)) {
					foundKeywords++
				}
			}
			assert.GreaterOrEqual(t, foundKeywords, 1,
				"Method %s documentation should mention at least one of %v (documentation: %s)",
				tc.methodName, tc.expectedKeywords, entry.Summary)

			// 4. Verify cross-references to types
			if tc.shouldMentionResult {
				resultQual := packagePath + "." + tc.resultTypeName
				assert.True(t, Documentation.Entries[resultQual] != nil,
					"Method %s result type %s should exist", tc.methodName, tc.resultTypeName)
			}

			t.Logf("✓ %s: %d chars, keywords found: %d/%d - %s",
				tc.methodName, len(entry.Summary), foundKeywords, len(tc.expectedKeywords), entry.Summary)
		})
	}
}

// TestArgumentTypeDocumentationAccuracy verifies that argument types are accurately documented
func TestArgumentTypeDocumentationAccuracy(t *testing.T) {
	packagePath := "github.com/Manu343726/cucaracha/pkg/ui/debugger"

	// Test argument types with specific expected content from godoc comments
	argTypeTests := []struct {
		typeName           string
		expectedSummary    string // Must be included in Summary
		minSummaryLength   int
		shouldHaveDetails  bool
	}{
		{
			typeName:         "StepArgs",
			expectedSummary:  "specifies how to execute a step operation",
			minSummaryLength: 20,
			shouldHaveDetails: true,
		},
		{
			typeName:         "BreakArgs",
			expectedSummary:  "specifies where to set a breakpoint",
			minSummaryLength: 20,
			shouldHaveDetails: true,
		},
		{
			typeName:         "WatchArgs",
			expectedSummary:  "specifies a memory range",
			minSummaryLength: 20,
			shouldHaveDetails: true,
		},
		{
			typeName:         "SourceArgs",
			expectedSummary:  "specifies parameters for displaying source code",
			minSummaryLength: 25,
			shouldHaveDetails: true,
		},
		{
			typeName:         "DisasmArgs",
			expectedSummary:  "specifies parameters for disassembling",
			minSummaryLength: 25,
			shouldHaveDetails: true,
		},
		{
			typeName:         "EvalArgs",
			expectedSummary:  "specifies an expression to evaluate",
			minSummaryLength: 20,
			shouldHaveDetails: true,
		},
	}

	for _, tc := range argTypeTests {
		t.Run(tc.typeName, func(t *testing.T) {
			qualName := packagePath + "." + tc.typeName
			entry, exists := Documentation.Entries[qualName]
			require.True(t, exists, "Type %s should be documented", tc.typeName)
			require.NotNil(t, entry, "Entry for %s should not be nil", tc.typeName)

			// 1. Verify summary contains expected text
			assert.NotEmpty(t, entry.Summary, "Type %s must have a summary", tc.typeName)
			summaryLower := strings.ToLower(entry.Summary)
			expectedLower := strings.ToLower(tc.expectedSummary)
			assert.True(t, strings.Contains(summaryLower, expectedLower),
				"Type %s summary must contain %q (got: %q)", tc.typeName, tc.expectedSummary, entry.Summary)

			// 2. Verify summary is substantive
			assert.GreaterOrEqual(t, len(entry.Summary), tc.minSummaryLength,
				"Type %s summary is too short (%d chars), expected at least %d chars: %q",
				tc.typeName, len(entry.Summary), tc.minSummaryLength, entry.Summary)

			// 3. Verify it has some documentation
			hasDocumentation := entry.Summary != "" || entry.Details != ""
			assert.True(t, hasDocumentation, "Type %s must have documentation", tc.typeName)

			t.Logf("✓ %s: Summary (%d chars accurate) - %s", tc.typeName, len(entry.Summary), entry.Summary)
			if entry.Details != "" {
				t.Logf("  Additional details: %d chars", len(entry.Details))
			}
		})
	}
}

// TestResultTypeDocumentation verifies that result types are properly documented with accuracy
func TestResultTypeDocumentation(t *testing.T) {
	packagePath := "github.com/Manu343726/cucaracha/pkg/ui/debugger"

	// Test result types with expected content validation
	resultTypeTests := []struct {
		typeName              string
		expectedSummaryKeyword string
		minSummaryLength      int
	}{
		{
			typeName:               "ExecutionResult",
			expectedSummaryKeyword: "execution",
			minSummaryLength:       30,
		},
		{
			typeName:               "BreakResult",
			expectedSummaryKeyword: "breakpoint",
			minSummaryLength:       25,
		},
		{
			typeName:               "WatchResult",
			expectedSummaryKeyword: "watchpoint",
			minSummaryLength:       25,
		},
		{
			typeName:               "SourceResult",
			expectedSummaryKeyword: "source",
			minSummaryLength:       25,
		},
		{
			typeName:               "ListResult",
			expectedSummaryKeyword: "active",
			minSummaryLength:       25,
		},
		{
			typeName:               "EvalResult",
			expectedSummaryKeyword: "value",
			minSummaryLength:       20,
		},
		{
			typeName:               "DisasmResult",
			expectedSummaryKeyword: "disassem",
			minSummaryLength:       20,
		},
	}

	for _, tc := range resultTypeTests {
		t.Run(tc.typeName, func(t *testing.T) {
			qualName := packagePath + "." + tc.typeName
			entry, exists := Documentation.Entries[qualName]
			require.True(t, exists, "Result type %s should be documented", tc.typeName)
			require.NotNil(t, entry, "Entry for %s should not be nil", tc.typeName)

			// 1. Verify summary contains expected keyword
			assert.NotEmpty(t, entry.Summary, "Result type %s must have a summary", tc.typeName)
			summaryLower := strings.ToLower(entry.Summary)
			assert.True(t, strings.Contains(summaryLower, strings.ToLower(tc.expectedSummaryKeyword)),
				"Result type %s summary should mention %q (got: %q)",
				tc.typeName, tc.expectedSummaryKeyword, entry.Summary)

			// 2. Verify summary is substantive
			assert.GreaterOrEqual(t, len(entry.Summary), tc.minSummaryLength,
				"Result type %s summary is too short (%d chars), expected at least %d: %q",
				tc.typeName, len(entry.Summary), tc.minSummaryLength, entry.Summary)

			// 3. Verify it has documentation
			hasDocumentation := entry.Summary != "" || entry.Details != ""
			assert.True(t, hasDocumentation, "Result type %s must have documentation", tc.typeName)

			t.Logf("✓ %s: Summary (%d chars) - %s", tc.typeName, len(entry.Summary), entry.Summary)
			if entry.Details != "" {
				t.Logf("  Additional details: %d chars", len(entry.Details))
			}
		})
	}
}

// TestMethodAndArgsResultConnectionsExist verifies cross-references between methods and their types
func TestMethodAndArgsResultConnectionsExist(t *testing.T) {
	packagePath := "github.com/Manu343726/cucaracha/pkg/ui/debugger"
	interfaceName := "DebuggerCommands"

	// Map of methods to their arg types and result types
	methodConnections := map[string]struct {
		args   string
		result string
	}{
		"Step": {args: "StepArgs", result: "ExecutionResult"},
		"Break": {args: "BreakArgs", result: "BreakResult"},
		"Watch": {args: "WatchArgs", result: "WatchResult"},
		"Disasm": {args: "DisasmArgs", result: "DisasmResult"},
		"Memory": {args: "MemoryArgs", result: "MemoryResult"},
		"Source": {args: "SourceArgs", result: "SourceResult"},
		"CurrentSource": {args: "CurrentSourceArgs", result: "SourceResult"},
		"Eval": {args: "EvalArgs", result: "EvalResult"},
		"Info": {args: "InfoArgs", result: "InfoResult"},
		"Symbols": {args: "SymbolsArgs", result: "SymbolsResult"},
		"RemoveBreakpoint": {args: "RemoveBreakpointArgs", result: "RemoveBreakpointResult"},
		"RemoveWatchpoint": {args: "RemoveWatchpointArgs", result: "RemoveWatchpointResult"},
		"LoadSystemFromFile": {args: "LoadSystemFromFileArgs", result: "LoadSystemFromFileResult"},
		"LoadProgramFromFile": {args: "LoadProgramFromFileArgs", result: "LoadProgramFromFileResult"},
		"LoadRuntime": {args: "LoadRuntimeArgs", result: "LoadRuntimeResult"},
	}

	for methodName, connections := range methodConnections {
		t.Run(methodName, func(t *testing.T) {
			// Check method exists
			methodQual := packagePath + "." + interfaceName + "." + methodName
			methodEntry, exists := Documentation.Entries[methodQual]
			require.True(t, exists, "Method %s should be documented", methodName)
			require.NotNil(t, methodEntry, "Method %s entry should not be nil", methodName)

			// Check args type exists
			argsQual := packagePath + "." + connections.args
			argsExists := false
			if argsEntry, ok := Documentation.Entries[argsQual]; ok {
				argsExists = true
				assert.NotNil(t, argsEntry, "Args entry should not be nil")
			}
			require.True(t, argsExists, "Args type %s for method %s should be documented", connections.args, methodName)

			// Check result type exists
			resultQual := packagePath + "." + connections.result
			resultExists := false
			if resultEntry, ok := Documentation.Entries[resultQual]; ok {
				resultExists = true
				assert.NotNil(t, resultEntry, "Result entry should not be nil")
			}
			require.True(t, resultExists, "Result type %s for method %s should be documented", connections.result, methodName)

			// Verify method summary mentions the arg/result types (via godoc references)
			methodSummaryLower := strings.ToLower(methodEntry.Summary)
			hasBothTypesReferenced := strings.Contains(methodSummaryLower, strings.ToLower(connections.args)) ||
				strings.Contains(methodSummaryLower, strings.ToLower(connections.result))

			if hasBothTypesReferenced {
				t.Logf("✓ %s properly references its argument and/or result types", methodName)
			} else {
				t.Logf("⚠️  %s does not explicitly reference argument/result types in summary (but types exist)", methodName)
			}

			t.Logf("  Args: %s ✓", connections.args)
			t.Logf("  Result: %s ✓", connections.result)
		})
	}
}

// TestDocumentationContentQualityMetrics verifies the quality of extracted documentation
func TestDocumentationContentQualityMetrics(t *testing.T) {
	packagePath := "github.com/Manu343726/cucaracha/pkg/ui/debugger"

	const (
		minSummaryLength = 10
		minDetailsLength = 20
	)

	var (
		totalEntries              = 0
		entriesWithSummary        = 0
		entriesWithDetails        = 0
		entriesWithLinks          = 0
		entriesWithExamples       = 0
		entriesWithSubstantiveDoc = 0  // Has summary AND (details OR links OR examples)
		entriesWithShortSummary   = 0  // Summary < minSummaryLength
		coreTypesInPackage        = 0
	)

	// Map of core types that absolutely need good documentation
	coreTypes := map[string]bool{
		"DebuggerCommands": true,
		"StepArgs": true, "BreakArgs": true, "WatchArgs": true,
		"ExecutionResult": true, "BreakResult": true, "WatchResult": true,
	}

	for qualName, entry := range Documentation.Entries {
		if !strings.HasPrefix(qualName, packagePath) {
			continue
		}

		totalEntries++

		hasSubstantiveSummary := entry.Summary != "" && len(entry.Summary) >= minSummaryLength
		hasSubstantiveDetails := entry.Details != "" && len(entry.Details) >= minDetailsLength

		if entry.Summary != "" {
			entriesWithSummary++
		}
		if entry.Details != "" {
			entriesWithDetails++
		}
		if entry.Links != nil && len(entry.Links) > 0 {
			entriesWithLinks++
		}
		if entry.Examples != nil && len(entry.Examples) > 0 {
			entriesWithExamples++
		}

		// Count entries with substantive documentation
		if hasSubstantiveSummary && (hasSubstantiveDetails || entriesWithLinks > 0 || entriesWithExamples > 0) {
			entriesWithSubstantiveDoc++
		}

		// Check for short summaries
		if entry.Summary != "" && len(entry.Summary) < minSummaryLength {
			entriesWithShortSummary++
			t.Logf("WARNING: %s has short summary (%d chars): %q", entry.LocalName, len(entry.Summary), entry.Summary)
		}

		// Track core types in package
		if coreTypes[entry.LocalName] {
			coreTypesInPackage++
		}
	}

	// Report metrics
	t.Logf("\n=== Documentation Content Quality Metrics ===")
	t.Logf("Total entries analyzed: %d", totalEntries)
	t.Logf("Entries with summaries: %d (%.1f%%)", entriesWithSummary, float64(entriesWithSummary)/float64(totalEntries)*100)
	t.Logf("Entries with details: %d (%.1f%%)", entriesWithDetails, float64(entriesWithDetails)/float64(totalEntries)*100)
	t.Logf("Entries with links: %d", entriesWithLinks)
	t.Logf("Entries with examples: %d", entriesWithExamples)
	t.Logf("")
	t.Logf("Substantive content (summary >= %d chars + details/links/examples): %d entries",
		minSummaryLength, entriesWithSubstantiveDoc)
	t.Logf("Short summaries (< %d chars): %d", minSummaryLength, entriesWithShortSummary)
	t.Logf("Core types in package: %d", coreTypesInPackage)

	// Verify quality thresholds
	percentSummary := float64(entriesWithSummary) / float64(totalEntries)
	assert.Greater(t, percentSummary, 0.9, "At least 90%% of entries should have summaries (got %.1f%%)", percentSummary*100)

	percentDetails := float64(entriesWithDetails) / float64(totalEntries)
	assert.Greater(t, percentDetails, 0.5, "At least 50%% of entries should have details (got %.1f%%)", percentDetails*100)

	percentSubstantive := float64(entriesWithSubstantiveDoc) / float64(totalEntries)
	assert.Greater(t, percentSubstantive, 0.7, "At least 70%% of entries should have substantive documentation (got %.1f%%)", percentSubstantive*100)

	t.Logf("\n✓ Documentation quality meets thresholds")
}

// TestNoEmptyDocumentationForCoreTypes verifies core types are not left undocumented
func TestNoEmptyDocumentationForCoreTypes(t *testing.T) {
	packagePath := "github.com/Manu343726/cucaracha/pkg/ui/debugger"

	// Core types that absolutely must have documentation
	coreTypes := []string{
		"DebuggerCommands", // Interface itself
		"StepArgs", "BreakArgs", "WatchArgs",
		"ExecutionResult", "BreakResult", "WatchResult",
	}

	for _, typeName := range coreTypes {
		t.Run(typeName, func(t *testing.T) {
			qualName := packagePath + "." + typeName
			entry, exists := Documentation.Entries[qualName]
			require.True(t, exists, "Core type %s must be documented", typeName)
			require.NotNil(t, entry, "Documentation entry for %s must not be nil", typeName)

			// Core types should have meaningful summaries
			assert.NotEmpty(t, entry.Summary, "Core type %s must have a summary", typeName)

			// Verify has some documentation content
			hasContent := entry.Summary != "" || entry.Details != ""
			assert.True(t, hasContent, "Core type %s must have documentation content", typeName)

			t.Logf("✓ %s properly documented with summary: %q", typeName, entry.Summary)
		})
	}
}
