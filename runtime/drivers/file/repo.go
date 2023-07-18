package file

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/fsnotify/fsnotify"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

var limit = 500

// Driver implements drivers.RepoStore.
func (c *connection) Driver() string {
	return "file"
}

// Root implements drivers.RepoStore.
func (c *connection) Root() string {
	return c.root
}

// ListRecursive implements drivers.RepoStore.
func (c *connection) ListRecursive(ctx context.Context, instID, glob string) ([]string, error) {
	// Check that folder hasn't been moved
	if err := c.checkRoot(); err != nil {
		return nil, err
	}

	fsRoot := os.DirFS(c.root)
	glob = path.Clean(path.Join("./", glob))

	var paths []string
	err := doublestar.GlobWalk(fsRoot, glob, func(p string, d fs.DirEntry) error {
		// Don't track directories
		if d.IsDir() {
			return nil
		}

		// Exit if we reached the limit
		if len(paths) == limit {
			return fmt.Errorf("glob exceeded limit of %d matched files", limit)
		}

		// Track file (p is already relative to the FS root)
		p = filepath.Join("/", p)
		paths = append(paths, p)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return paths, nil
}

// Get implements drivers.RepoStore.
func (c *connection) Get(ctx context.Context, instID, filePath string) (string, error) {
	filePath = filepath.Join(c.root, filePath)

	b, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// Stat implements drivers.RepoStore.
func (c *connection) Stat(ctx context.Context, instID, filePath string) (*drivers.RepoObjectStat, error) {
	filePath = filepath.Join(c.root, filePath)

	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	return &drivers.RepoObjectStat{
		LastUpdated: info.ModTime(),
	}, nil
}

// Put implements drivers.RepoStore.
func (c *connection) Put(ctx context.Context, instID, filePath string, reader io.Reader) error {
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

// Rename implements drivers.RepoStore.
func (c *connection) Rename(ctx context.Context, instID, fromPath, toPath string) error {
	toPath = path.Join(c.root, toPath)

	fromPath = path.Join(c.root, fromPath)
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
func (c *connection) Delete(ctx context.Context, instID, filePath string) error {
	filePath = filepath.Join(c.root, filePath)
	return os.Remove(filePath)
}

func (c *connection) Sync(ctx context.Context, instID string) error {
	return nil
}

func (c *connection) Watch(ctx context.Context, replay bool, callback drivers.WatchCallback) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	fsRoot := os.DirFS(c.Root())

	var dirs []string
	var files []string

	err = doublestar.GlobWalk(fsRoot, "**", func(p string, d fs.DirEntry) error {
		if d.IsDir() {
			dirs = append(dirs, p)
		} else {
			files = append(files, p)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if replay {
		for _, f := range files {
			err = callback(drivers.WatchEvent{
				Path: filepath.Join("/", f),
				Type: runtimev1.FileEvent_FILE_EVENT_WRITE,
			})
			if err != nil {
				return err
			}
		}
	}

	for _, path := range dirs {
		relativePath := filepath.Join(c.Root(), path)
		err := watcher.Add(relativePath)
		if err != nil {
			return err
		}
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			relativeName, err := filepath.Rel(c.Root(), event.Name)
			if err != nil {
				return err
			}

			e := drivers.WatchEvent{
				Path: filepath.Join("/", relativeName),
			}
			if event.Op&fsnotify.Create != 0 {
				e.Type = runtimev1.FileEvent_FILE_EVENT_WRITE
				fi, err := os.Stat(event.Name)
				// if err != nil the file is removed already
				if err == nil && fi.IsDir() {
					err := watcher.Add(event.Name)
					if err != nil {
						return err
					}

					e.Dir = true
				}
			} else if event.Op&fsnotify.Write != 0 {
				e.Type = runtimev1.FileEvent_FILE_EVENT_WRITE
			} else if event.Op&fsnotify.Remove != 0 {
				e.Type = runtimev1.FileEvent_FILE_EVENT_DELETE
			} else if event.Op&fsnotify.Rename != 0 {
				e.Type = runtimev1.FileEvent_FILE_EVENT_RENAME
			} else {
				e.Type = runtimev1.FileEvent_FILE_EVENT_UNSPECIFIED
			}
			err = callback(e)
			if err != nil {
				return err
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			return err
		case _, ok := <-ctx.Done():
			if !ok {
				return nil
			}

			return ctx.Err()
		}
	}
}
