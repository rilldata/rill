package canvas

import (
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/pathutil"
)

// ValidateRendererProperties validates the renderer properties for a component.
// The provided metricsViews should contain every valid metrics view referenced by the component (as determined in the parser).
// If the renderer properties reference a metrics view not in metricsViews, assume the metrics view is invalid or does not exist (don't look it up separately in the catalog).
//
// Note: that metrics views referenced through markdown or metrics_sql cannot be validated here.
// This is because the upstream parser can't extract refs from them, so the metrics views cannot be passed through to here.
// Warning: if you try to fix this, note that the refs must be added in the parser, not looked up dynamically here;
// a dynamic lookup will have a race condition where the metrics view may not have been reconciled yet.
func ValidateRendererProperties(renderer string, props map[string]any, metricsViews map[string]*runtimev1.MetricsViewSpec) error {
	switch renderer {
	case "line_chart":
		mvn, ok := pathutil.GetPathString(props, "metrics_view")
		if !ok {
			return errors.New("renderer properties must include a string 'metrics_view' property")
		}
		mv := metricsViews[mvn]
		if mv == nil {
			return fmt.Errorf("referenced metrics view %q is invalid", mvn)
		}

		xField, ok := pathutil.GetPathString(props, "x.field")
		if !ok {
			return errors.New("renderer properties for line_chart must include a string 'x.field' property")
		}
		if !metricsViewHasDimension(mv, xField) {
			return fmt.Errorf("referenced x.field %q is not a dimension in metrics view %q", xField, mvn)
		}

		yField, ok := pathutil.GetPathString(props, "y.field")
		if !ok {
			return errors.New("renderer properties for line_chart must include a string 'y.field' property")
		}
		if !metricsViewHasMeasure(mv, yField) {
			return fmt.Errorf("referenced y.field %q is not a measure in metrics view %q", yField, mvn)
		}

		// TODO: Any other validation for line charts?
	case "stacked_bar":
		// TODO: Implement
	case "bar_chart":
		// TODO: Implement
	case "stacked_bar_normalized":
		// TODO: Implement
	case "area_chart":
		// TODO: Implement
	case "donut_chart":
		// TODO: Implement
	case "pie_chart":
		// TODO: Implement
	case "heatmap":
		// TODO: Implement
	case "funnel_chart":
		// TODO: Implement
	case "combo_chart":
		// TODO: Implement
	case "scatter_plot":
		// TODO: Implement
	case "markdown":
		// TODO: Implement
	case "kpi":
		// TODO: Implement
	case "kpi_grid":
		// TODO: Implement
	case "image":
		// TODO: Implement
	case "table":
		// TODO: Implement
	case "pivot":
		// TODO: Implement
	case "leaderboard":
		// TODO: Implement
	default:
		return fmt.Errorf("unsupported renderer %q", renderer)
	}

	return nil
}

// metricsViewHasDimension returns true if the metrics view has a dimension with the given name.
func metricsViewHasDimension(mv *runtimev1.MetricsViewSpec, fieldName string) bool {
	for _, d := range mv.Dimensions {
		if d.Name == fieldName {
			return true
		}
	}
	return false
}

// metricsViewHasMeasure returns true if the metrics view has a measure with the given name.
func metricsViewHasMeasure(mv *runtimev1.MetricsViewSpec, fieldName string) bool {
	for _, m := range mv.Measures {
		if m.Name == fieldName {
			return true
		}
	}
	return false
}
