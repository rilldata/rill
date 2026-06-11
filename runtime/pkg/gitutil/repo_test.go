package gitutil

import (
	"context"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnsureInit(t *testing.T) {
	ctx := context.Background()

	t.Run("initializes a fresh directory on the given branch", func(t *testing.T) {
		path := t.TempDir()
		require.NoError(t, EnsureInit(ctx, path, "deploybranch"))

		branch, err := CurrentBranch(ctx, path)
		require.NoError(t, err)
		require.Equal(t, "deploybranch", branch)
	})

	t.Run("is a no-op on an existing repository", func(t *testing.T) {
		path := setupTestRepository(t) // on branch main

		require.NoError(t, EnsureInit(ctx, path, "otherbranch"))

		branch, err := CurrentBranch(ctx, path)
		require.NoError(t, err)
		require.Equal(t, "main", branch, "EnsureInit must not switch branches of an existing repo")
	})
}

func TestCloneWithConfig(t *testing.T) {
	ctx := context.Background()
	remote := setupBareRemote(t)
	baseURL := serveRepoOverHTTP(t, remote)

	config := &Config{
		Remote:        baseURL,
		Username:      "x-access-token",
		Password:      "SECRETTOKEN123",
		DefaultBranch: "main",
		ManagedRepo:   true,
	}

	path := filepath.Join(t.TempDir(), "clone")
	require.NoError(t, CloneWithConfig(ctx, path, config))

	// the remote is named after the config and persists the clean URL
	url, err := Run(ctx, path, "remote", "get-url", "__rill_remote")
	require.NoError(t, err)
	require.Equal(t, baseURL, url)

	// the default branch is checked out
	branch, err := CurrentBranch(ctx, path)
	require.NoError(t, err)
	require.Equal(t, "main", branch)

	// the token must not appear anywhere in .git (config, logs/, FETCH_HEAD, ...)
	err = filepath.WalkDir(filepath.Join(path, ".git"), func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		require.NotContains(t, string(data), "SECRETTOKEN123", "credentials leaked into %s", p)
		return nil
	})
	require.NoError(t, err)
}

// setupBareRemote creates a bare repository seeded with one commit on `main`.
func setupBareRemote(t *testing.T) string {
	t.Helper()
	remote := filepath.Join(t.TempDir(), "remote.git")
	_, err := Run(context.Background(), "", "init", "--bare", remote)
	require.NoError(t, err)

	// seed the remote through a temporary clone
	seed := filepath.Join(t.TempDir(), "seed")
	require.NoError(t, Clone(context.Background(), seed, remote, "", false, false))
	setupGitConfig(t, seed)
	_, err = Run(context.Background(), seed, "checkout", "-b", "main")
	require.NoError(t, err)
	createCommit(t, seed, "seed.txt", "seed", "seed commit")
	require.NoError(t, execGit(seed, "push", "origin", "main"))
	// Point the bare repo's HEAD at main so later clones (e.g. createRemoteCommit) check out main
	// regardless of the environment's init.defaultBranch, which is master on CI runners. Otherwise
	// the dangling HEAD leaves clones on an unborn master and pushes land on master, not main.
	require.NoError(t, execGit(remote, "symbolic-ref", "HEAD", "refs/heads/main"))
	return remote
}

// serveRepoOverHTTP serves the bare repository at dir over the dumb HTTP protocol and returns
// its clone URL. Unlike file:// remotes, http:// remotes can carry credentials in the URL,
// which several tests need. Call refreshServerInfo after pushing new commits to dir.
func serveRepoOverHTTP(t *testing.T, dir string) string {
	t.Helper()
	refreshServerInfo(t, dir)
	srv := httptest.NewServer(http.FileServer(http.Dir(filepath.Dir(dir))))
	t.Cleanup(srv.Close)
	return srv.URL + "/" + filepath.Base(dir)
}

// refreshServerInfo updates the metadata that the dumb HTTP protocol serves.
func refreshServerInfo(t *testing.T, dir string) {
	t.Helper()
	require.NoError(t, execGit(dir, "update-server-info"))
}
