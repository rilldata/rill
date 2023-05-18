package github

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	doublestar "github.com/bmatcuk/doublestar/v4"
	"github.com/eapache/go-resiliency/retrier"
	"github.com/rilldata/rill/runtime/drivers"
)

var limit = 500

// Driver implements drivers.RepoStore.
func (c *connection) Driver() string {
	return "github"
}

// Root implements drivers.RepoStore.
func (c *connection) Root() string {
	return c.projectdir
}

// ListRecursive implements drivers.RepoStore.
func (c *connection) ListRecursive(ctx context.Context, instID, glob string) ([]string, error) {
	fsRoot := os.DirFS(c.projectdir)
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
	filePath = filepath.Join(c.projectdir, filePath)

	b, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// Stat implements drivers.RepoStore.
func (c *connection) Stat(ctx context.Context, instID, filePath string) (*drivers.RepoObjectStat, error) {
	filePath = filepath.Join(c.projectdir, filePath)

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
	return fmt.Errorf("Put operation is unsupported")
}

// Rename implements drivers.RepoStore.
func (c *connection) Rename(ctx context.Context, instID, fromPath, toPath string) error {
	return fmt.Errorf("Rename operation is unsupported")
}

// Delete implements drivers.RepoStore.
func (c *connection) Delete(ctx context.Context, instID, filePath string) error {
	return fmt.Errorf("Delete operation is unsupported")
}

// Sync implements drivers.RepoStore.
func (c *connection) Sync(ctx context.Context, instID string) error {
	r := retrier.New(retrier.ExponentialBackoff(retryN, retryWait), nil)
	err := r.Run(func() error { return c.pull(ctx) })
	if err != nil {
		return err
	}
	return nil
}
