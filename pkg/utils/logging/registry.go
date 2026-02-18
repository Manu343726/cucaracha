package logging

import (
	"fmt"
	"slices"
	"strings"
	"sync"
)

// Registry manages registered logger and sink lifecycle.
// It supports hierarchical logger lookup with best-match resolution.
// RegisteredLoggers are immutable once registered.
// Logger instances (thin wrappers) are created on-demand via Get().
type Registry struct {
	registeredLoggers map[string]*RegisteredLogger
	sinks             map[string]*Sink
	mu                sync.RWMutex
}

// NewRegistry creates a new empty registry.
func NewRegistry() *Registry {
	return &Registry{
		registeredLoggers: make(map[string]*RegisteredLogger),
		sinks:             make(map[string]*Sink),
	}
}

// RegisterSink registers an immutable sink in the registry.
// Sinks can be used by multiple loggers.
// Returns an error if a sink with the same name already exists.
func (r *Registry) RegisterSink(sink *Sink) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sinks[sink.Name()]; exists {
		return fmt.Errorf("sink %q already registered", sink.Name())
	}

	r.sinks[sink.Name()] = sink
	return nil
}

// GetSink retrieves a sink by name.
func (r *Registry) GetSink(name string) (*Sink, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sink, exists := r.sinks[name]
	if !exists {
		return nil, fmt.Errorf("sink %q not found", name)
	}
	return sink, nil
}

// RemoveSink removes a sink from the registry.
// This is primarily used for registry cleanup and testing.
func (r *Registry) RemoveSink(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sinks[name]; !exists {
		return fmt.Errorf("sink %q not found", name)
	}

	delete(r.sinks, name)
	return nil
}

// RegisterLogger registers an immutable registered logger with a hierarchical name.
// The name should use dot notation: "package", "package.module", "package.module.Type"
// The logger must be fully configured with its sinks before registration.
// Returns an error if a logger with the same name already exists.
func (r *Registry) RegisterLogger(logger *RegisteredLogger) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.registeredLoggers[logger.Name()]; exists {
		return fmt.Errorf("logger %q already registered", logger.Name())
	}

	r.registeredLoggers[logger.Name()] = logger
	return nil
}

// GetRegisteredLogger retrieves a registered logger by exact name.
func (r *Registry) GetRegisteredLogger(name string) (*RegisteredLogger, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	logger, exists := r.registeredLoggers[name]
	if !exists {
		return nil, fmt.Errorf("logger %q not found", name)
	}
	return logger, nil
}

// Returns the set of registered loggers using the given sink
func (r *Registry) GetLoggersBySink(sink *Sink) []*RegisteredLogger {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.unsafe_getLoggersBySink(sink)
}

func (r *Registry) unsafe_getLoggersBySink(sink *Sink) []*RegisteredLogger {
	var loggers []*RegisteredLogger
	for _, logger := range r.registeredLoggers {
		for _, s := range logger.sinks {
			if s == sink {
				loggers = append(loggers, logger)
				break
			}
		}
	}
	return loggers
}

// Returns the set of registered loggers not currently using the given sink
func (r *Registry) GetLoggersNotUsingSink(sink *Sink) []*RegisteredLogger {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.unsafe_getLoggersNotUsingSink(sink)
}

func (r *Registry) unsafe_getLoggersNotUsingSink(sink *Sink) []*RegisteredLogger {

	var loggers []*RegisteredLogger
	for _, logger := range r.registeredLoggers {
		usesSink := false
		for _, s := range logger.sinks {
			if s == sink {
				usesSink = true
				break
			}
		}
		if !usesSink {
			loggers = append(loggers, logger)
		}
	}
	return loggers
}

// Removes a given sink from all registered loggers that use it.
func (r *Registry) RemoveSinkFromLoggers(sink *Sink) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, logger := range r.unsafe_getLoggersBySink(sink) {
		newLogger := logger.WithoutSink(sink)
		r.registeredLoggers[logger.Name()] = newLogger
	}
}

// Adds the given sink to all the given loggers.
func (r *Registry) AddSinkToLoggers(sink *Sink, loggers []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, loggerName := range loggers {
		if _, exists := r.registeredLoggers[loggerName]; !exists {
			return fmt.Errorf("logger %q not found", loggerName)
		}
	}

	for _, logger := range r.unsafe_getLoggersNotUsingSink(sink) {
		if slices.Contains(loggers, logger.Name()) {
			newLogger := logger.WithAddedSinks(sink)
			r.registeredLoggers[logger.Name()] = newLogger
		}
	}

	return nil
}

// Get returns a thin wrapper Logger for the given hierarchical name.
// This is the main method users call to get a logger.
// The logger forwards calls to the best-matching registered logger.
func (r *Registry) Get(name string) *Logger {
	return NewLogger(name, r)
}

// ResolveLogger finds the best matching registered logger for a hierarchical name.
// Resolution is recursive: "package.module.Type" will match in this order:
// 1. "package.module.Type" (exact)
// 2. "package.module" (parent module)
// 3. "package" (parent package)
// 4. nil (no match)
//
// Returns the best matching registered logger, or nil if no logger matches.
func (r *Registry) ResolveLogger(name string) *RegisteredLogger {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Try exact match first
	if logger, exists := r.registeredLoggers[name]; exists {
		return logger
	}

	// Try parent hierarchies
	parts := strings.Split(name, ".")
	for i := len(parts) - 1; i > 0; i-- {
		parentName := strings.Join(parts[:i], ".")
		if logger, exists := r.registeredLoggers[parentName]; exists {
			return logger
		}
	}

	return nil
}

// ListLoggers returns all registered logger names.
func (r *Registry) ListLoggers() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.registeredLoggers))
	for name := range r.registeredLoggers {
		names = append(names, name)
	}
	return names
}

// ListSinks returns all registered sink names.
func (r *Registry) ListSinks() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.sinks))
	for name := range r.sinks {
		names = append(names, name)
	}
	return names
}

// Clear removes all loggers and sinks from the registry.
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.registeredLoggers = make(map[string]*RegisteredLogger)
	r.sinks = make(map[string]*Sink)
}
