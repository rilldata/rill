package gitutil

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rilldata/rill/runtime/pkg/fileutil"
)

var ErrNotAGitRepository = errors.New("not a git repository")

// IsGitRepo reports whether path is inside a git working tree.
// Returns true for the repo root as well as any subdirectory of it.
func IsGitRepo(path string) bool {
	_, err := Run(context.Background(), path, "rev-parse", "--is-inside-work-tree")
	return err == nil
}

// EnsureInit initializes a git repository at path if one does not already exist.
// If defaultBranch is non-empty, the initial (unborn) branch is named after it.
// It is a no-op if path already contains a .git entry (a directory in a regular checkout,
// or a file in a linked worktree).
func EnsureInit(ctx context.Context, path, defaultBranch string) error {
	if _, err := os.Stat(filepath.Join(path, ".git")); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if _, err := Run(ctx, path, "init"); err != nil {
		return err
	}
	if defaultBranch != "" {
		// rename the unborn branch; avoids `git init -b`, which requires git >= 2.28
		if _, err := Run(ctx, path, "symbolic-ref", "HEAD", "refs/heads/"+defaultBranch); err != nil {
			return err
		}
	}
	return nil
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

// CloneWithConfig clones the remote described by config into path, naming the remote
// config.RemoteName() and checking out config.DefaultBranch.
// Only the clean config.Remote URL is persisted in the repo's config; credentials are
// passed exclusively on the command line.
func CloneWithConfig(ctx context.Context, path string, config *Config) error {
	remote, err := config.FullyQualifiedRemote()
	if err != nil {
		return err
	}
	args := []string{"clone", remote, path, "--origin", config.RemoteName(), "--single-branch"}
	if config.DefaultBranch != "" {
		args = append(args, "--branch", config.DefaultBranch)
	}
	if _, err := Run(ctx, "", args...); err != nil {
		return err
	}
	if config.Username == "" {
		return nil
	}
	// scrub the credential-embedded URL that clone persisted in .git/config
	if _, err := Run(ctx, path, "remote", "set-url", config.RemoteName(), config.Remote); err != nil {
		// never leave credentials on disk
		_ = os.RemoveAll(path)
		return err
	}
	return nil
}

// CloneRepo clones the repository at repoURL into the current working directory and returns the repository name.
func CloneRepo(ctx context.Context, repoURL string) (string, error) {
	_, remotePath, ok := splitRemote(repoURL)
	if !ok || remotePath == "" {
		return "", fmt.Errorf("invalid git remote %q", repoURL)
	}
	repoName := fileutil.Stem(remotePath)

	if _, err := Run(ctx, "", "clone", repoURL); err != nil {
		return "", err
	}
	return repoName, nil
}

// InferRepoRoot returns the root of the git working tree containing path.
// Returns ErrNotAGitRepository if path is not inside a git working tree.
func InferRepoRoot(path string) (string, error) {
	out, err := Run(context.Background(), path, "rev-parse", "--show-cdup")
	if err != nil {
		if strings.Contains(err.Error(), "not a git repository") {
			return "", ErrNotAGitRepository
		}
		return "", err
	}
	return filepath.Join(path, out), nil
}

// InferRepoRootAndSubpath infers the root of the Git repository and the subpath from the given path.
// Since the extraction stops at first .git directory it means that if a subpath in a github monorepo is deployed as a rill managed project it will prevent the subpath from being inferred.
// This means :
// - user will need to explicitly set the subpath if they want to connect this to Github.
// - When finding matching projects it will only list the rill managed projects for that subpath.
func InferRepoRootAndSubpath(path string) (string, string, error) {
	// check if is a git repository
	repoRoot, err := InferRepoRoot(path)
	if err != nil {
		return "", "", err
	}

	// infer subpath if it exists
	subPath, err := filepath.Rel(repoRoot, path)
	if err != nil {
		// should never happen because repoRoot is detected from path
		return "", "", err
	}
	if subPath == "." || subPath == "" {
		// no subpath
		return repoRoot, "", nil
	}
	// check if subpath is in .gitignore
	ignored, err := isGitIgnored(repoRoot, subPath)
	if err != nil {
		return "", "", err
	}
	if ignored {
		// if subpath is ignored this is not a valid git path
		return "", "", ErrNotAGitRepository
	}
	return repoRoot, subPath, nil
}

// isGitIgnored reports whether subpath is ignored by git in the repository at repoRoot.
func isGitIgnored(repoRoot, subpath string) (bool, error) {
	_, err := Run(context.Background(), repoRoot, "check-ignore", subpath)
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			// exit code 1 means the path is not ignored
			return false, nil
		}
		return false, fmt.Errorf("git check-ignore failed for path %q, subpath %q: %w", repoRoot, subpath, err)
	}
	// exit code 0 means the path is ignored
	return true, nil
}
