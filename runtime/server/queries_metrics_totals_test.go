package server_test

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/expressionpb"
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
		Where: expressionpb.NotIn(
			expressionpb.Identifier("device"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewStringValue("iphone"))},
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
		Where: expressionpb.NotIn(
			expressionpb.Identifier("device"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewNullValue())},
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
		Where: expressionpb.NotIn(
			expressionpb.Identifier("device"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewNullValue()), expressionpb.Value(structpb.NewStringValue("iphone"))},
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
		Where: expressionpb.In(
			expressionpb.Identifier("device"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewStringValue("iphone"))},
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
		Where: expressionpb.In(
			expressionpb.Identifier("device"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewNullValue())},
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
		Where: expressionpb.In(
			expressionpb.Identifier("device"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewNullValue()), expressionpb.Value(structpb.NewStringValue("iphone"))},
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
		Where: expressionpb.NotLike(
			expressionpb.Identifier("device"),
			expressionpb.Value(structpb.NewStringValue("iphone")),
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
		Where: expressionpb.And([]*runtimev1.Expression{
			expressionpb.NotIn(
				expressionpb.Identifier("device"),
				[]*runtimev1.Expression{expressionpb.Value(structpb.NewNullValue())},
			),
			expressionpb.NotLike(
				expressionpb.Identifier("device"),
				expressionpb.Value(structpb.NewStringValue("iphone")),
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
		Where: expressionpb.NotIn(
			expressionpb.Identifier("device"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewStringValue("doesntexist"))},
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
		Where: expressionpb.In(
			expressionpb.Identifier("domain"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewStringValue("msn.com"))},
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
		Where: expressionpb.In(
			expressionpb.Identifier("domain"),
			[]*runtimev1.Expression{
				expressionpb.Value(structpb.NewStringValue("msn.'com")),
				expressionpb.Value(structpb.NewStringValue("msn.\"com")),
				expressionpb.Value(structpb.NewStringValue("msn. com")),
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
		Where: expressionpb.In(
			expressionpb.Identifier("domain"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewStringValue("msn.com")), expressionpb.Value(structpb.NewStringValue("yahoo.com"))},
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
		Where: expressionpb.And([]*runtimev1.Expression{
			expressionpb.NotIn(
				expressionpb.Identifier("domain"),
				[]*runtimev1.Expression{expressionpb.Value(structpb.NewStringValue("yahoo.com"))},
			),
			expressionpb.NotLike(
				expressionpb.Identifier("publisher"),
				expressionpb.Value(structpb.NewStringValue("Yahoo")),
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
		Where: expressionpb.Like(
			expressionpb.Identifier("domain"),
			expressionpb.Value(structpb.NewStringValue("%com")),
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
		Where: expressionpb.Or([]*runtimev1.Expression{
			expressionpb.In(expressionpb.Identifier("domain"), []*runtimev1.Expression{expressionpb.Value(structpb.NewStringValue("yahoo"))}),
			expressionpb.Like(expressionpb.Identifier("domain"), expressionpb.Value(structpb.NewStringValue("%com"))),
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
		Where: expressionpb.Or([]*runtimev1.Expression{
			expressionpb.Like(expressionpb.Identifier("domain"), expressionpb.Value(structpb.NewStringValue("msn%"))),
			expressionpb.Like(expressionpb.Identifier("domain"), expressionpb.Value(structpb.NewStringValue("%com"))),
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
		Where: expressionpb.And([]*runtimev1.Expression{
			expressionpb.Like(expressionpb.Identifier("domain"), expressionpb.Value(structpb.NewStringValue("%com"))),
			expressionpb.NotIn(expressionpb.Identifier("domain"), []*runtimev1.Expression{expressionpb.Value(structpb.NewStringValue("yahoo.com"))}),
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
		Where: expressionpb.And([]*runtimev1.Expression{
			expressionpb.In(expressionpb.Identifier("publisher"), []*runtimev1.Expression{expressionpb.Value(structpb.NewNullValue())}),
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
		Where: expressionpb.And([]*runtimev1.Expression{
			expressionpb.In(expressionpb.Identifier("domain"), []*runtimev1.Expression{expressionpb.Value(structpb.NewStringValue("msn.com"))}),
			expressionpb.Like(expressionpb.Identifier("domain"), expressionpb.Value(structpb.NewStringValue("%yahoo%"))),
			expressionpb.NotIn(expressionpb.Identifier("publisher"), []*runtimev1.Expression{expressionpb.Value(structpb.NewNullValue())}),
			expressionpb.NotLike(expressionpb.Identifier("publisher"), expressionpb.Value(structpb.NewStringValue("Y%"))),
		}),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 0.0, tr.Data.Fields["measure_0"].GetNumberValue())
}
