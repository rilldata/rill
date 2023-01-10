package blob

import (
	"context"
	"fmt"
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
}

// object path is relative to bucket
func (b *BlobHandler) DownloadObject(ctx context.Context, objpath string) (string, error) {
	rc, err := b.bucket.NewReader(ctx, objpath, nil)
	if err != nil {
		return "", fmt.Errorf("Object(%q).NewReader: %w", objpath, err)
	}
	defer rc.Close()
	objName := filepath.Base(objpath)
	return fileutil.CopyToTempFile(rc, fileutil.Stem(objName), fileutil.FullExt(objName))
}

// object path is relative to bucket
func (b *BlobHandler) DownloadAll(ctx context.Context) error {
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
		fileutil.RemoveFiles(b.LocalPaths)
		return err
	}

	return nil
}
