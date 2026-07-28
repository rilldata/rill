package admin

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/rilldata/rill/runtime/pkg/gitutil"
	"github.com/stretchr/testify/require"
)

func TestGitWriterWriteAndReadFiles(t *testing.T) {
	ctx := context.Background()
	remote := setupBareRepoWithInitialCommit(t)
	cfg := &gitutil.Config{Remote: remote} // no username/password: file:// remotes need no credentials
	w := NewGitWriter()

	// Write two files and delete one of the initial files in a single commit.
	sha, err := w.WriteFiles(ctx, &GitWriteOptions{
		Config:     cfg,
		Branch:     "main",
		BaseBranch: "main",
		Message:    "Sync Rill-managed user files",
		Changes: []FileChange{
			{Path: "user_files/jane-u1/alerts/my-alert.yaml", Data: []byte("type: alert\n")},
			{Path: "user_files/jane-u1/canvas/my-canvas.yaml", Data: []byte("type: canvas\n")},
			{Path: "initial.txt", Data: nil},
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, sha)

	got, err := w.ReadFiles(ctx, cfg, "main", "", []string{
		"user_files/jane-u1/alerts/my-alert.yaml",
		"user_files/jane-u1/canvas/my-canvas.yaml",
		"initial.txt",
	})
	require.NoError(t, err)
	require.Equal(t, []byte("type: alert\n"), got["user_files/jane-u1/alerts/my-alert.yaml"])
	require.Equal(t, []byte("type: canvas\n"), got["user_files/jane-u1/canvas/my-canvas.yaml"])
	require.NotContains(t, got, "initial.txt")

	// Flushing identical content again has nothing to commit.
	_, err = w.WriteFiles(ctx, &GitWriteOptions{
		Config:     cfg,
		Branch:     "main",
		BaseBranch: "main",
		Message:    "Sync Rill-managed user files",
		Changes: []FileChange{
			{Path: "user_files/jane-u1/alerts/my-alert.yaml", Data: []byte("type: alert\n")},
			{Path: "initial.txt", Data: nil}, // already deleted; removal is a no-op
		},
	})
	require.ErrorIs(t, err, gitutil.ErrEmptyCommit)
}

func TestGitWriterWriteFilesSubpath(t *testing.T) {
	ctx := context.Background()
	remote := setupBareRepoWithInitialCommit(t)
	cfg := &gitutil.Config{Remote: remote}
	w := NewGitWriter()

	_, err := w.WriteFiles(ctx, &GitWriteOptions{
		Config:     cfg,
		Branch:     "main",
		BaseBranch: "main",
		Subpath:    "my/project",
		Message:    "Sync Rill-managed user files",
		Changes: []FileChange{
			{Path: "user_files/jane-u1/alerts/my-alert.yaml", Data: []byte("type: alert\n")},
		},
	})
	require.NoError(t, err)

	got, err := w.ReadFiles(ctx, cfg, "main", "my/project", []string{"user_files/jane-u1/alerts/my-alert.yaml"})
	require.NoError(t, err)
	require.Equal(t, []byte("type: alert\n"), got["user_files/jane-u1/alerts/my-alert.yaml"])
}

func TestGitWriterProtectedBranch(t *testing.T) {
	ctx := context.Background()
	remote := setupBareRepoWithInitialCommit(t)

	// Simulate GitHub branch protection with a pre-receive hook that rejects all pushes.
	remoteDir := remote[len("file://"):]
	hook := filepath.Join(remoteDir, "hooks", "pre-receive")
	require.NoError(t, os.WriteFile(hook, []byte("#!/bin/sh\necho \"GH006: Protected branch update failed\" >&2\nexit 1\n"), 0o755))

	w := NewGitWriter()
	_, err := w.WriteFiles(ctx, &GitWriteOptions{
		Config:     &gitutil.Config{Remote: remote},
		Branch:     "main",
		BaseBranch: "main",
		Message:    "Sync Rill-managed user files",
		Changes: []FileChange{
			{Path: "user_files/jane-u1/alerts/my-alert.yaml", Data: []byte("type: alert\n")},
		},
	})
	require.Error(t, err)
	require.False(t, errors.Is(err, gitutil.ErrEmptyCommit))
	require.True(t, IsProtectedBranchErr(err), "expected protected-branch error, got: %v", err)
}

func TestApplyChangesRejectsTraversal(t *testing.T) {
	dir := t.TempDir()
	for _, p := range []string{
		"../escape.yaml",
		"..",
		"user_files/../../escape.yaml",
		"/etc/passwd",
	} {
		err := applyChanges(dir, "", []FileChange{{Path: p, Data: []byte("x")}})
		require.Error(t, err, "expected path %q to be rejected", p)
	}

	// Deleting a file that doesn't exist is a no-op.
	require.NoError(t, applyChanges(dir, "", []FileChange{{Path: "missing.yaml", Data: nil}}))

	// A clean relative path inside the tree is written.
	require.NoError(t, applyChanges(dir, "sub", []FileChange{{Path: "a/b.yaml", Data: []byte("x")}}))
	data, err := os.ReadFile(filepath.Join(dir, "sub", "a", "b.yaml"))
	require.NoError(t, err)
	require.Equal(t, []byte("x"), data)
}

func TestIsProtectedBranchErr(t *testing.T) {
	require.False(t, IsProtectedBranchErr(nil))
	require.False(t, IsProtectedBranchErr(errors.New("connection refused")))
	require.True(t, IsProtectedBranchErr(errors.New("remote: error: GH006: Protected branch update failed for refs/heads/main")))
	require.True(t, IsProtectedBranchErr(errors.New("remote rejected: cannot push to a protected branch")))
	require.True(t, IsProtectedBranchErr(errors.New("required status check \"ci\" is expected")))
	require.True(t, IsProtectedBranchErr(errors.New("push declined due to repository rule violations on branch main")))
}

// setupBareRepoWithInitialCommit creates a bare Git repo with a single commit on main and
// returns its file:// URL.
func setupBareRepoWithInitialCommit(t *testing.T) string {
	remoteDir := t.TempDir()
	runGit(t, "", "init", "--bare", remoteDir)
	runGit(t, remoteDir, "symbolic-ref", "HEAD", "refs/heads/main")

	workDir := t.TempDir()
	runGit(t, "", "clone", remoteDir, workDir)
	runGit(t, workDir, "config", "user.name", "Test")
	runGit(t, workDir, "config", "user.email", "test@example.com")
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "initial.txt"), []byte("initial"), 0o644))
	runGit(t, workDir, "add", ".")
	runGit(t, workDir, "commit", "-m", "Initial commit")
	runGit(t, workDir, "branch", "-M", "main")
	runGit(t, workDir, "push", "origin", "main")

	return "file://" + remoteDir
}

func runGit(t *testing.T, dir string, args ...string) {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v failed: %s", args, out)
}
