package queries

import (
	"context"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
)

type RollupInterval struct {
	TableName  string
	ColumnName string
	Result     *runtimev1.ColumnRollupIntervalResponse
}

var _ runtime.Query = &RollupInterval{}

func (q *RollupInterval) Key() string {
	return fmt.Sprintf("RollupInterval:%s:%s", q.TableName, q.ColumnName)
}

func (q *RollupInterval) Deps() []string {
	return []string{q.TableName}
}

func (q *RollupInterval) MarshalResult() any {
	return q.Result
}

func (q *RollupInterval) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.ColumnRollupIntervalResponse)
	if !ok {
		return fmt.Errorf("RollupInterval: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *RollupInterval) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	ctr := &ColumnTimeRange{
		TableName:  q.TableName,
		ColumnName: q.ColumnName,
	}
	err := rt.Query(ctx, instanceID, ctr, priority)
	if err != nil {
		return err
	}
	if ctr.Result.Interval == nil {
		q.Result = &runtimev1.ColumnRollupIntervalResponse{}
		return nil
	}
	r := ctr.Result.Interval

	const (
		microsSecond = 1000 * 1000
		microsMinute = 1000 * 1000 * 60
		microsHour   = 1000 * 1000 * 60 * 60
		microsDay    = 1000 * 1000 * 60 * 60 * 24
	)

	var rollupInterval runtimev1.TimeGrain
	if r.Days == 0 && r.Micros <= microsMinute {
		rollupInterval = runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND
	} else if r.Days == 0 && r.Micros > microsMinute && r.Micros <= microsHour {
		rollupInterval = runtimev1.TimeGrain_TIME_GRAIN_SECOND
	} else if r.Days == 0 && r.Micros <= microsDay {
		rollupInterval = runtimev1.TimeGrain_TIME_GRAIN_MINUTE
	} else if r.Days <= 7 {
		rollupInterval = runtimev1.TimeGrain_TIME_GRAIN_HOUR
	} else if r.Days <= 365*20 {
		rollupInterval = runtimev1.TimeGrain_TIME_GRAIN_DAY
	} else if r.Days <= 365*500 {
		rollupInterval = runtimev1.TimeGrain_TIME_GRAIN_MONTH
	} else {
		rollupInterval = runtimev1.TimeGrain_TIME_GRAIN_YEAR
	}

	q.Result = &runtimev1.ColumnRollupIntervalResponse{
		Interval: rollupInterval,
		Start:    ctr.Result.Min,
		End:      ctr.Result.Max,
	}
	return nil
}
