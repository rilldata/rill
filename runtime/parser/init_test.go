package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitEmptyDefault(t *testing.T) {
	require := require.New(t)

	repo := makeRepo(t, map[string]string{})
	err := InitEmpty(t.Context(), repo, "test-instance", "Test Project", "")
	require.NoError(err)

	// Verify the contents of the rill.yaml file
	rillYAML, err := repo.Get(t.Context(), "rill.yaml")
	require.NoError(err)
	require.Contains(rillYAML, "compiler: ")
	require.Contains(rillYAML, "display_name: Test Project")
	require.Contains(rillYAML, "olap_connector: duckdb")

	// Verify the contents of the .gitignore file
	gitignore, err := repo.Get(t.Context(), ".gitignore")
	require.NoError(err)
	require.Contains(gitignore, ".DS_Store")
	require.Contains(gitignore, "# Rill")
	require.Contains(gitignore, ".env")
	require.Contains(gitignore, "tmp")

	// Verify that NO connector file is created for default (DuckDB) to allow user-guided initialization
	_, err = repo.Get(t.Context(), "connectors/duckdb.yaml")
	require.Error(err) // Should error because file doesn't exist
}

func TestInitEmptyDuckDB(t *testing.T) {
	require := require.New(t)

	repo := makeRepo(t, map[string]string{})
	err := InitEmpty(t.Context(), repo, "test-instance", "Test Project", "duckdb")
	require.NoError(err)

	// Verify the contents of the rill.yaml file
	rillYAML, err := repo.Get(t.Context(), "rill.yaml")
	require.NoError(err)
	require.Contains(rillYAML, "compiler: ")
	require.Contains(rillYAML, "display_name: Test Project")
	require.Contains(rillYAML, "olap_connector: duckdb")

	// Verify the contents of the .gitignore file
	gitignore, err := repo.Get(t.Context(), ".gitignore")
	require.NoError(err)
	require.Contains(gitignore, ".DS_Store")
	require.Contains(gitignore, "# Rill")
	require.Contains(gitignore, ".env")
	require.Contains(gitignore, "tmp")

	// Verify that NO connector file is created for DuckDB to allow user-guided initialization
	_, err = repo.Get(t.Context(), "connectors/duckdb.yaml")
	require.Error(err) // Should error because file doesn't exist
}

func TestInitEmptyCH(t *testing.T) {
	require := require.New(t)

	repo := makeRepo(t, map[string]string{})
	err := InitEmpty(t.Context(), repo, "test-instance", "Test Project", "clickhouse")
	require.NoError(err)

	// Verify the contents of the rill.yaml file
	rillYAML, err := repo.Get(t.Context(), "rill.yaml")
	require.NoError(err)
	require.Contains(rillYAML, "compiler: ")
	require.Contains(rillYAML, "display_name: Test Project")
	require.Contains(rillYAML, "olap_connector: clickhouse")

	// Verify the contents of the .gitignore file
	gitignore, err := repo.Get(t.Context(), ".gitignore")
	require.NoError(err)
	require.Contains(gitignore, ".DS_Store")
	require.Contains(gitignore, "# Rill")
	require.Contains(gitignore, ".env")
	require.Contains(gitignore, "tmp")

	// Verify the contents of the connector file
	connector, err := repo.Get(t.Context(), "connectors/clickhouse.yaml")
	require.NoError(err)
	require.Contains(connector, "type: connector")
	require.Contains(connector, "driver: clickhouse")
	require.Contains(connector, "managed: true")

	// Verify duckdb is not present
	require.NotContains(connector, "driver: duckdb")
}
