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

	return &RegisteredLogger{
		name:   name,
		logger: slog.New(handler),
	}
}

// Name returns the registered logger's name.
func (rl *RegisteredLogger) Name() string {
	return rl.name
}

// Debug logs at debug level with context.
func (rl *RegisteredLogger) DebugContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	rl.logger.LogAttrs(ctx, slog.LevelDebug, msg, attrs...)
}

// Debug logs at debug level with background context.
func (rl *RegisteredLogger) Debug(msg string, attrs ...slog.Attr) {
	rl.logger.LogAttrs(context.Background(), slog.LevelDebug, msg, attrs...)
}

// Info logs at info level with context.
func (rl *RegisteredLogger) InfoContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	rl.logger.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
}

// Info logs at info level with background context.
func (rl *RegisteredLogger) Info(msg string, attrs ...slog.Attr) {
	rl.logger.LogAttrs(context.Background(), slog.LevelInfo, msg, attrs...)
}

// Warn logs at warn level with context.
func (rl *RegisteredLogger) WarnContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	rl.logger.LogAttrs(ctx, slog.LevelWarn, msg, attrs...)
}

// Warn logs at warn level with background context.
func (rl *RegisteredLogger) Warn(msg string, attrs ...slog.Attr) {
	rl.logger.LogAttrs(context.Background(), slog.LevelWarn, msg, attrs...)
}

// Error logs at error level with context.
func (rl *RegisteredLogger) ErrorContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	rl.logger.LogAttrs(ctx, slog.LevelError, msg, attrs...)
}

// Error logs at error level with background context.
func (rl *RegisteredLogger) Error(msg string, attrs ...slog.Attr) {
	rl.logger.LogAttrs(context.Background(), slog.LevelError, msg, attrs...)
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
	rl.logger.LogAttrs(ctx, slog.LevelError, msg, allAttrs...)

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

func (l *Logger) Timed(name string, fn func(*Logger) error) error {
	start := time.Now()
	log := l.Child(name)
	log.Debug("starting")
	err := fn(log)
	elapsed := time.Since(start)
	log.Debug(name, slog.Duration("elapsed", elapsed))
	return err
}

// logContext is the internal method that forwards to the registered logger with context.
func (l *Logger) logContext(ctx context.Context, level slog.Level, msg string, attrs ...any) {
	// Resolve the registered logger via hierarchy
	registeredLogger := l.registry.ResolveLogger(l.name)
	if registeredLogger == nil {
		// No registered logger found - this shouldn't happen in normal use
		return
	}

	// Combine pre-configured attrs with provided attrs
	allAttrs := make([]any, len(l.attrs)+len(attrs))
	copy(allAttrs, l.attrs)
	copy(allAttrs[len(l.attrs):], attrs)

	// Convert to slog.Attr for the logger
	slogAttrs := make([]slog.Attr, 0, len(allAttrs))
	for _, a := range allAttrs {
		if attr, ok := a.(slog.Attr); ok {
			slogAttrs = append(slogAttrs, attr)
		}
	}

	registeredLogger.logger.LogAttrs(ctx, level, msg, slogAttrs...)
}

// log is the internal method that forwards to the registered logger with background context.
func (l *Logger) log(level slog.Level, msg string, attrs ...any) {
	l.logContext(context.Background(), level, msg, attrs...)
}

// DebugContext logs at debug level with context.
func (l *Logger) DebugContext(ctx context.Context, msg string, attrs ...any) {
	l.logContext(ctx, slog.LevelDebug, msg, attrs...)
}

// Debug logs at debug level with background context.
func (l *Logger) Debug(msg string, attrs ...any) {
	l.log(slog.LevelDebug, msg, attrs...)
}

// InfoContext logs at info level with context.
func (l *Logger) InfoContext(ctx context.Context, msg string, attrs ...any) {
	l.logContext(ctx, slog.LevelInfo, msg, attrs...)
}

// Info logs at info level with background context.
func (l *Logger) Info(msg string, attrs ...any) {
	l.log(slog.LevelInfo, msg, attrs...)
}

// WarnContext logs at warn level with context.
func (l *Logger) WarnContext(ctx context.Context, msg string, attrs ...any) {
	l.logContext(ctx, slog.LevelWarn, msg, attrs...)
}

// Warn logs at warn level with background context.
func (l *Logger) Warn(msg string, attrs ...any) {
	l.log(slog.LevelWarn, msg, attrs...)
}

// ErrorContext logs at error level with context.
func (l *Logger) ErrorContext(ctx context.Context, msg string, attrs ...any) {
	l.logContext(ctx, slog.LevelError, msg, attrs...)
}

// Error logs at error level with background context.
func (l *Logger) Error(msg string, attrs ...any) {
	l.log(slog.LevelError, msg, attrs...)
}

// Like Error(), but takes format string and arguments, builds an error through fmt.Errorf, logs it, and returns the error.
func (l *Logger) Errorf(msg string, args ...any) error {
	err := fmt.Errorf(msg, args...)
	l.Error(err.Error())
	return err
}

// LogfContext logs a formatted message at the specified level with context.
func (l *Logger) LogfContext(ctx context.Context, level slog.Level, format string, args ...interface{}) {
	l.logContext(ctx, level, fmt.Sprintf(format, args...))
}

// Logf logs a formatted message at the specified level with background context.
func (l *Logger) Logf(level slog.Level, format string, args ...interface{}) {
	l.log(level, fmt.Sprintf(format, args...))
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
	l.logContext(ctx, slog.LevelError, msg, allAttrs...)

	// Panic
	panic(msg)
}

// Panic logs at panic level with backtrace and then panics.
func (l *Logger) Panic(msg string, attrs ...any) {
	l.PanicContext(context.Background(), msg, attrs...)
}
