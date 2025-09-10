package druid

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

// In druid there are multiple schemas but all user tables are in druid schema.
// Other useful schema is INFORMATION_SCHEMA for metadata.
// There are 2 more schemas - sys (internal things) and lookup (druid specific lookup).
// While querying druid does not support db name just use schema.table
//
// Since all user tables are in `druid` schema so we hardcode schema as `druid` and does not query database
func (c *connection) All(ctx context.Context, like string, pageSize uint32, pageToken string) ([]*drivers.OlapTable, string, error) {
	var likeClause string
	var args []any
	if like != "" {
		likeClause = "AND LOWER(T.TABLE_NAME) LIKE LOWER(?)"
		args = []any{like}
	}

	q := fmt.Sprintf(`
		SELECT
			T.TABLE_SCHEMA AS SCHEMA,
			T.TABLE_NAME AS NAME,
			T.TABLE_TYPE AS TABLE_TYPE, 
			C.COLUMN_NAME AS COLUMNS,
			C.DATA_TYPE AS COLUMN_TYPE,
			C.IS_NULLABLE = 'YES' AS IS_NULLABLE
		FROM INFORMATION_SCHEMA.TABLES T 
		JOIN INFORMATION_SCHEMA.COLUMNS C ON T.TABLE_SCHEMA = C.TABLE_SCHEMA AND T.TABLE_NAME = C.TABLE_NAME
		WHERE T.TABLE_SCHEMA = 'druid'
		%s
		ORDER BY SCHEMA, NAME, TABLE_TYPE, C.ORDINAL_POSITION
	`, likeClause)

	rows, err := c.db.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	tables, err := scanTables(rows)
	if err != nil {
		return nil, "", err
	}
	return tables, "", nil
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
