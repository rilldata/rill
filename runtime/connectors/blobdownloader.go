package connectors

import (
	"context"
	"fmt"
	"math"
	"path"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/rilldata/rill/runtime/pkg/fileutil"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
)

type FetchConfigs struct {
	MaxSize       int64 `default:int64(10 * 1024 * 1024* 1024)`
	MaxDownload   int   `default:int64(100)`
	MaxIterations int64 `default:int64(10 * 1024 * 1024* 1024)`
}

type BlobResult struct {
	Bucket    *blob.Bucket
	Prefix    string
	FileNames []string
	BlobType  BlobType
	Path      string
}

// object path is realtive to bucket
func (b *BlobResult) DownloadObject(ctx context.Context, objpath string) (string, error) {
	if b.BlobType == File {
		return fmt.Sprintf("%s%s", b.Path, objpath), nil
	}
	rc, err := b.Bucket.NewReader(ctx, objpath, nil)
	if err != nil {
		return "", fmt.Errorf("Object(%q).NewReader: %w", objpath, err)
	}
	defer rc.Close()
	objName := filepath.Base(objpath)
	if name, ext, found := strings.Cut(objName, "."); found {
		return fileutil.CopyToTempFile(rc, name, ext)
	} else {
		//ideally code should never reach here
		return "", fmt.Errorf("malformed file name %s", objpath)
	}
}
func newFetchConfigs() FetchConfigs {
	return FetchConfigs{
		MaxSize:       int64(10 * 1024 * 1024 * 1024),
		MaxDownload:   100,
		MaxIterations: math.MaxInt64,
	}
}

type glob struct {
	glob      string
	recursive bool
}

func newGlob(path string) glob {
	g := glob{glob: path}
	g.recursive = strings.Contains(path, "**")
	return g
}

func Trigger() {
	ctx := context.Background()
	bucket2, err := blob.OpenBucket(ctx, "file:///Users/kanshul/Downloads/")
	if err != nil {
		return
	}
	defer bucket2.Close()
	names, err := FetchFileNames(ctx, bucket2, newFetchConfigs(), "202?/*/green_*.parquet", "file:///Users/kanshul/Downloads")
	if err == nil {
		fmt.Println(names.FileNames)
	} else {
		fmt.Print(err)
	}
}

func blobType(path string) BlobType {
	return File
}

func FetchFileNames(ctx context.Context, bucket *blob.Bucket, config FetchConfigs, glob string, bucketPath string) (*BlobResult, error) {
	g := newGlob(glob)
	prefix, glob := split(glob)
	result := &BlobResult{Prefix: prefix, Bucket: bucket, BlobType: blobType(bucketPath), Path: bucketPath}
	if glob == "" {
		// glob represent plain object
		name := path.Base(glob) // todo :: this returns with extensions
		ext := fileutil.FullExt(glob)
		rc, err := bucket.NewReader(ctx, glob, nil)
		if err != nil {
			return nil, fmt.Errorf("Object(%q).NewReader: %w", glob, err)
		}
		defer rc.Close()

		filename, err := fileutil.CopyToTempFile(rc, name, ext)
		result.FileNames = []string{filename}
		return result, err
	} else {
		before := func(as func(interface{}) bool) error {
			// Access storage.Query via q here.
			var q *storage.Query
			if as(&q) {
				// we only need name and size
				q.SetAttrSelection([]string{"Name", "Size"})
			}
			return nil
		}

		listOptions := blob.ListOptions{BeforeList: before}
		if prefix != "" {
			listOptions.Prefix = prefix
		}

		var size int64 = 0
		var matchCount = 0
		var fileNames []string = make([]string, 0)
		pageSize := int(math.Max(100, float64(config.MaxDownload)))
		fetched := 0
		for token := blob.FirstPageToken; token != nil && size < config.MaxSize && matchCount < config.MaxDownload && fetched < pageSize; {
			iter, nextToken, err := bucket.ListPage(ctx, token, pageSize, &listOptions)
			if err != nil {
				return nil, err
			}
			token = nextToken
			println("%s", string(token))
			// fetched += pageSize
			for _, obj := range iter {
				if match(g, obj.Key) {
					size += obj.Size
					matchCount++
					fileNames = append(fileNames, obj.Key)
				}
			}

		}
		result.FileNames = fileNames
		return result, nil
	}

}

func match(glob glob, fileName string) bool {
	if glob.recursive {
		// recursive match
	}
	matched, _ := path.Match(glob.glob, fileName)
	return matched
}

func split(glob string) (string, string) {
	var b strings.Builder
	for i := 0; i < len(glob); i++ {
		switch glob[i] {
		case '*', '?', '[', '\\':
			return b.String(), glob[i:]
		default:
			b.WriteByte(glob[i])
		}
	}
	return b.String(), ""
}

// // hasMeta reports whether path contains any of the magic characters
// // recognized by path.Match.
// func hasMeta(path string) bool {
// 	for i := 0; i < len(path); i++ {
// 		switch path[i] {
// 		case '*', '?', '[', '\\':
// 			return true
// 		}
// 	}
// 	return false
// }
