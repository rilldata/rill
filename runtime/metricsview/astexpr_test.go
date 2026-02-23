package metricsview

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func TestArrayContainsCondition(t *testing.T) {
	mv := &runtimev1.MetricsViewSpec{
		Table:     "test_table",
		Database:  "",
		Connector: "",
		Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
			{Name: "id", Column: "id"},
			{Name: "tags", Column: "tags", Unnest: true},
			{Name: "city", Column: "city"},
		},
		Measures: []*runtimev1.MetricsViewSpec_Measure{
			{Name: "count", Expression: "count(*)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
		},
	}
	sec := skipMetricsViewSecurity{}

	tests := []struct {
		name     string
		dialect  drivers.Dialect
		dims     []Dimension
		where    *Expression
		wantSQL  string
		wantArgs []any
	}{
		{
			name:    "duckdb: in on unnest dim uses list_has_any",
			dialect: drivers.DialectDuckDB,
			where: &Expression{Condition: &Condition{
				Operator: OperatorIn,
				Expressions: []*Expression{
					{Name: "tags"},
					{Value: []any{"a", "b", "c"}},
				},
			}},
			wantSQL:  `(list_has_any(("tags"), [?,?,?]))`,
			wantArgs: []any{"a", "b", "c"},
		},
		{
			name:    "duckdb: nin on unnest dim uses NOT list_has_any",
			dialect: drivers.DialectDuckDB,
			where: &Expression{Condition: &Condition{
				Operator: OperatorNin,
				Expressions: []*Expression{
					{Name: "tags"},
					{Value: []any{"a", "b"}},
				},
			}},
			wantSQL:  `(NOT list_has_any(("tags"), [?,?]))`,
			wantArgs: []any{"a", "b"},
		},
		{
			name:    "duckdb: in on unnest dim with empty list",
			dialect: drivers.DialectDuckDB,
			where: &Expression{Condition: &Condition{
				Operator: OperatorIn,
				Expressions: []*Expression{
					{Name: "tags"},
					{Value: []any{}},
				},
			}},
			wantSQL:  `FALSE`,
			wantArgs: nil,
		},
		{
			name:    "duckdb: nin on unnest dim with empty list",
			dialect: drivers.DialectDuckDB,
			where: &Expression{Condition: &Condition{
				Operator: OperatorNin,
				Expressions: []*Expression{
					{Name: "tags"},
					{Value: []any{}},
				},
			}},
			wantSQL:  `TRUE`,
			wantArgs: nil,
		},
		{
			name:    "duckdb: in on unnest dim with null value in list",
			dialect: drivers.DialectDuckDB,
			where: &Expression{Condition: &Condition{
				Operator: OperatorIn,
				Expressions: []*Expression{
					{Name: "tags"},
					{Value: []any{"a", nil, "b"}},
				},
			}},
			wantSQL:  `(list_has_any(("tags"), [?,?,?]))`,
			wantArgs: []any{"a", nil, "b"},
		},
		{
			name:    "clickhouse: in on unnest dim uses hasAny",
			dialect: drivers.DialectClickHouse,
			where: &Expression{Condition: &Condition{
				Operator: OperatorIn,
				Expressions: []*Expression{
					{Name: "tags"},
					{Value: []any{"a", "b"}},
				},
			}},
			wantSQL:  `(hasAny(("tags"), [?,?]))`,
			wantArgs: []any{"a", "b"},
		},
		{
			name:    "clickhouse: nin on unnest dim uses NOT hasAny",
			dialect: drivers.DialectClickHouse,
			where: &Expression{Condition: &Condition{
				Operator: OperatorNin,
				Expressions: []*Expression{
					{Name: "tags"},
					{Value: []any{"a", "b"}},
				},
			}},
			wantSQL:  `(NOT hasAny(("tags"), [?,?]))`,
			wantArgs: []any{"a", "b"},
		},
		{
			name:    "clickhouse: in on unnest dim with null values",
			dialect: drivers.DialectClickHouse,
			where: &Expression{Condition: &Condition{
				Operator: OperatorIn,
				Expressions: []*Expression{
					{Name: "tags"},
					{Value: []any{nil, "a"}},
				},
			}},
			wantSQL:  `(hasAny(("tags"), [?,?]))`,
			wantArgs: []any{nil, "a"},
		},
		{
			name:    "duckdb: in on non-unnest dim uses normal IN",
			dialect: drivers.DialectDuckDB,
			where: &Expression{Condition: &Condition{
				Operator: OperatorIn,
				Expressions: []*Expression{
					{Name: "city"},
					{Value: []any{"NYC", "LA"}},
				},
			}},
			wantSQL:  `(("city") IN (?,?))`,
			wantArgs: []any{"NYC", "LA"},
		},
		{
			name:    "duckdb: nin on non-unnest dim uses normal NOT IN",
			dialect: drivers.DialectDuckDB,
			where: &Expression{Condition: &Condition{
				Operator: OperatorNin,
				Expressions: []*Expression{
					{Name: "city"},
					{Value: []any{"NYC"}},
				},
			}},
			wantSQL:  `(("city") NOT IN (?) OR ("city") IS NULL)`,
			wantArgs: []any{"NYC"},
		},
		{
			name:    "duckdb: in on unnest dim nested in AND",
			dialect: drivers.DialectDuckDB,
			where: &Expression{Condition: &Condition{
				Operator: OperatorAnd,
				Expressions: []*Expression{
					{Condition: &Condition{
						Operator: OperatorIn,
						Expressions: []*Expression{
							{Name: "tags"},
							{Value: []any{"a", "b"}},
						},
					}},
					{Condition: &Condition{
						Operator: OperatorEq,
						Expressions: []*Expression{
							{Name: "city"},
							{Value: "NYC"},
						},
					}},
				},
			}},
			wantSQL:  `((list_has_any(("tags"), [?,?])) AND (("city") = ?))`,
			wantArgs: []any{"a", "b", "NYC"},
		},
		{
			name:    "duckdb: in on unnest dim already in select falls back to normal IN",
			dialect: drivers.DialectDuckDB,
			dims:    []Dimension{{Name: "tags"}},
			where: &Expression{Condition: &Condition{
				Operator: OperatorIn,
				Expressions: []*Expression{
					{Name: "tags"},
					{Value: []any{"a", "b"}},
				},
			}},
			// When tags is in the query dimensions, it's already unnested via lateral join, so falls back to normal IN.
			// The AST qualifies the column with a table alias from the lateral unnest join.
			wantSQL:  `(("t0"."tags") IN (?,?))`,
			wantArgs: []any{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qry := &Query{
				MetricsView: "test",
				Dimensions:  tt.dims,
				Measures:    []Measure{{Name: "count"}},
				Where:       tt.where,
			}

			ast, err := NewAST(mv, sec, qry, tt.dialect)
			require.NoError(t, err)

			sql, args, err := ast.SQLForExpression(tt.where, nil, false, false)
			require.NoError(t, err)
			require.Equal(t, tt.wantSQL, sql)
			require.Equal(t, tt.wantArgs, args)
		})
	}
}
