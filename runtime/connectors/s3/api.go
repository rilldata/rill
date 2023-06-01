package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/connectors"
	"gocloud.dev/blob"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c Connector) ListBuckets(ctx context.Context, env *connectors.Env) ([]string, error) {
	creds, err := getCredentials(env)
	if err != nil {
		return nil, err
	}
	if creds == credentials.AnonymousCredentials {
		return nil, fmt.Errorf("no credentials exist")
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

	// no region found, default to us-east-1
	if sess.Config.Region == nil || *sess.Config.Region == "" {
		sess = sess.Copy(&aws.Config{Region: aws.String("us-east-1")})
	}
	svc := s3.New(sess)
	output, err := svc.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}

	buckets := make([]string, 0, len(output.Buckets))
	for _, bucket := range output.Buckets {
		if bucket.Name != nil {
			buckets = append(buckets, *bucket.Name)
		}
	}
	return buckets, nil
}

func (c Connector) ListObjects(ctx context.Context, req *runtimev1.S3ListObjectsRequest, env *connectors.Env) ([]*runtimev1.S3Object, string, error) {
	// todo :: check for cases when accessing public buckets but env configured
	creds, err := getCredentials(env)
	if err != nil {
		return nil, "", err
	}

	bucket, err := openBucket(ctx, &Config{AWSRegion: req.Region}, req.Bucket, creds)
	if err != nil {
		return nil, "", err
	}
	defer bucket.Close()

	var pageToken []byte
	if req.PageToken == "" {
		pageToken = blob.FirstPageToken
	} else {
		pageToken = []byte(req.PageToken)
	}

	pageSize := int(req.PageSize)
	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	objects, nextToken, err := bucket.ListPage(ctx, pageToken, pageSize, &blob.ListOptions{
		Prefix:    req.Prefix,
		Delimiter: req.Delimiter,
		BeforeList: func(as func(interface{}) bool) error {
			if req.StartAfter == "" {
				return nil
			}
			var q *s3.ListObjectsV2Input
			if as(&q) {
				q.StartAfter = &req.StartAfter
			}
			return nil
		},
	})
	if err != nil {
		return nil, "", err
	}

	s3Objects := make([]*runtimev1.S3Object, len(objects))
	for i, object := range objects {
		s3Objects[i] = &runtimev1.S3Object{
			Name:       object.Key,
			ModifiedOn: timestamppb.New(object.ModTime),
			Size:       object.Size,
			IsDir:      object.IsDir,
		}
	}
	return s3Objects, string(nextToken), nil
}

func (c Connector) GetBucketMetadata(ctx context.Context, req *runtimev1.S3GetBucketMetadataRequest, env *connectors.Env) (string, error) {
	creds, err := getCredentials(env)
	if err != nil {
		return "", err
	}

	sess, err := getAwsSessionConfig(ctx, &Config{}, req.Bucket, creds)
	if err != nil {
		return "", err
	}

	if sess.Config.Region != nil {
		return *sess.Config.Region, nil
	}
	return "", fmt.Errorf("unable to get region")
}
