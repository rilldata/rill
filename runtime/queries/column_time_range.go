package queries

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/marcboeker/go-duckdb"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ColumnTimeRange struct {
	TableName  string
	ColumnName string
	Result     *runtimev1.TimeRangeSummary
}

var _ runtime.Query = &ColumnTimeRange{}

func (q *ColumnTimeRange) Key() string {
	return fmt.Sprintf("ColumnTimeRange:%s:%s", q.TableName, q.ColumnName)
}

func (q *ColumnTimeRange) Deps() []string {
	return []string{q.TableName}
}

func (q *ColumnTimeRange) MarshalResult() any {
	return q.Result
}

func (q *ColumnTimeRange) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.TimeRangeSummary)
	if !ok {
		return fmt.Errorf("ColumnTimeRange: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *ColumnTimeRange) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	olap, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}

	switch olap.Dialect() {
	case drivers.DialectDuckDB:
		return q.resolveDuckDB(ctx, olap, priority)
	case drivers.DialectDruid:
		return q.resolveDruid(ctx, olap, priority)
	default:
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}
}

func (q *ColumnTimeRange) resolveDuckDB(ctx context.Context, olap drivers.OLAPStore, priority int) error {
	rangeSQL := fmt.Sprintf(
		"SELECT min(%[1]s) as \"min\", max(%[1]s) as \"max\", max(%[1]s) - min(%[1]s) as \"interval\" FROM %[2]s",
		safeName(q.ColumnName),
		safeName(q.TableName),
	)

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    rangeSQL,
		Priority: priority,
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
		q.Result = summary
		return nil
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	return errors.New("no rows returned")
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

func (q *ColumnTimeRange) resolveDruid(ctx context.Context, olap drivers.OLAPStore, priority int) error {
	rangeSQL := fmt.Sprintf(
		"SELECT min(%[1]s) as \"min\", max(%[1]s) as \"max\", timestampdiff(SECOND, max(%[1]s), min(%[1]s)) as \"interval\" FROM %[2]s",
		safeName(q.ColumnName),
		safeName(q.TableName),
	)

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    rangeSQL,
		Priority: priority,
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

			interval, ok := rowMap["interval"].(int64)
			if !ok {
				return fmt.Errorf("cannot handle interval type %T", interval)
			}
			summary.Interval = &runtimev1.TimeRangeSummary_Interval{
				Micros: interval * int64(time.Second) / int64(time.Microsecond),
			}
		}
		q.Result = summary
		return nil
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	return errors.New("no rows returned")
}
