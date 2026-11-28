package reconcilers

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/robfig/cron/v3"
)

// checkRefs checks that all refs exist, are idle, and have no errors.
func checkRefs(ctx context.Context, c *runtime.Controller, refs []*runtimev1.ResourceName) error {
	// Look up the refs and collect any errors we find.
	var errs []string
	for _, ref := range refs {
		res, err := c.Get(ctx, ref, false)
		if err != nil {
			if !errors.Is(err, drivers.ErrResourceNotFound) {
				// Return immediately for other errors (most likely a context cancellation).
				return runtime.NewDependencyError(fmt.Errorf("failed to get resource %s/%s: %w", runtime.PrettifyResourceKind(ref.Kind), ref.Name, err))
			}
			errs = append(errs, fmt.Sprintf("resource %s/%s not found", runtime.PrettifyResourceKind(ref.Kind), ref.Name))
			continue
		}
		if res.Meta.ReconcileStatus != runtimev1.ReconcileStatus_RECONCILE_STATUS_IDLE {
			errs = append(errs, fmt.Sprintf("resource %s/%s is not idle", runtime.PrettifyResourceKind(ref.Kind), ref.Name))
			continue
		}
		if res.Meta.ReconcileError == "" {
			continue
		}
		switch res.Meta.Name.Kind {
		case runtime.ResourceKindComponent:
			// For components, we propagate the underlying error message to make canvas debugging easier (since components are usually defined inline in a canvas).
			// This is safe since components are already end-user facing resources.
			errs = append(errs, fmt.Sprintf("component %s has an error: %s", ref.Name, res.Meta.ReconcileError))
		default:
			// For other resources, we don't propagate errors for one of two reasons:
			// 1. In some cases, such as models, the error message may contain sensitive information that non-admins shouldn't see.
			// 2. In other cases, it's sufficient to reference the underlying resource, which the admin can investigate separately. This prevents an error from getting multiplied in logs and UIs.
			errs = append(errs, fmt.Sprintf("resource %s/%s has an error", runtime.PrettifyResourceKind(ref.Kind), ref.Name))
		}
	}

	// Only return the full first error; truncate any additional errors.
	if len(errs) == 1 {
		return runtime.NewDependencyError(errors.New(errs[0]))
	} else if len(errs) > 1 {
		return runtime.NewDependencyError(fmt.Errorf("%s; %d other refs also have errors", errs[0], len(errs)-1))
	}

	return nil
}

// checkStreamingRef returns true if one or more of the refs have data that may be updated outside of a reconcile.
// If so, it also returns whether any ref may scale to zero.
func checkStreamingRef(ctx context.Context, c *runtime.Controller, refs []*runtimev1.ResourceName) (ok, mayScaleToZero bool, err error) {
	for _, ref := range refs {
		// Currently only metrics views can be streaming.
		if ref.Kind != runtime.ResourceKindMetricsView {
			continue
		}

		res, err := c.Get(ctx, ref, false)
		if err != nil {
			if errors.Is(err, drivers.ErrResourceNotFound) {
				// Broken refs are not streaming.
				continue
			}
			return false, false, err
		}
		mv := res.GetMetricsView()

		// Don't consider invalid metrics views.
		if mv.State.ValidSpec == nil {
			continue
		}

		// Don't consider non-streaming metrics views.
		if !mv.State.Streaming {
			continue
		}

		// We found a streaming ref
		ok = true

		// Check if it may scale to zero
		olap, release, err := c.AcquireOLAP(ctx, mv.State.ValidSpec.Connector)
		if err != nil {
			return false, false, err
		}
		if olap.MayBeScaledToZero(ctx) {
			mayScaleToZero = true
		}
		release()
	}
	return ok, mayScaleToZero, nil
}

// nextRefreshTime returns the earliest time AFTER t that the schedule should trigger.
func nextRefreshTime(t time.Time, schedule *runtimev1.Schedule) (time.Time, error) {
	if schedule == nil || schedule.Disable {
		return time.Time{}, nil
	}

	var t1 time.Time
	if schedule.TickerSeconds > 0 {
		d := time.Duration(schedule.TickerSeconds) * time.Second
		t1 = t.Add(d)
	}

	var t2 time.Time
	if schedule.Cron != "" {
		crontab := schedule.Cron
		if schedule.TimeZone != "" {
			if !strings.HasPrefix(crontab, "TZ=") && !strings.HasPrefix(crontab, "CRON_TZ=") {
				crontab = fmt.Sprintf("CRON_TZ=%s %s", schedule.TimeZone, crontab)
			}
		}

		cs, err := cron.ParseStandard(crontab)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to parse cron schedule: %w", err)
		}
		t2 = cs.Next(t)
	}

	if t1.IsZero() {
		return t2, nil
	}
	if t2.IsZero() {
		return t1, nil
	}
	if t1.Before(t2) {
		return t1, nil
	}
	return t2, nil
}

// computeInheritedWatermark returns the earliest watermark of the refs of an alert or report, and false if none of them has one.
// Explores, which AI alerts and reports reference, inherit the watermark of their metrics view.
func computeInheritedWatermark(ctx context.Context, c *runtime.Controller, refs []*runtimev1.ResourceName, priority int) (time.Time, bool, error) {
	var t time.Time
	for _, ref := range refs {
		// Resolving the explore's metrics view here instead of in the watermark query keeps the query's cache keyed on the metrics view's data.
		if ref.Kind == runtime.ResourceKindExplore {
			res, err := c.Get(ctx, ref, false)
			if err != nil {
				return t, false, fmt.Errorf("failed to get explore %q: %w", ref.Name, err)
			}
			spec := res.GetExplore().State.ValidSpec
			if spec == nil {
				return t, false, fmt.Errorf("explore %q is not valid", ref.Name)
			}
			ref = &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: spec.MetricsView}
		}

		q := &queries.ResourceWatermark{
			ResourceKind: ref.Kind,
			ResourceName: ref.Name,
		}
		err := c.Runtime.Query(ctx, c.InstanceID, q, priority)
		if err != nil {
			return t, false, fmt.Errorf("failed to resolve watermark for %s/%s: %w", ref.Kind, ref.Name, err)
		}

		if q.Result != nil && (t.IsZero() || q.Result.Before(t)) {
			t = *q.Result
		}
	}

	return t, !t.IsZero(), nil
}

// exploreNameFromAnnotations extracts the explore name from resource annotations.
// It checks the "explore" annotation first, then falls back to parsing the "web_open_path" annotation.
// If neither is found, it returns fallback.
func exploreNameFromAnnotations(annotations map[string]string, fallback string) string {
	if e, ok := annotations["explore"]; ok {
		return e
	}
	if path, ok := annotations["web_open_path"]; ok {
		if strings.HasPrefix(path, "/explore/") {
			explore := path[9:]
			if explore != "" && explore[len(explore)-1] == '/' {
				explore = explore[:len(explore)-1]
			}
			if decoded, err := url.PathUnescape(explore); err == nil {
				explore = decoded
			}
			if explore != "" {
				return explore
			}
		}
	}
	return fallback
}

// dashboardConditionKinds returns the kinds of resources that transitive access to a dashboard grants in full, whatever the dashboard contains.
// Those are the themes that style the dashboard,
// and, if the claims can use AI, the skills that the dashboard's AI chat lists in its "/" picker.
// The AI agents load skills regardless of resource restrictions, so the picker would otherwise miss skills the agents use.
func dashboardConditionKinds(claims *runtime.SecurityClaims) []string {
	kinds := []string{runtime.ResourceKindTheme}
	if claims.Can(runtime.UseAI) {
		kinds = append(kinds, runtime.ResourceKindSkill)
	}
	return kinds
}

// analyzeTemplatedVariables analyzes strings nested in the provided props for template tags that reference instance variables.
// It returns a map of variable names referenced in the props mapped to their current value (if known).
func analyzeTemplatedVariables(ctx context.Context, c *runtime.Controller, props map[string]any) (map[string]string, error) {
	res := make(map[string]string)
	err := parser.AnalyzeTemplateRecursively(props, res)
	if err != nil {
		return nil, err
	}

	inst, err := c.Runtime.Instance(ctx, c.InstanceID)
	if err != nil {
		return nil, err
	}
	vars := inst.ResolveVariables(false)

	for k := range res {
		// Project variables are referenced with .env.name (current) or .vars.name (deprecated).
		// Other templated variable names are not project variable references.
		if k2 := strings.TrimPrefix(k, "env."); k != k2 {
			res[k] = vars[k2]
		} else if k2 := strings.TrimPrefix(k, "vars."); k != k2 {
			res[k] = vars[k2]
		}
	}

	return res, nil
}
