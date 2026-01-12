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

	// Fetch the default branch from remote
	err = remote.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", r.defaultBranch, r.defaultBranch))},
		Force:    true,
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("failed to fetch from remote: %w", err)
	}

	// Checkout the default branch (in case it was changed)
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}
	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/heads/" + r.defaultBranch),
		Force:  true,
	})
	if err != nil {
		return fmt.Errorf("failed to checkout branch %q: %w", r.defaultBranch, err)
	}

	// Pull in the latest changes
	err = worktree.PullContext(ctx, &git.PullOptions{
		RemoteURL:     r.remoteURL,
		ReferenceName: plumbing.ReferenceName("refs/heads/" + r.defaultBranch),
		SingleBranch:  !r.editable(),
		Force:         true,
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		rev, err := repo.ResolveRevision(plumbing.Revision(fmt.Sprintf("refs/remotes/origin/%s", r.defaultBranch)))
		if err != nil {
			return err
		}

		err = worktree.Reset(&git.ResetOptions{
			Commit: *rev,
			Mode:   git.HardReset,
		})
		if err != nil {
			return err
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

	// TODO: Implement editable mode.
	r.h.logger.Info("pullInner", zap.Bool("force", force), observability.ZapCtx(ctx))

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

	// TODO: Implement
	r.h.logger.Info("commitToEditBranch", observability.ZapCtx(ctx))

	return nil
}

// commitAndPush commits changes to the repository and pushes them to the remote.
func (r *gitRepo) commitAndPushToDefaultBranch(ctx context.Context, message string, force bool) error {
	if !r.editable() {
		return fmt.Errorf("cannot commit to this repository because it is not marked editable")
	}

	// TODO: Commit to r.editBranch, then merge it into r.defaultBranch and push it to the remote (respecting force).
	r.h.logger.Info("commitAndPushToDefaultBranch", zap.String("message", message), zap.Bool("force", force), observability.ZapCtx(ctx))

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
