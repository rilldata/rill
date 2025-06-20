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

	remoteURL string // Note that repo.checkSyncHandshake may update it at any time
	branch    string // Note that repo.checkSyncHandshake may update it at any time
	subpath   string // Note that repo.checkSyncHandshake may update it at any time
}

// sync clones or pulls from the remote Git repository.
func (r *gitRepo) sync(ctx context.Context) error {
	// Call syncInner with retries
	var err error
	for i := 0; i < gitRetryN; i++ {
		err = r.syncInner(ctx)
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

// syncInner contains the actual logic of r.sync without retries.
func (r *gitRepo) syncInner(ctx context.Context) error {
	// Check if repoDir exists and is a valid git repository
	repo, err := git.PlainOpen(r.repoDir)
	if err != nil {
		// Repository doesn't exist or is invalid, remove and clone fresh
		if err := os.RemoveAll(r.repoDir); err != nil {
			return err
		}

		cloneOptions := &git.CloneOptions{
			URL:           r.remoteURL,
			RemoteName:    "origin",
			ReferenceName: plumbing.ReferenceName("refs/heads/" + r.branch),
			SingleBranch:  true,
		}

		_, err = git.PlainCloneContext(ctx, r.repoDir, false, cloneOptions)
		return err
	}

	// Ensure the remote URL is correct
	_ = repo.DeleteRemote("origin")
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{r.remoteURL},
	})
	if err != nil {
		return fmt.Errorf("failed to set remote URL: %w", err)
	}

	// Repository exists, pull latest changes
	workTree, err := repo.Worktree()
	if err != nil {
		return err
	}
	err = workTree.PullContext(ctx, &git.PullOptions{
		RemoteURL:     r.remoteURL,
		ReferenceName: plumbing.ReferenceName("refs/heads/" + r.branch),
		SingleBranch:  true,
		Force:         true,
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		rev, err := repo.ResolveRevision(plumbing.Revision("remotes/origin/HEAD"))
		if err != nil {
			return err
		}

		return workTree.Reset(&git.ResetOptions{
			Commit: *rev,
			Mode:   git.HardReset,
		})
	}

	// Checkout in case the branch was changed
	ref, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}
	if ref.Name().Short() != r.branch {
		err = workTree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.ReferenceName("refs/heads/" + r.branch),
			Force:  true,
		})
		if err != nil {
			return fmt.Errorf("failed to checkout branch %s: %w", r.branch, err)
		}
	}

	return nil
}

func (r *gitRepo) root() string {
	if r.subpath != "" {
		return path.Join(r.repoDir, r.subpath)
	}
	return r.repoDir
}

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
