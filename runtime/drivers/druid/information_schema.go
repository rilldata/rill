package druid

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pagination"
)

// In druid there are multiple schemas but all user tables are in druid schema.
// Other useful schema is INFORMATION_SCHEMA for metadata.
// There are 2 more schemas - sys (internal things) and lookup (druid specific lookup).
// While querying druid does not support db name just use schema.table
//
// Since all user tables are in `druid` schema so we hardcode schema as `druid` and does not query database
func (c *connection) ListDatabaseSchemas(ctx context.Context, pageSize uint32, pageToken string) ([]*drivers.DatabaseSchemaInfo, string, error) {
	if pageToken != "" {
		return []*drivers.DatabaseSchemaInfo{}, "", nil
	}
	return []*drivers.DatabaseSchemaInfo{{Database: "", DatabaseSchema: "druid"}}, "", nil
}

func (c *connection) ListTables(ctx context.Context, database, databaseSchema string, pageSize uint32, pageToken string) ([]*drivers.TableInfo, string, error) {
	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)

	var filter string
	args := []any{databaseSchema}
	if pageToken != "" {
		var startAfter string
		if err := pagination.UnmarshalPageToken(pageToken, &startAfter); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		filter += " AND TABLE_NAME > ?"
		args = append(args, startAfter)
	}

	q := fmt.Sprintf(`
	SELECT
		TABLE_NAME,
		TABLE_TYPE = 'VIEW' AS view
	FROM INFORMATION_SCHEMA.TABLES
	WHERE TABLE_SCHEMA = ? %s
	ORDER BY TABLE_NAME 
	LIMIT %d
	`, filter, limit+1)

	rows, err := c.db.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var res []*drivers.TableInfo
	var name string
	var typ bool
	for rows.Next() {
		if err := rows.Scan(&name, &typ); err != nil {
			return nil, "", err
		}
		res = append(res, &drivers.TableInfo{
			Name: name,
			View: typ,
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

func (c *connection) GetTable(ctx context.Context, database, databaseSchema, table string) (*drivers.TableMetadata, error) {
	// Ensure Coordinator is ready.
	// The issues is that the request
	//	SELECT ...
	//	FROM INFORMATION_SCHEMA.TABLES T
	//	JOIN INFORMATION_SCHEMA.COLUMNS C ON T.TABLE_SCHEMA = C.TABLE_SCHEMA AND T.TABLE_NAME = C.TABLE_NAME
	//	WHERE T.TABLE_SCHEMA = 'druid' AND T.TABLE_NAME = ?
	//	ORDER BY SCHEMA, NAME, TABLE_TYPE, C.ORDINAL_POSITION
	// returns false-negative if the Coordinator is being restarted. Retrier is a more abstract component and it doesn't check
	// if SQL tries to retrieve the dynamic schema and the will be no error from Druid Router
	// (because if the dynamic schema is empty - it's considered OK by the Druid cluster).
	q := "SELECT * FROM sys.segments LIMIT 1"
	rows, err := c.db.QueryxContext(ctx, q)
	if err != nil {
		return nil, err
	}
	rows.Close()

	q = `
    SELECT
    	T.TABLE_TYPE = 'VIEW' AS view,
    	C.COLUMN_NAME,
    	C.DATA_TYPE,
    	C.IS_NULLABLE = 'YES' AS is_nullable
	FROM INFORMATION_SCHEMA.TABLES T
	LEFT JOIN INFORMATION_SCHEMA.COLUMNS C 
		ON T.TABLE_SCHEMA = C.TABLE_SCHEMA AND T.TABLE_NAME = C.TABLE_NAME
	WHERE T.TABLE_SCHEMA = ? AND T.TABLE_NAME = ?
	ORDER BY C.ORDINAL_POSITION
    `
	rows, err = c.db.QueryxContext(ctx, q, databaseSchema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schemaMap := make(map[string]string)
	var view, nullable bool
	var colName, dataType string
	for rows.Next() {
		if err := rows.Scan(&view, &colName, &dataType, &nullable); err != nil {
			return nil, err
		}
		pbType := databaseTypeToPB(dataType, nullable)
		if pbType.Code == runtimev1.Type_CODE_UNSPECIFIED {
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

func (c *connection) All(ctx context.Context, like string, pageSize uint32, pageToken string) ([]*drivers.OlapTable, string, error) {
	var filter string
	var args []any
	if like != "" {
		filter = " AND LOWER(T.TABLE_NAME) LIKE LOWER(?)"
		args = []any{like}
	}

	// Add pagination clause
	if pageToken != "" {
		var startAfterName string
		if err := pagination.UnmarshalPageToken(pageToken, &startAfterName); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		filter += " AND T.TABLE_NAME > ?"
		args = append(args, startAfterName)
	}
	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)

	q := fmt.Sprintf(`
		SELECT
			LT.TABLE_SCHEMA AS SCHEMA,
			LT.TABLE_NAME AS NAME,
			LT.TABLE_TYPE AS TABLE_TYPE, 
			C.COLUMN_NAME AS COLUMNS,
			C.DATA_TYPE AS COLUMN_TYPE,
			C.IS_NULLABLE = 'YES' AS IS_NULLABLE
		FROM (
			SELECT
				T.TABLE_SCHEMA,
				T.TABLE_NAME,
				T.TABLE_TYPE
			FROM INFORMATION_SCHEMA.TABLES T
			WHERE T.TABLE_SCHEMA = 'druid' %s
			ORDER BY TABLE_SCHEMA, TABLE_NAME, TABLE_TYPE
			LIMIT %d
		) LT
		JOIN INFORMATION_SCHEMA.COLUMNS C ON LT.TABLE_SCHEMA = C.TABLE_SCHEMA AND LT.TABLE_NAME = C.TABLE_NAME		
		ORDER BY SCHEMA, NAME, TABLE_TYPE, C.ORDINAL_POSITION
	`, filter, (limit + 1))

	rows, err := c.db.QueryxContext(ctx, q, args...)
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
		next = pagination.MarshalPageToken(lastTable.Name)
	}

	return tables, next, nil
}

func (c *connection) Lookup(ctx context.Context, db, schema, name string) (*drivers.OlapTable, error) {
	// Ensure Coordinator is ready.
	// The issues is that the request
	//	SELECT ...
	//	FROM INFORMATION_SCHEMA.TABLES T
	//	JOIN INFORMATION_SCHEMA.COLUMNS C ON T.TABLE_SCHEMA = C.TABLE_SCHEMA AND T.TABLE_NAME = C.TABLE_NAME
	//	WHERE T.TABLE_SCHEMA = 'druid' AND T.TABLE_NAME = ?
	//	ORDER BY SCHEMA, NAME, TABLE_TYPE, C.ORDINAL_POSITION
	// returns false-negative if the Coordinator is being restarted. Retrier is a more abstract component and it doesn't check
	// if SQL tries to retrieve the dynamic schema and the will be no error from Druid Router
	// (because if the dynamic schema is empty - it's considered OK by the Druid cluster).
	q := "SELECT * FROM sys.segments LIMIT 1"
	rows, err := c.db.QueryxContext(ctx, q, name)
	if err != nil {
		return nil, err
	}
	rows.Close()

	q = `
		SELECT
			T.TABLE_SCHEMA AS SCHEMA,
			T.TABLE_NAME AS NAME,
			T.TABLE_TYPE AS TABLE_TYPE, 
			C.COLUMN_NAME AS COLUMN_NAME,
			C.DATA_TYPE AS COLUMN_TYPE,
			C.IS_NULLABLE = 'YES' AS IS_NULLABLE
		FROM INFORMATION_SCHEMA.TABLES T 
		JOIN INFORMATION_SCHEMA.COLUMNS C ON T.TABLE_SCHEMA = C.TABLE_SCHEMA AND T.TABLE_NAME = C.TABLE_NAME
		WHERE T.TABLE_SCHEMA = 'druid' AND T.TABLE_NAME = ?
		ORDER BY SCHEMA, NAME, TABLE_TYPE, C.ORDINAL_POSITION
	`

	rows, err = c.db.QueryxContext(ctx, q, name)
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

func (c *connection) LoadPhysicalSize(ctx context.Context, tables []*drivers.OlapTable) error {
	q := `SELECT
    		datasource,
    		SUM("size") AS total_size
		FROM sys.segments
		WHERE is_active = 1
		GROUP BY 1`
	rows, err := c.db.QueryxContext(ctx, q)
	if err != nil {
		return err
	}
	defer rows.Close()
	res := make(map[string]uint64, len(tables))
	var (
		name string
		size uint64
	)
	for rows.Next() {
		if err := rows.Scan(&name, &size); err != nil {
			return err
		}
		res[name] = size
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, t := range tables {
		if size, ok := res[t.Name]; ok {
			t.PhysicalSizeBytes = int64(size)
		}
	}
	return nil
}

func scanTables(rows *sqlx.Rows) ([]*drivers.OlapTable, error) {
	var res []*drivers.OlapTable

	for rows.Next() {
		var schema string
		var name string
		var tableType string
		var columnName string
		var columnType string
		var nullable bool

		err := rows.Scan(&schema, &name, &tableType, &columnName, &columnType, &nullable)
		if err != nil {
			return nil, err
		}

		// set t to res[len(res)-1] if it's the same table, else set t to a new table and append it
		var t *drivers.OlapTable
		if len(res) > 0 {
			t = res[len(res)-1]
			if !(t.DatabaseSchema == schema && t.Name == name) {
				t = nil
			}
		}
		if t == nil {
			t = &drivers.OlapTable{
				DatabaseSchema:          schema,
				IsDefaultDatabaseSchema: true,
				Name:                    name,
				Schema:                  &runtimev1.StructType{},
				PhysicalSizeBytes:       -1,
			}
			res = append(res, t)
		}

		// append column
		t.Schema.Fields = append(t.Schema.Fields, &runtimev1.StructType_Field{
			Name: columnName,
			Type: databaseTypeToPB(columnType, nullable),
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func databaseTypeToPB(dbt string, nullable bool) *runtimev1.Type {
	t := &runtimev1.Type{Nullable: nullable}
	switch dbt {
	case "BOOLEAN":
		t.Code = runtimev1.Type_CODE_BOOL
	case "TINYINT":
		t.Code = runtimev1.Type_CODE_INT8
	case "SMALLINT":
		t.Code = runtimev1.Type_CODE_INT16
	case "INTEGER":
		t.Code = runtimev1.Type_CODE_INT32
	case "BIGINT":
		t.Code = runtimev1.Type_CODE_INT64
	case "FLOAT":
		t.Code = runtimev1.Type_CODE_FLOAT32
	case "DOUBLE":
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "REAL":
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "DECIMAL":
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "CHAR":
		t.Code = runtimev1.Type_CODE_STRING
	case "VARCHAR":
		t.Code = runtimev1.Type_CODE_STRING
	case "TIMESTAMP":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "DATE":
		t.Code = runtimev1.Type_CODE_DATE
	case "OTHER":
		t.Code = runtimev1.Type_CODE_JSON
	}

	return t
}
