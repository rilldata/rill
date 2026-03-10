package python

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rilldata/rill/runtime/pkg/fileutil"
)

// ExecuteOptions configures a Python script execution.
type ExecuteOptions struct {
	CodePath        string
	PythonPath      string
	RepoRoot        string
	AllowHostAccess bool
	TempDir         string
	Args            []string
	ExtraEnv        map[string]string
}

// ExecuteScript runs a Python script and returns the path to the output Parquet file.
// The script receives RILL_OUTPUT_PATH as an environment variable indicating where to write output.
func ExecuteScript(ctx context.Context, opts *ExecuteOptions) (string, error) {
	if opts.CodePath == "" {
		return "", fmt.Errorf("python: code_path is required")
	}

	// Resolve script path relative to repo root; prevents path traversal when AllowHostAccess is false
	scriptPath, err := fileutil.ResolveLocalPath(opts.CodePath, opts.RepoRoot, opts.AllowHostAccess)
	if err != nil {
		return "", fmt.Errorf("python: invalid code_path %q: %w", opts.CodePath, err)
	}

	// Verify the script exists
	if _, err := os.Stat(scriptPath); err != nil {
		return "", fmt.Errorf("python: script not found at %q: %w", opts.CodePath, err)
	}

	// Create a temp file for output
	outputPath := filepath.Join(opts.TempDir, fmt.Sprintf("rill_python_%d.parquet", time.Now().UnixNano()))

	// Resolve Python binary
	pythonPath := opts.PythonPath
	if pythonPath == "" {
		pythonPath = "python3"
	}

	// Build command
	args := append([]string{scriptPath}, opts.Args...)
	cmd := exec.CommandContext(ctx, pythonPath, args...)

	// Set working directory to repo root so relative imports work
	cmd.Dir = opts.RepoRoot

	// Build environment: inherit current env, add Rill-specific vars
	cmd.Env = append(os.Environ(),
		"RILL_OUTPUT_PATH="+outputPath,
		"RILL_REPO_ROOT="+opts.RepoRoot,
	)
	for k, v := range opts.ExtraEnv {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	// Capture output for error reporting
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute
	err = cmd.Run()
	if err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr == "" {
			stderrStr = "(no output)"
		}
		return "", fmt.Errorf("python: script %q failed: %w\nstderr:\n%s", opts.CodePath, err, stderrStr)
	}

	// Verify the script produced output
	if _, err := os.Stat(outputPath); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		return "", fmt.Errorf("python: script %q did not write output to RILL_OUTPUT_PATH (%s); stderr: %s",
			opts.CodePath, outputPath, stderrStr)
	}

	return outputPath, nil
}
