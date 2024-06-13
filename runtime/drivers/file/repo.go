package file

import (
	"context"
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

var limit = 1000

// Driver implements drivers.RepoStore.
func (c *connection) Driver() string {
	return "file"
}

// Root implements drivers.RepoStore.
func (c *connection) Root() string {
	return c.root
}

// CommitHash implements drivers.RepoStore.
func (c *connection) CommitHash(ctx context.Context) (string, error) {
	return "", nil
}

// ListRecursive implements drivers.RepoStore.
func (c *connection) ListRecursive(ctx context.Context, glob string, skipDirs bool) ([]drivers.DirEntry, error) {
	// Check that folder hasn't been moved
	if err := c.checkRoot(); err != nil {
		return nil, err
	}

	fsRoot := os.DirFS(c.root)
	glob = filepath.Clean(filepath.Join("./", glob))

	var entries []drivers.DirEntry
	err := doublestar.GlobWalk(fsRoot, glob, func(p string, d fs.DirEntry) error {
		if skipDirs && d.IsDir() {
			return nil
		}

		// Exit if we reached the limit
		if len(entries) == limit {
			return fmt.Errorf("glob exceeded limit of %d matched files", limit)
		}

		// Track file (p is already relative to the FS root)
		p = filepath.Join("/", p)
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
	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()
	for _, p := range c.cachedPaths {
		if strings.HasPrefix(strings.TrimLeft(filePath, "/"), strings.TrimLeft(p, "/")) {
			b := c.assetsCache[strings.TrimLeft(filePath, "/")]
			if b != nil {
				return string(b), nil
			}
			break
		}
	}

	filePath = filepath.Join(c.root, filePath)

	b, err := os.ReadFile(filePath)
	if err != nil {
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

func (c *connection) SetCachedPaths(paths []string) {
	c.cachedPaths = paths
}

func (c *connection) GetCachedPaths() []string {
	return c.cachedPaths
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
	cache := make(map[string][]byte)
	for _, p := range c.cachedPaths {
		p = filepath.Join(c.root, p)
		_, err := os.Stat(p)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		err = filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				b, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				rel, err := filepath.Rel(c.root, path)
				if err != nil {
					return err
				}
				cache[strings.TrimLeft(rel, "/")] = b
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	c.assetsCache = cache
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
