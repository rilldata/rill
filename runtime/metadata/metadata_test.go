package metadata_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/rilldata/rill/runtime/metadata"
	_ "github.com/rilldata/rill/runtime/metadata/postgres"
	_ "github.com/rilldata/rill/runtime/metadata/sqlite"
)

// TestAll st creates one of every driver and run
func TestAll(t *testing.T) {
	ctx := context.Background()

	// Start a Postgres test container
	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:14",
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor:   wait.ForListeningPort("5432/tcp"),
			Env: map[string]string{
				"POSTGRES_USER":     "postgres",
				"POSTGRES_PASSWORD": "postgres",
				"POSTGRES_DB":       "postgres",
			},
		},
	})
	require.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	// Get Postgres database URL
	host, err := pgContainer.Host(ctx)
	require.NoError(t, err)
	port, err := pgContainer.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)
	pgURL := fmt.Sprintf("postgres://postgres:postgres@%s:%d/postgres", host, port.Int())

	// Open Postgres driver
	pgDB, err := metadata.Open("postgres", pgURL)
	require.NoError(t, err)
	require.NotNil(t, pgDB)
	require.NoError(t, pgDB.Migrate(ctx))

	// Open SQLite driver
	sqliteDB, err := metadata.Open("sqlite", ":memory:")
	require.NoError(t, err)
	require.NotNil(t, sqliteDB)
	require.NoError(t, sqliteDB.Migrate(ctx))

	dbs := map[string]metadata.DB{
		"Postgres": pgDB,
		"SQLite":   sqliteDB,
	}

	// Run test matrix
	for driver, db := range dbs {
		t.Run("TestMigrations"+driver, func(t *testing.T) { testMigrations(t, db) })
		// Add new tests here

		require.NoError(t, db.Close())
	}
}

func testMigrations(t *testing.T, db metadata.DB) {
	ctx := context.Background()
	version, err := db.FindMigrationVersion(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, version)
}
