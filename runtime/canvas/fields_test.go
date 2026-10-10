package canvas

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
)

func TestFieldTemplateData(t *testing.T) {
	mv := &runtimev1.MetricsViewSpec{
		TimeDimension: "timestamp",
		Measures: []*runtimev1.MetricsViewSpec_Measure{
			{Name: "bid price sum", DisplayName: "Sum of Bid Price", FormatPreset: "currency_usd"},
			{Name: "total_records", DisplayName: "Total records", FormatD3: ",.0f"},
		},
		Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
			{Name: "publisher", DisplayName: "Publisher"},
			{Name: "created_at", DisplayName: "Created At", Type: runtimev1.MetricsViewSpec_DIMENSION_TYPE_TIME},
		},
	}
	metricsViews := map[string]*runtimev1.MetricsViewSpec{"mv1": mv}

	params := []*runtimev1.ComponentParam{
		{Name: "metrics_view", Type: "metrics_view"},
		{Name: "measure", Type: "measure", MetricsViewParam: "metrics_view"},
		{Name: "dim", Type: "dimension", MetricsViewParam: "metrics_view"},
		{Name: "time_dim", Type: "time_dimension", MetricsViewParam: "metrics_view"},
		{Name: "limit", Type: "number"},
	}

	fields := FieldTemplateData(params, map[string]any{
		"metrics_view": "mv1",
		"measure":      "bid price sum",
		"dim":          "publisher",
		"time_dim":     "timestamp",
		"limit":        100,
	}, metricsViews)

	require.Equal(t, map[string]any{
		"name":          "bid price sum",
		"display_name":  "Sum of Bid Price",
		"type":          "measure",
		"format_d3":     "",
		"format_preset": "currency_usd",
		// The measure name is reduced to a Vega identifier: " " is U+0020.
		"format_type": "rill_bid_u20_price_u20_sum",
	}, fields["measure"])

	require.Equal(t, "Publisher", fields["dim"].(map[string]any)["display_name"])
	// The primary time dimension is labelled "Time" rather than by its column name.
	require.Equal(t, "Time", fields["time_dim"].(map[string]any)["display_name"])
	// Only field-typed params carry metrics view metadata.
	require.NotContains(t, fields, "limit")
	require.NotContains(t, fields, "metrics_view")

	t.Run("secondary time dimension keeps its display name", func(t *testing.T) {
		fields := FieldTemplateData(params, map[string]any{
			"metrics_view": "mv1",
			"time_dim":     "created_at",
		}, metricsViews)
		require.Equal(t, "Created At", fields["time_dim"].(map[string]any)["display_name"])
	})

	t.Run("unresolvable bindings blank out rather than going missing", func(t *testing.T) {
		// A reference to a missing key resolves to Go's "<no value>" placeholder, which would then
		// reach the renderer as a field or formatter name, so every field-typed param gets an entry.
		fields := FieldTemplateData(params, map[string]any{
			// No metrics view bound, so no field metadata can be resolved.
			"measure": "total_records",
			// Templated and unbound values resolve at render time, not here.
			"dim": "{{ .env.dim }}",
		}, metricsViews)

		require.Equal(t, map[string]any{
			"name":          "total_records",
			"display_name":  "total_records",
			"type":          "measure",
			"format_d3":     "",
			"format_preset": "",
			"format_type":   "",
		}, fields["measure"])
		require.Equal(t, map[string]any{
			"name":          "",
			"display_name":  "",
			"type":          "dimension",
			"format_d3":     "",
			"format_preset": "",
			"format_type":   "",
		}, fields["dim"])
		// Unbound entirely, and still present.
		require.Contains(t, fields, "time_dim")

		for name, meta := range fields {
			for _, key := range FieldMetadataKeys {
				require.Contains(t, meta, key, "param %q is missing %q", name, key)
			}
		}
	})

	t.Run("unknown field falls back to its name", func(t *testing.T) {
		fields := FieldTemplateData(params, map[string]any{
			"metrics_view": "mv1",
			"dim":          "not_a_dimension",
		}, metricsViews)
		require.Equal(t, "not_a_dimension", fields["dim"].(map[string]any)["display_name"])
	})
}

func TestVegaFormatType(t *testing.T) {
	require.Equal(t, "rill_total_records", VegaFormatType("total_records"))
	require.Equal(t, "rill_bid_u20_price", VegaFormatType("bid price"))
	require.Equal(t, "rill_a_u25_b", VegaFormatType("a%b"))
	require.Equal(t, "rill_field", VegaFormatType(""))
}
