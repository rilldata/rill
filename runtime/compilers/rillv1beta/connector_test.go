package rillv1beta

import (
	"context"
	"fmt"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	_ "github.com/rilldata/rill/runtime/connectors/gcs"
	_ "github.com/rilldata/rill/runtime/connectors/s3"
	"github.com/rilldata/rill/runtime/drivers"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/file"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	"github.com/rilldata/rill/runtime/services/catalog/artifacts"
	"github.com/rilldata/rill/runtime/services/catalog/testutils"
	"github.com/stretchr/testify/require"
)

var AdBidsS3 = "s3://rill-developer.rilldata.io/AdBids.csv.gz"
var AdBidsGCS = "gs://scratch.rilldata.com/rill-developer/AdBids.csv.gz"

func Test_ExtractConnectors(t *testing.T) {
	s, dir := testutils.GetService(t)
	ctx := context.Background()

	require.NoError(t, artifacts.Write(ctx, s.Repo, s.InstID, &drivers.CatalogEntry{
		Name: "AdBidsS3",
		Type: drivers.ObjectTypeSource,
		Path: "sources/AdBidsS3.yaml",
		Object: &runtimev1.Source{
			Name:      "AdBidsS3",
			Connector: "s3",
			Properties: testutils.ToProtoStruct(map[string]any{
				"path": AdBidsS3,
			}),
		},
	}))
	testutils.CreateModel(t, s, "AdBidsGCS", fmt.Sprintf("select * from \"%s\"", AdBidsGCS), "models/AdBidsGCS.sql")

	connectors, err := ExtractConnectors(ctx, dir)
	require.NoError(t, err)
	require.Len(t, connectors, 2)

	var gcs *Connector
	var s3 *Connector

	if connectors[0].Name == "gcs" {
		gcs = connectors[0]
		s3 = connectors[1]
	} else {
		gcs = connectors[1]
		s3 = connectors[0]
	}

	require.Equal(t, "gcs", gcs.Name)
	require.Equal(t, false, gcs.AnonymousAccess)
	require.Equal(t, "s3", s3.Name)
	require.Equal(t, false, s3.AnonymousAccess)
}
