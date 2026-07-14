package gitutil

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommitAndForcePush(t *testing.T) {
	ctx := context.Background()
	author := Signature{Name: "Rill", Email: "noreply@rilldata.com"}

	t.Run("succeeds on matching branch with uncommitted changes", func(t *testing.T) {
		local, remote := setupRepoWithRemote(t)
		branch := getCurrentBranch(t, local)

		require.NoError(t, os.WriteFile(filepath.Join(local, "new.txt"), []byte("hello"), 0644))

		config := &Config{Remote: remote, DefaultBranch: branch}
		require.NoError(t, CommitAndForcePush(ctx, local, config, "force push commit", author))

		localTip, err := Hash(ctx, local, "HEAD")
		require.NoError(t, err)
		remoteTip, err := Hash(ctx, remote, "refs/heads/"+branch)
		require.NoError(t, err)
		require.Equal(t, localTip, remoteTip)
	})

	t.Run("pushes even when there is nothing to commit", func(t *testing.T) {
		local, remote := setupRepoWithRemote(t)
		branch := getCurrentBranch(t, local)

		// create a local-only commit with nothing left in the working tree
		createCommit(t, local, "unpushed.txt", "content", "local-only commit")
		localTip, err := Hash(ctx, local, "HEAD")
		require.NoError(t, err)

		config := &Config{Remote: remote, DefaultBranch: branch}
		require.NoError(t, CommitAndForcePush(ctx, local, config, "", author))

		remoteTip, err := Hash(ctx, remote, "refs/heads/"+branch)
		require.NoError(t, err)
		require.Equal(t, localTip, remoteTip, "local-only commit must reach the remote")
	})

	t.Run("succeeds from a detached HEAD", func(t *testing.T) {
		local, remote := setupRepoWithRemote(t)
		branch := getCurrentBranch(t, local)

		require.NoError(t, execGit(local, "checkout", "--detach"))
		require.NoError(t, os.WriteFile(filepath.Join(local, "detached.txt"), []byte("from detached HEAD"), 0644))

		config := &Config{Remote: remote, DefaultBranch: branch}
		require.NoError(t, CommitAndForcePush(ctx, local, config, "detached commit", author))

		localTip, err := Hash(ctx, local, "HEAD")
		require.NoError(t, err)
		remoteTip, err := Hash(ctx, remote, "refs/heads/"+branch)
		require.NoError(t, err)
		require.Equal(t, localTip, remoteTip, "force push from detached HEAD must update the remote branch")
	})

	t.Run("succeeds from a different local branch", func(t *testing.T) {
		local, remote := setupRepoWithRemote(t)
		defaultBranch := getCurrentBranch(t, local)

		require.NoError(t, execGit(local, "checkout", "-b", "feature"))
		require.NoError(t, os.WriteFile(filepath.Join(local, "feature.txt"), []byte("feature work"), 0644))

		config := &Config{Remote: remote, DefaultBranch: defaultBranch}
		require.NoError(t, CommitAndForcePush(ctx, local, config, "feature commit", author))

		localTip, err := Hash(ctx, local, "HEAD")
		require.NoError(t, err)
		remoteTip, err := Hash(ctx, remote, "refs/heads/"+defaultBranch)
		require.NoError(t, err)
		require.Equal(t, localTip, remoteTip, "remote default branch must be overwritten with the local feature branch tip")
	})

	t.Run("overwrites divergent remote history", func(t *testing.T) {
		local, remote := setupRepoWithRemote(t)
		branch := getCurrentBranch(t, local)

		// advance the remote with a commit that local doesn't have
		createRemoteCommit(t, remote, "remote-only.txt", "remote content", "remote-only commit")

		// advance local with a different commit, creating a true divergence from the initial tip
		createCommit(t, local, "local-only.txt", "local content", "local-only commit")

		// add uncommitted changes so CommitAll also has work to do
		require.NoError(t, os.WriteFile(filepath.Join(local, "extra.txt"), []byte("extra"), 0644))

		config := &Config{Remote: remote, DefaultBranch: branch}
		require.NoError(t, CommitAndForcePush(ctx, local, config, "force over divergence", author))

		localTip, err := Hash(ctx, local, "HEAD")
		require.NoError(t, err)
		remoteTip, err := Hash(ctx, remote, "refs/heads/"+branch)
		require.NoError(t, err)
		require.Equal(t, localTip, remoteTip, "force push must overwrite the divergent remote history")
	})
}

func TestCurrentBranch(t *testing.T) {
	ctx := context.Background()

	t.Run("returns the current branch", func(t *testing.T) {
		path := setupTestRepository(t)
		branch, err := CurrentBranch(ctx, path)
		require.NoError(t, err)
		require.Equal(t, "main", branch)
	})

	t.Run("returns the unborn branch of a fresh repository", func(t *testing.T) {
		path := t.TempDir()
		require.NoError(t, EnsureInit(ctx, path, "unborn"))
		branch, err := CurrentBranch(ctx, path)
		require.NoError(t, err)
		require.Equal(t, "unborn", branch)
	})

	t.Run("returns ErrDetachedHead on a detached HEAD", func(t *testing.T) {
		path := setupTestRepository(t)
		require.NoError(t, execGit(path, "checkout", "--detach"))
		_, err := CurrentBranch(ctx, path)
		require.ErrorIs(t, err, ErrDetachedHead)
	})
}

func TestUserSignature(t *testing.T) {
	ctx := context.Background()

	t.Run("returns the configured identity", func(t *testing.T) {
		path := setupTestRepository(t) // setupGitConfig sets Test User <test@rilldata.com>
		sig, err := UserSignature(ctx, path)
		require.NoError(t, err)
		require.Equal(t, Signature{Name: "Test User", Email: "test@rilldata.com"}, sig)
	})

	t.Run("errors when the identity is not configured", func(t *testing.T) {
		// hide the developer's global/system config so the lookup sees no identity
		t.Setenv("GIT_CONFIG_GLOBAL", "/dev/null")
		t.Setenv("GIT_CONFIG_SYSTEM", "/dev/null")

		path := t.TempDir()
		require.NoError(t, EnsureInit(ctx, path, "main"))
		_, err := UserSignature(ctx, path)
		require.Error(t, err)
		require.Contains(t, err.Error(), "is not set in git config")
	})
}

func TestCommitAndPush(t *testing.T) {
	ctx := context.Background()
	author := Signature{Name: "Rill", Email: "noreply@rilldata.com"}

	t.Run("initializes a fresh directory on the default branch and pushes", func(t *testing.T) {
		remote := filepath.Join(t.TempDir(), "remote.git")
		_, err := Run(ctx, "", "init", "--bare", remote)
		require.NoError(t, err)

		path := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(path, "file.txt"), []byte("content"), 0644))

		config := &Config{Remote: remote, DefaultBranch: "deploybranch", ManagedRepo: true}
		require.NoError(t, CommitAndPush(ctx, path, config, "", author))

		// the commit landed on the configured branch, regardless of the user's init.defaultBranch
		branch, err := CurrentBranch(ctx, path)
		require.NoError(t, err)
		require.Equal(t, "deploybranch", branch)

		// the remote received the branch
		_, err = Hash(ctx, remote, "refs/heads/deploybranch")
		require.NoError(t, err)

		// the managed remote was persisted with the clean URL
		url, err := Run(ctx, path, "remote", "get-url", "__rill_remote")
		require.NoError(t, err)
		require.Equal(t, remote, url)

		// the default commit message was used
		msg, err := Run(ctx, path, "log", "-1", "--format=%s")
		require.NoError(t, err)
		require.Equal(t, "Auto committed by Rill", msg)
	})

	t.Run("pushes even when there is nothing to commit", func(t *testing.T) {
		local, remoteDir := setupRepoWithRemote(t)
		branch := getCurrentBranch(t, local)

		// create a local commit that exists only locally
		createCommit(t, local, "unpushed.txt", "content", "local-only commit")
		localTip, err := Hash(ctx, local, "HEAD")
		require.NoError(t, err)

		// nothing left in the working tree, so the commit step is empty but the push must still run
		config := &Config{Remote: remoteDir, DefaultBranch: branch}
		require.NoError(t, CommitAndPush(ctx, local, config, "", author))

		remoteTip, err := Hash(ctx, remoteDir, "refs/heads/"+branch)
		require.NoError(t, err)
		require.Equal(t, localTip, remoteTip, "the local-only commit must reach the remote")
	})

	t.Run("errors when the current branch differs from the default branch", func(t *testing.T) {
		local, remoteDir := setupRepoWithRemote(t)
		config := &Config{Remote: remoteDir, DefaultBranch: "someotherbranch"}
		err := CommitAndPush(ctx, local, config, "", author)
		require.Error(t, err)
		require.Contains(t, err.Error(), "does not match deployed branch")
	})

	t.Run("errors on a detached HEAD", func(t *testing.T) {
		local, remoteDir := setupRepoWithRemote(t)
		require.NoError(t, execGit(local, "checkout", "--detach"))
		config := &Config{Remote: remoteDir, DefaultBranch: getCurrentBranch(t, local)}
		err := CommitAndPush(ctx, local, config, "", author)
		require.Error(t, err)
		require.Contains(t, err.Error(), "detached HEAD")
	})

	t.Run("fails fast on an unborn branch that differs from the default branch", func(t *testing.T) {
		path := t.TempDir()
		require.NoError(t, EnsureInit(ctx, path, "first"))
		require.NoError(t, os.WriteFile(filepath.Join(path, "file.txt"), []byte("content"), 0644))

		config := &Config{Remote: filepath.Join(t.TempDir(), "remote.git"), DefaultBranch: "second"}
		err := CommitAndPush(ctx, path, config, "", author)
		require.Error(t, err)
		require.Contains(t, err.Error(), "does not match deployed branch")
	})

	t.Run("scopes the commit to the subpath", func(t *testing.T) {
		local, remoteDir := setupRepoWithRemote(t)
		branch := getCurrentBranch(t, local)

		// seed the subpath with a committed file and a gitignore rule
		require.NoError(t, os.MkdirAll(filepath.Join(local, "sub"), 0755))
		createCommit(t, local, "sub/seeded.txt", "seeded", "seed subpath")
		createCommit(t, local, ".gitignore", "sub/ignored.txt\n", "add gitignore")
		require.NoError(t, execGit(local, "push", "origin", branch))

		// stage a mix of changes inside and outside the subpath
		require.NoError(t, os.WriteFile(filepath.Join(local, "sub", "inside.txt"), []byte("inside"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(local, "sub", ".dotfile"), []byte("dot"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(local, "sub", "ignored.txt"), []byte("ignored"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(local, "outside.txt"), []byte("outside"), 0644))
		require.NoError(t, os.Remove(filepath.Join(local, "sub", "seeded.txt")))

		config := &Config{Remote: remoteDir, DefaultBranch: branch, Subpath: "sub"}
		require.NoError(t, CommitAndPush(ctx, local, config, "subpath commit", author))

		committed, err := Run(ctx, local, "show", "--name-status", "--pretty=format:", "HEAD")
		require.NoError(t, err)
		require.Contains(t, committed, "sub/inside.txt")
		require.Contains(t, committed, "sub/.dotfile", "dotfiles inside the subpath must be committed")
		require.Contains(t, committed, "D\tsub/seeded.txt", "deletions inside the subpath must be committed")
		require.NotContains(t, committed, "ignored.txt", "gitignored files must not be committed")
		require.NotContains(t, committed, "outside.txt", "changes outside the subpath must not be committed")

		// the outside change is left untouched in the working tree
		status, err := Run(ctx, local, "status", "--porcelain")
		require.NoError(t, err)
		require.Contains(t, status, "outside.txt")
	})
}
