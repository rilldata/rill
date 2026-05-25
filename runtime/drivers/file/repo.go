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
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/filewatcher"
	rtgitutil "github.com/rilldata/rill/runtime/pkg/gitutil"
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
	if !c.isGitRepo() {
		return nil, "", errors.New("not a git repository")
	}

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
		credURL, err := cfg.FullyQualifiedRemote()
		if err != nil {
			return nil, "", err
		}
		if err := gitFetchAll(ctx, gitPath, credURL, remoteName); err != nil {
			return nil, "", err
		}
	} else {
		remotes, err := gitListRemotes(gitPath)
		if err != nil {
			return nil, "", err
		}
		for name := range remotes {
			if name == "__rill_remote" {
				remoteName = name
				break
			}
			remoteName = name
		}
	}

	localBranches, remoteBranches, err := gitListRefs(gitPath)
	if err != nil {
		return nil, "", err
	}

	branchSet := make(map[string]bool)
	for _, b := range localBranches {
		branchSet[b] = true
	}
	for _, b := range remoteBranches[remoteName] {
		branchSet[b] = true
	}

	currentBranch, err := rtgitutil.CurrentBranch(gitPath)
	if err != nil {
		return nil, "", err
	}

	return maps.Keys(branchSet), currentBranch, nil
}

// SwitchBranch implements drivers.RepoStore.
func (c *connection) SwitchBranch(ctx context.Context, branchName string, createIfNotExists, ignoreLocalChanges bool) error {
	if !c.isGitRepo() {
		return errors.New("not a git repository")
	}

	c.gitMu.Lock()
	defer c.gitMu.Unlock()

	gitPath, _, err := gitutil.InferRepoRootAndSubpath(c.root)
	if err != nil {
		return err
	}

	return rtgitutil.Checkout(gitPath, branchName, ignoreLocalChanges, createIfNotExists, "")
}

// ListCommits implements drivers.RepoStore.
func (c *connection) ListCommits(ctx context.Context, pageToken string, limit int) ([]drivers.Commit, string, error) {
	if !c.isGitRepo() {
		return nil, "", errors.New("not a git repository")
	}

	c.gitMu.Lock()
	defer c.gitMu.Unlock()

	gitPath, _, err := gitutil.InferRepoRootAndSubpath(c.root)
	if err != nil {
		return nil, "", err
	}

	// Use HEAD if no page token provided
	fromHash := pageToken
	if fromHash == "" {
		fromHash, err = gitHeadHash(gitPath)
		if err != nil {
			return nil, "", err
		}
	}

	rawCommits, err := gitLogCommits(gitPath, fromHash, limit)
	if err != nil {
		return nil, "", err
	}

	var commits []drivers.Commit
	var nextPageToken string
	for i, rc := range rawCommits {
		if limit > 0 && i >= limit {
			nextPageToken = rc.hash
			break
		}
		commits = append(commits, drivers.Commit{
			CommitSha:     rc.hash,
			AuthorName:    rc.authorName,
			AuthorEmail:   rc.authorEmail,
			CommitMessage: rc.message,
			CommittedOn:   timestamppb.New(rc.committedAt),
		})
	}

	return commits, nextPageToken, nil
}

func (c *connection) Status(ctx context.Context, remoteBranch string) (*drivers.RepoStatus, error) {
	if !c.isGitRepo() {
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
			st, err := gitutil.RunGitStatus(gitPath, subPath, "origin", remoteBranch)
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
	err = gitutil.GitFetch(ctx, gitPath, config)
	if err != nil {
		return nil, err
	}
	gs, err := gitutil.RunGitStatus(gitPath, subPath, config.RemoteName(), remoteBranch)
	if err != nil {
		return nil, err
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
	}, nil
}

// Pull implements drivers.RepoStore.
func (c *connection) Pull(ctx context.Context, opts *drivers.PullOptions) error {
	// If its a Git repository, pull the current branch. Otherwise, this is a no-op.
	if !c.isGitRepo() || !opts.UserTriggered {
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

	_, err = gitutil.RunGitPull(ctx, gitPath, opts.DiscardChanges, remote, gitConfig.RemoteName())
	if err != nil {
		return err
	}
	return nil
}

// Commit implements drivers.RepoStore.
func (c *connection) Commit(ctx context.Context, message string) (string, error) {
	// If its a Git repository, commit the changes with the given message to the current branch.
	if !c.isGitRepo() {
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
	authorName, authorEmail, err := c.gitSignature(ctx, client, gitPath)
	if err != nil {
		return "", err
	}

	hash, err := gitCommitAll(gitPath, subpath, message, authorName, authorEmail)
	if err != nil {
		return "", err
	}

	return hash, nil
}

// RestoreCommit implements drivers.RepoStore.
func (c *connection) RestoreCommit(ctx context.Context, commitSHA string) (string, error) {
	// If its a Git repository, revert the specified commit.
	if !c.isGitRepo() {
		return "", errors.New("not a git repository")
	}

	c.gitMu.Lock()
	defer c.gitMu.Unlock()

	gitPath, subpath, err := gitutil.InferRepoRootAndSubpath(c.root)
	if err != nil {
		return "", err
	}

	// check that commit exists
	if err := gitVerifyCommit(gitPath, commitSHA); err != nil {
		if errors.Is(err, errCommitNotFound) {
			return "", fmt.Errorf("commit %q not found", commitSHA)
		}
		return "", err
	}

	// commit existing changes if any
	client, err := c.getAdminClient()
	if err != nil {
		return "", err
	}
	authorName, authorEmail, err := c.gitSignature(ctx, client, gitPath)
	if err != nil {
		return "", err
	}

	_, err = gitCommitAll(gitPath, subpath, "WIP: commit before restore", authorName, authorEmail)
	if err != nil {
		return "", err
	}

	err = restoreToCommit(gitPath, subpath, commitSHA)
	if err != nil {
		return "", err
	}

	// Create the restore commit
	hash, err := gitCommitAll(gitPath, subpath, fmt.Sprintf("Restore commit %s", commitSHA[:7]), authorName, authorEmail)
	if err != nil {
		return "", err
	}
	if hash == "" {
		return "", fmt.Errorf("restore would result in no changes")
	}

	return hash, nil
}

// CommitAndPush commits local changes to the remote repository and pushes them.
func (c *connection) CommitAndPush(ctx context.Context, message string, force bool) error {
	// If its a Git repository, commit and push the changes with the given message to the current branch.
	if !c.isGitRepo() {
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
	authorName, authorEmail, err := c.gitSignature(ctx, client, gitPath)
	if err != nil {
		return err
	}

	// fetch the status
	gs, err := gitutil.RunGitStatus(gitPath, subpath, gitConfig.RemoteName(), "")
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
		err := gitutil.RunUpstreamMerge(ctx, gitConfig.RemoteName(), c.root, gitConfig.DefaultBranch, true)
		if err != nil {
			return fmt.Errorf("local is behind remote and failed to sync with remote: %w", err)
		}
		return gitutil.CommitAndPush(ctx, c.root, gitConfig, message, &object.Signature{
			Name:  authorName,
			Email: authorEmail,
			When:  time.Now(),
		})
	}
	err = gitutil.RunUpstreamMerge(ctx, gitConfig.RemoteName(), c.root, gitConfig.DefaultBranch, false)
	if err != nil {
		return fmt.Errorf("local is behind remote and failed to sync with remote: %w", err)
	}
	return gitutil.CommitAndPush(ctx, c.root, gitConfig, message, &object.Signature{
		Name:  authorName,
		Email: authorEmail,
		When:  time.Now(),
	})
}

func (c *connection) MergeToBranch(ctx context.Context, branch string, force bool) (resErr error) {
	// If its a Git repository, merge the current branch to the specified branch.
	if !c.isGitRepo() {
		return errors.New("not a git repository")
	}

	c.gitMu.Lock()
	defer c.gitMu.Unlock()

	gitPath, subpath, err := gitutil.InferRepoRootAndSubpath(c.root)
	if err != nil {
		return err
	}

	// Get the current branch
	currentBranch, err := rtgitutil.CurrentBranch(gitPath)
	if err != nil {
		return err
	}

	// Switch to the target branch
	if err := rtgitutil.Checkout(gitPath, branch, false, false, ""); err != nil {
		return err
	}
	defer func() {
		// Switch back to the original branch
		if err := rtgitutil.Checkout(gitPath, currentBranch, false, false, ""); err != nil {
			resErr = errors.Join(resErr, fmt.Errorf("failed to switch back to the original branch: %w", err))
		}
	}()

	if force {
		if subpath != "" {
			return fmt.Errorf("cannot force merge in a monorepo setup")
		}
		return rtgitutil.MergeWithStrategy(gitPath, branch, "theirs")
	}
	merged, err := rtgitutil.MergeWithBailOnConflict(gitPath, branch)
	if err != nil {
		return err
	}
	if !merged {
		return &drivers.MergeFailedError{
			Output:       "merge failed due to conflicts, use force merge to favour current changes",
			MergedBranch: branch,
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

func (c *connection) isGitRepo() bool {
	return isGitDir(c.root)
}

func gitCommitAll(path, subpath, message, authorName, authorEmail string) (string, error) {
	if message == "" {
		message = "Auto committed by Rill"
	}
	var glob string
	if subpath != "" {
		glob = filepath.Join(subpath, "**")
	}
	hash, err := rtgitutil.CommitAll(path, glob, message, authorName, authorEmail)
	if err != nil {
		return "", fmt.Errorf("failed to commit files to git: %w", err)
	}
	return hash, nil
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

// isGitDir reports whether path is inside a git repository.
func isGitDir(path string) bool {
	return exec.Command("git", "-C", path, "rev-parse", "--git-dir").Run() == nil
}

// gitFetchAll fetches all branches from credURL into refs/remotes/<remoteName>/*.
func gitFetchAll(ctx context.Context, repoDir, credURL, remoteName string) error {
	refSpec := fmt.Sprintf("+refs/heads/*:refs/remotes/%s/*", remoteName)
	out, err := exec.CommandContext(ctx, "git", "-C", repoDir, "fetch", "--force", credURL, refSpec).CombinedOutput()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			return fmt.Errorf("git fetch failed: %s", string(out))
		}
		return err
	}
	return nil
}

// gitListRemotes returns a map of remote name → URL for the repository.
func gitListRemotes(repoDir string) (map[string]string, error) {
	out, err := exec.Command("git", "-C", repoDir, "remote", "-v").Output()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			return nil, fmt.Errorf("git remote failed: %s", string(execErr.Stderr))
		}
		return nil, err
	}
	remotes := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			remotes[parts[0]] = parts[1]
		}
	}
	return remotes, nil
}

// gitListRefs returns local branch names and remote branch names keyed by remote name.
func gitListRefs(repoDir string) (localBranches []string, remoteBranches map[string][]string, err error) {
	out, err := exec.Command("git", "-C", repoDir, "for-each-ref", "--format=%(refname)", "refs/heads", "refs/remotes").Output()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			return nil, nil, fmt.Errorf("git for-each-ref failed: %s", string(execErr.Stderr))
		}
		return nil, nil, err
	}
	remoteBranches = make(map[string][]string)
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if after, ok := strings.CutPrefix(line, "refs/heads/"); ok {
			localBranches = append(localBranches, after)
		} else if after, ok := strings.CutPrefix(line, "refs/remotes/"); ok {
			parts := strings.SplitN(after, "/", 2)
			if len(parts) == 2 && parts[1] != "HEAD" {
				remoteBranches[parts[0]] = append(remoteBranches[parts[0]], parts[1])
			}
		}
	}
	return localBranches, remoteBranches, nil
}

// gitHeadHash returns the commit hash of HEAD.
func gitHeadHash(repoDir string) (string, error) {
	out, err := exec.Command("git", "-C", repoDir, "rev-parse", "HEAD").Output()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			if strings.Contains(string(execErr.Stderr), "unknown revision") {
				return "", nil
			}
			return "", fmt.Errorf("git rev-parse failed: %s", string(execErr.Stderr))
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// commitInfo holds data for a single git commit.
type commitInfo struct {
	hash        string
	authorName  string
	authorEmail string
	message     string
	committedAt time.Time
}

// gitLogCommits returns up to limit+1 commits starting from fromHash (for pagination).
func gitLogCommits(repoDir, fromHash string, limit int) ([]commitInfo, error) {
	args := []string{"-C", repoDir, "log", "--format=%H%x00%an%x00%ae%x00%cI%x00%s"}
	if fromHash != "" {
		args = append(args, fromHash)
	}
	if limit > 0 {
		args = append(args, fmt.Sprintf("-n%d", limit+1))
	}
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			return nil, fmt.Errorf("git log failed: %s", string(execErr.Stderr))
		}
		return nil, err
	}
	var commits []commitInfo
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\x00", 5)
		if len(parts) < 5 {
			continue
		}
		committedAt, _ := time.Parse(time.RFC3339, parts[3])
		commits = append(commits, commitInfo{
			hash:        parts[0],
			authorName:  parts[1],
			authorEmail: parts[2],
			message:     parts[4],
			committedAt: committedAt,
		})
	}
	return commits, nil
}

var errCommitNotFound = errors.New("commit not found")

// gitVerifyCommit checks that a commit hash exists in the repository.
func gitVerifyCommit(repoDir, sha string) error {
	if err := exec.Command("git", "-C", repoDir, "cat-file", "-e", sha+"^{commit}").Run(); err != nil {
		return errCommitNotFound
	}
	return nil
}
