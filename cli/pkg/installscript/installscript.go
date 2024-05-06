package installscript

import (
	"context"
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
)

//go:embed embed/*
var embedFS embed.FS

func Install(ctx context.Context, version string) error {
	if version != "" {
		return execScript(ctx, "--version", version)
	}
	return execScript(ctx)
}

func Uninstall(ctx context.Context) error {
	return execScript(ctx, "--uninstall")
}

func execScript(ctx context.Context, args ...string) error {
	script, err := createScriptFile()
	if err != nil {
		return err
	}
	defer os.Remove(script)

	// Execute the script with bash
	args = append([]string{script}, args...)
	cmd := exec.CommandContext(ctx, "/bin/bash", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func createScriptFile() (string, error) {
	// Open the embedded install script file
	in, err := embedFS.Open("embed/install.sh")
	if err != nil {
		return "", fmt.Errorf("install script not embedded (is this a dev build?): %w", err)
	}
	defer in.Close()

	// Write the install script to a temporary file
	out, err := os.CreateTemp("", "install*.sh")
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return "", err
	}

	// Return the temp script path
	return out.Name(), nil
}
