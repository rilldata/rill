package metricssql_test

import (
	"context"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"github.com/rilldata/rill/runtime/metricsview/metricssql"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestCompile(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")
	ctrl, err := rt.Controller(t.Context(), instanceID)
	require.NoError(t, err)
	olap, release, err := rt.OLAP(t.Context(), instanceID, "")
	require.NoError(t, err)
	defer release()

	resource := &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}
	mv, err := ctrl.Get(t.Context(), resource, false)
	require.NoError(t, err)

	resource = &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics_advanced"}
	advancedMV, err := ctrl.Get(t.Context(), resource, false)
	require.NoError(t, err)

	claims := &runtime.SecurityClaims{}
	compiler := metricssql.New(&metricssql.CompilerOptions{
		GetMetricsView: func(ctx context.Context, name string) (*runtimev1.Resource, error) {
			mv, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: name}, false)
			if err != nil {
				return nil, err
			}
			sec, err := rt.ResolveSecurity(ctx, ctrl.InstanceID, claims, mv)
			if err != nil {
				return nil, err
			}
			if !sec.CanAccess() {
				return nil, runtime.ErrForbidden
			}
			return mv, nil
		},
		GetTimestamps: func(ctx context.Context, mv *runtimev1.Resource, timeDim string) (metricsview.TimestampsResult, error) {
			sec, err := rt.ResolveSecurity(ctx, ctrl.InstanceID, claims, mv)
			if err != nil {
				return metricsview.TimestampsResult{}, err
			}
			e, err := executor.New(ctx, rt, instanceID, mv.GetMetricsView().State.ValidSpec, false, sec, 0)
			if err != nil {
				return metricsview.TimestampsResult{}, err
			}
			return e.Timestamps(ctx, timeDim)
		},
	})
	passTests := []struct {
		inSQL    string
		outSQL   string
		resource *runtimev1.Resource
		args     []any
	}{
		{
			"select pub, dom from ad_bids_metrics",
			"SELECT (\"publisher\") AS \"pub\", (\"domain\") AS \"dom\" FROM \"ad_bids\" GROUP BY 1, 2",
			mv,
			nil,
		},

		{
			"select pub, dom from ad_bids_metrics LIMIT 5",
			"SELECT (\"publisher\") AS \"pub\", (\"domain\") AS \"dom\" FROM \"ad_bids\" GROUP BY 1, 2 LIMIT 5",
			mv,
			nil,
		},
		{
			"select pub, dom from ad_bids_metrics order by pub desc",
			"SELECT (\"publisher\") AS \"pub\", (\"domain\") AS \"dom\" FROM \"ad_bids\" GROUP BY 1, 2 ORDER BY \"pub\" DESC NULLS LAST",
			mv,
			nil,
		},
		{
			"select pub, dom from ad_bids_metrics order by pub desc, dom asc",
			"SELECT (\"publisher\") AS \"pub\", (\"domain\") AS \"dom\" FROM \"ad_bids\" GROUP BY 1, 2 ORDER BY \"pub\" DESC NULLS LAST, \"dom\" NULLS LAST",
			mv,
			nil,
		},
		{
			"select pub, dom from ad_bids_metrics order by pub desc, dom asc limit 10",
			"SELECT (\"publisher\") AS \"pub\", (\"domain\") AS \"dom\" FROM \"ad_bids\" GROUP BY 1, 2 ORDER BY \"pub\" DESC NULLS LAST, \"dom\" NULLS LAST LIMIT 10",
			mv,
			nil,
		},
		{
			"select pub, dom from ad_bids_metrics where tld = 'Yahoo'",
			"SELECT (\"publisher\") AS \"pub\", (\"domain\") AS \"dom\" FROM \"ad_bids\" WHERE ((regexp_extract(domain, '(.*\\.)?(.*\\.com)', 2)) = ?) GROUP BY 1, 2",
			mv,
			[]any{"Yahoo"},
		},
		{
			"select pub, dom from ad_bids_metrics where dom like '%google%'",
			"SELECT (\"publisher\") AS \"pub\", (\"domain\") AS \"dom\" FROM \"ad_bids\" WHERE ((\"domain\") ILIKE ?) GROUP BY 1, 2",
			mv,
			[]any{"%google%"},
		},
		{
			"select pub, dom from ad_bids_metrics where tld = 'Yahoo' LIMIT 5",
			"SELECT (\"publisher\") AS \"pub\", (\"domain\") AS \"dom\" FROM \"ad_bids\" WHERE ((regexp_extract(domain, '(.*\\.)?(.*\\.com)', 2)) = ?) GROUP BY 1, 2 LIMIT 5",
			mv,
			[]any{"Yahoo"},
		},
		{
			"select pub, dom, measure_0, measure_1 from ad_bids_metrics",
			"SELECT (\"publisher\") AS \"pub\", (\"domain\") AS \"dom\", (count(*)) AS \"measure_0\", (avg(bid_price)) AS \"measure_1\" FROM \"ad_bids\" GROUP BY 1, 2",
			mv,
			nil,
		},
		{
			"select pub, dom,measure_0 from ad_bids_metrics  where tld = 'Yahoo' order by pub desc, dom asc limit 10",
			"SELECT (\"publisher\") AS \"pub\", (\"domain\") AS \"dom\", (count(*)) AS \"measure_0\" FROM \"ad_bids\" WHERE ((regexp_extract(domain, '(.*\\.)?(.*\\.com)', 2)) = ?) GROUP BY 1, 2 ORDER BY \"pub\" DESC NULLS LAST, \"dom\" NULLS LAST LIMIT 10",
			mv,
			[]any{"Yahoo"},
		},
		{
			"select pub, dom,measure_0 from ad_bids_metrics  where tld = 'Yahoo' having measure_0 > 10 order by pub desc, dom asc limit 10",
			"SELECT (\"t1\".\"pub\") AS \"pub\", (\"t1\".\"dom\") AS \"dom\", (\"t1\".\"measure_0\") AS \"measure_0\" FROM (SELECT (\"publisher\") AS \"pub\", (\"domain\") AS \"dom\", (count(*)) AS \"measure_0\" FROM \"ad_bids\" WHERE ((regexp_extract(domain, '(.*\\.)?(.*\\.com)', 2)) = ?) GROUP BY 1, 2) t1 WHERE ((\"t1\".\"measure_0\") > ?) ORDER BY \"pub\" DESC NULLS LAST, \"dom\" NULLS LAST LIMIT 10",
			mv,
			[]any{"Yahoo", 10},
		},
		{
			"select pub, dom from ad_bids_metrics where tld = '{{.user.domain}}'",
			"SELECT (\"publisher\") AS \"pub\", (\"domain\") AS \"dom\" FROM \"ad_bids\" WHERE ((regexp_extract(domain, '(.*\\.)?(.*\\.com)', 2)) = ?) GROUP BY 1, 2",
			mv,
			[]any{"{{.user.domain}}"},
		},
		{
			"select pub, dom, date_trunc('SECOND', timestamp) from ad_bids_metrics",
			"SELECT (\"publisher\") AS \"pub\", (\"domain\") AS \"dom\", (date_trunc('SECOND', \"timestamp\"::TIMESTAMP)::TIMESTAMP) AS \"DATE_TRUNC('SECOND', \"\"timestamp\"\")\" FROM \"ad_bids\" GROUP BY 1, 2, 3",
			mv,
			nil,
		},
		{
			"select pub, dom, measure_0 as \"click rate\" from ad_bids_metrics where (pub is not null and dom is null) or (pub = '__default__')",
			"SELECT (\"publisher\") AS \"pub\", (\"domain\") AS \"dom\", (count(*)) AS \"measure_0\" FROM \"ad_bids\" WHERE ((((\"publisher\") IS NOT NULL) AND ((\"domain\") IS NULL)) OR ((\"publisher\") = ?)) GROUP BY 1, 2",
			mv,
			[]any{"__default__"},
		},
		{
			"select pub, dom from ad_bids_metrics where pub in ('Yahoo', 'Google')",
			"SELECT (\"publisher\") AS \"pub\", (\"domain\") AS \"dom\" FROM \"ad_bids\" WHERE ((\"publisher\") IN (?,?)) GROUP BY 1, 2",
			mv,
			[]any{"Yahoo", "Google"},
		},
		{
			"select pub, dom from ad_bids_metrics where timestamp > '30-07-2024' - INTERVAL 90 DAY",
			"SELECT (\"publisher\") AS \"pub\", (\"domain\") AS \"dom\" FROM \"ad_bids\" WHERE ((\"timestamp\") > ?) GROUP BY 1, 2",
			mv,
			[]any{"2024-05-01T00:00:00Z"},
		},
		{
			"select pub, dom from ad_bids_metrics where timestamp > '30-07-2024' + INTERVAL 90 HOUR",
			"SELECT (\"publisher\") AS \"pub\", (\"domain\") AS \"dom\" FROM \"ad_bids\" WHERE ((\"timestamp\") > ?) GROUP BY 1, 2",
			mv,
			[]any{"2024-08-02T18:00:00Z"},
		},
		{
			"select timestamp, bids_1day_rolling_avg from ad_bids_metrics_advanced",
			"SELECT (\"t1\".\"timestamp\") AS \"timestamp\", (AVG(bids) OVER (ORDER BY \"t1\".\"timestamp\" RANGE BETWEEN INTERVAL 1 DAY PRECEDING AND CURRENT ROW)) AS \"bids_1day_rolling_avg\" FROM (SELECT (\"timestamp\") AS \"timestamp\", (count(*)) AS \"bids\" FROM \"ad_bids\" GROUP BY 1) t1",
			advancedMV,
			nil,
		},
		{
			"select pub, dom from ad_bids_metrics where timestamp > time_range_start('7D as of watermark/D+1D') and timestamp <= time_range_end('7D as of watermark/D+1D')",
			"SELECT (\"publisher\") AS \"pub\", (\"domain\") AS \"dom\" FROM \"ad_bids\" WHERE (((\"timestamp\") > ?) AND ((\"timestamp\") <= ?)) GROUP BY 1, 2",
			advancedMV,
			[]any{parseTestTime(t, "2022-03-24T00:00:00Z"), parseTestTime(t, "2022-03-31T00:00:00Z")},
		},
	}

	clm, err := rt.ResolveSecurity(t.Context(), instanceID, claims, mv)
	require.NoError(t, err)

	for _, test := range passTests {
		q, err := compiler.Parse(t.Context(), test.inSQL)
		require.NoError(t, err, "input = %v", test.inSQL)
		ast, err := metricsview.NewAST(test.resource.GetMetricsView().State.ValidSpec, clm, q, olap.Dialect())
		require.NoError(t, err)

		sql, args, err := ast.SQL()
		require.NoError(t, err)
		require.Equal(t, test.outSQL, sql)
		require.ElementsMatch(t, test.args, args)

		res, err := olap.Query(t.Context(), &drivers.Statement{Query: sql, Args: args})
		require.NoError(t, err)
		require.NoError(t, res.Close())
	}
}

func parseTestTime(tst *testing.T, t string) time.Time {
	ts, err := time.Parse(time.RFC3339, t)
	require.NoError(tst, err)
	return ts
}
