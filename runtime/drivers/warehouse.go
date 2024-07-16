package drivers

import (
	"context"
	"net/url"
)

type Warehouse interface {
	// QueryAsFiles downloads results into files and returns an iterator to iterate over them
	QueryAsFiles(ctx context.Context, props map[string]any, opt *QueryOption) (FileIterator, error)
	Export(ctx context.Context, props map[string]any, store ObjectStore, outputLocation string) (*ExportResult, error)
}

type QueryOption struct {
	// TotalLimitInBytes rerpresent the max limit on the bytes that should be downloaded in a file
	TotalLimitInBytes int64
}

type ExportResult struct {
	Path   string
	Format FileFormat
}

func (e *ExportResult) Glob() (string, error) {
	return url.JoinPath(e.Path, "*."+string(e.Format))
}
