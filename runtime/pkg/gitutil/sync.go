package gitutil

import (
	"bytes"
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
func Status(ctx context.Context, path, subpath, remoteName, remoteBranch string) (GitStatus, error) {
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

	// get the remote URL
	remoteURL, err := Run(ctx, path, "remote", "get-url", remoteName)
	if err == nil {
		status.RemoteURL = remoteURL
	}
	return status, nil
}

// ChangedFiles returns the files that differ between the working tree at path and the comparison
// ref, i.e. the changes that would land on the ref if merged.
//
// The diff is computed from the merge base of HEAD and the remote ref, not directly against the
// remote ref. This ensures that commits on the remote that have not been pulled locally do not
// appear as spurious inverse changes in the result.
//
// Paths are returned relative to subpath (with the subpath prefix stripped). Uncommitted and
// untracked changes are included, since those are committed before a merge.
// Renames are reported with status Renamed and the old path, but only when git can detect them (a
// committed or staged rename); a rename that is still uncommitted in the working tree appears as a
// deleted old path plus an added new path, because git cannot pair an untracked new file with a
// deleted old one.
func ChangedFiles(ctx context.Context, path, subpath, remoteName, remoteBranch string) ([]ChangedFile, error) {
	compareBranch := remoteBranch
	if compareBranch == "" {
		branch, err := Run(ctx, path, "rev-parse", "--abbrev-ref", "HEAD")
		if err != nil {
			return nil, err
		}
		compareBranch = branch
	}
	ref := fmt.Sprintf("%s/%s", remoteName, compareBranch)

	mergeBase, err := Run(ctx, path, "merge-base", "HEAD", ref)
	if err != nil {
		return nil, err
	}
	mergeBase = strings.TrimSpace(mergeBase)

	diffArgs := []string{"diff", "--name-status", "-M", mergeBase}
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

// maxFileDiffBytes caps the size of a single file's diff in the combined patch returned by Diff.
// A file whose diff exceeds this (e.g. a large generated or accidentally committed data file) is
// suppressed to a "diff too large" placeholder so one huge file cannot bloat the response.
const maxFileDiffBytes = 100 * 1024

// Diff returns the combined unified patch between the working tree at path and the comparison ref,
// i.e. the changes that would land on the ref if merged. It mirrors ChangedFiles: committed, staged,
// and untracked changes are all included.
//
// The result is git's standard unified diff format, ready for rendering. Per-file sections larger
// than maxFileDiffBytes are replaced with a "diff too large" placeholder. (Genuinely binary files
// stay small because git emits a "Binary files … differ" line for them rather than their content.)
func Diff(ctx context.Context, path, subpath, remoteName, remoteBranch string) (string, error) {
	compareBranch := remoteBranch
	if compareBranch == "" {
		branch, err := Run(ctx, path, "rev-parse", "--abbrev-ref", "HEAD")
		if err != nil {
			return "", err
		}
		compareBranch = branch
	}
	ref := fmt.Sprintf("%s/%s", remoteName, compareBranch)

	// Committed and staged changes, with rename detection (-M).
	diffArgs := []string{"diff", "-M", "--no-color", ref}
	if subpath != "" {
		diffArgs = append(diffArgs, "--", subpath)
	}
	tracked, err := Run(ctx, path, diffArgs...)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	if tracked != "" {
		b.WriteString(tracked)
		b.WriteByte('\n')
	}

	// `git diff` omits untracked files; append each one's patch (mirrors the untracked pass in
	// ChangedFiles so the diff covers the same files as the list).
	statusArgs := []string{"status", "--porcelain", "--untracked-files=all"}
	if subpath != "" {
		statusArgs = append(statusArgs, "--", subpath)
	}
	statusOut, err := Run(ctx, path, statusArgs...)
	if err != nil {
		return "", err
	}
	for line := range strings.SplitSeq(statusOut, "\n") {
		file, ok := strings.CutPrefix(line, "?? ")
		if !ok {
			continue
		}
		untracked, err := diffNoIndex(ctx, path, file)
		if err != nil {
			return "", err
		}
		if untracked != "" {
			b.WriteString(untracked)
			b.WriteByte('\n')
		}
	}

	return capLargeFileDiffs(b.String()), nil
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
	st, err := Status(ctx, path, "", remoteName, "")
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

// diffNoIndex runs `git diff --no-index` to produce a patch for an untracked file.
//
// It cannot use Run for two reasons: `git diff --no-index` exits with code 1 whenever the files
// differ (the expected case here, not an error), and Run discards stdout on any non-zero exit, so
// the patch we want would be lost. So we run the command directly, capture stdout, and treat exit
// code 1 as success.
func diffNoIndex(ctx context.Context, path, file string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", "-C", path, "diff", "--no-index", "--no-color", os.DevNull, file)
	cmd.Env = append(os.Environ(), "LC_ALL=C", "GIT_TERMINAL_PROMPT=0")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return stdout.String(), nil
		}
		return "", fmt.Errorf("git diff --no-index %s: %s (%w)", file, strings.TrimSpace(stderr.String()), err)
	}
	return stdout.String(), nil
}

// capLargeFileDiffs replaces any per-file section in a unified diff larger than maxFileDiffBytes with
// a "diff too large" placeholder, keeping every file present but bounding the total size.
func capLargeFileDiffs(diff string) string {
	if diff == "" {
		return ""
	}
	var out strings.Builder
	var section []string
	flush := func() {
		if len(section) == 0 {
			return
		}
		joined := strings.Join(section, "\n")
		if len(joined) > maxFileDiffBytes {
			out.WriteString(tooLargePlaceholder(section[0], len(joined)))
		} else {
			out.WriteString(joined)
		}
		out.WriteByte('\n')
		section = section[:0]
	}
	for line := range strings.SplitSeq(diff, "\n") {
		if strings.HasPrefix(line, "diff --git ") {
			flush()
		}
		section = append(section, line)
	}
	flush()
	return out.String()
}

// tooLargePlaceholder replaces an over-cap per-file diff section with a minimal unified-diff hunk
// whose single context line reports that the diff was elided. diff2html renders this as a neutral
// row; the "Binary files … differ" marker is avoided because diff2html always labels it "Binary
// file" regardless of the file's actual type, which would be misleading for a large text file.
// diffGitLine is the section's "diff --git a/X b/Y" header; size is the section size in bytes.
func tooLargePlaceholder(diffGitLine string, size int) string {
	a, b := "a/file", "b/file"
	if rest, ok := strings.CutPrefix(diffGitLine, "diff --git "); ok {
		if x, y, ok := strings.Cut(rest, " b/"); ok {
			a, b = x, "b/"+y
		}
	}
	return fmt.Sprintf("%s\n--- %s\n+++ %s\n@@ -1,1 +1,1 @@\n Diff too large to display (%d KB)", diffGitLine, a, b, size/1024)
}
