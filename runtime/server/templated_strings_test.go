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
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestResolveTemplatedString_BasicTemplate(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
		},
		Variables: map[string]string{
			"test_var": "hello",
		},
	})

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	// Test basic Sprig template functions
	res, err := server.ResolveTemplatedString(testCtx(), &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Body:       "Hello {{ upper .env.test_var }}!",
	})
	require.NoError(t, err)
	require.Equal(t, "Hello HELLO!", res.Body)
}

func TestResolveTemplatedString_UserAttributes(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
		},
	})

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	ctx := auth.WithClaims(context.Background(), &runtime.SecurityClaims{
		SkipChecks: true,
		UserAttributes: map[string]any{
			"name":  "John Doe",
			"email": "john@example.com",
		},
	})

	res, err := server.ResolveTemplatedString(ctx, &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Body:       "Welcome {{ .user.name }} ({{ .user.email }})",
	})
	require.NoError(t, err)
	require.Equal(t, "Welcome John Doe (john@example.com)", res.Body)
}

func TestResolveTemplatedString_MetricsSQL_SimpleQuery(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"model.sql": `
SELECT 'US' AS country, 100 AS revenue
UNION ALL
SELECT 'UK' AS country, 200 AS revenue
`,
			"mv.yaml": `
type: metrics_view
version: 1
model: model
dimensions:
- column: country
measures:
- name: total_revenue
  expression: SUM(revenue)
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	res, err := server.ResolveTemplatedString(testCtx(), &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Body:       `Total revenue is {{ metrics_sql "SELECT total_revenue FROM mv" }}`,
	})
	require.NoError(t, err)
	require.Equal(t, "Total revenue is 300", res.Body)
}

func TestResolveTemplatedString_MetricsSQL_WithFormatTokens(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"model.sql": `
SELECT 'US' AS country, 100.5 AS revenue
UNION ALL
SELECT 'UK' AS country, 200.3 AS revenue
`,
			"mv.yaml": `
type: metrics_view
version: 1
model: model
dimensions:
- column: country
measures:
- name: total_revenue
  expression: SUM(revenue)
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	res, err := server.ResolveTemplatedString(testCtx(), &runtimev1.ResolveTemplatedStringRequest{
		InstanceId:      instanceID,
		Body:            `Total: {{ metrics_sql "SELECT total_revenue FROM mv" }}`,
		UseFormatTokens: true,
	})
	require.NoError(t, err)
	require.Contains(t, res.Body, `__RILL__FORMAT__("mv", "total_revenue", 300.8)`)
}

func TestResolveTemplatedString_MetricsSQL_MultipleQueries(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"model.sql": `
SELECT 'US' AS country, 100 AS revenue, 5 AS orders
UNION ALL
SELECT 'UK' AS country, 200 AS revenue, 10 AS orders
`,
			"mv.yaml": `
type: metrics_view
version: 1
model: model
dimensions:
- column: country
measures:
- name: total_revenue
  expression: SUM(revenue)
- name: total_orders
  expression: SUM(orders)
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	res, err := server.ResolveTemplatedString(testCtx(), &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Body: `Revenue: {{ metrics_sql "SELECT total_revenue FROM mv" }}
Orders: {{ metrics_sql "SELECT total_orders FROM mv" }}`,
	})
	require.NoError(t, err)
	require.Contains(t, res.Body, "Revenue: 300")
	require.Contains(t, res.Body, "Orders: 15")
}

func TestResolveTemplatedString_MetricsSQL_WithFilters(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"model.sql": `
SELECT 'US' AS country, 100 AS revenue
UNION ALL
SELECT 'UK' AS country, 200 AS revenue
`,
			"mv.yaml": `
type: metrics_view
version: 1
model: model
dimensions:
- column: country
measures:
- name: total_revenue
  expression: SUM(revenue)
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	res, err := server.ResolveTemplatedString(testCtx(), &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Body:       `US Revenue: {{ metrics_sql "SELECT total_revenue FROM mv WHERE country = 'US'" }}`,
	})
	require.NoError(t, err)
	require.Equal(t, "US Revenue: 100", res.Body)
}

func TestResolveTemplatedString_MetricsSQL_WithAdditionalWhere(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"model.sql": `
SELECT 'US' AS country, 100 AS revenue
UNION ALL
SELECT 'UK' AS country, 200 AS revenue
UNION ALL
SELECT 'CA' AS country, 150 AS revenue
`,
			"mv.yaml": `
type: metrics_view
version: 1
model: model
dimensions:
- column: country
measures:
- name: total_revenue
  expression: SUM(revenue)
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	// Create additional where clause: country IN ('US', 'UK')
	additionalWhere := &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op: runtimev1.Operation_OPERATION_IN,
				Exprs: []*runtimev1.Expression{
					{Expression: &runtimev1.Expression_Ident{Ident: "country"}},
					{
						Expression: &runtimev1.Expression_Val{
							Val: structpb.NewListValue(&structpb.ListValue{
								Values: []*structpb.Value{
									structpb.NewStringValue("US"),
									structpb.NewStringValue("UK"),
								},
							}),
						},
					},
				},
			},
		},
	}

	res, err := server.ResolveTemplatedString(testCtx(), &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Body:       `Revenue: {{ metrics_sql "SELECT total_revenue FROM mv" }}`,
		AdditionalWhereByMetricsView: map[string]*runtimev1.Expression{
			"mv": additionalWhere,
		},
	})
	require.NoError(t, err)
	// Should exclude CA, so 100 + 200 = 300
	require.Equal(t, "Revenue: 300", res.Body)
}

func TestResolveTemplatedString_MetricsSQL_MultipleMetricsViewsWithDifferentFilters(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"model1.sql": `
SELECT 'US' AS country, 100 AS revenue
UNION ALL
SELECT 'UK' AS country, 200 AS revenue
UNION ALL
SELECT 'CA' AS country, 300 AS revenue
`,
			"mv1.yaml": `
type: metrics_view
version: 1
model: model1
dimensions:
- column: country
measures:
- name: total_revenue
  expression: SUM(revenue)
`,
			"model2.sql": `
SELECT 'Electronics' AS category, 500 AS sales
UNION ALL
SELECT 'Clothing' AS category, 250 AS sales
UNION ALL
SELECT 'Food' AS category, 150 AS sales
`,
			"mv2.yaml": `
type: metrics_view
version: 1
model: model2
dimensions:
- column: category
measures:
- name: total_sales
  expression: SUM(sales)
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 5, 0, 0)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	// Filter for mv1: only US and UK
	filterMv1 := &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op: runtimev1.Operation_OPERATION_IN,
				Exprs: []*runtimev1.Expression{
					{Expression: &runtimev1.Expression_Ident{Ident: "country"}},
					{
						Expression: &runtimev1.Expression_Val{
							Val: structpb.NewListValue(&structpb.ListValue{
								Values: []*structpb.Value{
									structpb.NewStringValue("US"),
									structpb.NewStringValue("UK"),
								},
							}),
						},
					},
				},
			},
		},
	}

	// Filter for mv2: only Electronics
	filterMv2 := &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op: runtimev1.Operation_OPERATION_EQ,
				Exprs: []*runtimev1.Expression{
					{Expression: &runtimev1.Expression_Ident{Ident: "category"}},
					{Expression: &runtimev1.Expression_Val{Val: structpb.NewStringValue("Electronics")}},
				},
			},
		},
	}

	res, err := server.ResolveTemplatedString(testCtx(), &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Body:       `Revenue: {{ metrics_sql "SELECT total_revenue FROM mv1" }}, Sales: {{ metrics_sql "SELECT total_sales FROM mv2" }}`,
		AdditionalWhereByMetricsView: map[string]*runtimev1.Expression{
			"mv1": filterMv1,
			"mv2": filterMv2,
		},
	})
	require.NoError(t, err)
	// mv1: 100 + 200 = 300 (excludes CA)
	// mv2: 500 (only Electronics)
	require.Equal(t, "Revenue: 300, Sales: 500", res.Body)
}

func TestResolveTemplatedString_MetricsSQL_ErrorNotSingleValue(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"model.sql": `
SELECT 'US' AS country, 100 AS revenue
UNION ALL
SELECT 'UK' AS country, 200 AS revenue
`,
			"mv.yaml": `
type: metrics_view
version: 1
model: model
dimensions:
- column: country
measures:
- name: total_revenue
  expression: SUM(revenue)
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	// Query returns multiple columns
	res, err := server.ResolveTemplatedString(testCtx(), &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Body:       `{{ metrics_sql "SELECT country, total_revenue FROM mv" }}`,
	})
	require.Error(t, err)
	require.Nil(t, res)
	require.Contains(t, err.Error(), "expected one field")
}

func TestResolveTemplatedString_MetricsSQL_ErrorMultipleRows(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"model.sql": `
SELECT 'US' AS country, 100 AS revenue
UNION ALL
SELECT 'UK' AS country, 200 AS revenue
`,
			"mv.yaml": `
type: metrics_view
version: 1
model: model
dimensions:
- column: country
measures:
- name: total_revenue
  expression: SUM(revenue)
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	// Query returns multiple rows - select dimension without aggregation
	res, err := server.ResolveTemplatedString(testCtx(), &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Body:       `{{ metrics_sql "SELECT country FROM mv" }}`,
	})
	require.Error(t, err)
	require.Nil(t, res)
	require.Contains(t, err.Error(), "multiple rows")
}

func TestResolveTemplatedString_MetricsSQL_ErrorInvalidQuery(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"model.sql": `SELECT 'US' AS country, 100 AS revenue`,
			"mv.yaml": `
type: metrics_view
version: 1
model: model
dimensions:
- column: country
measures:
- name: total_revenue
  expression: SUM(revenue)
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	// Invalid SQL syntax
	res, err := server.ResolveTemplatedString(testCtx(), &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Body:       `{{ metrics_sql "SELECT invalid syntax FROM mv" }}`,
	})
	require.Error(t, err)
	require.Nil(t, res)
	require.Contains(t, err.Error(), "selected column `invalid` not found")
}

func TestResolveTemplatedString_MetricsSQL_ErrorMetricsViewNotFound(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
		},
	})

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	res, err := server.ResolveTemplatedString(testCtx(), &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Body:       `{{ metrics_sql "SELECT total_revenue FROM nonexistent_mv" }}`,
	})
	require.Error(t, err)
	require.Nil(t, res)
}

func TestResolveTemplatedString_MetricsSQL_WithSecurity(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"model.sql": `
SELECT 'US' AS country, 100 AS revenue
UNION ALL
SELECT 'UK' AS country, 200 AS revenue
`,
			"mv.yaml": `
type: metrics_view
version: 1
model: model
dimensions:
- column: country
measures:
- name: total_revenue
  expression: SUM(revenue)
security:
  access: false
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	// Use a context with a user that doesn't have access
	ctx := auth.WithClaims(context.Background(), &runtime.SecurityClaims{
		SkipChecks: false,
		UserAttributes: map[string]any{
			"user": "restricted_user",
		},
	})

	res, err := server.ResolveTemplatedString(ctx, &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Body:       `{{ metrics_sql "SELECT total_revenue FROM mv" }}`,
	})
	require.Error(t, err)
	require.Nil(t, res)
	// Error occurs at the permission check level, not within the metrics_sql function
	require.Contains(t, err.Error(), "does not have access to query data")
}

func TestResolveTemplatedString_ErrorNoReadAPIPermission(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
		},
	})

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	// Context without ReadAPI permission
	ctx := auth.WithClaims(context.Background(), &runtime.SecurityClaims{
		SkipChecks: false,
	})

	res, err := server.ResolveTemplatedString(ctx, &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Body:       "Hello World",
	})
	require.Error(t, err)
	require.Nil(t, res)
	require.Contains(t, err.Error(), "does not have access to query data")
}

func TestResolveTemplatedString_InvalidTemplateError(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
		},
	})

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	res, err := server.ResolveTemplatedString(testCtx(), &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Body:       "Hello {{ .Invalid.Syntax",
	})
	require.Error(t, err)
	require.Nil(t, res)
	require.Contains(t, err.Error(), "failed to resolve template")
}

func TestResolveTemplatedString_EnvFunctionsDisabled(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
		},
	})

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	// env and expandenv should not be available
	res, err := server.ResolveTemplatedString(testCtx(), &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Body:       `{{ env "PATH" }}`,
	})
	require.Error(t, err)
	require.Nil(t, res)
	require.Contains(t, err.Error(), "function \"env\" not defined")
}

func TestResolveTemplatedString_ComplexMarkdown(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"model.sql": `
SELECT 'US' AS country, 100 AS revenue, 5 AS orders
UNION ALL
SELECT 'UK' AS country, 200 AS revenue, 10 AS orders
`,
			"mv.yaml": `
type: metrics_view
version: 1
model: model
dimensions:
- column: country
measures:
- name: total_revenue
  expression: SUM(revenue)
- name: total_orders
  expression: SUM(orders)
- name: avg_order_value
  expression: SUM(revenue) / SUM(orders)
`,
		},
		Variables: map[string]string{
			"company_name": "Acme Corp",
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	ctx := auth.WithClaims(context.Background(), &runtime.SecurityClaims{
		SkipChecks: true,
		UserAttributes: map[string]any{
			"name": "Jane Doe",
		},
	})

	markdown := `# Sales Report for {{ .env.company_name }}

Welcome, {{ .user.name }}!

## Key Metrics

- **Total Revenue**: {{ metrics_sql "SELECT total_revenue FROM mv" }}
- **Total Orders**: {{ metrics_sql "SELECT total_orders FROM mv" }}
- **Average Order Value**: {{ metrics_sql "SELECT avg_order_value FROM mv" }}

## Analysis

The total revenue is {{ metrics_sql "SELECT total_revenue FROM mv" }}, which is {{ if gt (metrics_sql "SELECT total_revenue FROM mv" | atoi) 250 }}above{{ else }}below{{ end }} target.
`

	res, err := server.ResolveTemplatedString(ctx, &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Body:       markdown,
	})
	require.NoError(t, err)
	require.Contains(t, res.Body, "# Sales Report for Acme Corp")
	require.Contains(t, res.Body, "Welcome, Jane Doe!")
	require.Contains(t, res.Body, "**Total Revenue**: 300")
	require.Contains(t, res.Body, "**Total Orders**: 15")
	require.Contains(t, res.Body, "**Average Order Value**: 20")
	require.Contains(t, res.Body, "above target")
}

func TestResolveTemplatedString_MetricsSQL_WithAdditionalTimeRange(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"model.sql": `
SELECT DATE '2024-01-01' AS order_date, 100 AS revenue
UNION ALL
SELECT DATE '2024-01-15' AS order_date, 200 AS revenue
UNION ALL
SELECT DATE '2024-02-01' AS order_date, 150 AS revenue
UNION ALL
SELECT DATE '2024-02-15' AS order_date, 250 AS revenue
`,
			"mv.yaml": `
type: metrics_view
version: 1
model: model
timeseries: order_date
dimensions:
- column: order_date
  name: order_date
measures:
- name: total_revenue
  expression: SUM(revenue)
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	// Test with time range filtering to January 2024
	startTime, err := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")
	require.NoError(t, err)
	endTime, err := time.Parse(time.RFC3339, "2024-02-01T00:00:00Z")
	require.NoError(t, err)

	additionalTimeRange := &runtimev1.TimeRange{
		Start: timestamppb.New(startTime),
		End:   timestamppb.New(endTime),
	}

	res, err := server.ResolveTemplatedString(testCtx(), &runtimev1.ResolveTemplatedStringRequest{
		InstanceId:          instanceID,
		Body:                `Revenue: {{ metrics_sql "SELECT total_revenue FROM mv" }}`,
		AdditionalTimeRange: additionalTimeRange,
	})
	require.NoError(t, err)
	// Should only include January
	require.Equal(t, "Revenue: 700", res.Body)
}
