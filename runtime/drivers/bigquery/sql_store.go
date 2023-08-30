package bigquery

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

// Query implements drivers.SQLStore
func (c *Connection) Query(ctx context.Context, props map[string]any, sql string) (drivers.RowIterator, error) {
	srcProps, err := parseSourceProperties(props)
	if err != nil {
		return nil, err
	}

	client, err := c.createClient(ctx, srcProps)
	if err != nil {
		if strings.Contains(err.Error(), "unable to detect projectID") {
			return nil, fmt.Errorf("projectID not detected in credentials. Please set `project_id` in source yaml")
		}
		return nil, fmt.Errorf("failed to create bigquery client: %w", err)
	}

	if err := client.EnableStorageReadClient(ctx); err != nil {
		client.Close()
		return nil, err
	}

	now := time.Now()
	q := client.Query(sql)
	it, err := q.ReadAsArrowObjects(ctx)
	if err != nil && !strings.Contains(err.Error(), "Syntax error") {
		// close the read storage API client
		client.Close()
		c.logger.Info("query failed, retrying without storage api", zap.Error(err))
		// the query results are always cached in a temporary table that storage api can use
		// there are some exceptions when results aren't cached
		// so we also try without storage api
		client, err = c.createClient(ctx, srcProps)
		if err != nil {
			return nil, fmt.Errorf("failed to create bigquery client: %w", err)
		}

		q := client.Query(sql)
		it, err = q.ReadAsArrowObjects(ctx)
	}
	if err != nil {
		client.Close()
		return nil, err
	}
	c.logger.Info("query took", zap.Duration("duration", time.Since(now)))

	return &rowIterator{
		client: client,
		bqIter: it,
	}, nil
}

type rowIterator struct {
	client *bigquery.Client
	bqIter *bigquery.ArrowIterator
	logger *zap.Logger
}

var _ drivers.RowIterator = &rowIterator{}

func (r *rowIterator) Schema(ctx context.Context) (*runtimev1.StructType, error) {
	return nil, drivers.ErrNotImplemented
}

func (r *rowIterator) Next(ctx context.Context) ([]any, error) {
	return nil, drivers.ErrNotImplemented
}

func (r *rowIterator) Close() error {
	return r.client.Close()
}

func (r *rowIterator) Size(unit drivers.ProgressUnit) (uint64, bool) {
	if unit == drivers.ProgressUnitRecord {
		return r.bqIter.TotalRows, true
	}

	return 0, false
}
