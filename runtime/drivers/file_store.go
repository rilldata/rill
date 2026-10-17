package drivers

import "context"

type FileStore interface {
	// FilePaths returns local absolute paths where files are stored.
	// The caller must remove the files after use.
	FilePaths(ctx context.Context, src map[string]any) ([]string, error)
}
