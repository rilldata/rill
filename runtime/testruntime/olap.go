package testruntime

import (
	"fmt"
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func RequireOLAPTable(t testing.TB, rt *runtime.Runtime, id, name string) {
	ctx := t.Context()
	olap, release, err := rt.OLAP(ctx, id, "")
	require.NoError(t, err)
	defer release()

	_, err = olap.InformationSchema().Lookup(ctx, "", "", name)
	require.NoError(t, err)
}

func RequireNoOLAPTable(t testing.TB, rt *runtime.Runtime, id, name string) {
	ctx := t.Context()
	olap, release, err := rt.OLAP(ctx, id, "")
	require.NoError(t, err)
	defer release()

	_, err = olap.InformationSchema().Lookup(ctx, "", "", name)
	require.ErrorIs(t, err, drivers.ErrNotFound)
}

func RequireOLAPTableCount(t testing.TB, rt *runtime.Runtime, id, name string, count int) {
	ctx := t.Context()
	olap, release, err := rt.OLAP(ctx, id, "")
	require.NoError(t, err)
	defer release()

	_, err = olap.InformationSchema().Lookup(ctx, "", "", name)
	require.NoError(t, err)

	rows, err := olap.Query(ctx, &drivers.Statement{Query: fmt.Sprintf(`SELECT count(*) FROM %s`, drivers.DialectDuckDB.EscapeIdentifier(name))})
	require.NoError(t, err)
	defer rows.Close()

	var n int
	for rows.Next() {
		err := rows.Scan(&n)
		require.NoError(t, err)
	}
	require.NoError(t, rows.Err())
	require.Equal(t, count, n)
}

func RequireIsView(t testing.TB, olap drivers.OLAPStore, tableName string, isView bool) {
	table, err := olap.InformationSchema().Lookup(t.Context(), "", "", tableName)
	require.NoError(t, err)
	// Assert that the model is a table now
	require.Equal(t, table.View, isView)
}
