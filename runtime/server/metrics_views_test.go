package server

import (
	"context"
	"testing"

	"github.com/rilldata/rill/runtime/api"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestMetricsView(t *testing.T) {
	// Open a test server
	ctx := context.Background()
	srv, instanceId, err := getTestServer()
	require.NoError(t, err)

	// Create a source
	_, err = srv.MigrateSingle(ctx, &api.MigrateSingleRequest{
		InstanceId: instanceId,
		Sql: `
			CREATE SOURCE ad_bids WITH connector = 'file', path = '../../web-local/test/data/AdBids.csv'
		`,
	})
	require.NoError(t, err)

	// Check source ingested correctly
	r1, err := srv.QueryDirect(ctx, &api.QueryDirectRequest{
		InstanceId: instanceId,
		Sql:        "select count(*) as count from ad_bids",
	})
	require.NoError(t, err)
	require.Equal(t, 100000, int(r1.Data[0].Fields["count"].GetNumberValue()))

	// Create a metrics view
	_, err = srv.MigrateSingle(ctx, &api.MigrateSingleRequest{
		InstanceId: instanceId,
		Sql: `
			CREATE METRICS VIEW bids_metrics
			DIMENSIONS publisher, domain
			MEASURES 
				count(*) AS "count",
				count(distinct domain) as domains,
				sum(bid_price) as total_bid,
				avg(bid_price) as avg_bid
			FROM main.ad_bids
		`,
	})
	require.NoError(t, err)

	// Query the metrics view
	r2, err := srv.Query(ctx, &api.QueryRequest{
		InstanceId: instanceId,
		Sql:        "SELECT publisher, domains, avg_bid FROM bids_metrics",
	})
	require.NoError(t, err)
	require.Equal(t, 5, len(r2.Data))

	// Send a similar query directly against the underlying DB
	r3, err := srv.QueryDirect(ctx, &api.QueryDirectRequest{
		InstanceId: instanceId,
		Sql: `
			SELECT
				publisher,
				count(distinct domain) as "DOMAINS",
				avg(bid_price) as "AVG_BID"
			FROM ad_bids
			GROUP BY publisher
		`,
	})
	require.NoError(t, err)

	// Compare the JSON representations of the metrics view query and the direct query
	j2, err := protojson.Marshal(r2)
	require.NoError(t, err)
	j3, err := protojson.Marshal(r3)
	require.NoError(t, err)
	require.Equal(t, j3, j2)
}
