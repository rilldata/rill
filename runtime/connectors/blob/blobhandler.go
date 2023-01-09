package blob

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	BlobType   BlobType
	FileNames  []string
	LocalPaths []string
	TempDir    string
}

func (b *BlobHandler) Close() {
	b.bucket.Close()
	os.RemoveAll(b.TempDir)
}

// object path is realtive to bucket
func (b *BlobHandler) DownloadObject(ctx context.Context, objpath string) (string, error) {
	if b.BlobType == File {
		return fmt.Sprintf("%s/%s", b.path, objpath), nil
	}
	rc, err := b.bucket.NewReader(ctx, objpath, nil)
	if err != nil {
		return "", fmt.Errorf("Object(%q).NewReader: %w", objpath, err)
	}
	defer rc.Close()
	objName := filepath.Base(objpath)
	if name, ext, found := strings.Cut(objName, "."); found {
		return fileutil.CopyToTempFile(rc, name, fmt.Sprintf(".%s", ext), b.TempDir)
	}
	//ideally code should never reach here
	return "", fmt.Errorf("malformed file name %s", objpath)
}

// object path is realtive to bucket
func (b *BlobHandler) DownloadAll(ctx context.Context) error {
	dir, err := os.MkdirTemp("", "download-temp")
	if err != nil {
		return fmt.Errorf("create temp dir for downloading files failed with error %w", err)
	}
	b.TempDir = dir
	b.LocalPaths = make([]string, len(b.FileNames))
	if b.BlobType == File {
		copy(b.LocalPaths, b.FileNames)
		return nil
	}
	g := errgroup.Group{}
	g.SetLimit(concurrentBlobDownloadLimit)
	for i, file := range b.FileNames {
		objectPath := file
		index := i
		g.Go(func() error {
			if localPath, err := b.DownloadObject(ctx, objectPath); err == nil {
				b.LocalPaths[index] = localPath
				return err
			} else {
				return err
			}
		})
	}
	if err := g.Wait(); err == nil {
		return nil
	} else {
		// one of the download failed
		// remove the temp directory
		os.RemoveAll(b.TempDir)
		return err
	}
}
