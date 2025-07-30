package admin

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

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
					editBranch:    "", // Non-editable
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
				gitRepo, err := git.PlainOpen(localDir)
				require.NoError(t, err)
				head, err := gitRepo.Head()
				require.NoError(t, err)
				require.Equal(t, "main", head.Name().Short())
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
				return &gitRepo{
					h:             &Handle{logger: zap.NewNop()},
					repoDir:       localDir,
					remoteURL:     remoteURL,
					defaultBranch: "main",
					editBranch:    "", // Non-editable
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
				gitRepo, err := git.PlainOpen(localDir)
				require.NoError(t, err)
				head, err := gitRepo.Head()
				require.NoError(t, err)
				require.Equal(t, "main", head.Name().Short())

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
				// Clone the repository first
				_, err := git.PlainClone(localDir, false, &git.CloneOptions{
					URL:           remoteURL,
					RemoteName:    "origin",
					ReferenceName: plumbing.ReferenceName("refs/heads/main"),
					SingleBranch:  false, // Allow multiple branches for editable mode
				})
				require.NoError(t, err)
				return &gitRepo{
					h:             &Handle{logger: zap.NewNop()},
					repoDir:       localDir,
					remoteURL:     remoteURL,
					defaultBranch: "main",
					editBranch:    "edit-branch", // Editable
					subpath:       "",
					managedRepo:   true,
				}
			},
			setupRemote: func(t *testing.T, remoteDir string) {
				// No additional remote changes needed
			},
			force:       false,
			expectError: false,
			validate: func(t *testing.T, repo *gitRepo, localDir string) {
				// Verify we're on the edit branch
				gitRepo, err := git.PlainOpen(localDir)
				require.NoError(t, err)
				head, err := gitRepo.Head()
				require.NoError(t, err)
				require.Equal(t, "edit-branch", head.Name().Short())
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
					},
				})
				require.NoError(t, err)

				return &gitRepo{
					h:             &Handle{logger: zap.NewNop()},
					repoDir:       localDir,
					remoteURL:     remoteURL,
					defaultBranch: "main",
					editBranch:    "edit-branch",
					subpath:       "",
					managedRepo:   true,
				}
			},
			setupRemote: func(t *testing.T, remoteDir string) {
				// Create the edit branch on remote with different content
				createRemoteBranch(t, remoteDir, "edit-branch", "remote_edit.txt", "remote edit content", "Remote edit change")
			},
			force:       true, // Force to handle potential conflicts
			expectError: false,
			validate: func(t *testing.T, repo *gitRepo, localDir string) {
				// Verify we're on the edit branch and it has been reset to remote
				gitRepo, err := git.PlainOpen(localDir)
				require.NoError(t, err)
				head, err := gitRepo.Head()
				require.NoError(t, err)
				require.Equal(t, "edit-branch", head.Name().Short())

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

				return &gitRepo{
					h:             &Handle{logger: zap.NewNop()},
					repoDir:       localDir,
					remoteURL:     remoteURL,
					defaultBranch: "main",
					editBranch:    "", // Non-editable, so force should always be true
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
				gitRepo, err := git.PlainOpen(localDir)
				require.NoError(t, err)
				head, err := gitRepo.Head()
				require.NoError(t, err)
				require.Equal(t, "main", head.Name().Short())

				// Verify remote changes are present
				test1Path := filepath.Join(localDir, "test1.txt")
				content, err := os.ReadFile(test1Path)
				require.NoError(t, err)
				require.Equal(t, "updated remote content", string(content))
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
			err := repo.pullInner(ctx, tt.force)

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

// setupTestGitRepository creates a bare Git repository with initial content for testing
func setupTestGitRepository(t *testing.T) string {
	// Create bare repository
	remoteDir := t.TempDir()
	cmd := exec.Command("git", "init", "--bare", remoteDir)
	err := cmd.Run()
	require.NoError(t, err, "failed to initialize bare git repository")

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
	cmd := exec.Command("git", "-C", repoPath, "config", "user.name", "Test User")
	err := cmd.Run()
	require.NoError(t, err, "failed to set user name")

	cmd = exec.Command("git", "-C", repoPath, "config", "user.email", "test@example.com")
	err = cmd.Run()
	require.NoError(t, err, "failed to set user email")
}

func TestGitRepo_commitAndPushToDefaultBranch(t *testing.T) {
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
			name: "commit and push changes to default branch without conflicts",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				// Clone repository and create edit branch
				repo, err := git.PlainClone(localDir, false, &git.CloneOptions{
					URL:           remoteURL,
					RemoteName:    "origin",
					ReferenceName: plumbing.ReferenceName("refs/heads/main"),
					SingleBranch:  false,
				})
				require.NoError(t, err)

				// Create and switch to edit branch
				worktree, err := repo.Worktree()
				require.NoError(t, err)
				err = worktree.Checkout(&git.CheckoutOptions{
					Branch: plumbing.ReferenceName("refs/heads/edit-branch"),
					Create: true,
				})
				require.NoError(t, err)

				return &gitRepo{
					h:             &Handle{logger: zap.NewNop()},
					repoDir:       localDir,
					remoteURL:     remoteURL,
					defaultBranch: "main",
					editBranch:    "edit-branch",
					subpath:       "",
					managedRepo:   true,
				}
			},
			setupChanges: func(t *testing.T, localDir string) {
				// Make changes in the edit branch
				filePath := filepath.Join(localDir, "new_feature.txt")
				err := os.WriteFile(filePath, []byte("new feature content"), 0644)
				require.NoError(t, err)
			},
			setupRemote: func(t *testing.T, remoteDir string) {
				// No additional remote changes
			},
			message:     "Add new feature",
			force:       false,
			expectError: false,
			validate: func(t *testing.T, repo *gitRepo, localDir, remoteDir string) {
				// Verify we're back on edit branch
				gitRepo, err := git.PlainOpen(localDir)
				require.NoError(t, err)
				head, err := gitRepo.Head()
				require.NoError(t, err)
				require.Equal(t, "edit-branch", head.Name().Short())

				// Verify changes were pushed to remote default branch
				workingDir := t.TempDir()
				cmd := exec.Command("git", "clone", "-b", "main", remoteDir, workingDir)
				err = cmd.Run()
				require.NoError(t, err)

				newFeaturePath := filepath.Join(workingDir, "new_feature.txt")
				_, err = os.Stat(newFeaturePath)
				require.NoError(t, err, "new_feature.txt should exist in remote main branch")
			},
		},
		{
			name: "commit and push with force when there are conflicts",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				// Clone repository and create edit branch
				repo, err := git.PlainClone(localDir, false, &git.CloneOptions{
					URL:           remoteURL,
					RemoteName:    "origin",
					ReferenceName: plumbing.ReferenceName("refs/heads/main"),
					SingleBranch:  false,
				})
				require.NoError(t, err)

				// Create and switch to edit branch
				worktree, err := repo.Worktree()
				require.NoError(t, err)
				err = worktree.Checkout(&git.CheckoutOptions{
					Branch: plumbing.ReferenceName("refs/heads/edit-branch"),
					Create: true,
				})
				require.NoError(t, err)

				return &gitRepo{
					h:             &Handle{logger: zap.NewNop()},
					repoDir:       localDir,
					remoteURL:     remoteURL,
					defaultBranch: "main",
					editBranch:    "edit-branch",
					subpath:       "",
					managedRepo:   true,
				}
			},
			setupChanges: func(t *testing.T, localDir string) {
				// Modify existing file in edit branch
				filePath := filepath.Join(localDir, "test1.txt")
				err := os.WriteFile(filePath, []byte("edit branch content"), 0644)
				require.NoError(t, err)
			},
			setupRemote: func(t *testing.T, remoteDir string) {
				// Add conflicting changes to remote main branch
				createRemoteCommit(t, remoteDir, "test1.txt", "conflicting remote content", "Conflicting remote change")
			},
			message:     "Edit branch changes",
			force:       true,
			expectError: false,
			validate: func(t *testing.T, repo *gitRepo, localDir, remoteDir string) {
				// Verify we're back on edit branch
				gitRepo, err := git.PlainOpen(localDir)
				require.NoError(t, err)
				head, err := gitRepo.Head()
				require.NoError(t, err)
				require.Equal(t, "edit-branch", head.Name().Short())

				// Verify changes were resolved and pushed (force merge should use "theirs" strategy)
				workingDir := t.TempDir()
				cmd := exec.Command("git", "clone", "-b", "main", remoteDir, workingDir)
				err = cmd.Run()
				require.NoError(t, err)

				test1Path := filepath.Join(workingDir, "test1.txt")
				content, err := os.ReadFile(test1Path)
				require.NoError(t, err)
				// With force merge using "theirs" strategy, edit branch content should win
				require.Equal(t, "edit branch content", string(content))
			},
		},
		{
			name: "abort merge when conflicts exist and force is false",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				// Clone repository and create edit branch
				repo, err := git.PlainClone(localDir, false, &git.CloneOptions{
					URL:           remoteURL,
					RemoteName:    "origin",
					ReferenceName: plumbing.ReferenceName("refs/heads/main"),
					SingleBranch:  false,
				})
				require.NoError(t, err)

				// Create and switch to edit branch
				worktree, err := repo.Worktree()
				require.NoError(t, err)
				err = worktree.Checkout(&git.CheckoutOptions{
					Branch: plumbing.ReferenceName("refs/heads/edit-branch"),
					Create: true,
				})
				require.NoError(t, err)

				return &gitRepo{
					h:             &Handle{logger: zap.NewNop()},
					repoDir:       localDir,
					remoteURL:     remoteURL,
					defaultBranch: "main",
					editBranch:    "edit-branch",
					subpath:       "",
					managedRepo:   true,
				}
			},
			setupChanges: func(t *testing.T, localDir string) {
				// Modify existing file in edit branch
				filePath := filepath.Join(localDir, "test1.txt")
				err := os.WriteFile(filePath, []byte("edit branch content"), 0644)
				require.NoError(t, err)
			},
			setupRemote: func(t *testing.T, remoteDir string) {
				// Add conflicting changes to remote main branch
				createRemoteCommit(t, remoteDir, "test1.txt", "conflicting remote content", "Conflicting remote change")
			},
			message:     "Edit branch changes",
			force:       false,
			expectError: false, // Should not error, but should abort merge
			validate: func(t *testing.T, repo *gitRepo, localDir, remoteDir string) {
				// Verify we're back on edit branch
				gitRepo, err := git.PlainOpen(localDir)
				require.NoError(t, err)
				head, err := gitRepo.Head()
				require.NoError(t, err)
				require.Equal(t, "edit-branch", head.Name().Short())

				// Verify remote main branch was NOT updated (merge was aborted)
				workingDir := t.TempDir()
				cmd := exec.Command("git", "clone", "-b", "main", remoteDir, workingDir)
				err = cmd.Run()
				require.NoError(t, err)

				test1Path := filepath.Join(workingDir, "test1.txt")
				content, err := os.ReadFile(test1Path)
				require.NoError(t, err)
				// Should still have the conflicting remote content, not the edit branch content
				require.Equal(t, "conflicting remote content", string(content))
			},
		},
		{
			name: "handle empty commit gracefully",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				// Clone repository and create edit branch
				repo, err := git.PlainClone(localDir, false, &git.CloneOptions{
					URL:           remoteURL,
					RemoteName:    "origin",
					ReferenceName: plumbing.ReferenceName("refs/heads/main"),
					SingleBranch:  false,
				})
				require.NoError(t, err)

				// Create and switch to edit branch
				worktree, err := repo.Worktree()
				require.NoError(t, err)
				err = worktree.Checkout(&git.CheckoutOptions{
					Branch: plumbing.ReferenceName("refs/heads/edit-branch"),
					Create: true,
				})
				require.NoError(t, err)

				return &gitRepo{
					h:             &Handle{logger: zap.NewNop()},
					repoDir:       localDir,
					remoteURL:     remoteURL,
					defaultBranch: "main",
					editBranch:    "edit-branch",
					subpath:       "",
					managedRepo:   true,
				}
			},
			setupChanges: func(t *testing.T, localDir string) {
				// No changes made - this should result in empty commit
			},
			setupRemote: func(t *testing.T, remoteDir string) {
				// No additional remote changes
			},
			message:     "Empty commit",
			force:       false,
			expectError: false, // Should handle empty commit gracefully
			validate: func(t *testing.T, repo *gitRepo, localDir, remoteDir string) {
				// Verify we're back on edit branch
				gitRepo, err := git.PlainOpen(localDir)
				require.NoError(t, err)
				head, err := gitRepo.Head()
				require.NoError(t, err)
				require.Equal(t, "edit-branch", head.Name().Short())

				// Verify no new commits were added to remote
				workingDir := t.TempDir()
				cmd := exec.Command("git", "clone", "-b", "main", remoteDir, workingDir)
				err = cmd.Run()
				require.NoError(t, err)

				// Count commits - should still be the initial commit only
				cmd = exec.Command("git", "-C", workingDir, "rev-list", "--count", "HEAD")
				output, err := cmd.Output()
				require.NoError(t, err)
				require.Equal(t, "1\n", string(output)) // Only initial commit
			},
		},
		{
			name: "error when repository is not editable",
			setupRepo: func(t *testing.T, localDir, remoteURL string) *gitRepo {
				// Clone repository in non-editable mode
				_, err := git.PlainClone(localDir, false, &git.CloneOptions{
					URL:           remoteURL,
					RemoteName:    "origin",
					ReferenceName: plumbing.ReferenceName("refs/heads/main"),
					SingleBranch:  true,
				})
				require.NoError(t, err)

				return &gitRepo{
					h:             &Handle{logger: zap.NewNop()},
					repoDir:       localDir,
					remoteURL:     remoteURL,
					defaultBranch: "main",
					editBranch:    "", // Non-editable
					subpath:       "",
					managedRepo:   true,
				}
			},
			setupChanges: func(t *testing.T, localDir string) {
				// Make some changes
				filePath := filepath.Join(localDir, "should_not_commit.txt")
				err := os.WriteFile(filePath, []byte("content"), 0644)
				require.NoError(t, err)
			},
			setupRemote: func(t *testing.T, remoteDir string) {
				// No additional remote changes
			},
			message:     "Should fail",
			force:       false,
			expectError: true,
			validate: func(t *testing.T, repo *gitRepo, localDir, remoteDir string) {
				// Verify we're still on main branch
				gitRepo, err := git.PlainOpen(localDir)
				require.NoError(t, err)
				head, err := gitRepo.Head()
				require.NoError(t, err)
				require.Equal(t, "main", head.Name().Short())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			localDir := t.TempDir()
			remoteDir := setupTestGitRepository(t)

			// Setup the gitRepo instance
			repo := tt.setupRepo(t, localDir, remoteDir)

			// Setup any changes
			tt.setupChanges(t, localDir)

			// Setup any remote changes
			tt.setupRemote(t, remoteDir)

			// Execute commitAndPushToDefaultBranch
			ctx := context.Background()
			err := repo.commitAndPushToDefaultBranch(ctx, tt.message, tt.force)

			// Verify error expectation
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Run validation
			if tt.validate != nil {
				tt.validate(t, repo, localDir, remoteDir)
			}
		})
	}
}
