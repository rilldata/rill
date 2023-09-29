package testruntime

import (
	"context"
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/stretchr/testify/require"
)

func RequireTableRowCount(t testing.TB, rt *runtime.Runtime, id, table string, limit int) {
	q := &queries.TableHead{
		TableName: table,
		Limit:     3,
	}
	require.NoError(t, rt.Query(context.Background(), id, q, 5))
	require.Len(t, q.Result, limit)
}
