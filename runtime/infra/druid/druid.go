package druid

import (
	"context"

	_ "github.com/apache/calcite-avatica-go/v5"
	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/infra"
)

func init() {
	infra.Register("druid", driver{})
}

type driver struct{}

// Open connects to a Druid cluster using Avatica. Note that the Druid connection string must have
// the form "http://host/druid/v2/sql/avatica-protobuf/".
func (d driver) Open(dsn string) (infra.Connection, error) {
	db, err := sqlx.Open("avatica", dsn)
	if err != nil {
		return nil, err
	}

	conn := &connection{db: db}
	return conn, nil
}

type connection struct {
	db *sqlx.DB
}

func (c *connection) Close() error {
	return c.db.Close()
}

func (c *connection) InformationSchema() string {
	return ""
}

func (c *connection) Execute(ctx context.Context, stmt *infra.Statement) (*sqlx.Rows, error) {
	if stmt.DryRun {
		// TODO: Find way to validate with args
		prepared, err := c.db.PrepareContext(ctx, stmt.Query)
		if err != nil {
			return nil, err
		}
		prepared.Close()
		return nil, nil
	}

	rows, err := c.db.QueryxContext(ctx, stmt.Query, stmt.Args...)
	if err != nil {
		return nil, err
	}

	return rows, nil
}
