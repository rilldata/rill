package archive

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDownloadCleanPreservesProjectOnFailure(t *testing.T) {
	// clean=true must behave transactionally: any pre-install failure preserves the active project.
	t.Run("HTTP transport failure", func(t *testing.T) {
		// A request that never produces a response must leave the live project and temp path untouched.
		root, projectPath, downloadPath := prepareExistingProject(t)
		transportErr := errors.New("transport unavailable")
		setDefaultHTTPClient(t, &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, transportErr
		})})

		err := Download(context.Background(), "http://archive.test/project.tar.gz", downloadPath, projectPath, true, false)
		if !errors.Is(err, transportErr) {
			t.Errorf("Download error = %v, want transport error", err)
		}
		assertExistingProjectPreserved(t, root, projectPath, downloadPath)
	})

	t.Run("non-200 response", func(t *testing.T) {
		// An HTTP error response is not an archive and must be rejected before replacement begins.
		root, projectPath, downloadPath := prepareExistingProject(t)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
		}))
		t.Cleanup(server.Close)

		err := Download(context.Background(), server.URL, downloadPath, projectPath, true, false)
		if err == nil {
			t.Error("Download accepted a non-200 response")
		}
		assertExistingProjectPreserved(t, root, projectPath, downloadPath)
	})

	t.Run("corrupt archive", func(t *testing.T) {
		// Gzip-compressed garbage reaches archive validation but must not delete the prior project.
		root, projectPath, downloadPath := prepareExistingProject(t)
		corruptBody := makeCorruptGzip(t)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write(corruptBody)
		}))
		t.Cleanup(server.Close)

		err := Download(context.Background(), server.URL, downloadPath, projectPath, true, false)
		if err == nil {
			t.Error("Download accepted a corrupt archive")
		}
		assertExistingProjectPreserved(t, root, projectPath, downloadPath)
	})

	t.Run("interrupted response body", func(t *testing.T) {
		// The synthetic body returns bytes and then an error, modeling a connection lost mid-download.
		root, projectPath, downloadPath := prepareExistingProject(t)
		bodyErr := errors.New("response body interrupted")
		setDefaultHTTPClient(t, &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Header:     make(http.Header),
				Body: io.NopCloser(io.MultiReader(
					strings.NewReader("partial archive bytes"),
					errorReader{err: bodyErr},
				)),
				Request: req,
			}, nil
		})})

		err := Download(context.Background(), "http://archive.test/project.tar.gz", downloadPath, projectPath, true, false)
		if !errors.Is(err, bodyErr) {
			t.Errorf("Download error = %v, want interrupted-body error", err)
		}
		assertExistingProjectPreserved(t, root, projectPath, downloadPath)
	})

	t.Run("context canceled while body is delayed", func(t *testing.T) {
		// Cancellation during a stalled response must abort before staging or replacing the live project.
		root, projectPath, downloadPath := prepareExistingProject(t)
		requestStarted := make(chan struct{})
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			close(requestStarted)
			<-r.Context().Done()
		}))
		t.Cleanup(server.Close)
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		go func() {
			done <- Download(ctx, server.URL, downloadPath, projectPath, true, false)
		}()
		<-requestStarted
		cancel()
		if err := <-done; !errors.Is(err, context.Canceled) {
			t.Errorf("Download error = %v, want context cancellation", err)
		}
		assertExistingProjectPreserved(t, root, projectPath, downloadPath)
	})

	t.Run("oversized response", func(t *testing.T) {
		// A declared oversized body is rejected before reading attacker-controlled data or touching the project.
		root, projectPath, downloadPath := prepareExistingProject(t)
		setDefaultHTTPClient(t, &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode:    http.StatusOK,
				Status:        "200 OK",
				Header:        make(http.Header),
				Body:          io.NopCloser(strings.NewReader("not read")),
				ContentLength: maxArchiveBytes + 1,
				Request:       req,
			}, nil
		})})

		err := Download(context.Background(), "http://archive.test/project.tar.gz", downloadPath, projectPath, true, false)
		if err == nil {
			t.Error("Download accepted an oversized archive")
		}
		assertExistingProjectPreserved(t, root, projectPath, downloadPath)
	})
}

func TestDownloadCleanReplacesProject(t *testing.T) {
	// A fully validated download should remove stale files and install the archive as the new project.
	_, projectPath, downloadPath := prepareExistingProject(t)
	archiveBody := makeTestArchive(t, []testTarEntry{
		{name: "bin/", typeflag: tar.TypeDir, mode: 0o700},
		{name: "bin/run.sh", typeflag: tar.TypeReg, mode: 0o751, body: []byte("#!/bin/sh\n")},
		{name: "rill.yaml", typeflag: tar.TypeReg, mode: 0o640, body: []byte("olap_connector: duckdb\n")},
	})
	serverURL := serveTestArchive(t, archiveBody)

	if err := Download(context.Background(), serverURL, downloadPath, projectPath, true, false); err != nil {
		t.Fatalf("Download returned an error: %v", err)
	}
	assertPathDoesNotExist(t, filepath.Join(projectPath, "keep.txt"))
	assertPathDoesNotExist(t, filepath.Join(projectPath, "state", "current.json"))
	assertFileContent(t, filepath.Join(projectPath, "bin", "run.sh"), "#!/bin/sh\n")
	assertFileContent(t, filepath.Join(projectPath, "rill.yaml"), "olap_connector: duckdb\n")
	assertPathDoesNotExist(t, downloadPath)
}

func TestDownloadPreservesArchivedModes(t *testing.T) {
	// Unix mode bits carry executable and access intent; platforms without them skip this contract.
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not preserve Unix archive permission bits")
	}

	root := t.TempDir()
	projectPath := filepath.Join(root, "project")
	downloadPath := filepath.Join(root, "download.tar.gz")
	// The explicit directory header isolates archived-mode restoration from implicit-parent creation.
	archiveBody := makeTestArchive(t, []testTarEntry{
		{name: "bin/", typeflag: tar.TypeDir, mode: 0o700},
		{name: "bin/run.sh", typeflag: tar.TypeReg, mode: 0o751, body: []byte("#!/bin/sh\n")},
		{name: "rill.yaml", typeflag: tar.TypeReg, mode: 0o640, body: []byte("olap_connector: duckdb\n")},
	})
	serverURL := serveTestArchive(t, archiveBody)

	if err := Download(context.Background(), serverURL, downloadPath, projectPath, true, false); err != nil {
		t.Fatalf("Download returned an error: %v", err)
	}
	assertFileMode(t, filepath.Join(projectPath, "bin"), 0o700)
	assertFileMode(t, filepath.Join(projectPath, "bin", "run.sh"), 0o751)
	assertFileMode(t, filepath.Join(projectPath, "rill.yaml"), 0o640)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type errorReader struct {
	err error
}

func (r errorReader) Read([]byte) (int, error) {
	return 0, r.err
}

func setDefaultHTTPClient(t *testing.T, client *http.Client) {
	t.Helper()

	previous := http.DefaultClient
	http.DefaultClient = client
	t.Cleanup(func() {
		http.DefaultClient = previous
	})
}

func serveTestArchive(t *testing.T, body []byte) string {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(body)
	}))
	t.Cleanup(server.Close)
	return server.URL
}

func prepareExistingProject(t *testing.T) (root, projectPath, downloadPath string) {
	t.Helper()

	root = t.TempDir()
	projectPath = filepath.Join(root, "project")
	downloadPath = filepath.Join(root, "download.tar.gz")
	if err := os.MkdirAll(filepath.Join(projectPath, "state"), 0o750); err != nil {
		t.Fatalf("create existing project: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectPath, "keep.txt"), []byte("keep me"), 0o640); err != nil {
		t.Fatalf("write existing project file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectPath, "state", "current.json"), []byte("{\"version\":1}"), 0o600); err != nil {
		t.Fatalf("write existing nested project file: %v", err)
	}
	return root, projectPath, downloadPath
}

func assertExistingProjectPreserved(t *testing.T, root, projectPath, downloadPath string) {
	t.Helper()

	assertFileContent(t, filepath.Join(projectPath, "keep.txt"), "keep me")
	assertFileContent(t, filepath.Join(projectPath, "state", "current.json"), "{\"version\":1}")
	assertPathDoesNotExist(t, filepath.Join(root, "unexpected.txt"))
	assertPathDoesNotExist(t, downloadPath)
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()

	got, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("read %q: %v", path, err)
		return
	}
	if string(got) != want {
		t.Errorf("content of %q = %q, want %q", path, got, want)
	}
}

func assertFileMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Errorf("stat %q: %v", path, err)
		return
	}
	if got := info.Mode().Perm(); got != want {
		t.Errorf("mode of %q = %#o, want %#o", path, got, want)
	}
}

func makeCorruptGzip(t *testing.T) []byte {
	t.Helper()

	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	if _, err := writer.Write([]byte("this is not a tar stream")); err != nil {
		t.Fatalf("write corrupt gzip fixture: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close corrupt gzip fixture: %v", err)
	}
	return buf.Bytes()
}
