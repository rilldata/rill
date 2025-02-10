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

func (h *Handle) Root(ctx context.Context) (string, error) {
	err := h.rlockEnsureCloned(ctx)
	if err != nil {
		return "", err
	}
	defer h.repoMu.RUnlock()

	return h.projPath, nil
}

func (h *Handle) CommitTimestamp(ctx context.Context) (time.Time, error) {
	err := h.rlockEnsureCloned(ctx)
	if err != nil {
		return time.Time{}, err
	}

	defer h.repoMu.RUnlock()

	if h.archiveDownloadURL != "" {
		return h.archiveCreatedOn, nil
	}

	repo, err := git.PlainOpen(h.repoPath)
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

func (h *Handle) CommitHash(ctx context.Context) (string, error) {
	err := h.rlockEnsureCloned(ctx)
	if err != nil {
		return "", err
	}
	defer h.repoMu.RUnlock()

	if h.archiveDownloadURL != "" {
		return h.archiveID, nil
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
	err := h.rlockEnsureCloned(ctx)
	if err != nil {
		return nil, err
	}
	defer h.repoMu.RUnlock()

	fsRoot := os.DirFS(h.projPath)
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
		if drivers.IsIgnored(p, h.ignorePaths) {
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

func (h *Handle) Get(ctx context.Context, filePath string) (string, error) {
	err := h.rlockEnsureCloned(ctx)
	if err != nil {
		return "", err
	}
	defer h.repoMu.RUnlock()

	fp := filepath.Join(h.projPath, filePath)

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

func (h *Handle) Stat(ctx context.Context, filePath string) (*drivers.RepoObjectStat, error) {
	err := h.rlockEnsureCloned(ctx)
	if err != nil {
		return nil, err
	}
	defer h.repoMu.RUnlock()

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

func (h *Handle) FileHash(ctx context.Context, paths []string) (string, error) {
	err := h.rlockEnsureCloned(ctx)
	if err != nil {
		return "", err
	}
	defer h.repoMu.RUnlock()

	hasher := md5.New()
	for _, path := range paths {
		path = filepath.Join(h.projPath, path)
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
	return h.cloneOrPull(ctx)
}

func (h *Handle) Watch(ctx context.Context, callback drivers.WatchCallback) error {
	return fmt.Errorf("watch operation is unsupported")
}
