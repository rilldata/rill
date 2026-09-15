package executor_test

import (
	"context"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime"
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
