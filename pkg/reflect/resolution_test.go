package reflect

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackageResolverCreation(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err, "Failed to get working directory")

	resolver, err := NewPackageResolver(cwd)
	require.NoError(t, err, "Failed to create package resolver")

	assert.NotNil(t, resolver, "Package resolver should not be nil")
	assert.NotEmpty(t, resolver.moduleRoot, "Module root should not be empty")
	assert.NotEmpty(t, resolver.gopath, "GOPATH should not be empty")
}

func TestResolvePackagePath(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err, "Failed to get working directory")

	resolver, err := NewPackageResolver(cwd)
	require.NoError(t, err, "Failed to create package resolver")

	tests := []struct {
		importPath string
		wantPath   bool
	}{
		{"fmt", true},
		{"os", true},
		{"encoding/json", true},
		{"github.com/Manu343726/cucaracha/pkg/reflect", true},
		{"github.com/Manu343726/cucaracha/pkg/hw", true},
	}

	for _, tt := range tests {
		t.Run(tt.importPath, func(t *testing.T) {
			path, err := resolver.ResolvePackagePath(tt.importPath)
			if tt.wantPath {
				require.NoError(t, err, "Failed to resolve %s", tt.importPath)
			}

			assert.NotEmpty(t, path, "Resolved path should not be empty for %s", tt.importPath)
			assert.True(t, filepath.IsAbs(path), "Resolved path should be absolute for %s: %s", tt.importPath, path)
		})
	}
}

func TestParsePackageFromImport(t *testing.T) {
	// Test parsing standard library
	pkg, err := ParsePackageFromImport("fmt")
	require.NoError(t, err, "Failed to parse fmt package")
	require.NotNil(t, pkg, "Parsed package should not be nil")
	assert.Equal(t, "fmt", pkg.Name, "Package name should be 'fmt'")
}

func TestFindGoMod(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err, "Failed to get working directory")

	gomodPath, err := findGoMod(cwd)
	require.NoError(t, err, "Failed to find go.mod")
	require.NotEmpty(t, gomodPath, "go.mod path should not be empty")

	// Check if file exists
	_, err = os.Stat(gomodPath)
	assert.NoError(t, err, "go.mod file should exist at %s", gomodPath)
}
