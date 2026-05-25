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
	if !isGitRepo(r.repoDir) {
		// Repository doesn't exist or is invalid, remove and clone fresh
		if err := os.RemoveAll(r.repoDir); err != nil {
			return err
		}

		if err := gitClone(ctx, r.repoDir, r.remoteURL, r.primaryBranch, r.primaryBranch == r.defaultBranch); err != nil {
			return err
		}

		if r.editableDepl {
			// set git config in the repo dir to ensure git commits/git merge etc pass on cloud
			if err := ensureGitConfig(r.repoDir, "user.name", "Rill"); err != nil {
				return err
			}
			if err := ensureGitConfig(r.repoDir, "user.email", "noreply@rilldata.com"); err != nil {
				return err
			}
		}
	} else {
		// Repository exists, pull latest changes

		// Ensure the remote URL is correct
		if err := setGitRemote(r.repoDir, "origin", r.remoteURL); err != nil {
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
			if err := fetchBranch(ctx, r.repoDir, r.remoteURL, branch, true); err != nil {
				return fmt.Errorf("failed to fetch from remote: %w", err)
			}
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
		remoteHash, err := gitResolveRef(r.repoDir, "refs/remotes/origin/"+r.defaultBranch)
		if err != nil {
			if errors.Is(err, gitutil.ErrRefNotFound) && r.editable() {
				// In editable mode, the default branch may not exist yet on remote. We will create it based on the primary branch.
				r.h.logger.Info("Default branch does not exist on remote, will create it based on primary branch", zap.String("defaultBranch", r.defaultBranch), zap.String("primaryBranch", r.primaryBranch))
				remoteHash, err = gitResolveRef(r.repoDir, "refs/remotes/origin/"+r.primaryBranch)
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
		err = gitutil.ResetToRemote(r.repoDir, r.defaultBranch)
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

	_, err := gitutil.CommitAll(r.repoDir, "", message, "Rill", "noreply@rilldata.com")
	if err != nil {
		return fmt.Errorf("failed to commit changes to edit branch: %w", err)
	}
	// if hash == "", there was nothing to commit; continue to push existing commits

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

	_, err := gitutil.CommitAll(r.repoDir, "", "Auto commit before merging to "+branch, "Rill", "noreply@rilldata.com")
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	// Fetch the branch to ensure we are up-to-date
	if err := r.fetchBranch(ctx, branch); err != nil {
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
		err = gitutil.ResetToRemote(r.repoDir, branch)
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
func (r *gitRepo) commitHash() (string, error) {
	return gitHeadHash(r.repoDir)
}

// commitTimestamp returns the timestamp of the latest commit on the current branch.
func (r *gitRepo) commitTimestamp() (time.Time, error) {
	hash, err := gitHeadHash(r.repoDir)
	if err != nil {
		return time.Time{}, err
	}
	if hash == "" {
		return time.Time{}, nil
	}
	return gitCommitTimestamp(r.repoDir, hash)
}

func (r *gitRepo) fetchCurrentBranch(ctx context.Context) error {
	branch, err := gitutil.CurrentBranch(r.repoDir)
	if err != nil {
		return err
	}
	return r.fetchBranch(ctx, branch)
}

// pushBranch pushes the specified branches to the remote repository.
// If the remote branch is already up-to-date, it does not return an error.
func (r *gitRepo) pushBranch(ctx context.Context, branches ...string) error {
	if len(branches) == 0 {
		return errors.New("at least one branch must be specified to push")
	}
	return gitPushBranches(ctx, r.repoDir, r.remoteURL, branches...)
}

// fetchBranch fetches the specified branch(es) from the remote repository.
// It does not return an error if the local branch is already up-to-date with the remote branch.
func (r *gitRepo) fetchBranch(ctx context.Context, branch ...string) error {
	for _, b := range branch {
		if err := fetchBranch(ctx, r.repoDir, r.remoteURL, b, false); err != nil {
			return err
		}
	}
	return nil
}

// isGitRepo reports whether repoDir is a valid git repository.
func isGitRepo(repoDir string) bool {
	return exec.Command("git", "-C", repoDir, "rev-parse", "--git-dir").Run() == nil
}

// gitClone clones a remote repository into repoDir on the specified branch.
func gitClone(ctx context.Context, repoDir, remoteURL, branch string, singleBranch bool) error {
	args := []string{"clone"}
	if singleBranch {
		args = append(args, "--single-branch")
	}
	args = append(args, "-b", branch, remoteURL, repoDir)
	out, err := exec.CommandContext(ctx, "git", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %s", string(out))
	}
	return nil
}

// setGitRemote replaces the named remote with the given URL, removing it first if it exists.
func setGitRemote(repoDir, name, url string) error {
	_ = exec.Command("git", "-C", repoDir, "remote", "remove", name).Run()
	out, err := exec.Command("git", "-C", repoDir, "remote", "add", name, url).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git remote add failed: %s", string(out))
	}
	return nil
}

// fetchBranch fetches a single branch from the remote URL into refs/remotes/origin/<branch>.
// It silently ignores "no matching remote ref" errors (branch doesn't exist on remote yet).
func fetchBranch(ctx context.Context, repoDir, remoteURL, branch string, force bool) error {
	refSpec := fmt.Sprintf("refs/heads/%s:refs/remotes/origin/%s", branch, branch)
	args := []string{"-C", repoDir, "fetch"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, remoteURL, refSpec)
	out, err := exec.CommandContext(ctx, "git", args...).CombinedOutput()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			if strings.Contains(string(out), "couldn't find remote ref") {
				return nil // branch doesn't exist on remote yet
			}
			return fmt.Errorf("git fetch failed: %s", string(out))
		}
		return err
	}
	return nil
}

// gitResolveRef resolves a full ref name (e.g. refs/remotes/origin/main) to a commit hash.
// Returns gitutil.ErrRefNotFound if the ref does not exist.
func gitResolveRef(repoDir, ref string) (string, error) {
	out, err := exec.Command("git", "-C", repoDir, "rev-parse", ref).Output()
	if err != nil {
		return "", gitutil.ErrRefNotFound
	}
	return strings.TrimSpace(string(out)), nil
}

// gitHeadHash returns the commit hash of HEAD, or "" if the repo has no commits yet.
func gitHeadHash(repoDir string) (string, error) {
	out, err := exec.Command("git", "-C", repoDir, "rev-parse", "HEAD").Output()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) && strings.Contains(string(execErr.Stderr), "unknown revision") {
			return "", nil
		}
		return "", fmt.Errorf("git rev-parse HEAD failed: %s", string(execErr.Stderr))
	}
	return strings.TrimSpace(string(out)), nil
}

// gitCommitTimestamp returns the author timestamp of a commit.
func gitCommitTimestamp(repoDir, hash string) (time.Time, error) {
	out, err := exec.Command("git", "-C", repoDir, "log", "-1", "--format=%aI", hash).Output()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			return time.Time{}, fmt.Errorf("git log failed: %s", string(execErr.Stderr))
		}
		return time.Time{}, err
	}
	ts := strings.TrimSpace(string(out))
	if ts == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, ts)
}

// gitPushBranches pushes one or more branches to the remote URL.
// Treats "Everything up-to-date" as success.
func gitPushBranches(ctx context.Context, repoDir, remoteURL string, branches ...string) error {
	args := []string{"-C", repoDir, "push", remoteURL}
	for _, branch := range branches {
		args = append(args, fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch))
	}
	out, err := exec.CommandContext(ctx, "git", args...).CombinedOutput()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			if strings.Contains(string(out), "Everything up-to-date") {
				return nil
			}
			return fmt.Errorf("git push failed: %s", string(out))
		}
		return err
	}
	return nil
}

// ensureGitConfig sets a git config key locally if not already set at any scope.
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
