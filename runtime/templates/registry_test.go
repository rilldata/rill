package templates

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRegistry(t *testing.T) {
	r, err := NewRegistry()
	require.NoError(t, err)
	require.NotNil(t, r)

	// Verify all definitions loaded
	all := r.List()
	require.Greater(t, len(all), 25, "expected at least 25 template definitions")
}

func TestRegistryGet(t *testing.T) {
	r, err := NewRegistry()
	require.NoError(t, err)

	// Known templates exist
	for _, name := range []string{"s3", "gcs", "clickhouse", "s3-duckdb", "iceberg-duckdb", "snowflake-duckdb"} {
		t.Run(name, func(t *testing.T) {
			tmpl, ok := r.Get(name)
			require.True(t, ok, "template %q should exist", name)
			require.Equal(t, name, tmpl.Name)
			require.NotEmpty(t, tmpl.DisplayName)
			require.NotEmpty(t, tmpl.Files)
		})
	}

	// Unknown template doesn't exist
	_, ok := r.Get("nonexistent")
	require.False(t, ok)
}

func TestRegistryListByTags(t *testing.T) {
	r, err := NewRegistry()
	require.NoError(t, err)

	// Filter by "duckdb" should return DuckDB-related templates
	duckdbTemplates := r.ListByTags([]string{"duckdb"})
	require.Greater(t, len(duckdbTemplates), 5, "expected several duckdb-tagged templates")
	for _, tmpl := range duckdbTemplates {
		require.Contains(t, tmpl.Tags, "duckdb")
	}

	// Filter by "olap" + "connector" should return OLAP connector templates
	olapConnectors := r.ListByTags([]string{"olap", "connector"})
	require.Greater(t, len(olapConnectors), 0)
	for _, tmpl := range olapConnectors {
		require.Contains(t, tmpl.Tags, "olap")
		require.Contains(t, tmpl.Tags, "connector")
	}

	// Empty tags returns all templates
	allTemplates := r.ListByTags(nil)
	require.Equal(t, len(r.List()), len(allTemplates))
}

func TestRegistryLookupByDriver(t *testing.T) {
	r, err := NewRegistry()
	require.NoError(t, err)

	// Connector lookup: driver name = template name
	tmpl, ok := r.LookupByDriver("s3", "connector")
	require.True(t, ok)
	require.Equal(t, "s3", tmpl.Name)

	// Model lookup for object stores: driver-duckdb
	tmpl, ok = r.LookupByDriver("s3", "model")
	require.True(t, ok)
	require.Equal(t, "s3-duckdb", tmpl.Name)

	// Model lookup for warehouses: driver-duckdb
	tmpl, ok = r.LookupByDriver("snowflake", "model")
	require.True(t, ok)
	require.Equal(t, "snowflake-duckdb", tmpl.Name)
}

func TestRegistryTemplatesSorted(t *testing.T) {
	r, err := NewRegistry()
	require.NoError(t, err)

	all := r.List()
	for i := 1; i < len(all); i++ {
		require.Less(t, all[i-1].Name, all[i].Name, "templates should be sorted by name")
	}
}

func TestRegistryAllDefinitionsValid(t *testing.T) {
	r, err := NewRegistry()
	require.NoError(t, err)

	for _, tmpl := range r.List() {
		t.Run(tmpl.Name, func(t *testing.T) {
			require.NotEmpty(t, tmpl.Name)
			require.NotEmpty(t, tmpl.DisplayName)
			require.NotEmpty(t, tmpl.Tags)
			require.NotEmpty(t, tmpl.Files)

			for _, f := range tmpl.Files {
				require.NotEmpty(t, f.Name, "file name required for template %s", tmpl.Name)
				require.NotEmpty(t, f.PathTemplate, "path template required for template %s file %s", tmpl.Name, f.Name)
				require.NotEmpty(t, f.CodeTemplate, "code template required for template %s file %s", tmpl.Name, f.Name)
				require.Contains(t, []string{"connector", "model"}, f.Name, "file name must be connector or model for template %s", tmpl.Name)
			}
		})
	}
}
