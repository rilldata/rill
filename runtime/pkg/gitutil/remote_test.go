package gitutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetRemote(t *testing.T) {
	ctx := context.Background()

	getURL := func(t *testing.T, path, name string) string {
		t.Helper()
		url, err := Run(ctx, path, "remote", "get-url", name)
		require.NoError(t, err)
		return url
	}

	t.Run("creates a missing remote", func(t *testing.T) {
		path := setupTestRepository(t)
		require.NoError(t, SetRemote(path, &Config{Remote: "https://example.com/repo.git"}))
		require.Equal(t, "https://example.com/repo.git", getURL(t, path, "origin"))
	})

	t.Run("no-op when the remote already has the same URL", func(t *testing.T) {
		path := setupTestRepository(t)
		require.NoError(t, execGit(path, "remote", "add", "origin", "https://example.com/repo.git"))
		require.NoError(t, SetRemote(path, &Config{Remote: "https://example.com/repo.git"}))
		require.Equal(t, "https://example.com/repo.git", getURL(t, path, "origin"))
	})

	t.Run("never overwrites an unmanaged remote", func(t *testing.T) {
		path := setupTestRepository(t)
		require.NoError(t, execGit(path, "remote", "add", "origin", "https://example.com/users-own.git"))
		require.NoError(t, SetRemote(path, &Config{Remote: "https://example.com/other.git"}))
		require.Equal(t, "https://example.com/users-own.git", getURL(t, path, "origin"))
	})

	t.Run("updates a managed remote with a different URL", func(t *testing.T) {
		path := setupTestRepository(t)
		require.NoError(t, execGit(path, "remote", "add", "__rill_remote", "https://example.com/old.git"))
		require.NoError(t, SetRemote(path, &Config{Remote: "https://example.com/new.git", ManagedRepo: true}))
		require.Equal(t, "https://example.com/new.git", getURL(t, path, "__rill_remote"))
	})

	t.Run("no-op when the config has no remote", func(t *testing.T) {
		path := setupTestRepository(t)
		require.NoError(t, SetRemote(path, &Config{}))
		_, err := Run(ctx, path, "remote", "get-url", "origin")
		require.Error(t, err)
	})
}

func TestRemoveRemote(t *testing.T) {
	t.Run("removes an existing remote", func(t *testing.T) {
		path := setupTestRepository(t)
		require.NoError(t, execGit(path, "remote", "add", "__rill_remote", "https://example.com/repo.git"))
		require.NoError(t, RemoveRemote(path, "__rill_remote"))
		_, err := Run(context.Background(), path, "remote", "get-url", "__rill_remote")
		require.Error(t, err)
	})

	t.Run("no-op when the remote does not exist", func(t *testing.T) {
		path := setupTestRepository(t)
		require.NoError(t, RemoveRemote(path, "doesnotexist"))
	})
}
