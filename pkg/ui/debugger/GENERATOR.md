# Debugger Command Generator

This document explains the code generation system for debugger commands in the cucaracha project.

## Overview

The debugger command generator is a Go code generation tool that automatically generates command infrastructure from the `DebuggerCommands` interface definition in [api.go](api.go). This ensures consistency between command definitions and their implementations, eliminating manual synchronization.

**Single source of truth:** The `DebuggerCommands` interface is the canonical definition of all debugger commands.

## Running the Generator

Generate all command-related code:

```bash
cd /workspaces/cucaracha/pkg/ui/debugger
go generate ./...
```

This automatically invokes the generator as specified in the `//go:generate` directive in [command.go](command.go).

## Generator Architecture

The generator is split into two files for clean separation of concerns:

### Entry Point: generator.go

- **Location:** [generator.go](generator.go)
- **Build tag:** `//go:build ignore` (excluded from normal builds)
- **Purpose:** Main entry point for code generation
- **Content:**
  - `main()` function that parses command-line flags
  - Calls exported functions from the debugger package
  - Writes generated code to output files
- **Flags:**
  - `-out`: Output file for Execute() method (default: `zz_execute_generated.go`)
  - `-api`: API file to parse (default: `api.go`)
  - `-docs-out`: Output file for documentation schema (default: `zz_commands_documentation_schema.go`)
  - `-enum-out`: Output file for enum constants (default: `zz_commands_enum_generated.go`)
  - `-structs-out`: Output file for structs (default: `zz_commands_structs_generated.go`)

### Implementation: generator_impl.go

- **Location:** [generator_impl.go](generator_impl.go)
- **Package:** `debugger` (part of debugger package)
- **Purpose:** All generation logic and AST parsing
- **Exported Functions:**
  - `ParseDebuggerInterface(apiPath string) ([]CommandInfo, error)` - Extract interface definitions
  - `GenerateExecuteMethod(commands []CommandInfo) ([]byte, error)` - Generate dispatch logic
  - `GenerateEnumConstants(commands []CommandInfo) (string, error)` - Generate enums
  - `GenerateCommandStructs(commands []CommandInfo) (string, error)` - Generate structs
  - `GenerateDocsSchema(commands []CommandInfo) (string, error)` - Generate documentation
- **Private Functions:**
  - `typeToString()` - Convert AST types to strings
  - `baseTypeName()` - Extract base type from pointers
  - `shouldIncludeInExecute()` - Filter methods from generation
  - `methodNameToCommandName()` - PascalCase to camelCase conversion

### Generation Trigger

- **Directive:** Line 1 of [command.go](command.go):
  ```go
  //go:generate go run generator.go -out zz_execute_generated.go -api api.go -docs-out zz_commands_documentation_schema.go -enum-out zz_commands_enum_generated.go -structs-out zz_commands_structs_generated.go
  ```
- **When it runs:** Every time `go generate ./...` is executed in the debugger package

## How It Works

### 1. Interface Parsing

The generator parses the `DebuggerCommands` interface from [api.go](api.go) using Go's AST (Abstract Syntax Tree) package:

```go
// parseDebuggerInterface() extracts:
// - Method names (e.g., Step, Continue, Disasm)
// - Parameter types (e.g., *StepArgs)
// - Return types (e.g., *ExecutionResult)
// - Documentation comments
```

**Naming conventions enforced:**
- Methods: PascalCase (e.g., `Step`, `RemoveBreakpoint`)
- Parameter types: `*<MethodName>Args` (e.g., `*StepArgs`)
- Return types: `*<Type>Result` (e.g., `*ExecutionResult`)
- Command names (JSON): camelCase (e.g., `step`, `removeBreakpoint`)

### 2. Command Info Extraction

For each interface method, the generator creates a `commandInfo` struct containing:

| Field | Description | Example |
|-------|-------------|---------|
| `MethodName` | Interface method name | `Step` |
| `CommandID` | Enum constant name | `DebuggerCommandStep` |
| `CommandName` | JSON identifier | `step` |
| `Comments` | Documentation | `"Executes a single execution step"` |
| `ArgsType` | Parameter type | `*StepArgs` |
| `ArgsFieldName` | Command struct field reference | `cmd.StepArgs` |
| `ResultType` | Return type | `*ExecutionResult` |
| `ResultFieldName` | Result struct field name | `StepResult` |

### 3. Code Generation

The generator produces four output files:

#### A. `zz_commands_enum_generated.go`

**Purpose:** Define the command identifier enum and conversion functions

**Generated:**

1. **`DebuggerCommandId` type definition**
   ```go
   type DebuggerCommandId int
   ```

2. **Enum constants (iota-based)**
   ```go
   const (
       DebuggerCommandStep DebuggerCommandId = iota
       DebuggerCommandContinue
       DebuggerCommandRun
       // ... one for each method
   )
   ```
   - Constants are generated in the order methods appear in the interface
   - Each gets a unique numeric value via iota
   - Comments from interface methods are preserved

3. **`String()` method**
   ```go
   func (c DebuggerCommandId) String() string {
       switch c {
       case DebuggerCommandStep:
           return "step"
       // ...
       }
   }
   ```
   - Maps enum values to camelCase command names for JSON

4. **`DebuggerCommandIdFromString()` function**
   ```go
   func DebuggerCommandIdFromString(s string) (DebuggerCommandId, error) {
       switch s {
       case "step":
           return DebuggerCommandStep, nil
       // ...
       }
   }
   ```
   - Parses JSON command names back to enum values
   - Returns error for unknown commands

#### B. `zz_commands_structs_generated.go`

**Purpose:** Define command and result message types

**Generated:**

1. **`DebuggerCommand` struct**
   ```go
   type DebuggerCommand struct {
       Id      uint64            `json:"id"`      // Unique ID
       Command DebuggerCommandId `json:"command"` // Command type
       StepArgs           *StepArgs           `json:"stepargs"`           // Conditional
       DisasmArgs         *DisasmArgs         `json:"disasmargs"`         // Conditional
       // ... one optional field per method with arguments
   }
   ```
   - `Id`: Unique identifier for this command invocation
   - `Command`: Enum identifying which command this is
   - Optional argument fields: Only methods with parameters generate fields
   - Field names: `<MethodName>Args` (e.g., `StepArgs`, `DisasmArgs`)
   - JSON tags: lowercase (e.g., `stepargs`, `disasmargs`)

2. **`DebuggerCommandResult` struct**
   ```go
   type DebuggerCommandResult struct {
       Id      uint64            `json:"id"`      // Matches command ID
       Command DebuggerCommandId `json:"command"` // Echo of command type
       StepResult         *ExecutionResult    `json:"stepresult"`         // One per method
       ContinueResult     *ExecutionResult    `json:"continueresult"`
       RunResult          *ExecutionResult    `json:"runresult"`
       // ... one field per method
   }
   ```
   - `Id`: Echoes the command ID for correlation
   - `Command`: Echoes the command type for debugging
   - Result fields: Always generated, one per method
   - Field names: `<MethodName>Result` (e.g., `StepResult`, `DisasmResult`)
   - Types: Use the exact return type from the interface method

#### C. `zz_execute_generated.go`

**Purpose:** Implement command dispatch logic

**Generated:**

1. **`Execute()` method**
   ```go
   func (d *commandBasedDebuggerAdapter) Execute(cmd *DebuggerCommand) (*DebuggerCommandResult, error) {
       switch cmd.Command {
       case DebuggerCommandStep:
           return &DebuggerCommandResult{
               Id:         cmd.Id,
               Command:    cmd.Command,
               StepResult: d.debugger.Step(cmd.StepArgs),
           }, nil
       case DebuggerCommandContinue:
           return &DebuggerCommandResult{
               Id:             cmd.Id,
               Command:        cmd.Command,
               ContinueResult: d.debugger.Continue(),
           }, nil
       // ... one case per method
       default:
           return nil, fmt.Errorf("unsupported command: %s", cmd.Command)
       }
   }
   ```
   - Routes commands to the appropriate interface method
   - Wraps results in `DebuggerCommandResult`
   - Always returns result structured with correct field populated

#### D. `zz_commands_documentation_schema.go`

**Purpose:** Runtime documentation for all commands

**Generated:**

1. **`CommandsDocsSchema` variable**
   ```go
   var CommandsDocsSchema = &CommandsDocumentationSchema{
       Version: "1.0",
       Commands: map[DebuggerCommandId]CommandDocumentation{
           0: { // DebuggerCommandStep
               ID:          "0",
               Name:        "step",
               Description: "Executes a single execution step",
               Arguments:   []CommandArgumentInfo{},
               Result:      "",
               ResultFields: []CommandResultField{},
           },
           // ... one entry per method
       },
       Enums: make(map[string]EnumDocumentation),
   }
   ```
   - Uses types from [commands_documentation_service.go](commands_documentation_service.go):
     - `CommandsDocumentationSchema`: Container for all documentation
     - `CommandDocumentation`: Documentation for one command
     - `CommandArgumentInfo`: Info about command parameters
     - `CommandResultField`: Info about result fields
   - Populated with interface method documentation
   - Indexed by numeric command ID (enum value)

## Naming Conventions

The generator enforces strict naming conventions to ensure consistency:

| Item | Convention | Example |
|------|-----------|---------|
| Interface methods | PascalCase | `Step`, `RemoveBreakpoint` |
| Command enum constants | `DebuggerCommand` + MethodName | `DebuggerCommandStep` |
| Command names (JSON) | camelCase | `step`, `removeBreakpoint` |
| Argument struct types | `<MethodName>Args` | `StepArgs`, `DisasmArgs` |
| Result struct types | `<MethodName>Result` | `ExecutionResult`, `DisasmResult` |
| Command struct fields | `<MethodName>Args` | `StepArgs`, `DisasmArgs` |
| Result struct fields | `<MethodName>Result` | `StepResult`, `DisasmResult` |
| JSON field names | lowercase | `stepargs`, `stepresult` |

## Types Used from Other Files

The generator reuses existing types defined in [commands_documentation_service.go](commands_documentation_service.go):

- `CommandDocumentation`: Describes one debugger command
- `CommandArgumentInfo`: Describes command arguments
- `CommandResultField`: Describes result fields
- `EnumDocumentation`: Maps enum value names to descriptions
- `CommandsDocumentationSchema`: Container for all command documentation

The generator also references types that must exist in the debugger package:
- All `*Args` types (e.g., `StepArgs`, `DisasmArgs`)
- All `*Result` types (e.g., `ExecutionResult`, `DisasmResult`)

## Adding a New Command

To add a new debugger command, follow these steps:

### Step 1: Add to Interface

Edit [api.go](api.go) and add method to `DebuggerCommands` interface:

```go
// MyCommand does something interesting
MyCommand(args *MyCommandArgs) *MyCommandResult
```

**Requirements:**
- Method name: PascalCase
- Parameter: `*<MethodName>Args` (or no parameter if nil)
- Return: `*<Type>Result`
- Include documentation comment

### Step 2: Create Argument Type (if needed)

If your command needs parameters, define a struct (typically in same file as result type):

```go
type MyCommandArgs struct {
    Param1 string `json:"param1"`
    Param2 int    `json:"param2"`
}
```

### Step 3: Create Result Type

Define the result type the command returns:

```go
type MyCommandResult struct {
    Output string `json:"output"`
    Status int    `json:"status"`
}
```

### Step 4: Implement Method

Implement the method in your debugger implementation:

```go
func (d *myDebuggerImpl) MyCommand(args *MyCommandArgs) *MyCommandResult {
    // Implementation
    return &MyCommandResult{
        Output: "result",
        Status: 0,
    }
}
```

### Step 5: Regenerate

Run code generation:

```bash
cd /workspaces/cucaracha/pkg/ui/debugger
go generate ./...
```

### What Gets Generated Automatically

The generator creates:

1. **Enum constant:** `DebuggerCommandMyCommand`
2. **Command field:** `MyCommandArgs` in `DebuggerCommand` struct
3. **Result field:** `MyCommandResult` in `DebuggerCommandResult` struct
4. **Switch case:** In `Execute()` method routing to your implementation
5. **Documentation:** Entry in `CommandsDocsSchema` with your method's comment

### Example: Adding StepInto Command

**api.go:**
```go
// StepInto executes until the next function call at the same level or higher
StepInto(args *StepIntoArgs) *ExecutionResult
```

**Generated automatically:**
- Enum: `DebuggerCommandStepInto`
- Switch case in `Execute()` dispatching to `d.debugger.StepInto(cmd.StepIntoArgs)`
- Struct fields: `StepIntoArgs` in `DebuggerCommand`, `StepIntoResult` in `DebuggerCommandResult`
- Documentation entry with comment "executes until the next function call at the same level or higher"

No manual updates needed elsewhere!

## Generator Functions

### Core Functions (in generator_impl.go)

#### ParseDebuggerInterface(apiPath string) (*[]CommandInfo, error)

**Purpose:** Extract command definitions from the DebuggerCommands interface

**Process:**
1. Parse `api.go` using Go's AST parser
2. Find the `DebuggerCommands` interface type
3. For each interface method:
   - Extract method name (PascalCase)
   - Extract parameter type name (must follow `*<MethodName>Args` pattern or be nil)
   - Extract return type name (must follow `*<Type>Result` pattern)
   - Extract documentation comment
4. Build `CommandInfo` struct for each method
5. Return slice of command info

**Returns:** Slice of `CommandInfo` with one entry per interface method

**CommandInfo struct:**
```go
type CommandInfo struct {
    MethodName      string  // Method name (PascalCase)
    CommandID       string  // Enum constant ("DebuggerCommand" + MethodName)
    CommandName     string  // Command name (camelCase)
    Comments        string  // Documentation string
    ArgsType        string  // Parameter type (e.g., "*StepArgs")
    ArgsFieldName   string  // Reference in struct (e.g., "cmd.StepArgs")
    ResultType      string  // Return type (e.g., "*ExecutionResult")
    ResultFieldName string  // Field name in result struct
}
```

#### GenerateExecuteMethod(commands []CommandInfo) ([]byte, error)

**Purpose:** Generate the `Execute()` method implementation

**Output:**
- File package declaration
- Imports (fmt, github.com/Manu343726/cucaracha/pkg/ui/debugger)
- `Execute()` method with switch statement
- One case per command method
- Each case calls the corresponding interface method
- Wraps result in `DebuggerCommandResult`
- Default case returns error for unknown commands

#### GenerateEnumConstants(commands []CommandInfo) (string, error)

**Purpose:** Generate DebuggerCommandId enum and conversion functions

**Output:**
- `type DebuggerCommandId int`
- Const block with iota-based enum constants
- `String()` method mapping enum values to camelCase names
- `DebuggerCommandIdFromString()` function for reverse lookup
- Comments preserved from original interface

#### GenerateCommandStructs(commands []CommandInfo) (string, error)

**Purpose:** Generate `DebuggerCommand` and `DebuggerCommandResult` structs

**Output:**
- `DebuggerCommand` struct:
  - `Id` field (uint64)
  - `Command` field (DebuggerCommandId)
  - Optional field for each method with args
- `DebuggerCommandResult` struct:
  - `Id` field (uint64)
  - `Command` field (DebuggerCommandId)
  - One result field per method

#### GenerateDocsSchema(commands []CommandInfo) (string, error)

**Purpose:** Generate runtime documentation for all commands

**Output:**
- `CommandsDocsSchema` variable declaration
- Populated with `CommandDocumentation` entries
- One entry per method with:
  - Command ID (numeric enum value)
  - Command name (camelCase)
  - Description (from interface comments)
  - Arguments list
  - Result type information

### Helper Functions (in generator_impl.go)

| Function | Purpose | Example |
|----------|---------|---------|
| `typeToString(expr ast.Expr) string` | Convert AST type expressions to Go type strings | `*StepArgs` |
| `baseTypeName(typeName string) string` | Extract base type from pointer | `*Foo` → `Foo` |
| `shouldIncludeInExecute(methodName string) bool` | Filter methods (excludes configuration methods like `SetEventCallback`) | `true` for command methods |
| `methodNameToCommandName(method string) string` | Convert PascalCase to camelCase | `Step` → `step` |

## Excluded Methods

The following methods are excluded from code generation:

- `SetEventCallback`: Not a command, is a configuration method

## Build Output Location

All generated files are created in `/workspaces/cucaracha/pkg/ui/debugger/`:
- `zz_commands_enum_generated.go` (enum + conversions)
- `zz_commands_structs_generated.go` (command & result structs)
- `zz_execute_generated.go` (dispatch logic)
- `zz_commands_documentation_schema.go` (runtime docs)

The `zz_` prefix ensures generated files sort after hand-written code and are easily identifiable.

## Runtime Usage

### How Commands Flow at Runtime

The generated code works together to create a command dispatch system:

```
1. Client sends DebuggerCommand (JSON):
   {
     "id": 1,
     "command": 2,
     "stepargs": { "steps": 5 }
   }
   ↓
2. JSON unmarshaled to DebuggerCommand struct
   ↓
3. Execute() method receives DebuggerCommand
   ↓
4. Switch statement routes by Command field:
   case DebuggerCommandStep:
     result := d.debugger.Step(cmd.StepArgs)
   ↓
5. DebuggerCommandResult created and populated:
   {
     "id": 1,
     "command": 2,
     "stepresult": { "pc": 0x1000, ... }
   }
   ↓
6. Result marshaled to JSON and sent to client
```

### Type Flow Diagram

```
DebuggerCommands Interface (api.go)
  ↓
  Generate: Enum + Structs + Execute() + Docs
  ↓
┌─────────────────────────────────────────────────────┐
│ DebuggerCommand (JSON)                              │
│   - Command: DebuggerCommandId enum value           │
│   - StepArgs, ContinueArgs, ... (optional fields)   │
└──────────────┬──────────────────────────────────────┘
               │
        Execute() method
               │
               ↓
┌─────────────────────────────────────────────────────┐
│ Interface Method Call                               │
│   d.debugger.Step(cmd.StepArgs) // or similar       │
└──────────────┬──────────────────────────────────────┘
               │
               ↓
          DebuggerCommands
          Implementation
               │
               ↓
┌─────────────────────────────────────────────────────┐
│ DebuggerCommandResult (JSON)                        │
│   - Command: echoed from request                    │
│   - StepResult, ContinueResult, ... (result field)  │
└─────────────────────────────────────────────────────┘
```

This flow is completely automated by the generated `Execute()` method - no manual routing code needed.

## Single Source of Truth

This generator makes the `DebuggerCommands` interface the single source of truth for:
- What commands exist
- What parameters each command accepts
- What each command returns
- How commands are dispatched
- Command documentation

Changes to the interface automatically propagate to all generated code on the next `go generate` invocation. There's no need to manually update enum values, struct fields, or dispatch logic.

## Complete Generation Workflow

When you execute `go generate ./...`, here's what happens:

```
1. Go reads //go:generate directive in command.go
   ↓
2. Executes: go run generator.go -out zz_execute_generated.go ...
   ↓
3. generator.go main() runs:
   - Parses command-line flags
   - Calls ParseDebuggerInterface() from debugger package
   ↓
4. ParseDebuggerInterface() executes:
   - Opens api.go
   - Parses Go AST
   - Finds DebuggerCommands interface
   - Extracts all method definitions
   - Returns []CommandInfo with all parsed commands
   ↓
5. generator.go calls four generation functions:
   ├→ GenerateExecuteMethod(commands) → writes zz_execute_generated.go
   ├→ GenerateEnumConstants(commands) → writes zz_commands_enum_generated.go
   ├→ GenerateCommandStructs(commands) → writes zz_commands_structs_generated.go
   └→ GenerateDocsSchema(commands) → writes zz_commands_documentation_schema.go
   ↓
6. All four files are now updated with latest command definitions
   ↓
7. Next compile will use fresh generated code
```

## Debugging the Generator

### Verify Generator Runs

Check that the `//go:generate` directive in [command.go](command.go) is correct and runs:

```bash
cd /workspaces/cucaracha/pkg/ui/debugger
go generate -v ./...
```

The `-v` flag shows each generation command as it runs.

### Examine Generated Code

After running `go generate`, verify the output files were created and contain expected code:

```bash
# Check all four files exist
ls -la zz_*_generated.go

# View generated enum constants
head -50 zz_commands_enum_generated.go

# View generated Execute method (should contain switch cases for each command)
head -100 zz_execute_generated.go

# View generated command structs
head -100 zz_commands_structs_generated.go

# View generated documentation schema
head -50 zz_commands_documentation_schema.go
```

### Verify Package Builds

After generation, ensure the package still compiles:

```bash
cd /workspaces/cucaracha && go build ./pkg/ui/debugger && echo "✓ Build successful"
```

If build fails, check that:
1. All `*Args` types referenced in cmd struct are defined
2. All `*Result` types referenced in execute method are defined
3. The `DebuggerCommands` interface imports match expected signatures
4. No syntax errors in generated code

### Add Debug Output

To debug the parser, temporarily add logging to [generator_impl.go](generator_impl.go):

**In ParseDebuggerInterface():**
```go
for _, method := range commands {
    fmt.Fprintf(os.Stderr, "DEBUG: Command=%s, Args=%s, Result=%s\n",
        method.MethodName, method.ArgsType, method.ResultType)
}
```

Then run:
```bash
cd /workspaces/cucaracha/pkg/ui/debugger
go generate ./... 2>&1 | grep DEBUG
```

This will print parsed command information to verify correct extraction.

### Common Issues

**Issue: "interface not found" error**
- Check that `api.go` exists and contains `DebuggerCommands` interface definition

**Issue: Generated enum values are wrong**
- Verify interface methods follow naming convention (PascalCase)
- Check that enum constants are generated with iota (should be 0, 1, 2, ...)

**Issue: Execute method doesn't compile**
- Verify all interface methods have corresponding `*Args` and `*Result` types
- Check that type names match between interface method signature and struct definitions

**Issue: Build fails after generation**
- Run `go generate ./...` again to refresh generated files
- Check for missing `*Args` or `*Result` type definitions
- Review compilation errors for typos in generated code
