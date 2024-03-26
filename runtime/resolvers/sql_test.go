package resolvers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestSimpleSQLApi(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	api, err := rt.APIForName(context.Background(), instanceID, "simple_sql_api")
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
	require.NoError(t, json.Unmarshal(res.Data, &rows))
	require.Equal(t, 5, len(rows))
	require.Equal(t, 5, len(rows[0]))
	require.Equal(t, 4.09, rows[0]["bid_price"])
	require.Equal(t, "msn.com", rows[0]["domain"])
	require.Equal(t, float64(4000), rows[0]["id"])
	require.Equal(t, nil, rows[0]["publisher"])
	require.Equal(t, "2022-03-05T14:49:50.459Z", rows[0]["timestamp"])
}

func TestTemplateSQLApi(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	api, err := rt.APIForName(context.Background(), instanceID, "templated_sql_api")
	require.NoError(t, err)

	res, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID:         instanceID,
		Resolver:           api.Spec.Resolver,
		ResolverProperties: api.Spec.ResolverProperties.AsMap(),
		Args:               map[string]any{"domain": "sports.yahoo.com"},
		UserAttributes:     nil,
	})
	require.NoError(t, err)
	require.NotNil(t, res)

	var rows []map[string]interface{}
	require.NoError(t, json.Unmarshal(res.Data, &rows))
	require.Equal(t, 5, len(rows))
	require.Equal(t, 5, len(rows[0]))
	require.Equal(t, 1.81, rows[0]["bid_price"])
	require.Equal(t, "sports.yahoo.com", rows[0]["domain"])
	require.Equal(t, float64(9000), rows[0]["id"])
	require.Equal(t, "Yahoo", rows[0]["publisher"])
	require.Equal(t, "2022-02-09T11:58:12.475Z", rows[0]["timestamp"])
	for _, row := range rows {
		require.Equal(t, "sports.yahoo.com", row["domain"])
	}
}

func TestTemplateSQLApi2(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	api, err := rt.APIForName(context.Background(), instanceID, "templated_sql_api_2")
	require.NoError(t, err)

	res, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID:         instanceID,
		Resolver:           api.Spec.Resolver,
		ResolverProperties: api.Spec.ResolverProperties.AsMap(),
		Args:               map[string]any{"pageSize": 5},
		UserAttributes:     map[string]any{"domain": "msn.com"},
	})
	require.NoError(t, err)
	require.NotNil(t, res)

	var rows []map[string]interface{}
	require.NoError(t, json.Unmarshal(res.Data, &rows))
	require.Equal(t, 5, len(rows))
	require.Equal(t, 5, len(rows[0]))
	require.Equal(t, 4.09, rows[0]["bid_price"])
	require.Equal(t, "msn.com", rows[0]["domain"])
	require.Equal(t, float64(4000), rows[0]["id"])
	require.Equal(t, nil, rows[0]["publisher"])
	require.Equal(t, "2022-03-05T14:49:50.459Z", rows[0]["timestamp"])
	for _, row := range rows {
		require.Equal(t, "msn.com", row["domain"])
	}
}
