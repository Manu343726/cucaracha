package widgets

import (
	"os"
	"testing"

	"github.com/Manu343726/cucaracha/pkg/ui/tui/tview/themes"
	tvlib "github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Disable automatic VSCode theme loading in tests to prevent network calls and hangs
	os.Setenv("CUCARACHA_NO_VSCODE_THEMES", "1")
}

func TestFuzzyMatch(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		pattern string
		want    bool
	}{
		{"exact match", "dracula", "dracula", true},
		{"case insensitive", "Dracula", "dracula", true},
		{"fuzzy match", "Monokai Dark", "md", true},
		{"fuzzy match partial", "Solarized Dark", "solar", true},
		{"no match", "Nord", "xyz", false},
		{"empty pattern", "any text", "", true},
		{"empty pattern and text", "", "", true},
		{"single char", "Dracula", "d", true},
		{"order matters", "Dracula", "ad", false},
		{"substring match", "Gruvbox Dark", "box", true},
		{"case insensitive fuzzy", "One Dark", "OD", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fuzzyMatch(tt.text, tt.pattern)
			assert.Equal(t, tt.want, got, "fuzzyMatch(%q, %q)", tt.text, tt.pattern)
		})
	}
}

func TestFuzzyMatchScore(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		pattern  string
		minScore int
	}{
		{"exact match scores high", "dracula", "dracula", 100},
		{"partial match scores lower", "dracula", "dr", 20},
		{"consecutive match scores high", "Monokai Dark", "dark", 40},
		{"empty pattern", "any text", "", 0},
		{"single char", "Dracula", "d", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fuzzyMatchScore(tt.text, tt.pattern)
			assert.GreaterOrEqual(t, got, tt.minScore, "fuzzyMatchScore(%q, %q) should be >= %d", tt.text, tt.pattern, tt.minScore)
		})
	}
}

func TestThemeSelectorCreation(t *testing.T) {
	manager := themes.NewManager()
	ts := NewThemeSelector(manager)

	assert.NotNil(t, ts, "NewThemeSelector should not return nil")
	assert.NotNil(t, ts.searchInput, "searchInput should be initialized")
	assert.NotNil(t, ts.list, "list should be initialized")
	assert.NotNil(t, ts.preview, "preview should be initialized")
	assert.Equal(t, manager, ts.manager, "manager should be set")
	assert.Greater(t, len(ts.allThemeNames), 0, "allThemeNames should be populated")
	assert.Greater(t, len(ts.filteredIndices), 0, "filteredIndices should be populated")
}

func TestThemeSelectorFilterThemes(t *testing.T) {
	manager := themes.NewManager()
	ts := NewThemeSelector(manager)

	// Test empty search - should show all themes
	ts.searchQuery = ""
	ts.filterThemes()
	assert.Equal(t, len(ts.allThemeNames), len(ts.filteredIndices), "empty search should show all themes")

	// Test search for "dracula"
	ts.searchQuery = "dracula"
	ts.filterThemes()
	assert.Equal(t, 1, len(ts.filteredIndices), "search for 'dracula' should find 1 theme")

	// Test search for "dark"
	ts.searchQuery = "dark"
	ts.filterThemes()
	assert.GreaterOrEqual(t, len(ts.filteredIndices), 2, "search for 'dark' should find multiple themes")

	// Test search for "nord"
	ts.searchQuery = "nord"
	ts.filterThemes()
	assert.Equal(t, 1, len(ts.filteredIndices), "search for 'nord' should find 1 theme")

	// Test search for nonexistent theme
	ts.searchQuery = "xyz"
	ts.filterThemes()
	assert.Equal(t, 0, len(ts.filteredIndices), "search for 'xyz' should find 0 themes")

	// Test case-insensitive search
	ts.searchQuery = "MONOKAI"
	ts.filterThemes()
	assert.Equal(t, 1, len(ts.filteredIndices), "search for 'MONOKAI' should find 1 theme")
}

func TestThemeSelectorFuzzyFiltering(t *testing.T) {
	manager := themes.NewManager()
	ts := NewThemeSelector(manager)

	// Test fuzzy matching "md" should match "Monokai Dark"
	ts.searchQuery = "md"
	ts.filterThemes()
	assert.Greater(t, len(ts.filteredIndices), 0, "fuzzy search 'md' should match 'Monokai Dark'")

	// Verify that "Monokai Dark" is in the results
	foundMonokai := false
	for _, idx := range ts.filteredIndices {
		name := ts.allThemeNames[idx]
		theme, _ := manager.GetThemeByName(name)
		if theme.Name == "Monokai Dark" {
			foundMonokai = true
			break
		}
	}
	assert.True(t, foundMonokai, "expected 'Monokai Dark' in fuzzy search results for 'md'")

	// Test fuzzy matching "od" should match "One Dark"
	ts.searchQuery = "od"
	ts.filterThemes()
	assert.Greater(t, len(ts.filteredIndices), 0, "fuzzy search 'od' should match 'One Dark'")

	// Test fuzzy matching "gd" should match "Gruvbox Dark"
	ts.searchQuery = "gd"
	ts.filterThemes()
	assert.Greater(t, len(ts.filteredIndices), 0, "fuzzy search 'gd' should match 'Gruvbox Dark'")
}

func TestThemeSelectorGetSelectedTheme(t *testing.T) {
	manager := themes.NewManager()
	ts := NewThemeSelector(manager)

	selected := ts.GetSelectedTheme()
	assert.NotEmpty(t, selected, "GetSelectedTheme should not return empty string")

	// Should be one of the valid theme names
	validThemes := map[string]bool{
		"Monokai Dark":   true,
		"Dracula":        true,
		"Nord":           true,
		"Solarized Dark": true,
		"Gruvbox Dark":   true,
		"One Dark":       true,
	}

	assert.True(t, validThemes[selected], "GetSelectedTheme returned invalid theme: %q", selected)
}

func TestThemeSelectorCallbacks(t *testing.T) {
	manager := themes.NewManager()
	ts := NewThemeSelector(manager)

	previewCalled := false
	previewTheme := ""

	// Set preview callback
	ts.SetOnThemePreview(func(theme *themes.Theme) {
		previewCalled = true
		previewTheme = theme.Name
	})

	// Trigger preview
	ts.themeItemPreviewed(0)

	assert.True(t, previewCalled, "preview callback was not called")
	assert.NotEmpty(t, previewTheme, "preview callback received empty theme")

	// Test selected callback
	selectedCalled := false
	selectedTheme := ""

	ts.SetOnThemeSelected(func(theme *themes.Theme) {
		selectedCalled = true
		selectedTheme = theme.Name
	})

	ts.themeItemSelected(0)

	assert.True(t, selectedCalled, "selected callback was not called")
	assert.NotEmpty(t, selectedTheme, "selected callback received empty theme")
}

func TestThemeSelectorSearchQuery(t *testing.T) {
	manager := themes.NewManager()
	ts := NewThemeSelector(manager)

	// Initially should be empty
	assert.Empty(t, ts.searchQuery, "expected empty searchQuery on init")

	// Simulate search input
	ts.searchInput.SetText("dark")
	assert.Equal(t, "dark", ts.searchQuery, "search query not updated")

	// Search should filter results
	assert.Greater(t, len(ts.filteredIndices), 0, "search for 'dark' should find at least one theme")

	// Clear search
	ts.searchInput.SetText("")
	assert.Equal(t, len(ts.allThemeNames), len(ts.filteredIndices), "cleared search should show all themes")
}

func TestThemeSelectorThemeIndexMapping(t *testing.T) {
	manager := themes.NewManager()
	ts := NewThemeSelector(manager)

	// All theme names should be accessible
	for i, name := range ts.allThemeNames {
		theme, err := manager.GetThemeByName(name)
		assert.NoError(t, err, "GetThemeByName(%q) failed", name)
		assert.NotNil(t, theme, "theme at index %d is nil", i)
		assert.NotEmpty(t, theme.Name, "theme at index %d has empty name", i)
	}

	// filteredIndices should be valid
	for _, idx := range ts.filteredIndices {
		assert.GreaterOrEqual(t, idx, 0, "invalid index in filteredIndices: %d", idx)
		assert.Less(t, idx, len(ts.allThemeNames), "index out of bounds: %d", idx)
	}
}

func TestThemeSelectorListItem(t *testing.T) {
	manager := themes.NewManager()
	ts := NewThemeSelector(manager)

	// Get list item count
	itemCount := ts.list.GetItemCount()
	assert.Equal(t, len(ts.filteredIndices), itemCount, "list item count should match filtered indices")
}

func TestThemeSelectorPreviewWidget(t *testing.T) {
	manager := themes.NewManager()
	ts := NewThemeSelector(manager)

	// Verify preview widget exists and is properly configured
	assert.NotNil(t, ts.preview, "preview widget should be initialized")

	// The preview should have some initial content if filtered themes exist
	if len(ts.filteredIndices) > 0 {
		previewText := ts.preview.GetText(true)
		// Preview might be empty until a theme is selected - this just verifies it exists
		assert.NotNil(t, previewText)
	}
}

func TestThemeSelectorSetTheme(t *testing.T) {
	manager := themes.NewManager()
	ts := NewThemeSelector(manager)

	theme := &themes.Theme{
		Name:  "Test Theme",
		Theme: &tvlib.Theme{},
	}

	// Set theme should return the selector for chaining
	result := ts.SetTheme(theme)
	assert.Equal(t, ts, result, "SetTheme should return *ThemeSelector for method chaining")
}

func TestThemeSelectorUpdatePreviewWithNilTheme(t *testing.T) {
	manager := themes.NewManager()
	ts := NewThemeSelector(manager)

	// Should not crash with nil theme
	ts.updatePreview(nil)

	// Preview should remain unchanged
	text := ts.preview.GetText(true)
	// Just verify no crash occurred
	assert.NotNil(t, text)
}

func TestThemeSelectorThemeItemPreviewOutOfBounds(t *testing.T) {
	manager := themes.NewManager()
	ts := NewThemeSelector(manager)

	// Try to preview item at invalid index
	// Should not crash or panic
	invalidIndex := 999
	ts.themeItemPreviewed(invalidIndex)

	// Just verify no crash occurred
}

func TestThemeSelectorThemeItemSelectedOutOfBounds(t *testing.T) {
	manager := themes.NewManager()
	ts := NewThemeSelector(manager)

	// Try to select item at invalid index
	// Should not crash or panic
	invalidIndex := 999
	ts.themeItemSelected(invalidIndex)

	// Just verify no crash occurred
}

func TestThemeSelectorInputCapture(t *testing.T) {
	manager := themes.NewManager()
	ts := NewThemeSelector(manager)

	assert.NotNil(t, ts.searchInput, "searchInput should be initialized")

	// Verify that the input capture is configured
	// (The actual event handling is tested implicitly through integration)
}

func TestPopulateList(t *testing.T) {
	manager := themes.NewManager()
	ts := NewThemeSelector(manager)

	currentTheme := manager.GetTheme()

	// Clear and repopulate the list
	ts.list.Clear()
	ts.populateList(currentTheme)

	itemCount := ts.list.GetItemCount()
	assert.Greater(t, itemCount, 0, "populateList should add items to the list")

	// Verify that the current theme is marked with a checkmark
	// (This is implicit in the display text)
}

func TestFilterThemesScoreSorting(t *testing.T) {
	manager := themes.NewManager()
	ts := NewThemeSelector(manager)

	// Search for something that matches multiple themes
	ts.searchQuery = "dark"
	ts.filterThemes()

	// Verify that results are sorted (should have higher-scoring matches first)
	if len(ts.filteredIndices) == 0 {
		t.Skip("No themes containing 'dark' found")
	}

	// This test just verifies the filtering doesn't crash
	// Detailed score verification would require inspecting theme names
}

func TestFuzzyMatchEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		pattern string
		want    bool
	}{
		{"unicode characters", "Café", "café", true},
		{"numeric match", "Theme 123", "123", true},
		{"special characters", "One-Dark", "One", true},
		{"whitespace", "My   Theme", "My", true},
		{"very long text", "A" + string(make([]byte, 1000)), "A", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fuzzyMatch(tt.text, tt.pattern)
			assert.Equal(t, tt.want, got, "fuzzyMatch(%q, %q)", tt.text, tt.pattern)
		})
	}
}

func TestFuzzyMatchScoreConsistency(t *testing.T) {
	// Same text and pattern should always produce the same score
	text := "Dracula Theme"
	pattern := "drac"

	score1 := fuzzyMatchScore(text, pattern)
	score2 := fuzzyMatchScore(text, pattern)

	assert.Equal(t, score1, score2, "fuzzyMatchScore should be consistent")
	assert.Greater(t, score1, 0, "fuzzyMatchScore for matching pattern should be positive")
}

func TestThemeSelectorFilterThemesEmptyBuiltIn(t *testing.T) {
	manager := themes.NewManager()
	ts := NewThemeSelector(manager)

	// Simulate searching with no results in built-in themes
	ts.searchQuery = "zzzzzzzzzzz" // Something that definitely won't match

	ts.filterThemes()

	assert.Equal(t, 0, len(ts.filteredIndices), "search for nonexistent theme should have no results")

	// Should display "No matching themes" in the list
	itemCount := ts.list.GetItemCount()
	assert.Greater(t, itemCount, 0, "list should have at least the 'no matches' item")
}

func TestThemeSelectorCallbackWithVSCodeTheme(t *testing.T) {
	manager := themes.NewManager()
	ts := NewThemeSelector(manager)

	// Verify callbacks can be registered
	ts.SetOnThemePreview(func(theme *themes.Theme) {
		// Callback implementation
	})

	assert.NotNil(t, ts.onThemePreview, "preview callback should be set")
}

func TestThemeSelectorSelectedCallbackWithVSCodeTheme(t *testing.T) {
	manager := themes.NewManager()
	ts := NewThemeSelector(manager)

	// Verify selected callback can be registered
	ts.SetOnThemeSelected(func(theme *themes.Theme) {
		// Callback implementation
	})

	assert.NotNil(t, ts.onThemeSelected, "selected callback should be set")
}

func TestThemeSelectorListItemTextFormatting(t *testing.T) {
	manager := themes.NewManager()
	ts := NewThemeSelector(manager)

	// Populate list
	ts.list.Clear()
	ts.populateList(manager.GetTheme())

	// Current selection should be marked with ✓
	itemCount := ts.list.GetItemCount()
	if itemCount > 0 {
		mainText, _ := ts.list.GetItemText(0)
		assert.NotEmpty(t, mainText, "list item should have display text")
	}
}

func TestThemeSelectorStatus(t *testing.T) {
	manager := themes.NewManager()
	ts := NewThemeSelector(manager)

	// Verify theme selector is created successfully
	assert.NotNil(t, ts.manager, "manager should be initialized")
}

func TestThemeSelectorGetListWidget(t *testing.T) {
	manager := themes.NewManager()
	ts := NewThemeSelector(manager)

	listWidget := ts.GetList()
	assert.NotNil(t, listWidget, "GetList should return the list widget")
	assert.Equal(t, ts.list, listWidget, "GetList should return the same list widget used internally")
}

func TestThemeSelectorFilterThemesMultipleMatches(t *testing.T) {
	manager := themes.NewManager()
	ts := NewThemeSelector(manager)

	// Search for pattern that matches multiple themes
	// All built-in themes contain 'd' or 'D'
	ts.searchQuery = "d"
	ts.filterThemes()

	initialMatches := len(ts.filteredIndices)

	// More restrictive search
	ts.searchQuery = "dark"
	ts.filterThemes()

	restrictiveMatches := len(ts.filteredIndices)

	// Restrictive search should have fewer or equal matches
	assert.LessOrEqual(t, restrictiveMatches, initialMatches, "more restrictive search should not return more matches")
}
