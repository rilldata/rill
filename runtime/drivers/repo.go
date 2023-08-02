package drivers

import (
	"context"
	"errors"
	"io"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// RepoStore is implemented by drivers capable of storing code artifacts.
// It mirrors a file system, but may be virtualized by a database for non-local deployments.
type RepoStore interface {
	Driver() string
	// Root returns directory where artifacts are stored.
	Root() string
	ListRecursive(ctx context.Context, instID string, glob string) ([]string, error)
	Get(ctx context.Context, instID string, path string) (string, error)
	Stat(ctx context.Context, instID string, path string) (*RepoObjectStat, error)
	Put(ctx context.Context, instID string, path string, reader io.Reader) error
	Rename(ctx context.Context, instID string, fromPath string, toPath string) error
	Delete(ctx context.Context, instID string, path string) error
	Sync(ctx context.Context, instID string) error
	Watch(ctx context.Context, instID string, cb WatchCallback) error
}

type WatchCallback func(event []WatchEvent)

type WatchEvent struct {
	Type runtimev1.FileEvent
	Path string
	Dir  bool
}

type RepoObjectStat struct {
	LastUpdated time.Time
}

var ErrFileAlreadyExists = errors.New("file already exists")
