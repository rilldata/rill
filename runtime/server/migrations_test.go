package server

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/api"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const AdBidsPath = "../../web-local/test/data/AdBids.csv"

func TestServer_MigrateSingleSources(t *testing.T) {
	server, instanceId, err := getTestServer(t)
	require.NoError(t, err)

	_, err = server.MigrateSingle(context.Background(), &api.MigrateSingleRequest{
		InstanceId: instanceId,
		Sql:        fmt.Sprintf("create source AdBids with connector = 'file', path = '%s'", AdBidsPath),
		DryRun:     false,
	})
	require.NoError(t, err)
	assertTablePresence(t, server, instanceId, "AdBids", 100000)

	_, err = server.MigrateSingle(context.Background(), &api.MigrateSingleRequest{
		InstanceId: instanceId,
		Sql:        fmt.Sprintf("create source adbids with connector = 'file', path = '%s'", AdBidsPath),
		DryRun:     false,
	})
	require.Error(t, err)

	_, err = server.MigrateSingle(context.Background(), &api.MigrateSingleRequest{
		InstanceId:      instanceId,
		Sql:             fmt.Sprintf("create source adbids with connector = 'file', path = '%s'", AdBidsPath),
		DryRun:          false,
		CreateOrReplace: true,
	})
	require.NoError(t, err)
	assertTablePresence(t, server, instanceId, "adbids", 100000)

	_, err = server.MigrateSingle(context.Background(), &api.MigrateSingleRequest{
		InstanceId:      instanceId,
		Sql:             fmt.Sprintf("create source AdBids_New with connector = 'file', path = '%s'", AdBidsPath),
		DryRun:          false,
		CreateOrReplace: true,
		RenameFrom:      "AdBids",
	})
	require.NoError(t, err)
	assertTablePresence(t, server, instanceId, "AdBids_New", 100000)
}

func assertTablePresence(t *testing.T, server *Server, instanceId string, tableName string, count int) {
	resp, err := server.QueryDirect(context.Background(), &api.QueryDirectRequest{
		InstanceId: instanceId,
		Sql:        fmt.Sprintf("select count(*) as count from %s", tableName),
		Args:       nil,
		Priority:   0,
		DryRun:     false,
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.Data)
	require.Equal(t, int(resp.Data[0].Fields["count"].GetNumberValue()), count)

	catalog, _ := server.GetCatalogObject(context.Background(), &api.GetCatalogObjectRequest{
		InstanceId: instanceId,
		Name:       tableName,
	})
	require.WithinDuration(t, time.Now(), catalog.GetObject().RefreshedOn.AsTime(), time.Second)
}
