package queries

import (
	"context"
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	sanitizedColumnName := quoteName(q.ColumnName)
	rangeSql := fmt.Sprintf(
		"SELECT min(%[1]s) as min, max(%[1]s) as max, max(%[1]s) - min(%[1]s) as interval FROM %[2]s",
		sanitizedColumnName,
		quoteName(q.TableName),
	)

	rows, err := rt.Execute(ctx, instanceID, priority, rangeSql)
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
			summary.Min = timestamppb.New(v.(time.Time))
			summary.Max = timestamppb.New(rowMap["max"].(time.Time))
			summary.Interval, err = server.handleInterval(rowMap["interval"])
			if err != nil {
				return err
			}
		}
		q.Result = summary
		return nil
	}
	return status.Error(codes.Internal, "no rows returned")
}
