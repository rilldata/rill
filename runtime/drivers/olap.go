package drivers

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/connectors"
)

// ErrUnsupportedConnector is returned from Ingest for unsupported connectors
var ErrUnsupportedConnector = errors.New("drivers: connector not supported")

// OLAPStore is implemented by drivers that are capable of storing, transforming and serving analytical queries
type OLAPStore interface {
	Execute(ctx context.Context, stmt *Statement) (*sqlx.Rows, error)
	Ingest(ctx context.Context, source *connectors.Source) error
	InformationSchema() InformationSchema
}

// Statement wraps a query to execute against an OLAP driver
type Statement struct {
	Query    string
	Args     []any
	DryRun   bool
	Priority int
}

// InformationSchema contains information about existing tables in an OLAP driver
type InformationSchema interface {
	All(ctx context.Context) ([]*Table, error)
	Lookup(ctx context.Context, name string) (*Table, error)
}

// Table represents a table in an information schema
type Table struct {
	Database string
	Schema   string
	Name     string
	Type     string
	Columns  []Column
}

// Column represents a column in a table
type Column struct {
	Name     string
	Type     string
	Nullable bool
}
