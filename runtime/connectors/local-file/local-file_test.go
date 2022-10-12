package file_test

import (
	"context"
	"testing"

	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/connectors/sources"
	"github.com/rilldata/rill/runtime/drivers"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
)

func TestLocalFileConnector(t *testing.T) {
	duckdb, err := drivers.Open("duckdb", t.TempDir()+"stage.db")
	if err != nil {
		t.Fatal(err)
	}
	defer duckdb.Close()
	olap, _ := duckdb.OLAPStore()

	connector, _ := connectors.Create(sources.LocalFileConnectorName)

	err = connector.Ingest(context.Background(), sources.Source{
		Name:         "AdBids",
		Connector:    sources.LocalFileConnectorName,
		SamplePolicy: sources.SamplePolicy{},
		Properties: map[string]any{
			"path": "../../../web-local/test/data/AdBids.csv",
		},
	}, olap)
	if err != nil {
		t.Fatal(err)
	}

	_, err = olap.Execute(context.Background(), &drivers.Statement{
		Query:    "SELECT * FROM AdBids LIMIT 1",
		Args:     nil,
		DryRun:   false,
		Priority: 0,
	})
	if err != nil {
		t.Fatal(err)
	}

	// TODO: assert that data is loaded properly
}
