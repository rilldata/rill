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

		f, err := d.Info()
		if err != nil {
			return err
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
			Size:  f.Size(),
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
				Size: e.Size,
			})
		}
		cb(watchEvents)
	})
}

// Pull implements drivers.RepoStore.
func (c *connection) Pull(ctx context.Context, discardChanges, forceHandshake bool) error {
	// TODO: If its a Git repository, pull the current branch. Otherwise, this is a no-op.
	return nil
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
