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
	"github.com/rilldata/rill/runtime/drivers"
)

var limit = 500

// Driver implements drivers.RepoStore.
func (c *connection) Driver() string {
	return "file"
}

// DSN implements drivers.RepoStore.
func (c *connection) DSN() string {
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

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
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
