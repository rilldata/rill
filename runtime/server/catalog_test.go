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

	instId, repoId := createInstanceAndRepo(t, server, ctx, dir)
	service, err := server.serviceCache.createCatalogService(ctx, server, instId, repoId)
	require.NoError(t, err)

	testutils.CreateSource(t, service, "AdBids", AdBidsCsvPath, AdBidsRepoPath)
	testutils.CreateModel(t, service, "AdBids_model", "select timestamp, publisher from AdBids", AdBidsModelRepoPath)
	migrateResp, err := service.Migrate(ctx, catalog.MigrationConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, migrateResp, 0, 2, 0, 0)
	testutils.AssertTable(t, service, "AdBids", AdBidsRepoPath)
	testutils.AssertTable(t, service, "AdBids_model", AdBidsModelRepoPath)

	// create a new service and make sure DAG is generated
	instId, repoId = createInstanceAndRepo(t, server, ctx, dir)
	service, err = server.serviceCache.createCatalogService(ctx, server, instId, repoId)
	require.NoError(t, err)

	// initial migrate to setup cache
	migrateResp, err = service.Migrate(ctx, catalog.MigrationConfig{})
	// force update the source
	migrateResp, err = service.Migrate(ctx, catalog.MigrationConfig{
		ChangedPaths: []string{AdBidsRepoPath},
		ForcedPaths:  []string{AdBidsRepoPath},
	})
	require.NoError(t, err)
	testutils.AssertMigration(t, migrateResp, 0, 0, 2, 0)
}

func createInstanceAndRepo(t *testing.T, server *Server, ctx context.Context, dir string) (string, string) {
	instResp, err := server.CreateInstance(ctx, &api.CreateInstanceRequest{
		Driver: "duckdb",
		// use persistent file to test fresh load
		Dsn:          filepath.Join(dir, "stage.db"),
		Exposed:      true,
		EmbedCatalog: true,
	})
	require.NoError(t, err)

	repoResp, err := server.CreateRepo(ctx, &api.CreateRepoRequest{
		Driver: "file",
		Dsn:    dir,
	})
	require.NoError(t, err)

	return instResp.Instance.InstanceId, repoResp.Repo.RepoId
}
