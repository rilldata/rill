package reconcilers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/robfig/cron/v3"
)

// checkRefs checks that all refs exist, are idle, and have no errors.
func checkRefs(ctx context.Context, c *runtime.Controller, refs []*runtimev1.ResourceName) error {
	for _, ref := range refs {
		res, err := c.Get(ctx, ref)
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
		cs, err := cron.ParseStandard(schedule.Cron)
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

	var typ string
	if view {
		typ = "VIEW"
	} else {
		typ = "TABLE"
	}

	_ = olap.Exec(ctx, &drivers.Statement{
		Query:       fmt.Sprintf("DROP %s IF EXISTS %s", typ, safeSQLName(table)),
		Priority:    100,
		LongRunning: true,
	})
}

// olapRenameTable renames the table from oldName to newName in the OLAP connector.
// oldName must exist and newName must not exist.
func olapRenameTable(ctx context.Context, c *runtime.Controller, connector, oldName, newName string, view bool) error {
	if oldName == "" || newName == "" {
		return fmt.Errorf("cannot rename empty table name: oldName=%q, newName=%q", oldName, newName)
	}

	if oldName == newName {
		return nil
	}

	olap, release, err := c.AcquireOLAP(ctx, connector)
	if err != nil {
		return err
	}
	defer release()

	var typ string
	if view {
		typ = "VIEW"
	} else {
		typ = "TABLE"
	}

	// TODO: Use a transaction?
	return olap.WithConnection(ctx, 100, true, func(ctx context.Context, ensuredCtx context.Context, conn *sql.Conn) error {
		// Renaming a table to the same name with different casing is not supported. Workaround by renaming to a temporary name first.
		if strings.EqualFold(oldName, newName) {
			tmp := "__rill_tmp_rename_%s_" + typ + newName
			err = olap.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("DROP %s IF EXISTS %s", typ, safeSQLName(tmp))})
			if err != nil {
				return err
			}

			err := olap.Exec(ctx, &drivers.Statement{
				Query:    fmt.Sprintf("ALTER %s %s RENAME TO %s", typ, safeSQLName(oldName), safeSQLName(tmp)),
				Priority: 100,
			})
			if err != nil {
				return err
			}
			oldName = tmp
		}

		return olap.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("ALTER TABLE %s RENAME TO %s", safeSQLName(oldName), safeSQLName(newName)),
			Priority: 100,
		})
	})
}

// safeSQLName returns a quoted SQL identifier.
func safeSQLName(name string) string {
	if name == "" {
		return name
	}
	return fmt.Sprintf("\"%s\"", strings.ReplaceAll(name, "\"", "\"\""))
}
