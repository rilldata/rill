package queries

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

type MetricsViewDataTypes struct {
	MetricsViewName    string                               `json:"metrics_view_name,omitempty"`
	MetricsView        *runtimev1.MetricsViewSpec           `json:"-"`
	ResolvedMVSecurity *runtime.ResolvedMetricsViewSecurity `json:"security"`

	Result []*runtimev1.MetricsViewDataType `json:"-"`
}

var _ runtime.Query = &MetricsViewDataTypes{}

func (q *MetricsViewDataTypes) Key() string {
	return fmt.Sprintf("TableColumns:%s", q.MetricsViewName)
}

func (q *MetricsViewDataTypes) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindSource, Name: q.MetricsView.Table},
		{Kind: runtime.ResourceKindModel, Name: q.MetricsView.Table},
		{Kind: runtime.ResourceKindMetricsView, Name: q.MetricsViewName},
	}
}

func (q *MetricsViewDataTypes) MarshalResult() *runtime.QueryResult {
	var size int64
	if len(q.Result) > 0 {
		// approx
		size = sizeProtoMessage(q.Result[0]) * int64(len(q.Result))
	}
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: size,
	}
}

func (q *MetricsViewDataTypes) UnmarshalResult(v any) error {
	res, ok := v.([]*runtimev1.MetricsViewDataType)
	if !ok {
		return fmt.Errorf("TableColumns: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewDataTypes) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	olap, release, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	if olap.Dialect() != drivers.DialectDuckDB {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	return olap.WithConnection(ctx, priority, false, false, func(ctx context.Context, ensuredCtx context.Context, _ *sql.Conn) error {
		// views return duplicate column names, so we need to create a temporary table
		temporaryTableName := tempName("profile_columns_")

		err = olap.Exec(ctx, &drivers.Statement{
			Query:            fmt.Sprintf(`CREATE TEMPORARY TABLE "%s" AS (%s)`, temporaryTableName, q.buildMetricsViewDataTypesSQL()),
			Priority:         priority,
			ExecutionTimeout: defaultExecutionTimeout,
		})
		if err != nil {
			return err
		}
		defer func() {
			// NOTE: Using ensuredCtx
			_ = olap.Exec(ensuredCtx, &drivers.Statement{
				Query:            `DROP TABLE "` + temporaryTableName + `"`,
				Priority:         priority,
				ExecutionTimeout: defaultExecutionTimeout,
			})
		}()

		rows, err := olap.Execute(ctx, &drivers.Statement{
			Query: fmt.Sprintf(`
				SELECT column_name AS name, data_type AS type
				FROM information_schema.columns
				WHERE table_catalog = 'temp' AND table_name = '%s'`, temporaryTableName),
			Priority:         priority,
			ExecutionTimeout: defaultExecutionTimeout,
		})
		if err != nil {
			return err
		}
		defer rows.Close()

		var pcs []*runtimev1.MetricsViewDataType
		i := 0
		for rows.Next() {
			pc := runtimev1.MetricsViewDataType{}
			if err := rows.StructScan(&pc); err != nil {
				return err
			}
			// TODO: Find a better way to handle this, this is ugly
			if strings.Contains(pc.Type, "ENUM") {
				pc.Type = "VARCHAR"
			}
			pcs = append(pcs, &pc)
			i++
		}

		q.Result = pcs[0:i]
		return nil
	})
}

func (q *MetricsViewDataTypes) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	return nil
}

func (q *MetricsViewDataTypes) buildMetricsViewDataTypesSQL() string {
	var dimensions []string
	var unnestClauses []string
	for _, dim := range q.MetricsView.Dimensions {
		sel, unnestClause := dimensionSelect(q.MetricsView, dim, drivers.DialectDuckDB)
		if unnestClause != "" {
			unnestClauses = append(unnestClauses, unnestClause)
		}
		dimensions = append(dimensions, sel)
	}

	var measures []string
	for _, meas := range q.MetricsView.Measures {
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
		`SELECT %[1]s FROM "%[2]s" %[3]s %[4]s LIMIT 1`,
		columns,                         // 1
		q.MetricsView.Table,             // 2
		strings.Join(unnestClauses, ""), // 3
		groupBy,                         // 4
	)
}
