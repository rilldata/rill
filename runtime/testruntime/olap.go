package testruntime

import (
	"context"
	"fmt"
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func RequireOLAPTable(t testing.TB, rt *runtime.Runtime, id, name string) {
	olap, release, err := rt.OLAP(context.Background(), id)
	require.NoError(t, err)
	defer release()

	_, err = olap.InformationSchema().Lookup(context.Background(), name)
	require.NoError(t, err)
}

func RequireNoOLAPTable(t testing.TB, rt *runtime.Runtime, id, name string) {
	olap, release, err := rt.OLAP(context.Background(), id)
	require.NoError(t, err)
	defer release()

	_, err = olap.InformationSchema().Lookup(context.Background(), name)
	require.ErrorIs(t, err, drivers.ErrNotFound)
}

func RequireOLAPTableCount(t testing.TB, rt *runtime.Runtime, id, name string, count int) {
	olap, release, err := rt.OLAP(context.Background(), id)
	require.NoError(t, err)
	defer release()

	_, err = olap.InformationSchema().Lookup(context.Background(), name)
	require.NoError(t, err)

	rows, err := olap.Execute(context.Background(), &drivers.Statement{Query: fmt.Sprintf(`SELECT count(*) FROM "%s"`, name)})
	require.NoError(t, err)
	defer rows.Close()

	var n int
	for rows.Next() {
		err := rows.Scan(&n)
		require.NoError(t, err)
	}

	require.Equal(t, count, n)
}
