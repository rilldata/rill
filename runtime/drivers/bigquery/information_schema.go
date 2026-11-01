package bigquery

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"cloud.google.com/go/bigquery"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/pagination"
	"go.uber.org/zap"
	bqv2 "google.golang.org/api/bigquery/v2"
	"google.golang.org/api/iterator"
)

func (c *Connection) ListDatabaseSchemas(ctx context.Context, pageSize uint32, pageToken string) ([]*drivers.DatabaseSchemaInfo, string, error) {
	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)

	// The page token is the project to resume from and the dataset page token within it.
	var tokProject, dsToken string
	if pageToken != "" {
		if err := pagination.UnmarshalPageToken(pageToken, &tokProject, &dsToken); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
	}

	client, err := c.getClient(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get BigQuery client: %w", err)
	}

	// List all projects the credentials have access to.
	// If listing fails, we still list datasets in the client's project.
	var projectIDs []string
	opts, err := c.clientOption(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get Google API client options: %w", err)
	}
	svc, err := bqv2.NewService(ctx, opts...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create BigQuery service: %w", err)
	}
	err = svc.Projects.List().Pages(ctx, func(page *bqv2.ProjectList) error {
		for _, p := range page.Projects {
			projectIDs = append(projectIDs, p.ProjectReference.ProjectId)
		}
		return nil
	})
	if err != nil {
		c.logger.Debug("failed to list projects", zap.Error(err), observability.ZapCtx(ctx))
	}

	projectIDs = append(projectIDs, client.Project()) // always append client's project in case listing projects failed or return an empty list (when having dataset only access)
	slices.Sort(projectIDs)
	projectIDs = slices.Compact(projectIDs) // compact to ensure client.Project() is not duplicated

	// Resume from the token's project, or the next one if it no longer exists.
	i, found := slices.BinarySearch(projectIDs, tokProject)
	if !found {
		dsToken = ""
	}

	var res []*drivers.DatabaseSchemaInfo
	for i < len(projectIDs) {
		if len(res) == limit {
			return res, pagination.MarshalPageToken(projectIDs[i], dsToken), nil
		}

		it := client.Datasets(ctx)
		it.ProjectID = projectIDs[i]
		var page []*bigquery.Dataset
		next, err := iterator.NewPager(it, limit-len(res), dsToken).NextPage(&page)
		if err != nil {
			if ctx.Err() != nil {
				return nil, "", ctx.Err()
			}
			// Skip projects we can't list datasets in.
			c.logger.Debug("failed to list datasets", zap.String("project", projectIDs[i]), zap.Error(err), observability.ZapCtx(ctx))
			i++
			dsToken = ""
			continue
		}
		for _, ds := range page {
			res = append(res, &drivers.DatabaseSchemaInfo{
				Database:       ds.ProjectID,
				DatabaseSchema: ds.DatasetID,
			})
		}
		dsToken = next
		if next == "" {
			i++
		}
	}
	return res, "", nil
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
			Name: row.TableName,
			View: row.TableType == "VIEW" || row.TableType == "MATERIALIZED VIEW",
		})
	}

	next := ""
	if len(res) > limit {
		res = res[:limit]
		next = pagination.MarshalPageToken(res[len(res)-1].Name)
	}
	return res, next, nil
}

func (c *Connection) Lookup(ctx context.Context, database, databaseSchema, name string) (*drivers.OlapTable, error) {
	client, err := c.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get BigQuery client: %w", err)
	}

	var table *bigquery.Table
	if database != "" {
		table = client.DatasetInProject(database, databaseSchema).Table(name)
	} else {
		table = client.Dataset(databaseSchema).Table(name)
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
		Database:          database,
		DatabaseSchema:    databaseSchema,
		Name:              name,
		View:              meta.Type == bigquery.ViewTable || meta.Type == bigquery.MaterializedView,
		Schema:            runtimeSchema,
		UnsupportedCols:   nil, // all columns are currently being mapped though may not be as specific as in BigQuery
		PhysicalSizeBytes: 0,
	}
	return tbl, nil
}

// All implements drivers.OLAPInformationSchema.
func (c *Connection) All(ctx context.Context, like string, pageSize uint32, pageToken string) ([]*drivers.OlapTable, string, error) {
	return drivers.AllFromInformationSchema(ctx, like, pageSize, pageToken, c)
}

// LoadPhysicalSize implements drivers.OLAPInformationSchema.
func (c *Connection) LoadPhysicalSize(ctx context.Context, tables []*drivers.OlapTable) error {
	return nil
}

// LoadDDL implements drivers.OLAPInformationSchema.
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
