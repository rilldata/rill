package drivers_test

import (
	"context"
	"testing"

	"github.com/rilldata/rill/admin/pkg/pgtestcontainer"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/file"
	_ "github.com/rilldata/rill/runtime/drivers/postgres"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
)

// TestAll runs sub-tests against all drivers.
// This should be the only "real" test in the package. Other tests should be added
// as subtests of TestAll.
func TestAll(t *testing.T) {
	var matrix = []func(t *testing.T, fn func(driver string, instanceID string, cfg map[string]any)) error{
		withDuckDB,
		withFile,
		withPostgres,
		withSQLite,
		// Druid only tested in driver due to complicated ingestion setup
	}

	for _, withDriver := range matrix {
		err := withDriver(t, func(driver string, instanceID string, cfg map[string]any) {
			// Open
			conn, err := drivers.Open(driver, instanceID, cfg, activity.NewNoopClient(), zap.NewNop())
			require.NoError(t, err)
			require.NotNil(t, conn)

			// Migrate
			ctx := context.Background()
			require.NoError(t, conn.Migrate(ctx))
			current, desired, err := conn.MigrationStatus(ctx)
			require.NoError(t, err)
			require.Equal(t, desired, current)

			// Run applicable sub-tests
			if registry, ok := conn.AsRegistry(); ok {
				t.Run("registry_"+driver, func(t *testing.T) { testRegistry(t, registry) })
			}
			if catalog, ok := conn.AsCatalogStore(""); ok {
				t.Run("catalog_"+driver, func(t *testing.T) { testCatalog(t, catalog) })
			}
			if repo, ok := conn.AsRepoStore(""); ok {
				t.Run("repo_"+driver, func(t *testing.T) { testRepo(t, repo) })
			}
			if olap, ok := conn.AsOLAP(""); ok {
				t.Run("olap_"+driver, func(t *testing.T) { testOLAP(t, olap) })
			}

			// Close
			require.NoError(t, conn.Close())
		})
		require.NoError(t, err)
	}
}

func withDuckDB(t *testing.T, fn func(driver string, instanceID string, cfg map[string]any)) error {
	fn("duckdb", "default", map[string]any{"dsn": ":memory:?access_mode=read_write", "pool_size": 4})
	return nil
}

func withFile(t *testing.T, fn func(driver string, instanceID string, cfg map[string]any)) error {
	dsn := t.TempDir()
	fn("file", "default", map[string]any{"dsn": dsn})
	return nil
}

func withPostgres(t *testing.T, fn func(driver string, instanceID string, cfg map[string]any)) error {
	pg := pgtestcontainer.New(t)
	defer pg.Terminate(t)

	fn("postgres", "default", map[string]any{"dsn": pg.DatabaseURL})
	return nil
}

func withSQLite(t *testing.T, fn func(driver string, instanceID string, cfg map[string]any)) error {
	fn("sqlite", "", map[string]any{"dsn": ":memory:"})
	return nil
}
