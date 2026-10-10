package runtime

import (
	"context"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/canvas"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview/metricssql"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// CollectCanvasComponentNames collects the names of all components referenced by the given rows,
// descending into tab groups (one level deep, since tabs cannot be nested).
func CollectCanvasComponentNames(rows []*runtimev1.CanvasRow, out map[string]bool) {
	for _, row := range rows {
		for _, item := range row.Items {
			out[item.Component] = true
		}
		if tg := row.GetTabGroup(); tg != nil {
			for _, tab := range tg.Tabs {
				CollectCanvasComponentNames(tab.Rows, out)
			}
		}
	}
}

// ResolveCanvasOptions configures ResolveCanvas.
type ResolveCanvasOptions struct {
	// Unsafe falls back to the canvas's unvalidated spec when it has no valid spec.
	// Only the visual editor sets it; it must never be set by Rill Cloud viewers, read-only previews, shared token access, or embedded viewers.
	Unsafe bool
	// IncludeHidden returns the canvas without removing the content hidden from the claims, and lists that content in ResolveCanvasResult.HiddenPaths.
	// The visual editor needs it, because it maps canvas elements to their position in the YAML.
	// It requires permission to edit the canvas; ResolveCanvas returns ErrForbidden otherwise.
	IncludeHidden bool
}

type ResolveCanvasResult struct {
	Canvas                 *runtimev1.Resource
	ResolvedComponents     map[string]*runtimev1.Resource
	ReferencedMetricsViews map[string]*runtimev1.Resource
	// HiddenPaths are the YAML paths of the canvas elements hidden from the claims, such as "rows.0.items.1" or "rows.3.tabs.0".
	// It's only set when ResolveCanvasOptions.IncludeHidden is set.
	HiddenPaths []string
}

func (r *Runtime) ResolveCanvas(ctx context.Context, instanceID, canvasName string, claims *SecurityClaims, opts *ResolveCanvasOptions) (*ResolveCanvasResult, error) {
	// Find the canvas resource
	ctrl, err := r.Controller(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	res, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: ResourceKindCanvas, Name: canvasName}, false)
	if err != nil {
		return nil, err
	}

	// Check if the user has access to the canvas
	sec, err := r.ResolveSecurity(ctx, instanceID, claims, res)
	if err != nil {
		return nil, fmt.Errorf("failed to apply security policy: %w", err)
	}
	if !sec.CanAccess() {
		return nil, ErrForbidden
	}

	// The visual editor gets the canvas with the content hidden from the claims.
	// Older editors only send unsafe, so it also includes hidden content for callers who may edit the canvas.
	includeHidden := opts.IncludeHidden
	if opts.IncludeHidden || opts.Unsafe {
		allowed := canIncludeHiddenCanvasContent(claims, res)
		if opts.IncludeHidden && !allowed {
			return nil, ErrForbidden
		}
		includeHidden = allowed
	}
	if !includeHidden {
		res = r.applyCanvasSecurity(res, sec)
	}

	// Use the valid spec if available. If unsafe is set, fall back to the unvalidated spec.
	spec := res.GetCanvas().State.ValidSpec
	if spec == nil && opts.Unsafe {
		spec = res.GetCanvas().Spec
	}
	if spec == nil {
		return &ResolveCanvasResult{
			Canvas: res,
		}, nil
	}

	components := make(map[string]*runtimev1.Resource)

	// Collect all referenced component names, descending into tab groups.
	componentNames := make(map[string]bool)
	CollectCanvasComponentNames(spec.Rows, componentNames)

	for componentName := range componentNames {
		// Get component resource.
		cmp, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: ResourceKindComponent, Name: componentName}, false)
		if err != nil {
			if errors.Is(err, drivers.ErrResourceNotFound) {
				return nil, fmt.Errorf("component %q in valid spec not found", componentName)
			}
			return nil, err
		}

		// Add to map without resolving templates. Use ResolveTemplatedString RPC for template resolution.
		components[componentName] = cmp
	}

	// Track the components the claims can't see or query; viewers don't get them, and the editor gets their paths.
	// An inline component keeps the `if` conditions of its item wherever it's referenced, including from another canvas.
	// Only its conditions are checked, not the full security policy, since claims granted just this canvas (e.g. alert links) can't access components directly.
	forbiddenComponents := make(map[string]bool)
	for componentName, cmp := range components {
		met, err := r.ComponentConditionMet(ctx, instanceID, claims, cmp)
		if err != nil {
			return nil, err
		}
		if !met {
			forbiddenComponents[componentName] = true
		}
	}

	// Resolve each metrics view and its security once.
	// Resolving security expands the claims' transitive access rules on every call, even when its result is cached.
	type resolvedMetricsView struct {
		res *runtimev1.Resource
		sec *ResolvedSecurity
		// getErr is set if the metrics view can't be found, and secErr if its security policy fails to evaluate.
		getErr, secErr error
	}
	resolvedMetricsViews := make(map[string]*resolvedMetricsView)
	resolveMetricsView := func(ctx context.Context, name string) *resolvedMetricsView {
		if rmv, ok := resolvedMetricsViews[name]; ok {
			return rmv
		}
		rmv := &resolvedMetricsView{}
		rmv.res, rmv.getErr = ctrl.Get(ctx, &runtimev1.ResourceName{Kind: ResourceKindMetricsView, Name: name}, false)
		if rmv.getErr == nil {
			rmv.sec, rmv.secErr = r.ResolveSecurity(ctx, instanceID, claims, rmv.res)
			// The SQL compiler looks a metrics view up by the name in the query, but returns queries with its resource name.
			resolvedMetricsViews[rmv.res.Meta.Name.Name] = rmv
		}
		resolvedMetricsViews[name] = rmv
		return rmv
	}

	// Extract metrics view names from components.
	// Components with a metrics_sql query the claims can't run are added to forbiddenComponents.
	var msqlParser *metricssql.Compiler
	metricsViews := make(map[string]bool)
	for componentName, cmp := range components {
		// Viewers don't get the metrics views of components hidden from them.
		if forbiddenComponents[componentName] && !includeHidden {
			continue
		}

		validSpec := cmp.GetComponent().State.ValidSpec
		if validSpec == nil && opts.Unsafe {
			validSpec = cmp.GetComponent().Spec
		}
		if validSpec == nil || validSpec.RendererProperties == nil {
			continue
		}

		for k, v := range validSpec.RendererProperties.Fields {
			switch k {
			case "metrics_view":
				if name := v.GetStringValue(); name != "" {
					metricsViews[name] = true
				}
			case "metrics_sql":
				// Instantiate a metrics SQL parser
				if msqlParser == nil {
					msqlParser = metricssql.New(&metricssql.CompilerOptions{
						GetMetricsView: func(ctx context.Context, name string) (*runtimev1.Resource, error) {
							rmv := resolveMetricsView(ctx, name)
							if rmv.getErr != nil {
								return nil, rmv.getErr
							}
							if rmv.secErr != nil {
								if ctx.Err() != nil {
									return nil, ctx.Err()
								}
								// A policy that fails to evaluate denies access, as for the metrics views below.
								return nil, ErrForbidden
							}
							if !rmv.sec.CanAccess() {
								return nil, ErrForbidden
							}
							return rmv.res, nil
						},
					})
				}

				// Create list of queries to analyze
				var queries []string
				if s := v.GetStringValue(); s != "" {
					queries = append(queries, s)
				} else if vals := v.GetListValue(); vals != nil {
					for _, val := range vals.Values {
						if s := val.GetStringValue(); s != "" {
							queries = append(queries, s)
						}
					}
				}

				// Analyze each query
				filter := validSpec.RendererProperties.Fields["dimension_filters"].GetStringValue()
				for _, sql := range queries {
					q, err := msqlParser.Parse(ctx, sql)
					if errors.Is(err, ErrForbidden) {
						forbiddenComponents[componentName] = true
						continue
					}
					if err != nil || q.MetricsView == "" {
						continue
					}
					metricsViews[q.MetricsView] = true

					// A query that selects or filters on a field the claims can't access fails, so the component can't render.
					// The frontend adds the component's own filter to each of its queries, so it's checked against each query's metrics view.
					// The compiler has already resolved that metrics view, and checked that the claims can access it.
					rmv := resolveMetricsView(ctx, q.MetricsView)
					if rmv.sec.CanAccessAllFields() {
						continue
					}
					timeDimension := rmv.res.GetMetricsView().GetState().GetValidSpec().GetTimeDimension()
					if canvas.QueryReferencesRestrictedFields(q, timeDimension, rmv.sec.CanAccessField) ||
						canvas.FilterReferencesRestrictedFields(filter, timeDimension, rmv.sec.CanAccessField) {
						forbiddenComponents[componentName] = true
					}
				}
			}
		}
	}

	// Lookup the metrics views and apply their security policies.
	// A metrics view whose policy fails to evaluate for the claims is treated as inaccessible instead of failing the canvas.
	// For example, a policy like '{{ .user.admin }}' doesn't evaluate for an embed token that carries no 'admin' attribute.
	referencedMetricsViews := make(map[string]*runtimev1.Resource, len(metricsViews))
	securities := make(map[string]*ResolvedSecurity, len(metricsViews))
	for name := range metricsViews {
		rmv := resolveMetricsView(ctx, name)
		if rmv.getErr != nil {
			if errors.Is(rmv.getErr, drivers.ErrResourceNotFound) {
				return nil, fmt.Errorf("metrics view %q in valid spec not found", name)
			}
			return nil, rmv.getErr
		}
		if rmv.secErr != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			r.Logger.Warn("treating metrics view as inaccessible because its security policy failed to evaluate", zap.String("instance_id", instanceID), zap.String("canvas", canvasName), zap.String("metrics_view", name), zap.Error(rmv.secErr), observability.ZapCtx(ctx))
			continue
		}
		if !rmv.sec.CanAccess() {
			continue
		}
		referencedMetricsViews[name] = r.applyMetricsViewSecurity(rmv.res, rmv.sec)
		securities[name] = rmv.sec
	}

	// The visual editor gets the full canvas and the paths of the content hidden from the claims:
	// the elements whose `if` conditions aren't met, and the components the claims can't query.
	if includeHidden {
		removed, _ := inaccessibleComponents(components, referencedMetricsViews, securities, forbiddenComponents)
		keep := CanvasConditionsKeep(sec.canvasConditions)
		itemMet := keep.Item
		keep.Item = func(item *runtimev1.CanvasItem) bool {
			return itemMet(item) && !removed[item.Component]
		}
		_, hiddenPaths := canvas.PruneRows(spec.Rows, keep)
		return &ResolveCanvasResult{
			Canvas:                 res,
			ResolvedComponents:     components,
			ReferencedMetricsViews: referencedMetricsViews,
			HiddenPaths:            hiddenPaths,
		}, nil
	}

	// Remove the fields and components the claims can't query, so viewers don't see errors for content they aren't allowed to see.
	res, components = pruneInaccessibleContent(res, components, referencedMetricsViews, securities, forbiddenComponents)
	res = pruneDefaultFilters(res, referencedMetricsViews)

	// Return the response
	return &ResolveCanvasResult{
		Canvas:                 res,
		ResolvedComponents:     components,
		ReferencedMetricsViews: referencedMetricsViews,
	}, nil
}

// pruneInaccessibleContent removes the fields that the claims can't query from the components' renderer properties,
// and removes the components that can't render without them, along with their items in the canvas rows (see inaccessibleComponents).
// The provided resources are not modified; anything that changes is cloned.
func pruneInaccessibleContent(canvasRes *runtimev1.Resource, components, metricsViews map[string]*runtimev1.Resource, securities map[string]*ResolvedSecurity, forbiddenComponents map[string]bool) (*runtimev1.Resource, map[string]*runtimev1.Resource) {
	removed, pruned := inaccessibleComponents(components, metricsViews, securities, forbiddenComponents)
	for name, cmp := range pruned {
		components[name] = cmp
	}
	if len(removed) == 0 {
		return canvasRes, components
	}

	for name := range removed {
		delete(components, name)
	}
	keep := canvas.Keep{Item: func(item *runtimev1.CanvasItem) bool { return !removed[item.Component] }}

	canvasRes = proto.Clone(canvasRes).(*runtimev1.Resource)
	if spec := canvasRes.GetCanvas().State.ValidSpec; spec != nil {
		spec.Rows, _ = canvas.PruneRows(spec.Rows, keep)
	}
	if spec := canvasRes.GetCanvas().Spec; spec != nil {
		spec.Rows, _ = canvas.PruneRows(spec.Rows, keep)
	}
	refs := canvasRes.Meta.Refs[:0]
	for _, ref := range canvasRes.Meta.Refs {
		if ref.Kind == ResourceKindComponent && removed[ref.Name] {
			continue
		}
		refs = append(refs, ref)
	}
	canvasRes.Meta.Refs = refs

	return canvasRes, components
}

// inaccessibleComponents determines which components the claims can't query.
// It returns the names of the components to remove,
// and clones of the components whose renderer properties lost references to restricted fields.
// A component is removed when its metrics view is inaccessible (absent from metricsViews),
// when a metrics_sql query targets a metrics view the claims can't access,
// or when it can't render without its restricted fields (e.g. a chart whose x or y field is restricted).
func inaccessibleComponents(components, metricsViews map[string]*runtimev1.Resource, securities map[string]*ResolvedSecurity, forbiddenComponents map[string]bool) (map[string]bool, map[string]*runtimev1.Resource) {
	removed := make(map[string]bool)
	pruned := make(map[string]*runtimev1.Resource)
	for name, cmp := range components {
		if forbiddenComponents[name] {
			removed[name] = true
			continue
		}

		validSpec := cmp.GetComponent().State.ValidSpec
		if validSpec == nil || validSpec.RendererProperties == nil {
			continue
		}
		mvName := validSpec.RendererProperties.Fields["metrics_view"].GetStringValue()
		if mvName == "" {
			continue
		}
		mv, ok := metricsViews[mvName]
		if !ok {
			removed[name] = true
			continue
		}
		sec := securities[mvName]
		if sec == nil || sec.CanAccessAllFields() {
			continue
		}

		timeDimension := mv.GetMetricsView().State.ValidSpec.GetTimeDimension()
		props := validSpec.RendererProperties.AsMap()
		keep, changed := canvas.PruneRendererProperties(validSpec.Renderer, props, timeDimension, sec.CanAccessField)
		if !keep {
			removed[name] = true
			continue
		}
		if !changed {
			continue
		}
		propsPB, err := structpb.NewStruct(props)
		if err != nil {
			removed[name] = true
			continue
		}

		// Viewers only use the valid spec, so the unvalidated spec gets the same pruned properties.
		cmp = proto.Clone(cmp).(*runtimev1.Resource)
		cmp.GetComponent().State.ValidSpec.RendererProperties = propsPB
		if spec := cmp.GetComponent().Spec; spec != nil {
			spec.RendererProperties = propsPB
		}
		pruned[name] = cmp
	}
	return removed, pruned
}

// pruneDefaultFilters removes the default filters for metrics views that are no longer on the canvas for the claims.
func pruneDefaultFilters(canvasRes *runtimev1.Resource, metricsViews map[string]*runtimev1.Resource) *runtimev1.Resource {
	stale := func(spec *runtimev1.CanvasSpec) bool {
		for mv := range spec.GetDefaultPreset().GetFilterExpr() {
			if _, ok := metricsViews[mv]; !ok {
				return true
			}
		}
		return false
	}
	c := canvasRes.GetCanvas()
	if !stale(c.GetSpec()) && !stale(c.GetState().GetValidSpec()) {
		return canvasRes
	}

	canvasRes = proto.Clone(canvasRes).(*runtimev1.Resource)
	c = canvasRes.GetCanvas()
	for _, spec := range []*runtimev1.CanvasSpec{c.GetSpec(), c.GetState().GetValidSpec()} {
		for mv := range spec.GetDefaultPreset().GetFilterExpr() {
			if _, ok := metricsViews[mv]; !ok {
				delete(spec.DefaultPreset.FilterExpr, mv)
			}
		}
	}
	return canvasRes
}
