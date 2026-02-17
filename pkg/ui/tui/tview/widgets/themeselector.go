package widgets

import (
	"strings"

	"github.com/Manu343726/cucaracha/pkg/ui/tui/tview/themes"
	"github.com/gdamore/tcell/v2"
	tvlib "github.com/rivo/tview"
)

type ThemeChangeCallback func(theme *themes.Theme)

// ThemeSelector is an interactive theme selection widget
type ThemeSelector struct {
	*tvlib.Flex
	searchInput      *tvlib.InputField
	list             *tvlib.List
	preview          *tvlib.TextView
	manager          *themes.Manager
	onThemeSelected  ThemeChangeCallback
	onThemePreview   ThemeChangeCallback
	selectedCallback func()
	allThemeNames    []string // All available theme names
	filteredIndices  []int    // Indices of filtered themes in allThemeNames
	searchQuery      string   // Current search query
}

// NewThemeSelector creates a new theme selector widget
func NewThemeSelector(manager *themes.Manager) *ThemeSelector {
	ts := &ThemeSelector{
		Flex:            tvlib.NewFlex(),
		manager:         manager,
		allThemeNames:   manager.GetThemeNames(),
		filteredIndices: make([]int, 0),
		searchQuery:     "",
	}

	// Create search input field
	ts.searchInput = tvlib.NewInputField().
		SetLabel("Search: ").
		SetFieldWidth(0).
		SetAcceptanceFunc(func(textToCheck string, lastChar rune) bool {
			return true // Accept all input
		})

	ts.searchInput.SetBorder(true).
		SetTitle("Find Theme").
		SetTitleAlign(tvlib.AlignLeft)

	// Create the list of themes
	ts.list = tvlib.NewList().
		SetWrapAround(false) // Don't wrap around

	ts.list.SetBorder(true).
		SetTitle("Select Theme").
		SetTitleAlign(tvlib.AlignLeft)

	currentTheme := manager.GetTheme()

	// Initialize filteredIndices with all theme indices
	for i := range ts.allThemeNames {
		ts.filteredIndices = append(ts.filteredIndices, i)
	}

	// Add themes to the list initially
	ts.populateList(currentTheme)

	// Create the preview area
	ts.preview = tvlib.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(true).
		SetScrollable(true)

	ts.preview.SetBorder(true).
		SetTitle("Preview").
		SetTitleAlign(tvlib.AlignLeft)

	// Set up search input callback for fuzzy filtering
	ts.searchInput.SetChangedFunc(func(text string) {
		ts.searchQuery = text
		ts.filterThemes()
	})

	// Handle keyboard input for search input
	ts.searchInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			// Clear search and keep focus on search
			ts.searchInput.SetText("")
			ts.searchQuery = ""
			ts.filterThemes()
			return nil
		}
		// Allow all other keys to be processed normally (for text input)
		return event
	})

	// Set up list change callback for real-time preview
	ts.list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		ts.themeItemPreviewed(index)
	})

	// Set selected callback for Enter key (tview List handles this automatically)
	ts.list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		ts.themeItemSelected(index)
	})

	// Set initial preview
	if len(ts.filteredIndices) > 0 {
		ts.themeItemPreviewed(0)
	}

	// Layout: search on top, list on left (1/3), preview on right (2/3)
	listAndPreview := tvlib.NewFlex().
		SetDirection(tvlib.FlexColumn).
		AddItem(ts.list, 0, 1, true).
		AddItem(ts.preview, 0, 2, false)

	ts.SetDirection(tvlib.FlexRow).
		AddItem(ts.searchInput, 3, 0, true).
		AddItem(listAndPreview, 0, 1, true)

	// Remove top-level input capture to avoid interference with child widgets
	// Let child widgets (searchInput and list) handle their own input
	ts.SetInputCapture(nil)

	return ts
}

// fuzzyMatch performs fuzzy matching of pattern against text
// Returns true if all characters in pattern appear in order in text (case-insensitive)
func fuzzyMatch(text, pattern string) bool {
	text = strings.ToLower(text)
	pattern = strings.ToLower(pattern)

	if pattern == "" {
		return true
	}

	patternIdx := 0
	for i := 0; i < len(text) && patternIdx < len(pattern); i++ {
		if text[i] == pattern[patternIdx] {
			patternIdx++
		}
	}

	return patternIdx == len(pattern)
}

// fuzzyMatchScore calculates a score for fuzzy matching
// Higher scores indicate better matches (consecutive matches score higher)
func fuzzyMatchScore(text, pattern string) int {
	text = strings.ToLower(text)
	pattern = strings.ToLower(pattern)

	if pattern == "" {
		return 0
	}

	score := 0
	consecutiveMatches := 0
	patternIdx := 0

	for i := 0; i < len(text) && patternIdx < len(pattern); i++ {
		if text[i] == pattern[patternIdx] {
			consecutiveMatches++
			score += consecutiveMatches * 10 // Bonus for consecutive matches
			patternIdx++
		} else {
			consecutiveMatches = 0
		}
	}

	return score
}

// filterThemes filters the theme list based on the search query
func (ts *ThemeSelector) filterThemes() {
	ts.filteredIndices = make([]int, 0)

	// Build list of matching indices
	type scoreMatch struct {
		index int
		score int
	}
	matches := make([]scoreMatch, 0)

	// Search built-in themes
	for i, name := range ts.allThemeNames {
		themeName, _ := ts.manager.GetThemeByName(name)
		displayName := themeName.Name

		if fuzzyMatch(displayName, ts.searchQuery) {
			score := fuzzyMatchScore(displayName, ts.searchQuery)
			matches = append(matches, scoreMatch{i, score})
		}
	}

	// Sort matches by score (descending)
	for i := 0; i < len(matches); i++ {
		maxIdx := i
		for j := i + 1; j < len(matches); j++ {
			if matches[j].score > matches[maxIdx].score {
				maxIdx = j
			}
		}
		matches[i], matches[maxIdx] = matches[maxIdx], matches[i]
		ts.filteredIndices = append(ts.filteredIndices, matches[i].index)
	}

	// Rebuild list
	ts.list.Clear()
	currentTheme := ts.manager.GetTheme()

	// Add built-in themes to list
	currentIndex := 0
	for listIdx, themeIdx := range ts.filteredIndices {
		name := ts.allThemeNames[themeIdx]
		theme, _ := ts.manager.GetThemeByName(name)
		displayText := theme.Name

		if theme == currentTheme {
			displayText = displayText + " ✓"
			currentIndex = listIdx
		}

		ts.list.AddItem(displayText, "", 0, nil)
	}

	if len(ts.filteredIndices) == 0 {
		ts.list.AddItem("No matching themes", "", 0, nil)
		ts.preview.Clear()
		return
	}

	// Set current item and trigger preview
	ts.list.SetCurrentItem(currentIndex)
	ts.themeItemPreviewed(currentIndex)
}

// populateList populates the list with all themes
func (ts *ThemeSelector) populateList(currentTheme *themes.Theme) {
	currentIndex := 0
	for i, name := range ts.allThemeNames {
		theme, _ := ts.manager.GetThemeByName(name)
		displayText := theme.Name
		if theme == currentTheme {
			displayText = displayText + " ✓"
			currentIndex = i
		}
		ts.list.AddItem(displayText, "", 0, nil)
	}

	// Set initial selection
	ts.list.SetCurrentItem(currentIndex)
}

func (ts *ThemeSelector) themeItemSelected(index int) {
	if index >= 0 && index < len(ts.filteredIndices) {
		// Theme selected
		themeIdx := ts.filteredIndices[index]
		themeName := ts.allThemeNames[themeIdx]
		theme, _ := ts.manager.GetThemeByName(themeName)
		ts.manager.SetTheme(themeName)
		if ts.onThemeSelected != nil {
			ts.onThemeSelected(theme)
		}
	}
}

func (ts *ThemeSelector) themeItemPreviewed(index int) {
	if index >= 0 && index < len(ts.filteredIndices) {
		// Theme previewed
		themeIdx := ts.filteredIndices[index]
		themeName := ts.allThemeNames[themeIdx]
		theme, _ := ts.manager.GetThemeByName(themeName)
		ts.updatePreview(theme)
		if ts.onThemePreview != nil {
			ts.onThemePreview(theme)
		}
	}
}

// updatePreview updates the preview with the theme's colors
func (ts *ThemeSelector) updatePreview(theme *themes.Theme) {
	if theme != nil {
		// Convert tcell.Color to hex for display
		colorToHex := func(c tcell.Color) string {
			r, g, b := c.RGB()
			return "#" + "0123456789ABCDEF"[r>>4:r>>4+1] + "0123456789ABCDEF"[r&15:r&15+1] +
				"0123456789ABCDEF"[g>>4:g>>4+1] + "0123456789ABCDEF"[g&15:g&15+1] +
				"0123456789ABCDEF"[b>>4:b>>4+1] + "0123456789ABCDEF"[b&15:b&15+1]
		}

		var preview string

		// Theme header
		primHex := colorToHex(theme.PrimaryTextColor)
		preview = "[" + primHex + "]" + theme.Name + "[-]\n"
		preview += "[" + colorToHex(theme.SecondaryTextColor) + "]" + theme.Description + "[-]\n\n"

		// User I/O Theme
		if theme.UserIO != nil {
			preview += "[" + colorToHex(theme.TitleColor) + "]━━━ User I/O ━━━[-]\n"
			preview += "[" + colorToHex(theme.UserIO.CommandPrompt) + "]→ command[-] "
			preview += "[" + colorToHex(theme.UserIO.Info) + "]ℹ info[-] "
			preview += "[" + colorToHex(theme.UserIO.Success) + "]✓ success[-] "
			preview += "[" + colorToHex(theme.UserIO.Warning) + "]⚠ warning[-] "
			preview += "[" + colorToHex(theme.UserIO.Error) + "]✗ error[-]\n\n"
		}

		// Source Snippet Theme
		if theme.SourceSnippet != nil {
			preview += "[" + colorToHex(theme.TitleColor) + "]━━━ Source Code ━━━[-]\n"
			preview += "[" + colorToHex(theme.SourceSnippet.LineNumber) + "]  42[-] "
			if theme.SourceSnippet.C != nil {
				preview += "[" + colorToHex(theme.SourceSnippet.C.Keyword) + "]mov[-] "
				preview += "[" + colorToHex(theme.SourceSnippet.C.Type) + "]rax[-], "
				preview += "[" + colorToHex(theme.SourceSnippet.C.Number) + "]0x1000[-]\n"
				preview += "[" + colorToHex(theme.SourceSnippet.LineNumber) + "]  43[-] "
				preview += "[" + colorToHex(theme.SourceSnippet.BreakpointMarker) + "]●[-] "
				preview += "[" + colorToHex(theme.SourceSnippet.C.Keyword) + "]call[-] "
				preview += "[" + colorToHex(theme.SourceSnippet.C.Function) + "]function[-] "
				preview += "[" + colorToHex(theme.SourceSnippet.C.Comment) + "]// comment[-]\n"
				preview += "[" + colorToHex(theme.SourceSnippet.LineNumber) + "]  44[-] "
				preview += "[" + colorToHex(theme.SourceSnippet.C.String) + "]\"string\"[-]\n\n"
			}
		}

		// Memory Dump Theme
		if theme.MemoryDump != nil {
			preview += "[" + colorToHex(theme.TitleColor) + "]━━━ Memory Dump ━━━[-]\n"
			preview += "[" + colorToHex(theme.MemoryDump.Address) + "]0x08100000[-] "
			preview += "[" + colorToHex(theme.MemoryDump.HexDump) + "]48 89 E5 FF 15[-] "
			preview += "[" + colorToHex(theme.MemoryDump.ASCII) + "]H..[-]\n"
			preview += "[" + colorToHex(theme.MemoryDump.WatchpointMarker) + "]●[-] "
			preview += "[" + colorToHex(theme.MemoryDump.Address) + "]0x08100010[-] "
			preview += "[" + colorToHex(theme.MemoryDump.HexDump) + "]E8 1A 00 00 00[-] "
			preview += "[" + colorToHex(theme.MemoryDump.ASCII) + "]. ...[-]\n\n"
		}

		// Disassembly Theme
		if theme.Disassembly != nil {
			preview += "[" + colorToHex(theme.TitleColor) + "]━━━ Disassembly ━━━[-]\n"
			preview += "[" + colorToHex(theme.Disassembly.Address) + "]08100000[-]: "
			preview += "[" + colorToHex(theme.Disassembly.Mnemonic) + "]mov[-] "
			preview += "[" + colorToHex(theme.Disassembly.RegisterOperand) + "]rax[-], "
			preview += "[" + colorToHex(theme.Disassembly.ImmediateOperand) + "]0x1000[-]\n"
			preview += "[" + colorToHex(theme.Disassembly.Address) + "]08100005[-]: "
			preview += "[" + colorToHex(theme.Disassembly.Mnemonic) + "]call[-] "
			preview += "[" + colorToHex(theme.Disassembly.ImmediateOperand) + "]0x08101000[-]\n\n"
		}

		// Registers Theme
		if theme.Registers != nil {
			preview += "[" + colorToHex(theme.TitleColor) + "]━━━ Registers ━━━[-]\n"
			preview += "[" + colorToHex(theme.Registers.RegisterName) + "]rax[-]: "
			preview += "[" + colorToHex(theme.Registers.RegisterValue_Hexadecimal) + "]0x00007fff[-] "
			preview += "[" + colorToHex(theme.Registers.RegisterValue_Decimal) + "]32767[-] "
			preview += "[" + colorToHex(theme.Registers.RegisterValue_Binary) + "]0111111111111111[-]\n"
			preview += "[" + colorToHex(theme.Registers.RegisterName) + "]rbx[-]: "
			preview += "[" + colorToHex(theme.Registers.RegisterValue_Hexadecimal) + "]0xdeadbeef[-] "
			preview += "[" + colorToHex(theme.Registers.RegisterValue_Decimal) + "]3735928559[-] "
			preview += "[" + colorToHex(theme.Registers.RegisterValue_Binary) + "]11010110101011011011111011101111[-]\n\n"
		}

		// tview Theme Fields
		if theme.Theme != nil {
			preview += "[" + colorToHex(theme.Theme.TitleColor) + "]━━━ UI Theme Fields ━━━[-]\n"

			// Show text color variations
			preview += "[" + colorToHex(theme.Theme.PrimaryTextColor) + "]Primary[-] "
			preview += "[" + colorToHex(theme.Theme.SecondaryTextColor) + "]Secondary[-] "
			preview += "[" + colorToHex(theme.Theme.TertiaryTextColor) + "]Tertiary[-] "
			preview += "[" + colorToHex(theme.Theme.ContrastSecondaryTextColor) + "]Contrast[-]\n"

			// Show background variations
			preview += "[" + colorToHex(theme.Theme.PrimitiveBackgroundColor) + "]█ Primitive[-] "
			preview += "[" + colorToHex(theme.Theme.ContrastBackgroundColor) + "]█ Contrast[-] "
			preview += "[" + colorToHex(theme.Theme.MoreContrastBackgroundColor) + "]█ More[-]\n"

			// Show border and graphics
			preview += "[" + colorToHex(theme.Theme.BorderColor) + "]═══╗ Border[-] "
			preview += "[" + colorToHex(theme.Theme.GraphicsColor) + "]◆ Graphics[-] "
			preview += "[" + colorToHex(theme.Theme.InverseTextColor) + "]■ Inverse[-]\n"
		}

		// Update preview text - SetText re-parses color markup with DynamicColors(true)
		ts.preview.SetText(preview)
		ts.preview.ScrollToBeginning()
	}
}

// SetOnThemeSelected sets the callback when a theme is selected (Enter key)
func (ts *ThemeSelector) SetOnThemeSelected(callback ThemeChangeCallback) {
	ts.onThemeSelected = callback
}

// SetOnThemePreview sets the callback when a theme is previewed (arrow keys)
func (ts *ThemeSelector) SetOnThemePreview(callback ThemeChangeCallback) {
	ts.onThemePreview = callback
}

// GetSelectedTheme returns the currently selected theme name
func (ts *ThemeSelector) GetSelectedTheme() string {
	return ts.manager.GetTheme().Name
}

// GetList returns the underlying list widget
func (ts *ThemeSelector) GetList() *tvlib.List {
	return ts.list
}

// SetTheme applies the theme to the ThemeSelector and its children recursively
func (ts *ThemeSelector) SetTheme(theme *themes.Theme) *ThemeSelector {
	ts.Flex.SetBackgroundColor(theme.PrimitiveBackgroundColor)
	ts.Flex.SetBorderColor(theme.BorderColor)
	ts.list.SetBackgroundColor(theme.PrimitiveBackgroundColor)
	ts.list.SetMainTextColor(theme.PrimaryTextColor)
	ts.list.SetBorderColor(theme.BorderColor)
	ts.preview.SetBackgroundColor(theme.PrimitiveBackgroundColor)
	ts.preview.SetTextColor(theme.PrimaryTextColor)
	ts.preview.SetBorderColor(theme.BorderColor)
	return ts
}
