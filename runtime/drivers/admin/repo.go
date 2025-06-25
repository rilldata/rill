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
	"slices"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/ctxsync"
	"github.com/rilldata/rill/runtime/pkg/filewatcher"
	"golang.org/x/sync/singleflight"
	"gopkg.in/yaml.v3"
)

const repoSyncTimeout = 10 * time.Minute

// repo implements the drivers.RepoStore interface.
// It does a handshake using GetRepoMeta on the admin service to discover code files (Git, tarball archive, and/or virtual files).
// It then wraps gitRepo, archiveRepo, and/or virtualRepo to provide the actual file access.
//
// For dev deployments that support file editing, it also supports updating files and committing those back.
// This currently only works for gitRepo, look in its implementation for details.
//
// It's external functions are safe for concurrent use, but the underlying gitRepo/archiveRepo/virtualRepo types are not.
type repo struct {
	// Handle for the parent driver, providing access to the admin service and storage client.
	h *Handle
	// mu is a read-write mutex for accessing and updating files in the repo. It ensures we don't sync files while they're being read.
	mu ctxsync.RWMutex
	// singleflight is used to deduplicate concurrent sync calls.
	singleflight *singleflight.Group

	// handshakeExpiresOn is the next time we should refresh the admin handshake, namely to ensure the Git credentials remain valid.
	handshakeExpiresOn time.Time
	// configUpdatedOn tracks the last_updated_on time from the handshake. We use it to find out if the configuration (apart from credentials) has changed.
	configUpdatedOn time.Time
	// configCtx is a context that is valid until a handshake changes the repo configuration (i.e. configUpdatedOn increases).
	configCtx context.Context
	// configCtxCancel is a cancel function for the configCtx.
	configCtxCancel context.CancelFunc
	// synced is true if files are have been synced successfully.
	// After the first successful sync, it remains true even if the latest sync fails (so syncErr is not nil).
	synced bool
	// syncErr is the last error encountered during sync. It is set to nil when a sync is successful.
	// Even if syncErr is not nil, synced can still be true if a previous sync was successful.
	syncErr error
	// ignorePaths is a list of paths to ignore when listing or accessing files. It's populated by parsing rill.yaml during sync.
	ignorePaths []string
	// git wraps files retrieved from a remote Git repository.
	git *gitRepo
	// archive wraps files retrieved from a remote archive (tarball).
	archive *archiveRepo
	// virtual wraps files that are stored directly in the admin service's virtual_files table in Postgres.
	// It's currently used for alert and reports files, which are not committed to Git or stored in the tarball archive.
	virtual *virtualRepo
}

var _ drivers.RepoStore = (*repo)(nil)

func newRepo(h *Handle) *repo {
	return &repo{
		h:            h,
		mu:           ctxsync.NewRWMutex(),
		singleflight: &singleflight.Group{},
	}
}

// Root implements drivers.RepoStore.
func (r *repo) Root(ctx context.Context) (string, error) {
	err := r.rlockEnsureSynced(ctx)
	if err != nil {
		return "", err
	}
	defer r.mu.RUnlock()

	// NOTE: Virtual files are not available at the root we return here.
	// This is not a problem for the current use cases, but worth keeping in mind.
	if r.archive != nil {
		return r.archive.root(), nil
	}
	return r.git.root(), nil
}

// ListGlob implements drivers.RepoStore.
func (r *repo) ListGlob(ctx context.Context, glob string, skipDirs bool) ([]drivers.DirEntry, error) {
	err := r.rlockEnsureSynced(ctx)
	if err != nil {
		return nil, err
	}
	defer r.mu.RUnlock()

	var entries []drivers.DirEntry
	for _, root := range r.roots() { // Incorporate matches from every underlying file system.
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

// Get implements drivers.RepoStore.
func (r *repo) Get(ctx context.Context, path string) (string, error) {
	if drivers.IsIgnored(path, r.ignorePaths) {
		return "", os.ErrNotExist
	}

	err := r.rlockEnsureSynced(ctx)
	if err != nil {
		return "", err
	}
	defer r.mu.RUnlock()

	var readErr error
	for _, root := range r.roots() { // Search in every underlying file system.
		fp := filepath.Join(root, path)
		b, err := os.ReadFile(fp)
		if err != nil {
			// Keep searching if it's a not exist error. Otherwise break and return the error immediately.
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

// Hash implements drivers.RepoStore.
func (r *repo) Hash(ctx context.Context, paths []string) (string, error) {
	err := r.rlockEnsureSynced(ctx)
	if err != nil {
		return "", err
	}
	defer r.mu.RUnlock()

	// NOTE: Virtual files are not supported here.
	// This is not a problem for the current use cases, but worth keeping in mind.
	var root string
	if r.archive != nil {
		root = r.archive.root()
	} else {
		root = r.git.root()
	}

	hasher := md5.New()
	for _, path := range paths {
		if drivers.IsIgnored(path, r.ignorePaths) {
			continue // Skip if file does not exist
		}
		fp := filepath.Join(root, path)
		file, err := os.Open(fp)
		if err != nil {
			if os.IsNotExist(err) {
				continue // Skip if file does not exist
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

// Stat implements drivers.RepoStore.
func (r *repo) Stat(ctx context.Context, path string) (*drivers.FileInfo, error) {
	if drivers.IsIgnored(path, r.ignorePaths) {
		return nil, os.ErrNotExist
	}

	err := r.rlockEnsureSynced(ctx)
	if err != nil {
		return nil, err
	}
	defer r.mu.RUnlock()

	var statErr error
	for _, root := range r.roots() { // Search in every underlying file system.
		fp := filepath.Join(root, path)
		info, err := os.Stat(fp)
		if err != nil {
			// Keep searching if it's a not exist error. Otherwise break and return the error immediately.
			statErr = err
			if !os.IsNotExist(err) {
				break
			}
			continue
		}
		return &drivers.FileInfo{
			LastUpdated: info.ModTime(),
			IsDir:       info.IsDir(),
		}, nil
	}

	return nil, statErr
}

// Put implements drivers.RepoStore.
func (r *repo) Put(ctx context.Context, path string, reader io.Reader) error {
	if drivers.IsIgnored(path, r.ignorePaths) {
		return fmt.Errorf("can't write to ignored path %q", path)
	}

	err := r.rlockEnsureSynced(ctx)
	if err != nil {
		return err
	}
	defer r.mu.RUnlock()

	if r.git != nil && !r.git.editable {
		return fmt.Errorf("repo is not editable")
	}

	var root string
	if r.archive != nil {
		root = r.archive.root()
	} else {
		root = r.git.root()
	}

	fp := filepath.Join(root, path)

	err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return err
	}

	f, err := os.Create(fp)
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

// MkdirAll implements drivers.RepoStore.
func (r *repo) MkdirAll(ctx context.Context, path string) error {
	if drivers.IsIgnored(path, r.ignorePaths) {
		return fmt.Errorf("can't write to ignored path %q", path)
	}

	err := r.rlockEnsureSynced(ctx)
	if err != nil {
		return err
	}
	defer r.mu.RUnlock()

	if r.git != nil && !r.git.editable {
		return fmt.Errorf("repo is not editable")
	}

	var root string
	if r.archive != nil {
		root = r.archive.root()
	} else {
		root = r.git.root()
	}

	fp := filepath.Join(root, path)

	err = os.MkdirAll(fp, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

// Rename implements drivers.RepoStore.
func (r *repo) Rename(ctx context.Context, fromPath, toPath string) error {
	if drivers.IsIgnored(fromPath, r.ignorePaths) {
		return fmt.Errorf("can't write from ignored path %q", fromPath)
	}
	if drivers.IsIgnored(toPath, r.ignorePaths) {
		return fmt.Errorf("can't write to ignored path %q", toPath)
	}

	err := r.rlockEnsureSynced(ctx)
	if err != nil {
		return err
	}
	defer r.mu.RUnlock()

	if r.git != nil && !r.git.editable {
		return fmt.Errorf("repo is not editable")
	}

	var root string
	if r.archive != nil {
		root = r.archive.root()
	} else {
		root = r.git.root()
	}

	fromPath = filepath.Join(root, fromPath)
	toPath = filepath.Join(root, toPath)

	if _, err := os.Stat(toPath); !strings.EqualFold(fromPath, toPath) && err == nil {
		return os.ErrExist
	}

	err = os.Rename(fromPath, toPath)
	if err != nil {
		return err
	}
	err = os.Chtimes(toPath, time.Now(), time.Now())
	if err != nil {
		return err
	}

	return nil
}

// Delete implements drivers.RepoStore.
func (r *repo) Delete(ctx context.Context, path string, force bool) error {
	if drivers.IsIgnored(path, r.ignorePaths) {
		return fmt.Errorf("can't write to ignored path %q", path)
	}

	err := r.rlockEnsureSynced(ctx)
	if err != nil {
		return err
	}
	defer r.mu.RUnlock()

	if r.git != nil && !r.git.editable {
		return fmt.Errorf("repo is not editable")
	}

	var root string
	if r.archive != nil {
		root = r.archive.root()
	} else {
		root = r.git.root()
	}

	fp := filepath.Join(root, path)

	if force {
		err = os.RemoveAll(fp)
		if err != nil {
			return err
		}
	} else {
		err = os.Remove(fp)
		if err != nil {
			return err
		}
	}

	return nil
}

// Watch implements drivers.RepoStore.
func (r *repo) Watch(ctx context.Context, cb drivers.WatchCallback) error {
	err := r.rlockEnsureSynced(ctx)
	if err != nil {
		return err
	}

	if r.git != nil && !r.git.editable {
		r.mu.RUnlock()
		return fmt.Errorf("repo is not watchable")
	}

	var root string
	if r.archive != nil {
		root = r.archive.root()
	} else {
		root = r.git.root()
	}

	// Derive a context that is cancelled if the repo config changes (i.e. if r.configCtx is cancelled).
	ctx, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(r.configCtx, cancel)
	defer stop()

	w, err := filewatcher.NewWatcher(root, r.ignorePaths, r.h.logger)
	if err != nil {
		r.mu.RUnlock()
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	r.mu.RUnlock()

	return w.Subscribe(ctx, func(events []filewatcher.WatchEvent) {
		if len(events) == 0 {
			return
		}
		var watchEvents []drivers.WatchEvent
		for _, e := range events {
			watchEvents = append(watchEvents, drivers.WatchEvent{
				Type: e.Type,
				Path: e.RelPath,
				Dir:  e.Dir,
			})
		}
		cb(watchEvents)
	})
}

// Sync implements drivers.RepoStore.
func (r *repo) Sync(ctx context.Context) error {
	return r.sync(ctx)
}

// CommitHash implements drivers.RepoStore.
func (r *repo) CommitHash(ctx context.Context) (string, error) {
	err := r.rlockEnsureSynced(ctx)
	if err != nil {
		return "", err
	}
	defer r.mu.RUnlock()

	if r.archive != nil {
		return r.archive.archiveID, nil
	}
	return r.git.commitHash()
}

// CommitTimestamp implements drivers.RepoStore.
func (r *repo) CommitTimestamp(ctx context.Context) (time.Time, error) {
	err := r.rlockEnsureSynced(ctx)
	if err != nil {
		return time.Time{}, err
	}
	defer r.mu.RUnlock()

	if r.archive != nil {
		return r.archive.archiveCreatedOn, nil
	}
	return r.git.commitTimestamp()
}

// close deletes the temporary directories used by the repo.
func (r *repo) close() {
	if r.archive != nil {
		_ = os.RemoveAll(r.archive.tmpDir)
	}
	if r.virtual != nil {
		_ = os.RemoveAll(r.virtual.tmpDir)
	}
}

// roots returns the actual local file system roots for the underlying repos, including the virtual files.
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

// rlockEnsureSynced acquires a read lock after ensuring that the repo is synced.
// If the repo is not synced, it triggers and waits for a sync. If the sync fails, it returns the error without acquiring the read lock.
// If the repo is already synced, it returns immediately and does not trigger a fresh sync (that requires an explicit call to Sync).
func (r *repo) rlockEnsureSynced(ctx context.Context) error {
	// Get read lock
	err := r.mu.RLock(ctx)
	if err != nil {
		return err
	}

	// Return with lock held if already synced.
	// Not checking r.syncErr because we prefer retrying the sync if it failed previously.
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

// sync clones or pulls/updates the repo with the latest code files.
// It is safe for concurrent use and deduplicates concurrent calls (using a singleflight).
func (r *repo) sync(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "r.sync")
	defer span.End()

	ch := r.singleflight.DoChan("sync", func() (any, error) {
		// Using context.Background to prevent context cancellation of the first caller to cause other callers to fail.
		ctx, cancel := context.WithTimeout(context.Background(), repoSyncTimeout)
		defer cancel()

		// Get a write lock. We want to prevent concurrent reads while we're mutating files.
		err := r.mu.Lock(ctx)
		if err != nil {
			return nil, err
		}
		defer r.mu.Unlock()

		// Do the actual sync.
		err = r.syncInner(ctx)
		r.synced = r.synced && (err == nil) // If a sync previously succeeded, we still consider the repo synced even though the latest sync failed.
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

// syncStatus returns the current sync status of the repo.
// If a sync is currently in progress, it waits for it to complete.
// If it returns an error, it may either be the most recent sync error or ctx.Err() from the provided context.
//
// It is safe for concurrent use.
func (r *repo) syncStatus(ctx context.Context) (bool, error) {
	err := r.mu.RLock(ctx)
	if err != nil {
		return false, err
	}
	defer r.mu.RUnlock()

	return r.synced, r.syncErr
}

// syncInner implements the actual sync logic.
// Unlike r.sync(), it is NOT safe for concurrent use and expects r.mu to be held with a write lock.
func (r *repo) syncInner(ctx context.Context) error {
	// Ensure the underlying repos are initialized and have valid credentials.
	err := r.checkSyncHandshake(ctx)
	if err != nil {
		return fmt.Errorf("repo handshake failed: %w", err)
	}

	// Push the sync into the underlying repos. These are created/updated by checkSyncHandshake.
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
	// This enables us to honor `ignore_paths` closer to the file system level, greatly improving performance.
	// NOTE: Not checking r.virtual for rill.yaml because it'll never be stored there.
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
		tmp := &struct {
			IgnorePaths []string `yaml:"ignore_paths"`
		}{}
		err = yaml.Unmarshal(rawYAML, tmp)
		if err == nil {
			if !slices.Equal(r.ignorePaths, tmp.IgnorePaths) {
				r.configChanged()
			}
			r.ignorePaths = tmp.IgnorePaths
		}
	}

	return nil
}

// checkSyncHandshake checks and possibly renews the repo details handshake with the admin server.
// Unsafe for concurrent use.
func (r *repo) checkSyncHandshake(ctx context.Context) error {
	// If the handshake is still valid, return early.
	if !r.handshakeExpiresOn.Before(time.Now()) {
		return nil
	}

	// Handshake with the admin service.
	meta, err := r.h.admin.GetRepoMeta(ctx, &adminv1.GetRepoMetaRequest{
		ProjectId: r.h.config.ProjectID,
	})
	if err != nil {
		return fmt.Errorf("failed to get repo meta: %w", err)
	}

	// Setup or refresh credentials for r.git.
	if meta.GitUrl != "" {
		if r.git == nil {
			repoDir, err := r.h.storage.DataDir("git")
			if err != nil {
				return fmt.Errorf("failed to get git data dir: %w", err)
			}
			r.git = &gitRepo{
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

	// Setup or refresh credentials for r.archive.
	if meta.ArchiveDownloadUrl != "" {
		if r.archive == nil {
			tmpDir, err := r.h.storage.RandomTempDir("archive")
			if err != nil {
				return err
			}

			r.archive = &archiveRepo{
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

	// Setup r.virtual on the first call.
	if r.virtual == nil {
		tmpDir, err := r.h.storage.RandomTempDir("virtual")
		if err != nil {
			return err
		}

		r.virtual = &virtualRepo{
			h:      r.h,
			tmpDir: tmpDir,
		}
	}

	if !r.configUpdatedOn.Equal(meta.LastUpdatedOn.AsTime()) {
		r.configChanged()
		r.configUpdatedOn = meta.LastUpdatedOn.AsTime()
	}
	r.handshakeExpiresOn = meta.ExpiresOn.AsTime()
	return nil
}

// configChanged should be called on changes to the repo configuration, such as branch or subpath (but not when the Git credentials, as happens routinely).
// It cancels the current configCtx and creates a new one, which will be used for
// It is not safe for concurrent use and should be called with a write lock is held.
func (r *repo) configChanged() {
	if r.configCtx != nil {
		r.configCtxCancel()
	}
	r.configCtx, r.configCtxCancel = context.WithCancel(context.Background())
}
