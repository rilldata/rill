package databricks_test

import (
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestListDatabaseSchemas(t *testing.T) {
	t.Skip("skipping due to inactive Databricks account")
	testmode.Expensive(t)

	conn, _ := acquireTestDatabricks(t)
	is, ok := conn.AsInformationSchema()
	require.True(t, ok)

	schemas, nextToken, err := is.ListDatabaseSchemas(t.Context(), 0, "")
	require.NoError(t, err)
	require.Empty(t, nextToken)

	found := false
	for _, s := range schemas {
		if s.DatabaseSchema == "integration_test" {
			found = true
		}
		require.NotEmpty(t, s.DatabaseSchema)
	}
	require.True(t, found, "expected integration_test schema to be present")
}

func TestListDatabaseSchemasWhitelist(t *testing.T) {
	t.Skip("skipping due to inactive Databricks account")
	testmode.Expensive(t)

	cfg := testruntime.AcquireConnector(t, "databricks")
	cfg["catalog_whitelist"] = "SAMPLES, "
	conn, err := drivers.Open(t.Context(), "databricks", "", "default", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })
	is, ok := conn.AsInformationSchema()
	require.True(t, ok)

	schemas, _, err := is.ListDatabaseSchemas(t.Context(), 0, "")
	require.NoError(t, err)
	require.NotEmpty(t, schemas)
	for _, s := range schemas {
		require.Equal(t, "samples", s.Database)
	}
}

func TestListTables(t *testing.T) {
	t.Skip("skipping due to inactive Databricks account")
	testmode.Expensive(t)

	conn, _ := acquireTestDatabricks(t)
	is, ok := conn.AsInformationSchema()
	require.True(t, ok)

	tables, nextToken, err := is.ListTables(t.Context(), "", "integration_test", 0, "")
	require.NoError(t, err)
	require.Empty(t, nextToken)

	found := false
	for _, tbl := range tables {
		if tbl.Name == "all_datatypes" {
			found = true
			require.False(t, tbl.View)
		}
	}
	require.True(t, found, "expected all_datatypes table to be present")
}

func TestLookup(t *testing.T) {
	t.Skip("skipping due to inactive Databricks account")
	testmode.Expensive(t)

	conn, _ := acquireTestDatabricks(t)
	is, ok := conn.AsInformationSchema()
	require.True(t, ok)

	meta, err := is.Lookup(t.Context(), "", "integration_test", "all_datatypes")
	require.NoError(t, err)
	require.NotNil(t, meta)
	require.False(t, meta.View)
	require.NotEmpty(t, meta.Schema)

	// Verify expected columns and types from the init SQL.
	// Databricks information_schema uses its own type aliases (e.g. SHORT, LONG, BYTE)
	// and strips precision/length from scalar types (e.g. DECIMAL instead of DECIMAL(18,6)).
	expected := map[string]string{
		"id":                "INT",
		"boolean_col":       "BOOLEAN",
		"tinyint_col":       "BYTE",
		"smallint_col":      "SHORT",
		"int32_col":         "INT",
		"int64_col":         "LONG",
		"float_col":         "FLOAT",
		"double_col":        "DOUBLE",
		"decimal_col":       "DECIMAL",
		"string_col":        "STRING",
		"varchar_col":       "STRING",
		"date_col":          "DATE",
		"timestamp_col":     "TIMESTAMP",
		"timestamp_ntz_col": "TIMESTAMP_NTZ",
		"binary_col":        "BINARY",
		"array_col":         "ARRAY",
		"map_col":           "MAP",
		"struct_col":        "STRUCT",
	}
	actual := make(map[string]string, len(meta.Schema.Fields))
	for _, f := range meta.Schema.Fields {
		actual[f.Name] = f.Type.RawType
	}
	for col, typ := range expected {
		require.Equal(t, typ, actual[col], "unexpected type for column %q", col)
	}
}
