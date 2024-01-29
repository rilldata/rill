package queries

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MetricsViewTimeRange struct {
	MetricsViewName    string                               `json:"name"`
	MetricsView        *runtimev1.MetricsViewSpec           `json:"-"`
	ResolvedMVSecurity *runtime.ResolvedMetricsViewSecurity `json:"security"`

	Result *runtimev1.MetricsViewTimeRangeResponse `json:"_"`
}

var _ runtime.Query = &MetricsViewTimeRange{}

func (q *MetricsViewTimeRange) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("MetricsViewTimeRange:%s", r)
}

func (q *MetricsViewTimeRange) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindMetricsView, Name: q.MetricsViewName},
	}
}

func (q *MetricsViewTimeRange) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
}

func (q *MetricsViewTimeRange) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewTimeRangeResponse)
	if !ok {
		return fmt.Errorf("MetricsViewTimeRange: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewTimeRange) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	policyFilter := ""
	if q.ResolvedMVSecurity != nil {
		policyFilter = q.ResolvedMVSecurity.RowFilter
	}

	if q.MetricsView.TimeDimension == "" {
		return fmt.Errorf("metrics view '%s' does not have a time dimension", q.MetricsViewName)
	}

	olap, release, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	switch olap.Dialect() {
	case drivers.DialectDuckDB:
		return q.resolveDuckDB(ctx, olap, q.MetricsView.TimeDimension, q.MetricsView.Table, policyFilter, priority)
	case drivers.DialectDruid:
		return q.resolveDruid(ctx, olap, q.MetricsView.TimeDimension, q.MetricsView.Table, policyFilter, priority)
	case drivers.DialectClickHouse:
		return q.resolveClickHouse(ctx, olap, q.MetricsView.TimeDimension, q.MetricsView.Table, policyFilter, priority)
	default:
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}
}

func (q *MetricsViewTimeRange) resolveDuckDB(ctx context.Context, olap drivers.OLAPStore, timeDim, tableName, filter string, priority int) error {
	if filter != "" {
		filter = fmt.Sprintf(" WHERE %s", filter)
	}

	rangeSQL := fmt.Sprintf(
		"SELECT min(%[1]s) as \"min\", max(%[1]s) as \"max\", max(%[1]s) - min(%[1]s) as \"interval\" FROM %[2]s %[3]s",
		safeName(timeDim),
		safeName(tableName),
		filter,
	)

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:            rangeSQL,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		summary := &runtimev1.TimeRangeSummary{}
		rowMap := make(map[string]any)
		err = rows.MapScan(rowMap)
		if err != nil {
			return err
		}
		if v := rowMap["min"]; v != nil {
			minTime, ok := v.(time.Time)
			if !ok {
				return fmt.Errorf("not a timestamp column")
			}
			summary.Min = timestamppb.New(minTime)
			summary.Max = timestamppb.New(rowMap["max"].(time.Time))
			summary.Interval, err = handleDuckDBInterval(rowMap["interval"])
			if err != nil {
				return err
			}
		}
		q.Result = &runtimev1.MetricsViewTimeRangeResponse{
			TimeRangeSummary: summary,
		}
		return nil
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	return errors.New("no rows returned")
}

func (q *MetricsViewTimeRange) resolveDruid(ctx context.Context, olap drivers.OLAPStore, timeDim, tableName, filter string, priority int) error {
	if filter != "" {
		filter = fmt.Sprintf(" WHERE %s", filter)
	}

	var minTime, maxTime time.Time
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		minSQL := fmt.Sprintf(
			"SELECT min(%[1]s) as \"min\" FROM %[2]s %[3]s",
			safeName(timeDim),
			safeName(tableName),
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
			safeName(timeDim),
			safeName(tableName),
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
		return err
	}

	summary := &runtimev1.TimeRangeSummary{}
	summary.Min = timestamppb.New(minTime)
	summary.Max = timestamppb.New(maxTime)
	summary.Interval = &runtimev1.TimeRangeSummary_Interval{
		Micros: maxTime.Sub(minTime).Microseconds(),
	}
	q.Result = &runtimev1.MetricsViewTimeRangeResponse{
		TimeRangeSummary: summary,
	}

	return nil
}

func (q *MetricsViewTimeRange) resolveClickHouse(ctx context.Context, olap drivers.OLAPStore, timeDim, tableName, filter string, priority int) error {
	if filter != "" {
		filter = fmt.Sprintf(" WHERE %s", filter)
	}

	rangeSQL := fmt.Sprintf(
		"SELECT min(%[1]s) as \"min\", max(%[1]s) as \"max\", max(%[1]s) - min(%[1]s) as \"interval\" FROM %[2]s %[3]s",
		safeName(timeDim),
		safeName(tableName),
		filter,
	)

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:            rangeSQL,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		summary := &runtimev1.TimeRangeSummary{}
		rowMap := make(map[string]any)
		err = rows.MapScan(rowMap)
		if err != nil {
			return err
		}
		if v := rowMap["min"]; v != nil {
			minTime, ok := v.(time.Time)
			if !ok {
				return fmt.Errorf("not a timestamp column")
			}
			maxTime := rowMap["max"].(time.Time)
			summary.Min = timestamppb.New(minTime)
			summary.Max = timestamppb.New(maxTime)
			summary.Interval = &runtimev1.TimeRangeSummary_Interval{
				Micros: maxTime.Sub(minTime).Microseconds(),
			}
		}
		q.Result = &runtimev1.MetricsViewTimeRangeResponse{
			TimeRangeSummary: summary,
		}
		return nil
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	return errors.New("no rows returned")
}

func (q *MetricsViewTimeRange) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	return ErrExportNotSupported
}
