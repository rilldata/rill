package cmdutil

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/pkg/printer"
	"github.com/stretchr/testify/require"
)

func minimalHelper() *Helper {
	return &Helper{
		Printer:     printer.NewPrinter(printer.FormatHuman),
		Interactive: false,
	}
}

func TestCommitAndSafePush_NoRemoteCommits(t *testing.T) {
	tempDir, _ := setupTestRepository(t)
	h := minimalHelper()

	// Create local changes
	createCommit(t, tempDir, "new-file.txt", "new content", "add new file")

	config := &gitutil.Config{
		Remote:        filepath.Join(tempDir, ".git"),
		DefaultBranch: getCurrentBranch(t, tempDir),
		Subpath:       "",
	}

	err := h.CommitAndSafePush(context.Background(), tempDir, config, "test commit", author, "1")
	require.NoError(t, err, "CommitAndSafePush should succeed with no remote commits")
	// verify file exists locally and was pushed
	assertFileExists(t, tempDir, "new-file.txt")
}

func TestCommitAndSafePush_WithRemoteCommits_DefaultChoice1(t *testing.T) {
	tempDir, remoteDir := setupTestRepository(t)
	h := minimalHelper()

	// Create local changes (uncommitted)
	createFile(t, tempDir, "local.txt", "local content")

	// Create remote changes (different file to avoid conflicts)
	createRemoteCommit(t, remoteDir, "remote.txt", "remote content", "add remote file")

	config := &gitutil.Config{
		Remote:        remoteDir,
		DefaultBranch: getCurrentBranch(t, tempDir),
		Subpath:       "",
	}

	err := h.CommitAndSafePush(context.Background(), tempDir, config, "test commit", author, "1")
	require.NoError(t, err, "CommitAndSafePush should succeed with choice 1 (merge)")

	// Verify both files exist
	assertFileExists(t, tempDir, "local.txt")
	assertFileExists(t, tempDir, "remote.txt")
}

func TestCommitAndSafePush_WithRemoteCommits_DefaultChoice2(t *testing.T) {
	tempDir, remoteDir := setupTestRepository(t)
	h := minimalHelper()

	// Create local changes (uncommitted)
	createFile(t, tempDir, "local.txt", "local content")

	// Create remote changes (different file)
	createRemoteCommit(t, remoteDir, "remote.txt", "remote content", "add remote file")

	config := &gitutil.Config{
		Remote:        remoteDir,
		DefaultBranch: getCurrentBranch(t, tempDir),
		Subpath:       "",
	}

	err := h.CommitAndSafePush(context.Background(), tempDir, config, "test commit", author, "2")
	require.NoError(t, err, "CommitAndSafePush should succeed with choice 2 (overwrite)")

	// Verify both files exist after merge with favourLocal
	assertFileExists(t, tempDir, "local.txt")
	assertFileExists(t, tempDir, "remote.txt")
}

func TestCommitAndSafePush_WithRemoteCommits_DefaultChoice3(t *testing.T) {
	tempDir, remoteDir := setupTestRepository(t)
	h := minimalHelper()

	// Create local changes (uncommitted)
	createFile(t, tempDir, "local.txt", "local content")

	// Create remote changes
	createRemoteCommit(t, remoteDir, "remote.txt", "remote content", "add remote file")

	config := &gitutil.Config{
		Remote:        remoteDir,
		DefaultBranch: getCurrentBranch(t, tempDir),
		Subpath:       "",
	}

	err := h.CommitAndSafePush(context.Background(), tempDir, config, "test commit", author, "3")
	require.Error(t, err, "CommitAndSafePush should fail with choice 3 (abort)")
	require.Contains(t, err.Error(), "aborting deploy")
}

func TestCommitAndSafePush_ConflictingChanges_Choice1(t *testing.T) {
	tempDir, remoteDir := setupTestRepository(t)
	h := minimalHelper()

	// Create conflicting changes - same file, different content
	createFile(t, tempDir, "conflict.txt", "local version")
	createRemoteCommit(t, remoteDir, "conflict.txt", "remote version", "add conflict file remotely")

	config := &gitutil.Config{
		Remote:        remoteDir,
		DefaultBranch: getCurrentBranch(t, tempDir),
		Subpath:       "",
	}

	err := h.CommitAndSafePush(context.Background(), tempDir, config, "test commit", author, "1")
	require.Error(t, err, "CommitAndSafePush should fail with merge conflicts and choice 1")
	require.Contains(t, err.Error(), "failed to sync with remote")
}

func TestCommitAndSafePush_ConflictingChanges_Choice2(t *testing.T) {
	tempDir, remoteDir := setupTestRepository(t)
	h := minimalHelper()

	// Create conflicting changes - same file, different content
	// Create and commit local version first
	createCommit(t, tempDir, "conflict.txt", "local version", "add conflict file locally")
	// Then create uncommitted changes to the same file
	createFile(t, tempDir, "conflict.txt", "local version modified")
	createRemoteCommit(t, remoteDir, "conflict.txt", "remote version", "add conflict file remotely")

	config := &gitutil.Config{
		Remote:        remoteDir,
		DefaultBranch: getCurrentBranch(t, tempDir),
		Subpath:       "",
	}

	err := h.CommitAndSafePush(context.Background(), tempDir, config, "test commit", author, "2")
	require.NoError(t, err, "CommitAndSafePush should succeed with choice 2 (favourLocal)")

	// Verify local version wins
	content := readFile(t, tempDir, "conflict.txt")
	require.Equal(t, "local version modified", content, "local version should win with choice 2")
}

func TestCommitAndSafePush_WithSubpath(t *testing.T) {
	tempDir, remoteDir := setupMonorepoTestRepository(t)
	h := minimalHelper()

	// Create local changes in subproject1 (uncommitted)
	createFile(t, tempDir, "subproject1/local.txt", "local content")

	// Create remote changes in subproject1
	createRemoteCommit(t, remoteDir, "subproject1/remote.txt", "remote content", "add remote file to subproject1")

	config := &gitutil.Config{
		Remote:        remoteDir,
		DefaultBranch: getCurrentBranch(t, tempDir),
		Subpath:       "subproject1",
	}

	err := h.CommitAndSafePush(context.Background(), tempDir, config, "test commit", author, "1")
	require.NoError(t, err, "CommitAndSafePush should succeed with subpath")

	// Verify both files exist in subproject1
	assertFileExists(t, tempDir, "subproject1/local.txt")
	assertFileExists(t, tempDir, "subproject1/remote.txt")
}

func TestCommitAndSafePush_SubpathIsolation(t *testing.T) {
	tempDir, remoteDir := setupMonorepoTestRepository(t)
	h := minimalHelper()

	// Create local changes in subproject1 (uncommitted)
	createFile(t, tempDir, "subproject1/local.txt", "local content")

	// Create remote changes in subproject2 (different subpath)
	createRemoteCommit(t, remoteDir, "subproject2/remote.txt", "remote content", "add remote file to subproject2")

	config := &gitutil.Config{
		Remote:        remoteDir,
		DefaultBranch: getCurrentBranch(t, tempDir),
		Subpath:       "subproject1",
	}

	// Should succeed because remote commits in subproject2 don't affect subproject1
	err := h.CommitAndSafePush(context.Background(), tempDir, config, "test commit", author, "1")
	require.NoError(t, err, "CommitAndSafePush should succeed when remote commits are in different subpath")

	// Verify local file exists in subproject1 and subproject2
	assertFileExists(t, tempDir, "subproject1/local.txt")
	assertFileExists(t, tempDir, "subproject2/remote.txt")
}

// Helper functions

func createFile(t *testing.T, repoPath, fileName, fileContent string) {
	t.Helper()
	filePath := filepath.Join(repoPath, fileName)

	// Create directory if needed
	dir := filepath.Dir(filePath)
	if dir != repoPath {
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err, "failed to create directory")
	}

	err := os.WriteFile(filePath, []byte(fileContent), 0644)
	require.NoError(t, err, "failed to create file")
}

func setupTestRepository(t *testing.T) (string, string) {
	t.Helper()
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

	// Create and commit an initial file
	initialFile := filepath.Join(tempDir, "README.md")
	err = os.WriteFile(initialFile, []byte("# Test Repo"), 0644)
	require.NoError(t, err, "failed to create initial file")

	cmd = exec.Command("git", "-C", tempDir, "add", "README.md")
	err = cmd.Run()
	require.NoError(t, err, "failed to stage initial file")

	cmd = exec.Command("git", "-C", tempDir, "commit", "-m", "initial commit")
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

	// Set remote to support fetching with fully qualified URL
	err = gitutil.SetRemote(tempDir, &gitutil.Config{
		Remote:        remoteDir,
		DefaultBranch: getCurrentBranch(t, tempDir),
	})
	require.NoError(t, err, "failed to set remote")

	return tempDir, remoteDir
}

func setupMonorepoTestRepository(t *testing.T) (string, string) {
	t.Helper()
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
	err = os.WriteFile(file1, []byte("content of file1"), 0644)
	require.NoError(t, err, "failed to create file1 in subproject1")

	// Create initial files in subproject2
	file2 := filepath.Join(subproject2Path, "file2.txt")
	err = os.WriteFile(file2, []byte("content of file2"), 0644)
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

	// Set remote to support fetching with fully qualified URL
	err = gitutil.SetRemote(tempDir, &gitutil.Config{
		Remote:        remoteDir,
		DefaultBranch: getCurrentBranch(t, tempDir),
	})
	require.NoError(t, err, "failed to set remote")

	return tempDir, remoteDir
}

func createCommit(t *testing.T, repoPath, fileName, fileContent, commitMessage string) {
	t.Helper()
	filePath := filepath.Join(repoPath, fileName)

	// Create directory if needed
	dir := filepath.Dir(filePath)
	if dir != repoPath {
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err, "failed to create directory")
	}

	err := os.WriteFile(filePath, []byte(fileContent), 0644)
	require.NoError(t, err, "failed to create file")

	cmd := exec.Command("git", "-C", repoPath, "add", fileName)
	_, err = cmd.Output()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			require.NoError(t, err, "failed to stage file: "+string(execErr.Stderr))
		}
		require.NoError(t, err, "failed to stage file")
	}

	cmd = exec.Command("git", "-C", repoPath, "commit", "-m", commitMessage)
	_, err = cmd.Output()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			require.NoError(t, err, "failed to commit file: "+string(execErr.Stderr))
		}
		require.NoError(t, err, "failed to commit file")
	}
}

func createRemoteCommit(t *testing.T, remoteDir, fileName, fileContent, commitMessage string) {
	t.Helper()
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

func setupGitConfig(t *testing.T, repoPath string) {
	t.Helper()
	// Set user name and email for the git repository
	cmd := exec.Command("git", "-C", repoPath, "config", "user.name", "Test User")
	err := cmd.Run()
	require.NoError(t, err, "failed to set user name in git config")

	cmd = exec.Command("git", "-C", repoPath, "config", "user.email", "test@rilldata.com")
	err = cmd.Run()
	require.NoError(t, err, "failed to set user email in git config")
}

func getCurrentBranch(t *testing.T, repoPath string) string {
	t.Helper()
	cmd := exec.Command("git", "-C", repoPath, "branch", "--show-current")
	output, err := cmd.Output()
	require.NoError(t, err, "failed to get current branch")
	return strings.TrimSpace(string(output))
}

func assertFileExists(t *testing.T, repoPath, relativePath string) {
	t.Helper()
	filePath := filepath.Join(repoPath, relativePath)
	_, err := os.Stat(filePath)
	require.NoError(t, err, "file %s should exist", relativePath)
}

func readFile(t *testing.T, repoPath, relativePath string) string {
	t.Helper()
	filePath := filepath.Join(repoPath, relativePath)
	content, err := os.ReadFile(filePath)
	require.NoError(t, err, "failed to read file %s", relativePath)
	return string(content)
}

var author = &object.Signature{
	Name:  "Test User",
	Email: "test@rilldata.com",
	When:  time.Now(),
}
