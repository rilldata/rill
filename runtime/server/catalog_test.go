package server

import (
	"context"
	"path/filepath"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/file"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	"github.com/rilldata/rill/runtime/services/catalog"
	_ "github.com/rilldata/rill/runtime/services/catalog/artifacts/sql"
	_ "github.com/rilldata/rill/runtime/services/catalog/artifacts/yaml"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/metrics_views"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/models"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/sources"
	"github.com/rilldata/rill/runtime/services/catalog/testutils"
	"github.com/stretchr/testify/require"
)

func TestServer_InitCatalogService(t *testing.T) {
	server, _ := getTestServer(t)

	ctx := context.Background()
	dir := t.TempDir()

	instId := createInstance(t, server, ctx, dir)
	service, err := server.serviceCache.createCatalogService(ctx, server, instId)
	require.NoError(t, err)

	testutils.CreateSource(t, service, "AdBids", AdBidsCsvPath, AdBidsRepoPath)
	testutils.CreateModel(t, service, "AdBids_model", "select timestamp, publisher from AdBids", AdBidsModelRepoPath)
	migrateResp, err := service.Migrate(ctx, catalog.MigrationConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, migrateResp, 0, 2, 0, 0, []string{AdBidsRepoPath, AdBidsModelRepoPath})
	testutils.AssertTable(t, service, "AdBids", AdBidsRepoPath)
	testutils.AssertTable(t, service, "AdBids_model", AdBidsModelRepoPath)

	// create a new service and make sure DAG is generated
	instId = createInstance(t, server, ctx, dir)
	service, err = server.serviceCache.createCatalogService(ctx, server, instId)
	require.NoError(t, err)

	// initial migrate to setup cache
	migrateResp, err = service.Migrate(ctx, catalog.MigrationConfig{})
	// force update the source
	migrateResp, err = service.Migrate(ctx, catalog.MigrationConfig{
		ChangedPaths: []string{AdBidsRepoPath},
		ForcedPaths:  []string{AdBidsRepoPath},
	})
	require.NoError(t, err)
	testutils.AssertMigration(t, migrateResp, 0, 0, 2, 0, []string{AdBidsRepoPath, AdBidsModelRepoPath})
}

func createInstance(t *testing.T, server *Server, ctx context.Context, dir string) string {
	instResp, err := server.CreateInstance(ctx, &runtimev1.CreateInstanceRequest{
		OlapDriver: "duckdb",
		// use persistent file to test fresh load
		OlapDsn:      filepath.Join(dir, "stage.db"),
		RepoDriver:   "file",
		RepoDsn:      dir,
		EmbedCatalog: true,
	})
	require.NoError(t, err)

	return instResp.Instance.InstanceId
}
