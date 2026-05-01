package admin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGitRepo_pullInner(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func(t *testing.T, localDir, remoteURL string) *gitRepo
		setupRemote func(t *testing.T, remoteDir string)
		force       bool
		expectError bool
		validate    func(t *testing.T, repo *gitRepo, localDir string)
	}{
		{
			name: "clone fresh repository when local doesn't exist",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				// Remove the local directory to simulate no existing repo
				require.NoError(t, os.RemoveAll(localDir))
				return &gitRepo{
					h:             &Handle{logger: zap.NewNop()},
					repoDir:       localDir,
					remoteURL:     remoteURL,
					defaultBranch: "main",
					primaryBranch: "main",
					subpath:       "",
					managedRepo:   true,
				}
			},
			setupRemote: func(t *testing.T, remoteDir string) {
				// Remote already has initial commit from setupTestGitRepository
			},
			force:       false,
			expectError: false,
			validate: func(t *testing.T, repo *gitRepo, localDir string) {
				// Verify repository exists and is on correct branch
				verifyCurrentBranch(t, localDir, "main")
			},
		},
		{
			name: "pull changes on default branch in non-editable mode",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				// Clone the repository first
				_, err := git.PlainClone(localDir, false, &git.CloneOptions{
					URL:           remoteURL,
					RemoteName:    "origin",
					ReferenceName: plumbing.ReferenceName("refs/heads/main"),
					SingleBranch:  true,
				})
				require.NoError(t, err)
				setupGitConfig(t, localDir) // Ensure git config is set up
				return &gitRepo{
					h:             &Handle{logger: zap.NewNop()},
					repoDir:       localDir,
					remoteURL:     remoteURL,
					defaultBranch: "main",
					primaryBranch: "main",
					subpath:       "",
					managedRepo:   true,
				}
			},
			setupRemote: func(t *testing.T, remoteDir string) {
				// Add a new commit to the remote
				createRemoteCommit(t, remoteDir, "new_file.txt", "new content", "Add new file")
			},
			force:       false,
			expectError: false,
			validate: func(t *testing.T, repo *gitRepo, localDir string) {
				// Verify repository is on correct branch
				verifyCurrentBranch(t, localDir, "main")

				// Verify the new file exists
				newFilePath := filepath.Join(localDir, "new_file.txt")
				content, err := os.ReadFile(newFilePath)
				require.NoError(t, err)
				require.Equal(t, "new content", string(content))
			},
		},
		{
			name: "create and switch to edit branch in editable mode",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				// Remove the local directory to simulate no existing repo
				require.NoError(t, os.RemoveAll(localDir))
				return newEditableGitRepo(localDir, remoteURL, "edit-branch", "main", "")
			},
			setupRemote: func(t *testing.T, remoteDir string) {
				// No additional remote changes needed
			},
			force:       false,
			expectError: false,
			validate: func(t *testing.T, repo *gitRepo, localDir string) {
				// Verify we're on the edit branch
				verifyCurrentBranch(t, localDir, "edit-branch")
			},
		},
		{
			name: "switch to edit branch when already on edit branch",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				// Clone and create edit branch
				repo, err := git.PlainClone(localDir, false, &git.CloneOptions{
					URL:           remoteURL,
					RemoteName:    "origin",
					ReferenceName: plumbing.ReferenceName("refs/heads/main"),
					SingleBranch:  false,
				})
				require.NoError(t, err)
				setupGitConfig(t, localDir) // Ensure git config is set up

				// Create and switch to edit branch
				worktree, err := repo.Worktree()
				require.NoError(t, err)
				err = worktree.Checkout(&git.CheckoutOptions{
					Branch: plumbing.ReferenceName("refs/heads/edit-branch"),
					Create: true,
				})
				require.NoError(t, err)

				// Make a local commit on edit branch
				filePath := filepath.Join(localDir, "edit_change.txt")
				err = os.WriteFile(filePath, []byte("edit content"), 0644)
				require.NoError(t, err)
				_, err = worktree.Add("edit_change.txt")
				require.NoError(t, err)
				_, err = worktree.Commit("Edit branch change", &git.CommitOptions{
					Author: &object.Signature{
						Name:  "Test User",
						Email: "test@example.com",
						When:  time.Now(),
					},
				})
				require.NoError(t, err)

				return newEditableGitRepo(localDir, remoteURL, "edit-branch", "main", "")
			},
			setupRemote: func(t *testing.T, remoteDir string) {
				// Create the edit branch on remote with different content
				createRemoteBranch(t, remoteDir, "edit-branch", "remote_edit.txt", "remote edit content", "Remote edit change")
			},
			force:       true, // Force to handle potential conflicts
			expectError: false,
			validate: func(t *testing.T, repo *gitRepo, localDir string) {
				// Verify we're on the edit branch and it has been reset to remote
				verifyCurrentBranch(t, localDir, "edit-branch")

				// The remote file should exist (local changes discarded due to force reset)
				remoteFilePath := filepath.Join(localDir, "remote_edit.txt")
				content, err := os.ReadFile(remoteFilePath)
				require.NoError(t, err)
				require.Equal(t, "remote edit content", string(content))
			},
		},
		{
			name: "force pull discards local changes",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				// Clone repository in non-editable mode
				_, err := git.PlainClone(localDir, false, &git.CloneOptions{
					URL:           remoteURL,
					RemoteName:    "origin",
					ReferenceName: plumbing.ReferenceName("refs/heads/main"),
					SingleBranch:  true,
				})
				require.NoError(t, err)
				setupGitConfig(t, localDir) // Ensure git config is set up

				return &gitRepo{
					h:             &Handle{logger: zap.NewNop()},
					repoDir:       localDir,
					remoteURL:     remoteURL,
					defaultBranch: "main",
					primaryBranch: "main",
					subpath:       "",
					managedRepo:   true,
				}
			},
			setupRemote: func(t *testing.T, remoteDir string) {
				// Add conflicting changes to remote
				createRemoteCommit(t, remoteDir, "test1.txt", "updated remote content", "Update test1.txt remotely")
			},
			force:       false, // Will be forced to true due to non-editable mode
			expectError: false,
			validate: func(t *testing.T, repo *gitRepo, localDir string) {
				// Verify repository is on correct branch
				verifyCurrentBranch(t, localDir, "main")

				// Verify remote changes are present
				test1Path := filepath.Join(localDir, "test1.txt")
				content, err := os.ReadFile(test1Path)
				require.NoError(t, err)
				require.Equal(t, "updated remote content", string(content))
			},
		},
		{
			name: "editable mode - pushes newly created default branch to remote when it doesn't exist",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				require.NoError(t, os.RemoveAll(localDir))
				return newEditableGitRepo(localDir, remoteURL, "edit-branch", "main", "")
			},
			setupRemote: func(t *testing.T, remoteDir string) {
				// Remote only has main; edit-branch does not exist yet.
			},
			force:       false,
			expectError: false,
			validate: func(t *testing.T, repo *gitRepo, localDir string) {
				verifyCurrentBranch(t, localDir, "edit-branch")

				// Verify edit-branch was pushed to remote.
				cmd := exec.Command("git", "ls-remote", "--heads", repo.remoteURL, "edit-branch")
				output, err := cmd.Output()
				require.NoError(t, err)
				require.NotEmpty(t, output, "edit-branch should have been pushed to remote")
			},
		},
		{
			name: "switch primary branch from main to rename (remote only has main initially)",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				// Remove the local directory to simulate no existing repo
				require.NoError(t, os.RemoveAll(localDir))
				return &gitRepo{
					h:             &Handle{logger: zap.NewNop()},
					repoDir:       localDir,
					remoteURL:     remoteURL,
					defaultBranch: "main",
					primaryBranch: "main",
					subpath:       "",
					managedRepo:   true,
				}
			},
			setupRemote: func(t *testing.T, remoteDir string) {
				// Initial pull happens with only main branch
				// After first pull, we'll create rename branch
			},
			force:       false,
			expectError: false,
			validate: func(t *testing.T, repo *gitRepo, localDir string) {
				// First pull should succeed with main branch
				verifyCurrentBranch(t, localDir, "main")

				// Now create rename branch on remote
				createRemoteBranch(t, repo.remoteURL, "rename", "rename_file.txt", "rename content", "Create rename branch")

				// Update primary branch and pull again
				repo.primaryBranch = "rename"
				ctx := context.Background()
				err := repo.pullInner(ctx, false, false)
				require.NoError(t, err)

				// Verify we're still on main (default branch unchanged)
				verifyCurrentBranch(t, localDir, "main")

				// Now change default branch to rename and pull again
				repo.defaultBranch = "rename"
				err = repo.pullInner(ctx, false, false)
				require.NoError(t, err)

				// Verify we've switched to rename branch
				verifyCurrentBranch(t, localDir, "rename")

				// Verify rename branch file exists
				renameFilePath := filepath.Join(localDir, "rename_file.txt")
				content, err := os.ReadFile(renameFilePath)
				require.NoError(t, err)
				require.Equal(t, "rename content", string(content))
			},
		},
		{
			name: "switch primary branch from main to rename (remote has both branches)",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				// Remove the local directory to simulate no existing repo
				require.NoError(t, os.RemoveAll(localDir))
				return &gitRepo{
					h:             &Handle{logger: zap.NewNop()},
					repoDir:       localDir,
					remoteURL:     remoteURL,
					defaultBranch: "main",
					primaryBranch: "main",
					subpath:       "",
					managedRepo:   true,
				}
			},
			setupRemote: func(t *testing.T, remoteDir string) {
				// Create rename branch on remote before first pull
				createRemoteBranch(t, remoteDir, "rename", "rename_file.txt", "rename content", "Create rename branch")
			},
			force:       false,
			expectError: false,
			validate: func(t *testing.T, repo *gitRepo, localDir string) {
				// First pull should succeed with main branch
				verifyCurrentBranch(t, localDir, "main")

				// Update primary branch and pull again
				repo.primaryBranch = "rename"
				ctx := context.Background()
				err := repo.pullInner(ctx, false, false)
				require.NoError(t, err)

				// Verify we're still on main (default branch unchanged)
				verifyCurrentBranch(t, localDir, "main")

				// Verify rename branch file doesn't exist on main branch
				renameFilePath := filepath.Join(localDir, "rename_file.txt")
				_, err = os.Stat(renameFilePath)
				require.Error(t, err, "rename_file.txt should not exist on main branch")

				// Now change default branch to rename and pull again
				repo.defaultBranch = "rename"
				err = repo.pullInner(ctx, false, false)
				require.NoError(t, err)

				// Verify we've switched to rename branch
				verifyCurrentBranch(t, localDir, "rename")

				// Verify rename branch file now exists
				content, err := os.ReadFile(renameFilePath)
				require.NoError(t, err)
				require.Equal(t, "rename content", string(content))
			},
		},
		{
			name: "pull preserves git-ignored files",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				// Clone the repository
				_, err := git.PlainClone(localDir, false, &git.CloneOptions{
					URL:           remoteURL,
					RemoteName:    "origin",
					ReferenceName: plumbing.ReferenceName("refs/heads/main"),
					SingleBranch:  true,
				})
				require.NoError(t, err)
				setupGitConfig(t, localDir)

				// Create a git-ignored file that should survive the pull
				err = os.WriteFile(filepath.Join(localDir, "data.db"), []byte("database content"), 0644)
				require.NoError(t, err)

				return &gitRepo{
					h:             &Handle{logger: zap.NewNop()},
					repoDir:       localDir,
					remoteURL:     remoteURL,
					defaultBranch: "main",
					primaryBranch: "main",
					subpath:       "",
					managedRepo:   true,
				}
			},
			setupRemote: func(t *testing.T, remoteDir string) {
				createRemoteCommit(t, remoteDir, ".gitignore", "*.db\n", "Add .gitignore")
				createRemoteCommit(t, remoteDir, "new_file.txt", "new content", "Add new file")
			},
			force:       false,
			expectError: false,
			validate: func(t *testing.T, repo *gitRepo, localDir string) {
				verifyCurrentBranch(t, localDir, "main")

				// Verify git-ignored file is preserved
				content, err := os.ReadFile(filepath.Join(localDir, "data.db"))
				require.NoError(t, err, "git-ignored file data.db should be preserved")
				require.Equal(t, "database content", string(content))

				// Verify the new file from remote also exists
				content, err = os.ReadFile(filepath.Join(localDir, "new_file.txt"))
				require.NoError(t, err)
				require.Equal(t, "new content", string(content))
			},
		},
		{
			name: "editable mode - primary branch changes on existing repo",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				// Remove the local directory to simulate no existing repo
				require.NoError(t, os.RemoveAll(localDir))
				return newEditableGitRepo(localDir, remoteURL, "edit-branch", "main", "")
			},
			setupRemote: func(t *testing.T, remoteDir string) {
				// Only main branch exists initially
			},
			force:       true,
			expectError: false,
			validate: func(t *testing.T, repo *gitRepo, localDir string) {
				// First pull (clone) succeeded — edit-branch created from main
				verifyCurrentBranch(t, localDir, "edit-branch")

				// Set git identity so merge commits can be created (required on CI where no global config exists).
				setupGitConfig(t, localDir)

				// Create "new-primary" branch on remote (simulates a primary branch rename)
				createRemoteBranch(t, repo.remoteURL, "new-primary", "new_file.txt", "new content", "Create new-primary")

				// Simulate primary branch change
				repo.primaryBranch = "new-primary"

				// Second pull uses the existing-repo fetch path.
				ctx := context.Background()
				err := repo.pullInner(ctx, false, false)
				require.NoError(t, err, "pullInner should succeed after primary branch change")

				// Verify the new primary branch content was merged into edit-branch
				newFilePath := filepath.Join(localDir, "new_file.txt")
				content, err := os.ReadFile(newFilePath)
				require.NoError(t, err, "new_file.txt should exist after merge from new primary branch")
				require.Equal(t, "new content", string(content))
			},
		},
		{
			name: "editable mode - force=false - conflicting local changes returns error",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				_, err := git.PlainClone(localDir, false, &git.CloneOptions{
					URL:           remoteURL,
					RemoteName:    "origin",
					ReferenceName: plumbing.ReferenceName("refs/heads/main"),
					SingleBranch:  false,
				})
				require.NoError(t, err)
				setupGitConfig(t, localDir)

				// Create edit-branch and push it to remote so the remote tracking ref exists.
				require.NoError(t, execCommand(exec.Command("git", "-C", localDir, "checkout", "-b", "edit-branch")))
				require.NoError(t, execCommand(exec.Command("git", "-C", localDir, "push", "origin", "edit-branch")))

				// Commit a local change to test1.txt on edit-branch (do NOT push).
				err = os.WriteFile(filepath.Join(localDir, "test1.txt"), []byte("local edit content"), 0644)
				require.NoError(t, err)
				require.NoError(t, execCommand(exec.Command("git", "-C", localDir, "add", "test1.txt")))
				require.NoError(t, execCommand(exec.Command("git", "-C", localDir, "commit", "-m", "Local edit change")))

				return newEditableGitRepo(localDir, remoteURL, "edit-branch", "main", "")
			},
			setupRemote: func(t *testing.T, remoteDir string) {
				// Push a conflicting change to the same file on edit-branch.
				createRemoteCommitOnBranch(t, remoteDir, "edit-branch", "test1.txt", "remote edit content", "Remote edit change")
			},
			force:       false,
			expectError: true,
			validate:    nil,
		},
		{
			name: "editable mode - force=false - non-conflicting local changes merged successfully",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				_, err := git.PlainClone(localDir, false, &git.CloneOptions{
					URL:           remoteURL,
					RemoteName:    "origin",
					ReferenceName: plumbing.ReferenceName("refs/heads/main"),
					SingleBranch:  false,
				})
				require.NoError(t, err)
				setupGitConfig(t, localDir)

				// Create edit-branch and push it to remote.
				require.NoError(t, execCommand(exec.Command("git", "-C", localDir, "checkout", "-b", "edit-branch")))
				require.NoError(t, execCommand(exec.Command("git", "-C", localDir, "push", "origin", "edit-branch")))

				// Commit a local change to a unique file (no overlap with remote change).
				err = os.WriteFile(filepath.Join(localDir, "local_change.txt"), []byte("local only content"), 0644)
				require.NoError(t, err)
				require.NoError(t, execCommand(exec.Command("git", "-C", localDir, "add", "local_change.txt")))
				require.NoError(t, execCommand(exec.Command("git", "-C", localDir, "commit", "-m", "Local only change")))

				return newEditableGitRepo(localDir, remoteURL, "edit-branch", "main", "")
			},
			setupRemote: func(t *testing.T, remoteDir string) {
				// Push a non-conflicting change (different file) to edit-branch on remote.
				createRemoteCommitOnBranch(t, remoteDir, "edit-branch", "remote_change.txt", "remote only content", "Remote only change")
			},
			force:       false,
			expectError: false,
			validate: func(t *testing.T, repo *gitRepo, localDir string) {
				verifyCurrentBranch(t, localDir, "edit-branch")

				// Both local and remote changes should be present after the merge.
				localContent, err := os.ReadFile(filepath.Join(localDir, "local_change.txt"))
				require.NoError(t, err)
				require.Equal(t, "local only content", string(localContent))

				remoteContent, err := os.ReadFile(filepath.Join(localDir, "remote_change.txt"))
				require.NoError(t, err)
				require.Equal(t, "remote only content", string(remoteContent))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			localDir := t.TempDir()
			remoteDir := setupTestGitRepository(t)

			// Setup the gitRepo instance
			repo := tt.setupRepo(t, localDir, remoteDir)

			// Setup any remote changes
			tt.setupRemote(t, remoteDir)

			// Execute pullInner
			ctx := context.Background()
			err := repo.pullInner(ctx, true, tt.force)

			// Verify error expectation
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Run validation if no error expected
			if !tt.expectError && tt.validate != nil {
				tt.validate(t, repo, localDir)
			}
		})
	}
}

func TestGitRepo_commitToDefaultBranch(t *testing.T) {
	tests := []struct {
		name         string
		setupRepo    func(t *testing.T, localDir, remoteURL string) *gitRepo
		setupChanges func(t *testing.T, localDir string)
		setupRemote  func(t *testing.T, remoteDir string)
		message      string
		force        bool
		expectError  bool
		validate     func(t *testing.T, repo *gitRepo, localDir, remoteDir string)
	}{
		{
			name: "force=false: commit and push to default branch with no remote conflicts",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				cloneAndCreateRemoteEditBranch(t, localDir, remoteURL)
				return newEditableGitRepo(localDir, remoteURL, "edit-branch", "main", "")
			},
			setupChanges: func(t *testing.T, localDir string) {
				require.NoError(t, os.WriteFile(filepath.Join(localDir, "new_feature.txt"), []byte("new feature content"), 0644))
			},
			setupRemote: func(t *testing.T, remoteDir string) {},
			message:     "Add new feature",
			force:       false,
			expectError: false,
			validate: func(t *testing.T, repo *gitRepo, localDir, remoteDir string) {
				verifyRemoteBranchFile(t, remoteDir, "edit-branch", "new_feature.txt", "new feature content")
			},
		},
		{
			name: "force=false: non-conflicting remote changes are merged in via theirs strategy",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				cloneAndCreateRemoteEditBranch(t, localDir, remoteURL)
				return newEditableGitRepo(localDir, remoteURL, "edit-branch", "main", "")
			},
			setupChanges: func(t *testing.T, localDir string) {
				require.NoError(t, os.WriteFile(filepath.Join(localDir, "local_only.txt"), []byte("local only"), 0644))
			},
			setupRemote: func(t *testing.T, remoteDir string) {
				createRemoteCommitOnBranch(t, remoteDir, "edit-branch", "remote_only.txt", "remote only", "Remote-only change")
			},
			message:     "Add local only",
			force:       false,
			expectError: false,
			validate: func(t *testing.T, repo *gitRepo, localDir, remoteDir string) {
				verifyRemoteBranchFile(t, remoteDir, "edit-branch", "local_only.txt", "local only")
				verifyRemoteBranchFile(t, remoteDir, "edit-branch", "remote_only.txt", "remote only")
			},
		},
		{
			name: "force=false: conflicting remote changes return an error",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				cloneAndCreateRemoteEditBranch(t, localDir, remoteURL)
				return newEditableGitRepo(localDir, remoteURL, "edit-branch", "main", "")
			},
			setupChanges: func(t *testing.T, localDir string) {
				require.NoError(t, os.WriteFile(filepath.Join(localDir, "test1.txt"), []byte("local edit"), 0644))
			},
			setupRemote: func(t *testing.T, remoteDir string) {
				createRemoteCommitOnBranch(t, remoteDir, "edit-branch", "test1.txt", "remote edit", "Remote edit on same file")
			},
			message:     "Local edit",
			force:       false,
			expectError: true,
			validate: func(t *testing.T, repo *gitRepo, localDir, remoteDir string) {
				// Remote edit-branch is unchanged: nothing was pushed.
				verifyRemoteBranchFile(t, remoteDir, "edit-branch", "test1.txt", "remote edit")
			},
		},
		{
			name: "force=true: conflicting remote changes are overridden via ours strategy",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				cloneAndCreateRemoteEditBranch(t, localDir, remoteURL)
				return newEditableGitRepo(localDir, remoteURL, "edit-branch", "main", "")
			},
			setupChanges: func(t *testing.T, localDir string) {
				require.NoError(t, os.WriteFile(filepath.Join(localDir, "test1.txt"), []byte("local force edit"), 0644))
			},
			setupRemote: func(t *testing.T, remoteDir string) {
				createRemoteCommitOnBranch(t, remoteDir, "edit-branch", "test1.txt", "remote edit", "Remote edit on same file")
			},
			message:     "Local force edit",
			force:       true,
			expectError: false,
			validate: func(t *testing.T, repo *gitRepo, localDir, remoteDir string) {
				// ours (local) wins on conflicts when force=true
				verifyRemoteBranchFile(t, remoteDir, "edit-branch", "test1.txt", "local force edit")
			},
		},
		{
			name: "force=true with subpath: pushes successfully when no remote conflicts",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				cloneAndCreateRemoteEditBranch(t, localDir, remoteURL)
				require.NoError(t, os.MkdirAll(filepath.Join(localDir, "sub"), 0755))
				return newEditableGitRepo(localDir, remoteURL, "edit-branch", "main", "sub")
			},
			setupChanges: func(t *testing.T, localDir string) {
				require.NoError(t, os.WriteFile(filepath.Join(localDir, "sub", "feature.txt"), []byte("subpath content"), 0644))
			},
			setupRemote: func(t *testing.T, remoteDir string) {},
			message:     "Add subpath feature",
			force:       true,
			expectError: false,
			validate: func(t *testing.T, repo *gitRepo, localDir, remoteDir string) {
				verifyRemoteBranchFile(t, remoteDir, "edit-branch", "sub/feature.txt", "subpath content")
			},
		},
		{
			name: "defaultBranch == primaryBranch: commit and push works",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				_, err := git.PlainClone(localDir, false, &git.CloneOptions{
					URL:           remoteURL,
					RemoteName:    "origin",
					ReferenceName: plumbing.ReferenceName("refs/heads/main"),
					SingleBranch:  false,
				})
				require.NoError(t, err)
				setupGitConfig(t, localDir)
				return newEditableGitRepo(localDir, remoteURL, "main", "main", "")
			},
			setupChanges: func(t *testing.T, localDir string) {
				require.NoError(t, os.WriteFile(filepath.Join(localDir, "feature.txt"), []byte("feature content"), 0644))
			},
			setupRemote: func(t *testing.T, remoteDir string) {},
			message:     "Add feature",
			force:       false,
			expectError: false,
			validate: func(t *testing.T, repo *gitRepo, localDir, remoteDir string) {
				verifyRemoteBranchFile(t, remoteDir, "main", "feature.txt", "feature content")
			},
		},
		{
			name: "no changes results in no error and no remote update",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				cloneAndCreateRemoteEditBranch(t, localDir, remoteURL)
				return newEditableGitRepo(localDir, remoteURL, "edit-branch", "main", "")
			},
			setupChanges: func(t *testing.T, localDir string) {},
			setupRemote:  func(t *testing.T, remoteDir string) {},
			message:      "Empty commit",
			force:        false,
			expectError:  false,
			validate: func(t *testing.T, repo *gitRepo, localDir, remoteDir string) {
				// Verify no extra commits beyond the branch creation point.
				workingDir := t.TempDir()
				require.NoError(t, execCommand(exec.Command("git", "clone", "-b", "edit-branch", remoteDir, workingDir)))
				out, err := exec.Command("git", "-C", workingDir, "rev-list", "--count", "HEAD").Output()
				require.NoError(t, err)
				require.Equal(t, "1\n", string(out))
			},
		},
		{
			name: "error when repository is not editable",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				_, err := git.PlainClone(localDir, false, &git.CloneOptions{
					URL:           remoteURL,
					RemoteName:    "origin",
					ReferenceName: plumbing.ReferenceName("refs/heads/main"),
					SingleBranch:  true,
				})
				require.NoError(t, err)
				setupGitConfig(t, localDir)
				return &gitRepo{
					h:             &Handle{logger: zap.NewNop()},
					repoDir:       localDir,
					remoteURL:     remoteURL,
					defaultBranch: "main",
					primaryBranch: "main",
					subpath:       "",
					managedRepo:   true,
					// editableDepl is false: not editable
				}
			},
			setupChanges: func(t *testing.T, localDir string) {
				require.NoError(t, os.WriteFile(filepath.Join(localDir, "should_not_commit.txt"), []byte("content"), 0644))
			},
			setupRemote: func(t *testing.T, remoteDir string) {},
			message:     "Should fail",
			force:       false,
			expectError: true,
			validate:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			localDir := t.TempDir()
			remoteDir := setupTestGitRepository(t)

			repo := tt.setupRepo(t, localDir, remoteDir)
			tt.setupChanges(t, localDir)
			tt.setupRemote(t, remoteDir)

			ctx := context.Background()
			err := repo.commitToDefaultBranch(ctx, tt.message, tt.force)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.validate != nil {
				tt.validate(t, repo, localDir, remoteDir)
			}
		})
	}
}

func TestGitRepo_mergeToBranch(t *testing.T) {
	tests := []struct {
		name         string
		setupRepo    func(t *testing.T, localDir, remoteURL string) *gitRepo
		setupChanges func(t *testing.T, localDir string)
		setupRemote  func(t *testing.T, remoteDir string)
		branch       string
		force        bool
		expectError  bool
		validate     func(t *testing.T, repo *gitRepo, localDir, remoteDir string)
	}{
		{
			name: "error when repository is not editable",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				_, err := git.PlainClone(localDir, false, &git.CloneOptions{
					URL:           remoteURL,
					RemoteName:    "origin",
					ReferenceName: plumbing.ReferenceName("refs/heads/main"),
					SingleBranch:  true,
				})
				require.NoError(t, err)
				setupGitConfig(t, localDir)
				return &gitRepo{
					h:             &Handle{logger: zap.NewNop()},
					repoDir:       localDir,
					remoteURL:     remoteURL,
					defaultBranch: "main",
					primaryBranch: "main",
					editableDepl:  false,
				}
			},
			setupChanges: func(t *testing.T, localDir string) {},
			setupRemote:  func(t *testing.T, remoteDir string) {},
			branch:       "main",
			force:        false,
			expectError:  true,
			validate:     nil,
		},
		{
			name: "no local changes returns nil without pushing",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				cloneAndCreateRemoteEditBranch(t, localDir, remoteURL)
				return newEditableGitRepo(localDir, remoteURL, "edit-branch", "main", "")
			},
			setupChanges: func(t *testing.T, localDir string) {},
			setupRemote:  func(t *testing.T, remoteDir string) {},
			branch:       "main",
			force:        false,
			expectError:  false,
			validate: func(t *testing.T, repo *gitRepo, localDir, remoteDir string) {
				verifyCurrentBranch(t, localDir, "edit-branch")
			},
		},
		{
			name: "defaultBranch == branch: commits and pushes changes",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				_, err := git.PlainClone(localDir, false, &git.CloneOptions{
					URL:           remoteURL,
					RemoteName:    "origin",
					ReferenceName: plumbing.ReferenceName("refs/heads/main"),
					SingleBranch:  false,
				})
				require.NoError(t, err)
				setupGitConfig(t, localDir)
				return newEditableGitRepo(localDir, remoteURL, "main", "main", "")
			},
			setupChanges: func(t *testing.T, localDir string) {
				require.NoError(t, os.WriteFile(filepath.Join(localDir, "new_feature.txt"), []byte("new feature content"), 0644))
			},
			setupRemote: func(t *testing.T, remoteDir string) {},
			branch:      "main",
			force:       false,
			expectError: false,
			validate: func(t *testing.T, repo *gitRepo, localDir, remoteDir string) {
				verifyRemoteBranchFile(t, remoteDir, "main", "new_feature.txt", "new feature content")
			},
		},
		{
			name: "merge to different branch, force=false, no conflicts: changes appear on target branch",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				cloneAndCreateRemoteEditBranch(t, localDir, remoteURL)
				return newEditableGitRepo(localDir, remoteURL, "edit-branch", "main", "")
			},
			setupChanges: func(t *testing.T, localDir string) {
				require.NoError(t, os.WriteFile(filepath.Join(localDir, "feature.txt"), []byte("feature content"), 0644))
			},
			setupRemote: func(t *testing.T, remoteDir string) {},
			branch:      "main",
			force:       false,
			expectError: false,
			validate: func(t *testing.T, repo *gitRepo, localDir, remoteDir string) {
				verifyRemoteBranchFile(t, remoteDir, "main", "feature.txt", "feature content")
				verifyRemoteBranchFile(t, remoteDir, "edit-branch", "feature.txt", "feature content")
				verifyCurrentBranch(t, localDir, "edit-branch")
			},
		},
		{
			name: "merge to different branch, force=false, with conflicts: merge aborted, nothing pushed",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				cloneAndCreateRemoteEditBranch(t, localDir, remoteURL)
				return newEditableGitRepo(localDir, remoteURL, "edit-branch", "main", "")
			},
			setupChanges: func(t *testing.T, localDir string) {
				require.NoError(t, os.WriteFile(filepath.Join(localDir, "test1.txt"), []byte("local content"), 0644))
			},
			setupRemote: func(t *testing.T, remoteDir string) {
				createRemoteCommit(t, remoteDir, "test1.txt", "remote content", "Conflicting change on main")
			},
			branch:      "main",
			force:       false,
			expectError: false,
			validate: func(t *testing.T, repo *gitRepo, localDir, remoteDir string) {
				// Remote main is unchanged: no push happened.
				verifyRemoteBranchFile(t, remoteDir, "main", "test1.txt", "remote content")

				// We are back on edit-branch.
				verifyCurrentBranch(t, localDir, "edit-branch")

				// The committed local change is still intact on edit-branch.
				content, err := os.ReadFile(filepath.Join(localDir, "test1.txt"))
				require.NoError(t, err)
				require.Equal(t, "local content", string(content))

				// Repo is in a clean state: no pending merge (no MERGE_HEAD file).
				_, err = os.Stat(filepath.Join(localDir, ".git", "MERGE_HEAD"))
				require.True(t, os.IsNotExist(err), "expected no MERGE_HEAD after aborted merge")
			},
		},
		{
			name: "merge to different branch, force=true, with conflicts: edit-branch wins via theirs strategy",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				cloneAndCreateRemoteEditBranch(t, localDir, remoteURL)
				return newEditableGitRepo(localDir, remoteURL, "edit-branch", "main", "")
			},
			setupChanges: func(t *testing.T, localDir string) {
				require.NoError(t, os.WriteFile(filepath.Join(localDir, "test1.txt"), []byte("local force content"), 0644))
			},
			setupRemote: func(t *testing.T, remoteDir string) {
				createRemoteCommit(t, remoteDir, "test1.txt", "remote content", "Conflicting change on main")
			},
			branch:      "main",
			force:       true,
			expectError: false,
			validate: func(t *testing.T, repo *gitRepo, localDir, remoteDir string) {
				// "theirs" = incoming branch (edit-branch) wins
				verifyRemoteBranchFile(t, remoteDir, "main", "test1.txt", "local force content")
				verifyCurrentBranch(t, localDir, "edit-branch")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			localDir := t.TempDir()
			remoteDir := setupTestGitRepository(t)

			repo := tt.setupRepo(t, localDir, remoteDir)
			tt.setupChanges(t, localDir)
			tt.setupRemote(t, remoteDir)

			ctx := context.Background()
			err := repo.mergeToBranch(ctx, tt.branch, tt.force)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.validate != nil {
				tt.validate(t, repo, localDir, remoteDir)
			}
		})
	}
}

func TestEnsureGitConfig(t *testing.T) {
	withCleanGitEnv(t)

	repo := t.TempDir()
	require.NoError(t, execCommand(exec.Command("git", "-C", repo, "init")))
	require.NoError(t, ensureGitConfig(repo, "user.name", "Test User"))
	out, err := exec.Command("git", "-C", repo, "config", "--local", "--get", "user.name").CombinedOutput()
	require.NoError(t, err)
	require.Equal(t, "Test User\n", string(out))
}

func newEditableGitRepo(localDir, remoteURL, defaultBranch, primaryBranch, subpath string) *gitRepo {
	return &gitRepo{
		h:             &Handle{logger: zap.NewNop()},
		repoDir:       localDir,
		remoteURL:     remoteURL,
		defaultBranch: defaultBranch,
		editableDepl:  true,
		primaryBranch: primaryBranch,
		subpath:       subpath,
		managedRepo:   true,
	}
}

// cloneAndCreateRemoteEditBranch clones the remote, creates "edit-branch" locally, and pushes it to remote.
func cloneAndCreateRemoteEditBranch(t *testing.T, localDir, remoteURL string) {
	t.Helper()
	_, err := git.PlainClone(localDir, false, &git.CloneOptions{
		URL:           remoteURL,
		RemoteName:    "origin",
		ReferenceName: plumbing.ReferenceName("refs/heads/main"),
		SingleBranch:  false,
	})
	require.NoError(t, err)
	setupGitConfig(t, localDir)
	require.NoError(t, execCommand(exec.Command("git", "-C", localDir, "checkout", "-b", "edit-branch")))
	require.NoError(t, execCommand(exec.Command("git", "-C", localDir, "push", "origin", "edit-branch")))
}

// verifyRemoteBranchFile clones the given branch from remote and verifies that the file has the expected content.
func verifyRemoteBranchFile(t *testing.T, remoteDir, branch, relPath, expectedContent string) {
	t.Helper()
	workingDir := t.TempDir()
	require.NoError(t, execCommand(exec.Command("git", "clone", "-b", branch, remoteDir, workingDir)))
	content, err := os.ReadFile(filepath.Join(workingDir, relPath))
	require.NoError(t, err, "expected file %q to exist on remote branch %q", relPath, branch)
	require.Equal(t, expectedContent, string(content))
}

// setupTestGitRepository creates a bare Git repository with initial content for testing
func setupTestGitRepository(t *testing.T) string {
	// Create bare repository
	remoteDir := t.TempDir()
	cmd := exec.Command("git", "init", "--bare", remoteDir)
	err := cmd.Run()
	require.NoError(t, err, "failed to initialize bare git repository")

	// Set the default branch to main in the bare repository
	cmd = exec.Command("git", "-C", remoteDir, "symbolic-ref", "HEAD", "refs/heads/main")
	err = cmd.Run()
	require.NoError(t, err, "failed to set default branch to main")

	// Create a temporary working directory to add initial content
	workingDir := t.TempDir()
	cmd = exec.Command("git", "clone", remoteDir, workingDir)
	err = cmd.Run()
	require.NoError(t, err, "failed to clone bare repository")

	// Setup git config
	setupGitConfig(t, workingDir)

	// Create initial files
	for i := 1; i <= 3; i++ {
		filePath := filepath.Join(workingDir, "test"+string(rune('0'+i))+".txt")
		content := "content of file " + string(rune('0'+i))
		err = os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err, "failed to create test file")
	}

	// Add and commit files
	cmd = exec.Command("git", "-C", workingDir, "add", ".")
	err = cmd.Run()
	require.NoError(t, err, "failed to stage files")

	cmd = exec.Command("git", "-C", workingDir, "commit", "-m", "Initial commit")
	err = cmd.Run()
	require.NoError(t, err, "failed to commit files")

	// Set the default branch to main before pushing
	cmd = exec.Command("git", "-C", workingDir, "branch", "-M", "main")
	err = cmd.Run()
	require.NoError(t, err, "failed to rename branch to main")

	// Push to bare repository
	cmd = exec.Command("git", "-C", workingDir, "push", "origin", "main")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "failed to push initial commit "+string(output))

	return remoteDir
}

// createRemoteCommit creates a new commit in the remote repository
func createRemoteCommit(t *testing.T, remoteDir, fileName, content, commitMessage string) {
	// Clone to temporary directory
	workingDir := t.TempDir()
	cmd := exec.Command("git", "clone", remoteDir, workingDir)
	err := cmd.Run()
	require.NoError(t, err, "failed to clone repository")

	setupGitConfig(t, workingDir)

	// Create/modify file
	filePath := filepath.Join(workingDir, fileName)
	err = os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err, "failed to create file")

	// Add, commit, and push
	cmd = exec.Command("git", "-C", workingDir, "add", fileName)
	err = cmd.Run()
	require.NoError(t, err, "failed to add file")

	cmd = exec.Command("git", "-C", workingDir, "commit", "-m", commitMessage)
	err = cmd.Run()
	require.NoError(t, err, "failed to commit file")

	cmd = exec.Command("git", "-C", workingDir, "push", "origin", "main")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "failed to push commit "+string(output))
}

// createRemoteCommitOnBranch creates a new commit on an existing branch in the remote repository.
func createRemoteCommitOnBranch(t *testing.T, remoteDir, branchName, fileName, content, commitMessage string) {
	t.Helper()
	workingDir := t.TempDir()
	require.NoError(t, execCommand(exec.Command("git", "clone", remoteDir, workingDir)), "failed to clone repository")

	setupGitConfig(t, workingDir)

	require.NoError(t, execCommand(exec.Command("git", "-C", workingDir, "checkout", branchName)), "failed to checkout branch")

	require.NoError(t, os.WriteFile(filepath.Join(workingDir, fileName), []byte(content), 0644), "failed to write file")

	require.NoError(t, execCommand(exec.Command("git", "-C", workingDir, "add", fileName)), "failed to stage file")
	require.NoError(t, execCommand(exec.Command("git", "-C", workingDir, "commit", "-m", commitMessage)), "failed to commit file")
	require.NoError(t, execCommand(exec.Command("git", "-C", workingDir, "push", "origin", branchName)), "failed to push commit")
}

// createRemoteBranch creates a new branch with content in the remote repository
func createRemoteBranch(t *testing.T, remoteDir, branchName, fileName, content, commitMessage string) {
	// Clone to temporary directory
	workingDir := t.TempDir()
	cmd := exec.Command("git", "clone", remoteDir, workingDir)
	err := cmd.Run()
	require.NoError(t, err, "failed to clone repository")

	setupGitConfig(t, workingDir)

	// Create and switch to new branch
	cmd = exec.Command("git", "-C", workingDir, "checkout", "-b", branchName)
	err = cmd.Run()
	require.NoError(t, err, "failed to create branch")

	// Create/modify file
	filePath := filepath.Join(workingDir, fileName)
	err = os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err, "failed to create file")

	// Add, commit, and push
	cmd = exec.Command("git", "-C", workingDir, "add", fileName)
	err = cmd.Run()
	require.NoError(t, err, "failed to add file")

	cmd = exec.Command("git", "-C", workingDir, "commit", "-m", commitMessage)
	err = cmd.Run()
	require.NoError(t, err, "failed to commit file")

	cmd = exec.Command("git", "-C", workingDir, "push", "origin", branchName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Git push output: %s", string(output))
	}
	require.NoError(t, err, "failed to push branch")
}

// setupGitConfig sets up git configuration for testing
func setupGitConfig(t *testing.T, repoPath string) {
	require.NoError(t, execCommand(exec.Command("git", "-C", repoPath, "config", "user.name", "Test User")))
	require.NoError(t, execCommand(exec.Command("git", "-C", repoPath, "config", "user.email", "test@example.com")))
}

func execCommand(cmd *exec.Cmd) error {
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %s, output: %s", err, string(out))
	}
	return nil
}

// verifyCurrentBranch verifies that the repository is currently on the expected branch
func verifyCurrentBranch(t *testing.T, repoPath, expectedBranch string) {
	repo, err := git.PlainOpen(repoPath)
	require.NoError(t, err, "failed to open repository")

	head, err := repo.Head()
	require.NoError(t, err, "failed to get HEAD")

	require.Equal(t, expectedBranch, head.Name().Short(), "unexpected branch")
}

func withCleanGitEnv(t *testing.T) {
	t.Helper()
	empty := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", empty)
	t.Setenv("GIT_CONFIG_SYSTEM", empty)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1") // belt-and-suspenders for older gits
}
