package logging

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"

	slogmulti "github.com/samber/slog-multi"
)

// RegisteredLogger represents the actual working logger with sinks.
// It's immutable: created with sinks and handlers, cannot be modified.
// RegisteredLoggers are stored in the Registry and do the actual logging work.
type RegisteredLogger struct {
	name   string
	sinks  []*Sink      // Stores the sinks for reconstruction
	logger *slog.Logger // with Fanout handler connecting multiple sinks
}

// NewRegisteredLogger creates a new immutable registered logger with a Fanout handler.
// The Fanout handler connects all provided sinks so logs go to all of them.
func NewRegisteredLogger(name string, sinks ...*Sink) *RegisteredLogger {
	// Convert sinks to handlers for Fanout
	handlers := make([]slog.Handler, 0, len(sinks))
	for _, sink := range sinks {
		handlers = append(handlers, sink.Handler())
	}

	// Create Fanout handler that writes to all sinks
	var handler slog.Handler
	if len(handlers) == 0 {
		// No sinks - use a no-op handler
		handler = slog.NewTextHandler(bytes.NewBuffer(nil), &slog.HandlerOptions{Level: slog.LevelDebug})
	} else if len(handlers) == 1 {
		handler = handlers[0]
	} else {
		handler = slogmulti.Fanout(handlers...)
	}

	// Store a copy of the sinks for potential reconstruction
	sinksCopy := make([]*Sink, len(sinks))
	copy(sinksCopy, sinks)

	return &RegisteredLogger{
		name:   name,
		sinks:  sinksCopy,
		logger: slog.New(handler),
	}
}

// GetSinks returns the sinks attached to this logger.
func (rl *RegisteredLogger) GetSinks() []*Sink {
	return rl.sinks
}

// Returns an equivalent logger with a different set of sinks
func (rl *RegisteredLogger) WithSinks(newSinks ...*Sink) *RegisteredLogger {
	return NewRegisteredLogger(rl.name, newSinks...)
}

// Returns an equivalent logger with the specified sinks added (if they're not already attached)
func (rl *RegisteredLogger) WithAddedSinks(newSinks ...*Sink) *RegisteredLogger {
	existingSinks := make(map[*Sink]struct{})
	for _, s := range rl.sinks {
		existingSinks[s] = struct{}{}
	}

	var combinedSinks []*Sink
	combinedSinks = append(combinedSinks, rl.sinks...)
	for _, s := range newSinks {
		if _, exists := existingSinks[s]; !exists {
			combinedSinks = append(combinedSinks, s)
		}
	}

	return NewRegisteredLogger(rl.name, combinedSinks...)
}

// Returns an equivalent logger without the specified sink (if it exists)
func (rl *RegisteredLogger) WithoutSink(sink *Sink) *RegisteredLogger {
	var newSinks []*Sink
	for _, s := range rl.sinks {
		if s != sink {
			newSinks = append(newSinks, s)
		}
	}
	return NewRegisteredLogger(rl.name, newSinks...)
}

// Returns an equivalent logger without the specified sink name (if it exists)
func (rl *RegisteredLogger) WithoutSinkNamed(sinkName string) *RegisteredLogger {
	var newSinks []*Sink
	for _, s := range rl.sinks {
		if s.Name() != sinkName {
			newSinks = append(newSinks, s)
		}
	}
	return NewRegisteredLogger(rl.name, newSinks...)
}

// Name returns the registered logger's name.
func (rl *RegisteredLogger) Name() string {
	return rl.name
}

func (rl *RegisteredLogger) log(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	rl.logger.LogAttrs(ctx, level, msg, attrs...)
}

// Log logs a message at the specified level with context.
func (rl *RegisteredLogger) LogContext(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	rl.logger.LogAttrs(ctx, level, msg, attrs...)
}

// Logf logs a formatted message at the specified level with context.
func (rl *RegisteredLogger) LogfContext(ctx context.Context, level slog.Level, format string, args ...interface{}) {
	rl.LogContext(ctx, level, fmt.Sprintf(format, args...))
}

// Log logs a message at the specified level with background context.
func (rl *RegisteredLogger) Log(level slog.Level, msg string, attrs ...slog.Attr) {
	rl.LogContext(context.Background(), level, msg, attrs...)
}

// Logf logs a formatted message at the specified level with background context.
func (rl *RegisteredLogger) Logf(level slog.Level, format string, args ...interface{}) {
	rl.Log(level, fmt.Sprintf(format, args...))
}

// Debug logs at debug level with context.
func (rl *RegisteredLogger) DebugContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	rl.log(ctx, slog.LevelDebug, msg, attrs...)
}

// Debug logs at debug level with background context.
func (rl *RegisteredLogger) Debug(msg string, attrs ...slog.Attr) {
	rl.DebugContext(context.Background(), msg, attrs...)
}

// Info logs at info level with context.
func (rl *RegisteredLogger) InfoContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	rl.log(ctx, slog.LevelInfo, msg, attrs...)
}

// Info logs at info level with background context.
func (rl *RegisteredLogger) Info(msg string, attrs ...slog.Attr) {
	rl.InfoContext(context.Background(), msg, attrs...)
}

// Warn logs at warn level with context.
func (rl *RegisteredLogger) WarnContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	rl.log(ctx, slog.LevelWarn, msg, attrs...)
}

// Warn logs at warn level with background context.
func (rl *RegisteredLogger) Warn(msg string, attrs ...slog.Attr) {
	rl.WarnContext(context.Background(), msg, attrs...)
}

// Error logs at error level with context.
func (rl *RegisteredLogger) ErrorContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	rl.log(ctx, slog.LevelError, msg, attrs...)
}

// Error logs at error level with background context.
func (rl *RegisteredLogger) Error(msg string, attrs ...slog.Attr) {
	rl.ErrorContext(context.Background(), msg, attrs...)
}

// PanicContext logs at panic level with backtrace and then panics.
func (rl *RegisteredLogger) PanicContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	// Add backtrace to attributes
	backtrace := debug.Stack()
	backtraceAttr := slog.String("backtrace", string(backtrace))

	// Combine provided attrs with backtrace
	allAttrs := make([]slog.Attr, len(attrs)+1)
	copy(allAttrs, attrs)
	allAttrs[len(attrs)] = backtraceAttr

	// Log as error level with backtrace
	rl.log(ctx, slog.LevelError, msg, allAttrs...)

	// Panic
	panic(msg)
}

// Panic logs at panic level with backtrace and then panics.
func (rl *RegisteredLogger) Panic(msg string, attrs ...slog.Attr) {
	rl.PanicContext(context.Background(), msg, attrs...)
}

// Logger represents a thin wrapper around a RegisteredLogger.
// It stores pre-configured context (attributes and groups) that get applied
// to every log call without repeating them.
// Loggers do NOT do actual logging; they forward to the registered logger in the registry.
type Logger struct {
	name     string
	registry *Registry
	attrs    []any // slog.Attr or slog.String, etc.
}

// NewLogger creates a new thin wrapper logger that forwards to the registry.
// The logger stores a reference to the registry for hierarchical resolution.
// The logger itself does not hold sinks; it forwards to the registered logger.
func NewLogger(name string, registry *Registry) *Logger {
	return &Logger{
		name:     name,
		registry: registry,
		attrs:    make([]any, 0),
	}
}

// Name returns the logger's hierarchical name.
func (l *Logger) Name() string {
	return l.name
}

// WithAttrs returns a new logger with additional pre-configured attributes.
// These attributes will be included in all log entries from the returned logger.
// The original logger is not modified (immutable).
func (l *Logger) WithAttrs(attrs ...any) *Logger {
	newAttrs := make([]any, len(l.attrs)+len(attrs))
	copy(newAttrs, l.attrs)
	copy(newAttrs[len(l.attrs):], attrs)

	return &Logger{
		name:     l.name,
		registry: l.registry,
		attrs:    newAttrs,
	}
}

// WithGroup returns a new logger with a pre-configured group.
// All log entries will be grouped under this name.
// The original logger is not modified (immutable).
func (l *Logger) WithGroup(group string) *Logger {
	newAttrs := make([]any, len(l.attrs)+1)
	copy(newAttrs, l.attrs)
	newAttrs[len(l.attrs)] = slog.Group(group)

	return &Logger{
		name:     l.name,
		registry: l.registry,
		attrs:    newAttrs,
	}
}

// Child creates a child logger with an extended hierarchical name.
// The child inherits all context (attributes and groups) from the parent.
// The child is a new logger and modifying it does not affect the parent.
func (l *Logger) Child(name string) *Logger {
	childName := l.name + "." + name
	newAttrs := make([]any, len(l.attrs))
	copy(newAttrs, l.attrs)

	return &Logger{
		name:     childName,
		registry: l.registry,
		attrs:    newAttrs,
	}
}

// Timed executes the provided function while logging the start and end of the operation with elapsed time.
func (l *Logger) Timed(name string, fn func(*Logger) error) error {
	start := time.Now()
	log := l.Child(name)
	log.Debug("starting")
	err := fn(log)
	elapsed := time.Since(start)
	log.Debug(name, slog.Duration("elapsed", elapsed))
	return err
}

// Logs a message at the specified level with context
func (l *Logger) LogContext(ctx context.Context, level slog.Level, msg string, attrs ...any) {
	// Resolve the registered logger via hierarchy
	registeredLogger := l.registry.ResolveLogger(l.name)
	if registeredLogger == nil {
		// No registered logger found - this shouldn't happen in normal use
		return
	}

	// Combine pre-configured attrs with provided attrs
	allAttrs := make([]any, 1+len(l.attrs)+len(attrs))
	allAttrs[0] = slog.String("cucaracha.logging.logger.name", l.name)
	copy(allAttrs[1:], l.attrs)
	copy(allAttrs[1+len(l.attrs):], attrs)

	registeredLogger.logger.Log(ctx, level, msg, allAttrs...)
}

// Logs at the specified level with background context.
func (l *Logger) Log(level slog.Level, msg string, attrs ...any) {
	l.LogContext(context.Background(), level, msg, attrs...)
}

// DebugContext logs at debug level with context.
func (l *Logger) DebugContext(ctx context.Context, msg string, attrs ...any) {
	l.LogContext(ctx, slog.LevelDebug, msg, attrs...)
}

// Debug logs at debug level with background context.
func (l *Logger) Debug(msg string, attrs ...any) {
	l.Log(slog.LevelDebug, msg, attrs...)
}

// InfoContext logs at info level with context.
func (l *Logger) InfoContext(ctx context.Context, msg string, attrs ...any) {
	l.LogContext(ctx, slog.LevelInfo, msg, attrs...)
}

// Info logs at info level with background context.
func (l *Logger) Info(msg string, attrs ...any) {
	l.Log(slog.LevelInfo, msg, attrs...)
}

// WarnContext logs at warn level with context.
func (l *Logger) WarnContext(ctx context.Context, msg string, attrs ...any) {
	l.LogContext(ctx, slog.LevelWarn, msg, attrs...)
}

// Warn logs at warn level with background context.
func (l *Logger) Warn(msg string, attrs ...any) {
	l.Log(slog.LevelWarn, msg, attrs...)
}

// ErrorContext logs at error level with context.
func (l *Logger) ErrorContext(ctx context.Context, msg string, attrs ...any) {
	l.LogContext(ctx, slog.LevelError, msg, attrs...)
}

// Error logs at error level with background context.
func (l *Logger) Error(msg string, attrs ...any) {
	l.Log(slog.LevelError, msg, attrs...)
}

// Like Error(), but takes format string and arguments, builds an error through fmt.Errorf, logs it, and returns the error.
func (l *Logger) Errorf(msg string, args ...any) error {
	err := fmt.Errorf(msg, args...)
	l.Error(err.Error())
	return err
}

// PanicContext logs at panic level with backtrace and then panics.
func (l *Logger) PanicContext(ctx context.Context, msg string, attrs ...any) {
	// Add backtrace to attributes
	backtrace := debug.Stack()
	backtraceAttr := slog.String("backtrace", string(backtrace))

	// Combine pre-configured attrs and provided attrs with backtrace
	allAttrs := make([]any, len(l.attrs)+len(attrs)+1)
	copy(allAttrs, l.attrs)
	copy(allAttrs[len(l.attrs):], attrs)
	allAttrs[len(l.attrs)+len(attrs)] = backtraceAttr

	// Log at error level (panic level is treated as error for slog)
	l.LogContext(ctx, slog.LevelError, msg, allAttrs...)

	// Panic
	panic(msg)
}

// Panic logs at panic level with backtrace and then panics.
func (l *Logger) Panic(msg string, attrs ...any) {
	l.PanicContext(context.Background(), msg, attrs...)
}
