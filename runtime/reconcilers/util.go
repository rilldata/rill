package reconcilers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/robfig/cron/v3"
)

// checkRefs checks that all refs exist, are idle, and have no errors.
func checkRefs(ctx context.Context, c *runtime.Controller, refs []*runtimev1.ResourceName) error {
	for _, ref := range refs {
		res, err := c.Get(ctx, ref, false)
		if err != nil {
			if errors.Is(err, drivers.ErrResourceNotFound) {
				return runtime.NewDependencyError(fmt.Errorf("resource %q (%s) not found", ref.Name, ref.Kind))
			}
			return runtime.NewDependencyError(fmt.Errorf("failed to get resource %q (%s): %w", ref.Name, ref.Kind, err))
		}
		if res.Meta.ReconcileStatus != runtimev1.ReconcileStatus_RECONCILE_STATUS_IDLE {
			return runtime.NewDependencyError(fmt.Errorf("resource %q (%s) is not idle", ref.Name, ref.Kind))
		}
		if res.Meta.ReconcileError != "" {
			return runtime.NewDependencyError(fmt.Errorf("resource %q (%s) has an error", ref.Name, ref.Kind))
		}
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
