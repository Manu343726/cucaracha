# Package Logger Integration

## Overview

Each Cucaracha package now has a `logger.go` file with a `log()` function that returns the appropriate logger for that package. The logging system uses a global default registry initialized on first use.

## Architecture

```
Global Registry (initialized once)
├─ Sink: console (default)
├─ RegisteredLogger: cucaracha (root)
└─ Created on-demand via hierarchical resolution

Package Usage:
pkg/runtime/logger.go → log() → logging.Get("cucaracha.runtime")
pkg/hw/cpu/logger.go → log() → logging.Get("cucaracha.hw.cpu")
```

## Files Created

### Main Packages

Each main package has a `logger.go`:

- `pkg/runtime/logger.go` - "cucaracha.runtime"
- `pkg/hw/logger.go` - "cucaracha.hw"
- `pkg/debugger/logger.go` - "cucaracha.debugger"
- `pkg/interpreter/logger.go` - "cucaracha.interpreter"
- `pkg/llvm/logger.go` - "cucaracha.llvm"
- `pkg/system/logger.go` - "cucaracha.system"
- `pkg/ui/logger.go` - "cucaracha.ui"

### Sub-Packages

Each sub-package has a `logger.go`:

- `pkg/hw/cpu/logger.go` - "cucaracha.hw.cpu"
- `pkg/hw/memory/logger.go` - "cucaracha.hw.memory"
- `pkg/hw/peripheral/logger.go` - "cucaracha.hw.peripheral"
- `pkg/hw/component/logger.go` - "cucaracha.hw.component"
- `pkg/hw/components/logger.go` - "cucaracha.hw.components"
- `pkg/hw/peripherals/logger.go` - "cucaracha.hw.peripherals"
- `pkg/runtime/program/logger.go` - "cucaracha.runtime.program"
- `pkg/debugger/core/logger.go` - "cucaracha.debugger.core"
- `pkg/llvm/templates/logger.go` - "cucaracha.llvm.templates"
- `pkg/ui/tui/logger.go` - "cucaracha.ui.tui"

## Global Registry

### Initialization

The global registry is initialized on first use with:

```go
// Created once, on first call to DefaultRegistry()
registry := logging.DefaultRegistry()

// Automatically has:
// - console sink (stdout, text format, debug level)
// - cucaracha root logger
```

### API

```go
// Get logger for a package from global registry
logger := logging.Get("cucaracha.runtime.cpu")

// Initialize with custom YAML configuration
logging.InitializeWithConfig("config/logging.yaml")

// Or with YAML string
logging.InitializeWithConfigString(yamlString)

// Get the registry directly
registry := logging.DefaultRegistry()
```

## Usage Pattern

### In Each Package

Each package's `logger.go` looks like:

```go
package runtime

import (
    "github.com/Manu343726/cucaracha/pkg/logging"
)

// log returns the logger for the runtime package.
func log() *logging.Logger {
    return logging.Get("cucaracha.runtime")
}
```

### In Package Code

```go
package runtime

func (r *Runner) Start() {
    ctx := context.Background()
    logger := log()
    logger.Info(ctx, "runner started", slog.Int("cores", r.cpus))
}

func (r *Runner) WithDebug() {
    logger := log().WithAttrs(slog.String("mode", "debug"))
    logger.Info(context.Background(), "debug mode enabled")
}
```

### Sub-Package Code

Sub-packages use the same pattern but with hierarchical names:

```go
// pkg/hw/cpu/executor.go
package cpu

func (e *Executor) Execute(instr *Instruction) {
    logger := log()  // "cucaracha.hw.cpu"
    logger.Info(ctx, "executing", slog.String("instr", instr.Name()))
}
```

## Hierarchical Resolution

When a logger is not explicitly registered, it resolves to the nearest parent:

```go
// If only "cucaracha" and "cucaracha.hw" are registered:

logging.Get("cucaracha.hw.cpu.alu.register").Info(ctx, "msg")
// Resolves to: cucaracha.hw
// Logs to: its sinks

logging.Get("cucaracha.interpreter.ast").Info(ctx, "msg")
// Resolves to: cucaracha
// Logs to: its sinks (console by default)
```

## Configuration

### Default Setup

The system starts with:
- **Sink**: console (stdout, text format, debug level)
- **Logger**: cucaracha (root, uses console sink)

All other loggers are created on-demand and resolve hierarchically.

### Custom Configuration

Create `config/logging.yaml`:

```yaml
sinks:
  - name: console
    type: stdout
    level: debug
    format: text
  
  - name: file
    type: file
    path: /var/log/cucaracha.log
    level: info
    format: json

loggers:
  - name: cucaracha
    sinks: [console, file]
  
  - name: cucaracha.hw
    sinks: [console, file]
  
  - name: cucaracha.hw.cpu
    sinks: [file]  # CPU logs only to file
  
  - name: cucaracha.debugger
    sinks: [console]
```

Then initialize in main:

```go
func init() {
    logging.InitializeWithConfig("config/logging.yaml")
}
```

## Context Propagation

Each package can create context-aware loggers:

```go
// pkg/runtime/runner.go
type Runner struct {
    logger *logging.Logger
}

func NewRunner(name string) *Runner {
    return &Runner{
        logger: log().WithAttrs(
            slog.String("runner", name),
            slog.Int("id", globalRunnerID++),
        ),
    }
}

func (r *Runner) Info(ctx context.Context, msg string, attrs ...any) {
    r.logger.Info(ctx, msg, attrs...)
    // Includes runner and id automatically
}
```

## Testing

Each package's logger is tested:

```go
func TestRuntime(t *testing.T) {
    logger := log()
    if logger == nil {
        t.Error("logger should not be nil")
    }
    
    // Logger forwards to global registry
    logger.Info(context.Background(), "test message")
}
```

## Best Practices

1. **Use `log()` in package code**: Don't call `logging.Get()` directly in business logic
2. **Pre-configure context**: Use `WithAttrs()` for package-level context
3. **One logger per request**: Use `WithAttrs()` to add request-specific context
4. **Hierarchical naming**: Follow `cucaracha.package.subpackage` pattern
5. **Initialize once**: Call `logging.InitializeWithConfig()` in main/init
6. **No circular imports**: All packages use `logging.Get()`, not parent imports

## Summary

- ✅ Each package has `logger.go` with `log()` function
- ✅ Global registry initialized on first use
- ✅ Hierarchical resolution for all loggers
- ✅ No circular imports between packages
- ✅ Automatic parent logger fallback
- ✅ Simple API: just call `log()` in package code
- ✅ Full integration with slog and Fanout
- ✅ Production-ready configuration via YAML
