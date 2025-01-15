package queries

import (
	"context"
	"fmt"
	"io"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
)

type RollupInterval struct {
	Connector      string
	Database       string
	DatabaseSchema string
	TableName      string
	ColumnName     string
	Result         *runtimev1.ColumnRollupIntervalResponse
}

var _ runtime.Query = &RollupInterval{}

func (q *RollupInterval) Key() string {
	return fmt.Sprintf("RollupInterval:%s:%s", q.TableName, q.ColumnName)
}

func (q *RollupInterval) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindSource, Name: q.TableName},
		{Kind: runtime.ResourceKindModel, Name: q.TableName},
	}
}

func (q *RollupInterval) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
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
		Connector:      q.Connector,
		Database:       q.Database,
		DatabaseSchema: q.DatabaseSchema,
		TableName:      q.TableName,
		ColumnName:     q.ColumnName,
	}
	err := rt.Query(ctx, instanceID, ctr, priority)
	if err != nil {
		return err
	}

	duration := ctr.Result.Max.AsTime().Sub(ctr.Result.Min.AsTime())
	hours := duration.Hours()
	days := int64(0)
	if hours >= hourInDay {
		days = int64(hours / hourInDay)
	}
	micros := duration.Microseconds() - microsInDay*days

	const (
		microsSecond = 1000 * 1000
		microsMinute = 1000 * 1000 * 60
		microsHour   = 1000 * 1000 * 60 * 60
		microsDay    = 1000 * 1000 * 60 * 60 * 24
	)

	var rollupInterval runtimev1.TimeGrain
	if days == 0 && micros <= microsMinute {
		rollupInterval = runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND
	} else if days == 0 && micros > microsMinute && micros <= microsHour {
		rollupInterval = runtimev1.TimeGrain_TIME_GRAIN_SECOND
	} else if days == 0 && micros <= microsDay {
		rollupInterval = runtimev1.TimeGrain_TIME_GRAIN_MINUTE
	} else if days <= 7 {
		rollupInterval = runtimev1.TimeGrain_TIME_GRAIN_HOUR
	} else if days <= 365*20 {
		rollupInterval = runtimev1.TimeGrain_TIME_GRAIN_DAY
	} else if days <= 365*500 {
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

func (q *RollupInterval) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	return ErrExportNotSupported
}
