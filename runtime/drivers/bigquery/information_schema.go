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
	client, err := c.getClient(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get BigQuery client: %w", err)
	}
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

	client, err := c.getClient(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get BigQuery client: %w", err)
	}

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
			Name:              row.TableName,
			View:              row.TableType == "VIEW",
			IsDefaultDatabase: database == c.config.ProjectID,
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
		CASE t.table_type WHEN 'VIEW' THEN true else false END AS is_view,
		c.column_name,
		c.data_type
	FROM `+"`%s.%s.INFORMATION_SCHEMA.TABLES`"+` AS t
	JOIN `+"`%s.%s.INFORMATION_SCHEMA.COLUMNS`"+` AS c
	ON t.table_name = c.table_name
	WHERE c.table_name = @table
	ORDER BY c.ordinal_position
	`, database, databaseSchema, database, databaseSchema)

	client, err := c.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get BigQuery client: %w", err)
	}
	cq := client.Query(q)
	cq.Parameters = []bigquery.QueryParameter{
		{Name: "table", Value: table},
	}

	it, err := cq.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to run INFORMATION_SCHEMA query: %w", err)
	}

	r := &drivers.TableMetadata{
		Schema: make(map[string]string),
	}
	var row struct {
		IsView     bool   `bigquery:"is_view"`
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
		r.Schema[row.ColumnName] = row.DataType
		r.View = row.IsView
	}

	return r, nil
}

// All implements drivers.InformationSchema.
func (c *Connection) All(ctx context.Context, like string, pageSize uint32, pageToken string) ([]*drivers.OlapTable, string, error) {
	return drivers.AllFromInformationSchema(ctx, like, pageSize, pageToken, c)
}

// LoadPhysicalSize implements drivers.InformationSchema.
func (c *Connection) LoadPhysicalSize(ctx context.Context, tables []*drivers.OlapTable) error {
	return nil
}

// LoadDDL implements drivers.InformationSchema.
func (c *Connection) LoadDDL(ctx context.Context, table *drivers.OlapTable) error {
	client, err := c.getClient(ctx)
	if err != nil {
		return err
	}

	q := fmt.Sprintf("SELECT ddl FROM `%s.%s.INFORMATION_SCHEMA.TABLES` WHERE table_name = @name", table.Database, table.DatabaseSchema)
	cq := client.Query(q)
	cq.Parameters = []bigquery.QueryParameter{
		{Name: "name", Value: table.Name},
	}

	it, err := cq.Read(ctx)
	if err != nil {
		return err
	}

	var row struct {
		DDL string `bigquery:"ddl"`
	}
	err = it.Next(&row)
	if err != nil {
		return err
	}
	table.DDL = row.DDL
	return nil
}

// Lookup implements drivers.InformationSchema.
func (c *Connection) Lookup(ctx context.Context, db, schema, name string) (*drivers.OlapTable, error) {
	client, err := c.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get BigQuery client: %w", err)
	}

	var table *bigquery.Table
	if db != "" {
		table = client.DatasetInProject(db, schema).Table(name)
	} else {
		table = client.Dataset(schema).Table(name)
	}

	meta, err := table.Metadata(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get table metadata: %w", err)
	}
	runtimeSchema, err := fromBQSchema(meta.Schema)
	if err != nil {
		return nil, err
	}
	tbl := &drivers.OlapTable{
		Database:          db,
		DatabaseSchema:    schema,
		Name:              name,
		View:              meta.Type == bigquery.ViewTable,
		Schema:            runtimeSchema,
		UnsupportedCols:   nil, // all columns are currently being mapped though may not be as specific as in BigQuery
		PhysicalSizeBytes: 0,
	}
	return tbl, nil
}
