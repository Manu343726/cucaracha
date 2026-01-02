package llvm

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// OutputFormat specifies the output format for compilation
type OutputFormat int

const (
	// OutputAssembly produces .cucaracha assembly files
	OutputAssembly OutputFormat = iota
	// OutputObject produces .o ELF object files
	OutputObject
	// OutputLLVMIR produces .ll LLVM IR files
	OutputLLVMIR
)

// String returns the file extension for the output format
func (f OutputFormat) String() string {
	switch f {
	case OutputAssembly:
		return "cucaracha"
	case OutputObject:
		return "o"
	case OutputLLVMIR:
		return "ll"
	default:
		return "o"
	}
}

// ClangFlags returns the clang flags for this output format
func (f OutputFormat) ClangFlags() []string {
	switch f {
	case OutputAssembly:
		return []string{"-S"}
	case OutputObject:
		return []string{"-c"}
	case OutputLLVMIR:
		return []string{"-S", "-emit-llvm"}
	default:
		return []string{"-c"}
	}
}

// OptLevel specifies the optimization level
type OptLevel int

const (
	OptNone           OptLevel = iota // -O0
	OptLess                           // -O1
	OptDefault                        // -O2
	OptAggressive                     // -O3
	OptSize                           // -Os
	OptSizeAggressive                 // -Oz
)

// String returns the clang flag for the optimization level
func (o OptLevel) String() string {
	switch o {
	case OptNone:
		return "-O0"
	case OptLess:
		return "-O1"
	case OptDefault:
		return "-O2"
	case OptAggressive:
		return "-O3"
	case OptSize:
		return "-Os"
	case OptSizeAggressive:
		return "-Oz"
	default:
		return "-O0"
	}
}

// ClangConfig holds configuration for the Clang toolchain
type ClangConfig struct {
	// ClangPath is the explicit path to the clang executable
	ClangPath string

	// LLVMRoot is the root directory of the LLVM installation or build
	LLVMRoot string

	// BuildDir is the build directory within LLVMRoot (e.g., "build_vs2022")
	BuildDir string

	// BuildConfig is the build configuration (e.g., "Release", "Debug")
	BuildConfig string

	// ProjectRoot is the root of the cucaracha project (for auto-discovery)
	ProjectRoot string
}

// DefaultConfig returns a default configuration that will use auto-discovery
func DefaultConfig() *ClangConfig {
	return &ClangConfig{
		BuildConfig: "Release",
	}
}

// ClangToolchain represents a discovered or configured Clang toolchain
type ClangToolchain struct {
	config    *ClangConfig
	clangPath string
	llvmRoot  string
	isSystem  bool // true if using system clang, false if using project build
}

// ClangPath returns the path to the clang executable
func (t *ClangToolchain) ClangPath() string {
	return t.clangPath
}

// LLVMRoot returns the LLVM root directory
func (t *ClangToolchain) LLVMRoot() string {
	return t.llvmRoot
}

// IsSystemClang returns true if using system-installed clang
func (t *ClangToolchain) IsSystemClang() bool {
	return t.isSystem
}

// DiscoverClang attempts to find or configure a Clang toolchain
// Search order:
// 1. Explicit ClangPath in config
// 2. Project build directory (llvm-project sibling to cucaracha)
// 3. System PATH
func DiscoverClang(config *ClangConfig) (*ClangToolchain, error) {
	if config == nil {
		config = DefaultConfig()
	}

	toolchain := &ClangToolchain{config: config}

	// 1. Check explicit ClangPath
	if config.ClangPath != "" {
		if _, err := os.Stat(config.ClangPath); err == nil {
			toolchain.clangPath = config.ClangPath
			toolchain.llvmRoot = config.LLVMRoot
			return toolchain, nil
		}
		return nil, fmt.Errorf("specified clang path not found: %s", config.ClangPath)
	}

	// 2. Check project build directory
	projectClang, llvmRoot := findProjectClang(config)
	if projectClang != "" {
		toolchain.clangPath = projectClang
		toolchain.llvmRoot = llvmRoot
		toolchain.isSystem = false
		return toolchain, nil
	}

	// 3. Check system PATH
	systemClang := findSystemClang()
	if systemClang != "" {
		// Verify it supports cucaracha target
		if supportsCucarachaTarget(systemClang) {
			toolchain.clangPath = systemClang
			toolchain.isSystem = true
			return toolchain, nil
		}
		// System clang found but doesn't support cucaracha
		return nil, fmt.Errorf("system clang found at %s but does not support cucaracha target", systemClang)
	}

	return nil, fmt.Errorf("could not find clang with cucaracha support; please build LLVM with -DLLVM_EXPERIMENTAL_TARGETS_TO_BUILD=Cucaracha")
}

// findProjectClang looks for clang in the project's LLVM build directory
func findProjectClang(config *ClangConfig) (clangPath, llvmRoot string) {
	// Determine project root
	projectRoot := config.ProjectRoot

	// Collect multiple candidate roots
	var candidateRoots []string

	// Add explicit project root if provided
	if projectRoot != "" {
		candidateRoots = append(candidateRoots, projectRoot)
	}

	// Add current working directory and its parents
	if cwd, err := os.Getwd(); err == nil {
		candidateRoots = append(candidateRoots, cwd)
		candidateRoots = append(candidateRoots, filepath.Dir(cwd))
		candidateRoots = append(candidateRoots, filepath.Dir(filepath.Dir(cwd)))
	}

	// Add executable directory and its parents
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidateRoots = append(candidateRoots, exeDir)
		candidateRoots = append(candidateRoots, filepath.Dir(exeDir))
		candidateRoots = append(candidateRoots, filepath.Dir(filepath.Dir(exeDir)))
	}

	// Build list of possible llvm-project locations
	var possibleRoots []string
	for _, root := range candidateRoots {
		possibleRoots = append(possibleRoots,
			filepath.Join(root, "llvm-project"),
			filepath.Join(root, "..", "llvm-project"),
		)
	}

	// Add explicit LLVM root if provided
	if config.LLVMRoot != "" {
		possibleRoots = append([]string{config.LLVMRoot}, possibleRoots...)
	}

	// Build directories to check
	buildDirs := []string{config.BuildDir}
	if config.BuildDir == "" {
		if runtime.GOOS == "windows" {
			buildDirs = []string{"build_vs2022", "build", "build_msvc"}
		} else {
			buildDirs = []string{"build", "build_gcc", "build_clang", "build_docker_linux_gcc"}
		}
	}

	// Build configs to check
	buildConfigs := []string{config.BuildConfig}
	if config.BuildConfig == "" {
		buildConfigs = []string{"Release", "Debug", "RelWithDebInfo", "MinSizeRel"}
	}

	clangExe := "clang"
	if runtime.GOOS == "windows" {
		clangExe = "clang.exe"
	}

	for _, root := range possibleRoots {
		if root == "" {
			continue
		}
		absRoot, err := filepath.Abs(root)
		if err != nil {
			continue
		}

		for _, buildDir := range buildDirs {
			for _, buildConfig := range buildConfigs {
				// Try different path patterns
				paths := []string{
					filepath.Join(absRoot, buildDir, buildConfig, "bin", clangExe),
					filepath.Join(absRoot, buildDir, "bin", clangExe),
					filepath.Join(absRoot, buildDir, buildConfig, clangExe),
				}

				for _, path := range paths {
					if _, err := os.Stat(path); err == nil {
						return path, absRoot
					}
				}
			}
		}
	}

	return "", ""
}

// findSystemClang looks for clang in the system PATH
func findSystemClang() string {
	clangExe := "clang"
	if runtime.GOOS == "windows" {
		clangExe = "clang.exe"
	}

	path, err := exec.LookPath(clangExe)
	if err == nil {
		return path
	}
	return ""
}

// supportsCucarachaTarget checks if a clang binary supports the cucaracha target
func supportsCucarachaTarget(clangPath string) bool {
	cmd := exec.Command(clangPath, "--print-targets")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "cucaracha")
}

// CompileOptions specifies options for compilation
type CompileOptions struct {
	// OutputFormat specifies the output format (assembly, object, LLVM IR)
	OutputFormat OutputFormat

	// OptLevel specifies the optimization level
	OptLevel OptLevel

	// OutputPath is the output file path (optional, auto-generated if empty)
	OutputPath string

	// IncludePaths is a list of additional include directories
	IncludePaths []string

	// Defines is a list of preprocessor definitions
	Defines []string

	// ExtraFlags is a list of additional flags to pass to clang
	ExtraFlags []string

	// Verbose prints the clang command before executing
	Verbose bool

	// KeepTempFiles keeps intermediate temporary files
	KeepTempFiles bool
}

// DefaultCompileOptions returns default compilation options
func DefaultCompileOptions() *CompileOptions {
	return &CompileOptions{
		OutputFormat: OutputObject,
		OptLevel:     OptNone,
	}
}

// CompileResult contains the result of a compilation
type CompileResult struct {
	// OutputPath is the path to the output file
	OutputPath string

	// Command is the clang command that was executed
	Command string

	// Stdout is the standard output from clang
	Stdout string

	// Stderr is the standard error from clang
	Stderr string

	// TempFiles is a list of temporary files created (if KeepTempFiles is false, these are deleted)
	TempFiles []string
}

// Cleanup removes temporary files created during compilation
func (r *CompileResult) Cleanup() {
	for _, f := range r.TempFiles {
		os.Remove(f)
	}
}

// Compile compiles a source file to Cucaracha target
func (t *ClangToolchain) Compile(inputPath string, opts *CompileOptions) (*CompileResult, error) {
	if opts == nil {
		opts = DefaultCompileOptions()
	}

	// Validate input file exists
	if _, err := os.Stat(inputPath); err != nil {
		return nil, fmt.Errorf("input file not found: %s", inputPath)
	}

	// Determine output path
	outputPath := opts.OutputPath
	if outputPath == "" {
		ext := opts.OutputFormat.String()
		base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
		outputPath = filepath.Join(filepath.Dir(inputPath), base+"."+ext)
	}

	// Build clang command
	args := []string{
		"--target=cucaracha",
		opts.OptLevel.String(),
	}

	// Add output format flags
	args = append(args, opts.OutputFormat.ClangFlags()...)

	// Add include paths
	for _, inc := range opts.IncludePaths {
		args = append(args, "-I"+inc)
	}

	// Add defines
	for _, def := range opts.Defines {
		args = append(args, "-D"+def)
	}

	// Add extra flags
	args = append(args, opts.ExtraFlags...)

	// Add output and input
	args = append(args, "-o", outputPath, inputPath)

	// Create command
	cmd := exec.Command(t.clangPath, args...)

	result := &CompileResult{
		OutputPath: outputPath,
		Command:    fmt.Sprintf("%s %s", t.clangPath, strings.Join(args, " ")),
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Running: %s\n", result.Command)
	}

	// Execute - in verbose mode, stream output directly to stderr
	var output []byte
	var err error

	if opts.Verbose {
		// Stream output directly to stderr for real-time visibility
		var stdoutBuf, stderrBuf strings.Builder
		cmd.Stdout = io.MultiWriter(os.Stderr, &stdoutBuf)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
		err = cmd.Run()
		result.Stdout = stdoutBuf.String()
		result.Stderr = stderrBuf.String()
		output = []byte(result.Stdout + result.Stderr)
	} else {
		output, err = cmd.CombinedOutput()
		result.Stderr = string(output)
	}

	if err != nil {
		return result, fmt.Errorf("compilation failed: %v\n%s", err, output)
	}

	return result, nil
}

// CompileToTemp compiles a source file to a temporary output file
func (t *ClangToolchain) CompileToTemp(inputPath string, opts *CompileOptions) (*CompileResult, error) {
	if opts == nil {
		opts = DefaultCompileOptions()
	}

	// Create temp file with appropriate extension
	ext := opts.OutputFormat.String()
	base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))

	tempDir := os.TempDir()
	tempPath := filepath.Join(tempDir, fmt.Sprintf("cucaracha_%s_%d.%s", base, os.Getpid(), ext))

	optsCopy := *opts
	optsCopy.OutputPath = tempPath

	result, err := t.Compile(inputPath, &optsCopy)
	if err != nil {
		return result, err
	}

	result.TempFiles = append(result.TempFiles, tempPath)
	return result, nil
}

// Version returns the clang version string
func (t *ClangToolchain) Version() (string, error) {
	cmd := exec.Command(t.clangPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}
	return "", nil
}

// BuildClang attempts to build clang from the LLVM sources
func (t *ClangToolchain) BuildClang(verbose bool) error {
	if t.llvmRoot == "" {
		return fmt.Errorf("LLVM root not set; cannot build clang")
	}

	// Determine build directory
	buildDir := t.config.BuildDir
	if buildDir == "" {
		if runtime.GOOS == "windows" {
			buildDir = "build_vs2022"
		} else {
			buildDir = "build"
		}
	}

	buildPath := filepath.Join(t.llvmRoot, buildDir)

	// Check if build directory exists, create if not
	if _, err := os.Stat(buildPath); os.IsNotExist(err) {
		if err := os.MkdirAll(buildPath, 0755); err != nil {
			return fmt.Errorf("failed to create build directory: %v", err)
		}
	}

	// Determine build configuration
	buildConfig := t.config.BuildConfig
	if buildConfig == "" {
		buildConfig = "Release"
	}

	// Run CMake configure if needed
	cmakeCachePath := filepath.Join(buildPath, "CMakeCache.txt")
	if _, err := os.Stat(cmakeCachePath); os.IsNotExist(err) {
		if err := t.runCMakeConfigure(buildPath, buildConfig, verbose); err != nil {
			return fmt.Errorf("CMake configure failed: %v", err)
		}
	}

	// Run CMake build
	if err := t.runCMakeBuild(buildPath, buildConfig, verbose); err != nil {
		return fmt.Errorf("CMake build failed: %v", err)
	}

	// Update clang path after successful build
	clangExe := "clang"
	if runtime.GOOS == "windows" {
		clangExe = "clang.exe"
	}

	newClangPath := filepath.Join(buildPath, buildConfig, "bin", clangExe)
	if _, err := os.Stat(newClangPath); err == nil {
		t.clangPath = newClangPath
		return nil
	}

	// Try alternate path
	newClangPath = filepath.Join(buildPath, "bin", clangExe)
	if _, err := os.Stat(newClangPath); err == nil {
		t.clangPath = newClangPath
		return nil
	}

	return fmt.Errorf("build completed but clang not found at expected location")
}

// runCMakeConfigure runs CMake configuration
func (t *ClangToolchain) runCMakeConfigure(buildPath, buildConfig string, verbose bool) error {
	llvmPath := filepath.Join(t.llvmRoot, "llvm")

	var args []string

	if runtime.GOOS == "windows" {
		args = []string{
			"-G", "Visual Studio 17 2022",
			"-A", "x64",
		}
	} else {
		// Try to use Ninja if available
		if _, err := exec.LookPath("ninja"); err == nil {
			args = []string{"-G", "Ninja"}
		} else {
			args = []string{"-G", "Unix Makefiles"}
		}
	}

	args = append(args,
		"-DLLVM_ENABLE_PROJECTS=clang",
		"-DLLVM_TARGETS_TO_BUILD=X86",
		"-DLLVM_EXPERIMENTAL_TARGETS_TO_BUILD=Cucaracha",
		fmt.Sprintf("-DCMAKE_BUILD_TYPE=%s", buildConfig),
		llvmPath,
	)

	cmd := exec.Command("cmake", args...)
	cmd.Dir = buildPath

	if verbose {
		fmt.Fprintf(os.Stderr, "Running: cmake %s\n", strings.Join(args, " "))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}

// runCMakeBuild runs CMake build
func (t *ClangToolchain) runCMakeBuild(buildPath, buildConfig string, verbose bool) error {
	args := []string{
		"--build", ".",
		"--target", "clang",
		"--config", buildConfig,
	}

	// Add parallel jobs
	args = append(args, "-j", fmt.Sprintf("%d", runtime.NumCPU()))

	cmd := exec.Command("cmake", args...)
	cmd.Dir = buildPath

	if verbose {
		fmt.Fprintf(os.Stderr, "Running: cmake %s\n", strings.Join(args, " "))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}

// IsSourceFile returns true if the file extension indicates a C/C++ source file
func IsSourceFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".c", ".cc", ".cpp", ".cxx", ".c++", ".m", ".mm":
		return true
	default:
		return false
	}
}

// IsCucarachaFile returns true if the file is a Cucaracha assembly or object file
func IsCucarachaFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".cucaracha", ".s", ".o":
		return true
	default:
		return false
	}
}

// ========== Legacy API (for backwards compatibility) ==========

// FindClang finds the full path to the clang binary in the system
func FindClang() (string, error) {
	toolchain, err := DiscoverClang(nil)
	if err != nil {
		return "", err
	}
	return toolchain.ClangPath(), nil
}

// CheckClang verifies a clang binary works
func CheckClang(path string) error {
	cmd := exec.Command(path, "--version")
	return cmd.Run()
}

// ClangDriver is a simple wrapper for clang operations (legacy API)
type ClangDriver struct {
	path string
}

// NewClangDriver creates a new clang driver
func NewClangDriver(paths ...string) (*ClangDriver, error) {
	var path string
	var err error

	if len(paths) > 0 && paths[0] != "" {
		path = paths[0]
	} else {
		path, err = FindClang()
		if err != nil {
			return nil, fmt.Errorf("clang binary not found: %w", err)
		}
	}

	if err := CheckClang(path); err != nil {
		return nil, fmt.Errorf("invalid clang binary '%s': %w", path, err)
	}

	return &ClangDriver{path: path}, nil
}

// Compile compiles a C source file into an object file (legacy API)
func (c *ClangDriver) Compile(source, output string) error {
	cmd := exec.Command(c.path, "-c", source, "-o", output)
	return cmd.Run()
}

// Link links a set of object files into an executable (legacy API)
func (c *ClangDriver) Link(objects []string, output string) error {
	args := []string{"-o", output}
	args = append(args, objects...)
	cmd := exec.Command(c.path, args...)
	return cmd.Run()
}

// CompileAndLink compiles a C source file into an executable (legacy API)
func (c *ClangDriver) CompileAndLink(source, output string) error {
	cmd := exec.Command(c.path, source, "-o", output)
	return cmd.Run()
}

// Version returns clang version (legacy API)
func (c *ClangDriver) Version() (string, error) {
	cmd := exec.Command(c.path, "--version")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
