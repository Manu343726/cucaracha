package repl

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Setting key constants for REPL configuration
const (
	// SettingKeyDisplayEvents controls whether incoming debugger events are displayed
	SettingKeyDisplayEvents = "display.events"
	// SettingKeyDisplayLogs specifies logger names to capture logs from
	SettingKeyDisplayLogs = "display.logs"
	// SettingKeyBuildAutoClang controls whether to automatically build clang when loading programs
	SettingKeyBuildAutoClang = "build.auto_clang"
	// SettingKeyBuildForceClang forces a rebuild of clang even if it's already built
	SettingKeyBuildForceClang = "build.force_clang"
)

// SettingChangeHandler is a callback function that is invoked when a setting value changes.
// It receives the setting name and the new value.
type SettingChangeHandler func(name string, newValue interface{}) error

// Setting represents a single REPL setting
type Setting struct {
	Name         string      // Setting name (e.g. "display.events")
	Description  string      // Description of what the setting does
	Value        interface{} // Current value (bool, string, int, etc)
	DefaultValue interface{} // Default value
}

// SettingCategory represents a hierarchical category of settings
type SettingCategory struct {
	Name          string                      // Category name (e.g. "display")
	Description   string                      // Category description
	Subcategories map[string]*SettingCategory // Nested subcategories
	Settings      map[string]*Setting         // Settings in this category
}

// Settings holds all available REPL settings
type Settings struct {
	settings     map[string]*Setting             // Flat map for fast lookup by full key name
	categoryRoot *SettingCategory                // Root of the category tree
	callbacks    map[string]SettingChangeHandler // Callbacks to invoke when settings change
}

// NewSettings creates a new settings collection with defaults
func NewSettings() *Settings {
	s := &Settings{
		settings: make(map[string]*Setting),
		categoryRoot: &SettingCategory{
			Name:          "Root",
			Subcategories: make(map[string]*SettingCategory),
			Settings:      make(map[string]*Setting),
		},
		callbacks: make(map[string]SettingChangeHandler),
	}

	// Register category descriptions
	s.RegisterCategoryDescription("display", "Display and output settings")
	s.RegisterCategoryDescription("build", "Build and compilation settings")

	// Register default settings
	s.Register(SettingKeyDisplayEvents, "Display incoming debugger events during execution", true)
	s.Register(SettingKeyDisplayLogs, "Logger names to capture logs from (space-separated)", []string{})
	s.Register(SettingKeyBuildAutoClang, "Automatically build clang when loading C/C++ programs", true)
	s.Register(SettingKeyBuildForceClang, "Force rebuild of clang even if already built", false)

	return s
}

// getCategoryPath navigates the category tree, creating subcategories as needed
// Returns the parent category and the setting/subcategory name
func (s *Settings) getCategoryPath(fullKey string) (*SettingCategory, string, error) {
	parts := strings.Split(fullKey, ".")
	if len(parts) == 0 {
		return nil, "", fmt.Errorf("invalid setting name: empty")
	}

	// Navigate to the parent category
	currentCategory := s.categoryRoot
	for _, part := range parts[:len(parts)-1] {
		if subcat, exists := currentCategory.Subcategories[part]; exists {
			currentCategory = subcat
		} else {
			// Create missing subcategory
			newCat := &SettingCategory{
				Name:          part,
				Subcategories: make(map[string]*SettingCategory),
				Settings:      make(map[string]*Setting),
			}
			currentCategory.Subcategories[part] = newCat
			currentCategory = newCat
		}
	}

	return currentCategory, parts[len(parts)-1], nil
}

// Register adds a new setting or updates an existing one
func (s *Settings) Register(name, description string, defaultValue interface{}) error {
	if _, exists := s.settings[name]; exists {
		return fmt.Errorf("setting %q already registered", name)
	}

	// Get the category path for this setting
	category, settingName, err := s.getCategoryPath(name)
	if err != nil {
		return err
	}

	setting := &Setting{
		Name:         name,
		Description:  description,
		Value:        defaultValue,
		DefaultValue: defaultValue,
	}

	// Add to flat map for fast lookup
	s.settings[name] = setting

	// Add to category tree
	category.Settings[settingName] = setting

	return nil
}

// RegisterCategoryDescription sets a description for a category
func (s *Settings) RegisterCategoryDescription(categoryPath, description string) error {
	parts := strings.Split(categoryPath, ".")
	if len(parts) == 0 {
		return fmt.Errorf("invalid category path: empty")
	}

	currentCategory := s.categoryRoot
	for _, part := range parts {
		if subcat, exists := currentCategory.Subcategories[part]; exists {
			currentCategory = subcat
		} else {
			// Create missing subcategory
			newCat := &SettingCategory{
				Name:          part,
				Subcategories: make(map[string]*SettingCategory),
				Settings:      make(map[string]*Setting),
			}
			currentCategory.Subcategories[part] = newCat
			currentCategory = newCat
		}
	}

	currentCategory.Description = description
	return nil
}

// SetChangeCallback registers a callback to be invoked when a setting changes.
// If a callback is already registered for this setting, it will be replaced.
func (s *Settings) SetChangeCallback(name string, handler SettingChangeHandler) error {
	if _, exists := s.settings[name]; !exists {
		return fmt.Errorf("unknown setting %q", name)
	}

	s.callbacks[name] = handler
	return nil
}

// Set sets a setting value. The setting is only updated if the callback (if registered) succeeds.
func (s *Settings) Set(name string, value interface{}) error {
	setting, exists := s.settings[name]
	if !exists {
		return fmt.Errorf("unknown setting %q", name)
	}

	// Convert and validate the new value without updating the setting yet
	var newValue interface{}

	switch setting.DefaultValue.(type) {
	case bool:
		// Try to parse as boolean
		switch v := value.(type) {
		case bool:
			newValue = v
		case string:
			boolVal, err := parseBool(v)
			if err != nil {
				return fmt.Errorf("invalid value for %s: %v (expected true/false)", name, v)
			}
			newValue = boolVal
		default:
			return fmt.Errorf("invalid value type for %s: %T", name, value)
		}

	case string:
		// Accept any type, convert to string
		newValue = fmt.Sprintf("%v", value)

	case int:
		// Try to parse as integer
		switch v := value.(type) {
		case int:
			newValue = v
		case string:
			intVal, err := strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("invalid value for %s: %v (expected integer)", name, v)
			}
			newValue = intVal
		default:
			return fmt.Errorf("invalid value type for %s: %T", name, value)
		}

	case []string:
		// Handle string slices (like logger names for show.logs)
		switch v := value.(type) {
		case []string:
			newValue = v
		case []interface{}:
			strs := make([]string, len(v))
			for i, item := range v {
				strs[i] = fmt.Sprintf("%v", item)
			}
			newValue = strs
		default:
			return fmt.Errorf("invalid value type for %s: %T", name, value)
		}

	default:
		newValue = value
	}

	// Call the change callback if one is registered for this setting
	if handler, exists := s.callbacks[name]; exists {
		if err := handler(name, newValue); err != nil {
			// Don't update the setting if the callback fails
			return err
		}
	}

	// Only update the setting if validation and callback (if exists) succeeded
	setting.Value = newValue

	return nil
}

// Get retrieves a setting value
func (s *Settings) Get(name string) (interface{}, error) {
	setting, exists := s.settings[name]
	if !exists {
		return nil, fmt.Errorf("unknown setting %q", name)
	}

	return setting.Value, nil
}

// GetBool retrieves a setting as a boolean
func (s *Settings) GetBool(name string) (bool, error) {
	val, err := s.Get(name)
	if err != nil {
		return false, err
	}

	if b, ok := val.(bool); ok {
		return b, nil
	}

	return false, fmt.Errorf("setting %q is not a boolean", name)
}

// GetString retrieves a setting as a string
func (s *Settings) GetString(name string) (string, error) {
	val, err := s.Get(name)
	if err != nil {
		return "", err
	}

	if s, ok := val.(string); ok {
		return s, nil
	}

	return "", fmt.Errorf("setting %q is not a string", name)
}

// GetInt retrieves a setting as an integer
func (s *Settings) GetInt(name string) (int, error) {
	val, err := s.Get(name)
	if err != nil {
		return 0, err
	}

	if i, ok := val.(int); ok {
		return i, nil
	}

	return 0, fmt.Errorf("setting %q is not an integer", name)
}

// GetStringSlice retrieves a setting as a string slice
func (s *Settings) GetStringSlice(name string) ([]string, error) {
	val, err := s.Get(name)
	if err != nil {
		return nil, err
	}

	if ss, ok := val.([]string); ok {
		return ss, nil
	}

	return nil, fmt.Errorf("setting %q is not a string slice", name)
}

// List returns all settings with their current values
func (s *Settings) List() []*Setting {
	settings := make([]*Setting, 0, len(s.settings))
	for _, setting := range s.settings {
		settings = append(settings, setting)
	}

	// Sort by name for consistent output
	sort.Slice(settings, func(i, j int) bool {
		return settings[i].Name < settings[j].Name
	})

	return settings
}

// ListByCategory returns settings organized by category hierarchy
func (s *Settings) ListByCategory() *SettingCategory {
	// Build a clean category tree by deep-copying non-empty categories
	return s.cloneCategoryTree(s.categoryRoot)
}

// cloneCategoryTree creates a deep copy of the category tree, removing empty branches
func (s *Settings) cloneCategoryTree(cat *SettingCategory) *SettingCategory {
	newCat := &SettingCategory{
		Name:          cat.Name,
		Description:   cat.Description,
		Subcategories: make(map[string]*SettingCategory),
		Settings:      make(map[string]*Setting),
	}

	// Copy non-empty subcategories
	for name, subcat := range cat.Subcategories {
		cloned := s.cloneCategoryTree(subcat)
		// Only include if the subcategory has settings or non-empty subcategories
		if len(cloned.Settings) > 0 || len(cloned.Subcategories) > 0 {
			newCat.Subcategories[name] = cloned
		}
	}

	// Copy settings
	for name, setting := range cat.Settings {
		newCat.Settings[name] = setting
	}

	return newCat
}

// CategoryPath represents the path through the category hierarchy
type CategoryPath struct {
	Path     []string // e.g., ["display"]
	Category *SettingCategory
	Indent   string
}

// IterateCategories yields all categories in the tree in hierarchical order
func (s *Settings) IterateCategories(callback func(path []string, category *SettingCategory, indent string)) {
	s.iterateCategoriesRec(s.categoryRoot, []string{}, "", callback)
}

func (s *Settings) iterateCategoriesRec(cat *SettingCategory, path []string, indent string, callback func([]string, *SettingCategory, string)) {
	// Process subcategories in sorted order
	subNames := make([]string, 0, len(cat.Subcategories))
	for name := range cat.Subcategories {
		subNames = append(subNames, name)
	}
	sort.Strings(subNames)

	for _, name := range subNames {
		subcat := cat.Subcategories[name]
		newPath := append(path, name)
		newIndent := indent + "  "
		callback(newPath, subcat, newIndent)
		s.iterateCategoriesRec(subcat, newPath, newIndent, callback)
	}
}

// Reset resets all settings to their default values
func (s *Settings) Reset() {
	for _, setting := range s.settings {
		setting.Value = setting.DefaultValue
	}
}

// parseBool parses a boolean value from string
func parseBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "yes", "on", "1":
		return true, nil
	case "false", "no", "off", "0":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value: %q", s)
	}
}

// LoadFromFile loads settings from a YAML file
func (s *Settings) LoadFromFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read settings file: %w", err)
	}

	// Parse YAML into a map
	var settingsMap map[string]interface{}
	if err := yaml.Unmarshal(data, &settingsMap); err != nil {
		return fmt.Errorf("failed to parse settings file as YAML: %w", err)
	}

	// Apply each setting
	for key, value := range settingsMap {
		if err := s.Set(key, value); err != nil {
			return fmt.Errorf("failed to set %q from file: %w", key, err)
		}
	}

	return nil
}

// ApplyKeyValue applies a single key=value setting string
func (s *Settings) ApplyKeyValue(kvStr string) error {
	parts := strings.SplitN(kvStr, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid setting format %q (expected key=value)", kvStr)
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	// Check if this setting expects a string slice and parse accordingly
	setting, exists := s.settings[key]
	if exists {
		if _, isList := setting.DefaultValue.([]string); isList {
			// Split by comma or space for string slice values
			// Support both comma-separated and space-separated values
			var values []string
			if strings.Contains(value, ",") {
				values = strings.Split(value, ",")
			} else {
				values = strings.Fields(value)
			}
			// Trim spaces from each value
			for i := range values {
				values[i] = strings.TrimSpace(values[i])
			}
			return s.Set(key, values)
		}
	}

	return s.Set(key, value)
}
