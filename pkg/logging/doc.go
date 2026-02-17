package logging

/*
Package logging provides hierarchical, structured logging using slog with
a registry-based approach for managing loggers and sinks.

# Core Concepts

Loggers:
  - Named using hierarchical dot notation: "package", "package.module", "package.module.Type"
  - IMMUTABLE: Configured with sinks at creation time, cannot be modified after
  - Can spawn child loggers that inherit attributes and sinks from the parent
  - Write messages to attached sinks
  - Can be resolved hierarchically - "package.module.Type" will match the
    closest registered logger in the hierarchy
  - New variations can be created using WithAttrs() and WithGroup() which return
    new logger instances

Sinks:
  - Destinations where logs are actually written (files, stdout, stderr, etc.)
  - IMMUTABLE: Fully configured at creation time, cannot be changed
  - Have a minimum log level (only write entries >= level)
  - Can be attached to multiple loggers
  - Support different formats (JSON, text)
  - Can be registered once and reused

Registry:
  - Central management of loggers and sinks
  - Registry itself is mutable for initial setup (RegisterLogger, RegisterSink)
  - But loggers and sinks registered in it are immutable
  - Handles hierarchical logger resolution
  - Ensures unique names for loggers and sinks
  - Provides configuration loading from YAML

# Example Usage

Basic setup:

	registry := logging.NewRegistry()

	// Create a sink (immutable)
	sink := logging.NewTextSink("console", os.Stdout, slog.LevelInfo)
	registry.RegisterSink(sink)

	// Create loggers with sinks already attached (immutable)
	runtime := logging.NewLogger("runtime", sink)
	registry.RegisterLogger(runtime)

	cpu := logging.NewLogger("runtime.cpu", sink)
	registry.RegisterLogger(cpu)

Using loggers:

	ctx := context.Background()

	// Resolve logger for hierarchical name
	logger := registry.ResolveLogger("runtime.cpu.ALU")
	logger.Info(ctx, "arithmetic operation", slog.String("op", "add"))

YAML Configuration:

	sinks:
	  - name: console
	    type: stdout
	    level: info
	    format: text
	  - name: errors
	    type: file
	    path: /var/log/errors.log
	    level: error
	    format: json

	loggers:
	  - name: runtime
	    sinks:
	      - console
	  - name: runtime.cpu
	    sinks:
	      - console
	  - name: debugger
	    sinks:
	      - console
	      - errors

	config, err := logging.NewConfigFromYAML("logging.yaml")
	registry := logging.NewRegistry()
	err = config.Apply(registry)

# Logger Resolution

Logger lookup is recursive and finds the best match:

Given registered loggers: "package", "package.module"

- "package" → "package"
- "package.module" → "package.module"
- "package.module.Type" → "package.module"
- "package.other.Type" → "package"
- "other" → nil (no match)

This allows you to have default loggers at broad levels and more specific
ones for particular subsystems.

# Thread Safety

Registry is thread-safe and can be used concurrently for registration and lookup.
Loggers and Sinks are immutable and inherently thread-safe for concurrent use.

# Performance

- Logging is only performed for entries >= sink's level
- No lock contention on the hot path for resolving loggers
- Suitable for high-throughput applications
*/

// TypeDoc marker - used to generate documentation
type TypeDoc struct{}
