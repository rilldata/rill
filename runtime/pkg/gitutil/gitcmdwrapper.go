package gitutil

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ErrRefNotFound is returned when a git ref does not exist.
var ErrRefNotFound = errors.New("reference not found")

// MergeWithStrategy merge a branch into the current branch using the specified strategy.
func MergeWithStrategy(path, branch, strategy string) error {
	var args []string
	switch strategy {
	case "theirs":
		args = []string{"-C", path, "merge", "-X", "theirs", branch}
	case "ours":
		args = []string{"-C", path, "merge", "-X", "ours", branch}
	case "":
		args = []string{"-C", path, "merge", branch}
	default:
		return fmt.Errorf("internal error: unsupported merge strategy: %s", strategy)
	}

	cmd := exec.Command("git", args...)
	_, err := cmd.Output()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			return fmt.Errorf("git merge failed: %s", string(execErr.Stderr))
		}
		return err
	}
	return nil
}

// MergeWithBailOnConflict attempts to merge a branch into the current branch and aborts if there are conflicts.
// Returns true if merge was successful, false if there were conflicts (but abort succeeded).
// Returns an error only if both merge and abort fail.
func MergeWithBailOnConflict(path, branch string) (bool, error) {
	// First try the merge
	cmd := exec.Command("git", "-C", path, "merge", "--no-ff", branch)
	_, err := cmd.Output()
	if err == nil {
		// Merge succeeded
		return true, nil
	}

	// Merge failed, try to abort
	abortCmd := exec.Command("git", "-C", path, "merge", "--abort")
	abortErr := abortCmd.Run()
	if abortErr != nil {
		// Both merge and abort failed
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			return false, fmt.Errorf("git merge failed and abort failed: %s", string(execErr.Stderr))
		}
		return false, fmt.Errorf("git merge failed and abort failed: %w", err)
	}

	// Merge failed but abort succeeded
	return false, nil
}

// Checkout checks out a branch. If force is true, discards local changes (--force).
// If create is true, creates or resets the branch to startPoint using -B.
// Returns ErrRefNotFound if the branch does not exist.
func Checkout(repoDir, branch string, force, create bool, startPoint string) error {
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
	_, err := exec.Command("git", args...).Output()
	if err != nil {
		var execErr *exec.ExitError
		if !errors.As(err, &execErr) {
			return err
		}
		stderr := string(execErr.Stderr)
		if strings.Contains(stderr, "did not match") {
			return ErrRefNotFound
		}
		return fmt.Errorf("git checkout failed: %s", stderr)
	}
	return nil
}

// ResetToRemote hard-resets the current branch to the state of its remote tracking branch.
// Returns ErrRefNotFound if the remote tracking branch does not exist.
func ResetToRemote(repoDir, branch string) error {
	_, err := exec.Command("git", "-C", repoDir, "reset", "--hard", "origin/"+branch).Output()
	if err != nil {
		var execErr *exec.ExitError
		if !errors.As(err, &execErr) {
			return err
		}
		if strings.Contains(string(execErr.Stderr), "unknown revision") {
			return ErrRefNotFound
		}
		return fmt.Errorf("git reset failed: %s", string(execErr.Stderr))
	}
	return nil
}

// CurrentBranch returns the short name of the current branch.
func CurrentBranch(repoDir string) (string, error) {
	out, err := exec.Command("git", "-C", repoDir, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			return "", fmt.Errorf("git rev-parse failed: %s", string(execErr.Stderr))
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// CommitAll stages changes and creates a commit with the given author.
// If glob is non-empty, only files matching glob are staged (git add -- glob); otherwise all changes are staged (git add -A).
// Returns ("", nil) if there is nothing to commit.
func CommitAll(repoDir, glob, message, authorName, authorEmail string) (string, error) {
	var addArgs []string
	if glob == "" {
		addArgs = []string{"-C", repoDir, "add", "-A"}
	} else {
		addArgs = []string{"-C", repoDir, "add", "--", glob}
	}
	if out, err := exec.Command("git", addArgs...).CombinedOutput(); err != nil {
		return "", fmt.Errorf("git add failed: %s", string(out))
	}

	cmd := exec.Command("git", "-C", repoDir, "commit", "-m", message)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME="+authorName,
		"GIT_AUTHOR_EMAIL="+authorEmail,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			if strings.Contains(string(out), "nothing to commit") {
				return "", nil
			}
			return "", fmt.Errorf("git commit failed: %s", string(out))
		}
		return "", err
	}

	hashOut, err := exec.Command("git", "-C", repoDir, "rev-parse", "HEAD").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(hashOut)), nil
}
