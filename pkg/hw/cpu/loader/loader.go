// Package loader provides high-level APIs for loading Cucaracha programs.
//
// This package abstracts the details of parsing different file formats (assembly,
// binary/ELF, source code) and provides a unified interface for loading programs
// into memory. It handles:
//
//   - File format detection based on extension
//   - Assembly file parsing via LLVM parser
//   - Binary/ELF file parsing with DWARF debug info
//   - Source code compilation via clang
//   - Symbol and address resolution
//
// Typical usage:
//
//	opts := &loader.Options{Verbose: true}
//	pf, cleanup, err := loader.LoadFile("program.c", opts)
//	if err != nil { ... }
//	defer cleanup()
//
// The returned ProgramFile is resolved (has addresses assigned) and ready
// for execution.
package loader

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/llvm"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
)

// Options configures the program loading process
type Options struct {
	// Verbose enables detailed output during loading
	Verbose bool

	// OutputFormat specifies the compilation output format for source files
	// Valid values: "assembly", "object" (default: "object" for DWARF support)
	OutputFormat string

	// MemoryConfig specifies the memory layout configuration
	// If nil, uses DefaultMemoryConfig()
	MemoryConfig *mc.MemoryResolverConfig

	// AutoBuildClang enables automatic building of clang from llvm-project
	// if llvm-project is found but clang hasn't been built yet
	AutoBuildClang bool
}

// DefaultMemoryConfig returns the default memory configuration.
// It places code at 0x10000 to avoid overlap with low absolute addresses
// used by LLVM for local variables.
func DefaultMemoryConfig() mc.MemoryResolverConfig {
	return mc.MemoryResolverConfig{
		BaseAddress:     0x10000, // 64KB offset
		MaxSize:         0,       // No limit
		DataAlignment:   4,
		InstructionSize: 4,
	}
}

// Result contains the result of a load operation
type Result struct {
	// Program is the resolved program file
	Program mc.ProgramFile

	// Cleanup is a function to call when done with the program
	// (removes temporary compiled files). May be nil.
	Cleanup func()

	// OriginalPath is the original file path provided
	OriginalPath string

	// CompiledPath is the path to the file that was loaded
	// (may be a temporary file if source was compiled)
	CompiledPath string

	// WasCompiled is true if the source was compiled
	WasCompiled bool

	// Format is the detected/used file format
	Format FileFormat

	// Warnings contains non-fatal warnings that occurred during loading
	Warnings []string
}

// FileFormat represents the type of program file
type FileFormat int

const (
	// FormatUnknown indicates an unknown file format
	FormatUnknown FileFormat = iota
	// FormatAssembly indicates an assembly file (.cucaracha, .s)
	FormatAssembly
	// FormatBinary indicates a binary/ELF file (.o)
	FormatBinary
	// FormatSource indicates a source file (.c, .cpp, etc.) that was compiled
	FormatSource
)

// String returns the string representation of a FileFormat
func (f FileFormat) String() string {
	switch f {
	case FormatAssembly:
		return "assembly"
	case FormatBinary:
		return "binary"
	case FormatSource:
		return "source"
	default:
		return "unknown"
	}
}

// LoadFile loads a program file from the given path.
// It automatically detects the file format and handles compilation if needed.
// The returned Result.Cleanup function should be called when done.
func LoadFile(path string, opts *Options) (*Result, error) {
	if opts == nil {
		opts = &Options{}
	}

	result := &Result{
		OriginalPath: path,
		CompiledPath: path,
		Cleanup:      func() {}, // No-op by default
	}

	// Check if it's a source file that needs compilation
	if llvm.IsSourceFile(path) {
		outputFormat := llvm.OutputObject // Default for DWARF support
		if opts.OutputFormat != "" {
			switch strings.ToLower(opts.OutputFormat) {
			case "assembly", "asm":
				outputFormat = llvm.OutputAssembly
			case "object", "obj", "o":
				outputFormat = llvm.OutputObject
			}
		}

		compiledPath, cleanup, warnings, err := compileSourceFile(path, outputFormat, opts.Verbose, opts.AutoBuildClang)
		if err != nil {
			return nil, fmt.Errorf("compiling source file: %w", err)
		}

		result.CompiledPath = compiledPath
		result.Cleanup = cleanup
		result.WasCompiled = true
		result.Format = FormatSource
		result.Warnings = append(result.Warnings, warnings...)

		// Update path to the compiled file
		path = compiledPath
	}

	// Detect file format by extension
	ext := strings.ToLower(filepath.Ext(path))

	var pf mc.ProgramFile
	var err error

	switch ext {
	case ".cucaracha", ".s":
		pf, err = llvm.ParseAssemblyFile(path)
		if result.Format == FormatUnknown {
			result.Format = FormatAssembly
		}
	case ".o":
		pf, err = llvm.ParseBinaryFile(path)
		if result.Format == FormatUnknown {
			result.Format = FormatBinary
		}
	default:
		return nil, fmt.Errorf("unsupported file extension '%s' (supported: .cucaracha, .s, .o, .c, .cpp)", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("parsing file: %w", err)
	}

	// Resolve the program
	memConfig := DefaultMemoryConfig()
	if opts.MemoryConfig != nil {
		memConfig = *opts.MemoryConfig
	}

	resolved, err := mc.Resolve(pf, memConfig)
	if err != nil {
		return nil, fmt.Errorf("resolving program: %w", err)
	}

	result.Program = resolved
	return result, nil
}

// IsSupportedFile returns true if the file extension is supported
func IsSupportedFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".cucaracha", ".s", ".o":
		return true
	}
	return llvm.IsSourceFile(path)
}

// compileSourceFile compiles a C/C++ source file to the specified output format.
// Returns the path to the compiled file, a cleanup function, any warnings, and any error.
func compileSourceFile(inputPath string, outputFormat llvm.OutputFormat, verbose bool, autoBuild bool) (string, func(), []string, error) {
	var warnings []string

	// Discover clang with auto-build option
	discoverOpts := &llvm.DiscoverClangOptions{
		AutoBuild: autoBuild,
		Verbose:   verbose,
	}
	toolchain, err := llvm.DiscoverClang(nil, discoverOpts)
	if err != nil {
		return "", nil, nil, fmt.Errorf("clang not found: %w", err)
	}

	// Warn if using system clang instead of llvm-project build
	if toolchain.IsSystemClang() {
		warnings = append(warnings, fmt.Sprintf(
			"using system clang (%s) instead of llvm-project build; "+
				"ensure your system clang supports the Cucaracha target",
			toolchain.ClangPath()))
	}

	// Compile to temp file
	compileOpts := &llvm.CompileOptions{
		OutputFormat: outputFormat,
		OptLevel:     llvm.OptNone,
		DebugInfo:    true, // Always include debug info for debugging support
		Verbose:      verbose,
	}

	compileResult, err := toolchain.CompileToTemp(inputPath, compileOpts)
	if err != nil {
		return "", nil, nil, err
	}

	// Return cleanup function
	cleanup := func() {
		compileResult.Cleanup()
	}

	return compileResult.OutputPath, cleanup, warnings, nil
}
