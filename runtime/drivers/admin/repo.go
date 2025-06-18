package admin

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-git/go-git/v5"
	"github.com/rilldata/rill/runtime/drivers"
)

func (c *Connection) Root(ctx context.Context) (string, error) {
	err := c.rlockEnsureCloned(ctx)
	if err != nil {
		return "", err
	}
	defer c.repoMu.RUnlock()

	return c.projPath, nil
}

func (c *Connection) CommitTimestamp(ctx context.Context) (time.Time, error) {
	err := c.rlockEnsureCloned(ctx)
	if err != nil {
		return time.Time{}, err
	}
	defer c.repoMu.RUnlock()

	if c.archiveDownloadURL != "" {
		return c.archiveCreatedOn, nil
	}

	repo, err := git.PlainOpen(c.repoPath)
	if err != nil {
		return time.Time{}, err
	}

	ref, err := repo.Head()
	if err != nil {
		return time.Time{}, err
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return time.Time{}, err
	}

	return commit.Author.When, nil
}

func (c *Connection) CommitHash(ctx context.Context) (string, error) {
	err := c.rlockEnsureCloned(ctx)
	if err != nil {
		return "", err
	}
	defer c.repoMu.RUnlock()

	if c.archiveDownloadURL != "" {
		return c.archiveID, nil
	}

	repo, err := git.PlainOpen(c.repoPath)
	if err != nil {
		return "", err
	}

	ref, err := repo.Head()
	if err != nil {
		return "", err
	}

	if ref.Hash().IsZero() {
		return "", nil
	}

	return ref.Hash().String(), nil
}

func (c *Connection) ListRecursive(ctx context.Context, glob string, skipDirs bool) ([]drivers.DirEntry, error) {
	err := c.rlockEnsureCloned(ctx)
	if err != nil {
		return nil, err
	}
	defer c.repoMu.RUnlock()

	fsRoot := os.DirFS(c.projPath)
	glob = path.Clean(path.Join(".", glob))

	var entries []drivers.DirEntry
	err = doublestar.GlobWalk(fsRoot, glob, func(p string, d fs.DirEntry) error {
		if skipDirs && d.IsDir() {
			return nil
		}

		// Exit if we reached the limit
		if len(entries) == drivers.RepoListLimit {
			return drivers.ErrRepoListLimitExceeded
		}

		// Track file (p is already relative to the FS root)
		p = path.Join("/", p)
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

func (c *Connection) Get(ctx context.Context, filePath string) (string, error) {
	err := c.rlockEnsureCloned(ctx)
	if err != nil {
		return "", err
	}
	defer c.repoMu.RUnlock()

	fp := filepath.Join(c.projPath, filePath)

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

func (c *Connection) Stat(ctx context.Context, filePath string) (*drivers.RepoObjectStat, error) {
	err := c.rlockEnsureCloned(ctx)
	if err != nil {
		return nil, err
	}
	defer c.repoMu.RUnlock()

	filePath = filepath.Join(c.projPath, filePath)

	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	return &drivers.RepoObjectStat{
		LastUpdated: info.ModTime(),
		IsDir:       info.IsDir(),
	}, nil
}

func (c *Connection) FileHash(ctx context.Context, paths []string) (string, error) {
	err := c.rlockEnsureCloned(ctx)
	if err != nil {
		return "", err
	}
	defer c.repoMu.RUnlock()

	hasher := md5.New()
	for _, path := range paths {
		path = filepath.Join(c.projPath, path)
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

func (c *Connection) Put(ctx context.Context, filePath string, reader io.Reader) error {
	return fmt.Errorf("put operation is unsupported")
}

func (c *Connection) MakeDir(ctx context.Context, dirPath string) error {
	return fmt.Errorf("make dir operation is unsupported")
}

func (c *Connection) Rename(ctx context.Context, fromPath, toPath string) error {
	return fmt.Errorf("rename operation is unsupported")
}

func (c *Connection) Delete(ctx context.Context, filePath string, force bool) error {
	return fmt.Errorf("delete operation is unsupported")
}

func (c *Connection) Sync(ctx context.Context) error {
	return c.cloneOrPull(ctx)
}

func (c *Connection) Watch(ctx context.Context, callback drivers.WatchCallback) error {
	return fmt.Errorf("watch operation is unsupported")
}
