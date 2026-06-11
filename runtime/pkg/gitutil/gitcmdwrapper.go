package gitutil

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var ErrRefNotFound = errors.New("git reference not found")

// ErrEmptyCommit is returned by CommitAll when there are no changes to commit.
var ErrEmptyCommit = errors.New("nothing to commit")

// IsGitRepo reports whether path is inside a git working tree.
// Returns true for the repo root as well as any subdirectory of it.
func IsGitRepo(path string) bool {
	_, err := Run(context.Background(), path, "rev-parse", "--is-inside-work-tree")
	return err == nil
}

func Clone(ctx context.Context, path, remote, checkoutBranch string, singleBranch, shallow bool) error {
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

	_, err := Run(ctx, "", args...)
	return err
}

// Checkout checks out a branch using the git command.
// If create is true, it creates the branch (using -B) at the given startPoint.
// go-git wipes out git-ignored changes during checkout so must use the git command.
func Checkout(repoDir, branch string, force, create bool, startPoint string) error {
	args := []string{"checkout"}
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
	_, err := Run(context.Background(), repoDir, args...)
	if err != nil {
		if strings.Contains(err.Error(), "did not match") {
			return ErrRefNotFound
		}
		return err
	}
	return nil
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
// Returns an error if the merge failed for a reason other than conflicts, or if both merge and abort fail.
func MergeWithBailOnConflict(path, branch string) (bool, error) {
	_, mergeErr := Run(context.Background(), path, "merge", "--no-ff", branch)
	if mergeErr == nil {
		return true, nil
	}

	// Detect "merge in progress" via MERGE_HEAD: presence means git stopped on conflicts and is waiting for resolution.
	// Other merge failures (e.g., invalid ref, untracked-file overwrite that git auto-aborts) leave no MERGE_HEAD,
	// and should surface as real errors rather than be silently treated as conflicts.
	merging, err := mergeInProgress(path)
	if err != nil {
		return false, fmt.Errorf("merge failed: %w; could not check merge state: %w", mergeErr, err)
	}
	if !merging {
		return false, mergeErr
	}

	if _, abortErr := Run(context.Background(), path, "merge", "--abort"); abortErr != nil {
		return false, fmt.Errorf("merge failed with error: %w, and abort also failed with error: %w", mergeErr, abortErr)
	}
	return false, nil
}

// mergeInProgress reports whether the repository at path is in the middle of a merge.
func mergeInProgress(path string) (bool, error) {
	mergeHeadPath, err := Run(context.Background(), path, "rev-parse", "--git-path", "MERGE_HEAD")
	if err != nil {
		return false, err
	}
	if !filepath.IsAbs(mergeHeadPath) {
		mergeHeadPath = filepath.Join(path, mergeHeadPath)
	}
	if _, err := os.Stat(mergeHeadPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
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

	return Hash(ctx, path, "HEAD")
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

// IsCommitHash reports whether s is a full hex commit hash (SHA-1 or SHA-256).
// Use it to validate untrusted hashes before passing them as git CLI arguments: it rules out
// strings that git would interpret as flags or other revision syntax.
func IsCommitHash(s string) bool {
	if len(s) != 40 && len(s) != 64 {
		return false
	}
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}

// Hash returns the commit hash for the given ref. Returns ErrRefNotFound if the ref does not resolve.
func Hash(ctx context.Context, path, ref string) (string, error) {
	out, err := Run(ctx, path, "rev-parse", "--verify", ref)
	if err != nil {
		if strings.Contains(err.Error(), "Needed a single revision") {
			return "", ErrRefNotFound
		}
		return "", err
	}
	return out, nil
}

// Run executes a git command with the specified arguments in the given path and returns its output or an error.
// If path is empty, the command runs without -C (use for commands like `clone` that take an explicit destination).
// Use it to run one-off git commands that don't fit into the other helper functions in this package.
func Run(ctx context.Context, path string, args ...string) (string, error) {
	fullArgs := args
	if path != "" {
		fullArgs = append([]string{"-C", path}, args...)
	}
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", fullArgs...)
	// Force English error messages so stderr substring checks are stable, and disable interactive credential prompts.
	cmd.Env = append(os.Environ(), "LC_ALL=C", "GIT_TERMINAL_PROMPT=0")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		// Redact credentials: args and git's stderr may contain credential-embedded remote URLs.
		msg := fmt.Sprintf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(stderr.String()))
		return "", fmt.Errorf("%s(%w)", redactURLCredentials(msg), err)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// urlCredentialsRegexp matches the userinfo component of a URL (e.g. "https://user:token@host").
var urlCredentialsRegexp = regexp.MustCompile(`([a-zA-Z][a-zA-Z0-9+.-]*://)[^@/\s]+@`)

// redactURLCredentials masks credentials embedded in URLs in s.
func redactURLCredentials(s string) string {
	return urlCredentialsRegexp.ReplaceAllString(s, "$1<redacted>@")
}
