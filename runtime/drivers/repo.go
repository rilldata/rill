package drivers

import (
	"context"
	"io"
)

// RepoStore is implemented by drivers capable of storing SQL file artifacts
type RepoStore interface {
	ListRecursive(ctx context.Context, repoID string) ([]string, error)
	Get(ctx context.Context, repoID string, path string) (string, error)
	PutBlob(ctx context.Context, repoID string, path string, blob string) error
	PutReader(ctx context.Context, repoID string, path string, reader io.Reader) (string, error)
	Delete(ctx context.Context, repoID string, path string) error
}
