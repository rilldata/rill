package duckdbsql

import (
	"context"
	databasesql "database/sql"
	"database/sql/driver"
	"fmt"
	"sync"

	"github.com/duckdb/duckdb-go/v2"
	"github.com/rilldata/rill/runtime/drivers/duckdb/extensions"
)

// queryString runs a DuckDB query and returns the result as a scalar string
func queryString(qry string, args ...any) ([]byte, error) {
	rows, err := query(qry, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var res []byte
	for rows.Next() {
		err := rows.Scan(&res)
		if err != nil {
			return nil, err
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

// Use a global in-memory DuckDB connection for invoking DuckDB's json_serialize_sql and json_deserialize_sql
var (
	db     *databasesql.DB
	dbOnce sync.Once
)

// query runs a DuckDB query
func query(qry string, args ...any) (*databasesql.Rows, error) {
	err := extensions.InstallExtensionsOnce()
	if err != nil {
		fmt.Printf("failed to install embedded DuckDB extensions, let DuckDB download them: %v\n", err)
	}

	// Lazily initialize db global as an in-memory DuckDB connection
	dbOnce.Do(func() {
		// Using NewConnector since DuckDB requires extensions to be loaded separately on each connection
		connector, err := duckdb.NewConnector("", func(conn driver.ExecerContext) error {
			// Load JSON extension
			_, err := conn.ExecContext(context.Background(), "INSTALL 'json'; LOAD 'json';", nil)
			if err != nil {
				return err
			}
			// Lock it down
			_, err = conn.ExecContext(context.Background(), "SET enable_external_access=false", nil)
			return err
		})
		if err != nil {
			panic(err)
		}

		// Set global
		db = databasesql.OpenDB(connector)
		db.SetMaxOpenConns(1)
	})

	return db.Query(qry, args...)
}
