package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/rilldata/rill/runtime/drivers"
)

func init() {
	drivers.RegisterConnector("s3", &Connection{})
}

var spec = drivers.Spec{
	DisplayName:        "Amazon S3",
	Description:        "Connect to AWS S3 Storage.",
	ServiceAccountDocs: "https://docs.rilldata.com/deploy/credentials/s3",
	Properties: []drivers.PropertySchema{
		{
			Key:         "path",
			DisplayName: "S3 URI",
			Description: "Path to file on the disk.",
			Placeholder: "s3://bucket-name/path/to/file.csv",
			Type:        drivers.StringPropertyType,
			Required:    true,
			Hint:        "Glob patterns are supported",
		},
		{
			Key:         "region",
			DisplayName: "AWS region",
			Description: "AWS Region for the bucket.",
			Placeholder: "us-east-1",
			Type:        drivers.StringPropertyType,
			Required:    false,
			Hint:        "Rill will use the default region in your local AWS config, unless set here.",
		},
		{
			Key:         "aws.credentials",
			DisplayName: "AWS credentials",
			Description: "AWS credentials inferred from your local environment.",
			Type:        drivers.InformationalPropertyType,
			Hint:        "Set your local credentials: <code>aws configure</code> Click to learn more.",
			Href:        "https://docs.rilldata.com/develop/import-data#configure-credentials-for-s3",
		},
	},
	ConnectorVariables: []drivers.VariableSchema{
		{
			Key:    "aws_access_key_id",
			Secret: true,
		},
		{
			Key:    "aws_secret_access_key",
			Secret: true,
		},
	},
}

func (c *Connection) Spec() drivers.Spec {
	return spec
}

func (c *Connection) HasAnonymousAccess(ctx context.Context, config map[string]any) (bool, error) {
	conf, err := parseConfig(config)
	if err != nil {
		return false, fmt.Errorf("failed to parse config: %w", err)
	}

	bucketObj, err := c.openBucket(ctx, conf, conf.url.Host, credentials.AnonymousCredentials)
	if err != nil {
		return false, fmt.Errorf("failed to open bucket %q, %w", conf.url.Host, err)
	}
	defer bucketObj.Close()

	return bucketObj.IsAccessible(ctx)
}
