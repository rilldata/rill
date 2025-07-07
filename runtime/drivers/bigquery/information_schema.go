package bigquery

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/api/iterator"
)

func (c *Connection) ListDatabaseSchemas(ctx context.Context) ([]*drivers.DatabaseSchemaInfo, error) {
	client, err := c.createClient(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get BigQuery client: %w", err)
	}
	defer client.Close()

	var allSchemas []*drivers.DatabaseSchemaInfo
	it := client.Datasets(ctx)
	for {
		ds, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error listing datasets: %w", err)
		}
		allSchemas = append(allSchemas, &drivers.DatabaseSchemaInfo{
			Database:       ds.ProjectID,
			DatabaseSchema: ds.DatasetID,
		})
	}

	return allSchemas, nil
}

func (c *Connection) ListTables(ctx context.Context, database, databaseSchema string) ([]*drivers.TableInfo, error) {
	client, err := c.createClient(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get BigQuery client: %w", err)
	}
	defer client.Close()

	dataset := client.DatasetInProject(database, databaseSchema)
	it := dataset.Tables(ctx)

	var res []*drivers.TableInfo
	for {
		tbl, err := it.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return nil, fmt.Errorf("error listing tables in dataset %s: %w", databaseSchema, err)
		}

		meta, err := tbl.Metadata(ctx)
		if err != nil {
			return nil, fmt.Errorf("error fetching table metadata: %w", err)
		}

		res = append(res, &drivers.TableInfo{
			Name: tbl.TableID,
			View: meta.Type == bigquery.ViewTable,
		})
	}

	return res, nil
}

func (c *Connection) GetTable(ctx context.Context, database, databaseSchema, table string) (*drivers.TableMetadata, error) {
	query := fmt.Sprintf(`
	SELECT 
		column_name,
		data_type
	FROM 
		`+"`%s.%s.INFORMATION_SCHEMA.COLUMNS`"+`
	WHERE  table_name = @table
	ORDER BY 
			ordinal_position`, database, databaseSchema)

	client, err := c.createClient(ctx, database)
	if err != nil {
		return nil, fmt.Errorf("failed to get BigQuery client: %w", err)
	}
	defer client.Close()
	q := client.Query(query)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "table", Value: table},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to run INFORMATION_SCHEMA query: %w", err)
	}

	schemaMap := make(map[string]string)
	var row struct {
		ColumnName string `bigquery:"column_name"`
		DataType   string `bigquery:"data_type"`
	}
	for {
		err := it.Next(&row)
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate over schema rows: %w", err)
		}
		schemaMap[row.ColumnName] = row.DataType
	}

	return &drivers.TableMetadata{
		Schema: schemaMap,
	}, nil
}
