package themes

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThemesGlobalMap(t *testing.T) {
	require.NotNil(t, Themes)
	assert.Greater(t, len(Themes), 0)

	expectedThemes := []string{"monokai-dark", "dracula", "nord", "solarized-dark", "gruvbox-dark", "one-dark"}
	for _, name := range expectedThemes {
		assert.Contains(t, Themes, name, "expected theme %q not found in Themes map", name)
	}
}

func TestMonokaiDarkTheme(t *testing.T) {
	theme := MonokaiDark

	require.NotNil(t, theme)
	assert.Equal(t, "Monokai Dark", theme.Name)

	assert.NotNil(t, theme.UserIO)
	assert.NotNil(t, theme.SourceSnippet)
	assert.NotNil(t, theme.MemoryDump)
	assert.NotNil(t, theme.Disassembly)
	assert.NotNil(t, theme.Registers)
	assert.NotNil(t, theme.Theme)

	verifyThemeColors(t, theme, "monokai-dark")
}

func TestDraculaTheme(t *testing.T) {
	theme := Dracula

	require.NotNil(t, theme)
	assert.Equal(t, "Dracula", theme.Name)
	assert.NotNil(t, theme.UserIO)

	verifyThemeColors(t, theme, "dracula")
}

func TestNordTheme(t *testing.T) {
	theme := Nord

	require.NotNil(t, theme)
	assert.Equal(t, "Nord", theme.Name)
	assert.NotNil(t, theme.UserIO)

	verifyThemeColors(t, theme, "nord")
}

func TestSolarizedDarkTheme(t *testing.T) {
	theme := SolarizedDark

	require.NotNil(t, theme)
	assert.Equal(t, "Solarized Dark", theme.Name)

	verifyThemeColors(t, theme, "solarized-dark")
}

func TestGruvboxDarkTheme(t *testing.T) {
	theme := GruvboxDark

	require.NotNil(t, theme)
	assert.Equal(t, "Gruvbox Dark", theme.Name)

	verifyThemeColors(t, theme, "gruvbox-dark")
}

func TestOneDarkTheme(t *testing.T) {
	theme := OneDark

	require.NotNil(t, theme)
	assert.Equal(t, "One Dark", theme.Name)

	verifyThemeColors(t, theme, "one-dark")
}

func verifyThemeColors(t *testing.T, theme *Theme, themeName string) {
	assert.NotEqual(t, tcell.ColorDefault, theme.UserIO.CommandPrompt, "%s: UserIO.CommandPrompt not set", themeName)
	assert.NotEqual(t, tcell.ColorDefault, theme.UserIO.Info, "%s: UserIO.Info not set", themeName)
	assert.NotEqual(t, tcell.ColorDefault, theme.UserIO.Error, "%s: UserIO.Error not set", themeName)
	assert.NotEqual(t, tcell.ColorDefault, theme.UserIO.Success, "%s: UserIO.Success not set", themeName)
	assert.NotEqual(t, tcell.ColorDefault, theme.UserIO.Warning, "%s: UserIO.Warning not set", themeName)

	assert.NotEqual(t, tcell.ColorDefault, theme.SourceSnippet.LineNumber, "%s: SourceSnippet.LineNumber not set", themeName)
	assert.NotEqual(t, tcell.ColorDefault, theme.SourceSnippet.CurrentLine, "%s: SourceSnippet.CurrentLine not set", themeName)
	assert.NotEqual(t, tcell.ColorDefault, theme.SourceSnippet.BreakpointMarker, "%s: SourceSnippet.BreakpointMarker not set", themeName)
	assert.NotNil(t, theme.SourceSnippet.C, "%s: SourceSnippet.C not set", themeName)
	if theme.SourceSnippet.C != nil {
		verifySyntaxTheme(t, theme.SourceSnippet.C, themeName)
	}

	assert.NotEqual(t, tcell.ColorDefault, theme.MemoryDump.Address, "%s: MemoryDump.Address not set", themeName)
	assert.NotEqual(t, tcell.ColorDefault, theme.MemoryDump.HexDump, "%s: MemoryDump.HexDump not set", themeName)
	assert.NotEqual(t, tcell.ColorDefault, theme.MemoryDump.ASCII, "%s: MemoryDump.ASCII not set", themeName)

	assert.NotEqual(t, tcell.ColorDefault, theme.Disassembly.Address, "%s: Disassembly.Address not set", themeName)
	assert.NotEqual(t, tcell.ColorDefault, theme.Disassembly.Mnemonic, "%s: Disassembly.Mnemonic not set", themeName)

	assert.NotEqual(t, tcell.ColorDefault, theme.Registers.RegisterName, "%s: Registers.RegisterName not set", themeName)

	assert.NotEqual(t, tcell.ColorDefault, theme.Theme.PrimitiveBackgroundColor, "%s: tview.Theme.PrimitiveBackgroundColor not set", themeName)
	assert.NotEqual(t, tcell.ColorDefault, theme.Theme.BorderColor, "%s: tview.Theme.BorderColor not set", themeName)
	assert.NotEqual(t, tcell.ColorDefault, theme.Theme.PrimaryTextColor, "%s: tview.Theme.PrimaryTextColor not set", themeName)
}

func verifySyntaxTheme(t *testing.T, cTheme *CSyntaxTheme, themeName string) {
	assert.NotEqual(t, tcell.ColorDefault, cTheme.Keyword, "%s: CSyntaxTheme.Keyword not set", themeName)
	assert.NotEqual(t, tcell.ColorDefault, cTheme.Type, "%s: CSyntaxTheme.Type not set", themeName)
	assert.NotEqual(t, tcell.ColorDefault, cTheme.Function, "%s: CSyntaxTheme.Function not set", themeName)
	assert.NotEqual(t, tcell.ColorDefault, cTheme.String, "%s: CSyntaxTheme.String not set", themeName)
	assert.NotEqual(t, tcell.ColorDefault, cTheme.Comment, "%s: CSyntaxTheme.Comment not set", themeName)
}

func TestThemeConsistency(t *testing.T) {
	allThemes := []*Theme{MonokaiDark, Dracula, Nord, SolarizedDark, GruvboxDark, OneDark}

	for _, theme := range allThemes {
		assert.NotNil(t, theme.UserIO, "%s: missing UserIO", theme.Name)
		assert.NotNil(t, theme.SourceSnippet, "%s: missing SourceSnippet", theme.Name)
		assert.NotNil(t, theme.MemoryDump, "%s: missing MemoryDump", theme.Name)
		assert.NotNil(t, theme.Disassembly, "%s: missing Disassembly", theme.Name)
		assert.NotNil(t, theme.Registers, "%s: missing Registers", theme.Name)
		assert.NotNil(t, theme.Theme, "%s: missing tview.Theme", theme.Name)
		assert.NotNil(t, theme.SourceSnippet.C, "%s: missing C syntax theme", theme.Name)
	}
}

func TestThemeMapConsistency(t *testing.T) {
	assert.Equal(t, Themes["monokai-dark"], MonokaiDark, "Themes[monokai-dark] does not match MonokaiDark")
	assert.Equal(t, Themes["dracula"], Dracula, "Themes[dracula] does not match Dracula")
	assert.Equal(t, Themes["nord"], Nord, "Themes[nord] does not match Nord")
	assert.Equal(t, Themes["solarized-dark"], SolarizedDark, "Themes[solarized-dark] does not match SolarizedDark")
	assert.Equal(t, Themes["gruvbox-dark"], GruvboxDark, "Themes[gruvbox-dark] does not match GruvboxDark")
	assert.Equal(t, Themes["one-dark"], OneDark, "Themes[one-dark] does not match OneDark")
}
