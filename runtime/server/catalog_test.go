package server

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/rilldata/rill/runtime/api"
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
	server, _, err := getTestServer(t)
	require.NoError(t, err)

	ctx := context.Background()

	dir := t.TempDir()
	serviceResp, err := server.InitCatalogService(ctx, &api.InitCatalogServiceRequest{
		Instance: &api.CreateInstanceRequest{
			Driver:       "duckdb",
			Dsn:          filepath.Join(dir, "stage.db"),
			Exposed:      true,
			EmbedCatalog: true,
		},
		Repo: &api.CreateRepoRequest{
			Driver: "file",
			Dsn:    dir,
		},
	})
	require.NoError(t, err)
	service, err := server.serviceCache.createCatalogService(ctx, server, serviceResp.Instance.InstanceId, serviceResp.Repo.RepoId)
	require.NoError(t, err)

	testutils.CreateSource(t, service, "AdBids", AdBidsCsvPath, AdBidsRepoPath)
	testutils.CreateModel(t, service, "AdBids_model", "select timestamp, publisher from AdBids", AdBidsModelRepoPath)
	migrateResp, err := service.Migrate(ctx, catalog.MigrationConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, migrateResp, 0, 2, 0, 0)
	testutils.AssertTable(t, service, "AdBids", AdBidsRepoPath)
	testutils.AssertTable(t, service, "AdBids_model", AdBidsModelRepoPath)

	// create a new service and make sure DAG is generated
	serviceResp, err = server.InitCatalogService(ctx, &api.InitCatalogServiceRequest{
		Instance: &api.CreateInstanceRequest{
			Driver:       "duckdb",
			Dsn:          filepath.Join(dir, "stage.db"),
			Exposed:      true,
			EmbedCatalog: true,
		},
		Repo: &api.CreateRepoRequest{
			Driver: "file",
			Dsn:    dir,
		},
	})
	require.NoError(t, err)
	service, err = server.serviceCache.createCatalogService(ctx, server, serviceResp.Instance.InstanceId, serviceResp.Repo.RepoId)
	require.NoError(t, err)

	// force update the source
	migrateResp, err = service.Migrate(ctx, catalog.MigrationConfig{
		ChangedPaths: []string{AdBidsRepoPath},
		ForcedPaths:  []string{AdBidsRepoPath},
	})
	require.NoError(t, err)
	testutils.AssertMigration(t, migrateResp, 0, 0, 2, 0)
}
