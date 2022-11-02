package duckdb

import (
	"context"
	"testing"

	"github.com/rilldata/rill/runtime/connectors"
	_ "github.com/rilldata/rill/runtime/connectors/gcs"
	_ "github.com/rilldata/rill/runtime/connectors/s3"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

const pathPrefix = "../../../web-local/test/data/"

func TestFileConnector(t *testing.T) {
	ctx := context.Background()
	conn, err := driver{}.Open("?access_mode=read_write")
	require.NoError(t, err)
	olap, _ := conn.OLAPStore()

	s := &connectors.Source{
		Name:      "foo",
		Connector: "file",
		Properties: map[string]any{
			"path": pathPrefix + "AdBids.csv",
		},
	}

	err = olap.Ingest(ctx, s)
	require.NoError(t, err)

	assertAdBidsTable(t, ctx, olap)

	s = &connectors.Source{
		Name:      "foo",
		Connector: "file",
		Properties: map[string]any{
			"path":          pathPrefix + "AdBids.csv",
			"csv.delimiter": ",",
		},
	}

	err = olap.Ingest(ctx, s)
	require.NoError(t, err)

	assertAdBidsTable(t, ctx, olap)
}

func TestConnectorWithSourceVariations(t *testing.T) {
	sources := []struct {
		Connector string
		Path      string
	}{
		{"file", pathPrefix + "AdBids.csv"},
		{"file", pathPrefix + "AdBids.csv.gz"},
		{"file", pathPrefix + "AdBids.parquet"},
		// something wrong with this particular file. duckdb fails to extract
		//{"file", pathPrefix + "AdBids.parquet.gz"},
		// only enable to do adhoc tests. needs credentials to work
		//{"s3", "s3://rill-developer.rilldata.io/AdBids.csv"},
		//{"s3", "s3://rill-developer.rilldata.io/AdBids.csv.gz"},
		//{"s3", "s3://rill-developer.rilldata.io/AdBids.parquet"},
		//{"s3", "s3://rill-developer.rilldata.io/AdBids.parquet.gz"},
		//{"gcs", "gs://scratch.rilldata.com/rill-developer/AdBids.csv"},
		//{"gcs", "gs://scratch.rilldata.com/rill-developer/AdBids.csv.gz"},
		//{"gcs", "gs://scratch.rilldata.com/rill-developer/AdBids.parquet"},
		//{"gcs", "gs://scratch.rilldata.com/rill-developer/AdBids.parquet.gz"},
	}

	ctx := context.Background()
	conn, err := driver{}.Open("?access_mode=read_write")
	require.NoError(t, err)
	olap, _ := conn.OLAPStore()

	for _, tt := range sources {
		s := &connectors.Source{
			Name:      "foo",
			Connector: tt.Connector,
			Properties: map[string]any{
				"path": tt.Path,
			},
		}
		err = olap.Ingest(ctx, s)
		require.NoError(t, err)

		assertAdBidsTable(t, ctx, olap)
	}
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
