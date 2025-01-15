package queries

import (
	"context"
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

const hourInDay = 24

var microsInDay = hourInDay * time.Hour.Microseconds()

type ColumnTimeRange struct {
	Connector      string
	Database       string
	DatabaseSchema string
	TableName      string
	ColumnName     string
	Result         *runtimev1.TimeRangeSummary
}

var _ runtime.Query = &ColumnTimeRange{}

func (q *ColumnTimeRange) Key() string {
	return fmt.Sprintf("ColumnTimeRange:%s:%s", q.TableName, q.ColumnName)
}

func (q *ColumnTimeRange) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindSource, Name: q.TableName},
		{Kind: runtime.ResourceKindModel, Name: q.TableName},
	}
}

func (q *ColumnTimeRange) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
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
	olap, release, err := rt.OLAP(ctx, instanceID, q.Connector)
	if err != nil {
		return err
	}
	defer release()

	switch olap.Dialect() {
	case drivers.DialectDuckDB:
		return q.resolveDuckDB(ctx, olap, priority)
	case drivers.DialectDruid:
		return q.resolveDruid(ctx, olap, priority)
	case drivers.DialectClickHouse:
		return q.resolveClickHouse(ctx, olap, priority)
	default:
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}
}

func (q *ColumnTimeRange) resolveDuckDB(ctx context.Context, olap drivers.OLAPStore, priority int) error {
	rangeSQL := fmt.Sprintf(
		"SELECT min(%[1]s) as \"min\", max(%[1]s) as \"max\", max(%[1]s) - min(%[1]s) as \"interval\" FROM %[2]s",
		safeName(q.ColumnName),
		drivers.DialectDuckDB.EscapeTable(q.Database, q.DatabaseSchema, q.TableName),
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

func (q *ColumnTimeRange) resolveDruid(ctx context.Context, olap drivers.OLAPStore, priority int) error {
	var minTime, maxTime time.Time
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		minSQL := fmt.Sprintf(
			"SELECT min(%[1]s) as \"min\" FROM %[2]s",
			safeName(q.ColumnName),
			drivers.DialectDruid.EscapeTable(q.Database, q.DatabaseSchema, q.TableName),
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
			"SELECT max(%[1]s) as \"max\" FROM %[2]s",
			safeName(q.ColumnName),
			drivers.DialectDruid.EscapeTable(q.Database, q.DatabaseSchema, q.TableName),
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
	q.Result = summary

	return nil
}

func (q *ColumnTimeRange) resolveClickHouse(ctx context.Context, olap drivers.OLAPStore, priority int) error {
	sql := fmt.Sprintf(
		"SELECT min(%[1]s) as \"min\", max(%[1]s) as \"max\" FROM %[2]s",
		safeName(q.ColumnName),
		drivers.DialectClickHouse.EscapeTable(q.Database, q.DatabaseSchema, q.TableName),
	)

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:            sql,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return err
	}
	defer rows.Close()

	var minTime, maxTime *time.Time
	for rows.Next() {
		err = rows.Scan(&minTime, &maxTime)
		if err != nil {
			return err
		}
	}

	summary := &runtimev1.TimeRangeSummary{}
	if minTime != nil {
		summary.Min = timestamppb.New(*minTime)
	}
	if maxTime != nil {
		summary.Max = timestamppb.New(*maxTime)
	}

	q.Result = summary

	return nil
}

func (q *ColumnTimeRange) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	return ErrExportNotSupported
}
