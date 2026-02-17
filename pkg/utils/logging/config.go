package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// SinkType represents the type of logging sink.
type SinkType string

const (
	SinkTypeFile   SinkType = "file"
	SinkTypeStdout SinkType = "stdout"
	SinkTypeStderr SinkType = "stderr"
	SinkTypeSyslog SinkType = "syslog"
)

// String implements fmt.Stringer for SinkType.
func (st SinkType) String() string {
	return string(st)
}

// Valid checks if the sink type is valid.
func (st SinkType) Valid() bool {
	switch st {
	case SinkTypeFile, SinkTypeStdout, SinkTypeStderr, SinkTypeSyslog:
		return true
	}
	return false
}

// LogFormat represents the format of log output.
type LogFormat string

const (
	LogFormatJSON LogFormat = "json"
	LogFormatText LogFormat = "text"
)

// String implements fmt.Stringer for LogFormat.
func (lf LogFormat) String() string {
	return string(lf)
}

// Valid checks if the log format is valid.
func (lf LogFormat) Valid() bool {
	switch lf {
	case LogFormatJSON, LogFormatText:
		return true
	}
	return false
}

// LogLevel represents the log level.
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelPanic LogLevel = "panic"
)

// String implements fmt.Stringer for LogLevel.
func (ll LogLevel) String() string {
	return string(ll)
}

// Valid checks if the log level is valid.
func (ll LogLevel) Valid() bool {
	switch ll {
	case LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError, LogLevelPanic:
		return true
	}
	return false
}

// ToSlogLevel converts LogLevel to slog.Level.
func (ll LogLevel) ToSlogLevel() slog.Level {
	switch ll {
	case LogLevelDebug:
		return slog.LevelDebug
	case LogLevelInfo:
		return slog.LevelInfo
	case LogLevelWarn:
		return slog.LevelWarn
	case LogLevelError:
		return slog.LevelError
	case LogLevelPanic:
		// Panic level is treated as Error level for slog purposes
		// The Panic() method handles the actual panic
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Config represents the YAML configuration structure for loggers and sinks.
type Config struct {
	Sinks   []SinkConfig   `yaml:"sinks"`
	Loggers []LoggerConfig `yaml:"loggers"`
}

// SinkConfig represents a single sink configuration.
type SinkConfig struct {
	Name    string    `yaml:"name"`
	Type    SinkType  `yaml:"type"`               // file, stdout, stderr, syslog
	Path    string    `yaml:"path,omitempty"`     // for file sinks and syslog tag
	Level   LogLevel  `yaml:"level,omitempty"`    // debug, info, warn, error
	Format  LogFormat `yaml:"format,omitempty"`   // json or text (default: json)
	MaxSize int       `yaml:"max_size,omitempty"` // MB, for rotating files
}

// LoggerConfig represents a single logger configuration.
type LoggerConfig struct {
	Name  string   `yaml:"name"`
	Sinks []string `yaml:"sinks"` // list of sink names to attach
}

// NewConfigFromYAML loads configuration from a YAML file.
func NewConfigFromYAML(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &config, nil
}

// NewConfigFromString parses configuration from a YAML string.
func NewConfigFromString(yamlStr string) (*Config, error) {
	var config Config
	if err := yaml.Unmarshal([]byte(yamlStr), &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	return &config, nil
}

// Apply applies the configuration to a registry.
// It creates and registers all sinks first, then creates and registers registered loggers
// with their sinks connected via Fanout handlers.
func (c *Config) Apply(registry *Registry) error {
	// Create and register sinks first
	sinkMap := make(map[string]*Sink)
	for _, sinkCfg := range c.Sinks {
		sink, err := createSink(sinkCfg)
		if err != nil {
			return fmt.Errorf("failed to create sink %q: %w", sinkCfg.Name, err)
		}

		if err := registry.RegisterSink(sink); err != nil {
			return err
		}
		sinkMap[sinkCfg.Name] = sink
	}

	// Create and register registered loggers with sinks already attached
	for _, loggerCfg := range c.Loggers {
		// Collect sinks for this logger
		sinks := make([]*Sink, 0, len(loggerCfg.Sinks))
		for _, sinkName := range loggerCfg.Sinks {
			sink, exists := sinkMap[sinkName]
			if !exists {
				return fmt.Errorf("sink %q referenced by logger %q not found", sinkName, loggerCfg.Name)
			}
			sinks = append(sinks, sink)
		}

		// Create registered logger with Fanout handler connecting all sinks
		registeredLogger := NewRegisteredLogger(loggerCfg.Name, sinks...)

		if regErr := registry.RegisterLogger(registeredLogger); regErr != nil {
			return regErr
		}
	}

	return nil
}

// createSink creates a sink from configuration.
func createSink(cfg SinkConfig) (*Sink, error) {
	// Validate sink type
	if !cfg.Type.Valid() {
		return nil, fmt.Errorf("invalid sink type: %s", cfg.Type)
	}

	// Validate and convert log level
	level := LogLevelInfo
	if cfg.Level != "" {
		if !cfg.Level.Valid() {
			return nil, fmt.Errorf("invalid log level: %s", cfg.Level)
		}
		level = cfg.Level
	}

	// Validate and normalize log format
	format := LogFormatJSON
	if cfg.Format != "" {
		if !cfg.Format.Valid() {
			return nil, fmt.Errorf("invalid log format: %s", cfg.Format)
		}
		format = cfg.Format
	}

	var writer io.Writer

	switch cfg.Type {
	case SinkTypeFile:
		if cfg.Path == "" {
			return nil, fmt.Errorf("file sink requires 'path'")
		}

		// Create directory if it doesn't exist
		dir := filepath.Dir(cfg.Path)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create directory: %w", err)
			}
		}

		file, err := os.OpenFile(cfg.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		writer = file

	case SinkTypeStdout:
		writer = os.Stdout
	case SinkTypeStderr:
		writer = os.Stderr

	case SinkTypeSyslog:
		// Syslog sink requires the tag parameter
		tag := cfg.Path
		if tag == "" {
			tag = "cucaracha"
		}
		syslogSink, err := NewSyslogSink(cfg.Name, tag, level.ToSlogLevel())
		if err != nil {
			return nil, fmt.Errorf("failed to create syslog sink: %w", err)
		}
		return syslogSink, nil

	default:
		return nil, fmt.Errorf("unknown sink type: %s", cfg.Type)
	}

	if format == LogFormatText {
		return NewTextSink(cfg.Name, writer, level.ToSlogLevel()), nil
	}
	return NewFileSink(cfg.Name, writer, level.ToSlogLevel()), nil
}

// Example returns a YAML example configuration.
func ExampleConfig() string {
	return `# Cucaracha Logging Configuration

sinks:
  # syslog sink (doesn't interfere with TUI)
  - name: syslog
    type: syslog
    level: debug

  # stdout with text formatting
  - name: console
    type: stdout
    level: info
    format: text

  # File sink with JSON formatting
  - name: file_json
    type: file
    path: /tmp/cucaracha.log
    level: debug
    format: json

  # File sink for errors only
  - name: file_errors
    type: file
    path: /tmp/cucaracha_errors.log
    level: error
    format: text

loggers:
  # Root logger for all runtime operations, uses syslog by default
  - name: cucaracha
    sinks:
      - syslog

  # Runtime logger
  - name: runtime
    sinks:
      - syslog
      - file_json

  # Sub-logger for CPU operations
  - name: runtime.cpu
    sinks:
      - syslog
      - file_json

  # Sub-logger for memory operations
  - name: runtime.memory
    sinks:
      - syslog

  # Debugger logs to syslog and error file
  - name: debugger
    sinks:
      - syslog
      - file_errors

  # LLVM compilation logs
  - name: llvm
    sinks:
      - syslog
      - file_json

  # UI logs to syslog only (not console to avoid TUI disruption)
  - name: ui
    sinks:
      - syslog
`
}
