package server_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestExport(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"m1.sql": `
SELECT 'US' AS country
`,
			"mv1.yaml": `
type: metrics_view
model: m1
dimensions:
- column: country
measures:
- name: c
  expression: COUNT(*)
explore:
  skip: true
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	// Check it generates a download token
	resp, err := server.Export(testCtx(), &runtimev1.ExportRequest{
		InstanceId: instanceID,
		Format:     runtimev1.ExportFormat_EXPORT_FORMAT_CSV,
		Query: &runtimev1.Query{
			Query: &runtimev1.Query_MetricsViewAggregationRequest{
				MetricsViewAggregationRequest: &runtimev1.MetricsViewAggregationRequest{
					InstanceId:  instanceID,
					MetricsView: "mv1",
					Dimensions:  []*runtimev1.MetricsViewAggregationDimension{{Name: "country"}},
					Measures:    []*runtimev1.MetricsViewAggregationMeasure{{Name: "count"}},
				},
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Greater(t, len(resp.DownloadUrlPath), 0)

	// Check it errors for very large download tokens
	v := make([]byte, 1<<18) // 256 KB
	_, err = rand.Read(v)
	require.NoError(t, err)
	resp, err = server.Export(testCtx(), &runtimev1.ExportRequest{
		InstanceId: instanceID,
		Format:     runtimev1.ExportFormat_EXPORT_FORMAT_CSV,
		Query: &runtimev1.Query{
			Query: &runtimev1.Query_MetricsViewAggregationRequest{
				MetricsViewAggregationRequest: &runtimev1.MetricsViewAggregationRequest{
					InstanceId:  instanceID,
					MetricsView: "mv1",
					Dimensions:  []*runtimev1.MetricsViewAggregationDimension{{Name: "country"}},
					Measures:    []*runtimev1.MetricsViewAggregationMeasure{{Name: "count"}},
					WhereSql:    fmt.Sprintf("country = '%s'", hex.EncodeToString(v)),
				},
			},
		},
	})
	require.Error(t, err)
	require.ErrorContains(t, err, "exceeds maximum allowed size")
}
