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

// CommitHash implements drivers.RepoStore.
func (c *connection) CommitHash(ctx context.Context, instID string) (string, error) {
	return "", nil
}

// ListRecursive implements drivers.RepoStore.
func (c *connection) ListRecursive(ctx context.Context, instID, glob string) ([]string, error) {
	// Check that folder hasn't been moved
	if err := c.checkRoot(); err != nil {
		return nil, err
	}

	fsRoot := os.DirFS(c.root)
	glob = filepath.Clean(filepath.Join("./", glob))

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
func (c *connection) Delete(ctx context.Context, instID, filePath string) error {
	filePath = filepath.Join(c.root, filePath)
	return os.Remove(filePath)
}

func (c *connection) Sync(ctx context.Context, instID string) error {
	return nil
}

func (c *connection) Watch(ctx context.Context, replay bool, batchInterval time.Duration, cb drivers.WatchCallback) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	var dirs []string
	var replayEvents []drivers.WatchEvent

	fsRoot := os.DirFS(c.root)
	err = doublestar.GlobWalk(fsRoot, "**", func(p string, d fs.DirEntry) error {
		if d.IsDir() {
			dirs = append(dirs, p)
		} else if replay {
			replayEvents = append(replayEvents, drivers.WatchEvent{
				Path: filepath.Join("/", p),
				Type: runtimev1.FileEvent_FILE_EVENT_WRITE,
			})
		}
		return nil
	})
	if err != nil {
		return err
	}

	if replay {
		err := cb(replayEvents)
		if err != nil {
			return err
		}
	}

	for _, path := range dirs {
		err := watcher.Add(filepath.Join(c.root, path))
		if err != nil {
			return err
		}
	}

	// Batch events for a short time
	var buffer []drivers.WatchEvent
	timer := time.NewTimer(batchInterval)

	for {
		select {
		case e, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			path, err := filepath.Rel(c.Root(), e.Name)
			if err != nil {
				return err
			}
			path = filepath.Join("/", path)

			we := drivers.WatchEvent{Path: path}

			if e.Has(fsnotify.Create) || e.Has(fsnotify.Write) {
				we.Type = runtimev1.FileEvent_FILE_EVENT_WRITE
			} else if e.Has(fsnotify.Remove) {
				we.Type = runtimev1.FileEvent_FILE_EVENT_DELETE
			} else if e.Has(fsnotify.Rename) {
				// TODO: Can there be a rename that's not either a delete or a write?
				we.Type = runtimev1.FileEvent_FILE_EVENT_RENAME
			}

			// Check if the file is a directory
			if !e.Has(fsnotify.Remove) {
				info, err := os.Stat(e.Name)
				we.Dir = err == nil && info.IsDir()
			}

			// We need to add new directories to the watcher, but the watcher automatically handles renames and deletes
			if e.Has(fsnotify.Create) && we.Dir {
				err := watcher.Add(e.Name)
				if err != nil {
					return err
				}
			}

			buffer = append(buffer, we)

			// Reset the timer when new events are added.
			// So we only emit a batch when no new events have been observed for a duration of batchInterval.
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(batchInterval)
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			return err
		case <-timer.C:
			if len(buffer) > 0 {
				err := cb(buffer)
				if err != nil {
					return err
				}
				buffer = nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
