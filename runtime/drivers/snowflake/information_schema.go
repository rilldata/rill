package snowflake

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
)

func (c *connection) ListDatabaseSchemas(ctx context.Context) ([]*drivers.DatabaseSchemaInfo, error) {
	// in snowflake it show all the schemas in all the databases
	q := "SHOW SCHEMAS"
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

	var names []*drivers.DatabaseSchemaInfo
	for rows.Next() {
		cols, _ := rows.Columns()
		vals := make([]interface{}, len(cols))
		for i := range vals {
			var v interface{}
			vals[i] = &v
		}

		if err := rows.Scan(vals...); err != nil {
			return nil, err
		}
		nameVal := *(vals[1].(*interface{})) // column at index 1 is "name"
		nameStr, ok := nameVal.(string)
		if !ok {
			return nil, fmt.Errorf("expected schema name as string got %T", nameVal)
		}

		databaseNameVal := *(vals[4].(*interface{})) // column at index 4 is "name"
		databaseNameStr, ok := databaseNameVal.(string)
		if !ok {
			return nil, fmt.Errorf("expected database name as string got %T", nameVal)
		}
		names = append(names, &drivers.DatabaseSchemaInfo{
			Database:       databaseNameStr,
			DatabaseSchema: nameStr,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return names, nil
}

func (c *connection) ListTables(ctx context.Context, database, schema string) ([]*drivers.TableInfo, error) {
	return nil, nil
}

func (c *connection) GetTable(ctx context.Context, database, schema, table string) (*drivers.TableMetadata, error) {
	return nil, nil
}
