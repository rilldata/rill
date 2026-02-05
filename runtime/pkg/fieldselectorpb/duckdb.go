package fieldselectorpb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rilldata/rill/runtime/drivers"

	// Import the DuckDB driver
	_ "github.com/duckdb/duckdb-go/v2"
)

// resolveDuckDBExpression implements Resolve for FieldSelector.duckdb_expression.
func resolveDuckDBExpression(expr string, all []string) ([]string, error) {
	var res []string
	err := withEphemeralDuckDBConn(func(ctx context.Context, conn *sql.Conn) error {
		var ddl strings.Builder
		ddl.WriteString("CREATE TEMPORARY TABLE t AS SELECT ")
		for i, f := range all {
			if i > 0 {
				ddl.WriteString(", ")
			}
			ddl.WriteString("1 AS ")
			ddl.WriteString(drivers.DialectDuckDB.EscapeIdentifier(f))
		}

		_, err := conn.ExecContext(ctx, ddl.String())
		if err != nil {
			return err
		}
		defer func() {
			_, _ = conn.ExecContext(ctx, "DROP TABLE t")
		}()

		rows, err := conn.QueryContext(ctx, fmt.Sprintf("SELECT %s FROM t", expr))
		if err != nil {
			return err
		}
		defer rows.Close()

		res, err = rows.Columns()
		if err != nil {
			return err
		}

		return rows.Err()
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Use a global in-memory DuckDB connection for invoking DuckDB's json_serialize_sql and json_deserialize_sql
var (
	db     *sql.DB
	dbOnce sync.Once
)

// withEphemeralDuckDBConn acquires an ephemeral in-memory DuckDB connection.
// It should only be used for temporary operations.
func withEphemeralDuckDBConn(fn func(ctx context.Context, conn *sql.Conn) error) error {
	// Lazily initialize db global as an in-memory DuckDB connection.
	dbOnce.Do(func() {
		var err error
		db, err = sql.Open("duckdb", "")
		if err != nil {
			panic(err)
		}
		db.SetMaxOpenConns(1)
	})

	// Prepare a short-lived context. This should never take long, but better be safe in case of panics that don't release connections or such.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Get a connection from the pool (currently limited to one)
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	return fn(ctx, conn)
}
