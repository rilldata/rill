package ai

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
)

const ListBucketsName = "list_buckets"

type ListBuckets struct {
	Runtime *runtime.Runtime
}

var _ Tool[*ListBucketsArgs, *ListBucketsResult] = (*ListBuckets)(nil)

type ListBucketsArgs struct {
	Connector string `json:"connector" jsonschema:"The name of an object store connector (e.g., s3, gcs, azure)."`
}

type ListBucketsResult struct {
	Buckets []string `json:"buckets"`
}

func (t *ListBuckets) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        ListBucketsName,
		Title:       "List Buckets",
		Description: "List buckets available in an object store connector. Note: This is best-effort - bucket listing may not return all accessible buckets depending on cloud provider permissions. There may be access to additional buckets not returned here.",
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Listing buckets...",
			"openai/toolInvocation/invoked":  "Listed buckets",
		},
	}
}

func (t *ListBuckets) CheckAccess(ctx context.Context) (bool, error) {
	return checkDeveloperAccess(ctx, t.Runtime, false)
}

func (t *ListBuckets) Handler(ctx context.Context, args *ListBucketsArgs) (*ListBucketsResult, error) {
	if args.Connector == "" {
		return nil, fmt.Errorf("connector name is required")
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

	// List buckets (collect all pages)
	var allBuckets []string
	pageToken := ""
	for {
		buckets, nextToken, err := os.ListBuckets(ctx, 100, pageToken)
		if err != nil {
			return nil, err
		}
		allBuckets = append(allBuckets, buckets...)
		if nextToken == "" {
			break
		}
		pageToken = nextToken
	}

	return &ListBucketsResult{Buckets: allBuckets}, nil
}
