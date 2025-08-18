package gitutil

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunGitStatus(t *testing.T) {
	tempDir, remoteDir := setupTestRepository(t)

	// Test case: Local commits
	createCommit(t, tempDir, "test4.txt", "add a local file", "local commit")

	// Run GitFetch
	require.NoError(t, GitFetch(t.Context(), tempDir, nil), "failed to fetch changes from remote repository")
	// Run the RunGitStatus function again
	status, err := RunGitStatus(tempDir, "origin")
	require.NoError(t, err, "RunGitStatus failed after local commit")

	// Validate the updated status
	require.Equal(t, int32(1), status.LocalCommits, "unexpected local commits after fourth commit")
	require.Equal(t, int32(0), status.RemoteCommits, "unexpected remote commits")

	// Test case: Remote commits
	createRemoteCommit(t, remoteDir, "test5.txt", "add a file in remote", "remote commit")

	// Fetch the latest changes from the remote repository before running RunGitStatus
	require.NoError(t, GitFetch(t.Context(), tempDir, nil), "failed to fetch changes from remote repository")

	// Run the RunGitStatus function again
	status, err = RunGitStatus(tempDir, "origin")
	require.NoError(t, err, "RunGitStatus failed after local commit")

	// Validate the updated status
	require.Equal(t, int32(1), status.LocalCommits, "unexpected local commits after fourth commit")
	require.Equal(t, int32(1), status.RemoteCommits, "unexpected remote commits")

	// Test case: Untracked files
	filePath := filepath.Join(tempDir, "untracked.txt")
	err = os.WriteFile(filePath, []byte("untracked content"), 0644)
	require.NoError(t, err, "failed to create untracked file")

	status, err = RunGitStatus(tempDir, "origin")
	require.NoError(t, err, "RunGitStatus failed with untracked files")
	require.True(t, status.LocalChanges, "expected local changes due to untracked files")

	// Test case: Staged but uncommitted changes
	err = os.WriteFile(filePath, []byte("staged content"), 0644)
	require.NoError(t, err, "failed to modify file for staging")

	cmd := exec.Command("git", "-C", tempDir, "add", "untracked.txt")
	err = cmd.Run()
	require.NoError(t, err, "failed to stage file")

	status, err = RunGitStatus(tempDir, "origin")
	require.NoError(t, err, "RunGitStatus failed with staged changes")
	require.True(t, status.LocalChanges, "expected local changes due to staged files")

	// Test case: Unstaged changes
	err = os.WriteFile(filePath, []byte("unstaged content"), 0644)
	require.NoError(t, err, "failed to modify file for unstaged changes")

	status, err = RunGitStatus(tempDir, "origin")
	require.NoError(t, err, "RunGitStatus failed with unstaged changes")
	require.True(t, status.LocalChanges, "expected local changes due to unstaged files")
}

func TestGitPull(t *testing.T) {
	tempDir, remoteDir := setupTestRepository(t)

	// Test case: Pull with no local changes
	output, err := RunGitPull(context.Background(), tempDir, false, "", "origin")
	require.NoError(t, err, "GitPull failed with no local changes")
	require.Empty(t, output, "unexpected output from GitPull with no local changes")

	// Test case: Pull with local changes (discardLocal = false)
	createCommit(t, tempDir, "local.txt", "local content", "local commit")
	createRemoteCommit(t, remoteDir, "local.txt", "remote content", "remote commit")
	output, err = RunGitPull(context.Background(), tempDir, false, "", "origin")
	if len(output) == 0 && err == nil {
		t.Fatalf("expected GitPull to fail with local changes and discardLocal=false, but it succeeded")
	}
	require.Contains(t, output, "Need to specify how to reconcile divergent branches", "unexpected output from GitPull with local changes")

	// Test case: Pull with local changes (discardLocal = true)
	output, err = RunGitPull(context.Background(), tempDir, true, "", "origin")
	require.NoError(t, err, "GitPull failed with local changes and discardLocal=true")
	require.Empty(t, output, "unexpected output from GitPull with discardLocal=true")

	// Test case: Pull with remote changes
	createRemoteCommit(t, remoteDir, "remote.txt", "remote content", "remote commit")
	output, err = RunGitPull(context.Background(), tempDir, false, "", "origin")
	require.NoError(t, err, "GitPull failed with remote changes")
	require.Empty(t, output, "unexpected output from GitPull with remote changes")
}

func TestRunGitPush(t *testing.T) {
	tempDir, _ := setupTestRepository(t)

	// Test case: Push with local commits to existing branch
	createCommit(t, tempDir, "push_test.txt", "content for push test", "local commit for push")

	err := RunGitPush(context.Background(), tempDir, "origin", "main")
	require.NoError(t, err, "RunGitPush failed for existing branch")

	// Verify the commit was pushed by checking remote status
	err = RunGitFetch(context.Background(), tempDir, "origin")
	require.NoError(t, err, "failed to fetch after push")

	status, err := RunGitStatus(tempDir, "origin")
	require.NoError(t, err, "RunGitStatus failed after push")
	require.Equal(t, int32(0), status.LocalCommits, "expected no local commits after successful push")

	// Test case: Push a new branch
	// Create a new branch
	cmd := exec.Command("git", "-C", tempDir, "checkout", "-b", "feature-branch")
	err = cmd.Run()
	require.NoError(t, err, "failed to create new branch")

	// Create a commit on the new branch
	createCommit(t, tempDir, "new_branch.txt", "content for new branch", "commit on new branch")

	err = RunGitPush(context.Background(), tempDir, "origin", "feature-branch")
	require.NoError(t, err, "RunGitPush failed for new branch")

	// Verify the new branch was pushed by checking if it exists on remote
	cmd = exec.Command("git", "-C", tempDir, "ls-remote", "--heads", "origin", "feature-branch")
	output, err := cmd.Output()
	require.NoError(t, err, "failed to check remote branches")
	require.Contains(t, string(output), "feature-branch", "new branch was not pushed to remote")

	// Test case: Push with no local commits (should succeed but do nothing)
	err = RunGitPush(context.Background(), tempDir, "origin", "feature-branch")
	require.NoError(t, err, "RunGitPush failed when there are no new commits")

	// Test case: Push to non-existent remote
	err = RunGitPush(context.Background(), tempDir, "nonexistent", "feature-branch")
	require.Error(t, err, "expected error when pushing to non-existent remote")
	require.Contains(t, err.Error(), "git push failed", "error message should mention git push failure")

	// Test case: Push non-existent branch
	err = RunGitPush(context.Background(), tempDir, "origin", "nonexistent-branch")
	require.Error(t, err, "expected error when pushing non-existent branch")
	require.Contains(t, err.Error(), "git push failed", "error message should mention git push failure")

	// Test case: Push with context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = RunGitPush(ctx, tempDir, "origin", "feature-branch")
	require.Error(t, err, "expected error when context is cancelled")
}

func TestRunGitPush_InvalidRepository(t *testing.T) {
	// Test case: Push from non-git directory
	tempDir := t.TempDir()

	err := RunGitPush(context.Background(), tempDir, "origin", "main")
	require.Error(t, err, "expected error when pushing from non-git directory")
	require.Contains(t, err.Error(), "git push failed", "error message should mention git push failure")
}

func TestInferGitRepoRoot_InRepoRoot(t *testing.T) {
	tempDir, _ := setupTestRepository(t)

	root, err := InferGitRepoRoot(tempDir)
	require.NoError(t, err, "InferGitRepoRoot failed on repo root")
	assertPathsEqual(t, root, tempDir)
}

func TestInferGitRepoRoot_InNestedDir(t *testing.T) {
	tempDir, _ := setupTestRepository(t)

	nested := filepath.Join(tempDir, "nested", "deep")
	err := os.MkdirAll(nested, 0o755)
	require.NoError(t, err, "failed to create nested directories")

	root, err := InferGitRepoRoot(nested)
	require.NoError(t, err, "InferGitRepoRoot failed on nested path")
	assertPathsEqual(t, root, tempDir)
}

func TestInferGitRepoRoot_NotRepo(t *testing.T) {
	dir := t.TempDir()

	root, err := InferGitRepoRoot(dir)
	require.Error(t, err, "expected error for non-repo directory")
	require.Equal(t, "", root, "expected empty root for error case")
}

// Helper: compare canonicalized paths
func assertPathsEqual(t *testing.T, p1, p2 string) {
	t.Helper()
	p1 = filepath.Clean(p1)
	p2 = filepath.Clean(p2)
	if r1, err := filepath.EvalSymlinks(p1); err == nil {
		p1 = r1
	}
	if r2, err := filepath.EvalSymlinks(p2); err == nil {
		p2 = r2
	}
	require.Equal(t, p1, p2)
}
func setupTestRepository(t *testing.T) (string, string) {
	tempDir := t.TempDir()

	// Initialize a new git repository in the temp directory
	cmd := exec.Command("git", "init", tempDir)
	err := cmd.Run()
	require.NoError(t, err, "failed to initialize git repository")
	setupGitConfig(t, tempDir)

	// Create a remote repository in another temp directory
	remoteDir := t.TempDir()
	cmd = exec.Command("git", "init", "--bare", remoteDir)
	err = cmd.Run()
	require.NoError(t, err, "failed to initialize remote git repository")

	// Add the remote to the local repository
	cmd = exec.Command("git", "-C", tempDir, "remote", "add", "origin", remoteDir)
	err = cmd.Run()
	require.NoError(t, err, "failed to add remote repository")

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

	// Push the initial commit to the remote repository
	cmd = exec.Command("git", "-C", tempDir, "push", "-u", "origin", "HEAD")
	err = cmd.Run()
	require.NoError(t, err, "failed to push initial commit")

	// Run the RunGitStatus function
	status, err := RunGitStatus(tempDir, "origin")
	require.NoError(t, err, "RunGitStatus failed")

	// Validate the status
	require.Equal(t, int32(0), status.LocalCommits, "unexpected local commits")
	require.Equal(t, remoteDir, status.RemoteURL, "unexpected remote URL")
	require.False(t, status.LocalChanges, "unexpected local changes")
	require.Equal(t, int32(0), status.RemoteCommits, "unexpected remote commits")

	return tempDir, remoteDir
}

func createCommit(t *testing.T, repoPath, fileName, fileContent, commitMessage string) {
	filePath := filepath.Join(repoPath, fileName)
	err := os.WriteFile(filePath, []byte(fileContent), 0644)
	require.NoError(t, err, "failed to create file")

	cmd := exec.Command("git", "-C", repoPath, "add", fileName)
	_, err = cmd.Output()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			// This is error msg returned by git when pull fails
			require.NoError(t, err, "failed to stage file"+fmt.Sprintf(": %s", execErr.Stderr))
		}
		require.NoError(t, err, "failed to stage file")
	}

	cmd = exec.Command("git", "-C", repoPath, "commit", "-m", commitMessage)
	_, err = cmd.Output()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			// This is error msg returned by git when pull fails
			require.NoError(t, err, "failed to commit file"+fmt.Sprintf(": %s", execErr.Stderr))
		}
		require.NoError(t, err, "failed to commit file")
	}

}

func createRemoteCommit(t *testing.T, remoteDir, fileName, fileContent, commitMessage string) {
	// Clone the bare repository to a temporary working directory
	workingDir := t.TempDir()
	cmd := exec.Command("git", "clone", remoteDir, workingDir)
	err := cmd.Run()
	require.NoError(t, err, "failed to clone remote repository")
	setupGitConfig(t, workingDir)

	// Create and commit the file in the working directory
	createCommit(t, workingDir, fileName, fileContent, commitMessage)

	// Push the changes back to the remote repository
	cmd = exec.Command("git", "-C", workingDir, "push", "origin", "HEAD")
	err = cmd.Run()
	require.NoError(t, err, "failed to push changes to remote repository")
}

// setupGitConfig sets up the git configuration for the repository at repoPath.
func setupGitConfig(t *testing.T, repoPath string) {
	// Set user name and email for the git repository
	cmd := exec.Command("git", "-C", repoPath, "config", "user.name", "Test User")
	err := cmd.Run()
	require.NoError(t, err, "failed to set user name in git config")

	cmd = exec.Command("git", "-C", repoPath, "config", "user.email", "test@rilldata.com")
	err = cmd.Run()
	require.NoError(t, err, "failed to set user email in git config")
}
