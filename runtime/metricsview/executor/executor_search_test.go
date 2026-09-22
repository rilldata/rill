package executor_test

import (
	"context"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers/druid"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"

	_ "github.com/rilldata/rill/runtime/resolvers"
)

func TestSearch(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"models/events.sql": `
SELECT * FROM (VALUES
	(TIMESTAMP '2024-01-01 00:00:00', 'Google', 'news.com'),
	(TIMESTAMP '2024-01-02 00:00:00', 'Facebook', 'sports.com'),
	(TIMESTAMP '2024-01-03 00:00:00', 'Microsoft', 'foo.com'),
	(TIMESTAMP '2024-02-01 00:00:00', 'Yahoo', 'news.com')
) t(timestamp, publisher, domain)
`,
			"metrics_views/mv.yaml": `
type: metrics_view
version: 1
model: events
timeseries: timestamp
dimensions:
  - name: publisher
    column: publisher
  - name: domain
    column: domain
measures:
  - name: count
    expression: count(*)
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	r := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindMetricsView, "mv")
	mv := r.GetMetricsView().State.ValidSpec
	require.NotNil(t, mv)

	e, err := executor.New(context.Background(), rt, instanceID, mv, false, runtime.ResolvedSecurityOpen, 0, nil)
	require.NoError(t, err)
	defer e.Close()

	t.Run("multiple dimensions", func(t *testing.T) {
		res, err := e.Search(context.Background(), &metricsview.SearchQuery{
			MetricsView: "mv",
			Dimensions:  []string{"publisher", "domain"},
			Search:      "oo",
		}, nil)
		require.NoError(t, err)
		require.ElementsMatch(t, []metricsview.SearchResult{
			{Dimension: "publisher", Value: "Google"},
			{Dimension: "publisher", Value: "Facebook"},
			{Dimension: "publisher", Value: "Yahoo"},
			{Dimension: "domain", Value: "foo.com"},
		}, res)
	})

	t.Run("with where and time range", func(t *testing.T) {
		res, err := e.Search(context.Background(), &metricsview.SearchQuery{
			MetricsView: "mv",
			Dimensions:  []string{"publisher", "domain"},
			Search:      "oo",
			Where: &metricsview.Expression{Condition: &metricsview.Condition{
				Operator:    metricsview.OperatorNeq,
				Expressions: []*metricsview.Expression{{Name: "domain"}, {Value: "sports.com"}},
			}},
			TimeRange: &metricsview.TimeRange{
				Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			},
		}, nil)
		require.NoError(t, err)
		require.ElementsMatch(t, []metricsview.SearchResult{
			{Dimension: "publisher", Value: "Google"},
			{Dimension: "domain", Value: "foo.com"},
		}, res)
	})
}

// TestSearchDruidSQLCastsNonStringDimensions checks the per-dimension SQL that the search fallback (UNION ALL) query generates for Druid.
// Druid has no ILIKE, so the search is compiled to REGEXP_LIKE, which only accepts string operands:
// dimensions with a known non-string type (including the time dimension) must be cast to VARCHAR, while string dimensions and dimensions of unknown type are matched directly.
func TestSearchDruidSQLCastsNonStringDimensions(t *testing.T) {
	mv := &runtimev1.MetricsViewSpec{
		Table:         "events",
		TimeDimension: "__time",
		Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
			{Name: "__time", Column: "__time", DataType: &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP}},
			{Name: "account_id", Column: "account_id", DataType: &runtimev1.Type{Code: runtimev1.Type_CODE_INT64}},
			{Name: "account_name", Column: "account_name", DataType: &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}},
			{Name: "domain", Column: "domain"},
		},
	}

	cases := map[string]string{
		"__time":       `SELECT ("__time") AS "__time" FROM "events" WHERE (REGEXP_LIKE(CAST(("__time") AS VARCHAR), ?)) GROUP BY 1`,
		"account_id":   `SELECT ("account_id") AS "account_id" FROM "events" WHERE (REGEXP_LIKE(CAST(("account_id") AS VARCHAR), ?)) GROUP BY 1`,
		"account_name": `SELECT ("account_name") AS "account_name" FROM "events" WHERE (REGEXP_LIKE(("account_name"), ?)) GROUP BY 1`,
		"domain":       `SELECT ("domain") AS "domain" FROM "events" WHERE (REGEXP_LIKE(("domain"), ?)) GROUP BY 1`,
	}
	for dim, want := range cases {
		t.Run(dim, func(t *testing.T) {
			qry := &metricsview.Query{
				MetricsView: "mv",
				Dimensions:  []metricsview.Dimension{{Name: dim}},
				Where: &metricsview.Expression{Condition: &metricsview.Condition{
					Operator:    metricsview.OperatorIlike,
					Expressions: []*metricsview.Expression{{Name: dim}, {Value: "%tvc%"}},
				}},
			}
			ast, err := metricsview.NewAST(mv, runtime.ResolvedSecurityOpen, qry, druid.DialectDruid)
			require.NoError(t, err)

			sql, args, err := ast.SQL()
			require.NoError(t, err)
			require.Equal(t, want, sql)
			require.Equal(t, []any{"^(?i).*tvc.*$"}, args)
		})
	}
}
