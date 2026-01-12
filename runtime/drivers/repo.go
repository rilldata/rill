package drivers

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

var ErrRemoteAhead = fmt.Errorf("remote ahead of local state, please pull first")

// RepoStore is implemented by drivers capable of storing project code files.
// All paths start with '/' and are relative to the repo root.
type RepoStore interface {
	// Root returns an absolute path to a physical directory where files are stored.
	// It's provided as an escape hatch, but should not generally be used as it does not have consistency guarantees.
	Root(ctx context.Context) (string, error)
	// ListGlob lists all files in the repo matching the glob pattern.
	ListGlob(ctx context.Context, glob string, skipDirs bool) ([]DirEntry, error)
	// Get retrieves the content of a file at the specified path.
	Get(ctx context.Context, path string) (string, error)
	// Hash returns a unique hash of the contents of the files at the specified paths.
	// If a file does not exist, it is skipped (does not return an error).
	Hash(ctx context.Context, paths []string) (string, error)
	// Stat returns metadata about a file.
	Stat(ctx context.Context, path string) (*FileInfo, error)
	// Put creates or overwrites a file.
	Put(ctx context.Context, path string, reader io.Reader) error
	// MkdirAll creates a directory and any required parent directories.
	MkdirAll(ctx context.Context, path string) error
	// Rename moves a file from one path to another.
	Rename(ctx context.Context, fromPath string, toPath string) error
	// Delete removes a file or directory.
	// If force is true, it will delete non-empty directories.
	Delete(ctx context.Context, path string, force bool) error
	// Watch sets up a watcher for changes in the repo.
	// The callback will be called with events for changes in the repo.
	// The function does not return until the context is cancelled or an error occurs.
	Watch(ctx context.Context, cb WatchCallback) error

	// Status returns the current status of the repository.
	Status(ctx context.Context) (*RepoStatus, error)
	// Pull synchronizes local and remote state.
	// If discardChanges is true, it will discard any local changes made using Put/Rename/etc. and force synchronize to the remote state.
	// If forceHandshake is true, it will re-verify any cached config. Specifically, this should be used when external config changes, such as the Git branch or file archive ID.
	Pull(ctx context.Context, opts *PullOptions) error
	// CommitAndPush commits local changes to the remote repository and pushes them.
	CommitAndPush(ctx context.Context, message string, force bool) error
	// CommitHash returns a unique ID for the state of the remote files currently served (does not change on uncommitted local changes).
	CommitHash(ctx context.Context) (string, error)
	// CommitTimestamp returns the update timestamp for the current remote files (does not change on uncommitted local changes).
	CommitTimestamp(ctx context.Context) (time.Time, error)
}

// FileInfo contains metadata about a file.
type FileInfo struct {
	LastUpdated time.Time
	IsDir       bool
}

// DirEntry represents an entry in a directory listing.
type DirEntry struct {
	Path  string
	IsDir bool
}

// WatchCallback is a function that will be called with file events.
type WatchCallback func(event []WatchEvent)

// WatchEvent represents a file event.
type WatchEvent struct {
	Type runtimev1.FileEvent
	Path string
	Dir  bool
}

// RepoListLimit is the maximum number of files that can be listed in a call to RepoStore.ListGlob.
// This limit is effectively a cap on the number of files in a project because `rill start` lists the project directory using a "**" glob.
const RepoListLimit = 2000

// ErrRepoListLimitExceeded should be returned when RepoListLimit is exceeded.
var ErrRepoListLimitExceeded = fmt.Errorf("glob exceeded limit of %d matched files", RepoListLimit)

// IsIgnored returns true if the path (and any files in nested directories) should be ignored.
// It checks ignoredPaths as well as any additional paths specified.
func IsIgnored(path string, additionalIgnoredPaths []string) bool {
	for _, dir := range ignoredPaths {
		if path == dir {
			return true
		}
		if strings.HasPrefix(path, dir) && path[len(dir)] == '/' {
			return true
		}
	}
	for _, dir := range additionalIgnoredPaths {
		if path == dir {
			return true
		}
		if strings.HasPrefix(path, dir) && path[len(dir)] == '/' {
			return true
		}
	}
	return false
}

type RepoStatus struct {
	// IsGitRepo indicates if the repo is backed by a Git repository.
	IsGitRepo     bool
	Branch        string
	RemoteURL     string
	Subpath       string
	ManagedRepo   bool
	LocalChanges  bool // true if there are local changes (staged, unstaged, or untracked)
	LocalCommits  int32
	RemoteCommits int32
}

type PullOptions struct {
	ForceHandshake bool

	// If userTriggered is true, the latest changes will be pulled from the remote repository honouring DiscardChanges.
	UserTriggered  bool
	DiscardChanges bool
}

// ignoredPaths is a list of paths that are always ignored by the parser.
var ignoredPaths = []string{
	"/tmp",
	"/.git",
	"/node_modules",
	"/.DS_Store",
	"/.vscode",
	"/.idea",
	"/.rillcloud",
}
