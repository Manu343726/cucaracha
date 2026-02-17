# Hierarchical Logging System - Immutable Design

## Overview

The logging system provides hierarchical, structured logging using Go's `slog` with immutable loggers and sinks. This design ensures thread-safety, predictability, and clean architecture.

## Immutability Model

### Sinks (Fully Immutable)

Sinks are **completely immutable** once created. They cannot be modified after instantiation:

```go
// Create sink - fully configured at instantiation
sink := logging.NewFileSink("console", os.Stdout, slog.LevelInfo)
// No SetLevel() method - level is fixed

registry.RegisterSink(sink)
// Sink cannot be changed
```

**Benefits:**
- No synchronization needed for reads
- Safe to share across goroutines
- Configuration is visible at creation time

### Loggers (Immutable Core, Extensible)

Loggers are **immutable in their core configuration** (name, sinks) but can be **extended** with new attributes and groups:

```go
// Create logger - sinks attached at instantiation
logger := logging.NewLogger("runtime", sink1, sink2)
registry.RegisterLogger(logger)

// Logger cannot have sinks added/removed after creation
// But you can create new logger variations:

logger2 := logger.WithAttrs(slog.String("request_id", "123"))
logger3 := logger.WithGroup("debug")
logger4 := logger.Child("cpu")

// All are new logger instances, originals unchanged
```

**Benefits:**
- Core configuration is immutable and thread-safe
- Creating contextual variations is lightweight
- No mutations affect other references
- Easy to reason about

### Registry (Mutable for Setup, Stateless Lookup)

The registry is mutable during setup but loggers/sinks are immutable once registered:

```go
registry := logging.NewRegistry()

// Setup phase (can be done once at startup)
sink := logging.NewFileSink("console", os.Stdout, slog.LevelInfo)
registry.RegisterSink(sink)

logger := logging.NewLogger("app", sink)
registry.RegisterLogger(logger)

// After setup, registry is effectively read-only for the application
logger := registry.ResolveLogger("app.module.Type")
// No possibility of logger being modified
```

## Configuration with YAML

Loggers and sinks are fully specified in YAML and created with all configuration:

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

loggers:
  - name: runtime
    sinks: [console, file]

  - name: runtime.cpu
    sinks: [console, file]

  - name: debugger
    sinks: [file]
```

```go
config, _ := logging.NewConfigFromYAML("logging.yaml")
registry := logging.NewRegistry()
config.Apply(registry)

// All loggers and sinks are now immutable and ready to use
logger := registry.ResolveLogger("runtime.cpu.executor")
logger.Info(ctx, "execution started")
```

## Hierarchical Resolution

Loggers are resolved hierarchically - looking for the best match:

```go
// Registered loggers:
// - "package"
// - "package.module"

// Resolution:
registry.ResolveLogger("package")              // → "package"
registry.ResolveLogger("package.module")       // → "package.module"
registry.ResolveLogger("package.module.Type")  // → "package.module"
registry.ResolveLogger("package.other")        // → "package"
registry.ResolveLogger("other")                // → nil
```

## Creating Logger Variations

New logger instances can be created with additional context without modifying the original:

```go
baseLogger := registry.ResolveLogger("runtime.cpu")

// Create variation with attributes
ctxLogger := baseLogger.WithAttrs(
	slog.String("request_id", "req-123"),
	slog.Int("user_id", 456),
)

// Create variation with group
debugLogger := baseLogger.WithGroup("execution")

// Create child logger
childLogger := baseLogger.Child("executor")
// childLogger.Name() == "runtime.cpu.executor"

// All original references remain unchanged
```

## Thread Safety

**Registry:**
- Thread-safe for concurrent `ResolveLogger()` and `GetLogger()` calls
- Registration phase should happen once at startup
- Lock-free reads are possible after registration

**Loggers:**
- Immutable = inherently thread-safe
- Can be safely accessed from multiple goroutines
- `WithAttrs()`, `WithGroup()`, `Child()` create new instances

**Sinks:**
- Immutable = inherently thread-safe
- Can be safely shared across loggers and goroutines

## Usage Patterns

### Pattern 1: Application Startup

```go
func init() {
	config, _ := logging.NewConfigFromYAML("logging.yaml")
	globalRegistry := logging.NewRegistry()
	globalRegistry.Apply(config)
}

func GetLogger(name string) *logging.Logger {
	return globalRegistry.ResolveLogger(name)
}
```

### Pattern 2: Request Tracing

```go
func HandleRequest(reqID string) {
	logger := GetLogger("app.handler")
	
	// Create logger with request context
	reqLogger := logger.WithAttrs(slog.String("request_id", reqID))
	
	reqLogger.Info(ctx, "handling request")
	processRequest(reqLogger, reqID)
	reqLogger.Info(ctx, "request completed")
}

func processRequest(logger *logging.Logger, reqID string) {
	// Logger automatically has request_id in all logs
	logger.Info(ctx, "processing")
}
```

### Pattern 3: Hierarchical Logging

```go
// CPU module
runtimeLogger := GetLogger("runtime")
cpuLogger := runtimeLogger.Child("cpu")
executorLogger := cpuLogger.Child("executor")

ctx := context.Background()
executorLogger.Info(ctx, "executing instruction")
```

### Pattern 4: Testing

```go
func TestCPUExecution(t *testing.T) {
	// Create test sink
	buf := &bytes.Buffer{}
	sink := logging.NewFileSink("test", buf, slog.LevelDebug)
	
	// Create test logger
	logger := logging.NewLogger("test.cpu", sink)
	
	// Run test
	ctx := context.Background()
	logger.Info(ctx, "test started")
	
	// Verify logs
	if !strings.Contains(buf.String(), "test started") {
		t.Fatal("expected log not found")
	}
}
```

## Differences from Mutable Approach

| Aspect | Mutable | Immutable |
|--------|---------|-----------|
| **Thread Safety** | Requires locks | Lock-free reads |
| **State Changes** | Possible bugs | Impossible bugs |
| **Testing** | Inject before operations | Create new instances |
| **Reasoning** | Complex state machine | Simple, predictable |
| **Context Passing** | Share single logger | Pass contextual variations |
| **Performance** | Lock overhead | No overhead |

## Sink and Logger Immutability Enforced By

1. **Sink:**
   - No `SetLevel()` method
   - No public fields
   - Created with all configuration

2. **Logger:**
   - No `AddSink()` or `RemoveSink()` methods
   - Sinks passed in constructor
   - `WithAttrs()`, `WithGroup()`, `Child()` return new instances
   - No mutations to existing instances

3. **Registry:**
   - No `AttachSinkToLogger()` or `DetachSinkFromLogger()` methods
   - Registry setup is one-time (typically in `init()`)
   - Loggers fully configured before registration

## Best Practices

1. **Load configuration once** - typically in `init()` or `main()`
2. **Use ResolveLogger()** - for hierarchical name lookup
3. **Create variations** - use `WithAttrs()` for contextual logging, not global state
4. **Pass loggers** - pass to functions as parameters, don't rely on globals
5. **Test with specific sinks** - create test loggers with buffer sinks

## Migration from Mutable Design

If you previously had mutable loggers, migrate to immutable by:

1. Move sink attachment to logger creation
2. Use `WithAttrs()` instead of modifying logger state
3. Use `Child()` for hierarchical relationships
4. Configure everything in YAML or at startup
