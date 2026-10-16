package executor_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

// The approximate uniq functions keep sketches in their states, which ClickHouse merges the same way across threads and parts,
// so the two-level query must return exactly what the single-pass query returns, also above the thresholds where they stop being exact.
// Two keys share a name and have overlapping user sets, so a rewrite that counted per key would overshoot.
func TestRewriteLookupGroupByKeyUniqFunctions(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		TestConnectors: []string{"clickhouse"},
		Files: map[string]string{
			"rill.yaml": "olap_connector: clickhouse",
			// Keys a and b share the name "shared" and 50,000 of their 100,000 users each; key c has seven users.
			"facts.sql": `
-- @connector: clickhouse
select 'a' as key, toUInt32(number) as user_id, toUInt8(number % 5) as clicks from numbers(100000)
union all select 'b' as key, toUInt32(number + 50000) as user_id, toUInt8(number % 5) as clicks from numbers(100000)
union all select 'c' as key, toUInt32(number % 7) as user_id, toUInt8(1) as clicks from numbers(100)
`,
			"names.yaml": `
type: model
connector: clickhouse
sql: select 'a' as key, 'shared' as name union all select 'b' as key, 'shared' as name union all select 'c' as key, 'small' as name
output:
  connector: clickhouse
  type: DICTIONARY
  primary_key: key
  dictionary_source_user: default
  dictionary_source_password: default
`,
			"mv.yaml": `
type: metrics_view
version: 1
model: facts
dimensions:
  - name: name
    column: key
    lookup_table: names
    lookup_key_column: key
    lookup_value_column: name
    lookup_group_by_key: true
measures:
  - name: uniq_users
    expression: uniq(user_id)
  - name: uniq_combined
    expression: uniqCombined(user_id)
  - name: uniq_combined64
    expression: uniqCombined64(user_id)
  - name: uniq_hll
    expression: uniqHLL12(user_id)
  - name: uniq_theta
    expression: uniqTheta(user_id)
  - name: uniq_upto
    expression: uniqUpTo(100)(user_id)
  - name: uniq_if
    expression: uniqIf(user_id, clicks > 0)
  - name: uniq_exact
    expression: uniqExact(user_id)
  - name: distinct_users
    expression: count(DISTINCT user_id)
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	r := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindMetricsView, "mv")
	mv := r.GetMetricsView().State.ValidSpec
	require.NotNil(t, mv)

	ctx := context.Background()
	olap, release, err := rt.OLAP(ctx, instanceID, mv.Connector)
	require.NoError(t, err)
	defer release()

	e, err := executor.New(ctx, rt, instanceID, mv, false, runtime.ResolvedSecurityOpen, 0, nil)
	require.NoError(t, err)
	defer e.Close()

	measures := make([]metricsview.Measure, 0, len(mv.Measures))
	measureExprs := make([]string, 0, len(mv.Measures))
	for _, m := range mv.Measures {
		measures = append(measures, metricsview.Measure{Name: m.Name})
		measureExprs = append(measureExprs, fmt.Sprintf("%s AS %s", m.Expression, olap.Dialect().EscapeAlias(m.Name)))
	}
	qry := &metricsview.Query{
		Dimensions: []metricsview.Dimension{{Name: "name"}},
		Measures:   measures,
		Sort:       []metricsview.Sort{{Name: "name"}},
	}

	// The single-pass query that the rewrite replaces.
	baseline := lookupTestRows(t, olap, fmt.Sprintf(
		`SELECT dictGet('names', 'name', "key") AS "name", %s FROM %s GROUP BY 1 ORDER BY 1`,
		strings.Join(measureExprs, ", "), olap.Dialect().EscapeTable(mv.Database, mv.DatabaseSchema, mv.Table),
	), nil)
	require.Len(t, baseline, 2)
	require.Equal(t, "shared", baseline[0]["name"])
	require.EqualValues(t, 150000, baseline[0]["uniq_exact"], "the shared name must be above the exact thresholds of the approximate functions")

	// The rewritten query, checked both as generated SQL and through the executor.
	ast, err := metricsview.NewAST(mv, runtime.ResolvedSecurityOpen, qry, olap.Dialect())
	require.NoError(t, err)
	e.RewriteLookupGroupByKey(ctx, ast)
	sql, args, err := ast.SQL()
	require.NoError(t, err)
	for _, fn := range []string{"uniqState(", "uniqCombinedState(", "uniqCombined64State(", "uniqHLL12State(", "uniqThetaState(", "uniqUpToState(100)(", "uniqIfState(", "uniqExactState("} {
		require.Contains(t, sql, fn)
	}
	require.Equal(t, baseline, lookupTestRows(t, olap, sql, args))

	res, err := e.Query(ctx, qry, nil)
	require.NoError(t, err)
	require.Equal(t, baseline, lookupTestResultRows(t, res))
}

func lookupTestRows(t *testing.T, olap drivers.OLAPStore, sql string, args []any) []map[string]any {
	t.Helper()
	res, err := olap.Query(context.Background(), &drivers.Statement{Query: sql, Args: args})
	require.NoError(t, err)
	return lookupTestResultRows(t, res)
}

func lookupTestResultRows(t *testing.T, res *drivers.Result) []map[string]any {
	t.Helper()
	defer res.Close()
	var rows []map[string]any
	for res.Next() {
		row := make(map[string]any)
		require.NoError(t, res.MapScan(row))
		rows = append(rows, row)
	}
	require.NoError(t, res.Err())
	return rows
}
