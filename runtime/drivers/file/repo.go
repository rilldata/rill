package file

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/rilldata/rill/runtime/drivers"
)

var limit = 500

// Driver implements drivers.RepoStore
func (c *connection) Driver() string {
	return "file"
}

// DSN implements drivers.RepoStore
func (c *connection) DSN() string {
	return c.root
}

// ListRecursive implements drivers.RepoStore.
func (c *connection) ListRecursive(ctx context.Context, repoID string, glob string) ([]string, error) {
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

// Get implements drivers.RepoStore
func (c *connection) Get(ctx context.Context, repoID string, filePath string) (string, error) {
	filePath = filepath.Join(c.root, filePath)

	b, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// Stat implements drivers.RepoStore
func (c *connection) Stat(ctx context.Context, repoID string, filePath string) (*drivers.RepoObjectStat, error) {
	filePath = filepath.Join(c.root, filePath)

	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	return &drivers.RepoObjectStat{
		LastUpdated: info.ModTime(),
	}, nil
}

// PutBlob implements drivers.RepoStore
func (c *connection) PutBlob(ctx context.Context, repoID string, filePath string, blob string) error {
	filePath = filepath.Join(c.root, filePath)

	err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, []byte(blob), 0644)
	if err != nil {
		return err
	}

	return nil
}

// PutReader implements drivers.RepoStore
func (c *connection) PutReader(ctx context.Context, repoID string, filePath string, reader io.Reader) (string, error) {
	originalPath := filePath
	filePath = filepath.Join(c.root, filePath)

	err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return "", err
	}

	f, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, reader)
	if err != nil {
		return "", err
	}

	return originalPath, nil
}

// Rename implements drivers.RepoStore
func (c *connection) Rename(ctx context.Context, repoID string, from string, filePath string) error {
	filePath = path.Join(c.root, filePath)
	from = path.Join(c.root, from)
	err := os.Rename(from, filePath)
	if err != nil {
		return err
	}
	return os.Chtimes(filePath, time.Now(), time.Now())
}

// Delete implements drivers.RepoStore
func (c *connection) Delete(ctx context.Context, repoID string, filePath string) error {
	filePath = filepath.Join(c.root, filePath)
	return os.Remove(filePath)
}
