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

func TestStatus(t *testing.T) {
	tempDir, remoteDir := setupRepoWithRemote(t)

	// Test case: Local commits
	createCommit(t, tempDir, "test4.txt", "add a local file", "local commit")

	// Run GitFetch
	require.NoError(t, Fetch(t.Context(), tempDir, nil), "failed to fetch changes from remote repository")
	// Run the Status function again
	status, err := Status(context.Background(), tempDir, "", "origin", "")
	require.NoError(t, err, "Status failed after local commit")

	// Validate the updated status
	require.Equal(t, int32(1), status.LocalCommits, "unexpected local commits after fourth commit")
	require.Equal(t, int32(0), status.RemoteCommits, "unexpected remote commits")

	// Test case: Remote commits
	createRemoteCommit(t, remoteDir, "test5.txt", "add a file in remote", "remote commit")

	// Fetch the latest changes from the remote repository before running Status
	require.NoError(t, Fetch(t.Context(), tempDir, nil), "failed to fetch changes from remote repository")

	// Run the Status function again
	status, err = Status(context.Background(), tempDir, "", "origin", "")
	require.NoError(t, err, "Status failed after local commit")

	// Validate the updated status
	require.Equal(t, int32(1), status.LocalCommits, "unexpected local commits after fourth commit")
	require.Equal(t, int32(1), status.RemoteCommits, "unexpected remote commits")

	// Test case: Untracked files
	filePath := filepath.Join(tempDir, "untracked.txt")
	err = os.WriteFile(filePath, []byte("untracked content"), 0644)
	require.NoError(t, err, "failed to create untracked file")

	status, err = Status(context.Background(), tempDir, "", "origin", "")
	require.NoError(t, err, "Status failed with untracked files")
	require.True(t, status.LocalChanges, "expected local changes due to untracked files")

	// Test case: Staged but uncommitted changes
	err = os.WriteFile(filePath, []byte("staged content"), 0644)
	require.NoError(t, err, "failed to modify file for staging")

	cmd := exec.Command("git", "-C", tempDir, "add", "untracked.txt")
	err = cmd.Run()
	require.NoError(t, err, "failed to stage file")

	status, err = Status(context.Background(), tempDir, "", "origin", "")
	require.NoError(t, err, "Status failed with staged changes")
	require.True(t, status.LocalChanges, "expected local changes due to staged files")

	// Test case: Unstaged changes
	err = os.WriteFile(filePath, []byte("unstaged content"), 0644)
	require.NoError(t, err, "failed to modify file for unstaged changes")

	status, err = Status(context.Background(), tempDir, "", "origin", "")
	require.NoError(t, err, "Status failed with unstaged changes")
	require.True(t, status.LocalChanges, "expected local changes due to unstaged files")
}

// TestStatus_Monorepo tests Status with subpath parameter for monorepo scenarios
func TestStatus_Monorepo(t *testing.T) {
	tempDir, _ := setupMonorepoTestRepository(t)

	// Test case 1: Check initial status of subprojects
	status, err := Status(context.Background(), tempDir, "subproject1", "origin", "")
	require.NoError(t, err, "Status failed for subproject1")
	require.Equal(t, int32(0), status.LocalCommits, "unexpected local commits for subproject1")
	require.Equal(t, int32(0), status.RemoteCommits, "unexpected remote commits for subproject1")
	require.False(t, status.LocalChanges, "unexpected local changes for subproject1")

	// Test case 2: Add local commit to subproject1 only
	createCommit(t, tempDir, "subproject1/local.txt", "local content in subproject1", "subproject1: local commit")
	require.NoError(t, Fetch(t.Context(), tempDir, nil), "failed to fetch changes")

	status, err = Status(context.Background(), tempDir, "subproject1", "origin", "")
	require.NoError(t, err, "Status failed for subproject1 after local commit")
	require.Equal(t, int32(1), status.LocalCommits, "expected 1 local commit for subproject1")
	require.Equal(t, int32(0), status.RemoteCommits, "unexpected remote commits for subproject1")

	// Test case 3: Verify subproject2 is unaffected by subproject1 changes
	status, err = Status(context.Background(), tempDir, "subproject2", "origin", "")
	require.NoError(t, err, "Status failed for subproject2")
	require.Equal(t, int32(0), status.LocalCommits, "subproject2 should have no local commits")
	require.Equal(t, int32(0), status.RemoteCommits, "subproject2 should have no remote commits")
}

// TestStatus_MonorepoLocalChanges tests local changes detection in monorepo subpaths
func TestStatus_MonorepoLocalChanges(t *testing.T) {
	tempDir, _ := setupMonorepoTestRepository(t)

	// Test case 1: Staged changes in subproject1
	stagedFile := filepath.Join(tempDir, "subproject1", "staged.txt")
	err := os.WriteFile(stagedFile, []byte("staged content"), 0644)
	require.NoError(t, err, "failed to create file")

	cmd := exec.Command("git", "-C", tempDir, "add", "subproject1/staged.txt")
	err = cmd.Run()
	require.NoError(t, err, "failed to stage file")

	status, err := Status(context.Background(), tempDir, "subproject1", "origin", "")
	require.NoError(t, err, "Status failed")
	require.True(t, status.LocalChanges, "expected local changes in subproject1")

	// Verify subproject2 is unaffected
	status, err = Status(context.Background(), tempDir, "subproject2", "origin", "")
	require.NoError(t, err, "Status failed")
	require.False(t, status.LocalChanges, "subproject2 should not have local changes")

	// Test case 2: Unstaged changes in subproject2
	existingFile := filepath.Join(tempDir, "subproject2", "file2.txt")
	err = os.WriteFile(existingFile, []byte("modified content"), 0644)
	require.NoError(t, err, "failed to modify file")

	status, err = Status(context.Background(), tempDir, "subproject2", "origin", "")
	require.NoError(t, err, "Status failed")
	require.True(t, status.LocalChanges, "expected local changes in subproject2")

	// Test case 3: Changes outside subpath should not be detected
	outsideFile := filepath.Join(tempDir, "outside.txt")
	err = os.WriteFile(outsideFile, []byte("outside content"), 0644)
	require.NoError(t, err, "failed to create file outside subprojects")

	status, err = Status(context.Background(), tempDir, "subproject1", "origin", "")
	require.NoError(t, err, "Status failed")
	// Should still show only the previously staged change
	require.True(t, status.LocalChanges, "expected local changes from staged file")

	status, err = Status(context.Background(), tempDir, "subproject2", "origin", "")
	require.NoError(t, err, "Status failed")
	// Should still show only the unstaged change to file2.txt
	require.True(t, status.LocalChanges, "expected local changes from modified file")
}

// TestStatus_MonorepoRemoteCommits tests remote commit tracking in monorepo subpaths
func TestStatus_MonorepoRemoteCommits(t *testing.T) {
	tempDir, remoteDir := setupMonorepoTestRepository(t)

	// Test case 1: Multiple remote commits to subproject1, verify isolation
	createRemoteCommit(t, remoteDir, "subproject1/feature1.txt", "feature 1", "subproject1: add feature 1")
	createRemoteCommit(t, remoteDir, "subproject1/feature2.txt", "feature 2", "subproject1: add feature 2")
	require.NoError(t, Fetch(t.Context(), tempDir, nil), "failed to fetch changes")

	status, err := Status(context.Background(), tempDir, "subproject1", "origin", "")
	require.NoError(t, err, "Status failed for subproject1")
	require.Equal(t, int32(2), status.RemoteCommits, "expected 2 remote commits for subproject1")

	status, err = Status(context.Background(), tempDir, "subproject2", "origin", "")
	require.NoError(t, err, "Status failed for subproject2")
	require.Equal(t, int32(0), status.RemoteCommits, "subproject2 should have no remote commits")

	// Test case 2: Mix of local and remote commits in different subprojects
	createCommit(t, tempDir, "subproject1/local.txt", "local content", "subproject1: local commit")
	createRemoteCommit(t, remoteDir, "subproject2/feature3.txt", "feature 3", "subproject2: add feature 3")
	require.NoError(t, Fetch(t.Context(), tempDir, nil), "failed to fetch changes")

	status, err = Status(context.Background(), tempDir, "subproject1", "origin", "")
	require.NoError(t, err, "Status failed for subproject1")
	require.Equal(t, int32(1), status.LocalCommits, "subproject1 should have 1 local commit")
	require.Equal(t, int32(2), status.RemoteCommits, "subproject1 should have 2 remote commits")

	status, err = Status(context.Background(), tempDir, "subproject2", "origin", "")
	require.NoError(t, err, "Status failed for subproject2")
	require.Equal(t, int32(0), status.LocalCommits, "subproject2 should have no local commits")
	require.Equal(t, int32(1), status.RemoteCommits, "subproject2 should have 1 remote commit")

	// Test case 3: Commits outside subpaths don't affect subproject counts
	createRemoteCommit(t, remoteDir, "root-file.txt", "root content", "add root file")
	require.NoError(t, Fetch(t.Context(), tempDir, nil), "failed to fetch changes")

	status, err = Status(context.Background(), tempDir, "subproject1", "origin", "")
	require.NoError(t, err, "Status failed for subproject1")
	require.Equal(t, int32(2), status.RemoteCommits, "subproject1 should still have 2 remote commits")

	status, err = Status(context.Background(), tempDir, "subproject2", "origin", "")
	require.NoError(t, err, "Status failed for subproject2")
	require.Equal(t, int32(1), status.RemoteCommits, "subproject2 should still have 1 remote commit")
}

// TestStatus_ExcludesLocalMergeCommits verifies that merge commits in the local branch
// are not counted as "ahead" commits.
func TestStatus_ExcludesLocalMergeCommits(t *testing.T) {
	tempDir, _ := setupRepoWithRemote(t)
	mainBranch := getCurrentBranch(t, tempDir)

	// Create a feature branch with one commit
	cmd := exec.Command("git", "-C", tempDir, "checkout", "-b", "feature")
	require.NoError(t, cmd.Run(), "failed to create feature branch")
	createCommit(t, tempDir, "feature.txt", "feature content", "feature commit")

	// Back to main and add a commit so the feature merge cannot fast-forward
	cmd = exec.Command("git", "-C", tempDir, "checkout", mainBranch)
	require.NoError(t, cmd.Run(), "failed to switch to main")
	createCommit(t, tempDir, "main.txt", "main content", "main commit")

	// Force a merge commit
	cmd = exec.Command("git", "-C", tempDir, "merge", "--no-ff", "feature", "-m", "merge feature")
	require.NoError(t, cmd.Run(), "failed to merge feature branch")

	require.NoError(t, Fetch(t.Context(), tempDir, nil), "failed to fetch")

	status, err := Status(context.Background(), tempDir, "", "origin", "")
	require.NoError(t, err, "Status failed")

	// History has 3 commits ahead (feature, main, merge), but the merge must be excluded.
	require.Equal(t, int32(2), status.LocalCommits, "merge commit should be excluded from local count")
	require.Equal(t, int32(0), status.RemoteCommits, "unexpected remote commits")
}

// TestStatus_ExcludesRemoteMergeCommits verifies that merge commits on the remote
// are not counted as "behind" commits.
func TestStatus_ExcludesRemoteMergeCommits(t *testing.T) {
	tempDir, remoteDir := setupRepoWithRemote(t)
	mainBranch := getCurrentBranch(t, tempDir)

	createRemoteMergeCommit(t, remoteDir, mainBranch)

	require.NoError(t, Fetch(t.Context(), tempDir, nil), "failed to fetch")

	status, err := Status(context.Background(), tempDir, "", "origin", "")
	require.NoError(t, err, "Status failed")

	require.Equal(t, int32(0), status.LocalCommits, "unexpected local commits")
	require.Equal(t, int32(2), status.RemoteCommits, "merge commit should be excluded from remote count")
}

// TestPull_DiscardsLocalMergeCommits verifies that discardLocal=true correctly resets
// the local branch even when its commits ahead include a merge commit (which would
// previously break the HEAD~N reset strategy).
func TestPull_DiscardsLocalMergeCommits(t *testing.T) {
	tempDir, _ := setupRepoWithRemote(t)
	mainBranch := getCurrentBranch(t, tempDir)

	cmd := exec.Command("git", "-C", tempDir, "checkout", "-b", "feature")
	require.NoError(t, cmd.Run(), "failed to create feature branch")
	createCommit(t, tempDir, "feature.txt", "feature content", "feature commit")

	cmd = exec.Command("git", "-C", tempDir, "checkout", mainBranch)
	require.NoError(t, cmd.Run(), "failed to switch to main")
	createCommit(t, tempDir, "main.txt", "main content", "main commit")

	cmd = exec.Command("git", "-C", tempDir, "merge", "--no-ff", "feature", "-m", "merge feature")
	require.NoError(t, cmd.Run(), "failed to merge feature branch")

	require.NoError(t, Fetch(t.Context(), tempDir, nil), "failed to fetch")

	_, err := Pull(context.Background(), tempDir, true, "", "origin")
	require.NoError(t, err, "Pull with discardLocal failed")

	// After discard, local should fully match remote.
	status, err := Status(context.Background(), tempDir, "", "origin", "")
	require.NoError(t, err, "Status failed after pull")
	require.Equal(t, int32(0), status.LocalCommits, "expected no local commits after discard pull")
	require.Equal(t, int32(0), status.RemoteCommits, "expected no remote commits after pull")
}

func TestPull(t *testing.T) {
	tempDir, remoteDir := setupRepoWithRemote(t)

	// Test case: Pull with no local changes
	output, err := Pull(context.Background(), tempDir, false, "", "origin")
	require.NoError(t, err, "Pull failed with no local changes")
	require.Empty(t, output, "unexpected output from Pull with no local changes")

	// Test case: Pull with local changes (discardLocal = false)
	createCommit(t, tempDir, "local.txt", "local content", "local commit")
	createRemoteCommit(t, remoteDir, "local.txt", "remote content", "remote commit")
	output, err = Pull(context.Background(), tempDir, false, "", "origin")
	if len(output) == 0 && err == nil {
		t.Fatalf("expected Pull to fail with local changes and discardLocal=false, but it succeeded")
	}
	require.Contains(t, output, "Need to specify how to reconcile divergent branches", "unexpected output from Pull with local changes")

	// Test case: Pull with local changes (discardLocal = true)
	output, err = Pull(context.Background(), tempDir, true, "", "origin")
	require.NoError(t, err, "Pull failed with local changes and discardLocal=true")
	require.Empty(t, output, "unexpected output from Pull with discardLocal=true")

	// Test case: Pull with remote changes
	createRemoteCommit(t, remoteDir, "remote.txt", "remote content", "remote commit")
	output, err = Pull(context.Background(), tempDir, false, "", "origin")
	require.NoError(t, err, "Pull failed with remote changes")
	require.Empty(t, output, "unexpected output from Pull with remote changes")
}

func TestPush_NoNewCommits(t *testing.T) {
	tempDir, _ := setupRepoWithRemote(t)
	branch := getCurrentBranch(t, tempDir)

	// Push with no new commits; remote is already up to date — should not error.
	err := Push(context.Background(), tempDir, "origin", branch)
	require.NoError(t, err, "Push should not fail when there are no new commits")
}

func TestInferRepoRoot_InRepoRoot(t *testing.T) {
	tempDir, _ := setupRepoWithRemote(t)

	root, err := InferRepoRoot(tempDir)
	require.NoError(t, err, "InferRepoRoot failed on repo root")
	assertPathsEqual(t, root, tempDir)
}

func TestInferRepoRoot_InNestedDir(t *testing.T) {
	tempDir, _ := setupRepoWithRemote(t)

	nested := filepath.Join(tempDir, "nested", "deep")
	err := os.MkdirAll(nested, 0o755)
	require.NoError(t, err, "failed to create nested directories")

	root, err := InferRepoRoot(nested)
	require.NoError(t, err, "InferRepoRoot failed on nested path")
	assertPathsEqual(t, root, tempDir)
}

func TestInferRepoRoot_NotRepo(t *testing.T) {
	dir := t.TempDir()

	root, err := InferRepoRoot(dir)
	require.Error(t, err, ErrNotAGitRepository)
	require.Equal(t, "", root, "expected empty root for error case")
}

func TestInferRepoRoot_SymlinkToRepoRoot(t *testing.T) {
	tempDir, _ := setupRepoWithRemote(t)

	// Create a separate temp directory for symlinks
	symlinkBase := t.TempDir()
	symlinkDir := filepath.Join(symlinkBase, "repo_symlink")
	err := os.Symlink(tempDir, symlinkDir)
	require.NoError(t, err, "failed to create symlink to repo root")

	root, err := InferRepoRoot(symlinkDir)
	require.NoError(t, err, "InferRepoRoot failed on symlink to repo root")

	// Get the canonical path of the expected repo root
	expectedRoot, err := filepath.EvalSymlinks(tempDir)
	require.NoError(t, err, "failed to resolve symlinks in expected root")

	// Get the canonical path of the actual result
	actualRoot, err := filepath.EvalSymlinks(root)
	require.NoError(t, err, "failed to resolve symlinks in actual root")

	require.Equal(t, expectedRoot, actualRoot, "repo root should match after symlink resolution")
}

func TestInferRepoRoot_SymlinkChain(t *testing.T) {
	tempDir, _ := setupRepoWithRemote(t)

	// Create a separate temp directory for symlinks
	symlinkBase := t.TempDir()
	symlink1 := filepath.Join(symlinkBase, "symlink1")
	symlink2 := filepath.Join(symlinkBase, "symlink2")

	err := os.Symlink(tempDir, symlink1)
	require.NoError(t, err, "failed to create first symlink")

	err = os.Symlink(symlink1, symlink2)
	require.NoError(t, err, "failed to create second symlink")

	root, err := InferRepoRoot(symlink2)
	require.NoError(t, err, "InferRepoRoot failed on symlink chain")

	// Get the canonical path of the expected repo root
	expectedRoot, err := filepath.EvalSymlinks(tempDir)
	require.NoError(t, err, "failed to resolve symlinks in expected root")

	// Get the canonical path of the actual result
	actualRoot, err := filepath.EvalSymlinks(root)
	require.NoError(t, err, "failed to resolve symlinks in actual root")

	require.Equal(t, expectedRoot, actualRoot, "repo root should match after symlink resolution")
}

func TestInferRepoRoot_SymlinkToNonRepo(t *testing.T) {
	nonRepoDir := t.TempDir()

	// Create a symlink to a non-repo directory in a different temp directory
	tempDirParent := t.TempDir()
	symlinkDir := filepath.Join(tempDirParent, "nonrepo_symlink")
	err := os.Symlink(nonRepoDir, symlinkDir)
	require.NoError(t, err, "failed to create symlink to non-repo directory")

	root, err := InferRepoRoot(symlinkDir)
	require.Error(t, err, "expected error when symlink points to non-repo")
	require.ErrorIs(t, err, ErrNotAGitRepository)
	require.Equal(t, "", root, "expected empty root for error case")
}

func TestUpstreamMerge(t *testing.T) {
	tempDir, remoteDir := setupRepoWithRemote(t)

	// Test case 1: Simple upstream merge without conflicts
	createRemoteCommit(t, remoteDir, "upstream1.txt", "upstream content 1", "upstream commit 1")
	require.NoError(t, Fetch(context.Background(), tempDir, nil), "failed to fetch changes")

	branch := getCurrentBranch(t, tempDir)
	err := UpstreamMerge(context.Background(), tempDir, "origin", branch, false)
	require.NoError(t, err, "UpstreamMerge failed")

	// Verify file exists after merge
	assertFileExists(t, tempDir, "upstream1.txt")

	// Test case 2: Upstream merge with local commit (no conflict)
	createCommit(t, tempDir, "local1.txt", "local content 1", "local commit 1")
	createRemoteCommit(t, remoteDir, "upstream2.txt", "upstream content 2", "upstream commit 2")
	require.NoError(t, Fetch(context.Background(), tempDir, nil), "failed to fetch changes")

	err = UpstreamMerge(context.Background(), tempDir, "origin", branch, false)
	require.NoError(t, err, "UpstreamMerge failed with local commits")

	// Both files should exist
	assertFileExists(t, tempDir, "local1.txt")
	assertFileExists(t, tempDir, "upstream2.txt")

	// Test case 3: Conflicting changes - first try without favourLocal (should fail), then with favourLocal (local wins)
	createCommit(t, tempDir, "conflict.txt", "local version", "add conflict file locally")

	createRemoteCommit(t, remoteDir, "conflict.txt", "upstream version", "add conflict file upstream")
	require.NoError(t, Fetch(context.Background(), tempDir, nil), "failed to fetch changes")

	// First try without favourLocal - should fail
	err = UpstreamMerge(context.Background(), tempDir, "origin", branch, false)
	require.Error(t, err, "UpstreamMerge should fail with conflicts and favourLocal=false")
	require.Contains(t, err.Error(), "git merge failed", "expected merge failure error")

	// Reset to before the failed merge attempt
	cmd := exec.Command("git", "-C", tempDir, "merge", "--abort")
	_ = cmd.Run() // Ignore error if no merge in progress

	// Now try with favourLocal=true - should succeed
	err = UpstreamMerge(context.Background(), tempDir, "origin", branch, true)
	require.NoError(t, err, "UpstreamMerge failed with conflicts and favourLocal=true")

	require.Equal(t, "local version", readFile(t, tempDir, "conflict.txt"), "local version should win with favourLocal=true")

	// Test case 4: Mixed changes - remote changes in A and B, local changes in A and C
	// Create local changes to file A and C
	createCommit(t, tempDir, "fileA.txt", "local content A", "add fileA locally")
	createCommit(t, tempDir, "fileC.txt", "local content C", "add fileC locally")

	// Create remote changes to file A and B
	createRemoteCommit(t, remoteDir, "fileA.txt", "upstream content A", "add fileA upstream")
	createRemoteCommit(t, remoteDir, "fileB.txt", "upstream content B", "add fileB upstream")
	require.NoError(t, Fetch(context.Background(), tempDir, nil), "failed to fetch changes")

	err = UpstreamMerge(context.Background(), tempDir, "origin", branch, true)
	require.NoError(t, err, "UpstreamMerge failed with mixed changes")

	// Verify local version of A is preserved
	require.Equal(t, "local content A", readFile(t, tempDir, "fileA.txt"), "local version of A should be preserved")

	// Verify remote version of B is present
	assertFileExists(t, tempDir, "fileB.txt")
	require.Equal(t, "upstream content B", readFile(t, tempDir, "fileB.txt"), "remote version of B should be present")

	// Verify local version of C is preserved
	require.Equal(t, "local content C", readFile(t, tempDir, "fileC.txt"), "local version of C should be preserved")
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
func setupRepoWithRemote(t *testing.T) (string, string) {
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

	// Run the Status function
	status, err := Status(context.Background(), tempDir, "", "origin", "")
	require.NoError(t, err, "Status failed")

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
	status, err := Status(context.Background(), tempDir, "subproject1", "origin", "")
	require.NoError(t, err, "Status failed for subproject1")
	require.Equal(t, int32(0), status.LocalCommits, "unexpected local commits in subproject1")
	require.False(t, status.LocalChanges, "unexpected local changes in subproject1")

	status, err = Status(context.Background(), tempDir, "subproject2", "origin", "")
	require.NoError(t, err, "Status failed for subproject2")
	require.Equal(t, int32(0), status.LocalCommits, "unexpected local commits in subproject2")
	require.False(t, status.LocalChanges, "unexpected local changes in subproject2")

	return tempDir, remoteDir
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

// createRemoteMergeCommit builds a merge commit on the remote by cloning it,
// creating a feature branch and a divergent commit, then merging with --no-ff and pushing.
// After it returns, the remote `branch` has 2 non-merge commits + 1 merge commit
// on top of its previous tip.
func createRemoteMergeCommit(t *testing.T, remoteDir, branch string) {
	workingDir := t.TempDir()
	cmd := exec.Command("git", "clone", remoteDir, workingDir)
	require.NoError(t, cmd.Run(), "failed to clone remote repository")
	setupGitConfig(t, workingDir)

	cmd = exec.Command("git", "-C", workingDir, "checkout", "-b", "remote-feature")
	require.NoError(t, cmd.Run(), "failed to create remote feature branch")
	createCommit(t, workingDir, "remote-feature.txt", "feature", "remote feature commit")

	cmd = exec.Command("git", "-C", workingDir, "checkout", branch)
	require.NoError(t, cmd.Run(), "failed to switch to branch")
	createCommit(t, workingDir, "remote-main.txt", "main", "remote main commit")

	cmd = exec.Command("git", "-C", workingDir, "merge", "--no-ff", "remote-feature", "-m", "remote merge")
	require.NoError(t, cmd.Run(), "failed to create remote merge commit")

	cmd = exec.Command("git", "-C", workingDir, "push", "origin", branch)
	require.NoError(t, cmd.Run(), "failed to push merge to remote")
}

// getCurrentBranch gets the current branch name for the repository at repoPath
func getCurrentBranch(t *testing.T, repoPath string) string {
	t.Helper()
	cmd := exec.Command("git", "-C", repoPath, "branch", "--show-current")
	output, err := cmd.Output()
	require.NoError(t, err, "failed to get current branch")
	return strings.TrimSpace(string(output))
}

// assertFileExists checks if a file exists at the specified path relative to repoPath
func assertFileExists(t *testing.T, repoPath, relativePath string) {
	t.Helper()
	filePath := filepath.Join(repoPath, relativePath)
	_, err := os.Stat(filePath)
	require.NoError(t, err, "file %s should exist", relativePath)
}

// readFile reads and returns the content of a file at the specified path relative to repoPath
func readFile(t *testing.T, repoPath, relativePath string) string {
	t.Helper()
	filePath := filepath.Join(repoPath, relativePath)
	content, err := os.ReadFile(filePath)
	require.NoError(t, err, "failed to read file %s", relativePath)
	return string(content)
}

// TestFetch_UpdatesRemoteTrackingRefs pins the load-bearing fetch refspec: a credentialed fetch
// must update refs/remotes/<remote-name>/* (a bare `git fetch <url>` would only write
// FETCH_HEAD), because Status, Pull, and UpstreamMerge compare against <remote-name>/<branch>.
func TestFetch_UpdatesRemoteTrackingRefs(t *testing.T) {
	ctx := context.Background()
	remote := setupBareRemote(t)
	baseURL := serveRepoOverHTTP(t, remote)

	config := &Config{
		Remote:        baseURL,
		Username:      "user",
		Password:      "SECRETTOKEN123",
		DefaultBranch: "main",
		ManagedRepo:   true,
	}

	path := filepath.Join(t.TempDir(), "clone")
	require.NoError(t, CloneWithConfig(ctx, path, config))

	before, err := Hash(ctx, path, "refs/remotes/__rill_remote/main")
	require.NoError(t, err)

	createRemoteCommit(t, remote, "new.txt", "new content", "remote commit")
	refreshServerInfo(t, remote)

	require.NoError(t, Fetch(ctx, path, config))

	after, err := Hash(ctx, path, "refs/remotes/__rill_remote/main")
	require.NoError(t, err)
	require.NotEqual(t, before, after, "fetch must advance the remote-tracking ref")

	st, err := Status(ctx, path, "", config.RemoteName(), "")
	require.NoError(t, err)
	require.Equal(t, int32(1), st.RemoteCommits)
	require.Equal(t, int32(0), st.LocalCommits)
}

func TestPush_RedactsCredentialsInError(t *testing.T) {
	path := setupTestRepository(t)
	err := Push(context.Background(), path, "https://user:secret-token@host.invalid/org/repo.git", "main")
	require.Error(t, err)
	require.NotContains(t, err.Error(), "secret-token", "credentials must not leak into push errors")
	require.Contains(t, err.Error(), "<redacted>@")
}

func TestPull_RedactsCredentialsInMessage(t *testing.T) {
	local, _ := setupRepoWithRemote(t)
	msg, err := Pull(context.Background(), local, false, "https://user:secret-token@host.invalid/org/repo.git", "origin")
	require.NoError(t, err, "a git-rejected pull must surface as a message, not an error")
	require.NotContains(t, msg, "secret-token", "credentials must not leak into the user-visible pull message")
	require.Contains(t, msg, "<redacted>@")
}
