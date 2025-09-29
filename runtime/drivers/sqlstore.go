package drivers

import (
	"context"
)

// SQLStore is implemented by drivers that support SQL querying.
type SQLStore interface {
	// Query executes a SQL query and returns a RowIterator to iterate over the results.
	Query(ctx context.Context, query string, args ...any) (RowIterator, error)
}

// RowIterator is used to iterate over rows returned by a SQL query.
type RowIterator interface {
	Close() error
	// Next returns the next row as a map of column names to values.
	Next(ctx context.Context) (map[string]any, error)

	// Add more methods as needed like ColumnTypes, etc.
}
