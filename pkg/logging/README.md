# Hierarchical Logging Package

A hierarchical, structured logging system for Go applications using `slog` with immutable loggers and sinks, configured through YAML.

## Features

- **Hierarchical Logger Resolution** - Loggers are resolved by closest match in dot-notation hierarchy
- **Immutable Design** - Sinks and loggers are configured once and cannot be modified
- **Multiple Sinks** - Loggers can write to multiple sinks simultaneously
- **YAML Configuration** - Define loggers and sinks in a configuration file
- **Thread-Safe** - No locks needed for reads, safe for concurrent use
- **slog Integration** - Built on top of Go's standard `log/slog` package
- **Extensible** - Create contextual logger variations with `WithAttrs()`, `WithGroup()`, and `Child()`

## Installation

```go
import "github.com/Manu343726/cucaracha/pkg/logging"
```

## Quick Start

### Programmatic Setup

```go
package main

import (
	"context"
	"log/slog"
	"os"
	"github.com/Manu343726/cucaracha/pkg/logging"
)

func main() {
	// Create registry
	registry := logging.NewRegistry()

	// Create sinks
	console := logging.NewTextSink("console", os.Stdout, slog.LevelInfo)
	registry.RegisterSink(console)

	// Create loggers with sinks attached
	runtime := logging.NewLogger("runtime", console)
	registry.RegisterLogger(runtime)

	cpu := logging.NewLogger("runtime.cpu", console)
	registry.RegisterLogger(cpu)

	// Use logger
	ctx := context.Background()
	logger := registry.ResolveLogger("runtime.cpu.executor")
	logger.Info(ctx, "CPU initialized", slog.Int("cores", 4))
}
```

### YAML Configuration

Create `logging.yaml`:

```yaml
sinks:
  - name: console
    type: stdout
    level: info
    format: text

  - name: file
    type: file
    path: /var/log/app.log
    level: debug
    format: json

  - name: errors
    type: file
    path: /var/log/errors.log
    level: error
    format: text

loggers:
  - name: runtime
    sinks: [console, file]

  - name: runtime.cpu
    sinks: [console, file]

  - name: runtime.memory
    sinks: [console, file]

  - name: debugger
    sinks: [console, errors]

  - name: llvm
    sinks: [file]
```

Use in code:

```go
func init() {
	config, err := logging.NewConfigFromYAML("logging.yaml")
	if err != nil {
		panic(err)
	}

	registry := logging.NewRegistry()
	if err := config.Apply(registry); err != nil {
		panic(err)
	}
}

func main() {
	logger := registry.ResolveLogger("runtime.cpu.executor")
	logger.Info(context.Background(), "execution started")
}
```

## Core Concepts

### Loggers

Loggers are identified by hierarchical names using dot notation:

- `"package"` - Top-level logger
- `"package.module"` - Module logger
- `"package.module.Type"` - Type/component logger

Loggers are **immutable** - configured with sinks at creation time and cannot be changed.

```go
// Create logger with sinks
logger := logging.NewLogger("runtime.cpu", sink1, sink2)

// Create contextual variations
requestLogger := logger.WithAttrs(slog.String("request_id", "123"))
groupLogger := logger.WithGroup("debug")
childLogger := logger.Child("executor")

// All are new instances, original unchanged
```

### Sinks

Sinks are where logs are actually written. They are **immutable** and fully configured at creation.

**Built-in Sink Types:**
- `NewTextSink()` - Text format to writer
- `NewFileSink()` - JSON format to file or writer
- `NewSink()` - Custom handler

**Supported in YAML:**
- `stdout` - Standard output
- `stderr` - Standard error
- `file` - File on disk

**Log Levels:**
- `debug` - Debug messages
- `info` - Informational messages
- `warn` - Warning messages
- `error` - Error messages

### Registry

Central management of loggers and sinks.

```go
registry := logging.NewRegistry()

// Register sinks and loggers
registry.RegisterSink(sink)
registry.RegisterLogger(logger)

// Resolve loggers
logger := registry.ResolveLogger("runtime.cpu.executor")

// Hierarchical resolution - finds best match
// "runtime.cpu.executor" → matches "runtime.cpu" logger
// "runtime.cache" → matches "runtime" logger
```

### Hierarchical Resolution

Logger lookup uses hierarchical resolution to find the best matching logger:

```go
// Registered loggers:
// - "runtime"
// - "runtime.cpu"

registry.ResolveLogger("runtime")              // → "runtime"
registry.ResolveLogger("runtime.cpu")          // → "runtime.cpu"
registry.ResolveLogger("runtime.cpu.executor") // → "runtime.cpu"
registry.ResolveLogger("runtime.memory")       // → "runtime"
registry.ResolveLogger("debugger")             // → nil (no match)
```

## Usage Examples

### Basic Logging

```go
ctx := context.Background()
logger := registry.ResolveLogger("runtime.cpu")

logger.Debug(ctx, "debug message")
logger.Info(ctx, "info message")
logger.Warn(ctx, "warning message")
logger.Error(ctx, "error message")
```

### Logging with Attributes

```go
logger.Info(ctx, "instruction executed",
	slog.String("opcode", "ADD"),
	slog.Uint64("address", 0x1000),
	slog.Int("cycles", 2),
)
```

### Request Tracing

```go
func HandleRequest(reqID string) {
	logger := registry.ResolveLogger("handler")
	
	// Create logger with request context
	reqLogger := logger.WithAttrs(slog.String("request_id", reqID))
	
	reqLogger.Info(ctx, "request started")
	process(reqLogger)
	reqLogger.Info(ctx, "request completed")
}

func process(logger *logging.Logger) {
	// Logger automatically includes request_id in all logs
	logger.Info(ctx, "processing")
}
```

### Hierarchical Logging

```go
// Module hierarchy
runtimeLogger := registry.ResolveLogger("runtime")
cpuLogger := runtimeLogger.Child("cpu")
aluLogger := cpuLogger.Child("alu")

aluLogger.Info(ctx, "arithmetic operation", slog.String("op", "add"))
```

### Grouped Logging

```go
logger := registry.ResolveLogger("runtime")

debugLogger := logger.WithGroup("debug")
debugLogger.Info(ctx, "debug info")
// Output includes {"debug":{"message":"debug info",...}}
```

## Testing

```go
func TestCPUExecution(t *testing.T) {
	// Create test sink
	buf := &bytes.Buffer{}
	sink := logging.NewFileSink("test", buf, slog.LevelDebug)
	
	// Create test logger
	logger := logging.NewLogger("test.cpu", sink)
	
	// Run code that uses logger
	ctx := context.Background()
	logger.Info(ctx, "test started")
	
	// Verify logs
	output := buf.String()
	if !strings.Contains(output, "test started") {
		t.Fatal("expected log not found")
	}
}
```

## Immutability Benefits

| Benefit | Description |
|---------|-------------|
| **Thread-Safe** | No locks needed for reads |
| **Predictable** | Configuration cannot change unexpectedly |
| **Testable** | Easy to create test loggers with specific sinks |
| **Simple** | No state mutations, easy to reason about |
| **Performant** | No synchronization overhead |

## Configuration File Format

### Sinks

```yaml
sinks:
  - name: sink_name           # Unique identifier
    type: stdout|stderr|file  # Sink type
    path: /path/to/file       # Required for file type
    level: debug|info|warn|error  # Minimum log level
    format: json|text         # Output format (default: json)
```

### Loggers

```yaml
loggers:
  - name: logger.hierarchical.name  # Hierarchical identifier
    sinks:                          # List of sink names
      - sink_name1
      - sink_name2
```

## API Reference

### Sink

```go
// Create sinks
sink := logging.NewSink(name, handler, level)
sink := logging.NewTextSink(name, writer, level)
sink := logging.NewFileSink(name, writer, level)

// Query
sink.Name() string
sink.Level() slog.Level
```

### Logger

```go
// Create loggers
logger := logging.NewLogger(name, sinks...)

// Query
logger.Name() string
logger.Sinks() []*Sink

// Create variations (return new instances)
logger := logger.WithAttrs(attrs...)
logger := logger.WithGroup(name)
logger := logger.Child(name)

// Logging
logger.Debug(ctx, msg, attrs...)
logger.Info(ctx, msg, attrs...)
logger.Warn(ctx, msg, attrs...)
logger.Error(ctx, msg, attrs...)
logger.Logf(ctx, level, format, args...)
```

### Registry

```go
// Create
registry := logging.NewRegistry()

// Register
registry.RegisterSink(sink) error
registry.RegisterLogger(logger) error

// Query
registry.GetSink(name) (*Sink, error)
registry.GetLogger(name) (*Logger, error)
registry.ResolveLogger(name) *Logger  // Hierarchical lookup
registry.ListSinks() []string
registry.ListLoggers() []string

// Cleanup
registry.RemoveSink(name) error
registry.RemoveLogger(name) error
registry.Clear()
```

### Config

```go
// Load from YAML
config, err := logging.NewConfigFromYAML(path)
config, err := logging.NewConfigFromString(yaml)

// Apply
config.Apply(registry) error

// Example
exampleYAML := logging.ExampleConfig()
```

## Running Tests

```bash
go test ./pkg/logging/...
```

Run with coverage:

```bash
go test -cover ./pkg/logging/...
```

## Documentation

For more details, see:
- [IMMUTABLE_DESIGN.md](IMMUTABLE_DESIGN.md) - Design rationale and patterns
- [examples.go](examples.go) - Code examples
- [doc.go](doc.go) - Package documentation

## Architecture

```
┌─────────────────────────────────────────┐
│          Application Code                │
│  (Calls logger.Info(), logger.Error())   │
└──────────────────┬──────────────────────┘
                   │
                   ▼
        ┌──────────────────────┐
        │  Logger (Immutable)   │
        │  - name               │
        │  - sinks (immutable)  │
        │  - attrs              │
        └──────────────────────┘
                   │
        ┌──────────┴──────────┬─────────┐
        ▼                     ▼         ▼
    ┌────────┐           ┌────────┐  ┌────────┐
    │ Sink 1 │           │ Sink 2 │  │ Sink N │
    │ (File) │           │ (JSON) │  │(Syslog)│
    └────────┘           └────────┘  └────────┘
        │                    │            │
        ▼                    ▼            ▼
    [File]              [Stdout]      [Syslog]
```

## Contributing

When adding new features to the logging package:

1. Maintain immutability of sinks and loggers
2. Add unit tests for new functionality
3. Update documentation and examples
4. Ensure thread-safety properties are preserved

## License

Same as Cucaracha project.
