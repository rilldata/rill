package s3

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/blob"
	"github.com/rilldata/rill/runtime/pkg/globutil"
	"gocloud.dev/blob/s3blob"
)

// ListObjects implements drivers.ObjectStore.
func (c *Connection) ListObjects(ctx context.Context, path string) ([]drivers.ObjectStoreEntry, error) {
	url, err := c.parseBucketURL(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %q: %w", path, err)
	}

	bucket, err := c.openBucket(ctx, url.Host, false)
	if err != nil {
		return nil, err
	}
	defer bucket.Close()

	return bucket.ListObjects(ctx, url.Path)
}

// DownloadFiles implements drivers.ObjectStore.
func (c *Connection) DownloadFiles(ctx context.Context, path string) (drivers.FileIterator, error) {
	url, err := c.parseBucketURL(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %q: %w", path, err)
	}

	bucket, err := c.openBucket(ctx, url.Host, false)
	if err != nil {
		return nil, err
	}

	tempDir, err := c.storage.TempDir()
	if err != nil {
		return nil, err
	}

	return bucket.Download(ctx, &blob.DownloadOptions{
		Glob:        url.Path,
		TempDir:     tempDir,
		CloseBucket: true,
	})
}

// BucketRegion returns the region to use for the given bucket.
func (c *Connection) BucketRegion(ctx context.Context, bucket string) (string, error) {
	creds, err := c.newCredentials()
	if err != nil {
		return "", err
	}

	sess, err := c.newSessionForBucket(ctx, bucket, c.config.Endpoint, c.config.Region, creds)
	if err != nil {
		return "", err
	}

	if sess.Config.Region != nil {
		return *sess.Config.Region, nil
	}
	return "", fmt.Errorf("unable to get region")
}

func (c *Connection) parseBucketURL(path string) (*globutil.URL, error) {
	path = c.rewriteToS3Path(path)
	url, err := globutil.ParseBucketURL(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %q: %w", path, err)
	}
	if url.Scheme != "s3" {
		return nil, fmt.Errorf("invalid S3 path %q: should start with s3://", path)
	}
	return url, nil
}

func (c *Connection) openBucket(ctx context.Context, bucket string, anonymous bool) (*blob.Bucket, error) {
	var creds *credentials.Credentials
	if anonymous {
		creds = credentials.AnonymousCredentials
	} else {
		var err error
		creds, err = c.newCredentials()
		if err != nil {
			return nil, fmt.Errorf("failed to create AWS credentials: %w", err)
		}
	}

	sess, err := c.newSessionForBucket(ctx, bucket, c.config.Endpoint, c.config.Region, creds)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	s3Bucket, err := s3blob.OpenBucket(ctx, sess, bucket, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open bucket %q: %w", bucket, err)
	}

	return blob.NewBucket(s3Bucket, c.logger)
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

func (c *Connection) rewriteToS3Path(s string) string {
	switch c.config.Endpoint {
	case "storage.googleapis.com":
		if after, ok := strings.CutPrefix(s, "gs://"); ok {
			return "s3://" + after
		}
		if after, ok := strings.CutPrefix(s, "gcs://"); ok {
			return "s3://" + after
		}
		return s
	default:
		return s
	}
}
