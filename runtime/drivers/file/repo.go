package file

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/filewatcher"
	"github.com/rilldata/rill/runtime/pkg/gitutil"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Root implements drivers.RepoStore.
func (c *connection) Root(ctx context.Context) (string, error) {
	return c.root, nil
}

// ListGlob implements drivers.RepoStore.
func (c *connection) ListGlob(ctx context.Context, glob string, skipDirs bool) ([]drivers.DirEntry, error) {
	// Check that folder hasn't been moved
	if err := c.checkRoot(); err != nil {
		return nil, err
	}

	fsRoot := os.DirFS(c.root)
	glob = filepath.Clean(filepath.Join(".", glob))

	var entries []drivers.DirEntry
	err := doublestar.GlobWalk(fsRoot, glob, func(p string, d fs.DirEntry) error {
		if skipDirs && d.IsDir() {
			return nil
		}

		// Exit if we reached the limit
		if len(entries) == drivers.RepoListLimit {
			return drivers.ErrRepoListLimitExceeded
		}

		// Track file (p is already relative to the FS root)
		p = filepath.Join(string(filepath.Separator), p)
		// Do not send files for ignored paths
		if drivers.IsIgnored(p, c.ignorePaths) {
			return nil
		}
		entries = append(entries, drivers.DirEntry{
			Path:  p,
			IsDir: d.IsDir(),
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// Get implements drivers.RepoStore.
func (c *connection) Get(ctx context.Context, filePath string) (string, error) {
	fp := filepath.Join(c.root, filePath)

	b, err := os.ReadFile(fp)
	if err != nil {
		// obscure the root directory location
		if perr, ok := err.(*fs.PathError); ok { // nolint:errorlint // we specifically check for a non-wrapped error
			perr.Path = filePath
			return "", perr
		}
		return "", err
	}

	return string(b), nil
}

// Hash implements drivers.RepoStore.
func (c *connection) Hash(ctx context.Context, paths []string) (string, error) {
	hasher := md5.New()
	for _, path := range paths {
		path = filepath.Join(c.root, path)
		file, err := os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", err
		}

		if _, err := io.Copy(hasher, file); err != nil {
			file.Close()
			return "", err
		}
		file.Close()
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// Stat implements drivers.RepoStore.
func (c *connection) Stat(ctx context.Context, filePath string) (*drivers.FileInfo, error) {
	filePath = filepath.Join(c.root, filePath)

	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	return &drivers.FileInfo{
		LastUpdated: info.ModTime(),
		IsDir:       info.IsDir(),
	}, nil
}

// Put implements drivers.RepoStore.
func (c *connection) Put(ctx context.Context, filePath string, reader io.Reader) error {
	filePath = filepath.Join(c.root, filePath)

	err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return err
	}

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, reader)
	if err != nil {
		return err
	}

	return nil
}

// MkdirAll implements drivers.RepoStore.
func (c *connection) MkdirAll(ctx context.Context, dirPath string) error {
	dirPath = filepath.Join(c.root, dirPath)

	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

// Rename implements drivers.RepoStore.
func (c *connection) Rename(ctx context.Context, fromPath, toPath string) error {
	toPath = filepath.Join(c.root, toPath)

	fromPath = filepath.Join(c.root, fromPath)
	if _, err := os.Stat(toPath); !strings.EqualFold(fromPath, toPath) && err == nil {
		return os.ErrExist
	}
	err := os.Rename(fromPath, toPath)
	if err != nil {
		return err
	}
	return os.Chtimes(toPath, time.Now(), time.Now())
}

// Delete implements drivers.RepoStore.
func (c *connection) Delete(ctx context.Context, filePath string, force bool) error {
	filePath = filepath.Join(c.root, filePath)
	if force {
		return os.RemoveAll(filePath)
	}
	return os.Remove(filePath)
}

// Watch implements drivers.RepoStore.
func (c *connection) Watch(ctx context.Context, cb drivers.WatchCallback) error {
	return c.watcher.Subscribe(ctx, func(events []filewatcher.WatchEvent) {
		if len(events) == 0 {
			return
		}
		watchEvents := make([]drivers.WatchEvent, 0, len(events))
		for _, e := range events {
			watchEvents = append(watchEvents, drivers.WatchEvent{
				Type: e.Type,
				Path: e.RelPath,
				Dir:  e.Dir,
			})
		}
		cb(watchEvents)
	})
}

// ListBranches implements drivers.RepoStore.
func (c *connection) ListBranches(ctx context.Context) ([]string, string, error) {
	c.gitMu.Lock()
	defer c.gitMu.Unlock()

	gitPath, _, err := gitutil.InferRepoRootAndSubpath(c.root)
	if err != nil {
		return nil, "", err
	}

	cfg, err := c.loadGitConfig(ctx)
	if err != nil && !errors.Is(err, errProjectNotFound) && !errors.Is(err, drivers.ErrNotAuthenticated) {
		return nil, "", err
	}
	var remoteName string
	if cfg != nil {
		remoteName = cfg.RemoteName()
		// Fetch all branches from the remote using the credential-embedded URL.
		remoteURL, err := cfg.FullyQualifiedRemote()
		if err != nil {
			return nil, "", err
		}
		refspec := "+refs/heads/*:refs/remotes/" + remoteName + "/*"
		if _, err := gitutil.Run(ctx, gitPath, "fetch", remoteURL, refspec); err != nil {
			return nil, "", err
		}
	} else {
		// No admin config: pick a remote name from the local config. Prefer "__rill_remote" if present.
		out, err := gitutil.Run(ctx, gitPath, "remote")
		if err != nil {
			return nil, "", err
		}
		for r := range strings.SplitSeq(out, "\n") {
			r = strings.TrimSpace(r)
			if r == "" {
				continue
			}
			if r == "__rill_remote" {
				remoteName = r
				break
			}
			remoteName = r
		}
	}

	// List all local and remote branch refs.
	refsOut, err := gitutil.Run(ctx, gitPath, "for-each-ref", "--format=%(refname)", "refs/heads/", "refs/remotes/")
	if err != nil {
		return nil, "", err
	}

	branchSet := make(map[string]bool)
	for line := range strings.SplitSeq(refsOut, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if local, ok := strings.CutPrefix(line, "refs/heads/"); ok {
			branchSet[local] = true
			continue
		}
		if remote, ok := strings.CutPrefix(line, "refs/remotes/"); ok {
			// Strip "<remoteName>/" prefix; skip the symbolic HEAD ref.
			if branchName, ok := strings.CutPrefix(remote, remoteName+"/"); ok && branchName != "HEAD" {
				branchSet[branchName] = true
			}
		}
	}

	currentBranch, err := gitutil.Run(ctx, gitPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return nil, "", err
	}

	return maps.Keys(branchSet), currentBranch, nil
}

// SwitchBranch implements drivers.RepoStore.
func (c *connection) SwitchBranch(ctx context.Context, branchName string, createIfNotExists, ignoreLocalChanges bool) error {
	c.gitMu.Lock()
	defer c.gitMu.Unlock()

	gitPath, _, err := gitutil.InferRepoRootAndSubpath(c.root)
	if err != nil {
		return err
	}

	return gitutil.Checkout(gitPath, branchName, ignoreLocalChanges, createIfNotExists, "")
}

// ListCommits implements drivers.RepoStore.
func (c *connection) ListCommits(ctx context.Context, pageToken string, limit int) ([]drivers.Commit, string, error) {
	c.gitMu.Lock()
	defer c.gitMu.Unlock()

	gitPath, _, err := gitutil.InferRepoRootAndSubpath(c.root)
	if err != nil {
		return nil, "", err
	}

	// Use `-z` so git separates commits with NUL bytes, which cannot appear in commit data.
	// Fields within a commit are separated by the ASCII unit separator (\x1f).
	const fieldSep = "\x1f"
	format := "--format=%H" + fieldSep + "%an" + fieldSep + "%ae" + fieldSep + "%cI" + fieldSep + "%B"

	args := []string{"log", "-z", format}
	if limit > 0 {
		// Fetch one extra commit so we can populate the next page token.
		args = append(args, "-n", strconv.Itoa(limit+1))
	}
	if pageToken != "" {
		// Validate before passing to git: an arbitrary string could be interpreted as a flag.
		if !gitutil.IsCommitHash(pageToken) {
			return nil, "", fmt.Errorf("invalid page token %q", pageToken)
		}
		args = append(args, pageToken)
	}

	out, err := gitutil.Run(ctx, gitPath, args...)
	if err != nil {
		return nil, "", err
	}

	var commits []drivers.Commit
	var nextPageToken string
	for rec := range strings.SplitSeq(out, "\x00") {
		rec = strings.TrimLeft(rec, "\n")
		if rec == "" {
			continue
		}
		fields := strings.SplitN(rec, fieldSep, 5)
		if len(fields) < 5 {
			return nil, "", fmt.Errorf("unexpected git log output: %q", rec)
		}
		if limit > 0 && len(commits) >= limit {
			nextPageToken = fields[0]
			break
		}
		t, err := time.Parse(time.RFC3339, fields[3])
		if err != nil {
			return nil, "", fmt.Errorf("failed to parse commit date %q: %w", fields[3], err)
		}
		commits = append(commits, drivers.Commit{
			CommitSha:     fields[0],
			AuthorName:    fields[1],
			AuthorEmail:   fields[2],
			CommitMessage: fields[4],
			CommittedOn:   timestamppb.New(t),
		})
	}

	return commits, nextPageToken, nil
}

func (c *connection) Status(ctx context.Context, remoteBranch string, opts drivers.RepoStatusOptions) (*drivers.RepoStatus, error) {
	if !gitutil.IsGitRepo(c.root) {
		return &drivers.RepoStatus{}, nil
	}

	c.gitMu.Lock()
	defer c.gitMu.Unlock()

	gitPath, subPath, err := gitutil.InferRepoRootAndSubpath(c.root)
	if err != nil {
		// should not happen because we already checked isGitRepo
		return nil, err
	}

	// get ephemeral git credentials
	config, err := c.loadGitConfig(ctx)
	if err != nil {
		if errors.Is(err, errProjectNotFound) || errors.Is(err, drivers.ErrNotAuthenticated) {
			// not connected to a rill project or not authenticated, return minimal status
			st, err := gitutil.Status(ctx, gitPath, subPath, "origin", remoteBranch)
			if err != nil {
				return nil, err
			}
			return &drivers.RepoStatus{
				IsGitRepo: true,
				Branch:    st.Branch,
				RemoteURL: st.RemoteURL,
				Subpath:   subPath,
			}, nil
		}
		return nil, err
	}

	// set remote
	err = gitutil.Fetch(ctx, gitPath, config)
	if err != nil {
		return nil, err
	}
	gs, err := gitutil.Status(ctx, gitPath, subPath, config.RemoteName(), remoteBranch)
	if err != nil {
		return nil, err
	}
	// Listing changed files (and the diff) is extra git work most callers do not need, so it is opt-in
	// and computed separately.
	var files []gitutil.ChangedFile
	if opts.ChangedFiles || opts.Diff {
		f, err := gitutil.ChangedFiles(ctx, gitPath, subPath, config.RemoteName(), remoteBranch)
		if err != nil {
			return nil, err
		}
		files = f
	}
	var diff string
	if opts.Diff {
		d, err := gitutil.Diff(ctx, gitPath, subPath, config.RemoteName(), remoteBranch)
		if err != nil {
			return nil, err
		}
		diff = d
	}
	return &drivers.RepoStatus{
		IsGitRepo:     true,
		Branch:        gs.Branch,
		RemoteURL:     gs.RemoteURL,
		Subpath:       subPath,
		ManagedRepo:   config.ManagedRepo,
		LocalChanges:  gs.LocalChanges,
		LocalCommits:  gs.LocalCommits,
		RemoteCommits: gs.RemoteCommits,
		ChangedFiles:  repoFileChanges(files),
		Diff:          diff,
	}, nil
}

// Pull implements drivers.RepoStore.
func (c *connection) Pull(ctx context.Context, opts *drivers.PullOptions) error {
	// If its a Git repository, pull the current branch. Otherwise, this is a no-op.
	if !gitutil.IsGitRepo(c.root) || !opts.UserTriggered {
		return nil
	}

	c.gitMu.Lock()
	defer c.gitMu.Unlock()

	gitPath, subpath, err := gitutil.InferRepoRootAndSubpath(c.root)
	if err != nil {
		// Not a git repo
		return err
	}

	gitConfig, err := c.loadGitConfig(ctx)
	if err != nil {
		return err
	}

	if gitConfig.Subpath != subpath {
		// should not happen
		return errors.New("detected subpath within git repo does not match project subpath")
	}

	remote, err := gitConfig.FullyQualifiedRemote()
	if err != nil {
		return err
	}

	_, err = gitutil.Pull(ctx, gitPath, opts.DiscardChanges, remote, gitConfig.RemoteName())
	if err != nil {
		return err
	}
	return nil
}

// Commit implements drivers.RepoStore.
func (c *connection) Commit(ctx context.Context, message string) (string, error) {
	// If its a Git repository, commit the changes with the given message to the current branch.
	if !gitutil.IsGitRepo(c.root) {
		return "", errors.New("not a git repository")
	}

	c.gitMu.Lock()
	defer c.gitMu.Unlock()

	gitPath, subpath, err := gitutil.InferRepoRootAndSubpath(c.root)
	if err != nil {
		return "", err
	}

	client, err := c.getAdminClient()
	if err != nil && !errors.Is(err, drivers.ErrNotAuthenticated) { // allow committing without auth
		return "", err
	}
	author, err := c.gitSignature(ctx, client, gitPath)
	if err != nil {
		return "", err
	}

	if message == "" {
		message = "Auto committed by Rill"
	}
	hash, err := gitutil.CommitAll(ctx, gitPath, subpath, message, author)
	if err != nil {
		if errors.Is(err, gitutil.ErrEmptyCommit) {
			// Nothing to commit - preserve the historical contract of returning (empty hash, no error).
			return "", nil
		}
		return "", err
	}

	return hash, nil
}

// RestoreCommit implements drivers.RepoStore.
func (c *connection) RestoreCommit(ctx context.Context, commitSHA string) (string, error) {
	// If its a Git repository, revert the specified commit.
	if !gitutil.IsGitRepo(c.root) {
		return "", errors.New("not a git repository")
	}

	c.gitMu.Lock()
	defer c.gitMu.Unlock()

	gitPath, subpath, err := gitutil.InferRepoRootAndSubpath(c.root)
	if err != nil {
		return "", err
	}

	// Require a full hash: prevents git interpreting the value as a flag and guarantees commitSHA[:7] below is safe.
	if !gitutil.IsCommitHash(commitSHA) {
		return "", fmt.Errorf("invalid commit SHA %q: must be a full commit hash", commitSHA)
	}

	// check that the commit exists and is actually a commit (not a tree/blob/tag).
	if _, err := gitutil.Run(ctx, gitPath, "cat-file", "-e", commitSHA+"^{commit}"); err != nil {
		return "", fmt.Errorf("commit %q not found", commitSHA)
	}

	// commit existing changes if any
	client, err := c.getAdminClient()
	if err != nil {
		return "", err
	}
	author, err := c.gitSignature(ctx, client, gitPath)
	if err != nil {
		return "", err
	}

	_, err = gitutil.CommitAll(ctx, gitPath, subpath, "WIP: commit before restore", author)
	if err != nil && !errors.Is(err, gitutil.ErrEmptyCommit) {
		return "", err
	}

	err = restoreToCommit(gitPath, subpath, commitSHA)
	if err != nil {
		return "", err
	}

	// Create the restore commit
	hash, err := gitutil.CommitAll(ctx, gitPath, subpath, fmt.Sprintf("Restore commit %s", commitSHA[:7]), author)
	if err != nil {
		if errors.Is(err, gitutil.ErrEmptyCommit) {
			return "", fmt.Errorf("restore would result in no changes")
		}
		return "", fmt.Errorf("failed to commit restore: %w", err)
	}

	return hash, nil
}

// CommitAndPush commits local changes to the remote repository and pushes them.
func (c *connection) CommitAndPush(ctx context.Context, message string, force bool) error {
	// If its a Git repository, commit and push the changes with the given message to the current branch.
	if !gitutil.IsGitRepo(c.root) {
		return errors.New("not a git repository")
	}

	c.gitMu.Lock()
	defer c.gitMu.Unlock()

	gitPath, subpath, err := gitutil.InferRepoRootAndSubpath(c.root)
	if err != nil {
		// Not a git repo - checked above
		return err
	}

	gitConfig, err := c.loadGitConfig(ctx)
	if err != nil {
		return err
	}

	if gitConfig.Subpath != subpath {
		// should not happen
		return errors.New("detected subpath within git repo does not match project subpath")
	}

	client, err := c.getAdminClient()
	if err != nil {
		return err
	}
	author, err := c.gitSignature(ctx, client, gitPath)
	if err != nil {
		return err
	}

	// fetch the status
	gs, err := gitutil.Status(ctx, gitPath, subpath, gitConfig.RemoteName(), "")
	if err != nil {
		return err
	}
	if gs.RemoteCommits > 0 && !force {
		return drivers.ErrRemoteAhead
	}

	if force {
		// Instead of a force push, we do a merge with favourLocal=true to ensure we don't lose history.
		// This is not equivalent to a force push but is safer for users.
		if gitConfig.Subpath != "" {
			// force pushing in a monorepo can overwrite other subpaths
			// we can check for changes in other subpaths but it is tricky and error prone
			// monorepo setups are advanced use cases and we can require users to manually resolve remote changes
			return fmt.Errorf("cannot overwrite remote changes in a monorepo setup. Merge remote changes manually")
		}
		err := gitutil.UpstreamMerge(ctx, c.root, gitConfig.RemoteName(), gitConfig.DefaultBranch, true)
		if err != nil {
			return fmt.Errorf("local is behind remote and failed to sync with remote: %w", err)
		}
		return gitutil.CommitAndPush(ctx, c.root, gitConfig, message, author)
	}
	err = gitutil.UpstreamMerge(ctx, c.root, gitConfig.RemoteName(), gitConfig.DefaultBranch, false)
	if err != nil {
		return fmt.Errorf("local is behind remote and failed to sync with remote: %w", err)
	}
	return gitutil.CommitAndPush(ctx, c.root, gitConfig, message, author)
}

func (c *connection) MergeToBranch(ctx context.Context, branch string, force bool) (resErr error) {
	// If its a Git repository, merge the current branch to the specified branch.
	if !gitutil.IsGitRepo(c.root) {
		return errors.New("not a git repository")
	}

	c.gitMu.Lock()
	defer c.gitMu.Unlock()

	gitPath, subpath, err := gitutil.InferRepoRootAndSubpath(c.root)
	if err != nil {
		return err
	}

	// Remember the current branch so we can restore it on return.
	currentBranch, err := gitutil.Run(ctx, gitPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return err
	}

	// Switch to the target branch.
	if err := gitutil.Checkout(gitPath, branch, false, false, ""); err != nil {
		return err
	}
	defer func() {
		if err := gitutil.Checkout(gitPath, currentBranch, false, false, ""); err != nil {
			resErr = errors.Join(resErr, fmt.Errorf("failed to switch back to the original branch: %w", err))
		}
	}()

	if force {
		if subpath != "" {
			return fmt.Errorf("cannot force merge in a monorepo setup")
		}
		return gitutil.MergeWithStrategy(gitPath, currentBranch, "theirs")
	}
	merged, err := gitutil.MergeWithBailOnConflict(gitPath, currentBranch)
	if err != nil {
		return err
	}
	if !merged {
		return &drivers.MergeFailedError{
			Output:       "merge failed due to conflicts, use force merge to favour current changes",
			MergedBranch: currentBranch,
			Conflict:     true,
		}
	}
	return nil
}

// CommitHash implements drivers.RepoStore.
func (c *connection) CommitHash(ctx context.Context) (string, error) {
	return "", nil
}

// CommitTimestamp implements drivers.RepoStore.
func (c *connection) CommitTimestamp(ctx context.Context) (time.Time, error) {
	return time.Time{}, nil
}

func restoreToCommit(path, subpath, commithash string) error {
	var args []string
	args = append(args, "-C", path, "restore", "--source", commithash, "--staged", "--worktree")
	if subpath != "" {
		args = append(args, "--", subpath)
	} else {
		args = append(args, "--", ".")
	}
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restore to commit: %s, %w", string(output), err)
	}
	return nil
}

func repoFileChanges(files []gitutil.ChangedFile) []drivers.RepoFileChange {
	if len(files) == 0 {
		return nil
	}
	out := make([]drivers.RepoFileChange, len(files))
	for i, f := range files {
		out[i] = drivers.RepoFileChange{
			Path:    f.Path,
			OldPath: f.OldPath,
			Status:  repoFileStatus(f.Status),
		}
	}
	return out
}

func repoFileStatus(s gitutil.ChangedFileStatus) drivers.RepoFileStatus {
	switch s {
	case gitutil.ChangedFileStatusAdded:
		return drivers.RepoFileStatusAdded
	case gitutil.ChangedFileStatusModified:
		return drivers.RepoFileStatusModified
	case gitutil.ChangedFileStatusDeleted:
		return drivers.RepoFileStatusDeleted
	case gitutil.ChangedFileStatusRenamed:
		return drivers.RepoFileStatusRenamed
	default:
		return drivers.RepoFileStatusUnspecified
	}
}
