package admin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rilldata/rill/runtime/pkg/archive"
)

type archiveRepo struct {
	h                  *Handle
	tmpDir             string
	archiveDownloadURL string
	archiveID          string
	archiveCreatedOn   time.Time

	syncedDownloadURL string
}

func (r *archiveRepo) sync(ctx context.Context) error {
	if r.syncedDownloadURL == r.archiveDownloadURL {
		return nil
	}

	_ = os.RemoveAll(r.tmpDir)

	dst, err := generateTmpPath(r.tmpDir, "admin_driver_zipped_repo", ".tar.gz")
	if err != nil {
		return fmt.Errorf("archiveRepo: %w", err)
	}
	defer func() { _ = os.Remove(dst) }()

	err = archive.Download(ctx, r.archiveDownloadURL, dst, r.tmpDir, true, false)
	if err != nil {
		return fmt.Errorf("archiveRepo: %w", err)
	}

	r.syncedDownloadURL = r.archiveDownloadURL
	return nil
}

func (r *archiveRepo) root() string {
	return r.tmpDir
}

func (r *archiveRepo) commitHash() string {
	return r.archiveID
}

func (r *archiveRepo) commitTimestamp() time.Time {
	return r.archiveCreatedOn
}

// generateTmpPath generates a temporary path with a random suffix.
// It uses the format <dir>/<base><random><ext>.
// The output path is absolute.
func generateTmpPath(dir, base, ext string) (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("generate tmp path: %w", err)
	}

	r := hex.EncodeToString(b)

	p := filepath.Join(dir, base+r+ext)

	p, err = filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("generate tmp path: %w", err)
	}

	return p, nil
}
