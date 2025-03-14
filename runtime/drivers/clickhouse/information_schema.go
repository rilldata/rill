package clickhouse

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

type informationSchema struct {
	c *connection
}

func (c *connection) InformationSchema() drivers.InformationSchema {
	return informationSchema{c: c}
}

func (i informationSchema) All(ctx context.Context, like string) ([]*drivers.Table, error) {
	conn, release, err := i.c.acquireMetaConn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = release() }()

	var likeClause string
	var args []any
	if like != "" {
		likeClause = "AND (LOWER(T.name) LIKE LOWER(?) OR CONCAT(T.database, '.', T.name) LIKE LOWER(?))"
		args = []any{like, like}
	}

	// Clickhouse does not have a concept of schemas. Both table_catalog and table_schema refer to the database where table is located.
	// Given the usual way of querying table in clickhouse is `SELECT * FROM table_name` or `SELECT * FROM database.table_name`.
	// We map clickhouse database to `database schema` and table_name to `table name`.
	q := fmt.Sprintf(`
		SELECT 
			T.database AS SCHEMA,
			T.database = currentDatabase() AS is_default_schema,
			T.name AS NAME,
			if(lower(T.engine) like '%%view%%', 'VIEW', 'TABLE') AS TABLE_TYPE,
			C.name AS COLUMNS,
			C.type AS COLUMN_TYPE,
			C.position AS ORDINAL_POSITION
		FROM system.tables T
		JOIN system.columns C ON T.database = C.database AND T.name = C.table
		-- allow fetching tables from system or information_schema if it is current database
		WHERE (T.database == currentDatabase() OR lower(T.database) NOT IN ('information_schema', 'system'))
		%s
		ORDER BY SCHEMA, NAME, TABLE_TYPE, ORDINAL_POSITION
	`, likeClause)

	rows, err := conn.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables, err := i.scanTables(rows)
	if err != nil {
		return nil, err
	}
	return tables, nil
}

func (i informationSchema) Lookup(ctx context.Context, db, schema, name string) (*drivers.Table, error) {
	conn, release, err := i.c.acquireMetaConn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = release() }()

	var q string
	var args []any
	q = `
		SELECT 
			T.database AS SCHEMA,
			T.database = currentDatabase() AS is_default_schema,
			T.name AS NAME,
			if(lower(T.engine) like '%view%', 'VIEW', 'TABLE') AS TABLE_TYPE,
			C.name AS COLUMNS,
			C.type AS COLUMN_TYPE,
			C.position AS ORDINAL_POSITION
		FROM system.tables T
		JOIN system.columns C ON T.database = C.database AND T.name = C.table
		WHERE T.database = coalesce(?, currentDatabase()) AND T.name = ?
		ORDER BY SCHEMA, NAME, TABLE_TYPE, ORDINAL_POSITION
	`
	if schema == "" {
		args = append(args, nil, name)
	} else {
		args = append(args, schema, name)
	}

	rows, err := conn.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables, err := i.scanTables(rows)
	if err != nil {
		return nil, err
	}

	if len(tables) == 0 {
		return nil, drivers.ErrNotFound
	}

	return tables[0], nil
}

func (i informationSchema) LoadPhysicalSize(ctx context.Context, tables []*drivers.Table) error {
	if len(tables) == 0 {
		return nil
	}
	conn, release, err := i.c.acquireMetaConn(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = release() }()

	var queryBuilder strings.Builder
	queryBuilder.WriteString(`
		SELECT 
			database, 
			table, 
			SUM(bytes_on_disk) AS total_size_bytes
		FROM system.parts
		WHERE active = 1 AND (database, table) IN (
	`)
	args := make([]interface{}, 0, len(tables)*2)
	placeholders := make([]string, 0, len(tables))

	for _, table := range tables {
		placeholders = append(placeholders, "(?, ?)")
		args = append(args, table.DatabaseSchema, table.Name)
	}

	queryBuilder.WriteString(strings.Join(placeholders, ", "))
	queryBuilder.WriteString(") GROUP BY database, table")

	rows, err := conn.QueryxContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	res := make(map[string]map[string]uint64, 0)
	var (
		name, schema string
		size         uint64
	)
	for rows.Next() {
		if err := rows.Scan(&schema, &name, &size); err != nil {
			return err
		}
		schemaTables, ok := res[schema]
		if !ok {
			schemaTables = make(map[string]uint64)
			res[schema] = schemaTables
		}
		schemaTables[name] = size
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, t := range tables {
		schemaTables, ok := res[t.DatabaseSchema]
		if !ok {
			continue
		}
		if size, ok := schemaTables[t.Name]; ok {
			t.PhysicalSizeBytes = int64(size)
		}
	}
	return err
}

func (i informationSchema) scanTables(rows *sqlx.Rows) ([]*drivers.Table, error) {
	var res []*drivers.Table

	for rows.Next() {
		var databaseSchema string
		var isDefaultSchema bool
		var name string
		var tableType string
		var columnName string
		var columnType string
		var oridinalPosition int

		err := rows.Scan(&databaseSchema, &isDefaultSchema, &name, &tableType, &columnName, &columnType, &oridinalPosition)
		if err != nil {
			return nil, err
		}

		// set t to res[len(res)-1] if it's the same table, else set t to a new table and append it
		var t *drivers.Table
		if len(res) > 0 {
			t = res[len(res)-1]
			if !(t.DatabaseSchema == databaseSchema && t.Name == name) {
				t = nil
			}
		}
		if t == nil {
			t = &drivers.Table{
				DatabaseSchema:          databaseSchema,
				IsDefaultDatabaseSchema: isDefaultSchema,
				Name:                    name,
				View:                    tableType == "VIEW",
				Schema:                  &runtimev1.StructType{},
			}
			if !t.View {
				t.PhysicalSizeBytes = -1
			}
			res = append(res, t)
		}

		// parse column type
		colType, err := databaseTypeToPB(columnType, false)
		if err != nil {
			if !errors.Is(err, errUnsupportedType) {
				return nil, err
			}
			if t.UnsupportedCols == nil {
				t.UnsupportedCols = make(map[string]string)
			}
			t.UnsupportedCols[columnName] = columnType
			continue
		}

		// append column
		t.Schema.Fields = append(t.Schema.Fields, &runtimev1.StructType_Field{
			Name: columnName,
			Type: colType,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (i informationSchema) entityType(ctx context.Context, db, name string) (typ string, onCluster bool, err error) {
	conn, release, err := i.c.acquireMetaConn(ctx)
	if err != nil {
		return "", false, err
	}
	defer func() { _ = release() }()

	var q string
	if i.c.config.Cluster == "" {
		q = `SELECT
    			multiIf(engine IN ('MaterializedView', 'View'), 'VIEW', engine = 'Dictionary', 'DICTIONARY', 'TABLE') AS type,
    			0 AS is_on_cluster
			FROM system.tables AS t
			JOIN system.databases AS db ON t.database = db.name
			WHERE t.database = coalesce(?, currentDatabase()) AND t.name = ?`
	} else {
		q = `SELECT
    			multiIf(engine IN ('MaterializedView', 'View'), 'VIEW', engine = 'Dictionary', 'DICTIONARY', 'TABLE') AS type,
    			countDistinct(_shard_num) > 1 AS is_on_cluster
			FROM clusterAllReplicas(` + safeSQLName(i.c.config.Cluster) + `, system.tables) AS t
			JOIN system.databases AS db ON t.database = db.name
			WHERE t.database = coalesce(?, currentDatabase()) AND t.name = ?
			GROUP BY engine, t.name`
	}
	var args []any
	if db == "" {
		args = []any{nil, name}
	} else {
		args = []any{db, name}
	}
	row := conn.QueryRowxContext(ctx, q, args...)
	err = row.Scan(&typ, &onCluster)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, drivers.ErrNotFound
		}
		return "", false, err
	}
	return typ, onCluster, nil
}
