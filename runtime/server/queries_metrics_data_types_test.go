package server_test

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
)

func TestMetricsViewDataTypes(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	res, err := server.MetricsViewDataTypes(
		testCtx(),
		&runtimev1.MetricsViewDataTypesRequest{
			InstanceId:      instanceId,
			MetricsViewName: "ad_bids_metrics",
		},
	)
	require.NoError(t, err)
	types := res.DataTypes
	require.Len(t, types, 7)

	i := 0
	require.Equal(t, types[i].Name, "pub")
	require.Equal(t, types[i].Type, "VARCHAR")

	i++
	require.Equal(t, types[i].Name, "dom")
	require.Equal(t, types[i].Type, "VARCHAR")

	i++
	require.Equal(t, types[i].Name, "domain_parts")
	require.Equal(t, types[i].Type, "VARCHAR")

	i++
	require.Equal(t, types[i].Name, "tld")
	require.Equal(t, types[i].Type, "VARCHAR")

	i++
	require.Equal(t, types[i].Name, "null_publisher")
	require.Equal(t, types[i].Type, "BOOLEAN")

	i++
	require.Equal(t, types[i].Name, "measure_0")
	require.Equal(t, types[i].Type, "BIGINT")

	i++
	require.Equal(t, types[i].Name, "measure_1")
	require.Equal(t, types[i].Type, "DOUBLE")
}
