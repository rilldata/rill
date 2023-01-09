package blob

import (
	"context"
	"fmt"
	"math"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/bmatcuk/doublestar/v4"

	"gocloud.dev/blob"
)

type FetchConfigs struct {
	MaxSize       int64 `default:int64(10 * 1024 * 1024* 1024)`
	MaxDownload   int   `default:int64(100)`
	MaxIterations int64 `default:int64(10 * 1024 * 1024* 1024)`
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

// func Trigger() {
// 	ctx := context.Background()
// 	bucket2, err := blob.OpenBucket(ctx, "gs://druid-demo.gorill-stage.io")
// 	if err != nil {
// 		return
// 	}
// 	defer bucket2.Close()
// 	names, err := FetchFileNames(ctx, bucket2, newFetchConfigs(),
// 		"safegraph/nytimes_hex/00000000016[0-9].csv.gz",
// 		"gs://druid-demo.gorill-stage.io/safegraph/nytimes_hex")
// 	if err == nil {
// 		fmt.Println(names.FileNames)
// 	} else {
// 		fmt.Print(err)
// 	}
// }

// todo :: anshul :: check caps here
func blobType(path string) BlobType {
	if strings.Contains(path, "file") {
		return File
	} else if strings.Contains(path, "gs") {
		return GCS
	} else if strings.Contains(path, "s3") {
		return S3
	}
	return File
}

// todo :: return error here
func FetchBlobHandler(ctx context.Context, bucket *blob.Bucket, config FetchConfigs, glob string, bucketPath string) (*BlobHandler, error) {
	g := newGlob(glob)
	prefix, glob := split(glob)
	result := &BlobHandler{prefix: prefix, bucket: bucket, BlobType: blobType(bucketPath), path: bucketPath}
	if glob == "" {
		// glob represent plain object
		result.FileNames = []string{glob}
		return result, nil
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
		for token := blob.FirstPageToken; token != nil; {
			iter, nextToken, err := bucket.ListPage(ctx, token, pageSize, &listOptions)
			if err != nil {
				return nil, err
			}
			token = nextToken
			fmt.Printf("listing %d\n", len(iter))
			// fetched += pageSize
			for _, obj := range iter {
				if match(g, obj.Key) {
					size += obj.Size
					matchCount++
					fileNames = append(fileNames, obj.Key)
				}
			}

			if size > config.MaxSize || matchCount > config.MaxDownload || fetched > pageSize {
				return nil, fmt.Errorf("glob pattern ")
			}

		}
		result.FileNames = fileNames
		return result, nil
	}

}

func match(glob glob, fileName string) bool {
	// if glob.recursive {
	// 	matched, _ := doublestar.Match(glob.glob, fileName)
	// 	return matched
	// }
	// matched, _ := path.Match(glob.glob, fileName)
	// return matched

	matched, _ := doublestar.Match(glob.glob, fileName)
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
