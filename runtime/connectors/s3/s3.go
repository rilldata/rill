package s3

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
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

type Config struct {
	Path      string `mapstructure:"path"`
	AWSRegion string `mapstructure:"region"`
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

	// The session the S3 Downloader will use
	sess, err := getAwsSessionConfig(conf)
	if err != nil {
		return "", fmt.Errorf("failed to start session: %w", err)
	}

	// Create a downloader with the session and default options
	downloader := s3manager.NewDownloader(sess)

	bucket, key, extension, err := awsURLParts(conf.Path)
	if err != nil {
		return "", fmt.Errorf("failed to parse path %s, %w", conf.Path, err)
	}

	f, err := os.CreateTemp(
		os.TempDir(),
		fmt.Sprintf("%s*%s", source.Name, extension),
	)
	if err != nil {
		return "", fmt.Errorf("os.Create: %w", err)
	}
	defer f.Close()

	// Write the contents of S3 Object to the f
	_, err = downloader.DownloadWithContext(ctx, f, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		os.Remove(f.Name())
		return "", fmt.Errorf("failed to download f, %w", err)
	}

	return f.Name(), nil
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

func awsURLParts(path string) (string, string, string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", "", "", err
	}
	return u.Host, u.Path, fileutil.FullExt(u.Path), nil
}
