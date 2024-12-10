package drivers

import (
	"context"
	"errors"
)

var ErrNoRows = errors.New("no rows found for the query")

type Warehouse interface {
	// QueryAsFiles downloads results into files and returns an iterator to iterate over them
	QueryAsFiles(ctx context.Context, props map[string]any) (FileIterator, error)
}
