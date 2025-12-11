package executor_test

import (
	"context"
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

// TestResolveQueryAttributesTemplate verifies template expressions
func TestResolveQueryAttributesTemplate(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		TestConnectors: []string{"clickhouse"},
		Files: map[string]string{
			"rill.yaml": "olap_connector: clickhouse",
			"model.sql": `
SELECT CURRENT_TIMESTAMP AS timestamp, 'publisher1' AS publisher
`,
			"metrics.yaml": `
type: metrics_view
model: model
timeseries: timestamp
dimensions:
  - column: publisher
measures:
  - expression: count(*)
query_attributes:
  test_compound_attr: '{{ .user.org_id }}_{{ .user.tenant_id }}'
  test_conditional_attr: '{{ if .user.is_premium }}premium{{ else }}standard{{ end }}'
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	mv := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindMetricsView, "metrics")
	spec := mv.GetMetricsView().Spec

	// Create user attributes
	userAttrs := map[string]any{
		"org_id":     "org123",
		"tenant_id":  "tenant456",
		"is_premium": true,
	}

	e, err := executor.New(context.Background(), rt, instanceID, spec, false, runtime.ResolvedSecurityOpen, 0, userAttrs)
	require.NoError(t, err)
	defer e.Close()

	// Verify cache key is computed successfully with complex templates
	cacheKey, _, err := e.CacheKey(context.Background())
	require.NoError(t, err)
	require.NotNil(t, cacheKey)
}
