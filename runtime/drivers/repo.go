package drivers

import (
	"context"
)

// RepoStore is implemented by drivers capable of storing SQL file artifacts
type RepoStore interface {
	ListRecursive(ctx context.Context, repoID string) []string
	Get(ctx context.Context, repoID string, path string) (string, error)
	Put(ctx context.Context, repoID string, path string, blob string) error
	Delete(ctx context.Context, repoID string, path string) error
}
