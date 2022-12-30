package drivers

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/connectors"
)

// ErrUnsupportedConnector is returned from Ingest for unsupported connectors.
var ErrUnsupportedConnector = errors.New("drivers: connector not supported")

// OLAPStore is implemented by drivers that are capable of storing, transforming and serving analytical queries.
type OLAPStore interface {
	Dialect() Dialect
	Execute(ctx context.Context, stmt *Statement) (*Result, error)
	Ingest(ctx context.Context, env *connectors.Env, source *connectors.Source) error
	InformationSchema() InformationSchema
}

// Statement wraps a query to execute against an OLAP driver.
type Statement struct {
	Query    string
	Args     []any
	DryRun   bool
	Priority int
}

// Result wraps the results of query.
type Result struct {
	*sqlx.Rows
	Schema *runtimev1.StructType
}

// InformationSchema contains information about existing tables in an OLAP driver.
type InformationSchema interface {
	All(ctx context.Context) ([]*Table, error)
	Lookup(ctx context.Context, name string) (*Table, error)
}

// Table represents a table in an information schema.
type Table struct {
	Database       string
	DatabaseSchema string
	Name           string
	Schema         *runtimev1.StructType
}

// Dialect enumerates OLAP query languages.
type Dialect int

const (
	DialectUnspecified Dialect = iota
	DialectDuckDB
	DialectDruid
)

func (d Dialect) String() string {
	switch d {
	case DialectUnspecified:
		return ""
	case DialectDuckDB:
		return "duckdb"
	case DialectDruid:
		return "druid"
	default:
		panic("not implemented")
	}
}
