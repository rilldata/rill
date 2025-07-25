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
	"github.com/rilldata/rill/cli/pkg/dotrillcloud"
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

func (c *connection) Status(ctx context.Context) (*drivers.GitStatus, error) {
	// If its a Git repository, return the status of the current branch.
	if !c.isGitRepo() {
		return nil, fmt.Errorf("not a git repository: %s", c.root)
	}

	// if there is a origin set, try with native git configurations
	remote, err := gitutil.ExtractGitRemote(c.root, "origin", false)
	if err == nil && remote.URL != "" {
		err = gitutil.GitFetch(ctx, c.root, nil)
		if err == nil {
			// if native git fetch succeeds, return the status
			gs, err := gitutil.RunGitStatus(c.root, "origin")
			if err != nil {
				return nil, err
			}
			return &drivers.GitStatus{
				Branch:        gs.Branch,
				RemoteURL:     gs.RemoteURL,
				LocalChanges:  gs.LocalChanges,
				LocalCommits:  gs.LocalCommits,
				RemoteCommits: gs.RemoteCommits,
			}, nil
		}
	}

	// if native git fetch fails, try with ephemeral token - this may be a managed git project

	// Get authenticated admin client
	if c.driverConfig.AccessToken == "" {
		// if the user is not authenticated, we cannot fetch the project
		// return the best effort status
		gs, err := gitutil.RunGitStatus(c.root, "origin")
		if err != nil {
			return nil, err
		}
		return &drivers.GitStatus{
			Branch:    gs.Branch,
			RemoteURL: gs.RemoteURL,
		}, nil
	}

	config, err := c.loadGitConfig(ctx)
	if err != nil {
		if !errors.Is(err, drivers.ErrNotFound) {
			return nil, err
		}
		// If the project is not found return the best effort status
		gs, err := gitutil.RunGitStatus(c.root, "origin")
		if err != nil {
			return nil, err
		}
		return &drivers.GitStatus{
			Branch:    gs.Branch,
			RemoteURL: gs.RemoteURL,
		}, nil
	}

	err = gitutil.GitFetch(ctx, c.root, config)
	if err != nil {
		return nil, err
	}
	gs, err := gitutil.RunGitStatus(c.root, config.RemoteName())
	if err != nil {
		return nil, err
	}
	return &drivers.GitStatus{
		Branch:        gs.Branch,
		RemoteURL:     gs.RemoteURL,
		LocalChanges:  gs.LocalChanges,
		LocalCommits:  gs.LocalCommits,
		RemoteCommits: gs.RemoteCommits,
	}, nil
}

// Pull implements drivers.RepoStore.
func (c *connection) Pull(ctx context.Context, discardChanges, forceHandshake bool) error {
	// If its a Git repository, pull the current branch. Otherwise, this is a no-op.
	if !c.isGitRepo() {
		return nil
	}
	origin, err := gitutil.ExtractGitRemote(c.root, "origin", false)
	if err == nil && origin.URL != "" {
		out, err := gitutil.RunGitPull(ctx, c.root, discardChanges, "", "origin")
		if err == nil && strings.Contains(out, "Already up to date") {
			return nil
		}
	}
	// if native git pull fails, try with ephemeral token - this may be a managed git project

	if c.driverConfig.AccessToken == "" {
		// This should ideally not happen since otherwise user would not be able to clone the repo
		return nil
	}

	gitConfig, err := c.loadGitConfig(ctx)
	if err != nil {
		return err
	}

	remote, err := gitConfig.FullyQualifiedRemote()
	if err != nil {
		return err
	}

	_, err = gitutil.RunGitPull(ctx, c.root, discardChanges, remote, gitConfig.RemoteName())
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

	remote, err := gitutil.ExtractGitRemote(c.root, "origin", false)
	if err == nil && remote.URL != "" {
		st, err := gitutil.RunGitStatus(c.root, "origin")
		if err != nil {
			return err
		}
		if st.RemoteCommits > 0 && !force {
			return nil
		}

		// generate git signature
		author, err := gitutil.NativeGitSignature(ctx, c.root)
		if err == nil {
			err = gitutil.CommitAndForcePush(ctx, c.root, &gitutil.Config{Remote: st.RemoteURL, DefaultBranch: st.Branch}, message, author)
			if err == nil {
				return nil
			}
		}
	}
	// if native git push fails, try with ephemeral token - this may be a managed git project

	// Get authenticated admin client
	if c.driverConfig.AccessToken == "" {
		// This should ideally not happen since otherwise user would not be able to clone the repo
		return nil
	}

	config, err := c.loadGitConfig(ctx)
	if err != nil {
		return err
	}

	// fetch the status again
	gs, err := gitutil.RunGitStatus(c.root, config.RemoteName())
	if err != nil {
		return err
	}
	if gs.RemoteCommits > 0 && !force {
		return nil
	}

	author, err := c.gitSignature(ctx, c.root)
	if err != nil {
		return err
	}

	err = gitutil.CommitAndForcePush(ctx, c.root, config, message, author)
	if err != nil {
		return err
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

// loadGitConfig loads the git configuration for the repository
// Should be called with c.gitMu held.
func (c *connection) loadGitConfig(ctx context.Context) (*gitutil.Config, error) {
	if c.gitConfig != nil && !c.gitConfig.IsExpired() {
		return c.gitConfig, nil
	}
	config, err := dotrillcloud.GetAll(c.root, c.driverConfig.AdminURL)
	if err != nil {
		return nil, err
	}
	if config == nil || config.ProjectID == "" {
		return nil, drivers.ErrNotFound
	}
	proj, err := c.admin.GetProjectByID(ctx, &adminv1.GetProjectByIDRequest{
		Id: config.ProjectID,
	})
	if err != nil {
		return nil, err
	}

	resp, err := c.admin.GetCloneCredentials(ctx, &adminv1.GetCloneCredentialsRequest{
		Organization: proj.Project.OrgName,
		Project:      proj.Project.Name,
	})
	if err != nil {
		return nil, err
	}

	c.gitConfig = &gitutil.Config{
		Remote:            resp.GitRepoUrl,
		Username:          resp.GitUsername,
		Password:          resp.GitPassword,
		PasswordExpiresAt: resp.GitPasswordExpiresAt.AsTime(),
		DefaultBranch:     resp.GitProdBranch,
		Subpath:           resp.GitSubpath,
		ManagedRepo:       resp.GitManagedRepo,
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
