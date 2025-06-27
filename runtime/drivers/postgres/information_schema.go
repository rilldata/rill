package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

var errUnsupportedType = errors.New("encountered unsupported postgres type")

func (c *connection) All(ctx context.Context, like string) ([]*drivers.Table, error) {
	var likeClause string
	var args []any
	if like != "" {
		likeClause = "AND (LOWER(T.table_name) LIKE LOWER($1) OR CONCAT(T.table_schema, '.', T.table_name) LIKE LOWER($2))"
		args = []any{like, like}
	}

	q := fmt.Sprintf(`
		SELECT 
			T.table_catalog AS database,
			T.table_schema AS database_schema,
			T.table_catalog = current_database() AS is_default_database,
			T.table_schema = current_schema() AS is_default_database_schema,
			T.table_name AS name,
			T.table_type,
			C.column_name AS columns,
			C.data_type AS column_type
		FROM information_schema.tables T
		JOIN information_schema.columns C ON T.table_catalog = C.table_catalog AND T.table_schema = C.table_schema AND T.table_name = C.table_name
		WHERE T.table_schema NOT IN ('information_schema', 'pg_catalog', 'pg_auto_copy')
		%s
		ORDER BY database, database_schema, name, table_type
	`, likeClause)

	db, err := c.getDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables, err := c.scanTables(rows)
	if err != nil {
		return nil, err
	}
	return tables, nil
}

func (c *connection) Lookup(ctx context.Context, dbName, schema, name string) (*drivers.Table, error) {
	var q string
	var args []any
	q = `
		SELECT 
			T.table_catalog AS database,
			T.table_schema AS database_schema,
			T.table_catalog = current_database() AS is_default_database,
			T.table_schema = current_schema() AS is_default_database_schema,
			T.table_name AS name,
			T.table_type,
			C.column_name AS columns,
			C.data_type AS column_type
		FROM information_schema.tables T
		JOIN information_schema.columns C ON T.table_catalog = C.table_catalog AND T.table_schema = C.table_schema AND T.table_name = C.table_name
		WHERE T.table_catalog = COALESCE($1, current_database()) AND T.table_schema = COALESCE($2, current_schema()) AND T.table_name = $3
		ORDER BY database, database_schema, name, table_type, ordinal_position
	`
	if dbName == "" {
		args = append(args, nil)
	} else {
		args = append(args, dbName)
	}
	if schema == "" {
		args = append(args, nil)
	} else {
		args = append(args, schema)
	}
	args = append(args, name)

	db, err := c.getDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables, err := c.scanTables(rows)
	if err != nil {
		return nil, err
	}

	if len(tables) == 0 {
		return nil, drivers.ErrNotFound
	}

	return tables[0], nil
}

func (c *connection) LoadPhysicalSize(ctx context.Context, tables []*drivers.Table) error {
	if len(tables) == 0 {
		return nil
	}
	for _, t := range tables {
		t.PhysicalSizeBytes = -1
	}
	return nil
}

func (c *connection) scanTables(rows *sqlx.Rows) ([]*drivers.Table, error) {
	var res []*drivers.Table

	for rows.Next() {
		var database string
		var databaseSchema string
		var isDefaultDatabase bool
		var isDefaultDatabaseSchema bool
		var name string
		var tableType string
		var columnName string
		var columnType string

		err := rows.Scan(&database, &databaseSchema, &isDefaultDatabase, &isDefaultDatabaseSchema, &name, &tableType, &columnName, &columnType)
		if err != nil {
			return nil, err
		}

		// set t to res[len(res)-1] if it's the same table, else set t to a new table and append it
		var t *drivers.Table
		if len(res) > 0 {
			t = res[len(res)-1]
			if !(t.Database == database && t.DatabaseSchema == databaseSchema && t.Name == name) {
				t = nil
			}
		}
		if t == nil {
			t = &drivers.Table{
				Database:                database,
				DatabaseSchema:          databaseSchema,
				IsDefaultDatabase:       isDefaultDatabase,
				IsDefaultDatabaseSchema: isDefaultDatabaseSchema,
				Name:                    name,
				View:                    tableType == "VIEW",
				Schema:                  &runtimev1.StructType{},
				PhysicalSizeBytes:       -1,
			}
			res = append(res, t)
		}

		// parse column type
		colType, err := databaseTypeToPB(columnType)
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

func databaseTypeToPB(postgresType string) (*runtimev1.Type, error) {
	typ := strings.ToLower(strings.TrimSpace(postgresType))

	switch {
	case typ == "boolean", typ == "bool":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_BOOL}, nil
	case typ == "smallint":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT16}, nil
	case typ == "int", typ == "integer":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT32}, nil
	case typ == "bigint":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT64}, nil
	case strings.HasPrefix(typ, "decimal"), strings.HasPrefix(typ, "numeric"):
		return &runtimev1.Type{Code: runtimev1.Type_CODE_DECIMAL}, nil
	case typ == "real", typ == "float4":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_FLOAT32}, nil
	case typ == "double precision", typ == "float8":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_FLOAT64}, nil
	case strings.HasPrefix(typ, "varchar"), strings.HasPrefix(typ, "char"), typ == "text":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}, nil
	case typ == "bytea":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_BYTES}, nil
	case typ == "json", typ == "jsonb":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_JSON}, nil
	case typ == "uuid":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_UUID}, nil
	case typ == "date":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_DATE}, nil
	case typ == "time", strings.HasPrefix(typ, "time"):
		return &runtimev1.Type{Code: runtimev1.Type_CODE_TIME}, nil
	case strings.HasPrefix(typ, "timestamp"):
		return &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP}, nil
	case typ == "interval":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INTERVAL}, nil
	case typ == "xml":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}, nil
	case typ == "money":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_DECIMAL}, nil
	case typ == "bit varying", typ == "bit":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}, nil

	// Array types are already flattened in your input to just "array"
	case typ == "array":
		// Cannot infer element type from just "array"
		return &runtimev1.Type{Code: runtimev1.Type_CODE_ARRAY}, nil

	// Geometric and Network types
	case typ == "point", typ == "line", typ == "lseg", typ == "box",
		typ == "path", typ == "polygon", typ == "circle",
		typ == "inet", typ == "cidr", typ == "macaddr":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}, nil

	// Range types
	case typ == "int4range", typ == "numrange", typ == "tsrange",
		typ == "tstzrange", typ == "daterange":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}, nil

	// User-defined type fallback
	case typ == "user-defined", typ == "hstore":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}, nil

	default:
		return nil, fmt.Errorf("%w: %s", errUnsupportedType, typ)
	}
}
