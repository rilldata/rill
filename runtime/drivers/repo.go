package drivers

import (
	"context"
	"errors"
	"io"
	"time"
)

// RepoStore is implemented by drivers capable of storing code artifacts.
// It mirrors a file system, but may be virtualized by a database for non-local deployments.
type RepoStore interface {
	Driver() string
	DSN() string
	ListRecursive(ctx context.Context, instID string, glob string) ([]string, error)
	Get(ctx context.Context, instID string, path string) (string, error)
	Stat(ctx context.Context, instID string, path string) (*RepoObjectStat, error)
	PutBlob(ctx context.Context, instID string, path string, blob string) error
	PutReader(ctx context.Context, instID string, path string, reader io.Reader) (string, error)
	Rename(ctx context.Context, instID string, from string, path string) error
	Delete(ctx context.Context, instID string, path string) error
}

type RepoObjectStat struct {
	LastUpdated time.Time
}

var ErrFileAlreadyExists = errors.New("file already exists")
