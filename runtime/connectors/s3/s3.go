package s3

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/connectors"
	rillblob "github.com/rilldata/rill/runtime/connectors/blob"
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
			Key:         "region",
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

type s3Config struct {
	connectors.Config     `mapstructure:",squash"`
	AWSRegion             string `mapstructure:"region"`
	GlobMaxTotalSize      int64  `mapstructure:"glob.max_total_size"`
	GlobMaxObjectsMatched int    `mapstructure:"glob.max_objects_matched"`
	GlobMaxObjectsListed  int64  `mapstructure:"glob.max_objects_listed"`
	GlobPageSize          int    `mapstructure:"glob.page_size"`
}

func parseConfig(props map[string]any) (*s3Config, error) {
	conf := &s3Config{}
	err := mapstructure.Decode(props, conf)
	if err != nil {
		return nil, err
	}
	if !doublestar.ValidatePattern(conf.Path) {
		return nil, fmt.Errorf("glob pattern %s is invalid", conf.Path)
	}
	return conf, nil
}

type connector struct{}

func (c connector) Spec() connectors.Spec {
	return spec
}

func (c connector) ConsumeAsFiles(ctx context.Context, env *connectors.Env, source *connectors.Source) ([]string, error) {
	conf, err := parseConfig(source.Properties)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	bucket, glob, err := s3URLParts(conf.Path)
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
		GlobMaxTotalSize:      conf.GlobMaxTotalSize,
		GlobMaxObjectsMatched: conf.GlobMaxObjectsMatched,
		GlobMaxObjectsListed:  conf.GlobMaxObjectsListed,
		GlobPageSize:          conf.GlobPageSize,
	}
	return rillblob.FetchFileNames(ctx, bucketObj, fetchConfigs, glob, bucket)
}

func s3URLParts(path string) (string, string, error) {
	trimmedPath := strings.Replace(path, "s3://", "", 1)
	bucket, glob, found := strings.Cut(trimmedPath, "/")
	if !found {
		return "", "", fmt.Errorf("failed to parse path %s", path)
	}
	return bucket, glob, nil
}

func getAwsSessionConfig(conf *s3Config) (*session.Session, error) {
	if conf.AWSRegion != "" {
		return session.NewSession(&aws.Config{
			Region: aws.String(conf.AWSRegion),
		})
	}
	return session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
}
