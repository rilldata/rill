package resolvers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

// TestMetricsResolverExpressionMeasure exercises the "metrics" resolver's declarative path for
// expression measures: properties are mapstructure-decoded straight into a metricsview.Query.
func TestMetricsResolverExpressionMeasure(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids_2rows")

	res, _, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID: instanceID,
		Resolver:   "metrics",
		ResolverProperties: map[string]any{
			"metrics_view": "ad_bids_metrics",
			"dimensions":   []map[string]any{{"name": "domain"}},
			"measures": []map[string]any{
				{"name": "measure_2"},
				{
					"name": "profit",
					"compute": map[string]any{
						"expression": map[string]any{
							"expression":   "measure_1 - measure_2",
							"display_name": "Profit",
						},
					},
				},
			},
			"sort": []map[string]any{{"name": "profit", "desc": true}},
		},
		Claims: &runtime.SecurityClaims{},
	})
	require.NoError(t, err)
	defer res.Close()

	var rows []map[string]any
	require.NoError(t, json.Unmarshal(must(res.MarshalJSON()), &rows))
	require.Len(t, rows, 2)
	require.Equal(t, "yahoo.com", rows[0]["domain"])
	require.Equal(t, 3.0, rows[0]["profit"])
	require.Equal(t, "msn.com", rows[1]["domain"])
	require.Equal(t, 2.0, rows[1]["profit"])

	// The ephemeral measure has no spec entry; its metadata must be synthesized.
	meta := res.Meta()
	require.NotNil(t, meta)
	fields, ok := meta["fields"].([]map[string]any)
	require.True(t, ok)
	var found bool
	for _, f := range fields {
		if f["name"] == "profit" {
			found = true
			require.Equal(t, "Profit", f["display_name"])
		}
	}
	require.True(t, found, "expected synthesized field metadata for the expression measure")
}

// TestMetricsResolverExpressionSecurity checks that expression measures respect measure-level
// security policies: the policy on ad_bids_mini_metrics_with_policy excludes "total volume"
// unless the user's domain is msn.com.
func TestMetricsResolverExpressionSecurity(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	props := map[string]any{
		"metrics_view": "ad_bids_mini_metrics_with_policy",
		"measures": []map[string]any{
			{
				"name": "net",
				"compute": map[string]any{
					"expression": map[string]any{"expression": `"total impressions" - "total volume"`},
				},
			},
		},
	}

	_, _, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID:         instanceID,
		Resolver:           "metrics",
		ResolverProperties: props,
		Claims:             &runtime.SecurityClaims{UserAttributes: map[string]any{"domain": "yahoo.com", "email": "user@yahoo.com"}},
	})
	require.Error(t, err)
	require.ErrorContains(t, err, "total volume")

	res, _, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID:         instanceID,
		Resolver:           "metrics",
		ResolverProperties: props,
		Claims:             &runtime.SecurityClaims{UserAttributes: map[string]any{"domain": "msn.com", "email": "user@msn.com"}},
	})
	require.NoError(t, err)
	defer res.Close()
	var rows []map[string]any
	require.NoError(t, json.Unmarshal(must(res.MarshalJSON()), &rows))
	require.Len(t, rows, 1)
}

// TestMetricsResolverExpressionInferredSecurityRules checks that inferred security rules for
// saved alerts/reports include the measures referenced by an expression measure.
// Without them, the exclusive field-access rule would deny the referenced measures at execution time.
func TestMetricsResolverExpressionInferredSecurityRules(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids_2rows")

	initializer, ok := runtime.ResolverInitializers["metrics"]
	require.True(t, ok)
	resolver, err := initializer(context.Background(), &runtime.ResolverOptions{
		Runtime:    rt,
		InstanceID: instanceID,
		Properties: map[string]any{
			"metrics_view": "ad_bids_metrics",
			"measures": []map[string]any{
				{
					"name": "profit",
					"compute": map[string]any{
						"expression": map[string]any{"expression": "measure_1 - measure_2"},
					},
				},
			},
		},
		Claims: &runtime.SecurityClaims{},
	})
	require.NoError(t, err)
	defer resolver.Close()

	rules, err := resolver.InferRequiredSecurityRules()
	require.NoError(t, err)

	var fields []string
	for _, rule := range rules {
		if fa := rule.GetFieldAccess(); fa != nil {
			fields = append(fields, fa.Fields...)
		}
	}
	require.Contains(t, fields, "measure_1")
	require.Contains(t, fields, "measure_2")
}
