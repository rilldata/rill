package file

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/filewatcher"
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
func (c *connection) ListBranches(ctx context.Context) ([]drivers.GitBranch, error) {
	if !c.isGitRepo() {
		return nil, errors.New("not a git repository")
	}

	c.gitMu.Lock()
	defer c.gitMu.Unlock()

	gitPath, _, err := gitutil.InferRepoRootAndSubpath(c.root)
	if err != nil {
		return nil, err
	}

	repo, err := git.PlainOpen(gitPath)
	if err != nil {
		return nil, err
	}

	// Fetch remote branches and preview deployments if authenticated
	var gitConfig *gitutil.Config
	previewBranches := make(map[string]bool)
	if c.driverConfig.AccessToken != "" {
		gitConfig, err = c.loadGitConfig(ctx)
		if err == nil && gitConfig != nil {
			// fetch to get latest remote branches
			_ = gitutil.GitFetch(ctx, gitPath, gitConfig)

			// fetch preview branches from deployments
			depls, err := c.admin.ListDeployments(ctx, &adminv1.ListDeploymentsRequest{
				Org:     c.driverConfig.Org,
				Project: c.project.Name,
			})
			if err != nil {
				return nil, err
			}

			for _, d := range depls.Deployments {
				if !d.Editable && d.Id != c.project.PrimaryDeploymentId {
					previewBranches[d.Branch] = true
				}
			}
		}
	}

	// List all references (local and remote)
	branchSet := make(map[string]bool)
	refs, err := repo.References()
	if err != nil {
		return nil, err
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
			if branchName, ok := strings.CutPrefix(refName.Short(), gitConfig.RemoteName()+"/"); ok && branchName != "HEAD" {
				branchSet[branchName] = true
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Get current branch
	head, err := repo.Head()
	if err != nil {
		return nil, err
	}
	currentBranch := head.Name().Short()

	branches := make([]drivers.GitBranch, 0, len(branchSet))
	for name := range branchSet {
		branches = append(branches, drivers.GitBranch{
			Name:                 name,
			IsCurrent:            name == currentBranch,
			HasPreviewDeployment: previewBranches[name],
		})
	}

	return branches, nil
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

	// Check if branch exists
	branchRef := plumbing.NewBranchReferenceName(branchName)
	_, err = repo.Reference(branchRef, true)
	branchExists := err == nil

	if !branchExists {
		if !createIfNotExists {
			return git.ErrBranchNotFound
		}

		// Create new branch from HEAD
		head, err := repo.Head()
		if err != nil {
			return err
		}

		// Create the branch reference
		err = repo.CreateBranch(&config.Branch{
			Name:   branchName,
			Remote: "origin",
			Merge:  plumbing.NewBranchReferenceName(branchName),
		})
		if err != nil {
			return err
		}

		// Create the reference pointing to HEAD's commit
		ref := plumbing.NewHashReference(branchRef, head.Hash())
		err = repo.Storer.SetReference(ref)
		if err != nil {
			return err
		}
	}

	// Checkout the branch
	err = w.Checkout(&git.CheckoutOptions{
		Branch: branchRef,
	})
	if err != nil {
		return err
	}

	return nil
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
		if errors.Is(err, errProjectNotFound) || errors.Is(err, errAuthRequired) {
			// not connected to a rill project or not authenticated, return minimal status
			st, err := gitutil.RunGitStatus(gitPath, subPath, "origin")
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
func (c *connection) Pull(ctx context.Context, discardChanges, forceHandshake bool) error {
	// TODO: If its a Git repository, pull the current branch. Otherwise, this is a no-op.
	return nil
}

// Commit implements drivers.RepoStore.
func (c *connection) Commit(ctx context.Context, message string) error {
	// If its a Git repository, commit the changes with the given message to the current branch.
	if !c.isGitRepo() {
		return nil
	}

	c.gitMu.Lock()
	defer c.gitMu.Unlock()

	gitPath, subpath, err := gitutil.InferRepoRootAndSubpath(c.root)
	if err != nil {
		return err
	}

	author, err := c.gitSignature(ctx, gitPath)
	if err != nil {
		return err
	}

	_, err = gitCommitAll(gitPath, subpath, message, author)
	if err != nil {
		return err
	}

	return nil
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

	// commit existing changes if any
	author, err := c.gitSignature(ctx, gitPath)
	if err != nil {
		return "", err
	}

	_, err = gitCommitAll(gitPath, subpath, "WIP: commit before restore", author)
	if err != nil {
		return "", err
	}

	err = restoreToCommit(gitPath, subpath, commitSHA)
	if err != nil {
		return "", fmt.Errorf("failed to restore to commit %s: %w", commitSHA, err)
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

// CommitAndPush commits local changes to the remote repository and pushes them.
func (c *connection) CommitAndPush(ctx context.Context, message string, force bool) error {
	// TODO: If its a Git repository, commit and push the changes with the given message to the current branch.
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
