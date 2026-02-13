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
	primaryBranch string // Primary branch of the project.
	editableDepl  bool   // Whether this is a dev deployment where editing is allowed
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
			SingleBranch:  r.primaryBranch == r.defaultBranch,
			ReferenceName: plumbing.ReferenceName("refs/heads/" + r.primaryBranch), // primary branch must exist, default branch may not exist yet in editable mode.
		}

		repo, err = git.PlainCloneContext(ctx, r.repoDir, false, cloneOptions)
		if err != nil {
			return err
		}
	} else {
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

		// Fetch the remote changes.
		// We fetch each branch individually because a NoMatchingRefSpecError for one branch (e.g., an edit branch
		// that doesn't exist on the remote yet) would cause a combined fetch to skip all branches.
		branches := []string{r.defaultBranch}
		if r.primaryBranch != r.defaultBranch {
			branches = append(branches, r.primaryBranch)
		}
		for _, branch := range branches {
			refSpec := config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/remotes/origin/%s", branch, branch))
			err = remote.Fetch(&git.FetchOptions{
				RefSpecs: []config.RefSpec{refSpec},
				Force:    true,
			})
			if err != nil && !(errors.Is(err, git.NoErrAlreadyUpToDate) || git.NoMatchingRefSpecError{}.Is(err)) {
				return fmt.Errorf("failed to fetch from remote: %w", err)
			}
		}
	}

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
		if !errors.Is(err, plumbing.ErrReferenceNotFound) {
			return fmt.Errorf("failed to checkout branch %q: %w", r.defaultBranch, err)
		}

		// The default branch does not exist locally, try to find it on remote. It can happen in cases:
		// a. when branch is created remotely after the last pull
		// b. primary branch was edited
		// c. in editable mode when the default branch may not exist at all.
		remoteHash, err := repo.Reference(plumbing.ReferenceName("refs/remotes/origin/"+r.defaultBranch), true)
		if err != nil {
			if errors.Is(err, plumbing.ErrReferenceNotFound) && r.editable() {
				// In editable mode, the default branch may not exist yet on remote. We will create it based on the primary branch.
				r.h.logger.Info("Default branch does not exist on remote, will create it based on primary branch", zap.String("defaultBranch", r.defaultBranch), zap.String("primaryBranch", r.primaryBranch))
				remoteHash, err = repo.Reference(plumbing.ReferenceName("refs/remotes/origin/"+r.primaryBranch), true)
				if err != nil {
					return fmt.Errorf("failed to get reference for primary branch %q: %w", r.primaryBranch, err)
				}
			} else {
				// In non-editable mode, the default branch must exist on remote, if not found, return error.
				return fmt.Errorf("failed to get remote tracking branch %q: %w", r.defaultBranch, err)
			}
		}

		// create the default branch
		err = worktree.Checkout(&git.CheckoutOptions{
			Hash:   remoteHash.Hash(),
			Branch: plumbing.ReferenceName("refs/heads/" + r.defaultBranch),
			Create: true,
			Force:  true,
		})
		if err != nil {
			return fmt.Errorf("failed to create and checkout default branch %q: %w", r.defaultBranch, err)
		}
	}

	// Hard reset to remote branch
	err = resetToRemoteTrackingBranch(repo, worktree, r.defaultBranch)
	if err != nil {
		if !(errors.Is(err, plumbing.ErrReferenceNotFound) && r.editable()) { // In editable mode, the default branch may not exist yet on remote.
			return fmt.Errorf("failed to reset to remote tracking branch %q: %w", r.defaultBranch, err)
		}
	}

	if !r.editable() {
		return nil
	}

	// We're in editable mode.
	// To reduce the chance of conflicts, we should also try to merge the primary branch into the default branch (but only force merge if `force` is true).

	// merge primary branch into edit branch
	mergeBranch := "origin/" + r.primaryBranch
	if force {
		err = gitutil.MergeWithTheirsStrategy(r.repoDir, mergeBranch)
	} else {
		_, err = gitutil.MergeWithBailOnConflict(r.repoDir, mergeBranch)
	}
	if err != nil {
		return fmt.Errorf("failed to merge primary branch %q into default branch %q: %w", r.primaryBranch, r.defaultBranch, err)
	}
	return nil
}

// editable returns true if its allowed to edit files in the repository.
// Will be true for dev deployments with an edit branch, and false for prod deployments that serve files on the default branch.
func (r *gitRepo) editable() bool {
	return r.editableDepl
}

// root returns the absolute path to the root of the Rill project.
func (r *gitRepo) root() string {
	if r.subpath != "" {
		return path.Join(r.repoDir, r.subpath)
	}
	return r.repoDir
}

// commitToDefaultBranch auto-commits any current changes to the default branch of the repository. This is only allowed if editable is true.
// This is done to checkpoint progress when the handle is closed.
// If there are conflicts, it should drop any local changes.
func (r *gitRepo) commitToDefaultBranch(ctx context.Context, message string) (string, error) {
	if !r.editable() {
		return "", fmt.Errorf("cannot commit to the default branch because it is not configured")
	}

	r.h.logger.Info("commitToDefaultBranch", observability.ZapCtx(ctx))
	repo, err := git.PlainOpen(r.repoDir)
	if err != nil {
		return "", err
	}

	hash, err := r.commitAll(repo, message)
	if err != nil {
		if errors.Is(err, git.ErrEmptyCommit) {
			return "", nil // No changes to commit
		}
		return "", fmt.Errorf("failed to commit changes to edit branch: %w", err)
	}

	// Push the changes to the remote edit branch
	err = repo.PushContext(ctx, &git.PushOptions{
		RemoteName: "origin",
		RemoteURL:  r.remoteURL,
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", r.defaultBranch, r.defaultBranch)),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to push changes to remote default branch: %w", err)
	}
	return hash, nil
}

// commitAndPush commits changes to the repository and pushes them to the remote.
func (r *gitRepo) commitAndPushToPrimaryBranch(ctx context.Context, message string, force bool) (resErr error) {
	if !r.editable() {
		return fmt.Errorf("cannot commit to this repository because it is not marked editable")
	}

	r.h.logger.Info("commitAndPushToPrimaryBranch", zap.String("message", message), zap.Bool("force", force), observability.ZapCtx(ctx))
	repo, err := git.PlainOpen(r.repoDir)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}
	_, err = r.commitAll(repo, message)
	if err != nil {
		if errors.Is(err, git.ErrEmptyCommit) {
			return nil // No changes to commit
		}
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	// Fetch the primary branch to ensure we are up-to-date
	err = repo.FetchContext(ctx, &git.FetchOptions{
		RemoteURL: r.remoteURL,
		RefSpecs:  []config.RefSpec{config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/remotes/origin/%s", r.primaryBranch, r.primaryBranch))},
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("failed to fetch primary branch %q: %w", r.primaryBranch, err)
	}

	// Switch to the primary branch
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}
	defer func() {
		// switch back to the default branch
		err := worktree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.ReferenceName("refs/heads/" + r.defaultBranch),
			Force:  true,
		})
		if err != nil {
			resErr = errors.Join(resErr, fmt.Errorf("failed to checkout default branch %q: %w", r.defaultBranch, err))
			return
		}
	}()

	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/heads/" + r.primaryBranch),
		Force:  true,
	})
	if err != nil {
		return fmt.Errorf("failed to checkout primary branch %q: %w", r.primaryBranch, err)
	}
	err = resetToRemoteTrackingBranch(repo, worktree, r.primaryBranch)
	if err != nil {
		return fmt.Errorf("failed to reset to remote tracking branch %q: %w", r.primaryBranch, err)
	}

	// Merge the default branch into the primary branch
	merged := true
	if force {
		err = gitutil.MergeWithTheirsStrategy(r.repoDir, r.defaultBranch)
	} else {
		merged, err = gitutil.MergeWithBailOnConflict(r.repoDir, r.defaultBranch)
	}
	if err != nil {
		return fmt.Errorf("failed to merge default branch %q into primary branch %q: %w", r.defaultBranch, r.primaryBranch, err)
	}

	if !merged {
		// If the merge was aborted no need to push the changes
		r.h.logger.Warn("Merge aborted due to conflicts, not pushing changes", zap.String("primaryBranch", r.primaryBranch), zap.String("defaultBranch", r.defaultBranch))
		return nil
	}

	// Push the changes to the remote default branch
	err = repo.PushContext(ctx, &git.PushOptions{
		RemoteName: "origin",
		RemoteURL:  r.remoteURL,
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", r.defaultBranch, r.defaultBranch)),
			config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", r.primaryBranch, r.primaryBranch)),
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

func (r *gitRepo) commitAll(repo *git.Repository, message string) (string, error) {
	worktree, err := repo.Worktree()
	if err != nil {
		return "", err
	}

	err = worktree.AddWithOptions(&git.AddOptions{
		All: true, // Add all changes
	})
	if err != nil {
		return "", err
	}

	hash, err := worktree.Commit(message, &git.CommitOptions{
		All: true, // Commit all changes
		Author: &object.Signature{
			Name:  "Rill Runtime",
			Email: "runtime@rilldata.com", // Use a generic author for the commit
		},
	})
	if err != nil {
		return "", err
	}
	return hash.String(), nil
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
