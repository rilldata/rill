package metricssql_test

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers/duckdb"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/metricssql"
	"github.com/stretchr/testify/require"
)

// An IN or NIN with no values has no SQL form of its own, so ExpressionToSQL writes it as a comparison of constants.
// Parsing that back and compiling it must keep the meaning: no row passes the IN and every row passes the NIN.
// Canvas default filters take this round trip when they are saved to YAML and parsed again.
func TestParseFilterRoundTripsMembershipWithoutValues(t *testing.T) {
	mv := &runtimev1.MetricsViewSpec{
		Table: "tbl",
		Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
			{Name: "dim", Column: "dim"},
		},
		Measures: []*runtimev1.MetricsViewSpec_Measure{
			{Name: "count", Expression: "count(*)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
		},
	}

	tests := []struct {
		name         string
		expr         *metricsview.Expression
		wantSQL      string
		wantCompiled string
		wantArgs     []any
	}{
		{
			name: "in without values",
			expr: &metricsview.Expression{Condition: &metricsview.Condition{
				Operator:    metricsview.OperatorIn,
				Expressions: []*metricsview.Expression{{Name: "dim"}},
			}},
			wantSQL:      "1 = 0",
			wantCompiled: "(? = ?)",
			wantArgs:     []any{1, 0},
		},
		{
			name: "in with empty list",
			expr: &metricsview.Expression{Condition: &metricsview.Condition{
				Operator:    metricsview.OperatorIn,
				Expressions: []*metricsview.Expression{{Name: "dim"}, {Value: []any{}}},
			}},
			wantSQL:      "1 = 0",
			wantCompiled: "(? = ?)",
			wantArgs:     []any{1, 0},
		},
		{
			name: "nin without values",
			expr: &metricsview.Expression{Condition: &metricsview.Condition{
				Operator:    metricsview.OperatorNin,
				Expressions: []*metricsview.Expression{{Name: "dim"}},
			}},
			wantSQL:      "1 = 1",
			wantCompiled: "(? = ?)",
			wantArgs:     []any{1, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, err := metricsview.ExpressionToSQL(tt.expr)
			require.NoError(t, err)
			require.Equal(t, tt.wantSQL, sql)

			parsed, err := metricssql.ParseFilter(sql)
			require.NoError(t, err)

			qry := &metricsview.Query{
				MetricsView: "mv",
				Measures:    []metricsview.Measure{{Name: "count"}},
				Where:       parsed,
			}
			ast, err := metricsview.NewAST(mv, allowAllSecurity{}, qry, duckdb.DialectDuckDB)
			require.NoError(t, err)

			compiled, args, err := ast.SQLForExpression(parsed, nil, false, false)
			require.NoError(t, err)
			require.Equal(t, tt.wantCompiled, compiled)
			require.Equal(t, tt.wantArgs, args)
		})
	}
}

// allowAllSecurity is a metrics view security policy that grants access to everything.
type allowAllSecurity struct{}

func (allowAllSecurity) CanAccessField(string) bool { return true }

func (allowAllSecurity) RowFilter() string { return "" }

func (allowAllSecurity) QueryFilter() *runtimev1.Expression { return nil }
