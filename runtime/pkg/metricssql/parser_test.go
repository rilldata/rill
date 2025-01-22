package metricssqlparser_test

import (
	"context"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	metricssqlparser "github.com/rilldata/rill/runtime/pkg/metricssql"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestCompile(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")
	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	olap, release, err := rt.OLAP(context.Background(), instanceID, "")
	require.NoError(t, err)
	defer release()

	resource := &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}
	mv, err := ctrl.Get(context.Background(), resource, false)
	require.NoError(t, err)

	resource = &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics_advanced"}
	advancedMV, err := ctrl.Get(context.Background(), resource, false)
	require.NoError(t, err)

	claims := &runtime.SecurityClaims{}
	compiler := metricssqlparser.New(ctrl, instanceID, claims, 1)
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
			"SELECT (t1.\"pub\") AS \"pub\", (t1.\"dom\") AS \"dom\", (t1.\"measure_0\") AS \"measure_0\" FROM (SELECT (\"publisher\") AS \"pub\", (\"domain\") AS \"dom\", (count(*)) AS \"measure_0\" FROM \"ad_bids\" WHERE ((regexp_extract(domain, '(.*\\.)?(.*\\.com)', 2)) = ?) GROUP BY 1, 2) t1 WHERE ((t1.\"measure_0\") > ?) ORDER BY \"pub\" DESC NULLS LAST, \"dom\" NULLS LAST LIMIT 10",
			mv,
			[]any{"Yahoo", "10"},
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
			"SELECT (t1.\"timestamp\") AS \"timestamp\", (AVG(bids) OVER (ORDER BY t1.\"timestamp\" RANGE BETWEEN INTERVAL 1 DAY PRECEDING AND CURRENT ROW)) AS \"bids_1day_rolling_avg\" FROM (SELECT (\"timestamp\") AS \"timestamp\", (count(*)) AS \"bids\" FROM \"ad_bids\" GROUP BY 1) t1",
			advancedMV,
			nil,
		},
		{
			"select pub, dom from ad_bids_metrics where timestamp > time_range_start('-7d,latest') and timestamp <= time_range_end('-7d,latest')",
			"SELECT (\"publisher\") AS \"pub\", (\"domain\") AS \"dom\" FROM \"ad_bids\" WHERE (((\"timestamp\") > ?) AND ((\"timestamp\") <= ?)) GROUP BY 1, 2",
			advancedMV,
			[]any{parseTestTime(t, "2022-03-23T00:00:00Z"), parseTestTime(t, "2022-03-30T23:59:44.2Z")},
		},
	}

	clm, err := rt.ResolveSecurity(instanceID, claims, mv)
	require.NoError(t, err)

	for _, test := range passTests {
		q, err := compiler.Rewrite(context.Background(), test.inSQL)
		require.NoError(t, err, "input = %v", test.inSQL)
		ast, err := metricsview.NewAST(test.resource.GetMetricsView().State.ValidSpec, clm, q, olap.Dialect())
		require.NoError(t, err)

		sql, args, err := ast.SQL()
		require.NoError(t, err)
		require.Equal(t, test.outSQL, sql)
		require.ElementsMatch(t, test.args, args)

		res, err := olap.Execute(context.Background(), &drivers.Statement{Query: sql, Args: args})
		require.NoError(t, err)
		require.NoError(t, res.Close())
	}
}

func parseTestTime(tst *testing.T, t string) time.Time {
	ts, err := time.Parse(time.RFC3339, t)
	require.NoError(tst, err)
	return ts
}
