package themes

import (
	"fmt"
	"sort"
)

// Manages the current theme
type Manager struct {
	current          *Theme
	themes           map[string]*Theme
	themeNamesSorted []string
}

// NewManager creates a new theme manager with monokai-dark as default
func NewManager() *Manager {
	return &Manager{
		current: Themes["monokai-dark"],
		themes:  Themes,
		themeNamesSorted: func() []string {
			names := make([]string, 0, len(Themes))
			for name := range Themes {
				names = append(names, name)
			}
			sort.Strings(names)
			return names
		}(),
	}
}

// GetTheme returns the current theme
func (tm *Manager) GetTheme() *Theme {
	return tm.current
}

// SetTheme sets the current theme by name
func (tm *Manager) SetTheme(name string) error {
	if theme, exists := tm.themes[name]; exists {
		tm.current = theme
		return nil
	}
	return fmt.Errorf("theme %q not found", name)
}

// GetThemeNames returns all available theme names
func (tm *Manager) GetThemeNames() []string {
	return tm.themeNamesSorted
}

// Returns a theme given its index inside the sorted theme names list
func (tm *Manager) GetThemeByIndex(index int) (*Theme, error) {
	if index >= 0 && index < len(tm.themeNamesSorted) {
		name := tm.themeNamesSorted[index]
		return tm.themes[name], nil
	}
	return nil, fmt.Errorf("index %d out of range", index)
}

// Selects the current theme by its index inside the sorted theme names list
func (tm *Manager) SelectThemeByIndex(index int) error {
	if index >= 0 && index < len(tm.themeNamesSorted) {
		name := tm.themeNamesSorted[index]
		tm.current = tm.themes[name]
		return nil
	}
	return fmt.Errorf("index %d out of range", index)
}

// GetThemeByName returns a theme by name
func (tm *Manager) GetThemeByName(name string) (*Theme, error) {
	if theme, exists := tm.themes[name]; exists {
		return theme, nil
	}
	return nil, fmt.Errorf("theme %q not found", name)
}

// GetCurrentThemeName returns the name of the current theme
func (tm *Manager) GetCurrentThemeName() string {
	for name, theme := range tm.themes {
		if theme == tm.current {
			return name
		}
	}
	return "monokai-dark"
}

// GetThemes returns all available themes
func (tm *Manager) GetThemes() map[string]*Theme {
	return tm.themes
}
