package gcs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"cloud.google.com/go/storage"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/connectors"
	"gocloud.dev/blob"
	"gocloud.dev/blob/gcsblob"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c Connector) ListBuckets(ctx context.Context, req *runtimev1.GCSListBucketsRequest, env *connectors.Env) ([]string, string, error) {
	credentials, err := resolvedCredentials(ctx, env)
	if err != nil {
		return nil, "", err
	}

	client, err := storage.NewClient(ctx, option.WithCredentials(credentials))
	if err != nil {
		return nil, "", err
	}
	defer client.Close()

	projectID, err := getProjectID(credentials)
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

func (c Connector) ListObjects(ctx context.Context, req *runtimev1.GCSListObjectsRequest, env *connectors.Env) ([]*runtimev1.GCSObject, string, error) {
	client, err := createClient(ctx, env)
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

func (c Connector) GetCredentialsInfo(ctx context.Context, env *connectors.Env) (string, bool, error) {
	creds, err := resolvedCredentials(ctx, env)
	if err != nil {
		if errors.Is(err, errNoCredentials) {
			return "", false, nil
		}
		return "", false, err
	}

	projectID, err := getProjectID(creds)
	return projectID, err == nil, err
}

func getProjectID(credentials *google.Credentials) (string, error) {
	projectID := credentials.ProjectID
	if projectID == "" {
		if len(credentials.JSON) == 0 {
			return "", fmt.Errorf("unable to get project ID")
		}
		f := &credentialsFile{}
		if err := json.Unmarshal(credentials.JSON, f); err != nil {
			return "", err
		}

		projectID = f.getProjectID()
	}
	return projectID, nil
}

// credentialsFile is the unmarshalled representation of a credentials file.
type credentialsFile struct {
	Type string `json:"type"`

	// Service Account fields
	ProjectID string `json:"project_id"`

	// External Account fields
	QuotaProjectID string `json:"quota_project_id"`

	// Service account impersonation
	SourceCredentials *credentialsFile `json:"source_credentials"`
}

func (c *credentialsFile) getProjectID() string {
	if c.Type == "impersonated_service_account" {
		return c.SourceCredentials.getProjectID()
	}
	if c.ProjectID != "" {
		return c.ProjectID
	}
	return c.QuotaProjectID
}
