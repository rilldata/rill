package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MetricsViewRows struct {
	MetricsViewName string                       `json:"metrics_view_name,omitempty"`
	TimeStart       *timestamppb.Timestamp       `json:"time_start,omitempty"`
	TimeEnd         *timestamppb.Timestamp       `json:"time_end,omitempty"`
	Filter          *runtimev1.MetricsViewFilter `json:"filter,omitempty"`
	Sort            []*runtimev1.MetricsViewSort `json:"sort,omitempty"`
	Limit           int32                        `json:"limit,omitempty"`
	Offset          int64                        `json:"offset,omitempty"`

	Result *runtimev1.MetricsViewRowsResponse `json:"-"`
}

var _ runtime.Query = &MetricsViewRows{}

func (q *MetricsViewRows) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("MetricsViewRows:%s", string(r))
}

func (q *MetricsViewRows) Deps() []string {
	return []string{q.MetricsViewName}
}

func (q *MetricsViewRows) MarshalResult() *runtime.CacheObject {
	return &runtime.CacheObject{
		Result:      q.Result,
		SizeInBytes: sizeProtoMessage(q.Result),
	}
}

func (q *MetricsViewRows) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewRowsResponse)
	if !ok {
		return fmt.Errorf("MetricsViewRows: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewRows) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	olap, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}

	if olap.Dialect() != drivers.DialectDuckDB && olap.Dialect() != drivers.DialectDruid {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	mv, err := lookupMetricsView(ctx, rt, instanceID, q.MetricsViewName)
	if err != nil {
		return err
	}

	if mv.TimeDimension == "" && (q.TimeStart != nil || q.TimeEnd != nil) {
		return fmt.Errorf("metrics view '%s' does not have a time dimension", q.MetricsViewName)
	}

	ql, args, err := q.buildMetricsRowsSQL(mv, olap.Dialect())
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	meta, data, err := metricsQuery(ctx, olap, priority, ql, args)
	if err != nil {
		return err
	}

	q.Result = &runtimev1.MetricsViewRowsResponse{
		Meta: meta,
		Data: data,
	}

	return nil
}

func (q *MetricsViewRows) buildMetricsRowsSQL(mv *runtimev1.MetricsView, dialect drivers.Dialect) (string, []any, error) {
	whereClause := "1=1"
	args := []any{}
	if mv.TimeDimension != "" {
		if q.TimeStart != nil {
			whereClause += fmt.Sprintf(" AND %s >= ?", safeName(mv.TimeDimension))
			args = append(args, q.TimeStart.AsTime())
		}
		if q.TimeEnd != nil {
			whereClause += fmt.Sprintf(" AND %s < ?", safeName(mv.TimeDimension))
			args = append(args, q.TimeEnd.AsTime())
		}
	}

	if q.Filter != nil {
		clause, clauseArgs, err := buildFilterClauseForMetricsViewFilter(q.Filter, dialect)
		if err != nil {
			return "", nil, err
		}
		whereClause += " " + clause
		args = append(args, clauseArgs...)
	}

	sortingCriteria := make([]string, 0, len(q.Sort))
	for _, s := range q.Sort {
		sortCriterion := safeName(s.Name)
		if !s.Ascending {
			sortCriterion += " DESC"
		}
		if dialect == drivers.DialectDuckDB {
			sortCriterion += " NULLS LAST"
		}
		sortingCriteria = append(sortingCriteria, sortCriterion)
	}
	orderClause := ""
	if len(sortingCriteria) > 0 {
		orderClause = "ORDER BY " + strings.Join(sortingCriteria, ", ")
	}

	if q.Limit == 0 {
		q.Limit = 100
	}

	sql := fmt.Sprintf("SELECT * FROM %q WHERE %s %s LIMIT %d",
		mv.Model,
		whereClause,
		orderClause,
		q.Limit,
	)

	return sql, args, nil
}
