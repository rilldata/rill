package resolvers

import (
	"context"
	"encoding/json"
	"testing"

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
