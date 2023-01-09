package gcs

import (
	"context"
	"fmt"
	_ "io"
	"net/url"
	_ "os"
	"strings"
	_ "sync"

	"cloud.google.com/go/storage"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/connectors"
	rillblob "github.com/rilldata/rill/runtime/connectors/blob"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"google.golang.org/api/option"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/gcsblob"
)

func init() {
	connectors.Register("gcs", connector{})
	// ctx := context.Background()
	// bucket, err := blob.OpenBucket(ctx, "gs://druid-demo.gorill-stage.io/?prefix=safegraph_social_distancing")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// defer bucket.Close()
	// before := func(as func(interface{}) bool) error {
	// 	// Access storage.Query via q here.
	// 	var q *storage.Query
	// 	if as(&q) {
	// 		q.SetAttrSelection([]string{"Name", "Size"})
	// 	}
	// 	return nil
	// }

	// var names []string
	// var wg sync.WaitGroup
	// if iter, _, err := bucket.ListPage(ctx, blob.FirstPageToken, 10, &blob.ListOptions{BeforeList: before}); err == nil {
	// 	names = make([]string, len(iter))
	// 	for i, obj := range iter {
	// 		names[i] = obj.Key
	// 		if err != nil {
	// 			fmt.Printf("failed to parse path %s, %s\n", obj.Key, err)
	// 		}
	// 		wg.Add(1)
	// 		go func(name string) {
	// 			defer wg.Done()
	// 			fmt.Printf("starting copying %s\n", name)
	// 			if rc, err := bucket.NewReader(ctx, name, nil); err == nil {
	// 				defer rc.Close()
	// 				f, err := os.OpenFile("/Users/kanshul/Downloads/test" + name, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	// 				if err != nil {
	// 					fmt.Println(err)
	// 				}
	// 				defer f.Close()
	// 				if _, err = io.Copy(f, rc); err != nil {
	// 					fmt.Println(err)
	// 				}
	// 			} else {
	// 				fmt.Printf("error in opening reader for %s %s", name, err)
	// 			}
	// 			fmt.Printf("ending copying %s\n", name)
	// 		} (obj.Key)
	// 	}
	// }
	// wg.Wait()
	// // fmt.Println(names)
	// // for {
	// // 	obj, err := iter.Next(ctx)
	// // 	if err == io.EOF {
	// // 		break
	// // 	}
	// // 	if err != nil {
	// // 		fmt.Println(err)
	// // 	}
	// // 	fmt.Println(obj)
	// // }
	// // if err := bucket.Copy(ctx, "/Users/kanshul/Documents/projects/rill-developer/000000000165.csv.gz", "000000000165.csv.gz", nil); err != nil {
	// // 	fmt.Println(err)
	// // }

}

var spec = connectors.Spec{
	DisplayName: "Google Cloud Storage",
	Description: "Connect to Google Cloud Storage.",
	Properties: []connectors.PropertySchema{
		{
			Key:         "path",
			DisplayName: "GS URI",
			Description: "Path to file on the disk.",
			Placeholder: "gs://bucket-name/path/to/file.csv",
			Type:        connectors.StringPropertyType,
			Required:    true,
			Hint:        "Note that glob patterns aren't yet supported",
		},
		{
			Key:         "gcp.credentials",
			DisplayName: "GCP credentials",
			Description: "GCP credentials inferred from your local environment.",
			Type:        connectors.InformationalPropertyType,
			Hint:        "Set your local credentials: <code>gcloud auth application-default login</code> Click to learn more.",
			Href:        "https://docs.rilldata.com/using-rill/import-data#setting-google-gcs-credentials",
		},
	},
}

type Config struct {
	Path          string `key:"path"`
	MaxSize       int64  `key:"glob.max_size"`
	MaxDownload   int    `key:"glob.max_download"`
	MaxIterations int    `key:"glob.max_iterations"`
}

func ParseConfig(props map[string]any) (*Config, error) {
	conf := &Config{}
	err := mapstructure.Decode(props, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

type connector struct{}

func (c connector) Spec() connectors.Spec {
	return spec
}

func (c connector) ConsumeAsFile(ctx context.Context, env *connectors.Env, source *connectors.Source) (string, error) {
	conf, err := ParseConfig(source.Properties)
	if err != nil {
		return "", fmt.Errorf("failed to parse config: %w", err)
	}

	client, err := getGcsClient(ctx)
	if err != nil {
		return "", fmt.Errorf("storage.NewClient: %w", err)
	}
	defer client.Close()

	bucket, object, extension, err := gcsURLParts(conf.Path)
	if err != nil {
		return "", fmt.Errorf("failed to parse path %s, %w", conf.Path, err)
	}

	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return "", fmt.Errorf("Object(%q).NewReader: %w", object, err)
	}
	defer rc.Close()

	return fileutil.CopyToTempFile(rc, source.Name, extension)
}

func main() {
	fmt.Println(gcsURLParts("gs://druid-demo.gorill-stage.io/**/nytimes_hex/00000000016[0-9]*.csv.gz"))
}

func gcsURLParts(path string) (string, string, string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", "", "", err
	}
	return u.Host, strings.Replace(u.Path, "/", "", 1), fileutil.FullExt(u.Path), nil
}

func getGcsClient(ctx context.Context) (*storage.Client, error) {
	client, err := storage.NewClient(ctx)
	if err == nil || !strings.Contains(err.Error(), "google: could not find default credentials") {
		return client, err
	}
	client, err = storage.NewClient(ctx, option.WithoutAuthentication())
	if err != nil {
		return nil, err
	}
	return client, nil
}

// Responsibility of caller to close bucketObject
func (c connector) PrepareBlob(ctx context.Context, source *connectors.Source) (*rillblob.BlobHandler, error) {
	conf, err := ParseConfig(source.Properties)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if !doublestar.ValidatePattern(conf.Path) {
		// ideally this should be validated at much earlier stage
		// keeping it here to have gcs specific validations
		return nil, fmt.Errorf("glob pattern %s is invalid", conf.Path)
	}

	bucket, glob, _, err := gcsURLParts(conf.Path)
	bucket = fmt.Sprintf("gs://%s", bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %s, %w", conf.Path, err)
	}
	bucketObj, err := blob.OpenBucket(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to open bucket %s, %w", bucket, err)
	}
	// fetchConfigs := connectors.FetchConfigs{
	// 	MaxSize: conf.MaxSize,
	// 	MaxDownload: conf.MaxDownload,
	// 	MaxIterations: int64(conf.MaxIterations),
	// }
	fetchConfigs := rillblob.FetchConfigs{
		MaxSize:       int64(10 * 1024 * 1024 * 1024),
		MaxDownload:   100,
		MaxIterations: int64(10 * 1024 * 1024 * 1024),
	}
	return rillblob.FetchBlobHandler(ctx, bucketObj, fetchConfigs, glob, bucket)
}
