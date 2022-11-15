package drivers

import (
	"context"
	"io"
	"time"
)

// RepoStore is implemented by drivers capable of storing SQL file artifacts
type RepoStore interface {
	ListRecursive(ctx context.Context, repoID string) ([]string, error)
	Get(ctx context.Context, repoID string, path string) (string, error)
	Stat(ctx context.Context, repoID string, path string) (*RepoObjectStat, error)
	PutBlob(ctx context.Context, repoID string, path string, blob string) error
	PutReader(ctx context.Context, repoID string, path string, reader io.Reader) (string, error)
	Rename(ctx context.Context, repoID string, from string, path string) error
	Delete(ctx context.Context, repoID string, path string) error
}

type RepoObjectStat struct {
	LastUpdated time.Time
}
