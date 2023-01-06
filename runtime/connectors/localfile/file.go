package localfile

import (
	"context"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/connectors"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
)

func init() {
	connectors.Register("local_file", connector{})
}

var spec = connectors.Spec{
	DisplayName: "Local file",
	Description: "Import Locally Stored File.",
	Properties: []connectors.PropertySchema{
		{
			Key:         "path",
			Type:        connectors.StringPropertyType,
			Required:    true,
			DisplayName: "Path",
			Description: "Path or URL to file",
			Placeholder: "/path/to/file",
		},
		{
			Key:         "format",
			Type:        connectors.StringPropertyType,
			Required:    false,
			DisplayName: "Format",
			Description: "Either CSV or Parquet. Inferred if not set.",
			Placeholder: "csv",
		},
		{
			Key:         "csv.delimiter",
			Type:        connectors.StringPropertyType,
			Required:    false,
			DisplayName: "CSV Delimiter",
			Description: "Force delimiter for a CSV file.",
			Placeholder: ",",
		},
	},
}

type Config struct {
	Path          string `mapstructure:"path"`
	Format        string `mapstructure:"format"`
	CSVDelimiter  string `mapstructure:"csv.delimiter"`
	MaxSize       int64  `mapstructure:"max_size" default:int64(10 * 1024 * 1024* 1024)`
	MaxDownload   int    `mapstructure:"max_download" default:int(100)`
	MaxIterations int64  `mapstructure:"max_iterations" default:int64(10000)`
}

func ParseConfig(props map[string]any) (*Config, error) {
	conf := &Config{}
	err := mapstructure.Decode(props, &conf)
	if err != nil {
		return nil, err
	}

	if conf.Format == "" {
		conf.Format = path.Ext(conf.Path)
	}

	return conf, nil
}

type connector struct{}

func (c connector) Spec() connectors.Spec {
	return spec
}

func (c connector) ConsumeAsFile(ctx context.Context, env *connectors.Env, source *connectors.Source) (string, error) {
	return "", errors.New("not implemented")
}

func (c connector) FetchFileNamesForGlob(ctx context.Context, source *connectors.Source) (*connectors.BlobResult, error) {
	conf, err := ParseConfig(source.Properties)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %s, %w", conf.Path, err)
	}
	//todo :: validate path
	bucket, glob := fetchFileParts(conf.Path)


	replaceBucket := fmt.Sprintf("%s%s", "file://", bucket)
	bucketObj, err := blob.OpenBucket(ctx, replaceBucket)
	if err != nil {
		return nil, fmt.Errorf("failed to open bucket %s, %w", bucket, err)
	}
	fetchConfigs := connectors.FetchConfigs{
		MaxSize:       int64(10 * 1024 * 1024 * 1024),
		MaxDownload:   100,
		MaxIterations: int64(10 * 1024 * 1024 * 1024),
	}
	return connectors.FetchFileNames(ctx, bucketObj, fetchConfigs, glob, bucket)
}

func fetchFileParts(path string) (string, string) {
	dir, file := filepath.Split(path)
	bucket, glob := split(dir)
	if glob != "" {
		glob = fmt.Sprintf("%s%s", bucket[strings.LastIndex(bucket, "/")+1:], glob)
		bucket = bucket[:strings.LastIndex(bucket, "/")+1]
	}
	if glob != "" {
		glob = filepath.Join(glob, file)
	} else {
		glob = file
	}
	fmt.Printf("%s ::::: %s %s\n", path, bucket, glob)
	return bucket, glob
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

// hasMeta reports whether path contains any of the magic characters
// recognized by path.Match.
func hasMeta(path string) bool {
	for i := 0; i < len(path); i++ {
		switch path[i] {
		case '*', '?', '[', '\\':
			return true
		}
	}
	return false
}
