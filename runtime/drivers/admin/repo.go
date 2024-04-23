package admin

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-git/go-git/v5"
	"github.com/rilldata/rill/runtime/drivers"
)

var listFileslimit = 2000

func (h *Handle) Root() string {
	return h.projPath
}

func (h *Handle) CommitHash(ctx context.Context) (string, error) {
	err := h.cloneOrPull(ctx, true)
	if err != nil {
		return "", err
	}

	repo, err := git.PlainOpen(h.repoPath)
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

func (h *Handle) ListRecursive(ctx context.Context, glob string, skipDirs bool) ([]drivers.DirEntry, error) {
	err := h.cloneOrPull(ctx, true)
	if err != nil {
		return nil, err
	}

	fsRoot := os.DirFS(h.projPath)
	glob = path.Clean(path.Join("./", glob))

	var entries []drivers.DirEntry
	err = doublestar.GlobWalk(fsRoot, glob, func(p string, d fs.DirEntry) error {
		if skipDirs && d.IsDir() {
			return nil
		}

		// Exit if we reached the limit
		if len(entries) == listFileslimit {
			return fmt.Errorf("glob exceeded limit of %d matched files", listFileslimit)
		}

		// Track file (p is already relative to the FS root)
		p = filepath.Join("/", p)
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

func (h *Handle) Get(ctx context.Context, filePath string) (string, error) {
	err := h.cloneOrPull(ctx, true)
	if err != nil {
		return "", err
	}

	filePath = filepath.Join(h.projPath, filePath)

	b, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (h *Handle) Stat(ctx context.Context, filePath string) (*drivers.RepoObjectStat, error) {
	err := h.cloneOrPull(ctx, true)
	if err != nil {
		return nil, err
	}

	filePath = filepath.Join(h.projPath, filePath)

	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	return &drivers.RepoObjectStat{
		LastUpdated: info.ModTime(),
		IsDir:       info.IsDir(),
	}, nil
}

func (h *Handle) Put(ctx context.Context, filePath string, reader io.Reader) error {
	return fmt.Errorf("put operation is unsupported")
}

func (h *Handle) MakeDir(ctx context.Context, dirPath string) error {
	return fmt.Errorf("make dir operation is unsupported")
}

func (h *Handle) Rename(ctx context.Context, fromPath, toPath string) error {
	return fmt.Errorf("rename operation is unsupported")
}

func (h *Handle) Delete(ctx context.Context, filePath string, force bool) error {
	return fmt.Errorf("delete operation is unsupported")
}

func (h *Handle) Sync(ctx context.Context) error {
	return h.cloneOrPull(ctx, false)
}

func (h *Handle) Watch(ctx context.Context, callback drivers.WatchCallback) error {
	return fmt.Errorf("watch operation is unsupported")
}
