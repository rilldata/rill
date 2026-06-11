package gitutil

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMergeWithStrategy(t *testing.T) {
	t.Run("theirs: successful merge without conflicts", func(t *testing.T) {
		tempDir := setupTestRepository(t)

		// Create a feature branch
		cmd := exec.Command("git", "-C", tempDir, "checkout", "-b", "feature")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Git checkout -b output: %s", string(output))
		}
		require.NoError(t, err, "failed to create feature branch")

		// Add a file in the feature branch
		createCommit(t, tempDir, "feature.txt", "feature content", "add feature file")

		// Switch back to main and add a different file
		cmd = exec.Command("git", "-C", tempDir, "checkout", "main")
		output, err = cmd.CombinedOutput()
		require.NoError(t, err, "failed to switch to main branch ", string(output))

		createCommit(t, tempDir, "main.txt", "main content", "add main file")

		// Test merging feature branch using theirs strategy
		err = MergeWithStrategy(tempDir, "feature", "theirs")
		require.NoError(t, err, "MergeWithStrategy(theirs) should succeed without conflicts")

		// Verify both files exist
		featureFile := filepath.Join(tempDir, "feature.txt")
		mainFile := filepath.Join(tempDir, "main.txt")
		require.FileExists(t, featureFile, "feature file should exist after merge")
		require.FileExists(t, mainFile, "main file should exist after merge")
	})

	t.Run("theirs: merge with conflicts resolved using theirs strategy", func(t *testing.T) {
		tempDir := setupTestRepository(t)

		// Create a feature branch
		cmd := exec.Command("git", "-C", tempDir, "checkout", "-b", "feature")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "failed to create feature branch ", string(output))

		// Modify the same file in feature branch
		createCommit(t, tempDir, "test1.txt", "feature version", "modify test1 in feature")

		// Switch back to main and modify the same file differently
		cmd = exec.Command("git", "-C", tempDir, "checkout", "main")
		err = cmd.Run()
		require.NoError(t, err, "failed to switch to main branch")

		createCommit(t, tempDir, "test1.txt", "main version", "modify test1 in main")

		// Test merging feature branch with conflicts using theirs strategy
		err = MergeWithStrategy(tempDir, "feature", "theirs")
		require.NoError(t, err, "MergeWithStrategy(theirs) should resolve conflicts using theirs strategy")

		// Verify the file has the feature branch content (theirs)
		content, err := os.ReadFile(filepath.Join(tempDir, "test1.txt"))
		require.NoError(t, err, "failed to read merged file")
		require.Equal(t, "feature version", string(content), "file should contain feature branch content")
	})

	t.Run("ours: merge with conflicts resolved using ours strategy", func(t *testing.T) {
		tempDir := setupTestRepository(t)

		// Create a feature branch and modify test1.txt
		cmd := exec.Command("git", "-C", tempDir, "checkout", "-b", "feature")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "failed to create feature branch ", string(output))

		createCommit(t, tempDir, "test1.txt", "feature version", "modify test1 in feature")

		// Switch back to main and modify test1.txt differently
		cmd = exec.Command("git", "-C", tempDir, "checkout", "main")
		err = cmd.Run()
		require.NoError(t, err, "failed to switch to main branch")

		createCommit(t, tempDir, "test1.txt", "main version", "modify test1 in main")

		// Test merging feature branch with conflicts using ours strategy
		err = MergeWithStrategy(tempDir, "feature", "ours")
		require.NoError(t, err, "MergeWithStrategy(ours) should resolve conflicts using ours strategy")

		// Verify the file has the main branch content (ours)
		content, err := os.ReadFile(filepath.Join(tempDir, "test1.txt"))
		require.NoError(t, err, "failed to read merged file")
		require.Equal(t, "main version", string(content), "file should contain main branch content")
	})

	t.Run("default: successful merge without conflicts", func(t *testing.T) {
		tempDir := setupTestRepository(t)

		// Create a feature branch with a non-conflicting change
		cmd := exec.Command("git", "-C", tempDir, "checkout", "-b", "feature")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "failed to create feature branch ", string(output))

		createCommit(t, tempDir, "feature.txt", "feature content", "add feature file")

		// Switch back to main and add a different file
		cmd = exec.Command("git", "-C", tempDir, "checkout", "main")
		output, err = cmd.CombinedOutput()
		require.NoError(t, err, "failed to switch to main branch ", string(output))

		createCommit(t, tempDir, "main.txt", "main content", "add main file")

		// Test merging feature branch using the default merge strategy
		err = MergeWithStrategy(tempDir, "feature", "")
		require.NoError(t, err, "MergeWithStrategy(default) should succeed without conflicts")

		// Verify both files exist
		require.FileExists(t, filepath.Join(tempDir, "feature.txt"), "feature file should exist after merge")
		require.FileExists(t, filepath.Join(tempDir, "main.txt"), "main file should exist after merge")
	})

	t.Run("default: merge with conflicts returns error", func(t *testing.T) {
		tempDir := setupTestRepository(t)

		// Create a feature branch and modify test1.txt
		cmd := exec.Command("git", "-C", tempDir, "checkout", "-b", "feature")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "failed to create feature branch ", string(output))

		createCommit(t, tempDir, "test1.txt", "feature version", "modify test1 in feature")

		// Switch back to main and modify test1.txt differently
		cmd = exec.Command("git", "-C", tempDir, "checkout", "main")
		err = cmd.Run()
		require.NoError(t, err, "failed to switch to main branch")

		createCommit(t, tempDir, "test1.txt", "main version", "modify test1 in main")

		// Default merge should fail on conflicts
		err = MergeWithStrategy(tempDir, "feature", "")
		require.Error(t, err, "MergeWithStrategy(default) should fail on conflicts")
	})

	t.Run("unsupported strategy returns error", func(t *testing.T) {
		tempDir := setupTestRepository(t)
		err := MergeWithStrategy(tempDir, "main", "bogus")
		require.Error(t, err, "MergeWithStrategy should reject unsupported strategy")
		require.Contains(t, err.Error(), "unsupported merge strategy")
	})
}

func TestMergeWithBailOnConflict(t *testing.T) {
	t.Run("successful merge without conflicts", func(t *testing.T) {
		tempDir := setupTestRepository(t)

		// Create a feature branch
		cmd := exec.Command("git", "-C", tempDir, "checkout", "-b", "feature")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "failed to create feature branch ", string(output))

		// Add a file in the feature branch
		createCommit(t, tempDir, "feature.txt", "feature content", "add feature file")

		// Switch back to main and add a different file
		cmd = exec.Command("git", "-C", tempDir, "checkout", "main")
		output, err = cmd.CombinedOutput()
		require.NoError(t, err, "failed to switch to main branch ", string(output))

		createCommit(t, tempDir, "main.txt", "main content", "add main file")

		// Test merging feature branch without conflicts
		success, err := MergeWithBailOnConflict(tempDir, "feature")
		require.NoError(t, err, "MergeWithBailOnConflict should succeed without conflicts")
		require.True(t, success, "merge should be successful")

		// Verify both files exist
		featureFile := filepath.Join(tempDir, "feature.txt")
		mainFile := filepath.Join(tempDir, "main.txt")
		require.FileExists(t, featureFile, "feature file should exist after merge")
		require.FileExists(t, mainFile, "main file should exist after merge")
	})

	t.Run("merge with conflicts should abort", func(t *testing.T) {
		tempDir := setupTestRepository(t)

		// Create a feature branch
		cmd := exec.Command("git", "-C", tempDir, "checkout", "-b", "feature")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "failed to create feature branch ", string(output))

		// Modify the same file in feature branch
		createCommit(t, tempDir, "test1.txt", "feature version", "modify test1 in feature")

		// Switch back to main and modify the same file differently
		cmd = exec.Command("git", "-C", tempDir, "checkout", "main")
		output, err = cmd.CombinedOutput()
		require.NoError(t, err, "failed to switch to main branch ", string(output))

		createCommit(t, tempDir, "test1.txt", "main version", "modify test1 in main")

		// Test merging feature branch with conflicts
		success, err := MergeWithBailOnConflict(tempDir, "feature")
		require.NoError(t, err, "MergeWithBailOnConflict should handle conflicts gracefully")
		require.False(t, success, "merge should fail due to conflicts")

		// Verify the file still has the main branch content (merge was aborted)
		content, err := os.ReadFile(filepath.Join(tempDir, "test1.txt"))
		require.NoError(t, err, "failed to read file after aborted merge")
		require.Equal(t, "main version", string(content), "file should contain original main branch content")

		// Verify we're still on main branch and not in merge state
		cmd = exec.Command("git", "-C", tempDir, "status", "--porcelain")
		output, err = cmd.Output()
		require.NoError(t, err, "failed to get git status ", string(output))
		require.Empty(t, string(output), "working directory should be clean after aborted merge")

		// MERGE_HEAD must be gone after the abort.
		require.NoFileExists(t, filepath.Join(tempDir, ".git", "MERGE_HEAD"), "MERGE_HEAD should not exist after abort")
	})

	t.Run("merge failure unrelated to conflicts surfaces as an error", func(t *testing.T) {
		tempDir := setupTestRepository(t)

		success, err := MergeWithBailOnConflict(tempDir, "does-not-exist")
		require.Error(t, err, "non-conflict failures must not be silently swallowed")
		require.False(t, success)
	})
}

func TestCommitAll(t *testing.T) {
	t.Run("returns ErrEmptyCommit when there are no changes", func(t *testing.T) {
		tempDir := setupTestRepository(t)

		_, err := CommitAll(context.Background(), tempDir, "", "noop", "Rill", "noreply@rilldata.com")
		require.ErrorIs(t, err, ErrEmptyCommit)
	})

	t.Run("returns ErrEmptyCommit when changes exist but fall outside the pathspec", func(t *testing.T) {
		tempDir := setupTestRepository(t)

		// Seed the pathspec directory with a committed file so the pathspec resolves to a real path.
		require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "sub"), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "sub", "seed.txt"), []byte("seed"), 0644))
		require.NoError(t, execGit(tempDir, "add", "sub/seed.txt"))
		require.NoError(t, execGit(tempDir, "commit", "-m", "seed"))

		// Introduce a change *outside* the pathspec.
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "outside.txt"), []byte("outside"), 0644))

		_, err := CommitAll(context.Background(), tempDir, "sub", "noop", "Rill", "noreply@rilldata.com")
		require.ErrorIs(t, err, ErrEmptyCommit)

		// The outside file must not have been committed.
		out, err := Run(context.Background(), tempDir, "status", "--porcelain")
		require.NoError(t, err)
		require.Contains(t, out, "outside.txt")
	})

	t.Run("only commits files matching the pathspec", func(t *testing.T) {
		tempDir := setupTestRepository(t)

		require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "sub"), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "sub", "inside.txt"), []byte("inside"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "outside.txt"), []byte("outside"), 0644))

		hash, err := CommitAll(context.Background(), tempDir, "sub", "scoped commit", "Rill", "noreply@rilldata.com")
		require.NoError(t, err)
		require.NotEmpty(t, hash)

		// The outside file should still be untracked (not committed).
		out, err := Run(context.Background(), tempDir, "status", "--porcelain")
		require.NoError(t, err)
		require.Contains(t, out, "outside.txt")
		require.NotContains(t, out, "inside.txt")

		// The commit should include only the inside file.
		filesOut, err := Run(context.Background(), tempDir, "show", "--name-only", "--pretty=format:", "HEAD")
		require.NoError(t, err)
		require.Contains(t, filesOut, "sub/inside.txt")
		require.NotContains(t, filesOut, "outside.txt")
	})

	t.Run("uses the provided author name and email", func(t *testing.T) {
		tempDir := setupTestRepository(t)

		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "new.txt"), []byte("hello"), 0644))

		_, err := CommitAll(context.Background(), tempDir, "", "msg", "Rill Bot", "bot@rilldata.com")
		require.NoError(t, err)

		name, err := Run(context.Background(), tempDir, "log", "-1", "--format=%an")
		require.NoError(t, err)
		require.Equal(t, "Rill Bot", name)

		email, err := Run(context.Background(), tempDir, "log", "-1", "--format=%ae")
		require.NoError(t, err)
		require.Equal(t, "bot@rilldata.com", email)
	})
}

func TestFetchBranches(t *testing.T) {
	t.Run("silently skips branches that do not exist on the remote", func(t *testing.T) {
		remote, local := setupRemoteAndClone(t)

		// Add a new branch on the remote with a commit.
		require.NoError(t, execGit(remote, "checkout", "-b", "feature"))
		createCommit(t, remote, "f.txt", "f", "feature commit")
		require.NoError(t, execGit(remote, "checkout", "main"))

		err := FetchBranches(context.Background(), local, "feature", "does-not-exist")
		require.NoError(t, err, "missing branch must not produce an error")

		// The existing branch must have been fetched.
		hash, err := Hash(context.Background(), local, "refs/remotes/origin/feature")
		require.NoError(t, err)
		require.NotEmpty(t, hash)
	})
}

func TestHash(t *testing.T) {
	t.Run("returns ErrRefNotFound for a missing ref", func(t *testing.T) {
		tempDir := setupTestRepository(t)

		_, err := Hash(context.Background(), tempDir, "refs/heads/does-not-exist")
		require.ErrorIs(t, err, ErrRefNotFound)
	})

	t.Run("returns ErrRefNotFound for HEAD in an unborn repository", func(t *testing.T) {
		tempDir := t.TempDir()
		require.NoError(t, exec.Command("git", "init", tempDir).Run())
		require.NoError(t, exec.Command("git", "-C", tempDir, "checkout", "-b", "main").Run())
		setupGitConfig(t, tempDir)

		_, err := Hash(context.Background(), tempDir, "HEAD")
		require.ErrorIs(t, err, ErrRefNotFound)
	})

	t.Run("returns the commit hash for HEAD", func(t *testing.T) {
		tempDir := setupTestRepository(t)

		hash, err := Hash(context.Background(), tempDir, "HEAD")
		require.NoError(t, err)
		require.Len(t, hash, 40)
	})
}

func TestIsCommitHash(t *testing.T) {
	require.True(t, IsCommitHash(strings.Repeat("a1", 20)), "40-char SHA-1")
	require.True(t, IsCommitHash(strings.Repeat("a1", 32)), "64-char SHA-256")
	require.True(t, IsCommitHash(strings.Repeat("A1", 20)), "uppercase hex")
	require.False(t, IsCommitHash(""), "empty")
	require.False(t, IsCommitHash("abc1"), "abbreviated hash")
	require.False(t, IsCommitHash("--output=/tmp/x"), "flag-like input")
	require.False(t, IsCommitHash("HEAD"), "symbolic ref")
	require.False(t, IsCommitHash(strings.Repeat("g", 40)), "non-hex characters")
}

func TestRunRedactsURLCredentials(t *testing.T) {
	tempDir := setupTestRepository(t)

	// Fetch from an unreachable credential-embedded URL: both the args and git's stderr contain the URL.
	_, err := Run(context.Background(), tempDir, "fetch", "https://user:secret-token@host.invalid/org/repo.git")
	require.Error(t, err)
	require.NotContains(t, err.Error(), "secret-token", "credentials must not leak into errors")
	require.Contains(t, err.Error(), "<redacted>@")
}

// setupRemoteAndClone creates a remote repository with a single commit on `main`
// and clones it into a separate local directory. It returns the remote and local paths.
func setupRemoteAndClone(t *testing.T) (string, string) {
	remote := setupTestRepository(t)
	local := t.TempDir()
	// t.TempDir() pre-creates the directory; remove it so `git clone` can populate it.
	require.NoError(t, os.RemoveAll(local))

	err := Clone(context.Background(), local, remote, "main", false, false)
	require.NoError(t, err)
	setupGitConfig(t, local)
	return remote, local
}

func execGit(repoPath string, args ...string) error {
	cmd := exec.Command("git", append([]string{"-C", repoPath}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %v failed: %s: %w", args, string(out), err)
	}
	return nil
}

func setupTestRepository(t *testing.T) string {
	tempDir := t.TempDir()

	// Initialize a new git repository in the temp directory
	cmd := exec.Command("git", "init", tempDir)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "failed to initialize git repository ", string(output))

	// Set the default branch to main
	cmd = exec.Command("git", "-C", tempDir, "checkout", "-b", "main")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "failed to create main branch ", string(output))

	setupGitConfig(t, tempDir)

	// Create and commit multiple files
	for i := 1; i <= 3; i++ {
		filePath := filepath.Join(tempDir, fmt.Sprintf("test%d.txt", i))
		err = os.WriteFile(filePath, []byte(fmt.Sprintf("content of file %d", i)), 0644)
		require.NoError(t, err, "failed to create test file")

		cmd = exec.Command("git", "-C", tempDir, "add", fmt.Sprintf("test%d.txt", i))
		err = cmd.Run()
		require.NoError(t, err, "failed to stage file")
	}

	cmd = exec.Command("git", "-C", tempDir, "commit", "-m", "initial commit")
	_, err = cmd.Output()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			require.NoError(t, err, "failed to commit files: "+fmt.Sprintf(": %s", execErr.Stderr))
		}
		require.NoError(t, err, "failed to commit files")
	}
	return tempDir
}

// createCommit creates a file and commits it to the repository
func createCommit(t *testing.T, repoPath, fileName, fileContent, commitMessage string) {
	filePath := filepath.Join(repoPath, fileName)
	err := os.WriteFile(filePath, []byte(fileContent), 0644)
	require.NoError(t, err, "failed to create file")

	cmd := exec.Command("git", "-C", repoPath, "add", fileName)
	_, err = cmd.Output()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			require.NoError(t, err, "failed to stage file"+fmt.Sprintf(": %s", execErr.Stderr))
		}
		require.NoError(t, err, "failed to stage file")
	}

	cmd = exec.Command("git", "-C", repoPath, "commit", "-m", commitMessage)
	_, err = cmd.Output()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			require.NoError(t, err, "failed to commit file"+fmt.Sprintf(": %s", execErr.Stderr))
		}
		require.NoError(t, err, "failed to commit file")
	}
}

// setupGitConfig sets up the git configuration for the repository at repoPath
func setupGitConfig(t *testing.T, repoPath string) {
	// Set user name and email for the git repository
	cmd := exec.Command("git", "-C", repoPath, "config", "user.name", "Test User")
	err := cmd.Run()
	require.NoError(t, err, "failed to set user name in git config")

	cmd = exec.Command("git", "-C", repoPath, "config", "user.email", "test@rilldata.com")
	err = cmd.Run()
	require.NoError(t, err, "failed to set user email in git config")
}
