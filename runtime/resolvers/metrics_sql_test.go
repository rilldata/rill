package resolvers

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func Test_parsedSQL(t *testing.T) {
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
			"select * from ad_bids_metrics",
			result{
				sql:  "select * FROM \"ad_bids\"",
				deps: []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}},
			},
		},
		{
			"simple quoted",
			"select * from \"ad_bids_metrics\"",
			result{
				sql:  "select * FROM \"ad_bids\"",
				deps: []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}},
			},
		},
		{
			"aggregate",
			"SELECT pub,domain_parts,dom,tld,null_publisher,AGGREGATE(measure_0),AGGREGATE(measure_1) FROM ad_bids_metrics GROUP BY ALL",
			result{
				sql:  "SELECT pub,domain_parts,dom,tld,null_publisher,count(*),avg(bid_price) FROM \"ad_bids\" GROUP BY ALL",
				deps: []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}},
			},
		},
		{
			"aggregate with mv appended",
			"SELECT pub,domain_parts,dom,tld,null_publisher,AGGREGATE(\"ad_bids_metrics\".measure_0),AGGREGATE(ad_bids_metrics.measure_1) FROM ad_bids_metrics GROUP BY ALL",
			result{
				sql:  "SELECT pub,domain_parts,dom,tld,null_publisher,count(*),avg(bid_price) FROM \"ad_bids\" GROUP BY ALL",
				deps: []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}},
			},
		},
		{
			"aggregate with mv appended and quoted",
			"SELECT pub,domain_parts,dom,tld,null_publisher,AGGREGATE(\"ad_bids_metrics\".\"measure_0\"),AGGREGATE(ad_bids_metrics.\"measure_1\") FROM ad_bids_metrics GROUP BY ALL",
			result{
				sql:  "SELECT pub,domain_parts,dom,tld,null_publisher,count(*),avg(bid_price) FROM \"ad_bids\" GROUP BY ALL",
				deps: []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}},
			},
		},
		{
			"aggregate and spaces with policy",
			`SELECT pub,dom,AGGREGATE("bid's number"),AGGREGATE("total volume"),Aggregate("total click""s") From ad_bids_mini_metrics_with_policy GROUP BY ALL`,
			result{
				sql:  "SELECT pub,dom,count(*),sum(volume),sum(clicks) FROM (SELECT * FROM \"ad_bids_mini\") GROUP BY ALL",
				deps: []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_mini_metrics_with_policy"}},
			},
		},
		{
			"aggregate and join",
			`with a as (
				select
					publisher,
					AGGREGATE(ad_bids_mini_metrics_with_policy."total volume") as total_volume,
					AGGREGATE(ad_bids_mini_metrics_with_policy."total click""s") as total_clicks
				from
					ad_bids_mini_metrics_with_policy
				group by
					publisher
				),
				b as (
				select
					publisher,
					AGGREGATE(ad_bids_metrics."measure_1") as avg_bids
				from
					ad_bids_metrics
				group by
					publisher
				)
				select
					a.publisher,
					a.total_volume,
					a.total_clicks,
					b.avg_bids
				from
					a
				join b on
					a.publisher = b.publisher
				`,
			result{
				sql: `with a as (
					select
						publisher,
						sum(volume) as total_volume,
						sum(clicks) as total_clicks
					FROM (SELECT * FROM "ad_bids_mini")
					group by
						publisher
					),
					b as (
					select
						publisher,
						avg(bid_price) as avg_bids
					FROM "ad_bids"
					group by
						publisher
					)
					select
						a.publisher,
						a.total_volume,
						a.total_clicks,
						b.avg_bids
					from
						a
					join b on
						a.publisher = b.publisher
					`,
				deps: []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_mini_metrics_with_policy"}, {Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &metricsSQLCompiler{
				instanceID: instanceID,
				ctrl:       ctrl,
				sql:        tt.sql,
				dialect:    drivers.DialectDuckDB,
			}

			got, _, deps, err := c.compile(context.Background())
			require.NoError(t, err)

			require.Subset(t, deps, tt.want.deps)
			require.Subset(t, tt.want.deps, deps)

			got = regexp.MustCompile(`\s+`).ReplaceAllString(strings.ReplaceAll(strings.ReplaceAll(got, "\n", " "), "\t", " "), " ")
			tt.want.sql = regexp.MustCompile(`\s+`).ReplaceAllString(strings.ReplaceAll(strings.ReplaceAll(tt.want.sql, "\n", " "), "\t", " "), " ")
			if got != tt.want.sql {
				t.Errorf("parsedSQL() = %v, want %v", got, tt.want.sql)
			}
		})
	}
}

func TestSimpleMVSQLApi(t *testing.T) {
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
	require.Equal(t, 3, len(rows[0]))
	require.Equal(t, "msn.com", rows[0]["domain"])
	require.Equal(t, nil, rows[0]["publisher"])
	require.Equal(t, "2022-03-05T14:49:50.459Z", rows[0]["timestamp"])
}

func TestTemplateMVSQLApi(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

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
	require.Equal(t, 3.0, rows[0]["total_imp"])
	require.Equal(t, "yahoo.com", rows[0]["domain"])
	require.Equal(t, "Yahoo", rows[0]["publisher"])
}

func TestPolicyMVSQLApi(t *testing.T) {
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
	require.Equal(t, nil, rows[0]["total_vol"])
	require.Equal(t, 3.0, rows[0]["total_imp"])
	require.Equal(t, "yahoo.com", rows[0]["domain"])
	require.Equal(t, "Yahoo", rows[0]["publisher"])

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
	require.Equal(t, 11.0, resp[0]["total_vol"])
	require.Equal(t, 3.0, resp[0]["total_imp"])
	require.Equal(t, "msn.com", resp[0]["domain"])
	require.Equal(t, nil, resp[0]["publisher"])
}
