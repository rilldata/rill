package local_file_test

import (
	"context"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/connectors/sources"
	"github.com/rilldata/rill/runtime/drivers"
	test_utils "github.com/rilldata/rill/runtime/test-utils"
	"os"
	"path/filepath"
	"testing"
)

var curPath, _ = os.Getwd()
var TestFilesPath = filepath.Join(curPath, "/../../../web-local/test/data/")

func TestLocalFileConnector(t *testing.T) {
	duckdb, err := test_utils.GetDuckdbDriver("stage.db")
	if err != nil {
		t.Fatal(err)
	}
	olap, _ := duckdb.OLAPStore()

	connector, _ := connectors.Create(sources.LocalFileConnectorName)

	_, err = connector.Ingest(context.Background(), sources.Source{
		Name:         "AdBids",
		Connector:    sources.LocalFileConnectorName,
		SamplePolicy: sources.SamplePolicy{},
		Properties: map[string]any{
			"path": TestFilesPath + "/AdBids.csv",
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
