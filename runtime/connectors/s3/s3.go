package s3

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/connectors"
	rillblob "github.com/rilldata/rill/runtime/connectors/blob"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"gocloud.dev/blob/s3blob"
)

func init() {
	connectors.Register("s3", connector{})
}

var spec = connectors.Spec{
	DisplayName: "Amazon S3",
	Description: "Connect to AWS S3 Storage.",
	Properties: []connectors.PropertySchema{
		{
			Key:         "path",
			DisplayName: "S3 URI",
			Description: "Path to file on the disk.",
			Placeholder: "s3://bucket-name/path/to/file.csv",
			Type:        connectors.StringPropertyType,
			Required:    true,
			Hint:        "Note that glob patterns aren't yet supported",
		},
		{
			Key:         "aws.region",
			DisplayName: "AWS region",
			Description: "AWS Region for the bucket.",
			Placeholder: "us-east-1",
			Type:        connectors.StringPropertyType,
			Required:    false,
			Hint:        "Rill will use the default region in your local AWS config, unless set here.",
		},
		{
			Key:         "aws.credentials",
			DisplayName: "AWS credentials",
			Description: "AWS credentials inferred from your local environment.",
			Type:        connectors.InformationalPropertyType,
			Hint:        "Set your local credentials: <code>aws configure</code> Click to learn more.",
			Href:        "https://docs.rilldata.com/using-rill/import-data#setting-amazon-s3-credentials",
		},
	},
}

type Config struct {
	Path              string `mapstructure:"path"`
	AWSRegion         string `mapstructure:"aws.region"`
	MaxTotalSize      int64  `mapstructure:"glob.max_total_size"`
	MaxMatchedObjects int    `mapstructure:"glob.max_matched_objects"`
	MaxObjectsListed  int64  `mapstructure:"glob.max_objects_listed"`
	PageSize          int    `mapstructure:"glob.page_size"`
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

func (c connector) ConsumeAsFile(ctx context.Context, env *connectors.Env, source *connectors.Source) ([]string, error) {
	conf, err := ParseConfig(source.Properties)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if !doublestar.ValidatePattern(conf.Path) {
		// ideally this should be validated at much earlier stage
		// keeping it here to have s3 specific validations
		return nil, fmt.Errorf("glob pattern %s is invalid", conf.Path)
	}

	bucket, glob, _, err := s3URLParts(conf.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %s, %w", conf.Path, err)
	}

	sess, err := getAwsSessionConfig(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}

	bucketObj, err := s3blob.OpenBucket(ctx, sess, bucket, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open bucket %s, %w", bucket, err)
	}
	defer bucketObj.Close()

	fetchConfigs := rillblob.FetchConfigs{
		MaxTotalSize:      conf.MaxTotalSize,
		MaxMatchedObjects: conf.MaxMatchedObjects,
		MaxObjectsListed:  conf.MaxObjectsListed,
		PageSize:          conf.PageSize,
	}
	return rillblob.FetchFileNames(ctx, bucketObj, fetchConfigs, glob, bucket)
}

func s3URLParts(path string) (string, string, string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", "", "", err
	}
	return u.Host, strings.Replace(u.Path, "/", "", 1), fileutil.FullExt(u.Path), nil
}

func getAwsSessionConfig(conf *Config) (*session.Session, error) {
	if conf.AWSRegion != "" {
		return session.NewSession(&aws.Config{
			Region: aws.String(conf.AWSRegion),
		})
	}
	return session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
}
