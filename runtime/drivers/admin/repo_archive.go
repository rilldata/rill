package admin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rilldata/rill/runtime/pkg/archive"
)

// archiveRepo represents a tarball archive file source.
// It is unsafe for concurrent reads and writes.
type archiveRepo struct {
	h                  *Handle
	tmpDir             string
	archiveDownloadURL string
	archiveID          string
	archiveCreatedOn   time.Time

	filesDir          string
	syncedDownloadURL string
}

func (r *archiveRepo) sync(ctx context.Context) error {
	if r.syncedDownloadURL == r.archiveDownloadURL {
		return nil
	}

	archivePath := filepath.Join(r.tmpDir, "archive.tar.gz")
	defer func() { _ = os.Remove(archivePath) }()

	filesDir, err := os.MkdirTemp(r.tmpDir, "files")
	if err != nil {
		return err
	}

	err = archive.Download(ctx, r.archiveDownloadURL, archivePath, filesDir, false, false)
	if err != nil {
		_ = os.RemoveAll(filesDir)
		return fmt.Errorf("archiveRepo: %w", err)
	}

	_ = os.RemoveAll(r.filesDir)
	r.filesDir = filesDir
	r.syncedDownloadURL = r.archiveDownloadURL
	return nil
}

func (r *archiveRepo) root() string {
	return r.filesDir
}
