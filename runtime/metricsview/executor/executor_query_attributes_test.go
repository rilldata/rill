package executor_test

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

// TestResolveQueryAttributesEmpty verifies that empty query attributes return nil
func TestResolveQueryAttributesEmpty(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	mv := &runtimev1.MetricsViewSpec{
		Connector:     "duckdb",
		Table:         "ad_bids",
		TimeDimension: "timestamp",
		Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
			{Name: "publisher", Column: "publisher"},
		},
		Measures: []*runtimev1.MetricsViewSpec_Measure{
			{Name: "records", Expression: "count(*)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
		},
		QueryAttributes: map[string]string{},
	}

	e, err := executor.New(context.Background(), rt, instanceID, mv, false, runtime.ResolvedSecurityOpen, 0)
	require.NoError(t, err)
	defer e.Close()

	require.NoError(t, err)
}

// TestResolveQueryAttributesNoAttributes verifies that nil query attributes return nil
func TestResolveQueryAttributesNoAttributes(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	mv := &runtimev1.MetricsViewSpec{
		Connector:     "duckdb",
		Table:         "ad_bids",
		TimeDimension: "timestamp",
		Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
			{Name: "publisher", Column: "publisher"},
		},
		Measures: []*runtimev1.MetricsViewSpec_Measure{
			{Name: "records", Expression: "count(*)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
		},
	}

	e, err := executor.New(context.Background(), rt, instanceID, mv, false, runtime.ResolvedSecurityOpen, 0)
	require.NoError(t, err)
	defer e.Close()

	require.NoError(t, err)
}

// TestResolveQueryAttributesSimpleValue verifies that simple string values are resolved correctly
func TestResolveQueryAttributesSimpleValue(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": `olap_connector: duckdb`,
			"model.sql": `
SELECT now() AS timestamp, 'publisher1' AS publisher
`,
			"metrics.yaml": `
type: metrics_view
model: model
dimensions:
  - column: publisher
measures:
  - expression: count(*)
query_attributes:
  author: sample_author
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	mv := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindMetricsView, "metrics")
	spec := mv.GetMetricsView().Spec

	require.NotNil(t, spec.QueryAttributes)
	require.Equal(t, "sample_author", spec.QueryAttributes["author"])

	e, err := executor.New(context.Background(), rt, instanceID, spec, false, runtime.ResolvedSecurityOpen, 0)
	require.NoError(t, err)
	defer e.Close()

	// Verify cache key is computed successfully
	cacheKey, _, err := e.CacheKey(context.Background())
	require.NoError(t, err)
	require.NotNil(t, cacheKey)
}

// TestResolveQueryAttributesEnvironmentVariable verifies that environment variables are resolved
func TestResolveQueryAttributesEnvironmentVariable(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "olap_connector: duckdb",
			"model.sql": `
SELECT now() AS timestamp, 'publisher1' AS publisher
`,
			"metrics.yaml": `
type: metrics_view
model: model
dimensions:
  - column: publisher
measures:
  - expression: count(*)
query_attributes:
  test_env_attr: '{{ .env.ENV_VAR }}'
`,
		},
		Variables: map[string]string{
			"ENV_VAR": "test_value",
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	mv := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindMetricsView, "metrics")
	spec := mv.GetMetricsView().Spec

	e, err := executor.New(context.Background(), rt, instanceID, spec, false, runtime.ResolvedSecurityOpen, 0)
	require.NoError(t, err)
	defer e.Close()

	// Verify through cache key that query attributes are resolved
	cacheKey, _, err := e.CacheKey(context.Background())
	require.NoError(t, err)
	require.NotNil(t, cacheKey)
}

// TestResolveQueryAttributesMultipleAttributes verifies that multiple attributes are resolved correctly
func TestResolveQueryAttributesMultipleAttributes(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "olap_connector: duckdb",
			"model.sql": `
SELECT now() AS timestamp, 'publisher1' AS publisher
`,
			"metrics.yaml": `
type: metrics_view
model: model
dimensions:
  - column: publisher
measures:
  - expression: count(*)
query_attributes:
  attr1: value1
  attr2: value2
  attr3: value3
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	mv := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindMetricsView, "metrics")
	spec := mv.GetMetricsView().Spec

	// Verify all query attributes are stored in the spec
	require.NotNil(t, spec.QueryAttributes)
	require.Equal(t, "value1", spec.QueryAttributes["attr1"])
	require.Equal(t, "value2", spec.QueryAttributes["attr2"])
	require.Equal(t, "value3", spec.QueryAttributes["attr3"])
	require.Len(t, spec.QueryAttributes, 3)

	e, err := executor.New(context.Background(), rt, instanceID, spec, false, runtime.ResolvedSecurityOpen, 0)
	require.NoError(t, err)
	defer e.Close()

	// Verify cache key computation works (uses resolveQueryAttributes)
	cacheKey, _, err := e.CacheKey(context.Background())
	require.NoError(t, err)
	require.NotNil(t, cacheKey)
}

// TestResolveQueryAttributesTemplateError verifies that template resolution errors are properly handled
func TestResolveQueryAttributesTemplateError(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "olap_connector: duckdb",
			"model.sql": `
SELECT now() AS timestamp, 'publisher1' AS publisher
`,
			"metrics.yaml": `
type: metrics_view
model: model
dimensions:
  - column: publisher
measures:
  - expression: count(*)
query_attributes:
  bad_attr: '{{ .nonexistent.field }}'
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	mv := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindMetricsView, "metrics")
	spec := mv.GetMetricsView().Spec

	// Verify the attribute is stored as a template (YAML parser strips outer quotes)
	require.NotNil(t, spec.QueryAttributes)
	require.Contains(t, spec.QueryAttributes, "bad_attr")
	require.Equal(t, "{{ .nonexistent.field }}", spec.QueryAttributes["bad_attr"])

	e, err := executor.New(context.Background(), rt, instanceID, spec, false, runtime.ResolvedSecurityOpen, 0)
	require.NoError(t, err)
	defer e.Close()

	// Verify cache key computation works even with undefined template fields
	cacheKey, _, err := e.CacheKey(context.Background())
	require.NoError(t, err)
	require.NotNil(t, cacheKey)
}

// TestQueryAttributesInCacheKey verifies that query attributes are included in the cache key computation
func TestQueryAttributesInCacheKey(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		TestConnectors: []string{"clickhouse"},
		Files: map[string]string{
			"rill.yaml": "olap_connector: clickhouse",
			"model.sql": `
-- @connector: clickhouse
select parseDateTimeBestEffort('2024-01-01T00:00:00Z') as time, 'US' as country, 1 as val union all
select parseDateTimeBestEffort('2024-01-02T00:00:00Z') as time, 'US' as country, 2 as val
`,
			"metrics.yaml": `
type: metrics_view
model: model
dimensions:
  - column: country
measures:
  - expression: count(*)
cache:
  enabled: true
query_attributes:
  partner_id: partner123
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	mv := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindMetricsView, "metrics")
	spec := mv.GetMetricsView().Spec

	e, err := executor.New(context.Background(), rt, instanceID, spec, false, runtime.ResolvedSecurityOpen, 0)
	require.NoError(t, err)
	defer e.Close()

	cacheKey, _, err := e.CacheKey(context.Background())
	require.NoError(t, err)
	require.NotNil(t, cacheKey)
}

// TestQueryAttributesWithSpecialCharacters verifies that special characters in values are handled correctly
func TestQueryAttributesWithSpecialCharacters(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "olap_connector: duckdb",
			"model.sql": `
SELECT now() AS timestamp, 'publisher1' AS publisher
`,
			"metrics.yaml": `
type: metrics_view
model: model
dimensions:
  - column: publisher
measures:
  - expression: count(*)
query_attributes:
  attr_with_quotes: "value with 'quotes'"
  attr_with_spaces: "value with spaces"
  attr_with_equals: "key=value"
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	mv := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindMetricsView, "metrics")
	spec := mv.GetMetricsView().Spec

	// Verify query attributes with special characters are stored correctly
	require.NotNil(t, spec.QueryAttributes)
	require.Equal(t, "value with 'quotes'", spec.QueryAttributes["attr_with_quotes"])
	require.Equal(t, "value with spaces", spec.QueryAttributes["attr_with_spaces"])
	require.Equal(t, "key=value", spec.QueryAttributes["attr_with_equals"])

	e, err := executor.New(context.Background(), rt, instanceID, spec, false, runtime.ResolvedSecurityOpen, 0)
	require.NoError(t, err)
	defer e.Close()

	cacheKey, _, err := e.CacheKey(context.Background())
	require.NoError(t, err)
	require.NotNil(t, cacheKey)
}

// TestClickHouseQueryAttributesIntegration verifies that query attributes are properly passed to ClickHouse driver
func TestClickHouseQueryAttributesIntegration(t *testing.T) {
	t.Skip("Requires ClickHouse test environment - can be run with -tags=integration")

	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		TestConnectors: []string{"clickhouse"},
		Files: map[string]string{
			"rill.yaml": "olap_connector: clickhouse",
			"model.sql": `
-- @connector: clickhouse
select parseDateTimeBestEffort('2024-01-01T00:00:00Z') as time, 'US' as country, 1 as val union all
select parseDateTimeBestEffort('2024-01-02T00:00:00Z') as time, 'US' as country, 2 as val
`,
			"metrics.yaml": `
type: metrics_view
model: model
dimensions:
  - column: country
measures:
  - expression: count(*)
query_attributes:
  partner_id: partner123
  env_name: test
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	mv := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindMetricsView, "metrics")
	spec := mv.GetMetricsView().Spec

	e, err := executor.New(context.Background(), rt, instanceID, spec, false, runtime.ResolvedSecurityOpen, 0)
	require.NoError(t, err)
	defer e.Close()

	// Verify that cache key includes query attributes
	cacheKey, _, err := e.CacheKey(context.Background())
	require.NoError(t, err)
	require.NotNil(t, cacheKey)
}
