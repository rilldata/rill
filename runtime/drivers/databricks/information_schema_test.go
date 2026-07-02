package databricks_test

import (
	"testing"

	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
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
	expected := []struct {
		Name string
		Type string
	}{
		{Name: "id", Type: "INT"},
		{Name: "boolean_col", Type: "BOOLEAN"},
		{Name: "tinyint_col", Type: "BYTE"},
		{Name: "smallint_col", Type: "SHORT"},
		{Name: "int32_col", Type: "INT"},
		{Name: "int64_col", Type: "LONG"},
		{Name: "float_col", Type: "FLOAT"},
		{Name: "double_col", Type: "DOUBLE"},
		{Name: "decimal_col", Type: "DECIMAL"},
		{Name: "string_col", Type: "STRING"},
		{Name: "tinyint_col", Type: "BYTE"},
		{Name: "varchar_col", Type: "STRING"},
		{Name: "date_col", Type: "DATE"},
		{Name: "timestamp_col", Type: "TIMESTAMP"},
		{Name: "timestamp_ntz_col", Type: "TIMESTAMP_NTZ"},
		{Name: "binary_col", Type: "BINARY"},
		{Name: "array_col", Type: "ARRAY"},
		{Name: "map_col", Type: "MAP"},
		{Name: "struct_col", Type: "STRUCT"},
	}
	for col, typ := range expected {
		require.Equal(t, typ, meta.Schema.Fields[col].Type.Code, "unexpected type for column %q", col)
	}
}
