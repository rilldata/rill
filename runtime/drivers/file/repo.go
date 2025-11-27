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
	"path/filepath"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
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
		if t, ok := err.(*fs.PathError); ok { // nolint:errorlint // we specifically check for a non-wrapped error
			return "", fmt.Errorf("%s %s %s", t.Op, filePath, t.Err.Error())
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

	// Get authenticated admin client
	if c.driverConfig.AccessToken == "" {
		// if not authenticated do not return local/remote changes info
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

	// get ephemeral git credentials
	config, err := c.loadGitConfig(ctx)
	if err != nil {
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

	if c.driverConfig.AccessToken == "" {
		return errors.New("must authenticate before performing this action")
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

// CommitAndPush commits local changes to the remote repository and pushes them.
func (c *connection) CommitAndPush(ctx context.Context, message string, force bool) error {
	// If its a Git repository, commit and push the changes with the given message to the current branch.
	if !c.isGitRepo() {
		return nil
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

	author, err := c.gitSignature(ctx, gitPath)
	if err != nil {
		return err
	}

	// fetch the status
	gs, err := gitutil.RunGitStatus(gitPath, subpath, gitConfig.RemoteName())
	if err != nil {
		return err
	}
	if gs.RemoteCommits > 0 && !force {
		return errors.New("cannot push with remote commits present, please pull first")
	}

	if force {
		// Instead of a force push, we do a merge with favourLocal=true to ensure we don't loose history.
		// This is not euivalent to a force push but is safer for users.
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

// loadGitConfig loads the git configuration for the repository
// Should be called with c.gitMu held.
func (c *connection) loadGitConfig(ctx context.Context) (*gitutil.Config, error) {
	if c.gitConfig != nil && !c.gitConfig.IsExpired() {
		return c.gitConfig, nil
	}

	// Build request
	req := &adminv1.ListProjectsForFingerprintRequest{
		DirectoryName: filepath.Base(c.root),
	}

	// extract subpath
	repoRoot, subpath, err := gitutil.InferRepoRootAndSubpath(c.root)
	if err == nil {
		req.SubPath = subpath
	}

	// extract remotes
	remote, err := gitutil.ExtractRemotes(repoRoot, false)
	if err == nil {
		for _, r := range remote {
			if r.Name == "__rill_remote" {
				req.RillMgdGitRemote = r.URL
			} else {
				gitRemote, err := r.Github()
				if err == nil {
					req.GitRemote = gitRemote
				}
			}
		}
	}
	resp, err := c.admin.ListProjectsForFingerprint(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(resp.Projects) == 0 {
		return nil, nil
	}

	orgFiltered := make([]*adminv1.Project, 0)
	for _, p := range resp.Projects {
		if p.OrgName == c.driverConfig.Org {
			orgFiltered = append(orgFiltered, p)
		}
	}
	if len(orgFiltered) == 0 {
		return nil, nil
	}
	p := orgFiltered[0]
	creds, err := c.admin.GetCloneCredentials(ctx, &adminv1.GetCloneCredentialsRequest{
		Org:     p.OrgName,
		Project: p.Name,
	})
	if err != nil {
		return nil, err
	}

	c.gitConfig = &gitutil.Config{
		Remote:            creds.GitRepoUrl,
		Username:          creds.GitUsername,
		Password:          creds.GitPassword,
		PasswordExpiresAt: creds.GitPasswordExpiresAt.AsTime(),
		DefaultBranch:     creds.GitProdBranch,
		Subpath:           creds.GitSubpath,
		ManagedRepo:       creds.GitManagedRepo,
	}
	return c.gitConfig, nil
}

func (c *connection) gitSignature(ctx context.Context, path string) (*object.Signature, error) {
	repo, err := git.PlainOpen(path)
	if err == nil {
		cfg, err := repo.ConfigScoped(config.SystemScope)
		if err == nil && cfg.User.Email != "" && cfg.User.Name != "" {
			// user has git properly configured use that
			return &object.Signature{
				Name:  cfg.User.Name,
				Email: cfg.User.Email,
				When:  time.Now(),
			}, nil
		}
	}

	// use email of rill user
	userResp, err := c.admin.GetCurrentUser(ctx, &adminv1.GetCurrentUserRequest{})
	if err != nil {
		return nil, err
	}

	return &object.Signature{
		Name:  userResp.User.DisplayName,
		Email: userResp.User.Email,
		When:  time.Now(),
	}, nil
}
