package logging

import (
	"log/slog"
	"sync"
)

// defaultRegistry is the global registry used across the entire application.
var defaultRegistry *Registry
var registryOnce sync.Once

// DefaultRegistry returns the global default registry.
// It's initialized once with a default syslog sink to avoid interfering with TUI output.
func DefaultRegistry() *Registry {
	registryOnce.Do(func() {
		defaultRegistry = NewRegistry()

		// Register a default syslog sink for initial logging (doesn't interfere with TUI)
		syslogSink, err := NewSyslogSink("syslog", "cucaracha", slog.LevelDebug)
		if err != nil {
			// If syslog is not available (e.g., no syslog daemon), fall back to dev/null
			// to prevent TUI disruption
			syslogSink = NewTextSink("fallback", devNull{}, slog.LevelDebug)
		}
		defaultRegistry.RegisterSink(syslogSink)

		// Register root logger
		rootLogger := NewRegisteredLogger("cucaracha", syslogSink)
		defaultRegistry.RegisterLogger(rootLogger)
	})

	return defaultRegistry
}

// Get returns a logger for the given hierarchical name from the default registry.
func Get(name string) *Logger {
	return DefaultRegistry().Get(name)
}

// GetRegistered returns the registered logger for the given name from the default registry.
func GetRegistered(name string) *RegisteredLogger {
	return DefaultRegistry().ResolveLogger(name)
}

// InitializeWithConfig initializes the default registry with YAML configuration.
func InitializeWithConfig(configPath string) error {
	config, err := NewConfigFromYAML(configPath)
	if err != nil {
		return err
	}

	return config.Apply(DefaultRegistry())
}

// InitializeWithConfigString initializes the default registry with YAML string.
func InitializeWithConfigString(configStr string) error {
	config, err := NewConfigFromString(configStr)
	if err != nil {
		return err
	}

	return config.Apply(DefaultRegistry())
}

// devNull is a no-op writer that discards all writes.
// Used as fallback when syslog is unavailable to prevent TUI disruption.
type devNull struct{}

func (d devNull) Write(p []byte) (int, error) {
	return len(p), nil
}
