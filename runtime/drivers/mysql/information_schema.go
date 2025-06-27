package mysql

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

var errUnsupportedType = errors.New("encountered unsupported mysql type")

func (c *connection) ListDatabases(ctx context.Context) ([]string, error) {
	return []string{"def"}, nil
}

func (c *connection) ListSchemas(ctx context.Context, database string) ([]string, error) {
	q := `
	SELECT schema_name FROM information_schema.schemata
	WHERE schema_name not in ('information_schema', 'performance_schema', 'sys')
	ORDER BY schema_name
	`

	db, err := c.getDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryxContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return names, nil
}

func (c *connection) All(ctx context.Context, like string) ([]*drivers.Table, error) {
	var likeClause string
	var args []any
	if like != "" {
		likeClause = "AND (LOWER(T.table_name) LIKE LOWER(?) OR CONCAT(T.table_schema, '.', T.table_name) LIKE LOWER(?))"
		args = []any{like, like}
	}
	q := fmt.Sprintf(`
		SELECT 
			T.table_catalog AS `+"`database`"+`,
			T.table_schema AS `+"`database_schema`"+`,
			true AS is_default_database,
			T.table_schema = DATABASE() AS is_default_database_schema,
			T.table_name AS name,
			T.table_type,
			C.column_name AS columns,
			C.data_type AS column_type
		FROM information_schema.tables T
		JOIN information_schema.columns C ON T.table_catalog = C.table_catalog AND T.table_schema = C.table_schema AND T.table_name = C.table_name
		WHERE T.table_schema NOT IN ('information_schema', 'mysql', 'performance_schema', 'sys')
		%s
		ORDER BY `+"`database`"+`, `+"`database_schema`"+`, name, table_type
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
			T.table_catalog AS ` + "`database`" + `,
			T.table_schema AS ` + "`database_schema`" + `,
			true AS is_default_database,
			T.table_schema = DATABASE() AS is_default_database_schema,
			T.table_name AS name,
			T.table_type,
			C.column_name AS columns,
			C.data_type AS column_type
		FROM information_schema.tables T
		JOIN information_schema.columns C ON T.table_catalog = C.table_catalog AND T.table_schema = C.table_schema AND T.table_name = C.table_name
		WHERE T.table_catalog = COALESCE(?, 'def') AND T.table_schema = COALESCE(?, DATABASE()) AND T.table_name = ?
		ORDER BY ` + "`database`" + `, ` + "`database_schema`" + `, name, table_type, ordinal_position
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

func databaseTypeToPB(mysqlType string) (*runtimev1.Type, error) {
	typ := strings.ToLower(strings.TrimSpace(mysqlType))

	switch {
	case typ == "boolean", typ == "bool", typ == "tinyint(1)":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_BOOL}, nil
	case typ == "tinyint":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT8}, nil
	case typ == "smallint":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT16}, nil
	case typ == "int", typ == "integer", typ == "mediumint":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT32}, nil
	case typ == "bigint":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT64}, nil
	case strings.HasPrefix(typ, "decimal"), strings.HasPrefix(typ, "numeric"):
		return &runtimev1.Type{Code: runtimev1.Type_CODE_DECIMAL}, nil
	case typ == "float":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_FLOAT32}, nil
	case typ == "double", typ == "double precision", typ == "real":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_FLOAT64}, nil
	case strings.HasPrefix(typ, "varchar"), strings.HasPrefix(typ, "char"),
		typ == "text", typ == "tinytext", typ == "mediumtext", typ == "longtext":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}, nil
	case typ == "binary", typ == "varbinary", typ == "blob",
		typ == "tinyblob", typ == "mediumblob", typ == "longblob":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_BYTES}, nil
	case typ == "json":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_JSON}, nil
	case typ == "date":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_DATE}, nil
	case typ == "time":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_TIME}, nil
	case strings.HasPrefix(typ, "timestamp"), strings.HasPrefix(typ, "datetime"):
		return &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP}, nil
	case typ == "year":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT32}, nil
	case typ == "enum", typ == "set":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}, nil
	case typ == "bit":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_BOOL}, nil
	case typ == "geometry", typ == "point", typ == "linestring", typ == "polygon",
		typ == "multipoint", typ == "multilinestring", typ == "multipolygon", typ == "geomcollection":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}, nil
	default:
		return nil, fmt.Errorf("%w: %s", errUnsupportedType, typ)
	}
}
