package logging

import (
	"log/slog"
	"strings"
)

// LoggingWriter wraps an io.Writer to log each line written to it.
// It buffers output and logs complete lines at the specified level.
type LoggingWriter struct {
	logger *Logger
	logLvl slog.Level
	buffer strings.Builder
}

// NewLoggingWriter creates a new LoggingWriter that logs each line at the specified level.
func NewLoggingWriter(logger *Logger, logLvl slog.Level) *LoggingWriter {
	return &LoggingWriter{
		logger: logger,
		logLvl: logLvl,
	}
}

// Write implements the io.Writer interface.
// It buffers input and logs complete lines (lines ending with \n).
func (lw *LoggingWriter) Write(p []byte) (n int, err error) {
	// Write to buffer
	lw.buffer.Write(p)

	// Process complete lines
	for {
		line := lw.readLine()
		if line == "" {
			break
		}
		if line != "" {
			lw.logger.Log(lw.logLvl, line)
		}
	}

	return len(p), nil
}

// readLine reads a complete line from the buffer and removes it.
// Returns empty string if no complete line is available.
func (lw *LoggingWriter) readLine() string {
	content := lw.buffer.String()
	idx := strings.Index(content, "\n")
	if idx < 0 {
		return "" // No complete line yet
	}

	line := content[:idx]
	remaining := content[idx+1:]
	lw.buffer.Reset()
	lw.buffer.WriteString(remaining)

	return strings.TrimSpace(line)
}

// Flush ensures any remaining buffered content is logged.
// Call this when you're done writing to ensure the last line (if any) is logged.
func (lw *LoggingWriter) Flush() {
	if lw.buffer.Len() > 0 {
		remaining := strings.TrimSpace(lw.buffer.String())
		if remaining != "" {
			lw.logger.Log(lw.logLvl, remaining)
		}
		lw.buffer.Reset()
	}
}
