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
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
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

	repo, err := git.PlainOpen(gitPath)
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
		// fetch all branches
		err := repo.FetchContext(ctx, &git.FetchOptions{
			RemoteName: cfg.RemoteName(),
			RemoteURL:  cfg.Remote,
			RefSpecs:   []config.RefSpec{config.RefSpec("+refs/heads/*:refs/remotes/" + cfg.RemoteName() + "/*")},
			Auth: &http.BasicAuth{
				Username: cfg.Username,
				Password: cfg.Password,
			},
		})
		if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return nil, "", err
		}
	} else {
		// List all remotes
		remotes, err := repo.Remotes()
		if err != nil {
			return nil, "", err
		}

		for _, r := range remotes {
			if r.Config().Name == "__rill_remote" {
				remoteName = r.Config().Name
				break
			}
			remoteName = r.Config().Name
		}
	}

	// List all references (local and remote)
	branchSet := make(map[string]bool)
	refs, err := repo.References()
	if err != nil {
		return nil, "", err
	}

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		refName := ref.Name()
		// Include local branches (refs/heads/*)
		if refName.IsBranch() {
			branchSet[refName.Short()] = true
		}
		// Include remote branches (refs/remotes/origin/*)
		if refName.IsRemote() {
			// Strip "<remote>/" prefix to get branch name
			// Skip HEAD reference
			if branchName, ok := strings.CutPrefix(refName.Short(), remoteName+"/"); ok && branchName != "HEAD" {
				branchSet[branchName] = true
			}
		}
		return nil
	})
	if err != nil {
		return nil, "", err
	}

	// Get current branch
	head, err := repo.Head()
	if err != nil {
		return nil, "", err
	}
	currentBranch := head.Name().Short()

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

	repo, err := git.PlainOpen(gitPath)
	if err != nil {
		return err
	}

	// Get the worktree
	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
		Create: createIfNotExists,
		Force:  ignoreLocalChanges,
	})
	if err != nil {
		return err
	}

	return nil
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

	repo, err := git.PlainOpen(gitPath)
	if err != nil {
		return nil, "", err
	}

	// Determine starting point: page token or HEAD
	var fromHash plumbing.Hash
	if pageToken != "" {
		fromHash = plumbing.NewHash(pageToken)
	} else {
		head, err := repo.Head()
		if err != nil {
			return nil, "", err
		}
		fromHash = head.Hash()
	}

	// Get commit iterator starting from the determined hash
	commitIter, err := repo.Log(&git.LogOptions{
		From:  fromHash,
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, "", err
	}
	defer commitIter.Close()

	var commits []drivers.Commit
	var nextPageToken string
	for {
		if limit > 0 && len(commits) >= limit {
			// Peek next commit to get the next page token
			nextCommit, err := commitIter.Next()
			if err == nil {
				nextPageToken = nextCommit.Hash.String()
			}
			break
		}
		commit, err := commitIter.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, "", err
		}
		commits = append(commits, drivers.Commit{
			CommitSha:     commit.Hash.String(),
			AuthorName:    commit.Author.Name,
			AuthorEmail:   commit.Author.Email,
			CommitMessage: commit.Message,
			CommittedOn:   timestamppb.New(commit.Committer.When),
		})
	}

	return commits, nextPageToken, nil
}

func (c *connection) Status(ctx context.Context) (*drivers.RepoStatus, error) {
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
			st, err := gitutil.RunGitStatus(gitPath, subPath, "origin")
			if err != nil {
				return nil, err
			}
			return &drivers.RepoStatus{
				IsGitRepo:    true,
				Branch:       st.Branch,
				RemoteURL:    st.RemoteURL,
				Subpath:      subPath,
				LocalChanges: st.LocalChanges,
			}, nil
		}
		return nil, err
	}

	// set remote
	err = gitutil.GitFetch(ctx, gitPath, config)
	if err != nil {
		return nil, err
	}
	gs, err := gitutil.RunGitStatus(gitPath, subPath, config.RemoteName())
	if err != nil {
		return nil, err
	}
	return &drivers.RepoStatus{
		IsGitRepo:     true,
		Branch:        gs.Branch,
		RemoteURL:     gs.RemoteURL,
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
	author, err := c.gitSignature(ctx, client, gitPath)
	if err != nil {
		return "", err
	}

	hash, err := gitCommitAll(gitPath, subpath, message, author)
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
	repo, err := git.PlainOpen(gitPath)
	if err != nil {
		return "", err
	}
	_, err = repo.CommitObject(plumbing.NewHash(commitSHA))
	if err != nil {
		if errors.Is(err, plumbing.ErrObjectNotFound) {
			return "", fmt.Errorf("commit %q not found", commitSHA)
		}
		return "", err
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

	_, err = gitCommitAll(gitPath, subpath, "WIP: commit before restore", author)
	if err != nil {
		return "", err
	}

	err = restoreToCommit(gitPath, subpath, commitSHA)
	if err != nil {
		return "", err
	}

	// Create the restore commit
	hash, err := gitCommitAll(gitPath, subpath, fmt.Sprintf("Restore commit %s", commitSHA[:7]), author)
	if err != nil {
		if errors.Is(err, git.ErrEmptyCommit) {
			return "", fmt.Errorf("restore would result in no changes")
		}
		return "", fmt.Errorf("failed to commit restore: %w", err)
	}

	return hash, nil
}

func (c *connection) ApplyPatch(ctx context.Context, patch string) error {
	// If its a Git repository, revert the specified commit.
	if !c.isGitRepo() {
		return errors.New("not a git repository")
	}

	c.gitMu.Lock()
	defer c.gitMu.Unlock()

	gitPath, subpath, err := gitutil.InferRepoRootAndSubpath(c.root)
	if err != nil {
		return err
	}

	return applyPatch(ctx, gitPath, subpath, patch)
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
	author, err := c.gitSignature(ctx, client, gitPath)
	if err != nil {
		return err
	}

	// fetch the status
	gs, err := gitutil.RunGitStatus(gitPath, subpath, gitConfig.RemoteName())
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
		return gitutil.CommitAndPush(ctx, c.root, gitConfig, message, author)
	}
	err = gitutil.RunUpstreamMerge(ctx, gitConfig.RemoteName(), c.root, gitConfig.DefaultBranch, false)
	if err != nil {
		return fmt.Errorf("local is behind remote and failed to sync with remote: %w", err)
	}
	return gitutil.CommitAndPush(ctx, c.root, gitConfig, message, author)
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

	repo, err := git.PlainOpen(gitPath)
	if err != nil {
		return err
	}

	// Get the current branch
	head, err := repo.Head()
	if err != nil {
		return err
	}
	currentBranch := head.Name().Short()

	// Switch to the target branch
	w, err := repo.Worktree()
	if err != nil {
		return err
	}
	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
	})
	if err != nil {
		return err
	}
	defer func() {
		// Switch back to the original branch
		err := w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(currentBranch),
		})
		if err != nil {
			resErr = errors.Join(resErr, fmt.Errorf("failed to switch back to the original branch: %w", err))
		}
	}()

	if force {
		if subpath != "" {
			return fmt.Errorf("cannot force merge in a monorepo setup")
		}
		return rtgitutil.MergeWithTheirsStrategy(gitPath, branch)
	}
	aborted, err := rtgitutil.MergeWithBailOnConflict(gitPath, branch)
	if err != nil {
		return err
	}
	if aborted {
		return fmt.Errorf("merge conflicts detected while merging to branch %s", branch)
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
	_, err := git.PlainOpen(c.root)
	return err == nil
}

func gitCommitAll(path, subpath, message string, author *object.Signature) (string, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return "", err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	// Stage all changes (git add -A for the subpath)
	var stagingPath string
	if subpath != "" {
		stagingPath = filepath.Join(subpath, "**")
	} else {
		stagingPath = "."
	}
	if err := wt.AddWithOptions(&git.AddOptions{Glob: stagingPath}); err != nil {
		return "", fmt.Errorf("failed to add files to git: %w", err)
	}

	// Commit the changes (git commit -m)
	if message == "" {
		message = "Auto committed by Rill"
	}
	hash, err := wt.Commit(message, &git.CommitOptions{Author: author, AllowEmptyCommits: false})
	if err != nil {
		if !errors.Is(err, git.ErrEmptyCommit) {
			return "", fmt.Errorf("failed to commit files to git: %w", err)
		}
		// empty commit - nothing to commit
		return "", nil
	}
	return hash.String(), nil
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

func applyPatch(ctx context.Context, path, subpath, patch string) error {
	patchFile, _, err := fileutil.CopyToTempFile(strings.NewReader(patch), "rill-patch", ".patch")
	if err != nil {
		return err
	}
	defer os.Remove(patchFile)

	cmd := exec.CommandContext(ctx, "git", "-C", path, "apply", "-R", patchFile)
	_, err = cmd.CombinedOutput()
	return err
}
