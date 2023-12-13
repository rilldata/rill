package server_test

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestServer_MetricsViewTotals(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 2.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_row_null_exclude(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Where: filterNotInClause(
			filterColumn("device"),
			[]*runtimev1.Expression{filterValue(structpb.NewStringValue("iphone"))},
		),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_row_null_exclude_null(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Where: filterNotInClause(
			filterColumn("device"),
			[]*runtimev1.Expression{filterValue(structpb.NewNullValue())},
		),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_row_null_exclude_all(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Where: filterNotInClause(
			filterColumn("device"),
			[]*runtimev1.Expression{filterValue(structpb.NewNullValue()), filterValue(structpb.NewStringValue("iphone"))},
		),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 0.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_row_null_include(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Where: filterInClause(
			filterColumn("device"),
			[]*runtimev1.Expression{filterValue(structpb.NewStringValue("iphone"))},
		),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_row_null_include_null(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Where: filterInClause(
			filterColumn("device"),
			[]*runtimev1.Expression{filterValue(structpb.NewNullValue())},
		),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_row_null_include_all(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Where: filterInClause(
			filterColumn("device"),
			[]*runtimev1.Expression{filterValue(structpb.NewNullValue()), filterValue(structpb.NewStringValue("iphone"))},
		),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 2.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_row_null_exclude_like(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Where: filterNotLikeClause(
			filterColumn("device"),
			filterValue(structpb.NewStringValue("iphone")),
		),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_row_null_exclude_like_and_null(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Where: filterAndClause([]*runtimev1.Expression{
			filterNotInClause(
				filterColumn("device"),
				[]*runtimev1.Expression{filterValue(structpb.NewNullValue())},
			),
			filterNotLikeClause(
				filterColumn("device"),
				filterValue(structpb.NewStringValue("iphone")),
			),
		}),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 0.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_row_null_exclude_like_doesntexist(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Where: filterNotInClause(
			filterColumn("device"),
			[]*runtimev1.Expression{filterValue(structpb.NewStringValue("doesntexist"))},
		),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 2.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_timestamp_name_with_spaces(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics_garbled",
		MeasureNames:    []string{"measure_0"},
		TimeEnd:         parseTimeToProtoTimeStamps(t, "2022-01-02T00:00:00Z"),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_EmptyModel(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "no_rows_metrics",
		MeasureNames:    []string{"measure_0", "measure_1"},
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Data.Fields))
	require.Equal(t, 0.0, tr.Data.Fields["measure_0"].GetNumberValue())
	require.Equal(t, 0.0, tr.Data.Fields["measure_2"].GetNumberValue())
}

func TestServer_MetricsViewTotals_2measures(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")
	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0", "measure_1"},
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Data.Fields))
	require.Equal(t, 2.0, tr.Data.Fields["measure_0"].GetNumberValue())
	require.Equal(t, 8.0, tr.Data.Fields["measure_1"].GetNumberValue())
}

func TestServer_MetricsViewTotals_TimeStart(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		TimeStart:       parseTimeToProtoTimeStamps(t, "2022-01-02T00:00:00Z"),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_TimeEnd(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		TimeEnd:         parseTimeToProtoTimeStamps(t, "2022-01-02T00:00:00Z"),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_TimeStart_TimeEnd(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		TimeStart:       parseTimeToProtoTimeStamps(t, "2022-01-01T00:00:00Z"),
		TimeEnd:         parseTimeToProtoTimeStamps(t, "2022-01-02T00:00:00Z"),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_1dim(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Where: filterInClause(
			filterColumn("domain"),
			[]*runtimev1.Expression{filterValue(structpb.NewStringValue("msn.com"))},
		),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_1dim_special_symbol_values(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Where: filterInClause(
			filterColumn("domain"),
			[]*runtimev1.Expression{
				filterValue(structpb.NewStringValue("msn.'com")),
				filterValue(structpb.NewStringValue("msn.\"com")),
				filterValue(structpb.NewStringValue("msn. com")),
			},
		),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 0.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_1dim_2In(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Where: filterInClause(
			filterColumn("domain"),
			[]*runtimev1.Expression{filterValue(structpb.NewStringValue("msn.com")), filterValue(structpb.NewStringValue("yahoo.com"))},
		),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 2.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_2dim(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Where: filterAndClause([]*runtimev1.Expression{
			filterNotInClause(
				filterColumn("domain"),
				[]*runtimev1.Expression{filterValue(structpb.NewStringValue("yahoo.com"))},
			),
			filterNotLikeClause(
				filterColumn("publisher"),
				filterValue(structpb.NewStringValue("Yahoo")),
			),
		}),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_1dim_like(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Where: filterLikeClause(
			filterColumn("domain"),
			filterValue(structpb.NewStringValue("%com")),
		),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 2.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_1dim_in_and_like(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Where: filterOrClause([]*runtimev1.Expression{
			filterInClause(filterColumn("domain"), []*runtimev1.Expression{filterValue(structpb.NewStringValue("yahoo"))}),
			filterLikeClause(filterColumn("domain"), filterValue(structpb.NewStringValue("%com"))),
		}),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 2.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_1dim_2like(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Where: filterOrClause([]*runtimev1.Expression{
			filterLikeClause(filterColumn("domain"), filterValue(structpb.NewStringValue("msn%"))),
			filterLikeClause(filterColumn("domain"), filterValue(structpb.NewStringValue("%com"))),
		}),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 2.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_1dim_include_and_exclude(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Where: filterAndClause([]*runtimev1.Expression{
			filterLikeClause(filterColumn("domain"), filterValue(structpb.NewStringValue("%com"))),
			filterNotInClause(filterColumn("domain"), []*runtimev1.Expression{filterValue(structpb.NewStringValue("yahoo.com"))}),
		}),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_1dim_null(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Where: filterAndClause([]*runtimev1.Expression{
			filterInClause(filterColumn("publisher"), []*runtimev1.Expression{filterValue(structpb.NewNullValue())}),
		}),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_1dim_include_and_exclude_in_and_like(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Where: filterAndClause([]*runtimev1.Expression{
			filterInClause(filterColumn("domain"), []*runtimev1.Expression{filterValue(structpb.NewStringValue("msn.com"))}),
			filterLikeClause(filterColumn("domain"), filterValue(structpb.NewStringValue("%yahoo%"))),
			filterNotInClause(filterColumn("publisher"), []*runtimev1.Expression{filterValue(structpb.NewNullValue())}),
			filterNotLikeClause(filterColumn("publisher"), filterValue(structpb.NewStringValue("Y%"))),
		}),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 0.0, tr.Data.Fields["measure_0"].GetNumberValue())
}
