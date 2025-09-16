package bigquery

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pagination"
	"google.golang.org/api/iterator"
)

func (c *Connection) ListDatabaseSchemas(ctx context.Context, pageSize uint32, pageToken string) ([]*drivers.DatabaseSchemaInfo, string, error) {
	client, err := c.createClient(ctx, "")
	if err != nil {
		return nil, "", fmt.Errorf("failed to get BigQuery client: %w", err)
	}
	defer client.Close()
	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)
	it := client.Datasets(ctx)
	pi := it.PageInfo()
	pi.MaxSize = limit
	if pageToken != "" {
		var startAfter string
		if err := pagination.UnmarshalPageToken(pageToken, &startAfter); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		pi.Token = startAfter
	}

	var res []*drivers.DatabaseSchemaInfo
	count := 0
	for {
		if count >= limit {
			break
		}
		ds, err := it.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return nil, "", fmt.Errorf("error listing datasets: %w", err)
		}
		res = append(res, &drivers.DatabaseSchemaInfo{
			Database:       ds.ProjectID,
			DatabaseSchema: ds.DatasetID,
		})
		count++
	}

	return res, pagination.MarshalPageToken(pi.Token), nil
}

func (c *Connection) ListTables(ctx context.Context, database, databaseSchema string, pageSize uint32, pageToken string) ([]*drivers.TableInfo, string, error) {
	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)
	q := fmt.Sprintf(`
	SELECT
		table_name,
		table_type
		FROM `+"`%s.%s.INFORMATION_SCHEMA.TABLES`"+`
	`, database, databaseSchema)

	var args []bigquery.QueryParameter
	if pageToken != "" {
		var startAfter string
		if err := pagination.UnmarshalPageToken(pageToken, &startAfter); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		q += `
		WHERE table_name > @startAfter
		ORDER BY table_name
		LIMIT @limit
		`
		args = append(args,
			bigquery.QueryParameter{Name: "startAfter", Value: startAfter},
			bigquery.QueryParameter{Name: "limit", Value: limit + 1},
		)
	} else {
		q += `
		ORDER BY table_name
		LIMIT @limit
		`
		args = append(args, bigquery.QueryParameter{Name: "limit", Value: limit + 1})
	}

	client, err := c.createClient(ctx, database)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get BigQuery client: %w", err)
	}
	defer client.Close()

	cq := client.Query(q)
	cq.Parameters = args

	it, err := cq.Read(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to query INFORMATION_SCHEMA.TABLES: %w", err)
	}

	var res []*drivers.TableInfo
	var row struct {
		TableName string `bigquery:"table_name"`
		TableType string `bigquery:"table_type"`
	}

	for {
		err := it.Next(&row)
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, "", fmt.Errorf("failed to iterate over tables: %w", err)
		}
		res = append(res, &drivers.TableInfo{
			Name: row.TableName,
			View: row.TableType == "VIEW",
		})
	}

	next := ""
	if len(res) > limit {
		res = res[:limit]
		next = pagination.MarshalPageToken(res[len(res)-1].Name)
	}
	return res, next, nil
}

func (c *Connection) GetTable(ctx context.Context, database, databaseSchema, table string) (*drivers.TableMetadata, error) {
	q := fmt.Sprintf(`
	SELECT 
		column_name,
		data_type
	FROM `+"`%s.%s.INFORMATION_SCHEMA.COLUMNS`"+`
	WHERE  table_name = @table
	ORDER BY ordinal_position
	`, database, databaseSchema)

	client, err := c.createClient(ctx, database)
	if err != nil {
		return nil, fmt.Errorf("failed to get BigQuery client: %w", err)
	}
	defer client.Close()
	cq := client.Query(q)
	cq.Parameters = []bigquery.QueryParameter{
		{Name: "table", Value: table},
	}

	it, err := cq.Read(ctx)
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
