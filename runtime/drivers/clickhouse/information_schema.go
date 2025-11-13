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
	"github.com/rilldata/rill/runtime/pkg/pagination"
)

func (c *Connection) ListDatabaseSchemas(ctx context.Context, pageSize uint32, pageToken string) ([]*drivers.DatabaseSchemaInfo, string, error) {
	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)

	var args []any
	var condFilter string
	if c.config.DatabaseWhitelist != "" {
		dbs := strings.Split(c.config.DatabaseWhitelist, ",")
		var sb strings.Builder
		for i, db := range dbs {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString("?")
			args = append(args, strings.TrimSpace(db))
		}
		condFilter = fmt.Sprintf("(schema_name IN (%s))", sb.String())
	} else {
		condFilter = "(schema_name == currentDatabase() OR lower(schema_name) NOT IN ('information_schema', 'system'))"
	}

	if pageToken != "" {
		var startAfter string
		if err := pagination.UnmarshalPageToken(pageToken, &startAfter); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		condFilter += "	AND schema_name > ?"
		args = append(args, startAfter)
	}

	q := fmt.Sprintf(`
	SELECT
		 name as schema_name
	FROM system.databases
	WHERE %s 
	ORDER BY schema_name 
	LIMIT ?
	`, condFilter)
	args = append(args, limit+1)

	conn, release, err := c.acquireMetaConn(ctx)
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = release() }()

	rows, err := conn.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var res []*drivers.DatabaseSchemaInfo
	var schema string
	for rows.Next() {
		if err := rows.Scan(&schema); err != nil {
			return nil, "", err
		}
		res = append(res, &drivers.DatabaseSchemaInfo{
			Database:       "",
			DatabaseSchema: schema,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	next := ""
	if len(res) > limit {
		res = res[:limit]
		next = pagination.MarshalPageToken(res[len(res)-1].DatabaseSchema)
	}
	return res, next, nil
}

func (c *Connection) ListTables(ctx context.Context, database, databaseSchema string, pageSize uint32, pageToken string) ([]*drivers.TableInfo, string, error) {
	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)

	q := `
	SELECT
		name AS table_name,
		CASE WHEN match(engine, 'View') THEN true ELSE false END AS view
	FROM system.tables
	WHERE is_temporary = 0 AND database = ?
	`
	args := []any{databaseSchema}
	if pageToken != "" {
		var startAfter string
		if err := pagination.UnmarshalPageToken(pageToken, &startAfter); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		q += "	AND table_name > ?"
		args = append(args, startAfter)
	}
	q += `
	ORDER BY table_name 
	LIMIT ?
	`
	args = append(args, limit+1)

	conn, release, err := c.acquireMetaConn(ctx)
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = release() }()

	rows, err := conn.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var res []*drivers.TableInfo
	var name string
	var view bool
	for rows.Next() {
		if err := rows.Scan(&name, &view); err != nil {
			return nil, "", err
		}
		res = append(res, &drivers.TableInfo{
			Name: name,
			View: view,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	next := ""
	if len(res) > limit {
		res = res[:limit]
		next = pagination.MarshalPageToken(res[len(res)-1].Name)
	}
	return res, next, nil
}

func (c *Connection) GetTable(ctx context.Context, database, databaseSchema, table string) (*drivers.TableMetadata, error) {
	conn, release, err := c.acquireMetaConn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = release() }()

	q := `
    SELECT
        CASE WHEN match(engine, 'View') THEN true ELSE false END AS view,
        c.name AS column_name,
        c.type AS data_type
    FROM system.tables t
    LEFT JOIN system.columns c
		ON t.database = c.database AND t.name = c.table
    WHERE t.database = ? AND t.name = ?
    ORDER BY c.position
    `
	rows, err := conn.QueryxContext(ctx, q, databaseSchema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schemaMap := make(map[string]string)
	var view bool
	var colName, dataType string
	for rows.Next() {
		if err := rows.Scan(&view, &colName, &dataType); err != nil {
			return nil, err
		}
		if pbType, err := databaseTypeToPB(dataType, false); err != nil {
			if errors.Is(err, errUnsupportedType) {
				schemaMap[colName] = fmt.Sprintf("UNKNOWN(%s)", dataType)
			} else {
				return nil, err
			}
		} else if pbType.Code == runtimev1.Type_CODE_UNSPECIFIED {
			schemaMap[colName] = fmt.Sprintf("UNKNOWN(%s)", dataType)
		} else {
			schemaMap[colName] = strings.TrimPrefix(pbType.Code.String(), "CODE_")
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &drivers.TableMetadata{
		Schema: schemaMap,
		View:   view,
	}, nil
}

func (c *Connection) All(ctx context.Context, like string, pageSize uint32, pageToken string) ([]*drivers.OlapTable, string, error) {
	conn, release, err := c.acquireMetaConn(ctx)
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = release() }()
	var args []any
	var filter string
	if c.config.DatabaseWhitelist != "" {
		dbs := strings.Split(c.config.DatabaseWhitelist, ",")
		var sb strings.Builder
		for i, db := range dbs {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString("?")
			args = append(args, strings.TrimSpace(db))
		}
		filter = fmt.Sprintf("(T.database IN (%s))", sb.String())
	} else {
		filter = "(T.database == currentDatabase() OR lower(T.database) NOT IN ('information_schema', 'system'))"
	}

	if like != "" {
		filter += " AND (LOWER(T.name) LIKE LOWER(?) OR CONCAT(T.database, '.', T.name) LIKE LOWER(?))"
		args = append(args, like, like)
	}

	if pageToken != "" {
		var startAfterSchema, startAfterName string
		if err := pagination.UnmarshalPageToken(pageToken, &startAfterSchema, &startAfterName); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		filter += " AND (T.database > ? OR (T.database = ? AND T.name > ?))"
		args = append(args, startAfterSchema, startAfterSchema, startAfterName)
	}

	// Clickhouse does not have a concept of schemas. Both table_catalog and table_schema refer to the database where table is located.
	// Given the usual way of querying table in clickhouse is `SELECT * FROM table_name` or `SELECT * FROM database.table_name`.
	// We map clickhouse database to `database schema` and table_name to `table name`.
	q := fmt.Sprintf(`
		SELECT 
			LT.database AS SCHEMA,
			LT.database = currentDatabase() AS is_default_schema,
			LT.name AS NAME,
			if(lower(LT.engine) like '%%view%%', 'VIEW', 'TABLE') AS TABLE_TYPE,
			C.name AS COLUMNS,
			C.type AS COLUMN_TYPE,
			C.position AS ORDINAL_POSITION
		FROM (
			SELECT 
				T.database,
				T.name,
				T.engine
			FROM system.tables T
			-- allow fetching tables from system or information_schema if it is current database
			WHERE %s
			ORDER BY database, name, engine
			LIMIT ?
		) LT
		JOIN system.columns C ON LT.database = C.database AND LT.name = C.table
		ORDER BY SCHEMA, NAME, TABLE_TYPE, ORDINAL_POSITION
	`, filter)

	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)
	args = append(args, limit+1)

	rows, err := conn.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	tables, err := scanTables(rows)
	if err != nil {
		return nil, "", err
	}

	next := ""
	if len(tables) > limit {
		tables = tables[:limit]
		lastTable := tables[len(tables)-1]
		next = pagination.MarshalPageToken(lastTable.DatabaseSchema, lastTable.Name)
	}

	return tables, next, nil
}

func (c *Connection) Lookup(ctx context.Context, db, schema, name string) (*drivers.OlapTable, error) {
	conn, release, err := c.acquireMetaConn(ctx)
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

	tables, err := scanTables(rows)
	if err != nil {
		return nil, err
	}

	if len(tables) == 0 {
		return nil, drivers.ErrNotFound
	}

	return tables[0], nil
}

func (c *Connection) LoadPhysicalSize(ctx context.Context, tables []*drivers.OlapTable) error {
	if len(tables) == 0 {
		return nil
	}
	conn, release, err := c.acquireMetaConn(ctx)
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

func scanTables(rows *sqlx.Rows) ([]*drivers.OlapTable, error) {
	var res []*drivers.OlapTable

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
		var t *drivers.OlapTable
		if len(res) > 0 {
			t = res[len(res)-1]
			if !(t.DatabaseSchema == databaseSchema && t.Name == name) {
				t = nil
			}
		}
		if t == nil {
			t = &drivers.OlapTable{
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

func (c *Connection) entityType(ctx context.Context, db, name string) (typ string, onCluster bool, err error) {
	conn, release, err := c.acquireMetaConn(ctx)
	if err != nil {
		return "", false, err
	}
	defer func() { _ = release() }()

	var q string
	if c.config.Cluster == "" {
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
			FROM clusterAllReplicas(` + safeSQLName(c.config.Cluster) + `, system.tables) AS t
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
