package reconcilers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	compilerv1 "github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/robfig/cron/v3"
)

// checkRefs checks that all refs exist, are idle, and have no errors.
func checkRefs(ctx context.Context, c *runtime.Controller, refs []*runtimev1.ResourceName) error {
	for _, ref := range refs {
		res, err := c.Get(ctx, ref, false)
		if err != nil {
			if errors.Is(err, drivers.ErrResourceNotFound) {
				return fmt.Errorf("dependency error: resource %q (%s) not found", ref.Name, ref.Kind)
			}
			return fmt.Errorf("dependency error: failed to get resource %q (%s): %w", ref.Name, ref.Kind, err)
		}
		if res.Meta.ReconcileStatus != runtimev1.ReconcileStatus_RECONCILE_STATUS_IDLE {
			return fmt.Errorf("dependency error: resource %q (%s) is not idle", ref.Name, ref.Kind)
		}
		if res.Meta.ReconcileError != "" {
			return fmt.Errorf("dependency error: resource %q (%s) has an error", ref.Name, ref.Kind)
		}
	}
	return nil
}

// hasStreamingRef returns true if one or more of the refs have data that may be updated outside of a reconcile.
func hasStreamingRef(ctx context.Context, c *runtime.Controller, refs []*runtimev1.ResourceName) bool {
	for _, ref := range refs {
		// Currently only metrics views can be streaming.
		if ref.Kind != runtime.ResourceKindMetricsView {
			continue
		}

		res, err := c.Get(ctx, ref, false)
		if err != nil {
			// Broken refs are not streaming.
			continue
		}
		mv := res.GetMetricsView()

		if mv.State.Streaming {
			return true
		}
	}
	return false
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

// olapTableInfo returns info about a table in an OLAP connector
func olapTableInfo(ctx context.Context, c *runtime.Controller, connector, table string) (*drivers.Table, bool) {
	if table == "" {
		return nil, false
	}

	olap, release, err := c.AcquireOLAP(ctx, connector)
	if err != nil {
		return nil, false
	}
	defer release()

	t, err := olap.InformationSchema().Lookup(ctx, "", "", table)
	if err != nil {
		return nil, false
	}

	return t, true
}

// olapDropTableIfExists drops a table from an OLAP connector.
func olapDropTableIfExists(ctx context.Context, c *runtime.Controller, connector, table string) {
	if table == "" {
		return
	}

	olap, release, err := c.AcquireOLAP(ctx, connector)
	if err != nil {
		return
	}
	defer release()

	_ = olap.DropTable(ctx, table)
}

// olapForceRenameTable renames a table or view from fromName to toName in the OLAP connector.
// If a view or table already exists with toName, it is overwritten.
func olapForceRenameTable(ctx context.Context, c *runtime.Controller, connector, fromName string, fromIsView bool, toName string) error {
	if fromName == "" || toName == "" {
		return fmt.Errorf("cannot rename empty table name: fromName=%q, toName=%q", fromName, toName)
	}

	if fromName == toName {
		return nil
	}

	olap, release, err := c.AcquireOLAP(ctx, connector)
	if err != nil {
		return err
	}
	defer release()

	// Infer SQL keyword for the table type
	var typ string
	if fromIsView {
		typ = "VIEW"
	} else {
		typ = "TABLE"
	}

	// Renaming a table to the same name with different casing is not supported. Workaround by renaming to a temporary name first.
	if strings.EqualFold(fromName, toName) {
		tmpName := fmt.Sprintf("__rill_tmp_rename_%s_%s", typ, toName)
		err = olap.RenameTable(ctx, fromName, tmpName)
		if err != nil {
			return err
		}
		fromName = tmpName
	}

	// Do the rename
	return olap.RenameTable(ctx, fromName, toName)
}

func resolveTemplatedProps(ctx context.Context, c *runtime.Controller, self compilerv1.TemplateResource, props map[string]any) (map[string]any, error) {
	inst, err := c.Runtime.Instance(ctx, c.InstanceID)
	if err != nil {
		return nil, err
	}
	vars := inst.ResolveVariables(false)

	templateData := compilerv1.TemplateData{
		Environment: inst.Environment,
		User:        map[string]interface{}{},
		Variables:   vars,
		ExtraProps:  nil,
		Self:        self,
		Resolve: func(ref compilerv1.ResourceName) (string, error) {
			// We don't actually know if this "ref" is from a "sql:" property or something else.
			// If it's a SQL property, we also don't know what the SQL dialect in question is. (Do we even want to support "ref" outside of SQL?)
			// Using the DuckDB dialect escaping is going to work correctly in basically all cases, but it's a bit hacky.
			return drivers.DialectDuckDB.EscapeIdentifier(ref.Name), nil
		},
		Lookup: func(name compilerv1.ResourceName) (compilerv1.TemplateResource, error) {
			if name.Kind == compilerv1.ResourceKindUnspecified {
				return compilerv1.TemplateResource{}, fmt.Errorf("can't resolve name %q without type specified", name.Name)
			}
			res, err := c.Get(ctx, runtime.ResourceNameFromCompiler(name), false)
			if err != nil {
				return compilerv1.TemplateResource{}, err
			}
			return compilerv1.TemplateResource{
				Meta:  res.Meta,
				Spec:  res.Resource.(*runtimev1.Resource_Model).Model.Spec,
				State: res.Resource.(*runtimev1.Resource_Model).Model.State,
			}, nil
		},
	}

	for key, value := range props {
		res, err := convert(value, &templateData)
		if err != nil {
			return nil, fmt.Errorf("failed to convert property %q: %w", key, err)
		}
		props[key] = res
	}
	return props, nil
}

func convert(value any, templateData *compilerv1.TemplateData) (res any, err error) {
	switch v := value.(type) {
	case string:
		res, err = compilerv1.ResolveTemplate(v, *templateData)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve template: %w", err)
		}
	case map[string]any:
		for key, item := range v {
			item, err = convert(item, templateData)
			if err != nil {
				return nil, err
			}
			v[key] = item
		}
		res = v
	case []any:
		for i, item := range v {
			item, err = convert(item, templateData)
			if err != nil {
				return nil, err
			}
			v[i] = item
		}
		res = v
	default:
		res = v
	}
	return
}
