# Integration Guide for Cucaracha Project

This guide shows how to integrate the hierarchical logging system into the Cucaracha project.

## Step 1: Create Logging Configuration

Create `config/logging.yaml` in the project root:

```yaml
sinks:
  console:
    name: console
    type: stdout
    level: info
    format: text

  debug_file:
    name: debug_file
    type: file
    path: /tmp/cucaracha_debug.log
    level: debug
    format: json

  error_file:
    name: error_file
    type: file
    path: /tmp/cucaracha_errors.log
    level: error
    format: text

loggers:
  # Runtime execution
  runtime:
    name: runtime
    sinks: [console, debug_file]

  runtime.cpu:
    name: runtime.cpu
    sinks: [console, debug_file]

  runtime.memory:
    name: runtime.memory
    sinks: [console, debug_file]

  runtime.peripheral:
    name: runtime.peripheral
    sinks: [console, debug_file]

  # Debugger
  debugger:
    name: debugger
    sinks: [console, debug_file, error_file]

  debugger.breakpoint:
    name: debugger.breakpoint
    sinks: [console, debug_file]

  debugger.watchpoint:
    name: debugger.watchpoint
    sinks: [console, debug_file]

  # LLVM Integration
  llvm:
    name: llvm
    sinks: [debug_file]

  llvm.cmake:
    name: llvm.cmake
    sinks: [debug_file]

  llvm.binary_parser:
    name: llvm.binary_parser
    sinks: [debug_file]

  # UI
  ui:
    name: ui
    sinks: [console]

  ui.tui:
    name: ui.tui
    sinks: [console]

  # Interpreter
  interpreter:
    name: interpreter
    sinks: [console, debug_file]
```

## Step 2: Initialize Global Logger Registry

Create `pkg/logging/init.go`:

```go
package logging

import (
	"log"
	"os"
	"path/filepath"
)

var globalRegistry *Registry

// Initialize sets up the global logging registry from config file
func Initialize(configPath string) error {
	// If no path provided, try default locations
	if configPath == "" {
		configPath = "config/logging.yaml"
		if _, err := os.Stat(configPath); err != nil {
			// Try relative to executable
			configPath = filepath.Join(filepath.Dir(os.Args[0]), "config", "logging.yaml")
		}
	}

	config, err := NewConfigFromYAML(configPath)
	if err != nil {
		return err
	}

	globalRegistry = NewRegistry()
	if err := config.Apply(globalRegistry); err != nil {
		return err
	}

	return nil
}

// GetRegistry returns the global logger registry
func GetRegistry() *Registry {
	if globalRegistry == nil {
		// Fallback: create minimal registry if not initialized
		log.Println("warning: logger registry not initialized, using minimal setup")
		globalRegistry = NewRegistry()
	}
	return globalRegistry
}

// Get retrieves a logger by hierarchical name
func Get(name string) *Logger {
	registry := GetRegistry()
	logger := registry.ResolveLogger(name)
	if logger == nil {
		// Return empty logger as fallback
		return NewLogger(name)
	}
	return logger
}
```

Add to `cmd/root.go`:

```go
package cmd

import (
	"github.com/Manu343726/cucaracha/pkg/logging"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cucaracha",
	Short: "Cucaracha emulator",
	Long:  `A complete emulator for the Cucaracha architecture`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize logging before running any command
		return logging.Initialize("")
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	// Add global flags if needed
	rootCmd.PersistentFlags().String(
		"log-config",
		"",
		"Path to logging configuration file",
	)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
```

## Step 3: Update Module Code

### pkg/runtime/runner.go

```go
package runtime

import (
	"context"
	"log/slog"
	"github.com/Manu343726/cucaracha/pkg/logging"
)

type Runner struct {
	// ... existing fields
	logger *logging.Logger
}

func NewRunner() *Runner {
	return &Runner{
		// ... initialize fields
		logger: logging.Get("runtime"),
	}
}

func (r *Runner) Start(ctx context.Context) error {
	ctx = logging.WithAttrs(ctx, slog.String("runner_id", r.id))

	r.logger.Info(ctx, "runner starting")

	if err := r.initialize(ctx); err != nil {
		r.logger.Error(ctx, "initialization failed", slog.String("error", err.Error()))
		return err
	}

	r.logger.Info(ctx, "runner ready")
	return nil
}

func (r *Runner) Execute(ctx context.Context) error {
	r.logger.Info(ctx, "execution started")

	// Use child logger for sub-operations
	cpuLogger := r.logger.Child("cpu")
	return r.executeCPU(ctx, cpuLogger)
}

func (r *Runner) executeCPU(ctx context.Context, logger *logging.Logger) error {
	logger.Debug(ctx, "executing CPU cycle")
	// ... CPU execution code
	return nil
}
```

### pkg/hw/cpu/cpu.go

```go
package cpu

import (
	"context"
	"log/slog"
	"github.com/Manu343726/cucaracha/pkg/logging"
)

type CPU struct {
	// ... existing fields
	logger *logging.Logger
}

func NewCPU() *CPU {
	return &CPU{
		// ... initialize fields
		logger: logging.Get("runtime.cpu"),
	}
}

func (c *CPU) Step(ctx context.Context) error {
	ctx = logging.WithAttrs(ctx, slog.Uint64("pc", c.registers.PC))

	c.logger.Debug(ctx, "CPU step")
	// ... step implementation
	return nil
}

func (c *CPU) Execute(ctx context.Context, instruction uint32) error {
	execLogger := c.logger.WithAttrs(slog.Uint64("instruction", uint64(instruction)))

	execLogger.Debug(ctx, "executing instruction")
	// ... execution code
	return nil
}
```

### pkg/debugger/api.go

```go
package debugger

import (
	"context"
	"log/slog"
	"github.com/Manu343726/cucaracha/pkg/logging"
)

type Debugger struct {
	// ... existing fields
	logger *logging.Logger
}

func NewDebugger() *Debugger {
	return &Debugger{
		// ... initialize fields
		logger: logging.Get("debugger"),
	}
}

func (d *Debugger) AddBreakpoint(ctx context.Context, addr uint32) error {
	bpLogger := d.logger.Child("breakpoint")
	bpLogger.Info(ctx, "breakpoint added", slog.Uint64("address", uint64(addr)))
	// ... add breakpoint
	return nil
}

func (d *Debugger) OnBreakpoint(ctx context.Context, addr uint32) {
	bpLogger := d.logger.Child("breakpoint").WithAttrs(slog.Uint64("address", uint64(addr)))
	bpLogger.Info(ctx, "breakpoint hit")
	// ... handle breakpoint
}
```

### pkg/llvm/cmake.go

```go
package llvm

import (
	"context"
	"log/slog"
	"github.com/Manu343726/cucaracha/pkg/logging"
)

func BuildClang(ctx context.Context) error {
	logger := logging.Get("llvm.cmake")

	logger.Info(ctx, "starting clang build")

	// Replace old fmt.Print calls with logger
	logger.Info(ctx, "using CMake preset", slog.String("preset", "docker-gcc"))

	logger.Debug(ctx, "running CMake configure")
	// ... CMake execution

	if err := runBuild(ctx); err != nil {
		logger.Error(ctx, "build failed", slog.String("error", err.Error()))
		return err
	}

	logger.Info(ctx, "build completed successfully")
	return nil
}
```

## Step 4: Replace Old Logging

Replace old logging patterns:

**Old pattern:**
```go
fmt.Printf("Warning: Failed to build with presets: %v\n", err)
fmt.Println("Building clang...")
// fmt.Printf("DEBUG: loOffset=0x%X\n", loOffset)
```

**New pattern:**
```go
logger.Warn(ctx, "failed to build with presets", slog.String("error", err.Error()))
logger.Info(ctx, "building clang")
logger.Debug(ctx, "offset calculation", slog.Uint64("lo_offset", uint64(loOffset)))
```

## Step 5: Usage Examples

### In Commands

```go
// In cmd/root.go
func Execute() {
	ctx := context.Background()

	// Add request ID for tracing
	ctx = logging.WithAttrs(ctx, slog.String("session_id", generateID()))

	cmd.Execute()
}
```

### In Tests

```go
func TestCPUExecution(t *testing.T) {
	// Create test logger
	buf := &bytes.Buffer{}
	sink := logging.NewFileSink("test", buf, slog.LevelDebug)
	logger := logging.NewLogger("test.cpu", sink)

	cpu := NewCPU()
	cpu.logger = logger

	ctx := context.Background()
	err := cpu.Step(ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify logs
	output := buf.String()
	if !strings.Contains(output, "CPU step") {
		t.Error("expected log not found")
	}
}
```

## Configuration Profiles

Create different configurations for different use cases:

### Development (config/logging.dev.yaml)

```yaml
sinks:
  - name: console
    type: stdout
    level: debug
    format: text
```

### Production (config/logging.prod.yaml)

```yaml
sinks:
  - name: file
    type: file
    path: /var/log/cucaracha/app.log
    level: warn
    format: json
```

Load based on environment:

```go
func Initialize(configPath string) error {
	if configPath == "" {
		env := os.Getenv("CUCARACHA_ENV")
		if env == "" {
			env = "dev"
		}
		configPath = fmt.Sprintf("config/logging.%s.yaml", env)
	}
	// ... rest of initialization
}
```

## Benefits of This Integration

1. **Structured Logging** - All logs are JSON/structured
2. **Context Propagation** - Request IDs flow through the system
3. **Hierarchical Organization** - Logs grouped by module
4. **Performance** - No string formatting overhead for filtered levels
5. **Testing** - Easy to capture and verify logs
6. **Configuration** - Change logging without code changes
7. **Thread-Safe** - Safe for concurrent use
8. **Immutable** - No accidental logger state changes

## Migration Timeline

1. **Phase 1** - Initialize logging in main, get it working
2. **Phase 2** - Update high-level modules (runner, debugger)
3. **Phase 3** - Update CPU and memory operations
4. **Phase 4** - Update LLVM integration
5. **Phase 5** - Remove old fmt.Printf calls
6. **Phase 6** - Enable debug logging in configuration

## Troubleshooting

### "Logger registry not initialized"

Ensure `logging.Initialize()` is called before using any loggers:

```go
// In PersistentPreRunE or init()
if err := logging.Initialize(""); err != nil {
	log.Fatal(err)
}
```

### Logs not appearing

Check:
1. Logger level >= message level
2. Sink level >= message level
3. Logger is registered in configuration
4. Correct sink is attached to logger

### Performance issues

If logging is slow:
1. Check sink is using appropriate level
2. Avoid logging in tight loops
3. Use Debug level for verbose logging and disable in production
