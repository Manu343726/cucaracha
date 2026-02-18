package logging

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// UILogEntry represents a log entry captured by the UI sink.
type UILogEntry struct {
	Time    time.Time
	Level   string
	Logger  string
	Message string
	Attrs   map[string]string
}

func (e UILogEntry) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "[%s][%s][%s] %s\n", e.Time.Format("2006-01-02 15:04:05"), e.Level, e.Logger, e.Message)

	// Show attributes as (key1=value1, key2=value2, ...)  after the message if there are any
	if len(e.Attrs) > 0 {
		buf.WriteString("(")
		first := true
		for k, v := range e.Attrs {
			if !first {
				buf.WriteString(", ")
			}
			fmt.Fprintf(&buf, "%s=%s", k, v)
			first = false
		}
		buf.WriteString(")")
	}
	return buf.String()
}

// UISink is a thread-safe sink that collects log entries for UI display.
// It stores recent log entries in memory and provides methods to retrieve them.
type UISink struct {
	mu         sync.RWMutex
	entries    []UILogEntry
	maxEntries int
	level      slog.Level
}

// uiSinkHandler wraps the UISink to implement slog.Handler interface
type uiSinkHandler struct {
	uiSink   *UISink
	callback UILogEntryCallback
}

// newUISinkHandler creates a new handler for the UI sink
func newUISinkHandler(uiSink *UISink, callback UILogEntryCallback) *uiSinkHandler {
	return &uiSinkHandler{
		callback: callback,
		uiSink:   uiSink,
	}
}

// Enabled implements slog.Handler
func (h *uiSinkHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.uiSink.level
}

// Handle implements slog.Handler
func (h *uiSinkHandler) Handle(ctx context.Context, record slog.Record) error {
	if !h.Enabled(ctx, record.Level) {
		return nil
	}

	h.uiSink.mu.Lock()
	defer h.uiSink.mu.Unlock()

	var loggerName string

	// Extract attributes from the record
	attrs := make(map[string]string)
	record.Attrs(func(a slog.Attr) bool {
		if a.Key == "cucaracha.logging.logger.name" {
			loggerName = fmt.Sprintf("%v", a.Value.Any())
		} else {
			attrs[a.Key] = fmt.Sprintf("%v", a.Value.Any())
		}
		return true
	})

	entry := UILogEntry{
		Time:    record.Time,
		Level:   record.Level.String(),
		Logger:  loggerName,
		Message: record.Message,
		Attrs:   attrs,
	}

	h.uiSink.entries = append(h.uiSink.entries, entry)

	// Remove oldest entries if we exceed maxEntries
	if len(h.uiSink.entries) > h.uiSink.maxEntries {
		h.uiSink.entries = h.uiSink.entries[len(h.uiSink.entries)-h.uiSink.maxEntries:]
	}

	// Invoke the callback if provided
	if h.callback != nil {
		h.callback(entry)
	}

	return nil
}

// WithAttrs implements slog.Handler
func (h *uiSinkHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h // For now, we ignore additional attributes
}

// WithGroup implements slog.Handler
func (h *uiSinkHandler) WithGroup(name string) slog.Handler {
	return h // For now, we ignore groups
}

type UILogEntryCallback func(entry UILogEntry)

// NewUISink creates a new UI sink with a configurable maximum number of entries.
// Older entries are removed when the limit is exceeded.
// Returns a *Sink that can be registered with the logging registry.
func NewUISink(name string, level slog.Level, maxEntries int, callback UILogEntryCallback) *Sink {
	uiSinkImpl := &UISink{
		maxEntries: maxEntries,
		level:      level,
		entries:    make([]UILogEntry, 0, maxEntries),
	}

	handler := newUISinkHandler(uiSinkImpl, callback)

	// Create and return a proper Sink
	sink := &Sink{
		name:    name,
		handler: handler,
		level:   level,
	}

	// Store a reference to the UISink implementation in the sink
	// We'll retrieve it later via GetUIEntries
	sink.uiSinkImpl = uiSinkImpl

	return sink
}

// GetUIEntries retrieves the UISink implementation from a Sink.
// Returns the UISink and true if this Sink contains a UI sink implementation, false otherwise.
func GetUIEntries(sink *Sink) (*UISink, bool) {
	if sink != nil && sink.uiSinkImpl != nil {
		if uiSink, ok := sink.uiSinkImpl.(*UISink); ok {
			return uiSink, true
		}
	}
	return nil, false
}

// GetEntries returns a copy of all captured log entries.
func (u *UISink) GetEntries() []UILogEntry {
	u.mu.RLock()
	defer u.mu.RUnlock()

	entries := make([]UILogEntry, len(u.entries))
	copy(entries, u.entries)
	return entries
}

// GetRecentEntries returns the n most recent log entries.
func (u *UISink) GetRecentEntries(n int) []UILogEntry {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if n >= len(u.entries) {
		entries := make([]UILogEntry, len(u.entries))
		copy(entries, u.entries)
		return entries
	}

	entries := make([]UILogEntry, n)
	copy(entries, u.entries[len(u.entries)-n:])
	return entries
}

// Clear removes all stored log entries.
func (u *UISink) Clear() {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.entries = u.entries[:0]
}

// Count returns the number of stored log entries.
func (u *UISink) Count() int {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return len(u.entries)
}

// FormatEntries formats log entries for display.
// Each entry is formatted as: [LEVEL] HH:MM:SS message
func (u *UISink) FormatEntries(entries []UILogEntry) string {
	var buf bytes.Buffer
	for _, entry := range entries {
		fmt.Fprintf(&buf, "[%s] %s %s\n", entry.Level, entry.Time.Format("15:04:05"), entry.Message)
		for k, v := range entry.Attrs {
			fmt.Fprintf(&buf, "  %s=%s\n", k, v)
		}
	}
	return buf.String()
}
