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
	"github.com/robfig/cron/v3"
	"golang.org/x/exp/slog"
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

// nextRefreshTime returns the earliest time AFTER t that the schedule should trigger.
func nextRefreshTime(t time.Time, schedule *runtimev1.Schedule) (time.Time, error) {
	if schedule == nil {
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

// olapTableInfo returns info about a table in an OLAP connector.
func olapTableInfo(ctx context.Context, c *runtime.Controller, connector, table string) (*drivers.Table, bool) {
	if table == "" {
		return nil, false
	}

	olap, release, err := c.AcquireOLAP(ctx, connector)
	if err != nil {
		return nil, false
	}
	defer release()

	t, err := olap.InformationSchema().Lookup(ctx, table)
	if err != nil {
		return nil, false
	}

	return t, true
}

// olapDropTableIfExists drops a table from an OLAP connector.
func olapDropTableIfExists(ctx context.Context, c *runtime.Controller, connector, table string, view bool) {
	if table == "" {
		return
	}

	olap, release, err := c.AcquireOLAP(ctx, connector)
	if err != nil {
		return
	}
	defer release()

	_ = olap.DropTable(ctx, table, view)
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
		err = olap.RenameTable(ctx, fromName, tmpName, fromIsView)
		if err != nil {
			return err
		}
		fromName = tmpName
	}

	// Do the rename
	return olap.RenameTable(ctx, fromName, toName, fromIsView)
}

// safeSQLName returns a quoted SQL identifier.
func safeSQLName(name string) string {
	if name == "" {
		return name
	}
	return fmt.Sprintf("\"%s\"", strings.ReplaceAll(name, "\"", "\"\""))
}

func logTableNameAndType(ctx context.Context, c *runtime.Controller, connector, name string) {
	olap, release, err := c.AcquireOLAP(ctx, connector)
	if err != nil {
		c.Logger.Error("LogTableNameAndType: failed to acquire OLAP", slog.Any("err", err))
		return
	}
	defer release()

	res, err := olap.Execute(context.Background(), &drivers.Statement{Query: "SELECT column_name, data_type FROM information_schema.columns WHERE table_name=? ORDER BY column_name ASC", Args: []any{name}})
	if err != nil {
		c.Logger.Error("LogTableNameAndType: failed information_schema.columns", slog.Any("err", err))
		return
	}
	defer res.Close()

	colTyp := make([]string, 0)
	var col, typ string
	for res.Next() {
		err = res.Scan(&col, &typ)
		if err != nil {
			c.Logger.Error("LogTableNameAndType: failed scan", slog.Any("err", err))
			return
		}
		colTyp = append(colTyp, fmt.Sprintf("%s:%s", col, typ))
	}

	c.Logger.Info("LogTableNameAndType: ", slog.String("name", name), slog.String("schema", strings.Join(colTyp, ", ")))
}
