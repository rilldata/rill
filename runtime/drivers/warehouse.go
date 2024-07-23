package drivers

import (
	"context"
)

type Warehouse interface {
	// QueryAsFiles downloads results into files and returns an iterator to iterate over them
	QueryAsFiles(ctx context.Context, props map[string]any, opt *QueryOption) (FileIterator, error)
}

type QueryOption struct {
	// TotalLimitInBytes rerpresent the max limit on the bytes that should be downloaded in a file
	TotalLimitInBytes int64
}
