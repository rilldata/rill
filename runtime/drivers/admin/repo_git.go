package admin

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
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
func (r *gitRepo) pull(ctx context.Context, userTriggered, force bool) error {
	// Call pullInner with retries
	var err error
	for i := 0; i < gitRetryN; i++ {
		err = r.pullInner(ctx, userTriggered, force)
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
func (r *gitRepo) pullInner(ctx context.Context, userTriggered, force bool) error {
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
		if r.editableDepl {
			// set git config in the repo dir to ensure git commits/git merge etc pass on cloud
			err = ensureGitConfig(r.repoDir, "user.name", "Rill")
			if err != nil {
				return err
			}
			err = ensureGitConfig(r.repoDir, "user.email", "noreply@rilldata.com")
			if err != nil {
				return err
			}
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
	var createDefault bool
	err = gitCheckout(r.repoDir, r.defaultBranch, force, false, "")
	if err != nil {
		if !errors.Is(err, errRefNotFound) {
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
				createDefault = true
			} else {
				// In non-editable mode, the default branch must exist on remote, if not found, return error.
				return fmt.Errorf("failed to get remote tracking branch %q: %w", r.defaultBranch, err)
			}
		}

		// Create the default branch at the resolved remote hash
		err = gitCheckout(r.repoDir, r.defaultBranch, true, true, remoteHash.Hash().String())
		if err != nil {
			return fmt.Errorf("failed to create and checkout default branch %q: %w", r.defaultBranch, err)
		}

		if createDefault {
			// Also push the newly created branch to remote so other operations like git status pass without error.
			err = r.pushBranch(ctx, r.defaultBranch)
			if err != nil {
				return fmt.Errorf("failed to push changes to remote default branch: %w", err)
			}
		}
	}

	if force {
		// Hard reset to remote branch
		err = resetToRemoteTrackingBranch(r.repoDir, r.defaultBranch)
		if err != nil {
			if !(errors.Is(err, errRefNotFound) && r.editable()) { // In editable mode, the default branch may not exist yet on remote.
				return fmt.Errorf("failed to reset to remote tracking branch %q: %w", r.defaultBranch, err)
			}
		}
	} else if !createDefault { // If we just created the default branch, there's no need to merge
		merged, err := gitutil.MergeWithBailOnConflict(r.repoDir, fmt.Sprintf("%s/%s", "origin", r.defaultBranch))
		if err != nil {
			return err
		}
		if !merged { // Only user triggered pulls should fail on conflicts
			if userTriggered {
				return fmt.Errorf("local is behind remote and failed to sync with remote due to conflicts")
			}
			r.h.logger.Warn("Merge aborted due to conflicts, local changes not synced with remote", zap.String("branch", r.defaultBranch))
		}
	}

	if !r.editable() || r.primaryBranch == r.defaultBranch {
		return nil
	}

	// We're in editable mode.
	// To reduce the chance of conflicts, we should also try to merge the primary branch into the default branch (but only force merge if `force` is true).

	// merge primary branch into edit branch
	mergeBranch := "origin/" + r.primaryBranch
	if force {
		err = gitutil.MergeWithStrategy(r.repoDir, mergeBranch, "theirs")
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
func (r *gitRepo) commitToDefaultBranch(ctx context.Context, message string, force bool) error {
	if !r.editable() {
		return fmt.Errorf("cannot commit to the default branch because it is not configured")
	}

	r.h.logger.Info("commitToDefaultBranch", observability.ZapCtx(ctx))
	repo, err := git.PlainOpen(r.repoDir)
	if err != nil {
		return err
	}

	_, err = r.commitAll(repo, message)
	if err != nil {
		if errors.Is(err, git.ErrEmptyCommit) {
			return nil // No changes to commit
		}
		return fmt.Errorf("failed to commit changes to edit branch: %w", err)
	}

	err = r.fetchCurrentBranch(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch default branch: %w", err)
	}

	if !force {
		err = gitutil.MergeWithStrategy(r.repoDir, fmt.Sprintf("%s/%s", "origin", r.defaultBranch), "")
		if err != nil {
			return fmt.Errorf("local is behind remote and failed to sync with remote: %w", err)
		}
		err = r.pushBranch(ctx, r.defaultBranch)
		if err != nil {
			return err
		}
		return nil
	}

	// Instead of a force push, we do a merge with 'ours' strategy to ensure we don't lose history.
	// This is not equivalent to a force push but is safer for users.
	if r.subpath != "" {
		// force pushing in a monorepo can overwrite other subpaths so just try with normal push
		// we can check for changes in other subpaths but it is tricky and error prone
		// monorepo setups are advanced use cases and we can require users to manually resolve remote changes
		err = r.pushBranch(ctx, r.defaultBranch)
		if err != nil {
			return err
		}
		return nil
	}
	err = gitutil.MergeWithStrategy(r.repoDir, fmt.Sprintf("%s/%s", "origin", r.defaultBranch), "ours")
	if err != nil {
		return fmt.Errorf("local is behind remote and failed to sync with remote: %w", err)
	}
	err = r.pushBranch(ctx, r.defaultBranch)
	if err != nil {
		return err
	}
	return nil
}

// mergeToBranch commits changes to the repository and merges them into the specified branch.
func (r *gitRepo) mergeToBranch(ctx context.Context, branch string, force bool) (resErr error) {
	if !r.editable() {
		return fmt.Errorf("cannot commit to this repository because it is not marked editable")
	}

	r.h.logger.Info("mergeToBranch", zap.String("branch", branch), zap.Bool("force", force), observability.ZapCtx(ctx))
	repo, err := git.PlainOpen(r.repoDir)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}
	_, err = r.commitAll(repo, "Auto commit before merging to "+branch)
	if err != nil && !errors.Is(err, git.ErrEmptyCommit) {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	// Fetch the branch to ensure we are up-to-date
	err = r.fetchBranch(ctx, repo, branch)
	if err != nil {
		return err
	}

	if r.defaultBranch != branch {
		defer func() {
			err := gitCheckout(r.repoDir, r.defaultBranch, true, false, "")
			if err != nil {
				resErr = errors.Join(resErr, fmt.Errorf("failed to checkout default branch %q: %w", r.defaultBranch, err))
				return
			}
		}()

		// Switch to the requested branch, then hard-reset it to the remote tracking ref.
		// Hard reset is safe here because local changes were already committed above.
		err = gitCheckout(r.repoDir, branch, true, false, "")
		if err != nil {
			return fmt.Errorf("failed to checkout branch %q: %w", branch, err)
		}
		err = resetToRemoteTrackingBranch(r.repoDir, branch)
		if err != nil {
			return fmt.Errorf("failed to reset to remote tracking branch %q: %w", branch, err)
		}
	}

	// Merge the default branch into the current branch
	merged := true
	if force {
		err = gitutil.MergeWithStrategy(r.repoDir, r.defaultBranch, "theirs")
	} else {
		merged, err = gitutil.MergeWithBailOnConflict(r.repoDir, r.defaultBranch)
	}
	if err != nil {
		return fmt.Errorf("failed to merge default branch %q into branch %q: %w", r.defaultBranch, branch, err)
	}

	if !merged {
		// If the merge was aborted no need to push the changes
		r.h.logger.Warn("Merge aborted due to conflicts, not pushing changes", zap.String("branch", branch), zap.String("defaultBranch", r.defaultBranch))
		return nil
	}

	// Push the changes
	branches := []string{branch}
	if r.defaultBranch != branch {
		branches = append(branches, r.defaultBranch)
	}
	return r.pushBranch(ctx, branches...)
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
			When: time.Now(),
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
	return r.fetchBranch(ctx, repo, head.Name().Short())
}

// pushBranch pushes the specified branches to the remote repository.
// If the remote branch is already up-to-date, it does not return an error.
// It re-opens the repository so that any objects written by prior shell-based git
// commands (e.g. merge, checkout) are visible to go-git's push.
func (r *gitRepo) pushBranch(ctx context.Context, branches ...string) error {
	if len(branches) == 0 {
		return errors.New("at least one branch must be specified to push")
	}
	repo, err := git.PlainOpen(r.repoDir)
	if err != nil {
		return err
	}
	refSpecs := make([]config.RefSpec, 0, len(branches))
	for _, branch := range branches {
		refSpecs = append(refSpecs, config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch)))
	}
	err = repo.PushContext(ctx, &git.PushOptions{
		RemoteName: "origin",
		RemoteURL:  r.remoteURL,
		RefSpecs:   refSpecs,
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return err
	}
	return nil
}

// fetchBranch fetches the specified branch from the remote repository.
// It does not return an error if the local branch is already up-to-date with the remote branch.
func (r *gitRepo) fetchBranch(ctx context.Context, repo *git.Repository, branch string) error {
	err := repo.FetchContext(ctx, &git.FetchOptions{
		RemoteURL: r.remoteURL,
		RefSpecs:  []config.RefSpec{config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/remotes/origin/%s", branch, branch))},
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return err
	}
	return nil
}

var errRefNotFound = errors.New("reference not found")

// gitCheckout checks out a branch using the git command.
// If create is true, it creates the branch (using -B) at the given startPoint.
// go-git wipes out git-ignored changes during checkout so must use the git command.
func gitCheckout(repoDir, branch string, force, create bool, startPoint string) error {
	args := []string{"-C", repoDir, "checkout"}
	if force {
		args = append(args, "--force")
	}
	if create {
		args = append(args, "-B", branch)
		if startPoint != "" {
			args = append(args, startPoint)
		}
	} else {
		args = append(args, branch)
	}
	cmd := exec.Command("git", args...)
	_, err := cmd.Output()
	if err != nil {
		var execErr *exec.ExitError
		if !errors.As(err, &execErr) {
			return err
		}
		stderr := string(execErr.Stderr)
		if strings.Contains(stderr, "did not match") {
			return errRefNotFound
		}
		return fmt.Errorf("git checkout failed: %s", stderr)
	}
	return nil
}

// resetToRemoteTrackingBranch resets to the commit pointed by the remote tracking branch.
// This is used to reset the local branch to the state of the remote branch so it is expected that the latest changes have been fetched.
// go-git wipes out git-ignored changes so must use the git command.
func resetToRemoteTrackingBranch(repoDir, branch string) error {
	cmd := exec.Command("git", "-C", repoDir, "reset", "--hard", "origin/"+branch)
	_, err := cmd.Output()
	if err != nil {
		var execErr *exec.ExitError
		if !errors.As(err, &execErr) {
			return err
		}
		if strings.Contains(string(execErr.Stderr), "unknown revision") {
			return errRefNotFound
		}
		return fmt.Errorf("git reset failed: %s", string(execErr.Stderr))
	}
	return nil
}

// ensureGitConfig ensures that the git config key is set.
// if not set then it sets the key to the given value locally in the repo
func ensureGitConfig(repoDir, key, value string) error {
	_, err := exec.Command("git", "-C", repoDir, "config", "--get", key).Output()
	if err == nil {
		return nil
	}

	// Exit code 1 means "key not set" — that's the case we want to handle.
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		return err
	}

	// set only locally
	return exec.Command("git", "-C", repoDir, "config", "--local", key, value).Run()
}
