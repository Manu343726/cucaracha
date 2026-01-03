package cpu

import (
	"fmt"
	"os"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/llvm"
	"github.com/spf13/cobra"
)

var (
	clangOutputPath   string
	clangOutputFormat string
	clangOptLevel     string
	clangVerbose      bool
	clangClangPath    string
	clangLLVMRoot     string
	clangBuildConfig  string
	clangBuildClang   bool
	clangIncludes     []string
	clangDefines      []string
	clangExtraFlags   []string
)

var clangCmd = &cobra.Command{
	Use:   "clang <source-file>",
	Short: "Compile C/C++ source files to Cucaracha target",
	Long: `Compiles C/C++ source files to Cucaracha assembly or object files.

This command acts as a driver for clang with the cucaracha target.
It will automatically discover the clang binary with Cucaracha support,
either from the project's LLVM build or from the system PATH.

Output formats:
  assembly  - Produces .cucaracha assembly files (default)
  object    - Produces .o ELF object files
  llvm-ir   - Produces .ll LLVM IR files

Optimization levels:
  0  - No optimization (default)
  1  - Basic optimization
  2  - Standard optimization
  3  - Aggressive optimization
  s  - Optimize for size
  z  - Optimize for size aggressively

Examples:
  # Compile to assembly
  cucaracha cpu clang hello.c

  # Compile to object file with optimization
  cucaracha cpu clang -f object -O 2 hello.c

  # Specify output path
  cucaracha cpu clang -o output.cucaracha hello.c

  # Use specific clang binary
  cucaracha cpu clang --clang-path /path/to/clang hello.c

  # Build clang from sources if not found
  cucaracha cpu clang --build-clang hello.c`,
	Args: cobra.ExactArgs(1),
	Run:  runClang,
}

func init() {
	CpuCmd.AddCommand(clangCmd)

	clangCmd.Flags().StringVarP(&clangOutputPath, "output", "o", "", "Output file path (default: based on input)")
	clangCmd.Flags().StringVarP(&clangOutputFormat, "format", "f", "assembly", "Output format: assembly, object, llvm-ir")
	clangCmd.Flags().StringVarP(&clangOptLevel, "opt", "O", "0", "Optimization level: 0, 1, 2, 3, s, z")
	clangCmd.Flags().BoolVarP(&clangVerbose, "verbose", "v", false, "Print verbose output")
	clangCmd.Flags().StringVar(&clangClangPath, "clang-path", "", "Explicit path to clang binary")
	clangCmd.Flags().StringVar(&clangLLVMRoot, "llvm-root", "", "LLVM project root directory")
	clangCmd.Flags().StringVar(&clangBuildConfig, "build-config", "Release", "Build configuration (Release, Debug)")
	clangCmd.Flags().BoolVar(&clangBuildClang, "build-clang", false, "Build clang from sources if not found")
	clangCmd.Flags().StringArrayVarP(&clangIncludes, "include", "I", nil, "Add include directory")
	clangCmd.Flags().StringArrayVarP(&clangDefines, "define", "D", nil, "Add preprocessor definition")
	clangCmd.Flags().StringArrayVarP(&clangExtraFlags, "Xclang", "X", nil, "Pass extra flag to clang")
}

func runClang(cmd *cobra.Command, args []string) {
	inputPath := args[0]

	// Validate input file
	if !llvm.IsSourceFile(inputPath) {
		fmt.Fprintf(os.Stderr, "Error: %s is not a recognized C/C++ source file\n", inputPath)
		fmt.Fprintln(os.Stderr, "Supported extensions: .c, .cc, .cpp, .cxx, .c++, .m, .mm")
		os.Exit(1)
	}

	// Configure toolchain discovery
	config := &llvm.ClangConfig{
		ClangPath:   clangClangPath,
		LLVMRoot:    clangLLVMRoot,
		BuildConfig: clangBuildConfig,
	}

	// Discover clang
	toolchain, err := llvm.DiscoverClang(config)
	if err != nil {
		if clangBuildClang {
			fmt.Fprintln(os.Stderr, "Clang not found, attempting to build from sources...")
			// Create a toolchain with LLVM root for building
			if config.LLVMRoot == "" {
				// Try to find llvm-project
				fmt.Fprintln(os.Stderr, "Error: --llvm-root required when using --build-clang")
				os.Exit(1)
			}
			toolchain = &llvm.ClangToolchain{}
			if buildErr := toolchain.BuildClang(clangVerbose); buildErr != nil {
				fmt.Fprintf(os.Stderr, "Error building clang: %v\n", buildErr)
				os.Exit(1)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			fmt.Fprintln(os.Stderr, "\nTo build clang from sources, use --build-clang with --llvm-root")
			os.Exit(1)
		}
	}

	if clangVerbose {
		version, _ := toolchain.Version()
		fmt.Fprintf(os.Stderr, "Using clang: %s\n", toolchain.ClangPath())
		if version != "" {
			fmt.Fprintf(os.Stderr, "Version: %s\n", version)
		}
	}

	// Parse output format
	var outputFormat llvm.OutputFormat
	switch strings.ToLower(clangOutputFormat) {
	case "assembly", "asm", "s":
		outputFormat = llvm.OutputAssembly
	case "object", "obj", "o":
		outputFormat = llvm.OutputObject
	case "llvm-ir", "llvm", "ir", "ll":
		outputFormat = llvm.OutputLLVMIR
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown output format '%s'\n", clangOutputFormat)
		os.Exit(1)
	}

	// Parse optimization level
	var optLevel llvm.OptLevel
	switch strings.ToLower(clangOptLevel) {
	case "0":
		optLevel = llvm.OptNone
	case "1":
		optLevel = llvm.OptLess
	case "2":
		optLevel = llvm.OptDefault
	case "3":
		optLevel = llvm.OptAggressive
	case "s":
		optLevel = llvm.OptSize
	case "z":
		optLevel = llvm.OptSizeAggressive
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown optimization level '%s'\n", clangOptLevel)
		os.Exit(1)
	}

	// Build compile options
	opts := &llvm.CompileOptions{
		OutputFormat: outputFormat,
		OptLevel:     optLevel,
		OutputPath:   clangOutputPath,
		IncludePaths: clangIncludes,
		Defines:      clangDefines,
		ExtraFlags:   clangExtraFlags,
		Verbose:      clangVerbose,
	}

	// Compile
	result, err := toolchain.Compile(inputPath, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Compilation failed: %v\n", err)
		if result != nil && result.Stderr != "" {
			fmt.Fprintln(os.Stderr, result.Stderr)
		}
		os.Exit(1)
	}

	if clangVerbose {
		fmt.Fprintf(os.Stderr, "Output: %s\n", result.OutputPath)
	} else {
		fmt.Println(result.OutputPath)
	}
}

// CompileSourceFile compiles a source file using the discovered clang toolchain
// This is a helper function for use by exec and debug commands
func CompileSourceFile(inputPath string, outputFormat llvm.OutputFormat, verbose bool) (string, func(), error) {
	// Discover clang
	toolchain, err := llvm.DiscoverClang(nil)
	if err != nil {
		return "", nil, fmt.Errorf("clang not found: %v", err)
	}

	if verbose {
		version, _ := toolchain.Version()
		fmt.Fprintf(os.Stderr, "Compiling %s with clang...\n", inputPath)
		if version != "" {
			fmt.Fprintf(os.Stderr, "Using: %s\n", version)
		}
	}

	// Compile to temp file
	opts := &llvm.CompileOptions{
		OutputFormat: outputFormat,
		OptLevel:     llvm.OptNone,
		DebugInfo:    true, // Always include debug info for exec/debug commands
		Verbose:      verbose,
	}

	result, err := toolchain.CompileToTemp(inputPath, opts)
	if err != nil {
		return "", nil, err
	}

	// Return cleanup function
	cleanup := func() {
		result.Cleanup()
	}

	return result.OutputPath, cleanup, nil
}
