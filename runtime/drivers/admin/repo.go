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
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/ctxsync"
	"golang.org/x/sync/singleflight"
	"gopkg.in/yaml.v3"
)

const muxSyncTimeout = 10 * time.Minute

type repo struct {
	h            *Handle
	mu           ctxsync.RWMutex
	singleflight *singleflight.Group

	handshakeExpiresAt time.Time
	synced             bool
	syncErr            error
	ignorePaths        []string
	git                *gitFS
	archive            *archiveFS
	virtual            *virtualFS
}

var _ drivers.RepoStore = (*repo)(nil)

func newRepo(h *Handle) *repo {
	return &repo{
		h:            h,
		mu:           ctxsync.NewRWMutex(),
		singleflight: &singleflight.Group{},
	}
}

func (r *repo) Driver() string {
	return r.h.Driver()
}

func (r *repo) Root(ctx context.Context) (string, error) {
	err := r.rlockEnsureSynced(ctx)
	if err != nil {
		return "", err
	}
	defer r.mu.RUnlock()

	if r.archive != nil {
		return r.archive.root(), nil
	}
	return r.git.root(), nil
}

func (r *repo) CommitHash(ctx context.Context) (string, error) {
	err := r.rlockEnsureSynced(ctx)
	if err != nil {
		return "", err
	}
	defer r.mu.RUnlock()

	if r.archive != nil {
		return r.archive.commitHash(), nil
	}
	return r.git.commitHash()
}

func (r *repo) CommitTimestamp(ctx context.Context) (time.Time, error) {
	err := r.rlockEnsureSynced(ctx)
	if err != nil {
		return time.Time{}, err
	}
	defer r.mu.RUnlock()

	if r.archive != nil {
		return r.archive.commitTimestamp(), nil
	}
	return r.git.commitTimestamp()
}

func (r *repo) ListRecursive(ctx context.Context, glob string, skipDirs bool) ([]drivers.DirEntry, error) {
	err := r.rlockEnsureSynced(ctx)
	if err != nil {
		return nil, err
	}
	defer r.mu.RUnlock()

	var entries []drivers.DirEntry
	for _, root := range r.roots() {
		err := doublestar.GlobWalk(os.DirFS(root), path.Clean(path.Join(".", glob)), func(p string, d fs.DirEntry) error {
			if skipDirs && d.IsDir() {
				return nil
			}
			if len(entries) == drivers.RepoListLimit {
				return drivers.ErrRepoListLimitExceeded
			}
			p = path.Join("/", p) // p is already relative to the root, not absolute
			if drivers.IsIgnored(p, r.ignorePaths) {
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
	}

	return entries, nil
}

func (r *repo) Get(ctx context.Context, path string) (string, error) {
	if drivers.IsIgnored(path, r.ignorePaths) {
		return "", fmt.Errorf("path %q is ignored", path)
	}

	err := r.rlockEnsureSynced(ctx)
	if err != nil {
		return "", err
	}
	defer r.mu.RUnlock()

	var readErr error
	for _, root := range r.roots() {
		fp := filepath.Join(root, path)
		b, err := os.ReadFile(fp)
		if err != nil {
			readErr = err
			if !os.IsNotExist(err) {
				break
			}
			continue
		}
		return string(b), nil
	}

	return "", readErr
}

func (r *repo) Stat(ctx context.Context, path string) (*drivers.RepoObjectStat, error) {
	if drivers.IsIgnored(path, r.ignorePaths) {
		return nil, fmt.Errorf("path %q is ignored", path)
	}

	err := r.rlockEnsureSynced(ctx)
	if err != nil {
		return nil, err
	}
	defer r.mu.RUnlock()

	var statErr error
	for _, root := range r.roots() {
		fp := filepath.Join(root, path)
		info, err := os.Stat(fp)
		if err != nil {
			statErr = err
			if !os.IsNotExist(err) {
				break
			}
			continue
		}
		return &drivers.RepoObjectStat{
			LastUpdated: info.ModTime(),
			IsDir:       info.IsDir(),
		}, nil
	}

	return nil, statErr
}

func (r *repo) FileHash(ctx context.Context, paths []string) (string, error) {
	err := r.rlockEnsureSynced(ctx)
	if err != nil {
		return "", err
	}
	defer r.mu.RUnlock()

	var root string
	if r.archive != nil {
		root = r.archive.root()
	} else {
		root = r.git.root()
	}

	hasher := md5.New()
	for _, path := range paths {
		if drivers.IsIgnored(path, r.ignorePaths) {
			return "", fmt.Errorf("path %q is ignored", path)
		}
		path = filepath.Join(root, path)
		file, err := os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", err
		}
		_, err = io.Copy(hasher, file)
		if err != nil {
			file.Close()
			return "", err
		}
		file.Close()
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (r *repo) Put(ctx context.Context, filePath string, reader io.Reader) error {
	return fmt.Errorf("put operation is unsupported")
}

func (r *repo) MakeDir(ctx context.Context, dirPath string) error {
	return fmt.Errorf("make dir operation is unsupported")
}

func (r *repo) Rename(ctx context.Context, fromPath, toPath string) error {
	return fmt.Errorf("rename operation is unsupported")
}

func (r *repo) Delete(ctx context.Context, filePath string, force bool) error {
	return fmt.Errorf("delete operation is unsupported")
}

func (r *repo) Sync(ctx context.Context) error {
	return r.sync(ctx)
}

func (r *repo) Watch(ctx context.Context, callback drivers.WatchCallback) error {
	return fmt.Errorf("watch operation is unsupported")
}

func (r *repo) close() {
	if r.archive != nil {
		_ = os.RemoveAll(r.archive.tmpDir)
	}
	if r.virtual != nil {
		_ = os.RemoveAll(r.virtual.tmpDir)
	}
}

func (r *repo) rlockEnsureSynced(ctx context.Context) error {
	// Get read lock
	err := r.mu.RLock(ctx)
	if err != nil {
		return err
	}

	// Return with lock held if already synced
	if r.synced {
		return nil
	}

	// Release read lock and clone (which uses a singleflight)
	r.mu.RUnlock()
	err = r.sync(ctx)
	if err != nil {
		return err
	}

	// We know it's synced now. Take read lock and return.
	return r.mu.RLock(ctx)
}

func (r *repo) sync(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "r.sync")
	defer span.End()

	ch := r.singleflight.DoChan("sync", func() (any, error) {
		ctx, cancel := context.WithTimeout(context.Background(), muxSyncTimeout)
		defer cancel()

		err := r.mu.Lock(ctx)
		if err != nil {
			return nil, err
		}
		defer r.mu.Unlock()

		err = r.syncInner(ctx)
		r.synced = err == nil
		r.syncErr = err
		return nil, r.syncErr
	})

	select {
	case res := <-ch:
		return res.Err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *repo) syncInner(ctx context.Context) error {
	err := r.checkSyncHandshake(ctx)
	if err != nil {
		return fmt.Errorf("repo handshake failed: %w", err)
	}

	if r.git != nil {
		err = r.git.sync(ctx)
		if err != nil {
			return fmt.Errorf("git sync failed: %w", err)
		}
	}
	if r.archive != nil {
		err = r.archive.sync(ctx)
		if err != nil {
			return fmt.Errorf("archive sync failed: %w", err)
		}
	}
	if r.virtual != nil {
		err = r.virtual.sync(ctx)
		if err != nil {
			return fmt.Errorf("virtual sync failed: %w", err)
		}
	}

	// Parse `ignore_paths` from `rill.yaml` without fully parsing the project.
	var root string
	if r.archive != nil {
		root = r.archive.root()
	} else {
		root = r.git.root()
	}
	rawYAML, err := os.ReadFile(filepath.Join(root, "rill.yaml"))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read rill.yaml: %w", err)
	}
	if rawYAML != nil {
		yml := &struct {
			IgnorePaths []string `yaml:"ignore_paths"`
		}{}
		err = yaml.Unmarshal(rawYAML, yml)
		if err == nil {
			r.ignorePaths = yml.IgnorePaths
		}
	}

	return nil
}

func (r *repo) syncStatus(ctx context.Context) (bool, error) {
	err := r.mu.RLock(ctx)
	if err != nil {
		return false, err
	}
	defer r.mu.RUnlock()

	return r.synced, r.syncErr
}

// checkSyncHandshake checks and possibly renews the repo details handshake with the admin server.
// Unsafe for concurrent use.
func (r *repo) checkSyncHandshake(ctx context.Context) error {
	if !r.handshakeExpiresAt.Before(time.Now()) {
		return nil
	}

	meta, err := r.h.admin.GetRepoMeta(ctx, &adminv1.GetRepoMetaRequest{
		ProjectId: r.h.config.ProjectID,
	})
	if err != nil {
		return fmt.Errorf("failed to get repo meta: %w", err)
	}

	if meta.GitUrl != "" {
		if r.git == nil {
			repoDir, err := r.h.storage.DataDir("git")
			if err != nil {
				return fmt.Errorf("failed to get git data dir: %w", err)
			}
			r.git = &gitFS{
				h:       r.h,
				repoDir: repoDir,
			}
		}

		r.git.remoteURL = meta.GitUrl
		r.git.branch = meta.GitBranch
		r.git.subpath = meta.GitSubpath
	} else {
		r.git = nil
	}

	if meta.ArchiveDownloadUrl != "" {
		if r.archive == nil {
			tmpDir, err := r.h.storage.RandomTempDir("archive")
			if err != nil {
				return err
			}

			r.archive = &archiveFS{
				h:      r.h,
				tmpDir: tmpDir,
			}
		}

		r.archive.archiveDownloadURL = meta.ArchiveDownloadUrl
		r.archive.archiveID = meta.ArchiveId
		r.archive.archiveCreatedOn = meta.ArchiveCreatedOn.AsTime()
	} else {
		r.archive = nil
	}

	if r.virtual == nil {
		tmpDir, err := r.h.storage.RandomTempDir("virtual")
		if err != nil {
			return err
		}

		r.virtual = &virtualFS{
			h:      r.h,
			tmpDir: tmpDir,
		}
	}

	r.handshakeExpiresAt = meta.ValidUntilTime.AsTime()
	return nil
}

func (r *repo) roots() []string {
	var roots []string
	if r.virtual != nil {
		roots = append(roots, r.virtual.root())
	}
	if r.archive != nil {
		roots = append(roots, r.archive.root())
	}
	if r.git != nil {
		roots = append(roots, r.git.root())
	}
	return roots
}
