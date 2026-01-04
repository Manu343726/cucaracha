// Package llvm provides CMake build system integration for building LLVM/Clang.
package llvm

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// CMakeGenerator represents a CMake generator type
type CMakeGenerator int

const (
	// CMakeGeneratorAuto automatically selects the best available generator
	CMakeGeneratorAuto CMakeGenerator = iota
	// CMakeGeneratorNinja uses the Ninja build system
	CMakeGeneratorNinja
	// CMakeGeneratorMake uses Unix Makefiles
	CMakeGeneratorMake
	// CMakeGeneratorVS2022 uses Visual Studio 2022
	CMakeGeneratorVS2022
	// CMakeGeneratorVS2019 uses Visual Studio 2019
	CMakeGeneratorVS2019
)

// String returns the CMake generator string
func (g CMakeGenerator) String() string {
	switch g {
	case CMakeGeneratorNinja:
		return "Ninja"
	case CMakeGeneratorMake:
		return "Unix Makefiles"
	case CMakeGeneratorVS2022:
		return "Visual Studio 17 2022"
	case CMakeGeneratorVS2019:
		return "Visual Studio 16 2019"
	default:
		return ""
	}
}

// =============================================================================
// CMake Presets Support
// =============================================================================

// CMakePresetsFile represents the structure of a CMakePresets.json file
type CMakePresetsFile struct {
	Version              any               `json:"version"` // Can be int or object, we don't validate
	CMakeMinimumRequired *CMakeVersion     `json:"cmakeMinimumRequired,omitempty"`
	ConfigurePresets     []ConfigurePreset `json:"configurePresets,omitempty"`
	BuildPresets         []BuildPreset     `json:"buildPresets,omitempty"`
}

// CMakeVersion represents a CMake version requirement
type CMakeVersion struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
}

// PresetCondition represents a condition for when a preset applies
type PresetCondition struct {
	Type string `json:"type"`
	Lhs  string `json:"lhs"`
	Rhs  string `json:"rhs"`
}

// ConfigurePreset represents a CMake configure preset
type ConfigurePreset struct {
	Name           string           `json:"name"`
	Inherits       string           `json:"inherits,omitempty"`
	Hidden         bool             `json:"hidden,omitempty"`
	DisplayName    string           `json:"displayName,omitempty"`
	Description    string           `json:"description,omitempty"`
	Generator      string           `json:"generator,omitempty"`
	BinaryDir      string           `json:"binaryDir,omitempty"`
	CacheVariables map[string]any   `json:"cacheVariables,omitempty"`
	Condition      *PresetCondition `json:"condition,omitempty"`
}

// BuildPreset represents a CMake build preset
type BuildPreset struct {
	Name            string   `json:"name"`
	Inherits        string   `json:"inherits,omitempty"`
	Hidden          bool     `json:"hidden,omitempty"`
	ConfigurePreset string   `json:"configurePreset,omitempty"`
	Jobs            int      `json:"jobs,omitempty"`
	Targets         []string `json:"targets,omitempty"`
	Verbose         bool     `json:"verbose,omitempty"`
}

// LoadCMakePresets loads a CMakePresets.json file
func LoadCMakePresets(path string) (*CMakePresetsFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read presets file: %w", err)
	}

	var presets CMakePresetsFile
	if err := json.Unmarshal(data, &presets); err != nil {
		return nil, fmt.Errorf("failed to parse presets file: %w", err)
	}

	return &presets, nil
}

// FindConfigurePreset finds a configure preset by name
func (p *CMakePresetsFile) FindConfigurePreset(name string) *ConfigurePreset {
	for i := range p.ConfigurePresets {
		if p.ConfigurePresets[i].Name == name {
			return &p.ConfigurePresets[i]
		}
	}
	return nil
}

// FindBuildPreset finds a build preset by name
func (p *CMakePresetsFile) FindBuildPreset(name string) *BuildPreset {
	for i := range p.BuildPresets {
		if p.BuildPresets[i].Name == name {
			return &p.BuildPresets[i]
		}
	}
	return nil
}

// GetHostSystemName returns the CMake hostSystemName for the current platform
func GetHostSystemName() string {
	switch runtime.GOOS {
	case "windows":
		return "Windows"
	case "linux":
		return "Linux"
	case "darwin":
		return "Darwin"
	default:
		return runtime.GOOS
	}
}

// EvaluateCondition evaluates a preset condition for the current platform
func (c *PresetCondition) EvaluateCondition() bool {
	if c == nil {
		return true // No condition means always applicable
	}

	switch c.Type {
	case "equals":
		// Expand variables in lhs
		lhs := c.Lhs
		if strings.Contains(lhs, "${hostSystemName}") {
			lhs = strings.ReplaceAll(lhs, "${hostSystemName}", GetHostSystemName())
		}
		return lhs == c.Rhs
	case "notEquals":
		lhs := c.Lhs
		if strings.Contains(lhs, "${hostSystemName}") {
			lhs = strings.ReplaceAll(lhs, "${hostSystemName}", GetHostSystemName())
		}
		return lhs != c.Rhs
	default:
		return true // Unknown condition type, assume true
	}
}

// GetApplicableConfigurePresets returns all configure presets that apply to the current platform
func (p *CMakePresetsFile) GetApplicableConfigurePresets() []ConfigurePreset {
	var result []ConfigurePreset
	for _, preset := range p.ConfigurePresets {
		if preset.Hidden {
			continue
		}
		if preset.Condition != nil && !preset.Condition.EvaluateCondition() {
			continue
		}
		result = append(result, preset)
	}
	return result
}

// GetApplicableBuildPresets returns all build presets that apply to the current platform
func (p *CMakePresetsFile) GetApplicableBuildPresets() []BuildPreset {
	var result []BuildPreset
	for _, preset := range p.BuildPresets {
		if preset.Hidden {
			continue
		}
		// Check if the corresponding configure preset is applicable
		if preset.ConfigurePreset != "" {
			configPreset := p.FindConfigurePreset(preset.ConfigurePreset)
			if configPreset != nil && configPreset.Condition != nil && !configPreset.Condition.EvaluateCondition() {
				continue
			}
		}
		result = append(result, preset)
	}
	return result
}

// SelectBestPreset selects the best configure preset for the current platform
// It prefers presets with "Docker" in the name when running in a container,
// otherwise selects the first applicable preset
func (p *CMakePresetsFile) SelectBestPreset() *ConfigurePreset {
	applicable := p.GetApplicableConfigurePresets()
	if len(applicable) == 0 {
		return nil
	}

	// Check if we're in a Docker container
	inDocker := isInDockerContainer()

	// First pass: look for Docker-specific preset if in container
	if inDocker {
		for i := range applicable {
			if strings.Contains(strings.ToLower(applicable[i].Name), "docker") {
				return &applicable[i]
			}
		}
	}

	// Second pass: prefer non-Docker preset if not in container
	if !inDocker {
		for i := range applicable {
			if !strings.Contains(strings.ToLower(applicable[i].Name), "docker") {
				return &applicable[i]
			}
		}
	}

	// Fall back to first applicable preset
	return &applicable[0]
}

// isInDockerContainer checks if we're running inside a Docker container
func isInDockerContainer() bool {
	// Check for .dockerenv file
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Check cgroup for docker
	if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		if strings.Contains(string(data), "docker") {
			return true
		}
	}

	// Check for devcontainer environment
	if os.Getenv("REMOTE_CONTAINERS") != "" || os.Getenv("CODESPACES") != "" {
		return true
	}

	return false
}

// =============================================================================
// CMake Configuration
// =============================================================================

// CMakeConfig holds configuration for CMake operations
type CMakeConfig struct {
	// SourceDir is the path to the CMake source directory (contains CMakeLists.txt)
	SourceDir string

	// BuildDir is the path to the build directory
	BuildDir string

	// Generator specifies which CMake generator to use (default: auto-detect)
	Generator CMakeGenerator

	// BuildType is the CMake build type (Release, Debug, RelWithDebInfo, MinSizeRel)
	BuildType string

	// Variables is a map of CMake variables to set (-D options)
	Variables map[string]string

	// CacheVariables are variables that should be cached
	CacheVariables map[string]string

	// Targets specifies which targets to build (empty = all)
	Targets []string

	// ParallelJobs is the number of parallel build jobs (0 = auto)
	ParallelJobs int

	// Verbose enables verbose output
	Verbose bool

	// Stdout is where to write stdout (default: discard if not verbose)
	Stdout io.Writer

	// Stderr is where to write stderr (default: discard if not verbose)
	Stderr io.Writer
}

// DefaultCMakeConfig returns a default CMake configuration
func DefaultCMakeConfig() *CMakeConfig {
	return &CMakeConfig{
		Generator:    CMakeGeneratorAuto,
		BuildType:    "Release",
		Variables:    make(map[string]string),
		ParallelJobs: 0, // Auto-detect
	}
}

// CMake provides an interface to the CMake build system
type CMake struct {
	// Path to the cmake executable
	cmakePath string
}

// NewCMake creates a new CMake instance
// It searches for cmake in PATH or uses the provided path
func NewCMake(cmakePath ...string) (*CMake, error) {
	var path string

	if len(cmakePath) > 0 && cmakePath[0] != "" {
		path = cmakePath[0]
	} else {
		// Search in PATH
		var err error
		path, err = exec.LookPath("cmake")
		if err != nil {
			return nil, fmt.Errorf("cmake not found in PATH: %w", err)
		}
	}

	// Verify cmake works
	cmd := exec.Command(path, "--version")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("cmake at %s is not functional: %w", path, err)
	}

	return &CMake{cmakePath: path}, nil
}

// Path returns the path to the cmake executable
func (c *CMake) Path() string {
	return c.cmakePath
}

// Version returns the CMake version string
func (c *CMake) Version() (string, error) {
	cmd := exec.Command(c.cmakePath, "--version")
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

// detectGenerator determines the best generator for the current platform
func (c *CMake) detectGenerator() CMakeGenerator {
	if runtime.GOOS == "windows" {
		// Prefer Ninja on Windows if available, otherwise VS2022
		if _, err := exec.LookPath("ninja"); err == nil {
			return CMakeGeneratorNinja
		}
		return CMakeGeneratorVS2022
	}

	// On Unix, prefer Ninja if available, otherwise Make
	if _, err := exec.LookPath("ninja"); err == nil {
		return CMakeGeneratorNinja
	}
	return CMakeGeneratorMake
}

// ConfigureResult contains the result of a CMake configure operation
type ConfigureResult struct {
	// Command is the full command that was executed
	Command string

	// CacheFile is the path to the generated CMakeCache.txt
	CacheFile string

	// Stdout contains the standard output
	Stdout string

	// Stderr contains the standard error
	Stderr string
}

// Configure runs CMake configuration
func (c *CMake) Configure(config *CMakeConfig) (*ConfigureResult, error) {
	if config == nil {
		config = DefaultCMakeConfig()
	}

	if config.SourceDir == "" {
		return nil, fmt.Errorf("source directory not specified")
	}

	if config.BuildDir == "" {
		return nil, fmt.Errorf("build directory not specified")
	}

	// Ensure source directory exists
	if _, err := os.Stat(config.SourceDir); err != nil {
		return nil, fmt.Errorf("source directory not found: %s", config.SourceDir)
	}

	// Create build directory if it doesn't exist
	if err := os.MkdirAll(config.BuildDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create build directory: %w", err)
	}

	// Build arguments
	args := []string{}

	// Add generator
	generator := config.Generator
	if generator == CMakeGeneratorAuto {
		generator = c.detectGenerator()
	}

	if genStr := generator.String(); genStr != "" {
		args = append(args, "-G", genStr)
		// Add architecture for Visual Studio generators
		if generator == CMakeGeneratorVS2022 || generator == CMakeGeneratorVS2019 {
			args = append(args, "-A", "x64")
		}
	}

	// Add build type
	if config.BuildType != "" {
		args = append(args, fmt.Sprintf("-DCMAKE_BUILD_TYPE=%s", config.BuildType))
	}

	// Add variables
	for key, value := range config.Variables {
		args = append(args, fmt.Sprintf("-D%s=%s", key, value))
	}

	// Add cache variables
	for key, value := range config.CacheVariables {
		args = append(args, fmt.Sprintf("-D%s:STRING=%s", key, value))
	}

	// Add source directory
	args = append(args, "-S", config.SourceDir)

	// Add build directory
	args = append(args, "-B", config.BuildDir)

	// Create command
	cmd := exec.Command(c.cmakePath, args...)

	result := &ConfigureResult{
		Command:   fmt.Sprintf("%s %s", c.cmakePath, strings.Join(args, " ")),
		CacheFile: filepath.Join(config.BuildDir, "CMakeCache.txt"),
	}

	// Set up output
	var stdoutBuf, stderrBuf strings.Builder
	if config.Verbose {
		cmd.Stdout = io.MultiWriter(os.Stderr, &stdoutBuf)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
		fmt.Fprintf(os.Stderr, "Running: %s\n", result.Command)
	} else {
		if config.Stdout != nil {
			cmd.Stdout = io.MultiWriter(config.Stdout, &stdoutBuf)
		} else {
			cmd.Stdout = &stdoutBuf
		}
		if config.Stderr != nil {
			cmd.Stderr = io.MultiWriter(config.Stderr, &stderrBuf)
		} else {
			cmd.Stderr = &stderrBuf
		}
	}

	// Execute
	if err := cmd.Run(); err != nil {
		result.Stdout = stdoutBuf.String()
		result.Stderr = stderrBuf.String()
		return result, fmt.Errorf("cmake configure failed: %w\n%s", err, result.Stderr)
	}

	result.Stdout = stdoutBuf.String()
	result.Stderr = stderrBuf.String()
	return result, nil
}

// BuildResult contains the result of a CMake build operation
type BuildResult struct {
	// Command is the full command that was executed
	Command string

	// Stdout contains the standard output
	Stdout string

	// Stderr contains the standard error
	Stderr string
}

// Build runs CMake build
func (c *CMake) Build(config *CMakeConfig) (*BuildResult, error) {
	if config == nil {
		config = DefaultCMakeConfig()
	}

	if config.BuildDir == "" {
		return nil, fmt.Errorf("build directory not specified")
	}

	// Ensure build directory exists
	if _, err := os.Stat(config.BuildDir); err != nil {
		return nil, fmt.Errorf("build directory not found: %s", config.BuildDir)
	}

	// Build arguments
	args := []string{
		"--build", config.BuildDir,
	}

	// Add config (for multi-config generators like VS)
	if config.BuildType != "" {
		args = append(args, "--config", config.BuildType)
	}

	// Add targets
	for _, target := range config.Targets {
		args = append(args, "--target", target)
	}

	// Add parallel jobs
	jobs := config.ParallelJobs
	if jobs <= 0 {
		jobs = runtime.NumCPU()
	}
	args = append(args, "-j", fmt.Sprintf("%d", jobs))

	// Create command
	cmd := exec.Command(c.cmakePath, args...)

	result := &BuildResult{
		Command: fmt.Sprintf("%s %s", c.cmakePath, strings.Join(args, " ")),
	}

	// Set up output
	var stdoutBuf, stderrBuf strings.Builder
	if config.Verbose {
		cmd.Stdout = io.MultiWriter(os.Stderr, &stdoutBuf)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
		fmt.Fprintf(os.Stderr, "Running: %s\n", result.Command)
	} else {
		if config.Stdout != nil {
			cmd.Stdout = io.MultiWriter(config.Stdout, &stdoutBuf)
		} else {
			cmd.Stdout = &stdoutBuf
		}
		if config.Stderr != nil {
			cmd.Stderr = io.MultiWriter(config.Stderr, &stderrBuf)
		} else {
			cmd.Stderr = &stderrBuf
		}
	}

	// Execute
	if err := cmd.Run(); err != nil {
		result.Stdout = stdoutBuf.String()
		result.Stderr = stderrBuf.String()
		return result, fmt.Errorf("cmake build failed: %w\n%s", err, result.Stderr)
	}

	result.Stdout = stdoutBuf.String()
	result.Stderr = stderrBuf.String()
	return result, nil
}

// IsConfigured checks if a build directory has been configured
func (c *CMake) IsConfigured(buildDir string) bool {
	cacheFile := filepath.Join(buildDir, "CMakeCache.txt")
	_, err := os.Stat(cacheFile)
	return err == nil
}

// Clean removes the build directory
func (c *CMake) Clean(buildDir string) error {
	return os.RemoveAll(buildDir)
}

// ConfigureAndBuild is a convenience method that configures and builds in one call
func (c *CMake) ConfigureAndBuild(config *CMakeConfig) (*BuildResult, error) {
	// Check if already configured
	if !c.IsConfigured(config.BuildDir) {
		if _, err := c.Configure(config); err != nil {
			return nil, err
		}
	}

	return c.Build(config)
}

// LLVMBuildConfig holds LLVM-specific build configuration
type LLVMBuildConfig struct {
	// LLVMRoot is the root of the llvm-project checkout
	LLVMRoot string

	// BuildDir is the build directory (default: llvm-project/build)
	BuildDir string

	// BuildType is the CMake build type (default: Release)
	BuildType string

	// EnableProjects is a list of LLVM projects to enable (default: clang)
	EnableProjects []string

	// TargetsToBuild is a list of targets to build (default: X86)
	TargetsToBuild []string

	// ExperimentalTargets is a list of experimental targets (default: Cucaracha)
	ExperimentalTargets []string

	// ExtraVariables is a map of additional CMake variables
	ExtraVariables map[string]string

	// Verbose enables verbose output
	Verbose bool
}

// DefaultLLVMBuildConfig returns default LLVM build configuration
func DefaultLLVMBuildConfig() *LLVMBuildConfig {
	return &LLVMBuildConfig{
		BuildType:           "Release",
		EnableProjects:      []string{"clang"},
		TargetsToBuild:      []string{"X86"},
		ExperimentalTargets: []string{"Cucaracha"},
	}
}

// BuildLLVM builds LLVM/Clang with Cucaracha support
// It automatically checks for CMakePresets.json and uses presets if available
func (c *CMake) BuildLLVM(config *LLVMBuildConfig) (*BuildResult, error) {
	if config == nil {
		config = DefaultLLVMBuildConfig()
	}

	if config.LLVMRoot == "" {
		return nil, fmt.Errorf("LLVM root not specified")
	}

	// Check llvm-project structure
	llvmDir := filepath.Join(config.LLVMRoot, "llvm")
	if _, err := os.Stat(llvmDir); err != nil {
		return nil, fmt.Errorf("llvm directory not found at %s", llvmDir)
	}

	// Check for CMakePresets.json
	presetsPath := filepath.Join(llvmDir, "CMakePresets.json")
	if _, err := os.Stat(presetsPath); err == nil {
		// Presets file exists, try to use it
		result, err := c.BuildLLVMWithPresets(config, presetsPath)
		if err == nil {
			return result, nil
		}
		// If presets failed, log and fall back to manual config
		if config.Verbose {
			fmt.Printf("Warning: Failed to build with presets: %v, falling back to manual configuration\n", err)
		}
	}

	// Fall back to manual configuration
	return c.buildLLVMManual(config, llvmDir)
}

// buildLLVMManual builds LLVM using manual CMake configuration (no presets)
func (c *CMake) buildLLVMManual(config *LLVMBuildConfig, llvmDir string) (*BuildResult, error) {
	// Determine build directory
	buildDir := config.BuildDir
	if buildDir == "" {
		if runtime.GOOS == "windows" {
			buildDir = filepath.Join(config.LLVMRoot, "build_vs2022")
		} else {
			buildDir = filepath.Join(config.LLVMRoot, "build")
		}
	}

	// Create CMake config
	cmakeConfig := &CMakeConfig{
		SourceDir:    llvmDir,
		BuildDir:     buildDir,
		BuildType:    config.BuildType,
		Variables:    make(map[string]string),
		Targets:      []string{"clang"},
		Verbose:      config.Verbose,
		ParallelJobs: 0, // Auto
	}

	// Set LLVM-specific variables
	if len(config.EnableProjects) > 0 {
		cmakeConfig.Variables["LLVM_ENABLE_PROJECTS"] = strings.Join(config.EnableProjects, ";")
	}

	if len(config.TargetsToBuild) > 0 {
		cmakeConfig.Variables["LLVM_TARGETS_TO_BUILD"] = strings.Join(config.TargetsToBuild, ";")
	}

	if len(config.ExperimentalTargets) > 0 {
		cmakeConfig.Variables["LLVM_EXPERIMENTAL_TARGETS_TO_BUILD"] = strings.Join(config.ExperimentalTargets, ";")
	}

	// Add extra variables
	for k, v := range config.ExtraVariables {
		cmakeConfig.Variables[k] = v
	}

	return c.ConfigureAndBuild(cmakeConfig)
}

// BuildLLVMWithPresets builds LLVM using CMake presets
func (c *CMake) BuildLLVMWithPresets(config *LLVMBuildConfig, presetsPath string) (*BuildResult, error) {
	presets, err := LoadCMakePresets(presetsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load presets: %w", err)
	}

	// Select the best preset for the current platform
	configurePreset := presets.SelectBestPreset()
	if configurePreset == nil {
		return nil, fmt.Errorf("no applicable configure preset found for this platform")
	}

	if config.Verbose {
		fmt.Printf("Using CMake preset: %s\n", configurePreset.Name)
	}

	// Find corresponding build preset
	buildPreset := presets.FindBuildPreset(configurePreset.Name)

	// Build using presets
	return c.ConfigureAndBuildWithPreset(config.LLVMRoot, configurePreset, buildPreset, config.Verbose)
}

// ConfigureAndBuildWithPreset configures and builds using CMake presets
func (c *CMake) ConfigureAndBuildWithPreset(llvmRoot string, configurePreset *ConfigurePreset, buildPreset *BuildPreset, verbose bool) (*BuildResult, error) {
	llvmDir := filepath.Join(llvmRoot, "llvm")

	// Step 1: Configure using preset
	if verbose {
		fmt.Printf("Configuring with preset: %s\n", configurePreset.Name)
	}

	configureArgs := []string{
		"--preset", configurePreset.Name,
	}

	cmd := exec.Command(c.cmakePath, configureArgs...)
	cmd.Dir = llvmDir

	var configureOutput strings.Builder
	if verbose {
		cmd.Stdout = io.MultiWriter(os.Stdout, &configureOutput)
		cmd.Stderr = io.MultiWriter(os.Stderr, &configureOutput)
	} else {
		cmd.Stdout = &configureOutput
		cmd.Stderr = &configureOutput
	}

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("configure failed: %w\nOutput: %s", err, configureOutput.String())
	}

	// Determine build directory from preset
	buildDir := resolveBinaryDir(configurePreset.BinaryDir, llvmDir)
	if buildDir == "" {
		buildDir = filepath.Join(llvmRoot, "build")
	}

	// Step 2: Build using preset if available, otherwise manual build
	if verbose {
		fmt.Println("Building clang...")
	}

	var buildArgs []string
	if buildPreset != nil {
		buildArgs = []string{
			"--build", "--preset", buildPreset.Name,
			"--target", "clang",
		}
	} else {
		// Fall back to manual build command
		buildArgs = []string{
			"--build", buildDir,
			"--target", "clang",
		}
		// Use parallel jobs based on CPU count
		buildArgs = append(buildArgs, "--parallel")
	}

	buildCmd := exec.Command(c.cmakePath, buildArgs...)
	buildCmd.Dir = llvmDir

	var buildOutput strings.Builder
	if verbose {
		buildCmd.Stdout = io.MultiWriter(os.Stdout, &buildOutput)
		buildCmd.Stderr = io.MultiWriter(os.Stderr, &buildOutput)
	} else {
		buildCmd.Stdout = &buildOutput
		buildCmd.Stderr = &buildOutput
	}

	if err := buildCmd.Run(); err != nil {
		return nil, fmt.Errorf("build failed: %w\nOutput: %s", err, buildOutput.String())
	}

	return &BuildResult{
		Command: fmt.Sprintf("%s %s && %s %s", c.cmakePath, strings.Join(configureArgs, " "), c.cmakePath, strings.Join(buildArgs, " ")),
		Stdout:  configureOutput.String() + "\n" + buildOutput.String(),
		Stderr:  "",
	}, nil
}

// resolveBinaryDir resolves CMake preset binaryDir variable substitution
func resolveBinaryDir(binaryDir string, sourceDir string) string {
	if binaryDir == "" {
		return ""
	}

	// Handle ${sourceDir} variable
	result := strings.ReplaceAll(binaryDir, "${sourceDir}", sourceDir)

	// Handle relative paths
	if !filepath.IsAbs(result) {
		result = filepath.Join(sourceDir, result)
	}

	// Clean the path
	result = filepath.Clean(result)

	return result
}

// findClangInBuildDir searches for clang executable in build directory
func findClangInBuildDir(buildDir string) string {
	candidates := []string{
		filepath.Join(buildDir, "bin", "clang"),
		filepath.Join(buildDir, "bin", "clang.exe"),
		filepath.Join(buildDir, "Debug", "bin", "clang.exe"),
		filepath.Join(buildDir, "Release", "bin", "clang.exe"),
		filepath.Join(buildDir, "RelWithDebInfo", "bin", "clang.exe"),
		filepath.Join(buildDir, "MinSizeRel", "bin", "clang.exe"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

// FindLLVMProject searches for llvm-project directory
// Returns the path if found, empty string otherwise
func FindLLVMProject(searchPaths ...string) string {
	// Default search paths
	if len(searchPaths) == 0 {
		// Add current directory and its parents
		if cwd, err := os.Getwd(); err == nil {
			searchPaths = append(searchPaths, cwd)
			searchPaths = append(searchPaths, filepath.Dir(cwd))
			searchPaths = append(searchPaths, filepath.Dir(filepath.Dir(cwd)))
		}

		// Add executable directory and its parents
		if exe, err := os.Executable(); err == nil {
			exeDir := filepath.Dir(exe)
			searchPaths = append(searchPaths, exeDir)
			searchPaths = append(searchPaths, filepath.Dir(exeDir))
		}

		// Common locations
		searchPaths = append(searchPaths, "/workspaces")
	}

	// Look for llvm-project in search paths
	for _, searchPath := range searchPaths {
		candidates := []string{
			filepath.Join(searchPath, "llvm-project"),
			filepath.Join(searchPath, "..", "llvm-project"),
		}

		for _, candidate := range candidates {
			absPath, err := filepath.Abs(candidate)
			if err != nil {
				continue
			}

			// Check if it looks like llvm-project
			llvmDir := filepath.Join(absPath, "llvm")
			if _, err := os.Stat(llvmDir); err == nil {
				return absPath
			}
		}
	}

	return ""
}
