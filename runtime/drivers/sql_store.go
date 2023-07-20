package drivers

import (
	"context"
)

type Field struct {
	Name string
	Type string
}

type Schema []Field

// SQLStore is implemented by drivers capable of running sql queries and generating an iterator to consume results.
// In future the results can be produced in other formats like arrow as well.
// This is kept differnt from OLAP store since results can be produced in other formats in future and drivers like gcs, athena may not produce sql.Rows.
type SQLStore interface {
	Exec(ctx context.Context, src *DatabaseSource) (RowIterator, error)
}

// RowIterator returns an iterator to iterate over result of a sql query
type RowIterator interface {
	// Schema of the underlying data
	// TODO :: some enum for types may be ?
	ResultSchema(ctx context.Context) (Schema, error)
	// Next fetches next row
	Next(ctx context.Context) ([]any, error)
	// Close closes the iterator and frees resources
	Close() error
	// Size returns size of data downloaded in unit.
	// Returns 0,false if not able to compute size in given unit
	Size(unit ProgressUnit) (uint64, bool)
}
