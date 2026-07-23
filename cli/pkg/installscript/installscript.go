package installscript

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

const (
	cdnBaseURL     = "https://cdn.rilldata.com/rill"
	releaseBaseURL = "https://raw.githubusercontent.com/rilldata/rill"
	maxScriptBytes = int64(1 << 20)
)

var errUnsafeRedirect = errors.New("refusing install script redirect to a different origin")

type processRunner func(ctx context.Context, executable string, args ...string) error

type installer struct {
	client         *http.Client
	cdnBaseURL     string
	releaseBaseURL string
	tempDir        string
	run            processRunner
}

var defaultInstaller = installer{
	client:         http.DefaultClient,
	cdnBaseURL:     cdnBaseURL,
	releaseBaseURL: releaseBaseURL,
	run:            runProcess,
}

func Install(ctx context.Context, version string) error {
	return defaultInstaller.install(ctx, version)
}

func Uninstall(ctx context.Context) error {
	return defaultInstaller.uninstall(ctx)
}

func (i installer) install(ctx context.Context, version string) error {
	return i.execScript(ctx, version, "--version", version)
}

func (i installer) uninstall(ctx context.Context) error {
	return i.execScript(ctx, "", "--uninstall")
}

func (i installer) execScript(ctx context.Context, version string, args ...string) error {
	script, err := i.createScriptFile(ctx, version)
	if err != nil {
		return err
	}
	defer os.Remove(script)

	if err := ctx.Err(); err != nil {
		return err
	}

	// Execution invariant: the downloaded file is only passed to /bin/sh, followed by the caller's exact arguments.
	scriptArgs := make([]string, 0, len(args)+1)
	scriptArgs = append(scriptArgs, script)
	scriptArgs = append(scriptArgs, args...)
	if err := i.run(ctx, "/bin/sh", scriptArgs...); err != nil {
		// Cancellation invariant: one context controls download and execution, and cancellation wins over a process error.
		if ctxErr := ctx.Err(); ctxErr != nil {
			return fmt.Errorf("install script canceled: %w", ctxErr)
		}
		return fmt.Errorf("install script failed: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("install script canceled: %w", err)
	}
	return nil
}

func (i installer) createScriptFile(ctx context.Context, version string) (string, error) {
	scriptURL, err := i.scriptURL(version)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, scriptURL, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Download invariant: only a complete, bounded HTTP 200 response from the selected release URL can become executable input.
	client := i.redirectSafeClient()
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download install script from %s: %w", scriptURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download install script from %s: HTTP %d", scriptURL, resp.StatusCode)
	}
	if resp.ContentLength > maxScriptBytes {
		return "", fmt.Errorf("failed to download install script from %s: response exceeds %d bytes", scriptURL, maxScriptBytes)
	}

	// Integrity invariant: this package defines no detached checksum contract for these URLs, so partial or oversized transfers fail closed.
	contents, err := io.ReadAll(io.LimitReader(resp.Body, maxScriptBytes+1))
	if err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return "", fmt.Errorf("failed to download install script from %s: truncated response: %w", scriptURL, err)
		}
		return "", fmt.Errorf("failed to read install script from %s: %w", scriptURL, err)
	}
	if int64(len(contents)) > maxScriptBytes {
		return "", fmt.Errorf("failed to download install script from %s: response exceeds %d bytes", scriptURL, maxScriptBytes)
	}
	if resp.ContentLength >= 0 && int64(len(contents)) != resp.ContentLength {
		return "", fmt.Errorf("failed to download install script from %s: truncated response (expected %d bytes, got %d)", scriptURL, resp.ContentLength, len(contents))
	}
	if len(contents) == 0 {
		return "", fmt.Errorf("failed to download install script from %s: empty response", scriptURL)
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	// Temp-file invariant: CreateTemp provides mode 0600, and every write or close failure removes the file.
	out, err := os.CreateTemp(i.tempDir, "rill-install-*.sh")
	if err != nil {
		return "", fmt.Errorf("failed to create install script file: %w", err)
	}
	path := out.Name()
	keep := false
	defer func() {
		_ = out.Close()
		if !keep {
			_ = os.Remove(path)
		}
	}()

	if _, err := out.Write(contents); err != nil {
		return "", fmt.Errorf("failed to write install script file: %w", err)
	}
	if err := out.Close(); err != nil {
		return "", fmt.Errorf("failed to close install script file: %w", err)
	}
	keep = true
	return path, nil
}

func (i installer) scriptURL(version string) (string, error) {
	baseURL := i.cdnBaseURL
	path := []string{"install.sh"}
	switch version {
	case "nightly":
		path = []string{"nightly", "install.sh"}
	case "latest", "":
	default:
		if version == "." || version == ".." || strings.ContainsAny(version, "/\\") {
			return "", fmt.Errorf("invalid release version %q", version)
		}
		baseURL = i.releaseBaseURL
		path = []string{version, "scripts", "install.sh"}
	}

	u, err := url.JoinPath(baseURL, path...)
	if err != nil {
		return "", fmt.Errorf("failed to construct install script URL: %w", err)
	}
	parsed, err := url.Parse(u)
	if err != nil || !parsed.IsAbs() || parsed.Host == "" {
		return "", fmt.Errorf("failed to construct install script URL from %q", baseURL)
	}
	return u, nil
}

func (i installer) redirectSafeClient() *http.Client {
	client := http.Client{}
	if i.client != nil {
		client = *i.client
	}
	previousCheck := client.CheckRedirect
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		// Redirect invariant: a redirect may change the path, but not the scheme, host, or effective port.
		if len(via) > 0 && !sameOrigin(via[0].URL, req.URL) {
			return fmt.Errorf("%w: %s", errUnsafeRedirect, req.URL.Redacted())
		}
		if previousCheck != nil {
			return previousCheck(req, via)
		}
		return nil
	}
	return &client
}

func sameOrigin(a, b *url.URL) bool {
	return strings.EqualFold(a.Scheme, b.Scheme) &&
		strings.EqualFold(a.Hostname(), b.Hostname()) &&
		effectivePort(a) == effectivePort(b)
}

func effectivePort(u *url.URL) string {
	if port := u.Port(); port != "" {
		return port
	}
	switch strings.ToLower(u.Scheme) {
	case "http":
		return "80"
	case "https":
		return "443"
	default:
		return ""
	}
}

func runProcess(ctx context.Context, executable string, args ...string) error {
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
