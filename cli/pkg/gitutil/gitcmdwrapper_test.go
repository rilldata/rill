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
	status, err := RunGitStatus(tempDir, "", "origin")
	require.NoError(t, err, "RunGitStatus failed after local commit")

	// Validate the updated status
	require.Equal(t, int32(1), status.LocalCommits, "unexpected local commits after fourth commit")
	require.Equal(t, int32(0), status.RemoteCommits, "unexpected remote commits")

	// Test case: Remote commits
	createRemoteCommit(t, remoteDir, "test5.txt", "add a file in remote", "remote commit")

	// Fetch the latest changes from the remote repository before running RunGitStatus
	require.NoError(t, GitFetch(t.Context(), tempDir, nil), "failed to fetch changes from remote repository")

	// Run the RunGitStatus function again
	status, err = RunGitStatus(tempDir, "", "origin")
	require.NoError(t, err, "RunGitStatus failed after local commit")

	// Validate the updated status
	require.Equal(t, int32(1), status.LocalCommits, "unexpected local commits after fourth commit")
	require.Equal(t, int32(1), status.RemoteCommits, "unexpected remote commits")

	// Test case: Untracked files
	filePath := filepath.Join(tempDir, "untracked.txt")
	err = os.WriteFile(filePath, []byte("untracked content"), 0644)
	require.NoError(t, err, "failed to create untracked file")

	status, err = RunGitStatus(tempDir, "", "origin")
	require.NoError(t, err, "RunGitStatus failed with untracked files")
	require.True(t, status.LocalChanges, "expected local changes due to untracked files")

	// Test case: Staged but uncommitted changes
	err = os.WriteFile(filePath, []byte("staged content"), 0644)
	require.NoError(t, err, "failed to modify file for staging")

	cmd := exec.Command("git", "-C", tempDir, "add", "untracked.txt")
	err = cmd.Run()
	require.NoError(t, err, "failed to stage file")

	status, err = RunGitStatus(tempDir, "", "origin")
	require.NoError(t, err, "RunGitStatus failed with staged changes")
	require.True(t, status.LocalChanges, "expected local changes due to staged files")

	// Test case: Unstaged changes
	err = os.WriteFile(filePath, []byte("unstaged content"), 0644)
	require.NoError(t, err, "failed to modify file for unstaged changes")

	status, err = RunGitStatus(tempDir, "", "origin")
	require.NoError(t, err, "RunGitStatus failed with unstaged changes")
	require.True(t, status.LocalChanges, "expected local changes due to unstaged files")
}

// TestRunGitStatus_Monorepo tests RunGitStatus with subpath parameter for monorepo scenarios
func TestRunGitStatus_Monorepo(t *testing.T) {
	tempDir, _ := setupMonorepoTestRepository(t)

	// Test case 1: Check initial status of subprojects
	status, err := RunGitStatus(tempDir, "subproject1", "origin")
	require.NoError(t, err, "RunGitStatus failed for subproject1")
	require.Equal(t, int32(0), status.LocalCommits, "unexpected local commits for subproject1")
	require.Equal(t, int32(0), status.RemoteCommits, "unexpected remote commits for subproject1")
	require.False(t, status.LocalChanges, "unexpected local changes for subproject1")

	// Test case 2: Add local commit to subproject1 only
	createCommit(t, tempDir, "subproject1/local.txt", "local content in subproject1", "subproject1: local commit")
	require.NoError(t, GitFetch(t.Context(), tempDir, nil), "failed to fetch changes")

	status, err = RunGitStatus(tempDir, "subproject1", "origin")
	require.NoError(t, err, "RunGitStatus failed for subproject1 after local commit")
	require.Equal(t, int32(1), status.LocalCommits, "expected 1 local commit for subproject1")
	require.Equal(t, int32(0), status.RemoteCommits, "unexpected remote commits for subproject1")

	// Test case 3: Verify subproject2 is unaffected by subproject1 changes
	status, err = RunGitStatus(tempDir, "subproject2", "origin")
	require.NoError(t, err, "RunGitStatus failed for subproject2")
	require.Equal(t, int32(0), status.LocalCommits, "subproject2 should have no local commits")
	require.Equal(t, int32(0), status.RemoteCommits, "subproject2 should have no remote commits")
}

// TestRunGitStatus_MonorepoLocalChanges tests local changes detection in monorepo subpaths
func TestRunGitStatus_MonorepoLocalChanges(t *testing.T) {
	tempDir, _ := setupMonorepoTestRepository(t)

	// Test case 1: Staged changes in subproject1
	stagedFile := filepath.Join(tempDir, "subproject1", "staged.txt")
	err := os.WriteFile(stagedFile, []byte("staged content"), 0644)
	require.NoError(t, err, "failed to create file")

	cmd := exec.Command("git", "-C", tempDir, "add", "subproject1/staged.txt")
	err = cmd.Run()
	require.NoError(t, err, "failed to stage file")

	status, err := RunGitStatus(tempDir, "subproject1", "origin")
	require.NoError(t, err, "RunGitStatus failed")
	require.True(t, status.LocalChanges, "expected local changes in subproject1")

	// Verify subproject2 is unaffected
	status, err = RunGitStatus(tempDir, "subproject2", "origin")
	require.NoError(t, err, "RunGitStatus failed")
	require.False(t, status.LocalChanges, "subproject2 should not have local changes")

	// Test case 2: Unstaged changes in subproject2
	existingFile := filepath.Join(tempDir, "subproject2", "file2.txt")
	err = os.WriteFile(existingFile, []byte("modified content"), 0644)
	require.NoError(t, err, "failed to modify file")

	status, err = RunGitStatus(tempDir, "subproject2", "origin")
	require.NoError(t, err, "RunGitStatus failed")
	require.True(t, status.LocalChanges, "expected local changes in subproject2")

	// Test case 3: Changes outside subpath should not be detected
	outsideFile := filepath.Join(tempDir, "outside.txt")
	err = os.WriteFile(outsideFile, []byte("outside content"), 0644)
	require.NoError(t, err, "failed to create file outside subprojects")

	status, err = RunGitStatus(tempDir, "subproject1", "origin")
	require.NoError(t, err, "RunGitStatus failed")
	// Should still show only the previously staged change
	require.True(t, status.LocalChanges, "expected local changes from staged file")

	status, err = RunGitStatus(tempDir, "subproject2", "origin")
	require.NoError(t, err, "RunGitStatus failed")
	// Should still show only the unstaged change to file2.txt
	require.True(t, status.LocalChanges, "expected local changes from modified file")
}

// TestRunGitStatus_MonorepoRemoteCommits tests remote commit tracking in monorepo subpaths
func TestRunGitStatus_MonorepoRemoteCommits(t *testing.T) {
	tempDir, remoteDir := setupMonorepoTestRepository(t)

	// Test case 1: Multiple remote commits to subproject1, verify isolation
	createRemoteCommit(t, remoteDir, "subproject1/feature1.txt", "feature 1", "subproject1: add feature 1")
	createRemoteCommit(t, remoteDir, "subproject1/feature2.txt", "feature 2", "subproject1: add feature 2")
	require.NoError(t, GitFetch(t.Context(), tempDir, nil), "failed to fetch changes")

	status, err := RunGitStatus(tempDir, "subproject1", "origin")
	require.NoError(t, err, "RunGitStatus failed for subproject1")
	require.Equal(t, int32(2), status.RemoteCommits, "expected 2 remote commits for subproject1")

	status, err = RunGitStatus(tempDir, "subproject2", "origin")
	require.NoError(t, err, "RunGitStatus failed for subproject2")
	require.Equal(t, int32(0), status.RemoteCommits, "subproject2 should have no remote commits")

	// Test case 2: Mix of local and remote commits in different subprojects
	createCommit(t, tempDir, "subproject1/local.txt", "local content", "subproject1: local commit")
	createRemoteCommit(t, remoteDir, "subproject2/feature3.txt", "feature 3", "subproject2: add feature 3")
	require.NoError(t, GitFetch(t.Context(), tempDir, nil), "failed to fetch changes")

	status, err = RunGitStatus(tempDir, "subproject1", "origin")
	require.NoError(t, err, "RunGitStatus failed for subproject1")
	require.Equal(t, int32(1), status.LocalCommits, "subproject1 should have 1 local commit")
	require.Equal(t, int32(2), status.RemoteCommits, "subproject1 should have 2 remote commits")

	status, err = RunGitStatus(tempDir, "subproject2", "origin")
	require.NoError(t, err, "RunGitStatus failed for subproject2")
	require.Equal(t, int32(0), status.LocalCommits, "subproject2 should have no local commits")
	require.Equal(t, int32(1), status.RemoteCommits, "subproject2 should have 1 remote commit")

	// Test case 3: Commits outside subpaths don't affect subproject counts
	createRemoteCommit(t, remoteDir, "root-file.txt", "root content", "add root file")
	require.NoError(t, GitFetch(t.Context(), tempDir, nil), "failed to fetch changes")

	status, err = RunGitStatus(tempDir, "subproject1", "origin")
	require.NoError(t, err, "RunGitStatus failed for subproject1")
	require.Equal(t, int32(2), status.RemoteCommits, "subproject1 should still have 2 remote commits")

	status, err = RunGitStatus(tempDir, "subproject2", "origin")
	require.NoError(t, err, "RunGitStatus failed for subproject2")
	require.Equal(t, int32(1), status.RemoteCommits, "subproject2 should still have 1 remote commit")
}

func TestGitPull(t *testing.T) {
	tempDir, remoteDir := setupTestRepository(t)

	// Test case: Pull with no local changes
	output, err := RunGitPull(context.Background(), tempDir, false, false, "", "origin")
	require.NoError(t, err, "GitPull failed with no local changes")
	require.Empty(t, output, "unexpected output from GitPull with no local changes")

	// Test case: Pull with local changes (discardLocal = false)
	createCommit(t, tempDir, "local.txt", "local content", "local commit")
	createRemoteCommit(t, remoteDir, "local.txt", "remote content", "remote commit")
	output, err = RunGitPull(context.Background(), tempDir, false, false, "", "origin")
	if len(output) == 0 && err == nil {
		t.Fatalf("expected GitPull to fail with local changes and discardLocal=false, but it succeeded")
	}
	require.Contains(t, output, "Need to specify how to reconcile divergent branches", "unexpected output from GitPull with local changes")

	// Test case: Pull with local changes (discardLocal = true)
	output, err = RunGitPull(context.Background(), tempDir, true, false, "", "origin")
	require.NoError(t, err, "GitPull failed with local changes and discardLocal=true")
	require.Empty(t, output, "unexpected output from GitPull with discardLocal=true")

	// Test case: Pull with remote changes
	createRemoteCommit(t, remoteDir, "remote.txt", "remote content", "remote commit")
	output, err = RunGitPull(context.Background(), tempDir, false, false, "", "origin")
	require.NoError(t, err, "GitPull failed with remote changes")
	require.Empty(t, output, "unexpected output from GitPull with remote changes")
}

// TestRunGitPull_Monorepo tests RunGitPull behavior in monorepo scenarios
func TestRunGitPull_Monorepo(t *testing.T) {
	tempDir, remoteDir := setupMonorepoTestRepository(t)

	// Test case 1: Pull with remote commits in one subproject
	createRemoteCommit(t, remoteDir, "subproject1/feature.txt", "feature", "subproject1: add feature")
	require.NoError(t, GitFetch(t.Context(), tempDir, nil), "failed to fetch changes")

	// Verify status before pull
	status, err := RunGitStatus(tempDir, "subproject1", "origin")
	require.NoError(t, err, "RunGitStatus failed")
	require.Equal(t, int32(1), status.RemoteCommits, "expected 1 remote commit before pull")

	// Pull the changes
	output, err := RunGitPull(context.Background(), tempDir, false, false, "", "origin")
	require.NoError(t, err, "RunGitPull failed")
	require.Empty(t, output, "unexpected output from RunGitPull")

	// Verify remote commits are now local
	status, err = RunGitStatus(tempDir, "subproject1", "origin")
	require.NoError(t, err, "RunGitStatus failed after pull")
	require.Equal(t, int32(0), status.RemoteCommits, "expected no remote commits after pull")

	// Test case 2: Pull with divergent branches - local commit in subproject1, remote in subproject2
	createCommit(t, tempDir, "subproject1/local.txt", "local content", "subproject1: local commit")
	createRemoteCommit(t, remoteDir, "subproject2/remote.txt", "remote content", "subproject2: remote commit")
	require.NoError(t, GitFetch(t.Context(), tempDir, nil), "failed to fetch changes")

	// This should fail because of divergent branches (no strategy specified)
	output, err = RunGitPull(context.Background(), tempDir, false, false, "", "origin")
	if len(output) == 0 && err == nil {
		t.Fatalf("expected RunGitPull to fail with divergent branches, but it succeeded")
	}
	require.Contains(t, output, "Need to specify how to reconcile divergent branches", "expected divergent branches error")

	// Test case 3: Pull with preferLocal=true to resolve divergence with local winning
	output, err = RunGitPull(context.Background(), tempDir, false, true, "", "origin")
	require.NoError(t, err, "RunGitPull with preferLocal failed")
	require.Empty(t, output, "unexpected output from RunGitPull with preferLocal")

	// After merge, we now have a merge commit that includes both local and remote changes
	// The merge commit will show as a local commit until pushed
	status, err = RunGitStatus(tempDir, "subproject1", "origin")
	require.NoError(t, err, "RunGitStatus failed for subproject1")
	require.Equal(t, int32(0), status.RemoteCommits, "subproject1 should have no remote commits after merge")
	// Note: LocalCommits will be > 0 because the merge commit hasn't been pushed yet

	status, err = RunGitStatus(tempDir, "subproject2", "origin")
	require.NoError(t, err, "RunGitStatus failed for subproject2")
	require.Equal(t, int32(0), status.RemoteCommits, "subproject2 should have no remote commits")

	// Verify local file in subproject1 still exists (local changes preserved)
	localFilePath := filepath.Join(tempDir, "subproject1", "local.txt")
	_, err = os.Stat(localFilePath)
	require.NoError(t, err, "local file should exist after preferLocal merge")

	content, err := os.ReadFile(localFilePath)
	require.NoError(t, err, "failed to read local file")
	require.Equal(t, "local content", string(content), "local file content should be preserved")

	// Test case 4: Pull with conflicting changes and preferLocal=true (local wins)
	// Modify the same file locally and remotely
	conflictFile := filepath.Join(tempDir, "subproject1", "file1.txt")
	err = os.WriteFile(conflictFile, []byte("local modified content"), 0644)
	require.NoError(t, err, "failed to modify file locally")
	createCommit(t, tempDir, "subproject1/file1.txt", "local modified content", "subproject1: modify file1")

	createRemoteCommit(t, remoteDir, "subproject1/file1.txt", "remote modified content", "subproject1: remote modify file1")
	require.NoError(t, GitFetch(t.Context(), tempDir, nil), "failed to fetch changes")

	// Pull with preferLocal=true should keep local version
	output, err = RunGitPull(context.Background(), tempDir, false, true, "", "origin")
	require.NoError(t, err, "RunGitPull with preferLocal and conflicts failed")
	require.Empty(t, output, "unexpected output from RunGitPull")

	// Verify local version is preserved
	content, err = os.ReadFile(conflictFile)
	require.NoError(t, err, "failed to read conflict file")
	require.Equal(t, "local modified content", string(content), "local version should win in conflict")

	// Test case 5: Pull with discardLocal=true to completely discard local work
	createCommit(t, tempDir, "subproject1/discard.txt", "will be discarded", "subproject1: commit to discard")
	createRemoteCommit(t, remoteDir, "subproject1/keep.txt", "will be kept", "subproject1: commit to keep")
	require.NoError(t, GitFetch(t.Context(), tempDir, nil), "failed to fetch changes")

	output, err = RunGitPull(context.Background(), tempDir, true, false, "", "origin")
	require.NoError(t, err, "RunGitPull with discardLocal failed")
	require.Empty(t, output, "unexpected output from RunGitPull with discardLocal")

	// Verify local commit is discarded
	_, err = os.Stat(filepath.Join(tempDir, "subproject1", "discard.txt"))
	require.True(t, os.IsNotExist(err), "discarded file should not exist")

	// Verify remote commit is pulled
	_, err = os.Stat(filepath.Join(tempDir, "subproject1", "keep.txt"))
	require.NoError(t, err, "remote file should exist after pull")
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
	require.Error(t, err, ErrNotAGitRepository)
	require.Equal(t, "", root, "expected empty root for error case")
}

func TestInferGitRepoRoot_SymlinkToRepoRoot(t *testing.T) {
	tempDir, _ := setupTestRepository(t)

	// Create a separate temp directory for symlinks
	symlinkBase := t.TempDir()
	symlinkDir := filepath.Join(symlinkBase, "repo_symlink")
	err := os.Symlink(tempDir, symlinkDir)
	require.NoError(t, err, "failed to create symlink to repo root")

	root, err := InferGitRepoRoot(symlinkDir)
	require.NoError(t, err, "InferGitRepoRoot failed on symlink to repo root")

	// Get the canonical path of the expected repo root
	expectedRoot, err := filepath.EvalSymlinks(tempDir)
	require.NoError(t, err, "failed to resolve symlinks in expected root")

	// Get the canonical path of the actual result
	actualRoot, err := filepath.EvalSymlinks(root)
	require.NoError(t, err, "failed to resolve symlinks in actual root")

	require.Equal(t, expectedRoot, actualRoot, "repo root should match after symlink resolution")
}

func TestInferGitRepoRoot_SymlinkChain(t *testing.T) {
	tempDir, _ := setupTestRepository(t)

	// Create a separate temp directory for symlinks
	symlinkBase := t.TempDir()
	symlink1 := filepath.Join(symlinkBase, "symlink1")
	symlink2 := filepath.Join(symlinkBase, "symlink2")

	err := os.Symlink(tempDir, symlink1)
	require.NoError(t, err, "failed to create first symlink")

	err = os.Symlink(symlink1, symlink2)
	require.NoError(t, err, "failed to create second symlink")

	root, err := InferGitRepoRoot(symlink2)
	require.NoError(t, err, "InferGitRepoRoot failed on symlink chain")

	// Get the canonical path of the expected repo root
	expectedRoot, err := filepath.EvalSymlinks(tempDir)
	require.NoError(t, err, "failed to resolve symlinks in expected root")

	// Get the canonical path of the actual result
	actualRoot, err := filepath.EvalSymlinks(root)
	require.NoError(t, err, "failed to resolve symlinks in actual root")

	require.Equal(t, expectedRoot, actualRoot, "repo root should match after symlink resolution")
}

func TestInferGitRepoRoot_SymlinkToNonRepo(t *testing.T) {
	nonRepoDir := t.TempDir()

	// Create a symlink to a non-repo directory in a different temp directory
	tempDirParent := t.TempDir()
	symlinkDir := filepath.Join(tempDirParent, "nonrepo_symlink")
	err := os.Symlink(nonRepoDir, symlinkDir)
	require.NoError(t, err, "failed to create symlink to non-repo directory")

	root, err := InferGitRepoRoot(symlinkDir)
	require.Error(t, err, "expected error when symlink points to non-repo")
	require.ErrorIs(t, err, ErrNotAGitRepository)
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
	status, err := RunGitStatus(tempDir, "", "origin")
	require.NoError(t, err, "RunGitStatus failed")

	// Validate the status
	require.Equal(t, int32(0), status.LocalCommits, "unexpected local commits")
	require.Equal(t, remoteDir, status.RemoteURL, "unexpected remote URL")
	require.False(t, status.LocalChanges, "unexpected local changes")
	require.Equal(t, int32(0), status.RemoteCommits, "unexpected remote commits")

	return tempDir, remoteDir
}

// setupMonorepoTestRepository creates a test repository with a monorepo structure
// It creates two subprojects: subproject1 and subproject2, each with initial files
func setupMonorepoTestRepository(t *testing.T) (string, string) {
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

	// Create monorepo structure with multiple subprojects
	subproject1Path := filepath.Join(tempDir, "subproject1")
	subproject2Path := filepath.Join(tempDir, "subproject2")
	err = os.MkdirAll(subproject1Path, 0755)
	require.NoError(t, err, "failed to create subproject1 directory")
	err = os.MkdirAll(subproject2Path, 0755)
	require.NoError(t, err, "failed to create subproject2 directory")

	// Create initial files in subproject1
	file1 := filepath.Join(subproject1Path, "file1.txt")
	err = os.WriteFile(file1, []byte("content of file1 in subproject1"), 0644)
	require.NoError(t, err, "failed to create file1 in subproject1")

	// Create initial files in subproject2
	file2 := filepath.Join(subproject2Path, "file2.txt")
	err = os.WriteFile(file2, []byte("content of file2 in subproject2"), 0644)
	require.NoError(t, err, "failed to create file2 in subproject2")

	// Add root level README
	readmePath := filepath.Join(tempDir, "README.md")
	err = os.WriteFile(readmePath, []byte("# Monorepo Test\n"), 0644)
	require.NoError(t, err, "failed to create README")

	// Stage all files
	cmd = exec.Command("git", "-C", tempDir, "add", ".")
	err = cmd.Run()
	require.NoError(t, err, "failed to stage files")

	// Commit all files
	cmd = exec.Command("git", "-C", tempDir, "commit", "-m", "initial monorepo commit")
	_, err = cmd.Output()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			require.NoError(t, err, "failed to commit files: "+string(execErr.Stderr))
		}
		require.NoError(t, err, "failed to commit files")
	}

	// Push the initial commit to the remote repository
	cmd = exec.Command("git", "-C", tempDir, "push", "-u", "origin", "HEAD")
	err = cmd.Run()
	require.NoError(t, err, "failed to push initial commit")

	// Verify the initial status for both subprojects
	status, err := RunGitStatus(tempDir, "subproject1", "origin")
	require.NoError(t, err, "RunGitStatus failed for subproject1")
	require.Equal(t, int32(0), status.LocalCommits, "unexpected local commits in subproject1")
	require.False(t, status.LocalChanges, "unexpected local changes in subproject1")

	status, err = RunGitStatus(tempDir, "subproject2", "origin")
	require.NoError(t, err, "RunGitStatus failed for subproject2")
	require.Equal(t, int32(0), status.LocalCommits, "unexpected local commits in subproject2")
	require.False(t, status.LocalChanges, "unexpected local changes in subproject2")

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
