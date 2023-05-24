package s3

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/connectors"
	rillblob "github.com/rilldata/rill/runtime/connectors/blob"
	"github.com/rilldata/rill/runtime/pkg/globutil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
	"gocloud.dev/gcerrors"
)

func init() {
	connectors.Register("s3", connector{})
}

var spec = connectors.Spec{
	DisplayName:        "Amazon S3",
	Description:        "Connect to AWS S3 Storage.",
	ServiceAccountDocs: "https://docs.rilldata.com/connectors/s3",
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
			Href:        "https://docs.rilldata.com/develop/import-data#setting-local-credentials-for-s3",
		},
	},
	ConnectorVariables: []connectors.VariableSchema{
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

type Config struct {
	Path                  string `mapstructure:"path"`
	AWSRegion             string `mapstructure:"region"`
	GlobMaxTotalSize      int64  `mapstructure:"glob.max_total_size"`
	GlobMaxObjectsMatched int    `mapstructure:"glob.max_objects_matched"`
	GlobMaxObjectsListed  int64  `mapstructure:"glob.max_objects_listed"`
	GlobPageSize          int    `mapstructure:"glob.page_size"`
	S3Endpoint            string `mapstructure:"endpoint"`
	url                   *globutil.URL
}

func ParseConfig(props map[string]any) (*Config, error) {
	conf := &Config{}
	err := mapstructure.Decode(props, conf)
	if err != nil {
		return nil, err
	}

	if !doublestar.ValidatePattern(conf.Path) {
		return nil, fmt.Errorf("glob pattern %s is invalid", conf.Path)
	}

	url, err := globutil.ParseBucketURL(conf.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %q, %w", conf.Path, err)
	}

	if url.Scheme != "s3" {
		return nil, fmt.Errorf("invalid s3 path %q, should start with s3://", conf.Path)
	}
	conf.url = url
	return conf, nil
}

type connector struct{}

func (c connector) Spec() connectors.Spec {
	return spec
}

// ConsumeAsIterator returns a file iterator over objects stored in gcs.
//
// The credentials are read from following env variables
//   - AWS_ACCESS_KEY_ID
//   - AWS_SECRET_ACCESS_KEY
//   - AWS_SESSION_TOKEN
//
// Additionally in case env.AllowHostCredentials is true it looks for credentials stored on host machine as well
func (c connector) ConsumeAsIterator(ctx context.Context, env *connectors.Env, source *connectors.Source, logger *zap.Logger) (connectors.FileIterator, error) {
	conf, err := ParseConfig(source.Properties)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	creds, err := getCredentials(env)
	if err != nil {
		return nil, err
	}

	bucketObj, err := openBucket(ctx, conf, conf.url.Host, creds)
	if err != nil {
		return nil, fmt.Errorf("failed to open bucket %q, %w", conf.url.Host, err)
	}

	// prepare fetch configs
	opts := rillblob.Options{
		GlobMaxTotalSize:      conf.GlobMaxTotalSize,
		GlobMaxObjectsMatched: conf.GlobMaxObjectsMatched,
		GlobMaxObjectsListed:  conf.GlobMaxObjectsListed,
		GlobPageSize:          conf.GlobPageSize,
		GlobPattern:           conf.url.Path,
		ExtractPolicy:         source.ExtractPolicy,
		StorageLimitInBytes:   env.StorageLimitInBytes,
	}

	it, err := rillblob.NewIterator(ctx, bucketObj, opts, logger)
	if err != nil {
		// s3 throws error (possibly inconsistent) in case we are trying to access public buckets and passing some credentials
		// go cdk wraps some s3's errors into gcerrors.Unknown
		// we try again with anonymous credentials in case bucket is public
		errCode := gcerrors.Code(err)
		if (errCode == gcerrors.PermissionDenied || errCode == gcerrors.Unknown) && creds != credentials.AnonymousCredentials {
			logger.Info("s3 list objects failed, re-trying with anonymous credential", zap.Error(err), observability.ZapCtx(ctx))
			creds = credentials.AnonymousCredentials
			bucketObj, bucketErr := openBucket(ctx, conf, conf.url.Host, creds)
			if bucketErr != nil {
				return nil, fmt.Errorf("failed to open bucket %q, %w", conf.url.Host, bucketErr)
			}
			it, err = rillblob.NewIterator(ctx, bucketObj, opts, logger)
		}

		// aws returns StatusForbidden in cases like no creds passed, wrong creds passed and incorrect bucket
		// r2 returns StatusBadRequest in all cases above
		var failureErr awserr.RequestFailure
		if errors.As(err, &failureErr) && (failureErr.StatusCode() == http.StatusForbidden || failureErr.StatusCode() == http.StatusBadRequest) {
			return nil, connectors.NewPermissionDeniedError(fmt.Sprintf("can't access remote source %q err: %v", source.Name, failureErr))
		}
	}

	return it, err
}

func (c connector) HasAnonymousAccess(ctx context.Context, env *connectors.Env, source *connectors.Source) (bool, error) {
	conf, err := ParseConfig(source.Properties)
	if err != nil {
		return false, fmt.Errorf("failed to parse config: %w", err)
	}

	creds, err := getCredentials(env)
	if err != nil {
		return false, err
	}

	bucketObj, err := openBucket(ctx, conf, conf.url.Host, creds)
	if err != nil {
		return false, fmt.Errorf("failed to open bucket %q, %w", conf.url.Host, err)
	}

	return bucketObj.IsAccessible(ctx)
}

func openBucket(ctx context.Context, conf *Config, bucket string, creds *credentials.Credentials) (*blob.Bucket, error) {
	sess, err := getAwsSessionConfig(ctx, conf, bucket, creds)
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}

	return s3blob.OpenBucket(ctx, sess, bucket, nil)
}

func getAwsSessionConfig(ctx context.Context, conf *Config, bucket string, creds *credentials.Credentials) (*session.Session, error) {
	// If S3Endpoint is set, we assume we're targeting an S3 compatible API (but not AWS)
	if len(conf.S3Endpoint) > 0 {
		region := conf.AWSRegion
		if region == "" {
			// Set the default region for bwd compatibility reasons
			// cloudflare and minio ignore if us-east-1 is set, not tested for others
			region = "us-east-1"
		}
		return session.NewSession(&aws.Config{
			Region:           aws.String(region),
			Endpoint:         &conf.S3Endpoint,
			S3ForcePathStyle: aws.Bool(true),
			Credentials:      creds,
		})
	}
	// The logic below is AWS-specific, so we ignore it when conf.S3Endpoint is set
	// The complexity below relates to AWS being pretty strict about regions (probably to avoid unexpected cross-region traffic).

	// If the user explicitly set a region, we use that
	if conf.AWSRegion != "" {
		return session.NewSession(&aws.Config{
			Region:      aws.String(conf.AWSRegion),
			Credentials: creds,
		})
	}

	// Create a session that tries to infer the region from the environment
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable, // Tells to look for default region set with `aws configure`
		Config: aws.Config{
			Credentials: creds,
		},
	})
	if err != nil {
		return nil, err
	}

	// If no region was found, we default to us-east-1 (which will be used to resolve the lookup in the next step)
	if sess.Config.Region == nil || *sess.Config.Region == "" {
		sess = sess.Copy(&aws.Config{Region: aws.String("us-east-1")})
	}

	// Bucket names are globally unique, but requests will fail if their region doesn't match the one configured in the session.
	// So we do a lookup for the bucket's region and configure the session to use that.
	reg, err := s3manager.GetBucketRegion(ctx, sess, bucket, "")
	if err != nil {
		return nil, err
	}
	if reg != "" {
		sess = sess.Copy(&aws.Config{Region: aws.String(reg)})
	}

	return sess, nil
}

func getCredentials(env *connectors.Env) (*credentials.Credentials, error) {
	providers := make([]credentials.Provider, 0)

	staticProvider := &credentials.StaticProvider{}
	staticProvider.AccessKeyID = env.Variables["AWS_ACCESS_KEY_ID"]
	staticProvider.SecretAccessKey = env.Variables["AWS_SECRET_ACCESS_KEY"]
	staticProvider.SessionToken = env.Variables["AWS_SESSION_TOKEN"]
	staticProvider.ProviderName = credentials.StaticProviderName
	// in case user doesn't set access key id and secret access key the credentials retreival will fail
	// the credential lookup will proceed to next provider in chain
	providers = append(providers, staticProvider)

	if env.AllowHostAccess {
		// allowed to access host credentials so we add them in chain
		// The chain used here is a duplicate of defaults.CredProviders(), but without the remote credentials lookup (since they resolve too slowly).
		providers = append(providers, &credentials.EnvProvider{}, &credentials.SharedCredentialsProvider{Filename: "", Profile: ""})
	}
	// Find credentials to use.
	creds := credentials.NewChainCredentials(providers)
	if _, err := creds.Get(); err != nil {
		if !errors.Is(err, credentials.ErrNoValidProvidersFoundInChain) {
			return nil, err
		}
		// If no local credentials are found, you must explicitly set AnonymousCredentials to fetch public objects.
		// AnonymousCredentials can't be chained, so we try to resolve local creds, and use anon if none were found.
		creds = credentials.AnonymousCredentials
	}

	return creds, nil
}
