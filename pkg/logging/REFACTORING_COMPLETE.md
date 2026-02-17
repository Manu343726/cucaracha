# Logging Architecture Refactoring - Complete

## Summary

Successfully refactored the logging system to implement the correct three-tier architecture:

1. **Sinks** (destination) - Immutable handlers that write logs
2. **RegisteredLoggers** (workers) - Real slog.Logger instances with Fanout handlers connecting multiple sinks
3. **Loggers** (interface) - Thin wrappers that forward to registered loggers via hierarchy

## Key Changes

### Architecture Shift

**Before**: Logger directly held sinks and wrote logs
**After**: Logger is thin wrapper → Registry resolves → RegisteredLogger does actual logging

```
User calls logger.Info()
  ↓
Logger looks up name in registry
  ↓
Registry returns RegisteredLogger
  ↓
RegisteredLogger writes to Fanout handler
  ↓
Fanout distributes to all sinks
```

### New Types

#### RegisteredLogger
```go
type RegisteredLogger struct {
    name   string
    logger *slog.Logger // with Fanout handler
}
```

The **actual working logger** with sinks. Created with `NewRegisteredLogger(name, sink1, sink2, ...)` and stored in registry.

#### Logger (refactored)
```go
type Logger struct {
    name     string
    registry *Registry
    attrs    []any     // pre-configured context
}
```

**Thin wrapper** for user interaction. Created with `registry.Get(name)`. Does NOT hold sinks; forwards to registered logger.

### API Changes

#### Creating Loggers

**Before**:
```go
logger := NewLogger("app", sink1, sink2)
registry.RegisterLogger(logger)
```

**After**:
```go
// Setup (one time)
regLogger := NewRegisteredLogger("app", sink1, sink2)
registry.RegisterLogger(regLogger)

// Usage (many times)
logger := registry.Get("app")
```

#### Logging with Context

**Before**: Needed to repeat attributes
```go
logger.Info(ctx, "msg1", slog.String("module", "cpu"))
logger.Info(ctx, "msg2", slog.String("module", "cpu"))
```

**After**: Pre-configure context once
```go
cpuLogger := logger.WithAttrs(slog.String("module", "cpu"))
cpuLogger.Info(ctx, "msg1")  // "module" automatically included
cpuLogger.Info(ctx, "msg2")  // "module" automatically included
```

#### Multiple Sinks per Logger

**Before**: Limited to multiple separate calls
**After**: Fanout handler automatically distributes to all sinks
```go
regLogger := NewRegisteredLogger("app", sink1, sink2, sink3)
// One call to Info → goes to all 3 sinks via Fanout
```

### Implementation Details

#### Fanout Handler (slog-multi)

Uses `github.com/samber/slog-multi` to create a Fanout handler that:
- Accepts multiple handlers (one per sink)
- Distributes each log record to all handlers
- Maintains independence: each sink can have different levels

```go
handlers := []slog.Handler{sink1.Handler(), sink2.Handler()}
handler := slogmulti.Fanout(handlers...)
logger := slog.New(handler)
```

#### Hierarchical Resolution

Registry still supports hierarchical lookup:
```
"app.cpu.executor" → resolved to "app.cpu" (if registered)
                  → else "app" (if registered)
                  → else nil
```

Logger automatically uses hierarchical resolution via `registry.ResolveLogger()`.

#### Context Propagation

Logger stores pre-configured attributes that get applied to every call:
```go
logger := registry.Get("app")
logger2 := logger.WithAttrs(slog.String("user", "alice"))
logger2.Info(ctx, "action")  // "user" included automatically

// Chain operations
contextLogger := logger.
    WithAttrs(slog.String("module", "cpu")).
    WithGroup("performance").
    Child("monitor")
```

### Configuration (YAML)

No changes needed - configuration still works the same:

```yaml
sinks:
  - name: console
    type: stdout
    level: info

loggers:
  - name: app
    sinks: [console]
```

Config now creates RegisteredLoggers instead of Loggers.

### Migration Path

For existing code using old API:

```go
// Old
logger := NewLogger("app", sink)

// New
regLogger := NewRegisteredLogger("app", sink)
registry.RegisterLogger(regLogger)
logger := registry.Get("app")
```

## Benefits

1. **Clean separation**: Logger = interface, RegisteredLogger = implementation
2. **Context reuse**: Pre-configure attributes once, use everywhere
3. **Fanout simplicity**: Multiple sinks handled by single handler
4. **Immutability preserved**: Both RegisteredLogger and Logger immutable
5. **Extensibility**: Logger methods return new instances
6. **Hierarchical**: Still supports parent-child resolution
7. **Thread-safe**: Immutable design eliminates locks on critical path

## Test Results

✅ All 27 tests passing
✅ Coverage maintained
✅ Build clean (no errors/warnings)

### Test Coverage

- Sink creation and immutability ✓
- RegisteredLogger creation and immutability ✓
- Logger wrapper creation ✓
- Logger.WithAttrs() ✓
- Logger.WithGroup() ✓
- Logger.Child() ✓
- Registry.RegisterSinks() ✓
- Registry.RegisterLoggers() ✓
- Registry.ResolveLogger() (6 cases) ✓
- Registry.Get() ✓
- Registry.ListLoggers() ✓
- Registry.Clear() ✓
- Actual logging ✓
- Context propagation ✓
- YAML config parsing ✓
- YAML config application ✓
- Hierarchical logging ✓
- Multiple attributes ✓
- Chained operations ✓

## Files Modified

1. **go.mod** - Added `github.com/samber/slog-multi v1.0.2`
2. **pkg/logging/logger.go** - Complete refactor:
   - New `RegisteredLogger` type (actual worker)
   - Refactored `Logger` (thin wrapper)
   - Fanout handler setup
3. **pkg/logging/sink.go** - Added `Handler()` method
4. **pkg/logging/registry.go** - Updated to work with RegisteredLogger:
   - Changed `loggers` to `registeredLoggers`
   - Added `Get()` method for thin wrapper creation
   - Updated `ResolveLogger()` return type
   - Updated `RegisterLogger()` parameter type
5. **pkg/logging/config.go** - Updated to create RegisteredLoggers
6. **pkg/logging/examples.go** - Updated examples to use new API
7. **pkg/logging/logging_test.go** - Complete rewrite with 27 tests

## Backward Compatibility

❌ Breaking changes - Users must migrate to new API

Old API no longer exists:
- `NewLogger(name, ...sinks)` → use `NewRegisteredLogger() + registry.Get()`
- `registry.GetLogger(name)` → use `registry.Get(name)`
- `registry.RegisterLogger(logger)` → pass `RegisteredLogger`, not `Logger`

## Next Steps

Documentation should be updated to reflect:
1. Three-tier architecture (Sinks → RegisteredLoggers → Loggers)
2. New API for creating and registering loggers
3. Using `registry.Get()` for standard usage
4. Context pre-configuration patterns
5. Fanout behavior with multiple sinks
