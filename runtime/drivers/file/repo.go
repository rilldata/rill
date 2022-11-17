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

	"github.com/rilldata/rill/runtime/drivers"
)

var excludes = []string{"__pycache__", "build", "dist", "node_modules", "venv"}
var maxDepth = 2

// ListRecursive implements drivers.RepoStore.
// This implementation has some hard-coded rules: it only returns .sql files, it searches
// to a max-depth of 3, and it excludes some common large folders (such as node_modules).
func (c *connection) ListRecursive(ctx context.Context, repoID string) ([]string, error) {
	// Check that folder hasn't been moved
	if err := c.checkRoot(); err != nil {
		return nil, err
	}

	var paths []string
	cleanRoot := path.Clean(c.root)
	rootDepth := strings.Count(cleanRoot, "/")

	err := filepath.WalkDir(c.root, func(p string, d fs.DirEntry, err error) error {
		// Determine whether to skip the directory
		if d.IsDir() {
			// Skip if too deep
			depth := strings.Count(path.Clean(p), "/")
			if depth-rootDepth > maxDepth {
				return filepath.SkipDir
			}
			// Skip if name is excluded
			for _, bad := range excludes {
				if d.Name() == bad {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Track file if it's a .sql file
		if hasSupportForExt(p) {
			pathFromRoot := strings.TrimPrefix(p, cleanRoot)
			paths = append(paths, pathFromRoot)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return paths, nil
}

// Get implements drivers.RepoStore
func (c *connection) Get(ctx context.Context, repoID string, filePath string) (string, error) {
	filePath = path.Join(c.root, filePath)

	b, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// Stat implements drivers.RepoStore by returning the file's stat
func (c *connection) Stat(ctx context.Context, repoID string, filePath string) (*drivers.RepoObjectStat, error) {
	filePath = path.Join(c.root, filePath)
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
	if !hasSupportForExt(filePath) {
		return fmt.Errorf("file repo: can only edit .sql files")
	}

	filePath = path.Join(c.root, filePath)

	err := os.MkdirAll(path.Dir(filePath), os.ModePerm)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, []byte(blob), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (c *connection) PutReader(ctx context.Context, repoID string, filePath string, reader io.Reader) (string, error) {
	filePath = path.Join(c.root, filePath)

	err := os.MkdirAll(path.Dir(filePath), os.ModePerm)
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

	return filePath, nil
}

func (c *connection) Rename(ctx context.Context, repoID string, from string, filePath string) error {
	filePath = path.Join(c.root, filePath)
	from = path.Join(c.root, from)
	return os.Rename(from, filePath)
}

// Delete implements drivers.RepoStore
func (c *connection) Delete(ctx context.Context, repoID string, filePath string) error {
	if !hasSupportForExt(filePath) {
		return fmt.Errorf("file repo: can only edit .sql files")
	}
	filePath = path.Join(c.root, filePath)
	return os.Remove(filePath)
}

func hasSupportForExt(filePath string) bool {
	ext := path.Ext(filePath)
	return ext == ".sql" || ext == ".yaml"
}
