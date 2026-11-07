package local

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCurrentGitBranch(t *testing.T) {
	ctx := context.Background()

	gitExec := func(t *testing.T, dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "git %v: %s", args, out)
	}

	setupRepo := func(t *testing.T) string {
		t.Helper()
		dir := t.TempDir()
		gitExec(t, dir, "init")
		gitExec(t, dir, "checkout", "-b", "main")
		gitExec(t, dir, "config", "user.name", "Test User")
		gitExec(t, dir, "config", "user.email", "test@rilldata.com")
		require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0644))
		gitExec(t, dir, "add", ".")
		gitExec(t, dir, "commit", "-m", "initial commit")
		return dir
	}

	t.Run("returns empty for a non-repo directory", func(t *testing.T) {
		branch, err := currentGitBranch(ctx, t.TempDir())
		require.NoError(t, err)
		require.Equal(t, "", branch)
	})

	t.Run("returns empty for a non-repo directory nested inside a repo", func(t *testing.T) {
		dir := setupRepo(t)
		nested := filepath.Join(dir, "nested")
		require.NoError(t, os.MkdirAll(nested, 0755))
		branch, err := currentGitBranch(ctx, nested)
		require.NoError(t, err)
		require.Equal(t, "", branch, "must not report the parent repo's branch")
	})

	t.Run("returns the current branch", func(t *testing.T) {
		dir := setupRepo(t)
		branch, err := currentGitBranch(ctx, dir)
		require.NoError(t, err)
		require.Equal(t, "main", branch)
	})

	t.Run("returns the unborn branch of a fresh repository", func(t *testing.T) {
		dir := t.TempDir()
		gitExec(t, dir, "init")
		gitExec(t, dir, "symbolic-ref", "HEAD", "refs/heads/main")
		branch, err := currentGitBranch(ctx, dir)
		require.NoError(t, err)
		require.Equal(t, "main", branch, "a git-init'ed project without commits must be deployable")
	})

	t.Run("errors on a detached HEAD", func(t *testing.T) {
		dir := setupRepo(t)
		gitExec(t, dir, "checkout", "--detach")
		_, err := currentGitBranch(ctx, dir)
		require.Error(t, err)
		require.Contains(t, err.Error(), "HEAD is not a branch")
	})
}
