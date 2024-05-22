package drivers

import (
	"context"
	"database/sql/driver"
	"errors"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

var ErrIteratorDone = errors.New("empty iterator")

var ErrNoRows = errors.New("no rows found for the query")

// SQLStore is implemented by drivers capable of running sql queries and generating an iterator to consume results.
// In future the results can be produced in other formats like arrow as well.
// May be call it DataWarehouse to differentiate from OLAP or postgres?
type SQLStore interface {
	// Query returns driver.RowIterator to iterate over results row by row
	Query(ctx context.Context, props map[string]any) (RowIterator, error)
	// QueryAsFiles downloads results into files and returns an iterator to iterate over them
	QueryAsFiles(ctx context.Context, props map[string]any, opt *QueryOption, p Progress) (FileIterator, error)
}

type QueryOption struct {
	// TotalLimitInBytes rerpresent the max limit on the bytes that should be downloaded in a file
	TotalLimitInBytes int64
}

// RowIterator returns an iterator to iterate over result of a sql query
type RowIterator interface {
	// Schema of the underlying data
	Schema(ctx context.Context) (*runtimev1.StructType, error)
	// Next fetches next row
	Next(ctx context.Context) ([]driver.Value, error)
	// Close closes the iterator and frees resources
	Close() error
	// Size returns total size of data downloaded in unit.
	// Returns 0,false if not able to compute size in given unit
	Size(unit ProgressUnit) (uint64, bool)
}
