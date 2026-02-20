package reflect

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PackageResolver handles resolution of Go import paths to filesystem locations
type PackageResolver struct {
	gomodPath  string // Path to go.mod file
	moduleRoot string // Root directory of the Go module
	gopath     string // GOPATH environment variable
	cacheDir   string // Directory where the resolver is invoked from
}

// NewPackageResolver creates a new package resolver for a given directory
func NewPackageResolver(workDir string) (*PackageResolver, error) {
	gomodPath, err := findGoMod(workDir)
	if err != nil {
		return nil, fmt.Errorf("failed to find go.mod: %w", err)
	}

	moduleRoot := filepath.Dir(gomodPath)

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		gopath = filepath.Join(home, "go")
	}

	return &PackageResolver{
		gomodPath:  gomodPath,
		moduleRoot: moduleRoot,
		gopath:     gopath,
		cacheDir:   workDir,
	}, nil
}

// ResolvePackagePath resolves a Go import path to a filesystem path
// Examples:
//   - "fmt" -> "/usr/lib/go/src/fmt"
//   - "github.com/user/repo" -> "/home/user/go/pkg/mod/github.com/user/repo@v1.0.0"
//   - "github.com/user/repo/subpkg" -> "/home/user/go/pkg/mod/github.com/user/repo@v1.0.0/subpkg"
func (pr *PackageResolver) ResolvePackagePath(importPath string) (string, error) {
	// Primary: Try to resolve using go list (handles stdlib, modules, and packages)
	if path, err := pr.resolveUsingGoList(importPath); err == nil && path != "" {
		return path, nil
	}

	// Secondary: Try to resolve as local package in the current module
	if path, err := pr.resolveLocal(importPath); err == nil {
		return path, nil
	}

	// Tertiary: Try to resolve from module cache
	if path, err := pr.resolveFromModuleCache(importPath); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("could not resolve package %q", importPath)
}

// resolveUsingGoList uses "go list" to resolve the package path
// It first tries "go list -m" for modules, then "go list -json" with Dir fallback
func (pr *PackageResolver) resolveUsingGoList(importPath string) (string, error) {
	// First try to use "go list -json" to get package information directly
	cmd := exec.Command("go", "list", "-json", importPath)
	cmd.Dir = pr.moduleRoot

	output, err := cmd.Output()
	if err == nil {
		var pkgInfo struct {
			Dir string `json:"Dir"`
		}

		if err := json.Unmarshal(output, &pkgInfo); err == nil && pkgInfo.Dir != "" {
			return pkgInfo.Dir, nil
		}
	}

	// Fall back to "go list -m" for module resolution
	cmd = exec.Command("go", "list", "-m", "-json", importPath)
	cmd.Dir = pr.moduleRoot

	output, err = cmd.Output()
	if err != nil {
		return "", err
	}

	var moduleInfo struct {
		Path      string `json:"Path"`
		Main      bool   `json:"Main"`
		Dir       string `json:"Dir"`
		GoMod     string `json:"GoMod"`
		GoVersion string `json:"GoVersion"`
	}

	if err := json.Unmarshal(output, &moduleInfo); err != nil {
		return "", fmt.Errorf("failed to unmarshal module info: %w", err)
	}

	if moduleInfo.Dir == "" {
		return "", fmt.Errorf("module directory not found")
	}

	// If this is the main module, the path is relative to the module root
	if moduleInfo.Main {
		// Calculate the relative path of the import within the module
		relPath := strings.TrimPrefix(importPath, moduleInfo.Path)
		relPath = strings.TrimPrefix(relPath, "/")
		if relPath != "" {
			return filepath.Join(moduleInfo.Dir, relPath), nil
		}
		return moduleInfo.Dir, nil
	}

	// For external modules, go list gives us the module directory
	// We need to append any sub-package path
	relPath := strings.TrimPrefix(importPath, moduleInfo.Path)
	relPath = strings.TrimPrefix(relPath, "/")
	if relPath != "" {
		return filepath.Join(moduleInfo.Dir, relPath), nil
	}

	return moduleInfo.Dir, nil
}

// resolveStdlib attempts to resolve standard library packages
func (pr *PackageResolver) resolveStdlib(importPath string) (string, error) {
	// Standard library packages should be found via go list
	cmd := exec.Command("go", "list", "-m", "-json", importPath)
	cmd.Dir = pr.moduleRoot

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	var moduleInfo struct {
		Dir string `json:"Dir"`
	}

	if err := json.Unmarshal(output, &moduleInfo); err != nil {
		return "", err
	}

	if moduleInfo.Dir == "" {
		return "", fmt.Errorf("standard library package %q not found", importPath)
	}

	return moduleInfo.Dir, nil
}

// resolveFromModuleCache tries to find the package in the Go module cache
func (pr *PackageResolver) resolveFromModuleCache(importPath string) (string, error) {
	moduleCache := filepath.Join(pr.gopath, "pkg", "mod")

	// Parse the import path to extract the module and version
	// For "github.com/user/repo/subpkg", we need to find "github.com/user/repo@version"
	parts := strings.Split(importPath, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid import path: %q", importPath)
	}

	// Try progressively longer prefixes to find the module
	for i := len(parts); i >= 2; i-- {
		modulePath := strings.Join(parts[:i], "/")
		moduleCachePath := filepath.Join(moduleCache, modulePath)

		// Check if this directory exists (might have version suffix)
		parentDir := filepath.Dir(moduleCachePath)
		entries, err := os.ReadDir(parentDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() && strings.HasPrefix(entry.Name(), filepath.Base(moduleCachePath)+"@") {
				fullModulePath := filepath.Join(parentDir, entry.Name())
				// Calculate remaining path
				remainingPath := strings.Join(parts[i:], "/")
				if remainingPath != "" {
					return filepath.Join(fullModulePath, remainingPath), nil
				}
				return fullModulePath, nil
			}
		}
	}

	return "", fmt.Errorf("could not find module in cache: %q", importPath)
}

// resolveLocal attempts to resolve the package as a local path within the current module
func (pr *PackageResolver) resolveLocal(importPath string) (string, error) {
	// Read go.mod to find the module name
	modContent, err := os.ReadFile(pr.gomodPath)
	if err != nil {
		return "", err
	}

	moduleNameLine := ""
	for _, line := range strings.Split(string(modContent), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			moduleNameLine = line
			break
		}
	}

	if moduleNameLine == "" {
		return "", fmt.Errorf("could not find module name in go.mod")
	}

	moduleName := strings.TrimPrefix(moduleNameLine, "module ")
	moduleName = strings.TrimSpace(moduleName)

	// If the import path starts with the module name, it's a local package
	if strings.HasPrefix(importPath, moduleName) {
		relPath := strings.TrimPrefix(importPath, moduleName)
		relPath = strings.TrimPrefix(relPath, "/")
		fullPath := filepath.Join(pr.moduleRoot, relPath)
		return fullPath, nil
	}

	return "", fmt.Errorf("import path %q does not belong to module %q", importPath, moduleName)
}

// findGoMod searches for go.mod file starting from the given directory and working upwards
func findGoMod(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}

	for {
		gomodPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(gomodPath); err == nil {
			return gomodPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("go.mod not found in %q or parent directories", startDir)
}

// ParsePackageFromImport parses a Go package using its import path
// Examples:
//   - reflect.ParsePackageFromImport("fmt")
//   - reflect.ParsePackageFromImport("github.com/user/repo/pkg")
//   - reflect.ParsePackageFromImport("github.com/Manu343726/cucaracha/pkg/hw")
func ParsePackageFromImport(importPath string) (*Package, error) {
	// Get the current working directory to initialize the resolver
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	return ParsePackageFromImportInDir(importPath, cwd)
}

// ParsePackageFromImportInDir parses a Go package using its import path, starting resolution from a specific directory
func ParsePackageFromImportInDir(importPath string, startDir string) (*Package, error) {
	// Create a package resolver
	resolver, err := NewPackageResolver(startDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create package resolver: %w", err)
	}

	// Resolve the import path to a filesystem path
	packagePath, err := resolver.ResolvePackagePath(importPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve import path %q: %w", importPath, err)
	}

	// Check if the resolved path exists
	if _, err := os.Stat(packagePath); err != nil {
		return nil, fmt.Errorf("resolved package path does not exist: %s: %w", packagePath, err)
	}

	// Parse the package using the existing ParsePackage function
	pkg, err := ParsePackage(packagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse package at %q: %w", packagePath, err)
	}

	return pkg, nil
}

// GetModuleInfo retrieves information about a module using "go list"
func GetModuleInfo(importPath string, workDir string) (map[string]interface{}, error) {
	cmd := exec.Command("go", "list", "-m", "-json", importPath)
	cmd.Dir = workDir

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get module info for %q: %w", importPath, err)
	}

	var moduleInfo map[string]interface{}
	if err := json.Unmarshal(output, &moduleInfo); err != nil {
		return nil, fmt.Errorf("failed to parse module info: %w", err)
	}

	return moduleInfo, nil
}
