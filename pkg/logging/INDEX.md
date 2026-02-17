# Hierarchical Logging System - Package Contents

Complete hierarchical logging system for Cucaracha project with immutable loggers and sinks.

## Package Location

`/workspaces/cucaracha/pkg/logging/`

## Files Overview

### Implementation (5 files)

| File | Lines | Purpose |
|------|-------|---------|
| **sink.go** | ~75 | Immutable sink implementation |
| **logger.go** | ~150 | Immutable logger with context support |
| **registry.go** | ~130 | Logger/sink registry with hierarchical resolution |
| **config.go** | ~220 | YAML configuration loading and application |
| **logging_test.go** | ~475 | Comprehensive unit tests (20+ test cases) |

### Documentation (5 files)

| File | Purpose |
|------|---------|
| **README.md** | Complete user guide with quick start and API reference |
| **IMMUTABLE_DESIGN.md** | Design rationale, patterns, and best practices |
| **INTEGRATION.md** | Step-by-step integration guide for Cucaracha |
| **doc.go** | Package documentation and overview |
| **examples.go** | Code examples for different use cases |

### Reference (2 files)

| File | Purpose |
|------|---------|
| **SUMMARY.md** | Executive summary of the implementation |
| **INDEX.md** | This file |

**Total Size**: ~92 KB  
**Total Lines**: ~1800 lines of code and documentation

## Quick Navigation

### I want to...

**Get started quickly**
→ Read [README.md](README.md) - Quick Start section

**Understand the design**
→ Read [IMMUTABLE_DESIGN.md](IMMUTABLE_DESIGN.md)

**Integrate into Cucaracha**
→ Read [INTEGRATION.md](INTEGRATION.md)

**See code examples**
→ Look at [examples.go](examples.go) or README.md

**Understand how it works**
→ Read [doc.go](doc.go)

**Run tests**
```bash
cd /workspaces/cucaracha
go test -v ./pkg/logging/...
```

## Key Features

✅ Hierarchical logger resolution  
✅ Immutable loggers and sinks  
✅ YAML configuration support  
✅ Multiple sinks per logger  
✅ Thread-safe design  
✅ Built on Go's standard slog  
✅ Extensible logging contexts  
✅ Production-ready  

## Core Concepts

**Logger**: Named entity (e.g., "runtime.cpu.executor") that writes to sinks  
**Sink**: Destination (file, stdout) with a minimum log level  
**Registry**: Manages loggers and sinks, enables hierarchical resolution  

```
Application Code
      ↓
    Logger (immutable)
      ↓
   Sinks (immutable)
      ↓
 File/Stdout/Stderr
```

## Test Results

```
=== RUN   TestSink_Creation
--- PASS: TestSink_Creation (0.00s)
=== RUN   TestLogger_Creation
--- PASS: TestLogger_Creation (0.00s)
... (18 more tests)
PASS
ok      github.com/Manu343726/cucaracha/pkg/logging     0.002s
```

All 20+ tests passing ✓

## Hierarchical Resolution

Loggers are resolved by best match:

```go
// Registered: "runtime", "runtime.cpu"

"runtime"              → "runtime"
"runtime.cpu"          → "runtime.cpu"  
"runtime.cpu.executor" → "runtime.cpu"
"runtime.memory"       → "runtime"
"other"                → nil (no match)
```

## Immutable Design

**Sinks**: Cannot be modified after creation
- Fixed name
- Fixed handler
- Fixed log level

**Loggers**: Core is immutable, context is extensible
- Immutable: name, sinks
- Extensible: WithAttrs(), WithGroup(), Child()
- All variations return new instances

**Benefits**: Thread-safe, predictable, no accidental state changes

## Configuration Example

```yaml
sinks:
  - name: console
    type: stdout
    level: info
    format: text

loggers:
  - name: runtime
    sinks: [console]
  - name: runtime.cpu
    sinks: [console]
```

## Usage Example

```go
// Setup
config, _ := logging.NewConfigFromYAML("logging.yaml")
registry := logging.NewRegistry()
config.Apply(registry)

// Usage
logger := registry.ResolveLogger("runtime.cpu.executor")
ctx := context.Background()
logger.Info(ctx, "CPU initialized", slog.Int("cores", 4))
```

## Integration Checklist

- [ ] Read README.md for overview
- [ ] Read IMMUTABLE_DESIGN.md for patterns  
- [ ] Follow INTEGRATION.md steps
- [ ] Create config/logging.yaml
- [ ] Initialize in cmd/root.go
- [ ] Update modules to use loggers
- [ ] Run tests: `go test ./pkg/logging/...`
- [ ] Verify logs output correctly
- [ ] Remove old fmt.Print calls

## Documentation Index

### For Users
1. [README.md](README.md) - Complete user guide
2. [examples.go](examples.go) - Working code examples

### For Developers
1. [IMMUTABLE_DESIGN.md](IMMUTABLE_DESIGN.md) - Design patterns
2. [doc.go](doc.go) - Package documentation
3. [logging_test.go](logging_test.go) - Test examples

### For Integration
1. [INTEGRATION.md](INTEGRATION.md) - Step-by-step guide
2. [SUMMARY.md](SUMMARY.md) - What was created

### API Reference
- Sink API: [sink.go](sink.go)
- Logger API: [logger.go](logger.go)
- Registry API: [registry.go](registry.go)
- Config API: [config.go](config.go)

## Module Dependencies

```
logging/
├── Uses: log/slog (Go standard library)
├── Uses: strings, fmt, sync (Go standard library)
├── Uses: gopkg.in/yaml.v3 (YAML parsing)
└── No external dependencies
```

## Compliance

✓ Follows Go best practices  
✓ Uses standard library where possible  
✓ Thread-safe design  
✓ Immutability principles  
✓ Comprehensive tests  
✓ Clear documentation  
✓ Integration examples  

## Performance

- Logging is O(1) for number of sinks
- Logger resolution is O(n) where n = depth of hierarchy
- Immutable design = no locks for reads
- Suitable for high-throughput applications

## Support

For issues or questions:

1. Check [README.md](README.md#troubleshooting) troubleshooting section
2. Review [examples.go](examples.go) for patterns
3. Look at [logging_test.go](logging_test.go) for test examples
4. Read [IMMUTABLE_DESIGN.md](IMMUTABLE_DESIGN.md) for design info

## License

Same as Cucaracha project
