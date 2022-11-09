package server

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	"github.com/stretchr/testify/require"
)

const AdBidsPath = "../../web-local/test/data/AdBids.csv"

func TestServer_MigrateSingleSources(t *testing.T) {
	server, instanceId, err := getTestServer()
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

func getTestServer() (*Server, string, error) {
	metastore, err := drivers.Open("sqlite", "file:rill?mode=memory&cache=shared")
	if err != nil {
		return nil, "", err
	}
	err = metastore.Migrate(context.Background())
	if err != nil {
		return nil, "", err
	}

	server, err := NewServer(&ServerOptions{
		ConnectionCacheSize:  100,
		CatalogCacheSize:     100,
		CatalogCacheDuration: 10 * time.Second,
	}, metastore, nil)
	if err != nil {
		return nil, "", err
	}

	resp, err := server.CreateInstance(context.Background(), &api.CreateInstanceRequest{
		Driver:       "duckdb",
		Dsn:          "",
		Exposed:      true,
		EmbedCatalog: true,
	})
	if err != nil {
		return nil, "", err
	}

	return server, resp.InstanceId, nil
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
