package drivers

import "context"

type FileStore interface {
	// FilePaths returns local absolute paths where files are stored.
	// The caller must call release after use.
	FilePaths(ctx context.Context, src map[string]any) (paths []string, release func(), err error)
}
