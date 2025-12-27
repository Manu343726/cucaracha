package llvm

import (
	"fmt"
	"os/exec"
	"runtime"
)

// FindClang finds the full path to the clang binary in the system
func FindClang() (string, error) {
	var binaryName string
	if runtime.GOOS == "windows" {
		binaryName = "clang.exe"
	} else {
		binaryName = "clang"
	}

	path, err := exec.LookPath(binaryName)
	if err != nil {
		return "", err
	}
	return path, nil
}

func CheckClang(path string) error {
	cmd := exec.Command(path, "--version")
	return cmd.Run()
}

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

// Compile compiles a C source file into an object file
func (c *ClangDriver) Compile(source, output string) error {
	cmd := exec.Command(c.path, "-c", source, "-o", output)
	return cmd.Run()
}

// Link links a set of object files into an executable
func (c *ClangDriver) Link(objects []string, output string) error {
	args := []string{"-o", output}
	args = append(args, objects...)
	cmd := exec.Command(c.path, args...)
	return cmd.Run()
}

// CompileAndLink compiles a C source file into an executable
func (c *ClangDriver) CompileAndLink(source, output string) error {
	cmd := exec.Command(c.path, source, "-o", output)
	return cmd.Run()
}

// Returns clang version
func (c *ClangDriver) Version() (string, error) {
	cmd := exec.Command(c.path, "--version")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
