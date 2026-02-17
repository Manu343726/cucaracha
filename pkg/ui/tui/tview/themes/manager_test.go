package themes

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()

	require.NotNil(t, manager)
	assert.NotNil(t, manager.current)
	assert.Greater(t, len(manager.themes), 0)
	assert.Equal(t, "Monokai Dark", manager.current.Name)
}

func TestManagerGetTheme(t *testing.T) {
	manager := NewManager()

	theme := manager.GetTheme()
	assert.NotNil(t, theme)
	assert.NotEmpty(t, theme.Name)
}

func TestManagerSetTheme(t *testing.T) {
	manager := NewManager()

	err := manager.SetTheme("dracula")
	assert.NoError(t, err)
	assert.Equal(t, "Dracula", manager.current.Name)

	err = manager.SetTheme("nonexistent-theme")
	assert.Error(t, err)
}

func TestManagerGetThemeNames(t *testing.T) {
	manager := NewManager()

	names := manager.GetThemeNames()
	assert.Greater(t, len(names), 0)

	// Verify themes are sorted
	if len(names) > 1 {
		for i := 0; i < len(names)-1; i++ {
			assert.LessOrEqual(t, names[i], names[i+1], "themes should be sorted")
		}
	}

	expectedThemes := []string{"dracula", "gruvbox-dark", "monokai-dark", "nord", "one-dark", "solarized-dark"}
	for _, expected := range expectedThemes {
		found := false
		for _, name := range names {
			if name == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "expected theme %q not found in theme names", expected)
	}
}

func TestManagerGetThemeByIndex(t *testing.T) {
	manager := NewManager()

	theme, err := manager.GetThemeByIndex(0)
	assert.NoError(t, err)
	assert.NotNil(t, theme)

	_, err = manager.GetThemeByIndex(999)
	assert.Error(t, err)
}

func TestManagerSelectThemeByIndex(t *testing.T) {
	manager := NewManager()

	targetIndex := 1

	err := manager.SelectThemeByIndex(targetIndex)
	assert.NoError(t, err)

	theme, _ := manager.GetThemeByIndex(targetIndex)
	assert.Equal(t, theme.Name, manager.current.Name)

	err = manager.SelectThemeByIndex(999)
	assert.Error(t, err)
}

func TestManagerGetThemeByName(t *testing.T) {
	manager := NewManager()

	theme, err := manager.GetThemeByName("dracula")
	assert.NoError(t, err)
	assert.NotNil(t, theme)
	assert.Equal(t, "Dracula", theme.Name)

	_, err = manager.GetThemeByName("nonexistent")
	assert.Error(t, err)
}

func TestManagerGetCurrentThemeName(t *testing.T) {
	manager := NewManager()

	name := manager.GetCurrentThemeName()
	assert.Equal(t, "monokai-dark", name)

	manager.SetTheme("nord")
	name = manager.GetCurrentThemeName()
	assert.Equal(t, "nord", name)
}

func TestManagerGetThemes(t *testing.T) {
	manager := NewManager()

	themes := manager.GetThemes()
	assert.Greater(t, len(themes), 0)
	assert.Equal(t, 6, len(themes))

	expectedThemes := []string{"dracula", "gruvbox-dark", "monokai-dark", "nord", "one-dark", "solarized-dark"}
	for _, name := range expectedThemes {
		assert.Contains(t, themes, name, "expected theme %q not found in themes map", name)
	}
}

func TestManagerConsistency(t *testing.T) {
	manager := NewManager()

	current := manager.GetTheme()
	assert.Equal(t, manager.current, current)

	byName, err := manager.GetThemeByName(manager.GetCurrentThemeName())
	assert.NoError(t, err)
	assert.Equal(t, current.Name, byName.Name)
}

func TestManagerThemeProperties(t *testing.T) {
	manager := NewManager()

	theme := manager.GetTheme()

	assert.NotNil(t, theme.UserIO)
	assert.NotNil(t, theme.SourceSnippet)
	assert.NotNil(t, theme.MemoryDump)
	assert.NotNil(t, theme.Disassembly)
	assert.NotNil(t, theme.Registers)
	assert.NotNil(t, theme.Theme)
}

func TestManagerThemeColors(t *testing.T) {
	manager := NewManager()

	theme := manager.GetTheme()

	assert.NotEqual(t, tcell.ColorDefault, theme.UserIO.CommandPrompt)
	assert.NotEqual(t, tcell.ColorDefault, theme.SourceSnippet.LineNumber)
	assert.NotEqual(t, tcell.ColorDefault, theme.MemoryDump.Address)
	assert.NotEqual(t, tcell.ColorDefault, theme.Disassembly.Mnemonic)
	assert.NotEqual(t, tcell.ColorDefault, theme.Registers.RegisterName)
}
