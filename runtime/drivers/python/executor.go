package python

import (
	"context"
	"encoding/json"
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
	// ConnectorEnvVars are env vars derived from connector configs (via env_from_connectors).
	// Keys are uppercase env var names, values are the config values.
	// JSON-object values are written to temp files; the env var is set to the file path.
	ConnectorEnvVars map[string]string
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

	// Process connector env vars: write JSON values to temp files, pass strings directly
	var tempFiles []string
	for envName, val := range opts.ConnectorEnvVars {
		if val == "" {
			continue
		}
		if looksLikeJSON(val) {
			// Write JSON credentials to a temp file so SDKs that expect file paths work
			tmpFile := filepath.Join(opts.TempDir, fmt.Sprintf("rill_creds_%s_%d.json", strings.ToLower(envName), time.Now().UnixNano()))
			if err := os.WriteFile(tmpFile, []byte(val), 0600); err != nil {
				return "", fmt.Errorf("python: failed to write credentials for %s: %w", envName, err)
			}
			tempFiles = append(tempFiles, tmpFile)
			cmd.Env = append(cmd.Env, envName+"="+tmpFile)
		} else {
			cmd.Env = append(cmd.Env, envName+"="+val)
		}
	}
	defer func() {
		for _, f := range tempFiles {
			os.Remove(f)
		}
	}()

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

// looksLikeJSON returns true if the string appears to be a JSON object.
func looksLikeJSON(s string) bool {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "{") {
		return false
	}
	return json.Valid([]byte(s))
}
