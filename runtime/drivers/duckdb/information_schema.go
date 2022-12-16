package duckdb

import (
	"context"
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
	return &informationSchema{c: c}
}

func (i informationSchema) All(ctx context.Context) ([]*drivers.Table, error) {
	db := <-i.c.pool
	defer func() { i.c.pool <- db }()

	q := `
		select
			coalesce(t.table_catalog, '') as "database",
			t.table_schema as "schema",
			t.table_name as "name",
			t.table_type as "type", 
			array_agg(c.column_name order by c.ordinal_position) as "column_names",
			array_agg(c.data_type order by c.ordinal_position) as "column_types",
			array_agg(c.is_nullable = 'YES' order by c.ordinal_position) as "column_nullable"
		from information_schema.tables t
		join information_schema.columns c on t.table_schema = c.table_schema and t.table_name = c.table_name
		where t.table_schema = 'main'
		group by 1, 2, 3, 4
		order by 1, 2, 3, 4
	`

	rows, err := db.QueryxContext(ctx, q)
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

func (i informationSchema) Lookup(ctx context.Context, name string) (*drivers.Table, error) {
	db := <-i.c.pool
	defer func() { i.c.pool <- db }()

	q := `
		select
			coalesce(t.table_catalog, '') as "database",
			t.table_schema as "schema",
			t.table_name as "name",
			t.table_type as "type", 
			array_agg(c.column_name order by c.ordinal_position) as "column_names",
			array_agg(c.data_type order by c.ordinal_position) as "column_types",
			array_agg(c.is_nullable = 'YES' order by c.ordinal_position) as "column_nullable"
		from information_schema.tables t
		join information_schema.columns c on t.table_schema = c.table_schema and t.table_name = c.table_name
		where t.table_schema = 'main' and t.table_name = ?
		group by 1, 2, 3, 4
		order by 1, 2, 3, 4
	`

	rows, err := db.QueryxContext(ctx, q, name)
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

func (i informationSchema) scanTables(rows *sqlx.Rows) ([]*drivers.Table, error) {
	var res []*drivers.Table

	for rows.Next() {
		var database string
		var schema string
		var name string
		var tableType string
		var columnNames []any
		var columnTypes []any
		var columnNullable []any

		err := rows.Scan(&database, &schema, &name, &tableType, &columnNames, &columnTypes, &columnNullable)
		if err != nil {
			return nil, err
		}

		t := &drivers.Table{
			Database:       database,
			DatabaseSchema: schema,
			Name:           name,
			Schema:         &runtimev1.StructType{},
		}

		// should NEVER happen, but just to be safe
		if len(columnNames) != len(columnTypes) {
			panic(fmt.Errorf("duckdb: column slices have different length"))
		}

		for idx, colName := range columnNames {
			databaseType := columnTypes[idx].(string)
			nullable := columnNullable[idx].(bool)
			colType, err := databaseTypeToPB(databaseType, nullable)
			if err != nil {
				return nil, err
			}

			t.Schema.Fields = append(t.Schema.Fields, &runtimev1.StructType_Field{
				Name: colName.(string),
				Type: colType,
			})
		}

		res = append(res, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func databaseTypeToPB(dbt string, nullable bool) (*runtimev1.Type, error) {
	t := &runtimev1.Type{Nullable: nullable}
	match := true
	switch dbt {
	case "INVALID":
		return nil, fmt.Errorf("encountered invalid duckdb type")
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
	case "UTINYINT":
		t.Code = runtimev1.Type_CODE_UINT8
	case "USMALLINT":
		t.Code = runtimev1.Type_CODE_UINT16
	case "UINTEGER":
		t.Code = runtimev1.Type_CODE_UINT32
	case "UBIGINT":
		t.Code = runtimev1.Type_CODE_UINT64
	case "FLOAT":
		t.Code = runtimev1.Type_CODE_FLOAT32
	case "DOUBLE":
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "TIMESTAMP":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "DATE":
		t.Code = runtimev1.Type_CODE_DATE
	case "TIME":
		t.Code = runtimev1.Type_CODE_TIME
	case "INTERVAL":
		t.Code = runtimev1.Type_CODE_UNSPECIFIED // TODO - Consider adding interval type
	case "HUGEINT":
		t.Code = runtimev1.Type_CODE_INT128
	case "VARCHAR":
		t.Code = runtimev1.Type_CODE_STRING
	case "BLOB":
		t.Code = runtimev1.Type_CODE_BYTES
	case "TIMESTAMP_S":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "TIMESTAMP_MS":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "TIMESTAMP_NS":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "ENUM":
		t.Code = runtimev1.Type_CODE_UNSPECIFIED // TODO - Consider how to handle enums
	case "UUID":
		t.Code = runtimev1.Type_CODE_UUID
	case "JSON":
		t.Code = runtimev1.Type_CODE_JSON
	case "CHAR":
		t.Code = runtimev1.Type_CODE_STRING
	case "TIMESTAMP WITH TIME ZONE":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "TIME WITH TIME ZONE":
		t.Code = runtimev1.Type_CODE_TIME
	case "NULL":
		t.Code = runtimev1.Type_CODE_UNSPECIFIED
	default:
		match = false
	}

	if match {
		return t, nil
	}

	// Handle complex types

	// Handle arrays. They have the format "type[]"
	if strings.HasSuffix(dbt, "[]") {
		at, err := databaseTypeToPB(dbt[0:len(dbt)-2], true)
		if err != nil {
			return nil, err
		}

		t.Code = runtimev1.Type_CODE_ARRAY
		t.ArrayElementType = at
		return t, nil
	}

	// All other complex types have details in parentheses after the type name.
	base, args, ok := splitBaseAndArgs(dbt)
	if !ok {
		return nil, fmt.Errorf("encountered unsupported duckdb type '%s'", dbt)
	}

	switch base {
	// Example: "DECIMAL(10,20)"
	case "DECIMAL":
		t.Code = runtimev1.Type_CODE_DECIMAL
	// Example: "STRUCT(a INT, b INT)"
	case "STRUCT":
		t.Code = runtimev1.Type_CODE_STRUCT
		t.StructType = &runtimev1.StructType{}

		fieldStrs := splitCommasUnlessNestedInParens(args)
		for _, fieldStr := range fieldStrs {
			// Each field has format "name TYPE"
			fieldName, fieldTypeStr, ok := strings.Cut(fieldStr, " ")
			if !ok {
				return nil, fmt.Errorf("encountered unsupported duckdb type '%s'", dbt)
			}

			// Convert to type
			fieldType, err := databaseTypeToPB(fieldTypeStr, true)
			if err != nil {
				return nil, err
			}

			// Add to fields
			t.StructType.Fields = append(t.StructType.Fields, &runtimev1.StructType_Field{
				Name: fieldName,
				Type: fieldType,
			})
		}
	// Example: "MAP(VARCHAR, INT)"
	case "MAP":
		fieldStrs := splitCommasUnlessNestedInParens(args)
		if len(fieldStrs) != 2 {
			return nil, fmt.Errorf("encountered unsupported duckdb type '%s'", dbt)
		}

		keyType, err := databaseTypeToPB(fieldStrs[0], true)
		if err != nil {
			return nil, err
		}

		valType, err := databaseTypeToPB(fieldStrs[1], true)
		if err != nil {
			return nil, err
		}

		t.Code = runtimev1.Type_CODE_MAP
		t.MapType = &runtimev1.MapType{
			KeyType:   keyType,
			ValueType: valType,
		}
	default:
		return nil, fmt.Errorf("encountered unsupported duckdb type '%s'", dbt)
	}

	// Done
	return t, nil
}

// Splits a type with args in parentheses, for example:
//
//	"STRUCT(a INT, b INT)" -> ("STRUCT", "a INT, b INT", true)
func splitBaseAndArgs(s string) (string, string, bool) {
	// Split on opening parenthesis
	base, rest, found := strings.Cut(s, "(")
	if !found {
		return "", "", false
	}

	// Remove closing parenthesis
	rest = rest[0 : len(rest)-1]

	return base, rest, true
}

// Splits a comma-separated list, but ignores commas nested in parentheses.
//
// Examples:
//
//	"10,20" -> ["10", "20"]
//	"foo INT, bar STRUCT(a INT, b INT)" -> ["foo INT", "bar STRUCT(a INT, b INT)"]
func splitCommasUnlessNestedInParens(s string) []string {
	// Result slice
	splits := []string{}
	// Starting idx of current split
	fromIdx := 0
	// Nesting level
	nestCount := 0

	// Consume input character-by-character
	for idx, char := range s {
		// Increase nesting on opening paren
		if char == '(' {
			nestCount++
			continue
		}
		// Decrease nesting on closing paren
		if char == ')' {
			nestCount--
			continue
		}
		// If nested, don't parse for commas
		if nestCount != 0 {
			continue
		}
		// If not nested and there's a comma, add split to result
		if char == ',' {
			splits = append(splits, s[fromIdx:idx])
			fromIdx = idx + 1
			continue
		}
		// If not nested, and there's a space at the start of the split, skip it
		if fromIdx == idx && char == ' ' {
			fromIdx++
			continue
		}
	}

	// Add last split to result and return
	splits = append(splits, s[fromIdx:])
	return splits
}
