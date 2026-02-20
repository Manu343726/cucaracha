package reflect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPackageResolverCreation(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	resolver, err := NewPackageResolver(cwd)
	if err != nil {
		t.Fatalf("Failed to create package resolver: %v", err)
	}

	if resolver == nil {
		t.Error("Package resolver is nil")
	}

	if resolver.moduleRoot == "" {
		t.Error("Module root is empty")
	}

	if resolver.gopath == "" {
		t.Error("GOPATH is empty")
	}
}

func TestResolvePackagePath(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	resolver, err := NewPackageResolver(cwd)
	if err != nil {
		t.Fatalf("Failed to create package resolver: %v", err)
	}

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
			if err != nil {
				if tt.wantPath {
					t.Errorf("Failed to resolve %q: %v", tt.importPath, err)
				}
				return
			}

			if path == "" {
				t.Errorf("Resolved path is empty for %q", tt.importPath)
				return
			}

			// Check if path is absolute
			if !filepath.IsAbs(path) {
				t.Errorf("Resolved path is not absolute for %q: %s", tt.importPath, path)
			}
		})
	}
}

func TestParsePackageFromImport(t *testing.T) {
	// Test parsing standard library
	pkg, err := ParsePackageFromImport("fmt")
	if err != nil {
		t.Errorf("Failed to parse fmt package: %v", err)
		return
	}

	if pkg == nil {
		t.Error("Parsed package is nil")
		return
	}

	if pkg.Name != "fmt" {
		t.Errorf("Expected package name 'fmt', got %q", pkg.Name)
	}
}

func TestFindGoMod(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	gomodPath, err := findGoMod(cwd)
	if err != nil {
		t.Fatalf("Failed to find go.mod: %v", err)
	}

	if gomodPath == "" {
		t.Error("go.mod path is empty")
	}

	// Check if file exists
	if _, err := os.Stat(gomodPath); err != nil {
		t.Errorf("go.mod file does not exist at %s: %v", gomodPath, err)
	}
}
