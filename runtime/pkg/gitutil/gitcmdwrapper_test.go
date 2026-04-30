package gitutil

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
