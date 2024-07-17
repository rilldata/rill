package gcs

import (
	"context"
	"errors"

	"cloud.google.com/go/storage"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/gcputil"
	"gocloud.dev/blob"
	"gocloud.dev/blob/gcsblob"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const defaultPageSize = 20

func (c *Connection) ListBuckets(ctx context.Context, req *runtimev1.GCSListBucketsRequest) ([]string, string, error) {
	credentials, err := gcputil.Credentials(ctx, c.config.SecretJSON, c.config.AllowHostAccess)
	if err != nil {
		return nil, "", err
	}

	client, err := storage.NewClient(ctx, option.WithCredentials(credentials))
	if err != nil {
		return nil, "", err
	}
	defer client.Close()

	projectID, err := gcputil.ProjectID(credentials)
	if err != nil {
		return nil, "", err
	}

	pageSize := int(req.PageSize)
	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	pager := iterator.NewPager(client.Buckets(ctx, projectID), pageSize, req.PageToken)
	buckets := make([]*storage.BucketAttrs, 0)
	next, err := pager.NextPage(&buckets)
	if err != nil {
		return nil, "", err
	}

	names := make([]string, len(buckets))
	for i := 0; i < len(buckets); i++ {
		names[i] = buckets[i].Name
	}
	return names, next, nil
}

func (c *Connection) ListObjectsRaw(ctx context.Context, req *runtimev1.GCSListObjectsRequest) ([]*runtimev1.GCSObject, string, error) {
	client, err := c.newClient(ctx)
	if err != nil {
		return nil, "", err
	}

	bucket, err := gcsblob.OpenBucket(ctx, client, req.Bucket, nil)
	if err != nil {
		return nil, "", err
	}
	defer bucket.Close()

	pageSize := int(req.PageSize)
	if pageSize == 0 {
		pageSize = defaultPageSize
	}

	var pageToken []byte
	if req.PageToken == "" {
		pageToken = blob.FirstPageToken
	} else {
		pageToken = []byte(req.PageToken)
	}

	objects, nextToken, err := bucket.ListPage(ctx, pageToken, pageSize, &blob.ListOptions{
		Prefix:    req.Prefix,
		Delimiter: req.Delimiter,
		BeforeList: func(as func(interface{}) bool) error {
			var q *storage.Query
			if as(&q) {
				q.StartOffset = req.StartOffset
				q.EndOffset = req.EndOffset
			} else {
				panic("Listobjects failed")
			}
			return nil
		},
	})
	if err != nil {
		return nil, "", err
	}

	gcsObjects := make([]*runtimev1.GCSObject, len(objects))
	for i, object := range objects {
		gcsObjects[i] = &runtimev1.GCSObject{
			Name:       object.Key,
			ModifiedOn: timestamppb.New(object.ModTime),
			Size:       object.Size,
			IsDir:      object.IsDir,
		}
	}
	return gcsObjects, string(nextToken), nil
}

func (c *Connection) GetCredentialsInfo(ctx context.Context) (string, bool, error) {
	creds, err := gcputil.Credentials(ctx, c.config.SecretJSON, c.config.AllowHostAccess)
	if err != nil {
		if errors.Is(err, gcputil.ErrNoCredentials) {
			return "", false, nil
		}
		return "", false, err
	}

	projectID, err := gcputil.ProjectID(creds)
	return projectID, err == nil, err
}
