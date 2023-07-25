package drivers

import "context"

type FileStore interface {
	// FilePaths returns local absolute paths where files are stored
	FilePaths(ctx context.Context, src *FileSource) ([]string, error)
}
