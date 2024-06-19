package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/druid"
	"github.com/rilldata/rill/runtime/pkg/expressionpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MetricsViewSearch struct {
	MetricsViewName    string                                `json:"metrics_view_name,omitempty"`
	Dimensions         []string                              `json:"dimensions,omitempty"`
	Search             string                                `json:"search,omitempty"`
	TimeRange          *runtimev1.TimeRange                  `json:"time_range,omitempty"`
	Where              *runtimev1.Expression                 `json:"where,omitempty"`
	Having             *runtimev1.Expression                 `json:"having,omitempty"`
	Priority           int32                                 `json:"priority,omitempty"`
	Limit              *int64                                `json:"limit,omitempty"`
	SecurityAttributes map[string]any                        `json:"security_attributes,omitempty"`
	SecurityPolicy     *runtimev1.MetricsViewSpec_SecurityV2 `json:"security_policy,omitempty"`

	Result *runtimev1.MetricsViewSearchResponse
}

func (q *MetricsViewSearch) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("MetricsViewSearch:%s", string(r))
}

func (q *MetricsViewSearch) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindMetricsView, Name: q.MetricsViewName},
	}
}

func (q *MetricsViewSearch) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
}

func (q *MetricsViewSearch) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewSearchResponse)
	if !ok {
		return fmt.Errorf("MetricsViewSearch: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewSearch) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	mv, sec, err := resolveMVAndSecurityFromAttributes(ctx, rt, instanceID, q.MetricsViewName, q.SecurityAttributes, q.SecurityPolicy, nil, nil)
	if err != nil {
		return err
	}
	for _, d := range q.Dimensions {
		if !checkFieldAccess(d, sec) {
			return ErrForbidden
		}
	}

	olap, release, err := rt.OLAP(ctx, instanceID, mv.Connector)
	if err != nil {
		return err
	}
	defer release()

	if olap.Dialect() != drivers.DialectDuckDB && olap.Dialect() != drivers.DialectDruid && olap.Dialect() != drivers.DialectClickHouse && olap.Dialect() != drivers.DialectPinot {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	if mv.TimeDimension == "" && !isTimeRangeNil(q.TimeRange) {
		return fmt.Errorf("metrics view '%s' does not have a time dimension", mv)
	}

	if !isTimeRangeNil(q.TimeRange) {
		start, end, err := ResolveTimeRange(q.TimeRange, mv)
		if err != nil {
			return err
		}
		q.TimeRange = &runtimev1.TimeRange{
			Start: timestamppb.New(start.In(time.UTC)),
			End:   timestamppb.New(end.In(time.UTC)),
		}
	}

	if olap.Dialect() == drivers.DialectDruid {
		ok, err := q.executeSearchInDruid(ctx, rt, olap, instanceID, mv.Table, sec)
		if err != nil || ok {
			return err
		}
	}

	sql, args, err := q.buildSearchQuerySQL(mv, olap.Dialect(), sec)
	if err != nil {
		return err
	}

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:            sql,
		Args:             args,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return nil
	}
	defer rows.Close()

	q.Result = &runtimev1.MetricsViewSearchResponse{Results: make([]*runtimev1.MetricsViewSearchResponse_SearchResult, 0)}
	for rows.Next() {
		res := map[string]any{}
		err := rows.MapScan(res)
		if err != nil {
			return err
		}

		dimName, ok := res["dimension"].(string)
		if !ok {
			return fmt.Errorf("unknown result dimension: %q", dimName)
		}

		v, err := structpb.NewValue(res["value"])
		if err != nil {
			return err
		}

		q.Result.Results = append(q.Result.Results, &runtimev1.MetricsViewSearchResponse_SearchResult{
			Dimension: dimName,
			Value:     v,
		})
	}

	return nil
}

func (q *MetricsViewSearch) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	return nil
}

var druidSQLDSN = regexp.MustCompile(`/v2/sql/?`)

func (q *MetricsViewSearch) executeSearchInDruid(ctx context.Context, rt *runtime.Runtime, olap drivers.OLAPStore, instanceID, table string, policy *runtime.ResolvedMetricsViewSecurity) (bool, error) {
	var query map[string]interface{}
	if policy != nil && policy.RowFilter != "" {
		rows, err := olap.Execute(ctx, &drivers.Statement{
			Query:            fmt.Sprintf("EXPLAIN PLAN FOR SELECT 1 FROM %s WHERE %s", table, policy.RowFilter),
			Args:             nil,
			DryRun:           false,
			Priority:         0,
			LongRunning:      false,
			ExecutionTimeout: 0,
		})
		if err != nil {
			return false, err
		}

		if !rows.Next() {
			return false, fmt.Errorf("failed to parse policy filter")
		}

		var planRaw string
		var resRaw string
		var attrRaw string
		err = rows.Scan(&planRaw, &resRaw, &attrRaw)
		if err != nil {
			return false, err
		}

		var plan []druid.QueryPlan
		err = json.Unmarshal([]byte(planRaw), &plan)
		if err != nil {
			return false, err
		}

		if len(plan) == 0 {
			return false, fmt.Errorf("failed to parse policy filter")
		}
		if plan[0].Query.Filter == nil {
			// if we failed to parse a filter we return and run UNION query.
			// this can happen when the row filter is complex
			// TODO: iterate over this and integrate more parts like joins and subfilter in policy filter
			return false, nil
		}
		query = *plan[0].Query.Filter
	}

	inst, err := rt.Instance(ctx, instanceID)
	if err != nil {
		return false, err
	}

	dsn := ""
	for _, c := range inst.Connectors {
		if c.Name == "druid" {
			dsn, err = druid.GetDSN(c.Config)
			if err != nil {
				return false, err
			}
			break
		}
	}
	if dsn == "" {
		return false, fmt.Errorf("druid connector config not found in instance")
	}

	nq := druid.NewNativeQuery(druidSQLDSN.ReplaceAllString(dsn, "/v2/"))
	req := druid.NewNativeSearchQueryRequest(table, q.Search, q.Dimensions, q.TimeRange.Start.AsTime(), q.TimeRange.End.AsTime(), query)
	var res druid.NativeSearchQueryResponse
	err = nq.Do(ctx, req, &res, req.Context.QueryID)
	if err != nil {
		return false, err
	}

	q.Result = &runtimev1.MetricsViewSearchResponse{Results: make([]*runtimev1.MetricsViewSearchResponse_SearchResult, 0)}
	for _, re := range res {
		for _, r := range re.Result {
			v, err := structpb.NewValue(r.Value)
			if err != nil {
				return false, err
			}
			q.Result.Results = append(q.Result.Results, &runtimev1.MetricsViewSearchResponse_SearchResult{
				Dimension: r.Dimension,
				Value:     v,
			})
		}
	}

	return true, nil
}

func (q *MetricsViewSearch) buildSearchQuerySQL(mv *runtimev1.MetricsViewSpec, dialect drivers.Dialect, policy *runtime.ResolvedMetricsViewSecurity) (string, []any, error) {
	var baseWhereClause string
	if policy != nil && policy.RowFilter != "" {
		baseWhereClause += fmt.Sprintf(" AND (%s)", policy.RowFilter)
	}

	var args []any

	unions := make([]string, len(q.Dimensions))
	for i, dimName := range q.Dimensions {
		var dim *runtimev1.MetricsViewSpec_DimensionV2
		for _, d := range mv.Dimensions {
			if d.Name == dimName {
				dim = d
				break
			}
		}
		if dim == nil {
			return "", nil, fmt.Errorf("dimension not found: %q", q.Dimensions[i])
		}

		expr, _, unnest := dialect.DimensionSelectPair(mv.Database, mv.DatabaseSchema, mv.Table, dim)
		filterBuilder := &ExpressionBuilder{
			mv:      mv,
			dialect: dialect,
		}
		clause, clauseArgs, err := filterBuilder.buildExpression(expressionpb.Like(expressionpb.Identifier(dimName), expressionpb.String(fmt.Sprintf("%%%s%%", q.Search))))
		if err != nil {
			return "", nil, err
		}
		if clause != "" {
			clause = " AND " + clause
			args = append(args, clauseArgs...)
		}

		unions[i] = fmt.Sprintf(
			`SELECT %s as "value", '%s' as dimension from %s %s WHERE 1=1 %s %s GROUP BY 1`,
			expr,
			dimName,
			mv.Table,
			unnest,
			baseWhereClause,
			clause,
		)
	}

	return strings.Join(unions, " UNION ALL "), args, nil
}
