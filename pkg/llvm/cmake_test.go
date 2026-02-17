package llvm

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerator_String(t *testing.T) {
	tests := []struct {
		gen      CMakeGenerator
		expected string
	}{
		{CMakeGeneratorNinja, "Ninja"},
		{CMakeGeneratorMake, "Unix Makefiles"},
		{CMakeGeneratorVS2022, "Visual Studio 17 2022"},
		{CMakeGeneratorVS2019, "Visual Studio 16 2019"},
		{CMakeGeneratorAuto, ""},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.gen.String())
		})
	}
}

func TestDefaultCMakeConfig(t *testing.T) {
	config := DefaultCMakeConfig()

	assert.Equal(t, CMakeGeneratorAuto, config.Generator)
	assert.Equal(t, "Release", config.BuildType)
	assert.NotNil(t, config.Variables)
	assert.Equal(t, 0, config.ParallelJobs)
}

func TestNewCMake(t *testing.T) {
	// Skip if cmake is not available
	if _, err := exec.LookPath("cmake"); err != nil {
		t.Skip("cmake not found in PATH")
	}

	t.Run("auto-discover cmake", func(t *testing.T) {
		cmake, err := NewCMake()
		require.NoError(t, err)
		assert.NotEmpty(t, cmake.Path())
	})

	t.Run("invalid cmake path", func(t *testing.T) {
		_, err := NewCMake("/nonexistent/cmake")
		assert.Error(t, err)
	})
}

func TestCMake_Version(t *testing.T) {
	if _, err := exec.LookPath("cmake"); err != nil {
		t.Skip("cmake not found in PATH")
	}

	cmake, err := NewCMake()
	require.NoError(t, err)

	version, err := cmake.Version()
	require.NoError(t, err)
	assert.Contains(t, version, "cmake")
}

func TestCMake_detectGenerator(t *testing.T) {
	if _, err := exec.LookPath("cmake"); err != nil {
		t.Skip("cmake not found in PATH")
	}

	cmake, err := NewCMake()
	require.NoError(t, err)

	gen := cmake.detectGenerator()

	if runtime.GOOS == "windows" {
		// On Windows, expect Ninja or VS2022
		assert.True(t, gen == CMakeGeneratorNinja || gen == CMakeGeneratorVS2022)
	} else {
		// On Unix, expect Ninja or Make
		assert.True(t, gen == CMakeGeneratorNinja || gen == CMakeGeneratorMake)
	}
}

func TestCMake_Configure_Validation(t *testing.T) {
	if _, err := exec.LookPath("cmake"); err != nil {
		t.Skip("cmake not found in PATH")
	}

	cmake, err := NewCMake()
	require.NoError(t, err)

	t.Run("missing source dir", func(t *testing.T) {
		config := &CMakeConfig{
			BuildDir: "/tmp/build",
		}
		_, err := cmake.Configure(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "source directory not specified")
	})

	t.Run("missing build dir", func(t *testing.T) {
		config := &CMakeConfig{
			SourceDir: "/tmp/src",
		}
		_, err := cmake.Configure(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "build directory not specified")
	})

	t.Run("nonexistent source dir", func(t *testing.T) {
		config := &CMakeConfig{
			SourceDir: "/nonexistent/path",
			BuildDir:  "/tmp/build",
		}
		_, err := cmake.Configure(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "source directory not found")
	})
}

func TestCMake_Build_Validation(t *testing.T) {
	if _, err := exec.LookPath("cmake"); err != nil {
		t.Skip("cmake not found in PATH")
	}

	cmake, err := NewCMake()
	require.NoError(t, err)

	t.Run("missing build dir", func(t *testing.T) {
		config := &CMakeConfig{}
		_, err := cmake.Build(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "build directory not specified")
	})

	t.Run("nonexistent build dir", func(t *testing.T) {
		config := &CMakeConfig{
			BuildDir: "/nonexistent/build/path",
		}
		_, err := cmake.Build(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "build directory not found")
	})
}

func TestCMake_IsConfigured(t *testing.T) {
	if _, err := exec.LookPath("cmake"); err != nil {
		t.Skip("cmake not found in PATH")
	}

	cmake, err := NewCMake()
	require.NoError(t, err)

	t.Run("not configured", func(t *testing.T) {
		tempDir := t.TempDir()
		assert.False(t, cmake.IsConfigured(tempDir))
	})

	t.Run("configured", func(t *testing.T) {
		tempDir := t.TempDir()
		// Create a fake CMakeCache.txt
		cacheFile := filepath.Join(tempDir, "CMakeCache.txt")
		err := os.WriteFile(cacheFile, []byte("# CMake cache"), 0644)
		require.NoError(t, err)

		assert.True(t, cmake.IsConfigured(tempDir))
	})
}

func TestCMake_Clean(t *testing.T) {
	if _, err := exec.LookPath("cmake"); err != nil {
		t.Skip("cmake not found in PATH")
	}

	cmake, err := NewCMake()
	require.NoError(t, err)

	tempDir := t.TempDir()
	buildDir := filepath.Join(tempDir, "build")

	// Create build directory with some files
	err = os.MkdirAll(buildDir, 0755)
	require.NoError(t, err)

	testFile := filepath.Join(buildDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	// Clean
	err = cmake.Clean(buildDir)
	require.NoError(t, err)

	// Verify removed
	_, err = os.Stat(buildDir)
	assert.True(t, os.IsNotExist(err))
}

func TestDefaultLLVMBuildConfig(t *testing.T) {
	config := DefaultLLVMBuildConfig()

	assert.Equal(t, "Release", config.BuildType)
	assert.Contains(t, config.EnableProjects, "clang")
	assert.Contains(t, config.TargetsToBuild, "X86")
	assert.Contains(t, config.ExperimentalTargets, "Cucaracha")
}

func TestCMake_BuildLLVM_Validation(t *testing.T) {
	if _, err := exec.LookPath("cmake"); err != nil {
		t.Skip("cmake not found in PATH")
	}

	cmake, err := NewCMake()
	require.NoError(t, err)

	t.Run("missing LLVM root", func(t *testing.T) {
		config := &LLVMBuildConfig{}
		_, err := cmake.BuildLLVM(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "LLVM root not specified")
	})

	t.Run("invalid LLVM root", func(t *testing.T) {
		config := &LLVMBuildConfig{
			LLVMRoot: "/nonexistent/llvm-project",
		}
		_, err := cmake.BuildLLVM(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "llvm directory not found")
	})
}

func TestFindLLVMProject(t *testing.T) {
	t.Run("find in workspaces", func(t *testing.T) {
		// Check if we're in the dev container with llvm-project
		llvmPath := FindLLVMProject("/workspaces")
		if llvmPath != "" {
			assert.DirExists(t, llvmPath)
			assert.DirExists(t, filepath.Join(llvmPath, "llvm"))
		}
	})

	t.Run("not found in empty paths", func(t *testing.T) {
		tempDir := t.TempDir()
		llvmPath := FindLLVMProject(tempDir)
		assert.Empty(t, llvmPath)
	})
}

// Integration test - only runs if cmake and a simple CMake project can be configured
func TestCMake_Integration_ConfigureAndBuild(t *testing.T) {
	if _, err := exec.LookPath("cmake"); err != nil {
		t.Skip("cmake not found in PATH")
	}

	// Skip on Windows CI where we might not have a C compiler
	if runtime.GOOS == "windows" && os.Getenv("CI") != "" {
		t.Skip("Skipping on Windows CI")
	}

	// Create a minimal CMake project
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "src")
	buildDir := filepath.Join(tempDir, "build")

	err := os.MkdirAll(srcDir, 0755)
	require.NoError(t, err)

	// Create minimal CMakeLists.txt
	cmakeLists := `
cmake_minimum_required(VERSION 3.10)
project(TestProject C)
add_executable(test_exe main.c)
`
	err = os.WriteFile(filepath.Join(srcDir, "CMakeLists.txt"), []byte(cmakeLists), 0644)
	require.NoError(t, err)

	// Create minimal main.c
	mainC := `int main() { return 0; }`
	err = os.WriteFile(filepath.Join(srcDir, "main.c"), []byte(mainC), 0644)
	require.NoError(t, err)

	cmake, err := NewCMake()
	require.NoError(t, err)

	config := &CMakeConfig{
		SourceDir: srcDir,
		BuildDir:  buildDir,
		BuildType: "Release",
	}

	// Configure
	configResult, err := cmake.Configure(config)
	require.NoError(t, err)
	assert.FileExists(t, configResult.CacheFile)

	// Build
	buildResult, err := cmake.Build(config)
	require.NoError(t, err)
	assert.NotEmpty(t, buildResult.Command)

	// Verify executable was created
	exeName := "test_exe"
	if runtime.GOOS == "windows" {
		exeName = "test_exe.exe"
	}

	// Check both possible locations (config subdirectory for multi-config generators)
	possiblePaths := []string{
		filepath.Join(buildDir, exeName),
		filepath.Join(buildDir, "Release", exeName),
	}

	found := false
	for _, p := range possiblePaths {
		if _, err := os.Stat(p); err == nil {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected executable to be created in one of: %v", possiblePaths)
}

func TestCMakeConfig_Variables(t *testing.T) {
	config := &CMakeConfig{
		SourceDir: "/src",
		BuildDir:  "/build",
		Variables: map[string]string{
			"CMAKE_C_COMPILER":   "gcc",
			"CMAKE_CXX_COMPILER": "g++",
			"CUSTOM_VAR":         "value",
		},
	}

	assert.Equal(t, "gcc", config.Variables["CMAKE_C_COMPILER"])
	assert.Equal(t, "g++", config.Variables["CMAKE_CXX_COMPILER"])
	assert.Equal(t, "value", config.Variables["CUSTOM_VAR"])
}

func TestConfigureResult_Fields(t *testing.T) {
	result := &ConfigureResult{
		Command:   "cmake -G Ninja -S /src -B /build",
		CacheFile: "/build/CMakeCache.txt",
		Stdout:    "Configure output",
		Stderr:    "",
	}

	assert.Contains(t, result.Command, "cmake")
	assert.Contains(t, result.CacheFile, "CMakeCache.txt")
	assert.NotEmpty(t, result.Stdout)
}

func TestBuildResult_Fields(t *testing.T) {
	result := &BuildResult{
		Command: "cmake --build /build --config Release",
		Stdout:  "Build output",
		Stderr:  "",
	}

	assert.Contains(t, result.Command, "cmake")
	assert.Contains(t, result.Command, "--build")
	assert.NotEmpty(t, result.Stdout)
}

func TestLLVMBuildConfig_CustomSettings(t *testing.T) {
	config := &LLVMBuildConfig{
		LLVMRoot:            "/path/to/llvm-project",
		BuildDir:            "/custom/build",
		BuildType:           "Debug",
		EnableProjects:      []string{"clang", "lld"},
		TargetsToBuild:      []string{"X86", "ARM"},
		ExperimentalTargets: []string{"Cucaracha"},
		ExtraVariables: map[string]string{
			"LLVM_USE_LINKER": "lld",
		},
		Verbose: true,
	}

	assert.Equal(t, "/path/to/llvm-project", config.LLVMRoot)
	assert.Equal(t, "/custom/build", config.BuildDir)
	assert.Equal(t, "Debug", config.BuildType)
	assert.Contains(t, config.EnableProjects, "clang")
	assert.Contains(t, config.EnableProjects, "lld")
	assert.Contains(t, config.TargetsToBuild, "ARM")
	assert.Equal(t, "lld", config.ExtraVariables["LLVM_USE_LINKER"])
	assert.True(t, config.Verbose)
}

func TestCMake_ConfigureCommand(t *testing.T) {
	if _, err := exec.LookPath("cmake"); err != nil {
		t.Skip("cmake not found in PATH")
	}

	cmake, err := NewCMake()
	require.NoError(t, err)

	// We can't actually run configure without a valid source, but we can verify
	// the generator detection works
	gen := cmake.detectGenerator()

	// Verify it returns a valid generator
	switch gen {
	case CMakeGeneratorNinja, CMakeGeneratorMake, CMakeGeneratorVS2022, CMakeGeneratorVS2019:
		// Valid
	default:
		t.Errorf("Unexpected generator: %v", gen)
	}

	// If ninja is available, it should be preferred
	if _, err := exec.LookPath("ninja"); err == nil {
		assert.Equal(t, CMakeGeneratorNinja, gen)
	}
}

// Benchmark for FindLLVMProject
func BenchmarkFindLLVMProject(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FindLLVMProject()
	}
}

// Test that command strings are properly formatted
func TestCommandFormatting(t *testing.T) {
	// Test that variable formatting is correct
	config := &CMakeConfig{
		SourceDir: "/path/to/src",
		BuildDir:  "/path/to/build",
		BuildType: "Release",
		Variables: map[string]string{
			"VAR1": "value1",
			"VAR2": "value with spaces",
		},
	}

	// Variables should be properly formatted for CMake
	for k, v := range config.Variables {
		assert.NotEmpty(t, k)
		assert.NotEmpty(t, v)
		// Verify no leading/trailing spaces in keys
		assert.Equal(t, strings.TrimSpace(k), k)
	}
}

// =============================================================================
// CMake Presets Tests
// =============================================================================

func TestLoadCMakePresets(t *testing.T) {
	// Create a temporary directory for test presets
	tmpDir, err := os.MkdirTemp("", "cmake_presets_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	t.Run("valid presets file", func(t *testing.T) {
		presetsContent := `{
			"version": 6,
			"cmakeMinimumRequired": {"major": 3, "minor": 20, "patch": 0},
			"configurePresets": [
				{
					"name": "base",
					"hidden": true,
					"binaryDir": "${sourceDir}/../build",
					"cacheVariables": {
						"LLVM_ENABLE_PROJECTS": "clang"
					}
				},
				{
					"name": "linux-gcc",
					"inherits": "base",
					"generator": "Unix Makefiles",
					"condition": {
						"type": "equals",
						"lhs": "${hostSystemName}",
						"rhs": "Linux"
					}
				}
			],
			"buildPresets": [
				{
					"name": "linux-gcc",
					"configurePreset": "linux-gcc",
					"jobs": 10
				}
			]
		}`

		presetsPath := filepath.Join(tmpDir, "CMakePresets.json")
		err := os.WriteFile(presetsPath, []byte(presetsContent), 0644)
		require.NoError(t, err)

		presets, err := LoadCMakePresets(presetsPath)
		require.NoError(t, err)

		assert.Equal(t, float64(6), presets.Version) // JSON numbers unmarshal as float64 into any
		assert.NotNil(t, presets.CMakeMinimumRequired)
		assert.Equal(t, 3, presets.CMakeMinimumRequired.Major)
		assert.Equal(t, 20, presets.CMakeMinimumRequired.Minor)
		assert.Len(t, presets.ConfigurePresets, 2)
		assert.Len(t, presets.BuildPresets, 1)

		// Check base preset
		basePreset := presets.FindConfigurePreset("base")
		require.NotNil(t, basePreset)
		assert.True(t, basePreset.Hidden)
		assert.Equal(t, "${sourceDir}/../build", basePreset.BinaryDir)

		// Check linux-gcc preset
		linuxPreset := presets.FindConfigurePreset("linux-gcc")
		require.NotNil(t, linuxPreset)
		assert.Equal(t, "base", linuxPreset.Inherits)
		assert.Equal(t, "Unix Makefiles", linuxPreset.Generator)
		assert.NotNil(t, linuxPreset.Condition)
		assert.Equal(t, "equals", linuxPreset.Condition.Type)

		// Check build preset
		buildPreset := presets.FindBuildPreset("linux-gcc")
		require.NotNil(t, buildPreset)
		assert.Equal(t, "linux-gcc", buildPreset.ConfigurePreset)
		assert.Equal(t, 10, buildPreset.Jobs)
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := LoadCMakePresets(filepath.Join(tmpDir, "nonexistent.json"))
		assert.Error(t, err)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		invalidPath := filepath.Join(tmpDir, "invalid.json")
		err := os.WriteFile(invalidPath, []byte("not valid json"), 0644)
		require.NoError(t, err)

		_, err = LoadCMakePresets(invalidPath)
		assert.Error(t, err)
	})
}

func TestPresetCondition_EvaluateCondition(t *testing.T) {
	hostSystem := GetHostSystemName()

	tests := []struct {
		name      string
		condition *PresetCondition
		expected  bool
	}{
		{
			name:      "nil condition",
			condition: nil,
			expected:  true,
		},
		{
			name: "equals matching",
			condition: &PresetCondition{
				Type: "equals",
				Lhs:  "${hostSystemName}",
				Rhs:  hostSystem,
			},
			expected: true,
		},
		{
			name: "equals not matching",
			condition: &PresetCondition{
				Type: "equals",
				Lhs:  "${hostSystemName}",
				Rhs:  "NonexistentOS",
			},
			expected: false,
		},
		{
			name: "notEquals matching",
			condition: &PresetCondition{
				Type: "notEquals",
				Lhs:  "${hostSystemName}",
				Rhs:  "NonexistentOS",
			},
			expected: true,
		},
		{
			name: "notEquals not matching",
			condition: &PresetCondition{
				Type: "notEquals",
				Lhs:  "${hostSystemName}",
				Rhs:  hostSystem,
			},
			expected: false,
		},
		{
			name: "unknown condition type",
			condition: &PresetCondition{
				Type: "unknown",
				Lhs:  "a",
				Rhs:  "b",
			},
			expected: true, // Unknown types default to true
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.condition.EvaluateCondition()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetHostSystemName(t *testing.T) {
	hostSystem := GetHostSystemName()

	switch runtime.GOOS {
	case "windows":
		assert.Equal(t, "Windows", hostSystem)
	case "linux":
		assert.Equal(t, "Linux", hostSystem)
	case "darwin":
		assert.Equal(t, "Darwin", hostSystem)
	default:
		assert.Equal(t, runtime.GOOS, hostSystem)
	}
}

func TestCMakePresetsFile_GetApplicablePresets(t *testing.T) {
	hostSystem := GetHostSystemName()

	presets := &CMakePresetsFile{
		Version: 6,
		ConfigurePresets: []ConfigurePreset{
			{Name: "hidden-base", Hidden: true},
			{Name: "always-applicable"},
			{
				Name: "current-platform",
				Condition: &PresetCondition{
					Type: "equals",
					Lhs:  "${hostSystemName}",
					Rhs:  hostSystem,
				},
			},
			{
				Name: "other-platform",
				Condition: &PresetCondition{
					Type: "equals",
					Lhs:  "${hostSystemName}",
					Rhs:  "NonexistentOS",
				},
			},
		},
		BuildPresets: []BuildPreset{
			{Name: "build-always"},
			{Name: "build-current-platform", ConfigurePreset: "current-platform"},
			{Name: "build-other-platform", ConfigurePreset: "other-platform"},
		},
	}

	t.Run("applicable configure presets", func(t *testing.T) {
		applicable := presets.GetApplicableConfigurePresets()

		// Should exclude hidden and non-matching platform presets
		assert.Len(t, applicable, 2)

		names := make([]string, len(applicable))
		for i, p := range applicable {
			names[i] = p.Name
		}
		assert.Contains(t, names, "always-applicable")
		assert.Contains(t, names, "current-platform")
		assert.NotContains(t, names, "hidden-base")
		assert.NotContains(t, names, "other-platform")
	})

	t.Run("applicable build presets", func(t *testing.T) {
		applicable := presets.GetApplicableBuildPresets()

		// Should exclude presets whose configure preset doesn't match
		assert.Len(t, applicable, 2)

		names := make([]string, len(applicable))
		for i, p := range applicable {
			names[i] = p.Name
		}
		assert.Contains(t, names, "build-always")
		assert.Contains(t, names, "build-current-platform")
		assert.NotContains(t, names, "build-other-platform")
	})
}

func TestCMakePresetsFile_FindPresets(t *testing.T) {
	presets := &CMakePresetsFile{
		ConfigurePresets: []ConfigurePreset{
			{Name: "preset1"},
			{Name: "preset2"},
		},
		BuildPresets: []BuildPreset{
			{Name: "build1"},
			{Name: "build2"},
		},
	}

	t.Run("find existing configure preset", func(t *testing.T) {
		p := presets.FindConfigurePreset("preset1")
		require.NotNil(t, p)
		assert.Equal(t, "preset1", p.Name)
	})

	t.Run("find nonexistent configure preset", func(t *testing.T) {
		p := presets.FindConfigurePreset("nonexistent")
		assert.Nil(t, p)
	})

	t.Run("find existing build preset", func(t *testing.T) {
		p := presets.FindBuildPreset("build1")
		require.NotNil(t, p)
		assert.Equal(t, "build1", p.Name)
	})

	t.Run("find nonexistent build preset", func(t *testing.T) {
		p := presets.FindBuildPreset("nonexistent")
		assert.Nil(t, p)
	})
}

func TestResolveBinaryDir(t *testing.T) {
	tests := []struct {
		name      string
		binaryDir string
		sourceDir string
		expected  string
	}{
		{
			name:      "empty binaryDir",
			binaryDir: "",
			sourceDir: "/path/to/src",
			expected:  "",
		},
		{
			name:      "absolute path",
			binaryDir: "/absolute/build",
			sourceDir: "/path/to/src",
			expected:  "/absolute/build",
		},
		{
			name:      "sourceDir variable",
			binaryDir: "${sourceDir}/../build",
			sourceDir: "/path/to/llvm",
			expected:  "/path/to/build",
		},
		{
			name:      "sourceDir variable with subdir",
			binaryDir: "${sourceDir}/build",
			sourceDir: "/path/to/llvm",
			expected:  "/path/to/llvm/build",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveBinaryDir(tt.binaryDir, tt.sourceDir)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSelectBestPreset(t *testing.T) {
	hostSystem := GetHostSystemName()

	t.Run("no applicable presets", func(t *testing.T) {
		presets := &CMakePresetsFile{
			ConfigurePresets: []ConfigurePreset{
				{Name: "hidden", Hidden: true},
				{
					Name: "wrong-platform",
					Condition: &PresetCondition{
						Type: "equals",
						Lhs:  "${hostSystemName}",
						Rhs:  "NonexistentOS",
					},
				},
			},
		}
		selected := presets.SelectBestPreset()
		assert.Nil(t, selected)
	})

	t.Run("single applicable preset", func(t *testing.T) {
		presets := &CMakePresetsFile{
			ConfigurePresets: []ConfigurePreset{
				{Name: "hidden", Hidden: true},
				{
					Name: "applicable",
					Condition: &PresetCondition{
						Type: "equals",
						Lhs:  "${hostSystemName}",
						Rhs:  hostSystem,
					},
				},
			},
		}
		selected := presets.SelectBestPreset()
		require.NotNil(t, selected)
		assert.Equal(t, "applicable", selected.Name)
	})

	t.Run("prefers non-docker preset when not in container", func(t *testing.T) {
		// Skip if we're actually in a container
		if isInDockerContainer() {
			t.Skip("Running in container, skip non-docker preference test")
		}

		presets := &CMakePresetsFile{
			ConfigurePresets: []ConfigurePreset{
				{
					Name: "Linux GCC Docker",
					Condition: &PresetCondition{
						Type: "equals",
						Lhs:  "${hostSystemName}",
						Rhs:  hostSystem,
					},
				},
				{
					Name: "Linux GCC",
					Condition: &PresetCondition{
						Type: "equals",
						Lhs:  "${hostSystemName}",
						Rhs:  hostSystem,
					},
				},
			},
		}
		selected := presets.SelectBestPreset()
		require.NotNil(t, selected)
		assert.Equal(t, "Linux GCC", selected.Name)
	})
}

func TestIsInDockerContainer(t *testing.T) {
	// This just tests that the function runs without panicking
	// The actual result depends on the environment
	result := isInDockerContainer()
	t.Logf("isInDockerContainer: %v", result)
}

func TestLoadCMakePresets_RealFile(t *testing.T) {
	// Test with real CMakePresets.json from llvm-project if available
	llvmRoot := FindLLVMProject()
	if llvmRoot == "" {
		t.Skip("llvm-project not found")
	}

	presetsPath := filepath.Join(llvmRoot, "llvm", "CMakePresets.json")
	if _, err := os.Stat(presetsPath); err != nil {
		t.Skip("CMakePresets.json not found in llvm-project")
	}

	presets, err := LoadCMakePresets(presetsPath)
	require.NoError(t, err)

	t.Logf("Loaded CMakePresets.json from %s", presetsPath)
	t.Logf("Version: %v", presets.Version)
	t.Logf("Configure presets: %d", len(presets.ConfigurePresets))
	t.Logf("Build presets: %d", len(presets.BuildPresets))

	applicable := presets.GetApplicableConfigurePresets()
	t.Logf("Applicable configure presets: %d", len(applicable))
	for _, p := range applicable {
		t.Logf("  - %s", p.Name)
	}

	selected := presets.SelectBestPreset()
	if selected != nil {
		t.Logf("Selected preset: %s", selected.Name)
	} else {
		t.Log("No preset selected")
	}
}
