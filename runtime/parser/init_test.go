package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitDefaultOLAP(t *testing.T) {
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

	duckdbYAML, err := repo.Get(t.Context(), "connectors/duckdb.yaml")
	require.NoError(err)
	require.Contains(duckdbYAML, "type: connector")
	require.Contains(duckdbYAML, "driver: duckdb")
}

func TestInitWithDuckDB(t *testing.T) {
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
	duckdbYAML, err := repo.Get(t.Context(), "connectors/duckdb.yaml")
	require.NoError(err)
	require.Contains(duckdbYAML, "type: connector")
	require.Contains(duckdbYAML, "driver: duckdb")
	require.Contains(duckdbYAML, "managed: true")
}

func TestInitWithClickHouse(t *testing.T) {
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

	// Verify the contents of the connectors/clickhouse.yaml file
	clickhouseYAML, err := repo.Get(t.Context(), "connectors/clickhouse.yaml")
	require.NoError(err)
	require.Contains(clickhouseYAML, "type: connector")
	require.Contains(clickhouseYAML, "driver: clickhouse")
	require.Contains(clickhouseYAML, "managed: true")

	// Verify that duckdb.yaml is NOT created for ClickHouse projects
	_, err = repo.Get(t.Context(), "connectors/duckdb.yaml")
	require.Error(err)
}
