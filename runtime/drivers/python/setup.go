package python

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PythonInfo describes a detected Python installation.
type PythonInfo struct {
	Found   bool
	Path    string
	Version string
}

// SetupOptions configures environment creation.
type SetupOptions struct {
	ProjectRoot string
	VenvPath    string   // defaults to ".rill/.venv"
	Packages    []string // pip packages to install (pandas and pyarrow are always included)
	PythonPath  string   // override auto-detect
}

// SetupResult describes the created environment.
type SetupResult struct {
	PythonPath string
	VenvPath   string
	Installed  []string
}

// basePackages are always installed in every Python environment.
var basePackages = []string{"pandas", "pyarrow"}

// DetectPython finds a Python installation on the system.
// If pythonPath is provided, it checks that specific path. Otherwise, it checks python3 then python.
func DetectPython(pythonPath string) (*PythonInfo, error) {
	candidates := []string{"python3", "python"}
	if pythonPath != "" {
		candidates = []string{pythonPath}
	}

	for _, candidate := range candidates {
		path, err := exec.LookPath(candidate)
		if err != nil {
			continue
		}

		// Get version
		out, err := exec.Command(path, "--version").CombinedOutput()
		if err != nil {
			continue
		}

		version := strings.TrimSpace(string(out))
		// "Python 3.11.0" → "3.11.0"
		version = strings.TrimPrefix(version, "Python ")

		return &PythonInfo{
			Found:   true,
			Path:    path,
			Version: version,
		}, nil
	}

	if pythonPath != "" {
		return &PythonInfo{Found: false}, fmt.Errorf("python not found at %q", pythonPath)
	}
	return &PythonInfo{Found: false}, fmt.Errorf("python not found; install Python 3.9+ or specify python_path in connector config")
}

// SetupEnvironment creates a virtual environment and installs packages.
// It uses DetectPython if opts.PythonPath is empty.
func SetupEnvironment(ctx context.Context, opts *SetupOptions) (*SetupResult, error) {
	// Detect Python
	info, err := DetectPython(opts.PythonPath)
	if err != nil {
		return nil, err
	}

	// Resolve venv path
	venvPath := opts.VenvPath
	if venvPath == "" {
		venvPath = ".rill/.venv"
	}
	if !filepath.IsAbs(venvPath) {
		venvPath = filepath.Join(opts.ProjectRoot, venvPath)
	}

	// Create venv
	cmd := exec.CommandContext(ctx, info.Path, "-m", "venv", venvPath)
	cmd.Dir = opts.ProjectRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to create venv at %s: %w\n%s", venvPath, err, string(out))
	}

	// Resolve pip path inside venv
	pipPath := filepath.Join(venvPath, "bin", "pip")
	if _, err := os.Stat(pipPath); err != nil {
		// Windows fallback
		pipPath = filepath.Join(venvPath, "Scripts", "pip.exe")
		if _, err := os.Stat(pipPath); err != nil {
			return nil, fmt.Errorf("pip not found in venv at %s", venvPath)
		}
	}

	// Merge base packages with requested packages (deduplicated)
	allPackages := deduplicatePackages(append(basePackages, opts.Packages...))

	// Install packages
	args := append([]string{"install"}, allPackages...)
	cmd = exec.CommandContext(ctx, pipPath, args...)
	cmd.Dir = opts.ProjectRoot
	out, err = cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to install packages: %w\n%s", err, string(out))
	}

	// Resolve Python path inside venv
	venvPythonPath := filepath.Join(venvPath, "bin", "python")
	if _, err := os.Stat(venvPythonPath); err != nil {
		venvPythonPath = filepath.Join(venvPath, "Scripts", "python.exe")
	}

	return &SetupResult{
		PythonPath: venvPythonPath,
		VenvPath:   venvPath,
		Installed:  allPackages,
	}, nil
}

func deduplicatePackages(pkgs []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, pkg := range pkgs {
		lower := strings.ToLower(strings.TrimSpace(pkg))
		if lower != "" && !seen[lower] {
			seen[lower] = true
			result = append(result, lower)
		}
	}
	return result
}
