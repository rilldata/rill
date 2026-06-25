package gitutil

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type GitStatus struct {
	Branch        string
	RemoteURL     string
	LocalChanges  bool // true if there are local changes (staged, unstaged, or untracked)
	LocalCommits  int32
	RemoteCommits int32
	// ChangedFiles lists the files that differ from the comparison ref. It is only populated
	// when Status is called with changedFiles set to true.
	ChangedFiles []ChangedFile
}

// ChangedFileStatus describes how a file changed relative to a comparison ref.
type ChangedFileStatus int

const (
	ChangedFileStatusUnspecified ChangedFileStatus = iota
	ChangedFileStatusAdded
	ChangedFileStatusModified
	ChangedFileStatusDeleted
	ChangedFileStatusRenamed
)

// ChangedFile is a single file that differs from a comparison ref.
type ChangedFile struct {
	Path string
	// OldPath is the previous path; only set when Status is ChangedFileStatusRenamed.
	OldPath string
	Status  ChangedFileStatus
}

// Status returns the status of the git repo at path.
// If subpath is non-empty, local changes and ahead/behind counts are scoped to it.
// If remoteBranch is non-empty, ahead/behind counts compare the local branch against
// `<remoteName>/<remoteBranch>` instead of `<remoteName>/<localBranch>`.
// If changedFiles is true, the changed files relative to the comparison ref are listed.
func Status(ctx context.Context, path, subpath, remoteName, remoteBranch string, changedFiles bool) (GitStatus, error) {
	args := []string{"status", "--porcelain=v2", "--branch"}
	if subpath != "" {
		args = append(args, "--", subpath)
	}
	out, err := Run(ctx, path, args...)
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
	lines := strings.SplitSeq(strings.TrimSpace(out), "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		switch {
		// standard headers - all may not be present
		case strings.HasPrefix(line, "# branch.oid "):
		case strings.HasPrefix(line, "# branch.head "):
			if strings.HasSuffix(line, "(detached)") {
				return status, ErrDetachedHead
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

	compareBranch := remoteBranch
	if compareBranch == "" {
		compareBranch = status.Branch
	}
	remoteRef := fmt.Sprintf("%s/%s", remoteName, compareBranch)

	ahead, behind, err := countAheadBehind(ctx, path, subpath, status.Branch, remoteRef)
	if err == nil {
		status.LocalCommits = ahead
		status.RemoteCommits = behind
	}

	// Only list changed files when the caller opts in; skip the extra git work for the
	// frequently-polled current-branch status. Best-effort: like the ahead/behind counts above,
	// a failure here (e.g. the remote-tracking ref not resolving) must not break the status,
	// since the merge flow depends on it.
	if changedFiles {
		if files, err := listChangedFiles(ctx, path, subpath, remoteRef); err == nil {
			status.ChangedFiles = files
		}
	}

	// get the remote URL
	remoteURL, err := Run(ctx, path, "remote", "get-url", remoteName)
	if err == nil {
		status.RemoteURL = remoteURL
	}
	return status, nil
}

// changedFiles returns the files that differ between the working tree at path and ref, i.e. the
// changes that would land on ref if merged. Paths are returned relative to subpath (with the
// subpath prefix stripped). Uncommitted and untracked changes are included, since those are
// committed before a merge (see MergeToBranch). Renames are reported with status Renamed and the
// old path, but only when git can detect them (a committed or staged rename); a rename that is
// still uncommitted in the working tree appears as a deleted old path plus an added new path,
// because git cannot pair an untracked new file with a deleted old one.
func listChangedFiles(ctx context.Context, path, subpath, ref string) ([]ChangedFile, error) {
	diffArgs := []string{"diff", "--name-status", "-M", ref}
	if subpath != "" {
		diffArgs = append(diffArgs, "--", subpath)
	}
	diffOut, err := Run(ctx, path, diffArgs...)
	if err != nil {
		return nil, err
	}

	// Keyed by new path so a later untracked-file pass can override a stale entry for the same path.
	changes := map[string]ChangedFile{}
	for line := range strings.SplitSeq(strings.TrimSpace(diffOut), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format is "<code>\t<path>", or "<code>\t<oldpath>\t<newpath>" for renames/copies.
		fields := strings.Split(line, "\t")
		if len(fields) < 2 {
			continue
		}
		status := changedFileStatusFromCode(fields[0][0])
		if status == ChangedFileStatusUnspecified {
			continue
		}
		file := fields[len(fields)-1]
		change := ChangedFile{Path: file, Status: status}
		if status == ChangedFileStatusRenamed && len(fields) >= 3 {
			change.OldPath = fields[1]
		}
		changes[file] = change
	}

	// `git diff` does not list untracked files; treat them as added.
	statusArgs := []string{"status", "--porcelain", "--untracked-files=all"}
	if subpath != "" {
		statusArgs = append(statusArgs, "--", subpath)
	}
	statusOut, err := Run(ctx, path, statusArgs...)
	if err != nil {
		return nil, err
	}
	for line := range strings.SplitSeq(statusOut, "\n") {
		if file, ok := strings.CutPrefix(line, "?? "); ok {
			changes[file] = ChangedFile{Path: file, Status: ChangedFileStatusAdded}
		}
	}

	result := make([]ChangedFile, 0, len(changes))
	for _, change := range changes {
		change.Path = trimSubpath(change.Path, subpath)
		if change.OldPath != "" {
			change.OldPath = trimSubpath(change.OldPath, subpath)
		}
		result = append(result, change)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result, nil
}

func trimSubpath(file, subpath string) string {
	return strings.TrimPrefix(strings.TrimPrefix(file, subpath), "/")
}

// changedFileStatusFromCode maps a `git diff --name-status` status code to a ChangedFileStatus.
// Copies are reported as added against their destination path.
func changedFileStatusFromCode(code byte) ChangedFileStatus {
	switch code {
	case 'A', 'C':
		return ChangedFileStatusAdded
	case 'M', 'T':
		return ChangedFileStatusModified
	case 'D':
		return ChangedFileStatusDeleted
	case 'R':
		return ChangedFileStatusRenamed
	default:
		return ChangedFileStatusUnspecified
	}
}

// Fetch fetches the latest changes from the remote described by config, updating the
// remote-tracking refs under `refs/remotes/<remote-name>/`.
// If config is nil or carries no credentials, it fetches from origin relying on git's own
// configuration and credential helpers.
func Fetch(ctx context.Context, path string, config *Config) error {
	if config == nil || config.Username == "" {
		_, err := Run(ctx, path, "fetch", "origin")
		return err
	}
	remote, err := config.FullyQualifiedRemote()
	if err != nil {
		return err
	}
	// Fetching from a URL does not update remote-tracking refs by default, so pass an explicit
	// refspec; Status, Pull, and UpstreamMerge compare against <remote-name>/<branch>.
	_, err = Run(ctx, path, "fetch", remote, "+refs/heads/*:refs/remotes/"+config.RemoteName()+"/*")
	return err
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

// Pull pulls the latest changes from the remote into the current branch.
// If discardLocal is true, local changes and local commits are stashed away first.
// If git rejects the pull (e.g. on divergent branches), the (credential-redacted) git error
// message is returned as the output with a nil error, so callers can surface it to users.
func Pull(ctx context.Context, path string, discardLocal bool, remote, remoteName string) (string, error) {
	// current status of the full repo
	st, err := Status(ctx, path, "", remoteName, "", false)
	if err != nil {
		return "", err
	}

	if discardLocal {
		// when discarding local changes it is okay to discard in full repo and not just in subpath
		// instead of doing a hard clean, do a stash instead
		if st.LocalChanges {
			if _, err := Run(ctx, path, "stash", "--include-untracked"); err != nil {
				return "", fmt.Errorf("failed to remove local changes: %w", err)
			}
		}

		if st.LocalCommits > 0 {
			// reset local commits by moving HEAD to the remote tip; HEAD~N would be wrong
			// because LocalCommits excludes merge commits.
			if _, err := Run(ctx, path, "reset", "--mixed", fmt.Sprintf("%s/%s", remoteName, st.Branch)); err != nil {
				return "", fmt.Errorf("failed to reset local commits: %w", err)
			}
			// stash the changes
			if _, err := Run(ctx, path, "stash", "--include-untracked"); err != nil {
				return "", fmt.Errorf("failed to remove local changes: %w", err)
			}
		}
	}

	args := []string{"pull"}
	if remote != "" {
		args = append(args, remote, st.Branch)
	}
	_, err = Run(ctx, path, args...)
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// git rejected the pull: return its message so callers can display it
			return err.Error(), nil
		}
		return "", err
	}
	// Skip the normal output of git pull, just return an empty string
	return "", nil
}

// Push pushes the given refspec (commonly just a branch name) to the remote.
// remote may be a remote name or a URL, optionally with embedded credentials (which are
// passed as an argument only and redacted from any error).
func Push(ctx context.Context, path, remote, refspec string) error {
	_, err := Run(ctx, path, "push", remote, refspec)
	return err
}

// UpstreamMerge merges the remote tracking branch `<remoteName>/<branch>` into the current branch.
// If favourLocal is true, merge conflicts are resolved in favour of local changes.
func UpstreamMerge(ctx context.Context, path, remoteName, branch string, favourLocal bool) error {
	args := []string{"merge"}
	if favourLocal {
		args = append(args, "-X", "ours")
	}
	args = append(args, fmt.Sprintf("%s/%s", remoteName, branch))
	if _, err := Run(ctx, path, args...); err != nil {
		return fmt.Errorf("git merge failed: %w", err)
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

// countAheadBehind returns the number of non-merge commits on `local` not in `remote` (ahead)
// and on `remote` not in `local` (behind), in a single git invocation.
func countAheadBehind(ctx context.Context, path, subpath, local, remote string) (int32, int32, error) {
	args := []string{"rev-list", "--left-right", "--count", "--no-merges", fmt.Sprintf("%s...%s", local, remote)}
	if subpath != "" {
		args = append(args, "--", subpath)
	}
	out, err := Run(ctx, path, args...)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to count commits: %w", err)
	}
	parts := strings.Fields(out)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected rev-list output: %q", out)
	}
	ahead, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse ahead count: %w", err)
	}
	behind, err := strconv.ParseInt(parts[1], 10, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse behind count: %w", err)
	}
	return int32(ahead), int32(behind), nil
}
