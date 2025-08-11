// gitutil package provides utility functions for working with git repositories.
// To execute git operations we use the go-git library.
// However the library does not support all git operations and in those cases we directly run git commands.
// The utility functions in this file directly run git commands.
package gitutil

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

var errDetachedHead = errors.New("detached HEAD state detected, please checkout a branch")

type GitStatus struct {
	Branch        string
	RemoteURL     string
	LocalChanges  bool // true if there are local changes (staged, unstaged, or untracked)
	LocalCommits  int32
	RemoteCommits int32
}

func (s GitStatus) Equal(v GitStatus) bool {
	return s.Branch == v.Branch && s.LocalCommits == v.LocalCommits && s.RemoteCommits == v.RemoteCommits && s.LocalChanges == v.LocalChanges
}

func RunGitStatus(path, remoteName string) (GitStatus, error) {
	cmd := exec.Command("git", "-C", path, "status", "--porcelain=v2", "--branch")
	data, err := cmd.CombinedOutput()
	if err != nil {
		return GitStatus{}, err
	}

	// parse the output
	// Format is
	// # branch.oid 4954f542d4b1f652bba02064aa8ee64ece38d02a
	// # branch.head mgd_repo_poc
	// # branch.upstream origin/mgd_repo_poc
	// # branch.ab +0 -0
	// lines describing the status of the working tree
	status := GitStatus{}
	lines := strings.SplitSeq(strings.TrimSpace(string(data)), "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		switch {
		// standard headers - all may not be present
		case strings.HasPrefix(line, "# branch.oid "):
		case strings.HasPrefix(line, "# branch.head "):
			if strings.HasSuffix(line, "(detached)") {
				return status, errDetachedHead
			}
			status.Branch = strings.TrimPrefix(line, "# branch.head ")
		case strings.HasPrefix(line, "# branch.upstream "):
		case strings.HasPrefix(line, "# branch.ab "):
			// do not use this as the remote tracking branch may not be set/may be set to a different remote
		default:
			// any non header line means staged, unstaged or untracked changes
			status.LocalChanges = true
		}
	}

	localCommits, err := countCommitsAhead(path, fmt.Sprintf("%s/%s", remoteName, status.Branch), status.Branch)
	if err == nil {
		status.LocalCommits = localCommits
	}
	remoteCommits, err := countCommitsAhead(path, status.Branch, fmt.Sprintf("%s/%s", remoteName, status.Branch))
	if err == nil {
		status.RemoteCommits = remoteCommits
	}

	// get the remote URL
	data, err = exec.Command("git", "-C", path, "remote", "get-url", remoteName).Output()
	if err == nil {
		status.RemoteURL = strings.TrimSpace(string(data))
	}
	return status, nil
}

func RunGitFetch(ctx context.Context, path, remote string) error {
	args := []string{"-C", path, "fetch"}
	if remote != "" {
		// if remote is specified, fetch from that remote
		args = append(args, remote)
	}
	cmd := exec.CommandContext(ctx, "git", args...)
	_, err := cmd.Output()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			return fmt.Errorf("git fetch failed: %s", string(execErr.Stderr))
		}
		return err
	}
	return nil
}

// RunGitPull runs a git pull command in the specified path.
func RunGitPull(ctx context.Context, path string, discardLocal bool, remote, remoteName string) (string, error) {
	st, err := RunGitStatus(path, remoteName)
	if err != nil {
		return "", err
	}

	if discardLocal {
		// instead of doing a hard clean, do a stash instead
		if st.LocalChanges {
			cmd := exec.CommandContext(ctx, "git", "-C", path, "stash", "--include-untracked")
			if err := cmd.Run(); err != nil {
				return "", fmt.Errorf("failed to remove local changes: %w", err)
			}
		}

		if st.LocalCommits > 0 {
			// reset the local commits and stash the changes
			cmd := exec.CommandContext(ctx, "git", "-C", path, "reset", "--mixed", fmt.Sprintf("HEAD~%d", st.LocalCommits))
			if err := cmd.Run(); err != nil {
				return "", fmt.Errorf("failed to reset local commits: %w", err)
			}
			// stash the changes
			cmd = exec.CommandContext(ctx, "git", "-C", path, "stash", "--include-untracked")
			if err := cmd.Run(); err != nil {
				return "", fmt.Errorf("failed to remove local changes: %w", err)
			}
		}
	}

	// git -C <path> pull <remote> <branch>
	args := []string{"-C", path, "pull"}

	if remote != "" {
		args = append(args, remote, st.Branch)
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	_, err = cmd.Output()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			// This is error msg returned by git when pull fails
			return string(execErr.Stderr), nil
		}
		return "", err
	}
	// Skip the normal output of git pull, just return an empty string
	return "", nil
}

func InferGitRepoRoot(path string) (string, error) {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--show-toplevel")
	data, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// countCommitsAhead counts the number of commits in `from` branch not present in `to` branch.
func countCommitsAhead(path, to, from string) (int32, error) {
	cmd := exec.Command("git", "-C", path, "rev-list", "--count", fmt.Sprintf("%s..%s", to, from))
	data, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("failed to count commits: %w", err)
	}
	count, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to parse commit count: %w", err)
	}
	return int32(count), nil
}
