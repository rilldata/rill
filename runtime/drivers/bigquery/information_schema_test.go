package bigquery_test

import (
	"testing"

	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
)

func TestInformationSchemaListTables(t *testing.T) {
	testmode.Expensive(t)

	conn, _ := acquireTestBigQuery(t)
	ctx := t.Context()

	infoSchema, ok := conn.AsInformationSchema()
	require.True(t, ok)

	// Get first available schema to find a valid project/dataset pair
	schemas, _, err := infoSchema.ListDatabaseSchemas(ctx, 1, "")
	require.NoError(t, err)
	require.NotEmpty(t, schemas)

	db := schemas[0].Database
	dbSchema := schemas[0].DatabaseSchema

	tables, _, err := infoSchema.ListTables(ctx, db, dbSchema, 0, "")
	require.NoError(t, err)
	require.NotEmpty(t, tables)

	for _, tbl := range tables {
		// IsDefaultDatabase is true when the project matches the configured project ID
		require.True(t, tbl.IsDefaultDatabase, "table %s: expected IsDefaultDatabase=true", tbl.Name)
		// BigQuery has no default dataset concept; IsDefaultDatabaseSchema is always false
		require.False(t, tbl.IsDefaultDatabaseSchema, "table %s: expected IsDefaultDatabaseSchema=false", tbl.Name)
	}
}
