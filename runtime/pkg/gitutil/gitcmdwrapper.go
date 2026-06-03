package gitutil

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var ErrRefNotFound = errors.New("git reference not found")

// ErrEmptyCommit is returned by CommitAll when there are no changes to commit.
var ErrEmptyCommit = errors.New("nothing to commit")

func Clone(ctx context.Context, path string, remote, checkoutBranch string, singleBranch, shallow bool) error {
	args := []string{"clone", remote, path}
	if singleBranch {
		args = append(args, "--single-branch")
	}
	if shallow {
		args = append(args, "--depth", "1")
	}
	if checkoutBranch != "" {
		args = append(args, "--branch", checkoutBranch)
	}

	_, err := Run(ctx, ".", args...)
	return err
}

// MergeWithStrategy merge a branch into the current branch using the specified strategy.
func MergeWithStrategy(path, branch, strategy string) error {
	var args []string
	switch strategy {
	case "theirs":
		args = []string{"merge", "-X", "theirs", branch}
	case "ours":
		args = []string{"merge", "-X", "ours", branch}
	case "":
		args = []string{"merge", branch}
	default:
		return fmt.Errorf("internal error: unsupported merge strategy: %s", strategy)
	}

	_, err := Run(context.Background(), path, args...)
	if err != nil {
		return err
	}
	return nil
}

// MergeWithBailOnConflict attempts to merge a branch into the current branch and aborts if there are conflicts.
// Returns true if merge was successful, false if there were conflicts (but abort succeeded).
// Returns an error only if both merge and abort fail.
func MergeWithBailOnConflict(path, branch string) (bool, error) {
	// First try the merge
	_, mergeErr := Run(context.Background(), path, "merge", "--no-ff", branch)
	if mergeErr != nil {
		if strings.Contains(mergeErr.Error(), "Aborting") || strings.Contains(mergeErr.Error(), "Merge with strategy") {
			return false, nil // Merge failed due to conflicts, but git already aborted the merge, so we can just return.
		}
		// fall through to abort
	} else {
		return true, nil // Merge succeeded
	}

	// Merge succeeded with conflicts, now try to abort
	_, abortErr := Run(context.Background(), path, "merge", "--abort")
	if abortErr != nil {
		// Both merge and abort failed
		return false, fmt.Errorf("merge failed with error: %v, and abort also failed with error: %v", mergeErr, abortErr)
	}

	// Merge failed but abort succeeded
	return false, nil
}

// CommitAll stages all changes in the working tree and creates a commit with the given message.
// If pathspec is non-empty, staging and the empty-commit check are scoped to that pathspec.
// The commit's author and committer are set to the provided name and email.
// Returns the new commit hash, or ErrEmptyCommit if there are no changes to commit.
func CommitAll(ctx context.Context, path, pathspec, message, authorName, authorEmail string) (string, error) {
	addArgs := []string{"add", "--all"}
	statusArgs := []string{"status", "--porcelain"}
	if pathspec != "" {
		addArgs = append(addArgs, "--", pathspec)
		statusArgs = append(statusArgs, "--", pathspec)
	}

	if _, err := Run(ctx, path, addArgs...); err != nil {
		return "", err
	}

	status, err := Run(ctx, path, statusArgs...)
	if err != nil {
		return "", err
	}
	if status == "" {
		return "", ErrEmptyCommit
	}

	args := []string{
		"-c", "user.name=" + authorName,
		"-c", "user.email=" + authorEmail,
		"commit", "-m", message,
	}
	if _, err := Run(ctx, path, args...); err != nil {
		return "", err
	}

	return Hash(path, "HEAD")
}

// FetchBranches fetches the specified branches from the remote repository.
// If a branch doesn't exist on the remote, it will be skipped without returning an error.
func FetchBranches(ctx context.Context, path string, branches ...string) error {
	for _, branch := range branches {
		// fetch separately to avoid NoMatchingRefSpecError when one of the branches doesn't exist on remote
		_, err := Run(ctx, path, "fetch", "origin", branch)
		if err != nil {
			if strings.Contains(err.Error(), "find remote ref") {
				continue
			}
			return err
		}
	}
	return nil
}

func Hash(path, ref string) (string, error) {
	out, err := Run(context.Background(), path, "rev-parse", "--verify", ref)
	if err != nil {
		if strings.Contains(err.Error(), "Needed a single revision") {
			return "", ErrRefNotFound
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// Run executes a git command with the specified arguments in the given path and returns its output or an error.
// Use it to run one-off git commands that don't fit into the other helper functions in this package.
func Run(ctx context.Context, path string, args ...string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", path}, args...)...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %s(%w)", strings.Join(args, " "), strings.TrimSpace(stderr.String()), err)
	}
	return strings.TrimSpace(stdout.String()), nil
}
