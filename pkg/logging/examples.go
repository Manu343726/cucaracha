package logging

import (
	"log/slog"
	"os"
)

/*
This file contains examples of how to use the hierarchical logging system.
These are not runnable examples in the traditional sense, but they show
the patterns and workflows for different use cases.

Example 1: Basic Setup

	// Create registry
	registry := NewRegistry()

	// Create sinks
	consoleSink := NewTextSink("console", os.Stdout, slog.LevelInfo)
	fileSink := NewFileSink("file", os.Stderr, slog.LevelDebug)

	// Register sinks
	registry.RegisterSink(consoleSink)
	registry.RegisterSink(fileSink)

	// Create and register loggers
	runtime := NewLogger("runtime")
	runtime.AddSink(consoleSink)
	runtime.AddSink(fileSink)
	registry.RegisterLogger(runtime)

	cpu := NewLogger("runtime.cpu")
	cpu.AddSink(consoleSink)
	cpu.AddSink(fileSink)
	registry.RegisterLogger(cpu)

	// Use logger
	ctx := context.Background()
	cpu.Info(ctx, "CPU initialized", slog.Int("cores", 4))

Example 2: Configuration from YAML

	config, err := NewConfigFromYAML("logging.yaml")
	if err != nil {
		panic(err)
	}

	registry := NewRegistry()
	if err := config.Apply(registry); err != nil {
		panic(err)
	}

	// Resolve and use logger
	ctx := context.Background()
	logger := registry.ResolveLogger("runtime.memory.manager")
	if logger != nil {
		logger.Info(ctx, "memory allocated",
			slog.Uint64("address", 0x1000),
			slog.Uint64("size", 4096),
		)
	}

Example 3: Hierarchical Logger Resolution

	registry := NewRegistry()

	// Register loggers at different levels
	root := NewLogger("myapp")
	registry.RegisterLogger(root)

	component := NewLogger("myapp.database")
	registry.RegisterLogger(component)

	// When resolving "myapp.database.connection", it will find "myapp.database"
	logger := registry.ResolveLogger("myapp.database.connection")
	// logger.Name() == "myapp.database"

	// When resolving "myapp.cache", it will find "myapp"
	logger = registry.ResolveLogger("myapp.cache")
	// logger.Name() == "myapp"

Example 4: Child Loggers with Inherited Context

	ctx := context.Background()
	parentLogger := NewLogger("service")

	// Parent has attributes
	parentLogger = parentLogger.WithAttrs(
		slog.String("service", "user-api"),
		slog.String("version", "1.0"),
	)

	// Create child logger - inherits attributes from parent
	childLogger := parentLogger.Child("handler")
	childLogger.Info(ctx, "handling request")
	// Will log with both parent and child attributes:
	// "service": "user-api", "version": "1.0"

Example 5: Rotating Sinks by Level

	registry := NewRegistry()

	// Console for normal output
	consoleSink := NewTextSink("console", os.Stdout, slog.LevelInfo)
	registry.RegisterSink(consoleSink)

	// File for errors only
	errorFile, _ := os.OpenFile("errors.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	errorSink := NewFileSink("errors", errorFile, slog.LevelError)
	registry.RegisterSink(errorSink)

	// Logger writes to both, but errorSink filters to only errors
	logger := NewLogger("app")
	logger.AddSink(consoleSink)
	logger.AddSink(errorSink)
	registry.RegisterLogger(logger)

	ctx := context.Background()
	logger.Info(ctx, "info message")    // Goes to console only
	logger.Error(ctx, "error message")  // Goes to both console and error file

Example 6: Runtime Logger Modification

	registry := NewRegistry()

	logger := NewLogger("runtime")
	registry.RegisterLogger(logger)

	// Dynamically register sink
	sink := NewTextSink("dynamic", os.Stdout, slog.LevelDebug)
	registry.RegisterSink(sink)

	// Attach to existing logger
	registry.AttachSinkToLogger("runtime", "dynamic")

	// Detach from logger
	registry.DetachSinkFromLogger("runtime", "dynamic")

	// Change sink level
	registry.SetSinkLevel("dynamic", slog.LevelWarn)

Example 7: Cucaracha Project Structure

YAML configuration for the Cucaracha project:

	sinks:
	  - name: console
	    type: stdout
	    format: text
	    level: info

	  - name: debug_file
	    type: file
	    path: /tmp/cucaracha_debug.log
	    format: json
	    level: debug

	  - name: error_file
	    type: file
	    path: /tmp/cucaracha_errors.log
	    format: text
	    level: error

	loggers:
	  - name: runtime
	    sinks: [console, debug_file]

	  - name: runtime.cpu
	    sinks: [console, debug_file]

	  - name: runtime.memory
	    sinks: [console, debug_file]

	  - name: debugger
	    sinks: [console, debug_file, error_file]

	  - name: llvm
	    sinks: [debug_file]

	  - name: ui
	    sinks: [console]

Usage in cucaracha:

	// In main.go
	config, _ := NewConfigFromYAML("logging.yaml")
	registry := NewRegistry()
	config.Apply(registry)

	// In pkg/runtime/runner.go
	logger := registry.ResolveLogger("runtime.cpu.executor")
	ctx := context.Background()
	logger.Info(ctx, "executing instruction", slog.String("opcode", "ADD"))

	// In pkg/debugger/api.go
	logger := registry.ResolveLogger("debugger.breakpoint")
	logger.Debug(ctx, "breakpoint hit", slog.Uint64("address", 0x1000))

	// In pkg/llvm/compiler.go
	logger := registry.ResolveLogger("llvm.compilation")
	logger.Warn(ctx, "deprecated instruction", slog.String("instr", "movw"))

Example 8: Testing with Loggers

	func TestCPUExecution(t *testing.T) {
		// Create test logger
		buf := &bytes.Buffer{}
		sink := NewFileSink("test", buf, slog.LevelDebug)

		logger := NewLogger("test.cpu")
		logger.AddSink(sink)

		// Execute code that uses logger
		ctx := context.Background()
		logger.Info(ctx, "test start")

		// Verify logs
		output := buf.String()
		if !strings.Contains(output, "test start") {
			t.Fatal("expected log not found")
		}
	}

Example 9: Global Registry Pattern

	var globalRegistry *Registry

	func init() {
		globalRegistry = NewRegistry()
		config, _ := NewConfigFromYAML("logging.yaml")
		globalRegistry.Apply(config)
	}

	func GetLogger(hierarchicalName string) *Logger {
		return globalRegistry.ResolveLogger(hierarchicalName)
	}

	// In modules:
	logger := GetLogger("runtime.cpu.executor")
	logger.Info(ctx, "operation executed")

Example 10: Contextual Logging with Request Tracing

	// Request enters the system
	func HandleRequest(registry *Registry, reqID string, command string) {
		ctx := context.Background()

		logger := registry.ResolveLogger("runtime.executor")
		requestLogger := logger.WithAttrs(slog.String("request_id", reqID))

		requestLogger.Info(ctx, "request started")

		// Pass to sub-operations
		executeCommand(requestLogger, command)

		requestLogger.Info(ctx, "request completed")
	}

	func executeCommand(logger *Logger, command string) {
		ctx := context.Background()

		// Sub-operations inherit request_id
		opLogger := logger.WithAttrs(slog.String("operation", command))
		opLogger.Debug(ctx, "executing")

		// All logs from this operation will have both request_id and operation
	}

Key Patterns:

1. Registry Creation: Always create a registry first and register sinks/loggers

2. Configuration: Use YAML for complex setups with multiple sinks/loggers

3. Logger Resolution: Prefer ResolveLogger() over GetLogger() for hierarchical names

4. Child Loggers: Use Child() when you need to extend the hierarchy

5. Attributes: Use WithAttrs() to add context that persists across logs

6. Testing: Create test sinks with buffers to verify logging behavior

7. Performance: Filter at sink level to avoid expensive string formatting

8. Global Access: Consider a GetLogger() function for convenience, but prefer
   passing loggers as parameters for testability
*/

// ExampleBasicSetup shows how to set up basic logging
func ExampleBasicSetup() {
	registry := NewRegistry()

	// Create sink
	sink := NewTextSink("console", os.Stdout, slog.LevelInfo)
	registry.RegisterSink(sink)

	// Create and register a registered logger with sinks attached
	registeredLogger := NewRegisteredLogger("myapp", sink)
	registry.RegisterLogger(registeredLogger)

	// Get a thin wrapper logger for use
	logger := registry.Get("myapp")

	// Use logger
	logger.Info("application started", slog.String("module", "main"))
}

// ExampleYAMLConfiguration shows how to load from YAML
func ExampleYAMLConfiguration() {
	yaml := `
sinks:
  - name: console
    type: stdout
    level: info

loggers:
  - name: runtime
    sinks: [console]
`

	config, _ := NewConfigFromString(yaml)
	registry := NewRegistry()
	_ = config.Apply(registry)

	// Logger is now ready to use - get a thin wrapper
	logger := registry.Get("runtime")
	logger.Info("ready", slog.String("module", "main"))
}

// ExampleHierarchicalResolution shows logger resolution
func ExampleHierarchicalResolution() {
	registry := NewRegistry()

	// Register sinks first
	sink := NewTextSink("console", os.Stdout, slog.LevelInfo)
	registry.RegisterSink(sink)

	// Register registered loggers with sinks attached
	root := NewRegisteredLogger("app", sink)
	component := NewRegisteredLogger("app.database", sink)
	registry.RegisterLogger(root)
	registry.RegisterLogger(component)

	// Resolution examples:
	// "app" matches "app"
	// "app.database" matches "app.database"
	// "app.database.connection" matches "app.database"
	// "app.cache" matches "app"

	resolved := registry.ResolveLogger("app.database.connection")
	if resolved != nil {
		// resolved.Name() == "app.database"
	}
}
