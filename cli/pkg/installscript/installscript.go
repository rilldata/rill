package installscript

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

func Install(ctx context.Context, version string, force bool) error {
	var env []string
	if force {
		env = append(os.Environ(), "RILL_INSTALL_FORCE=1")
	}
	return execScript(ctx, version, env, "--version", version)
}

func Uninstall(ctx context.Context) error {
	return execScript(ctx, "", nil, "--uninstall")
}

func execScript(ctx context.Context, version string, env []string, args ...string) error {
	script, err := createScriptFile(ctx, version)
	if err != nil {
		return err
	}
	defer os.Remove(script)

	scriptArgs := append([]string{script}, args...)
	cmd := exec.CommandContext(ctx, "/bin/sh", scriptArgs...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func createScriptFile(ctx context.Context, version string) (string, error) {
	var url string
	switch version {
	case "nightly":
		url = "https://cdn.rilldata.com/rill/nightly/install.sh"
	case "latest", "":
		url = "https://cdn.rilldata.com/rill/install.sh"
	default:
		url = fmt.Sprintf("https://raw.githubusercontent.com/rilldata/rill/%s/scripts/install.sh", version)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download install script from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download install script from %s: HTTP %d", url, resp.StatusCode)
	}

	out, err := os.CreateTemp("", "install*.sh")
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return out.Name(), nil
}
