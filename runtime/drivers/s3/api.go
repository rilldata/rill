package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"gocloud.dev/blob"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *Connection) ListObjectsRaw(ctx context.Context, req *runtimev1.ListObjectsRequest) ([]*runtimev1.Object, string, error) {
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

	bucket, err := c.openBucket(ctx, req.Bucket, false)
	if err != nil {
		return nil, "", err
	}
	defer bucket.Close()

	objects, nextToken, err := fetchObjects(ctx, bucket.Underlying(), pageToken, pageSize, req)
	if err != nil {
		if isPermissionError(err) {
			bucket, err = c.openBucket(ctx, req.Bucket, true)
			if err != nil {
				return nil, "", fmt.Errorf("failed to open bucket %q, %w", req.Bucket, err)
			}
			defer bucket.Close()

			objects, nextToken, err = fetchObjects(ctx, bucket.Underlying(), pageToken, pageSize, req)
		}
	}
	if err != nil {
		return nil, "", err
	}

	s3Objects := make([]*runtimev1.Object, len(objects))
	for i, object := range objects {
		s3Objects[i] = &runtimev1.Object{
			Name:       object.Key,
			ModifiedOn: timestamppb.New(object.ModTime),
			Size:       object.Size,
			IsDir:      object.IsDir,
		}
	}
	return s3Objects, string(nextToken), nil
}

func fetchObjects(ctx context.Context, bucket *blob.Bucket, pageToken []byte, pageSize int, req *runtimev1.ListObjectsRequest) ([]*blob.ListObject, []byte, error) {
	objects, nextToken, err := bucket.ListPage(ctx, pageToken, pageSize, &blob.ListOptions{
		Prefix:    req.Prefix,
		Delimiter: req.Delimiter,
		BeforeList: func(as func(interface{}) bool) error {
			if req.StartOffset == "" {
				return nil
			}
			var q *s3.ListObjectsV2Input
			if as(&q) {
				q.StartAfter = &req.StartOffset
			}
			return nil
		},
	})
	return objects, nextToken, err
}

func isPermissionError(err error) bool {
	errStr := err.Error()
	return errStr == "403" ||
		errStr == "Forbidden" ||
		errStr == "Access Denied" ||
		errStr == "400" ||
		errStr == "Bad Request"
}
