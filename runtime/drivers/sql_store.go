package drivers

import (
	"context"
)

type Field struct {
	Name string
	Type string
}

type Schema []Field

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
}
