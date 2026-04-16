package mysql_test

import (
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func TestInformationSchemaListTables(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	conn, olap := acquireTestMySQL(t)
	ctx := t.Context()

	rows, err := olap.Query(ctx, &drivers.Statement{Query: "SELECT DATABASE()"})
	require.NoError(t, err)
	defer rows.Close()
	require.True(t, rows.Next())
	var curSchema string
	err = rows.Scan(&curSchema)
	require.NoError(t, err)
	require.NotEmpty(t, curSchema)

	infoSchema, ok := conn.AsInformationSchema()
	require.True(t, ok)

	// MySQL has no database tier; pass empty string for database
	tables, _, err := infoSchema.ListTables(ctx, "", curSchema, 0, "")
	require.NoError(t, err)
	require.NotEmpty(t, tables)

	for _, tbl := range tables {
		require.True(t, tbl.IsDefaultDatabase, "table %s: expected IsDefaultDatabase=true", tbl.Name)
		require.True(t, tbl.IsDefaultDatabaseSchema, "table %s: expected IsDefaultDatabaseSchema=true", tbl.Name)
	}
}
