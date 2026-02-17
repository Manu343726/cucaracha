# New Logging API - Quick Reference

## Three-Tier Architecture

```
┌─────────────────────────────────────────┐
│  User Code: Logger (Wrapper)            │
│  ├─ registry.Get("app")                 │
│  ├─ logger.WithAttrs(...)               │
│  └─ logger.Info/Debug/Warn/Error        │
└────────────┬────────────────────────────┘
             │ hierarchical resolution
             ↓
┌─────────────────────────────────────────┐
│  Registry: Maps names to workers        │
│  ├─ register.RegisterLogger(regLogger)  │
│  ├─ registry.Get(name) → Logger         │
│  └─ registry.ResolveLogger(name)        │
└────────────┬────────────────────────────┘
             │ returns RegisteredLogger
             ↓
┌─────────────────────────────────────────┐
│  RegisteredLogger: Real Worker          │
│  ├─ slog.Logger with Fanout handler     │
│  ├─ connects multiple sinks             │
│  └─ does actual logging work            │
└────────────┬────────────────────────────┘
             │ distributes via Fanout
             ↓
┌─────────────────────────────────────────┐
│  Sinks: Destinations                    │
│  ├─ stdout/stderr                       │
│  ├─ file                                │
│  └─ any slog.Handler                    │
└─────────────────────────────────────────┘
```

## Setup (One Time)

### Option 1: Programmatic Setup

```go
registry := logging.NewRegistry()

// Create sinks
consoleSink := logging.NewTextSink("console", os.Stdout, slog.LevelInfo)
fileSink := logging.NewFileSink("file", file, slog.LevelDebug)

// Register sinks
registry.RegisterSink(consoleSink)
registry.RegisterSink(fileSink)

// Create and register registered loggers
runtime := logging.NewRegisteredLogger("runtime", consoleSink, fileSink)
cpu := logging.NewRegisteredLogger("runtime.cpu", consoleSink)

registry.RegisterLogger(runtime)
registry.RegisterLogger(cpu)
```

### Option 2: YAML Setup (Recommended)

```yaml
# config/logging.yaml
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

loggers:
  - name: runtime
    sinks: [console, file]
  - name: runtime.cpu
    sinks: [console]
  - name: debugger
    sinks: [file]
```

```go
config, _ := logging.NewConfigFromYAML("config/logging.yaml")
registry := logging.NewRegistry()
config.Apply(registry)
```

## Usage (Many Times)

### Basic Logging

```go
logger := registry.Get("runtime.cpu")
ctx := context.Background()

// Log messages
logger.Info(ctx, "CPU initialized", slog.Int("cores", 4))
logger.Error(ctx, "CPU error", slog.String("reason", "timeout"))
```

### Pre-Configure Context

```go
// Create logger with pre-configured attributes
cpuLogger := registry.Get("runtime.cpu").
    WithAttrs(slog.String("component", "ALU"))

// Automatically includes component="ALU" in all logs
cpuLogger.Info(ctx, "ALU operation started")
cpuLogger.Error(ctx, "ALU fault detected")
```

### Chain Operations

```go
logger := registry.Get("app")

// Chain to build complex context
debugLogger := logger.
    WithAttrs(slog.String("mode", "debug")).
    WithGroup("performance").
    Child("monitor")

debugLogger.Info(ctx, "monitoring started")
```

### Hierarchical Resolution

```go
// Only "runtime" is registered
registry.Get("runtime.cpu.executor").Info(ctx, "msg")
// → Resolves to "runtime" logger
// → Logs via console+file sinks

registry.Get("runtime").Info(ctx, "msg")
// → Exact match
// → Logs via console+file sinks

registry.Get("unknown.path").Info(ctx, "msg")
// → No match
// → Silent (no logging)
```

## Key Differences from Old API

### Creating Loggers

```go
// OLD (no longer works)
logger := logging.NewLogger("app", sink1, sink2)
registry.RegisterLogger(logger)

// NEW
regLogger := logging.NewRegisteredLogger("app", sink1, sink2)
registry.RegisterLogger(regLogger)
logger := registry.Get("app")
```

### Getting Loggers

```go
// OLD (no longer works)
logger, err := registry.GetLogger("app")

// NEW
logger := registry.Get("app")
```

### Using Logger

```go
// OLD
logger.Info(ctx, "message")

// NEW
logger.Info(ctx, "message", slog.String("key", "value"))
```

### Context Attributes

```go
// OLD - repeat attributes each time
logger.Info(ctx, "msg1", slog.String("module", "cpu"))
logger.Info(ctx, "msg2", slog.String("module", "cpu"))

// NEW - pre-configure once
cpuLogger := logger.WithAttrs(slog.String("module", "cpu"))
cpuLogger.Info(ctx, "msg1")
cpuLogger.Info(ctx, "msg2")
```

## Multiple Sinks

With Fanout handler, one call goes to all sinks:

```go
regLogger := logging.NewRegisteredLogger("app", sink1, sink2, sink3)

// Call Info once
regLogger.Info(ctx, "message")
// ↓ automatically goes to all 3 sinks
// sink1 ✓
// sink2 ✓
// sink3 ✓
```

## Complete Example

```go
package main

import (
    "context"
    "log/slog"
    "os"
    "github.com/Manu343726/cucaracha/pkg/logging"
)

func main() {
    // Setup
    registry := logging.NewRegistry()
    
    console := logging.NewTextSink("console", os.Stdout, slog.LevelInfo)
    registry.RegisterSink(console)
    
    app := logging.NewRegisteredLogger("app", console)
    registry.RegisterLogger(app)
    
    // Usage
    ctx := context.Background()
    
    // Get logger
    logger := registry.Get("app")
    
    // Pre-configure context
    userLogger := logger.WithAttrs(slog.String("user", "alice"))
    
    // Log with pre-configured context
    userLogger.Info(ctx, "user login")        // includes user="alice"
    userLogger.Info(ctx, "user action")       // includes user="alice"
    
    // Log without context
    logger.Info(ctx, "system event")          // user not included
}
```

## Common Patterns

### Module-Specific Logger

```go
type Database struct {
    logger *logging.Logger
}

func NewDatabase(registry *logging.Registry) *Database {
    return &Database{
        logger: registry.Get("app.database"),
    }
}

func (db *Database) Connect() error {
    db.logger.Info(ctx, "connecting")
    // ...
}
```

### Request-Scoped Logger

```go
handler := func(w http.ResponseWriter, r *http.Request) {
    requestID := r.Header.Get("X-Request-ID")
    
    // Create request-scoped logger with context
    logger := baseLogger.WithAttrs(slog.String("request_id", requestID))
    
    logger.Info(r.Context(), "request start")
    // ... request handling ...
    logger.Info(r.Context(), "request complete")
}
```

### Testing

```go
func TestDatabase(t *testing.T) {
    registry := logging.NewRegistry()
    
    buf := &bytes.Buffer{}
    testSink := logging.NewTextSink("test", buf, slog.LevelDebug)
    registry.RegisterSink(testSink)
    
    db := logging.NewRegisteredLogger("test.db", testSink)
    registry.RegisterLogger(db)
    
    // Test code
    logger := registry.Get("test.db")
    logger.Info(context.Background(), "test message")
    
    output := buf.String()
    assert.Contains(t, output, "test message")
}
```

## Advanced: Fanout with Multiple Formats

```go
// Same information, different formats
jsonFile := logging.NewFileSink("json_log", jsonFile, slog.LevelDebug)
textFile := logging.NewFileSink("text_log", textFile, slog.LevelInfo)

// One logger, two outputs, different formats
app := logging.NewRegisteredLogger("app", jsonFile, textFile)

// Call once, Fanout distributes to both:
app.Info(ctx, "message")
// → jsonFile gets JSON format
// → textFile gets text format
```
