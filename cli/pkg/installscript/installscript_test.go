package installscript

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestScriptURL(t *testing.T) {
	// Version selection must resolve to the exact trusted release channel or tagged source URL.
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{name: "empty means latest", version: "", want: "https://cdn.rilldata.com/rill/install.sh"},
		{name: "latest", version: "latest", want: "https://cdn.rilldata.com/rill/install.sh"},
		{name: "nightly", version: "nightly", want: "https://cdn.rilldata.com/rill/nightly/install.sh"},
		{name: "tag", version: "v1.2.3", want: "https://raw.githubusercontent.com/rilldata/rill/v1.2.3/scripts/install.sh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := defaultInstaller.scriptURL(tt.version)
			if err != nil {
				t.Fatalf("scriptURL(%q) returned error: %v", tt.version, err)
			}
			if got != tt.want {
				t.Fatalf("scriptURL(%q) = %q, want %q", tt.version, got, tt.want)
			}
		})
	}
}

func TestInstallRejects404(t *testing.T) {
	// An HTTP error page is never valid executable input and must leave no temporary script behind.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	runnerCalled := false
	i := testInstaller(server, tempDir, func(context.Context, string, ...string) error {
		runnerCalled = true
		return nil
	})

	err := i.install(context.Background(), "latest")
	if err == nil || !strings.Contains(err.Error(), "HTTP 404") {
		t.Fatalf("install error = %v, want HTTP 404", err)
	}
	if runnerCalled {
		t.Fatal("runner was called for a 404 response")
	}
	assertDirEmpty(t, tempDir)
}

func TestInstallRejectsCrossHostRedirect(t *testing.T) {
	// Redirects cannot move an installer download to a different host controlled outside the selected channel.
	destinationHit := make(chan struct{}, 1)
	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		destinationHit <- struct{}{}
		_, _ = io.WriteString(w, "#!/bin/sh\n")
	}))
	defer destination.Close()

	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, destination.URL+"/install.sh", http.StatusFound)
	}))
	defer source.Close()

	tempDir := t.TempDir()
	i := testInstaller(source, tempDir, func(context.Context, string, ...string) error {
		t.Fatal("runner was called after an unsafe redirect")
		return nil
	})

	err := i.install(context.Background(), "latest")
	if !errors.Is(err, errUnsafeRedirect) {
		t.Fatalf("install error = %v, want errUnsafeRedirect", err)
	}
	select {
	case <-destinationHit:
		t.Fatal("redirect destination was contacted")
	default:
	}
	assertDirEmpty(t, tempDir)
}

func TestInstallRejectsOversizedResponse(t *testing.T) {
	// Enforce the response limit for both declared and chunked bodies before invoking the shell.
	tests := []struct {
		name    string
		handler http.HandlerFunc
	}{
		{
			name: "declared size",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Length", strconv.FormatInt(maxScriptBytes+1, 10))
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			name: "streamed size",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.(http.Flusher).Flush()
				_, _ = w.Write(bytes.Repeat([]byte("x"), int(maxScriptBytes)+1))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			tempDir := t.TempDir()
			runnerCalled := false
			i := testInstaller(server, tempDir, func(context.Context, string, ...string) error {
				runnerCalled = true
				return nil
			})

			err := i.install(context.Background(), "latest")
			if err == nil || !strings.Contains(err.Error(), "response exceeds") {
				t.Fatalf("install error = %v, want oversized response error", err)
			}
			if runnerCalled {
				t.Fatal("runner was called for an oversized response")
			}
			assertDirEmpty(t, tempDir)
		})
	}
}

func TestInstallFailsClosedOnTruncatedResponse(t *testing.T) {
	// A short transfer must be detected before incomplete remote content can be executed.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "100")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "#!/bin/sh\n")
	}))
	defer server.Close()

	tempDir := t.TempDir()
	runnerCalled := false
	i := testInstaller(server, tempDir, func(context.Context, string, ...string) error {
		runnerCalled = true
		return nil
	})

	// This package's release URL contract has no detached checksum; transfer truncation must therefore never reach execution.
	err := i.install(context.Background(), "latest")
	if err == nil || !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("install error = %v, want unexpected EOF", err)
	}
	if runnerCalled {
		t.Fatal("runner was called for a truncated response")
	}
	assertDirEmpty(t, tempDir)
}

func TestInstallCancellation(t *testing.T) {
	// Cancellation must interrupt either network transfer or process execution and always clean temporary files.
	t.Run("during download", func(t *testing.T) {
		started := make(chan struct{})
		server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			close(started)
			<-r.Context().Done()
		}))
		defer server.Close()

		tempDir := t.TempDir()
		runnerCalled := false
		i := testInstaller(server, tempDir, func(context.Context, string, ...string) error {
			runnerCalled = true
			return nil
		})
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		go func() {
			done <- i.install(ctx, "latest")
		}()

		<-started
		cancel()
		select {
		case err := <-done:
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("install error = %v, want context.Canceled", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("install did not stop after cancellation")
		}
		if runnerCalled {
			t.Fatal("runner was called after download cancellation")
		}
		assertDirEmpty(t, tempDir)
	})

	t.Run("during execution", func(t *testing.T) {
		server := scriptServer("#!/bin/sh\n")
		defer server.Close()

		tempDir := t.TempDir()
		ctx, cancel := context.WithCancel(context.Background())
		var scriptPath string
		i := testInstaller(server, tempDir, func(runCtx context.Context, _ string, args ...string) error {
			scriptPath = args[0]
			cancel()
			<-runCtx.Done()
			return errors.New("process stopped")
		})

		err := i.install(ctx, "latest")
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("install error = %v, want context.Canceled", err)
		}
		if _, statErr := os.Stat(scriptPath); !errors.Is(statErr, os.ErrNotExist) {
			t.Fatalf("temporary script still exists after cancellation: %v", statErr)
		}
		assertDirEmpty(t, tempDir)
	})
}

func TestInstallReturnsNonzeroScriptExitAndCleansUp(t *testing.T) {
	// Preserve the script's nonzero exit status for callers while still deleting the downloaded file.
	server := scriptServer("#!/bin/sh\nexit 7\n")
	defer server.Close()

	tempDir := t.TempDir()
	i := testInstaller(server, tempDir, runProcess)
	err := i.install(context.Background(), "latest")
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("install error = %v, want exec.ExitError", err)
	}
	if exitErr.ExitCode() != 7 {
		t.Fatalf("exit code = %d, want 7", exitErr.ExitCode())
	}
	assertDirEmpty(t, tempDir)
}

func TestInstallPassesExactArgumentsAndCleansUp(t *testing.T) {
	// A tagged install executes the downloaded script once with the exact version arguments and restrictive mode.
	requestedPath := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath <- r.URL.Path
		_, _ = io.WriteString(w, "#!/bin/sh\n")
	}))
	defer server.Close()

	tempDir := t.TempDir()
	var executable string
	var args []string
	i := testInstaller(server, tempDir, func(_ context.Context, gotExecutable string, gotArgs ...string) error {
		executable = gotExecutable
		args = append([]string(nil), gotArgs...)
		contents, err := os.ReadFile(gotArgs[0])
		if err != nil {
			t.Fatalf("read temporary script: %v", err)
		}
		if string(contents) != "#!/bin/sh\n" {
			t.Fatalf("temporary script contents = %q", contents)
		}
		info, err := os.Stat(gotArgs[0])
		if err != nil {
			t.Fatalf("stat temporary script: %v", err)
		}
		if got := info.Mode().Perm(); got != 0o600 {
			t.Fatalf("temporary script mode = %o, want 600", got)
		}
		return nil
	})

	err := i.install(context.Background(), "v1.2.3")
	if err != nil {
		t.Fatalf("install returned error: %v", err)
	}
	if got := <-requestedPath; got != "/v1.2.3/scripts/install.sh" {
		t.Fatalf("request path = %q, want tag install script path", got)
	}
	if executable != "/bin/sh" {
		t.Fatalf("executable = %q, want /bin/sh", executable)
	}
	if len(args) != 3 || args[1] != "--version" || args[2] != "v1.2.3" {
		t.Fatalf("runner args = %q, want [script --version v1.2.3]", args)
	}
	if filepath.Dir(args[0]) != tempDir {
		t.Fatalf("script path = %q, want file in %q", args[0], tempDir)
	}
	info, err := os.Stat(args[0])
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("temporary script was not removed: %v", err)
	}
	if info != nil {
		t.Fatalf("temporary script still exists: %s", info.Name())
	}
	assertDirEmpty(t, tempDir)
}

func TestUninstallSucceedsAndCleansUp(t *testing.T) {
	// Uninstall uses the stable installer with only the uninstall flag and removes its temporary copy afterward.
	requestedPath := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath <- r.URL.Path
		_, _ = fmt.Fprint(w, "#!/bin/sh\n[ \"$#\" -eq 1 ] && [ \"$1\" = \"--uninstall\" ]\n")
	}))
	defer server.Close()

	tempDir := t.TempDir()
	i := testInstaller(server, tempDir, runProcess)
	if err := i.uninstall(context.Background()); err != nil {
		t.Fatalf("uninstall returned error: %v", err)
	}
	if got := <-requestedPath; got != "/install.sh" {
		t.Fatalf("request path = %q, want /install.sh", got)
	}
	assertDirEmpty(t, tempDir)
}

func testInstaller(server *httptest.Server, tempDir string, run processRunner) installer {
	return installer{
		client:         server.Client(),
		cdnBaseURL:     server.URL,
		releaseBaseURL: server.URL,
		tempDir:        tempDir,
		run:            run,
	}
}

func scriptServer(script string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, script)
	}))
}

func assertDirEmpty(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read temp directory: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("temporary files were not cleaned up: %v", entries)
	}
}
