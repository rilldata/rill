package postgres

import (
	"context"

	"github.com/rilldata/rill/runtime/drivers"
)

func (c *connection) ListSchemas(ctx context.Context) ([]*drivers.DatabaseSchemaInfo, error) {
	return nil, nil
}

func (c *connection) ListTables(ctx context.Context, database, schema string) ([]*drivers.TableInfo, error) {
	return nil, nil
}

func (c *connection) GetTable(ctx context.Context, database, schema, table string) (*drivers.TableMetadata, error) {
	return nil, nil
}
