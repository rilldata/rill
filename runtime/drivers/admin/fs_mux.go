package admin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/ctxsync"
	"golang.org/x/sync/singleflight"
	"gopkg.in/yaml.v3"
)

const muxSyncTimeout = 10 * time.Minute

type muxFS struct {
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

func (fs *muxFS) close() error {
	if fs.archive != nil {
		_ = os.RemoveAll(fs.archive.tmpDir)
	}
	if fs.virtual != nil {
		_ = os.RemoveAll(fs.virtual.tmpDir)
	}
	return nil
}

func (fs *muxFS) sync(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "fs.sync")
	defer span.End()

	ch := fs.singleflight.DoChan("sync", func() (any, error) {
		ctx, cancel := context.WithTimeout(context.Background(), muxSyncTimeout)
		defer cancel()

		err := fs.mu.Lock(ctx)
		if err != nil {
			return nil, err
		}
		defer fs.mu.Unlock()

		err = fs.syncInner(ctx)
		fs.synced = err == nil
		fs.syncErr = err
		return nil, fs.syncErr
	})

	select {
	case res := <-ch:
		return res.Err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (fs *muxFS) syncInner(ctx context.Context) error {
	err := fs.checkHandshake(ctx)
	if err != nil {
		return fmt.Errorf("repo handshake failed: %w", err)
	}

	if fs.git != nil {
		err = fs.git.sync(ctx)
		if err != nil {
			return fmt.Errorf("git sync failed: %w", err)
		}
	}
	if fs.archive != nil {
		err = fs.archive.sync(ctx)
		if err != nil {
			return fmt.Errorf("archive sync failed: %w", err)
		}
	}
	if fs.virtual != nil {
		err = fs.virtual.sync(ctx)
		if err != nil {
			return fmt.Errorf("virtual sync failed: %w", err)
		}
	}

	// Parse `ignore_paths` from `rill.yaml` without fully parsing the project.
	var rawYAML []byte
	if fs.git != nil {
		// TODO
	} else if fs.archive != nil {
		// TODO
	}
	if rawYAML != nil {
		yml := &struct {
			IgnorePaths []string `yaml:"ignore_paths"`
		}{}
		err = yaml.Unmarshal(rawYAML, yml)
		if err == nil {
			fs.ignorePaths = yml.IgnorePaths
		}
	}

	return nil
}

// checkHandshake checks and possibly renews the repo details handshake with the admin server.
// Unsafe for concurrent use.
func (fs *muxFS) checkHandshake(ctx context.Context) error {
	if !fs.handshakeExpiresAt.Before(time.Now()) {
		return nil
	}

	meta, err := fs.h.admin.GetRepoMeta(ctx, &adminv1.GetRepoMetaRequest{
		ProjectId: fs.h.config.ProjectID,
	})
	if err != nil {
		return fmt.Errorf("failed to get repo meta: %w", err)
	}

	if meta.GitUrl != "" {
		if fs.git == nil {
			repoDir, err := fs.h.storage.DataDir("git")
			if err != nil {
				return fmt.Errorf("failed to get git data dir: %w", err)
			}
			fs.git = &gitFS{
				h:       fs.h,
				repoDir: repoDir,
			}
		}

		fs.git.remoteURL = meta.GitUrl
		fs.git.branch = meta.GitBranch
		fs.git.subpath = meta.GitSubpath
	} else {
		fs.git = nil
	}

	if meta.ArchiveDownloadUrl != "" {
		if fs.archive == nil {
			tmpDir, err := fs.h.storage.RandomTempDir("archive")
			if err != nil {
				return err
			}

			fs.archive = &archiveFS{
				h:      fs.h,
				tmpDir: tmpDir,
			}
		}

		fs.archive.archiveDownloadURL = meta.ArchiveDownloadUrl
		fs.archive.archiveID = meta.ArchiveId
		fs.archive.archiveCreatedOn = meta.ArchiveCreatedOn.AsTime()
	} else {
		fs.archive = nil
	}

	if fs.virtual == nil {
		tmpDir, err := fs.h.storage.RandomTempDir("virtual")
		if err != nil {
			return err
		}

		fs.virtual = &virtualFS{
			h:      fs.h,
			tmpDir: tmpDir,
		}
	}

	fs.handshakeExpiresAt = meta.ValidUntilTime.AsTime()
	return nil
}

func (fs *muxFS) syncStatus(ctx context.Context) (bool, error) {
	err := fs.mu.RLock(ctx)
	if err != nil {
		return false, err
	}
	defer fs.mu.RUnlock()

	return fs.synced, fs.syncErr
}

func (fs *muxFS) rlockEnsureSynced(ctx context.Context) error {
	// Get read lock
	err := fs.mu.RLock(ctx)
	if err != nil {
		return err
	}

	// Return with lock held if already synced
	if fs.synced {
		return nil
	}

	// Release read lock and clone (which uses a singleflight)
	fs.mu.RUnlock()
	err = fs.sync(ctx)
	if err != nil {
		return err
	}

	// We know it's synced now. Take read lock and return.
	return fs.mu.RLock(ctx)
}

// generateVirtualPath generates a virtual path inside the project path.
func generateVirtualPath(projPath string) string {
	return filepath.Join(projPath, "__virtual__")
}
