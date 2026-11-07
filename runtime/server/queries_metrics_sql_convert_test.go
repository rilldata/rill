package server_test

import (
	"context"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestConvertMetricsSQLToQuery(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"m1.sql":    `SELECT 'US' AS country, 1 AS value, TIMESTAMP '2024-01-01' AS ts`,
			"mv1.yaml": `
type: metrics_view
version: 1
model: m1
timeseries: ts
dimensions:
- column: country
measures:
- name: total
  expression: SUM(value)

security:
  access: "'{{ .user.domain }}' = 'rilldata.com'"
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	srv, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	allowed := auth.WithClaims(context.Background(), &runtime.SecurityClaims{
		UserAttributes: map[string]any{"domain": "rilldata.com"},
		Permissions:    []runtime.Permission{runtime.ReadMetrics},
	})
	sql := "SELECT country, total FROM mv1 WHERE ts < now() HAVING total > 10 ORDER BY total DESC LIMIT 5"

	// The compiled query keeps the HAVING apart from the WHERE, and now() is the given execution time.
	res, err := srv.ConvertMetricsSQLToQuery(allowed, &runtimev1.ConvertMetricsSQLToQueryRequest{
		InstanceId:    instanceID,
		Sql:           sql,
		ExecutionTime: timestamppb.New(time.Date(2024, 2, 1, 7, 0, 0, 0, time.UTC)),
	})
	require.NoError(t, err)
	q := res.Query.AsMap()
	require.Equal(t, "mv1", q["metrics_view"])
	require.Equal(t, []any{map[string]any{"name": "country"}}, q["dimensions"])
	require.Equal(t, map[string]any{"cond": map[string]any{"op": "lt", "exprs": []any{
		map[string]any{"name": "ts"},
		map[string]any{"val": "2024-02-01T07:00:00Z"},
	}}}, q["where"])
	require.Equal(t, map[string]any{"cond": map[string]any{"op": "gt", "exprs": []any{
		map[string]any{"name": "total"},
		map[string]any{"val": float64(10)},
	}}}, q["having"])

	// A caller that can't access the metrics view doesn't get its query.
	denied := auth.WithClaims(context.Background(), &runtime.SecurityClaims{
		UserAttributes: map[string]any{"domain": "example.com"},
		Permissions:    []runtime.Permission{runtime.ReadMetrics},
	})
	_, err = srv.ConvertMetricsSQLToQuery(denied, &runtimev1.ConvertMetricsSQLToQueryRequest{InstanceId: instanceID, Sql: sql})
	require.Equal(t, codes.PermissionDenied, status.Code(err))

	// Nor does one without the permission to read metrics.
	noRead := auth.WithClaims(context.Background(), &runtime.SecurityClaims{
		UserAttributes: map[string]any{"domain": "rilldata.com"},
	})
	_, err = srv.ConvertMetricsSQLToQuery(noRead, &runtimev1.ConvertMetricsSQLToQueryRequest{InstanceId: instanceID, Sql: sql})
	require.Equal(t, codes.PermissionDenied, status.Code(err))

	// Invalid metrics SQL is the caller's error.
	_, err = srv.ConvertMetricsSQLToQuery(allowed, &runtimev1.ConvertMetricsSQLToQueryRequest{InstanceId: instanceID, Sql: "SELECT 1"})
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}
