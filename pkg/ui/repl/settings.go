package repl

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Setting key constants for REPL configuration
const (
	// SettingKeyDisplayEvents controls whether incoming debugger events are displayed
	SettingKeyDisplayEvents = "display.events"
	// SettingKeyShowLogs specifies logger names to capture logs from
	SettingKeyShowLogs = "show.logs"
)

// Setting represents a single REPL setting
type Setting struct {
	Name         string      // Setting name (e.g. "display.events")
	Description  string      // Description of what the setting does
	Value        interface{} // Current value (bool, string, int, etc)
	DefaultValue interface{} // Default value
}

// Settings holds all available REPL settings
type Settings struct {
	settings map[string]*Setting
}

// NewSettings creates a new settings collection with defaults
func NewSettings() *Settings {
	s := &Settings{
		settings: make(map[string]*Setting),
	}

	// Register default settings
	s.Register(SettingKeyDisplayEvents, "Display incoming debugger events while the program is running", true)
	s.Register(SettingKeyShowLogs, "Logger names to capture logs from (space-separated)", []string{})

	return s
}

// Register adds a new setting or updates an existing one
func (s *Settings) Register(name, description string, defaultValue interface{}) error {
	if _, exists := s.settings[name]; exists {
		return fmt.Errorf("setting %q already registered", name)
	}

	s.settings[name] = &Setting{
		Name:         name,
		Description:  description,
		Value:        defaultValue,
		DefaultValue: defaultValue,
	}

	return nil
}

// Set sets a setting value
func (s *Settings) Set(name string, value interface{}) error {
	setting, exists := s.settings[name]
	if !exists {
		return fmt.Errorf("unknown setting %q", name)
	}

	// Type conversion based on the default value type
	switch setting.DefaultValue.(type) {
	case bool:
		// Try to parse as boolean
		switch v := value.(type) {
		case bool:
			setting.Value = v
		case string:
			boolVal, err := parseBool(v)
			if err != nil {
				return fmt.Errorf("invalid value for %s: %v (expected true/false)", name, v)
			}
			setting.Value = boolVal
		default:
			return fmt.Errorf("invalid value type for %s: %T", name, value)
		}

	case string:
		// Accept any type, convert to string
		setting.Value = fmt.Sprintf("%v", value)

	case int:
		// Try to parse as integer
		switch v := value.(type) {
		case int:
			setting.Value = v
		case string:
			intVal, err := strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("invalid value for %s: %v (expected integer)", name, v)
			}
			setting.Value = intVal
		default:
			return fmt.Errorf("invalid value type for %s: %T", name, value)
		}

	case []string:
		// Handle string slices (like logger names for show.logs)
		switch v := value.(type) {
		case []string:
			setting.Value = v
		case []interface{}:
			strs := make([]string, len(v))
			for i, item := range v {
				strs[i] = fmt.Sprintf("%v", item)
			}
			setting.Value = strs
		default:
			return fmt.Errorf("invalid value type for %s: %T", name, value)
		}

	default:
		setting.Value = value
	}

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
	for i := 0; i < len(settings); i++ {
		for j := i + 1; j < len(settings); j++ {
			if settings[i].Name > settings[j].Name {
				settings[i], settings[j] = settings[j], settings[i]
			}
		}
	}

	return settings
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
