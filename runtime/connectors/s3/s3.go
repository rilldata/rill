package s3

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/connectors"
)

func init() {
	connectors.Register("s3", connector{})
}

var spec = connectors.Spec{
	DisplayName: "S3",
	Description: "Connect to CSV or Parquet files in an Amazon S3 bucket. For private buckets, provide an <a href=https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html target='_blank'>access key</a>.",
	Properties: []connectors.PropertySchema{
		{
			Key:         "path",
			DisplayName: "S3 URI",
			Description: "Path to file on the disk.",
			Placeholder: "s3://bucket-name/path/to/file.csv",
			Type:        connectors.StringPropertyType,
			Required:    true,
			Hint:        "Tip: use glob patterns to select multiple files",
		},
		{
			Key:         "aws.region",
			DisplayName: "AWS region",
			Description: "AWS Region for the bucket.",
			Placeholder: "us-east-1",
			Type:        connectors.StringPropertyType,
			Required:    true,
		},
		{
			Key:         "aws.access.key",
			DisplayName: "AWS access key",
			Description: "",
			Placeholder: "...",
			Type:        connectors.StringPropertyType,
			Required:    false,
		},
		{
			Key:         "aws.access.secret",
			DisplayName: "AWS access secret",
			Description: "",
			Placeholder: "...",
			Type:        connectors.StringPropertyType,
			Required:    false,
		},
	},
}

type Config struct {
	Path       string `mapstructure:"path" ignored:"true"`
	AWSRegion  string `mapstructure:"aws.region" envconfig:"AWS_DEFAULT_REGION"`
	AWSKey     string `mapstructure:"aws.access.key" envconfig:"AWS_ACCESS_KEY_ID"`
	AWSSecret  string `mapstructure:"aws.access.secret" envconfig:"AWS_SECRET_ACCESS_KEY"`
	AWSSession string `mapstructure:"aws.access.session" ignored:"true"`
}

func ParseConfig(props map[string]any) (*Config, error) {
	conf := &Config{}
	// will be needed when we leverage duckdb's s3 import
	//err := envconfig.Process("aws", conf)
	//if err != nil {
	//	return nil, err
	//}
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

func (c connector) ConsumeAsFile(ctx context.Context, source *connectors.Source, callback func(filename string) error) error {
	conf, err := ParseConfig(source.Properties)
	if err != nil {
		return fmt.Errorf("failed to parse config: %v", err)
	}

	// The session the S3 Downloader will use
	sess, err := session.NewSession(&aws.Config{
		Region:      &conf.AWSRegion,
		Credentials: getAwsCredentials(conf),
	})
	if err != nil {
		return fmt.Errorf("failed to start session: %v", err)
	}

	// Create a downloader with the session and default options
	downloader := s3manager.NewDownloader(sess)

	bucket, key, extension, err := getAwsUrlParts(conf.Path)
	if err != nil {
		return fmt.Errorf("failed to parse path %s, %v", conf.Path, err)
	}

	f, err := os.CreateTemp(
		os.TempDir(),
		fmt.Sprintf("%s*.%s", source.Name, extension),
	)
	if err != nil {
		return fmt.Errorf("os.Create: %v", err)
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	// Write the contents of S3 Object to the f
	_, err = downloader.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to download f, %v", err)
	}

	err = callback(f.Name())
	if err != nil {
		return fmt.Errorf("failed to ingest f, %v", err)
	}

	return nil
}

func getAwsCredentials(conf *Config) *credentials.Credentials {
	if conf.AWSSession != "" {
		return credentials.NewStaticCredentialsFromCreds(credentials.Value{
			SessionToken: conf.AWSSession,
		})
	} else if conf.AWSKey != "" && conf.AWSSecret != "" {
		return credentials.NewStaticCredentialsFromCreds(credentials.Value{
			AccessKeyID:     conf.AWSKey,
			SecretAccessKey: conf.AWSSecret,
		})
	}
	return nil
}

func getAwsUrlParts(path string) (string, string, string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", "", "", err
	}

	p := strings.Split(u.Path, ".")

	return u.Host, u.Path, p[len(p)-1], nil
}
