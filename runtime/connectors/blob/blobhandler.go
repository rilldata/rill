package blob

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"gocloud.dev/blob"
)

type BlobHandler struct {
	bucket *blob.Bucket
	prefix string
	path   string

	BlobType  BlobType
	FileNames []string
}

func (b *BlobHandler) Close() {
	b.bucket.Close()
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
		return fileutil.CopyToTempFile(rc, name, fmt.Sprintf(".%s", ext))
	}
	//ideally code should never reach here
	return "", fmt.Errorf("malformed file name %s", objpath)
}
