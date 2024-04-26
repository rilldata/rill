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
	CommitHash(ctx context.Context) (string, error)
	ListRecursive(ctx context.Context, glob string, skipDirs bool) ([]DirEntry, error)
	Get(ctx context.Context, path string) (string, error)
	Stat(ctx context.Context, path string) (*RepoObjectStat, error)
	Put(ctx context.Context, path string, reader io.Reader) error
	MakeDir(ctx context.Context, path string) error
	Rename(ctx context.Context, fromPath string, toPath string) error
	Delete(ctx context.Context, path string, force bool) error
	Sync(ctx context.Context) error
	Watch(ctx context.Context, cb WatchCallback) error
}

type WatchCallback func(event []WatchEvent)

type WatchEvent struct {
	Type runtimev1.FileEvent
	Path string
	Dir  bool
}

type RepoObjectStat struct {
	LastUpdated time.Time
	IsDir       bool
}

var ErrFileAlreadyExists = errors.New("file already exists")

type DirEntry struct {
	Path  string
	IsDir bool
}
