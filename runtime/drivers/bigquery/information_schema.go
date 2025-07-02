package bigquery

import (
	"context"

	"github.com/rilldata/rill/runtime/drivers"
)

func (c *Connection) ListSchemas(ctx context.Context) ([]*drivers.DatabaseSchemaInfo, error) {
	return nil, nil
}

func (c *Connection) ListTables(ctx context.Context, database, schema string) ([]*drivers.TableInfo, error) {
	return nil, nil
}

func (c *Connection) GetTable(ctx context.Context, database, schema, table string) (*drivers.TableMetadata, error) {
	return nil, nil
}
