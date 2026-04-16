package snowflake_test

import (
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
)

func TestInformationSchemaListTables(t *testing.T) {
	testmode.Expensive(t)

	conn, olap := acquireTestSnowflake(t)
	ctx := t.Context()

	rows, err := olap.Query(ctx, &drivers.Statement{Query: "SELECT CURRENT_DATABASE(), CURRENT_SCHEMA()"})
	require.NoError(t, err)
	defer rows.Close()
	require.True(t, rows.Next())
	var curDB, curSchema string
	err = rows.Scan(&curDB, &curSchema)
	require.NoError(t, err)
	require.NotEmpty(t, curDB)
	require.NotEmpty(t, curSchema)

	infoSchema, ok := conn.AsInformationSchema()
	require.True(t, ok)

	tables, _, err := infoSchema.ListTables(ctx, curDB, curSchema, 0, "")
	require.NoError(t, err)
	require.NotEmpty(t, tables)

	for _, tbl := range tables {
		require.True(t, tbl.IsDefaultDatabase, "table %s: expected IsDefaultDatabase=true", tbl.Name)
		require.True(t, tbl.IsDefaultDatabaseSchema, "table %s: expected IsDefaultDatabaseSchema=true", tbl.Name)
	}
}
