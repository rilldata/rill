package gitutil

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var ErrRefNotFound = errors.New("git reference not found")

// ErrEmptyCommit is returned by CommitAll when there are no changes to commit.
var ErrEmptyCommit = errors.New("nothing to commit")

// ErrDetachedHead is returned when an operation requires HEAD to point to a branch.
var ErrDetachedHead = errors.New("detached HEAD state detected, please checkout a branch")

// CurrentBranch returns the name of the branch HEAD points to. It also succeeds on unborn
// branches in repositories without commits. Returns ErrDetachedHead if HEAD is detached.
func CurrentBranch(ctx context.Context, path string) (string, error) {
	out, err := Run(ctx, path, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			// with --quiet, exit code 1 means HEAD is not a symbolic ref
			return "", ErrDetachedHead
		}
		return "", err
	}
	return out, nil
}

// UserSignature returns the git author configured for the repository at path, resolved from
// the combined local, global, and system git config.
// Returns an error if user.name or user.email is not configured.
func UserSignature(ctx context.Context, path string) (Signature, error) {
	name, err := Run(ctx, path, "config", "--get", "user.name")
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return Signature{}, errors.New("git user.name is not set in git config")
		}
		return Signature{}, err
	}
	email, err := Run(ctx, path, "config", "--get", "user.email")
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return Signature{}, errors.New("git user.email is not set in git config")
		}
		return Signature{}, err
	}
	return Signature{Name: name, Email: email}, nil
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

// CommitAll stages all changes in the working tree and creates a commit with the given message.
// If pathspec is non-empty, staging and the empty-commit check are scoped to that pathspec.
// The commit's author and committer are set to the provided author.
// Returns the new commit hash, or ErrEmptyCommit if there are no changes to commit.
func CommitAll(ctx context.Context, path, pathspec, message string, author Signature) (string, error) {
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
		"-c", "user.name=" + author.Name,
		"-c", "user.email=" + author.Email,
		"commit", "-m", message,
	}
	if _, err := Run(ctx, path, args...); err != nil {
		return "", err
	}

	return Hash(ctx, path, "HEAD")
}

// CommitAndPush stages and commits all changes at path (scoped to config.Subpath if set) and
// pushes the current branch to the remote described by config. It initializes a git repository
// at path if one does not exist. If there is nothing new to commit, it still attempts the push.
// Only the clean config.Remote URL is persisted in the repo's config; credentials are passed
// exclusively on the command line.
func CommitAndPush(ctx context.Context, path string, config *Config, commitMsg string, author Signature) error {
	err := EnsureInit(ctx, path, config.DefaultBranch)
	if err != nil {
		return fmt.Errorf("failed to init git repo: %w", err)
	}

	// check current branch matches deployed branch
	branch, err := CurrentBranch(ctx, path)
	if err != nil {
		if errors.Is(err, ErrDetachedHead) {
			return fmt.Errorf("detached HEAD state detected. Checkout a branch")
		}
		return err
	}
	if branch != config.DefaultBranch {
		return fmt.Errorf("current branch %q does not match deployed branch %q", branch, config.DefaultBranch)
	}

	if commitMsg == "" {
		commitMsg = "Auto committed by Rill"
	}
	_, err = CommitAll(ctx, path, config.Subpath, commitMsg, author)
	if err != nil && !errors.Is(err, ErrEmptyCommit) {
		// on ErrEmptyCommit we still trigger the push
		return fmt.Errorf("failed to commit files to git: %w", err)
	}

	// set remote and push the changes
	err = SetRemote(path, config)
	if err != nil {
		return err
	}

	if config.Username == "" {
		// if no credentials are provided we assume it is the user's self-managed repo and auth is already set up in git
		return Push(ctx, path, config.RemoteName(), config.DefaultBranch)
	}

	remote, err := config.FullyQualifiedRemote()
	if err != nil {
		return err
	}
	return Push(ctx, path, remote, config.DefaultBranch)
}

// CommitAndForcePush is similar to CommitAndPush but force pushes the local changes to the remote.
// Unlike CommitAndPush, the current local branch need not match config.DefaultBranch, and HEAD may
// be detached.
func CommitAndForcePush(ctx context.Context, path string, config *Config, commitMsg string, author Signature) error {
	err := EnsureInit(ctx, path, config.DefaultBranch)
	if err != nil {
		return fmt.Errorf("failed to init git repo: %w", err)
	}

	if commitMsg == "" {
		commitMsg = "Auto committed by Rill"
	}
	_, err = CommitAll(ctx, path, config.Subpath, commitMsg, author)
	if err != nil && !errors.Is(err, ErrEmptyCommit) {
		return fmt.Errorf("failed to commit files to git: %w", err)
	}

	err = SetRemote(path, config)
	if err != nil {
		return err
	}

	if config.Username == "" {
		return ForcePush(ctx, path, config.RemoteName(), config.DefaultBranch)
	}

	remote, err := config.FullyQualifiedRemote()
	if err != nil {
		return err
	}
	return ForcePush(ctx, path, remote, config.DefaultBranch)
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
