package installscript

import (
	"context"
	"embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

//go:embed embed/*
var embedFS embed.FS

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
	var in io.Reader
	var err error

	if version == "" {
		in, err = getEmbeddedScript()
	} else {
		// Download script for specific version
		in, err = downloadScript(ctx, version)
	}

	if err != nil {
		return "", err
	}

	// Write script to temporary file
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

func getEmbeddedScript() (io.Reader, error) {
	file, err := embedFS.Open("embed/install.sh")
	if err != nil {
		return nil, fmt.Errorf("install script not embedded (is this a dev build?): %w", err)
	}
	return file, nil
}

func downloadScript(ctx context.Context, version string) (io.Reader, error) {
	var url string
	if version == "nightly" {
		url = "https://raw.githubusercontent.com/rilldata/rill/main/scripts/install.sh"
	} else {
		url = fmt.Sprintf("https://raw.githubusercontent.com/rilldata/rill/%s/scripts/install.sh", version)
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
