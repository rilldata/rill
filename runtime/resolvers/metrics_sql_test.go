package resolvers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestSimpleMetricsSQLApi(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	api, err := rt.APIForName(context.Background(), instanceID, "simple_mv_sql_api")
	require.NoError(t, err)

	res, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID:         instanceID,
		Resolver:           api.Spec.Resolver,
		ResolverProperties: api.Spec.ResolverProperties.AsMap(),
		Args:               nil,
		Claims:             &runtime.SecurityClaims{},
	})
	require.NoError(t, err)
	defer res.Close()

	require.NotNil(t, res)
	var rows []map[string]interface{}
	require.NoError(t, json.Unmarshal(must(res.MarshalJSON()), &rows))
	require.Equal(t, 1, len(rows))
	require.Equal(t, 2, len(rows[0]))
	require.Equal(t, "msn.com", rows[0]["dom"])
	require.Equal(t, "Microsoft", rows[0]["pub"])
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
		Claims:             &runtime.SecurityClaims{},
	})
	require.NoError(t, err)
	defer res.Close()

	require.NotNil(t, res)
	var rows []map[string]interface{}
	require.NoError(t, json.Unmarshal(must(res.MarshalJSON()), &rows))
	require.Equal(t, 1, len(rows))
	require.Equal(t, 3.0, rows[0]["measure_2"])
	require.Equal(t, "yahoo.com", rows[0]["domain"])
	require.Equal(t, "Yahoo", rows[0]["publisher"])
}

func TestComplexTemplateMetricsSQLAPI(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	testruntime.RequireParseErrors(t, rt, instanceID, nil)

	api, err := rt.APIForName(context.Background(), instanceID, "templated_mv_sql_api_2")
	require.NoError(t, err)

	res, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID:         instanceID,
		Resolver:           api.Spec.Resolver,
		ResolverProperties: api.Spec.ResolverProperties.AsMap(),
		Args:               map[string]any{"domain": "yahoo.com", "pageSize": ""},
		Claims:             &runtime.SecurityClaims{UserAttributes: map[string]any{"domain": "yahoo.com"}},
	})
	require.NoError(t, err)
	defer res.Close()

	require.NotNil(t, res)
	var rows []map[string]interface{}
	require.NoError(t, json.Unmarshal(must(res.MarshalJSON()), &rows))
	require.Equal(t, 1, len(rows))
	require.Equal(t, 3.0, rows[0]["measure_2"])
	require.Equal(t, "yahoo.com", rows[0]["domain"])
	require.Equal(t, "Yahoo", rows[0]["publisher"])
}

func TestPolicyMetricsSQLAPI(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	api, err := rt.APIForName(context.Background(), instanceID, "mv_sql_policy_api")
	require.NoError(t, err)

	_, err = rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID:         instanceID,
		Resolver:           api.Spec.Resolver,
		ResolverProperties: api.Spec.ResolverProperties.AsMap(),
		Args:               nil,
		Claims:             &runtime.SecurityClaims{UserAttributes: map[string]any{"domain": "yahoo.com", "email": "user@yahoo.com"}},
	})
	require.Error(t, err)

	api, err = rt.APIForName(context.Background(), instanceID, "mv_sql_policy_api")
	require.NoError(t, err)

	res, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID:         instanceID,
		Resolver:           api.Spec.Resolver,
		ResolverProperties: api.Spec.ResolverProperties.AsMap(),
		Args:               nil,
		Claims:             &runtime.SecurityClaims{UserAttributes: map[string]any{"domain": "msn.com", "email": "user@msn.com"}},
	})
	require.NoError(t, err)
	defer res.Close()

	require.NotNil(t, res)
	var resp []map[string]interface{}
	require.NoError(t, json.Unmarshal(must(res.MarshalJSON()), &resp))
	require.Equal(t, 1, len(resp))
	require.Equal(t, 11.0, resp[0]["total volume"])
	require.Equal(t, 3.0, resp[0]["total impressions"])
	require.Equal(t, "msn.com", resp[0]["domain"])
	require.Equal(t, nil, resp[0]["publisher"])
}

func TestMetricsSQLWithAdditionalWhere(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	testruntime.RequireParseErrors(t, rt, instanceID, nil)

	res, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID: instanceID,
		Resolver:   "metrics_sql",
		ResolverProperties: map[string]any{
			"sql": "SELECT dom, pub FROM ad_bids_metrics",
			"additional_where": map[string]any{
				"cond": map[string]any{
					"op":    "eq",
					"exprs": []map[string]any{{"name": "dom"}, {"val": "msn.com"}},
				},
			},
		},
		Args:   nil,
		Claims: &runtime.SecurityClaims{},
	})
	require.NoError(t, err)
	defer res.Close()

	require.NotNil(t, res)
	var rows []map[string]interface{}
	require.NoError(t, json.Unmarshal(must(res.MarshalJSON()), &rows))

	// Should return filtered results for msn.com only
	require.Greater(t, len(rows), 0, "Should return filtered results")

	// All returned rows should have dom = "msn.com" due to additional_where filter
	for _, row := range rows {
		require.Equal(t, "msn.com", row["dom"], "All rows should have dom = msn.com due to additional_where filter")
	}

	resNoFilter, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID: instanceID,
		Resolver:   "metrics_sql",
		ResolverProperties: map[string]any{
			"sql": "SELECT dom, pub FROM ad_bids_metrics",
		},
		Args:   nil,
		Claims: &runtime.SecurityClaims{},
	})
	require.NoError(t, err)
	defer resNoFilter.Close()

	var rowsNoFilter []map[string]interface{}
	require.NoError(t, json.Unmarshal(must(resNoFilter.MarshalJSON()), &rowsNoFilter))

	// Without the filter, we should get more domains than just msn.com
	require.GreaterOrEqual(t, len(rowsNoFilter), len(rows), "Unfiltered results should have same or more rows")

	// Verify unfiltered results include multiple domains
	domains := make(map[string]bool)
	for _, row := range rowsNoFilter {
		if domain, ok := row["dom"].(string); ok {
			domains[domain] = true
		}
	}
	require.Greater(t, len(domains), 1, "Unfiltered results should include multiple domains")
}

func TestMetricsSQLWithAdditionalTimeRange(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	additionalTimeRange := map[string]any{
		"start": time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		"end":   time.Date(2022, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	testruntime.RequireParseErrors(t, rt, instanceID, nil)
	res, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID: instanceID,
		Resolver:   "metrics_sql",
		ResolverProperties: map[string]any{
			"sql":                   "SELECT dom, pub FROM ad_bids_metrics",
			"additional_time_range": additionalTimeRange,
		},
		Args:   nil,
		Claims: &runtime.SecurityClaims{},
	})
	require.NoError(t, err)
	defer res.Close()

	require.NotNil(t, res)
	var rows []map[string]interface{}
	require.NoError(t, json.Unmarshal(must(res.MarshalJSON()), &rows))
	require.Greater(t, len(rows), 0, "Should return filtered results for the additional time range")

	// Use additional_time_range and additional_where to filter results
	resWithFilter, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID: instanceID,
		Resolver:   "metrics_sql",
		ResolverProperties: map[string]any{
			"sql":                   "SELECT dom, pub FROM ad_bids_metrics",
			"additional_time_range": additionalTimeRange,
			"additional_where": map[string]any{
				"cond": map[string]any{
					"op":    "eq",
					"exprs": []map[string]any{{"name": "dom"}, {"val": "msn.com"}},
				},
			},
		},
		Args:   nil,
		Claims: &runtime.SecurityClaims{},
	})
	require.NoError(t, err)
	defer resWithFilter.Close()

	var rowsWithFilter []map[string]interface{}
	require.NoError(t, json.Unmarshal(must(resWithFilter.MarshalJSON()), &rowsWithFilter))

	t.Logf("Rows with filter: %v", rowsWithFilter)

	require.Greater(t, len(rowsWithFilter), 0, "Should return filtered results for the additional time range and where clause")
}

func TestMetricsSQLWithAdditionalTimeZone(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	// Use a timezone and check that the query does not error and returns data
	res, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID: instanceID,
		Resolver:   "metrics_sql",
		ResolverProperties: map[string]any{
			"sql":       "SELECT dom, pub FROM ad_bids_metrics",
			"time_zone": "America/New_York",
		},
		Args:   nil,
		Claims: &runtime.SecurityClaims{},
	})
	require.NoError(t, err)
	defer res.Close()

	var rows []map[string]interface{}
	require.NoError(t, json.Unmarshal(must(res.MarshalJSON()), &rows))
	require.Greater(t, len(rows), 0, "Should return rows for valid timezone filter")

	// Use an invalid timezone and expect an error
	_, err = rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID: instanceID,
		Resolver:   "metrics_sql",
		ResolverProperties: map[string]any{
			"sql":       "SELECT dom, pub FROM ad_bids_metrics",
			"time_zone": "Invalid/Timezone",
		},
		Args:   nil,
		Claims: &runtime.SecurityClaims{},
	})

	require.Error(t, err, "Should error for invalid timezone")
}
