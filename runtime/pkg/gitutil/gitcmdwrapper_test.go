package gitutil

import (
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
	})
}

func TestCheckout(t *testing.T) {
	t.Run("checkout existing branch", func(t *testing.T) {
		tempDir := setupTestRepository(t)

		cmd := exec.Command("git", "-C", tempDir, "checkout", "-b", "feature")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "failed to create feature branch: "+string(output))

		cmd = exec.Command("git", "-C", tempDir, "checkout", "main")
		output, err = cmd.CombinedOutput()
		require.NoError(t, err, "failed to switch to main: "+string(output))

		err = Checkout(tempDir, "feature", false, false, "")
		require.NoError(t, err)

		branch, err := CurrentBranch(tempDir)
		require.NoError(t, err)
		require.Equal(t, "feature", branch)
	})

	t.Run("checkout non-existent branch returns ErrRefNotFound", func(t *testing.T) {
		tempDir := setupTestRepository(t)

		err := Checkout(tempDir, "does-not-exist", false, false, "")
		require.ErrorIs(t, err, ErrRefNotFound)
	})

	t.Run("checkout with force discards local changes", func(t *testing.T) {
		tempDir := setupTestRepository(t)

		// Dirty the working tree
		err := os.WriteFile(filepath.Join(tempDir, "test1.txt"), []byte("dirty"), 0644)
		require.NoError(t, err)

		err = Checkout(tempDir, "main", true, false, "")
		require.NoError(t, err)

		// File should be restored to its committed state
		content, err := os.ReadFile(filepath.Join(tempDir, "test1.txt"))
		require.NoError(t, err)
		require.Equal(t, "content of file 1", string(content))
	})

	t.Run("create branch with -B", func(t *testing.T) {
		tempDir := setupTestRepository(t)

		err := Checkout(tempDir, "new-branch", false, true, "")
		require.NoError(t, err)

		branch, err := CurrentBranch(tempDir)
		require.NoError(t, err)
		require.Equal(t, "new-branch", branch)
	})

	t.Run("create branch at startPoint", func(t *testing.T) {
		tempDir := setupTestRepository(t)

		// Get current HEAD hash to use as startPoint
		out, err := exec.Command("git", "-C", tempDir, "rev-parse", "HEAD").Output()
		require.NoError(t, err)
		startPoint := strings.TrimSpace(string(out))

		createCommit(t, tempDir, "extra.txt", "extra", "add extra file")

		// Create a new branch pointing at the earlier commit
		err = Checkout(tempDir, "at-start", false, true, startPoint)
		require.NoError(t, err)

		branch, err := CurrentBranch(tempDir)
		require.NoError(t, err)
		require.Equal(t, "at-start", branch)

		// extra.txt should not exist on this branch
		_, statErr := os.Stat(filepath.Join(tempDir, "extra.txt"))
		require.True(t, os.IsNotExist(statErr), "extra.txt should not exist at startPoint")
	})
}

func TestResetToRemote(t *testing.T) {
	t.Run("resets local branch to remote state", func(t *testing.T) {
		repoDir, _ := setupTestRepositoryWithRemote(t)

		// Make a local commit that hasn't been pushed
		createCommit(t, repoDir, "local-only.txt", "local only content", "local commit")

		err := ResetToRemote(repoDir, "main")
		require.NoError(t, err)

		// local-only.txt should be gone after the hard reset
		_, statErr := os.Stat(filepath.Join(repoDir, "local-only.txt"))
		require.True(t, os.IsNotExist(statErr), "local-only.txt should not exist after reset to remote")
	})

	t.Run("non-existent remote branch returns ErrRefNotFound", func(t *testing.T) {
		repoDir, _ := setupTestRepositoryWithRemote(t)

		err := ResetToRemote(repoDir, "does-not-exist")
		require.ErrorIs(t, err, ErrRefNotFound)
	})
}

func TestCurrentBranch(t *testing.T) {
	t.Run("returns main on initial branch", func(t *testing.T) {
		tempDir := setupTestRepository(t)

		branch, err := CurrentBranch(tempDir)
		require.NoError(t, err)
		require.Equal(t, "main", branch)
	})

	t.Run("returns correct branch after checkout", func(t *testing.T) {
		tempDir := setupTestRepository(t)

		cmd := exec.Command("git", "-C", tempDir, "checkout", "-b", "my-feature")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "failed to create branch: "+string(output))

		branch, err := CurrentBranch(tempDir)
		require.NoError(t, err)
		require.Equal(t, "my-feature", branch)
	})
}

func TestCommitAll(t *testing.T) {
	t.Run("commits all changes and returns hash", func(t *testing.T) {
		tempDir := setupTestRepository(t)

		err := os.WriteFile(filepath.Join(tempDir, "new.txt"), []byte("hello"), 0644)
		require.NoError(t, err)

		hash, err := CommitAll(tempDir, "", "test commit", "Alice", "alice@example.com")
		require.NoError(t, err)
		require.NotEmpty(t, hash, "should return commit hash")

		// Verify HEAD matches the returned hash
		out, err := exec.Command("git", "-C", tempDir, "rev-parse", "HEAD").Output()
		require.NoError(t, err)
		require.Equal(t, hash, strings.TrimSpace(string(out)))
	})

	t.Run("nothing to commit returns empty string", func(t *testing.T) {
		tempDir := setupTestRepository(t)

		hash, err := CommitAll(tempDir, "", "empty commit", "Alice", "alice@example.com")
		require.NoError(t, err)
		require.Empty(t, hash, "should return empty string when nothing to commit")
	})

	t.Run("commits only files matching glob", func(t *testing.T) {
		tempDir := setupTestRepository(t)

		err := os.WriteFile(filepath.Join(tempDir, "include.yaml"), []byte("included"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tempDir, "exclude.txt"), []byte("excluded"), 0644)
		require.NoError(t, err)

		hash, err := CommitAll(tempDir, "*.yaml", "yaml only", "Alice", "alice@example.com")
		require.NoError(t, err)
		require.NotEmpty(t, hash)

		// exclude.txt should be untracked (not staged/committed)
		out, err := exec.Command("git", "-C", tempDir, "status", "--porcelain").Output()
		require.NoError(t, err)
		require.Contains(t, string(out), "exclude.txt", "exclude.txt should remain untracked")
	})

	t.Run("author name and email are recorded on commit", func(t *testing.T) {
		tempDir := setupTestRepository(t)

		err := os.WriteFile(filepath.Join(tempDir, "authored.txt"), []byte("content"), 0644)
		require.NoError(t, err)

		_, err = CommitAll(tempDir, "", "authored commit", "Bob", "bob@example.com")
		require.NoError(t, err)

		out, err := exec.Command("git", "-C", tempDir, "log", "-1", "--format=%an %ae").Output()
		require.NoError(t, err)
		require.Equal(t, "Bob bob@example.com", strings.TrimSpace(string(out)))
	})
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

// setupTestRepositoryWithRemote creates a test repo with a bare remote and pushes main to it.
func setupTestRepositoryWithRemote(t *testing.T) (repoDir, remoteDir string) {
	remoteDir = t.TempDir()
	cmd := exec.Command("git", "init", "--bare", remoteDir)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "failed to init bare remote: "+string(output))

	repoDir = setupTestRepository(t)

	cmd = exec.Command("git", "-C", repoDir, "remote", "add", "origin", remoteDir)
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "failed to add remote: "+string(output))

	cmd = exec.Command("git", "-C", repoDir, "push", "-u", "origin", "main")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "failed to push to remote: "+string(output))

	return repoDir, remoteDir
}
