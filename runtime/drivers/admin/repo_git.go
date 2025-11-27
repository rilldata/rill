package admin

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rilldata/rill/runtime/pkg/gitutil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

const (
	gitRetryN    = 3
	gitRetryWait = 2 * time.Second
)

// gitRepo represents a remote Git repository.
// It is unsafe for concurrent reads and writes.
type gitRepo struct {
	h       *Handle
	repoDir string // The persistent directory where we store the Git repository

	remoteURL     string // Note that repo.checkSyncHandshake may update it at any time
	defaultBranch string // Note that repo.checkSyncHandshake may update it at any time
	editBranch    string // Does not change. Only set for dev deployments.
	subpath       string // Note that repo.checkSyncHandshake may update it at any time
	managedRepo   bool   // Whether the repo is managed by Rill
}

// pull clones or pulls from the remote Git repository.
func (r *gitRepo) pull(ctx context.Context, force bool) error {
	// Call pullInner with retries
	var err error
	for i := 0; i < gitRetryN; i++ {
		err = r.pullInner(ctx, force)
		if err == nil {
			break
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			break
		}
		select {
		case <-time.After(gitRetryWait):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return err
}

// pullInner contains the actual logic of r.pull without retries.
func (r *gitRepo) pullInner(ctx context.Context, force bool) error {
	// If the repository is not editable, there shouldn't be any local changes, but just to be safe, we always force pull.
	if !r.editable() {
		force = true
	}

	// Check if repoDir exists and is a valid Git repository
	repo, err := git.PlainOpen(r.repoDir)
	if err != nil {
		// Repository doesn't exist or is invalid, remove and clone fresh
		if err := os.RemoveAll(r.repoDir); err != nil {
			return err
		}

		cloneOptions := &git.CloneOptions{
			URL:           r.remoteURL,
			RemoteName:    "origin",
			ReferenceName: plumbing.ReferenceName("refs/heads/" + r.defaultBranch),
			SingleBranch:  !r.editable(),
		}

		_, err = git.PlainCloneContext(ctx, r.repoDir, false, cloneOptions)
		return err
	}

	// Repository exists, pull latest changes

	// Ensure the remote URL is correct
	_ = repo.DeleteRemote("origin")
	remote, err := repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{r.remoteURL},
	})
	if err != nil {
		return fmt.Errorf("failed to set remote URL: %w", err)
	}

	// check what is the current branch
	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	onEditBranch := head.Name().Short() == r.editBranch

	refSpecs := []config.RefSpec{
		config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/remotes/origin/%s", r.defaultBranch, r.defaultBranch)),
	}
	if onEditBranch {
		refSpecs = append(refSpecs, config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/remotes/origin/%s", r.editBranch, r.editBranch)))
	}

	// Fetch the remote changes
	err = remote.Fetch(&git.FetchOptions{
		RefSpecs: refSpecs,
		Force:    true,
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("failed to fetch from remote: %w", err)
	}

	// If we are not on the edit branch, we checkout the default branch.
	if !onEditBranch {
		// Checkout the default branch
		worktree, err := repo.Worktree()
		if err != nil {
			return fmt.Errorf("failed to get worktree: %w", err)
		}
		err = worktree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.ReferenceName("refs/heads/" + r.defaultBranch),
			Force:  true,
		})
		if err != nil {
			return fmt.Errorf("failed to checkout branch %q: %w", r.defaultBranch, err)
		}

		// Hard reset to remote branch
		err = resetToRemoteTrackingBranch(repo, worktree, r.defaultBranch)
		if err != nil {
			return fmt.Errorf("failed to reset to remote tracking branch %q: %w", r.defaultBranch, err)
		}
	}

	// If not editable, we stay on the default branch.
	if !r.editable() {
		return nil
	}

	// We're in editable mode, so r.editBranch is set. We want to pull/create it and switch to it.
	// The edit branch enables us to commit progress when closing (e.g. due to hibernation) without affecting the default branch.
	// When pulling the editBranch, we should force pull if there are conflicts even if `force` is false (to bring us in sync with changes made in a split-brain scenario).
	// To reduce the chance of conflicts, we should also try to merge the default branch into the edit branch (but only force merge if `force` is true).

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}
	if !onEditBranch {
		// Create the edit branch if it doesn't exist
		err = worktree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", r.editBranch)),
			Create: true,
			Force:  true,
		})
		if err != nil {
			return fmt.Errorf("failed to create edit branch %q: %w", r.editBranch, err)
		}
	} else {
		// Hard reset to remote branch (this discards all local changes)
		err = resetToRemoteTrackingBranch(repo, worktree, r.editBranch)
		if err != nil {
			return fmt.Errorf("failed to reset to remote tracking branch %q: %w", r.editBranch, err)
		}
	}

	// merge default branch into edit branch
	if force {
		// Maybe instead of merge with the "theirs" strategy should we just reset to the default branch?
		err = gitutil.MergeWithTheirsStrategy(r.repoDir, r.defaultBranch)
	} else {
		_, err = gitutil.MergeWithBailOnConflict(r.repoDir, r.defaultBranch)
	}
	if err != nil {
		return fmt.Errorf("failed to merge default branch %q into edit branch %q: %w", r.defaultBranch, r.editBranch, err)
	}
	return nil
}

// editable returns true if its allowed to edit files in the repository.
// Will be true for dev deployments with an edit branch, and false for prod deployments that serve files on the default branch.
func (r *gitRepo) editable() bool {
	return r.editBranch != ""
}

// root returns the absolute path to the root of the Rill project.
func (r *gitRepo) root() string {
	if r.subpath != "" {
		return path.Join(r.repoDir, r.subpath)
	}
	return r.repoDir
}

// commitToEditBranch auto-commits any current changes to the edit branch of the repository.
// This is done to checkpoint progress when the handle is closed.
// If there are conflicts, it should drop any local changes.
func (r *gitRepo) commitToEditBranch(ctx context.Context) error {
	if !r.editable() {
		return fmt.Errorf("cannot commit to the edit branch because it is not configured")
	}

	r.h.logger.Info("commitToEditBranch", observability.ZapCtx(ctx))
	repo, err := git.PlainOpen(r.repoDir)
	if err != nil {
		return err
	}

	err = r.commitAll(repo, "Checkpoint commit")
	if err != nil {
		if errors.Is(err, git.ErrEmptyCommit) {
			return nil // No changes to commit
		}
		return fmt.Errorf("failed to commit changes to edit branch: %w", err)
	}

	// Push the changes to the remote edit branch
	err = repo.PushContext(ctx, &git.PushOptions{
		RemoteName: "origin",
		RemoteURL:  r.remoteURL,
	})
	if err != nil {
		return fmt.Errorf("failed to push changes to remote edit branch: %w", err)
	}
	return nil
}

// commitAndPush commits changes to the repository and pushes them to the remote.
func (r *gitRepo) commitAndPushToDefaultBranch(ctx context.Context, message string, force bool) error {
	if !r.editable() {
		return fmt.Errorf("cannot commit to this repository because it is not marked editable")
	}

	r.h.logger.Info("commitAndPushToDefaultBranch", zap.String("message", message), zap.Bool("force", force), observability.ZapCtx(ctx))
	repo, err := git.PlainOpen(r.repoDir)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}
	err = r.commitAll(repo, message)
	if err != nil {
		if errors.Is(err, git.ErrEmptyCommit) {
			return nil // No changes to commit
		}
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	// Fetch the default branch to ensure we are up-to-date
	err = repo.FetchContext(ctx, &git.FetchOptions{
		RemoteURL: r.remoteURL,
		RefSpecs:  []config.RefSpec{config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/remotes/origin/%s", r.defaultBranch, r.defaultBranch))},
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("failed to fetch default branch %q: %w", r.defaultBranch, err)
	}

	// Switch to the default branch
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}
	defer func() {
		// switch back to the edit branch
		err := worktree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.ReferenceName("refs/heads/" + r.editBranch),
			Force:  true,
		})
		if err != nil {
			r.h.logger.Error("failed to switch back to edit branch after commit", zap.String("editBranch", r.editBranch), zap.Error(err))
		}
	}()

	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/heads/" + r.defaultBranch),
		Force:  true,
	})
	if err != nil {
		return fmt.Errorf("failed to checkout default branch %q: %w", r.defaultBranch, err)
	}
	err = resetToRemoteTrackingBranch(repo, worktree, r.defaultBranch)
	if err != nil {
		return fmt.Errorf("failed to reset to remote tracking branch %q: %w", r.defaultBranch, err)
	}

	// Merge the edit branch into the default branch
	merged := true
	if force {
		err = gitutil.MergeWithTheirsStrategy(r.repoDir, r.editBranch)
	} else {
		merged, err = gitutil.MergeWithBailOnConflict(r.repoDir, r.editBranch)
	}
	if err != nil {
		return fmt.Errorf("failed to merge edit branch %q into default branch %q: %w", r.editBranch, r.defaultBranch, err)
	}

	if !merged {
		// If the merge was aborted no need to push the changes
		r.h.logger.Warn("Merge aborted due to conflicts, not pushing changes", zap.String("editBranch", r.editBranch), zap.String("defaultBranch", r.defaultBranch))
		return nil
	}

	// Push the changes to the remote default branch
	err = repo.PushContext(ctx, &git.PushOptions{
		RemoteName: "origin",
		RemoteURL:  r.remoteURL,
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", r.defaultBranch, r.defaultBranch)),
			config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", r.editBranch, r.editBranch)),
		},
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("failed to push changes to remote default branch %q: %w", r.defaultBranch, err)
	}

	return nil
}

// commitHash returns the current commit hash of the repository.
func (r *gitRepo) commitHash() (string, error) {
	repo, err := git.PlainOpen(r.repoDir)
	if err != nil {
		return "", err
	}

	ref, err := repo.Head()
	if err != nil {
		return "", err
	}

	if ref.Hash().IsZero() {
		return "", nil
	}

	return ref.Hash().String(), nil
}

// commitTimestamp returns the timestamp of the latest commit on the current branch.
func (r *gitRepo) commitTimestamp() (time.Time, error) {
	repo, err := git.PlainOpen(r.repoDir)
	if err != nil {
		return time.Time{}, err
	}

	ref, err := repo.Head()
	if err != nil {
		return time.Time{}, err
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return time.Time{}, err
	}

	return commit.Author.When, nil
}

func (r *gitRepo) commitAll(repo *git.Repository, message string) error {
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = worktree.AddWithOptions(&git.AddOptions{
		All: true, // Add all changes
	})
	if err != nil {
		return err
	}

	_, err = worktree.Commit(message, &git.CommitOptions{
		All: true, // Commit all changes
		Author: &object.Signature{
			Name:  "Rill Runtime",
			Email: "runtime@rilldata.com", // Use a generic author for the commit
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *gitRepo) fetchCurrentBranch(ctx context.Context) error {
	repo, err := git.PlainOpen(r.repoDir)
	if err != nil {
		return err
	}
	head, err := repo.Head()
	if err != nil {
		return err
	}
	err = repo.FetchContext(ctx, &git.FetchOptions{
		RefSpecs: []config.RefSpec{config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/remotes/origin/%s", head.Name().Short(), head.Name().Short()))},
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return err
	}
	return nil
}

// resetToRemoteTrackingBranch resets to the commit pointed by the remote tracking branch.
// This is used to reset the local branch to the state of the remote branch so it is expected that the latest changes have been fetched.
func resetToRemoteTrackingBranch(repo *git.Repository, wt *git.Worktree, branch string) error {
	trackingRef, err := repo.Reference(plumbing.ReferenceName(fmt.Sprintf("refs/remotes/origin/%s", branch)), true)
	if err != nil {
		return err
	}

	err = wt.Reset(&git.ResetOptions{
		Commit: trackingRef.Hash(),
		Mode:   git.HardReset,
	})
	if err != nil {
		return err
	}
	return nil
}
