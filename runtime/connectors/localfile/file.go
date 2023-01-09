package localfile

import (
	"context"
	"errors"
	"fmt"
	"path"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/connectors"
	rillblob "github.com/rilldata/rill/runtime/connectors/blob"

	"github.com/bmatcuk/doublestar/v4"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob" // blank import required
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
	MaxSize       int64  `mapstructure:"glob.max_size"`
	MaxDownload   int    `mapstructure:"glob.max_download"`
	MaxIterations int64  `mapstructure:"glob.max_iterations"`
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

func (c connector) PrepareBlob(ctx context.Context, source *connectors.Source) (*rillblob.BlobHandler, error) {
	conf, err := ParseConfig(source.Properties)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %s, %w", conf.Path, err)
	}
	//todo :: validate path
	// bucket, glob := fetchFileParts(conf.Path)

	bucket, glob := doublestar.SplitPattern(conf.Path)

	replaceBucket := fmt.Sprintf("%s%s", "file://", bucket)
	bucketObj, err := blob.OpenBucket(ctx, replaceBucket)
	if err != nil {
		return nil, fmt.Errorf("failed to open bucket %s, %w", bucket, err)
	}
	fetchConfigs := rillblob.FetchConfigs{
		MaxSize:       conf.MaxSize,
		MaxDownload:   conf.MaxDownload,
		MaxIterations: conf.MaxIterations,
	}
	return rillblob.FetchBlobHandler(ctx, bucketObj, fetchConfigs, glob, bucket)
}
