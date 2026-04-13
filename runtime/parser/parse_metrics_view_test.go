package parser

import (
	"context"
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestMetricsViewDimensionLookup(t *testing.T) {
	files := map[string]string{
		// rill.yaml
		`rill.yaml`: ``,
		// model m1
		`models/m1.sql`: `SELECT 1 AS id`,
		// model m2
		`models/m2.sql`: `SELECT 1 AS id, 2 AS value`,
		// metrics view
		`metrics_views/mv1.yaml`: `
type: metrics_view
version: 1
model: m1
dimensions:
- name: foo
  expression: id
  lookup_table: default.m2 # Expect ref to m2 after stripping the schema prefix since it is a model in the same project
  lookup_key_column: id
  lookup_value_column: value
- name: bar
  expression: id
  lookup_table: ext # Expect no ref since it is not a known model (assumed to be a pre-existing table)
  lookup_key_column: id
  lookup_value_column: value
measures:
- name: count
  expression: COUNT(*)
`,
	}

	resources := []*Resource{
		// model m1
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
			Paths: []string{"/models/m1.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
				InputConnector:  "duckdb",
				InputProperties: must(structpb.NewStruct(map[string]any{"sql": strings.TrimSpace(files["models/m1.sql"])})),
				OutputConnector: "duckdb",
				ChangeMode:      runtimev1.ModelChangeMode_MODEL_CHANGE_MODE_RESET,
			},
		},
		// model m2
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m2"},
			Paths: []string{"/models/m2.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
				InputConnector:  "duckdb",
				InputProperties: must(structpb.NewStruct(map[string]any{"sql": strings.TrimSpace(files["models/m2.sql"])})),
				OutputConnector: "duckdb",
				ChangeMode:      runtimev1.ModelChangeMode_MODEL_CHANGE_MODE_RESET,
			},
		},
		// metrics view
		{
			Name: ResourceName{Kind: ResourceKindMetricsView, Name: "mv1"},
			// Note: Expecting a ref to m2 since it's used as a lookup table and exists as a model in the same project.
			Refs:  []ResourceName{{Kind: ResourceKindModel, Name: "m1"}, {Kind: ResourceKindModel, Name: "m2"}},
			Paths: []string{"/metrics_views/mv1.yaml"},
			MetricsViewSpec: &runtimev1.MetricsViewSpec{
				Connector:   "duckdb",
				Model:       "m1",
				DisplayName: "Mv1",
				Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
					{
						Name:              "foo",
						DisplayName:       "Foo",
						Expression:        "id",
						LookupTable:       "default.m2",
						LookupKeyColumn:   "id",
						LookupValueColumn: "value",
					},
					{
						Name:              "bar",
						DisplayName:       "Bar",
						Expression:        "id",
						LookupTable:       "ext",
						LookupKeyColumn:   "id",
						LookupValueColumn: "value",
					},
				},
				Measures: []*runtimev1.MetricsViewSpec_Measure{
					{
						Name:        "count",
						DisplayName: "Count",
						Expression:  "COUNT(*)",
						Type:        runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE,
					},
				},
			},
		},
	}

	ctx := context.Background()
	repo := makeRepo(t, files)
	p, err := Parse(ctx, repo, "", "", "duckdb", true)
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}

func TestMetricsViewDimensionSmallestTimeGrain(t *testing.T) {
	files := map[string]string{
		// rill.yaml
		`rill.yaml`: ``,
		// model m1
		`models/m1.sql`: `SELECT 1 AS id, '2025-01-01T00:00:00Z'::TIMESTAMP AS ts1, '2025-01-01'::DATE AS ts2`,
		// metrics view
		`metrics_views/mv1.yaml`: `
type: metrics_view
version: 1
model: m1
timeseries: ts1
smallest_time_grain: hour
dimensions:
- column: id
- column: ts2
  smallest_time_grain: day
measures:
- name: count
  expression: COUNT(*)
`,
	}

	resources := []*Resource{
		// model m1
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
			Paths: []string{"/models/m1.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
				InputConnector:  "duckdb",
				InputProperties: must(structpb.NewStruct(map[string]any{"sql": strings.TrimSpace(files["models/m1.sql"])})),
				OutputConnector: "duckdb",
				ChangeMode:      runtimev1.ModelChangeMode_MODEL_CHANGE_MODE_RESET,
			},
		},
		// metrics view
		{
			Name: ResourceName{Kind: ResourceKindMetricsView, Name: "mv1"},
			// Note: Expecting a ref to m2 since it's used as a lookup table and exists as a model in the same project.
			Refs:  []ResourceName{{Kind: ResourceKindModel, Name: "m1"}},
			Paths: []string{"/metrics_views/mv1.yaml"},
			MetricsViewSpec: &runtimev1.MetricsViewSpec{
				Connector:         "duckdb",
				Model:             "m1",
				DisplayName:       "Mv1",
				TimeDimension:     "ts1",
				SmallestTimeGrain: runtimev1.TimeGrain_TIME_GRAIN_HOUR,
				Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
					{
						Name:        "ts1",
						DisplayName: "Ts1",
						Column:      "ts1",
					},
					{
						Name:        "id",
						DisplayName: "Id",
						Column:      "id",
					},
					{
						Name:              "ts2",
						DisplayName:       "Ts2",
						Column:            "ts2",
						SmallestTimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
					},
				},
				Measures: []*runtimev1.MetricsViewSpec_Measure{
					{
						Name:        "count",
						DisplayName: "Count",
						Expression:  "COUNT(*)",
						Type:        runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE,
					},
				},
			},
		},
	}

	ctx := context.Background()
	repo := makeRepo(t, files)
	p, err := Parse(ctx, repo, "", "", "duckdb", true)
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}

func TestMetricsViewTags(t *testing.T) {
	files := map[string]string{
		// rill.yaml
		`rill.yaml`: ``,
		// model m1
		`models/m1.sql`: `SELECT 1 AS id, 'test' AS category`,
		// metrics view with tags
		`metrics_views/mv1.yaml`: `
type: metrics_view
version: 1
model: m1
dimensions:
- name: dim_with_tags
  column: id
  tags:
    - dimension
    - primary
    - test
- name: dim_without_tags
  column: category
- name: dim_with_empty_tags
  column: id
  tags: []
measures:
- name: measure_with_tags
  expression: COUNT(*)
  tags:
    - measure
    - count
    - test
- name: measure_without_tags
  expression: SUM(id)
- name: measure_with_empty_tags
  expression: AVG(id)
  tags: []
`,
	}

	resources := []*Resource{
		// model m1
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
			Paths: []string{"/models/m1.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
				InputConnector:  "duckdb",
				InputProperties: must(structpb.NewStruct(map[string]any{"sql": strings.TrimSpace(files["models/m1.sql"])})),
				OutputConnector: "duckdb",
				ChangeMode:      runtimev1.ModelChangeMode_MODEL_CHANGE_MODE_RESET,
			},
		},
		// metrics view
		{
			Name:  ResourceName{Kind: ResourceKindMetricsView, Name: "mv1"},
			Refs:  []ResourceName{{Kind: ResourceKindModel, Name: "m1"}},
			Paths: []string{"/metrics_views/mv1.yaml"},
			MetricsViewSpec: &runtimev1.MetricsViewSpec{
				Connector:   "duckdb",
				Model:       "m1",
				DisplayName: "Mv1",
				Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
					{
						Name:        "dim_with_tags",
						DisplayName: "Dim With Tags",
						Column:      "id",
						Tags:        []string{"dimension", "primary", "test"},
					},
					{
						Name:        "dim_without_tags",
						DisplayName: "Dim Without Tags",
						Column:      "category",
					},
					{
						Name:        "dim_with_empty_tags",
						DisplayName: "Dim With Empty Tags",
						Column:      "id",
						Tags:        []string{},
					},
				},
				Measures: []*runtimev1.MetricsViewSpec_Measure{
					{
						Name:        "measure_with_tags",
						DisplayName: "Measure With Tags",
						Expression:  "COUNT(*)",
						Type:        runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE,
						Tags:        []string{"measure", "count", "test"},
					},
					{
						Name:        "measure_without_tags",
						DisplayName: "Measure Without Tags",
						Expression:  "SUM(id)",
						Type:        runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE,
					},
					{
						Name:        "measure_with_empty_tags",
						DisplayName: "Measure With Empty Tags",
						Expression:  "AVG(id)",
						Type:        runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE,
						Tags:        []string{},
					},
				},
			},
		},
	}

	ctx := context.Background()
	repo := makeRepo(t, files)
	p, err := Parse(ctx, repo, "", "", "duckdb", true)
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)

	mv := p.Resources[ResourceName{Kind: ResourceKindMetricsView, Name: "mv1"}]
	require.NotNil(t, mv)
	require.NotNil(t, mv.MetricsViewSpec)

	require.Len(t, mv.MetricsViewSpec.Dimensions, 3)
	require.Equal(t, []string{"dimension", "primary", "test"}, mv.MetricsViewSpec.Dimensions[0].Tags)
	require.Nil(t, mv.MetricsViewSpec.Dimensions[1].Tags)
	require.Equal(t, []string{}, mv.MetricsViewSpec.Dimensions[2].Tags)

	require.Len(t, mv.MetricsViewSpec.Measures, 3)
	require.Equal(t, []string{"measure", "count", "test"}, mv.MetricsViewSpec.Measures[0].Tags)
	require.Nil(t, mv.MetricsViewSpec.Measures[1].Tags)
	require.Equal(t, []string{}, mv.MetricsViewSpec.Measures[2].Tags)
}

func TestValidateQueryAttributes(t *testing.T) {
	tests := []struct {
		name    string
		attrs   map[string]string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid simple attributes",
			attrs:   map[string]string{"partner_id": "acme_corp", "region": "us-west"},
			wantErr: false,
		},
		{
			name:    "valid with underscores and hyphens",
			attrs:   map[string]string{"partner_id": "value1", "user-role": "admin", "app.env": "prod"},
			wantErr: false,
		},
		{
			name:    "valid with dots in key",
			attrs:   map[string]string{"app.environment": "production"},
			wantErr: false,
		},
		{
			name:    "valid with template",
			attrs:   map[string]string{"partner_id": "{{ .user.partner_id }}"},
			wantErr: false,
		},
		{
			name:    "empty attributes map",
			attrs:   map[string]string{},
			wantErr: false,
		},
		{
			name:    "nil attributes map",
			attrs:   nil,
			wantErr: false,
		},
		{
			name:    "empty key",
			attrs:   map[string]string{"": "value"},
			wantErr: true,
		},
		{
			name:    "invalid key with spaces",
			attrs:   map[string]string{"partner id": "value"},
			wantErr: true,
			errMsg:  "contains invalid characters",
		},
		{
			name:    "invalid key with special chars",
			attrs:   map[string]string{"partner@id": "value"},
			wantErr: true,
			errMsg:  "contains invalid characters",
		},
		{
			name:    "invalid key with SQL injection",
			attrs:   map[string]string{"partner'; DROP TABLE users--": "value"},
			wantErr: true,
			errMsg:  "contains invalid characters",
		},
		{
			name:    "template with dangerous pattern should pass",
			attrs:   map[string]string{"query": "{{ .user.custom_query }}"},
			wantErr: false,
		},
		{
			name:    "mixed safe and template values",
			attrs:   map[string]string{"env": "production", "partner_id": "{{ .user.partner_id }}"},
			wantErr: false,
		},
		{
			name:    "valid uppercase key",
			attrs:   map[string]string{"PARTNER_ID": "value"},
			wantErr: false,
		},
		{
			name:    "valid numeric in key",
			attrs:   map[string]string{"partner_id_123": "value"},
			wantErr: false,
		},
		{
			name:    "simple alphanumeric",
			attrs:   map[string]string{"partner_id": ""},
			wantErr: false,
		},
		{
			name:    "with hyphen",
			attrs:   map[string]string{"partner-id": ""},
			wantErr: false,
		},
		{
			name:    "with dot",
			attrs:   map[string]string{"app.environment": ""},
			wantErr: false,
		},
		{
			name:    "with numbers",
			attrs:   map[string]string{"key123": ""},
			wantErr: false,
		},
		{
			name:    "uppercase",
			attrs:   map[string]string{"PARTNER_ID": ""},
			wantErr: false,
		},
		{
			name:    "mixed case",
			attrs:   map[string]string{"PartnerId": ""},
			wantErr: false,
		},
		{
			name:    "empty string",
			attrs:   map[string]string{"": ""},
			wantErr: true,
		},
		{
			name:    "with space",
			attrs:   map[string]string{"partner id": ""},
			wantErr: true,
		},
		{
			name:    "with special char",
			attrs:   map[string]string{"partner@id": ""},
			wantErr: true,
		},
		{
			name:    "with slash",
			attrs:   map[string]string{"partner/id": ""},
			wantErr: true,
		},
		{
			name:    "with quotes",
			attrs:   map[string]string{"partner'id": ""},
			wantErr: true,
		},
		{
			name:    "with semicolon",
			attrs:   map[string]string{"partner;id": ""},
			wantErr: true,
		},
		{
			name:    "unicode",
			attrs:   map[string]string{"партнер": ""},
			wantErr: true,
		},
		{
			name:    "emoji",
			attrs:   map[string]string{"partner🎉": ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateQueryAttributes(tt.attrs)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMetricsViewRollups(t *testing.T) {
	files := map[string]string{
		`rill.yaml`:               ``,
		`models/m1.sql`:           `SELECT 1 AS id, 'a' AS publisher, 'b' AS domain`,
		`models/rollup_daily.sql`: `SELECT 1 AS id`,
		`metrics_views/mv1.yaml`: `
type: metrics_view
version: 1
model: m1
dimensions:
- name: publisher
  column: publisher
- name: domain
  column: domain
measures:
- name: total_impressions
  expression: "SUM(impressions)"
- name: total_clicks
  expression: "SUM(clicks)"
rollups:
  - model: rollup_daily
    time_grain: day
    dimensions:
      - publisher
      - domain
    measures:
      - total_impressions
      - total_clicks
`,
	}

	ctx := context.Background()
	repo := makeRepo(t, files)
	p, err := Parse(ctx, repo, "", "", "duckdb", true)
	require.NoError(t, err)
	require.Empty(t, p.Errors)

	// Find the metrics view resource
	var mvSpec *runtimev1.MetricsViewSpec
	for _, r := range p.Resources {
		if r.Name.Kind == ResourceKindMetricsView && r.Name.Name == "mv1" {
			mvSpec = r.MetricsViewSpec
			break
		}
	}
	require.NotNil(t, mvSpec)
	require.Len(t, mvSpec.Rollups, 1)

	rollup := mvSpec.Rollups[0]
	require.Equal(t, "rollup_daily", rollup.Model)
	require.Equal(t, runtimev1.TimeGrain_TIME_GRAIN_DAY, rollup.TimeGrain)
	require.Equal(t, []string{"publisher", "domain"}, rollup.Dimensions)
	require.Equal(t, []string{"total_impressions", "total_clicks"}, rollup.Measures)
	require.Nil(t, rollup.DimensionsSelector)
	require.Nil(t, rollup.MeasuresSelector)
}

func TestMetricsViewRollupsStarSelector(t *testing.T) {
	files := map[string]string{
		`rill.yaml`:               ``,
		`models/m1.sql`:           `SELECT 1 AS id, 'a' AS publisher`,
		`models/rollup_daily.sql`: `SELECT 1 AS id`,
		`metrics_views/mv1.yaml`: `
type: metrics_view
version: 1
model: m1
dimensions:
- name: publisher
  column: publisher
measures:
- name: total_impressions
  expression: "SUM(impressions)"
rollups:
  - model: rollup_daily
    time_grain: day
    dimensions: "*"
    measures: "*"
`,
	}

	ctx := context.Background()
	repo := makeRepo(t, files)
	p, err := Parse(ctx, repo, "", "", "duckdb", true)
	require.NoError(t, err)
	require.Empty(t, p.Errors)

	var mvSpec *runtimev1.MetricsViewSpec
	for _, r := range p.Resources {
		if r.Name.Kind == ResourceKindMetricsView && r.Name.Name == "mv1" {
			mvSpec = r.MetricsViewSpec
			break
		}
	}
	require.NotNil(t, mvSpec)
	require.Len(t, mvSpec.Rollups, 1)

	rollup := mvSpec.Rollups[0]
	// '*' cannot be resolved at parse time; should produce selectors
	require.Empty(t, rollup.Dimensions)
	require.Empty(t, rollup.Measures)
	require.NotNil(t, rollup.DimensionsSelector)
	require.NotNil(t, rollup.MeasuresSelector)
}

func TestMetricsViewRollupsExcludeSelector(t *testing.T) {
	files := map[string]string{
		`rill.yaml`:               ``,
		`models/m1.sql`:           `SELECT 1 AS id, 'a' AS publisher`,
		`models/rollup_daily.sql`: `SELECT 1 AS id`,
		`metrics_views/mv1.yaml`: `
type: metrics_view
version: 1
model: m1
dimensions:
- name: publisher
  column: publisher
measures:
- name: total_impressions
  expression: "SUM(impressions)"
- name: total_clicks
  expression: "SUM(clicks)"
rollups:
  - model: rollup_daily
    time_grain: day
    dimensions:
      - publisher
    measures:
      exclude:
        - total_clicks
`,
	}

	ctx := context.Background()
	repo := makeRepo(t, files)
	p, err := Parse(ctx, repo, "", "", "duckdb", true)
	require.NoError(t, err)
	require.Empty(t, p.Errors)

	var mvSpec *runtimev1.MetricsViewSpec
	for _, r := range p.Resources {
		if r.Name.Kind == ResourceKindMetricsView && r.Name.Name == "mv1" {
			mvSpec = r.MetricsViewSpec
			break
		}
	}
	require.NotNil(t, mvSpec)
	require.Len(t, mvSpec.Rollups, 1)

	rollup := mvSpec.Rollups[0]
	require.Equal(t, []string{"publisher"}, rollup.Dimensions)
	// Exclude selector cannot be resolved at parse time
	require.Empty(t, rollup.Measures)
	require.NotNil(t, rollup.MeasuresSelector)
}

func TestMetricsViewRollupsRequiredTimeGrain(t *testing.T) {
	files := map[string]string{
		`rill.yaml`:               ``,
		`models/m1.sql`:           `SELECT 1 AS id, 'a' AS publisher`,
		`models/rollup_daily.sql`: `SELECT 1 AS id`,
		`metrics_views/mv1.yaml`: `
type: metrics_view
version: 1
model: m1
dimensions:
- name: publisher
  column: publisher
measures:
- name: total_impressions
  expression: "SUM(impressions)"
rollups:
  - model: rollup_daily
    dimensions:
      - publisher
    measures:
      - total_impressions
`,
	}

	ctx := context.Background()
	repo := makeRepo(t, files)
	p, err := Parse(ctx, repo, "", "", "duckdb", true)
	require.NoError(t, err)
	require.NotEmpty(t, p.Errors)
	require.Contains(t, p.Errors[0].Message, `"time_grain" is required`)
}

func TestMetricsViewRollupsValidation(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name: "missing model",
			yaml: `
type: metrics_view
version: 1
model: m1
dimensions:
- name: publisher
  column: publisher
measures:
- name: count
  expression: "COUNT(*)"
rollups:
  - time_grain: day
    measures:
      - count
`,
			wantErr: `"model" is required`,
		},
		{
			name: "invalid time_grain",
			yaml: `
type: metrics_view
version: 1
model: m1
dimensions:
- name: publisher
  column: publisher
measures:
- name: count
  expression: "COUNT(*)"
rollups:
  - model: r1
    time_grain: fortnight
    measures:
      - count
`,
			wantErr: `invalid "time_grain"`,
		},
		{
			name: "dimension not in metrics view",
			yaml: `
type: metrics_view
version: 1
model: m1
dimensions:
- name: publisher
  column: publisher
measures:
- name: count
  expression: "COUNT(*)"
rollups:
  - model: r1
    time_grain: day
    dimensions:
      - nonexistent
    measures:
      - count
`,
			wantErr: `dimension "nonexistent" does not exist`,
		},
		{
			name: "measure not in metrics view",
			yaml: `
type: metrics_view
version: 1
model: m1
dimensions:
- name: publisher
  column: publisher
measures:
- name: count
  expression: "COUNT(*)"
rollups:
  - model: r1
    time_grain: day
    measures:
      - nonexistent
`,
			wantErr: `measure "nonexistent" does not exist`,
		},
		{
			name: "invalid timezone",
			yaml: `
type: metrics_view
version: 1
model: m1
dimensions:
- name: publisher
  column: publisher
measures:
- name: count
  expression: "COUNT(*)"
rollups:
  - model: r1
    time_grain: day
    time_zone: Not/A_Timezone
    measures:
      - count
`,
			wantErr: `invalid "time_zone"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := map[string]string{
				`rill.yaml`:              ``,
				`models/m1.sql`:          `SELECT 1 AS id`,
				`metrics_views/mv1.yaml`: tt.yaml,
			}
			ctx := context.Background()
			repo := makeRepo(t, files)
			p, err := Parse(ctx, repo, "", "", "duckdb", true)
			require.NoError(t, err)
			require.NotEmpty(t, p.Errors)
			require.Contains(t, p.Errors[0].Message, tt.wantErr)
		})
	}
}
