package file

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/rilldata/rill/runtime/drivers"
)

// Driver implements drivers.RepoStore.
func (c *connection) Driver() string {
	return c.driverName
}

// Root implements drivers.RepoStore.
func (c *connection) Root(ctx context.Context) (string, error) {
	return c.root, nil
}

// CommitHash implements drivers.RepoStore.
func (c *connection) CommitHash(ctx context.Context) (string, error) {
	return "", nil
}

// CommitTimestamp implements drivers.RepoStore.
func (c *connection) CommitTimestamp(ctx context.Context) (time.Time, error) {
	return time.Time{}, nil
}

// ListRecursive implements drivers.RepoStore.
func (c *connection) ListRecursive(ctx context.Context, glob string, skipDirs bool) ([]drivers.DirEntry, error) {
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

// Stat implements drivers.RepoStore.
func (c *connection) Stat(ctx context.Context, filePath string) (*drivers.RepoObjectStat, error) {
	filePath = filepath.Join(c.root, filePath)

	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	return &drivers.RepoObjectStat{
		LastUpdated: info.ModTime(),
		IsDir:       info.IsDir(),
	}, nil
}

func (c *connection) FileHash(ctx context.Context, paths []string) (string, error) {
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

func (c *connection) MakeDir(ctx context.Context, dirPath string) error {
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
		return drivers.ErrFileAlreadyExists
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

// Sync implements drivers.RepoStore.
func (c *connection) Sync(ctx context.Context) error {
	return nil
}

// Watch implements drivers.RepoStore.
func (c *connection) Watch(ctx context.Context, cb drivers.WatchCallback) error {
	c.watcherMu.Lock()
	if c.watcher == nil {
		w, err := newWatcher(c.root, c.ignorePaths, c.logger)
		if err != nil {
			c.watcherMu.Unlock()
			return err
		}
		c.watcher = w
	}
	c.watcherCount++
	c.watcherMu.Unlock()

	defer func() {
		c.watcherMu.Lock()
		c.watcherCount--
		if c.watcherCount == 0 {
			c.watcher.close()
			c.watcher = nil
		}
		c.watcherMu.Unlock()
	}()

	return c.watcher.subscribe(ctx, cb)
}
