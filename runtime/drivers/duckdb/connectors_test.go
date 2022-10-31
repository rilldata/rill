package duckdb

import (
	"context"
	"testing"

	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func TestFileConnector(t *testing.T) {
	ctx := context.Background()
	conn, err := driver{}.Open("?access_mode=read_write")
	require.NoError(t, err)
	olap, _ := conn.OLAPStore()

	s := &connectors.Source{
		Name:      "foo",
		Connector: "file",
		Properties: map[string]any{
			"path": "../../../web-local/test/data/AdBids.csv",
		},
	}

	err = olap.Ingest(ctx, s)
	require.NoError(t, err)

	assertAdBidsTable(t, ctx, olap)

	s = &connectors.Source{
		Name:      "foo",
		Connector: "file",
		Properties: map[string]any{
			"path":          "../../../web-local/test/data/AdBids.csv",
			"csv.delimiter": ",",
		},
	}

	err = olap.Ingest(ctx, s)
	require.NoError(t, err)

	assertAdBidsTable(t, ctx, olap)
}

func TestFileConnectorWithGzip(t *testing.T) {
	ctx := context.Background()
	conn, err := driver{}.Open("?access_mode=read_write")
	require.NoError(t, err)
	olap, _ := conn.OLAPStore()

	s := &connectors.Source{
		Name:      "foo",
		Connector: "file",
		Properties: map[string]any{
			"path": "../../../web-local/test/data/AdBids.csv.gz",
		},
	}
	err = olap.Ingest(ctx, s)
	require.NoError(t, err)
}

func assertAdBidsTable(t *testing.T, ctx context.Context, olap drivers.OLAPStore) {
	var count int
	rows, err := olap.Execute(ctx, &drivers.Statement{Query: "SELECT count(timestamp) FROM foo"})
	require.NoError(t, err)
	require.True(t, rows.Next())
	require.NoError(t, rows.Scan(&count))
	require.Equal(t, 100000, count)
	require.False(t, rows.Next())
	require.NoError(t, rows.Close())
}
