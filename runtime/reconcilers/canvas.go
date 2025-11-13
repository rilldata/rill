package reconcilers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/metricssql"
	"github.com/rilldata/rill/runtime/pkg/pathutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindCanvas, newCanvasReconciler)
}

type CanvasReconciler struct {
	C *runtime.Controller
}

func newCanvasReconciler(ctx context.Context, c *runtime.Controller) (runtime.Reconciler, error) {
	return &CanvasReconciler{C: c}, nil
}

func (r *CanvasReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *CanvasReconciler) AssignSpec(from, to *runtimev1.Resource) error {
	a := from.GetCanvas()
	b := to.GetCanvas()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign spec from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *CanvasReconciler) AssignState(from, to *runtimev1.Resource) error {
	a := from.GetCanvas()
	b := to.GetCanvas()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign state from %T to %T", from.Resource, to.Resource)
	}
	b.State = a.State
	return nil
}

func (r *CanvasReconciler) ResetState(res *runtimev1.Resource) error {
	res.GetCanvas().State = &runtimev1.CanvasState{}
	return nil
}

func (r *CanvasReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	c := self.GetCanvas()
	if c == nil {
		return runtime.ReconcileResult{Err: errors.New("not a canvas")}
	}

	// Exit early for deletion
	if self.Meta.DeletedOn != nil {
		return runtime.ReconcileResult{}
	}

	// Get instance config
	cfg, err := r.C.Runtime.InstanceConfig(ctx, r.C.InstanceID)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	// Get all referenced components.
	// If a referenced component is not found, we still add it to the map with a nil value.
	components := make(map[string]*runtimev1.Resource)
	for _, ref := range self.Meta.Refs {
		if ref.Kind != runtime.ResourceKindComponent {
			continue
		}
		res, err := r.C.Get(ctx, ref, false)
		if err != nil {
			if !errors.Is(err, drivers.ErrResourceNotFound) {
				return runtime.ReconcileResult{Err: err}
			}
			components[ref.Name] = nil // Component not found, add it to the map with nil value
		} else {
			components[ref.Name] = res
		}
	}

	// Find most recent data refresh time across all components.
	var dataRefreshedOn *timestamppb.Timestamp
	for _, c := range components {
		if c == nil {
			continue
		}
		t := c.GetComponent().State.DataRefreshedOn
		if t == nil {
			continue
		}
		if dataRefreshedOn == nil || t.AsTime().After(dataRefreshedOn.AsTime()) {
			dataRefreshedOn = t
		}
	}

	// Validate refs
	validateErr := checkRefs(ctx, r.C, self.Meta.Refs)
	if validateErr == nil {
		validateErr = r.validateMetricsViewTimeConsistency(ctx, components)
	}

	// Capture the valid spec in the state
	if validateErr == nil {
		c.State.ValidSpec = c.Spec
		c.State.DataRefreshedOn = dataRefreshedOn
	} else if cfg.StageChanges && r.checkAnyComponentHasValidSpec(components) {
		// When StageChanges is enabled, we want to make a best effort to serve the canvas anyway.
		// If any of the components referenced by the spec have a ValidSpec, we'll try to serve the canvas.
		c.State.ValidSpec = c.Spec
		c.State.DataRefreshedOn = dataRefreshedOn
	} else {
		c.State.ValidSpec = nil
		c.State.DataRefreshedOn = nil
	}

	// Update state. Even if the validation result is unchanged, we always update the state to ensure the state version is incremented.
	err = r.C.UpdateState(ctx, self.Meta.Name, self)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	return runtime.ReconcileResult{Err: validateErr}
}

// checkAnyComponentHasValidSpec returns true if one or more components have a valid spec.
func (r *CanvasReconciler) checkAnyComponentHasValidSpec(components map[string]*runtimev1.Resource) bool {
	for _, res := range components {
		if res == nil {
			// Component not found, skip it
			continue
		}
		if res.GetComponent().State.ValidSpec != nil {
			// Found component ref with a valid spec
			return true
		}
	}
	return false
}

func (r *CanvasReconciler) ResolveTransitiveAccess(ctx context.Context, claims *runtime.SecurityClaims, res *runtimev1.Resource) ([]*runtimev1.SecurityRule, error) {
	var rules []*runtimev1.SecurityRule
	var conditionKinds []string
	var conditionResources []*runtimev1.ResourceName
	refs := &rendererRefs{
		metricsViews: make(map[string]bool),
		mvFields:     make(map[string]map[string]bool),
		mvFilters:    make(map[string][]string),
	}

	canvas := res.GetCanvas()
	if canvas == nil {
		return nil, fmt.Errorf("resource is not a canvas")
	}

	spec := canvas.GetState().GetValidSpec()
	if spec == nil {
		spec = canvas.GetSpec() // Fallback to spec if ValidSpec is not available
	}
	if spec == nil {
		return nil, fmt.Errorf("canvas spec is nil")
	}

	// explicitly allow access to the canvas itself
	conditionResources = append(conditionResources, res.Meta.Name)
	conditionKinds = append(conditionKinds, runtime.ResourceKindTheme)

	// Get controller to fetch components
	ctr, err := r.C.Runtime.Controller(ctx, r.C.InstanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get controller: %w", err)
	}

	// Collect all component names referenced by the canvas
	componentNames := make(map[string]bool)
	for _, row := range spec.Rows {
		for _, item := range row.Items {
			componentNames[item.Component] = true
		}
	}

	// Process each component
	for componentName := range componentNames {
		componentRef := &runtimev1.ResourceName{
			Kind: runtime.ResourceKindComponent,
			Name: componentName,
		}
		// Allow access to the component itself
		conditionResources = append(conditionResources, componentRef)

		// Get component resource
		componentRes, err := ctr.Get(ctx, componentRef, false)
		if err != nil {
			// If component is not found, skip it but still allow access to the component name
			continue
		}

		// Get component spec to extract renderer properties
		componentSpec := componentRes.GetComponent().State.ValidSpec
		if componentSpec == nil {
			componentSpec = componentRes.GetComponent().Spec
		}

		if componentSpec.RendererProperties == nil {
			continue
		}

		rendererProps := componentSpec.RendererProperties.AsMap()
		err = populateRendererRefs(refs, componentSpec.Renderer, rendererProps)
		if err != nil {
			return nil, fmt.Errorf("failed to parse renderer properties for component %q: %w", componentName, err)
		}
	}

	// Now build security rules based on the collected references
	// First, allow access to all referenced metrics views
	// Then, for each metrics view, add field access and row filter rules as needed
	for mv := range refs.metricsViews {
		// allow access to the referenced metrics view
		conditionResources = append(conditionResources, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: mv})

		mvf, ok := refs.mvFields[mv]
		if ok && len(mvf) > 0 {
			fields := make([]string, 0, len(mvf))
			for f := range mvf {
				fields = append(fields, f)
			}
			rules = append(rules, &runtimev1.SecurityRule{
				Rule: &runtimev1.SecurityRule_FieldAccess{
					FieldAccess: &runtimev1.SecurityRuleFieldAccess{
						ConditionResources: []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: mv}},
						Fields:             fields,
						Allow:              true,
					},
				},
			})
		}

		mvr, ok := refs.mvFilters[mv]
		if ok && len(mvr) > 0 {
			// Combine multiple row filters with OR
			rowFilter := strings.Join(mvr, " OR ")
			rules = append(rules, &runtimev1.SecurityRule{
				Rule: &runtimev1.SecurityRule_RowFilter{
					RowFilter: &runtimev1.SecurityRuleRowFilter{
						ConditionResources: []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: mv}},
						Sql:                rowFilter,
					},
				},
			})
		}
	}

	if len(conditionKinds) > 0 || len(conditionResources) > 0 {
		rules = append(rules, &runtimev1.SecurityRule{
			Rule: &runtimev1.SecurityRule_Access{
				Access: &runtimev1.SecurityRuleAccess{
					ConditionKinds:     conditionKinds,
					ConditionResources: conditionResources,
					Allow:              true,
					Exclusive:          true,
				},
			},
		})
	}

	return rules, nil
}

// validateMetricsViewTimeConsistency checks that all the metrics views referenced by the canvas' components have the same first_day_of_week and first_month_of_year.
func (r *CanvasReconciler) validateMetricsViewTimeConsistency(ctx context.Context, components map[string]*runtimev1.Resource) error {
	metricsViews := make(map[string]*runtimev1.Resource)
	for _, component := range components {
		if component == nil {
			// Component not found, skip it
			continue
		}

		// Skip non-metrics view refs
		for _, ref := range component.Meta.Refs {
			if ref.Kind != runtime.ResourceKindMetricsView {
				continue
			}
			// Skip if the metrics view is already in the map
			if _, ok := metricsViews[ref.Name]; ok {
				continue
			}
			// Get the metrics view
			mv, err := r.C.Get(ctx, ref, false)
			if err != nil {
				continue
			}

			// Skip if the metrics view is not valid
			mvSpec := mv.GetMetricsView().State.ValidSpec
			if mvSpec == nil {
				continue
			}

			metricsViews[ref.Name] = mv
		}
	}

	// Validate all metrics views have consistent first_day_of_week or first_month_of_year
	if len(metricsViews) > 0 {
		first := false
		var firstDayOfWeek uint32
		var firstMonthOfYear uint32
		var firstViewName string

		for mvName, mv := range metricsViews {
			mvSpec := mv.GetMetricsView().State.ValidSpec
			if mvSpec == nil {
				continue
			}

			if !first {
				// This is the first metrics view, store its settings as reference
				first = true
				firstDayOfWeek = mvSpec.FirstDayOfWeek
				firstMonthOfYear = mvSpec.FirstMonthOfYear
				firstViewName = mvName
			} else {
				// Compare subsequent views with the first one
				if firstDayOfWeek != mvSpec.FirstDayOfWeek {
					return status.Errorf(codes.InvalidArgument, "metrics views %q and %q have inconsistent first_day_of_week", firstViewName, mvName)
				} else if firstMonthOfYear != mvSpec.FirstMonthOfYear {
					return status.Errorf(codes.InvalidArgument, "metrics views %q and %q have inconsistent first_month_of_year", firstViewName, mvName)
				}
			}
		}
	}

	return nil
}

// populateRendererRefs extracts all metricsview and its field names and filters from renderer properties based on the renderer type
// Depending on the component, fields will be named differently - Also there can be computed time dimension like <time_dim>_rill_TIME_GRAIN_<GRAIN>
//
//		"leaderboard" - "dimensions" and "measures"
//		"kpi_grid" - "dimensions" and "measures"
//		"table" - "columns" (can have computed time dim)
//		"pivot" - "row_dimensions", "col_dimensions" and "measures" (row/col can have computed time dim)
//		"heatmap" - "color"."field", "x"."field" and "y"."field"
//	 	"multi_metric_chart" - "measures" and "x"."field"
//		"funnel_chart" - "stage"."field", "measure"."field"
//		"donut_chart" - "color"."field", "measure"."field"
//		"bar_chart" - "color"."field", "x"."field" and "y"."field"
//		"line_chart" - "color"."field", "x"."field" and "y"."field"
//		"area_chart" - "color"."field", "x"."field" and "y"."field"
//		"stacked_bar" - "color"."field", "x"."field" and "y"."field"
//		"stacked_bar_normalized" - "color"."field", "x"."field" and "y"."field"
//		"markdown" - content may contain metrics_sql template functions; metrics views are resolved at query time via ResolveTemplatedString RPC
func populateRendererRefs(res *rendererRefs, renderer string, rendererProps map[string]any) error {
	mv, ok := pathutil.GetPath(rendererProps, "metrics_view")
	if !ok {
		return nil
	}
	err := res.metricsView(mv)
	if err != nil {
		return err
	}
	if filter, ok := pathutil.GetPath(rendererProps, "dimension_filters"); ok {
		err = res.metricsViewRowFilter(mv, filter)
		if err != nil {
			return err
		}
	}
	switch renderer {
	case "leaderboard":
		if dims, ok := pathutil.GetPath(rendererProps, "dimensions"); ok {
			err = res.metricsViewFields(mv, dims)
			if err != nil {
				return err
			}
		}
		if meas, ok := pathutil.GetPath(rendererProps, "measures"); ok {
			err = res.metricsViewFields(mv, meas)
			if err != nil {
				return err
			}
		}
	case "kpi_grid":
		if dims, ok := pathutil.GetPath(rendererProps, "dimensions"); ok {
			err = res.metricsViewFields(mv, dims)
			if err != nil {
				return err
			}
		}
		if meas, ok := pathutil.GetPath(rendererProps, "measures"); ok {
			err = res.metricsViewFields(mv, meas)
			if err != nil {
				return err
			}
		}
	case "table":
		if cols, ok := pathutil.GetPath(rendererProps, "columns"); ok {
			err = res.metricsViewFields(mv, cols)
			if err != nil {
				return err
			}
		}
	case "pivot":
		if rowDims, ok := pathutil.GetPath(rendererProps, "row_dimensions"); ok {
			err = res.metricsViewFields(mv, rowDims)
			if err != nil {
				return err
			}
		}
		if colDims, ok := pathutil.GetPath(rendererProps, "col_dimensions"); ok {
			err = res.metricsViewFields(mv, colDims)
			if err != nil {
				return err
			}
		}
		if meas, ok := pathutil.GetPath(rendererProps, "measures"); ok {
			err = res.metricsViewFields(mv, meas)
			if err != nil {
				return err
			}
		}
	case "heatmap":
		if colorField, ok := pathutil.GetPath(rendererProps, "color.field"); ok {
			err = res.metricsViewField(mv, colorField)
			if err != nil {
				return err
			}
		}
		if xField, ok := pathutil.GetPath(rendererProps, "x.field"); ok {
			err = res.metricsViewField(mv, xField)
			if err != nil {
				return err
			}
		}
		if yField, ok := pathutil.GetPath(rendererProps, "y.field"); ok {
			err = res.metricsViewField(mv, yField)
			if err != nil {
				return err
			}
		}
	case "multi_metric_chart":
		if meas, ok := pathutil.GetPath(rendererProps, "measures"); ok {
			err = res.metricsViewFields(mv, meas)
			if err != nil {
				return err
			}
		}
		if xField, ok := pathutil.GetPath(rendererProps, "x.field"); ok {
			err = res.metricsViewField(mv, xField)
			if err != nil {
				return err
			}
		}
	case "funnel_chart":
		if stageField, ok := pathutil.GetPath(rendererProps, "stage.field"); ok {
			err = res.metricsViewField(mv, stageField)
			if err != nil {
				return err
			}
		}
		if measureField, ok := pathutil.GetPath(rendererProps, "measure.field"); ok {
			err = res.metricsViewField(mv, measureField)
			if err != nil {
				return err
			}
		}
	case "donut_chart":
		if colorField, ok := pathutil.GetPath(rendererProps, "color.field"); ok {
			err = res.metricsViewField(mv, colorField)
			if err != nil {
				return err
			}
		}
		if measureField, ok := pathutil.GetPath(rendererProps, "measure.field"); ok {
			err = res.metricsViewField(mv, measureField)
			if err != nil {
				return err
			}
		}
	case "bar_chart", "line_chart", "area_chart", "stacked_bar", "stacked_bar_normalized":
		if colorField, ok := pathutil.GetPath(rendererProps, "color.field"); ok {
			err = res.metricsViewField(mv, colorField)
			if err != nil {
				return err
			}
		}
		if xField, ok := pathutil.GetPath(rendererProps, "x.field"); ok {
			err = res.metricsViewField(mv, xField)
			if err != nil {
				return err
			}
		}
		if yField, ok := pathutil.GetPath(rendererProps, "y.field"); ok {
			err = res.metricsViewField(mv, yField)
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unknown renderer type %q", renderer)
	}
	return nil
}

// extractDimension return the dimension or extracts the base time dimension from computed time field if present
// example - from "<time_dim>_rill_TIME_GRAIN_<GRAIN>" extracts "<time_dim>"
func extractDimension(field string) string {
	if strings.Contains(field, "_rill_TIME_GRAIN_") {
		parts := strings.Split(field, "_rill_TIME_GRAIN_")
		if len(parts) > 0 {
			return parts[0]
		}
	}
	return field
}

type rendererRefs struct {
	metricsViews map[string]bool
	mvFields     map[string]map[string]bool
	mvFilters    map[string][]string
}

func (r *rendererRefs) metricsView(mv any) error {
	if mv, ok := mv.(string); ok {
		r.metricsViews[mv] = true
		return nil
	}
	return fmt.Errorf("metrics view field is not a string")
}

func (r *rendererRefs) metricsViewFields(mv, fields any) error {
	metricsView, ok1 := mv.(string)
	fs, ok2 := fields.([]interface{})
	if !ok1 || !ok2 {
		return fmt.Errorf("metrics view field is not a string or fields is not a list")
	}
	if r.mvFields[metricsView] == nil {
		r.mvFields[metricsView] = make(map[string]bool)
	}
	for _, f := range fs {
		fstr, ok := f.(string)
		if !ok {
			return fmt.Errorf("field is not a string")
		}
		r.mvFields[metricsView][extractDimension(fstr)] = true
	}
	return nil
}

func (r *rendererRefs) metricsViewField(mv, field any) error {
	metricsView, ok1 := mv.(string)
	f, ok2 := field.(string)
	if !ok1 || !ok2 {
		return fmt.Errorf("metrics view field is not a string or field is not a string")
	}
	if f == "" {
		return nil
	}
	if r.mvFields[metricsView] == nil {
		r.mvFields[metricsView] = make(map[string]bool)
	}
	r.mvFields[metricsView][extractDimension(f)] = true

	return nil
}

func (r *rendererRefs) metricsViewRowFilter(mv, filter any) error {
	metricsView, ok1 := mv.(string)
	f, ok2 := filter.(string)
	if !ok1 || !ok2 {
		return fmt.Errorf("metrics view field is not a string or filter is not a string")
	}
	if f == "" {
		return nil
	}
	r.mvFilters[metricsView] = append(r.mvFilters[metricsView], fmt.Sprintf("(%s)", f)) // wrap in () to ensure correct precedence when combining multiple filters with OR
	// Extract fields from dimension_filters SQL expression
	ex, err := metricssql.ParseFilter(f)
	if err != nil {
		return fmt.Errorf("failed to parse dimension_filters SQL expression %q: %w", f, err)
	}
	dimFilterFields := metricsview.AnalyzeExpressionFields(ex)
	if r.mvFields[metricsView] == nil {
		r.mvFields[metricsView] = make(map[string]bool)
	}
	for _, f := range dimFilterFields {
		r.mvFields[metricsView][extractDimension(f)] = true
	}
	return nil
}
