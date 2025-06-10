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
	LocalChanges  bool // true if there are local changes (staged, unstaged, or untracked)
	LocalCommits  int32
	RemoteCommits int32
}

func (s GitStatus) Equal(v GitStatus) bool {
	return s.Branch == v.Branch && s.LocalCommits == v.LocalCommits && s.RemoteCommits == v.RemoteCommits && s.LocalChanges == v.LocalChanges
}

func RunGitStatus(path string) (GitStatus, error) {
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
			// Should handle detached state ?
			status.Branch = strings.TrimPrefix(line, "# branch.head ")
		case strings.HasPrefix(line, "# branch.upstream "):
		case strings.HasPrefix(line, "# branch.ab "):
			s := strings.Split(line, " ")

			ahead, err := strconv.ParseInt(s[2], 10, 32)
			if err != nil {
				return status, err
			}
			status.LocalCommits = int32(ahead)

			behind, err := strconv.ParseInt(s[3], 10, 32)
			if err != nil {
				return status, err
			}
			if behind < 0 {
				behind = 0 - behind // git status reports negative behind if there are remote commits
			}
			status.RemoteCommits = int32(behind)
		default:
			// any non header line means staged, unstaged or untracked changes
			status.LocalChanges = true
			return status, nil
		}
	}
	return status, nil
}

func GitFetch(ctx context.Context, path, remote string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", path, "fetch", remote)
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

func GitPull(ctx context.Context, path string, discardLocal bool, remote string) (string, error) {
	if discardLocal {
		// instead of doing a hard clean, do a stash instead
		cmd := exec.CommandContext(ctx, "git", "-C", path, "stash", "--include-untracked")
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to remove local changes: %w", err)
		}
	}

	// git -C <path> pull <remote> <branch>
	args := []string{"-C", path, "pull"}
	st, err := RunGitStatus(path)
	if err != nil {
		return "", err
	}
	args = append(args, remote, st.Branch)

	cmd := exec.CommandContext(ctx, "git", args...)
	out, err := cmd.Output()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			return "", fmt.Errorf("git pull failed: %s", string(execErr.Stderr))
		}
		return "", err
	}
	return string(out), nil
}
