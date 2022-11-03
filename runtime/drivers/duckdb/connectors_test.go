package duckdb

import (
	"context"
	"fmt"
	"testing"

	"github.com/rilldata/rill/runtime/connectors"
	_ "github.com/rilldata/rill/runtime/connectors/gcs"
	_ "github.com/rilldata/rill/runtime/connectors/s3"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

const testdataFolder = "../../../web-local/test/data/"

func TestConnectorWithSourceVariations(t *testing.T) {
	sources := []struct {
		Connector       string
		Path            string
		AdditionalProps map[string]any
	}{
		{"file", testdataFolder + "AdBids.csv", nil},
		{"file", testdataFolder + "AdBids.csv", map[string]any{"csv.delimiter": ","}},
		{"file", testdataFolder + "AdBids.csv.gz", nil},
		{"file", testdataFolder + "AdBids.parquet", nil},
		// something wrong with this particular file. duckdb fails to extract
		// TODO: move the generator to go and fix the parquet file
		//{"file", testdataFolder + "AdBids.parquet.gz", nil},
		// only enable to do adhoc tests. needs credentials to work
		//{"s3", "s3://rill-developer.rilldata.io/AdBids.csv", nil},
		//{"s3", "s3://rill-developer.rilldata.io/AdBids.csv.gz", nil},
		//{"s3", "s3://rill-developer.rilldata.io/AdBids.parquet", nil},
		//{"s3", "s3://rill-developer.rilldata.io/AdBids.parquet.gz", nil},
		//{"gcs", "gs://scratch.rilldata.com/rill-developer/AdBids.csv", nil},
		//{"gcs", "gs://scratch.rilldata.com/rill-developer/AdBids.csv.gz", nil},
		//{"gcs", "gs://scratch.rilldata.com/rill-developer/AdBids.parquet", nil},
		//{"gcs", "gs://scratch.rilldata.com/rill-developer/AdBids.parquet.gz", nil},
	}

	ctx := context.Background()
	conn, err := driver{}.Open("?access_mode=read_write")
	require.NoError(t, err)
	olap, _ := conn.OLAPStore()

	for _, tt := range sources {
		t.Run(fmt.Sprintf("%s - %s", tt.Connector, tt.Path), func(t *testing.T) {
			var props map[string]any
			if tt.AdditionalProps != nil {
				props = tt.AdditionalProps
			} else {
				props = make(map[string]any)
			}
			props["path"] = tt.Path

			s := &connectors.Source{
				Name:       "foo",
				Connector:  tt.Connector,
				Properties: props,
			}
			err = olap.Ingest(ctx, s)
			require.NoError(t, err)

			var count int
			rows, err := olap.Execute(ctx, &drivers.Statement{Query: "SELECT count(timestamp) FROM foo"})
			require.NoError(t, err)
			require.True(t, rows.Next())
			require.NoError(t, rows.Scan(&count))
			require.Equal(t, 100000, count)
			require.False(t, rows.Next())
			require.NoError(t, rows.Close())
		})
	}
}
