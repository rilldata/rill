package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
)

const ListBucketObjectsName = "list_bucket_objects"

type ListBucketObjects struct {
	Runtime *runtime.Runtime
}

var _ Tool[*ListBucketObjectsArgs, *ListBucketObjectsResult] = (*ListBucketObjects)(nil)

type ListBucketObjectsArgs struct {
	Connector string `json:"connector" jsonschema:"The name of an object store connector (e.g., s3, gcs, azure)."`
	Bucket    string `json:"bucket" jsonschema:"The bucket name to list objects from."`
	Path      string `json:"path,omitempty" jsonschema:"Optional path prefix to list objects under. Defaults to root."`
	PageSize  int    `json:"page_size,omitempty" jsonschema:"Maximum number of objects to return. Defaults to 10."`
	PageToken string `json:"page_token,omitempty" jsonschema:"Token for pagination. Use next_page_token from previous response to get next page."`
}

type ListBucketObjectsResult struct {
	Objects       []ObjectInfo `json:"objects"`
	NextPageToken string       `json:"next_page_token,omitempty"`
}

type ObjectInfo struct {
	Path       string     `json:"path"`
	IsDir      bool       `json:"is_dir"`
	Size       int64      `json:"size,omitempty"`
	ModifiedOn *time.Time `json:"modified_on,omitempty"`
}

func (t *ListBucketObjects) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        ListBucketObjectsName,
		Title:       "List Bucket Objects",
		Description: "List objects (files and directories) in a bucket from an object store connector.",
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Listing bucket objects...",
			"openai/toolInvocation/invoked":  "Listed bucket objects",
		},
	}
}

func (t *ListBucketObjects) CheckAccess(ctx context.Context) (bool, error) {
	return checkDeveloperAgentAccess(ctx, t.Runtime)
}

func (t *ListBucketObjects) Handler(ctx context.Context, args *ListBucketObjectsArgs) (*ListBucketObjectsResult, error) {
	if args.Connector == "" {
		return nil, fmt.Errorf("connector name is required")
	}
	if args.Bucket == "" {
		return nil, fmt.Errorf("bucket name is required")
	}

	s := GetSession(ctx)

	// Acquire handle for the connector
	handle, release, err := t.Runtime.AcquireHandle(ctx, s.InstanceID(), args.Connector)
	if err != nil {
		return nil, err
	}
	defer release()

	// Cast to object store
	os, ok := handle.AsObjectStore()
	if !ok {
		return nil, fmt.Errorf("connector %q does not implement object store", args.Connector)
	}

	// Apply defaults
	pageSize := args.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	// List objects
	objects, nextToken, err := os.ListObjects(ctx, args.Bucket, args.Path, "/", uint32(pageSize), args.PageToken)
	if err != nil {
		return nil, err
	}

	// Convert to result format
	result := &ListBucketObjectsResult{
		Objects:       make([]ObjectInfo, 0, len(objects)),
		NextPageToken: nextToken,
	}
	for _, obj := range objects {
		info := ObjectInfo{
			Path:  obj.Path,
			IsDir: obj.IsDir,
			Size:  obj.Size,
		}
		if !obj.UpdatedOn.IsZero() {
			info.ModifiedOn = &obj.UpdatedOn
		}
		result.Objects = append(result.Objects, info)
	}

	return result, nil
}
