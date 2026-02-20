# Cucaracha Reflect and Codegen Packages

This document describes two complementary packages for Go code reflection and generation:

## Philosophy

Both packages follow the same philosophy as the debugger generator:

1. **Use Go's Standard Tools**: Leverage `go/parser` and `go/ast` packages for robust, standards-compliant code analysis
2. **Extract Semantic Information**: Parse AST structures to extract meaningful metadata about types, functions, constants, and enums
3. **Generate Using Conventions**: Generate code based on naming conventions and patterns extracted from the parsed metadata
4. **Separate Concerns**: Keep reflection logic separate from generation logic, with clear APIs between them

## Package: reflect

The `reflect` package provides an interface for parsing Go packages and files, extracting structural information about types, functions, constants, and enums with their documentation.

### Core Types

- **Package**: Represents a complete Go package with all its files, types, functions, constants, and identified enums
- **File**: Represents a single Go source file with its imports and declarations
- **Type**: Represents a named type (struct, interface, or alias) with fields and methods
- **Function**: Represents a top-level function or method with parameters and results
- **Constant**: Represents a package-level constant or variable
- **Enum**: Represents a logical grouping of constants that form an enum type
- **Parameter**: Represent function parameter with name and type

### Query Methods

The `Package` type provides comprehensive query methods for finding types, functions, and other metadata:

```go
pkg, _ := reflect.ParsePackage("./mypackage")

// Find specific elements
typ := pkg.FindType("MyStruct")
fn := pkg.FindFunction("MyFunction")
enum := pkg.FindEnum("MyEnum")
constant := pkg.FindConstantByName("MyConst")

// Find multiple elements by criteria
structs := pkg.FindTypesByKind(reflect.TypeKindStruct)
interfaces := pkg.FindTypesByKind(reflect.TypeKindInterface)

// Get exported (public) items
publicTypes := pkg.GetPublicTypes()
publicFns := pkg.GetPublicFunctions()

// Find functions by name pattern
helpers := pkg.FindFunctionsByPrefix("Helper")

// Query structural information
methods := pkg.GetMethodsForType("MyType")
fields := pkg.GetStructFields("MyStruct")
ifaceMethods := pkg.GetInterfaceMethods("MyInterface")

// Find types with specific characteristics
structsWithField := pkg.FindStructsWithField("ID")
structsWithFieldType := pkg.FindStructsWithFieldType("string")
structsWithMethod := pkg.GetMethodsByName("String")

// Find constants
constants := pkg.FindConstantsByType("int")

// Navigate files
file := pkg.GetFileByName("types.go")
```

### Main APIs

#### Parsing a Single File

```go
import "github.com/Manu343726/cucaracha/pkg/reflect"

file, err := reflect.ParseFile("mypackage/file.go")
if err != nil {
    log.Fatal(err)
}

for name, typ := range file.Types {
    println(name, typ.Kind)
}
```

#### Parsing an Entire Package

```go
pkg, err := reflect.ParsePackage("./mypackage")
if err != nil {
    log.Fatal(err)
}

// Access all types, functions, and identified enums
for name, typ := range pkg.Types {
    println(name)
}

// Search for specific elements
step := pkg.FindFunction("Step")
if step != nil {
    for _, arg := range step.Args {
        println(arg.Name, arg.Type)
    }
}
```

#### Parsing a Package by Import Path

The `reflect` package can resolve Go import paths to their filesystem locations, automatically handling go.mod, GOPATH, and the standard library:

```go
// Parse a standard library package
pkg, err := reflect.ParsePackageFromImport("fmt")
if err != nil {
    log.Fatal(err)
}

// Parse an external package
pkg, err := reflect.ParsePackageFromImport("github.com/Manu343726/cucaracha/pkg/hw")
if err != nil {
    log.Fatal(err)
}

// Parse a local package (from the current module)
pkg, err := reflect.ParsePackageFromImport("github.com/myuser/myrepo/pkg/utils")
if err != nil {
    log.Fatal(err)
}
```

### Package Resolution

The `PackageResolver` handles Go's standard package resolution:

```go
// Create a resolver starting from a specific directory
resolver, err := reflect.NewPackageResolver("./myproject")
if err != nil {
    log.Fatal(err)
}

// Resolve import paths to filesystem locations
path, err := resolver.ResolvePackagePath("github.com/user/repo/pkg")
if err != nil {
    log.Fatal(err)
}

// Get detailed module information
info, err := reflect.GetModuleInfo("github.com/user/repo", "./myproject")
if err != nil {
    log.Fatal(err)
}
```

The resolver automatically:
- Searches for go.mod in the current directory and parent directories
- Queries the Go module cache using `go list`
- Handles standard library packages
- Resolves external module imports
- Supports local packages within the current module

### Enum Identification

The package automatically identifies enums by looking for:
1. Types that correspond to constant groups
2. All constants sharing the same underlying type
3. Optional String() methods on the type

Example:
```go
// From the parsed code
type DebuggerCommandId int

const (
    DebuggerCommandStep DebuggerCommandId = iota
    DebuggerCommandContinue
    DebuggerCommandBreakpoint
)
```

After parsing, `pkg.Enums["DebuggerCommandId"]` contains metadata about this enum.

## Package: codegen

The `codegen` package takes reflect metadata describing Go packages, files, types, and functions, then generates equivalent Go code.

### Core Types

- **Generator**: Main interface for generating code from reflect metadata
- **FileBuilder**: Fluent builder for constructing Go source files programmatically
- **StructBuilder**: Builder for creating struct definitions
- **InterfaceBuilder**: Builder for creating interface definitions
- **EnumBuilder**: Builder for creating enum types with constants

### Generation Patterns

#### From Existing Reflect Metadata

```go
import "github.com/Manu343726/cucaracha/pkg/codegen"

gen := codegen.NewGenerator(pkg)

// Generate struct definition
code, err := gen.GenerateStructCode(myStructType)

// Generate interface definition
code, err := gen.GenerateInterfaceCode(myInterfaceType)

// Generate enum definition with String() method
code := gen.GenerateEnumCode(myEnum)
```

#### Building Code Programmatically

```go
// Build a new struct definition
myStruct := codegen.NewStructBuilder("MyType").
    WithDoc("MyType represents something.").
    AddField("Name", "string", "Name of the item", `json:"name"`).
    AddField("Count", "int", "How many", `json:"count"`).
    Build()

code, _ := gen.GenerateStructCode(myStruct)

// Build a new interface definition
myInterface := codegen.NewInterfaceBuilder("Reader").
    WithDoc("Reader reads data.").
    AddMethod("Read", 
        []*reflect.Parameter{{Name: "p", Type: "[]byte"}},
        []*reflect.Parameter{{Type: "int"}, {Type: "error"}},
        "Read some bytes").
    Build()

code, _ := gen.GenerateInterfaceCode(myInterface)

// Build an enum
myEnum := codegen.NewEnumBuilder("Status", "int").
    WithDoc("Status represents the current status.").
    AddValue("StatusPending", "Pending state").
    AddValue("StatusActive", "Active state").
    AddValue("StatusDone", "Completed state").
    Build()

code := gen.GenerateEnumCode(myEnum)
```

#### Building Complete Files

```go
file := codegen.NewFileBuilder("mypackage").
    AddImport("fmt").
    AddImport("errors").
    AddCode("// Some custom code").
    AddStruct(myStructType).
    AddInterface(myInterfaceType).
    AddEnum(myEnum).
    Build()

// Write to file
os.WriteFile("generated.go", []byte(file), 0644)
```

## Usage Patterns

### Pattern 0: Package Resolution with Import Syntax

Parse packages using standard Go import paths - the resolver automatically handles package discovery:

```go
// Parse a standard library package
pkg, err := reflect.ParsePackageFromImport("fmt")
if err != nil {
    log.Fatal(err)
}

// Parse an external package from go.mod dependencies
pkg, err := reflect.ParsePackageFromImport("github.com/Manu343726/cucaracha/pkg/hw")
if err != nil {
    log.Fatal(err)
}

// Parse a local package from the current module
pkg, err := reflect.ParsePackageFromImport("github.com/user/myrepo/internal/utils")
if err != nil {
    log.Fatal(err)
}

// Detailed control with explicit directory
pkg, err := reflect.ParsePackageFromImportInDir("fmt", "./myproject")
if err != nil {
    log.Fatal(err)
}

// Get resolver to resolve multiple paths
resolver, err := reflect.NewPackageResolver("./myproject")
if err != nil {
    log.Fatal(err)
}

// Resolve import path to filesystem location
fspath, err := resolver.ResolvePackagePath("github.com/user/repo/pkg")
if err != nil {
    log.Fatal(err)
}
println("Package location:", fspath)

// Get detailed module information
info, err := reflect.GetModuleInfo("github.com/user/repo", "./myproject")
if err != nil {
    log.Fatal(err)
}
println("Module version:", info["Version"])
```

### Pattern 1: Code Generator Tool

Create a standalone generator tool similar to the debugger generator:

```go
//go:build ignore

package main

import (
    "flag"
    "os"
    "github.com/Manu343726/cucaracha/pkg/reflect"
    "github.com/Manu343726/cucaracha/pkg/codegen"
)

func main() {
    apiPath := flag.String("api", "", "path to interface file")
    outPath := flag.String("out", "", "output path")
    flag.Parse()

    pkg, _ := reflect.ParsePackage(filepath.Dir(*apiPath))
    iface := pkg.FindType("MyInterface")
    
    gen := codegen.NewGenerator(pkg)
    code, _ := gen.GenerateInterfaceCode(iface)
    
    os.WriteFile(*outPath, []byte(code), 0644)
}
```

### Pattern 2: Programmatic Code Generation

Generate code within your application:

```go
pkg, err := reflect.ParsePackage("./mypackage")
if err != nil {
    return err
}

gen := codegen.NewGenerator(pkg)

// Generate all enums
for name, enum := range pkg.Enums {
    code := gen.GenerateEnumCode(enum)
    // Use the generated code...
}
```

### Pattern 3: Building New Types from Scratch

Create types programmatically:

```go
resultType := codegen.NewStructBuilder("CommandResult").
    AddField("Id", "string").
    AddField("Status", "int").
    AddField("Data", "interface{}").
    Build()

code, _ := codegen.NewGenerator(pkg).GenerateStructCode(resultType)
```

## Design Principles

1. **Separation of Concerns**
   - `reflect` handles parsing and metadata extraction
   - `codegen` handles code generation
   - They communicate through immutable metadata types

2. **Naming Conventions**
   - Types are identified by their Go names
   - Functions and methods preserve their original names
   - Generated code maintains idiomatic Go formatting

3. **Documentation Preservation**
   - All documentation comments are extracted and preserved
   - Generated code includes original documentation

4. **Extensibility**
   - Builders provide fluent APIs for programmatic code construction
   - Generators can be extended for custom generation logic
   - Type information is preserved for downstream tools

5. **Robustness**
   - Uses Go's standard `go/parser` and `go/ast` packages
   - Handles various code patterns and edge cases
   - Error handling throughout the API

## Differences from Debugger Generator

While following the same philosophy, these packages are more general-purpose:

| Aspect | Debugger Generator | reflect/codegen |
|--------|-------------------|-----------------|
| Scope | Parse a single interface | Parse entire packages/files |
| Metadata | Command methods only | All types, functions, constants, enums |
| Output | Generate Execute() methods and related code | Generate any Go code from metadata |
| Builders | N/A | Fluent builders for programmatic construction |
| Reusability | Specific to debugger domain | General-purpose, domain-agnostic |

## Example: Complete Workflow

```go
package main

import (
    "os"
    "github.com/Manu343726/cucaracha/pkg/reflect"
    "github.com/Manu343726/cucaracha/pkg/codegen"
)

func main() {
    // Parse the package
    pkg, _ := reflect.ParsePackage("./mypackage")
    
    // Find a type
    myType := pkg.FindType("MyStruct")
    
    // Create a generator
    gen := codegen.NewGenerator(pkg)
    
    // Generate code for that type
    code, _ := gen.GenerateStructCode(myType)
    
    // Build a file with the generated code
    file := codegen.NewFileBuilder(pkg.Name).
        AddCode(code).
        Build()
    
    // Write to output
    os.WriteFile("output.go", []byte(file), 0644)
}
```
