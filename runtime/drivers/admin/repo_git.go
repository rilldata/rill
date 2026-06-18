package admin

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
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
	if !isRepoRoot(r.repoDir) {
		// Repository doesn't exist or is invalid, remove and clone fresh
		if err := os.RemoveAll(r.repoDir); err != nil {
			return err
		}
		// for non editable make the clone faster by only cloning the primary branch with no history
		// for editable deployments, we need the full history and all branches to support editing and merging
		err := gitutil.Clone(ctx, r.repoDir, r.remoteURL, r.primaryBranch, !r.editableDepl, !r.editableDepl)
		if err != nil {
			return err
		}

		if r.editableDepl {
			// set git config in the repo dir to ensure git commits/git merge etc pass on cloud
			err = setGitConfig(r.repoDir, "user.name", "Rill")
			if err != nil {
				return err
			}
			err = setGitConfig(r.repoDir, "user.email", "noreply@rilldata.com")
			if err != nil {
				return err
			}
		}
	} else {
		// Repository exists, pull latest changes

		// Ensure the remote URL is correct
		err := setRemoteURL(r.repoDir, r.remoteURL)
		if err != nil {
			return err
		}

		if !r.editableDepl {
			// For non-editable repos, we only care about the primary branch, so set the fetch refspec to only fetch the primary branch.
			err := setFetchBranch(r.repoDir, r.primaryBranch)
			if err != nil {
				return err
			}
		}

		// Fetch the remote changes.
		branchesToFetch := []string{r.primaryBranch}
		if r.editableDepl && r.primaryBranch != r.defaultBranch {
			branchesToFetch = append(branchesToFetch, r.defaultBranch)
		}
		err = gitutil.FetchBranches(ctx, r.repoDir, branchesToFetch...)
		if err != nil {
			return err
		}
	}

	// Checkout the default branch
	var createDefault bool
	err := gitutil.Checkout(r.repoDir, r.defaultBranch, force, false, "")
	if err != nil {
		if !errors.Is(err, gitutil.ErrRefNotFound) {
			return fmt.Errorf("failed to checkout branch %q: %w", r.defaultBranch, err)
		}

		// The default branch does not exist locally, try to find it on remote. It can happen in cases:
		// a. when branch is created remotely after the last pull
		// b. primary branch was edited
		// c. in editable mode when the default branch may not exist at all.
		remoteHash, err := gitutil.Hash(ctx, r.repoDir, "refs/remotes/origin/"+r.defaultBranch)
		if err != nil {
			if errors.Is(err, gitutil.ErrRefNotFound) && r.editable() {
				// In editable mode, the default branch may not exist yet on remote. We will create it based on the primary branch.
				r.h.logger.Info("Default branch does not exist on remote, will create it based on primary branch", zap.String("defaultBranch", r.defaultBranch), zap.String("primaryBranch", r.primaryBranch))
				remoteHash, err = gitutil.Hash(ctx, r.repoDir, "refs/remotes/origin/"+r.primaryBranch)
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
		err = gitutil.Checkout(r.repoDir, r.defaultBranch, true, true, remoteHash)
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
			if !(errors.Is(err, gitutil.ErrRefNotFound) && r.editable()) { // In editable mode, the default branch may not exist yet on remote.
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
				return &drivers.MergeFailedError{
					Output:       "local is behind remote and failed to sync with remote due to conflicts, use force pull to discard local changes and sync with remote",
					MergedBranch: r.defaultBranch,
					Conflict:     true,
				}
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
		var merged bool
		merged, err = gitutil.MergeWithBailOnConflict(r.repoDir, mergeBranch)
		if !merged && userTriggered { // only user triggered pulls should fail on conflicts
			return &drivers.MergeFailedError{
				Output:       "failed to merge primary branch, use force pull to discard local changes and sync with primary branch",
				MergedBranch: r.primaryBranch,
				Conflict:     true,
			}
		}
	}
	if err != nil {
		return &drivers.MergeFailedError{
			Output:       fmt.Sprintf("failed to merge primary branch %q into default branch %q: %v", r.primaryBranch, r.defaultBranch, err),
			MergedBranch: r.primaryBranch,
		}
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

	_, err := gitutil.CommitAll(ctx, r.repoDir, r.subpath, message, "Rill", "noreply@rilldata.com")
	if err != nil {
		if !errors.Is(err, gitutil.ErrEmptyCommit) {
			return fmt.Errorf("failed to commit changes to edit branch: %w", err)
		}
		// continue to push existing commits, if any
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
	_, err := gitutil.CommitAll(ctx, r.repoDir, r.subpath, "Auto commit before merging to "+branch, "Rill", "noreply@rilldata.com")
	if err != nil && !errors.Is(err, gitutil.ErrEmptyCommit) {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	// Fetch the branch to ensure we are up-to-date
	err = gitutil.FetchBranches(ctx, r.repoDir, branch)
	if err != nil {
		return err
	}

	if r.defaultBranch != branch {
		defer func() {
			err := gitutil.Checkout(r.repoDir, r.defaultBranch, true, false, "")
			if err != nil {
				resErr = errors.Join(resErr, fmt.Errorf("failed to checkout default branch %q: %w", r.defaultBranch, err))
				return
			}
		}()

		// Switch to the requested branch, then hard-reset it to the remote tracking ref.
		// Hard reset is safe here because local changes were already committed above.
		err = gitutil.Checkout(r.repoDir, branch, true, false, "")
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
		// wrap with drivers.ErrMergeFailed
		return &drivers.MergeFailedError{
			Output:       fmt.Sprintf("failed to merge default branch %q into branch %q: %v", r.defaultBranch, branch, err),
			MergedBranch: r.defaultBranch,
		}
	}

	if !merged {
		return &drivers.MergeFailedError{
			Output:       "merge failed due to conflicts, use force merge to favour current changes",
			MergedBranch: r.defaultBranch,
			Conflict:     true,
		}
	}

	// Push the changes
	branches := []string{branch}
	if r.defaultBranch != branch {
		branches = append(branches, r.defaultBranch)
	}
	return r.pushBranch(ctx, branches...)
}

// commitHash returns the current commit hash of the repository.
// It returns an empty string (without error) when HEAD points to an unborn branch.
func (r *gitRepo) commitHash(ctx context.Context) (string, error) {
	hash, err := gitutil.Hash(ctx, r.repoDir, "HEAD")
	if errors.Is(err, gitutil.ErrRefNotFound) {
		return "", nil
	}
	return hash, err
}

// commitTimestamp returns the timestamp of the latest commit on the current branch.
func (r *gitRepo) commitTimestamp(ctx context.Context) (time.Time, error) {
	out, err := gitutil.Run(ctx, r.repoDir, "log", "-1", "--format=%aI", "HEAD")
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339, out)
}

func (r *gitRepo) fetchCurrentBranch(ctx context.Context) error {
	branch, err := currentBranch(r.repoDir)
	if err != nil {
		return err
	}
	return gitutil.FetchBranches(ctx, r.repoDir, branch)
}

// pushBranch pushes the specified branches to the remote repository.
func (r *gitRepo) pushBranch(ctx context.Context, branches ...string) error {
	if len(branches) == 0 {
		return errors.New("at least one branch must be specified to push")
	}
	args := append([]string{"push", "origin"}, branches...)
	_, err := gitutil.Run(ctx, r.repoDir, args...)
	return err
}

// resetToRemoteTrackingBranch resets to the commit pointed by the remote tracking branch.
// This is used to reset the local branch to the state of the remote branch so it is expected that the latest changes have been fetched.
// go-git wipes out git-ignored changes so must use the git command.
func resetToRemoteTrackingBranch(repoDir, branch string) error {
	_, err := gitutil.Run(context.Background(), repoDir, "reset", "--hard", "origin/"+branch)
	if err != nil {
		if strings.Contains(err.Error(), "unknown revision") {
			return gitutil.ErrRefNotFound
		}
		return err
	}
	return nil
}

// setGitConfig sets the git config key locally in the repo.
func setGitConfig(repoDir, key, value string) error {
	_, err := gitutil.Run(context.Background(), repoDir, "config", "--local", key, value)
	return err
}

func isRepoRoot(path string) bool {
	out, err := gitutil.Run(context.Background(), path, "rev-parse", "--show-cdup")
	if err != nil {
		return false
	}
	return out == ""
}

func setRemoteURL(path, remoteURL string) error {
	_, err := gitutil.Run(context.Background(), path, "remote", "set-url", "origin", remoteURL)
	return err
}

func setFetchBranch(path, branch string) error {
	_, err := gitutil.Run(context.Background(), path, "remote", "set-branches", "origin", branch)
	return err
}

func currentBranch(path string) (string, error) {
	return gitutil.Run(context.Background(), path, "rev-parse", "--abbrev-ref", "HEAD")
}
