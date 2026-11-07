package queries_test

import (
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/stretchr/testify/require"
)

// TestMetricsViewRowsKey checks that the cache key depends on the security claims,
// so callers with different row filters never share cached results.
func TestMetricsViewRowsKey(t *testing.T) {
	key := func(attrs map[string]any) string {
		limit := int64(10)
		q := &queries.MetricsViewRows{
			MetricsViewName: "mv",
			Limit:           &limit,
			SecurityClaims: &runtime.SecurityClaims{
				UserAttributes: attrs,
				Permissions:    []runtime.Permission{runtime.ReadMetrics},
			},
		}
		return q.Key()
	}

	a := key(map[string]any{"org_id": "a"})
	require.Equal(t, a, key(map[string]any{"org_id": "a"}))
	require.NotEqual(t, a, key(map[string]any{"org_id": "b"}))
	require.NotEqual(t, a, key(nil))
}
