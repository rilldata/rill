package executor_test

import (
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

// The exact expressions below are deliberately different from the default ones (rather than a real approximate/exact pair),
// so that the tests can tell which expression was used.
var exactMeasuresTestFiles = map[string]string{
	"rill.yaml": "",
	"m1.sql":    `SELECT * FROM (VALUES (1, 'a'), (2, 'b'), (3, 'a')) t(id, publisher)`,
	"mv1.yaml": `
type: metrics_view
model: m1
dimensions:
  - name: publisher
    column: publisher
measures:
  - name: records
    expression: count(*)
    exact_expression: count(*) + 100
  - name: ids
    expression: sum(id)
  - name: records_per_id
    requires: [records, ids]
    expression: records / ids
  - name: publishers
    expression: count(distinct publisher)
    exact_expression: count(distinct publisher) + 1000
  - name: records_per_publisher
    requires: [records, publishers]
    expression: records / publishers
`,
	// m2 adds a time dimension: the rows of m1 on 2024-01-02, plus rows on 2024-01-01 that serve as a comparison period.
	// Per publisher: a has 2 rows on the 2nd and 1 on the 1st; b has 1 row on the 2nd and 2 on the 1st.
	"m2.sql": `SELECT * FROM (VALUES
		(1, 'a', TIMESTAMP '2024-01-02'), (2, 'b', TIMESTAMP '2024-01-02'), (3, 'a', TIMESTAMP '2024-01-02'),
		(4, 'a', TIMESTAMP '2024-01-01'), (5, 'b', TIMESTAMP '2024-01-01'), (6, 'b', TIMESTAMP '2024-01-01')
	) t(id, publisher, ts)`,
	"mv2.yaml": `
type: metrics_view
model: m2
timeseries: ts
dimensions:
  - name: publisher
    column: publisher
measures:
  - name: records
    expression: count(*)
    exact_expression: count(*) + 100
`,
}

// recordsFilterSubquery is a `publisher IN (SELECT publisher ... HAVING records > threshold)` filter
// whose subquery requests records exactly when exact is set.
func recordsFilterSubquery(threshold int, exact bool) map[string]any {
	return map[string]any{"cond": map[string]any{"op": "in", "exprs": []map[string]any{
		{"name": "publisher"},
		{"subquery": map[string]any{
			"dimension": map[string]any{"name": "publisher"},
			"measures":  []map[string]any{{"name": "records", "exact": exact}},
			"having":    map[string]any{"cond": map[string]any{"op": "gt", "exprs": []map[string]any{{"name": "records"}, {"val": threshold}}}},
		}},
	}}}
}

func resolveMetrics(t *testing.T, rt *runtime.Runtime, instanceID string, props map[string]any) ([]map[string]any, error) {
	if _, ok := props["metrics_view"]; !ok {
		props["metrics_view"] = "mv1"
	}
	res, _, err := rt.Resolve(t.Context(), &runtime.ResolveOptions{
		InstanceID:         instanceID,
		Resolver:           "metrics",
		ResolverProperties: props,
		Claims:             &runtime.SecurityClaims{Permissions: runtime.AllPermissions},
	})
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var rows []map[string]any
	for {
		row, err := res.Next()
		if err != nil {
			break
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func TestExactMeasures(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{Files: exactMeasuresTestFiles})
	testruntime.RequireReconcileState(t, rt, instanceID, -1, 0, 0)

	t.Run("approximate by default", func(t *testing.T) {
		rows, err := resolveMetrics(t, rt, instanceID, map[string]any{
			"measures": []map[string]any{{"name": "records"}, {"name": "records_per_id"}},
		})
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.EqualValues(t, 3, rows[0]["records"])
		require.InDelta(t, 0.5, rows[0]["records_per_id"], 1e-9)
	})

	t.Run("flagged measure is exact", func(t *testing.T) {
		rows, err := resolveMetrics(t, rt, instanceID, map[string]any{
			"measures": []map[string]any{{"name": "records", "exact": true}, {"name": "ids"}},
		})
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.EqualValues(t, 103, rows[0]["records"])
		require.EqualValues(t, 6, rows[0]["ids"])
	})

	t.Run("flagged derived measure computes its referenced measure exactly", func(t *testing.T) {
		rows, err := resolveMetrics(t, rt, instanceID, map[string]any{
			"measures": []map[string]any{{"name": "records_per_id", "exact": true}},
		})
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.InDelta(t, 103.0/6.0, rows[0]["records_per_id"], 1e-9)
	})

	t.Run("a measure requested both ways is exact", func(t *testing.T) {
		rows, err := resolveMetrics(t, rt, instanceID, map[string]any{
			"measures": []map[string]any{{"name": "records"}, {"name": "records_per_id", "exact": true}},
		})
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.EqualValues(t, 103, rows[0]["records"])
		require.InDelta(t, 103.0/6.0, rows[0]["records_per_id"], 1e-9)
	})

	t.Run("flag on a measure without an exact expression is a no-op", func(t *testing.T) {
		rows, err := resolveMetrics(t, rt, instanceID, map[string]any{
			"measures": []map[string]any{{"name": "ids", "exact": true}},
		})
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.EqualValues(t, 6, rows[0]["ids"])
	})

	t.Run("percent of total uses the same precision for the totals", func(t *testing.T) {
		rows, err := resolveMetrics(t, rt, instanceID, map[string]any{
			"dimensions": []map[string]any{{"name": "publisher"}},
			"measures": []map[string]any{
				{"name": "records", "exact": true},
				{"name": "records_pct", "compute": map[string]any{"percent_of_total": map[string]any{"measure": "records"}}, "exact": true},
			},
			"sort": []map[string]any{{"name": "publisher"}},
		})
		require.NoError(t, err)
		require.Len(t, rows, 2)
		// publisher a has 2 rows, b has 1 row. Exact: a=102, b=101, total=103 (the totals query also adds 100 once).
		require.EqualValues(t, 102, rows[0]["records"])
		require.InDelta(t, 102.0/103.0, rows[0]["records_pct"], 1e-9)
		require.EqualValues(t, 101, rows[1]["records"])
		require.InDelta(t, 101.0/103.0, rows[1]["records_pct"], 1e-9)
	})

	t.Run("percent of total follows the query's resolved precision when only the base measure is flagged", func(t *testing.T) {
		rows, err := resolveMetrics(t, rt, instanceID, map[string]any{
			"dimensions": []map[string]any{{"name": "publisher"}},
			"measures": []map[string]any{
				{"name": "records", "exact": true},
				{"name": "records_pct", "compute": map[string]any{"percent_of_total": map[string]any{"measure": "records"}}},
			},
			"sort": []map[string]any{{"name": "publisher"}},
		})
		require.NoError(t, err)
		require.Len(t, rows, 2)
		// records is exact everywhere in the query, so the totals query must be exact too: 102/103, not 102/3.
		require.EqualValues(t, 102, rows[0]["records"])
		require.InDelta(t, 102.0/103.0, rows[0]["records_pct"], 1e-9)
		require.EqualValues(t, 101, rows[1]["records"])
		require.InDelta(t, 101.0/103.0, rows[1]["records_pct"], 1e-9)
	})

	t.Run("two-phase comparison keeps the base query exact when only the comparison is flagged", func(t *testing.T) {
		rows, err := resolveMetrics(t, rt, instanceID, map[string]any{
			"metrics_view": "mv2",
			"dimensions":   []map[string]any{{"name": "publisher"}},
			"measures": []map[string]any{
				{"name": "records"},
				{"name": "records_prev", "compute": map[string]any{"comparison_value": map[string]any{"measure": "records"}}, "exact": true},
			},
			"time_range":            map[string]any{"start": "2024-01-02T00:00:00Z", "end": "2024-01-03T00:00:00Z"},
			"comparison_time_range": map[string]any{"start": "2024-01-01T00:00:00Z", "end": "2024-01-02T00:00:00Z"},
			"sort":                  []map[string]any{{"name": "records", "desc": true}},
			// Under the two-phase limit, so the base values are queried first (without the comparison measure that carries the flag).
			"limit": 10,
		})
		require.NoError(t, err)
		require.Len(t, rows, 2)
		require.Equal(t, "a", rows[0]["publisher"])
		require.EqualValues(t, 102, rows[0]["records"])
		require.EqualValues(t, 101, rows[0]["records_prev"])
		require.Equal(t, "b", rows[1]["publisher"])
		require.EqualValues(t, 101, rows[1]["records"])
		require.EqualValues(t, 102, rows[1]["records_prev"])
	})

	t.Run("a measure flagged in a filter subquery is exact in the whole query", func(t *testing.T) {
		rows, err := resolveMetrics(t, rt, instanceID, map[string]any{
			"dimensions": []map[string]any{{"name": "publisher"}},
			"measures":   []map[string]any{{"name": "records"}},
			"where":      recordsFilterSubquery(101, true),
			"sort":       []map[string]any{{"name": "publisher"}},
		})
		require.NoError(t, err)
		// Exact: a=102, b=101, so only a passes the filter, and the outer query computes records exactly as well.
		require.Len(t, rows, 1)
		require.Equal(t, "a", rows[0]["publisher"])
		require.EqualValues(t, 102, rows[0]["records"])
	})

	t.Run("a derived measure calculated from an exact measure computes its other inputs exactly", func(t *testing.T) {
		rows, err := resolveMetrics(t, rt, instanceID, map[string]any{
			"measures": []map[string]any{{"name": "records", "exact": true}, {"name": "publishers"}, {"name": "records_per_publisher"}},
		})
		require.NoError(t, err)
		require.Len(t, rows, 1)
		// records_per_publisher is calculated from records, so it never divides an exact count by an approximate one: publishers is exact too (2 + 1000).
		require.EqualValues(t, 103, rows[0]["records"])
		require.EqualValues(t, 1002, rows[0]["publishers"])
		require.InDelta(t, 103.0/1002.0, rows[0]["records_per_publisher"], 1e-9)
	})

	t.Run("without the derived measure, the other input stays approximate", func(t *testing.T) {
		rows, err := resolveMetrics(t, rt, instanceID, map[string]any{
			"measures": []map[string]any{{"name": "records", "exact": true}, {"name": "publishers"}},
		})
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.EqualValues(t, 103, rows[0]["records"])
		require.EqualValues(t, 2, rows[0]["publishers"])
	})

	t.Run("two-phase comparison applies the query's precision to a filter subquery", func(t *testing.T) {
		rows, err := resolveMetrics(t, rt, instanceID, map[string]any{
			"metrics_view": "mv2",
			"dimensions":   []map[string]any{{"name": "publisher"}},
			"measures": []map[string]any{
				{"name": "records"},
				{"name": "records_prev", "compute": map[string]any{"comparison_value": map[string]any{"measure": "records"}}, "exact": true},
			},
			// The subquery does not flag records itself: it is exact because the query computes records exactly.
			// Exact, both publishers pass the filter (102 and 101 are above 1); approximate, only a would (2 is, 1 is not).
			"where":                 recordsFilterSubquery(1, false),
			"time_range":            map[string]any{"start": "2024-01-02T00:00:00Z", "end": "2024-01-03T00:00:00Z"},
			"comparison_time_range": map[string]any{"start": "2024-01-01T00:00:00Z", "end": "2024-01-02T00:00:00Z"},
			"sort":                  []map[string]any{{"name": "records", "desc": true}},
			"limit":                 10,
		})
		require.NoError(t, err)
		require.Len(t, rows, 2)
		require.Equal(t, "a", rows[0]["publisher"])
		require.EqualValues(t, 102, rows[0]["records"])
		require.EqualValues(t, 101, rows[0]["records_prev"])
		require.Equal(t, "b", rows[1]["publisher"])
		require.EqualValues(t, 101, rows[1]["records"])
		require.EqualValues(t, 102, rows[1]["records_prev"])
	})
}

func TestExactMeasuresDisabled(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files:     exactMeasuresTestFiles,
		Variables: map[string]string{"rill.metrics.exact_measures": "false"},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, -1, 0, 0)

	// Approximate queries keep working.
	rows, err := resolveMetrics(t, rt, instanceID, map[string]any{
		"measures": []map[string]any{{"name": "records"}},
	})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.EqualValues(t, 3, rows[0]["records"])

	// A flag on a measure without an exact expression computes nothing exactly, so it is still allowed.
	rows, err = resolveMetrics(t, rt, instanceID, map[string]any{
		"measures": []map[string]any{{"name": "ids", "exact": true}},
	})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.EqualValues(t, 6, rows[0]["ids"])

	// Exact queries fail with a clear error instead of silently returning approximate values.
	_, err = resolveMetrics(t, rt, instanceID, map[string]any{
		"measures": []map[string]any{{"name": "records", "exact": true}},
	})
	require.ErrorContains(t, err, "exact measures are disabled")

	// Including when the exact measure is requested inside a filter subquery.
	_, err = resolveMetrics(t, rt, instanceID, map[string]any{
		"dimensions": []map[string]any{{"name": "publisher"}},
		"measures":   []map[string]any{{"name": "records"}},
		"where":      recordsFilterSubquery(0, true),
	})
	require.ErrorContains(t, err, "exact measures are disabled")
}
