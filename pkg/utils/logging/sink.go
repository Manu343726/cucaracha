package logging

import (
	"io"
	"log/slog"
	"log/syslog"
)

// Sink represents an immutable destination where log entries are written.
// A sink handles the actual writing (to file, stdout, syslog, etc.)
// and enforces a minimum log level.
// Sinks are configured at creation time and cannot be modified afterward.
type Sink struct {
	name       string
	handler    slog.Handler
	level      slog.Level
	uiSinkImpl interface{} // Optional: stores *UISink if this is a UI sink (for internal use only)
}

// NewSink creates a new immutable sink with the given name, handler, and level.
func NewSink(name string, handler slog.Handler, level slog.Level) *Sink {
	return &Sink{
		name:    name,
		handler: handler,
		level:   level,
	}
}

// NewFileSink creates an immutable sink that writes to a file with JSON formatting.
func NewFileSink(name string, w io.Writer, level slog.Level) *Sink {
	opts := &slog.HandlerOptions{Level: level}
	handler := slog.NewJSONHandler(w, opts)
	return NewSink(name, handler, level)
}

// NewTextSink creates an immutable sink that writes to a writer with text formatting.
func NewTextSink(name string, w io.Writer, level slog.Level) *Sink {
	opts := &slog.HandlerOptions{Level: level}
	handler := slog.NewTextHandler(w, opts)
	return NewSink(name, handler, level)
}

// NewSyslogSink creates an immutable sink that writes to syslog with JSON formatting.
// It connects to the system's syslog service using Unix domain sockets.
// The tag identifies the application in syslog entries.
func NewSyslogSink(name string, tag string, level slog.Level) (*Sink, error) {
	// Connect to syslog with INFO priority (can be overridden by handler level)
	writer, err := syslog.New(syslog.LOG_USER|syslog.LOG_INFO, tag)
	if err != nil {
		return nil, err
	}

	opts := &slog.HandlerOptions{Level: level}
	handler := slog.NewJSONHandler(writer, opts)
	return NewSink(name, handler, level), nil
}

// Name returns the sink's name.
func (s *Sink) Name() string {
	return s.name
}

// Level returns the sink's minimum log level.
func (s *Sink) Level() slog.Level {
	return s.level
}

// Handle writes a log record through the sink's handler.
// Returns nil if successful, or an error if the record's level is below the sink's level.
func (s *Sink) Handle(record slog.Record) error {
	if record.Level < s.level {
		return nil
	}

	return s.handler.Handle(nil, record)
}

// Handler returns the sink's underlying slog handler.
func (s *Sink) Handler() slog.Handler {
	return s.handler
}
