package s3

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/c2h5oh/datasize"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	rillblob "github.com/rilldata/rill/runtime/drivers/blob"
	"github.com/rilldata/rill/runtime/pkg/globutil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
)

type sourceProperties struct {
	AWSRegion             string         `mapstructure:"region"`
	S3Endpoint            string         `mapstructure:"endpoint"`
	Path                  string         `mapstructure:"path"`
	URI                   string         `mapstructure:"uri"`
	GlobMaxTotalSize      int64          `mapstructure:"glob.max_total_size"`
	GlobMaxObjectsMatched int            `mapstructure:"glob.max_objects_matched"`
	GlobMaxObjectsListed  int64          `mapstructure:"glob.max_objects_listed"`
	GlobPageSize          int            `mapstructure:"glob.page_size"`
	Extract               map[string]any `mapstructure:"extract"`
	BatchSize             string         `mapstructure:"batch_size"`
	url                   *globutil.URL
	extractPolicy         *rillblob.ExtractPolicy
}

func parseSourceProperties(props map[string]any) (*sourceProperties, error) {
	conf := &sourceProperties{}
	err := mapstructure.WeakDecode(props, conf)
	if err != nil {
		return nil, err
	}

	// Backwards compatibility for "uri" renamed to "path"
	if conf.URI != "" {
		conf.Path = conf.URI
	}

	if !doublestar.ValidatePattern(conf.Path) {
		return nil, fmt.Errorf("glob pattern %s is invalid", conf.Path)
	}

	url, err := globutil.ParseBucketURL(conf.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %q, %w", conf.Path, err)
	}
	conf.url = url

	if url.Scheme != "s3" {
		return nil, fmt.Errorf("invalid s3 path %q, should start with s3://", conf.Path)
	}

	conf.extractPolicy, err = rillblob.ParseExtractPolicy(conf.Extract)
	if err != nil {
		return nil, fmt.Errorf("failed to parse extract config: %w", err)
	}

	return conf, nil
}

// ListObjects implements drivers.ObjectStore.
func (c *Connection) ListObjects(ctx context.Context, propsMap map[string]any) ([]drivers.ObjectStoreEntry, error) {
	props, err := parseSourceProperties(propsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	creds, err := c.newCredentials()
	if err != nil {
		return nil, err
	}

	s3Bucket, err := c.openBucket(ctx, props, props.url.Host, creds)
	if err != nil {
		return nil, fmt.Errorf("failed to open bucket %q: %w", props.url.Host, err)
	}

	bucket, err := rillblob.NewBucket(s3Bucket, c.logger)
	if err != nil {
		return nil, err
	}
	defer bucket.Close()

	return bucket.ListObjects(ctx, props.url.Path)
}

// DownloadFiles implements drivers.ObjectStore.
func (c *Connection) DownloadFiles(ctx context.Context, src map[string]any) (drivers.FileIterator, error) {
	conf, err := parseSourceProperties(src)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	creds, err := c.newCredentials()
	if err != nil {
		return nil, err
	}

	bucketObj, err := c.openBucket(ctx, conf, conf.url.Host, creds)
	if err != nil {
		return nil, fmt.Errorf("failed to open bucket %q, %w", conf.url.Host, err)
	}

	var batchSize datasize.ByteSize
	if conf.BatchSize == "-1" {
		batchSize = math.MaxInt64 // download everything in one batch
	} else {
		batchSize, err = datasize.ParseString(conf.BatchSize)
		if err != nil {
			return nil, err
		}
	}
	// prepare fetch configs
	opts := rillblob.Options{
		GlobMaxTotalSize:      conf.GlobMaxTotalSize,
		GlobMaxObjectsMatched: conf.GlobMaxObjectsMatched,
		GlobMaxObjectsListed:  conf.GlobMaxObjectsListed,
		GlobPageSize:          conf.GlobPageSize,
		GlobPattern:           conf.url.Path,
		ExtractPolicy:         conf.extractPolicy,
		BatchSizeBytes:        int64(batchSize.Bytes()),
		KeepFilesUntilClose:   conf.BatchSize == "-1",
		RetainFiles:           c.config.RetainFiles,
		TempDir:               c.config.TempDir,
	}

	it, err := rillblob.NewIterator(ctx, bucketObj, opts, c.logger)
	if err != nil {
		// TODO :: fix this for single file access. for single file first call only happens during download
		var failureErr awserr.RequestFailure
		if !errors.As(err, &failureErr) {
			return nil, err
		}

		// aws returns StatusForbidden in cases like no creds passed, wrong creds passed and incorrect bucket
		// r2 returns StatusBadRequest in all cases above
		// we try again with anonymous credentials in case bucket is public
		if (failureErr.StatusCode() == http.StatusForbidden || failureErr.StatusCode() == http.StatusBadRequest) && creds != credentials.AnonymousCredentials {
			c.logger.Debug("s3 list objects failed, re-trying with anonymous credential", zap.Error(err), observability.ZapCtx(ctx))
			creds = credentials.AnonymousCredentials
			bucketObj, bucketErr := c.openBucket(ctx, conf, conf.url.Host, creds)
			if bucketErr != nil {
				return nil, fmt.Errorf("failed to open bucket %q, %w", conf.url.Host, bucketErr)
			}

			anonIt, anonErr := rillblob.NewIterator(ctx, bucketObj, opts, c.logger)
			if anonErr == nil {
				return anonIt, nil
			}
		}

		// check again
		if failureErr.StatusCode() == http.StatusForbidden || failureErr.StatusCode() == http.StatusBadRequest {
			return nil, drivers.NewPermissionDeniedError(fmt.Sprintf("can't access remote err: %v", failureErr))
		}
		return nil, err
	}

	return it, err
}

func (c *Connection) openBucket(ctx context.Context, conf *sourceProperties, bucket string, creds *credentials.Credentials) (*blob.Bucket, error) {
	sess, err := c.newSessionForBucket(ctx, bucket, conf.S3Endpoint, conf.AWSRegion, creds)
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}

	return s3blob.OpenBucket(ctx, sess, bucket, nil)
}

func (c *Connection) newSessionForBucket(ctx context.Context, bucket, endpoint, region string, creds *credentials.Credentials) (*session.Session, error) {
	// If S3Endpoint is set, we assume we're targeting an S3 compatible API (but not AWS)
	if endpoint != "" {
		if region == "" {
			// Set the default region for bwd compatibility reasons
			// cloudflare and minio ignore if us-east-1 is set, not tested for others
			region = "us-east-1"
		}
		return session.NewSession(&aws.Config{
			Region:           aws.String(region),
			Endpoint:         &endpoint,
			S3ForcePathStyle: aws.Bool(true),
			Credentials:      creds,
		})
	}
	// The logic below is AWS-specific, so we ignore it when conf.S3Endpoint is set
	// The complexity below relates to AWS being pretty strict about regions (probably to avoid unexpected cross-region traffic).

	// If the user explicitly set a region, we use that
	if region != "" {
		return session.NewSession(&aws.Config{
			Region:      aws.String(region),
			Credentials: creds,
		})
	}

	sharedConfigState := session.SharedConfigDisable
	if c.config.AllowHostAccess {
		sharedConfigState = session.SharedConfigEnable // Tells to look for default region set with `aws configure`
	}

	// Create a session that tries to infer the region from the environment
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: sharedConfigState,
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
