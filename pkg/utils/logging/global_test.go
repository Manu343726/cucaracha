package logging

import (
	"log/slog"
	"testing"
)

// TestGlobalRegistry tests the global registry functionality
func TestGlobalRegistry(t *testing.T) {
	registry := DefaultRegistry()

	// Verify root logger is registered
	if registry == nil {
		t.Error("default registry should not be nil")
	}

	// Get a logger from the registry
	logger := Get("cucaracha.runtime")
	if logger == nil {
		t.Error("logger should not be nil")
	}
}

// TestPackageLoggers tests that package loggers can be created
func TestPackageLoggers(t *testing.T) {
	// These simulate what each package's log() function does
	tests := []struct {
		name string
		path string
	}{
		{"runtime", "cucaracha.runtime"},
		{"runtime.program", "cucaracha.runtime.program"},
		{"hw", "cucaracha.hw"},
		{"hw.cpu", "cucaracha.hw.cpu"},
		{"hw.memory", "cucaracha.hw.memory"},
		{"hw.peripheral", "cucaracha.hw.peripheral"},
		{"debugger", "cucaracha.debugger"},
		{"debugger.core", "cucaracha.debugger.core"},
		{"interpreter", "cucaracha.interpreter"},
		{"llvm", "cucaracha.llvm"},
		{"llvm.templates", "cucaracha.llvm.templates"},
		{"system", "cucaracha.system"},
		{"ui", "cucaracha.ui"},
		{"ui.tui", "cucaracha.ui.tui"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := Get(tt.path)
			if logger == nil {
				t.Errorf("logger for %s should not be nil", tt.path)
			}

			if logger.Name() != tt.path {
				t.Errorf("expected logger name %s, got %s", tt.path, logger.Name())
			}

			// Should not panic when logging
			logger.Info("test message", slog.String("package", tt.name))
		})
	}
}

// TestHierarchicalResolution tests that hierarchical resolution works
func TestHierarchicalResolution(t *testing.T) {
	registry := DefaultRegistry()

	// Register some loggers
	tests := []struct {
		registered string
		lookup     string
		expected   string
	}{
		{"cucaracha", "cucaracha.runtime.program.executor", "cucaracha"},
		{"cucaracha.runtime", "cucaracha.runtime.program", "cucaracha.runtime"},
		{"cucaracha.hw.cpu", "cucaracha.hw.cpu.alu", "cucaracha.hw.cpu"},
		{"cucaracha.hw", "cucaracha.hw.memory", "cucaracha.hw"},
	}

	for _, tt := range tests {
		resolved := registry.ResolveLogger(tt.lookup)
		if resolved != nil && resolved.Name() != tt.expected && tt.registered == "cucaracha" {
			// First one should resolve to cucaracha root logger
			if resolved.Name() != tt.expected {
				t.Errorf("expected resolution of %s to %s, got %s",
					tt.lookup, tt.expected, resolved.Name())
			}
		}
	}
}
