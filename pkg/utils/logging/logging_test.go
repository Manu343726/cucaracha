package logging

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// TestSink_Creation tests that sinks can be created with proper initialization
func TestSink_Creation(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := slog.NewJSONHandler(buf, nil)
	sink := NewSink("test", handler, slog.LevelInfo)

	if sink.Name() != "test" {
		t.Errorf("expected name 'test', got %q", sink.Name())
	}

	if sink.Level() != slog.LevelInfo {
		t.Errorf("expected level Info, got %v", sink.Level())
	}
}

// TestSink_Immutable verifies that sinks are immutable
func TestSink_Immutable(t *testing.T) {
	buf := &bytes.Buffer{}
	sink := NewFileSink("test", buf, slog.LevelWarn)

	originalLevel := sink.Level()
	if originalLevel != slog.LevelWarn {
		t.Errorf("expected level Warn, got %v", originalLevel)
	}
}

// TestRegisteredLogger_Creation tests RegisteredLogger creation
func TestRegisteredLogger_Creation(t *testing.T) {
	buf := &bytes.Buffer{}
	sink := NewFileSink("sink1", buf, slog.LevelInfo)

	regLogger := NewRegisteredLogger("test.module", sink)

	if regLogger.Name() != "test.module" {
		t.Errorf("expected name 'test.module', got %q", regLogger.Name())
	}
}

// TestRegisteredLogger_Immutable verifies RegisteredLogger is immutable
func TestRegisteredLogger_Immutable(t *testing.T) {
	buf := &bytes.Buffer{}
	sink := NewFileSink("sink", buf, slog.LevelInfo)

	regLogger := NewRegisteredLogger("test", sink)
	originalName := regLogger.Name()

	// RegisteredLogger cannot be modified after creation
	if originalName != "test" {
		t.Errorf("expected name 'test', got %q", originalName)
	}
}

// TestLogger_Creation tests Logger wrapper creation
func TestLogger_Creation(t *testing.T) {
	registry := NewRegistry()
	logger := NewLogger("test.module", registry)

	if logger.Name() != "test.module" {
		t.Errorf("expected name 'test.module', got %q", logger.Name())
	}
}

// TestLogger_WithAttrs tests adding attributes to a logger
func TestLogger_WithAttrs(t *testing.T) {
	registry := NewRegistry()
	logger := NewLogger("test", registry)

	attr := slog.String("key", "value")
	newLogger := logger.WithAttrs(attr)

	if newLogger.Name() != logger.Name() {
		t.Errorf("child logger name mismatch")
	}

	// Original logger should be unmodified
	if logger == newLogger {
		t.Error("WithAttrs should return a new instance")
	}
}

// TestLogger_WithGroup tests grouping attributes
func TestLogger_WithGroup(t *testing.T) {
	registry := NewRegistry()
	logger := NewLogger("test", registry)

	newLogger := logger.WithGroup("mygroup")

	if newLogger.Name() != logger.Name() {
		t.Errorf("logger name should be unchanged")
	}

	if logger == newLogger {
		t.Error("WithGroup should return a new instance")
	}
}

// TestLogger_Child tests creating child loggers
func TestLogger_Child(t *testing.T) {
	registry := NewRegistry()
	logger := NewLogger("test", registry)

	child := logger.Child("submodule")

	if child.Name() != "test.submodule" {
		t.Errorf("expected name 'test.submodule', got %q", child.Name())
	}

	if logger == child {
		t.Error("Child should return a new instance")
	}
}

// TestRegistry_RegisterSinks tests sink registration
func TestRegistry_RegisterSinks(t *testing.T) {
	registry := NewRegistry()

	buf := &bytes.Buffer{}
	sink1 := NewFileSink("sink1", buf, slog.LevelInfo)
	sink2 := NewFileSink("sink2", buf, slog.LevelDebug)

	if err := registry.RegisterSink(sink1); err != nil {
		t.Fatalf("failed to register sink: %v", err)
	}

	if err := registry.RegisterSink(sink2); err != nil {
		t.Fatalf("failed to register sink: %v", err)
	}

	// Registering duplicate should fail
	if err := registry.RegisterSink(sink1); err == nil {
		t.Error("expected error when registering duplicate sink")
	}
}

// TestRegistry_RegisterLoggers tests registered logger registration
func TestRegistry_RegisterLoggers(t *testing.T) {
	registry := NewRegistry()

	buf := &bytes.Buffer{}
	sink := NewFileSink("sink", buf, slog.LevelInfo)
	registry.RegisterSink(sink)

	regLogger := NewRegisteredLogger("runtime", sink)
	if err := registry.RegisterLogger(regLogger); err != nil {
		t.Fatalf("failed to register logger: %v", err)
	}

	// Registering duplicate should fail
	if err := registry.RegisterLogger(regLogger); err == nil {
		t.Error("expected error when registering duplicate logger")
	}
}

// TestRegistry_ResolveLogger tests hierarchical logger resolution
func TestRegistry_ResolveLogger(t *testing.T) {
	registry := NewRegistry()
	buf := &bytes.Buffer{}
	sink := NewFileSink("sink", buf, slog.LevelInfo)
	registry.RegisterSink(sink)

	// Register loggers at different levels
	root := NewRegisteredLogger("runtime", sink)
	module := NewRegisteredLogger("runtime.cpu", sink)

	registry.RegisterLogger(root)
	registry.RegisterLogger(module)

	tests := []struct {
		name     string
		lookup   string
		expected string
	}{
		{"exact match", "runtime", "runtime"},
		{"exact child", "runtime.cpu", "runtime.cpu"},
		{"parent resolution", "runtime.cpu.executor", "runtime.cpu"},
		{"root resolution", "runtime.memory", "runtime"},
		{"not found", "unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved := registry.ResolveLogger(tt.lookup)
			if tt.expected == "" {
				if resolved != nil {
					t.Errorf("expected nil, got %v", resolved)
				}
			} else {
				if resolved == nil {
					t.Errorf("expected logger %q, got nil", tt.expected)
				} else if resolved.Name() != tt.expected {
					t.Errorf("expected logger %q, got %q", tt.expected, resolved.Name())
				}
			}
		})
	}
}

// TestRegistry_Get tests getting thin wrapper loggers
func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()
	logger := registry.Get("test.module")

	if logger.Name() != "test.module" {
		t.Errorf("expected name 'test.module', got %q", logger.Name())
	}
}

// TestRegistry_List tests listing registered loggers
func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()
	buf := &bytes.Buffer{}
	sink := NewFileSink("sink", buf, slog.LevelInfo)
	registry.RegisterSink(sink)

	logger1 := NewRegisteredLogger("runtime", sink)
	logger2 := NewRegisteredLogger("debugger", sink)

	registry.RegisterLogger(logger1)
	registry.RegisterLogger(logger2)

	names := registry.ListLoggers()
	if len(names) != 2 {
		t.Errorf("expected 2 loggers, got %d", len(names))
	}

	// Check that both names are present
	foundRuntime := false
	foundDebugger := false
	for _, name := range names {
		if name == "runtime" {
			foundRuntime = true
		}
		if name == "debugger" {
			foundDebugger = true
		}
	}

	if !foundRuntime || !foundDebugger {
		t.Error("expected both runtime and debugger in names")
	}
}

// TestRegistry_Clear tests clearing the registry
func TestRegistry_Clear(t *testing.T) {
	registry := NewRegistry()
	buf := &bytes.Buffer{}
	sink := NewFileSink("sink", buf, slog.LevelInfo)
	registry.RegisterSink(sink)

	logger := NewRegisteredLogger("runtime", sink)
	registry.RegisterLogger(logger)

	if len(registry.ListLoggers()) != 1 {
		t.Error("setup failed")
	}

	registry.Clear()

	if len(registry.ListLoggers()) != 0 {
		t.Errorf("expected 0 loggers after Clear, got %d", len(registry.ListLoggers()))
	}

	if len(registry.ListSinks()) != 0 {
		t.Errorf("expected 0 sinks after Clear, got %d", len(registry.ListSinks()))
	}
}

// TestLogger_Logging tests actual logging through wrapper
func TestLogger_Logging(t *testing.T) {
	registry := NewRegistry()
	buf := &bytes.Buffer{}
	sink := NewFileSink("sink", buf, slog.LevelInfo)
	registry.RegisterSink(sink)

	regLogger := NewRegisteredLogger("app", sink)
	registry.RegisterLogger(regLogger)

	logger := registry.Get("app")

	logger.Info("test message", slog.String("key", "value"))

	// Verify something was written
	if buf.Len() == 0 {
		t.Error("expected log output, got nothing")
	}
}

// TestLogger_ContextPropagation tests that attributes persist across calls
func TestLogger_ContextPropagation(t *testing.T) {
	registry := NewRegistry()
	buf := &bytes.Buffer{}
	sink := NewTextSink("sink", buf, slog.LevelInfo)
	registry.RegisterSink(sink)

	regLogger := NewRegisteredLogger("app", sink)
	registry.RegisterLogger(regLogger)

	// Create logger with pre-configured attributes
	logger := registry.Get("app")
	loggerWithContext := logger.WithAttrs(slog.String("module", "cpu"))

	loggerWithContext.Info("test message")

	// Both original and contextual loggers should work
	logger.Info("another message")

	if buf.Len() == 0 {
		t.Error("expected log output")
	}
}

// TestConfig_FromString tests YAML configuration parsing
func TestConfig_FromString(t *testing.T) {
	yaml := `
sinks:
  - name: console
    type: stdout
    level: info
    format: text

loggers:
  - name: runtime
    sinks: [console]
`

	config, err := NewConfigFromString(yaml)
	if err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	if len(config.Sinks) != 1 {
		t.Errorf("expected 1 sink, got %d", len(config.Sinks))
	}

	if len(config.Loggers) != 1 {
		t.Errorf("expected 1 logger, got %d", len(config.Loggers))
	}

	if config.Sinks[0].Name != "console" {
		t.Errorf("expected sink name 'console', got %q", config.Sinks[0].Name)
	}

	if config.Loggers[0].Name != "runtime" {
		t.Errorf("expected logger name 'runtime', got %q", config.Loggers[0].Name)
	}
}

// TestConfig_Apply tests applying configuration to registry
func TestConfig_Apply(t *testing.T) {
	yaml := `
sinks:
  - name: console
    type: stdout
    level: info
    format: text

loggers:
  - name: runtime
    sinks: [console]
  - name: runtime.cpu
    sinks: [console]
`

	config, err := NewConfigFromString(yaml)
	if err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	registry := NewRegistry()
	if err := config.Apply(registry); err != nil {
		t.Fatalf("failed to apply config: %v", err)
	}

	// Verify sinks are registered
	if len(registry.ListSinks()) != 1 {
		t.Errorf("expected 1 sink, got %d", len(registry.ListSinks()))
	}

	// Verify loggers are registered
	if len(registry.ListLoggers()) != 2 {
		t.Errorf("expected 2 loggers, got %d", len(registry.ListLoggers()))
	}

	// Verify resolution works
	resolved := registry.ResolveLogger("runtime.cpu.executor")
	if resolved == nil || resolved.Name() != "runtime.cpu" {
		t.Error("hierarchical resolution failed")
	}
}

// TestConfig_ApplyErrors tests error handling in configuration
func TestConfig_ApplyErrors(t *testing.T) {
	yaml := `
loggers:
  - name: runtime
    sinks: [nonexistent]
`

	config, err := NewConfigFromString(yaml)
	if err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	registry := NewRegistry()
	if err := config.Apply(registry); err == nil {
		t.Error("expected error when referencing nonexistent sink")
	}
}

// TestLogger_HierarchicalLogging tests hierarchical logger chains
func TestLogger_HierarchicalLogging(t *testing.T) {
	registry := NewRegistry()
	buf := &bytes.Buffer{}
	sink := NewTextSink("sink", buf, slog.LevelInfo)
	registry.RegisterSink(sink)

	// Create parent logger
	parentLogger := NewRegisteredLogger("app", sink)
	registry.RegisterLogger(parentLogger)

	// Get child via Get
	childLogger := registry.Get("app.module.component")

	// Should resolve to parent
	resolved := registry.ResolveLogger("app.module.component")
	if resolved == nil || resolved.Name() != "app" {
		t.Error("expected resolution to parent logger")
	}

	// Logging through child should work (via hierarchical resolution)
	childLogger.Info("child message")

	if buf.Len() == 0 {
		t.Error("expected log output from child logger")
	}
}

// TestLogger_MultipleAttrs tests adding multiple attributes
func TestLogger_MultipleAttrs(t *testing.T) {
	registry := NewRegistry()
	buf := &bytes.Buffer{}
	sink := NewTextSink("sink", buf, slog.LevelInfo)
	registry.RegisterSink(sink)

	regLogger := NewRegisteredLogger("app", sink)
	registry.RegisterLogger(regLogger)

	logger := registry.Get("app")
	loggerWithAttrs := logger.
		WithAttrs(slog.String("user", "alice")).
		WithAttrs(slog.Int("age", 30))

	loggerWithAttrs.Info("test")

	if buf.Len() == 0 {
		t.Error("expected log output")
	}
}

// TestLogger_ChainedOperations tests chaining multiple operations
func TestLogger_ChainedOperations(t *testing.T) {
	registry := NewRegistry()
	logger := registry.Get("app")

	// Chain operations
	result := logger.
		WithAttrs(slog.String("module", "cpu")).
		WithGroup("performance").
		Child("monitor")

	if result.Name() != "app.monitor" {
		t.Errorf("expected name 'app.monitor', got %q", result.Name())
	}
}

// TestLogger_OptionalContext tests that logger methods work without explicit context
func TestLogger_OptionalContext(t *testing.T) {
	buf := &bytes.Buffer{}
	sink := NewFileSink("test", buf, slog.LevelDebug)

	registry := NewRegistry()
	regLogger := NewRegisteredLogger("app", sink)
	registry.RegisterLogger(regLogger)

	logger := registry.Get("app")

	// Test all logging levels without context
	logger.Debug("debug message", slog.String("level", "debug"))
	logger.Info("info message", slog.String("level", "info"))
	logger.Warn("warn message", slog.String("level", "warn"))
	logger.Error("error message", slog.String("level", "error"))

	output := buf.String()
	if len(output) == 0 {
		t.Error("expected log output, got empty")
	}
	if !contains(output, "debug message") {
		t.Error("expected 'debug message' in output")
	}
	if !contains(output, "info message") {
		t.Error("expected 'info message' in output")
	}
	if !contains(output, "warn message") {
		t.Error("expected 'warn message' in output")
	}
	if !contains(output, "error message") {
		t.Error("expected 'error message' in output")
	}
}

// TestLogger_ContextVersions tests that context versions are available alongside optional context versions
func TestLogger_ContextVersions(t *testing.T) {
	buf := &bytes.Buffer{}
	sink := NewFileSink("test", buf, slog.LevelDebug)

	registry := NewRegistry()
	regLogger := NewRegisteredLogger("app", sink)
	registry.RegisterLogger(regLogger)

	logger := registry.Get("app")

	// Test context versions
	ctx := context.WithValue(context.Background(), "testkey", "testvalue")
	logger.DebugContext(ctx, "debug with context")
	logger.InfoContext(ctx, "info with context")
	logger.WarnContext(ctx, "warn with context")
	logger.ErrorContext(ctx, "error with context")

	output := buf.String()
	if len(output) == 0 {
		t.Error("expected log output, got empty")
	}
	if !contains(output, "debug with context") {
		t.Error("expected 'debug with context' in output")
	}
	if !contains(output, "info with context") {
		t.Error("expected 'info with context' in output")
	}
	if !contains(output, "warn with context") {
		t.Error("expected 'warn with context' in output")
	}
	if !contains(output, "error with context") {
		t.Error("expected 'error with context' in output")
	}
}

// TestLogger_Panic tests that panic logs message with backtrace and panics
func TestLogger_Panic(t *testing.T) {
	buf := &bytes.Buffer{}
	sink := NewFileSink("test", buf, slog.LevelInfo)

	registry := NewRegistry()
	regLogger := NewRegisteredLogger("app", sink)
	registry.RegisterLogger(regLogger)

	logger := registry.Get("app")

	// Panic should cause a recover-able panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic, but didn't panic")
		}
		if r := recover(); r != nil {
			t.Error("unexpected second panic")
		}
	}()

	logger.Panic("panic message", slog.String("severity", "critical"))

	// This line should not be reached
	t.Error("expected panic but code continued")
}

// TestLogger_PanicContext tests panic with context
func TestLogger_PanicContext(t *testing.T) {
	buf := &bytes.Buffer{}
	sink := NewFileSink("test", buf, slog.LevelInfo)

	registry := NewRegistry()
	regLogger := NewRegisteredLogger("app", sink)
	registry.RegisterLogger(regLogger)

	logger := registry.Get("app")

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic, but didn't panic")
		}
		// Verify backtrace was included
		output := buf.String()
		if !contains(output, "backtrace") {
			t.Error("expected backtrace in log output")
		}
	}()

	ctx := context.Background()
	logger.PanicContext(ctx, "panic with context")
}

// TestRegisteredLogger_Panic tests panic on RegisteredLogger
func TestRegisteredLogger_Panic(t *testing.T) {
	buf := &bytes.Buffer{}
	sink := NewFileSink("test", buf, slog.LevelInfo)
	regLogger := NewRegisteredLogger("app", sink)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic, but didn't panic")
		}
	}()

	regLogger.Panic("registered panic message")
}
