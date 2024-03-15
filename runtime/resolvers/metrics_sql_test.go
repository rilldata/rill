package resolvers

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestCompiler(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")
	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)

	type result struct {
		sql  string
		deps []*runtimev1.ResourceName
	}
	tests := []struct {
		name string
		sql  string
		want result
	}{
		{
			"simple",
			"select pub,dom,AGGREGATE(measure_0) from ad_bids_metrics GROUP BY ALL",
			result{
				sql:  "SELECT \"publisher\" AS pub, \"domain\" AS dom, count(*) AS measure_0 FROM \"ad_bids\" GROUP BY ALL",
				deps: []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}},
			},
		},
		{
			"simple quoted",
			"select pub,  dom,   AGGREGATE( measure_0) from \"ad_bids_metrics\" GROUP BY pub, dom",
			result{
				sql:  "SELECT \"publisher\" AS pub, \"domain\" AS dom, count(*) AS measure_0 FROM \"ad_bids\" GROUP BY pub, dom",
				deps: []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}},
			},
		},
		{
			"aggregate and spaces with policy",
			`SELECT publisher,domain,AGGREGATE("bid's number"),AGGREGATE("total volume"),AGGREGATE("total click""s") From ad_bids_mini_metrics_with_policy`,
			result{
				sql:  `SELECT upper(publisher) AS publisher, "domain" AS domain, count(*) AS "bid's number", sum(volume) AS "total volume", sum(clicks) AS "total click""s" FROM (SELECT * FROM "ad_bids_mini") `,
				deps: []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_mini_metrics_with_policy"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &metricsSQLCompiler{
				instanceID: instanceID,
				ctrl:       ctrl,
				sql:        tt.sql,
			}

			got, _, deps, err := c.compile(context.Background())
			require.NoError(t, err)

			require.Subset(t, deps, tt.want.deps)
			require.Subset(t, tt.want.deps, deps)

			got = regexp.MustCompile(`\s+`).ReplaceAllString(strings.ReplaceAll(strings.ReplaceAll(got, "\n", " "), "\t", " "), " ")
			tt.want.sql = regexp.MustCompile(`\s+`).ReplaceAllString(strings.ReplaceAll(strings.ReplaceAll(tt.want.sql, "\n", " "), "\t", " "), " ")
			if got != tt.want.sql {
				t.Errorf("parsedSQL() = %q, want %q", got, tt.want.sql)
			}
		})
	}
}

func TestSimpleMetricsSQLApi(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	api, err := rt.APIForName(context.Background(), instanceID, "simple_mv_sql_api")
	require.NoError(t, err)

	res, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID:         instanceID,
		Resolver:           api.Spec.Resolver,
		ResolverProperties: api.Spec.ResolverProperties.AsMap(),
		Args:               nil,
		UserAttributes:     nil,
	})

	require.NoError(t, err)
	require.NotNil(t, res)
	var rows []map[string]interface{}
	require.NoError(t, json.Unmarshal(res, &rows))
	require.Equal(t, 5, len(rows))
	require.Equal(t, 2, len(rows[0]))
	require.Equal(t, "msn.com", rows[0]["dom"])
	require.Equal(t, nil, rows[0]["pub"])
}

func TestTemplateMetricsSQLAPI(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	testruntime.RequireParseErrors(t, rt, instanceID, nil)

	api, err := rt.APIForName(context.Background(), instanceID, "templated_mv_sql_api")
	require.NoError(t, err)

	res, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID:         instanceID,
		Resolver:           api.Spec.Resolver,
		ResolverProperties: api.Spec.ResolverProperties.AsMap(),
		Args:               map[string]any{"domain": "yahoo.com"},
		UserAttributes:     nil,
	})

	require.NoError(t, err)
	require.NotNil(t, res)
	var rows []map[string]interface{}
	require.NoError(t, json.Unmarshal(res, &rows))
	require.Equal(t, 1, len(rows))
	require.Equal(t, 3.0, rows[0]["measure_2"])
	require.Equal(t, "yahoo.com", rows[0]["domain"])
	require.Equal(t, "Yahoo", rows[0]["publisher"])
}

func TestPolicyMetricsSQLAPI(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	api, err := rt.APIForName(context.Background(), instanceID, "mv_sql_policy_api")
	require.NoError(t, err)

	res, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID:         instanceID,
		Resolver:           api.Spec.Resolver,
		ResolverProperties: api.Spec.ResolverProperties.AsMap(),
		Args:               nil,
		UserAttributes:     map[string]any{"domain": "yahoo.com", "email": "user@yahoo.com"},
	})

	require.NoError(t, err)
	require.NotNil(t, res)
	var rows []map[string]interface{}
	require.NoError(t, json.Unmarshal(res, &rows))
	require.Equal(t, 1, len(rows))
	require.Equal(t, nil, rows[0]["total volume"])
	require.Equal(t, 3.0, rows[0]["total impressions"])
	require.Equal(t, "yahoo.com", rows[0]["domain"])
	require.Equal(t, "YAHOO", rows[0]["publisher"])

	api, err = rt.APIForName(context.Background(), instanceID, "mv_sql_policy_api")
	require.NoError(t, err)

	res, err = rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID:         instanceID,
		Resolver:           api.Spec.Resolver,
		ResolverProperties: api.Spec.ResolverProperties.AsMap(),
		Args:               nil,
		UserAttributes:     map[string]any{"domain": "msn.com", "email": "user@msn.com"},
	})

	require.NoError(t, err)
	require.NotNil(t, res)
	var resp []map[string]interface{}
	require.NoError(t, json.Unmarshal(res, &resp))
	require.Equal(t, 1, len(resp))
	require.Equal(t, 11.0, resp[0]["total volume"])
	require.Equal(t, 3.0, resp[0]["total impressions"])
	require.Equal(t, "msn.com", resp[0]["domain"])
	require.Equal(t, nil, resp[0]["publisher"])
}
