package gcs

import (
	"context"

	"cloud.google.com/go/storage"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"gocloud.dev/blob"
	"gocloud.dev/blob/gcsblob"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const defaultPageSize = 20

func (c *Connection) ListObjectsRaw(ctx context.Context, req *runtimev1.ListObjectsRequest) ([]*runtimev1.Object, string, error) {
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

	gcsObjects := make([]*runtimev1.Object, len(objects))
	for i, object := range objects {
		gcsObjects[i] = &runtimev1.Object{
			Name:       object.Key,
			ModifiedOn: timestamppb.New(object.ModTime),
			Size:       object.Size,
			IsDir:      object.IsDir,
		}
	}
	return gcsObjects, string(nextToken), nil
}
