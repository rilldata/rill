package queries

import (
	"context"
	"fmt"
	"io"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

type MetricsViewSchema struct {
	MetricsViewName string                                   `json:"metrics_view_name,omitempty"`
	TableName       string                                   `json:"table_name,omitempty"`
	Measures        []*runtimev1.MetricsViewSpec_MeasureV2   `json:"measures,omitempty"`
	Dimensions      []*runtimev1.MetricsViewSpec_DimensionV2 `json:"dimensions,omitempty"`

	Result *runtimev1.MetricsViewSchemaResponse `json:"-"`
}

var _ runtime.Query = &MetricsViewSchema{}

func (q *MetricsViewSchema) Key() string {
	return fmt.Sprintf("MetricsViewSchema:%s", q.MetricsViewName)
}

func (q *MetricsViewSchema) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindMetricsView, Name: q.MetricsViewName},
	}
}

func (q *MetricsViewSchema) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
}

func (q *MetricsViewSchema) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewSchemaResponse)
	if !ok {
		return fmt.Errorf("MetricsViewSchema: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewSchema) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	olap, release, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	schema, _, err := olapQuery(ctx, olap, priority, q.buildMetricsViewDataTypesSQL(olap.Dialect()), nil)
	if err != nil {
		return err
	}

	q.Result = &runtimev1.MetricsViewSchemaResponse{
		Schema: schema,
	}

	return nil
}

func (q *MetricsViewSchema) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	return nil
}

func (q *MetricsViewSchema) buildMetricsViewDataTypesSQL(dialect drivers.Dialect) string {
	var dimensions []string
	var unnestClauses []string
	for _, dim := range q.Dimensions {
		sel, unnestClause := dimensionSelect(q.TableName, dim, dialect)
		if unnestClause != "" {
			unnestClauses = append(unnestClauses, unnestClause)
		}
		dimensions = append(dimensions, sel)
	}

	var measures []string
	for _, meas := range q.Measures {
		measures = append(measures, fmt.Sprintf("%s as %s", meas.Expression, safeName(meas.Name)))
	}

	groups := make([]string, len(dimensions))
	for i := range dimensions {
		groups[i] = fmt.Sprintf("%d", i+1)
	}

	dimensionColumns := strings.Join(dimensions, ",")
	measureColumns := strings.Join(measures, ",")

	columns := dimensionColumns
	if measureColumns != "" {
		if columns != "" {
			columns += ","
		}
		columns += measureColumns
	}

	groupBy := strings.Join(groups, ",")
	if groupBy != "" {
		groupBy = fmt.Sprintf("GROUP BY %s", groupBy)
	}

	return fmt.Sprintf(
		`SELECT %[1]s FROM %[2]s %[3]s %[4]s LIMIT 0`,
		columns,                         // 1
		safeName(q.TableName),           // 2
		strings.Join(unnestClauses, ""), // 3
		groupBy,                         // 4
	)
}
