package installscript

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

func Install(ctx context.Context, version string) error {
	return execScript(ctx, version, "--version", version)
}

func Uninstall(ctx context.Context) error {
	return execScript(ctx, "", "--uninstall")
}

func execScript(ctx context.Context, version string, args ...string) error {
	script, err := createScriptFile(ctx, version)
	if err != nil {
		return err
	}
	defer os.Remove(script)

	scriptArgs := append([]string{script}, args...)
	cmd := exec.CommandContext(ctx, "/bin/sh", scriptArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func createScriptFile(ctx context.Context, version string) (string, error) {
	if version == "" {
		version = "latest"
	}

	in, err := downloadScript(ctx, version)
	if err != nil {
		return "", err
	}

	out, err := os.CreateTemp("", "install*.sh")
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return "", err
	}

	return out.Name(), nil
}

func downloadScript(ctx context.Context, version string) (io.Reader, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/rilldata/rill/%s/scripts/install.sh", version)
	if version == "nightly" || version == "latest" {
		url = "https://raw.githubusercontent.com/rilldata/rill/main/scripts/install.sh"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download install script from %s: %w", url, err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("failed to download install script from %s: HTTP %d", url, resp.StatusCode)
	}

	return resp.Body, nil
}
