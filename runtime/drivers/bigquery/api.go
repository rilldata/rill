package bigquery

import (
	"context"

	"cloud.google.com/go/bigquery"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"google.golang.org/api/iterator"
)

const defaultPageSize = 20

func (c *Connection) ListDatasets(ctx context.Context, req *runtimev1.BigQueryListDatasetsRequest) ([]string, string, error) {
	opts, err := c.clientOption(ctx)
	if err != nil {
		return nil, "", err
	}

	client, err := bigquery.NewClient(ctx, bigquery.DetectProjectID, opts...)
	if err != nil {
		return nil, "", err
	}
	defer client.Close()

	pageSize := int(req.PageSize)
	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	pager := iterator.NewPager(client.Datasets(ctx), pageSize, req.PageToken)
	datasets := make([]*bigquery.Dataset, 0)
	nextToken, err := pager.NextPage(&datasets)
	if err != nil {
		return nil, "", err
	}

	names := make([]string, len(datasets))
	for i := 0; i < len(datasets); i++ {
		names[i] = datasets[i].DatasetID
	}
	return names, nextToken, nil
}

func (c *Connection) ListBigQueryTables(ctx context.Context, req *runtimev1.BigQueryListTablesRequest) ([]string, string, error) {
	opts, err := c.clientOption(ctx)
	if err != nil {
		return nil, "", err
	}

	client, err := bigquery.NewClient(ctx, bigquery.DetectProjectID, opts...)
	if err != nil {
		return nil, "", err
	}
	defer client.Close()

	pageSize := int(req.PageSize)
	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	pager := iterator.NewPager(client.Dataset(req.Dataset).Tables(ctx), pageSize, req.PageToken)
	tables := make([]*bigquery.Table, 0)
	nextToken, err := pager.NextPage(&tables)
	if err != nil {
		return nil, "", err
	}

	names := make([]string, len(tables))
	for i := 0; i < len(tables); i++ {
		names[i] = tables[i].TableID
	}
	return names, nextToken, nil
}
