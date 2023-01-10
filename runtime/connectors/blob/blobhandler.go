package blob

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"gocloud.dev/blob"
	"golang.org/x/sync/errgroup"
)

// increasing this limit can increase speed ingestion
// but may increase bottleneck at duckdb or network/db IO
// set without any benchamarks
const concurrentBlobDownloadLimit = 8

type BlobHandler struct {
	bucket *blob.Bucket
	prefix string
	path   string

	FileNames  []string
	LocalPaths []string
	TempDir    string
}

func (b *BlobHandler) Close() {
	b.bucket.Close()
	os.RemoveAll(b.TempDir)
}

// object path is relative to bucket
func (b *BlobHandler) DownloadObject(ctx context.Context, objpath string) (string, error) {
	rc, err := b.bucket.NewReader(ctx, objpath, nil)
	if err != nil {
		return "", fmt.Errorf("Object(%q).NewReader: %w", objpath, err)
	}
	defer rc.Close()
	objName := filepath.Base(objpath)
	return fileutil.CopyToTempFile(rc, fileutil.Stem(objName), fileutil.FullExt(objName), b.TempDir)
}

// object path is relative to bucket
func (b *BlobHandler) DownloadAll(ctx context.Context) error {
	dir, err := os.MkdirTemp("", "download-temp")
	if err != nil {
		return fmt.Errorf("create temp dir for downloading files failed with error %w", err)
	}
	b.TempDir = dir
	b.LocalPaths = make([]string, len(b.FileNames))

	g, grpCtx := errgroup.WithContext(ctx)
	g.SetLimit(concurrentBlobDownloadLimit)
	for i, file := range b.FileNames {
		objectPath := file
		index := i
		g.Go(func() error {
			localPath, err := b.DownloadObject(grpCtx, objectPath)
			if err != nil {
				return err
			}
			b.LocalPaths[index] = localPath
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		// one of the download failed
		// remove the temp directory
		os.RemoveAll(b.TempDir)
		return err
	}

	return nil
}
