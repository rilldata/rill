package resolvers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/marcboeker/go-duckdb"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	defaultExecutionTimeout = time.Minute * 3
	hourInDay               = 24
)

var microsInDay = hourInDay * time.Hour.Microseconds()

func init() {
	runtime.RegisterResolverInitializer("metrics_time_range", newMetricsViewTimeRangeResolver)
}

type metricsViewTimeRangeResolver struct {
	runtime            *runtime.Runtime
	instanceID         string
	mvName             string
	mv                 *runtimev1.MetricsViewSpec
	resolvedMVSecurity *runtime.ResolvedSecurity
	args               *metricsViewTimeRangeResolverArgs
}

type metricsViewTimeRangeResolverArgs struct {
	Priority int `mapstructure:"priority"`
}

type metricsViewTimeRange struct {
	MetricsView string `mapstructure:"metrics_view"`
}

func newMetricsViewTimeRangeResolver(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	tr := &metricsViewTimeRange{}
	if err := mapstructureutil.WeakDecode(opts.Properties, tr); err != nil {
		return nil, err
	}

	args := &metricsViewTimeRangeResolverArgs{}
	if err := mapstructureutil.WeakDecode(opts.Args, args); err != nil {
		return nil, err
	}

	ctrl, err := opts.Runtime.Controller(ctx, opts.InstanceID)
	if err != nil {
		return nil, err
	}

	res, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: tr.MetricsView}, false)
	if err != nil {
		return nil, err
	}

	mv := res.GetMetricsView().State.ValidSpec
	if mv == nil {
		return nil, fmt.Errorf("metrics view %q is invalid", res.Meta.Name.Name)
	}

	if mv.TimeDimension == "" {
		return nil, fmt.Errorf("metrics view '%s' does not have a time dimension", tr.MetricsView)
	}

	security, err := opts.Runtime.ResolveSecurity(opts.InstanceID, opts.Claims, res)
	if err != nil {
		return nil, err
	}

	if !security.CanAccess() {
		return nil, runtime.ErrForbidden
	}

	return &metricsViewTimeRangeResolver{
		runtime:            opts.Runtime,
		instanceID:         opts.InstanceID,
		mvName:             tr.MetricsView,
		mv:                 mv,
		resolvedMVSecurity: security,
		args:               args,
	}, nil
}

func (r *metricsViewTimeRangeResolver) Close() error {
	return nil
}

func (r *metricsViewTimeRangeResolver) CacheKey(ctx context.Context) ([]byte, bool, error) {
	// todo : fix the implementation to use executor
	// this resolver is only used in health check so okay to not cache for now
	return nil, false, nil
}

func (r *metricsViewTimeRangeResolver) Refs() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: r.mvName}}
}

func (r *metricsViewTimeRangeResolver) Validate(ctx context.Context) error {
	return nil
}

func (r *metricsViewTimeRangeResolver) ResolveInteractive(ctx context.Context) (runtime.ResolverResult, error) {
	olap, release, err := r.runtime.OLAP(ctx, r.instanceID, r.mv.Connector)
	if err != nil {
		return nil, err
	}
	defer release()

	var tr *runtimev1.TimeRangeSummary
	switch olap.Dialect() {
	case drivers.DialectDuckDB:
		tr, err = r.resolveDuckDB(ctx, olap, r.mv.TimeDimension, escapeMetricsViewTable(drivers.DialectDuckDB, r.mv), r.resolvedMVSecurity.RowFilter(), r.args.Priority)
	case drivers.DialectDruid:
		tr, err = r.resolveDruid(ctx, olap, r.mv.TimeDimension, escapeMetricsViewTable(drivers.DialectDruid, r.mv), r.resolvedMVSecurity.RowFilter(), r.args.Priority)
	case drivers.DialectClickHouse:
		tr, err = r.resolveClickHouseAndPinot(ctx, olap, r.mv.TimeDimension, escapeMetricsViewTable(drivers.DialectClickHouse, r.mv), r.resolvedMVSecurity.RowFilter(), r.args.Priority)
	case drivers.DialectPinot:
		tr, err = r.resolveClickHouseAndPinot(ctx, olap, r.mv.TimeDimension, escapeMetricsViewTable(drivers.DialectPinot, r.mv), r.resolvedMVSecurity.RowFilter(), r.args.Priority)
	default:
		return nil, fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}
	if err != nil {
		return nil, err
	}

	row := map[string]any{}
	if tr.Min != nil {
		row["min"] = tr.Min.AsTime()
		row["max"] = tr.Max.AsTime()
		row["interval"] = map[string]any{
			"days":   tr.Interval.Days,
			"months": tr.Interval.Months,
			"micros": tr.Interval.Micros,
		}
	}
	schema := &runtimev1.StructType{
		Fields: []*runtimev1.StructType_Field{
			{Name: "min", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP, Nullable: true}},
			{Name: "max", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP, Nullable: true}},
			{Name: "interval", Type: &runtimev1.Type{
				Code: runtimev1.Type_CODE_STRUCT,
				StructType: &runtimev1.StructType{
					Fields: []*runtimev1.StructType_Field{
						{Name: "days", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_INT32}},
						{Name: "months", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_INT32}},
						{Name: "micros", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_INT64}},
					},
				},
				Nullable: true,
			}},
		},
	}
	return runtime.NewMapsResolverResult([]map[string]any{row}, schema), nil
}

func (r *metricsViewTimeRangeResolver) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ResolverExportOptions) error {
	return errors.New("not implemented")
}

func (r *metricsViewTimeRangeResolver) resolveDuckDB(ctx context.Context, olap drivers.OLAPStore, timeDim, escapedTableName, filter string, priority int) (*runtimev1.TimeRangeSummary, error) {
	rangeSQL := fmt.Sprintf(
		"SELECT min(%[1]s) as \"min\", max(%[1]s) as \"max\", max(%[1]s) - min(%[1]s) as \"interval\" FROM %[2]s %[3]s",
		olap.Dialect().EscapeIdentifier(timeDim),
		escapedTableName,
		filter,
	)

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:            rangeSQL,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		summary := &runtimev1.TimeRangeSummary{}
		rowMap := make(map[string]any)
		err = rows.MapScan(rowMap)
		if err != nil {
			return nil, err
		}
		if v := rowMap["min"]; v != nil {
			minTime, ok := v.(time.Time)
			if !ok {
				return nil, fmt.Errorf("not a timestamp column")
			}
			summary.Min = timestamppb.New(minTime)
			summary.Max = timestamppb.New(rowMap["max"].(time.Time))
			summary.Interval, err = handleDuckDBInterval(rowMap["interval"])
			if err != nil {
				return nil, err
			}
		}
		return summary, nil
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return nil, errors.New("no rows returned")
}

func (r *metricsViewTimeRangeResolver) resolveDruid(ctx context.Context, olap drivers.OLAPStore, timeDim, escapedTableName, filter string, priority int) (*runtimev1.TimeRangeSummary, error) {
	if filter != "" {
		filter = fmt.Sprintf(" WHERE %s", filter)
	}

	var minTime, maxTime time.Time
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		minSQL := fmt.Sprintf(
			"SELECT min(%[1]s) as \"min\" FROM %[2]s %[3]s",
			olap.Dialect().EscapeIdentifier(timeDim),
			escapedTableName,
			filter,
		)

		rows, err := olap.Execute(ctx, &drivers.Statement{
			Query:            minSQL,
			Priority:         priority,
			ExecutionTimeout: defaultExecutionTimeout,
		})
		if err != nil {
			return err
		}
		defer rows.Close()

		if rows.Next() {
			err = rows.Scan(&minTime)
			if err != nil {
				return err
			}
		} else {
			err = rows.Err()
			if err != nil {
				return err
			}
			return errors.New("no rows returned for min time")
		}

		return nil
	})

	group.Go(func() error {
		maxSQL := fmt.Sprintf(
			"SELECT max(%[1]s) as \"max\" FROM %[2]s %[3]s",
			olap.Dialect().EscapeIdentifier(timeDim),
			escapedTableName,
			filter,
		)

		rows, err := olap.Execute(ctx, &drivers.Statement{
			Query:            maxSQL,
			Priority:         priority,
			ExecutionTimeout: defaultExecutionTimeout,
		})
		if err != nil {
			return err
		}
		defer rows.Close()

		if rows.Next() {
			err = rows.Scan(&maxTime)
			if err != nil {
				return err
			}
		} else {
			err = rows.Err()
			if err != nil {
				return err
			}
			return errors.New("no rows returned for max time")
		}
		return nil
	})

	err := group.Wait()
	if err != nil {
		return nil, err
	}

	summary := &runtimev1.TimeRangeSummary{}
	summary.Min = timestamppb.New(minTime)
	summary.Max = timestamppb.New(maxTime)
	summary.Interval = &runtimev1.TimeRangeSummary_Interval{
		Micros: maxTime.Sub(minTime).Microseconds(),
	}
	return summary, nil
}

func (r *metricsViewTimeRangeResolver) resolveClickHouseAndPinot(ctx context.Context, olap drivers.OLAPStore, timeDim, escapedTableName, filter string, priority int) (*runtimev1.TimeRangeSummary, error) {
	if filter != "" {
		filter = fmt.Sprintf(" WHERE %s", filter)
	}

	rangeSQL := fmt.Sprintf(
		"SELECT min(%[1]s) AS \"min\", max(%[1]s) AS \"max\" FROM %[2]s %[3]s",
		olap.Dialect().EscapeIdentifier(timeDim),
		escapedTableName,
		filter,
	)

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:            rangeSQL,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		summary := &runtimev1.TimeRangeSummary{}
		var minVal, maxVal *time.Time
		err = rows.Scan(&minVal, &maxVal)
		if err != nil {
			return nil, err
		}

		if minVal != nil {
			summary.Min = timestamppb.New(*minVal)
		}
		if maxVal != nil {
			summary.Max = timestamppb.New(*maxVal)
		}
		if minVal != nil && maxVal != nil {
			// ignoring months for now since its hard to compute and anyways not being used
			summary.Interval = &runtimev1.TimeRangeSummary_Interval{}
			duration := maxVal.Sub(*minVal)
			hours := duration.Hours()
			if hours >= hourInDay {
				summary.Interval.Days = int32(hours / hourInDay)
			}
			summary.Interval.Micros = duration.Microseconds() - microsInDay*int64(summary.Interval.Days)
		}
		return summary, nil
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return nil, errors.New("no rows returned")
}

func handleDuckDBInterval(interval any) (*runtimev1.TimeRangeSummary_Interval, error) {
	switch i := interval.(type) {
	case duckdb.Interval:
		result := new(runtimev1.TimeRangeSummary_Interval)
		result.Days = i.Days
		result.Months = i.Months
		result.Micros = i.Micros
		return result, nil
	case int64:
		// for date type column interval is difference in num days for two dates
		result := new(runtimev1.TimeRangeSummary_Interval)
		result.Days = int32(i)
		return result, nil
	}
	return nil, fmt.Errorf("cannot handle interval type %T", interval)
}

func escapeMetricsViewTable(d drivers.Dialect, mv *runtimev1.MetricsViewSpec) string {
	return d.EscapeTable(mv.Database, mv.DatabaseSchema, mv.Table)
}
