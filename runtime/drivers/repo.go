package drivers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// RepoStore is implemented by drivers capable of storing code artifacts.
// It mirrors a file system, but may be virtualized by a database for non-local deployments.
type RepoStore interface {
	Driver() string
	// Root returns directory where artifacts are stored.
	Root(ctx context.Context) (string, error)
	CommitHash(ctx context.Context) (string, error)
	ListRecursive(ctx context.Context, glob string, skipDirs bool) ([]DirEntry, error)
	Get(ctx context.Context, path string) (string, error)
	FileHash(ctx context.Context, paths []string) (string, error)
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

// RepoListLimit is the maximum number of files that can be listed in a call to RepoStore.ListRecursive.
// This limit is effectively a cap on the number of files in a project because `rill start` lists the project directory using a "**" glob.
const RepoListLimit = 2000

// ErrRepoListLimitExceeded should be returned when RepoListLimit is exceeded.
var ErrRepoListLimitExceeded = fmt.Errorf("glob exceeded limit of %d matched files", RepoListLimit)

// ignoredPaths is a list of paths that are ignored by the parser.
var ignoredPaths = []string{
	"/tmp",
	"/.git",
	"/node_modules",
	"/.DS_Store",
	"/.vscode",
	"/.idea",
	"/.rillcloud",
}

// IsIgnored returns true if the path (and any files in nested directories) should be ignored.
func IsIgnored(path string, ignorePathsConfig []string) bool {
	for _, dir := range ignoredPaths {
		if path == dir {
			return true
		}
		if strings.HasPrefix(path, dir) && path[len(dir)] == '/' {
			return true
		}
	}
	for _, dir := range ignorePathsConfig {
		if path == dir {
			return true
		}
		if strings.HasPrefix(path, dir) && path[len(dir)] == '/' {
			return true
		}
	}
	return false
}
