package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rilldata/rill/runtime/api"
)

func TestServer_MetricsViewTotals(t *testing.T) {
	server, instanceId := getTestServer(t)
	rows := CreateSimpleTimeseriesTable(server, instanceId, t, "timeseries")
	rows.Close()

	cntName := "cnt"
	r, err := server.MetricsViewTotals(context.Background(), &api.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "timeseries",
		BasicMeasures: []*api.BasicMeasureDefinition{
			{
				SqlName:    &cntName,
				Expression: "count(*)",
			},
		},
		TimestampColumnName: "time",
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(r.Data.Fields))
	require.Equal(t, 2.0, r.Data.Fields["cnt"].GetNumberValue())
}

func TestServer_MetricsViewTotals_2measures(t *testing.T) {
	server, instanceId := getTestServer(t)
	rows := CreateSimpleTimeseriesTable(server, instanceId, t, "timeseries")
	rows.Close()

	cntName := "cnt"
	mx := "max"
	r, err := server.MetricsViewTotals(context.Background(), &api.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "timeseries",
		BasicMeasures: []*api.BasicMeasureDefinition{
			{
				SqlName:    &cntName,
				Expression: "count(*)",
			},
			{
				SqlName:    &mx,
				Expression: "max(clicks)",
			},
		},
		TimestampColumnName: "time",
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(r.Data.Fields))
	require.Equal(t, 2.0, r.Data.Fields["cnt"].GetNumberValue())
	require.Equal(t, 1.0, r.Data.Fields["max"].GetNumberValue())
}
