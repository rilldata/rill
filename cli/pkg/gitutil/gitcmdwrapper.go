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

var (
	errDetachedHead = errors.New("detached HEAD state detected, please checkout a branch")
	ErrLocalBehind  = errors.New("local branch is behind the remote branch")
)

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

func RunGitStatus(path, subpath, remoteName string) (GitStatus, error) {
	var args []string
	if subpath == "" {
		args = []string{"-C", path, "status", "--porcelain=v2", "--branch"}
	} else {
		args = []string{"-C", path, "status", "--porcelain=v2", "--branch", "--", subpath}
	}
	cmd := exec.Command("git", args...)
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

	localCommits, err := countCommitsAhead(path, fmt.Sprintf("%s/%s", remoteName, status.Branch), status.Branch, subpath)
	if err == nil {
		status.LocalCommits = localCommits
	}
	remoteCommits, err := countCommitsAhead(path, status.Branch, fmt.Sprintf("%s/%s", remoteName, status.Branch), subpath)
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
func RunGitPull(ctx context.Context, path string, discardLocal, overwriteLocal bool, remote, remoteName string) (string, error) {
	if discardLocal && overwriteLocal {
		return "", errors.New("cannot discard and overwrite local changes at the same time")
	}
	// current status of the full repo
	st, err := RunGitStatus(path, "", remoteName)
	if err != nil {
		return "", err
	}

	if discardLocal {
		// when discarding local changes it is okay to discard in full repo and not just in subpath
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
			if out, err := cmd.CombinedOutput(); err != nil {
				return "", fmt.Errorf("failed to reset local commits: %s (%w)", string(out), err)
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
	if overwriteLocal {
		args = append(args, "--strategy=ours")
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

func RunGitPush(ctx context.Context, path, remoteName, branchName string, force bool) error {
	var cmd *exec.Cmd
	if force {
		// instead of simply doing a push --force we do a pull and then push to not loose history
		_, err := RunGitPull(ctx, path, false, true, "", remoteName)
		if err != nil {
			return fmt.Errorf("git pull before push failed: %w", err)
		}
		cmd = exec.CommandContext(ctx, "git", "-C", path, "push", remoteName, branchName)
	} else {
		cmd = exec.CommandContext(ctx, "git", "-C", path, "push", remoteName, branchName)
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		if strings.Contains(string(out), "Updates were rejected because the tip of your current branch is behind") {
			return ErrLocalBehind
		}
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			return fmt.Errorf("git push failed: %s(%s)", string(out), string(execErr.Stderr))
		}
		return fmt.Errorf("git push failed: %s(%w)", string(out), err)
	}
	return nil
}

func InferGitRepoRoot(path string) (string, error) {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--show-toplevel")
	data, err := cmd.Output()
	if err != nil {
		var execErr *exec.ExitError
		if !errors.As(err, &execErr) {
			return "", err
		}
		errStr := strings.TrimSpace(string(execErr.Stderr))
		if strings.Contains(errStr, "not a git repository") {
			return "", ErrNotAGitRepository
		}
		return "", errors.New(string(execErr.Stderr))
	}
	return strings.TrimSpace(string(data)), nil
}

func isGitIgnored(repoRoot, subpath string) (bool, error) {
	cmd := exec.Command("git", "-C", repoRoot, "check-ignore", subpath)
	err := cmd.Run()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			// exit code 1 means the file is not ignored
			if execErr.ExitCode() == 1 {
				return false, nil
			}
			// any other exit code is an error
			return false, fmt.Errorf("git check-ignore failed: %s", string(execErr.Stderr))
		}
		return false, fmt.Errorf("git check-ignore failed: %w", err)
	}
	// exit code 0 means the file is ignored
	return true, nil
}

// countCommitsAhead counts the number of commits in `from` branch not present in `to` branch.
func countCommitsAhead(to, from, path, subpath string) (int32, error) {
	var args []string
	if subpath == "" {
		args = []string{"-C", path, "rev-list", "--count", fmt.Sprintf("%s..%s", to, from)}
	} else {
		args = []string{"-C", path, "rev-list", "--count", fmt.Sprintf("%s..%s", to, from), "--", subpath}
	}
	cmd := exec.Command("git", args...)
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
